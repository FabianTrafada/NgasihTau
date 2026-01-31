package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"ngasihtau/internal/common/config"
	"ngasihtau/internal/common/health"
	"ngasihtau/internal/common/middleware"
	"ngasihtau/internal/search/application"
	"ngasihtau/internal/search/infrastructure/meilisearch"
	"ngasihtau/internal/search/infrastructure/mock"
	searchopenai "ngasihtau/internal/search/infrastructure/openai"
	"ngasihtau/internal/search/infrastructure/qdrant"
	"ngasihtau/internal/search/interfaces/http"
	"ngasihtau/pkg/jwt"
	"ngasihtau/pkg/nats"
)

func main() {
	// Initialize logger
	zerolog.TimeFieldFormat = time.RFC3339
	if os.Getenv("APP_ENV") == "development" {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}

	// Load configuration
	configPath := os.Getenv("CONFIG_PATH")
	cfg, err := config.Load(configPath)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to load configuration")
	}

	// Set log level
	level, err := zerolog.ParseLevel(cfg.App.LogLevel)
	if err != nil {
		level = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(level)

	log.Info().
		Str("service", "search-service").
		Str("env", cfg.App.Env).
		Msg("starting search service")

	// Create application context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize Meilisearch client
	meiliClient, err := meilisearch.NewClient(cfg.Meilisearch.Host, cfg.Meilisearch.APIKey)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to initialize meilisearch client")
	}

	// Setup Meilisearch indexes
	if err := meiliClient.SetupIndexes(context.Background()); err != nil {
		log.Fatal().Err(err).Msg("failed to setup meilisearch indexes")
	}

	// Initialize Qdrant client
	qdrantClient, err := qdrant.NewClient(cfg.Qdrant.Host, cfg.Qdrant.Port)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to initialize qdrant client")
	}
	defer qdrantClient.Close()

	// Initialize NATS client for event subscription
	natsConfig := nats.Config{
		URL:       cfg.NATS.URL,
		ClusterID: cfg.NATS.ClusterID,
		ClientID:  "search-service",
	}
	natsClient, err := nats.NewClient(natsConfig)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to initialize nats client")
	}
	defer natsClient.Close()
	// Initialize mock repositories for database-backed functionality
	historyRepo := mock.NewMockSearchHistoryRepository()
	trendingRepo := mock.NewMockTrendingRepository()

	// Create search service
	service := application.NewService(meiliClient, meiliClient, qdrantClient, historyRepo, trendingRepo)

	// Initialize OpenAI embedding client for semantic search
	if cfg.OpenAI.APIKey != "" {
		embeddingClient := searchopenai.NewEmbeddingClient(searchopenai.Config{
			APIKey:         cfg.OpenAI.APIKey,
			EmbeddingModel: cfg.OpenAI.EmbeddingModel,
		})
		service.SetEmbeddingGenerator(embeddingClient)
		log.Info().Msg("OpenAI embedding client initialized for semantic search")
	} else {
		log.Warn().Msg("OpenAI API key not configured, semantic search will return empty results")
	}

	// Start search indexing worker
	worker := application.NewWorker(service, natsClient)
	if err := worker.Start(ctx); err != nil {
		log.Fatal().Err(err).Msg("failed to start search indexing worker")
	}

	// Initialize JWT manager
	jwtConfig := jwt.Config{
		Secret:             cfg.JWT.Secret,
		AccessTokenExpiry:  cfg.JWT.AccessTokenExpiry,
		RefreshTokenExpiry: cfg.JWT.RefreshTokenExpiry,
		Issuer:             "ngasihtau-search-service",
	}
	jwtManager := jwt.NewManager(jwtConfig)

	// Initialize Fiber app
	app := fiber.New(fiber.Config{
		AppName:      "NgasihTau Search Service",
		ErrorHandler: errorHandler,
	})

	// Global middleware
	app.Use(recover.New())
	app.Use(middleware.RequestID())
	corsOrigins := "http://localhost:3000,http://localhost:5173"
	if len(cfg.CORS.AllowedOrigins) > 0 {
		corsOrigins = cfg.CORS.AllowedOrigins[0]
		for _, origin := range cfg.CORS.AllowedOrigins[1:] {
			corsOrigins += "," + origin
		}
	}
	app.Use(cors.New(cors.Config{
		AllowOrigins:     corsOrigins,
		AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders:     "Origin,Content-Type,Accept,Authorization,X-Request-ID",
		AllowCredentials: true,
	}))

	// Initialize handler with service and JWT manager
	handler := http.NewHandler(service, jwtManager)
	handler.RegisterRoutes(app)

	// Initialize health checker
	healthChecker := health.NewChecker("search-service", "1.0.0")
	healthChecker.AddDependency("meilisearch", health.MeilisearchCheck("meilisearch", meiliClient))
	healthChecker.AddDependency("qdrant", health.QdrantCheck("qdrant", qdrantClient))
	healthChecker.AddDependency("nats", health.NATSCheck("nats", natsClient))
	healthChecker.RegisterRoutes(app)

	// Start server
	port := 8085 // Default search service port
	if cfg.SearchSvc.Port > 0 {
		port = cfg.SearchSvc.Port
	}
	addr := fmt.Sprintf(":%d", port)

	go func() {
		log.Info().Str("addr", addr).Msg("HTTP server starting")
		if err := app.Listen(addr); err != nil {
			log.Error().Err(err).Msg("server error")
			cancel()
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info().Msg("shutting down server...")

	// Graceful shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := app.ShutdownWithContext(shutdownCtx); err != nil {
		log.Error().Err(err).Msg("server forced to shutdown")
	}

	log.Info().Msg("server exited")
}

// errorHandler is the global error handler for Fiber.
func errorHandler(c *fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError
	message := "Internal Server Error"

	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
		message = e.Message
	}

	requestID := middleware.GetRequestID(c)

	return c.Status(code).JSON(fiber.Map{
		"success": false,
		"error": fiber.Map{
			"code":    "INTERNAL_ERROR",
			"message": message,
		},
		"meta": fiber.Map{
			"timestamp":  time.Now().UTC().Format(time.RFC3339),
			"request_id": requestID,
		},
	})
}

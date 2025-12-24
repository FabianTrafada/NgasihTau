// Package main is the entry point for the Pod Service.
// Pod Service handles Knowledge Pod CRUD operations, collaboration, and activity feeds.
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
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"ngasihtau/internal/common/config"
	"ngasihtau/internal/common/health"
	"ngasihtau/internal/common/middleware"
	"ngasihtau/internal/pod/application"
	"ngasihtau/internal/pod/infrastructure/postgres"
	podhttp "ngasihtau/internal/pod/interfaces/http"
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
		Str("service", "pod-service").
		Str("env", cfg.App.Env).
		Int("port", cfg.PodService.Port).
		Msg("starting pod service")

	// Create application context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize application
	app, err := initializeApp(ctx, cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to initialize application")
	}

	// Start HTTP server in goroutine
	go func() {
		addr := fmt.Sprintf("%s:%d", cfg.PodService.Host, cfg.PodService.Port)
		if err := app.Start(addr); err != nil {
			log.Error().Err(err).Msg("server error")
			cancel()
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info().Msg("shutting down server...")

	// Graceful shutdown with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := app.Shutdown(shutdownCtx); err != nil {
		log.Error().Err(err).Msg("server forced to shutdown")
	}

	log.Info().Msg("server exited")
}

// App represents the pod service application.
type App struct {
	httpServer    *fiber.App
	db            *pgxpool.Pool
	natsClient    *nats.Client
	healthChecker *health.Checker
}

// initializeApp creates and configures the application with all dependencies.
func initializeApp(ctx context.Context, cfg *config.Config) (*App, error) {
	// Initialize database connection pool
	dbConfig, err := pgxpool.ParseConfig(cfg.PodDB.DSN())
	if err != nil {
		return nil, fmt.Errorf("failed to parse database config: %w", err)
	}

	dbConfig.MaxConns = int32(cfg.PodDB.MaxOpenConns)
	dbConfig.MinConns = int32(cfg.PodDB.MaxIdleConns / 2)
	dbConfig.MaxConnLifetime = cfg.PodDB.ConnMaxLifetime
	dbConfig.MaxConnIdleTime = 5 * time.Minute

	db, err := pgxpool.NewWithConfig(ctx, dbConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Verify database connection
	if err := db.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}
	log.Info().Msg("connected to database")

	// Initialize JWT manager
	jwtManager := jwt.NewManager(jwt.Config{
		Secret:             cfg.JWT.Secret,
		AccessTokenExpiry:  cfg.JWT.AccessTokenExpiry,
		RefreshTokenExpiry: cfg.JWT.RefreshTokenExpiry,
		Issuer:             "ngasihtau",
	})

	// Initialize NATS client for event publishing
	var natsClient *nats.Client
	var eventPublisher application.EventPublisher

	natsConfig := nats.Config{
		URL:       cfg.NATS.URL,
		ClusterID: cfg.NATS.ClusterID,
		ClientID:  "pod-service",
	}
	natsClient, err = nats.NewClient(natsConfig)
	if err != nil {
		log.Warn().Err(err).Msg("failed to connect to NATS, events will not be published")
		eventPublisher = application.NewNoOpEventPublisher()
	} else {
		// Ensure POD stream exists
		if err := natsClient.EnsureStream(ctx, nats.StreamPod, nats.StreamSubjects[nats.StreamPod]); err != nil {
			log.Warn().Err(err).Msg("failed to ensure POD stream")
		}
		publisher := nats.NewPublisher(natsClient)
		eventPublisher = application.NewNATSEventPublisher(publisher)
		log.Info().Msg("connected to NATS for event publishing")
	}

	// Initialize repositories
	podRepo := postgres.NewPodRepository(db)
	collaboratorRepo := postgres.NewCollaboratorRepository(db)
	starRepo := postgres.NewPodStarRepository(db)
	followRepo := postgres.NewPodFollowRepository(db)
	activityRepo := postgres.NewActivityRepository(db)

	// Initialize recommendation repositories
	interactionRepo := postgres.NewInteractionRepository(db)
	userCategoryScoreRepo := postgres.NewUserCategoryScoreRepository(db)
	userTagScoreRepo := postgres.NewUserTagScoreRepository(db)
	podPopularityRepo := postgres.NewPodPopularityRepository(db)
	recommendationRepo := postgres.NewRecommendationRepository(
		db,
		podRepo,
		userCategoryScoreRepo,
		userTagScoreRepo,
		podPopularityRepo,
		interactionRepo,
	)

	// Initialize services
	podService := application.NewPodService(
		podRepo,
		collaboratorRepo,
		starRepo,
		followRepo,
		activityRepo,
		eventPublisher,
	)

	recommendationService := application.NewRecommendationService(
		interactionRepo,
		userCategoryScoreRepo,
		userTagScoreRepo,
		podPopularityRepo,
		recommendationRepo,
		podRepo,
		log.Logger,
	)

	// Initialize HTTP handlers
	handler := podhttp.NewHandler(podService, recommendationService, jwtManager)

	// Initialize Fiber app
	fiberApp := fiber.New(fiber.Config{
		AppName:      "NgasihTau Pod Service",
		ErrorHandler: errorHandler,
	})

	// Global middleware
	fiberApp.Use(recover.New())
	fiberApp.Use(middleware.RequestID())

	// Configure CORS
	corsOrigins := "*"
	if len(cfg.CORS.AllowedOrigins) > 0 {
		corsOrigins = cfg.CORS.AllowedOrigins[0]
		for i := 1; i < len(cfg.CORS.AllowedOrigins); i++ {
			corsOrigins += "," + cfg.CORS.AllowedOrigins[i]
		}
	}
	fiberApp.Use(cors.New(cors.Config{
		AllowOrigins:     corsOrigins,
		AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders:     "Origin,Content-Type,Accept,Authorization,X-Request-ID",
		AllowCredentials: true,
	}))

	// Register routes
	handler.RegisterRoutes(fiberApp)

	// Initialize health checker
	healthChecker := health.NewChecker("pod-service", "1.0.0")
	healthChecker.AddDependency("postgres", health.PostgresCheck("postgres", db))
	if natsClient != nil {
		healthChecker.AddDependency("nats", health.NATSCheck("nats", natsClient))
	}
	healthChecker.RegisterRoutes(fiberApp)

	log.Info().Msg("application initialized")

	return &App{
		httpServer:    fiberApp,
		db:            db,
		natsClient:    natsClient,
		healthChecker: healthChecker,
	}, nil
}

// Start starts the HTTP server.
func (a *App) Start(addr string) error {
	log.Info().Str("addr", addr).Msg("HTTP server starting")
	return a.httpServer.Listen(addr)
}

// Shutdown gracefully shuts down the application.
func (a *App) Shutdown(ctx context.Context) error {
	log.Info().Msg("shutting down application")

	// Shutdown HTTP server
	if err := a.httpServer.ShutdownWithContext(ctx); err != nil {
		log.Error().Err(err).Msg("failed to shutdown HTTP server")
	}

	// Close NATS connection
	if a.natsClient != nil {
		a.natsClient.Close()
		log.Info().Msg("NATS connection closed")
	}

	// Close database pool
	if a.db != nil {
		a.db.Close()
		log.Info().Msg("database connection closed")
	}

	return nil
}

// errorHandler is the global error handler for Fiber.
func errorHandler(c *fiber.Ctx, err error) error {
	// Default to 500 Internal Server Error
	code := fiber.StatusInternalServerError
	message := "Internal Server Error"

	// Check if it's a Fiber error
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

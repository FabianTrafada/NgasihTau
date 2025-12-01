// Package main is the entry point for the Material Service.
// Material Service handles file uploads, versioning, comments, ratings, and bookmarks.
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
	"ngasihtau/internal/common/middleware"
	"ngasihtau/internal/material/application"
	minioclient "ngasihtau/internal/material/infrastructure/minio"
	"ngasihtau/internal/material/infrastructure/postgres"
	materialhttp "ngasihtau/internal/material/interfaces/http"
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
		Str("service", "material-service").
		Str("env", cfg.App.Env).
		Int("port", cfg.MatService.Port).
		Msg("starting material service")

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
		addr := fmt.Sprintf("%s:%d", cfg.MatService.Host, cfg.MatService.Port)
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

// App represents the material service application.
type App struct {
	httpServer *fiber.App
	db         *pgxpool.Pool
	natsClient *nats.Client
}

// initializeApp creates and configures the application with all dependencies.
func initializeApp(ctx context.Context, cfg *config.Config) (*App, error) {
	// Initialize database connection pool
	dbConfig, err := pgxpool.ParseConfig(cfg.MaterialDB.DSN())
	if err != nil {
		return nil, fmt.Errorf("failed to parse database config: %w", err)
	}

	dbConfig.MaxConns = int32(cfg.MaterialDB.MaxOpenConns)
	dbConfig.MinConns = int32(cfg.MaterialDB.MaxIdleConns / 2)
	dbConfig.MaxConnLifetime = cfg.MaterialDB.ConnMaxLifetime
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

	// Initialize MinIO client
	minioClient, err := minioclient.NewClient(minioclient.Config{
		Endpoint:        cfg.MinIO.Endpoint,
		AccessKey:       cfg.MinIO.AccessKey,
		SecretKey:       cfg.MinIO.SecretKey,
		UseSSL:          cfg.MinIO.UseSSL,
		BucketMaterials: cfg.MinIO.BucketMaterials,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create MinIO client: %w", err)
	}

	// Ensure materials bucket exists
	if err := minioClient.EnsureBucket(ctx); err != nil {
		log.Warn().Err(err).Msg("failed to ensure materials bucket exists")
	}
	log.Info().Msg("connected to MinIO")

	// Initialize NATS client for event publishing
	var natsClient *nats.Client
	var eventPublisher nats.EventPublisher
	if cfg.NATS.URL != "" {
		natsClient, err = nats.NewClient(nats.Config{
			URL:       cfg.NATS.URL,
			ClusterID: cfg.NATS.ClusterID,
			ClientID:  "material-service",
		})
		if err != nil {
			log.Warn().Err(err).Msg("failed to connect to NATS, event publishing disabled")
			eventPublisher = nats.NewNoOpPublisher()
		} else {
			log.Info().Msg("connected to NATS")
			eventPublisher = nats.NewPublisher(natsClient)

			// Ensure MATERIAL stream exists for event publishing
			if err := natsClient.EnsureStream(ctx, nats.StreamMaterial, nats.StreamSubjects[nats.StreamMaterial]); err != nil {
				log.Warn().Err(err).Msg("failed to ensure MATERIAL stream exists")
			}
		}
	} else {
		log.Info().Msg("NATS not configured, event publishing disabled")
		eventPublisher = nats.NewNoOpPublisher()
	}

	// Initialize JWT manager
	jwtManager := jwt.NewManager(jwt.Config{
		Secret:             cfg.JWT.Secret,
		AccessTokenExpiry:  cfg.JWT.AccessTokenExpiry,
		RefreshTokenExpiry: cfg.JWT.RefreshTokenExpiry,
		Issuer:             "ngasihtau",
	})

	// Initialize repositories
	materialRepo := postgres.NewMaterialRepository(db)
	versionRepo := postgres.NewMaterialVersionRepository(db)
	commentRepo := postgres.NewCommentRepository(db)
	ratingRepo := postgres.NewRatingRepository(db)
	bookmarkRepo := postgres.NewBookmarkRepository(db)

	// Initialize services
	materialService := application.NewService(
		materialRepo,
		versionRepo,
		commentRepo,
		ratingRepo,
		bookmarkRepo,
		minioClient,
		eventPublisher,
	)

	// Initialize HTTP handlers
	handler := materialhttp.NewHandler(materialService, jwtManager)

	// Initialize Fiber app
	fiberApp := fiber.New(fiber.Config{
		AppName:      "NgasihTau Material Service",
		ErrorHandler: errorHandler,
		BodyLimit:    100 * 1024 * 1024, // 100MB for file metadata
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

	// Health check endpoint
	fiberApp.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "healthy",
			"service": "material-service",
		})
	})

	log.Info().Msg("application initialized")

	return &App{
		httpServer: fiberApp,
		db:         db,
		natsClient: natsClient,
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

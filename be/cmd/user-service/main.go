// Package main is the entry point for the User Service.
// User Service handles authentication, authorization, user profiles, and social features.
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
	"github.com/gofiber/swagger"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	_ "ngasihtau/api/swagger" // Swagger generated docs
	"ngasihtau/internal/common/config"
	"ngasihtau/internal/common/health"
	"ngasihtau/internal/common/middleware"
	"ngasihtau/internal/user/application"
	"ngasihtau/internal/user/infrastructure/postgres"
	userhttp "ngasihtau/internal/user/interfaces/http"
	"ngasihtau/pkg/jwt"
	natspkg "ngasihtau/pkg/nats"
	"ngasihtau/pkg/oauth"
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
		Str("service", "user-service").
		Str("env", cfg.App.Env).
		Int("port", cfg.UserService.Port).
		Msg("starting user service")

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
		addr := fmt.Sprintf("%s:%d", cfg.UserService.Host, cfg.UserService.Port)
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

// App represents the user service application.
type App struct {
	httpServer    *fiber.App
	db            *pgxpool.Pool
	aiDB          *pgxpool.Pool
	podDB         *pgxpool.Pool
	natsClient    *natspkg.Client
	healthChecker *health.Checker
}

// initializeApp creates and configures the application with all dependencies.
func initializeApp(ctx context.Context, cfg *config.Config) (*App, error) {
	// Initialize database connection pool
	dbConfig, err := pgxpool.ParseConfig(cfg.UserDB.DSN())
	if err != nil {
		return nil, fmt.Errorf("failed to parse database config: %w", err)
	}

	dbConfig.MaxConns = int32(cfg.UserDB.MaxOpenConns)
	dbConfig.MinConns = int32(cfg.UserDB.MaxIdleConns / 2)
	dbConfig.MaxConnLifetime = cfg.UserDB.ConnMaxLifetime
	dbConfig.MaxConnIdleTime = 5 * time.Minute // Default idle time

	db, err := pgxpool.NewWithConfig(ctx, dbConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Verify database connection
	if err := db.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}
	log.Info().Msg("connected to user database")

	// Initialize optional database connections for behavior tracking
	var aiDB *pgxpool.Pool
	var podDB *pgxpool.Pool

	// Try to connect to AI database (optional - for chat behavior data)
	if cfg.AIDB.Host != "" {
		aiDBConfig, err := pgxpool.ParseConfig(cfg.AIDB.DSN())
		if err == nil {
			aiDBConfig.MaxConns = 5
			aiDBConfig.MinConns = 1
			aiDB, err = pgxpool.NewWithConfig(ctx, aiDBConfig)
			if err != nil {
				log.Warn().Err(err).Msg("failed to connect to AI database - chat behavior data will be unavailable")
			} else {
				if err := aiDB.Ping(ctx); err != nil {
					log.Warn().Err(err).Msg("failed to ping AI database - chat behavior data will be unavailable")
					aiDB = nil
				} else {
					log.Info().Msg("connected to AI database for behavior tracking")
				}
			}
		}
	}

	// Try to connect to Pod database (optional - for material behavior data)
	if cfg.PodDB.Host != "" {
		podDBConfig, err := pgxpool.ParseConfig(cfg.PodDB.DSN())
		if err == nil {
			podDBConfig.MaxConns = 5
			podDBConfig.MinConns = 1
			podDB, err = pgxpool.NewWithConfig(ctx, podDBConfig)
			if err != nil {
				log.Warn().Err(err).Msg("failed to connect to Pod database - material behavior data will be unavailable")
			} else {
				if err := podDB.Ping(ctx); err != nil {
					log.Warn().Err(err).Msg("failed to ping Pod database - material behavior data will be unavailable")
					podDB = nil
				} else {
					log.Info().Msg("connected to Pod database for behavior tracking")
				}
			}
		}
	}

	// Initialize JWT manager
	jwtManager := jwt.NewManager(jwt.Config{
		Secret:             cfg.JWT.Secret,
		AccessTokenExpiry:  cfg.JWT.AccessTokenExpiry,
		RefreshTokenExpiry: cfg.JWT.RefreshTokenExpiry,
		Issuer:             "ngasihtau",
	})

	// Initialize repositories
	userRepo := postgres.NewUserRepository(db)
	oauthRepo := postgres.NewOAuthRepository(db)
	refreshTokenRepo := postgres.NewRefreshTokenRepository(db)
	backupCodeRepo := postgres.NewBackupCodeRepository(db)
	followRepo := postgres.NewFollowRepository(db)
	verificationTokenRepo := postgres.NewVerificationTokenRepository(db)
	teacherVerificationRepo := postgres.NewTeacherVerificationRepository(db)
	predefinedInterestRepo := postgres.NewPredefinedInterestRepository(db)
	userLearningInterestRepo := postgres.NewUserLearningInterestRepository(db)
	storageRepo := postgres.NewStorageRepository(db)

	// Initialize Google OAuth client (optional - may not be configured)
	var googleClient *oauth.GoogleClient
	if cfg.OAuth.Google.ClientID != "" && cfg.OAuth.Google.ClientSecret != "" {
		googleClient = oauth.NewGoogleClient(oauth.GoogleConfig{
			ClientID:     cfg.OAuth.Google.ClientID,
			ClientSecret: cfg.OAuth.Google.ClientSecret,
			RedirectURL:  cfg.OAuth.Google.RedirectURL,
		})
		log.Info().Msg("Google OAuth client initialized")
	} else {
		log.Warn().Msg("Google OAuth not configured - Google login will be disabled")
	}

	// Initialize NATS client (optional - may not be configured)
	var natsClient *natspkg.Client
	var eventPublisher natspkg.EventPublisher
	if cfg.NATS.URL != "" {
		var err error
		natsClient, err = natspkg.NewClient(natspkg.Config{
			URL:       cfg.NATS.URL,
			ClusterID: cfg.NATS.ClusterID,
			ClientID:  cfg.NATS.ClientID + "-user-service",
		})
		if err != nil {
			log.Warn().Err(err).Msg("failed to connect to NATS - event publishing will be disabled")
			eventPublisher = natspkg.NewNoOpPublisher()
		} else {
			log.Info().Msg("NATS client initialized")

			// Ensure required streams exist for publishing
			for streamName, subjects := range natspkg.StreamSubjects {
				if err := natsClient.EnsureStream(ctx, streamName, subjects); err != nil {
					log.Warn().Err(err).Str("stream", streamName).Msg("failed to ensure stream")
				} else {
					log.Debug().Str("stream", streamName).Msg("stream ensured")
				}
			}

			eventPublisher = natspkg.NewPublisher(natsClient)
		}
	} else {
		log.Warn().Msg("NATS not configured - event publishing will be disabled")
		eventPublisher = natspkg.NewNoOpPublisher()
	}

	// Initialize services
	userService := application.NewUserService(
		userRepo,
		oauthRepo,
		refreshTokenRepo,
		backupCodeRepo,
		followRepo,
		verificationTokenRepo,
		teacherVerificationRepo,
		jwtManager,
		googleClient,
		eventPublisher,
	)

	// Initialize learning interest service
	interestService := application.NewLearningInterestService(
		predefinedInterestRepo,
		userLearningInterestRepo,
		userRepo,
	)

	// Initialize storage service
	storageService := application.NewStorageService(
		userRepo,
		storageRepo,
		cfg.Storage,
	)

	// Initialize behavior repository and service
	behaviorRepo := postgres.NewBehaviorRepository(db, aiDB, podDB)
	behaviorService := application.NewBehaviorService(behaviorRepo)

	// Initialize HTTP handlers
	handler := userhttp.NewHandlerWithBehavior(userService, interestService, storageService, behaviorService, jwtManager)

	// Initialize Fiber app
	fiberApp := fiber.New(fiber.Config{
		AppName:      "NgasihTau User Service",
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

	// Register Swagger UI routes
	// Swagger UI at /api/docs
	fiberApp.Get("/api/docs/*", swagger.HandlerDefault)
	// OpenAPI spec at /api/openapi.json
	fiberApp.Get("/api/openapi.json", func(c *fiber.Ctx) error {
		return c.SendFile("./api/swagger/swagger.json")
	})
	// OpenAPI spec at /api/openapi.yaml
	fiberApp.Get("/api/openapi.yaml", func(c *fiber.Ctx) error {
		return c.SendFile("./api/swagger/swagger.yaml")
	})

	// Initialize health checker
	healthChecker := health.NewChecker("user-service", "1.0.0")
	healthChecker.AddDependency("postgres", health.PostgresCheck("postgres", db))
	if natsClient != nil {
		healthChecker.AddDependency("nats", health.NATSCheck("nats", natsClient))
	}
	healthChecker.RegisterRoutes(fiberApp)

	log.Info().Msg("application initialized")

	return &App{
		httpServer:    fiberApp,
		db:            db,
		aiDB:          aiDB,
		podDB:         podDB,
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
		log.Info().Msg("user database connection closed")
	}

	// Close AI database pool
	if a.aiDB != nil {
		a.aiDB.Close()
		log.Info().Msg("AI database connection closed")
	}

	// Close Pod database pool
	if a.podDB != nil {
		a.podDB.Close()
		log.Info().Msg("Pod database connection closed")
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

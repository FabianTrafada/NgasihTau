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
	"ngasihtau/internal/notification/application"
	"ngasihtau/internal/notification/infrastructure/postgres"
	notificationhttp "ngasihtau/internal/notification/interfaces/http"
	"ngasihtau/pkg/email"
	"ngasihtau/pkg/jwt"
	natspkg "ngasihtau/pkg/nats"
)

func main() {
	zerolog.TimeFieldFormat = time.RFC3339
	if os.Getenv("APP_ENV") == "development" {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}

	configPath := os.Getenv("CONFIG_PATH")
	cfg, err := config.Load(configPath)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to load configuration")
	}

	level, err := zerolog.ParseLevel(cfg.App.LogLevel)
	if err != nil {
		level = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(level)

	log.Info().
		Str("service", "notification-service").
		Str("env", cfg.App.Env).
		Int("port", cfg.NotifSvc.Port).
		Msg("starting notification service")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	app, err := initializeApp(ctx, cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to initialize application")
	}

	go func() {
		addr := fmt.Sprintf("%s:%d", cfg.NotifSvc.Host, cfg.NotifSvc.Port)
		if err := app.Start(addr); err != nil {
			log.Error().Err(err).Msg("server error")
			cancel()
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info().Msg("shutting down server...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := app.Shutdown(shutdownCtx); err != nil {
		log.Error().Err(err).Msg("server forced to shutdown")
	}

	log.Info().Msg("server exited")
}

type App struct {
	httpServer    *fiber.App
	db            *pgxpool.Pool
	natsClient    *natspkg.Client
	emailWorker   *application.EmailWorker
	healthChecker *health.Checker
}

func initializeApp(ctx context.Context, cfg *config.Config) (*App, error) {
	dbConfig, err := pgxpool.ParseConfig(cfg.NotifDB.DSN())
	if err != nil {
		return nil, fmt.Errorf("failed to parse database config: %w", err)
	}

	dbConfig.MaxConns = int32(cfg.NotifDB.MaxOpenConns)
	dbConfig.MinConns = int32(cfg.NotifDB.MaxIdleConns / 2)
	dbConfig.MaxConnLifetime = cfg.NotifDB.ConnMaxLifetime
	dbConfig.MaxConnIdleTime = 5 * time.Minute

	db, err := pgxpool.NewWithConfig(ctx, dbConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	if err := db.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}
	log.Info().Msg("connected to database")

	jwtManager := jwt.NewManager(jwt.Config{
		Secret:             cfg.JWT.Secret,
		AccessTokenExpiry:  cfg.JWT.AccessTokenExpiry,
		RefreshTokenExpiry: cfg.JWT.RefreshTokenExpiry,
		Issuer:             "ngasihtau",
	})

	notificationRepo := postgres.NewNotificationRepository(db)
	preferenceRepo := postgres.NewNotificationPreferenceRepository(db)

	var natsClient *natspkg.Client
	if cfg.NATS.URL != "" {
		var err error
		natsClient, err = natspkg.NewClient(natspkg.Config{
			URL:       cfg.NATS.URL,
			ClusterID: cfg.NATS.ClusterID,
			ClientID:  cfg.NATS.ClientID + "-notification-service",
		})
		if err != nil {
			log.Warn().Err(err).Msg("failed to connect to NATS - event consuming will be disabled")
		} else {
			log.Info().Err(err).Msg("NATS client initialized")
		}
	} else {
		log.Warn().Msg("NATS not configured - event consuming will be disabled")
	}

	var emailWorker *application.EmailWorker
	if natsClient != nil {
		emailProvider, err := email.NewProvider(ctx, cfg.Email)
		if err != nil {
			log.Warn().Err(err).Msg("failed to initialize email provider - email sending will be disabled")
		} else {
			emailWorkerConfig := application.DefaultEmailWorkerConfig()
			// Use configured frontend URL if available
			if cfg.App.FrontendURL != "" {
				emailWorkerConfig.AppUrl = cfg.App.FrontendURL
			}

			emailWorker = application.NewEmailWorker(natsClient, emailProvider, emailWorkerConfig)
			if err := emailWorker.Start(ctx); err != nil {
				log.Error().Err(err).Msg("failed to start email worker")
			} else {
				log.Info().Msg("email worker started")
			}
		}
	}

	notificationService := application.NewNotificationService(
		notificationRepo,
		preferenceRepo,
	)

	handler := notificationhttp.NewHandler(notificationService, jwtManager)

	fiberApp := fiber.New(fiber.Config{
		AppName:      "NgasihTau - Notification Service",
		ErrorHandler: errorHandler,
	})

	fiberApp.Use(recover.New())
	fiberApp.Use(middleware.RequestID())

	// Configure CORS from config
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

	handler.RegisterRoutes(fiberApp)

	// Initialize health checker
	healthChecker := health.NewChecker("notification-service", "1.0.0")
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
		emailWorker:   emailWorker,
		healthChecker: healthChecker,
	}, nil
}

func (a *App) Start(addr string) error {
	log.Info().Str("addr", addr).Msg("HTTP server starting")
	return a.httpServer.Listen(addr)
}

func (a *App) Shutdown(ctx context.Context) error {
	log.Info().Msg("shutting down service")

	if err := a.httpServer.ShutdownWithContext(ctx); err != nil {
		log.Error().Err(err).Msg("failed to shutdown HTTP server")
	}

	if a.emailWorker != nil {
		a.emailWorker.Stop()
		log.Info().Msg("email worker stoppedx")
	}

	if a.natsClient != nil {
		a.natsClient.Close()
		log.Info().Msg("NATS connection closed")
	}

	if a.db != nil {
		a.db.Close()
		log.Info().Msg("database connection closed")
	}

	return nil
}

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

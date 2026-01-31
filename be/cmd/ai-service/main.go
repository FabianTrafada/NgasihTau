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
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"ngasihtau/internal/ai/application"
	"ngasihtau/internal/ai/infrastructure/gemini"
	"ngasihtau/internal/ai/infrastructure/openai"
	"ngasihtau/internal/ai/infrastructure/postgres"
	"ngasihtau/internal/ai/infrastructure/qdrant"
	aihttp "ngasihtau/internal/ai/interfaces/http"
	"ngasihtau/internal/common/config"
	"ngasihtau/internal/common/health"
	"ngasihtau/internal/common/middleware"
	userapplication "ngasihtau/internal/user/application"
	userpostgres "ngasihtau/internal/user/infrastructure/postgres"
	userredis "ngasihtau/internal/user/infrastructure/redis"
	"ngasihtau/pkg/jwt"
	natspkg "ngasihtau/pkg/nats"
)

type App struct {
	httpServer    *fiber.App
	db            *pgxpool.Pool
	userDB        *pgxpool.Pool
	redisClient   *redis.Client
	natsClient    *natspkg.Client
	qdrantClient  *qdrant.Client
	healthChecker *health.Checker
}

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
		Str("service", "ai-service").
		Str("env", cfg.App.Env).
		Int("port", cfg.AISvc.Port).
		Msg("starting ai service")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	app, err := initializeApp(ctx, cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to initialize application")
	}

	go func() {
		addr := fmt.Sprintf("%s:%d", cfg.AISvc.Host, cfg.AISvc.Port)
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

func initializeApp(ctx context.Context, cfg *config.Config) (*App, error) {
	dbConfig, err := pgxpool.ParseConfig(cfg.AIDB.DSN())
	if err != nil {
		return nil, fmt.Errorf("failed to parse database config: %w", err)
	}

	dbConfig.MaxConns = int32(cfg.AIDB.MaxOpenConns)
	dbConfig.MinConns = int32(cfg.AIDB.MaxIdleConns / 2)
	dbConfig.MaxConnLifetime = cfg.AIDB.ConnMaxLifetime
	dbConfig.MaxConnIdleTime = 5 * time.Minute

	db, err := pgxpool.NewWithConfig(ctx, dbConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	if err := db.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}
	log.Info().Msg("connected to AI database")

	// Connect to User database for AI limit checks
	userDBConfig, err := pgxpool.ParseConfig(cfg.UserDB.DSN())
	if err != nil {
		return nil, fmt.Errorf("failed to parse user database config: %w", err)
	}

	userDBConfig.MaxConns = int32(cfg.UserDB.MaxOpenConns)
	userDBConfig.MinConns = int32(cfg.UserDB.MaxIdleConns / 2)
	userDBConfig.MaxConnLifetime = cfg.UserDB.ConnMaxLifetime
	userDBConfig.MaxConnIdleTime = 5 * time.Minute

	userDB, err := pgxpool.NewWithConfig(ctx, userDBConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to user database: %w", err)
	}

	if err := userDB.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping user database: %w", err)
	}
	log.Info().Msg("connected to User database")

	// Connect to Redis for AI usage tracking
	redisClient := redis.NewClient(&redis.Options{
		Addr:            cfg.Redis.Addr(),
		Password:        cfg.Redis.Password,
		DB:              cfg.Redis.DB,
		PoolSize:        cfg.Redis.PoolSize,
		MinIdleConns:    cfg.Redis.MinIdleConns,
		PoolTimeout:     cfg.Redis.PoolTimeout,
		ConnMaxIdleTime: cfg.Redis.IdleTimeout,
	})

	if err := redisClient.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}
	log.Info().Msg("connected to Redis")

	qdrantClient, err := qdrant.NewClient(qdrant.Config{
		Host:           cfg.Qdrant.Host,
		Port:           cfg.Qdrant.Port,
		CollectionName: cfg.Qdrant.Collection,
		VectorSize:     1536,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Qdrant client: %w", err)
	}

	if err := qdrantClient.EnsureCollection(ctx); err != nil {
		log.Warn().Err(err).Msg("failed to ensure Qdrant collection exists")
	}
	log.Info().Msg("connected to Qdrant")

	// Initialize AI clients based on provider config
	var embeddingClient application.EmbeddingClient
	var chatClient application.ChatClient

	switch cfg.AI.Provider {
	case "gemini":
		geminiClient, err := gemini.NewClient(ctx, gemini.Config{
			APIKey:         cfg.Gemini.APIKey,
			ChatModel:      cfg.Gemini.ChatModel,
			EmbeddingModel: cfg.Gemini.EmbeddingModel,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to create Gemini client: %w", err)
		}
		embeddingClient = geminiClient
		chatClient = geminiClient
		log.Info().
			Str("chat_model", cfg.Gemini.ChatModel).
			Str("embedding_model", cfg.Gemini.EmbeddingModel).
			Msg("using Gemini as AI provider")
	default:
		// Default to OpenAI
		openaiClient := openai.NewClient(openai.Config{
			APIKey:         cfg.OpenAI.APIKey,
			EmbeddingModel: cfg.OpenAI.EmbeddingModel,
			ChatModel:      cfg.OpenAI.ChatModel,
		})
		embeddingClient = openaiClient
		chatClient = openaiClient
		log.Info().
			Str("chat_model", cfg.OpenAI.ChatModel).
			Str("embedding_model", cfg.OpenAI.EmbeddingModel).
			Msg("using OpenAI as AI provider")
	}

	var natsClient *natspkg.Client
	if cfg.NATS.URL != "" {
		natsClient, err = natspkg.NewClient(natspkg.Config{
			URL:       cfg.NATS.URL,
			ClusterID: cfg.NATS.ClusterID,
			ClientID:  "ai-service",
		})
		if err != nil {
			log.Warn().Err(err).Msg("failed to connect to NATS, event consuming disabled")
		} else {
			log.Info().Msg("connected to NATS")
		}
	} else {
		log.Info().Msg("NATS not configured, event consuming disabled")
	}

	jwtManager := jwt.NewManager(jwt.Config{
		Secret:             cfg.JWT.Secret,
		AccessTokenExpiry:  cfg.JWT.AccessTokenExpiry,
		RefreshTokenExpiry: cfg.JWT.RefreshTokenExpiry,
		Issuer:             "ngasihtau",
	})

	chatSessionRepo := postgres.NewChatSessionRepository(db)
	chatMessageRepo := postgres.NewChatMessageRepository(db)

	// Initialize User repositories for AI limit checking
	userRepo := userpostgres.NewUserRepository(userDB)
	aiUsageRepo := userredis.NewAIUsageRepository(redisClient)

	// Create AIService for limit checking
	aiLimitChecker := userapplication.NewAIService(userRepo, aiUsageRepo, cfg.AILimit)

	aiService := application.NewService(
		chatSessionRepo,
		chatMessageRepo,
		qdrantClient,
		embeddingClient,
		chatClient,
		cfg.FileProcSvc.URL,
		aiLimitChecker,
	)

	worker := application.NewWorker(aiService, natsClient, cfg.FileProcSvc.URL)
	go func() {
		if err := worker.Start(ctx); err != nil {
			log.Error().Err(err).Msg("failed to start material processing worker")
		}
	}()

	handler := aihttp.NewHandler(aiService, jwtManager)

	fiberApp := fiber.New(fiber.Config{
		AppName:      "NgasihTau AI Service",
		ErrorHandler: errorHandler,
	})

	fiberApp.Use(recover.New())
	fiberApp.Use(middleware.RequestID())

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
	healthChecker := health.NewChecker("ai-service", "1.0.0")
	healthChecker.AddDependency("postgres", health.PostgresCheck("postgres", db))
	healthChecker.AddDependency("user_postgres", health.PostgresCheck("user_postgres", userDB))
	healthChecker.AddDependency("redis", health.CustomCheck("redis", func(ctx context.Context) error {
		return redisClient.Ping(ctx).Err()
	}))
	healthChecker.AddDependency("qdrant", health.QdrantCheck("qdrant", qdrantClient))
	if natsClient != nil {
		healthChecker.AddDependency("nats", health.NATSCheck("nats", natsClient))
	}
	healthChecker.RegisterRoutes(fiberApp)

	log.Info().Msg("application initialized")

	return &App{
		httpServer:    fiberApp,
		db:            db,
		userDB:        userDB,
		redisClient:   redisClient,
		natsClient:    natsClient,
		qdrantClient:  qdrantClient,
		healthChecker: healthChecker,
	}, nil
}

func (a *App) Start(addr string) error {
	log.Info().Str("addr", addr).Msg("HTTP server starting")
	return a.httpServer.Listen(addr)
}

func (a *App) Shutdown(ctx context.Context) error {
	log.Info().Msg("shutting down application")

	if err := a.httpServer.ShutdownWithContext(ctx); err != nil {
		log.Error().Err(err).Msg("failed to shutdown HTTP server")
	}

	if a.natsClient != nil {
		a.natsClient.Close()
		log.Info().Msg("NATS connection closed")
	}

	if a.redisClient != nil {
		if err := a.redisClient.Close(); err != nil {
			log.Error().Err(err).Msg("failed to close Redis connection")
		} else {
			log.Info().Msg("Redis connection closed")
		}
	}

	if a.userDB != nil {
		a.userDB.Close()
		log.Info().Msg("User database connection closed")
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

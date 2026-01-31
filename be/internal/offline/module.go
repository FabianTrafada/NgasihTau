// Package offline provides the Offline Material Service module.
// This module handles device management, license management, encrypted material downloads,
// and background encryption jobs for offline access to learning materials.
package offline

import (
	"context"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"

	"ngasihtau/internal/common/health"
	"ngasihtau/internal/offline/application"
	offlinematerial "ngasihtau/internal/offline/infrastructure/material"
	"ngasihtau/internal/offline/infrastructure/metrics"
	"ngasihtau/internal/offline/infrastructure/postgres"
	offlineredis "ngasihtau/internal/offline/infrastructure/redis"
	offlinehttp "ngasihtau/internal/offline/interfaces/http"
	"ngasihtau/pkg/jwt"
	"ngasihtau/pkg/nats"
)

// Module represents the Offline Material Service module.
// It encapsulates all offline-related services, repositories, and handlers.
type Module struct {
	// Services
	DeviceService   application.DeviceService
	LicenseService  application.LicenseService
	SecurityService application.SecurityService
	RateLimiter     application.RateLimiter

	// Infrastructure
	Cache   *offlineredis.OfflineCache
	Metrics *metrics.Metrics

	// HTTP Handler
	Handler *offlinehttp.Handler

	// Health Checker
	HealthChecker *offlinehttp.HealthHandler
}

// Config holds configuration for the Offline module.
type Config struct {
	// Database connection pool for offline tables
	DB *pgxpool.Pool

	// Material database connection pool (for material access verification)
	// Required if MaterialAccessChecker is not provided
	MaterialDB *pgxpool.Pool

	// Pod database connection pool (for pod access verification)
	// Required if MaterialAccessChecker is not provided
	PodDB *pgxpool.Pool

	// Redis client for caching
	RedisClient *redis.Client

	// NATS event publisher
	EventPublisher nats.EventPublisher

	// JWT manager for authentication
	JWTManager *jwt.Manager

	// Material access checker (from material service)
	// If not provided, will be created using MaterialDB and PodDB
	MaterialAccessChecker application.LicenseMaterialAccessChecker

	// Signing secret for request signature validation
	SigningSecret string
}

// NewModule creates and initializes a new Offline module.
func NewModule(cfg Config) (*Module, error) {
	if cfg.DB == nil {
		return nil, fmt.Errorf("database connection is required")
	}
	if cfg.JWTManager == nil {
		return nil, fmt.Errorf("JWT manager is required")
	}

	// Initialize repositories
	deviceRepo := postgres.NewDeviceRepository(cfg.DB)
	licenseRepo := postgres.NewLicenseRepository(cfg.DB)
	cekRepo := postgres.NewCEKRepository(cfg.DB)
	auditRepo := postgres.NewAuditLogRepository(cfg.DB)

	// Initialize cache (optional)
	var cache *offlineredis.OfflineCache
	if cfg.RedisClient != nil {
		cache = offlineredis.NewOfflineCache(cfg.RedisClient)
	}

	// Initialize event publisher (use no-op if not provided)
	eventPublisher := cfg.EventPublisher
	if eventPublisher == nil {
		eventPublisher = nats.NewNoOpPublisher()
	}

	// Initialize material access checker
	// If not provided, create one using MaterialDB and PodDB
	// Implements Requirement 6.2: Integrate with material-service to verify material access permissions
	materialAccessChecker := cfg.MaterialAccessChecker
	if materialAccessChecker == nil {
		if cfg.MaterialDB == nil || cfg.PodDB == nil {
			log.Warn().Msg("material access checker not configured - license issuance will skip access verification")
		} else {
			materialAccessChecker = offlinematerial.NewAccessChecker(cfg.MaterialDB, cfg.PodDB)
			log.Info().Msg("material access checker initialized with database connections")
		}
	}

	// Initialize services
	deviceService := application.NewDeviceService(
		deviceRepo,
		licenseRepo,
		cekRepo,
		eventPublisher,
	)

	licenseService := application.NewLicenseService(
		licenseRepo,
		deviceRepo,
		materialAccessChecker,
		eventPublisher,
	)

	// Initialize rate limiter (requires Redis)
	var rateLimiter application.RateLimiter
	if cfg.RedisClient != nil {
		rateLimiter = application.NewRateLimiter(cfg.RedisClient)
	}

	// Initialize security service (requires Redis for replay protection)
	var securityService application.SecurityService
	if cfg.RedisClient != nil {
		securityService = application.NewSecurityService(
			cfg.RedisClient,
			auditRepo,
			eventPublisher,
			cfg.SigningSecret,
		)
	}

	// Initialize metrics
	offlineMetrics := metrics.NewMetrics()

	// Initialize HTTP handler
	handler := offlinehttp.NewHandlerWithDeviceAndLicense(
		deviceService,
		licenseService,
		cfg.JWTManager,
	)

	// Initialize health handler with dependencies
	healthDeps := &offlinehttp.HealthDependencies{
		PostgresPool: cfg.DB,
		RedisClient:  cfg.RedisClient,
	}
	healthHandler := offlinehttp.NewHealthHandler("offline-module", "1.0.0", healthDeps)

	log.Info().Msg("offline module initialized")

	return &Module{
		DeviceService:   deviceService,
		LicenseService:  licenseService,
		SecurityService: securityService,
		RateLimiter:     rateLimiter,
		Cache:           cache,
		Metrics:         offlineMetrics,
		Handler:         handler,
		HealthChecker:   healthHandler,
	}, nil
}

// RegisterRoutes registers all offline module routes on the Fiber app.
func (m *Module) RegisterRoutes(app *fiber.App) {
	m.Handler.RegisterRoutes(app)
	log.Info().Msg("offline module routes registered")
}

// RegisterHealthChecks adds offline module health checks to the health checker.
func (m *Module) RegisterHealthChecks(checker *health.Checker, db *pgxpool.Pool) {
	// Add database health check
	checker.AddDependency("offline-postgres", health.PostgresCheck("offline-postgres", db))

	// Add Redis health check if cache is available
	if m.Cache != nil {
		checker.AddDependency("offline-redis", func(ctx context.Context) health.DependencyStatus {
			if err := m.Cache.Ping(ctx); err != nil {
				return health.DependencyStatus{
					Name:   "offline-redis",
					Status: health.StatusUnhealthy,
					Error:  err.Error(),
				}
			}
			return health.DependencyStatus{
				Name:   "offline-redis",
				Status: health.StatusHealthy,
			}
		})
	}

	log.Info().Msg("offline module health checks registered")
}

// Close gracefully shuts down the offline module.
func (m *Module) Close() error {
	if m.Cache != nil {
		if err := m.Cache.Close(); err != nil {
			log.Error().Err(err).Msg("failed to close offline cache")
			return err
		}
	}
	log.Info().Msg("offline module closed")
	return nil
}

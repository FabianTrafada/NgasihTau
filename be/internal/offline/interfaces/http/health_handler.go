// Package http provides HTTP handlers for the Offline Material Service API.
package http

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"ngasihtau/internal/common/health"
)

// HealthHandler handles health check endpoints for the Offline Material Service.
type HealthHandler struct {
	checker *health.Checker
}

// HealthDependencies contains dependencies for health checks.
type HealthDependencies struct {
	PostgresPool *pgxpool.Pool
	RedisClient  *redis.Client
	MinIOClient  health.MinIOChecker
	NATSClient   health.NATSChecker
}

// NewHealthHandler creates a new HealthHandler with the given dependencies.
func NewHealthHandler(serviceName, version string, deps *HealthDependencies) *HealthHandler {
	checker := health.NewChecker(serviceName, version)

	// Register dependency checks
	if deps != nil {
		if deps.PostgresPool != nil {
			checker.AddDependency("postgres", health.PostgresCheck("postgres", deps.PostgresPool))
		}
		if deps.RedisClient != nil {
			checker.AddDependency("redis", RedisCheck("redis", deps.RedisClient))
		}
		if deps.MinIOClient != nil {
			checker.AddDependency("minio", health.MinIOCheck("minio", deps.MinIOClient))
		}
		if deps.NATSClient != nil {
			checker.AddDependency("nats", health.NATSCheck("nats", deps.NATSClient))
		}
	}

	return &HealthHandler{
		checker: checker,
	}
}

// RegisterRoutes registers health check routes.
// Routes:
//   - GET /health/live - Liveness probe (always returns 200 if service is running)
//   - GET /health/ready - Readiness probe (checks all dependencies)
//   - GET /health/full - Full health check with detailed dependency status
func (h *HealthHandler) RegisterRoutes(app *fiber.App) {
	healthGroup := app.Group("/health")

	// Liveness probe - simple check that service is running
	healthGroup.Get("/live", h.LivenessProbe)

	// Readiness probe - checks if service is ready to accept traffic
	healthGroup.Get("/ready", h.ReadinessProbe)

	// Full health check - detailed status of all dependencies
	healthGroup.Get("/full", h.FullHealthCheck)

	// Alias for backward compatibility
	app.Get("/health", h.FullHealthCheck)
}

// LivenessProbe handles GET /health/live.
// Returns 200 OK if the service is running.
// @Summary Liveness probe
// @Description Check if the service is alive
// @Tags Health
// @Produce json
// @Success 200 {object} health.LiveResponse
// @Router /health/live [get]
func (h *HealthHandler) LivenessProbe(c *fiber.Ctx) error {
	return c.JSON(h.checker.Live())
}

// ReadinessProbe handles GET /health/ready.
// Returns 200 OK if all dependencies are healthy, 503 otherwise.
// @Summary Readiness probe
// @Description Check if the service is ready to accept traffic
// @Tags Health
// @Produce json
// @Success 200 {object} health.ReadyResponse
// @Failure 503 {object} health.ReadyResponse
// @Router /health/ready [get]
func (h *HealthHandler) ReadinessProbe(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
	defer cancel()

	resp := h.checker.Ready(ctx)
	statusCode := fiber.StatusOK
	if resp.Status == health.StatusUnhealthy {
		statusCode = fiber.StatusServiceUnavailable
	}
	return c.Status(statusCode).JSON(resp)
}

// FullHealthCheck handles GET /health/full.
// Returns detailed health status including all dependencies.
// @Summary Full health check
// @Description Get detailed health status with all dependencies
// @Tags Health
// @Produce json
// @Success 200 {object} health.HealthResponse
// @Failure 503 {object} health.HealthResponse
// @Router /health/full [get]
func (h *HealthHandler) FullHealthCheck(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
	defer cancel()

	resp := h.checker.Health(ctx)
	statusCode := fiber.StatusOK
	if resp.Status == health.StatusUnhealthy {
		statusCode = fiber.StatusServiceUnavailable
	}
	return c.Status(statusCode).JSON(resp)
}

// RedisCheck creates a dependency check for Redis.
func RedisCheck(name string, client *redis.Client) health.DependencyCheck {
	return func(ctx context.Context) health.DependencyStatus {
		start := time.Now()
		status := health.DependencyStatus{
			Name:   name,
			Status: health.StatusHealthy,
		}

		if client == nil {
			status.Status = health.StatusUnhealthy
			status.Error = "Redis client is nil"
			return status
		}

		if err := client.Ping(ctx).Err(); err != nil {
			status.Status = health.StatusUnhealthy
			status.Error = "ping failed: " + err.Error()
		}

		status.Latency = time.Since(start).Round(time.Millisecond).String()
		return status
	}
}

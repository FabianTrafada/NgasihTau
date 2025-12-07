// Package health provides health check functionality for microservices.
// It supports liveness, readiness, and detailed health checks with dependency status.
package health

import (
	"context"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
)

// Status represents the health status of a component.
type Status string

const (
	StatusHealthy   Status = "healthy"
	StatusUnhealthy Status = "unhealthy"
	StatusDegraded  Status = "degraded"
)

// DependencyCheck is a function that checks the health of a dependency.
type DependencyCheck func(ctx context.Context) DependencyStatus

// DependencyStatus represents the health status of a single dependency.
type DependencyStatus struct {
	Name    string `json:"name"`
	Status  Status `json:"status"`
	Latency string `json:"latency,omitempty"`
	Error   string `json:"error,omitempty"`
}

// HealthResponse represents the response for detailed health check.
type HealthResponse struct {
	Status       Status             `json:"status"`
	Service      string             `json:"service"`
	Version      string             `json:"version,omitempty"`
	Uptime       string             `json:"uptime"`
	Timestamp    string             `json:"timestamp"`
	Dependencies []DependencyStatus `json:"dependencies,omitempty"`
}

// LiveResponse represents the response for liveness check.
type LiveResponse struct {
	Status string `json:"status"`
}

// ReadyResponse represents the response for readiness check.
type ReadyResponse struct {
	Status       Status             `json:"status"`
	Dependencies []DependencyStatus `json:"dependencies,omitempty"`
}

// Checker manages health checks for a service.
type Checker struct {
	serviceName  string
	version      string
	startTime    time.Time
	dependencies map[string]DependencyCheck
	mu           sync.RWMutex
}

// NewChecker creates a new health checker for a service.
func NewChecker(serviceName, version string) *Checker {
	return &Checker{
		serviceName:  serviceName,
		version:      version,
		startTime:    time.Now(),
		dependencies: make(map[string]DependencyCheck),
	}
}

// AddDependency registers a dependency check.
func (c *Checker) AddDependency(name string, check DependencyCheck) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.dependencies[name] = check
}

// RemoveDependency removes a dependency check.
func (c *Checker) RemoveDependency(name string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.dependencies, name)
}

// checkDependencies runs all dependency checks concurrently.
func (c *Checker) checkDependencies(ctx context.Context) []DependencyStatus {
	c.mu.RLock()
	deps := make(map[string]DependencyCheck, len(c.dependencies))
	for k, v := range c.dependencies {
		deps[k] = v
	}
	c.mu.RUnlock()

	if len(deps) == 0 {
		return nil
	}

	results := make([]DependencyStatus, 0, len(deps))
	resultsChan := make(chan DependencyStatus, len(deps))
	var wg sync.WaitGroup

	for name, check := range deps {
		wg.Add(1)
		go func(n string, ch DependencyCheck) {
			defer wg.Done()
			resultsChan <- ch(ctx)
		}(name, check)
	}

	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	for result := range resultsChan {
		results = append(results, result)
	}

	return results
}

// aggregateStatus determines overall status from dependency statuses.
func aggregateStatus(deps []DependencyStatus) Status {
	if len(deps) == 0 {
		return StatusHealthy
	}

	hasUnhealthy := false
	hasDegraded := false

	for _, dep := range deps {
		switch dep.Status {
		case StatusUnhealthy:
			hasUnhealthy = true
		case StatusDegraded:
			hasDegraded = true
		}
	}

	if hasUnhealthy {
		return StatusUnhealthy
	}
	if hasDegraded {
		return StatusDegraded
	}
	return StatusHealthy
}

// Health returns detailed health status including all dependencies.
func (c *Checker) Health(ctx context.Context) HealthResponse {
	deps := c.checkDependencies(ctx)
	status := aggregateStatus(deps)

	return HealthResponse{
		Status:       status,
		Service:      c.serviceName,
		Version:      c.version,
		Uptime:       time.Since(c.startTime).Round(time.Second).String(),
		Timestamp:    time.Now().UTC().Format(time.RFC3339),
		Dependencies: deps,
	}
}

// Live returns simple liveness status (always healthy if service is running).
func (c *Checker) Live() LiveResponse {
	return LiveResponse{
		Status: "alive",
	}
}

// Ready returns readiness status based on critical dependencies.
func (c *Checker) Ready(ctx context.Context) ReadyResponse {
	deps := c.checkDependencies(ctx)
	status := aggregateStatus(deps)

	return ReadyResponse{
		Status:       status,
		Dependencies: deps,
	}
}

// RegisterRoutes registers health check endpoints on a Fiber app.
func (c *Checker) RegisterRoutes(app *fiber.App) {
	// Detailed health check with all dependencies
	app.Get("/health", func(ctx *fiber.Ctx) error {
		checkCtx, cancel := context.WithTimeout(ctx.Context(), 5*time.Second)
		defer cancel()

		resp := c.Health(checkCtx)
		statusCode := fiber.StatusOK
		if resp.Status == StatusUnhealthy {
			statusCode = fiber.StatusServiceUnavailable
		}
		return ctx.Status(statusCode).JSON(resp)
	})

	// Simple liveness probe
	app.Get("/health/live", func(ctx *fiber.Ctx) error {
		return ctx.JSON(c.Live())
	})

	// Readiness probe with dependency checks
	app.Get("/ready", func(ctx *fiber.Ctx) error {
		checkCtx, cancel := context.WithTimeout(ctx.Context(), 5*time.Second)
		defer cancel()

		resp := c.Ready(checkCtx)
		statusCode := fiber.StatusOK
		if resp.Status == StatusUnhealthy {
			statusCode = fiber.StatusServiceUnavailable
		}
		return ctx.Status(statusCode).JSON(resp)
	})
}

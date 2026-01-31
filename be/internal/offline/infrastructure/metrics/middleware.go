// Package metrics provides Prometheus metrics middleware for the Offline Material Service.
package metrics

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
)

// Middleware creates a Fiber middleware that records request metrics.
func (m *Metrics) Middleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()

		// Process request
		err := c.Next()

		// Record metrics
		duration := time.Since(start).Seconds()
		status := c.Response().StatusCode()
		method := c.Method()
		path := c.Route().Path

		m.RecordRequest(method, path, status, duration)

		return err
	}
}

// MetricsHandler returns a handler that exposes Prometheus metrics.
// This is typically mounted at /metrics endpoint.
func MetricsHandler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// The actual metrics are exposed via promhttp.Handler()
		// This is a placeholder - actual implementation uses adaptor
		return c.SendString("# Metrics endpoint - use promhttp.Handler() with adaptor")
	}
}

// RateLimitHeadersMiddleware adds rate limit headers to responses.
func RateLimitHeadersMiddleware(limit, remaining int, resetTime time.Time) fiber.Handler {
	return func(c *fiber.Ctx) error {
		c.Set("X-RateLimit-Limit", strconv.Itoa(limit))
		c.Set("X-RateLimit-Remaining", strconv.Itoa(remaining))
		c.Set("X-RateLimit-Reset", strconv.FormatInt(resetTime.Unix(), 10))
		return c.Next()
	}
}

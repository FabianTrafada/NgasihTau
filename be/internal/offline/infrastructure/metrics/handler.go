// Package metrics provides Prometheus metrics endpoint for the Offline Material Service.
package metrics

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// RegisterMetricsEndpoint registers the /metrics endpoint for Prometheus scraping.
func RegisterMetricsEndpoint(app *fiber.App) {
	app.Get("/metrics", adaptor.HTTPHandler(promhttp.Handler()))
}

// RegisterMetricsEndpointWithPath registers the metrics endpoint at a custom path.
func RegisterMetricsEndpointWithPath(app *fiber.App, path string) {
	app.Get(path, adaptor.HTTPHandler(promhttp.Handler()))
}

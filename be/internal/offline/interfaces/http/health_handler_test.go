package http

import (
	"encoding/json"
	"io"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"

	"ngasihtau/internal/common/health"
)

func TestHealthHandler_LivenessProbe(t *testing.T) {
	app := fiber.New()
	handler := NewHealthHandler("offline-service", "1.0.0", nil)
	handler.RegisterRoutes(app)

	req := httptest.NewRequest("GET", "/health/live", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Failed to test request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != fiber.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	var result health.LiveResponse
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if result.Status != "alive" {
		t.Errorf("Expected status 'alive', got '%s'", result.Status)
	}
}

func TestHealthHandler_ReadinessProbe_NoDependencies(t *testing.T) {
	app := fiber.New()
	handler := NewHealthHandler("offline-service", "1.0.0", nil)
	handler.RegisterRoutes(app)

	req := httptest.NewRequest("GET", "/health/ready", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Failed to test request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != fiber.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	var result health.ReadyResponse
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if result.Status != health.StatusHealthy {
		t.Errorf("Expected status 'healthy', got '%s'", result.Status)
	}
}

func TestHealthHandler_FullHealthCheck(t *testing.T) {
	app := fiber.New()
	handler := NewHealthHandler("offline-service", "1.0.0", nil)
	handler.RegisterRoutes(app)

	req := httptest.NewRequest("GET", "/health/full", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Failed to test request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != fiber.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	var result health.HealthResponse
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if result.Status != health.StatusHealthy {
		t.Errorf("Expected status 'healthy', got '%s'", result.Status)
	}

	if result.Service != "offline-service" {
		t.Errorf("Expected service 'offline-service', got '%s'", result.Service)
	}

	if result.Version != "1.0.0" {
		t.Errorf("Expected version '1.0.0', got '%s'", result.Version)
	}

	if result.Uptime == "" {
		t.Error("Expected uptime to be set")
	}

	if result.Timestamp == "" {
		t.Error("Expected timestamp to be set")
	}
}

func TestHealthHandler_HealthAlias(t *testing.T) {
	app := fiber.New()
	handler := NewHealthHandler("offline-service", "1.0.0", nil)
	handler.RegisterRoutes(app)

	// Test /health alias
	req := httptest.NewRequest("GET", "/health", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Failed to test request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != fiber.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

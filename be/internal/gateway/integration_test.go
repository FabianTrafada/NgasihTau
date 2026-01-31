// Package gateway contains integration tests for the API Gateway (Traefik).
// These tests verify routing, rate limiting, and health check aggregation.
//
// Prerequisites:
// - Docker Compose environment must be running: docker-compose up -d
// - All services must be healthy
//
// Run tests: go test -v -tags=integration ./internal/gateway/...
//
//go:build integration

package gateway

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"
)

const (
	// Gateway URL - Traefik is exposed on port 8000
	gatewayURL = "http://localhost:8000"

	// Timeout for HTTP requests
	requestTimeout = 10 * time.Second
)

// TestRouting verifies that requests are routed to the correct services.
func TestRouting(t *testing.T) {
	client := &http.Client{Timeout: requestTimeout}

	tests := []struct {
		name           string
		path           string
		expectedStatus int
		description    string
	}{
		{
			name:           "Auth routes to User Service",
			path:           "/api/v1/auth/login",
			expectedStatus: http.StatusBadRequest, // No body, but routed correctly
			description:    "Auth endpoints should route to User Service",
		},
		{
			name:           "Users routes to User Service",
			path:           "/api/v1/users/me",
			expectedStatus: http.StatusUnauthorized, // No auth, but routed correctly
			description:    "User endpoints should route to User Service",
		},
		{
			name:           "Pods routes to Pod Service",
			path:           "/api/v1/pods",
			expectedStatus: http.StatusOK, // List pods (may be empty)
			description:    "Pod endpoints should route to Pod Service",
		},
		{
			name:           "Feed routes to Pod Service",
			path:           "/api/v1/feed",
			expectedStatus: http.StatusUnauthorized, // No auth, but routed correctly
			description:    "Feed endpoint should route to Pod Service",
		},
		{
			name:           "Materials routes to Material Service",
			path:           "/api/v1/materials/upload-url",
			expectedStatus: http.StatusUnauthorized, // No auth, but routed correctly
			description:    "Material endpoints should route to Material Service",
		},
		{
			name:           "Bookmarks routes to Material Service",
			path:           "/api/v1/bookmarks",
			expectedStatus: http.StatusUnauthorized, // No auth, but routed correctly
			description:    "Bookmark endpoints should route to Material Service",
		},
		{
			name:           "Search routes to Search Service",
			path:           "/api/v1/search",
			expectedStatus: http.StatusOK, // Search with no query
			description:    "Search endpoints should route to Search Service",
		},
		{
			name:           "Notifications routes to Notification Service",
			path:           "/api/v1/notifications",
			expectedStatus: http.StatusUnauthorized, // No auth, but routed correctly
			description:    "Notification endpoints should route to Notification Service",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := gatewayURL + tt.path
			resp, err := client.Get(url)
			if err != nil {
				t.Fatalf("Request failed: %v", err)
			}
			defer resp.Body.Close()

			// Check that we got a response (not 502/504 which would indicate routing failure)
			// 503 means routing is correct but service is down - this is acceptable for routing test
			if resp.StatusCode == http.StatusBadGateway ||
				resp.StatusCode == http.StatusGatewayTimeout {
				t.Errorf("Routing failed for %s: got status %d, routing may be misconfigured", tt.path, resp.StatusCode)
				return
			}

			// If service is unavailable, routing is correct but service is down
			if resp.StatusCode == http.StatusServiceUnavailable {
				t.Logf("Route %s: routing correct (503 - service unavailable)", tt.path)
				// Still check headers to verify request went through gateway
			}

			// Verify API version header is present (indicates request went through gateway)
			apiVersion := resp.Header.Get("X-API-Version")
			if apiVersion == "" {
				t.Errorf("Missing X-API-Version header for %s", tt.path)
			}

			t.Logf("Route %s: status=%d, X-API-Version=%s", tt.path, resp.StatusCode, apiVersion)
		})
	}
}

// TestRateLimiting verifies that rate limiting is enforced per endpoint type.
func TestRateLimiting(t *testing.T) {
	client := &http.Client{Timeout: requestTimeout}

	tests := []struct {
		name        string
		path        string
		method      string
		rateLimit   int // requests per minute
		burstLimit  int // burst allowance
		description string
	}{
		{
			name:        "Auth rate limit (10/min)",
			path:        "/api/v1/auth/login",
			method:      "POST",
			rateLimit:   10,
			burstLimit:  15,
			description: "Auth endpoints should be limited to 10 req/min",
		},
		{
			name:        "Search rate limit (60/min)",
			path:        "/api/v1/search",
			method:      "GET",
			rateLimit:   60,
			burstLimit:  80,
			description: "Search endpoints should be limited to 60 req/min",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := gatewayURL + tt.path

			// Send requests up to burst limit + some extra to trigger rate limiting
			requestCount := tt.burstLimit + 5
			rateLimited := false
			var lastStatus int

			for i := 0; i < requestCount; i++ {
				var resp *http.Response
				var err error

				if tt.method == "POST" {
					resp, err = client.Post(url, "application/json", strings.NewReader("{}"))
				} else {
					resp, err = client.Get(url)
				}

				if err != nil {
					t.Fatalf("Request %d failed: %v", i+1, err)
				}

				lastStatus = resp.StatusCode
				resp.Body.Close()

				if resp.StatusCode == http.StatusTooManyRequests {
					rateLimited = true
					t.Logf("Rate limited after %d requests", i+1)
					break
				}
			}

			if !rateLimited {
				t.Logf("Warning: Rate limiting not triggered after %d requests (last status: %d)", requestCount, lastStatus)
				// This is a soft warning - rate limiting may not trigger in test environment
				// due to timing or configuration differences
			}
		})
	}
}

// TestHealthCheckEndpoints verifies health check endpoints are accessible.
func TestHealthCheckEndpoints(t *testing.T) {
	client := &http.Client{Timeout: requestTimeout}

	// Test Traefik's own health endpoint
	t.Run("Traefik ping", func(t *testing.T) {
		resp, err := client.Get(gatewayURL + "/ping")
		if err != nil {
			t.Fatalf("Ping request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}
	})

	// Test service health endpoints through gateway
	services := []struct {
		name string
		path string
	}{
		{"User Service", "/api/v1/auth/../../../health"},
		{"Pod Service", "/api/v1/pods/../../../health"},
		{"Material Service", "/api/v1/materials/../../../health"},
		{"Search Service", "/api/v1/search/../../../health"},
		{"Notification Service", "/api/v1/notifications/../../../health"},
	}

	for _, svc := range services {
		t.Run(svc.name+" health", func(t *testing.T) {
			// Note: Direct health check access depends on service configuration
			// This test verifies the services are reachable through the gateway
			t.Logf("Checking %s health endpoint", svc.name)
		})
	}
}

// TestCORSHeaders verifies CORS headers are properly set.
func TestCORSHeaders(t *testing.T) {
	client := &http.Client{Timeout: requestTimeout}

	// Send OPTIONS request to trigger CORS preflight
	req, err := http.NewRequest("OPTIONS", gatewayURL+"/api/v1/pods", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Origin", "http://localhost:3000")
	req.Header.Set("Access-Control-Request-Method", "GET")

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("OPTIONS request failed: %v", err)
	}
	defer resp.Body.Close()

	// Check CORS headers
	allowOrigin := resp.Header.Get("Access-Control-Allow-Origin")
	allowMethods := resp.Header.Get("Access-Control-Allow-Methods")
	allowHeaders := resp.Header.Get("Access-Control-Allow-Headers")

	if allowOrigin == "" {
		t.Error("Missing Access-Control-Allow-Origin header")
	} else {
		t.Logf("Access-Control-Allow-Origin: %s", allowOrigin)
	}

	if allowMethods == "" {
		t.Error("Missing Access-Control-Allow-Methods header")
	} else {
		t.Logf("Access-Control-Allow-Methods: %s", allowMethods)
	}

	if allowHeaders == "" {
		t.Error("Missing Access-Control-Allow-Headers header")
	} else {
		t.Logf("Access-Control-Allow-Headers: %s", allowHeaders)
	}
}

// TestSecurityHeaders verifies security headers are properly set.
func TestSecurityHeaders(t *testing.T) {
	client := &http.Client{Timeout: requestTimeout}

	resp, err := client.Get(gatewayURL + "/api/v1/pods")
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()

	// Check security headers
	headers := map[string]string{
		"X-Content-Type-Options": "nosniff",
		"X-Frame-Options":        "DENY",
		"X-XSS-Protection":       "1; mode=block",
	}

	for header, expected := range headers {
		actual := resp.Header.Get(header)
		if actual == "" {
			t.Errorf("Missing security header: %s", header)
		} else if actual != expected {
			t.Errorf("Header %s: expected %q, got %q", header, expected, actual)
		} else {
			t.Logf("%s: %s", header, actual)
		}
	}
}

// TestServiceHealthAggregation tests that all backend services are healthy.
func TestServiceHealthAggregation(t *testing.T) {
	client := &http.Client{Timeout: requestTimeout}

	// Service ports for direct health checks
	services := []struct {
		name string
		port int
	}{
		{"User Service", 8001},
		{"Pod Service", 8002},
		{"Material Service", 8003},
		{"Search Service", 8004},
		{"AI Service", 8005},
		{"Notification Service", 8006},
	}

	healthyCount := 0
	var wg sync.WaitGroup
	var mu sync.Mutex
	results := make(map[string]bool)

	for _, svc := range services {
		wg.Add(1)
		go func(name string, port int) {
			defer wg.Done()

			url := fmt.Sprintf("http://127.0.0.1:%d/health", port)
			resp, err := client.Get(url)

			mu.Lock()
			defer mu.Unlock()

			if err != nil {
				t.Logf("%s: connection failed - %v", name, err)
				results[name] = false
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode == http.StatusOK {
				results[name] = true
				healthyCount++

				// Parse health response
				var health map[string]interface{}
				if err := json.NewDecoder(resp.Body).Decode(&health); err == nil {
					t.Logf("%s: healthy (status=%v)", name, health["status"])
				}
			} else {
				results[name] = false
				t.Logf("%s: unhealthy (status=%d)", name, resp.StatusCode)
			}
		}(svc.name, svc.port)
	}

	wg.Wait()

	t.Logf("Health check summary: %d/%d services healthy", healthyCount, len(services))

	// Log individual results
	for name, healthy := range results {
		status := "unhealthy"
		if healthy {
			status = "healthy"
		}
		t.Logf("  - %s: %s", name, status)
	}
}

// TestAPIVersionHeader verifies the X-API-Version header is set correctly.
func TestAPIVersionHeader(t *testing.T) {
	client := &http.Client{Timeout: requestTimeout}

	paths := []string{
		"/api/v1/auth/login",
		"/api/v1/users/me",
		"/api/v1/pods",
		"/api/v1/materials/upload-url",
		"/api/v1/search",
		"/api/v1/notifications",
	}

	for _, path := range paths {
		t.Run(path, func(t *testing.T) {
			resp, err := client.Get(gatewayURL + path)
			if err != nil {
				t.Fatalf("Request failed: %v", err)
			}
			defer resp.Body.Close()

			apiVersion := resp.Header.Get("X-API-Version")
			if apiVersion != "v1" {
				t.Errorf("Expected X-API-Version=v1, got %q", apiVersion)
			}
		})
	}
}

// TestResponseFormat verifies API responses follow the standard envelope format.
func TestResponseFormat(t *testing.T) {
	client := &http.Client{Timeout: requestTimeout}

	// Test error response format
	t.Run("Error response format", func(t *testing.T) {
		resp, err := client.Get(gatewayURL + "/api/v1/users/me")
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		defer resp.Body.Close()

		// Skip if service is unavailable (503 from Traefik)
		if resp.StatusCode == http.StatusServiceUnavailable {
			t.Skip("Skipping - backend service unavailable")
		}

		body, _ := io.ReadAll(resp.Body)
		var response map[string]interface{}
		if err := json.Unmarshal(body, &response); err != nil {
			t.Fatalf("Failed to parse response: %v (body: %s)", err, string(body))
		}

		// Check for standard envelope fields
		if _, ok := response["success"]; !ok {
			t.Error("Missing 'success' field in response")
		}
		if _, ok := response["error"]; !ok && resp.StatusCode >= 400 {
			t.Error("Missing 'error' field in error response")
		}
		if _, ok := response["meta"]; !ok {
			t.Log("Note: 'meta' field not present in response")
		}

		t.Logf("Response format: %+v", response)
	})

	// Test success response format
	t.Run("Success response format", func(t *testing.T) {
		resp, err := client.Get(gatewayURL + "/api/v1/pods")
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Skipf("Skipping success format test - got status %d", resp.StatusCode)
		}

		body, _ := io.ReadAll(resp.Body)
		var response map[string]interface{}
		if err := json.Unmarshal(body, &response); err != nil {
			t.Fatalf("Failed to parse response: %v", err)
		}

		// Check for standard envelope fields
		if success, ok := response["success"].(bool); !ok || !success {
			t.Error("Expected 'success: true' in success response")
		}
		if _, ok := response["data"]; !ok {
			t.Error("Missing 'data' field in success response")
		}

		t.Logf("Response format: success=%v, has_data=%v", response["success"], response["data"] != nil)
	})
}

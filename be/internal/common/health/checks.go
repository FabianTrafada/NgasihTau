package health

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresCheck creates a dependency check for PostgreSQL.
func PostgresCheck(name string, pool *pgxpool.Pool) DependencyCheck {
	return func(ctx context.Context) DependencyStatus {
		start := time.Now()
		status := DependencyStatus{
			Name:   name,
			Status: StatusHealthy,
		}

		if pool == nil {
			status.Status = StatusUnhealthy
			status.Error = "database pool is nil"
			return status
		}

		if err := pool.Ping(ctx); err != nil {
			status.Status = StatusUnhealthy
			status.Error = fmt.Sprintf("ping failed: %v", err)
		}

		status.Latency = time.Since(start).Round(time.Millisecond).String()
		return status
	}
}

// HTTPServiceCheck creates a dependency check for an HTTP service.
func HTTPServiceCheck(name, healthURL string) DependencyCheck {
	return func(ctx context.Context) DependencyStatus {
		start := time.Now()
		status := DependencyStatus{
			Name:   name,
			Status: StatusHealthy,
		}

		client := &http.Client{Timeout: 3 * time.Second}
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, healthURL, nil)
		if err != nil {
			status.Status = StatusUnhealthy
			status.Error = fmt.Sprintf("failed to create request: %v", err)
			return status
		}

		resp, err := client.Do(req)
		if err != nil {
			status.Status = StatusUnhealthy
			status.Error = fmt.Sprintf("request failed: %v", err)
			status.Latency = time.Since(start).Round(time.Millisecond).String()
			return status
		}
		defer resp.Body.Close()

		if resp.StatusCode >= 400 {
			status.Status = StatusUnhealthy
			status.Error = fmt.Sprintf("unhealthy status code: %d", resp.StatusCode)
		}

		status.Latency = time.Since(start).Round(time.Millisecond).String()
		return status
	}
}

// NATSChecker is an interface for NATS client health check.
type NATSChecker interface {
	IsConnected() bool
}

// NATSCheck creates a dependency check for NATS.
func NATSCheck(name string, client NATSChecker) DependencyCheck {
	return func(ctx context.Context) DependencyStatus {
		status := DependencyStatus{
			Name:   name,
			Status: StatusHealthy,
		}

		if client == nil {
			status.Status = StatusDegraded
			status.Error = "NATS client not configured"
			return status
		}

		if !client.IsConnected() {
			status.Status = StatusUnhealthy
			status.Error = "NATS connection lost"
		}

		return status
	}
}

// MinIOChecker is an interface for MinIO client health check.
type MinIOChecker interface {
	HealthCheck(ctx context.Context) error
}

// MinIOCheck creates a dependency check for MinIO.
func MinIOCheck(name string, client MinIOChecker) DependencyCheck {
	return func(ctx context.Context) DependencyStatus {
		start := time.Now()
		status := DependencyStatus{
			Name:   name,
			Status: StatusHealthy,
		}

		if client == nil {
			status.Status = StatusUnhealthy
			status.Error = "MinIO client is nil"
			return status
		}

		if err := client.HealthCheck(ctx); err != nil {
			status.Status = StatusUnhealthy
			status.Error = fmt.Sprintf("health check failed: %v", err)
		}

		status.Latency = time.Since(start).Round(time.Millisecond).String()
		return status
	}
}

// QdrantChecker is an interface for Qdrant client health check.
type QdrantChecker interface {
	HealthCheck(ctx context.Context) error
}

// QdrantCheck creates a dependency check for Qdrant.
func QdrantCheck(name string, client QdrantChecker) DependencyCheck {
	return func(ctx context.Context) DependencyStatus {
		start := time.Now()
		status := DependencyStatus{
			Name:   name,
			Status: StatusHealthy,
		}

		if client == nil {
			status.Status = StatusUnhealthy
			status.Error = "Qdrant client is nil"
			return status
		}

		if err := client.HealthCheck(ctx); err != nil {
			status.Status = StatusUnhealthy
			status.Error = fmt.Sprintf("health check failed: %v", err)
		}

		status.Latency = time.Since(start).Round(time.Millisecond).String()
		return status
	}
}

// MeilisearchChecker is an interface for Meilisearch client health check.
type MeilisearchChecker interface {
	HealthCheck(ctx context.Context) error
}

// MeilisearchCheck creates a dependency check for Meilisearch.
func MeilisearchCheck(name string, client MeilisearchChecker) DependencyCheck {
	return func(ctx context.Context) DependencyStatus {
		start := time.Now()
		status := DependencyStatus{
			Name:   name,
			Status: StatusHealthy,
		}

		if client == nil {
			status.Status = StatusUnhealthy
			status.Error = "Meilisearch client is nil"
			return status
		}

		if err := client.HealthCheck(ctx); err != nil {
			status.Status = StatusUnhealthy
			status.Error = fmt.Sprintf("health check failed: %v", err)
		}

		status.Latency = time.Since(start).Round(time.Millisecond).String()
		return status
	}
}

// CustomCheck creates a custom dependency check with a provided function.
func CustomCheck(name string, checkFn func(ctx context.Context) error) DependencyCheck {
	return func(ctx context.Context) DependencyStatus {
		start := time.Now()
		status := DependencyStatus{
			Name:   name,
			Status: StatusHealthy,
		}

		if err := checkFn(ctx); err != nil {
			status.Status = StatusUnhealthy
			status.Error = err.Error()
		}

		status.Latency = time.Since(start).Round(time.Millisecond).String()
		return status
	}
}

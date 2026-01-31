# API Gateway Integration Tests

This package contains integration tests for the Traefik API Gateway configuration.

## Prerequisites

1. Docker Compose environment must be running:
   ```bash
   cd be
   docker-compose up -d
   ```

2. All backend services must be healthy and running on their respective ports:
   - User Service: 8001
   - Pod Service: 8002
   - Material Service: 8003
   - Search Service: 8004
   - AI Service: 8005
   - Notification Service: 8006

3. Traefik must be accessible on port 8000

## Running Tests

Run all integration tests:
```bash
cd be
go test -v -tags=integration ./internal/gateway/...
```

Run specific test:
```bash
go test -v -tags=integration -run TestRouting ./internal/gateway/...
```

## Test Coverage

| Test | Description |
|------|-------------|
| TestRouting | Verifies requests are routed to correct services |
| TestRateLimiting | Verifies rate limiting per endpoint type |
| TestHealthCheckEndpoints | Verifies health endpoints are accessible |
| TestCORSHeaders | Verifies CORS headers are set correctly |
| TestSecurityHeaders | Verifies security headers are present |
| TestServiceHealthAggregation | Checks all backend services health |
| TestAPIVersionHeader | Verifies X-API-Version header |
| TestResponseFormat | Verifies standard API response format |

## Rate Limits Tested

Per Requirement 10.8:
- Auth endpoints: 10 req/min (burst: 15)
- Search endpoints: 60 req/min (burst: 80)
- AI Chat endpoints: 30 req/min (burst: 40)
- File upload endpoints: 10 req/min (burst: 15)
- General API: 100 req/min (burst: 150)

# ===========================================
# NgasihTau Backend Services - Multi-stage Dockerfile
# ===========================================
# Build stage
FROM golang:1.23-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /app

# Copy go mod files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build argument for service selection
ARG SERVICE=user-service

# Build the service
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /app/service ./cmd/${SERVICE}

# ===========================================
# Runtime stage
FROM alpine:3.19

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata wget

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/service .

# Copy migrations (needed for some services)
COPY --from=builder /app/migrations ./migrations

# Copy API swagger docs
COPY --from=builder /app/api ./api

# Create non-root user
RUN adduser -D -g '' appuser
USER appuser

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD wget -q --spider http://localhost:${PORT:-8001}/health || exit 1

# Default command
CMD ["./service"]

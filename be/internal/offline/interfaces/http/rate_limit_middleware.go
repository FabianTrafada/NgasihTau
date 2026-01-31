// Package http provides HTTP handlers for the Offline Material Service API.
package http

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"ngasihtau/internal/common/errors"
	"ngasihtau/internal/common/middleware"
	"ngasihtau/internal/offline/application"
	"ngasihtau/internal/offline/domain"
)

// sendMiddlewareError sends an error response from middleware.
func sendMiddlewareError(c *fiber.Ctx, status int, code errors.Code, message string) error {
	requestID := middleware.GetRequestID(c)
	appErr := errors.New(code, message)
	resp := errors.BuildResponse(requestID, appErr)
	return c.Status(status).JSON(resp)
}

// RateLimitMiddleware creates a middleware that enforces rate limiting.
// Implements Requirement 6.6: Add rate limit headers to responses.
func RateLimitMiddleware(rateLimiter application.RateLimiter) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get user ID from context (set by auth middleware)
		userIDStr := c.Locals("user_id")
		if userIDStr == nil {
			return c.Next()
		}

		userID, err := uuid.Parse(userIDStr.(string))
		if err != nil {
			return c.Next()
		}

		// Check rate limit
		allowed, remaining, resetTime, err := rateLimiter.CheckDownloadLimit(c.Context(), userID)
		if err != nil {
			// Log error but don't block request
			return c.Next()
		}

		// Set rate limit headers
		c.Set("X-RateLimit-Limit", strconv.Itoa(domain.MaxDownloadsPerHour))
		c.Set("X-RateLimit-Remaining", strconv.Itoa(remaining))
		c.Set("X-RateLimit-Reset", strconv.FormatInt(resetTime.Unix(), 10))

		if !allowed {
			c.Set("Retry-After", strconv.FormatInt(int64(time.Until(resetTime).Seconds()), 10))
			return sendMiddlewareError(c, fiber.StatusTooManyRequests, errors.CodeRateLimited, "Download rate limit exceeded")
		}

		return c.Next()
	}
}

// DeviceBlockMiddleware creates a middleware that checks if a device is blocked.
// Implements Requirement 6.4: Device blocking after 5 failures.
func DeviceBlockMiddleware(rateLimiter application.RateLimiter) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get device ID from request body or params
		deviceIDStr := c.Params("device_id")
		if deviceIDStr == "" {
			// Try to get from request body
			var body struct {
				DeviceID string `json:"device_id"`
			}
			if err := c.BodyParser(&body); err == nil && body.DeviceID != "" {
				deviceIDStr = body.DeviceID
			}
		}

		if deviceIDStr == "" {
			return c.Next()
		}

		deviceID, err := uuid.Parse(deviceIDStr)
		if err != nil {
			return c.Next()
		}

		// Check if device is blocked
		isBlocked, blockedUntil, err := rateLimiter.IsDeviceBlocked(c.Context(), deviceID)
		if err != nil {
			// Log error but don't block request
			return c.Next()
		}

		if isBlocked {
			c.Set("Retry-After", strconv.FormatInt(int64(time.Until(blockedUntil).Seconds()), 10))
			return sendMiddlewareError(c, fiber.StatusTooManyRequests, errors.CodeRateLimited, "Device is temporarily blocked")
		}

		return c.Next()
	}
}

// SecurityMiddleware creates a middleware that validates request signatures and prevents replay attacks.
// Implements Requirement 7.1: Request signing validation.
// Implements Requirement 7.2: Replay attack prevention.
func SecurityMiddleware(securityService application.SecurityService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Check if signature header is present
		signatureHeader := c.Get("X-Signature")
		if signatureHeader == "" {
			// Signature not required for all endpoints
			return c.Next()
		}

		// Parse signature header
		timestamp, nonce, signature, err := application.ParseSignatureHeader(signatureHeader)
		if err != nil {
			return sendMiddlewareError(c, fiber.StatusBadRequest, errors.CodeBadRequest, "Invalid signature header format")
		}

		// Check for replay attack
		if err := securityService.CheckReplayAttack(c.Context(), nonce, timestamp); err != nil {
			return sendMiddlewareError(c, fiber.StatusForbidden, errors.CodeForbidden, "Request rejected")
		}

		// Validate signature
		input := application.ValidateSignatureInput{
			Method:    c.Method(),
			Path:      c.Path(),
			Body:      c.Body(),
			Timestamp: timestamp,
			Nonce:     nonce,
			Signature: signature,
			DeviceID:  c.Get("X-Device-ID"),
		}

		if err := securityService.ValidateRequestSignature(c.Context(), input); err != nil {
			return sendMiddlewareError(c, fiber.StatusForbidden, errors.CodeForbidden, "Request rejected")
		}

		// Record the nonce to prevent replay
		if err := securityService.RecordRequest(c.Context(), nonce); err != nil {
			// Log error but don't block request
		}

		return c.Next()
	}
}

// AuditMiddleware creates a middleware that logs audit events for security-relevant operations.
// Implements Requirement 7.5: Audit logging.
func AuditMiddleware(securityService application.SecurityService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Process request
		err := c.Next()

		// Get user ID
		userIDStr := c.Locals("user_id")
		if userIDStr == nil {
			return err
		}

		userID, parseErr := uuid.Parse(userIDStr.(string))
		if parseErr != nil {
			return err
		}

		// Determine action based on path and method
		action := getAuditAction(c.Method(), c.Path())
		if action == "" {
			return err
		}

		// Get device ID if present
		var deviceID *uuid.UUID
		deviceIDStr := c.Get("X-Device-ID")
		if deviceIDStr != "" {
			if id, parseErr := uuid.Parse(deviceIDStr); parseErr == nil {
				deviceID = &id
			}
		}

		// Determine resource and resource ID
		resource, resourceID := getAuditResourceFromParams(c)

		// Create audit log
		success := c.Response().StatusCode() < 400
		var errorCode *string
		if !success {
			code := getErrorCodeFromResponse(c)
			if code != "" {
				errorCode = &code
			}
		}

		auditLog := domain.NewAuditLog(
			userID,
			deviceID,
			action,
			resource,
			resourceID,
			c.IP(),
			c.Get("User-Agent"),
			success,
			errorCode,
		)

		// Log asynchronously (fire and forget)
		go func() {
			_ = securityService.LogAuditEvent(c.Context(), auditLog)
		}()

		return err
	}
}

// getAuditAction determines the audit action based on HTTP method and path.
func getAuditAction(method, path string) string {
	switch {
	case method == "POST" && contains(path, "/devices"):
		return domain.AuditActionDeviceRegister
	case method == "DELETE" && contains(path, "/devices"):
		return domain.AuditActionDeviceDeregister
	case method == "POST" && contains(path, "/license"):
		return domain.AuditActionLicenseIssue
	case method == "POST" && contains(path, "/validate"):
		return domain.AuditActionLicenseValidate
	case method == "POST" && contains(path, "/renew"):
		return domain.AuditActionLicenseRenew
	case method == "GET" && contains(path, "/download"):
		return domain.AuditActionMaterialDownload
	default:
		return ""
	}
}

// getAuditResourceFromParams determines the resource type and ID from the Fiber context.
func getAuditResourceFromParams(c *fiber.Ctx) (string, uuid.UUID) {
	path := c.Path()
	switch {
	case contains(path, "/devices"):
		if id := c.Params("device_id"); id != "" {
			if uid, err := uuid.Parse(id); err == nil {
				return domain.AuditResourceDevice, uid
			}
		}
		return domain.AuditResourceDevice, uuid.Nil
	case contains(path, "/licenses"):
		if id := c.Params("license_id"); id != "" {
			if uid, err := uuid.Parse(id); err == nil {
				return domain.AuditResourceLicense, uid
			}
		}
		return domain.AuditResourceLicense, uuid.Nil
	case contains(path, "/materials"):
		if id := c.Params("material_id"); id != "" {
			if uid, err := uuid.Parse(id); err == nil {
				return domain.AuditResourceMaterial, uid
			}
		}
		return domain.AuditResourceMaterial, uuid.Nil
	default:
		return "", uuid.Nil
	}
}

// getErrorCodeFromResponse extracts error code from response body.
func getErrorCodeFromResponse(c *fiber.Ctx) string {
	switch c.Response().StatusCode() {
	case 400:
		return "INVALID_REQUEST"
	case 401:
		return "UNAUTHORIZED"
	case 403:
		return "FORBIDDEN"
	case 404:
		return "NOT_FOUND"
	case 429:
		return "RATE_LIMIT_EXCEEDED"
	default:
		return "INTERNAL_ERROR"
	}
}

// contains checks if a string contains a substring.
func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// Package middleware provides HTTP middleware for the API.
// Implements requirement 10.3 for request ID and logging middleware.
package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

const (
	// RequestIDHeader is the header name for request ID.
	RequestIDHeader = "X-Request-ID"
	// RequestIDKey is the context key for request ID.
	RequestIDKey = "request_id"
	// UserIDKey is the context key for authenticated user ID.
	UserIDKey = "user_id"
	// UserRoleKey is the context key for authenticated user role.
	UserRoleKey = "user_role"
)

// RequestID middleware generates or extracts a request ID for each request.
func RequestID() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Check if request ID is provided in header
		requestID := c.Get(RequestIDHeader)
		if requestID == "" {
			requestID = uuid.New().String()
		}

		// Store in context and set response header
		c.Locals(RequestIDKey, requestID)
		c.Set(RequestIDHeader, requestID)

		return c.Next()
	}
}

// GetRequestID extracts the request ID from the Fiber context.
func GetRequestID(c *fiber.Ctx) string {
	if id, ok := c.Locals(RequestIDKey).(string); ok {
		return id
	}
	return ""
}

// GetUserID extracts the authenticated user ID from the Fiber context.
func GetUserID(c *fiber.Ctx) (uuid.UUID, bool) {
	if id, ok := c.Locals(UserIDKey).(uuid.UUID); ok {
		return id, true
	}
	return uuid.Nil, false
}

// GetUserRole extracts the authenticated user role from the Fiber context.
func GetUserRole(c *fiber.Ctx) string {
	if role, ok := c.Locals(UserRoleKey).(string); ok {
		return role
	}
	return ""
}

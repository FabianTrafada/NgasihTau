package middleware

import (
	stderrors "errors"
	"strings"

	"github.com/gofiber/fiber/v2"

	"ngasihtau/internal/common/errors"
	"ngasihtau/pkg/jwt"
)

// Auth middleware validates JWT tokens and sets user context.
// Implements requirement 1 for JWT validation on protected endpoints.
// Implements requirement 10 for API Gateway token validation.
func Auth(jwtManager *jwt.Manager) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Extract token from Authorization header
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return sendError(c, errors.Unauthorized("missing authorization header"))
		}

		// Check Bearer prefix
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			return sendError(c, errors.Unauthorized("invalid authorization header format"))
		}

		token := parts[1]
		if token == "" {
			return sendError(c, errors.Unauthorized("missing token"))
		}

		// Validate token
		claims, err := jwtManager.ValidateAccessToken(token)
		if err != nil {
			// Handle specific token errors for better client feedback
			return sendError(c, mapTokenError(err))
		}

		// Set user context (UserID is already a uuid.UUID from claims)
		c.Locals(UserIDKey, claims.UserID)
		c.Locals(UserRoleKey, claims.Role)

		return c.Next()
	}
}

// mapTokenError maps JWT package errors to appropriate AppError responses.
// This provides specific error messages for different token validation failures.
func mapTokenError(err error) *errors.AppError {
	switch {
	case stderrors.Is(err, jwt.ErrExpiredToken):
		return errors.Unauthorized("token has expired")
	case stderrors.Is(err, jwt.ErrTokenNotYetValid):
		return errors.Unauthorized("token is not yet valid")
	case stderrors.Is(err, jwt.ErrInvalidSignature):
		return errors.Unauthorized("invalid token signature")
	case stderrors.Is(err, jwt.ErrInvalidClaims):
		return errors.Unauthorized("invalid token claims")
	case stderrors.Is(err, jwt.ErrInvalidToken):
		return errors.Unauthorized("invalid token")
	default:
		return errors.Unauthorized("invalid or expired token")
	}
}

// OptionalAuth middleware validates JWT tokens if present but doesn't require them.
// Useful for endpoints that behave differently for authenticated users.
func OptionalAuth(jwtManager *jwt.Manager) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Next()
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			return c.Next()
		}

		token := parts[1]
		if token == "" {
			return c.Next()
		}

		claims, err := jwtManager.ValidateAccessToken(token)
		if err != nil {
			return c.Next()
		}

		// Set user context (UserID is already a uuid.UUID from claims)
		c.Locals(UserIDKey, claims.UserID)
		c.Locals(UserRoleKey, claims.Role)

		return c.Next()
	}
}

// sendError sends an error response using the standard error format.
func sendError(c *fiber.Ctx, err *errors.AppError) error {
	requestID := GetRequestID(c)
	resp := errors.BuildResponse(requestID, err)
	return c.Status(err.HTTPStatus()).JSON(resp)
}

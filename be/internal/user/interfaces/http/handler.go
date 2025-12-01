// Package http provides HTTP handlers for the User Service API.
// Implements the interfaces layer in Clean Architecture.
package http

import (
	"github.com/gofiber/fiber/v2"

	"ngasihtau/internal/common/middleware"
	"ngasihtau/internal/user/application"
	"ngasihtau/pkg/jwt"
)

// Handler contains all HTTP handlers for the User Service.
type Handler struct {
	authHandler *AuthHandler
	userHandler *UserHandler
	jwtManager  *jwt.Manager
}

// NewHandler creates a new Handler with the given dependencies.
func NewHandler(userService application.UserService, jwtManager *jwt.Manager) *Handler {
	return &Handler{
		authHandler: NewAuthHandler(userService),
		userHandler: NewUserHandler(userService),
		jwtManager:  jwtManager,
	}
}

// RegisterRoutes registers all User Service routes on the given Fiber app.
// Routes are organized by functionality:
//
// Authentication (public):
//   - POST /api/v1/auth/register
//   - POST /api/v1/auth/login
//   - POST /api/v1/auth/google (Google OAuth login)
//   - POST /api/v1/auth/refresh
//   - POST /api/v1/auth/logout
//   - POST /api/v1/auth/2fa/login (2FA login verification)
//   - POST /api/v1/auth/verify-email (email verification with token)
//   - POST /api/v1/auth/password/forgot (request password reset)
//   - POST /api/v1/auth/password/reset (reset password with token)
//
// 2FA (protected):
//   - POST /api/v1/auth/2fa/enable
//   - POST /api/v1/auth/2fa/verify
//   - POST /api/v1/auth/2fa/disable
//
// Email Verification (protected):
//   - POST /api/v1/auth/send-verification (send verification email)
//
// User Profile (protected):
//   - GET  /api/v1/users/me
//   - PUT  /api/v1/users/me
//
// User Public Profile (public with optional auth):
//   - GET  /api/v1/users/:id
//
// Follow (protected):
//   - POST   /api/v1/users/:id/follow
//   - DELETE /api/v1/users/:id/follow
//
// Followers/Following (public):
//   - GET    /api/v1/users/:id/followers
//   - GET    /api/v1/users/:id/following
func (h *Handler) RegisterRoutes(app *fiber.App) {
	// API v1 group
	api := app.Group("/api/v1")

	// Auth routes (public)
	auth := api.Group("/auth")
	auth.Post("/register", h.authHandler.Register)
	auth.Post("/login", h.authHandler.Login)
	auth.Post("/google", h.authHandler.GoogleLogin)
	auth.Post("/refresh", h.authHandler.RefreshToken)
	auth.Post("/logout", h.authHandler.Logout)

	// Email verification and password reset (public - uses token)
	auth.Post("/verify-email", h.authHandler.VerifyEmail)
	auth.Post("/password/forgot", h.authHandler.RequestPasswordReset)
	auth.Post("/password/reset", h.authHandler.ResetPassword)

	// 2FA login verification (public - uses temp token)
	auth.Post("/2fa/login", h.authHandler.Verify2FALogin)

	// 2FA management routes (protected)
	twoFA := auth.Group("/2fa", middleware.Auth(h.jwtManager))
	twoFA.Post("/enable", h.authHandler.Enable2FA)
	twoFA.Post("/verify", h.authHandler.Verify2FA)
	twoFA.Post("/disable", h.authHandler.Disable2FA)

	// Send verification email (protected - requires authentication)
	auth.Post("/send-verification", middleware.Auth(h.jwtManager), h.authHandler.SendVerificationEmail)

	// User routes
	users := api.Group("/users")

	// Protected routes - require authentication
	protected := users.Group("", middleware.Auth(h.jwtManager))
	protected.Get("/me", h.userHandler.GetCurrentUser)
	protected.Put("/me", h.userHandler.UpdateCurrentUser)

	// Public user profile with optional auth (to check if following)
	users.Get("/:id", middleware.OptionalAuth(h.jwtManager), h.userHandler.GetUser)

	// Follow routes - require authentication
	users.Post("/:id/follow", middleware.Auth(h.jwtManager), h.userHandler.FollowUser)
	users.Delete("/:id/follow", middleware.Auth(h.jwtManager), h.userHandler.UnfollowUser)

	// Public follower/following lists
	users.Get("/:id/followers", h.userHandler.GetFollowers)
	users.Get("/:id/following", h.userHandler.GetFollowing)
}

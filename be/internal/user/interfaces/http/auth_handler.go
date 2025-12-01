package http

import (
	"github.com/gofiber/fiber/v2"

	"ngasihtau/internal/common/errors"
	"ngasihtau/internal/common/middleware"
	"ngasihtau/internal/common/response"
	"ngasihtau/internal/common/validator"
	"ngasihtau/internal/user/application"
	"ngasihtau/internal/user/domain"
)

// AuthHandler handles authentication-related HTTP requests.
type AuthHandler struct {
	userService application.UserService
}

// NewAuthHandler creates a new AuthHandler.
func NewAuthHandler(userService application.UserService) *AuthHandler {
	return &AuthHandler{
		userService: userService,
	}
}

// RegisterRequest represents the request body for user registration.
type RegisterRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8,password"`
	Name     string `json:"name" validate:"required,min=2,max=100"`
}

// LoginRequest represents the request body for user login.
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// GoogleLoginRequest represents the request body for Google OAuth login.
type GoogleLoginRequest struct {
	Code        string `json:"code" validate:"required"`
	RedirectURI string `json:"redirect_uri" validate:"required,url"`
}

// RefreshTokenRequest represents the request body for token refresh.
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// LogoutRequest represents the request body for logout.
type LogoutRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// Verify2FARequest represents the request body for 2FA verification.
// Code can be either a 6-digit TOTP code or a backup code (format: XXXX-XXXX).
type Verify2FARequest struct {
	Code string `json:"code" validate:"required,min=6,max=10"`
}

// Verify2FALoginRequest represents the request body for 2FA login verification.
// Code can be either a 6-digit TOTP code or a backup code (format: XXXX-XXXX).
type Verify2FALoginRequest struct {
	TempToken string `json:"temp_token" validate:"required"`
	Code      string `json:"code" validate:"required,min=6,max=10"`
}

// PasswordResetRequest represents the request body for password reset request.
type PasswordResetRequest struct {
	Email string `json:"email" validate:"required,email"`
}

// PasswordResetConfirmRequest represents the request body for password reset confirmation.
type PasswordResetConfirmRequest struct {
	Token       string `json:"token" validate:"required"`
	NewPassword string `json:"new_password" validate:"required,min=8,password"`
}

// VerifyEmailRequest represents the request body for email verification.
type VerifyEmailRequest struct {
	Token string `json:"token" validate:"required"`
}

// AuthResponse represents the response for successful authentication.
type AuthResponse struct {
	User         *UserResponse `json:"user"`
	AccessToken  string        `json:"access_token"`
	RefreshToken string        `json:"refresh_token"`
	ExpiresIn    int64         `json:"expires_in"`
}

// TwoFactorSetupResponse represents the response for 2FA setup.
type TwoFactorSetupResponse struct {
	Secret      string   `json:"secret"`
	QRCodeURL   string   `json:"qr_code_url"`
	BackupCodes []string `json:"backup_codes"`
}

// TwoFactorRequiredResponse represents the response when 2FA is required.
type TwoFactorRequiredResponse struct {
	TempToken       string `json:"temp_token"`
	TwoFactorMethod string `json:"two_factor_method"` // "totp" or "backup_code"
}

// toAuthResponse converts an AuthResult to AuthResponse.
func toAuthResponse(result *application.AuthResult) *AuthResponse {
	return &AuthResponse{
		User:         ToUserResponse(result.User, true),
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
		ExpiresIn:    result.ExpiresIn,
	}
}

// Register handles user registration.
// POST /api/v1/auth/register
func (h *AuthHandler) Register(c *fiber.Ctx) error {
	requestID := middleware.GetRequestID(c)

	var req RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return sendError(c, requestID, errors.BadRequest("invalid request body"))
	}

	// Validate request
	if err := validator.Get().Struct(&req); err != nil {
		return sendError(c, requestID, err)
	}

	// Call service
	result, err := h.userService.Register(c.Context(), application.RegisterInput{
		Email:    req.Email,
		Password: req.Password,
		Name:     req.Name,
	})
	if err != nil {
		return sendError(c, requestID, err)
	}

	return c.Status(fiber.StatusCreated).JSON(response.Created(requestID, toAuthResponse(result)))
}

// Login handles user login.
// POST /api/v1/auth/login
func (h *AuthHandler) Login(c *fiber.Ctx) error {
	requestID := middleware.GetRequestID(c)

	var req LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return sendError(c, requestID, errors.BadRequest("invalid request body"))
	}

	// Validate request
	if err := validator.Get().Struct(&req); err != nil {
		return sendError(c, requestID, err)
	}

	// Extract device info from request
	deviceInfo := extractDeviceInfo(c)

	// Call service
	result, err := h.userService.Login(c.Context(), application.LoginInput{
		Email:      req.Email,
		Password:   req.Password,
		DeviceInfo: deviceInfo,
	})
	if err != nil {
		return sendError(c, requestID, err)
	}

	// Check if 2FA is required
	if result.Requires2FA {
		return c.Status(fiber.StatusOK).JSON(response.OK(requestID, TwoFactorRequiredResponse{
			TempToken:       result.TempToken,
			TwoFactorMethod: "totp",
		}))
	}

	return c.Status(fiber.StatusOK).JSON(response.OK(requestID, toAuthResponse(result)))
}

// RefreshToken handles token refresh.
// POST /api/v1/auth/refresh
func (h *AuthHandler) RefreshToken(c *fiber.Ctx) error {
	requestID := middleware.GetRequestID(c)

	var req RefreshTokenRequest
	if err := c.BodyParser(&req); err != nil {
		return sendError(c, requestID, errors.BadRequest("invalid request body"))
	}

	// Validate request
	if err := validator.Get().Struct(&req); err != nil {
		return sendError(c, requestID, err)
	}

	// Call service
	result, err := h.userService.RefreshToken(c.Context(), req.RefreshToken)
	if err != nil {
		return sendError(c, requestID, err)
	}

	return c.Status(fiber.StatusOK).JSON(response.OK(requestID, toAuthResponse(result)))
}

// Logout handles user logout.
// POST /api/v1/auth/logout
func (h *AuthHandler) Logout(c *fiber.Ctx) error {
	requestID := middleware.GetRequestID(c)

	var req LogoutRequest
	if err := c.BodyParser(&req); err != nil {
		return sendError(c, requestID, errors.BadRequest("invalid request body"))
	}

	// Validate request
	if err := validator.Get().Struct(&req); err != nil {
		return sendError(c, requestID, err)
	}

	// Call service
	if err := h.userService.Logout(c.Context(), req.RefreshToken); err != nil {
		return sendError(c, requestID, err)
	}

	return c.Status(fiber.StatusOK).JSON(response.Empty(requestID))
}

// GoogleLogin handles Google OAuth login.
// POST /api/v1/auth/google
func (h *AuthHandler) GoogleLogin(c *fiber.Ctx) error {
	requestID := middleware.GetRequestID(c)

	var req GoogleLoginRequest
	if err := c.BodyParser(&req); err != nil {
		return sendError(c, requestID, errors.BadRequest("invalid request body"))
	}

	// Validate request
	if err := validator.Get().Struct(&req); err != nil {
		return sendError(c, requestID, err)
	}

	// Call service
	result, err := h.userService.LoginWithGoogle(c.Context(), application.GoogleLoginInput{
		Code:        req.Code,
		RedirectURI: req.RedirectURI,
	})
	if err != nil {
		return sendError(c, requestID, err)
	}

	// Check if 2FA is required
	if result.Requires2FA {
		return c.Status(fiber.StatusOK).JSON(response.OK(requestID, TwoFactorRequiredResponse{
			TempToken:       result.TempToken,
			TwoFactorMethod: "totp",
		}))
	}

	return c.Status(fiber.StatusOK).JSON(response.OK(requestID, toAuthResponse(result)))
}

// Enable2FA initiates 2FA setup for the authenticated user.
// POST /api/v1/auth/2fa/enable
func (h *AuthHandler) Enable2FA(c *fiber.Ctx) error {
	requestID := middleware.GetRequestID(c)

	// Get user ID from context (set by auth middleware)
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return sendError(c, requestID, errors.Unauthorized("authentication required"))
	}

	// Call service
	result, err := h.userService.Enable2FA(c.Context(), userID)
	if err != nil {
		return sendError(c, requestID, err)
	}

	return c.Status(fiber.StatusOK).JSON(response.OK(requestID, TwoFactorSetupResponse{
		Secret:      result.Secret,
		QRCodeURL:   result.QRCodeURL,
		BackupCodes: result.BackupCodes,
	}))
}

// Verify2FA completes 2FA setup by verifying a TOTP code.
// POST /api/v1/auth/2fa/verify
func (h *AuthHandler) Verify2FA(c *fiber.Ctx) error {
	requestID := middleware.GetRequestID(c)

	// Get user ID from context (set by auth middleware)
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return sendError(c, requestID, errors.Unauthorized("authentication required"))
	}

	var req Verify2FARequest
	if err := c.BodyParser(&req); err != nil {
		return sendError(c, requestID, errors.BadRequest("invalid request body"))
	}

	// Validate request
	if err := validator.Get().Struct(&req); err != nil {
		return sendError(c, requestID, err)
	}

	// Call service
	if err := h.userService.Verify2FA(c.Context(), userID, req.Code); err != nil {
		return sendError(c, requestID, err)
	}

	return c.Status(fiber.StatusOK).JSON(response.OK(requestID, map[string]bool{
		"two_factor_enabled": true,
	}))
}

// Disable2FA disables 2FA for the authenticated user.
// POST /api/v1/auth/2fa/disable
func (h *AuthHandler) Disable2FA(c *fiber.Ctx) error {
	requestID := middleware.GetRequestID(c)

	// Get user ID from context (set by auth middleware)
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return sendError(c, requestID, errors.Unauthorized("authentication required"))
	}

	var req Verify2FARequest
	if err := c.BodyParser(&req); err != nil {
		return sendError(c, requestID, errors.BadRequest("invalid request body"))
	}

	// Validate request
	if err := validator.Get().Struct(&req); err != nil {
		return sendError(c, requestID, err)
	}

	// Call service
	if err := h.userService.Disable2FA(c.Context(), userID, req.Code); err != nil {
		return sendError(c, requestID, err)
	}

	return c.Status(fiber.StatusOK).JSON(response.OK(requestID, map[string]bool{
		"two_factor_enabled": false,
	}))
}

// Verify2FALogin completes login for users with 2FA enabled.
// POST /api/v1/auth/2fa/login
func (h *AuthHandler) Verify2FALogin(c *fiber.Ctx) error {
	requestID := middleware.GetRequestID(c)

	var req Verify2FALoginRequest
	if err := c.BodyParser(&req); err != nil {
		return sendError(c, requestID, errors.BadRequest("invalid request body"))
	}

	// Validate request
	if err := validator.Get().Struct(&req); err != nil {
		return sendError(c, requestID, err)
	}

	// Call service
	result, err := h.userService.Verify2FALogin(c.Context(), req.TempToken, req.Code)
	if err != nil {
		return sendError(c, requestID, err)
	}

	return c.Status(fiber.StatusOK).JSON(response.OK(requestID, toAuthResponse(result)))
}

// VerifyEmail handles email verification.
// POST /api/v1/auth/verify-email
func (h *AuthHandler) VerifyEmail(c *fiber.Ctx) error {
	requestID := middleware.GetRequestID(c)

	var req VerifyEmailRequest
	if err := c.BodyParser(&req); err != nil {
		return sendError(c, requestID, errors.BadRequest("invalid request body"))
	}

	// Validate request
	if err := validator.Get().Struct(&req); err != nil {
		return sendError(c, requestID, err)
	}

	// Call service
	if err := h.userService.VerifyEmail(c.Context(), req.Token); err != nil {
		return sendError(c, requestID, err)
	}

	return c.Status(fiber.StatusOK).JSON(response.OK(requestID, map[string]bool{
		"email_verified": true,
	}))
}

// RequestPasswordReset handles password reset request.
// POST /api/v1/auth/password/forgot
func (h *AuthHandler) RequestPasswordReset(c *fiber.Ctx) error {
	requestID := middleware.GetRequestID(c)

	var req PasswordResetRequest
	if err := c.BodyParser(&req); err != nil {
		return sendError(c, requestID, errors.BadRequest("invalid request body"))
	}

	// Validate request
	if err := validator.Get().Struct(&req); err != nil {
		return sendError(c, requestID, err)
	}

	// Call service - always returns success to prevent email enumeration
	if err := h.userService.RequestPasswordReset(c.Context(), req.Email); err != nil {
		return sendError(c, requestID, err)
	}

	return c.Status(fiber.StatusOK).JSON(response.OK(requestID, map[string]string{
		"message": "If an account with that email exists, a password reset link has been sent",
	}))
}

// ResetPassword handles password reset confirmation.
// POST /api/v1/auth/password/reset
func (h *AuthHandler) ResetPassword(c *fiber.Ctx) error {
	requestID := middleware.GetRequestID(c)

	var req PasswordResetConfirmRequest
	if err := c.BodyParser(&req); err != nil {
		return sendError(c, requestID, errors.BadRequest("invalid request body"))
	}

	// Validate request
	if err := validator.Get().Struct(&req); err != nil {
		return sendError(c, requestID, err)
	}

	// Call service
	if err := h.userService.ResetPassword(c.Context(), req.Token, req.NewPassword); err != nil {
		return sendError(c, requestID, err)
	}

	return c.Status(fiber.StatusOK).JSON(response.OK(requestID, map[string]bool{
		"password_reset": true,
	}))
}

// SendVerificationEmail handles sending verification email.
// POST /api/v1/auth/send-verification (protected)
func (h *AuthHandler) SendVerificationEmail(c *fiber.Ctx) error {
	requestID := middleware.GetRequestID(c)

	// Get user ID from context (set by auth middleware)
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return sendError(c, requestID, errors.Unauthorized("authentication required"))
	}

	// Call service
	if err := h.userService.SendVerificationEmail(c.Context(), userID); err != nil {
		return sendError(c, requestID, err)
	}

	return c.Status(fiber.StatusOK).JSON(response.OK(requestID, map[string]string{
		"message": "Verification email has been sent",
	}))
}

// extractDeviceInfo extracts device information from the request.
func extractDeviceInfo(c *fiber.Ctx) *domain.DeviceInfo {
	return &domain.DeviceInfo{
		UserAgent: c.Get("User-Agent"),
		IPAddress: c.IP(),
	}
}

// sendError sends an error response using the standard error format.
func sendError(c *fiber.Ctx, requestID string, err error) error {
	resp := errors.BuildResponse(requestID, err)
	status := errors.GetHTTPStatus(err)
	return c.Status(status).JSON(resp)
}

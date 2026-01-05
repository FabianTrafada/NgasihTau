// Package application contains the business logic and use cases for the User Service.
// This layer orchestrates the domain entities and repositories to implement features.
package application

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"ngasihtau/internal/common/errors"
	"ngasihtau/internal/user/domain"
	"ngasihtau/pkg/hash"
	"ngasihtau/pkg/jwt"
	natspkg "ngasihtau/pkg/nats"
	"ngasihtau/pkg/oauth"
	"ngasihtau/pkg/totp"
)

// UserService defines the interface for user-related business operations.
type UserService interface {
	// Authentication operations (Requirement 1)
	Register(ctx context.Context, input RegisterInput) (*AuthResult, error)
	Login(ctx context.Context, input LoginInput) (*AuthResult, error)
	LoginWithGoogle(ctx context.Context, input GoogleLoginInput) (*AuthResult, error)
	RefreshToken(ctx context.Context, refreshToken string) (*AuthResult, error)
	Logout(ctx context.Context, refreshToken string) error
	LogoutAll(ctx context.Context, userID uuid.UUID) error

	// 2FA operations (Requirement 1.1)
	Enable2FA(ctx context.Context, userID uuid.UUID) (*TwoFactorSetup, error)
	Verify2FA(ctx context.Context, userID uuid.UUID, code string) error
	Disable2FA(ctx context.Context, userID uuid.UUID, code string) error
	Verify2FALogin(ctx context.Context, tempToken, code string) (*AuthResult, error)

	// Email verification and password reset (Requirement 1.2)
	SendVerificationEmail(ctx context.Context, userID uuid.UUID) error
	VerifyEmail(ctx context.Context, token string) error
	RequestPasswordReset(ctx context.Context, email string) error
	ResetPassword(ctx context.Context, token, newPassword string) error

	// Profile operations (Requirement 2.1)
	GetProfile(ctx context.Context, userID uuid.UUID) (*domain.UserProfile, error)
	GetCurrentUser(ctx context.Context, userID uuid.UUID) (*domain.User, error)
	UpdateProfile(ctx context.Context, userID uuid.UUID, input UpdateProfileInput) error

	// Follow operations (Requirement 2.2)
	Follow(ctx context.Context, followerID, followingID uuid.UUID) error
	Unfollow(ctx context.Context, followerID, followingID uuid.UUID) error
	GetFollowers(ctx context.Context, userID uuid.UUID, page, perPage int) (*FollowListResult, error)
	GetFollowing(ctx context.Context, userID uuid.UUID, page, perPage int) (*FollowListResult, error)
	IsFollowing(ctx context.Context, followerID, followingID uuid.UUID) (bool, error)

	// Session management (Requirement 1.3)
	GetActiveSessions(ctx context.Context, userID uuid.UUID, currentTokenHash string) ([]*SessionInfo, error)
	RevokeSession(ctx context.Context, userID, sessionID uuid.UUID) error

	// Teacher Verification operations
	SubmitTeacherVerification(ctx context.Context, userID uuid.UUID, input TeacherVerificationInput) (*domain.TeacherVerification, error)
	GetVerificationStatus(ctx context.Context, userID uuid.UUID) (*domain.TeacherVerification, error)
	ApproveVerification(ctx context.Context, verificationID uuid.UUID, reviewerID uuid.UUID) error
	RejectVerification(ctx context.Context, verificationID uuid.UUID, reviewerID uuid.UUID, reason string) error
	GetPendingVerifications(ctx context.Context, page, perPage int) (*VerificationListResult, error)
}

// RegisterInput contains the data required for user registration.
type RegisterInput struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8,password"`
	Name     string `json:"name" validate:"required,min=2,max=100"`
}

// LoginInput contains the data required for user login.
type LoginInput struct {
	Email      string             `json:"email" validate:"required,email"`
	Password   string             `json:"password" validate:"required"`
	DeviceInfo *domain.DeviceInfo `json:"device_info,omitempty"`
}

// GoogleLoginInput contains the data from Google OAuth callback.
type GoogleLoginInput struct {
	Code        string `json:"code" validate:"required"`
	RedirectURI string `json:"redirect_uri" validate:"required,url"`
}

// UpdateProfileInput contains the data for updating a user profile.
type UpdateProfileInput struct {
	Name      *string `json:"name,omitempty" validate:"omitempty,min=2,max=100"`
	Bio       *string `json:"bio,omitempty" validate:"omitempty,max=500"`
	AvatarURL *string `json:"avatar_url,omitempty" validate:"omitempty,url"`
	Language  *string `json:"language,omitempty" validate:"omitempty,oneof=id en"`
}

// AuthResult contains the result of a successful authentication.
type AuthResult struct {
	User         *domain.User `json:"user"`
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	ExpiresIn    int64        `json:"expires_in"` // Access token expiry in seconds
	Requires2FA  bool         `json:"requires_2fa,omitempty"`
	TempToken    string       `json:"temp_token,omitempty"` // Temporary token for 2FA flow
}

// TwoFactorSetup contains the data for setting up 2FA.
type TwoFactorSetup struct {
	Secret      string   `json:"secret"`
	QRCodeURL   string   `json:"qr_code_url"`
	BackupCodes []string `json:"backup_codes"`
}

// FollowListResult contains a paginated list of users.
type FollowListResult struct {
	Users      []*domain.User `json:"users"`
	Total      int            `json:"total"`
	Page       int            `json:"page"`
	PerPage    int            `json:"per_page"`
	TotalPages int            `json:"total_pages"`
}

// SessionInfo contains information about an active session.
type SessionInfo struct {
	ID         uuid.UUID          `json:"id"`
	DeviceInfo *domain.DeviceInfo `json:"device_info,omitempty"`
	CreatedAt  time.Time          `json:"created_at"`
	ExpiresAt  time.Time          `json:"expires_at,omitempty"`
	IsCurrent  bool               `json:"is_current"`
}

// TeacherVerificationInput contains the data required for teacher verification request.
type TeacherVerificationInput struct {
	FullName       string                `json:"full_name" validate:"required,min=3,max=255"`
	IDNumber       string                `json:"id_number" validate:"required,min=10,max=100"`
	CredentialType domain.CredentialType `json:"credential_type" validate:"required,oneof=government_id educator_card professional_cert"`
	DocumentRef    string                `json:"document_ref" validate:"required,max=500"`
}

// VerificationListResult contains a paginated list of teacher verifications.
type VerificationListResult struct {
	Verifications []*domain.TeacherVerification `json:"verifications"`
	Total         int                           `json:"total"`
	Page          int                           `json:"page"`
	PerPage       int                           `json:"per_page"`
	TotalPages    int                           `json:"total_pages"`
}

// userService implements the UserService interface.
type userService struct {
	userRepo                domain.UserRepository
	oauthRepo               domain.OAuthAccountRepository
	refreshTokenRepo        domain.RefreshTokenRepository
	backupCodeRepo          domain.BackupCodeRepository
	followRepo              domain.FollowRepository
	verificationTokenRepo   domain.VerificationTokenRepository
	teacherVerificationRepo domain.TeacherVerificationRepository
	jwtManager              *jwt.Manager
	googleClient            *oauth.GoogleClient
	eventPublisher          natspkg.EventPublisher
}

// NewUserService creates a new UserService instance.
func NewUserService(
	userRepo domain.UserRepository,
	oauthRepo domain.OAuthAccountRepository,
	refreshTokenRepo domain.RefreshTokenRepository,
	backupCodeRepo domain.BackupCodeRepository,
	followRepo domain.FollowRepository,
	verificationTokenRepo domain.VerificationTokenRepository,
	teacherVerificationRepo domain.TeacherVerificationRepository,
	jwtManager *jwt.Manager,
	googleClient *oauth.GoogleClient,
	eventPublisher natspkg.EventPublisher,
) UserService {
	return &userService{
		userRepo:                userRepo,
		oauthRepo:               oauthRepo,
		refreshTokenRepo:        refreshTokenRepo,
		backupCodeRepo:          backupCodeRepo,
		followRepo:              followRepo,
		verificationTokenRepo:   verificationTokenRepo,
		teacherVerificationRepo: teacherVerificationRepo,
		jwtManager:              jwtManager,
		googleClient:            googleClient,
		eventPublisher:          eventPublisher,
	}
}

// Register creates a new user account.
// Implements requirement 1: User Authentication & Authorization.
func (s *userService) Register(ctx context.Context, input RegisterInput) (*AuthResult, error) {
	// Check if email already exists
	exists, err := s.userRepo.ExistsByEmail(ctx, input.Email)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.Conflict("user", input.Email)
	}

	// Hash password using bcrypt (cost 12 as per requirement 10.7)
	passwordHash, err := hash.Password(input.Password)
	if err != nil {
		return nil, errors.Internal("failed to hash password", err)
	}

	// Create user with default role "student"
	user := domain.NewUser(input.Email, passwordHash, input.Name)

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	// Generate JWT tokens
	accessToken, refreshToken, err := s.jwtManager.GenerateTokenPair(user.ID, string(user.Role))
	if err != nil {
		return nil, errors.Internal("failed to generate tokens", err)
	}

	// Hash refresh token for storage (SHA-256)
	tokenHash, err := hash.SHA256(refreshToken)
	if err != nil {
		return nil, errors.Internal("failed to hash refresh token", err)
	}

	// Store refresh token
	refreshTokenEntity := domain.NewRefreshToken(
		user.ID,
		tokenHash,
		nil, // No device info for registration
		s.jwtManager.RefreshTokenExpiry(),
	)
	if err := s.refreshTokenRepo.Create(ctx, refreshTokenEntity); err != nil {
		return nil, err
	}

	// Publish user.created event
	if s.eventPublisher != nil {
		event := natspkg.UserCreatedEvent{
			UserID: user.ID,
			Email:  user.Email,
			Name:   user.Name,
			Role:   string(user.Role),
		}
		if err := s.eventPublisher.PublishUserCreated(ctx, event); err != nil {
			log.Error().Err(err).Str("user_id", user.ID.String()).Msg("failed to publish user created event")
		}
	}

	// Send verification email automatically after registration
	// This is done asynchronously via event publishing
	if err := s.SendVerificationEmail(ctx, user.ID); err != nil {
		// Log but don't fail registration - user can request verification email later
		log.Error().Err(err).Str("user_id", user.ID.String()).Msg("failed to send verification email after registration")
	}

	return &AuthResult{
		User:         user,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(s.jwtManager.GetAccessTokenExpiry().Seconds()),
	}, nil
}

// Login authenticates a user with email and password.
// Implements requirement 1: User Authentication & Authorization.
func (s *userService) Login(ctx context.Context, input LoginInput) (*AuthResult, error) {
	// Find user by email
	user, err := s.userRepo.FindByEmail(ctx, input.Email)
	if err != nil {
		// Return generic error to prevent email enumeration
		return nil, errors.Unauthorized("invalid email or password")
	}

	// Verify password
	if err := hash.VerifyPassword(input.Password, user.PasswordHash); err != nil {
		return nil, errors.Unauthorized("invalid email or password")
	}

	// Check if 2FA is enabled
	if user.TwoFactorEnabled {
		// Generate temporary token for 2FA flow
		tempToken, err := s.jwtManager.GenerateAccessToken(user.ID, string(user.Role))
		if err != nil {
			return nil, errors.Internal("failed to generate temp token", err)
		}
		return &AuthResult{
			Requires2FA: true,
			TempToken:   tempToken,
		}, nil
	}

	// Generate JWT tokens
	accessToken, refreshToken, err := s.jwtManager.GenerateTokenPair(user.ID, string(user.Role))
	if err != nil {
		return nil, errors.Internal("failed to generate tokens", err)
	}

	// Hash refresh token for storage
	tokenHash, err := hash.SHA256(refreshToken)
	if err != nil {
		return nil, errors.Internal("failed to hash refresh token", err)
	}

	// Store refresh token with device info
	refreshTokenEntity := domain.NewRefreshToken(
		user.ID,
		tokenHash,
		input.DeviceInfo,
		s.jwtManager.RefreshTokenExpiry(),
	)
	if err := s.refreshTokenRepo.Create(ctx, refreshTokenEntity); err != nil {
		return nil, err
	}

	return &AuthResult{
		User:         user,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(s.jwtManager.GetAccessTokenExpiry().Seconds()),
	}, nil
}

// LoginWithGoogle authenticates a user via Google OAuth.
// Implements requirement 1: Google OAuth login.
func (s *userService) LoginWithGoogle(ctx context.Context, input GoogleLoginInput) (*AuthResult, error) {
	// Check if Google OAuth is configured
	if s.googleClient == nil {
		return nil, errors.Internal("Google OAuth is not configured", nil)
	}

	// Exchange authorization code for tokens and get user info
	googleUser, err := s.googleClient.ExchangeCode(ctx, input.Code, input.RedirectURI)
	if err != nil {
		return nil, errors.Unauthorized("failed to authenticate with Google")
	}

	// Validate that we got an email
	if googleUser.Email == "" {
		return nil, errors.BadRequest("Google account does not have an email")
	}

	// Check if OAuth account already exists
	oauthAccount, err := s.oauthRepo.FindByProvider(ctx, domain.OAuthProviderGoogle, googleUser.ID)
	if err != nil {
		// Check if it's a "not found" error
		if appErr, ok := err.(*errors.AppError); ok && appErr.Code == errors.CodeNotFound {
			// OAuth account doesn't exist, need to create or link
			return s.handleNewGoogleOAuth(ctx, googleUser)
		}
		return nil, err
	}

	// OAuth account exists, get the user
	user, err := s.userRepo.FindByID(ctx, oauthAccount.UserID)
	if err != nil {
		return nil, err
	}

	// Check if 2FA is enabled
	if user.TwoFactorEnabled {
		// Generate temporary token for 2FA flow
		tempToken, err := s.jwtManager.GenerateAccessToken(user.ID, string(user.Role))
		if err != nil {
			return nil, errors.Internal("failed to generate temp token", err)
		}
		return &AuthResult{
			Requires2FA: true,
			TempToken:   tempToken,
		}, nil
	}

	// Generate JWT tokens
	return s.generateAuthResult(ctx, user, nil)
}

// handleNewGoogleOAuth handles the case where a Google OAuth account doesn't exist yet.
// It either links to an existing user with the same email or creates a new user.
func (s *userService) handleNewGoogleOAuth(ctx context.Context, googleUser *oauth.GoogleUserInfo) (*AuthResult, error) {
	// Check if a user with this email already exists
	existingUser, err := s.userRepo.FindByEmail(ctx, googleUser.Email)
	if err != nil {
		// Check if it's a "not found" error
		if appErr, ok := err.(*errors.AppError); ok && appErr.Code == errors.CodeNotFound {
			// No existing user, create a new one
			return s.createUserFromGoogle(ctx, googleUser)
		}
		return nil, err
	}

	// User exists with this email, link the OAuth account
	oauthAccount := domain.NewOAuthAccount(existingUser.ID, domain.OAuthProviderGoogle, googleUser.ID)
	if err := s.oauthRepo.Create(ctx, oauthAccount); err != nil {
		return nil, err
	}

	// If user's email wasn't verified but Google's is, mark it as verified
	if !existingUser.EmailVerified && googleUser.VerifiedEmail {
		if err := s.userRepo.VerifyEmail(ctx, existingUser.ID); err != nil {
			// Log but don't fail - the OAuth link was successful
		}
		existingUser.EmailVerified = true
	}

	// Check if 2FA is enabled
	if existingUser.TwoFactorEnabled {
		tempToken, err := s.jwtManager.GenerateAccessToken(existingUser.ID, string(existingUser.Role))
		if err != nil {
			return nil, errors.Internal("failed to generate temp token", err)
		}
		return &AuthResult{
			Requires2FA: true,
			TempToken:   tempToken,
		}, nil
	}

	return s.generateAuthResult(ctx, existingUser, nil)
}

// createUserFromGoogle creates a new user from Google OAuth information.
func (s *userService) createUserFromGoogle(ctx context.Context, googleUser *oauth.GoogleUserInfo) (*AuthResult, error) {
	// Create user without password (OAuth-only user)
	user := &domain.User{
		ID:            uuid.New(),
		Email:         googleUser.Email,
		PasswordHash:  "", // No password for OAuth-only users
		Name:          googleUser.Name,
		Role:          domain.RoleStudent,
		EmailVerified: googleUser.VerifiedEmail,
		Language:      "id",
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	// Set avatar if provided
	if googleUser.Picture != "" {
		user.AvatarURL = &googleUser.Picture
	}

	// Create the user
	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	// Create OAuth account link
	oauthAccount := domain.NewOAuthAccount(user.ID, domain.OAuthProviderGoogle, googleUser.ID)
	if err := s.oauthRepo.Create(ctx, oauthAccount); err != nil {
		// Rollback user creation would be ideal here, but for simplicity we'll just return the error
		return nil, err
	}

	return s.generateAuthResult(ctx, user, nil)
}

// generateAuthResult generates JWT tokens and creates a refresh token for the user.
func (s *userService) generateAuthResult(ctx context.Context, user *domain.User, deviceInfo *domain.DeviceInfo) (*AuthResult, error) {
	// Generate JWT tokens
	accessToken, refreshToken, err := s.jwtManager.GenerateTokenPair(user.ID, string(user.Role))
	if err != nil {
		return nil, errors.Internal("failed to generate tokens", err)
	}

	// Hash refresh token for storage
	tokenHash, err := hash.SHA256(refreshToken)
	if err != nil {
		return nil, errors.Internal("failed to hash refresh token", err)
	}

	// Store refresh token
	refreshTokenEntity := domain.NewRefreshToken(
		user.ID,
		tokenHash,
		deviceInfo,
		s.jwtManager.RefreshTokenExpiry(),
	)
	if err := s.refreshTokenRepo.Create(ctx, refreshTokenEntity); err != nil {
		return nil, err
	}

	return &AuthResult{
		User:         user,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(s.jwtManager.GetAccessTokenExpiry().Seconds()),
	}, nil
}

// RefreshToken generates new tokens using a valid refresh token.
// Implements requirement 1.3: Session Management & Refresh Tokens.
func (s *userService) RefreshToken(ctx context.Context, refreshToken string) (*AuthResult, error) {
	// Hash the refresh token to find it in database
	tokenHash, err := hash.SHA256(refreshToken)
	if err != nil {
		return nil, errors.Internal("failed to hash refresh token", err)
	}

	// Find refresh token in database
	storedToken, err := s.refreshTokenRepo.FindByTokenHash(ctx, tokenHash)
	if err != nil {
		return nil, errors.Unauthorized("invalid refresh token")
	}

	// Check if token is expired
	if storedToken.IsExpired() {
		// Delete expired token
		_ = s.refreshTokenRepo.Delete(ctx, storedToken.ID)
		return nil, errors.Unauthorized("refresh token expired")
	}

	// Get user
	user, err := s.userRepo.FindByID(ctx, storedToken.UserID)
	if err != nil {
		return nil, err
	}

	// Delete old refresh token (token rotation)
	if err := s.refreshTokenRepo.Delete(ctx, storedToken.ID); err != nil {
		return nil, err
	}

	// Generate new tokens
	newAccessToken, newRefreshToken, err := s.jwtManager.GenerateTokenPair(user.ID, string(user.Role))
	if err != nil {
		return nil, errors.Internal("failed to generate tokens", err)
	}

	// Hash new refresh token for storage
	newTokenHash, err := hash.SHA256(newRefreshToken)
	if err != nil {
		return nil, errors.Internal("failed to hash refresh token", err)
	}

	// Store new refresh token with same device info
	newRefreshTokenEntity := domain.NewRefreshToken(
		user.ID,
		newTokenHash,
		storedToken.DeviceInfo,
		s.jwtManager.RefreshTokenExpiry(),
	)
	if err := s.refreshTokenRepo.Create(ctx, newRefreshTokenEntity); err != nil {
		return nil, err
	}

	return &AuthResult{
		User:         user,
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
		ExpiresIn:    int64(s.jwtManager.GetAccessTokenExpiry().Seconds()),
	}, nil
}

// Logout invalidates the current refresh token.
// Implements requirement 1.3: Session Management & Refresh Tokens.
func (s *userService) Logout(ctx context.Context, refreshToken string) error {
	// Hash the refresh token
	tokenHash, err := hash.SHA256(refreshToken)
	if err != nil {
		return errors.Internal("failed to hash refresh token", err)
	}

	// Delete from database
	if err := s.refreshTokenRepo.DeleteByTokenHash(ctx, tokenHash); err != nil {
		// Ignore not found errors - token may already be deleted
		if appErr, ok := err.(*errors.AppError); ok && appErr.Code == errors.CodeNotFound {
			return nil
		}
		return err
	}

	return nil
}

// LogoutAll invalidates all refresh tokens for a user.
// Implements requirement 1.3: Session Management & Refresh Tokens.
func (s *userService) LogoutAll(ctx context.Context, userID uuid.UUID) error {
	return s.refreshTokenRepo.DeleteAllByUserID(ctx, userID)
}

// Enable2FA initiates 2FA setup for a user.
// Implements requirement 1.1: Two-Factor Authentication.
// This returns the secret and QR code URL for the user to set up their authenticator app.
// The user must call Verify2FA with a valid code to complete the setup.
func (s *userService) Enable2FA(ctx context.Context, userID uuid.UUID) (*TwoFactorSetup, error) {
	// Get user to check if 2FA is already enabled
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if user.TwoFactorEnabled {
		return nil, errors.BadRequest("2FA is already enabled")
	}

	// Generate TOTP secret
	secret, qrCodeURL, err := totp.GenerateSecret(user.Email)
	if err != nil {
		return nil, errors.Internal("failed to generate TOTP secret", err)
	}

	// Generate backup codes (shown to user once, will be stored after verification)
	backupCodes, err := totp.GenerateBackupCodes()
	if err != nil {
		return nil, errors.Internal("failed to generate backup codes", err)
	}

	// Store the pending secret in the user record (two_factor_enabled stays false)
	// The secret will be "activated" when the user verifies with a valid code
	user.TwoFactorSecret = &secret
	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, err
	}

	return &TwoFactorSetup{
		Secret:      secret,
		QRCodeURL:   qrCodeURL,
		BackupCodes: backupCodes,
	}, nil
}

// Verify2FA completes 2FA setup by verifying a TOTP code.
// Implements requirement 1.1: Two-Factor Authentication.
func (s *userService) Verify2FA(ctx context.Context, userID uuid.UUID, code string) error {
	// Get user to retrieve the pending secret
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return err
	}

	if user.TwoFactorEnabled {
		return errors.BadRequest("2FA is already enabled")
	}

	if user.TwoFactorSecret == nil || *user.TwoFactorSecret == "" {
		return errors.BadRequest("2FA setup not initiated, please call enable first")
	}

	// Validate the TOTP code
	if !totp.ValidateCode(*user.TwoFactorSecret, code) {
		return errors.Unauthorized("invalid 2FA code")
	}

	// Generate backup codes
	backupCodes, err := totp.GenerateBackupCodes()
	if err != nil {
		return errors.Internal("failed to generate backup codes", err)
	}

	// Delete any existing backup codes
	if err := s.backupCodeRepo.DeleteAllByUserID(ctx, userID); err != nil {
		return err
	}

	// Hash and store backup codes
	backupCodeEntities := make([]*domain.BackupCode, len(backupCodes))
	for i, code := range backupCodes {
		// Normalize and hash the backup code
		normalizedCode := totp.NormalizeCode(code)
		codeHash, err := hash.Password(normalizedCode)
		if err != nil {
			return errors.Internal("failed to hash backup code", err)
		}
		backupCodeEntities[i] = domain.NewBackupCode(userID, codeHash)
	}

	if err := s.backupCodeRepo.CreateBatch(ctx, backupCodeEntities); err != nil {
		return err
	}

	// Enable 2FA
	if err := s.userRepo.Enable2FA(ctx, userID, *user.TwoFactorSecret); err != nil {
		return err
	}

	return nil
}

// Disable2FA disables 2FA for a user.
// Implements requirement 1.1: Two-Factor Authentication.
func (s *userService) Disable2FA(ctx context.Context, userID uuid.UUID, code string) error {
	// Get user
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return err
	}

	if !user.TwoFactorEnabled {
		return errors.BadRequest("2FA is not enabled")
	}

	if user.TwoFactorSecret == nil || *user.TwoFactorSecret == "" {
		return errors.Internal("2FA secret not found", nil)
	}

	// Validate the TOTP code
	if !totp.ValidateCode(*user.TwoFactorSecret, code) {
		return errors.Unauthorized("invalid 2FA code")
	}

	// Delete all backup codes
	if err := s.backupCodeRepo.DeleteAllByUserID(ctx, userID); err != nil {
		return err
	}

	// Disable 2FA
	if err := s.userRepo.Disable2FA(ctx, userID); err != nil {
		return err
	}

	return nil
}

// Verify2FALogin completes login for users with 2FA enabled.
// Implements requirement 1.1: Two-Factor Authentication.
func (s *userService) Verify2FALogin(ctx context.Context, tempToken, code string) (*AuthResult, error) {
	// Validate the temporary token (use access token validation since temp token is an access token)
	claims, err := s.jwtManager.ValidateAccessToken(tempToken)
	if err != nil {
		return nil, errors.Unauthorized("invalid or expired temporary token")
	}

	userID := claims.UserID

	// Get user
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if !user.TwoFactorEnabled {
		return nil, errors.BadRequest("2FA is not enabled for this account")
	}

	if user.TwoFactorSecret == nil || *user.TwoFactorSecret == "" {
		return nil, errors.Internal("2FA secret not found", nil)
	}

	// Try TOTP code first
	codeValid := totp.ValidateCode(*user.TwoFactorSecret, code)

	// If TOTP code is invalid, try backup codes
	if !codeValid {
		normalizedCode := totp.NormalizeCode(code)
		backupCodes, err := s.backupCodeRepo.FindUnusedByUserID(ctx, userID)
		if err != nil {
			return nil, err
		}

		for _, backupCode := range backupCodes {
			if err := hash.VerifyPassword(normalizedCode, backupCode.CodeHash); err == nil {
				// Backup code matches, mark it as used
				if err := s.backupCodeRepo.MarkAsUsed(ctx, backupCode.ID); err != nil {
					return nil, err
				}
				codeValid = true
				break
			}
		}
	}

	if !codeValid {
		return nil, errors.Unauthorized("invalid 2FA code")
	}

	// Generate full tokens
	return s.generateAuthResult(ctx, user, nil)
}

// SendVerificationEmail sends an email verification link.
// Implements requirement 1.2: Email Verification & Password Reset.
func (s *userService) SendVerificationEmail(ctx context.Context, userID uuid.UUID) error {
	// Get user
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return err
	}

	// Check if email is already verified
	if user.EmailVerified {
		return errors.BadRequest("email is already verified")
	}

	// Delete any existing verification tokens for this user
	if err := s.verificationTokenRepo.DeleteByUserIDAndType(ctx, userID, domain.TokenTypeEmailVerification); err != nil {
		return err
	}

	// Generate a secure random token
	rawToken := uuid.New().String()
	tokenHash, err := hash.SHA256(rawToken)
	if err != nil {
		return errors.Internal("failed to hash verification token", err)
	}

	// Create verification token (valid for 24 hours as per requirement 1.2)
	expiresAt := time.Now().Add(24 * time.Hour)
	verificationToken := domain.NewVerificationToken(
		userID,
		tokenHash,
		domain.TokenTypeEmailVerification,
		expiresAt,
	)

	if err := s.verificationTokenRepo.Create(ctx, verificationToken); err != nil {
		return err
	}

	// Publish email.verification event via NATS for email sending
	// The Notification Service will consume this event and send the verification email
	if s.eventPublisher != nil {
		event := natspkg.EmailVerificationEvent{
			UserID:    userID,
			Email:     user.Email,
			Name:      user.Name,
			Token:     rawToken,
			ExpiresAt: expiresAt.UTC().Format(time.RFC3339),
		}
		if err := s.eventPublisher.PublishEmailVerification(ctx, event); err != nil {
			// Log the error but don't fail the request - the token is still created
			log.Error().Err(err).Str("user_id", userID.String()).Msg("failed to publish email verification event")
		}
	}

	return nil
}

// VerifyEmail verifies a user's email address.
// Implements requirement 1.2: Email Verification & Password Reset.
func (s *userService) VerifyEmail(ctx context.Context, token string) error {
	// Hash the token to find it in database
	tokenHash, err := hash.SHA256(token)
	if err != nil {
		return errors.Internal("failed to hash token", err)
	}

	// Find the verification token
	verificationToken, err := s.verificationTokenRepo.FindByTokenHash(ctx, tokenHash)
	if err != nil {
		return errors.BadRequest("invalid or expired verification token")
	}

	// Check token type
	if verificationToken.TokenType != domain.TokenTypeEmailVerification {
		return errors.BadRequest("invalid token type")
	}

	// Check if token is expired
	if verificationToken.IsExpired() {
		return errors.BadRequest("verification token has expired")
	}

	// Check if token is already used
	if verificationToken.IsUsed() {
		return errors.BadRequest("verification token has already been used")
	}

	// Mark token as used
	if err := s.verificationTokenRepo.MarkAsUsed(ctx, verificationToken.ID); err != nil {
		return err
	}

	// Mark email as verified
	if err := s.userRepo.VerifyEmail(ctx, verificationToken.UserID); err != nil {
		return err
	}

	return nil
}

// RequestPasswordReset sends a password reset email.
// Implements requirement 1.2: Email Verification & Password Reset.
func (s *userService) RequestPasswordReset(ctx context.Context, email string) error {
	// Find user by email
	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		// Return success even if user not found to prevent email enumeration
		if appErr, ok := err.(*errors.AppError); ok && appErr.Code == errors.CodeNotFound {
			return nil
		}
		return err
	}

	// Delete any existing password reset tokens for this user
	if err := s.verificationTokenRepo.DeleteByUserIDAndType(ctx, user.ID, domain.TokenTypePasswordReset); err != nil {
		return err
	}

	// Generate a secure random token
	rawToken := uuid.New().String()
	tokenHash, err := hash.SHA256(rawToken)
	if err != nil {
		return errors.Internal("failed to hash reset token", err)
	}

	// Create password reset token (valid for 1 hour as per requirement 1.2)
	expiresAt := time.Now().Add(1 * time.Hour)
	resetToken := domain.NewVerificationToken(
		user.ID,
		tokenHash,
		domain.TokenTypePasswordReset,
		expiresAt,
	)

	if err := s.verificationTokenRepo.Create(ctx, resetToken); err != nil {
		return err
	}

	// Publish email.password_reset event via NATS for email sending
	// The Notification Service will consume this event and send the password reset email
	if s.eventPublisher != nil {
		event := natspkg.EmailPasswordResetEvent{
			UserID:    user.ID,
			Email:     user.Email,
			Name:      user.Name,
			Token:     rawToken,
			ExpiresAt: expiresAt.UTC().Format(time.RFC3339),
		}
		if err := s.eventPublisher.PublishEmailPasswordReset(ctx, event); err != nil {
			// Log the error but don't fail the request - the token is still created
			log.Error().Err(err).Str("user_id", user.ID.String()).Msg("failed to publish password reset event")
		}
	}

	return nil
}

// ResetPassword resets a user's password using a reset token.
// Implements requirement 1.2: Email Verification & Password Reset.
func (s *userService) ResetPassword(ctx context.Context, token, newPassword string) error {
	// Hash the token to find it in database
	tokenHash, err := hash.SHA256(token)
	if err != nil {
		return errors.Internal("failed to hash token", err)
	}

	// Find the reset token
	resetToken, err := s.verificationTokenRepo.FindByTokenHash(ctx, tokenHash)
	if err != nil {
		return errors.BadRequest("invalid or expired reset token")
	}

	// Check token type
	if resetToken.TokenType != domain.TokenTypePasswordReset {
		return errors.BadRequest("invalid token type")
	}

	// Check if token is expired
	if resetToken.IsExpired() {
		return errors.BadRequest("reset token has expired")
	}

	// Check if token is already used
	if resetToken.IsUsed() {
		return errors.BadRequest("reset token has already been used")
	}

	// Hash the new password
	passwordHash, err := hash.Password(newPassword)
	if err != nil {
		return errors.Internal("failed to hash password", err)
	}

	// Mark token as used
	if err := s.verificationTokenRepo.MarkAsUsed(ctx, resetToken.ID); err != nil {
		return err
	}

	// Update the password
	if err := s.userRepo.UpdatePassword(ctx, resetToken.UserID, passwordHash); err != nil {
		return err
	}

	// Invalidate all refresh tokens for security
	if err := s.refreshTokenRepo.DeleteAllByUserID(ctx, resetToken.UserID); err != nil {
		// Log but don't fail - password was already updated
		_ = err
	}

	return nil
}

// GetProfile retrieves a user's public profile.
// Implements requirement 2.1: User Profiles.
func (s *userService) GetProfile(ctx context.Context, userID uuid.UUID) (*domain.UserProfile, error) {
	return s.userRepo.GetProfile(ctx, userID)
}

// GetCurrentUser retrieves the current user's full information.
// Implements requirement 2.1: User Profiles.
func (s *userService) GetCurrentUser(ctx context.Context, userID uuid.UUID) (*domain.User, error) {
	return s.userRepo.FindByID(ctx, userID)
}

// UpdateProfile updates a user's profile information.
// Implements requirement 2.1: User Profiles.
func (s *userService) UpdateProfile(ctx context.Context, userID uuid.UUID, input UpdateProfileInput) error {
	// Get current user
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return err
	}

	// Apply updates
	if input.Name != nil {
		user.Name = *input.Name
	}
	if input.Bio != nil {
		user.Bio = input.Bio
	}
	if input.AvatarURL != nil {
		user.AvatarURL = input.AvatarURL
	}
	if input.Language != nil {
		user.Language = *input.Language
	}

	// Save changes
	if err := s.userRepo.Update(ctx, user); err != nil {
		return err
	}

	// Publish user.updated event via NATS for cache invalidation
	if s.eventPublisher != nil {
		if err := s.eventPublisher.PublishUserUpdated(ctx, userID); err != nil {
			log.Error().Err(err).Str("user_id", userID.String()).Msg("failed to publish user updated event")
		}
	}

	return nil
}

// Follow creates a follow relationship.
// Implements requirement 2.2: Follow Users.
func (s *userService) Follow(ctx context.Context, followerID, followingID uuid.UUID) error {
	// Cannot follow yourself
	if followerID == followingID {
		return errors.BadRequest("cannot follow yourself")
	}

	// Check if target user exists
	_, err := s.userRepo.FindByID(ctx, followingID)
	if err != nil {
		return err
	}

	// Check if already following
	exists, err := s.followRepo.Exists(ctx, followerID, followingID)
	if err != nil {
		return err
	}
	if exists {
		return errors.ConflictMsg("already following this user")
	}

	// Create follow relationship
	follow := domain.NewFollow(followerID, followingID)
	if err := s.followRepo.Create(ctx, follow); err != nil {
		return err
	}

	// Publish user.followed event via NATS for notifications
	if s.eventPublisher != nil {
		if err := s.eventPublisher.PublishUserFollowed(ctx, followerID, followingID); err != nil {
			log.Error().Err(err).
				Str("follower_id", followerID.String()).
				Str("following_id", followingID.String()).
				Msg("failed to publish user followed event")
		}
	}

	return nil
}

// Unfollow removes a follow relationship.
// Implements requirement 2.2: Follow Users.
func (s *userService) Unfollow(ctx context.Context, followerID, followingID uuid.UUID) error {
	return s.followRepo.Delete(ctx, followerID, followingID)
}

// GetFollowers retrieves a paginated list of followers.
// Implements requirement 2.2: Follow Users.
func (s *userService) GetFollowers(ctx context.Context, userID uuid.UUID, page, perPage int) (*FollowListResult, error) {
	// Validate pagination
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}

	offset := (page - 1) * perPage
	users, total, err := s.followRepo.GetFollowers(ctx, userID, perPage, offset)
	if err != nil {
		return nil, err
	}

	totalPages := (total + perPage - 1) / perPage

	return &FollowListResult{
		Users:      users,
		Total:      total,
		Page:       page,
		PerPage:    perPage,
		TotalPages: totalPages,
	}, nil
}

// GetFollowing retrieves a paginated list of users being followed.
// Implements requirement 2.2: Follow Users.
func (s *userService) GetFollowing(ctx context.Context, userID uuid.UUID, page, perPage int) (*FollowListResult, error) {
	// Validate pagination
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}

	offset := (page - 1) * perPage
	users, total, err := s.followRepo.GetFollowing(ctx, userID, perPage, offset)
	if err != nil {
		return nil, err
	}

	totalPages := (total + perPage - 1) / perPage

	return &FollowListResult{
		Users:      users,
		Total:      total,
		Page:       page,
		PerPage:    perPage,
		TotalPages: totalPages,
	}, nil
}

// IsFollowing checks if a user is following another user.
// Implements requirement 2.2: Follow Users.
func (s *userService) IsFollowing(ctx context.Context, followerID, followingID uuid.UUID) (bool, error) {
	return s.followRepo.Exists(ctx, followerID, followingID)
}

// GetActiveSessions retrieves all active sessions for a user.
// Implements requirement 1.3: Session Management & Refresh Tokens.
func (s *userService) GetActiveSessions(ctx context.Context, userID uuid.UUID, currentTokenHash string) ([]*SessionInfo, error) {
	sessions, err := s.refreshTokenRepo.GetActiveSessions(ctx, userID, currentTokenHash)
	if err != nil {
		return nil, err
	}

	result := make([]*SessionInfo, len(sessions))
	for i, session := range sessions {
		result[i] = &SessionInfo{
			ID:         session.ID,
			DeviceInfo: session.DeviceInfo,
			CreatedAt:  session.CreatedAt,
			IsCurrent:  session.IsCurrent,
		}
	}

	return result, nil
}

// RevokeSession revokes a specific session.
// Implements requirement 1.3: Session Management & Refresh Tokens.
func (s *userService) RevokeSession(ctx context.Context, userID, sessionID uuid.UUID) error {
	// Verify the session belongs to the user
	tokens, err := s.refreshTokenRepo.FindByUserID(ctx, userID)
	if err != nil {
		return err
	}

	for _, token := range tokens {
		if token.ID == sessionID {
			return s.refreshTokenRepo.Delete(ctx, sessionID)
		}
	}

	return errors.NotFound("session", sessionID.String())
}

// SubmitTeacherVerification submits a teacher verification request.
func (s *userService) SubmitTeacherVerification(ctx context.Context, userID uuid.UUID, input TeacherVerificationInput) (*domain.TeacherVerification, error) {
	// Validate input fields (Requirement 3.3)
	if err := s.validateTeacherVerificationInput(input); err != nil {
		return nil, err
	}

	// Check if user exists
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Check if user is already a teacher
	if user.Role == domain.RoleTeacher {
		return nil, errors.BadRequest("user is already a verified teacher")
	}

	// Check if a pending verification already exists for this user
	existingPending, err := s.teacherVerificationRepo.ExistsPendingByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if existingPending {
		return nil, errors.ConflictMsg("a pending verification request already exists for this user")
	}

	// Create new teacher verification with pending status (Requirement 2.1)
	verification := domain.NewTeacherVerification(
		userID,
		input.FullName,
		input.IDNumber,
		input.CredentialType,
		input.DocumentRef,
	)

	// Save verification to database
	if err := s.teacherVerificationRepo.Create(ctx, verification); err != nil {
		return nil, err
	}

	log.Info().
		Str("user_id", userID.String()).
		Str("verification_id", verification.ID.String()).
		Str("credential_type", string(input.CredentialType)).
		Msg("teacher verification request submitted")

	return verification, nil
}

// validateTeacherVerificationInput validates the teacher verification input fields.
// Implements requirement 3.3: Validation of required fields.
func (s *userService) validateTeacherVerificationInput(input TeacherVerificationInput) error {
	var details []errors.ErrorDetail

	// Validate full_name (required, min 3, max 255)
	if input.FullName == "" {
		details = append(details, errors.ErrorDetail{Field: "full_name", Message: "full name is required"})
	} else if len(input.FullName) < 3 {
		details = append(details, errors.ErrorDetail{Field: "full_name", Message: "full name must be at least 3 characters"})
	} else if len(input.FullName) > 255 {
		details = append(details, errors.ErrorDetail{Field: "full_name", Message: "full name must not exceed 255 characters"})
	}

	// Validate id_number (required, min 10, max 100)
	if input.IDNumber == "" {
		details = append(details, errors.ErrorDetail{Field: "id_number", Message: "ID number is required"})
	} else if len(input.IDNumber) < 10 {
		details = append(details, errors.ErrorDetail{Field: "id_number", Message: "ID number must be at least 10 characters"})
	} else if len(input.IDNumber) > 100 {
		details = append(details, errors.ErrorDetail{Field: "id_number", Message: "ID number must not exceed 100 characters"})
	}

	// Validate credential_type (required, must be valid type)
	if input.CredentialType == "" {
		details = append(details, errors.ErrorDetail{Field: "credential_type", Message: "credential type is required"})
	} else if !domain.IsValidCredentialType(input.CredentialType) {
		details = append(details, errors.ErrorDetail{Field: "credential_type", Message: "credential type must be one of: government_id, educator_card, professional_cert"})
	}

	// Validate document_ref (required, max 500)
	if input.DocumentRef == "" {
		details = append(details, errors.ErrorDetail{Field: "document_ref", Message: "document reference is required"})
	} else if len(input.DocumentRef) > 500 {
		details = append(details, errors.ErrorDetail{Field: "document_ref", Message: "document reference must not exceed 500 characters"})
	}

	if len(details) > 0 {
		return errors.Validation("validation failed", details...)
	}

	return nil
}

// GetVerificationStatus retrieves the current user's verification status.
// Implements requirement 2.1: Teacher verification status check.
func (s *userService) GetVerificationStatus(ctx context.Context, userID uuid.UUID) (*domain.TeacherVerification, error) {
	return s.teacherVerificationRepo.FindByUserID(ctx, userID)
}

// ApproveVerification approves a teacher verification request.
// Implements requirement 2.2: Teacher verification approval.
// This method updates the verification status to approved and changes the user's role from student to teacher.
func (s *userService) ApproveVerification(ctx context.Context, verificationID uuid.UUID, reviewerID uuid.UUID) error {
	// Find the verification request
	verification, err := s.teacherVerificationRepo.FindByID(ctx, verificationID)
	if err != nil {
		return err
	}

	// Check if verification is still pending
	if !verification.IsPending() {
		return errors.BadRequest("verification request has already been reviewed")
	}

	// Get the user to update their role
	user, err := s.userRepo.FindByID(ctx, verification.UserID)
	if err != nil {
		return err
	}

	// Check if user is already a teacher (shouldn't happen, but defensive check)
	if user.Role == domain.RoleTeacher {
		return errors.BadRequest("user is already a teacher")
	}

	// Update verification status to approved
	verification.Approve(reviewerID)
	if err := s.teacherVerificationRepo.Update(ctx, verification); err != nil {
		return err
	}

	// Update user role from student to teacher (Requirement 2.2)
	user.Role = domain.RoleTeacher
	user.UpdatedAt = time.Now()
	if err := s.userRepo.Update(ctx, user); err != nil {
		return err
	}

	log.Info().
		Str("verification_id", verificationID.String()).
		Str("user_id", verification.UserID.String()).
		Str("reviewer_id", reviewerID.String()).
		Msg("teacher verification approved, user role updated to teacher")

	// Publish events for cache invalidation and notification
	if s.eventPublisher != nil {
		// Publish user.updated event for cache invalidation
		if err := s.eventPublisher.PublishUserUpdated(ctx, verification.UserID); err != nil {
			log.Error().Err(err).Str("user_id", verification.UserID.String()).Msg("failed to publish user updated event")
		}

		// Publish teacher verified event for notification
		event := natspkg.TeacherVerifiedEvent{
			UserID:         verification.UserID,
			VerificationID: verification.ID,
			FullName:       verification.FullName,
			CredentialType: string(verification.CredentialType),
		}
		if err := s.eventPublisher.PublishTeacherVerified(ctx, event); err != nil {
			log.Error().Err(err).Str("user_id", verification.UserID.String()).Msg("failed to publish teacher verified event")
		}
	}

	return nil
}

// RejectVerification rejects a teacher verification request.
// Implements requirement 2.2: Teacher verification rejection.
// This method updates the verification status to rejected with a reason.
// The user's role remains unchanged (stays as student).
func (s *userService) RejectVerification(ctx context.Context, verificationID uuid.UUID, reviewerID uuid.UUID, reason string) error {
	// Find the verification request
	verification, err := s.teacherVerificationRepo.FindByID(ctx, verificationID)
	if err != nil {
		return err
	}

	// Check if verification is still pending
	if !verification.IsPending() {
		return errors.BadRequest("verification request has already been reviewed")
	}

	// Validate that a reason is provided
	if reason == "" {
		return errors.BadRequest("rejection reason is required")
	}

	// Update verification status to rejected with reason
	verification.Reject(reviewerID, reason)
	if err := s.teacherVerificationRepo.Update(ctx, verification); err != nil {
		return err
	}

	log.Info().
		Str("verification_id", verificationID.String()).
		Str("user_id", verification.UserID.String()).
		Str("reviewer_id", reviewerID.String()).
		Str("reason", reason).
		Msg("teacher verification rejected")

	return nil
}

// GetPendingVerifications retrieves paginated pending verification requests.
// Used by admins to review pending verifications.
func (s *userService) GetPendingVerifications(ctx context.Context, page, perPage int) (*VerificationListResult, error) {
	// Validate pagination
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}

	offset := (page - 1) * perPage
	verifications, total, err := s.teacherVerificationRepo.FindPending(ctx, perPage, offset)
	if err != nil {
		return nil, err
	}

	totalPages := (total + perPage - 1) / perPage

	return &VerificationListResult{
		Verifications: verifications,
		Total:         total,
		Page:          page,
		PerPage:       perPage,
		TotalPages:    totalPages,
	}, nil
}

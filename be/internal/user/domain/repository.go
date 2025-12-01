package domain

import (
	"context"

	"github.com/google/uuid"
)

// UserRepository defines the interface for user data access.
// Implements the Repository pattern for data access abstraction (requirement 10.3).
type UserRepository interface {
	// Create creates a new user.
	Create(ctx context.Context, user *User) error

	// FindByID finds a user by ID.
	FindByID(ctx context.Context, id uuid.UUID) (*User, error)

	// FindByEmail finds a user by email address.
	FindByEmail(ctx context.Context, email string) (*User, error)

	// Update updates an existing user.
	Update(ctx context.Context, user *User) error

	// Delete soft-deletes a user.
	Delete(ctx context.Context, id uuid.UUID) error

	// ExistsByEmail checks if a user with the given email exists.
	ExistsByEmail(ctx context.Context, email string) (bool, error)

	// GetProfile retrieves a user's public profile with stats.
	GetProfile(ctx context.Context, id uuid.UUID) (*UserProfile, error)

	// UpdateProfile updates a user's profile information.
	UpdateProfile(ctx context.Context, id uuid.UUID, name string, bio *string, avatarURL *string) error

	// Enable2FA enables two-factor authentication for a user.
	Enable2FA(ctx context.Context, id uuid.UUID, secret string) error

	// Disable2FA disables two-factor authentication for a user.
	Disable2FA(ctx context.Context, id uuid.UUID) error

	// UpdatePassword updates a user's password hash.
	UpdatePassword(ctx context.Context, id uuid.UUID, passwordHash string) error

	// VerifyEmail marks a user's email as verified.
	VerifyEmail(ctx context.Context, id uuid.UUID) error
}

// OAuthAccountRepository defines the interface for OAuth account data access.
type OAuthAccountRepository interface {
	// Create creates a new OAuth account link.
	Create(ctx context.Context, account *OAuthAccount) error

	// FindByProvider finds an OAuth account by provider and provider user ID.
	FindByProvider(ctx context.Context, provider OAuthProvider, providerUserID string) (*OAuthAccount, error)

	// FindByUserID finds all OAuth accounts for a user.
	FindByUserID(ctx context.Context, userID uuid.UUID) ([]*OAuthAccount, error)

	// Delete removes an OAuth account link.
	Delete(ctx context.Context, id uuid.UUID) error
}

// RefreshTokenRepository defines the interface for refresh token data access.
type RefreshTokenRepository interface {
	// Create creates a new refresh token.
	Create(ctx context.Context, token *RefreshToken) error

	// FindByTokenHash finds a refresh token by its hash.
	FindByTokenHash(ctx context.Context, tokenHash string) (*RefreshToken, error)

	// FindByUserID finds all refresh tokens for a user.
	FindByUserID(ctx context.Context, userID uuid.UUID) ([]*RefreshToken, error)

	// GetActiveSessions returns all active (non-expired) sessions for a user.
	// Used for the "view active sessions" feature (requirement 1.3).
	GetActiveSessions(ctx context.Context, userID uuid.UUID, currentTokenHash string) ([]*Session, error)

	// Delete removes a refresh token.
	Delete(ctx context.Context, id uuid.UUID) error

	// DeleteByTokenHash removes a refresh token by its hash.
	DeleteByTokenHash(ctx context.Context, tokenHash string) error

	// DeleteAllByUserID removes all refresh tokens for a user.
	DeleteAllByUserID(ctx context.Context, userID uuid.UUID) error

	// DeleteExpired removes all expired refresh tokens.
	DeleteExpired(ctx context.Context) (int64, error)
}

// BackupCodeRepository defines the interface for backup code data access.
type BackupCodeRepository interface {
	// CreateBatch creates multiple backup codes for a user.
	CreateBatch(ctx context.Context, codes []*BackupCode) error

	// FindByUserID finds all backup codes for a user.
	FindByUserID(ctx context.Context, userID uuid.UUID) ([]*BackupCode, error)

	// FindUnusedByUserID finds all unused backup codes for a user.
	FindUnusedByUserID(ctx context.Context, userID uuid.UUID) ([]*BackupCode, error)

	// MarkAsUsed marks a backup code as used.
	MarkAsUsed(ctx context.Context, id uuid.UUID) error

	// DeleteAllByUserID removes all backup codes for a user.
	DeleteAllByUserID(ctx context.Context, userID uuid.UUID) error
}

// FollowRepository defines the interface for follow relationship data access.
type FollowRepository interface {
	// Create creates a new follow relationship.
	Create(ctx context.Context, follow *Follow) error

	// Delete removes a follow relationship.
	Delete(ctx context.Context, followerID, followingID uuid.UUID) error

	// Exists checks if a follow relationship exists.
	Exists(ctx context.Context, followerID, followingID uuid.UUID) (bool, error)

	// GetFollowers returns paginated followers for a user.
	GetFollowers(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*User, int, error)

	// GetFollowing returns paginated users that a user is following.
	GetFollowing(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*User, int, error)

	// CountFollowers returns the number of followers for a user.
	CountFollowers(ctx context.Context, userID uuid.UUID) (int, error)

	// CountFollowing returns the number of users a user is following.
	CountFollowing(ctx context.Context, userID uuid.UUID) (int, error)
}

// APIKeyRepository defines the interface for API key data access.
type APIKeyRepository interface {
	// Create creates a new API key.
	Create(ctx context.Context, key *APIKey) error

	// FindByID finds an API key by ID.
	FindByID(ctx context.Context, id uuid.UUID) (*APIKey, error)

	// FindByKeyHash finds an API key by its hash.
	FindByKeyHash(ctx context.Context, keyHash string) (*APIKey, error)

	// FindByUserID finds all API keys for a user.
	FindByUserID(ctx context.Context, userID uuid.UUID) ([]*APIKey, error)

	// UpdateLastUsed updates the last used timestamp for an API key.
	UpdateLastUsed(ctx context.Context, id uuid.UUID) error

	// Delete removes an API key.
	Delete(ctx context.Context, id uuid.UUID) error
}

// VerificationTokenRepository defines the interface for verification token data access.
// Implements requirement 1.2: Email Verification & Password Reset.
type VerificationTokenRepository interface {
	// Create creates a new verification token.
	Create(ctx context.Context, token *VerificationToken) error

	// FindByTokenHash finds a verification token by its hash.
	FindByTokenHash(ctx context.Context, tokenHash string) (*VerificationToken, error)

	// FindActiveByUserIDAndType finds an active (unused, non-expired) token for a user and type.
	FindActiveByUserIDAndType(ctx context.Context, userID uuid.UUID, tokenType TokenType) (*VerificationToken, error)

	// MarkAsUsed marks a verification token as used.
	MarkAsUsed(ctx context.Context, id uuid.UUID) error

	// DeleteByUserIDAndType deletes all tokens for a user and type.
	DeleteByUserIDAndType(ctx context.Context, userID uuid.UUID, tokenType TokenType) error

	// DeleteExpired removes all expired verification tokens.
	DeleteExpired(ctx context.Context) (int64, error)
}

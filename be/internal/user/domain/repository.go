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

	// SetOnboardingCompleted marks a user's onboarding as completed.
	SetOnboardingCompleted(ctx context.Context, id uuid.UUID, completed bool) error
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

// PredefinedInterestRepository defines the interface for predefined interest data access.
type PredefinedInterestRepository interface {
	// FindAll returns all active predefined interests.
	FindAll(ctx context.Context) ([]*PredefinedInterest, error)

	// FindByID finds a predefined interest by ID.
	FindByID(ctx context.Context, id uuid.UUID) (*PredefinedInterest, error)

	// FindBySlug finds a predefined interest by slug.
	FindBySlug(ctx context.Context, slug string) (*PredefinedInterest, error)

	// FindByIDs finds predefined interests by multiple IDs.
	FindByIDs(ctx context.Context, ids []uuid.UUID) ([]*PredefinedInterest, error)

	// FindByCategory returns all active predefined interests in a category.
	FindByCategory(ctx context.Context, category string) ([]*PredefinedInterest, error)

	// GetCategories returns all unique categories.
	GetCategories(ctx context.Context) ([]string, error)
}

// UserLearningInterestRepository defines the interface for user learning interest data access.
type UserLearningInterestRepository interface {
	// Create creates a new user learning interest.
	Create(ctx context.Context, interest *UserLearningInterest) error

	// CreateBatch creates multiple user learning interests.
	CreateBatch(ctx context.Context, interests []*UserLearningInterest) error

	// FindByUserID returns all learning interests for a user.
	FindByUserID(ctx context.Context, userID uuid.UUID) ([]*UserLearningInterest, error)

	// FindByUserIDWithDetails returns all learning interests for a user with predefined interest details.
	FindByUserIDWithDetails(ctx context.Context, userID uuid.UUID) ([]*UserLearningInterest, error)

	// Delete removes a user learning interest.
	Delete(ctx context.Context, id uuid.UUID) error

	// DeleteByUserID removes all learning interests for a user.
	DeleteByUserID(ctx context.Context, userID uuid.UUID) error

	// DeleteByUserIDAndPredefinedID removes a specific predefined interest for a user.
	DeleteByUserIDAndPredefinedID(ctx context.Context, userID, predefinedInterestID uuid.UUID) error

	// Exists checks if a user already has a specific predefined interest.
	ExistsByUserIDAndPredefinedID(ctx context.Context, userID, predefinedInterestID uuid.UUID) (bool, error)

	// ExistsByUserIDAndCustom checks if a user already has a specific custom interest.
	ExistsByUserIDAndCustom(ctx context.Context, userID uuid.UUID, customInterest string) (bool, error)

	// CountByUserID returns the number of interests for a user.
	CountByUserID(ctx context.Context, userID uuid.UUID) (int, error)

	// GetInterestSummaries returns simplified interest summaries for a user.
	GetInterestSummaries(ctx context.Context, userID uuid.UUID) ([]*InterestSummary, error)
}

// TeacherVerificationRepository defines the interface for teacher verification data access.
type TeacherVerificationRepository interface {
	// Create creates a new teacher verification request.
	Create(ctx context.Context, verification *TeacherVerification) error

	// FindByID finds a teacher verification by ID.
	FindByID(ctx context.Context, id uuid.UUID) (*TeacherVerification, error)

	// FindByUserID finds a teacher verification by user ID.
	// Returns the most recent verification for the user.
	FindByUserID(ctx context.Context, userID uuid.UUID) (*TeacherVerification, error)

	// FindPending returns paginated pending verification requests.
	// Used by admins to review pending verifications.
	FindPending(ctx context.Context, limit, offset int) ([]*TeacherVerification, int, error)

	// Update updates an existing teacher verification.
	Update(ctx context.Context, verification *TeacherVerification) error

	// UpdateStatus updates the status of a verification request.
	// Used when approving or rejecting a verification.
	UpdateStatus(ctx context.Context, id uuid.UUID, status VerificationStatus, reviewedBy uuid.UUID, reason *string) error

	// ExistsByUserID checks if a verification request exists for a user.
	ExistsByUserID(ctx context.Context, userID uuid.UUID) (bool, error)

	// ExistsPendingByUserID checks if a pending verification request exists for a user.
	ExistsPendingByUserID(ctx context.Context, userID uuid.UUID) (bool, error)
}

// StorageRepository defines the interface for storage calculation operations.
// Implements requirement 8.2: Define StorageRepository interface in user domain layer.
type StorageRepository interface {
	// GetUserStorageUsage returns total bytes used by a user.
	// Calculates storage by summing file_size of all non-deleted materials uploaded by the user.
	GetUserStorageUsage(ctx context.Context, userID uuid.UUID) (int64, error)
}

// AIUsageRepository defines the interface for AI usage tracking operations.
// Implements requirement 8.2: Define AIUsageRepository interface in user domain layer.
type AIUsageRepository interface {
	// GetDailyUsage returns today's AI message count for a user.
	GetDailyUsage(ctx context.Context, userID uuid.UUID) (int, error)

	// IncrementDailyUsage increments the daily AI message count for a user.
	// The count should reset at midnight UTC.
	IncrementDailyUsage(ctx context.Context, userID uuid.UUID) error
}

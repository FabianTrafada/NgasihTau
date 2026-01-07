// Package domain contains the core business entities and repository interfaces
// for the User Service. This layer is independent of external frameworks and databases.
package domain

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Role represents a user's role in the system.
type Role string

const (
	// RoleStudent is the default role for new users.
	RoleStudent Role = "student"
	// RoleVerifiedStudent is for students verified by a teacher in a pod context.
	RoleVerifiedStudent Role = "verified_student"
	// RoleTeacher is for users who can create Knowledge Pods.
	RoleTeacher Role = "teacher"
)

// Tier represents a user's subscription tier.
type Tier string

const (
	// TierFree is the default tier for new users.
	TierFree Tier = "free"
	// TierPremium is for users with premium subscription.
	TierPremium Tier = "premium"
	// TierPro is for users with pro subscription.
	TierPro Tier = "pro"
)

// Order returns the numeric order of tiers for comparison.
// Higher order means higher tier level.
func (t Tier) Order() int {
	switch t {
	case TierFree:
		return 0
	case TierPremium:
		return 1
	case TierPro:
		return 2
	default:
		return 0
	}
}

// IsAtLeast checks if the tier is at least the given tier level.
func (t Tier) IsAtLeast(other Tier) bool {
	return t.Order() >= other.Order()
}

// ValidTiers returns all valid tier values.
func ValidTiers() []Tier {
	return []Tier{TierFree, TierPremium, TierPro}
}

// IsValidTier checks if the given tier is valid.
func IsValidTier(t Tier) bool {
	for _, valid := range ValidTiers() {
		if t == valid {
			return true
		}
	}
	return false
}

// User represents a user account in the system.
type User struct {
	ID                  uuid.UUID  `json:"id"`
	Email               string     `json:"email"`
	PasswordHash        string     `json:"-"` // Never expose in JSON
	Name                string     `json:"name"`
	AvatarURL           *string    `json:"avatar_url,omitempty"`
	Bio                 *string    `json:"bio,omitempty"`
	Role                Role       `json:"role"`
	Tier                Tier       `json:"tier"`
	EmailVerified       bool       `json:"email_verified"`
	TwoFactorEnabled    bool       `json:"two_factor_enabled"`
	TwoFactorSecret     *string    `json:"-"` // Never expose in JSON
	Language            string     `json:"language"`
	OnboardingCompleted bool       `json:"onboarding_completed"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`
	DeletedAt           *time.Time `json:"-"`
}

// IsVerified returns true if the user can create Knowledge Pods.
func (u *User) IsVerified() bool {
	return u.Role == RoleTeacher || u.Role == RoleVerifiedStudent
}

// CanCreatePod returns true if the user has permission to create Knowledge Pods.
func (u *User) CanCreatePod() bool {
	return u.Role == RoleTeacher
}

// OAuthProvider represents supported OAuth providers.
type OAuthProvider string

const (
	// OAuthProviderGoogle represents Google OAuth.
	OAuthProviderGoogle OAuthProvider = "google"
)

// OAuthAccount represents a linked OAuth account.
// Implements requirement 1 for Google OAuth login.
type OAuthAccount struct {
	ID             uuid.UUID     `json:"id"`
	UserID         uuid.UUID     `json:"user_id"`
	Provider       OAuthProvider `json:"provider"`
	ProviderUserID string        `json:"provider_user_id"`
	CreatedAt      time.Time     `json:"created_at"`
}

// DeviceInfo represents device/session metadata for refresh tokens.
// Stored as JSONB in the database for tracking active sessions.
type DeviceInfo struct {
	UserAgent  string `json:"user_agent,omitempty"`
	IPAddress  string `json:"ip_address,omitempty"`
	DeviceType string `json:"device_type,omitempty"` // desktop, mobile, tablet
	Browser    string `json:"browser,omitempty"`
	OS         string `json:"os,omitempty"`
}

// Value implements driver.Valuer for database storage.
func (d *DeviceInfo) Value() (driver.Value, error) {
	if d == nil {
		return nil, nil
	}
	return json.Marshal(d)
}

// Scan implements sql.Scanner for database retrieval.
func (d *DeviceInfo) Scan(value interface{}) error {
	if value == nil {
		return nil
	}

	var data []byte
	switch v := value.(type) {
	case []byte:
		data = v
	case string:
		data = []byte(v)
	default:
		return fmt.Errorf("cannot scan type %T into DeviceInfo", value)
	}

	return json.Unmarshal(data, d)
}

// RefreshToken represents a refresh token for session management.
// Implements requirement 1.3.
type RefreshToken struct {
	ID         uuid.UUID   `json:"id"`
	UserID     uuid.UUID   `json:"user_id"`
	TokenHash  string      `json:"-"` // SHA-256 hash of the token
	DeviceInfo *DeviceInfo `json:"device_info,omitempty"`
	ExpiresAt  time.Time   `json:"expires_at"`
	CreatedAt  time.Time   `json:"created_at"`
}

// IsExpired returns true if the refresh token has expired.
func (r *RefreshToken) IsExpired() bool {
	return time.Now().After(r.ExpiresAt)
}

// Session represents an active user session for the sessions list endpoint.
// Used when users view their active sessions (requirement 1.3).
type Session struct {
	ID         uuid.UUID   `json:"id"`
	DeviceInfo *DeviceInfo `json:"device_info,omitempty"`
	CreatedAt  time.Time   `json:"created_at"`
	LastUsedAt time.Time   `json:"last_used_at"`
	IsCurrent  bool        `json:"is_current"`
}

// BackupCode represents a 2FA backup code.
// Implements requirement 1.1.
type BackupCode struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	CodeHash  string    `json:"-"` // Hashed backup code
	Used      bool      `json:"used"`
	CreatedAt time.Time `json:"created_at"`
}

// Follow represents a follow relationship between users.
// Implements requirement 2.2.
type Follow struct {
	FollowerID  uuid.UUID `json:"follower_id"`
	FollowingID uuid.UUID `json:"following_id"`
	CreatedAt   time.Time `json:"created_at"`
}

// APIKey represents an API key for programmatic access.
// Implements requirement 22.
type APIKey struct {
	ID          uuid.UUID  `json:"id"`
	UserID      uuid.UUID  `json:"user_id"`
	Name        string     `json:"name"`
	KeyHash     string     `json:"-"` // Hashed API key
	Permissions []string   `json:"permissions,omitempty"`
	LastUsedAt  *time.Time `json:"last_used_at,omitempty"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
}

// IsExpired returns true if the API key has expired.
func (a *APIKey) IsExpired() bool {
	if a.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*a.ExpiresAt)
}

// UserProfile represents the public profile information of a user.
// Used for profile views and public API responses.
type UserProfile struct {
	ID             uuid.UUID `json:"id"`
	Name           string    `json:"name"`
	AvatarURL      *string   `json:"avatar_url,omitempty"`
	Bio            *string   `json:"bio,omitempty"`
	Role           Role      `json:"role"`
	FollowerCount  int       `json:"follower_count"`
	FollowingCount int       `json:"following_count"`
	PodCount       int       `json:"pod_count"`
	MaterialCount  int       `json:"material_count"`
	CreatedAt      time.Time `json:"created_at"`
}

// UserStats represents user statistics.
type UserStats struct {
	TotalMaterials int `json:"total_materials"`
	TotalViews     int `json:"total_views"`
	TotalPods      int `json:"total_pods"`
	FollowerCount  int `json:"follower_count"`
	FollowingCount int `json:"following_count"`
}

// NewUser creates a new User with default values.
// Sets default role to student, tier to free, and language to Indonesian.
func NewUser(email, passwordHash, name string) *User {
	now := time.Now()
	return &User{
		ID:                  uuid.New(),
		Email:               email,
		PasswordHash:        passwordHash,
		Name:                name,
		Role:                RoleStudent,
		Tier:                TierFree,
		EmailVerified:       false,
		TwoFactorEnabled:    false,
		Language:            "id",
		OnboardingCompleted: false,
		CreatedAt:           now,
		UpdatedAt:           now,
	}
}

// NewOAuthAccount creates a new OAuth account link.
func NewOAuthAccount(userID uuid.UUID, provider OAuthProvider, providerUserID string) *OAuthAccount {
	return &OAuthAccount{
		ID:             uuid.New(),
		UserID:         userID,
		Provider:       provider,
		ProviderUserID: providerUserID,
		CreatedAt:      time.Now(),
	}
}

// NewRefreshToken creates a new refresh token.
// tokenHash should be the SHA-256 hash of the actual token.
func NewRefreshToken(userID uuid.UUID, tokenHash string, deviceInfo *DeviceInfo, expiresAt time.Time) *RefreshToken {
	return &RefreshToken{
		ID:         uuid.New(),
		UserID:     userID,
		TokenHash:  tokenHash,
		DeviceInfo: deviceInfo,
		ExpiresAt:  expiresAt,
		CreatedAt:  time.Now(),
	}
}

// NewBackupCode creates a new backup code.
// codeHash should be the bcrypt hash of the actual code.
func NewBackupCode(userID uuid.UUID, codeHash string) *BackupCode {
	return &BackupCode{
		ID:        uuid.New(),
		UserID:    userID,
		CodeHash:  codeHash,
		Used:      false,
		CreatedAt: time.Now(),
	}
}

// NewFollow creates a new follow relationship.
func NewFollow(followerID, followingID uuid.UUID) *Follow {
	return &Follow{
		FollowerID:  followerID,
		FollowingID: followingID,
		CreatedAt:   time.Now(),
	}
}

// NewAPIKey creates a new API key.
// keyHash should be the SHA-256 hash of the actual key.
func NewAPIKey(userID uuid.UUID, name, keyHash string, permissions []string, expiresAt *time.Time) *APIKey {
	return &APIKey{
		ID:          uuid.New(),
		UserID:      userID,
		Name:        name,
		KeyHash:     keyHash,
		Permissions: permissions,
		ExpiresAt:   expiresAt,
		CreatedAt:   time.Now(),
	}
}

// TokenType represents the type of verification token.
type TokenType string

const (
	// TokenTypeEmailVerification is for email verification tokens.
	TokenTypeEmailVerification TokenType = "email_verification"
	// TokenTypePasswordReset is for password reset tokens.
	TokenTypePasswordReset TokenType = "password_reset"
)

// VerificationToken represents a token for email verification or password reset.
// Implements requirement 1.2: Email Verification & Password Reset.
type VerificationToken struct {
	ID        uuid.UUID  `json:"id"`
	UserID    uuid.UUID  `json:"user_id"`
	TokenHash string     `json:"-"` // SHA-256 hash of the token
	TokenType TokenType  `json:"token_type"`
	ExpiresAt time.Time  `json:"expires_at"`
	UsedAt    *time.Time `json:"used_at,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
}

// IsExpired returns true if the verification token has expired.
func (v *VerificationToken) IsExpired() bool {
	return time.Now().After(v.ExpiresAt)
}

// IsUsed returns true if the verification token has been used.
func (v *VerificationToken) IsUsed() bool {
	return v.UsedAt != nil
}

// NewVerificationToken creates a new verification token.
// tokenHash should be the SHA-256 hash of the actual token.
func NewVerificationToken(userID uuid.UUID, tokenHash string, tokenType TokenType, expiresAt time.Time) *VerificationToken {
	return &VerificationToken{
		ID:        uuid.New(),
		UserID:    userID,
		TokenHash: tokenHash,
		TokenType: tokenType,
		ExpiresAt: expiresAt,
		CreatedAt: time.Now(),
	}
}

// PredefinedInterest represents a system-defined learning interest option.
// These are the default interests that users can select from during onboarding.
type PredefinedInterest struct {
	ID           uuid.UUID `json:"id"`
	Name         string    `json:"name"`
	Slug         string    `json:"slug"`
	Description  *string   `json:"description,omitempty"`
	Icon         *string   `json:"icon,omitempty"`
	Category     *string   `json:"category,omitempty"`
	DisplayOrder int       `json:"display_order"`
	IsActive     bool      `json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// UserLearningInterest represents a user's selected learning interest.
// Can be either a predefined interest or a custom one created by the user.
type UserLearningInterest struct {
	ID                   uuid.UUID  `json:"id"`
	UserID               uuid.UUID  `json:"user_id"`
	PredefinedInterestID *uuid.UUID `json:"predefined_interest_id,omitempty"`
	CustomInterest       *string    `json:"custom_interest,omitempty"`
	CreatedAt            time.Time  `json:"created_at"`

	// Joined fields (populated when fetching with predefined interest)
	PredefinedInterest *PredefinedInterest `json:"predefined_interest,omitempty"`
}

// GetInterestName returns the name of the interest (either predefined or custom).
func (u *UserLearningInterest) GetInterestName() string {
	if u.CustomInterest != nil {
		return *u.CustomInterest
	}
	if u.PredefinedInterest != nil {
		return u.PredefinedInterest.Name
	}
	return ""
}

// IsPredefined returns true if this is a predefined interest.
func (u *UserLearningInterest) IsPredefined() bool {
	return u.PredefinedInterestID != nil
}

// InterestSummary represents a simplified view of a user's interest.
// Used for API responses and recommendations.
type InterestSummary struct {
	ID       uuid.UUID `json:"id"`
	Name     string    `json:"name"`
	Slug     *string   `json:"slug,omitempty"`
	Icon     *string   `json:"icon,omitempty"`
	Category *string   `json:"category,omitempty"`
	IsCustom bool      `json:"is_custom"`
}

// NewUserLearningInterestFromPredefined creates a new user learning interest from a predefined interest.
func NewUserLearningInterestFromPredefined(userID, predefinedInterestID uuid.UUID) *UserLearningInterest {
	return &UserLearningInterest{
		ID:                   uuid.New(),
		UserID:               userID,
		PredefinedInterestID: &predefinedInterestID,
		CreatedAt:            time.Now(),
	}
}

// NewUserLearningInterestCustom creates a new custom user learning interest.
func NewUserLearningInterestCustom(userID uuid.UUID, customInterest string) *UserLearningInterest {
	return &UserLearningInterest{
		ID:             uuid.New(),
		UserID:         userID,
		CustomInterest: &customInterest,
		CreatedAt:      time.Now(),
	}
}

// VerificationStatus represents the status of teacher verification.
type VerificationStatus string

const (
	// VerificationStatusPending is the initial status when verification is submitted.
	VerificationStatusPending VerificationStatus = "pending"
	// VerificationStatusApproved is when verification is approved by admin.
	VerificationStatusApproved VerificationStatus = "approved"
	// VerificationStatusRejected is when verification is rejected by admin.
	VerificationStatusRejected VerificationStatus = "rejected"
)

// CredentialType represents the type of verification document.
type CredentialType string

const (
	// CredentialTypeGovernmentID represents KTP/NIK.
	CredentialTypeGovernmentID CredentialType = "government_id"
	// CredentialTypeEducatorCard represents Kartu Pendidik.
	CredentialTypeEducatorCard CredentialType = "educator_card"
	// CredentialTypeProfessionalCert represents BNSP Certificate.
	CredentialTypeProfessionalCert CredentialType = "professional_cert"
)

// TeacherVerification represents a teacher verification request.
type TeacherVerification struct {
	ID              uuid.UUID          `json:"id"`
	UserID          uuid.UUID          `json:"user_id"`
	FullName        string             `json:"full_name"`
	IDNumber        string             `json:"id_number"`
	CredentialType  CredentialType     `json:"credential_type"`
	DocumentRef     string             `json:"document_ref"` // Reference only, not actual file
	Status          VerificationStatus `json:"status"`
	ReviewedBy      *uuid.UUID         `json:"reviewed_by,omitempty"`
	ReviewedAt      *time.Time         `json:"reviewed_at,omitempty"`
	RejectionReason *string            `json:"rejection_reason,omitempty"`
	CreatedAt       time.Time          `json:"created_at"`
	UpdatedAt       time.Time          `json:"updated_at"`
}

// IsPending returns true if the verification is pending review.
func (tv *TeacherVerification) IsPending() bool {
	return tv.Status == VerificationStatusPending
}

// IsApproved returns true if the verification has been approved.
func (tv *TeacherVerification) IsApproved() bool {
	return tv.Status == VerificationStatusApproved
}

// IsRejected returns true if the verification has been rejected.
func (tv *TeacherVerification) IsRejected() bool {
	return tv.Status == VerificationStatusRejected
}

// NewTeacherVerification creates a new teacher verification request.
// Sets default status to pending.
func NewTeacherVerification(userID uuid.UUID, fullName, idNumber string, credentialType CredentialType, documentRef string) *TeacherVerification {
	now := time.Now()
	return &TeacherVerification{
		ID:             uuid.New(),
		UserID:         userID,
		FullName:       fullName,
		IDNumber:       idNumber,
		CredentialType: credentialType,
		DocumentRef:    documentRef,
		Status:         VerificationStatusPending,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
}

// Approve marks the verification as approved.
func (tv *TeacherVerification) Approve(reviewerID uuid.UUID) {
	now := time.Now()
	tv.Status = VerificationStatusApproved
	tv.ReviewedBy = &reviewerID
	tv.ReviewedAt = &now
	tv.UpdatedAt = now
}

// Reject marks the verification as rejected with a reason.
func (tv *TeacherVerification) Reject(reviewerID uuid.UUID, reason string) {
	now := time.Now()
	tv.Status = VerificationStatusRejected
	tv.ReviewedBy = &reviewerID
	tv.ReviewedAt = &now
	tv.RejectionReason = &reason
	tv.UpdatedAt = now
}

// ValidCredentialTypes returns all valid credential types.
func ValidCredentialTypes() []CredentialType {
	return []CredentialType{
		CredentialTypeGovernmentID,
		CredentialTypeEducatorCard,
		CredentialTypeProfessionalCert,
	}
}

// IsValidCredentialType checks if the given credential type is valid.
func IsValidCredentialType(ct CredentialType) bool {
	for _, valid := range ValidCredentialTypes() {
		if ct == valid {
			return true
		}
	}
	return false
}

// StorageInfo represents a user's storage usage information.
// Implements Requirements 2.1-2.5 for storage information API response.
type StorageInfo struct {
	UsedBytes      int64   `json:"used_bytes"`
	QuotaBytes     int64   `json:"quota_bytes"`
	RemainingBytes int64   `json:"remaining_bytes"`
	UsagePercent   float64 `json:"usage_percent"`
	Tier           Tier    `json:"tier"`
	Warning        string  `json:"warning,omitempty"`         // "warning" at 80%, "critical" at 90%
	NextTier       *Tier   `json:"next_tier,omitempty"`       // Next tier upgrade option
	NextTierQuota  *int64  `json:"next_tier_quota,omitempty"` // Quota for next tier in bytes
}

// AIUsageInfo represents a user's AI usage information.
// Implements Requirements 10.1-10.4 for AI usage information API response.
type AIUsageInfo struct {
	UsedToday   int       `json:"used_today"`
	DailyLimit  int       `json:"daily_limit"` // -1 for unlimited
	Remaining   int       `json:"remaining"`   // -1 for unlimited
	ResetAt     time.Time `json:"reset_at"`
	Tier        Tier      `json:"tier"`
	IsUnlimited bool      `json:"is_unlimited"`
}

// Package postgres provides PostgreSQL implementations of the User Service repositories.
// Implements the Repository pattern for data access abstraction (requirement 10.3).
package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"ngasihtau/internal/common/errors"
	"ngasihtau/internal/user/domain"
)

// UserRepository implements domain.UserRepository using PostgreSQL.
type UserRepository struct {
	db DBTX
}

// NewUserRepository creates a new UserRepository.
func NewUserRepository(db DBTX) *UserRepository {
	return &UserRepository{db: db}
}

// Create creates a new user in the database.
func (r *UserRepository) Create(ctx context.Context, user *domain.User) error {
	query := `
		INSERT INTO users (id, email, password_hash, name, avatar_url, bio, role, 
			email_verified, two_factor_enabled, two_factor_secret, language, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`

	now := time.Now().UTC()
	if user.ID == uuid.Nil {
		user.ID = uuid.New()
	}
	user.CreatedAt = now
	user.UpdatedAt = now

	_, err := r.db.Exec(ctx, query,
		user.ID,
		user.Email,
		user.PasswordHash,
		user.Name,
		user.AvatarURL,
		user.Bio,
		user.Role,
		user.EmailVerified,
		user.TwoFactorEnabled,
		user.TwoFactorSecret,
		user.Language,
		user.CreatedAt,
		user.UpdatedAt,
	)
	if err != nil {
		return errors.Internal("failed to create user", err)
	}

	return nil
}

// FindByID finds a user by their ID.
func (r *UserRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	query := `
		SELECT id, email, password_hash, name, avatar_url, bio, role, 
			email_verified, two_factor_enabled, two_factor_secret, language, 
			created_at, updated_at, deleted_at
		FROM users
		WHERE id = $1 AND deleted_at IS NULL
	`

	user := &domain.User{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.Name,
		&user.AvatarURL,
		&user.Bio,
		&user.Role,
		&user.EmailVerified,
		&user.TwoFactorEnabled,
		&user.TwoFactorSecret,
		&user.Language,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.DeletedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, errors.NotFound("user", id.String())
	}
	if err != nil {
		return nil, errors.Internal("failed to find user", err)
	}

	return user, nil
}

// FindByEmail finds a user by their email address.
func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	query := `
		SELECT id, email, password_hash, name, avatar_url, bio, role, 
			email_verified, two_factor_enabled, two_factor_secret, language, 
			created_at, updated_at, deleted_at
		FROM users
		WHERE email = $1 AND deleted_at IS NULL
	`

	user := &domain.User{}
	err := r.db.QueryRow(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.Name,
		&user.AvatarURL,
		&user.Bio,
		&user.Role,
		&user.EmailVerified,
		&user.TwoFactorEnabled,
		&user.TwoFactorSecret,
		&user.Language,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.DeletedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, errors.NotFound("user", email)
	}
	if err != nil {
		return nil, errors.Internal("failed to find user", err)
	}

	return user, nil
}

// Update updates an existing user.
func (r *UserRepository) Update(ctx context.Context, user *domain.User) error {
	query := `
		UPDATE users
		SET email = $2, password_hash = $3, name = $4, avatar_url = $5, bio = $6,
			role = $7, email_verified = $8, two_factor_enabled = $9, 
			two_factor_secret = $10, language = $11, updated_at = $12
		WHERE id = $1 AND deleted_at IS NULL
	`

	user.UpdatedAt = time.Now().UTC()

	tag, err := r.db.Exec(ctx, query,
		user.ID,
		user.Email,
		user.PasswordHash,
		user.Name,
		user.AvatarURL,
		user.Bio,
		user.Role,
		user.EmailVerified,
		user.TwoFactorEnabled,
		user.TwoFactorSecret,
		user.Language,
		user.UpdatedAt,
	)
	if err != nil {
		return errors.Internal("failed to update user", err)
	}

	if tag.RowsAffected() == 0 {
		return errors.NotFound("user", user.ID.String())
	}

	return nil
}

// Delete soft-deletes a user.
func (r *UserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE users
		SET deleted_at = $2, updated_at = $2
		WHERE id = $1 AND deleted_at IS NULL
	`

	now := time.Now().UTC()
	tag, err := r.db.Exec(ctx, query, id, now)
	if err != nil {
		return errors.Internal("failed to delete user", err)
	}

	if tag.RowsAffected() == 0 {
		return errors.NotFound("user", id.String())
	}

	return nil
}

// ExistsByEmail checks if a user with the given email exists.
func (r *UserRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1 AND deleted_at IS NULL)`

	var exists bool
	err := r.db.QueryRow(ctx, query, email).Scan(&exists)
	fmt.Printf("%v\n", err)
	if err != nil {
		return false, errors.Internal("failed to check email existence", err)
	}

	return exists, nil
}

// GetProfile retrieves a user's public profile with stats.
func (r *UserRepository) GetProfile(ctx context.Context, id uuid.UUID) (*domain.UserProfile, error) {
	query := `
		SELECT 
			u.id, u.name, u.avatar_url, u.bio, u.role, u.created_at,
			COALESCE(f1.follower_count, 0) as follower_count,
			COALESCE(f2.following_count, 0) as following_count
		FROM users u
		LEFT JOIN (
			SELECT following_id, COUNT(*) as follower_count
			FROM follows
			WHERE following_id = $1
			GROUP BY following_id
		) f1 ON u.id = f1.following_id
		LEFT JOIN (
			SELECT follower_id, COUNT(*) as following_count
			FROM follows
			WHERE follower_id = $1
			GROUP BY follower_id
		) f2 ON u.id = f2.follower_id
		WHERE u.id = $1 AND u.deleted_at IS NULL
	`

	profile := &domain.UserProfile{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&profile.ID,
		&profile.Name,
		&profile.AvatarURL,
		&profile.Bio,
		&profile.Role,
		&profile.CreatedAt,
		&profile.FollowerCount,
		&profile.FollowingCount,
	)
	if err == pgx.ErrNoRows {
		return nil, errors.NotFound("user", id.String())
	}
	if err != nil {
		return nil, errors.Internal("failed to get user profile", err)
	}

	// Pod and material counts are fetched from their respective services
	// In a microservices architecture, this would be done via:
	// 1. API Gateway aggregation
	// 2. Event-driven cache updates
	// 3. BFF (Backend for Frontend) pattern
	// For now, these are set to 0 and should be populated by the API layer
	profile.PodCount = 0
	profile.MaterialCount = 0

	return profile, nil
}

// UpdateProfile updates a user's profile information.
func (r *UserRepository) UpdateProfile(ctx context.Context, id uuid.UUID, name string, bio *string, avatarURL *string) error {
	query := `
		UPDATE users
		SET name = $2, bio = $3, avatar_url = $4, updated_at = $5
		WHERE id = $1 AND deleted_at IS NULL
	`

	now := time.Now().UTC()
	tag, err := r.db.Exec(ctx, query, id, name, bio, avatarURL, now)
	if err != nil {
		return errors.Internal("failed to update profile", err)
	}

	if tag.RowsAffected() == 0 {
		return errors.NotFound("user", id.String())
	}

	return nil
}

// Enable2FA enables two-factor authentication for a user.
func (r *UserRepository) Enable2FA(ctx context.Context, id uuid.UUID, secret string) error {
	query := `
		UPDATE users
		SET two_factor_enabled = true, two_factor_secret = $2, updated_at = $3
		WHERE id = $1 AND deleted_at IS NULL
	`

	now := time.Now().UTC()
	tag, err := r.db.Exec(ctx, query, id, secret, now)
	if err != nil {
		return errors.Internal("failed to enable 2FA", err)
	}

	if tag.RowsAffected() == 0 {
		return errors.NotFound("user", id.String())
	}

	return nil
}

// Disable2FA disables two-factor authentication for a user.
func (r *UserRepository) Disable2FA(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE users
		SET two_factor_enabled = false, two_factor_secret = NULL, updated_at = $2
		WHERE id = $1 AND deleted_at IS NULL
	`

	now := time.Now().UTC()
	tag, err := r.db.Exec(ctx, query, id, now)
	if err != nil {
		return errors.Internal("failed to disable 2FA", err)
	}

	if tag.RowsAffected() == 0 {
		return errors.NotFound("user", id.String())
	}

	return nil
}

// UpdatePassword updates a user's password hash.
func (r *UserRepository) UpdatePassword(ctx context.Context, id uuid.UUID, passwordHash string) error {
	query := `
		UPDATE users
		SET password_hash = $2, updated_at = $3
		WHERE id = $1 AND deleted_at IS NULL
	`

	now := time.Now().UTC()
	tag, err := r.db.Exec(ctx, query, id, passwordHash, now)
	if err != nil {
		return errors.Internal("failed to update password", err)
	}

	if tag.RowsAffected() == 0 {
		return errors.NotFound("user", id.String())
	}

	return nil
}

// VerifyEmail marks a user's email as verified.
func (r *UserRepository) VerifyEmail(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE users
		SET email_verified = true, updated_at = $2
		WHERE id = $1 AND deleted_at IS NULL
	`

	now := time.Now().UTC()
	tag, err := r.db.Exec(ctx, query, id, now)
	if err != nil {
		return errors.Internal("failed to verify email", err)
	}

	if tag.RowsAffected() == 0 {
		return errors.NotFound("user", id.String())
	}

	return nil
}

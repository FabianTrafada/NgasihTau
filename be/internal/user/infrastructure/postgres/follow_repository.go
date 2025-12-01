package postgres

import (
	"context"
	"time"

	"github.com/google/uuid"

	"ngasihtau/internal/common/errors"
	"ngasihtau/internal/user/domain"
)

// FollowRepository implements domain.FollowRepository using PostgreSQL.
type FollowRepository struct {
	db DBTX
}

// NewFollowRepository creates a new FollowRepository.
func NewFollowRepository(db DBTX) *FollowRepository {
	return &FollowRepository{db: db}
}

// Create creates a new follow relationship.
func (r *FollowRepository) Create(ctx context.Context, follow *domain.Follow) error {
	query := `
		INSERT INTO follows (follower_id, following_id, created_at)
		VALUES ($1, $2, $3)
		ON CONFLICT (follower_id, following_id) DO NOTHING
	`

	follow.CreatedAt = time.Now().UTC()

	_, err := r.db.Exec(ctx, query,
		follow.FollowerID,
		follow.FollowingID,
		follow.CreatedAt,
	)
	if err != nil {
		return errors.Internal("failed to create follow", err)
	}

	return nil
}

// Delete removes a follow relationship.
func (r *FollowRepository) Delete(ctx context.Context, followerID, followingID uuid.UUID) error {
	query := `DELETE FROM follows WHERE follower_id = $1 AND following_id = $2`

	tag, err := r.db.Exec(ctx, query, followerID, followingID)
	if err != nil {
		return errors.Internal("failed to delete follow", err)
	}

	if tag.RowsAffected() == 0 {
		return errors.NotFoundMsg("follow relationship not found")
	}

	return nil
}

// Exists checks if a follow relationship exists.
func (r *FollowRepository) Exists(ctx context.Context, followerID, followingID uuid.UUID) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM follows WHERE follower_id = $1 AND following_id = $2)`

	var exists bool
	err := r.db.QueryRow(ctx, query, followerID, followingID).Scan(&exists)
	if err != nil {
		return false, errors.Internal("failed to check follow existence", err)
	}

	return exists, nil
}


// GetFollowers returns paginated followers for a user.
func (r *FollowRepository) GetFollowers(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*domain.User, int, error) {
	// Get total count
	countQuery := `SELECT COUNT(*) FROM follows WHERE following_id = $1`
	var total int
	err := r.db.QueryRow(ctx, countQuery, userID).Scan(&total)
	if err != nil {
		return nil, 0, errors.Internal("failed to count followers", err)
	}

	// Get followers with user details
	query := `
		SELECT u.id, u.email, u.name, u.avatar_url, u.bio, u.role, 
			u.email_verified, u.two_factor_enabled, u.language, 
			u.created_at, u.updated_at
		FROM follows f
		JOIN users u ON f.follower_id = u.id
		WHERE f.following_id = $1 AND u.deleted_at IS NULL
		ORDER BY f.created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, 0, errors.Internal("failed to get followers", err)
	}
	defer rows.Close()

	var users []*domain.User
	for rows.Next() {
		user := &domain.User{}
		err := rows.Scan(
			&user.ID,
			&user.Email,
			&user.Name,
			&user.AvatarURL,
			&user.Bio,
			&user.Role,
			&user.EmailVerified,
			&user.TwoFactorEnabled,
			&user.Language,
			&user.CreatedAt,
			&user.UpdatedAt,
		)
		if err != nil {
			return nil, 0, errors.Internal("failed to scan user", err)
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, errors.Internal("error iterating followers", err)
	}

	return users, total, nil
}

// GetFollowing returns paginated users that a user is following.
func (r *FollowRepository) GetFollowing(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*domain.User, int, error) {
	// Get total count
	countQuery := `SELECT COUNT(*) FROM follows WHERE follower_id = $1`
	var total int
	err := r.db.QueryRow(ctx, countQuery, userID).Scan(&total)
	if err != nil {
		return nil, 0, errors.Internal("failed to count following", err)
	}

	// Get following with user details
	query := `
		SELECT u.id, u.email, u.name, u.avatar_url, u.bio, u.role, 
			u.email_verified, u.two_factor_enabled, u.language, 
			u.created_at, u.updated_at
		FROM follows f
		JOIN users u ON f.following_id = u.id
		WHERE f.follower_id = $1 AND u.deleted_at IS NULL
		ORDER BY f.created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, 0, errors.Internal("failed to get following", err)
	}
	defer rows.Close()

	var users []*domain.User
	for rows.Next() {
		user := &domain.User{}
		err := rows.Scan(
			&user.ID,
			&user.Email,
			&user.Name,
			&user.AvatarURL,
			&user.Bio,
			&user.Role,
			&user.EmailVerified,
			&user.TwoFactorEnabled,
			&user.Language,
			&user.CreatedAt,
			&user.UpdatedAt,
		)
		if err != nil {
			return nil, 0, errors.Internal("failed to scan user", err)
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, errors.Internal("error iterating following", err)
	}

	return users, total, nil
}

// CountFollowers returns the number of followers for a user.
func (r *FollowRepository) CountFollowers(ctx context.Context, userID uuid.UUID) (int, error) {
	query := `SELECT COUNT(*) FROM follows WHERE following_id = $1`

	var count int
	err := r.db.QueryRow(ctx, query, userID).Scan(&count)
	if err != nil {
		return 0, errors.Internal("failed to count followers", err)
	}

	return count, nil
}

// CountFollowing returns the number of users a user is following.
func (r *FollowRepository) CountFollowing(ctx context.Context, userID uuid.UUID) (int, error) {
	query := `SELECT COUNT(*) FROM follows WHERE follower_id = $1`

	var count int
	err := r.db.QueryRow(ctx, query, userID).Scan(&count)
	if err != nil {
		return 0, errors.Internal("failed to count following", err)
	}

	return count, nil
}

// Package adapter provides external service adapters for the pod service.
package adapter

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"ngasihtau/internal/common/errors"
)

// UserFinderAdapter implements application.UserFinder interface
// by querying the user database directly.
type UserFinderAdapter struct {
	db *pgxpool.Pool
}

// NewUserFinderAdapter creates a new UserFinderAdapter.
func NewUserFinderAdapter(db *pgxpool.Pool) *UserFinderAdapter {
	return &UserFinderAdapter{db: db}
}

// FindUserByEmail finds a user by their email address.
// Returns the user ID if found, or an error if not found.
// This does not expose user enumeration - only used internally for collaboration invites.
func (a *UserFinderAdapter) FindUserByEmail(ctx context.Context, email string) (uuid.UUID, error) {
	var userID uuid.UUID
	query := `SELECT id FROM users WHERE email = $1 AND deleted_at IS NULL`
	err := a.db.QueryRow(ctx, query, email).Scan(&userID)
	if err != nil {
		return uuid.Nil, errors.NotFound("user", "user with this email not found")
	}
	return userID, nil
}

// GetUserDetails gets user details by user ID.
// Returns name, email, and avatar_url for a user.
func (a *UserFinderAdapter) GetUserDetails(ctx context.Context, userID uuid.UUID) (string, string, *string, error) {
	var name, email string
	var avatarURL *string
	query := `SELECT name, email, avatar_url FROM users WHERE id = $1 AND deleted_at IS NULL`
	err := a.db.QueryRow(ctx, query, userID).Scan(&name, &email, &avatarURL)
	if err != nil {
		return "", "", nil, errors.NotFound("user", userID.String())
	}
	return name, email, avatarURL, nil
}

// Package application contains the business logic for the pod service.
package application

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"ngasihtau/internal/common/errors"
)

// PostgresUserRoleChecker implements UserRoleChecker using direct database access.
// This is used to check user roles from the user database.
// Implements requirement 4.1: Validate requester is teacher, target pod owner is teacher.
type PostgresUserRoleChecker struct {
	db *pgxpool.Pool
}

// NewPostgresUserRoleChecker creates a new PostgresUserRoleChecker.
func NewPostgresUserRoleChecker(db *pgxpool.Pool) *PostgresUserRoleChecker {
	return &PostgresUserRoleChecker{db: db}
}

// IsTeacher checks if the user with the given ID has the teacher role.
// Returns true if the user has the "teacher" role, false otherwise.
func (c *PostgresUserRoleChecker) IsTeacher(ctx context.Context, userID uuid.UUID) (bool, error) {
	query := `SELECT role FROM users WHERE id = $1 AND deleted_at IS NULL`

	var role string
	err := c.db.QueryRow(ctx, query, userID).Scan(&role)
	if err != nil {
		return false, errors.NotFound("user", "user not found")
	}

	return role == "teacher", nil
}

// NoOpUserRoleChecker is a no-op implementation that always returns true.
// Used when user role checking is not available or not required.
type NoOpUserRoleChecker struct{}

// NewNoOpUserRoleChecker creates a new NoOpUserRoleChecker.
func NewNoOpUserRoleChecker() *NoOpUserRoleChecker {
	return &NoOpUserRoleChecker{}
}

// IsTeacher always returns true for the no-op implementation.
// This allows the system to work without user role validation when needed.
func (c *NoOpUserRoleChecker) IsTeacher(ctx context.Context, userID uuid.UUID) (bool, error) {
	return true, nil
}

package postgres

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"ngasihtau/internal/common/errors"
	"ngasihtau/internal/pod/domain"
)

// CollaboratorRepository implements domain.CollaboratorRepository using PostgreSQL.
type CollaboratorRepository struct {
	db *pgxpool.Pool
}

// NewCollaboratorRepository creates a new CollaboratorRepository.
func NewCollaboratorRepository(db *pgxpool.Pool) *CollaboratorRepository {
	return &CollaboratorRepository{db: db}
}

// Create creates a new collaborator.
func (r *CollaboratorRepository) Create(ctx context.Context, collaborator *domain.Collaborator) error {
	query := `
		INSERT INTO collaborators (id, pod_id, user_id, role, status, invited_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err := r.db.Exec(ctx, query,
		collaborator.ID, collaborator.PodID, collaborator.UserID, collaborator.Role,
		collaborator.Status, collaborator.InvitedBy, collaborator.CreatedAt, collaborator.UpdatedAt,
	)
	if err != nil {
		return errors.Internal("failed to create collaborator", err)
	}
	return nil
}

// FindByID finds a collaborator by ID.
func (r *CollaboratorRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Collaborator, error) {
	query := `
		SELECT id, pod_id, user_id, role, status, invited_by, created_at, updated_at
		FROM collaborators WHERE id = $1
	`
	return r.scanCollaborator(r.db.QueryRow(ctx, query, id))
}

// FindByPodAndUser finds a collaborator by pod ID and user ID.
func (r *CollaboratorRepository) FindByPodAndUser(ctx context.Context, podID, userID uuid.UUID) (*domain.Collaborator, error) {
	query := `
		SELECT id, pod_id, user_id, role, status, invited_by, created_at, updated_at
		FROM collaborators WHERE pod_id = $1 AND user_id = $2
	`
	return r.scanCollaborator(r.db.QueryRow(ctx, query, podID, userID))
}

// FindByPodID finds all collaborators for a pod.
func (r *CollaboratorRepository) FindByPodID(ctx context.Context, podID uuid.UUID) ([]*domain.Collaborator, error) {
	query := `
		SELECT id, pod_id, user_id, role, status, invited_by, created_at, updated_at
		FROM collaborators WHERE pod_id = $1 ORDER BY created_at ASC
	`
	rows, err := r.db.Query(ctx, query, podID)
	if err != nil {
		return nil, errors.Internal("failed to query collaborators", err)
	}
	defer rows.Close()
	return r.scanCollaborators(rows)
}

// FindByPodIDWithUsers finds all collaborators for a pod with user details.
func (r *CollaboratorRepository) FindByPodIDWithUsers(ctx context.Context, podID uuid.UUID) ([]*domain.CollaboratorWithUser, error) {
	query := `
		SELECT c.id, c.pod_id, c.user_id, c.role, c.status, c.invited_by, c.created_at, c.updated_at,
			u.name, u.email, u.avatar_url
		FROM collaborators c
		JOIN users u ON c.user_id = u.id
		WHERE c.pod_id = $1 ORDER BY c.created_at ASC
	`
	rows, err := r.db.Query(ctx, query, podID)
	if err != nil {
		return nil, errors.Internal("failed to query collaborators", err)
	}
	defer rows.Close()

	var collaborators []*domain.CollaboratorWithUser
	for rows.Next() {
		var c domain.CollaboratorWithUser
		err := rows.Scan(
			&c.ID, &c.PodID, &c.UserID, &c.Role, &c.Status, &c.InvitedBy, &c.CreatedAt, &c.UpdatedAt,
			&c.UserName, &c.UserEmail, &c.UserAvatarURL,
		)
		if err != nil {
			return nil, errors.Internal("failed to scan collaborator", err)
		}
		collaborators = append(collaborators, &c)
	}
	return collaborators, nil
}

// FindByUserID finds all collaborations for a user.
func (r *CollaboratorRepository) FindByUserID(ctx context.Context, userID uuid.UUID) ([]*domain.Collaborator, error) {
	query := `
		SELECT id, pod_id, user_id, role, status, invited_by, created_at, updated_at
		FROM collaborators WHERE user_id = $1 ORDER BY created_at DESC
	`
	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, errors.Internal("failed to query collaborators", err)
	}
	defer rows.Close()
	return r.scanCollaborators(rows)
}

// Update updates a collaborator.
func (r *CollaboratorRepository) Update(ctx context.Context, collaborator *domain.Collaborator) error {
	query := `
		UPDATE collaborators SET role = $2, status = $3, updated_at = $4
		WHERE id = $1
	`
	result, err := r.db.Exec(ctx, query, collaborator.ID, collaborator.Role, collaborator.Status, time.Now())
	if err != nil {
		return errors.Internal("failed to update collaborator", err)
	}
	if result.RowsAffected() == 0 {
		return errors.NotFound("collaborator", collaborator.ID.String())
	}
	return nil
}

// Delete removes a collaborator.
func (r *CollaboratorRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM collaborators WHERE id = $1`
	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return errors.Internal("failed to delete collaborator", err)
	}
	if result.RowsAffected() == 0 {
		return errors.NotFound("collaborator", id.String())
	}
	return nil
}

// DeleteByPodAndUser removes a collaborator by pod ID and user ID.
func (r *CollaboratorRepository) DeleteByPodAndUser(ctx context.Context, podID, userID uuid.UUID) error {
	query := `DELETE FROM collaborators WHERE pod_id = $1 AND user_id = $2`
	_, err := r.db.Exec(ctx, query, podID, userID)
	if err != nil {
		return errors.Internal("failed to delete collaborator", err)
	}
	return nil
}

// Exists checks if a collaborator relationship exists.
func (r *CollaboratorRepository) Exists(ctx context.Context, podID, userID uuid.UUID) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM collaborators WHERE pod_id = $1 AND user_id = $2)`
	var exists bool
	if err := r.db.QueryRow(ctx, query, podID, userID).Scan(&exists); err != nil {
		return false, errors.Internal("failed to check collaborator existence", err)
	}
	return exists, nil
}

// UpdateStatus updates a collaborator's status.
func (r *CollaboratorRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.CollaboratorStatus) error {
	query := `UPDATE collaborators SET status = $2, updated_at = $3 WHERE id = $1`
	result, err := r.db.Exec(ctx, query, id, status, time.Now())
	if err != nil {
		return errors.Internal("failed to update collaborator status", err)
	}
	if result.RowsAffected() == 0 {
		return errors.NotFound("collaborator", id.String())
	}
	return nil
}

// UpdateRole updates a collaborator's role.
func (r *CollaboratorRepository) UpdateRole(ctx context.Context, id uuid.UUID, role domain.CollaboratorRole) error {
	query := `UPDATE collaborators SET role = $2, updated_at = $3 WHERE id = $1`
	result, err := r.db.Exec(ctx, query, id, role, time.Now())
	if err != nil {
		return errors.Internal("failed to update collaborator role", err)
	}
	if result.RowsAffected() == 0 {
		return errors.NotFound("collaborator", id.String())
	}
	return nil
}

// scanCollaborator scans a single collaborator from a row.
func (r *CollaboratorRepository) scanCollaborator(row pgx.Row) (*domain.Collaborator, error) {
	var c domain.Collaborator
	err := row.Scan(&c.ID, &c.PodID, &c.UserID, &c.Role, &c.Status, &c.InvitedBy, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, errors.NotFound("collaborator", "")
		}
		return nil, errors.Internal("failed to scan collaborator", err)
	}
	return &c, nil
}

// scanCollaborators scans multiple collaborators from rows.
func (r *CollaboratorRepository) scanCollaborators(rows pgx.Rows) ([]*domain.Collaborator, error) {
	var collaborators []*domain.Collaborator
	for rows.Next() {
		var c domain.Collaborator
		err := rows.Scan(&c.ID, &c.PodID, &c.UserID, &c.Role, &c.Status, &c.InvitedBy, &c.CreatedAt, &c.UpdatedAt)
		if err != nil {
			return nil, errors.Internal("failed to scan collaborator", err)
		}
		collaborators = append(collaborators, &c)
	}
	return collaborators, nil
}

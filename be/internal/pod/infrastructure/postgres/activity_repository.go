package postgres

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"ngasihtau/internal/common/errors"
	"ngasihtau/internal/pod/domain"
)

// ActivityRepository implements domain.ActivityRepository using PostgreSQL.
type ActivityRepository struct {
	db *pgxpool.Pool
}

// NewActivityRepository creates a new ActivityRepository.
func NewActivityRepository(db *pgxpool.Pool) *ActivityRepository {
	return &ActivityRepository{db: db}
}

// Create creates a new activity.
func (r *ActivityRepository) Create(ctx context.Context, activity *domain.Activity) error {
	query := `
		INSERT INTO activities (id, pod_id, user_id, action, metadata, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := r.db.Exec(ctx, query,
		activity.ID, activity.PodID, activity.UserID, activity.Action,
		activity.Metadata, activity.CreatedAt,
	)
	if err != nil {
		return errors.Internal("failed to create activity", err)
	}
	return nil
}

// FindByPodID finds activities for a pod.
func (r *ActivityRepository) FindByPodID(ctx context.Context, podID uuid.UUID, limit, offset int) ([]*domain.Activity, int, error) {
	countQuery := `SELECT COUNT(*) FROM activities WHERE pod_id = $1`
	var total int
	if err := r.db.QueryRow(ctx, countQuery, podID).Scan(&total); err != nil {
		return nil, 0, errors.Internal("failed to count activities", err)
	}

	query := `
		SELECT id, pod_id, user_id, action, metadata, created_at
		FROM activities WHERE pod_id = $1
		ORDER BY created_at DESC LIMIT $2 OFFSET $3
	`
	rows, err := r.db.Query(ctx, query, podID, limit, offset)
	if err != nil {
		return nil, 0, errors.Internal("failed to query activities", err)
	}
	defer rows.Close()

	var activities []*domain.Activity
	for rows.Next() {
		var a domain.Activity
		err := rows.Scan(&a.ID, &a.PodID, &a.UserID, &a.Action, &a.Metadata, &a.CreatedAt)
		if err != nil {
			return nil, 0, errors.Internal("failed to scan activity", err)
		}
		activities = append(activities, &a)
	}
	return activities, total, nil
}

// FindByPodIDWithDetails finds activities for a pod with user details.
func (r *ActivityRepository) FindByPodIDWithDetails(ctx context.Context, podID uuid.UUID, limit, offset int) ([]*domain.ActivityWithDetails, int, error) {
	countQuery := `SELECT COUNT(*) FROM activities WHERE pod_id = $1`
	var total int
	if err := r.db.QueryRow(ctx, countQuery, podID).Scan(&total); err != nil {
		return nil, 0, errors.Internal("failed to count activities", err)
	}

	query := `
		SELECT a.id, a.pod_id, a.user_id, a.action, a.metadata, a.created_at,
			u.name, u.avatar_url, p.name, p.slug
		FROM activities a
		JOIN users u ON a.user_id = u.id
		JOIN pods p ON a.pod_id = p.id
		WHERE a.pod_id = $1
		ORDER BY a.created_at DESC LIMIT $2 OFFSET $3
	`
	rows, err := r.db.Query(ctx, query, podID, limit, offset)
	if err != nil {
		return nil, 0, errors.Internal("failed to query activities", err)
	}
	defer rows.Close()

	var activities []*domain.ActivityWithDetails
	for rows.Next() {
		var a domain.ActivityWithDetails
		err := rows.Scan(
			&a.ID, &a.PodID, &a.UserID, &a.Action, &a.Metadata, &a.CreatedAt,
			&a.UserName, &a.UserAvatarURL, &a.PodName, &a.PodSlug,
		)
		if err != nil {
			return nil, 0, errors.Internal("failed to scan activity", err)
		}
		activities = append(activities, &a)
	}
	return activities, total, nil
}

// GetUserFeed returns activity feed for a user based on followed pods.
func (r *ActivityRepository) GetUserFeed(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*domain.ActivityWithDetails, int, error) {
	countQuery := `
		SELECT COUNT(*) FROM activities a
		JOIN pod_follows f ON a.pod_id = f.pod_id
		WHERE f.user_id = $1
	`
	var total int
	if err := r.db.QueryRow(ctx, countQuery, userID).Scan(&total); err != nil {
		return nil, 0, errors.Internal("failed to count feed activities", err)
	}

	query := `
		SELECT a.id, a.pod_id, a.user_id, a.action, a.metadata, a.created_at,
			u.name, u.avatar_url, p.name, p.slug
		FROM activities a
		JOIN pod_follows f ON a.pod_id = f.pod_id
		JOIN users u ON a.user_id = u.id
		JOIN pods p ON a.pod_id = p.id
		WHERE f.user_id = $1 AND p.deleted_at IS NULL
		ORDER BY a.created_at DESC LIMIT $2 OFFSET $3
	`
	rows, err := r.db.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, 0, errors.Internal("failed to query feed activities", err)
	}
	defer rows.Close()

	var activities []*domain.ActivityWithDetails
	for rows.Next() {
		var a domain.ActivityWithDetails
		err := rows.Scan(
			&a.ID, &a.PodID, &a.UserID, &a.Action, &a.Metadata, &a.CreatedAt,
			&a.UserName, &a.UserAvatarURL, &a.PodName, &a.PodSlug,
		)
		if err != nil {
			return nil, 0, errors.Internal("failed to scan activity", err)
		}
		activities = append(activities, &a)
	}
	return activities, total, nil
}

// DeleteByPodID deletes all activities for a pod.
func (r *ActivityRepository) DeleteByPodID(ctx context.Context, podID uuid.UUID) error {
	query := `DELETE FROM activities WHERE pod_id = $1`
	_, err := r.db.Exec(ctx, query, podID)
	if err != nil {
		return errors.Internal("failed to delete activities", err)
	}
	return nil
}

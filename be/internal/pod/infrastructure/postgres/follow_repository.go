package postgres

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"ngasihtau/internal/common/errors"
	"ngasihtau/internal/pod/domain"
)

// PodFollowRepository implements domain.PodFollowRepository using PostgreSQL.
type PodFollowRepository struct {
	db *pgxpool.Pool
}

// NewPodFollowRepository creates a new PodFollowRepository.
func NewPodFollowRepository(db *pgxpool.Pool) *PodFollowRepository {
	return &PodFollowRepository{db: db}
}

// Create creates a new follow.
func (r *PodFollowRepository) Create(ctx context.Context, follow *domain.PodFollow) error {
	query := `INSERT INTO pod_follows (user_id, pod_id, created_at) VALUES ($1, $2, $3)`
	_, err := r.db.Exec(ctx, query, follow.UserID, follow.PodID, follow.CreatedAt)
	if err != nil {
		return errors.Internal("failed to create follow", err)
	}
	return nil
}

// Delete removes a follow.
func (r *PodFollowRepository) Delete(ctx context.Context, userID, podID uuid.UUID) error {
	query := `DELETE FROM pod_follows WHERE user_id = $1 AND pod_id = $2`
	result, err := r.db.Exec(ctx, query, userID, podID)
	if err != nil {
		return errors.Internal("failed to delete follow", err)
	}
	if result.RowsAffected() == 0 {
		return errors.NotFound("follow", "")
	}
	return nil
}

// Exists checks if a follow exists.
func (r *PodFollowRepository) Exists(ctx context.Context, userID, podID uuid.UUID) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM pod_follows WHERE user_id = $1 AND pod_id = $2)`
	var exists bool
	if err := r.db.QueryRow(ctx, query, userID, podID).Scan(&exists); err != nil {
		return false, errors.Internal("failed to check follow existence", err)
	}
	return exists, nil
}

// GetFollowedPods returns paginated followed pods for a user.
func (r *PodFollowRepository) GetFollowedPods(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*domain.Pod, int, error) {
	countQuery := `SELECT COUNT(*) FROM pod_follows WHERE user_id = $1`
	var total int
	if err := r.db.QueryRow(ctx, countQuery, userID).Scan(&total); err != nil {
		return nil, 0, errors.Internal("failed to count followed pods", err)
	}

	query := `
		SELECT p.id, p.owner_id, p.name, p.slug, p.description, p.visibility, p.categories, p.tags,
			p.star_count, p.fork_count, p.view_count, p.forked_from_id, p.created_at, p.updated_at, p.deleted_at
		FROM pods p
		JOIN pod_follows f ON p.id = f.pod_id
		WHERE f.user_id = $1 AND p.deleted_at IS NULL
		ORDER BY f.created_at DESC LIMIT $2 OFFSET $3
	`
	rows, err := r.db.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, 0, errors.Internal("failed to query followed pods", err)
	}
	defer rows.Close()

	var pods []*domain.Pod
	for rows.Next() {
		var pod domain.Pod
		var categories, tags []string
		err := rows.Scan(
			&pod.ID, &pod.OwnerID, &pod.Name, &pod.Slug, &pod.Description, &pod.Visibility,
			&categories, &tags,
			&pod.StarCount, &pod.ForkCount, &pod.ViewCount, &pod.ForkedFromID,
			&pod.CreatedAt, &pod.UpdatedAt, &pod.DeletedAt,
		)
		if err != nil {
			return nil, 0, errors.Internal("failed to scan pod", err)
		}
		pod.Categories = categories
		pod.Tags = tags
		pods = append(pods, &pod)
	}
	return pods, total, nil
}

// GetFollowerIDs returns all user IDs following a pod.
func (r *PodFollowRepository) GetFollowerIDs(ctx context.Context, podID uuid.UUID) ([]uuid.UUID, error) {
	query := `SELECT user_id FROM pod_follows WHERE pod_id = $1`
	rows, err := r.db.Query(ctx, query, podID)
	if err != nil {
		return nil, errors.Internal("failed to query followers", err)
	}
	defer rows.Close()

	var userIDs []uuid.UUID
	for rows.Next() {
		var userID uuid.UUID
		if err := rows.Scan(&userID); err != nil {
			return nil, errors.Internal("failed to scan user ID", err)
		}
		userIDs = append(userIDs, userID)
	}
	return userIDs, nil
}

// CountByPodID returns the follower count for a pod.
func (r *PodFollowRepository) CountByPodID(ctx context.Context, podID uuid.UUID) (int, error) {
	query := `SELECT COUNT(*) FROM pod_follows WHERE pod_id = $1`
	var count int
	if err := r.db.QueryRow(ctx, query, podID).Scan(&count); err != nil {
		return 0, errors.Internal("failed to count followers", err)
	}
	return count, nil
}

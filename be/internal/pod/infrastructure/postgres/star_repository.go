package postgres

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"ngasihtau/internal/common/errors"
	"ngasihtau/internal/pod/domain"
)

// PodStarRepository implements domain.PodStarRepository using PostgreSQL.
type PodStarRepository struct {
	db *pgxpool.Pool
}

// NewPodStarRepository creates a new PodStarRepository.
func NewPodStarRepository(db *pgxpool.Pool) *PodStarRepository {
	return &PodStarRepository{db: db}
}

// Create creates a new star.
func (r *PodStarRepository) Create(ctx context.Context, star *domain.PodStar) error {
	query := `INSERT INTO pod_stars (user_id, pod_id, created_at) VALUES ($1, $2, $3)`
	_, err := r.db.Exec(ctx, query, star.UserID, star.PodID, star.CreatedAt)
	if err != nil {
		return errors.Internal("failed to create star", err)
	}
	return nil
}

// Delete removes a star.
func (r *PodStarRepository) Delete(ctx context.Context, userID, podID uuid.UUID) error {
	query := `DELETE FROM pod_stars WHERE user_id = $1 AND pod_id = $2`
	result, err := r.db.Exec(ctx, query, userID, podID)
	if err != nil {
		return errors.Internal("failed to delete star", err)
	}
	if result.RowsAffected() == 0 {
		return errors.NotFound("star", "")
	}
	return nil
}

// Exists checks if a star exists.
func (r *PodStarRepository) Exists(ctx context.Context, userID, podID uuid.UUID) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM pod_stars WHERE user_id = $1 AND pod_id = $2)`
	var exists bool
	if err := r.db.QueryRow(ctx, query, userID, podID).Scan(&exists); err != nil {
		return false, errors.Internal("failed to check star existence", err)
	}
	return exists, nil
}

// GetStarredPods returns paginated starred pods for a user.
func (r *PodStarRepository) GetStarredPods(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*domain.Pod, int, error) {
	countQuery := `SELECT COUNT(*) FROM pod_stars WHERE user_id = $1`
	var total int
	if err := r.db.QueryRow(ctx, countQuery, userID).Scan(&total); err != nil {
		return nil, 0, errors.Internal("failed to count starred pods", err)
	}

	query := `
		SELECT p.id, p.owner_id, p.name, p.slug, p.description, p.visibility, p.categories, p.tags,
			p.star_count, p.fork_count, p.view_count, p.is_verified, p.upvote_count, p.forked_from_id, p.created_at, p.updated_at, p.deleted_at
		FROM pods p
		JOIN pod_stars s ON p.id = s.pod_id
		WHERE s.user_id = $1 AND p.deleted_at IS NULL
		ORDER BY s.created_at DESC LIMIT $2 OFFSET $3
	`
	rows, err := r.db.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, 0, errors.Internal("failed to query starred pods", err)
	}
	defer rows.Close()

	var pods []*domain.Pod
	for rows.Next() {
		var pod domain.Pod
		var categories, tags []string
		err := rows.Scan(
			&pod.ID, &pod.OwnerID, &pod.Name, &pod.Slug, &pod.Description, &pod.Visibility,
			&categories, &tags,
			&pod.StarCount, &pod.ForkCount, &pod.ViewCount, &pod.IsVerified, &pod.UpvoteCount, &pod.ForkedFromID,
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

// CountByPodID returns the star count for a pod.
func (r *PodStarRepository) CountByPodID(ctx context.Context, podID uuid.UUID) (int, error) {
	query := `SELECT COUNT(*) FROM pod_stars WHERE pod_id = $1`
	var count int
	if err := r.db.QueryRow(ctx, query, podID).Scan(&count); err != nil {
		return 0, errors.Internal("failed to count stars", err)
	}
	return count, nil
}

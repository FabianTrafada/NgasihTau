package postgres

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"ngasihtau/internal/common/errors"
	"ngasihtau/internal/pod/domain"
)

// PodDownvoteRepository implements domain.PodDownvoteRepository using PostgreSQL.
// Follows the existing PodUpvoteRepository pattern for consistency.
type PodDownvoteRepository struct {
	db *pgxpool.Pool
}

// NewPodDownvoteRepository creates a new PodDownvoteRepository.
func NewPodDownvoteRepository(db *pgxpool.Pool) *PodDownvoteRepository {
	return &PodDownvoteRepository{db: db}
}

// Create creates a new downvote.
func (r *PodDownvoteRepository) Create(ctx context.Context, downvote *domain.PodDownvote) error {
	query := `INSERT INTO pod_downvotes (user_id, pod_id, created_at) VALUES ($1, $2, $3)`
	_, err := r.db.Exec(ctx, query, downvote.UserID, downvote.PodID, downvote.CreatedAt)
	if err != nil {
		return errors.Internal("failed to create downvote", err)
	}
	return nil
}

// Delete removes a downvote.
func (r *PodDownvoteRepository) Delete(ctx context.Context, userID, podID uuid.UUID) error {
	query := `DELETE FROM pod_downvotes WHERE user_id = $1 AND pod_id = $2`
	result, err := r.db.Exec(ctx, query, userID, podID)
	if err != nil {
		return errors.Internal("failed to delete downvote", err)
	}
	if result.RowsAffected() == 0 {
		return errors.NotFound("downvote", "")
	}
	return nil
}

// Exists checks if a downvote exists.
func (r *PodDownvoteRepository) Exists(ctx context.Context, userID, podID uuid.UUID) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM pod_downvotes WHERE user_id = $1 AND pod_id = $2)`
	var exists bool
	if err := r.db.QueryRow(ctx, query, userID, podID).Scan(&exists); err != nil {
		return false, errors.Internal("failed to check downvote existence", err)
	}
	return exists, nil
}

// CountByPodID returns the downvote count for a pod.
func (r *PodDownvoteRepository) CountByPodID(ctx context.Context, podID uuid.UUID) (int, error) {
	query := `SELECT COUNT(*) FROM pod_downvotes WHERE pod_id = $1`
	var count int
	if err := r.db.QueryRow(ctx, query, podID).Scan(&count); err != nil {
		return 0, errors.Internal("failed to count downvotes", err)
	}
	return count, nil
}

// GetDownvotedPods returns paginated downvoted pods for a user.
func (r *PodDownvoteRepository) GetDownvotedPods(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*domain.Pod, int, error) {
	countQuery := `SELECT COUNT(*) FROM pod_downvotes WHERE user_id = $1`
	var total int
	if err := r.db.QueryRow(ctx, countQuery, userID).Scan(&total); err != nil {
		return nil, 0, errors.Internal("failed to count downvoted pods", err)
	}

	query := `
		SELECT p.id, p.owner_id, p.name, p.slug, p.description, p.visibility, p.categories, p.tags,
			p.star_count, p.fork_count, p.view_count, p.is_verified, p.upvote_count, p.forked_from_id, p.created_at, p.updated_at, p.deleted_at
		FROM pods p
		JOIN pod_downvotes d ON p.id = d.pod_id
		WHERE d.user_id = $1 AND p.deleted_at IS NULL
		ORDER BY d.created_at DESC LIMIT $2 OFFSET $3
	`
	rows, err := r.db.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, 0, errors.Internal("failed to query downvoted pods", err)
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

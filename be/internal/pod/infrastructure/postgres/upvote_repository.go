package postgres

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"ngasihtau/internal/common/errors"
	"ngasihtau/internal/pod/domain"
)

// PodUpvoteRepository implements domain.PodUpvoteRepository using PostgreSQL.
// Follows the existing PodStarRepository pattern for consistency.
// Implements requirements 5.1, 5.2, 5.3.
type PodUpvoteRepository struct {
	db *pgxpool.Pool
}

// NewPodUpvoteRepository creates a new PodUpvoteRepository.
func NewPodUpvoteRepository(db *pgxpool.Pool) *PodUpvoteRepository {
	return &PodUpvoteRepository{db: db}
}

// Create creates a new upvote.
// Implements requirement 5.1: WHEN a user upvotes a knowledge pod.
func (r *PodUpvoteRepository) Create(ctx context.Context, upvote *domain.PodUpvote) error {
	query := `INSERT INTO pod_upvotes (user_id, pod_id, created_at) VALUES ($1, $2, $3)`
	_, err := r.db.Exec(ctx, query, upvote.UserID, upvote.PodID, upvote.CreatedAt)
	if err != nil {
		return errors.Internal("failed to create upvote", err)
	}
	return nil
}

// Delete removes an upvote.
// Implements requirement 5.2: WHEN a user removes their upvote.
func (r *PodUpvoteRepository) Delete(ctx context.Context, userID, podID uuid.UUID) error {
	query := `DELETE FROM pod_upvotes WHERE user_id = $1 AND pod_id = $2`
	result, err := r.db.Exec(ctx, query, userID, podID)
	if err != nil {
		return errors.Internal("failed to delete upvote", err)
	}
	if result.RowsAffected() == 0 {
		return errors.NotFound("upvote", "")
	}
	return nil
}

// Exists checks if an upvote exists.
// Implements requirement 5.3: THE Pod Service SHALL allow each user to upvote a knowledge pod only once.
func (r *PodUpvoteRepository) Exists(ctx context.Context, userID, podID uuid.UUID) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM pod_upvotes WHERE user_id = $1 AND pod_id = $2)`
	var exists bool
	if err := r.db.QueryRow(ctx, query, userID, podID).Scan(&exists); err != nil {
		return false, errors.Internal("failed to check upvote existence", err)
	}
	return exists, nil
}

// CountByPodID returns the upvote count for a pod.
func (r *PodUpvoteRepository) CountByPodID(ctx context.Context, podID uuid.UUID) (int, error) {
	query := `SELECT COUNT(*) FROM pod_upvotes WHERE pod_id = $1`
	var count int
	if err := r.db.QueryRow(ctx, query, podID).Scan(&count); err != nil {
		return 0, errors.Internal("failed to count upvotes", err)
	}
	return count, nil
}

// GetUpvotedPods returns paginated upvoted pods for a user.
func (r *PodUpvoteRepository) GetUpvotedPods(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*domain.Pod, int, error) {
	countQuery := `SELECT COUNT(*) FROM pod_upvotes WHERE user_id = $1`
	var total int
	if err := r.db.QueryRow(ctx, countQuery, userID).Scan(&total); err != nil {
		return nil, 0, errors.Internal("failed to count upvoted pods", err)
	}

	query := `
		SELECT p.id, p.owner_id, p.name, p.slug, p.description, p.visibility, p.categories, p.tags,
			p.star_count, p.fork_count, p.view_count, p.is_verified, p.upvote_count, p.forked_from_id, p.created_at, p.updated_at, p.deleted_at
		FROM pods p
		JOIN pod_upvotes u ON p.id = u.pod_id
		WHERE u.user_id = $1 AND p.deleted_at IS NULL
		ORDER BY u.created_at DESC LIMIT $2 OFFSET $3
	`
	rows, err := r.db.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, 0, errors.Internal("failed to query upvoted pods", err)
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

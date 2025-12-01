package postgres

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"ngasihtau/internal/common/errors"
	"ngasihtau/internal/pod/domain"
)

// PodRepository implements domain.PodRepository using PostgreSQL.
type PodRepository struct {
	db *pgxpool.Pool
}

// NewPodRepository creates a new PodRepository.
func NewPodRepository(db *pgxpool.Pool) *PodRepository {
	return &PodRepository{db: db}
}

// Create creates a new pod.
func (r *PodRepository) Create(ctx context.Context, pod *domain.Pod) error {
	query := `
		INSERT INTO pods (id, owner_id, name, slug, description, visibility, categories, tags, 
			star_count, fork_count, view_count, forked_from_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
	`
	_, err := r.db.Exec(ctx, query,
		pod.ID, pod.OwnerID, pod.Name, pod.Slug, pod.Description, pod.Visibility,
		pod.Categories, pod.Tags,
		pod.StarCount, pod.ForkCount, pod.ViewCount, pod.ForkedFromID,
		pod.CreatedAt, pod.UpdatedAt,
	)
	if err != nil {
		return errors.Internal("failed to create pod", err)
	}
	return nil
}

// FindByID finds a pod by ID.
func (r *PodRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Pod, error) {
	query := `
		SELECT id, owner_id, name, slug, description, visibility, categories, tags,
			star_count, fork_count, view_count, forked_from_id, created_at, updated_at, deleted_at
		FROM pods WHERE id = $1 AND deleted_at IS NULL
	`
	return r.scanPod(r.db.QueryRow(ctx, query, id))
}

// FindBySlug finds a pod by slug.
func (r *PodRepository) FindBySlug(ctx context.Context, slug string) (*domain.Pod, error) {
	query := `
		SELECT id, owner_id, name, slug, description, visibility, categories, tags,
			star_count, fork_count, view_count, forked_from_id, created_at, updated_at, deleted_at
		FROM pods WHERE slug = $1 AND deleted_at IS NULL
	`
	return r.scanPod(r.db.QueryRow(ctx, query, slug))
}

// FindByOwnerID finds all pods owned by a user.
func (r *PodRepository) FindByOwnerID(ctx context.Context, ownerID uuid.UUID, limit, offset int) ([]*domain.Pod, int, error) {
	countQuery := `SELECT COUNT(*) FROM pods WHERE owner_id = $1 AND deleted_at IS NULL`
	var total int
	if err := r.db.QueryRow(ctx, countQuery, ownerID).Scan(&total); err != nil {
		return nil, 0, errors.Internal("failed to count pods", err)
	}

	query := `
		SELECT id, owner_id, name, slug, description, visibility, categories, tags,
			star_count, fork_count, view_count, forked_from_id, created_at, updated_at, deleted_at
		FROM pods WHERE owner_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC LIMIT $2 OFFSET $3
	`
	rows, err := r.db.Query(ctx, query, ownerID, limit, offset)
	if err != nil {
		return nil, 0, errors.Internal("failed to query pods", err)
	}
	defer rows.Close()

	pods, err := r.scanPods(rows)
	if err != nil {
		return nil, 0, err
	}
	return pods, total, nil
}

// Update updates an existing pod.
func (r *PodRepository) Update(ctx context.Context, pod *domain.Pod) error {
	query := `
		UPDATE pods SET name = $2, description = $3, visibility = $4, categories = $5, 
			tags = $6, updated_at = $7
		WHERE id = $1 AND deleted_at IS NULL
	`
	result, err := r.db.Exec(ctx, query,
		pod.ID, pod.Name, pod.Description, pod.Visibility,
		pod.Categories, pod.Tags, pod.UpdatedAt,
	)
	if err != nil {
		return errors.Internal("failed to update pod", err)
	}
	if result.RowsAffected() == 0 {
		return errors.NotFound("pod", pod.ID.String())
	}
	return nil
}

// Delete soft-deletes a pod.
func (r *PodRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE pods SET deleted_at = $2 WHERE id = $1 AND deleted_at IS NULL`
	result, err := r.db.Exec(ctx, query, id, time.Now())
	if err != nil {
		return errors.Internal("failed to delete pod", err)
	}
	if result.RowsAffected() == 0 {
		return errors.NotFound("pod", id.String())
	}
	return nil
}

// ExistsBySlug checks if a pod with the given slug exists.
func (r *PodRepository) ExistsBySlug(ctx context.Context, slug string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM pods WHERE slug = $1 AND deleted_at IS NULL)`
	var exists bool
	if err := r.db.QueryRow(ctx, query, slug).Scan(&exists); err != nil {
		return false, errors.Internal("failed to check slug existence", err)
	}
	return exists, nil
}

// IncrementStarCount increments the star count for a pod.
func (r *PodRepository) IncrementStarCount(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE pods SET star_count = star_count + 1 WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return errors.Internal("failed to increment star count", err)
	}
	return nil
}

// DecrementStarCount decrements the star count for a pod.
func (r *PodRepository) DecrementStarCount(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE pods SET star_count = GREATEST(star_count - 1, 0) WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return errors.Internal("failed to decrement star count", err)
	}
	return nil
}

// IncrementForkCount increments the fork count for a pod.
func (r *PodRepository) IncrementForkCount(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE pods SET fork_count = fork_count + 1 WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return errors.Internal("failed to increment fork count", err)
	}
	return nil
}

// IncrementViewCount increments the view count for a pod.
func (r *PodRepository) IncrementViewCount(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE pods SET view_count = view_count + 1 WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return errors.Internal("failed to increment view count", err)
	}
	return nil
}

// Search searches pods with filters.
func (r *PodRepository) Search(ctx context.Context, query string, filters domain.PodFilters, limit, offset int) ([]*domain.Pod, int, error) {
	var conditions []string
	var args []interface{}
	argIndex := 1

	conditions = append(conditions, "deleted_at IS NULL")

	if filters.OwnerID != nil {
		conditions = append(conditions, fmt.Sprintf("owner_id = $%d", argIndex))
		args = append(args, *filters.OwnerID)
		argIndex++
	}

	if filters.Category != nil {
		conditions = append(conditions, fmt.Sprintf("$%d = ANY(categories)", argIndex))
		args = append(args, *filters.Category)
		argIndex++
	}

	if filters.Visibility != nil {
		conditions = append(conditions, fmt.Sprintf("visibility = $%d", argIndex))
		args = append(args, *filters.Visibility)
		argIndex++
	} else {
		// Default to public only if no visibility filter
		conditions = append(conditions, "visibility = 'public'")
	}

	whereClause := strings.Join(conditions, " AND ")

	// Count query
	countQuery := "SELECT COUNT(*) FROM pods WHERE " + whereClause
	var total int
	if err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, errors.Internal("failed to count pods", err)
	}

	// Add pagination args
	limitArgIdx := argIndex
	offsetArgIdx := argIndex + 1
	args = append(args, limit, offset)

	selectQuery := fmt.Sprintf(`
		SELECT id, owner_id, name, slug, description, visibility, categories, tags,
			star_count, fork_count, view_count, forked_from_id, created_at, updated_at, deleted_at
		FROM pods WHERE %s
		ORDER BY created_at DESC LIMIT $%d OFFSET $%d`, whereClause, limitArgIdx, offsetArgIdx)

	rows, err := r.db.Query(ctx, selectQuery, args...)
	if err != nil {
		return nil, 0, errors.Internal("failed to search pods", err)
	}
	defer rows.Close()

	pods, err := r.scanPods(rows)
	if err != nil {
		return nil, 0, err
	}
	return pods, total, nil
}

// GetPublicPods returns paginated public pods.
func (r *PodRepository) GetPublicPods(ctx context.Context, limit, offset int) ([]*domain.Pod, int, error) {
	countQuery := `SELECT COUNT(*) FROM pods WHERE visibility = 'public' AND deleted_at IS NULL`
	var total int
	if err := r.db.QueryRow(ctx, countQuery).Scan(&total); err != nil {
		return nil, 0, errors.Internal("failed to count pods", err)
	}

	query := `
		SELECT id, owner_id, name, slug, description, visibility, categories, tags,
			star_count, fork_count, view_count, forked_from_id, created_at, updated_at, deleted_at
		FROM pods WHERE visibility = 'public' AND deleted_at IS NULL
		ORDER BY created_at DESC LIMIT $1 OFFSET $2
	`
	rows, err := r.db.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, errors.Internal("failed to query pods", err)
	}
	defer rows.Close()

	pods, err := r.scanPods(rows)
	if err != nil {
		return nil, 0, err
	}
	return pods, total, nil
}

// scanPod scans a single pod from a row.
func (r *PodRepository) scanPod(row pgx.Row) (*domain.Pod, error) {
	var pod domain.Pod
	var categories, tags []string
	err := row.Scan(
		&pod.ID, &pod.OwnerID, &pod.Name, &pod.Slug, &pod.Description, &pod.Visibility,
		&categories, &tags,
		&pod.StarCount, &pod.ForkCount, &pod.ViewCount, &pod.ForkedFromID,
		&pod.CreatedAt, &pod.UpdatedAt, &pod.DeletedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, errors.NotFound("pod", "")
		}
		return nil, errors.Internal("failed to scan pod", err)
	}
	pod.Categories = categories
	pod.Tags = tags
	return &pod, nil
}

// scanPods scans multiple pods from rows.
func (r *PodRepository) scanPods(rows pgx.Rows) ([]*domain.Pod, error) {
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
			return nil, errors.Internal("failed to scan pod", err)
		}
		pod.Categories = categories
		pod.Tags = tags
		pods = append(pods, &pod)
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Internal("error iterating pods", err)
	}
	return pods, nil
}

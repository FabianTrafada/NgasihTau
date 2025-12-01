package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"ngasihtau/internal/material/domain"
)

// MaterialRepository implements domain.MaterialRepository using PostgreSQL.
type MaterialRepository struct {
	db DBTX
}

// NewMaterialRepository creates a new MaterialRepository.
func NewMaterialRepository(db DBTX) *MaterialRepository {
	return &MaterialRepository{db: db}
}

// Create creates a new material.
func (r *MaterialRepository) Create(ctx context.Context, material *domain.Material) error {
	query := `
		INSERT INTO materials (
			id, pod_id, uploader_id, title, description, file_type, file_url, file_size,
			current_version, status, view_count, download_count, average_rating, rating_count,
			created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
	`

	_, err := r.db.Exec(ctx, query,
		material.ID,
		material.PodID,
		material.UploaderID,
		material.Title,
		material.Description,
		material.FileType,
		material.FileURL,
		material.FileSize,
		material.CurrentVersion,
		material.Status,
		material.ViewCount,
		material.DownloadCount,
		material.AverageRating,
		material.RatingCount,
		material.CreatedAt,
		material.UpdatedAt,
	)

	return err
}

// FindByID finds a material by ID.
func (r *MaterialRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Material, error) {
	query := `
		SELECT id, pod_id, uploader_id, title, description, file_type, file_url, file_size,
			current_version, status, view_count, download_count, average_rating, rating_count,
			created_at, updated_at, deleted_at
		FROM materials
		WHERE id = $1 AND deleted_at IS NULL
	`

	var m domain.Material
	err := r.db.QueryRow(ctx, query, id).Scan(
		&m.ID,
		&m.PodID,
		&m.UploaderID,
		&m.Title,
		&m.Description,
		&m.FileType,
		&m.FileURL,
		&m.FileSize,
		&m.CurrentVersion,
		&m.Status,
		&m.ViewCount,
		&m.DownloadCount,
		&m.AverageRating,
		&m.RatingCount,
		&m.CreatedAt,
		&m.UpdatedAt,
		&m.DeletedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, errors.New("material not found")
	}

	return &m, err
}

// FindByPodID finds all materials in a pod.
func (r *MaterialRepository) FindByPodID(ctx context.Context, podID uuid.UUID, limit, offset int) ([]*domain.Material, int, error) {
	// Get total count
	countQuery := `SELECT COUNT(*) FROM materials WHERE pod_id = $1 AND deleted_at IS NULL`
	var total int
	if err := r.db.QueryRow(ctx, countQuery, podID).Scan(&total); err != nil {
		return nil, 0, err
	}

	// Get materials
	query := `
		SELECT id, pod_id, uploader_id, title, description, file_type, file_url, file_size,
			current_version, status, view_count, download_count, average_rating, rating_count,
			created_at, updated_at, deleted_at
		FROM materials
		WHERE pod_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(ctx, query, podID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var materials []*domain.Material
	for rows.Next() {
		var m domain.Material
		if err := rows.Scan(
			&m.ID,
			&m.PodID,
			&m.UploaderID,
			&m.Title,
			&m.Description,
			&m.FileType,
			&m.FileURL,
			&m.FileSize,
			&m.CurrentVersion,
			&m.Status,
			&m.ViewCount,
			&m.DownloadCount,
			&m.AverageRating,
			&m.RatingCount,
			&m.CreatedAt,
			&m.UpdatedAt,
			&m.DeletedAt,
		); err != nil {
			return nil, 0, err
		}
		materials = append(materials, &m)
	}

	return materials, total, rows.Err()
}

// FindByUploaderID finds all materials uploaded by a user.
func (r *MaterialRepository) FindByUploaderID(ctx context.Context, uploaderID uuid.UUID, limit, offset int) ([]*domain.Material, int, error) {
	// Get total count
	countQuery := `SELECT COUNT(*) FROM materials WHERE uploader_id = $1 AND deleted_at IS NULL`
	var total int
	if err := r.db.QueryRow(ctx, countQuery, uploaderID).Scan(&total); err != nil {
		return nil, 0, err
	}

	// Get materials
	query := `
		SELECT id, pod_id, uploader_id, title, description, file_type, file_url, file_size,
			current_version, status, view_count, download_count, average_rating, rating_count,
			created_at, updated_at, deleted_at
		FROM materials
		WHERE uploader_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(ctx, query, uploaderID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var materials []*domain.Material
	for rows.Next() {
		var m domain.Material
		if err := rows.Scan(
			&m.ID,
			&m.PodID,
			&m.UploaderID,
			&m.Title,
			&m.Description,
			&m.FileType,
			&m.FileURL,
			&m.FileSize,
			&m.CurrentVersion,
			&m.Status,
			&m.ViewCount,
			&m.DownloadCount,
			&m.AverageRating,
			&m.RatingCount,
			&m.CreatedAt,
			&m.UpdatedAt,
			&m.DeletedAt,
		); err != nil {
			return nil, 0, err
		}
		materials = append(materials, &m)
	}

	return materials, total, rows.Err()
}

// Update updates an existing material.
func (r *MaterialRepository) Update(ctx context.Context, material *domain.Material) error {
	query := `
		UPDATE materials
		SET title = $2, description = $3, file_url = $4, file_size = $5,
			current_version = $6, status = $7, updated_at = $8
		WHERE id = $1 AND deleted_at IS NULL
	`

	_, err := r.db.Exec(ctx, query,
		material.ID,
		material.Title,
		material.Description,
		material.FileURL,
		material.FileSize,
		material.CurrentVersion,
		material.Status,
		time.Now(),
	)

	return err
}

// Delete soft-deletes a material.
func (r *MaterialRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE materials SET deleted_at = $2 WHERE id = $1 AND deleted_at IS NULL`
	_, err := r.db.Exec(ctx, query, id, time.Now())
	return err
}

// UpdateStatus updates a material's processing status.
func (r *MaterialRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.MaterialStatus) error {
	query := `UPDATE materials SET status = $2, updated_at = $3 WHERE id = $1 AND deleted_at IS NULL`
	_, err := r.db.Exec(ctx, query, id, status, time.Now())
	return err
}

// IncrementViewCount increments the view count for a material.
func (r *MaterialRepository) IncrementViewCount(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE materials SET view_count = view_count + 1 WHERE id = $1 AND deleted_at IS NULL`
	_, err := r.db.Exec(ctx, query, id)
	return err
}

// IncrementDownloadCount increments the download count for a material.
func (r *MaterialRepository) IncrementDownloadCount(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE materials SET download_count = download_count + 1 WHERE id = $1 AND deleted_at IS NULL`
	_, err := r.db.Exec(ctx, query, id)
	return err
}

// UpdateRatingStats updates the average rating and rating count.
func (r *MaterialRepository) UpdateRatingStats(ctx context.Context, id uuid.UUID, avgRating float64, ratingCount int) error {
	query := `UPDATE materials SET average_rating = $2, rating_count = $3, updated_at = $4 WHERE id = $1 AND deleted_at IS NULL`
	_, err := r.db.Exec(ctx, query, id, avgRating, ratingCount, time.Now())
	return err
}

// IncrementVersion increments the current version number.
func (r *MaterialRepository) IncrementVersion(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE materials SET current_version = current_version + 1, updated_at = $2 WHERE id = $1 AND deleted_at IS NULL`
	_, err := r.db.Exec(ctx, query, id, time.Now())
	return err
}

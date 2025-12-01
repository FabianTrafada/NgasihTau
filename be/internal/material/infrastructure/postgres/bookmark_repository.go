package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"ngasihtau/internal/material/domain"
)

// BookmarkRepository implements domain.BookmarkRepository using PostgreSQL.
type BookmarkRepository struct {
	db DBTX
}

// NewBookmarkRepository creates a new BookmarkRepository.
func NewBookmarkRepository(db DBTX) *BookmarkRepository {
	return &BookmarkRepository{db: db}
}

// Create creates a new bookmark.
func (r *BookmarkRepository) Create(ctx context.Context, bookmark *domain.Bookmark) error {
	query := `
		INSERT INTO bookmarks (id, user_id, material_id, folder, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`

	_, err := r.db.Exec(ctx, query,
		bookmark.ID,
		bookmark.UserID,
		bookmark.MaterialID,
		bookmark.Folder,
		bookmark.CreatedAt,
	)

	return err
}

// FindByID finds a bookmark by ID.
func (r *BookmarkRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Bookmark, error) {
	query := `
		SELECT id, user_id, material_id, folder, created_at
		FROM bookmarks
		WHERE id = $1
	`

	var b domain.Bookmark
	err := r.db.QueryRow(ctx, query, id).Scan(
		&b.ID,
		&b.UserID,
		&b.MaterialID,
		&b.Folder,
		&b.CreatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, errors.New("bookmark not found")
	}

	return &b, err
}

// FindByUserAndMaterial finds a bookmark by user and material.
func (r *BookmarkRepository) FindByUserAndMaterial(ctx context.Context, userID, materialID uuid.UUID) (*domain.Bookmark, error) {
	query := `
		SELECT id, user_id, material_id, folder, created_at
		FROM bookmarks
		WHERE user_id = $1 AND material_id = $2
	`

	var b domain.Bookmark
	err := r.db.QueryRow(ctx, query, userID, materialID).Scan(
		&b.ID,
		&b.UserID,
		&b.MaterialID,
		&b.Folder,
		&b.CreatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}

	return &b, err
}

// FindByUserID finds all bookmarks for a user.
func (r *BookmarkRepository) FindByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*domain.Bookmark, int, error) {
	// Get total count
	countQuery := `SELECT COUNT(*) FROM bookmarks WHERE user_id = $1`
	var total int
	if err := r.db.QueryRow(ctx, countQuery, userID).Scan(&total); err != nil {
		return nil, 0, err
	}

	// Get bookmarks
	query := `
		SELECT id, user_id, material_id, folder, created_at
		FROM bookmarks
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var bookmarks []*domain.Bookmark
	for rows.Next() {
		var b domain.Bookmark
		if err := rows.Scan(
			&b.ID,
			&b.UserID,
			&b.MaterialID,
			&b.Folder,
			&b.CreatedAt,
		); err != nil {
			return nil, 0, err
		}
		bookmarks = append(bookmarks, &b)
	}

	return bookmarks, total, rows.Err()
}

// FindByUserIDWithMaterials finds all bookmarks for a user with material details.
func (r *BookmarkRepository) FindByUserIDWithMaterials(ctx context.Context, userID uuid.UUID, folder *string, limit, offset int) ([]*domain.MaterialWithUploader, int, error) {
	// Build query based on folder filter
	var countQuery string
	var query string
	var args []any

	if folder != nil {
		countQuery = `
			SELECT COUNT(*) 
			FROM bookmarks b
			JOIN materials m ON b.material_id = m.id
			WHERE b.user_id = $1 AND b.folder = $2 AND m.deleted_at IS NULL
		`
		query = `
			SELECT m.id, m.pod_id, m.uploader_id, m.title, m.description, m.file_type, m.file_url, m.file_size,
				m.current_version, m.status, m.view_count, m.download_count, m.average_rating, m.rating_count,
				m.created_at, m.updated_at, m.deleted_at
			FROM bookmarks b
			JOIN materials m ON b.material_id = m.id
			WHERE b.user_id = $1 AND b.folder = $2 AND m.deleted_at IS NULL
			ORDER BY b.created_at DESC
			LIMIT $3 OFFSET $4
		`
		args = []any{userID, *folder}
	} else {
		countQuery = `
			SELECT COUNT(*) 
			FROM bookmarks b
			JOIN materials m ON b.material_id = m.id
			WHERE b.user_id = $1 AND m.deleted_at IS NULL
		`
		query = `
			SELECT m.id, m.pod_id, m.uploader_id, m.title, m.description, m.file_type, m.file_url, m.file_size,
				m.current_version, m.status, m.view_count, m.download_count, m.average_rating, m.rating_count,
				m.created_at, m.updated_at, m.deleted_at
			FROM bookmarks b
			JOIN materials m ON b.material_id = m.id
			WHERE b.user_id = $1 AND m.deleted_at IS NULL
			ORDER BY b.created_at DESC
			LIMIT $2 OFFSET $3
		`
		args = []any{userID}
	}

	// Get total count
	var total int
	if folder != nil {
		if err := r.db.QueryRow(ctx, countQuery, userID, *folder).Scan(&total); err != nil {
			return nil, 0, err
		}
		args = append(args, limit, offset)
	} else {
		if err := r.db.QueryRow(ctx, countQuery, userID).Scan(&total); err != nil {
			return nil, 0, err
		}
		args = append(args, limit, offset)
	}

	// Get materials
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var materials []*domain.MaterialWithUploader
	for rows.Next() {
		var m domain.MaterialWithUploader
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
		// Uploader info would be populated via gRPC call to User Service
		m.UploaderName = "Unknown"
		materials = append(materials, &m)
	}

	return materials, total, rows.Err()
}

// Delete removes a bookmark.
func (r *BookmarkRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM bookmarks WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}

// DeleteByUserAndMaterial removes a bookmark by user and material.
func (r *BookmarkRepository) DeleteByUserAndMaterial(ctx context.Context, userID, materialID uuid.UUID) error {
	query := `DELETE FROM bookmarks WHERE user_id = $1 AND material_id = $2`
	_, err := r.db.Exec(ctx, query, userID, materialID)
	return err
}

// Exists checks if a bookmark exists.
func (r *BookmarkRepository) Exists(ctx context.Context, userID, materialID uuid.UUID) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM bookmarks WHERE user_id = $1 AND material_id = $2)`
	var exists bool
	err := r.db.QueryRow(ctx, query, userID, materialID).Scan(&exists)
	return exists, err
}

// GetFolders returns all unique folders for a user.
func (r *BookmarkRepository) GetFolders(ctx context.Context, userID uuid.UUID) ([]string, error) {
	query := `
		SELECT DISTINCT folder 
		FROM bookmarks 
		WHERE user_id = $1 AND folder IS NOT NULL
		ORDER BY folder
	`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var folders []string
	for rows.Next() {
		var folder string
		if err := rows.Scan(&folder); err != nil {
			return nil, err
		}
		folders = append(folders, folder)
	}

	return folders, rows.Err()
}

package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"ngasihtau/internal/material/domain"
)

// MaterialVersionRepository implements domain.MaterialVersionRepository using PostgreSQL.
type MaterialVersionRepository struct {
	db DBTX
}

// NewMaterialVersionRepository creates a new MaterialVersionRepository.
func NewMaterialVersionRepository(db DBTX) *MaterialVersionRepository {
	return &MaterialVersionRepository{db: db}
}

// Create creates a new material version.
func (r *MaterialVersionRepository) Create(ctx context.Context, version *domain.MaterialVersion) error {
	query := `
		INSERT INTO material_versions (id, material_id, version, file_url, file_size, uploader_id, changelog, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := r.db.Exec(ctx, query,
		version.ID,
		version.MaterialID,
		version.Version,
		version.FileURL,
		version.FileSize,
		version.UploaderID,
		version.Changelog,
		version.CreatedAt,
	)

	return err
}

// FindByID finds a version by ID.
func (r *MaterialVersionRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.MaterialVersion, error) {
	query := `
		SELECT id, material_id, version, file_url, file_size, uploader_id, changelog, created_at
		FROM material_versions
		WHERE id = $1
	`

	var v domain.MaterialVersion
	err := r.db.QueryRow(ctx, query, id).Scan(
		&v.ID,
		&v.MaterialID,
		&v.Version,
		&v.FileURL,
		&v.FileSize,
		&v.UploaderID,
		&v.Changelog,
		&v.CreatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, errors.New("version not found")
	}

	return &v, err
}

// FindByMaterialID finds all versions for a material.
func (r *MaterialVersionRepository) FindByMaterialID(ctx context.Context, materialID uuid.UUID) ([]*domain.MaterialVersion, error) {
	query := `
		SELECT id, material_id, version, file_url, file_size, uploader_id, changelog, created_at
		FROM material_versions
		WHERE material_id = $1
		ORDER BY version DESC
	`

	rows, err := r.db.Query(ctx, query, materialID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var versions []*domain.MaterialVersion
	for rows.Next() {
		var v domain.MaterialVersion
		if err := rows.Scan(
			&v.ID,
			&v.MaterialID,
			&v.Version,
			&v.FileURL,
			&v.FileSize,
			&v.UploaderID,
			&v.Changelog,
			&v.CreatedAt,
		); err != nil {
			return nil, err
		}
		versions = append(versions, &v)
	}

	return versions, rows.Err()
}

// FindByMaterialIDAndVersion finds a specific version of a material.
func (r *MaterialVersionRepository) FindByMaterialIDAndVersion(ctx context.Context, materialID uuid.UUID, version int) (*domain.MaterialVersion, error) {
	query := `
		SELECT id, material_id, version, file_url, file_size, uploader_id, changelog, created_at
		FROM material_versions
		WHERE material_id = $1 AND version = $2
	`

	var v domain.MaterialVersion
	err := r.db.QueryRow(ctx, query, materialID, version).Scan(
		&v.ID,
		&v.MaterialID,
		&v.Version,
		&v.FileURL,
		&v.FileSize,
		&v.UploaderID,
		&v.Changelog,
		&v.CreatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, errors.New("version not found")
	}

	return &v, err
}

// GetLatestVersion gets the latest version number for a material.
func (r *MaterialVersionRepository) GetLatestVersion(ctx context.Context, materialID uuid.UUID) (int, error) {
	query := `SELECT COALESCE(MAX(version), 0) FROM material_versions WHERE material_id = $1`

	var version int
	err := r.db.QueryRow(ctx, query, materialID).Scan(&version)
	return version, err
}

// DeleteByMaterialID deletes all versions for a material.
func (r *MaterialVersionRepository) DeleteByMaterialID(ctx context.Context, materialID uuid.UUID) error {
	query := `DELETE FROM material_versions WHERE material_id = $1`
	_, err := r.db.Exec(ctx, query, materialID)
	return err
}

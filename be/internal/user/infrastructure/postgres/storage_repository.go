// Package postgres provides PostgreSQL implementations of the User Service repositories.
package postgres

import (
	"context"

	"github.com/google/uuid"

	"ngasihtau/internal/common/errors"
	"ngasihtau/internal/user/domain"
)

// StorageRepository implements domain.StorageRepository using PostgreSQL.
// It calculates storage usage by summing file_size from the materials table.
// Implements requirements 5.1, 5.3, 4.3.
type StorageRepository struct {
	db DBTX
}

// NewStorageRepository creates a new StorageRepository.
func NewStorageRepository(db DBTX) *StorageRepository {
	return &StorageRepository{db: db}
}

// GetUserStorageUsage returns total bytes used by a user.
// It calculates storage by summing file_size of all non-deleted materials uploaded by the user.
// Uses int64 to support files larger than 2GB (requirement 5.3).
// Only counts non-deleted materials (requirement 4.3).
func (r *StorageRepository) GetUserStorageUsage(ctx context.Context, userID uuid.UUID) (int64, error) {
	query := `
		SELECT COALESCE(SUM(file_size), 0)
		FROM materials
		WHERE uploader_id = $1 AND deleted_at IS NULL
	`

	var totalBytes int64
	err := r.db.QueryRow(ctx, query, userID).Scan(&totalBytes)
	if err != nil {
		return 0, errors.Internal("failed to calculate storage usage", err)
	}

	return totalBytes, nil
}

// Compile-time interface implementation check.
var _ domain.StorageRepository = (*StorageRepository)(nil)

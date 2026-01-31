package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"ngasihtau/internal/offline/domain"
)

// CEKRepository implements domain.CEKRepository using PostgreSQL.
type CEKRepository struct {
	db DBTX
}

// NewCEKRepository creates a new CEKRepository.
func NewCEKRepository(db DBTX) *CEKRepository {
	return &CEKRepository{db: db}
}

// Create creates a new Content Encryption Key record.
func (r *CEKRepository) Create(ctx context.Context, cek *domain.ContentEncryptionKey) error {
	query := `
		INSERT INTO offline_ceks (id, user_id, material_id, device_id, encrypted_key, key_version, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err := r.db.Exec(ctx, query,
		cek.ID,
		cek.UserID,
		cek.MaterialID,
		cek.DeviceID,
		cek.EncryptedKey,
		cek.KeyVersion,
		cek.CreatedAt,
	)

	return err
}

// FindByID finds a CEK by its ID.
func (r *CEKRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.ContentEncryptionKey, error) {
	query := `
		SELECT id, user_id, material_id, device_id, encrypted_key, key_version, created_at
		FROM offline_ceks
		WHERE id = $1
	`

	var cek domain.ContentEncryptionKey
	err := r.db.QueryRow(ctx, query, id).Scan(
		&cek.ID,
		&cek.UserID,
		&cek.MaterialID,
		&cek.DeviceID,
		&cek.EncryptedKey,
		&cek.KeyVersion,
		&cek.CreatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return &cek, nil
}

// FindByComposite finds a CEK by the composite key (user_id, material_id, device_id).
func (r *CEKRepository) FindByComposite(ctx context.Context, userID, materialID, deviceID uuid.UUID) (*domain.ContentEncryptionKey, error) {
	query := `
		SELECT id, user_id, material_id, device_id, encrypted_key, key_version, created_at
		FROM offline_ceks
		WHERE user_id = $1 AND material_id = $2 AND device_id = $3
	`

	var cek domain.ContentEncryptionKey
	err := r.db.QueryRow(ctx, query, userID, materialID, deviceID).Scan(
		&cek.ID,
		&cek.UserID,
		&cek.MaterialID,
		&cek.DeviceID,
		&cek.EncryptedKey,
		&cek.KeyVersion,
		&cek.CreatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return &cek, nil
}

// DeleteByDeviceID deletes all CEKs associated with a device.
func (r *CEKRepository) DeleteByDeviceID(ctx context.Context, deviceID uuid.UUID) error {
	query := `DELETE FROM offline_ceks WHERE device_id = $1`
	_, err := r.db.Exec(ctx, query, deviceID)
	return err
}

// DeleteByMaterialID deletes all CEKs associated with a material.
func (r *CEKRepository) DeleteByMaterialID(ctx context.Context, materialID uuid.UUID) error {
	query := `DELETE FROM offline_ceks WHERE material_id = $1`
	_, err := r.db.Exec(ctx, query, materialID)
	return err
}

// FindByKeyVersion finds all CEKs with a specific key version.
// Used for key rotation operations.
func (r *CEKRepository) FindByKeyVersion(ctx context.Context, keyVersion int) ([]*domain.ContentEncryptionKey, error) {
	query := `
		SELECT id, user_id, material_id, device_id, encrypted_key, key_version, created_at
		FROM offline_ceks
		WHERE key_version = $1
		ORDER BY created_at ASC
	`

	rows, err := r.db.Query(ctx, query, keyVersion)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ceks []*domain.ContentEncryptionKey
	for rows.Next() {
		var cek domain.ContentEncryptionKey
		if err := rows.Scan(
			&cek.ID,
			&cek.UserID,
			&cek.MaterialID,
			&cek.DeviceID,
			&cek.EncryptedKey,
			&cek.KeyVersion,
			&cek.CreatedAt,
		); err != nil {
			return nil, err
		}
		ceks = append(ceks, &cek)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return ceks, nil
}

// UpdateKeyVersion updates the encrypted key and key version for a CEK.
// Used during key rotation to re-encrypt CEKs with a new KEK.
func (r *CEKRepository) UpdateKeyVersion(ctx context.Context, id uuid.UUID, encryptedKey []byte, keyVersion int) error {
	query := `
		UPDATE offline_ceks
		SET encrypted_key = $2, key_version = $3
		WHERE id = $1
	`

	result, err := r.db.Exec(ctx, query, id, encryptedKey, keyVersion)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return errors.New("cek not found")
	}

	return nil
}

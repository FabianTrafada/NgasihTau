package postgres

import (
	"context"
	"time"

	"github.com/google/uuid"

	"ngasihtau/internal/common/errors"
	"ngasihtau/internal/user/domain"
)

// BackupCodeRepository implements domain.BackupCodeRepository using PostgreSQL.
type BackupCodeRepository struct {
	db DBTX
}

// NewBackupCodeRepository creates a new BackupCodeRepository.
func NewBackupCodeRepository(db DBTX) *BackupCodeRepository {
	return &BackupCodeRepository{db: db}
}

// CreateBatch creates multiple backup codes for a user.
func (r *BackupCodeRepository) CreateBatch(ctx context.Context, codes []*domain.BackupCode) error {
	query := `
		INSERT INTO backup_codes (id, user_id, code_hash, used, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`

	now := time.Now().UTC()
	for _, code := range codes {
		if code.ID == uuid.Nil {
			code.ID = uuid.New()
		}
		code.CreatedAt = now

		_, err := r.db.Exec(ctx, query,
			code.ID,
			code.UserID,
			code.CodeHash,
			code.Used,
			code.CreatedAt,
		)
		if err != nil {
			return errors.Internal("failed to create backup code", err)
		}
	}

	return nil
}

// FindByUserID finds all backup codes for a user.
func (r *BackupCodeRepository) FindByUserID(ctx context.Context, userID uuid.UUID) ([]*domain.BackupCode, error) {
	query := `
		SELECT id, user_id, code_hash, used, created_at
		FROM backup_codes
		WHERE user_id = $1
		ORDER BY created_at ASC
	`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, errors.Internal("failed to find backup codes", err)
	}
	defer rows.Close()

	var codes []*domain.BackupCode
	for rows.Next() {
		code := &domain.BackupCode{}
		err := rows.Scan(
			&code.ID,
			&code.UserID,
			&code.CodeHash,
			&code.Used,
			&code.CreatedAt,
		)
		if err != nil {
			return nil, errors.Internal("failed to scan backup code", err)
		}
		codes = append(codes, code)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.Internal("error iterating backup codes", err)
	}

	return codes, nil
}

// FindUnusedByUserID finds all unused backup codes for a user.
func (r *BackupCodeRepository) FindUnusedByUserID(ctx context.Context, userID uuid.UUID) ([]*domain.BackupCode, error) {
	query := `
		SELECT id, user_id, code_hash, used, created_at
		FROM backup_codes
		WHERE user_id = $1 AND used = false
		ORDER BY created_at ASC
	`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, errors.Internal("failed to find unused backup codes", err)
	}
	defer rows.Close()

	var codes []*domain.BackupCode
	for rows.Next() {
		code := &domain.BackupCode{}
		err := rows.Scan(
			&code.ID,
			&code.UserID,
			&code.CodeHash,
			&code.Used,
			&code.CreatedAt,
		)
		if err != nil {
			return nil, errors.Internal("failed to scan backup code", err)
		}
		codes = append(codes, code)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.Internal("error iterating backup codes", err)
	}

	return codes, nil
}

// MarkAsUsed marks a backup code as used.
func (r *BackupCodeRepository) MarkAsUsed(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE backup_codes SET used = true WHERE id = $1 AND used = false`

	tag, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return errors.Internal("failed to mark backup code as used", err)
	}

	if tag.RowsAffected() == 0 {
		return errors.NotFoundMsg("backup code not found or already used")
	}

	return nil
}

// DeleteAllByUserID removes all backup codes for a user.
func (r *BackupCodeRepository) DeleteAllByUserID(ctx context.Context, userID uuid.UUID) error {
	query := `DELETE FROM backup_codes WHERE user_id = $1`

	_, err := r.db.Exec(ctx, query, userID)
	if err != nil {
		return errors.Internal("failed to delete backup codes", err)
	}

	return nil
}

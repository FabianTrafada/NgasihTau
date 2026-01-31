// Package postgres provides PostgreSQL implementations of the User Service repositories.
package postgres

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"ngasihtau/internal/common/errors"
	"ngasihtau/internal/user/domain"
)

// TeacherVerificationRepository implements domain.TeacherVerificationRepository using PostgreSQL.
type TeacherVerificationRepository struct {
	db DBTX
}

// NewTeacherVerificationRepository creates a new TeacherVerificationRepository.
func NewTeacherVerificationRepository(db DBTX) *TeacherVerificationRepository {
	return &TeacherVerificationRepository{db: db}
}

// Create creates a new teacher verification request in the database.
func (r *TeacherVerificationRepository) Create(ctx context.Context, verification *domain.TeacherVerification) error {
	query := `
		INSERT INTO teacher_verifications (id, user_id, full_name, id_number, credential_type, 
			document_ref, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	now := time.Now().UTC()
	if verification.ID == uuid.Nil {
		verification.ID = uuid.New()
	}
	verification.CreatedAt = now
	verification.UpdatedAt = now

	_, err := r.db.Exec(ctx, query,
		verification.ID,
		verification.UserID,
		verification.FullName,
		verification.IDNumber,
		verification.CredentialType,
		verification.DocumentRef,
		verification.Status,
		verification.CreatedAt,
		verification.UpdatedAt,
	)
	if err != nil {
		return errors.Internal("failed to create teacher verification", err)
	}

	return nil
}

// FindByID finds a teacher verification by ID.
func (r *TeacherVerificationRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.TeacherVerification, error) {
	query := `
		SELECT id, user_id, full_name, id_number, credential_type, document_ref, 
			status, reviewed_by, reviewed_at, rejection_reason, created_at, updated_at
		FROM teacher_verifications
		WHERE id = $1
	`

	verification := &domain.TeacherVerification{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&verification.ID,
		&verification.UserID,
		&verification.FullName,
		&verification.IDNumber,
		&verification.CredentialType,
		&verification.DocumentRef,
		&verification.Status,
		&verification.ReviewedBy,
		&verification.ReviewedAt,
		&verification.RejectionReason,
		&verification.CreatedAt,
		&verification.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, errors.NotFound("teacher_verification", id.String())
	}
	if err != nil {
		return nil, errors.Internal("failed to find teacher verification", err)
	}

	return verification, nil
}

// FindByUserID finds a teacher verification by user ID.
// Returns the most recent verification for the user.
func (r *TeacherVerificationRepository) FindByUserID(ctx context.Context, userID uuid.UUID) (*domain.TeacherVerification, error) {
	query := `
		SELECT id, user_id, full_name, id_number, credential_type, document_ref, 
			status, reviewed_by, reviewed_at, rejection_reason, created_at, updated_at
		FROM teacher_verifications
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT 1
	`

	verification := &domain.TeacherVerification{}
	err := r.db.QueryRow(ctx, query, userID).Scan(
		&verification.ID,
		&verification.UserID,
		&verification.FullName,
		&verification.IDNumber,
		&verification.CredentialType,
		&verification.DocumentRef,
		&verification.Status,
		&verification.ReviewedBy,
		&verification.ReviewedAt,
		&verification.RejectionReason,
		&verification.CreatedAt,
		&verification.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, errors.NotFound("teacher_verification", userID.String())
	}
	if err != nil {
		return nil, errors.Internal("failed to find teacher verification", err)
	}

	return verification, nil
}

// FindPending returns paginated pending verification requests.
// Used by admins to review pending verifications.
func (r *TeacherVerificationRepository) FindPending(ctx context.Context, limit, offset int) ([]*domain.TeacherVerification, int, error) {
	// Get total count of pending verifications
	countQuery := `SELECT COUNT(*) FROM teacher_verifications WHERE status = 'pending'`
	var total int
	err := r.db.QueryRow(ctx, countQuery).Scan(&total)
	if err != nil {
		return nil, 0, errors.Internal("failed to count pending verifications", err)
	}

	// Get paginated pending verifications
	query := `
		SELECT id, user_id, full_name, id_number, credential_type, document_ref, 
			status, reviewed_by, reviewed_at, rejection_reason, created_at, updated_at
		FROM teacher_verifications
		WHERE status = 'pending'
		ORDER BY created_at ASC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, errors.Internal("failed to find pending verifications", err)
	}
	defer rows.Close()

	var verifications []*domain.TeacherVerification
	for rows.Next() {
		verification := &domain.TeacherVerification{}
		err := rows.Scan(
			&verification.ID,
			&verification.UserID,
			&verification.FullName,
			&verification.IDNumber,
			&verification.CredentialType,
			&verification.DocumentRef,
			&verification.Status,
			&verification.ReviewedBy,
			&verification.ReviewedAt,
			&verification.RejectionReason,
			&verification.CreatedAt,
			&verification.UpdatedAt,
		)
		if err != nil {
			return nil, 0, errors.Internal("failed to scan teacher verification", err)
		}
		verifications = append(verifications, verification)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, errors.Internal("failed to iterate verifications", err)
	}

	return verifications, total, nil
}

// Update updates an existing teacher verification.
func (r *TeacherVerificationRepository) Update(ctx context.Context, verification *domain.TeacherVerification) error {
	query := `
		UPDATE teacher_verifications
		SET full_name = $2, id_number = $3, credential_type = $4, document_ref = $5,
			status = $6, reviewed_by = $7, reviewed_at = $8, rejection_reason = $9, updated_at = $10
		WHERE id = $1
	`

	verification.UpdatedAt = time.Now().UTC()

	tag, err := r.db.Exec(ctx, query,
		verification.ID,
		verification.FullName,
		verification.IDNumber,
		verification.CredentialType,
		verification.DocumentRef,
		verification.Status,
		verification.ReviewedBy,
		verification.ReviewedAt,
		verification.RejectionReason,
		verification.UpdatedAt,
	)
	if err != nil {
		return errors.Internal("failed to update teacher verification", err)
	}

	if tag.RowsAffected() == 0 {
		return errors.NotFound("teacher_verification", verification.ID.String())
	}

	return nil
}

// UpdateStatus updates the status of a verification request.
// Used when approving or rejecting a verification.
func (r *TeacherVerificationRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.VerificationStatus, reviewedBy uuid.UUID, reason *string) error {
	query := `
		UPDATE teacher_verifications
		SET status = $2, reviewed_by = $3, reviewed_at = $4, rejection_reason = $5, updated_at = $4
		WHERE id = $1
	`

	now := time.Now().UTC()
	tag, err := r.db.Exec(ctx, query, id, status, reviewedBy, now, reason)
	if err != nil {
		return errors.Internal("failed to update verification status", err)
	}

	if tag.RowsAffected() == 0 {
		return errors.NotFound("teacher_verification", id.String())
	}

	return nil
}

// ExistsByUserID checks if a verification request exists for a user.
func (r *TeacherVerificationRepository) ExistsByUserID(ctx context.Context, userID uuid.UUID) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM teacher_verifications WHERE user_id = $1)`

	var exists bool
	err := r.db.QueryRow(ctx, query, userID).Scan(&exists)
	if err != nil {
		return false, errors.Internal("failed to check verification existence", err)
	}

	return exists, nil
}

// ExistsPendingByUserID checks if a pending verification request exists for a user.
func (r *TeacherVerificationRepository) ExistsPendingByUserID(ctx context.Context, userID uuid.UUID) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM teacher_verifications WHERE user_id = $1 AND status = 'pending')`

	var exists bool
	err := r.db.QueryRow(ctx, query, userID).Scan(&exists)
	if err != nil {
		return false, errors.Internal("failed to check pending verification existence", err)
	}

	return exists, nil
}

// Package postgres provides PostgreSQL implementations of domain repositories.
package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"ngasihtau/internal/offline/domain"
)

// EncryptionJobRepository implements domain.EncryptionJobRepository using PostgreSQL.
type EncryptionJobRepository struct {
	db DBTX
}

// NewEncryptionJobRepository creates a new EncryptionJobRepository.
func NewEncryptionJobRepository(db DBTX) *EncryptionJobRepository {
	return &EncryptionJobRepository{db: db}
}

// Create creates a new encryption job.
func (r *EncryptionJobRepository) Create(ctx context.Context, job *domain.EncryptionJob) error {
	query := `
		INSERT INTO offline_encryption_jobs (id, material_id, user_id, device_id, license_id, priority, status, retry_count, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	_, err := r.db.Exec(ctx, query,
		job.ID,
		job.MaterialID,
		job.UserID,
		job.DeviceID,
		job.LicenseID,
		job.Priority,
		string(job.Status),
		job.RetryCount,
		job.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create encryption job: %w", err)
	}

	return nil
}

// FindByID finds an encryption job by ID.
func (r *EncryptionJobRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.EncryptionJob, error) {
	query := `
		SELECT id, material_id, user_id, device_id, license_id, priority, status, error, retry_count, created_at, started_at, completed_at
		FROM offline_encryption_jobs
		WHERE id = $1
	`

	job, err := r.scanJob(r.db.QueryRow(ctx, query, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.NewOfflineError(domain.ErrCodeJobNotFound, "encryption job not found")
		}
		return nil, fmt.Errorf("failed to find encryption job: %w", err)
	}

	return job, nil
}

// FindByMaterialID finds all encryption jobs for a material.
func (r *EncryptionJobRepository) FindByMaterialID(ctx context.Context, materialID uuid.UUID) ([]*domain.EncryptionJob, error) {
	query := `
		SELECT id, material_id, user_id, device_id, license_id, priority, status, error, retry_count, created_at, started_at, completed_at
		FROM offline_encryption_jobs
		WHERE material_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(ctx, query, materialID)
	if err != nil {
		return nil, fmt.Errorf("failed to find encryption jobs by material: %w", err)
	}
	defer rows.Close()

	return r.scanJobs(rows)
}

// FindByLicenseID finds an encryption job by license ID.
func (r *EncryptionJobRepository) FindByLicenseID(ctx context.Context, licenseID uuid.UUID) (*domain.EncryptionJob, error) {
	query := `
		SELECT id, material_id, user_id, device_id, license_id, priority, status, error, retry_count, created_at, started_at, completed_at
		FROM offline_encryption_jobs
		WHERE license_id = $1
		ORDER BY created_at DESC
		LIMIT 1
	`

	job, err := r.scanJob(r.db.QueryRow(ctx, query, licenseID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.NewOfflineError(domain.ErrCodeJobNotFound, "encryption job not found")
		}
		return nil, fmt.Errorf("failed to find encryption job by license: %w", err)
	}

	return job, nil
}

// FindPending finds pending jobs ordered by priority and creation time.
func (r *EncryptionJobRepository) FindPending(ctx context.Context, limit int) ([]*domain.EncryptionJob, error) {
	query := `
		SELECT id, material_id, user_id, device_id, license_id, priority, status, error, retry_count, created_at, started_at, completed_at
		FROM offline_encryption_jobs
		WHERE status = $1
		ORDER BY priority ASC, created_at ASC
		LIMIT $2
	`

	rows, err := r.db.Query(ctx, query, string(domain.JobStatusPending), limit)
	if err != nil {
		return nil, fmt.Errorf("failed to find pending jobs: %w", err)
	}
	defer rows.Close()

	return r.scanJobs(rows)
}

// UpdateStatus updates the status of an encryption job.
func (r *EncryptionJobRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.JobStatus) error {
	query := `UPDATE offline_encryption_jobs SET status = $1 WHERE id = $2`

	result, err := r.db.Exec(ctx, query, string(status), id)
	if err != nil {
		return fmt.Errorf("failed to update job status: %w", err)
	}

	if result.RowsAffected() == 0 {
		return domain.NewOfflineError(domain.ErrCodeJobNotFound, "encryption job not found")
	}

	return nil
}

// UpdateStarted marks a job as started.
func (r *EncryptionJobRepository) UpdateStarted(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE offline_encryption_jobs SET status = $1, started_at = $2 WHERE id = $3`

	result, err := r.db.Exec(ctx, query, string(domain.JobStatusProcessing), time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to update job started: %w", err)
	}

	if result.RowsAffected() == 0 {
		return domain.NewOfflineError(domain.ErrCodeJobNotFound, "encryption job not found")
	}

	return nil
}

// UpdateCompleted marks a job as completed.
func (r *EncryptionJobRepository) UpdateCompleted(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE offline_encryption_jobs SET status = $1, completed_at = $2 WHERE id = $3`

	result, err := r.db.Exec(ctx, query, string(domain.JobStatusCompleted), time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to update job completed: %w", err)
	}

	if result.RowsAffected() == 0 {
		return domain.NewOfflineError(domain.ErrCodeJobNotFound, "encryption job not found")
	}

	return nil
}

// UpdateFailed marks a job as failed with an error message.
func (r *EncryptionJobRepository) UpdateFailed(ctx context.Context, id uuid.UUID, errorMsg string) error {
	query := `UPDATE offline_encryption_jobs SET status = $1, error = $2, completed_at = $3 WHERE id = $4`

	result, err := r.db.Exec(ctx, query, string(domain.JobStatusFailed), errorMsg, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to update job failed: %w", err)
	}

	if result.RowsAffected() == 0 {
		return domain.NewOfflineError(domain.ErrCodeJobNotFound, "encryption job not found")
	}

	return nil
}

// IncrementRetryCount increments the retry count for a job.
func (r *EncryptionJobRepository) IncrementRetryCount(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE offline_encryption_jobs SET retry_count = retry_count + 1 WHERE id = $1`

	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to increment retry count: %w", err)
	}

	if result.RowsAffected() == 0 {
		return domain.NewOfflineError(domain.ErrCodeJobNotFound, "encryption job not found")
	}

	return nil
}

// DeleteOldCompleted deletes completed jobs older than the specified time.
func (r *EncryptionJobRepository) DeleteOldCompleted(ctx context.Context, olderThan time.Time) error {
	query := `DELETE FROM offline_encryption_jobs WHERE status = $1 AND completed_at < $2`

	_, err := r.db.Exec(ctx, query, string(domain.JobStatusCompleted), olderThan)
	if err != nil {
		return fmt.Errorf("failed to delete old completed jobs: %w", err)
	}

	return nil
}

// scanJob scans a single job row.
func (r *EncryptionJobRepository) scanJob(row pgx.Row) (*domain.EncryptionJob, error) {
	var job domain.EncryptionJob
	var status string
	var errorMsg *string
	var startedAt, completedAt *time.Time

	err := row.Scan(
		&job.ID,
		&job.MaterialID,
		&job.UserID,
		&job.DeviceID,
		&job.LicenseID,
		&job.Priority,
		&status,
		&errorMsg,
		&job.RetryCount,
		&job.CreatedAt,
		&startedAt,
		&completedAt,
	)
	if err != nil {
		return nil, err
	}

	job.Status = domain.JobStatus(status)
	job.Error = errorMsg
	job.StartedAt = startedAt
	job.CompletedAt = completedAt

	return &job, nil
}

// scanJobs scans multiple job rows.
func (r *EncryptionJobRepository) scanJobs(rows pgx.Rows) ([]*domain.EncryptionJob, error) {
	var jobs []*domain.EncryptionJob

	for rows.Next() {
		var job domain.EncryptionJob
		var status string
		var errorMsg *string
		var startedAt, completedAt *time.Time

		err := rows.Scan(
			&job.ID,
			&job.MaterialID,
			&job.UserID,
			&job.DeviceID,
			&job.LicenseID,
			&job.Priority,
			&status,
			&errorMsg,
			&job.RetryCount,
			&job.CreatedAt,
			&startedAt,
			&completedAt,
		)
		if err != nil {
			return nil, err
		}

		job.Status = domain.JobStatus(status)
		job.Error = errorMsg
		job.StartedAt = startedAt
		job.CompletedAt = completedAt

		jobs = append(jobs, &job)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return jobs, nil
}

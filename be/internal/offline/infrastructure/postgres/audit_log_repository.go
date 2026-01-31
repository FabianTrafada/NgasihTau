// Package postgres provides PostgreSQL implementations of the offline domain repositories.
package postgres

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"ngasihtau/internal/common/errors"
	"ngasihtau/internal/offline/domain"
)

// AuditLogRepository implements domain.AuditLogRepository using PostgreSQL.
type AuditLogRepository struct {
	db *pgxpool.Pool
}

// NewAuditLogRepository creates a new AuditLogRepository.
func NewAuditLogRepository(db *pgxpool.Pool) *AuditLogRepository {
	return &AuditLogRepository{db: db}
}

// Create creates a new audit log entry.
// Implements Requirement 7.5: Audit logging repository.
func (r *AuditLogRepository) Create(ctx context.Context, log *domain.AuditLog) error {
	query := `
		INSERT INTO offline_audit_logs (id, user_id, device_id, action, resource, resource_id, ip_address, user_agent, success, error_code, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`
	_, err := r.db.Exec(ctx, query,
		log.ID,
		log.UserID,
		log.DeviceID,
		log.Action,
		log.Resource,
		log.ResourceID,
		log.IPAddress,
		log.UserAgent,
		log.Success,
		log.ErrorCode,
		log.CreatedAt,
	)
	if err != nil {
		return errors.Internal("failed to create audit log", err)
	}
	return nil
}

// FindByUserID finds audit logs for a user with pagination.
func (r *AuditLogRepository) FindByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*domain.AuditLog, int, error) {
	// Get total count
	countQuery := `SELECT COUNT(*) FROM offline_audit_logs WHERE user_id = $1`
	var total int
	if err := r.db.QueryRow(ctx, countQuery, userID).Scan(&total); err != nil {
		return nil, 0, errors.Internal("failed to count audit logs", err)
	}

	// Get logs
	query := `
		SELECT id, user_id, device_id, action, resource, resource_id, ip_address, user_agent, success, error_code, created_at
		FROM offline_audit_logs
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.db.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, 0, errors.Internal("failed to query audit logs", err)
	}
	defer rows.Close()

	logs, err := r.scanAuditLogs(rows)
	if err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

// FindByDeviceID finds audit logs for a device with pagination.
func (r *AuditLogRepository) FindByDeviceID(ctx context.Context, deviceID uuid.UUID, limit, offset int) ([]*domain.AuditLog, int, error) {
	// Get total count
	countQuery := `SELECT COUNT(*) FROM offline_audit_logs WHERE device_id = $1`
	var total int
	if err := r.db.QueryRow(ctx, countQuery, deviceID).Scan(&total); err != nil {
		return nil, 0, errors.Internal("failed to count audit logs", err)
	}

	// Get logs
	query := `
		SELECT id, user_id, device_id, action, resource, resource_id, ip_address, user_agent, success, error_code, created_at
		FROM offline_audit_logs
		WHERE device_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.db.Query(ctx, query, deviceID, limit, offset)
	if err != nil {
		return nil, 0, errors.Internal("failed to query audit logs", err)
	}
	defer rows.Close()

	logs, err := r.scanAuditLogs(rows)
	if err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

// FindByAction finds audit logs by action type with pagination.
func (r *AuditLogRepository) FindByAction(ctx context.Context, action string, limit, offset int) ([]*domain.AuditLog, int, error) {
	// Get total count
	countQuery := `SELECT COUNT(*) FROM offline_audit_logs WHERE action = $1`
	var total int
	if err := r.db.QueryRow(ctx, countQuery, action).Scan(&total); err != nil {
		return nil, 0, errors.Internal("failed to count audit logs", err)
	}

	// Get logs
	query := `
		SELECT id, user_id, device_id, action, resource, resource_id, ip_address, user_agent, success, error_code, created_at
		FROM offline_audit_logs
		WHERE action = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.db.Query(ctx, query, action, limit, offset)
	if err != nil {
		return nil, 0, errors.Internal("failed to query audit logs", err)
	}
	defer rows.Close()

	logs, err := r.scanAuditLogs(rows)
	if err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

// CountFailedByDevice counts failed operations for a device since a given time.
// Used for rate limiting and security monitoring.
func (r *AuditLogRepository) CountFailedByDevice(ctx context.Context, deviceID uuid.UUID, since time.Time) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM offline_audit_logs
		WHERE device_id = $1 AND success = false AND created_at >= $2
	`
	var count int
	if err := r.db.QueryRow(ctx, query, deviceID, since).Scan(&count); err != nil {
		return 0, errors.Internal("failed to count failed operations", err)
	}
	return count, nil
}

// CountFailedByDeviceAndAction counts failed operations for a device and specific action since a given time.
func (r *AuditLogRepository) CountFailedByDeviceAndAction(ctx context.Context, deviceID uuid.UUID, action string, since time.Time) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM offline_audit_logs
		WHERE device_id = $1 AND action = $2 AND success = false AND created_at >= $3
	`
	var count int
	if err := r.db.QueryRow(ctx, query, deviceID, action, since).Scan(&count); err != nil {
		return 0, errors.Internal("failed to count failed operations", err)
	}
	return count, nil
}

// scanAuditLogs scans multiple audit log rows.
func (r *AuditLogRepository) scanAuditLogs(rows pgx.Rows) ([]*domain.AuditLog, error) {
	var logs []*domain.AuditLog
	for rows.Next() {
		var log domain.AuditLog
		err := rows.Scan(
			&log.ID,
			&log.UserID,
			&log.DeviceID,
			&log.Action,
			&log.Resource,
			&log.ResourceID,
			&log.IPAddress,
			&log.UserAgent,
			&log.Success,
			&log.ErrorCode,
			&log.CreatedAt,
		)
		if err != nil {
			return nil, errors.Internal("failed to scan audit log", err)
		}
		logs = append(logs, &log)
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Internal("error iterating audit logs", err)
	}
	return logs, nil
}

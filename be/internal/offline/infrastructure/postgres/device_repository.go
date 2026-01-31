// Package postgres provides PostgreSQL implementations of the offline domain repositories.
package postgres

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"ngasihtau/internal/common/errors"
	"ngasihtau/internal/offline/domain"
)

// DeviceRepository implements domain.DeviceRepository using PostgreSQL.
type DeviceRepository struct {
	db *pgxpool.Pool
}

// NewDeviceRepository creates a new DeviceRepository.
func NewDeviceRepository(db *pgxpool.Pool) *DeviceRepository {
	return &DeviceRepository{db: db}
}

// Create creates a new device.
// Implements Requirement 5.1: Store device fingerprint with user association.
func (r *DeviceRepository) Create(ctx context.Context, device *domain.Device) error {
	query := `
		INSERT INTO offline_devices (id, user_id, fingerprint, name, platform, last_used_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := r.db.Exec(ctx, query,
		device.ID,
		device.UserID,
		device.Fingerprint,
		device.Name,
		device.Platform,
		device.LastUsedAt,
		device.CreatedAt,
	)
	if err != nil {
		// Check for unique constraint violation (user already has this fingerprint)
		if isUniqueViolation(err) {
			return errors.Conflict("device", device.Fingerprint)
		}
		return errors.Internal("failed to create device", err)
	}
	return nil
}

// FindById finds a device by ID.
func (r *DeviceRepository) FindById(ctx context.Context, id uuid.UUID) (*domain.Device, error) {
	query := `
		SELECT id, user_id, fingerprint, name, platform, last_used_at, created_at, revoked_at
		FROM offline_devices
		WHERE id = $1
	`
	return r.scanDevice(r.db.QueryRow(ctx, query, id))
}


// FindByUserID finds all active devices for a user.
// Implements Requirement 5: Device Management - list devices.
func (r *DeviceRepository) FindByUserID(ctx context.Context, userID uuid.UUID) ([]*domain.Device, error) {
	query := `
		SELECT id, user_id, fingerprint, name, platform, last_used_at, created_at, revoked_at
		FROM offline_devices
		WHERE user_id = $1 AND revoked_at IS NULL
		ORDER BY last_used_at DESC
	`
	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, errors.Internal("failed to query devices", err)
	}
	defer rows.Close()

	return r.scanDevices(rows)
}

// FindByFingerprint finds a device by user ID and fingerprint.
// Implements Requirement 5.6: Validate device fingerprints.
func (r *DeviceRepository) FindByFingerprint(ctx context.Context, userID uuid.UUID, fingerprint string) (*domain.Device, error) {
	query := `
		SELECT id, user_id, fingerprint, name, platform, last_used_at, created_at, revoked_at
		FROM offline_devices
		WHERE user_id = $1 AND fingerprint = $2 AND revoked_at IS NULL
	`
	return r.scanDevice(r.db.QueryRow(ctx, query, userID, fingerprint))
}

// CountActiveByUserID counts active (non-revoked) devices for a user.
// Implements Requirement 5.2: Enforce maximum of 5 registered devices per user.
func (r *DeviceRepository) CountActiveByUserID(ctx context.Context, userID uuid.UUID) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM offline_devices
		WHERE user_id = $1 AND revoked_at IS NULL
	`
	var count int
	if err := r.db.QueryRow(ctx, query, userID).Scan(&count); err != nil {
		return 0, errors.Internal("failed to count devices", err)
	}
	return count, nil
}

// UpdateLastUsed updates the last_used_at timestamp for a device.
// Implements Requirement 5.8: Track last_used_at timestamp for each device.
func (r *DeviceRepository) UpdateLastUsed(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE offline_devices
		SET last_used_at = $2
		WHERE id = $1 AND revoked_at IS NULL
	`
	result, err := r.db.Exec(ctx, query, id, time.Now())
	if err != nil {
		return errors.Internal("failed to update device last used", err)
	}
	if result.RowsAffected() == 0 {
		return errors.NotFound("device", id.String())
	}
	return nil
}

// Revoke marks a device as revoked.
// Implements Requirement 5.4: Allow users to deregister devices.
func (r *DeviceRepository) Revoke(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE offline_devices
		SET revoked_at = $2
		WHERE id = $1 AND revoked_at IS NULL
	`
	result, err := r.db.Exec(ctx, query, id, time.Now())
	if err != nil {
		return errors.Internal("failed to revoke device", err)
	}
	if result.RowsAffected() == 0 {
		return errors.NotFound("device", id.String())
	}
	return nil
}

// scanDevice scans a single device row.
func (r *DeviceRepository) scanDevice(row pgx.Row) (*domain.Device, error) {
	var device domain.Device
	err := row.Scan(
		&device.ID,
		&device.UserID,
		&device.Fingerprint,
		&device.Name,
		&device.Platform,
		&device.LastUsedAt,
		&device.CreatedAt,
		&device.RevokedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, errors.NotFound("device", "")
		}
		return nil, errors.Internal("failed to scan device", err)
	}
	return &device, nil
}

// scanDevices scans multiple device rows.
func (r *DeviceRepository) scanDevices(rows pgx.Rows) ([]*domain.Device, error) {
	var devices []*domain.Device
	for rows.Next() {
		var device domain.Device
		err := rows.Scan(
			&device.ID,
			&device.UserID,
			&device.Fingerprint,
			&device.Name,
			&device.Platform,
			&device.LastUsedAt,
			&device.CreatedAt,
			&device.RevokedAt,
		)
		if err != nil {
			return nil, errors.Internal("failed to scan device", err)
		}
		devices = append(devices, &device)
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Internal("error iterating devices", err)
	}
	return devices, nil
}

// isUniqueViolation checks if the error is a unique constraint violation.
func isUniqueViolation(err error) bool {
	// PostgreSQL error code for unique_violation is 23505
	if pgErr, ok := err.(*pgconn.PgError); ok {
		return pgErr.Code == "23505"
	}
	return false
}

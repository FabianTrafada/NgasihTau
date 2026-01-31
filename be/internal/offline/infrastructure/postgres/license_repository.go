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

// LicenseRepository implements domain.LicenseRepository using PostgreSQL.
type LicenseRepository struct {
	db *pgxpool.Pool
}

// NewLicenseRepository creates a new LicenseRepository.
func NewLicenseRepository(db *pgxpool.Pool) *LicenseRepository {
	return &LicenseRepository{db: db}
}

// Create creates a new license.
// Implements Requirement 3.1: Issue licenses for offline access.
func (r *LicenseRepository) Create(ctx context.Context, license *domain.License) error {
	query := `
		INSERT INTO offline_licenses (
			id, user_id, material_id, device_id, status, expires_at,
			offline_grace_period, last_validated_at, nonce, created_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`
	_, err := r.db.Exec(ctx, query,
		license.ID,
		license.UserID,
		license.MaterialID,
		license.DeviceID,
		license.Status,
		license.ExpiresAt,
		license.OfflineGracePeriod,
		license.LastValidatedAt,
		license.Nonce,
		license.CreatedAt,
	)
	if err != nil {
		if isUniqueViolation(err) {
			return errors.Conflict("license", license.ID.String())
		}
		return errors.Internal("failed to create license", err)
	}
	return nil
}

// FindByID finds a license by ID.
func (r *LicenseRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.License, error) {
	query := `
		SELECT id, user_id, material_id, device_id, status, expires_at,
			   offline_grace_period, last_validated_at, nonce, created_at, revoked_at
		FROM offline_licenses
		WHERE id = $1
	`
	return r.scanLicense(r.db.QueryRow(ctx, query, id))
}


// FindByUserAndMaterial finds a license by user, material, and device.
// Implements Requirement 3: License Management - find existing license.
func (r *LicenseRepository) FindByUserAndMaterial(ctx context.Context, userID, materialID, deviceID uuid.UUID) (*domain.License, error) {
	query := `
		SELECT id, user_id, material_id, device_id, status, expires_at,
			   offline_grace_period, last_validated_at, nonce, created_at, revoked_at
		FROM offline_licenses
		WHERE user_id = $1 AND material_id = $2 AND device_id = $3 AND revoked_at IS NULL
	`
	return r.scanLicense(r.db.QueryRow(ctx, query, userID, materialID, deviceID))
}

// FindActiveByDeviceID finds all active licenses for a device.
// Implements Requirement 3.5: Revoke licenses when device is deregistered.
func (r *LicenseRepository) FindActiveByDeviceID(ctx context.Context, deviceID uuid.UUID) ([]*domain.License, error) {
	query := `
		SELECT id, user_id, material_id, device_id, status, expires_at,
			   offline_grace_period, last_validated_at, nonce, created_at, revoked_at
		FROM offline_licenses
		WHERE device_id = $1 AND status = 'active' AND revoked_at IS NULL
		ORDER BY created_at DESC
	`
	rows, err := r.db.Query(ctx, query, deviceID)
	if err != nil {
		return nil, errors.Internal("failed to query licenses", err)
	}
	defer rows.Close()

	return r.scanLicenses(rows)
}

// FindActiveByUserID finds all active licenses for a user.
func (r *LicenseRepository) FindActiveByUserID(ctx context.Context, userID uuid.UUID) ([]*domain.License, error) {
	query := `
		SELECT id, user_id, material_id, device_id, status, expires_at,
			   offline_grace_period, last_validated_at, nonce, created_at, revoked_at
		FROM offline_licenses
		WHERE user_id = $1 AND status = 'active' AND revoked_at IS NULL
		ORDER BY created_at DESC
	`
	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, errors.Internal("failed to query licenses", err)
	}
	defer rows.Close()

	return r.scanLicenses(rows)
}

// UpdateValidation updates the last_validated_at timestamp and nonce.
// Implements Requirement 3.4: Update last_validated_at on validation.
// Implements Property 13: License Validation Timestamp Update.
func (r *LicenseRepository) UpdateValidation(ctx context.Context, id uuid.UUID, nonce string) error {
	query := `
		UPDATE offline_licenses
		SET last_validated_at = $2, nonce = $3
		WHERE id = $1 AND status = 'active' AND revoked_at IS NULL
	`
	result, err := r.db.Exec(ctx, query, id, time.Now(), nonce)
	if err != nil {
		return errors.Internal("failed to update license validation", err)
	}
	if result.RowsAffected() == 0 {
		return errors.NotFound("license", id.String())
	}
	return nil
}

// UpdateExpiration updates the expires_at timestamp.
// Implements Requirement 3.6: Allow license renewal.
// Implements Property 15: License Renewal Extension.
func (r *LicenseRepository) UpdateExpiration(ctx context.Context, id uuid.UUID, expiresAt time.Time) error {
	query := `
		UPDATE offline_licenses
		SET expires_at = $2
		WHERE id = $1 AND status = 'active' AND revoked_at IS NULL
	`
	result, err := r.db.Exec(ctx, query, id, expiresAt)
	if err != nil {
		return errors.Internal("failed to update license expiration", err)
	}
	if result.RowsAffected() == 0 {
		return errors.NotFound("license", id.String())
	}
	return nil
}

// Revoke marks a license as revoked.
// Implements Requirement 3.5: Revoke licenses.
func (r *LicenseRepository) Revoke(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE offline_licenses
		SET status = 'revoked', revoked_at = $2
		WHERE id = $1 AND revoked_at IS NULL
	`
	result, err := r.db.Exec(ctx, query, id, time.Now())
	if err != nil {
		return errors.Internal("failed to revoke license", err)
	}
	if result.RowsAffected() == 0 {
		return errors.NotFound("license", id.String())
	}
	return nil
}

// RevokeByDeviceID revokes all licenses for a device.
// Implements Requirement 5.5: Revoke all licenses bound to deregistered device.
// Implements Property 14: License Cascading Revocation.
func (r *LicenseRepository) RevokeByDeviceID(ctx context.Context, deviceID uuid.UUID) error {
	query := `
		UPDATE offline_licenses
		SET status = 'revoked', revoked_at = $2
		WHERE device_id = $1 AND revoked_at IS NULL
	`
	_, err := r.db.Exec(ctx, query, deviceID, time.Now())
	if err != nil {
		return errors.Internal("failed to revoke licenses by device", err)
	}
	return nil
}

// RevokeByMaterialID revokes all licenses for a material.
// Implements Property 14: License Cascading Revocation.
func (r *LicenseRepository) RevokeByMaterialID(ctx context.Context, materialID uuid.UUID) error {
	query := `
		UPDATE offline_licenses
		SET status = 'revoked', revoked_at = $2
		WHERE material_id = $1 AND revoked_at IS NULL
	`
	_, err := r.db.Exec(ctx, query, materialID, time.Now())
	if err != nil {
		return errors.Internal("failed to revoke licenses by material", err)
	}
	return nil
}

// RevokeByUserAndMaterial revokes all licenses for a user and material combination.
func (r *LicenseRepository) RevokeByUserAndMaterial(ctx context.Context, userID, materialID uuid.UUID) error {
	query := `
		UPDATE offline_licenses
		SET status = 'revoked', revoked_at = $2
		WHERE user_id = $1 AND material_id = $3 AND revoked_at IS NULL
	`
	_, err := r.db.Exec(ctx, query, userID, time.Now(), materialID)
	if err != nil {
		return errors.Internal("failed to revoke licenses by user and material", err)
	}
	return nil
}

// scanLicense scans a single license row.
func (r *LicenseRepository) scanLicense(row pgx.Row) (*domain.License, error) {
	var license domain.License
	var gracePeriod string
	err := row.Scan(
		&license.ID,
		&license.UserID,
		&license.MaterialID,
		&license.DeviceID,
		&license.Status,
		&license.ExpiresAt,
		&gracePeriod,
		&license.LastValidatedAt,
		&license.Nonce,
		&license.CreatedAt,
		&license.RevokedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, errors.NotFound("license", "")
		}
		return nil, errors.Internal("failed to scan license", err)
	}
	// Parse interval string to duration
	license.OfflineGracePeriod = parseInterval(gracePeriod)
	return &license, nil
}

// scanLicenses scans multiple license rows.
func (r *LicenseRepository) scanLicenses(rows pgx.Rows) ([]*domain.License, error) {
	var licenses []*domain.License
	for rows.Next() {
		var license domain.License
		var gracePeriod string
		err := rows.Scan(
			&license.ID,
			&license.UserID,
			&license.MaterialID,
			&license.DeviceID,
			&license.Status,
			&license.ExpiresAt,
			&gracePeriod,
			&license.LastValidatedAt,
			&license.Nonce,
			&license.CreatedAt,
			&license.RevokedAt,
		)
		if err != nil {
			return nil, errors.Internal("failed to scan license", err)
		}
		license.OfflineGracePeriod = parseInterval(gracePeriod)
		licenses = append(licenses, &license)
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Internal("error iterating licenses", err)
	}
	return licenses, nil
}

// parseInterval parses a PostgreSQL interval string to time.Duration.
// PostgreSQL returns intervals like "72:00:00" for 72 hours.
func parseInterval(interval string) time.Duration {
	// Try parsing as duration directly (e.g., "72h")
	if d, err := time.ParseDuration(interval); err == nil {
		return d
	}
	// Default to 72 hours if parsing fails
	return domain.DefaultOfflineGracePeriod
}

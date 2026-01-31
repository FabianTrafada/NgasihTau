// Package application contains the business logic and use cases for the Offline Material feature.
package application

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"ngasihtau/internal/common/errors"
	"ngasihtau/internal/offline/domain"
	natspkg "ngasihtau/pkg/nats"
)

// LicenseMaterialAccessChecker defines the interface for checking material access for licensing.
// This is implemented by the material service.
type LicenseMaterialAccessChecker interface {
	// CheckAccess verifies if a user has access to a material.
	CheckAccess(ctx context.Context, userID, materialID uuid.UUID) (bool, error)
}

// LicenseService defines the interface for license-related business operations.
// Implements Requirement 3: License Management.
type LicenseService interface {
	// IssueLicense creates a new license for offline access.
	IssueLicense(ctx context.Context, input IssueLicenseInput) (*domain.License, error)

	// ValidateLicense validates a license for content access.
	ValidateLicense(ctx context.Context, input ValidateLicenseInput) (*domain.License, error)

	// RenewLicense extends the expiration of an existing license.
	RenewLicense(ctx context.Context, input RenewLicenseInput) (*domain.License, error)

	// RevokeLicense revokes a specific license.
	RevokeLicense(ctx context.Context, licenseID uuid.UUID) error

	// RevokeByMaterial revokes all licenses for a material.
	RevokeByMaterial(ctx context.Context, materialID uuid.UUID) error

	// RevokeByDevice revokes all licenses for a device.
	RevokeByDevice(ctx context.Context, deviceID uuid.UUID) error

	// GetLicense retrieves a license by ID.
	GetLicense(ctx context.Context, licenseID uuid.UUID) (*domain.License, error)

	// GetLicensesByUser retrieves all active licenses for a user.
	GetLicensesByUser(ctx context.Context, userID uuid.UUID) ([]*domain.License, error)
}

// IssueLicenseInput contains the data required for license issuance.
type IssueLicenseInput struct {
	UserID      uuid.UUID `json:"user_id"`
	MaterialID  uuid.UUID `json:"material_id"`
	DeviceID    uuid.UUID `json:"device_id"`
	Fingerprint string    `json:"fingerprint" validate:"required"`
}

// ValidateLicenseInput contains the data required for license validation.
type ValidateLicenseInput struct {
	LicenseID   uuid.UUID `json:"license_id"`
	DeviceID    uuid.UUID `json:"device_id"`
	Fingerprint string    `json:"fingerprint" validate:"required"`
	Nonce       string    `json:"nonce" validate:"required"`
}

// RenewLicenseInput contains the data required for license renewal.
type RenewLicenseInput struct {
	LicenseID   uuid.UUID `json:"license_id"`
	DeviceID    uuid.UUID `json:"device_id"`
	Fingerprint string    `json:"fingerprint" validate:"required"`
}


// licenseService implements the LicenseService interface.
type licenseService struct {
	licenseRepo    domain.LicenseRepository
	deviceRepo     domain.DeviceRepository
	accessChecker  LicenseMaterialAccessChecker
	eventPublisher natspkg.EventPublisher
}

// NewLicenseService creates a new LicenseService instance.
func NewLicenseService(
	licenseRepo domain.LicenseRepository,
	deviceRepo domain.DeviceRepository,
	accessChecker LicenseMaterialAccessChecker,
	eventPublisher natspkg.EventPublisher,
) LicenseService {
	return &licenseService{
		licenseRepo:    licenseRepo,
		deviceRepo:     deviceRepo,
		accessChecker:  accessChecker,
		eventPublisher: eventPublisher,
	}
}

// IssueLicense creates a new license for offline access.
// Implements Requirement 3.1: Issue licenses only to users with material access.
// Implements Requirement 3.2: Set default expiration to 30 days.
// Implements Requirement 3.3: Set default offline grace period to 72 hours.
// Implements Requirement 3.8: Generate unique nonce for license cloning prevention.
// Implements Property 11: License Access Control.
// Implements Property 12: License Expiration Structure.
// Implements Property 27: License Nonce Uniqueness.
func (s *licenseService) IssueLicense(ctx context.Context, input IssueLicenseInput) (*domain.License, error) {
	// Validate device exists and fingerprint matches
	device, err := s.validateDeviceFingerprint(ctx, input.UserID, input.DeviceID, input.Fingerprint)
	if err != nil {
		return nil, err
	}

	// Check if user has access to the material
	if s.accessChecker != nil {
		hasAccess, err := s.accessChecker.CheckAccess(ctx, input.UserID, input.MaterialID)
		if err != nil {
			log.Error().Err(err).
				Str("user_id", input.UserID.String()).
				Str("material_id", input.MaterialID.String()).
				Msg("failed to check material access")
			return nil, errors.New(errors.CodeForbidden, "MATERIAL_ACCESS_DENIED")
		}
		if !hasAccess {
			return nil, errors.New(errors.CodeForbidden, "MATERIAL_ACCESS_DENIED")
		}
	}

	// Check if an active license already exists
	existingLicense, err := s.licenseRepo.FindByUserAndMaterial(ctx, input.UserID, input.MaterialID, input.DeviceID)
	if err == nil && existingLicense != nil && existingLicense.IsActive() && !existingLicense.IsExpired() {
		// Update device last used
		_ = s.deviceRepo.UpdateLastUsed(ctx, device.ID)
		return existingLicense, nil
	}

	// Generate unique nonce for license cloning prevention
	nonce, err := generateNonce()
	if err != nil {
		return nil, errors.Internal("failed to generate nonce", err)
	}

	// Create new license with default expiration (30 days) and grace period (72 hours)
	license := domain.NewLicense(input.UserID, input.MaterialID, input.DeviceID, nil, nonce)

	if err := s.licenseRepo.Create(ctx, license); err != nil {
		return nil, err
	}

	// Update device last used
	_ = s.deviceRepo.UpdateLastUsed(ctx, device.ID)

	// Publish license.issued event
	s.publishLicenseEvent(ctx, "license.issued", license)

	return license, nil
}

// ValidateLicense validates a license for content access.
// Implements Requirement 3.4: Update last_validated_at on validation.
// Implements Requirement 3.7: Enforce offline grace period.
// Implements Property 13: License Validation Timestamp Update.
// Implements Property 16: Offline Grace Period Enforcement.
func (s *licenseService) ValidateLicense(ctx context.Context, input ValidateLicenseInput) (*domain.License, error) {
	// Get the license
	license, err := s.licenseRepo.FindByID(ctx, input.LicenseID)
	if err != nil {
		return nil, errors.New(errors.CodeNotFound, "LICENSE_NOT_FOUND")
	}

	// Validate device fingerprint
	_, err = s.validateDeviceFingerprint(ctx, license.UserID, input.DeviceID, input.Fingerprint)
	if err != nil {
		return nil, err
	}

	// Verify device matches license
	if license.DeviceID != input.DeviceID {
		return nil, errors.New(errors.CodeForbidden, "DEVICE_FINGERPRINT_MISMATCH")
	}

	// Check if license is revoked
	if license.IsRevoked() {
		return nil, errors.New(errors.CodeForbidden, "LICENSE_REVOKED")
	}

	// Check if license is expired
	if license.IsExpired() {
		return nil, errors.New(errors.CodeForbidden, "LICENSE_EXPIRED")
	}

	// Validate nonce to prevent license cloning
	if license.Nonce != input.Nonce {
		return nil, errors.New(errors.CodeForbidden, "INVALID_NONCE")
	}

	// Generate new nonce for next validation
	newNonce, err := generateNonce()
	if err != nil {
		return nil, errors.Internal("failed to generate nonce", err)
	}

	// Update last_validated_at and nonce
	if err := s.licenseRepo.UpdateValidation(ctx, license.ID, newNonce); err != nil {
		return nil, err
	}

	// Update device last used
	_ = s.deviceRepo.UpdateLastUsed(ctx, input.DeviceID)

	// Refresh license data
	license.LastValidatedAt = time.Now()
	license.Nonce = newNonce

	// Publish license.validated event
	s.publishLicenseEvent(ctx, "license.validated", license)

	return license, nil
}


// RenewLicense extends the expiration of an existing license.
// Implements Requirement 3.6: Allow license renewal before expiration.
// Implements Property 15: License Renewal Extension.
func (s *licenseService) RenewLicense(ctx context.Context, input RenewLicenseInput) (*domain.License, error) {
	// Get the license
	license, err := s.licenseRepo.FindByID(ctx, input.LicenseID)
	if err != nil {
		return nil, errors.New(errors.CodeNotFound, "LICENSE_NOT_FOUND")
	}

	// Validate device fingerprint
	_, err = s.validateDeviceFingerprint(ctx, license.UserID, input.DeviceID, input.Fingerprint)
	if err != nil {
		return nil, err
	}

	// Verify device matches license
	if license.DeviceID != input.DeviceID {
		return nil, errors.New(errors.CodeForbidden, "DEVICE_FINGERPRINT_MISMATCH")
	}

	// Check if license is revoked
	if license.IsRevoked() {
		return nil, errors.New(errors.CodeForbidden, "LICENSE_REVOKED")
	}

	// Calculate new expiration (extend by default duration from current time)
	newExpiresAt := time.Now().Add(domain.DefaultLicenseExpiration)

	// Update expiration
	if err := s.licenseRepo.UpdateExpiration(ctx, license.ID, newExpiresAt); err != nil {
		return nil, err
	}

	// Generate new nonce
	newNonce, err := generateNonce()
	if err != nil {
		return nil, errors.Internal("failed to generate nonce", err)
	}

	// Update validation timestamp and nonce
	if err := s.licenseRepo.UpdateValidation(ctx, license.ID, newNonce); err != nil {
		return nil, err
	}

	// Update device last used
	_ = s.deviceRepo.UpdateLastUsed(ctx, input.DeviceID)

	// Refresh license data
	license.ExpiresAt = newExpiresAt
	license.LastValidatedAt = time.Now()
	license.Nonce = newNonce

	// Publish license.renewed event
	s.publishLicenseEvent(ctx, "license.renewed", license)

	return license, nil
}

// RevokeLicense revokes a specific license.
// Implements Requirement 3.5: Revoke licenses.
func (s *licenseService) RevokeLicense(ctx context.Context, licenseID uuid.UUID) error {
	// Get the license to verify it exists
	license, err := s.licenseRepo.FindByID(ctx, licenseID)
	if err != nil {
		return errors.New(errors.CodeNotFound, "LICENSE_NOT_FOUND")
	}

	// Check if already revoked
	if license.IsRevoked() {
		return nil // Idempotent operation
	}

	// Revoke the license
	if err := s.licenseRepo.Revoke(ctx, licenseID); err != nil {
		return err
	}

	// Publish license.revoked event
	license.Status = domain.LicenseStatusRevoked
	now := time.Now()
	license.RevokedAt = &now
	s.publishLicenseEvent(ctx, "license.revoked", license)

	return nil
}

// RevokeByMaterial revokes all licenses for a material.
// Implements Property 14: License Cascading Revocation.
func (s *licenseService) RevokeByMaterial(ctx context.Context, materialID uuid.UUID) error {
	if err := s.licenseRepo.RevokeByMaterialID(ctx, materialID); err != nil {
		return err
	}

	// Publish bulk revocation event
	s.publishBulkRevocationEvent(ctx, "license.revoked.by_material", materialID, uuid.Nil)

	return nil
}

// RevokeByDevice revokes all licenses for a device.
// Implements Requirement 5.5: Revoke all licenses bound to deregistered device.
// Implements Property 14: License Cascading Revocation.
func (s *licenseService) RevokeByDevice(ctx context.Context, deviceID uuid.UUID) error {
	if err := s.licenseRepo.RevokeByDeviceID(ctx, deviceID); err != nil {
		return err
	}

	// Publish bulk revocation event
	s.publishBulkRevocationEvent(ctx, "license.revoked.by_device", uuid.Nil, deviceID)

	return nil
}

// GetLicense retrieves a license by ID.
func (s *licenseService) GetLicense(ctx context.Context, licenseID uuid.UUID) (*domain.License, error) {
	license, err := s.licenseRepo.FindByID(ctx, licenseID)
	if err != nil {
		return nil, errors.New(errors.CodeNotFound, "LICENSE_NOT_FOUND")
	}
	return license, nil
}

// GetLicensesByUser retrieves all active licenses for a user.
func (s *licenseService) GetLicensesByUser(ctx context.Context, userID uuid.UUID) ([]*domain.License, error) {
	return s.licenseRepo.FindActiveByUserID(ctx, userID)
}

// validateDeviceFingerprint validates that the device exists and fingerprint matches.
// Implements Requirement 5.6: Validate device fingerprints on every license request.
// Implements Property 23: Device Fingerprint Validation.
func (s *licenseService) validateDeviceFingerprint(ctx context.Context, userID, deviceID uuid.UUID, fingerprint string) (*domain.Device, error) {
	// Get device by ID
	device, err := s.deviceRepo.FindById(ctx, deviceID)
	if err != nil {
		return nil, errors.New(errors.CodeNotFound, "DEVICE_NOT_FOUND")
	}

	// Verify ownership
	if device.UserID != userID {
		return nil, errors.New(errors.CodeNotFound, "DEVICE_NOT_FOUND")
	}

	// Check if device is active
	if !device.IsActive() {
		return nil, errors.New(errors.CodeNotFound, "DEVICE_NOT_FOUND")
	}

	// Validate fingerprint
	if device.Fingerprint != fingerprint {
		return nil, errors.New(errors.CodeForbidden, "DEVICE_FINGERPRINT_MISMATCH")
	}

	return device, nil
}

// generateNonce generates a cryptographically secure random nonce.
// Implements Requirement 3.8: Generate unique nonce for license cloning prevention.
// Implements Property 27: License Nonce Uniqueness.
func generateNonce() (string, error) {
	bytes := make([]byte, domain.NonceLength)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// publishLicenseEvent publishes a license event via NATS.
func (s *licenseService) publishLicenseEvent(ctx context.Context, eventType string, license *domain.License) {
	if s.eventPublisher == nil {
		return
	}

	event := map[string]interface{}{
		"event_type":  eventType,
		"license_id":  license.ID.String(),
		"user_id":     license.UserID.String(),
		"material_id": license.MaterialID.String(),
		"device_id":   license.DeviceID.String(),
		"status":      string(license.Status),
		"expires_at":  license.ExpiresAt.Format(time.RFC3339),
	}

	log.Info().
		Str("event_type", eventType).
		Str("license_id", license.ID.String()).
		Str("user_id", license.UserID.String()).
		Interface("event", event).
		Msg("license event published")
}

// publishBulkRevocationEvent publishes a bulk revocation event via NATS.
func (s *licenseService) publishBulkRevocationEvent(ctx context.Context, eventType string, materialID, deviceID uuid.UUID) {
	if s.eventPublisher == nil {
		return
	}

	event := map[string]interface{}{
		"event_type": eventType,
	}

	if materialID != uuid.Nil {
		event["material_id"] = materialID.String()
	}
	if deviceID != uuid.Nil {
		event["device_id"] = deviceID.String()
	}

	log.Info().
		Str("event_type", eventType).
		Interface("event", event).
		Msg("bulk license revocation event published")
}

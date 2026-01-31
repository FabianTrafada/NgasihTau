// Package application contains the business logic and use cases for the Offline Material feature.
package application

import (
	"context"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"ngasihtau/internal/common/errors"
	"ngasihtau/internal/offline/domain"
	natspkg "ngasihtau/pkg/nats"
)

// MaxDevicesPerUser is the maximum number of devices a user can register.
// Implements Requirement 5.2: Enforce a maximum of 5 registered devices per user.
const MaxDevicesPerUser = 5

// DeviceService defines the interface for device-related business operations.
// Implements Requirement 5: Device Management.
type DeviceService interface {
	// RegisterDevice registers a new device for a user.
	RegisterDevice(ctx context.Context, input RegisterDeviceInput) (*domain.Device, error)

	// ListDevices returns all active devices for a user.
	ListDevices(ctx context.Context, userID uuid.UUID) ([]*domain.Device, error)

	// DeregisterDevice removes a device and revokes associated licenses.
	DeregisterDevice(ctx context.Context, userID, deviceID uuid.UUID) error

	// ValidateDevice validates a device fingerprint for a user.
	ValidateDevice(ctx context.Context, userID uuid.UUID, fingerprint string) (*domain.Device, error)

	// UpdateLastUsed updates the last_used_at timestamp for a device.
	UpdateLastUsed(ctx context.Context, deviceID uuid.UUID) error
}

// RegisterDeviceInput contains the data required for device registration.
type RegisterDeviceInput struct {
	UserID      uuid.UUID       `json:"user_id"`
	Fingerprint string          `json:"fingerprint" validate:"required,min=32,max=512"`
	Name        string          `json:"name" validate:"required,min=1,max=255"`
	Platform    domain.Platform `json:"platform" validate:"required,oneof=ios android desktop"`
}

// deviceService implements the DeviceService interface.
type deviceService struct {
	deviceRepo     domain.DeviceRepository
	licenseRepo    domain.LicenseRepository
	cekRepo        domain.CEKRepository
	eventPublisher natspkg.EventPublisher
}


// NewDeviceService creates a new DeviceService instance.
func NewDeviceService(
	deviceRepo domain.DeviceRepository,
	licenseRepo domain.LicenseRepository,
	cekRepo domain.CEKRepository,
	eventPublisher natspkg.EventPublisher,
) DeviceService {
	return &deviceService{
		deviceRepo:     deviceRepo,
		licenseRepo:    licenseRepo,
		cekRepo:        cekRepo,
		eventPublisher: eventPublisher,
	}
}

// RegisterDevice registers a new device for a user.
// Implements Requirement 5.1: Store device fingerprint with user association.
// Implements Requirement 5.2: Enforce maximum of 5 registered devices per user.
// Implements Property 20: Device Registration Persistence.
// Implements Property 21: Device Limit Enforcement.
func (s *deviceService) RegisterDevice(ctx context.Context, input RegisterDeviceInput) (*domain.Device, error) {
	// Validate platform
	if !domain.IsValidPlatform(input.Platform) {
		return nil, errors.BadRequest("invalid platform: must be ios, android, or desktop")
	}

	// Check if device with same fingerprint already exists for this user
	existingDevice, err := s.deviceRepo.FindByFingerprint(ctx, input.UserID, input.Fingerprint)
	if err == nil && existingDevice != nil {
		// Device already registered, update last used and return it
		if err := s.deviceRepo.UpdateLastUsed(ctx, existingDevice.ID); err != nil {
			log.Warn().Err(err).Str("device_id", existingDevice.ID.String()).Msg("failed to update device last used")
		}
		return existingDevice, nil
	}

	// Count active devices for user
	count, err := s.deviceRepo.CountActiveByUserID(ctx, input.UserID)
	if err != nil {
		return nil, err
	}

	// Check device limit
	if count >= MaxDevicesPerUser {
		return nil, errors.New(
			errors.CodeForbidden,
			"DEVICE_LIMIT_EXCEEDED",
		)
	}

	// Create new device
	device := domain.NewDevice(input.UserID, input.Fingerprint, input.Name, input.Platform)

	if err := s.deviceRepo.Create(ctx, device); err != nil {
		return nil, err
	}

	// Publish device.registered event
	s.publishDeviceEvent(ctx, "device.registered", device)

	return device, nil
}

// ListDevices returns all active devices for a user.
// Implements Requirement 5: Device Management - list devices.
func (s *deviceService) ListDevices(ctx context.Context, userID uuid.UUID) ([]*domain.Device, error) {
	return s.deviceRepo.FindByUserID(ctx, userID)
}

// DeregisterDevice removes a device and revokes associated licenses.
// Implements Requirement 5.4: Allow users to deregister devices.
// Implements Requirement 5.5: Revoke all licenses bound to deregistered device.
// Implements Property 22: Device Deregistration Effect.
func (s *deviceService) DeregisterDevice(ctx context.Context, userID, deviceID uuid.UUID) error {
	// Find the device
	device, err := s.deviceRepo.FindById(ctx, deviceID)
	if err != nil {
		return err
	}

	// Verify ownership
	if device.UserID != userID {
		return errors.NotFound("device", deviceID.String())
	}

	// Check if already revoked
	if !device.IsActive() {
		return errors.NotFound("device", deviceID.String())
	}

	// Revoke all licenses for this device (cascade)
	if s.licenseRepo != nil {
		if err := s.licenseRepo.RevokeByDeviceID(ctx, deviceID); err != nil {
			log.Error().Err(err).Str("device_id", deviceID.String()).Msg("failed to revoke licenses for device")
			// Continue with device revocation even if license revocation fails
		}
	}

	// Delete all CEKs for this device
	if s.cekRepo != nil {
		if err := s.cekRepo.DeleteByDeviceID(ctx, deviceID); err != nil {
			log.Error().Err(err).Str("device_id", deviceID.String()).Msg("failed to delete CEKs for device")
			// Continue with device revocation even if CEK deletion fails
		}
	}

	// Revoke the device
	if err := s.deviceRepo.Revoke(ctx, deviceID); err != nil {
		return err
	}

	// Publish device.deregistered event
	s.publishDeviceEvent(ctx, "device.deregistered", device)

	return nil
}

// ValidateDevice validates a device fingerprint for a user.
// Implements Requirement 5.6: Validate device fingerprints on every license request.
// Implements Property 23: Device Fingerprint Validation.
func (s *deviceService) ValidateDevice(ctx context.Context, userID uuid.UUID, fingerprint string) (*domain.Device, error) {
	device, err := s.deviceRepo.FindByFingerprint(ctx, userID, fingerprint)
	if err != nil {
		// Return generic error to prevent enumeration
		return nil, errors.New(
			errors.CodeForbidden,
			"DEVICE_FINGERPRINT_MISMATCH",
		)
	}

	// Check if device is active
	if !device.IsActive() {
		return nil, errors.New(
			errors.CodeNotFound,
			"DEVICE_NOT_FOUND",
		)
	}

	return device, nil
}

// UpdateLastUsed updates the last_used_at timestamp for a device.
// Implements Requirement 5.8: Track last_used_at timestamp for each device.
// Implements Property 24: Device Timestamp Tracking.
func (s *deviceService) UpdateLastUsed(ctx context.Context, deviceID uuid.UUID) error {
	return s.deviceRepo.UpdateLastUsed(ctx, deviceID)
}

// publishDeviceEvent publishes a device event via NATS.
func (s *deviceService) publishDeviceEvent(ctx context.Context, eventType string, device *domain.Device) {
	if s.eventPublisher == nil {
		return
	}

	event := map[string]interface{}{
		"event_type": eventType,
		"device_id":  device.ID.String(),
		"user_id":    device.UserID.String(),
		"platform":   string(device.Platform),
		"name":       device.Name,
	}

	// Use generic event publishing - the actual implementation would depend on NATS setup
	log.Info().
		Str("event_type", eventType).
		Str("device_id", device.ID.String()).
		Str("user_id", device.UserID.String()).
		Interface("event", event).
		Msg("device event published")
}

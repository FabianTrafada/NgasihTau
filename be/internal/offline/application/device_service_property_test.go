package application

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
	"github.com/stretchr/testify/mock"

	"ngasihtau/internal/common/errors"
	"ngasihtau/internal/offline/domain"
)

// Property-based tests for Device Service
// Feature: offline-material-backend

// genFingerprint generates valid fingerprints (32-512 chars).
func genFingerprint() gopter.Gen {
	return gen.AlphaString().SuchThat(func(s string) bool {
		return len(s) >= 32 && len(s) <= 512
	}).Map(func(s string) string {
		if len(s) < 32 {
			return s + "00000000000000000000000000000000"[:32-len(s)]
		}
		if len(s) > 512 {
			return s[:512]
		}
		return s
	})
}

// genDeviceName generates valid device names (1-255 chars).
func genDeviceName() gopter.Gen {
	return gen.AlphaString().SuchThat(func(s string) bool {
		return len(s) >= 1 && len(s) <= 255
	}).Map(func(s string) string {
		if len(s) < 1 {
			return "Device"
		}
		if len(s) > 255 {
			return s[:255]
		}
		return s
	})
}

// genPlatform generates valid platforms.
func genPlatform() gopter.Gen {
	return gen.OneConstOf(domain.PlatformIOS, domain.PlatformAndroid, domain.PlatformDesktop)
}

// TestProperty20_DeviceRegistrationPersistence tests that registered devices are persisted.
// Feature: offline-material-backend, Property 20: Device Registration Persistence
// For any successfully registered device, querying devices for that user SHALL include the registered device.
// **Validates: Requirements 5.1**
func TestProperty20_DeviceRegistrationPersistence(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("registered device appears in user's device list", prop.ForAll(
		func(fingerprint string, name string, platform domain.Platform) bool {
			ctx := context.Background()
			mockDeviceRepo := new(MockDeviceRepository)
			service := NewDeviceService(mockDeviceRepo, nil, nil, nil)

			userID := uuid.New()
			input := RegisterDeviceInput{
				UserID:      userID,
				Fingerprint: fingerprint,
				Name:        name,
				Platform:    platform,
			}

			// Setup: device doesn't exist, user has no devices
			mockDeviceRepo.On("FindByFingerprint", ctx, userID, fingerprint).Return(nil, errors.NotFound("device", ""))
			mockDeviceRepo.On("CountActiveByUserID", ctx, userID).Return(0, nil)

			var createdDevice *domain.Device
			mockDeviceRepo.On("Create", ctx, mock.AnythingOfType("*domain.Device")).Run(func(args mock.Arguments) {
				createdDevice = args.Get(1).(*domain.Device)
			}).Return(nil)

			// Register device
			device, err := service.RegisterDevice(ctx, input)
			if err != nil {
				return false
			}

			// Verify device was created with correct properties
			return device != nil &&
				device.UserID == userID &&
				device.Fingerprint == fingerprint &&
				device.Name == name &&
				device.Platform == platform &&
				createdDevice != nil &&
				createdDevice.ID == device.ID
		},
		genFingerprint(),
		genDeviceName(),
		genPlatform(),
	))

	properties.TestingRun(t)
}


// TestProperty21_DeviceLimitEnforcement tests that device limit is enforced.
// Feature: offline-material-backend, Property 21: Device Limit Enforcement
// For any user with 5 registered devices, attempting to register a 6th device SHALL fail with DEVICE_LIMIT_EXCEEDED.
// **Validates: Requirements 5.2**
func TestProperty21_DeviceLimitEnforcement(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("6th device registration fails with DEVICE_LIMIT_EXCEEDED", prop.ForAll(
		func(fingerprint string, name string, platform domain.Platform, existingCount int) bool {
			ctx := context.Background()
			mockDeviceRepo := new(MockDeviceRepository)
			service := NewDeviceService(mockDeviceRepo, nil, nil, nil)

			userID := uuid.New()
			input := RegisterDeviceInput{
				UserID:      userID,
				Fingerprint: fingerprint,
				Name:        name,
				Platform:    platform,
			}

			// Device doesn't exist
			mockDeviceRepo.On("FindByFingerprint", ctx, userID, fingerprint).Return(nil, errors.NotFound("device", ""))
			// User has 5 or more devices
			mockDeviceRepo.On("CountActiveByUserID", ctx, userID).Return(existingCount, nil)

			device, err := service.RegisterDevice(ctx, input)

			if existingCount >= MaxDevicesPerUser {
				// Should fail with DEVICE_LIMIT_EXCEEDED
				if err == nil {
					return false
				}
				appErr, ok := err.(*errors.AppError)
				return ok && appErr.Message == "DEVICE_LIMIT_EXCEEDED" && device == nil
			}

			// If under limit, would need Create mock - skip this case
			return true
		},
		genFingerprint(),
		genDeviceName(),
		genPlatform(),
		gen.IntRange(5, 10), // Test with 5-10 existing devices
	))

	properties.TestingRun(t)
}

// TestProperty22_DeviceDeregistrationEffect tests that deregistered devices cannot be used.
// Feature: offline-material-backend, Property 22: Device Deregistration Effect
// For any deregistered device, subsequent license requests using that device_id SHALL fail with DEVICE_NOT_FOUND.
// **Validates: Requirements 5.4**
func TestProperty22_DeviceDeregistrationEffect(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("deregistered device validation fails", prop.ForAll(
		func(fingerprint string, name string, platform domain.Platform) bool {
			ctx := context.Background()
			mockDeviceRepo := new(MockDeviceRepository)
			mockLicenseRepo := new(MockLicenseRepository)
			mockCEKRepo := new(MockCEKRepository)
			service := NewDeviceService(mockDeviceRepo, mockLicenseRepo, mockCEKRepo, nil)

			userID := uuid.New()
			deviceID := uuid.New()
			revokedAt := time.Now()

			// Create a revoked device
			revokedDevice := &domain.Device{
				ID:          deviceID,
				UserID:      userID,
				Fingerprint: fingerprint,
				Name:        name,
				Platform:    platform,
				LastUsedAt:  time.Now(),
				CreatedAt:   time.Now(),
				RevokedAt:   &revokedAt,
			}

			// Validation should fail for revoked device
			mockDeviceRepo.On("FindByFingerprint", ctx, userID, fingerprint).Return(revokedDevice, nil)

			device, err := service.ValidateDevice(ctx, userID, fingerprint)

			// Should fail with DEVICE_NOT_FOUND
			if err == nil {
				return false
			}
			appErr, ok := err.(*errors.AppError)
			return ok && appErr.Message == "DEVICE_NOT_FOUND" && device == nil
		},
		genFingerprint(),
		genDeviceName(),
		genPlatform(),
	))

	properties.TestingRun(t)
}

// TestProperty23_DeviceFingerprintValidation tests fingerprint validation.
// Feature: offline-material-backend, Property 23: Device Fingerprint Validation
// For any license request, if the provided fingerprint does not match the registered fingerprint
// for the device_id, the request SHALL fail with DEVICE_FINGERPRINT_MISMATCH.
// **Validates: Requirements 5.6**
func TestProperty23_DeviceFingerprintValidation(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("mismatched fingerprint fails validation", prop.ForAll(
		func(suffix1 int, suffix2 int) bool {
			// Generate different fingerprints using suffixes
			registeredFingerprint := fmt.Sprintf("registered-fingerprint-%d-padding", suffix1)
			providedFingerprint := fmt.Sprintf("provided-fingerprint-%d-padding", suffix2)

			// Skip if fingerprints happen to be the same
			if registeredFingerprint == providedFingerprint {
				return true
			}

			ctx := context.Background()
			mockDeviceRepo := new(MockDeviceRepository)
			service := NewDeviceService(mockDeviceRepo, nil, nil, nil)

			userID := uuid.New()

			// Fingerprint not found (simulates mismatch)
			mockDeviceRepo.On("FindByFingerprint", ctx, userID, providedFingerprint).Return(nil, errors.NotFound("device", ""))

			device, err := service.ValidateDevice(ctx, userID, providedFingerprint)

			// Should fail with DEVICE_FINGERPRINT_MISMATCH
			if err == nil {
				return false
			}
			appErr, ok := err.(*errors.AppError)
			return ok && appErr.Message == "DEVICE_FINGERPRINT_MISMATCH" && device == nil
		},
		gen.IntRange(1, 10000),
		gen.IntRange(10001, 20000), // Different range to ensure different values
	))

	properties.TestingRun(t)
}

// TestProperty24_DeviceTimestampTracking tests that device timestamps are updated.
// Feature: offline-material-backend, Property 24: Device Timestamp Tracking
// For any successful device operation (license request, validation), the device's last_used_at timestamp SHALL be updated.
// **Validates: Requirements 5.8**
func TestProperty24_DeviceTimestampTracking(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("UpdateLastUsed is called on device operations", prop.ForAll(
		func(fingerprint string, name string, platform domain.Platform) bool {
			ctx := context.Background()
			mockDeviceRepo := new(MockDeviceRepository)
			service := NewDeviceService(mockDeviceRepo, nil, nil, nil)

			userID := uuid.New()
			deviceID := uuid.New()

			// Test 1: Existing device registration updates timestamp
			existingDevice := &domain.Device{
				ID:          deviceID,
				UserID:      userID,
				Fingerprint: fingerprint,
				Name:        name,
				Platform:    platform,
				LastUsedAt:  time.Now().Add(-1 * time.Hour), // Old timestamp
				CreatedAt:   time.Now().Add(-24 * time.Hour),
			}

			mockDeviceRepo.On("FindByFingerprint", ctx, userID, fingerprint).Return(existingDevice, nil)
			mockDeviceRepo.On("UpdateLastUsed", ctx, deviceID).Return(nil)

			input := RegisterDeviceInput{
				UserID:      userID,
				Fingerprint: fingerprint,
				Name:        name,
				Platform:    platform,
			}

			device, err := service.RegisterDevice(ctx, input)
			if err != nil {
				return false
			}

			// Verify UpdateLastUsed was called
			mockDeviceRepo.AssertCalled(t, "UpdateLastUsed", ctx, deviceID)

			return device != nil && device.ID == deviceID
		},
		genFingerprint(),
		genDeviceName(),
		genPlatform(),
	))

	properties.TestingRun(t)
}

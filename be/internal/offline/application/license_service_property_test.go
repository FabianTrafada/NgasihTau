package application

import (
	"context"
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

// Property-based tests for License Service
// Feature: offline-material-backend

// genUUID generates random UUIDs.
func genUUID() gopter.Gen {
	return gen.SliceOfN(16, gen.UInt8()).Map(func(bytes []byte) uuid.UUID {
		var id uuid.UUID
		copy(id[:], bytes)
		return id
	})
}

// genLicenseFingerprint generates valid fingerprints for license tests.
func genLicenseFingerprint() gopter.Gen {
	return gen.AlphaString().Map(func(s string) string {
		if len(s) < 32 {
			return s + "00000000000000000000000000000000"[:32-len(s)]
		}
		if len(s) > 512 {
			return s[:512]
		}
		return s
	})
}

// TestProperty11_LicenseAccessControl tests that licenses are only issued to users with material access.
// Feature: offline-material-backend, Property 11: License Access Control
// For any license request, if the user does not have access to the material, the license SHALL NOT be issued.
// **Validates: Requirements 3.1**
func TestProperty11_LicenseAccessControl(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("license not issued when user lacks material access", prop.ForAll(
		func(fingerprint string) bool {
			ctx := context.Background()
			mockLicenseRepo := new(MockLicenseRepository)
			mockDeviceRepo := new(MockDeviceRepository)
			mockAccessChecker := new(MockLicenseMaterialAccessChecker)

			service := NewLicenseService(mockLicenseRepo, mockDeviceRepo, mockAccessChecker, nil)

			userID := uuid.New()
			materialID := uuid.New()
			deviceID := uuid.New()

			device := &domain.Device{
				ID:          deviceID,
				UserID:      userID,
				Fingerprint: fingerprint,
				Name:        "Test Device",
				Platform:    domain.PlatformIOS,
				LastUsedAt:  time.Now(),
				CreatedAt:   time.Now(),
			}

			input := IssueLicenseInput{
				UserID:      userID,
				MaterialID:  materialID,
				DeviceID:    deviceID,
				Fingerprint: fingerprint,
			}

			// Device exists and fingerprint matches
			mockDeviceRepo.On("FindById", ctx, deviceID).Return(device, nil)
			// User does NOT have access to material
			mockAccessChecker.On("CheckAccess", ctx, userID, materialID).Return(false, nil)

			license, err := service.IssueLicense(ctx, input)

			// License should NOT be issued
			return license == nil && err != nil
		},
		genLicenseFingerprint(),
	))

	properties.TestingRun(t)
}


// TestProperty12_LicenseExpirationStructure tests that licenses have proper expiration structure.
// Feature: offline-material-backend, Property 12: License Expiration Structure
// For any issued license, the expires_at field SHALL be set to a future timestamp (default 30 days),
// and offline_grace_period SHALL be set (default 72 hours).
// **Validates: Requirements 3.2, 3.3**
func TestProperty12_LicenseExpirationStructure(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("issued license has valid expiration and grace period", prop.ForAll(
		func(fingerprint string) bool {
			ctx := context.Background()
			mockLicenseRepo := new(MockLicenseRepository)
			mockDeviceRepo := new(MockDeviceRepository)
			mockAccessChecker := new(MockLicenseMaterialAccessChecker)

			service := NewLicenseService(mockLicenseRepo, mockDeviceRepo, mockAccessChecker, nil)

			userID := uuid.New()
			materialID := uuid.New()
			deviceID := uuid.New()

			device := &domain.Device{
				ID:          deviceID,
				UserID:      userID,
				Fingerprint: fingerprint,
				Name:        "Test Device",
				Platform:    domain.PlatformIOS,
				LastUsedAt:  time.Now(),
				CreatedAt:   time.Now(),
			}

			input := IssueLicenseInput{
				UserID:      userID,
				MaterialID:  materialID,
				DeviceID:    deviceID,
				Fingerprint: fingerprint,
			}

			mockDeviceRepo.On("FindById", ctx, deviceID).Return(device, nil)
			mockAccessChecker.On("CheckAccess", ctx, userID, materialID).Return(true, nil)
			mockLicenseRepo.On("FindByUserAndMaterial", ctx, userID, materialID, deviceID).Return(nil, errors.NotFound("license", ""))
			mockLicenseRepo.On("Create", ctx, mock.AnythingOfType("*domain.License")).Return(nil)
			mockDeviceRepo.On("UpdateLastUsed", ctx, deviceID).Return(nil)

			beforeIssue := time.Now()
			license, err := service.IssueLicense(ctx, input)
			if err != nil {
				return false
			}

			// Verify expiration is approximately 30 days in the future
			expectedExpiration := beforeIssue.Add(domain.DefaultLicenseExpiration)
			expirationDiff := license.ExpiresAt.Sub(expectedExpiration)
			expirationValid := expirationDiff > -time.Minute && expirationDiff < time.Minute

			// Verify grace period is 72 hours
			gracePeriodValid := license.OfflineGracePeriod == domain.DefaultOfflineGracePeriod

			// Verify expiration is in the future
			expiresInFuture := license.ExpiresAt.After(beforeIssue)

			return expirationValid && gracePeriodValid && expiresInFuture
		},
		genLicenseFingerprint(),
	))

	properties.TestingRun(t)
}

// TestProperty13_LicenseValidationTimestampUpdate tests that validation updates the timestamp.
// Feature: offline-material-backend, Property 13: License Validation Timestamp Update
// For any successful license validation, the last_validated_at timestamp SHALL be updated to the current time.
// **Validates: Requirements 3.4**
func TestProperty13_LicenseValidationTimestampUpdate(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("validation updates last_validated_at timestamp", prop.ForAll(
		func(fingerprint string) bool {
			ctx := context.Background()
			mockLicenseRepo := new(MockLicenseRepository)
			mockDeviceRepo := new(MockDeviceRepository)

			service := NewLicenseService(mockLicenseRepo, mockDeviceRepo, nil, nil)

			userID := uuid.New()
			materialID := uuid.New()
			deviceID := uuid.New()
			licenseID := uuid.New()
			nonce := "test-nonce-1234567890abcdef1234567890abcdef"

			device := &domain.Device{
				ID:          deviceID,
				UserID:      userID,
				Fingerprint: fingerprint,
				Name:        "Test Device",
				Platform:    domain.PlatformIOS,
				LastUsedAt:  time.Now(),
				CreatedAt:   time.Now(),
			}

			// License was validated 1 hour ago
			oldValidatedAt := time.Now().Add(-1 * time.Hour)
			license := &domain.License{
				ID:                 licenseID,
				UserID:             userID,
				MaterialID:         materialID,
				DeviceID:           deviceID,
				Status:             domain.LicenseStatusActive,
				ExpiresAt:          time.Now().Add(30 * 24 * time.Hour),
				OfflineGracePeriod: 72 * time.Hour,
				LastValidatedAt:    oldValidatedAt,
				Nonce:              nonce,
				CreatedAt:          time.Now().Add(-24 * time.Hour),
			}

			input := ValidateLicenseInput{
				LicenseID:   licenseID,
				DeviceID:    deviceID,
				Fingerprint: fingerprint,
				Nonce:       nonce,
			}

			mockLicenseRepo.On("FindByID", ctx, licenseID).Return(license, nil)
			mockDeviceRepo.On("FindById", ctx, deviceID).Return(device, nil)
			mockLicenseRepo.On("UpdateValidation", ctx, licenseID, mock.AnythingOfType("string")).Return(nil)
			mockDeviceRepo.On("UpdateLastUsed", ctx, deviceID).Return(nil)

			beforeValidation := time.Now()
			result, err := service.ValidateLicense(ctx, input)
			if err != nil {
				return false
			}

			// Verify last_validated_at was updated to approximately now
			timeDiff := result.LastValidatedAt.Sub(beforeValidation)
			return timeDiff >= 0 && timeDiff < time.Second
		},
		genLicenseFingerprint(),
	))

	properties.TestingRun(t)
}


// TestProperty14_LicenseCascadingRevocation tests that device deregistration revokes licenses.
// Feature: offline-material-backend, Property 14: License Cascading Revocation
// For any device deregistration or material access revocation, all associated licenses SHALL be marked as revoked.
// **Validates: Requirements 3.5, 5.5**
func TestProperty14_LicenseCascadingRevocation(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("revoking by device revokes all device licenses", prop.ForAll(
		func(numLicenses int) bool {
			if numLicenses < 1 {
				numLicenses = 1
			}
			if numLicenses > 10 {
				numLicenses = 10
			}

			ctx := context.Background()
			mockLicenseRepo := new(MockLicenseRepository)

			service := NewLicenseService(mockLicenseRepo, nil, nil, nil)

			deviceID := uuid.New()

			// Setup: RevokeByDeviceID should be called
			mockLicenseRepo.On("RevokeByDeviceID", ctx, deviceID).Return(nil)

			err := service.RevokeByDevice(ctx, deviceID)

			// Verify revocation was called
			mockLicenseRepo.AssertCalled(t, "RevokeByDeviceID", ctx, deviceID)
			return err == nil
		},
		gen.IntRange(1, 10),
	))

	properties.Property("revoking by material revokes all material licenses", prop.ForAll(
		func(numLicenses int) bool {
			if numLicenses < 1 {
				numLicenses = 1
			}
			if numLicenses > 10 {
				numLicenses = 10
			}

			ctx := context.Background()
			mockLicenseRepo := new(MockLicenseRepository)

			service := NewLicenseService(mockLicenseRepo, nil, nil, nil)

			materialID := uuid.New()

			// Setup: RevokeByMaterialID should be called
			mockLicenseRepo.On("RevokeByMaterialID", ctx, materialID).Return(nil)

			err := service.RevokeByMaterial(ctx, materialID)

			// Verify revocation was called
			mockLicenseRepo.AssertCalled(t, "RevokeByMaterialID", ctx, materialID)
			return err == nil
		},
		gen.IntRange(1, 10),
	))

	properties.TestingRun(t)
}

// TestProperty15_LicenseRenewalExtension tests that renewal extends expiration.
// Feature: offline-material-backend, Property 15: License Renewal Extension
// For any valid license renewal request, the new expires_at SHALL be greater than the previous expires_at.
// **Validates: Requirements 3.6**
func TestProperty15_LicenseRenewalExtension(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("renewal extends expiration beyond original", prop.ForAll(
		func(fingerprint string, daysUntilExpiry int) bool {
			if daysUntilExpiry < 1 {
				daysUntilExpiry = 1
			}
			if daysUntilExpiry > 29 {
				daysUntilExpiry = 29 // Less than default 30 days
			}

			ctx := context.Background()
			mockLicenseRepo := new(MockLicenseRepository)
			mockDeviceRepo := new(MockDeviceRepository)

			service := NewLicenseService(mockLicenseRepo, mockDeviceRepo, nil, nil)

			userID := uuid.New()
			materialID := uuid.New()
			deviceID := uuid.New()
			licenseID := uuid.New()

			device := &domain.Device{
				ID:          deviceID,
				UserID:      userID,
				Fingerprint: fingerprint,
				Name:        "Test Device",
				Platform:    domain.PlatformIOS,
				LastUsedAt:  time.Now(),
				CreatedAt:   time.Now(),
			}

			// License expires in daysUntilExpiry days (less than default 30)
			originalExpiration := time.Now().Add(time.Duration(daysUntilExpiry) * 24 * time.Hour)
			license := &domain.License{
				ID:                 licenseID,
				UserID:             userID,
				MaterialID:         materialID,
				DeviceID:           deviceID,
				Status:             domain.LicenseStatusActive,
				ExpiresAt:          originalExpiration,
				OfflineGracePeriod: 72 * time.Hour,
				LastValidatedAt:    time.Now(),
				Nonce:              "test-nonce",
				CreatedAt:          time.Now().Add(-24 * time.Hour),
			}

			input := RenewLicenseInput{
				LicenseID:   licenseID,
				DeviceID:    deviceID,
				Fingerprint: fingerprint,
			}

			mockLicenseRepo.On("FindByID", ctx, licenseID).Return(license, nil)
			mockDeviceRepo.On("FindById", ctx, deviceID).Return(device, nil)
			mockLicenseRepo.On("UpdateExpiration", ctx, licenseID, mock.AnythingOfType("time.Time")).Return(nil)
			mockLicenseRepo.On("UpdateValidation", ctx, licenseID, mock.AnythingOfType("string")).Return(nil)
			mockDeviceRepo.On("UpdateLastUsed", ctx, deviceID).Return(nil)

			result, err := service.RenewLicense(ctx, input)
			if err != nil {
				return false
			}

			// New expiration should be after original expiration
			return result.ExpiresAt.After(originalExpiration)
		},
		genLicenseFingerprint(),
		gen.IntRange(1, 29),
	))

	properties.TestingRun(t)
}


// TestProperty16_OfflineGracePeriodEnforcement tests offline grace period enforcement.
// Feature: offline-material-backend, Property 16: Offline Grace Period Enforcement
// For any license where (current_time - last_validated_at) exceeds offline_grace_period,
// validation SHALL fail with LICENSE_OFFLINE_EXPIRED.
// **Validates: Requirements 3.7**
func TestProperty16_OfflineGracePeriodEnforcement(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("license entity detects grace period expiration", prop.ForAll(
		func(hoursOverGrace int) bool {
			if hoursOverGrace < 1 {
				hoursOverGrace = 1
			}
			if hoursOverGrace > 100 {
				hoursOverGrace = 100
			}

			userID := uuid.New()
			materialID := uuid.New()
			deviceID := uuid.New()
			licenseID := uuid.New()

			// License was validated more than 72 hours ago (grace period exceeded)
			gracePeriod := 72 * time.Hour
			lastValidated := time.Now().Add(-gracePeriod - time.Duration(hoursOverGrace)*time.Hour)

			license := &domain.License{
				ID:                 licenseID,
				UserID:             userID,
				MaterialID:         materialID,
				DeviceID:           deviceID,
				Status:             domain.LicenseStatusActive,
				ExpiresAt:          time.Now().Add(30 * 24 * time.Hour), // Not expired
				OfflineGracePeriod: gracePeriod,
				LastValidatedAt:    lastValidated,
				Nonce:              "test-nonce",
				CreatedAt:          time.Now().Add(-100 * time.Hour),
			}

			// Verify the license entity correctly identifies grace period expiration
			return license.IsOfflineGraceExpired()
		},
		gen.IntRange(1, 100),
	))

	properties.Property("license entity allows access within grace period", prop.ForAll(
		func(hoursWithinGrace int) bool {
			if hoursWithinGrace < 0 {
				hoursWithinGrace = 0
			}
			if hoursWithinGrace > 71 {
				hoursWithinGrace = 71
			}

			userID := uuid.New()
			materialID := uuid.New()
			deviceID := uuid.New()
			licenseID := uuid.New()

			// License was validated within grace period
			gracePeriod := 72 * time.Hour
			lastValidated := time.Now().Add(-time.Duration(hoursWithinGrace) * time.Hour)

			license := &domain.License{
				ID:                 licenseID,
				UserID:             userID,
				MaterialID:         materialID,
				DeviceID:           deviceID,
				Status:             domain.LicenseStatusActive,
				ExpiresAt:          time.Now().Add(30 * 24 * time.Hour), // Not expired
				OfflineGracePeriod: gracePeriod,
				LastValidatedAt:    lastValidated,
				Nonce:              "test-nonce",
				CreatedAt:          time.Now().Add(-100 * time.Hour),
			}

			// Verify the license entity allows access within grace period
			return !license.IsOfflineGraceExpired()
		},
		gen.IntRange(0, 71),
	))

	properties.TestingRun(t)
}

// TestProperty27_LicenseNonceUniqueness tests that nonces are unique.
// Feature: offline-material-backend, Property 27: License Nonce Uniqueness
// For any two licenses (even for the same user, material, device combination after revocation
// and re-issuance), the nonces SHALL be different.
// **Validates: Requirements 8.4**
func TestProperty27_LicenseNonceUniqueness(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("generated nonces are unique", prop.ForAll(
		func(count int) bool {
			if count < 2 {
				count = 2
			}
			if count > 1000 {
				count = 1000
			}

			nonces := make(map[string]bool)
			for i := 0; i < count; i++ {
				nonce, err := generateNonce()
				if err != nil {
					return false
				}
				if nonces[nonce] {
					return false // Duplicate found
				}
				nonces[nonce] = true
			}
			return true
		},
		gen.IntRange(2, 1000),
	))

	properties.Property("nonces have correct length", prop.ForAll(
		func(_ int) bool {
			nonce, err := generateNonce()
			if err != nil {
				return false
			}
			return len(nonce) == domain.NonceHexLength
		},
		gen.Int(),
	))

	properties.TestingRun(t)
}

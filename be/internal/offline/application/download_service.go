// Package application contains the business logic and use cases for the Offline Material Service.
package application

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"ngasihtau/internal/offline/domain"
)

// DownloadService handles material download operations.
// Implements Requirement 4: Download API.
type DownloadService interface {
	// PrepareDownload prepares a material for download, returning manifest and presigned URL.
	PrepareDownload(ctx context.Context, input PrepareDownloadInput) (*PrepareDownloadOutput, error)
}

// PrepareDownloadInput contains input for preparing a download.
type PrepareDownloadInput struct {
	UserID      uuid.UUID
	MaterialID  uuid.UUID
	DeviceID    uuid.UUID
	LicenseID   uuid.UUID
	Fingerprint string
}

// PrepareDownloadOutput contains the result of preparing a download.
type PrepareDownloadOutput struct {
	Manifest    *domain.DownloadManifest
	DownloadURL string
	ExpiresAt   time.Time
}

// downloadService implements DownloadService.
type downloadService struct {
	licenseService        LicenseService
	deviceService         DeviceService
	encryptionService     *EncryptionService
	encryptedMaterialRepo domain.EncryptedMaterialRepository
	rateLimiter           RateLimiter
	eventPublisher        OfflineEventPublisher
	storage               MinIOStorageClient
	presignedURLExpiry    time.Duration
}

// DownloadServiceConfig holds configuration for the Download Service.
type DownloadServiceConfig struct {
	PresignedURLExpiry time.Duration
}

// DefaultDownloadServiceConfig returns the default download service configuration.
func DefaultDownloadServiceConfig() DownloadServiceConfig {
	return DownloadServiceConfig{
		PresignedURLExpiry: 1 * time.Hour,
	}
}

// NewDownloadService creates a new Download Service.
func NewDownloadService(
	licenseService LicenseService,
	deviceService DeviceService,
	encryptionService *EncryptionService,
	encryptedMaterialRepo domain.EncryptedMaterialRepository,
	rateLimiter RateLimiter,
	eventPublisher OfflineEventPublisher,
	storage MinIOStorageClient,
	config DownloadServiceConfig,
) DownloadService {
	return &downloadService{
		licenseService:        licenseService,
		deviceService:         deviceService,
		encryptionService:     encryptionService,
		encryptedMaterialRepo: encryptedMaterialRepo,
		rateLimiter:           rateLimiter,
		eventPublisher:        eventPublisher,
		storage:               storage,
		presignedURLExpiry:    config.PresignedURLExpiry,
	}
}

// PrepareDownload prepares a material for download.
// Implements Requirement 4.1-4.8: Download API with validation chain.
func (s *downloadService) PrepareDownload(ctx context.Context, input PrepareDownloadInput) (*PrepareDownloadOutput, error) {
	// Step 1: Validate device (Requirement 4.3)
	device, err := s.deviceService.ValidateDevice(ctx, input.UserID, input.Fingerprint)
	if err != nil {
		log.Warn().
			Err(err).
			Str("user_id", input.UserID.String()).
			Str("device_id", input.DeviceID.String()).
			Msg("device validation failed for download")
		return nil, err
	}

	// Verify device ID matches
	if device.ID != input.DeviceID {
		return nil, domain.ErrDeviceFingerprintMismatch
	}

	// Step 2: Validate license (Requirement 4.2)
	license, err := s.licenseService.ValidateLicense(ctx, ValidateLicenseInput{
		LicenseID:   input.LicenseID,
		DeviceID:    input.DeviceID,
		Fingerprint: input.Fingerprint,
		Nonce:       "", // Nonce validation is optional for download
	})
	if err != nil {
		log.Warn().
			Err(err).
			Str("license_id", input.LicenseID.String()).
			Msg("license validation failed for download")
		return nil, err
	}

	// Verify license is for the correct material and user
	if license.MaterialID != input.MaterialID {
		return nil, domain.NewOfflineError(domain.ErrCodeLicenseNotFound, "license does not match material")
	}
	if license.UserID != input.UserID {
		return nil, domain.NewOfflineError(domain.ErrCodeMaterialAccessDenied, "license does not belong to user")
	}

	// Step 3: Check rate limit (Requirement 4.6)
	if s.rateLimiter != nil {
		allowed, _, _, err := s.rateLimiter.CheckDownloadLimit(ctx, input.UserID)
		if err != nil {
			log.Warn().Err(err).Msg("rate limit check failed")
			// Continue on rate limit check failure (fail open)
		} else if !allowed {
			return nil, domain.ErrRateLimitExceeded
		}
	}

	// Step 4: Get encrypted material
	materials, err := s.encryptedMaterialRepo.FindByMaterialID(ctx, input.MaterialID)
	if err != nil || len(materials) == 0 {
		return nil, domain.NewOfflineError(domain.ErrCodeMaterialNotFound, "encrypted material not found, encryption may be in progress")
	}

	// Use the first available encrypted material
	encryptedMaterial := materials[0]

	// Step 5: Generate presigned URL (Requirement 4.4)
	presignedURL, err := s.storage.GeneratePresignedGetURL(ctx, encryptedMaterial.EncryptedFileURL, s.presignedURLExpiry)
	if err != nil {
		log.Error().Err(err).Msg("failed to generate presigned URL")
		return nil, domain.WrapOfflineError(domain.ErrCodeStorageError, "failed to generate download URL", err)
	}

	// Step 6: Increment download counter
	if s.rateLimiter != nil {
		if _, err := s.rateLimiter.IncrementDownload(ctx, input.UserID); err != nil {
			log.Warn().Err(err).Msg("failed to increment download count")
			// Continue on failure (non-critical)
		}
	}

	// Step 7: Publish download event
	if s.eventPublisher != nil {
		s.eventPublisher.PublishMaterialDownloaded(ctx, MaterialDownloadEvent{
			UserID:     input.UserID,
			MaterialID: input.MaterialID,
			DeviceID:   input.DeviceID,
			LicenseID:  input.LicenseID,
		})
	}

	log.Info().
		Str("user_id", input.UserID.String()).
		Str("material_id", input.MaterialID.String()).
		Str("device_id", input.DeviceID.String()).
		Msg("download prepared successfully")

	return &PrepareDownloadOutput{
		Manifest:    &encryptedMaterial.Manifest,
		DownloadURL: presignedURL,
		ExpiresAt:   time.Now().Add(s.presignedURLExpiry),
	}, nil
}

// Note: RateLimiter interface is defined in rate_limiter.go

// MaterialDownloadEvent represents a material download event.
type MaterialDownloadEvent struct {
	UserID     uuid.UUID
	MaterialID uuid.UUID
	DeviceID   uuid.UUID
	LicenseID  uuid.UUID
}

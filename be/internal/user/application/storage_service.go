// Package application contains the business logic and use cases for the User Service.
package application

import (
	"context"

	"github.com/google/uuid"

	"ngasihtau/internal/common/config"
	"ngasihtau/internal/common/errors"
	"ngasihtau/internal/user/domain"
)

// StorageService defines the interface for storage-related business operations.
// Implements requirements 2.1-2.5, 7.1-7.3 for storage information.
type StorageService interface {
	// GetStorageInfo returns storage usage information for a user.
	// Implements requirements 2.1-2.5, 7.1-7.3.
	GetStorageInfo(ctx context.Context, userID uuid.UUID) (*domain.StorageInfo, error)

	// CheckStorageQuota checks if user has enough storage for a file.
	// Implements requirements 3.1-3.4.
	CheckStorageQuota(ctx context.Context, userID uuid.UUID, fileSize int64) error

	// UpdateTier updates a user's subscription tier.
	// Implements requirements 6.1-6.4.
	UpdateTier(ctx context.Context, userID uuid.UUID, tier domain.Tier) error
}

// storageService implements the StorageService interface.
type storageService struct {
	userRepo      domain.UserRepository
	storageRepo   domain.StorageRepository
	storageConfig config.StorageConfig
}

// NewStorageService creates a new StorageService instance.
func NewStorageService(
	userRepo domain.UserRepository,
	storageRepo domain.StorageRepository,
	storageConfig config.StorageConfig,
) StorageService {
	return &storageService{
		userRepo:      userRepo,
		storageRepo:   storageRepo,
		storageConfig: storageConfig,
	}
}

// GetStorageInfo returns storage usage information for a user.
// Implements requirements 2.1-2.5, 7.1-7.3.
func (s *storageService) GetStorageInfo(ctx context.Context, userID uuid.UUID) (*domain.StorageInfo, error) {
	// Get user to retrieve their tier
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Get storage usage from repository
	usedBytes, err := s.storageRepo.GetUserStorageUsage(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Calculate quota based on tier from config
	quotaBytes := s.getQuotaForTier(user.Tier)

	// Calculate remaining bytes (can be negative if over quota after downgrade)
	remainingBytes := quotaBytes - usedBytes
	if remainingBytes < 0 {
		remainingBytes = 0
	}

	// Calculate usage percentage
	var usagePercent float64
	if quotaBytes > 0 {
		usagePercent = (float64(usedBytes) / float64(quotaBytes)) * 100
	}

	// Set warning flag based on thresholds (80% warning, 90% critical)
	// Implements requirements 7.1, 7.2
	var warning string
	if usagePercent >= 90 {
		warning = "critical"
	} else if usagePercent >= 80 {
		warning = "warning"
	}

	// Include next tier upgrade option if available
	// Implements requirement 7.3
	var nextTier *domain.Tier
	var nextTierQuota *int64
	switch user.Tier {
	case domain.TierFree:
		premium := domain.TierPremium
		nextTier = &premium
		premiumQuota := s.getQuotaForTier(domain.TierPremium)
		nextTierQuota = &premiumQuota
	case domain.TierPremium:
		pro := domain.TierPro
		nextTier = &pro
		proQuota := s.getQuotaForTier(domain.TierPro)
		nextTierQuota = &proQuota
	case domain.TierPro:
		// Pro is the highest tier, no upgrade available
		nextTier = nil
		nextTierQuota = nil
	}

	return &domain.StorageInfo{
		UsedBytes:      usedBytes,
		QuotaBytes:     quotaBytes,
		RemainingBytes: remainingBytes,
		UsagePercent:   usagePercent,
		Tier:           user.Tier,
		Warning:        warning,
		NextTier:       nextTier,
		NextTierQuota:  nextTierQuota,
	}, nil
}

// getQuotaForTier returns the storage quota in bytes for a given tier.
// Converts GB from config to bytes.
func (s *storageService) getQuotaForTier(tier domain.Tier) int64 {
	const bytesPerGB = 1024 * 1024 * 1024 // 1 GB = 1,073,741,824 bytes

	switch tier {
	case domain.TierFree:
		return s.storageConfig.FreeQuotaGB * bytesPerGB
	case domain.TierPremium:
		return s.storageConfig.PremiumQuotaGB * bytesPerGB
	case domain.TierPro:
		return s.storageConfig.ProQuotaGB * bytesPerGB
	default:
		// Default to free tier quota for unknown tiers
		return s.storageConfig.FreeQuotaGB * bytesPerGB
	}
}

// CheckStorageQuota checks if user has enough storage for a file.
// Implements requirements 3.1-3.4.
func (s *storageService) CheckStorageQuota(ctx context.Context, userID uuid.UUID, fileSize int64) error {
	// Get user to retrieve their tier
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return err
	}

	// Get current storage usage
	usedBytes, err := s.storageRepo.GetUserStorageUsage(ctx, userID)
	if err != nil {
		return err
	}

	// Get quota for user's tier
	quotaBytes := s.getQuotaForTier(user.Tier)

	// Check if upload would exceed quota
	// Implements requirements 3.1, 3.2
	if usedBytes+fileSize > quotaBytes {
		return errors.StorageQuotaExceeded(usedBytes, quotaBytes, fileSize, string(user.Tier))
	}

	return nil
}

// UpdateTier updates a user's subscription tier.
// Implements requirements 6.1-6.4.
func (s *storageService) UpdateTier(ctx context.Context, userID uuid.UUID, tier domain.Tier) error {
	// Validate tier value
	if !domain.IsValidTier(tier) {
		return errors.BadRequest("invalid tier value")
	}

	// Get user
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return err
	}

	// Update user's tier
	user.Tier = tier

	// Save changes - new quota applies immediately on next query
	// Implements requirements 6.2, 6.3
	if err := s.userRepo.Update(ctx, user); err != nil {
		return err
	}

	return nil
}

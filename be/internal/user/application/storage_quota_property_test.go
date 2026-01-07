// Package application contains property-based tests for storage quota enforcement.
package application

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"

	"ngasihtau/internal/common/config"
	"ngasihtau/internal/common/errors"
	"ngasihtau/internal/user/domain"
)

// **Feature: user-storage-limit, Property 3: Storage Quota Enforcement**
//
// *For any* user attempting to upload a file:
// - IF `current_usage + file_size > quota` THEN the upload SHALL be rejected
// - IF `current_usage + file_size <= quota` THEN the upload SHALL be allowed
//
// **Validates: Requirements 3.1, 3.2, 3.3**

func TestProperty_StorageQuotaEnforcement(t *testing.T) {
	parameters := gopter.DefaultTestParametersWithSeed(12345)
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	// Default configuration
	defaultConfig := config.StorageConfig{
		FreeQuotaGB:    1,
		PremiumQuotaGB: 5,
		ProQuotaGB:     20,
	}

	const bytesPerGB = int64(1024 * 1024 * 1024)

	// Property 3.1: Upload exceeding quota is rejected
	// Validates: Requirement 3.2 - IF the upload would exceed the user's storage quota
	// THEN THE Storage_Limit_System SHALL reject the upload with error code STORAGE_QUOTA_EXCEEDED
	properties.Property("upload exceeding quota is rejected", prop.ForAll(
		func(tier domain.Tier, usagePercent float64, fileSizePercent float64) bool {
			svc, userRepo, storageRepo := newTestStorageService(defaultConfig)
			ctx := context.Background()

			// Create user with specified tier
			user := domain.NewUser("test@example.com", "hash", "Test User")
			user.Tier = tier
			userRepo.users[user.ID] = user
			userRepo.emailIndex[user.Email] = user

			// Calculate quota for tier
			quotaBytes := getQuotaForTier(tier, defaultConfig)

			// Set current usage as percentage of quota (0-99%)
			currentUsage := int64(float64(quotaBytes) * usagePercent / 100.0)
			storageRepo.usage[user.ID] = currentUsage

			// Calculate remaining space
			remainingSpace := quotaBytes - currentUsage

			// File size that exceeds remaining space (101% to 200% of remaining)
			fileSize := int64(float64(remainingSpace) * (1.0 + fileSizePercent/100.0))
			if fileSize <= 0 {
				fileSize = 1 // Ensure at least 1 byte
			}

			// Verify this would exceed quota
			if currentUsage+fileSize <= quotaBytes {
				return true // Skip if doesn't actually exceed
			}

			err := svc.CheckStorageQuota(ctx, user.ID, fileSize)

			// Should return storage quota exceeded error
			if err == nil {
				return false
			}

			appErr, ok := err.(*errors.AppError)
			if !ok {
				return false
			}

			return appErr.Code == errors.CodeStorageQuotaExceeded
		},
		genValidTierForStorage(),
		gen.Float64Range(0, 99),  // Current usage 0-99% of quota
		gen.Float64Range(1, 100), // File size 101-200% of remaining space
	))

	// Property 3.2: Upload within quota is allowed
	// Validates: Requirement 3.3 - IF the upload would not exceed the quota
	// THEN THE Storage_Limit_System SHALL allow the upload to proceed
	properties.Property("upload within quota is allowed", prop.ForAll(
		func(tier domain.Tier, usagePercent float64, fileSizePercent float64) bool {
			svc, userRepo, storageRepo := newTestStorageService(defaultConfig)
			ctx := context.Background()

			// Create user with specified tier
			user := domain.NewUser("test@example.com", "hash", "Test User")
			user.Tier = tier
			userRepo.users[user.ID] = user
			userRepo.emailIndex[user.Email] = user

			// Calculate quota for tier
			quotaBytes := getQuotaForTier(tier, defaultConfig)

			// Set current usage as percentage of quota (0-90%)
			currentUsage := int64(float64(quotaBytes) * usagePercent / 100.0)
			storageRepo.usage[user.ID] = currentUsage

			// Calculate remaining space
			remainingSpace := quotaBytes - currentUsage
			if remainingSpace <= 0 {
				return true // Skip if no remaining space
			}

			// File size within remaining space (1% to 99% of remaining)
			fileSize := int64(float64(remainingSpace) * fileSizePercent / 100.0)
			if fileSize <= 0 {
				fileSize = 1 // Ensure at least 1 byte
			}

			// Verify this would NOT exceed quota
			if currentUsage+fileSize > quotaBytes {
				return true // Skip if would exceed
			}

			err := svc.CheckStorageQuota(ctx, user.ID, fileSize)

			// Should return nil (no error)
			return err == nil
		},
		genValidTierForStorage(),
		gen.Float64Range(0, 90), // Current usage 0-90% of quota
		gen.Float64Range(1, 99), // File size 1-99% of remaining space
	))

	// Property 3.3: Upload exactly at quota limit is allowed
	// Validates: Requirement 3.1 - WHEN a user attempts to upload a file
	// THEN THE Storage_Limit_System SHALL calculate if the new file would exceed the quota
	properties.Property("upload exactly at quota limit is allowed", prop.ForAll(
		func(tier domain.Tier, usagePercent float64) bool {
			svc, userRepo, storageRepo := newTestStorageService(defaultConfig)
			ctx := context.Background()

			// Create user with specified tier
			user := domain.NewUser("test@example.com", "hash", "Test User")
			user.Tier = tier
			userRepo.users[user.ID] = user
			userRepo.emailIndex[user.Email] = user

			// Calculate quota for tier
			quotaBytes := getQuotaForTier(tier, defaultConfig)

			// Set current usage as percentage of quota
			currentUsage := int64(float64(quotaBytes) * usagePercent / 100.0)
			storageRepo.usage[user.ID] = currentUsage

			// File size exactly fills remaining space
			fileSize := quotaBytes - currentUsage
			if fileSize <= 0 {
				return true // Skip if no remaining space
			}

			err := svc.CheckStorageQuota(ctx, user.ID, fileSize)

			// Should return nil (no error) - exactly at limit is allowed
			return err == nil
		},
		genValidTierForStorage(),
		gen.Float64Range(0, 99), // Current usage 0-99% of quota
	))

	// Property 3.4: Zero-size file is always allowed (when not already over quota)
	properties.Property("zero-size file is allowed when not over quota", prop.ForAll(
		func(tier domain.Tier, usagePercent float64) bool {
			svc, userRepo, storageRepo := newTestStorageService(defaultConfig)
			ctx := context.Background()

			// Create user with specified tier
			user := domain.NewUser("test@example.com", "hash", "Test User")
			user.Tier = tier
			userRepo.users[user.ID] = user
			userRepo.emailIndex[user.Email] = user

			// Calculate quota for tier
			quotaBytes := getQuotaForTier(tier, defaultConfig)

			// Set current usage as percentage of quota (0-100%)
			currentUsage := int64(float64(quotaBytes) * usagePercent / 100.0)
			storageRepo.usage[user.ID] = currentUsage

			// Zero-size file
			fileSize := int64(0)

			err := svc.CheckStorageQuota(ctx, user.ID, fileSize)

			// Should return nil (no error) for zero-size file when not over quota
			if currentUsage <= quotaBytes {
				return err == nil
			}
			// If already over quota, zero-size file should still be allowed
			// (current_usage + 0 > quota only if already over)
			return true
		},
		genValidTierForStorage(),
		gen.Float64Range(0, 100), // Current usage 0-100% of quota
	))

	// Property 3.5: Quota enforcement is consistent across all tiers
	properties.Property("quota enforcement is consistent across tiers", prop.ForAll(
		func(usageBytes int64, fileSize int64) bool {
			ctx := context.Background()

			for _, tier := range domain.ValidTiers() {
				svc, userRepo, storageRepo := newTestStorageService(defaultConfig)

				// Create user with specified tier
				user := domain.NewUser("test@example.com", "hash", "Test User")
				user.Tier = tier
				userRepo.users[user.ID] = user
				userRepo.emailIndex[user.Email] = user

				// Set current usage
				storageRepo.usage[user.ID] = usageBytes

				// Get quota for tier
				quotaBytes := getQuotaForTier(tier, defaultConfig)

				err := svc.CheckStorageQuota(ctx, user.ID, fileSize)

				// Check if result matches expected behavior
				wouldExceed := usageBytes+fileSize > quotaBytes

				if wouldExceed {
					// Should return error
					if err == nil {
						return false
					}
					appErr, ok := err.(*errors.AppError)
					if !ok || appErr.Code != errors.CodeStorageQuotaExceeded {
						return false
					}
				} else {
					// Should not return error
					if err != nil {
						return false
					}
				}
			}
			return true
		},
		gen.Int64Range(0, 5*bytesPerGB), // Usage 0-5GB
		gen.Int64Range(0, 2*bytesPerGB), // File size 0-2GB
	))

	// Property 3.6: Quota check is deterministic
	properties.Property("quota check is deterministic", prop.ForAll(
		func(tier domain.Tier, usageBytes int64, fileSize int64) bool {
			svc, userRepo, storageRepo := newTestStorageService(defaultConfig)
			ctx := context.Background()

			// Create user with specified tier
			user := domain.NewUser("test@example.com", "hash", "Test User")
			user.Tier = tier
			userRepo.users[user.ID] = user
			userRepo.emailIndex[user.Email] = user

			// Set current usage
			storageRepo.usage[user.ID] = usageBytes

			// Check twice
			err1 := svc.CheckStorageQuota(ctx, user.ID, fileSize)
			err2 := svc.CheckStorageQuota(ctx, user.ID, fileSize)

			// Results should be the same
			if err1 == nil && err2 == nil {
				return true
			}
			if err1 != nil && err2 != nil {
				appErr1, ok1 := err1.(*errors.AppError)
				appErr2, ok2 := err2.(*errors.AppError)
				if ok1 && ok2 {
					return appErr1.Code == appErr2.Code
				}
			}
			return false
		},
		genValidTierForStorage(),
		gen.Int64Range(0, 5*bytesPerGB), // Usage 0-5GB
		gen.Int64Range(0, 2*bytesPerGB), // File size 0-2GB
	))

	properties.TestingRun(t)
}

// Helper function to get quota for tier
func getQuotaForTier(tier domain.Tier, cfg config.StorageConfig) int64 {
	const bytesPerGB = int64(1024 * 1024 * 1024)
	switch tier {
	case domain.TierFree:
		return cfg.FreeQuotaGB * bytesPerGB
	case domain.TierPremium:
		return cfg.PremiumQuotaGB * bytesPerGB
	case domain.TierPro:
		return cfg.ProQuotaGB * bytesPerGB
	default:
		return cfg.FreeQuotaGB * bytesPerGB
	}
}

// Helper function to create test storage service with mocks
func newTestStorageService(cfg config.StorageConfig) (StorageService, *mockUserRepo, *mockStorageRepo) {
	userRepo := newMockUserRepo()
	storageRepo := newMockStorageRepo()

	svc := NewStorageService(userRepo, storageRepo, cfg)

	return svc, userRepo, storageRepo
}

// Mock storage repository for testing
type mockStorageRepo struct {
	usage map[uuid.UUID]int64
}

func newMockStorageRepo() *mockStorageRepo {
	return &mockStorageRepo{
		usage: make(map[uuid.UUID]int64),
	}
}

func (m *mockStorageRepo) GetUserStorageUsage(ctx context.Context, userID uuid.UUID) (int64, error) {
	if usage, ok := m.usage[userID]; ok {
		return usage, nil
	}
	return 0, nil
}

// Generator for valid tiers
func genValidTierForStorage() gopter.Gen {
	return gopter.Gen(func(params *gopter.GenParameters) *gopter.GenResult {
		tiers := domain.ValidTiers()
		idx := params.Rng.Intn(len(tiers))
		return gopter.NewGenResult(tiers[idx], gopter.NoShrinker)
	})
}

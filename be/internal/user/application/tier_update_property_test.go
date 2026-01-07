// Package application contains property-based tests for tier update and over-quota downgrade.
package application

import (
	"context"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"

	"ngasihtau/internal/common/config"
	"ngasihtau/internal/common/errors"
	"ngasihtau/internal/user/domain"
)

// **Feature: user-storage-limit, Property 5: Tier Update Applies Immediately**
//
// *For any* user who changes tier, the new quota SHALL be immediately reflected
// in storage info queries.
//
// **Validates: Requirements 6.1, 6.2, 6.3**

// **Feature: user-storage-limit, Property 6: Over-Quota After Downgrade Blocks New Uploads**
//
// *For any* user who downgrades to a tier where `current_usage > new_quota`:
// - Existing files SHALL remain accessible
// - New uploads SHALL be rejected until usage is below quota
//
// **Validates: Requirements 6.4**

func TestProperty_TierUpdateAppliesImmediately(t *testing.T) {
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

	// Property 5.1: After tier update, GetStorageInfo returns the new tier's quota
	properties.Property("tier update immediately reflects new quota in storage info", prop.ForAll(
		func(initialTier, newTier domain.Tier, usagePercent float64) bool {
			svc, userRepo, storageRepo := newTestStorageService(defaultConfig)
			ctx := context.Background()

			// Create user with initial tier
			user := domain.NewUser("test@example.com", "hash", "Test User")
			user.Tier = initialTier
			userRepo.users[user.ID] = user
			userRepo.emailIndex[user.Email] = user

			// Set some storage usage
			initialQuota := getQuotaForTier(initialTier, defaultConfig)
			usedBytes := int64(float64(initialQuota) * usagePercent / 100.0)
			storageRepo.usage[user.ID] = usedBytes

			// Update tier
			err := svc.UpdateTier(ctx, user.ID, newTier)
			if err != nil {
				return false
			}

			// Get storage info - should reflect new tier immediately
			info, err := svc.GetStorageInfo(ctx, user.ID)
			if err != nil {
				return false
			}

			// Verify the new tier is reflected
			expectedQuota := getQuotaForTier(newTier, defaultConfig)
			return info.Tier == newTier && info.QuotaBytes == expectedQuota
		},
		genValidTierForStorage(),
		genValidTierForStorage(),
		gen.Float64Range(0, 50), // Usage 0-50% of initial quota
	))

	// Property 5.2: Tier update is persisted and consistent across multiple queries
	properties.Property("tier update is persisted across multiple queries", prop.ForAll(
		func(initialTier, newTier domain.Tier) bool {
			svc, userRepo, storageRepo := newTestStorageService(defaultConfig)
			ctx := context.Background()

			// Create user with initial tier
			user := domain.NewUser("test@example.com", "hash", "Test User")
			user.Tier = initialTier
			userRepo.users[user.ID] = user
			userRepo.emailIndex[user.Email] = user
			storageRepo.usage[user.ID] = 0

			// Update tier
			err := svc.UpdateTier(ctx, user.ID, newTier)
			if err != nil {
				return false
			}

			// Query multiple times - should be consistent
			info1, err := svc.GetStorageInfo(ctx, user.ID)
			if err != nil {
				return false
			}

			info2, err := svc.GetStorageInfo(ctx, user.ID)
			if err != nil {
				return false
			}

			return info1.Tier == newTier && info2.Tier == newTier &&
				info1.QuotaBytes == info2.QuotaBytes
		},
		genValidTierForStorage(),
		genValidTierForStorage(),
	))

	// Property 5.3: Upgrade from any tier to higher tier increases quota
	properties.Property("upgrade increases quota immediately", prop.ForAll(
		func(usageBytes int64) bool {
			ctx := context.Background()

			// Test free -> premium upgrade
			svc1, userRepo1, storageRepo1 := newTestStorageService(defaultConfig)
			user1 := domain.NewUser("test1@example.com", "hash", "Test User 1")
			user1.Tier = domain.TierFree
			userRepo1.users[user1.ID] = user1
			userRepo1.emailIndex[user1.Email] = user1
			storageRepo1.usage[user1.ID] = usageBytes

			infoBefore1, _ := svc1.GetStorageInfo(ctx, user1.ID)
			svc1.UpdateTier(ctx, user1.ID, domain.TierPremium)
			infoAfter1, _ := svc1.GetStorageInfo(ctx, user1.ID)

			if infoAfter1.QuotaBytes <= infoBefore1.QuotaBytes {
				return false
			}

			// Test premium -> pro upgrade
			svc2, userRepo2, storageRepo2 := newTestStorageService(defaultConfig)
			user2 := domain.NewUser("test2@example.com", "hash", "Test User 2")
			user2.Tier = domain.TierPremium
			userRepo2.users[user2.ID] = user2
			userRepo2.emailIndex[user2.Email] = user2
			storageRepo2.usage[user2.ID] = usageBytes

			infoBefore2, _ := svc2.GetStorageInfo(ctx, user2.ID)
			svc2.UpdateTier(ctx, user2.ID, domain.TierPro)
			infoAfter2, _ := svc2.GetStorageInfo(ctx, user2.ID)

			if infoAfter2.QuotaBytes <= infoBefore2.QuotaBytes {
				return false
			}

			// Test free -> pro upgrade
			svc3, userRepo3, storageRepo3 := newTestStorageService(defaultConfig)
			user3 := domain.NewUser("test3@example.com", "hash", "Test User 3")
			user3.Tier = domain.TierFree
			userRepo3.users[user3.ID] = user3
			userRepo3.emailIndex[user3.Email] = user3
			storageRepo3.usage[user3.ID] = usageBytes

			infoBefore3, _ := svc3.GetStorageInfo(ctx, user3.ID)
			svc3.UpdateTier(ctx, user3.ID, domain.TierPro)
			infoAfter3, _ := svc3.GetStorageInfo(ctx, user3.ID)

			return infoAfter3.QuotaBytes > infoBefore3.QuotaBytes
		},
		gen.Int64Range(0, 500*1024*1024), // 0-500MB usage
	))

	// Property 5.4: Downgrade from any tier to lower tier decreases quota
	properties.Property("downgrade decreases quota immediately", prop.ForAll(
		func(usageBytes int64) bool {
			ctx := context.Background()

			// Test pro -> premium downgrade
			svc1, userRepo1, storageRepo1 := newTestStorageService(defaultConfig)
			user1 := domain.NewUser("test1@example.com", "hash", "Test User 1")
			user1.Tier = domain.TierPro
			userRepo1.users[user1.ID] = user1
			userRepo1.emailIndex[user1.Email] = user1
			storageRepo1.usage[user1.ID] = usageBytes

			infoBefore1, _ := svc1.GetStorageInfo(ctx, user1.ID)
			svc1.UpdateTier(ctx, user1.ID, domain.TierPremium)
			infoAfter1, _ := svc1.GetStorageInfo(ctx, user1.ID)

			if infoAfter1.QuotaBytes >= infoBefore1.QuotaBytes {
				return false
			}

			// Test premium -> free downgrade
			svc2, userRepo2, storageRepo2 := newTestStorageService(defaultConfig)
			user2 := domain.NewUser("test2@example.com", "hash", "Test User 2")
			user2.Tier = domain.TierPremium
			userRepo2.users[user2.ID] = user2
			userRepo2.emailIndex[user2.Email] = user2
			storageRepo2.usage[user2.ID] = usageBytes

			infoBefore2, _ := svc2.GetStorageInfo(ctx, user2.ID)
			svc2.UpdateTier(ctx, user2.ID, domain.TierFree)
			infoAfter2, _ := svc2.GetStorageInfo(ctx, user2.ID)

			if infoAfter2.QuotaBytes >= infoBefore2.QuotaBytes {
				return false
			}

			// Test pro -> free downgrade
			svc3, userRepo3, storageRepo3 := newTestStorageService(defaultConfig)
			user3 := domain.NewUser("test3@example.com", "hash", "Test User 3")
			user3.Tier = domain.TierPro
			userRepo3.users[user3.ID] = user3
			userRepo3.emailIndex[user3.Email] = user3
			storageRepo3.usage[user3.ID] = usageBytes

			infoBefore3, _ := svc3.GetStorageInfo(ctx, user3.ID)
			svc3.UpdateTier(ctx, user3.ID, domain.TierFree)
			infoAfter3, _ := svc3.GetStorageInfo(ctx, user3.ID)

			return infoAfter3.QuotaBytes < infoBefore3.QuotaBytes
		},
		gen.Int64Range(0, 500*1024*1024), // 0-500MB usage
	))

	// Property 5.5: Invalid tier update is rejected
	properties.Property("invalid tier update is rejected", prop.ForAll(
		func(initialTier domain.Tier) bool {
			svc, userRepo, storageRepo := newTestStorageService(defaultConfig)
			ctx := context.Background()

			// Create user with initial tier
			user := domain.NewUser("test@example.com", "hash", "Test User")
			user.Tier = initialTier
			userRepo.users[user.ID] = user
			userRepo.emailIndex[user.Email] = user
			storageRepo.usage[user.ID] = 0

			// Try to update to invalid tier
			err := svc.UpdateTier(ctx, user.ID, domain.Tier("invalid_tier"))

			// Should return error
			if err == nil {
				return false
			}

			// Tier should remain unchanged
			info, _ := svc.GetStorageInfo(ctx, user.ID)
			return info.Tier == initialTier
		},
		genValidTierForStorage(),
	))

	// Property 5.6: Same tier update is idempotent
	properties.Property("same tier update is idempotent", prop.ForAll(
		func(tier domain.Tier, usageBytes int64) bool {
			svc, userRepo, storageRepo := newTestStorageService(defaultConfig)
			ctx := context.Background()

			// Create user with tier
			user := domain.NewUser("test@example.com", "hash", "Test User")
			user.Tier = tier
			userRepo.users[user.ID] = user
			userRepo.emailIndex[user.Email] = user
			storageRepo.usage[user.ID] = usageBytes

			// Get info before
			infoBefore, _ := svc.GetStorageInfo(ctx, user.ID)

			// Update to same tier
			err := svc.UpdateTier(ctx, user.ID, tier)
			if err != nil {
				return false
			}

			// Get info after
			infoAfter, _ := svc.GetStorageInfo(ctx, user.ID)

			// Should be identical
			return infoBefore.Tier == infoAfter.Tier &&
				infoBefore.QuotaBytes == infoAfter.QuotaBytes &&
				infoBefore.UsedBytes == infoAfter.UsedBytes
		},
		genValidTierForStorage(),
		gen.Int64Range(0, 500*1024*1024),
	))

	properties.TestingRun(t)
}

func TestProperty_OverQuotaAfterDowngradeBlocksNewUploads(t *testing.T) {
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

	// Property 6.1: After downgrade with over-quota, new uploads are blocked
	properties.Property("over-quota after downgrade blocks new uploads", prop.ForAll(
		func(overQuotaPercent float64, newFileSize int64) bool {
			svc, userRepo, storageRepo := newTestStorageService(defaultConfig)
			ctx := context.Background()

			// Create user with premium tier
			user := domain.NewUser("test@example.com", "hash", "Test User")
			user.Tier = domain.TierPremium
			userRepo.users[user.ID] = user
			userRepo.emailIndex[user.Email] = user

			// Set usage that's valid for premium but exceeds free quota
			// Free quota is 1GB, premium is 5GB
			// Set usage between 1GB and 5GB (over free quota but under premium)
			freeQuota := getQuotaForTier(domain.TierFree, defaultConfig)
			premiumQuota := getQuotaForTier(domain.TierPremium, defaultConfig)

			// Usage is freeQuota + (overQuotaPercent% of the difference)
			usageBytes := freeQuota + int64(float64(premiumQuota-freeQuota)*overQuotaPercent/100.0)
			storageRepo.usage[user.ID] = usageBytes

			// Verify upload works before downgrade
			errBefore := svc.CheckStorageQuota(ctx, user.ID, newFileSize)
			if usageBytes+newFileSize <= premiumQuota && errBefore != nil {
				return false // Should be allowed before downgrade
			}

			// Downgrade to free tier
			err := svc.UpdateTier(ctx, user.ID, domain.TierFree)
			if err != nil {
				return false
			}

			// Now any new upload should be blocked (since we're over quota)
			errAfter := svc.CheckStorageQuota(ctx, user.ID, newFileSize)

			// Should return storage quota exceeded error
			if errAfter == nil {
				return false
			}

			appErr, ok := errAfter.(*errors.AppError)
			if !ok {
				return false
			}

			return appErr.Code == errors.CodeStorageQuotaExceeded
		},
		gen.Float64Range(10, 90),         // 10-90% over free quota (but under premium)
		gen.Int64Range(1, 100*1024*1024), // 1 byte to 100MB new file
	))

	// Property 6.2: After downgrade with over-quota, existing files remain (usage unchanged)
	properties.Property("existing files remain after downgrade with over-quota", prop.ForAll(
		func(overQuotaPercent float64) bool {
			svc, userRepo, storageRepo := newTestStorageService(defaultConfig)
			ctx := context.Background()

			// Create user with pro tier
			user := domain.NewUser("test@example.com", "hash", "Test User")
			user.Tier = domain.TierPro
			userRepo.users[user.ID] = user
			userRepo.emailIndex[user.Email] = user

			// Set usage that exceeds free quota
			freeQuota := getQuotaForTier(domain.TierFree, defaultConfig)
			proQuota := getQuotaForTier(domain.TierPro, defaultConfig)

			// Usage is freeQuota + (overQuotaPercent% of the difference)
			usageBytes := freeQuota + int64(float64(proQuota-freeQuota)*overQuotaPercent/100.0)
			storageRepo.usage[user.ID] = usageBytes

			// Get info before downgrade
			infoBefore, _ := svc.GetStorageInfo(ctx, user.ID)

			// Downgrade to free tier
			err := svc.UpdateTier(ctx, user.ID, domain.TierFree)
			if err != nil {
				return false
			}

			// Get info after downgrade
			infoAfter, _ := svc.GetStorageInfo(ctx, user.ID)

			// Used bytes should remain the same (existing files not deleted)
			return infoAfter.UsedBytes == infoBefore.UsedBytes &&
				infoAfter.UsedBytes == usageBytes
		},
		gen.Float64Range(10, 90), // 10-90% over free quota
	))

	// Property 6.3: Over-quota state is correctly reflected in storage info
	properties.Property("over-quota state is correctly reflected in storage info", prop.ForAll(
		func(overQuotaPercent float64) bool {
			svc, userRepo, storageRepo := newTestStorageService(defaultConfig)
			ctx := context.Background()

			// Create user with premium tier
			user := domain.NewUser("test@example.com", "hash", "Test User")
			user.Tier = domain.TierPremium
			userRepo.users[user.ID] = user
			userRepo.emailIndex[user.Email] = user

			// Set usage that exceeds free quota
			freeQuota := getQuotaForTier(domain.TierFree, defaultConfig)
			premiumQuota := getQuotaForTier(domain.TierPremium, defaultConfig)

			usageBytes := freeQuota + int64(float64(premiumQuota-freeQuota)*overQuotaPercent/100.0)
			storageRepo.usage[user.ID] = usageBytes

			// Downgrade to free tier
			svc.UpdateTier(ctx, user.ID, domain.TierFree)

			// Get storage info
			info, _ := svc.GetStorageInfo(ctx, user.ID)

			// Should show:
			// - UsedBytes > QuotaBytes (over quota)
			// - RemainingBytes = 0 (no space left)
			// - UsagePercent > 100 (over 100%)
			// - Warning = "critical" (since > 90%)
			return info.UsedBytes > info.QuotaBytes &&
				info.RemainingBytes == 0 &&
				info.UsagePercent > 100 &&
				info.Warning == "critical"
		},
		gen.Float64Range(10, 90), // 10-90% over free quota
	))

	// Property 6.4: Zero-size upload is still blocked when over quota
	properties.Property("zero-size upload blocked when over quota after downgrade", prop.ForAll(
		func(overQuotaPercent float64) bool {
			svc, userRepo, storageRepo := newTestStorageService(defaultConfig)
			ctx := context.Background()

			// Create user with premium tier
			user := domain.NewUser("test@example.com", "hash", "Test User")
			user.Tier = domain.TierPremium
			userRepo.users[user.ID] = user
			userRepo.emailIndex[user.Email] = user

			// Set usage that exceeds free quota
			freeQuota := getQuotaForTier(domain.TierFree, defaultConfig)
			premiumQuota := getQuotaForTier(domain.TierPremium, defaultConfig)

			usageBytes := freeQuota + int64(float64(premiumQuota-freeQuota)*overQuotaPercent/100.0)
			storageRepo.usage[user.ID] = usageBytes

			// Downgrade to free tier
			svc.UpdateTier(ctx, user.ID, domain.TierFree)

			// Try zero-size upload - should still be blocked because current usage > quota
			err := svc.CheckStorageQuota(ctx, user.ID, 0)

			// When over quota, even zero-size should be blocked
			// (current_usage + 0 > quota is true)
			if err == nil {
				return false
			}

			appErr, ok := err.(*errors.AppError)
			if !ok {
				return false
			}

			return appErr.Code == errors.CodeStorageQuotaExceeded
		},
		gen.Float64Range(10, 90), // 10-90% over free quota
	))

	// Property 6.5: Upgrade after over-quota downgrade allows uploads again
	properties.Property("upgrade after over-quota downgrade allows uploads again", prop.ForAll(
		func(overQuotaPercent float64, newFileSize int64) bool {
			svc, userRepo, storageRepo := newTestStorageService(defaultConfig)
			ctx := context.Background()

			// Create user with premium tier
			user := domain.NewUser("test@example.com", "hash", "Test User")
			user.Tier = domain.TierPremium
			userRepo.users[user.ID] = user
			userRepo.emailIndex[user.Email] = user

			// Set usage that exceeds free quota but is under premium
			freeQuota := getQuotaForTier(domain.TierFree, defaultConfig)
			premiumQuota := getQuotaForTier(domain.TierPremium, defaultConfig)

			// Usage between free and premium quotas
			usageBytes := freeQuota + int64(float64(premiumQuota-freeQuota)*overQuotaPercent/100.0)
			storageRepo.usage[user.ID] = usageBytes

			// Downgrade to free tier
			svc.UpdateTier(ctx, user.ID, domain.TierFree)

			// Verify blocked
			errBlocked := svc.CheckStorageQuota(ctx, user.ID, newFileSize)
			if errBlocked == nil {
				return false // Should be blocked
			}

			// Upgrade back to premium
			svc.UpdateTier(ctx, user.ID, domain.TierPremium)

			// Now should be allowed (if file fits in remaining space)
			errAfterUpgrade := svc.CheckStorageQuota(ctx, user.ID, newFileSize)

			// Should be allowed if usage + newFileSize <= premiumQuota
			if usageBytes+newFileSize <= premiumQuota {
				return errAfterUpgrade == nil
			}
			// Otherwise should still be blocked
			return errAfterUpgrade != nil
		},
		gen.Float64Range(10, 50),         // 10-50% over free quota (leaves room for new files in premium)
		gen.Int64Range(1, 100*1024*1024), // 1 byte to 100MB new file
	))

	// Property 6.6: Pro to free downgrade with large usage blocks uploads
	properties.Property("pro to free downgrade with large usage blocks uploads", prop.ForAll(
		func(usageGB float64, newFileMB int64) bool {
			svc, userRepo, storageRepo := newTestStorageService(defaultConfig)
			ctx := context.Background()

			// Create user with pro tier
			user := domain.NewUser("test@example.com", "hash", "Test User")
			user.Tier = domain.TierPro
			userRepo.users[user.ID] = user
			userRepo.emailIndex[user.Email] = user

			// Set usage between 1GB (free quota) and 20GB (pro quota)
			// usageGB is 1.5 to 15 GB
			usageBytes := int64(usageGB * float64(bytesPerGB))
			storageRepo.usage[user.ID] = usageBytes

			// Downgrade to free tier
			svc.UpdateTier(ctx, user.ID, domain.TierFree)

			// Try to upload
			newFileSize := newFileMB * 1024 * 1024
			err := svc.CheckStorageQuota(ctx, user.ID, newFileSize)

			// Should be blocked since usage > free quota (1GB)
			freeQuota := getQuotaForTier(domain.TierFree, defaultConfig)
			if usageBytes > freeQuota {
				if err == nil {
					return false
				}
				appErr, ok := err.(*errors.AppError)
				return ok && appErr.Code == errors.CodeStorageQuotaExceeded
			}
			return true
		},
		gen.Float64Range(1.5, 15), // 1.5GB to 15GB usage
		gen.Int64Range(1, 100),    // 1MB to 100MB new file
	))

	properties.TestingRun(t)
}

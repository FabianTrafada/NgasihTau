// Package domain contains property-based tests for storage info calculation.
package domain

import (
	"math"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"

	"ngasihtau/internal/common/config"
)

// **Feature: user-storage-limit, Property 2: Storage Info Calculation Accuracy**
//
// *For any* user with a set of materials, the storage info SHALL satisfy:
// - `used_bytes` equals the sum of all non-deleted material `file_size` values
// - `remaining_bytes` equals `quota_bytes - used_bytes`
// - `usage_percent` equals `(used_bytes / quota_bytes) * 100`
//
// **Validates: Requirements 2.1, 2.3, 2.4, 5.1**

// CalculateStorageInfo calculates storage information for a user based on their
// used bytes, tier, and configuration. This is the function under test.
func CalculateStorageInfo(usedBytes int64, tier Tier, cfg *config.StorageConfig) *StorageInfo {
	quotaBytes := GetQuotaBytesForTier(tier, cfg)

	remainingBytes := quotaBytes - usedBytes
	if remainingBytes < 0 {
		remainingBytes = 0
	}

	var usagePercent float64
	if quotaBytes > 0 {
		usagePercent = (float64(usedBytes) / float64(quotaBytes)) * 100
	}

	// Determine warning level
	var warning string
	if usagePercent >= 90 {
		warning = "critical"
	} else if usagePercent >= 80 {
		warning = "warning"
	}

	// Determine next tier upgrade option
	var nextTier *Tier
	var nextTierQuota *int64
	switch tier {
	case TierFree:
		nt := TierPremium
		nextTier = &nt
		nq := GetQuotaBytesForTier(TierPremium, cfg)
		nextTierQuota = &nq
	case TierPremium:
		nt := TierPro
		nextTier = &nt
		nq := GetQuotaBytesForTier(TierPro, cfg)
		nextTierQuota = &nq
	case TierPro:
		// Pro is the highest tier, no upgrade available
		nextTier = nil
		nextTierQuota = nil
	}

	return &StorageInfo{
		UsedBytes:      usedBytes,
		QuotaBytes:     quotaBytes,
		RemainingBytes: remainingBytes,
		UsagePercent:   usagePercent,
		Tier:           tier,
		Warning:        warning,
		NextTier:       nextTier,
		NextTierQuota:  nextTierQuota,
	}
}

// SumMaterialSizes calculates the total size of non-deleted materials.
// This simulates what the repository does when calculating storage usage.
func SumMaterialSizes(materials []MaterialForTest) int64 {
	var total int64
	for _, m := range materials {
		if !m.IsDeleted {
			total += m.FileSize
		}
	}
	return total
}

// MaterialForTest represents a material with file size and deletion status for testing.
type MaterialForTest struct {
	FileSize  int64
	IsDeleted bool
}

func TestProperty_StorageInfoCalculationAccuracy(t *testing.T) {
	parameters := gopter.DefaultTestParametersWithSeed(12345)
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	// Default configuration
	defaultConfig := &config.StorageConfig{
		FreeQuotaGB:    1,
		PremiumQuotaGB: 5,
		ProQuotaGB:     20,
	}

	const bytesPerGB = int64(1024 * 1024 * 1024)

	// Property 2.1: used_bytes equals sum of all non-deleted material file_size values
	properties.Property("used_bytes equals sum of non-deleted materials", prop.ForAll(
		func(materials []MaterialForTest, tier Tier) bool {
			expectedUsedBytes := SumMaterialSizes(materials)
			info := CalculateStorageInfo(expectedUsedBytes, tier, defaultConfig)
			return info.UsedBytes == expectedUsedBytes
		},
		genMaterialList(),
		genValidTier(),
	))

	// Property 2.2: remaining_bytes equals quota_bytes - used_bytes (non-negative)
	properties.Property("remaining_bytes equals quota_bytes minus used_bytes", prop.ForAll(
		func(usedBytes int64, tier Tier) bool {
			info := CalculateStorageInfo(usedBytes, tier, defaultConfig)
			quotaBytes := GetQuotaBytesForTier(tier, defaultConfig)

			expectedRemaining := quotaBytes - usedBytes
			if expectedRemaining < 0 {
				expectedRemaining = 0
			}

			return info.RemainingBytes == expectedRemaining
		},
		genStorageBytes(),
		genValidTier(),
	))

	// Property 2.3: usage_percent equals (used_bytes / quota_bytes) * 100
	properties.Property("usage_percent is calculated correctly", prop.ForAll(
		func(usedBytes int64, tier Tier) bool {
			info := CalculateStorageInfo(usedBytes, tier, defaultConfig)
			quotaBytes := GetQuotaBytesForTier(tier, defaultConfig)

			expectedPercent := (float64(usedBytes) / float64(quotaBytes)) * 100

			// Use approximate comparison for floating point
			return math.Abs(info.UsagePercent-expectedPercent) < 0.0001
		},
		genStorageBytes(),
		genValidTier(),
	))

	// Property 2.4: quota_bytes matches tier configuration
	properties.Property("quota_bytes matches tier configuration", prop.ForAll(
		func(usedBytes int64, tier Tier) bool {
			info := CalculateStorageInfo(usedBytes, tier, defaultConfig)
			expectedQuota := GetQuotaBytesForTier(tier, defaultConfig)
			return info.QuotaBytes == expectedQuota
		},
		genStorageBytes(),
		genValidTier(),
	))

	// Property 2.5: tier in storage info matches input tier
	properties.Property("tier in storage info matches input tier", prop.ForAll(
		func(usedBytes int64, tier Tier) bool {
			info := CalculateStorageInfo(usedBytes, tier, defaultConfig)
			return info.Tier == tier
		},
		genStorageBytes(),
		genValidTier(),
	))

	// Property 2.6: remaining_bytes is never negative
	properties.Property("remaining_bytes is never negative", prop.ForAll(
		func(usedBytes int64, tier Tier) bool {
			info := CalculateStorageInfo(usedBytes, tier, defaultConfig)
			return info.RemainingBytes >= 0
		},
		genStorageBytes(),
		genValidTier(),
	))

	// Property 2.7: when used_bytes exceeds quota, remaining_bytes is zero
	properties.Property("when over quota, remaining_bytes is zero", prop.ForAll(
		func(tier Tier) bool {
			quotaBytes := GetQuotaBytesForTier(tier, defaultConfig)
			// Use 150% of quota to ensure we're over
			overQuotaBytes := quotaBytes + (quotaBytes / 2)
			info := CalculateStorageInfo(overQuotaBytes, tier, defaultConfig)
			return info.RemainingBytes == 0
		},
		genValidTier(),
	))

	// Property 2.8: storage info calculation is deterministic
	properties.Property("storage info calculation is deterministic", prop.ForAll(
		func(usedBytes int64, tier Tier) bool {
			info1 := CalculateStorageInfo(usedBytes, tier, defaultConfig)
			info2 := CalculateStorageInfo(usedBytes, tier, defaultConfig)

			return info1.UsedBytes == info2.UsedBytes &&
				info1.QuotaBytes == info2.QuotaBytes &&
				info1.RemainingBytes == info2.RemainingBytes &&
				info1.UsagePercent == info2.UsagePercent &&
				info1.Tier == info2.Tier
		},
		genStorageBytes(),
		genValidTier(),
	))

	// Property 2.9: deleted materials are not counted in used_bytes
	properties.Property("deleted materials are not counted", prop.ForAll(
		func(materials []MaterialForTest) bool {
			// Calculate expected: only non-deleted materials
			var expectedUsed int64
			for _, m := range materials {
				if !m.IsDeleted {
					expectedUsed += m.FileSize
				}
			}

			actualUsed := SumMaterialSizes(materials)
			return actualUsed == expectedUsed
		},
		genMaterialListWithDeletions(),
	))

	// Property 2.10: usage_percent is 0 when used_bytes is 0
	properties.Property("usage_percent is 0 when used_bytes is 0", prop.ForAll(
		func(tier Tier) bool {
			info := CalculateStorageInfo(0, tier, defaultConfig)
			return info.UsagePercent == 0
		},
		genValidTier(),
	))

	// Property 2.11: usage_percent can exceed 100 when over quota
	properties.Property("usage_percent can exceed 100 when over quota", prop.ForAll(
		func(tier Tier) bool {
			quotaBytes := GetQuotaBytesForTier(tier, defaultConfig)
			// Use 200% of quota
			overQuotaBytes := quotaBytes * 2
			info := CalculateStorageInfo(overQuotaBytes, tier, defaultConfig)
			return info.UsagePercent >= 200
		},
		genValidTier(),
	))

	properties.TestingRun(t)
}

// **Feature: user-storage-limit, Property 7: Warning Thresholds**
//
// *For any* user's storage info:
// - IF `usage_percent >= 90` THEN `warning` SHALL be "critical"
// - ELSE IF `usage_percent >= 80` THEN `warning` SHALL be "warning"
// - ELSE `warning` SHALL be empty
//
// **Validates: Requirements 7.1, 7.2**
func TestProperty_WarningThresholds(t *testing.T) {
	parameters := gopter.DefaultTestParametersWithSeed(12345)
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	// Default configuration
	defaultConfig := &config.StorageConfig{
		FreeQuotaGB:    1,
		PremiumQuotaGB: 5,
		ProQuotaGB:     20,
	}

	// Property 7.1: When usage_percent >= 90, warning is "critical"
	properties.Property("critical warning when usage >= 90%", prop.ForAll(
		func(tier Tier, percentOver90 float64) bool {
			quotaBytes := GetQuotaBytesForTier(tier, defaultConfig)
			// Calculate used bytes to achieve 90% + percentOver90 (0-10%)
			usagePercent := 90.0 + percentOver90
			usedBytes := int64(float64(quotaBytes) * usagePercent / 100.0)

			info := CalculateStorageInfo(usedBytes, tier, defaultConfig)
			return info.Warning == "critical"
		},
		genValidTier(),
		gen.Float64Range(0, 110), // 90% to 200% usage
	))

	// Property 7.2: When usage_percent >= 80 and < 90, warning is "warning"
	properties.Property("warning when usage >= 80% and < 90%", prop.ForAll(
		func(tier Tier, percentInRange float64) bool {
			quotaBytes := GetQuotaBytesForTier(tier, defaultConfig)
			// Calculate used bytes to achieve 80% + percentInRange (0-9.99%)
			usagePercent := 80.0 + percentInRange
			usedBytes := int64(float64(quotaBytes) * usagePercent / 100.0)

			info := CalculateStorageInfo(usedBytes, tier, defaultConfig)

			// Recalculate actual usage percent to handle rounding
			actualPercent := info.UsagePercent
			if actualPercent >= 90 {
				// Due to rounding, we might have crossed into critical territory
				return info.Warning == "critical"
			}
			return info.Warning == "warning"
		},
		genValidTier(),
		gen.Float64Range(0, 9.99), // 80% to 89.99% usage
	))

	// Property 7.3: When usage_percent < 80, warning is empty
	properties.Property("no warning when usage < 80%", prop.ForAll(
		func(tier Tier, percentUnder80 float64) bool {
			quotaBytes := GetQuotaBytesForTier(tier, defaultConfig)
			// Calculate used bytes to achieve percentUnder80 (0-79.99%)
			usedBytes := int64(float64(quotaBytes) * percentUnder80 / 100.0)

			info := CalculateStorageInfo(usedBytes, tier, defaultConfig)
			return info.Warning == ""
		},
		genValidTier(),
		gen.Float64Range(0, 79.99), // 0% to 79.99% usage
	))

	// Property 7.4: Warning thresholds are consistent across all tiers
	properties.Property("warning thresholds consistent across tiers", prop.ForAll(
		func(usagePercent float64) bool {
			// Test all tiers with the same usage percentage
			for _, tier := range ValidTiers() {
				quotaBytes := GetQuotaBytesForTier(tier, defaultConfig)
				usedBytes := int64(float64(quotaBytes) * usagePercent / 100.0)
				info := CalculateStorageInfo(usedBytes, tier, defaultConfig)

				// Recalculate actual percent due to int64 rounding
				actualPercent := info.UsagePercent

				var expectedWarning string
				if actualPercent >= 90 {
					expectedWarning = "critical"
				} else if actualPercent >= 80 {
					expectedWarning = "warning"
				} else {
					expectedWarning = ""
				}

				if info.Warning != expectedWarning {
					return false
				}
			}
			return true
		},
		gen.Float64Range(0, 150), // 0% to 150% usage
	))

	// Property 7.5: Warning at or just above 80% threshold
	// Note: Due to floating-point to int64 conversion, exact percentages may not be achievable.
	// We test that values calculated to be at 80% result in the correct warning based on actual percentage.
	properties.Property("warning at 80% threshold respects actual percentage", prop.ForAll(
		func(tier Tier) bool {
			quotaBytes := GetQuotaBytesForTier(tier, defaultConfig)
			// Calculate bytes for 80% - may be slightly under due to truncation
			usedBytes := int64(float64(quotaBytes) * 80.0 / 100.0)
			info := CalculateStorageInfo(usedBytes, tier, defaultConfig)

			// The warning should match the actual calculated percentage
			if info.UsagePercent >= 90 {
				return info.Warning == "critical"
			} else if info.UsagePercent >= 80 {
				return info.Warning == "warning"
			}
			return info.Warning == ""
		},
		genValidTier(),
	))

	// Property 7.6: Critical warning at or just above 90% threshold
	properties.Property("critical at 90% threshold respects actual percentage", prop.ForAll(
		func(tier Tier) bool {
			quotaBytes := GetQuotaBytesForTier(tier, defaultConfig)
			// Calculate bytes for 90% - may be slightly under due to truncation
			usedBytes := int64(float64(quotaBytes) * 90.0 / 100.0)
			info := CalculateStorageInfo(usedBytes, tier, defaultConfig)

			// The warning should match the actual calculated percentage
			if info.UsagePercent >= 90 {
				return info.Warning == "critical"
			} else if info.UsagePercent >= 80 {
				return info.Warning == "warning"
			}
			return info.Warning == ""
		},
		genValidTier(),
	))

	// Property 7.7: Warning is deterministic for same input
	properties.Property("warning calculation is deterministic", prop.ForAll(
		func(usedBytes int64, tier Tier) bool {
			info1 := CalculateStorageInfo(usedBytes, tier, defaultConfig)
			info2 := CalculateStorageInfo(usedBytes, tier, defaultConfig)
			return info1.Warning == info2.Warning
		},
		genStorageBytes(),
		genValidTier(),
	))

	// Property 7.8: Warning only has three possible values
	properties.Property("warning has only valid values", prop.ForAll(
		func(usedBytes int64, tier Tier) bool {
			info := CalculateStorageInfo(usedBytes, tier, defaultConfig)
			return info.Warning == "" || info.Warning == "warning" || info.Warning == "critical"
		},
		genStorageBytes(),
		genValidTier(),
	))

	properties.TestingRun(t)
}

// **Feature: user-storage-limit, Property 8: Next Tier Upgrade Option**
//
// *For any* user's storage info:
// - IF tier is "free" THEN `next_tier` SHALL be "premium"
// - IF tier is "premium" THEN `next_tier` SHALL be "pro"
// - IF tier is "pro" THEN `next_tier` SHALL be nil
//
// **Validates: Requirements 7.3**
func TestProperty_NextTierUpgradeOption(t *testing.T) {
	parameters := gopter.DefaultTestParametersWithSeed(12345)
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	// Default configuration
	defaultConfig := &config.StorageConfig{
		FreeQuotaGB:    1,
		PremiumQuotaGB: 5,
		ProQuotaGB:     20,
	}

	// Property 8.1: Free tier has premium as next tier
	properties.Property("free tier has premium as next tier", prop.ForAll(
		func(usedBytes int64) bool {
			info := CalculateStorageInfo(usedBytes, TierFree, defaultConfig)
			return info.NextTier != nil && *info.NextTier == TierPremium
		},
		genStorageBytes(),
	))

	// Property 8.2: Premium tier has pro as next tier
	properties.Property("premium tier has pro as next tier", prop.ForAll(
		func(usedBytes int64) bool {
			info := CalculateStorageInfo(usedBytes, TierPremium, defaultConfig)
			return info.NextTier != nil && *info.NextTier == TierPro
		},
		genStorageBytes(),
	))

	// Property 8.3: Pro tier has no next tier (nil)
	properties.Property("pro tier has no next tier", prop.ForAll(
		func(usedBytes int64) bool {
			info := CalculateStorageInfo(usedBytes, TierPro, defaultConfig)
			return info.NextTier == nil
		},
		genStorageBytes(),
	))

	// Property 8.4: Free tier next tier quota matches premium quota
	properties.Property("free tier next tier quota matches premium quota", prop.ForAll(
		func(usedBytes int64, premiumGB int64) bool {
			cfg := &config.StorageConfig{
				FreeQuotaGB:    1,
				PremiumQuotaGB: premiumGB,
				ProQuotaGB:     20,
			}
			info := CalculateStorageInfo(usedBytes, TierFree, cfg)
			expectedQuota := GetQuotaBytesForTier(TierPremium, cfg)
			return info.NextTierQuota != nil && *info.NextTierQuota == expectedQuota
		},
		genStorageBytes(),
		genPositiveInt64(),
	))

	// Property 8.5: Premium tier next tier quota matches pro quota
	properties.Property("premium tier next tier quota matches pro quota", prop.ForAll(
		func(usedBytes int64, proGB int64) bool {
			cfg := &config.StorageConfig{
				FreeQuotaGB:    1,
				PremiumQuotaGB: 5,
				ProQuotaGB:     proGB,
			}
			info := CalculateStorageInfo(usedBytes, TierPremium, cfg)
			expectedQuota := GetQuotaBytesForTier(TierPro, cfg)
			return info.NextTierQuota != nil && *info.NextTierQuota == expectedQuota
		},
		genStorageBytes(),
		genPositiveInt64(),
	))

	// Property 8.6: Pro tier has no next tier quota (nil)
	properties.Property("pro tier has no next tier quota", prop.ForAll(
		func(usedBytes int64) bool {
			info := CalculateStorageInfo(usedBytes, TierPro, defaultConfig)
			return info.NextTierQuota == nil
		},
		genStorageBytes(),
	))

	// Property 8.7: Next tier is always higher than current tier (when not nil)
	properties.Property("next tier is always higher than current tier", prop.ForAll(
		func(usedBytes int64, tier Tier) bool {
			info := CalculateStorageInfo(usedBytes, tier, defaultConfig)
			if info.NextTier == nil {
				// Pro tier has no next tier, which is valid
				return tier == TierPro
			}
			// Next tier should have higher order than current tier
			return info.NextTier.Order() > tier.Order()
		},
		genStorageBytes(),
		genValidTier(),
	))

	// Property 8.8: Next tier quota is always greater than current quota (when not nil)
	properties.Property("next tier quota is always greater than current quota", prop.ForAll(
		func(usedBytes int64, tier Tier) bool {
			info := CalculateStorageInfo(usedBytes, tier, defaultConfig)
			if info.NextTierQuota == nil {
				// Pro tier has no next tier quota, which is valid
				return tier == TierPro
			}
			// Next tier quota should be greater than current quota
			return *info.NextTierQuota > info.QuotaBytes
		},
		genStorageBytes(),
		genValidTier(),
	))

	// Property 8.9: Next tier and next tier quota are both nil or both non-nil
	properties.Property("next tier and quota are consistent (both nil or both non-nil)", prop.ForAll(
		func(usedBytes int64, tier Tier) bool {
			info := CalculateStorageInfo(usedBytes, tier, defaultConfig)
			bothNil := info.NextTier == nil && info.NextTierQuota == nil
			bothNonNil := info.NextTier != nil && info.NextTierQuota != nil
			return bothNil || bothNonNil
		},
		genStorageBytes(),
		genValidTier(),
	))

	// Property 8.10: Next tier calculation is deterministic
	properties.Property("next tier calculation is deterministic", prop.ForAll(
		func(usedBytes int64, tier Tier) bool {
			info1 := CalculateStorageInfo(usedBytes, tier, defaultConfig)
			info2 := CalculateStorageInfo(usedBytes, tier, defaultConfig)

			// Compare NextTier
			if info1.NextTier == nil && info2.NextTier == nil {
				// Both nil, check NextTierQuota
			} else if info1.NextTier != nil && info2.NextTier != nil {
				if *info1.NextTier != *info2.NextTier {
					return false
				}
			} else {
				return false // One nil, one not
			}

			// Compare NextTierQuota
			if info1.NextTierQuota == nil && info2.NextTierQuota == nil {
				return true
			} else if info1.NextTierQuota != nil && info2.NextTierQuota != nil {
				return *info1.NextTierQuota == *info2.NextTierQuota
			}
			return false // One nil, one not
		},
		genStorageBytes(),
		genValidTier(),
	))

	properties.TestingRun(t)
}

// Generator for storage bytes (0 to 50GB range)
func genStorageBytes() gopter.Gen {
	const maxBytes = int64(50 * 1024 * 1024 * 1024) // 50GB
	return gen.Int64Range(0, maxBytes)
}

// Generator for a list of materials
func genMaterialList() gopter.Gen {
	return gen.SliceOf(genMaterial())
}

// Generator for a list of materials with some deletions
func genMaterialListWithDeletions() gopter.Gen {
	return gen.SliceOf(genMaterialWithDeletion())
}

// Generator for a single material (non-deleted)
func genMaterial() gopter.Gen {
	return gen.Int64Range(0, 100*1024*1024).Map(func(size int64) MaterialForTest {
		return MaterialForTest{
			FileSize:  size,
			IsDeleted: false,
		}
	})
}

// Generator for a single material (may be deleted)
func genMaterialWithDeletion() gopter.Gen {
	return gopter.CombineGens(
		gen.Int64Range(0, 100*1024*1024), // 0 to 100MB
		gen.Bool(),
	).Map(func(vals []interface{}) MaterialForTest {
		return MaterialForTest{
			FileSize:  vals[0].(int64),
			IsDeleted: vals[1].(bool),
		}
	})
}

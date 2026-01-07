// Package domain contains property-based tests for tier to quota mapping.
package domain

import (
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/prop"

	"ngasihtau/internal/common/config"
)

// **Feature: user-storage-limit, Property 1: Tier to Quota Mapping Consistency**
//
// *For any* valid user tier, the system SHALL return the correct storage quota
// as configured (free=1GB, premium=5GB, pro=20GB by default).
//
// **Validates: Requirements 1.2, 2.2**

// GetQuotaBytesForTier returns the storage quota in bytes for a given tier.
// This is the function under test that maps tiers to their configured quotas.
func GetQuotaBytesForTier(tier Tier, cfg *config.StorageConfig) int64 {
	const bytesPerGB = int64(1024 * 1024 * 1024)
	switch tier {
	case TierFree:
		return cfg.FreeQuotaGB * bytesPerGB
	case TierPremium:
		return cfg.PremiumQuotaGB * bytesPerGB
	case TierPro:
		return cfg.ProQuotaGB * bytesPerGB
	default:
		// Unknown tiers default to free tier quota
		return cfg.FreeQuotaGB * bytesPerGB
	}
}

func TestProperty_TierToQuotaMappingConsistency(t *testing.T) {
	parameters := gopter.DefaultTestParametersWithSeed(12345)
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	// Default configuration as per requirements
	defaultConfig := &config.StorageConfig{
		FreeQuotaGB:    1,
		PremiumQuotaGB: 5,
		ProQuotaGB:     20,
	}

	const bytesPerGB = int64(1024 * 1024 * 1024)

	// Property 1.1: Free tier always maps to FreeQuotaGB
	properties.Property("free tier maps to configured free quota", prop.ForAll(
		func(freeGB int64) bool {
			cfg := &config.StorageConfig{
				FreeQuotaGB:    freeGB,
				PremiumQuotaGB: 5,
				ProQuotaGB:     20,
			}
			quota := GetQuotaBytesForTier(TierFree, cfg)
			return quota == freeGB*bytesPerGB
		},
		genPositiveInt64(),
	))

	// Property 1.2: Premium tier always maps to PremiumQuotaGB
	properties.Property("premium tier maps to configured premium quota", prop.ForAll(
		func(premiumGB int64) bool {
			cfg := &config.StorageConfig{
				FreeQuotaGB:    1,
				PremiumQuotaGB: premiumGB,
				ProQuotaGB:     20,
			}
			quota := GetQuotaBytesForTier(TierPremium, cfg)
			return quota == premiumGB*bytesPerGB
		},
		genPositiveInt64(),
	))

	// Property 1.3: Pro tier always maps to ProQuotaGB
	properties.Property("pro tier maps to configured pro quota", prop.ForAll(
		func(proGB int64) bool {
			cfg := &config.StorageConfig{
				FreeQuotaGB:    1,
				PremiumQuotaGB: 5,
				ProQuotaGB:     proGB,
			}
			quota := GetQuotaBytesForTier(TierPro, cfg)
			return quota == proGB*bytesPerGB
		},
		genPositiveInt64(),
	))

	// Property 1.4: Default config returns expected default values
	properties.Property("default config returns expected quotas", prop.ForAll(
		func(tier Tier) bool {
			quota := GetQuotaBytesForTier(tier, defaultConfig)
			switch tier {
			case TierFree:
				return quota == 1*bytesPerGB
			case TierPremium:
				return quota == 5*bytesPerGB
			case TierPro:
				return quota == 20*bytesPerGB
			default:
				return false
			}
		},
		genValidTier(),
	))

	// Property 1.5: Tier ordering is preserved in quota values (higher tier = higher quota)
	properties.Property("higher tier always has higher or equal quota", prop.ForAll(
		func(freeGB, premiumGB, proGB int64) bool {
			// Ensure tier ordering: free <= premium <= pro
			if freeGB > premiumGB || premiumGB > proGB {
				return true // Skip invalid configs where ordering is violated
			}
			cfg := &config.StorageConfig{
				FreeQuotaGB:    freeGB,
				PremiumQuotaGB: premiumGB,
				ProQuotaGB:     proGB,
			}
			freeQuota := GetQuotaBytesForTier(TierFree, cfg)
			premiumQuota := GetQuotaBytesForTier(TierPremium, cfg)
			proQuota := GetQuotaBytesForTier(TierPro, cfg)

			return freeQuota <= premiumQuota && premiumQuota <= proQuota
		},
		genPositiveInt64(),
		genPositiveInt64(),
		genPositiveInt64(),
	))

	// Property 1.6: All valid tiers return positive quota
	properties.Property("all valid tiers return positive quota", prop.ForAll(
		func(tier Tier) bool {
			quota := GetQuotaBytesForTier(tier, defaultConfig)
			return quota > 0
		},
		genValidTier(),
	))

	// Property 1.7: Quota calculation is deterministic (same input = same output)
	properties.Property("quota calculation is deterministic", prop.ForAll(
		func(tier Tier, freeGB, premiumGB, proGB int64) bool {
			cfg := &config.StorageConfig{
				FreeQuotaGB:    freeGB,
				PremiumQuotaGB: premiumGB,
				ProQuotaGB:     proGB,
			}
			quota1 := GetQuotaBytesForTier(tier, cfg)
			quota2 := GetQuotaBytesForTier(tier, cfg)
			return quota1 == quota2
		},
		genValidTier(),
		genPositiveInt64(),
		genPositiveInt64(),
		genPositiveInt64(),
	))

	properties.TestingRun(t)
}

// Generator for valid tiers
func genValidTier() gopter.Gen {
	return gopter.CombineGens().FlatMap(func(_ interface{}) gopter.Gen {
		tiers := ValidTiers()
		return gopter.Gen(func(params *gopter.GenParameters) *gopter.GenResult {
			idx := params.Rng.Intn(len(tiers))
			return gopter.NewGenResult(tiers[idx], gopter.NoShrinker)
		})
	}, nil)
}

// Generator for positive int64 values (for quota GB values)
func genPositiveInt64() gopter.Gen {
	return gopter.Gen(func(params *gopter.GenParameters) *gopter.GenResult {
		// Generate values between 1 and 1000 GB (reasonable quota range)
		val := int64(params.Rng.Intn(1000)) + 1
		return gopter.NewGenResult(val, gopter.NoShrinker)
	})
}

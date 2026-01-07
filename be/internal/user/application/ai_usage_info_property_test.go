// Package application contains property-based tests for AI usage info calculation.
package application

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"

	"ngasihtau/internal/common/config"
	"ngasihtau/internal/user/domain"
)

// **Feature: user-storage-limit, Property 10: AI Usage Info Calculation**
//
// *For any* user's AI usage info:
// - `remaining` equals `daily_limit - used_today` (or -1 if unlimited)
// - `is_unlimited` is true IFF tier is "pro"
//
// **Validates: Requirements 10.1, 10.2, 10.3**

func TestProperty_AIUsageInfoCalculation(t *testing.T) {
	parameters := gopter.DefaultTestParametersWithSeed(12345)
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	// Default configuration
	defaultConfig := config.AILimitConfig{
		FreeDailyLimit:    20,
		PremiumDailyLimit: 100,
		ProDailyLimit:     -1, // Unlimited
	}

	// Property 10.1: Remaining equals daily_limit - used_today for non-pro tiers
	// Validates: Requirement 10.3 - WHEN a user requests their AI usage
	// THEN THE Storage_Limit_System SHALL return the remaining messages for today
	properties.Property("remaining equals daily_limit minus used_today for non-pro tiers", prop.ForAll(
		func(tier domain.Tier, usedToday int) bool {
			// Skip pro tier for this property (handled separately)
			if tier == domain.TierPro {
				return true
			}

			svc, userRepo, aiUsageRepo := newTestAIService(defaultConfig)
			ctx := context.Background()

			// Create user with specified tier
			user := domain.NewUser("test@example.com", "hash", "Test User")
			user.Tier = tier
			userRepo.users[user.ID] = user
			userRepo.emailIndex[user.Email] = user

			// Set daily usage
			aiUsageRepo.usage[user.ID] = usedToday

			// Get AI usage info
			info, err := svc.GetAIUsageInfo(ctx, user.ID)
			if err != nil {
				return false
			}

			// Get expected daily limit
			expectedLimit := getDailyLimitForTier(tier, defaultConfig)

			// Calculate expected remaining
			expectedRemaining := expectedLimit - usedToday
			if expectedRemaining < 0 {
				expectedRemaining = 0
			}

			return info.Remaining == expectedRemaining &&
				info.DailyLimit == expectedLimit &&
				info.UsedToday == usedToday
		},
		genNonProTier(),
		gen.IntRange(0, 150), // Usage 0-150 messages
	))

	// Property 10.2: Pro tier has unlimited remaining (-1)
	// Validates: Requirement 10.2 - WHEN a user requests their AI usage
	// THEN THE Storage_Limit_System SHALL return the daily limit based on user's tier
	properties.Property("pro tier has unlimited remaining", prop.ForAll(
		func(usedToday int) bool {
			svc, userRepo, aiUsageRepo := newTestAIService(defaultConfig)
			ctx := context.Background()

			// Create pro user
			user := domain.NewUser("test@example.com", "hash", "Test User")
			user.Tier = domain.TierPro
			userRepo.users[user.ID] = user
			userRepo.emailIndex[user.Email] = user

			// Set daily usage (shouldn't matter for pro)
			aiUsageRepo.usage[user.ID] = usedToday

			// Get AI usage info
			info, err := svc.GetAIUsageInfo(ctx, user.ID)
			if err != nil {
				return false
			}

			// Pro tier should have remaining = -1 (unlimited)
			return info.Remaining == -1 &&
				info.DailyLimit == -1 &&
				info.UsedToday == usedToday
		},
		gen.IntRange(0, 1000), // Usage 0-1000 messages (shouldn't matter)
	))

	// Property 10.3: IsUnlimited is true IFF tier is pro
	// Validates: Requirement 10.2 - daily limit based on user's tier
	properties.Property("is_unlimited is true iff tier is pro", prop.ForAll(
		func(tier domain.Tier, usedToday int) bool {
			svc, userRepo, aiUsageRepo := newTestAIService(defaultConfig)
			ctx := context.Background()

			// Create user with specified tier
			user := domain.NewUser("test@example.com", "hash", "Test User")
			user.Tier = tier
			userRepo.users[user.ID] = user
			userRepo.emailIndex[user.Email] = user

			// Set daily usage
			aiUsageRepo.usage[user.ID] = usedToday

			// Get AI usage info
			info, err := svc.GetAIUsageInfo(ctx, user.ID)
			if err != nil {
				return false
			}

			// IsUnlimited should be true only for pro tier
			expectedUnlimited := tier == domain.TierPro
			return info.IsUnlimited == expectedUnlimited
		},
		genValidTierForAI(),
		gen.IntRange(0, 100),
	))

	// Property 10.4: Tier is correctly reflected in AI usage info
	// Validates: Requirement 10.1 - WHEN a user requests their AI usage
	// THEN THE Storage_Limit_System SHALL return the current daily message count
	properties.Property("tier is correctly reflected in ai usage info", prop.ForAll(
		func(tier domain.Tier, usedToday int) bool {
			svc, userRepo, aiUsageRepo := newTestAIService(defaultConfig)
			ctx := context.Background()

			// Create user with specified tier
			user := domain.NewUser("test@example.com", "hash", "Test User")
			user.Tier = tier
			userRepo.users[user.ID] = user
			userRepo.emailIndex[user.Email] = user

			// Set daily usage
			aiUsageRepo.usage[user.ID] = usedToday

			// Get AI usage info
			info, err := svc.GetAIUsageInfo(ctx, user.ID)
			if err != nil {
				return false
			}

			return info.Tier == tier
		},
		genValidTierForAI(),
		gen.IntRange(0, 100),
	))

	// Property 10.5: Daily limit matches tier configuration
	properties.Property("daily limit matches tier configuration", prop.ForAll(
		func(tier domain.Tier) bool {
			svc, userRepo, aiUsageRepo := newTestAIService(defaultConfig)
			ctx := context.Background()

			// Create user with specified tier
			user := domain.NewUser("test@example.com", "hash", "Test User")
			user.Tier = tier
			userRepo.users[user.ID] = user
			userRepo.emailIndex[user.Email] = user
			aiUsageRepo.usage[user.ID] = 0

			// Get AI usage info
			info, err := svc.GetAIUsageInfo(ctx, user.ID)
			if err != nil {
				return false
			}

			expectedLimit := getDailyLimitForTier(tier, defaultConfig)
			return info.DailyLimit == expectedLimit
		},
		genValidTierForAI(),
	))

	// Property 10.6: Remaining is never negative for non-pro tiers
	properties.Property("remaining is never negative for non-pro tiers", prop.ForAll(
		func(tier domain.Tier, usedToday int) bool {
			// Skip pro tier
			if tier == domain.TierPro {
				return true
			}

			svc, userRepo, aiUsageRepo := newTestAIService(defaultConfig)
			ctx := context.Background()

			// Create user with specified tier
			user := domain.NewUser("test@example.com", "hash", "Test User")
			user.Tier = tier
			userRepo.users[user.ID] = user
			userRepo.emailIndex[user.Email] = user

			// Set daily usage (can exceed limit)
			aiUsageRepo.usage[user.ID] = usedToday

			// Get AI usage info
			info, err := svc.GetAIUsageInfo(ctx, user.ID)
			if err != nil {
				return false
			}

			// Remaining should be >= 0 for non-pro tiers
			return info.Remaining >= 0
		},
		genNonProTier(),
		gen.IntRange(0, 500), // Usage can exceed limit
	))

	// Property 10.7: AI usage info is deterministic
	properties.Property("ai usage info is deterministic", prop.ForAll(
		func(tier domain.Tier, usedToday int) bool {
			svc, userRepo, aiUsageRepo := newTestAIService(defaultConfig)
			ctx := context.Background()

			// Create user with specified tier
			user := domain.NewUser("test@example.com", "hash", "Test User")
			user.Tier = tier
			userRepo.users[user.ID] = user
			userRepo.emailIndex[user.Email] = user
			aiUsageRepo.usage[user.ID] = usedToday

			// Get AI usage info twice
			info1, err1 := svc.GetAIUsageInfo(ctx, user.ID)
			info2, err2 := svc.GetAIUsageInfo(ctx, user.ID)

			if err1 != nil || err2 != nil {
				return false
			}

			// Results should be identical (except ResetAt which depends on time)
			return info1.UsedToday == info2.UsedToday &&
				info1.DailyLimit == info2.DailyLimit &&
				info1.Remaining == info2.Remaining &&
				info1.Tier == info2.Tier &&
				info1.IsUnlimited == info2.IsUnlimited
		},
		genValidTierForAI(),
		gen.IntRange(0, 100),
	))

	properties.TestingRun(t)
}

// Helper function to get daily limit for tier
func getDailyLimitForTier(tier domain.Tier, cfg config.AILimitConfig) int {
	switch tier {
	case domain.TierFree:
		return cfg.FreeDailyLimit
	case domain.TierPremium:
		return cfg.PremiumDailyLimit
	case domain.TierPro:
		if cfg.ProDailyLimit <= 0 {
			return -1
		}
		return cfg.ProDailyLimit
	default:
		return cfg.FreeDailyLimit
	}
}

// Helper function to create test AI service with mocks
func newTestAIService(cfg config.AILimitConfig) (AIService, *mockUserRepo, *mockAIUsageRepo) {
	userRepo := newMockUserRepo()
	aiUsageRepo := newMockAIUsageRepo()

	svc := NewAIService(userRepo, aiUsageRepo, cfg)

	return svc, userRepo, aiUsageRepo
}

// Mock AI usage repository for testing
type mockAIUsageRepo struct {
	usage map[uuid.UUID]int
}

func newMockAIUsageRepo() *mockAIUsageRepo {
	return &mockAIUsageRepo{
		usage: make(map[uuid.UUID]int),
	}
}

func (m *mockAIUsageRepo) GetDailyUsage(ctx context.Context, userID uuid.UUID) (int, error) {
	if usage, ok := m.usage[userID]; ok {
		return usage, nil
	}
	return 0, nil
}

func (m *mockAIUsageRepo) IncrementDailyUsage(ctx context.Context, userID uuid.UUID) error {
	m.usage[userID]++
	return nil
}

// Generator for valid tiers (all tiers)
func genValidTierForAI() gopter.Gen {
	return gopter.Gen(func(params *gopter.GenParameters) *gopter.GenResult {
		tiers := domain.ValidTiers()
		idx := params.Rng.Intn(len(tiers))
		return gopter.NewGenResult(tiers[idx], gopter.NoShrinker)
	})
}

// Generator for non-pro tiers only
func genNonProTier() gopter.Gen {
	return gopter.Gen(func(params *gopter.GenParameters) *gopter.GenResult {
		tiers := []domain.Tier{domain.TierFree, domain.TierPremium}
		idx := params.Rng.Intn(len(tiers))
		return gopter.NewGenResult(tiers[idx], gopter.NoShrinker)
	})
}

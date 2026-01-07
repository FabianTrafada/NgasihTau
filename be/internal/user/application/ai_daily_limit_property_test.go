// Package application contains property-based tests for AI daily limit enforcement.
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

// **Feature: user-storage-limit, Property 9: AI Daily Limit Enforcement**
//
// *For any* user with tier T and daily usage count C:
// - IF T is "pro" THEN AI requests SHALL always be allowed (unlimited)
// - IF C >= daily_limit[T] THEN AI requests SHALL be rejected
// - IF C < daily_limit[T] THEN AI requests SHALL be allowed and C SHALL increment
//
// **Validates: Requirements 9.1, 9.3, 9.4, 9.5**

func TestProperty_AIDailyLimitEnforcement(t *testing.T) {
	parameters := gopter.DefaultTestParametersWithSeed(12345)
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	// Default configuration
	defaultConfig := config.AILimitConfig{
		FreeDailyLimit:    20,
		PremiumDailyLimit: 100,
		ProDailyLimit:     -1, // Unlimited
	}

	// Property 9.1: Pro tier users always pass AI limit check (unlimited)
	// Validates: Requirement 9.1 - THE Storage_Limit_System SHALL enforce daily AI chat message limits per tier (pro=unlimited)
	properties.Property("pro tier always passes AI limit check", prop.ForAll(
		func(usedToday int) bool {
			svc, userRepo, aiUsageRepo := newTestAIService(defaultConfig)
			ctx := context.Background()

			// Create pro user
			user := domain.NewUser("test@example.com", "hash", "Test User")
			user.Tier = domain.TierPro
			userRepo.users[user.ID] = user
			userRepo.emailIndex[user.Email] = user

			// Set daily usage to any value (even very high)
			aiUsageRepo.usage[user.ID] = usedToday

			// Check AI limit - should always pass for pro tier
			err := svc.CheckAILimit(ctx, user.ID)

			return err == nil
		},
		gen.IntRange(0, 10000), // Usage 0-10000 messages (shouldn't matter for pro)
	))

	// Property 9.2: Non-pro users at or above limit are rejected
	// Validates: Requirement 9.4 - WHEN a user sends an AI chat message THEN THE Storage_Limit_System SHALL check if the daily limit is exceeded
	// Validates: Requirement 9.5 - IF the AI daily limit is exceeded THEN THE Storage_Limit_System SHALL reject the request with error code AI_LIMIT_EXCEEDED
	properties.Property("non-pro users at or above limit are rejected", prop.ForAll(
		func(tier domain.Tier, limitOffset int) bool {
			// Skip pro tier (handled separately)
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

			// Get daily limit for tier
			dailyLimit := getDailyLimitForTier(tier, defaultConfig)

			// Set usage to be at or above limit (limit + offset where offset >= 0)
			usedToday := dailyLimit + limitOffset
			aiUsageRepo.usage[user.ID] = usedToday

			// Check AI limit - should be rejected
			err := svc.CheckAILimit(ctx, user.ID)

			// Should return AI limit exceeded error
			if err == nil {
				return false
			}

			appErr, ok := err.(*errors.AppError)
			if !ok {
				return false
			}

			return appErr.Code == errors.CodeAILimitExceeded
		},
		genNonProTier(),
		gen.IntRange(0, 100), // Offset 0-100 above limit
	))

	// Property 9.3: Non-pro users below limit are allowed
	// Validates: Requirement 9.4 - WHEN a user sends an AI chat message THEN THE Storage_Limit_System SHALL check if the daily limit is exceeded
	properties.Property("non-pro users below limit are allowed", prop.ForAll(
		func(tier domain.Tier, usagePercent int) bool {
			// Skip pro tier (handled separately)
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

			// Get daily limit for tier
			dailyLimit := getDailyLimitForTier(tier, defaultConfig)

			// Set usage to be below limit (percentage of limit, 0-99%)
			usedToday := (dailyLimit * usagePercent) / 100
			if usedToday >= dailyLimit {
				usedToday = dailyLimit - 1
			}
			aiUsageRepo.usage[user.ID] = usedToday

			// Check AI limit - should be allowed
			err := svc.CheckAILimit(ctx, user.ID)

			return err == nil
		},
		genNonProTier(),
		gen.IntRange(0, 99), // Usage percentage 0-99%
	))

	// Property 9.4: Limit enforcement is consistent with tier configuration
	// Validates: Requirement 9.1 - THE Storage_Limit_System SHALL enforce daily AI chat message limits per tier (default: free=20, premium=100, pro=unlimited)
	properties.Property("limit enforcement matches tier configuration", prop.ForAll(
		func(tier domain.Tier) bool {
			svc, userRepo, aiUsageRepo := newTestAIService(defaultConfig)
			ctx := context.Background()

			// Create user with specified tier
			user := domain.NewUser("test@example.com", "hash", "Test User")
			user.Tier = tier
			userRepo.users[user.ID] = user
			userRepo.emailIndex[user.Email] = user

			// Get expected daily limit for tier
			expectedLimit := getDailyLimitForTier(tier, defaultConfig)

			// For pro tier, any usage should be allowed
			if tier == domain.TierPro {
				aiUsageRepo.usage[user.ID] = 1000
				err := svc.CheckAILimit(ctx, user.ID)
				return err == nil
			}

			// For non-pro tiers, test boundary conditions
			// At limit - 1: should be allowed
			aiUsageRepo.usage[user.ID] = expectedLimit - 1
			errBelowLimit := svc.CheckAILimit(ctx, user.ID)
			if errBelowLimit != nil {
				return false
			}

			// At limit: should be rejected
			aiUsageRepo.usage[user.ID] = expectedLimit
			errAtLimit := svc.CheckAILimit(ctx, user.ID)
			if errAtLimit == nil {
				return false
			}

			return true
		},
		genValidTierForAI(),
	))

	// Property 9.5: Free tier has correct limit (20)
	// Validates: Requirement 9.1 - default: free=20
	properties.Property("free tier has limit of 20", prop.ForAll(
		func(usedToday int) bool {
			svc, userRepo, aiUsageRepo := newTestAIService(defaultConfig)
			ctx := context.Background()

			// Create free user
			user := domain.NewUser("test@example.com", "hash", "Test User")
			user.Tier = domain.TierFree
			userRepo.users[user.ID] = user
			userRepo.emailIndex[user.Email] = user

			aiUsageRepo.usage[user.ID] = usedToday

			err := svc.CheckAILimit(ctx, user.ID)

			// Should be allowed if usedToday < 20, rejected if >= 20
			if usedToday < 20 {
				return err == nil
			}
			return err != nil
		},
		gen.IntRange(0, 40), // Test around the boundary
	))

	// Property 9.6: Premium tier has correct limit (100)
	// Validates: Requirement 9.1 - default: premium=100
	properties.Property("premium tier has limit of 100", prop.ForAll(
		func(usedToday int) bool {
			svc, userRepo, aiUsageRepo := newTestAIService(defaultConfig)
			ctx := context.Background()

			// Create premium user
			user := domain.NewUser("test@example.com", "hash", "Test User")
			user.Tier = domain.TierPremium
			userRepo.users[user.ID] = user
			userRepo.emailIndex[user.Email] = user

			aiUsageRepo.usage[user.ID] = usedToday

			err := svc.CheckAILimit(ctx, user.ID)

			// Should be allowed if usedToday < 100, rejected if >= 100
			if usedToday < 100 {
				return err == nil
			}
			return err != nil
		},
		gen.IntRange(0, 150), // Test around the boundary
	))

	// Property 9.7: AI limit check is deterministic
	// Validates: Requirement 9.3 - THE Storage_Limit_System SHALL track daily AI message count per user
	properties.Property("AI limit check is deterministic", prop.ForAll(
		func(tier domain.Tier, usedToday int) bool {
			svc, userRepo, aiUsageRepo := newTestAIService(defaultConfig)
			ctx := context.Background()

			// Create user with specified tier
			user := domain.NewUser("test@example.com", "hash", "Test User")
			user.Tier = tier
			userRepo.users[user.ID] = user
			userRepo.emailIndex[user.Email] = user
			aiUsageRepo.usage[user.ID] = usedToday

			// Check AI limit twice
			err1 := svc.CheckAILimit(ctx, user.ID)
			err2 := svc.CheckAILimit(ctx, user.ID)

			// Results should be identical
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
		genValidTierForAI(),
		gen.IntRange(0, 150),
	))

	// Property 9.8: Error details contain correct information when limit exceeded
	// Validates: Requirement 9.5 - IF the AI daily limit is exceeded THEN THE Storage_Limit_System SHALL reject the request with error code AI_LIMIT_EXCEEDED
	properties.Property("error details contain correct information", prop.ForAll(
		func(tier domain.Tier, limitOffset int) bool {
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

			// Get daily limit for tier
			dailyLimit := getDailyLimitForTier(tier, defaultConfig)

			// Set usage to be at or above limit
			usedToday := dailyLimit + limitOffset
			aiUsageRepo.usage[user.ID] = usedToday

			// Check AI limit
			err := svc.CheckAILimit(ctx, user.ID)

			if err == nil {
				return false
			}

			appErr, ok := err.(*errors.AppError)
			if !ok {
				return false
			}

			// Verify error code
			if appErr.Code != errors.CodeAILimitExceeded {
				return false
			}

			// Verify error details contain expected fields
			hasUsedToday := false
			hasDailyLimit := false
			hasTier := false

			for _, detail := range appErr.Details {
				switch detail.Field {
				case "used_today":
					if detail.Value == usedToday {
						hasUsedToday = true
					}
				case "daily_limit":
					if detail.Value == dailyLimit {
						hasDailyLimit = true
					}
				case "tier":
					if detail.Value == string(tier) {
						hasTier = true
					}
				}
			}

			return hasUsedToday && hasDailyLimit && hasTier
		},
		genNonProTier(),
		gen.IntRange(0, 50),
	))

	properties.TestingRun(t)
}

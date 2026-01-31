// Package application contains property-based tests for tier-based feature access.
package application

import (
	"context"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/prop"

	"ngasihtau/internal/common/config"
	"ngasihtau/internal/common/errors"
	"ngasihtau/internal/user/domain"
)

// **Feature: user-storage-limit, Property 11: Tier-Based Feature Access**
//
// *For any* user attempting to access a premium feature:
// - Chat export: allowed IFF tier is "premium" OR "pro"
// - Pod-wide AI chat: allowed IFF tier is "pro"
// - Question generation: allowed IFF tier is "pro"
//
// **Validates: Requirements 11.1, 11.2, 11.3, 12.1, 12.6**

func TestProperty_TierBasedFeatureAccess(t *testing.T) {
	parameters := gopter.DefaultTestParametersWithSeed(12345)
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	// Default configuration
	defaultConfig := config.AILimitConfig{
		FreeDailyLimit:    20,
		PremiumDailyLimit: 100,
		ProDailyLimit:     -1, // Unlimited
	}

	// Property 11.1: Chat export is allowed for premium and pro tiers only
	// Validates: Requirement 11.1 - WHERE a user has premium or pro tier
	// THEN THE Storage_Limit_System SHALL allow access to chat export feature
	properties.Property("chat export allowed for premium and pro tiers", prop.ForAll(
		func(tier domain.Tier) bool {
			svc, userRepo, _ := newTestAIService(defaultConfig)
			ctx := context.Background()

			// Create user with specified tier
			user := domain.NewUser("test@example.com", "hash", "Test User")
			user.Tier = tier
			userRepo.users[user.ID] = user
			userRepo.emailIndex[user.Email] = user

			// Check feature access
			err := svc.CanAccessFeature(ctx, user.ID, FeatureChatExport)

			// Chat export should be allowed for premium and pro tiers
			expectedAllowed := tier == domain.TierPremium || tier == domain.TierPro

			if expectedAllowed {
				return err == nil
			}
			// Should be rejected with PREMIUM_FEATURE_REQUIRED for free tier
			if err == nil {
				return false
			}
			appErr, ok := err.(*errors.AppError)
			if !ok {
				return false
			}
			return appErr.Code == errors.CodePremiumFeatureRequired
		},
		genValidTierForAI(),
	))

	// Property 11.2: Pod-wide AI chat is allowed for pro tier only
	// Validates: Requirement 11.2 - WHERE a user has pro tier
	// THEN THE Storage_Limit_System SHALL allow access to pod-wide AI chat
	properties.Property("pod-wide AI chat allowed for pro tier only", prop.ForAll(
		func(tier domain.Tier) bool {
			svc, userRepo, _ := newTestAIService(defaultConfig)
			ctx := context.Background()

			// Create user with specified tier
			user := domain.NewUser("test@example.com", "hash", "Test User")
			user.Tier = tier
			userRepo.users[user.ID] = user
			userRepo.emailIndex[user.Email] = user

			// Check feature access
			err := svc.CanAccessFeature(ctx, user.ID, FeaturePodWideChat)

			// Pod-wide chat should be allowed only for pro tier
			expectedAllowed := tier == domain.TierPro

			if expectedAllowed {
				return err == nil
			}
			// Should be rejected with PRO_FEATURE_REQUIRED for non-pro tiers
			if err == nil {
				return false
			}
			appErr, ok := err.(*errors.AppError)
			if !ok {
				return false
			}
			return appErr.Code == errors.CodeProFeatureRequired
		},
		genValidTierForAI(),
	))

	// Property 11.3: Question generation is allowed for pro tier only
	// Validates: Requirement 12.1 - WHERE a user has pro tier
	// THEN THE Storage_Limit_System SHALL allow access to AI question generation feature
	// Validates: Requirement 12.6 - IF a non-pro user attempts to generate questions
	// THEN THE Storage_Limit_System SHALL reject with error code PRO_FEATURE_REQUIRED
	properties.Property("question generation allowed for pro tier only", prop.ForAll(
		func(tier domain.Tier) bool {
			svc, userRepo, _ := newTestAIService(defaultConfig)
			ctx := context.Background()

			// Create user with specified tier
			user := domain.NewUser("test@example.com", "hash", "Test User")
			user.Tier = tier
			userRepo.users[user.ID] = user
			userRepo.emailIndex[user.Email] = user

			// Check feature access
			err := svc.CanAccessFeature(ctx, user.ID, FeatureQuestionGeneration)

			// Question generation should be allowed only for pro tier
			expectedAllowed := tier == domain.TierPro

			if expectedAllowed {
				return err == nil
			}
			// Should be rejected with PRO_FEATURE_REQUIRED for non-pro tiers
			if err == nil {
				return false
			}
			appErr, ok := err.(*errors.AppError)
			if !ok {
				return false
			}
			return appErr.Code == errors.CodeProFeatureRequired
		},
		genValidTierForAI(),
	))

	// Property 11.4: Free tier is rejected for premium features with correct error
	// Validates: Requirement 11.3 - IF a free user attempts to access premium AI features
	// THEN THE Storage_Limit_System SHALL reject with error code PREMIUM_FEATURE_REQUIRED
	properties.Property("free tier rejected for chat export with PREMIUM_FEATURE_REQUIRED", prop.ForAll(
		func(_ bool) bool {
			svc, userRepo, _ := newTestAIService(defaultConfig)
			ctx := context.Background()

			// Create free user
			user := domain.NewUser("test@example.com", "hash", "Test User")
			user.Tier = domain.TierFree
			userRepo.users[user.ID] = user
			userRepo.emailIndex[user.Email] = user

			// Check feature access for chat export
			err := svc.CanAccessFeature(ctx, user.ID, FeatureChatExport)

			if err == nil {
				return false
			}
			appErr, ok := err.(*errors.AppError)
			if !ok {
				return false
			}
			// Verify error code and details
			if appErr.Code != errors.CodePremiumFeatureRequired {
				return false
			}
			// Verify error details contain feature and required_tier
			hasFeature := false
			hasRequiredTier := false
			for _, detail := range appErr.Details {
				if detail.Field == "feature" && detail.Value == FeatureChatExport {
					hasFeature = true
				}
				if detail.Field == "required_tier" && detail.Value == "premium" {
					hasRequiredTier = true
				}
			}
			return hasFeature && hasRequiredTier
		},
		genBool(),
	))

	// Property 11.5: Non-pro tiers rejected for pro features with correct error
	// Validates: Requirement 12.6 - IF a non-pro user attempts to generate questions
	// THEN THE Storage_Limit_System SHALL reject with error code PRO_FEATURE_REQUIRED
	properties.Property("non-pro tiers rejected for pro features with PRO_FEATURE_REQUIRED", prop.ForAll(
		func(tier domain.Tier, feature string) bool {
			// Skip pro tier
			if tier == domain.TierPro {
				return true
			}

			svc, userRepo, _ := newTestAIService(defaultConfig)
			ctx := context.Background()

			// Create user with non-pro tier
			user := domain.NewUser("test@example.com", "hash", "Test User")
			user.Tier = tier
			userRepo.users[user.ID] = user
			userRepo.emailIndex[user.Email] = user

			// Check feature access for pro-only features
			err := svc.CanAccessFeature(ctx, user.ID, feature)

			if err == nil {
				return false
			}
			appErr, ok := err.(*errors.AppError)
			if !ok {
				return false
			}
			// Verify error code
			if appErr.Code != errors.CodeProFeatureRequired {
				return false
			}
			// Verify error details contain feature and required_tier
			hasFeature := false
			hasRequiredTier := false
			for _, detail := range appErr.Details {
				if detail.Field == "feature" && detail.Value == feature {
					hasFeature = true
				}
				if detail.Field == "required_tier" && detail.Value == "pro" {
					hasRequiredTier = true
				}
			}
			return hasFeature && hasRequiredTier
		},
		genNonProTier(),
		genProOnlyFeature(),
	))

	// Property 11.6: Feature access is deterministic
	properties.Property("feature access is deterministic", prop.ForAll(
		func(tier domain.Tier, feature string) bool {
			svc, userRepo, _ := newTestAIService(defaultConfig)
			ctx := context.Background()

			// Create user with specified tier
			user := domain.NewUser("test@example.com", "hash", "Test User")
			user.Tier = tier
			userRepo.users[user.ID] = user
			userRepo.emailIndex[user.Email] = user

			// Check feature access twice
			err1 := svc.CanAccessFeature(ctx, user.ID, feature)
			err2 := svc.CanAccessFeature(ctx, user.ID, feature)

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
		genValidFeature(),
	))

	// Property 11.7: Pro tier has access to all features
	properties.Property("pro tier has access to all features", prop.ForAll(
		func(feature string) bool {
			svc, userRepo, _ := newTestAIService(defaultConfig)
			ctx := context.Background()

			// Create pro user
			user := domain.NewUser("test@example.com", "hash", "Test User")
			user.Tier = domain.TierPro
			userRepo.users[user.ID] = user
			userRepo.emailIndex[user.Email] = user

			// Check feature access
			err := svc.CanAccessFeature(ctx, user.ID, feature)

			// Pro tier should have access to all valid features
			return err == nil
		},
		genValidFeature(),
	))

	// Property 11.8: Premium tier has access to premium features but not pro features
	properties.Property("premium tier has access to premium features but not pro features", prop.ForAll(
		func(_ bool) bool {
			svc, userRepo, _ := newTestAIService(defaultConfig)
			ctx := context.Background()

			// Create premium user
			user := domain.NewUser("test@example.com", "hash", "Test User")
			user.Tier = domain.TierPremium
			userRepo.users[user.ID] = user
			userRepo.emailIndex[user.Email] = user

			// Premium should have access to chat export
			errChatExport := svc.CanAccessFeature(ctx, user.ID, FeatureChatExport)
			if errChatExport != nil {
				return false
			}

			// Premium should NOT have access to pod-wide chat
			errPodWideChat := svc.CanAccessFeature(ctx, user.ID, FeaturePodWideChat)
			if errPodWideChat == nil {
				return false
			}

			// Premium should NOT have access to question generation
			errQuestionGen := svc.CanAccessFeature(ctx, user.ID, FeatureQuestionGeneration)
			if errQuestionGen == nil {
				return false
			}

			return true
		},
		genBool(),
	))

	// Property 11.9: Unknown features are rejected
	properties.Property("unknown features are rejected with bad request", prop.ForAll(
		func(tier domain.Tier) bool {
			svc, userRepo, _ := newTestAIService(defaultConfig)
			ctx := context.Background()

			// Create user with specified tier
			user := domain.NewUser("test@example.com", "hash", "Test User")
			user.Tier = tier
			userRepo.users[user.ID] = user
			userRepo.emailIndex[user.Email] = user

			// Check access for unknown feature
			err := svc.CanAccessFeature(ctx, user.ID, "unknown_feature")

			if err == nil {
				return false
			}
			appErr, ok := err.(*errors.AppError)
			if !ok {
				return false
			}
			return appErr.Code == errors.CodeBadRequest
		},
		genValidTierForAI(),
	))

	properties.TestingRun(t)
}

// Generator for valid features
func genValidFeature() gopter.Gen {
	return gopter.Gen(func(params *gopter.GenParameters) *gopter.GenResult {
		features := []string{FeatureChatExport, FeaturePodWideChat, FeatureQuestionGeneration}
		idx := params.Rng.Intn(len(features))
		return gopter.NewGenResult(features[idx], gopter.NoShrinker)
	})
}

// Generator for pro-only features
func genProOnlyFeature() gopter.Gen {
	return gopter.Gen(func(params *gopter.GenParameters) *gopter.GenResult {
		features := []string{FeaturePodWideChat, FeatureQuestionGeneration}
		idx := params.Rng.Intn(len(features))
		return gopter.NewGenResult(features[idx], gopter.NoShrinker)
	})
}

// Generator for boolean values (used as dummy input for properties that don't need random input)
func genBool() gopter.Gen {
	return gopter.Gen(func(params *gopter.GenParameters) *gopter.GenResult {
		return gopter.NewGenResult(params.Rng.Intn(2) == 1, gopter.NoShrinker)
	})
}

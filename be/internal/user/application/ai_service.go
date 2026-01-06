// Package application contains the business logic and use cases for the User Service.
package application

import (
	"context"
	"time"

	"github.com/google/uuid"

	"ngasihtau/internal/common/config"
	"ngasihtau/internal/common/errors"
	"ngasihtau/internal/user/domain"
)

// Feature constants define the premium and pro features.
// Used by CanAccessFeature to check tier-based access.
const (
	// FeatureChatExport allows exporting chat conversations.
	// Requires premium tier or higher.
	// Implements requirement 11.1.
	FeatureChatExport = "chat_export"

	// FeaturePodWideChat allows AI chat across all materials in a pod.
	// Requires pro tier.
	// Implements requirement 11.2.
	FeaturePodWideChat = "pod_wide_chat"

	// FeatureQuestionGeneration allows generating quiz questions from materials.
	// Requires pro tier.
	// Implements requirement 12.1.
	FeatureQuestionGeneration = "question_generation"
)

// AIService defines the interface for AI-related business operations.
// Implements requirements 9.1-9.6, 10.1-10.5 for AI usage limits.
type AIService interface {
	// GetAIUsageInfo returns AI usage information for a user.
	// Implements requirements 10.1-10.5.
	GetAIUsageInfo(ctx context.Context, userID uuid.UUID) (*domain.AIUsageInfo, error)

	// CheckAILimit checks if a user has remaining AI messages for today.
	// Returns nil if the user can send AI messages, or an error if the limit is exceeded.
	// Pro tier users always pass this check (unlimited).
	// Implements requirements 9.1, 9.4, 9.5.
	CheckAILimit(ctx context.Context, userID uuid.UUID) error

	// IncrementAIUsage increments the user's daily AI usage count.
	// Should be called after a successful AI chat message is processed.
	// Implements requirement 9.3.
	IncrementAIUsage(ctx context.Context, userID uuid.UUID) error

	// CanAccessFeature checks if a user's tier allows access to a specific feature.
	// Returns nil if access is allowed, or an appropriate error if not.
	// Implements requirements 11.1, 11.2, 11.3, 11.4, 12.1, 12.6.
	CanAccessFeature(ctx context.Context, userID uuid.UUID, feature string) error
}

// aiService implements the AIService interface.
type aiService struct {
	userRepo      domain.UserRepository
	aiUsageRepo   domain.AIUsageRepository
	aiLimitConfig config.AILimitConfig
}

// NewAIService creates a new AIService instance.
func NewAIService(
	userRepo domain.UserRepository,
	aiUsageRepo domain.AIUsageRepository,
	aiLimitConfig config.AILimitConfig,
) AIService {
	return &aiService{
		userRepo:      userRepo,
		aiUsageRepo:   aiUsageRepo,
		aiLimitConfig: aiLimitConfig,
	}
}

// nextMidnightUTC returns the time of the next midnight UTC.
func nextMidnightUTC() time.Time {
	now := time.Now().UTC()
	return time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, time.UTC)
}

// GetAIUsageInfo returns AI usage information for a user.
// Implements requirements 10.1-10.5.
func (s *aiService) GetAIUsageInfo(ctx context.Context, userID uuid.UUID) (*domain.AIUsageInfo, error) {
	// Get user to retrieve their tier
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Get daily usage from AIUsageRepository
	usedToday, err := s.aiUsageRepo.GetDailyUsage(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Calculate daily limit based on tier (pro = unlimited/-1)
	dailyLimit := s.getDailyLimitForTier(user.Tier)

	// Check if unlimited (pro tier)
	isUnlimited := user.Tier == domain.TierPro

	// Calculate remaining messages
	var remaining int
	if isUnlimited {
		remaining = -1 // -1 indicates unlimited
	} else {
		remaining = dailyLimit - usedToday
		if remaining < 0 {
			remaining = 0
		}
	}

	// Set reset time to next midnight UTC
	resetAt := nextMidnightUTC()

	return &domain.AIUsageInfo{
		UsedToday:   usedToday,
		DailyLimit:  dailyLimit,
		Remaining:   remaining,
		ResetAt:     resetAt,
		Tier:        user.Tier,
		IsUnlimited: isUnlimited,
	}, nil
}

// getDailyLimitForTier returns the daily AI message limit for a given tier.
// Returns -1 for unlimited (pro tier).
func (s *aiService) getDailyLimitForTier(tier domain.Tier) int {
	switch tier {
	case domain.TierFree:
		return s.aiLimitConfig.FreeDailyLimit
	case domain.TierPremium:
		return s.aiLimitConfig.PremiumDailyLimit
	case domain.TierPro:
		// Pro tier has unlimited messages, represented as -1
		if s.aiLimitConfig.ProDailyLimit <= 0 {
			return -1
		}
		return s.aiLimitConfig.ProDailyLimit
	default:
		// Default to free tier limit for unknown tiers
		return s.aiLimitConfig.FreeDailyLimit
	}
}

// CheckAILimit checks if a user has remaining AI messages for today.
// Returns nil if the user can send AI messages, or an error if the limit is exceeded.
// Pro tier users always pass this check (unlimited).
// Implements requirements 9.1, 9.4, 9.5.
func (s *aiService) CheckAILimit(ctx context.Context, userID uuid.UUID) error {
	// Get user to retrieve their tier
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return err
	}

	// Pro tier has unlimited AI messages - always allow
	if user.Tier == domain.TierPro {
		return nil
	}

	// Get daily usage count from AIUsageRepository
	usedToday, err := s.aiUsageRepo.GetDailyUsage(ctx, userID)
	if err != nil {
		return err
	}

	// Get daily limit for user's tier
	dailyLimit := s.getDailyLimitForTier(user.Tier)

	// Compare against tier limit
	if usedToday >= dailyLimit {
		return errors.AILimitExceeded(usedToday, dailyLimit, string(user.Tier))
	}

	return nil
}

// IncrementAIUsage increments the user's daily AI usage count.
// Should be called after a successful AI chat message is processed.
// Implements requirement 9.3.
func (s *aiService) IncrementAIUsage(ctx context.Context, userID uuid.UUID) error {
	return s.aiUsageRepo.IncrementDailyUsage(ctx, userID)
}

// CanAccessFeature checks if a user's tier allows access to a specific feature.
// Returns nil if access is allowed, or an appropriate error if not.
// Implements requirements 11.1, 11.2, 11.3, 11.4, 12.1, 12.6.
func (s *aiService) CanAccessFeature(ctx context.Context, userID uuid.UUID, feature string) error {
	// Get user to retrieve their tier
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return err
	}

	// Check feature access based on tier
	switch feature {
	case FeatureChatExport:
		// Chat export requires premium tier or higher
		// Implements requirement 11.1
		if !user.Tier.IsAtLeast(domain.TierPremium) {
			return errors.PremiumFeatureRequired(feature)
		}
	case FeaturePodWideChat:
		// Pod-wide AI chat requires pro tier
		// Implements requirement 11.2
		if !user.Tier.IsAtLeast(domain.TierPro) {
			return errors.ProFeatureRequired(feature)
		}
	case FeatureQuestionGeneration:
		// Question generation requires pro tier
		// Implements requirement 12.1, 12.6
		if !user.Tier.IsAtLeast(domain.TierPro) {
			return errors.ProFeatureRequired(feature)
		}
	default:
		// Unknown feature - deny access by default
		return errors.BadRequest("unknown feature: " + feature)
	}

	return nil
}

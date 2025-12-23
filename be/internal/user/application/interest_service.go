// Package application contains the business logic for learning interests.
package application

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"ngasihtau/internal/common/errors"
	"ngasihtau/internal/user/domain"
)

// LearningInterestService defines the interface for learning interest operations.
type LearningInterestService interface {
	// GetPredefinedInterests returns all active predefined interests.
	GetPredefinedInterests(ctx context.Context) (*PredefinedInterestsResult, error)

	// GetPredefinedInterestsByCategory returns predefined interests grouped by category.
	GetPredefinedInterestsByCategory(ctx context.Context) (*GroupedInterestsResult, error)

	// GetUserInterests returns all learning interests for a user.
	GetUserInterests(ctx context.Context, userID uuid.UUID) (*UserInterestsResult, error)

	// SetUserInterests sets/replaces all learning interests for a user (used during onboarding).
	SetUserInterests(ctx context.Context, userID uuid.UUID, input SetInterestsInput) error

	// AddUserInterest adds a single interest to a user's list.
	AddUserInterest(ctx context.Context, userID uuid.UUID, input AddInterestInput) (*domain.InterestSummary, error)

	// RemoveUserInterest removes a single interest from a user's list.
	RemoveUserInterest(ctx context.Context, userID uuid.UUID, interestID uuid.UUID) error

	// CompleteOnboarding marks user onboarding as completed.
	CompleteOnboarding(ctx context.Context, userID uuid.UUID) error

	// CheckOnboardingStatus returns whether user has completed onboarding.
	CheckOnboardingStatus(ctx context.Context, userID uuid.UUID) (*OnboardingStatus, error)
}

// PredefinedInterestsResult contains the list of predefined interests.
type PredefinedInterestsResult struct {
	Interests []*domain.PredefinedInterest `json:"interests"`
	Total     int                          `json:"total"`
}

// GroupedInterestsResult contains predefined interests grouped by category.
type GroupedInterestsResult struct {
	Categories []CategoryGroup `json:"categories"`
	Total      int             `json:"total"`
}

// CategoryGroup represents a group of interests in a category.
type CategoryGroup struct {
	Category  string                       `json:"category"`
	Interests []*domain.PredefinedInterest `json:"interests"`
}

// UserInterestsResult contains the user's selected interests.
type UserInterestsResult struct {
	Interests []*domain.InterestSummary `json:"interests"`
	Total     int                       `json:"total"`
}

// SetInterestsInput contains the data for setting user interests during onboarding.
type SetInterestsInput struct {
	PredefinedInterestIDs []uuid.UUID `json:"predefined_interest_ids" validate:"omitempty,dive,uuid"`
	CustomInterests       []string    `json:"custom_interests" validate:"omitempty,dive,min=2,max=100"`
}

// AddInterestInput contains the data for adding a single interest.
type AddInterestInput struct {
	PredefinedInterestID *uuid.UUID `json:"predefined_interest_id,omitempty" validate:"omitempty,uuid"`
	CustomInterest       *string    `json:"custom_interest,omitempty" validate:"omitempty,min=2,max=100"`
}

// OnboardingStatus contains the user's onboarding status.
type OnboardingStatus struct {
	OnboardingCompleted bool                      `json:"onboarding_completed"`
	InterestsCount      int                       `json:"interests_count"`
	Interests           []*domain.InterestSummary `json:"interests,omitempty"`
}

// learningInterestService implements the LearningInterestService interface.
type learningInterestService struct {
	predefinedInterestRepo   domain.PredefinedInterestRepository
	userLearningInterestRepo domain.UserLearningInterestRepository
	userRepo                 domain.UserRepository
}

// NewLearningInterestService creates a new LearningInterestService instance.
func NewLearningInterestService(
	predefinedInterestRepo domain.PredefinedInterestRepository,
	userLearningInterestRepo domain.UserLearningInterestRepository,
	userRepo domain.UserRepository,
) LearningInterestService {
	return &learningInterestService{
		predefinedInterestRepo:   predefinedInterestRepo,
		userLearningInterestRepo: userLearningInterestRepo,
		userRepo:                 userRepo,
	}
}

// GetPredefinedInterests returns all active predefined interests.
func (s *learningInterestService) GetPredefinedInterests(ctx context.Context) (*PredefinedInterestsResult, error) {
	interests, err := s.predefinedInterestRepo.FindAll(ctx)
	if err != nil {
		return nil, err
	}

	return &PredefinedInterestsResult{
		Interests: interests,
		Total:     len(interests),
	}, nil
}

// GetPredefinedInterestsByCategory returns predefined interests grouped by category.
func (s *learningInterestService) GetPredefinedInterestsByCategory(ctx context.Context) (*GroupedInterestsResult, error) {
	interests, err := s.predefinedInterestRepo.FindAll(ctx)
	if err != nil {
		return nil, err
	}

	// Group by category
	categoryMap := make(map[string][]*domain.PredefinedInterest)
	categoryOrder := make([]string, 0)

	for _, interest := range interests {
		category := "Other"
		if interest.Category != nil {
			category = *interest.Category
		}

		if _, exists := categoryMap[category]; !exists {
			categoryOrder = append(categoryOrder, category)
		}
		categoryMap[category] = append(categoryMap[category], interest)
	}

	// Build result maintaining order
	categories := make([]CategoryGroup, 0, len(categoryOrder))
	for _, category := range categoryOrder {
		categories = append(categories, CategoryGroup{
			Category:  category,
			Interests: categoryMap[category],
		})
	}

	return &GroupedInterestsResult{
		Categories: categories,
		Total:      len(interests),
	}, nil
}

// GetUserInterests returns all learning interests for a user.
func (s *learningInterestService) GetUserInterests(ctx context.Context, userID uuid.UUID) (*UserInterestsResult, error) {
	summaries, err := s.userLearningInterestRepo.GetInterestSummaries(ctx, userID)
	if err != nil {
		return nil, err
	}

	return &UserInterestsResult{
		Interests: summaries,
		Total:     len(summaries),
	}, nil
}

// SetUserInterests sets/replaces all learning interests for a user.
// This is typically used during onboarding when the user selects their initial interests.
func (s *learningInterestService) SetUserInterests(ctx context.Context, userID uuid.UUID, input SetInterestsInput) error {
	// Validate that at least one interest is provided
	if len(input.PredefinedInterestIDs) == 0 && len(input.CustomInterests) == 0 {
		return errors.BadRequest("at least one interest must be selected")
	}

	// Validate predefined interest IDs exist
	if len(input.PredefinedInterestIDs) > 0 {
		existingInterests, err := s.predefinedInterestRepo.FindByIDs(ctx, input.PredefinedInterestIDs)
		if err != nil {
			return err
		}
		if len(existingInterests) != len(input.PredefinedInterestIDs) {
			return errors.BadRequest("one or more predefined interest IDs are invalid")
		}
	}

	// Delete existing interests for the user
	if err := s.userLearningInterestRepo.DeleteByUserID(ctx, userID); err != nil {
		return err
	}

	// Create new interests
	var interests []*domain.UserLearningInterest

	// Add predefined interests
	for _, predefinedID := range input.PredefinedInterestIDs {
		interest := domain.NewUserLearningInterestFromPredefined(userID, predefinedID)
		interests = append(interests, interest)
	}

	// Add custom interests
	for _, customInterest := range input.CustomInterests {
		trimmed := strings.TrimSpace(customInterest)
		if trimmed == "" {
			continue
		}
		interest := domain.NewUserLearningInterestCustom(userID, trimmed)
		interests = append(interests, interest)
	}

	// Batch create
	if err := s.userLearningInterestRepo.CreateBatch(ctx, interests); err != nil {
		return err
	}

	log.Info().
		Str("user_id", userID.String()).
		Int("predefined_count", len(input.PredefinedInterestIDs)).
		Int("custom_count", len(input.CustomInterests)).
		Msg("user interests updated")

	return nil
}

// AddUserInterest adds a single interest to a user's list.
func (s *learningInterestService) AddUserInterest(ctx context.Context, userID uuid.UUID, input AddInterestInput) (*domain.InterestSummary, error) {
	// Validate input - must have exactly one of predefined or custom
	if input.PredefinedInterestID == nil && input.CustomInterest == nil {
		return nil, errors.BadRequest("either predefined_interest_id or custom_interest must be provided")
	}
	if input.PredefinedInterestID != nil && input.CustomInterest != nil {
		return nil, errors.BadRequest("cannot specify both predefined_interest_id and custom_interest")
	}

	var interest *domain.UserLearningInterest
	var summary *domain.InterestSummary

	if input.PredefinedInterestID != nil {
		// Validate predefined interest exists
		predefined, err := s.predefinedInterestRepo.FindByID(ctx, *input.PredefinedInterestID)
		if err != nil {
			return nil, err
		}

		// Check if already exists
		exists, err := s.userLearningInterestRepo.ExistsByUserIDAndPredefinedID(ctx, userID, *input.PredefinedInterestID)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, errors.Conflict("user learning interest", "interest already selected")
		}

		interest = domain.NewUserLearningInterestFromPredefined(userID, *input.PredefinedInterestID)
		summary = &domain.InterestSummary{
			ID:       interest.ID,
			Name:     predefined.Name,
			Slug:     &predefined.Slug,
			Icon:     predefined.Icon,
			Category: predefined.Category,
			IsCustom: false,
		}
	} else {
		// Custom interest
		trimmed := strings.TrimSpace(*input.CustomInterest)
		if trimmed == "" {
			return nil, errors.BadRequest("custom interest cannot be empty")
		}

		// Check if already exists
		exists, err := s.userLearningInterestRepo.ExistsByUserIDAndCustom(ctx, userID, trimmed)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, errors.Conflict("user learning interest", "custom interest already exists")
		}

		interest = domain.NewUserLearningInterestCustom(userID, trimmed)
		summary = &domain.InterestSummary{
			ID:       interest.ID,
			Name:     trimmed,
			IsCustom: true,
		}
	}

	if err := s.userLearningInterestRepo.Create(ctx, interest); err != nil {
		return nil, err
	}

	return summary, nil
}

// RemoveUserInterest removes a single interest from a user's list.
func (s *learningInterestService) RemoveUserInterest(ctx context.Context, userID uuid.UUID, interestID uuid.UUID) error {
	// Get all user interests to verify ownership
	interests, err := s.userLearningInterestRepo.FindByUserID(ctx, userID)
	if err != nil {
		return err
	}

	// Find the interest to delete
	found := false
	for _, interest := range interests {
		if interest.ID == interestID {
			found = true
			break
		}
	}

	if !found {
		return errors.NotFound("user learning interest", interestID.String())
	}

	return s.userLearningInterestRepo.Delete(ctx, interestID)
}

// CompleteOnboarding marks user onboarding as completed.
func (s *learningInterestService) CompleteOnboarding(ctx context.Context, userID uuid.UUID) error {
	// Check if user exists
	_, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return err
	}

	// Check if user has selected at least one interest
	count, err := s.userLearningInterestRepo.CountByUserID(ctx, userID)
	if err != nil {
		return err
	}

	if count == 0 {
		return errors.BadRequest("please select at least one learning interest before completing onboarding")
	}

	// Update user to mark onboarding as completed
	if err := s.userRepo.SetOnboardingCompleted(ctx, userID, true); err != nil {
		return err
	}

	log.Info().
		Str("user_id", userID.String()).
		Int("interests_count", count).
		Msg("user onboarding completed")

	return nil
}

// CheckOnboardingStatus returns whether user has completed onboarding.
func (s *learningInterestService) CheckOnboardingStatus(ctx context.Context, userID uuid.UUID) (*OnboardingStatus, error) {
	// Get user to check onboarding_completed flag
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Get interest count
	count, err := s.userLearningInterestRepo.CountByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Get interest summaries
	summaries, err := s.userLearningInterestRepo.GetInterestSummaries(ctx, userID)
	if err != nil {
		return nil, err
	}

	return &OnboardingStatus{
		OnboardingCompleted: user.OnboardingCompleted,
		InterestsCount:      count,
		Interests:           summaries,
	}, nil
}

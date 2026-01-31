// Package application provides the recommendation service for personalized feeds.
package application

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"ngasihtau/internal/common/errors"
	"ngasihtau/internal/pod/domain"
)

// RecommendationService defines the interface for recommendation operations.
type RecommendationService interface {
	// Interaction tracking
	TrackInteraction(ctx context.Context, input TrackInteractionInput) error
	TrackView(ctx context.Context, userID, podID uuid.UUID) error
	TrackTimeSpent(ctx context.Context, userID, podID uuid.UUID, seconds int) error

	// Personalized feed
	GetPersonalizedFeed(ctx context.Context, userID uuid.UUID, page, perPage int) (*RecommendedFeedResult, error)
	GetTrendingFeed(ctx context.Context, page, perPage int) (*PodListResult, error)
	GetSimilarPods(ctx context.Context, podID uuid.UUID, limit int) ([]*domain.Pod, error)

	// User preferences
	GetUserPreferences(ctx context.Context, userID uuid.UUID) (*domain.UserPreferenceProfile, error)

	// Admin/background tasks
	RecalculatePodPopularity(ctx context.Context, podID uuid.UUID) error
	RecalculateAllPopularity(ctx context.Context) error
	ApplyScoreDecay(ctx context.Context) error
}

// TrackInteractionInput contains data for tracking an interaction.
type TrackInteractionInput struct {
	UserID          uuid.UUID                   `json:"-"`
	PodID           uuid.UUID                   `json:"pod_id" validate:"required"`
	InteractionType domain.InteractionType      `json:"interaction_type" validate:"required"`
	Metadata        *domain.InteractionMetadata `json:"metadata,omitempty"`
	SessionID       *uuid.UUID                  `json:"session_id,omitempty"`
}

// RecommendedFeedResult contains paginated recommended pods.
type RecommendedFeedResult struct {
	Pods           []*domain.RecommendedPod `json:"pods"`
	Total          int                      `json:"total"`
	Page           int                      `json:"page"`
	PerPage        int                      `json:"per_page"`
	HasMore        bool                     `json:"has_more"`
	IsPersonalized bool                     `json:"is_personalized"`
}

// recommendationService implements RecommendationService.
type recommendationService struct {
	interactionRepo    domain.InteractionRepository
	categoryScoreRepo  domain.UserCategoryScoreRepository
	tagScoreRepo       domain.UserTagScoreRepository
	popularityRepo     domain.PodPopularityRepository
	recommendationRepo domain.RecommendationRepository
	podRepo            domain.PodRepository
	config             *domain.RecommendationConfig
	logger             zerolog.Logger
}

// NewRecommendationService creates a new recommendation service.
func NewRecommendationService(
	interactionRepo domain.InteractionRepository,
	categoryScoreRepo domain.UserCategoryScoreRepository,
	tagScoreRepo domain.UserTagScoreRepository,
	popularityRepo domain.PodPopularityRepository,
	recommendationRepo domain.RecommendationRepository,
	podRepo domain.PodRepository,
	logger zerolog.Logger,
) RecommendationService {
	return &recommendationService{
		interactionRepo:    interactionRepo,
		categoryScoreRepo:  categoryScoreRepo,
		tagScoreRepo:       tagScoreRepo,
		popularityRepo:     popularityRepo,
		recommendationRepo: recommendationRepo,
		podRepo:            podRepo,
		config:             domain.DefaultRecommendationConfig(),
		logger:             logger.With().Str("service", "recommendation").Logger(),
	}
}

// TrackInteraction records a user interaction and updates preference scores.
func (s *recommendationService) TrackInteraction(ctx context.Context, input TrackInteractionInput) error {
	// Get pod to extract categories and tags
	pod, err := s.podRepo.FindByID(ctx, input.PodID)
	if err != nil {
		return err
	}
	if pod == nil {
		return errors.NotFound("pod", input.PodID.String())
	}

	// Create the interaction
	interaction := domain.NewPodInteraction(
		input.UserID,
		input.PodID,
		input.InteractionType,
		input.Metadata,
	)
	interaction.SessionID = input.SessionID

	// Store the interaction
	if err := s.interactionRepo.Create(ctx, interaction); err != nil {
		return err
	}

	// Update user preference scores asynchronously (in-line for now)
	go s.updateUserScores(context.Background(), input.UserID, pod, interaction)

	// Update pod popularity asynchronously
	go s.popularityRepo.RecalculateForPod(context.Background(), input.PodID)

	s.logger.Debug().
		Str("user_id", input.UserID.String()).
		Str("pod_id", input.PodID.String()).
		Str("type", string(input.InteractionType)).
		Float64("weight", interaction.Weight).
		Msg("Interaction tracked")

	return nil
}

// updateUserScores updates category and tag scores based on interaction.
func (s *recommendationService) updateUserScores(ctx context.Context, userID uuid.UUID, pod *domain.Pod, interaction *domain.PodInteraction) {
	// Update category scores
	for _, category := range pod.Categories {
		if err := s.categoryScoreRepo.IncrementScore(ctx, userID, category, interaction.Weight, interaction.InteractionType); err != nil {
			s.logger.Error().Err(err).
				Str("user_id", userID.String()).
				Str("category", category).
				Msg("Failed to update category score")
		}
	}

	// Update tag scores
	if len(pod.Tags) > 0 {
		if err := s.tagScoreRepo.IncrementScoreBatch(ctx, userID, pod.Tags, interaction.Weight); err != nil {
			s.logger.Error().Err(err).
				Str("user_id", userID.String()).
				Msg("Failed to update tag scores")
		}
	}
}

// TrackView is a convenience method for tracking a view interaction.
func (s *recommendationService) TrackView(ctx context.Context, userID, podID uuid.UUID) error {
	return s.TrackInteraction(ctx, TrackInteractionInput{
		UserID:          userID,
		PodID:           podID,
		InteractionType: domain.InteractionView,
	})
}

// TrackTimeSpent tracks time spent on a pod.
func (s *recommendationService) TrackTimeSpent(ctx context.Context, userID, podID uuid.UUID, seconds int) error {
	return s.TrackInteraction(ctx, TrackInteractionInput{
		UserID:          userID,
		PodID:           podID,
		InteractionType: domain.InteractionTimeSpent,
		Metadata: &domain.InteractionMetadata{
			TimeSpentSeconds: seconds,
		},
	})
}

// GetPersonalizedFeed returns a personalized feed for the user.
func (s *recommendationService) GetPersonalizedFeed(ctx context.Context, userID uuid.UUID, page, perPage int) (*RecommendedFeedResult, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 50 {
		perPage = 20
	}

	offset := (page - 1) * perPage

	// Get personalized recommendations
	recommendations, err := s.recommendationRepo.GetPersonalizedFeed(ctx, userID, s.config, perPage+1, offset)
	if err != nil {
		return nil, err
	}

	// Check if user has enough data for personalization
	profile, _ := s.recommendationRepo.GetUserPreferenceProfile(ctx, userID)
	isPersonalized := profile != nil && profile.HasEnoughData

	// Check if there are more results
	hasMore := len(recommendations) > perPage
	if hasMore {
		recommendations = recommendations[:perPage]
	}

	return &RecommendedFeedResult{
		Pods:           recommendations,
		Total:          0, // We don't know exact total without counting
		Page:           page,
		PerPage:        perPage,
		HasMore:        hasMore,
		IsPersonalized: isPersonalized,
	}, nil
}

// GetTrendingFeed returns trending pods.
func (s *recommendationService) GetTrendingFeed(ctx context.Context, page, perPage int) (*PodListResult, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 50 {
		perPage = 20
	}

	offset := (page - 1) * perPage

	pods, err := s.recommendationRepo.GetTrendingFeed(ctx, perPage+1, offset)
	if err != nil {
		return nil, err
	}

	hasMore := len(pods) > perPage
	if hasMore {
		pods = pods[:perPage]
	}

	return &PodListResult{
		Pods:    pods,
		Total:   0,
		Page:    page,
		PerPage: perPage,
	}, nil
}

// GetSimilarPods returns pods similar to a given pod.
func (s *recommendationService) GetSimilarPods(ctx context.Context, podID uuid.UUID, limit int) ([]*domain.Pod, error) {
	if limit < 1 || limit > 20 {
		limit = 10
	}

	return s.recommendationRepo.GetSimilarPods(ctx, podID, limit)
}

// GetUserPreferences returns the user's preference profile.
func (s *recommendationService) GetUserPreferences(ctx context.Context, userID uuid.UUID) (*domain.UserPreferenceProfile, error) {
	return s.recommendationRepo.GetUserPreferenceProfile(ctx, userID)
}

// RecalculatePodPopularity recalculates popularity for a specific pod.
func (s *recommendationService) RecalculatePodPopularity(ctx context.Context, podID uuid.UUID) error {
	return s.popularityRepo.RecalculateForPod(ctx, podID)
}

// RecalculateAllPopularity recalculates popularity for all pods.
func (s *recommendationService) RecalculateAllPopularity(ctx context.Context) error {
	s.logger.Info().Msg("Starting popularity recalculation for all pods")
	err := s.popularityRepo.RecalculateAll(ctx)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to recalculate all popularity")
		return err
	}
	s.logger.Info().Msg("Completed popularity recalculation")
	return nil
}

// ApplyScoreDecay applies time decay to old scores.
func (s *recommendationService) ApplyScoreDecay(ctx context.Context) error {
	// Apply 50% decay to scores older than half-life period
	decayThreshold := time.Now().AddDate(0, 0, -s.config.DecayHalfLifeDays)
	decayFactor := 0.5

	s.logger.Info().
		Time("threshold", decayThreshold).
		Float64("factor", decayFactor).
		Msg("Applying score decay")

	return s.categoryScoreRepo.ApplyDecay(ctx, decayFactor, decayThreshold)
}

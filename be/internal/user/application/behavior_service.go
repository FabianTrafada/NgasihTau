package application

import (
	"context"

	"github.com/google/uuid"

	"ngasihtau/internal/user/domain"
)

// BehaviorService defines the interface for user behavior data operations.
type BehaviorService interface {
	// GetUserBehavior retrieves aggregated behavior data for a user over the last 30 days.
	GetUserBehavior(ctx context.Context, userID uuid.UUID) (*domain.UserBehavior, error)
}

// BehaviorRepository defines the interface for behavior data storage operations.
type BehaviorRepository interface {
	// GetChatBehavior retrieves chat behavior metrics for a user.
	GetChatBehavior(ctx context.Context, userID uuid.UUID, days int) (*domain.ChatBehavior, error)
	// GetMaterialBehavior retrieves material interaction metrics for a user.
	GetMaterialBehavior(ctx context.Context, userID uuid.UUID, days int) (*domain.MaterialBehavior, error)
	// GetActivityBehavior retrieves activity pattern metrics for a user.
	GetActivityBehavior(ctx context.Context, userID uuid.UUID, days int) (*domain.ActivityBehavior, error)
	// GetQuizBehavior retrieves quiz performance metrics for a user.
	GetQuizBehavior(ctx context.Context, userID uuid.UUID, days int) (*domain.QuizBehavior, error)
}

// behaviorService implements BehaviorService.
type behaviorService struct {
	behaviorRepo BehaviorRepository
}

// NewBehaviorService creates a new BehaviorService.
func NewBehaviorService(behaviorRepo BehaviorRepository) BehaviorService {
	return &behaviorService{
		behaviorRepo: behaviorRepo,
	}
}

// GetUserBehavior retrieves aggregated behavior data for a user.
func (s *behaviorService) GetUserBehavior(ctx context.Context, userID uuid.UUID) (*domain.UserBehavior, error) {
	analysisPeriodDays := 30

	// Get chat behavior
	chatBehavior, err := s.behaviorRepo.GetChatBehavior(ctx, userID, analysisPeriodDays)
	if err != nil {
		// Use defaults if error
		chatBehavior = &domain.ChatBehavior{}
	}

	// Get material behavior
	materialBehavior, err := s.behaviorRepo.GetMaterialBehavior(ctx, userID, analysisPeriodDays)
	if err != nil {
		// Use defaults if error
		materialBehavior = &domain.MaterialBehavior{AvgScrollDepth: 0.5}
	}

	// Get activity behavior
	activityBehavior, err := s.behaviorRepo.GetActivityBehavior(ctx, userID, analysisPeriodDays)
	if err != nil {
		// Use defaults if error
		activityBehavior = &domain.ActivityBehavior{PeakHour: 12}
	}

	// Get quiz behavior
	quizBehavior, err := s.behaviorRepo.GetQuizBehavior(ctx, userID, analysisPeriodDays)
	if err != nil {
		// Use defaults if error (quiz tracking not yet implemented)
		quizBehavior = &domain.QuizBehavior{}
	}

	return &domain.UserBehavior{
		UserID: userID.String(),
		BehaviorData: domain.BehaviorData{
			UserID:             userID.String(),
			AnalysisPeriodDays: analysisPeriodDays,
			Chat:               *chatBehavior,
			Material:           *materialBehavior,
			Activity:           *activityBehavior,
			Quiz:               *quizBehavior,
		},
		QuizScore:       quizBehavior.AvgScore,
		PreviousPersona: "false", // TODO: Store and retrieve from user preferences
	}, nil
}

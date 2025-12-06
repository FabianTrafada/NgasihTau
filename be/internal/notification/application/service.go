package application

import (
	"context"

	"github.com/google/uuid"

	"ngasihtau/internal/notification/domain"
)

type NotificationService struct {
	notificationRepo domain.NotificationRepository
	preferenceRepo   domain.NotificationPreferenceRepository
}

func NewNotificationService(
	notificationRepo domain.NotificationRepository,
	preferenceRepo domain.NotificationPreferenceRepository,
) *NotificationService {
	return &NotificationService{
		notificationRepo: notificationRepo,
		preferenceRepo:   preferenceRepo,
	}
}

type GetNotificationsInput struct {
	UserID uuid.UUID
	Limit  int
	Offset int
}

type GetNotificationsOutput struct {
	Notifications []*domain.Notification
	Total         int
	UnreadCount   int
}

func (s *NotificationService) GetNotifications(ctx context.Context, input GetNotificationsInput) (*GetNotificationsOutput, error) {
	if input.Limit <= 0 {
		input.Limit = 20
	}
	if input.Limit > 100 {
		input.Limit = 100
	}

	notifications, total, err := s.notificationRepo.FindByUserID(ctx, input.UserID, input.Limit, input.Offset)
	if err != nil {
		return nil, err
	}

	unreadCount, err := s.notificationRepo.CountUnread(ctx, input.UserID)
	if err != nil {
		return nil, err
	}

	return &GetNotificationsOutput{
		Notifications: notifications,
		Total:         total,
		UnreadCount:   unreadCount,
	}, nil
}

func (s *NotificationService) MarkAsRead(ctx context.Context, userID, notificationID uuid.UUID) error {
	notification, err := s.notificationRepo.FindByID(ctx, notificationID)
	if err != nil {
		return err
	}

	if notification.UserID != userID {
		return ErrNotificationNotFound
	}

	return s.notificationRepo.MarkAsRead(ctx, notificationID)
}

func (s *NotificationService) MarkAllAsRead(ctx context.Context, userID uuid.UUID) error {
	return s.notificationRepo.MarkAllAsRead(ctx, userID)
}

func (s *NotificationService) GetPreferences(ctx context.Context, userID uuid.UUID) (*domain.NotificationPreference, error) {
	pref, err := s.preferenceRepo.FindByUserID(ctx, userID)
	if err != nil {
		return domain.NewDefaultNotificationPreference(userID), nil
	}
	return pref, nil
}

type UpdatePreferencesInput struct {
	UserID            uuid.UUID
	EmailPodInvite    *bool
	EmailNewMaterial  *bool
	EmailCommentReply *bool
	InAppPodInvite    *bool
	InAppNewMaterial  *bool
	InAppCommentReply *bool
}

func (s *NotificationService) UpdatePreferences(ctx context.Context, input UpdatePreferencesInput) (*domain.NotificationPreference, error) {
	pref, err := s.preferenceRepo.FindByUserID(ctx, input.UserID)
	if err != nil {
		pref = domain.NewDefaultNotificationPreference(input.UserID)
	}

	if input.EmailPodInvite != nil {
		pref.EmailPodInvite = *input.EmailPodInvite
	}
	if input.EmailNewMaterial != nil {
		pref.EmailNewMaterial = *input.EmailNewMaterial
	}
	if input.EmailCommentReply != nil {
		pref.EmailCommentReply = *input.EmailCommentReply
	}
	if input.InAppPodInvite != nil {
		pref.InAppPodInvite = *input.InAppPodInvite
	}
	if input.InAppNewMaterial != nil {
		pref.InAppNewMaterial = *input.InAppNewMaterial
	}
	if input.InAppCommentReply != nil {
		pref.InAppCommentReply = *input.InAppCommentReply
	}

	if err := s.preferenceRepo.Upsert(ctx, pref); err != nil {
		return nil, err
	}

	return pref, nil
}

func (s *NotificationService) CreateNotification(ctx context.Context, notification *domain.Notification) error {
	pref, _ := s.preferenceRepo.FindByUserID(ctx, notification.UserID)
	if pref != nil {
		switch notification.Type {
		case domain.NotificationTypePodInvite:
			if !pref.InAppPodInvite {
				return nil // Skip notification
			}
		case domain.NotificationTypeNewMaterial:
			if !pref.InAppNewMaterial {
				return nil
			}
		case domain.NotificationTypeCommentReply:
			if !pref.InAppCommentReply {
				return nil
			}
		}
	}

	return s.notificationRepo.Create(ctx, notification)
}

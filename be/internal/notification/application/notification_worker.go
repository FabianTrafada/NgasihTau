package application

import (
	"context"
	"fmt"

	"github.com/rs/zerolog/log"

	"ngasihtau/internal/notification/domain"
	natspkg "ngasihtau/pkg/nats"
)

// NotificationWorker handles NATS events and creates in-app notifications.
type NotificationWorker struct {
	natsClient    *natspkg.Client
	service       *NotificationService
	subscriptions []*natspkg.Subscription
}

// NewNotificationWorker creates a new notification worker.
func NewNotificationWorker(
	natsClient *natspkg.Client,
	service *NotificationService,
) *NotificationWorker {
	return &NotificationWorker{
		natsClient:    natsClient,
		service:       service,
		subscriptions: make([]*natspkg.Subscription, 0),
	}
}

// Start starts the notification worker and subscribes to events.
func (w *NotificationWorker) Start(ctx context.Context) error {
	// Ensure streams exist
	if err := w.natsClient.EnsureStream(ctx, natspkg.StreamCollaborator, natspkg.StreamSubjects[natspkg.StreamCollaborator]); err != nil {
		return fmt.Errorf("failed to ensure Collaborator stream: %w", err)
	}
	if err := w.natsClient.EnsureStream(ctx, natspkg.StreamMaterial, natspkg.StreamSubjects[natspkg.StreamMaterial]); err != nil {
		return fmt.Errorf("failed to ensure Material stream: %w", err)
	}
	if err := w.natsClient.EnsureStream(ctx, natspkg.StreamComment, natspkg.StreamSubjects[natspkg.StreamComment]); err != nil {
		return fmt.Errorf("failed to ensure Comment stream: %w", err)
	}

	// Subscribe to collaborator.invited events
	collaboratorSub, err := w.natsClient.SubscribeSimple(
		ctx,
		natspkg.DefaultSubscribeConfig(
			natspkg.StreamCollaborator,
			"notification-collaborator-invited",
			[]string{natspkg.SubjectCollaboratorInvited},
		),
		w.handleCollaboratorInvited,
	)
	if err != nil {
		return fmt.Errorf("failed to subscribe to collaborator.invited events: %w", err)
	}
	w.subscriptions = append(w.subscriptions, collaboratorSub)
	log.Info().Msg("subscribed to collaborator.invited events")

	// Subscribe to material.processed events
	materialSub, err := w.natsClient.SubscribeSimple(
		ctx,
		natspkg.DefaultSubscribeConfig(
			natspkg.StreamMaterial,
			"notification-material-processed",
			[]string{natspkg.SubjectMaterialProcessed},
		),
		w.handleMaterialProcessed,
	)
	if err != nil {
		return fmt.Errorf("failed to subscribe to material.processed events: %w", err)
	}
	w.subscriptions = append(w.subscriptions, materialSub)
	log.Info().Msg("subscribed to material.processed events")

	// Subscribe to comment.created events
	commentSub, err := w.natsClient.SubscribeSimple(
		ctx,
		natspkg.DefaultSubscribeConfig(
			natspkg.StreamComment,
			"notification-comment-created",
			[]string{natspkg.SubjectCommentCreated},
		),
		w.handleCommentCreated,
	)
	if err != nil {
		return fmt.Errorf("failed to subscribe to comment.created events: %w", err)
	}
	w.subscriptions = append(w.subscriptions, commentSub)
	log.Info().Msg("subscribed to comment.created events")

	log.Info().Msg("notification worker started successfully")
	return nil
}

// Stop stops the notification worker.
func (w *NotificationWorker) Stop() {
	for _, sub := range w.subscriptions {
		sub.Stop()
	}
	log.Info().Msg("notification worker stopped")
}

// handleCollaboratorInvited handles collaborator.invited events.
func (w *NotificationWorker) handleCollaboratorInvited(ctx context.Context, event natspkg.CloudEvent) error {
	data, err := natspkg.ParseEventData[natspkg.CollaboratorInvitedEvent](event)
	if err != nil {
		log.Error().Err(err).Str("event_id", event.ID).Msg("failed to parse collaborator.invited event")
		return err
	}

	log.Info().
		Str("pod_id", data.PodID.String()).
		Str("invitee_id", data.InviteeID.String()).
		Str("inviter_id", data.InviterID.String()).
		Msg("processing collaborator.invited event")

	notification := domain.NewNotification(
		data.InviteeID,
		domain.NotificationTypePodInvite,
		"You've been invited to collaborate",
		fmt.Sprintf("You have been invited to collaborate on a pod with role: %s", data.Role),
		&domain.NotificationData{
			PodID:     &data.PodID,
			UserID:    &data.InviterID,
			ActionURL: fmt.Sprintf("/pods/%s", data.PodID.String()),
		},
	)

	if err := w.service.CreateNotification(ctx, notification); err != nil {
		log.Error().Err(err).
			Str("invitee_id", data.InviteeID.String()).
			Msg("failed to create collaborator invite notification")
		return err
	}

	log.Info().
		Str("invitee_id", data.InviteeID.String()).
		Msg("collaborator invite notification created")
	return nil
}

// handleMaterialProcessed handles material.processed events.
func (w *NotificationWorker) handleMaterialProcessed(ctx context.Context, event natspkg.CloudEvent) error {
	data, err := natspkg.ParseEventData[natspkg.MaterialProcessedEvent](event)
	if err != nil {
		log.Error().Err(err).Str("event_id", event.ID).Msg("failed to parse material.processed event")
		return err
	}

	log.Info().
		Str("material_id", data.MaterialID.String()).
		Str("pod_id", data.PodID.String()).
		Str("status", data.Status).
		Msg("processing material.processed event")

	// Only notify if processing was successful
	if data.Status != "ready" {
		log.Info().
			Str("material_id", data.MaterialID.String()).
			Str("status", data.Status).
			Msg("skipping notification for non-ready material")
		return nil
	}

	// Notify the uploader that their material is ready
	notification := domain.NewNotification(
		data.UploaderID,
		domain.NotificationTypeMaterialProcessed,
		"Your material is ready",
		fmt.Sprintf("Your material '%s' has been processed and is now ready for AI chat", data.Title),
		&domain.NotificationData{
			MaterialID:    &data.MaterialID,
			MaterialTitle: data.Title,
			PodID:         &data.PodID,
			ActionURL:     fmt.Sprintf("/materials/%s", data.MaterialID.String()),
		},
	)

	if err := w.service.CreateNotification(ctx, notification); err != nil {
		log.Error().Err(err).
			Str("uploader_id", data.UploaderID.String()).
			Msg("failed to create material processed notification")
		return err
	}

	log.Info().
		Str("uploader_id", data.UploaderID.String()).
		Str("material_id", data.MaterialID.String()).
		Msg("material processed notification created")
	return nil
}

// handleCommentCreated handles comment.created events.
func (w *NotificationWorker) handleCommentCreated(ctx context.Context, event natspkg.CloudEvent) error {
	data, err := natspkg.ParseEventData[natspkg.CommentCreatedEvent](event)
	if err != nil {
		log.Error().Err(err).Str("event_id", event.ID).Msg("failed to parse comment.created event")
		return err
	}

	log.Info().
		Str("comment_id", data.CommentID.String()).
		Str("material_id", data.MaterialID.String()).
		Str("user_id", data.UserID.String()).
		Msg("processing comment.created event")

	// Only notify if this is a reply to another comment
	if data.ParentID == nil || data.ParentAuthorID == nil {
		log.Info().
			Str("comment_id", data.CommentID.String()).
			Msg("skipping notification for top-level comment")
		return nil
	}

	if *data.ParentAuthorID == data.UserID {
		log.Info().
			Str("comment_id", data.CommentID.String()).
			Msg("skipping self-reply notification")
		return nil
	}

	notification := domain.NewNotification(
		*data.ParentAuthorID,
		domain.NotificationTypeCommentReply,
		"Someone replied to your comment",
		truncateContent(data.Content, 100),
		&domain.NotificationData{
			MaterialID: &data.MaterialID,
			CommentID:  &data.CommentID,
			UserID:     &data.UserID,
			ActionURL:  fmt.Sprintf("/materials/%s#comment-%s", data.MaterialID.String(), data.CommentID.String()),
		},
	)

	if err := w.service.CreateNotification(ctx, notification); err != nil {
		log.Error().Err(err).
			Str("comment_id", data.CommentID.String()).
			Msg("failed to create comment reply notification")
		return err
	}

	log.Info().
		Str("comment_id", data.CommentID.String()).
		Msg("comment reply notification created")
	return nil
}

// truncateContent truncates content to the specified length.
func truncateContent(content string, maxLen int) string {
	if len(content) <= maxLen {
		return content
	}
	return content[:maxLen-3] + "..."
}

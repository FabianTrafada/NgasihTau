// Package application contains the business logic for the pod service.
package application

import (
	"context"

	"ngasihtau/internal/pod/domain"
	natspkg "ngasihtau/pkg/nats"

	"github.com/google/uuid"
)

// NATSEventPublisher implements EventPublisher using NATS.
type NATSEventPublisher struct {
	publisher *natspkg.Publisher
}

// NewNATSEventPublisher creates a new NATS event publisher.
func NewNATSEventPublisher(publisher *natspkg.Publisher) *NATSEventPublisher {
	return &NATSEventPublisher{
		publisher: publisher,
	}
}

// PublishPodCreated publishes a pod.created event.
func (p *NATSEventPublisher) PublishPodCreated(ctx context.Context, pod *domain.Pod) error {
	description := ""
	if pod.Description != nil {
		description = *pod.Description
	}

	event := natspkg.PodCreatedEvent{
		PodID:       pod.ID,
		OwnerID:     pod.OwnerID,
		Name:        pod.Name,
		Slug:        pod.Slug,
		Description: description,
		Visibility:  string(pod.Visibility),
		Categories:  pod.Categories,
		Tags:        pod.Tags,
		IsVerified:  pod.IsVerified,  // True if created by teacher (Requirement 6.1)
		UpvoteCount: pod.UpvoteCount, // Trust indicator, always 0 for new pods (Requirement 6.2)
	}

	return p.publisher.PublishPodCreated(ctx, event)
}

// PublishPodUpdated publishes a pod.updated event.
func (p *NATSEventPublisher) PublishPodUpdated(ctx context.Context, pod *domain.Pod) error {
	description := ""
	if pod.Description != nil {
		description = *pod.Description
	}

	event := natspkg.PodUpdatedEvent{
		PodID:       pod.ID,
		OwnerID:     pod.OwnerID,
		Name:        pod.Name,
		Slug:        pod.Slug,
		Description: description,
		Visibility:  string(pod.Visibility),
		Categories:  pod.Categories,
		Tags:        pod.Tags,
		StarCount:   pod.StarCount,
		ForkCount:   pod.ForkCount,
		ViewCount:   pod.ViewCount,
		IsVerified:  pod.IsVerified,  // True if created by teacher (Requirement 6.1)
		UpvoteCount: pod.UpvoteCount, // Trust indicator (Requirement 6.2)
	}

	return p.publisher.PublishPodUpdated(ctx, event)
}

// PublishCollaboratorInvited publishes a collaborator.invited event.
func (p *NATSEventPublisher) PublishCollaboratorInvited(ctx context.Context, collaborator *domain.Collaborator, podName string) error {
	event := natspkg.CollaboratorInvitedEvent{
		PodID:     collaborator.PodID,
		InviterID: collaborator.InvitedBy,
		InviteeID: collaborator.UserID,
		Role:      string(collaborator.Role),
	}

	return p.publisher.PublishCollaboratorInvited(ctx, event)
}

// PublishUploadRequestCreated publishes an upload request created event for notification.
// Implements requirement 4.2: WHEN a pod owner receives an upload request, THE Notification Service SHALL send a notification.
func (p *NATSEventPublisher) PublishUploadRequestCreated(ctx context.Context, request *domain.UploadRequest, podName string, requesterName string) error {
	message := ""
	if request.Message != nil {
		message = *request.Message
	}

	event := natspkg.UploadRequestEvent{
		RequestID:     request.ID,
		RequesterID:   request.RequesterID,
		RequesterName: requesterName,
		PodID:         request.PodID,
		PodName:       podName,
		PodOwnerID:    request.PodOwnerID,
		Status:        string(request.Status),
		Message:       message,
	}

	return p.publisher.PublishUploadRequest(ctx, event)
}

// PublishUploadRequestApproved publishes an upload request approved event for notification.
// Implements requirement 4.3: WHEN a pod owner approves an upload request.
func (p *NATSEventPublisher) PublishUploadRequestApproved(ctx context.Context, request *domain.UploadRequest, podName string, requesterName string) error {
	event := natspkg.UploadRequestEvent{
		RequestID:     request.ID,
		RequesterID:   request.RequesterID,
		RequesterName: requesterName,
		PodID:         request.PodID,
		PodName:       podName,
		PodOwnerID:    request.PodOwnerID,
		Status:        string(domain.UploadRequestStatusApproved),
	}

	return p.publisher.PublishUploadRequest(ctx, event)
}

// PublishUploadRequestRejected publishes an upload request rejected event for notification.
// Implements requirement 4.4: WHEN a pod owner rejects an upload request, THE Pod Service SHALL notify the requesting teacher.
func (p *NATSEventPublisher) PublishUploadRequestRejected(ctx context.Context, request *domain.UploadRequest, podName string, reason *string) error {
	rejectionReason := ""
	if reason != nil {
		rejectionReason = *reason
	}

	event := natspkg.UploadRequestEvent{
		RequestID:       request.ID,
		RequesterID:     request.RequesterID,
		PodID:           request.PodID,
		PodName:         podName,
		PodOwnerID:      request.PodOwnerID,
		Status:          string(domain.UploadRequestStatusRejected),
		RejectionReason: rejectionReason,
	}

	return p.publisher.PublishUploadRequest(ctx, event)
}

// PublishPodShared publishes a pod shared event for notification.
// Implements requirement 7.3: WHEN a teacher shares a pod with a student, THE Notification Service SHALL notify the student.
func (p *NATSEventPublisher) PublishPodShared(ctx context.Context, sharedPod *domain.SharedPod, podName string, teacherName string) error {
	message := ""
	if sharedPod.Message != nil {
		message = *sharedPod.Message
	}

	event := natspkg.PodSharedEvent{
		ShareID:     sharedPod.ID,
		PodID:       sharedPod.PodID,
		PodName:     podName,
		TeacherID:   sharedPod.TeacherID,
		TeacherName: teacherName,
		StudentID:   sharedPod.StudentID,
		Message:     message,
	}

	return p.publisher.PublishPodShared(ctx, event)
}

// PublishPodUpvoted publishes a pod upvoted event.
// Consumed by Search Service for re-indexing upvote count.
func (p *NATSEventPublisher) PublishPodUpvoted(ctx context.Context, podID uuid.UUID, userID uuid.UUID, upvoteCount int, isUpvote bool) error {
	event := natspkg.PodUpvotedEvent{
		PodID:       podID,
		UserID:      userID,
		UpvoteCount: upvoteCount,
		IsUpvote:    isUpvote,
	}

	return p.publisher.PublishPodUpvoted(ctx, event)
}

// NoOpEventPublisher is a no-op implementation of EventPublisher.
type NoOpEventPublisher struct{}

// NewNoOpEventPublisher creates a new no-op event publisher.
func NewNoOpEventPublisher() *NoOpEventPublisher {
	return &NoOpEventPublisher{}
}

// PublishPodCreated is a no-op.
func (p *NoOpEventPublisher) PublishPodCreated(ctx context.Context, pod *domain.Pod) error {
	return nil
}

// PublishPodUpdated is a no-op.
func (p *NoOpEventPublisher) PublishPodUpdated(ctx context.Context, pod *domain.Pod) error {
	return nil
}

// PublishCollaboratorInvited is a no-op.
func (p *NoOpEventPublisher) PublishCollaboratorInvited(ctx context.Context, collaborator *domain.Collaborator, podName string) error {
	return nil
}

// PublishUploadRequestCreated is a no-op.
func (p *NoOpEventPublisher) PublishUploadRequestCreated(ctx context.Context, request *domain.UploadRequest, podName string, requesterName string) error {
	return nil
}

// PublishUploadRequestApproved is a no-op.
func (p *NoOpEventPublisher) PublishUploadRequestApproved(ctx context.Context, request *domain.UploadRequest, podName string, requesterName string) error {
	return nil
}

// PublishUploadRequestRejected is a no-op.
func (p *NoOpEventPublisher) PublishUploadRequestRejected(ctx context.Context, request *domain.UploadRequest, podName string, reason *string) error {
	return nil
}

// PublishPodShared is a no-op.
func (p *NoOpEventPublisher) PublishPodShared(ctx context.Context, sharedPod *domain.SharedPod, podName string, teacherName string) error {
	return nil
}

// PublishPodUpvoted is a no-op.
func (p *NoOpEventPublisher) PublishPodUpvoted(ctx context.Context, podID uuid.UUID, userID uuid.UUID, upvoteCount int, isUpvote bool) error {
	return nil
}

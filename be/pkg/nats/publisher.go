// Package nats provides event publishing functionality.
package nats

import (
	"context"

	"github.com/google/uuid"
)

// EventPublisher defines the interface for publishing events.
// This interface allows for dependency injection and easier testing.
type EventPublisher interface {
	// PublishEmailVerification publishes an email verification event.
	PublishEmailVerification(ctx context.Context, event EmailVerificationEvent) error

	// PublishEmailPasswordReset publishes a password reset email event.
	PublishEmailPasswordReset(ctx context.Context, event EmailPasswordResetEvent) error

	// PublishUserCreated publishes a user created event.
	PublishUserCreated(ctx context.Context, event UserCreatedEvent) error

	// PublishUserUpdated publishes a user updated event.
	PublishUserUpdated(ctx context.Context, userID uuid.UUID) error

	// PublishUserFollowed publishes a user followed event.
	PublishUserFollowed(ctx context.Context, followerID, followingID uuid.UUID) error

	// PublishMaterialUploaded publishes a material uploaded event.
	PublishMaterialUploaded(ctx context.Context, event MaterialUploadedEvent) error

	// PublishMaterialDeleted publishes a material deleted event.
	PublishMaterialDeleted(ctx context.Context, event MaterialDeletedEvent) error

	// PublishPodCreated publishes a pod created event.
	PublishPodCreated(ctx context.Context, event PodCreatedEvent) error

	// PublishPodUpdated publishes a pod updated event.
	PublishPodUpdated(ctx context.Context, event PodUpdatedEvent) error

	// PublishCollaboratorInvited publishes a collaborator invited event.
	PublishCollaboratorInvited(ctx context.Context, event CollaboratorInvitedEvent) error

	// PublishCommentCreated publishes a comment created event.
	PublishCommentCreated(ctx context.Context, event CommentCreatedEvent) error

	// PublishTeacherVerified publishes a teacher verified event.
	PublishTeacherVerified(ctx context.Context, event TeacherVerifiedEvent) error

	// PublishUploadRequest publishes an upload request event.
	PublishUploadRequest(ctx context.Context, event UploadRequestEvent) error

	// PublishPodShared publishes a pod shared event.
	PublishPodShared(ctx context.Context, event PodSharedEvent) error

	// PublishPodUpvoted publishes a pod upvoted event.
	PublishPodUpvoted(ctx context.Context, event PodUpvotedEvent) error
}

// Publisher implements EventPublisher using NATS JetStream.
type Publisher struct {
	client *Client
}

// NewPublisher creates a new Publisher.
func NewPublisher(client *Client) *Publisher {
	return &Publisher{client: client}
}

// PublishEmailVerification publishes an email verification event.
func (p *Publisher) PublishEmailVerification(ctx context.Context, event EmailVerificationEvent) error {
	return p.client.Publish(ctx, SubjectEmailVerification, "email.verification", event)
}

// PublishEmailPasswordReset publishes a password reset email event.
func (p *Publisher) PublishEmailPasswordReset(ctx context.Context, event EmailPasswordResetEvent) error {
	return p.client.Publish(ctx, SubjectEmailPasswordReset, "email.password_reset", event)
}

// PublishUserCreated publishes a user created event.
func (p *Publisher) PublishUserCreated(ctx context.Context, event UserCreatedEvent) error {
	return p.client.Publish(ctx, SubjectUserCreated, "user.created", event)
}

// PublishUserUpdated publishes a user updated event.
func (p *Publisher) PublishUserUpdated(ctx context.Context, userID uuid.UUID) error {
	event := UserUpdatedEvent{UserID: userID}
	return p.client.Publish(ctx, SubjectUserUpdated, "user.updated", event)
}

// PublishUserFollowed publishes a user followed event.
func (p *Publisher) PublishUserFollowed(ctx context.Context, followerID, followingID uuid.UUID) error {
	event := UserFollowedEvent{
		FollowerID:  followerID,
		FollowingID: followingID,
	}
	return p.client.Publish(ctx, SubjectUserFollowed, "user.followed", event)
}

// PublishMaterialUploaded publishes a material uploaded event.
func (p *Publisher) PublishMaterialUploaded(ctx context.Context, event MaterialUploadedEvent) error {
	return p.client.Publish(ctx, SubjectMaterialUploaded, "material.uploaded", event)
}

// PublishMaterialDeleted publishes a material deleted event.
func (p *Publisher) PublishMaterialDeleted(ctx context.Context, event MaterialDeletedEvent) error {
	return p.client.Publish(ctx, SubjectMaterialDeleted, "material.deleted", event)
}

// PublishPodCreated publishes a pod created event.
func (p *Publisher) PublishPodCreated(ctx context.Context, event PodCreatedEvent) error {
	return p.client.Publish(ctx, SubjectPodCreated, "pod.created", event)
}

// PublishPodUpdated publishes a pod updated event.
func (p *Publisher) PublishPodUpdated(ctx context.Context, event PodUpdatedEvent) error {
	return p.client.Publish(ctx, SubjectPodUpdated, "pod.updated", event)
}

// PublishCollaboratorInvited publishes a collaborator invited event.
func (p *Publisher) PublishCollaboratorInvited(ctx context.Context, event CollaboratorInvitedEvent) error {
	return p.client.Publish(ctx, SubjectCollaboratorInvited, "collaborator.invited", event)
}

// PublishCommentCreated publishes a comment created event.
func (p *Publisher) PublishCommentCreated(ctx context.Context, event CommentCreatedEvent) error {
	return p.client.Publish(ctx, SubjectCommentCreated, "comment.created", event)
}

// PublishTeacherVerified publishes a teacher verified event.
func (p *Publisher) PublishTeacherVerified(ctx context.Context, event TeacherVerifiedEvent) error {
	return p.client.Publish(ctx, SubjectTeacherVerified, "user.teacher_verified", event)
}

// PublishUploadRequest publishes an upload request event.
func (p *Publisher) PublishUploadRequest(ctx context.Context, event UploadRequestEvent) error {
	return p.client.Publish(ctx, SubjectUploadRequest, "pod.upload_request", event)
}

// PublishPodShared publishes a pod shared event.
func (p *Publisher) PublishPodShared(ctx context.Context, event PodSharedEvent) error {
	return p.client.Publish(ctx, SubjectPodShared, "pod.shared", event)
}

// PublishPodUpvoted publishes a pod upvoted event.
func (p *Publisher) PublishPodUpvoted(ctx context.Context, event PodUpvotedEvent) error {
	return p.client.Publish(ctx, SubjectPodUpvoted, "pod.upvoted", event)
}

// NoOpPublisher is a no-op implementation of EventPublisher for when NATS is not configured.
type NoOpPublisher struct{}

// NewNoOpPublisher creates a new NoOpPublisher.
func NewNoOpPublisher() *NoOpPublisher {
	return &NoOpPublisher{}
}

// PublishEmailVerification is a no-op.
func (p *NoOpPublisher) PublishEmailVerification(ctx context.Context, event EmailVerificationEvent) error {
	return nil
}

// PublishEmailPasswordReset is a no-op.
func (p *NoOpPublisher) PublishEmailPasswordReset(ctx context.Context, event EmailPasswordResetEvent) error {
	return nil
}

// PublishUserCreated is a no-op.
func (p *NoOpPublisher) PublishUserCreated(ctx context.Context, event UserCreatedEvent) error {
	return nil
}

// PublishUserUpdated is a no-op.
func (p *NoOpPublisher) PublishUserUpdated(ctx context.Context, userID uuid.UUID) error {
	return nil
}

// PublishUserFollowed is a no-op.
func (p *NoOpPublisher) PublishUserFollowed(ctx context.Context, followerID, followingID uuid.UUID) error {
	return nil
}

// PublishMaterialUploaded is a no-op.
func (p *NoOpPublisher) PublishMaterialUploaded(ctx context.Context, event MaterialUploadedEvent) error {
	return nil
}

// PublishMaterialDeleted is a no-op.
func (p *NoOpPublisher) PublishMaterialDeleted(ctx context.Context, event MaterialDeletedEvent) error {
	return nil
}

// PublishPodCreated is a no-op.
func (p *NoOpPublisher) PublishPodCreated(ctx context.Context, event PodCreatedEvent) error {
	return nil
}

// PublishPodUpdated is a no-op.
func (p *NoOpPublisher) PublishPodUpdated(ctx context.Context, event PodUpdatedEvent) error {
	return nil
}

// PublishCollaboratorInvited is a no-op.
func (p *NoOpPublisher) PublishCollaboratorInvited(ctx context.Context, event CollaboratorInvitedEvent) error {
	return nil
}

// PublishCommentCreated is a no-op.
func (p *NoOpPublisher) PublishCommentCreated(ctx context.Context, event CommentCreatedEvent) error {
	return nil
}

// PublishTeacherVerified is a no-op.
func (p *NoOpPublisher) PublishTeacherVerified(ctx context.Context, event TeacherVerifiedEvent) error {
	return nil
}

// PublishUploadRequest is a no-op.
func (p *NoOpPublisher) PublishUploadRequest(ctx context.Context, event UploadRequestEvent) error {
	return nil
}

// PublishPodShared is a no-op.
func (p *NoOpPublisher) PublishPodShared(ctx context.Context, event PodSharedEvent) error {
	return nil
}

// PublishPodUpvoted is a no-op.
func (p *NoOpPublisher) PublishPodUpvoted(ctx context.Context, event PodUpvotedEvent) error {
	return nil
}

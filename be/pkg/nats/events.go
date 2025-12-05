// Package nats provides event types and subjects for NATS messaging.
package nats

import "github.com/google/uuid"

// Stream names for JetStream.
const (
	// StreamEmail is the stream for email-related events.
	StreamEmail = "EMAIL"
	// StreamUser is the stream for user-related events.
	StreamUser = "USER"
	// StreamMaterial is the stream for material-related events.
	StreamMaterial = "MATERIAL"
	// StreamPod is the stream for pod-related events.
	StreamPod = "POD"
	// StreamCollaborator is the stream for collaborator-related events.
	StreamCollaborator = "COLLABORATOR"
	// StreamComment is the stream for comment-related events.
	StreamComment = "COMMENT"
)

// Event subjects for NATS JetStream.
const (
	// Email events - consumed by Notification Service
	SubjectEmailVerification       = "email.verification"
	SubjectEmailPasswordReset      = "email.password_reset"
	SubjectEmailCollaboratorInvite = "email.collaborator_invite"

	// User events
	SubjectUserCreated  = "user.created"
	SubjectUserUpdated  = "user.updated"
	SubjectUserFollowed = "user.followed"

	// Material events
	SubjectMaterialUploaded  = "material.uploaded"
	SubjectMaterialProcessed = "material.processed"
	SubjectMaterialDeleted   = "material.deleted"

	// Pod events
	SubjectPodCreated = "pod.created"
	SubjectPodUpdated = "pod.updated"

	// Collaborator events
	SubjectCollaboratorInvited = "collaborator.invited"

	// Comment events
	SubjectCommentCreated = "comment.created"
)

// StreamSubjects maps stream names to their subjects.
var StreamSubjects = map[string][]string{
	StreamEmail:        {SubjectEmailVerification, SubjectEmailPasswordReset, SubjectEmailCollaboratorInvite},
	StreamUser:         {SubjectUserCreated, SubjectUserUpdated, SubjectUserFollowed},
	StreamMaterial:     {SubjectMaterialUploaded, SubjectMaterialProcessed, SubjectMaterialDeleted},
	StreamPod:          {SubjectPodCreated, SubjectPodUpdated},
	StreamCollaborator: {SubjectCollaboratorInvited},
	StreamComment:      {SubjectCommentCreated},
}

// EmailVerificationEvent is published when a user needs to verify their email.
// Consumed by Notification Service to send verification email.
type EmailVerificationEvent struct {
	UserID    uuid.UUID `json:"user_id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	Token     string    `json:"token"`
	ExpiresAt string    `json:"expires_at"` // ISO 8601 format
}

// EmailPasswordResetEvent is published when a user requests a password reset.
// Consumed by Notification Service to send password reset email.
type EmailPasswordResetEvent struct {
	UserID    uuid.UUID `json:"user_id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	Token     string    `json:"token"`
	ExpiresAt string    `json:"expires_at"` // ISO 8601 format
}

// EmailCollaboratorInviteEvent is published when a user is invited to a pod.
// Consumed by Notification Service to send invitation email.
type EmailCollaboratorInviteEvent struct {
	InviterID   uuid.UUID `json:"inviter_id"`
	InviterName string    `json:"inviter_name"`
	InviteeID   uuid.UUID `json:"invitee_id"`
	Email       string    `json:"email"`
	PodID       uuid.UUID `json:"pod_id"`
	PodName     string    `json:"pod_name"`
	Role        string    `json:"role"`
}

// UserCreatedEvent is published when a new user is created.
type UserCreatedEvent struct {
	UserID uuid.UUID `json:"user_id"`
	Email  string    `json:"email"`
	Name   string    `json:"name"`
	Role   string    `json:"role"`
}

// UserUpdatedEvent is published when a user profile is updated.
type UserUpdatedEvent struct {
	UserID uuid.UUID `json:"user_id"`
}

// UserFollowedEvent is published when a user follows another user.
type UserFollowedEvent struct {
	FollowerID  uuid.UUID `json:"follower_id"`
	FollowingID uuid.UUID `json:"following_id"`
}

// MaterialUploadedEvent is published when a material is uploaded.
// Consumed by AI Service for processing and embedding generation.
type MaterialUploadedEvent struct {
	MaterialID  uuid.UUID `json:"material_id"`
	PodID       uuid.UUID `json:"pod_id"`
	FileURL     string    `json:"file_url"`
	FileType    string    `json:"file_type"`
	UploaderID  uuid.UUID `json:"uploader_id"`
	Title       string    `json:"title"`
	Description string    `json:"description,omitempty"`
	// Inherited from pod for search indexing
	Categories []string `json:"categories,omitempty"`
	Tags       []string `json:"tags,omitempty"`
}

// MaterialProcessedEvent is published when a material has been processed.
// Consumed by Search Service for indexing and Material Service for status update.
type MaterialProcessedEvent struct {
	MaterialID  uuid.UUID `json:"material_id"`
	PodID       uuid.UUID `json:"pod_id"`
	UploaderID  uuid.UUID `json:"uploader_id"`
	Title       string    `json:"title"`
	Description string    `json:"description,omitempty"`
	FileType    string    `json:"file_type"`
	Status      string    `json:"status"` // "ready" or "processing_failed"
	ChunkCount  int       `json:"chunk_count,omitempty"`
	WordCount   int       `json:"word_count,omitempty"`
	Error       string    `json:"error,omitempty"`
	// Inherited from pod for search indexing
	Categories []string `json:"categories,omitempty"`
	Tags       []string `json:"tags,omitempty"`
}

// MaterialDeletedEvent is published when a material is deleted.
type MaterialDeletedEvent struct {
	MaterialID uuid.UUID `json:"material_id"`
	PodID      uuid.UUID `json:"pod_id"`
}

// PodCreatedEvent is published when a pod is created.
// Consumed by Search Service for indexing.
type PodCreatedEvent struct {
	PodID       uuid.UUID `json:"pod_id"`
	OwnerID     uuid.UUID `json:"owner_id"`
	Name        string    `json:"name"`
	Slug        string    `json:"slug"`
	Description string    `json:"description,omitempty"`
	Visibility  string    `json:"visibility"`
	Categories  []string  `json:"categories,omitempty"`
	Tags        []string  `json:"tags,omitempty"`
}

// PodUpdatedEvent is published when a pod is updated.
// Consumed by Search Service for re-indexing.
type PodUpdatedEvent struct {
	PodID       uuid.UUID `json:"pod_id"`
	OwnerID     uuid.UUID `json:"owner_id"`
	Name        string    `json:"name"`
	Slug        string    `json:"slug"`
	Description string    `json:"description,omitempty"`
	Visibility  string    `json:"visibility"`
	Categories  []string  `json:"categories,omitempty"`
	Tags        []string  `json:"tags,omitempty"`
	StarCount   int       `json:"star_count"`
	ForkCount   int       `json:"fork_count"`
	ViewCount   int       `json:"view_count"`
}

// CollaboratorInvitedEvent is published when a collaborator is invited.
type CollaboratorInvitedEvent struct {
	PodID     uuid.UUID `json:"pod_id"`
	InviterID uuid.UUID `json:"inviter_id"`
	InviteeID uuid.UUID `json:"invitee_id"`
	Role      string    `json:"role"`
}

// CommentCreatedEvent is published when a comment is created.
type CommentCreatedEvent struct {
	CommentID  uuid.UUID  `json:"comment_id"`
	MaterialID uuid.UUID  `json:"material_id"`
	UserID     uuid.UUID  `json:"user_id"`
	ParentID   *uuid.UUID `json:"parent_id,omitempty"`
	Content    string     `json:"content"`
}

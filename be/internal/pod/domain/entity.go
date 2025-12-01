// Package domain contains the core business entities and repository interfaces
// for the Pod Service. This layer is independent of external frameworks and databases.
package domain

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Visibility represents the visibility level of a Knowledge Pod.
type Visibility string

const (
	// VisibilityPublic means the pod is visible to everyone.
	VisibilityPublic Visibility = "public"
	// VisibilityPrivate means the pod is only visible to owner and collaborators.
	VisibilityPrivate Visibility = "private"
)

// CollaboratorRole represents a collaborator's role in a Knowledge Pod.
type CollaboratorRole string

const (
	// CollaboratorRoleViewer can only view the pod content.
	CollaboratorRoleViewer CollaboratorRole = "viewer"
	// CollaboratorRoleContributor can upload materials to the pod.
	CollaboratorRoleContributor CollaboratorRole = "contributor"
	// CollaboratorRoleAdmin can manage collaborators and pod settings.
	CollaboratorRoleAdmin CollaboratorRole = "admin"
)

// CollaboratorStatus represents the status of a collaborator invitation.
type CollaboratorStatus string

const (
	// CollaboratorStatusPending means the invitation is pending acceptance.
	CollaboratorStatusPending CollaboratorStatus = "pending"
	// CollaboratorStatusPendingVerification means accepted but awaiting owner verification.
	CollaboratorStatusPendingVerification CollaboratorStatus = "pending_verification"
	// CollaboratorStatusVerified means the collaborator is fully verified.
	CollaboratorStatusVerified CollaboratorStatus = "verified"
)

// ActivityAction represents the type of activity in a pod.
type ActivityAction string

const (
	// ActivityActionMaterialUploaded when a new material is uploaded.
	ActivityActionMaterialUploaded ActivityAction = "material_uploaded"
	// ActivityActionCollaboratorAdded when a new collaborator is added.
	ActivityActionCollaboratorAdded ActivityAction = "collaborator_added"
	// ActivityActionPodUpdated when pod details are updated.
	ActivityActionPodUpdated ActivityAction = "pod_updated"
	// ActivityActionPodForked when the pod is forked.
	ActivityActionPodForked ActivityAction = "pod_forked"
)

// Pod represents a Knowledge Pod in the system.
// Implements requirements 3, 3.1, 3.2.
type Pod struct {
	ID           uuid.UUID  `json:"id"`
	OwnerID      uuid.UUID  `json:"owner_id"`
	Name         string     `json:"name"`
	Slug         string     `json:"slug"`
	Description  *string    `json:"description,omitempty"`
	Visibility   Visibility `json:"visibility"`
	Categories   []string   `json:"categories,omitempty"`
	Tags         []string   `json:"tags,omitempty"`
	StarCount    int        `json:"star_count"`
	ForkCount    int        `json:"fork_count"`
	ViewCount    int        `json:"view_count"`
	ForkedFromID *uuid.UUID `json:"forked_from_id,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	DeletedAt    *time.Time `json:"-"`
}

// IsPublic returns true if the pod is publicly visible.
func (p *Pod) IsPublic() bool {
	return p.Visibility == VisibilityPublic
}

// IsOwner returns true if the given user ID is the pod owner.
func (p *Pod) IsOwner(userID uuid.UUID) bool {
	return p.OwnerID == userID
}

// Collaborator represents a user's collaboration on a Knowledge Pod.
// Implements requirement 4.
type Collaborator struct {
	ID        uuid.UUID          `json:"id"`
	PodID     uuid.UUID          `json:"pod_id"`
	UserID    uuid.UUID          `json:"user_id"`
	Role      CollaboratorRole   `json:"role"`
	Status    CollaboratorStatus `json:"status"`
	InvitedBy uuid.UUID          `json:"invited_by"`
	CreatedAt time.Time          `json:"created_at"`
	UpdatedAt time.Time          `json:"updated_at"`
}

// IsVerified returns true if the collaborator is verified.
func (c *Collaborator) IsVerified() bool {
	return c.Status == CollaboratorStatusVerified
}

// CanUpload returns true if the collaborator can upload materials.
func (c *Collaborator) CanUpload() bool {
	return c.IsVerified() && (c.Role == CollaboratorRoleContributor || c.Role == CollaboratorRoleAdmin)
}

// CanManage returns true if the collaborator can manage the pod.
func (c *Collaborator) CanManage() bool {
	return c.IsVerified() && c.Role == CollaboratorRoleAdmin
}

// PodStar represents a star on a Knowledge Pod.
// Implements requirement 3.2.
type PodStar struct {
	UserID    uuid.UUID `json:"user_id"`
	PodID     uuid.UUID `json:"pod_id"`
	CreatedAt time.Time `json:"created_at"`
}

// PodFollow represents a follow relationship with a Knowledge Pod.
// Implements requirement 12.
type PodFollow struct {
	UserID    uuid.UUID `json:"user_id"`
	PodID     uuid.UUID `json:"pod_id"`
	CreatedAt time.Time `json:"created_at"`
}

// ActivityMetadata represents additional data for an activity.
type ActivityMetadata map[string]any

// Value implements driver.Valuer for database storage.
func (m ActivityMetadata) Value() (driver.Value, error) {
	if m == nil {
		return nil, nil
	}
	return json.Marshal(m)
}

// Scan implements sql.Scanner for database retrieval.
func (m *ActivityMetadata) Scan(value interface{}) error {
	if value == nil {
		*m = nil
		return nil
	}

	var data []byte
	switch v := value.(type) { //nolint:gocritic // type switch is appropriate here
	case []byte:
		data = v
	case string:
		data = []byte(v)
	default:
		return fmt.Errorf("cannot scan type %T into ActivityMetadata", value)
	}

	return json.Unmarshal(data, m)
}

// Activity represents an activity event in a Knowledge Pod.
// Implements requirement 12.
type Activity struct {
	ID        uuid.UUID        `json:"id"`
	PodID     uuid.UUID        `json:"pod_id"`
	UserID    uuid.UUID        `json:"user_id"`
	Action    ActivityAction   `json:"action"`
	Metadata  ActivityMetadata `json:"metadata,omitempty"`
	CreatedAt time.Time        `json:"created_at"`
}

// PodWithOwner represents a pod with owner information for API responses.
type PodWithOwner struct {
	Pod
	OwnerName      string  `json:"owner_name"`
	OwnerAvatarURL *string `json:"owner_avatar_url,omitempty"`
}

// CollaboratorWithUser represents a collaborator with user information.
type CollaboratorWithUser struct {
	Collaborator
	UserName      string  `json:"user_name"`
	UserEmail     string  `json:"user_email"`
	UserAvatarURL *string `json:"user_avatar_url,omitempty"`
}

// ActivityWithDetails represents an activity with user and pod details.
type ActivityWithDetails struct {
	Activity
	UserName      string  `json:"user_name"`
	UserAvatarURL *string `json:"user_avatar_url,omitempty"`
	PodName       string  `json:"pod_name"`
	PodSlug       string  `json:"pod_slug"`
}

// NewPod creates a new Pod with default values.
func NewPod(ownerID uuid.UUID, name, slug string, visibility Visibility) *Pod {
	now := time.Now()
	return &Pod{
		ID:         uuid.New(),
		OwnerID:    ownerID,
		Name:       name,
		Slug:       slug,
		Visibility: visibility,
		StarCount:  0,
		ForkCount:  0,
		ViewCount:  0,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
}

// NewCollaborator creates a new Collaborator with pending status.
func NewCollaborator(podID, userID, invitedBy uuid.UUID, role CollaboratorRole) *Collaborator {
	now := time.Now()
	return &Collaborator{
		ID:        uuid.New(),
		PodID:     podID,
		UserID:    userID,
		Role:      role,
		Status:    CollaboratorStatusPending,
		InvitedBy: invitedBy,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// NewPodStar creates a new PodStar.
func NewPodStar(userID, podID uuid.UUID) *PodStar {
	return &PodStar{
		UserID:    userID,
		PodID:     podID,
		CreatedAt: time.Now(),
	}
}

// NewPodFollow creates a new PodFollow.
func NewPodFollow(userID, podID uuid.UUID) *PodFollow {
	return &PodFollow{
		UserID:    userID,
		PodID:     podID,
		CreatedAt: time.Now(),
	}
}

// NewActivity creates a new Activity.
func NewActivity(podID, userID uuid.UUID, action ActivityAction, metadata ActivityMetadata) *Activity {
	return &Activity{
		ID:        uuid.New(),
		PodID:     podID,
		UserID:    userID,
		Action:    action,
		Metadata:  metadata,
		CreatedAt: time.Now(),
	}
}

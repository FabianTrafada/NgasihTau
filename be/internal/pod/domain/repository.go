package domain

import (
	"context"

	"github.com/google/uuid"
)

// PodRepository defines the interface for pod data access.
// Implements the Repository pattern for data access abstraction (requirement 10.3).
type PodRepository interface {
	// Create creates a new pod.
	Create(ctx context.Context, pod *Pod) error

	// FindByID finds a pod by ID.
	FindByID(ctx context.Context, id uuid.UUID) (*Pod, error)

	// FindBySlug finds a pod by slug.
	FindBySlug(ctx context.Context, slug string) (*Pod, error)

	// FindByOwnerID finds all pods owned by a user.
	FindByOwnerID(ctx context.Context, ownerID uuid.UUID, limit, offset int) ([]*Pod, int, error)

	// Update updates an existing pod.
	Update(ctx context.Context, pod *Pod) error

	// Delete soft-deletes a pod.
	Delete(ctx context.Context, id uuid.UUID) error

	// ExistsBySlug checks if a pod with the given slug exists.
	ExistsBySlug(ctx context.Context, slug string) (bool, error)

	// IncrementStarCount increments the star count for a pod.
	IncrementStarCount(ctx context.Context, id uuid.UUID) error

	// DecrementStarCount decrements the star count for a pod.
	DecrementStarCount(ctx context.Context, id uuid.UUID) error

	// IncrementForkCount increments the fork count for a pod.
	IncrementForkCount(ctx context.Context, id uuid.UUID) error

	// IncrementViewCount increments the view count for a pod.
	IncrementViewCount(ctx context.Context, id uuid.UUID) error

	// Search searches pods with filters.
	Search(ctx context.Context, query string, filters PodFilters, limit, offset int) ([]*Pod, int, error)

	// GetPublicPods returns paginated public pods.
	GetPublicPods(ctx context.Context, limit, offset int) ([]*Pod, int, error)
}

// PodFilters contains filters for pod search.
type PodFilters struct {
	OwnerID    *uuid.UUID
	Category   *string
	Visibility *Visibility
	Tags       []string
}

// CollaboratorRepository defines the interface for collaborator data access.
type CollaboratorRepository interface {
	// Create creates a new collaborator.
	Create(ctx context.Context, collaborator *Collaborator) error

	// FindByID finds a collaborator by ID.
	FindByID(ctx context.Context, id uuid.UUID) (*Collaborator, error)

	// FindByPodAndUser finds a collaborator by pod ID and user ID.
	FindByPodAndUser(ctx context.Context, podID, userID uuid.UUID) (*Collaborator, error)

	// FindByPodID finds all collaborators for a pod.
	FindByPodID(ctx context.Context, podID uuid.UUID) ([]*Collaborator, error)

	// FindByPodIDWithUsers finds all collaborators for a pod with user details.
	FindByPodIDWithUsers(ctx context.Context, podID uuid.UUID) ([]*CollaboratorWithUser, error)

	// FindByUserID finds all collaborations for a user.
	FindByUserID(ctx context.Context, userID uuid.UUID) ([]*Collaborator, error)

	// Update updates a collaborator.
	Update(ctx context.Context, collaborator *Collaborator) error

	// Delete removes a collaborator.
	Delete(ctx context.Context, id uuid.UUID) error

	// DeleteByPodAndUser removes a collaborator by pod ID and user ID.
	DeleteByPodAndUser(ctx context.Context, podID, userID uuid.UUID) error

	// Exists checks if a collaborator relationship exists.
	Exists(ctx context.Context, podID, userID uuid.UUID) (bool, error)

	// UpdateStatus updates a collaborator's status.
	UpdateStatus(ctx context.Context, id uuid.UUID, status CollaboratorStatus) error

	// UpdateRole updates a collaborator's role.
	UpdateRole(ctx context.Context, id uuid.UUID, role CollaboratorRole) error
}

// PodStarRepository defines the interface for pod star data access.
type PodStarRepository interface {
	// Create creates a new star.
	Create(ctx context.Context, star *PodStar) error

	// Delete removes a star.
	Delete(ctx context.Context, userID, podID uuid.UUID) error

	// Exists checks if a star exists.
	Exists(ctx context.Context, userID, podID uuid.UUID) (bool, error)

	// GetStarredPods returns paginated starred pods for a user.
	GetStarredPods(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*Pod, int, error)

	// CountByPodID returns the star count for a pod.
	CountByPodID(ctx context.Context, podID uuid.UUID) (int, error)
}

// PodFollowRepository defines the interface for pod follow data access.
type PodFollowRepository interface {
	// Create creates a new follow.
	Create(ctx context.Context, follow *PodFollow) error

	// Delete removes a follow.
	Delete(ctx context.Context, userID, podID uuid.UUID) error

	// Exists checks if a follow exists.
	Exists(ctx context.Context, userID, podID uuid.UUID) (bool, error)

	// GetFollowedPods returns paginated followed pods for a user.
	GetFollowedPods(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*Pod, int, error)

	// GetFollowerIDs returns all user IDs following a pod.
	GetFollowerIDs(ctx context.Context, podID uuid.UUID) ([]uuid.UUID, error)

	// CountByPodID returns the follower count for a pod.
	CountByPodID(ctx context.Context, podID uuid.UUID) (int, error)
}

// ActivityRepository defines the interface for activity data access.
type ActivityRepository interface {
	// Create creates a new activity.
	Create(ctx context.Context, activity *Activity) error

	// FindByPodID finds activities for a pod.
	FindByPodID(ctx context.Context, podID uuid.UUID, limit, offset int) ([]*Activity, int, error)

	// FindByPodIDWithDetails finds activities for a pod with user details.
	FindByPodIDWithDetails(ctx context.Context, podID uuid.UUID, limit, offset int) ([]*ActivityWithDetails, int, error)

	// GetUserFeed returns activity feed for a user based on followed pods.
	GetUserFeed(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*ActivityWithDetails, int, error)

	// DeleteByPodID deletes all activities for a pod.
	DeleteByPodID(ctx context.Context, podID uuid.UUID) error
}

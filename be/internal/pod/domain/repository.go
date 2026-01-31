package domain

import (
	"context"
	"time"

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

	// IncrementUpvoteCount increments the upvote count for a pod.
	// Implements requirement 5.1.
	IncrementUpvoteCount(ctx context.Context, id uuid.UUID) error

	// DecrementUpvoteCount decrements the upvote count for a pod.
	// Implements requirement 5.2.
	DecrementUpvoteCount(ctx context.Context, id uuid.UUID) error

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

// ===========================================
// Recommendation System Repositories
// ===========================================

// InteractionRepository defines the interface for tracking user interactions.
type InteractionRepository interface {
	// Create records a new interaction.
	Create(ctx context.Context, interaction *PodInteraction) error

	// CreateBatch records multiple interactions at once.
	CreateBatch(ctx context.Context, interactions []*PodInteraction) error

	// FindByUserID finds interactions for a user.
	FindByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*PodInteraction, error)

	// FindByUserAndPod finds interactions for a user on a specific pod.
	FindByUserAndPod(ctx context.Context, userID, podID uuid.UUID, limit int) ([]*PodInteraction, error)

	// CountByUserID returns the total interaction count for a user.
	CountByUserID(ctx context.Context, userID uuid.UUID) (int, error)

	// GetRecentInteractionTypes returns distinct recent interaction types for a user-pod pair.
	GetRecentInteractionTypes(ctx context.Context, userID, podID uuid.UUID, since time.Time) ([]InteractionType, error)

	// GetUserInteractedPodIDs returns pod IDs the user has interacted with.
	GetUserInteractedPodIDs(ctx context.Context, userID uuid.UUID, limit int) ([]uuid.UUID, error)
}

// UserCategoryScoreRepository manages aggregated category preferences.
type UserCategoryScoreRepository interface {
	// Upsert creates or updates a category score.
	Upsert(ctx context.Context, score *UserCategoryScore) error

	// FindByUserID returns all category scores for a user, ordered by score.
	FindByUserID(ctx context.Context, userID uuid.UUID, limit int) ([]*UserCategoryScore, error)

	// FindByUserAndCategory returns a specific category score.
	FindByUserAndCategory(ctx context.Context, userID uuid.UUID, category string) (*UserCategoryScore, error)

	// IncrementScore increments the score for a category.
	IncrementScore(ctx context.Context, userID uuid.UUID, category string, delta float64, interactionType InteractionType) error

	// ApplyDecay applies time decay to scores older than the threshold.
	ApplyDecay(ctx context.Context, decayFactor float64, olderThan time.Time) error
}

// UserTagScoreRepository manages aggregated tag preferences.
type UserTagScoreRepository interface {
	// Upsert creates or updates a tag score.
	Upsert(ctx context.Context, score *UserTagScore) error

	// FindByUserID returns all tag scores for a user, ordered by score.
	FindByUserID(ctx context.Context, userID uuid.UUID, limit int) ([]*UserTagScore, error)

	// IncrementScore increments the score for a tag.
	IncrementScore(ctx context.Context, userID uuid.UUID, tag string, delta float64) error

	// IncrementScoreBatch increments scores for multiple tags.
	IncrementScoreBatch(ctx context.Context, userID uuid.UUID, tags []string, delta float64) error
}

// PodPopularityRepository manages pod popularity metrics.
type PodPopularityRepository interface {
	// Upsert creates or updates popularity score.
	Upsert(ctx context.Context, score *PodPopularityScore) error

	// FindByPodID returns popularity score for a pod.
	FindByPodID(ctx context.Context, podID uuid.UUID) (*PodPopularityScore, error)

	// GetTrendingPods returns pods ordered by trending score.
	GetTrendingPods(ctx context.Context, limit, offset int) ([]*PodPopularityScore, error)

	// RecalculateForPod recalculates popularity metrics for a specific pod.
	RecalculateForPod(ctx context.Context, podID uuid.UUID) error

	// RecalculateAll recalculates popularity metrics for all pods.
	RecalculateAll(ctx context.Context) error
}

// RecommendationRepository provides methods for generating recommendations.
type RecommendationRepository interface {
	// GetPersonalizedFeed returns personalized pod recommendations for a user.
	GetPersonalizedFeed(ctx context.Context, userID uuid.UUID, config *RecommendationConfig, limit, offset int) ([]*RecommendedPod, error)

	// GetTrendingFeed returns trending pods (for cold start or anonymous users).
	GetTrendingFeed(ctx context.Context, limit, offset int) ([]*Pod, error)

	// GetSimilarPods returns pods similar to a given pod.
	GetSimilarPods(ctx context.Context, podID uuid.UUID, limit int) ([]*Pod, error)

	// GetUserPreferenceProfile returns aggregated user preferences.
	GetUserPreferenceProfile(ctx context.Context, userID uuid.UUID) (*UserPreferenceProfile, error)

	// ExcludePods allows filtering out specific pods (e.g., already seen).
	GetPersonalizedFeedExcluding(ctx context.Context, userID uuid.UUID, excludePodIDs []uuid.UUID, config *RecommendationConfig, limit int) ([]*RecommendedPod, error)
}

// ===========================================
// Student-Teacher Roles Repositories
// ===========================================

// PodUpvoteRepository defines the interface for pod upvote data access.
// Follows the existing PodStarRepository pattern for consistency.
// Implements requirements 5.1, 5.2, 5.3.
type PodUpvoteRepository interface {
	// Create creates a new upvote.
	Create(ctx context.Context, upvote *PodUpvote) error

	// Delete removes an upvote.
	Delete(ctx context.Context, userID, podID uuid.UUID) error

	// Exists checks if an upvote exists.
	Exists(ctx context.Context, userID, podID uuid.UUID) (bool, error)

	// CountByPodID returns the upvote count for a pod.
	CountByPodID(ctx context.Context, podID uuid.UUID) (int, error)

	// GetUpvotedPods returns paginated upvoted pods for a user.
	GetUpvotedPods(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*Pod, int, error)
}

// PodDownvoteRepository defines the interface for pod downvote data access.
// Follows the existing PodUpvoteRepository pattern for consistency.
type PodDownvoteRepository interface {
	// Create creates a new downvote.
	Create(ctx context.Context, downvote *PodDownvote) error

	// Delete removes a downvote.
	Delete(ctx context.Context, userID, podID uuid.UUID) error

	// Exists checks if a downvote exists.
	Exists(ctx context.Context, userID, podID uuid.UUID) (bool, error)

	// CountByPodID returns the downvote count for a pod.
	CountByPodID(ctx context.Context, podID uuid.UUID) (int, error)

	// GetDownvotedPods returns paginated downvoted pods for a user.
	GetDownvotedPods(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*Pod, int, error)
}

// UploadRequestRepository defines the interface for upload request data access.
// Enables teacher-to-teacher collaboration for quality educational content.
// Implements requirements 4.1, 4.3, 4.6.
type UploadRequestRepository interface {
	// Create creates a new upload request.
	Create(ctx context.Context, request *UploadRequest) error

	// FindByID finds an upload request by ID.
	FindByID(ctx context.Context, id uuid.UUID) (*UploadRequest, error)

	// FindByRequesterAndPod finds an upload request by requester ID and pod ID.
	FindByRequesterAndPod(ctx context.Context, requesterID, podID uuid.UUID) (*UploadRequest, error)

	// FindByPodOwner finds upload requests for a pod owner with optional status filter.
	FindByPodOwner(ctx context.Context, ownerID uuid.UUID, status *UploadRequestStatus, limit, offset int) ([]*UploadRequest, int, error)

	// FindByRequester finds upload requests made by a requester.
	FindByRequester(ctx context.Context, requesterID uuid.UUID, limit, offset int) ([]*UploadRequest, int, error)

	// FindApprovedByRequesterAndPod finds an approved upload request for a requester and pod.
	FindApprovedByRequesterAndPod(ctx context.Context, requesterID, podID uuid.UUID) (*UploadRequest, error)

	// Update updates an upload request.
	Update(ctx context.Context, request *UploadRequest) error

	// UpdateStatus updates the status of an upload request with optional rejection reason.
	UpdateStatus(ctx context.Context, id uuid.UUID, status UploadRequestStatus, reason *string) error
}

// SharedPodRepository defines the interface for shared pod data access.
// Enables teachers to share pods with students for guided learning.
// Implements requirement 7.2.
type SharedPodRepository interface {
	// Create creates a new shared pod record.
	Create(ctx context.Context, share *SharedPod) error

	// Delete removes a shared pod record.
	Delete(ctx context.Context, id uuid.UUID) error

	// FindByStudent finds shared pods for a student with pagination.
	FindByStudent(ctx context.Context, studentID uuid.UUID, limit, offset int) ([]*SharedPod, int, error)

	// FindByStudentWithDetails finds shared pods for a student with pod and teacher details.
	FindByStudentWithDetails(ctx context.Context, studentID uuid.UUID, limit, offset int) ([]*SharedPodWithDetails, int, error)

	// FindByTeacherAndStudent finds shared pods between a specific teacher and student.
	FindByTeacherAndStudent(ctx context.Context, teacherID, studentID uuid.UUID) ([]*SharedPod, error)

	// Exists checks if a pod is already shared with a student.
	Exists(ctx context.Context, podID, studentID uuid.UUID) (bool, error)
}

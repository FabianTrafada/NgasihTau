// Package application contains the business logic and use cases for the Pod Service.
// This layer orchestrates the domain entities and repositories to implement features.
package application

import (
	"context"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"

	"ngasihtau/internal/common/errors"
	"ngasihtau/internal/pod/domain"
)

// PodService defines the interface for pod-related business operations.
// Implements requirements 3, 3.1, 3.2, 4, 5, 12.
type PodService interface {
	// Pod CRUD operations (Requirement 3)
	CreatePod(ctx context.Context, input CreatePodInput) (*domain.Pod, error)
	GetPod(ctx context.Context, id uuid.UUID, viewerID *uuid.UUID) (*domain.Pod, error)
	GetPodBySlug(ctx context.Context, slug string, viewerID *uuid.UUID) (*domain.Pod, error)
	UpdatePod(ctx context.Context, id uuid.UUID, userID uuid.UUID, input UpdatePodInput) (*domain.Pod, error)
	DeletePod(ctx context.Context, id, userID uuid.UUID) error
	ListPods(ctx context.Context, filters domain.PodFilters, page, perPage int) (*PodListResult, error)
	ListUserPods(ctx context.Context, ownerID uuid.UUID, page, perPage int) (*PodListResult, error)

	// Fork operations (Requirement 3.1)
	ForkPod(ctx context.Context, podID, userID uuid.UUID) (*domain.Pod, error)

	// Star operations (Requirement 3.2)
	StarPod(ctx context.Context, podID, userID uuid.UUID) error
	UnstarPod(ctx context.Context, podID, userID uuid.UUID) error
	GetStarredPods(ctx context.Context, userID uuid.UUID, page, perPage int) (*PodListResult, error)
	IsStarred(ctx context.Context, podID, userID uuid.UUID) (bool, error)

	// Upvote operations (Requirements 5.1, 5.2, 5.3, 5.4)
	// Upvotes are trust indicators, distinct from stars (bookmarks/favorites)
	UpvotePod(ctx context.Context, podID, userID uuid.UUID) error
	RemoveUpvote(ctx context.Context, podID, userID uuid.UUID) error
	HasUpvoted(ctx context.Context, podID, userID uuid.UUID) (bool, error)
	GetUpvotedPods(ctx context.Context, userID uuid.UUID, page, perPage int) (*PodListResult, error)

	// Collaborator operations (Requirement 4)
	InviteCollaborator(ctx context.Context, input InviteCollaboratorInput) (*domain.Collaborator, error)
	AcceptInvitation(ctx context.Context, podID, userID uuid.UUID) error
	VerifyCollaborator(ctx context.Context, podID, collaboratorID, ownerID uuid.UUID) error
	RemoveCollaborator(ctx context.Context, podID, collaboratorID, requesterID uuid.UUID) error
	UpdateCollaboratorRole(ctx context.Context, podID, collaboratorID, ownerID uuid.UUID, role domain.CollaboratorRole) error
	GetCollaborators(ctx context.Context, podID uuid.UUID) ([]*domain.CollaboratorWithUser, error)
	GetUserCollaborations(ctx context.Context, userID uuid.UUID) ([]*domain.Collaborator, error)

	// Follow operations (Requirement 12)
	FollowPod(ctx context.Context, podID, userID uuid.UUID) error
	UnfollowPod(ctx context.Context, podID, userID uuid.UUID) error
	GetFollowedPods(ctx context.Context, userID uuid.UUID, page, perPage int) (*PodListResult, error)
	IsFollowing(ctx context.Context, podID, userID uuid.UUID) (bool, error)

	// Activity feed (Requirement 12)
	GetPodActivity(ctx context.Context, podID uuid.UUID, page, perPage int) (*ActivityListResult, error)
	GetUserFeed(ctx context.Context, userID uuid.UUID, page, perPage int) (*ActivityListResult, error)

	// Permission checks
	CanUserAccessPod(ctx context.Context, podID uuid.UUID, userID *uuid.UUID) (bool, error)
	CanUserEditPod(ctx context.Context, podID, userID uuid.UUID) (bool, error)
	CanUserUploadToPod(ctx context.Context, podID, userID uuid.UUID) (bool, error)
}

// CreatePodInput contains the data required for pod creation.
type CreatePodInput struct {
	OwnerID     uuid.UUID         `json:"-"` // Set from auth context
	Name        string            `json:"name" validate:"required,min=3,max=100"`
	Description *string           `json:"description,omitempty" validate:"omitempty,max=2000"`
	Visibility  domain.Visibility `json:"visibility" validate:"required,oneof=public private"`
	Categories  []string          `json:"categories,omitempty" validate:"omitempty,max=5,dive,max=50"`
	Tags        []string          `json:"tags,omitempty" validate:"omitempty,max=10,dive,max=30"`
}

// UpdatePodInput contains the data for updating a pod.
type UpdatePodInput struct {
	Name        *string            `json:"name,omitempty" validate:"omitempty,min=3,max=100"`
	Description *string            `json:"description,omitempty" validate:"omitempty,max=2000"`
	Visibility  *domain.Visibility `json:"visibility,omitempty" validate:"omitempty,oneof=public private"`
	Categories  []string           `json:"categories,omitempty" validate:"omitempty,max=5,dive,max=50"`
	Tags        []string           `json:"tags,omitempty" validate:"omitempty,max=10,dive,max=30"`
}

// InviteCollaboratorInput contains the data for inviting a collaborator.
type InviteCollaboratorInput struct {
	PodID     uuid.UUID               `json:"-"` // Set from URL param
	InviterID uuid.UUID               `json:"-"` // Set from auth context
	UserID    uuid.UUID               `json:"user_id" validate:"required"`
	Role      domain.CollaboratorRole `json:"role" validate:"required,oneof=viewer contributor admin"`
}

// PodListResult contains a paginated list of pods.
type PodListResult struct {
	Pods       []*domain.Pod `json:"pods"`
	Total      int           `json:"total"`
	Page       int           `json:"page"`
	PerPage    int           `json:"per_page"`
	TotalPages int           `json:"total_pages"`
}

// ActivityListResult contains a paginated list of activities.
type ActivityListResult struct {
	Activities []*domain.ActivityWithDetails `json:"activities"`
	Total      int                           `json:"total"`
	Page       int                           `json:"page"`
	PerPage    int                           `json:"per_page"`
	TotalPages int                           `json:"total_pages"`
}

// podService implements the PodService interface.
type podService struct {
	podRepo          domain.PodRepository
	collaboratorRepo domain.CollaboratorRepository
	starRepo         domain.PodStarRepository
	upvoteRepo       domain.PodUpvoteRepository
	followRepo       domain.PodFollowRepository
	activityRepo     domain.ActivityRepository
	eventPublisher   EventPublisher
}

// EventPublisher defines the interface for publishing pod events.
type EventPublisher interface {
	PublishPodCreated(ctx context.Context, pod *domain.Pod) error
	PublishPodUpdated(ctx context.Context, pod *domain.Pod) error
	PublishCollaboratorInvited(ctx context.Context, collaborator *domain.Collaborator, podName string) error
}

// NewPodService creates a new PodService instance.
func NewPodService(
	podRepo domain.PodRepository,
	collaboratorRepo domain.CollaboratorRepository,
	starRepo domain.PodStarRepository,
	upvoteRepo domain.PodUpvoteRepository,
	followRepo domain.PodFollowRepository,
	activityRepo domain.ActivityRepository,
	eventPublisher EventPublisher,
) PodService {
	return &podService{
		podRepo:          podRepo,
		collaboratorRepo: collaboratorRepo,
		starRepo:         starRepo,
		upvoteRepo:       upvoteRepo,
		followRepo:       followRepo,
		activityRepo:     activityRepo,
		eventPublisher:   eventPublisher,
	}
}

// generateSlug creates a URL-friendly slug from a name.
func generateSlug(name string) string {
	// Convert to lowercase
	slug := strings.ToLower(name)
	// Replace spaces with hyphens
	slug = strings.ReplaceAll(slug, " ", "-")
	// Remove non-alphanumeric characters except hyphens
	reg := regexp.MustCompile("[^a-z0-9-]+")
	slug = reg.ReplaceAllString(slug, "")
	// Remove multiple consecutive hyphens
	reg = regexp.MustCompile("-+")
	slug = reg.ReplaceAllString(slug, "-")
	// Trim hyphens from start and end
	slug = strings.Trim(slug, "-")
	return slug
}

// calculateTotalPages calculates the total number of pages.
func calculateTotalPages(total, perPage int) int {
	if perPage <= 0 {
		return 0
	}
	pages := total / perPage
	if total%perPage > 0 {
		pages++
	}
	return pages
}

// CreatePod creates a new Knowledge Pod.
// Implements requirement 3: Knowledge Pod Creation.
func (s *podService) CreatePod(ctx context.Context, input CreatePodInput) (*domain.Pod, error) {
	// Generate slug from name
	baseSlug := generateSlug(input.Name)
	slug := baseSlug

	// Ensure slug is unique by appending a suffix if needed
	suffix := 1
	for {
		exists, err := s.podRepo.ExistsBySlug(ctx, slug)
		if err != nil {
			return nil, err
		}
		if !exists {
			break
		}
		slug = baseSlug + "-" + uuid.New().String()[:8]
		suffix++
		if suffix > 10 {
			// Fallback to UUID-based slug
			slug = baseSlug + "-" + uuid.New().String()[:8]
			break
		}
	}

	// Create pod
	// Note: isCreatorTeacher is set to false by default here.
	// Task 13.1 will update this to check the actual user role.
	pod := domain.NewPod(input.OwnerID, input.Name, slug, input.Visibility, false)
	pod.Description = input.Description
	pod.Categories = input.Categories
	pod.Tags = input.Tags

	if err := s.podRepo.Create(ctx, pod); err != nil {
		return nil, err
	}

	// Publish pod.created event for search indexing
	if s.eventPublisher != nil {
		go func() {
			if err := s.eventPublisher.PublishPodCreated(context.Background(), pod); err != nil {
				// Log error but don't fail the request
				// The pod is already created, event publishing is best-effort
			}
		}()
	}

	return pod, nil
}

// GetPod retrieves a pod by ID.
func (s *podService) GetPod(ctx context.Context, id uuid.UUID, viewerID *uuid.UUID) (*domain.Pod, error) {
	pod, err := s.podRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Check access permission
	canAccess, err := s.CanUserAccessPod(ctx, id, viewerID)
	if err != nil {
		return nil, err
	}
	if !canAccess {
		return nil, errors.Forbidden("you do not have access to this pod")
	}

	// Increment view count (fire and forget)
	go func() {
		_ = s.podRepo.IncrementViewCount(context.Background(), id)
	}()

	return pod, nil
}

// GetPodBySlug retrieves a pod by slug.
func (s *podService) GetPodBySlug(ctx context.Context, slug string, viewerID *uuid.UUID) (*domain.Pod, error) {
	pod, err := s.podRepo.FindBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}

	// Check access permission
	canAccess, err := s.CanUserAccessPod(ctx, pod.ID, viewerID)
	if err != nil {
		return nil, err
	}
	if !canAccess {
		return nil, errors.Forbidden("you do not have access to this pod")
	}

	// Increment view count (fire and forget)
	go func() {
		_ = s.podRepo.IncrementViewCount(context.Background(), pod.ID)
	}()

	return pod, nil
}

// UpdatePod updates a pod's information.
func (s *podService) UpdatePod(ctx context.Context, id uuid.UUID, userID uuid.UUID, input UpdatePodInput) (*domain.Pod, error) {
	// Check if user can edit
	canEdit, err := s.CanUserEditPod(ctx, id, userID)
	if err != nil {
		return nil, err
	}
	if !canEdit {
		return nil, errors.Forbidden("you do not have permission to edit this pod")
	}

	pod, err := s.podRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Update fields if provided
	if input.Name != nil {
		pod.Name = *input.Name
	}
	if input.Description != nil {
		pod.Description = input.Description
	}
	if input.Visibility != nil {
		pod.Visibility = *input.Visibility
	}
	if input.Categories != nil {
		pod.Categories = input.Categories
	}
	if input.Tags != nil {
		pod.Tags = input.Tags
	}
	pod.UpdatedAt = time.Now()

	if err := s.podRepo.Update(ctx, pod); err != nil {
		return nil, err
	}

	// Log activity
	activity := domain.NewActivity(pod.ID, userID, domain.ActivityActionPodUpdated, domain.ActivityMetadata{
		"changes": "pod_updated",
	})
	_ = s.activityRepo.Create(ctx, activity)

	// Publish pod.updated event for search re-indexing
	if s.eventPublisher != nil {
		go func() {
			if err := s.eventPublisher.PublishPodUpdated(context.Background(), pod); err != nil {
				// Log error but don't fail the request
			}
		}()
	}

	return pod, nil
}

// DeletePod soft-deletes a pod.
func (s *podService) DeletePod(ctx context.Context, id, userID uuid.UUID) error {
	pod, err := s.podRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	// Only owner can delete
	if !pod.IsOwner(userID) {
		return errors.Forbidden("only the owner can delete this pod")
	}

	return s.podRepo.Delete(ctx, id)
}

// ListPods returns a paginated list of pods with filters.
func (s *podService) ListPods(ctx context.Context, filters domain.PodFilters, page, perPage int) (*PodListResult, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}
	offset := (page - 1) * perPage

	pods, total, err := s.podRepo.Search(ctx, "", filters, perPage, offset)
	if err != nil {
		return nil, err
	}

	return &PodListResult{
		Pods:       pods,
		Total:      total,
		Page:       page,
		PerPage:    perPage,
		TotalPages: calculateTotalPages(total, perPage),
	}, nil
}

// ListUserPods returns pods owned by a user.
func (s *podService) ListUserPods(ctx context.Context, ownerID uuid.UUID, page, perPage int) (*PodListResult, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}
	offset := (page - 1) * perPage

	pods, total, err := s.podRepo.FindByOwnerID(ctx, ownerID, perPage, offset)
	if err != nil {
		return nil, err
	}

	return &PodListResult{
		Pods:       pods,
		Total:      total,
		Page:       page,
		PerPage:    perPage,
		TotalPages: calculateTotalPages(total, perPage),
	}, nil
}

// ForkPod creates a copy of a pod.
// Implements requirement 3.1: Knowledge Pod Fork/Clone.
func (s *podService) ForkPod(ctx context.Context, podID, userID uuid.UUID) (*domain.Pod, error) {
	// Get original pod
	originalPod, err := s.podRepo.FindByID(ctx, podID)
	if err != nil {
		return nil, err
	}

	// Check if user can access the original pod
	canAccess, err := s.CanUserAccessPod(ctx, podID, &userID)
	if err != nil {
		return nil, err
	}
	if !canAccess {
		return nil, errors.Forbidden("you do not have access to fork this pod")
	}

	// Generate new slug
	baseSlug := generateSlug(originalPod.Name + "-fork")
	slug := baseSlug
	suffix := 1
	for {
		exists, err := s.podRepo.ExistsBySlug(ctx, slug)
		if err != nil {
			return nil, err
		}
		if !exists {
			break
		}
		slug = baseSlug + "-" + uuid.New().String()[:8]
		suffix++
		if suffix > 10 {
			slug = baseSlug + "-" + uuid.New().String()[:8]
			break
		}
	}

	// Create forked pod
	// Note: isCreatorTeacher is set to false by default here.
	// Task 13.1 will update this to check the actual user role.
	forkedPod := domain.NewPod(userID, originalPod.Name, slug, domain.VisibilityPublic, false)
	forkedPod.Description = originalPod.Description
	forkedPod.Categories = originalPod.Categories
	forkedPod.Tags = originalPod.Tags
	forkedPod.ForkedFromID = &podID

	if err := s.podRepo.Create(ctx, forkedPod); err != nil {
		return nil, err
	}

	// Increment fork count on original pod
	_ = s.podRepo.IncrementForkCount(ctx, podID)

	// Log activity on original pod
	activity := domain.NewActivity(podID, userID, domain.ActivityActionPodForked, domain.ActivityMetadata{
		"forked_pod_id": forkedPod.ID.String(),
	})
	_ = s.activityRepo.Create(ctx, activity)

	return forkedPod, nil
}

// StarPod adds a star to a pod.
// Implements requirement 3.2: Star/Bookmark Pods.
func (s *podService) StarPod(ctx context.Context, podID, userID uuid.UUID) error {
	// Check if pod exists and user can access it
	canAccess, err := s.CanUserAccessPod(ctx, podID, &userID)
	if err != nil {
		return err
	}
	if !canAccess {
		return errors.Forbidden("you do not have access to this pod")
	}

	// Check if already starred
	exists, err := s.starRepo.Exists(ctx, userID, podID)
	if err != nil {
		return err
	}
	if exists {
		return errors.Conflict("star", "already starred")
	}

	// Create star
	star := domain.NewPodStar(userID, podID)
	if err := s.starRepo.Create(ctx, star); err != nil {
		return err
	}

	// Increment star count
	return s.podRepo.IncrementStarCount(ctx, podID)
}

// UnstarPod removes a star from a pod.
func (s *podService) UnstarPod(ctx context.Context, podID, userID uuid.UUID) error {
	// Check if starred
	exists, err := s.starRepo.Exists(ctx, userID, podID)
	if err != nil {
		return err
	}
	if !exists {
		return errors.NotFound("star", "not starred")
	}

	// Delete star
	if err := s.starRepo.Delete(ctx, userID, podID); err != nil {
		return err
	}

	// Decrement star count
	return s.podRepo.DecrementStarCount(ctx, podID)
}

// GetStarredPods returns pods starred by a user.
func (s *podService) GetStarredPods(ctx context.Context, userID uuid.UUID, page, perPage int) (*PodListResult, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}
	offset := (page - 1) * perPage

	pods, total, err := s.starRepo.GetStarredPods(ctx, userID, perPage, offset)
	if err != nil {
		return nil, err
	}

	return &PodListResult{
		Pods:       pods,
		Total:      total,
		Page:       page,
		PerPage:    perPage,
		TotalPages: calculateTotalPages(total, perPage),
	}, nil
}

// IsStarred checks if a user has starred a pod.
func (s *podService) IsStarred(ctx context.Context, podID, userID uuid.UUID) (bool, error) {
	return s.starRepo.Exists(ctx, userID, podID)
}

// UpvotePod adds an upvote to a pod (trust indicator).
// Implements requirement 5.1: WHEN a user upvotes a knowledge pod.
// Implements requirement 5.3: Each user can upvote a pod only once.
func (s *podService) UpvotePod(ctx context.Context, podID, userID uuid.UUID) error {
	// Check if pod exists and user can access it
	canAccess, err := s.CanUserAccessPod(ctx, podID, &userID)
	if err != nil {
		return err
	}
	if !canAccess {
		return errors.Forbidden("you do not have access to this pod")
	}

	// Check if already upvoted (requirement 5.3)
	exists, err := s.upvoteRepo.Exists(ctx, userID, podID)
	if err != nil {
		return err
	}
	if exists {
		return errors.Conflict("upvote", "already upvoted")
	}

	// Create upvote
	upvote := domain.NewPodUpvote(userID, podID)
	if err := s.upvoteRepo.Create(ctx, upvote); err != nil {
		return err
	}

	// Increment upvote count (requirement 5.1)
	return s.podRepo.IncrementUpvoteCount(ctx, podID)
}

// RemoveUpvote removes an upvote from a pod.
// Implements requirement 5.2: WHEN a user removes their upvote.
func (s *podService) RemoveUpvote(ctx context.Context, podID, userID uuid.UUID) error {
	// Check if upvoted
	exists, err := s.upvoteRepo.Exists(ctx, userID, podID)
	if err != nil {
		return err
	}
	if !exists {
		return errors.NotFound("upvote", "not upvoted")
	}

	// Delete upvote
	if err := s.upvoteRepo.Delete(ctx, userID, podID); err != nil {
		return err
	}

	// Decrement upvote count (requirement 5.2)
	return s.podRepo.DecrementUpvoteCount(ctx, podID)
}

// HasUpvoted checks if a user has upvoted a pod.
// Implements requirement 5.3: Each user can upvote a pod only once.
func (s *podService) HasUpvoted(ctx context.Context, podID, userID uuid.UUID) (bool, error) {
	return s.upvoteRepo.Exists(ctx, userID, podID)
}

// GetUpvotedPods returns pods upvoted by a user.
// Implements requirement 5.4: Show total upvote count when displaying pod details.
func (s *podService) GetUpvotedPods(ctx context.Context, userID uuid.UUID, page, perPage int) (*PodListResult, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}
	offset := (page - 1) * perPage

	pods, total, err := s.upvoteRepo.GetUpvotedPods(ctx, userID, perPage, offset)
	if err != nil {
		return nil, err
	}

	return &PodListResult{
		Pods:       pods,
		Total:      total,
		Page:       page,
		PerPage:    perPage,
		TotalPages: calculateTotalPages(total, perPage),
	}, nil
}

// InviteCollaborator invites a user to collaborate on a pod.
// Implements requirement 4: Knowledge Pod Collaboration.
func (s *podService) InviteCollaborator(ctx context.Context, input InviteCollaboratorInput) (*domain.Collaborator, error) {
	// Check if inviter can manage the pod
	canEdit, err := s.CanUserEditPod(ctx, input.PodID, input.InviterID)
	if err != nil {
		return nil, err
	}
	if !canEdit {
		return nil, errors.Forbidden("you do not have permission to invite collaborators")
	}

	// Check if user is already a collaborator
	exists, err := s.collaboratorRepo.Exists(ctx, input.PodID, input.UserID)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.Conflict("collaborator", "user is already a collaborator")
	}

	// Check if user is the owner
	pod, err := s.podRepo.FindByID(ctx, input.PodID)
	if err != nil {
		return nil, err
	}
	if pod.IsOwner(input.UserID) {
		return nil, errors.BadRequest("cannot invite the owner as a collaborator")
	}

	// Create collaborator with pending status
	collaborator := domain.NewCollaborator(input.PodID, input.UserID, input.InviterID, input.Role)
	if err := s.collaboratorRepo.Create(ctx, collaborator); err != nil {
		return nil, err
	}

	// Log activity
	activity := domain.NewActivity(input.PodID, input.InviterID, domain.ActivityActionCollaboratorAdded, domain.ActivityMetadata{
		"collaborator_id": input.UserID.String(),
		"role":            string(input.Role),
	})
	_ = s.activityRepo.Create(ctx, activity)

	// Publish collaborator.invited event for notification
	if s.eventPublisher != nil {
		go func() {
			if err := s.eventPublisher.PublishCollaboratorInvited(context.Background(), collaborator, pod.Name); err != nil {
				// Log error but don't fail the request
			}
		}()
	}

	return collaborator, nil
}

// AcceptInvitation accepts a collaboration invitation.
func (s *podService) AcceptInvitation(ctx context.Context, podID, userID uuid.UUID) error {
	collaborator, err := s.collaboratorRepo.FindByPodAndUser(ctx, podID, userID)
	if err != nil {
		return err
	}

	if collaborator.Status != domain.CollaboratorStatusPending {
		return errors.BadRequest("invitation is not pending")
	}

	// Update status to pending verification
	return s.collaboratorRepo.UpdateStatus(ctx, collaborator.ID, domain.CollaboratorStatusPendingVerification)
}

// VerifyCollaborator verifies a collaborator (owner action).
func (s *podService) VerifyCollaborator(ctx context.Context, podID, collaboratorID, ownerID uuid.UUID) error {
	// Check if requester is the owner
	pod, err := s.podRepo.FindByID(ctx, podID)
	if err != nil {
		return err
	}
	if !pod.IsOwner(ownerID) {
		return errors.Forbidden("only the owner can verify collaborators")
	}

	collaborator, err := s.collaboratorRepo.FindByID(ctx, collaboratorID)
	if err != nil {
		return err
	}

	if collaborator.PodID != podID {
		return errors.BadRequest("collaborator does not belong to this pod")
	}

	if collaborator.Status == domain.CollaboratorStatusVerified {
		return errors.BadRequest("collaborator is already verified")
	}

	return s.collaboratorRepo.UpdateStatus(ctx, collaboratorID, domain.CollaboratorStatusVerified)
}

// RemoveCollaborator removes a collaborator from a pod.
func (s *podService) RemoveCollaborator(ctx context.Context, podID, collaboratorID, requesterID uuid.UUID) error {
	collaborator, err := s.collaboratorRepo.FindByID(ctx, collaboratorID)
	if err != nil {
		return err
	}

	if collaborator.PodID != podID {
		return errors.BadRequest("collaborator does not belong to this pod")
	}

	// Check permissions: owner can remove anyone, user can remove themselves
	pod, err := s.podRepo.FindByID(ctx, podID)
	if err != nil {
		return err
	}

	if !pod.IsOwner(requesterID) && collaborator.UserID != requesterID {
		return errors.Forbidden("you do not have permission to remove this collaborator")
	}

	return s.collaboratorRepo.Delete(ctx, collaboratorID)
}

// UpdateCollaboratorRole updates a collaborator's role.
func (s *podService) UpdateCollaboratorRole(ctx context.Context, podID, collaboratorID, ownerID uuid.UUID, role domain.CollaboratorRole) error {
	// Check if requester is the owner
	pod, err := s.podRepo.FindByID(ctx, podID)
	if err != nil {
		return err
	}
	if !pod.IsOwner(ownerID) {
		return errors.Forbidden("only the owner can update collaborator roles")
	}

	collaborator, err := s.collaboratorRepo.FindByID(ctx, collaboratorID)
	if err != nil {
		return err
	}

	if collaborator.PodID != podID {
		return errors.BadRequest("collaborator does not belong to this pod")
	}

	return s.collaboratorRepo.UpdateRole(ctx, collaboratorID, role)
}

// GetCollaborators returns all collaborators for a pod.
func (s *podService) GetCollaborators(ctx context.Context, podID uuid.UUID) ([]*domain.CollaboratorWithUser, error) {
	return s.collaboratorRepo.FindByPodIDWithUsers(ctx, podID)
}

// GetUserCollaborations returns all pods a user collaborates on.
func (s *podService) GetUserCollaborations(ctx context.Context, userID uuid.UUID) ([]*domain.Collaborator, error) {
	return s.collaboratorRepo.FindByUserID(ctx, userID)
}

// FollowPod follows a pod.
// Implements requirement 12: Activity Feed.
func (s *podService) FollowPod(ctx context.Context, podID, userID uuid.UUID) error {
	// Check if pod exists and user can access it
	canAccess, err := s.CanUserAccessPod(ctx, podID, &userID)
	if err != nil {
		return err
	}
	if !canAccess {
		return errors.Forbidden("you do not have access to this pod")
	}

	// Check if already following
	exists, err := s.followRepo.Exists(ctx, userID, podID)
	if err != nil {
		return err
	}
	if exists {
		return errors.Conflict("follow", "already following")
	}

	// Create follow
	follow := domain.NewPodFollow(userID, podID)
	return s.followRepo.Create(ctx, follow)
}

// UnfollowPod unfollows a pod.
func (s *podService) UnfollowPod(ctx context.Context, podID, userID uuid.UUID) error {
	// Check if following
	exists, err := s.followRepo.Exists(ctx, userID, podID)
	if err != nil {
		return err
	}
	if !exists {
		return errors.NotFound("follow", "not following")
	}

	return s.followRepo.Delete(ctx, userID, podID)
}

// GetFollowedPods returns pods followed by a user.
func (s *podService) GetFollowedPods(ctx context.Context, userID uuid.UUID, page, perPage int) (*PodListResult, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}
	offset := (page - 1) * perPage

	pods, total, err := s.followRepo.GetFollowedPods(ctx, userID, perPage, offset)
	if err != nil {
		return nil, err
	}

	return &PodListResult{
		Pods:       pods,
		Total:      total,
		Page:       page,
		PerPage:    perPage,
		TotalPages: calculateTotalPages(total, perPage),
	}, nil
}

// IsFollowing checks if a user is following a pod.
func (s *podService) IsFollowing(ctx context.Context, podID, userID uuid.UUID) (bool, error) {
	return s.followRepo.Exists(ctx, userID, podID)
}

// GetPodActivity returns activity for a pod.
func (s *podService) GetPodActivity(ctx context.Context, podID uuid.UUID, page, perPage int) (*ActivityListResult, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}
	offset := (page - 1) * perPage

	activities, total, err := s.activityRepo.FindByPodIDWithDetails(ctx, podID, perPage, offset)
	if err != nil {
		return nil, err
	}

	return &ActivityListResult{
		Activities: activities,
		Total:      total,
		Page:       page,
		PerPage:    perPage,
		TotalPages: calculateTotalPages(total, perPage),
	}, nil
}

// GetUserFeed returns activity feed for a user based on followed pods.
func (s *podService) GetUserFeed(ctx context.Context, userID uuid.UUID, page, perPage int) (*ActivityListResult, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}
	offset := (page - 1) * perPage

	activities, total, err := s.activityRepo.GetUserFeed(ctx, userID, perPage, offset)
	if err != nil {
		return nil, err
	}

	return &ActivityListResult{
		Activities: activities,
		Total:      total,
		Page:       page,
		PerPage:    perPage,
		TotalPages: calculateTotalPages(total, perPage),
	}, nil
}

// CanUserAccessPod checks if a user can access a pod.
func (s *podService) CanUserAccessPod(ctx context.Context, podID uuid.UUID, userID *uuid.UUID) (bool, error) {
	pod, err := s.podRepo.FindByID(ctx, podID)
	if err != nil {
		return false, err
	}

	// Public pods are accessible to everyone
	if pod.IsPublic() {
		return true, nil
	}

	// Private pods require authentication
	if userID == nil {
		return false, nil
	}

	// Owner can always access
	if pod.IsOwner(*userID) {
		return true, nil
	}

	// Check if user is a collaborator
	exists, err := s.collaboratorRepo.Exists(ctx, podID, *userID)
	if err != nil {
		return false, err
	}

	return exists, nil
}

// CanUserEditPod checks if a user can edit a pod.
func (s *podService) CanUserEditPod(ctx context.Context, podID, userID uuid.UUID) (bool, error) {
	pod, err := s.podRepo.FindByID(ctx, podID)
	if err != nil {
		return false, err
	}

	// Owner can always edit
	if pod.IsOwner(userID) {
		return true, nil
	}

	// Check if user is an admin collaborator
	collaborator, err := s.collaboratorRepo.FindByPodAndUser(ctx, podID, userID)
	if err != nil {
		// Not found means no permission
		if appErr, ok := err.(*errors.AppError); ok && appErr.Code == errors.CodeNotFound {
			return false, nil
		}
		return false, err
	}

	return collaborator.CanManage(), nil
}

// CanUserUploadToPod checks if a user can upload materials to a pod.
func (s *podService) CanUserUploadToPod(ctx context.Context, podID, userID uuid.UUID) (bool, error) {
	pod, err := s.podRepo.FindByID(ctx, podID)
	if err != nil {
		return false, err
	}

	// Owner can always upload
	if pod.IsOwner(userID) {
		return true, nil
	}

	// Check if user is a verified contributor or admin
	collaborator, err := s.collaboratorRepo.FindByPodAndUser(ctx, podID, userID)
	if err != nil {
		// Not found means no permission
		if appErr, ok := err.(*errors.AppError); ok && appErr.Code == errors.CodeNotFound {
			return false, nil
		}
		return false, err
	}

	return collaborator.CanUpload(), nil
}

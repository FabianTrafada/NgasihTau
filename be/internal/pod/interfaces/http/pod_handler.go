package http

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"ngasihtau/internal/common/errors"
	"ngasihtau/internal/common/middleware"
	"ngasihtau/internal/common/response"
	"ngasihtau/internal/common/validator"
	"ngasihtau/internal/pod/application"
	"ngasihtau/internal/pod/domain"

	_ "ngasihtau/docs" // Swagger docs
)

// PodHandler handles HTTP requests for pod operations.
type PodHandler struct {
	podService            application.PodService
	recommendationService application.RecommendationService
	validator             *validator.Validator
}

// NewPodHandler creates a new PodHandler.
func NewPodHandler(podService application.PodService, recommendationService application.RecommendationService) *PodHandler {
	return &PodHandler{
		podService:            podService,
		recommendationService: recommendationService,
		validator:             validator.Get(),
	}
}

// CreatePod handles POST /api/v1/pods
// @Summary Create a new Knowledge Pod
// @Description Create a new Knowledge Pod. Only verified users (teachers) can create pods.
// @Tags Pods
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body application.CreatePodInput true "Pod creation data"
// @Success 201 {object} response.Response[domain.Pod] "Pod created successfully"
// @Failure 400 {object} errors.ErrorResponse "Invalid request body"
// @Failure 401 {object} errors.ErrorResponse "Authentication required"
// @Failure 403 {object} errors.ErrorResponse "Only verified users can create pods"
// @Router /pods [post]
func (h *PodHandler) CreatePod(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok || userID == uuid.Nil {
		return errors.Unauthorized("authentication required")
	}

	var input application.CreatePodInput
	if err := c.BodyParser(&input); err != nil {
		return errors.BadRequest("invalid request body")
	}

	if err := h.validator.Struct(input); err != nil {
		return err
	}

	input.OwnerID = userID

	pod, err := h.podService.CreatePod(c.Context(), input)
	if err != nil {
		return err
	}

	requestID := middleware.GetRequestID(c)
	return c.Status(fiber.StatusCreated).JSON(response.Created(requestID, pod))
}

// GetPod handles GET /api/v1/pods/:id
// @Summary Get a Knowledge Pod
// @Description Get a Knowledge Pod by ID. Private pods require authentication and access.
// @Tags Pods
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Pod ID" format(uuid)
// @Success 200 {object} response.Response[domain.Pod] "Pod details"
// @Failure 400 {object} errors.ErrorResponse "Invalid pod ID"
// @Failure 403 {object} errors.ErrorResponse "Access denied to private pod"
// @Failure 404 {object} errors.ErrorResponse "Pod not found"
// @Router /pods/{id} [get]
func (h *PodHandler) GetPod(c *fiber.Ctx) error {
	// Pod ID is extracted and validated by middleware
	podID, ok := GetPodID(c)
	if !ok {
		// Fallback to parsing from params if middleware not used
		idParam := c.Params("id")
		var err error
		podID, err = uuid.Parse(idParam)
		if err != nil {
			return errors.BadRequest("invalid pod ID")
		}
	}

	var viewerID *uuid.UUID
	if uid, ok := middleware.GetUserID(c); ok && uid != uuid.Nil {
		viewerID = &uid
	}

	pod, err := h.podService.GetPod(c.Context(), podID, viewerID)
	if err != nil {
		return err
	}

	// Auto-track view interaction for authenticated users
	if viewerID != nil {
		go h.recommendationService.TrackView(c.Context(), *viewerID, podID)
	}

	requestID := middleware.GetRequestID(c)
	return c.JSON(response.OK(requestID, pod))
}

// ListPods handles GET /api/v1/pods
// @Summary List Knowledge Pods
// @Description Get a paginated list of Knowledge Pods with optional filters
// @Tags Pods
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param per_page query int false "Items per page" default(20)
// @Param owner_id query string false "Filter by owner ID" format(uuid)
// @Param category query string false "Filter by category"
// @Param visibility query string false "Filter by visibility" Enums(public, private)
// @Success 200 {object} response.PaginatedResponse[domain.Pod] "List of pods"
// @Router /pods [get]
func (h *PodHandler) ListPods(c *fiber.Ctx) error {
	page := c.QueryInt("page", 1)
	perPage := c.QueryInt("per_page", 20)

	filters := domain.PodFilters{}

	if ownerID := c.Query("owner_id"); ownerID != "" {
		if uid, err := uuid.Parse(ownerID); err == nil {
			filters.OwnerID = &uid
		}
	}

	if category := c.Query("category"); category != "" {
		filters.Category = &category
	}

	if visibility := c.Query("visibility"); visibility != "" {
		v := domain.Visibility(visibility)
		filters.Visibility = &v
	}

	result, err := h.podService.ListPods(c.Context(), filters, page, perPage)
	if err != nil {
		return err
	}

	requestID := middleware.GetRequestID(c)
	return c.JSON(response.List(requestID, result.Pods, result.Page, result.PerPage, result.Total))
}

// UpdatePod handles PUT /api/v1/pods/:id
// @Summary Update a Knowledge Pod
// @Description Update a Knowledge Pod. Requires owner or admin collaborator access.
// @Tags Pods
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Pod ID" format(uuid)
// @Param request body application.UpdatePodInput true "Pod update data"
// @Success 200 {object} response.Response[domain.Pod] "Updated pod"
// @Failure 400 {object} errors.ErrorResponse "Invalid request body"
// @Failure 401 {object} errors.ErrorResponse "Authentication required"
// @Failure 403 {object} errors.ErrorResponse "Edit access required"
// @Failure 404 {object} errors.ErrorResponse "Pod not found"
// @Router /pods/{id} [put]
func (h *PodHandler) UpdatePod(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok || userID == uuid.Nil {
		return errors.Unauthorized("authentication required")
	}

	// Pod ID is extracted and validated by middleware
	podID, ok := GetPodID(c)
	if !ok {
		var err error
		podID, err = uuid.Parse(c.Params("id"))
		if err != nil {
			return errors.BadRequest("invalid pod ID")
		}
	}

	var input application.UpdatePodInput
	if err := c.BodyParser(&input); err != nil {
		return errors.BadRequest("invalid request body")
	}

	if err := h.validator.Struct(input); err != nil {
		return err
	}

	pod, err := h.podService.UpdatePod(c.Context(), podID, userID, input)
	if err != nil {
		return err
	}

	requestID := middleware.GetRequestID(c)
	return c.JSON(response.OK(requestID, pod))
}

// DeletePod handles DELETE /api/v1/pods/:id
// @Summary Delete a Knowledge Pod
// @Description Delete a Knowledge Pod. Only the owner can delete a pod.
// @Tags Pods
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Pod ID" format(uuid)
// @Success 204 "Pod deleted successfully"
// @Failure 401 {object} errors.ErrorResponse "Authentication required"
// @Failure 403 {object} errors.ErrorResponse "Owner access required"
// @Failure 404 {object} errors.ErrorResponse "Pod not found"
// @Router /pods/{id} [delete]
func (h *PodHandler) DeletePod(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok || userID == uuid.Nil {
		return errors.Unauthorized("authentication required")
	}

	// Pod ID is extracted and validated by middleware
	podID, ok := GetPodID(c)
	if !ok {
		var err error
		podID, err = uuid.Parse(c.Params("id"))
		if err != nil {
			return errors.BadRequest("invalid pod ID")
		}
	}

	if err := h.podService.DeletePod(c.Context(), podID, userID); err != nil {
		return err
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// ForkPod handles POST /api/v1/pods/:id/fork
// @Summary Fork a Knowledge Pod
// @Description Create a copy of a Knowledge Pod. Only verified users can fork pods.
// @Tags Pods
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Pod ID to fork" format(uuid)
// @Success 201 {object} response.Response[domain.Pod] "Forked pod"
// @Failure 401 {object} errors.ErrorResponse "Authentication required"
// @Failure 403 {object} errors.ErrorResponse "Access denied or not verified"
// @Failure 404 {object} errors.ErrorResponse "Pod not found"
// @Router /pods/{id}/fork [post]
func (h *PodHandler) ForkPod(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok || userID == uuid.Nil {
		return errors.Unauthorized("authentication required")
	}

	// Pod ID is extracted and validated by middleware
	podID, ok := GetPodID(c)
	if !ok {
		var err error
		podID, err = uuid.Parse(c.Params("id"))
		if err != nil {
			return errors.BadRequest("invalid pod ID")
		}
	}

	pod, err := h.podService.ForkPod(c.Context(), podID, userID)
	if err != nil {
		return err
	}

	// Auto-track fork interaction (on source pod, not the new fork)
	go h.recommendationService.TrackInteraction(c.Context(), application.TrackInteractionInput{
		UserID:          userID,
		PodID:           podID,
		InteractionType: domain.InteractionFork,
	})

	requestID := middleware.GetRequestID(c)
	return c.Status(fiber.StatusCreated).JSON(response.Created(requestID, pod))
}

// StarPod handles POST /api/v1/pods/:id/star
// @Summary Star a Knowledge Pod
// @Description Add a pod to your starred list
// @Tags Pods
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Pod ID" format(uuid)
// @Success 200 {object} response.Response[map[string]bool] "Pod starred"
// @Failure 401 {object} errors.ErrorResponse "Authentication required"
// @Failure 404 {object} errors.ErrorResponse "Pod not found"
// @Failure 409 {object} errors.ErrorResponse "Already starred"
// @Router /pods/{id}/star [post]
func (h *PodHandler) StarPod(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok || userID == uuid.Nil {
		return errors.Unauthorized("authentication required")
	}

	// Pod ID is extracted and validated by middleware
	podID, ok := GetPodID(c)
	if !ok {
		var err error
		podID, err = uuid.Parse(c.Params("id"))
		if err != nil {
			return errors.BadRequest("invalid pod ID")
		}
	}

	if err := h.podService.StarPod(c.Context(), podID, userID); err != nil {
		return err
	}

	// Auto-track star interaction
	go h.recommendationService.TrackInteraction(c.Context(), application.TrackInteractionInput{
		UserID:          userID,
		PodID:           podID,
		InteractionType: domain.InteractionStar,
	})

	requestID := middleware.GetRequestID(c)
	return c.JSON(response.OK(requestID, fiber.Map{"starred": true}))
}

// UnstarPod handles DELETE /api/v1/pods/:id/star
// @Summary Unstar a Knowledge Pod
// @Description Remove a pod from your starred list
// @Tags Pods
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Pod ID" format(uuid)
// @Success 200 {object} response.Response[map[string]bool] "Pod unstarred"
// @Failure 401 {object} errors.ErrorResponse "Authentication required"
// @Failure 404 {object} errors.ErrorResponse "Pod not found or not starred"
// @Router /pods/{id}/star [delete]
func (h *PodHandler) UnstarPod(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok || userID == uuid.Nil {
		return errors.Unauthorized("authentication required")
	}

	// Pod ID is extracted and validated by middleware
	podID, ok := GetPodID(c)
	if !ok {
		var err error
		podID, err = uuid.Parse(c.Params("id"))
		if err != nil {
			return errors.BadRequest("invalid pod ID")
		}
	}

	if err := h.podService.UnstarPod(c.Context(), podID, userID); err != nil {
		return err
	}

	// Auto-track unstar interaction
	go h.recommendationService.TrackInteraction(c.Context(), application.TrackInteractionInput{
		UserID:          userID,
		PodID:           podID,
		InteractionType: domain.InteractionUnstar,
	})

	requestID := middleware.GetRequestID(c)
	return c.JSON(response.OK(requestID, fiber.Map{"starred": false}))
}

// FollowPod handles POST /api/v1/pods/:id/follow
// @Summary Follow a Knowledge Pod
// @Description Follow a pod to receive notifications about new materials
// @Tags Pods
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Pod ID" format(uuid)
// @Success 200 {object} response.Response[map[string]bool] "Pod followed"
// @Failure 401 {object} errors.ErrorResponse "Authentication required"
// @Failure 404 {object} errors.ErrorResponse "Pod not found"
// @Router /pods/{id}/follow [post]
func (h *PodHandler) FollowPod(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok || userID == uuid.Nil {
		return errors.Unauthorized("authentication required")
	}

	// Pod ID is extracted and validated by middleware
	podID, ok := GetPodID(c)
	if !ok {
		var err error
		podID, err = uuid.Parse(c.Params("id"))
		if err != nil {
			return errors.BadRequest("invalid pod ID")
		}
	}

	if err := h.podService.FollowPod(c.Context(), podID, userID); err != nil {
		return err
	}

	// Auto-track follow interaction
	go h.recommendationService.TrackInteraction(c.Context(), application.TrackInteractionInput{
		UserID:          userID,
		PodID:           podID,
		InteractionType: domain.InteractionFollow,
	})

	requestID := middleware.GetRequestID(c)
	return c.JSON(response.OK(requestID, fiber.Map{"following": true}))
}

// UnfollowPod handles DELETE /api/v1/pods/:id/follow
// @Summary Unfollow a Knowledge Pod
// @Description Stop following a pod
// @Tags Pods
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Pod ID" format(uuid)
// @Success 200 {object} response.Response[map[string]bool] "Pod unfollowed"
// @Failure 401 {object} errors.ErrorResponse "Authentication required"
// @Failure 404 {object} errors.ErrorResponse "Pod not found"
// @Router /pods/{id}/follow [delete]
func (h *PodHandler) UnfollowPod(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok || userID == uuid.Nil {
		return errors.Unauthorized("authentication required")
	}

	// Pod ID is extracted and validated by middleware
	podID, ok := GetPodID(c)
	if !ok {
		var err error
		podID, err = uuid.Parse(c.Params("id"))
		if err != nil {
			return errors.BadRequest("invalid pod ID")
		}
	}

	if err := h.podService.UnfollowPod(c.Context(), podID, userID); err != nil {
		return err
	}

	// Auto-track unfollow interaction
	go h.recommendationService.TrackInteraction(c.Context(), application.TrackInteractionInput{
		UserID:          userID,
		PodID:           podID,
		InteractionType: domain.InteractionUnfollow,
	})

	requestID := middleware.GetRequestID(c)
	return c.JSON(response.OK(requestID, fiber.Map{"following": false}))
}

// InviteCollaborator handles POST /api/v1/pods/:id/collaborators
// @Summary Invite a collaborator
// @Description Invite a user to collaborate on a pod. Requires owner or admin access.
// @Tags Pods
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Pod ID" format(uuid)
// @Param request body application.InviteCollaboratorInput true "Collaborator invitation"
// @Success 201 {object} response.Response[domain.Collaborator] "Collaborator invited"
// @Failure 400 {object} errors.ErrorResponse "Invalid request body"
// @Failure 401 {object} errors.ErrorResponse "Authentication required"
// @Failure 403 {object} errors.ErrorResponse "Collaborator management access required"
// @Failure 404 {object} errors.ErrorResponse "Pod or user not found"
// @Router /pods/{id}/collaborators [post]
func (h *PodHandler) InviteCollaborator(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok || userID == uuid.Nil {
		return errors.Unauthorized("authentication required")
	}

	// Pod ID is extracted and validated by middleware
	podID, ok := GetPodID(c)
	if !ok {
		var err error
		podID, err = uuid.Parse(c.Params("id"))
		if err != nil {
			return errors.BadRequest("invalid pod ID")
		}
	}

	var input application.InviteCollaboratorInput
	if err := c.BodyParser(&input); err != nil {
		return errors.BadRequest("invalid request body")
	}

	if err := h.validator.Struct(input); err != nil {
		return err
	}

	input.PodID = podID
	input.InviterID = userID

	collaborator, err := h.podService.InviteCollaborator(c.Context(), input)
	if err != nil {
		return err
	}

	requestID := middleware.GetRequestID(c)
	return c.Status(fiber.StatusCreated).JSON(response.Created(requestID, collaborator))
}

// UpdateCollaborator handles PUT /api/v1/pods/:id/collaborators/:userId
// @Summary Update a collaborator
// @Description Verify a collaborator or update their role. Requires owner or admin access.
// @Tags Pods
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Pod ID" format(uuid)
// @Param userId path string true "Collaborator user ID" format(uuid)
// @Param request body object{action=string,role=string} true "Action (verify/update_role) and optional role"
// @Success 200 {object} response.Response[map[string]bool] "Collaborator updated"
// @Failure 400 {object} errors.ErrorResponse "Invalid request body or action"
// @Failure 401 {object} errors.ErrorResponse "Authentication required"
// @Failure 403 {object} errors.ErrorResponse "Collaborator management access required"
// @Failure 404 {object} errors.ErrorResponse "Pod or collaborator not found"
// @Router /pods/{id}/collaborators/{userId} [put]
func (h *PodHandler) UpdateCollaborator(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok || userID == uuid.Nil {
		return errors.Unauthorized("authentication required")
	}

	// Pod ID is extracted and validated by middleware
	podID, ok := GetPodID(c)
	if !ok {
		var err error
		podID, err = uuid.Parse(c.Params("id"))
		if err != nil {
			return errors.BadRequest("invalid pod ID")
		}
	}

	collaboratorUserID, err := uuid.Parse(c.Params("userId"))
	if err != nil {
		return errors.BadRequest("invalid user ID")
	}

	var input struct {
		Action string                  `json:"action"` // "verify" or "update_role"
		Role   domain.CollaboratorRole `json:"role,omitempty"`
	}
	if err := c.BodyParser(&input); err != nil {
		return errors.BadRequest("invalid request body")
	}

	// Get collaborator by pod and user
	collaborators, err := h.podService.GetCollaborators(c.Context(), podID)
	if err != nil {
		return err
	}

	var collaboratorID uuid.UUID
	for _, collab := range collaborators {
		if collab.UserID == collaboratorUserID {
			collaboratorID = collab.ID
			break
		}
	}

	if collaboratorID == uuid.Nil {
		return errors.NotFound("collaborator", collaboratorUserID.String())
	}

	switch input.Action {
	case "verify":
		if err := h.podService.VerifyCollaborator(c.Context(), podID, collaboratorID, userID); err != nil {
			return err
		}
	case "update_role":
		if input.Role == "" {
			return errors.BadRequest("role is required for update_role action")
		}
		if err := h.podService.UpdateCollaboratorRole(c.Context(), podID, collaboratorID, userID, input.Role); err != nil {
			return err
		}
	default:
		return errors.BadRequest("invalid action, must be 'verify' or 'update_role'")
	}

	requestID := middleware.GetRequestID(c)
	return c.JSON(response.OK(requestID, fiber.Map{"updated": true}))
}

// RemoveCollaborator handles DELETE /api/v1/pods/:id/collaborators/:userId
// @Summary Remove a collaborator
// @Description Remove a collaborator from a pod. Requires owner or admin access.
// @Tags Pods
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Pod ID" format(uuid)
// @Param userId path string true "Collaborator user ID" format(uuid)
// @Success 204 "Collaborator removed"
// @Failure 401 {object} errors.ErrorResponse "Authentication required"
// @Failure 403 {object} errors.ErrorResponse "Collaborator management access required"
// @Failure 404 {object} errors.ErrorResponse "Pod or collaborator not found"
// @Router /pods/{id}/collaborators/{userId} [delete]
func (h *PodHandler) RemoveCollaborator(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok || userID == uuid.Nil {
		return errors.Unauthorized("authentication required")
	}

	// Pod ID is extracted and validated by middleware
	podID, ok := GetPodID(c)
	if !ok {
		var err error
		podID, err = uuid.Parse(c.Params("id"))
		if err != nil {
			return errors.BadRequest("invalid pod ID")
		}
	}

	collaboratorUserID, err := uuid.Parse(c.Params("userId"))
	if err != nil {
		return errors.BadRequest("invalid user ID")
	}

	// Get collaborator by pod and user
	collaborators, err := h.podService.GetCollaborators(c.Context(), podID)
	if err != nil {
		return err
	}

	var collaboratorID uuid.UUID
	for _, collab := range collaborators {
		if collab.UserID == collaboratorUserID {
			collaboratorID = collab.ID
			break
		}
	}

	if collaboratorID == uuid.Nil {
		return errors.NotFound("collaborator", collaboratorUserID.String())
	}

	if err := h.podService.RemoveCollaborator(c.Context(), podID, collaboratorID, userID); err != nil {
		return err
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// GetCollaborators handles GET /api/v1/pods/:id/collaborators
// @Summary Get pod collaborators
// @Description Get a list of collaborators for a pod
// @Tags Pods
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Pod ID" format(uuid)
// @Success 200 {object} response.Response[[]domain.Collaborator] "List of collaborators"
// @Failure 400 {object} errors.ErrorResponse "Invalid pod ID"
// @Failure 403 {object} errors.ErrorResponse "Access denied"
// @Failure 404 {object} errors.ErrorResponse "Pod not found"
// @Router /pods/{id}/collaborators [get]
func (h *PodHandler) GetCollaborators(c *fiber.Ctx) error {
	// Pod ID is extracted and validated by middleware
	podID, ok := GetPodID(c)
	if !ok {
		var err error
		podID, err = uuid.Parse(c.Params("id"))
		if err != nil {
			return errors.BadRequest("invalid pod ID")
		}
	}

	collaborators, err := h.podService.GetCollaborators(c.Context(), podID)
	if err != nil {
		return err
	}

	requestID := middleware.GetRequestID(c)
	return c.JSON(response.OK(requestID, collaborators))
}

// GetPodActivity handles GET /api/v1/pods/:id/activity
// @Summary Get pod activity
// @Description Get a paginated list of activities for a pod
// @Tags Pods
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Pod ID" format(uuid)
// @Param page query int false "Page number" default(1)
// @Param per_page query int false "Items per page" default(20)
// @Success 200 {object} response.PaginatedResponse[domain.Activity] "List of activities"
// @Failure 400 {object} errors.ErrorResponse "Invalid pod ID"
// @Failure 403 {object} errors.ErrorResponse "Access denied"
// @Failure 404 {object} errors.ErrorResponse "Pod not found"
// @Router /pods/{id}/activity [get]
func (h *PodHandler) GetPodActivity(c *fiber.Ctx) error {
	// Pod ID is extracted and validated by middleware
	podID, ok := GetPodID(c)
	if !ok {
		var err error
		podID, err = uuid.Parse(c.Params("id"))
		if err != nil {
			return errors.BadRequest("invalid pod ID")
		}
	}

	page := c.QueryInt("page", 1)
	perPage := c.QueryInt("per_page", 20)

	result, err := h.podService.GetPodActivity(c.Context(), podID, page, perPage)
	if err != nil {
		return err
	}

	requestID := middleware.GetRequestID(c)
	return c.JSON(response.List(requestID, result.Activities, result.Page, result.PerPage, result.Total))
}

// GetUserFeed handles GET /api/v1/feed
// @Summary Get user activity feed
// @Description Get a paginated activity feed from followed pods and users
// @Tags Pods
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param per_page query int false "Items per page" default(20)
// @Success 200 {object} response.PaginatedResponse[domain.Activity] "Activity feed"
// @Failure 401 {object} errors.ErrorResponse "Authentication required"
// @Router /feed [get]
func (h *PodHandler) GetUserFeed(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok || userID == uuid.Nil {
		return errors.Unauthorized("authentication required")
	}

	page := c.QueryInt("page", 1)
	perPage := c.QueryInt("per_page", 20)

	result, err := h.podService.GetUserFeed(c.Context(), userID, page, perPage)
	if err != nil {
		return err
	}

	requestID := middleware.GetRequestID(c)
	return c.JSON(response.List(requestID, result.Activities, result.Page, result.PerPage, result.Total))
}

// GetUserPods handles GET /api/v1/users/:id/pods
// @Summary Get user's pods
// @Description Get a paginated list of pods owned by a user
// @Tags Pods
// @Accept json
// @Produce json
// @Param id path string true "User ID" format(uuid)
// @Param page query int false "Page number" default(1)
// @Param per_page query int false "Items per page" default(20)
// @Success 200 {object} response.PaginatedResponse[domain.Pod] "List of user's pods"
// @Failure 400 {object} errors.ErrorResponse "Invalid user ID"
// @Router /users/{id}/pods [get]
func (h *PodHandler) GetUserPods(c *fiber.Ctx) error {
	ownerID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return errors.BadRequest("invalid user ID")
	}

	page := c.QueryInt("page", 1)
	perPage := c.QueryInt("per_page", 20)

	result, err := h.podService.ListUserPods(c.Context(), ownerID, page, perPage)
	if err != nil {
		return err
	}

	requestID := middleware.GetRequestID(c)
	return c.JSON(response.List(requestID, result.Pods, result.Page, result.PerPage, result.Total))
}

// GetUserStarredPods handles GET /api/v1/users/:id/starred
// @Summary Get user's starred pods
// @Description Get a paginated list of pods starred by a user
// @Tags Pods
// @Accept json
// @Produce json
// @Param id path string true "User ID" format(uuid)
// @Param page query int false "Page number" default(1)
// @Param per_page query int false "Items per page" default(20)
// @Success 200 {object} response.PaginatedResponse[domain.Pod] "List of starred pods"
// @Failure 400 {object} errors.ErrorResponse "Invalid user ID"
// @Router /users/{id}/starred [get]
func (h *PodHandler) GetUserStarredPods(c *fiber.Ctx) error {
	userID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return errors.BadRequest("invalid user ID")
	}

	page := c.QueryInt("page", 1)
	perPage := c.QueryInt("per_page", 20)

	result, err := h.podService.GetStarredPods(c.Context(), userID, page, perPage)
	if err != nil {
		return err
	}

	requestID := middleware.GetRequestID(c)
	return c.JSON(response.List(requestID, result.Pods, result.Page, result.PerPage, result.Total))
}

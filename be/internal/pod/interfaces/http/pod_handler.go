package http

import (
	"fmt"

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

// CollaboratorInviteResponse is the response body for POST /api/v1/pods/:id/collaborators.
type CollaboratorInviteResponse struct {
	CreatedAt string `json:"created_at"`
	ID        string `json:"id"`
	InvitedBy string `json:"invited_by"`
	PodID     string `json:"pod_id"`
	Role      string `json:"role"`
	Status    string `json:"status"`
	UpdatedAt string `json:"updated_at"`
	UserID    string `json:"user_id"`
}

func toCollaboratorInviteResponse(c *domain.Collaborator) CollaboratorInviteResponse {
	return CollaboratorInviteResponse{
		CreatedAt: c.CreatedAt.UTC().Format("2006-01-02T15:04:05Z07:00"),
		ID:        c.ID.String(),
		InvitedBy: c.InvitedBy.String(),
		PodID:     c.PodID.String(),
		Role:      string(c.Role),
		Status:    string(c.Status),
		UpdatedAt: c.UpdatedAt.UTC().Format("2006-01-02T15:04:05Z07:00"),
		UserID:    c.UserID.String(),
	}
}

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

// GetPodBySlug handles GET /api/v1/pods/slug/:slug
// @Summary Get a Knowledge Pod by slug
// @Description Get a Knowledge Pod by its URL-friendly slug. Private pods require authentication and access.
// @Tags Pods
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param slug path string true "Pod slug"
// @Success 200 {object} response.Response[domain.Pod] "Pod details"
// @Failure 403 {object} errors.ErrorResponse "Access denied to private pod"
// @Failure 404 {object} errors.ErrorResponse "Pod not found"
// @Router /pods/slug/{slug} [get]
func (h *PodHandler) GetPodBySlug(c *fiber.Ctx) error {
	slug := c.Params("slug")
	if slug == "" {
		return errors.BadRequest("slug is required")
	}

	var viewerID *uuid.UUID
	if uid, ok := middleware.GetUserID(c); ok && uid != uuid.Nil {
		viewerID = &uid
	}

	pod, err := h.podService.GetPodBySlug(c.Context(), slug, viewerID)
	if err != nil {
		return err
	}

	// Auto-track view interaction for authenticated users
	if viewerID != nil {
		go h.recommendationService.TrackView(c.Context(), *viewerID, pod.ID)
	}

	requestID := middleware.GetRequestID(c)
	return c.JSON(response.OK(requestID, pod))
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
	println("=== HANDLER START: GetPod ===")

	// Pod ID is extracted and validated by middleware
	podID, ok := GetPodID(c)
	println("GetPodID result:", ok, "podID:", podID.String())
	if !ok {
		// Fallback to parsing from params if middleware not used
		idParam := c.Params("id")
		println("Parsing podID from params:", idParam)
		var err error
		podID, err = uuid.Parse(idParam)
		if err != nil {
			println("ERROR: Invalid pod ID:", err.Error())
			return errors.BadRequest("invalid pod ID")
		}
	}

	var viewerID *uuid.UUID
	if uid, ok := middleware.GetUserID(c); ok && uid != uuid.Nil {
		viewerID = &uid
		println("Got viewerID:", uid.String())
	} else {
		println("No viewerID (public request)")
	}

	println("Calling podService.GetPod with podID:", podID.String())
	pod, err := h.podService.GetPod(c.Context(), podID, viewerID)
	if err != nil {
		println("ERROR from GetPod service:", err.Error())
		// Log the actual error for debugging
		c.App().Config().ErrorHandler(c, err)
		return err
	}
	println("Got pod successfully:", pod.Name)

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
// @Param request body application.InviteCollaboratorInput true "Collaborator invitation (role, user_id)"
// @Success 201 {object} response.Response[CollaboratorInviteResponse] "Collaborator invited"
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
	return c.Status(fiber.StatusCreated).JSON(response.Created(requestID, toCollaboratorInviteResponse(collaborator)))
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
	fmt.Println(userID)
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

// UpvotePod handles POST /api/v1/pods/:id/upvote
// @Summary Upvote a Knowledge Pod
// @Description Add an upvote to a pod as a trust indicator. Each user can only upvote a pod once.
// @Tags Pods
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Pod ID" format(uuid)
// @Success 200 {object} response.Response[map[string]bool] "Pod upvoted"
// @Failure 401 {object} errors.ErrorResponse "Authentication required"
// @Failure 404 {object} errors.ErrorResponse "Pod not found"
// @Failure 409 {object} errors.ErrorResponse "Already upvoted"
// @Router /pods/{id}/upvote [post]
func (h *PodHandler) UpvotePod(c *fiber.Ctx) error {
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

	if err := h.podService.UpvotePod(c.Context(), podID, userID); err != nil {
		return err
	}

	// Auto-track upvote interaction
	go h.recommendationService.TrackInteraction(c.Context(), application.TrackInteractionInput{
		UserID:          userID,
		PodID:           podID,
		InteractionType: domain.InteractionUpvote,
	})

	requestID := middleware.GetRequestID(c)
	return c.JSON(response.OK(requestID, fiber.Map{"upvoted": true}))
}

// RemoveUpvote handles DELETE /api/v1/pods/:id/upvote
// @Summary Remove upvote from a Knowledge Pod
// @Description Remove your upvote from a pod
// @Tags Pods
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Pod ID" format(uuid)
// @Success 200 {object} response.Response[map[string]bool] "Upvote removed"
// @Failure 401 {object} errors.ErrorResponse "Authentication required"
// @Failure 404 {object} errors.ErrorResponse "Pod not found or not upvoted"
// @Router /pods/{id}/upvote [delete]
func (h *PodHandler) RemoveUpvote(c *fiber.Ctx) error {
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

	if err := h.podService.RemoveUpvote(c.Context(), podID, userID); err != nil {
		return err
	}

	// Auto-track remove upvote interaction
	go h.recommendationService.TrackInteraction(c.Context(), application.TrackInteractionInput{
		UserID:          userID,
		PodID:           podID,
		InteractionType: domain.InteractionRemoveUpvote,
	})

	requestID := middleware.GetRequestID(c)
	return c.JSON(response.OK(requestID, fiber.Map{"upvoted": false}))
}

// GetUserUpvotedPods handles GET /api/v1/users/me/upvoted-pods
// @Summary Get current user's upvoted pods
// @Description Get a paginated list of pods upvoted by the current user
// @Tags Pods
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param per_page query int false "Items per page" default(20)
// @Success 200 {object} response.PaginatedResponse[domain.Pod] "List of upvoted pods"
// @Failure 401 {object} errors.ErrorResponse "Authentication required"
// @Router /users/me/upvoted-pods [get]
func (h *PodHandler) GetUserUpvotedPods(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok || userID == uuid.Nil {
		return errors.Unauthorized("authentication required")
	}

	page := c.QueryInt("page", 1)
	perPage := c.QueryInt("per_page", 20)

	result, err := h.podService.GetUpvotedPods(c.Context(), userID, page, perPage)
	if err != nil {
		return err
	}

	requestID := middleware.GetRequestID(c)
	return c.JSON(response.List(requestID, result.Pods, result.Page, result.PerPage, result.Total))
}

// DownvotePod handles POST /api/v1/pods/:id/downvote
// @Summary Downvote a Knowledge Pod
// @Description Add a downvote to a pod as a negative trust indicator. Each user can only downvote a pod once.
// @Tags Pods
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Pod ID" format(uuid)
// @Success 200 {object} response.Response[map[string]bool] "Pod downvoted"
// @Failure 401 {object} errors.ErrorResponse "Authentication required"
// @Failure 404 {object} errors.ErrorResponse "Pod not found"
// @Failure 409 {object} errors.ErrorResponse "Already downvoted"
// @Router /pods/{id}/downvote [post]
func (h *PodHandler) DownvotePod(c *fiber.Ctx) error {
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

	if err := h.podService.DownvotePod(c.Context(), podID, userID); err != nil {
		return err
	}

	// Auto-track downvote interaction
	go h.recommendationService.TrackInteraction(c.Context(), application.TrackInteractionInput{
		UserID:          userID,
		PodID:           podID,
		InteractionType: domain.InteractionDownvote,
	})

	requestID := middleware.GetRequestID(c)
	return c.JSON(response.OK(requestID, fiber.Map{"downvoted": true}))
}

// RemoveDownvote handles DELETE /api/v1/pods/:id/downvote
// @Summary Remove downvote from a Knowledge Pod
// @Description Remove your downvote from a pod
// @Tags Pods
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Pod ID" format(uuid)
// @Success 200 {object} response.Response[map[string]bool] "Downvote removed"
// @Failure 401 {object} errors.ErrorResponse "Authentication required"
// @Failure 404 {object} errors.ErrorResponse "Pod not found or not downvoted"
// @Router /pods/{id}/downvote [delete]
func (h *PodHandler) RemoveDownvote(c *fiber.Ctx) error {
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

	if err := h.podService.RemoveDownvote(c.Context(), podID, userID); err != nil {
		return err
	}

	// Auto-track remove downvote interaction
	go h.recommendationService.TrackInteraction(c.Context(), application.TrackInteractionInput{
		UserID:          userID,
		PodID:           podID,
		InteractionType: domain.InteractionRemoveDownvote,
	})

	requestID := middleware.GetRequestID(c)
	return c.JSON(response.OK(requestID, fiber.Map{"downvoted": false}))
}

// GetUserDownvotedPods handles GET /api/v1/users/me/downvoted-pods
// @Summary Get current user's downvoted pods
// @Description Get a paginated list of pods downvoted by the current user
// @Tags Pods
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param per_page query int false "Items per page" default(20)
// @Success 200 {object} response.PaginatedResponse[domain.Pod] "List of downvoted pods"
// @Failure 401 {object} errors.ErrorResponse "Authentication required"
// @Router /users/me/downvoted-pods [get]
func (h *PodHandler) GetUserDownvotedPods(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok || userID == uuid.Nil {
		return errors.Unauthorized("authentication required")
	}

	page := c.QueryInt("page", 1)
	perPage := c.QueryInt("per_page", 20)

	result, err := h.podService.GetDownvotedPods(c.Context(), userID, page, perPage)
	if err != nil {
		return err
	}

	requestID := middleware.GetRequestID(c)
	return c.JSON(response.List(requestID, result.Pods, result.Page, result.PerPage, result.Total))
}

// CreateUploadRequestInput represents the input for creating an upload request.
// @Description Create upload request input
type CreateUploadRequestInput struct {
	Message *string `json:"message,omitempty" example:"I would like to contribute additional materials on advanced topics."`
}

// CreateUploadRequest handles POST /api/v1/pods/:id/upload-request
// @Summary Request upload permission to a pod
// @Description Request permission to upload materials to another teacher's pod. Only teachers can make this request.
// @Tags Upload Requests
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Pod ID" format(uuid)
// @Param request body CreateUploadRequestInput false "Optional message for the request"
// @Success 201 {object} response.Response[domain.UploadRequest] "Upload request created"
// @Failure 400 {object} errors.ErrorResponse "Invalid request body"
// @Failure 401 {object} errors.ErrorResponse "Authentication required"
// @Failure 403 {object} errors.ErrorResponse "Only teachers can request upload permission"
// @Failure 404 {object} errors.ErrorResponse "Pod not found"
// @Failure 409 {object} errors.ErrorResponse "Upload request already exists"
// @Router /pods/{id}/upload-request [post]
func (h *PodHandler) CreateUploadRequest(c *fiber.Ctx) error {
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

	var input CreateUploadRequestInput
	if err := c.BodyParser(&input); err != nil {
		// Body is optional, so ignore parse errors for empty body
		input = CreateUploadRequestInput{}
	}

	uploadRequest, err := h.podService.CreateUploadRequest(c.Context(), userID, podID, input.Message)
	if err != nil {
		return err
	}

	requestID := middleware.GetRequestID(c)
	return c.Status(fiber.StatusCreated).JSON(response.Created(requestID, uploadRequest))
}

// GetUploadRequests handles GET /api/v1/users/me/upload-requests
// @Summary Get upload requests for the current user
// @Description Get a paginated list of upload requests received by the current user (as pod owner)
// @Tags Upload Requests
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param per_page query int false "Items per page" default(20)
// @Param status query string false "Filter by status" Enums(pending, approved, rejected, revoked)
// @Success 200 {object} response.PaginatedResponse[domain.UploadRequest] "List of upload requests"
// @Failure 401 {object} errors.ErrorResponse "Authentication required"
// @Router /users/me/upload-requests [get]
func (h *PodHandler) GetUploadRequests(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok || userID == uuid.Nil {
		return errors.Unauthorized("authentication required")
	}

	page := c.QueryInt("page", 1)
	perPage := c.QueryInt("per_page", 20)

	var status *domain.UploadRequestStatus
	if statusStr := c.Query("status"); statusStr != "" {
		s := domain.UploadRequestStatus(statusStr)
		status = &s
	}

	result, err := h.podService.GetUploadRequestsForOwner(c.Context(), userID, status, page, perPage)
	if err != nil {
		return err
	}

	requestID := middleware.GetRequestID(c)
	return c.JSON(response.List(requestID, result.UploadRequests, result.Page, result.PerPage, result.Total))
}

// ApproveUploadRequest handles POST /api/v1/upload-requests/:id/approve
// @Summary Approve an upload request
// @Description Approve an upload request, granting the requester permission to upload to your pod
// @Tags Upload Requests
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Upload Request ID" format(uuid)
// @Success 200 {object} response.Response[map[string]bool] "Upload request approved"
// @Failure 400 {object} errors.ErrorResponse "Invalid request ID"
// @Failure 401 {object} errors.ErrorResponse "Authentication required"
// @Failure 403 {object} errors.ErrorResponse "Not the pod owner"
// @Failure 404 {object} errors.ErrorResponse "Upload request not found"
// @Router /upload-requests/{id}/approve [post]
func (h *PodHandler) ApproveUploadRequest(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok || userID == uuid.Nil {
		return errors.Unauthorized("authentication required")
	}

	requestID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return errors.BadRequest("invalid request ID")
	}

	if err := h.podService.ApproveUploadRequest(c.Context(), requestID, userID); err != nil {
		return err
	}

	reqID := middleware.GetRequestID(c)
	return c.JSON(response.OK(reqID, fiber.Map{"approved": true}))
}

// RejectUploadRequestInput represents the input for rejecting an upload request.
// @Description Reject upload request input
type RejectUploadRequestInput struct {
	Reason *string `json:"reason,omitempty" example:"Not accepting contributions at this time."`
}

// RejectUploadRequest handles POST /api/v1/upload-requests/:id/reject
// @Summary Reject an upload request
// @Description Reject an upload request with an optional reason
// @Tags Upload Requests
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Upload Request ID" format(uuid)
// @Param request body RejectUploadRequestInput false "Optional rejection reason"
// @Success 200 {object} response.Response[map[string]bool] "Upload request rejected"
// @Failure 400 {object} errors.ErrorResponse "Invalid request ID"
// @Failure 401 {object} errors.ErrorResponse "Authentication required"
// @Failure 403 {object} errors.ErrorResponse "Not the pod owner"
// @Failure 404 {object} errors.ErrorResponse "Upload request not found"
// @Router /upload-requests/{id}/reject [post]
func (h *PodHandler) RejectUploadRequest(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok || userID == uuid.Nil {
		return errors.Unauthorized("authentication required")
	}

	requestID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return errors.BadRequest("invalid request ID")
	}

	var input RejectUploadRequestInput
	if err := c.BodyParser(&input); err != nil {
		// Body is optional, so ignore parse errors for empty body
		input = RejectUploadRequestInput{}
	}

	if err := h.podService.RejectUploadRequest(c.Context(), requestID, userID, input.Reason); err != nil {
		return err
	}

	reqID := middleware.GetRequestID(c)
	return c.JSON(response.OK(reqID, fiber.Map{"rejected": true}))
}

// RevokeUploadPermission handles DELETE /api/v1/upload-requests/:id
// @Summary Revoke upload permission
// @Description Revoke a previously approved upload permission
// @Tags Upload Requests
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Upload Request ID" format(uuid)
// @Success 200 {object} response.Response[map[string]bool] "Upload permission revoked"
// @Failure 400 {object} errors.ErrorResponse "Invalid request ID"
// @Failure 401 {object} errors.ErrorResponse "Authentication required"
// @Failure 403 {object} errors.ErrorResponse "Not the pod owner"
// @Failure 404 {object} errors.ErrorResponse "Upload request not found"
// @Router /upload-requests/{id} [delete]
func (h *PodHandler) RevokeUploadPermission(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok || userID == uuid.Nil {
		return errors.Unauthorized("authentication required")
	}

	requestID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return errors.BadRequest("invalid request ID")
	}

	if err := h.podService.RevokeUploadPermission(c.Context(), requestID, userID); err != nil {
		return err
	}

	reqID := middleware.GetRequestID(c)
	return c.JSON(response.OK(reqID, fiber.Map{"revoked": true}))
}

// SharePodInput represents the input for sharing a pod with a student.
// @Description Share pod with student input
type SharePodInput struct {
	StudentID uuid.UUID `json:"student_id" validate:"required" example:"550e8400-e29b-41d4-a716-446655440000"`
	Message   *string   `json:"message,omitempty" example:"I recommend this pod for your studies on Go programming."`
}

// SharePod handles POST /api/v1/pods/:id/share
// @Summary Share a pod with a student
// @Description Share a knowledge pod with a student. Only teachers can share pods.
// @Tags Shared Pods
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Pod ID" format(uuid)
// @Param request body SharePodInput true "Share pod input"
// @Success 201 {object} response.Response[domain.SharedPod] "Pod shared successfully"
// @Failure 400 {object} errors.ErrorResponse "Invalid request body"
// @Failure 401 {object} errors.ErrorResponse "Authentication required"
// @Failure 403 {object} errors.ErrorResponse "Only teachers can share pods"
// @Failure 404 {object} errors.ErrorResponse "Pod or student not found"
// @Failure 409 {object} errors.ErrorResponse "Pod already shared with this student"
// @Router /pods/{id}/share [post]
func (h *PodHandler) SharePod(c *fiber.Ctx) error {
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

	var input SharePodInput
	if err := c.BodyParser(&input); err != nil {
		return errors.BadRequest("invalid request body")
	}

	if err := h.validator.Struct(input); err != nil {
		return err
	}

	sharedPod, err := h.podService.SharePodWithStudent(c.Context(), userID, podID, input.StudentID, input.Message)
	if err != nil {
		return err
	}

	requestID := middleware.GetRequestID(c)
	return c.Status(fiber.StatusCreated).JSON(response.Created(requestID, sharedPod))
}

// GetSharedPods handles GET /api/v1/users/me/shared-pods
// @Summary Get pods shared with the current user
// @Description Get a paginated list of pods that have been shared with the current user by teachers
// @Tags Shared Pods
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param per_page query int false "Items per page" default(20)
// @Success 200 {object} response.PaginatedResponse[domain.SharedPodWithDetails] "List of shared pods"
// @Failure 401 {object} errors.ErrorResponse "Authentication required"
// @Router /users/me/shared-pods [get]
func (h *PodHandler) GetSharedPods(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok || userID == uuid.Nil {
		return errors.Unauthorized("authentication required")
	}

	page := c.QueryInt("page", 1)
	perPage := c.QueryInt("per_page", 20)

	result, err := h.podService.GetSharedPods(c.Context(), userID, page, perPage)
	if err != nil {
		return err
	}

	requestID := middleware.GetRequestID(c)
	return c.JSON(response.List(requestID, result.SharedPods, result.Page, result.PerPage, result.Total))
}

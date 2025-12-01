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
)

// PodHandler handles HTTP requests for pod operations.
type PodHandler struct {
	podService application.PodService
	validator  *validator.Validator
}

// NewPodHandler creates a new PodHandler.
func NewPodHandler(podService application.PodService) *PodHandler {
	return &PodHandler{
		podService: podService,
		validator:  validator.Get(),
	}
}

// CreatePod handles POST /api/v1/pods
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
// Permission check is done by middleware (RequireReadAccess).
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

	requestID := middleware.GetRequestID(c)
	return c.JSON(response.OK(requestID, pod))
}

// ListPods handles GET /api/v1/pods
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
// Permission check is done by middleware (RequireEditAccess).
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
// Permission check is done by middleware (RequireOwnerAccess).
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
// Permission check is done by middleware (RequireReadAccess).
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

	requestID := middleware.GetRequestID(c)
	return c.Status(fiber.StatusCreated).JSON(response.Created(requestID, pod))
}

// StarPod handles POST /api/v1/pods/:id/star
// Permission check is done by middleware (RequireReadAccess).
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

	requestID := middleware.GetRequestID(c)
	return c.JSON(response.OK(requestID, fiber.Map{"starred": true}))
}

// UnstarPod handles DELETE /api/v1/pods/:id/star
// Permission check is done by middleware (RequireReadAccess).
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

	requestID := middleware.GetRequestID(c)
	return c.JSON(response.OK(requestID, fiber.Map{"starred": false}))
}

// FollowPod handles POST /api/v1/pods/:id/follow
// Permission check is done by middleware (RequireReadAccess).
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

	requestID := middleware.GetRequestID(c)
	return c.JSON(response.OK(requestID, fiber.Map{"following": true}))
}

// UnfollowPod handles DELETE /api/v1/pods/:id/follow
// Permission check is done by middleware (RequireReadAccess).
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

	requestID := middleware.GetRequestID(c)
	return c.JSON(response.OK(requestID, fiber.Map{"following": false}))
}

// InviteCollaborator handles POST /api/v1/pods/:id/collaborators
// Permission check is done by middleware (RequireCollaboratorManagement).
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
// Permission check is done by middleware (RequireCollaboratorManagement).
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
// Permission check is done by middleware (RequireCollaboratorManagement).
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
// Permission check is done by middleware (RequireReadAccess).
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
// Permission check is done by middleware (RequireReadAccess).
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

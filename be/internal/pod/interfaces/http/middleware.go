// Package http provides HTTP handlers and middleware for the Pod Service API.
package http

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"ngasihtau/internal/common/errors"
	"ngasihtau/internal/common/middleware"
	"ngasihtau/internal/pod/application"
)

// PodPermissionMiddleware provides middleware functions for pod permission checks.
// Implements requirements 3 and 4 for access control.
type PodPermissionMiddleware struct {
	podService application.PodService
}

// NewPodPermissionMiddleware creates a new PodPermissionMiddleware.
func NewPodPermissionMiddleware(podService application.PodService) *PodPermissionMiddleware {
	return &PodPermissionMiddleware{
		podService: podService,
	}
}

// PodIDKey is the context key for the parsed pod ID.
const PodIDKey = "pod_id"

// GetPodID extracts the pod ID from the Fiber context.
func GetPodID(c *fiber.Ctx) (uuid.UUID, bool) {
	if id, ok := c.Locals(PodIDKey).(uuid.UUID); ok {
		return id, true
	}
	return uuid.Nil, false
}

// ExtractPodID middleware parses and validates the pod ID from URL params.
// It stores the parsed UUID in context for downstream handlers.
func (m *PodPermissionMiddleware) ExtractPodID() fiber.Handler {
	return func(c *fiber.Ctx) error {
		idParam := c.Params("id")
		if idParam == "" {
			return errors.BadRequest("pod ID is required")
		}

		podID, err := uuid.Parse(idParam)
		if err != nil {
			return errors.BadRequest("invalid pod ID format")
		}

		c.Locals(PodIDKey, podID)
		return c.Next()
	}
}

// RequireReadAccess middleware checks if the user can read the pod.
// Public pods are accessible to everyone.
// Private pods require the user to be owner or collaborator.
// Implements requirement 3 for pod visibility access control.
func (m *PodPermissionMiddleware) RequireReadAccess() fiber.Handler {
	return func(c *fiber.Ctx) error {
		podID, ok := GetPodID(c)
		if !ok {
			return errors.Internal("pod ID not found in context", nil)
		}

		var userID *uuid.UUID
		if uid, ok := middleware.GetUserID(c); ok && uid != uuid.Nil {
			userID = &uid
		}

		canAccess, err := m.podService.CanUserAccessPod(c.Context(), podID, userID)
		if err != nil {
			return err
		}

		if !canAccess {
			return errors.Forbidden("you do not have access to this pod")
		}

		return c.Next()
	}
}

// RequireEditAccess middleware checks if the user can edit the pod.
// Only owner or admin collaborators can edit.
// Implements requirements 3 and 4 for update/delete operations.
func (m *PodPermissionMiddleware) RequireEditAccess() fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID, ok := middleware.GetUserID(c)
		if !ok || userID == uuid.Nil {
			return errors.Unauthorized("authentication required")
		}

		podID, ok := GetPodID(c)
		if !ok {
			return errors.Internal("pod ID not found in context", nil)
		}

		canEdit, err := m.podService.CanUserEditPod(c.Context(), podID, userID)
		if err != nil {
			return err
		}

		if !canEdit {
			return errors.Forbidden("you do not have permission to edit this pod")
		}

		return c.Next()
	}
}

// RequireOwnerAccess middleware checks if the user is the pod owner.
// Only the owner can perform certain operations like delete.
// Implements requirement 3 for owner-only operations.
func (m *PodPermissionMiddleware) RequireOwnerAccess() fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID, ok := middleware.GetUserID(c)
		if !ok || userID == uuid.Nil {
			return errors.Unauthorized("authentication required")
		}

		podID, ok := GetPodID(c)
		if !ok {
			return errors.Internal("pod ID not found in context", nil)
		}

		// Get pod to check ownership
		pod, err := m.podService.GetPod(c.Context(), podID, &userID)
		if err != nil {
			return err
		}

		if !pod.IsOwner(userID) {
			return errors.Forbidden("only the owner can perform this action")
		}

		return c.Next()
	}
}

// RequireUploadAccess middleware checks if the user can upload materials to the pod.
// Owner or verified contributors/admins can upload.
// Implements requirement 4 for verified collaborator upload permissions.
func (m *PodPermissionMiddleware) RequireUploadAccess() fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID, ok := middleware.GetUserID(c)
		if !ok || userID == uuid.Nil {
			return errors.Unauthorized("authentication required")
		}

		podID, ok := GetPodID(c)
		if !ok {
			return errors.Internal("pod ID not found in context", nil)
		}

		canUpload, err := m.podService.CanUserUploadToPod(c.Context(), podID, userID)
		if err != nil {
			return err
		}

		if !canUpload {
			return errors.Forbidden("you do not have permission to upload to this pod")
		}

		return c.Next()
	}
}

// RequireCollaboratorManagement middleware checks if the user can manage collaborators.
// Only owner or admin collaborators can manage other collaborators.
// Implements requirement 4 for collaborator management.
func (m *PodPermissionMiddleware) RequireCollaboratorManagement() fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID, ok := middleware.GetUserID(c)
		if !ok || userID == uuid.Nil {
			return errors.Unauthorized("authentication required")
		}

		podID, ok := GetPodID(c)
		if !ok {
			return errors.Internal("pod ID not found in context", nil)
		}

		canEdit, err := m.podService.CanUserEditPod(c.Context(), podID, userID)
		if err != nil {
			return err
		}

		if !canEdit {
			return errors.Forbidden("you do not have permission to manage collaborators")
		}

		return c.Next()
	}
}

// Package http provides HTTP handlers for the Pod Service API.
// Implements the interfaces layer in Clean Architecture.
package http

import (
	"github.com/gofiber/fiber/v2"

	"ngasihtau/internal/common/middleware"
	"ngasihtau/internal/pod/application"
	"ngasihtau/pkg/jwt"
)

// Handler contains all HTTP handlers for the Pod Service.
type Handler struct {
	podHandler    *PodHandler
	jwtManager    *jwt.Manager
	podPermission *PodPermissionMiddleware
}

// NewHandler creates a new Handler with the given dependencies.
func NewHandler(podService application.PodService, jwtManager *jwt.Manager) *Handler {
	return &Handler{
		podHandler:    NewPodHandler(podService),
		jwtManager:    jwtManager,
		podPermission: NewPodPermissionMiddleware(podService),
	}
}

// RegisterRoutes registers all Pod Service routes on the given Fiber app.
// Routes are organized by functionality with permission middleware:
//
// Pod CRUD (mixed auth):
//   - POST   /api/v1/pods (protected - create pod)
//   - GET    /api/v1/pods (public - list pods with filters)
//   - GET    /api/v1/pods/:id (public with optional auth - get pod, checks visibility)
//   - PUT    /api/v1/pods/:id (protected - update pod, requires edit access)
//   - DELETE /api/v1/pods/:id (protected - delete pod, requires owner access)
//
// Fork (protected):
//   - POST   /api/v1/pods/:id/fork (requires read access)
//
// Star (protected):
//   - POST   /api/v1/pods/:id/star (requires read access)
//   - DELETE /api/v1/pods/:id/star (requires read access)
//
// Collaborators (protected):
//   - GET    /api/v1/pods/:id/collaborators (requires read access)
//   - POST   /api/v1/pods/:id/collaborators (requires collaborator management)
//   - PUT    /api/v1/pods/:id/collaborators/:userId (requires collaborator management)
//   - DELETE /api/v1/pods/:id/collaborators/:userId (requires collaborator management)
//
// Follow (protected):
//   - POST   /api/v1/pods/:id/follow (requires read access)
//   - DELETE /api/v1/pods/:id/follow (requires read access)
//
// Activity (public with optional auth):
//   - GET    /api/v1/pods/:id/activity (requires read access)
//   - GET    /api/v1/feed (protected - user's activity feed)
//
// User's pods and starred (public):
//   - GET    /api/v1/users/:id/pods
//   - GET    /api/v1/users/:id/starred
func (h *Handler) RegisterRoutes(app *fiber.App) {
	// API v1 group
	api := app.Group("/api/v1")

	// Pod routes
	pods := api.Group("/pods")

	// Public routes - list pods
	pods.Get("", middleware.OptionalAuth(h.jwtManager), h.podHandler.ListPods)

	// Routes with pod ID that need permission checks
	// GET pod - public with optional auth, checks visibility
	pods.Get("/:id",
		middleware.OptionalAuth(h.jwtManager),
		h.podPermission.ExtractPodID(),
		h.podPermission.RequireReadAccess(),
		h.podHandler.GetPod,
	)

	// GET activity - public with optional auth, checks visibility
	pods.Get("/:id/activity",
		middleware.OptionalAuth(h.jwtManager),
		h.podPermission.ExtractPodID(),
		h.podPermission.RequireReadAccess(),
		h.podHandler.GetPodActivity,
	)

	// GET collaborators - public with optional auth, checks visibility
	pods.Get("/:id/collaborators",
		middleware.OptionalAuth(h.jwtManager),
		h.podPermission.ExtractPodID(),
		h.podPermission.RequireReadAccess(),
		h.podHandler.GetCollaborators,
	)

	// Protected routes - require authentication
	// Create pod (no pod ID yet)
	pods.Post("", middleware.Auth(h.jwtManager), h.podHandler.CreatePod)

	// Update pod - requires edit access (owner or admin collaborator)
	pods.Put("/:id",
		middleware.Auth(h.jwtManager),
		h.podPermission.ExtractPodID(),
		h.podPermission.RequireEditAccess(),
		h.podHandler.UpdatePod,
	)

	// Delete pod - requires owner access
	pods.Delete("/:id",
		middleware.Auth(h.jwtManager),
		h.podPermission.ExtractPodID(),
		h.podPermission.RequireOwnerAccess(),
		h.podHandler.DeletePod,
	)

	// Fork - requires read access
	pods.Post("/:id/fork",
		middleware.Auth(h.jwtManager),
		h.podPermission.ExtractPodID(),
		h.podPermission.RequireReadAccess(),
		h.podHandler.ForkPod,
	)

	// Star - requires read access
	pods.Post("/:id/star",
		middleware.Auth(h.jwtManager),
		h.podPermission.ExtractPodID(),
		h.podPermission.RequireReadAccess(),
		h.podHandler.StarPod,
	)
	pods.Delete("/:id/star",
		middleware.Auth(h.jwtManager),
		h.podPermission.ExtractPodID(),
		h.podPermission.RequireReadAccess(),
		h.podHandler.UnstarPod,
	)

	// Follow - requires read access
	pods.Post("/:id/follow",
		middleware.Auth(h.jwtManager),
		h.podPermission.ExtractPodID(),
		h.podPermission.RequireReadAccess(),
		h.podHandler.FollowPod,
	)
	pods.Delete("/:id/follow",
		middleware.Auth(h.jwtManager),
		h.podPermission.ExtractPodID(),
		h.podPermission.RequireReadAccess(),
		h.podHandler.UnfollowPod,
	)

	// Collaborators management - requires collaborator management permission
	pods.Post("/:id/collaborators",
		middleware.Auth(h.jwtManager),
		h.podPermission.ExtractPodID(),
		h.podPermission.RequireCollaboratorManagement(),
		h.podHandler.InviteCollaborator,
	)
	pods.Put("/:id/collaborators/:userId",
		middleware.Auth(h.jwtManager),
		h.podPermission.ExtractPodID(),
		h.podPermission.RequireCollaboratorManagement(),
		h.podHandler.UpdateCollaborator,
	)
	pods.Delete("/:id/collaborators/:userId",
		middleware.Auth(h.jwtManager),
		h.podPermission.ExtractPodID(),
		h.podPermission.RequireCollaboratorManagement(),
		h.podHandler.RemoveCollaborator,
	)

	// Activity feed (protected)
	api.Get("/feed", middleware.Auth(h.jwtManager), h.podHandler.GetUserFeed)

	// User-related pod routes (public)
	users := api.Group("/users")
	users.Get("/:id/pods", h.podHandler.GetUserPods)
	users.Get("/:id/starred", h.podHandler.GetUserStarredPods)
}

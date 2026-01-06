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
	podHandler            *PodHandler
	recommendationHandler *RecommendationHandler
	jwtManager            *jwt.Manager
	podPermission         *PodPermissionMiddleware
}

// NewHandler creates a new Handler with the given dependencies.
func NewHandler(
	podService application.PodService,
	recommendationService application.RecommendationService,
	jwtManager *jwt.Manager,
) *Handler {
	return &Handler{
		podHandler:            NewPodHandler(podService, recommendationService),
		recommendationHandler: NewRecommendationHandler(recommendationService),
		jwtManager:            jwtManager,
		podPermission:         NewPodPermissionMiddleware(podService),
	}
}

// RegisterRoutes registers all Pod Service routes on the given Fiber app.
// Routes are organized by functionality with permission middleware:
//
// Pod CRUD (mixed auth):
//   - POST   /api/v1/pods (protected - create pod)
//   - GET    /api/v1/pods (public - list pods with filters)
//   - GET    /api/v1/pods/slug/:slug (public with optional auth - get pod by slug)
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
// Upvote (protected - trust indicator):
//   - POST   /api/v1/pods/:id/upvote (requires read access)
//   - DELETE /api/v1/pods/:id/upvote (requires read access)
//
// Upload Request (protected - teacher collaboration):
//   - POST   /api/v1/pods/:id/upload-request (requires read access)
//   - GET    /api/v1/users/me/upload-requests (protected)
//   - POST   /api/v1/upload-requests/:id/approve (protected)
//   - POST   /api/v1/upload-requests/:id/reject (protected)
//   - DELETE /api/v1/upload-requests/:id (protected)
//
// Shared Pods (protected - teacher-student sharing):
//   - POST   /api/v1/pods/:id/share (requires read access)
//   - GET    /api/v1/users/me/shared-pods (protected)
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
// Recommendations (mixed auth):
//   - POST   /api/v1/pods/:id/track (protected - track interaction)
//   - POST   /api/v1/pods/:id/track/time (protected - track time spent)
//   - GET    /api/v1/pods/:id/similar (public - similar pods)
//   - GET    /api/v1/feed/recommended (protected - personalized feed)
//   - GET    /api/v1/feed/trending (public - trending feed)
//   - GET    /api/v1/users/me/preferences (protected - user preferences)
//
// User's pods and starred (public):
//   - GET    /api/v1/users/:id/pods
//   - GET    /api/v1/users/:id/starred
//
// User's upvoted pods (protected):
//   - GET    /api/v1/users/me/upvoted-pods
func (h *Handler) RegisterRoutes(app *fiber.App) {
	// API v1 group
	api := app.Group("/api/v1")

	// Pod routes
	pods := api.Group("/pods")

	// Public routes - list pods
	pods.Get("", middleware.OptionalAuth(h.jwtManager), h.podHandler.ListPods)

	// GET pod by slug - public with optional auth, checks visibility
	pods.Get("/slug/:slug",
		middleware.OptionalAuth(h.jwtManager),
		h.podHandler.GetPodBySlug,
	)

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

	// Upvote - requires read access (trust indicator)
	pods.Post("/:id/upvote",
		middleware.Auth(h.jwtManager),
		h.podPermission.ExtractPodID(),
		h.podPermission.RequireReadAccess(),
		h.podHandler.UpvotePod,
	)
	pods.Delete("/:id/upvote",
		middleware.Auth(h.jwtManager),
		h.podPermission.ExtractPodID(),
		h.podPermission.RequireReadAccess(),
		h.podHandler.RemoveUpvote,
	)

	// Upload Request - requires read access (teacher collaboration)
	pods.Post("/:id/upload-request",
		middleware.Auth(h.jwtManager),
		h.podPermission.ExtractPodID(),
		h.podPermission.RequireReadAccess(),
		h.podHandler.CreateUploadRequest,
	)

	// Share Pod - requires read access (teacher-student sharing)
	pods.Post("/:id/share",
		middleware.Auth(h.jwtManager),
		h.podPermission.ExtractPodID(),
		h.podPermission.RequireReadAccess(),
		h.podHandler.SharePod,
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

	// === Upload Request management routes ===
	uploadRequests := api.Group("/upload-requests")
	uploadRequests.Post("/:id/approve",
		middleware.Auth(h.jwtManager),
		h.podHandler.ApproveUploadRequest,
	)
	uploadRequests.Post("/:id/reject",
		middleware.Auth(h.jwtManager),
		h.podHandler.RejectUploadRequest,
	)
	uploadRequests.Delete("/:id",
		middleware.Auth(h.jwtManager),
		h.podHandler.RevokeUploadPermission,
	)

	// === Recommendation routes ===
	// Track interaction (protected)
	pods.Post("/:id/track",
		middleware.Auth(h.jwtManager),
		h.recommendationHandler.TrackInteraction,
	)

	// Track time spent (protected)
	pods.Post("/:id/track/time",
		middleware.Auth(h.jwtManager),
		h.recommendationHandler.TrackTimeSpent,
	)

	// Get similar pods (public)
	pods.Get("/:id/similar",
		h.recommendationHandler.GetSimilarPods,
	)

	// === Feed routes ===
	// Activity feed (protected)
	api.Get("/feed", middleware.Auth(h.jwtManager), h.podHandler.GetUserFeed)

	// Personalized feed (protected)
	api.Get("/feed/recommended",
		middleware.Auth(h.jwtManager),
		h.recommendationHandler.GetPersonalizedFeed,
	)

	// Trending feed (public)
	api.Get("/feed/trending", h.recommendationHandler.GetTrendingFeed)

	// User-related pod routes (public)
	users := api.Group("/users")
	users.Get("/:id/pods", h.podHandler.GetUserPods)
	users.Get("/:id/starred", h.podHandler.GetUserStarredPods)

	// User's upvoted pods (protected)
	users.Get("/me/upvoted-pods",
		middleware.Auth(h.jwtManager),
		h.podHandler.GetUserUpvotedPods,
	)

	// User's upload requests (protected)
	users.Get("/me/upload-requests",
		middleware.Auth(h.jwtManager),
		h.podHandler.GetUploadRequests,
	)

	// User's shared pods (protected)
	users.Get("/me/shared-pods",
		middleware.Auth(h.jwtManager),
		h.podHandler.GetSharedPods,
	)

	// User preferences (protected)
	users.Get("/me/preferences",
		middleware.Auth(h.jwtManager),
		h.recommendationHandler.GetUserPreferences,
	)
}

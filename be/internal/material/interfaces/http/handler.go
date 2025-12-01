// Package http provides HTTP handlers for the Material Service.
package http

import (
	"time"

	"github.com/gofiber/fiber/v2"

	"ngasihtau/internal/common/middleware"
	"ngasihtau/internal/material/application"
	"ngasihtau/pkg/jwt"
)

// Handler handles HTTP requests for the Material Service.
type Handler struct {
	service    *application.Service
	jwtManager *jwt.Manager
}

// NewHandler creates a new Handler.
func NewHandler(service *application.Service, jwtManager *jwt.Manager) *Handler {
	return &Handler{
		service:    service,
		jwtManager: jwtManager,
	}
}

// RegisterRoutes registers all routes for the Material Service.
func (h *Handler) RegisterRoutes(app *fiber.App) {
	api := app.Group("/api/v1")

	// Material upload routes (require authentication)
	materials := api.Group("/materials", middleware.Auth(h.jwtManager))
	materials.Post("/upload-url", h.GetUploadURL)
	materials.Post("/confirm", h.ConfirmUpload)
	materials.Get("/:id", h.GetMaterial)
	materials.Put("/:id", h.UpdateMaterial)
	materials.Delete("/:id", h.DeleteMaterial)
	materials.Get("/:id/preview", h.GetPreviewURL)
	materials.Get("/:id/download", h.GetDownloadURL)
	materials.Post("/:id/versions", h.CreateVersion)
	materials.Get("/:id/versions", h.GetVersionHistory)
	materials.Post("/:id/comments", h.AddComment)
	materials.Get("/:id/comments", h.GetComments)
	materials.Post("/:id/ratings", h.RateMaterial)
	materials.Get("/:id/ratings", h.GetRatings)
	materials.Post("/:id/bookmark", h.BookmarkMaterial)
	materials.Delete("/:id/bookmark", h.RemoveBookmark)

	// Comment routes
	comments := api.Group("/comments", middleware.Auth(h.jwtManager))
	comments.Put("/:id", h.UpdateComment)
	comments.Delete("/:id", h.DeleteComment)

	// Bookmark routes
	bookmarks := api.Group("/bookmarks", middleware.Auth(h.jwtManager))
	bookmarks.Get("/", h.GetBookmarks)
	bookmarks.Get("/folders", h.GetBookmarkFolders)

	// Pod materials route (list materials in a pod)
	pods := api.Group("/pods", middleware.Auth(h.jwtManager))
	pods.Get("/:podId/materials", h.GetMaterialsByPod)
}

// successResponse creates a standard success response.
func successResponse(c *fiber.Ctx, data interface{}) error {
	requestID := middleware.GetRequestID(c)
	return c.JSON(fiber.Map{
		"success": true,
		"data":    data,
		"meta": fiber.Map{
			"timestamp":  time.Now().UTC().Format(time.RFC3339),
			"request_id": requestID,
		},
	})
}

// paginatedResponse creates a standard paginated response.
func paginatedResponse(c *fiber.Ctx, data interface{}, page, perPage, total int) error {
	requestID := middleware.GetRequestID(c)
	totalPages := (total + perPage - 1) / perPage

	return c.JSON(fiber.Map{
		"success": true,
		"data":    data,
		"pagination": fiber.Map{
			"page":        page,
			"per_page":    perPage,
			"total":       total,
			"total_pages": totalPages,
		},
		"meta": fiber.Map{
			"timestamp":  time.Now().UTC().Format(time.RFC3339),
			"request_id": requestID,
		},
	})
}

// errorResponse creates a standard error response.
func errorResponse(c *fiber.Ctx, code int, errorCode, message string) error {
	requestID := middleware.GetRequestID(c)
	return c.Status(code).JSON(fiber.Map{
		"success": false,
		"error": fiber.Map{
			"code":    errorCode,
			"message": message,
		},
		"meta": fiber.Map{
			"timestamp":  time.Now().UTC().Format(time.RFC3339),
			"request_id": requestID,
		},
	})
}

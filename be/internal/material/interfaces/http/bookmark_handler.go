package http

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"ngasihtau/internal/common/middleware"
	"ngasihtau/internal/material/application"
)

// BookmarkMaterialRequest represents the request body for bookmarking a material.
type BookmarkMaterialRequest struct {
	Folder *string `json:"folder" validate:"omitempty,max=100"`
}

// BookmarkMaterial bookmarks a material.
// POST /api/v1/materials/:id/bookmark
func (h *Handler) BookmarkMaterial(c *fiber.Ctx) error {
	materialID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "VALIDATION_ERROR", "Invalid material ID")
	}

	var req BookmarkMaterialRequest
	if err := c.BodyParser(&req); err != nil {
		// Body is optional for bookmarks
		req = BookmarkMaterialRequest{}
	}

	userID, ok := middleware.GetUserID(c)
	if !ok {
		return errorResponse(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated")
	}

	bookmark, err := h.service.BookmarkMaterial(c.Context(), application.BookmarkMaterialInput{
		UserID:     userID,
		MaterialID: materialID,
		Folder:     req.Folder,
	})
	if err != nil {
		return errorResponse(c, fiber.StatusConflict, "CONFLICT", err.Error())
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"data":    bookmark,
	})
}

// RemoveBookmark removes a bookmark.
// DELETE /api/v1/materials/:id/bookmark
func (h *Handler) RemoveBookmark(c *fiber.Ctx) error {
	materialID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "VALIDATION_ERROR", "Invalid material ID")
	}

	userID, ok := middleware.GetUserID(c)
	if !ok {
		return errorResponse(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated")
	}

	if err := h.service.RemoveBookmark(c.Context(), userID, materialID); err != nil {
		return errorResponse(c, fiber.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
	}

	return c.Status(fiber.StatusNoContent).Send(nil)
}

// GetBookmarks retrieves bookmarks for the current user.
// GET /api/v1/bookmarks
func (h *Handler) GetBookmarks(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return errorResponse(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated")
	}

	page := c.QueryInt("page", 1)
	perPage := c.QueryInt("per_page", 20)
	if perPage > 100 {
		perPage = 100
	}
	offset := (page - 1) * perPage

	// Optional folder filter
	var folder *string
	if f := c.Query("folder"); f != "" {
		folder = &f
	}

	materials, total, err := h.service.GetBookmarks(c.Context(), userID, folder, perPage, offset)
	if err != nil {
		return errorResponse(c, fiber.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
	}

	return paginatedResponse(c, materials, page, perPage, total)
}

// GetBookmarkFolders retrieves all bookmark folders for the current user.
// GET /api/v1/bookmarks/folders
func (h *Handler) GetBookmarkFolders(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return errorResponse(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated")
	}

	folders, err := h.service.GetBookmarkFolders(c.Context(), userID)
	if err != nil {
		return errorResponse(c, fiber.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
	}

	return successResponse(c, fiber.Map{"folders": folders})
}

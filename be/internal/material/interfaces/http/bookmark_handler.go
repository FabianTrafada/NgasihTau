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
// @Summary Bookmark a material
// @Description Add a material to bookmarks with optional folder
// @Tags Bookmarks
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Material ID" format(uuid)
// @Param request body BookmarkMaterialRequest false "Optional folder name"
// @Success 201 {object} fiber.Map "Created bookmark"
// @Failure 400 {object} fiber.Map "Invalid material ID"
// @Failure 401 {object} fiber.Map "Authentication required"
// @Failure 409 {object} fiber.Map "Already bookmarked"
// @Router /materials/{id}/bookmark [post]
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
// @Summary Remove a bookmark
// @Description Remove a material from bookmarks
// @Tags Bookmarks
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Material ID" format(uuid)
// @Success 204 "Bookmark removed"
// @Failure 400 {object} fiber.Map "Invalid material ID"
// @Failure 401 {object} fiber.Map "Authentication required"
// @Router /materials/{id}/bookmark [delete]
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
// @Summary Get bookmarks
// @Description Get a paginated list of bookmarked materials
// @Tags Bookmarks
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param folder query string false "Filter by folder name"
// @Param page query int false "Page number" default(1)
// @Param per_page query int false "Items per page" default(20) maximum(100)
// @Success 200 {object} fiber.Map "List of bookmarked materials"
// @Failure 401 {object} fiber.Map "Authentication required"
// @Router /bookmarks [get]
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
// @Summary Get bookmark folders
// @Description Get a list of all bookmark folders
// @Tags Bookmarks
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} fiber.Map "List of folders"
// @Failure 401 {object} fiber.Map "Authentication required"
// @Router /bookmarks/folders [get]
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

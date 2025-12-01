package http

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"ngasihtau/internal/common/middleware"
	"ngasihtau/internal/material/application"
)

// AddCommentRequest represents the request body for adding a comment.
type AddCommentRequest struct {
	Content  string  `json:"content" validate:"required,max=2000"`
	ParentID *string `json:"parent_id" validate:"omitempty,uuid"`
}

// AddComment adds a comment to a material.
// POST /api/v1/materials/:id/comments
func (h *Handler) AddComment(c *fiber.Ctx) error {
	materialID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "VALIDATION_ERROR", "Invalid material ID")
	}

	var req AddCommentRequest
	if err := c.BodyParser(&req); err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "VALIDATION_ERROR", "Invalid request body")
	}

	if req.Content == "" {
		return errorResponse(c, fiber.StatusBadRequest, "VALIDATION_ERROR", "Content is required")
	}

	var parentID *uuid.UUID
	if req.ParentID != nil {
		pid, err := uuid.Parse(*req.ParentID)
		if err != nil {
			return errorResponse(c, fiber.StatusBadRequest, "VALIDATION_ERROR", "Invalid parent ID")
		}
		parentID = &pid
	}

	userID, ok := middleware.GetUserID(c)
	if !ok {
		return errorResponse(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated")
	}

	comment, err := h.service.AddComment(c.Context(), application.AddCommentInput{
		MaterialID: materialID,
		UserID:     userID,
		Content:    req.Content,
		ParentID:   parentID,
	})
	if err != nil {
		return errorResponse(c, fiber.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"data":    comment,
	})
}

// GetComments retrieves comments for a material.
// GET /api/v1/materials/:id/comments
func (h *Handler) GetComments(c *fiber.Ctx) error {
	materialID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "VALIDATION_ERROR", "Invalid material ID")
	}

	page := c.QueryInt("page", 1)
	perPage := c.QueryInt("per_page", 20)
	if perPage > 100 {
		perPage = 100
	}
	offset := (page - 1) * perPage

	comments, total, err := h.service.GetComments(c.Context(), materialID, perPage, offset)
	if err != nil {
		return errorResponse(c, fiber.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
	}

	return paginatedResponse(c, comments, page, perPage, total)
}

// UpdateCommentRequest represents the request body for updating a comment.
type UpdateCommentRequest struct {
	Content string `json:"content" validate:"required,max=2000"`
}

// UpdateComment updates a comment.
// PUT /api/v1/comments/:id
func (h *Handler) UpdateComment(c *fiber.Ctx) error {
	commentID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "VALIDATION_ERROR", "Invalid comment ID")
	}

	var req UpdateCommentRequest
	if err := c.BodyParser(&req); err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "VALIDATION_ERROR", "Invalid request body")
	}

	if req.Content == "" {
		return errorResponse(c, fiber.StatusBadRequest, "VALIDATION_ERROR", "Content is required")
	}

	userID, ok := middleware.GetUserID(c)
	if !ok {
		return errorResponse(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated")
	}

	comment, err := h.service.UpdateComment(c.Context(), application.UpdateCommentInput{
		ID:      commentID,
		UserID:  userID,
		Content: req.Content,
	})
	if err != nil {
		return errorResponse(c, fiber.StatusForbidden, "FORBIDDEN", err.Error())
	}

	return successResponse(c, comment)
}

// DeleteComment soft-deletes a comment.
// DELETE /api/v1/comments/:id
func (h *Handler) DeleteComment(c *fiber.Ctx) error {
	commentID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "VALIDATION_ERROR", "Invalid comment ID")
	}

	userID, ok := middleware.GetUserID(c)
	if !ok {
		return errorResponse(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated")
	}

	if err := h.service.DeleteComment(c.Context(), commentID, userID); err != nil {
		return errorResponse(c, fiber.StatusForbidden, "FORBIDDEN", err.Error())
	}

	return c.Status(fiber.StatusNoContent).Send(nil)
}

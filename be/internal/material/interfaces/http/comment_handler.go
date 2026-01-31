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
// @Summary Add a comment
// @Description Add a comment to a material. Supports threaded replies.
// @Tags Comments
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Material ID" format(uuid)
// @Param request body AddCommentRequest true "Comment content"
// @Success 201 {object} fiber.Map "Created comment"
// @Failure 400 {object} fiber.Map "Invalid request"
// @Failure 401 {object} fiber.Map "Authentication required"
// @Router /materials/{id}/comments [post]
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
// @Summary Get comments
// @Description Get a paginated list of comments for a material
// @Tags Comments
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Material ID" format(uuid)
// @Param page query int false "Page number" default(1)
// @Param per_page query int false "Items per page" default(20) maximum(100)
// @Success 200 {object} fiber.Map "List of comments"
// @Failure 400 {object} fiber.Map "Invalid material ID"
// @Router /materials/{id}/comments [get]
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
// @Summary Update a comment
// @Description Update a comment. Only the comment author can update.
// @Tags Comments
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Comment ID" format(uuid)
// @Param request body UpdateCommentRequest true "Updated content"
// @Success 200 {object} fiber.Map "Updated comment"
// @Failure 400 {object} fiber.Map "Invalid request"
// @Failure 401 {object} fiber.Map "Authentication required"
// @Failure 403 {object} fiber.Map "Not comment author"
// @Router /comments/{id} [put]
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
// @Summary Delete a comment
// @Description Soft-delete a comment. Only the comment author can delete.
// @Tags Comments
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Comment ID" format(uuid)
// @Success 204 "Comment deleted"
// @Failure 400 {object} fiber.Map "Invalid comment ID"
// @Failure 401 {object} fiber.Map "Authentication required"
// @Failure 403 {object} fiber.Map "Not comment author"
// @Router /comments/{id} [delete]
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

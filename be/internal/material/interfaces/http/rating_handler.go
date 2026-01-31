package http

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"ngasihtau/internal/common/middleware"
	"ngasihtau/internal/material/application"
)

// RateMaterialRequest represents the request body for rating a material.
type RateMaterialRequest struct {
	Score  int     `json:"score" validate:"required,min=1,max=5"`
	Review *string `json:"review" validate:"omitempty,max=2000"`
}

// RateMaterial rates a material.
// @Summary Rate a material
// @Description Rate a material (1-5 stars) with optional review. Updates existing rating if already rated.
// @Tags Ratings
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Material ID" format(uuid)
// @Param request body RateMaterialRequest true "Rating and optional review"
// @Success 201 {object} fiber.Map "Created/updated rating"
// @Failure 400 {object} fiber.Map "Invalid request"
// @Failure 401 {object} fiber.Map "Authentication required"
// @Router /materials/{id}/ratings [post]
func (h *Handler) RateMaterial(c *fiber.Ctx) error {
	materialID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "VALIDATION_ERROR", "Invalid material ID")
	}

	var req RateMaterialRequest
	if err := c.BodyParser(&req); err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "VALIDATION_ERROR", "Invalid request body")
	}

	if req.Score < 1 || req.Score > 5 {
		return errorResponse(c, fiber.StatusBadRequest, "VALIDATION_ERROR", "Score must be between 1 and 5")
	}

	userID, ok := middleware.GetUserID(c)
	if !ok {
		return errorResponse(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated")
	}

	rating, err := h.service.RateMaterial(c.Context(), application.RateMaterialInput{
		MaterialID: materialID,
		UserID:     userID,
		Score:      req.Score,
		Review:     req.Review,
	})
	if err != nil {
		return errorResponse(c, fiber.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"data":    rating,
	})
}

// GetRatings retrieves ratings for a material.
// @Summary Get ratings
// @Description Get ratings for a material. Use summary=true for rating summary.
// @Tags Ratings
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Material ID" format(uuid)
// @Param summary query bool false "Return summary instead of list"
// @Param page query int false "Page number" default(1)
// @Param per_page query int false "Items per page" default(10) maximum(50)
// @Success 200 {object} fiber.Map "Ratings or summary"
// @Failure 400 {object} fiber.Map "Invalid material ID"
// @Router /materials/{id}/ratings [get]
func (h *Handler) GetRatings(c *fiber.Ctx) error {
	materialID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "VALIDATION_ERROR", "Invalid material ID")
	}

	// Check if summary is requested
	if c.Query("summary") == "true" {
		summary, err := h.service.GetRatingSummary(c.Context(), materialID)
		if err != nil {
			return errorResponse(c, fiber.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		}
		return successResponse(c, summary)
	}

	page := c.QueryInt("page", 1)
	perPage := c.QueryInt("per_page", 10)
	if perPage > 50 {
		perPage = 50
	}
	offset := (page - 1) * perPage

	ratings, total, err := h.service.GetRatings(c.Context(), materialID, perPage, offset)
	if err != nil {
		return errorResponse(c, fiber.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
	}

	return paginatedResponse(c, ratings, page, perPage, total)
}

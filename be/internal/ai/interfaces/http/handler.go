package http

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"ngasihtau/internal/ai/application"
	"ngasihtau/internal/ai/domain"
	"ngasihtau/internal/common/middleware"
	"ngasihtau/pkg/jwt"
)

type Handler struct {
	service    *application.Service
	jwtManager *jwt.Manager
}

func NewHandler(service *application.Service, jwtManager *jwt.Manager) *Handler {
	return &Handler{
		service:    service,
		jwtManager: jwtManager,
	}
}

func (h *Handler) RegisterRoutes(app *fiber.App) {
	api := app.Group("/api/v1")

	materials := api.Group("/materials/:id")
	materials.Post("/chat", h.authMiddleware, h.Chat)
	materials.Get("/chat/history", h.authMiddleware, h.GetChatHistory)
	materials.Get("/chat/suggestions", h.authMiddleware, h.GetSuggestions)
	materials.Post("/chat/export", h.authMiddleware, h.ExportChat)

	pods := api.Group("/pods/:id")
	pods.Post("/chat", h.authMiddleware, h.PodChat)

	api.Post("/chat/:messageId/feedback", h.authMiddleware, h.SubmitFeedback)
}

type ChatRequest struct {
	Message string `json:"message" validate:"required,min=1,max=2000"`
}

type ExportChatRequest struct {
	Format string `json:"format" validate:"required,oneof=pdf markdown"`
}

func (h *Handler) ExportChat(c *fiber.Ctx) error {
	materialID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return h.errorResponse(c, fiber.StatusBadRequest, "VALIDATION_ERROR", "Invalid material ID")
	}

	var req ExportChatRequest
	if err := c.BodyParser(&req); err != nil {
		return h.errorResponse(c, fiber.StatusBadRequest, "VALIDATION_ERROR", "Invalid request body")
	}

	userID := c.Locals("user_id").(uuid.UUID)

	output, err := h.service.ExportChat(c.Context(), application.ExportChatInput{
		UserID:     userID,
		MaterialID: materialID,
		Format:     domain.ExportFormat(req.Format),
	})
	if err != nil {
		return h.errorResponse(c, fiber.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
	}

	c.Set("Content-Type", output.ContentType)
	c.Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", output.Filename))

	return c.Send(output.Content)
}

func (h *Handler) Chat(c *fiber.Ctx) error {
	materialID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return h.errorResponse(c, fiber.StatusBadRequest, "VALIDATION_ERROR", "Invalid material ID")
	}

	var req ChatRequest
	if err := c.BodyParser(&req); err != nil {
		return h.errorResponse(c, fiber.StatusBadRequest, "VALIDATION_ERROR", "Invalid request body")
	}

	userID := c.Locals("user_id").(uuid.UUID)

	output, err := h.service.Chat(c.Context(), application.ChatInput{
		UserID:     userID,
		MaterialID: &materialID,
		Message:    req.Message,
	})
	if err != nil {
		return h.errorResponse(c, fiber.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
	}

	return h.successResponse(c, fiber.StatusOK, output.Message)
}

func (h *Handler) PodChat(c *fiber.Ctx) error {
	podID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return h.errorResponse(c, fiber.StatusBadRequest, "VALIDATION_ERROR", "Invalid pod ID")
	}

	var req ChatRequest
	if err := c.BodyParser(&req); err != nil {
		return h.errorResponse(c, fiber.StatusBadRequest, "VALIDATION_ERROR", "Invalid request body")
	}

	userID := c.Locals("user_id").(uuid.UUID)

	output, err := h.service.Chat(c.Context(), application.ChatInput{
		UserID:  userID,
		PodID:   &podID,
		Message: req.Message,
	})
	if err != nil {
		return h.errorResponse(c, fiber.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
	}

	return h.successResponse(c, fiber.StatusOK, output.Message)
}

func (h *Handler) GetChatHistory(c *fiber.Ctx) error {
	materialID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return h.errorResponse(c, fiber.StatusBadRequest, "VALIDATION_ERROR", "Invalid material ID")
	}

	userID := c.Locals("user_id").(uuid.UUID)
	limit := c.QueryInt("limit", 50)
	offset := c.QueryInt("offset", 0)

	messages, total, err := h.service.GetChatHistory(c.Context(), userID, &materialID, nil, limit, offset)
	if err != nil {
		return h.errorResponse(c, fiber.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
	}

	return h.paginatedResponse(c, messages, limit, offset, total)
}

func (h *Handler) GetSuggestions(c *fiber.Ctx) error {
	materialID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return h.errorResponse(c, fiber.StatusBadRequest, "VALIDATION_ERROR", "Invalid material ID")
	}

	userID := c.Locals("user_id").(uuid.UUID)

	suggestions, err := h.service.GetSuggestions(c.Context(), application.GetSuggestionsInput{
		UserID:     userID,
		MaterialID: materialID,
	})
	if err != nil {
		return h.errorResponse(c, fiber.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
	}

	return h.successResponse(c, fiber.StatusOK, suggestions)
}

type FeedbackRequest struct {
	Feedback     string  `json:"feedback" validate:"required,oneof=thumbs_up thumbs_down"`
	FeedbackText *string `json:"feedback_text,omitempty"`
}

func (h *Handler) SubmitFeedback(c *fiber.Ctx) error {
	messageID, err := uuid.Parse(c.Params("messageId"))
	if err != nil {
		return h.errorResponse(c, fiber.StatusBadRequest, "VALIDATION_ERROR", "Invalid message ID")
	}

	var req FeedbackRequest
	if err := c.BodyParser(&req); err != nil {
		return h.errorResponse(c, fiber.StatusBadRequest, "VALIDATION_ERROR", "Invalid request body")
	}

	feedback := domain.FeedbackType(req.Feedback)
	if err := h.service.SubmitFeedback(c.Context(), messageID, feedback, req.FeedbackText); err != nil {
		return h.errorResponse(c, fiber.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
	}

	return h.successResponse(c, fiber.StatusOK, fiber.Map{"message": "Feedback submitted"})
}

func (h *Handler) authMiddleware(c *fiber.Ctx) error {
	token := c.Get("Authorization")
	if token == "" {
		return h.errorResponse(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "Missing authorization header")
	}

	if len(token) > 7 && token[:7] == "Bearer " {
		token = token[7:]
	}

	claims, err := h.jwtManager.ValidateAccessToken(token)
	if err != nil {
		return h.errorResponse(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "Invalid token")
	}

	c.Locals("user_id", claims.UserID)
	c.Locals("user_role", claims.Role)

	return c.Next()
}

func (h *Handler) successResponse(c *fiber.Ctx, status int, data interface{}) error {
	return c.Status(status).JSON(fiber.Map{
		"success": true,
		"data":    data,
		"meta": fiber.Map{
			"timestamp":  time.Now().UTC().Format(time.RFC3339),
			"request_id": middleware.GetRequestID(c),
		},
	})
}

func (h *Handler) errorResponse(c *fiber.Ctx, status int, code, message string) error {
	return c.Status(status).JSON(fiber.Map{
		"success": false,
		"error": fiber.Map{
			"code":    code,
			"message": message,
		},
		"meta": fiber.Map{
			"timestamp":  time.Now().UTC().Format(time.RFC3339),
			"request_id": middleware.GetRequestID(c),
		},
	})
}

func (h *Handler) paginatedResponse(c *fiber.Ctx, data interface{}, limit, offset, total int) error {
	page := (offset / limit) + 1
	totalPages := (total + limit - 1) / limit

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"data":    data,
		"pagination": fiber.Map{
			"page":        page,
			"per_page":    limit,
			"total":       total,
			"total_pages": totalPages,
		},
		"meta": fiber.Map{
			"timestamp":  time.Now().UTC().Format(time.RFC3339),
			"request_id": middleware.GetRequestID(c),
		},
	})
}

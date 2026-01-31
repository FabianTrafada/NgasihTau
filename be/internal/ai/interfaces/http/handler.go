package http

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"ngasihtau/docs"
	"ngasihtau/internal/ai/application"
	"ngasihtau/internal/ai/domain"
	"ngasihtau/internal/common/errors"
	"ngasihtau/internal/common/middleware"
	"ngasihtau/pkg/jwt"
)

// Ensure docs package is used for swagger
var _ = docs.Meta{}

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
	materials.Post("/generate-questions", h.authMiddleware, h.GenerateQuestions)

	ai := api.Group("/ai")
	ai.Post("/materials/:id/generate-questions", h.authMiddleware, h.GenerateQuestions)

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

// GenerateQuestionsRequest represents the request body for question generation.
// Implements requirement 12.2.
type GenerateQuestionsRequest struct {
	Count        int    `json:"count,omitempty"`         // default 5, max 20
	QuestionType string `json:"question_type,omitempty"` // multiple_choice, true_false, short_answer, mixed
}

// ExportChat exports chat history to PDF or Markdown.
// @Summary Export chat history
// @Description Export chat history for a material to PDF or Markdown format
// @Tags AI
// @Accept json
// @Produce application/pdf,text/markdown
// @Security BearerAuth
// @Param id path string true "Material ID" format(uuid)
// @Param request body ExportChatRequest true "Export format"
// @Success 200 {file} binary "Exported chat file"
// @Failure 400 {object} docs.ErrorResponse "Invalid request"
// @Failure 401 {object} docs.ErrorResponse "Authentication required"
// @Failure 403 {object} docs.ErrorResponse "Premium subscription required"
// @Router /materials/{id}/chat/export [post]
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

	// Check if user has access to chat export feature (requires premium tier or higher)
	// Implements requirements 11.1, 11.3
	if err := h.service.CheckFeatureAccess(c.Context(), userID, application.FeatureChatExport); err != nil {
		return h.handleFeatureAccessError(c, err)
	}

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

// GenerateQuestions generates quiz questions from material content.
// @Summary Generate quiz questions
// @Description Generate quiz questions from material content using AI
// @Tags AI
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Material ID" format(uuid)
// @Param request body GenerateQuestionsRequest true "Question generation options"
// @Success 200 {object} docs.SuccessResponse "Generated questions"
// @Failure 400 {object} docs.ErrorResponse "Invalid request"
// @Failure 401 {object} docs.ErrorResponse "Authentication required"
// @Failure 403 {object} docs.ErrorResponse "Pro subscription required"
// @Router /ai/materials/{id}/generate-questions [post]
func (h *Handler) GenerateQuestions(c *fiber.Ctx) error {
	materialID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return h.errorResponse(c, fiber.StatusBadRequest, "VALIDATION_ERROR", "Invalid material ID")
	}

	var req GenerateQuestionsRequest
	if err := c.BodyParser(&req); err != nil {
		return h.errorResponse(c, fiber.StatusBadRequest, "VALIDATION_ERROR", "Invalid request body")
	}

	userID := c.Locals("user_id").(uuid.UUID)

	output, err := h.service.GenerateQuestions(c.Context(), application.GenerateQuestionsInput{
		UserID:       userID,
		MaterialID:   materialID,
		Count:        req.Count,
		QuestionType: req.QuestionType,
	})
	if err != nil {
		return h.handleFeatureAccessError(c, err)
	}

	return h.successResponse(c, fiber.StatusOK, output)
}

// Chat handles AI chat with material context.
// @Summary Chat with material
// @Description Send a message to AI with material context using RAG
// @Tags AI
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Material ID" format(uuid)
// @Param request body ChatRequest true "Chat message"
// @Success 200 {object} docs.SuccessResponse "AI response"
// @Failure 400 {object} docs.ErrorResponse "Invalid request"
// @Failure 401 {object} docs.ErrorResponse "Authentication required"
// @Router /materials/{id}/chat [post]
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

// PodChat handles AI chat with pod-wide context.
// @Summary Chat with pod
// @Description Send a message to AI with context from all materials in a pod
// @Tags AI
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Pod ID" format(uuid)
// @Param request body ChatRequest true "Chat message"
// @Success 200 {object} docs.SuccessResponse "AI response with material citations"
// @Failure 400 {object} docs.ErrorResponse "Invalid request"
// @Failure 401 {object} docs.ErrorResponse "Authentication required"
// @Failure 403 {object} docs.ErrorResponse "Pro subscription required"
// @Router /pods/{id}/chat [post]
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

	// Check if user has access to pod-wide chat feature (requires pro tier)
	// Implements requirements 11.2, 11.4
	if err := h.service.CheckFeatureAccess(c.Context(), userID, application.FeaturePodWideChat); err != nil {
		return h.handleFeatureAccessError(c, err)
	}

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

// GetChatHistory retrieves chat history for a material.
// @Summary Get chat history
// @Description Get paginated chat history for a material
// @Tags AI
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Material ID" format(uuid)
// @Param limit query int false "Maximum messages" default(50)
// @Param offset query int false "Offset for pagination" default(0)
// @Success 200 {object} docs.SuccessResponse "Chat history"
// @Failure 400 {object} docs.ErrorResponse "Invalid material ID"
// @Failure 401 {object} docs.ErrorResponse "Authentication required"
// @Router /materials/{id}/chat/history [get]
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

// GetSuggestions retrieves suggested questions for a material.
// @Summary Get suggested questions
// @Description Get AI-generated suggested questions based on material content
// @Tags AI
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Material ID" format(uuid)
// @Success 200 {object} docs.SuggestedQuestionsResponse "Suggested questions"
// @Failure 400 {object} docs.ErrorResponse "Invalid material ID"
// @Failure 401 {object} docs.ErrorResponse "Authentication required"
// @Router /materials/{id}/chat/suggestions [get]
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

	// Transform to string array for response
	questions := make([]string, len(suggestions))
	for i, s := range suggestions {
		questions[i] = s.Question
	}

	return h.successResponse(c, fiber.StatusOK, fiber.Map{
		"questions": questions,
	})
}

type FeedbackRequest struct {
	Feedback     string  `json:"feedback" validate:"required,oneof=thumbs_up thumbs_down"`
	FeedbackText *string `json:"feedback_text,omitempty"`
}

// SubmitFeedback submits feedback for an AI response.
// @Summary Submit feedback
// @Description Submit thumbs up/down feedback for an AI response
// @Tags AI
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param messageId path string true "Message ID" format(uuid)
// @Param request body FeedbackRequest true "Feedback data"
// @Success 200 {object} docs.MessageResponse "Feedback submitted"
// @Failure 400 {object} docs.ErrorResponse "Invalid request"
// @Failure 401 {object} docs.ErrorResponse "Authentication required"
// @Router /chat/{messageId}/feedback [post]
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

// handleFeatureAccessError handles errors from feature access checks.
// It returns appropriate HTTP status codes based on the error type.
// Implements requirements 11.3, 11.4.
func (h *Handler) handleFeatureAccessError(c *fiber.Ctx, err error) error {
	// Check if it's an AppError with a specific code
	if appErr, ok := err.(*errors.AppError); ok {
		return h.errorResponse(c, appErr.HTTPStatus(), string(appErr.Code), appErr.Message)
	}
	// Default to internal error for unknown error types
	return h.errorResponse(c, fiber.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
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

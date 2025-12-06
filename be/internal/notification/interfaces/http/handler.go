package http

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"ngasihtau/internal/common/middleware"
	"ngasihtau/internal/notification/application"
	"ngasihtau/pkg/jwt"
)

type Handler struct {
	service    *application.NotificationService
	jwtManager *jwt.Manager
}

func NewHandler(service *application.NotificationService, jwtManager *jwt.Manager) *Handler {
	return &Handler{
		service:    service,
		jwtManager: jwtManager,
	}
}

func (h *Handler) RegisterRoutes(app *fiber.App) {
	api := app.Group("/api/v1")

	notifications := api.Group("/notifications", middleware.Auth(h.jwtManager))
	notifications.Get("/", h.GetNotifications)
	notifications.Put("/:id/read", h.MarkAsRead)
	notifications.Put("/read-all", h.MarkAllAsRead)
	notifications.Get("/preferences", h.GetPreferences)
	notifications.Put("/preferences", h.UpdatePreferences)
}

// GetNotifications handles GET /api/v1/notifications
func (h *Handler) GetNotifications(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return errorResponse(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated")
	}

	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	offset, _ := strconv.Atoi(c.Query("offset", "0"))

	output, err := h.service.GetNotifications(c.Context(), application.GetNotificationsInput{
		UserID: userID,
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return errorResponse(c, fiber.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get notifications")
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    output.Notifications,
		"pagination": fiber.Map{
			"total":        output.Total,
			"limit":        limit,
			"offset":       offset,
			"unread_count": output.UnreadCount,
		},
		"meta": fiber.Map{
			"timestamp":  time.Now().UTC().Format(time.RFC3339),
			"request_id": middleware.GetRequestID(c),
		},
	})
}

// MarkAsRead handles PUT /api/v1/notifications/:id/read
func (h *Handler) MarkAsRead(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return errorResponse(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated")
	}

	notificationID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "VALIDATION_ERROR", "Invalid notification ID")
	}

	if err := h.service.MarkAsRead(c.Context(), userID, notificationID); err != nil {
		if err == application.ErrNotificationNotFound {
			return errorResponse(c, fiber.StatusNotFound, "NOT_FOUND", "Notification not found")
		}
		return errorResponse(c, fiber.StatusInternalServerError, "INTERNAL_ERROR", "Failed to mark notification as read")
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{
			"message": "Notification marked as read",
		},
		"meta": fiber.Map{
			"timestamp":  time.Now().UTC().Format(time.RFC3339),
			"request_id": middleware.GetRequestID(c),
		},
	})
}

// MarkAllAsRead handles PUT /api/v1/notifications/read-all
func (h *Handler) MarkAllAsRead(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return errorResponse(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated")
	}

	if err := h.service.MarkAllAsRead(c.Context(), userID); err != nil {
		return errorResponse(c, fiber.StatusInternalServerError, "INTERNAL_ERROR", "Failed to mark all notifications as read")
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{
			"message": "All notifications marked as read",
		},
		"meta": fiber.Map{
			"timestamp":  time.Now().UTC().Format(time.RFC3339),
			"request_id": middleware.GetRequestID(c),
		},
	})
}

// GetPreferences handles GET /api/v1/notifications/preferences
func (h *Handler) GetPreferences(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return errorResponse(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated")
	}

	pref, err := h.service.GetPreferences(c.Context(), userID)
	if err != nil {
		return errorResponse(c, fiber.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get preferences")
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    pref,
		"meta": fiber.Map{
			"timestamp":  time.Now().UTC().Format(time.RFC3339),
			"request_id": middleware.GetRequestID(c),
		},
	})
}

// UpdatePreferencesRequest represents the request body for updating preferences.
type UpdatePreferencesRequest struct {
	EmailPodInvite    *bool `json:"email_pod_invite"`
	EmailNewMaterial  *bool `json:"email_new_material"`
	EmailCommentReply *bool `json:"email_comment_reply"`
	InAppPodInvite    *bool `json:"inapp_pod_invite"`
	InAppNewMaterial  *bool `json:"inapp_new_material"`
	InAppCommentReply *bool `json:"inapp_comment_reply"`
}

// UpdatePreferences handles PUT /api/v1/notifications/preferences
func (h *Handler) UpdatePreferences(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return errorResponse(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated")
	}

	var req UpdatePreferencesRequest
	if err := c.BodyParser(&req); err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "VALIDATION_ERROR", "Invalid request body")
	}

	pref, err := h.service.UpdatePreferences(c.Context(), application.UpdatePreferencesInput{
		UserID:            userID,
		EmailPodInvite:    req.EmailPodInvite,
		EmailNewMaterial:  req.EmailNewMaterial,
		EmailCommentReply: req.EmailCommentReply,
		InAppPodInvite:    req.InAppPodInvite,
		InAppNewMaterial:  req.InAppNewMaterial,
		InAppCommentReply: req.InAppCommentReply,
	})
	if err != nil {
		return errorResponse(c, fiber.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update preferences")
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    pref,
		"meta": fiber.Map{
			"timestamp":  time.Now().UTC().Format(time.RFC3339),
			"request_id": middleware.GetRequestID(c),
		},
	})
}

// errorResponse returns a standardized error response.
func errorResponse(c *fiber.Ctx, status int, code, message string) error {
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

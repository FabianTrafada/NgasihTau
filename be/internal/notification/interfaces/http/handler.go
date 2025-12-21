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
// @Summary Get notifications
// @Description Get a paginated list of notifications for the authenticated user
// @Tags Notifications
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param limit query int false "Maximum notifications" default(20)
// @Param offset query int false "Offset for pagination" default(0)
// @Success 200 {object} fiber.Map "Notifications with unread count"
// @Failure 401 {object} fiber.Map "Authentication required"
// @Router /notifications [get]
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
// @Summary Mark notification as read
// @Description Mark a single notification as read
// @Tags Notifications
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Notification ID" format(uuid)
// @Success 200 {object} fiber.Map "Notification marked as read"
// @Failure 400 {object} fiber.Map "Invalid notification ID"
// @Failure 401 {object} fiber.Map "Authentication required"
// @Failure 404 {object} fiber.Map "Notification not found"
// @Router /notifications/{id}/read [put]
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
// @Summary Mark all notifications as read
// @Description Mark all notifications as read for the authenticated user
// @Tags Notifications
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} fiber.Map "All notifications marked as read"
// @Failure 401 {object} fiber.Map "Authentication required"
// @Router /notifications/read-all [put]
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
// @Summary Get notification preferences
// @Description Get notification preferences for the authenticated user
// @Tags Notifications
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} fiber.Map "Notification preferences"
// @Failure 401 {object} fiber.Map "Authentication required"
// @Router /notifications/preferences [get]
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
// @Summary Update notification preferences
// @Description Update notification preferences for the authenticated user
// @Tags Notifications
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body UpdatePreferencesRequest true "Preference updates"
// @Success 200 {object} fiber.Map "Updated preferences"
// @Failure 400 {object} fiber.Map "Invalid request body"
// @Failure 401 {object} fiber.Map "Authentication required"
// @Router /notifications/preferences [put]
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

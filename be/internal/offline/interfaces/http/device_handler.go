package http

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"ngasihtau/internal/common/errors"
	"ngasihtau/internal/common/middleware"
	"ngasihtau/internal/common/response"
	"ngasihtau/internal/common/validator"
	"ngasihtau/internal/offline/application"
	"ngasihtau/internal/offline/domain"
)

// DeviceHandler handles device-related HTTP requests.
type DeviceHandler struct {
	deviceService application.DeviceService
}

// NewDeviceHandler creates a new DeviceHandler.
func NewDeviceHandler(deviceService application.DeviceService) *DeviceHandler {
	return &DeviceHandler{
		deviceService: deviceService,
	}
}

// RegisterDeviceRequest represents the request body for device registration.
type RegisterDeviceRequest struct {
	Fingerprint string `json:"fingerprint" validate:"required,min=32,max=512"`
	Name        string `json:"name" validate:"required,min=1,max=255"`
	Platform    string `json:"platform" validate:"required,oneof=ios android desktop"`
}

// DeviceResponse represents a device in API responses.
type DeviceResponse struct {
	ID          uuid.UUID `json:"id"`
	UserID      uuid.UUID `json:"user_id"`
	Fingerprint string    `json:"fingerprint"`
	Name        string    `json:"name"`
	Platform    string    `json:"platform"`
	Status      string    `json:"status"`
	LastUsedAt  time.Time `json:"last_used_at"`
	CreatedAt   time.Time `json:"created_at"`
}

// ToDeviceResponse converts a domain.Device to DeviceResponse.
func ToDeviceResponse(device *domain.Device) *DeviceResponse {
	if device == nil {
		return nil
	}

	return &DeviceResponse{
		ID:          device.ID,
		UserID:      device.UserID,
		Fingerprint: device.Fingerprint,
		Name:        device.Name,
		Platform:    string(device.Platform),
		Status:      string(device.Status()), // Status() is a method
		LastUsedAt:  device.LastUsedAt,
		CreatedAt:   device.CreatedAt,
	}
}

// ToDeviceResponseList converts a slice of domain.Device to DeviceResponse slice.
func ToDeviceResponseList(devices []*domain.Device) []*DeviceResponse {
	result := make([]*DeviceResponse, len(devices))
	for i, device := range devices {
		result[i] = ToDeviceResponse(device)
	}
	return result
}

// RegisterDevice handles device registration.
// @Summary Register a new device
// @Description Register a new device for offline material access. Maximum 5 devices per user.
// @Tags Offline
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body RegisterDeviceRequest true "Device registration data"
// @Success 201 {object} response.Response[DeviceResponse] "Device registered successfully"
// @Failure 400 {object} errors.ErrorResponse "Invalid request body"
// @Failure 401 {object} errors.ErrorResponse "Authentication required"
// @Failure 403 {object} errors.ErrorResponse "Device limit exceeded (max 5)"
// @Router /offline/devices [post]
func (h *DeviceHandler) RegisterDevice(c *fiber.Ctx) error {
	requestID := middleware.GetRequestID(c)

	// Get user ID from context (set by auth middleware)
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return sendError(c, requestID, errors.Unauthorized(""))
	}

	var req RegisterDeviceRequest
	if err := c.BodyParser(&req); err != nil {
		return sendError(c, requestID, errors.BadRequest("invalid request body"))
	}

	// Validate request
	if err := validator.Get().Struct(&req); err != nil {
		return sendError(c, requestID, err)
	}

	// Call service
	device, err := h.deviceService.RegisterDevice(c.Context(), application.RegisterDeviceInput{
		UserID:      userID,
		Fingerprint: req.Fingerprint,
		Name:        req.Name,
		Platform:    domain.Platform(req.Platform),
	})
	if err != nil {
		return sendError(c, requestID, err)
	}

	return c.Status(fiber.StatusCreated).JSON(response.Created(requestID, ToDeviceResponse(device)))
}

// ListDevices handles listing user's devices.
// @Summary List user's devices
// @Description Get all registered devices for the authenticated user
// @Tags Offline
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response[[]DeviceResponse] "List of devices"
// @Failure 401 {object} errors.ErrorResponse "Authentication required"
// @Router /offline/devices [get]
func (h *DeviceHandler) ListDevices(c *fiber.Ctx) error {
	requestID := middleware.GetRequestID(c)

	// Get user ID from context
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return sendError(c, requestID, errors.Unauthorized(""))
	}

	// Call service
	devices, err := h.deviceService.ListDevices(c.Context(), userID)
	if err != nil {
		return sendError(c, requestID, err)
	}

	return c.Status(fiber.StatusOK).JSON(response.OK(requestID, ToDeviceResponseList(devices)))
}

// DeregisterDevice handles device deregistration.
// @Summary Deregister a device
// @Description Remove a device and revoke all associated licenses
// @Tags Offline
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param device_id path string true "Device ID" format(uuid)
// @Success 200 {object} response.Response[any] "Device deregistered successfully"
// @Failure 400 {object} errors.ErrorResponse "Invalid device ID"
// @Failure 401 {object} errors.ErrorResponse "Authentication required"
// @Failure 404 {object} errors.ErrorResponse "Device not found"
// @Router /offline/devices/{device_id} [delete]
func (h *DeviceHandler) DeregisterDevice(c *fiber.Ctx) error {
	requestID := middleware.GetRequestID(c)

	// Get user ID from context
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return sendError(c, requestID, errors.Unauthorized(""))
	}

	// Parse device ID from path
	deviceIDStr := c.Params("device_id")
	deviceID, err := uuid.Parse(deviceIDStr)
	if err != nil {
		return sendError(c, requestID, errors.BadRequest("invalid device ID"))
	}

	// Call service
	if err := h.deviceService.DeregisterDevice(c.Context(), userID, deviceID); err != nil {
		return sendError(c, requestID, err)
	}

	return c.Status(fiber.StatusOK).JSON(response.Empty(requestID))
}

// sendError sends an error response using the standard error format.
func sendError(c *fiber.Ctx, requestID string, err error) error {
	appErr, ok := err.(*errors.AppError)
	if !ok {
		appErr = errors.Internal("internal server error", err)
	}
	resp := errors.BuildResponse(requestID, appErr)
	return c.Status(appErr.HTTPStatus()).JSON(resp)
}

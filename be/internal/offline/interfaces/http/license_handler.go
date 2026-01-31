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

// LicenseHandler handles license-related HTTP requests.
type LicenseHandler struct {
	licenseService application.LicenseService
}

// NewLicenseHandler creates a new LicenseHandler.
func NewLicenseHandler(licenseService application.LicenseService) *LicenseHandler {
	return &LicenseHandler{
		licenseService: licenseService,
	}
}

// IssueLicenseRequest represents the request body for license issuance.
type IssueLicenseRequest struct {
	DeviceID    string `json:"device_id" validate:"required,uuid"`
	Fingerprint string `json:"fingerprint" validate:"required,min=32,max=512"`
}

// ValidateLicenseRequest represents the request body for license validation.
type ValidateLicenseRequest struct {
	Fingerprint string `json:"fingerprint" validate:"required,min=32,max=512"`
	Nonce       string `json:"nonce" validate:"required,min=16,max=64"`
}

// LicenseResponse represents a license in API responses.
type LicenseResponse struct {
	ID                 uuid.UUID `json:"id"`
	UserID             uuid.UUID `json:"user_id"`
	MaterialID         uuid.UUID `json:"material_id"`
	DeviceID           uuid.UUID `json:"device_id"`
	Status             string    `json:"status"`
	ExpiresAt          time.Time `json:"expires_at"`
	OfflineGracePeriod string    `json:"offline_grace_period"`
	LastValidatedAt    time.Time `json:"last_validated_at"`
	Nonce              string    `json:"nonce"`
	CreatedAt          time.Time `json:"created_at"`
}

// ToLicenseResponse converts a domain.License to LicenseResponse.
func ToLicenseResponse(license *domain.License) *LicenseResponse {
	if license == nil {
		return nil
	}

	return &LicenseResponse{
		ID:                 license.ID,
		UserID:             license.UserID,
		MaterialID:         license.MaterialID,
		DeviceID:           license.DeviceID,
		Status:             string(license.Status),
		ExpiresAt:          license.ExpiresAt,
		OfflineGracePeriod: license.OfflineGracePeriod.String(),
		LastValidatedAt:    license.LastValidatedAt,
		Nonce:              license.Nonce,
		CreatedAt:          license.CreatedAt,
	}
}

// IssueLicense handles license issuance for a material.
// @Summary Issue a license for offline access
// @Description Issue a new license for offline access to a material on a specific device
// @Tags Offline
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param material_id path string true "Material ID" format(uuid)
// @Param request body IssueLicenseRequest true "License issuance data"
// @Success 201 {object} response.Response[LicenseResponse] "License issued successfully"
// @Failure 400 {object} errors.ErrorResponse "Invalid request body"
// @Failure 401 {object} errors.ErrorResponse "Authentication required"
// @Failure 403 {object} errors.ErrorResponse "Access denied to material"
// @Failure 404 {object} errors.ErrorResponse "Material or device not found"
// @Router /offline/materials/{material_id}/license [post]
func (h *LicenseHandler) IssueLicense(c *fiber.Ctx) error {
	requestID := middleware.GetRequestID(c)

	// Get user ID from context
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return sendError(c, requestID, errors.Unauthorized(""))
	}

	// Parse material ID from path
	materialIDStr := c.Params("material_id")
	materialID, err := uuid.Parse(materialIDStr)
	if err != nil {
		return sendError(c, requestID, errors.BadRequest("invalid material ID"))
	}

	var req IssueLicenseRequest
	if err := c.BodyParser(&req); err != nil {
		return sendError(c, requestID, errors.BadRequest("invalid request body"))
	}

	// Validate request
	if err := validator.Get().Struct(&req); err != nil {
		return sendError(c, requestID, err)
	}

	// Parse device ID
	deviceID, err := uuid.Parse(req.DeviceID)
	if err != nil {
		return sendError(c, requestID, errors.BadRequest("invalid device ID"))
	}

	// Call service
	license, err := h.licenseService.IssueLicense(c.Context(), application.IssueLicenseInput{
		UserID:      userID,
		MaterialID:  materialID,
		DeviceID:    deviceID,
		Fingerprint: req.Fingerprint,
	})
	if err != nil {
		return sendError(c, requestID, err)
	}

	return c.Status(fiber.StatusCreated).JSON(response.Created(requestID, ToLicenseResponse(license)))
}

// ValidateLicense handles license validation.
// @Summary Validate a license
// @Description Validate a license for offline access, updating last validated timestamp
// @Tags Offline
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param license_id path string true "License ID" format(uuid)
// @Param request body ValidateLicenseRequest true "License validation data"
// @Success 200 {object} response.Response[LicenseResponse] "License validated successfully"
// @Failure 400 {object} errors.ErrorResponse "Invalid request body"
// @Failure 401 {object} errors.ErrorResponse "Authentication required"
// @Failure 403 {object} errors.ErrorResponse "License expired or invalid"
// @Failure 404 {object} errors.ErrorResponse "License not found"
// @Router /offline/licenses/{license_id}/validate [post]
func (h *LicenseHandler) ValidateLicense(c *fiber.Ctx) error {
	requestID := middleware.GetRequestID(c)

	// Get user ID from context
	if _, ok := middleware.GetUserID(c); !ok {
		return sendError(c, requestID, errors.Unauthorized(""))
	}

	// Parse license ID from path
	licenseIDStr := c.Params("license_id")
	licenseID, err := uuid.Parse(licenseIDStr)
	if err != nil {
		return sendError(c, requestID, errors.BadRequest("invalid license ID"))
	}

	var req ValidateLicenseRequest
	if err := c.BodyParser(&req); err != nil {
		return sendError(c, requestID, errors.BadRequest("invalid request body"))
	}

	// Validate request
	if err := validator.Get().Struct(&req); err != nil {
		return sendError(c, requestID, err)
	}

	// Call service
	license, err := h.licenseService.ValidateLicense(c.Context(), application.ValidateLicenseInput{
		LicenseID:   licenseID,
		DeviceID:    uuid.Nil, // Will be validated from license
		Fingerprint: req.Fingerprint,
		Nonce:       req.Nonce,
	})
	if err != nil {
		return sendError(c, requestID, err)
	}

	return c.Status(fiber.StatusOK).JSON(response.OK(requestID, ToLicenseResponse(license)))
}

// RenewLicense handles license renewal.
// @Summary Renew a license
// @Description Renew a license before it expires, extending the expiration date
// @Tags Offline
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param license_id path string true "License ID" format(uuid)
// @Success 200 {object} response.Response[LicenseResponse] "License renewed successfully"
// @Failure 400 {object} errors.ErrorResponse "Invalid license ID"
// @Failure 401 {object} errors.ErrorResponse "Authentication required"
// @Failure 403 {object} errors.ErrorResponse "License already expired or revoked"
// @Failure 404 {object} errors.ErrorResponse "License not found"
// @Router /offline/licenses/{license_id}/renew [post]
func (h *LicenseHandler) RenewLicense(c *fiber.Ctx) error {
	requestID := middleware.GetRequestID(c)

	// Get user ID from context
	_, ok := middleware.GetUserID(c)
	if !ok {
		return sendError(c, requestID, errors.Unauthorized(""))
	}

	// Parse license ID from path
	licenseIDStr := c.Params("license_id")
	licenseID, err := uuid.Parse(licenseIDStr)
	if err != nil {
		return sendError(c, requestID, errors.BadRequest("invalid license ID"))
	}

	// Call service - RenewLicense takes RenewLicenseInput
	// For renewal, we need fingerprint from request body
	var req struct {
		Fingerprint string `json:"fingerprint" validate:"required,min=32,max=512"`
	}
	if err := c.BodyParser(&req); err != nil {
		return sendError(c, requestID, errors.BadRequest("invalid request body"))
	}

	if err := validator.Get().Struct(&req); err != nil {
		return sendError(c, requestID, err)
	}

	license, err := h.licenseService.RenewLicense(c.Context(), application.RenewLicenseInput{
		LicenseID:   licenseID,
		DeviceID:    uuid.Nil, // Will be validated from license
		Fingerprint: req.Fingerprint,
	})
	if err != nil {
		return sendError(c, requestID, err)
	}

	return c.Status(fiber.StatusOK).JSON(response.OK(requestID, ToLicenseResponse(license)))
}

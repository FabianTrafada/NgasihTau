package http

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"ngasihtau/internal/common/middleware"
	"ngasihtau/internal/material/application"
)

// UploadURLRequest represents the request body for getting an upload URL.
type UploadURLRequest struct {
	Filename    string `json:"filename" validate:"required"`
	ContentType string `json:"content_type" validate:"required"`
	Size        int64  `json:"size" validate:"required,max=104857600"` // 100MB max
}

// GetUploadURL generates a presigned URL for uploading a file.
// POST /api/v1/materials/upload-url
func (h *Handler) GetUploadURL(c *fiber.Ctx) error {
	var req UploadURLRequest
	if err := c.BodyParser(&req); err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "VALIDATION_ERROR", "Invalid request body")
	}

	// Validate content type
	validTypes := map[string]bool{
		"application/pdf": true,
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document":   true,
		"application/vnd.openxmlformats-officedocument.presentationml.presentation": true,
	}
	if !validTypes[req.ContentType] {
		return errorResponse(c, fiber.StatusBadRequest, "VALIDATION_ERROR", "Unsupported file type. Allowed: PDF, DOCX, PPTX")
	}

	result, err := h.service.GetUploadURL(c.Context(), application.UploadURLInput{
		Filename:    req.Filename,
		ContentType: req.ContentType,
		Size:        req.Size,
	})
	if err != nil {
		return errorResponse(c, fiber.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
	}

	return successResponse(c, result)
}

// ConfirmUploadRequest represents the request body for confirming an upload.
type ConfirmUploadRequest struct {
	ObjectKey   string  `json:"object_key" validate:"required"`
	PodID       string  `json:"pod_id" validate:"required,uuid"`
	Title       string  `json:"title" validate:"required,max=255"`
	Description *string `json:"description" validate:"omitempty,max=2000"`
}

// ConfirmUpload confirms a file upload and creates a material record.
// POST /api/v1/materials/confirm
func (h *Handler) ConfirmUpload(c *fiber.Ctx) error {
	var req ConfirmUploadRequest
	if err := c.BodyParser(&req); err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "VALIDATION_ERROR", "Invalid request body")
	}

	podID, err := uuid.Parse(req.PodID)
	if err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "VALIDATION_ERROR", "Invalid pod ID")
	}

	userID, ok := middleware.GetUserID(c)
	if !ok {
		return errorResponse(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated")
	}

	material, err := h.service.ConfirmUpload(c.Context(), application.ConfirmUploadInput{
		ObjectKey:   req.ObjectKey,
		PodID:       podID,
		UploaderID:  userID,
		Title:       req.Title,
		Description: req.Description,
	})
	if err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "UPLOAD_ERROR", err.Error())
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"data":    material,
	})
}

// GetMaterial retrieves a material by ID.
// GET /api/v1/materials/:id
func (h *Handler) GetMaterial(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "VALIDATION_ERROR", "Invalid material ID")
	}

	material, err := h.service.GetMaterial(c.Context(), id)
	if err != nil {
		return errorResponse(c, fiber.StatusNotFound, "NOT_FOUND", "Material not found")
	}

	return successResponse(c, material)
}

// UpdateMaterialRequest represents the request body for updating a material.
type UpdateMaterialRequest struct {
	Title       *string `json:"title" validate:"omitempty,max=255"`
	Description *string `json:"description" validate:"omitempty,max=2000"`
}

// UpdateMaterial updates a material's metadata.
// PUT /api/v1/materials/:id
func (h *Handler) UpdateMaterial(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "VALIDATION_ERROR", "Invalid material ID")
	}

	var req UpdateMaterialRequest
	if err := c.BodyParser(&req); err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "VALIDATION_ERROR", "Invalid request body")
	}

	material, err := h.service.UpdateMaterial(c.Context(), application.UpdateMaterialInput{
		ID:          id,
		Title:       req.Title,
		Description: req.Description,
	})
	if err != nil {
		return errorResponse(c, fiber.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
	}

	return successResponse(c, material)
}

// DeleteMaterial soft-deletes a material.
// DELETE /api/v1/materials/:id
func (h *Handler) DeleteMaterial(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "VALIDATION_ERROR", "Invalid material ID")
	}

	if err := h.service.DeleteMaterial(c.Context(), id); err != nil {
		return errorResponse(c, fiber.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
	}

	return c.Status(fiber.StatusNoContent).Send(nil)
}

// GetMaterialsByPod retrieves all materials in a pod.
// GET /api/v1/pods/:podId/materials
func (h *Handler) GetMaterialsByPod(c *fiber.Ctx) error {
	podID, err := uuid.Parse(c.Params("podId"))
	if err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "VALIDATION_ERROR", "Invalid pod ID")
	}

	page := c.QueryInt("page", 1)
	perPage := c.QueryInt("per_page", 20)
	if perPage > 100 {
		perPage = 100
	}
	offset := (page - 1) * perPage

	materials, total, err := h.service.GetMaterialsByPod(c.Context(), podID, perPage, offset)
	if err != nil {
		return errorResponse(c, fiber.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
	}

	return paginatedResponse(c, materials, page, perPage, total)
}

// GetPreviewURL generates a presigned URL for previewing a material.
// GET /api/v1/materials/:id/preview
func (h *Handler) GetPreviewURL(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "VALIDATION_ERROR", "Invalid material ID")
	}

	url, err := h.service.GetPreviewURL(c.Context(), id)
	if err != nil {
		return errorResponse(c, fiber.StatusNotFound, "NOT_FOUND", "Material not found")
	}

	return successResponse(c, fiber.Map{"preview_url": url})
}

// GetDownloadURL generates a presigned URL for downloading a material.
// GET /api/v1/materials/:id/download
func (h *Handler) GetDownloadURL(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "VALIDATION_ERROR", "Invalid material ID")
	}

	url, err := h.service.GetDownloadURL(c.Context(), id)
	if err != nil {
		return errorResponse(c, fiber.StatusNotFound, "NOT_FOUND", "Material not found")
	}

	return successResponse(c, fiber.Map{"download_url": url})
}

// CreateVersionRequest represents the request body for creating a new version.
type CreateVersionRequest struct {
	ObjectKey string  `json:"object_key" validate:"required"`
	Changelog *string `json:"changelog" validate:"omitempty,max=1000"`
}

// CreateVersion creates a new version of a material.
// POST /api/v1/materials/:id/versions
func (h *Handler) CreateVersion(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "VALIDATION_ERROR", "Invalid material ID")
	}

	var req CreateVersionRequest
	if err := c.BodyParser(&req); err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "VALIDATION_ERROR", "Invalid request body")
	}

	userID, ok := middleware.GetUserID(c)
	if !ok {
		return errorResponse(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated")
	}

	version, err := h.service.CreateVersion(c.Context(), application.CreateVersionInput{
		MaterialID: id,
		ObjectKey:  req.ObjectKey,
		UploaderID: userID,
		Changelog:  req.Changelog,
	})
	if err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "VERSION_ERROR", err.Error())
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"data":    version,
	})
}

// GetVersionHistory retrieves all versions of a material.
// GET /api/v1/materials/:id/versions
func (h *Handler) GetVersionHistory(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "VALIDATION_ERROR", "Invalid material ID")
	}

	versions, err := h.service.GetVersionHistory(c.Context(), id)
	if err != nil {
		return errorResponse(c, fiber.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
	}

	return successResponse(c, versions)
}

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
// @Summary Get presigned upload URL
// @Description Generate a presigned URL for direct file upload to storage
// @Tags Materials
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body UploadURLRequest true "Upload request details"
// @Success 200 {object} fiber.Map "Upload URL and object key"
// @Failure 400 {object} fiber.Map "Invalid request or unsupported file type"
// @Failure 401 {object} fiber.Map "Authentication required"
// @Router /materials/upload-url [post]
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
// @Summary Confirm file upload
// @Description Confirm a file upload and create a material record
// @Tags Materials
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body ConfirmUploadRequest true "Upload confirmation details"
// @Success 201 {object} fiber.Map "Created material"
// @Failure 400 {object} fiber.Map "Invalid request or upload error"
// @Failure 401 {object} fiber.Map "Authentication required"
// @Router /materials/confirm [post]
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
// @Summary Get a material
// @Description Get a material by ID
// @Tags Materials
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Material ID" format(uuid)
// @Success 200 {object} fiber.Map "Material details"
// @Failure 400 {object} fiber.Map "Invalid material ID"
// @Failure 404 {object} fiber.Map "Material not found"
// @Router /materials/{id} [get]
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
// @Summary Update a material
// @Description Update a material's title and description
// @Tags Materials
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Material ID" format(uuid)
// @Param request body UpdateMaterialRequest true "Material update data"
// @Success 200 {object} fiber.Map "Updated material"
// @Failure 400 {object} fiber.Map "Invalid request"
// @Failure 401 {object} fiber.Map "Authentication required"
// @Router /materials/{id} [put]
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
// @Summary Delete a material
// @Description Soft-delete a material
// @Tags Materials
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Material ID" format(uuid)
// @Success 204 "Material deleted"
// @Failure 400 {object} fiber.Map "Invalid material ID"
// @Failure 401 {object} fiber.Map "Authentication required"
// @Router /materials/{id} [delete]
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
// @Summary List materials in a pod
// @Description Get a paginated list of materials in a Knowledge Pod
// @Tags Materials
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param podId path string true "Pod ID" format(uuid)
// @Param page query int false "Page number" default(1)
// @Param per_page query int false "Items per page" default(20) maximum(100)
// @Success 200 {object} fiber.Map "List of materials"
// @Failure 400 {object} fiber.Map "Invalid pod ID"
// @Router /pods/{podId}/materials [get]
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
// @Summary Get preview URL
// @Description Generate a presigned URL for previewing a material
// @Tags Materials
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Material ID" format(uuid)
// @Success 200 {object} fiber.Map "Preview URL"
// @Failure 400 {object} fiber.Map "Invalid material ID"
// @Failure 404 {object} fiber.Map "Material not found"
// @Router /materials/{id}/preview [get]
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
// @Summary Get download URL
// @Description Generate a presigned URL for downloading a material
// @Tags Materials
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Material ID" format(uuid)
// @Success 200 {object} fiber.Map "Download URL"
// @Failure 400 {object} fiber.Map "Invalid material ID"
// @Failure 404 {object} fiber.Map "Material not found"
// @Router /materials/{id}/download [get]
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
// @Summary Create a new version
// @Description Upload a new version of an existing material
// @Tags Materials
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Material ID" format(uuid)
// @Param request body CreateVersionRequest true "Version details"
// @Success 201 {object} fiber.Map "Created version"
// @Failure 400 {object} fiber.Map "Invalid request"
// @Failure 401 {object} fiber.Map "Authentication required"
// @Router /materials/{id}/versions [post]
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
// @Summary Get version history
// @Description Get all versions of a material
// @Tags Materials
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Material ID" format(uuid)
// @Success 200 {object} fiber.Map "List of versions"
// @Failure 400 {object} fiber.Map "Invalid material ID"
// @Router /materials/{id}/versions [get]
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

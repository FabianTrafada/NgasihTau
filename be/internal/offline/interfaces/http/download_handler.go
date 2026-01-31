package http

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"ngasihtau/internal/common/errors"
	"ngasihtau/internal/common/middleware"
	"ngasihtau/internal/common/response"
	"ngasihtau/internal/offline/application"
	"ngasihtau/internal/offline/domain"
)

// DownloadHandler handles download-related HTTP requests.
type DownloadHandler struct {
	downloadService   application.DownloadService
	licenseService    application.LicenseService
	encryptionService *application.EncryptionService
}

// NewDownloadHandler creates a new DownloadHandler.
func NewDownloadHandler(
	downloadService application.DownloadService,
	licenseService application.LicenseService,
	encryptionService *application.EncryptionService,
) *DownloadHandler {
	return &DownloadHandler{
		downloadService:   downloadService,
		licenseService:    licenseService,
		encryptionService: encryptionService,
	}
}

// DownloadRequest represents the query parameters for download.
type DownloadRequest struct {
	DeviceID    string `query:"device_id" validate:"required,uuid"`
	Fingerprint string `query:"fingerprint" validate:"required,min=32,max=512"`
	LicenseID   string `query:"license_id" validate:"required,uuid"`
}

// ManifestResponse represents the download manifest in API responses.
type ManifestResponse struct {
	MaterialID    uuid.UUID              `json:"material_id"`
	LicenseID     uuid.UUID              `json:"license_id"`
	TotalChunks   int                    `json:"total_chunks"`
	TotalSize     int64                  `json:"total_size"`
	OriginalHash  string                 `json:"original_hash"`
	EncryptedHash string                 `json:"encrypted_hash"`
	ChunkSize     int64                  `json:"chunk_size"`
	Chunks        []ChunkResponse        `json:"chunks"`
	FileType      string                 `json:"file_type"`
	CreatedAt     time.Time              `json:"created_at"`
}

// ChunkResponse represents a chunk in API responses.
type ChunkResponse struct {
	Index   int    `json:"index"`
	Offset  int64  `json:"offset"`
	Size    int64  `json:"size"`
}

// DownloadResponse represents the download response.
type DownloadResponse struct {
	Manifest    ManifestResponse `json:"manifest"`
	DownloadURL string           `json:"download_url"`
	ExpiresAt   time.Time        `json:"expires_at"`
}

// ToManifestResponse converts a domain.DownloadManifest to ManifestResponse.
func ToManifestResponse(manifest *domain.DownloadManifest) *ManifestResponse {
	if manifest == nil {
		return nil
	}

	chunks := make([]ChunkResponse, len(manifest.Chunks))
	for i, chunk := range manifest.Chunks {
		chunks[i] = ChunkResponse{
			Index:  chunk.Index,
			Offset: chunk.Offset,
			Size:   chunk.Size,
		}
	}

	return &ManifestResponse{
		MaterialID:    manifest.MaterialID,
		LicenseID:     manifest.LicenseID,
		TotalChunks:   manifest.TotalChunks,
		TotalSize:     manifest.TotalSize,
		OriginalHash:  manifest.OriginalHash,
		EncryptedHash: manifest.EncryptedHash,
		ChunkSize:     manifest.ChunkSize,
		Chunks:        chunks,
		FileType:      manifest.FileType,
		CreatedAt:     manifest.CreatedAt,
	}
}

// DownloadMaterial handles material download requests.
// @Summary Download encrypted material
// @Description Download encrypted material with manifest for offline access
// @Tags Offline
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param material_id path string true "Material ID" format(uuid)
// @Param device_id query string true "Device ID" format(uuid)
// @Param fingerprint query string true "Device fingerprint"
// @Param license_id query string true "License ID" format(uuid)
// @Param Range header string false "Range header for partial content"
// @Success 200 {object} response.Response[DownloadResponse] "Download manifest and URL"
// @Success 206 {object} response.Response[DownloadResponse] "Partial content (range request)"
// @Failure 400 {object} errors.ErrorResponse "Invalid request"
// @Failure 401 {object} errors.ErrorResponse "Authentication required"
// @Failure 403 {object} errors.ErrorResponse "Access denied or license invalid"
// @Failure 404 {object} errors.ErrorResponse "Material or license not found"
// @Failure 429 {object} errors.ErrorResponse "Rate limit exceeded"
// @Router /offline/materials/{material_id}/download [get]
func (h *DownloadHandler) DownloadMaterial(c *fiber.Ctx) error {
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

	// Parse query parameters
	deviceIDStr := c.Query("device_id")
	if deviceIDStr == "" {
		return sendError(c, requestID, errors.BadRequest("device_id is required"))
	}
	deviceID, err := uuid.Parse(deviceIDStr)
	if err != nil {
		return sendError(c, requestID, errors.BadRequest("invalid device ID"))
	}

	fingerprint := c.Query("fingerprint")
	if fingerprint == "" {
		return sendError(c, requestID, errors.BadRequest("fingerprint is required"))
	}
	if len(fingerprint) < domain.MinFingerprintLength || len(fingerprint) > domain.MaxFingerprintLength {
		return sendError(c, requestID, errors.BadRequest(fmt.Sprintf("fingerprint must be between %d and %d characters", domain.MinFingerprintLength, domain.MaxFingerprintLength)))
	}

	licenseIDStr := c.Query("license_id")
	if licenseIDStr == "" {
		return sendError(c, requestID, errors.BadRequest("license_id is required"))
	}
	licenseID, err := uuid.Parse(licenseIDStr)
	if err != nil {
		return sendError(c, requestID, errors.BadRequest("invalid license ID"))
	}

	// Check for range header
	rangeHeader := c.Get("Range")
	var rangeStart, rangeEnd int64 = 0, -1
	hasRange := false

	if rangeHeader != "" {
		start, end, err := parseRangeHeader(rangeHeader)
		if err == nil {
			rangeStart = start
			rangeEnd = end
			hasRange = true
		}
	}

	// Call download service
	result, err := h.downloadService.PrepareDownload(c.Context(), application.PrepareDownloadInput{
		UserID:      userID,
		MaterialID:  materialID,
		DeviceID:    deviceID,
		LicenseID:   licenseID,
		Fingerprint: fingerprint,
	})
	if err != nil {
		return sendError(c, requestID, err)
	}

	// Build response
	downloadResp := DownloadResponse{
		Manifest:    *ToManifestResponse(result.Manifest),
		DownloadURL: result.DownloadURL,
		ExpiresAt:   result.ExpiresAt,
	}

	// Handle range request
	if hasRange {
		// Add range-specific headers
		totalSize := result.Manifest.TotalSize
		if rangeEnd < 0 || rangeEnd >= totalSize {
			rangeEnd = totalSize - 1
		}
		contentLength := rangeEnd - rangeStart + 1

		c.Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", rangeStart, rangeEnd, totalSize))
		c.Set("Accept-Ranges", "bytes")
		c.Set("Content-Length", strconv.FormatInt(contentLength, 10))

		return c.Status(fiber.StatusPartialContent).JSON(response.OK(requestID, downloadResp))
	}

	// Add accept-ranges header for full response
	c.Set("Accept-Ranges", "bytes")

	return c.Status(fiber.StatusOK).JSON(response.OK(requestID, downloadResp))
}

// parseRangeHeader parses an HTTP Range header.
// Supports format: "bytes=start-end" or "bytes=start-"
func parseRangeHeader(header string) (start, end int64, err error) {
	if !strings.HasPrefix(header, "bytes=") {
		return 0, 0, fmt.Errorf("invalid range header format")
	}

	rangeSpec := strings.TrimPrefix(header, "bytes=")
	parts := strings.Split(rangeSpec, "-")

	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("invalid range header format")
	}

	// Parse start
	if parts[0] == "" {
		return 0, 0, fmt.Errorf("range start is required")
	}
	start, err = strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return 0, 0, fmt.Errorf("invalid range start: %w", err)
	}

	// Parse end (optional)
	if parts[1] == "" {
		end = -1 // Indicates "to end of file"
	} else {
		end, err = strconv.ParseInt(parts[1], 10, 64)
		if err != nil {
			return 0, 0, fmt.Errorf("invalid range end: %w", err)
		}
	}

	if end >= 0 && start > end {
		return 0, 0, fmt.Errorf("range start cannot be greater than end")
	}

	return start, end, nil
}

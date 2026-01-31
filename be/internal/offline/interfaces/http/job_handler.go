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

// JobHandler handles job-related HTTP requests.
type JobHandler struct {
	jobService *application.JobService
}

// NewJobHandler creates a new JobHandler.
func NewJobHandler(jobService *application.JobService) *JobHandler {
	return &JobHandler{
		jobService: jobService,
	}
}

// JobResponse represents a job in API responses.
type JobResponse struct {
	ID          uuid.UUID  `json:"id"`
	MaterialID  uuid.UUID  `json:"material_id"`
	UserID      uuid.UUID  `json:"user_id"`
	DeviceID    uuid.UUID  `json:"device_id"`
	LicenseID   uuid.UUID  `json:"license_id"`
	Priority    int        `json:"priority"`
	Status      string     `json:"status"`
	Error       *string    `json:"error,omitempty"`
	RetryCount  int        `json:"retry_count"`
	CreatedAt   time.Time  `json:"created_at"`
	StartedAt   *time.Time `json:"started_at,omitempty"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
}

// ToJobResponse converts a domain.EncryptionJob to JobResponse.
func ToJobResponse(job *domain.EncryptionJob) *JobResponse {
	if job == nil {
		return nil
	}

	return &JobResponse{
		ID:          job.ID,
		MaterialID:  job.MaterialID,
		UserID:      job.UserID,
		DeviceID:    job.DeviceID,
		LicenseID:   job.LicenseID,
		Priority:    job.Priority,
		Status:      string(job.Status),
		Error:       job.Error,
		RetryCount:  job.RetryCount,
		CreatedAt:   job.CreatedAt,
		StartedAt:   job.StartedAt,
		CompletedAt: job.CompletedAt,
	}
}

// SyncRequest represents the request body for sync.
type SyncRequest struct {
	DeviceID       string `json:"device_id" validate:"required,uuid"`
	Fingerprint    string `json:"fingerprint" validate:"required,min=32,max=512"`
	LastSyncAt     string `json:"last_sync_at,omitempty"`
	PendingChanges []byte `json:"pending_changes,omitempty"`
}

// SyncResponse represents the sync response.
type SyncResponse struct {
	DeviceID       uuid.UUID `json:"device_id"`
	LastSyncAt     time.Time `json:"last_sync_at"`
	SyncVersion    int       `json:"sync_version"`
	PendingChanges []byte    `json:"pending_changes,omitempty"`
	Licenses       []LicenseSyncInfo `json:"licenses,omitempty"`
}

// LicenseSyncInfo represents license info for sync.
type LicenseSyncInfo struct {
	LicenseID  uuid.UUID `json:"license_id"`
	MaterialID uuid.UUID `json:"material_id"`
	Status     string    `json:"status"`
	ExpiresAt  time.Time `json:"expires_at"`
}

// GetJob handles getting a job by ID.
// @Summary Get job status
// @Description Get the status of an encryption job by ID
// @Tags Offline
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param job_id path string true "Job ID" format(uuid)
// @Success 200 {object} response.Response[JobResponse] "Job details"
// @Failure 400 {object} errors.ErrorResponse "Invalid job ID"
// @Failure 401 {object} errors.ErrorResponse "Authentication required"
// @Failure 404 {object} errors.ErrorResponse "Job not found"
// @Router /offline/jobs/{job_id} [get]
func (h *JobHandler) GetJob(c *fiber.Ctx) error {
	requestID := middleware.GetRequestID(c)

	// Get user ID from context
	_, ok := middleware.GetUserID(c)
	if !ok {
		return sendError(c, requestID, errors.Unauthorized(""))
	}

	// Parse job ID from path
	jobIDStr := c.Params("job_id")
	jobID, err := uuid.Parse(jobIDStr)
	if err != nil {
		return sendError(c, requestID, errors.BadRequest("invalid job ID"))
	}

	// Get job
	job, err := h.jobService.GetJob(c.Context(), jobID)
	if err != nil {
		return sendError(c, requestID, err)
	}

	return c.Status(fiber.StatusOK).JSON(response.OK(requestID, ToJobResponse(job)))
}

// GetJobByMaterial handles getting a job by material ID.
// @Summary Get job for material
// @Description Get the latest encryption job for a material
// @Tags Offline
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param material_id path string true "Material ID" format(uuid)
// @Success 200 {object} response.Response[JobResponse] "Job details"
// @Failure 400 {object} errors.ErrorResponse "Invalid material ID"
// @Failure 401 {object} errors.ErrorResponse "Authentication required"
// @Failure 404 {object} errors.ErrorResponse "No job found for material"
// @Router /offline/materials/{material_id}/job [get]
func (h *JobHandler) GetJobByMaterial(c *fiber.Ctx) error {
	requestID := middleware.GetRequestID(c)

	// Get user ID from context
	_, ok := middleware.GetUserID(c)
	if !ok {
		return sendError(c, requestID, errors.Unauthorized(""))
	}

	// Parse material ID from path
	materialIDStr := c.Params("material_id")
	materialID, err := uuid.Parse(materialIDStr)
	if err != nil {
		return sendError(c, requestID, errors.BadRequest("invalid material ID"))
	}

	// Get job by material
	job, err := h.jobService.GetJobByMaterial(c.Context(), materialID)
	if err != nil {
		return sendError(c, requestID, err)
	}

	return c.Status(fiber.StatusOK).JSON(response.OK(requestID, ToJobResponse(job)))
}

// Sync handles sync requests from devices.
// @Summary Sync offline state
// @Description Sync offline state between device and server
// @Tags Offline
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body SyncRequest true "Sync data"
// @Success 200 {object} response.Response[SyncResponse] "Sync result"
// @Failure 400 {object} errors.ErrorResponse "Invalid request"
// @Failure 401 {object} errors.ErrorResponse "Authentication required"
// @Failure 403 {object} errors.ErrorResponse "Device validation failed"
// @Router /offline/sync [post]
func (h *JobHandler) Sync(c *fiber.Ctx) error {
	requestID := middleware.GetRequestID(c)

	// Get user ID from context
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return sendError(c, requestID, errors.Unauthorized(""))
	}

	var req SyncRequest
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

	// For now, return a basic sync response
	// Full sync implementation would involve:
	// 1. Validating device
	// 2. Checking for license updates
	// 3. Returning pending changes
	syncResp := SyncResponse{
		DeviceID:    deviceID,
		LastSyncAt:  time.Now(),
		SyncVersion: 1,
		Licenses:    []LicenseSyncInfo{},
	}

	// Log sync request
	_ = userID // Will be used in full implementation

	return c.Status(fiber.StatusOK).JSON(response.OK(requestID, syncResp))
}

// Package http provides HTTP handlers for the Offline Material Service API.
// Implements the interfaces layer in Clean Architecture.
package http

import (
	"github.com/gofiber/fiber/v2"

	"ngasihtau/internal/common/middleware"
	"ngasihtau/internal/offline/application"
	"ngasihtau/pkg/jwt"
)

// Handler contains all HTTP handlers for the Offline Material Service.
type Handler struct {
	deviceHandler   *DeviceHandler
	licenseHandler  *LicenseHandler
	downloadHandler *DownloadHandler
	jobHandler      *JobHandler
	jwtManager      *jwt.Manager
}

// NewHandler creates a new Handler with the given dependencies.
func NewHandler(
	deviceService application.DeviceService,
	licenseService application.LicenseService,
	downloadService application.DownloadService,
	jobService *application.JobService,
	encryptionService *application.EncryptionService,
	jwtManager *jwt.Manager,
) *Handler {
	return &Handler{
		deviceHandler:   NewDeviceHandler(deviceService),
		licenseHandler:  NewLicenseHandler(licenseService),
		downloadHandler: NewDownloadHandler(downloadService, licenseService, encryptionService),
		jobHandler:      NewJobHandler(jobService),
		jwtManager:      jwtManager,
	}
}

// NewHandlerWithDeviceOnly creates a new Handler with only device service.
func NewHandlerWithDeviceOnly(
	deviceService application.DeviceService,
	jwtManager *jwt.Manager,
) *Handler {
	return &Handler{
		deviceHandler: NewDeviceHandler(deviceService),
		jwtManager:    jwtManager,
	}
}

// NewHandlerWithDeviceAndLicense creates a new Handler with device and license services.
func NewHandlerWithDeviceAndLicense(
	deviceService application.DeviceService,
	licenseService application.LicenseService,
	jwtManager *jwt.Manager,
) *Handler {
	return &Handler{
		deviceHandler:  NewDeviceHandler(deviceService),
		licenseHandler: NewLicenseHandler(licenseService),
		jwtManager:     jwtManager,
	}
}

// RegisterRoutes registers all Offline Material Service routes on the given Fiber app.
// Routes are organized by functionality:
//
// Device Management (protected):
//   - POST   /api/v1/offline/devices (register device)
//   - GET    /api/v1/offline/devices (list devices)
//   - DELETE /api/v1/offline/devices/:device_id (deregister device)
//
// License Management (protected):
//   - POST   /api/v1/offline/materials/:material_id/license (issue license)
//   - POST   /api/v1/offline/licenses/:license_id/validate (validate license)
//   - POST   /api/v1/offline/licenses/:license_id/renew (renew license)
//
// Download (protected):
//   - GET    /api/v1/offline/materials/:material_id/download (download encrypted material)
//
// Job Management (protected):
//   - GET    /api/v1/offline/jobs/:job_id (get job status)
//   - GET    /api/v1/offline/materials/:material_id/job (get job for material)
//
// Sync (protected):
//   - POST   /api/v1/offline/sync (sync offline state)
func (h *Handler) RegisterRoutes(app *fiber.App) {
	// API v1 group
	api := app.Group("/api/v1")

	// Offline routes group
	offline := api.Group("/offline")

	// Device routes (all protected)
	devices := offline.Group("/devices", middleware.Auth(h.jwtManager))
	devices.Post("", h.deviceHandler.RegisterDevice)
	devices.Get("", h.deviceHandler.ListDevices)
	devices.Delete("/:device_id", h.deviceHandler.DeregisterDevice)

	// License routes (all protected)
	if h.licenseHandler != nil {
		// Issue license for a material
		offline.Post("/materials/:material_id/license",
			middleware.Auth(h.jwtManager),
			h.licenseHandler.IssueLicense,
		)

		// License operations
		licenses := offline.Group("/licenses", middleware.Auth(h.jwtManager))
		licenses.Post("/:license_id/validate", h.licenseHandler.ValidateLicense)
		licenses.Post("/:license_id/renew", h.licenseHandler.RenewLicense)
	}

	// Download routes (protected)
	if h.downloadHandler != nil {
		offline.Get("/materials/:material_id/download",
			middleware.Auth(h.jwtManager),
			h.downloadHandler.DownloadMaterial,
		)
	}

	// Job routes (protected)
	if h.jobHandler != nil {
		jobs := offline.Group("/jobs", middleware.Auth(h.jwtManager))
		jobs.Get("/:job_id", h.jobHandler.GetJob)

		// Material job route
		offline.Get("/materials/:material_id/job",
			middleware.Auth(h.jwtManager),
			h.jobHandler.GetJobByMaterial,
		)

		// Sync route
		offline.Post("/sync",
			middleware.Auth(h.jwtManager),
			h.jobHandler.Sync,
		)
	}
}

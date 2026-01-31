// Package http provides HTTP handlers for the User Service API.
// This file contains handlers for teacher verification operations.
package http

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"ngasihtau/internal/common/errors"
	"ngasihtau/internal/common/middleware"
	"ngasihtau/internal/common/response"
	"ngasihtau/internal/common/validator"
	"ngasihtau/internal/user/application"
	"ngasihtau/internal/user/domain"

	_ "ngasihtau/docs" // Swagger docs
)

// VerificationHandler handles teacher verification-related HTTP requests.
// Implements requirements 2.1, 2.2 for teacher verification.
type VerificationHandler struct {
	userService application.UserService
}

// NewVerificationHandler creates a new VerificationHandler.
func NewVerificationHandler(userService application.UserService) *VerificationHandler {
	return &VerificationHandler{
		userService: userService,
	}
}

// SubmitVerificationRequest represents the request body for teacher verification submission.
// @Description Teacher verification submission request
type SubmitVerificationRequest struct {
	FullName       string `json:"full_name" validate:"required,min=3,max=255" example:"Dr. John Smith"`
	IDNumber       string `json:"id_number" validate:"required,min=10,max=100" example:"3275012345678901"`
	CredentialType string `json:"credential_type" validate:"required,oneof=government_id educator_card professional_cert" example:"educator_card"`
	DocumentRef    string `json:"document_ref" validate:"required,max=500" example:"doc_ref_abc123xyz"`
}

// RejectVerificationRequest represents the request body for rejecting a verification.
// @Description Reject verification request
type RejectVerificationRequest struct {
	Reason string `json:"reason" validate:"required,min=1,max=1000" example:"The submitted document is not clear. Please resubmit with a higher quality image."`
}

// TeacherVerificationResponse represents a teacher verification in API responses.
// @Description Teacher verification information
type TeacherVerificationResponse struct {
	ID              uuid.UUID  `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	UserID          uuid.UUID  `json:"user_id" example:"550e8400-e29b-41d4-a716-446655440001"`
	FullName        string     `json:"full_name" example:"Dr. John Smith"`
	IDNumber        string     `json:"id_number" example:"3275012345678901"`
	CredentialType  string     `json:"credential_type" example:"educator_card"`
	DocumentRef     string     `json:"document_ref" example:"doc_ref_abc123xyz"`
	Status          string     `json:"status" example:"pending"`
	ReviewedBy      *uuid.UUID `json:"reviewed_by,omitempty" example:"550e8400-e29b-41d4-a716-446655440002"`
	ReviewedAt      *time.Time `json:"reviewed_at,omitempty"`
	RejectionReason *string    `json:"rejection_reason,omitempty" example:"Document not clear"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

// ToTeacherVerificationResponse converts a domain.TeacherVerification to TeacherVerificationResponse.
func ToTeacherVerificationResponse(v *domain.TeacherVerification) *TeacherVerificationResponse {
	if v == nil {
		return nil
	}

	return &TeacherVerificationResponse{
		ID:              v.ID,
		UserID:          v.UserID,
		FullName:        v.FullName,
		IDNumber:        v.IDNumber,
		CredentialType:  string(v.CredentialType),
		DocumentRef:     v.DocumentRef,
		Status:          string(v.Status),
		ReviewedBy:      v.ReviewedBy,
		ReviewedAt:      v.ReviewedAt,
		RejectionReason: v.RejectionReason,
		CreatedAt:       v.CreatedAt,
		UpdatedAt:       v.UpdatedAt,
	}
}

// toVerificationResponseList converts a slice of domain.TeacherVerification to response slice.
func toVerificationResponseList(verifications []*domain.TeacherVerification) []*TeacherVerificationResponse {
	result := make([]*TeacherVerificationResponse, len(verifications))
	for i, v := range verifications {
		result[i] = ToTeacherVerificationResponse(v)
	}
	return result
}

// SubmitTeacherVerification handles teacher verification submission.
// @Summary Submit teacher verification request
// @Description Submit a request to become a verified teacher with identity documents
// @Tags Verification
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body SubmitVerificationRequest true "Verification details"
// @Success 201 {object} response.Response[TeacherVerificationResponse] "Verification request created"
// @Failure 400 {object} errors.ErrorResponse "Invalid request body or validation error"
// @Failure 401 {object} errors.ErrorResponse "Authentication required"
// @Failure 409 {object} errors.ErrorResponse "Verification request already exists or user is already a teacher"
// @Router /users/verification/teacher [post]
func (h *VerificationHandler) SubmitTeacherVerification(c *fiber.Ctx) error {
	requestID := middleware.GetRequestID(c)

	// Get user ID from context (set by auth middleware)
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return sendError(c, requestID, errors.Unauthorized(""))
	}

	var req SubmitVerificationRequest
	if err := c.BodyParser(&req); err != nil {
		return sendError(c, requestID, errors.BadRequest("invalid request body"))
	}

	// Validate request
	if err := validator.Get().Struct(&req); err != nil {
		return sendError(c, requestID, err)
	}

	// Call service
	verification, err := h.userService.SubmitTeacherVerification(c.Context(), userID, application.TeacherVerificationInput{
		FullName:       req.FullName,
		IDNumber:       req.IDNumber,
		CredentialType: domain.CredentialType(req.CredentialType),
		DocumentRef:    req.DocumentRef,
	})
	if err != nil {
		return sendError(c, requestID, err)
	}

	return c.Status(fiber.StatusCreated).JSON(response.Created(requestID, ToTeacherVerificationResponse(verification)))
}

// GetVerificationStatus handles getting the current user's verification status.
// @Summary Get verification status
// @Description Get the current user's teacher verification status
// @Tags Verification
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response[TeacherVerificationResponse] "Verification status"
// @Failure 401 {object} errors.ErrorResponse "Authentication required"
// @Failure 404 {object} errors.ErrorResponse "No verification request found"
// @Router /users/verification/status [get]
func (h *VerificationHandler) GetVerificationStatus(c *fiber.Ctx) error {
	requestID := middleware.GetRequestID(c)

	// Get user ID from context (set by auth middleware)
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return sendError(c, requestID, errors.Unauthorized(""))
	}

	// Call service
	verification, err := h.userService.GetVerificationStatus(c.Context(), userID)
	if err != nil {
		return sendError(c, requestID, err)
	}

	return c.Status(fiber.StatusOK).JSON(response.OK(requestID, ToTeacherVerificationResponse(verification)))
}

// GetPendingVerifications handles getting pending verification requests (admin only).
// @Summary Get pending verifications
// @Description Get a paginated list of pending teacher verification requests (admin only)
// @Tags Verification
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param per_page query int false "Items per page" default(20)
// @Success 200 {object} response.PaginatedResponse[TeacherVerificationResponse] "List of pending verifications"
// @Failure 401 {object} errors.ErrorResponse "Authentication required"
// @Failure 403 {object} errors.ErrorResponse "Admin access required"
// @Router /admin/verifications [get]
func (h *VerificationHandler) GetPendingVerifications(c *fiber.Ctx) error {
	requestID := middleware.GetRequestID(c)

	// Get pagination parameters
	page := c.QueryInt("page", 1)
	perPage := c.QueryInt("per_page", 20)

	// Call service
	result, err := h.userService.GetPendingVerifications(c.Context(), page, perPage)
	if err != nil {
		return sendError(c, requestID, err)
	}

	return c.Status(fiber.StatusOK).JSON(response.List(
		requestID,
		toVerificationResponseList(result.Verifications),
		result.Page,
		result.PerPage,
		result.Total,
	))
}

// ApproveVerification handles approving a teacher verification request (admin only).
// @Summary Approve verification
// @Description Approve a teacher verification request and update user role to teacher (admin only)
// @Tags Verification
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Verification ID" format(uuid)
// @Success 200 {object} response.Response[any] "Verification approved successfully"
// @Failure 400 {object} errors.ErrorResponse "Invalid verification ID or already reviewed"
// @Failure 401 {object} errors.ErrorResponse "Authentication required"
// @Failure 403 {object} errors.ErrorResponse "Admin access required"
// @Failure 404 {object} errors.ErrorResponse "Verification not found"
// @Router /admin/verifications/{id}/approve [post]
func (h *VerificationHandler) ApproveVerification(c *fiber.Ctx) error {
	requestID := middleware.GetRequestID(c)

	// Get reviewer ID from context (set by auth middleware)
	reviewerID, ok := middleware.GetUserID(c)
	if !ok {
		return sendError(c, requestID, errors.Unauthorized(""))
	}

	// Parse verification ID from path
	verificationIDStr := c.Params("id")
	verificationID, err := uuid.Parse(verificationIDStr)
	if err != nil {
		return sendError(c, requestID, errors.BadRequest("invalid verification ID"))
	}

	// Call service
	if err := h.userService.ApproveVerification(c.Context(), verificationID, reviewerID); err != nil {
		return sendError(c, requestID, err)
	}

	return c.Status(fiber.StatusOK).JSON(response.OK(requestID, map[string]bool{
		"approved": true,
	}))
}

// RejectVerification handles rejecting a teacher verification request (admin only).
// @Summary Reject verification
// @Description Reject a teacher verification request with a reason (admin only)
// @Tags Verification
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Verification ID" format(uuid)
// @Param request body RejectVerificationRequest true "Rejection reason"
// @Success 200 {object} response.Response[any] "Verification rejected successfully"
// @Failure 400 {object} errors.ErrorResponse "Invalid verification ID, already reviewed, or missing reason"
// @Failure 401 {object} errors.ErrorResponse "Authentication required"
// @Failure 403 {object} errors.ErrorResponse "Admin access required"
// @Failure 404 {object} errors.ErrorResponse "Verification not found"
// @Router /admin/verifications/{id}/reject [post]
func (h *VerificationHandler) RejectVerification(c *fiber.Ctx) error {
	requestID := middleware.GetRequestID(c)

	// Get reviewer ID from context (set by auth middleware)
	reviewerID, ok := middleware.GetUserID(c)
	if !ok {
		return sendError(c, requestID, errors.Unauthorized(""))
	}

	// Parse verification ID from path
	verificationIDStr := c.Params("id")
	verificationID, err := uuid.Parse(verificationIDStr)
	if err != nil {
		return sendError(c, requestID, errors.BadRequest("invalid verification ID"))
	}

	var req RejectVerificationRequest
	if err := c.BodyParser(&req); err != nil {
		return sendError(c, requestID, errors.BadRequest("invalid request body"))
	}

	// Validate request
	if err := validator.Get().Struct(&req); err != nil {
		return sendError(c, requestID, err)
	}

	// Call service
	if err := h.userService.RejectVerification(c.Context(), verificationID, reviewerID, req.Reason); err != nil {
		return sendError(c, requestID, err)
	}

	return c.Status(fiber.StatusOK).JSON(response.OK(requestID, map[string]bool{
		"rejected": true,
	}))
}

package http

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"ngasihtau/internal/common/errors"
	"ngasihtau/internal/common/middleware"
	"ngasihtau/internal/common/response"
	"ngasihtau/internal/common/validator"
	"ngasihtau/internal/user/application"
)

// InterestHandler handles learning interest HTTP requests.
type InterestHandler struct {
	interestService application.LearningInterestService
}

// NewInterestHandler creates a new InterestHandler.
func NewInterestHandler(interestService application.LearningInterestService) *InterestHandler {
	return &InterestHandler{
		interestService: interestService,
	}
}

// SetInterestsRequest represents the request body for setting user interests.
type SetInterestsRequest struct {
	PredefinedInterestIDs []string `json:"predefined_interest_ids" validate:"omitempty,dive,uuid"`
	CustomInterests       []string `json:"custom_interests" validate:"omitempty,dive,min=2,max=100"`
}

// AddInterestRequest represents the request body for adding a single interest.
type AddInterestRequest struct {
	PredefinedInterestID *string `json:"predefined_interest_id,omitempty" validate:"omitempty,uuid"`
	CustomInterest       *string `json:"custom_interest,omitempty" validate:"omitempty,min=2,max=100"`
}

// GetPredefinedInterests godoc
// @Summary Get all predefined learning interests
// @Description Returns a list of all active predefined learning interests that users can select from
// @Tags Interests
// @Accept json
// @Produce json
// @Success 200 {object} response.Response[application.PredefinedInterestsResult]
// @Failure 500 {object} errors.ErrorResponse
// @Router /api/v1/interests/predefined [get]
func (h *InterestHandler) GetPredefinedInterests(c *fiber.Ctx) error {
	requestID := middleware.GetRequestID(c)

	result, err := h.interestService.GetPredefinedInterests(c.Context())
	if err != nil {
		return sendInterestError(c, requestID, err)
	}

	return c.Status(fiber.StatusOK).JSON(response.OK(requestID, result))
}

// GetPredefinedInterestsByCategory godoc
// @Summary Get predefined interests grouped by category
// @Description Returns predefined learning interests organized by their categories
// @Tags Interests
// @Accept json
// @Produce json
// @Success 200 {object} response.Response[application.GroupedInterestsResult]
// @Failure 500 {object} errors.ErrorResponse
// @Router /api/v1/interests/predefined/categories [get]
func (h *InterestHandler) GetPredefinedInterestsByCategory(c *fiber.Ctx) error {
	requestID := middleware.GetRequestID(c)

	result, err := h.interestService.GetPredefinedInterestsByCategory(c.Context())
	if err != nil {
		return sendInterestError(c, requestID, err)
	}

	return c.Status(fiber.StatusOK).JSON(response.OK(requestID, result))
}

// GetUserInterests godoc
// @Summary Get current user's learning interests
// @Description Returns all learning interests selected by the authenticated user
// @Tags Interests
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response[application.UserInterestsResult]
// @Failure 401 {object} errors.ErrorResponse
// @Failure 500 {object} errors.ErrorResponse
// @Router /api/v1/interests/me [get]
func (h *InterestHandler) GetUserInterests(c *fiber.Ctx) error {
	requestID := middleware.GetRequestID(c)

	userID, ok := middleware.GetUserID(c)
	if !ok {
		return sendInterestError(c, requestID, errors.Unauthorized("authentication required"))
	}

	result, err := h.interestService.GetUserInterests(c.Context(), userID)
	if err != nil {
		return sendInterestError(c, requestID, err)
	}

	return c.Status(fiber.StatusOK).JSON(response.OK(requestID, result))
}

// SetUserInterests godoc
// @Summary Set user's learning interests (onboarding)
// @Description Sets or replaces all learning interests for the authenticated user. Used during onboarding.
// @Tags Interests
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body SetInterestsRequest true "Interests to set"
// @Success 200 {object} response.Response[any]
// @Failure 400 {object} errors.ErrorResponse
// @Failure 401 {object} errors.ErrorResponse
// @Failure 500 {object} errors.ErrorResponse
// @Router /api/v1/interests/me [put]
func (h *InterestHandler) SetUserInterests(c *fiber.Ctx) error {
	requestID := middleware.GetRequestID(c)

	userID, ok := middleware.GetUserID(c)
	if !ok {
		return sendInterestError(c, requestID, errors.Unauthorized("authentication required"))
	}

	var req SetInterestsRequest
	if err := c.BodyParser(&req); err != nil {
		return sendInterestError(c, requestID, errors.BadRequest("invalid request body"))
	}

	if err := validator.Get().Struct(&req); err != nil {
		return sendInterestError(c, requestID, err)
	}

	// Convert string UUIDs to uuid.UUID
	predefinedIDs := make([]uuid.UUID, 0, len(req.PredefinedInterestIDs))
	for _, idStr := range req.PredefinedInterestIDs {
		id, parseErr := uuid.Parse(idStr)
		if parseErr != nil {
			return sendInterestError(c, requestID, errors.BadRequest("invalid predefined interest ID: "+idStr))
		}
		predefinedIDs = append(predefinedIDs, id)
	}

	input := application.SetInterestsInput{
		PredefinedInterestIDs: predefinedIDs,
		CustomInterests:       req.CustomInterests,
	}

	if err := h.interestService.SetUserInterests(c.Context(), userID, input); err != nil {
		return sendInterestError(c, requestID, err)
	}

	return c.Status(fiber.StatusOK).JSON(response.Empty(requestID))
}

// AddUserInterest godoc
// @Summary Add a single interest to user's list
// @Description Adds a predefined or custom interest to the authenticated user's interest list
// @Tags Interests
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body AddInterestRequest true "Interest to add"
// @Success 201 {object} response.Response[domain.InterestSummary]
// @Failure 400 {object} errors.ErrorResponse
// @Failure 401 {object} errors.ErrorResponse
// @Failure 409 {object} errors.ErrorResponse
// @Failure 500 {object} errors.ErrorResponse
// @Router /api/v1/interests/me [post]
func (h *InterestHandler) AddUserInterest(c *fiber.Ctx) error {
	requestID := middleware.GetRequestID(c)

	userID, ok := middleware.GetUserID(c)
	if !ok {
		return sendInterestError(c, requestID, errors.Unauthorized("authentication required"))
	}

	var req AddInterestRequest
	if err := c.BodyParser(&req); err != nil {
		return sendInterestError(c, requestID, errors.BadRequest("invalid request body"))
	}

	if err := validator.Get().Struct(&req); err != nil {
		return sendInterestError(c, requestID, err)
	}

	// Convert to service input
	var input application.AddInterestInput
	if req.PredefinedInterestID != nil {
		id, parseErr := uuid.Parse(*req.PredefinedInterestID)
		if parseErr != nil {
			return sendInterestError(c, requestID, errors.BadRequest("invalid predefined interest ID"))
		}
		input.PredefinedInterestID = &id
	}
	input.CustomInterest = req.CustomInterest

	summary, err := h.interestService.AddUserInterest(c.Context(), userID, input)
	if err != nil {
		return sendInterestError(c, requestID, err)
	}

	return c.Status(fiber.StatusCreated).JSON(response.Created(requestID, summary))
}

// RemoveUserInterest godoc
// @Summary Remove an interest from user's list
// @Description Removes a specific interest from the authenticated user's interest list
// @Tags Interests
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Interest ID"
// @Success 200 {object} response.Response[any]
// @Failure 401 {object} errors.ErrorResponse
// @Failure 404 {object} errors.ErrorResponse
// @Failure 500 {object} errors.ErrorResponse
// @Router /api/v1/interests/me/{id} [delete]
func (h *InterestHandler) RemoveUserInterest(c *fiber.Ctx) error {
	requestID := middleware.GetRequestID(c)

	userID, ok := middleware.GetUserID(c)
	if !ok {
		return sendInterestError(c, requestID, errors.Unauthorized("authentication required"))
	}

	interestIDStr := c.Params("id")
	interestID, parseErr := uuid.Parse(interestIDStr)
	if parseErr != nil {
		return sendInterestError(c, requestID, errors.BadRequest("invalid interest ID"))
	}

	if err := h.interestService.RemoveUserInterest(c.Context(), userID, interestID); err != nil {
		return sendInterestError(c, requestID, err)
	}

	return c.Status(fiber.StatusOK).JSON(response.Empty(requestID))
}

// CompleteOnboarding godoc
// @Summary Complete user onboarding
// @Description Marks the user's onboarding as complete. User must have selected at least one interest.
// @Tags Interests
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response[any]
// @Failure 400 {object} errors.ErrorResponse
// @Failure 401 {object} errors.ErrorResponse
// @Failure 500 {object} errors.ErrorResponse
// @Router /api/v1/interests/onboarding/complete [post]
func (h *InterestHandler) CompleteOnboarding(c *fiber.Ctx) error {
	requestID := middleware.GetRequestID(c)

	userID, ok := middleware.GetUserID(c)
	if !ok {
		return sendInterestError(c, requestID, errors.Unauthorized("authentication required"))
	}

	if err := h.interestService.CompleteOnboarding(c.Context(), userID); err != nil {
		return sendInterestError(c, requestID, err)
	}

	return c.Status(fiber.StatusOK).JSON(response.Empty(requestID))
}

// GetOnboardingStatus godoc
// @Summary Get user onboarding status
// @Description Returns whether the user has completed onboarding and their current interests
// @Tags Interests
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response[application.OnboardingStatus]
// @Failure 401 {object} errors.ErrorResponse
// @Failure 500 {object} errors.ErrorResponse
// @Router /api/v1/interests/onboarding/status [get]
func (h *InterestHandler) GetOnboardingStatus(c *fiber.Ctx) error {
	requestID := middleware.GetRequestID(c)

	userID, ok := middleware.GetUserID(c)
	if !ok {
		return sendInterestError(c, requestID, errors.Unauthorized("authentication required"))
	}

	status, err := h.interestService.CheckOnboardingStatus(c.Context(), userID)
	if err != nil {
		return sendInterestError(c, requestID, err)
	}

	return c.Status(fiber.StatusOK).JSON(response.OK(requestID, status))
}

// sendInterestError sends an error response using the standard error format.
func sendInterestError(c *fiber.Ctx, requestID string, err error) error {
	resp := errors.BuildResponse(requestID, err)
	status := errors.GetHTTPStatus(err)
	return c.Status(status).JSON(resp)
}

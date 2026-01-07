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

// UserHandler handles user profile and follow-related HTTP requests.
type UserHandler struct {
	userService    application.UserService
	storageService application.StorageService
	aiService      application.AIService
}

// NewUserHandler creates a new UserHandler.
func NewUserHandler(userService application.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

// NewUserHandlerWithStorage creates a new UserHandler with storage service.
func NewUserHandlerWithStorage(userService application.UserService, storageService application.StorageService) *UserHandler {
	return &UserHandler{
		userService:    userService,
		storageService: storageService,
	}
}

// NewUserHandlerWithServices creates a new UserHandler with storage and AI services.
func NewUserHandlerWithServices(userService application.UserService, storageService application.StorageService, aiService application.AIService) *UserHandler {
	return &UserHandler{
		userService:    userService,
		storageService: storageService,
		aiService:      aiService,
	}
}

// UserResponse represents a user in API responses.
type UserResponse struct {
	ID                  uuid.UUID `json:"id"`
	Email               string    `json:"email,omitempty"` // Only included for own profile
	Name                string    `json:"name"`
	AvatarURL           *string   `json:"avatar_url,omitempty"`
	Bio                 *string   `json:"bio,omitempty"`
	Role                string    `json:"role"`
	EmailVerified       bool      `json:"email_verified,omitempty"`       // Only included for own profile
	TwoFactorEnabled    bool      `json:"two_factor_enabled,omitempty"`   // Only included for own profile
	Language            string    `json:"language,omitempty"`             // Only included for own profile
	OnboardingCompleted bool      `json:"onboarding_completed,omitempty"` // Only included for own profile
	CreatedAt           time.Time `json:"created_at"`
}

// UserProfileResponse represents a user's public profile in API responses.
type UserProfileResponse struct {
	ID             uuid.UUID `json:"id"`
	Name           string    `json:"name"`
	AvatarURL      *string   `json:"avatar_url,omitempty"`
	Bio            *string   `json:"bio,omitempty"`
	Role           string    `json:"role"`
	FollowerCount  int       `json:"follower_count"`
	FollowingCount int       `json:"following_count"`
	PodCount       int       `json:"pod_count"`
	MaterialCount  int       `json:"material_count"`
	CreatedAt      time.Time `json:"created_at"`
	IsFollowing    bool      `json:"is_following,omitempty"` // Only when viewing another user's profile
}

// UpdateProfileRequest represents the request body for profile update.
type UpdateProfileRequest struct {
	Name      *string `json:"name,omitempty" validate:"omitempty,min=2,max=100"`
	Bio       *string `json:"bio,omitempty" validate:"omitempty,max=500"`
	AvatarURL *string `json:"avatar_url,omitempty" validate:"omitempty,url"`
	Language  *string `json:"language,omitempty" validate:"omitempty,oneof=id en"`
}

// FollowListResponse represents a paginated list of users.
type FollowListResponse struct {
	Users      []*UserResponse `json:"users"`
	Total      int             `json:"total"`
	Page       int             `json:"page"`
	PerPage    int             `json:"per_page"`
	TotalPages int             `json:"total_pages"`
}

// ToUserResponse converts a domain.User to UserResponse.
func ToUserResponse(user *domain.User, includePrivate bool) *UserResponse {
	if user == nil {
		return nil
	}

	resp := &UserResponse{
		ID:        user.ID,
		Name:      user.Name,
		AvatarURL: user.AvatarURL,
		Bio:       user.Bio,
		Role:      string(user.Role),
		CreatedAt: user.CreatedAt,
	}

	if includePrivate {
		resp.Email = user.Email
		resp.EmailVerified = user.EmailVerified
		resp.TwoFactorEnabled = user.TwoFactorEnabled
		resp.Language = user.Language
		resp.OnboardingCompleted = user.OnboardingCompleted
	}

	return resp
}

// ToUserProfileResponse converts a domain.UserProfile to UserProfileResponse.
func ToUserProfileResponse(profile *domain.UserProfile) *UserProfileResponse {
	if profile == nil {
		return nil
	}

	return &UserProfileResponse{
		ID:             profile.ID,
		Name:           profile.Name,
		AvatarURL:      profile.AvatarURL,
		Bio:            profile.Bio,
		Role:           string(profile.Role),
		FollowerCount:  profile.FollowerCount,
		FollowingCount: profile.FollowingCount,
		PodCount:       profile.PodCount,
		MaterialCount:  profile.MaterialCount,
		CreatedAt:      profile.CreatedAt,
	}
}

// toUserResponseList converts a slice of domain.User to UserResponse slice.
func toUserResponseList(users []*domain.User) []*UserResponse {
	result := make([]*UserResponse, len(users))
	for i, user := range users {
		result[i] = ToUserResponse(user, false)
	}
	return result
}

// GetCurrentUser handles getting the current user's profile.
// @Summary Get current user profile
// @Description Get the authenticated user's full profile including private fields
// @Tags Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response[UserResponse] "User profile"
// @Failure 401 {object} errors.ErrorResponse "Authentication required"
// @Router /users/me [get]
func (h *UserHandler) GetCurrentUser(c *fiber.Ctx) error {
	requestID := middleware.GetRequestID(c)

	// Get user ID from context (set by auth middleware)
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return sendError(c, requestID, errors.Unauthorized(""))
	}

	// Call service
	user, err := h.userService.GetCurrentUser(c.Context(), userID)
	if err != nil {
		return sendError(c, requestID, err)
	}

	return c.Status(fiber.StatusOK).JSON(response.OK(requestID, ToUserResponse(user, true)))
}

// UpdateCurrentUser handles updating the current user's profile.
// @Summary Update current user profile
// @Description Update the authenticated user's profile information
// @Tags Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body UpdateProfileRequest true "Profile update data"
// @Success 200 {object} response.Response[UserResponse] "Updated user profile"
// @Failure 400 {object} errors.ErrorResponse "Invalid request body"
// @Failure 401 {object} errors.ErrorResponse "Authentication required"
// @Router /users/me [put]
func (h *UserHandler) UpdateCurrentUser(c *fiber.Ctx) error {
	requestID := middleware.GetRequestID(c)

	// Get user ID from context
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return sendError(c, requestID, errors.Unauthorized(""))
	}

	var req UpdateProfileRequest
	if err := c.BodyParser(&req); err != nil {
		return sendError(c, requestID, errors.BadRequest("invalid request body"))
	}

	// Validate request
	if err := validator.Get().Struct(&req); err != nil {
		return sendError(c, requestID, err)
	}

	// Call service
	if err := h.userService.UpdateProfile(c.Context(), userID, application.UpdateProfileInput{
		Name:      req.Name,
		Bio:       req.Bio,
		AvatarURL: req.AvatarURL,
		Language:  req.Language,
	}); err != nil {
		return sendError(c, requestID, err)
	}

	// Get updated user
	user, err := h.userService.GetCurrentUser(c.Context(), userID)
	if err != nil {
		return sendError(c, requestID, err)
	}

	return c.Status(fiber.StatusOK).JSON(response.OK(requestID, ToUserResponse(user, true)))
}

// GetUser handles getting a user's public profile.
// @Summary Get user public profile
// @Description Get a user's public profile by ID. Includes follow status if authenticated.
// @Tags Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "User ID" format(uuid)
// @Success 200 {object} response.Response[UserProfileResponse] "User public profile"
// @Failure 400 {object} errors.ErrorResponse "Invalid user ID"
// @Failure 404 {object} errors.ErrorResponse "User not found"
// @Router /users/{id} [get]
func (h *UserHandler) GetUser(c *fiber.Ctx) error {
	requestID := middleware.GetRequestID(c)

	// Parse user ID from path
	userIDStr := c.Params("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return sendError(c, requestID, errors.BadRequest("invalid user ID"))
	}

	// Get profile
	profile, err := h.userService.GetProfile(c.Context(), userID)
	if err != nil {
		return sendError(c, requestID, err)
	}

	resp := ToUserProfileResponse(profile)

	// Check if current user is following this user
	currentUserID, ok := middleware.GetUserID(c)
	if ok && currentUserID != userID {
		isFollowing, err := h.userService.IsFollowing(c.Context(), currentUserID, userID)
		if err == nil {
			resp.IsFollowing = isFollowing
		}
	}

	return c.Status(fiber.StatusOK).JSON(response.OK(requestID, resp))
}

// FollowUser handles following a user.
// @Summary Follow a user
// @Description Follow another user to see their activities in your feed
// @Tags Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "User ID to follow" format(uuid)
// @Success 200 {object} response.Response[any] "Successfully followed user"
// @Failure 400 {object} errors.ErrorResponse "Invalid user ID or cannot follow yourself"
// @Failure 401 {object} errors.ErrorResponse "Authentication required"
// @Failure 404 {object} errors.ErrorResponse "User not found"
// @Failure 409 {object} errors.ErrorResponse "Already following this user"
// @Router /users/{id}/follow [post]
func (h *UserHandler) FollowUser(c *fiber.Ctx) error {
	requestID := middleware.GetRequestID(c)

	// Get current user ID from context
	currentUserID, ok := middleware.GetUserID(c)
	if !ok {
		return sendError(c, requestID, errors.Unauthorized(""))
	}

	// Parse target user ID from path
	targetUserIDStr := c.Params("id")
	targetUserID, err := uuid.Parse(targetUserIDStr)
	if err != nil {
		return sendError(c, requestID, errors.BadRequest("invalid user ID"))
	}

	// Call service
	if err := h.userService.Follow(c.Context(), currentUserID, targetUserID); err != nil {
		return sendError(c, requestID, err)
	}

	return c.Status(fiber.StatusOK).JSON(response.Empty(requestID))
}

// UnfollowUser handles unfollowing a user.
// @Summary Unfollow a user
// @Description Stop following a user
// @Tags Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "User ID to unfollow" format(uuid)
// @Success 200 {object} response.Response[any] "Successfully unfollowed user"
// @Failure 400 {object} errors.ErrorResponse "Invalid user ID"
// @Failure 401 {object} errors.ErrorResponse "Authentication required"
// @Failure 404 {object} errors.ErrorResponse "Not following this user"
// @Router /users/{id}/follow [delete]
func (h *UserHandler) UnfollowUser(c *fiber.Ctx) error {
	requestID := middleware.GetRequestID(c)

	// Get current user ID from context
	currentUserID, ok := middleware.GetUserID(c)
	if !ok {
		return sendError(c, requestID, errors.Unauthorized(""))
	}

	// Parse target user ID from path
	targetUserIDStr := c.Params("id")
	targetUserID, err := uuid.Parse(targetUserIDStr)
	if err != nil {
		return sendError(c, requestID, errors.BadRequest("invalid user ID"))
	}

	// Call service
	if err := h.userService.Unfollow(c.Context(), currentUserID, targetUserID); err != nil {
		return sendError(c, requestID, err)
	}

	return c.Status(fiber.StatusOK).JSON(response.Empty(requestID))
}

// GetFollowers handles getting a user's followers.
// @Summary Get user's followers
// @Description Get a paginated list of users following the specified user
// @Tags Users
// @Accept json
// @Produce json
// @Param id path string true "User ID" format(uuid)
// @Param page query int false "Page number" default(1)
// @Param per_page query int false "Items per page" default(20)
// @Success 200 {object} response.PaginatedResponse[UserResponse] "List of followers"
// @Failure 400 {object} errors.ErrorResponse "Invalid user ID"
// @Failure 404 {object} errors.ErrorResponse "User not found"
// @Router /users/{id}/followers [get]
func (h *UserHandler) GetFollowers(c *fiber.Ctx) error {
	requestID := middleware.GetRequestID(c)

	// Parse user ID from path
	userIDStr := c.Params("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return sendError(c, requestID, errors.BadRequest("invalid user ID"))
	}

	// Get pagination parameters
	page := c.QueryInt("page", 1)
	perPage := c.QueryInt("per_page", 20)

	// Call service
	result, err := h.userService.GetFollowers(c.Context(), userID, page, perPage)
	if err != nil {
		return sendError(c, requestID, err)
	}

	return c.Status(fiber.StatusOK).JSON(response.List(
		requestID,
		toUserResponseList(result.Users),
		result.Page,
		result.PerPage,
		result.Total,
	))
}

// GetFollowing handles getting users that a user is following.
// @Summary Get users being followed
// @Description Get a paginated list of users that the specified user is following
// @Tags Users
// @Accept json
// @Produce json
// @Param id path string true "User ID" format(uuid)
// @Param page query int false "Page number" default(1)
// @Param per_page query int false "Items per page" default(20)
// @Success 200 {object} response.PaginatedResponse[UserResponse] "List of followed users"
// @Failure 400 {object} errors.ErrorResponse "Invalid user ID"
// @Failure 404 {object} errors.ErrorResponse "User not found"
// @Router /users/{id}/following [get]
func (h *UserHandler) GetFollowing(c *fiber.Ctx) error {
	requestID := middleware.GetRequestID(c)

	// Parse user ID from path
	userIDStr := c.Params("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return sendError(c, requestID, errors.BadRequest("invalid user ID"))
	}

	// Get pagination parameters
	page := c.QueryInt("page", 1)
	perPage := c.QueryInt("per_page", 20)

	// Call service
	result, err := h.userService.GetFollowing(c.Context(), userID, page, perPage)
	if err != nil {
		return sendError(c, requestID, err)
	}

	return c.Status(fiber.StatusOK).JSON(response.List(
		requestID,
		toUserResponseList(result.Users),
		result.Page,
		result.PerPage,
		result.Total,
	))
}

// StorageInfoResponse represents storage information in API responses.
type StorageInfoResponse struct {
	UsedBytes      int64   `json:"used_bytes"`
	QuotaBytes     int64   `json:"quota_bytes"`
	RemainingBytes int64   `json:"remaining_bytes"`
	UsagePercent   float64 `json:"usage_percent"`
	Tier           string  `json:"tier"`
	Warning        string  `json:"warning,omitempty"`
	NextTier       *string `json:"next_tier,omitempty"`
	NextTierQuota  *int64  `json:"next_tier_quota,omitempty"`
}

// ToStorageInfoResponse converts a domain.StorageInfo to StorageInfoResponse.
func ToStorageInfoResponse(info *domain.StorageInfo) *StorageInfoResponse {
	if info == nil {
		return nil
	}

	resp := &StorageInfoResponse{
		UsedBytes:      info.UsedBytes,
		QuotaBytes:     info.QuotaBytes,
		RemainingBytes: info.RemainingBytes,
		UsagePercent:   info.UsagePercent,
		Tier:           string(info.Tier),
		Warning:        info.Warning,
	}

	if info.NextTier != nil {
		nextTierStr := string(*info.NextTier)
		resp.NextTier = &nextTierStr
	}

	if info.NextTierQuota != nil {
		resp.NextTierQuota = info.NextTierQuota
	}

	return resp
}

// GetStorageInfo handles getting the current user's storage information.
// @Summary Get current user storage info
// @Description Get the authenticated user's storage usage, quota, and tier information
// @Tags Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response[StorageInfoResponse] "Storage information"
// @Failure 401 {object} errors.ErrorResponse "Authentication required"
// @Failure 500 {object} errors.ErrorResponse "Internal server error"
// @Router /users/me/storage [get]
func (h *UserHandler) GetStorageInfo(c *fiber.Ctx) error {
	requestID := middleware.GetRequestID(c)

	// Get user ID from context (set by auth middleware)
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return sendError(c, requestID, errors.Unauthorized(""))
	}

	// Check if storage service is available
	if h.storageService == nil {
		return sendError(c, requestID, errors.Internal("storage service not available", nil))
	}

	// Call service
	info, err := h.storageService.GetStorageInfo(c.Context(), userID)
	if err != nil {
		return sendError(c, requestID, err)
	}

	return c.Status(fiber.StatusOK).JSON(response.OK(requestID, ToStorageInfoResponse(info)))
}

// UpdateTierRequest represents the request body for tier update.
type UpdateTierRequest struct {
	Tier string `json:"tier" validate:"required,oneof=free premium pro"`
}

// UpdateTierResponse represents the response for tier update.
type UpdateTierResponse struct {
	ID   uuid.UUID `json:"id"`
	Tier string    `json:"tier"`
}

// UpdateTier handles updating the current user's subscription tier.
// @Summary Update current user tier
// @Description Update the authenticated user's subscription tier
// @Tags Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body UpdateTierRequest true "Tier update data"
// @Success 200 {object} response.Response[UpdateTierResponse] "Updated tier info"
// @Failure 400 {object} errors.ErrorResponse "Invalid request body or tier value"
// @Failure 401 {object} errors.ErrorResponse "Authentication required"
// @Failure 500 {object} errors.ErrorResponse "Internal server error"
// @Router /users/me/tier [put]
func (h *UserHandler) UpdateTier(c *fiber.Ctx) error {
	requestID := middleware.GetRequestID(c)

	// Get user ID from context (set by auth middleware)
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return sendError(c, requestID, errors.Unauthorized(""))
	}

	// Check if storage service is available
	if h.storageService == nil {
		return sendError(c, requestID, errors.Internal("storage service not available", nil))
	}

	var req UpdateTierRequest
	if err := c.BodyParser(&req); err != nil {
		return sendError(c, requestID, errors.BadRequest("invalid request body"))
	}

	// Validate request
	if err := validator.Get().Struct(&req); err != nil {
		return sendError(c, requestID, err)
	}

	// Convert string to domain.Tier
	tier := domain.Tier(req.Tier)

	// Call service to update tier
	if err := h.storageService.UpdateTier(c.Context(), userID, tier); err != nil {
		return sendError(c, requestID, err)
	}

	// Return updated tier info
	return c.Status(fiber.StatusOK).JSON(response.OK(requestID, &UpdateTierResponse{
		ID:   userID,
		Tier: req.Tier,
	}))
}

// AIUsageInfoResponse represents AI usage information in API responses.
type AIUsageInfoResponse struct {
	UsedToday   int    `json:"used_today"`
	DailyLimit  int    `json:"daily_limit"`
	Remaining   int    `json:"remaining"`
	ResetAt     string `json:"reset_at"`
	Tier        string `json:"tier"`
	IsUnlimited bool   `json:"is_unlimited"`
}

// ToAIUsageInfoResponse converts a domain.AIUsageInfo to AIUsageInfoResponse.
func ToAIUsageInfoResponse(info *domain.AIUsageInfo) *AIUsageInfoResponse {
	if info == nil {
		return nil
	}

	return &AIUsageInfoResponse{
		UsedToday:   info.UsedToday,
		DailyLimit:  info.DailyLimit,
		Remaining:   info.Remaining,
		ResetAt:     info.ResetAt.Format(time.RFC3339),
		Tier:        string(info.Tier),
		IsUnlimited: info.IsUnlimited,
	}
}

// GetAIUsageInfo handles getting the current user's AI usage information.
// @Summary Get current user AI usage info
// @Description Get the authenticated user's AI usage, daily limit, and remaining messages
// @Tags Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response[AIUsageInfoResponse] "AI usage information"
// @Failure 401 {object} errors.ErrorResponse "Authentication required"
// @Failure 500 {object} errors.ErrorResponse "Internal server error"
// @Router /users/me/ai-usage [get]
func (h *UserHandler) GetAIUsageInfo(c *fiber.Ctx) error {
	requestID := middleware.GetRequestID(c)

	// Get user ID from context (set by auth middleware)
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return sendError(c, requestID, errors.Unauthorized(""))
	}

	// Check if AI service is available
	if h.aiService == nil {
		return sendError(c, requestID, errors.Internal("AI service not available", nil))
	}

	// Call service
	info, err := h.aiService.GetAIUsageInfo(c.Context(), userID)
	if err != nil {
		return sendError(c, requestID, err)
	}

	return c.Status(fiber.StatusOK).JSON(response.OK(requestID, ToAIUsageInfoResponse(info)))
}

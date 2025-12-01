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
)

// UserHandler handles user profile and follow-related HTTP requests.
type UserHandler struct {
	userService application.UserService
}

// NewUserHandler creates a new UserHandler.
func NewUserHandler(userService application.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

// UserResponse represents a user in API responses.
type UserResponse struct {
	ID               uuid.UUID `json:"id"`
	Email            string    `json:"email,omitempty"` // Only included for own profile
	Name             string    `json:"name"`
	AvatarURL        *string   `json:"avatar_url,omitempty"`
	Bio              *string   `json:"bio,omitempty"`
	Role             string    `json:"role"`
	EmailVerified    bool      `json:"email_verified,omitempty"` // Only included for own profile
	TwoFactorEnabled bool      `json:"two_factor_enabled,omitempty"` // Only included for own profile
	Language         string    `json:"language,omitempty"` // Only included for own profile
	CreatedAt        time.Time `json:"created_at"`
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
// GET /api/v1/users/me
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
// PUT /api/v1/users/me
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
// GET /api/v1/users/:id
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
// POST /api/v1/users/:id/follow
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
// DELETE /api/v1/users/:id/follow
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
// GET /api/v1/users/:id/followers
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
// GET /api/v1/users/:id/following
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

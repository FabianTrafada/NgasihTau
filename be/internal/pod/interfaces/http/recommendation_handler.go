// Package http provides HTTP handlers for recommendation endpoints.
package http

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"ngasihtau/internal/common/errors"
	"ngasihtau/internal/common/middleware"
	"ngasihtau/internal/common/response"
	"ngasihtau/internal/common/validator"
	"ngasihtau/internal/pod/application"
	"ngasihtau/internal/pod/domain"
)

// RecommendationHandler handles recommendation-related HTTP requests.
type RecommendationHandler struct {
	recommendationService application.RecommendationService
	validator             *validator.Validator
}

// NewRecommendationHandler creates a new RecommendationHandler.
func NewRecommendationHandler(service application.RecommendationService) *RecommendationHandler {
	return &RecommendationHandler{
		recommendationService: service,
		validator:             validator.Get(),
	}
}

// TrackInteractionRequest represents the request body for tracking interactions.
type TrackInteractionRequest struct {
	InteractionType string                      `json:"interaction_type" validate:"required,oneof=view star unstar follow unfollow fork share time_spent material_view material_bookmark search_click"`
	Metadata        *domain.InteractionMetadata `json:"metadata,omitempty"`
	SessionID       *uuid.UUID                  `json:"session_id,omitempty"`
}

// TrackInteraction handles POST /api/v1/pods/:id/track
// @Summary Track user interaction with a pod
// @Description Records user interaction for recommendation algorithm
// @Tags Recommendations
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Pod ID"
// @Param body body TrackInteractionRequest true "Interaction data"
// @Success 204 "Interaction tracked"
// @Failure 400 {object} response.ErrorResponse "Invalid request"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Failure 404 {object} response.ErrorResponse "Pod not found"
// @Router /api/v1/pods/{id}/track [post]
func (h *RecommendationHandler) TrackInteraction(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return errors.Unauthorized("authentication required")
	}

	podIDStr := c.Params("id")
	podID, err := uuid.Parse(podIDStr)
	if err != nil {
		return errors.BadRequest("invalid pod ID")
	}

	var req TrackInteractionRequest
	if err := c.BodyParser(&req); err != nil {
		return errors.BadRequest("invalid request body")
	}

	if err := h.validator.Struct(req); err != nil {
		return err
	}

	input := application.TrackInteractionInput{
		UserID:          userID,
		PodID:           podID,
		InteractionType: domain.InteractionType(req.InteractionType),
		Metadata:        req.Metadata,
		SessionID:       req.SessionID,
	}

	if err := h.recommendationService.TrackInteraction(c.Context(), input); err != nil {
		return err
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// TrackTimeSpentRequest represents the request for tracking time spent.
type TrackTimeSpentRequest struct {
	Seconds int `json:"seconds" validate:"required,min=1,max=3600"`
}

// TrackTimeSpent handles POST /api/v1/pods/:id/track/time
// @Summary Track time spent on a pod
// @Description Records time spent viewing a pod for recommendations
// @Tags Recommendations
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Pod ID"
// @Param body body TrackTimeSpentRequest true "Time spent data"
// @Success 204 "Time tracked"
// @Failure 400 {object} response.ErrorResponse "Invalid request"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Router /api/v1/pods/{id}/track/time [post]
func (h *RecommendationHandler) TrackTimeSpent(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return errors.Unauthorized("authentication required")
	}

	podIDStr := c.Params("id")
	podID, err := uuid.Parse(podIDStr)
	if err != nil {
		return errors.BadRequest("invalid pod ID")
	}

	var req TrackTimeSpentRequest
	if err := c.BodyParser(&req); err != nil {
		return errors.BadRequest("invalid request body")
	}

	if err := h.validator.Struct(req); err != nil {
		return err
	}

	if err := h.recommendationService.TrackTimeSpent(c.Context(), userID, podID, req.Seconds); err != nil {
		return err
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// GetPersonalizedFeed handles GET /api/v1/feed/recommended
// @Summary Get personalized feed
// @Description Returns pods recommended based on user's interaction history
// @Tags Recommendations
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param per_page query int false "Items per page" default(20) maximum(50)
// @Success 200 {object} response.SuccessResponse{data=RecommendedFeedResponse}
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Router /api/v1/feed/recommended [get]
func (h *RecommendationHandler) GetPersonalizedFeed(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return errors.Unauthorized("authentication required")
	}

	page := c.QueryInt("page", 1)
	perPage := c.QueryInt("per_page", 20)

	result, err := h.recommendationService.GetPersonalizedFeed(c.Context(), userID, page, perPage)
	if err != nil {
		return err
	}

	resp := RecommendedFeedResponse{
		Pods:           mapRecommendedPods(result.Pods),
		Page:           result.Page,
		PerPage:        result.PerPage,
		HasMore:        result.HasMore,
		IsPersonalized: result.IsPersonalized,
	}

	requestID := middleware.GetRequestID(c)
	return c.JSON(response.OK(requestID, resp))
}

// GetTrendingFeed handles GET /api/v1/feed/trending
// @Summary Get trending feed
// @Description Returns trending pods based on popularity
// @Tags Recommendations
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param per_page query int false "Items per page" default(20) maximum(50)
// @Success 200 {object} response.SuccessResponse{data=TrendingFeedResponse}
// @Router /api/v1/feed/trending [get]
func (h *RecommendationHandler) GetTrendingFeed(c *fiber.Ctx) error {
	page := c.QueryInt("page", 1)
	perPage := c.QueryInt("per_page", 20)

	result, err := h.recommendationService.GetTrendingFeed(c.Context(), page, perPage)
	if err != nil {
		return err
	}

	resp := TrendingFeedResponse{
		Pods:    mapPods(result.Pods),
		Page:    result.Page,
		PerPage: result.PerPage,
	}

	requestID := middleware.GetRequestID(c)
	return c.JSON(response.OK(requestID, resp))
}

// GetSimilarPods handles GET /api/v1/pods/:id/similar
// @Summary Get similar pods
// @Description Returns pods similar to the specified pod
// @Tags Recommendations
// @Produce json
// @Param id path string true "Pod ID"
// @Param limit query int false "Maximum results" default(10) maximum(20)
// @Success 200 {object} response.SuccessResponse{data=SimilarPodsResponse}
// @Failure 404 {object} response.ErrorResponse "Pod not found"
// @Router /api/v1/pods/{id}/similar [get]
func (h *RecommendationHandler) GetSimilarPods(c *fiber.Ctx) error {
	podIDStr := c.Params("id")
	podID, err := uuid.Parse(podIDStr)
	if err != nil {
		return errors.BadRequest("invalid pod ID")
	}

	limit := c.QueryInt("limit", 10)

	pods, err := h.recommendationService.GetSimilarPods(c.Context(), podID, limit)
	if err != nil {
		return err
	}

	resp := SimilarPodsResponse{
		Pods: mapPods(pods),
	}

	requestID := middleware.GetRequestID(c)
	return c.JSON(response.OK(requestID, resp))
}

// GetUserPreferences handles GET /api/v1/users/me/preferences
// @Summary Get user preferences
// @Description Returns user's content preferences based on interaction history
// @Tags Recommendations
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.SuccessResponse{data=UserPreferencesResponse}
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Router /api/v1/users/me/preferences [get]
func (h *RecommendationHandler) GetUserPreferences(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return errors.Unauthorized("authentication required")
	}

	profile, err := h.recommendationService.GetUserPreferences(c.Context(), userID)
	if err != nil {
		return err
	}

	resp := UserPreferencesResponse{
		TopCategories:     mapCategoryPreferences(profile.TopCategories),
		TopTags:           mapTagPreferences(profile.TopTags),
		TotalInteractions: profile.TotalInteractions,
		HasEnoughData:     profile.HasEnoughData,
	}

	requestID := middleware.GetRequestID(c)
	return c.JSON(response.OK(requestID, resp))
}

// ===========================================
// Response Types
// ===========================================

// RecommendedFeedResponse represents the personalized feed response.
type RecommendedFeedResponse struct {
	Pods           []RecommendedPodResponse `json:"pods"`
	Page           int                      `json:"page"`
	PerPage        int                      `json:"per_page"`
	HasMore        bool                     `json:"has_more"`
	IsPersonalized bool                     `json:"is_personalized"`
}

// RecommendedPodResponse represents a recommended pod with score breakdown.
type RecommendedPodResponse struct {
	ID                  uuid.UUID `json:"id"`
	Name                string    `json:"name"`
	Slug                string    `json:"slug"`
	Description         *string   `json:"description,omitempty"`
	Categories          []string  `json:"categories,omitempty"`
	Tags                []string  `json:"tags,omitempty"`
	StarCount           int       `json:"star_count"`
	ViewCount           int       `json:"view_count"`
	RecommendationScore float64   `json:"recommendation_score"`
	MatchedCategories   []string  `json:"matched_categories,omitempty"`
	MatchedTags         []string  `json:"matched_tags,omitempty"`
}

// TrendingFeedResponse represents the trending feed response.
type TrendingFeedResponse struct {
	Pods    []PodResponse `json:"pods"`
	Page    int           `json:"page"`
	PerPage int           `json:"per_page"`
}

// SimilarPodsResponse represents similar pods response.
type SimilarPodsResponse struct {
	Pods []PodResponse `json:"pods"`
}

// PodResponse represents a pod in API responses.
type PodResponse struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Slug        string    `json:"slug"`
	Description *string   `json:"description,omitempty"`
	Categories  []string  `json:"categories,omitempty"`
	Tags        []string  `json:"tags,omitempty"`
	StarCount   int       `json:"star_count"`
	ViewCount   int       `json:"view_count"`
}

// UserPreferencesResponse represents user preferences.
type UserPreferencesResponse struct {
	TopCategories     []CategoryPreferenceResponse `json:"top_categories"`
	TopTags           []TagPreferenceResponse      `json:"top_tags"`
	TotalInteractions int                          `json:"total_interactions"`
	HasEnoughData     bool                         `json:"has_enough_data"`
}

// CategoryPreferenceResponse represents a category preference.
type CategoryPreferenceResponse struct {
	Category string  `json:"category"`
	Score    float64 `json:"score"`
	Rank     int     `json:"rank"`
}

// TagPreferenceResponse represents a tag preference.
type TagPreferenceResponse struct {
	Tag   string  `json:"tag"`
	Score float64 `json:"score"`
	Rank  int     `json:"rank"`
}

// ===========================================
// Mappers
// ===========================================

func mapRecommendedPods(pods []*domain.RecommendedPod) []RecommendedPodResponse {
	result := make([]RecommendedPodResponse, len(pods))
	for i, p := range pods {
		result[i] = RecommendedPodResponse{
			ID:                  p.Pod.ID,
			Name:                p.Pod.Name,
			Slug:                p.Pod.Slug,
			Description:         p.Pod.Description,
			Categories:          p.Pod.Categories,
			Tags:                p.Pod.Tags,
			StarCount:           p.Pod.StarCount,
			ViewCount:           p.Pod.ViewCount,
			RecommendationScore: p.RecommendationScore,
			MatchedCategories:   p.MatchedCategories,
			MatchedTags:         p.MatchedTags,
		}
	}
	return result
}

func mapPods(pods []*domain.Pod) []PodResponse {
	result := make([]PodResponse, len(pods))
	for i, p := range pods {
		result[i] = PodResponse{
			ID:          p.ID,
			Name:        p.Name,
			Slug:        p.Slug,
			Description: p.Description,
			Categories:  p.Categories,
			Tags:        p.Tags,
			StarCount:   p.StarCount,
			ViewCount:   p.ViewCount,
		}
	}
	return result
}

func mapCategoryPreferences(prefs []domain.CategoryPreference) []CategoryPreferenceResponse {
	result := make([]CategoryPreferenceResponse, len(prefs))
	for i, p := range prefs {
		result[i] = CategoryPreferenceResponse{
			Category: p.Category,
			Score:    p.Score,
			Rank:     p.Rank,
		}
	}
	return result
}

func mapTagPreferences(prefs []domain.TagPreference) []TagPreferenceResponse {
	result := make([]TagPreferenceResponse, len(prefs))
	for i, p := range prefs {
		result[i] = TagPreferenceResponse{
			Tag:   p.Tag,
			Score: p.Score,
			Rank:  p.Rank,
		}
	}
	return result
}

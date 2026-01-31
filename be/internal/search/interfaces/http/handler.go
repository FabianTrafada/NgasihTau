package http

import (
	"strconv"

	"github.com/gofiber/fiber/v2"

	"ngasihtau/docs"
	"ngasihtau/internal/common/errors"
	"ngasihtau/internal/common/middleware"
	"ngasihtau/internal/common/response"
	"ngasihtau/internal/search/application"
	"ngasihtau/internal/search/domain"
	"ngasihtau/pkg/jwt"
)

// Ensure docs package is used for swagger
var _ = docs.Meta{}

// Handler handles HTTP requests for search service
type Handler struct {
	service    *application.Service
	jwtManager *jwt.Manager
}

// NewHandler creates a new search HTTP handler
func NewHandler(service *application.Service, jwtManager *jwt.Manager) *Handler {
	return &Handler{
		service:    service,
		jwtManager: jwtManager,
	}
}

// RegisterRoutes registers search routes on the Fiber app
func (h *Handler) RegisterRoutes(app *fiber.App) {
	api := app.Group("/api/v1")

	// Public routes (with optional auth for search history)
	api.Get("/search", middleware.OptionalAuth(h.jwtManager), h.Search)
	api.Get("/search/semantic", h.SemanticSearch)
	api.Get("/search/hybrid", middleware.OptionalAuth(h.jwtManager), h.HybridSearch)
	api.Get("/search/suggestions", h.GetSuggestions)
	api.Get("/search/trending", h.GetTrending)
	api.Get("/search/popular", h.GetPopular)

	// Protected routes
	api.Get("/search/history", middleware.Auth(h.jwtManager), h.GetSearchHistory)
	api.Delete("/search/history", middleware.Auth(h.jwtManager), h.ClearSearchHistory)
}

// Search handles full-text search requests
// @Summary Full-text search
// @Description Search for pods and materials using full-text search
// @Tags Search
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param q query string true "Search query"
// @Param type query string false "Filter by type (pod, material)"
// @Param category query string false "Filter by category"
// @Param file_type query string false "Filter by file type (pdf, docx, pptx)"
// @Param pod_id query string false "Filter by pod ID" format(uuid)
// @Param verified query bool false "Filter by verified status (teacher-created pods)"
// @Param sort query string false "Sort by (relevance, upvotes, trust_score, recent, popular)" default(relevance)
// @Param page query int false "Page number" default(1)
// @Param per_page query int false "Items per page" default(20)
// @Success 200 {object} response.PaginatedResponse[any] "Search results"
// @Router /search [get]
func (h *Handler) Search(c *fiber.Ctx) error {
	requestID := middleware.GetRequestID(c)

	query := c.Query("q")
	typesParam := c.Query("type")
	categoryParam := c.Query("category")
	fileTypeParam := c.Query("file_type")
	podID := c.Query("pod_id")
	verifiedParam := c.Query("verified")
	sortParam := c.Query("sort", "relevance")
	page, _ := strconv.Atoi(c.Query("page", "1"))
	perPage, _ := strconv.Atoi(c.Query("per_page", "20"))

	// Parse array params
	var typesArr, categoriesArr, fileTypesArr []string
	if typesParam != "" {
		typesArr = []string{typesParam}
	}
	if categoryParam != "" {
		categoriesArr = []string{categoryParam}
	}
	if fileTypeParam != "" {
		fileTypesArr = []string{fileTypeParam}
	}

	// Parse verified filter (Requirement 6.3)
	var verified *bool
	if verifiedParam != "" {
		v, err := strconv.ParseBool(verifiedParam)
		if err == nil {
			verified = &v
		}
	}

	// Parse sort option (Requirement 6.4, 6.5)
	sortBy := parseSortBy(sortParam)

	// Get user ID if authenticated (optional)
	var userIDStr string
	if userID, ok := middleware.GetUserID(c); ok {
		userIDStr = userID.String()
	}

	input := application.SearchInput{
		Query:      query,
		Types:      typesArr,
		Categories: categoriesArr,
		FileTypes:  fileTypesArr,
		PodID:      podID,
		Verified:   verified,
		SortBy:     sortBy,
		Page:       page,
		PerPage:    perPage,
		UserID:     userIDStr,
	}

	output, err := h.service.Search(c.Context(), input)
	if err != nil {
		return sendError(c, requestID, err)
	}

	return c.Status(fiber.StatusOK).JSON(response.List(requestID, output.Results, output.Page, output.PerPage, int(output.Total)))
}

// SemanticSearch handles semantic search requests
// @Summary Semantic search
// @Description Search using natural language with vector similarity
// @Tags Search
// @Accept json
// @Produce json
// @Param q query string true "Natural language query"
// @Param pod_id query string false "Filter by pod ID" format(uuid)
// @Param limit query int false "Maximum results" default(10)
// @Param min_score query number false "Minimum similarity score" default(0)
// @Success 200 {object} response.Response[any] "Semantic search results"
// @Failure 400 {object} errors.ErrorResponse "Query parameter required"
// @Router /search/semantic [get]
func (h *Handler) SemanticSearch(c *fiber.Ctx) error {
	requestID := middleware.GetRequestID(c)

	query := c.Query("q")
	if query == "" {
		return sendError(c, requestID, errors.BadRequest("query parameter 'q' is required"))
	}

	podID := c.Query("pod_id")
	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	minScore, _ := strconv.ParseFloat(c.Query("min_score", "0"), 64)

	input := application.SemanticSearchInput{
		Query:    query,
		PodID:    podID,
		Limit:    limit,
		MinScore: minScore,
	}

	results, err := h.service.SemanticSearch(c.Context(), input)
	if err != nil {
		return sendError(c, requestID, err)
	}

	return c.Status(fiber.StatusOK).JSON(response.OK(requestID, results))
}

// HybridSearch handles hybrid search requests combining keyword and semantic search
// @Summary Hybrid search
// @Description Combine keyword and semantic search for best results
// @Tags Search
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param q query string true "Search query"
// @Param type query string false "Filter by type"
// @Param category query string false "Filter by category"
// @Param file_type query string false "Filter by file type"
// @Param pod_id query string false "Filter by pod ID" format(uuid)
// @Param verified query bool false "Filter by verified status (teacher-created pods)"
// @Param sort query string false "Sort by (relevance, upvotes, trust_score, recent, popular)" default(relevance)
// @Param page query int false "Page number" default(1)
// @Param per_page query int false "Items per page" default(20)
// @Param semantic_weight query number false "Weight for semantic results (0-1)" default(0.3)
// @Success 200 {object} response.PaginatedResponse[any] "Hybrid search results"
// @Router /search/hybrid [get]
func (h *Handler) HybridSearch(c *fiber.Ctx) error {
	requestID := middleware.GetRequestID(c)

	query := c.Query("q")
	typesParam := c.Query("type")
	categoryParam := c.Query("category")
	fileTypeParam := c.Query("file_type")
	podID := c.Query("pod_id")
	verifiedParam := c.Query("verified")
	sortParam := c.Query("sort", "relevance")
	page, _ := strconv.Atoi(c.Query("page", "1"))
	perPage, _ := strconv.Atoi(c.Query("per_page", "20"))
	semanticWeight, _ := strconv.ParseFloat(c.Query("semantic_weight", "0.3"), 64)

	// Parse array params
	var typesArr, categoriesArr, fileTypesArr []string
	if typesParam != "" {
		typesArr = []string{typesParam}
	}
	if categoryParam != "" {
		categoriesArr = []string{categoryParam}
	}
	if fileTypeParam != "" {
		fileTypesArr = []string{fileTypeParam}
	}

	// Parse verified filter (Requirement 6.3)
	var verified *bool
	if verifiedParam != "" {
		v, err := strconv.ParseBool(verifiedParam)
		if err == nil {
			verified = &v
		}
	}

	// Parse sort option (Requirement 6.4, 6.5)
	sortBy := parseSortBy(sortParam)

	// Get user ID if authenticated (optional)
	var userIDStr string
	if userID, ok := middleware.GetUserID(c); ok {
		userIDStr = userID.String()
	}

	input := application.HybridSearchInput{
		Query:          query,
		Types:          typesArr,
		Categories:     categoriesArr,
		FileTypes:      fileTypesArr,
		PodID:          podID,
		Verified:       verified,
		SortBy:         sortBy,
		Page:           page,
		PerPage:        perPage,
		SemanticWeight: semanticWeight,
		UserID:         userIDStr,
	}

	output, err := h.service.HybridSearch(c.Context(), input)
	if err != nil {
		return sendError(c, requestID, err)
	}

	return c.Status(fiber.StatusOK).JSON(response.List(requestID, output.Results, output.Page, output.PerPage, int(output.Total)))
}

// GetSuggestions handles autocomplete suggestion requests
// @Summary Get search suggestions
// @Description Get autocomplete suggestions based on query prefix
// @Tags Search
// @Accept json
// @Produce json
// @Param q query string true "Query prefix"
// @Param limit query int false "Maximum suggestions" default(10)
// @Success 200 {object} response.Response[map[string][]string] "Suggestions"
// @Failure 400 {object} errors.ErrorResponse "Query parameter required"
// @Router /search/suggestions [get]
func (h *Handler) GetSuggestions(c *fiber.Ctx) error {
	requestID := middleware.GetRequestID(c)

	prefix := c.Query("q")
	if prefix == "" {
		return sendError(c, requestID, errors.BadRequest("query parameter 'q' is required"))
	}

	limit, _ := strconv.Atoi(c.Query("limit", "10"))

	suggestions, err := h.service.GetSuggestions(c.Context(), prefix, limit)
	if err != nil {
		return sendError(c, requestID, err)
	}

	return c.Status(fiber.StatusOK).JSON(response.OK(requestID, fiber.Map{
		"suggestions": suggestions,
	}))
}

// GetTrending handles trending materials requests
// @Summary Get trending materials
// @Description Get materials ranked by recent engagement (last 7 days)
// @Tags Search
// @Accept json
// @Produce json
// @Param category query string false "Filter by category"
// @Param limit query int false "Maximum results" default(20)
// @Success 200 {object} response.Response[any] "Trending materials"
// @Router /search/trending [get]
func (h *Handler) GetTrending(c *fiber.Ctx) error {
	requestID := middleware.GetRequestID(c)

	category := c.Query("category")
	limit, _ := strconv.Atoi(c.Query("limit", "20"))

	materials, err := h.service.GetTrending(c.Context(), category, limit)
	if err != nil {
		return sendError(c, requestID, err)
	}

	return c.Status(fiber.StatusOK).JSON(response.OK(requestID, materials))
}

// GetPopular handles popular materials requests
// @Summary Get popular materials
// @Description Get materials ranked by all-time engagement
// @Tags Search
// @Accept json
// @Produce json
// @Param category query string false "Filter by category"
// @Param limit query int false "Maximum results" default(20)
// @Success 200 {object} response.Response[any] "Popular materials"
// @Router /search/popular [get]
func (h *Handler) GetPopular(c *fiber.Ctx) error {
	requestID := middleware.GetRequestID(c)

	category := c.Query("category")
	limit, _ := strconv.Atoi(c.Query("limit", "20"))

	materials, err := h.service.GetPopular(c.Context(), category, limit)
	if err != nil {
		return sendError(c, requestID, err)
	}

	return c.Status(fiber.StatusOK).JSON(response.OK(requestID, materials))
}

// GetSearchHistory handles search history requests
// @Summary Get search history
// @Description Get the authenticated user's search history
// @Tags Search
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param limit query int false "Maximum results" default(20)
// @Success 200 {object} docs.SuccessResponse "Search history"
// @Failure 401 {object} errors.ErrorResponse "Authentication required"
// @Router /search/history [get]
func (h *Handler) GetSearchHistory(c *fiber.Ctx) error {
	requestID := middleware.GetRequestID(c)

	userID, ok := middleware.GetUserID(c)
	if !ok {
		return sendError(c, requestID, errors.Unauthorized("authentication required"))
	}

	limit, _ := strconv.Atoi(c.Query("limit", "20"))

	history, err := h.service.GetSearchHistory(c.Context(), userID.String(), limit)
	if err != nil {
		return sendError(c, requestID, err)
	}

	return c.Status(fiber.StatusOK).JSON(response.OK(requestID, fiber.Map{
		"history": history,
	}))
}

// ClearSearchHistory handles clearing search history
// @Summary Clear search history
// @Description Clear the authenticated user's search history
// @Tags Search
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response[map[string]string] "History cleared"
// @Failure 401 {object} errors.ErrorResponse "Authentication required"
// @Router /search/history [delete]
func (h *Handler) ClearSearchHistory(c *fiber.Ctx) error {
	requestID := middleware.GetRequestID(c)

	userID, ok := middleware.GetUserID(c)
	if !ok {
		return sendError(c, requestID, errors.Unauthorized("authentication required"))
	}

	err := h.service.ClearSearchHistory(c.Context(), userID.String())
	if err != nil {
		return sendError(c, requestID, err)
	}

	return c.Status(fiber.StatusOK).JSON(response.OK(requestID, fiber.Map{
		"message": "Search history cleared",
	}))
}

// parseSortBy converts a string sort parameter to domain.SortBy
// Implements Requirement 6.4, 6.5: Trust score sorting
func parseSortBy(s string) domain.SortBy {
	switch s {
	case "upvotes":
		return domain.SortByUpvotes
	case "trust_score":
		return domain.SortByTrustScore
	case "recent":
		return domain.SortByRecent
	case "popular":
		return domain.SortByPopular
	case "relevance":
		return domain.SortByRelevance
	default:
		return domain.SortByRelevance
	}
}

// sendError sends an error response using the standard error format.
func sendError(c *fiber.Ctx, requestID string, err error) error {
	resp := errors.BuildResponse(requestID, err)
	status := errors.GetHTTPStatus(err)
	return c.Status(status).JSON(resp)
}

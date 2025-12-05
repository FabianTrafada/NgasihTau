package http

import (
	"strconv"

	"github.com/gofiber/fiber/v2"

	"ngasihtau/internal/common/errors"
	"ngasihtau/internal/common/middleware"
	"ngasihtau/internal/common/response"
	"ngasihtau/internal/search/application"
	"ngasihtau/pkg/jwt"
)

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
// GET /api/v1/search
func (h *Handler) Search(c *fiber.Ctx) error {
	requestID := middleware.GetRequestID(c)

	query := c.Query("q")
	typesParam := c.Query("type")
	categoryParam := c.Query("category")
	fileTypeParam := c.Query("file_type")
	podID := c.Query("pod_id")
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
// GET /api/v1/search/semantic
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
// GET /api/v1/search/hybrid
func (h *Handler) HybridSearch(c *fiber.Ctx) error {
	requestID := middleware.GetRequestID(c)

	query := c.Query("q")
	typesParam := c.Query("type")
	categoryParam := c.Query("category")
	fileTypeParam := c.Query("file_type")
	podID := c.Query("pod_id")
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
// GET /api/v1/search/suggestions
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
// GET /api/v1/search/trending
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
// GET /api/v1/search/popular
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
// GET /api/v1/search/history
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
// DELETE /api/v1/search/history
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

// sendError sends an error response using the standard error format.
func sendError(c *fiber.Ctx, requestID string, err error) error {
	resp := errors.BuildResponse(requestID, err)
	status := errors.GetHTTPStatus(err)
	return c.Status(status).JSON(resp)
}

package application

import (
	"context"
	"sort"

	"ngasihtau/internal/search/domain"
)

// Service handles search business logic
type Service struct {
	searchRepo   domain.SearchRepository
	indexRepo    domain.IndexRepository
	vectorRepo   domain.VectorSearchRepository
	historyRepo  domain.SearchHistoryRepository
	trendingRepo domain.TrendingRepository
	embeddingGen domain.EmbeddingGenerator
}

// NewService creates a new search service
func NewService(
	searchRepo domain.SearchRepository,
	indexRepo domain.IndexRepository,
	vectorRepo domain.VectorSearchRepository,
	historyRepo domain.SearchHistoryRepository,
	trendingRepo domain.TrendingRepository,
) *Service {
	return &Service{
		searchRepo:   searchRepo,
		indexRepo:    indexRepo,
		vectorRepo:   vectorRepo,
		historyRepo:  historyRepo,
		trendingRepo: trendingRepo,
	}
}

// SetEmbeddingGenerator sets the embedding generator for semantic search
func (s *Service) SetEmbeddingGenerator(gen domain.EmbeddingGenerator) {
	s.embeddingGen = gen
}

// SearchInput represents search request input
type SearchInput struct {
	Query      string
	Types      []string
	Categories []string
	FileTypes  []string
	PodID      string
	Verified   *bool         // Filter by verified status (teacher-created pods)
	SortBy     domain.SortBy // Sorting option (relevance, upvotes, trust_score, recent, popular)
	Page       int
	PerPage    int
	UserID     string // for saving search history
}

// SearchOutput represents search response
type SearchOutput struct {
	Results    []domain.SearchResult
	Total      int64
	Page       int
	PerPage    int
	TotalPages int64
}

// Search performs full-text search across pods and materials
// Implements Requirement 7: Material Search
func (s *Service) Search(ctx context.Context, input SearchInput) (*SearchOutput, error) {
	if input.Page < 1 {
		input.Page = 1
	}
	if input.PerPage < 1 || input.PerPage > 100 {
		input.PerPage = 20
	}

	query := domain.SearchQuery{
		Query:      input.Query,
		Types:      input.Types,
		Categories: input.Categories,
		FileTypes:  input.FileTypes,
		PodID:      input.PodID,
		Verified:   input.Verified,
		SortBy:     input.SortBy,
		Page:       input.Page,
		PerPage:    input.PerPage,
	}

	results, total, err := s.searchRepo.Search(ctx, query)
	if err != nil {
		return nil, err
	}

	// Save to search history if user is authenticated
	if input.UserID != "" && input.Query != "" {
		_ = s.historyRepo.SaveSearch(ctx, input.UserID, input.Query)
	}

	totalPages := (total + int64(input.PerPage) - 1) / int64(input.PerPage)

	return &SearchOutput{
		Results:    results,
		Total:      total,
		Page:       input.Page,
		PerPage:    input.PerPage,
		TotalPages: totalPages,
	}, nil
}

// SemanticSearchInput represents semantic search request
type SemanticSearchInput struct {
	Query    string
	PodID    string
	Limit    int
	MinScore float64
}

// SemanticSearch performs vector similarity search using Qdrant
// Implements Requirement 7.1: Semantic Search
func (s *Service) SemanticSearch(ctx context.Context, input SemanticSearchInput) ([]domain.SearchResult, error) {
	if input.Limit < 1 || input.Limit > 50 {
		input.Limit = 10
	}

	// Generate embedding for the query using OpenAI
	var embedding []float32
	var err error

	if s.embeddingGen != nil {
		embedding, err = s.embeddingGen.GenerateEmbedding(ctx, input.Query)
		if err != nil {
			// Log error but continue with zero embedding (will return no results)
			embedding = make([]float32, 1536)
		}
	} else {
		// No embedding generator configured, return empty results
		embedding = make([]float32, 1536)
	}

	query := domain.SemanticSearchQuery{
		Query:    input.Query,
		PodID:    input.PodID,
		Limit:    input.Limit,
		MinScore: input.MinScore,
	}

	return s.vectorRepo.SemanticSearch(ctx, query, embedding)
}

// HybridSearchInput represents hybrid search request
type HybridSearchInput struct {
	Query          string
	Types          []string
	Categories     []string
	FileTypes      []string
	PodID          string
	Verified       *bool         // Filter by verified status (teacher-created pods)
	SortBy         domain.SortBy // Sorting option (relevance, upvotes, trust_score, recent, popular)
	Page           int
	PerPage        int
	SemanticWeight float64 // 0.0 to 1.0, weight for semantic results (default 0.3)
	UserID         string
}

// HybridSearch combines keyword search with semantic search for better results
// Implements Requirement 7.1: Semantic Search (hybrid ranking)
func (s *Service) HybridSearch(ctx context.Context, input HybridSearchInput) (*SearchOutput, error) {
	if input.Page < 1 {
		input.Page = 1
	}
	if input.PerPage < 1 || input.PerPage > 100 {
		input.PerPage = 20
	}
	if input.SemanticWeight < 0 || input.SemanticWeight > 1 {
		input.SemanticWeight = 0.3 // Default 30% semantic, 70% keyword
	}

	// Perform keyword search
	keywordQuery := domain.SearchQuery{
		Query:      input.Query,
		Types:      input.Types,
		Categories: input.Categories,
		FileTypes:  input.FileTypes,
		PodID:      input.PodID,
		Verified:   input.Verified,
		Page:       1,
		PerPage:    input.PerPage * 2, // Fetch more for merging
	}

	keywordResults, keywordTotal, err := s.searchRepo.Search(ctx, keywordQuery)
	if err != nil {
		return nil, err
	}

	// Perform semantic search if embedding generator is available
	var semanticResults []domain.SearchResult
	if s.embeddingGen != nil && input.Query != "" {
		embedding, err := s.embeddingGen.GenerateEmbedding(ctx, input.Query)
		if err == nil {
			semanticQuery := domain.SemanticSearchQuery{
				Query:    input.Query,
				PodID:    input.PodID,
				Limit:    input.PerPage * 2,
				MinScore: 0.5, // Minimum similarity threshold
			}
			semanticResults, _ = s.vectorRepo.SemanticSearch(ctx, semanticQuery, embedding)
		}
	}

	// Merge and rank results using reciprocal rank fusion
	mergedResults := s.mergeResults(keywordResults, semanticResults, input.SemanticWeight)

	// Apply pagination
	start := (input.Page - 1) * input.PerPage
	end := start + input.PerPage
	if start >= len(mergedResults) {
		mergedResults = []domain.SearchResult{}
	} else if end > len(mergedResults) {
		mergedResults = mergedResults[start:]
	} else {
		mergedResults = mergedResults[start:end]
	}

	// Save to search history if user is authenticated
	if input.UserID != "" && input.Query != "" {
		_ = s.historyRepo.SaveSearch(ctx, input.UserID, input.Query)
	}

	// Estimate total (use keyword total as base)
	total := keywordTotal
	totalPages := (total + int64(input.PerPage) - 1) / int64(input.PerPage)

	return &SearchOutput{
		Results:    mergedResults,
		Total:      total,
		Page:       input.Page,
		PerPage:    input.PerPage,
		TotalPages: totalPages,
	}, nil
}

// mergeResults combines keyword and semantic results using reciprocal rank fusion
func (s *Service) mergeResults(keywordResults, semanticResults []domain.SearchResult, semanticWeight float64) []domain.SearchResult {
	keywordWeight := 1.0 - semanticWeight
	k := 60.0 // RRF constant

	// Build score map
	scoreMap := make(map[string]float64)
	resultMap := make(map[string]domain.SearchResult)

	// Score keyword results
	for i, result := range keywordResults {
		rank := float64(i + 1)
		score := keywordWeight * (1.0 / (k + rank))
		scoreMap[result.ID] = score
		resultMap[result.ID] = result
	}

	// Score semantic results and merge
	for i, result := range semanticResults {
		rank := float64(i + 1)
		score := semanticWeight * (1.0 / (k + rank))
		if existing, ok := scoreMap[result.ID]; ok {
			scoreMap[result.ID] = existing + score
		} else {
			scoreMap[result.ID] = score
			resultMap[result.ID] = result
		}
	}

	// Convert to slice and sort by score
	type scoredResult struct {
		result domain.SearchResult
		score  float64
	}

	var scored []scoredResult
	for id, score := range scoreMap {
		result := resultMap[id]
		result.Score = score
		scored = append(scored, scoredResult{result: result, score: score})
	}

	sort.Slice(scored, func(i, j int) bool {
		return scored[i].score > scored[j].score
	})

	// Extract results
	results := make([]domain.SearchResult, len(scored))
	for i, sr := range scored {
		results[i] = sr.result
	}

	return results
}

// GetSuggestions returns autocomplete suggestions based on prefix
// Implements Requirement 7.3: Search History & Suggestions
func (s *Service) GetSuggestions(ctx context.Context, prefix string, limit int) ([]string, error) {
	if limit < 1 || limit > 20 {
		limit = 10
	}
	return s.searchRepo.GetSuggestions(ctx, prefix, limit)
}

// GetTrending returns trending materials based on recent engagement
// Implements Requirement 7.2: Trending & Popular Materials
func (s *Service) GetTrending(ctx context.Context, category string, limit int) ([]domain.TrendingMaterial, error) {
	if limit < 1 || limit > 50 {
		limit = 20
	}
	return s.trendingRepo.GetTrending(ctx, category, limit)
}

// GetPopular returns popular materials based on all-time engagement
// Implements Requirement 7.2: Trending & Popular Materials
func (s *Service) GetPopular(ctx context.Context, category string, limit int) ([]domain.MaterialDocument, error) {
	if limit < 1 || limit > 50 {
		limit = 20
	}
	return s.trendingRepo.GetPopular(ctx, category, limit)
}

// GetSearchHistory returns user's search history
// Implements Requirement 7.3: Search History & Suggestions
func (s *Service) GetSearchHistory(ctx context.Context, userID string, limit int) ([]domain.SearchHistory, error) {
	if limit < 1 || limit > 50 {
		limit = 20
	}
	return s.historyRepo.GetHistory(ctx, userID, limit)
}

// ClearSearchHistory clears user's search history
// Implements Requirement 7.3: Search History & Suggestions
func (s *Service) ClearSearchHistory(ctx context.Context, userID string) error {
	return s.historyRepo.ClearHistory(ctx, userID)
}

// IndexPod indexes a pod document for search
func (s *Service) IndexPod(ctx context.Context, pod domain.PodDocument) error {
	return s.indexRepo.IndexPod(ctx, pod)
}

// IndexMaterial indexes a material document for search
func (s *Service) IndexMaterial(ctx context.Context, material domain.MaterialDocument) error {
	return s.indexRepo.IndexMaterial(ctx, material)
}

// DeletePodIndex removes a pod from the search index
func (s *Service) DeletePodIndex(ctx context.Context, podID string) error {
	return s.indexRepo.DeletePod(ctx, podID)
}

// DeleteMaterialIndex removes a material from the search index
func (s *Service) DeleteMaterialIndex(ctx context.Context, materialID string) error {
	return s.indexRepo.DeleteMaterial(ctx, materialID)
}

// UpdatePodUpvoteCount updates the upvote count for a pod in the search index.
// Implements Requirements 6.2, 6.4, 6.5: Trust indicator updates.
func (s *Service) UpdatePodUpvoteCount(ctx context.Context, podID string, upvoteCount int) error {
	return s.indexRepo.UpdatePodUpvoteCount(ctx, podID, upvoteCount)
}

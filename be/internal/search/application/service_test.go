// Package application contains unit tests for the Search Service.
package application

import (
	"context"
	"strings"
	"testing"
	"time"

	"ngasihtau/internal/search/domain"
)

// Mock implementations for repositories

type mockSearchRepo struct {
	results     []domain.SearchResult
	total       int64
	suggestions []string
	searchErr   error
	suggestErr  error
}

func newMockSearchRepo() *mockSearchRepo {
	return &mockSearchRepo{
		results:     []domain.SearchResult{},
		suggestions: []string{},
	}
}

func (m *mockSearchRepo) Search(ctx context.Context, query domain.SearchQuery) ([]domain.SearchResult, int64, error) {
	if m.searchErr != nil {
		return nil, 0, m.searchErr
	}
	return m.results, m.total, nil
}

func (m *mockSearchRepo) GetSuggestions(ctx context.Context, prefix string, limit int) ([]string, error) {
	if m.suggestErr != nil {
		return nil, m.suggestErr
	}
	// Filter suggestions by prefix
	var filtered []string
	for _, s := range m.suggestions {
		if strings.HasPrefix(strings.ToLower(s), strings.ToLower(prefix)) {
			filtered = append(filtered, s)
			if len(filtered) >= limit {
				break
			}
		}
	}
	return filtered, nil
}

type mockIndexRepo struct {
	pods      map[string]domain.PodDocument
	materials map[string]domain.MaterialDocument
	indexErr  error
	deleteErr error
}

func newMockIndexRepo() *mockIndexRepo {
	return &mockIndexRepo{
		pods:      make(map[string]domain.PodDocument),
		materials: make(map[string]domain.MaterialDocument),
	}
}

func (m *mockIndexRepo) IndexPod(ctx context.Context, pod domain.PodDocument) error {
	if m.indexErr != nil {
		return m.indexErr
	}
	m.pods[pod.ID] = pod
	return nil
}

func (m *mockIndexRepo) UpdatePod(ctx context.Context, pod domain.PodDocument) error {
	if m.indexErr != nil {
		return m.indexErr
	}
	m.pods[pod.ID] = pod
	return nil
}

func (m *mockIndexRepo) DeletePod(ctx context.Context, podID string) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	delete(m.pods, podID)
	return nil
}

func (m *mockIndexRepo) IndexMaterial(ctx context.Context, material domain.MaterialDocument) error {
	if m.indexErr != nil {
		return m.indexErr
	}
	m.materials[material.ID] = material
	return nil
}

func (m *mockIndexRepo) UpdateMaterial(ctx context.Context, material domain.MaterialDocument) error {
	if m.indexErr != nil {
		return m.indexErr
	}
	m.materials[material.ID] = material
	return nil
}

func (m *mockIndexRepo) DeleteMaterial(ctx context.Context, materialID string) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	delete(m.materials, materialID)
	return nil
}

func (m *mockIndexRepo) UpdatePodUpvoteCount(ctx context.Context, podID string, upvoteCount int) error {
	if m.indexErr != nil {
		return m.indexErr
	}
	if pod, ok := m.pods[podID]; ok {
		pod.UpvoteCount = upvoteCount
		m.pods[podID] = pod
	}
	return nil
}

type mockVectorSearchRepo struct {
	results   []domain.SearchResult
	searchErr error
}

func newMockVectorSearchRepo() *mockVectorSearchRepo {
	return &mockVectorSearchRepo{
		results: []domain.SearchResult{},
	}
}

func (m *mockVectorSearchRepo) SemanticSearch(ctx context.Context, query domain.SemanticSearchQuery, embedding []float32) ([]domain.SearchResult, error) {
	if m.searchErr != nil {
		return nil, m.searchErr
	}
	return m.results, nil
}

type mockSearchHistoryRepo struct {
	history  map[string][]domain.SearchHistory
	saveErr  error
	getErr   error
	clearErr error
}

func newMockSearchHistoryRepo() *mockSearchHistoryRepo {
	return &mockSearchHistoryRepo{
		history: make(map[string][]domain.SearchHistory),
	}
}

func (m *mockSearchHistoryRepo) SaveSearch(ctx context.Context, userID, query string) error {
	if m.saveErr != nil {
		return m.saveErr
	}
	entry := domain.SearchHistory{
		ID:        "hist_" + userID + "_" + query,
		UserID:    userID,
		Query:     query,
		CreatedAt: time.Now(),
	}
	m.history[userID] = append(m.history[userID], entry)
	return nil
}

func (m *mockSearchHistoryRepo) GetHistory(ctx context.Context, userID string, limit int) ([]domain.SearchHistory, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	history := m.history[userID]
	if limit > 0 && limit < len(history) {
		return history[:limit], nil
	}
	return history, nil
}

func (m *mockSearchHistoryRepo) ClearHistory(ctx context.Context, userID string) error {
	if m.clearErr != nil {
		return m.clearErr
	}
	delete(m.history, userID)
	return nil
}

type mockTrendingRepo struct {
	trending []domain.TrendingMaterial
	popular  []domain.MaterialDocument
	trendErr error
	popErr   error
}

func newMockTrendingRepo() *mockTrendingRepo {
	return &mockTrendingRepo{
		trending: []domain.TrendingMaterial{},
		popular:  []domain.MaterialDocument{},
	}
}

func (m *mockTrendingRepo) GetTrending(ctx context.Context, category string, limit int) ([]domain.TrendingMaterial, error) {
	if m.trendErr != nil {
		return nil, m.trendErr
	}
	if limit > 0 && limit < len(m.trending) {
		return m.trending[:limit], nil
	}
	return m.trending, nil
}

func (m *mockTrendingRepo) GetPopular(ctx context.Context, category string, limit int) ([]domain.MaterialDocument, error) {
	if m.popErr != nil {
		return nil, m.popErr
	}
	if limit > 0 && limit < len(m.popular) {
		return m.popular[:limit], nil
	}
	return m.popular, nil
}

type mockEmbeddingGenerator struct {
	embedding []float32
	err       error
}

func newMockEmbeddingGenerator() *mockEmbeddingGenerator {
	return &mockEmbeddingGenerator{
		embedding: make([]float32, 1536),
	}
}

func (m *mockEmbeddingGenerator) GenerateEmbedding(ctx context.Context, text string) ([]float32, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.embedding, nil
}

// Helper to create a test service
func newTestService() (*Service, *mockSearchRepo, *mockIndexRepo, *mockVectorSearchRepo, *mockSearchHistoryRepo, *mockTrendingRepo) {
	searchRepo := newMockSearchRepo()
	indexRepo := newMockIndexRepo()
	vectorRepo := newMockVectorSearchRepo()
	historyRepo := newMockSearchHistoryRepo()
	trendingRepo := newMockTrendingRepo()

	svc := NewService(
		searchRepo,
		indexRepo,
		vectorRepo,
		historyRepo,
		trendingRepo,
	)

	return svc, searchRepo, indexRepo, vectorRepo, historyRepo, trendingRepo
}

// Test: Full-text Search
func TestSearch_Success(t *testing.T) {
	svc, searchRepo, _, _, _, _ := newTestService()
	ctx := context.Background()

	// Setup mock data
	searchRepo.results = []domain.SearchResult{
		{
			ID:          "mat_1",
			Type:        "material",
			Title:       "Introduction to Machine Learning",
			Description: "A comprehensive guide to ML",
			Score:       0.95,
		},
		{
			ID:          "pod_1",
			Type:        "pod",
			Title:       "Machine Learning Course",
			Description: "Complete ML course materials",
			Score:       0.85,
		},
	}
	searchRepo.total = 2

	input := SearchInput{
		Query:   "machine learning",
		Page:    1,
		PerPage: 20,
	}

	result, err := svc.Search(ctx, input)
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	if len(result.Results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(result.Results))
	}
	if result.Total != 2 {
		t.Errorf("Expected total 2, got %d", result.Total)
	}
	if result.Page != 1 {
		t.Errorf("Expected page 1, got %d", result.Page)
	}
}

func TestSearch_WithFilters(t *testing.T) {
	svc, searchRepo, _, _, _, _ := newTestService()
	ctx := context.Background()

	searchRepo.results = []domain.SearchResult{
		{
			ID:    "mat_1",
			Type:  "material",
			Title: "Python Tutorial",
			Score: 0.9,
		},
	}
	searchRepo.total = 1

	input := SearchInput{
		Query:      "python",
		Types:      []string{"material"},
		Categories: []string{"programming"},
		FileTypes:  []string{"pdf"},
		Page:       1,
		PerPage:    10,
	}

	result, err := svc.Search(ctx, input)
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	if len(result.Results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(result.Results))
	}
}

func TestSearch_SavesHistory(t *testing.T) {
	svc, searchRepo, _, _, historyRepo, _ := newTestService()
	ctx := context.Background()

	searchRepo.results = []domain.SearchResult{}
	searchRepo.total = 0

	input := SearchInput{
		Query:   "test query",
		UserID:  "user_123",
		Page:    1,
		PerPage: 20,
	}

	_, err := svc.Search(ctx, input)
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	// Verify history was saved
	history := historyRepo.history["user_123"]
	if len(history) != 1 {
		t.Errorf("Expected 1 history entry, got %d", len(history))
	}
	if history[0].Query != "test query" {
		t.Errorf("Expected query 'test query', got '%s'", history[0].Query)
	}
}

func TestSearch_DefaultPagination(t *testing.T) {
	svc, searchRepo, _, _, _, _ := newTestService()
	ctx := context.Background()

	searchRepo.results = []domain.SearchResult{}
	searchRepo.total = 0

	// Test with invalid pagination values
	input := SearchInput{
		Query:   "test",
		Page:    0,  // Invalid, should default to 1
		PerPage: -1, // Invalid, should default to 20
	}

	result, err := svc.Search(ctx, input)
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	if result.Page != 1 {
		t.Errorf("Expected page 1, got %d", result.Page)
	}
	if result.PerPage != 20 {
		t.Errorf("Expected perPage 20, got %d", result.PerPage)
	}
}

func TestSearch_CalculatesTotalPages(t *testing.T) {
	svc, searchRepo, _, _, _, _ := newTestService()
	ctx := context.Background()

	searchRepo.results = []domain.SearchResult{}
	searchRepo.total = 45

	input := SearchInput{
		Query:   "test",
		Page:    1,
		PerPage: 10,
	}

	result, err := svc.Search(ctx, input)
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	if result.TotalPages != 5 {
		t.Errorf("Expected 5 total pages, got %d", result.TotalPages)
	}
}

// Test: Semantic Search
func TestSemanticSearch_Success(t *testing.T) {
	svc, _, _, vectorRepo, _, _ := newTestService()
	ctx := context.Background()

	// Setup embedding generator
	embeddingGen := newMockEmbeddingGenerator()
	svc.SetEmbeddingGenerator(embeddingGen)

	vectorRepo.results = []domain.SearchResult{
		{
			ID:          "mat_1",
			Type:        "material",
			Title:       "Deep Learning Fundamentals",
			Description: "Neural networks and deep learning",
			Score:       0.92,
		},
		{
			ID:          "mat_2",
			Type:        "material",
			Title:       "AI Concepts",
			Description: "Introduction to artificial intelligence",
			Score:       0.85,
		},
	}

	input := SemanticSearchInput{
		Query:    "how do neural networks work",
		Limit:    10,
		MinScore: 0.5,
	}

	results, err := svc.SemanticSearch(ctx, input)
	if err != nil {
		t.Fatalf("SemanticSearch failed: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results))
	}
}

func TestSemanticSearch_WithPodFilter(t *testing.T) {
	svc, _, _, vectorRepo, _, _ := newTestService()
	ctx := context.Background()

	embeddingGen := newMockEmbeddingGenerator()
	svc.SetEmbeddingGenerator(embeddingGen)

	vectorRepo.results = []domain.SearchResult{
		{
			ID:    "mat_1",
			Type:  "material",
			Title: "Pod-specific content",
			Score: 0.9,
		},
	}

	input := SemanticSearchInput{
		Query: "specific topic",
		PodID: "pod_123",
		Limit: 5,
	}

	results, err := svc.SemanticSearch(ctx, input)
	if err != nil {
		t.Fatalf("SemanticSearch failed: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(results))
	}
}

func TestSemanticSearch_NoEmbeddingGenerator(t *testing.T) {
	svc, _, _, vectorRepo, _, _ := newTestService()
	ctx := context.Background()

	// Don't set embedding generator
	vectorRepo.results = []domain.SearchResult{}

	input := SemanticSearchInput{
		Query: "test query",
		Limit: 10,
	}

	results, err := svc.SemanticSearch(ctx, input)
	if err != nil {
		t.Fatalf("SemanticSearch failed: %v", err)
	}

	// Should return empty results when no embedding generator
	if len(results) != 0 {
		t.Errorf("Expected 0 results without embedding generator, got %d", len(results))
	}
}

func TestSemanticSearch_DefaultLimit(t *testing.T) {
	svc, _, _, _, _, _ := newTestService()
	ctx := context.Background()

	embeddingGen := newMockEmbeddingGenerator()
	svc.SetEmbeddingGenerator(embeddingGen)

	input := SemanticSearchInput{
		Query: "test",
		Limit: 0, // Invalid, should default to 10
	}

	_, err := svc.SemanticSearch(ctx, input)
	if err != nil {
		t.Fatalf("SemanticSearch failed: %v", err)
	}
}

// Test: Autocomplete Suggestions
func TestGetSuggestions_Success(t *testing.T) {
	svc, searchRepo, _, _, _, _ := newTestService()
	ctx := context.Background()

	searchRepo.suggestions = []string{
		"machine learning",
		"machine vision",
		"mathematics",
		"marketing",
	}

	suggestions, err := svc.GetSuggestions(ctx, "mac", 10)
	if err != nil {
		t.Fatalf("GetSuggestions failed: %v", err)
	}

	if len(suggestions) != 2 {
		t.Errorf("Expected 2 suggestions, got %d", len(suggestions))
	}
}

func TestGetSuggestions_LimitResults(t *testing.T) {
	svc, searchRepo, _, _, _, _ := newTestService()
	ctx := context.Background()

	searchRepo.suggestions = []string{
		"python basics",
		"python advanced",
		"python web",
		"python data",
		"python ml",
	}

	suggestions, err := svc.GetSuggestions(ctx, "python", 3)
	if err != nil {
		t.Fatalf("GetSuggestions failed: %v", err)
	}

	if len(suggestions) != 3 {
		t.Errorf("Expected 3 suggestions, got %d", len(suggestions))
	}
}

func TestGetSuggestions_DefaultLimit(t *testing.T) {
	svc, _, _, _, _, _ := newTestService()
	ctx := context.Background()

	// Test with invalid limit
	_, err := svc.GetSuggestions(ctx, "test", 0)
	if err != nil {
		t.Fatalf("GetSuggestions failed: %v", err)
	}
}

// Test: Hybrid Search
func TestHybridSearch_Success(t *testing.T) {
	svc, searchRepo, _, vectorRepo, _, _ := newTestService()
	ctx := context.Background()

	embeddingGen := newMockEmbeddingGenerator()
	svc.SetEmbeddingGenerator(embeddingGen)

	// Setup keyword results
	searchRepo.results = []domain.SearchResult{
		{ID: "mat_1", Type: "material", Title: "Keyword Match 1", Score: 0.9},
		{ID: "mat_2", Type: "material", Title: "Keyword Match 2", Score: 0.8},
	}
	searchRepo.total = 2

	// Setup semantic results
	vectorRepo.results = []domain.SearchResult{
		{ID: "mat_1", Type: "material", Title: "Keyword Match 1", Score: 0.85},
		{ID: "mat_3", Type: "material", Title: "Semantic Match", Score: 0.75},
	}

	input := HybridSearchInput{
		Query:          "machine learning",
		SemanticWeight: 0.3,
		Page:           1,
		PerPage:        10,
	}

	result, err := svc.HybridSearch(ctx, input)
	if err != nil {
		t.Fatalf("HybridSearch failed: %v", err)
	}

	// Should have merged results
	if len(result.Results) == 0 {
		t.Error("Expected merged results")
	}
}

func TestHybridSearch_DefaultSemanticWeight(t *testing.T) {
	svc, searchRepo, _, _, _, _ := newTestService()
	ctx := context.Background()

	searchRepo.results = []domain.SearchResult{}
	searchRepo.total = 0

	input := HybridSearchInput{
		Query:          "test",
		SemanticWeight: -1, // Invalid, should default to 0.3
		Page:           1,
		PerPage:        10,
	}

	_, err := svc.HybridSearch(ctx, input)
	if err != nil {
		t.Fatalf("HybridSearch failed: %v", err)
	}
}

// Test: Trending Materials
func TestGetTrending_Success(t *testing.T) {
	svc, _, _, _, _, trendingRepo := newTestService()
	ctx := context.Background()

	trendingRepo.trending = []domain.TrendingMaterial{
		{
			MaterialDocument: domain.MaterialDocument{
				ID:            "mat_1",
				Title:         "Trending Material 1",
				ViewCount:     1000,
				DownloadCount: 500,
			},
			TrendingScore: 0.95,
		},
		{
			MaterialDocument: domain.MaterialDocument{
				ID:            "mat_2",
				Title:         "Trending Material 2",
				ViewCount:     800,
				DownloadCount: 400,
			},
			TrendingScore: 0.85,
		},
	}

	results, err := svc.GetTrending(ctx, "", 10)
	if err != nil {
		t.Fatalf("GetTrending failed: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("Expected 2 trending materials, got %d", len(results))
	}
}

func TestGetTrending_WithCategory(t *testing.T) {
	svc, _, _, _, _, trendingRepo := newTestService()
	ctx := context.Background()

	trendingRepo.trending = []domain.TrendingMaterial{
		{
			MaterialDocument: domain.MaterialDocument{
				ID:         "mat_1",
				Title:      "Programming Tutorial",
				Categories: []string{"programming"},
			},
			TrendingScore: 0.9,
		},
	}

	results, err := svc.GetTrending(ctx, "programming", 10)
	if err != nil {
		t.Fatalf("GetTrending failed: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("Expected 1 trending material, got %d", len(results))
	}
}

// Test: Popular Materials
func TestGetPopular_Success(t *testing.T) {
	svc, _, _, _, _, trendingRepo := newTestService()
	ctx := context.Background()

	trendingRepo.popular = []domain.MaterialDocument{
		{
			ID:            "mat_1",
			Title:         "Popular Material 1",
			ViewCount:     5000,
			DownloadCount: 2500,
			AverageRating: 4.8,
		},
		{
			ID:            "mat_2",
			Title:         "Popular Material 2",
			ViewCount:     4000,
			DownloadCount: 2000,
			AverageRating: 4.5,
		},
	}

	results, err := svc.GetPopular(ctx, "", 10)
	if err != nil {
		t.Fatalf("GetPopular failed: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("Expected 2 popular materials, got %d", len(results))
	}
}

// Test: Search History
func TestGetSearchHistory_Success(t *testing.T) {
	svc, _, _, _, historyRepo, _ := newTestService()
	ctx := context.Background()

	// Add some history
	historyRepo.history["user_123"] = []domain.SearchHistory{
		{ID: "h1", UserID: "user_123", Query: "python", CreatedAt: time.Now()},
		{ID: "h2", UserID: "user_123", Query: "machine learning", CreatedAt: time.Now()},
	}

	history, err := svc.GetSearchHistory(ctx, "user_123", 10)
	if err != nil {
		t.Fatalf("GetSearchHistory failed: %v", err)
	}

	if len(history) != 2 {
		t.Errorf("Expected 2 history entries, got %d", len(history))
	}
}

func TestClearSearchHistory_Success(t *testing.T) {
	svc, _, _, _, historyRepo, _ := newTestService()
	ctx := context.Background()

	// Add some history
	historyRepo.history["user_123"] = []domain.SearchHistory{
		{ID: "h1", UserID: "user_123", Query: "test"},
	}

	err := svc.ClearSearchHistory(ctx, "user_123")
	if err != nil {
		t.Fatalf("ClearSearchHistory failed: %v", err)
	}

	// Verify history was cleared
	if _, exists := historyRepo.history["user_123"]; exists {
		t.Error("Expected history to be cleared")
	}
}

// Test: Indexing Operations
func TestIndexPod_Success(t *testing.T) {
	svc, _, indexRepo, _, _, _ := newTestService()
	ctx := context.Background()

	pod := domain.PodDocument{
		ID:          "pod_123",
		OwnerID:     "user_1",
		Name:        "Test Pod",
		Slug:        "test-pod",
		Description: "A test knowledge pod",
		Categories:  []string{"education"},
		Visibility:  "public",
	}

	err := svc.IndexPod(ctx, pod)
	if err != nil {
		t.Fatalf("IndexPod failed: %v", err)
	}

	// Verify pod was indexed
	if _, exists := indexRepo.pods["pod_123"]; !exists {
		t.Error("Expected pod to be indexed")
	}
}

func TestIndexMaterial_Success(t *testing.T) {
	svc, _, indexRepo, _, _, _ := newTestService()
	ctx := context.Background()

	material := domain.MaterialDocument{
		ID:          "mat_123",
		PodID:       "pod_1",
		UploaderID:  "user_1",
		Title:       "Test Material",
		Description: "A test material",
		FileType:    "pdf",
		Status:      "ready",
	}

	err := svc.IndexMaterial(ctx, material)
	if err != nil {
		t.Fatalf("IndexMaterial failed: %v", err)
	}

	// Verify material was indexed
	if _, exists := indexRepo.materials["mat_123"]; !exists {
		t.Error("Expected material to be indexed")
	}
}

func TestDeletePodIndex_Success(t *testing.T) {
	svc, _, indexRepo, _, _, _ := newTestService()
	ctx := context.Background()

	// Add pod to index
	indexRepo.pods["pod_123"] = domain.PodDocument{ID: "pod_123", Name: "Test"}

	err := svc.DeletePodIndex(ctx, "pod_123")
	if err != nil {
		t.Fatalf("DeletePodIndex failed: %v", err)
	}

	// Verify pod was removed
	if _, exists := indexRepo.pods["pod_123"]; exists {
		t.Error("Expected pod to be removed from index")
	}
}

func TestDeleteMaterialIndex_Success(t *testing.T) {
	svc, _, indexRepo, _, _, _ := newTestService()
	ctx := context.Background()

	// Add material to index
	indexRepo.materials["mat_123"] = domain.MaterialDocument{ID: "mat_123", Title: "Test"}

	err := svc.DeleteMaterialIndex(ctx, "mat_123")
	if err != nil {
		t.Fatalf("DeleteMaterialIndex failed: %v", err)
	}

	// Verify material was removed
	if _, exists := indexRepo.materials["mat_123"]; exists {
		t.Error("Expected material to be removed from index")
	}
}

// Test: mergeResults helper function
func TestMergeResults_CombinesResults(t *testing.T) {
	svc, _, _, _, _, _ := newTestService()

	keywordResults := []domain.SearchResult{
		{ID: "mat_1", Title: "Result 1", Score: 0.9},
		{ID: "mat_2", Title: "Result 2", Score: 0.8},
	}

	semanticResults := []domain.SearchResult{
		{ID: "mat_1", Title: "Result 1", Score: 0.85},
		{ID: "mat_3", Title: "Result 3", Score: 0.75},
	}

	merged := svc.mergeResults(keywordResults, semanticResults, 0.3)

	// Should have 3 unique results
	if len(merged) != 3 {
		t.Errorf("Expected 3 merged results, got %d", len(merged))
	}

	// mat_1 should be first (appears in both)
	if merged[0].ID != "mat_1" {
		t.Errorf("Expected mat_1 to be first, got %s", merged[0].ID)
	}
}

func TestMergeResults_EmptySemanticResults(t *testing.T) {
	svc, _, _, _, _, _ := newTestService()

	keywordResults := []domain.SearchResult{
		{ID: "mat_1", Title: "Result 1", Score: 0.9},
	}

	merged := svc.mergeResults(keywordResults, nil, 0.3)

	if len(merged) != 1 {
		t.Errorf("Expected 1 result, got %d", len(merged))
	}
}

func TestMergeResults_EmptyKeywordResults(t *testing.T) {
	svc, _, _, _, _, _ := newTestService()

	semanticResults := []domain.SearchResult{
		{ID: "mat_1", Title: "Result 1", Score: 0.9},
	}

	merged := svc.mergeResults(nil, semanticResults, 0.3)

	if len(merged) != 1 {
		t.Errorf("Expected 1 result, got %d", len(merged))
	}
}

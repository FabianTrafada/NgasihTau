package mock

import (
	"context"
	"fmt"
	"sync"
	"time"

	"ngasihtau/internal/search/domain"
	"ngasihtau/internal/common/errors"
)

// MockSearchHistoryRepository implements SearchHistoryRepository in memory
type MockSearchHistoryRepository struct {
	mu      sync.RWMutex
	storage map[string][]domain.SearchHistory // userID -> history
}

// NewMockSearchHistoryRepository creates a new mock search history repository
func NewMockSearchHistoryRepository() *MockSearchHistoryRepository {
	return &MockSearchHistoryRepository{
		storage: make(map[string][]domain.SearchHistory),
	}
}

// SaveSearch saves a search query to history
func (r *MockSearchHistoryRepository) SaveSearch(ctx context.Context, userID, query string) error {
	if userID == "" || query == "" {
		return errors.BadRequest("userID and query are required")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	history := r.storage[userID]
	newEntry := domain.SearchHistory{
		ID:        fmt.Sprintf("hist_%d", len(history)+1),
		UserID:    userID,
		Query:     query,
		CreatedAt: time.Now(), // Need to add time import
	}
	
	// Keep only last 50 entries
	if len(history) >= 50 {
		history = append(history[1:], newEntry)
	} else {
		history = append(history, newEntry)
	}
	
	r.storage[userID] = history
	return nil
}

// GetHistory returns user's search history
func (r *MockSearchHistoryRepository) GetHistory(ctx context.Context, userID string, limit int) ([]domain.SearchHistory, error) {
	if userID == "" {
		return nil, errors.BadRequest("userID is required")
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	history := r.storage[userID]
	if limit < len(history) && limit > 0 {
		history = history[:limit]
	}
	
	return history, nil
}

// ClearHistory clears user's search history
func (r *MockSearchHistoryRepository) ClearHistory(ctx context.Context, userID string) error {
	if userID == "" {
		return errors.BadRequest("userID is required")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.storage, userID)
	return nil
}

// MockTrendingRepository implements TrendingRepository in memory
type MockTrendingRepository struct {
	mu sync.RWMutex
}

// NewMockTrendingRepository creates a new mock trending repository
func NewMockTrendingRepository() *MockTrendingRepository {
	return &MockTrendingRepository{}
}

// GetTrending returns trending materials
func (r *MockTrendingRepository) GetTrending(ctx context.Context, category string, limit int) ([]domain.TrendingMaterial, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Return mock trending materials
	return []domain.TrendingMaterial{
		{
			MaterialDocument: domain.MaterialDocument{
				ID:            "mat_1",
				PodID:         "pod_1",
				UploaderID:    "user_1",
				Title:         "Sample Material 1",
				Description:   "This is a trending material",
				FileType:      "pdf",
				Status:        "published",
				ViewCount:     1000,
				DownloadCount: 500,
				AverageRating: 4.5,
				RatingCount:   100,
				Categories:    []string{"education", "technology"},
				Tags:          []string{"programming", "tutorial"},
				CreatedAt:     time.Now().Unix(),
				UpdatedAt:     time.Now().Unix(),
			},
			TrendingScore: 0.95,
		},
	}, nil
}

// GetPopular returns popular materials
func (r *MockTrendingRepository) GetPopular(ctx context.Context, category string, limit int) ([]domain.MaterialDocument, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Return mock popular materials
	return []domain.MaterialDocument{
		{
			ID:            "mat_2",
			PodID:         "pod_2",
			UploaderID:    "user_2",
			Title:         "Popular Material",
			Description:   "This is a popular material",
			FileType:      "video",
			Status:        "published",
			ViewCount:     2000,
			DownloadCount: 1000,
			AverageRating: 4.8,
			RatingCount:   200,
			Categories:    []string{"entertainment"},
			Tags:          []string{"fun", "creative"},
			CreatedAt:     time.Now().Unix(),
			UpdatedAt:     time.Now().Unix(),
		},
	}, nil
}

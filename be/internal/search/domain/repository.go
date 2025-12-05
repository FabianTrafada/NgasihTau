package domain

import "context"

// SearchRepository handles full-text search operations
type SearchRepository interface {
	Search(ctx context.Context, query SearchQuery) ([]SearchResult, int64, error)

	GetSuggestions(ctx context.Context, prefix string, limit int) ([]string, error)
}

// EmbeddingGenerator generates vector embeddings for text
type EmbeddingGenerator interface {
	GenerateEmbedding(ctx context.Context, text string) ([]float32, error)
}

type IndexRepository interface {
	// Pod Indexing
	IndexPod(ctx context.Context, pod PodDocument) error
	UpdatePod(ctx context.Context, pod PodDocument) error
	DeletePod(ctx context.Context, podID string) error

	// Material Indexing
	IndexMaterial(ctx context.Context, material MaterialDocument) error
	UpdateMaterial(ctx context.Context, material MaterialDocument) error
	DeleteMaterial(ctx context.Context, materialID string) error
}

type VectorSearchRepository interface {
	SemanticSearch(ctx context.Context, query SemanticSearchQuery, embedding []float32) ([]SearchResult, error)
}

type SearchHistoryRepository interface {
	SaveSearch(ctx context.Context, userID, query string) error
	GetHistory(ctx context.Context, userID string, limit int) ([]SearchHistory, error)
	ClearHistory(ctx context.Context, userID string) error
}

type TrendingRepository interface {
	GetTrending(ctx context.Context, category string, limit int) ([]TrendingMaterial, error)

	GetPopular(ctx context.Context, category string, limit int) ([]MaterialDocument, error)
}

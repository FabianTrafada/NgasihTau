package domain

import "time"

type SearchResult struct {
	ID          string              `json:"id"`
	Type        string              `json:"type"` // "pod" atau "material"
	Title       string              `json:"title"`
	Description string              `json:"description"`
	Score       float64             `json:"score"`
	Highlights  map[string][]string `json:"highlights,omitempty"`
	Metadata    map[string]any      `json:"metadata,omitempty"`
}

type PodDocument struct {
	ID          string   `json:"id"`
	OwnerID     string   `json:"owner_id"`
	Name        string   `json:"name"`
	Slug        string   `json:"slug"`
	Description string   `json:"description"`
	Categories  []string `json:"categories"`
	Tags        []string `json:"tags"`
	Visibility  string   `json:"visibility"`
	StarCount   int      `json:"star_count"`
	ForkCount   int      `json:"fork_count"`
	ViewCount   int      `json:"view_count"`
	IsVerified  bool     `json:"is_verified"`  // True if created by teacher (Requirement 6.1)
	UpvoteCount int      `json:"upvote_count"` // Trust indicator (Requirement 6.2)
	TrustScore  float64  `json:"trust_score"`  // Computed: (0.6 * is_verified) + (0.4 * normalized_upvotes) (Requirement 6.4, 6.5)
	CreatedAt   int64    `json:"created_at"`
	UpdatedAt   int64    `json:"updated_at"`
}

type MaterialDocument struct {
	ID            string   `json:"id"`
	PodID         string   `json:"pod_id"`
	UploaderID    string   `json:"uploader_id"`
	Title         string   `json:"title"`
	Description   string   `json:"description"`
	FileType      string   `json:"file_type"`
	Status        string   `json:"status"`
	ViewCount     int      `json:"view_count"`
	DownloadCount int      `json:"download_count"`
	AverageRating float64  `json:"average_rating"`
	RatingCount   int      `json:"rating_count"`
	Categories    []string `json:"categories"` // inherited from pod
	Tags          []string `json:"tags"`       // inherited from pod
	CreatedAt     int64    `json:"created_at"`
	UpdatedAt     int64    `json:"updated_at"`
}

// SortBy represents the sorting option for search results
type SortBy string

const (
	SortByRelevance  SortBy = "relevance"   // Default Meilisearch relevance
	SortByUpvotes    SortBy = "upvotes"     // Sort by upvote_count descending
	SortByTrustScore SortBy = "trust_score" // Sort by combination of is_verified and upvote_count
	SortByRecent     SortBy = "recent"      // Sort by created_at descending
	SortByPopular    SortBy = "popular"     // Sort by view_count descending
)

type SearchQuery struct {
	Query      string   `json:"query"`
	Types      []string `json:"types,omitempty"`
	Categories []string `json:"categories,omitempty"`
	FileTypes  []string `json:"file_types,omitempty"`
	PodID      string   `json:"pod_id,omitempty"`
	Verified   *bool    `json:"verified,omitempty"` // Filter by verified status (teacher-created pods)
	SortBy     SortBy   `json:"sort_by,omitempty"`  // Sorting option (relevance, upvotes, trust_score, recent, popular)
	Page       int      `json:"page"`
	PerPage    int      `json:"per_page"`
}

type SemanticSearchQuery struct {
	Query    string  `json:"query"`
	PodID    string  `json:"pod_id,omitempty"`
	Limit    int     `json:"limit"`
	MinScore float64 `json:"min_score,omitempty"`
}

// HybridSearchQuery represents a hybrid search request combining keyword and semantic search
type HybridSearchQuery struct {
	Query          string   `json:"query"`
	Types          []string `json:"types,omitempty"`
	Categories     []string `json:"categories,omitempty"`
	FileTypes      []string `json:"file_types,omitempty"`
	PodID          string   `json:"pod_id,omitempty"`
	Verified       *bool    `json:"verified,omitempty"` // Filter by verified status (teacher-created pods)
	Page           int      `json:"page"`
	PerPage        int      `json:"per_page"`
	SemanticWeight float64  `json:"semantic_weight"` // 0.0 to 1.0, weight for semantic results
}

type SearchHistory struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Query     string    `json:"query"`
	CreatedAt time.Time `json:"created_at"`
}

type TrendingMaterial struct {
	MaterialDocument
	TrendingScore float64 `json:"trending_score"`
}

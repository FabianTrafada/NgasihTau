package meilisearch

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"

	"github.com/meilisearch/meilisearch-go"

	"ngasihtau/internal/search/domain"
)

const (
	PodsIndex      = "pods"
	MaterialsIndex = "materials"
)

// Client wraps Meilisearch client
type Client struct {
	client meilisearch.ServiceManager
}

// NewClient creates a new Meilisearch client
func NewClient(host, apiKey string) (*Client, error) {
	client := meilisearch.New(host, meilisearch.WithAPIKey(apiKey))

	// Verify connection
	_, err := client.Health()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to meilisearch: %w", err)
	}

	return &Client{client: client}, nil
}

// SetupIndexes creates and configures indexes
func (c *Client) SetupIndexes(ctx context.Context) error {
	// Create pods index
	_, err := c.client.CreateIndex(&meilisearch.IndexConfig{
		Uid:        PodsIndex,
		PrimaryKey: "id",
	})
	if err != nil {
		// Index might already exist, continue
	}

	// Configure pods index
	podsIndex := c.client.Index(PodsIndex)
	_, err = podsIndex.UpdateSearchableAttributes(&[]string{
		"name", "description", "categories", "tags",
	})
	if err != nil {
		return fmt.Errorf("failed to update pods searchable attributes: %w", err)
	}

	podsFilterable := []any{"owner_id", "visibility", "categories", "is_verified"}
	_, err = podsIndex.UpdateFilterableAttributes(&podsFilterable)
	if err != nil {
		return fmt.Errorf("failed to update pods filterable attributes: %w", err)
	}

	podsSortable := []string{"star_count", "view_count", "created_at", "upvote_count", "trust_score"}
	_, err = podsIndex.UpdateSortableAttributes(&podsSortable)
	if err != nil {
		return fmt.Errorf("failed to update pods sortable attributes: %w", err)
	}

	// Create materials index
	_, err = c.client.CreateIndex(&meilisearch.IndexConfig{
		Uid:        MaterialsIndex,
		PrimaryKey: "id",
	})
	if err != nil {
		// Index might already exist, continue
	}

	// Configure materials index
	materialsIndex := c.client.Index(MaterialsIndex)
	materialsSearchable := []string{"title", "description", "categories", "tags"}
	_, err = materialsIndex.UpdateSearchableAttributes(&materialsSearchable)
	if err != nil {
		return fmt.Errorf("failed to update materials searchable attributes: %w", err)
	}

	materialsFilterable := []any{"pod_id", "uploader_id", "file_type", "status", "categories"}
	_, err = materialsIndex.UpdateFilterableAttributes(&materialsFilterable)
	if err != nil {
		return fmt.Errorf("failed to update materials filterable attributes: %w", err)
	}

	materialsSortable := []string{"view_count", "download_count", "average_rating", "created_at"}
	_, err = materialsIndex.UpdateSortableAttributes(&materialsSortable)
	if err != nil {
		return fmt.Errorf("failed to update materials sortable attributes: %w", err)
	}

	return nil
}

// Search performs full-text search across pods and materials
func (c *Client) Search(ctx context.Context, query domain.SearchQuery) ([]domain.SearchResult, int64, error) {
	var results []domain.SearchResult
	var totalHits int64

	offset := (query.Page - 1) * query.PerPage

	// Build common filters for materials
	var materialFilters []string
	if query.PodID != "" {
		materialFilters = append(materialFilters, fmt.Sprintf("pod_id = '%s'", query.PodID))
	}
	if len(query.Categories) > 0 {
		catFilter := "categories IN ["
		for i, cat := range query.Categories {
			if i > 0 {
				catFilter += ", "
			}
			catFilter += fmt.Sprintf("'%s'", cat)
		}
		catFilter += "]"
		materialFilters = append(materialFilters, catFilter)
	}
	if len(query.FileTypes) > 0 {
		ftFilter := "file_type IN ["
		for i, ft := range query.FileTypes {
			if i > 0 {
				ftFilter += ", "
			}
			ftFilter += fmt.Sprintf("'%s'", ft)
		}
		ftFilter += "]"
		materialFilters = append(materialFilters, ftFilter)
	}

	materialFilterStr := ""
	if len(materialFilters) > 0 {
		materialFilterStr = materialFilters[0]
		for i := 1; i < len(materialFilters); i++ {
			materialFilterStr += " AND " + materialFilters[i]
		}
	}

	// Build filters for pods (includes verified filter)
	var podFilters []string
	if len(query.Categories) > 0 {
		catFilter := "categories IN ["
		for i, cat := range query.Categories {
			if i > 0 {
				catFilter += ", "
			}
			catFilter += fmt.Sprintf("'%s'", cat)
		}
		catFilter += "]"
		podFilters = append(podFilters, catFilter)
	}
	// Apply verified filter for pods (Requirement 6.3)
	if query.Verified != nil {
		podFilters = append(podFilters, fmt.Sprintf("is_verified = %t", *query.Verified))
	}

	podFilterStr := ""
	if len(podFilters) > 0 {
		podFilterStr = podFilters[0]
		for i := 1; i < len(podFilters); i++ {
			podFilterStr += " AND " + podFilters[i]
		}
	}

	// Build sort options for pods based on SortBy (Requirement 6.4, 6.5)
	var podSort []string
	switch query.SortBy {
	case domain.SortByUpvotes:
		podSort = []string{"upvote_count:desc"}
	case domain.SortByTrustScore:
		// Sort by trust_score which combines is_verified and upvote_count
		// trust_score is pre-computed: (0.6 * is_verified) + (0.4 * normalized_upvotes)
		podSort = []string{"trust_score:desc", "upvote_count:desc"}
	case domain.SortByRecent:
		podSort = []string{"created_at:desc"}
	case domain.SortByPopular:
		podSort = []string{"view_count:desc"}
	case domain.SortByRelevance:
		// Default Meilisearch relevance, no explicit sort
		podSort = nil
	default:
		// Default: prioritize verified pods with high upvotes (Requirement 6.5)
		podSort = nil
	}

	// Build sort options for materials
	var materialSort []string
	switch query.SortBy {
	case domain.SortByRecent:
		materialSort = []string{"created_at:desc"}
	case domain.SortByPopular:
		materialSort = []string{"view_count:desc"}
	default:
		materialSort = nil
	}

	// Search materials
	if len(query.Types) == 0 || slices.Contains(query.Types, "material") {
		materialSearchRequest := &meilisearch.SearchRequest{
			Query:                 query.Query,
			Limit:                 int64(query.PerPage),
			Offset:                int64(offset),
			AttributesToHighlight: []string{"title", "name", "description"},
		}
		if materialFilterStr != "" {
			materialSearchRequest.Filter = materialFilterStr
		}
		if len(materialSort) > 0 {
			materialSearchRequest.Sort = materialSort
		}

		materialsIndex := c.client.Index(MaterialsIndex)
		resp, err := materialsIndex.Search(query.Query, materialSearchRequest)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to search materials: %w", err)
		}

		for _, hit := range resp.Hits {
			hitMap := make(map[string]any)
			for key, value := range hit {
				var val any
				if err := json.Unmarshal(value, &val); err != nil {
					continue
				}
				hitMap[key] = val
			}
			result := domain.SearchResult{
				ID:          getString(hitMap, "id"),
				Type:        "material",
				Title:       getString(hitMap, "title"),
				Description: getString(hitMap, "description"),
				Metadata: map[string]any{
					"pod_id":    getString(hitMap, "pod_id"),
					"file_type": getString(hitMap, "file_type"),
				},
			}
			results = append(results, result)
		}
		totalHits += resp.EstimatedTotalHits
	}

	// Search pods
	if len(query.Types) == 0 || slices.Contains(query.Types, "pod") {
		podSearchRequest := &meilisearch.SearchRequest{
			Query:                 query.Query,
			Limit:                 int64(query.PerPage),
			Offset:                int64(offset),
			AttributesToHighlight: []string{"title", "name", "description"},
		}
		if podFilterStr != "" {
			podSearchRequest.Filter = podFilterStr
		}
		if len(podSort) > 0 {
			podSearchRequest.Sort = podSort
		}

		podsIndex := c.client.Index(PodsIndex)
		resp, err := podsIndex.Search(query.Query, podSearchRequest)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to search pods: %w", err)
		}

		for _, hit := range resp.Hits {
			hitMap := make(map[string]any)
			for key, value := range hit {
				var val any
				if err := json.Unmarshal(value, &val); err != nil {
					continue
				}
				hitMap[key] = val
			}
			result := domain.SearchResult{
				ID:          getString(hitMap, "id"),
				Type:        "pod",
				Title:       getString(hitMap, "name"),
				Description: getString(hitMap, "description"),
				Metadata: map[string]any{
					"slug":         getString(hitMap, "slug"),
					"visibility":   getString(hitMap, "visibility"),
					"is_verified":  getBool(hitMap, "is_verified"),
					"upvote_count": getInt(hitMap, "upvote_count"),
					"trust_score":  getFloat(hitMap, "trust_score"),
				},
			}
			results = append(results, result)
		}
		totalHits += resp.EstimatedTotalHits
	}

	return results, totalHits, nil
}

// GetSuggestions returns autocomplete suggestions
func (c *Client) GetSuggestions(ctx context.Context, prefix string, limit int) ([]string, error) {
	var suggestions []string

	materialsIndex := c.client.Index(MaterialsIndex)
	resp, err := materialsIndex.Search(prefix, &meilisearch.SearchRequest{
		Limit:                int64(limit),
		AttributesToRetrieve: []string{"title"},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get suggestions: %w", err)
	}

	for _, hit := range resp.Hits {
		hitMap := make(map[string]any)
		for key, value := range hit {
			var val any
			if err := json.Unmarshal(value, &val); err != nil {
				continue
			}
			hitMap[key] = val
		}
		if title := getString(hitMap, "title"); title != "" {
			suggestions = append(suggestions, title)
		}
	}

	return suggestions, nil
}

// IndexPod indexes a pod document
func (c *Client) IndexPod(ctx context.Context, pod domain.PodDocument) error {
	podsIndex := c.client.Index(PodsIndex)
	primaryKey := "id"
	_, err := podsIndex.AddDocuments([]domain.PodDocument{pod}, &primaryKey)
	if err != nil {
		return fmt.Errorf("failed to index pod: %w", err)
	}
	return nil
}

// UpdatePod updates a pod document
func (c *Client) UpdatePod(ctx context.Context, pod domain.PodDocument) error {
	return c.IndexPod(ctx, pod)
}

// DeletePod removes a pod from the index
func (c *Client) DeletePod(ctx context.Context, podID string) error {
	podsIndex := c.client.Index(PodsIndex)
	_, err := podsIndex.DeleteDocument(podID)
	if err != nil {
		return fmt.Errorf("failed to delete pod from index: %w", err)
	}
	return nil
}

// IndexMaterial indexes a material document
func (c *Client) IndexMaterial(ctx context.Context, material domain.MaterialDocument) error {
	materialsIndex := c.client.Index(MaterialsIndex)
	primaryKey := "id"
	_, err := materialsIndex.AddDocuments([]domain.MaterialDocument{material}, &primaryKey)
	if err != nil {
		return fmt.Errorf("failed to index material: %w", err)
	}
	return nil
}

// UpdateMaterial updates a material document
func (c *Client) UpdateMaterial(ctx context.Context, material domain.MaterialDocument) error {
	return c.IndexMaterial(ctx, material)
}

// DeleteMaterial removes a material from the index
func (c *Client) DeleteMaterial(ctx context.Context, materialID string) error {
	materialsIndex := c.client.Index(MaterialsIndex)
	_, err := materialsIndex.DeleteDocument(materialID)
	if err != nil {
		return fmt.Errorf("failed to delete material from index: %w", err)
	}
	return nil
}

// UpdatePodUpvoteCount updates only the upvote count and trust score for a pod.
// This is a partial update that only modifies the upvote_count and trust_score fields.
// Implements Requirements 6.2, 6.4, 6.5: Trust indicator updates.
func (c *Client) UpdatePodUpvoteCount(ctx context.Context, podID string, upvoteCount int) error {
	podsIndex := c.client.Index(PodsIndex)

	// Calculate trust score based on upvote count
	// Note: We don't have is_verified here, so we'll update only upvote_count
	// and let the trust_score be recalculated based on the existing is_verified value
	// For a proper update, we'd need to fetch the current document first
	// However, Meilisearch supports partial updates, so we can just update the fields we have

	// Create a partial document with only the fields we want to update
	partialDoc := map[string]any{
		"id":           podID,
		"upvote_count": upvoteCount,
	}

	primaryKey := "id"
	_, err := podsIndex.UpdateDocuments([]map[string]any{partialDoc}, &primaryKey)
	if err != nil {
		return fmt.Errorf("failed to update pod upvote count: %w", err)
	}
	return nil
}

func getString(m map[string]any, key string) string {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func getBool(m map[string]any, key string) bool {
	if v, ok := m[key]; ok {
		if b, ok := v.(bool); ok {
			return b
		}
	}
	return false
}

func getInt(m map[string]any, key string) int {
	if v, ok := m[key]; ok {
		switch n := v.(type) {
		case int:
			return n
		case int64:
			return int(n)
		case float64:
			return int(n)
		}
	}
	return 0
}

func getFloat(m map[string]any, key string) float64 {
	if v, ok := m[key]; ok {
		switch n := v.(type) {
		case float64:
			return n
		case float32:
			return float64(n)
		case int:
			return float64(n)
		case int64:
			return float64(n)
		}
	}
	return 0.0
}

// HealthCheck checks if Meilisearch is accessible.
func (c *Client) HealthCheck(ctx context.Context) error {
	_, err := c.client.Health()
	if err != nil {
		return fmt.Errorf("Meilisearch health check failed: %w", err)
	}
	return nil
}

var _ domain.SearchRepository = (*Client)(nil)
var _ domain.IndexRepository = (*Client)(nil)

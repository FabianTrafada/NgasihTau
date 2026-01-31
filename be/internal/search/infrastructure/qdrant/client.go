package qdrant

import (
	"context"
	"fmt"

	"github.com/qdrant/go-client/qdrant"

	"ngasihtau/internal/search/domain"
)

const (
	CollectionName = "material_chunks"
	VectorSize     = 1536
)

// Client wraps Qdrant client for semantic search
type Client struct {
	client *qdrant.Client
}

// NewClient creates a new Qdrant client
func NewClient(host string, port int) (*Client, error) {
	client, err := qdrant.NewClient(&qdrant.Config{
		Host: host,
		Port: port,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to qdrant: %w", err)
	}

	return &Client{client: client}, nil
}

// SemanticSearch performs vector similarity search
func (c *Client) SemanticSearch(ctx context.Context, query domain.SemanticSearchQuery, embedding []float32) ([]domain.SearchResult, error) {
	filter := &qdrant.Filter{}

	if query.PodID != "" {
		filter.Must = append(filter.Must, &qdrant.Condition{
			ConditionOneOf: &qdrant.Condition_Field{
				Field: &qdrant.FieldCondition{
					Key: "pod_id",
					Match: &qdrant.Match{
						MatchValue: &qdrant.Match_Keyword{
							Keyword: query.PodID,
						},
					},
				},
			},
		})
	}

	searchResult, err := c.client.Query(ctx, &qdrant.QueryPoints{
		CollectionName: CollectionName,
		Query:          qdrant.NewQuery(embedding...),
		Limit:          qdrant.PtrOf(uint64(query.Limit)),
		Filter:         filter,
		WithPayload:    qdrant.NewWithPayload(true),
		ScoreThreshold: qdrant.PtrOf(float32(query.MinScore)),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to perform semantic search: %w", err)
	}

	var results []domain.SearchResult
	for _, point := range searchResult {
		payload := point.Payload
		result := domain.SearchResult{
			ID:    getPayloadString(payload, "material_id"),
			Type:  "material",
			Title: getPayloadString(payload, "title"),
			Score: float64(point.Score),
			Metadata: map[string]interface{}{
				"pod_id":      getPayloadString(payload, "pod_id"),
				"chunk_index": getPayloadInt(payload, "chunk_index"),
				"text":        getPayloadString(payload, "text"),
			},
		}
		results = append(results, result)
	}

	return results, nil
}

// Close closes the client connection
func (c *Client) Close() error {
	return c.client.Close()
}

// Helper functions
func getPayloadString(payload map[string]*qdrant.Value, key string) string {
	if v, ok := payload[key]; ok {
		if sv := v.GetStringValue(); sv != "" {
			return sv
		}
	}
	return ""
}

func getPayloadInt(payload map[string]*qdrant.Value, key string) int64 {
	if v, ok := payload[key]; ok {
		return v.GetIntegerValue()
	}
	return 0
}

// HealthCheck checks if Qdrant is accessible.
func (c *Client) HealthCheck(ctx context.Context) error {
	_, err := c.client.ListCollections(ctx)
	if err != nil {
		return fmt.Errorf("Qdrant health check failed: %w", err)
	}
	return nil
}

// Ensure Client implements interface
var _ domain.VectorSearchRepository = (*Client)(nil)

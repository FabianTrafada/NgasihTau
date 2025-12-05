package openai

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/sashabaranov/go-openai"

	"ngasihtau/internal/search/domain"
)

var (
	ErrRateLimited        = errors.New("rate limited by OpenAI API")
	ErrInvalidAPIKey      = errors.New("invalid API key")
	ErrServiceUnavailable = errors.New("OpenAI service unavailable")
)

// Config holds OpenAI client configuration
type Config struct {
	APIKey         string
	EmbeddingModel string
	MaxRetries     int
	RetryDelay     time.Duration
}

// EmbeddingClient wraps OpenAI client for embedding generation
type EmbeddingClient struct {
	client         *openai.Client
	embeddingModel string
	maxRetries     int
	retryDelay     time.Duration
}

// NewEmbeddingClient creates a new OpenAI embedding client
func NewEmbeddingClient(cfg Config) *EmbeddingClient {
	embeddingModel := cfg.EmbeddingModel
	if embeddingModel == "" {
		embeddingModel = "text-embedding-3-small"
	}

	maxRetries := cfg.MaxRetries
	if maxRetries == 0 {
		maxRetries = 3
	}

	retryDelay := cfg.RetryDelay
	if retryDelay == 0 {
		retryDelay = 1 * time.Second
	}

	return &EmbeddingClient{
		client:         openai.NewClient(cfg.APIKey),
		embeddingModel: embeddingModel,
		maxRetries:     maxRetries,
		retryDelay:     retryDelay,
	}
}

// GenerateEmbedding generates a vector embedding for the given text
func (c *EmbeddingClient) GenerateEmbedding(ctx context.Context, text string) ([]float32, error) {
	var embedding []float32

	err := c.retryWithBackoff(ctx, func() error {
		resp, err := c.client.CreateEmbeddings(ctx, openai.EmbeddingRequest{
			Model: openai.EmbeddingModel(c.embeddingModel),
			Input: []string{text},
		})
		if err != nil {
			return err
		}

		if len(resp.Data) == 0 {
			return fmt.Errorf("no embedding returned")
		}

		embedding = resp.Data[0].Embedding
		return nil
	})

	if err != nil {
		return nil, wrapOpenAIError(err)
	}

	return embedding, nil
}

func (c *EmbeddingClient) retryWithBackoff(ctx context.Context, fn func() error) error {
	var lastErr error
	delay := c.retryDelay

	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(delay):
				delay *= 2 // exponential backoff
				if delay > 30*time.Second {
					delay = 30 * time.Second
				}
			}
		}

		lastErr = fn()
		if lastErr == nil {
			return nil
		}

		if !isRetryableError(lastErr) {
			return lastErr
		}
	}

	return fmt.Errorf("max retries exceeded: %w", lastErr)
}

func isRetryableError(err error) bool {
	if err == nil {
		return false
	}

	var apiErr *openai.APIError
	if errors.As(err, &apiErr) {
		switch apiErr.HTTPStatusCode {
		case 429:
			return true
		case 500, 502, 503, 504:
			return true
		}
	}

	if errors.Is(err, context.DeadlineExceeded) {
		return true
	}

	return false
}

func wrapOpenAIError(err error) error {
	if err == nil {
		return nil
	}

	var apiErr *openai.APIError
	if errors.As(err, &apiErr) {
		switch apiErr.HTTPStatusCode {
		case 401:
			return fmt.Errorf("%w: %s", ErrInvalidAPIKey, apiErr.Message)
		case 429:
			return fmt.Errorf("%w: %s", ErrRateLimited, apiErr.Message)
		case 500, 502, 503, 504:
			return fmt.Errorf("%w: %s", ErrServiceUnavailable, apiErr.Message)
		}
	}

	return err
}

// Ensure EmbeddingClient implements EmbeddingGenerator interface
var _ domain.EmbeddingGenerator = (*EmbeddingClient)(nil)

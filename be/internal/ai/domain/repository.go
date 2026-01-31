package domain

import (
	"context"

	"github.com/google/uuid"
)

type ChatSessionRepository interface {
	Create(ctx context.Context, session *ChatSession) error

	FindByID(ctx context.Context, id uuid.UUID) (*ChatSession, error)

	FindByUserAndMaterial(ctx context.Context, userID, materialID uuid.UUID) (*ChatSession, error)

	FindByUserAndPod(ctx context.Context, userID, podID uuid.UUID) (*ChatSession, error)

	Update(ctx context.Context, session *ChatSession) error

	Delete(ctx context.Context, id uuid.UUID) error
}

type ChatMessageRepository interface {
	Create(ctx context.Context, message *ChatMessage) error

	FindByID(ctx context.Context, id uuid.UUID) (*ChatMessage, error)

	FindBySessionID(ctx context.Context, sessionID uuid.UUID, limit, offset int) ([]ChatMessage, error)

	UpdateFeedback(ctx context.Context, id uuid.UUID, feedback FeedbackType, feedbackText *string) error

	CountBySessionID(ctx context.Context, sessionID uuid.UUID) (int, error)
}

type VectorRepository interface {
	Upsert(ctx context.Context, chunks []MaterialChunk) error

	Search(ctx context.Context, embedding []float32, materialID *uuid.UUID, podID *uuid.UUID, limit int) ([]MaterialChunk, error)

	DeleteByMaterialID(ctx context.Context, materialID uuid.UUID) error

	DeleteByPodID(ctx context.Context, podID uuid.UUID) error
}

// EmbeddingService defines the interface for generating embeddings.
type EmbeddingService interface {
	// GenerateEmbedding generates an embedding for the given text.
	GenerateEmbedding(ctx context.Context, text string) ([]float32, error)

	// GenerateEmbeddings generates embeddings for multiple texts.
	GenerateEmbeddings(ctx context.Context, texts []string) ([][]float32, error)
}

// ChatCompletionService defines the interface for chat completion.
type ChatCompletionService interface {
	// GenerateResponse generates a response based on context and query.
	GenerateResponse(ctx context.Context, systemPrompt, userQuery string, context []string) (string, error)

	// GenerateSuggestions generates suggested questions based on content.
	GenerateSuggestions(ctx context.Context, content string, existingQuestions []string) ([]string, error)
}

// TextExtractor defines the interface for extracting text from files.
type TextExtractor interface {
	// ExtractText extracts text from a file URL.
	ExtractText(ctx context.Context, fileURL, fileType string) (string, map[string]interface{}, error)
}

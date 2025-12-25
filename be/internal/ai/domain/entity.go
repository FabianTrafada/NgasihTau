package domain

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type ChatMode string
type MessageRole string
type FeedbackType string
type ExportFormat string

const (
	// ChatModeMaterial is for single material context.
	ChatModeMaterial ChatMode = "material"

	// ChatModePod is for pod-wide context.
	ChatModePod ChatMode = "pod"
)

const (
	MessageRoleUser      MessageRole = "user"
	MessageRoleAssistant MessageRole = "assistant"
)

const (
	FeedbackThumbsUp   FeedbackType = "thumbs_up"
	FeedbackThumbsDown FeedbackType = "thumbs_down"
)

const (
	ExportFormatPDF      ExportFormat = "pdf"
	ExportFormatMarkdown ExportFormat = "markdown"
)

type ChatSession struct {
	ID         uuid.UUID  `json:"id"`
	UserID     uuid.UUID  `json:"user_id"`
	MaterialID *uuid.UUID `json:"material_id,omitempty"`
	PodID      *uuid.UUID `json:"pod_id,omitempty"`
	Mode       ChatMode   `json:"mode"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}

type ChatMessage struct {
	ID           uuid.UUID     `json:"id"`
	SessionID    uuid.UUID     `json:"session_id"`
	Role         MessageRole   `json:"role"`
	Content      string        `json:"content"`
	Sources      []ChunkSource `json:"sources,omitempty"`
	Feedback     *FeedbackType `json:"feedback,omitempty"`
	FeedbackText *string       `json:"feedback_text,omitempty"`
	CreatedAt    time.Time     `json:"created_at"`
}

type ChunkSource struct {
	MaterialID uuid.UUID `json:"material_id"`
	ChunkIndex int       `json:"chunk_index"`
	Text       string    `json:"text"`
	Score      float64   `json:"score"`
}

type Sources []ChunkSource

type ChatExport struct {
	Format      ExportFormat  `json:"format"`
	Content     []byte        `json:"-"`
	Filename    string        `json:"filename"`
	ContentType string        `json:"content_type"`
	Session     *ChatSession  `json:"session,omitempty"`
	Messages    []ChatMessage `json:"messages,omitempty"`
}

func (s Sources) Value() (driver.Value, error) {
	if s == nil {
		return nil, nil
	}
	return json.Marshal(s)
}

func (s *Sources) Scan(value interface{}) error {
	if value == nil {
		*s = nil
		return nil
	}

	var data []byte
	switch v := value.(type) {
	case []byte:
		data = v
	case string:
		data = []byte(v)
	default:
		return fmt.Errorf("cannot scan type %T into Sources", value)
	}

	return json.Unmarshal(data, s)
}

type MaterialChunk struct {
	ID         string            `json:"id"`
	MaterialID uuid.UUID         `json:"material_id"`
	PodID      uuid.UUID         `json:"pod_id"`
	ChunkIndex int               `json:"chunk_index"`
	Text       string            `json:"text"`
	Embedding  []float32         `json:"embedding"`
	Metadata   map[string]string `json:"metadata,omitempty"`
}

type ChatHistory struct {
	Session  ChatSession   `json:"session"`
	Messages []ChatMessage `json:"messages"`
}

type SuggestedQuestion struct {
	Question string `json:"question"`
}

func NewChatSession(userID uuid.UUID, materialID, podID *uuid.UUID, mode ChatMode) *ChatSession {
	now := time.Now()
	return &ChatSession{
		ID:         uuid.New(),
		UserID:     userID,
		MaterialID: materialID,
		PodID:      podID,
		Mode:       mode,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
}

func NewChatMessage(sessionID uuid.UUID, role MessageRole, content string, sources []ChunkSource) *ChatMessage {
	return &ChatMessage{
		ID:        uuid.New(),
		SessionID: sessionID,
		Role:      role,
		Content:   content,
		Sources:   sources,
		CreatedAt: time.Now(),
	}
}

func NewMaterialChunk(materialID, podID uuid.UUID, chunkIndex int, text string, embedding []float32, metadata map[string]string) *MaterialChunk {
	return &MaterialChunk{
		ID:         uuid.New().String(),
		MaterialID: materialID,
		PodID:      podID,
		ChunkIndex: chunkIndex,
		Text:       text,
		Embedding:  embedding,
		Metadata:   metadata,
	}
}

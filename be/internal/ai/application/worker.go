package application

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"ngasihtau/internal/ai/domain"
	natspkg "ngasihtau/pkg/nats"
)

type Worker struct {
	service          *Service
	natsClient       *natspkg.Client
	fileProcessorURL string
	httpClient       *http.Client
}

func NewWorker(service *Service, natsClient *natspkg.Client, fileProcessorURL string) *Worker {
	return &Worker{
		service:          service,
		natsClient:       natsClient,
		fileProcessorURL: fileProcessorURL,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

func (w *Worker) Start(ctx context.Context) error {
	if w.natsClient == nil {
		log.Warn().Msg("NATS client not configured, worker not started")
		return nil
	}

	if err := w.natsClient.EnsureStream(ctx, natspkg.StreamMaterial, natspkg.StreamSubjects[natspkg.StreamMaterial]); err != nil {
		return fmt.Errorf("failed to ensure material stream: %w", err)
	}

	cfg := natspkg.DefaultSubscribeConfig(
		natspkg.StreamMaterial,
		"ai-service-material-processor",
		[]string{natspkg.SubjectMaterialUploaded},
	)
	cfg.MaxDeliver = 3
	cfg.AckWait = 120 * time.Second

	_, err := w.natsClient.Subscribe(ctx, cfg, w.handleMaterialUploaded)
	if err != nil {
		return fmt.Errorf("failed to subscribe to material.uploaded: %w", err)
	}

	deleteCfg := natspkg.DefaultSubscribeConfig(
		natspkg.StreamMaterial,
		"ai-service-material-deleter",
		[]string{natspkg.SubjectMaterialDeleted},
	)
	deleteCfg.MaxDeliver = 3
	deleteCfg.AckWait = 30 * time.Second

	_, err = w.natsClient.Subscribe(ctx, deleteCfg, w.handleMaterialDeleted)
	if err != nil {
		return fmt.Errorf("failed to subscribe to material.deleted: %w", err)
	}

	log.Info().Msg("material processing worker started")
	log.Info().Msg("material deletion worker started")
	return nil
}

func (w *Worker) handleMaterialUploaded(ctx context.Context, event natspkg.CloudEvent) error {
	var data natspkg.MaterialUploadedEvent
	if err := json.Unmarshal(event.Data, &data); err != nil {
		log.Error().Err(err).Msg("failed to parse material.uploaded event")
		return err
	}

	log.Info().
		Str("material_id", data.MaterialID.String()).
		Str("pod_id", data.PodID.String()).
		Str("file_type", data.FileType).
		Msg("processing material")

	err := w.processMaterial(ctx, data)
	if err != nil {
		log.Error().Err(err).
			Str("material_id", data.MaterialID.String()).
			Msg("failed to process material")

		w.publishProcessedEvent(ctx, data.MaterialID, "processing_failed", 0, 0, err.Error(), nil)
		return nil
	}

	return nil
}

func (w *Worker) handleMaterialDeleted(ctx context.Context, event natspkg.CloudEvent) error {
	var data natspkg.MaterialDeletedEvent
	if err := json.Unmarshal(event.Data, &data); err != nil {
		log.Error().Err(err).Msg("failed to parse material.deleted event")
		return err
	}

	log.Info().
		Str("material_id", data.MaterialID.String()).
		Str("pod_id", data.PodID.String()).
		Msg("deleting material chunks from Qdrant")

	if err := w.service.vectorRepo.DeleteByMaterialID(ctx, data.MaterialID); err != nil {
		log.Error().Err(err).
			Str("material_id", data.MaterialID.String()).
			Msg("failed to delete material chunks from Qdrant")
		return err
	}

	log.Info().
		Str("material_id", data.MaterialID.String()).
		Msg("successfully deleted material chunks from Qdrant")

	return nil
}

func (w *Worker) processMaterial(ctx context.Context, data natspkg.MaterialUploadedEvent) error {
	text, metadata, err := w.extractText(ctx, data.FileURL, data.FileType)
	if err != nil {
		return fmt.Errorf("failed to extract text: %w", err)
	}

	if text == "" {
		return fmt.Errorf("no text extracted from file")
	}

	chunker := NewChunker(DefaultChunkerConfig())
	chunks := chunker.ChunkText(text)

	if len(chunks) == 0 {
		return fmt.Errorf("no chunks generated from text")
	}

	log.Info().
		Str("material_id", data.MaterialID.String()).
		Int("chunk_count", len(chunks)).
		Msg("generated chunks")

	var texts []string
	for _, chunk := range chunks {
		texts = append(texts, chunk.Text)
	}

	embeddings, err := w.service.embeddingClient.GenerateEmbeddings(ctx, texts)
	if err != nil {
		return fmt.Errorf("failed to generate embeddings: %w", err)
	}

	// Debug: log embedding dimensions
	if len(embeddings) > 0 {
		log.Debug().
			Int("embedding_count", len(embeddings)).
			Int("embedding_dimension", len(embeddings[0])).
			Msg("generated embeddings")
	}

	var materialChunks []domain.MaterialChunk
	for i, chunk := range chunks {
		materialChunks = append(materialChunks, domain.MaterialChunk{
			ID:         uuid.New().String(),
			MaterialID: data.MaterialID,
			PodID:      data.PodID,
			ChunkIndex: chunk.Index,
			Text:       chunk.Text,
			Embedding:  embeddings[i],
		})
	}

	if err := w.service.vectorRepo.Upsert(ctx, materialChunks); err != nil {
		return fmt.Errorf("failed to store embeddings: %w", err)
	}

	log.Info().
		Str("material_id", data.MaterialID.String()).
		Int("chunk_count", len(materialChunks)).
		Msg("stored embedding in Qdrant")

	wordCount := 0
	if wc, ok := metadata["word_count"].(float64); ok {
		wordCount = int(wc)
	}

	// Build material info for search indexing from the uploaded event
	info := &MaterialInfo{
		PodID:       data.PodID,
		UploaderID:  data.UploaderID,
		Title:       data.Title,
		Description: data.Description,
		FileType:    data.FileType,
		Categories:  data.Categories,
		Tags:        data.Tags,
	}

	w.publishProcessedEvent(ctx, data.MaterialID, "ready", len(materialChunks), wordCount, "", info)

	return nil
}

type ExtractRequest struct {
	FileURL  string `json:"file_url"`
	FileType string `json:"file_type"`
}

type ExtractResponse struct {
	Text     string                 `json:"text"`
	Metadata map[string]interface{} `json:"metadata"`
}

func (w *Worker) extractText(ctx context.Context, fileURL, fileType string) (string, map[string]interface{}, error) {
	reqBody := ExtractRequest{
		FileURL:  fileURL,
		FileType: fileType,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return "", nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/extract", w.fileProcessorURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(jsonBody))
	if err != nil {
		return "", nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := w.httpClient.Do(req)
	if err != nil {
		return "", nil, fmt.Errorf("failed to call file processor: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", nil, fmt.Errorf("file processor returned status %d: %s", resp.StatusCode, string(body))
	}

	var extractResp ExtractResponse
	if err := json.NewDecoder(resp.Body).Decode(&extractResp); err != nil {
		return "", nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return extractResp.Text, extractResp.Metadata, nil
}

// MaterialInfo contains material information for event publishing.
type MaterialInfo struct {
	PodID       uuid.UUID
	UploaderID  uuid.UUID
	Title       string
	Description string
	FileType    string
	Categories  []string
	Tags        []string
}

func (w *Worker) publishProcessedEvent(ctx context.Context, materialID uuid.UUID, status string, chunkCount, wordCount int, errMsg string, info *MaterialInfo) {
	if w.natsClient == nil {
		return
	}

	event := natspkg.MaterialProcessedEvent{
		MaterialID: materialID,
		Status:     status,
		ChunkCount: chunkCount,
		WordCount:  wordCount,
		Error:      errMsg,
	}

	// Include material info if available (for search indexing)
	if info != nil {
		event.PodID = info.PodID
		event.UploaderID = info.UploaderID
		event.Title = info.Title
		event.Description = info.Description
		event.FileType = info.FileType
		event.Categories = info.Categories
		event.Tags = info.Tags
	}

	if err := w.natsClient.Publish(ctx, natspkg.SubjectMaterialProcessed, "material.processed", event); err != nil {
		log.Error().Err(err).
			Str("material_id", materialID.String()).
			Msg("failed to publish material.processed event")
	}
}

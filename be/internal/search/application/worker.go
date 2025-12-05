// Package application contains the business logic for the search service.
package application

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/rs/zerolog/log"

	"ngasihtau/internal/search/domain"
	natspkg "ngasihtau/pkg/nats"
)

// Worker handles background event processing for search indexing.
// Implements requirement 7: Material Search.
type Worker struct {
	service    *Service
	natsClient *natspkg.Client
}

// NewWorker creates a new search indexing worker.
func NewWorker(service *Service, natsClient *natspkg.Client) *Worker {
	return &Worker{
		service:    service,
		natsClient: natsClient,
	}
}

// Start begins listening for events and processing them.
func (w *Worker) Start(ctx context.Context) error {
	if w.natsClient == nil {
		log.Warn().Msg("NATS client not configured, search indexing worker not started")
		return nil
	}

	// Ensure streams exist
	if err := w.natsClient.EnsureStream(ctx, natspkg.StreamPod, natspkg.StreamSubjects[natspkg.StreamPod]); err != nil {
		return fmt.Errorf("failed to ensure pod stream: %w", err)
	}

	if err := w.natsClient.EnsureStream(ctx, natspkg.StreamMaterial, natspkg.StreamSubjects[natspkg.StreamMaterial]); err != nil {
		return fmt.Errorf("failed to ensure material stream: %w", err)
	}

	// Subscribe to pod events
	podCfg := natspkg.DefaultSubscribeConfig(
		natspkg.StreamPod,
		"search-service-pod-indexer",
		[]string{natspkg.SubjectPodCreated, natspkg.SubjectPodUpdated},
	)
	podCfg.MaxDeliver = 3
	podCfg.AckWait = 30 * time.Second

	_, err := w.natsClient.Subscribe(ctx, podCfg, w.handlePodEvent)
	if err != nil {
		return fmt.Errorf("failed to subscribe to pod events: %w", err)
	}
	log.Info().Msg("subscribed to pod.created and pod.updated events")

	// Subscribe to material.processed events
	materialCfg := natspkg.DefaultSubscribeConfig(
		natspkg.StreamMaterial,
		"search-service-material-indexer",
		[]string{natspkg.SubjectMaterialProcessed},
	)
	materialCfg.MaxDeliver = 3
	materialCfg.AckWait = 30 * time.Second

	_, err = w.natsClient.Subscribe(ctx, materialCfg, w.handleMaterialProcessed)
	if err != nil {
		return fmt.Errorf("failed to subscribe to material.processed events: %w", err)
	}
	log.Info().Msg("subscribed to material.processed events")

	// Subscribe to material.deleted events for cleanup
	deleteCfg := natspkg.DefaultSubscribeConfig(
		natspkg.StreamMaterial,
		"search-service-material-deleter",
		[]string{natspkg.SubjectMaterialDeleted},
	)
	deleteCfg.MaxDeliver = 3
	deleteCfg.AckWait = 30 * time.Second

	_, err = w.natsClient.Subscribe(ctx, deleteCfg, w.handleMaterialDeleted)
	if err != nil {
		return fmt.Errorf("failed to subscribe to material.deleted events: %w", err)
	}
	log.Info().Msg("subscribed to material.deleted events")

	log.Info().Msg("search indexing worker started")
	return nil
}

// handlePodEvent handles pod.created and pod.updated events.
func (w *Worker) handlePodEvent(ctx context.Context, event natspkg.CloudEvent) error {
	switch event.Type {
	case "pod.created":
		return w.handlePodCreated(ctx, event)
	case "pod.updated":
		return w.handlePodUpdated(ctx, event)
	default:
		log.Warn().Str("type", event.Type).Msg("unknown pod event type")
		return nil
	}
}

// handlePodCreated handles pod.created events.
func (w *Worker) handlePodCreated(ctx context.Context, event natspkg.CloudEvent) error {
	var data natspkg.PodCreatedEvent
	if err := json.Unmarshal(event.Data, &data); err != nil {
		log.Error().Err(err).Msg("failed to parse pod.created event")
		return err
	}

	log.Info().
		Str("pod_id", data.PodID.String()).
		Str("name", data.Name).
		Str("slug", data.Slug).
		Str("visibility", data.Visibility).
		Msg("indexing new pod")

	// Create pod document for indexing with full data from event
	podDoc := domain.PodDocument{
		ID:          data.PodID.String(),
		OwnerID:     data.OwnerID.String(),
		Name:        data.Name,
		Slug:        data.Slug,
		Description: data.Description,
		Visibility:  data.Visibility,
		Categories:  data.Categories,
		Tags:        data.Tags,
		StarCount:   0,
		ForkCount:   0,
		ViewCount:   0,
		CreatedAt:   event.Time.Unix(),
		UpdatedAt:   event.Time.Unix(),
	}

	if err := w.service.IndexPod(ctx, podDoc); err != nil {
		log.Error().Err(err).
			Str("pod_id", data.PodID.String()).
			Msg("failed to index pod")
		return err
	}

	log.Info().
		Str("pod_id", data.PodID.String()).
		Msg("successfully indexed pod")

	return nil
}

// handlePodUpdated handles pod.updated events.
func (w *Worker) handlePodUpdated(ctx context.Context, event natspkg.CloudEvent) error {
	var data natspkg.PodUpdatedEvent
	if err := json.Unmarshal(event.Data, &data); err != nil {
		log.Error().Err(err).Msg("failed to parse pod.updated event")
		return err
	}

	log.Info().
		Str("pod_id", data.PodID.String()).
		Str("name", data.Name).
		Str("visibility", data.Visibility).
		Msg("re-indexing updated pod")

	// Create pod document for re-indexing with full data from event
	podDoc := domain.PodDocument{
		ID:          data.PodID.String(),
		OwnerID:     data.OwnerID.String(),
		Name:        data.Name,
		Slug:        data.Slug,
		Description: data.Description,
		Visibility:  data.Visibility,
		Categories:  data.Categories,
		Tags:        data.Tags,
		StarCount:   data.StarCount,
		ForkCount:   data.ForkCount,
		ViewCount:   data.ViewCount,
		UpdatedAt:   event.Time.Unix(),
	}

	if err := w.service.IndexPod(ctx, podDoc); err != nil {
		log.Error().Err(err).
			Str("pod_id", data.PodID.String()).
			Msg("failed to re-index pod")
		return err
	}

	log.Info().
		Str("pod_id", data.PodID.String()).
		Msg("successfully re-indexed pod")

	return nil
}

// handleMaterialProcessed handles material.processed events.
func (w *Worker) handleMaterialProcessed(ctx context.Context, event natspkg.CloudEvent) error {
	var data natspkg.MaterialProcessedEvent
	if err := json.Unmarshal(event.Data, &data); err != nil {
		log.Error().Err(err).Msg("failed to parse material.processed event")
		return err
	}

	log.Info().
		Str("material_id", data.MaterialID.String()).
		Str("status", data.Status).
		Int("chunk_count", data.ChunkCount).
		Int("word_count", data.WordCount).
		Msg("received material.processed event")

	// Only index materials that are ready
	if data.Status != "ready" {
		log.Info().
			Str("material_id", data.MaterialID.String()).
			Str("status", data.Status).
			Msg("skipping indexing for non-ready material")
		return nil
	}

	log.Info().
		Str("material_id", data.MaterialID.String()).
		Str("title", data.Title).
		Str("file_type", data.FileType).
		Msg("indexing processed material")

	// Create material document for indexing with full data from event
	materialDoc := domain.MaterialDocument{
		ID:          data.MaterialID.String(),
		PodID:       data.PodID.String(),
		UploaderID:  data.UploaderID.String(),
		Title:       data.Title,
		Description: data.Description,
		FileType:    data.FileType,
		Status:      data.Status,
		Categories:  data.Categories,
		Tags:        data.Tags,
		CreatedAt:   event.Time.Unix(),
		UpdatedAt:   event.Time.Unix(),
	}

	if err := w.service.IndexMaterial(ctx, materialDoc); err != nil {
		log.Error().Err(err).
			Str("material_id", data.MaterialID.String()).
			Msg("failed to index material")
		return err
	}

	log.Info().
		Str("material_id", data.MaterialID.String()).
		Msg("successfully indexed material")

	return nil
}

// handleMaterialDeleted handles material.deleted events.
func (w *Worker) handleMaterialDeleted(ctx context.Context, event natspkg.CloudEvent) error {
	var data natspkg.MaterialDeletedEvent
	if err := json.Unmarshal(event.Data, &data); err != nil {
		log.Error().Err(err).Msg("failed to parse material.deleted event")
		return err
	}

	log.Info().
		Str("material_id", data.MaterialID.String()).
		Str("pod_id", data.PodID.String()).
		Msg("removing material from search index")

	if err := w.service.DeleteMaterialIndex(ctx, data.MaterialID.String()); err != nil {
		log.Error().Err(err).
			Str("material_id", data.MaterialID.String()).
			Msg("failed to delete material from index")
		return err
	}

	log.Info().
		Str("material_id", data.MaterialID.String()).
		Msg("successfully removed material from search index")

	return nil
}

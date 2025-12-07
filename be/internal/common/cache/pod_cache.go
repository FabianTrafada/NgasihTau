// Package cache provides cached repository implementations.
// Implements requirement 24: Cache-aside pattern for pod details.
package cache

import (
	"context"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"ngasihtau/internal/pod/domain"
)

// PodCache provides cached access to pod data.
type PodCache struct {
	cache *Client
}

// NewPodCache creates a new pod cache.
func NewPodCache(cache *Client) *PodCache {
	return &PodCache{cache: cache}
}

// GetPod retrieves a pod from cache or calls the loader if not cached.
// Implements cache-aside pattern with 30-minute TTL.
func (c *PodCache) GetPod(ctx context.Context, podID uuid.UUID, loader func() (*domain.Pod, error)) (*domain.Pod, error) {
	key := PodDetailsKey(podID.String())

	// Try cache first
	var pod domain.Pod
	found, err := c.cache.Get(ctx, key, &pod)
	if err != nil {
		log.Warn().Err(err).Str("pod_id", podID.String()).Msg("cache get failed for pod")
	}
	if found {
		log.Debug().Str("pod_id", podID.String()).Msg("pod cache hit")
		return &pod, nil
	}

	// Cache miss - call loader
	log.Debug().Str("pod_id", podID.String()).Msg("pod cache miss")
	result, err := loader()
	if err != nil {
		return nil, err
	}

	// Store in cache
	if cacheErr := c.cache.Set(ctx, key, result, TTLPodDetails); cacheErr != nil {
		log.Warn().Err(cacheErr).Str("pod_id", podID.String()).Msg("failed to cache pod")
	}

	return result, nil
}

// GetPodWithOwner retrieves a pod with owner info from cache or calls the loader.
func (c *PodCache) GetPodWithOwner(ctx context.Context, podID uuid.UUID, loader func() (*domain.PodWithOwner, error)) (*domain.PodWithOwner, error) {
	key := PodDetailsKey(podID.String())

	// Try cache first
	var pod domain.PodWithOwner
	found, err := c.cache.Get(ctx, key, &pod)
	if err != nil {
		log.Warn().Err(err).Str("pod_id", podID.String()).Msg("cache get failed for pod with owner")
	}
	if found {
		log.Debug().Str("pod_id", podID.String()).Msg("pod with owner cache hit")
		return &pod, nil
	}

	// Cache miss - call loader
	log.Debug().Str("pod_id", podID.String()).Msg("pod with owner cache miss")
	result, err := loader()
	if err != nil {
		return nil, err
	}

	// Store in cache
	if cacheErr := c.cache.Set(ctx, key, result, TTLPodDetails); cacheErr != nil {
		log.Warn().Err(cacheErr).Str("pod_id", podID.String()).Msg("failed to cache pod with owner")
	}

	return result, nil
}

// GetCollaborators retrieves pod collaborators from cache or calls the loader.
// Uses 15-minute TTL as collaborators may change more frequently.
func (c *PodCache) GetCollaborators(ctx context.Context, podID uuid.UUID, loader func() ([]*domain.CollaboratorWithUser, error)) ([]*domain.CollaboratorWithUser, error) {
	key := PodCollaboratorsKey(podID.String())

	// Try cache first
	var collaborators []*domain.CollaboratorWithUser
	found, err := c.cache.Get(ctx, key, &collaborators)
	if err != nil {
		log.Warn().Err(err).Str("pod_id", podID.String()).Msg("cache get failed for collaborators")
	}
	if found {
		log.Debug().Str("pod_id", podID.String()).Msg("collaborators cache hit")
		return collaborators, nil
	}

	// Cache miss - call loader
	log.Debug().Str("pod_id", podID.String()).Msg("collaborators cache miss")
	result, err := loader()
	if err != nil {
		return nil, err
	}

	// Store in cache
	if cacheErr := c.cache.Set(ctx, key, result, TTLPodCollaborators); cacheErr != nil {
		log.Warn().Err(cacheErr).Str("pod_id", podID.String()).Msg("failed to cache collaborators")
	}

	return result, nil
}

// Invalidate removes a pod from cache.
func (c *PodCache) Invalidate(ctx context.Context, podID uuid.UUID) error {
	keys := []string{
		PodDetailsKey(podID.String()),
		PodCollaboratorsKey(podID.String()),
	}
	return c.cache.Delete(ctx, keys...)
}

// SetPod stores a pod in cache.
func (c *PodCache) SetPod(ctx context.Context, podID uuid.UUID, pod *domain.Pod) error {
	key := PodDetailsKey(podID.String())
	return c.cache.Set(ctx, key, pod, TTLPodDetails)
}

// InvalidateCollaborators removes collaborators from cache.
func (c *PodCache) InvalidateCollaborators(ctx context.Context, podID uuid.UUID) error {
	key := PodCollaboratorsKey(podID.String())
	return c.cache.Delete(ctx, key)
}

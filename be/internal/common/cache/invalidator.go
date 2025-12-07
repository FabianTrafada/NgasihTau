// Package cache provides cache invalidation via NATS events.
// Implements requirement 24: Cache invalidation on data updates.
package cache

import (
	"context"
	"sync"

	"github.com/rs/zerolog/log"

	"ngasihtau/pkg/nats"
)

// Invalidator handles cache invalidation based on NATS events.
type Invalidator struct {
	cache         *Client
	natsClient    *nats.Client
	subscriptions []*nats.Subscription
	mu            sync.Mutex
}

// NewInvalidator creates a new cache invalidator.
func NewInvalidator(cache *Client, natsClient *nats.Client) *Invalidator {
	return &Invalidator{
		cache:         cache,
		natsClient:    natsClient,
		subscriptions: make([]*nats.Subscription, 0),
	}
}

// Start begins listening for cache invalidation events.
// It subscribes to user, pod, material, and collaborator update events
// and invalidates relevant cache entries.
func (inv *Invalidator) Start(ctx context.Context) error {
	inv.mu.Lock()
	defer inv.mu.Unlock()

	// Ensure the streams exist
	streams := []struct {
		name     string
		subjects []string
	}{
		{nats.StreamUser, nats.StreamSubjects[nats.StreamUser]},
		{nats.StreamPod, nats.StreamSubjects[nats.StreamPod]},
		{nats.StreamMaterial, nats.StreamSubjects[nats.StreamMaterial]},
		{nats.StreamCollaborator, nats.StreamSubjects[nats.StreamCollaborator]},
	}

	for _, s := range streams {
		if err := inv.natsClient.EnsureStream(ctx, s.name, s.subjects); err != nil {
			log.Warn().Err(err).Str("stream", s.name).Msg("failed to ensure stream for cache invalidation")
		}
	}

	// Subscribe to user events for cache invalidation
	userCfg := nats.SubscribeConfig{
		Stream:     nats.StreamUser,
		Consumer:   "cache-invalidator-user",
		Subjects:   []string{nats.SubjectUserUpdated},
		MaxDeliver: 3,
	}

	userSub, err := inv.natsClient.SubscribeSimple(ctx, userCfg, inv.handleUserEvent)
	if err != nil {
		log.Warn().Err(err).Msg("failed to subscribe to user events for cache invalidation")
	} else {
		inv.subscriptions = append(inv.subscriptions, userSub)
		log.Info().Msg("cache invalidator subscribed to user events")
	}

	// Subscribe to pod events
	podCfg := nats.SubscribeConfig{
		Stream:     nats.StreamPod,
		Consumer:   "cache-invalidator-pod",
		Subjects:   []string{nats.SubjectPodUpdated},
		MaxDeliver: 3,
	}

	podSub, err := inv.natsClient.SubscribeSimple(ctx, podCfg, inv.handlePodEvent)
	if err != nil {
		log.Warn().Err(err).Msg("failed to subscribe to pod events for cache invalidation")
	} else {
		inv.subscriptions = append(inv.subscriptions, podSub)
		log.Info().Msg("cache invalidator subscribed to pod events")
	}

	// Subscribe to material events
	matCfg := nats.SubscribeConfig{
		Stream:     nats.StreamMaterial,
		Consumer:   "cache-invalidator-material",
		Subjects:   []string{nats.SubjectMaterialProcessed, nats.SubjectMaterialDeleted},
		MaxDeliver: 3,
	}

	matSub, err := inv.natsClient.SubscribeSimple(ctx, matCfg, inv.handleMaterialEvent)
	if err != nil {
		log.Warn().Err(err).Msg("failed to subscribe to material events for cache invalidation")
	} else {
		inv.subscriptions = append(inv.subscriptions, matSub)
		log.Info().Msg("cache invalidator subscribed to material events")
	}

	// Subscribe to collaborator events for pod cache invalidation
	collabCfg := nats.SubscribeConfig{
		Stream:     nats.StreamCollaborator,
		Consumer:   "cache-invalidator-collaborator",
		Subjects:   []string{nats.SubjectCollaboratorInvited},
		MaxDeliver: 3,
	}

	collabSub, err := inv.natsClient.SubscribeSimple(ctx, collabCfg, inv.handleCollaboratorEvent)
	if err != nil {
		log.Warn().Err(err).Msg("failed to subscribe to collaborator events for cache invalidation")
	} else {
		inv.subscriptions = append(inv.subscriptions, collabSub)
		log.Info().Msg("cache invalidator subscribed to collaborator events")
	}

	return nil
}

// Stop stops the cache invalidator and all its subscriptions.
func (inv *Invalidator) Stop() {
	inv.mu.Lock()
	defer inv.mu.Unlock()

	for _, sub := range inv.subscriptions {
		if sub != nil {
			sub.Stop()
		}
	}
	inv.subscriptions = nil
	log.Info().Msg("cache invalidator stopped")
}

// handleUserEvent handles user-related events for cache invalidation.
func (inv *Invalidator) handleUserEvent(ctx context.Context, event nats.CloudEvent) error {
	switch event.Type {
	case nats.SubjectUserUpdated:
		return inv.invalidateUserCache(ctx, event)
	}
	return nil
}

// handlePodEvent handles pod-related events for cache invalidation.
func (inv *Invalidator) handlePodEvent(ctx context.Context, event nats.CloudEvent) error {
	switch event.Type {
	case nats.SubjectPodUpdated:
		return inv.invalidatePodCache(ctx, event)
	}
	return nil
}

// handleMaterialEvent handles material-related events for cache invalidation.
func (inv *Invalidator) handleMaterialEvent(ctx context.Context, event nats.CloudEvent) error {
	switch event.Type {
	case nats.SubjectMaterialProcessed, nats.SubjectMaterialDeleted:
		return inv.invalidateMaterialCache(ctx, event)
	}
	return nil
}

// invalidateUserCache invalidates user-related cache entries.
func (inv *Invalidator) invalidateUserCache(ctx context.Context, event nats.CloudEvent) error {
	data, err := nats.ParseEventData[nats.UserUpdatedEvent](event)
	if err != nil {
		log.Error().Err(err).Msg("failed to parse user updated event")
		return err
	}

	userID := data.UserID.String()
	key := UserProfileKey(userID)

	if err := inv.cache.Delete(ctx, key); err != nil {
		log.Error().Err(err).Str("user_id", userID).Msg("failed to invalidate user profile cache")
		return err
	}

	log.Debug().Str("user_id", userID).Msg("invalidated user profile cache")
	return nil
}

// invalidatePodCache invalidates pod-related cache entries.
func (inv *Invalidator) invalidatePodCache(ctx context.Context, event nats.CloudEvent) error {
	data, err := nats.ParseEventData[nats.PodUpdatedEvent](event)
	if err != nil {
		log.Error().Err(err).Msg("failed to parse pod updated event")
		return err
	}

	podID := data.PodID.String()

	// Invalidate pod details and collaborators cache
	keys := []string{
		PodDetailsKey(podID),
		PodCollaboratorsKey(podID),
	}

	if err := inv.cache.Delete(ctx, keys...); err != nil {
		log.Error().Err(err).Str("pod_id", podID).Msg("failed to invalidate pod cache")
		return err
	}

	log.Debug().Str("pod_id", podID).Msg("invalidated pod cache")
	return nil
}

// invalidateMaterialCache invalidates material-related cache entries.
func (inv *Invalidator) invalidateMaterialCache(ctx context.Context, event nats.CloudEvent) error {
	// Try to parse as MaterialProcessedEvent first
	processedData, err := nats.ParseEventData[nats.MaterialProcessedEvent](event)
	if err == nil {
		materialID := processedData.MaterialID.String()
		key := MaterialMetaKey(materialID)

		if err := inv.cache.Delete(ctx, key); err != nil {
			log.Error().Err(err).Str("material_id", materialID).Msg("failed to invalidate material cache")
			return err
		}

		log.Debug().Str("material_id", materialID).Msg("invalidated material cache")
		return nil
	}

	// Try to parse as MaterialDeletedEvent
	deletedData, err := nats.ParseEventData[nats.MaterialDeletedEvent](event)
	if err != nil {
		log.Error().Err(err).Msg("failed to parse material event")
		return err
	}

	materialID := deletedData.MaterialID.String()
	key := MaterialMetaKey(materialID)

	if err := inv.cache.Delete(ctx, key); err != nil {
		log.Error().Err(err).Str("material_id", materialID).Msg("failed to invalidate material cache")
		return err
	}

	log.Debug().Str("material_id", materialID).Msg("invalidated material cache on delete")
	return nil
}

// handleCollaboratorEvent handles collaborator-related events for cache invalidation.
func (inv *Invalidator) handleCollaboratorEvent(ctx context.Context, event nats.CloudEvent) error {
	switch event.Type {
	case nats.SubjectCollaboratorInvited:
		return inv.invalidateCollaboratorCache(ctx, event)
	}
	return nil
}

// invalidateCollaboratorCache invalidates pod collaborators cache when collaborators change.
func (inv *Invalidator) invalidateCollaboratorCache(ctx context.Context, event nats.CloudEvent) error {
	data, err := nats.ParseEventData[nats.CollaboratorInvitedEvent](event)
	if err != nil {
		log.Error().Err(err).Msg("failed to parse collaborator invited event")
		return err
	}

	podID := data.PodID.String()
	key := PodCollaboratorsKey(podID)

	if err := inv.cache.Delete(ctx, key); err != nil {
		log.Error().Err(err).Str("pod_id", podID).Msg("failed to invalidate pod collaborators cache")
		return err
	}

	log.Debug().Str("pod_id", podID).Msg("invalidated pod collaborators cache")
	return nil
}

// InvalidateUserProfile manually invalidates a user profile cache.
// Useful for direct cache invalidation without events.
func (inv *Invalidator) InvalidateUserProfile(ctx context.Context, userID string) error {
	return inv.cache.Delete(ctx, UserProfileKey(userID))
}

// InvalidatePodDetails manually invalidates pod details cache.
func (inv *Invalidator) InvalidatePodDetails(ctx context.Context, podID string) error {
	return inv.cache.Delete(ctx, PodDetailsKey(podID), PodCollaboratorsKey(podID))
}

// InvalidateMaterialMeta manually invalidates material metadata cache.
func (inv *Invalidator) InvalidateMaterialMeta(ctx context.Context, materialID string) error {
	return inv.cache.Delete(ctx, MaterialMetaKey(materialID))
}

// Package cache provides cached repository implementations.
// Implements requirement 24: Cache-aside pattern for material metadata.
package cache

import (
	"context"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"ngasihtau/internal/material/domain"
)

// MaterialCache provides cached access to material data.
type MaterialCache struct {
	cache *Client
}

// NewMaterialCache creates a new material cache.
func NewMaterialCache(cache *Client) *MaterialCache {
	return &MaterialCache{cache: cache}
}

// GetMaterial retrieves a material from cache or calls the loader if not cached.
// Implements cache-aside pattern with 1-hour TTL.
func (c *MaterialCache) GetMaterial(ctx context.Context, materialID uuid.UUID, loader func() (*domain.Material, error)) (*domain.Material, error) {
	key := MaterialMetaKey(materialID.String())

	// Try cache first
	var material domain.Material
	found, err := c.cache.Get(ctx, key, &material)
	if err != nil {
		log.Warn().Err(err).Str("material_id", materialID.String()).Msg("cache get failed for material")
	}
	if found {
		log.Debug().Str("material_id", materialID.String()).Msg("material cache hit")
		return &material, nil
	}

	// Cache miss - call loader
	log.Debug().Str("material_id", materialID.String()).Msg("material cache miss")
	result, err := loader()
	if err != nil {
		return nil, err
	}

	// Store in cache
	if cacheErr := c.cache.Set(ctx, key, result, TTLMaterialMeta); cacheErr != nil {
		log.Warn().Err(cacheErr).Str("material_id", materialID.String()).Msg("failed to cache material")
	}

	return result, nil
}

// GetMaterialWithUploader retrieves a material with uploader info from cache or calls the loader.
func (c *MaterialCache) GetMaterialWithUploader(ctx context.Context, materialID uuid.UUID, loader func() (*domain.MaterialWithUploader, error)) (*domain.MaterialWithUploader, error) {
	key := MaterialMetaKey(materialID.String())

	// Try cache first
	var material domain.MaterialWithUploader
	found, err := c.cache.Get(ctx, key, &material)
	if err != nil {
		log.Warn().Err(err).Str("material_id", materialID.String()).Msg("cache get failed for material with uploader")
	}
	if found {
		log.Debug().Str("material_id", materialID.String()).Msg("material with uploader cache hit")
		return &material, nil
	}

	// Cache miss - call loader
	log.Debug().Str("material_id", materialID.String()).Msg("material with uploader cache miss")
	result, err := loader()
	if err != nil {
		return nil, err
	}

	// Store in cache
	if cacheErr := c.cache.Set(ctx, key, result, TTLMaterialMeta); cacheErr != nil {
		log.Warn().Err(cacheErr).Str("material_id", materialID.String()).Msg("failed to cache material with uploader")
	}

	return result, nil
}

// Invalidate removes a material from cache.
func (c *MaterialCache) Invalidate(ctx context.Context, materialID uuid.UUID) error {
	key := MaterialMetaKey(materialID.String())
	return c.cache.Delete(ctx, key)
}

// SetMaterial stores a material in cache.
func (c *MaterialCache) SetMaterial(ctx context.Context, materialID uuid.UUID, material *domain.Material) error {
	key := MaterialMetaKey(materialID.String())
	return c.cache.Set(ctx, key, material, TTLMaterialMeta)
}

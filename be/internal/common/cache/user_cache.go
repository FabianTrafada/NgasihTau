// Package cache provides cached repository implementations.
// Implements requirement 24: Cache-aside pattern for user profiles.
package cache

import (
	"context"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"ngasihtau/internal/user/domain"
)

// UserProfileCache provides cached access to user profiles.
type UserProfileCache struct {
	cache *Client
}

// NewUserProfileCache creates a new user profile cache.
func NewUserProfileCache(cache *Client) *UserProfileCache {
	return &UserProfileCache{cache: cache}
}

// GetProfile retrieves a user profile from cache or calls the loader if not cached.
// Implements cache-aside pattern with 1-hour TTL.
func (c *UserProfileCache) GetProfile(ctx context.Context, userID uuid.UUID, loader func() (*domain.UserProfile, error)) (*domain.UserProfile, error) {
	key := UserProfileKey(userID.String())

	// Try cache first
	var profile domain.UserProfile
	found, err := c.cache.Get(ctx, key, &profile)
	if err != nil {
		log.Warn().Err(err).Str("user_id", userID.String()).Msg("cache get failed for user profile")
	}
	if found {
		log.Debug().Str("user_id", userID.String()).Msg("user profile cache hit")
		return &profile, nil
	}

	// Cache miss - call loader
	log.Debug().Str("user_id", userID.String()).Msg("user profile cache miss")
	result, err := loader()
	if err != nil {
		return nil, err
	}

	// Store in cache
	if cacheErr := c.cache.Set(ctx, key, result, TTLUserProfile); cacheErr != nil {
		log.Warn().Err(cacheErr).Str("user_id", userID.String()).Msg("failed to cache user profile")
	}

	return result, nil
}

// GetUser retrieves a user from cache or calls the loader if not cached.
// Uses the same cache key as profile but stores full user data.
func (c *UserProfileCache) GetUser(ctx context.Context, userID uuid.UUID, loader func() (*domain.User, error)) (*domain.User, error) {
	key := UserProfileKey(userID.String())

	// Try cache first
	var user domain.User
	found, err := c.cache.Get(ctx, key, &user)
	if err != nil {
		log.Warn().Err(err).Str("user_id", userID.String()).Msg("cache get failed for user")
	}
	if found {
		log.Debug().Str("user_id", userID.String()).Msg("user cache hit")
		return &user, nil
	}

	// Cache miss - call loader
	log.Debug().Str("user_id", userID.String()).Msg("user cache miss")
	result, err := loader()
	if err != nil {
		return nil, err
	}

	// Store in cache
	if cacheErr := c.cache.Set(ctx, key, result, TTLUserProfile); cacheErr != nil {
		log.Warn().Err(cacheErr).Str("user_id", userID.String()).Msg("failed to cache user")
	}

	return result, nil
}

// Invalidate removes a user profile from cache.
func (c *UserProfileCache) Invalidate(ctx context.Context, userID uuid.UUID) error {
	key := UserProfileKey(userID.String())
	return c.cache.Delete(ctx, key)
}

// InvalidateOnRoleChange removes a user profile from cache when role changes.
// This should be called when a user's role changes from student to teacher.
// Implements requirement 2.2: Cache invalidation when user role changes.
func (c *UserProfileCache) InvalidateOnRoleChange(ctx context.Context, userID uuid.UUID) error {
	key := UserProfileKey(userID.String())
	if err := c.cache.Delete(ctx, key); err != nil {
		log.Warn().Err(err).Str("user_id", userID.String()).Msg("failed to invalidate user cache on role change")
		return err
	}
	log.Debug().Str("user_id", userID.String()).Msg("invalidated user cache on role change")
	return nil
}

// SetProfile stores a user profile in cache.
func (c *UserProfileCache) SetProfile(ctx context.Context, userID uuid.UUID, profile *domain.UserProfile) error {
	key := UserProfileKey(userID.String())
	return c.cache.Set(ctx, key, profile, TTLUserProfile)
}

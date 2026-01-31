// Package application contains the business logic and use cases for the Offline Material feature.
package application

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"

	"ngasihtau/internal/offline/domain"
)

// RateLimiter defines the interface for rate limiting operations.
// Implements Requirement 6: Rate Limiting.
type RateLimiter interface {
	// CheckDownloadLimit checks if a user has exceeded their download limit.
	// Returns (allowed, remaining, resetTime, error).
	CheckDownloadLimit(ctx context.Context, userID uuid.UUID) (bool, int, time.Time, error)

	// IncrementDownload increments the download counter for a user.
	// Returns (newCount, error).
	IncrementDownload(ctx context.Context, userID uuid.UUID) (int, error)

	// CheckMaterialDownloadLimit checks if a material has exceeded its download limit.
	// Returns (allowed, remaining, resetTime, error).
	CheckMaterialDownloadLimit(ctx context.Context, materialID uuid.UUID) (bool, int, time.Time, error)

	// IncrementMaterialDownload increments the download counter for a material.
	// Returns (newCount, error).
	IncrementMaterialDownload(ctx context.Context, materialID uuid.UUID) (int, error)

	// RecordValidationFailure records a validation failure for a device.
	// Returns (failureCount, isBlocked, error).
	RecordValidationFailure(ctx context.Context, deviceID uuid.UUID) (int, bool, error)

	// ResetValidationFailures resets the validation failure counter for a device.
	ResetValidationFailures(ctx context.Context, deviceID uuid.UUID) error

	// IsDeviceBlocked checks if a device is blocked due to too many failures.
	IsDeviceBlocked(ctx context.Context, deviceID uuid.UUID) (bool, time.Time, error)

	// BlockDevice blocks a device for a specified duration.
	BlockDevice(ctx context.Context, deviceID uuid.UUID, duration time.Duration) error

	// UnblockDevice removes the block from a device.
	UnblockDevice(ctx context.Context, deviceID uuid.UUID) error

	// GetRateLimitInfo returns rate limit information for response headers.
	GetRateLimitInfo(ctx context.Context, userID uuid.UUID) (*RateLimitInfo, error)
}

// RateLimitInfo contains rate limit information for HTTP headers.
type RateLimitInfo struct {
	Limit     int       `json:"limit"`
	Remaining int       `json:"remaining"`
	Reset     time.Time `json:"reset"`
}

// rateLimiter implements the RateLimiter interface using Redis.
type rateLimiter struct {
	client *redis.Client
}

// NewRateLimiter creates a new RateLimiter instance.
func NewRateLimiter(client *redis.Client) RateLimiter {
	return &rateLimiter{client: client}
}

// Redis key prefixes for rate limiting.
const (
	rateLimitDownloadPrefix         = "offline:ratelimit:download:user:"
	rateLimitMaterialDownloadPrefix = "offline:ratelimit:download:material:"
	rateLimitValidationPrefix       = "offline:ratelimit:validation:"
	rateLimitBlockPrefix            = "offline:ratelimit:block:"
)

// CheckDownloadLimit checks if a user has exceeded their download limit.
// Implements Requirement 6.1: Per-user download limits (10/hour).
// Implements Property 19: Download Rate Limiting.
func (r *rateLimiter) CheckDownloadLimit(ctx context.Context, userID uuid.UUID) (bool, int, time.Time, error) {
	key := rateLimitDownloadPrefix + userID.String()

	count, err := r.client.Get(ctx, key).Int()
	if err == redis.Nil {
		// No counter exists, user is allowed
		resetTime := time.Now().Add(domain.RateLimitWindow)
		return true, domain.MaxDownloadsPerHour, resetTime, nil
	}
	if err != nil {
		log.Error().Err(err).Str("user_id", userID.String()).Msg("failed to get download count")
		return false, 0, time.Time{}, fmt.Errorf("failed to check download limit: %w", err)
	}

	// Get TTL for reset time
	ttl, err := r.client.TTL(ctx, key).Result()
	if err != nil {
		ttl = domain.RateLimitWindow
	}
	resetTime := time.Now().Add(ttl)

	remaining := domain.MaxDownloadsPerHour - count
	if remaining < 0 {
		remaining = 0
	}

	allowed := count < domain.MaxDownloadsPerHour
	return allowed, remaining, resetTime, nil
}

// IncrementDownload increments the download counter for a user.
func (r *rateLimiter) IncrementDownload(ctx context.Context, userID uuid.UUID) (int, error) {
	key := rateLimitDownloadPrefix + userID.String()

	pipe := r.client.Pipeline()
	incr := pipe.Incr(ctx, key)
	// Set expiry only if key is new (NX flag via SetNX pattern)
	pipe.Expire(ctx, key, domain.RateLimitWindow)

	_, err := pipe.Exec(ctx)
	if err != nil {
		log.Error().Err(err).Str("user_id", userID.String()).Msg("failed to increment download count")
		return 0, fmt.Errorf("failed to increment download count: %w", err)
	}

	return int(incr.Val()), nil
}

// CheckMaterialDownloadLimit checks if a material has exceeded its download limit.
// Implements Requirement 6.2: Per-material download limits.
func (r *rateLimiter) CheckMaterialDownloadLimit(ctx context.Context, materialID uuid.UUID) (bool, int, time.Time, error) {
	key := rateLimitMaterialDownloadPrefix + materialID.String()

	// Material limit is higher - 100 downloads per hour
	const maxMaterialDownloads = 100

	count, err := r.client.Get(ctx, key).Int()
	if err == redis.Nil {
		resetTime := time.Now().Add(domain.RateLimitWindow)
		return true, maxMaterialDownloads, resetTime, nil
	}
	if err != nil {
		log.Error().Err(err).Str("material_id", materialID.String()).Msg("failed to get material download count")
		return false, 0, time.Time{}, fmt.Errorf("failed to check material download limit: %w", err)
	}

	ttl, err := r.client.TTL(ctx, key).Result()
	if err != nil {
		ttl = domain.RateLimitWindow
	}
	resetTime := time.Now().Add(ttl)

	remaining := maxMaterialDownloads - count
	if remaining < 0 {
		remaining = 0
	}

	allowed := count < maxMaterialDownloads
	return allowed, remaining, resetTime, nil
}

// IncrementMaterialDownload increments the download counter for a material.
func (r *rateLimiter) IncrementMaterialDownload(ctx context.Context, materialID uuid.UUID) (int, error) {
	key := rateLimitMaterialDownloadPrefix + materialID.String()

	pipe := r.client.Pipeline()
	incr := pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, domain.RateLimitWindow)

	_, err := pipe.Exec(ctx)
	if err != nil {
		log.Error().Err(err).Str("material_id", materialID.String()).Msg("failed to increment material download count")
		return 0, fmt.Errorf("failed to increment material download count: %w", err)
	}

	return int(incr.Val()), nil
}

// RecordValidationFailure records a validation failure for a device.
// Implements Requirement 6.3: Device validation failure tracking.
// Implements Property 28: Device Blocking on Failures.
func (r *rateLimiter) RecordValidationFailure(ctx context.Context, deviceID uuid.UUID) (int, bool, error) {
	key := rateLimitValidationPrefix + deviceID.String()

	pipe := r.client.Pipeline()
	incr := pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, domain.RateLimitWindow)

	_, err := pipe.Exec(ctx)
	if err != nil {
		log.Error().Err(err).Str("device_id", deviceID.String()).Msg("failed to record validation failure")
		return 0, false, fmt.Errorf("failed to record validation failure: %w", err)
	}

	count := int(incr.Val())
	shouldBlock := count >= domain.MaxValidationFailuresPerHour

	// Auto-block device if threshold exceeded
	if shouldBlock {
		if err := r.BlockDevice(ctx, deviceID, domain.DeviceBlockDuration); err != nil {
			log.Error().Err(err).Str("device_id", deviceID.String()).Msg("failed to auto-block device")
		}
	}

	return count, shouldBlock, nil
}

// ResetValidationFailures resets the validation failure counter for a device.
func (r *rateLimiter) ResetValidationFailures(ctx context.Context, deviceID uuid.UUID) error {
	key := rateLimitValidationPrefix + deviceID.String()

	if err := r.client.Del(ctx, key).Err(); err != nil {
		log.Error().Err(err).Str("device_id", deviceID.String()).Msg("failed to reset validation failures")
		return fmt.Errorf("failed to reset validation failures: %w", err)
	}

	return nil
}

// IsDeviceBlocked checks if a device is blocked due to too many failures.
// Implements Requirement 6.4: Device blocking after 5 failures.
func (r *rateLimiter) IsDeviceBlocked(ctx context.Context, deviceID uuid.UUID) (bool, time.Time, error) {
	key := rateLimitBlockPrefix + deviceID.String()

	blockedUntilStr, err := r.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return false, time.Time{}, nil
	}
	if err != nil {
		log.Error().Err(err).Str("device_id", deviceID.String()).Msg("failed to check device block status")
		return false, time.Time{}, fmt.Errorf("failed to check device block status: %w", err)
	}

	blockedUntilUnix, err := strconv.ParseInt(blockedUntilStr, 10, 64)
	if err != nil {
		return false, time.Time{}, fmt.Errorf("failed to parse blocked until time: %w", err)
	}

	blockedUntil := time.Unix(blockedUntilUnix, 0)
	if time.Now().After(blockedUntil) {
		// Block has expired, clean up
		_ = r.client.Del(ctx, key)
		return false, time.Time{}, nil
	}

	return true, blockedUntil, nil
}

// BlockDevice blocks a device for a specified duration.
func (r *rateLimiter) BlockDevice(ctx context.Context, deviceID uuid.UUID, duration time.Duration) error {
	key := rateLimitBlockPrefix + deviceID.String()
	blockedUntil := time.Now().Add(duration)

	if err := r.client.Set(ctx, key, blockedUntil.Unix(), duration).Err(); err != nil {
		log.Error().Err(err).Str("device_id", deviceID.String()).Msg("failed to block device")
		return fmt.Errorf("failed to block device: %w", err)
	}

	log.Warn().
		Str("device_id", deviceID.String()).
		Time("blocked_until", blockedUntil).
		Msg("device blocked due to too many validation failures")

	return nil
}

// UnblockDevice removes the block from a device.
func (r *rateLimiter) UnblockDevice(ctx context.Context, deviceID uuid.UUID) error {
	key := rateLimitBlockPrefix + deviceID.String()

	if err := r.client.Del(ctx, key).Err(); err != nil {
		log.Error().Err(err).Str("device_id", deviceID.String()).Msg("failed to unblock device")
		return fmt.Errorf("failed to unblock device: %w", err)
	}

	// Also reset validation failures
	return r.ResetValidationFailures(ctx, deviceID)
}

// GetRateLimitInfo returns rate limit information for response headers.
func (r *rateLimiter) GetRateLimitInfo(ctx context.Context, userID uuid.UUID) (*RateLimitInfo, error) {
	allowed, remaining, resetTime, err := r.CheckDownloadLimit(ctx, userID)
	if err != nil {
		return nil, err
	}

	_ = allowed // Not used here, just for info

	return &RateLimitInfo{
		Limit:     domain.MaxDownloadsPerHour,
		Remaining: remaining,
		Reset:     resetTime,
	}, nil
}

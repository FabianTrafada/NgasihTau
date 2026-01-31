package application

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"ngasihtau/internal/offline/domain"
)

func setupTestRedis(t *testing.T) (*redis.Client, func()) {
	mr, err := miniredis.Run()
	require.NoError(t, err)

	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	return client, func() {
		client.Close()
		mr.Close()
	}
}

func TestRateLimiter_CheckDownloadLimit(t *testing.T) {
	client, cleanup := setupTestRedis(t)
	defer cleanup()

	limiter := NewRateLimiter(client)
	ctx := context.Background()
	userID := uuid.New()

	// First check - should be allowed with full remaining
	allowed, remaining, _, err := limiter.CheckDownloadLimit(ctx, userID)
	require.NoError(t, err)
	assert.True(t, allowed)
	assert.Equal(t, domain.MaxDownloadsPerHour, remaining)

	// Increment some downloads
	for i := 0; i < 5; i++ {
		_, err := limiter.IncrementDownload(ctx, userID)
		require.NoError(t, err)
	}

	// Check again - should still be allowed with reduced remaining
	allowed, remaining, _, err = limiter.CheckDownloadLimit(ctx, userID)
	require.NoError(t, err)
	assert.True(t, allowed)
	assert.Equal(t, domain.MaxDownloadsPerHour-5, remaining)

	// Exhaust the limit
	for i := 0; i < domain.MaxDownloadsPerHour-5; i++ {
		_, err := limiter.IncrementDownload(ctx, userID)
		require.NoError(t, err)
	}

	// Check again - should not be allowed
	allowed, remaining, _, err = limiter.CheckDownloadLimit(ctx, userID)
	require.NoError(t, err)
	assert.False(t, allowed)
	assert.Equal(t, 0, remaining)
}

func TestRateLimiter_IncrementDownload(t *testing.T) {
	client, cleanup := setupTestRedis(t)
	defer cleanup()

	limiter := NewRateLimiter(client)
	ctx := context.Background()
	userID := uuid.New()

	// First increment
	count, err := limiter.IncrementDownload(ctx, userID)
	require.NoError(t, err)
	assert.Equal(t, 1, count)

	// Second increment
	count, err = limiter.IncrementDownload(ctx, userID)
	require.NoError(t, err)
	assert.Equal(t, 2, count)
}

func TestRateLimiter_RecordValidationFailure(t *testing.T) {
	client, cleanup := setupTestRedis(t)
	defer cleanup()

	limiter := NewRateLimiter(client)
	ctx := context.Background()
	deviceID := uuid.New()

	// Record failures up to threshold
	for i := 1; i < domain.MaxValidationFailuresPerHour; i++ {
		count, blocked, err := limiter.RecordValidationFailure(ctx, deviceID)
		require.NoError(t, err)
		assert.Equal(t, i, count)
		assert.False(t, blocked)
	}

	// Record one more failure - should trigger block
	count, blocked, err := limiter.RecordValidationFailure(ctx, deviceID)
	require.NoError(t, err)
	assert.Equal(t, domain.MaxValidationFailuresPerHour, count)
	assert.True(t, blocked)

	// Verify device is blocked
	isBlocked, _, err := limiter.IsDeviceBlocked(ctx, deviceID)
	require.NoError(t, err)
	assert.True(t, isBlocked)
}

func TestRateLimiter_IsDeviceBlocked(t *testing.T) {
	client, cleanup := setupTestRedis(t)
	defer cleanup()

	limiter := NewRateLimiter(client)
	ctx := context.Background()
	deviceID := uuid.New()

	// Initially not blocked
	isBlocked, _, err := limiter.IsDeviceBlocked(ctx, deviceID)
	require.NoError(t, err)
	assert.False(t, isBlocked)

	// Block the device
	err = limiter.BlockDevice(ctx, deviceID, 1*time.Hour)
	require.NoError(t, err)

	// Now should be blocked
	isBlocked, blockedUntil, err := limiter.IsDeviceBlocked(ctx, deviceID)
	require.NoError(t, err)
	assert.True(t, isBlocked)
	assert.True(t, blockedUntil.After(time.Now()))
}

func TestRateLimiter_UnblockDevice(t *testing.T) {
	client, cleanup := setupTestRedis(t)
	defer cleanup()

	limiter := NewRateLimiter(client)
	ctx := context.Background()
	deviceID := uuid.New()

	// Block the device
	err := limiter.BlockDevice(ctx, deviceID, 1*time.Hour)
	require.NoError(t, err)

	// Verify blocked
	isBlocked, _, err := limiter.IsDeviceBlocked(ctx, deviceID)
	require.NoError(t, err)
	assert.True(t, isBlocked)

	// Unblock
	err = limiter.UnblockDevice(ctx, deviceID)
	require.NoError(t, err)

	// Verify unblocked
	isBlocked, _, err = limiter.IsDeviceBlocked(ctx, deviceID)
	require.NoError(t, err)
	assert.False(t, isBlocked)
}

func TestRateLimiter_GetRateLimitInfo(t *testing.T) {
	client, cleanup := setupTestRedis(t)
	defer cleanup()

	limiter := NewRateLimiter(client)
	ctx := context.Background()
	userID := uuid.New()

	// Get info for new user
	info, err := limiter.GetRateLimitInfo(ctx, userID)
	require.NoError(t, err)
	assert.Equal(t, domain.MaxDownloadsPerHour, info.Limit)
	assert.Equal(t, domain.MaxDownloadsPerHour, info.Remaining)
	assert.True(t, info.Reset.After(time.Now()))

	// Increment some downloads
	for i := 0; i < 3; i++ {
		_, err := limiter.IncrementDownload(ctx, userID)
		require.NoError(t, err)
	}

	// Get info again
	info, err = limiter.GetRateLimitInfo(ctx, userID)
	require.NoError(t, err)
	assert.Equal(t, domain.MaxDownloadsPerHour, info.Limit)
	assert.Equal(t, domain.MaxDownloadsPerHour-3, info.Remaining)
}

func TestRateLimiter_MaterialDownloadLimit(t *testing.T) {
	client, cleanup := setupTestRedis(t)
	defer cleanup()

	limiter := NewRateLimiter(client)
	ctx := context.Background()
	materialID := uuid.New()

	// First check - should be allowed
	allowed, remaining, _, err := limiter.CheckMaterialDownloadLimit(ctx, materialID)
	require.NoError(t, err)
	assert.True(t, allowed)
	assert.Equal(t, 100, remaining) // Material limit is 100

	// Increment
	count, err := limiter.IncrementMaterialDownload(ctx, materialID)
	require.NoError(t, err)
	assert.Equal(t, 1, count)

	// Check again
	allowed, remaining, _, err = limiter.CheckMaterialDownloadLimit(ctx, materialID)
	require.NoError(t, err)
	assert.True(t, allowed)
	assert.Equal(t, 99, remaining)
}

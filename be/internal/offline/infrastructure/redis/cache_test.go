package redis

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

// setupTestCache creates a test cache with miniredis.
func setupTestCache(t *testing.T) (*OfflineCache, *miniredis.Miniredis) {
	mr, err := miniredis.Run()
	require.NoError(t, err)

	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	cache := NewOfflineCache(client)
	return cache, mr
}

// createTestLicense creates a test license.
func createTestLicense() *domain.License {
	return &domain.License{
		ID:                 uuid.New(),
		UserID:             uuid.New(),
		MaterialID:         uuid.New(),
		DeviceID:           uuid.New(),
		Status:             domain.LicenseStatusActive,
		ExpiresAt:          time.Now().Add(30 * 24 * time.Hour),
		OfflineGracePeriod: 72 * time.Hour,
		LastValidatedAt:    time.Now(),
		Nonce:              "test-nonce-1234567890abcdef",
		CreatedAt:          time.Now(),
	}
}

// createTestDeviceForCache creates a test device.
func createTestDeviceForCache() *domain.Device {
	return &domain.Device{
		ID:          uuid.New(),
		UserID:      uuid.New(),
		Fingerprint: "test-fingerprint-12345678901234567890",
		Name:        "Test Device",
		Platform:    domain.PlatformIOS,
		LastUsedAt:  time.Now(),
		CreatedAt:   time.Now(),
	}
}

func TestOfflineCache_Ping(t *testing.T) {
	cache, mr := setupTestCache(t)
	defer mr.Close()

	ctx := context.Background()
	err := cache.Ping(ctx)
	assert.NoError(t, err)
}

// ============================================================================
// License Cache Tests
// ============================================================================

func TestOfflineCache_SetGetLicense(t *testing.T) {
	cache, mr := setupTestCache(t)
	defer mr.Close()

	ctx := context.Background()
	license := createTestLicense()

	// Set license
	err := cache.SetLicense(ctx, license)
	assert.NoError(t, err)

	// Get license
	cached, err := cache.GetLicense(ctx, license.ID)
	assert.NoError(t, err)
	assert.NotNil(t, cached)
	assert.Equal(t, license.ID, cached.ID)
	assert.Equal(t, license.UserID, cached.UserID)
	assert.Equal(t, license.MaterialID, cached.MaterialID)
	assert.Equal(t, license.DeviceID, cached.DeviceID)
	assert.Equal(t, license.Status, cached.Status)
	assert.Equal(t, license.Nonce, cached.Nonce)
}

func TestOfflineCache_GetLicense_NotFound(t *testing.T) {
	cache, mr := setupTestCache(t)
	defer mr.Close()

	ctx := context.Background()
	cached, err := cache.GetLicense(ctx, uuid.New())
	assert.NoError(t, err)
	assert.Nil(t, cached)
}

func TestOfflineCache_GetLicenseByUserMaterialDevice(t *testing.T) {
	cache, mr := setupTestCache(t)
	defer mr.Close()

	ctx := context.Background()
	license := createTestLicense()

	// Set license (also sets lookup key)
	err := cache.SetLicense(ctx, license)
	assert.NoError(t, err)

	// Get by user/material/device
	cachedID, err := cache.GetLicenseByUserMaterialDevice(ctx, license.UserID, license.MaterialID, license.DeviceID)
	assert.NoError(t, err)
	assert.NotNil(t, cachedID)
	assert.Equal(t, license.ID, *cachedID)
}

func TestOfflineCache_InvalidateLicense(t *testing.T) {
	cache, mr := setupTestCache(t)
	defer mr.Close()

	ctx := context.Background()
	license := createTestLicense()

	// Set license
	err := cache.SetLicense(ctx, license)
	assert.NoError(t, err)

	// Verify it exists
	cached, err := cache.GetLicense(ctx, license.ID)
	assert.NoError(t, err)
	assert.NotNil(t, cached)

	// Invalidate
	err = cache.InvalidateLicense(ctx, license)
	assert.NoError(t, err)

	// Verify it's gone
	cached, err = cache.GetLicense(ctx, license.ID)
	assert.NoError(t, err)
	assert.Nil(t, cached)

	// Verify lookup is also gone
	cachedID, err := cache.GetLicenseByUserMaterialDevice(ctx, license.UserID, license.MaterialID, license.DeviceID)
	assert.NoError(t, err)
	assert.Nil(t, cachedID)
}


// ============================================================================
// Device Cache Tests
// ============================================================================

func TestOfflineCache_SetGetDevice(t *testing.T) {
	cache, mr := setupTestCache(t)
	defer mr.Close()

	ctx := context.Background()
	device := createTestDeviceForCache()

	// Set device
	err := cache.SetDevice(ctx, device)
	assert.NoError(t, err)

	// Get device
	cached, err := cache.GetDevice(ctx, device.ID)
	assert.NoError(t, err)
	assert.NotNil(t, cached)
	assert.Equal(t, device.ID, cached.ID)
	assert.Equal(t, device.UserID, cached.UserID)
	assert.Equal(t, device.Fingerprint, cached.Fingerprint)
	assert.Equal(t, device.Name, cached.Name)
	assert.Equal(t, device.Platform, cached.Platform)
}

func TestOfflineCache_GetDevice_NotFound(t *testing.T) {
	cache, mr := setupTestCache(t)
	defer mr.Close()

	ctx := context.Background()
	cached, err := cache.GetDevice(ctx, uuid.New())
	assert.NoError(t, err)
	assert.Nil(t, cached)
}

func TestOfflineCache_GetDeviceByFingerprint(t *testing.T) {
	cache, mr := setupTestCache(t)
	defer mr.Close()

	ctx := context.Background()
	device := createTestDeviceForCache()

	// Set device (also sets fingerprint lookup)
	err := cache.SetDevice(ctx, device)
	assert.NoError(t, err)

	// Get by fingerprint
	cachedID, err := cache.GetDeviceByFingerprint(ctx, device.UserID, device.Fingerprint)
	assert.NoError(t, err)
	assert.NotNil(t, cachedID)
	assert.Equal(t, device.ID, *cachedID)
}

func TestOfflineCache_InvalidateDevice(t *testing.T) {
	cache, mr := setupTestCache(t)
	defer mr.Close()

	ctx := context.Background()
	device := createTestDeviceForCache()

	// Set device
	err := cache.SetDevice(ctx, device)
	assert.NoError(t, err)

	// Verify it exists
	cached, err := cache.GetDevice(ctx, device.ID)
	assert.NoError(t, err)
	assert.NotNil(t, cached)

	// Invalidate
	err = cache.InvalidateDevice(ctx, device)
	assert.NoError(t, err)

	// Verify it's gone
	cached, err = cache.GetDevice(ctx, device.ID)
	assert.NoError(t, err)
	assert.Nil(t, cached)

	// Verify fingerprint lookup is also gone
	cachedID, err := cache.GetDeviceByFingerprint(ctx, device.UserID, device.Fingerprint)
	assert.NoError(t, err)
	assert.Nil(t, cachedID)
}

// ============================================================================
// User Device List Cache Tests
// ============================================================================

func TestOfflineCache_SetGetUserDevices(t *testing.T) {
	cache, mr := setupTestCache(t)
	defer mr.Close()

	ctx := context.Background()
	userID := uuid.New()
	devices := []*domain.Device{
		createTestDeviceForCache(),
		createTestDeviceForCache(),
	}
	// Set same user ID for all devices
	for _, d := range devices {
		d.UserID = userID
	}

	// Set user devices
	err := cache.SetUserDevices(ctx, userID, devices)
	assert.NoError(t, err)

	// Get user devices
	cached, err := cache.GetUserDevices(ctx, userID)
	assert.NoError(t, err)
	assert.NotNil(t, cached)
	assert.Len(t, cached, 2)
}

func TestOfflineCache_GetUserDevices_NotFound(t *testing.T) {
	cache, mr := setupTestCache(t)
	defer mr.Close()

	ctx := context.Background()
	cached, err := cache.GetUserDevices(ctx, uuid.New())
	assert.NoError(t, err)
	assert.Nil(t, cached)
}

func TestOfflineCache_InvalidateUserDevices(t *testing.T) {
	cache, mr := setupTestCache(t)
	defer mr.Close()

	ctx := context.Background()
	userID := uuid.New()
	devices := []*domain.Device{createTestDeviceForCache()}

	// Set user devices
	err := cache.SetUserDevices(ctx, userID, devices)
	assert.NoError(t, err)

	// Verify it exists
	cached, err := cache.GetUserDevices(ctx, userID)
	assert.NoError(t, err)
	assert.NotNil(t, cached)

	// Invalidate
	err = cache.InvalidateUserDevices(ctx, userID)
	assert.NoError(t, err)

	// Verify it's gone
	cached, err = cache.GetUserDevices(ctx, userID)
	assert.NoError(t, err)
	assert.Nil(t, cached)
}

// ============================================================================
// Rate Limiting Tests
// ============================================================================

func TestOfflineCache_DownloadCount(t *testing.T) {
	cache, mr := setupTestCache(t)
	defer mr.Close()

	ctx := context.Background()
	userID := uuid.New()

	// Initial count should be 0
	count, err := cache.GetDownloadCount(ctx, userID)
	assert.NoError(t, err)
	assert.Equal(t, 0, count)

	// Increment
	count, err = cache.IncrementDownloadCount(ctx, userID)
	assert.NoError(t, err)
	assert.Equal(t, 1, count)

	// Increment again
	count, err = cache.IncrementDownloadCount(ctx, userID)
	assert.NoError(t, err)
	assert.Equal(t, 2, count)

	// Get count
	count, err = cache.GetDownloadCount(ctx, userID)
	assert.NoError(t, err)
	assert.Equal(t, 2, count)
}

func TestOfflineCache_ValidationFailures(t *testing.T) {
	cache, mr := setupTestCache(t)
	defer mr.Close()

	ctx := context.Background()
	deviceID := uuid.New()

	// Initial count should be 0
	count, err := cache.GetValidationFailureCount(ctx, deviceID)
	assert.NoError(t, err)
	assert.Equal(t, 0, count)

	// Device should not be blocked
	blocked, err := cache.IsDeviceBlocked(ctx, deviceID)
	assert.NoError(t, err)
	assert.False(t, blocked)

	// Increment failures up to threshold
	for i := 0; i < domain.MaxValidationFailuresPerHour; i++ {
		_, err = cache.IncrementValidationFailure(ctx, deviceID)
		assert.NoError(t, err)
	}

	// Device should now be blocked
	blocked, err = cache.IsDeviceBlocked(ctx, deviceID)
	assert.NoError(t, err)
	assert.True(t, blocked)

	// Reset failures
	err = cache.ResetValidationFailures(ctx, deviceID)
	assert.NoError(t, err)

	// Device should no longer be blocked
	blocked, err = cache.IsDeviceBlocked(ctx, deviceID)
	assert.NoError(t, err)
	assert.False(t, blocked)
}

// ============================================================================
// Cache-Aside Pattern Tests
// ============================================================================

func TestOfflineCache_GetOrSetLicense(t *testing.T) {
	cache, mr := setupTestCache(t)
	defer mr.Close()

	ctx := context.Background()
	license := createTestLicense()
	loaderCalled := false

	loader := func() (*domain.License, error) {
		loaderCalled = true
		return license, nil
	}

	// First call - should call loader
	result, err := cache.GetOrSetLicense(ctx, license.ID, loader)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, loaderCalled)
	assert.Equal(t, license.ID, result.ID)

	// Reset flag
	loaderCalled = false

	// Second call - should use cache
	result, err = cache.GetOrSetLicense(ctx, license.ID, loader)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.False(t, loaderCalled) // Loader should NOT be called
	assert.Equal(t, license.ID, result.ID)
}

func TestOfflineCache_GetOrSetDevice(t *testing.T) {
	cache, mr := setupTestCache(t)
	defer mr.Close()

	ctx := context.Background()
	device := createTestDeviceForCache()
	loaderCalled := false

	loader := func() (*domain.Device, error) {
		loaderCalled = true
		return device, nil
	}

	// First call - should call loader
	result, err := cache.GetOrSetDevice(ctx, device.ID, loader)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, loaderCalled)
	assert.Equal(t, device.ID, result.ID)

	// Reset flag
	loaderCalled = false

	// Second call - should use cache
	result, err = cache.GetOrSetDevice(ctx, device.ID, loader)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.False(t, loaderCalled) // Loader should NOT be called
	assert.Equal(t, device.ID, result.ID)
}

func TestOfflineCache_GetOrSetUserDevices(t *testing.T) {
	cache, mr := setupTestCache(t)
	defer mr.Close()

	ctx := context.Background()
	userID := uuid.New()
	devices := []*domain.Device{createTestDeviceForCache(), createTestDeviceForCache()}
	loaderCalled := false

	loader := func() ([]*domain.Device, error) {
		loaderCalled = true
		return devices, nil
	}

	// First call - should call loader
	result, err := cache.GetOrSetUserDevices(ctx, userID, loader)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, loaderCalled)
	assert.Len(t, result, 2)

	// Reset flag
	loaderCalled = false

	// Second call - should use cache
	result, err = cache.GetOrSetUserDevices(ctx, userID, loader)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.False(t, loaderCalled) // Loader should NOT be called
	assert.Len(t, result, 2)
}

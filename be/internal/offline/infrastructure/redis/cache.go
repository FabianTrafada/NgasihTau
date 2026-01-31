// Package redis provides Redis-based caching implementations for the Offline Material Service.
package redis

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	"ngasihtau/internal/offline/domain"
)

// OfflineCache provides caching operations for the offline material service.
// Implements cache-aside pattern for licenses, devices, and rate limiting.
type OfflineCache struct {
	client *redis.Client
}

// NewOfflineCache creates a new OfflineCache instance.
func NewOfflineCache(client *redis.Client) *OfflineCache {
	return &OfflineCache{client: client}
}

// Ping checks if Redis is reachable.
func (c *OfflineCache) Ping(ctx context.Context) error {
	return c.client.Ping(ctx).Err()
}

// Close closes the Redis connection.
func (c *OfflineCache) Close() error {
	return c.client.Close()
}

// Redis returns the underlying Redis client for advanced operations.
func (c *OfflineCache) Redis() *redis.Client {
	return c.client
}

// ============================================================================
// License Caching (5 min TTL)
// ============================================================================

// licenseKey generates the cache key for a license.
func licenseKey(licenseID uuid.UUID) string {
	return domain.RedisKeyPrefixLicense + licenseID.String()
}

// licenseByUserMaterialDeviceKey generates the cache key for license lookup.
func licenseByUserMaterialDeviceKey(userID, materialID, deviceID uuid.UUID) string {
	return fmt.Sprintf("%slookup:%s:%s:%s", domain.RedisKeyPrefixLicense, userID.String(), materialID.String(), deviceID.String())
}

// GetLicense retrieves a license from cache.
// Returns nil if not found in cache.
func (c *OfflineCache) GetLicense(ctx context.Context, licenseID uuid.UUID) (*domain.License, error) {
	key := licenseKey(licenseID)
	data, err := c.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("cache get license error: %w", err)
	}

	var license domain.License
	if err := json.Unmarshal(data, &license); err != nil {
		return nil, fmt.Errorf("cache unmarshal license error: %w", err)
	}

	return &license, nil
}

// SetLicense stores a license in cache with 5 min TTL.
func (c *OfflineCache) SetLicense(ctx context.Context, license *domain.License) error {
	key := licenseKey(license.ID)
	data, err := json.Marshal(license)
	if err != nil {
		return fmt.Errorf("cache marshal license error: %w", err)
	}

	if err := c.client.Set(ctx, key, data, domain.LicenseCacheTTL).Err(); err != nil {
		return fmt.Errorf("cache set license error: %w", err)
	}

	// Also cache the lookup key
	lookupKey := licenseByUserMaterialDeviceKey(license.UserID, license.MaterialID, license.DeviceID)
	if err := c.client.Set(ctx, lookupKey, license.ID.String(), domain.LicenseCacheTTL).Err(); err != nil {
		// Non-fatal, just log
		return nil
	}

	return nil
}


// GetLicenseByUserMaterialDevice retrieves a license ID from cache by user, material, and device.
func (c *OfflineCache) GetLicenseByUserMaterialDevice(ctx context.Context, userID, materialID, deviceID uuid.UUID) (*uuid.UUID, error) {
	lookupKey := licenseByUserMaterialDeviceKey(userID, materialID, deviceID)
	idStr, err := c.client.Get(ctx, lookupKey).Result()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("cache get license lookup error: %w", err)
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		return nil, fmt.Errorf("cache parse license ID error: %w", err)
	}

	return &id, nil
}

// InvalidateLicense removes a license from cache.
func (c *OfflineCache) InvalidateLicense(ctx context.Context, license *domain.License) error {
	keys := []string{
		licenseKey(license.ID),
		licenseByUserMaterialDeviceKey(license.UserID, license.MaterialID, license.DeviceID),
	}
	if err := c.client.Del(ctx, keys...).Err(); err != nil {
		return fmt.Errorf("cache invalidate license error: %w", err)
	}
	return nil
}

// InvalidateLicenseByID removes a license from cache by ID only.
func (c *OfflineCache) InvalidateLicenseByID(ctx context.Context, licenseID uuid.UUID) error {
	key := licenseKey(licenseID)
	if err := c.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("cache invalidate license error: %w", err)
	}
	return nil
}

// InvalidateLicensesByDevice removes all licenses for a device from cache.
func (c *OfflineCache) InvalidateLicensesByDevice(ctx context.Context, deviceID uuid.UUID) error {
	pattern := fmt.Sprintf("%s*%s*", domain.RedisKeyPrefixLicense, deviceID.String())
	return c.deleteByPattern(ctx, pattern)
}

// InvalidateLicensesByMaterial removes all licenses for a material from cache.
func (c *OfflineCache) InvalidateLicensesByMaterial(ctx context.Context, materialID uuid.UUID) error {
	pattern := fmt.Sprintf("%s*%s*", domain.RedisKeyPrefixLicense, materialID.String())
	return c.deleteByPattern(ctx, pattern)
}

// ============================================================================
// Device Caching (10 min TTL)
// ============================================================================

// deviceKey generates the cache key for a device.
func deviceKey(deviceID uuid.UUID) string {
	return domain.RedisKeyPrefixDevice + deviceID.String()
}

// deviceByFingerprintKey generates the cache key for device lookup by fingerprint.
func deviceByFingerprintKey(userID uuid.UUID, fingerprint string) string {
	return fmt.Sprintf("%sfp:%s:%s", domain.RedisKeyPrefixDevice, userID.String(), fingerprint)
}

// GetDevice retrieves a device from cache.
// Returns nil if not found in cache.
func (c *OfflineCache) GetDevice(ctx context.Context, deviceID uuid.UUID) (*domain.Device, error) {
	key := deviceKey(deviceID)
	data, err := c.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("cache get device error: %w", err)
	}

	var device domain.Device
	if err := json.Unmarshal(data, &device); err != nil {
		return nil, fmt.Errorf("cache unmarshal device error: %w", err)
	}

	return &device, nil
}

// SetDevice stores a device in cache with 10 min TTL.
func (c *OfflineCache) SetDevice(ctx context.Context, device *domain.Device) error {
	key := deviceKey(device.ID)
	data, err := json.Marshal(device)
	if err != nil {
		return fmt.Errorf("cache marshal device error: %w", err)
	}

	if err := c.client.Set(ctx, key, data, domain.DeviceCacheTTL).Err(); err != nil {
		return fmt.Errorf("cache set device error: %w", err)
	}

	// Also cache the fingerprint lookup
	fpKey := deviceByFingerprintKey(device.UserID, device.Fingerprint)
	if err := c.client.Set(ctx, fpKey, device.ID.String(), domain.DeviceCacheTTL).Err(); err != nil {
		// Non-fatal
		return nil
	}

	return nil
}

// GetDeviceByFingerprint retrieves a device ID from cache by user and fingerprint.
func (c *OfflineCache) GetDeviceByFingerprint(ctx context.Context, userID uuid.UUID, fingerprint string) (*uuid.UUID, error) {
	fpKey := deviceByFingerprintKey(userID, fingerprint)
	idStr, err := c.client.Get(ctx, fpKey).Result()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("cache get device fingerprint error: %w", err)
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		return nil, fmt.Errorf("cache parse device ID error: %w", err)
	}

	return &id, nil
}

// InvalidateDevice removes a device from cache.
func (c *OfflineCache) InvalidateDevice(ctx context.Context, device *domain.Device) error {
	keys := []string{
		deviceKey(device.ID),
		deviceByFingerprintKey(device.UserID, device.Fingerprint),
	}
	if err := c.client.Del(ctx, keys...).Err(); err != nil {
		return fmt.Errorf("cache invalidate device error: %w", err)
	}
	return nil
}

// InvalidateDeviceByID removes a device from cache by ID only.
func (c *OfflineCache) InvalidateDeviceByID(ctx context.Context, deviceID uuid.UUID) error {
	key := deviceKey(deviceID)
	if err := c.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("cache invalidate device error: %w", err)
	}
	return nil
}


// ============================================================================
// User Device List Caching
// ============================================================================

// userDevicesKey generates the cache key for a user's device list.
func userDevicesKey(userID uuid.UUID) string {
	return domain.RedisKeyPrefixUserDevices + userID.String()
}

// GetUserDevices retrieves a user's device list from cache.
// Returns nil if not found in cache.
func (c *OfflineCache) GetUserDevices(ctx context.Context, userID uuid.UUID) ([]*domain.Device, error) {
	key := userDevicesKey(userID)
	data, err := c.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("cache get user devices error: %w", err)
	}

	var devices []*domain.Device
	if err := json.Unmarshal(data, &devices); err != nil {
		return nil, fmt.Errorf("cache unmarshal user devices error: %w", err)
	}

	return devices, nil
}

// SetUserDevices stores a user's device list in cache with 10 min TTL.
func (c *OfflineCache) SetUserDevices(ctx context.Context, userID uuid.UUID, devices []*domain.Device) error {
	key := userDevicesKey(userID)
	data, err := json.Marshal(devices)
	if err != nil {
		return fmt.Errorf("cache marshal user devices error: %w", err)
	}

	if err := c.client.Set(ctx, key, data, domain.DeviceCacheTTL).Err(); err != nil {
		return fmt.Errorf("cache set user devices error: %w", err)
	}

	return nil
}

// InvalidateUserDevices removes a user's device list from cache.
func (c *OfflineCache) InvalidateUserDevices(ctx context.Context, userID uuid.UUID) error {
	key := userDevicesKey(userID)
	if err := c.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("cache invalidate user devices error: %w", err)
	}
	return nil
}

// ============================================================================
// Rate Limiting Counters
// ============================================================================

// downloadCountKey generates the cache key for download rate limiting.
func downloadCountKey(userID uuid.UUID) string {
	return domain.RedisKeyPrefixDownloadCount + userID.String()
}

// validationFailureKey generates the cache key for validation failure tracking.
func validationFailureKey(deviceID uuid.UUID) string {
	return domain.RedisKeyPrefixValidationFailure + deviceID.String()
}

// GetDownloadCount retrieves the current download count for a user.
func (c *OfflineCache) GetDownloadCount(ctx context.Context, userID uuid.UUID) (int, error) {
	key := downloadCountKey(userID)
	count, err := c.client.Get(ctx, key).Int()
	if err == redis.Nil {
		return 0, nil
	}
	if err != nil {
		return 0, fmt.Errorf("cache get download count error: %w", err)
	}
	return count, nil
}

// IncrementDownloadCount increments the download count for a user.
// Sets TTL to 1 hour for automatic reset.
func (c *OfflineCache) IncrementDownloadCount(ctx context.Context, userID uuid.UUID) (int, error) {
	key := downloadCountKey(userID)
	pipe := c.client.Pipeline()
	incr := pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, domain.RateLimitCacheTTL)
	_, err := pipe.Exec(ctx)
	if err != nil {
		return 0, fmt.Errorf("cache increment download count error: %w", err)
	}
	return int(incr.Val()), nil
}

// GetValidationFailureCount retrieves the validation failure count for a device.
func (c *OfflineCache) GetValidationFailureCount(ctx context.Context, deviceID uuid.UUID) (int, error) {
	key := validationFailureKey(deviceID)
	count, err := c.client.Get(ctx, key).Int()
	if err == redis.Nil {
		return 0, nil
	}
	if err != nil {
		return 0, fmt.Errorf("cache get validation failure count error: %w", err)
	}
	return count, nil
}

// IncrementValidationFailure increments the validation failure count for a device.
// Sets TTL to 1 hour for automatic reset.
func (c *OfflineCache) IncrementValidationFailure(ctx context.Context, deviceID uuid.UUID) (int, error) {
	key := validationFailureKey(deviceID)
	pipe := c.client.Pipeline()
	incr := pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, domain.RateLimitCacheTTL)
	_, err := pipe.Exec(ctx)
	if err != nil {
		return 0, fmt.Errorf("cache increment validation failure error: %w", err)
	}
	return int(incr.Val()), nil
}

// ResetValidationFailures resets the validation failure count for a device.
func (c *OfflineCache) ResetValidationFailures(ctx context.Context, deviceID uuid.UUID) error {
	key := validationFailureKey(deviceID)
	if err := c.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("cache reset validation failures error: %w", err)
	}
	return nil
}

// IsDeviceBlocked checks if a device is blocked due to too many validation failures.
func (c *OfflineCache) IsDeviceBlocked(ctx context.Context, deviceID uuid.UUID) (bool, error) {
	count, err := c.GetValidationFailureCount(ctx, deviceID)
	if err != nil {
		return false, err
	}
	return count >= domain.MaxValidationFailuresPerHour, nil
}

// ============================================================================
// Helper Methods
// ============================================================================

// deleteByPattern removes all keys matching the pattern.
func (c *OfflineCache) deleteByPattern(ctx context.Context, pattern string) error {
	var cursor uint64
	var keys []string

	for {
		var err error
		var batch []string
		batch, cursor, err = c.client.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			return fmt.Errorf("cache scan error: %w", err)
		}
		keys = append(keys, batch...)
		if cursor == 0 {
			break
		}
	}

	if len(keys) > 0 {
		if err := c.client.Del(ctx, keys...).Err(); err != nil {
			return fmt.Errorf("cache delete error: %w", err)
		}
	}

	return nil
}

// GetOrSetLicense implements cache-aside pattern for licenses.
func (c *OfflineCache) GetOrSetLicense(ctx context.Context, licenseID uuid.UUID, loader func() (*domain.License, error)) (*domain.License, error) {
	// Try cache first
	license, err := c.GetLicense(ctx, licenseID)
	if err != nil {
		// Log error but continue to loader
		_ = err
	}
	if license != nil {
		return license, nil
	}

	// Cache miss - call loader
	license, err = loader()
	if err != nil {
		return nil, err
	}
	if license == nil {
		return nil, nil
	}

	// Store in cache (fire and forget)
	_ = c.SetLicense(ctx, license)

	return license, nil
}

// GetOrSetDevice implements cache-aside pattern for devices.
func (c *OfflineCache) GetOrSetDevice(ctx context.Context, deviceID uuid.UUID, loader func() (*domain.Device, error)) (*domain.Device, error) {
	// Try cache first
	device, err := c.GetDevice(ctx, deviceID)
	if err != nil {
		// Log error but continue to loader
		_ = err
	}
	if device != nil {
		return device, nil
	}

	// Cache miss - call loader
	device, err = loader()
	if err != nil {
		return nil, err
	}
	if device == nil {
		return nil, nil
	}

	// Store in cache (fire and forget)
	_ = c.SetDevice(ctx, device)

	return device, nil
}

// GetOrSetUserDevices implements cache-aside pattern for user device lists.
func (c *OfflineCache) GetOrSetUserDevices(ctx context.Context, userID uuid.UUID, loader func() ([]*domain.Device, error)) ([]*domain.Device, error) {
	// Try cache first
	devices, err := c.GetUserDevices(ctx, userID)
	if err != nil {
		// Log error but continue to loader
		_ = err
	}
	if devices != nil {
		return devices, nil
	}

	// Cache miss - call loader
	devices, err = loader()
	if err != nil {
		return nil, err
	}

	// Store in cache (fire and forget)
	_ = c.SetUserDevices(ctx, userID, devices)

	return devices, nil
}

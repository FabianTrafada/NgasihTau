// Package cache provides a Redis-based caching layer with cache-aside pattern.
// Implements requirement 24: Caching Strategy.
package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"runtime"
	"time"

	"github.com/redis/go-redis/v9"
)

// Config holds Redis cache configuration.
type Config struct {
	Host         string
	Port         int
	Password     string
	DB           int
	PoolSize     int
	MinIdleConns int
	PoolTimeout  time.Duration
	IdleTimeout  time.Duration
}

// DefaultConfig returns default cache configuration.
func DefaultConfig() Config {
	return Config{
		Host:         "localhost",
		Port:         6379,
		Password:     "",
		DB:           0,
		PoolSize:     10,
		MinIdleConns: 5,
		PoolTimeout:  4 * time.Second,
		IdleTimeout:  5 * time.Minute,
	}
}

// TTL constants for different data types.
// Based on design document caching strategy.
const (
	// TTLUserProfile is the TTL for user profile data (1 hour).
	TTLUserProfile = 1 * time.Hour
	// TTLUserSession is the TTL for user session data (15 minutes).
	TTLUserSession = 15 * time.Minute
	// TTLPodDetails is the TTL for pod details (30 minutes).
	TTLPodDetails = 30 * time.Minute
	// TTLPodCollaborators is the TTL for pod collaborators list (15 minutes).
	TTLPodCollaborators = 15 * time.Minute
	// TTLMaterialMeta is the TTL for material metadata (1 hour).
	TTLMaterialMeta = 1 * time.Hour
	// TTLSearchResults is the TTL for search results (5 minutes).
	TTLSearchResults = 5 * time.Minute
	// TTLAutocompleteSuggestions is the TTL for autocomplete suggestions (10 minutes).
	TTLAutocompleteSuggestions = 10 * time.Minute
	// TTLTrendingMaterials is the TTL for trending materials (15 minutes).
	TTLTrendingMaterials = 15 * time.Minute
)

// Key prefixes for cache key naming convention.
// Format: {service}:{entity}:{identifier}:{field?}
const (
	// PrefixUserProfile is the prefix for user profile cache keys.
	PrefixUserProfile = "user:profile:"
	// PrefixUserSession is the prefix for user session cache keys.
	PrefixUserSession = "user:session:"
	// PrefixPodDetails is the prefix for pod details cache keys.
	PrefixPodDetails = "pod:details:"
	// PrefixPodCollaborators is the prefix for pod collaborators cache keys.
	PrefixPodCollaborators = "pod:collaborators:"
	// PrefixMaterialMeta is the prefix for material metadata cache keys.
	PrefixMaterialMeta = "material:meta:"
	// PrefixSearchResults is the prefix for search results cache keys.
	PrefixSearchResults = "search:results:"
	// PrefixSearchSuggestions is the prefix for search suggestions cache keys.
	PrefixSearchSuggestions = "search:suggestions:"
	// PrefixRateLimitUser is the prefix for user rate limit counters.
	PrefixRateLimitUser = "rate:user:"
	// PrefixRateLimitIP is the prefix for IP rate limit counters.
	PrefixRateLimitIP = "rate:ip:"
)

// Client wraps Redis client with caching utilities.
type Client struct {
	rdb *redis.Client
}

// NewClient creates a new cache client.
func NewClient(cfg Config) (*Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:            fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password:        cfg.Password,
		DB:              cfg.DB,
		PoolSize:        cfg.PoolSize * runtime.NumCPU(),
		MinIdleConns:    cfg.MinIdleConns,
		PoolTimeout:     cfg.PoolTimeout,
		ConnMaxIdleTime: cfg.IdleTimeout,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &Client{rdb: rdb}, nil
}

// Close closes the Redis connection.
func (c *Client) Close() error {
	return c.rdb.Close()
}

// Ping checks if Redis is reachable.
func (c *Client) Ping(ctx context.Context) error {
	return c.rdb.Ping(ctx).Err()
}

// Redis returns the underlying Redis client for advanced operations.
func (c *Client) Redis() *redis.Client {
	return c.rdb
}

// Get retrieves a value from cache and unmarshals it into the target.
// Returns false if the key doesn't exist.
func (c *Client) Get(ctx context.Context, key string, target interface{}) (bool, error) {
	data, err := c.rdb.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("cache get error: %w", err)
	}

	if err := json.Unmarshal(data, target); err != nil {
		return false, fmt.Errorf("cache unmarshal error: %w", err)
	}

	return true, nil
}

// Set stores a value in cache with the specified TTL.
func (c *Client) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("cache marshal error: %w", err)
	}

	if err := c.rdb.Set(ctx, key, data, ttl).Err(); err != nil {
		return fmt.Errorf("cache set error: %w", err)
	}

	return nil
}

// Delete removes a key from cache.
func (c *Client) Delete(ctx context.Context, keys ...string) error {
	if len(keys) == 0 {
		return nil
	}
	if err := c.rdb.Del(ctx, keys...).Err(); err != nil {
		return fmt.Errorf("cache delete error: %w", err)
	}
	return nil
}

// DeleteByPattern removes all keys matching the pattern.
// Use with caution as SCAN can be slow on large datasets.
func (c *Client) DeleteByPattern(ctx context.Context, pattern string) error {
	var cursor uint64
	var keys []string

	for {
		var err error
		var batch []string
		batch, cursor, err = c.rdb.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			return fmt.Errorf("cache scan error: %w", err)
		}
		keys = append(keys, batch...)
		if cursor == 0 {
			break
		}
	}

	if len(keys) > 0 {
		if err := c.rdb.Del(ctx, keys...).Err(); err != nil {
			return fmt.Errorf("cache delete error: %w", err)
		}
	}

	return nil
}

// Exists checks if a key exists in cache.
func (c *Client) Exists(ctx context.Context, key string) (bool, error) {
	n, err := c.rdb.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("cache exists error: %w", err)
	}
	return n > 0, nil
}

// SetNX sets a value only if the key doesn't exist (for distributed locks).
func (c *Client) SetNX(ctx context.Context, key string, value interface{}, ttl time.Duration) (bool, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return false, fmt.Errorf("cache marshal error: %w", err)
	}

	ok, err := c.rdb.SetNX(ctx, key, data, ttl).Result()
	if err != nil {
		return false, fmt.Errorf("cache setnx error: %w", err)
	}

	return ok, nil
}

// GetOrSet implements the cache-aside pattern.
// It first tries to get the value from cache. If not found, it calls the loader
// function to fetch the data, stores it in cache, and returns it.
func (c *Client) GetOrSet(ctx context.Context, key string, target interface{}, ttl time.Duration, loader func() (interface{}, error)) error {
	// Try cache first
	found, getErr := c.Get(ctx, key, target)
	if getErr != nil {
		// Log error but continue to loader (cache miss behavior)
		// Silently ignore cache errors and fall back to loader
		_ = getErr
	}
	if found {
		return nil
	}

	// Cache miss - call loader
	data, err := loader()
	if err != nil {
		return err
	}

	// Store in cache (fire and forget, don't fail if cache write fails)
	_ = c.Set(ctx, key, data, ttl)

	// Marshal and unmarshal to populate target
	// This ensures consistent behavior whether data comes from cache or loader
	bytes, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal loader result: %w", err)
	}
	if err := json.Unmarshal(bytes, target); err != nil {
		return fmt.Errorf("failed to unmarshal loader result: %w", err)
	}

	return nil
}

// UserProfileKey returns the cache key for a user profile.
func UserProfileKey(userID string) string {
	return PrefixUserProfile + userID
}

// UserSessionKey returns the cache key for a user session.
func UserSessionKey(userID string) string {
	return PrefixUserSession + userID
}

// PodDetailsKey returns the cache key for pod details.
func PodDetailsKey(podID string) string {
	return PrefixPodDetails + podID
}

// PodCollaboratorsKey returns the cache key for pod collaborators.
func PodCollaboratorsKey(podID string) string {
	return PrefixPodCollaborators + podID
}

// MaterialMetaKey returns the cache key for material metadata.
func MaterialMetaKey(materialID string) string {
	return PrefixMaterialMeta + materialID
}

// SearchResultsKey returns the cache key for search results.
func SearchResultsKey(queryHash string) string {
	return PrefixSearchResults + queryHash
}

// SearchSuggestionsKey returns the cache key for search suggestions.
func SearchSuggestionsKey(prefix string) string {
	return PrefixSearchSuggestions + prefix
}

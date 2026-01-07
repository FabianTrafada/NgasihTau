// Package redis provides Redis implementations of the User Service repositories.
package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	"ngasihtau/internal/user/domain"
)

// AIUsageRepository implements domain.AIUsageRepository using Redis.
// It tracks daily AI message counts per user with automatic expiry at midnight UTC.
// Implements requirements 9.3, 9.6.
type AIUsageRepository struct {
	client *redis.Client
}

// NewAIUsageRepository creates a new AIUsageRepository.
func NewAIUsageRepository(client *redis.Client) *AIUsageRepository {
	return &AIUsageRepository{client: client}
}

// keyForUser generates the Redis key for a user's daily AI usage.
// Key pattern: ai_usage:{userID}:{date}
func (r *AIUsageRepository) keyForUser(userID uuid.UUID) string {
	date := time.Now().UTC().Format("2006-01-02")
	return fmt.Sprintf("ai_usage:%s:%s", userID.String(), date)
}

// nextMidnightUTC returns the time of the next midnight UTC.
func nextMidnightUTC() time.Time {
	now := time.Now().UTC()
	return time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, time.UTC)
}

// GetDailyUsage returns today's AI message count for a user.
// Returns 0 if no usage exists (handles redis.Nil as 0 count).
// Implements requirement 9.3.
func (r *AIUsageRepository) GetDailyUsage(ctx context.Context, userID uuid.UUID) (int, error) {
	key := r.keyForUser(userID)
	count, err := r.client.Get(ctx, key).Int()
	if err == redis.Nil {
		return 0, nil
	}
	if err != nil {
		return 0, fmt.Errorf("failed to get daily AI usage: %w", err)
	}
	return count, nil
}

// IncrementDailyUsage increments the daily AI message count for a user.
// Uses INCR and sets ExpireAt to next midnight UTC for automatic reset.
// Implements requirements 9.3, 9.6.
func (r *AIUsageRepository) IncrementDailyUsage(ctx context.Context, userID uuid.UUID) error {
	key := r.keyForUser(userID)
	pipe := r.client.Pipeline()
	pipe.Incr(ctx, key)
	pipe.ExpireAt(ctx, key, nextMidnightUTC())
	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to increment daily AI usage: %w", err)
	}
	return nil
}

// Compile-time interface implementation check.
var _ domain.AIUsageRepository = (*AIUsageRepository)(nil)

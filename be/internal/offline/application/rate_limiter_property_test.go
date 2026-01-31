package application

import (
	"context"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/google/uuid"
	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
	"github.com/redis/go-redis/v9"

	"ngasihtau/internal/offline/domain"
)

// TestProperty19_DownloadRateLimiting tests Property 19: Download Rate Limiting.
// Property: Users cannot exceed MaxDownloadsPerHour downloads within the rate limit window.
func TestProperty19_DownloadRateLimiting(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 50
	parameters.MaxSize = 20

	properties := gopter.NewProperties(parameters)

	// Property 19.1: Download count starts at zero for new users
	properties.Property("download count starts at zero for new users", prop.ForAll(
		func(userIDBytes []byte) bool {
			mr, _ := miniredis.Run()
			defer mr.Close()
			client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
			defer client.Close()

			limiter := NewRateLimiter(client)
			ctx := context.Background()

			// Create a deterministic UUID from bytes
			userID := uuid.New()

			allowed, remaining, _, err := limiter.CheckDownloadLimit(ctx, userID)
			if err != nil {
				return false
			}

			return allowed && remaining == domain.MaxDownloadsPerHour
		},
		gen.SliceOfN(16, gen.UInt8()),
	))

	// Property 19.2: Download count increments correctly
	properties.Property("download count increments correctly", prop.ForAll(
		func(numDownloads int) bool {
			if numDownloads < 1 || numDownloads > 20 {
				return true // Skip invalid inputs
			}

			mr, _ := miniredis.Run()
			defer mr.Close()
			client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
			defer client.Close()

			limiter := NewRateLimiter(client)
			ctx := context.Background()
			userID := uuid.New()

			for i := 0; i < numDownloads; i++ {
				count, err := limiter.IncrementDownload(ctx, userID)
				if err != nil {
					return false
				}
				if count != i+1 {
					return false
				}
			}

			_, remaining, _, err := limiter.CheckDownloadLimit(ctx, userID)
			if err != nil {
				return false
			}

			expectedRemaining := domain.MaxDownloadsPerHour - numDownloads
			if expectedRemaining < 0 {
				expectedRemaining = 0
			}

			return remaining == expectedRemaining
		},
		gen.IntRange(1, 15),
	))

	// Property 19.3: Rate limit is enforced at MaxDownloadsPerHour
	properties.Property("rate limit is enforced at max downloads", prop.ForAll(
		func(extraDownloads int) bool {
			mr, _ := miniredis.Run()
			defer mr.Close()
			client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
			defer client.Close()

			limiter := NewRateLimiter(client)
			ctx := context.Background()
			userID := uuid.New()

			// Exhaust the limit
			for i := 0; i < domain.MaxDownloadsPerHour; i++ {
				_, err := limiter.IncrementDownload(ctx, userID)
				if err != nil {
					return false
				}
			}

			// Check that limit is enforced
			allowed, remaining, _, err := limiter.CheckDownloadLimit(ctx, userID)
			if err != nil {
				return false
			}

			return !allowed && remaining == 0
		},
		gen.IntRange(1, 5),
	))

	// Property 19.4: Different users have independent rate limits
	properties.Property("different users have independent rate limits", prop.ForAll(
		func(user1Downloads, user2Downloads int) bool {
			if user1Downloads < 0 || user1Downloads > 15 || user2Downloads < 0 || user2Downloads > 15 {
				return true
			}

			mr, _ := miniredis.Run()
			defer mr.Close()
			client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
			defer client.Close()

			limiter := NewRateLimiter(client)
			ctx := context.Background()
			user1 := uuid.New()
			user2 := uuid.New()

			// Increment for user1
			for i := 0; i < user1Downloads; i++ {
				_, err := limiter.IncrementDownload(ctx, user1)
				if err != nil {
					return false
				}
			}

			// Increment for user2
			for i := 0; i < user2Downloads; i++ {
				_, err := limiter.IncrementDownload(ctx, user2)
				if err != nil {
					return false
				}
			}

			// Check user1
			_, remaining1, _, err := limiter.CheckDownloadLimit(ctx, user1)
			if err != nil {
				return false
			}

			// Check user2
			_, remaining2, _, err := limiter.CheckDownloadLimit(ctx, user2)
			if err != nil {
				return false
			}

			expected1 := domain.MaxDownloadsPerHour - user1Downloads
			if expected1 < 0 {
				expected1 = 0
			}
			expected2 := domain.MaxDownloadsPerHour - user2Downloads
			if expected2 < 0 {
				expected2 = 0
			}

			return remaining1 == expected1 && remaining2 == expected2
		},
		gen.IntRange(0, 15),
		gen.IntRange(0, 15),
	))

	// Property 19.5: Rate limit info is consistent
	properties.Property("rate limit info is consistent with check", prop.ForAll(
		func(numDownloads int) bool {
			if numDownloads < 0 || numDownloads > 15 {
				return true
			}

			mr, _ := miniredis.Run()
			defer mr.Close()
			client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
			defer client.Close()

			limiter := NewRateLimiter(client)
			ctx := context.Background()
			userID := uuid.New()

			for i := 0; i < numDownloads; i++ {
				_, err := limiter.IncrementDownload(ctx, userID)
				if err != nil {
					return false
				}
			}

			info, err := limiter.GetRateLimitInfo(ctx, userID)
			if err != nil {
				return false
			}

			_, remaining, _, err := limiter.CheckDownloadLimit(ctx, userID)
			if err != nil {
				return false
			}

			return info.Limit == domain.MaxDownloadsPerHour && info.Remaining == remaining
		},
		gen.IntRange(0, 15),
	))

	properties.TestingRun(t)
}

// TestProperty28_DeviceBlockingOnFailures tests Property 28: Device Blocking on Failures.
// Property: Devices are blocked after MaxValidationFailuresPerHour consecutive failures.
func TestProperty28_DeviceBlockingOnFailures(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 50
	parameters.MaxSize = 10

	properties := gopter.NewProperties(parameters)

	// Property 28.1: Device is not blocked initially
	properties.Property("device is not blocked initially", prop.ForAll(
		func(deviceIDBytes []byte) bool {
			mr, _ := miniredis.Run()
			defer mr.Close()
			client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
			defer client.Close()

			limiter := NewRateLimiter(client)
			ctx := context.Background()
			deviceID := uuid.New()

			isBlocked, _, err := limiter.IsDeviceBlocked(ctx, deviceID)
			if err != nil {
				return false
			}

			return !isBlocked
		},
		gen.SliceOfN(16, gen.UInt8()),
	))

	// Property 28.2: Failure count increments correctly
	properties.Property("failure count increments correctly", prop.ForAll(
		func(numFailures int) bool {
			if numFailures < 1 || numFailures > domain.MaxValidationFailuresPerHour-1 {
				return true
			}

			mr, _ := miniredis.Run()
			defer mr.Close()
			client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
			defer client.Close()

			limiter := NewRateLimiter(client)
			ctx := context.Background()
			deviceID := uuid.New()

			for i := 0; i < numFailures; i++ {
				count, blocked, err := limiter.RecordValidationFailure(ctx, deviceID)
				if err != nil {
					return false
				}
				if count != i+1 {
					return false
				}
				// Should not be blocked until threshold
				if blocked && count < domain.MaxValidationFailuresPerHour {
					return false
				}
			}

			return true
		},
		gen.IntRange(1, domain.MaxValidationFailuresPerHour-1),
	))

	// Property 28.3: Device is blocked after max failures
	properties.Property("device is blocked after max failures", prop.ForAll(
		func(extraFailures int) bool {
			mr, _ := miniredis.Run()
			defer mr.Close()
			client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
			defer client.Close()

			limiter := NewRateLimiter(client)
			ctx := context.Background()
			deviceID := uuid.New()

			// Record exactly MaxValidationFailuresPerHour failures
			for i := 0; i < domain.MaxValidationFailuresPerHour; i++ {
				_, _, err := limiter.RecordValidationFailure(ctx, deviceID)
				if err != nil {
					return false
				}
			}

			// Device should now be blocked
			isBlocked, _, err := limiter.IsDeviceBlocked(ctx, deviceID)
			if err != nil {
				return false
			}

			return isBlocked
		},
		gen.IntRange(0, 3),
	))

	// Property 28.4: Unblocking device clears block status
	properties.Property("unblocking device clears block status", prop.ForAll(
		func(dummy int) bool {
			mr, _ := miniredis.Run()
			defer mr.Close()
			client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
			defer client.Close()

			limiter := NewRateLimiter(client)
			ctx := context.Background()
			deviceID := uuid.New()

			// Block the device
			for i := 0; i < domain.MaxValidationFailuresPerHour; i++ {
				_, _, err := limiter.RecordValidationFailure(ctx, deviceID)
				if err != nil {
					return false
				}
			}

			// Verify blocked
			isBlocked, _, err := limiter.IsDeviceBlocked(ctx, deviceID)
			if err != nil || !isBlocked {
				return false
			}

			// Unblock
			err = limiter.UnblockDevice(ctx, deviceID)
			if err != nil {
				return false
			}

			// Verify unblocked
			isBlocked, _, err = limiter.IsDeviceBlocked(ctx, deviceID)
			if err != nil {
				return false
			}

			return !isBlocked
		},
		gen.IntRange(0, 1),
	))

	// Property 28.5: Different devices have independent failure counts
	properties.Property("different devices have independent failure counts", prop.ForAll(
		func(device1Failures, device2Failures int) bool {
			if device1Failures < 0 || device1Failures > 10 || device2Failures < 0 || device2Failures > 10 {
				return true
			}

			mr, _ := miniredis.Run()
			defer mr.Close()
			client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
			defer client.Close()

			limiter := NewRateLimiter(client)
			ctx := context.Background()
			device1 := uuid.New()
			device2 := uuid.New()

			// Record failures for device1
			for i := 0; i < device1Failures; i++ {
				_, _, err := limiter.RecordValidationFailure(ctx, device1)
				if err != nil {
					return false
				}
			}

			// Record failures for device2
			for i := 0; i < device2Failures; i++ {
				_, _, err := limiter.RecordValidationFailure(ctx, device2)
				if err != nil {
					return false
				}
			}

			// Check block status
			blocked1, _, _ := limiter.IsDeviceBlocked(ctx, device1)
			blocked2, _, _ := limiter.IsDeviceBlocked(ctx, device2)

			expected1 := device1Failures >= domain.MaxValidationFailuresPerHour
			expected2 := device2Failures >= domain.MaxValidationFailuresPerHour

			return blocked1 == expected1 && blocked2 == expected2
		},
		gen.IntRange(0, 10),
		gen.IntRange(0, 10),
	))

	properties.TestingRun(t)
}

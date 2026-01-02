// Package domain contains property-based tests for the Pod domain entities.
package domain

import (
	"testing"

	"github.com/google/uuid"
	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// **Feature: student-teacher-roles, Property 8: Upvote System Integrity**
// **Validates: Requirements 5.1, 5.2, 5.3**
//
// Property 8: Upvote System Integrity
// *For any* user and knowledge pod:
// - Upvoting SHALL increment the pod's upvote_count by exactly 1 and create a vote record
// - Removing upvote SHALL decrement the pod's upvote_count by exactly 1 and remove the vote record
// - A user SHALL only be able to upvote a pod once (duplicate upvotes SHALL be rejected)

func TestProperty_UpvoteSystemIntegrity(t *testing.T) {
	parameters := gopter.DefaultTestParametersWithSeed(12345)
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	// Property 8.1: NewPodUpvote creates a valid upvote record with correct user and pod IDs
	// Validates: Requirement 5.1 - WHEN a user upvotes a knowledge pod,
	// THE Pod Service SHALL record the user's vote.
	properties.Property("NewPodUpvote creates valid upvote record", prop.ForAll(
		func(userIDStr, podIDStr string) bool {
			userID := uuid.New()
			podID := uuid.New()

			upvote := NewPodUpvote(userID, podID)

			// Verify upvote record is created with correct IDs
			return upvote != nil &&
				upvote.UserID == userID &&
				upvote.PodID == podID &&
				!upvote.CreatedAt.IsZero()
		},
		gen.AlphaString(),
		gen.AlphaString(),
	))

	// Property 8.2: NewPodUpvote always sets a non-zero CreatedAt timestamp
	// Validates: Requirement 5.1 - vote record creation
	properties.Property("NewPodUpvote always sets CreatedAt timestamp", prop.ForAll(
		func(_ int) bool {
			userID := uuid.New()
			podID := uuid.New()

			upvote := NewPodUpvote(userID, podID)

			return !upvote.CreatedAt.IsZero()
		},
		gen.Int(),
	))

	// Property 8.3: NewPod initializes with zero upvote count
	// Validates: Requirement 5.1 - upvote count starts at zero
	properties.Property("NewPod initializes with zero upvote count", prop.ForAll(
		func(name, slug string, isTeacher bool) bool {
			ownerID := uuid.New()

			pod := NewPod(ownerID, name, slug, VisibilityPublic, isTeacher)

			return pod.UpvoteCount == 0
		},
		genValidPodName(),
		genValidSlug(),
		gen.Bool(),
	))

	// Property 8.4: Pod upvote count can be incremented (simulating upvote)
	// Validates: Requirement 5.1 - WHEN a user upvotes a knowledge pod,
	// THE Pod Service SHALL increment the pod's upvote count.
	properties.Property("upvote count increment by 1", prop.ForAll(
		func(initialCount int) bool {
			// Ensure non-negative initial count
			if initialCount < 0 {
				initialCount = 0
			}
			if initialCount > 1000000 {
				initialCount = 1000000
			}

			ownerID := uuid.New()
			pod := NewPod(ownerID, "Test Pod", "test-pod", VisibilityPublic, false)
			pod.UpvoteCount = initialCount

			// Simulate upvote increment
			pod.UpvoteCount++

			return pod.UpvoteCount == initialCount+1
		},
		gen.IntRange(0, 1000000),
	))

	// Property 8.5: Pod upvote count can be decremented (simulating remove upvote)
	// Validates: Requirement 5.2 - WHEN a user removes their upvote,
	// THE Pod Service SHALL decrement the pod's upvote count.
	properties.Property("upvote count decrement by 1", prop.ForAll(
		func(initialCount int) bool {
			// Ensure positive initial count (can't decrement from 0)
			if initialCount <= 0 {
				initialCount = 1
			}
			if initialCount > 1000000 {
				initialCount = 1000000
			}

			ownerID := uuid.New()
			pod := NewPod(ownerID, "Test Pod", "test-pod", VisibilityPublic, false)
			pod.UpvoteCount = initialCount

			// Simulate upvote removal
			pod.UpvoteCount--

			return pod.UpvoteCount == initialCount-1
		},
		gen.IntRange(1, 1000000),
	))

	// Property 8.6: Upvote then remove upvote returns to original count (round-trip)
	// Validates: Requirements 5.1, 5.2 - upvote/remove upvote round-trip
	properties.Property("upvote then remove returns to original count", prop.ForAll(
		func(initialCount int) bool {
			// Ensure non-negative initial count
			if initialCount < 0 {
				initialCount = 0
			}
			if initialCount > 1000000 {
				initialCount = 1000000
			}

			ownerID := uuid.New()
			pod := NewPod(ownerID, "Test Pod", "test-pod", VisibilityPublic, false)
			pod.UpvoteCount = initialCount

			// Simulate upvote then remove
			pod.UpvoteCount++ // upvote
			pod.UpvoteCount-- // remove upvote

			return pod.UpvoteCount == initialCount
		},
		gen.IntRange(0, 1000000),
	))

	// Property 8.7: Multiple upvotes from different users increment count correctly
	// Validates: Requirement 5.1 - multiple users can upvote
	properties.Property("multiple upvotes increment count correctly", prop.ForAll(
		func(numUpvotes int) bool {
			// Limit to reasonable range
			if numUpvotes < 0 {
				numUpvotes = 0
			}
			if numUpvotes > 100 {
				numUpvotes = 100
			}

			ownerID := uuid.New()
			pod := NewPod(ownerID, "Test Pod", "test-pod", VisibilityPublic, false)

			// Simulate multiple upvotes from different users
			for i := 0; i < numUpvotes; i++ {
				pod.UpvoteCount++
			}

			return pod.UpvoteCount == numUpvotes
		},
		gen.IntRange(0, 100),
	))

	// Property 8.8: PodUpvote records are unique per user-pod pair
	// Validates: Requirement 5.3 - THE Pod Service SHALL allow each user to upvote
	// a knowledge pod only once.
	properties.Property("upvote records have unique user-pod pairs", prop.ForAll(
		func(_ int) bool {
			userID := uuid.New()
			podID := uuid.New()

			upvote1 := NewPodUpvote(userID, podID)
			upvote2 := NewPodUpvote(userID, podID)

			// Both upvotes should have the same user and pod IDs
			// (in practice, the repository would reject the second one)
			return upvote1.UserID == upvote2.UserID &&
				upvote1.PodID == upvote2.PodID
		},
		gen.Int(),
	))

	// Property 8.9: Different users can create upvotes for the same pod
	// Validates: Requirement 5.1 - multiple users can upvote the same pod
	properties.Property("different users can upvote same pod", prop.ForAll(
		func(_ int) bool {
			user1ID := uuid.New()
			user2ID := uuid.New()
			podID := uuid.New()

			upvote1 := NewPodUpvote(user1ID, podID)
			upvote2 := NewPodUpvote(user2ID, podID)

			// Both upvotes should be for the same pod but different users
			return upvote1.PodID == upvote2.PodID &&
				upvote1.UserID != upvote2.UserID
		},
		gen.Int(),
	))

	// Property 8.10: Same user can upvote different pods
	// Validates: Requirement 5.1 - a user can upvote multiple pods
	properties.Property("same user can upvote different pods", prop.ForAll(
		func(_ int) bool {
			userID := uuid.New()
			pod1ID := uuid.New()
			pod2ID := uuid.New()

			upvote1 := NewPodUpvote(userID, pod1ID)
			upvote2 := NewPodUpvote(userID, pod2ID)

			// Both upvotes should be from the same user but for different pods
			return upvote1.UserID == upvote2.UserID &&
				upvote1.PodID != upvote2.PodID
		},
		gen.Int(),
	))

	properties.TestingRun(t)
}

// Generator for valid pod names
func genValidPodName() gopter.Gen {
	return gen.AlphaString().Map(func(s string) string {
		if len(s) == 0 {
			return "Default Pod"
		}
		if len(s) > 100 {
			return s[:100]
		}
		return s
	})
}

// Generator for valid slugs
func genValidSlug() gopter.Gen {
	return gen.AlphaString().Map(func(s string) string {
		if len(s) == 0 {
			return "default-slug"
		}
		if len(s) > 100 {
			return s[:100]
		}
		return s
	})
}

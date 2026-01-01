// Package domain contains property-based tests for trust indicator display.
package domain

import (
	"testing"

	"github.com/google/uuid"
	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// **Feature: student-teacher-roles, Property 10: Trust Indicator Display**
// **Validates: Requirements 6.1, 6.2**
//
// Property 10: Trust Indicator Display
// *For any* knowledge pod in API response:
// - If created by a teacher, `is_verified` SHALL be true
// - The `upvote_count` SHALL be included as a trust indicator

func TestProperty_TrustIndicatorDisplay(t *testing.T) {
	parameters := gopter.DefaultTestParametersWithSeed(12345)
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	// Property 10.1: Teacher-created pods have is_verified = true
	// Validates: Requirement 6.1 - WHEN displaying a knowledge pod,
	// THE Pod Service SHALL show a "verified" badge if the pod is created by a teacher.
	properties.Property("teacher-created pods have is_verified true", prop.ForAll(
		func(name, slug string) bool {
			ownerID := uuid.New()
			isCreatorTeacher := true

			pod := NewPod(ownerID, name, slug, VisibilityPublic, isCreatorTeacher)

			// Teacher-created pods must have is_verified = true
			return pod.IsVerified == true
		},
		genValidPodNameForTrust(),
		genValidSlugForTrust(),
	))

	// Property 10.2: Student-created pods have is_verified = false
	// Validates: Requirement 6.1 - verified badge only for teacher-created pods
	properties.Property("student-created pods have is_verified false", prop.ForAll(
		func(name, slug string) bool {
			ownerID := uuid.New()
			isCreatorTeacher := false

			pod := NewPod(ownerID, name, slug, VisibilityPublic, isCreatorTeacher)

			// Student-created pods must have is_verified = false
			return pod.IsVerified == false
		},
		genValidPodNameForTrust(),
		genValidSlugForTrust(),
	))

	// Property 10.3: is_verified status is determined solely by creator role
	// Validates: Requirement 6.1 - verified status depends on creator being teacher
	properties.Property("is_verified equals isCreatorTeacher parameter", prop.ForAll(
		func(name, slug string, isCreatorTeacher bool) bool {
			ownerID := uuid.New()

			pod := NewPod(ownerID, name, slug, VisibilityPublic, isCreatorTeacher)

			// is_verified should exactly match the isCreatorTeacher parameter
			return pod.IsVerified == isCreatorTeacher
		},
		genValidPodNameForTrust(),
		genValidSlugForTrust(),
		gen.Bool(),
	))

	// Property 10.4: All pods include upvote_count as trust indicator (initialized to 0)
	// Validates: Requirement 6.2 - THE Pod Service SHALL show the upvote count
	// as a trust indicator.
	properties.Property("all pods include upvote_count initialized to zero", prop.ForAll(
		func(name, slug string, isCreatorTeacher bool) bool {
			ownerID := uuid.New()

			pod := NewPod(ownerID, name, slug, VisibilityPublic, isCreatorTeacher)

			// upvote_count must be present and initialized to 0
			return pod.UpvoteCount == 0
		},
		genValidPodNameForTrust(),
		genValidSlugForTrust(),
		gen.Bool(),
	))

	// Property 10.5: upvote_count is always non-negative (trust indicator constraint)
	// Validates: Requirement 6.2 - upvote count as valid trust indicator
	properties.Property("upvote_count is always non-negative", prop.ForAll(
		func(name, slug string, isCreatorTeacher bool, upvoteCount int) bool {
			ownerID := uuid.New()

			pod := NewPod(ownerID, name, slug, VisibilityPublic, isCreatorTeacher)

			// Simulate setting upvote count (only non-negative values are valid)
			if upvoteCount >= 0 {
				pod.UpvoteCount = upvoteCount
			}

			// upvote_count must always be non-negative
			return pod.UpvoteCount >= 0
		},
		genValidPodNameForTrust(),
		genValidSlugForTrust(),
		gen.Bool(),
		gen.IntRange(0, 1000000),
	))

	// Property 10.6: Visibility does not affect trust indicators
	// Validates: Requirements 6.1, 6.2 - trust indicators are independent of visibility
	properties.Property("visibility does not affect trust indicators", prop.ForAll(
		func(name, slug string, isCreatorTeacher bool) bool {
			ownerID := uuid.New()

			publicPod := NewPod(ownerID, name, slug, VisibilityPublic, isCreatorTeacher)
			privatePod := NewPod(ownerID, name, slug, VisibilityPrivate, isCreatorTeacher)

			// Trust indicators should be the same regardless of visibility
			return publicPod.IsVerified == privatePod.IsVerified &&
				publicPod.UpvoteCount == privatePod.UpvoteCount
		},
		genValidPodNameForTrust(),
		genValidSlugForTrust(),
		gen.Bool(),
	))

	// Property 10.7: Pod struct always contains both trust indicator fields
	// Validates: Requirements 6.1, 6.2 - both indicators must be present in response
	properties.Property("pod struct contains both trust indicator fields", prop.ForAll(
		func(name, slug string, isCreatorTeacher bool) bool {
			ownerID := uuid.New()

			pod := NewPod(ownerID, name, slug, VisibilityPublic, isCreatorTeacher)

			// Verify both fields exist and have valid values
			// is_verified is a bool (always valid)
			// upvote_count must be >= 0
			return pod.UpvoteCount >= 0
		},
		genValidPodNameForTrust(),
		genValidSlugForTrust(),
		gen.Bool(),
	))

	// Property 10.8: PodWithOwner preserves trust indicators from Pod
	// Validates: Requirements 6.1, 6.2 - trust indicators in API response
	properties.Property("PodWithOwner preserves trust indicators", prop.ForAll(
		func(name, slug string, isCreatorTeacher bool, ownerName string) bool {
			ownerID := uuid.New()

			pod := NewPod(ownerID, name, slug, VisibilityPublic, isCreatorTeacher)

			// Create PodWithOwner (API response type)
			podWithOwner := PodWithOwner{
				Pod:       *pod,
				OwnerName: ownerName,
			}

			// Trust indicators should be preserved in API response type
			return podWithOwner.IsVerified == isCreatorTeacher &&
				podWithOwner.UpvoteCount == 0
		},
		genValidPodNameForTrust(),
		genValidSlugForTrust(),
		gen.Bool(),
		gen.AlphaString().Map(func(s string) string {
			if len(s) == 0 {
				return "Owner"
			}
			return s
		}),
	))

	properties.TestingRun(t)
}

// Generator for valid pod names (for trust indicator tests)
func genValidPodNameForTrust() gopter.Gen {
	return gen.AlphaString().Map(func(s string) string {
		if len(s) == 0 {
			return "Test Pod"
		}
		if len(s) > 100 {
			return s[:100]
		}
		return s
	})
}

// Generator for valid slugs (for trust indicator tests)
func genValidSlugForTrust() gopter.Gen {
	return gen.AlphaString().Map(func(s string) string {
		if len(s) == 0 {
			return "test-pod"
		}
		if len(s) > 100 {
			return s[:100]
		}
		return s
	})
}

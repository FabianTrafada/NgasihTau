// Package domain contains property-based tests for teacher pod creation.
package domain

import (
	"testing"

	"github.com/google/uuid"
	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"

	userDomain "ngasihtau/internal/user/domain"
)

// **Feature: student-teacher-roles, Property 4: Teacher Pod Creation with Verified Status**
// **Validates: Requirements 2.3, 2.4**
//
// Property 4: Teacher Pod Creation with Verified Status
// *For any* user with "teacher" role creating a knowledge pod:
// - The teacher SHALL be able to create pods with verified status
// - The resulting pod SHALL automatically have `is_verified = true`

func TestProperty_TeacherPodCreationWithVerifiedStatus(t *testing.T) {
	parameters := gopter.DefaultTestParametersWithSeed(12345)
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	// Property 4.1: Teachers can create knowledge pods with verified status
	// Validates: Requirement 2.3 - WHILE a user has "teacher" role,
	// THE Pod Service SHALL allow the user to create knowledge pods with "verified" status.
	properties.Property("teachers can create knowledge pods with verified status", prop.ForAll(
		func(name, slug string, visibility Visibility) bool {
			// Create a teacher user (simulating a user who has been verified)
			teacher := userDomain.NewUser("teacher@example.com", "hash", "Teacher User")
			teacher.Role = userDomain.RoleTeacher

			// Verify user is a teacher
			if teacher.Role != userDomain.RoleTeacher {
				return false
			}

			// Teacher creates a pod (isCreatorTeacher = true for teachers)
			pod := NewPod(teacher.ID, name, slug, visibility, true)

			// Pod should be created successfully with the teacher as owner
			// and with verified status
			return pod != nil &&
				pod.OwnerID == teacher.ID &&
				pod.Name == name &&
				pod.Slug == slug &&
				pod.Visibility == visibility &&
				pod.IsVerified == true
		},
		genValidTeacherPodName(),
		genValidTeacherSlug(),
		genTeacherVisibility(),
	))

	// Property 4.2: Pods created by teachers automatically have is_verified = true
	// Validates: Requirement 2.4 - WHEN a teacher creates a knowledge pod,
	// THE Pod Service SHALL automatically mark the pod with "verified" badge.
	properties.Property("pods created by teachers automatically have is_verified true", prop.ForAll(
		func(name, slug string, visibility Visibility) bool {
			teacherID := uuid.New()

			// Teacher creates a pod (isCreatorTeacher = true)
			pod := NewPod(teacherID, name, slug, visibility, true)

			// The pod must have is_verified = true
			return pod.IsVerified == true
		},
		genValidTeacherPodName(),
		genValidTeacherSlug(),
		genTeacherVisibility(),
	))

	// Property 4.3: Teacher role is the determining factor for verified status
	// Validates: Requirements 2.3, 2.4 - verified status is based on creator role
	properties.Property("isCreatorTeacher=true always results in is_verified=true", prop.ForAll(
		func(name, slug string, visibility Visibility) bool {
			ownerID := uuid.New()

			// Create pod with isCreatorTeacher = true
			pod := NewPod(ownerID, name, slug, visibility, true)

			// is_verified must be true when isCreatorTeacher is true
			return pod.IsVerified == true
		},
		genValidTeacherPodName(),
		genValidTeacherSlug(),
		genTeacherVisibility(),
	))

	// Property 4.4: Teacher-created pods have all standard fields initialized correctly
	// Validates: Requirement 2.3 - teachers can create pods with all standard features
	properties.Property("teacher-created pods have all fields initialized correctly", prop.ForAll(
		func(name, slug string, visibility Visibility) bool {
			teacherID := uuid.New()

			pod := NewPod(teacherID, name, slug, visibility, true)

			// Verify all fields are initialized correctly
			return pod != nil &&
				pod.ID != uuid.Nil &&
				pod.OwnerID == teacherID &&
				pod.Name == name &&
				pod.Slug == slug &&
				pod.Visibility == visibility &&
				pod.IsVerified == true &&
				pod.StarCount == 0 &&
				pod.ForkCount == 0 &&
				pod.ViewCount == 0 &&
				pod.UpvoteCount == 0 &&
				!pod.CreatedAt.IsZero() &&
				!pod.UpdatedAt.IsZero()
		},
		genValidTeacherPodName(),
		genValidTeacherSlug(),
		genTeacherVisibility(),
	))

	// Property 4.5: Teacher-created pods can have any visibility setting
	// Validates: Requirement 2.3 - teachers can create pods with any visibility
	properties.Property("teacher-created pods support all visibility settings", prop.ForAll(
		func(name, slug string) bool {
			teacherID := uuid.New()

			publicPod := NewPod(teacherID, name, slug, VisibilityPublic, true)
			privatePod := NewPod(teacherID, name, slug, VisibilityPrivate, true)

			// Both visibility settings should work and both should be verified
			return publicPod.IsVerified == true &&
				privatePod.IsVerified == true &&
				publicPod.Visibility == VisibilityPublic &&
				privatePod.Visibility == VisibilityPrivate
		},
		genValidTeacherPodName(),
		genValidTeacherSlug(),
	))

	// Property 4.6: Verified status is independent of other pod attributes
	// Validates: Requirement 2.4 - verified badge is automatic for teacher-created pods
	properties.Property("verified status is independent of pod name and slug", prop.ForAll(
		func(name1, slug1, name2, slug2 string) bool {
			teacherID := uuid.New()

			pod1 := NewPod(teacherID, name1, slug1, VisibilityPublic, true)
			pod2 := NewPod(teacherID, name2, slug2, VisibilityPublic, true)

			// Both pods should be verified regardless of name/slug
			return pod1.IsVerified == true && pod2.IsVerified == true
		},
		genValidTeacherPodName(),
		genValidTeacherSlug(),
		genValidTeacherPodName(),
		genValidTeacherSlug(),
	))

	// Property 4.7: Teacher-created pods have unique IDs
	// Validates: Requirement 2.3 - each pod created is a distinct entity
	properties.Property("each teacher-created pod has a unique ID", prop.ForAll(
		func(name, slug string) bool {
			teacherID := uuid.New()

			pod1 := NewPod(teacherID, name, slug, VisibilityPublic, true)
			pod2 := NewPod(teacherID, name, slug, VisibilityPublic, true)

			// Each pod should have a unique ID
			return pod1.ID != pod2.ID
		},
		genValidTeacherPodName(),
		genValidTeacherSlug(),
	))

	// Property 4.8: Contrast with student-created pods (is_verified = false)
	// Validates: Requirements 2.3, 2.4 - teacher pods are verified, student pods are not
	properties.Property("teacher pods are verified while student pods are not", prop.ForAll(
		func(name, slug string) bool {
			ownerID := uuid.New()

			teacherPod := NewPod(ownerID, name, slug, VisibilityPublic, true)  // Teacher
			studentPod := NewPod(ownerID, name, slug, VisibilityPublic, false) // Student

			// Teacher pod should be verified, student pod should not
			return teacherPod.IsVerified == true && studentPod.IsVerified == false
		},
		genValidTeacherPodName(),
		genValidTeacherSlug(),
	))

	properties.TestingRun(t)
}

// Generator for valid pod names (for teacher pod creation tests)
func genValidTeacherPodName() gopter.Gen {
	return gen.AlphaString().Map(func(s string) string {
		if len(s) == 0 {
			return "Teacher Pod"
		}
		if len(s) > 100 {
			return s[:100]
		}
		return s
	})
}

// Generator for valid slugs (for teacher pod creation tests)
func genValidTeacherSlug() gopter.Gen {
	return gen.AlphaString().Map(func(s string) string {
		if len(s) == 0 {
			return "teacher-pod"
		}
		if len(s) > 100 {
			return s[:100]
		}
		return s
	})
}

// Generator for visibility settings
func genTeacherVisibility() gopter.Gen {
	return gen.OneConstOf(VisibilityPublic, VisibilityPrivate)
}

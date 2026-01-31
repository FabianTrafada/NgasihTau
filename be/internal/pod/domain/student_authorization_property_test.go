// Package domain contains property-based tests for student authorization constraints.
package domain

import (
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"

	userDomain "ngasihtau/internal/user/domain"
)

// **Feature: student-teacher-roles, Property 2: Student Authorization Constraints**
// **Validates: Requirements 1.2, 1.3, 1.4**
//
// Property 2: Student Authorization Constraints
// *For any* user with "student" role and any knowledge pod:
// - The student SHALL be able to create new knowledge pods
// - The student SHALL only be able to upload materials to pods they own
// - Any pod created by the student SHALL have `is_verified = false`

func TestProperty_StudentAuthorizationConstraints(t *testing.T) {
	parameters := gopter.DefaultTestParametersWithSeed(12345)
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	// Property 2.1: Students can create knowledge pods
	// Validates: Requirement 1.2 - WHILE a user has "student" role,
	// THE Pod Service SHALL allow the user to create knowledge pods.
	properties.Property("students can create knowledge pods", prop.ForAll(
		func(name, slug string, visibility Visibility) bool {
			// Create a student user
			student := userDomain.NewUser("student@example.com", "hash", "Student User")
			// Verify user is a student
			if student.Role != userDomain.RoleStudent {
				return false
			}

			// Student creates a pod (isCreatorTeacher = false for students)
			pod := NewPod(student.ID, name, slug, visibility, false)

			// Pod should be created successfully with the student as owner
			return pod != nil &&
				pod.OwnerID == student.ID &&
				pod.Name == name &&
				pod.Slug == slug &&
				pod.Visibility == visibility
		},
		genValidStudentPodName(),
		genValidStudentSlug(),
		genVisibility(),
	))

	// Property 2.2: Pods created by students are unverified
	// Validates: Requirement 1.4 - WHEN a student creates a knowledge pod,
	// THE Pod Service SHALL mark the pod with "unverified" status.
	properties.Property("pods created by students are unverified", prop.ForAll(
		func(name, slug string, visibility Visibility) bool {
			// Create a student user
			student := userDomain.NewUser("student@example.com", "hash", "Student User")

			// Student creates a pod (isCreatorTeacher = false for students)
			pod := NewPod(student.ID, name, slug, visibility, false)

			// Pod must have is_verified = false
			return pod.IsVerified == false
		},
		genValidStudentPodName(),
		genValidStudentSlug(),
		genVisibility(),
	))

	// Property 2.3: Students can only upload to pods they own
	// Validates: Requirement 1.3 - WHILE a user has "student" role,
	// THE Material Service SHALL allow the user to upload materials only to
	// knowledge pods owned by that user.
	properties.Property("student is owner of their created pod", prop.ForAll(
		func(name, slug string, visibility Visibility) bool {
			// Create a student user
			student := userDomain.NewUser("student@example.com", "hash", "Student User")

			// Student creates a pod
			pod := NewPod(student.ID, name, slug, visibility, false)

			// Student should be the owner of the pod
			return pod.IsOwner(student.ID)
		},
		genValidStudentPodName(),
		genValidStudentSlug(),
		genVisibility(),
	))

	// Property 2.4: Student is not owner of other users' pods
	// Validates: Requirement 1.3 - students can only upload to their own pods
	properties.Property("student is not owner of other users pods", prop.ForAll(
		func(name, slug string, visibility Visibility) bool {
			// Create two student users
			student1 := userDomain.NewUser("student1@example.com", "hash", "Student 1")
			student2 := userDomain.NewUser("student2@example.com", "hash", "Student 2")

			// Student1 creates a pod
			pod := NewPod(student1.ID, name, slug, visibility, false)

			// Student2 should NOT be the owner of student1's pod
			return !pod.IsOwner(student2.ID)
		},
		genValidStudentPodName(),
		genValidStudentSlug(),
		genVisibility(),
	))

	properties.TestingRun(t)
}

// Generator for valid pod names (for student authorization tests)
func genValidStudentPodName() gopter.Gen {
	return gen.AlphaString().Map(func(s string) string {
		if len(s) == 0 {
			return "Student Pod"
		}
		if len(s) > 100 {
			return s[:100]
		}
		return s
	})
}

// Generator for valid slugs (for student authorization tests)
func genValidStudentSlug() gopter.Gen {
	return gen.AlphaString().Map(func(s string) string {
		if len(s) == 0 {
			return "student-pod"
		}
		if len(s) > 100 {
			return s[:100]
		}
		return s
	})
}

// Generator for visibility settings (for student authorization tests)
func genVisibility() gopter.Gen {
	return gen.OneConstOf(VisibilityPublic, VisibilityPrivate)
}

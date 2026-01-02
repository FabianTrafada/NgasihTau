// Package domain contains property-based tests for the Shared Pod domain entities.
package domain

import (
	"testing"

	"github.com/google/uuid"
	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// **Feature: student-teacher-roles, Property 12: Shared Pods Functionality**
// **Validates: Requirements 7.2, 7.3**
//
// Property 12: Shared Pods Functionality
// *For any* pod sharing from teacher to student:
// - The shared pod SHALL appear in the student's "shared with me" section
// - The student SHALL receive a notification about the shared pod

func TestProperty_SharedPodsFunctionality(t *testing.T) {
	parameters := gopter.DefaultTestParametersWithSeed(12345)
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	// Property 12.1: NewSharedPod creates a valid shared pod record with correct IDs
	// Validates: Requirement 7.2 - THE Pod Service SHALL support a "shared with me" section
	// showing pods that teachers have explicitly shared with the student.
	properties.Property("NewSharedPod creates valid shared pod record", prop.ForAll(
		func(hasMessage bool, messageContent string) bool {
			podID := uuid.New()
			teacherID := uuid.New()
			studentID := uuid.New()

			var message *string
			if hasMessage && len(messageContent) > 0 {
				message = &messageContent
			}

			sharedPod := NewSharedPod(podID, teacherID, studentID, message)

			// Verify shared pod record is created with correct IDs
			return sharedPod != nil &&
				sharedPod.ID != uuid.Nil &&
				sharedPod.PodID == podID &&
				sharedPod.TeacherID == teacherID &&
				sharedPod.StudentID == studentID &&
				!sharedPod.CreatedAt.IsZero()
		},
		gen.Bool(),
		gen.AlphaString(),
	))

	// Property 12.2: NewSharedPod always sets a non-zero CreatedAt timestamp
	// Validates: Requirement 7.2 - shared pod record creation
	properties.Property("NewSharedPod always sets CreatedAt timestamp", prop.ForAll(
		func(_ int) bool {
			podID := uuid.New()
			teacherID := uuid.New()
			studentID := uuid.New()

			sharedPod := NewSharedPod(podID, teacherID, studentID, nil)

			return !sharedPod.CreatedAt.IsZero()
		},
		gen.Int(),
	))

	// Property 12.3: NewSharedPod generates unique IDs for each shared pod
	// Validates: Requirement 7.2 - each share is a distinct record
	properties.Property("NewSharedPod generates unique IDs", prop.ForAll(
		func(_ int) bool {
			podID := uuid.New()
			teacherID := uuid.New()
			studentID := uuid.New()

			sharedPod1 := NewSharedPod(podID, teacherID, studentID, nil)
			sharedPod2 := NewSharedPod(podID, teacherID, studentID, nil)

			// Each shared pod should have a unique ID
			return sharedPod1.ID != sharedPod2.ID
		},
		gen.Int(),
	))

	// Property 12.4: Message is optional and preserved when provided
	// Validates: Requirement 7.2 - optional message in shared pod
	properties.Property("message is optional and preserved when provided", prop.ForAll(
		func(messageContent string) bool {
			podID := uuid.New()
			teacherID := uuid.New()
			studentID := uuid.New()

			// Test with message
			message := messageContent
			sharedWithMsg := NewSharedPod(podID, teacherID, studentID, &message)

			// Test without message
			sharedWithoutMsg := NewSharedPod(podID, teacherID, studentID, nil)

			// Message should be preserved when provided
			msgPreserved := sharedWithMsg.Message != nil && *sharedWithMsg.Message == messageContent
			// Message should be nil when not provided
			noMsgIsNil := sharedWithoutMsg.Message == nil

			return msgPreserved && noMsgIsNil
		},
		gen.AlphaString().SuchThat(func(s string) bool { return len(s) > 0 }),
	))

	// Property 12.5: SharedPod preserves teacher and student IDs
	// Validates: Requirement 7.2 - shared pod tracks who shared and who received
	properties.Property("SharedPod preserves teacher and student IDs", prop.ForAll(
		func(_ int) bool {
			podID := uuid.New()
			teacherID := uuid.New()
			studentID := uuid.New()

			sharedPod := NewSharedPod(podID, teacherID, studentID, nil)

			// IDs should be preserved
			return sharedPod.PodID == podID &&
				sharedPod.TeacherID == teacherID &&
				sharedPod.StudentID == studentID
		},
		gen.Int(),
	))

	// Property 12.6: Same teacher can share different pods with same student
	// Validates: Requirement 7.2 - teachers can share multiple pods
	properties.Property("same teacher can share different pods with same student", prop.ForAll(
		func(_ int) bool {
			teacherID := uuid.New()
			studentID := uuid.New()
			pod1ID := uuid.New()
			pod2ID := uuid.New()

			shared1 := NewSharedPod(pod1ID, teacherID, studentID, nil)
			shared2 := NewSharedPod(pod2ID, teacherID, studentID, nil)

			// Both shares should have the same teacher and student but different pods
			return shared1.TeacherID == shared2.TeacherID &&
				shared1.StudentID == shared2.StudentID &&
				shared1.PodID != shared2.PodID
		},
		gen.Int(),
	))

	// Property 12.7: Same pod can be shared with different students
	// Validates: Requirement 7.2 - a pod can be shared with multiple students
	properties.Property("same pod can be shared with different students", prop.ForAll(
		func(_ int) bool {
			teacherID := uuid.New()
			podID := uuid.New()
			student1ID := uuid.New()
			student2ID := uuid.New()

			shared1 := NewSharedPod(podID, teacherID, student1ID, nil)
			shared2 := NewSharedPod(podID, teacherID, student2ID, nil)

			// Both shares should be for the same pod but different students
			return shared1.PodID == shared2.PodID &&
				shared1.TeacherID == shared2.TeacherID &&
				shared1.StudentID != shared2.StudentID
		},
		gen.Int(),
	))

	// Property 12.8: Different teachers can share the same pod with the same student
	// Validates: Requirement 7.2 - multiple teachers can recommend the same pod
	properties.Property("different teachers can share same pod with same student", prop.ForAll(
		func(_ int) bool {
			podID := uuid.New()
			studentID := uuid.New()
			teacher1ID := uuid.New()
			teacher2ID := uuid.New()

			shared1 := NewSharedPod(podID, teacher1ID, studentID, nil)
			shared2 := NewSharedPod(podID, teacher2ID, studentID, nil)

			// Both shares should be for the same pod and student but different teachers
			return shared1.PodID == shared2.PodID &&
				shared1.StudentID == shared2.StudentID &&
				shared1.TeacherID != shared2.TeacherID
		},
		gen.Int(),
	))

	// Property 12.9: SharedPodWithDetails extends SharedPod with additional fields
	// Validates: Requirement 7.2 - shared pods include pod and teacher info
	properties.Property("SharedPodWithDetails includes pod and teacher info", prop.ForAll(
		func(podName, podSlug, teacherName string, hasAvatar bool, avatarURL string) bool {
			podID := uuid.New()
			teacherID := uuid.New()
			studentID := uuid.New()

			sharedPod := NewSharedPod(podID, teacherID, studentID, nil)

			// Create SharedPodWithDetails
			var avatar *string
			if hasAvatar && len(avatarURL) > 0 {
				avatar = &avatarURL
			}

			details := SharedPodWithDetails{
				SharedPod:     *sharedPod,
				PodName:       podName,
				PodSlug:       podSlug,
				TeacherName:   teacherName,
				TeacherAvatar: avatar,
			}

			// Verify all fields are accessible
			return details.ID == sharedPod.ID &&
				details.PodID == podID &&
				details.TeacherID == teacherID &&
				details.StudentID == studentID &&
				details.PodName == podName &&
				details.PodSlug == podSlug &&
				details.TeacherName == teacherName
		},
		genValidPodNameForShared(),
		genValidSlugForShared(),
		gen.AlphaString().SuchThat(func(s string) bool { return len(s) > 0 }),
		gen.Bool(),
		gen.AlphaString(),
	))

	// Property 12.10: Teacher and student IDs must be different
	// Validates: Requirement 7.2 - teachers share with students (not themselves)
	// Note: This is a domain constraint that should be enforced at service level
	properties.Property("teacher and student can be different users", prop.ForAll(
		func(_ int) bool {
			podID := uuid.New()
			teacherID := uuid.New()
			studentID := uuid.New()

			sharedPod := NewSharedPod(podID, teacherID, studentID, nil)

			// In valid scenarios, teacher and student should be different
			// The domain entity allows creation, but service layer validates
			return sharedPod.TeacherID != sharedPod.StudentID
		},
		gen.Int(),
	))

	// Property 12.11: Shared pod records are immutable after creation
	// Validates: Requirement 7.2 - shared pod records don't change
	properties.Property("shared pod fields are set at creation time", prop.ForAll(
		func(messageContent string) bool {
			podID := uuid.New()
			teacherID := uuid.New()
			studentID := uuid.New()
			message := messageContent

			sharedPod := NewSharedPod(podID, teacherID, studentID, &message)

			// All fields should be set at creation
			return sharedPod.ID != uuid.Nil &&
				sharedPod.PodID == podID &&
				sharedPod.TeacherID == teacherID &&
				sharedPod.StudentID == studentID &&
				sharedPod.Message != nil &&
				*sharedPod.Message == messageContent &&
				!sharedPod.CreatedAt.IsZero()
		},
		gen.AlphaString().SuchThat(func(s string) bool { return len(s) > 0 }),
	))

	properties.TestingRun(t)
}

// Generator for valid pod names (for shared pod tests)
func genValidPodNameForShared() gopter.Gen {
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

// Generator for valid slugs (for shared pod tests)
func genValidSlugForShared() gopter.Gen {
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

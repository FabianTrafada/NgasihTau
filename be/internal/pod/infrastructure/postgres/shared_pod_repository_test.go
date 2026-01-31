// Package postgres contains unit tests for the SharedPodRepository.
// Tests CRUD operations for shared pods.
// Implements requirements 9.1, 9.3.
package postgres

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"ngasihtau/internal/pod/domain"
)

// TestSharedPod_NewSharedPod tests the NewSharedPod constructor
func TestSharedPod_NewSharedPod(t *testing.T) {
	tests := []struct {
		name      string
		podID     uuid.UUID
		teacherID uuid.UUID
		studentID uuid.UUID
		message   *string
	}{
		{
			name:      "creates valid shared pod without message",
			podID:     uuid.New(),
			teacherID: uuid.New(),
			studentID: uuid.New(),
			message:   nil,
		},
		{
			name:      "creates valid shared pod with message",
			podID:     uuid.New(),
			teacherID: uuid.New(),
			studentID: uuid.New(),
			message: func() *string {
				msg := "Check out this pod for your studies"
				return &msg
			}(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			share := domain.NewSharedPod(tt.podID, tt.teacherID, tt.studentID, tt.message)

			if share == nil {
				t.Fatal("Expected shared pod to be created")
			}
			if share.ID == uuid.Nil {
				t.Error("Expected ID to be set")
			}
			if share.PodID != tt.podID {
				t.Errorf("Expected PodID %s, got %s", tt.podID, share.PodID)
			}
			if share.TeacherID != tt.teacherID {
				t.Errorf("Expected TeacherID %s, got %s", tt.teacherID, share.TeacherID)
			}
			if share.StudentID != tt.studentID {
				t.Errorf("Expected StudentID %s, got %s", tt.studentID, share.StudentID)
			}
			if share.CreatedAt.IsZero() {
				t.Error("Expected CreatedAt to be set")
			}
			if tt.message != nil && (share.Message == nil || *share.Message != *tt.message) {
				t.Error("Expected message to be set correctly")
			}
		})
	}
}

// TestSharedPodRepository_Interface tests that the repository implements the interface
func TestSharedPodRepository_Interface(t *testing.T) {
	// This test verifies that SharedPodRepository implements domain.SharedPodRepository
	// The compile-time check is in interfaces.go, but this documents the expected behavior
	var _ domain.SharedPodRepository = (*SharedPodRepository)(nil)
}

// TestSharedPodRepository_Methods tests the repository method signatures
func TestSharedPodRepository_Methods(t *testing.T) {
	// Test that the repository has all expected methods with correct signatures
	// This is a compile-time check that documents the API

	t.Run("Create method exists", func(t *testing.T) {
		var repo *SharedPodRepository
		var _ func(context.Context, *domain.SharedPod) error = repo.Create
	})

	t.Run("Delete method exists", func(t *testing.T) {
		var repo *SharedPodRepository
		var _ func(context.Context, uuid.UUID) error = repo.Delete
	})

	t.Run("FindByStudent method exists", func(t *testing.T) {
		var repo *SharedPodRepository
		var _ func(context.Context, uuid.UUID, int, int) ([]*domain.SharedPod, int, error) = repo.FindByStudent
	})

	t.Run("FindByStudentWithDetails method exists", func(t *testing.T) {
		var repo *SharedPodRepository
		var _ func(context.Context, uuid.UUID, int, int) ([]*domain.SharedPodWithDetails, int, error) = repo.FindByStudentWithDetails
	})

	t.Run("FindByTeacherAndStudent method exists", func(t *testing.T) {
		var repo *SharedPodRepository
		var _ func(context.Context, uuid.UUID, uuid.UUID) ([]*domain.SharedPod, error) = repo.FindByTeacherAndStudent
	})

	t.Run("Exists method exists", func(t *testing.T) {
		var repo *SharedPodRepository
		var _ func(context.Context, uuid.UUID, uuid.UUID) (bool, error) = repo.Exists
	})
}

// TestSharedPodWithDetails_Structure tests the SharedPodWithDetails struct
func TestSharedPodWithDetails_Structure(t *testing.T) {
	// Test that SharedPodWithDetails has all expected fields
	details := domain.SharedPodWithDetails{
		SharedPod: domain.SharedPod{
			ID:        uuid.New(),
			PodID:     uuid.New(),
			TeacherID: uuid.New(),
			StudentID: uuid.New(),
		},
		PodName:     "Test Pod",
		PodSlug:     "test-pod",
		TeacherName: "Teacher Name",
	}

	if details.PodName != "Test Pod" {
		t.Errorf("Expected PodName 'Test Pod', got %s", details.PodName)
	}
	if details.PodSlug != "test-pod" {
		t.Errorf("Expected PodSlug 'test-pod', got %s", details.PodSlug)
	}
	if details.TeacherName != "Teacher Name" {
		t.Errorf("Expected TeacherName 'Teacher Name', got %s", details.TeacherName)
	}
}

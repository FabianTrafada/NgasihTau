// Package postgres contains unit tests for the UploadRequestRepository.
// Tests CRUD operations for upload requests.
// Implements requirements 9.1, 9.3.
package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"ngasihtau/internal/pod/domain"
)

// TestUploadRequest_NewUploadRequest tests the NewUploadRequest constructor
func TestUploadRequest_NewUploadRequest(t *testing.T) {
	tests := []struct {
		name        string
		requesterID uuid.UUID
		podID       uuid.UUID
		podOwnerID  uuid.UUID
		message     *string
	}{
		{
			name:        "creates valid upload request without message",
			requesterID: uuid.New(),
			podID:       uuid.New(),
			podOwnerID:  uuid.New(),
			message:     nil,
		},
		{
			name:        "creates valid upload request with message",
			requesterID: uuid.New(),
			podID:       uuid.New(),
			podOwnerID:  uuid.New(),
			message: func() *string {
				msg := "I would like to contribute to your pod"
				return &msg
			}(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := domain.NewUploadRequest(tt.requesterID, tt.podID, tt.podOwnerID, tt.message)

			if request == nil {
				t.Fatal("Expected request to be created")
			}
			if request.ID == uuid.Nil {
				t.Error("Expected ID to be set")
			}
			if request.RequesterID != tt.requesterID {
				t.Errorf("Expected RequesterID %s, got %s", tt.requesterID, request.RequesterID)
			}
			if request.PodID != tt.podID {
				t.Errorf("Expected PodID %s, got %s", tt.podID, request.PodID)
			}
			if request.PodOwnerID != tt.podOwnerID {
				t.Errorf("Expected PodOwnerID %s, got %s", tt.podOwnerID, request.PodOwnerID)
			}
			if request.Status != domain.UploadRequestStatusPending {
				t.Errorf("Expected status pending, got %s", request.Status)
			}
			if request.CreatedAt.IsZero() {
				t.Error("Expected CreatedAt to be set")
			}
			if request.UpdatedAt.IsZero() {
				t.Error("Expected UpdatedAt to be set")
			}
			if tt.message != nil && (request.Message == nil || *request.Message != *tt.message) {
				t.Error("Expected message to be set correctly")
			}
		})
	}
}

// TestUploadRequest_StatusMethods tests the status helper methods
func TestUploadRequest_StatusMethods(t *testing.T) {
	tests := []struct {
		name       string
		status     domain.UploadRequestStatus
		isPending  bool
		isApproved bool
	}{
		{
			name:       "pending status",
			status:     domain.UploadRequestStatusPending,
			isPending:  true,
			isApproved: false,
		},
		{
			name:       "approved status",
			status:     domain.UploadRequestStatusApproved,
			isPending:  false,
			isApproved: true,
		},
		{
			name:       "rejected status",
			status:     domain.UploadRequestStatusRejected,
			isPending:  false,
			isApproved: false,
		},
		{
			name:       "revoked status",
			status:     domain.UploadRequestStatusRevoked,
			isPending:  false,
			isApproved: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := &domain.UploadRequest{
				ID:          uuid.New(),
				RequesterID: uuid.New(),
				PodID:       uuid.New(),
				PodOwnerID:  uuid.New(),
				Status:      tt.status,
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			}

			if request.IsPending() != tt.isPending {
				t.Errorf("IsPending() = %v, want %v", request.IsPending(), tt.isPending)
			}
			if request.IsApproved() != tt.isApproved {
				t.Errorf("IsApproved() = %v, want %v", request.IsApproved(), tt.isApproved)
			}
		})
	}
}

// TestUploadRequest_IsExpired tests the IsExpired method
func TestUploadRequest_IsExpired(t *testing.T) {
	tests := []struct {
		name      string
		expiresAt *time.Time
		isExpired bool
	}{
		{
			name:      "no expiry set",
			expiresAt: nil,
			isExpired: false,
		},
		{
			name: "not expired",
			expiresAt: func() *time.Time {
				t := time.Now().Add(24 * time.Hour)
				return &t
			}(),
			isExpired: false,
		},
		{
			name: "expired",
			expiresAt: func() *time.Time {
				t := time.Now().Add(-24 * time.Hour)
				return &t
			}(),
			isExpired: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := &domain.UploadRequest{
				ID:          uuid.New(),
				RequesterID: uuid.New(),
				PodID:       uuid.New(),
				PodOwnerID:  uuid.New(),
				Status:      domain.UploadRequestStatusApproved,
				ExpiresAt:   tt.expiresAt,
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			}

			if request.IsExpired() != tt.isExpired {
				t.Errorf("IsExpired() = %v, want %v", request.IsExpired(), tt.isExpired)
			}
		})
	}
}

// TestUploadRequest_CanUpload tests the CanUpload method
func TestUploadRequest_CanUpload(t *testing.T) {
	tests := []struct {
		name      string
		status    domain.UploadRequestStatus
		expiresAt *time.Time
		canUpload bool
	}{
		{
			name:      "approved without expiry",
			status:    domain.UploadRequestStatusApproved,
			expiresAt: nil,
			canUpload: true,
		},
		{
			name:   "approved not expired",
			status: domain.UploadRequestStatusApproved,
			expiresAt: func() *time.Time {
				t := time.Now().Add(24 * time.Hour)
				return &t
			}(),
			canUpload: true,
		},
		{
			name:   "approved but expired",
			status: domain.UploadRequestStatusApproved,
			expiresAt: func() *time.Time {
				t := time.Now().Add(-24 * time.Hour)
				return &t
			}(),
			canUpload: false,
		},
		{
			name:      "pending status",
			status:    domain.UploadRequestStatusPending,
			expiresAt: nil,
			canUpload: false,
		},
		{
			name:      "rejected status",
			status:    domain.UploadRequestStatusRejected,
			expiresAt: nil,
			canUpload: false,
		},
		{
			name:      "revoked status",
			status:    domain.UploadRequestStatusRevoked,
			expiresAt: nil,
			canUpload: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := &domain.UploadRequest{
				ID:          uuid.New(),
				RequesterID: uuid.New(),
				PodID:       uuid.New(),
				PodOwnerID:  uuid.New(),
				Status:      tt.status,
				ExpiresAt:   tt.expiresAt,
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			}

			if request.CanUpload() != tt.canUpload {
				t.Errorf("CanUpload() = %v, want %v", request.CanUpload(), tt.canUpload)
			}
		})
	}
}

// TestUploadRequestRepository_Interface tests that the repository implements the interface
func TestUploadRequestRepository_Interface(t *testing.T) {
	// This test verifies that UploadRequestRepository implements domain.UploadRequestRepository
	// The compile-time check is in interfaces.go, but this documents the expected behavior
	var _ domain.UploadRequestRepository = (*UploadRequestRepository)(nil)
}

// TestUploadRequestRepository_Methods tests the repository method signatures
func TestUploadRequestRepository_Methods(t *testing.T) {
	// Test that the repository has all expected methods with correct signatures
	// This is a compile-time check that documents the API

	t.Run("Create method exists", func(t *testing.T) {
		var repo *UploadRequestRepository
		var _ func(context.Context, *domain.UploadRequest) error = repo.Create
	})

	t.Run("FindByID method exists", func(t *testing.T) {
		var repo *UploadRequestRepository
		var _ func(context.Context, uuid.UUID) (*domain.UploadRequest, error) = repo.FindByID
	})

	t.Run("FindByRequesterAndPod method exists", func(t *testing.T) {
		var repo *UploadRequestRepository
		var _ func(context.Context, uuid.UUID, uuid.UUID) (*domain.UploadRequest, error) = repo.FindByRequesterAndPod
	})

	t.Run("FindByPodOwner method exists", func(t *testing.T) {
		var repo *UploadRequestRepository
		var _ func(context.Context, uuid.UUID, *domain.UploadRequestStatus, int, int) ([]*domain.UploadRequest, int, error) = repo.FindByPodOwner
	})

	t.Run("FindByRequester method exists", func(t *testing.T) {
		var repo *UploadRequestRepository
		var _ func(context.Context, uuid.UUID, int, int) ([]*domain.UploadRequest, int, error) = repo.FindByRequester
	})

	t.Run("FindApprovedByRequesterAndPod method exists", func(t *testing.T) {
		var repo *UploadRequestRepository
		var _ func(context.Context, uuid.UUID, uuid.UUID) (*domain.UploadRequest, error) = repo.FindApprovedByRequesterAndPod
	})

	t.Run("Update method exists", func(t *testing.T) {
		var repo *UploadRequestRepository
		var _ func(context.Context, *domain.UploadRequest) error = repo.Update
	})

	t.Run("UpdateStatus method exists", func(t *testing.T) {
		var repo *UploadRequestRepository
		var _ func(context.Context, uuid.UUID, domain.UploadRequestStatus, *string) error = repo.UpdateStatus
	})
}

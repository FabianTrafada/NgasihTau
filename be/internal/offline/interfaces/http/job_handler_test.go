package http

import (
	"testing"
	"time"

	"github.com/google/uuid"

	"ngasihtau/internal/offline/domain"
)

func TestToJobResponse(t *testing.T) {
	// Test nil input
	result := ToJobResponse(nil)
	if result != nil {
		t.Error("ToJobResponse(nil) should return nil")
	}

	// Test valid job
	now := time.Now()
	startedAt := now.Add(-5 * time.Minute)
	errorMsg := "test error"
	job := &domain.EncryptionJob{
		ID:          uuid.New(),
		MaterialID:  uuid.New(),
		UserID:      uuid.New(),
		DeviceID:    uuid.New(),
		LicenseID:   uuid.New(),
		Priority:    domain.JobPriorityNormal,
		Status:      domain.JobStatusProcessing,
		Error:       &errorMsg,
		RetryCount:  1,
		CreatedAt:   now.Add(-10 * time.Minute),
		StartedAt:   &startedAt,
		CompletedAt: nil,
	}

	resp := ToJobResponse(job)
	if resp == nil {
		t.Fatal("ToJobResponse should not return nil for valid job")
	}

	if resp.ID != job.ID {
		t.Errorf("ID mismatch: got %v, want %v", resp.ID, job.ID)
	}
	if resp.MaterialID != job.MaterialID {
		t.Errorf("MaterialID mismatch: got %v, want %v", resp.MaterialID, job.MaterialID)
	}
	if resp.Status != string(job.Status) {
		t.Errorf("Status mismatch: got %v, want %v", resp.Status, job.Status)
	}
	if resp.Priority != job.Priority {
		t.Errorf("Priority mismatch: got %v, want %v", resp.Priority, job.Priority)
	}
	if resp.RetryCount != job.RetryCount {
		t.Errorf("RetryCount mismatch: got %v, want %v", resp.RetryCount, job.RetryCount)
	}
	if resp.Error == nil || *resp.Error != errorMsg {
		t.Errorf("Error mismatch: got %v, want %v", resp.Error, &errorMsg)
	}
	if resp.StartedAt == nil {
		t.Error("StartedAt should not be nil")
	}
	if resp.CompletedAt != nil {
		t.Error("CompletedAt should be nil")
	}
}

func TestToJobResponse_AllStatuses(t *testing.T) {
	statuses := []domain.JobStatus{
		domain.JobStatusPending,
		domain.JobStatusProcessing,
		domain.JobStatusCompleted,
		domain.JobStatusFailed,
	}

	for _, status := range statuses {
		t.Run(string(status), func(t *testing.T) {
			job := &domain.EncryptionJob{
				ID:         uuid.New(),
				MaterialID: uuid.New(),
				UserID:     uuid.New(),
				DeviceID:   uuid.New(),
				LicenseID:  uuid.New(),
				Priority:   domain.JobPriorityNormal,
				Status:     status,
				CreatedAt:  time.Now(),
			}

			resp := ToJobResponse(job)
			if resp == nil {
				t.Fatal("ToJobResponse should not return nil")
			}
			if resp.Status != string(status) {
				t.Errorf("Status mismatch: got %v, want %v", resp.Status, status)
			}
		})
	}
}

func TestToJobResponse_Priorities(t *testing.T) {
	priorities := []int{
		domain.JobPriorityHigh,
		domain.JobPriorityNormal,
		domain.JobPriorityLow,
	}

	for _, priority := range priorities {
		t.Run("priority_"+string(rune('0'+priority)), func(t *testing.T) {
			job := &domain.EncryptionJob{
				ID:         uuid.New(),
				MaterialID: uuid.New(),
				UserID:     uuid.New(),
				DeviceID:   uuid.New(),
				LicenseID:  uuid.New(),
				Priority:   priority,
				Status:     domain.JobStatusPending,
				CreatedAt:  time.Now(),
			}

			resp := ToJobResponse(job)
			if resp == nil {
				t.Fatal("ToJobResponse should not return nil")
			}
			if resp.Priority != priority {
				t.Errorf("Priority mismatch: got %v, want %v", resp.Priority, priority)
			}
		})
	}
}

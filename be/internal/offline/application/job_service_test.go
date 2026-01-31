package application

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"ngasihtau/internal/offline/domain"
)

// mockEncryptionJobRepository is a mock implementation of domain.EncryptionJobRepository.
type mockEncryptionJobRepository struct {
	mu        sync.RWMutex
	jobs      map[uuid.UUID]*domain.EncryptionJob
	createErr error
	findErr   error
	updateErr error
}

func newMockEncryptionJobRepository() *mockEncryptionJobRepository {
	return &mockEncryptionJobRepository{
		jobs: make(map[uuid.UUID]*domain.EncryptionJob),
	}
}

func (m *mockEncryptionJobRepository) Create(ctx context.Context, job *domain.EncryptionJob) error {
	if m.createErr != nil {
		return m.createErr
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.jobs[job.ID] = job
	return nil
}

func (m *mockEncryptionJobRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.EncryptionJob, error) {
	if m.findErr != nil {
		return nil, m.findErr
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	job, ok := m.jobs[id]
	if !ok {
		return nil, domain.NewOfflineError(domain.ErrCodeJobNotFound, "job not found")
	}
	return job, nil
}

func (m *mockEncryptionJobRepository) FindByMaterialID(ctx context.Context, materialID uuid.UUID) ([]*domain.EncryptionJob, error) {
	if m.findErr != nil {
		return nil, m.findErr
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	var result []*domain.EncryptionJob
	for _, job := range m.jobs {
		if job.MaterialID == materialID {
			result = append(result, job)
		}
	}
	return result, nil
}

func (m *mockEncryptionJobRepository) FindByLicenseID(ctx context.Context, licenseID uuid.UUID) (*domain.EncryptionJob, error) {
	if m.findErr != nil {
		return nil, m.findErr
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, job := range m.jobs {
		if job.LicenseID == licenseID {
			return job, nil
		}
	}
	return nil, domain.NewOfflineError(domain.ErrCodeJobNotFound, "job not found")
}

func (m *mockEncryptionJobRepository) FindPending(ctx context.Context, limit int) ([]*domain.EncryptionJob, error) {
	if m.findErr != nil {
		return nil, m.findErr
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	var result []*domain.EncryptionJob
	for _, job := range m.jobs {
		if job.Status == domain.JobStatusPending {
			result = append(result, job)
			if len(result) >= limit {
				break
			}
		}
	}
	return result, nil
}

func (m *mockEncryptionJobRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.JobStatus) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	job, ok := m.jobs[id]
	if !ok {
		return domain.NewOfflineError(domain.ErrCodeJobNotFound, "job not found")
	}
	job.Status = status
	return nil
}

func (m *mockEncryptionJobRepository) UpdateStarted(ctx context.Context, id uuid.UUID) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	job, ok := m.jobs[id]
	if !ok {
		return domain.NewOfflineError(domain.ErrCodeJobNotFound, "job not found")
	}
	job.Status = domain.JobStatusProcessing
	now := time.Now()
	job.StartedAt = &now
	return nil
}

func (m *mockEncryptionJobRepository) UpdateCompleted(ctx context.Context, id uuid.UUID) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	job, ok := m.jobs[id]
	if !ok {
		return domain.NewOfflineError(domain.ErrCodeJobNotFound, "job not found")
	}
	job.Status = domain.JobStatusCompleted
	now := time.Now()
	job.CompletedAt = &now
	return nil
}

func (m *mockEncryptionJobRepository) UpdateFailed(ctx context.Context, id uuid.UUID, errorMsg string) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	job, ok := m.jobs[id]
	if !ok {
		return domain.NewOfflineError(domain.ErrCodeJobNotFound, "job not found")
	}
	job.Status = domain.JobStatusFailed
	job.Error = &errorMsg
	now := time.Now()
	job.CompletedAt = &now
	return nil
}

func (m *mockEncryptionJobRepository) IncrementRetryCount(ctx context.Context, id uuid.UUID) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	job, ok := m.jobs[id]
	if !ok {
		return domain.NewOfflineError(domain.ErrCodeJobNotFound, "job not found")
	}
	job.RetryCount++
	return nil
}

func (m *mockEncryptionJobRepository) DeleteOldCompleted(ctx context.Context, olderThan time.Time) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for id, job := range m.jobs {
		if job.Status == domain.JobStatusCompleted && job.CompletedAt != nil && job.CompletedAt.Before(olderThan) {
			delete(m.jobs, id)
		}
	}
	return nil
}

// mockJobEventPublisher is a mock implementation of OfflineEventPublisher for job tests.
type mockJobEventPublisher struct {
	mu                  sync.Mutex
	encryptionRequested []EncryptionJobEvent
	encryptionCompleted []EncryptionJobEvent
	encryptionFailed    []EncryptionJobEvent
}

func newMockJobEventPublisher() *mockJobEventPublisher {
	return &mockJobEventPublisher{}
}

func (m *mockJobEventPublisher) PublishKeyGenerated(ctx context.Context, event KeyEvent) error {
	return nil
}

func (m *mockJobEventPublisher) PublishKeyRetrieved(ctx context.Context, event KeyEvent) error {
	return nil
}

func (m *mockJobEventPublisher) PublishLicenseIssued(ctx context.Context, event LicenseEvent) error {
	return nil
}

func (m *mockJobEventPublisher) PublishLicenseValidated(ctx context.Context, event LicenseEvent) error {
	return nil
}

func (m *mockJobEventPublisher) PublishLicenseRevoked(ctx context.Context, event LicenseEvent) error {
	return nil
}

func (m *mockJobEventPublisher) PublishLicenseRenewed(ctx context.Context, event LicenseEvent) error {
	return nil
}

func (m *mockJobEventPublisher) PublishDeviceRegistered(ctx context.Context, event DeviceEvent) error {
	return nil
}

func (m *mockJobEventPublisher) PublishDeviceDeregistered(ctx context.Context, event DeviceEvent) error {
	return nil
}

func (m *mockJobEventPublisher) PublishEncryptionRequested(ctx context.Context, event EncryptionJobEvent) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.encryptionRequested = append(m.encryptionRequested, event)
	return nil
}

func (m *mockJobEventPublisher) PublishEncryptionCompleted(ctx context.Context, event EncryptionJobEvent) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.encryptionCompleted = append(m.encryptionCompleted, event)
	return nil
}

func (m *mockJobEventPublisher) PublishEncryptionFailed(ctx context.Context, event EncryptionJobEvent) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.encryptionFailed = append(m.encryptionFailed, event)
	return nil
}

func (m *mockJobEventPublisher) PublishMaterialDownloaded(ctx context.Context, event MaterialDownloadEvent) error {
	return nil
}

func TestJobService_CreateJob(t *testing.T) {
	ctx := context.Background()
	jobRepo := newMockEncryptionJobRepository()
	publisher := newMockJobEventPublisher()

	service := NewJobService(
		jobRepo,
		nil, // encryptionService not needed for create
		nil, // keyMgmtService not needed for create
		publisher,
		DefaultJobServiceConfig(),
	)

	input := CreateJobInput{
		MaterialID: uuid.New(),
		UserID:     uuid.New(),
		DeviceID:   uuid.New(),
		LicenseID:  uuid.New(),
		Priority:   domain.JobPriorityNormal,
	}

	job, err := service.CreateJob(ctx, input)
	require.NoError(t, err)
	assert.NotNil(t, job)
	assert.Equal(t, input.MaterialID, job.MaterialID)
	assert.Equal(t, input.UserID, job.UserID)
	assert.Equal(t, input.DeviceID, job.DeviceID)
	assert.Equal(t, input.LicenseID, job.LicenseID)
	assert.Equal(t, domain.JobPriorityNormal, job.Priority)
	assert.Equal(t, domain.JobStatusPending, job.Status)
	assert.Equal(t, 0, job.RetryCount)

	// Verify event was published
	assert.Len(t, publisher.encryptionRequested, 1)
	assert.Equal(t, job.ID, publisher.encryptionRequested[0].JobID)
}

func TestJobService_CreateJob_InvalidPriority(t *testing.T) {
	ctx := context.Background()
	jobRepo := newMockEncryptionJobRepository()
	publisher := newMockJobEventPublisher()

	service := NewJobService(
		jobRepo,
		nil,
		nil,
		publisher,
		DefaultJobServiceConfig(),
	)

	input := CreateJobInput{
		MaterialID: uuid.New(),
		UserID:     uuid.New(),
		DeviceID:   uuid.New(),
		LicenseID:  uuid.New(),
		Priority:   99, // Invalid priority
	}

	job, err := service.CreateJob(ctx, input)
	require.NoError(t, err)
	assert.Equal(t, domain.JobPriorityNormal, job.Priority) // Should default to normal
}

func TestJobService_CreateJob_DatabaseError(t *testing.T) {
	ctx := context.Background()
	jobRepo := newMockEncryptionJobRepository()
	jobRepo.createErr = errors.New("database error")
	publisher := newMockJobEventPublisher()

	service := NewJobService(
		jobRepo,
		nil,
		nil,
		publisher,
		DefaultJobServiceConfig(),
	)

	input := CreateJobInput{
		MaterialID: uuid.New(),
		UserID:     uuid.New(),
		DeviceID:   uuid.New(),
		LicenseID:  uuid.New(),
		Priority:   domain.JobPriorityNormal,
	}

	job, err := service.CreateJob(ctx, input)
	assert.Error(t, err)
	assert.Nil(t, job)
}

func TestJobService_GetJob(t *testing.T) {
	ctx := context.Background()
	jobRepo := newMockEncryptionJobRepository()

	// Create a job directly in the repo
	existingJob := domain.NewEncryptionJob(uuid.New(), uuid.New(), uuid.New(), uuid.New(), domain.JobPriorityNormal)
	jobRepo.jobs[existingJob.ID] = existingJob

	service := NewJobService(
		jobRepo,
		nil,
		nil,
		nil,
		DefaultJobServiceConfig(),
	)

	job, err := service.GetJob(ctx, existingJob.ID)
	require.NoError(t, err)
	assert.Equal(t, existingJob.ID, job.ID)
}

func TestJobService_GetJob_NotFound(t *testing.T) {
	ctx := context.Background()
	jobRepo := newMockEncryptionJobRepository()

	service := NewJobService(
		jobRepo,
		nil,
		nil,
		nil,
		DefaultJobServiceConfig(),
	)

	job, err := service.GetJob(ctx, uuid.New())
	assert.Error(t, err)
	assert.Nil(t, job)
}

func TestJobService_GetJobByMaterial(t *testing.T) {
	ctx := context.Background()
	jobRepo := newMockEncryptionJobRepository()

	materialID := uuid.New()
	existingJob := domain.NewEncryptionJob(materialID, uuid.New(), uuid.New(), uuid.New(), domain.JobPriorityNormal)
	jobRepo.jobs[existingJob.ID] = existingJob

	service := NewJobService(
		jobRepo,
		nil,
		nil,
		nil,
		DefaultJobServiceConfig(),
	)

	job, err := service.GetJobByMaterial(ctx, materialID)
	require.NoError(t, err)
	assert.Equal(t, materialID, job.MaterialID)
}

func TestJobService_GetJobByLicense(t *testing.T) {
	ctx := context.Background()
	jobRepo := newMockEncryptionJobRepository()

	licenseID := uuid.New()
	existingJob := domain.NewEncryptionJob(uuid.New(), uuid.New(), uuid.New(), licenseID, domain.JobPriorityNormal)
	jobRepo.jobs[existingJob.ID] = existingJob

	service := NewJobService(
		jobRepo,
		nil,
		nil,
		nil,
		DefaultJobServiceConfig(),
	)

	job, err := service.GetJobByLicense(ctx, licenseID)
	require.NoError(t, err)
	assert.Equal(t, licenseID, job.LicenseID)
}

func TestJobService_CalculateRetryDelay(t *testing.T) {
	config := DefaultJobServiceConfig()
	config.RetryBaseDelay = 1 * time.Second
	config.RetryMaxDelay = 1 * time.Minute

	service := NewJobService(nil, nil, nil, nil, config)

	tests := []struct {
		retryCount    int
		expectedDelay time.Duration
	}{
		{0, 1 * time.Second},
		{1, 2 * time.Second},
		{2, 4 * time.Second},
		{3, 8 * time.Second},
		{4, 16 * time.Second},
		{5, 32 * time.Second},
		{6, 1 * time.Minute},  // Capped at max
		{10, 1 * time.Minute}, // Still capped
	}

	for _, tt := range tests {
		delay := service.CalculateRetryDelay(tt.retryCount)
		assert.Equal(t, tt.expectedDelay, delay, "retry count %d", tt.retryCount)
	}
}

func TestJobService_CleanupOldJobs(t *testing.T) {
	ctx := context.Background()
	jobRepo := newMockEncryptionJobRepository()

	// Create old completed job
	oldJob := domain.NewEncryptionJob(uuid.New(), uuid.New(), uuid.New(), uuid.New(), domain.JobPriorityNormal)
	oldJob.Status = domain.JobStatusCompleted
	oldTime := time.Now().Add(-48 * time.Hour)
	oldJob.CompletedAt = &oldTime
	jobRepo.jobs[oldJob.ID] = oldJob

	// Create recent completed job
	recentJob := domain.NewEncryptionJob(uuid.New(), uuid.New(), uuid.New(), uuid.New(), domain.JobPriorityNormal)
	recentJob.Status = domain.JobStatusCompleted
	recentTime := time.Now().Add(-1 * time.Hour)
	recentJob.CompletedAt = &recentTime
	jobRepo.jobs[recentJob.ID] = recentJob

	service := NewJobService(
		jobRepo,
		nil,
		nil,
		nil,
		DefaultJobServiceConfig(),
	)

	err := service.CleanupOldJobs(ctx, 24*time.Hour)
	require.NoError(t, err)

	// Old job should be deleted
	_, exists := jobRepo.jobs[oldJob.ID]
	assert.False(t, exists, "old job should be deleted")

	// Recent job should still exist
	_, exists = jobRepo.jobs[recentJob.ID]
	assert.True(t, exists, "recent job should still exist")
}

func TestEncryptionJob_StatusMethods(t *testing.T) {
	job := domain.NewEncryptionJob(uuid.New(), uuid.New(), uuid.New(), uuid.New(), domain.JobPriorityNormal)

	// Initial state
	assert.True(t, job.IsPending())
	assert.False(t, job.IsProcessing())
	assert.False(t, job.IsCompleted())
	assert.False(t, job.IsFailed())
	assert.True(t, job.CanRetry())

	// Processing state
	job.Status = domain.JobStatusProcessing
	assert.False(t, job.IsPending())
	assert.True(t, job.IsProcessing())

	// Completed state
	job.Status = domain.JobStatusCompleted
	assert.True(t, job.IsCompleted())

	// Failed state
	job.Status = domain.JobStatusFailed
	assert.True(t, job.IsFailed())

	// Retry limit
	job.RetryCount = domain.MaxJobRetries
	assert.False(t, job.CanRetry())
}

func TestEncryptionWorker_StartStop(t *testing.T) {
	ctx := context.Background()
	jobRepo := newMockEncryptionJobRepository()

	config := DefaultWorkerConfig()
	config.PollInterval = 100 * time.Millisecond
	config.ShutdownTimeout = 1 * time.Second

	worker := NewEncryptionWorker(nil, jobRepo, config)

	// Start worker
	err := worker.Start(ctx)
	require.NoError(t, err)
	assert.True(t, worker.IsRunning())

	// Try to start again (should fail)
	err = worker.Start(ctx)
	assert.Error(t, err)

	// Stop worker
	err = worker.Stop()
	require.NoError(t, err)
	assert.False(t, worker.IsRunning())

	// Stop again (should be no-op)
	err = worker.Stop()
	require.NoError(t, err)
}

func TestDefaultConfigs(t *testing.T) {
	// Test DefaultJobServiceConfig
	jobConfig := DefaultJobServiceConfig()
	assert.Equal(t, domain.DefaultJobPollInterval, jobConfig.PollInterval)
	assert.Equal(t, domain.DefaultJobTimeout, jobConfig.JobTimeout)
	assert.Equal(t, domain.MaxJobRetries, jobConfig.MaxRetries)

	// Test DefaultWorkerConfig
	workerConfig := DefaultWorkerConfig()
	assert.Equal(t, domain.DefaultWorkerConcurrency, workerConfig.Concurrency)
	assert.Equal(t, domain.DefaultJobPollInterval, workerConfig.PollInterval)
	assert.Equal(t, domain.DefaultShutdownTimeout, workerConfig.ShutdownTimeout)

	// Test DefaultNATSJobQueueConfig
	natsConfig := DefaultNATSJobQueueConfig()
	assert.Equal(t, domain.JetStreamEncryptionStream, natsConfig.StreamName)
	assert.Equal(t, domain.JetStreamEncryptionConsumer, natsConfig.ConsumerName)
	assert.Equal(t, domain.NATSSubjectEncryptionRequested, natsConfig.Subject)
}

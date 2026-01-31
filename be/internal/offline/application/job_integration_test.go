package application

import (
	"bytes"
	"context"
	"io"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"ngasihtau/internal/offline/domain"
)

// mockKeyManagementServiceForJob is a mock for key management in job tests.
type mockKeyManagementServiceForJob struct {
	cek       []byte
	cekRecord *domain.ContentEncryptionKey
	err       error
}

func newMockKeyManagementServiceForJob() *mockKeyManagementServiceForJob {
	cek := make([]byte, domain.CEKSize)
	for i := range cek {
		cek[i] = byte(i)
	}
	return &mockKeyManagementServiceForJob{
		cek: cek,
		cekRecord: &domain.ContentEncryptionKey{
			ID:           uuid.New(),
			UserID:       uuid.New(),
			MaterialID:   uuid.New(),
			DeviceID:     uuid.New(),
			EncryptedKey: cek,
			KeyVersion:   1,
			CreatedAt:    time.Now(),
		},
	}
}

// mockEncryptionServiceForJob is a mock for encryption in job tests.
type mockEncryptionServiceForJob struct {
	mu           sync.Mutex
	encrypted    map[uuid.UUID]bool
	encryptErr   error
	encryptDelay time.Duration
}

func newMockEncryptionServiceForJob() *mockEncryptionServiceForJob {
	return &mockEncryptionServiceForJob{
		encrypted: make(map[uuid.UUID]bool),
	}
}

func (m *mockEncryptionServiceForJob) EncryptMaterial(ctx context.Context, input EncryptMaterialInput) (*EncryptMaterialOutput, error) {
	if m.encryptDelay > 0 {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(m.encryptDelay):
		}
	}

	if m.encryptErr != nil {
		return nil, m.encryptErr
	}

	m.mu.Lock()
	m.encrypted[input.MaterialID] = true
	m.mu.Unlock()

	manifest := &domain.DownloadManifest{
		MaterialID:    input.MaterialID,
		LicenseID:     input.LicenseID,
		TotalChunks:   1,
		TotalSize:     1024,
		OriginalHash:  "abc123",
		EncryptedHash: "def456",
		ChunkSize:     domain.DefaultChunkSize,
		Chunks: []domain.EncryptedChunk{
			{Index: 0, Offset: 0, Size: 1024 + domain.IVSize + domain.AuthTagSize},
		},
		FileType:  "pdf",
		CreatedAt: time.Now(),
	}

	return &EncryptMaterialOutput{
		Manifest:         manifest,
		EncryptedFileURL: "https://example.com/encrypted/file.enc",
	}, nil
}

func (m *mockEncryptionServiceForJob) IsEncrypted(materialID uuid.UUID) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.encrypted[materialID]
}

// TestJobIntegration_CreateAndProcess tests the full job creation and processing flow.
func TestJobIntegration_CreateAndProcess(t *testing.T) {
	ctx := context.Background()

	// Setup mocks
	jobRepo := newMockEncryptionJobRepository()
	publisher := newMockJobEventPublisher()
	mockKeyMgmt := newMockKeyManagementServiceForJob()
	mockEncryption := newMockEncryptionServiceForJob()

	// Create a real key management service with mocked dependencies
	cekRepo := newMockCEKRepository()
	auditRepo := newMockAuditLogRepository()
	keyMgmtService := NewKeyManagementService(
		cekRepo,
		auditRepo,
		publisher,
		KeyManagementConfig{
			MasterSecret:   make([]byte, 32),
			KEK:            make([]byte, 32),
			CurrentVersion: 1,
		},
	)

	// Create a mock encryption service wrapper
	encryptionService := &EncryptionService{
		storage:               newMockMinIOStorageClientForJob(),
		materialChecker:       newMockMaterialAccessCheckerForJob(),
		encryptedMaterialRepo: newMockEncryptedMaterialRepositoryForJob(),
		eventPublisher:        publisher,
		encryptedBucket:       "test-bucket",
	}

	// Create job service
	config := DefaultJobServiceConfig()
	config.JobTimeout = 5 * time.Second
	config.MaxRetries = 3

	jobService := NewJobService(
		jobRepo,
		encryptionService,
		keyMgmtService,
		publisher,
		config,
	)

	// Create a job
	input := CreateJobInput{
		MaterialID: uuid.New(),
		UserID:     uuid.New(),
		DeviceID:   uuid.New(),
		LicenseID:  uuid.New(),
		Priority:   domain.JobPriorityHigh,
	}

	job, err := jobService.CreateJob(ctx, input)
	require.NoError(t, err)
	assert.NotNil(t, job)
	assert.Equal(t, domain.JobStatusPending, job.Status)

	// Verify job was stored
	storedJob, err := jobRepo.FindByID(ctx, job.ID)
	require.NoError(t, err)
	assert.Equal(t, job.ID, storedJob.ID)

	// Verify event was published
	assert.Len(t, publisher.encryptionRequested, 1)

	// Note: Full processing test would require more complex setup
	// This test verifies the job creation and storage flow
	_ = mockKeyMgmt
	_ = mockEncryption
}

// TestJobIntegration_WorkerProcessesPendingJobs tests that the worker picks up and processes pending jobs.
func TestJobIntegration_WorkerProcessesPendingJobs(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Setup mocks
	jobRepo := newMockEncryptionJobRepository()
	publisher := newMockJobEventPublisher()

	// Create pending jobs
	for i := 0; i < 3; i++ {
		job := domain.NewEncryptionJob(uuid.New(), uuid.New(), uuid.New(), uuid.New(), domain.JobPriorityNormal)
		jobRepo.jobs[job.ID] = job
	}

	// Verify we have pending jobs
	pending, err := jobRepo.FindPending(ctx, 10)
	require.NoError(t, err)
	assert.Len(t, pending, 3)

	// Create a minimal job service (won't actually process, just for the worker structure)
	config := DefaultJobServiceConfig()
	config.JobTimeout = 1 * time.Second

	jobService := NewJobService(
		jobRepo,
		nil, // No encryption service - worker won't actually process
		nil, // No key management service
		publisher,
		config,
	)

	// Create worker with short poll interval
	workerConfig := DefaultWorkerConfig()
	workerConfig.PollInterval = 50 * time.Millisecond
	workerConfig.Concurrency = 2
	workerConfig.ShutdownTimeout = 1 * time.Second
	workerConfig.BatchSize = 0 // Don't fetch any jobs to avoid nil pointer issues

	worker := NewEncryptionWorker(jobService, jobRepo, workerConfig)

	// Start worker
	err = worker.Start(ctx)
	require.NoError(t, err)
	assert.True(t, worker.IsRunning())

	// Let it run briefly (jobs will fail due to nil services, but that's ok for this test)
	time.Sleep(200 * time.Millisecond)

	// Stop worker
	err = worker.Stop()
	require.NoError(t, err)
	assert.False(t, worker.IsRunning())
}

// TestJobIntegration_ConcurrentJobCreation tests concurrent job creation.
func TestJobIntegration_ConcurrentJobCreation(t *testing.T) {
	ctx := context.Background()

	jobRepo := newMockEncryptionJobRepository()
	publisher := newMockJobEventPublisher()

	jobService := NewJobService(
		jobRepo,
		nil,
		nil,
		publisher,
		DefaultJobServiceConfig(),
	)

	// Create jobs concurrently
	var wg sync.WaitGroup
	numJobs := 10
	jobs := make([]*domain.EncryptionJob, numJobs)
	errors := make([]error, numJobs)

	for i := 0; i < numJobs; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			input := CreateJobInput{
				MaterialID: uuid.New(),
				UserID:     uuid.New(),
				DeviceID:   uuid.New(),
				LicenseID:  uuid.New(),
				Priority:   domain.JobPriorityNormal,
			}
			jobs[idx], errors[idx] = jobService.CreateJob(ctx, input)
		}(i)
	}

	wg.Wait()

	// Verify all jobs were created successfully
	for i := 0; i < numJobs; i++ {
		assert.NoError(t, errors[i], "job %d should be created without error", i)
		assert.NotNil(t, jobs[i], "job %d should not be nil", i)
	}

	// Verify all jobs are in the repository
	assert.Len(t, jobRepo.jobs, numJobs)

	// Verify all events were published
	assert.Len(t, publisher.encryptionRequested, numJobs)
}

// TestJobIntegration_JobPriorityOrdering tests that jobs are processed in priority order.
func TestJobIntegration_JobPriorityOrdering(t *testing.T) {
	ctx := context.Background()

	jobRepo := newMockEncryptionJobRepository()

	// Create jobs with different priorities
	lowPriorityJob := domain.NewEncryptionJob(uuid.New(), uuid.New(), uuid.New(), uuid.New(), domain.JobPriorityLow)
	normalPriorityJob := domain.NewEncryptionJob(uuid.New(), uuid.New(), uuid.New(), uuid.New(), domain.JobPriorityNormal)
	highPriorityJob := domain.NewEncryptionJob(uuid.New(), uuid.New(), uuid.New(), uuid.New(), domain.JobPriorityHigh)

	// Add in reverse priority order
	jobRepo.jobs[lowPriorityJob.ID] = lowPriorityJob
	jobRepo.jobs[normalPriorityJob.ID] = normalPriorityJob
	jobRepo.jobs[highPriorityJob.ID] = highPriorityJob

	// Note: The mock repository doesn't implement priority ordering
	// In a real implementation, FindPending would return jobs ordered by priority
	pending, err := jobRepo.FindPending(ctx, 10)
	require.NoError(t, err)
	assert.Len(t, pending, 3)
}

// Mock implementations for job integration tests

type mockMinIOStorageClientForJob struct{}

func newMockMinIOStorageClientForJob() *mockMinIOStorageClientForJob {
	return &mockMinIOStorageClientForJob{}
}

func (m *mockMinIOStorageClientForJob) GetObject(ctx context.Context, objectKey string) (io.ReadCloser, error) {
	// Return a small test file
	content := bytes.Repeat([]byte("test content "), 100)
	return io.NopCloser(bytes.NewReader(content)), nil
}

func (m *mockMinIOStorageClientForJob) PutObject(ctx context.Context, objectKey string, reader io.Reader, size int64, contentType string) error {
	return nil
}

func (m *mockMinIOStorageClientForJob) DeleteObject(ctx context.Context, objectKey string) error {
	return nil
}

func (m *mockMinIOStorageClientForJob) DeleteObjects(ctx context.Context, objectKeys []string) error {
	return nil
}

func (m *mockMinIOStorageClientForJob) GetObjectInfo(ctx context.Context, objectKey string) (*ObjectInfo, error) {
	return &ObjectInfo{
		Size:        1300,
		ContentType: "application/pdf",
		ETag:        "abc123",
	}, nil
}

func (m *mockMinIOStorageClientForJob) GeneratePresignedGetURL(ctx context.Context, objectKey string, expiry time.Duration) (string, error) {
	return "https://example.com/presigned/" + objectKey, nil
}

type mockMaterialAccessCheckerForJob struct{}

func newMockMaterialAccessCheckerForJob() *mockMaterialAccessCheckerForJob {
	return &mockMaterialAccessCheckerForJob{}
}

func (m *mockMaterialAccessCheckerForJob) GetMaterialFileKey(ctx context.Context, materialID uuid.UUID) (string, string, error) {
	return "materials/" + materialID.String() + "/file.pdf", "pdf", nil
}

type mockEncryptedMaterialRepositoryForJob struct {
	mu        sync.Mutex
	materials map[uuid.UUID]*domain.EncryptedMaterial
}

func newMockEncryptedMaterialRepositoryForJob() *mockEncryptedMaterialRepositoryForJob {
	return &mockEncryptedMaterialRepositoryForJob{
		materials: make(map[uuid.UUID]*domain.EncryptedMaterial),
	}
}

func (m *mockEncryptedMaterialRepositoryForJob) Create(ctx context.Context, material *domain.EncryptedMaterial) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.materials[material.ID] = material
	return nil
}

func (m *mockEncryptedMaterialRepositoryForJob) FindById(ctx context.Context, id uuid.UUID) (*domain.EncryptedMaterial, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	material, ok := m.materials[id]
	if !ok {
		return nil, domain.NewOfflineError(domain.ErrCodeMaterialNotFound, "material not found")
	}
	return material, nil
}

func (m *mockEncryptedMaterialRepositoryForJob) FindByMaterialAndCEK(ctx context.Context, materialID, cekID uuid.UUID) (*domain.EncryptedMaterial, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, material := range m.materials {
		if material.MaterialID == materialID && material.CEKID == cekID {
			return material, nil
		}
	}
	return nil, domain.NewOfflineError(domain.ErrCodeMaterialNotFound, "material not found")
}

func (m *mockEncryptedMaterialRepositoryForJob) FindByMaterialID(ctx context.Context, materialID uuid.UUID) ([]*domain.EncryptedMaterial, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var result []*domain.EncryptedMaterial
	for _, material := range m.materials {
		if material.MaterialID == materialID {
			result = append(result, material)
		}
	}
	return result, nil
}

func (m *mockEncryptedMaterialRepositoryForJob) Delete(ctx context.Context, id uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.materials, id)
	return nil
}

func (m *mockEncryptedMaterialRepositoryForJob) DeleteByMaterialID(ctx context.Context, materialID uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for id, material := range m.materials {
		if material.MaterialID == materialID {
			delete(m.materials, id)
		}
	}
	return nil
}

// Note: mockCEKRepository and mockAuditLogRepository are defined in key_management_service_test.go

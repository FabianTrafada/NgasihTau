// Package application contains the business logic and use cases for the Offline Material Service.
package application

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"ngasihtau/internal/offline/domain"
)

// JobService handles encryption job management and processing.
// Implements Requirement 5: Background Job Processing.
type JobService struct {
	jobRepo           domain.EncryptionJobRepository
	encryptionService *EncryptionService
	keyMgmtService    *KeyManagementService
	eventPublisher    OfflineEventPublisher
	config            JobServiceConfig
}

// JobServiceConfig holds configuration for the Job Service.
type JobServiceConfig struct {
	// PollInterval is the interval between job queue polls.
	PollInterval time.Duration
	// JobTimeout is the maximum time a job can run before being considered timed out.
	JobTimeout time.Duration
	// MaxRetries is the maximum number of retry attempts for failed jobs.
	MaxRetries int
	// RetryBaseDelay is the base delay for exponential backoff.
	RetryBaseDelay time.Duration
	// RetryMaxDelay is the maximum delay for exponential backoff.
	RetryMaxDelay time.Duration
}

// DefaultJobServiceConfig returns the default job service configuration.
func DefaultJobServiceConfig() JobServiceConfig {
	return JobServiceConfig{
		PollInterval:   domain.DefaultJobPollInterval,
		JobTimeout:     domain.DefaultJobTimeout,
		MaxRetries:     domain.MaxJobRetries,
		RetryBaseDelay: 1 * time.Second,
		RetryMaxDelay:  5 * time.Minute,
	}
}

// NewJobService creates a new Job Service.
func NewJobService(
	jobRepo domain.EncryptionJobRepository,
	encryptionService *EncryptionService,
	keyMgmtService *KeyManagementService,
	eventPublisher OfflineEventPublisher,
	config JobServiceConfig,
) *JobService {
	return &JobService{
		jobRepo:           jobRepo,
		encryptionService: encryptionService,
		keyMgmtService:    keyMgmtService,
		eventPublisher:    eventPublisher,
		config:            config,
	}
}

// CreateJobInput contains input for creating an encryption job.
type CreateJobInput struct {
	MaterialID uuid.UUID
	UserID     uuid.UUID
	DeviceID   uuid.UUID
	LicenseID  uuid.UUID
	Priority   int // 1=high, 2=normal, 3=low
}

// CreateJob creates a new encryption job.
// Implements Requirement 5.1: Create encryption job entity.
func (s *JobService) CreateJob(ctx context.Context, input CreateJobInput) (*domain.EncryptionJob, error) {
	// Validate priority
	if input.Priority < domain.JobPriorityHigh || input.Priority > domain.JobPriorityLow {
		input.Priority = domain.JobPriorityNormal
	}

	// Create job entity
	job := domain.NewEncryptionJob(
		input.MaterialID,
		input.UserID,
		input.DeviceID,
		input.LicenseID,
		input.Priority,
	)

	// Store in database
	if err := s.jobRepo.Create(ctx, job); err != nil {
		return nil, domain.WrapOfflineError(domain.ErrCodeDatabaseError, "failed to create encryption job", err)
	}

	// Publish event
	s.publishJobEvent(ctx, domain.NATSSubjectEncryptionRequested, job, "")

	log.Info().
		Str("job_id", job.ID.String()).
		Str("material_id", input.MaterialID.String()).
		Int("priority", input.Priority).
		Msg("encryption job created")

	return job, nil
}

// GetJob retrieves a job by ID.
func (s *JobService) GetJob(ctx context.Context, jobID uuid.UUID) (*domain.EncryptionJob, error) {
	job, err := s.jobRepo.FindByID(ctx, jobID)
	if err != nil {
		return nil, err
	}
	return job, nil
}

// GetJobByMaterial retrieves the latest job for a material.
func (s *JobService) GetJobByMaterial(ctx context.Context, materialID uuid.UUID) (*domain.EncryptionJob, error) {
	jobs, err := s.jobRepo.FindByMaterialID(ctx, materialID)
	if err != nil {
		return nil, err
	}
	if len(jobs) == 0 {
		return nil, domain.NewOfflineError(domain.ErrCodeJobNotFound, "no encryption job found for material")
	}
	return jobs[0], nil
}

// GetJobByLicense retrieves the job for a license.
func (s *JobService) GetJobByLicense(ctx context.Context, licenseID uuid.UUID) (*domain.EncryptionJob, error) {
	return s.jobRepo.FindByLicenseID(ctx, licenseID)
}

// ProcessJob processes a single encryption job.
// Implements Requirement 5.3, 5.4: Job processing with status tracking.
func (s *JobService) ProcessJob(ctx context.Context, job *domain.EncryptionJob) error {
	// Mark job as started
	if err := s.jobRepo.UpdateStarted(ctx, job.ID); err != nil {
		return fmt.Errorf("failed to mark job as started: %w", err)
	}

	// Create a context with timeout
	jobCtx, cancel := context.WithTimeout(ctx, s.config.JobTimeout)
	defer cancel()

	// Get CEK for encryption
	cekRecord, err := s.keyMgmtService.GetOrCreateCEK(jobCtx, job.UserID, job.MaterialID, job.DeviceID)
	if err != nil {
		return s.handleJobFailure(ctx, job, fmt.Errorf("failed to get CEK: %w", err))
	}

	// Decrypt CEK for use
	cek, err := s.keyMgmtService.DecryptCEK(jobCtx, cekRecord)
	if err != nil {
		return s.handleJobFailure(ctx, job, fmt.Errorf("failed to decrypt CEK: %w", err))
	}

	// Perform encryption
	input := EncryptMaterialInput{
		MaterialID: job.MaterialID,
		LicenseID:  job.LicenseID,
		CEKID:      cekRecord.ID,
		CEK:        cek,
		UserID:     job.UserID,
		DeviceID:   job.DeviceID,
	}

	_, err = s.encryptionService.EncryptMaterial(jobCtx, input)
	if err != nil {
		return s.handleJobFailure(ctx, job, fmt.Errorf("encryption failed: %w", err))
	}

	// Mark job as completed
	if err := s.jobRepo.UpdateCompleted(ctx, job.ID); err != nil {
		log.Error().Err(err).Str("job_id", job.ID.String()).Msg("failed to mark job as completed")
	}

	// Publish completion event
	s.publishJobEvent(ctx, domain.NATSSubjectEncryptionCompleted, job, "")

	log.Info().
		Str("job_id", job.ID.String()).
		Str("material_id", job.MaterialID.String()).
		Msg("encryption job completed successfully")

	return nil
}

// handleJobFailure handles a job failure with retry logic.
// Implements Requirement 5.5: Job retry logic with exponential backoff.
func (s *JobService) handleJobFailure(ctx context.Context, job *domain.EncryptionJob, err error) error {
	errorMsg := err.Error()

	// Increment retry count
	if retryErr := s.jobRepo.IncrementRetryCount(ctx, job.ID); retryErr != nil {
		log.Error().Err(retryErr).Str("job_id", job.ID.String()).Msg("failed to increment retry count")
	}

	// Check if we should retry
	if job.RetryCount+1 < s.config.MaxRetries {
		// Reset to pending for retry
		if statusErr := s.jobRepo.UpdateStatus(ctx, job.ID, domain.JobStatusPending); statusErr != nil {
			log.Error().Err(statusErr).Str("job_id", job.ID.String()).Msg("failed to reset job status for retry")
		}

		log.Warn().
			Err(err).
			Str("job_id", job.ID.String()).
			Int("retry_count", job.RetryCount+1).
			Msg("encryption job failed, will retry")

		return err
	}

	// Max retries exceeded, mark as failed
	if failErr := s.jobRepo.UpdateFailed(ctx, job.ID, errorMsg); failErr != nil {
		log.Error().Err(failErr).Str("job_id", job.ID.String()).Msg("failed to mark job as failed")
	}

	// Publish failure event
	s.publishJobEvent(ctx, domain.NATSSubjectEncryptionFailed, job, errorMsg)

	log.Error().
		Err(err).
		Str("job_id", job.ID.String()).
		Int("retry_count", job.RetryCount+1).
		Msg("encryption job failed permanently")

	return err
}

// CalculateRetryDelay calculates the delay before retrying a job using exponential backoff.
// Implements Requirement 5.5: Exponential backoff.
func (s *JobService) CalculateRetryDelay(retryCount int) time.Duration {
	delay := s.config.RetryBaseDelay * time.Duration(1<<uint(retryCount))
	if delay > s.config.RetryMaxDelay {
		delay = s.config.RetryMaxDelay
	}
	return delay
}

// CleanupOldJobs removes completed jobs older than the specified duration.
func (s *JobService) CleanupOldJobs(ctx context.Context, olderThan time.Duration) error {
	cutoff := time.Now().Add(-olderThan)
	return s.jobRepo.DeleteOldCompleted(ctx, cutoff)
}

// publishJobEvent publishes a job-related event.
func (s *JobService) publishJobEvent(ctx context.Context, subject string, job *domain.EncryptionJob, errorMsg string) {
	if s.eventPublisher == nil {
		return
	}

	event := EncryptionJobEvent{
		JobID:      job.ID,
		MaterialID: job.MaterialID,
		UserID:     job.UserID,
		DeviceID:   job.DeviceID,
		LicenseID:  job.LicenseID,
		Error:      errorMsg,
	}

	var err error
	switch subject {
	case domain.NATSSubjectEncryptionRequested:
		err = s.eventPublisher.PublishEncryptionRequested(ctx, event)
	case domain.NATSSubjectEncryptionCompleted:
		err = s.eventPublisher.PublishEncryptionCompleted(ctx, event)
	case domain.NATSSubjectEncryptionFailed:
		err = s.eventPublisher.PublishEncryptionFailed(ctx, event)
	}

	if err != nil {
		log.Error().Err(err).Str("subject", subject).Msg("failed to publish job event")
	}
}

// EncryptionWorker processes encryption jobs from the queue.
// Implements Requirement 5.3: Encryption worker with concurrency control.
type EncryptionWorker struct {
	jobService *JobService
	jobRepo    domain.EncryptionJobRepository
	config     WorkerConfig
	running    atomic.Bool
	wg         sync.WaitGroup
	stopCh     chan struct{}
	jobCh      chan *domain.EncryptionJob
}

// WorkerConfig holds configuration for the encryption worker.
type WorkerConfig struct {
	// Concurrency is the number of concurrent workers.
	Concurrency int
	// PollInterval is the interval between job queue polls.
	PollInterval time.Duration
	// BatchSize is the number of jobs to fetch per poll.
	BatchSize int
	// ShutdownTimeout is the timeout for graceful shutdown.
	ShutdownTimeout time.Duration
}

// DefaultWorkerConfig returns the default worker configuration.
func DefaultWorkerConfig() WorkerConfig {
	return WorkerConfig{
		Concurrency:     domain.DefaultWorkerConcurrency,
		PollInterval:    domain.DefaultJobPollInterval,
		BatchSize:       10,
		ShutdownTimeout: domain.DefaultShutdownTimeout,
	}
}

// NewEncryptionWorker creates a new encryption worker.
func NewEncryptionWorker(
	jobService *JobService,
	jobRepo domain.EncryptionJobRepository,
	config WorkerConfig,
) *EncryptionWorker {
	return &EncryptionWorker{
		jobService: jobService,
		jobRepo:    jobRepo,
		config:     config,
		stopCh:     make(chan struct{}),
		jobCh:      make(chan *domain.EncryptionJob, config.BatchSize),
	}
}

// Start starts the encryption worker.
// Implements Requirement 5.2, 5.3: Job queue processing with concurrency.
func (w *EncryptionWorker) Start(ctx context.Context) error {
	if w.running.Load() {
		return fmt.Errorf("worker is already running")
	}

	w.running.Store(true)
	w.stopCh = make(chan struct{})

	log.Info().
		Int("concurrency", w.config.Concurrency).
		Dur("poll_interval", w.config.PollInterval).
		Msg("starting encryption worker")

	// Start worker goroutines
	for i := 0; i < w.config.Concurrency; i++ {
		w.wg.Add(1)
		go w.worker(ctx, i)
	}

	// Start job fetcher
	w.wg.Add(1)
	go w.fetchJobs(ctx)

	return nil
}

// Stop stops the encryption worker gracefully.
// Implements Requirement 5.7: Graceful worker shutdown.
func (w *EncryptionWorker) Stop() error {
	if !w.running.Load() {
		return nil
	}

	log.Info().Msg("stopping encryption worker")

	// Signal stop
	close(w.stopCh)

	// Wait for workers with timeout
	done := make(chan struct{})
	go func() {
		w.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		log.Info().Msg("encryption worker stopped gracefully")
	case <-time.After(w.config.ShutdownTimeout):
		log.Warn().Msg("encryption worker shutdown timed out")
	}

	w.running.Store(false)
	return nil
}

// IsRunning returns whether the worker is running.
func (w *EncryptionWorker) IsRunning() bool {
	return w.running.Load()
}

// fetchJobs fetches pending jobs from the database.
func (w *EncryptionWorker) fetchJobs(ctx context.Context) {
	defer w.wg.Done()

	ticker := time.NewTicker(w.config.PollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-w.stopCh:
			return
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.pollAndDispatch(ctx)
		}
	}
}

// pollAndDispatch polls for pending jobs and dispatches them to workers.
func (w *EncryptionWorker) pollAndDispatch(ctx context.Context) {
	if w.config.BatchSize <= 0 {
		return // Skip polling if batch size is 0 or negative
	}

	jobs, err := w.jobRepo.FindPending(ctx, w.config.BatchSize)
	if err != nil {
		log.Error().Err(err).Msg("failed to fetch pending jobs")
		return
	}

	for _, job := range jobs {
		select {
		case <-w.stopCh:
			return
		case <-ctx.Done():
			return
		case w.jobCh <- job:
			// Job dispatched
		default:
			// Channel full, skip this job for now
			log.Debug().Str("job_id", job.ID.String()).Msg("job channel full, skipping")
		}
	}
}

// worker processes jobs from the job channel.
func (w *EncryptionWorker) worker(ctx context.Context, id int) {
	defer w.wg.Done()

	log.Debug().Int("worker_id", id).Msg("worker started")

	for {
		select {
		case <-w.stopCh:
			log.Debug().Int("worker_id", id).Msg("worker stopping")
			return
		case <-ctx.Done():
			log.Debug().Int("worker_id", id).Msg("worker context cancelled")
			return
		case job := <-w.jobCh:
			w.processJob(ctx, id, job)
		}
	}
}

// processJob processes a single job with timeout handling.
// Implements Requirement 5.6: Job timeout handling.
func (w *EncryptionWorker) processJob(ctx context.Context, workerID int, job *domain.EncryptionJob) {
	log.Debug().
		Int("worker_id", workerID).
		Str("job_id", job.ID.String()).
		Msg("processing job")

	// Create timeout context
	jobCtx, cancel := context.WithTimeout(ctx, w.jobService.config.JobTimeout)
	defer cancel()

	// Process the job
	err := w.jobService.ProcessJob(jobCtx, job)
	if err != nil {
		if jobCtx.Err() == context.DeadlineExceeded {
			log.Error().
				Str("job_id", job.ID.String()).
				Dur("timeout", w.jobService.config.JobTimeout).
				Msg("job timed out")
		}
	}
}

// Package application contains the business logic and use cases for the Offline Material Service.
package application

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/rs/zerolog/log"

	"ngasihtau/internal/offline/domain"
)

// JobQueueMessage represents a message in the job queue.
type JobQueueMessage struct {
	JobID      uuid.UUID `json:"job_id"`
	MaterialID uuid.UUID `json:"material_id"`
	UserID     uuid.UUID `json:"user_id"`
	DeviceID   uuid.UUID `json:"device_id"`
	LicenseID  uuid.UUID `json:"license_id"`
	Priority   int       `json:"priority"`
	CreatedAt  time.Time `json:"created_at"`
}

// NATSJobQueue implements a job queue using NATS JetStream.
// Implements Requirement 5.2: NATS JetStream job queue setup.
type NATSJobQueue struct {
	conn     *nats.Conn
	js       jetstream.JetStream
	stream   jetstream.Stream
	consumer jetstream.Consumer
	config   NATSJobQueueConfig
}

// NATSJobQueueConfig holds configuration for the NATS job queue.
type NATSJobQueueConfig struct {
	// StreamName is the name of the JetStream stream.
	StreamName string
	// ConsumerName is the name of the durable consumer.
	ConsumerName string
	// Subject is the subject for job messages.
	Subject string
	// MaxDeliver is the maximum number of delivery attempts.
	MaxDeliver int
	// AckWait is the time to wait for acknowledgment.
	AckWait time.Duration
	// MaxAckPending is the maximum number of pending acknowledgments.
	MaxAckPending int
	// MaxAge is the maximum age of messages in the stream.
	MaxAge time.Duration
}

// DefaultNATSJobQueueConfig returns the default NATS job queue configuration.
func DefaultNATSJobQueueConfig() NATSJobQueueConfig {
	return NATSJobQueueConfig{
		StreamName:    domain.JetStreamEncryptionStream,
		ConsumerName:  domain.JetStreamEncryptionConsumer,
		Subject:       domain.NATSSubjectEncryptionRequested,
		MaxDeliver:    domain.MaxJobRetries + 1,
		AckWait:       domain.DefaultJobTimeout,
		MaxAckPending: 100,
		MaxAge:        7 * 24 * time.Hour, // 7 days
	}
}

// NewNATSJobQueue creates a new NATS job queue.
func NewNATSJobQueue(conn *nats.Conn, config NATSJobQueueConfig) (*NATSJobQueue, error) {
	js, err := jetstream.New(conn)
	if err != nil {
		return nil, fmt.Errorf("failed to create JetStream context: %w", err)
	}

	queue := &NATSJobQueue{
		conn:   conn,
		js:     js,
		config: config,
	}

	return queue, nil
}

// Setup creates or updates the stream and consumer.
func (q *NATSJobQueue) Setup(ctx context.Context) error {
	// Create or update stream
	stream, err := q.js.CreateOrUpdateStream(ctx, jetstream.StreamConfig{
		Name:        q.config.StreamName,
		Description: "Offline material encryption job queue",
		Subjects:    []string{q.config.Subject, domain.NATSSubjectEncryptionCompleted, domain.NATSSubjectEncryptionFailed},
		Retention:   jetstream.WorkQueuePolicy, // Each message delivered to one consumer
		MaxAge:      q.config.MaxAge,
		Storage:     jetstream.FileStorage,
		Replicas:    1,
		Discard:     jetstream.DiscardOld,
	})
	if err != nil {
		return fmt.Errorf("failed to create/update stream: %w", err)
	}
	q.stream = stream

	// Create or update consumer
	consumer, err := q.js.CreateOrUpdateConsumer(ctx, q.config.StreamName, jetstream.ConsumerConfig{
		Durable:        q.config.ConsumerName,
		AckPolicy:      jetstream.AckExplicitPolicy,
		AckWait:        q.config.AckWait,
		MaxDeliver:     q.config.MaxDeliver,
		MaxAckPending:  q.config.MaxAckPending,
		FilterSubjects: []string{q.config.Subject},
	})
	if err != nil {
		return fmt.Errorf("failed to create/update consumer: %w", err)
	}
	q.consumer = consumer

	log.Info().
		Str("stream", q.config.StreamName).
		Str("consumer", q.config.ConsumerName).
		Msg("NATS job queue setup complete")

	return nil
}

// Enqueue adds a job to the queue.
func (q *NATSJobQueue) Enqueue(ctx context.Context, job *domain.EncryptionJob) error {
	msg := JobQueueMessage{
		JobID:      job.ID,
		MaterialID: job.MaterialID,
		UserID:     job.UserID,
		DeviceID:   job.DeviceID,
		LicenseID:  job.LicenseID,
		Priority:   job.Priority,
		CreatedAt:  job.CreatedAt,
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal job message: %w", err)
	}

	_, err = q.js.Publish(ctx, q.config.Subject, data)
	if err != nil {
		return fmt.Errorf("failed to publish job message: %w", err)
	}

	log.Debug().
		Str("job_id", job.ID.String()).
		Msg("job enqueued to NATS")

	return nil
}

// Close closes the job queue connection.
func (q *NATSJobQueue) Close() {
	// Connection is managed externally
}

// NATSJobWorker processes jobs from the NATS queue.
// Implements Requirement 5.3: Encryption worker with concurrency control.
type NATSJobWorker struct {
	queue      *NATSJobQueue
	jobService *JobService
	jobRepo    domain.EncryptionJobRepository
	config     WorkerConfig
	running    atomic.Bool
	wg         sync.WaitGroup
	stopCh     chan struct{}
	consumeCtx jetstream.ConsumeContext
}

// NewNATSJobWorker creates a new NATS job worker.
func NewNATSJobWorker(
	queue *NATSJobQueue,
	jobService *JobService,
	jobRepo domain.EncryptionJobRepository,
	config WorkerConfig,
) *NATSJobWorker {
	return &NATSJobWorker{
		queue:      queue,
		jobService: jobService,
		jobRepo:    jobRepo,
		config:     config,
		stopCh:     make(chan struct{}),
	}
}

// Start starts the NATS job worker.
func (w *NATSJobWorker) Start(ctx context.Context) error {
	if w.running.Load() {
		return fmt.Errorf("worker is already running")
	}

	w.running.Store(true)
	w.stopCh = make(chan struct{})

	log.Info().
		Int("concurrency", w.config.Concurrency).
		Msg("starting NATS job worker")

	// Create a semaphore for concurrency control
	sem := make(chan struct{}, w.config.Concurrency)

	// Start consuming messages
	consumeCtx, err := w.queue.consumer.Consume(func(msg jetstream.Msg) {
		// Acquire semaphore
		select {
		case <-w.stopCh:
			msg.Nak()
			return
		case sem <- struct{}{}:
		}

		// Process in goroutine
		w.wg.Add(1)
		go func() {
			defer w.wg.Done()
			defer func() { <-sem }() // Release semaphore

			w.processMessage(ctx, msg)
		}()
	})
	if err != nil {
		w.running.Store(false)
		return fmt.Errorf("failed to start consuming: %w", err)
	}

	w.consumeCtx = consumeCtx
	return nil
}

// Stop stops the NATS job worker gracefully.
// Implements Requirement 5.7: Graceful worker shutdown.
func (w *NATSJobWorker) Stop() error {
	if !w.running.Load() {
		return nil
	}

	log.Info().Msg("stopping NATS job worker")

	// Signal stop
	close(w.stopCh)

	// Stop consuming new messages
	if w.consumeCtx != nil {
		w.consumeCtx.Stop()
	}

	// Wait for in-flight jobs with timeout
	done := make(chan struct{})
	go func() {
		w.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		log.Info().Msg("NATS job worker stopped gracefully")
	case <-time.After(w.config.ShutdownTimeout):
		log.Warn().Msg("NATS job worker shutdown timed out")
	}

	w.running.Store(false)
	return nil
}

// IsRunning returns whether the worker is running.
func (w *NATSJobWorker) IsRunning() bool {
	return w.running.Load()
}

// processMessage processes a single NATS message.
func (w *NATSJobWorker) processMessage(ctx context.Context, msg jetstream.Msg) {
	// Parse message
	var queueMsg JobQueueMessage
	if err := json.Unmarshal(msg.Data(), &queueMsg); err != nil {
		log.Error().Err(err).Msg("failed to unmarshal job message")
		msg.Term() // Terminal failure, don't retry
		return
	}

	log.Debug().
		Str("job_id", queueMsg.JobID.String()).
		Msg("processing job from NATS queue")

	// Get job from database
	job, err := w.jobRepo.FindByID(ctx, queueMsg.JobID)
	if err != nil {
		log.Error().Err(err).Str("job_id", queueMsg.JobID.String()).Msg("failed to find job")
		msg.Nak() // Retry
		return
	}

	// Skip if already completed or failed
	if job.IsCompleted() || job.IsFailed() {
		log.Debug().
			Str("job_id", job.ID.String()).
			Str("status", string(job.Status)).
			Msg("skipping job with terminal status")
		msg.Ack()
		return
	}

	// Skip if already processing (another worker picked it up)
	if job.IsProcessing() {
		log.Debug().
			Str("job_id", job.ID.String()).
			Msg("skipping job already being processed")
		msg.Ack()
		return
	}

	// Process the job
	err = w.jobService.ProcessJob(ctx, job)
	if err != nil {
		// Check if we should retry
		if job.CanRetry() {
			log.Warn().
				Err(err).
				Str("job_id", job.ID.String()).
				Int("retry_count", job.RetryCount).
				Msg("job failed, will retry")
			msg.NakWithDelay(w.jobService.CalculateRetryDelay(job.RetryCount))
		} else {
			log.Error().
				Err(err).
				Str("job_id", job.ID.String()).
				Msg("job failed permanently")
			msg.Ack() // Don't retry, job is marked as failed in DB
		}
		return
	}

	// Success
	msg.Ack()
}

// HybridJobWorker combines database polling with NATS queue for job processing.
// This provides resilience - jobs can be processed even if NATS is temporarily unavailable.
type HybridJobWorker struct {
	dbWorker   *EncryptionWorker
	natsWorker *NATSJobWorker
	useNATS    bool
}

// NewHybridJobWorker creates a new hybrid job worker.
func NewHybridJobWorker(
	dbWorker *EncryptionWorker,
	natsWorker *NATSJobWorker,
	useNATS bool,
) *HybridJobWorker {
	return &HybridJobWorker{
		dbWorker:   dbWorker,
		natsWorker: natsWorker,
		useNATS:    useNATS,
	}
}

// Start starts the hybrid worker.
func (w *HybridJobWorker) Start(ctx context.Context) error {
	if w.useNATS && w.natsWorker != nil {
		if err := w.natsWorker.Start(ctx); err != nil {
			log.Warn().Err(err).Msg("failed to start NATS worker, falling back to DB polling")
			return w.dbWorker.Start(ctx)
		}
		return nil
	}

	return w.dbWorker.Start(ctx)
}

// Stop stops the hybrid worker.
func (w *HybridJobWorker) Stop() error {
	if w.useNATS && w.natsWorker != nil && w.natsWorker.IsRunning() {
		return w.natsWorker.Stop()
	}

	if w.dbWorker != nil && w.dbWorker.IsRunning() {
		return w.dbWorker.Stop()
	}

	return nil
}

// IsRunning returns whether the worker is running.
func (w *HybridJobWorker) IsRunning() bool {
	if w.useNATS && w.natsWorker != nil {
		return w.natsWorker.IsRunning()
	}
	if w.dbWorker != nil {
		return w.dbWorker.IsRunning()
	}
	return false
}

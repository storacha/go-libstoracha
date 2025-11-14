package queuepoller

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	logging "github.com/ipfs/go-log/v2"
	"github.com/storacha/go-libstoracha/jobqueue"
)

const (
	defaultJobBatchSize = 10
	defaultConcurrency  = 100

	maxJobBatchSize      = 10
	maxJobProcessingTime = 5 * time.Minute
)

var log = logging.Logger("queuepoller")

type (
	WithID[Job any] struct {
		ID  string
		Job Job
	}

	// QueueQueuer is an interface for queuing jobs.
	QueueQueuer[Job any] interface {
		Queue(ctx context.Context, job Job) error
	}

	// QueueReader is an interface for reading jobs from the queue.
	QueueReader[Job any] interface {
		Read(ctx context.Context, maxJobs int) ([]WithID[Job], error)
	}

	// QueueReleaser is an interface for releasing jobs from the queue.
	QueueReleaser interface {
		Release(ctx context.Context, jobID string) error
	}

	// QueueDeleter is an interface for deleting jobs from the queue.
	QueueDeleter interface {
		Delete(ctx context.Context, jobID string) error
	}

	// Queue is an interface for a job queue,
	// combining queuing, reading, releasing, and deleting jobs.
	Queue[Job any] interface {
		QueueQueuer[Job]
		QueueReader[Job]
		QueueReleaser
		QueueDeleter
	}

	// config
	config struct {
		jobBatchSize int
		concurrency  int
	}

	// Option configures the CachingQueuePoller
	Option func(*config)

	// Handler processes jobs of the given type.
	Handler[Job any] interface {
		toJobQueue(queue Queue[Job], concurrency int) jobQueue[Job]
	}

	singleHandler[Job any] struct {
		handler func(ctx context.Context, j Job) error
	}

	batchHandler[Job any] struct {
		handler func(ctx context.Context, jobs []WithID[Job]) map[string]error
	}

	jobQueue[Job any] interface {
		Startup()
		Shutdown(ctx context.Context) error
		Queue(ctx context.Context, job []WithID[Job]) error
	}

	//lint:ignore U1000 https://github.com/dominikh/go-tools/issues/1440
	singleJobQueue[Job any] struct {
		jq *jobqueue.JobQueue[WithID[Job]]
	}

	//lint:ignore U1000 https://github.com/dominikh/go-tools/issues/1440
	batchJobQueue[Job any] struct {
		jq *jobqueue.JobQueue[[]WithID[Job]]
	}

	// QueuePoller polls a queue for  jobs and processes them
	// using the provided JobHandler.
	QueuePoller[Job any] struct {
		queue        Queue[Job]
		handler      Handler[Job]
		jq           jobQueue[Job]
		jobBatchSize int
		ctx          context.Context
		cancel       context.CancelFunc
		stopped      chan struct{}
		startOnce    sync.Once
		stopOnce     sync.Once
	}
)

// WithJobBatchSize sets the maximum number of jobs to process in a batch
func WithJobBatchSize(size int) Option {
	return func(cfg *config) {
		cfg.jobBatchSize = size
	}
}

// WithConcurrency sets the maximum number of concurrent job processing
func WithConcurrency(concurrency int) Option {
	return func(cfg *config) {
		cfg.concurrency = concurrency
	}
}

// JobHandler creates a Handler from a function that processes single jobs.
func JobHandler[Job any](handler func(ctx context.Context, j Job) error) Handler[Job] {
	return &singleHandler[Job]{handler: handler}
}

// BatchJobHandler creates a Handler from a function that processes multiple jobs.
func BatchJobHandler[Job any](handler func(ctx context.Context, jobs []WithID[Job]) map[string]error) Handler[Job] {
	return &batchHandler[Job]{handler: handler}
}

// NewQueuePoller creates a new QueuePoller instance.
func NewQueuePoller[Job any](queue Queue[Job], handler Handler[Job], opts ...Option) (*QueuePoller[Job], error) {
	cfg := &config{
		jobBatchSize: defaultJobBatchSize,
		concurrency:  defaultConcurrency,
	}

	for _, opt := range opts {
		opt(cfg)
	}

	poller := &QueuePoller[Job]{
		queue:        queue,
		handler:      handler,
		jq:           handler.toJobQueue(queue, cfg.concurrency),
		jobBatchSize: cfg.jobBatchSize,
		stopped:      make(chan struct{}),
	}

	if poller.jobBatchSize > maxJobBatchSize {
		return nil, fmt.Errorf("job batch size %d exceeds maximum allowed %d", poller.jobBatchSize, maxJobBatchSize)
	}

	return poller, nil
}

// Start begins polling the queue for caching jobs.
func (p *QueuePoller[Job]) Start() {
	p.startOnce.Do(func() {
		// Create root context
		p.ctx, p.cancel = context.WithCancel(context.Background())
		p.jq.Startup()
		log.Info("Starting caching queue poller")

		go func() {
			for {
				select {
				case <-p.ctx.Done():
					log.Info("Stopping polling loop")
					close(p.stopped)
					return
				default:
					p.processJobs(p.ctx)
				}
			}
		}()
	})
}

// Stop stops the polling loop and waits for it to finish.
func (p *QueuePoller[Job]) Stop() {
	p.stopOnce.Do(func() {
		// Cancel the root context, which will cancel all child contexts
		if p.cancel != nil {
			p.cancel()
		}

		// Wait for the polling loop to finish
		<-p.stopped

		p.jq.Shutdown(p.ctx)
	})
}

// processJobs reads and processes all available jobs from the queue in batches
func (p *QueuePoller[Job]) processJobs(ctx context.Context) {
	// Read a batch of jobs and queue them in the job queue
	jobs, err := p.queue.Read(ctx, p.jobBatchSize)
	if err != nil {
		log.Errorf("Error reading jobs from queue: %v", err)
		return
	}

	err = p.jq.Queue(ctx, jobs)
	if err != nil {
		log.Errorf("Error queuing jobs: %v", err)
	}
}

//lint:ignore U1000 https://github.com/dominikh/go-tools/issues/1440
func (s *singleHandler[Job]) toJobQueue(queue Queue[Job], concurrency int) jobQueue[Job] {
	handler := jobqueue.JobHandler(func(ctx context.Context, job WithID[Job]) error {
		jobCtx, cancel := context.WithTimeout(ctx, maxJobProcessingTime)
		defer cancel()

		// Process the job
		err := s.handler(jobCtx, job.Job)
		if err != nil && !errors.Is(err, context.DeadlineExceeded) {
			// if the error is not a timeout, make the job visible so that it can be retried
			if err := queue.Release(ctx, job.ID); err != nil {
				log.Warnf("Failed to release job %s: %s", job.ID, err)
			}

			return fmt.Errorf("failed to perform job %s: %w", job.ID, err)
		}

		// Do not hold up the queue by re-attempting a job that times out.
		// Log the error and proceed with deletion.
		if errors.Is(err, context.DeadlineExceeded) {
			log.Warnf("Not retrying provider job for %s: %s", job.ID, err)
		}

		// Delete the job too if processing was successful
		if err := queue.Delete(ctx, job.ID); err != nil {
			return fmt.Errorf("failed to delete job %s: %w", job.ID, err)
		}

		return nil
	})
	return &singleJobQueue[Job]{
		jq: jobqueue.NewJobQueue[WithID[Job]](
			handler,
			jobqueue.WithConcurrency(concurrency),
			jobqueue.WithErrorHandler(func(err error) {
				log.Errorw("processing job", "error", err)
			})),
	}
}

func (s *singleJobQueue[Job]) Startup() {
	s.jq.Startup()
}

func (s *singleJobQueue[Job]) Shutdown(ctx context.Context) error {
	return s.jq.Shutdown(ctx)
}

func (s *singleJobQueue[Job]) Queue(ctx context.Context, job []WithID[Job]) error {
	for _, j := range job {
		if err := s.jq.Queue(ctx, j); err != nil {
			return err
		}
	}
	return nil
}

//lint:ignore U1000 https://github.com/dominikh/go-tools/issues/1440
func (b *batchHandler[Job]) toJobQueue(queue Queue[Job], concurrency int) jobQueue[Job] {
	handler := jobqueue.JobHandler(func(ctx context.Context, jobs []WithID[Job]) error {
		jobCtx, cancel := context.WithTimeout(ctx, maxJobProcessingTime)
		defer cancel()

		// Process the jobs
		errMap := b.handler(jobCtx, jobs)

		// Handle individual job results
		for _, job := range jobs {
			jobErr := errMap[job.ID]
			if jobErr != nil && !errors.Is(jobErr, context.DeadlineExceeded) {
				// if the error is not a timeout, make the job visible so that it can be retried
				if err := queue.Release(ctx, job.ID); err != nil {
					log.Warnf("Failed to release job %s: %s", job.ID, err)
				}

				log.Errorf("Failed to perform job %s: %s", job.ID, jobErr)
				continue
			}

			// Do not hold up the queue by re-attempting a job that times out.
			// Log the error and proceed with deletion.
			if errors.Is(jobErr, context.DeadlineExceeded) {
				log.Warnf("Not retrying provider job for %s: %s", job.ID, jobErr)
			}

			// Delete the job too if processing was successful
			if err := queue.Delete(ctx, job.ID); err != nil {
				log.Errorf("Failed to delete job %s: %s", job.ID, err)
			}
		}

		return nil
	})
	return &batchJobQueue[Job]{
		jq: jobqueue.NewJobQueue[[]WithID[Job]](
			handler,
			jobqueue.WithConcurrency(concurrency),
			jobqueue.WithErrorHandler(func(err error) {
				log.Errorw("processing job", "error", err)
			}),
		),
	}
}

func (b *batchJobQueue[Job]) Startup() {
	b.jq.Startup()
}

func (b *batchJobQueue[Job]) Shutdown(ctx context.Context) error {
	return b.jq.Shutdown(ctx)
}

func (b *batchJobQueue[Job]) Queue(ctx context.Context, jobs []WithID[Job]) error {
	return b.jq.Queue(ctx, jobs)
}

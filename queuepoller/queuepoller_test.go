package queuepoller

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"

	"testing/synctest"
)

type testJob struct {
	id string
}

// mockQueue implements the Queue[testJob] interface for testing
type mockQueue struct {
	mu              sync.Mutex
	jobs            []WithID[testJob]
	jobsWaiting     chan struct{}
	readCount       int
	deletedJobs     map[string]struct{}
	releasedJobs    map[string]struct{}
	readBehavior    func(ctx context.Context, maxJobs int) ([]WithID[testJob], error)
	deleteBehavior  func(ctx context.Context, jobID string) error
	releaseBehavior func(ctx context.Context, jobID string) error
}

func newMockQueue() *mockQueue {
	return &mockQueue{
		jobsWaiting:  make(chan struct{}, 1),
		deletedJobs:  make(map[string]struct{}),
		releasedJobs: make(map[string]struct{}),
	}
}

func (m *mockQueue) Queue(ctx context.Context, job testJob) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.jobs = append(m.jobs, WithID[testJob]{ID: job.id, Job: job})
	select {
	case m.jobsWaiting <- struct{}{}:
	default:
	}
	return nil
}

func (m *mockQueue) Read(ctx context.Context, maxJobs int) ([]WithID[testJob], error) {

	select {
	case <-m.jobsWaiting:
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	// Allow custom read behavior
	if m.readBehavior != nil {
		return m.readBehavior(ctx, maxJobs)
	}

	m.readCount++
	if len(m.jobs) == 0 {
		return nil, nil
	}

	// Return up to maxJobs
	n := min(maxJobs, len(m.jobs))

	result := m.jobs[:n]
	m.jobs = m.jobs[n:]
	if len(m.jobs) > 0 {
		select {
		case m.jobsWaiting <- struct{}{}:
		default:
		}
	}
	return result, nil
}

func (m *mockQueue) Delete(ctx context.Context, jobID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.deleteBehavior != nil {
		return m.deleteBehavior(ctx, jobID)
	}

	m.deletedJobs[jobID] = struct{}{}
	return nil
}

func (m *mockQueue) Release(ctx context.Context, jobID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.releaseBehavior != nil {
		return m.releaseBehavior(ctx, jobID)
	}

	m.releasedJobs[jobID] = struct{}{}
	return nil
}

func (m *mockQueue) getDeletedJobs() []string {
	m.mu.Lock()
	defer m.mu.Unlock()
	result := make([]string, 0, len(m.deletedJobs))
	for id := range m.deletedJobs {
		result = append(result, id)
	}
	return result
}

func (m *mockQueue) getReleasedJobs() []string {
	m.mu.Lock()
	defer m.mu.Unlock()
	result := make([]string, 0, len(m.releasedJobs))
	for id := range m.releasedJobs {
		result = append(result, id)
	}
	return result
}

// Test that jobs in the queue are read and processed successfully
func TestQueuePollerJobHandlerSuccess(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		queue := newMockQueue()
		processedJobs := make([]testJob, 0)

		handler := JobHandler(func(ctx context.Context, job testJob) error {
			processedJobs = append(processedJobs, job)
			return nil
		})

		poller, err := NewQueuePoller(
			queue,
			handler,
			WithJobBatchSize(10),
			WithConcurrency(1),
		)
		if err != nil {
			t.Fatalf("failed to create poller: %v", err)
		}

		// Queue some jobs
		testJobs := []testJob{{id: "job1"}, {id: "job2"}, {id: "job3"}}
		for _, job := range testJobs {
			if err := queue.Queue(context.Background(), job); err != nil {
				t.Fatalf("failed to queue job: %v", err)
			}
		}

		// Start the poller and let it process jobs
		poller.Start()
		defer poller.Stop()

		// wait for jobs to be processed
		synctest.Wait()

		// Verify all jobs were processed
		if len(processedJobs) != len(testJobs) {
			t.Errorf("expected %d processed jobs, got %d", len(testJobs), len(processedJobs))
		}

		// Verify all jobs were deleted
		deletedJobs := queue.getDeletedJobs()
		if len(deletedJobs) != len(testJobs) {
			t.Errorf("expected %d deleted jobs, got %d", len(testJobs), len(deletedJobs))
		}

		// Verify no jobs were released
		releasedJobs := queue.getReleasedJobs()
		if len(releasedJobs) != 0 {
			t.Errorf("expected 0 released jobs, got %d", len(releasedJobs))
		}
	})
}

// Test that jobs are released when a non-timeout error occurs
func TestQueuePollerJobHandlerNonTimeoutError(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		queue := newMockQueue()

		testErr := errors.New("processing failed")
		handler := JobHandler(func(ctx context.Context, job testJob) error {
			return testErr
		})

		poller, err := NewQueuePoller(
			queue,
			handler,
			WithJobBatchSize(10),
			WithConcurrency(1),
		)
		if err != nil {
			t.Fatalf("failed to create poller: %v", err)
		}

		// Queue a job
		job := testJob{id: "job1"}
		if err := queue.Queue(context.Background(), job); err != nil {
			t.Fatalf("failed to queue job: %v", err)
		}

		// Start the poller
		poller.Start()
		defer poller.Stop()

		// Give the poller time to process jobs
		synctest.Wait()

		// Verify the job was released
		releasedJobs := queue.getReleasedJobs()
		if len(releasedJobs) != 1 || releasedJobs[0] != "job1" {
			t.Errorf("expected job1 to be released, got %v", releasedJobs)
		}

		// Verify the job was NOT deleted
		deletedJobs := queue.getDeletedJobs()
		if len(deletedJobs) != 0 {
			t.Errorf("expected 0 deleted jobs, got %d", len(deletedJobs))
		}
	})
}

// Test that jobs are deleted even when context.DeadlineExceeded occurs
func TestQueuePollerJobHandlerTimeoutError(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		queue := newMockQueue()

		handler := JobHandler(func(ctx context.Context, job testJob) error {
			return context.DeadlineExceeded
		})

		poller, err := NewQueuePoller(
			queue,
			handler,
			WithJobBatchSize(10),
			WithConcurrency(1),
		)
		if err != nil {
			t.Fatalf("failed to create poller: %v", err)
		}

		// Queue a job
		job := testJob{id: "job1"}
		if err := queue.Queue(context.Background(), job); err != nil {
			t.Fatalf("failed to queue job: %v", err)
		}

		// Start the poller
		poller.Start()
		defer poller.Stop()

		// Give the poller time to process jobs
		synctest.Wait()

		// Verify the job was deleted (not released)
		deletedJobs := queue.getDeletedJobs()
		if len(deletedJobs) != 1 || deletedJobs[0] != "job1" {
			t.Errorf("expected job1 to be deleted, got %v", deletedJobs)
		}

		// Verify the job was NOT released
		releasedJobs := queue.getReleasedJobs()
		if len(releasedJobs) != 0 {
			t.Errorf("expected 0 released jobs, got %d", len(releasedJobs))
		}
	})
}

// Test that BatchJobHandler processes batches correctly with mixed results
func TestQueuePollerBatchJobHandlerMixedResults(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		queue := newMockQueue()

		// Create a handler that returns different results for different jobs
		handler := BatchJobHandler(func(ctx context.Context, jobs []WithID[testJob]) map[string]error {
			results := make(map[string]error)
			for _, job := range jobs {
				switch job.ID {
				case "job1":
					// Success
					results[job.ID] = nil
				case "job2":
					// Non-timeout error
					results[job.ID] = errors.New("processing failed")
				case "job3":
					// Timeout error
					results[job.ID] = context.DeadlineExceeded
				}
			}
			return results
		})

		poller, err := NewQueuePoller(
			queue,
			handler,
			WithJobBatchSize(10),
			WithConcurrency(1),
		)
		if err != nil {
			t.Fatalf("failed to create poller: %v", err)
		}

		// Queue the test jobs
		testJobs := []testJob{{id: "job1"}, {id: "job2"}, {id: "job3"}}
		for _, job := range testJobs {
			if err := queue.Queue(context.Background(), job); err != nil {
				t.Fatalf("failed to queue job: %v", err)
			}
		}

		// Start the poller
		poller.Start()
		defer poller.Stop()

		// Give the poller time to process jobs
		synctest.Wait()

		// Verify job1 was deleted (success)
		deletedJobs := queue.getDeletedJobs()
		if len(deletedJobs) != 2 {
			t.Errorf("expected 2 deleted jobs (job1, job3), got %d: %v", len(deletedJobs), deletedJobs)
		}

		// Verify job2 was released (non-timeout error)
		releasedJobs := queue.getReleasedJobs()
		if len(releasedJobs) != 1 || releasedJobs[0] != "job2" {
			t.Errorf("expected job2 to be released, got %v", releasedJobs)
		}
	})
}

// Test that the poller respects the batch size
func TestQueuePollerBatchSize(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		queue := newMockQueue()

		batchCount := 0
		handler := BatchJobHandler(func(ctx context.Context, jobs []WithID[testJob]) map[string]error {
			batchCount++
			return nil
		})

		poller, err := NewQueuePoller(
			queue,
			handler,
			WithJobBatchSize(2),
			WithConcurrency(1),
		)
		if err != nil {
			t.Fatalf("failed to create poller: %v", err)
		}

		// Queue 5 jobs (should be read in batches of 2)
		for i := 1; i <= 5; i++ {
			job := testJob{id: fmt.Sprintf("job%d", i)}
			if err := queue.Queue(context.Background(), job); err != nil {
				t.Fatalf("failed to queue job: %v", err)
			}
		}

		// Start the poller
		poller.Start()
		defer poller.Stop()

		// Give the poller time to process jobs
		synctest.Wait()

		// Verify that 3 batches were processed (2, 2, 1)
		if batchCount != 3 {
			t.Errorf("expected 3 batches processed, got %d", batchCount)
		}

		// Verify all jobs were deleted
		deletedJobs := queue.getDeletedJobs()
		if len(deletedJobs) != 5 {
			t.Errorf("expected 5 deleted jobs, got %d", len(deletedJobs))
		}
	})
}

// Test that rejecting a batch size exceeding the maximum fails
func TestQueuePollerMaxBatchSize(t *testing.T) {
	queue := newMockQueue()
	handler := JobHandler(func(ctx context.Context, job testJob) error {
		return nil
	})

	// Try to create a poller with batch size exceeding the maximum
	_, err := NewQueuePoller(
		queue,
		handler,
		WithJobBatchSize(20), // max is 10
		WithConcurrency(1),
	)

	if err == nil {
		t.Error("expected error when batch size exceeds maximum")
	}
}

package jobqueue_test

import (
	"context"
	"slices"
	"sync"
	"testing"
	"time"

	"github.com/storacha/go-libstoracha/jobqueue"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Enqueues three jobs into a single-handler queue and verifies that all are processed before shutdown.
func TestJobQueueSingleHandlerBasic(t *testing.T) {
	var mu sync.Mutex
	processed := make([]int, 0)

	// Single handler that appends each job to a slice under a lock.
	h := jobqueue.JobHandler(func(ctx context.Context, job int) error {
		mu.Lock()
		defer mu.Unlock()
		processed = append(processed, job)
		return nil
	})

	q := jobqueue.NewJobQueue[int](h)
	q.Startup()

	ctx := context.Background()
	require.NoError(t, q.Queue(ctx, 1))
	require.NoError(t, q.Queue(ctx, 2))
	require.NoError(t, q.Queue(ctx, 3))

	shutdownCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	require.NoError(t, q.Shutdown(shutdownCtx))

	// Check that all jobs have been processed
	assert.ElementsMatch(t, []int{1, 2, 3}, processed)
}

// Exercises concurrency by queueing multiple jobs in parallel with a concurrency level of 4. Verifies all jobs are handled
func TestJobQueueSingleHandlerConcurrency(t *testing.T) {
	var mu sync.Mutex
	var processed []int

	h := jobqueue.JobHandler(func(ctx context.Context, j int) error {
		// Simulate random processing time
		time.Sleep(5 * time.Millisecond)
		mu.Lock()
		defer mu.Unlock()
		processed = append(processed, j)
		return nil
	})

	q := jobqueue.NewJobQueue[int](h,
		jobqueue.WithConcurrency(4),
	)

	q.Startup()

	ctx := context.Background()
	jobCount := 20
	for i := 1; i <= jobCount; i++ {
		require.NoError(t, q.Queue(ctx, i))
	}

	shutdownCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	require.NoError(t, q.Shutdown(shutdownCtx))

	// We expect all jobs to be processed
	require.Len(t, processed, jobCount, "expected all jobs to be processed")
}

// Uses a multi-handler that reads all available jobs into a slice. Verifies that all jobs are seen in total.
func TestJobQueueMultiHandlerBasicBatching(t *testing.T) {
	var mu sync.Mutex
	var allBatches [][]int

	// multi handler that collects all the items that come in a batch
	mh := jobqueue.MultiJobHandler(func(ctx context.Context, jobs []int) error {
		mu.Lock()
		defer mu.Unlock()
		copied := make([]int, len(jobs))
		copy(copied, jobs)
		allBatches = append(allBatches, copied)
		return nil
	})

	q := jobqueue.NewJobQueue[int](mh, jobqueue.WithConcurrency(1))
	q.Startup()

	ctx := context.Background()
	for i := 1; i <= 5; i++ {
		require.NoError(t, q.Queue(ctx, i))
	}

	shutdownCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	require.NoError(t, q.Shutdown(shutdownCtx))

	// Because multiHandler tries to read all available jobs at once, we typically
	// expect them all in a single batch, but it depends on scheduling. This test
	// at least ensures we see all five in total.
	foundItems := 0
	for _, batch := range allBatches {
		foundItems += len(batch)
	}
	assert.Equal(t, 5, foundItems, "expecting total of 5 items processed in batches")
}

// Verifies that queueing a job after calling Shutdown returns ErrQueueShutdown
func TestJobQueueQueueShutdown(t *testing.T) {
	h := jobqueue.JobHandler(func(ctx context.Context, j int) error {
		return nil
	})

	q := jobqueue.NewJobQueue[int](h)
	q.Startup()

	ctx := context.Background()
	require.NoError(t, q.Queue(ctx, 42))

	shutdownCtx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()
	require.NoError(t, q.Shutdown(shutdownCtx))

	// After shutdown, queueing a job should fail with ErrQueueShutdown
	err := q.Queue(ctx, 99)
	assert.ErrorIs(t, err, jobqueue.ErrQueueShutdown)
}

// Verifies that if the context is canceled before calling Queue, the call returns context.Canceled.
func TestJobQueueContextCancellation(t *testing.T) {
	h := jobqueue.JobHandler(func(ctx context.Context, job int) error {
		return nil
	})

	q := jobqueue.NewJobQueue[int](h)
	q.Startup()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// queueing after the context is canceled
	err := q.Queue(ctx, 123)
	require.Error(t, err)
	assert.ErrorIs(t, err, context.Canceled)
}

// Demonstrates buffering behavior. The queue is given a small concurrency level (1) but a buffer of 2.
// We enqueue 4 jobs, confirming that after the first 3 are accepted, the fourth must wait until the queue has capacity,
// and eventually, all 4 get processed
func TestJobQueueBuffer(t *testing.T) {
	var mu sync.Mutex
	var processed []int

	h := jobqueue.JobHandler(func(ctx context.Context, j int) error {
		time.Sleep(50 * time.Millisecond)
		mu.Lock()
		defer mu.Unlock()
		processed = append(processed, j)
		return nil
	})

	q := jobqueue.NewJobQueue[int](h,
		jobqueue.WithBuffer(2),
		jobqueue.WithConcurrency(1),
	)

	q.Startup()

	ctx := context.Background()
	// We push more jobs than concurrency. The buffer is 2, so total of 3 can be
	// queued at once (one actively processed, 2 in the channel).
	for i := 0; i <= 2; i++ {
		require.NoError(t, q.Queue(ctx, i))
	}

	// This next queue operation should block until at least one job is processed
	// or context canceled. We'll do it in a separate goroutine with a short delay.
	var queueErr error
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		queueErr = q.Queue(ctx, 99)
		wg.Done()
	}()

	wg.Wait()

	shutdownCtx, cancel := context.WithTimeout(ctx, 4*time.Second)
	defer cancel()
	assert.NoError(t, q.Shutdown(shutdownCtx), "queue shut down should be successful")

	// The queueErr must not be an error at this point; the capacity eventually
	// freed up and job #99 was queued
	assert.NoError(t, queueErr)

	// We expect all 4 jobs to eventually be processed
	assert.Len(t, processed, 4, "all jobs should have processed")
}

func TestJobQueueStress(t *testing.T) {
	ctx := context.Background()
	var mu sync.Mutex
	var processed []int

	h := jobqueue.JobHandler(func(ctx context.Context, j int) error {
		time.Sleep(1 * time.Millisecond)
		mu.Lock()
		defer mu.Unlock()
		processed = append(processed, j)
		return nil
	})

	q := jobqueue.NewJobQueue[int](h,
		jobqueue.WithBuffer(5),
		jobqueue.WithConcurrency(5),
	)

	q.Startup()

	for i := range 10_000 {
		require.NoError(t, q.Queue(ctx, i))
	}

	require.NoError(t, q.Shutdown(ctx))

	require.Equal(t, len(processed), 10_000)
	for i := range 10_000 {
		require.True(t, slices.Contains(processed, i))
	}
}

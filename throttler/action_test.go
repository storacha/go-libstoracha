package throttler_test

import (
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/storacha/go-libstoracha/throttler"
	"github.com/stretchr/testify/require"
)

func TestActionImmediateExecution(t *testing.T) {
	var called atomic.Int32
	a := throttler.NewAction(func(v int) error {
		called.Add(1)
		return nil
	}, 100*time.Millisecond)
	defer a.Close()

	err := a.Execute(1)
	require.NoError(t, err)
	require.Equal(t, int32(1), called.Load())
}

func TestActionReturnsError(t *testing.T) {
	want := errors.New("boom")
	a := throttler.NewAction(func(v int) error {
		return want
	}, 100*time.Millisecond)
	defer a.Close()

	err := a.Execute(1)
	require.ErrorIs(t, err, want)
}

func TestActionThrottlesRapidCalls(t *testing.T) {
	var mu sync.Mutex
	var args []int
	a := throttler.NewAction(func(v int) error {
		mu.Lock()
		args = append(args, v)
		mu.Unlock()
		return nil
	}, 50*time.Millisecond)
	defer a.Close()

	// First call executes immediately (leading edge).
	require.NoError(t, a.Execute(1))

	// Queue several calls rapidly; only the last should execute after the delay.
	var wg sync.WaitGroup
	for i := 2; i <= 5; i++ {
		wg.Add(1)
		v := i
		go func() {
			defer wg.Done()
			_ = a.Execute(v)
		}()
		time.Sleep(5 * time.Millisecond)
	}
	wg.Wait()

	mu.Lock()
	defer mu.Unlock()
	require.Equal(t, 1, args[0], "first arg should be 1 (immediate)")
	require.Equal(t, 5, args[len(args)-1], "last arg should be 5 (most recent wins)")
}

func TestActionMostRecentArgWins(t *testing.T) {
	// Block the first execution so we can queue multiple pendingExec replacements.
	block := make(chan struct{})
	var mu sync.Mutex
	var args []int

	a := throttler.NewAction(func(v int) error {
		<-block
		mu.Lock()
		args = append(args, v)
		mu.Unlock()
		return nil
	}, 10*time.Millisecond)
	defer a.Close()

	// Start first execution (will block).
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		_ = a.Execute(1)
	}()

	// Give goroutine time to start and block inside the action.
	time.Sleep(20 * time.Millisecond)

	// Queue three more calls while the first is in-flight; each replaces the pending.
	wg.Add(3)
	for _, v := range []int{2, 3, 4} {
		val := v
		go func() {
			defer wg.Done()
			_ = a.Execute(val)
		}()
		time.Sleep(5 * time.Millisecond)
	}

	// Unblock the first execution.
	close(block)
	wg.Wait()

	mu.Lock()
	defer mu.Unlock()
	require.Equal(t, 1, args[0], "first executed arg should be 1")
	require.Equal(t, 4, args[len(args)-1], "last executed arg should be 4")
}

func TestActionClosePreventsFutureExecution(t *testing.T) {
	var called atomic.Int32
	a := throttler.NewAction(func(v int) error {
		called.Add(1)
		return nil
	}, 10*time.Millisecond)

	a.Close()

	err := a.Execute(1)
	require.Error(t, err, "expected error after Close")
	require.Equal(t, int32(0), called.Load(), "action should not be called after Close")
}

func TestActionSubsequentCallAfterDelay(t *testing.T) {
	var called atomic.Int32
	a := throttler.NewAction(func(v int) error {
		called.Add(1)
		return nil
	}, 30*time.Millisecond)
	defer a.Close()

	require.NoError(t, a.Execute(1))
	// Wait for delay to expire.
	time.Sleep(60 * time.Millisecond)
	require.NoError(t, a.Execute(2))

	require.Equal(t, int32(2), called.Load())
}

func TestActionConcurrentExecuteSafety(t *testing.T) {
	var called atomic.Int32
	a := throttler.NewAction(func(v int) error {
		called.Add(1)
		time.Sleep(5 * time.Millisecond)
		return nil
	}, 10*time.Millisecond)
	defer a.Close()

	const n = 50
	var wg sync.WaitGroup
	wg.Add(n)
	for i := range n {
		v := i
		go func() {
			defer wg.Done()
			_ = a.Execute(v)
		}()
	}
	wg.Wait()

	require.Positive(t, called.Load(), "expected at least one execution")
}

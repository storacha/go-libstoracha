package throttler

import (
	"context"
	"sync"
	"time"
)

type Action[T any] struct {
	action      func(T) error
	mu          sync.Mutex
	pendingExec *execRequest[T]
	executing   bool
	delay       time.Duration
	ctx         context.Context
	cancel      context.CancelFunc
}

// execRequest is a request for execution of the action with an argument.
type execRequest[T any] struct {
	arg T
	// result is a channel that will receive the error from executing the action,
	// or nil if the execution was throttled and replaced by a newer request
	// before it could execute.
	result chan error
}

// NewAction creates a new action throttler that will call the provided function
// with the most recent argument at most once every delay duration.
//
// The action is executed on the leading edge of the delay duration, meaning
// that the first execution request will trigger an immediate execution of the
// action, and then subsequent execution requests will be throttled until the
// delay duration has passed since the last execution.
//
// If multiple execution requests are throttled within the delay duration, only
// the most recent execution argument will be passed to the action when the
// delay expires.
func NewAction[T any](action func(T) error, delay time.Duration) *Action[T] {
	ctx, cancel := context.WithCancel(context.Background())
	return &Action[T]{
		action: action,
		ctx:    ctx,
		cancel: cancel,
		delay:  delay,
	}
}

func (t *Action[T]) maybeExec() {
	t.mu.Lock()
	defer t.mu.Unlock()

	// If there is no pending execution, then there is nothing to execute, or if
	// there is already an action executing, then do not execute.
	if t.pendingExec == nil || t.executing {
		return
	}

	exec := t.pendingExec
	t.pendingExec = nil

	// If the throttler is closed then signal that the pending execution was not
	// executed and return
	if t.ctx.Err() != nil {
		exec.result <- t.ctx.Err()
		return
	}

	t.executing = true
	go func() {
		exec.result <- t.action(exec.arg)
		select {
		case <-time.After(t.delay):
		case <-t.ctx.Done():
		}
		t.mu.Lock()
		t.executing = false
		t.mu.Unlock()
		t.maybeExec() // try execute the next pending execution, if there is one
	}()
}

// Execute instructs the throttler to execute the action with the provided
// argument, subject to the throttling constraints.
func (t *Action[T]) Execute(arg T) error {
	t.mu.Lock()
	if t.pendingExec != nil {
		// signal that the pending execution was throttled and will not be
		// executed, allowing the sender to return from Execute.
		t.pendingExec.result <- nil
	}
	exec := execRequest[T]{arg: arg, result: make(chan error, 1)}
	t.pendingExec = &exec
	t.mu.Unlock()

	t.maybeExec() // trigger execution in case there is not already an execution in progress

	return <-exec.result
}

// Close prevents any pending or future execution requests from executing but
// does not interrupt any execution that is already in progress. It does not
// wait for any in-progress execution to complete.
func (t *Action[T]) Close() {
	t.cancel()
}

package session

import (
	"context"
	"errors"
	"sync"
	"time"
)

// idleTracker belongs to one admitted Run. Only the watchdog owns its timer;
// callers serialize deadline/activity/closure without doing transport work.
type idleTracker struct {
	mu       sync.Mutex
	ctx      context.Context
	cancel   context.CancelCauseFunc
	budget   time.Duration
	deadline time.Time
	expired  bool
	closed   bool
	wake     chan struct{}
	stop     chan struct{}
	done     chan struct{}
}

func newIdleTracker(parent context.Context, budget time.Duration) *idleTracker {
	ctx, cancel := context.WithCancelCause(parent)
	idle := &idleTracker{
		ctx: ctx, cancel: cancel, budget: budget,
		deadline: time.Now().Add(budget),
		wake:     make(chan struct{}, 1),
		stop:     make(chan struct{}),
		done:     make(chan struct{}),
	}
	go idle.watch()
	return idle
}

func (idle *idleTracker) watch() {
	defer close(idle.done)
	idle.mu.Lock()
	remaining := time.Until(idle.deadline)
	idle.mu.Unlock()
	timer := time.NewTimer(remaining)
	defer timer.Stop()
	for {
		select {
		case <-idle.stop:
			return
		case <-idle.ctx.Done():
			return
		case <-idle.wake:
		case <-timer.C:
		}
		idle.mu.Lock()
		if idle.closed || idle.ctx.Err() != nil {
			idle.mu.Unlock()
			return
		}
		remaining := time.Until(idle.deadline)
		if remaining <= 0 {
			idle.expired = true
			idle.cancel(context.DeadlineExceeded)
			idle.mu.Unlock()
			return
		}
		timer.Reset(remaining)
		idle.mu.Unlock()
	}
}

func (idle *idleTracker) activity() {
	if idle == nil {
		return
	}
	idle.mu.Lock()
	defer idle.mu.Unlock()
	if idle.closed || idle.expired || idle.ctx.Err() != nil {
		return
	}
	now := time.Now()
	if !now.Before(idle.deadline) {
		idle.expired = true
		idle.cancel(context.DeadlineExceeded)
		return
	}
	idle.deadline = now.Add(idle.budget)
	select {
	case idle.wake <- struct{}{}:
	default:
	}
}

// close is shared by Run finalization and Stop. It never waits under either
// lifecycle or tracker locks. The acknowledgement covers timer/goroutine exit.
func (idle *idleTracker) close() {
	if idle == nil {
		return
	}
	idle.mu.Lock()
	if !idle.closed {
		idle.closed = true
		close(idle.stop)
	}
	idle.mu.Unlock()
	<-idle.done
	idle.cancel(context.Canceled)
}

func (idle *idleTracker) result(parent context.Context, err error) error {
	if idle == nil {
		return err
	}
	// close has joined the watchdog, so expiry cannot change after this read.
	if parent.Err() != nil {
		return errors.Join(parent.Err(), err)
	}
	idle.mu.Lock()
	expired := idle.expired
	idle.mu.Unlock()
	if expired {
		return errors.Join(err, context.DeadlineExceeded)
	}
	return err
}

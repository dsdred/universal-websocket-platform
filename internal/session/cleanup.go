package session

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
)

var (
	errNilCancellationObservation = errors.New("connection cancellation observation is nil")
	errNilCancellationOperation   = errors.New("connection cancellation operation is nil")
)

type transportCleanupOutcome uint8

const (
	transportCleanupSucceeded transportCleanupOutcome = iota + 1
	transportCleanupFailed
	transportCleanupPanicked
)

type cancellationOutcome uint8

const (
	cancellationConfirmed cancellationOutcome = iota + 1
	cancellationAnomaly
)

type cleanupPanicCategory uint8

const cleanupPanicNone cleanupPanicCategory = 0

const (
	cleanupPanicStop cleanupPanicCategory = 1 << iota
	cleanupPanicCancellation
)

// cleanupAcknowledgement is a detached immutable summary of observed cleanup.
type cleanupAcknowledgement struct {
	transport    transportCleanupOutcome
	cancellation cancellationOutcome
	panics       cleanupPanicCategory
}

type cancellationDependency struct {
	done   <-chan struct{}
	cancel func()
}

type cancellationCell struct {
	done    <-chan struct{}
	cancel  func()
	invoked atomic.Bool
}

func newCancellationCell(dependency cancellationDependency) (*cancellationCell, error) {
	if dependency.done == nil {
		return nil, errNilCancellationObservation
	}
	if dependency.cancel == nil {
		return nil, errNilCancellationOperation
	}
	return &cancellationCell{
		done:   dependency.done,
		cancel: dependency.cancel,
	}, nil
}

func (cell *cancellationCell) invoke() (panicked bool) {
	if !cell.invoked.CompareAndSwap(false, true) {
		return false
	}
	defer func() {
		if recover() != nil {
			panicked = true
		}
	}()
	cell.cancel()
	return false
}

func (cell *cancellationCell) confirmed() bool {
	select {
	case <-cell.done:
		return true
	default:
		return false
	}
}

type sessionCleanup struct {
	mu           sync.Mutex
	session      *DefaultSession
	cancellation *cancellationCell
	started      bool
	done         chan struct{}
	result       cleanupAcknowledgement
}

func newSessionCleanup(
	runtimeSession *DefaultSession,
	cancellation *cancellationCell,
) *sessionCleanup {
	return &sessionCleanup{
		session:      runtimeSession,
		cancellation: cancellation,
		done:         make(chan struct{}),
	}
}

func (cleanup *sessionCleanup) run(ctx context.Context) cleanupAcknowledgement {
	cleanup.mu.Lock()
	if cleanup.started {
		done := cleanup.done
		cleanup.mu.Unlock()
		<-done
		cleanup.mu.Lock()
		result := cleanup.result
		cleanup.mu.Unlock()
		return result
	}
	cleanup.started = true
	cleanup.mu.Unlock()

	result := cleanup.perform(ctx)

	cleanup.mu.Lock()
	cleanup.result = result
	close(cleanup.done)
	cleanup.mu.Unlock()
	return result
}

func (cleanup *sessionCleanup) perform(ctx context.Context) cleanupAcknowledgement {
	stopErr, stopPanicked := invokeSessionStop(cleanup.session, ctx)
	result := cleanupAcknowledgement{}
	switch {
	case stopPanicked:
		result.transport = transportCleanupPanicked
		result.panics |= cleanupPanicStop
	case stopErr != nil:
		result.transport = transportCleanupFailed
	default:
		result.transport = transportCleanupSucceeded
	}

	if cleanup.cancellation.invoke() {
		result.panics |= cleanupPanicCancellation
	}
	if cleanup.cancellation.confirmed() {
		result.cancellation = cancellationConfirmed
	} else {
		result.cancellation = cancellationAnomaly
	}
	return result
}

func invokeSessionStop(runtimeSession *DefaultSession, ctx context.Context) (
	err error,
	panicked bool,
) {
	defer func() {
		if recover() != nil {
			err = nil
			panicked = true
		}
	}()
	return runtimeSession.Stop(ctx), false
}

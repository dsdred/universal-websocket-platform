// Package sessionmanager provides Runtime Session lifecycle tracking foundations.
package sessionmanager

import (
	"context"
	"errors"
	"sync"
)

var (
	// ErrShutdownNotStarted indicates that Wait was called while Manager is Open.
	ErrShutdownNotStarted = errors.New("Session Manager shutdown has not started")
)

// State is the read-only lifecycle state of a Manager.
type State uint8

const (
	// StateOpen accepts normal lifecycle work.
	StateOpen State = iota
	// StateClosing rejects new lifecycle work and is completing shutdown.
	StateClosing
	// StateClosed is terminal.
	StateClosed
)

// Manager owns the minimal Runtime Session Manager lifecycle.
type Manager struct {
	mu    sync.RWMutex
	state State
}

// New creates an Open Manager.
func New() *Manager {
	return &Manager{state: StateOpen}
}

// BeginShutdown atomically starts the single Manager shutdown cycle.
// It is nonblocking and idempotent in Closing and Closed.
func (manager *Manager) BeginShutdown() {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	if manager.state == StateOpen {
		manager.state = StateClosing
	}
}

// Wait observes shutdown completion.
// With no accounting in the lifecycle skeleton, Closing completes immediately.
// Wait does not implicitly start shutdown while Manager is Open.
func (manager *Manager) Wait(_ context.Context) error {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	switch manager.state {
	case StateOpen:
		return ErrShutdownNotStarted
	case StateClosing:
		manager.state = StateClosed
	}

	return nil
}

// State returns the current lifecycle state without exposing mutable state.
func (manager *Manager) State() State {
	manager.mu.RLock()
	defer manager.mu.RUnlock()
	return manager.state
}

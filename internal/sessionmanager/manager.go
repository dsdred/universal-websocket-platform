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
	// ErrManagerNotOpen indicates that an operation requires an Open Manager.
	ErrManagerNotOpen = errors.New("Session Manager is not open")
	// ErrInvalidSessionID indicates that a Session identifier is empty or too long.
	ErrInvalidSessionID = errors.New("invalid Session ID")
	// ErrSessionIDReserved indicates that a Session identifier has an active Reservation.
	ErrSessionIDReserved = errors.New("Session ID is already reserved")
	// ErrRegistrationIDExhausted indicates that Manager cannot allocate another identity.
	ErrRegistrationIDExhausted = errors.New("Registration ID space exhausted")
	// ErrReservationAborted indicates that Commit was attempted after Abort won.
	ErrReservationAborted = errors.New("Session Reservation is aborted")
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
	mu                 sync.RWMutex
	state              State
	nextRegistrationID uint64
	reservations       map[RegistrationID]*reservation
	reservedSessions   map[SessionID]RegistrationID
	registrations      map[RegistrationID]*registration
	registeredSessions map[SessionID]RegistrationID
}

// New creates an Open Manager.
func New() *Manager {
	return &Manager{
		state:              StateOpen,
		nextRegistrationID: 1,
		reservations:       make(map[RegistrationID]*reservation),
		reservedSessions:   make(map[SessionID]RegistrationID),
		registrations:      make(map[RegistrationID]*registration),
		registeredSessions: make(map[SessionID]RegistrationID),
	}
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

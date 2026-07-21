package sessionmanager

import (
	"math"

	"github.com/dsdred/universal-websocket-platform/internal/executionbinding"
)

const maxSessionIDBytes = 255

// ReservationHandle owns one pending Reservation.
type ReservationHandle interface {
	Commit(CommitInput) (CommitResult, error)
	Abort()
	AbortUnlessCommitted()
}

type reservationState uint8

const (
	reservationActive reservationState = iota
	reservationCommitted
	reservationAborted
)

type reservation struct {
	registrationID RegistrationID
	sessionID      SessionID
	state          reservationState
}

type registration struct {
	registrationID RegistrationID
	sessionID      SessionID
	commitResult   CommitResult
	stop           StopPublicationBinding
}

type reservationHandle struct {
	manager     *Manager
	reservation *reservation
}

// Reserve creates one unpublished Reservation owned by the returned Handle.
func (manager *Manager) Reserve(sessionID SessionID) (ReservationHandle, error) {
	if len(sessionID) == 0 || len(sessionID) > maxSessionIDBytes {
		return nil, ErrInvalidSessionID
	}

	manager.mu.Lock()
	defer manager.mu.Unlock()

	if manager.state != StateOpen {
		return nil, ErrManagerNotOpen
	}
	if _, exists := manager.reservedSessions[sessionID]; exists {
		return nil, ErrSessionIDReserved
	}
	if _, exists := manager.registeredSessions[sessionID]; exists {
		return nil, ErrSessionIDReserved
	}
	if manager.nextRegistrationID == 0 {
		return nil, ErrRegistrationIDExhausted
	}

	registrationID := RegistrationID{value: manager.nextRegistrationID}
	if manager.nextRegistrationID == math.MaxUint64 {
		manager.nextRegistrationID = 0
	} else {
		manager.nextRegistrationID++
	}

	reservation := &reservation{
		registrationID: registrationID,
		sessionID:      sessionID,
	}
	handle := &reservationHandle{
		manager:     manager,
		reservation: reservation,
	}
	manager.reservations[registrationID] = reservation
	manager.reservedSessions[sessionID] = registrationID
	return handle, nil
}

// Commit atomically publishes the Reservation as one committed Registration.
// Repeated calls return the same identity only while that Registration exists.
func (handle *reservationHandle) Commit(input CommitInput) (CommitResult, error) {
	return handle.manager.commit(handle.reservation, input)
}

// Abort completes the Reservation without publishing it.
func (handle *reservationHandle) Abort() {
	handle.manager.abort(handle.reservation)
}

// AbortUnlessCommitted aborts an active Reservation and preserves a committed one.
func (handle *reservationHandle) AbortUnlessCommitted() {
	handle.Abort()
}

func (manager *Manager) commit(target *reservation, input CommitInput) (CommitResult, error) {
	manager.mu.Lock()
	switch target.state {
	case reservationCommitted:
		result, err := manager.committedResultLocked(target)
		manager.mu.Unlock()
		return result, err
	case reservationAborted:
		manager.mu.Unlock()
		return CommitResult{}, ErrReservationAborted
	}
	if manager.state != StateOpen {
		manager.mu.Unlock()
		return CommitResult{}, ErrManagerNotOpen
	}
	current, exists := manager.reservations[target.registrationID]
	if !exists || current != target {
		manager.mu.Unlock()
		return CommitResult{}, ErrReservationAborted
	}
	manager.mu.Unlock()

	if err := input.validate(); err != nil {
		return CommitResult{}, err
	}
	result, err := prepareCommitResult(manager, target.registrationID)
	if err != nil {
		return CommitResult{}, err
	}

	manager.mu.Lock()
	defer manager.mu.Unlock()

	switch target.state {
	case reservationCommitted:
		return manager.committedResultLocked(target)
	case reservationAborted:
		return CommitResult{}, ErrReservationAborted
	}

	if manager.state != StateOpen {
		return CommitResult{}, ErrManagerNotOpen
	}

	current, exists = manager.reservations[target.registrationID]
	if !exists || current != target {
		return CommitResult{}, ErrReservationAborted
	}
	if err := input.validate(); err != nil {
		return CommitResult{}, err
	}
	committed := &registration{
		registrationID: target.registrationID,
		sessionID:      target.sessionID,
		commitResult:   result,
		stop:           input.stop,
	}
	target.state = reservationCommitted
	delete(manager.reservations, target.registrationID)
	delete(manager.reservedSessions, target.sessionID)
	manager.registrations[target.registrationID] = committed
	manager.registeredSessions[target.sessionID] = target.registrationID
	manager.lifetimeLeases[target.registrationID] = struct{}{}

	outcome, publicationErr := input.publisher.PublishCommitted()
	if publicationErr != nil || outcome != executionbinding.OutcomeCommitted {
		panic("Session Manager Commit publication invariant violated")
	}

	return result, nil
}

func (manager *Manager) committedResultLocked(target *reservation) (CommitResult, error) {
	registration, exists := manager.registrations[target.registrationID]
	if !exists || registration.sessionID != target.sessionID {
		return CommitResult{}, ErrRegistrationRemoved
	}
	currentID, exists := manager.registeredSessions[target.sessionID]
	if !exists || currentID != target.registrationID {
		return CommitResult{}, ErrRegistrationRemoved
	}
	return registration.commitResult, nil
}

func (manager *Manager) abort(target *reservation) {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	if target.state != reservationActive {
		return
	}

	current, exists := manager.reservations[target.registrationID]
	if !exists || current != target {
		return
	}

	target.state = reservationAborted
	delete(manager.reservations, target.registrationID)
	delete(manager.reservedSessions, target.sessionID)
	manager.closeIfAccountingEmptyLocked()
}

package sessionmanager

import "math"

const maxSessionIDBytes = 255

// ReservationHandle owns one pending Reservation.
// Commit is intentionally absent from the current transaction scope.
type ReservationHandle interface {
	Abort()
	AbortUnlessCommitted()
}

type reservation struct {
	registrationID RegistrationID
	sessionID      SessionID
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

// Abort completes the Reservation without publishing it.
func (handle *reservationHandle) Abort() {
	handle.manager.abort(handle.reservation)
}

// AbortUnlessCommitted currently has the same effect as Abort because Commit
// is outside the current transaction scope.
func (handle *reservationHandle) AbortUnlessCommitted() {
	handle.Abort()
}

func (manager *Manager) abort(target *reservation) {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	current, exists := manager.reservations[target.registrationID]
	if !exists || current != target {
		return
	}

	delete(manager.reservations, target.registrationID)
	delete(manager.reservedSessions, target.sessionID)
}

package sessionmanager

import "slices"

// ShutdownRegistration is immutable committed Registration metadata captured
// at the BeginShutdown linearization point.
type ShutdownRegistration struct {
	sessionID      SessionID
	registrationID RegistrationID
}

// SessionID returns the captured Session identity.
func (registration ShutdownRegistration) SessionID() SessionID {
	return registration.sessionID
}

// RegistrationID returns the captured Registration identity.
func (registration ShutdownRegistration) RegistrationID() RegistrationID {
	return registration.registrationID
}

// ShutdownSnapshot is an immutable value fixed by the first BeginShutdown.
type ShutdownSnapshot struct {
	registrations []ShutdownRegistration
}

// Registrations returns a detached copy of committed registrations captured by
// this Snapshot.
func (snapshot ShutdownSnapshot) Registrations() []ShutdownRegistration {
	return slices.Clone(snapshot.registrations)
}

func (snapshot ShutdownSnapshot) clone() ShutdownSnapshot {
	return ShutdownSnapshot{
		registrations: slices.Clone(snapshot.registrations),
	}
}

func (manager *Manager) captureShutdownSnapshotLocked() ShutdownSnapshot {
	registrations := make([]ShutdownRegistration, 0, len(manager.registrations))
	for _, registration := range manager.registrations {
		registrations = append(registrations, ShutdownRegistration{
			sessionID:      registration.sessionID,
			registrationID: registration.registrationID,
		})
	}
	return ShutdownSnapshot{registrations: registrations}
}

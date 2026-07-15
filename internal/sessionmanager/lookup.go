package sessionmanager

// Lookup returns immutable identity metadata for a committed Registration.
// The returned View does not retain or extend the Registration lifetime.
func (manager *Manager) Lookup(sessionID SessionID) (RegistrationView, bool) {
	manager.mu.RLock()
	defer manager.mu.RUnlock()

	if manager.state == StateClosed {
		return RegistrationView{}, false
	}

	registrationID, exists := manager.registeredSessions[sessionID]
	if !exists {
		return RegistrationView{}, false
	}
	registration, exists := manager.registrations[registrationID]
	if !exists || registration.sessionID != sessionID {
		return RegistrationView{}, false
	}

	return RegistrationView{
		sessionID:      registration.sessionID,
		registrationID: registration.registrationID,
		state:          StateRegistered,
	}, true
}

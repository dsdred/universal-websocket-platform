package sessionmanager

// Complete atomically removes one committed Registration.
// It returns true only for the first completion of a known RegistrationID.
func (manager *Manager) Complete(registrationID RegistrationID) bool {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	registration, exists := manager.registrations[registrationID]
	if !exists {
		return false
	}

	delete(manager.registrations, registrationID)
	if currentID, exists := manager.registeredSessions[registration.sessionID]; exists && currentID == registrationID {
		delete(manager.registeredSessions, registration.sessionID)
	}
	manager.closeIfAccountingEmptyLocked()

	return true
}

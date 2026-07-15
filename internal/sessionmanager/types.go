package sessionmanager

// SessionID is an opaque Runtime Session identifier.
type SessionID string

// RegistrationID is an opaque registration identity.
type RegistrationID struct {
	value uint64
}

// RegistrationView is the immutable shape reserved for future registry lookup.
// It intentionally exposes no Session or lifecycle operations.
type RegistrationView struct {
	sessionID      SessionID
	registrationID RegistrationID
}

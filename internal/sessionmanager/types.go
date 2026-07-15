package sessionmanager

// SessionID is an opaque Runtime Session identifier.
type SessionID string

// RegistrationID is an opaque registration identity.
type RegistrationID struct {
	value uint64
}

// RegistrationView is immutable identity metadata for a committed Registration.
// It intentionally exposes no Session or lifecycle operations.
type RegistrationView struct {
	sessionID      SessionID
	registrationID RegistrationID
}

// SessionID returns the Session identity captured by this View.
func (view RegistrationView) SessionID() SessionID {
	return view.sessionID
}

// RegistrationID returns the Registration identity captured by this View.
func (view RegistrationView) RegistrationID() RegistrationID {
	return view.registrationID
}

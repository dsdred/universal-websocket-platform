package sessionmanager

// SessionID is an opaque Runtime Session identifier.
type SessionID string

// RegistrationID is an opaque registration identity.
type RegistrationID struct {
	value uint64
}

// RegistrationState is the manager-visible state captured by a RegistrationView.
type RegistrationState uint8

const (
	// StateRegistered identifies a committed Registration visible to Manager.
	StateRegistered RegistrationState = iota + 1
)

// RegistrationView is immutable identity metadata for a committed Registration.
// It intentionally exposes no Session or lifecycle operations.
type RegistrationView struct {
	sessionID      SessionID
	registrationID RegistrationID
	state          RegistrationState
}

// SessionID returns the Session identity captured by this View.
func (view RegistrationView) SessionID() SessionID {
	return view.sessionID
}

// RegistrationID returns the Registration identity captured by this View.
func (view RegistrationView) RegistrationID() RegistrationID {
	return view.registrationID
}

// State returns the manager-visible state captured by this View.
func (view RegistrationView) State() RegistrationState {
	return view.state
}

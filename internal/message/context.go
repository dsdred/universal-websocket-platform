package message

import (
	"errors"
	"fmt"
	"reflect"
)

// ErrInvalidContext indicates that Runtime Message Context construction received inconsistent input.
var ErrInvalidContext = errors.New("invalid Runtime Message Context")

// Context is an immutable transport-neutral envelope for one Runtime Message.
type Context struct {
	runtimeMessage         Message
	sender                 Sender
	sessionID              string
	authenticated          bool
	anonymous              bool
	authenticationType     string
	authenticationProvider string
}

// NewContext validates and copies the values for one Runtime Message Context.
func NewContext(
	runtimeMessage *Message,
	sender Sender,
	sessionID string,
	authenticated bool,
	anonymous bool,
	authenticationType string,
	authenticationProvider string,
) (Context, error) {
	if runtimeMessage == nil {
		return Context{}, invalidContext("Message is nil")
	}
	if runtimeMessage.Type() != TypeText && runtimeMessage.Type() != TypeBinary {
		return Context{}, invalidContext("Message type is invalid")
	}
	if nilSender(sender) {
		return Context{}, invalidContext("Sender is nil")
	}
	if sessionID == "" {
		return Context{}, invalidContext("Session ID is empty")
	}
	if authenticated == anonymous {
		return Context{}, invalidContext("identity must be either authenticated or anonymous")
	}
	if authenticated && (authenticationType == "" || authenticationProvider == "") {
		return Context{}, invalidContext("authenticated identity requires Authentication metadata")
	}
	if anonymous && (authenticationType != "" || authenticationProvider != "") {
		return Context{}, invalidContext("anonymous identity must not contain Authentication metadata")
	}

	return Context{
		runtimeMessage:         cloneMessage(*runtimeMessage),
		sender:                 sender,
		sessionID:              sessionID,
		authenticated:          authenticated,
		anonymous:              anonymous,
		authenticationType:     authenticationType,
		authenticationProvider: authenticationProvider,
	}, nil
}

// Message returns an independent immutable Runtime Message value.
func (runtimeContext Context) Message() Message {
	return cloneMessage(runtimeContext.runtimeMessage)
}

// MessageType returns the Runtime Message type without copying its payload.
func (runtimeContext Context) MessageType() Type {
	return runtimeContext.runtimeMessage.Type()
}

// Sender returns the transport-neutral outbound capability for the current Session.
func (runtimeContext Context) Sender() Sender {
	return runtimeContext.sender
}

// SessionID returns the opaque Session identifier.
func (runtimeContext Context) SessionID() string {
	return runtimeContext.sessionID
}

// Authenticated reports whether the Context represents an authenticated identity.
func (runtimeContext Context) Authenticated() bool {
	return runtimeContext.authenticated
}

// Anonymous reports whether the Context represents an explicitly anonymous identity.
func (runtimeContext Context) Anonymous() bool {
	return runtimeContext.anonymous
}

// AuthenticationType returns the Authentication Provider type, or an empty value for anonymous identity.
func (runtimeContext Context) AuthenticationType() string {
	return runtimeContext.authenticationType
}

// AuthenticationProvider returns the configured Provider name, or an empty value for anonymous identity.
func (runtimeContext Context) AuthenticationProvider() string {
	return runtimeContext.authenticationProvider
}

func cloneMessage(runtimeMessage Message) Message {
	runtimeMessage.data = append([]byte(nil), runtimeMessage.data...)
	return runtimeMessage
}

func nilSender(sender Sender) bool {
	if sender == nil {
		return true
	}

	value := reflect.ValueOf(sender)
	switch value.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return value.IsNil()
	default:
		return false
	}
}

func invalidContext(reason string) error {
	return fmt.Errorf("%w: %s", ErrInvalidContext, reason)
}

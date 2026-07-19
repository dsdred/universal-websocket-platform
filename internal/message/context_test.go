package message

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"sync"
	"testing"
)

func TestNewContextCreatesAuthenticatedContext(t *testing.T) {
	runtimeMessage := mustContextMessage(t, TypeText, []byte("authenticated payload"))
	sender := &contextSender{}

	runtimeContext, err := NewContext(
		&runtimeMessage,
		sender,
		"session-authenticated",
		true,
		false,
		"jwt",
		"internal-jwt",
	)
	if err != nil {
		t.Fatalf("NewContext() error = %v", err)
	}

	assertContextMessage(t, runtimeContext, runtimeMessage)
	if runtimeContext.Sender() != sender {
		t.Fatal("Sender() did not return the supplied capability")
	}
	if runtimeContext.SessionID() != "session-authenticated" {
		t.Fatalf("SessionID() = %q", runtimeContext.SessionID())
	}
	if !runtimeContext.Authenticated() || runtimeContext.Anonymous() {
		t.Fatalf("identity flags = authenticated:%t anonymous:%t", runtimeContext.Authenticated(), runtimeContext.Anonymous())
	}
	if runtimeContext.AuthenticationType() != "jwt" || runtimeContext.AuthenticationProvider() != "internal-jwt" {
		t.Fatalf(
			"Authentication metadata = (%q, %q)",
			runtimeContext.AuthenticationType(),
			runtimeContext.AuthenticationProvider(),
		)
	}
}

func TestNewContextCreatesAnonymousContext(t *testing.T) {
	runtimeMessage := mustContextMessage(t, TypeBinary, []byte{0x00, 0x01, 0xff})
	sender := &contextSender{}

	runtimeContext, err := NewContext(
		&runtimeMessage,
		sender,
		"session-anonymous",
		false,
		true,
		"",
		"",
	)
	if err != nil {
		t.Fatalf("NewContext() error = %v", err)
	}

	assertContextMessage(t, runtimeContext, runtimeMessage)
	if runtimeContext.Sender() != sender {
		t.Fatal("Sender() did not return the supplied capability")
	}
	if runtimeContext.SessionID() != "session-anonymous" {
		t.Fatalf("SessionID() = %q", runtimeContext.SessionID())
	}
	if runtimeContext.Authenticated() || !runtimeContext.Anonymous() {
		t.Fatalf("identity flags = authenticated:%t anonymous:%t", runtimeContext.Authenticated(), runtimeContext.Anonymous())
	}
	if runtimeContext.AuthenticationType() != "" || runtimeContext.AuthenticationProvider() != "" {
		t.Fatalf(
			"anonymous Authentication metadata = (%q, %q)",
			runtimeContext.AuthenticationType(),
			runtimeContext.AuthenticationProvider(),
		)
	}
}

func TestNewContextRejectsInvalidArguments(t *testing.T) {
	validMessage := mustContextMessage(t, TypeText, []byte("payload"))
	invalidMessage := Message{messageType: Type("ping")}
	validSender := &contextSender{}
	var typedNilSender *contextSender

	tests := []struct {
		name                   string
		runtimeMessage         *Message
		sender                 Sender
		sessionID              string
		authenticated          bool
		anonymous              bool
		authenticationType     string
		authenticationProvider string
	}{
		{name: "nil Message", sender: validSender, sessionID: "session", authenticated: true, authenticationType: "jwt", authenticationProvider: "provider"},
		{name: "invalid Message", runtimeMessage: &invalidMessage, sender: validSender, sessionID: "session", authenticated: true, authenticationType: "jwt", authenticationProvider: "provider"},
		{name: "nil Sender", runtimeMessage: &validMessage, sessionID: "session", authenticated: true, authenticationType: "jwt", authenticationProvider: "provider"},
		{name: "typed nil Sender", runtimeMessage: &validMessage, sender: typedNilSender, sessionID: "session", authenticated: true, authenticationType: "jwt", authenticationProvider: "provider"},
		{name: "empty Session ID", runtimeMessage: &validMessage, sender: validSender, authenticated: true, authenticationType: "jwt", authenticationProvider: "provider"},
		{name: "both identity flags", runtimeMessage: &validMessage, sender: validSender, sessionID: "session", authenticated: true, anonymous: true, authenticationType: "jwt", authenticationProvider: "provider"},
		{name: "neither identity flag", runtimeMessage: &validMessage, sender: validSender, sessionID: "session"},
		{name: "authenticated without type", runtimeMessage: &validMessage, sender: validSender, sessionID: "session", authenticated: true, authenticationProvider: "provider"},
		{name: "authenticated without provider", runtimeMessage: &validMessage, sender: validSender, sessionID: "session", authenticated: true, authenticationType: "jwt"},
		{name: "anonymous with type", runtimeMessage: &validMessage, sender: validSender, sessionID: "session", anonymous: true, authenticationType: "jwt"},
		{name: "anonymous with provider", runtimeMessage: &validMessage, sender: validSender, sessionID: "session", anonymous: true, authenticationProvider: "provider"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := NewContext(
				test.runtimeMessage,
				test.sender,
				test.sessionID,
				test.authenticated,
				test.anonymous,
				test.authenticationType,
				test.authenticationProvider,
			)
			if !errors.Is(err, ErrInvalidContext) {
				t.Fatalf("NewContext() error = %v, want ErrInvalidContext", err)
			}
			if !reflect.DeepEqual(got, Context{}) {
				t.Fatalf("NewContext() Context = %#v, want zero value", got)
			}
		})
	}
}

func TestContextPayloadIsImmutable(t *testing.T) {
	input := []byte("original")
	runtimeMessage := mustContextMessage(t, TypeBinary, input)
	runtimeContext := mustAuthenticatedContext(t, &runtimeMessage)
	wantReceivedAt := runtimeMessage.ReceivedAt()

	input[0] = 'X'
	runtimeMessage.data[1] = 'X'
	returnedMessage := runtimeContext.Message()
	returnedMessage.data[2] = 'X'
	returnedData := runtimeContext.Message().Data()
	returnedData[3] = 'X'

	got := runtimeContext.Message()
	if string(got.Data()) != "original" {
		t.Fatalf("Context Message payload = %q, want original", got.Data())
	}
	if !got.ReceivedAt().Equal(wantReceivedAt) {
		t.Fatalf("Context Message ReceivedAt = %v, want %v", got.ReceivedAt(), wantReceivedAt)
	}
}

func TestContextCopiesScalarValues(t *testing.T) {
	runtimeMessage := mustContextMessage(t, TypeText, []byte("payload"))
	sessionID := "session-original"
	authenticationType := "api-key"
	authenticationProvider := "internal-api-key"

	runtimeContext, err := NewContext(
		&runtimeMessage,
		&contextSender{},
		sessionID,
		true,
		false,
		authenticationType,
		authenticationProvider,
	)
	if err != nil {
		t.Fatalf("NewContext() error = %v", err)
	}

	sessionID = "session-changed"
	authenticationType = "changed"
	authenticationProvider = "changed"
	if runtimeContext.SessionID() != "session-original" ||
		runtimeContext.AuthenticationType() != "api-key" ||
		runtimeContext.AuthenticationProvider() != "internal-api-key" {
		t.Fatalf(
			"Context scalars = (%q, %q, %q)",
			runtimeContext.SessionID(),
			runtimeContext.AuthenticationType(),
			runtimeContext.AuthenticationProvider(),
		)
	}
}

func TestContextContainsOnlyApprovedTransportNeutralFields(t *testing.T) {
	contextType := reflect.TypeOf(Context{})
	wantFieldTypes := []reflect.Type{
		reflect.TypeOf(Message{}),
		reflect.TypeOf((*Sender)(nil)).Elem(),
		reflect.TypeOf(""),
		reflect.TypeOf(false),
		reflect.TypeOf(false),
		reflect.TypeOf(""),
		reflect.TypeOf(""),
	}
	if contextType.NumField() != len(wantFieldTypes) {
		t.Fatalf("Context fields = %d, want %d", contextType.NumField(), len(wantFieldTypes))
	}
	for index, wantType := range wantFieldTypes {
		field := contextType.Field(index)
		if field.IsExported() {
			t.Fatalf("Context field %q is exported", field.Name)
		}
		if field.Type != wantType {
			t.Fatalf("Context field %q type = %v, want %v", field.Name, field.Type, wantType)
		}
	}

	wantMethods := map[string]struct{}{
		"Anonymous":              {},
		"Authenticated":          {},
		"AuthenticationProvider": {},
		"AuthenticationType":     {},
		"Message":                {},
		"Sender":                 {},
		"SessionID":              {},
	}
	if contextType.NumMethod() != len(wantMethods) {
		t.Fatalf("Context exported methods = %d, want %d", contextType.NumMethod(), len(wantMethods))
	}
	for index := 0; index < contextType.NumMethod(); index++ {
		method := contextType.Method(index)
		if _, exists := wantMethods[method.Name]; !exists {
			t.Fatalf("Context exposes unexpected method %q", method.Name)
		}
	}
}

func TestContextSupportsConcurrentReads(t *testing.T) {
	runtimeMessage := mustContextMessage(t, TypeBinary, []byte("concurrent payload"))
	runtimeContext := mustAuthenticatedContext(t, &runtimeMessage)

	const readers = 64
	start := make(chan struct{})
	errorsFound := make(chan error, readers)
	var waitGroup sync.WaitGroup
	waitGroup.Add(readers)
	for index := 0; index < readers; index++ {
		go func() {
			defer waitGroup.Done()
			<-start
			returnedMessage := runtimeContext.Message()
			payload := returnedMessage.Data()
			payload[0] = 'X'
			if returnedMessage.Type() != TypeBinary || string(runtimeContext.Message().Data()) != "concurrent payload" {
				errorsFound <- fmt.Errorf("Message accessor returned inconsistent data")
				return
			}
			if runtimeContext.Sender() == nil || runtimeContext.SessionID() != "session" ||
				!runtimeContext.Authenticated() || runtimeContext.Anonymous() ||
				runtimeContext.AuthenticationType() != "jwt" ||
				runtimeContext.AuthenticationProvider() != "provider" {
				errorsFound <- fmt.Errorf("scalar accessor returned inconsistent data")
			}
		}()
	}
	close(start)
	waitGroup.Wait()
	close(errorsFound)
	for err := range errorsFound {
		t.Error(err)
	}
}

func mustContextMessage(t *testing.T, messageType Type, payload []byte) Message {
	t.Helper()
	runtimeMessage, err := New(messageType, payload)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	return runtimeMessage
}

func mustAuthenticatedContext(t *testing.T, runtimeMessage *Message) Context {
	t.Helper()
	runtimeContext, err := NewContext(
		runtimeMessage,
		&contextSender{},
		"session",
		true,
		false,
		"jwt",
		"provider",
	)
	if err != nil {
		t.Fatalf("NewContext() error = %v", err)
	}
	return runtimeContext
}

func assertContextMessage(t *testing.T, runtimeContext Context, want Message) {
	t.Helper()
	got := runtimeContext.Message()
	if got.Type() != want.Type() || string(got.Data()) != string(want.Data()) || !got.ReceivedAt().Equal(want.ReceivedAt()) {
		t.Fatalf("Message() = %#v, want %#v", got, want)
	}
}

type contextSender struct{}

func (*contextSender) Send(context.Context, Message) error {
	return nil
}

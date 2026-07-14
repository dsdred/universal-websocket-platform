package session

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/coder/websocket"
	"github.com/dsdred/universal-websocket-platform/internal/authentication"
	"github.com/dsdred/universal-websocket-platform/internal/message"
)

var (
	// ErrNilConnection indicates that Session construction received no WebSocket connection.
	ErrNilConnection = errors.New("WebSocket connection is nil")
	// ErrInvalidPrincipal indicates that Session construction received an invalid Principal.
	ErrInvalidPrincipal = errors.New("invalid Session Principal")
	// ErrSessionAlreadyRunning indicates that a Session cannot transition to Running.
	ErrSessionAlreadyRunning = errors.New("Session already running or stopped")
	// ErrSessionNotRunning indicates that Run was called outside the Running state.
	ErrSessionNotRunning = errors.New("Session is not running")
	// ErrSessionReadLoopAlreadyRunning indicates that one Run call already owns reading.
	ErrSessionReadLoopAlreadyRunning = errors.New("Session read loop already running")
)

// Session owns one authenticated WebSocket connection and its minimal lifecycle.
type Session interface {
	ID() string
	Principal() authentication.Principal
	RemoteAddress() string
	CreatedAt() time.Time
	Running() bool
	Start(context.Context) error
	Run(context.Context) error
	Send(context.Context, message.Message) error
	Stop(context.Context) error
}

type lifecycleState uint8

const (
	stateCreated lifecycleState = iota
	stateRunning
	stateStopping
	stateStopped
)

type idGenerator func() (string, error)
type messageObserver func(message.Message)

// DefaultSession is the default minimal Runtime Session.
type DefaultSession struct {
	mu            sync.RWMutex
	writeMu       sync.Mutex
	id            string
	principal     authentication.Principal
	connection    *websocket.Conn
	remoteAddress string
	createdAt     time.Time
	state         lifecycleState
	readLoop      bool
	readLoopDone  chan struct{}
	stopDone      chan struct{}
	stopErr       error
	observe       messageObserver
	handler       message.Handler
}

// New creates a Session with a cryptographically random identifier.
func New(
	connection *websocket.Conn,
	principal authentication.Principal,
	remoteAddress string,
) (*DefaultSession, error) {
	return newWithDependencies(connection, principal, remoteAddress, generateID, nil, nil)
}

// NewWithHandler creates a Session with an explicitly injected Runtime Message Handler.
func NewWithHandler(
	connection *websocket.Conn,
	principal authentication.Principal,
	remoteAddress string,
	handler message.Handler,
) (*DefaultSession, error) {
	return newWithDependencies(connection, principal, remoteAddress, generateID, nil, handler)
}

func newWithIDGenerator(
	connection *websocket.Conn,
	principal authentication.Principal,
	remoteAddress string,
	generate idGenerator,
) (*DefaultSession, error) {
	return newWithDependencies(connection, principal, remoteAddress, generate, nil, nil)
}

func newWithObserver(
	connection *websocket.Conn,
	principal authentication.Principal,
	remoteAddress string,
	observe messageObserver,
) (*DefaultSession, error) {
	return newWithDependencies(connection, principal, remoteAddress, generateID, observe, nil)
}

func newWithDependencies(
	connection *websocket.Conn,
	principal authentication.Principal,
	remoteAddress string,
	generate idGenerator,
	observe messageObserver,
	handler message.Handler,
) (*DefaultSession, error) {
	if connection == nil {
		return nil, ErrNilConnection
	}
	if !principal.Authenticated || principal.Anonymous {
		return nil, ErrInvalidPrincipal
	}
	id, err := generate()
	if err != nil {
		return nil, fmt.Errorf("generate Session ID: %w", err)
	}

	return &DefaultSession{
		id:            id,
		principal:     clonePrincipal(principal),
		connection:    connection,
		remoteAddress: strings.TrimSpace(remoteAddress),
		createdAt:     time.Now().UTC(),
		state:         stateCreated,
		observe:       observe,
		handler:       handler,
	}, nil
}

// ID returns the immutable Session identifier.
func (session *DefaultSession) ID() string {
	return session.id
}

// Principal returns an independent copy of the authenticated Principal.
func (session *DefaultSession) Principal() authentication.Principal {
	return clonePrincipal(session.principal)
}

// RemoteAddress returns the normalized peer address retained by the Session.
func (session *DefaultSession) RemoteAddress() string {
	return session.remoteAddress
}

// CreatedAt returns the UTC Session creation time.
func (session *DefaultSession) CreatedAt() time.Time {
	return session.createdAt
}

// Running reports whether the Session is in the Running state.
func (session *DefaultSession) Running() bool {
	session.mu.RLock()
	defer session.mu.RUnlock()
	return session.state == stateRunning
}

// Start moves a newly created Session to Running without starting message loops.
func (session *DefaultSession) Start(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	session.mu.Lock()
	defer session.mu.Unlock()
	if session.state != stateCreated {
		return ErrSessionAlreadyRunning
	}
	session.state = stateRunning
	return nil
}

// Run blocks in the calling goroutine while reading application messages.
func (session *DefaultSession) Run(ctx context.Context) error {
	session.mu.Lock()
	if session.state != stateRunning {
		session.mu.Unlock()
		return ErrSessionNotRunning
	}
	if session.readLoop {
		session.mu.Unlock()
		return ErrSessionReadLoopAlreadyRunning
	}
	session.readLoop = true
	session.readLoopDone = make(chan struct{})
	done := session.readLoopDone
	connection := session.connection
	observe := session.observe
	session.mu.Unlock()

	defer func() {
		session.mu.Lock()
		session.readLoop = false
		close(done)
		session.mu.Unlock()
	}()

	for {
		websocketType, payload, err := connection.Read(ctx)
		if err != nil {
			return sessionReadError(ctx, err)
		}

		var messageType message.Type
		switch websocketType {
		case websocket.MessageText:
			messageType = message.TypeText
		case websocket.MessageBinary:
			messageType = message.TypeBinary
		default:
			continue
		}
		runtimeMessage, err := message.New(messageType, payload)
		if err != nil {
			return fmt.Errorf("create Runtime Message: %w", err)
		}
		if observe != nil {
			observe(runtimeMessage)
		}
		if session.handler != nil {
			if err := session.handler.Handle(ctx, session, runtimeMessage); err != nil {
				return fmt.Errorf("handle Runtime Message: %w", err)
			}
		}
	}
}

// Send writes one transport-neutral Runtime Message while the Session is Running.
func (session *DefaultSession) Send(ctx context.Context, runtimeMessage message.Message) error {
	var websocketType websocket.MessageType
	switch runtimeMessage.Type() {
	case message.TypeText:
		websocketType = websocket.MessageText
	case message.TypeBinary:
		websocketType = websocket.MessageBinary
	default:
		return message.ErrInvalidMessageType
	}
	payload := append([]byte(nil), runtimeMessage.Data()...)

	session.writeMu.Lock()
	defer session.writeMu.Unlock()

	session.mu.RLock()
	if session.state != stateRunning {
		session.mu.RUnlock()
		return ErrSessionNotRunning
	}
	connection := session.connection
	writeErr := connection.Write(ctx, websocketType, payload)
	session.mu.RUnlock()
	if writeErr != nil {
		return fmt.Errorf("write WebSocket message: %w", writeErr)
	}
	return nil
}

// Stop sends a normal closure without holding the lifecycle mutex and is idempotent.
func (session *DefaultSession) Stop(ctx context.Context) error {
	session.mu.Lock()
	switch session.state {
	case stateStopping:
		done := session.stopDone
		session.mu.Unlock()
		select {
		case <-done:
			session.mu.RLock()
			defer session.mu.RUnlock()
			return session.stopErr
		case <-ctx.Done():
			return ctx.Err()
		}
	case stateStopped:
		err := session.stopErr
		session.mu.Unlock()
		return err
	case stateCreated, stateRunning:
		session.state = stateStopping
		session.stopDone = make(chan struct{})
		done := session.stopDone
		readLoopDone := session.readLoopDone
		connection := session.connection
		session.mu.Unlock()

		closeErr := connection.Close(websocket.StatusNormalClosure, "")
		_ = connection.CloseNow()
		if errors.Is(closeErr, net.ErrClosed) {
			closeErr = nil
		}
		if readLoopDone != nil {
			<-readLoopDone
		}

		session.mu.Lock()
		session.state = stateStopped
		session.stopErr = closeErr
		close(done)
		session.mu.Unlock()
		return closeErr
	default:
		session.mu.Unlock()
		return nil
	}
}

func sessionReadError(ctx context.Context, err error) error {
	if contextErr := ctx.Err(); contextErr != nil {
		return contextErr
	}
	status := websocket.CloseStatus(err)
	if status == websocket.StatusNormalClosure || status == websocket.StatusGoingAway {
		return nil
	}
	return fmt.Errorf("read WebSocket message: %w", err)
}

func generateID() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func clonePrincipal(principal authentication.Principal) authentication.Principal {
	principal.Claims = cloneStrings(principal.Claims)
	principal.Roles = append([]string(nil), principal.Roles...)
	principal.Attributes = cloneStrings(principal.Attributes)
	principal.Metadata = cloneStrings(principal.Metadata)
	return principal
}

func cloneStrings(values map[string]string) map[string]string {
	if values == nil {
		return nil
	}
	result := make(map[string]string, len(values))
	for name, value := range values {
		result[name] = value
	}
	return result
}

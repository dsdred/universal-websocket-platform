package session

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/coder/websocket"
	"github.com/dsdred/universal-websocket-platform/internal/authentication"
)

var (
	// ErrNilConnection indicates that Session construction received no WebSocket connection.
	ErrNilConnection = errors.New("WebSocket connection is nil")
	// ErrInvalidPrincipal indicates that Session construction received an invalid Principal.
	ErrInvalidPrincipal = errors.New("invalid Session Principal")
	// ErrSessionAlreadyRunning indicates that a Session cannot transition to Running.
	ErrSessionAlreadyRunning = errors.New("Session already running or stopped")
)

// Session owns one authenticated WebSocket connection and its minimal lifecycle.
type Session interface {
	ID() string
	Principal() authentication.Principal
	RemoteAddress() string
	CreatedAt() time.Time
	Running() bool
	Start(context.Context) error
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

// DefaultSession is the default minimal Runtime Session.
type DefaultSession struct {
	mu            sync.RWMutex
	id            string
	principal     authentication.Principal
	connection    *websocket.Conn
	remoteAddress string
	createdAt     time.Time
	state         lifecycleState
	stopDone      chan struct{}
	stopErr       error
}

// New creates a Session with a cryptographically random identifier.
func New(
	connection *websocket.Conn,
	principal authentication.Principal,
	remoteAddress string,
) (*DefaultSession, error) {
	return newWithIDGenerator(connection, principal, remoteAddress, generateID)
}

func newWithIDGenerator(
	connection *websocket.Conn,
	principal authentication.Principal,
	remoteAddress string,
	generate idGenerator,
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
		connection := session.connection
		session.mu.Unlock()

		closeErr := connection.Close(websocket.StatusNormalClosure, "")
		connection.CloseNow()

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

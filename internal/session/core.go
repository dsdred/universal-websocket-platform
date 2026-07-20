package session

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/dsdred/universal-websocket-platform/internal/authentication"
	"github.com/dsdred/universal-websocket-platform/internal/message"
)

var (
	errNilIDGenerator            = errors.New("Session ID generator is nil")
	errInvalidGeneratedSessionID = errors.New("generated Session ID is empty")
	errNilSessionCore            = errors.New("Session Core is nil")
)

// sessionCore contains immutable Session identity and non-transport state.
// It owns no connection and exposes no lifecycle operations.
type sessionCore struct {
	id            string
	principal     authentication.Principal
	remoteAddress string
	createdAt     time.Time
	observe       messageObserver
	handler       message.Handler
}

func newSessionCore(
	principal authentication.Principal,
	remoteAddress string,
	generate idGenerator,
	observe messageObserver,
	handler message.Handler,
) (*sessionCore, error) {
	validAuthenticated := principal.Authenticated && !principal.Anonymous
	validAnonymous := principal.Anonymous && !principal.Authenticated && principal.ID == "anonymous"
	if !validAuthenticated && !validAnonymous {
		return nil, ErrInvalidPrincipal
	}
	if generate == nil {
		return nil, errNilIDGenerator
	}

	id, err := generate()
	if err != nil {
		return nil, fmt.Errorf("generate Session ID: %w", err)
	}
	if id == "" {
		return nil, errInvalidGeneratedSessionID
	}

	return &sessionCore{
		id:            id,
		principal:     clonePrincipal(principal),
		remoteAddress: strings.TrimSpace(remoteAddress),
		createdAt:     time.Now().UTC(),
		observe:       observe,
		handler:       handler,
	}, nil
}

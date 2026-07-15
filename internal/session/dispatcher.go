package session

import (
	"context"

	"github.com/coder/websocket"

	"github.com/dsdred/universal-websocket-platform/internal/authentication"
	"github.com/dsdred/universal-websocket-platform/internal/connection"
	"github.com/dsdred/universal-websocket-platform/internal/message"
)

type sessionFactory interface {
	Create(*websocket.Conn, authentication.Principal, string) (Session, error)
}

type handlerSessionFactory struct {
	handler message.Handler
}

func (factory handlerSessionFactory) Create(
	websocketConnection *websocket.Conn,
	principal authentication.Principal,
	remoteAddress string,
) (Session, error) {
	return NewWithHandler(websocketConnection, principal, remoteAddress, factory.handler)
}

// Dispatcher creates and runs a minimal Session for each authenticated connection.
type Dispatcher struct {
	factory sessionFactory
}

// NewDispatcher creates the production Session Dispatcher.
func NewDispatcher(handler message.Handler) *Dispatcher {
	return newDispatcher(handlerSessionFactory{handler: handler})
}

func newDispatcher(factory sessionFactory) *Dispatcher {
	return &Dispatcher{factory: factory}
}

// DispatchAuthenticated transfers connection ownership to a new Session.
func (dispatcher *Dispatcher) DispatchAuthenticated(authenticatedContext connection.AuthenticatedContext) (bool, error) {
	connectionContext := authenticatedContext.ConnectionContext()
	ctx := connectionContext.Context()
	if err := ctx.Err(); err != nil {
		return false, err
	}

	runtimeSession, err := dispatcher.factory.Create(
		connectionContext.Connection(),
		authenticatedContext.Principal(),
		connectionContext.Request().RemoteAddr,
	)
	if err != nil {
		return false, err
	}
	if err := runtimeSession.Start(ctx); err != nil {
		return false, err
	}
	defer connectionContext.Cancel()
	runErr := runtimeSession.Run(ctx)
	stopErr := runtimeSession.Stop(context.Background())
	if runErr != nil {
		return true, runErr
	}
	return true, stopErr
}

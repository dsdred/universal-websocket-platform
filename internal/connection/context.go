package connection

import (
	"context"
	"net/http"

	"github.com/coder/websocket"
)

// ConnectionContext contains transport references and the derived lifecycle for one upgraded connection.
type ConnectionContext struct {
	ctx        context.Context
	cancel     context.CancelFunc
	connection *websocket.Conn
	request    *http.Request
}

// NewContext creates a transport-only ConnectionContext.
func NewContext(
	ctx context.Context,
	websocketConnection *websocket.Conn,
	request *http.Request,
) ConnectionContext {
	return ConnectionContext{
		ctx:        ctx,
		connection: websocketConnection,
		request:    request,
	}
}

// NewRuntimeContext creates a connection context derived from the Host-owned Runtime context.
func NewRuntimeContext(
	runtimeContext context.Context,
	websocketConnection *websocket.Conn,
	request *http.Request,
) ConnectionContext {
	connectionContext, cancel := context.WithCancel(runtimeContext)
	return ConnectionContext{
		ctx:        connectionContext,
		cancel:     cancel,
		connection: websocketConnection,
		request:    request,
	}
}

// Context returns the lifecycle Context associated with the connection.
func (connectionContext ConnectionContext) Context() context.Context {
	return connectionContext.ctx
}

// Cancel ends the derived connection lifecycle when its current owner is finished.
func (connectionContext ConnectionContext) Cancel() {
	if connectionContext.cancel != nil {
		connectionContext.cancel()
	}
}

// Connection returns the upgraded WebSocket connection.
func (connectionContext ConnectionContext) Connection() *websocket.Conn {
	return connectionContext.connection
}

// Request returns the HTTP request that initiated the Upgrade.
func (connectionContext ConnectionContext) Request() *http.Request {
	return connectionContext.request
}

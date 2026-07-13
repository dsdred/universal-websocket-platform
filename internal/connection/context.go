package connection

import (
	"context"
	"net/http"

	"github.com/coder/websocket"
)

// ConnectionContext contains immutable transport references for one upgraded connection.
type ConnectionContext struct {
	ctx        context.Context
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

// Context returns the lifecycle Context associated with the connection.
func (connectionContext ConnectionContext) Context() context.Context {
	return connectionContext.ctx
}

// Connection returns the upgraded WebSocket connection.
func (connectionContext ConnectionContext) Connection() *websocket.Conn {
	return connectionContext.connection
}

// Request returns the HTTP request that initiated the Upgrade.
func (connectionContext ConnectionContext) Request() *http.Request {
	return connectionContext.request
}

package message

import "context"

// EchoHandler sends every incoming Runtime Message back through its Session capability.
type EchoHandler struct{}

// NewEchoHandler creates a stateless Echo Runtime Message Handler.
func NewEchoHandler() EchoHandler {
	return EchoHandler{}
}

// Handle returns the original immutable Message through Sender without transport access.
func (EchoHandler) Handle(ctx context.Context, runtimeContext Context) error {
	return runtimeContext.Sender().Send(ctx, runtimeContext.Message())
}

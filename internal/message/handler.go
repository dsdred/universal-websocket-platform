package message

import "context"

// Sender is the minimal Session capability available to Runtime Message Handlers.
// It keeps this package independent from transport and the concrete Session package.
type Sender interface {
	Send(context.Context, Message) error
}

// Handler processes one transport-neutral Runtime Message.
type Handler interface {
	Handle(context.Context, Sender, Message) error
}

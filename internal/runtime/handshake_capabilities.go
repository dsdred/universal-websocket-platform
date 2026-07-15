package runtime

import "context"

// handshakeCapabilities is a stable read-only bridge over Host-owned lifecycle state.
type handshakeCapabilities struct {
	canAccept      func() bool
	runtimeContext func() context.Context
}

func (capabilities *handshakeCapabilities) CanAccept() bool {
	return capabilities.canAccept()
}

func (capabilities *handshakeCapabilities) RuntimeContext() context.Context {
	return capabilities.runtimeContext()
}

package listener

import (
	"strings"

	"github.com/dsdred/universal-websocket-platform/internal/connection"
	"github.com/dsdred/universal-websocket-platform/internal/runtimeconfig"
)

// Bootstrap builds an effective Runtime Listener from an immutable Snapshot.
type Bootstrap interface {
	Build(runtimeconfig.ListenerSnapshot) (Listener, error)
}

// DefaultBootstrap builds the default Runtime Listener with an injectable Dispatcher.
type DefaultBootstrap struct {
	dispatcher connection.Dispatcher
}

// NewBootstrap creates a Listener Bootstrap with the supplied Connection Dispatcher.
func NewBootstrap(dispatcher connection.Dispatcher) DefaultBootstrap {
	return DefaultBootstrap{dispatcher: dispatcher}
}

// Build validates and copies Listener Snapshot metadata without opening a socket.
func (bootstrap DefaultBootstrap) Build(snapshot runtimeconfig.ListenerSnapshot) (Listener, error) {
	host := strings.TrimSpace(snapshot.Host)
	certificateRef := strings.TrimSpace(snapshot.TLS.CertificateRef)
	privateKeyRef := strings.TrimSpace(snapshot.TLS.PrivateKeyRef)
	minVersion := strings.TrimSpace(snapshot.TLS.MinVersion)
	if host == "" || snapshot.Port == 0 {
		return nil, ErrInvalidListenerConfiguration
	}
	if minVersion != "1.2" && minVersion != "1.3" {
		return nil, ErrInvalidListenerConfiguration
	}
	if snapshot.TLS.Enabled && (certificateRef == "" || privateKeyRef == "") {
		return nil, ErrInvalidListenerConfiguration
	}

	dispatcher := bootstrap.dispatcher
	if dispatcher == nil {
		dispatcher = connection.DefaultDispatcher{}
	}

	return &DefaultListener{
		host: host,
		port: snapshot.Port,
		tls: tlsConfiguration{
			enabled:        snapshot.TLS.Enabled,
			certificateRef: certificateRef,
			privateKeyRef:  privateKeyRef,
			minVersion:     minVersion,
		},
		state:      listenerCreated,
		dispatcher: dispatcher,
	}, nil
}

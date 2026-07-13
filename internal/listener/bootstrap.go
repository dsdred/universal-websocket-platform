package listener

import (
	"strings"

	"github.com/dsdred/universal-websocket-platform/internal/runtimeconfig"
)

// Bootstrap builds an effective Runtime Listener from an immutable Snapshot.
type Bootstrap interface {
	Build(runtimeconfig.ListenerSnapshot) (Listener, error)
}

// DefaultBootstrap builds the default metadata-only Runtime Listener.
type DefaultBootstrap struct{}

// Build validates and copies Listener Snapshot metadata without opening a socket.
func (DefaultBootstrap) Build(snapshot runtimeconfig.ListenerSnapshot) (Listener, error) {
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

	return &DefaultListener{
		host: host,
		port: snapshot.Port,
		tls: tlsConfiguration{
			enabled:        snapshot.TLS.Enabled,
			certificateRef: certificateRef,
			privateKeyRef:  privateKeyRef,
			minVersion:     minVersion,
		},
		state: listenerCreated,
	}, nil
}

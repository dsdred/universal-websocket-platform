package listener

import (
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/dsdred/universal-websocket-platform/internal/runtimeconfig"
	"github.com/dsdred/universal-websocket-platform/internal/runtimediagnostics"
)

// Bootstrap builds an effective Runtime Listener from an immutable Snapshot.
type Bootstrap interface {
	Build(runtimeconfig.ListenerSnapshot) (Listener, error)
}

// DefaultBootstrap builds the default Runtime Listener with an injected Handshake boundary.
type DefaultBootstrap struct {
	handshakeHandler http.Handler
	reportError      func(error)
	preparedTLS      *PreparedTLS
	diagnostics      runtimediagnostics.Producer
}

// NewBootstrapWithTLSAndHandshake binds already prepared, opaque TLS material.
// Secret resolution belongs to Host startup composition, never Build.
func NewBootstrapWithTLSAndHandshake(handler http.Handler, reporter func(error), prepared *PreparedTLS) DefaultBootstrap {
	return DefaultBootstrap{handshakeHandler: handler, reportError: reporter, preparedTLS: prepared}
}

// NewBootstrapWithHandshake creates a Listener Bootstrap with a pre-Upgrade Handshake handler.
func NewBootstrapWithHandshake(handler http.Handler) DefaultBootstrap {
	return DefaultBootstrap{handshakeHandler: handler}
}

// NewBootstrapWithHandshakeAndTerminalErrorReporter creates a Listener Bootstrap with explicit error reporting.
// The callback is synchronous and must return promptly; a callback panic is isolated from Listener lifecycle.
func NewBootstrapWithHandshakeAndTerminalErrorReporter(
	handler http.Handler,
	reportError func(error),
) DefaultBootstrap {
	return DefaultBootstrap{handshakeHandler: handler, reportError: reportError}
}

// WithDiagnosticsProducer returns a Bootstrap copy with an optional mailbox capability.
// The value is captured before Build; zero disables diagnostics.
func (bootstrap DefaultBootstrap) WithDiagnosticsProducer(producer runtimediagnostics.Producer) DefaultBootstrap {
	bootstrap.diagnostics = producer
	return bootstrap
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
	if snapshot.TLS.Enabled && (bootstrap.preparedTLS == nil || bootstrap.preparedTLS.config == nil) {
		return nil, ErrInvalidListenerConfiguration
	}
	if snapshot.TLS.Enabled && bootstrap.preparedTLS.settings != snapshot.TLS {
		return nil, ErrInvalidListenerConfiguration
	}
	if snapshot.TLS.Enabled && (snapshot.Timeouts.HandshakeSeconds < 1 || snapshot.Timeouts.HandshakeSeconds > 300) {
		return nil, ErrInvalidListenerConfiguration
	}
	if bootstrap.handshakeHandler == nil {
		return nil, ErrInvalidListenerConfiguration
	}

	created := &DefaultListener{
		host: host,
		port: snapshot.Port,
		tls: tlsConfiguration{
			enabled:        snapshot.TLS.Enabled,
			certificateRef: certificateRef,
			privateKeyRef:  privateKeyRef,
			minVersion:     minVersion,
		},
		state:            listenerCreated,
		handshakeHandler: bootstrap.handshakeHandler,
		reportError:      bootstrap.reportError,
		diagnostics:      bootstrap.diagnostics,
		serveHTTP: func(server *http.Server, listener net.Listener) error {
			return server.Serve(listener)
		},
	}
	if snapshot.TLS.Enabled {
		created.tlsConfig = bootstrap.preparedTLS.config.Clone()
		created.handshakeTimeout = time.Duration(snapshot.Timeouts.HandshakeSeconds) * time.Second
	}
	return created, nil
}

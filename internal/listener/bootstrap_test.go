package listener

import (
	"errors"
	"net/http"
	"testing"

	"github.com/dsdred/universal-websocket-platform/internal/runtimeconfig"
)

func TestDefaultBootstrapImplementsBootstrap(t *testing.T) {
	var _ Bootstrap = DefaultBootstrap{}
}

func TestBootstrapBuild(t *testing.T) {
	created, err := NewBootstrapWithHandshake(http.NotFoundHandler()).Build(validListenerSnapshot())
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	listener, ok := created.(*DefaultListener)
	if !ok {
		t.Fatalf("Build() Listener type = %T, want *DefaultListener", created)
	}
	if listener.Address() != "127.0.0.1:9443" || listener.Running() {
		t.Fatalf("Listener = (Address %q, Running %t)", listener.Address(), listener.Running())
	}
	wantTLS := tlsConfiguration{
		enabled:        true,
		certificateRef: "certificates/listener",
		privateKeyRef:  "secrets/listener-key",
		minVersion:     "1.3",
	}
	if listener.tls != wantTLS {
		t.Fatalf("TLS configuration = %+v, want %+v", listener.tls, wantTLS)
	}
	var _ Listener = listener
}

func TestBootstrapBuildRejectsMissingHandshake(t *testing.T) {
	created, err := (DefaultBootstrap{}).Build(validListenerSnapshot())
	if created != nil || !errors.Is(err, ErrInvalidListenerConfiguration) {
		t.Fatalf("Build() = (%v, %v), want nil and ErrInvalidListenerConfiguration", created, err)
	}
}

func TestBootstrapBuildDeepCopiesSnapshot(t *testing.T) {
	snapshot := validListenerSnapshot()
	created, err := NewBootstrapWithHandshake(http.NotFoundHandler()).Build(snapshot)
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	listener := created.(*DefaultListener)

	snapshot.Host = "changed"
	snapshot.Port = 8080
	snapshot.TLS.Enabled = false
	snapshot.TLS.CertificateRef = "changed-certificate"
	snapshot.TLS.PrivateKeyRef = "changed-key"
	snapshot.TLS.MinVersion = "1.2"

	if listener.Address() != "127.0.0.1:9443" {
		t.Fatalf("Address() = %q, want 127.0.0.1:9443", listener.Address())
	}
	wantTLS := tlsConfiguration{
		enabled:        true,
		certificateRef: "certificates/listener",
		privateKeyRef:  "secrets/listener-key",
		minVersion:     "1.3",
	}
	if listener.tls != wantTLS {
		t.Fatalf("TLS configuration changed with Snapshot: %+v", listener.tls)
	}
}

func TestBootstrapBuildRejectsInvalidSnapshot(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*runtimeconfig.ListenerSnapshot)
	}{
		{name: "empty host", mutate: func(snapshot *runtimeconfig.ListenerSnapshot) { snapshot.Host = " " }},
		{name: "zero port", mutate: func(snapshot *runtimeconfig.ListenerSnapshot) { snapshot.Port = 0 }},
		{name: "empty minimum TLS version", mutate: func(snapshot *runtimeconfig.ListenerSnapshot) { snapshot.TLS.MinVersion = "" }},
		{name: "unsupported TLS version", mutate: func(snapshot *runtimeconfig.ListenerSnapshot) { snapshot.TLS.MinVersion = "1.1" }},
		{name: "missing certificate", mutate: func(snapshot *runtimeconfig.ListenerSnapshot) { snapshot.TLS.CertificateRef = "" }},
		{name: "missing private key", mutate: func(snapshot *runtimeconfig.ListenerSnapshot) { snapshot.TLS.PrivateKeyRef = "" }},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			snapshot := validListenerSnapshot()
			test.mutate(&snapshot)
			created, err := NewBootstrapWithHandshake(http.NotFoundHandler()).Build(snapshot)
			if created != nil || !errors.Is(err, ErrInvalidListenerConfiguration) {
				t.Fatalf("Build() = (%v, %v), want nil and ErrInvalidListenerConfiguration", created, err)
			}
		})
	}
}

func validListenerSnapshot() runtimeconfig.ListenerSnapshot {
	return runtimeconfig.ListenerSnapshot{
		Host: "127.0.0.1",
		Port: 9443,
		TLS: runtimeconfig.TLSSnapshot{
			Enabled:        true,
			CertificateRef: "certificates/listener",
			PrivateKeyRef:  "secrets/listener-key",
			MinVersion:     "1.3",
		},
	}
}

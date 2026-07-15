package runtime

import (
	"context"
	"net/http"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/coder/websocket"
	"github.com/dsdred/universal-websocket-platform/internal/authentication"
	"github.com/dsdred/universal-websocket-platform/internal/connection"
	"github.com/dsdred/universal-websocket-platform/internal/handshake"
	"github.com/dsdred/universal-websocket-platform/internal/listener"
	"github.com/dsdred/universal-websocket-platform/internal/message"
	"github.com/dsdred/universal-websocket-platform/internal/runtimeconfig"
	"github.com/dsdred/universal-websocket-platform/internal/secretresolver"
)

func TestHostShutdownDuringAuthenticationPreventsUpgrade(t *testing.T) {
	entered := make(chan struct{})
	release := make(chan struct{})
	service := &blockingAuthenticationService{entered: entered, release: release}
	handoff := &countingAuthenticatedDispatcher{}
	snapshot := validSnapshot()
	snapshot.Listener.Port = availablePort(t)
	snapshot.Listener.TLS.Enabled = false

	var host *DefaultHost
	composer := func(
		snapshot runtimeconfig.Snapshot,
		_ secretresolver.Resolver,
		_ message.Handler,
	) (listener.Listener, error) {
		handler, err := handshake.NewHandler(host.capabilities, host.capabilities, service, handoff)
		if err != nil {
			return nil, err
		}
		return listener.NewBootstrapWithHandshake(handler).Build(snapshot.Listener)
	}
	created, err := newHost(snapshot, emptyResolver(t), nil, composer)
	if err != nil {
		t.Fatalf("newHost() error = %v", err)
	}
	host = created
	if err := host.Build(); err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	if err := host.Start(context.Background()); err != nil {
		t.Fatalf("Start() error = %v", err)
	}

	dialResult := make(chan handshakeDialResult, 1)
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		connection, response, dialErr := websocket.Dial(
			ctx,
			"ws://127.0.0.1:"+portString(snapshot.Listener.Port)+"/ws",
			nil,
		)
		if connection != nil {
			connection.CloseNow()
		}
		dialResult <- handshakeDialResult{response: response, err: dialErr}
	}()
	waitForSignal(t, entered, "Authentication entry")

	stopping := make(chan struct{})
	var stoppingOnce sync.Once
	host.stateObserver = func(state hostState) {
		if state == hostStopping {
			stoppingOnce.Do(func() { close(stopping) })
		}
	}
	stopResult := make(chan error, 1)
	go func() { stopResult <- host.Stop(context.Background()) }()
	waitForSignal(t, stopping, "Host shutdown admission closure")
	if host.CanAccept() {
		t.Fatal("Host still accepts admission after shutdown began")
	}
	close(release)

	select {
	case result := <-dialResult:
		if result.err == nil || result.response == nil {
			t.Fatalf("websocket.Dial() = (%v, %v), want HTTP rejection", result.response, result.err)
		}
		defer result.response.Body.Close()
		if result.response.StatusCode != http.StatusServiceUnavailable {
			t.Fatalf("status = %d, want %d", result.response.StatusCode, http.StatusServiceUnavailable)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("Handshake did not finish after Authentication release")
	}
	waitForResult(t, stopResult, "Host Stop after in-flight Authentication")
	if handoff.calls.Load() != 0 {
		t.Fatalf("Session handoff calls = %d, want 0", handoff.calls.Load())
	}
	if service.calls.Load() != 1 {
		t.Fatalf("Authentication calls = %d, want 1", service.calls.Load())
	}
}

type blockingAuthenticationService struct {
	entered chan<- struct{}
	release <-chan struct{}
	calls   atomic.Int32
}

func (service *blockingAuthenticationService) Authenticate(
	context.Context,
	authentication.AuthenticationRequest,
) (authentication.AuthenticationResult, error) {
	service.calls.Add(1)
	service.entered <- struct{}{}
	<-service.release
	return authentication.AuthenticationResult{
		Success: true,
		Principal: &authentication.Principal{
			ID:            "authenticated-client",
			Authenticated: true,
		},
	}, nil
}

type countingAuthenticatedDispatcher struct {
	calls atomic.Int32
}

func (dispatcher *countingAuthenticatedDispatcher) DispatchAuthenticated(
	connection.AuthenticatedContext,
) (bool, error) {
	dispatcher.calls.Add(1)
	return true, nil
}

type handshakeDialResult struct {
	response *http.Response
	err      error
}

package runtime

import (
	"context"
	"testing"
	"time"

	"github.com/coder/websocket"
	"github.com/dsdred/universal-websocket-platform/internal/configurationversion"
	"github.com/dsdred/universal-websocket-platform/internal/executionowner"
	"github.com/dsdred/universal-websocket-platform/internal/message"
	"github.com/dsdred/universal-websocket-platform/internal/session"
	"github.com/dsdred/universal-websocket-platform/internal/sessionmanager"
)

func TestRuntimeReadTimeoutFromPublishedSnapshot(t *testing.T) {
	for _, seconds := range []uint32{0, 1, 86400} {
		t.Run((time.Duration(seconds) * time.Second).String(), func(t *testing.T) {
			snapshot, diagnostics := buildSnapshotForTest(func(version *configurationversion.ConfigurationVersion) {
				version.Listener.Timeouts.ReadSeconds = seconds
			})
			if len(diagnostics) != 0 {
				t.Fatal(diagnostics)
			}
			copy := snapshot.Listener()
			copy.Timeouts.ReadSeconds = 99
			calls := 0
			_, err := composeRuntimeWithFactories(snapshot, emptyResolver(t), nil,
				&handshakeCapabilities{canAccept: func() bool { return true }, runtimeContext: context.Background},
				nil, buildMessageRouter,
				func(manager *sessionmanager.Manager, root session.RuntimeContextProvider, handler message.Handler, observer executionowner.TerminalObserver, readTimeout, writeTimeout time.Duration) (*session.TransactionalDispatcher, error) {
					calls++
					if readTimeout != time.Duration(seconds)*time.Second {
						t.Fatalf("composed read timeout %v, want %ds", readTimeout, seconds)
					}
					if writeTimeout != time.Duration(snapshot.Listener().Timeouts.WriteSeconds)*time.Second {
						t.Fatalf("composition changed write timeout: %v", writeTimeout)
					}
					return session.NewTransactionalDispatcherWithTimeouts(manager, root, handler, observer, readTimeout, writeTimeout)
				})
			if err != nil || calls != 1 {
				t.Fatalf("composition calls %d, error %v", calls, err)
			}
		})
	}
}

func TestBootstrapReadTimeoutClosesSessionAndPreservesRuntime(t *testing.T) {
	snapshot, diagnostics := buildSnapshotForTest(func(version *configurationversion.ConfigurationVersion) {
		version.Listener.Port = availablePort(t)
		version.Listener.Timeouts.ReadSeconds = 1
	})
	if len(diagnostics) != 0 {
		t.Fatal(diagnostics)
	}
	reported := make(chan error, 2)
	outcome := Bootstrap(&BootstrapRequest{
		Snapshot: snapshot, StartupContext: context.Background(),
		Dependencies: &DependencyBindings{
			SecretResolver: emptyResolver(t), LegacyMessageHandler: message.NewEchoHandler(),
			TerminalErrorReporter: func(err error) { reported <- err },
		},
	})
	host, ok := outcome.Success()
	if !ok {
		t.Fatalf("Bootstrap did not succeed: %#v", outcome)
	}
	t.Cleanup(func() { _ = host.Stop(context.Background()) })
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	connection, _, err := websocket.Dial(ctx, "ws://127.0.0.1:"+portString(snapshot.Listener().Port)+"/ws", nil)
	if err != nil {
		t.Fatal(err)
	}
	defer connection.CloseNow()
	for _, kind := range []websocket.MessageType{websocket.MessageText, websocket.MessageBinary} {
		if err := connection.Write(ctx, kind, []byte("read policy echo")); err != nil {
			t.Fatal(err)
		}
		gotKind, payload, err := connection.Read(ctx)
		if err != nil || gotKind != kind || string(payload) != "read policy echo" {
			t.Fatalf("echo: kind %v, payload %q, error %v", gotKind, payload, err)
		}
	}
	if _, _, err := connection.Read(ctx); err == nil || ctx.Err() != nil {
		t.Fatalf("expected server read timeout before caller deadline: %v, caller %v", err, ctx.Err())
	}
	if !host.Ready() || !host.CanAccept() || host.RuntimeContext().Err() != nil {
		t.Fatal("Session timeout changed Runtime readiness or root lifetime")
	}
	// Stop joins existing Manager/Owner accounting after the failed read.
	if err := host.Stop(ctx); err != nil {
		t.Fatalf("Runtime Stop after read timeout: %v", err)
	}
	select {
	case err := <-reported:
		t.Fatalf("post-Commit Session failure escaped Terminal Observer: %v", err)
	default:
	}
}

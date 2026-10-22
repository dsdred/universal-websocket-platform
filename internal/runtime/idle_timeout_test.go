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

func TestRuntimeIdleTimeoutFromPublishedSnapshot(t *testing.T) {
	for _, seconds := range []uint32{0, 1, 86400} {
		t.Run((time.Duration(seconds) * time.Second).String(), func(t *testing.T) {
			snapshot, diagnostics := buildSnapshotForTest(func(version *configurationversion.ConfigurationVersion) {
				version.Listener.Timeouts.IdleSeconds = seconds
			})
			if len(diagnostics) != 0 {
				t.Fatal(diagnostics)
			}
			copy := snapshot.Listener()
			copy.Timeouts.IdleSeconds = 99
			calls := 0
			_, err := composeRuntimeWithFactories(snapshot, emptyResolver(t), nil, &handshakeCapabilities{canAccept: func() bool { return true }, runtimeContext: context.Background}, nil, buildMessageRouter,
				func(manager *sessionmanager.Manager, root session.RuntimeContextProvider, handler message.Handler, observer executionowner.TerminalObserver, readTimeout, writeTimeout, idleTimeout time.Duration) (*session.TransactionalDispatcher, error) {
					calls++
					if idleTimeout != time.Duration(seconds)*time.Second {
						t.Fatalf("idle=%v want %ds", idleTimeout, seconds)
					}
					return session.NewTransactionalDispatcherWithIdleTimeout(manager, root, handler, observer, readTimeout, writeTimeout, idleTimeout)
				})
			if err != nil || calls != 1 {
				t.Fatalf("composition calls=%d error=%v", calls, err)
			}
		})
	}
}

func TestBootstrapIdleTimeoutEchoAndRuntimeSurvival(t *testing.T) {
	snapshot, diagnostics := buildSnapshotForTest(func(version *configurationversion.ConfigurationVersion) {
		version.Listener.Port = availablePort(t)
		version.Listener.Timeouts.ReadSeconds = 0
		version.Listener.Timeouts.IdleSeconds = 1
	})
	if len(diagnostics) != 0 {
		t.Fatal(diagnostics)
	}
	reported := make(chan error, 2)
	outcome := Bootstrap(&BootstrapRequest{Snapshot: snapshot, StartupContext: context.Background(), Dependencies: &DependencyBindings{SecretResolver: emptyResolver(t), LegacyMessageHandler: message.NewEchoHandler(), TerminalErrorReporter: func(err error) { reported <- err }}})
	host, ok := outcome.Success()
	if !ok {
		t.Fatalf("Bootstrap=%#v", outcome)
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
		if err := connection.Write(ctx, kind, []byte("idle echo")); err != nil {
			t.Fatal(err)
		}
		got, payload, err := connection.Read(ctx)
		if err != nil || got != kind || string(payload) != "idle echo" {
			t.Fatalf("echo kind=%v payload=%q err=%v", got, payload, err)
		}
	}
	if _, _, err := connection.Read(ctx); err == nil || ctx.Err() != nil {
		t.Fatalf("idle did not close before caller: %v caller=%v", err, ctx.Err())
	}
	if !host.Ready() || !host.CanAccept() || host.RuntimeContext().Err() != nil {
		t.Fatal("idle changed Runtime root/readiness")
	}
	if err := host.Stop(ctx); err != nil {
		t.Fatalf("Host Stop failed Manager/Owner join: %v", err)
	}
	select {
	case err := <-reported:
		t.Fatalf("post-Commit idle escaped Observer: %v", err)
	default:
	}
}

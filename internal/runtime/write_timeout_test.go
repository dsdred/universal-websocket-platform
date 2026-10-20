package runtime

import (
	"context"
	"testing"
	"time"

	"github.com/dsdred/universal-websocket-platform/internal/configurationversion"
	"github.com/dsdred/universal-websocket-platform/internal/executionowner"
	"github.com/dsdred/universal-websocket-platform/internal/message"
	"github.com/dsdred/universal-websocket-platform/internal/session"
	"github.com/dsdred/universal-websocket-platform/internal/sessionmanager"
)

func TestRuntimeWriteTimeoutFromPublishedSnapshot(t *testing.T) {
	for _, seconds := range []uint32{1, 300} {
		t.Run((time.Duration(seconds) * time.Second).String(), func(t *testing.T) {
			snapshot, diagnostics := buildSnapshotForTest(func(version *configurationversion.ConfigurationVersion) {
				version.Listener.Timeouts.WriteSeconds = seconds
			})
			if len(diagnostics) != 0 {
				t.Fatal(diagnostics)
			}
			// A consumer cannot mutate the detached Snapshot policy.
			copy := snapshot.Listener()
			copy.Timeouts.WriteSeconds = 99
			calls := 0
			_, err := composeRuntimeWithFactories(snapshot, emptyResolver(t), nil,
				&handshakeCapabilities{canAccept: func() bool { return true }, runtimeContext: context.Background},
				nil, buildMessageRouter,
				func(manager *sessionmanager.Manager, root session.RuntimeContextProvider, handler message.Handler, observer executionowner.TerminalObserver, timeout time.Duration) (*session.TransactionalDispatcher, error) {
					calls++
					if timeout != time.Duration(seconds)*time.Second {
						t.Fatalf("composed write timeout %v, want %ds", timeout, seconds)
					}
					return session.NewTransactionalDispatcherWithWriteTimeout(manager, root, handler, observer, timeout)
				})
			if err != nil || calls != 1 {
				t.Fatalf("composition calls %d, error %v", calls, err)
			}
		})
	}
}

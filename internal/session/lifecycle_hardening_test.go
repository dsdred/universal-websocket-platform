package session

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/coder/websocket"

	"github.com/dsdred/universal-websocket-platform/internal/message"
)

func TestSessionBlockedWriteDoesNotBlockStopLifecycle(t *testing.T) {
	connection := newBlockingSessionConnection()
	runtimeSession, err := newWithConnectionDependencies(
		connection,
		validPrincipal(),
		"",
		generateID,
		nil,
		nil,
	)
	if err != nil {
		t.Fatalf("create Session: %v", err)
	}
	if err := runtimeSession.Start(context.Background()); err != nil {
		t.Fatalf("Start() error = %v", err)
	}

	connection.cleanup(t)
	runtimeMessage, err := message.New(message.TypeText, []byte("admitted"))
	if err != nil {
		t.Fatalf("message.New() error = %v", err)
	}
	lateMessage, err := message.New(message.TypeText, []byte("too late"))
	if err != nil {
		t.Fatalf("message.New() error = %v", err)
	}

	sendResult := make(chan error, 1)
	go func() { sendResult <- runtimeSession.Send(context.Background(), runtimeMessage) }()
	waitSessionSignal(t, connection.writeEntered, "Write entry")

	stopResult := make(chan error, 1)
	go func() { stopResult <- runtimeSession.Stop(context.Background()) }()
	waitSessionSignal(t, connection.closeEntered, "Close entry while Write is blocked")

	runtimeSession.mu.RLock()
	state := runtimeSession.state
	runtimeSession.mu.RUnlock()
	if state != stateStopping {
		t.Fatalf("state = %d, want stateStopping", state)
	}

	lateSendResult := make(chan error, 1)
	go func() { lateSendResult <- runtimeSession.Send(context.Background(), lateMessage) }()
	if got := connection.writeCalls.Load(); got != 1 {
		t.Fatalf("Write calls before release = %d, want 1", got)
	}

	connection.releaseCloseOnce.Do(func() { close(connection.releaseClose) })
	if err := waitSessionResult(t, stopResult, "Stop"); err != nil {
		t.Fatalf("Stop() error = %v", err)
	}
	select {
	case err := <-sendResult:
		t.Fatalf("admitted Send() returned before Write release: %v", err)
	default:
	}

	connection.releaseWriteOnce.Do(func() { close(connection.releaseWrite) })
	if err := waitSessionResult(t, sendResult, "admitted Send"); err != nil {
		t.Fatalf("admitted Send() error = %v", err)
	}
	if err := waitSessionResult(t, lateSendResult, "late Send"); !errors.Is(err, ErrSessionNotRunning) {
		t.Fatalf("late Send() error = %v, want ErrSessionNotRunning", err)
	}
	if got := connection.writeCalls.Load(); got != 1 {
		t.Fatalf("Write calls = %d, want exactly one admitted write", got)
	}
	if runtimeSession.Running() {
		t.Fatal("Running() = true after Stop")
	}
}

func TestSessionConcurrentStopSharesTerminalResult(t *testing.T) {
	wantErr := errors.New("close failed")
	connection := newBlockingSessionConnection()
	connection.closeErr = wantErr
	runtimeSession, err := newWithConnectionDependencies(
		connection,
		validPrincipal(),
		"",
		generateID,
		nil,
		nil,
	)
	if err != nil {
		t.Fatalf("create Session: %v", err)
	}
	if err := runtimeSession.Start(context.Background()); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	connection.cleanup(t)

	const callers = 32
	begin := make(chan struct{})
	results := make(chan error, callers)
	for range callers {
		go func() {
			<-begin
			results <- runtimeSession.Stop(context.Background())
		}()
	}
	close(begin)
	waitSessionSignal(t, connection.closeEntered, "primary Close entry")
	connection.releaseCloseOnce.Do(func() { close(connection.releaseClose) })

	for caller := range callers {
		err := waitSessionResult(t, results, "concurrent Stop")
		if !errors.Is(err, wantErr) {
			t.Fatalf("Stop caller %d error = %v, want close error", caller, err)
		}
	}
	if got := connection.closeCalls.Load(); got != 1 {
		t.Fatalf("Close calls = %d, want 1", got)
	}
	if got := connection.closeNowCalls.Load(); got != 1 {
		t.Fatalf("CloseNow calls = %d, want 1", got)
	}
	if err := runtimeSession.Stop(context.Background()); !errors.Is(err, wantErr) {
		t.Fatalf("repeated Stop() error = %v, want stored close error", err)
	}
	if err := runtimeSession.Start(context.Background()); !errors.Is(err, ErrSessionAlreadyRunning) {
		t.Fatalf("Start() after terminal Stop error = %v, want ErrSessionAlreadyRunning", err)
	}
}

type blockingSessionConnection struct {
	writeEntered     chan struct{}
	releaseWrite     chan struct{}
	closeEntered     chan struct{}
	releaseClose     chan struct{}
	writeEnteredOnce sync.Once
	closeEnteredOnce sync.Once
	releaseWriteOnce sync.Once
	releaseCloseOnce sync.Once
	writeCalls       atomic.Int32
	closeCalls       atomic.Int32
	closeNowCalls    atomic.Int32
	writeErr         error
	closeErr         error
}

func newBlockingSessionConnection() *blockingSessionConnection {
	return &blockingSessionConnection{
		writeEntered: make(chan struct{}),
		releaseWrite: make(chan struct{}),
		closeEntered: make(chan struct{}),
		releaseClose: make(chan struct{}),
	}
}

func (connection *blockingSessionConnection) Read(ctx context.Context) (websocket.MessageType, []byte, error) {
	<-ctx.Done()
	return 0, nil, ctx.Err()
}

func (connection *blockingSessionConnection) Write(
	context.Context,
	websocket.MessageType,
	[]byte,
) error {
	connection.writeCalls.Add(1)
	connection.writeEnteredOnce.Do(func() { close(connection.writeEntered) })
	<-connection.releaseWrite
	return connection.writeErr
}

func (connection *blockingSessionConnection) Close(websocket.StatusCode, string) error {
	connection.closeCalls.Add(1)
	connection.closeEnteredOnce.Do(func() { close(connection.closeEntered) })
	<-connection.releaseClose
	return connection.closeErr
}

func (connection *blockingSessionConnection) CloseNow() error {
	connection.closeNowCalls.Add(1)
	return nil
}

func (connection *blockingSessionConnection) cleanup(t *testing.T) {
	t.Helper()
	t.Cleanup(func() {
		connection.releaseWriteOnce.Do(func() { close(connection.releaseWrite) })
		connection.releaseCloseOnce.Do(func() { close(connection.releaseClose) })
	})
}

func waitSessionSignal(t *testing.T, signal <-chan struct{}, operation string) {
	t.Helper()
	select {
	case <-signal:
	case <-time.After(time.Second):
		t.Fatalf("timeout waiting for %s", operation)
	}
}

func waitSessionResult(t *testing.T, result <-chan error, operation string) error {
	t.Helper()
	select {
	case err := <-result:
		return err
	case <-time.After(time.Second):
		t.Fatalf("timeout waiting for %s", operation)
		return nil
	}
}

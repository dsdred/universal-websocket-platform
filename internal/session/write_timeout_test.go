package session

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/coder/websocket"
	"github.com/dsdred/universal-websocket-platform/internal/executionowner"
	"github.com/dsdred/universal-websocket-platform/internal/message"
	"github.com/dsdred/universal-websocket-platform/internal/sessionmanager"
)

type timeoutTestConnection struct {
	write func(context.Context) error
	calls atomic.Int32
}

type timeoutStartedContext struct {
	context.Context
	started chan time.Time
}

func (c timeoutStartedContext) Deadline() (time.Time, bool) {
	select {
	case c.started <- time.Now():
	default:
	}
	return c.Context.Deadline()
}

func (c *timeoutTestConnection) Read(ctx context.Context) (websocket.MessageType, []byte, error) {
	<-ctx.Done()
	return 0, nil, ctx.Err()
}
func (c *timeoutTestConnection) Write(ctx context.Context, _ websocket.MessageType, _ []byte) error {
	c.calls.Add(1)
	return c.write(ctx)
}
func (*timeoutTestConnection) Close(websocket.StatusCode, string) error { return nil }
func (*timeoutTestConnection) CloseNow() error                          { return nil }

func timeoutTestSession(t *testing.T, timeout time.Duration, write func(context.Context) error) (*DefaultSession, *timeoutTestConnection, message.Message) {
	t.Helper()
	c := &timeoutTestConnection{write: write}
	s, err := newWithConnectionDependencies(c, validPrincipal(), "", generateID, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	s.writeTimeout = timeout
	if err := s.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	m, err := message.New(message.TypeText, []byte("bounded write"))
	if err != nil {
		t.Fatal(err)
	}
	return s, c, m
}

func TestWriteTimeoutConfiguredNetworkDeadline(t *testing.T) {
	const budget = 30 * time.Millisecond
	before := time.Now()
	s, c, m := timeoutTestSession(t, budget, func(ctx context.Context) error {
		deadline, ok := ctx.Deadline()
		if !ok || deadline.Before(before.Add(budget)) || deadline.After(time.Now().Add(budget)) {
			t.Errorf("network deadline = %v, present %t", deadline, ok)
		}
		<-ctx.Done()
		return ctx.Err()
	})
	if err := s.Send(context.Background(), m); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("Send = %v", err)
	}
	if c.calls.Load() != 1 || !s.Running() {
		t.Fatal("timeout changed lifecycle or write count")
	}
}

func TestWriteTimeoutPreservesEarlierCallerDeadline(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
	defer cancel()
	want, _ := ctx.Deadline()
	s, _, m := timeoutTestSession(t, time.Second, func(ctx context.Context) error {
		got, ok := ctx.Deadline()
		if !ok || !got.Equal(want) {
			t.Errorf("deadline = %v, want %v", got, want)
		}
		<-ctx.Done()
		return ctx.Err()
	})
	if err := s.Send(ctx, m); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("Send = %v", err)
	}
}

func TestWriteTimeoutCallerCancellationDuringNetworkWrite(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	entered := make(chan struct{})
	s, _, m := timeoutTestSession(t, time.Second, func(ctx context.Context) error {
		close(entered)
		<-ctx.Done()
		return ctx.Err()
	})
	result := make(chan error, 1)
	go func() { result <- s.Send(ctx, m) }()
	waitSessionSignal(t, entered, "network Write")
	cancel()
	if err := waitSessionResult(t, result, "cancelled Write"); !errors.Is(err, context.Canceled) {
		t.Fatalf("Send = %v", err)
	}
}

func TestWriteTimeoutConcurrentStopDoesNotWaitForAdmittedWrite(t *testing.T) {
	entered := make(chan struct{})
	release := make(chan struct{})
	s, c, m := timeoutTestSession(t, time.Second, func(ctx context.Context) error {
		close(entered)
		select {
		case <-release:
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	})
	result := make(chan error, 1)
	go func() { result <- s.Send(context.Background(), m) }()
	waitSessionSignal(t, entered, "admitted Write")
	stopResult := make(chan error, 1)
	go func() { stopResult <- s.Stop(context.Background()) }()
	if err := waitSessionResult(t, stopResult, "Stop while admitted Write waits"); err != nil {
		t.Fatalf("Stop = %v", err)
	}
	if s.Running() {
		t.Fatal("Session remained Running after Stop")
	}
	close(release)
	if err := waitSessionResult(t, result, "admitted Write after Stop"); err != nil {
		t.Fatalf("Send = %v", err)
	}
	if c.calls.Load() != 1 {
		t.Fatal("Write retried")
	}
}

func TestWriteTimeoutQueuedDeadlineDoesNotWrite(t *testing.T) {
	s, c, m := timeoutTestSession(t, 30*time.Millisecond, func(context.Context) error { return nil })
	s.writeGate <- struct{}{}
	defer func() { <-s.writeGate }()
	if err := s.Send(context.Background(), m); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("Send = %v", err)
	}
	if c.calls.Load() != 0 || !s.Running() {
		t.Fatal("queued timeout wrote or stopped Session")
	}
}

func TestWriteTimeoutQueuedCallerCancellationDoesNotWrite(t *testing.T) {
	s, c, m := timeoutTestSession(t, time.Second, func(context.Context) error { return nil })
	s.writeGate <- struct{}{}
	defer func() { <-s.writeGate }()
	ctx, cancel := context.WithCancel(context.Background())
	result := make(chan error, 1)
	go func() { result <- s.Send(ctx, m) }()
	cancel()
	if err := waitSessionResult(t, result, "queued cancellation"); !errors.Is(err, context.Canceled) {
		t.Fatalf("Send = %v", err)
	}
	if c.calls.Load() != 0 {
		t.Fatal("cancelled queued Send wrote")
	}
}

func TestWriteTimeoutQueueAndNetworkUseOneBudget(t *testing.T) {
	const budget = 150 * time.Millisecond
	deadlineSeen := make(chan time.Time, 1)
	s, _, m := timeoutTestSession(t, budget, func(ctx context.Context) error {
		deadline, _ := ctx.Deadline()
		deadlineSeen <- deadline
		<-ctx.Done()
		return ctx.Err()
	})
	s.writeGate <- struct{}{}
	started := make(chan time.Time, 1)
	ctx := timeoutStartedContext{Context: context.Background(), started: started}
	result := make(chan error, 1)
	go func() { result <- s.Send(ctx, m) }()
	var start time.Time
	select {
	case start = <-started:
	case <-time.After(time.Second):
		<-s.writeGate
		t.Fatal("Send did not establish its context before waiting")
	}
	// Wait for Send's context budget to elapse in the queue before admitting I/O.
	time.Sleep(40 * time.Millisecond)
	<-s.writeGate
	select {
	case deadline := <-deadlineSeen:
		if deadline.After(start.Add(budget)) {
			t.Fatalf("deadline reset after queue: %v", deadline.Sub(start))
		}
	case <-time.After(time.Second):
		t.Fatal("Write did not enter")
	}
	if err := waitSessionResult(t, result, "shared deadline"); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("Send = %v", err)
	}
}

func TestWriteTimeoutPreservesTransportErrorAndLegacyContext(t *testing.T) {
	want := errors.New("transport failure")
	s, _, m := timeoutTestSession(t, 0, func(ctx context.Context) error {
		if _, ok := ctx.Deadline(); ok {
			t.Error("legacy Session imposed a deadline")
		}
		return want
	})
	if err := s.Send(context.Background(), m); !errors.Is(err, want) {
		t.Fatalf("Send = %v", err)
	}
}

func TestWriteTimeoutDispatcherConstructorBoundsAndLegacy(t *testing.T) {
	root := transactionalRuntimeContextProvider{ctx: context.Background()}
	observer := &transactionalTestObserver{}
	for _, duration := range []time.Duration{-1, 0, 300*time.Second + 1, time.Duration(1<<63 - 1)} {
		if d, err := NewTransactionalDispatcherWithWriteTimeout(sessionmanager.New(), root, nil, observer, duration); err == nil || d != nil {
			t.Errorf("duration %v accepted", duration)
		}
	}
	for _, duration := range []time.Duration{time.Second, 300 * time.Second} {
		d, err := NewTransactionalDispatcherWithWriteTimeout(sessionmanager.New(), root, nil, observer, duration)
		if err != nil || d.writeTimeout != duration {
			t.Fatalf("duration %v: %v", duration, err)
		}
	}
	d, err := NewTransactionalDispatcher(sessionmanager.New(), root, nil, observer)
	if err != nil || d.writeTimeout != 0 {
		t.Fatalf("legacy constructor: %v", err)
	}
}

type timeoutCheckingReservation struct {
	sessionmanager.ReservationHandle
	beforeCommit func()
}

func (r timeoutCheckingReservation) Commit(input sessionmanager.CommitInput) (sessionmanager.CommitResult, error) {
	r.beforeCommit()
	return r.ReservationHandle.Commit(input)
}

func TestWriteTimeoutDispatcherAssignsBudgetBeforeCommit(t *testing.T) {
	const budget = 2 * time.Second
	manager := sessionmanager.New()
	root := context.Background()
	observer := &transactionalTestObserver{observed: make(chan executionowner.TerminalResult, 1)}
	dispatcher, err := NewTransactionalDispatcherWithWriteTimeout(manager, transactionalRuntimeContextProvider{ctx: root}, nil, observer, budget)
	if err != nil {
		t.Fatal(err)
	}
	handle, err := manager.Reserve("write-timeout")
	if err != nil {
		t.Fatal(err)
	}
	var prepared *provisionalSession
	checked := false
	dispatcher.manager = &transactionalReservationManager{reservation: timeoutCheckingReservation{
		ReservationHandle: handle,
		beforeCommit: func() {
			checked = true
			if prepared == nil || prepared.session.writeTimeout != budget {
				t.Error("immutable budget missing before Commit")
			}
		},
	}}
	dispatcher.generate = fixedSessionID("write-timeout")
	transport := newTransactionalTestConnection()
	dispatcher.prepare = func(core *sessionCore, _ websocketConnection, cancellation cancellationDependency) (*provisionalSession, error) {
		prepared, err = prepareProvisionalSession(core, transport, cancellation)
		return prepared, err
	}
	authenticated, cancel := transactionalAuthenticatedContext(t, root)
	defer cancel()
	accepted, err := dispatcher.DispatchAuthenticated(authenticated)
	if err != nil || !accepted || !checked {
		t.Fatalf("Dispatch = (%t,%v), checked %t", accepted, err, checked)
	}
	for _, registration := range manager.BeginShutdown().Registrations() {
		registration.RequestStop()
	}
	awaitTerminalResult(t, observer.observed)
	awaitManagerWait(t, manager)
}

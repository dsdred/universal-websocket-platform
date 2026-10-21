package session

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/coder/websocket"
	"github.com/dsdred/universal-websocket-platform/internal/executionowner"
	"github.com/dsdred/universal-websocket-platform/internal/message"
	"github.com/dsdred/universal-websocket-platform/internal/sessionmanager"
)

type readTimeoutConnection struct {
	read  func(context.Context) (websocket.MessageType, []byte, error)
	calls atomic.Int32
}

type observedReadTimeoutConnection struct {
	websocketConnection
	entered chan context.Context
}

func (c observedReadTimeoutConnection) Read(ctx context.Context) (websocket.MessageType, []byte, error) {
	c.entered <- ctx
	return c.websocketConnection.Read(ctx)
}

func (c *readTimeoutConnection) Read(ctx context.Context) (websocket.MessageType, []byte, error) {
	c.calls.Add(1)
	return c.read(ctx)
}
func (*readTimeoutConnection) Write(context.Context, websocket.MessageType, []byte) error { return nil }
func (*readTimeoutConnection) Close(websocket.StatusCode, string) error                   { return nil }
func (*readTimeoutConnection) CloseNow() error                                            { return nil }

func readTimeoutSession(t *testing.T, budget time.Duration, handler message.Handler, read func(context.Context) (websocket.MessageType, []byte, error)) (*DefaultSession, *readTimeoutConnection) {
	t.Helper()
	c := &readTimeoutConnection{read: read}
	s, err := newWithConnectionDependencies(c, validPrincipal(), "", generateID, nil, handler)
	if err != nil {
		t.Fatal(err)
	}
	s.readTimeout = budget
	if err := s.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	return s, c
}

func TestReadTimeoutDisabledPreservesOriginalContext(t *testing.T) {
	parent, cancel := context.WithCancel(context.Background())
	defer cancel()
	s, c := readTimeoutSession(t, 0, nil, func(ctx context.Context) (websocket.MessageType, []byte, error) {
		if ctx != parent {
			t.Error("disabled timeout replaced original context")
		}
		if _, ok := ctx.Deadline(); ok {
			t.Error("disabled timeout added deadline")
		}
		return 0, nil, websocket.CloseError{Code: websocket.StatusNormalClosure}
	})
	if err := s.Run(parent); err != nil {
		t.Fatal(err)
	}
	if c.calls.Load() != 1 || parent.Err() != nil {
		t.Fatal("normal close retried Read or cancelled parent")
	}
}

func TestReadTimeoutConfiguredDeadlineAndError(t *testing.T) {
	const budget = 20 * time.Millisecond
	before := time.Now()
	var child context.Context
	s, c := readTimeoutSession(t, budget, nil, func(ctx context.Context) (websocket.MessageType, []byte, error) {
		child = ctx
		deadline, ok := ctx.Deadline()
		if !ok || deadline.Before(before.Add(budget)) || deadline.After(time.Now().Add(budget)) {
			t.Errorf("read deadline %v present %t", deadline, ok)
		}
		<-ctx.Done()
		return 0, nil, fmt.Errorf("transport context: %w", ctx.Err())
	})
	if err := s.Run(context.Background()); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("Run = %v", err)
	}
	if child.Err() != context.DeadlineExceeded || c.calls.Load() != 1 || !s.Running() {
		t.Fatal("deadline retried Read or changed lifecycle")
	}
}

func TestReadTimeoutPreservesEarlierCallerDeadline(t *testing.T) {
	parent, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	want, _ := parent.Deadline()
	s, _ := readTimeoutSession(t, time.Second, nil, func(ctx context.Context) (websocket.MessageType, []byte, error) {
		got, ok := ctx.Deadline()
		if !ok || !got.Equal(want) {
			t.Errorf("deadline %v want %v", got, want)
		}
		<-ctx.Done()
		return 0, nil, ctx.Err()
	})
	if err := s.Run(parent); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("Run = %v", err)
	}
}

func TestReadTimeoutCallerCancellation(t *testing.T) {
	parent, cancel := context.WithCancel(context.Background())
	defer cancel()
	entered := make(chan struct{})
	s, _ := readTimeoutSession(t, time.Second, nil, func(ctx context.Context) (websocket.MessageType, []byte, error) {
		close(entered)
		<-ctx.Done()
		return 0, nil, ctx.Err()
	})
	result := make(chan error, 1)
	go func() { result <- s.Run(parent) }()
	waitSessionSignal(t, entered, "read entry")
	cancel()
	if err := waitSessionResult(t, result, "cancelled Read"); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
}

func TestReadTimeoutFreshBudgetExcludesHandlerAndCancelsChildren(t *testing.T) {
	const budget = 20 * time.Millisecond
	parent, cancel := context.WithCancel(context.Background())
	defer cancel()
	var previous context.Context
	var handlerDone time.Time
	reads := 0
	handler := messageHandlerFunc(func(ctx context.Context, _ message.Context) error {
		if ctx != parent || ctx.Err() != nil {
			t.Error("handler did not receive live original Run context")
		}
		if previous.Err() != context.Canceled {
			t.Errorf("read child not cancelled before handler: %v", previous.Err())
		}
		// Waiting beyond the completed Read budget proves handler time is excluded.
		<-time.After(2 * budget)
		if ctx.Err() != nil {
			t.Errorf("handler charged read budget: %v", ctx.Err())
		}
		handlerDone = time.Now()
		return nil
	})
	s, _ := readTimeoutSession(t, budget, handler, func(ctx context.Context) (websocket.MessageType, []byte, error) {
		reads++
		if reads == 1 {
			previous = ctx
			return websocket.MessageText, []byte("one"), nil
		}
		if reads != 2 {
			t.Fatal("normal close retried")
		}
		if ctx == previous {
			t.Error("next Read reused child")
		}
		deadline, ok := ctx.Deadline()
		if !ok || deadline.Before(handlerDone.Add(budget)) {
			t.Errorf("next budget charged handler time: %v", deadline)
		}
		previous = ctx
		return 0, nil, websocket.CloseError{Code: websocket.StatusGoingAway}
	})
	if err := s.Run(parent); err != nil {
		t.Fatal(err)
	}
	if reads != 2 || previous.Err() != context.Canceled || parent.Err() != nil {
		t.Fatal("failed read child cleanup or parent isolation")
	}
}

func TestReadTimeoutErrorClassificationBeforeChildCleanup(t *testing.T) {
	transportErr := errors.New("sentinel read transport")
	for _, tc := range []struct {
		name    string
		failure error
		want    error
	}{
		{"normal", websocket.CloseError{Code: websocket.StatusNormalClosure}, nil},
		{"going-away", websocket.CloseError{Code: websocket.StatusGoingAway}, nil},
		{"transport", transportErr, transportErr},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var child context.Context
			s, c := readTimeoutSession(t, time.Second, nil, func(ctx context.Context) (websocket.MessageType, []byte, error) {
				child = ctx
				return 0, nil, tc.failure
			})
			err := s.Run(context.Background())
			if !errors.Is(err, tc.want) {
				t.Fatalf("Run = %v want %v", err, tc.want)
			}
			if child.Err() != context.Canceled || c.calls.Load() != 1 {
				t.Fatal("child not cancelled or failed Read retried")
			}
		})
	}
}

func TestReadTimeoutDeadlineTakesPriorityOverCloseStatus(t *testing.T) {
	s, c := readTimeoutSession(t, 10*time.Millisecond, nil, func(ctx context.Context) (websocket.MessageType, []byte, error) {
		<-ctx.Done()
		return 0, nil, websocket.CloseError{Code: websocket.StatusNormalClosure}
	})
	if err := s.Run(context.Background()); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("Run = %v", err)
	}
	if c.calls.Load() != 1 {
		t.Fatal("expired Read retried")
	}
}

func TestReadTimeoutSuccessfulReadWinsConcurrentExpiry(t *testing.T) {
	handled := 0
	sentinel := errors.New("handler stops after complete message")
	s, _ := readTimeoutSession(t, 10*time.Millisecond, messageHandlerFunc(func(ctx context.Context, _ message.Context) error {
		handled++
		if ctx.Err() != nil {
			t.Fatal("handler inherited expired read child")
		}
		return sentinel
	}), func(ctx context.Context) (websocket.MessageType, []byte, error) {
		<-ctx.Done()
		return websocket.MessageBinary, []byte("complete"), nil
	})
	if err := s.Run(context.Background()); !errors.Is(err, sentinel) || handled != 1 {
		t.Fatalf("Run=%v handled=%d", err, handled)
	}
}

func TestReadTimeoutPanickingReadCancelsChild(t *testing.T) {
	var child context.Context
	s, _ := readTimeoutSession(t, time.Second, nil, func(ctx context.Context) (websocket.MessageType, []byte, error) { child = ctx; panic("read panic") })
	func() {
		defer func() {
			if recover() != "read panic" {
				t.Error("Read panic not propagated")
			}
		}()
		_ = s.Run(context.Background())
	}()
	if child == nil || child.Err() != context.Canceled {
		t.Fatal("panicking Read retained timer")
	}
	s.mu.RLock()
	active := s.readLoop
	s.mu.RUnlock()
	if active {
		t.Fatal("panic retained read-loop admission")
	}
}

func TestReadTimeoutDispatcherBoundsAndLegacy(t *testing.T) {
	root := transactionalRuntimeContextProvider{ctx: context.Background()}
	observer := &transactionalTestObserver{}
	for _, budget := range []time.Duration{-1, 86400*time.Second + 1, time.Duration(1<<63 - 1)} {
		if d, err := NewTransactionalDispatcherWithTimeouts(sessionmanager.New(), root, nil, observer, budget, time.Second); err == nil || d != nil {
			t.Errorf("read duration %v accepted", budget)
		}
	}
	for _, budget := range []time.Duration{0, time.Second, 86400 * time.Second} {
		d, err := NewTransactionalDispatcherWithTimeouts(sessionmanager.New(), root, nil, observer, budget, time.Second)
		if err != nil || d.readTimeout != budget || d.writeTimeout != time.Second {
			t.Fatalf("duration %v: %v", budget, err)
		}
	}
	legacy, err := NewTransactionalDispatcher(sessionmanager.New(), root, nil, observer)
	if err != nil || legacy.readTimeout != 0 || legacy.writeTimeout != 0 {
		t.Fatalf("legacy: %v", err)
	}
	write, err := NewTransactionalDispatcherWithWriteTimeout(sessionmanager.New(), root, nil, observer, time.Second)
	if err != nil || write.readTimeout != 0 || write.writeTimeout != time.Second {
		t.Fatalf("write-only: %v", err)
	}
	for _, budget := range []time.Duration{-1, 0, 300*time.Second + 1} {
		if d, err := NewTransactionalDispatcherWithTimeouts(sessionmanager.New(), root, nil, observer, time.Second, budget); err == nil || d != nil {
			t.Errorf("write duration %v accepted", budget)
		}
	}
}

func TestReadTimeoutDispatcherAssignsBothBudgetsBeforeCommit(t *testing.T) {
	manager := sessionmanager.New()
	observer := &transactionalTestObserver{observed: make(chan executionowner.TerminalResult, 1)}
	d, err := NewTransactionalDispatcherWithTimeouts(manager, transactionalRuntimeContextProvider{ctx: context.Background()}, nil, observer, 2*time.Second, 3*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	handle, err := manager.Reserve("read-budget")
	if err != nil {
		t.Fatal(err)
	}
	var prepared *provisionalSession
	checked := false
	d.manager = &transactionalReservationManager{reservation: timeoutCheckingReservation{ReservationHandle: handle, beforeCommit: func() {
		checked = true
		if prepared == nil || prepared.session.readTimeout != 2*time.Second || prepared.session.writeTimeout != 3*time.Second {
			t.Error("budgets missing before Commit")
		}
	}}}
	d.generate = fixedSessionID("read-budget")
	transport := newTransactionalTestConnection()
	d.prepare = func(core *sessionCore, _ websocketConnection, cancellation cancellationDependency) (*provisionalSession, error) {
		prepared, err = prepareProvisionalSession(core, transport, cancellation)
		return prepared, err
	}
	authenticated, cancel := transactionalAuthenticatedContext(t, context.Background())
	defer cancel()
	accepted, err := d.DispatchAuthenticated(authenticated)
	if err != nil || !accepted || !checked {
		t.Fatalf("Dispatch=(%t,%v) checked=%t", accepted, err, checked)
	}
	for _, r := range manager.BeginShutdown().Registrations() {
		r.RequestStop()
	}
	awaitTerminalResult(t, observer.observed)
	awaitManagerWait(t, manager)
}

func TestReadTimeoutOwnerFailureAndConcurrentTerminationConverge(t *testing.T) {
	for _, mode := range []string{"timeout", "explicit-stop", "root-cancel"} {
		t.Run(mode, func(t *testing.T) {
			manager := sessionmanager.New()
			root, cancelRoot := context.WithCancel(context.Background())
			defer cancelRoot()
			observer := &transactionalTestObserver{observed: make(chan executionowner.TerminalResult, 2)}
			budget := 5 * time.Second
			if mode == "timeout" {
				budget = 20 * time.Millisecond
			}
			d, err := NewTransactionalDispatcherWithTimeouts(manager, transactionalRuntimeContextProvider{ctx: root}, nil, observer, budget, time.Second)
			if err != nil {
				t.Fatal(err)
			}
			transport := newTransactionalTestConnection()
			d.prepare = func(core *sessionCore, _ websocketConnection, cancellation cancellationDependency) (*provisionalSession, error) {
				return prepareProvisionalSession(core, transport, cancellation)
			}
			authenticated, cancel := transactionalAuthenticatedContext(t, root)
			defer cancel()
			accepted, err := d.DispatchAuthenticated(authenticated)
			if err != nil || !accepted {
				t.Fatalf("Dispatch=(%t,%v)", accepted, err)
			}
			waitSessionSignal(t, transport.readStarted, "owner Read")
			if mode == "root-cancel" {
				cancelRoot()
			}
			if mode == "explicit-stop" {
				for _, r := range manager.BeginShutdown().Registrations() {
					r.RequestStop()
				}
			}
			result := awaitTerminalResult(t, observer.observed)
			if mode == "timeout" && (result.RunCategory() != executionowner.RunCategoryFailed || result.PrimaryCause() != executionowner.TerminationCauseExecutionFailure) {
				t.Fatalf("timeout result: %#v", result)
			}
			if mode == "explicit-stop" && result.PrimaryCause() != executionowner.TerminationCauseExplicitStop {
				t.Fatalf("Stop did not retain first cause: %#v", result)
			}
			if mode == "root-cancel" && result.PrimaryCause() != executionowner.TerminationCauseRuntimeCanceled && !result.SecondaryCauses().Contains(executionowner.TerminationCauseRuntimeCanceled) {
				t.Fatalf("root cancellation not recorded: %#v", result)
			}
			if result.CleanupCategory() != executionowner.CleanupCategorySucceeded || result.CompletionCategory() != executionowner.CompletionCategoryCompleted || result.CancellationCategory() != executionowner.CancellationCategoryConfirmed {
				t.Fatalf("cleanup/accounting result: %#v", result)
			}
			manager.BeginShutdown()
			awaitManagerWait(t, manager)
			if transport.reads.Load() != 1 || transport.closes.Load() != 1 || transport.closeNows.Load() != 1 || len(observer.observed) != 0 {
				t.Fatal("retry, duplicate cleanup, or duplicate terminal observation")
			}
		})
	}
}

func TestReadTimeoutPartialMessageAndControlFramesDoNotResetBudget(t *testing.T) {
	server, client := testWebSocketPair(t)
	entered := make(chan context.Context, 1)
	s, err := newWithConnectionDependencies(observedReadTimeoutConnection{websocketConnection: server, entered: entered}, validPrincipal(), "", generateID, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	const budget = 150 * time.Millisecond
	s.readTimeout = budget
	if err := s.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	writer, err := client.Writer(ctx, websocket.MessageText)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := writer.Write([]byte("unfinished fragment")); err != nil {
		t.Fatal(err)
	}
	// No final frame: the convenience Read must await the rest of the body.
	clientDone := make(chan error, 1)
	go func() { _, _, err := client.Read(ctx); clientDone <- err }()
	result := make(chan error, 1)
	go func() { result <- s.Run(context.Background()) }()
	var child context.Context
	select {
	case child = <-entered:
	case <-ctx.Done():
		t.Fatal("transport Read not entered")
	}
	wantDeadline, ok := child.Deadline()
	if !ok {
		t.Fatal("transport Read missing fixed deadline")
	}
	// Each successful ping proves the peer is processing control frames while
	// awaiting the complete application message. Ping does not complete Read.
	for i := 0; i < 3; i++ {
		if err := client.Ping(ctx); err != nil {
			t.Fatalf("Ping before expiry: %v", err)
		}
		if i < 2 {
			<-time.After(budget / 4)
		}
	}
	if err := waitSessionResult(t, result, "unfinished body timeout"); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("Run=%v", err)
	}
	if got, _ := child.Deadline(); !got.Equal(wantDeadline) || child.Err() != context.DeadlineExceeded {
		t.Fatal("fragment/control handling changed initial transport deadline")
	}
	_ = waitSessionResult(t, clientDone, "client reader transport close")
	// The transport is closed by read cancellation. Closing the uncompleted
	// writer is intentionally not required to make Run finish.
	_ = writer
}

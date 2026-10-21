package session

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/coder/websocket"
	"github.com/dsdred/universal-websocket-platform/internal/executionowner"
	"github.com/dsdred/universal-websocket-platform/internal/message"
	"github.com/dsdred/universal-websocket-platform/internal/sessionmanager"
)

type idleTestConnection struct {
	*transactionalTestConnection
	read        func(context.Context) (websocket.MessageType, []byte, error)
	write       func(context.Context) error
	beforeClose func()
}

func (c *idleTestConnection) Read(ctx context.Context) (websocket.MessageType, []byte, error) {
	if c.read != nil {
		return c.read(ctx)
	}
	return c.transactionalTestConnection.Read(ctx)
}
func (c *idleTestConnection) Write(ctx context.Context, kind websocket.MessageType, payload []byte) error {
	if c.write != nil {
		return c.write(ctx)
	}
	return c.transactionalTestConnection.Write(ctx, kind, payload)
}
func (c *idleTestConnection) Close(code websocket.StatusCode, reason string) error {
	if c.beforeClose != nil {
		c.beforeClose()
	}
	return c.transactionalTestConnection.Close(code, reason)
}

func idleTestSession(t *testing.T, idle time.Duration, handler message.Handler, c *idleTestConnection) *DefaultSession {
	t.Helper()
	if c.transactionalTestConnection == nil {
		c.transactionalTestConnection = newTransactionalTestConnection()
	}
	s, err := newWithConnectionDependencies(c, validPrincipal(), "", generateID, nil, handler)
	if err != nil {
		t.Fatal(err)
	}
	s.idleTimeout = idle
	if err := s.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	return s
}

func idleTrackerForTest(t *testing.T, s *DefaultSession) *idleTracker {
	t.Helper()
	s.mu.RLock()
	tracker := s.idle
	s.mu.RUnlock()
	if tracker == nil {
		t.Fatal("enabled Run did not publish idle tracker")
	}
	return tracker
}

func idleDeadlineForTest(tracker *idleTracker) time.Time {
	tracker.mu.Lock()
	defer tracker.mu.Unlock()
	return tracker.deadline
}

func assertIdleJoined(t *testing.T, tracker *idleTracker) {
	t.Helper()
	select {
	case <-tracker.done:
	default:
		t.Error("terminal path retained idle watchdog")
	}
}

func TestIdleTimeoutDisabledPreservesContextAndNoWatchdog(t *testing.T) {
	parent := context.WithValue(context.Background(), struct{}{}, "parent")
	var s *DefaultSession
	handled := 0
	sentinel := errors.New("finished")
	s = idleTestSession(t, 0, messageHandlerFunc(func(ctx context.Context, _ message.Context) error {
		handled++
		if ctx != parent {
			t.Error("disabled idle replaced Handler context")
		}
		return sentinel
	}), &idleTestConnection{read: func(ctx context.Context) (websocket.MessageType, []byte, error) {
		if ctx != parent || s.idle != nil {
			t.Error("disabled idle created scope")
		}
		return websocket.MessageText, []byte("complete"), nil
	}})
	if s.idle != nil {
		t.Fatal("watchdog before Run")
	}
	if err := s.Run(parent); !errors.Is(err, sentinel) || handled != 1 {
		t.Fatalf("Run=%v handled=%d", err, handled)
	}
	if s.idle != nil || parent.Err() != nil {
		t.Fatal("disabled idle left resources or cancelled parent")
	}
}

func TestIdleTimeoutExpiryIsolatedAndJoined(t *testing.T) {
	parent, cancel := context.WithCancel(context.Background())
	defer cancel()
	c := &idleTestConnection{}
	s := idleTestSession(t, 25*time.Millisecond, nil, c)
	if s.idle != nil {
		t.Fatal("idle scope before Run")
	}
	result := make(chan error, 1)
	go func() { result <- s.Run(parent) }()
	waitSessionSignal(t, c.readStarted, "idle Read")
	tracker := idleTrackerForTest(t, s)
	if err := waitSessionResult(t, result, "idle expiry"); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("Run=%v", err)
	}
	assertIdleJoined(t, tracker)
	if parent.Err() != nil || !s.Running() || c.closes.Load() != 0 {
		t.Fatal("idle changed root/lifecycle or performed Owner cleanup")
	}
}

func TestIdleTimeoutSuccessfulReadRefreshesBeforeHandler(t *testing.T) {
	var s *DefaultSession
	var before time.Time
	sentinel := errors.New("handler done")
	s = idleTestSession(t, time.Second, messageHandlerFunc(func(ctx context.Context, _ message.Context) error {
		tracker := idleTrackerForTest(t, s)
		if !idleDeadlineForTest(tracker).After(before) {
			t.Error("successful complete Read did not refresh idle")
		}
		if ctx != tracker.ctx {
			t.Error("Handler did not receive idle scope")
		}
		return sentinel
	}), &idleTestConnection{read: func(context.Context) (websocket.MessageType, []byte, error) {
		before = idleDeadlineForTest(idleTrackerForTest(t, s))
		time.Sleep(20 * time.Millisecond) // Windows clock ticks may coalesce adjacent calls.
		return websocket.MessageBinary, []byte("complete"), nil
	}})
	if err := s.Run(context.Background()); !errors.Is(err, sentinel) {
		t.Fatal(err)
	}
	assertIdleJoined(t, idleTrackerForTest(t, s))
}

func TestIdleTimeoutSuccessfulWriteRefreshesBlockedRead(t *testing.T) {
	parent, cancel := context.WithCancel(context.Background())
	defer cancel()
	c := &idleTestConnection{}
	s := idleTestSession(t, time.Second, nil, c)
	result := make(chan error, 1)
	go func() { result <- s.Run(parent) }()
	waitSessionSignal(t, c.readStarted, "blocked Read")
	tracker := idleTrackerForTest(t, s)
	before := idleDeadlineForTest(tracker)
	time.Sleep(20 * time.Millisecond)
	m, _ := message.New(message.TypeBinary, []byte("outbound"))
	if err := s.Send(context.Background(), m); err != nil {
		t.Fatal(err)
	}
	if !idleDeadlineForTest(tracker).After(before) {
		t.Fatal("successful Write did not reset idle")
	}
	cancel()
	if err := waitSessionResult(t, result, "parent cancellation"); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	assertIdleJoined(t, tracker)
}

func TestIdleTimeoutFailedAndQueuedWritesDoNotRefresh(t *testing.T) {
	parent, cancel := context.WithCancel(context.Background())
	defer cancel()
	sentinel := errors.New("failed network Write")
	c := &idleTestConnection{write: func(context.Context) error { return sentinel }}
	s := idleTestSession(t, time.Second, nil, c)
	result := make(chan error, 1)
	go func() { result <- s.Run(parent) }()
	waitSessionSignal(t, c.readStarted, "blocked Read")
	tracker := idleTrackerForTest(t, s)
	before := idleDeadlineForTest(tracker)
	m, _ := message.New(message.TypeText, []byte("failed"))
	if err := s.Send(parent, m); !errors.Is(err, sentinel) {
		t.Fatal(err)
	}
	s.writeGate <- struct{}{}
	queued, cancelQueued := context.WithCancel(parent)
	cancelQueued()
	if err := s.Send(queued, m); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	<-s.writeGate
	if !idleDeadlineForTest(tracker).Equal(before) {
		t.Fatal("failed/queued Write reset idle")
	}
	cancel()
	_ = waitSessionResult(t, result, "cancelled Run")
}

func TestIdleTimeoutHandlerTimeAndReadChildIsolation(t *testing.T) {
	parent := context.Background()
	sentinel := errors.New("handler returned after idle")
	var readChild context.Context
	s := idleTestSession(t, 25*time.Millisecond, messageHandlerFunc(func(ctx context.Context, _ message.Context) error {
		if ctx == parent || ctx == readChild || ctx.Err() != nil || readChild.Err() != context.Canceled {
			t.Error("Handler context/read child lifetime incorrect")
		}
		<-ctx.Done()
		return sentinel
	}), &idleTestConnection{read: func(ctx context.Context) (websocket.MessageType, []byte, error) {
		readChild = ctx
		return websocket.MessageText, []byte("inbound"), nil
	}})
	s.readTimeout = time.Second
	if err := s.Run(parent); !errors.Is(err, sentinel) || !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("Run lost Handler error/idle cause: %v", err)
	}
	if parent.Err() != nil {
		t.Fatal("idle cancelled parent")
	}
	assertIdleJoined(t, idleTrackerForTest(t, s))
}

func TestIdleTimeoutReadBudgetUnaffectedByWriteActivity(t *testing.T) {
	entered := make(chan context.Context, 1)
	c := &idleTestConnection{read: func(ctx context.Context) (websocket.MessageType, []byte, error) {
		entered <- ctx
		<-ctx.Done()
		return 0, nil, ctx.Err()
	}}
	s := idleTestSession(t, time.Second, nil, c)
	s.readTimeout = 40 * time.Millisecond
	result := make(chan error, 1)
	go func() { result <- s.Run(context.Background()) }()
	var read context.Context
	select {
	case read = <-entered:
	case <-time.After(time.Second):
		t.Fatal("Read did not enter")
	}
	deadline, ok := read.Deadline()
	if !ok {
		t.Fatal("missing read deadline")
	}
	m, _ := message.New(message.TypeText, []byte("activity"))
	if err := s.Send(context.Background(), m); err != nil {
		t.Fatal(err)
	}
	if got, _ := read.Deadline(); !got.Equal(deadline) {
		t.Fatal("Write activity extended read budget")
	}
	if err := waitSessionResult(t, result, "fixed read budget"); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatal(err)
	}
}

func TestIdleTimeoutAdmittedWriteBudgetAndLateSuccess(t *testing.T) {
	entered := make(chan context.Context, 1)
	release := make(chan struct{})
	c := &idleTestConnection{write: func(ctx context.Context) error { entered <- ctx; <-release; return nil }}
	s := idleTestSession(t, 25*time.Millisecond, nil, c)
	s.writeTimeout = time.Second
	run := make(chan error, 1)
	go func() { run <- s.Run(context.Background()) }()
	waitSessionSignal(t, c.readStarted, "Read before Send")
	tracker := idleTrackerForTest(t, s)
	m, _ := message.New(message.TypeText, []byte("late success"))
	send := make(chan error, 1)
	go func() { send <- s.Send(context.Background(), m) }()
	var write context.Context
	select {
	case write = <-entered:
	case <-time.After(time.Second):
		t.Fatal("Write not admitted")
	}
	if err := waitSessionResult(t, run, "idle during admitted Write"); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatal(err)
	}
	if write.Err() != nil {
		t.Fatal("idle cancelled independently admitted Write")
	}
	before := idleDeadlineForTest(tracker)
	close(release)
	if err := waitSessionResult(t, send, "late successful Write"); err != nil {
		t.Fatal(err)
	}
	if !idleDeadlineForTest(tracker).Equal(before) {
		t.Fatal("late successful Write revived idle")
	}
	assertIdleJoined(t, tracker)
}

func TestIdleTimeoutPanicAndNormalCloseJoinWatchdog(t *testing.T) {
	for _, mode := range []string{"panic", "normal", "transport"} {
		t.Run(mode, func(t *testing.T) {
			sentinel := errors.New("transport sentinel")
			var tracker *idleTracker
			var s *DefaultSession
			s = idleTestSession(t, time.Second, nil, &idleTestConnection{read: func(context.Context) (websocket.MessageType, []byte, error) {
				tracker = idleTrackerForTest(t, s)
				if mode == "panic" {
					panic("read panic")
				}
				if mode == "normal" {
					return 0, nil, websocket.CloseError{Code: websocket.StatusNormalClosure}
				}
				return 0, nil, sentinel
			}})
			func() {
				defer func() {
					got := recover()
					if mode == "panic" && got != "read panic" {
						t.Errorf("panic=%v", got)
					}
					if mode != "panic" && got != nil {
						t.Errorf("unexpected panic=%v", got)
					}
				}()
				err := s.Run(context.Background())
				if mode == "normal" && err != nil {
					t.Error(err)
				}
				if mode == "transport" && !errors.Is(err, sentinel) {
					t.Error(err)
				}
			}()
			assertIdleJoined(t, tracker)
		})
	}
}

func TestIdleTimeoutStopJoinsBeforeTransportClose(t *testing.T) {
	c := &idleTestConnection{}
	s := idleTestSession(t, time.Hour, nil, c)
	run := make(chan error, 1)
	go func() { run <- s.Run(context.Background()) }()
	waitSessionSignal(t, c.readStarted, "Read before Stop")
	tracker := idleTrackerForTest(t, s)
	c.beforeClose = func() { assertIdleJoined(t, tracker) }
	var calls sync.WaitGroup
	for i := 0; i < 8; i++ {
		calls.Add(1)
		go func() {
			defer calls.Done()
			if err := s.Stop(context.Background()); err != nil {
				t.Error(err)
			}
		}()
	}
	calls.Wait()
	_ = waitSessionResult(t, run, "Stop Run join")
	if c.closes.Load() != 1 || c.closeNows.Load() != 1 {
		t.Fatal("duplicate transport cleanup")
	}
	assertIdleJoined(t, tracker)
}

func TestIdleTimeoutDispatcherBoundsAndPreCommitAssignment(t *testing.T) {
	root := transactionalRuntimeContextProvider{ctx: context.Background()}
	observer := &transactionalTestObserver{observed: make(chan executionowner.TerminalResult, 1)}
	for _, idle := range []time.Duration{-1, 86400*time.Second + 1, time.Duration(1<<63 - 1)} {
		if d, err := NewTransactionalDispatcherWithIdleTimeout(sessionmanager.New(), root, nil, observer, 0, time.Second, idle); err == nil || d != nil {
			t.Errorf("invalid idle %v accepted", idle)
		}
	}
	for _, idle := range []time.Duration{0, time.Second, 86400 * time.Second} {
		d, err := NewTransactionalDispatcherWithIdleTimeout(sessionmanager.New(), root, nil, observer, 0, time.Second, idle)
		if err != nil || d.idleTimeout != idle {
			t.Fatalf("idle %v: %v", idle, err)
		}
	}
	legacy, err := NewTransactionalDispatcherWithTimeouts(sessionmanager.New(), root, nil, observer, time.Second, time.Second)
	if err != nil || legacy.idleTimeout != 0 {
		t.Fatal("legacy enabled idle")
	}
	manager := sessionmanager.New()
	d, err := NewTransactionalDispatcherWithIdleTimeout(manager, root, nil, observer, 2*time.Second, 3*time.Second, 4*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	handle, err := manager.Reserve("idle-budget")
	if err != nil {
		t.Fatal(err)
	}
	var prepared *provisionalSession
	checked := false
	d.manager = &transactionalReservationManager{reservation: timeoutCheckingReservation{ReservationHandle: handle, beforeCommit: func() {
		checked = true
		if prepared == nil || prepared.session.idleTimeout != 4*time.Second || prepared.session.idle != nil {
			t.Error("idle assignment/start before Commit invalid")
		}
	}}}
	d.generate = fixedSessionID("idle-budget")
	transport := newTransactionalTestConnection()
	d.prepare = func(core *sessionCore, _ websocketConnection, cancellation cancellationDependency) (*provisionalSession, error) {
		prepared, err = prepareProvisionalSession(core, transport, cancellation)
		return prepared, err
	}
	authenticated, cancel := transactionalAuthenticatedContext(t, context.Background())
	defer cancel()
	accepted, err := d.DispatchAuthenticated(authenticated)
	if err != nil || !accepted || !checked {
		t.Fatalf("Dispatch=%v,%v checked=%v", accepted, err, checked)
	}
	for _, r := range manager.BeginShutdown().Registrations() {
		r.RequestStop()
	}
	awaitTerminalResult(t, observer.observed)
	awaitManagerWait(t, manager)
}

func TestIdleTimeoutOwnerCleanupAndCancellationConverge(t *testing.T) {
	for _, mode := range []string{"idle", "explicit-stop", "root-cancel"} {
		t.Run(mode, func(t *testing.T) {
			manager := sessionmanager.New()
			root, cancelRoot := context.WithCancel(context.Background())
			defer cancelRoot()
			observer := &transactionalTestObserver{observed: make(chan executionowner.TerminalResult, 2)}
			budget := time.Second
			if mode == "idle" {
				budget = 25 * time.Millisecond
			}
			d, err := NewTransactionalDispatcherWithIdleTimeout(manager, transactionalRuntimeContextProvider{ctx: root}, nil, observer, 0, time.Second, budget)
			if err != nil {
				t.Fatal(err)
			}
			transport := newTransactionalTestConnection()
			var prepared *provisionalSession
			d.prepare = func(core *sessionCore, _ websocketConnection, cancellation cancellationDependency) (*provisionalSession, error) {
				prepared, err = prepareProvisionalSession(core, transport, cancellation)
				return prepared, err
			}
			authenticated, cancel := transactionalAuthenticatedContext(t, root)
			defer cancel()
			accepted, err := d.DispatchAuthenticated(authenticated)
			if err != nil || !accepted {
				t.Fatalf("Dispatch=%v,%v", accepted, err)
			}
			waitSessionSignal(t, transport.readStarted, "Owner Read")
			tracker := idleTrackerForTest(t, prepared.session)
			if mode == "root-cancel" {
				cancelRoot()
			}
			if mode == "explicit-stop" {
				for _, r := range manager.BeginShutdown().Registrations() {
					r.RequestStop()
				}
			}
			result := awaitTerminalResult(t, observer.observed)
			if mode == "idle" && (result.RunCategory() != executionowner.RunCategoryFailed || result.PrimaryCause() != executionowner.TerminationCauseExecutionFailure) {
				t.Fatalf("idle result=%#v", result)
			}
			if mode == "explicit-stop" && result.PrimaryCause() != executionowner.TerminationCauseExplicitStop {
				t.Fatalf("Stop cause=%#v", result)
			}
			if mode == "root-cancel" && result.PrimaryCause() != executionowner.TerminationCauseRuntimeCanceled && !result.SecondaryCauses().Contains(executionowner.TerminationCauseRuntimeCanceled) {
				t.Fatalf("root cause=%#v", result)
			}
			if result.CleanupCategory() != executionowner.CleanupCategorySucceeded || result.CompletionCategory() != executionowner.CompletionCategoryCompleted || result.CancellationCategory() != executionowner.CancellationCategoryConfirmed {
				t.Fatalf("cleanup=%#v", result)
			}
			manager.BeginShutdown()
			awaitManagerWait(t, manager)
			assertIdleJoined(t, tracker)
			if transport.closes.Load() != 1 || transport.closeNows.Load() != 1 || len(observer.observed) != 0 {
				t.Fatal("duplicate cleanup/observation")
			}
		})
	}
}

func TestIdleTimeoutFragmentsAndControlFramesDoNotReset(t *testing.T) {
	server, client := testWebSocketPair(t)
	entered := make(chan context.Context, 1)
	s, err := newWithConnectionDependencies(observedReadTimeoutConnection{websocketConnection: server, entered: entered}, validPrincipal(), "", generateID, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	s.idleTimeout = 200 * time.Millisecond
	if err := s.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	writer, err := client.Writer(ctx, websocket.MessageText)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := writer.Write([]byte("unfinished")); err != nil {
		t.Fatal(err)
	}
	clientDone := make(chan error, 1)
	go func() { _, _, err := client.Read(ctx); clientDone <- err }()
	run := make(chan error, 1)
	go func() { run <- s.Run(context.Background()) }()
	select {
	case <-entered:
	case <-ctx.Done():
		t.Fatal("Read not entered")
	}
	tracker := idleTrackerForTest(t, s)
	before := idleDeadlineForTest(tracker)
	for i := 0; i < 3; i++ {
		if err := client.Ping(ctx); err != nil {
			t.Fatal(err)
		}
		if i < 2 {
			<-time.After(35 * time.Millisecond)
		}
	}
	if !idleDeadlineForTest(tracker).Equal(before) {
		t.Fatal("fragments/control reset idle")
	}
	if err := waitSessionResult(t, run, "partial/control idle expiry"); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatal(err)
	}
	_ = waitSessionResult(t, clientDone, "client reader")
	assertIdleJoined(t, tracker)
}

func TestIdleTimeoutHandlerSendRefreshesWithoutExtendingReadBudget(t *testing.T) {
	var s *DefaultSession
	sentinel := errors.New("handler finished")
	s = idleTestSession(t, time.Second, messageHandlerFunc(func(ctx context.Context, _ message.Context) error {
		tracker := idleTrackerForTest(t, s)
		before := idleDeadlineForTest(tracker)
		time.Sleep(20 * time.Millisecond)
		m, _ := message.New(message.TypeText, []byte("handler activity"))
		if err := s.Send(ctx, m); err != nil {
			return err
		}
		if !idleDeadlineForTest(tracker).After(before) {
			t.Error("successful Handler Send did not refresh idle")
		}
		return sentinel
	}), &idleTestConnection{read: func(context.Context) (websocket.MessageType, []byte, error) {
		return websocket.MessageText, []byte("inbound"), nil
	}})
	if err := s.Run(context.Background()); !errors.Is(err, sentinel) {
		t.Fatal(err)
	}
}

func TestIdleTimeoutLateReadSuccessIsProcessedWithoutRevival(t *testing.T) {
	sentinel := errors.New("complete message processed")
	handled := 0
	s := idleTestSession(t, 20*time.Millisecond, messageHandlerFunc(func(ctx context.Context, _ message.Context) error {
		handled++
		if !errors.Is(context.Cause(ctx), context.DeadlineExceeded) {
			t.Error("late Read revived Handler scope")
		}
		return sentinel
	}), &idleTestConnection{read: func(ctx context.Context) (websocket.MessageType, []byte, error) {
		<-ctx.Done()
		return websocket.MessageBinary, []byte("late complete message"), nil
	}})
	if err := s.Run(context.Background()); handled != 1 || !errors.Is(err, sentinel) || !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("handled=%d Run=%v", handled, err)
	}
	assertIdleJoined(t, idleTrackerForTest(t, s))
}

func TestIdleTimeoutOverdueActivityCannotResurrectUnscheduledExpiry(t *testing.T) {
	tracker := newIdleTracker(context.Background(), time.Hour)
	defer tracker.close()
	// Move the real policy deadline behind now while the timer still targets the
	// old future deadline. Activity must enforce expiry without awaiting timer scheduling.
	tracker.mu.Lock()
	tracker.deadline = time.Now().Add(-time.Second)
	tracker.mu.Unlock()
	before := idleDeadlineForTest(tracker)
	tracker.activity()
	if !errors.Is(context.Cause(tracker.ctx), context.DeadlineExceeded) {
		t.Fatal("overdue activity revived policy")
	}
	tracker.activity()
	if !idleDeadlineForTest(tracker).Equal(before) {
		t.Fatal("expired tracker accepted later activity")
	}
}

func TestIdleTimeoutStaleWakeRevalidatesRefreshedDeadline(t *testing.T) {
	tracker := newIdleTracker(context.Background(), 200*time.Millisecond)
	defer tracker.close()
	// Queue a coalesced stale activity notification while holding the same mutex
	// that protects a newer deadline. Its wake must revalidate current policy.
	tracker.mu.Lock()
	tracker.deadline = time.Now().Add(time.Second)
	select {
	case tracker.wake <- struct{}{}:
	default:
	}
	tracker.mu.Unlock()
	select {
	case <-tracker.ctx.Done():
		t.Fatal("stale wake expired refreshed policy")
	case <-time.After(225 * time.Millisecond):
	}
	if tracker.ctx.Err() != nil {
		t.Fatal("stale timer ignored current deadline")
	}
}

func TestIdleTimeoutActivityExpiryAndCloseContention(t *testing.T) {
	tracker := newIdleTracker(context.Background(), time.Millisecond)
	start := make(chan struct{})
	var calls sync.WaitGroup
	for i := 0; i < 24; i++ {
		calls.Add(1)
		go func(closeScope bool) {
			defer calls.Done()
			<-start
			for n := 0; n < 20; n++ {
				tracker.activity()
			}
			if closeScope {
				tracker.close()
			}
		}(i%3 == 0)
	}
	close(start)
	calls.Wait()
	tracker.close()
	assertIdleJoined(t, tracker)
	before := idleDeadlineForTest(tracker)
	tracker.activity()
	if !idleDeadlineForTest(tracker).Equal(before) {
		t.Fatal("activity reopened terminal tracker")
	}
}

func TestIdleTimeoutReadPreservesTransportCause(t *testing.T) {
	for _, readBudget := range []time.Duration{0, time.Hour} {
		t.Run(readBudget.String(), func(t *testing.T) {
			parent, cancel := context.WithCancel(context.Background())
			defer cancel()
			sentinel := errors.New("independent Read transport failure")
			s := idleTestSession(t, 20*time.Millisecond, nil, &idleTestConnection{read: func(ctx context.Context) (websocket.MessageType, []byte, error) {
				<-ctx.Done()
				return 0, nil, sentinel
			}})
			s.readTimeout = readBudget
			err := s.Run(parent)
			if !errors.Is(err, sentinel) || !errors.Is(err, context.DeadlineExceeded) {
				t.Fatalf("Run must retain transport and idle causes: %v", err)
			}
			if parent.Err() != nil {
				t.Fatal("idle expiry cancelled original parent")
			}
			assertIdleJoined(t, idleTrackerForTest(t, s))
		})
	}
}

func TestIdleTimeoutReadPreservesOtherCancellationClassification(t *testing.T) {
	for _, mode := range []string{"disabled-parent", "enabled-parent", "read-budget", "stop"} {
		t.Run(mode, func(t *testing.T) {
			parent, cancel := context.WithCancel(context.Background())
			defer cancel()
			sentinel := errors.New("Read transport failure after other cancellation")
			entered := make(chan struct{})
			idleBudget := time.Hour
			if mode == "disabled-parent" {
				idleBudget = 0
			}
			s := idleTestSession(t, idleBudget, nil, &idleTestConnection{read: func(ctx context.Context) (websocket.MessageType, []byte, error) {
				close(entered)
				<-ctx.Done()
				return 0, nil, sentinel
			}})
			want := error(context.Canceled)
			if mode == "read-budget" {
				s.readTimeout = 20 * time.Millisecond
				want = context.DeadlineExceeded
			}
			run := make(chan error, 1)
			go func() { run <- s.Run(parent) }()
			waitSessionSignal(t, entered, "Read before cancellation control")
			switch mode {
			case "disabled-parent", "enabled-parent":
				cancel()
			case "stop":
				if err := s.Stop(context.Background()); err != nil {
					t.Fatal(err)
				}
			}
			err := waitSessionResult(t, run, "other cancellation classification")
			if !errors.Is(err, want) || errors.Is(err, sentinel) {
				t.Fatalf("unrelated cancellation classification changed: %v", err)
			}
		})
	}
}

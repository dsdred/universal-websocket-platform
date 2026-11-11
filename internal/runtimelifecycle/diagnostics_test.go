package runtimelifecycle

import (
	"context"
	"errors"
	"reflect"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/dsdred/universal-websocket-platform/internal/runtime"
	"github.com/dsdred/universal-websocket-platform/internal/runtimediagnostics"
	"github.com/dsdred/universal-websocket-platform/internal/runtimediagnosticsconsumer"
	"github.com/dsdred/universal-websocket-platform/internal/runtimediagnosticspull"
)

func TestDiagnosticsB01FailClosedAdmissionAndNoEarlyAllocation(t *testing.T) {
	producer, _ := runtimediagnostics.New()
	var launchCalls atomic.Int32
	if owner, err := newOwner(
		11,
		22,
		"runtime-prebound",
		sequentialAttemptSource(),
		&runtime.DependencyBindings{DiagnosticsProducer: *producer},
		func(*runtime.BootstrapRequest) launchResult {
			launchCalls.Add(1)
			return launchResult{}
		},
	); !errors.Is(err, ErrInvalidOwner) || owner != nil {
		t.Fatalf("newOwner(prebound) = %#v, %v; want nil, ErrInvalidOwner", owner, err)
	}
	if launchCalls.Load() != 0 {
		t.Fatalf("prebound launch calls = %d, want 0", launchCalls.Load())
	}

	owner := mustOwner(t, nil)
	preparation := mustPrepare(t, owner)
	assertAttemptHasNoDiagnostics(t, owner, preparation.attempt)

	failure := errors.New("preparation failed")
	outcome, err := owner.Start(
		context.Background(),
		preparation,
		FailedPreparation(failure),
	)
	if err != nil || outcome.Kind() != StartPreparationFailed {
		t.Fatalf("Start(failure) = %#v, %v", outcome, err)
	}
	assertAttemptHasNoDiagnostics(t, owner, preparation.attempt)

	stoppedOwner := mustOwner(t, nil)
	stoppedPreparation := mustPrepare(t, stoppedOwner)
	stopped, err := stoppedOwner.Stop(context.Background())
	if err != nil || stopped.Kind() != StopStopped {
		t.Fatalf("Stop(before acceptance) = %#v, %v", stopped, err)
	}
	assertAttemptHasNoDiagnostics(t, stoppedOwner, stoppedPreparation.attempt)
}

func TestDiagnosticsB02B05PrivateBindingAndExactAttemptPoll(t *testing.T) {
	baseReporter := func(error) {}
	base := &runtime.DependencyBindings{TerminalErrorReporter: baseReporter}
	controller := newLaunchController()
	owner, err := newOwner(
		11,
		22,
		"runtime-binding",
		sequentialAttemptSource(),
		base,
		controller.launch,
	)
	if err != nil {
		t.Fatalf("newOwner() error = %v", err)
	}
	preparation := mustPrepare(t, owner)
	result := startDiagnosticsAsync(t, owner, preparation)
	request := receiveLaunchRequest(t, controller.started)

	if request.Dependencies == base {
		t.Fatal("Launch Dependencies reused Owner.dependencies pointer")
	}
	owner.mu.Lock()
	ownedDependencies := preparation.attempt.launchDependencies
	owner.mu.Unlock()
	if ownedDependencies != request.Dependencies {
		t.Fatal("accepted attempt does not own the private launch dependency copy")
	}
	if base.DiagnosticsProducer != (runtimediagnostics.Producer{}) {
		t.Fatal("Owner.dependencies DiagnosticsProducer mutated")
	}
	if request.Dependencies.DiagnosticsProducer == (runtimediagnostics.Producer{}) {
		t.Fatal("private launch DiagnosticsProducer is disabled")
	}
	if reflect.ValueOf(request.Dependencies.TerminalErrorReporter).Pointer() !=
		reflect.ValueOf(baseReporter).Pointer() {
		t.Fatal("private dependency copy did not preserve TerminalErrorReporter")
	}

	attemptID := preparation.LoadRequest().LaunchAttemptID()
	if !request.Dependencies.DiagnosticsProducer.Offer(runtimediagnostics.ListenerServeFailed) {
		t.Fatal("Offer() = false")
	}
	if rejected, ok := owner.PollDiagnosticsExpectedAttempt(
		context.Background(), "another-attempt", 1,
	); ok || rejected != (zeroPollResult()) {
		t.Fatalf("mismatched Poll = %#v, %t", rejected, ok)
	}
	if rejected, ok := owner.PollDiagnosticsExpectedAttempt(
		context.Background(), attemptID, 0,
	); ok || rejected != (zeroPollResult()) {
		t.Fatalf("invalid-budget Poll = %#v, %t", rejected, ok)
	}

	poll, ok := owner.PollDiagnosticsExpectedAttempt(context.Background(), attemptID, 1)
	if !ok || poll.Count != 1 || poll.Events[0].Code != runtimediagnostics.ListenerServeFailed ||
		poll.Events[0].Sequence != 1 || poll.Receipt.Stop != runtimediagnosticsconsumer.Budget {
		t.Fatalf("exact Poll = %#v, %t", poll, ok)
	}
	if !request.Dependencies.DiagnosticsProducer.Offer(runtimediagnostics.ListenerServeFailed) {
		t.Fatal("second Offer() = false")
	}
	canceled, cancel := context.WithCancel(context.Background())
	cancel()
	if rejected, admitted := owner.PollDiagnosticsExpectedAttempt(canceled, attemptID, 1); admitted || rejected != (zeroPollResult()) {
		t.Fatalf("canceled Poll = %#v, %t", rejected, admitted)
	}
	poll, ok = owner.PollDiagnosticsExpectedAttempt(context.Background(), attemptID, 1)
	if !ok || poll.Count != 1 || poll.Events[0].Sequence != 2 {
		t.Fatalf("Poll after canceled rejection = %#v, %t", poll, ok)
	}

	controller.results <- launchResult{host: newProofHost(nil, nil), success: true}
	started := receiveStartResult(t, result)
	if started.err != nil || started.outcome.Kind() != StartRunning || controller.calls.Load() != 1 {
		t.Fatalf("Start() = %#v, %v; launch calls = %d", started.outcome, started.err, controller.calls.Load())
	}
	owner.mu.Lock()
	retainedLaunchDependencies := preparation.attempt.launchDependencies
	owner.mu.Unlock()
	if retainedLaunchDependencies != nil {
		t.Fatal("launch dependency copy retained after Launcher convergence")
	}
	duplicate, err := owner.Start(context.Background(), preparation, PreparationResult{})
	if err != nil || duplicate.Kind() != StartRunning || controller.calls.Load() != 1 {
		t.Fatalf("duplicate Start() = %#v, %v; launch calls = %d", duplicate, err, controller.calls.Load())
	}
}

func TestDiagnosticsB03StopDuringLaunchRetainsPairUntilCleanup(t *testing.T) {
	controller := newLaunchController()
	owner := mustOwner(t, controller.launch)
	preparation := mustPrepare(t, owner)
	startResult := startDiagnosticsAsync(t, owner, preparation)
	request := receiveLaunchRequest(t, controller.started)
	attemptID := preparation.LoadRequest().LaunchAttemptID()

	stopResult := make(chan diagnosticsStopResult, 1)
	go func() {
		outcome, err := owner.Stop(context.Background())
		stopResult <- diagnosticsStopResult{outcome: outcome, err: err}
	}()
	waitForActual(t, owner, ActualStopping)

	if !request.Dependencies.DiagnosticsProducer.Offer(runtimediagnostics.ListenerServeFailed) {
		t.Fatal("Offer() while Stop overlaps Launch = false")
	}
	poll, ok := owner.PollDiagnosticsExpectedAttempt(context.Background(), attemptID, 1)
	if !ok || poll.Count != 1 || poll.Events[0].Sequence != 1 {
		t.Fatalf("Poll() while Stop overlaps Launch = %#v, %t", poll, ok)
	}

	host := newProofHost(nil, nil)
	controller.results <- launchResult{host: host, success: true}
	started := receiveStartResult(t, startResult)
	if started.err != nil || started.outcome.Kind() != StartStoppedBeforeRunning {
		t.Fatalf("Start() = %#v, %v", started.outcome, started.err)
	}
	stopped := receiveStopResult(t, stopResult)
	if stopped.err != nil || stopped.outcome.Kind() != StopStopped || host.stopCalls.Load() != 1 {
		t.Fatalf("Stop() = %#v, %v; calls = %d", stopped.outcome, stopped.err, host.stopCalls.Load())
	}
	if _, ok := owner.PollDiagnosticsExpectedAttempt(context.Background(), attemptID, 1); ok {
		t.Fatal("Poll() admitted after successful cleanup")
	}
}

func TestDiagnosticsB04DistinctAttemptsUseIndependentPairs(t *testing.T) {
	var mu sync.Mutex
	var requests []*runtime.BootstrapRequest
	owner := mustOwner(t, func(request *runtime.BootstrapRequest) launchResult {
		mu.Lock()
		requests = append(requests, request)
		mu.Unlock()
		return launchResult{host: newProofHost(nil, nil), success: true}
	})

	firstPreparation := mustPrepare(t, owner)
	first := mustStart(t, owner, firstPreparation)
	firstID := first.Attempt().LaunchAttemptID()
	mu.Lock()
	firstProducer := requests[0].Dependencies.DiagnosticsProducer
	mu.Unlock()
	if !firstProducer.Offer(runtimediagnostics.ListenerServeFailed) {
		t.Fatal("first Offer() = false")
	}
	firstPoll, ok := owner.PollDiagnosticsExpectedAttempt(context.Background(), firstID, 1)
	if !ok || firstPoll.Count != 1 || firstPoll.Events[0].Sequence != 1 {
		t.Fatalf("first Poll() = %#v, %t", firstPoll, ok)
	}
	if outcome, err := owner.Stop(context.Background()); err != nil || outcome.Kind() != StopStopped {
		t.Fatalf("first Stop() = %#v, %v", outcome, err)
	}

	secondPreparation := mustPrepareRequest(t, owner, NewStartRequest(11, 22, 34))
	second := mustStart(t, owner, secondPreparation)
	secondID := second.Attempt().LaunchAttemptID()
	mu.Lock()
	secondProducer := requests[1].Dependencies.DiagnosticsProducer
	mu.Unlock()
	if firstProducer == secondProducer {
		t.Fatal("distinct attempts received the same Producer capability")
	}
	if !secondProducer.Offer(runtimediagnostics.AuthenticationProviderFailed) {
		t.Fatal("second Offer() = false")
	}
	if _, ok := owner.PollDiagnosticsExpectedAttempt(context.Background(), firstID, 1); ok {
		t.Fatal("historical attempt admitted against successor lifetime")
	}
	secondPoll, ok := owner.PollDiagnosticsExpectedAttempt(context.Background(), secondID, 1)
	if !ok || secondPoll.Count != 1 || secondPoll.Events[0].Sequence != 1 ||
		secondPoll.Events[0].Code != runtimediagnostics.AuthenticationProviderFailed {
		t.Fatalf("second Poll() = %#v, %t", secondPoll, ok)
	}
}

func TestDiagnosticsB04DistinctOwnersUseIndependentPairs(t *testing.T) {
	var firstRequest *runtime.BootstrapRequest
	firstOwner := mustOwner(t, func(request *runtime.BootstrapRequest) launchResult {
		firstRequest = request
		return launchResult{host: newProofHost(nil, nil), success: true}
	})
	var secondRequest *runtime.BootstrapRequest
	secondOwner := mustOwner(t, func(request *runtime.BootstrapRequest) launchResult {
		secondRequest = request
		return launchResult{host: newProofHost(nil, nil), success: true}
	})
	firstPreparation := mustPrepare(t, firstOwner)
	secondPreparation := mustPrepare(t, secondOwner)
	first := mustStart(t, firstOwner, firstPreparation)
	second := mustStart(t, secondOwner, secondPreparation)

	if firstRequest.Dependencies.DiagnosticsProducer == secondRequest.Dependencies.DiagnosticsProducer {
		t.Fatal("distinct Owners received the same Producer capability")
	}
	if !firstRequest.Dependencies.DiagnosticsProducer.Offer(runtimediagnostics.ListenerServeFailed) ||
		!secondRequest.Dependencies.DiagnosticsProducer.Offer(runtimediagnostics.AuthenticationProviderFailed) {
		t.Fatal("Offer() to independent Owner mailbox = false")
	}
	firstPoll, firstOK := firstOwner.PollDiagnosticsExpectedAttempt(
		context.Background(), first.Attempt().LaunchAttemptID(), 1,
	)
	secondPoll, secondOK := secondOwner.PollDiagnosticsExpectedAttempt(
		context.Background(), second.Attempt().LaunchAttemptID(), 1,
	)
	if !firstOK || !secondOK || firstPoll.Count != 1 || secondPoll.Count != 1 ||
		firstPoll.Events[0].Sequence != 1 || secondPoll.Events[0].Sequence != 1 ||
		firstPoll.Events[0].Code != runtimediagnostics.ListenerServeFailed ||
		secondPoll.Events[0].Code != runtimediagnostics.AuthenticationProviderFailed {
		t.Fatalf("independent Polls = %#v/%t, %#v/%t", firstPoll, firstOK, secondPoll, secondOK)
	}
}

func TestDiagnosticsB06PollClaimMayFinishAfterRevocation(t *testing.T) {
	host := newProofHost(nil, nil)
	var request *runtime.BootstrapRequest
	owner := mustOwner(t, func(candidate *runtime.BootstrapRequest) launchResult {
		request = candidate
		return launchResult{host: host, success: true}
	})
	preparation := mustPrepare(t, owner)
	started := mustStart(t, owner, preparation)
	attemptID := started.Attempt().LaunchAttemptID()
	if !request.Dependencies.DiagnosticsProducer.Offer(runtimediagnostics.ListenerServeFailed) {
		t.Fatal("Offer() = false")
	}

	ctx := newClaimGateContext()
	pollDone := make(chan diagnosticsPollResult, 1)
	go func() {
		poll, ok := owner.PollDiagnosticsExpectedAttempt(ctx, attemptID, 1)
		pollDone <- diagnosticsPollResult{poll: poll, ok: ok}
	}()
	select {
	case <-ctx.claimed:
	case <-time.After(time.Second):
		t.Fatal("Poll did not reach post-claim context sample")
	}

	stopDone := make(chan diagnosticsStopResult, 1)
	go func() {
		outcome, err := owner.Stop(context.Background())
		stopDone <- diagnosticsStopResult{outcome: outcome, err: err}
	}()
	stopped := receiveStopResult(t, stopDone)
	if stopped.err != nil || stopped.outcome.Kind() != StopStopped {
		t.Fatalf("Stop() while Poll blocked = %#v, %v", stopped.outcome, stopped.err)
	}
	close(ctx.release)
	select {
	case result := <-pollDone:
		if !result.ok || result.poll.Count != 1 || result.poll.Events[0].Sequence != 1 {
			t.Fatalf("admitted Poll = %#v, %t", result.poll, result.ok)
		}
	case <-time.After(time.Second):
		t.Fatal("admitted Poll did not finish")
	}
	if _, ok := owner.PollDiagnosticsExpectedAttempt(context.Background(), attemptID, 1); ok {
		t.Fatal("new Poll admitted after revocation")
	}
}

func TestDiagnosticsB07LaunchFailureAndStopOutcomes(t *testing.T) {
	t.Run("caller cancellation retains accepted lifetime", func(t *testing.T) {
		controller := newLaunchController()
		owner := mustOwner(t, controller.launch)
		preparation := mustPrepare(t, owner)
		snapshot := mustSnapshot(t, preparation)
		ctx, cancel := context.WithCancel(context.Background())
		startDone := make(chan diagnosticsStartResult, 1)
		go func() {
			outcome, err := owner.Start(ctx, preparation, PreparedSnapshot(snapshot))
			startDone <- diagnosticsStartResult{outcome: outcome, err: err}
		}()
		request := receiveLaunchRequest(t, controller.started)
		cancel()
		canceled := receiveStartResult(t, startDone)
		if !errors.Is(canceled.err, context.Canceled) {
			t.Fatalf("Start() after caller cancellation = %#v, %v", canceled.outcome, canceled.err)
		}

		attemptID := preparation.LoadRequest().LaunchAttemptID()
		if !request.Dependencies.DiagnosticsProducer.Offer(runtimediagnostics.ListenerServeFailed) {
			t.Fatal("Offer() after caller cancellation = false")
		}
		poll, ok := owner.PollDiagnosticsExpectedAttempt(context.Background(), attemptID, 1)
		if !ok || poll.Count != 1 || poll.Events[0].Sequence != 1 {
			t.Fatalf("Poll() after caller cancellation = %#v, %t", poll, ok)
		}

		controller.results <- launchResult{host: newProofHost(nil, nil), success: true}
		waitForActual(t, owner, ActualRunning)
		converged, err := owner.Start(context.Background(), preparation, PreparationResult{})
		if err != nil || converged.Kind() != StartRunning || controller.calls.Load() != 1 {
			t.Fatalf("converged Start() = %#v, %v; calls = %d", converged, err, controller.calls.Load())
		}
	})

	t.Run("launch failure retires", func(t *testing.T) {
		var request *runtime.BootstrapRequest
		owner := mustOwner(t, func(candidate *runtime.BootstrapRequest) launchResult {
			request = candidate
			return launchResult{success: false}
		})
		preparation := mustPrepare(t, owner)
		outcome := mustStart(t, owner, preparation)
		if outcome.Kind() != StartLaunchFailed {
			t.Fatalf("Start() kind = %q", outcome.Kind())
		}
		if request.Dependencies.DiagnosticsProducer == (runtimediagnostics.Producer{}) {
			t.Fatal("failed Launch did not receive private Producer")
		}
		if _, ok := owner.PollDiagnosticsExpectedAttempt(
			context.Background(), preparation.LoadRequest().LaunchAttemptID(), 1,
		); ok {
			t.Fatal("Poll admitted after launch failure")
		}
	})

	t.Run("failed Stop retains pair and repeated request", func(t *testing.T) {
		stopCause := errors.New("stop failed")
		host := newProofHost(nil, stopCause)
		var request *runtime.BootstrapRequest
		owner := mustOwner(t, func(candidate *runtime.BootstrapRequest) launchResult {
			request = candidate
			return launchResult{host: host, success: true}
		})
		preparation := mustPrepare(t, owner)
		started := mustStart(t, owner, preparation)
		attemptID := started.Attempt().LaunchAttemptID()
		failed, err := owner.Stop(context.Background())
		if err != nil || failed.Kind() != StopFailed || host.stopCalls.Load() != 1 {
			t.Fatalf("Stop() = %#v, %v; calls = %d", failed, err, host.stopCalls.Load())
		}
		if !request.Dependencies.DiagnosticsProducer.Offer(runtimediagnostics.ListenerServeFailed) {
			t.Fatal("Offer() after failed Stop = false")
		}
		poll, ok := owner.PollDiagnosticsExpectedAttempt(context.Background(), attemptID, 1)
		if !ok || poll.Count != 1 || poll.Events[0].Sequence != 1 {
			t.Fatalf("Poll() after failed Stop = %#v, %t", poll, ok)
		}
		repeated, err := owner.StopExpectedAttempt(context.Background(), attemptID)
		if err != nil || repeated.Kind() != StopFailed || host.stopCalls.Load() != 1 {
			t.Fatalf("repeated Stop() = %#v, %v; calls = %d", repeated, err, host.stopCalls.Load())
		}
		if !request.Dependencies.DiagnosticsProducer.Offer(runtimediagnostics.ListenerServeFailed) {
			t.Fatal("second Offer() after repeated Stop = false")
		}
		poll, ok = owner.PollDiagnosticsExpectedAttempt(context.Background(), attemptID, 1)
		if !ok || poll.Count != 1 || poll.Events[0].Sequence != 2 {
			t.Fatalf("second Poll() after repeated Stop = %#v, %t", poll, ok)
		}
	})

	t.Run("successful Stop retires", func(t *testing.T) {
		owner := mustOwner(t, nil)
		preparation := mustPrepare(t, owner)
		started := mustStart(t, owner, preparation)
		attemptID := started.Attempt().LaunchAttemptID()
		if outcome, err := owner.Stop(context.Background()); err != nil || outcome.Kind() != StopStopped {
			t.Fatalf("Stop() = %#v, %v", outcome, err)
		}
		if _, ok := owner.PollDiagnosticsExpectedAttempt(context.Background(), attemptID, 1); ok {
			t.Fatal("Poll admitted after successful Stop")
		}
	})
}

type diagnosticsStartResult struct {
	outcome StartOutcome
	err     error
}

type diagnosticsStopResult struct {
	outcome StopOutcome
	err     error
}

type diagnosticsPollResult struct {
	poll runtimediagnosticspull.PollResult
	ok   bool
}

type claimGateContext struct {
	context.Context
	calls   atomic.Int32
	claimed chan struct{}
	release chan struct{}
}

func newClaimGateContext() *claimGateContext {
	return &claimGateContext{
		Context: context.Background(),
		claimed: make(chan struct{}),
		release: make(chan struct{}),
	}
}

func (ctx *claimGateContext) Err() error {
	if ctx.calls.Add(1) == 2 {
		close(ctx.claimed)
		<-ctx.release
	}
	return nil
}

func startDiagnosticsAsync(
	t *testing.T,
	owner *Owner,
	preparation LaunchPreparation,
) <-chan diagnosticsStartResult {
	t.Helper()
	snapshot := mustSnapshot(t, preparation)
	done := make(chan diagnosticsStartResult, 1)
	go func() {
		outcome, err := owner.Start(
			context.Background(), preparation, PreparedSnapshot(snapshot),
		)
		done <- diagnosticsStartResult{outcome: outcome, err: err}
	}()
	return done
}

func receiveLaunchRequest(
	t *testing.T,
	requests <-chan *runtime.BootstrapRequest,
) *runtime.BootstrapRequest {
	t.Helper()
	select {
	case request := <-requests:
		return request
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for Launch request")
		return nil
	}
}

func receiveStartResult(t *testing.T, results <-chan diagnosticsStartResult) diagnosticsStartResult {
	t.Helper()
	select {
	case result := <-results:
		return result
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for Start result")
		return diagnosticsStartResult{}
	}
}

func receiveStopResult(t *testing.T, results <-chan diagnosticsStopResult) diagnosticsStopResult {
	t.Helper()
	select {
	case result := <-results:
		return result
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for Stop result")
		return diagnosticsStopResult{}
	}
}

func assertAttemptHasNoDiagnostics(t *testing.T, owner *Owner, attempt *attemptState) {
	t.Helper()
	owner.mu.Lock()
	defer owner.mu.Unlock()
	if attempt.diagnostics != nil {
		t.Fatalf("attempt diagnostics = %#v, want nil", attempt.diagnostics)
	}
}

func zeroPollResult() runtimediagnosticspull.PollResult {
	return runtimediagnosticspull.PollResult{}
}

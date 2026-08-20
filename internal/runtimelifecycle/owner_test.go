package runtimelifecycle

import (
	"context"
	"errors"
	"fmt"
	"net"
	"reflect"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/dsdred/universal-websocket-platform/internal/configurationversion"
	"github.com/dsdred/universal-websocket-platform/internal/runtime"
	"github.com/dsdred/universal-websocket-platform/internal/runtimeconfig"
	"github.com/dsdred/universal-websocket-platform/internal/runtimeconfigload"
	"github.com/dsdred/universal-websocket-platform/internal/secretresolver"
)

func TestPrepareStartCreatesOwnedAttemptAndPinnedLoadRequest(t *testing.T) {
	owner := mustOwner(t, nil)

	preparation, err := owner.PrepareStart(NewStartRequest(11, 22, 33))
	if err != nil {
		t.Fatalf("PrepareStart() error = %v", err)
	}

	request := preparation.LoadRequest()
	if request.WorkspaceID() != 11 ||
		request.ConfigurationID() != 22 ||
		request.ConfigurationVersionID() != 33 ||
		request.RuntimeInstanceID() != "runtime-1" ||
		request.LaunchAttemptID() != "attempt-1" {
		t.Fatalf("LoadRequest() = %#v", request)
	}
	if preparation.Context() == nil || preparation.Context().Err() != nil {
		t.Fatalf("Context() = %v", preparation.Context())
	}

	observation := owner.Observe()
	if observation.DesiredState() != DesiredRunning || observation.ActualState() != ActualStarting {
		t.Fatalf("states = %q/%q", observation.DesiredState(), observation.ActualState())
	}
	attempt, ok := observation.ActiveAttempt()
	if !ok || attempt.Phase() != AttemptPreparing ||
		attempt.ConfigurationVersionID() != 33 ||
		attempt.LaunchAttemptID() != "attempt-1" {
		t.Fatalf("ActiveAttempt() = %#v, %t", attempt, ok)
	}

	if _, err := owner.PrepareStart(NewStartRequest(11, 22, 34)); !errors.Is(err, ErrStartConflict) {
		t.Fatalf("duplicate PrepareStart() error = %v", err)
	}
}

func TestStartSuccessConvergesAndLaunchesOnce(t *testing.T) {
	host := newProofHost(nil, nil)
	var calls atomic.Int32
	owner := mustOwner(t, func(*runtime.BootstrapRequest) launchResult {
		calls.Add(1)
		return launchResult{host: host, success: true}
	})
	preparation := mustPrepare(t, owner)
	snapshot := mustSnapshot(t, preparation)

	outcome, err := owner.Start(context.Background(), preparation, PreparedSnapshot(snapshot))
	if err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	if outcome.Kind() != StartRunning || !outcome.Attempt().RunningPublished() {
		t.Fatalf("Start() = %#v", outcome)
	}
	if calls.Load() != 1 {
		t.Fatalf("launch calls = %d", calls.Load())
	}

	duplicate, err := owner.Start(context.Background(), preparation, PreparationResult{})
	if err != nil {
		t.Fatalf("duplicate Start() error = %v", err)
	}
	if duplicate.Kind() != StartRunning || calls.Load() != 1 {
		t.Fatalf("duplicate Start() = %#v, calls = %d", duplicate, calls.Load())
	}

	observation := owner.Observe()
	if observation.ActualState() != ActualRunning {
		t.Fatalf("ActualState() = %q", observation.ActualState())
	}
}

func TestStopSuccessConvergesReleasesOwnershipAndPreservesStart(t *testing.T) {
	host := newProofHost(nil, nil)
	owner := mustOwner(t, func(*runtime.BootstrapRequest) launchResult {
		return launchResult{host: host, success: true}
	})
	preparation := mustPrepare(t, owner)
	start := mustStart(t, owner, preparation)

	stop, err := owner.Stop(context.Background())
	if err != nil {
		t.Fatalf("Stop() error = %v", err)
	}
	if stop.Kind() != StopStopped || host.stopCalls.Load() != 1 {
		t.Fatalf("Stop() = %#v, calls = %d", stop, host.stopCalls.Load())
	}
	fact, ok := stop.Attempt()
	if !ok || fact.StopOrigin() != StopAfterRunning || fact.TerminalKind() != AttemptStopped {
		t.Fatalf("Stop Attempt() = %#v, %t", fact, ok)
	}

	duplicate, err := owner.Stop(context.Background())
	if err != nil || duplicate.Kind() != StopStopped {
		t.Fatalf("duplicate Stop() = %#v, %v", duplicate, err)
	}
	if _, ok := duplicate.Attempt(); ok {
		t.Fatal("idempotent Stop unexpectedly exposes an attempt")
	}
	if host.stopCalls.Load() != 1 {
		t.Fatalf("duplicate Stop calls = %d", host.stopCalls.Load())
	}
	if preparation.attempt.host != nil {
		t.Fatal("successful Stop retained direct Host reference")
	}
	if retainedHost, ok := preparation.attempt.launchOutcome.Success(); ok || retainedHost != nil {
		t.Fatal("successful attempt retained Host through BootstrapOutcome")
	}

	repeatedStart, err := owner.Start(context.Background(), preparation, PreparationResult{})
	if err != nil || repeatedStart.Kind() != start.Kind() || repeatedStart.Kind() != StartRunning {
		t.Fatalf("Start after Stop = %#v, %v", repeatedStart, err)
	}

	next, err := owner.PrepareStart(NewStartRequest(11, 22, 44))
	if err != nil {
		t.Fatalf("PrepareStart after ownership release error = %v", err)
	}
	if next.LoadRequest().LaunchAttemptID() != "attempt-2" {
		t.Fatalf("next LaunchAttemptID = %q", next.LoadRequest().LaunchAttemptID())
	}
}

func TestPreparationFailureAndLaunchFailuresPreserveExactOutcomes(t *testing.T) {
	t.Run("preparation failure", func(t *testing.T) {
		var calls atomic.Int32
		owner := mustOwner(t, func(*runtime.BootstrapRequest) launchResult {
			calls.Add(1)
			return launchResult{}
		})
		preparation := mustPrepare(t, owner)
		cause := errors.New("loader failed")

		outcome, err := owner.Start(context.Background(), preparation, FailedPreparation(cause))
		if err != nil {
			t.Fatalf("Start() error = %v", err)
		}
		got, ok := outcome.PreparationFailure()
		if outcome.Kind() != StartPreparationFailed || !ok || got != cause || calls.Load() != 0 {
			t.Fatalf("Start() = %#v, failure = %v/%t, calls = %d", outcome, got, ok, calls.Load())
		}
		if owner.Observe().ActualState() != ActualFailed {
			t.Fatalf("ActualState() = %q", owner.Observe().ActualState())
		}
	})

	t.Run("Bootstrap failure", func(t *testing.T) {
		expected := runtime.Launch(nil)
		owner := mustOwner(t, func(*runtime.BootstrapRequest) launchResult {
			return launchResult{outcome: expected}
		})
		preparation := mustPrepare(t, owner)

		outcome, err := owner.Start(
			context.Background(),
			preparation,
			PreparedSnapshot(mustSnapshot(t, preparation)),
		)
		if err != nil {
			t.Fatalf("Start() error = %v", err)
		}
		got, ok := outcome.LaunchOutcome()
		expectedFailure, expectedOK := expected.BootstrapFailure()
		gotFailure, gotOK := got.BootstrapFailure()
		if outcome.Kind() != StartLaunchFailed || !ok || !expectedOK || !gotOK ||
			gotFailure != expectedFailure {
			t.Fatalf("LaunchOutcome() did not preserve Bootstrap failure")
		}
	})

	t.Run("startup failure", func(t *testing.T) {
		preparationOwner := mustOwner(t, nil)
		preparation := mustPrepare(t, preparationOwner)
		snapshot := mustSnapshot(t, preparation)
		expected := startupFailureOutcome(t, snapshot)

		owner := mustOwner(t, func(*runtime.BootstrapRequest) launchResult {
			return launchResult{outcome: expected}
		})
		preparation = mustPrepare(t, owner)
		snapshot = mustSnapshot(t, preparation)
		outcome, err := owner.Start(
			context.Background(),
			preparation,
			PreparedSnapshot(snapshot),
		)
		if err != nil {
			t.Fatalf("Start() error = %v", err)
		}
		got, ok := outcome.LaunchOutcome()
		expectedFailure, expectedOK := expected.StartupFailure()
		gotFailure, gotOK := got.StartupFailure()
		if outcome.Kind() != StartLaunchFailed || !ok || !expectedOK || !gotOK ||
			gotFailure != expectedFailure {
			t.Fatalf("LaunchOutcome() did not preserve Startup failure")
		}
	})
}

func TestInvalidResultDoesNotConsumePreparation(t *testing.T) {
	var calls atomic.Int32
	owner := mustOwner(t, func(*runtime.BootstrapRequest) launchResult {
		calls.Add(1)
		return launchResult{host: newProofHost(nil, nil), success: true}
	})
	preparation := mustPrepare(t, owner)

	if _, err := owner.Start(context.Background(), preparation, PreparationResult{}); !errors.Is(err, ErrInvalidPreparationResult) {
		t.Fatalf("zero result error = %v", err)
	}

	foreignOwner := mustOwner(t, nil)
	foreignPreparation := mustPrepareRequest(
		t,
		foreignOwner,
		NewStartRequest(11, 22, 44),
	)
	if _, err := owner.Start(
		context.Background(),
		preparation,
		PreparedSnapshot(mustSnapshot(t, foreignPreparation)),
	); !errors.Is(err, ErrInvalidPreparationResult) {
		t.Fatalf("mismatched Snapshot error = %v", err)
	}
	if calls.Load() != 0 {
		t.Fatalf("launch calls after invalid result = %d", calls.Load())
	}

	outcome, err := owner.Start(
		context.Background(),
		preparation,
		PreparedSnapshot(mustSnapshot(t, preparation)),
	)
	if err != nil || outcome.Kind() != StartRunning || calls.Load() != 1 {
		t.Fatalf("valid retry = %#v, %v, calls = %d", outcome, err, calls.Load())
	}
}

func TestConcurrentStartUsesOneLaunchAndIgnoresLateArguments(t *testing.T) {
	controller := newLaunchController()
	host := newProofHost(nil, nil)
	owner := mustOwner(t, controller.launch)
	preparation := mustPrepare(t, owner)
	snapshot := mustSnapshot(t, preparation)

	const callers = 12
	results := make(chan startCallResult, callers)
	for range callers {
		go func() {
			outcome, err := owner.Start(
				context.Background(),
				preparation,
				PreparedSnapshot(snapshot),
			)
			results <- startCallResult{outcome: outcome, err: err}
		}()
	}

	<-controller.started
	late := make(chan startCallResult, 1)
	go func() {
		outcome, err := owner.Start(
			context.Background(),
			preparation,
			FailedPreparation(nonComparableError{"must", "be", "ignored"}),
		)
		late <- startCallResult{outcome: outcome, err: err}
	}()
	controller.results <- launchResult{host: host, success: true}

	for range callers {
		result := <-results
		if result.err != nil || result.outcome.Kind() != StartRunning {
			t.Errorf("concurrent Start = %#v, %v", result.outcome, result.err)
		}
	}
	lateResult := <-late
	if lateResult.err != nil || lateResult.outcome.Kind() != StartRunning {
		t.Fatalf("late Start = %#v, %v", lateResult.outcome, lateResult.err)
	}
	if controller.calls.Load() != 1 {
		t.Fatalf("launch calls = %d", controller.calls.Load())
	}
}

func TestConcurrentStopUsesOneHostStop(t *testing.T) {
	stopRelease := make(chan error, 1)
	host := newProofHost(stopRelease, nil)
	owner := mustOwner(t, func(*runtime.BootstrapRequest) launchResult {
		return launchResult{host: host, success: true}
	})
	preparation := mustPrepare(t, owner)
	mustStart(t, owner, preparation)

	const callers = 12
	results := make(chan stopCallResult, callers)
	for range callers {
		go func() {
			outcome, err := owner.Stop(context.Background())
			results <- stopCallResult{outcome: outcome, err: err}
		}()
	}
	<-host.stopStarted
	if host.stopCalls.Load() != 1 {
		t.Fatalf("Host.Stop calls before release = %d", host.stopCalls.Load())
	}
	stopRelease <- nil

	for range callers {
		result := <-results
		if result.err != nil || result.outcome.Kind() != StopStopped {
			t.Errorf("concurrent Stop = %#v, %v", result.outcome, result.err)
		}
	}
	if host.stopCalls.Load() != 1 {
		t.Fatalf("Host.Stop calls = %d", host.stopCalls.Load())
	}
}

func TestStopWhileLaunchFailureConvergesOnStoppedBeforeRunning(t *testing.T) {
	controller := newLaunchController()
	owner := mustOwner(t, controller.launch)
	preparation := mustPrepare(t, owner)

	startResult := make(chan startCallResult, 1)
	go func() {
		outcome, err := owner.Start(
			context.Background(),
			preparation,
			PreparedSnapshot(mustSnapshot(t, preparation)),
		)
		startResult <- startCallResult{outcome: outcome, err: err}
	}()
	<-controller.started

	stopResult := make(chan stopCallResult, 1)
	go func() {
		outcome, err := owner.Stop(context.Background())
		stopResult <- stopCallResult{outcome: outcome, err: err}
	}()
	<-preparation.Context().Done()
	controller.results <- launchResult{outcome: runtime.Launch(nil)}

	start := <-startResult
	stop := <-stopResult
	if start.err != nil || start.outcome.Kind() != StartStoppedBeforeRunning {
		t.Fatalf("Start = %#v, %v", start.outcome, start.err)
	}
	if _, ok := start.outcome.LaunchOutcome(); ok {
		t.Fatal("StartStoppedBeforeRunning exposed secondary Launch failure")
	}
	if stop.err != nil || stop.outcome.Kind() != StopStopped {
		t.Fatalf("Stop = %#v, %v", stop.outcome, stop.err)
	}
	if owner.Observe().ActualState() != ActualStopped {
		t.Fatalf("ActualState() = %q", owner.Observe().ActualState())
	}
}

func TestStopBeforeRunningFailurePreservesStartAndRetainsOwnership(t *testing.T) {
	controller := newLaunchController()
	stopCause := errors.New("late Host cleanup failed")
	host := newProofHost(nil, stopCause)
	owner := mustOwner(t, controller.launch)
	preparation := mustPrepare(t, owner)

	startResult := make(chan startCallResult, 1)
	go func() {
		outcome, err := owner.Start(
			context.Background(),
			preparation,
			PreparedSnapshot(mustSnapshot(t, preparation)),
		)
		startResult <- startCallResult{outcome: outcome, err: err}
	}()
	<-controller.started

	stopResult := make(chan stopCallResult, 1)
	go func() {
		outcome, err := owner.Stop(context.Background())
		stopResult <- stopCallResult{outcome: outcome, err: err}
	}()
	<-preparation.Context().Done()
	controller.results <- launchResult{host: host, success: true}

	start := <-startResult
	stop := <-stopResult
	if start.err != nil || start.outcome.Kind() != StartStoppedBeforeRunning {
		t.Fatalf("Start = %#v, %v", start.outcome, start.err)
	}
	if stop.err != nil || stop.outcome.Kind() != StopFailed {
		t.Fatalf("Stop = %#v, %v", stop.outcome, stop.err)
	}
	failure, ok := stop.outcome.Failure()
	if !ok || failure != stopCause {
		t.Fatalf("Stop Failure() = %v, %t", failure, ok)
	}
	repeated, err := owner.Start(context.Background(), preparation, PreparationResult{})
	if err != nil || repeated.Kind() != StartStoppedBeforeRunning {
		t.Fatalf("repeated Start = %#v, %v", repeated, err)
	}
	if _, err := owner.PrepareStart(NewStartRequest(11, 22, 44)); !errors.Is(err, ErrStartConflict) {
		t.Fatalf("PrepareStart after retained Host error = %v", err)
	}
}

func TestStopWhileStartCombinesOwnershipAndNeverPublishesRunning(t *testing.T) {
	controller := newLaunchController()
	host := newProofHost(nil, nil)
	owner := mustOwner(t, controller.launch)
	preparation := mustPrepare(t, owner)
	snapshot := mustSnapshot(t, preparation)

	startResult := make(chan startCallResult, 1)
	go func() {
		outcome, err := owner.Start(context.Background(), preparation, PreparedSnapshot(snapshot))
		startResult <- startCallResult{outcome: outcome, err: err}
	}()
	<-controller.started

	stopResult := make(chan stopCallResult, 1)
	go func() {
		outcome, err := owner.Stop(context.Background())
		stopResult <- stopCallResult{outcome: outcome, err: err}
	}()
	select {
	case <-preparation.Context().Done():
	case <-time.After(time.Second):
		t.Fatal("Stop did not cancel preparation context")
	}
	controller.results <- launchResult{host: host, success: true}

	start := <-startResult
	stop := <-stopResult
	if start.err != nil || start.outcome.Kind() != StartStoppedBeforeRunning {
		t.Fatalf("Start = %#v, %v", start.outcome, start.err)
	}
	if stop.err != nil || stop.outcome.Kind() != StopStopped {
		t.Fatalf("Stop = %#v, %v", stop.outcome, stop.err)
	}
	if start.outcome.Attempt().RunningPublished() || host.stopCalls.Load() != 1 {
		t.Fatalf("RunningPublished = %t, Host.Stop calls = %d",
			start.outcome.Attempt().RunningPublished(), host.stopCalls.Load())
	}
	if owner.Observe().ActualState() != ActualStopped {
		t.Fatalf("ActualState() = %q", owner.Observe().ActualState())
	}
}

func TestStopFromFailedWithoutHostPublishesStopped(t *testing.T) {
	owner := mustOwner(t, nil)
	preparation := mustPrepare(t, owner)
	cause := errors.New("Builder failed")
	start, err := owner.Start(context.Background(), preparation, FailedPreparation(cause))
	if err != nil || start.Kind() != StartPreparationFailed {
		t.Fatalf("Start = %#v, %v", start, err)
	}

	stop, err := owner.Stop(context.Background())
	if err != nil || stop.Kind() != StopStopped {
		t.Fatalf("Stop = %#v, %v", stop, err)
	}
	fact, ok := stop.Attempt()
	if !ok || fact.TerminalKind() != AttemptPreparationFailed {
		t.Fatalf("Stop Attempt() = %#v, %t", fact, ok)
	}
	observation := owner.Observe()
	if observation.DesiredState() != DesiredStopped || observation.ActualState() != ActualStopped {
		t.Fatalf("states = %q/%q", observation.DesiredState(), observation.ActualState())
	}
}

func TestStartWhileStopReturnsStoredRunningOutcome(t *testing.T) {
	stopRelease := make(chan error, 1)
	host := newProofHost(stopRelease, nil)
	owner := mustOwner(t, func(*runtime.BootstrapRequest) launchResult {
		return launchResult{host: host, success: true}
	})
	preparation := mustPrepare(t, owner)
	mustStart(t, owner, preparation)

	stopResult := make(chan stopCallResult, 1)
	go func() {
		outcome, err := owner.Stop(context.Background())
		stopResult <- stopCallResult{outcome: outcome, err: err}
	}()
	<-host.stopStarted

	start, err := owner.Start(context.Background(), preparation, PreparationResult{})
	if err != nil || start.Kind() != StartRunning {
		t.Fatalf("Start while Stop = %#v, %v", start, err)
	}
	if fact := start.Attempt(); !fact.RunningPublished() || fact.StopOrigin() != StopNotClaimed {
		t.Fatalf("stored Start fact = %#v", fact)
	}
	stopRelease <- nil
	if result := <-stopResult; result.err != nil || result.outcome.Kind() != StopStopped {
		t.Fatalf("Stop = %#v, %v", result.outcome, result.err)
	}
}

func TestStopDuringPreparationPreventsLaunch(t *testing.T) {
	var calls atomic.Int32
	owner := mustOwner(t, func(*runtime.BootstrapRequest) launchResult {
		calls.Add(1)
		return launchResult{}
	})
	preparation := mustPrepare(t, owner)

	stop, err := owner.Stop(context.Background())
	if err != nil || stop.Kind() != StopStopped {
		t.Fatalf("Stop() = %#v, %v", stop, err)
	}
	if preparation.Context().Err() != context.Canceled {
		t.Fatalf("preparation Context error = %v", preparation.Context().Err())
	}

	start, err := owner.Start(context.Background(), preparation, PreparationResult{})
	if err != nil || start.Kind() != StartStoppedBeforeRunning || calls.Load() != 0 {
		t.Fatalf("Start after Stop = %#v, %v, calls = %d", start, err, calls.Load())
	}
}

func TestStopFailureRetainsOwnershipAndDoesNotRetry(t *testing.T) {
	stopCause := errors.New("shutdown failed")
	host := newProofHost(nil, stopCause)
	owner := mustOwner(t, func(*runtime.BootstrapRequest) launchResult {
		return launchResult{host: host, success: true}
	})
	preparation := mustPrepare(t, owner)
	mustStart(t, owner, preparation)

	stop, err := owner.Stop(context.Background())
	if err != nil || stop.Kind() != StopFailed {
		t.Fatalf("Stop() = %#v, %v", stop, err)
	}
	failure, ok := stop.Failure()
	if !ok || failure != stopCause {
		t.Fatalf("Stop Failure() = %v, %t", failure, ok)
	}
	if owner.Observe().ActualState() != ActualFailed {
		t.Fatalf("ActualState() = %q", owner.Observe().ActualState())
	}
	if _, err := owner.PrepareStart(NewStartRequest(11, 22, 44)); !errors.Is(err, ErrStartConflict) {
		t.Fatalf("PrepareStart after Stop failure error = %v", err)
	}

	duplicate, err := owner.Stop(context.Background())
	if err != nil || duplicate.Kind() != StopFailed {
		t.Fatalf("duplicate Stop = %#v, %v", duplicate, err)
	}
	duplicateFailure, ok := duplicate.Failure()
	if !ok || duplicateFailure != stopCause || host.stopCalls.Load() != 1 {
		t.Fatalf("duplicate failure = %v/%t, calls = %d", duplicateFailure, ok, host.stopCalls.Load())
	}
}

func TestStopExpectedAttemptValidationAndMismatchDoNotMutate(t *testing.T) {
	var nilOwner *Owner
	if _, err := nilOwner.StopExpectedAttempt(context.Background(), "attempt-1"); !errors.Is(err, ErrInvalidOwner) {
		t.Fatalf("nil Owner error = %v", err)
	}

	owner := mustOwner(t, nil)
	preparation := mustPrepare(t, owner)
	if _, err := owner.StopExpectedAttempt(context.Background(), ""); !errors.Is(err, ErrInvalidExpectedAttempt) {
		t.Fatalf("empty expected ID error = %v", err)
	}
	if preparation.Context().Err() != nil {
		t.Fatalf("validation canceled preparation: %v", preparation.Context().Err())
	}

	mismatch, err := owner.StopExpectedAttempt(context.Background(), "another-attempt")
	if err != nil || mismatch.Kind() != StopAttemptMismatch {
		t.Fatalf("mismatch = %#v, %v", mismatch, err)
	}
	fact, ok := mismatch.Attempt()
	if !ok || fact.LaunchAttemptID() != preparation.LoadRequest().LaunchAttemptID() {
		t.Fatalf("mismatch Attempt() = %#v, %t", fact, ok)
	}
	if failure, ok := mismatch.Failure(); ok || failure != nil {
		t.Fatalf("mismatch Failure() = %v, %t", failure, ok)
	}
	if preparation.Context().Err() != nil || owner.Observe().ActualState() != ActualStarting {
		t.Fatalf("mismatch mutated attempt: context=%v state=%q", preparation.Context().Err(), owner.Observe().ActualState())
	}

	emptyOwner := mustOwner(t, nil)
	none, err := emptyOwner.StopExpectedAttempt(context.Background(), "attempt-1")
	if err != nil || none.Kind() != StopAttemptMismatch {
		t.Fatalf("no-attempt mismatch = %#v, %v", none, err)
	}
	if _, ok := none.Attempt(); ok {
		t.Fatal("no-attempt mismatch exposed an attempt")
	}
}

func TestStopExpectedAttemptSelectsActiveBeforeRetainedLast(t *testing.T) {
	owner := mustOwner(t, nil)
	first := mustPrepare(t, owner)
	firstID := first.LoadRequest().LaunchAttemptID()
	if outcome, err := owner.StopExpectedAttempt(context.Background(), firstID); err != nil || outcome.Kind() != StopStopped {
		t.Fatalf("stop first = %#v, %v", outcome, err)
	}

	second := mustPrepareRequest(t, owner, NewStartRequest(11, 22, 44))
	secondID := second.LoadRequest().LaunchAttemptID()
	mismatch, err := owner.StopExpectedAttempt(context.Background(), firstID)
	if err != nil || mismatch.Kind() != StopAttemptMismatch {
		t.Fatalf("old expected ID = %#v, %v", mismatch, err)
	}
	fact, ok := mismatch.Attempt()
	if !ok || fact.LaunchAttemptID() != secondID || second.Context().Err() != nil {
		t.Fatalf("relevant successor = %#v/%t, context=%v", fact, ok, second.Context().Err())
	}
	if outcome, err := owner.StopExpectedAttempt(context.Background(), secondID); err != nil || outcome.Kind() != StopStopped {
		t.Fatalf("stop successor = %#v, %v", outcome, err)
	}
}

func TestStopExpectedAttemptPreservesActivePhaseSemantics(t *testing.T) {
	t.Run("Preparing", func(t *testing.T) {
		owner := mustOwner(t, nil)
		preparation := mustPrepare(t, owner)
		outcome, err := owner.StopExpectedAttempt(context.Background(), preparation.LoadRequest().LaunchAttemptID())
		fact, ok := outcome.Attempt()
		if err != nil || outcome.Kind() != StopStopped || !ok ||
			fact.TerminalKind() != AttemptStoppedBeforeRunning || preparation.Context().Err() != context.Canceled {
			t.Fatalf("StopExpectedAttempt = %#v/%#v/%t, %v, context=%v", outcome, fact, ok, err, preparation.Context().Err())
		}
	})

	t.Run("Launching", func(t *testing.T) {
		controller := newLaunchController()
		owner := mustOwner(t, controller.launch)
		preparation := mustPrepare(t, owner)
		startResult := make(chan startCallResult, 1)
		go func() {
			outcome, err := owner.Start(context.Background(), preparation, PreparedSnapshot(mustSnapshot(t, preparation)))
			startResult <- startCallResult{outcome: outcome, err: err}
		}()
		<-controller.started

		stopResult := make(chan stopCallResult, 1)
		go func() {
			outcome, err := owner.StopExpectedAttempt(context.Background(), preparation.LoadRequest().LaunchAttemptID())
			stopResult <- stopCallResult{outcome: outcome, err: err}
		}()
		<-preparation.Context().Done()
		controller.results <- launchResult{outcome: runtime.Launch(nil)}
		if result := <-startResult; result.err != nil || result.outcome.Kind() != StartStoppedBeforeRunning {
			t.Fatalf("Start = %#v, %v", result.outcome, result.err)
		}
		if result := <-stopResult; result.err != nil || result.outcome.Kind() != StopStopped {
			t.Fatalf("StopExpectedAttempt = %#v, %v", result.outcome, result.err)
		}
	})

	t.Run("Running and Stopping convergence", func(t *testing.T) {
		stopRelease := make(chan error, 1)
		host := newProofHost(stopRelease, nil)
		owner := mustOwner(t, func(*runtime.BootstrapRequest) launchResult {
			return launchResult{host: host, success: true}
		})
		preparation := mustPrepare(t, owner)
		mustStart(t, owner, preparation)
		expectedID := preparation.LoadRequest().LaunchAttemptID()

		results := make(chan stopCallResult, 2)
		go func() {
			outcome, err := owner.StopExpectedAttempt(context.Background(), expectedID)
			results <- stopCallResult{outcome: outcome, err: err}
		}()
		<-host.stopStarted
		attachmentContext := &observedDoneContext{
			Context:  context.Background(),
			observed: make(chan struct{}),
		}
		go func() {
			outcome, err := owner.StopExpectedAttempt(attachmentContext, expectedID)
			results <- stopCallResult{outcome: outcome, err: err}
		}()
		<-attachmentContext.observed
		mismatch, err := owner.StopExpectedAttempt(context.Background(), "another-attempt")
		if err != nil || mismatch.Kind() != StopAttemptMismatch {
			t.Fatalf("different-ID call = %#v, %v", mismatch, err)
		}
		fact, ok := mismatch.Attempt()
		if !ok || fact.Phase() != AttemptStopping || host.stopCalls.Load() != 1 {
			t.Fatalf("different-ID relevant fact = %#v/%t, calls=%d", fact, ok, host.stopCalls.Load())
		}
		stopRelease <- nil
		for range 2 {
			result := <-results
			if result.err != nil || result.outcome.Kind() != StopStopped {
				t.Errorf("same-ID convergence = %#v, %v", result.outcome, result.err)
			}
		}
		if host.stopCalls.Load() != 1 {
			t.Fatalf("Host.Stop calls = %d", host.stopCalls.Load())
		}
	})
}

func TestStopExpectedAttemptRetainedOutcomes(t *testing.T) {
	t.Run("Stop failure", func(t *testing.T) {
		stopCause := errors.New("shutdown failed")
		host := newProofHost(nil, stopCause)
		owner := mustOwner(t, func(*runtime.BootstrapRequest) launchResult {
			return launchResult{host: host, success: true}
		})
		preparation := mustPrepare(t, owner)
		mustStart(t, owner, preparation)
		expectedID := preparation.LoadRequest().LaunchAttemptID()
		if outcome, err := owner.StopExpectedAttempt(context.Background(), expectedID); err != nil || outcome.Kind() != StopFailed {
			t.Fatalf("first StopExpectedAttempt = %#v, %v", outcome, err)
		}
		outcome, err := owner.StopExpectedAttempt(context.Background(), expectedID)
		failure, ok := outcome.Failure()
		if err != nil || outcome.Kind() != StopFailed || !ok || failure != stopCause || host.stopCalls.Load() != 1 {
			t.Fatalf("retained StopExpectedAttempt = %#v, %v/%t, err=%v calls=%d", outcome, failure, ok, err, host.stopCalls.Load())
		}
	})

	for _, running := range []bool{false, true} {
		name := "Stopped before running replay"
		if running {
			name = "Stopped after running replay"
		}
		t.Run(name, func(t *testing.T) {
			owner := mustOwner(t, nil)
			preparation := mustPrepare(t, owner)
			if running {
				mustStart(t, owner, preparation)
			}
			expectedID := preparation.LoadRequest().LaunchAttemptID()
			if _, err := owner.StopExpectedAttempt(context.Background(), expectedID); err != nil {
				t.Fatalf("first StopExpectedAttempt error = %v", err)
			}
			outcome, err := owner.StopExpectedAttempt(context.Background(), expectedID)
			fact, ok := outcome.Attempt()
			if err != nil || outcome.Kind() != StopStopped || !ok || fact.LaunchAttemptID() != expectedID {
				t.Fatalf("retained replay = %#v/%#v/%t, %v", outcome, fact, ok, err)
			}
			wantTerminal := AttemptStoppedBeforeRunning
			if running {
				wantTerminal = AttemptStopped
			}
			if fact.TerminalKind() != wantTerminal {
				t.Fatalf("TerminalKind() = %q, want %q", fact.TerminalKind(), wantTerminal)
			}
		})
	}

	tests := []struct {
		name  string
		start func(*testing.T, *Owner, LaunchPreparation)
		kind  AttemptTerminalKind
	}{
		{
			name: "Preparation failure",
			start: func(t *testing.T, owner *Owner, preparation LaunchPreparation) {
				if outcome, err := owner.Start(context.Background(), preparation, FailedPreparation(errors.New("load failed"))); err != nil || outcome.Kind() != StartPreparationFailed {
					t.Fatalf("Start = %#v, %v", outcome, err)
				}
			},
			kind: AttemptPreparationFailed,
		},
		{
			name: "Launch failure",
			start: func(t *testing.T, owner *Owner, preparation LaunchPreparation) {
				if outcome, err := owner.Start(context.Background(), preparation, PreparedSnapshot(mustSnapshot(t, preparation))); err != nil || outcome.Kind() != StartLaunchFailed {
					t.Fatalf("Start = %#v, %v", outcome, err)
				}
			},
			kind: AttemptLaunchFailed,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			owner := mustOwner(t, func(*runtime.BootstrapRequest) launchResult {
				return launchResult{outcome: runtime.Launch(nil)}
			})
			preparation := mustPrepare(t, owner)
			test.start(t, owner, preparation)
			outcome, err := owner.StopExpectedAttempt(context.Background(), preparation.LoadRequest().LaunchAttemptID())
			fact, ok := outcome.Attempt()
			observation := owner.Observe()
			if err != nil || outcome.Kind() != StopStopped || !ok || fact.TerminalKind() != test.kind ||
				observation.DesiredState() != DesiredStopped || observation.ActualState() != ActualStopped {
				t.Fatalf("StopExpectedAttempt = %#v/%#v/%t, %v, observation=%#v", outcome, fact, ok, err, observation)
			}
		})
	}

	t.Run("Impossible retained state", func(t *testing.T) {
		owner := mustOwner(t, nil)
		preparation := mustPrepare(t, owner)
		expectedID := preparation.LoadRequest().LaunchAttemptID()
		owner.mu.Lock()
		owner.active = nil
		owner.last = preparation.attempt
		owner.actual = ActualStopped
		before := preparation.attempt.fact
		owner.mu.Unlock()

		if _, err := owner.StopExpectedAttempt(context.Background(), expectedID); !errors.Is(err, ErrStartConflict) {
			t.Fatalf("StopExpectedAttempt error = %v", err)
		}
		owner.mu.Lock()
		after := preparation.attempt.fact
		actual := owner.actual
		owner.mu.Unlock()
		if after != before || actual != ActualStopped {
			t.Fatalf("impossible state mutated: before=%#v after=%#v actual=%q", before, after, actual)
		}
	})
}

func TestStopExpectedAttemptCancellationLinearization(t *testing.T) {
	t.Run("visible before locked check", func(t *testing.T) {
		owner := mustOwner(t, nil)
		preparation := mustPrepare(t, owner)
		ctx, cancel := context.WithCancel(context.Background())
		owner.mu.Lock()
		result := make(chan stopCallResult, 1)
		go func() {
			outcome, err := owner.StopExpectedAttempt(ctx, preparation.LoadRequest().LaunchAttemptID())
			result <- stopCallResult{outcome: outcome, err: err}
		}()
		cancel()
		owner.mu.Unlock()
		if got := <-result; !errors.Is(got.err, context.Canceled) {
			t.Fatalf("StopExpectedAttempt error = %v", got.err)
		}
		if preparation.Context().Err() != nil || owner.Observe().ActualState() != ActualStarting {
			t.Fatalf("canceled call mutated state: context=%v state=%q", preparation.Context().Err(), owner.Observe().ActualState())
		}
	})

	t.Run("after claim cancels only waiter", func(t *testing.T) {
		stopRelease := make(chan error, 1)
		host := newProofHost(stopRelease, nil)
		owner := mustOwner(t, func(*runtime.BootstrapRequest) launchResult {
			return launchResult{host: host, success: true}
		})
		preparation := mustPrepare(t, owner)
		mustStart(t, owner, preparation)
		expectedID := preparation.LoadRequest().LaunchAttemptID()
		ctx, cancel := context.WithCancel(context.Background())
		result := make(chan stopCallResult, 1)
		go func() {
			outcome, err := owner.StopExpectedAttempt(ctx, expectedID)
			result <- stopCallResult{outcome: outcome, err: err}
		}()
		<-host.stopStarted
		cancel()
		if got := <-result; !errors.Is(got.err, context.Canceled) {
			t.Fatalf("wait error = %v", got.err)
		}
		stopRelease <- nil
		waitForActual(t, owner, ActualStopped)
		outcome, err := owner.StopExpectedAttempt(context.Background(), expectedID)
		if err != nil || outcome.Kind() != StopStopped || host.stopCalls.Load() != 1 {
			t.Fatalf("converged replay = %#v, %v, calls=%d", outcome, err, host.stopCalls.Load())
		}
	})
}

func TestStopExpectedAttemptRunsWorkOutsideMutexAndOwnersRemainIndependent(t *testing.T) {
	cancelOwner := mustOwner(t, nil)
	cancelPreparation := mustPrepare(t, cancelOwner)
	cancelObserved := make(chan struct{})
	originalCancel := cancelPreparation.attempt.cancel
	cancelPreparation.attempt.cancel = func() {
		cancelOwner.Observe()
		originalCancel()
		close(cancelObserved)
	}
	cancelDone := make(chan stopCallResult, 1)
	go func() {
		outcome, err := cancelOwner.StopExpectedAttempt(
			context.Background(),
			cancelPreparation.LoadRequest().LaunchAttemptID(),
		)
		cancelDone <- stopCallResult{outcome: outcome, err: err}
	}()
	select {
	case <-cancelObserved:
	case <-time.After(time.Second):
		t.Fatal("preparation cancellation appears to run under Owner mutex")
	}
	if result := <-cancelDone; result.err != nil || result.outcome.Kind() != StopStopped {
		t.Fatalf("preparing StopExpectedAttempt = %#v, %v", result.outcome, result.err)
	}

	var first *Owner
	firstHost := newProofHost(nil, nil)
	firstHost.onStop = func() { first.Observe() }
	first = mustOwner(t, func(*runtime.BootstrapRequest) launchResult {
		return launchResult{host: firstHost, success: true}
	})
	firstPreparation := mustPrepare(t, first)
	mustStart(t, first, firstPreparation)

	second := mustOwner(t, nil)
	secondPreparation := mustPrepare(t, second)
	secondDone := make(chan stopCallResult, 1)
	go func() {
		outcome, err := second.StopExpectedAttempt(context.Background(), secondPreparation.LoadRequest().LaunchAttemptID())
		secondDone <- stopCallResult{outcome: outcome, err: err}
	}()
	select {
	case result := <-secondDone:
		if result.err != nil || result.outcome.Kind() != StopStopped {
			t.Fatalf("independent Owner StopExpectedAttempt = %#v, %v", result.outcome, result.err)
		}
	case <-time.After(time.Second):
		t.Fatal("independent Owner did not progress")
	}

	firstDone := make(chan stopCallResult, 1)
	go func() {
		outcome, err := first.StopExpectedAttempt(context.Background(), firstPreparation.LoadRequest().LaunchAttemptID())
		firstDone <- stopCallResult{outcome: outcome, err: err}
	}()
	select {
	case result := <-firstDone:
		if result.err != nil || result.outcome.Kind() != StopStopped {
			t.Fatalf("StopExpectedAttempt = %#v, %v", result.outcome, result.err)
		}
	case <-time.After(time.Second):
		t.Fatal("Host.Stop appears to run under Owner mutex")
	}
}

func TestCallerCancellationOnlyCancelsWaitAfterClaim(t *testing.T) {
	t.Run("Start", func(t *testing.T) {
		controller := newLaunchController()
		owner := mustOwner(t, controller.launch)
		preparation := mustPrepare(t, owner)
		snapshot := mustSnapshot(t, preparation)

		ctx, cancel := context.WithCancel(context.Background())
		result := make(chan startCallResult, 1)
		go func() {
			outcome, err := owner.Start(ctx, preparation, PreparedSnapshot(snapshot))
			result <- startCallResult{outcome: outcome, err: err}
		}()
		<-controller.started
		cancel()
		if got := <-result; !errors.Is(got.err, context.Canceled) {
			t.Fatalf("Start wait error = %v", got.err)
		}

		controller.results <- launchResult{host: newProofHost(nil, nil), success: true}
		waitForActual(t, owner, ActualRunning)
		outcome, err := owner.Start(context.Background(), preparation, PreparationResult{})
		if err != nil || outcome.Kind() != StartRunning {
			t.Fatalf("convergent Start = %#v, %v", outcome, err)
		}
	})

	t.Run("Stop", func(t *testing.T) {
		stopRelease := make(chan error, 1)
		host := newProofHost(stopRelease, nil)
		owner := mustOwner(t, func(*runtime.BootstrapRequest) launchResult {
			return launchResult{host: host, success: true}
		})
		preparation := mustPrepare(t, owner)
		mustStart(t, owner, preparation)

		ctx, cancel := context.WithCancel(context.Background())
		result := make(chan stopCallResult, 1)
		go func() {
			outcome, err := owner.Stop(ctx)
			result <- stopCallResult{outcome: outcome, err: err}
		}()
		<-host.stopStarted
		cancel()
		if got := <-result; !errors.Is(got.err, context.Canceled) {
			t.Fatalf("Stop wait error = %v", got.err)
		}

		stopRelease <- nil
		waitForActual(t, owner, ActualStopped)
		if host.stopCalls.Load() != 1 {
			t.Fatalf("Host.Stop calls = %d", host.stopCalls.Load())
		}
	})
}

func TestCancellationVisibleBeforeClaimDoesNotMutate(t *testing.T) {
	var calls atomic.Int32
	owner := mustOwner(t, func(*runtime.BootstrapRequest) launchResult {
		calls.Add(1)
		return launchResult{host: newProofHost(nil, nil), success: true}
	})
	preparation := mustPrepare(t, owner)
	snapshot := mustSnapshot(t, preparation)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if _, err := owner.Start(ctx, preparation, PreparedSnapshot(snapshot)); !errors.Is(err, context.Canceled) {
		t.Fatalf("Start error = %v", err)
	}
	if calls.Load() != 0 || owner.Observe().ActualState() != ActualStarting {
		t.Fatalf("calls/state = %d/%q", calls.Load(), owner.Observe().ActualState())
	}
	if _, err := owner.Stop(ctx); !errors.Is(err, context.Canceled) {
		t.Fatalf("Stop error = %v", err)
	}
	if owner.Observe().ActualState() != ActualStarting {
		t.Fatalf("Stop mutated state to %q", owner.Observe().ActualState())
	}

	mustStart(t, owner, preparation)
}

func TestConcurrentPrepareStartCommitsAtMostOneAttempt(t *testing.T) {
	var next atomic.Int32
	owner, err := newOwner(
		11,
		22,
		"runtime-1",
		func() (runtimeconfigload.LaunchAttemptID, error) {
			return runtimeconfigload.LaunchAttemptID(
				fmt.Sprintf("candidate-%d", next.Add(1)),
			), nil
		},
		&runtime.DependencyBindings{},
		launchRuntime,
	)
	if err != nil {
		t.Fatalf("newOwner() error = %v", err)
	}

	const callers = 20
	var successes atomic.Int32
	var wait sync.WaitGroup
	wait.Add(callers)
	for range callers {
		go func() {
			defer wait.Done()
			if _, err := owner.PrepareStart(NewStartRequest(11, 22, 33)); err == nil {
				successes.Add(1)
			} else if !errors.Is(err, ErrStartConflict) {
				t.Errorf("PrepareStart() error = %v", err)
			}
		}()
	}
	wait.Wait()
	if successes.Load() != 1 {
		t.Fatalf("successful claims = %d", successes.Load())
	}
	attempt, ok := owner.Observe().ActiveAttempt()
	if !ok || attempt.Phase() != AttemptPreparing {
		t.Fatalf("ActiveAttempt() = %#v, %t", attempt, ok)
	}
}

func TestAttemptIDFailuresAndReuseDoNotCreateAttempts(t *testing.T) {
	sourceCause := errors.New("allocator unavailable")
	owner, err := NewOwner(
		11,
		22,
		"runtime-1",
		func() (runtimeconfigload.LaunchAttemptID, error) { return "", sourceCause },
		&runtime.DependencyBindings{},
	)
	if err != nil {
		t.Fatalf("NewOwner() error = %v", err)
	}
	if _, err := owner.PrepareStart(NewStartRequest(11, 22, 33)); !errors.Is(err, ErrAttemptIDSourceFailed) ||
		!errors.Is(err, sourceCause) {
		t.Fatalf("source error = %v", err)
	}
	if _, ok := owner.Observe().ActiveAttempt(); ok {
		t.Fatal("source failure created an attempt")
	}

	ids := make(chan runtimeconfigload.LaunchAttemptID, 2)
	ids <- "same-attempt"
	ids <- "same-attempt"
	close(ids)
	owner, err = NewOwner(
		11,
		22,
		"runtime-1",
		func() (runtimeconfigload.LaunchAttemptID, error) { return <-ids, nil },
		&runtime.DependencyBindings{},
	)
	if err != nil {
		t.Fatalf("NewOwner() error = %v", err)
	}
	preparation := mustPrepareRequest(t, owner, NewStartRequest(11, 22, 33))
	if _, err := owner.Stop(context.Background()); err != nil {
		t.Fatalf("Stop() error = %v", err)
	}
	if _, err := owner.PrepareStart(NewStartRequest(11, 22, 44)); !errors.Is(err, ErrAttemptIDReused) {
		t.Fatalf("reused ID error = %v", err)
	}
	if outcome, err := owner.Start(context.Background(), preparation, PreparationResult{}); err != nil ||
		outcome.Kind() != StartStoppedBeforeRunning {
		t.Fatalf("historical Start = %#v, %v", outcome, err)
	}
}

func TestForeignPreparationAndIndependentOwners(t *testing.T) {
	release := make(chan launchResult, 1)
	started := make(chan struct{}, 1)
	first := mustOwner(t, func(*runtime.BootstrapRequest) launchResult {
		started <- struct{}{}
		return <-release
	})
	second := mustOwner(t, func(*runtime.BootstrapRequest) launchResult {
		return launchResult{host: newProofHost(nil, nil), success: true}
	})
	firstPreparation := mustPrepare(t, first)
	secondPreparation := mustPrepare(t, second)

	if _, err := first.Start(
		context.Background(),
		secondPreparation,
		PreparedSnapshot(mustSnapshot(t, secondPreparation)),
	); !errors.Is(err, ErrPreparationNotOwned) {
		t.Fatalf("foreign preparation error = %v", err)
	}

	firstResult := make(chan startCallResult, 1)
	go func() {
		outcome, err := first.Start(
			context.Background(),
			firstPreparation,
			PreparedSnapshot(mustSnapshot(t, firstPreparation)),
		)
		firstResult <- startCallResult{outcome: outcome, err: err}
	}()
	<-started

	secondOutcome := mustStart(t, second, secondPreparation)
	if secondOutcome.Kind() != StartRunning {
		t.Fatalf("second Owner Start = %#v", secondOutcome)
	}
	release <- launchResult{host: newProofHost(nil, nil), success: true}
	if result := <-firstResult; result.err != nil || result.outcome.Kind() != StartRunning {
		t.Fatalf("first Owner Start = %#v, %v", result.outcome, result.err)
	}
}

func TestRuntimeCallsOccurOutsideOwnerMutex(t *testing.T) {
	var owner *Owner
	host := newProofHost(nil, nil)
	host.onStop = func() {
		owner.Observe()
	}
	var err error
	owner, err = newOwner(
		11,
		22,
		"runtime-1",
		sequentialAttemptSource(),
		&runtime.DependencyBindings{},
		func(*runtime.BootstrapRequest) launchResult {
			owner.Observe()
			return launchResult{host: host, success: true}
		},
	)
	if err != nil {
		t.Fatalf("newOwner() error = %v", err)
	}

	preparation := mustPrepare(t, owner)
	startDone := make(chan startCallResult, 1)
	go func() {
		outcome, err := owner.Start(
			context.Background(),
			preparation,
			PreparedSnapshot(mustSnapshot(t, preparation)),
		)
		startDone <- startCallResult{outcome: outcome, err: err}
	}()
	select {
	case result := <-startDone:
		if result.err != nil || result.outcome.Kind() != StartRunning {
			t.Fatalf("Start = %#v, %v", result.outcome, result.err)
		}
	case <-time.After(time.Second):
		t.Fatal("Launcher appears to run under Owner mutex")
	}

	stopDone := make(chan stopCallResult, 1)
	go func() {
		outcome, err := owner.Stop(context.Background())
		stopDone <- stopCallResult{outcome: outcome, err: err}
	}()
	select {
	case result := <-stopDone:
		if result.err != nil || result.outcome.Kind() != StopStopped {
			t.Fatalf("Stop = %#v, %v", result.outcome, result.err)
		}
	case <-time.After(time.Second):
		t.Fatal("Host.Stop appears to run under Owner mutex")
	}
}

func TestAttemptStateDoesNotRetainSnapshot(t *testing.T) {
	attemptType := reflect.TypeFor[attemptState]()
	snapshotType := reflect.TypeFor[runtimeconfig.Snapshot]()
	for index := range attemptType.NumField() {
		field := attemptType.Field(index)
		if field.Type == snapshotType {
			t.Fatalf("attemptState retains Snapshot in field %q", field.Name)
		}
	}
}

func TestConstructorAndRequestValidation(t *testing.T) {
	validSource := sequentialAttemptSource()
	validDependencies := &runtime.DependencyBindings{}
	tests := []struct {
		name          string
		workspace     uint64
		configuration uint64
		instance      runtimeconfigload.RuntimeInstanceID
		source        LaunchAttemptIDSource
		dependencies  *runtime.DependencyBindings
	}{
		{"workspace", 0, 22, "runtime-1", validSource, validDependencies},
		{"configuration", 11, 0, "runtime-1", validSource, validDependencies},
		{"instance", 11, 22, "", validSource, validDependencies},
		{"source", 11, 22, "runtime-1", nil, validDependencies},
		{"dependencies", 11, 22, "runtime-1", validSource, nil},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if _, err := NewOwner(
				test.workspace,
				test.configuration,
				test.instance,
				test.source,
				test.dependencies,
			); !errors.Is(err, ErrInvalidOwner) {
				t.Fatalf("NewOwner() error = %v", err)
			}
		})
	}

	owner := mustOwner(t, nil)
	invalid := []StartRequest{
		NewStartRequest(0, 22, 33),
		NewStartRequest(11, 0, 33),
		NewStartRequest(11, 22, 0),
		NewStartRequest(99, 22, 33),
		NewStartRequest(11, 99, 33),
	}
	for _, request := range invalid {
		if _, err := owner.PrepareStart(request); !errors.Is(err, ErrInvalidStartRequest) {
			t.Errorf("PrepareStart(%#v) error = %v", request, err)
		}
	}
	if _, ok := owner.Observe().ActiveAttempt(); ok {
		t.Fatal("invalid request created an attempt")
	}
}

type startCallResult struct {
	outcome StartOutcome
	err     error
}

type stopCallResult struct {
	outcome StopOutcome
	err     error
}

type nonComparableError []string

func (nonComparableError) Error() string { return "non-comparable preparation failure" }

type observedDoneContext struct {
	context.Context
	observed chan struct{}
	once     sync.Once
}

func (ctx *observedDoneContext) Done() <-chan struct{} {
	ctx.once.Do(func() { close(ctx.observed) })
	return ctx.Context.Done()
}

type launchController struct {
	calls   atomic.Int32
	started chan *runtime.BootstrapRequest
	results chan launchResult
}

func newLaunchController() *launchController {
	return &launchController{
		started: make(chan *runtime.BootstrapRequest, 1),
		results: make(chan launchResult, 1),
	}
}

func (controller *launchController) launch(request *runtime.BootstrapRequest) launchResult {
	controller.calls.Add(1)
	controller.started <- request
	return <-controller.results
}

type proofHost struct {
	stopCalls   atomic.Int32
	stopStarted chan struct{}
	stopRelease <-chan error
	stopErr     error
	onStop      func()
}

func newProofHost(stopRelease <-chan error, stopErr error) *proofHost {
	return &proofHost{
		stopStarted: make(chan struct{}, 1),
		stopRelease: stopRelease,
		stopErr:     stopErr,
	}
}

func (host *proofHost) Snapshot() runtimeconfig.Snapshot { return runtimeconfig.Snapshot{} }
func (host *proofHost) RuntimeContext() context.Context  { return context.Background() }
func (host *proofHost) Build() error                     { return nil }
func (host *proofHost) Start(context.Context) error      { return nil }
func (host *proofHost) Running() bool                    { return true }
func (host *proofHost) Ready() bool                      { return true }
func (host *proofHost) CanAccept() bool                  { return true }

func (host *proofHost) Stop(context.Context) error {
	host.stopCalls.Add(1)
	host.stopStarted <- struct{}{}
	if host.onStop != nil {
		host.onStop()
	}
	if host.stopRelease != nil {
		return <-host.stopRelease
	}
	return host.stopErr
}

func mustOwner(t *testing.T, launch launchFunction) *Owner {
	t.Helper()
	if launch == nil {
		launch = func(*runtime.BootstrapRequest) launchResult {
			return launchResult{host: newProofHost(nil, nil), success: true}
		}
	}
	owner, err := newOwner(
		11,
		22,
		"runtime-1",
		sequentialAttemptSource(),
		&runtime.DependencyBindings{},
		launch,
	)
	if err != nil {
		t.Fatalf("newOwner() error = %v", err)
	}
	return owner
}

func sequentialAttemptSource() LaunchAttemptIDSource {
	var next atomic.Int32
	return func() (runtimeconfigload.LaunchAttemptID, error) {
		return runtimeconfigload.LaunchAttemptID(
			fmt.Sprintf("attempt-%d", next.Add(1)),
		), nil
	}
}

func mustPrepare(t *testing.T, owner *Owner) LaunchPreparation {
	t.Helper()
	return mustPrepareRequest(t, owner, NewStartRequest(11, 22, 33))
}

func mustPrepareRequest(
	t *testing.T,
	owner *Owner,
	request StartRequest,
) LaunchPreparation {
	t.Helper()
	preparation, err := owner.PrepareStart(request)
	if err != nil {
		t.Fatalf("PrepareStart() error = %v", err)
	}
	return preparation
}

func mustStart(t *testing.T, owner *Owner, preparation LaunchPreparation) StartOutcome {
	t.Helper()
	outcome, err := owner.Start(
		context.Background(),
		preparation,
		PreparedSnapshot(mustSnapshot(t, preparation)),
	)
	if err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	return outcome
}

func mustSnapshot(t *testing.T, preparation LaunchPreparation) runtimeconfig.Snapshot {
	t.Helper()
	request := preparation.LoadRequest()
	version := configurationversion.ConfigurationVersion{
		ID:              request.ConfigurationVersionID(),
		ConfigurationID: request.ConfigurationID(),
		Number:          1,
		State:           configurationversion.Published,
		Listener: configurationversion.ListenerSettings{
			Host: "127.0.0.1",
			Port: 8080,
			TLS: configurationversion.TLSSettings{
				MinVersion: "1.2",
			},
			Timeouts: configurationversion.TimeoutSettings{
				HandshakeSeconds: 10,
				ReadSeconds:      30,
				WriteSeconds:     20,
				IdleSeconds:      60,
			},
		},
	}
	input := runtimeconfigload.NewDetachedLoadResult(
		request,
		version,
		version.Number,
		true,
		"uwp.configuration",
		1,
	)
	snapshot, diagnostics := runtimeconfig.NewBuilder().Build(input)
	if len(diagnostics) != 0 {
		t.Fatalf("Build() diagnostics = %#v", diagnostics)
	}
	return snapshot
}

func startupFailureOutcome(
	t *testing.T,
	snapshot runtimeconfig.Snapshot,
) runtime.BootstrapOutcome {
	t.Helper()
	resolver, err := secretresolver.NewMemory(nil)
	if err != nil {
		t.Fatalf("NewMemory() error = %v", err)
	}
	reservation, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("reserve listener address: %v", err)
	}
	defer reservation.Close()
	_, portText, err := net.SplitHostPort(reservation.Addr().String())
	if err != nil {
		t.Fatalf("SplitHostPort(%q) error = %v", reservation.Addr(), err)
	}
	portValue, err := strconv.ParseUint(portText, 10, 16)
	if err != nil {
		t.Fatalf("ParseUint(%q) error = %v", portText, err)
	}
	port := uint16(portValue)
	preparation := snapshot.Provenance()
	request := runtimeconfigload.NewLoadRequest(
		preparation.WorkspaceID,
		preparation.ConfigurationID,
		preparation.ConfigurationVersionID,
		preparation.RuntimeInstanceID,
		preparation.LaunchAttemptID,
	)
	version := configurationversion.ConfigurationVersion{
		ID:              preparation.ConfigurationVersionID,
		ConfigurationID: preparation.ConfigurationID,
		Number:          preparation.ConfigurationVersionNumber,
		State:           configurationversion.Published,
		Listener: configurationversion.ListenerSettings{
			Host: "127.0.0.1",
			Port: port,
			TLS:  configurationversion.TLSSettings{MinVersion: "1.2"},
			Timeouts: configurationversion.TimeoutSettings{
				HandshakeSeconds: 10,
				ReadSeconds:      30,
				WriteSeconds:     20,
				IdleSeconds:      60,
			},
		},
	}
	input := runtimeconfigload.NewDetachedLoadResult(
		request,
		version,
		version.Number,
		true,
		preparation.SchemaIdentity,
		preparation.SchemaVersion,
	)
	startupSnapshot, diagnostics := runtimeconfig.NewBuilder().Build(input)
	if len(diagnostics) != 0 {
		t.Fatalf("Build startup Snapshot diagnostics = %#v", diagnostics)
	}

	outcome := runtime.Launch(&runtime.BootstrapRequest{
		Snapshot:       startupSnapshot,
		StartupContext: context.Background(),
		Dependencies:   &runtime.DependencyBindings{SecretResolver: resolver},
	})
	if _, ok := outcome.StartupFailure(); !ok {
		bootstrapFailure, bootstrapFailed := outcome.BootstrapFailure()
		host, succeeded := outcome.Success()
		t.Fatalf(
			"runtime.Launch() did not produce StartupFailure: BootstrapFailure=%v/%t Success=%v/%t",
			bootstrapFailure,
			bootstrapFailed,
			host,
			succeeded,
		)
	}
	return outcome
}

func waitForActual(t *testing.T, owner *Owner, expected ActualState) {
	t.Helper()
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		if owner.Observe().ActualState() == expected {
			return
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatalf("ActualState() = %q, want %q", owner.Observe().ActualState(), expected)
}

package executionowner

import (
	"context"
	"errors"
	"reflect"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestExplicitStopAndRuntimeCancellationShareOnePrimaryCause(t *testing.T) {
	owner := committedOwner(t)
	callback := installCapturedRuntimeCallback(t, owner, func() error { return nil })

	start := make(chan struct{})
	var ready sync.WaitGroup
	ready.Add(2)
	var calls sync.WaitGroup
	calls.Add(2)
	stopResult := make(chan bool, 1)
	go func() {
		defer calls.Done()
		ready.Done()
		<-start
		stopResult <- owner.RequestStop()
	}()
	go func() {
		defer calls.Done()
		ready.Done()
		<-start
		callback()
	}()
	ready.Wait()
	close(start)
	calls.Wait()

	primary, secondary, _, outstanding := controlState(owner)
	if primary != terminationExplicitStop && primary != terminationRuntimeCanceled {
		t.Fatalf("primary cause = %d, want explicit Stop or Runtime cancellation", primary)
	}
	if got := <-stopResult; got != (primary == terminationExplicitStop) {
		t.Fatalf("RequestStop() = %t with primary cause %d", got, primary)
	}
	loser := terminationExplicitStop
	if primary == terminationExplicitStop {
		loser = terminationRuntimeCanceled
	}
	if !secondary.contains(loser) {
		t.Fatalf("secondary causes = %08b, want losing cause %d", secondary, loser)
	}
	if outstanding != 0 {
		t.Fatalf("outstanding calls = %d, want 0", outstanding)
	}
}

func TestCanceledRootBeforeObservationInstallationIsDetected(t *testing.T) {
	owner := committedOwner(t)
	root, cancel := context.WithCancel(context.Background())
	cancel()

	var registrations atomic.Int32
	observation := runtimeCancellationObservation{
		root: root,
		register: func(func()) (callbackRegistration, error) {
			registrations.Add(1)
			return callbackRegistration{unregister: func() error { return nil }}, nil
		},
	}
	if err := owner.installRuntimeCancellation(observation); err != nil {
		t.Fatalf("installRuntimeCancellation() error = %v", err)
	}

	primary, _, _, outstanding := controlState(owner)
	if primary != terminationRuntimeCanceled {
		t.Fatalf("primary cause = %d, want Runtime cancellation", primary)
	}
	if registrations.Load() != 1 || outstanding != 0 {
		t.Fatalf("registrations/outstanding = %d/%d, want 1/0", registrations.Load(), outstanding)
	}
}

func TestCancellationBetweenRegistrationAndSynchronousCheckIsNotLost(t *testing.T) {
	owner := committedOwner(t)
	root, cancel := context.WithCancel(context.Background())

	observation := runtimeCancellationObservation{
		root: root,
		register: func(func()) (callbackRegistration, error) {
			cancel()
			return callbackRegistration{unregister: func() error { return nil }}, nil
		},
	}
	if err := owner.installRuntimeCancellation(observation); err != nil {
		t.Fatalf("installRuntimeCancellation() error = %v", err)
	}
	primary, _, _, _ := controlState(owner)
	if primary != terminationRuntimeCanceled {
		t.Fatalf("primary cause = %d, want Runtime cancellation", primary)
	}
}

func TestCancellationAfterInstallationUsesSameCausalState(t *testing.T) {
	owner := committedOwner(t)
	root, cancel := context.WithCancel(context.Background())
	executionCanceled := make(chan struct{})
	if err := owner.bindExecutionCancellation(func() { close(executionCanceled) }); err != nil {
		t.Fatalf("bindExecutionCancellation() error = %v", err)
	}
	if err := owner.installRuntimeCancellation(newRuntimeCancellationObservation(root)); err != nil {
		t.Fatalf("installRuntimeCancellation() error = %v", err)
	}

	cancel()
	waitClosed(t, executionCanceled, "execution cancellation")
	if result := owner.unregisterAndDrain(); result.status != callbackCleanupConfirmed {
		t.Fatalf("drain status = %d, want Confirmed", result.status)
	}
	primary, _, _, outstanding := controlState(owner)
	if primary != terminationRuntimeCanceled || outstanding != 0 {
		t.Fatalf("primary/outstanding = %d/%d, want RuntimeCanceled/0", primary, outstanding)
	}
}

func TestRepeatedRuntimeCallbackHasOnePrimaryEffect(t *testing.T) {
	owner := committedOwner(t)
	var cancellations atomic.Int32
	if err := owner.bindExecutionCancellation(func() { cancellations.Add(1) }); err != nil {
		t.Fatalf("bindExecutionCancellation() error = %v", err)
	}
	callback := installCapturedRuntimeCallback(t, owner, func() error { return nil })

	callback()
	callback()

	primary, secondary, _, outstanding := controlState(owner)
	if primary != terminationRuntimeCanceled || secondary.contains(terminationRuntimeCanceled) {
		t.Fatalf("primary/secondary = %d/%08b, want one Runtime cancellation", primary, secondary)
	}
	if cancellations.Load() != 1 || outstanding != 0 {
		t.Fatalf("cancellations/outstanding = %d/%d, want 1/0", cancellations.Load(), outstanding)
	}
}

func TestExplicitStopRecordsCauseBeforeCancelingExecution(t *testing.T) {
	owner := committedOwner(t)
	observedPrimary := make(chan terminationCause, 1)
	if err := owner.bindExecutionCancellation(func() {
		primary, _, _, _ := controlState(owner)
		observedPrimary <- primary
	}); err != nil {
		t.Fatalf("bindExecutionCancellation() error = %v", err)
	}

	if !owner.RequestStop() {
		t.Fatal("RequestStop() = false, want true")
	}
	if primary := <-observedPrimary; primary != terminationExplicitStop {
		t.Fatalf("primary observed by cancellation = %d, want explicit Stop", primary)
	}
}

func TestPrimaryCauseIsImmutableAcrossBoundedTerminationSources(t *testing.T) {
	owner := committedOwner(t)
	copy := *owner
	if !owner.recordTermination(terminationNaturalCompletion) {
		t.Fatal("NaturalCompletion did not establish the empty causal cell")
	}
	if copy.recordTermination(terminationExecutionFailure) {
		t.Fatal("ExecutionFailure replaced the primary cause")
	}
	if copy.recordTermination(terminationRecoveredPanic) {
		t.Fatal("RecoveredPanic replaced the primary cause")
	}
	if owner.RequestStop() {
		t.Fatal("RequestStop() replaced NaturalCompletion")
	}

	primary, secondary, _, outstanding := controlState(owner)
	if primary != terminationNaturalCompletion {
		t.Fatalf("primary cause = %d, want NaturalCompletion", primary)
	}
	for _, cause := range []terminationCause{
		terminationExecutionFailure,
		terminationRecoveredPanic,
		terminationExplicitStop,
	} {
		if !secondary.contains(cause) {
			t.Errorf("secondary causes = %08b, want cause %d", secondary, cause)
		}
	}
	if outstanding != 0 {
		t.Fatalf("outstanding calls = %d, want 0", outstanding)
	}
}

func TestRuntimeCallbackPanicIsContainedAndAccountingReturnsToZero(t *testing.T) {
	owner := committedOwner(t)
	if err := owner.bindExecutionCancellation(func() { panic("private payload") }); err != nil {
		t.Fatalf("bindExecutionCancellation() error = %v", err)
	}
	callback := installCapturedRuntimeCallback(t, owner, func() error { return nil })

	callback()

	primary, secondary, anomalies, outstanding := controlState(owner)
	if primary != terminationRuntimeCanceled {
		t.Fatalf("primary cause = %d, want Runtime cancellation", primary)
	}
	if !secondary.contains(terminationExecutionFailure) {
		t.Fatalf("secondary causes = %08b, want execution failure", secondary)
	}
	if anomalies&callbackAnomalyInvocationPanic == 0 {
		t.Fatalf("callback anomalies = %08b, want invocation panic", anomalies)
	}
	if outstanding != 0 {
		t.Fatalf("outstanding calls = %d, want 0", outstanding)
	}
}

func TestDrainClosesAdmissionAndWaitsForAdmittedCallback(t *testing.T) {
	owner := committedOwner(t)
	callbackEntered := make(chan struct{})
	releaseCallback := make(chan struct{})
	var enterOnce sync.Once
	if err := owner.bindExecutionCancellation(func() {
		enterOnce.Do(func() { close(callbackEntered) })
		<-releaseCallback
	}); err != nil {
		t.Fatalf("bindExecutionCancellation() error = %v", err)
	}
	unregisterCalled := make(chan struct{})
	callback := installCapturedRuntimeCallback(t, owner, func() error {
		close(unregisterCalled)
		return nil
	})

	callbackDone := make(chan struct{})
	go func() {
		defer close(callbackDone)
		callback()
	}()
	waitClosed(t, callbackEntered, "callback entry")

	drainResult := make(chan callbackCleanupResult, 1)
	go func() { drainResult <- owner.unregisterAndDrain() }()
	waitClosed(t, unregisterCalled, "callback unregister")
	waitForOutstandingDrain(t, owner)
	assertNoValue(t, drainResult, "drain while callback is admitted")
	repeatedDrainResult := make(chan callbackCleanupResult, 1)
	go func() { repeatedDrainResult <- owner.unregisterAndDrain() }()
	assertNoValue(t, repeatedDrainResult, "repeated drain while callback is admitted")

	if owner.RequestStop() {
		t.Fatal("RequestStop() after admission closure = true")
	}
	callback()
	_, _, _, outstanding := controlState(owner)
	if outstanding != 1 {
		t.Fatalf("outstanding calls = %d, want the one blocked callback", outstanding)
	}

	close(releaseCallback)
	waitClosed(t, callbackDone, "callback return")
	result := waitResult(t, drainResult, "drain completion")
	repeatedResult := waitResult(t, repeatedDrainResult, "repeated drain completion")
	if result.status != callbackCleanupConfirmed {
		t.Fatalf("drain status = %d, want Confirmed", result.status)
	}
	if repeatedResult != result {
		t.Fatalf("repeated drain result = %+v, want %+v", repeatedResult, result)
	}
	_, _, _, outstanding = controlState(owner)
	if outstanding != 0 {
		t.Fatalf("outstanding calls after drain = %d, want 0", outstanding)
	}
}

func TestUnregisterFailureAndPanicReturnStableUnconfirmedResult(t *testing.T) {
	tests := []struct {
		name       string
		unregister func() error
		anomaly    callbackAnomaly
	}{
		{
			name: "failure",
			unregister: func() error {
				return errors.New("unregister failure")
			},
			anomaly: callbackAnomalyUnregisterFailure,
		},
		{
			name: "panic",
			unregister: func() error {
				panic("private payload")
			},
			anomaly: callbackAnomalyUnregisterPanic,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			owner := committedOwner(t)
			var unregisterCalls atomic.Int32
			installCapturedRuntimeCallback(t, owner, func() error {
				unregisterCalls.Add(1)
				return test.unregister()
			})

			first := owner.unregisterAndDrain()
			second := owner.unregisterAndDrain()
			if first != second || first.status != callbackCleanupUnconfirmed {
				t.Fatalf("drain results = %+v and %+v, want same Unconfirmed value", first, second)
			}
			if first.anomaly != test.anomaly {
				t.Fatalf("drain anomaly = %08b, want %08b", first.anomaly, test.anomaly)
			}
			if unregisterCalls.Load() != 1 {
				t.Fatalf("unregister calls = %d, want 1", unregisterCalls.Load())
			}
			if err := owner.sealControl(); !errors.Is(err, errControlCleanupUnconfirmed) {
				t.Fatalf("sealControl() error = %v, want unconfirmed cleanup", err)
			}
		})
	}
}

func TestSealRequiresConfirmedDrainAndIsSharedByOwnerCopies(t *testing.T) {
	owner := committedOwner(t)
	copy1 := *owner
	copy2 := copy1
	callback := installCapturedRuntimeCallback(t, owner, func() error { return nil })

	if err := owner.sealControl(); !errors.Is(err, errControlCleanupUnconfirmed) {
		t.Fatalf("sealControl() before drain error = %v, want unconfirmed cleanup", err)
	}
	result := copy1.unregisterAndDrain()
	if result.status != callbackCleanupConfirmed {
		t.Fatalf("drain status = %d, want Confirmed", result.status)
	}
	if err := copy2.sealControl(); err != nil {
		t.Fatalf("sealControl() error = %v", err)
	}
	if err := owner.sealControl(); err != nil {
		t.Fatalf("repeated sealControl() error = %v", err)
	}
	if owner.RequestStop() || copy1.RequestStop() || copy2.RequestStop() {
		t.Fatal("copied Stop capability accepted a request after seal")
	}
	callback()
	primary, _, _, outstanding := controlState(owner)
	if primary != terminationNone || outstanding != 0 {
		t.Fatalf("primary/outstanding after sealed callback = %d/%d, want none/0", primary, outstanding)
	}
}

func TestCallbackCleanupResultIsDetachedBoundedScalarState(t *testing.T) {
	owner := committedOwner(t)
	result := owner.unregisterAndDrain()
	copy := result
	if copy != result || result.status != callbackCleanupConfirmed {
		t.Fatalf("copied cleanup result = %+v, want %+v", copy, result)
	}

	resultType := reflect.TypeOf(result)
	for index := 0; index < resultType.NumField(); index++ {
		field := resultType.Field(index)
		switch field.Type.Kind() {
		case reflect.Map, reflect.Pointer, reflect.Slice, reflect.Interface,
			reflect.Func, reflect.Chan:
			t.Fatalf("cleanup result field %q exposes mutable or identity-bearing kind %s", field.Name, field.Type.Kind())
		}
	}
}

func TestConcurrentStopCallbackDrainAndObservationPreserveControlInvariants(t *testing.T) {
	owner := committedOwner(t)
	callback := installCapturedRuntimeCallback(t, owner, func() error { return nil })

	start := make(chan struct{})
	var calls sync.WaitGroup
	const callers = 64
	calls.Add(callers)
	for index := 0; index < callers; index++ {
		go func(index int) {
			defer calls.Done()
			<-start
			switch index % 3 {
			case 0:
				owner.RequestStop()
			case 1:
				callback()
			default:
				owner.StopRequested()
			}
		}(index)
	}
	close(start)
	calls.Wait()

	result := owner.unregisterAndDrain()
	if result.status != callbackCleanupConfirmed {
		t.Fatalf("drain status = %d, want Confirmed", result.status)
	}
	primary, _, _, outstanding := controlState(owner)
	if primary != terminationExplicitStop && primary != terminationRuntimeCanceled {
		t.Fatalf("primary cause = %d, want one control cause", primary)
	}
	if outstanding != 0 {
		t.Fatalf("outstanding calls = %d, want 0", outstanding)
	}
	if owner.RequestStop() {
		t.Fatal("RequestStop() after drain = true")
	}
}

func TestObservationInstallationFailureAttemptsExecutionFailureWithoutSideEffects(t *testing.T) {
	owner := committedOwner(t)
	installErr := errors.New("install failure")
	err := owner.installRuntimeCancellation(runtimeCancellationObservation{
		root: context.Background(),
		register: func(func()) (callbackRegistration, error) {
			return callbackRegistration{}, installErr
		},
	})
	if !errors.Is(err, installErr) {
		t.Fatalf("installRuntimeCancellation() error = %v, want install failure", err)
	}
	primary, _, anomalies, outstanding := controlState(owner)
	if primary != terminationExecutionFailure {
		t.Fatalf("primary cause = %d, want execution failure", primary)
	}
	if anomalies&callbackAnomalyInstallFailure == 0 || outstanding != 0 {
		t.Fatalf("anomalies/outstanding = %08b/%d", anomalies, outstanding)
	}
	if owner.State() != StateCommitted {
		t.Fatalf("Owner state = %d, want Committed", owner.State())
	}
	if result := owner.unregisterAndDrain(); result.status != callbackCleanupUnconfirmed {
		t.Fatalf("drain status = %d, want Unconfirmed", result.status)
	}
}

func TestObservationInstallationPanicIsSanitized(t *testing.T) {
	owner := committedOwner(t)
	err := owner.installRuntimeCancellation(runtimeCancellationObservation{
		root: context.Background(),
		register: func(func()) (callbackRegistration, error) {
			panic("private payload")
		},
	})
	if err == nil {
		t.Fatal("installRuntimeCancellation() error = nil after registration panic")
	}
	primary, _, anomalies, outstanding := controlState(owner)
	if primary != terminationExecutionFailure {
		t.Fatalf("primary cause = %d, want execution failure", primary)
	}
	if anomalies&callbackAnomalyInstallPanic == 0 || outstanding != 0 {
		t.Fatalf("anomalies/outstanding = %08b/%d", anomalies, outstanding)
	}
	if result := owner.unregisterAndDrain(); result.status != callbackCleanupUnconfirmed {
		t.Fatalf("drain status = %d, want Unconfirmed", result.status)
	}
}

func committedOwner(t *testing.T) *Owner {
	t.Helper()
	owner := New()
	if err := owner.Transition(StatePreCommit, StateCommitted); err != nil {
		t.Fatalf("commit Transition() error = %v", err)
	}
	return owner
}

func installCapturedRuntimeCallback(
	t *testing.T,
	owner *Owner,
	unregister func() error,
) func() {
	t.Helper()
	var callback func()
	err := owner.installRuntimeCancellation(runtimeCancellationObservation{
		root: context.Background(),
		register: func(installed func()) (callbackRegistration, error) {
			callback = installed
			return callbackRegistration{unregister: unregister}, nil
		},
	})
	if err != nil {
		t.Fatalf("installRuntimeCancellation() error = %v", err)
	}
	if callback == nil {
		t.Fatal("Runtime callback was not registered")
	}
	return callback
}

func controlState(owner *Owner) (
	terminationCause,
	terminationSet,
	callbackAnomaly,
	uint64,
) {
	state := owner.state
	state.mu.RLock()
	defer state.mu.RUnlock()
	return state.control.primary,
		state.control.secondary,
		state.control.anomalies,
		state.control.outstanding
}

func waitClosed(t *testing.T, done <-chan struct{}, description string) {
	t.Helper()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatalf("timed out waiting for %s", description)
	}
}

func waitResult(
	t *testing.T,
	result <-chan callbackCleanupResult,
	description string,
) callbackCleanupResult {
	t.Helper()
	select {
	case value := <-result:
		return value
	case <-time.After(5 * time.Second):
		t.Fatalf("timed out waiting for %s", description)
		return callbackCleanupResult{}
	}
}

func assertNoValue(
	t *testing.T,
	result <-chan callbackCleanupResult,
	description string,
) {
	t.Helper()
	select {
	case value := <-result:
		t.Fatalf("%s returned early: %+v", description, value)
	default:
	}
}

func waitForOutstandingDrain(t *testing.T, owner *Owner) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for {
		state := owner.state
		state.mu.RLock()
		waiting := state.control.outstanding > 0 && state.control.outstandingDone != nil
		state.mu.RUnlock()
		if waiting {
			return
		}
		if time.Now().After(deadline) {
			t.Fatal("timed out waiting for drain to observe the admitted callback")
		}
		runtime.Gosched()
	}
}

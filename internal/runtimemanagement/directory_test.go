package runtimemanagement

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"

	"github.com/dsdred/universal-websocket-platform/internal/configurationloader"
	"github.com/dsdred/universal-websocket-platform/internal/runtime"
	"github.com/dsdred/universal-websocket-platform/internal/runtimeconfigload"
	"github.com/dsdred/universal-websocket-platform/internal/runtimelifecycle"
)

const (
	testWorkspaceID     uint64 = 11
	testConfigurationID uint64 = 22
	testVersionID       uint64 = 33
)

type unavailableSource struct {
	calls atomic.Int32
}

func (s *unavailableSource) LoadExact(uint64, uint64, uint64) (configurationloader.SourceObservation, error) {
	s.calls.Add(1)
	return configurationloader.SourceObservation{}, configurationloader.ErrSourceUnavailable
}

func TestTargetValidationAndAccessors(t *testing.T) {
	for _, input := range []struct {
		workspaceID, configurationID uint64
		instanceID                   runtimeconfigload.RuntimeInstanceID
	}{
		{0, testConfigurationID, "runtime-1"},
		{testWorkspaceID, 0, "runtime-1"},
		{testWorkspaceID, testConfigurationID, ""},
	} {
		if _, err := NewTarget(input.workspaceID, input.configurationID, input.instanceID); err != ErrInvalidTarget {
			t.Fatalf("NewTarget(%d, %d, %q) error = %v, want bare ErrInvalidTarget", input.workspaceID, input.configurationID, input.instanceID, err)
		}
	}

	target := mustTarget(t, testWorkspaceID, testConfigurationID, "runtime-1")
	if target.WorkspaceID() != testWorkspaceID ||
		target.ConfigurationID() != testConfigurationID ||
		target.RuntimeInstanceID() != "runtime-1" {
		t.Fatalf("Target accessors = (%d, %d, %q), want exact identities", target.WorkspaceID(), target.ConfigurationID(), target.RuntimeInstanceID())
	}
}

func TestSentinelStringsAndActionsAreExact(t *testing.T) {
	wantErrors := map[error]string{
		ErrInvalidBinding:          "invalid Runtime management binding",
		ErrInvalidDirectory:        "invalid Runtime management directory",
		ErrInvalidContext:          "invalid Runtime management context",
		ErrInvalidTarget:           "invalid Runtime management target",
		ErrRuntimeInstanceNotFound: "Runtime Instance not found",
	}
	for sentinel, want := range wantErrors {
		if sentinel.Error() != want {
			t.Fatalf("sentinel string = %q, want %q", sentinel.Error(), want)
		}
	}
	if ActionStart != "start" || ActionStop != "stop" || ActionObserve != "observe" {
		t.Fatalf("actions = %q/%q/%q, want exact strings", ActionStart, ActionStop, ActionObserve)
	}
}

func TestBindingRequiresExactOwnerIdentity(t *testing.T) {
	target := mustTarget(t, testWorkspaceID, testConfigurationID, "runtime-1")
	owner := mustOwner(t, testWorkspaceID, testConfigurationID, "runtime-1")
	loader := configurationloader.New(&unavailableSource{})

	for name, candidate := range map[string]func() (Binding, error){
		"nil owner":  func() (Binding, error) { return NewBinding(target, nil, loader) },
		"nil loader": func() (Binding, error) { return NewBinding(target, owner, nil) },
		"workspace mismatch": func() (Binding, error) {
			return NewBinding(target, mustOwner(t, 99, testConfigurationID, "runtime-1"), loader)
		},
		"configuration mismatch": func() (Binding, error) {
			return NewBinding(target, mustOwner(t, testWorkspaceID, 99, "runtime-1"), loader)
		},
		"instance mismatch": func() (Binding, error) {
			return NewBinding(target, mustOwner(t, testWorkspaceID, testConfigurationID, "runtime-2"), loader)
		},
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := candidate(); err != ErrInvalidBinding {
				t.Fatalf("NewBinding() error = %v, want bare ErrInvalidBinding", err)
			}
		})
	}
	if _, err := NewBinding(target, owner, loader); err != nil {
		t.Fatalf("NewBinding(valid) error = %v", err)
	}
}

func TestDirectoryRejectsInvalidAndDuplicateBindings(t *testing.T) {
	target := mustTarget(t, testWorkspaceID, testConfigurationID, "runtime-1")
	owner := mustOwner(t, testWorkspaceID, testConfigurationID, "runtime-1")
	binding := mustBinding(t, target, owner, configurationloader.New(&unavailableSource{}))
	authorize := Authorize(func(context.Context, Action, Target, uint64) error { return nil })

	for name, construct := range map[string]func() (*Directory, error){
		"nil authorize":      func() (*Directory, error) { return NewDirectory(nil, binding) },
		"no bindings":        func() (*Directory, error) { return NewDirectory(authorize) },
		"zero binding":       func() (*Directory, error) { return NewDirectory(authorize, Binding{}) },
		"duplicate instance": func() (*Directory, error) { return NewDirectory(authorize, binding, binding) },
		"duplicate owner": func() (*Directory, error) {
			other := mustTarget(t, testWorkspaceID, testConfigurationID, "runtime-2")
			return NewDirectory(authorize, binding, Binding{target: other, owner: owner, loader: configurationloader.New(&unavailableSource{})})
		},
	} {
		t.Run(name, func(t *testing.T) {
			if directory, err := construct(); directory != nil || err != ErrInvalidDirectory {
				t.Fatalf("NewDirectory() = %#v/%v, want nil/bare ErrInvalidDirectory", directory, err)
			}
		})
	}
}

func TestCommandValidationPrecedenceSkipsAuthorizationAndLifecycle(t *testing.T) {
	var authorizationCalls atomic.Int32
	var attemptCalls atomic.Int32
	target := mustTarget(t, testWorkspaceID, testConfigurationID, "runtime-1")
	source := &unavailableSource{}
	owner := mustOwnerWithCounter(t, testWorkspaceID, testConfigurationID, "runtime-1", &attemptCalls)
	directory := mustDirectory(t, func(context.Context, Action, Target, uint64) error {
		authorizationCalls.Add(1)
		return nil
	}, mustBinding(t, target, owner, configurationloader.New(source)))

	cancelled, cancel := context.WithCancel(context.Background())
	cancel()
	mismatched := mustTarget(t, 99, testConfigurationID, "runtime-1")
	missing := mustTarget(t, testWorkspaceID, testConfigurationID, "runtime-missing")
	var nilDirectory *Directory

	cases := []struct {
		name string
		call func() error
		want error
	}{
		{"invalid directory before context", func() error { _, err := nilDirectory.Start(nil, Target{}, 0); return err }, ErrInvalidDirectory},
		{"nil context before target", func() error { _, err := directory.Start(nil, Target{}, 0); return err }, ErrInvalidContext},
		{"invalid target before cancellation", func() error { _, err := directory.Start(cancelled, Target{}, 0); return err }, ErrInvalidTarget},
		{"zero version", func() error { _, err := directory.Start(context.Background(), target, 0); return err }, ErrInvalidTarget},
		{"cancelled before lookup", func() error { _, err := directory.Stop(cancelled, missing); return err }, context.Canceled},
		{"missing", func() error { _, err := directory.Observe(context.Background(), missing); return err }, ErrRuntimeInstanceNotFound},
		{"mismatch normalized", func() error { _, err := directory.Observe(context.Background(), mismatched); return err }, ErrRuntimeInstanceNotFound},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			if err := test.call(); err != test.want {
				t.Fatalf("error = %v, want exact %v", err, test.want)
			}
		})
	}
	if authorizationCalls.Load() != 0 || attemptCalls.Load() != 0 || source.calls.Load() != 0 {
		t.Fatalf("validation performed work: authorize=%d attempts=%d source=%d", authorizationCalls.Load(), attemptCalls.Load(), source.calls.Load())
	}
}

func TestStartAuthorizesExactScopeThenDelegatesToFlow(t *testing.T) {
	target := mustTarget(t, testWorkspaceID, testConfigurationID, "runtime-start")
	source := &unavailableSource{}
	owner := mustOwner(t, testWorkspaceID, testConfigurationID, "runtime-start")
	var calls atomic.Int32
	directory := mustDirectory(t, func(_ context.Context, action Action, got Target, version uint64) error {
		calls.Add(1)
		if action != ActionStart || got != target || version != testVersionID {
			t.Fatalf("Authorize() = (%q, %#v, %d), want exact Start scope", action, got, version)
		}
		return nil
	}, mustBinding(t, target, owner, configurationloader.New(source)))

	outcome, err := directory.Start(context.Background(), target, testVersionID)
	if err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	if outcome.Kind() != runtimelifecycle.StartPreparationFailed {
		t.Fatalf("Start() kind = %q, want %q", outcome.Kind(), runtimelifecycle.StartPreparationFailed)
	}
	failure, ok := outcome.PreparationFailure()
	if !ok || failure != configurationloader.ErrSourceUnavailable {
		t.Fatalf("PreparationFailure() = %v/%t, want exact ErrSourceUnavailable", failure, ok)
	}
	if calls.Load() != 1 || source.calls.Load() != 1 {
		t.Fatalf("calls authorize/source = %d/%d, want 1/1", calls.Load(), source.calls.Load())
	}
	attempt := outcome.Attempt()
	if attempt.WorkspaceID() != testWorkspaceID || attempt.ConfigurationID() != testConfigurationID ||
		attempt.ConfigurationVersionID() != testVersionID || attempt.RuntimeInstanceID() != target.RuntimeInstanceID() {
		t.Fatalf("Start attempt = %#v, want exact routed identities", attempt)
	}
}

func TestAuthorizationErrorsArePreservedWithoutMutation(t *testing.T) {
	denied := errors.New("denied")
	for _, action := range []Action{ActionStart, ActionStop, ActionObserve} {
		t.Run(string(action), func(t *testing.T) {
			var attemptCalls atomic.Int32
			target := mustTarget(t, testWorkspaceID, testConfigurationID, "runtime-"+runtimeconfigload.RuntimeInstanceID(action))
			source := &unavailableSource{}
			owner := mustOwnerWithCounter(t, testWorkspaceID, testConfigurationID, target.RuntimeInstanceID(), &attemptCalls)
			var authorizationCalls atomic.Int32
			directory := mustDirectory(t, func(_ context.Context, got Action, routed Target, version uint64) error {
				authorizationCalls.Add(1)
				if got != action || routed != target || version != map[Action]uint64{ActionStart: testVersionID}[action] {
					t.Fatalf("Authorize() = (%q, %#v, %d), unexpected", got, routed, version)
				}
				return denied
			}, mustBinding(t, target, owner, configurationloader.New(source)))

			var err error
			switch action {
			case ActionStart:
				_, err = directory.Start(context.Background(), target, testVersionID)
			case ActionStop:
				_, err = directory.Stop(context.Background(), target)
			case ActionObserve:
				_, err = directory.Observe(context.Background(), target)
			}
			if err != denied {
				t.Fatalf("command error = %v, want exact authorization error", err)
			}
			if authorizationCalls.Load() != 1 || attemptCalls.Load() != 0 || source.calls.Load() != 0 {
				t.Fatalf("denial performed work: authorize=%d attempts=%d source=%d", authorizationCalls.Load(), attemptCalls.Load(), source.calls.Load())
			}
			observation := owner.Observe()
			if observation.DesiredState() != runtimelifecycle.DesiredStopped || observation.ActualState() != runtimelifecycle.ActualStopped {
				t.Fatalf("Owner state mutated after denial: %#v", observation)
			}
		})
	}
}

func TestObserveChecksCancellationAfterAuthorization(t *testing.T) {
	target := mustTarget(t, testWorkspaceID, testConfigurationID, "runtime-observe")
	ctx, cancel := context.WithCancel(context.Background())
	directory := mustDirectory(t, func(got context.Context, action Action, routed Target, version uint64) error {
		if got != ctx || action != ActionObserve || routed != target || version != 0 {
			t.Fatalf("Authorize() received unexpected arguments")
		}
		cancel()
		return nil
	}, mustBinding(t, target, mustOwner(t, testWorkspaceID, testConfigurationID, "runtime-observe"), configurationloader.New(&unavailableSource{})))

	if _, err := directory.Observe(ctx, target); err != context.Canceled {
		t.Fatalf("Observe() error = %v, want exact context.Canceled", err)
	}
}

func TestDownstreamCancellationErrorsArePreserved(t *testing.T) {
	for _, action := range []Action{ActionStart, ActionStop} {
		t.Run(string(action), func(t *testing.T) {
			var attemptCalls atomic.Int32
			target := mustTarget(t, testWorkspaceID, testConfigurationID, "runtime-downstream-"+runtimeconfigload.RuntimeInstanceID(action))
			source := &unavailableSource{}
			owner := mustOwnerWithCounter(t, testWorkspaceID, testConfigurationID, target.RuntimeInstanceID(), &attemptCalls)
			ctx, cancel := context.WithCancel(context.Background())
			directory := mustDirectory(t, func(got context.Context, gotAction Action, _ Target, _ uint64) error {
				if got != ctx || gotAction != action {
					t.Fatalf("Authorize() received unexpected context/action")
				}
				cancel()
				return nil
			}, mustBinding(t, target, owner, configurationloader.New(source)))

			var err error
			if action == ActionStart {
				_, err = directory.Start(ctx, target, testVersionID)
			} else {
				_, err = directory.Stop(ctx, target)
			}
			if err != context.Canceled {
				t.Fatalf("command error = %v, want exact downstream context.Canceled", err)
			}
			if attemptCalls.Load() != 0 || source.calls.Load() != 0 {
				t.Fatalf("cancelled downstream gate performed work: attempts=%d source=%d", attemptCalls.Load(), source.calls.Load())
			}
		})
	}
}

func TestCancellationDuringAuthorizationReturnsWithoutDetachedWork(t *testing.T) {
	var attemptCalls atomic.Int32
	target := mustTarget(t, testWorkspaceID, testConfigurationID, "runtime-authorize-cancel")
	source := &unavailableSource{}
	owner := mustOwnerWithCounter(t, testWorkspaceID, testConfigurationID, target.RuntimeInstanceID(), &attemptCalls)
	entered := make(chan struct{})
	exited := make(chan struct{})
	directory := mustDirectory(t, func(ctx context.Context, _ Action, _ Target, _ uint64) error {
		close(entered)
		<-ctx.Done()
		close(exited)
		return ctx.Err()
	}, mustBinding(t, target, owner, configurationloader.New(source)))
	ctx, cancel := context.WithCancel(context.Background())
	result := make(chan error, 1)
	go func() {
		_, err := directory.Start(ctx, target, testVersionID)
		result <- err
	}()

	<-entered
	cancel()
	if err := <-result; err != context.Canceled {
		t.Fatalf("Start() error = %v, want exact context.Canceled", err)
	}
	select {
	case <-exited:
	default:
		t.Fatal("Authorize() work remained detached after command return")
	}
	if attemptCalls.Load() != 0 || source.calls.Load() != 0 {
		t.Fatalf("authorization cancellation performed work: attempts=%d source=%d", attemptCalls.Load(), source.calls.Load())
	}
}

func TestStopAndObserveAuthorizeExactScope(t *testing.T) {
	target := mustTarget(t, testWorkspaceID, testConfigurationID, "runtime-commands")
	owner := mustOwner(t, testWorkspaceID, testConfigurationID, target.RuntimeInstanceID())
	wantActions := []Action{ActionObserve, ActionStop}
	var next atomic.Int32
	directory := mustDirectory(t, func(_ context.Context, action Action, routed Target, version uint64) error {
		index := int(next.Add(1)) - 1
		if index >= len(wantActions) || action != wantActions[index] || routed != target || version != 0 {
			t.Fatalf("Authorize[%d]() = (%q, %#v, %d), want exact scope", index, action, routed, version)
		}
		return nil
	}, mustBinding(t, target, owner, configurationloader.New(&unavailableSource{})))

	observation, err := directory.Observe(context.Background(), target)
	if err != nil || observation.RuntimeInstanceID() != target.RuntimeInstanceID() {
		t.Fatalf("Observe() = %q/%v, want exact scope/nil", observation.RuntimeInstanceID(), err)
	}
	stop, err := directory.Stop(context.Background(), target)
	if err != nil || stop.Kind() != runtimelifecycle.StopStopped {
		t.Fatalf("Stop() = %q/%v, want StopStopped/nil", stop.Kind(), err)
	}
	if next.Load() != 2 {
		t.Fatalf("Authorize() calls = %d, want 2", next.Load())
	}
}

func TestAuthorizationPanicPropagatesBeforeLifecycle(t *testing.T) {
	var attemptCalls atomic.Int32
	target := mustTarget(t, testWorkspaceID, testConfigurationID, "runtime-panic")
	source := &unavailableSource{}
	owner := mustOwnerWithCounter(t, testWorkspaceID, testConfigurationID, target.RuntimeInstanceID(), &attemptCalls)
	directory := mustDirectory(t, func(context.Context, Action, Target, uint64) error {
		panic("policy defect")
	}, mustBinding(t, target, owner, configurationloader.New(source)))

	deferred := false
	func() {
		defer func() {
			if recovered := recover(); recovered != "policy defect" {
				t.Fatalf("recovered panic = %#v, want exact policy defect", recovered)
			}
			deferred = true
		}()
		_, _ = directory.Start(context.Background(), target, testVersionID)
	}()
	if !deferred || attemptCalls.Load() != 0 || source.calls.Load() != 0 {
		t.Fatalf("panic handling mutated lifecycle: recovered=%t attempts=%d source=%d", deferred, attemptCalls.Load(), source.calls.Load())
	}
}

func TestBlockedAuthorizationDoesNotBlockAnotherScope(t *testing.T) {
	first := mustTarget(t, testWorkspaceID, testConfigurationID, "runtime-blocked")
	second := mustTarget(t, testWorkspaceID, testConfigurationID, "runtime-independent")
	entered := make(chan struct{})
	release := make(chan struct{})
	directory := mustDirectory(t, func(_ context.Context, _ Action, target Target, _ uint64) error {
		if target == first {
			close(entered)
			<-release
		}
		return nil
	},
		mustBinding(t, first, mustOwner(t, testWorkspaceID, testConfigurationID, first.RuntimeInstanceID()), configurationloader.New(&unavailableSource{})),
		mustBinding(t, second, mustOwner(t, testWorkspaceID, testConfigurationID, second.RuntimeInstanceID()), configurationloader.New(&unavailableSource{})),
	)

	firstResult := make(chan error, 1)
	go func() {
		_, err := directory.Observe(context.Background(), first)
		firstResult <- err
	}()
	<-entered
	if observation, err := directory.Observe(context.Background(), second); err != nil || observation.RuntimeInstanceID() != second.RuntimeInstanceID() {
		t.Fatalf("independent Observe() = %q/%v, want second scope/nil", observation.RuntimeInstanceID(), err)
	}
	close(release)
	if err := <-firstResult; err != nil {
		t.Fatalf("blocked Observe() error = %v", err)
	}
}

func TestAuthorizeCanRunConcurrentlyForSameAndDifferentTargets(t *testing.T) {
	first := mustTarget(t, testWorkspaceID, testConfigurationID, "runtime-concurrent-1")
	second := mustTarget(t, testWorkspaceID, testConfigurationID, "runtime-concurrent-2")
	entered := make(chan struct{}, 3)
	release := make(chan struct{})
	var active atomic.Int32
	var maximum atomic.Int32
	directory := mustDirectory(t, func(_ context.Context, _ Action, _ Target, _ uint64) error {
		current := active.Add(1)
		for previous := maximum.Load(); current > previous && !maximum.CompareAndSwap(previous, current); previous = maximum.Load() {
		}
		entered <- struct{}{}
		<-release
		active.Add(-1)
		return nil
	},
		mustBinding(t, first, mustOwner(t, testWorkspaceID, testConfigurationID, first.RuntimeInstanceID()), configurationloader.New(&unavailableSource{})),
		mustBinding(t, second, mustOwner(t, testWorkspaceID, testConfigurationID, second.RuntimeInstanceID()), configurationloader.New(&unavailableSource{})),
	)

	results := make(chan error, 3)
	for _, target := range []Target{first, first, second} {
		go func(target Target) {
			_, err := directory.Observe(context.Background(), target)
			results <- err
		}(target)
	}
	for range 3 {
		<-entered
	}
	if maximum.Load() != 3 {
		t.Fatalf("maximum concurrent Authorize() calls = %d, want 3", maximum.Load())
	}
	close(release)
	for range 3 {
		if err := <-results; err != nil {
			t.Fatalf("Observe() error = %v", err)
		}
	}
	if active.Load() != 0 {
		t.Fatalf("active Authorize() calls after return = %d, want 0", active.Load())
	}
}

func mustTarget(t *testing.T, workspaceID, configurationID uint64, instanceID runtimeconfigload.RuntimeInstanceID) Target {
	t.Helper()
	target, err := NewTarget(workspaceID, configurationID, instanceID)
	if err != nil {
		t.Fatalf("NewTarget() error = %v", err)
	}
	return target
}

func mustOwner(t *testing.T, workspaceID, configurationID uint64, instanceID runtimeconfigload.RuntimeInstanceID) *runtimelifecycle.Owner {
	t.Helper()
	return mustOwnerWithCounter(t, workspaceID, configurationID, instanceID, nil)
}

func mustOwnerWithCounter(t *testing.T, workspaceID, configurationID uint64, instanceID runtimeconfigload.RuntimeInstanceID, calls *atomic.Int32) *runtimelifecycle.Owner {
	t.Helper()
	owner, err := runtimelifecycle.NewOwner(workspaceID, configurationID, instanceID, func() (runtimeconfigload.LaunchAttemptID, error) {
		if calls != nil {
			calls.Add(1)
		}
		return runtimeconfigload.LaunchAttemptID("attempt-" + instanceID), nil
	}, &runtime.DependencyBindings{})
	if err != nil {
		t.Fatalf("NewOwner() error = %v", err)
	}
	return owner
}

func mustBinding(t *testing.T, target Target, owner *runtimelifecycle.Owner, loader *configurationloader.Loader) Binding {
	t.Helper()
	binding, err := NewBinding(target, owner, loader)
	if err != nil {
		t.Fatalf("NewBinding() error = %v", err)
	}
	return binding
}

func mustDirectory(t *testing.T, authorize Authorize, bindings ...Binding) *Directory {
	t.Helper()
	directory, err := NewDirectory(authorize, bindings...)
	if err != nil {
		t.Fatalf("NewDirectory() error = %v", err)
	}
	return directory
}

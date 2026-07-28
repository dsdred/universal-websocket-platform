package runtime

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/dsdred/universal-websocket-platform/internal/authentication"
	"github.com/dsdred/universal-websocket-platform/internal/configurationversion"
)

func TestLaunchReturnsActiveHost(t *testing.T) {
	request := validBootstrapProofRequest(t)

	outcome := Launch(request)

	host, ok := outcome.Success()
	if !ok || host == nil {
		t.Fatalf("Launch() = %#v, want Success", outcome)
	}
	t.Cleanup(func() {
		if err := host.Stop(context.Background()); err != nil {
			t.Errorf("cleanup Stop() error = %v", err)
		}
	})
	if !host.Running() || !host.Ready() || !host.CanAccept() {
		t.Fatal("Launch() did not return the active Host published by Bootstrap")
	}
	assertExclusiveBootstrapOutcome(t, outcome, bootstrapOutcomeSuccess)
}

func TestLaunchReturnsBootstrapFailure(t *testing.T) {
	outcome := Launch(nil)

	failure, ok := outcome.BootstrapFailure()
	if !ok || failure == nil {
		t.Fatalf("Launch(nil) = %#v, want BootstrapFailure", outcome)
	}
	if failure.Stage() != BootstrapStageInputValidation ||
		failure.Code() != BootstrapCodeInvalidStartupContext {
		t.Fatalf(
			"Launch(nil) failure = %q/%q, want %q/%q",
			failure.Stage(),
			failure.Code(),
			BootstrapStageInputValidation,
			BootstrapCodeInvalidStartupContext,
		)
	}
	assertExclusiveBootstrapOutcome(t, outcome, bootstrapOutcomeBootstrapFailure)
}

func TestLaunchReturnsStartupFailureCause(t *testing.T) {
	request := &BootstrapRequest{
		Snapshot: snapshotWithAuthentication(t, configurationversion.AuthenticationSettings{
			Enabled: true,
			Providers: []configurationversion.AuthenticationProvider{{
				Name:    "basic",
				Type:    configurationversion.AuthenticationProviderBasic,
				Enabled: true,
				Basic: &configurationversion.BasicSettings{
					Realm:     "Universal WebSocket Platform",
					SecretRef: "secrets/basic/main",
				},
			}},
		}),
		StartupContext: context.Background(),
		Dependencies: &DependencyBindings{
			SecretResolver: emptyResolver(t),
		},
	}

	outcome := Launch(request)

	failure, ok := outcome.StartupFailure()
	if !ok || failure == nil {
		t.Fatalf("Launch() = %#v, want StartupFailure", outcome)
	}
	if !errors.Is(failure, authentication.ErrFactoryNotFound) {
		t.Fatalf("Launch() StartupFailure = %v, want ErrFactoryNotFound", failure)
	}
	if host, success := outcome.Success(); success || host != nil {
		t.Fatal("Launch() published a Host for StartupFailure")
	}
	assertExclusiveBootstrapOutcome(t, outcome, bootstrapOutcomeStartupFailure)
}

func TestLaunchConcurrentCallsAreIndependent(t *testing.T) {
	const invocations = 4

	var wait sync.WaitGroup
	outcomes := make(chan BootstrapOutcome, invocations)
	for range invocations {
		request := validBootstrapProofRequest(t)
		wait.Add(1)
		go func() {
			defer wait.Done()
			outcomes <- Launch(request)
		}()
	}
	wait.Wait()
	close(outcomes)

	hosts := make(map[Host]struct{}, invocations)
	for outcome := range outcomes {
		host, ok := outcome.Success()
		if !ok || host == nil {
			t.Fatalf("concurrent Launch() = %#v, want Success", outcome)
		}
		if _, duplicate := hosts[host]; duplicate {
			t.Fatal("concurrent Launch() calls returned the same Host")
		}
		hosts[host] = struct{}{}
		t.Cleanup(func() {
			if err := host.Stop(context.Background()); err != nil {
				t.Errorf("cleanup Stop() error = %v", err)
			}
		})
	}
}

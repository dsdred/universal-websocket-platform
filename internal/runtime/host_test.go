package runtime

import (
	"context"
	"errors"
	"net"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/dsdred/universal-websocket-platform/internal/runtimeconfig"
	"github.com/dsdred/universal-websocket-platform/internal/secretresolver"
)

func TestDefaultHostImplementsHost(t *testing.T) {
	var _ Host = (*DefaultHost)(nil)
}

func TestNewHostCreatesHostAndContainer(t *testing.T) {
	snapshot := validSnapshot()
	host, err := NewHost(snapshot, emptyResolver(t), nil)
	if err != nil {
		t.Fatalf("NewHost() error = %v", err)
	}
	if host.container == nil {
		t.Fatal("NewHost() Container = nil")
	}
	if host.Running() {
		t.Fatal("new Host is Running, want Created")
	}
	got := host.Snapshot()
	if got.ConfigurationID != snapshot.ConfigurationID || got.VersionID != snapshot.VersionID {
		t.Fatalf("Snapshot identifiers = (%d, %d)", got.ConfigurationID, got.VersionID)
	}
}

func TestNewHostRejectsZeroSnapshot(t *testing.T) {
	host, err := NewHost(runtimeconfig.Snapshot{}, emptyResolver(t), nil)
	if host != nil || !errors.Is(err, ErrNilSnapshot) {
		t.Fatalf("NewHost() = (%v, %v), want nil and ErrNilSnapshot", host, err)
	}
}

func TestNewHostRejectsNilResolver(t *testing.T) {
	host, err := NewHost(validSnapshot(), nil, nil)
	if host != nil || !errors.Is(err, ErrNilSecretResolver) {
		t.Fatalf("NewHost() = (%v, %v), want nil and ErrNilSecretResolver", host, err)
	}
}

func TestHostSnapshotIsDeepCopy(t *testing.T) {
	snapshot := validSnapshot()
	host, err := NewHost(snapshot, emptyResolver(t), nil)
	if err != nil {
		t.Fatalf("NewHost() error = %v", err)
	}

	snapshot.Listener.Host = "changed-source"
	snapshot.Authentication.Providers[0].JWT.SigningKeys[0].SecretRef = "changed-source-key"
	first := host.Snapshot()
	first.Listener.Host = "changed-result"
	first.Authentication.Providers[0].JWT.SigningKeys[0].SecretRef = "changed-result-key"
	first.Authentication.Providers = append(first.Authentication.Providers, runtimeconfig.AuthenticationProviderSnapshot{Name: "new"})

	assertOriginalSnapshot(t, host.Snapshot())
	assertOriginalSnapshot(t, host.container.Snapshot())
}

func TestHostBuildComposesRuntime(t *testing.T) {
	host := newUnbuiltHost(t)
	if err := host.Build(); err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	if host.runtimeListener == nil {
		t.Fatal("Build() Listener = nil")
	}
	if host.state != hostBuilt {
		t.Fatalf("Build() state = %v, want hostBuilt", host.state)
	}
	if err := host.Build(); !errors.Is(err, ErrHostAlreadyBuilt) {
		t.Fatalf("second Build() error = %v, want ErrHostAlreadyBuilt", err)
	}
}

func TestHostStartBeforeBuild(t *testing.T) {
	host := newUnbuiltHost(t)
	if err := host.Start(context.Background()); !errors.Is(err, ErrHostNotBuilt) {
		t.Fatalf("Start() error = %v, want ErrHostNotBuilt", err)
	}
}

func TestHostStartAndStop(t *testing.T) {
	host := mustHost(t)
	if err := host.Start(context.Background()); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	if !host.Running() {
		t.Fatal("Running() = false after Start")
	}
	if err := host.Stop(context.Background()); err != nil {
		t.Fatalf("Stop() error = %v", err)
	}
	if host.Running() {
		t.Fatal("Running() = true after Stop")
	}
}

func TestHostDoubleStart(t *testing.T) {
	host := mustHost(t)
	if err := host.Start(context.Background()); err != nil {
		t.Fatalf("first Start() error = %v", err)
	}
	if err := host.Start(context.Background()); !errors.Is(err, ErrHostAlreadyRunning) {
		t.Fatalf("second Start() error = %v, want ErrHostAlreadyRunning", err)
	}
}

func TestHostDoesNotSupportRestart(t *testing.T) {
	host := mustHost(t)
	if err := host.Start(context.Background()); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	if err := host.Stop(context.Background()); err != nil {
		t.Fatalf("Stop() error = %v", err)
	}
	if err := host.Start(context.Background()); !errors.Is(err, ErrHostAlreadyRunning) {
		t.Fatalf("Start() after Stop error = %v, want ErrHostAlreadyRunning", err)
	}
}

func TestHostStopWithoutStartIsNoOp(t *testing.T) {
	host := mustHost(t)
	if err := host.Stop(context.Background()); err != nil {
		t.Fatalf("Stop() error = %v", err)
	}
	if host.Running() {
		t.Fatal("Running() = true after no-op Stop")
	}
	if err := host.Start(context.Background()); err != nil {
		t.Fatalf("Start() after no-op Stop error = %v", err)
	}
}

func TestHostConcurrentStart(t *testing.T) {
	host := mustHost(t)
	const goroutines = 64
	var successes atomic.Int64
	var alreadyRunning atomic.Int64
	var waitGroup sync.WaitGroup
	errorsChannel := make(chan error, goroutines)
	for range goroutines {
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			err := host.Start(context.Background())
			switch {
			case err == nil:
				successes.Add(1)
			case errors.Is(err, ErrHostAlreadyRunning):
				alreadyRunning.Add(1)
			default:
				errorsChannel <- err
			}
		}()
	}
	waitGroup.Wait()
	close(errorsChannel)
	for err := range errorsChannel {
		t.Errorf("Start() unexpected error = %v", err)
	}
	if successes.Load() != 1 || alreadyRunning.Load() != goroutines-1 {
		t.Fatalf("Start outcomes = (%d success, %d already running)", successes.Load(), alreadyRunning.Load())
	}
	if !host.Running() {
		t.Fatal("Running() = false after concurrent Start")
	}
}

func TestHostConcurrentStop(t *testing.T) {
	host := mustHost(t)
	if err := host.Start(context.Background()); err != nil {
		t.Fatalf("Start() error = %v", err)
	}

	const goroutines = 64
	var waitGroup sync.WaitGroup
	errorsChannel := make(chan error, goroutines)
	for range goroutines {
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			if err := host.Stop(context.Background()); err != nil {
				errorsChannel <- err
			}
		}()
	}
	waitGroup.Wait()
	close(errorsChannel)
	for err := range errorsChannel {
		t.Errorf("Stop() error = %v", err)
	}
	if host.Running() {
		t.Fatal("Running() = true after concurrent Stop")
	}
}

func TestHostSmokeScenario(t *testing.T) {
	host := mustHost(t)
	if err := host.Start(context.Background()); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	t.Logf("Snapshot -> Host -> Start: Running=%t", host.Running())
	if !host.Running() {
		t.Fatal("Running() = false after Start")
	}
	if err := host.Stop(context.Background()); err != nil {
		t.Fatalf("Stop() error = %v", err)
	}
	t.Logf("Host -> Stop: Running=%t", host.Running())
	if host.Running() {
		t.Fatal("Running() = true after Stop")
	}
}

func mustHost(t *testing.T) *DefaultHost {
	t.Helper()
	host := newUnbuiltHost(t)
	if err := host.Build(); err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	t.Cleanup(func() {
		if err := host.Stop(context.Background()); err != nil {
			t.Errorf("cleanup Stop() error = %v", err)
		}
	})
	return host
}

func newUnbuiltHost(t *testing.T) *DefaultHost {
	t.Helper()
	snapshot := validSnapshot()
	snapshot.Listener.Port = availablePort(t)
	snapshot.Authentication.Providers = snapshot.Authentication.Providers[:2]
	host, err := NewHost(snapshot, emptyResolver(t), nil)
	if err != nil {
		t.Fatalf("NewHost() error = %v", err)
	}
	return host
}

func emptyResolver(t *testing.T) secretresolver.Resolver {
	t.Helper()
	resolver, err := secretresolver.NewMemory(nil)
	if err != nil {
		t.Fatalf("NewMemory() error = %v", err)
	}
	return resolver
}

func availablePort(t *testing.T) uint16 {
	t.Helper()
	tcpListener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("reserve port: %v", err)
	}
	port := tcpListener.Addr().(*net.TCPAddr).Port
	if err := tcpListener.Close(); err != nil {
		t.Fatalf("release port: %v", err)
	}
	return uint16(port)
}

package runtime

import (
	"context"
	"errors"
	"sync"

	"github.com/dsdred/universal-websocket-platform/internal/runtimeconfig"
)

var (
	// ErrHostAlreadyRunning indicates that the Host lifecycle has already been started.
	ErrHostAlreadyRunning = errors.New("runtime Host already running")
	// ErrNilSnapshot indicates that NewHost received a zero Snapshot value.
	ErrNilSnapshot = errors.New("runtime Snapshot is nil")
)

// Host owns an immutable Runtime Snapshot and coordinates its lifecycle.
type Host interface {
	Snapshot() runtimeconfig.Snapshot
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
}

type hostState uint8

const (
	hostCreated hostState = iota
	hostRunning
	hostStopped
)

// DefaultHost is the Runtime composition root and lifecycle coordinator.
type DefaultHost struct {
	mu        sync.RWMutex
	snapshot  runtimeconfig.Snapshot
	container Container
	state     hostState
}

// NewHost creates a Runtime Host and its dependency Container from independent Snapshot copies.
func NewHost(snapshot runtimeconfig.Snapshot) (*DefaultHost, error) {
	if isZeroSnapshot(snapshot) {
		return nil, ErrNilSnapshot
	}

	container, err := New(snapshot)
	if err != nil {
		return nil, err
	}

	return &DefaultHost{
		snapshot:  container.Snapshot(),
		container: container,
		state:     hostCreated,
	}, nil
}

// Snapshot returns an independent copy of the Host Snapshot.
func (host *DefaultHost) Snapshot() runtimeconfig.Snapshot {
	host.mu.RLock()
	defer host.mu.RUnlock()
	return cloneSnapshot(host.snapshot)
}

// Start moves a newly created Host to Running without starting Runtime components.
func (host *DefaultHost) Start(_ context.Context) error {
	host.mu.Lock()
	defer host.mu.Unlock()
	if host.state != hostCreated {
		return ErrHostAlreadyRunning
	}
	host.state = hostRunning
	return nil
}

// Stop moves a Running Host to Stopped and is otherwise a no-op.
func (host *DefaultHost) Stop(_ context.Context) error {
	host.mu.Lock()
	defer host.mu.Unlock()
	if host.state == hostRunning {
		host.state = hostStopped
	}
	return nil
}

// Running reports whether the Host is in the Running state.
func (host *DefaultHost) Running() bool {
	host.mu.RLock()
	defer host.mu.RUnlock()
	return host.state == hostRunning
}

func isZeroSnapshot(snapshot runtimeconfig.Snapshot) bool {
	return snapshot.ConfigurationID == 0 &&
		snapshot.VersionID == 0 &&
		snapshot.Listener == (runtimeconfig.ListenerSnapshot{}) &&
		!snapshot.Authentication.Enabled &&
		snapshot.Authentication.Providers == nil
}

package runtime

import (
	"context"
	"errors"
	"sync"

	"github.com/dsdred/universal-websocket-platform/internal/listener"
	"github.com/dsdred/universal-websocket-platform/internal/message"
	"github.com/dsdred/universal-websocket-platform/internal/runtimeconfig"
	"github.com/dsdred/universal-websocket-platform/internal/secretresolver"
)

var (
	// ErrHostAlreadyBuilt indicates that the Host component graph has already been assembled.
	ErrHostAlreadyBuilt = errors.New("runtime Host already built")
	// ErrHostNotBuilt indicates that Start was called before the Host component graph was assembled.
	ErrHostNotBuilt = errors.New("runtime Host is not built")
	// ErrHostAlreadyRunning indicates that the Host lifecycle has already been started.
	ErrHostAlreadyRunning = errors.New("runtime Host already running")
	// ErrNilSnapshot indicates that NewHost received a zero Snapshot value.
	ErrNilSnapshot = errors.New("runtime Snapshot is nil")
	// ErrNilSecretResolver indicates that Host composition received no Secret Resolver.
	ErrNilSecretResolver = errors.New("runtime Secret Resolver is nil")
)

// Host owns an immutable Runtime Snapshot, assembles components, and coordinates their lifecycle.
type Host interface {
	Snapshot() runtimeconfig.Snapshot
	Build() error
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	Running() bool
}

type hostState uint8

const (
	hostCreated hostState = iota
	hostBuilt
	hostStarting
	hostRunning
	hostStopping
	hostStopped
)

// DefaultHost is the Runtime composition root and lifecycle coordinator.
type DefaultHost struct {
	mu              sync.RWMutex
	snapshot        runtimeconfig.Snapshot
	container       Container
	resolver        secretresolver.Resolver
	handler         message.Handler
	compose         runtimeComposer
	runtimeListener listener.Listener
	state           hostState
	startDone       chan struct{}
	startErr        error
	stopDone        chan struct{}
	stopErr         error
}

// NewHost creates an unbuilt Runtime Host from explicit composition inputs.
func NewHost(
	snapshot runtimeconfig.Snapshot,
	resolver secretresolver.Resolver,
	handler message.Handler,
) (*DefaultHost, error) {
	return newHost(snapshot, resolver, handler, composeRuntime)
}

type runtimeComposer func(
	runtimeconfig.Snapshot,
	secretresolver.Resolver,
	message.Handler,
) (listener.Listener, error)

func newHost(
	snapshot runtimeconfig.Snapshot,
	resolver secretresolver.Resolver,
	handler message.Handler,
	compose runtimeComposer,
) (*DefaultHost, error) {
	if isZeroSnapshot(snapshot) {
		return nil, ErrNilSnapshot
	}
	if resolver == nil {
		return nil, ErrNilSecretResolver
	}

	container, err := New(snapshot)
	if err != nil {
		return nil, err
	}

	return &DefaultHost{
		snapshot:  container.Snapshot(),
		container: container,
		resolver:  resolver,
		handler:   handler,
		compose:   compose,
		state:     hostCreated,
	}, nil
}

// Snapshot returns an independent copy of the Host Snapshot.
func (host *DefaultHost) Snapshot() runtimeconfig.Snapshot {
	host.mu.RLock()
	defer host.mu.RUnlock()
	return cloneSnapshot(host.snapshot)
}

// Build assembles the existing Runtime vertical without starting network resources.
func (host *DefaultHost) Build() error {
	host.mu.Lock()
	defer host.mu.Unlock()
	if host.state != hostCreated {
		return ErrHostAlreadyBuilt
	}

	runtimeListener, err := host.compose(host.snapshot, host.resolver, host.handler)
	if err != nil {
		return err
	}

	host.runtimeListener = runtimeListener
	host.state = hostBuilt
	return nil
}

// Start delegates startup to the composed Listener.
func (host *DefaultHost) Start(ctx context.Context) error {
	host.mu.Lock()
	if host.state == hostCreated {
		host.mu.Unlock()
		return ErrHostNotBuilt
	}
	if host.state != hostBuilt {
		host.mu.Unlock()
		return ErrHostAlreadyRunning
	}
	runtimeListener := host.runtimeListener
	host.state = hostStarting
	host.startDone = make(chan struct{})
	host.startErr = nil
	startDone := host.startDone
	host.mu.Unlock()

	err := runtimeListener.Start(ctx)

	host.mu.Lock()
	host.startErr = err
	if host.state == hostStarting {
		if err != nil {
			host.state = hostBuilt
		} else {
			host.state = hostRunning
		}
	}
	close(startDone)
	host.mu.Unlock()
	return err
}

// Stop delegates shutdown to the composed Listener and otherwise remains a no-op.
func (host *DefaultHost) Stop(ctx context.Context) error {
	host.mu.Lock()
	switch host.state {
	case hostStarting:
		runtimeListener := host.runtimeListener
		startDone := host.startDone
		host.state = hostStopping
		host.stopDone = make(chan struct{})
		host.stopErr = nil
		stopDone := host.stopDone
		host.mu.Unlock()

		<-startDone

		host.mu.RLock()
		startErr := host.startErr
		host.mu.RUnlock()
		if startErr != nil {
			host.mu.Lock()
			host.state = hostBuilt
			host.stopErr = nil
			close(stopDone)
			host.mu.Unlock()
			return nil
		}

		return host.stopListener(ctx, runtimeListener, stopDone)
	case hostRunning:
		runtimeListener := host.runtimeListener
		host.state = hostStopping
		host.stopDone = make(chan struct{})
		host.stopErr = nil
		stopDone := host.stopDone
		host.mu.Unlock()

		return host.stopListener(ctx, runtimeListener, stopDone)
	case hostStopping:
		stopDone := host.stopDone
		host.mu.Unlock()
		<-stopDone

		host.mu.RLock()
		defer host.mu.RUnlock()
		return host.stopErr
	default:
		host.mu.Unlock()
		return nil
	}
}

func (host *DefaultHost) stopListener(
	ctx context.Context,
	runtimeListener listener.Listener,
	stopDone chan struct{},
) error {
	err := runtimeListener.Stop(ctx)

	host.mu.Lock()
	host.state = hostStopped
	host.stopErr = err
	close(stopDone)
	host.mu.Unlock()
	return err
}

// Running reports whether the Host and its composed Listener are Running.
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

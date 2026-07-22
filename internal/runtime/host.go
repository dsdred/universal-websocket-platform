package runtime

import (
	"context"
	"errors"
	"sync"

	"github.com/dsdred/universal-websocket-platform/internal/listener"
	"github.com/dsdred/universal-websocket-platform/internal/message"
	"github.com/dsdred/universal-websocket-platform/internal/runtimeconfig"
	"github.com/dsdred/universal-websocket-platform/internal/secretresolver"
	"github.com/dsdred/universal-websocket-platform/internal/sessionmanager"
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
	ErrNilSecretResolver  = errors.New("runtime Secret Resolver is nil")
	errNilRuntimeListener = errors.New("runtime Listener is nil")
)

// Host owns an immutable Runtime Snapshot, assembles components, and coordinates their lifecycle.
type Host interface {
	Snapshot() runtimeconfig.Snapshot
	RuntimeContext() context.Context
	Build() error
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	Running() bool
	Ready() bool
	CanAccept() bool
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
	mu                sync.RWMutex
	snapshot          runtimeconfig.Snapshot
	container         Container
	resolver          secretresolver.Resolver
	handler           message.Handler
	compose           productionRuntimeComposer
	newRuntimeContext runtimeContextFactory
	stateObserver     func(hostState)
	reportError       func(error)
	admission         admissionGate
	capabilities      *handshakeCapabilities
	runtimeListener   listener.Listener
	sessionManager    *sessionmanager.Manager
	runtimeContext    context.Context
	runtimeCancel     context.CancelFunc
	state             hostState
	startDone         chan struct{}
	startErr          error
	stopDone          chan struct{}
	stopErr           error
}

// NewHost creates an unbuilt Runtime Host from explicit composition inputs.
func NewHost(
	snapshot runtimeconfig.Snapshot,
	resolver secretresolver.Resolver,
	handler message.Handler,
) (*DefaultHost, error) {
	return newHostWithTerminalErrorReporter(snapshot, resolver, handler, nil, nil)
}

type runtimeComposer func(
	runtimeconfig.Snapshot,
	secretresolver.Resolver,
	message.Handler,
) (listener.Listener, error)

type productionRuntimeComposer func(
	runtimeconfig.Snapshot,
	secretresolver.Resolver,
	message.Handler,
) (runtimeComposition, error)

type runtimeContextFactory func() (context.Context, context.CancelFunc)

func newHost(
	snapshot runtimeconfig.Snapshot,
	resolver secretresolver.Resolver,
	handler message.Handler,
	compose runtimeComposer,
) (*DefaultHost, error) {
	return newHostWithTerminalErrorReporter(snapshot, resolver, handler, compose, nil)
}

func newHostWithTerminalErrorReporter(
	snapshot runtimeconfig.Snapshot,
	resolver secretresolver.Resolver,
	handler message.Handler,
	compose runtimeComposer,
	reportError func(error),
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

	host := &DefaultHost{
		snapshot:    container.Snapshot(),
		container:   container,
		resolver:    resolver,
		handler:     handler,
		reportError: reportError,
		newRuntimeContext: func() (context.Context, context.CancelFunc) {
			return context.WithCancel(context.Background())
		},
		state: hostCreated,
	}
	host.capabilities = &handshakeCapabilities{
		canAccept:      host.CanAccept,
		runtimeContext: host.RuntimeContext,
	}
	if compose == nil {
		host.compose = func(
			snapshot runtimeconfig.Snapshot,
			resolver secretresolver.Resolver,
			handler message.Handler,
		) (runtimeComposition, error) {
			return composeRuntime(snapshot, resolver, handler, host.capabilities, host.reportError)
		}
	} else {
		host.compose = func(
			snapshot runtimeconfig.Snapshot,
			resolver secretresolver.Resolver,
			handler message.Handler,
		) (runtimeComposition, error) {
			runtimeListener, err := compose(snapshot, resolver, handler)
			if err != nil {
				return runtimeComposition{}, err
			}
			return runtimeComposition{listener: runtimeListener}, nil
		}
	}
	return host, nil
}

// Snapshot returns an independent copy of the Host Snapshot.
func (host *DefaultHost) Snapshot() runtimeconfig.Snapshot {
	host.mu.RLock()
	defer host.mu.RUnlock()
	return cloneSnapshot(host.snapshot)
}

// RuntimeContext returns the Host-owned context for the running Runtime.
// It is nil until Listener startup succeeds and does not expose cancellation.
func (host *DefaultHost) RuntimeContext() context.Context {
	host.mu.RLock()
	defer host.mu.RUnlock()
	return host.runtimeContext
}

// Build prepares the Host for an atomic startup transaction.
func (host *DefaultHost) Build() error {
	host.mu.Lock()
	defer host.mu.Unlock()
	if host.state != hostCreated {
		return ErrHostAlreadyBuilt
	}
	host.state = hostBuilt
	return nil
}

// Start atomically acquires and starts the current Runtime component graph.
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
	host.state = hostStarting
	host.admission.close()
	host.startDone = make(chan struct{})
	host.startErr = nil
	startDone := host.startDone
	host.mu.Unlock()

	components, err := host.startTransaction(ctx)

	host.mu.Lock()
	host.startErr = err
	if err == nil {
		host.runtimeListener = components.listener
		host.sessionManager = components.sessionManager
		host.runtimeContext, host.runtimeCancel = host.newRuntimeContext()
	} else {
		host.runtimeListener = nil
		host.sessionManager = nil
		host.runtimeContext = nil
		host.runtimeCancel = nil
	}
	if host.state == hostStarting {
		if err != nil {
			host.state = hostBuilt
		} else {
			host.state = hostRunning
			host.admission.allow()
		}
	}
	close(startDone)
	host.mu.Unlock()
	return err
}

func (host *DefaultHost) startTransaction(ctx context.Context) (runtimeComposition, error) {
	transaction := startupTransaction{}
	var components runtimeComposition

	startupErr := transaction.acquire(func() (startupRollback, error) {
		createdComponents, err := host.compose(host.snapshot, host.resolver, host.handler)
		if err != nil {
			return nil, err
		}
		if createdComponents.listener == nil {
			return nil, errNilRuntimeListener
		}
		components = createdComponents
		return createdComponents.listener.Stop, nil
	})
	if startupErr == nil {
		startupErr = components.listener.Start(ctx)
	}
	if startupErr != nil {
		rollbackErr := transaction.rollback(context.Background())
		if rollbackErr != nil {
			return runtimeComposition{}, errors.Join(startupErr, rollbackErr)
		}
		return runtimeComposition{}, startupErr
	}

	transaction.commit()
	return components, nil
}

// Stop coordinates the single Runtime shutdown and publishes one shared result.
func (host *DefaultHost) Stop(ctx context.Context) error {
	host.mu.Lock()
	switch host.state {
	case hostStarting:
		startDone := host.startDone
		host.admission.close()
		host.state = hostStopping
		host.observeStateLocked(hostStopping)
		host.stopDone = make(chan struct{})
		host.stopErr = nil
		stopDone := host.stopDone
		host.mu.Unlock()

		<-startDone

		host.mu.RLock()
		startErr := host.startErr
		runtimeListener := host.runtimeListener
		manager := host.sessionManager
		runtimeCancel := host.runtimeCancel
		host.mu.RUnlock()
		if startErr != nil {
			host.mu.Lock()
			host.state = hostBuilt
			host.stopErr = nil
			close(stopDone)
			host.mu.Unlock()
			return nil
		}

		return host.stopRuntime(ctx, runtimeListener, manager, runtimeCancel, stopDone)
	case hostRunning:
		runtimeListener := host.runtimeListener
		manager := host.sessionManager
		runtimeCancel := host.runtimeCancel
		host.admission.close()
		host.state = hostStopping
		host.observeStateLocked(hostStopping)
		host.stopDone = make(chan struct{})
		host.stopErr = nil
		stopDone := host.stopDone
		host.mu.Unlock()

		return host.stopRuntime(ctx, runtimeListener, manager, runtimeCancel, stopDone)
	case hostStopping:
		stopDone := host.stopDone
		host.mu.Unlock()
		<-stopDone

		host.mu.RLock()
		defer host.mu.RUnlock()
		return host.stopErr
	case hostStopped:
		err := host.stopErr
		host.mu.Unlock()
		return err
	default:
		host.mu.Unlock()
		return nil
	}
}

func (host *DefaultHost) observeStateLocked(state hostState) {
	if host.stateObserver != nil {
		host.stateObserver(state)
	}
}

func (host *DefaultHost) stopRuntime(
	ctx context.Context,
	runtimeListener listener.Listener,
	manager *sessionmanager.Manager,
	runtimeCancel context.CancelFunc,
	stopDone chan struct{},
) error {
	if manager != nil {
		shutdown := manager.BeginShutdown()
		for _, registration := range shutdown.Registrations() {
			registration.RequestStop()
		}
	}

	runtimeCancel()
	listenerErr := runtimeListener.Stop(ctx)
	var managerErr error
	if manager != nil {
		managerErr = manager.Wait(ctx)
	}
	stopErr := errors.Join(listenerErr, managerErr)

	host.mu.Lock()
	host.state = hostStopped
	host.stopErr = stopErr
	close(stopDone)
	host.mu.Unlock()
	return stopErr
}

// Running reports whether the Host and its composed Listener are Running.
func (host *DefaultHost) Running() bool {
	host.mu.RLock()
	defer host.mu.RUnlock()
	return host.state == hostRunning
}

// Ready reports whether the startup transaction committed and the Host is Running.
func (host *DefaultHost) Ready() bool {
	host.mu.RLock()
	defer host.mu.RUnlock()
	return host.state == hostRunning
}

// CanAccept reports whether Runtime lifecycle currently permits new connections.
func (host *DefaultHost) CanAccept() bool {
	host.mu.RLock()
	defer host.mu.RUnlock()
	return host.state == hostRunning && host.admission.canAccept()
}

func isZeroSnapshot(snapshot runtimeconfig.Snapshot) bool {
	return snapshot.ConfigurationID == 0 &&
		snapshot.VersionID == 0 &&
		snapshot.Listener == (runtimeconfig.ListenerSnapshot{}) &&
		!snapshot.Authentication.Enabled &&
		snapshot.Authentication.Providers == nil
}

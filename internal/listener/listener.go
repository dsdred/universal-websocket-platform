package listener

import (
	"context"
	"errors"
	"net"
	"net/http"
	"strconv"
	"sync"
)

var (
	ErrListenerAlreadyRunning       = errors.New("runtime Listener already running")
	ErrInvalidListenerConfiguration = errors.New("invalid runtime Listener configuration")
)

// Listener exposes the lifecycle of a configured Runtime Listener without transport details.
type Listener interface {
	Address() string
	Running() bool
	Start(context.Context) error
	Stop(context.Context) error
}

type listenerState uint8

const (
	listenerCreated listenerState = iota
	listenerRunning
	listenerStopping
	listenerStopped
)

type tlsConfiguration struct {
	enabled        bool
	certificateRef string
	privateKeyRef  string
	minVersion     string
}

// DefaultListener stores effective Listener metadata and coordinates its lifecycle.
type DefaultListener struct {
	mu               sync.RWMutex
	host             string
	port             uint16
	tls              tlsConfiguration
	state            listenerState
	listener         net.Listener
	server           *http.Server
	serverStop       context.CancelFunc
	handshakeHandler http.Handler
	reportError      func(error)
	serveHTTP        func(*http.Server, net.Listener) error
	shutdownHTTP     func(*http.Server, context.Context) error
	closeTCP         func(net.Listener) error
	wg               sync.WaitGroup
	handlerWG        sync.WaitGroup
	stopDone         chan struct{}
	stopErr          error
}

// Address returns the configured host and port without opening a socket.
func (listener *DefaultListener) Address() string {
	listener.mu.RLock()
	defer listener.mu.RUnlock()
	return net.JoinHostPort(listener.host, strconv.Itoa(int(listener.port)))
}

// Running reports whether the Listener is in the Running state.
func (listener *DefaultListener) Running() bool {
	listener.mu.RLock()
	defer listener.mu.RUnlock()
	return listener.state == listenerRunning
}

// Start opens the configured TCP address and starts accepting connections.
func (listener *DefaultListener) Start(context.Context) error {
	listener.mu.Lock()
	defer listener.mu.Unlock()
	if listener.state != listenerCreated {
		return ErrListenerAlreadyRunning
	}

	tcpListener, err := net.Listen("tcp", net.JoinHostPort(listener.host, strconv.Itoa(int(listener.port))))
	if err != nil {
		return err
	}
	serverContext, serverStop := context.WithCancel(context.Background())
	httpHandler := newHTTPHandlerWithHandshake(listener.handshakeHandler)
	httpServer := &http.Server{
		Handler: trackedHandler(httpHandler, &listener.handlerWG),
		BaseContext: func(net.Listener) context.Context {
			return serverContext
		},
	}

	listener.listener = tcpListener
	listener.server = httpServer
	listener.serverStop = serverStop
	listener.state = listenerRunning
	listener.wg.Add(1)
	go listener.serve(httpServer, tcpListener)
	return nil
}

// Stop gracefully shuts down the HTTP Server and waits for its accept loop.
func (listener *DefaultListener) Stop(ctx context.Context) error {
	listener.mu.Lock()
	switch listener.state {
	case listenerCreated:
		listener.mu.Unlock()
		return nil
	case listenerStopping:
		stopDone := listener.stopDone
		listener.mu.Unlock()
		// This caller does not own shutdown. Its context bounds only the wait and
		// cannot cancel or replace the primary shutdown attempt or its stored result.
		select {
		case <-stopDone:
			listener.mu.RLock()
			defer listener.mu.RUnlock()
			return listener.stopErr
		case <-ctx.Done():
			return ctx.Err()
		}
	case listenerStopped:
		err := listener.stopErr
		listener.mu.Unlock()
		return err
	case listenerRunning:
		// The caller that performs this transition owns the one shutdown attempt.
	default:
		listener.mu.Unlock()
		return nil
	}
	tcpListener := listener.listener
	httpServer := listener.server
	serverStop := listener.serverStop
	stopDone := make(chan struct{})
	listener.stopDone = stopDone
	listener.stopErr = nil
	listener.state = listenerStopping
	listener.mu.Unlock()

	serverStop()
	shutdownHTTP := listener.shutdownHTTP
	if shutdownHTTP == nil {
		shutdownHTTP = func(server *http.Server, shutdownContext context.Context) error {
			return server.Shutdown(shutdownContext)
		}
	}
	closeTCP := listener.closeTCP
	if closeTCP == nil {
		closeTCP = func(listener net.Listener) error { return listener.Close() }
	}
	shutdownErr := shutdownHTTP(httpServer, ctx)
	closeErr := closeTCP(tcpListener)
	listener.wg.Wait()
	listener.handlerWG.Wait()
	if errors.Is(closeErr, net.ErrClosed) {
		closeErr = nil
	}
	stopErr := errors.Join(shutdownErr, closeErr)

	listener.mu.Lock()
	listener.listener = nil
	listener.server = nil
	listener.serverStop = nil
	listener.state = listenerStopped
	listener.stopErr = stopErr
	close(stopDone)
	listener.mu.Unlock()
	return stopErr
}

func trackedHandler(handler http.Handler, waitGroup *sync.WaitGroup) http.Handler {
	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		waitGroup.Add(1)
		defer waitGroup.Done()
		handler.ServeHTTP(response, request)
	})
}

func (listener *DefaultListener) serve(httpServer *http.Server, tcpListener net.Listener) {
	defer listener.wg.Done()
	err := listener.serveHTTP(httpServer, tcpListener)
	if err != nil && !errors.Is(err, http.ErrServerClosed) && !errors.Is(err, net.ErrClosed) {
		reportListenerTerminalError(listener.reportError, listenerServeError{cause: err})
	}
}

type listenerServeError struct {
	cause error
}

func (err listenerServeError) Error() string {
	return "HTTP Server Serve failed"
}

func (err listenerServeError) Unwrap() error {
	return err.cause
}

func reportListenerTerminalError(reporter func(error), err error) {
	if reporter == nil || err == nil {
		return
	}
	// Reporting is observer-only: a faulty consumer cannot terminate the serve owner or alter Stop.
	func() {
		defer func() {
			_ = recover()
		}()
		reporter(err)
	}()
}

func notImplementedHandler(response http.ResponseWriter, _ *http.Request) {
	response.WriteHeader(http.StatusNotImplemented)
}

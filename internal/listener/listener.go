package listener

import (
	"context"
	"errors"
	"net"
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
	mu       sync.RWMutex
	host     string
	port     uint16
	tls      tlsConfiguration
	state    listenerState
	listener net.Listener
	wg       sync.WaitGroup
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

	listener.listener = tcpListener
	listener.state = listenerRunning
	listener.wg.Add(1)
	go listener.acceptLoop(tcpListener)
	return nil
}

// Stop closes the TCP Listener, waits for the accept loop, and is idempotent.
func (listener *DefaultListener) Stop(context.Context) error {
	listener.mu.Lock()
	if listener.state != listenerRunning {
		listener.mu.Unlock()
		return nil
	}
	tcpListener := listener.listener
	listener.state = listenerStopping
	listener.mu.Unlock()

	err := tcpListener.Close()
	listener.wg.Wait()

	listener.mu.Lock()
	listener.listener = nil
	listener.state = listenerStopped
	listener.mu.Unlock()

	if err != nil && !errors.Is(err, net.ErrClosed) {
		return err
	}
	return nil
}

func (listener *DefaultListener) acceptLoop(tcpListener net.Listener) {
	defer listener.wg.Done()
	for {
		connection, err := tcpListener.Accept()
		if err != nil {
			return
		}
		_ = connection.Close()
	}
}

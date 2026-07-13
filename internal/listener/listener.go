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
	mu    sync.RWMutex
	host  string
	port  uint16
	tls   tlsConfiguration
	state listenerState
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

// Start moves a newly created Listener to Running without opening a socket.
func (listener *DefaultListener) Start(context.Context) error {
	listener.mu.Lock()
	defer listener.mu.Unlock()
	if listener.state != listenerCreated {
		return ErrListenerAlreadyRunning
	}
	listener.state = listenerRunning
	return nil
}

// Stop moves a Running Listener to Stopped and is otherwise a no-op.
func (listener *DefaultListener) Stop(context.Context) error {
	listener.mu.Lock()
	defer listener.mu.Unlock()
	if listener.state == listenerRunning {
		listener.state = listenerStopped
	}
	return nil
}

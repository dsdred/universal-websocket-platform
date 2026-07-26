package runtime

import (
	"errors"

	"github.com/dsdred/universal-websocket-platform/internal/runtimeconfig"
)

// Container exposes dependencies available to Runtime components.
type Container interface {
	Snapshot() runtimeconfig.Snapshot
}

// DefaultContainer stores the immutable Runtime Configuration Snapshot.
type DefaultContainer struct {
	snapshot runtimeconfig.Snapshot
}

// New creates a Runtime dependency Container with its own Snapshot copy.
func New(snapshot runtimeconfig.Snapshot) (*DefaultContainer, error) {
	provenance := snapshot.Provenance()
	if provenance.ConfigurationVersionID == 0 {
		return nil, errors.New("create runtime container: VersionID must not be zero")
	}
	if provenance.ConfigurationID == 0 {
		return nil, errors.New("create runtime container: ConfigurationID must not be zero")
	}
	listener := snapshot.Listener()
	if listener.Host == "" || listener.Port == 0 {
		return nil, errors.New("create runtime container: Listener must contain Host and Port")
	}

	return &DefaultContainer{snapshot: snapshot}, nil
}

// Snapshot returns an independent copy of the Runtime Configuration Snapshot.
func (container *DefaultContainer) Snapshot() runtimeconfig.Snapshot {
	return container.snapshot
}

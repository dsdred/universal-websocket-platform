package runtime

import (
	"errors"
	"fmt"
	"strings"

	"github.com/dsdred/universal-websocket-platform/internal/runtimeconfig"
)

var (
	// ErrInvalidRuntimeConfiguration identifies a Snapshot that cannot be executed safely.
	ErrInvalidRuntimeConfiguration = errors.New("invalid runtime executable configuration")
	// ErrUnsupportedRuntimeCapability identifies an active startup capability absent from this Runtime build.
	ErrUnsupportedRuntimeCapability = errors.New("unsupported runtime capability")
)

const (
	minimumHandshakeSeconds = 1
	maximumHandshakeSeconds = 300
)

// validateExecutableSnapshot validates startup-critical capabilities only.
// Read, write, and idle timeouts are retained as configured-but-inactive Runtime
// capabilities until the Listener Settings roadmap gate defines their execution.
func validateExecutableSnapshot(snapshot runtimeconfig.Snapshot) error {
	if snapshot.ConfigurationID == 0 || snapshot.VersionID == 0 {
		return invalidRuntimeField("identity")
	}
	if strings.TrimSpace(snapshot.Listener.Host) == "" || snapshot.Listener.Port == 0 {
		return invalidRuntimeField("listener")
	}
	if snapshot.Listener.TLS.MinVersion != "1.2" && snapshot.Listener.TLS.MinVersion != "1.3" {
		return invalidRuntimeField("TLS minimum version")
	}
	if snapshot.Listener.TLS.Enabled {
		return fmt.Errorf(
			"%w: %w: TLS",
			ErrInvalidRuntimeConfiguration,
			ErrUnsupportedRuntimeCapability,
		)
	}
	if snapshot.Listener.Timeouts.HandshakeSeconds < minimumHandshakeSeconds ||
		snapshot.Listener.Timeouts.HandshakeSeconds > maximumHandshakeSeconds {
		return invalidRuntimeField("handshake timeout")
	}
	return nil
}

func invalidRuntimeField(field string) error {
	return fmt.Errorf("%w: %s", ErrInvalidRuntimeConfiguration, field)
}

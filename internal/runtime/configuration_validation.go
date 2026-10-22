package runtime

import (
	"errors"
	"fmt"
	"strings"

	"github.com/dsdred/universal-websocket-platform/internal/runtimeconfig"
	"github.com/dsdred/universal-websocket-platform/internal/secretresolver"
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

// validateExecutableSnapshot rejects nonexecutable fields before resources.
func validateExecutableSnapshot(snapshot runtimeconfig.Snapshot) error {
	provenance := snapshot.Provenance()
	if provenance.ConfigurationID == 0 || provenance.ConfigurationVersionID == 0 {
		return invalidRuntimeField("identity")
	}
	listener := snapshot.Listener()
	if strings.TrimSpace(listener.Host) == "" || listener.Port == 0 {
		return invalidRuntimeField("listener")
	}
	if listener.TLS.MinVersion != "1.2" && listener.TLS.MinVersion != "1.3" {
		return invalidRuntimeField("TLS minimum version")
	}
	if listener.TLS.Enabled {
		if secretresolver.ValidateReference(listener.TLS.CertificateRef) != nil ||
			secretresolver.ValidateReference(listener.TLS.PrivateKeyRef) != nil {
			return invalidRuntimeField("TLS references")
		}
	}
	if listener.Timeouts.HandshakeSeconds < minimumHandshakeSeconds ||
		listener.Timeouts.HandshakeSeconds > maximumHandshakeSeconds {
		return invalidRuntimeField("handshake timeout")
	}
	if listener.Timeouts.WriteSeconds < 1 || listener.Timeouts.WriteSeconds > 300 {
		return invalidRuntimeField("write timeout")
	}
	if err := validateReadSeconds(listener.Timeouts.ReadSeconds); err != nil {
		return err
	}
	return validateIdleSeconds(listener.Timeouts.IdleSeconds)
}

func validateReadSeconds(seconds uint32) error {
	if seconds > 86400 {
		return invalidRuntimeField("read timeout")
	}
	return nil
}

func validateIdleSeconds(seconds uint32) error {
	if seconds > 86400 {
		return invalidRuntimeField("idle timeout")
	}
	return nil
}

func invalidRuntimeField(field string) error {
	return fmt.Errorf("%w: %s", ErrInvalidRuntimeConfiguration, field)
}

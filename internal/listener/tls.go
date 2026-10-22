package listener

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"time"

	"github.com/dsdred/universal-websocket-platform/internal/runtimeconfig"
	"github.com/dsdred/universal-websocket-platform/internal/secretresolver"
)

// PreparedTLS is immutable transport material, never Configuration or Snapshot data.
// Only PrepareTLS can create it; callers cannot inject callbacks or mutate keys.
type PreparedTLS struct {
	config   *tls.Config
	settings runtimeconfig.TLSSnapshot
}

// PrepareTLS resolves one leaf-first PEM chain and unencrypted PEM private key.
// It opens no sockets and retains no startup context or raw PEM buffers.
func PrepareTLS(ctx context.Context, snapshot runtimeconfig.ListenerSnapshot, resolver secretresolver.Resolver) (*PreparedTLS, error) {
	if !snapshot.TLS.Enabled {
		return nil, nil
	}
	if ctx == nil || resolver == nil {
		return nil, tlsFailure("input", nil)
	}
	minimum := uint16(tls.VersionTLS12)
	switch snapshot.TLS.MinVersion {
	case "1.2":
	case "1.3":
		minimum = tls.VersionTLS13
	default:
		return nil, tlsFailure("minimum version", nil)
	}
	if err := secretresolver.ValidateReference(snapshot.TLS.CertificateRef); err != nil {
		return nil, tlsFailure("certificate reference", err)
	}
	if err := secretresolver.ValidateReference(snapshot.TLS.PrivateKeyRef); err != nil {
		return nil, tlsFailure("key reference", err)
	}
	if err := ctx.Err(); err != nil {
		return nil, tlsFailure("cancellation", err)
	}
	certificate, err := resolver.Resolve(ctx, snapshot.TLS.CertificateRef)
	defer clear(certificate.Value)
	if err != nil {
		return nil, tlsFailure("certificate resolution", err)
	}
	if err := ctx.Err(); err != nil {
		return nil, tlsFailure("cancellation", err)
	}
	key, err := resolver.Resolve(ctx, snapshot.TLS.PrivateKeyRef)
	defer clear(key.Value)
	if err != nil {
		return nil, tlsFailure("key resolution", err)
	}
	if err := ctx.Err(); err != nil {
		return nil, tlsFailure("cancellation", err)
	}
	pair, err := tls.X509KeyPair(certificate.Value, key.Value)
	if err != nil {
		return nil, tlsFailure("key pair", err)
	}
	for index, der := range pair.Certificate {
		parsed, err := x509.ParseCertificate(der)
		if err != nil {
			return nil, tlsFailure("certificate chain", err)
		}
		if index == 0 {
			pair.Leaf = parsed
		}
	}
	leaf := pair.Leaf
	now := time.Now()
	if leaf == nil || now.Before(leaf.NotBefore) || now.After(leaf.NotAfter) {
		return nil, tlsFailure("certificate validity", nil)
	}
	if leaf.KeyUsage != 0 && leaf.KeyUsage&x509.KeyUsageDigitalSignature == 0 {
		return nil, tlsFailure("key usage", nil)
	}
	serverUsage := len(leaf.ExtKeyUsage) == 0 && len(leaf.UnknownExtKeyUsage) == 0
	for _, usage := range leaf.ExtKeyUsage {
		if usage == x509.ExtKeyUsageAny || usage == x509.ExtKeyUsageServerAuth {
			serverUsage = true
		}
	}
	if !serverUsage {
		return nil, tlsFailure("extended key usage", nil)
	}
	if err := ctx.Err(); err != nil {
		return nil, tlsFailure("cancellation", err)
	}
	return &PreparedTLS{settings: snapshot.TLS, config: &tls.Config{
		Certificates: []tls.Certificate{pair},
		MinVersion:   minimum,
		MaxVersion:   tls.VersionTLS13,
		NextProtos:   []string{"http/1.1"},
		ClientAuth:   tls.NoClientCert,
	}}, nil
}

type tlsPreparationError struct {
	stage string
	cause error
}

func tlsFailure(stage string, cause error) error {
	return &tlsPreparationError{stage: stage, cause: cause}
}

func (err *tlsPreparationError) Error() string {
	return "invalid runtime Listener configuration: TLS " + err.stage
}

// Format prevents debug formatting from exposing the retained private cause.
func (err *tlsPreparationError) Format(state fmt.State, _ rune) {
	_, _ = fmt.Fprint(state, err.Error())
}

func (*tlsPreparationError) Unwrap() error { return ErrInvalidListenerConfiguration }

// Is preserves exact resolver/parser identity without exposing raw diagnostic text.
func (err *tlsPreparationError) Is(target error) bool { return errors.Is(err.cause, target) }

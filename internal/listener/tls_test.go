package listener

import (
	"bufio"
	"bytes"
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"encoding/pem"
	"errors"
	"fmt"
	"math/big"
	"net"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/dsdred/universal-websocket-platform/internal/runtimeconfig"
	"github.com/dsdred/universal-websocket-platform/internal/secretresolver"
)

type tlsResolverFunc func(context.Context, string) (secretresolver.Secret, error)

func (f tlsResolverFunc) Resolve(ctx context.Context, ref string) (secretresolver.Secret, error) {
	return f(ctx, ref)
}

func tlsFixture(t *testing.T, mutate func(*x509.Certificate)) ([]byte, []byte, *x509.CertPool) {
	t.Helper()
	public, private, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	certificate := &x509.Certificate{
		SerialNumber: big.NewInt(1), Subject: pkix.Name{CommonName: "local test"},
		NotBefore: time.Now().Add(-time.Hour), NotAfter: time.Now().Add(time.Hour),
		IPAddresses: []net.IP{net.ParseIP("127.0.0.1")},
		KeyUsage:    x509.KeyUsageDigitalSignature, ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}
	if mutate != nil {
		mutate(certificate)
	}
	der, err := x509.CreateCertificate(rand.Reader, certificate, certificate, public, private)
	if err != nil {
		t.Fatal(err)
	}
	keyDER, err := x509.MarshalPKCS8PrivateKey(private)
	if err != nil {
		t.Fatal(err)
	}
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: keyDER})
	pool := x509.NewCertPool()
	if !pool.AppendCertsFromPEM(certPEM) {
		t.Fatal("trust fixture")
	}
	return certPEM, keyPEM, pool
}

func tlsSnapshot(t *testing.T) runtimeconfig.ListenerSnapshot {
	t.Helper()
	snapshot := validListenerSnapshot()
	snapshot.Port = availableTCPPort(t)
	snapshot.TLS = runtimeconfig.TLSSnapshot{Enabled: true, CertificateRef: "certificates/test", PrivateKeyRef: "secrets/test-key", MinVersion: "1.2"}
	snapshot.Timeouts.HandshakeSeconds = 1
	return snapshot
}

func tlsMemory(t *testing.T, cert, key []byte) secretresolver.Resolver {
	t.Helper()
	resolver, err := secretresolver.NewMemory(map[string][]byte{"certificates/test": cert, "secrets/test-key": key})
	if err != nil {
		t.Fatal(err)
	}
	return resolver
}

func TestTLSPreparationDisabledDoesNotResolve(t *testing.T) {
	snapshot := tlsSnapshot(t)
	snapshot.TLS.Enabled = false
	prepared, err := PrepareTLS(context.Background(), snapshot, tlsResolverFunc(func(context.Context, string) (secretresolver.Secret, error) {
		t.Fatal("disabled TLS resolved material")
		return secretresolver.Secret{}, nil
	}))
	if err != nil || prepared != nil {
		t.Fatalf("disabled preparation = %v, %v", prepared, err)
	}
}

func TestTLSPreparationValidationAndRawBufferCleanup(t *testing.T) {
	cert, key, _ := tlsFixture(t, nil)
	_, otherKey, _ := tlsFixture(t, nil)
	for _, test := range []struct {
		name      string
		cert, key []byte
		good      bool
	}{
		{"valid", cert, key, true},
		{"empty certificate", nil, key, false},
		{"empty key", cert, nil, false},
		{"malformed certificate", []byte("SECRET-CERT-INVALID"), key, false},
		{"malformed key", cert, []byte("SECRET-KEY-INVALID"), false},
		{"mismatched key", cert, otherKey, false},
		{"malformed later chain certificate", append(bytes.Clone(cert), pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: []byte("invalid-der")})...), key, false},
	} {
		t.Run(test.name, func(t *testing.T) {
			var borrowed [][]byte
			var refs []string
			ctx := context.WithValue(context.Background(), tlsTestContextKey{}, "exact context")
			resolver := tlsResolverFunc(func(got context.Context, ref string) (secretresolver.Secret, error) {
				if got != ctx {
					t.Fatal("resolver did not receive exact preparation context")
				}
				refs = append(refs, ref)
				value := test.cert
				if ref == "secrets/test-key" {
					value = test.key
				}
				value = bytes.Clone(value)
				borrowed = append(borrowed, value)
				return secretresolver.Secret{Value: value}, nil
			})
			prepared, err := PrepareTLS(ctx, tlsSnapshot(t), resolver)
			if test.good {
				if err != nil || prepared == nil {
					t.Fatalf("valid preparation: %v", err)
				}
				if strings.Join(refs, ",") != "certificates/test,secrets/test-key" {
					t.Fatalf("resolution order = %v", refs)
				}
			} else if prepared != nil || !errors.Is(err, ErrInvalidListenerConfiguration) {
				t.Fatalf("invalid preparation = %v, %v", prepared, err)
			}
			for _, value := range borrowed {
				for _, b := range value {
					if b != 0 {
						t.Fatal("raw returned material was not cleared")
					}
				}
			}
		})
	}
}

type tlsTestContextKey struct{}
type tlsSensitiveError struct{}

func (*tlsSensitiveError) Error() string { return "SECRET-RESOLVER-PAYLOAD" }

func TestTLSPreparationPreservesIdentityWithoutExposingCause(t *testing.T) {
	cert, _, _ := tlsFixture(t, nil)
	for _, stage := range []string{"certificate", "key"} {
		t.Run(stage, func(t *testing.T) {
			cause := &tlsSensitiveError{}
			resolver := tlsResolverFunc(func(_ context.Context, ref string) (secretresolver.Secret, error) {
				if stage == "certificate" || ref == "secrets/test-key" {
					return secretresolver.Secret{}, cause
				}
				return secretresolver.Secret{Value: bytes.Clone(cert)}, nil
			})
			_, err := PrepareTLS(context.Background(), tlsSnapshot(t), resolver)
			if !errors.Is(err, cause) || !errors.Is(err, ErrInvalidListenerConfiguration) {
				t.Fatalf("missing identities: %v", err)
			}
			if errors.Unwrap(err) != ErrInvalidListenerConfiguration {
				t.Fatal("ordinary Unwrap exposes nonclassification cause")
			}
			var secret *tlsSensitiveError
			if errors.As(err, &secret) {
				t.Fatal("As exposed retained secret cause")
			}
			for _, format := range []string{"%s", "%v", "%+v", "%#v", "%q", "%x", "%X"} {
				value := fmt.Sprintf(format, err)
				if strings.Contains(value, "SECRET") || strings.Contains(value, "certificates/test") || strings.Contains(value, "secrets/test-key") {
					t.Fatalf("unsafe format %s: %s", format, value)
				}
			}
		})
	}
}

func TestTLSPreparationRejectsLeafPolicy(t *testing.T) {
	for _, test := range []struct {
		name   string
		change func(*x509.Certificate)
		good   bool
	}{
		{"expired", func(c *x509.Certificate) { c.NotAfter = time.Now().Add(-time.Minute) }, false},
		{"future", func(c *x509.Certificate) { c.NotBefore = time.Now().Add(time.Minute) }, false},
		{"client only", func(c *x509.Certificate) { c.ExtKeyUsage = []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth} }, false},
		{"unknown only", func(c *x509.Certificate) {
			c.ExtKeyUsage = nil
			c.UnknownExtKeyUsage = []asn1.ObjectIdentifier{{1, 2, 3, 4}}
		}, false},
		{"no digital signature", func(c *x509.Certificate) { c.KeyUsage = x509.KeyUsageKeyEncipherment }, false},
		{"empty usage", func(c *x509.Certificate) { c.ExtKeyUsage = nil; c.KeyUsage = 0 }, true},
		{"any usage", func(c *x509.Certificate) { c.ExtKeyUsage = []x509.ExtKeyUsage{x509.ExtKeyUsageAny} }, true},
	} {
		t.Run(test.name, func(t *testing.T) {
			cert, key, _ := tlsFixture(t, test.change)
			prepared, err := PrepareTLS(context.Background(), tlsSnapshot(t), tlsMemory(t, cert, key))
			if test.good {
				if err != nil || prepared == nil {
					t.Fatalf("accepted policy: %v", err)
				}
			} else if prepared != nil || !errors.Is(err, ErrInvalidListenerConfiguration) {
				t.Fatalf("rejected policy: %v", err)
			}
		})
	}
}

func TestTLSPreparationCancellationBoundaries(t *testing.T) {
	cert, key, _ := tlsFixture(t, nil)
	for _, after := range []int{0, 1, 2} {
		t.Run(fmt.Sprint(after), func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			if after == 0 {
				cancel()
			}
			calls := 0
			prepared, err := PrepareTLS(ctx, tlsSnapshot(t), tlsResolverFunc(func(got context.Context, ref string) (secretresolver.Secret, error) {
				if got != ctx {
					t.Fatal("context identity")
				}
				calls++
				value := cert
				if ref == "secrets/test-key" {
					value = key
				}
				if calls == after {
					cancel()
				}
				return secretresolver.Secret{Value: bytes.Clone(value)}, nil
			}))
			if prepared != nil || !errors.Is(err, context.Canceled) || calls != after {
				t.Fatalf("prepared=%v err=%v calls=%d want %d", prepared, err, calls, after)
			}
		})
	}
}

func TestTLSBootstrapPreparedBindingAndStopBeforeStart(t *testing.T) {
	cert, key, pool := tlsFixture(t, nil)
	snapshot := tlsSnapshot(t)
	prepared, err := PrepareTLS(context.Background(), snapshot, tlsMemory(t, cert, key))
	if err != nil {
		t.Fatal(err)
	}
	for _, change := range []func(*runtimeconfig.ListenerSnapshot){
		func(s *runtimeconfig.ListenerSnapshot) { s.TLS.MinVersion = "1.3" },
		func(s *runtimeconfig.ListenerSnapshot) { s.TLS.CertificateRef = "certificates/other" },
		func(s *runtimeconfig.ListenerSnapshot) { s.TLS.PrivateKeyRef = "secrets/other" },
	} {
		other := snapshot
		change(&other)
		if created, err := NewBootstrapWithTLSAndHandshake(http.NotFoundHandler(), nil, prepared).Build(other); created != nil || !errors.Is(err, ErrInvalidListenerConfiguration) {
			t.Fatalf("mismatched prepared capability accepted: %v", err)
		}
	}
	created, err := NewBootstrapWithTLSAndHandshake(http.NotFoundHandler(), nil, prepared).Build(snapshot)
	if err != nil {
		t.Fatal(err)
	}
	if err := created.Stop(context.Background()); err != nil {
		t.Fatal(err)
	}
	if err := created.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = created.Stop(context.Background()) })
	conn, err := tls.DialWithDialer(&net.Dialer{Timeout: time.Second}, "tcp", created.Address(), &tls.Config{RootCAs: pool, ServerName: "127.0.0.1", MinVersion: tls.VersionTLS12})
	if err != nil {
		t.Fatalf("pre-start Stop lost TLS material: %v", err)
	}
	conn.Close()
}

func startTLSListener(t *testing.T, min string, handler http.Handler) (*DefaultListener, *x509.CertPool) {
	t.Helper()
	cert, key, pool := tlsFixture(t, nil)
	snapshot := tlsSnapshot(t)
	snapshot.TLS.MinVersion = min
	prepared, err := PrepareTLS(context.Background(), snapshot, tlsMemory(t, cert, key))
	if err != nil {
		t.Fatal(err)
	}
	created, err := NewBootstrapWithTLSAndHandshake(handler, nil, prepared).Build(snapshot)
	if err != nil {
		t.Fatal(err)
	}
	current := created.(*DefaultListener)
	if err := current.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		if err := current.Stop(ctx); err != nil {
			t.Errorf("TLS cleanup: %v", err)
		}
	})
	return current, pool
}

func TestTLSListenerVersionAndProtocolNegotiation(t *testing.T) {
	for _, minimum := range []string{"1.2", "1.3"} {
		t.Run(minimum, func(t *testing.T) {
			current, pool := startTLSListener(t, minimum, http.NotFoundHandler())
			for _, test := range []struct {
				name       string
				version    uint16
				protos     []string
				good       bool
				negotiated string
			}{
				{"tls12", tls.VersionTLS12, []string{"http/1.1"}, minimum == "1.2", "http/1.1"},
				{"tls13", tls.VersionTLS13, []string{"http/1.1"}, true, "http/1.1"},
				{"no ALPN", tls.VersionTLS13, nil, true, ""},
				{"h2 only", tls.VersionTLS13, []string{"h2"}, false, ""},
			} {
				t.Run(test.name, func(t *testing.T) {
					conn, err := tls.DialWithDialer(&net.Dialer{Timeout: 2 * time.Second}, "tcp", current.Address(), &tls.Config{RootCAs: pool, ServerName: "127.0.0.1", MinVersion: test.version, MaxVersion: test.version, NextProtos: test.protos})
					if !test.good {
						if err == nil {
							conn.Close()
							t.Fatal("unsupported protocol succeeded")
						}
						return
					}
					if err != nil {
						t.Fatal(err)
					}
					defer conn.Close()
					state := conn.ConnectionState()
					if state.Version != test.version || state.NegotiatedProtocol != test.negotiated {
						t.Fatalf("negotiated = %#v", state)
					}
				})
			}
			if current.server.ReadHeaderTimeout != time.Second || current.server.ReadTimeout != 0 || current.server.WriteTimeout != 0 || current.server.IdleTimeout != 0 {
				t.Fatal("TLS handshake budget corrupted HTTP/Session timeout boundary")
			}
		})
	}
}

func TestTLSListenerPlaintextNeverReachesApplication(t *testing.T) {
	var calls atomic.Int32
	current, _ := startTLSListener(t, "1.2", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { calls.Add(1); w.WriteHeader(http.StatusNoContent) }))
	conn, err := net.DialTimeout("tcp", current.Address(), time.Second)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	conn.SetDeadline(time.Now().Add(2 * time.Second))
	fmt.Fprint(conn, "GET /ws HTTP/1.1\r\nHost: localhost\r\nConnection: Upgrade\r\nUpgrade: websocket\r\nSec-WebSocket-Version: 13\r\nSec-WebSocket-Key: dGhlIHNhbXBsZSBub25jZQ==\r\n\r\n")
	line, _ := bufio.NewReader(conn).ReadString('\n')
	if strings.Contains(line, "101") || calls.Load() != 0 {
		t.Fatalf("plaintext application fallback: %q calls%d", line, calls.Load())
	}
}

func TestTLSListenerHandshakeBudgetAndStop(t *testing.T) {
	for _, test := range []struct {
		name   string
		prefix []byte
		stop   bool
	}{
		{"silent deadline", nil, false}, {"partial deadline", []byte{0x16, 0x03, 0x01, 0x00, 0x80, 0x01}, false}, {"silent Stop", nil, true},
	} {
		t.Run(test.name, func(t *testing.T) {
			current, _ := startTLSListener(t, "1.2", http.NotFoundHandler())
			conn, err := net.DialTimeout("tcp", current.Address(), time.Second)
			if err != nil {
				t.Fatal(err)
			}
			defer conn.Close()
			if len(test.prefix) > 0 {
				if _, err := conn.Write(test.prefix); err != nil {
					t.Fatal(err)
				}
			}
			start := time.Now()
			if test.stop {
				ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
				defer cancel()
				if err := current.Stop(ctx); err != nil {
					t.Fatal(err)
				}
			}
			conn.SetReadDeadline(time.Now().Add(2500 * time.Millisecond))
			_, err = conn.Read(make([]byte, 1))
			if err == nil {
				t.Fatal("pending TLS peer did not close")
			}
			if timeout, ok := err.(net.Error); ok && timeout.Timeout() {
				t.Fatalf("caller timeout instead of TLS termination: %v", err)
			}
			if !test.stop && time.Since(start) < 700*time.Millisecond {
				t.Fatalf("handshake expired prematurely: %v", time.Since(start))
			}
		})
	}
}

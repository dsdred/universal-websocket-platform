package runtime

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"fmt"
	"math/big"
	"net"
	"net/http"
	"reflect"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/coder/websocket"
	"github.com/dsdred/universal-websocket-platform/internal/configurationversion"
	"github.com/dsdred/universal-websocket-platform/internal/listener"
	"github.com/dsdred/universal-websocket-platform/internal/message"
	"github.com/dsdred/universal-websocket-platform/internal/runtimeconfig"
	"github.com/dsdred/universal-websocket-platform/internal/secretresolver"
	"github.com/dsdred/universal-websocket-platform/internal/sessionmanager"
	"github.com/golang-jwt/jwt/v5"
)

type runtimeTLSResolverFunc func(context.Context, string) (secretresolver.Secret, error)

func (f runtimeTLSResolverFunc) Resolve(ctx context.Context, ref string) (secretresolver.Secret, error) {
	return f(ctx, ref)
}

func runtimeTLSMaterial(t *testing.T) ([]byte, []byte, *http.Client) {
	t.Helper()
	public, private, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	certificate := &x509.Certificate{SerialNumber: big.NewInt(107), Subject: pkix.Name{CommonName: "runtime local TLS"}, NotBefore: time.Now().Add(-time.Hour), NotAfter: time.Now().Add(time.Hour), IPAddresses: []net.IP{net.ParseIP("127.0.0.1")}, KeyUsage: x509.KeyUsageDigitalSignature, ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth}}
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
	transport := &http.Transport{TLSClientConfig: &tls.Config{RootCAs: pool, MinVersion: tls.VersionTLS12}, Proxy: nil}
	t.Cleanup(transport.CloseIdleConnections)
	return certPEM, keyPEM, &http.Client{Transport: transport, Timeout: 5 * time.Second}
}

func runtimeTLSSnapshot(t *testing.T, change func(*configurationversion.ConfigurationVersion)) runtimeconfig.Snapshot {
	t.Helper()
	snapshot, diagnostics := buildSnapshotForTest(func(version *configurationversion.ConfigurationVersion) {
		version.Listener.Port = availablePort(t)
		version.Listener.TLS = configurationversion.TLSSettings{Enabled: true, CertificateRef: "certificates/runtime-tls", PrivateKeyRef: "secrets/runtime-tls-key", MinVersion: "1.2"}
		version.Listener.Timeouts = configurationversion.TimeoutSettings{HandshakeSeconds: 1, WriteSeconds: 1}
		if change != nil {
			change(version)
		}
	})
	if len(diagnostics) != 0 {
		t.Fatal(diagnostics)
	}
	return snapshot
}

func runtimeTLSMemory(t *testing.T, cert, key []byte) *secretresolver.MemoryResolver {
	t.Helper()
	resolver, err := secretresolver.NewMemory(map[string][]byte{"certificates/runtime-tls": cert, "secrets/runtime-tls-key": key, "secrets/api-key/runtime": []byte("runtime-secret"), "secrets/jwt/runtime": []byte("runtime-jwt-secret")})
	if err != nil {
		t.Fatal(err)
	}
	return resolver
}

func startTLSHost(t *testing.T, snapshot runtimeconfig.Snapshot, resolver secretresolver.Resolver, handler message.Handler) *DefaultHost {
	t.Helper()
	host, err := NewHost(snapshot, resolver, handler)
	if err != nil {
		t.Fatal(err)
	}
	if err := host.Build(); err != nil {
		t.Fatal(err)
	}
	if err := host.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
		defer cancel()
		if err := host.Stop(ctx); err != nil {
			t.Errorf("TLS host cleanup: %v", err)
		}
	})
	return host
}

func dialTLSWebSocket(t *testing.T, snapshot runtimeconfig.Snapshot, client *http.Client, header http.Header) *websocket.Conn {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
	defer cancel()
	conn, response, err := websocket.Dial(ctx, "wss://127.0.0.1:"+portString(snapshot.Listener().Port)+"/ws", &websocket.DialOptions{HTTPClient: client, HTTPHeader: header})
	if err != nil {
		t.Fatalf("WSS dial: %v (response %v)", err, response)
	}
	return conn
}

func assertTLSEcho(t *testing.T, conn *websocket.Conn) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	for _, kind := range []websocket.MessageType{websocket.MessageText, websocket.MessageBinary} {
		payload := []byte("WSS echo")
		if kind == websocket.MessageBinary {
			payload = []byte{0, 1, 0xfe, 0xff}
		}
		if err := conn.Write(ctx, kind, payload); err != nil {
			t.Fatal(err)
		}
		got, data, err := conn.Read(ctx)
		if err != nil || got != kind || !bytes.Equal(data, payload) {
			t.Fatalf("WSS echo kind=%v data=%v err=%v", got, data, err)
		}
	}
}

func TestTLSRuntimeDisabledDoesNotResolveAndStartupContextIsNotLifetime(t *testing.T) {
	cert, key, client := runtimeTLSMaterial(t)
	for _, enabled := range []bool{false, true} {
		t.Run(fmt.Sprint(enabled), func(t *testing.T) {
			snapshot := runtimeTLSSnapshot(t, func(v *configurationversion.ConfigurationVersion) { v.Listener.TLS.Enabled = enabled })
			memory := runtimeTLSMemory(t, cert, key)
			var calls atomic.Int32
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			resolver := runtimeTLSResolverFunc(func(got context.Context, ref string) (secretresolver.Secret, error) {
				if got != ctx {
					t.Error("exact Start context not propagated")
				}
				calls.Add(1)
				return memory.Resolve(got, ref)
			})
			host, err := NewHost(snapshot, resolver, message.NewEchoHandler())
			if err != nil {
				t.Fatal(err)
			}
			if err := host.Build(); err != nil {
				t.Fatal(err)
			}
			if calls.Load() != 0 {
				t.Fatal("NewHost/Build resolved TLS")
			}
			if err := host.Start(ctx); err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = host.Stop(context.Background()) })
			cancel()
			want := int32(0)
			if enabled {
				want = 2
			}
			if calls.Load() != want {
				t.Fatalf("startup resolutions=%d want%d", calls.Load(), want)
			}
			if host.RuntimeContext().Err() != nil || !host.Ready() {
				t.Fatal("Start cancellation changed committed lifetime")
			}
			url := "ws://127.0.0.1:" + portString(snapshot.Listener().Port) + "/ws"
			options := &websocket.DialOptions{}
			if enabled {
				url = "wss" + strings.TrimPrefix(url, "ws")
				options.HTTPClient = client
			}
			dialContext, stop := context.WithTimeout(context.Background(), 3*time.Second)
			defer stop()
			conn, _, err := websocket.Dial(dialContext, url, options)
			if err != nil {
				t.Fatal(err)
			}
			assertTLSEcho(t, conn)
			conn.CloseNow()
			if calls.Load() != want {
				t.Fatal("TLS re-resolved on request")
			}
			if err := host.Stop(context.Background()); err != nil {
				t.Fatal(err)
			}
			if host.sessionManager.State() != sessionmanager.StateClosed {
				t.Fatal("Manager accounting not closed")
			}
		})
	}
}

func TestTLSRuntimePreparationFailureBeforeSocketAndRedactedIdentity(t *testing.T) {
	cert, key, _ := runtimeTLSMaterial(t)
	for _, stage := range []string{"certificate resolution", "key resolution", "malformed certificate", "malformed key", "canceled"} {
		t.Run(stage, func(t *testing.T) {
			snapshot := runtimeTLSSnapshot(t, nil)
			wantSnapshot := snapshot
			sentinel := errors.New("SECRET-RAW-RESOLVER-ERROR-PEM")
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			if stage == "canceled" {
				cancel()
			}
			resolver := runtimeTLSResolverFunc(func(_ context.Context, ref string) (secretresolver.Secret, error) {
				if stage == "certificate resolution" || stage == "key resolution" && ref == "secrets/runtime-tls-key" {
					return secretresolver.Secret{}, sentinel
				}
				value := cert
				if ref == "secrets/runtime-tls-key" {
					value = key
				}
				if stage == "malformed certificate" && ref == "certificates/runtime-tls" || stage == "malformed key" && ref == "secrets/runtime-tls-key" {
					value = []byte("SECRET-INVALID-PEM")
				}
				return secretresolver.Secret{Value: bytes.Clone(value)}, nil
			})
			host, err := NewHost(snapshot, resolver, nil)
			if err != nil {
				t.Fatal(err)
			}
			if err := host.Build(); err != nil {
				t.Fatal(err)
			}
			err = host.Start(ctx)
			if !errors.Is(err, ErrInvalidRuntimeConfiguration) || !errors.Is(err, listener.ErrInvalidListenerConfiguration) {
				t.Fatalf("TLS startup classifications: %v", err)
			}
			if strings.Contains(stage, "resolution") && !errors.Is(err, sentinel) {
				t.Fatal("lost raw resolver errors.Is identity")
			}
			if stage == "canceled" && !errors.Is(err, context.Canceled) {
				t.Fatal("lost cancellation identity")
			}
			for current := err; current != nil; current = errors.Unwrap(current) {
				for _, format := range []string{"%v", "%+v", "%#v", "%s", "%q"} {
					if strings.Contains(fmt.Sprintf(format, current), "SECRET") {
						t.Fatalf("unsafe error formatting %s", format)
					}
				}
			}
			if host.Ready() || host.CanAccept() || host.RuntimeContext() != nil || currentHostState(host) != hostBuilt {
				t.Fatal("TLS failure published lifecycle state")
			}
			if !reflect.DeepEqual(snapshot, wantSnapshot) {
				t.Fatal("Snapshot mutated")
			}
			assertPortAvailable(t, snapshot.Listener().Port)
		})
	}
}

func TestTLSRuntimeStartCancellationWhileResolverBlockedAndStopWaits(t *testing.T) {
	snapshot := runtimeTLSSnapshot(t, nil)
	entered := make(chan struct{})
	release := make(chan struct{})
	var once sync.Once
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	resolver := runtimeTLSResolverFunc(func(got context.Context, _ string) (secretresolver.Secret, error) {
		if got != ctx {
			t.Error("Start context identity")
		}
		once.Do(func() { close(entered) })
		<-release
		return secretresolver.Secret{}, got.Err()
	})
	host, err := NewHost(snapshot, resolver, nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := host.Build(); err != nil {
		t.Fatal(err)
	}
	start := make(chan error, 1)
	go func() { start <- host.Start(ctx) }()
	select {
	case <-entered:
	case <-time.After(time.Second):
		t.Fatal("resolver entry")
	}
	stop := make(chan error, 1)
	go func() { stop <- host.Stop(context.Background()) }()
	if host.Ready() || host.CanAccept() {
		t.Fatal("starting TLS was admitted")
	}
	select {
	case err := <-stop:
		t.Fatalf("Stop bypassed Starting wait: %v", err)
	default:
	}
	cancel()
	close(release)
	select {
	case err := <-start:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("Start cancellation: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Start did not return")
	}
	select {
	case err := <-stop:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Stop did not return")
	}
	assertPortAvailable(t, snapshot.Listener().Port)
}

func TestTLSRuntimeWSSAuthenticationAndEcho(t *testing.T) {
	cert, key, client := runtimeTLSMaterial(t)
	for _, kind := range []string{"api-key", "jwt"} {
		t.Run(kind, func(t *testing.T) {
			snapshot := runtimeTLSSnapshot(t, func(v *configurationversion.ConfigurationVersion) {
				v.Authentication = configurationversion.AuthenticationSettings{Enabled: true, Providers: []configurationversion.AuthenticationProvider{{Name: "tls-auth", Type: configurationversion.AuthenticationProviderAPIKey, Enabled: true, APIKey: &configurationversion.APIKeySettings{Header: "X-API-Key", SecretRef: "secrets/api-key/runtime"}}}}
				if kind == "jwt" {
					v.Authentication.Providers = []configurationversion.AuthenticationProvider{{Name: "tls-jwt", Type: configurationversion.AuthenticationProviderJWT, Enabled: true, JWT: &configurationversion.JWTSettings{SigningKeys: []configurationversion.JWTSigningKey{{Name: "primary", SecretRef: "secrets/jwt/runtime"}}, AllowedAlgorithms: []configurationversion.JWTAlgorithm{configurationversion.HS256}}}}
				}
			})
			host := startTLSHost(t, snapshot, runtimeTLSMemory(t, cert, key), message.NewEchoHandler())
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()
			conn, response, err := websocket.Dial(ctx, "wss://127.0.0.1:"+portString(snapshot.Listener().Port)+"/ws", &websocket.DialOptions{HTTPClient: client})
			if conn != nil {
				conn.CloseNow()
				t.Fatal("missing credentials upgraded")
			}
			if err == nil || response == nil || response.StatusCode != http.StatusUnauthorized {
				t.Fatalf("pre-Upgrade reject: response%v err%v", response, err)
			}
			response.Body.Close()
			header := make(http.Header)
			header.Set("X-API-Key", "runtime-secret")
			if kind == "jwt" {
				token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"sub": "tls-client", "exp": time.Now().Add(time.Minute).Unix()}).SignedString([]byte("runtime-jwt-secret"))
				if err != nil {
					t.Fatal(err)
				}
				header = make(http.Header)
				header.Set("Authorization", "Bearer "+token)
			}
			conn = dialTLSWebSocket(t, snapshot, client, header)
			assertTLSEcho(t, conn)
			conn.CloseNow()
			if err := host.Stop(ctx); err != nil {
				t.Fatal(err)
			}
			if host.sessionManager.State() != sessionmanager.StateClosed {
				t.Fatal("WSS Manager convergence")
			}
		})
	}
}

func TestTLSRuntimeReadAndIdleTimeoutsPreserveRoot(t *testing.T) {
	cert, key, client := runtimeTLSMaterial(t)
	for _, kind := range []string{"read", "idle"} {
		t.Run(kind, func(t *testing.T) {
			snapshot := runtimeTLSSnapshot(t, func(v *configurationversion.ConfigurationVersion) {
				if kind == "read" {
					v.Listener.Timeouts.ReadSeconds = 1
				} else {
					v.Listener.Timeouts.IdleSeconds = 1
				}
			})
			host := startTLSHost(t, snapshot, runtimeTLSMemory(t, cert, key), message.NewEchoHandler())
			conn := dialTLSWebSocket(t, snapshot, client, nil)
			defer conn.CloseNow()
			assertTLSEcho(t, conn)
			ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
			defer cancel()
			if _, _, err := conn.Read(ctx); err == nil || ctx.Err() != nil {
				t.Fatalf("server timeout expected: %v caller%v", err, ctx.Err())
			}
			if !host.Ready() || !host.CanAccept() || host.RuntimeContext().Err() != nil {
				t.Fatal("Session policy changed Runtime lifetime")
			}
			if err := host.Stop(ctx); err != nil {
				t.Fatal(err)
			}
			if host.sessionManager.State() != sessionmanager.StateClosed {
				t.Fatal("timeout accounting not joined")
			}
		})
	}
}

type tlsLargeReply struct {
	reply  message.Message
	result chan error
}

func (h tlsLargeReply) Handle(ctx context.Context, input message.Context) error {
	err := input.Sender().Send(ctx, h.reply)
	h.result <- err
	return err
}

func TestTLSRuntimeWriteTimeoutOnNonreadingWSSPeer(t *testing.T) {
	cert, key, client := runtimeTLSMaterial(t)
	snapshot := runtimeTLSSnapshot(t, nil)
	reply, err := message.New(message.TypeBinary, make([]byte, 16<<20))
	if err != nil {
		t.Fatal(err)
	}
	result := make(chan error, 1)
	host := startTLSHost(t, snapshot, runtimeTLSMemory(t, cert, key), tlsLargeReply{reply: reply, result: result})
	conn := dialTLSWebSocket(t, snapshot, client, nil)
	defer conn.CloseNow()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := conn.Write(ctx, websocket.MessageText, []byte("send-large-reply")); err != nil {
		t.Fatal(err)
	}
	select {
	case err := <-result:
		if !errors.Is(err, context.DeadlineExceeded) {
			t.Fatalf("WriteSeconds identity on blocked TLS Write: %v", err)
		}
	case <-ctx.Done():
		t.Fatal("TLS Write exceeded configured budget")
	}
	conn.CloseNow()
	if err := host.Stop(ctx); err != nil {
		t.Fatal(err)
	}
	if host.sessionManager.State() != sessionmanager.StateClosed {
		t.Fatal("write timeout accounting not joined")
	}
}

func TestTLSRuntimeConcurrentWSSAndStopConverge(t *testing.T) {
	cert, key, client := runtimeTLSMaterial(t)
	snapshot := runtimeTLSSnapshot(t, nil)
	host := startTLSHost(t, snapshot, runtimeTLSMemory(t, cert, key), message.NewEchoHandler())
	var clients []*websocket.Conn
	for range 4 {
		conn := dialTLSWebSocket(t, snapshot, client, nil)
		assertTLSEcho(t, conn)
		clients = append(clients, conn)
	}
	defer func() {
		for _, conn := range clients {
			conn.CloseNow()
		}
	}()
	var calls atomic.Int32
	current := host.runtimeListener.(*listener.DefaultListener)
	_ = current
	// Existing Host shutdown tests count the one Listener Stop owner. Here real WSS
	// clients prove that the same Manager wait set converges through TLS transport.
	var group sync.WaitGroup
	failures := make(chan error, 8)
	for range 8 {
		group.Add(1)
		go func() {
			defer group.Done()
			ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
			defer cancel()
			failures <- host.Stop(ctx)
			calls.Add(1)
		}()
	}
	group.Wait()
	close(failures)
	for err := range failures {
		if err != nil {
			t.Fatal(err)
		}
	}
	if calls.Load() != 8 || host.sessionManager.State() != sessionmanager.StateClosed || host.Ready() || host.CanAccept() {
		t.Fatal("concurrent TLS shutdown did not converge")
	}
	assertPortAvailable(t, snapshot.Listener().Port)
}

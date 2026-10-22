# Runtime TLS execution v1 — TASK-107

[Russian version](../../ru/design/runtime-tls-enforcement.md)

Design status: Architecture Confirmation PASS, recorded in TASK-107 E-107-003.
Implementation status: implemented candidate; verification and acceptance resolve
from the TASK-107 Recovery Evidence Envelope. This focused addition does not amend
Frozen ARCH-002, Active ARCH-003/005 or Approved DP-003/004.

## Startup and material ownership

Certificate input is a leaf-first PEM chain; key input is unencrypted PEM
accepted by tls.X509KeyPair. Encrypted/unsupported key input fails closed; no
password, DER/JWK conversion or file-path fallback is introduced.

One certificate chain/private key pair is resolved once per Host.Start attempt
from the copied Snapshot through the existing borrowed Resolver. Disabled TLS
never resolves certificate/key references and preserves the plaintext path.
NewHost, Bootstrap and Build do not resolve Secrets or open sockets.

The private production composer receives the exact Start context; existing
injected test composer signatures may remain through adapters. Production checks
ctx.Err before resolution, between certificate and key, after resolution/parsing
and before socket acquisition. Resolver receives the exact Start context. No
detached goroutine forces an uncooperative Resolver. Stop during Starting waits
Start as before and gains no new startup cancellation authority. After successful
startup, Start context is not retained as Runtime/Listener lifetime context.

Enabled scalar/reference metadata is validated with existing validators before
resources. Resolve certificate then key once each without retry/fallback; own
returned bytes and defer clearing raw PEM buffers on success and failure. Parse
with tls.X509KeyPair and parse every supplied chain certificate with x509.
Reject empty/malformed/mismatched data, expired/future leaf, other-only extended
key usage and nonzero key usage lacking DigitalSignature. Empty EKU, Any and
ServerAuth are accepted. Startup does not verify DNS against bind Host, chain
trust/roots or revocation; trusted clients verify hostname and trust. Standard
library supported key parsing/signing supplies crypto policy.

A focused Listener transport helper prepares opaque TLS material during Host
composition. A new explicit TLS-aware Listener Bootstrap accepts that material
plus existing Handshake/reporter dependencies; existing constructors retain
their signatures and disabled behavior. Enabled TLS without prepared material
rejects. Prepared configuration is private/immutable, cloned for Listener use
without mutable caller accessors. Raw Secrets never enter Snapshot, provenance,
Configuration or Container. Failed composition releases parsed material; terminal
Listener Stop releases its references where safe. No secure zeroization of Go
cryptographic objects or caller-retained material is promised.

## Protocol and timeout execution

Configuration has exactly one certificate pair, minimum TLS1.2 or TLS1.3 from
Snapshot, maximum TLS1.3, NextProtos only http/1.1, NoClientCert and no certificate
selection callbacks, key logging or name map. Server.Protocols explicitly enables
HTTP1 only and TLSNextProto is empty. No ALPN is compatible with HTTP1; an h2-only
offer fails negotiation. There is no rotation/reload, ACME, SNI selection or mTLS.

Listener wraps its TCP socket with tls.NewListener and uses existing Server.Serve,
preserving concrete tls.Conn handling. With TLS only, ReadHeaderTimeout equals
H = validated HandshakeSeconds (1..300). ReadTimeout/WriteTimeout/IdleTimeout
remain unset: Session settings must not become HTTP server deadlines.
Go1.26.5 net/http uses the minimum positive read-header/read/write timeout for
TLS handshake read/write deadlines and clears them after success. Thus TLS
handshake has one non-resettable budget H; HTTP header reading has a separate H;
the existing request/pre-Upgrade Handshake context has its existing separate H.
This is not one shared TCP-accept-to-Upgrade budget. ReadSeconds/WriteSeconds/
IdleSeconds retain their existing Session semantics. No product timer/watchdog
or goroutine is added.

## Shutdown and failures

Frozen/current ordering remains close Admission -> Manager snapshot RequestStop
-> root cancellation -> Listener Stop (serverStop, Shutdown(ctx), listener close,
serve/handler joins) -> Manager.Wait. net/http owns connections until Hijack;
Session acquires sole ownership only at existing Commit. Listener BaseContext
cancellation reaches TLS HandshakeContext; the standard library closes pending
TLS connections, and H bounds transport work. No custom connection accounting
or new lifecycle state is introduced. Deadline-returning Shutdown/Stop does not
prove all net/http internal goroutines joined; the existing cooperative shutdown
limitation remains. Successful Shutdown completes non-hijacked connection shutdown.
One product Stop owner remains; defensive standard library closes are not a
second product cleanup owner.

TLS preparation failure uses a private safe error with fixed stage text. Unwrap
exposes only existing ErrInvalidListenerConfiguration classification; Is delegates
to the retained original cause, preserving errors.Is for resolver sentinels,
parser errors and context cancellation. Raw cause is not exposed by ordinary
Unwrap/As, Error or formatting. No underlying errors.As promise is made. Runtime
adds existing ErrInvalidRuntimeConfiguration classification with safe wrapping;
unrelated startup/rollback error semantics remain unchanged. No reference, PEM,
resolver payload or raw parser text is emitted. TLS-only Server.ErrorLog discards
unstructured standard library diagnostics; existing safe Serve terminal reporting
remains. Peer TLS rejection is per-connection, not a new Host/terminal category.

Plaintext on a TLS port never reaches application Handler, Upgrade/101 or Session.
The standard library may return its fixed HTTP400 diagnostic to such a client;
this is not plaintext application fallback. Startup preparation failures precede
socket acquisition, leave readiness/admission false and publish no Runtime root.
No new Configuration fields, dependencies, Secret backend or activation is added.

## Required verification

Disabled resolver spy; resolve/parse/mismatch/expiry/EKU/usage/cancel negatives;
failure-before-socket and safe errors.Is/redaction; TLS1.2/1.3 minimum boundaries,
h2-only/no-ALPN/plaintext rejection; silent/partial handshake expiry and Stop;
trusted real WSS text/binary/echo, API Key/JWT pre-Upgrade and read/write/idle
compatibility; concurrent Start/Stop/admission, rollback, single cleanup and
Manager convergence. Full regression, independent roles and exact manifest are
required; historical CI is not candidate evidence. Gate 9 closure requires its
own complete evidence assessment and is not implied by timeout work or this note.

# Исполнение Runtime TLS v1 — TASK-107

[English version](../../en/design/runtime-tls-enforcement.md)

Design status: Architecture Confirmation PASS, запись TASK-107 E-107-003.
Implementation status: implemented candidate; verification и acceptance
определяются Recovery Evidence Envelope TASK-107. Focused addition не изменяет Frozen ARCH-002,
Active ARCH-003/005 или Approved DP-003/004.

## Startup и владение материалом

Certificate input — leaf-first PEM chain; key input — unencrypted PEM,
принимаемый tls.X509KeyPair. Encrypted/unsupported key input fail-closed; password,
DER/JWK conversion или file-path fallback не добавляются.

Одна certificate chain/private key pair разрешается один раз на попытку
Host.Start из скопированного Snapshot через существующий borrowed Resolver.
Disabled TLS никогда не разрешает certificate/key references и сохраняет
plaintext path. NewHost, Bootstrap и Build не разрешают Secrets и не открывают socket.

Private production composer получает exact Start context; существующие injected
test composer signatures могут сохраняться через adapters. Production проверяет
ctx.Err до resolution, между certificate и key, после resolution/parsing и перед
socket acquisition. Resolver получает exact Start context. Detached goroutine
для принудительной остановки uncooperative Resolver нет. Stop во время Starting
как прежде ожидает Start и не получает новой startup cancellation authority.
После успешного startup Start context не является lifetime context Runtime/Listener.

Enabled scalar/reference metadata проверяется существующими validators до ресурсов.
Resolve certificate, затем key, по одному разу без retry/fallback; returned bytes
принадлежат caller, raw PEM buffers очищаются defer на success и failure.
Используются tls.X509KeyPair и x509 parsing каждого supplied chain certificate.
Отклоняются empty/malformed/mismatched data, expired/future leaf, other-only EKU
и nonzero KeyUsage без DigitalSignature. Empty EKU, Any и ServerAuth допустимы.
Startup не проверяет DNS по bind Host, chain trust/roots или revocation: hostname
и trust проверяет trusted client. Crypto policy — supported key parsing/signing
стандартной библиотеки.

Focused transport helper Listener подготавливает opaque TLS material во время
Host composition. Новый explicit TLS-aware Listener Bootstrap получает material
и прежние Handshake/reporter dependencies; существующие constructors сохраняют
signatures и disabled behavior. Enabled TLS без material отклоняется. Prepared
configuration private/immutable, клонируется для Listener без mutable caller
accessors. Raw Secrets не входят в Snapshot, provenance, Configuration, Container.
Failed composition освобождает parsed material; terminal Listener Stop освобождает
свои references там, где это безопасно. Secure zeroization Go cryptographic
objects или caller-retained material не обещается.

## Protocol и timeout execution

Configuration содержит ровно одну certificate pair, minimum TLS1.2 или TLS1.3 из
Snapshot, maximum TLS1.3, NextProtos только http/1.1, NoClientCert; certificate
selection callbacks, key logging и name map отсутствуют. Server.Protocols явно
включает только HTTP1, TLSNextProto пуст. Отсутствие ALPN совместимо с HTTP1;
h2-only offer отклоняется. Rotation/reload, ACME, SNI selection и mTLS отсутствуют.

Listener оборачивает TCP socket в tls.NewListener и использует прежний Server.Serve,
сохраняя concrete tls.Conn handling. Только для TLS ReadHeaderTimeout равен
H = проверенному HandshakeSeconds (1..300). ReadTimeout/WriteTimeout/IdleTimeout
не заданы: Session settings не становятся deadlines HTTP server.
Go1.26.5 net/http использует minimum positive read-header/read/write timeout для
TLS handshake read/write deadlines и очищает их после success. TLS handshake
имеет один non-resettable budget H; HTTP headers имеют отдельный H; прежний
request/pre-Upgrade Handshake context сохраняет отдельный H. Это не единый budget
от TCP accept до Upgrade. ReadSeconds/WriteSeconds/IdleSeconds сохраняют прежнюю
Session semantics. Product timer/watchdog или goroutine не добавляются.

## Shutdown и failures

Frozen/current ordering сохраняется: close Admission -> Manager snapshot
RequestStop -> root cancellation -> Listener Stop (serverStop, Shutdown(ctx),
listener close, serve/handler joins) -> Manager.Wait. net/http владеет connections
до Hijack; Session получает sole ownership только при существующем Commit.
Listener BaseContext cancellation достигает TLS HandshakeContext; standard library
закрывает pending TLS connections, H ограничивает transport work. Custom connection
accounting и новых lifecycle states нет. Shutdown/Stop, вернувший deadline, не
доказывает join всех внутренних goroutines net/http: прежняя cooperative shutdown
limitation сохраняется. Успешный Shutdown завершает non-hijacked connection shutdown.
Остаётся один product Stop owner; defensive standard library Close не является
вторым владельцем product cleanup.

TLS preparation failure использует private safe error с fixed stage text. Unwrap
раскрывает только прежнюю classification ErrInvalidListenerConfiguration; Is
делегирует retained original cause, сохраняя errors.Is для resolver sentinel,
parser error и context cancellation. Raw cause недоступен обычным Unwrap/As,
Error или formatting. Обещания underlying errors.As нет. Runtime добавляет прежнюю
classification ErrInvalidRuntimeConfiguration безопасным wrapping; unrelated
startup/rollback error semantics не меняется. References, PEM, resolver payload
и raw parser text не выводятся. TLS-only Server.ErrorLog отбрасывает unstructured
standard library diagnostics; прежний safe Serve terminal reporting сохраняется.
Peer TLS rejection относится к connection, не создаёт Host/terminal category.

Plaintext на TLS port не достигает application Handler, Upgrade/101 или Session.
Standard library может вернуть fixed HTTP400 diagnostic; это не plaintext
application fallback. Startup preparation failure предшествует socket acquisition,
оставляет readiness/admission false и не публикует Runtime root. Новых Configuration
fields, dependencies, Secret backend или activation нет.

## Необходимые проверки

Disabled resolver spy; resolve/parse/mismatch/expiry/EKU/usage/cancel negatives;
failure-before-socket и safe errors.Is/redaction; minimum boundaries TLS1.2/1.3,
h2-only/no-ALPN/plaintext rejection; silent/partial handshake expiry и Stop;
trusted real WSS text/binary/echo, API Key/JWT pre-Upgrade, read/write/idle
compatibility; concurrent Start/Stop/admission, rollback, single cleanup и Manager
convergence. Нужны full regression, independent roles и exact manifest; historical
CI не является candidate evidence. Закрытие Gate 9 требует отдельной полной
evidence assessment и не следует автоматически из timeout work или этой note.

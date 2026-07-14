# DP-001: Runtime Handshake Pipeline

[Russian version](../../ru/design/DP-001-runtime-handshake-pipeline.md)

## 1. Status

**Status:** Draft

This document proposes a Runtime Handshake architecture. It contains no implementation contract, Go interface, package declaration, or commitment to future policy APIs.

## 2. Background

The Alpha Runtime has a working WebSocket vertical. Listener owns the TCP listener and HTTP server, handles `GET /ws`, calls the WebSocket library to accept the Upgrade, and then transfers the upgraded connection through a Dispatcher. AuthenticationDispatcher converts selected HTTP request metadata into a transport-neutral `AuthenticationRequest`, invokes Authentication Service, and either closes the upgraded connection or transfers it with a Principal to Session Dispatcher. Session then owns the WebSocket connection and starts its read loop.

The current request path is:

```text
TCP accept
    -> HTTP request
    -> GET /ws handler
    -> WebSocket Upgrade (101)
    -> ConnectionContext with WebSocket connection
    -> AuthenticationDispatcher
    -> Authentication Service and Providers
    -> reject with WebSocket close
       or
       AuthenticatedContext
    -> Session creation and lifecycle
```

This path is verified by implementation and integration tests. The [Runtime Alpha Architecture Review](../reviews/runtime-alpha-review.md) records post-Upgrade Authentication as High finding F-01. It also identifies Runtime Host composition and lifecycle hardening as related work that must not be hidden inside Handshake.

[DP-001: Authentication](../proposals/DP-001-authentication.md) already states that required identity checks complete before a WebSocket Session is opened and that rejected credentials produce HTTP `401`. The implementation does not yet satisfy that design.

## 3. Problem Statement

Authentication after Upgrade is an architecture problem, not merely an HTTP status limitation.

### Ownership

After `websocket.Accept` succeeds, the HTTP exchange has ended and an upgraded mutable connection exists. AuthenticationDispatcher must own and close that connection even for clients that should never have been admitted. A security decision is therefore made after transport ownership has already crossed the admission boundary.

### Transport Semantics

Before Upgrade, Runtime can express rejection as an HTTP response. After Upgrade, the same outcome must be encoded as a WebSocket close. Clients observe a successful `101` followed by closure rather than a rejected handshake. This makes Authentication semantics depend on when transport conversion occurred.

### Lifecycle

Post-Upgrade evaluation creates WebSocket resources before admission completes. Cancellation, Provider failure, downstream construction failure, and Listener shutdown must all close a connection that did not need to exist. The connection-specific lifecycle begins too early.

### Observability

An accepted transport followed by policy rejection is harder to classify. Upgrade success, Authentication rejection, Provider outage, and Session startup failure can be conflated unless every later close is interpreted with hidden pipeline knowledge.

### Extensibility

Future Origin, maintenance, rate-limit, or IP filtering decisions would face the same choice: either execute after Upgrade and allocate resources unnecessarily, or be embedded directly into Listener. Neither provides a stable admission boundary.

### Future Policies

Adding each policy at a different transport stage would make order, short-circuit behavior, ownership, and error mapping unpredictable. A conceptual Handshake boundary is needed, but it must not become a universal Policy Engine.

## 4. Existing Architecture

```text
Listener
  owns TCP listener and HTTP server
        |
        v
HTTP /ws Handler
  validates method
        |
        v
websocket.Accept
  commits HTTP 101 and creates WebSocket connection
        |
        v
Connection Dispatcher
  transfers upgraded transport ownership
        |
        v
Authentication Dispatcher
  copies HTTP metadata
  invokes transport-neutral Authentication
  closes on rejection/error
        |
        v
Authenticated Dispatcher
  transfers connection and Principal
        |
        v
Session
  owns WebSocket, read loop, and serialized writes
```

The responsibilities below should be preserved:

- Listener owns network acceptance and HTTP/WebSocket execution.
- Authentication owns credential evaluation and remains independent of `net/http` and WebSocket.
- Connection handoff makes mutable transport ownership explicit.
- Session is created only for an authenticated Principal and owns the accepted WebSocket thereafter.
- Runtime Snapshot supplies effective Configuration without Repository access.

The ordering of Upgrade and Authentication is the boundary that must change.

## 5. Goals

- Complete required Authentication before WebSocket Upgrade.
- Preserve transport-neutral AuthenticationRequest, Authentication Service, and Providers.
- Preserve Runtime independence from Control Plane repositories and management HTTP API.
- Provide a clear conceptual location for future handshake-scoped Evaluation.
- Make ownership of HTTP and WebSocket resources explicit at every outcome.
- Keep connection and Session lifecycles predictable under rejection, cancellation, and failure.
- Distinguish negative admission decisions from internal and dependency failures.
- Keep Listener responsible for transport execution without containing Provider-specific logic.
- Remain compatible with the current Snapshot, Authentication, and Session vertical.

## 6. Non Goals

This document does not design:

- Router;
- Session Manager;
- Middleware;
- Plugin ABI;
- another transport or a transport abstraction;
- QUIC;
- HTTP/3;
- OAuth;
- mTLS implementation;
- Provider-specific credential verification;
- an Origin, rate-limit, maintenance, or IP filtering API;
- Runtime Host composition details.

## 7. Constraints

The design follows [ARCH-001: Runtime Architectural Pattern](../architecture/ARCH-001-runtime-architectural-pattern.md).

### Configuration First

Handshake behavior comes from Published Configuration through an immutable Runtime Snapshot. Unsupported active settings are rejected explicitly rather than ignored.

### Ownership

Every mutable transport resource has one current owner. Transfer is explicit, and failure before or during transfer has a defined closer.

### Lifecycle

Evaluation observes request cancellation. Long-running work and network I/O do not execute under lifecycle locks. Handshake completes before Session lifecycle begins.

### No Magic

Dependencies are composed explicitly. Handshake does not discover Providers, read repositories, use global state, or use `context.Value` as a service locator.

### Boring Core

The design reuses existing Authentication and transport boundaries. It does not create a generic pipeline framework, universal Policy Engine, or common Decision model for unrelated subsystems.

## 8. Design Alternatives

### Alternative A — Keep the Current Architecture

Authentication remains after `websocket.Accept` in AuthenticationDispatcher.

**Advantages:**

- No implementation change.
- Current integration tests and ownership transfer remain intact.
- Authentication has access to normalized request metadata already associated with a WebSocket connection.

**Disadvantages:**

- Rejected credentials receive `101` before closure.
- Unauthenticated clients consume upgraded connection resources.
- Admission ownership begins before the admission decision.
- Future handshake policies would also run too late or require separate placement.
- It conflicts with the existing Authentication proposal and Alpha review finding.

**Reasons rejected:** The design preserves the exact security and lifecycle boundary that TASK-BETA-003 must correct.

### Alternative B — Put Authentication Logic Inside Listener

Listener directly reads Authentication Configuration, selects Providers, verifies credentials, and decides whether to call Upgrade.

**Advantages:**

- Authentication can happen before Upgrade.
- HTTP rejection is straightforward.
- Fewer visible components in the request path.

**Disadvantages:**

- Listener gains JWT, API Key, Basic, Registry, and Provider orchestration knowledge.
- Authentication becomes coupled to HTTP and WebSocket transport behavior.
- Provider extensibility requires Listener changes.
- Unit boundaries and dependency direction from ADR-003 are lost.

**Reasons rejected:** It fixes ordering by violating single responsibility, transport-neutral Authentication, and Boring Core.

### Alternative C — Put Authentication Inside Session

Listener upgrades the connection and creates a provisional Session that authenticates before normal message processing.

**Advantages:**

- Session could centralize connection lifecycle after Upgrade.
- Authentication and later message processing would share one connection owner.
- Listener would remain unaware of Providers.

**Disadvantages:**

- Authentication still occurs after Upgrade.
- Session would exist before a valid Principal, contradicting its current invariant.
- Session would need HTTP handshake metadata or another hidden transport bridge.
- Rejection remains a WebSocket close and cannot be an HTTP response.
- Session responsibility expands into admission policy.

**Reasons rejected:** It moves the problem deeper into Runtime and breaks the established authenticated-Session boundary.

### Alternative D — Dedicated Handshake Pipeline

A conceptual Handshake boundary normalizes request metadata, performs configured Evaluation, produces an admission Decision, and only then asks the transport owner to reject or Upgrade.

**Advantages:**

- Authentication completes before WebSocket allocation.
- Authentication Service and Providers remain transport-neutral.
- Listener retains HTTP/WebSocket execution without Provider-specific logic.
- Ownership transfer occurs only after an allow Decision.
- Future handshake-scoped evaluations have one ordered boundary.
- Negative decisions and operational errors can have distinct HTTP semantics.

**Disadvantages:**

- The request path gains an explicit orchestration stage.
- Migration must avoid duplicate Authentication and double ownership.
- Ordering and failure semantics for future evaluations require focused design.
- Slow evaluations retain HTTP request resources until completion.

**Reasons selected:** It corrects the security boundary while preserving the current responsibilities and extension seams. “Pipeline” is a conceptual sequence, not a generic framework or a declaration of Go interfaces.

## 9. Proposed Architecture

```text
Runtime Snapshot and composed dependencies
                 |
                 v
Listener / HTTP transport owner
  accepts request and owns ResponseWriter
                 |
                 v
Handshake Context normalization
  copies only required metadata
                 |
                 v
Handshake Evaluation
  required method/path checks
  configured Authentication
  future handshake-scoped checks
                 |
                 v
Handshake Decision
          /                 \
         v                   v
reject or fail              allow
HTTP response                 |
no WebSocket                  v
                        WebSocket Upgrade
                              |
                              v
                    authenticated handoff
                              |
                              v
                           Session
```

The architecture is conceptual. It does not prescribe a Handshake package, interface, method, or universal list of evaluators.

Listener remains the transport executor. It supplies request metadata for normalization and applies the final transport effect. Authentication remains an existing transport-neutral service. A successful result supplies Principal data to the post-Upgrade handoff. Session remains unchanged in purpose: it is created only after successful admission and owns the WebSocket connection.

Future Handshake policies may be composed at the Evaluation stage only when their Configuration and focused design exist. They do not become generic Runtime policies automatically.

## 10. Handshake Stages

```text
Transport
    -> Handshake Context
    -> Evaluation
    -> Decision
    -> Upgrade
    -> Session
```

### Transport

Listener and HTTP server accept the request. They own transport execution and expose only the request data required for Handshake.

### Handshake Context

Required metadata is copied from the HTTP request into narrow evaluation input. Credential-bearing values remain request-scoped and are not stored in Session or Snapshot.

### Evaluation

Evaluation applies the configured admission checks. Authentication uses the existing Authentication Service and Providers. “Evaluation” is the architecture term from ARCH-001, not a mandatory framework, engine, or Go abstraction.

### Decision

The result distinguishes allow, expected reject, Configuration failure, dependency failure, cancellation, and internal failure as needed by the subsystem. This document does not define one global Decision type or its exact representation.

### Upgrade

Only an allow Decision permits WebSocket Upgrade. Listener performs the transport operation. A failure during Upgrade is a transport failure, not an Authentication rejection.

### Session

After successful Upgrade, the connection and copied authenticated Principal are transferred downstream. Session is then created and its connection lifecycle begins.

## 11. Ownership Model

### HTTP Request

The HTTP server owns the request lifecycle. The Handshake path borrows the request synchronously and copies only metadata needed for Evaluation. It must not retain `http.Request`, its Body, Headers, URL, or cancellation context after the handler completes. Authentication Providers continue to receive normalized data, not the request object.

### ResponseWriter

The HTTP handler borrows ResponseWriter from the HTTP server and is the only Handshake-side component allowed to apply the HTTP Decision. Evaluation does not write responses. Before Upgrade, the handler may send one rejection or failure response. After a response is committed, Upgrade is forbidden.

After successful Upgrade, HTTP response semantics are no longer available. The handler must not attempt another HTTP write.

### TCP Connection

Listener and HTTP server own the accepted TCP connection before Upgrade. Handshake Evaluation does not access or close the raw connection. On HTTP rejection, normal HTTP server semantics determine whether the transport can be reused or closed. On Listener shutdown, the server remains responsible for active pre-Upgrade requests.

### WebSocket Connection

No WebSocket connection exists during Evaluation. A successful Upgrade creates it under the transport execution boundary. That boundary temporarily owns the connection and must either:

- transfer it exactly once to the authenticated downstream path; or
- close it if handoff or Session creation fails.

Authentication rejection cannot own or close a WebSocket because none has been created.

### Authentication Result

Authentication Service produces result data during Evaluation. The Handshake path owns that result for the request duration. Credential values and resolved Secrets are not part of the result. On success, the Principal is copied into the allow Decision or authenticated handoff so downstream mutation cannot alter the evaluated identity.

The result owns no transport resource. A negative result leads to HTTP rejection without connection transfer.

### Session

Session does not exist before an allow Decision and successful Upgrade. Once constructed successfully, Session becomes the sole owner of the WebSocket connection, read loop, and write serialization. If Session construction fails, the pre-Session handoff owner closes the connection.

Session stores the Principal and minimal connection metadata required by its contract. It does not retain the HTTP request, ResponseWriter, credential Headers, Query, or Handshake Context.

### Ownership Timeline

```text
HTTP server owns TCP and request
        |
        | Handshake borrows request/ResponseWriter
        | Evaluation owns copied metadata and result values
        v
allow Decision
        |
        | transport boundary creates and temporarily owns WebSocket
        v
authenticated handoff
        |
        | Session construction succeeds
        v
Session owns WebSocket until Stop
```

Every failure path terminates before the next transfer or closes the resource owned at that point.

## 12. Error Model

### Before Upgrade

Before Upgrade, failures use HTTP semantics and do not create a WebSocket connection.

- Invalid request shape or unsupported method is an HTTP request error.
- Rejected credentials are an expected negative Authentication decision and produce a safe authentication rejection, consistent with the existing Authentication proposal.
- A future Origin or admission policy rejection is an expected negative decision with policy-appropriate HTTP semantics.
- Cancellation stops Evaluation and must not be reported as successful admission.
- Provider or dependency failure is operational, observable, and distinct from rejected credentials.
- Internal failure returns a generic response without implementation details.
- Configuration failure should normally prevent Listener startup. If detected per request, the system fails closed and reports it operationally.

Exact response bodies, challenge headers, and status mapping beyond already accepted behavior require implementation-focused design. Sensitive values never appear in a response.

### During Upgrade

Failure while performing Upgrade is a transport failure. It is not reclassified as Authentication rejection. Listener retains responsibility for safe diagnostics and cleanup according to the WebSocket library contract.

### After Upgrade

HTTP errors are no longer possible. Failures during authenticated handoff, Session creation, or Session execution use WebSocket close or immediate transport cleanup according to the component lifecycle. They must not be represented as pre-Upgrade policy rejection.

### Error Separation

Expected negative Decision, Configuration failure, dependency failure, cancellation, internal failure, Upgrade failure, and Session failure remain distinguishable for observability even when client-facing details are intentionally generic.

## 13. Lifecycle

The overall Runtime and Listener are already Running before a request enters Handshake. Handshake is a bounded per-request lifecycle within the HTTP handler:

```text
request accepted
    -> context normalized
    -> Evaluation started
    -> Decision produced
       -> rejected/failed: HTTP response, Handshake ends
       -> allowed: Upgrade attempted
          -> failed: transport cleanup, Handshake ends
          -> succeeded: authenticated handoff
             -> Session created and started
             -> Handshake ownership ends
```

An HTTP rejection remains possible until a response or Upgrade is committed. Evaluation must not commit either.

The connection-specific Session lifecycle begins only after allow, successful Upgrade, and successful transfer. The Session read loop never runs concurrently with unresolved Handshake Evaluation for the same connection.

Listener shutdown cancels active Handshake contexts and prevents new admissions. Evaluation must propagate cancellation to Authentication Providers. Shutdown waits and error propagation follow the lifecycle requirements in ARCH-001 and the debt identified by Alpha finding F-04.

## 14. Security

### Origin

Current WebSocket origin handling uses library defaults. A future explicit Origin Policy belongs before Upgrade in Handshake Evaluation and must be Configuration First. This document does not define its schema, proxy model, or matching rules.

### Credentials

Only required credential-bearing request fields are copied into AuthenticationRequest. Providers resolve Secret References close to use. Raw credentials, tokens, API keys, passwords, and resolved Secrets are not retained after Evaluation.

### Logging and Observability

Logs, metrics, traces, and errors must not contain credential Headers, tokens, query credentials, Secret values, private key material, or full request dumps. Operational signals distinguish rejection from failure using safe categories and bounded labels.

### Sensitive Data Lifetime

Handshake metadata and Authentication results are request-scoped. Secret byte ownership follows SecretResolver contracts and Provider cleanup. Session receives Principal identity data, not credential input.

### Configuration

Only Published Snapshot data controls Handshake. Runtime does not read Control Plane repositories. Unsupported Authentication or future admission settings fail explicitly before traffic where possible.

## 15. Future Extension Points

The proposed boundary can later host focused, configured evaluations such as:

- Origin Policy;
- maintenance admission;
- rate limiting;
- IP filtering;
- enterprise-specific admission extensions.

These are future possibilities, not implemented features or API commitments. Each requires Configuration metadata, validation, explicit ordering and failure semantics, and focused design. They must not be folded into a Universal Policy Engine.

Extensions receive only the minimum Handshake Context required by their responsibility. They do not receive arbitrary Runtime state, raw transport ownership, Repository access, or permission to perform Upgrade themselves.

## 16. Migration Strategy

Migration should proceed through small, testable steps:

1. Add characterization tests for current HTTP method handling, successful Upgrade, post-Upgrade Authentication rejection, connection cleanup, and Listener continuation.
2. Introduce a pre-Upgrade orchestration seam in the HTTP request path without changing Authentication Service, Providers, Session, or Snapshot.
3. Normalize the existing AuthenticationRequest from HTTP metadata before `websocket.Accept` and verify copying, cancellation, and sensitive-data handling.
4. Invoke the existing Authentication Service before Upgrade and map expected rejection to safe HTTP rejection.
5. Carry a copied successful Principal across Upgrade into the existing authenticated Session handoff.
6. Make the Upgrade boundary close the connection when post-Upgrade handoff fails.
7. Remove the obsolete post-Upgrade Authentication path so each request is evaluated exactly once.
8. Add integration tests for rejection without `101`, Provider failure, cancellation, successful Upgrade, Session start, concurrent requests, and Listener shutdown.
9. Add safe operational reporting for Handshake stages without logging credentials.
10. Consider additional admission evaluations only in separate tasks after the Authentication path is stable.

Each step keeps the existing vertical buildable and avoids simultaneous redesign of Router, Session Manager, Runtime Host, or Provider contracts.

## 17. Risks

- HTTP response may be committed before all required Evaluation completes, making the Decision impossible to apply correctly.
- Migration may accidentally authenticate twice or leave both old and new paths active.
- Principal or request metadata may alias mutable input across the Upgrade boundary.
- Ownership ambiguity may cause connection leaks or double close during failed handoff.
- Slow or unavailable Providers may retain HTTP request resources until timeout or cancellation.
- Proxy and trusted-address assumptions may make Origin or IP evaluation incorrect.
- Error mapping may expose sensitive Provider details or hide operational failure as credential rejection.
- Listener shutdown may race with Evaluation or Upgrade unless cancellation and transfer are explicit.
- Future checks may be added without deterministic ordering, recreating hidden policy behavior.
- Disabled Authentication semantics may remain inconsistent with the proposed anonymous Principal model.

## 18. Open Questions

- What exact Principal is produced when Authentication is disabled?
- Which HTTP status, body, and challenge behavior are required for each rejection category?
- What timeout and cancellation policy applies to individual Providers and total Handshake Evaluation?
- How are trusted proxy data, remote address, TLS state, and forwarded headers normalized?
- What Configuration model and matching semantics define Origin Policy?
- In what order would maintenance, rate limiting, IP filtering, Origin, and Authentication execute?
- May an expected Provider “not applicable” result continue to another Provider, and how is it distinguished from rejection?
- Which Handshake events and metrics are required, and which fields are safe?
- At which startup stage are unavailable Providers and unsupported admission settings rejected?
- What HTTP connection reuse behavior is expected after each pre-Upgrade rejection?

These questions are intentionally deferred to implementation tasks or focused DPs. They do not block selection of the pre-Upgrade architecture boundary.

## 19. Relationship

### ARCH-001

The design applies [ARCH-001](../architecture/ARCH-001-runtime-architectural-pattern.md): HTTP metadata forms Context, configured admission checks perform Evaluation, allow/reject is Decision, and Listener executes HTTP rejection or Upgrade as the transport owner. Ownership and lifecycle remain explicit. The pattern does not become a generic framework.

### Runtime Alpha Review

The design directly addresses [High finding F-01](../reviews/runtime-alpha-review.md). It preserves the stable decisions identified by the review and does not claim to resolve F-02 Runtime Host composition or F-04 lifecycle hardening by documentation alone.

### MASTER_PLAN

The [Master Engineering Plan](../roadmap/MASTER_PLAN.md) identifies Handshake as the first Beta foundation gate before Router. This DP supplies the architectural direction for that epic without changing milestone criteria or becoming an implementation backlog.

### Existing Decisions and Proposals

- [ADR-0002](../adr/0002-configuration-dsl.md) requires Configuration First and explicit rejection of unsupported effective behavior.
- [ADR-0003](../adr/0003-runtime-architecture.md) requires component responsibility, dependency injection, and transport-neutral Authentication.
- [DP-001: Authentication](../proposals/DP-001-authentication.md) establishes pre-connection identity evaluation and HTTP rejection intent.
- [DP-004](../proposals/DP-004-authentication-runtime-contracts.md) defines the transport-neutral Authentication contract direction preserved here.

## 20. Decision

The selected direction is **Alternative D: Dedicated Handshake Pipeline**.

Required configured Authentication is evaluated before WebSocket Upgrade. Listener remains responsible for HTTP and WebSocket transport effects but does not implement Provider logic. Authentication remains transport-neutral. An allow Decision carries copied identity data across Upgrade, after which the existing authenticated handoff creates Session and transfers WebSocket ownership.

The Handshake Pipeline is a conceptual subsystem sequence based on ARCH-001. It is not a Universal Policy Engine, generic framework, new package declaration, or fixed set of Go interfaces. Future admission checks require their own Configuration and design. Implementation details remain a separate task.

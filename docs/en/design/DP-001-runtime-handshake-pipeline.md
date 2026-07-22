# DP-001: Runtime Handshake Pipeline

[Russian version](../../ru/design/DP-001-runtime-handshake-pipeline.md)

## 1. Status

**Status:** Draft

**Implementation status (TASK-ARCH-REVIEW-010):** Partially implemented. The
pre-Upgrade Authentication boundary, live admission checks, bounded Handshake,
Runtime-owned connection context, and transactional ownership handoff are in
production. The document remains Draft because Host shutdown does not yet use
the Session Manager wait set and the operational diagnostics/supervision
contracts remain incomplete.

This document proposes the Runtime Handshake architecture. It defines conceptual boundaries and invariants, not Go interfaces, packages, or a reusable policy framework.

## 2. Background

The Alpha Runtime upgrades an HTTP request before Authentication. The resulting WebSocket then passes through Connection Dispatcher, Authentication Dispatcher, and Session Dispatcher. Rejected credentials therefore receive `101 Switching Protocols` before the server closes the connection.

The [Runtime Alpha Architecture Review](../reviews/runtime-alpha-review.md) records this as High finding F-01. [DP-001: Authentication](../proposals/DP-001-authentication.md) already requires identity evaluation before a WebSocket Session is admitted.

The current path is:

```text
TCP accept
    -> HTTP request
    -> WebSocket Accept
    -> Authentication
    -> reject with WebSocket close
       or
       Session
```

This proposal changes the admission order without moving Provider logic into Listener or coupling Authentication to HTTP or WebSocket.

## 3. Problem Statement

Authentication after Upgrade crosses the transport ownership boundary before the admission decision is known. It allocates WebSocket resources for rejected clients, prevents an HTTP authentication rejection, complicates shutdown, and conflates policy rejection with post-Upgrade failures.

The design must also resolve the ownership and lifecycle gaps identified by the independent Architecture Stress Review: request context must not become Session context, shutdown must close the admission gate atomically, the WebSocket handoff point must be exact, and the real behavior of `websocket.Accept` must be reflected in the error model.

## 4. Existing Architecture

The current responsibilities worth preserving are:

- Listener owns its listening socket and HTTP server lifecycle.
- `net/http` owns accepted HTTP connections until a successful hijack or Upgrade.
- Authentication Service and Providers operate on transport-neutral input.
- Runtime Snapshot supplies effective Configuration without Repository access.
- Session owns an admitted WebSocket, its read loop, and serialized writes.

The current ordering and context transfer are not preserved. Authentication moves before Upgrade, and Session receives a Runtime-owned lifecycle context rather than `http.Request.Context()`.

## 5. Goals

- Complete configured Authentication before WebSocket Upgrade.
- Preserve transport-neutral Authentication Service and Providers.
- Define one owner for each transport resource at every stage.
- Define an admission state machine and its shutdown transitions.
- Separate request-scoped Handshake context from connection-scoped Session context.
- Distinguish rejection, cancellation, dependency, configuration, protocol, internal, and handoff failures.
- Keep request processing independent from Snapshot reads and Control Plane repositories.
- Keep the first implementation limited to pre-Upgrade Authentication.

## 6. Non Goals

This document does not design:

- Router or Delivery;
- Session Manager API;
- Middleware or a universal Policy Engine;
- Plugin ABI;
- an Origin Policy beyond current library behavior;
- rate limiting, maintenance, IP filtering, or general admission rules;
- another transport, QUIC, or HTTP/3;
- Provider-specific credential verification;
- Runtime Host composition details.

## 7. Constraints

The design follows [ARCH-001: Runtime Architectural Pattern](../architecture/ARCH-001-runtime-architectural-pattern.md).

### Configuration First

Effective behavior is built from a Published Configuration Snapshot before Listener starts. Unsupported active behavior is rejected rather than ignored.

### Ownership

Every mutable transport resource has exactly one current owner. Ownership transfer is explicit and has one conceptual acceptance point.

### Lifecycle

Handshake and Session have separate contexts and lifetimes. Network I/O and dependency calls do not run while a lifecycle lock is held.

### No Magic

Dependencies are composed explicitly. The request path does not discover Providers, read repositories or Snapshot, use global state, or use `context.Value` as a service locator.

### Boring Core

The first implementation performs only configured Authentication before Upgrade. The conceptual sequence is not a generic evaluator chain and does not promise future policies.

## 8. Design Alternatives

### Alternative A — Keep Authentication After Upgrade

Rejected clients continue to receive `101`, consume upgraded resources, and enter connection lifecycle before admission. This preserves the defect and is rejected.

### Alternative B — Put Provider Logic in Listener

Listener could authenticate before Upgrade, but it would gain Provider, Secret, and credential-verification knowledge. This violates transport neutrality and single responsibility and is rejected.

### Alternative C — Authenticate in Session

A provisional Session would exist without a valid Principal and could reject only through WebSocket closure. This violates the authenticated Session boundary and is rejected.

### Alternative D — Dedicated Handshake Boundary

A conceptual Handshake boundary normalizes request metadata, performs configured Authentication, produces a one-use admission Decision, and only then permits the transport executor to commit an Upgrade.

This alternative is selected. “Pipeline” describes the ordered lifecycle of one request. It does not define a generic pipeline API, an evaluator registry, or a universal policy model.

## 9. Proposed Architecture

```text
Published Snapshot
    -> Runtime Host bootstrap and readiness validation
    -> Host-owned Admission Gate and Runtime context holder
    -> read-only Runtime capabilities
    -> composition bridge
    -> composed Handshake, Listener, Authentication Service, timeout, and dependencies
    -> Listener Start

HTTP request
    -> initial admission check
    -> Handshake Context
    -> pre-Upgrade Authentication
    -> one-use Decision
       -> Reject or operational failure: HTTP handling, no WebSocket
       -> Allow: final admission validation immediately before commit
    -> websocket.Accept
       -> Accept rejection: library may commit HTTP response
       -> success: Upgrade boundary owns WebSocket
    -> Runtime-owned connection context
    -> Session registered in Runtime shutdown wait set
    -> explicit Session ownership acceptance
    -> Session lifecycle
```

Runtime composition bridges Host-owned lifecycle state to Handshake through live, read-only capabilities. Handshake does not depend on Runtime Host and cannot mutate the Admission Gate or cancel the Runtime context. Listener remains responsible for transport execution but contains no Provider-specific logic. Authentication remains transport-neutral. Session is admitted only with a valid Principal and never receives the original HTTP request or its context as lifecycle state.

## 10. Context Model

### Handshake Context

The Handshake executor owns one request-scoped context. Its lifetime ends at `Rejected`, `Aborted`, or successful `HandedOff`.

The effective deadline is the earliest of:

- cancellation of `http.Request.Context()`;
- Runtime or Listener shutdown cancellation;
- the configured Handshake timeout from the effective Runtime configuration.

Only required request metadata is copied into transport-neutral Authentication input. The HTTP request, Body, Headers, URL, ResponseWriter, and request context are not transferred to Session.

### Runtime-Owned Connection Context

After successful Accept and before handoff, the Upgrade boundary creates a separate connection/session context from the active Runtime context supplied by the read-only Runtime context capability. It is not derived as a child whose lifetime depends on `http.Request.Context()`.

Runtime lifecycle owns its cancellation path. It is canceled by connection or Session termination and by Runtime shutdown. A future Session Manager may participate in that lifecycle, but its API is outside this document.

Once handoff succeeds, the HTTP handler may finish without canceling the Session.

## 11. Decision Invariants

An admission Decision follows these invariants:

- Allow contains a valid Principal.
- When Authentication is enabled, the Principal represents successful Authentication.
- When Authentication is disabled, the Principal is the explicit anonymous Principal defined below.
- Reject is an expected negative outcome, not a Go error.
- Operational failure is distinct from Reject.
- A Decision loses authority if its context is canceled or the admission gate closes before commit begins.
- A Decision may be applied at most once.
- Only Allow may attempt admission commit.
- An execution error does not rewrite or reclassify the original Authentication result.

The exact Go representation remains an implementation detail.

## 12. Admission State Machine

The conceptual states are:

```text
Evaluating
    -> Allowed
    -> Rejected
    -> Aborted

Allowed
    -> Committing
    -> Aborted

Committing
    -> Upgraded
    -> Rejected     (Accept performs protocol or Origin rejection)
    -> Aborted      (cancellation or transport/internal failure)

Upgraded
    -> HandedOff
    -> Aborted
```

`Rejected`, `Aborted`, and `HandedOff` are terminal Handshake states. This state model is conceptual and is not a required Go enum.

### Shutdown Transitions

Runtime Host owns the Admission Gate. Handshake receives only a live, read-only admission capability through the composition bridge; it cannot open or close the Gate and does not depend on Runtime Host.

Handshake checks that capability before Authentication so a request observed while admission is closed does not evaluate credentials. It checks the same capability again immediately before `websocket.Accept`; this successful final admission validation is the linearization point for entering `Committing`.

At shutdown start, Host closes the Gate before waiting for active work:

- `Evaluating` is canceled and ends as `Aborted`; an Evaluation returning Allow afterward cannot commit.
- `Allowed` cannot enter `Committing` after the gate closes and becomes `Aborted`.
- Entry into `Committing` requires an atomic successful final admission validation immediately before `websocket.Accept`.
- Work already in `Committing` must either fail and clean up, or reach `Upgraded` and complete the controlled handoff.
- An upgraded connection must enter the Runtime shutdown wait set before handoff completes; otherwise the Upgrade boundary closes it before shutdown completes.
- `HandedOff` work is owned by Session lifecycle and participates in Runtime shutdown.

Shutdown therefore cannot finish while an admitted connection exists outside both the Upgrade boundary and Runtime shutdown tracking.

## 13. Ownership Model

### Listening Socket

Listener exclusively owns the listening socket from successful `net.Listen` until Listener Stop closes it.

### Accepted HTTP Connection

The `net/http` transport owns each accepted HTTP connection before successful hijack or Upgrade. Handshake Evaluation never owns or closes the raw TCP connection. HTTP rejection leaves reuse or closure to `net/http` semantics.

### HTTP Request and ResponseWriter

The HTTP server owns request lifecycle. Handshake borrows the request and ResponseWriter synchronously. Evaluation never writes a response.

Before Accept, the transport/Handshake executor may apply one HTTP rejection or failure response. Once any response is committed, the upper layer must not write another response or attempt Upgrade.

### Upgrade Boundary

After successful `websocket.Accept`, the Upgrade boundary exclusively owns the WebSocket. It remains the owner while the Runtime connection context is created and Session is prepared for acceptance.

### Session Ownership Acceptance

The conceptual handoff point occurs only after all of the following are true:

1. Session construction succeeded.
2. The Runtime-owned connection context exists.
3. Session was entered into the Runtime shutdown wait set.
4. Session explicitly accepted responsibility for the WebSocket and its closure.

Before this point, any failure is closed by the Upgrade boundary. At this point ownership transfers exactly once to Session. After it, Session is the sole closer. A downstream error after acceptance is a Session failure and never causes a second Close by the upper boundary.

[DP-004](DP-004-per-session-execution-boundary.md) is the normative refinement of this point. Its single synchronous atomic handoff publishes Session construction, Runtime-owned connection context, shutdown accounting, and transport ownership acceptance together. `accepted=false` means no transfer occurred and the Upgrade boundary cleans the WebSocket; `accepted=true` means transfer occurred and the Upgrade boundary never cleans it.

The general `AuthenticatedDispatcher` ownership contract permits an implementation to return an error with either ownership result: `accepted` alone determines cleanup ownership, and Handshake never reclaims transport after `accepted=true`. Generic Handshake tests for `accepted=true, error!=nil` therefore remain valid ownership-boundary tests for arbitrary implementations of that interface.

The target production Session Dispatcher is narrowed normatively by [DP-004](DP-004-per-session-execution-boundary.md): every pre-Commit failure returns `accepted=false` with its safe error, successful Commit returns only `accepted=true, nil`, and post-Commit failures belong to terminal accounting and, when known before immutable Terminal Result construction, its single Terminal Observer invocation. The target implementation does not produce `accepted=true, error!=nil`; this production restriction does not change the generic Handshake ownership contract.

### Authentication Result and Principal

Authentication result data is request-scoped and owns no transport. Principal is copied across the commit boundary so downstream mutation cannot change evaluated identity. Credentials and resolved Secrets are never included.

## 14. `websocket.Accept` Semantics

`websocket.Accept` is not a passive transport conversion. The library:

- validates WebSocket transport and protocol requirements;
- may apply its default Origin validation;
- may write and commit an HTTP response on rejection;
- creates the WebSocket only after a successful handshake.

The executor distinguishes three phases:

### Pre-Accept Rejection

Authentication Reject, cancellation, readiness invariant failure, or another pre-commit failure occurs before Accept. The executor may write one appropriate HTTP response. No WebSocket exists.

### Accept Rejection

Accept rejects protocol, method, header, or default Origin conditions and may already have written the HTTP response. The executor records the terminal category and performs no second HTTP write.

### Post-Accept Failure

Accept succeeded and a WebSocket exists, but connection-context creation, shutdown registration, Session construction, or handoff fails. HTTP semantics are no longer available. The current WebSocket owner closes the connection according to the ownership rules.

Origin remains a transport/library concern in the first implementation. Moving an explicit configurable Origin Policy before Accept requires a separate focused DP.

## 15. Disabled Authentication

Disabled Authentication produces an explicit anonymous Principal. Principal remains mandatory for every Allow Decision and every Session.

The anonymous Principal is clearly marked anonymous and unauthenticated and contains no fabricated credentials, roles, or claims. This preserves the Configuration intent that `authentication.enabled: false` admits clients while giving downstream components one stable identity shape. It also avoids optional identity checks throughout Session, Router, and future delivery code.

The current Session implementation may require internal adaptation to accept this explicit anonymous identity; no public API is defined here.

## 16. Timeout and Cancellation

Handshake has a bounded lifecycle. The configured `HandshakeSeconds` value supplies its maximum duration, while request cancellation or Runtime shutdown may end it earlier.

Runtime readiness validation rejects a missing, zero, unsupported, or out-of-range effective Handshake timeout before Listener Start. There is no unbounded fallback and request processing does not invent a default.

Cancellation must be propagated to Authentication Service, Providers, and their dependencies. A dependency that ignores context cannot be forcibly stopped by Go context cancellation. Its operation may outlive the caller; Runtime must not classify that limitation as successful admission, and shutdown boundedness cannot be guaranteed for such a dependency without a dependency-specific isolation mechanism outside this DP.

## 17. Error Model

The following minimal categories must remain distinguishable internally:

- **rejection** — expected negative Authentication result; not an operational error;
- **cancellation or deadline** — request, configured Handshake timeout, or Runtime shutdown ended Evaluation;
- **dependency unavailable** — Provider dependency could not serve the request;
- **invalid Runtime configuration** — a readiness invariant was violated;
- **protocol or Upgrade rejection** — Accept rejected the HTTP/WebSocket handshake;
- **internal failure** — an unexpected Handshake or transport executor defect;
- **handoff or Session failure** — failure after successful Accept while preparing or operating Session.

This document does not prescribe Go error types or exact client response bodies. Client-facing responses remain generic and never contain credentials or internal details.

Execution preserves the original Authentication outcome. For example, a transport failure after Allow does not become credential rejection, and an Accept Origin rejection does not become dependency failure.

## 18. Operational Error Ownership

The transport/Handshake executor is the single owner of the terminal outcome for one Handshake. It receives or observes the final Evaluation, response, Accept, and pre-handoff execution result and reports one safe terminal category.

After Session ownership acceptance, Session lifecycle owns later failures. Listener lifecycle owns terminal HTTP server `Serve` failures that are not normal shutdown.

A diagnostics sink, event schema, metrics, and logging integration require separate design. Until then, terminal Dispatcher, Handshake, and unexpected `Serve` errors must not be silently ignored. Components may wrap or propagate errors, but must not independently emit duplicate terminal reports for the same Handshake.

## 19. Configuration and Readiness

Before Listener Start, Runtime bootstrap validates and composes:

- Published Snapshot compatibility;
- configured Handshake timeout;
- Authentication Service;
- enabled Provider factories and Provider metadata;
- SecretResolver and other required dependencies;
- the Host-owned Admission Gate, read-only Runtime capabilities, composition bridge, and transport/Handshake executor.

Failure prevents Listener Start. Request-time configuration failure is not the normal model; if an invariant is nevertheless violated, it is an internal or invalid Runtime configuration failure and fails closed.

The request path uses only already composed dependencies and effective values. It does not read Runtime Snapshot, ConfigurationVersion, management API state, or Control Plane repositories.

The composition bridge follows [ADR-0004](../adr/0004-handshake-runtime-dependencies.md): Runtime composition supplies Handshake with live, read-only admission permission and Runtime context access. Handshake neither imports nor controls Runtime Host. This DP defines how those capabilities are consumed without designing Host APIs.

## 20. Session Manager Relationship

The future Session Manager does not participate in Authentication Evaluation. Its relevant architectural obligation is at handoff: the new Session must enter the Runtime shutdown wait set before Session ownership acceptance completes.

After acceptance, Session lifecycle, including shutdown cancellation and completion, is independent from the HTTP request. This document does not define Session Manager interfaces, storage, lookup, grouping, or routing behavior.

## 21. Scope of the First Implementation

The first implementation contains only:

- request metadata normalization required by existing Authentication;
- bounded pre-Upgrade Authentication;
- one-use allow, reject, or operational outcome;
- admission gate and shutdown coordination;
- controlled Accept and ownership handoff;
- explicit anonymous Principal when Authentication is disabled;
- safe terminal error propagation.

It does not add a general evaluator list. Origin remains controlled by the WebSocket library. Rate Limit, Maintenance, IP rules, and other admission checks are not guaranteed extension points of the current contract.

A shared policy framework is deferred until several real use cases establish common ordering, data, ownership, and failure semantics. Each future behavior remains Configuration First and requires focused design.

## 22. Migration Strategy

1. Preserve characterization coverage for current method handling, Upgrade, rejection, cleanup, and Listener continuation.
2. Add startup readiness checks for timeout, Authentication, Providers, and dependencies before Listener Start.
3. Inject the Host-owned admission and Runtime context capabilities through the composition bridge, then introduce the request-scoped Handshake context without changing Provider logic.
4. Execute existing Authentication before Accept and produce a one-use Decision.
5. Implement explicit anonymous Principal behavior for disabled Authentication.
6. Apply pre-Accept rejection exactly once and respect responses committed by Accept.
7. Create a Runtime-owned connection context after successful Accept from the Runtime context capability.
8. Register the prospective Session in Runtime shutdown tracking before explicit ownership acceptance.
9. Remove the obsolete post-Upgrade Authentication path so every request is authenticated at most once.
10. Preserve and propagate terminal Handshake and unexpected Serve errors without logging sensitive input.

Each step keeps the vertical buildable and does not introduce Router, Session Manager API, a concrete Host dependency in Handshake, or a generic policy framework.

## 23. Remaining Risks

- Dependencies that ignore context can outlive the bounded Handshake caller.
- Proxy and trusted-address assumptions remain undefined.
- Incorrect response-commit tracking can still cause a second response attempt.
- Incorrect implementation of the admission gate can admit work during shutdown.
- Session shutdown tracking remains dependent on the separate Runtime Host and Session Manager epics.
- Operational diagnostics remain incomplete until a focused diagnostics design is accepted.

## 24. Open Questions

- Which exact HTTP statuses, bodies, and challenge headers apply to each authentication rejection?
- How are trusted proxy data, remote address, TLS state, and forwarded headers normalized?
- What Configuration and matching semantics will a future explicit Origin Policy use?
- Does a Provider need a distinct “not applicable” result in addition to rejection?
- Which bounded Handshake events and metrics are required, and which fields are safe?
- What diagnostics sink receives terminal Runtime errors?
- What HTTP connection reuse behavior is expected after each pre-Accept rejection?

Anonymous identity, total Handshake timeout, startup readiness, shutdown admission, and ownership transfer are no longer open questions in this DP.

## 25. Findings Traceability

| Finding | Status | Resolution |
|---|---|---|
| F-01 — Request context and Session lifecycle conflict | Resolved | Session uses a separate Runtime-owned context; request context ends with Handshake. |
| F-02 — Shutdown is not atomic across admission | Resolved | Host owns the Gate; Handshake uses its read-only capability before Authentication and immediately before Accept. Commits after shutdown are forbidden, and in-flight commits must hand off into the wait set or clean up. |
| F-03 — TCP connection has two stated owners | Resolved | Listening socket, accepted HTTP connection, Upgrade boundary, and Session ownership are separated. |
| F-04 — Accept error classification contradicts library behavior | Resolved | Pre-Accept rejection, Accept rejection, and post-Accept failure are distinct; committed library responses are not rewritten. |
| F-05 — Handoff point is undefined | Resolved | Ownership acceptance occurs only after construction, Runtime context creation, and shutdown registration. |
| F-06 — Disabled Authentication cannot produce a Session | Resolved | Disabled Authentication produces an explicit anonymous Principal. |
| F-07 — Handshake timeout is undefined | Resolved | Effective Handshake timeout is mandatory, validated before start, and combined with request and shutdown cancellation. |
| F-08 — Decision invariants are incomplete | Resolved | Allow, Reject, operational failure, cancellation validity, and one-use semantics are explicit. |
| F-09 — Error categories are not guaranteed | Clarified | Minimal categories must remain distinguishable without prescribing Go error types. |
| F-10 — Terminal operational error has no owner | Clarified | Handshake executor owns the terminal Handshake outcome; Listener and Session own errors after their boundaries. |
| F-11 — Future extensibility is unproven | Deferred | First scope is Authentication only; Origin and other policies require focused DPs and are not promised extension points. |
| F-12 — Configuration validation owner is ambiguous | Resolved | Runtime bootstrap readiness precedes Listener Start; request path reads neither Snapshot nor Control Plane. |
| F-13 — Session Manager depends on an undefined handoff | Clarified | Session enters the Runtime shutdown wait set before ownership acceptance; no Session Manager API is designed. |
| F-14 — Pipeline may become a premature policy framework | Deferred | No evaluator framework is introduced; common policy machinery waits for multiple demonstrated use cases. |

## 26. Relationship

### ARCH-001

The design applies [ARCH-001](../architecture/ARCH-001-runtime-architectural-pattern.md): request metadata forms Context, Authentication performs Evaluation, admission produces Decision, and the transport/Handshake executor performs the effect. Ownership, cancellation, and lifecycle are explicit.

### Runtime Alpha Review

The design resolves the pre-Upgrade Authentication boundary identified by [Runtime Alpha Architecture Review](../reviews/runtime-alpha-review.md). Runtime Host composition and broader lifecycle hardening remain separate work.

### MASTER_PLAN

The [Master Engineering Plan](../roadmap/MASTER_PLAN.md) identifies Handshake as the first Beta foundation gate before Router. This DP defines that gate without becoming an implementation backlog.

### Existing Decisions and Proposals

- [ADR-0002](../adr/0002-configuration-dsl.md) requires Configuration First and explicit rejection of unsupported effective behavior.
- [ADR-0003](../adr/0003-runtime-architecture.md) requires component responsibility, dependency injection, and transport-neutral Authentication.
- [ADR-0004](../adr/0004-handshake-runtime-dependencies.md) defines the read-only capability bridge from Host-owned admission and Runtime context state to Handshake.
- [DP-001: Authentication](../proposals/DP-001-authentication.md) defines identity evaluation before admission and HTTP rejection.
- [DP-004](../proposals/DP-004-authentication-runtime-contracts.md) defines the direction for transport-neutral Authentication contracts.

## 27. Decision

The selected direction remains **Alternative D: Dedicated Handshake Boundary**.

Handshake receives only live, read-only Runtime capabilities and has no dependency on Runtime Host. It checks the Host-owned Admission Gate before Authentication and performs the final admission validation immediately before `websocket.Accept`. Configured Authentication executes before WebSocket Accept. A one-use Allow Decision with a valid authenticated or explicit anonymous Principal may enter admission commit only after that final validation succeeds. The library may reject and commit an HTTP response during Accept. After successful Accept, the Upgrade boundary owns the WebSocket until Session is registered for Runtime shutdown and explicitly accepts ownership. Session is created with a Runtime-owned context obtained through the Runtime context capability and independent from the HTTP request.

This remains a conceptual subsystem sequence, not a Universal Policy Engine, generic framework, package declaration, or fixed Go API. The first implementation supports Authentication only. Future admission behavior requires its own Configuration and focused design.

# Runtime Alpha Architecture Review

[Russian version](../../ru/reviews/runtime-alpha-review.md)

**Status:** Completed  
**Review date:** 2026-07-14  
**Assessment:** Ready with findings

## 1. Scope

This review assesses the implemented Runtime architecture before Router development. It covers the production code and tests in:

- `internal/runtimeconfig`;
- `internal/runtime`;
- `internal/listener`;
- `internal/connection`;
- `internal/authentication`;
- `internal/session`;
- `internal/message`;
- `internal/secretresolver`.

The review compares the implementation with [ADR-0002: Configuration DSL](../adr/0002-configuration-dsl.md), [ADR-0003: Runtime Architecture](../adr/0003-runtime-architecture.md), and [DP-001](../proposals/DP-001-authentication.md), [DP-002](../proposals/DP-002-secret-references.md), [DP-003](../proposals/DP-003-jwt-provider.md), and [DP-004](../proposals/DP-004-authentication-runtime-contracts.md). Control Service behavior, persistence, and proposed Router behavior are outside the review except where they define a Runtime boundary.

Evidence was taken from the repository state on the review date, the actual Go import graph, and the full test suite. Planned behavior is not treated as implemented behavior.

## 2. Current Runtime Vertical

The implemented configuration path is:

```text
Published ConfigurationVersion
    -> runtimeconfig.Builder
    -> immutable Runtime Snapshot
```

The implemented connection path, when components are assembled in tests or by a caller, is:

```text
TCP Listener
    -> HTTP Server
    -> GET /ws
    -> RFC 6455 Upgrade (101)
    -> ConnectionContext
    -> AuthenticationDispatcher
    -> Authentication Service
    -> ordered Provider(s)
    -> AuthenticatedContext
    -> Session Dispatcher
    -> Session read loop
    -> transport-neutral Message
    -> Message Handler (currently EchoHandler)
    -> Session.Send
    -> WebSocket frame
```

This is a working vertical in integration tests, but it is not assembled by a production composition root. `runtime.DefaultHost` currently stores a Snapshot and Container and changes lifecycle state only. It does not build or own Listener, Authentication, Dispatcher, or Session dependencies.

Authentication currently happens after the WebSocket handshake. Rejected credentials therefore close an already upgraded connection with `PolicyViolation`; they do not produce an HTTP `401` response.

## 3. Package Responsibilities

| Package | Owns | Must not own |
|---|---|---|
| `runtimeconfig` | Runtime-only Snapshot models, Published-version check, and deep-copy conversion from ConfigurationVersion | Repository access, HTTP, secret values, live Runtime services |
| `runtime` | Snapshot Container and Host lifecycle shell | HTTP details, Repository, concrete Provider behavior, connection processing |
| `listener` | TCP socket, HTTP server, `/ws` Upgrade, server contexts, handler and serve goroutine coordination | Authentication rules, Session behavior, message routing, secret resolution |
| `connection` | Transport handoff, temporary transport context, Authentication adapter, authenticated handoff | Provider-specific verification, Session internals, Repository, routing |
| `authentication` | Transport-neutral contracts, ordered Service, Bootstrap, Registry, Factory adapters, API Key and JWT Providers | Listener, WebSocket, HTTP request types, Repository, ConfigurationVersion |
| `session` | One authenticated WebSocket connection, Principal copy, lifecycle, read loop, serialized writes, Handler invocation | HTTP request metadata, Listener lifecycle, routing policy, persistence |
| `message` | Immutable text/binary Message, minimal Sender and Handler contracts, EchoHandler | WebSocket, HTTP, Session implementation, Authentication |
| `secretresolver` | Secret Reference validation, Resolver contract, caller-owned secret bytes, in-memory implementation | Provider logic, HTTP, ConfigurationVersion, logging |

The actual imports support these responsibilities. There are no Go import cycles. Runtime packages do not import a Control Plane Repository. `message` does not import `session` or WebSocket, `session` does not import `listener` or `net/http`, and `authentication` does not import Listener or WebSocket packages.

One deliberate adapter dependency remains: `runtimeconfig.Builder` imports the ConfigurationVersion domain model. The resulting Snapshot itself does not retain that model or a Repository reference.

## 4. Dependency Direction

The effective dependency direction is:

```text
runtimeconfig.Builder -> configurationversion

runtime -> runtimeconfig
listener -> runtimeconfig, connection
connection -> authentication, WebSocket transport
session -> connection, authentication, message, WebSocket transport
authentication -> runtimeconfig, secretresolver
message -> standard library only
secretresolver -> standard library only
```

The network path uses dependency injection at the main seams: Listener accepts a `connection.Dispatcher`; AuthenticationDispatcher accepts an Authentication `Service` and a downstream authenticated dispatcher; Session accepts a `message.Handler`; Bootstrap uses a Registry and Resolver; Registry stores Factory interfaces.

Dependency direction is broadly consistent with ADR-003. The missing production assembly means this direction is validated component-by-component and by integration tests, not yet by the executable Runtime lifecycle.

## 5. Resource Ownership

| Resource | Owner | Acquisition | Release and wait path |
|---|---|---|---|
| Published configuration data | `runtimeconfig.Snapshot` consumers | Builder deep-copies a Published ConfigurationVersion | Value lifetime; Container and Host return independent copies |
| TCP socket | `listener.DefaultListener` | `net.Listen` in `Start` | `http.Server.Shutdown`, listener `Close`, serve `WaitGroup` |
| HTTP server and base context | `listener.DefaultListener` | Created in `Start` | Base context cancel, `Shutdown`, tracked handler wait |
| Upgraded WebSocket before Authentication | `connection.AuthenticationDispatcher` | Ownership received from WebSocket handler | Closed on rejection/error or transferred downstream |
| Authenticated WebSocket | `session.DefaultSession` | Received through Session Dispatcher | Normal close plus `CloseNow`; read loop completion awaited |
| Session read loop | Calling HTTP handler goroutine | `Session.Run` blocks in the caller | Returns on peer close, context cancellation, Stop, handler error, or read error |
| Concurrent writes | Calling goroutines, serialized by Session | `Session.Send` | Each call returns after one write or error |
| Resolved Secret bytes | Authentication Provider call | Resolver returns a caller-owned copy | API Key and JWT Providers clear resolved byte slices after use |

The Listener owns its serve goroutine and tracks active HTTP handlers. Session Dispatcher does not detach Session work: it runs `Start`, blocking `Run`, and `Stop` in the request goroutine. Authentication, Registry, Providers, Message handlers, and SecretResolver create no background goroutines.

Ownership transfer is explicit at the major boundaries. The principal weakness is shutdown boundedness: some final waits are not governed by the caller's context, as recorded in F-04.

## 6. Lifecycle Review

### Runtime Host

`Created -> Running -> Stopped` is thread-safe and restart is rejected. Start and Stop are state transitions only and ignore their contexts. No Runtime resources are currently attached to this lifecycle.

### Listener

`Created -> Running -> Stopping -> Stopped` is protected by `sync.RWMutex`. Network and wait operations occur after the lifecycle mutex is released. Stop cancels the server base context, calls `http.Server.Shutdown`, closes the TCP Listener, and waits for server and handler completion. A concurrent Stop that observes `Stopping` returns immediately rather than waiting for the winning Stop to finish.

### Session

`Created -> Running -> Stopping -> Stopped` is protected by a lifecycle mutex. A single blocking read loop is enforced. Stop performs WebSocket close outside the lifecycle lock and concurrent Stop callers can wait on `stopDone`. However, the first Stop waits for the read loop without selecting on its context, and Send holds a lifecycle read lock during the network write.

### Cancellation and goroutines

Listener handlers receive a cancellable server context. Authentication and Provider calls propagate request cancellation. Session reads and writes receive contexts. There are no intentionally detached application goroutines. Normal implemented paths terminate cleanly, but custom downstream components that do not honor cancellation can make Listener or Session shutdown exceed the requested deadline.

## 7. Security Boundaries

- Configuration and Snapshot models contain Secret References, not secret values.
- API Key and JWT Providers resolve secrets per Authentication attempt. Constructors and Factories do not resolve them.
- MemoryResolver copies values on input and output. Providers clear returned secret byte slices after use.
- API Key comparison uses `crypto/subtle.ConstantTimeCompare`.
- Runtime packages contain no credential logging. Authentication dispatcher errors expose a generic message while retaining the wrapped cause for `errors.Is`.
- `AuthenticationRequest` temporarily contains copied handshake Headers and Query values. Session stores neither that request nor the original `http.Request`; it retains only a copied Principal, RemoteAddress, ID, and creation time.
- The current boundary authenticates after `101 Switching Protocols`. This conflicts with DP-001's pre-connection Authentication and HTTP `401` model and increases unauthenticated resource allocation.
- WebSocket origin handling uses the library default. There is no explicit per-Configuration origin policy yet.
- JWT Runtime verification currently supports HMAC algorithms HS256, HS384, and HS512 only. Asymmetric algorithms remain DSL metadata but are rejected by the Runtime Provider constructor.
- Listener TLS metadata and timeout metadata are present in Snapshot but are not enforced by the current Listener implementation.

## 8. Extensibility Review

| Extension | Existing extension point | Assessment |
|---|---|---|
| Router | `message.Handler` injected into Session Dispatcher | A Router can implement Handler without changing Session or Listener. Routing contracts and failure semantics still need design. |
| Middleware | Handler wrappers can compose around another Handler | Technically possible without core changes, but ordering, configuration, and lifecycle contracts are not defined. |
| Basic Provider | Factory and Provider Registry by provider type | A concrete Basic Provider and Factory can be added without Registry or Service changes. |
| Session Manager | No registration/removal contract | Adding management, limits, or coordinated shutdown will require a new seam around Session creation and completion. |
| Message persistence | Handler or Handler decorator | Storage can be invoked behind Handler, but delivery guarantees, retries, backpressure, and error policy are undefined. |
| Plugins | Factory registration and Handler injection | Compile-time extension is possible. Discovery, compatibility, isolation, and dynamic plugin loading do not exist. |

The architecture has useful narrow seams and avoids type switches in Registry, Service, and Bootstrap. Router is the best-developed next seam. Session Manager and dynamic plugins are not yet extensible without adding contracts, which is acceptable at alpha.

## 9. Findings

### F-01 — Authentication occurs after WebSocket Upgrade

**Severity:** High  
**Observation:** `websocket.Accept` sends `101 Switching Protocols` before AuthenticationDispatcher runs. Rejection is expressed as a WebSocket close frame rather than HTTP `401`.  
**Risk:** Unauthenticated clients consume upgraded-connection resources, HTTP clients cannot observe the documented Authentication failure contract, and later correction would reshape the Listener-to-Dispatcher boundary.  
**Recommendation:** Decide and implement a pre-Upgrade Authentication boundary, or formally replace DP-001 with an accepted decision and an explicit post-Upgrade threat model.  
**Timing:** Must be resolved before Router.

### F-02 — Runtime Host is not the production composition root

**Severity:** High  
**Observation:** The working Listener -> Authentication -> Session -> Handler vertical is assembled only by callers and integration tests. Host stores Snapshot and Container but starts no component.  
**Risk:** There is no production lifecycle that proves dependency ordering, startup rollback, shutdown ordering, or the single ownership chain required by ADR-003.  
**Recommendation:** Make Host assemble and own Resolver, Provider Registry/Factories, Authentication Bootstrap, dispatchers, Handler, and Listener through existing interfaces; add rollback for partial startup.  
**Timing:** Must be resolved before Router.

### F-03 — Listener configuration is only partially effective

**Severity:** Medium  
**Observation:** Listener validates and stores TLS metadata but opens plain TCP. Snapshot timeout values are not applied to `http.Server` or WebSocket operations.  
**Risk:** A valid Published Snapshot can imply transport protections and limits that Runtime does not enforce. Operators may assume configuration is active when it is not.  
**Recommendation:** Either implement the configured TLS and timeout behavior or reject unsupported enabled settings at Runtime build/start with an explicit error.  
**Timing:** May follow Router, but must be completed before beta.

### F-04 — Shutdown deadlines are not uniformly authoritative

**Severity:** Medium  
**Observation:** Listener waits for tracked handlers after `Shutdown(ctx)` without a deadline-aware wait; a concurrent Listener Stop returns during `Stopping`; the first Session Stop waits for its read loop without observing `ctx`; Session Send holds the lifecycle read lock across WebSocket I/O.  
**Risk:** A faulty or slow Handler/write can delay shutdown beyond the caller deadline, and concurrent lifecycle callers do not all observe completion consistently.  
**Recommendation:** Define one lifecycle contract, make all completion waits context-aware, provide a shared completion signal for Listener Stop, and avoid holding lifecycle locks across network I/O.  
**Timing:** Must be resolved before Router because Router introduces arbitrary Handler work.

### F-05 — Runtime failures lack an operational reporting path

**Severity:** Medium  
**Observation:** Listener discards `http.Server.Serve` errors and WebSocket handler discards Dispatcher errors. Runtime components intentionally do not log sensitive input, but no safe event, metric, or injected logger reports failures.  
**Risk:** Provider outages, handler failures, and unexpected server termination can be invisible in production, slowing detection and diagnosis.  
**Recommendation:** Add a transport-neutral, dependency-injected observability/error reporting seam with a documented redaction policy; preserve generic client-facing errors.  
**Timing:** May follow Router, but must be completed before beta.

### F-06 — The Snapshot adapter shares the Runtime model package

**Severity:** Low  
**Observation:** `runtimeconfig.Builder` imports `configurationversion`, while the Snapshot types themselves are otherwise Control Plane independent.  
**Risk:** Future convenience changes could pull additional Control Plane concepts into a package imported throughout Runtime.  
**Recommendation:** Keep this as the sole explicit adapter or move conversion to a dedicated composition/adapter package when a production Loader is introduced. Never add Repository access to `runtimeconfig`.  
**Timing:** Backlog; revisit with the Loader and before v1.0.

### F-07 — Provider coverage is intentionally incomplete

**Severity:** Accepted limitation  
**Observation:** API Key and HMAC JWT are implemented; Basic, asymmetric JWT, JWKS, revocation, and distributed Secret backends are absent.  
**Risk:** Supported deployment scenarios remain narrow, but current code rejects unsupported JWT algorithms rather than silently accepting them.  
**Recommendation:** Keep documentation explicit and add Provider implementations incrementally through Factory registration without widening Service or Registry.  
**Timing:** Product-driven; Basic and required JWT algorithms should be selected before beta.

### F-08 — Origin policy relies on library defaults

**Severity:** Accepted limitation  
**Observation:** WebSocket Accept uses the library's standard Origin behavior with no Configuration policy.  
**Risk:** Deployments cannot express stricter or proxy-aware origin rules through the DSL.  
**Recommendation:** Design an explicit Configuration-first origin policy before exposing Runtime beyond controlled environments.  
**Timing:** Before beta.

### F-09 — Higher-level Runtime services are absent

**Severity:** Accepted limitation  
**Observation:** Router, Middleware contracts, Session Manager, persistence, backpressure, broadcast, monitoring, and dynamic plugins are not implemented.  
**Risk:** The current vertical is suitable for architecture validation and Echo behavior, not production workloads.  
**Recommendation:** Add only the next required contracts, beginning with Router after the must-fix findings, and keep transport-neutral Message and Handler boundaries.  
**Timing:** Milestone backlog; not a defect in the alpha foundation.

## 10. Stable Decisions

The following decisions are sufficiently demonstrated by code and tests to treat as stable for the next milestone:

- Published ConfigurationVersion is the only input accepted by Snapshot Builder.
- Snapshot consumers receive deep copies; mutable Provider and JWT collections do not alias ConfigurationVersion or caller values.
- Runtime is separated from Control Service HTTP API and Repository implementations.
- Dependencies are supplied explicitly through constructors and narrow interfaces.
- Authentication Service is ordered and Provider-neutral; Factory and Registry isolate concrete Provider construction.
- Secret values are resolved close to use and are not stored in Configuration, Snapshot, Factory, or Provider configuration.
- Runtime Message and Handler contracts are transport-neutral.
- Session owns the authenticated WebSocket after handoff, uses one read loop, and serializes writes.
- Router can be introduced as a `message.Handler` without exposing WebSocket transport.

## 11. Decisions Not Yet Stable

- Whether Authentication must occur before Upgrade, as DP-001 specifies, or after Upgrade.
- The production Host composition, startup rollback, and shutdown ordering.
- Uniform lifecycle cancellation and concurrent Stop semantics.
- Effective TLS, timeout, and origin policy behavior.
- Runtime observability, audit, metrics, and redaction contracts.
- Authentication failure taxonomy beyond success, rejection, and returned error.
- Basic and asymmetric JWT Provider scope.
- Router, Middleware, Session Manager, persistence, backpressure, and plugin contracts.
- Loader ownership and the long-term location of the ConfigurationVersion-to-Snapshot adapter.

## 12. Milestone Alpha Assessment

**Verdict: Ready with findings.**

The implementation validates the central ADR-003 direction: immutable Snapshot data, explicit dependency injection, transport-neutral Authentication and Message contracts, Registry/Factory extensibility, clear connection handoff, and testable components without import cycles or Repository coupling.

The project is not production-ready. Router development should begin only after F-01, F-02, and F-04 are resolved. Those findings affect the security boundary, the actual composition root, and lifecycle behavior that every Router implementation would otherwise inherit. Medium findings F-03 and F-05 must be closed before beta. Accepted limitations must remain explicit and must not be represented as implemented capabilities.

## 13. Recommended Next Tasks

### Must be resolved before Router

1. Move Authentication to a pre-Upgrade boundary, or record an accepted replacement decision with an explicit security model and client failure contract.
2. Turn Runtime Host into the production composition root for the existing Snapshot, Authentication, Dispatcher, Session Handler, and Listener components, including startup rollback and ordered shutdown.
3. Harden lifecycle contracts: context-aware waits, shared Stop completion, no lifecycle locks across network I/O, and tests with blocking/failing downstream handlers.

### Can be completed after Router, but before beta

4. Apply Listener TLS and timeout Snapshot settings, or fail startup for unsupported active configuration.
5. Add injected, secret-safe Runtime error reporting, metrics, and lifecycle events.
6. Add a Configuration-first WebSocket Origin policy and integration tests for allowed, denied, absent, and proxy-mediated Origin values.

### Backlog

7. Introduce Session Manager, persistence, and plugin contracts only when concrete use cases define ownership, delivery guarantees, and compatibility requirements.

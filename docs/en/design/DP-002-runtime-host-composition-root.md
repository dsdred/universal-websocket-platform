# DP-002: Runtime Host Composition Root

[Russian version](../../ru/design/DP-002-runtime-host-composition-root.md)

## 1. Status

**Status:** Draft

This document proposes the production composition boundary for one Runtime instance. It defines component assembly, readiness, lifecycle coordination, startup rollback, and shutdown ordering. It does not define Go APIs or subsystem business behavior.

## 2. Background

The Alpha Runtime contains the components of a working vertical, but no production path assembles and owns them as one system.

Today:

- `runtime.DefaultHost` stores copied Snapshot and Container values and performs only `Created -> Running -> Stopped` state changes;
- `runtime.Container` exposes only a copied Snapshot;
- Authentication Bootstrap builds a Service from an Authentication Snapshot, Registry, Factories, and Secret Resolver;
- Listener Bootstrap builds a Listener from Listener Snapshot and an injected Dispatcher;
- Session Dispatcher, Authentication Dispatcher, Message Handler, Registry, Resolver, and Listener are manually wired by callers and integration tests;
- Host does not start Listener, roll back partial construction, coordinate Session shutdown, own a Runtime context, or expose meaningful readiness.

These boundaries are useful and remain narrow, but subsystem Bootstrap is no longer sufficient for Beta. Each Bootstrap knows how to construct one subsystem; none owns the complete dependency graph, startup transaction, reverse-order cleanup, or Runtime-wide failure state. Expanding Container into a bag of services would hide rather than solve that problem.

The [Runtime Alpha Architecture Review](../reviews/runtime-alpha-review.md) records the missing production composition root as High finding F-02. [DP-001: Runtime Handshake Pipeline](DP-001-runtime-handshake-pipeline.md) also requires startup readiness, a Runtime-owned context, an admission gate, and Session participation in the Runtime shutdown wait set.

## 3. Goals

- Make Runtime Host the only production composition root for one Runtime instance.
- Assemble the supported Runtime graph from one immutable Published Snapshot and explicit operational dependencies.
- Validate all supported effective behavior before serving traffic.
- Define deterministic construction, startup, rollback, shutdown, and failure ordering.
- Give Runtime one owned root context and one readiness boundary.
- Preserve narrow subsystem contracts and dependency direction.
- Keep component creation explicit, ordinary, and reviewable.
- Allow tests to replace narrow dependencies without creating a production service locator.

## 4. Non Goals

Runtime Host does not:

- route Runtime Messages;
- authenticate clients or implement Provider logic;
- own WebSocket transport after Session handoff;
- store, search, group, or address Sessions;
- implement Delivery or Persistence;
- write application data to a database;
- resolve request credentials itself;
- define Plugin ABI or discover Plugins dynamically;
- interpret business rules or make subsystem Decisions;
- reload or restart a Runtime instance;
- design Control Plane loading or process supervision.

Host makes composition and lifecycle decisions only. It delegates all subsystem behavior to the component that owns that responsibility.

## 5. Responsibilities

Runtime Host is responsible for:

- accepting an immutable Runtime Snapshot and explicit operational dependencies;
- validating Snapshot compatibility and effective Runtime support;
- creating concrete infrastructure dependencies through explicit constructors;
- registering the explicitly supported Authentication Factories;
- invoking subsystem Bootstrap in dependency order;
- wiring Message Handler, Session handoff, Handshake, Authentication, and Listener boundaries;
- creating and canceling the root Runtime context;
- owning the Admission Gate and stable Runtime context holder while exposing only read-only capabilities to Handshake;
- starting externally visible components only after internal construction succeeds;
- rolling back partially initialized or started components;
- marking Runtime Ready only after every mandatory component is operational;
- closing admission before shutdown cancellation;
- stopping and cleaning up owned components in reverse dependency order;
- retaining primary and cleanup failures for diagnostics.

Host is not a general-purpose dependency container. Its fields and construction steps correspond only to supported production components. Adding a component requires an explicit composition change backed by that subsystem's design.

## 6. Dependency Graph

### Construction Graph

```text
Host
    -> read-only Runtime capabilities
    -> Handshake
    -> Listener

Published ConfigurationVersion
    -> runtimeconfig.Builder
    -> immutable Snapshot
    -> Runtime Host
         |
         +-> lifecycle state, Admission Gate, and inactive Runtime context holder
         |      +-> read-only Runtime capabilities
         |
         +-> concrete Secret Resolver
         |
         +-> Authentication Registry
         |      +-> explicit API Key Factory
         |      +-> explicit JWT Factory
         |
         +-> Authentication Bootstrap
         |      +-> Authentication Service
         |             +-> ordered Providers
         |                    +-> Secret Resolver
         |
         +-> Message Handler
         +-> Session handoff/Dispatcher
         |      +-> Execution Owner
         |             +-> read-only root Runtime context observation input
         +-> Handshake executor
         |      +-> read-only Runtime capabilities
         |      +-> Authentication Service
         |      +-> Session handoff
         |
         +-> Listener Bootstrap
                +-> Listener
                       +-> Handshake executor
```

The graph describes explicit production composition, not a package mandate or a generic construction API. The composition bridge supplies Handshake with live, read-only admission permission and Runtime context access while preserving Host ownership. Through that read-only capability boundary, composition also supplies each Execution Owner with the read-only root Runtime context observation input required by DP-004; the root `CancelFunc` is never exposed, and ownership of the root context remains with Runtime Host. Existing Authentication Factory and Registry contracts remain subsystem-specific; Host does not introduce a universal Factory or Registry.

The Runtime-derived connection/execution context remains a separate Session-owned context. The Runtime-cancellation callback observes only the Host-owned root Runtime context. Session Cleanup cancels the derived context and therefore cannot itself produce `RuntimeCanceled`.

### Runtime Call Direction

```text
Listener
    -> Handshake executor
         -> read-only Runtime capabilities
         -> Authentication Service
              -> Provider
                   -> Secret Resolver
         -> Session handoff
              -> Session
                   -> Message Handler
```

Reverse dependencies are prohibited:

- Listener does not know Host, concrete Providers, Session internals, Router, or Persistence.
- Handshake does not know Host, lifecycle mutation, Admission Gate implementation, or Runtime context cancellation.
- Authentication does not know Host, Listener, HTTP, WebSocket, Session, or repositories.
- Session does not know Host, Listener, Snapshot, Control Plane, or repositories.
- Secret Resolver does not know Host, Authentication behavior, or ConfigurationVersion.
- Message Handler does not know Host or WebSocket transport.
- no Runtime component locates another component through Container, globals, reflection, or `context.Value`.

`runtime.Container` remains a Snapshot holder unless a later focused decision changes it. This proposal does not turn it into a service locator.

## 7. Startup Pipeline

Startup is one ordered transaction:

```text
Immutable Snapshot
    -> compatibility and effective-setting validation
    -> Secret Resolver construction and readiness
    -> required startup Secret resolution
    -> Authentication Registry and Factory registration
    -> Authentication Service construction
    -> Message Handler and Session handoff construction
    -> Handshake executor construction with stable read-only Runtime capabilities
    -> Listener construction without opening the socket
    -> Listener Start
    -> startup commit: publish active Runtime context and open admission gate
    -> Runtime Ready
```

The details are:

1. Host copies and validates Snapshot identity, Listener metadata, timeouts, TLS support, Authentication metadata, and every active setting it claims to execute.
2. The stable Runtime context holder and read-only capabilities already exist with Host but remain inactive. The caller's startup context limits startup and does not become the lifetime of a successfully running Runtime.
3. Host constructs the selected Secret Resolver explicitly and verifies its startup readiness.
4. Secret material required to start a component, such as future TLS key material, is resolved before Listener Start and transferred only to that owner. Authentication Providers that intentionally resolve credentials per request retain that contract; Host validates their references and construction but does not eagerly store their secret values.
5. Host creates Authentication Registry and registers only production Factories explicitly supported by this build.
6. Authentication Bootstrap constructs the ordered Service and Providers from Authentication Snapshot and Resolver.
7. Host creates the selected Message Handler, Session handoff, and the DP-001 Handshake executor with the stable read-only Runtime capabilities. Admission Gate ownership remains with Host.
8. Listener Bootstrap validates Listener Snapshot and constructs Listener without network side effects.
9. Listener starts last. No network traffic is accepted before all mandatory downstream components exist.
10. As one successful startup commit, Host creates and publishes the active Runtime context, opens the admission gate, and enters `Running`. Successful Start means Runtime is Ready.

Construction is deterministic. Map iteration, reflection, package discovery, dynamic registration order, and generic component factories do not determine behavior.

## 8. Shutdown Pipeline

Shutdown reverses externally visible and dependency ownership order:

```text
Runtime Ready = false
    -> close admission gate
    -> Session Manager BeginShutdown and immutable Stop snapshot
    -> request Stop for snapshot Sessions
    -> cancel root Runtime context
    -> initiate Listener Stop
       |-> finish or abort active Handshakes and drain Listener handlers
       |-> owners terminalize and release eligible lifetime leases
    -> after Listener Stop returns, Session Manager Wait
    -> stop Session handoff and message-processing owners
    -> release Authentication components
    -> release startup-resolved Secrets and Resolver resources
    -> Runtime Stopped
```

For active Sessions, [DP-004](DP-004-per-session-execution-boundary.md) is the normative refinement of the interval between closing admission and stopping remaining components. This ordering follows the implemented ARCH-002 rule that root Runtime context cancellation precedes Listener Stop.

Listener handler drain and owner drain proceed in parallel after Listener Stop is initiated. Neither branch waits for the other by contract. Session Manager Wait starts after Listener Stop returns and succeeds only after truthful Registration and owner-lease accounting converge.

Guarantees:

- readiness becomes false before shutdown work starts;
- no new admission commit begins after the gate closes;
- already committing work follows DP-001: it enters Runtime shutdown tracking through successful handoff or closes before shutdown completes;
- every owned component is stopped or released at most once;
- cleanup order is the reverse of successful acquisition/start order;
- Stop is idempotent, and concurrent callers observe the same completion and terminal result;
- shutdown errors are accumulated without hiding the first causal failure;
- a caller deadline bounds that caller's wait, not Host's ownership obligation;
- if cleanup must continue after a caller deadline, it remains Host-owned in `Stopping` and has an explicit completion path;
- a dependency that ignores cancellation cannot be forcibly stopped; Runtime reports the timeout and remains responsible for eventual cleanup rather than declaring a false `Stopped` state.

No component is allowed to detach cleanup work without a Runtime owner.

## 9. Ownership

| Resource or component | Creator | Lifecycle owner | Cleanup responsibility |
|---|---|---|---|
| Published ConfigurationVersion | Control Plane/Loader boundary | Outside Runtime | Not retained by Runtime services |
| Immutable Snapshot copy | Builder, then copied by Host | Host | Value lifetime; never mutated |
| Host state, Admission Gate, Runtime context holder, and root Runtime context | Host | Host | Close Gate and cancel context on shutdown, rollback, or terminal failure |
| Secret Resolver | Host through an explicit concrete constructor or explicit owned dependency | Host | Close/release if its contract has lifecycle |
| Startup-resolved secret material | Resolver, transferred to the consuming component | Consuming component | Clear/release according to component and Resolver contracts |
| Authentication Registry | Host | Host | Value cleanup; no dynamic production mutation after initialization |
| Authentication Providers and Service | Authentication Bootstrap under Host composition | Host for component lifecycle; Provider owns per-call secret copies | Release component resources; clear per-call secrets close to use |
| Message Handler | Host through an explicit subsystem constructor | Host or its future subsystem owner | Stop only if its focused contract defines lifecycle |
| Read-only Runtime capabilities | Host/composition bridge | Host | Expose live admission permission and active Runtime context without mutation or cancellation |
| Handshake executor | Host | Host | Terminal Handshake tracking; consumes only read-only Runtime capabilities |
| Listener | Listener Bootstrap under Host composition | Host coordinates; Listener owns socket/server | Host calls Listener Stop; Listener closes and waits for its resources |
| Upgraded WebSocket before handoff | Upgrade boundary defined by DP-001 | Upgrade boundary | Close on pre-handoff failure |
| Session after ownership acceptance | Session | Session; Runtime tracks shutdown completion | Session closes WebSocket; Host never double-closes it |

Borrowed operational dependencies must be explicitly identified at composition time. Production defaults prefer Host-owned lifecycle dependencies. Host never stores raw credential values merely to make them globally available.

## 10. Lifecycle

The conceptual Host lifecycle is:

```text
Created
    -> Initialized
    -> Starting
    -> Running
    -> Stopping
    -> Stopped

Created / Initialized / Starting / Running / Stopping
    -> Failed
```

- **Created:** Host owns copied input but no complete component graph.
- **Initialized:** Snapshot and readiness checks passed; all mandatory components are constructed without externally visible traffic.
- **Starting:** lifecycle-bearing components are starting in order; admission remains closed.
- **Running:** all mandatory components are operational, admission is open, and Runtime is Ready.
- **Stopping:** readiness and admission are closed; cancellation, waits, and reverse cleanup are in progress.
- **Stopped:** all owned resources completed cleanup. The state is terminal.
- **Failed:** initialization, startup, runtime supervision, or cleanup failed. The causal and cleanup errors remain observable. Cleanup still runs before resource ownership is considered discharged.

Restart and in-place reload are not supported. A new Published Snapshot creates a new Runtime Host instance.

Stop before Running performs reverse cleanup of everything acquired so far. A normal explicit Stop reaches `Stopped`; startup failure, startup cancellation, or unexpected component termination reaches `Failed` after rollback. A cleanup failure also leaves the terminal state `Failed`.

## 11. Failure Model

### Invalid Snapshot

Initialization fails before any network resource starts. Host records invalid Runtime configuration, releases any earlier owned values, and enters `Failed`.

### Secret Resolution Failure

Failure of a Secret required at startup prevents Listener Start. Partially resolved material is released and initialized components roll back in reverse order. A per-request Provider resolution failure remains an Authentication operational error and does not retroactively change Runtime readiness unless a later focused health policy says otherwise.

### Authentication Construction Failure

Missing Factory, unsupported enabled Provider, invalid Provider runtime configuration, or unavailable mandatory Authentication dependency prevents Listener Start. Registry, Resolver, and other earlier components are cleaned up.

### Listener Start Failure

No Runtime is Ready. Host stops any partially started Listener state, then releases Handshake, Session handoff, Authentication, and Resolver resources in reverse order.

### Startup Cancellation

Startup observes both caller cancellation and the Host root context. Cancellation prevents admission, rolls back acquired resources, and ends in `Failed` with a cancellation/deadline category.

### Unexpected Failure While Running

Unexpected termination of a mandatory lifecycle component immediately clears readiness, closes admission, triggers coordinated shutdown, and ends in `Failed`. Runtime must not remain Ready with a missing mandatory component.

### Rollback Failure

Rollback continues after an individual cleanup error. The primary startup or runtime failure and all cleanup failures remain distinguishable. Host never reports successful Start or clean Stop when owned resources are unresolved.

## 12. Runtime Readiness

Runtime is Ready only when all of these conditions hold:

- Host state is `Running`;
- Snapshot and all active settings passed support validation;
- required startup Secrets were resolved and transferred to their owners;
- Authentication Service and all enabled Providers were constructed;
- Handshake executor, Host-owned admission gate, and read-only Runtime capabilities are operational;
- Listener successfully started and is accepting traffic;
- root Runtime context is active;
- every required lifecycle component is supervised by Host.

Before Ready:

- admission gate remains closed;
- Listener must not accept externally usable WebSocket Sessions;
- no caller may treat successful construction or `Initialized` as successful Runtime startup;
- Runtime must not advertise itself as available.

Readiness becomes false atomically when shutdown or terminal failure begins. Readiness is a Host lifecycle fact; this document does not design an HTTP health endpoint, service-discovery protocol, or monitoring API.

## 13. Runtime Context

Host creates the stable Runtime context holder together with Host. The holder remains inactive through construction and startup and publishes an active root Runtime context only as part of successful startup commit. Host owns that context, which is separate from the context passed by the caller to Start:

- startup context controls how long the caller waits for startup;
- root Runtime context controls the lifetime of the running instance;
- Host cancellation ends Runtime-owned work during rollback, Stop, or terminal failure.

Components receive the root context or narrowly derived child contexts through explicit construction or lifecycle calls. Handshake receives only the live, read-only Runtime context capability from the composition bridge; it cannot publish, replace, or cancel the root context. Context values do not carry services or configuration.

DP-001 Handshake context observes request cancellation, configured Handshake timeout, and Runtime shutdown through the read-only capabilities. After successful admission, Session receives a separate Runtime-owned connection context derived through the Runtime context capability, not from `http.Request.Context()`. Host shutdown cancellation reaches that connection context, while Session remains the owner of the WebSocket after handoff.

Closing the admission gate precedes root cancellation so an Evaluation completing during shutdown cannot begin a new Upgrade commit.

## 14. Extension Points

Host may later compose additional focused components, but it does not define their APIs:

- **Metrics:** a diagnostics/metrics component may observe bounded lifecycle and subsystem events without becoming a dependency lookup mechanism.
- **Persistence:** a focused Storage component may be constructed before the Handler or Delivery component that requires it and stopped after its dependents.
- **Router:** a designed Router may replace the current explicit Message Handler selection without changing Listener or Authentication.
- **Plugins:** supported Plugin implementations may be supplied through subsystem-specific contracts after compatibility and isolation design; Host does not scan or expose a universal Plugin registry.
- **Delivery:** a focused Delivery component may be wired between Router decisions and Session capabilities after its semantics are designed.

Each extension must define dependencies, ownership, readiness, startup failure, and shutdown order in its own DP. Host changes only to add explicit production wiring. No extension receives arbitrary Host or Container access.

## 15. Migration

Migration proceeds in small, verifiable steps:

1. Preserve characterization tests for current Host, Bootstrap, Listener, Authentication, Session, and Echo vertical.
2. Introduce the richer Host lifecycle and readiness semantics without starting network components.
3. Move Snapshot support validation and effective-setting rejection into the pre-start composition phase.
4. Give Host an owned Runtime context, stable inactive context holder, Host-owned Admission Gate, read-only capability bridge, and deterministic rollback ledger for explicitly acquired components.
5. Create Resolver, Registry, and production Authentication Factories through explicit wiring; build Authentication Service through existing Bootstrap.
6. Compose the Message Handler and Session handoff through existing narrow contracts.
7. Compose the DP-001 Handshake executor with the Host-owned Gate and Runtime context exposed only through stable read-only capabilities when that boundary is implemented.
8. Build Listener last and make it the only externally visible component started by Host.
9. Add failure-injection tests for every initialization/start boundary and reverse cleanup step.
10. Replace manual production assembly with Host while retaining direct component construction in unit and focused integration tests.
11. Add supervision tests for unexpected Listener termination, concurrent Stop, deadlines, and partial cleanup failure.

No migration step turns Container into a service locator, introduces reflection, or requires a general component framework.

## 16. Relationship

### ARCH-001

This proposal applies [ARCH-001](../architecture/ARCH-001-runtime-architectural-pattern.md): dependencies are explicit, every resource has one owner, lifecycle and cancellation are visible, and Host delegates Evaluation and Execution to focused components. The composition root is intentionally boring rather than a god object.

### MASTER_PLAN

The [Master Engineering Plan](../roadmap/MASTER_PLAN.md) identifies production Runtime Host, startup rollback, shutdown ordering, lifecycle hardening, and Configuration validation as Beta foundation gates before Router. This proposal defines their composition boundary without becoming a task backlog.

### DP-001

[DP-001: Runtime Handshake Pipeline](DP-001-runtime-handshake-pipeline.md) defines pre-Upgrade Authentication, two admission checks, Runtime-owned Session context, and shutdown handoff requirements. Host owns the Admission Gate and Runtime context holder, while the composition bridge gives Handshake only live read-only capabilities. Host does not implement Handshake decisions.

### Accepted ADRs

- [ADR-0002](../adr/0002-configuration-dsl.md) makes Published ConfigurationVersion the source of truth and requires unsupported behavior to fail explicitly.
- [ADR-0003](../adr/0003-runtime-architecture.md) requires independent components, explicit dependency injection, immutable Snapshot, and a production composition root without selecting a DI framework.
- [ADR-0004](../adr/0004-handshake-runtime-dependencies.md) fixes the read-only capability boundary between Host-owned lifecycle state and Handshake.

## 17. Open Questions

- Which concrete production Secret Resolver is selected for the first Beta environment?
- Which startup readiness operation, if any, can validate a remote Secret backend without reading request-scoped credentials?
- How are multiple cleanup errors represented and exposed without losing the primary failure?
- What supervision signal represents unexpected Listener or other mandatory component termination?
- Which component owns the Runtime shutdown wait set before a focused Session Manager design exists?
- Which diagnostics sink receives lifecycle, startup, rollback, and terminal component errors?
- What process-level owner creates and replaces Host instances when a new Published Snapshot is activated?

Reload, restart, Router behavior, Delivery semantics, Plugin ABI, and service-discovery APIs are outside scope rather than open questions for this proposal.

## 18. Decision

Runtime Host is the only production composition root for one Runtime instance.

It receives one immutable Snapshot and explicit operational inputs, constructs the supported component graph through ordinary constructors and existing subsystem Bootstrap, validates readiness before opening traffic, starts Listener last, and owns the Admission Gate, stable Runtime context holder, and root Runtime context. The holder becomes active only at successful startup commit. Runtime composition supplies Handshake with read-only Runtime capabilities, and Host performs rollback or shutdown in reverse acquisition order.

The selected design centralizes composition without centralizing business behavior. Host does not route, authenticate, store Sessions, persist data, or expose internal dependencies. It uses no service locator, DI framework, reflection, generic component factories, or Universal Component Registry. Explicit wiring is preferred because it keeps dependency direction, ownership, failure, and cleanup visible in code and review.

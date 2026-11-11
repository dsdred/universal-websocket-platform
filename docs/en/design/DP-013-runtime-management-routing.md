# DP-013: Runtime Management Routing

[Russian version](../../ru/design/DP-013-runtime-management-routing.md)

## 1. Status

- **Design Status:** Draft
- **Implementation Status:** Start/Stop/Observe implemented in isolation; C2-Exposure section 32 Planned
- **Integration Readiness:** Blocked; full integration and Production
  Activation require the Approved/Planned DP-019 prerequisites and wiring below

This proposal remains non-normative until approved. Its isolated in-process
management command boundary is implemented in `internal/runtimemanagement`.
Production Control Service routing, a concrete authorization policy,
external/process-restart persistence, recovery, and Production Activation
remain absent.

## 2. Purpose

Define the smallest boundary that routes authorized Start, Stop, and Observe
commands to exactly one immutable Runtime Instance scope containing exactly
one Runtime Lifecycle Owner and one Runtime Launch Flow.

## 3. Authority

This proposal refines, without overriding:

- [ADR-0002](../adr/0002-configuration-dsl.md);
- [ADR-0003](../adr/0003-runtime-architecture.md);
- [ADR-0004](../adr/0004-handshake-runtime-dependencies.md);
- [ARCH-002](../architecture/ARCH-002-runtime-foundation-freeze.md);
- [ARCH-004](../architecture/ARCH-004-runtime-deployment-and-identity-model.md);
- [ARCH-005](../architecture/ARCH-005-runtime-configuration-snapshot-and-loading-model.md);
- [DP-010](DP-010-runtime-lifecycle-owner-contract.md);
- [DP-011](DP-011-runtime-launch-pipeline-integration.md);
- [DP-012](DP-012-runtime-source-composition.md).
- [DP-021](DP-021-private-exact-scope-managed-start-invoker.md) for the exact
  composition-private managed Start invocation contract and partial isolated
  implementation.

Accepted ADR, Frozen foundation, and Active architecture remain authoritative.

## 4. Scope

The design covers one process-local command directory, immutable Runtime
Instance bindings, exact identity routing, a policy-neutral authorization
seam, authorization-before-mutation ordering, concurrency, cancellation,
failure preservation, composition audit, and isolated proof evidence.

It does not define transport resources or data transfer objects.

## 5. Package and responsibility

The isolated package is `internal/runtimemanagement`.

It owns only:

- validation of management targets;
- exact lookup of one immutable Runtime Instance binding;
- authorization admission for Start, Stop, and Observe;
- delegation to the binding's exact Flow or Owner.

It does not own Runtime lifecycle state, Configuration loading, Snapshot
construction, Host resources, authorization policy, persistence, recovery, or
transport mapping.

## 6. Terms

### Target

An immutable assertion of Workspace, Configuration, and Runtime Instance
identity.

### Binding

An opaque immutable construction input containing one Target, one Owner, and
one Loader. A caller never supplies a Flow.

### Scope

The private accepted directory entry containing one Target, the Binding's
Owner, and the one Flow constructed by the Directory from that same Owner and
Loader.

### Directory

The single focused management command boundary. Its private map is immutable
after construction and resolves only Runtime Instance scopes. It is not a
dynamic registry, generic resolver, or service locator.

## 7. Exact public surface

```go
package runtimemanagement

var (
    ErrInvalidBinding          error
    ErrInvalidDirectory        error
    ErrInvalidContext          error
    ErrInvalidTarget           error
    ErrRuntimeInstanceNotFound error
)

type Action string

const (
    ActionStart   Action = "start"
    ActionStop    Action = "stop"
    ActionObserve Action = "observe"
)

type Target struct {
    // private immutable identities
}

func NewTarget(
    workspaceID uint64,
    configurationID uint64,
    runtimeInstanceID runtimeconfigload.RuntimeInstanceID,
) (Target, error)

func (t Target) WorkspaceID() uint64
func (t Target) ConfigurationID() uint64
func (t Target) RuntimeInstanceID() runtimeconfigload.RuntimeInstanceID

type Authorize func(
    context.Context,
    Action,
    Target,
    uint64,
) error

type Binding struct {
    // private immutable Target, Owner, and Loader
}

func NewBinding(
    target Target,
    owner *runtimelifecycle.Owner,
    loader *configurationloader.Loader,
) (Binding, error)

type Directory struct {
    // private immutable Runtime Instance scope map
}

func NewDirectory(
    authorize Authorize,
    bindings ...Binding,
) (*Directory, error)

func (d *Directory) Start(
    ctx context.Context,
    target Target,
    configurationVersionID uint64,
) (runtimelifecycle.StartOutcome, error)

func (d *Directory) Stop(
    ctx context.Context,
    target Target,
) (runtimelifecycle.StopOutcome, error)

func (d *Directory) Observe(
    ctx context.Context,
    target Target,
) (runtimelifecycle.Observation, error)
```

No exported command DTO, interface, option, mutable scope, registration
method, list operation, replacement operation, or lifecycle declaration is
added.

## 8. Sentinel identities

The exact sentinels and strings are:

```go
var (
    ErrInvalidBinding = errors.New(
        "invalid Runtime management binding",
    )
    ErrInvalidDirectory = errors.New(
        "invalid Runtime management directory",
    )
    ErrInvalidContext = errors.New(
        "invalid Runtime management context",
    )
    ErrInvalidTarget = errors.New(
        "invalid Runtime management target",
    )
    ErrRuntimeInstanceNotFound = errors.New(
        "Runtime Instance not found",
    )
)
```

Callers use `errors.Is` for these identities. The boundary adds no
authorization, persistence, retry, transport, or diagnostics sentinel.

## 9. Target

`NewTarget` requires non-zero Workspace and Configuration IDs and a non-empty
Runtime Instance ID. Failure returns bare `ErrInvalidTarget`.

Target fields are private and exposed only through value accessors. The zero
value is invalid. Target contains no ConfigurationVersion, Launch Attempt,
Host, PID, Listener, Principal, policy, or transport identity.

## 10. Policy-neutral authorization seam

`Authorize` is a named function type rather than an interface. A nil function
is directly detectable without reflection or a typed-nil detector.

The arguments are:

- caller context;
- exact `ActionStart`, `ActionStop`, or `ActionObserve`;
- the verified Target;
- exact ConfigurationVersion ID for Start, or zero for Stop and Observe.

A non-nil authorization error is returned as the same error interface,
unchanged and unwrapped. The Directory performs no text inspection,
normalization, logging, `errors.Join`, transport mapping, or recovery.

A panic from a non-nil authorization function is a composition or policy
defect. The Directory does not recover it. Lifecycle mutation has not begun
because authorization precedes delegation.

This seam defines ordering, not a concrete Principal model or policy.

Directory borrows one immutable `Authorize` function value for its lifetime.
The function and every captured dependency must be safe for concurrent
invocation. Directory may call it concurrently for the same or different
Targets and adds no synchronization around it.

A conforming Authorize permits different Runtime Instance Targets to progress
independently. It may briefly synchronize internal policy state, but must not
hold a cross-scope or global lock across external I/O, blocking waits, or
callback execution. Same-Target calls may overlap; every result belongs only
to its invocation and is never cached as lifecycle authority.

## 11. Binding construction

`NewBinding` accepts Target, Owner, and Loader, but never a Flow.

Before accepting a Binding it:

1. validates the Target;
2. rejects a nil Owner or Loader;
3. reads `Owner.Observe()` without mutation;
4. verifies exact Workspace, Configuration, and Runtime Instance identities.

Any failure returns bare `ErrInvalidBinding`. Binding construction performs no
Flow construction, loading, building, launch, Stop, authorization, goroutine,
or repository access.

## 12. Static Owner-to-Flow binding

`NewDirectory` first validates the complete Binding set. Only after every
Binding passes does it call:

```go
runtimelaunchflow.New(binding.owner, binding.loader)
```

exactly once for each accepted Binding. The resulting Flow is stored in the
same private scope as that exact Owner.

Because callers cannot supply a Flow, the public API cannot pair a Flow with a
different Owner. Flow introspection, a new Flow accessor, and changes to
existing packages are unnecessary.

The constructor can prevent duplicate Runtime Instance IDs and duplicate
Owner pointers inside one Directory. It cannot detect a separately
constructed bypass Flow or another Directory in a different composition.
Those are Composition Audit obligations, not reasons to add global state.

## 13. Directory construction

`NewDirectory` rejects with bare `ErrInvalidDirectory`:

- nil `Authorize`;
- zero Bindings;
- a zero or invalid Binding;
- duplicate Runtime Instance ID;
- reuse of one Owner pointer in more than one Binding;
- an impossible Flow construction failure after validation.

All Bindings are validated before any Flow is constructed. No partial
Directory is returned. Construction creates no lifecycle state, command,
Launch Attempt, Runtime resource, goroutine, cache, or background work.

The constructor copies the accepted routing entries into a private map that is
never mutated afterward.

The accepted Authorize value is also immutable after construction. Directory
does not replace, wrap, serialize, or otherwise manage it.

## 14. Exact command targets

Start accepts Target plus one non-zero exact ConfigurationVersion ID.

Stop and Observe accept only Target. They do not accept or infer a version.

Commands never target ConfigurationVersion as execution identity, Host
pointer, PID, Listener address, Session, goroutine, context, or transport
resource. There is no latest/current selection, replacement identity, or
fallback scope.

## 15. Validation and error precedence

Every command applies this precedence:

1. nil or invalid Directory receiver returns bare `ErrInvalidDirectory`;
2. nil context returns bare `ErrInvalidContext`;
3. invalid Target, and additionally a zero Start version, returns bare
   `ErrInvalidTarget`;
4. an already visible `ctx.Err()` returns that exact context error;
5. exact lookup uses Runtime Instance ID;
6. absent Runtime Instance or Workspace/Configuration mismatch returns the
   same bare `ErrRuntimeInstanceNotFound`;
7. `Authorize` is called exactly once;
8. only a nil authorization result permits downstream delegation.

Lookup and identity comparison are read-only and perform no lifecycle
mutation. Normalizing absence and mismatch prevents this boundary from
creating a separate identity-oracle distinction.

## 16. Start

After successful validation, routing, and authorization, Start invokes only:

```go
scope.flow.Start(
    ctx,
    runtimelifecycle.NewStartRequest(
        scope.target.WorkspaceID(),
        scope.target.ConfigurationID(),
        configurationVersionID,
    ),
)
```

The exact requested version is preserved. Directory does not call
`Owner.PrepareStart`, `Owner.Start`, Loader, Builder, `runtime.Launch`,
`runtime.Bootstrap`, or `Host.Start` directly.

The returned `StartOutcome` and error are returned unchanged.

The implemented-in-isolation DP-016 orchestration and planned DP-017 recovery
keep this exported `Directory.Start` surface unchanged and use one private
management-only Start-claim continuation seam in the stored Flow. The exact
management composition supplies
that continuation with borrowed capabilities for DP-015 pending-Stop
coordination and DP-014 conditional execution binding. It also supplies the
opaque Control Service execution generation. Directory does not allocate the
generation, access persistence directly, or expose either capability.

After Owner claims the exact attempt and before Flow begins Load, the seam:

1. resolves an already admitted pending Stop first;
2. otherwise conditionally binds the exact attempt and expected aggregate
   revision to the exact generation through DP-014;
3. after confirmed binding, atomically orders one final Stop claim against
   release of `Continue` to Flow.

A pending Stop signals Owner claim to the original blocked Stop call stack;
that claimant alone invokes this binding's exact `scope.owner.Stop`, publishes
its result, and returns `StopConverged`. A confirmed binding plus final release
returns `Continue`. Only coherently proven binding absence for the exact still-
active attempt at the expected revision enters the same final gate against
pending Stop and may return `BindingFailed` without publishing any lifecycle or
command fact. Different generation, stale/conflicting/inactive facts,
unavailable state, or unknown is re-read and converges to an exact existing
terminal outcome or returns `Blocked`; it never becomes BindingFailed. Flow then
uses the authentic preparation token to
call existing Owner.Start with `FailedPreparation`; Owner's mutex orders that
failure against a later Stop. Only the exact returned Owner outcome, followed
by confirmed DP-014 terminal publication, permits command/phase terminalization.
Permit loss, unproven Stop convergence, binding/terminal-publication
indeterminacy, or unknown exact inspection returns `Blocked`, releases no
preparation work, and closes the linked DP-015 barrier.

The command-admission decisions occur before either call stack waits. The
continuation carries no command/recovery permit or caller context, and no
admission or Owner lock is held across persistence, signal, result wait, or
Owner convergence. A Stop losing the final release gate uses the ordinary
section 17 route and reaches the already-claimed attempt. The DP-013 binding
supplies the internal-package-callable `StartClaimContinuation` when
constructing Flow; this adds no exported Directory/Replace/Rollback operation,
transfers no mutable `LaunchPreparation`. The managed continuation is
implemented and independently accepted in isolation outside this package.

Approved [DP-019](DP-019-runtime-activation-orchestration-prerequisites.md)
closes the design ambiguity around this seam. It adds a separate exact
orchestration authorization tuple and a private composition-only lifecycle
invoker used only by the original primitive/phase permit holder. That invoker
passes one per-call immutable `StartExecutionBinding` to the managed Start
surface of the one already stored Flow; revision, generation, authorization,
linked execution, and rendezvous facts are never Flow-construction state or a
mutable Directory slot. Existing public `Start`, `Stop`, `Observe`, and their
authorization behavior remain unchanged. The managed command, binding, Flow,
and continuation surfaces are implemented in their focused internal packages;
TASK-043 now implements the DP-013 composition-private invoker that joins the
scope-bound Flow surface in isolation.
Draft [DP-021](DP-021-private-exact-scope-managed-start-invoker.md) now fixes
that invoker's ownership and exact contract: existing
`runtimemanagement` composition constructs one scope-bound managed Flow exactly
once, then gives the invoker copied domain/Target and a borrowed immutable
reference to that preconstructed Flow. The invoker exposes only synchronous
`InvokeManagedStart`, creates zero Flow, performs no
second policy authorization, stores no per-call capability, never calls the
command Boundary, and adds no public Directory operation. TASK-043 implements
this exact invoker in isolation. TASK-026 now implements callback custody,
terminal work, and orchestration in isolation; production composition and
wiring remain absent.

If the linked `Directory.Start` path returns a definitive cancellation or error
before Owner claim, it signals `StartNoClaim` to the original pending Stop call
stack. That claimant alone terminalizes its Stop satisfied without invoking
section 17. An indeterminate return or lost signal yields `Blocked`, not
`StartNoClaim`.

## 17. Stop

After successful validation, routing, and authorization, Stop invokes only:

```go
scope.owner.Stop(ctx)
```

It does not call Flow, select an attempt, retry cleanup, add a timeout, or
become another shutdown owner. The returned `StopOutcome` and error are
returned unchanged.

## 18. Observe

Observe is authorization-gated even though it performs no lifecycle mutation.
After authorization it performs one final `ctx.Err()` check. If nil, it calls:

```go
scope.owner.Observe()
```

exactly once and returns the coherent immutable Observation. It exposes no
Host, context, cancellation, raw failure, Configuration payload, or Secret.

## 19. Authorization-before-mutation

No Flow or Owner lifecycle method is called before authorization succeeds.

Invalid input, missing or mismatched identity, context cancellation,
authorization denial, and authorization failure therefore cause zero
lifecycle mutation. Directory neither caches an authorization result nor
turns it into ongoing lifecycle authority.

Policy revocation, authorization audit storage, and concealment at a
particular transport are separate contracts.

A nil authorization return is per-command admission only. It is not a
lifecycle linearization point: Flow and Owner retain all claim, conflict,
operation, and outcome linearization.

## 20. Concurrency and linearization

Directory adds no per-Instance mutex. Owner remains the only serialization
boundary for state-changing operations on one Runtime Instance.

- concurrent Start commands reach only the same Flow; its Owner arbitrates
  them, accepts at most one claim and operation, and preserves the existing
  conflict or outcome semantics for losing calls;
- concurrent Stop commands reach only the same Owner;
- Observe reads one coherent Owner snapshot;
- Authorize calls for the same or different Targets may overlap;
- different immutable scopes progress independently when the borrowed
  Authorize conforms to its cross-scope progress contract.

The private directory map is read-only after construction. No Directory lock
is held across authorization, Flow, Owner, resource work, or waiting.
Directory contains no mutex, queue, semaphore, or authorization goroutine. One
blocked authorization call blocks only its caller and does not prevent another
scope from entering Directory or a conforming Authorize.

## 21. Cancellation

Cancellation visible before authorization returns the exact context error
without lifecycle mutation. Authorization receives the original context.

A conforming Authorize observes cancellation for its own blocking work and
returns without detached work. Directory cannot force interruption of a
function that ignores its context and does not hide that limitation behind a
goroutine or timeout.

After authorization, Start passes that same context to Flow, whose Caller
Cancellation Gate remains authoritative. Directory does not interrupt,
detach, or override a Start operation after that gate wins.

Stop passes the same context to Owner, whose locked cancellation gate and
wait semantics remain authoritative.

Observe performs the final post-authorization context check described above.
Directory creates no substitute context, timeout, goroutine, channel, or
detached operation.

If cancellation races with a nil authorization return, the original context
reaches the existing Flow or Owner gate; Observe uses its final check. The gate
that observes cancellation determines whether lifecycle mutation may begin,
exactly as specified by DP-010 and DP-011.

## 22. Outcome and failure boundaries

Directory owns only its declared validation and routing errors. Authorization
errors pass unchanged. Existing Flow and Owner method errors, Start outcomes,
Stop outcomes, and Observation values also pass unchanged.

An authorization error affects only its invocation and causes zero downstream
calls. A panic is a policy or composition defect before lifecycle delegation:
Directory performs no recovery, the panic propagates on the caller goroutine,
and no Flow or Owner method is called.

Directory does not collapse or reclassify:

- preparation, Loader, or Builder failure;
- Bootstrap or startup failure;
- Start conflict or attempt-ID failure;
- Stop failure or retained Host ownership;
- caller cancellation after a downstream gate.

An impossible post-routing identity divergence is a composition defect. It
does not trigger fallback to another scope.

## 23. Ownership and lifetime

| Value or resource | Owner |
| --- | --- |
| Binding declaration | Composition root until Directory construction |
| Private routing map and scopes | Directory |
| Authorization policy/function | External composition; borrowed by Directory |
| Runtime Instance and Launch Attempts | Scope Owner |
| Start pipeline | Scope Flow |
| Loader and Source access | Configured Loader |
| Runtime Host reference | Scope Owner |
| Runtime resources | Runtime Host |
| Caller wait | Individual command invocation |

Owner, Loader, Authorize, and all dependencies captured by Authorize must
outlive Directory. Captured policy state remains externally owned and must
preserve the concurrent-safety and cross-scope progress contract. Directory
owns no shutdown hook and transfers no Runtime ownership.

## 24. Composition Audit

Before Directory construction and again before Production Activation,
composition evidence must prove:

1. one immutable scope per Runtime Instance;
2. one Owner per scope and no Owner reused across scopes;
3. one Flow constructed by Directory from that scope's Owner and Loader;
4. no separately constructed or bypass Flow;
5. all Start calls use Directory and the stored Flow;
6. all Stop and Observe calls use Directory and the stored Owner;
7. authorization is the only admission path before delegation;
8. Authorize and its captured state are concurrent-safe and contain no
   cross-scope lock held across blocking work;
9. no package-global directory, dynamic writer, service locator, importer, or
   alternate management path exists.

The Audit is reference-graph and ownership evidence. It is not a runtime
detector, registry, reflection check, global lock, or persistence substitute.

## 25. Dependency direction

The allowed direction is:

```text
future management composition
    -> runtimemanagement
        ├── runtimelaunchflow
        │   ├── runtimelifecycle
        │   └── configurationloader
        ├── runtimelifecycle
        ├── configurationloader
        └── runtimeconfigload
```

Lower-level Runtime, Flow, Owner, Loader, Source, and Host packages must not
import `runtimemanagement`. Directory is never passed into Runtime or used as
a generic dependency container.

## 26. Mandatory implementation prerequisites

The Draft design contract is implemented in a bounded isolated slice. Approved
DP-014 through DP-018 close all focused design gates required by ARCH-004
section 19(2)–(6). Full integration remains blocked until the required
external durability, activation, recovery, and reporting implementations exist
and the isolated routing, identity, and command packages are composed through
an explicit Control Service boundary.

Loader, Snapshot provenance, and schema compatibility are already resolved by
ARCH-005 and DP-007 through DP-012. The isolated implementation precedent of
DP-010 through DP-012 does not create an exception to the higher-status
ARCH-004 gate. Approved DP-014 through DP-018 define the dependency-ordered
contracts. DP-014 through DP-016 now provide isolated process-local packages;
external adapters, schemas, wiring, and DP-017/DP-018 implementations remain
absent.

## 27. Isolated implementation proofs

The bounded isolated implementation proves the following local contract. Full
integration must repeat the applicable composition proofs after every required
dependency in section 26 is implemented:

1. exact exported surface and sentinel strings;
2. Target validation and immutable accessors;
3. nil authorization is rejected without reflection;
4. Binding rejects nil dependencies and Owner identity mismatch;
5. Directory rejects zero, duplicate, and reused-Owner bindings before Flow
   construction;
6. Directory constructs exactly one Flow from each accepted Owner and Loader;
7. invalid, missing, mismatched, denied, failed, or cancelled commands perform
   zero downstream lifecycle calls;
8. authorization occurs exactly once before Start, Stop, and Observe
   delegation;
9. Start preserves exact version and calls only the exact stored Flow once;
10. Stop calls only the exact stored Owner once;
11. Observe is authorized, coherent, and non-mutating;
12. authorization and downstream errors retain exact interface identity and
    cause chains, and each authorization error affects only one invocation;
13. authorization panic propagates without recovery and performs zero
    downstream calls;
14. concurrent same-Target and different-Target Authorize calls are race-safe;
    one blocked scope does not prevent another conforming scope from
    authorizing and delegating;
15. cancellation before and during Authorize performs no detached work, and a
    nil return racing cancellation remains governed by the existing
    Flow/Owner or final Observe gate;
16. same-scope lifecycle concurrency remains governed by Owner, while
    different conforming scopes progress independently;
17. no fallback, latest selection, retry, re-read, detached work, new mutex,
    dynamic registration, or package-global state exists;
18. Composition Audit proves no bypass path or cross-scope authorization
    serialization;
19. targeted, stress, race, dependency, and full repository checks pass when
    technically available.

Tests may use local fakes or package-private seams. They do not authorize
Control Service wiring or Runtime activation.

## 28. Explicit deferrals

The section 19 focused design gates are closed. DP-014 operational identity and
DP-015 command idempotency now have isolated process-local in-memory
implementations. The following production dependencies remain required before
integration:

- external durable operational identity storage and its integration adapter;
- external durable management command storage and its integration adapter;
- activation, replacement, and rollback orchestration;
- recovery and reconciliation;
- operational error reporting and redaction.

Additional concerns deferred beyond this design:

- HTTP paths, JSON DTOs, status codes, and transport error concealment;
- authentication transport, Principal representation, and concrete
  authorization policy;
- Runtime Instance create, delete, list, or rebinding;
- persistence implementation, schema, transactions, and hydration after its
  focused design;
- retry, automatic restart, reload, and process supervision beyond the
  mandatory focused contracts;
- metrics, audit storage, and alerting beyond the mandatory reporting and
  redaction design;
- application production wiring and Production Activation.

None may be hidden inside Directory, Binding, Authorize, Owner, or Flow.

## 29. Implementation boundary

Implementation Status is Implemented in isolation. Package
`internal/runtimemanagement` provides the exact Target, Binding, Directory,
authorization seam, routing, and local proof tests from this design. It has no
Control Service binding, endpoint, concrete policy, persistence adapter,
recovery path, or activation path.

Approved DP-014 through DP-018 close the focused design gates in ARCH-004
section 19(2)–(6). Packages `internal/runtimeidentity` and
`internal/runtimecommandidempotency` implement DP-014 and DP-015 in isolation
over process-local in-memory storage. External durability, activation,
recovery, reporting, Control Service integration, and Production Activation
remain absent. Any integration slice must first provide its exact required
dependencies, including composition of the existing isolated private
Start-claim continuation and execution-generation binding/load gate defined by
Approved DP-019. Draft DP-021 defines the exact composition-private invoker and
its Implementation Status is Partial — implemented in isolation by TASK-043.
TASK-026 now implements callback/terminal/orchestrator integration in isolation;
production composition remains absent.

## 30. Decision

The Runtime Management Directory is one immutable process-local
command boundary. It validates and routes exact Runtime Instance identity,
authorizes Start, Stop, and Observe before lifecycle delegation, constructs
one Flow from the same Owner and Loader for each accepted scope, and preserves
all downstream lifecycle outcomes.

It adds no second lifecycle owner, dynamic registry, service locator,
persistence simulation, retry, transport API, or Production Activation.

The Draft design is implemented in isolation. Full integration and Production
Activation remain blocked by the unimplemented dependencies and wiring
identified above.

## 31. C2 diagnostics consume-admission boundary (TASK-121)

This section records the TASK-121 boundary required by C2-Binding. Section 32 supplies
the TASK-123 selected exposure design; neither adds a current Directory method,
endpoint or production capability.

`ActionObserve` authorizes only detached lifecycle observation and never authorizes a
destructive diagnostics Poll. DP-018 command/replay authority is also inapplicable.
The C2-Exposure design in section 32 selects a distinct diagnostics-consume authorization
contract whose input includes the exact `Target` and expected non-empty Launch Attempt
identity. It must not overload the current `Authorize` version parameter.

The mandatory order for each future request is:

1. validate context, syntactic Target, expected attempt identity and Poll budget;
2. call the dedicated consume authorizer before target lookup, Owner access or Poll;
3. on denial, cancellation or invalid input, return without Directory lookup that can
   reveal existence and with zero mailbox read;
4. after authorization, select the exact immutable Directory scope; absent/mismatched
   targets use the same externally safe unavailable class;
5. invoke Owner's expected-attempt diagnostics operation, which atomically matches the
   active attempt and non-retired private lifetime before copying its View;
6. perform at most one bounded `View.Poll` outside Directory/Owner locks.

Current security authorization that wins permits one admitted invocation; later grant
revocation affects subsequent requests. Separately, Owner lifetime retirement visible
before its locked View claim prevents Poll; retirement after that claim does not
retract the admitted Poll. Section 32.2 specifies these independent cuts. A successor attempt never satisfies an
older expected-attempt request. Unauthorized, nonexistent, retired and mismatched
requests consume no event and must not disclose which condition occurred.

Section 32 and DP-027 section 37 specify concrete policy, rate/concurrency/pending-request
bounds, transport, output rendering, retention and redaction as Draft design. It may not add an
unbounded target/handle registry, background worker or lifecycle wait. Directory keeps
only its existing Owner scope; the raw Reader/View handle remains private to Owner.

## 32. C2-Exposure consume contract (TASK-123)

This is the selected **Draft / Planned** exposure design, not an implemented route
or production authorization. Section 31 records the TASK-121 prerequisite; this
section supplies its concrete consume policy and admission contract. Existing
Start/Stop/Observe behavior is unchanged. Resource, result, retention and proof
requirements are canonical in [DP-027 section 37](DP-027-runtime-diagnostics-v1.md#37-c2-exposure-resources-results-and-proofs-task-123).

### 32.1 Topology and trust

The only selected deployment is a future single Control Service process containing
one immutable Directory, one Lifecycle Owner per Runtime Instance and its in-process
Host. Current `cmd/control-service` constructs none of these; its HTTP server and
stdout logger gain no diagnostics authority from this design. Host remains the sole
operational graph/lifecycle owner. Reader/View remain private to Owner; Producer uses
only the existing launch-private binding and never the external API. No fifth Bootstrap
binding, mutable dependency rebind, worker, logger fallback, dynamic target registry
or cross-process capability transfer is introduced.

The planned wire adapter uses **TLS 1.3 mutually authenticated HTTP/1.1**, with HTTP/2
disabled for this first contract. Trusted application authentication maps a verified,
configured client identity to an opaque construction-issued principal handle.
Request bodies, headers, query parameters, arbitrary context values, unverified
forwarded identities and client-asserted grants cannot construct that handle.
There is no cookie, bearer, ambient-browser, redirect or CORS authority.

Composition preconfigures at most **32 principal slots** and **64 immutable Target
slots**. Unknown/zero/unconfigured handles fail closed. Requests create no principal,
Target or attempt slots. The future adapter must enforce the bounded ingress in
DP-027 section 37 before exposing the route; this does not modify the current server.

### 32.2 Exact current consume authority

The dedicated action is `runtime.diagnostics.consume`, separate from ActionObserve,
Start/Stop, the existing Authorize version parameter and DP-018 reporting/replay or
durable command authority. One evaluation receives the verified principal handle,
this action, exact Workspace/Configuration/Runtime Instance Target, nonempty expected
Launch Attempt ID, requested Poll budget and current caller context.

The concrete policy is **explicit current exact-grant membership**. The trusted grant
snapshot must enable that principal/action/full Target/expected-attempt tuple. No
wildcard, admin/role inheritance, cross-Workspace fallback, latest-attempt selection
or client-asserted grant authorizes consumption. A snapshot contains at most **2,048
grants** (32 principals x 64 Targets), with at most one current expected-attempt grant
per principal/Target; it retains no grant history or diagnostic batch.

A trusted policy provider atomically supplies the current snapshot. Missing provider,
unavailable evaluation, malformed/unknown snapshot, absent/disabled/stale exact grant
or canceled evaluation denies. Missing production policy must never become allow-all.
The endpoint cannot provision/update grants. Authenticated trusted application grant
provisioning and revocation remain a separate **production activation prerequisite**;
this document does not claim that backend exists. The provider must satisfy the
trusted prompt, concurrent-safe and nonpanicking callback contract; no remote policy
I/O, detached evaluation or timeout goroutine is hidden inside this boundary.

Successful current evaluation is the **security admission linearization point** for
exactly one invocation. Its result is not serialized, replayed, cached or transferred
to another request. Security-grant revocation before this cut denies; after it,
revocation affects subsequent requests and does not retract the admitted invocation.
Independently, Owner lifetime retirement before its locked View capture prevents
Poll; retirement after capture does not retract that Poll. No external-policy check
is promised under Owner's mutex, and no atomic multi-lock security/lifetime cut is
invented. This preserves the existing DP-010 seam.

### 32.3 Admission and exact routing order

1. Attempt the process request slot without waiting; reject closed admission before
   any lookup. Ingress bounds apply even before parsing or principal resolution.
2. Validate method, content type, body bound and one fixed DTO. Reject unknown or
   duplicate fields, compression and noncanonical identities without target lookup.
3. Resolve the already-authenticated trusted principal handle, sample cancellation,
   charge process/principal rate buckets and acquire its principal slot independently
   of the requested Target or its existence. Unknown handles fail closed.
4. Evaluate the dedicated current consume authorizer exactly once on the full request.
5. Recheck cancellation. Denial/cancellation before routing makes **zero Directory
   lookup, Owner call and mailbox read**.
6. Only after authorization select the existing immutable Directory scope and compare
   the complete Target. Never route by ConfigurationVersion, PID, Sequence or Host
   pointer; never fall back to another Workspace or a successor/latest attempt.
7. Attempt that configured Target's slot and charge its rate bucket without waiting.
   Absent/mismatched binding and Target quota refusal use the same unavailable result.
8. Invoke Owner.PollDiagnosticsExpectedAttempt once with the original expected attempt,
   context and budget. Owner atomically matches active attempt/lifetime under its
   mutex, captures View, unlocks and performs at most one bounded Poll/Drain.
9. Validate the detached result and construct only the bounded allowlisted response
   in DP-027 section 37. Clear temporary copies on ordinary return/unwind.
10. Release Target/principal/process in-flight slots after response handling finishes,
    on every returning/unwinding path. Never queue, retry, reinsert or detach work.

Admission/counter updates use short nonblocking critical sections; contention refuses
admission instead of creating a hidden waiting queue. No adapter/Directory/quota lock
spans authorizer callbacks, Owner/Poll, serialization or network I/O. Contexts retain
the existing prompt/nonpanicking precondition. Deadlines do not guarantee termination
of hostile or nonreturning trusted callbacks. Existing Directory.admit, which looks up
before management authorization, is not reused as the consume path. Neither d.admit nor
per-request d.valid/Observe/scope iteration may run before consume authorization;
construction-time immutable binding validation is allowed. After authorization, the
new path may directly select the existing package-private immutable map.

### 32.4 Planned request and failure API

The planned resource is `POST /runtime-diagnostics/v1/poll`, Content-Type
`application/json`, one request/response. There is no GET consumption, streaming,
WebSocket, pagination, cursor, compression, automatic retry or idempotency/replay key.
No route or handler is added by this design.

```json
{"workspaceId":"1","configurationId":"2","runtimeInstanceId":"runtime-opaque-id","expectedLaunchAttemptId":"attempt-opaque-id","budget":16}
```

Workspace/Configuration IDs are canonical nonzero uint64 decimal strings: 1..20
digits, within uint64 range, without leading zero/sign/whitespace. Runtime Instance
and expected-attempt IDs are exact opaque 1..128 ASCII characters `[A-Za-z0-9_-]`.
No trim, case folding or normalization occurs. This wire grammar changes no internal
domain type; other internal identities are unavailable through this API version.
Budget is a required JSON integer 1..64. Expected attempt comes from separately
authorized management knowledge; this endpoint never discovers it.

| Condition | Exact external behavior |
| --- | --- |
| TLS authentication fails | TLS handshake failure; no target-aware HTTP response |
| Wrong method | 405, Allow: POST, fixed unavailable body; no lookup |
| Wrong content type, malformed/oversized/noncanonical DTO, duplicate/unknown field or compression | 400, fixed body `{"error":"invalid_diagnostics_request"}`; no lookup/read |
| Process/principal prelookup capacity/rate refusal | 429, fixed unavailable body, independent of requested Target |
| Process shutting down | 503, fixed unavailable body, independent of Target |
| Unknown trusted principal; consume deny/failure/unavailable; absent/mismatched Target; successor/retired lifetime; Target quota refusal; Poll Busy or invalid result | 404, identical bytes `{"error":"diagnostics_unavailable"}` |
| Valid detached result | 200, only the response in DP-027 section 37 |
| Caller disconnect/cancellation | Stop this invocation; if no usable connection, no response; no retry/replay/reinsert |

The fixed unavailable body for 405/429/503 is also
`{"error":"diagnostics_unavailable"}`. All responses use application/json and
Cache-Control: no-store; no ETag, Last-Modified, shared cache or Retry-After. Error
bodies never echo Target/attempt/principal, grant epoch/reason, quotas, existence,
raw causes or internal status. Target-specific Busy/rate refusals are **404**, never
target-aware 429. Anti-enumeration promises the same external class/body and no denied
lookup/read, not exact microtiming equality. Failure after Poll admission may already
have consumed events; its body never promises zero consumption or retraction.

### 32.5 Implementation split and activation gates

The first candidate is an isolated diagnostics admission/resource adapter in only
`internal/runtimemanagement/diagnostics.go` and `diagnostics_test.go`, borrowing the
existing Directory's package-private immutable scopes. It needs no changes to existing
NewDirectory/Start/Stop/Observe/Authorize APIs. A new dedicated adapter API may issue
opaque principal handles at trusted construction; it exposes no raw Owner/Reader/View.
Construction rejects missing evaluator, invalid clock/configuration, more than 32
principals or 64 Targets, and duplicate/reused bindings or principal slots within its
supplied configuration. Copies of one adapter share all quota state. A constructor
cannot detect separately constructed adapters or other Directories without a new
global registry. Exactly one sanctioned adapter per process and absence of bypass
paths are mandatory Composition Audit/activation proofs, not global enforcement
claimed by construction. It creates no production authentication backend, route or
handler. Test grants are not production policy.

This core is a candidate **Ready for separate implementation intake only after final
design Acceptance**, exact scope/coverage confirmation and separate Owner permission.
A later independent candidate is bounded mTLS transport/serialization against a fake
core. Neither candidate is activated here. Production integration is **NOT READY**
until trusted mTLS configuration, identity-to-handle mapping, actual current exact-grant
provisioning/revocation, bounded ingress, one sanctioned adapter/Directory/Owner
composition, relevant ARCH-004 section 19 dependencies and integrated P01-P08 proofs
exist. An unavailable fail-closed backend is not useful production consume authority.
No isolated allow-all fixture may cross that gate. C2/C3 remain Not Activated;
Gate 11 remains OPEN. No Frozen/Approved contract is amended.

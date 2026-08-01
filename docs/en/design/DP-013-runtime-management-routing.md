# DP-013: Runtime Management Routing

[Russian version](../../ru/design/DP-013-runtime-management-routing.md)

## 1. Status

- **Design Status:** Draft
- **Implementation Status:** Planned
- **Implementation Readiness:** Blocked by the remaining mandatory focused
  designs in ARCH-004 section 19

This proposal is non-normative until approved. It defines a planned isolated
in-process management command boundary. The package, production Control
Service routing, authorization policy, persistence, recovery, and Production
Activation do not exist as a result of this document.

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

Accepted ADR, Frozen foundation, and Active architecture remain authoritative.

## 4. Scope

The design covers one process-local command directory, immutable Runtime
Instance bindings, exact identity routing, a policy-neutral authorization
seam, authorization-before-mutation ordering, concurrency, cancellation,
failure preservation, composition audit, and future isolated proof
requirements.

It does not define transport resources or data transfer objects.

## 5. Package and responsibility

The planned package is `internal/runtimemanagement`.

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

## 7. Exact planned public surface

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

The future DP-016/DP-017 implementation keeps this exported `Directory.Start`
surface unchanged but requires one private management-only Start-claim
continuation seam in the stored Flow. The exact management composition supplies
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
transfers no mutable `LaunchPreparation`, and is Planned rather than
implemented.

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

The design contract is valid, but implementation readiness is blocked.
ARCH-004 section 19 is Active and takes precedence over this Draft. No
`internal/runtimemanagement` package or local proof implementation may begin
until focused designs resolve all remaining mandatory prerequisites:

1. approval/status decision for the section 19(2) candidate persistence
   contract;
2. management command durable idempotency;
3. activation, replacement, and rollback ordering;
4. recovery and reconciliation after Control Service termination;
5. operational error reporting and redaction.

Loader, Snapshot provenance, and schema compatibility are already resolved by
ARCH-005 and DP-007 through DP-012. The isolated implementation precedent of
DP-010 through DP-012 does not create an exception to the higher-status
ARCH-004 gate.

Draft [DP-014](DP-014-runtime-operational-identity-persistence.md) proposes the
candidate Runtime Instance and Launch Attempt persistence contract required by
ARCH-004 section 19(2). It defines durable aggregate, identity, history,
conditional revision, and indeterminate-outcome semantics without creating
persistence implementation. Because it remains non-normative Draft, section
19(2) remains a formal implementation blocker until a separate
approval/status decision.

Draft [DP-015](DP-015-runtime-management-command-idempotency.md) now proposes
the candidate durable management command idempotency contract for section
19(3). It binds one authorized command identity to immutable intent before
lifecycle delegation and defines non-mutating replay. As a non-normative Draft
it does not remove the section 19(2) or 19(3) gates or activate implementation.

Draft [DP-016](DP-016-runtime-activation-replacement-rollback.md) now proposes
the candidate activation, replacement, and explicit rollback ordering contract
for section 19(4). It preserves exact-version attempts and Stop-to-proven-release
before any replacement or rollback Start. As a non-normative Draft it does not
remove the section 19(2), 19(3), or 19(4) gates or activate implementation.

## 27. Future implementation proofs

After every prerequisite in section 26 is resolved, a future implementation
task must prove:

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

Remaining blocking architecture prerequisites before any implementation:

- approval/status decision for the section 19(2) candidate persistence
  contract;
- durable management command idempotency;
- activation, replacement, and rollback ordering;
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

Implementation Status remains Planned. No `internal/runtimemanagement`
package, test, Control Service binding, endpoint, policy, persistence adapter,
or activation path exists as a result of this document.

Implementation Readiness is Blocked. Neither an isolated package nor local
proof code is authorized until every mandatory focused design in section 26
exists. DP-015, DP-016, and
[DP-017](DP-017-runtime-recovery-reconciliation.md) are the candidate section
19(3), 19(4), and 19(5) designs, not implementation tasks; the section
19(2)–(5) approval gates remain. By dependency ordering the next design
recommendation may address operational error reporting and redaction in
section 19(6).

## 30. Decision

The planned Runtime Management Directory is one immutable process-local
command boundary. It validates and routes exact Runtime Instance identity,
authorizes Start, Stop, and Observe before lifecycle delegation, constructs
one Flow from the same Owner and Loader for each accepted scope, and preserves
all downstream lifecycle outcomes.

It adds no second lifecycle owner, dynamic registry, service locator,
persistence simulation, retry, transport API, or Production Activation.

The design is ready for review and possible acceptance as Draft/Planned.
Implementation remains blocked by ARCH-004 section 19(2)-(6).

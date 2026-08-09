# DP-011: Runtime Launch Pipeline Integration

[Russian version](../../ru/design/DP-011-runtime-launch-pipeline-integration.md)

## 1. Status

**Design Status:** Draft

**Implementation Status:** Implemented in isolation; the private DP-016
Start-claim continuation extension defined by Approved DP-019 is Planned

**Architecture Status:** focused integration contract over approved ARCH-004
and ARCH-005 and the existing Draft DP-007 through DP-010.

The `internal/runtimelaunchflow` package implements this contract in isolation.
Concrete Source composition and management routing also exist as isolated
packages under DP-012 and DP-013. Their production composition, Control Service
routing, and Production Activation remain absent. Implementation does not claim
that the production launch capability is implemented and does not raise the
status of any related Draft DP. The current package does not implement the
private claim-continuation gate required by DP-016; that extension requires a
separate implementation task.

## 2. Purpose

Define the minimal in-process flow that connects one existing Runtime
Lifecycle Owner to the Configuration Loader, Snapshot Builder, and stateless
Runtime Launcher.

The Flow must:

- begin a Launch Attempt only through `Owner.PrepareStart`;
- pass exactly the Owner-issued `LoadRequest` to the Loader;
- pass exactly one successful `DetachedLoadResult` to the Builder;
- convert a Loader failure or Builder Diagnostics into the closed
  `PreparationResult`;
- return that result to the same Owner through `Owner.Start`; and
- preserve the ownership, cancellation, and concurrency contracts of
  DP-007 through DP-010.

## 3. Sources of Authority

The contract is constrained by:

- [ADR-0002](../adr/0002-configuration-dsl.md): a Published
  ConfigurationVersion is the immutable declarative source;
- [ADR-0003](../adr/0003-runtime-architecture.md): Runtime dependencies are
  explicit and Runtime does not read Control Plane repositories;
- [ARCH-002](../architecture/ARCH-002-runtime-foundation-freeze.md): Host owns
  production composition, startup, rollback, and shutdown;
- [ARCH-004](../architecture/ARCH-004-runtime-deployment-and-identity-model.md):
  the Lifecycle Owner owns the Runtime Instance, Launch Attempt, and lifecycle
  serialization;
- [ARCH-005](../architecture/ARCH-005-runtime-configuration-snapshot-and-loading-model.md):
  Launch preparation owns construction of the detached Snapshot;
- [DP-007](DP-007-configuration-loader-contract.md): Loader accepts an exact
  five-identity `LoadRequest` and returns one detached result or failure;
- [DP-008](DP-008-snapshot-builder-contract.md): Builder returns one complete
  Snapshot or non-empty exhaustive Diagnostics;
- [DP-009](DP-009-runtime-bootstrap-contract.md): every launch passes through
  stateless `runtime.Launch`;
- [DP-010](DP-010-runtime-lifecycle-owner-contract.md): Owner creates a
  preparation, accepts a closed result, and alone calls the Launcher.

A higher-status source takes precedence over this Draft.

## 4. Scope

DP-011 defines:

- one package-level integration boundary;
- immutable binding of one Flow to one Owner and one Loader;
- the exact `PrepareStart -> Load -> Build -> Start` flow;
- representation of Builder Diagnostics as a preparation failure;
- a synchronous attempt operation and caller cancellation gate;
- Stop/cancellation races;
- dependency direction, ownership, and lifetime;
- local and future production activation proofs.

DP-011 does not define:

- an HTTP endpoint, management command DTO, or authorization;
- a repository adapter or selection of a concrete
  `configurationloader.Source`;
- persistence of Runtime Instance, Launch Attempt, desired, or actual state;
- durable idempotency, retry, restart, replacement, or reconciliation;
- supervision, a terminal Host signal, diagnostics transport, or metrics;
- new Loader, Builder, Launcher, or Owner semantics.

## 5. Terms

### Runtime Launch Flow

An immutable in-process integration object that borrows one Owner and one
Loader and performs exactly one preparation sequence for each accepted Start
request.

### Start Operation

Exactly one synchronous call-stack operation associated with one `Flow.Start`
invocation. After a successful `PrepareStart`, the same caller goroutine
performs Load, Build, and result delivery to the Owner in order. The Flow
creates no goroutine, detached work, or background worker.

### Build Failure

An immutable error containing the complete non-empty set of blocking
Diagnostics from one Builder invocation. It converts the Builder result into
the existing `FailedPreparation` without turning Diagnostics into a Loader or
Bootstrap failure.

### Production Activation

Connection of the Flow to the single authorized management boundary, concrete
Source composition, and operational state storage. Creating the Flow package
alone is not Production Activation.

## 6. Decision

One flow is selected:

```text
authorized caller
    -> Runtime Launch Flow.Start
        -> Owner.PrepareStart
        -> Configuration Loader.Load
        -> Snapshot Builder.Build
        -> Owner.Start
            -> runtime.Launch
                -> Runtime Bootstrap
                    -> Runtime Host
```

The Flow does not independently create a Launch Attempt, Snapshot, Bootstrap
request, or Host. It only connects the existing owners of those objects in
order.

## 7. Package and Responsibility

The first implementation slice belongs in
`internal/runtimelaunchflow`.

The package is responsible only for:

- binding one Owner and one configured Loader;
- invoking the existing stateless Builder;
- converting a complete Diagnostics set into a Build Failure;
- synchronously executing one Start Operation;
- passing a closed preparation result to the original Owner.

The package is not a Lifecycle Owner, source adapter, repository, generic
orchestrator, registry, supervisor, or policy engine.

## 8. Exact Initial Public Surface

The first implementation is limited to this conceptual Go contract:

```go
package runtimelaunchflow

var (
    ErrInvalidFlow         error
    ErrInvalidStartContext error
)

type Flow struct {
    // package-private immutable dependencies
}

func New(
    owner *runtimelifecycle.Owner,
    loader *configurationloader.Loader,
) (*Flow, error)

func (f *Flow) Start(
    ctx context.Context,
    request runtimelifecycle.StartRequest,
) (runtimelifecycle.StartOutcome, error)

type BuildFailure struct {
    // package-private immutable diagnostics
}

func (f *BuildFailure) Error() string
func (f *BuildFailure) Diagnostics() []runtimeconfig.Diagnostic
```

The implemented isolated surface introduces no additional exported interface,
callback, registry, option, stage enum, result union, or lifecycle state. The
private continuation described in section 10 is a planned management
integration seam, not part of the current exported API.

`New` rejects nil Flow dependencies with `ErrInvalidFlow`. `Start` rejects a
nil context with `ErrInvalidStartContext`. Owner, Loader, and context errors
are not reclassified.

## 9. Construction and Dependency Binding

A `Flow` is permanently bound to one `*runtimelifecycle.Owner` and one
`*configurationloader.Loader`.

The constructor:

- does not create a Runtime Instance or Launch Attempt;
- does not invoke Loader, Builder, Launcher, or Host;
- does not read a repository;
- does not create a goroutine;
- does not copy or mutate dependency state.

The Builder is constructed as stateless `runtimeconfig.NewBuilder()` inside the
package or held as a value without mutable state. The production constructor
does not accept a replaceable Builder or Launcher seam.

The concrete Source is selected and passed when constructing the Loader
outside the Flow. The Flow therefore knows no transport, repository, or
deployment adapter.

## 10. Claim and Startup Request

Before lifecycle mutation, `Flow.Start` checks for a non-nil context and reads
`ctx.Err()` exactly once. This read is the **Caller Cancellation Gate**
linearization point. The Flow then calls:

```go
preparation, err := owner.PrepareStart(request)
```

The Flow does not create or repair identity. Start request validation, Launch
Attempt ID allocation, conflict detection, and claim linearization remain
solely with the Owner.

When cancellation races with the Gate:

- cancellation observed by the Gate wins and prevents a claim;
- nil observed by the Gate permits the claim attempt, and later caller
  cancellation does not cancel this Start Operation even when it occurs before
  the Owner's internal claim linearization.

When `PrepareStart` returns an error, the Flow:

- begins no Load or Build;
- does not call Loader, Builder, Owner.Start, or Launcher;
- returns the exact error to the caller.

DP-016 management orchestration and DP-017 recovery require one future private
**Start-claim continuation gate**. It does not move claim authority out of
Owner. Immediately after `Owner.PrepareStart` returns a successful preparation
and before Load, Build, or Launcher work begins, Flow must synchronously offer
an immutable view of that exact claim to the management continuation associated
with the primitive or linked Start command. The Flow-provided view contains
only the exact Runtime Instance and Launch Attempt claimed by Owner. The exact
management continuation is already bound immutably to the expected aggregate
revision and composition-owned execution generation required for DP-014
binding and rejects a mismatched claim view.

The exact Control Service composition creates one opaque execution generation
for its process-containment boundary. DP-014 owns conditional durable
Attempt-to-generation binding. The management continuation coordinates that
binding; Flow, Owner, DP-013 Directory, and Runtime Host neither allocate a
generation nor persist the binding.

The continuation follows this order without holding an Owner or admission lock
across persistence:

1. an already pending Stop is signalled and converged before binding;
2. otherwise the continuation conditionally binds the exact active attempt to
   the exact generation using its expected aggregate revision;
3. after confirmed binding, one final per-Instance gate atomically orders a
   Stop claim against release of `Continue` to Flow;
4. Stop winning that gate is converged before Load; `Continue` winning permits
   Load, and a later Stop reaches the already claimed attempt normally.

The decision meanings are:

- `Continue` means the exact execution binding is confirmed and the final gate
  released preparation before any pending Stop; Flow may begin Load and Build;
- `StopConverged` means the original Stop claiming path used its own permit to
  invoke exact DP-013 Stop and converged that Owner attempt; Flow begins no Load
  or Build and returns the Owner-equivalent stopped-before-running outcome;
- `BindingFailed` means one coherent exact read proves that the binding did not
  commit for the still-active exact attempt at the expected revision and that
  no external preparation began. The continuation terminalizes nothing. Flow
  begins no Load or Build and submits `FailedPreparation(bindingFailure)` with
  the original authentic token to Owner.Start;
- `Blocked` means the pending claimant lost its permit, returned without a
  definitive result, reported unproven Stop convergence, binding or terminal
  publication was indeterminate, exact binding inspection remained unknown, or
  the rendezvous was indeterminate; Flow begins no external preparation work
  and the linked set remains unresolved.

Binding publication uses the exact attempt and expected DP-014 aggregate
revision. Definitive failure performs zero external preparation. After an
indeterminate outcome the same path inspects the exact attempt/generation and
revision: exact same-generation presence permits the final gate; coherently
proven absence for the still-active attempt and expected revision enters a final
gate that orders pending Stop against `BindingFailed`; different generation,
stale revision, conflicting or inactive facts, unavailable state, or still-
unknown is re-read and then converges to an exact existing terminal outcome or
returns `Blocked`. None of those conflicts becomes BindingFailed. Caller
cancellation after the existing Caller Cancellation Gate does not cancel or
detach this binding duty.

If `BindingFailed` wins its final gate, Flow uses the same internal non-caller-
cancelled wait context as other preparation convergence and calls the existing
Owner-owned operation exactly once:

```go
owner.Start(waitContext, preparation, FailedPreparation(bindingFailure))
```

Owner's mutex orders this failure acceptance against a later ordinary Stop. If
failure acceptance wins, Owner returns `StartPreparationFailed`; if Stop wins,
the same token converges to `StartStoppedBeforeRunning`. Only the returned exact
Owner outcome may drive DP-014 terminal publication and later command/phase
terminalization. An absent or indeterminate durable terminal publication is
unresolved. The continuation never publishes lifecycle or command outcomes.

The continuation carries no mutable Host, Snapshot, lifecycle ownership,
permit, or caller-selected identity. It only signals the original pending Stop
call stack that Owner claim succeeded, then waits for that same claimant's
durable outcome. It runs outside Owner's mutex; neither the command-admission
nor Owner lock is held across persistence, either wait, or Stop convergence.
Process loss makes the rendezvous and linked command set unresolved for DP-017
recovery. If process loss occurs after Owner claim but before a confirmed
binding, the durable attempt already exists and is Starting; recovery may prove
only that no external preparation began, not that no lifecycle mutation
occurred.

Because Flow and management routing are separate Go packages, the future
extension uses an internal-package-callable managed surface. One long-lived
Flow is bound immutably at construction to a stateless
`StartClaimContinuation`; each managed Start invocation separately receives
the exact per-call `StartExecutionBinding` defined by DP-019. The capability
exposes one synchronous `AfterOwnerClaim` decision with `Continue`,
`StopConverged`, `BindingFailed`, or `Blocked`; it exposes no mutable
`LaunchPreparation`, command permit, recovery permit, or persistence
implementation. A Go symbol may be exported across the repository's
`internal/` package boundary, but it is not a public management or HTTP API.
The exact current `New` and `Start` implementation remains unchanged
and has no such seam, so it implements neither the DP-016 continuation nor the
DP-017 binding gate.

Approved [DP-019](DP-019-runtime-activation-orchestration-prerequisites.md)
defines the exact per-call binding and claim view, callback-scoped parent/phase
authority, attempt-publication/binding order, and closed continuation outcomes.
It does not change this Flow's Owner-first claim or synchronous preparation
semantics. That prerequisite remains Planned and must be implemented before
DP-016 orchestration work resumes.

## 11. Synchronous Operation and Caller Lifetime

After a successful claim, the same `Flow.Start` invocation synchronously
performs Load, Build, and `Owner.Start`. The Flow creates no goroutine, channel,
detached work, or package-owned operation state.

After the Gate wins, the caller context no longer controls operation lifetime
and is not checked again. `Flow.Start` returns only after one Owner outcome or
exact operation error. This retains one explicit owner of the blocking call
stack and leaves no unawaited preparation work.

If the configured Source blocks, `Flow.Start` and the caller goroutine remain
blocked until the Source returns. The current contract promises no timeout or
forced cancellation. Production Activation must expose that limitation rather
than mask it with a detached worker.

## 12. Exact Loader Handoff

The Start Operation calls the Loader exactly once and only with:

```go
preparation.LoadRequest()
```

It is forbidden to:

- construct a second `LoadRequest`;
- replace any of the five identities;
- select a latest or replacement ConfigurationVersion;
- perform fallback, retry, or a second Load;
- pass Owner, Snapshot dependencies, or management authority to the Loader.

On Loader success, exactly the returned `DetachedLoadResult` is passed to the
Builder. The Flow does not normalize, repair, or reread source material.

## 13. Loader Failure

On a Loader error:

1. Builder is not called;
2. the operation creates `runtimelifecycle.FailedPreparation` with the exact
   error interface;
3. the same Owner receives that result through `Owner.Start`;
4. Launcher, Bootstrap, and Host are not called.

The Flow does not wrap or edit the Loader error. Sentinel identity and cause
chain remain available through `StartOutcome.PreparationFailure()`.

## 14. Exact Builder Handoff

On Loader success, the operation invokes one stateless Builder exactly once:

```go
snapshot, diagnostics := runtimeconfig.NewBuilder().Build(loadResult)
```

The Flow does not perform semantic validation, normalization, or provenance
construction itself.

The result is handled only as the closed DP-008 contract:

- empty Diagnostics means one complete Snapshot;
- non-empty Diagnostics means no Snapshot and a Build Failure.

Malformed output is impossible from the concrete Builder and does not create a
separate production policy boundary.

## 15. Build Failure

For non-empty Diagnostics, the operation creates one `*BuildFailure`.

`BuildFailure`:

- preserves ordering of the complete Diagnostics set;
- owns a detached copy of the slice;
- returns a detached copy from `Diagnostics()`;
- has the exact constant `Error()` `Runtime Snapshot build failed` without
  concatenating locations, values, or source material;
- implements no retry, severity mapping, HTTP mapping, logging, or redaction
  policy.

The exact Build Failure pointer is passed to
`runtimelifecycle.FailedPreparation`. Builder Diagnostics remain a
Builder-owned semantic category and do not become a Loader, Bootstrap, or
Startup failure.

## 16. Snapshot Success

On a successful Build, the operation creates only:

```go
runtimelifecycle.PreparedSnapshot(snapshot)
```

and passes the result to the same Owner with the original opaque preparation
token.

The Owner defensively checks matching five-part provenance before acceptance.
The Flow does not bypass that validation and does not directly invoke
`runtime.Launch`, `runtime.Bootstrap`, or `Host.Start`.

After `Owner.Start` returns, local copies of the `DetachedLoadResult`, Snapshot,
and Diagnostics are not retained by the Flow.

## 17. Stop and Preparation Cancellation

`LaunchPreparation.Context()` is the only signal from the Owner to the
synchronous Start Operation. The Flow receives no cancellation authority.

The operation checks this context:

- before Load;
- after Load returns and before Build;
- after Build returns and before passing the result to the Owner.

The current synchronous Loader contract does not accept a context. Stop
therefore does not promise to interrupt an in-progress `Source.LoadExact`;
after it returns, the operation does not begin the next stage if the Owner has
already cancelled preparation. Timeout, forced cancellation, and a Source API
change are not introduced by this Draft.

If Stop terminalized AttemptPreparing, the operation publishes no second
failure and does not call the Launcher. After observing the cancelled
preparation context, the operation calls `Owner.Start` with the original token,
a zero result,
and an internal non-cancelled wait context. Under DP-010 same-token
convergence, the late result is not validated and the operation receives the
stored `StartStoppedBeforeRunning` without changing the Stop terminal fact.

If Stop races with `Owner.Start`, winner semantics, Host cleanup, and
`StartStoppedBeforeRunning` remain exactly as defined by DP-010.

The same rule applies when `Owner.Start` carries the binding-failure
`FailedPreparation`: a Stop winning Owner's mutex supplies the stored stopped-
before-running outcome; failure winning supplies the Owner-confirmed preparation
failure. Flow and continuation never pre-publish either result.

## 18. Owner.Start Wait Context

After the Caller Cancellation Gate wins, the caller context is not passed as
preparation, Owner wait, or Runtime startup authority. To pass the result, the
operation uses a wait context that is not cancelled by the caller; the
Owner-owned `LaunchPreparation.Context()` still becomes
`BootstrapRequest.StartupContext` inside DP-010.

This separation is required so caller cancellation after the Gate does not
leave AttemptPreparing without a terminal outcome. It does not hide Stop: Stop
mutates Owner state and cancels the preparation context, while Owner
convergence determines the outcome.

The Flow creates no new root Runtime context and receives no Host `CancelFunc`.

## 19. Concurrency and Linearization

One `Flow` may receive concurrent `Start` calls. It does not serialize them
with its own mutex:

- Owner remains the only per-Instance serialization boundary;
- no more than one `PrepareStart` claim becomes active;
- losing calls begin no Load or Build;
- one successful claim performs exactly one synchronous Start Operation;
- one operation performs no more than one Load, Build, and Owner.Start;
- only the Owner may call the Launcher;
- concurrent Flows mistakenly bound to one Owner still converge through that
  Owner, but production composition must create exactly one Flow per Owner.

Different Owner/Flow pairs share no mutable state and may progress
independently.

## 20. Ownership and Lifetime

| Object | Owner |
| --- | --- |
| Runtime Instance and Launch Attempt | Runtime Lifecycle Owner |
| Configured source access | Configuration Loader |
| Load operation/source material before detachment | Loader |
| Detached Load Result between Load and Build | Synchronous Start Operation |
| Diagnostics before Build Failure | Synchronous Start Operation |
| Build Failure after acceptance | Owner outcome |
| Snapshot before acceptance | Synchronous Start Operation |
| Accepted Snapshot and launch operation | Runtime Lifecycle Owner |
| Bootstrap request and construction inputs | Bootstrap for the call |
| Runtime resources | Runtime Host |
| Caller wait | One `Flow.Start` invocation |

The Flow owns no Host reference, lifecycle state, repository transaction,
Snapshot cache, or retry state.

## 21. Dependency Rules

The allowed production dependency direction is:

```text
management composition
    -> runtimelaunchflow
        -> runtimelifecycle
        -> configurationloader
        -> runtimeconfig

runtimelifecycle
    -> runtimeconfigload
    -> runtime
```

Reverse imports from `runtime`, `runtimeconfig`, `configurationloader`, or
`runtimelifecycle` into `runtimelaunchflow` are forbidden.

The Flow is not passed to Host, Bootstrap, Builder, or Loader as an arbitrary
capability. Registries, global singletons, reflection, and service locators
are forbidden.

## 22. Failure Matrix

| Stage | Observed result | Subsequent work |
| --- | --- | --- |
| Invalid context or cancellation observed by Caller Cancellation Gate | `ErrInvalidStartContext` or exact context error | no claim, Load, Build, Start, or Launch |
| Nil observed by Gate; cancellation before or after Owner claim | cancellation does not override the Gate | synchronous operation continues to Owner outcome |
| Owner claim failure | exact Owner error | no Load, Build, Start, or Launch |
| Loader failure | `StartPreparationFailed` with exact Loader error | no Build or Launch |
| Builder Diagnostics | `StartPreparationFailed` with `*BuildFailure` | no Launch |
| Snapshot provenance mismatch | exact Owner `ErrInvalidPreparationResult` | no Launch; blocking contract defect |
| Bootstrap/Startup failure | `StartLaunchFailed` with unchanged `BootstrapOutcome` | Owner records launch failure |
| Runtime success | `StartRunning` | Owner owns Host reference |
| Stop during preparation | DP-010 `StartStoppedBeforeRunning` convergence | no new stage after observed preparation cancellation |

The Flow does not classify management, authorization, persistence, or recovery
failures.

## 23. Acceptance Proofs

The first implementation task must prove:

1. constructor and context validation perform zero lifecycle mutation;
2. exactly `LaunchPreparation.LoadRequest()` reaches the Loader;
3. Loader failure causes zero Build and zero Launch;
4. Loader success reaches the Builder unchanged;
5. Builder is invoked no more than once;
6. the complete Diagnostics set is available through immutable
   `BuildFailure`;
7. Builder failure causes zero Launch;
8. a Snapshot reaches the same Owner and passes five-identity validation;
9. one accepted Snapshot creates exactly one Owner-to-`runtime.Launch` path;
10. production Flow neither imports nor calls `runtime.Bootstrap` or
    `Host.Start`;
11. the Caller Cancellation Gate has exact winner semantics, and cancellation
    after the Gate wins does not interrupt the synchronous operation;
12. Stop before, during, and after Load/Build creates no second result,
    operation, or Launcher call;
13. concurrent Start on one Flow begins no more than one Start Operation;
14. different Owner/Flow pairs progress independently;
15. Flow, Loader, Builder, and Launcher share no package-global mutable state;
16. no resource call or wait occurs under a new Flow mutex because no such
    mutex exists;
17. package and full repository tests pass, including race tests when the
    toolchain supports them.

These proofs close the integration part of DP-009 AP-003 and AP-011 only for
the introduced in-process Flow. Full Production Activation additionally
requires proof that the single management start boundary uses this Flow and
that no bypassing production path exists.

## 24. Production Activation Gates

Before the capability is described as production-integrated, separate tasks
must compose and verify:

- the existing isolated DP-012 Source composition in the production path;
- the existing isolated DP-013 command boundary, authorization-before-mutation
  seam, and exact Owner/Flow routing in the production path;
- the existing isolated DP-014/DP-015 aggregate and command stores together
  with the required external/process-restart durability;
- the Planned DP-016 activation/replacement/rollback orchestrator, including
  the private DP-011/DP-013 Start-claim continuation and DP-017-required
  execution-binding/load gate;
- implementation of the Approved/Planned DP-017 recovery/reconciliation
  contract, or explicit startup rejection while that implementation is absent;
- implementation of the Approved/Planned DP-018 operational reporting and
  redaction contract for preparation/launch failures.

DP-011 does not select the order or API of those tasks. The isolated package
implementation is an independently verified prerequisite, but by itself it
does not change `spec/current-state.md` to claim that Runtime management from
the Control Service is implemented.

## 25. Intentionally Deferred

Deferred:

- additional Source adapters beyond the existing isolated in-memory adapter,
  including PostgreSQL, YAML, or remote transport;
- HTTP/CLI/API surface and authorization;
- external durable persistence schema and transactions;
- process-restart command/result persistence, retention, and recovery;
- activation/replacement/rollback orchestration, private Start-claim
  continuation, and execution-binding/load gate;
- retry, backoff, restart, replacement, rollback policy, and reconciliation;
- terminal Host supervision and unexpected failure;
- timeout/force policy for a blocking Source;
- diagnostics transport, metrics, audit, and redaction policy;
- process isolation, scheduling, clustering, and federation.

None of these concerns may be implemented as hidden Flow behavior.

## 26. Implementation Boundary

The first code slice is implemented and limited to:

- `internal/runtimelaunchflow`;
- local proof tests;
- factual documentation synchronization.

It does not change the DP-007 through DP-010 packages, Control Service,
repositories, management API, persistence, Host, or Bootstrap.

It is not connected to the Control Service and is not Production Activation.
Implementation does not automatically raise Design Status.

## 27. Consequences

Positive:

- the single explicit gap between the implemented isolated contracts is
  closed at the design level;
- Owner remains the lifecycle authority;
- Loader and Builder responsibilities are not mixed;
- AP-003/AP-011 gain a concrete integration proof boundary;
- caller cancellation has one explicit Gate linearization and leaves no
  claimed attempt without a synchronous owner duty;
- production activation gates remain visible.

Costs:

- after the Gate wins, the caller may wait on a blocking Source without
  cancellation;
- a synchronous Source cannot be force-cancelled and a detached workaround is
  forbidden;
- package implementation still does not create a user-facing management
  capability;
- external persistence, concrete authorization, and production integration
  still require separate tasks.

## 28. Conclusion

Runtime Launch Flow is a narrow immutable integration boundary. It uses the
Owner-issued preparation, exact Loader handoff, stateless Builder, and Owner
acceptance in order, after which only the Owner calls the stateless Runtime
Launcher.

The Flow does not become a second lifecycle owner and does not hide product,
persistence, or recovery policy. Until separate Production Activation, the
repository retains the truthful state: the integration contract is defined,
but Runtime management from the Control Service is not implemented.

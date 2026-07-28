# DP-010: Runtime Lifecycle Owner Contract

[Russian version](../../ru/design/DP-010-runtime-lifecycle-owner-contract.md)

## 1. Status

**Design Status:** Draft

**Implementation Status:** Implemented in isolation

**Architecture status:** implementation contract for the approved operational
identity model in
[ARCH-004](../architecture/ARCH-004-runtime-deployment-and-identity-model.md)
and loading model in
[ARCH-005](../architecture/ARCH-005-runtime-configuration-snapshot-and-loading-model.md).

Runtime Lifecycle Owner is implemented in isolation in
`internal/runtimelifecycle`. Production Loader-to-Builder-to-Launcher wiring
is not implemented. This Draft does not revise approved architecture.

## 2. Purpose

Define the smallest in-process Control Service-side owner that:

- owns exactly one Workspace, Configuration, and Runtime Instance scope;
- creates each Launch Attempt and pins its exact ConfigurationVersion before
  Loader or Builder work;
- serializes lifecycle operations for that Instance;
- invokes only the existing stateless Runtime Launcher;
- owns the active Host reference; and
- publishes truthful desired, actual, attempt, and terminal facts.

The first implementation is isolated. Persistence, management APIs, and
production wiring remain deferred.

## 3. Sources of Authority

This contract is constrained by:

- [ADR-0002](../adr/0002-configuration-dsl.md): Published
  ConfigurationVersion is the immutable behavior source;
- [ADR-0003](../adr/0003-runtime-architecture.md): dependencies are explicit
  and Runtime has no Repository or management API access;
- [ARCH-002](../architecture/ARCH-002-runtime-foundation-freeze.md): Host is
  the production composition root and owns startup and shutdown;
- [ARCH-004](../architecture/ARCH-004-runtime-deployment-and-identity-model.md):
  Lifecycle Owner creates Launch Attempts and owns per-Instance orchestration;
- [ARCH-005](../architecture/ARCH-005-runtime-configuration-snapshot-and-loading-model.md):
  Owner pins the exact version before Loader and Builder and supplies complete
  operational provenance;
- [DP-007](DP-007-configuration-loader-contract.md): Loader accepts the neutral
  exact-version `LoadRequest`;
- [DP-008](DP-008-snapshot-builder-contract.md): Builder produces a Snapshot
  with matching five-part provenance;
- [DP-009](DP-009-runtime-bootstrap-contract.md): `runtime.Launch` is the only
  stateless launch boundary.

A higher-status source wins over this Draft.

## 4. Scope and Boundary

DP-010 defines:

- exact exported local declarations and sentinel errors;
- construction and immutable identity binding;
- two-phase `PrepareStart` then `Start`;
- Owner-issued attempt identity and exact version pin;
- preparation, launch, Start, Stop, and observation semantics;
- one synchronization boundary and linearization points;
- caller cancellation and same-token convergence;
- truthful failure retention;
- local proof requirements and future integration gates.

It does not implement or specify an adapter that calls Loader or Builder.

## 5. Package and Responsibility

The package is `internal/runtimelifecycle`, not `internal/runtime`.

One `Owner` is permanently bound to exactly one Workspace, Configuration, and
`runtimeconfigload.RuntimeInstanceID`. It is not a Host, repository, generic
manager, registry, service locator, or policy engine.

Different Owners share no lifecycle state and may progress independently.

## 6. Exact Exported Declarations

The first implementation uses these declarations without adding another public
lifecycle abstraction:

```go
package runtimelifecycle

type DesiredState string

const (
    DesiredStopped DesiredState = "stopped"
    DesiredRunning DesiredState = "running"
)

type ActualState string

const (
    ActualStopped  ActualState = "stopped"
    ActualStarting ActualState = "starting"
    ActualRunning  ActualState = "running"
    ActualStopping ActualState = "stopping"
    ActualFailed   ActualState = "failed"
)

type LaunchAttemptIDSource func() (runtimeconfigload.LaunchAttemptID, error)

type StartRequest struct { /* unexported identities */ }

func NewStartRequest(workspaceID, configurationID, configurationVersionID uint64) StartRequest
func (r StartRequest) WorkspaceID() uint64
func (r StartRequest) ConfigurationID() uint64
func (r StartRequest) ConfigurationVersionID() uint64

type LaunchPreparation struct { /* opaque owner/claim token, immutable LoadRequest, read-only context */ }

func (p LaunchPreparation) LoadRequest() runtimeconfigload.LoadRequest
func (p LaunchPreparation) Context() context.Context

type PreparationResultKind string

const (
    PreparationSnapshot PreparationResultKind = "snapshot"
    PreparationFailure  PreparationResultKind = "failure"
)

type PreparationResult struct { /* closed union, fields unexported */ }

func PreparedSnapshot(snapshot runtimeconfig.Snapshot) PreparationResult
func FailedPreparation(cause error) PreparationResult
func (r PreparationResult) Kind() PreparationResultKind
func (r PreparationResult) Snapshot() (runtimeconfig.Snapshot, bool)
func (r PreparationResult) Failure() (error, bool)

type StartOutcomeKind string

const (
    StartRunning              StartOutcomeKind = "running"
    StartPreparationFailed    StartOutcomeKind = "preparation-failed"
    StartLaunchFailed         StartOutcomeKind = "launch-failed"
    StartStoppedBeforeRunning StartOutcomeKind = "stopped-before-running"
)

type StartOutcome struct { /* immutable */ }

func (r StartOutcome) Kind() StartOutcomeKind
func (r StartOutcome) Attempt() AttemptFact
func (r StartOutcome) PreparationFailure() (error, bool)
func (r StartOutcome) LaunchOutcome() (runtime.BootstrapOutcome, bool)

type StopOutcomeKind string

const (
    StopStopped StopOutcomeKind = "stopped"
    StopFailed  StopOutcomeKind = "stop-failed"
)

type StopOutcome struct { /* immutable */ }

func (r StopOutcome) Kind() StopOutcomeKind
func (r StopOutcome) Attempt() (AttemptFact, bool)
func (r StopOutcome) Failure() (error, bool)

type AttemptPhase string

const (
    AttemptPreparing  AttemptPhase = "preparing"
    AttemptLaunching  AttemptPhase = "launching"
    AttemptRunning    AttemptPhase = "running"
    AttemptStopping   AttemptPhase = "stopping"
    AttemptHistorical AttemptPhase = "historical"
)

type StopOrigin string

const (
    StopNotClaimed    StopOrigin = ""
    StopBeforeRunning StopOrigin = "before-running"
    StopAfterRunning  StopOrigin = "after-running"
)

type AttemptTerminalKind string

const (
    AttemptNotTerminal          AttemptTerminalKind = ""
    AttemptPreparationFailed    AttemptTerminalKind = "preparation-failed"
    AttemptLaunchFailed         AttemptTerminalKind = "launch-failed"
    AttemptStoppedBeforeRunning AttemptTerminalKind = "stopped-before-running"
    AttemptStopped              AttemptTerminalKind = "stopped"
    AttemptStopFailed           AttemptTerminalKind = "stop-failed"
)

type AttemptFact struct { /* immutable identities/facts only; no Host/raw causes */ }

func (f AttemptFact) WorkspaceID() uint64
func (f AttemptFact) ConfigurationID() uint64
func (f AttemptFact) ConfigurationVersionID() uint64
func (f AttemptFact) RuntimeInstanceID() runtimeconfigload.RuntimeInstanceID
func (f AttemptFact) LaunchAttemptID() runtimeconfigload.LaunchAttemptID
func (f AttemptFact) Phase() AttemptPhase
func (f AttemptFact) StopOrigin() StopOrigin
func (f AttemptFact) RunningPublished() bool
func (f AttemptFact) TerminalKind() AttemptTerminalKind

type Observation struct { /* immutable coherent snapshot */ }

func (s Observation) RuntimeInstanceID() runtimeconfigload.RuntimeInstanceID
func (s Observation) WorkspaceID() uint64
func (s Observation) ConfigurationID() uint64
func (s Observation) DesiredState() DesiredState
func (s Observation) ActualState() ActualState
func (s Observation) ActiveAttempt() (AttemptFact, bool)
func (s Observation) LastAttempt() (AttemptFact, bool)

type Owner struct { /* unexported synchronized state */ }

func NewOwner(
    workspaceID uint64,
    configurationID uint64,
    instanceID runtimeconfigload.RuntimeInstanceID,
    nextAttemptID LaunchAttemptIDSource,
    dependencies *runtime.DependencyBindings,
) (*Owner, error)
func (o *Owner) PrepareStart(
    request StartRequest,
) (LaunchPreparation, error)
func (o *Owner) Start(ctx context.Context, preparation LaunchPreparation, result PreparationResult) (StartOutcome, error)
func (o *Owner) Stop(ctx context.Context) (StopOutcome, error)
func (o *Owner) Observe() Observation
```

No exported Launcher interface, Loader adapter, Builder adapter, repository, or
generic manager is part of the package.

## 7. Sentinel Errors

The package exposes stable sentinels with these exact names:

```go
var (
    ErrInvalidOwner             error
    ErrInvalidStartRequest      error
    ErrAttemptIDSourceFailed    error
    ErrAttemptIDReused          error
    ErrStartConflict            error
    ErrPreparationNotOwned      error
    ErrInvalidPreparationResult error
)
```

Callers distinguish them with `errors.Is`. An attempt-ID source failure wraps
both `ErrAttemptIDSourceFailed` and the exact source error so that both remain
discoverable. Validation and conflict errors do not mutate lifecycle state.

## 8. Construction

`NewOwner` rejects with `ErrInvalidOwner`:

- zero Workspace ID;
- zero Configuration ID;
- empty Runtime Instance ID;
- nil `LaunchAttemptIDSource`;
- nil `DependencyBindings` pointer.

Owner borrows one stable bindings envelope for its lifetime and neither mutates
nor closes it. DP-009 Bootstrap retains typed-nil and semantic dependency
validation.

Initial desired and actual states are `DesiredStopped` and `ActualStopped`.
Construction performs no ID allocation, loading, building, launch, or resource
acquisition.

## 9. Start Request

`StartRequest` contains exactly Workspace, Configuration, and
ConfigurationVersion identities. It contains no Launch Attempt ID, Snapshot,
Host, Loader, or Builder capability.

`PrepareStart` rejects zero request identities or Workspace/Configuration
different from Owner with `ErrInvalidStartRequest`, without allocating or
mutating state.

The version identity is pinned into the attempt created by the Owner. A later
publication never redirects that preparation.

## 10. Attempt ID Allocation

Owner calls `LaunchAttemptIDSource` outside its state mutex. The source only
allocates an opaque candidate value; it does not create the Launch Attempt.

- Source error returns an error wrapping `ErrAttemptIDSourceFailed` and the
  exact source error.
- Empty candidate returns `ErrAttemptIDSourceFailed`.
- A candidate already committed during this Owner lifetime returns
  `ErrAttemptIDReused`.
- None of these outcomes creates an attempt or changes lifecycle state.

Concurrent `PrepareStart` calls may allocate multiple candidates outside the
mutex. Candidates that lose the later claim remain unused. A Launch Attempt
exists only when Owner commits one candidate at the claim linearization point.

## 11. PrepareStart Claim

After request validation and candidate allocation, `PrepareStart` locks Owner.
The claim is allowed only from:

- `ActualStopped`; or
- `ActualFailed` with no retained Host and no active attempt.

Any active preparation, launch, Host, stop, or retained failed-cleanup Host
returns `ErrStartConflict` without claim.

The claim linearization point is one locked mutation that:

1. creates the Launch Attempt owned by Owner;
2. commits the candidate ID to the lifetime used-ID set;
3. pins Workspace, Configuration, exact ConfigurationVersion, Runtime
   Instance, and Launch Attempt identities;
4. constructs the exact neutral `runtimeconfigload.LoadRequest`;
5. creates the Owner-owned preparation/start context;
6. publishes desired `DesiredRunning` and actual `ActualStarting`;
7. sets phase `AttemptPreparing`;
8. creates one opaque, non-forgeable `LaunchPreparation`.

`LoadRequest()` is the only Loader input produced by this contract.
`Context()` is read-only and permits external preparation work to observe Stop.

## 12. Launch Preparation Ownership

`LaunchPreparation` is an opaque token bound to exactly one Owner claim. Its
zero value, a token from another Owner, and a forged or stale token return
`ErrPreparationNotOwned`.

The token exposes no mutation, cancellation, Host, dependencies, or operation
record. It contains the exact `LoadRequest` and Owner-owned context through
read-only accessors.

One preparation accepts at most one result. Its stored result and operation
remain the convergence target for every later Start with the same authentic
token. The token is not a durable idempotency key and does not survive process
restart.

## 13. Closed Preparation Result

`PreparationResult` contains exactly one of:

- a complete Snapshot created with `PreparedSnapshot`; or
- an exact preparation error created with `FailedPreparation`.

`Kind()` reports `PreparationSnapshot` or `PreparationFailure`. The zero value,
a nil failure, an accessor combination inconsistent with the declared kind, or
any otherwise malformed union returns `ErrInvalidPreparationResult` while the
preparation remains unaccepted.

Before Start mutation, a prepared Snapshot must match all five identities
pinned in `LaunchPreparation.LoadRequest()`:

1. Workspace;
2. Configuration;
3. exact ConfigurationVersion;
4. Runtime Instance;
5. Launch Attempt.

A mismatch returns `ErrInvalidPreparationResult`, does not accept the
preparation, and performs zero Launcher calls. Loader and Builder semantics are
not repeated beyond this identity handoff check.

## 14. Start Acceptance

`Start` first authenticates the preparation token for this Owner and claim.
Zero, unknown, or other-Owner tokens return `ErrPreparationNotOwned`.

Immediately before acceptance or attachment, while holding the state mutex,
Start checks `ctx.Err()`:

- non-nil returns that exact context error without mutation, acceptance, or
  attachment;
- nil and the following acceptance or attachment under the same lock win the
  race with later cancellation.

For a live unaccepted preparation, the first valid result accepted under the
mutex wins:

- a failure result terminalizes preparation without a Launcher call;
- a Snapshot result moves `AttemptPreparing -> AttemptLaunching`, creates one
  tracked operation, and schedules exactly one Launcher call after unlock.

Owner accepts that exact Snapshot value or error interface identity as the sole
preparation result for the attempt. The tracked launch operation retains the
Snapshot only until Launcher returns; Owner does not retain a Snapshot in the
historical attempt. The exact preparation error identity remains retained in
the convergent outcome. Owner performs no structural, pointer, deep, text,
`errors.Is`, chain, comparability, or semantic-equivalence comparison between
results.

After first acceptance, every later Start with the same authentic token ignores
its supplied `PreparationResult` completely, including a zero or different
value, and attaches to or returns the stored operation or outcome. Later
arguments cannot alter the attempt and are intentionally not validated.

All runtime calls and waits occur outside the mutex.

## 15. Fixed Launcher Call

For an accepted Snapshot result, the production path is exactly:

```go
runtime.Launch(&runtime.BootstrapRequest{
    Snapshot:       snapshot,
    StartupContext: ownerPreparationContext,
    Dependencies:   owner.dependencies,
})
```

There is one and only one `runtime.Launch` call for the attempt. Owner never
calls `runtime.Bootstrap` or `Host.Start` directly.

A package-private immutable test seam may prove scheduling and results. It must
not become exported, mutable package state, a registry, or a production policy
boundary.

## 16. Preparation Failure

An accepted `FailedPreparation(cause)`:

- preserves the exact error identity and cause chain;
- performs zero Launcher calls;
- marks phase `AttemptHistorical`;
- sets terminal kind `AttemptPreparationFailed`;
- clears the active attempt;
- publishes actual `ActualFailed`;
- leaves desired `DesiredRunning`;
- completes Start with kind `StartPreparationFailed`.

The exact cause is available from the convergent `StartOutcome`, not
`Observation`.

If Stop already terminalized the preparation, later Start with the same token
returns the stored `StartStoppedBeforeRunning` outcome without accepting a new
failure or calling Launcher.

## 17. Launch Success and Failure

Launch failure without a Stop claim:

- preserves the exact `runtime.BootstrapOutcome`, failure identity, and cause
  chain;
- retains no Host;
- marks the attempt historical with `AttemptLaunchFailed`;
- publishes actual `ActualFailed`, while desired remains `DesiredRunning`;
- completes Start with `StartLaunchFailed`.

Launch success publishes the Host, phase `AttemptRunning`, actual
`ActualRunning`, `RunningPublished() == true`, and one immutable
`StartRunning` outcome only if Stop has not claimed the attempt and desired
remains `DesiredRunning`. DP-009 success already proves Host readiness. Once
published, the Running fact and stored Start outcome never regress.

## 18. Stop Claims by Phase

Stop always checks `ctx.Err()` under the mutex immediately before claim or
attachment. A non-nil error wins without mutation or attachment. A nil check
and the following locked mutation win over later caller cancellation.

| Attempt phase or state | Required Stop behavior |
| --- | --- |
| No active attempt, `ActualStopped` | Return idempotent `StopStopped`. |
| `AttemptPreparing` | Set `StopBeforeRunning`, keep Running unpublished, set desired Stopped, cancel Owner context after unlock, terminalize `AttemptStoppedBeforeRunning`, clear active attempt, publish ActualStopped. |
| `AttemptLaunching` | Set `StopBeforeRunning`, keep Running unpublished, set desired Stopped and ActualStopping, set phase AttemptStopping, cancel context after unlock, attach to the same combined operation. |
| `AttemptRunning` | Set `StopAfterRunning`, preserve Running published, set desired Stopped and ActualStopping, set phase AttemptStopping, create exactly one tracked Host Stop operation. |
| `AttemptStopping` | Attach to the existing operation; do not create another shutdown owner. |
| `ActualFailed` without Host | Publish ActualStopped immediately while preserving LastAttempt. |
| `ActualFailed` with retained Host | Return the stored `StopFailed`; do not retry cleanup. |

Concurrent operations follow mutex claim order. `StopOrigin()` and
`RunningPublished()` are recorded at Stop claim and never regress.

## 19. Stop During Preparation or Launch

Stop in `AttemptPreparing` completes without Host. A later Start with the same
preparation token converges on stored `StartStoppedBeforeRunning` and performs
zero Launch calls.

Stop in `AttemptLaunching` cancels the Owner-owned context outside the mutex.
If Launch later returns a Host:

- Running is never published;
- Owner retains the Host in `AttemptStopping`;
- `Host.Stop` is called exactly once.

If Launch returns failure, the exact Launch outcome is retained internally as a
secondary attempt fact, while the Start kind is
`StartStoppedBeforeRunning`. `LaunchOutcome()` remains false because the
primary Start kind is not `StartLaunchFailed`; no Host Stop is needed.

Whether late Host Stop succeeds or fails, a before-Running Start remains
`StartStoppedBeforeRunning`. Stop failure is reported only by `StopOutcome`.

## 20. Host Stop and Terminal Proof

`Host.Stop` receives the Owner-owned non-caller `context.Background()` and runs
outside the mutex. DP-010 defines no timeout, force, or retry policy.

While it blocks, actual state remains `ActualStopping`.

Nil is the only proof that permits Owner to:

- clear Host and active-attempt references;
- mark the attempt historical with `AttemptStopped`;
- publish actual `ActualStopped`;
- complete Stop with `StopStopped`.

A non-nil error cannot prove resource absence:

- actual becomes `ActualFailed`;
- desired remains `DesiredStopped`;
- the retained active attempt remains in `AttemptStopping` with terminal kind
  `AttemptStopFailed`;
- exact error identity and cause chain are retained;
- Host and attempt references remain retained;
- Stop completes with `StopFailed`;
- repeated Stop returns the stored failure without retry;
- new `PrepareStart` returns `ErrStartConflict`.

For `StopBeforeRunning`, Start remains `StartStoppedBeforeRunning` after either
Stop success or failure. For `StopAfterRunning`, the already stored
`StartRunning` remains unchanged during Stopping and after either Stop result.
Stop never retroactively changes whether Running was published.

## 21. Same-Token Convergence

For the exact preparation token:

- in `AttemptPreparing`, the first valid Start result may be accepted;
- after first result acceptance, every repeated Start ignores its supplied
  result without comparing or validating it;
- in `AttemptLaunching`, repeated Start attaches to the one launch operation;
- in `AttemptRunning`, repeated Start returns the stored `StartRunning`;
- in `AttemptStopping` with `StopBeforeRunning`, repeated Start attaches to the
  combined operation and returns `StartStoppedBeforeRunning` after either Stop
  result;
- in `AttemptStopping` with `StopAfterRunning`, repeated Start returns the
  immutable stored `StartRunning` during and after Stop;
- historical preparation or launch failure returns its stored exact
  `StartOutcome`;
- a preparation stopped before Start returns stored
  `StartStoppedBeforeRunning`.

A foreign or forged token returns `ErrPreparationNotOwned`. No equality or
equivalence comparison is performed on later Snapshots or errors. Convergence
never invokes Launcher or Host Stop twice.

## 22. Caller Cancellation

`PrepareStart` has no caller context.

For Start and Stop, the locked `ctx.Err()` check immediately before
acceptance, claim, or attachment defines the race:

- cancellation already visible at that check returns the context error and
  changes nothing;
- a nil check plus the immediately following locked mutation wins;
- cancellation after that point interrupts only this caller's wait.

Owner-owned tracked work, context, and cleanup duties continue after a waiter
returns. Callers can repeat with the same token or use `Observe` to see truth.
Caller cancellation is not a terminal lifecycle outcome.

## 23. Outcomes and Access

`StartOutcome` is immutable and has exactly one declared
`StartOutcomeKind`. It always exposes `AttemptFact`.

- `PreparationFailure()` succeeds only for
  `StartPreparationFailed`.
- `LaunchOutcome()` succeeds only for `StartLaunchFailed`.
- `StartRunning` and `StartStoppedBeforeRunning` expose neither failure
  accessor.
- Stop failure is never duplicated as a raw Start failure.

`StopOutcome` is immutable and has exactly one declared `StopOutcomeKind`.
`Attempt()` is absent only for an idempotent Stop with no applicable attempt.
`Failure()` succeeds only for `StopFailed` and returns the exact Host Stop
error.

Method-level errors represent validation, conflict, ownership, ID-source, or
this caller's wait. They do not replace accepted lifecycle outcomes.

## 24. Observation

`Observe` returns one detached immutable coherent value built under the state
mutex.

It exposes Owner Workspace, Configuration, and Runtime Instance identity;
desired and actual state; optional active `AttemptFact`; and optional last
`AttemptFact`. Attempt facts include immutable Stop origin, whether Running was
ever published, and terminal category.

`Observation` and `AttemptFact` expose no Host, dependency, context,
cancellation function, mutable operation, or raw cause. Exact failures are
available only through convergent operation outcomes until diagnostics and
redaction have a separate contract.

## 25. Owned State and Locking

One short-held mutex protects:

- bound identities and desired/actual state;
- used attempt IDs;
- active and last attempt facts;
- the live opaque preparation token and consumption fact;
- tracked launch and stop operations;
- Owner-owned preparation context and cancellation;
- active or conservatively retained Host;
- exact terminal outcomes used for convergence.

The mutex is never held during ID-source calls, Loader/Builder work,
`runtime.Launch`, `Host.Stop`, resource/channel waits, context waits, or caller
waits.

## 26. Ownership

| Value or resource | Ownership |
| --- | --- |
| Workspace, Configuration, Runtime Instance scope | Bound immutably to Owner at construction. |
| Launch Attempt identity and record | Candidate allocated by injected source; attempt created and owned by Owner at claim. |
| Exact ConfigurationVersion pin | Selected by StartRequest and pinned by Owner before Loader/Builder. |
| LoadRequest | Constructed by Owner and exposed read-only through preparation. |
| Snapshot or preparation failure | Produced externally and returned through closed PreparationResult. Snapshot is retained only by the tracked launch operation until Launcher returns; exact failure is retained in the convergent outcome. |
| Dependency bindings | Externally owned; borrowed stable and unchanged by Owner. |
| Preparation/start context | Created, canceled, and owned by Owner. |
| Runtime Launcher | Stateless and owns no lifecycle state. |
| Runtime Host reference | Owned by Owner after Launch until cleanup is proven. |
| Runtime graph and resources | Owned exclusively by Host under ARCH-002. |

## 27. Truthfulness Limitation

Current Host has no completion or supervision signal for unexpected
termination after Running publication. Owner must not infer `Running ->
Failed` by polling `Running()` or guessing from unrelated facts.

Unexpected Runtime termination remains integration-gated until a Host-owned
terminal signal exists.

## 28. Local Acceptance Proofs

The isolated implementation must prove:

1. constructor, request, ID-source, and identity validation;
2. Owner creates the attempt and pins exact version before external
   preparation;
3. `LoadRequest` contains the five exact identities;
4. concurrent preparation commits at most one claim;
5. losing ID candidates never become attempts;
6. foreign, stale, and reused preparations use the declared errors;
7. invalid or mismatched preparation performs zero Launch calls and remains
   unaccepted;
8. the first valid concurrent result wins, later same-token arguments are
   ignored, and non-comparable Snapshots or errors trigger no equality call or
   panic;
9. exact preparation and Launch failures converge unchanged;
10. accepted Snapshot causes exactly one Launcher call, and Owner releases its
    Snapshot copy after Launcher returns;
11. Running publishes only after successful ready Host return;
12. Stop works in Preparing, Launching, Running, Stopping, and both Failed
    forms;
13. before-Running Stop preserves `StartStoppedBeforeRunning` on Stop success
    and failure;
14. after-Running Stop preserves the exact stored `StartRunning` during and
    after Stop success or failure;
15. same-token Start converges in every phase without validating later result
    arguments;
16. concurrent Stop causes one Host Stop and one exact result;
17. the locked `ctx.Err()` race has the specified winner semantics;
18. caller cancellation after claim affects only waiting;
19. Stop failure retains Host and blocks new preparation;
20. no resource call or wait occurs under the mutex;
21. Observation is coherent, capability-safe, and reports immutable Stop
    origin and Running publication;
22. separate Owners progress independently;
23. package race tests and relevant Runtime regression tests pass when the
    toolchain supports them.

## 29. Future Integration Proofs

The following remain future production evidence:

- Loader uses exactly `LaunchPreparation.LoadRequest`;
- Builder returns Snapshot provenance matching the same preparation;
- every production path is Owner-to-`runtime.Launch`, with no Bootstrap or
  `Host.Start` bypass;
- exactly one Owner scope is routed per Runtime Instance;
- management authorization occurs before mutation;
- dependency composition is stable and has no hidden registry;
- durable allocation, history, idempotency, and recovery survive restart;
- Host terminal supervision enables truthful unexpected failure observation.

DP-009 AP-003 and AP-011 remain integration-gated.

## 30. Explicit Deferrals

DP-010 does not define:

- actual Loader/Builder adapter or production wiring;
- persistence schema, transaction, or durable history;
- HTTP API, authorization, command DTO, or durable idempotency;
- concrete distributed or persistent ID-allocation strategy;
- retry, restart, replacement, rollback, recovery, reconciliation, or reload;
- PID, worker, process supervision, scheduling, or clustering;
- diagnostics, logging, redaction, metrics, or alerting;
- Stop timeout, force, or cleanup-retry policy;
- Host API changes or unexpected-termination polling;
- generic manager, registry, service locator, or policy framework.

## 31. Implementation Boundary

The implemented first code slice adds only `internal/runtimelifecycle` and
local proof tests for this contract. It uses fakes around the package-private
immutable launch seam and external preparation boundary.

It does not wire Loader, Builder, Control Service HTTP handlers, repositories,
persistence, or production routing. Implementation does not promote the design
status.

## 32. Summary

Runtime Lifecycle Owner first creates and pins one Launch Attempt through
`PrepareStart`, exposing the exact neutral LoadRequest required by Loader and
Builder. `Start` then accepts one closed preparation result and invokes only
the stateless Runtime Launcher.

One short-held state boundary preserves per-Instance serialization,
same-token convergence, caller-cancellation rules, conservative Host
ownership, and truthful operational state without prematurely designing
persistence, management, policy, supervision, or production wiring.

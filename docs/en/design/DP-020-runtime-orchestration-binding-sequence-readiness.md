# DP-020: Runtime Orchestration Binding Sequence Readiness

[Russian version](../../ru/design/DP-020-runtime-orchestration-binding-sequence-readiness.md)

## 1. Status

- **Design Status:** Draft
- **Implementation Status:** Planned overall; Slice 3 implemented and
  independently accepted in isolation

Implementation progress: TASK-031 and TASK-032 produced Coordinator-Accepted
isolated partial implementations of Slices 1 and 2, and TASK-034 defined their
required conformance repair. TASK-035 implements Slice 2R in isolation: the
six-field authorization and complete binding now live in the dependency-leaf
package, primitive managed claims use the sole `ExecuteManagedStart` adapter,
and command-owned rendezvous identities are unique and callback-scoped.
TASK-043 implements the concrete DP-013 composition-private invoker in
isolation; production wiring remains absent. TASK-036 resolves the remaining Slice-3 command-gate and continuation
API ambiguity, and TASK-037 implements that Slice 3 protocol in isolation:
managed primitive and linked gates, the stateless OwnerClaim-to-DP-014
continuation, exact revision threading and managed Flow outcome adaptation.
TASK-037 is independently accepted. Slice 4 is completed and Coordinator
Accepted as TASK-038 with verdict `TASK-026 REMAINS BLOCKED`; its first
design-only next candidate completed as Coordinator-Accepted TASK-039.
Completed and Coordinator-Accepted TASK-040 implements and verifies those
accepted Draft DP-010 semantics in isolation; repeat final Reviewer is
`APPROVED` 0/0. The overall status remains Planned.

This focused design decomposes the remaining Approved DP-019 prerequisites —
exact orchestration authorization, private managed invocation, and
OwnerClaim-to-DP-014 binding — into ordered, independently testable
implementation slices and discharges the deferred design decisions they need.

It defines no production code and changes no approved semantics. TASK-026
remains `Blocked by Architecture`.

## 2. Purpose

Approved DP-019 fixes the ordering, ownership, concurrency, permit, and
fail-closed invariants for activation orchestration prerequisites but describes
the authorization tuple, the private managed invoker, the managed Flow seam,
and the binding sequence only conceptually. This proposal records the focused,
implementable decomposition of exactly those seams so each can be implemented
and independently accepted before the DP-016 orchestrator is reconsidered.

## 3. Authority

This proposal refines, without overriding:

- [ARCH-004](../architecture/ARCH-004-runtime-deployment-and-identity-model.md),
  section 19;
- [DP-010](DP-010-runtime-lifecycle-owner-contract.md) — the Owner remains the
  sole lifecycle and attempt-claim authority;
- [DP-011](DP-011-runtime-launch-pipeline-integration.md) — the synchronous
  `PrepareStart -> Load -> Build -> Start` path and the private continuation
  location;
- [DP-013](DP-013-runtime-management-routing.md) — exact routing,
  authorization-before-mutation, and the planned private seam;
- [DP-014](DP-014-runtime-operational-identity-persistence.md) — conditional
  aggregate revision and the Launch Attempt membership / execution-generation
  binding boundary;
- [DP-015](DP-015-runtime-management-command-idempotency.md) — the claim /
  replay / permit / unresolved-barrier boundary, including the TASK-028 parent
  /phase core and the TASK-029 Continue gate and pending-Stop rendezvous;
- [DP-016](DP-016-runtime-activation-replacement-rollback.md) — the unchanged
  activation / replacement / rollback ordering and acceptance proofs;
- [DP-017](DP-017-runtime-recovery-reconciliation.md) — fail-closed recovery;
- [DP-019](DP-019-runtime-activation-orchestration-prerequisites.md) — the
  Approved prerequisite contract this proposal decomposes.

Accepted ADRs and Active or Frozen architecture remain authoritative. This
Draft never overrides an Approved document, and it does not raise or lower any
existing Design or Implementation Status.

## 4. Scope

This design covers only the focused decomposition of the conceptual seams of
Approved DP-019 §7–§8 and §14–§16:

- the policy-neutral orchestration authorizer representation and ownership, the
  validated authorization request value, the failed-authorization error
  contract, and the DP-013 authorization-before-mutation adaptation;
- the private managed invoker versus managed Flow seam package split and
  invocation direction, the managed construction and per-call managed Start
  surfaces, the immutable per-invocation lifecycle of `StartExecutionBinding`,
  and the exact failed private-invocation error contract;
- the immutable `OwnerClaimView` tuple and the OwnerClaim-to-DP-014 binding
  sequence: composition inputs, the closed four-outcome continuation result,
  conditional attempt-membership publication and same-generation binding order
  against the existing pending-Stop rendezvous, and the final
  Stop-versus-Continue gate placement;
- the provable activation-adaptation invariant;
- the lock / ownership proof plan required before implementation;
- the ordered list of bounded implementation slices and their dependency
  order.

## 5. Non-goals

This design does not define:

- production implementation of any slice, a public HTTP/CLI/DTO/API route, or
  wiring;
- a concrete authorization policy, Principal model, credential mapping, or
  transport;
- a storage schema, migration, external persistence, cross-process lease, or
  recovery execution;
- the DP-016 orchestrator implementation, DP-017 recovery implementation,
  DP-018 reporting implementation, automatic rollback/restart, supervision,
  scheduling, or Production Activation;
- any change to the DP-010 lifecycle contract, the DP-011 base Flow behavior,
  the DP-013/DP-015 public surfaces, or the approved semantics of
  DP-014/DP-015/DP-016/DP-017/DP-019;
- any new Runtime Instance domain or deployment domain concept.

## 6. Decision Summary

The remaining DP-019 prerequisites are recordable as three original ordered
implementation slices, one conformance-repair slice inserted before Slice 3,
and one gated re-assessment:

1. the exact orchestration authorizer surface at the existing DP-013 / DP-019
   command boundary;
2. the private managed invoker and the managed Flow per-call seam with the
   per-invocation `StartExecutionBinding`;
3. Slice 2R, the authoritative binding and primitive managed-adapter
   conformance repair;
4. the OwnerClaimView-to-DP-014 conditional attempt publication and
   same-generation binding before Load, integrated with the existing
   pending-Stop rendezvous;
5. re-assessment of DP-016 orchestrator readiness after Slices 1, 2, 2R, and 3
   are implemented and independently accepted.

No slice bypasses another. Slice 1 may stand independently; Slice 2 depends on
Slice 1; Slice 2R repairs the accepted partial Slices 1 and 2; Slice 3 depends
on Slice 2R; Slice 4 depends on independent acceptance of Slice 3.

## 7. Deferred Decision: Orchestration Authorizer and Request Representations

### 7.1 Authorization request value

One immutable, validated value is the only authorization input:

```text
OrchestrationAuthorizationRequest {
    OperationalDomain           string
    WorkspaceID                  uint64
    ConfigurationID              uint64
    RuntimeInstanceID            runtimeconfigload.RuntimeInstanceID
    Action                       OrchestrationAction
    TargetConfigurationVersionID uint64
}
```

Validation requires every identity to be non-zero and exact and the target
version to be a valid Published identity belonging to that Configuration.
Validation failure, authorization denial/failure/panic, absent scope, or
pre-claim cancellation performs zero command, aggregate, and lifecycle
mutation.

### 7.2 Retaining the `OperationalDomain` field

Approved DP-019 §7 requires `OperationalDomain`; this Draft cannot remove it.
The value is the exact non-empty opaque domain already present in the accepted
DP-015 command `Scope`. It is carried into every initial and replay
authorization request and into the same per-invocation binding. It is not
inferred from Workspace, Configuration, Runtime Instance, process locality, or
a default constant. `runtimemanagement.Target` remains unchanged; a private
managed invoker receives its domain as a separate immutable composition input
and validates it together with that Target.

### 7.3 Authorizer representation

The seam is one policy-neutral synchronous named function type owned by the
future composition-facing orchestration package (not stored in
`runtimemanagement.Directory` and not replacing the existing DP-015
`Authorize(...) error` parameter). It is validated per submission, never
cached, and re-evaluated on every initial and replay submission.

A conforming authorizer returns no error or one exact error. Denial,
unavailability, invalid scope, panic, and cancellation each produce a
truthful, specific error outcome and no mutation. The DP-013 existing exported
authorization surface for `Directory.Start/Stop/Observe` remains unchanged;
orchestration actions never pass through that surface without the additional
exact orchestration authorization.

## 8. Resolved Design Boundary: Private Managed Invoker and Managed Flow Seam

TASK-042 closes only the remaining concrete-invoker design ambiguity through
Draft/Partial [DP-021](DP-021-private-exact-scope-managed-start-invoker.md).
The following existing decomposition remains authoritative context: DP-021
fixes `runtimemanagement` ownership, preconstructed-Flow custody, the sole
`InvokeManagedStart` operation, cancellation delegation, capability custody,
failure behavior, and absence of legacy fallback. A future TASK-026
orchestrator-owned DP-015 callback closure calls that invoker as its sole
lifecycle subcall and owns `TerminalOutcome` mapping, publication, and
terminalization outside DP-021; the invoker is not itself the callback.
TASK-043 implements only that invoker in isolation; no callback, terminal work,
or orchestrator is activated.

### 8.1 Package split and invocation direction

- The private managed invoker lives at the DP-013 composition side and
  validates the binding against its immutable scope and the `StartRequest`
  before calling the stored Flow.
- The authoritative immutable cross-layer values live in dependency-leaf
  `internal/runtimeorchestrationbinding`. It may import `runtimeconfigload`
  identity types, but imports neither `runtimecommandidempotency`,
  `runtimeidentity`, `runtimemanagement`, nor `runtimelaunchflow`.
- The managed Flow seam lives in `internal/runtimelaunchflow`. The existing
  unmanaged `New(owner, loader)` constructor and its `Start(ctx, request)`
  semantics remain unchanged.
- The invocation is one synchronous per-invocation call
  `Flow.StartManaged(context, StartRequest, StartExecutionBinding)`.
- The four higher packages may depend on the neutral binding package; the
  neutral package depends on none of them. `runtimelaunchflow` imports neither
  `internal/runtimecommandidempotency` nor `internal/runtimeidentity`.

#### 8.1.1 Primitive managed adapter seam

`runtimecommandidempotency.Boundary` owns one additive internal-repository
primitive adapter, conceptually:

```text
ExecuteManagedStart(
    context,
    Start Scope,
    CommandKey,
    Start Intent,
    ExpectedAggregateRevision,
    ExecutionGeneration,
    AuthorizeOrchestration,
    invoke(StartExecutionBinding) -> TerminalOutcome, error,
) -> Admission, error
```

It accepts only an exact primitive Start scope/intent. It validates every
input, derives the six-field `ActivateExactTarget` authorization request, runs
authorization and the pre-claim cancellation gate, and only then enters the
same command admission linearization point used by `Execute`. Authorization is
performed on every initial, in-progress, and replay submission before command
inspection; denial, panic, cancellation, or invalid input performs zero
mutation.

Only a newly committed primitive claim receives the callback. In the same
locked claim transaction, before releasing admission locks, Boundary commits
the command record and live permit, allocates a unique generation-bound
`StartRendezvous` identity, installs its private lookup entry, and determines
the complete primitive `StartExecutionBinding`. All caller-supplied binding
facts are validated before claim, so deterministic binding construction cannot
fail after command mutation. The callback then runs once, synchronously and
outside all command locks, with no parent/phase identity.

In-progress and replay submissions return their normal Admission and receive
no callback, binding, rendezvous allocation, or permit. A record originally
claimed through legacy `Execute` is likewise only observed; the managed seam
never adopts or recreates its execution authority.

The permit, rendezvous lookup, and callback authority expire when the
callback/permit returns, panics, executes `runtime.Goexit`, or loses its
Boundary generation. This does not mutate or invalidate the structurally valid
binding value; its non-reuse depends on callback custody and absence of bypass.
A valid terminal outcome is published by the existing primitive permit rules.
A callback error, panic, invalid outcome, missing terminal publication, or
indeterminate return leaves the command Claimed/unresolved, expires live
authority, blocks rendezvous resolution, and returns the existing truthful
indeterminate error; no callback is retried or detached.

Existing `Boundary.Execute` remains unchanged as an isolated legacy primitive
surface for current non-orchestration callers and tests. Production activation
orchestration must use `ExecuteManagedStart`; it may not call `Execute` and
then synthesize a binding, invoke the private managed invoker directly, or
fall back to legacy execution after any managed failure. The DP-013 public
`Directory.Start/Stop/Observe` and existing DP-015 surfaces remain unchanged;
the new method is repository-internal and creates no transport/API path.

#### 8.1.2 Managed parent and StartTarget adapter seam

Replacement and rollback use one additive managed parent path. Conceptually,
the surfaces are:

```text
Boundary.ExecuteManagedParent(
    context, Replace|Rollback Scope, CommandKey, matching Intent,
    AuthorizeOrchestration,
    invoke(*ManagedParentExecution) error,
) -> ParentAdmission, error

ManagedParentExecution.ContinueOrExecuteManagedStartTarget(
    context, ExpectedAggregateRevision, ExecutionGeneration,
    invoke(StartExecutionBinding) -> TerminalOutcome, error,
) -> PhaseAdmission, prevented bool, error
```

`ExecuteManagedParent` accepts only exact Replace or Rollback,
uses `AuthorizeOrchestration` on every initial, in-progress, and replay
submission, and is the only source of a `ManagedParentExecution`. A parent
admitted through legacy `ExecuteParent` cannot be adopted or upgraded.

`ManagedParentExecution.ContinueOrExecuteManagedStartTarget(...)` preserves the
existing pre-phase Continue ordering. Only a newly committed `StartTarget`
phase receives a callback. In the phase-claim transaction it derives the exact
parent identity and ordinal-one StartTarget identity, commits the phase and its
permit, allocates and indexes a unique generation-bound rendezvous identity,
constructs the complete linked `StartExecutionBinding`, and then invokes the
callback once outside command locks. In-progress and replay observations
receive no binding or execution authority.

The existing `ExecuteParent`, `ParentExecution`, and
`ContinueOrExecuteStartTarget` remain unchanged compatibility surfaces. The
managed path does not synthesize authority from their records.

### 8.2 Immutable per-invocation `StartExecutionBinding`

`StartExecutionBinding` is the single authoritative immutable value owned by
`runtimeorchestrationbinding`, constructed by the permit-holding stack and
passed inward. It contains the validated six-field
`OrchestrationAuthorizationRequest`, the expected aggregate revision, the
composition-owned exact `ExecutionGeneration`, the parent and phase identity
when applicable, and the opaque `StartRendezvous` for this live primitive or
phase execution. It contains no primitive, parent, phase, or Stop permit, no
preparation token, no Host or Snapshot, no context cancellation authority, and
no mutable Owner state. It is validated before any Owner mutation, retained
only on that synchronous call stack and invoked at most once; it is never
stored as a Flow field. Callback return expires live authority without mutating
or invalidating the structurally valid binding value.

Linked execution identity is an explicit all-or-none variant: primitive Start
has neither parent nor phase; Replace/Rollback `StartTarget` has both the exact
parent command identity and its command-boundary-derived `StartTarget`, ordinal
one phase identity. A lone parent, lone phase, caller-selected phase, or an
action/variant mismatch is invalid. Store-owned `Revision` and
`ExecutionGeneration` remain authoritative persistence concepts; the leaf
package carries validated lossless values converted explicitly at the
`runtimeidentity` boundary.

The managed construction binds the existing stateless private
`StartClaimContinuation` exactly once. The binding creates no Registry entry,
mutable slot, goroutine, detached callback, or new lifecycle state.

### 8.3 Opaque `StartRendezvous` cross-package mechanism

The existing pending-Stop rendezvous primitive in
`internal/runtimecommandidempotency` remains the sole owner of its signals and
its locks, per DP-019 §13 and §17. The cross-package seam is an opaque identity
value with no signaling or waiting methods, defined in
`runtimeorchestrationbinding`, so:

- the DP-015 command boundary allocates one collision-safe identity for each
  live primitive Start or `StartTarget` execution and indexes its private
  concrete rendezvous by that identity plus the active Boundary generation;
- it exposes only the opaque identity in the binding;
- `StartExecutionBinding` holds the opaque handle;
- `runtimemanagement` and `runtimelaunchflow` pass it through without
  importing the command-boundary package.

The handle carries no pointer, channel, function, capability, permit, or
mutable state. Only `runtimecommandidempotency` resolves it, and only while the
original callback-scoped execution and Boundary generation are live. Missing,
forged, reused, cross-generation, or identity-mismatched handles yield
`Blocked`; they never mean no pending Stop. Callback return and Boundary
replacement expire resolution authority but do not erase durable unresolved
facts. No global rendezvous registry exists outside the command boundary.

The command boundary exposes only three repository-internal operations over a
complete binding; their result types are closed, not booleans:

```text
ResolveManagedStartEarly(binding, launchAttemptID)
    -> GateClear | GateStopConverged | GateBlocked

ResolveManagedStartFinal(binding, FinalContinue | FinalBindingFailed)
    -> FinalContinue | FinalBindingFailed | GateStopConverged | GateBlocked

SignalManagedStartNoClaim(binding, Cancelled | Rejected | Failed)
```

`ManagedStartGateOutcome`, `ManagedStartFinalDisposition`, and their constants
belong to `runtimecommandidempotency`. The neutral immutable
`StartNoClaimCause` and its three constants belong to
`runtimeorchestrationbinding`; the existing command-package names may remain
aliases for compatibility. `StartClaimOutcome` and its four constants belong
to `runtimelaunchflow`. This placement avoids every reverse import.

Each operation resolves the opaque handle together with the complete
authorization tuple, primitive-or-linked identity, exact command or phase,
live permit, and active Boundary generation. The caller supplies no command
key: the private lookup entry owns the exact command identity. Missing, forged,
reused, mismatched, cross-generation, or expired handles return `GateBlocked`
with the existing indeterminate-execution error; they never mean clear.

One managed primitive or StartTarget rendezvous starts in `PreOwner`. A Stop
admitted there occupies the single tracked-Start exception on its original
`Boundary.Execute` stack and waits. The early operation stores the exact Owner
attempt and moves to `Binding`; an already pending Stop is signalled and only
its original claimant may invoke Stop and publish its result. Exact convergence
returns `GateStopConverged`; cancellation before delegation clears the gate;
lost or conflicting evidence blocks it. During `Binding`, one later Stop may
claim and wait. The final operation is the single admission linearization point
between that Stop and `FinalContinue` or `FinalBindingFailed`. Stop-first must
converge on its original stack; disposition-first seals the result. After
Continue, a later Stop uses the ordinary tracked-Start behavior. No second Stop
or unrelated lifecycle mutation bypasses either gate.

Callback return, panic, `runtime.Goexit`, or Boundary generation replacement
expires the handle and wakes every waiter as Blocked while durable command or
phase facts remain Claimed. Command locks are held only to validate, transition,
and capture a notification; no command lock is held while waiting or calling
identity, continuation, Flow, Owner, or external work.

### 8.4 Failed private-invocation error contract

Binding validation failure before Owner mutation returns an exact,
distinguishable error (fail-closed) and performs zero Owner or aggregate
mutation. Owner or continuation errors are returned unchanged and unwrapped;
there is no reclassification, wrapping, or recovery at this seam.

## 9. Deferred Decision: OwnerClaimView, Binding Sequence, and Outcomes

### 9.1 `OwnerClaimView` and the `LaunchAttemptID` accessor

Immediately after a successful authentic `Owner.PrepareStart` and before any
Load, Build, or Launcher work, the Flow assembles one immutable
`OwnerClaimView`:

```text
OwnerClaimView {
    WorkspaceID                  uint64
    ConfigurationID              uint64
    RuntimeInstanceID            runtimeconfigload.RuntimeInstanceID
    LaunchAttemptID              runtimeconfigload.LaunchAttemptID
    TargetConfigurationVersionID uint64
}
```

The LaunchAttemptID accessor evidence is present today: the authentic
`LaunchPreparation.LoadRequest()` returns `runtimeconfigload.LoadRequest`,
which carries `LaunchAttemptID()`. The claim view is sourced from that
Owner-issued `LoadRequest`; no `runtimelifecycle.StartRequest` shape change is
needed.

### 9.2 Closed continuation contract and outcome

`runtimelaunchflow` owns the interface and lifecycle adaptation so it keeps no
dependency on command or identity packages:

```text
StartNoClaim(context, StartExecutionBinding, StartNoClaimCause) error
AfterOwnerClaim(context, StartExecutionBinding, OwnerClaimView)
    -> StartClaimOutcome, error
```

The new focused `internal/runtimeorchestrationcontinuation` package owns the
long-lived stateless implementation. It depends on
`runtimeorchestrationbinding`, `runtimecommandidempotency`, `runtimeidentity`,
and `runtimelaunchflow`; none depends back on it. It retains command-boundary
and narrow identity-boundary dependencies only, never a per-call binding,
Owner preparation, permit, mutable rendezvous, goroutine, or registry entry.

`AfterOwnerClaim` returns exactly one closed outcome:

- `Continue`: the already-admitted pending-Stop rendezvous is definitively
  absent; the exact durable attempt membership and the same-generation binding
  are committed at their expected revisions; the final Stop-versus-Continue
  gate released Flow.
- `StopConverged`: the original pending Stop claimant invoked the exact Stop
  and converged the claimed attempt; Flow begins no Load and performs no DP-014
  publication or binding.
- `BindingFailed`: coherent exact evidence proves that attempt-membership
  publication or generation binding did not commit and no conflict exists and
  no external preparation began; Flow submits the failure through the authentic
  Owner preparation.
- `Blocked`: permit or signal loss, stale or conflicting revision, different
  generation, unavailable or unknown facts, unproven Stop convergence, or an
  indeterminate publication; Flow begins no Load and the linked set remains
  unresolved.

`Continue` and `StopConverged` require a nil error. `BindingFailed` and
`Blocked` require one exact non-nil cause. An invalid combination or panic is
`Blocked` with the managed continuation validation error.

Caller cancellation is observed only at the existing DP-011 gate before the
Owner claim. After successful `PrepareStart`, Flow derives a non-cancelable
continuation context (equivalent to `context.WithoutCancel(ctx)`, preserving
values but not cancellation or deadline) and uses it for `AfterOwnerClaim`.
Every same-preparation Owner convergence call uses a local non-cancelable
context as well. `StartNoClaim` is also signalled with a local non-cancelable
context after a definitive pre-claim outcome, so caller cancellation cannot
erase the command signal. Dependency failure or unavailable identity/gate
state remains Blocked; it is not reclassified as caller cancellation.

Flow maps the outcomes without letting the continuation receive its
preparation token: Continue calls the existing prepared-start path;
StopConverged calls `Owner.Start` with the same authentic preparation and an
empty result, retrieving the stored `StartStoppedBeforeRunning` outcome without
Load; BindingFailed passes `FailedPreparation(cause)` through the authentic
Owner preparation and returns the semantic Owner outcome; Blocked also
converges the in-memory Owner claim locally to avoid a leaked preparation but
returns the exact Blocked cause so the command or phase remains unresolved.
Owner convergence failure supersedes the continuation cause. A definitive
pre-Owner cancellation, rejection, or failure calls `StartNoClaim`; a signal
failure makes the execution Blocked.

The continuation never publishes a lifecycle, phase, or parent terminal
outcome itself; only the exact Owner result followed by the DP-014 and DP-015
conditional publications may terminalize them.

### 9.3 Binding-sequence order and revision threading

The fixed order, after the sole Owner claim and before Load, is:

1. resolve the already-admitted pending-Stop rendezvous through the original
   Stop claimant only; `StopConverged` exits without either write, and
   `Blocked` leaves the linked set unresolved;
2. only after definitive proof that no pending Stop remains, read the exact
   Runtime Instance and prove workspace, configuration, instance identity, and
   expected revision before any mutation; read failure or any mismatch is
   `Blocked` with zero writes;
3. after that pre-mutation proof, conditionally
   publish the exact Launch Attempt membership and version pin in DP-014 at the
   expected aggregate revision;
4. read the committed revision returned by that write and conditionally bind
   the exact active attempt to the composition-owned execution generation at
   that new expected revision;
5. after any indeterminate result, inspect exact aggregate facts by
   `ReadRuntimeInstance` and `ReadLaunchAttemptHistory` and converge to an
   exact existing terminal outcome or return `Blocked`;
6. execute the final Stop-versus-Continue gate for a Stop admitted after the
   early rendezvous check, then release `Continue` only under confirmed, exact
   same-generation binding.

The continuation depends on a narrow `IdentityStore` interface matching the
existing Store operations so unavailable and indeterminate behavior is
testable; `runtimeidentity.Store` itself remains unchanged. Conversion at this
boundary is explicit and lossless:
`runtimeidentity.Revision(uint64(binding.ExpectedAggregateRevision()))` and
`runtimeidentity.ExecutionGeneration(string(binding.ExecutionGeneration()))`.
Zero or round-trip mismatch is Blocked; no value is allocated or inferred.

The revision returned by a committed attempt claim is the only expected
revision accepted by the generation bind. No stale write is retried. After any
error or non-commit, exact inspection uses a revision sandwich:
`ReadRuntimeInstance A -> ReadLaunchAttemptHistory -> ReadRuntimeInstance B`.
The pre-mutation scope/revision read is not sandwich observation A; inspection
starts with a fresh read after the ambiguous operation.
The result is coherent only when A and B have equal revisions and identical
immutable identity and active-attempt facts. Exact active attempt, version,
Claimed phase, and generation may prove satisfied convergence. A coherent
absence at the still-current relevant expected revision, with no conflicting
history, active attempt, or generation and no external preparation, may yield
BindingFailed. Stale revision, changed sandwich, different or reused attempt,
different version or generation, inactive or terminal attempt, read failure,
unavailability, or unknown facts is Blocked. Neither outcome permits a blind
retry or repair.

The interface contains exactly the operations needed by this sequence:
`ConditionalClaimLaunchAttempt`, `ConditionalBindExecutionGeneration`,
`ReadRuntimeInstance`, and `ReadLaunchAttemptHistory`, with the existing
`runtimeidentity` parameter and result types. It has no terminal-publication,
recovery, allocation, or mutation method beyond claim and bind.

Same-attempt/same-version membership and same-generation binding already
present are zero-mutation satisfied convergence observations. Different
attempt, different version, different generation, stale revision, inactive
fact, or unknown state is never auto-repaired or auto-replaced.

## 10. Activation Adaptation Invariant

Initial activation preserves the primitive Start identity and intent:

- orchestration uses `Boundary.ExecuteManagedStart` as its sole primitive
  managed admission; legacy `Boundary.Execute` remains unchanged only for
  isolated compatibility callers and cannot be adopted by orchestration;
- `runtimelifecycle.StartRequest` remains the immutable existing shape
  `(WorkspaceID, ConfigurationID, ConfigurationVersionID)` and carries no
  parent, phase, or authorization authority;
- the composition-facing authorization adapter labels the exact Start
  submission authorization action as `ActivateExactTarget`;
- no one-phase parent is created for API uniformity; parent shapes are used
  only for `Replace` and `Rollback`.

Replacement and rollback map to `ReplaceWithExactTarget` and
`RollbackToExactTarget` through the existing `ExecuteParent` parent path
introduced by TASK-028 and completed by TASK-029. There is no fallback between
actions and no cached authorization result.

## 11. Lock and Ownership Proof Plan

Before any implementation slice is accepted, the Roles responsible for it must
prove, with focused tests and static review of every `Lock`/`Unlock` site,
that none of these locks is held across the named operations:

- the DP-015 command-admission ledger and storage-client locks;
- the DP-014 aggregate store lock and per-aggregate lock;
- the DP-013 Directory state (which intentionally has no process lock);
- the DP-011 managed Flow (which intentionally has no process lock);
- the Runtime Lifecycle Owner mutex;
- any synchronization private to a new slice.

None is held across: (1) the authorizer callback; (2) a DP-014 storage
operation or exact inspection; (3) a continuation signal or wait; (4) a DP-013
or Owner lifecycle invocation; (5) Load, Build, Launcher, Host, or external
I/O.

Ownership remains: the external authorization policy is composition-owned and
borrowed per submission; DP-015 parent/phase records and live permits stay in
the command boundary; DP-014 aggregate and attempt facts stay in the identity
store; the Lifecycle Owner stays the sole lifecycle decision maker and live
Host owner; DP-013 stays the exact routing/private-invocation composition;
DP-011 stays the synchronous preparation sequence; the pending-Stop rendezvous
stays in the bounded orchestration execution. No row transfers ownership to
the orchestrator.

## 12. Ordered Implementation Slices (Readiness Output)

### Slice 1 — Orchestration authorizer surface

Current slice status: partial isolated implementation accepted historically by
TASK-031; its missing `OperationalDomain` is repaired in isolation by TASK-035
Slice 2R.

- Introduce the `OrchestrationAuthorizationRequest` validated value, the
  named policy-neutral `AuthorizeOrchestration` function type, the
  `OrchestrationAction` set, and the failed-authorization error contract at
  the existing command boundary without changing DP-013 / DP-015 public
  surfaces.
- Prove: exact initial/replay authorization, zero mutation on
  denial/failure/cancellation, activation zero-path on the immutable primitive
  Start submission, and no public or private bypass.
- No DP-014 binding, no Flow change, no orchestrator.

### Slice 2 — Private managed invoker and managed Flow seam

Current slice status: partial isolated implementation. TASK-032 historically
implemented the managed Flow seam, TASK-035 Slice 2R supplies the complete
authoritative binding repair, and TASK-043 implements the concrete exact-scope
invoker in isolation. Future callback custody and terminal integration remain
outside this slice's implemented proof.

- Add the managed construction and `StartManaged` per-call seam and the
  `StartExecutionBinding` / `OwnerClaimView` immutable values, the opaque
  `StartRendezvous` handle, and the failed private-invocation error contract,
  using Slice 1.
- Prove: validate-before-Owner-mutation, invoke-at-most-once, callback-scoped
  authority expiry, custody-based no-reuse, and never-stored binding without
  mutating the structural value; unchanged unmanaged `New` and `Start`; no
  goroutine, registry entry, or detached callback.
- No DP-014 binding logic yet.

### Slice 2R — Managed binding conformance repair

Current slice status: implemented and independently accepted in isolation by
TASK-035. Production composition/private-invoker wiring is not part of this
slice.

- Introduce the dependency-leaf authoritative binding values, restore
  `OperationalDomain`, carry the complete authorization tuple and the
  all-or-none linked identity, replace the constant token with a unique
  command-owned rendezvous identity, remove the Flow-to-`runtimeidentity`
  import, provide the composition-private invoker validation path, and add
  `Boundary.ExecuteManagedStart` as the sole primitive orchestration claim-to-
  binding adapter while leaving `Execute` unchanged.
- Prove constructor and cross-field validation, exact command-to-binding and
  Owner-request mapping, unique/stale/forged rendezvous rejection, synchronous
  call-stack lifetime, unchanged unmanaged Flow and DP-013 public behavior,
  acyclic imports, and zero mutation on every validation failure.
- Prove authorization-before-inspection on initial/replay, callback only for a
  new claim, atomic claim/permit/rendezvous installation, no replay adoption,
  callback lifetime expiry, unresolved-on-panic/error/indeterminate behavior,
  and no orchestration fallback to legacy `Execute`.
- Do not implement DP-014 publication/binding, continuation outcomes, an
  orchestrator, policy, persistence, API, or production wiring.

### Slice 3 — OwnerClaim-to-DP-014 binding sequence

Current slice status: implemented and independently accepted in isolation by
TASK-037.

- Implement the managed parent/StartTarget adapter, the common command-owned
  early/final rendezvous gates, and the stateless
  `StartClaimContinuation.AfterOwnerClaim` using a narrow identity-store
  boundary over existing `runtimeidentity.Store` operations, using Slices 1,
  2, and 2R.
- Prove: attempt membership and same-generation binding before Load,
  stale/different/unknown facts yield `Blocked` without preparation,
  definitive binding absence converges through the authentic Owner outcome,
  and a Stop admitted after the early check is ordered by the final gate for
  primitive and linked execution.
- Do not claim DP-014 terminal publication or DP-015 command/phase
  terminalization: the later orchestration callback may publish those only
  after the exact Owner outcome. Isolated Blocked and BindingFailed paths may
  intentionally leave command or phase Claimed.

### Slice 4 — DP-016 orchestrator readiness re-assessment

Current slice status: completed and Coordinator Accepted as TASK-038. Its 19-row reassessment records
`TASK-026 REMAINS BLOCKED`: 7 Direct, 10 Compositional, 2 Missing, and 0 Deferred
proofs after Reviewer rework and repeat Reviewer APPROVED 0/0. The accepted
verdict does not activate implementation or change TASK-026 status automatically.

- Only after Slices 1–3, including Slice 2R, are implemented and independently
  accepted, re-assess
  whether TASK-026 can be unblocked against the unmodified DP-016 §25 proofs.
- TASK-038 identifies a design-only atomic expected-attempt Owner Stop contract
  as the first bounded prerequisite. TASK-039 completed and received
  Coordinator Acceptance after recording that design in Draft DP-010;
  completed and Coordinator-Accepted TASK-040 implements and verifies the
  isolated Owner extension, with repeat final Reviewer `APPROVED` 0/0. The
  private exact-scope composition-invoker design is fixed by Draft DP-021, and
  TASK-043 implements that invoker in isolation without activating callback,
  terminal, orchestrator, or production work.

Each slice requires its own task intake, Existing Coverage Report, Verification
Matrix, Independent Review, PROCESS-002, and Coordinator Acceptance.

## 13. Acceptance Proofs

This design itself proves:

1. the chosen decomposition preserves every Approved DP-019 §21 acceptance
   proof obligation and maps each remaining proof to exactly one slice;
2. the existing DP-013, DP-014, DP-015, and DP-016 surfaces require no
   semantic change to host the original three implementation slices and the
   Slice 2R conformance repair;
3. activation stays on the primitive immutable Start command path, uses no
   synthetic parent/phase identity, and cannot bypass the same managed
   continuation prerequisite required by Approved DP-019;
4. locks are provably not held across the forbidden boundaries;
5. EN and RU mirrors are semantically equal, with equal headings and equal
   code-fence counts, and every relative link resolves.

These proofs are verified by documentation, parity, link, status, and
regression checks and by the fresh Independent Review of this proposal.

## 14. Implementation Boundary

Implementation Status remains Planned overall. This design task itself
implemented no slice; successor TASK-031 and TASK-032 produced historically
Coordinator-Accepted partial isolated implementations of Slices 1 and 2.
TASK-034 identified the remaining conformance gap, TASK-035 implements and
independently accepts its Slice 2R repair in isolation, and TASK-036 resolves
the remaining Slice-3 command-gate and continuation API ambiguity. TASK-037
implements and independently accepts Slice 3 in isolation. The
repository contains the accepted Draft design, completed TASK-040 isolated
implementation of atomic expected-attempt Owner Stop, and the TASK-043
isolated concrete private exact-scope composition invoker defined by Draft
DP-021, but still lacks the activation orchestrator,
external persistence, API, recovery worker, and production wiring. Later DP-014 terminal publication and DP-015 command/phase
terminalization after the Owner result belong to the TASK-026 orchestrator,
not to a separate prerequisite. TASK-026 therefore remains Blocked; Slice 4 is
completed and accepted as TASK-038.

## 15. Consequences

Positive:

- each remaining DP-019 prerequisite becomes independently testable and
  independently acceptable;
- authorization, lifecycle ownership, durability ordering, and permit
  ownership stay explicit;
- no production or public contract is decided here.

Costs:

- TASK-038 performs the permitted readiness reassessment, and completed
  TASK-039 records only its design-only atomic expected-attempt Stop candidate;
- the synchronous pending-Stop rendezvous may block callers;
- process restart still requires Planned DP-017 implementation;
- production integration still requires external durability and a composition
  audit.

## 16. Decision

UWP records the readiness decomposition of the remaining Approved DP-019
prerequisites in this Draft/Planned proposal and implements each slice only
through a separate, individually reviewed task. Slices 1 and 2 remain
historically accepted partial implementations; TASK-035 implements and
independently accepts Slice 2R in isolation; TASK-036 fixes the exact Slice-3
protocol; and TASK-037 implements and independently accepts Slice 3 in
isolation. Slice 4 is completed and accepted as TASK-038 with a remains-Blocked
verdict. This proposal does not approximate DP-016
with an adapter, does not add replacement/rollback operations to the Owner,
does not transfer permits, does not change any Approved status or semantic,
and does not treat planned capability as implemented.

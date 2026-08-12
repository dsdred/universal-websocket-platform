# DP-019: Runtime Activation Orchestration Prerequisites

[Russian version](../../ru/design/DP-019-runtime-activation-orchestration-prerequisites.md)

## 1. Status

- **Design Status:** Approved
- **Implementation Status:** Planned overall; parent/phase storage, callback
  capability, sequential phase core, and command-boundary Continue/pending-Stop
  rendezvous implemented in isolation; Slice 3 managed gates and continuation
  implemented and independently accepted in isolation

This focused design closes only the integration-contract ambiguity discovered
by TASK-026. The repository implements the isolated DP-015 parent/phase core,
command-boundary Continue/pending-Stop rendezvous, exact authorization and
dependency-leaf binding values, primitive managed adapter, and managed
Flow/OwnerClaimView seam. TASK-037 implements the managed parent/StartTarget
adapter, common managed gates, concrete stateless continuation, DP-014
attempt/generation binding sequence, and exact managed Flow outcome adaptation,
implemented and independently accepted in isolation. The concrete private
scoped invoker, later terminal publication, activation orchestrator, external
persistence, API, recovery worker, and production wiring remain absent.
TASK-026 remains Blocked until the prerequisites defined here are implemented
and independently accepted.

## 2. Purpose

Define the exact internal contracts required to implement Approved DP-016
without weakening its ordering or acceptance proofs:

- a bounded DP-015 parent/phase claim surface for replacement and rollback;
- exact authorization of activation, replacement, and rollback intents;
- a private DP-011/DP-013 continuation after the sole Owner claim and before
  Load;
- durable publication of that exact attempt and its execution generation
  before external preparation work.

## 3. Authority

This proposal refines, without overriding:

- [ARCH-004](../architecture/ARCH-004-runtime-deployment-and-identity-model.md),
  especially section 19(4);
- [DP-010](DP-010-runtime-lifecycle-owner-contract.md), whose Owner remains the
  sole lifecycle authority;
- [DP-011](DP-011-runtime-launch-pipeline-integration.md) for the synchronous
  launch path and private continuation location;
- [DP-013](DP-013-runtime-management-routing.md) for exact routing and
  authorization-before-mutation;
- [DP-014](DP-014-runtime-operational-identity-persistence.md) for aggregate,
  attempt, revision, and execution-generation facts;
- [DP-015](DP-015-runtime-management-command-idempotency.md) for claim,
  replay, permits, and unresolved barriers;
- [DP-016](DP-016-runtime-activation-replacement-rollback.md) for the complete
  activation/replacement/rollback ordering;
- [DP-017](DP-017-runtime-recovery-reconciliation.md) for fail-closed recovery.

Accepted ADRs and Active or Frozen architecture remain authoritative.

## 4. Scope

DP-019 defines one internal integration contract containing:

- orchestration actions and exact authorization input;
- replacement/rollback parent identity and two finite linked phase kinds;
- callback-scoped parent and phase execution capabilities;
- phase admission, replay, terminalization, and unresolved behavior;
- private exact-scope lifecycle invocation after authorization;
- the Start-claim continuation construction, input, outcomes, and ordering;
- exact attempt publication and generation binding before Load;
- cancellation, lock, ownership, and failure boundaries.

It does not implement these contracts or change Runtime lifecycle behavior.

## 5. Non-goals

This design does not define:

- a public Activate, Replace, or Rollback API, DTO, route, or status code;
- Principal representation or concrete policy;
- a generic workflow engine, saga framework, command bus, or registry;
- new Owner states, operations, or ownership;
- storage schema, transaction product, migration, or cross-process lease;
- recovery execution, reporting, automatic rollback/restart, zero-downtime
  replacement, supervision, scheduling, or Production Activation.

## 6. Decision Summary

UWP selects the DP-015 parent/phase API path rather than inventing new
lifecycle operations.

- Initial activation remains one primitive exact-version Start command.
- Replacement and rollback are distinct parent operation kinds.
- Each accepted parent has at most `StopOld` followed by `StartTarget`.
- A parent permit never invokes lifecycle work.
- Only a newly committed phase permit invokes one exact scoped lifecycle
  operation, and the permit remains on its original synchronous call stack.
- Owner alone claims and converges a Launch Attempt.
- The private continuation publishes the exact Owner claim into DP-014 and
  binds it to the composition-owned generation before Load.

No DP-016 acceptance proof is deferred or weakened.

## 7. Exact Authorization Model

The orchestration boundary defines these caller actions:

```text
ActivateExactTarget
ReplaceWithExactTarget
RollbackToExactTarget
```

Every initial submission and replay is authorized before command inspection or
claim. The immutable authorization input is exactly:

```text
OperationalDomain
WorkspaceID
ConfigurationID
RuntimeInstanceID
Action
TargetConfigurationVersionID
```

The planned policy-neutral seam is one synchronous function conceptually
equivalent to:

```text
AuthorizeOrchestration(context, AuthorizationRequest) -> error
```

`AuthorizationRequest` is a validated immutable value containing exactly the
tuple above. Initial activation adapts its existing primitive DP-015 Start
submission to `ActivateExactTarget`; `ExecuteParent` uses
`ReplaceWithExactTarget` or `RollbackToExactTarget`. There is no fallback from
one action to another and no cached authorization result.

All identities are non-zero and exact. The target version must be Published
and belong to the same Configuration. Rollback differs by caller intent and
the DP-016 historical-target precondition; it does not infer a previous or
latest version. Principal, credential, deadline, trace data, current aggregate
state, and inferred source version are not durable intent fields.

Authorization denial, failure, panic, invalid target, absent scope, or
pre-claim cancellation causes zero command, aggregate, and lifecycle mutation.
Authorization is evaluated again for replay and is never stored as durable
authority.

## 8. Authorization and Phase Authority

Authorization applies to the complete external orchestration intent. Linked
phases are not new caller submissions and are not independently rebound to a
different Principal, Target, version, or policy result. Their authority is
derived only from the accepted immutable parent and remains limited to the
declared phase.

The exact management scope exposes a private composition-only lifecycle
invoker after authorization. It routes only to the already bound Flow/Owner;
it is not a public Directory bypass. Its planned Start path is conceptually:

```text
InvokeManagedStart(context, StartRequest, StartExecutionBinding) -> exact Owner outcome
```

`StartExecutionBinding` is a per-invocation immutable value containing the
validated `AuthorizationRequest`, expected aggregate revision, exact execution
generation, parent/phase identity when applicable, and the opaque
`StartRendezvous` for that live primitive/phase execution. It contains no
primitive, parent, phase, or Stop permit. The call stack holding the exact live
permit constructs the binding and synchronously invokes this operation once;
it cannot obtain another scope, change the target, or transfer its permit.

The private invoker validates the binding against its immutable scope and the
StartRequest, then calls the one stored Flow through its planned managed Start
surface with that same binding. It creates no Flow, registry entry, mutable
current-operation slot, goroutine, or detached callback.

Existing public DP-013 Start/Stop/Observe authorization behavior remains
unchanged for primitive submissions. Production composition must prove there
is no unauthenticated transport path to the private invoker.

## 9. Parent Identity and Immutable Intent

A parent command identity remains `(CommandScope, CommandKey)`. The parent
scope operation is exactly `Replace` or `Rollback`. Its immutable intent
contains the exact authorization tuple and target version.

The coherent source observation, old attempt identity, old pinned version,
expected aggregate revision, and execution generation are bound execution
preconditions. Once the parent is claimed they may only be narrowed by exact
conditional reads; they cannot retarget the immutable caller intent.

Initial activation uses the existing primitive Start scope and intent. It does
not create a one-phase parent solely for API uniformity.

## 10. Finite Parent/Phase State Machine

The only legal linked shapes are:

```text
Replace parent: [StopOld when required] -> StartTarget
Rollback parent: [StopOld when required] -> StartTarget
```

If no active attempt exists, `StopOld` is omitted. If the exact target is
already Running, the parent may terminalize Satisfied with no phase. No other
phase kind, repetition, reordering, branch, loop, retry phase, or caller-
selected phase identity is valid.

Each phase identity is derived collision-safely from parent identity, phase
kind, and fixed ordinal. It is immutable, not caller-selectable, cannot be
reused by another parent, and remains retained with the parent.

## 11. Parent/Phase Claim API

The DP-015 extension is callback-scoped and preserves the existing
non-transferable permit invariant. Its durable parent/phase storage,
generation-bound callback capability, and strict sequential core are
implemented in isolation; the command-boundary Continue/pending-Stop surface is
also implemented there in isolation. Managed continuation and binding remain
Planned.

Conceptually it provides:

```text
ExecuteParent(submission, authorize, invokeParent) -> admission/result

invokeParent(parentExecution):
    InspectOrClaimPhase(StopOld, invokeExactStop)
    ContinueOrClaimPhase(StartTarget, invokeManagedStartWithBinding)
    PublishParentTerminal(exact linked outcomes)
```

`ExecuteParent` performs validation, exact authorization, cancellation gate,
same-key decision, per-Instance barrier evaluation, and parent claim under the
existing admission boundary. Only the call that commits a new parent receives
`parentExecution`.

`parentExecution` is valid only during its synchronous callback and for its
exact storage-client generation. Retaining it, returning from the callback,
panic, `runtime.Goexit`, or generation replacement expires every unpublished
live capability and leaves any committed non-terminal parent/phase unresolved.

## 12. Phase Claim Semantics

Claiming a phase atomically verifies:

- exact parent identity, intent, state, revision, and live parent execution;
- expected next phase kind and ordinal;
- the per-Instance barrier and the one DP-016 Stop exception;
- absence of an already committed conflicting phase;
- current storage-client generation.

A newly committed phase creates one durable linked record and one private live
phase permit. Its callback performs at most one exact lifecycle invocation.
Same-phase replay returns in-progress or terminal facts without a permit.
Definitive claim failure performs no lifecycle mutation. Indeterminate claim,
lost callback, invalid outcome, panic, or publication uncertainty leaves the
linked set unresolved.

Parent terminal publication is permitted only after every required phase has
a definitive durable terminal outcome or the parent has a definitive
zero-mutation outcome before the next phase. A parent never fabricates a phase
result from aggregate observation alone.

## 13. Continue Gate

`ContinueOrClaimPhase(StartTarget, ...)` shares one per-Instance admission
linearization point with an independent Stop claim:

- Stop first: parent terminalizes Stopped/Cancelled and no Start phase exists;
- Start phase first: exactly one Start phase permit exists, and one later Stop
  may occupy the tracked-Start exception;
- indeterminate: neither side may infer a winner; the linked set is unresolved.

The gate never pre-creates a Launch Attempt and does not run lifecycle or
storage callbacks while holding the admission lock.

## 14. Private Start-Claim Continuation

The one long-lived managed Flow is constructed immutably with one stateless
private synchronous `StartClaimContinuation` service. Per-operation facts are
never stored in the Flow or bound at construction. The existing unmanaged Flow
constructor and its isolated semantics remain unchanged. Production activation
may use only this managed construction and invocation path.

For each primitive Start or `StartTarget` phase, the original permit-holding
stack calls:

```text
Flow.StartManaged(context, StartRequest, StartExecutionBinding)
```

`StartManaged` validates the per-call binding before Owner mutation and retains
it only on that synchronous call stack. After `Owner.PrepareStart`, it calls
`StartClaimContinuation.AfterOwnerClaim(StartExecutionBinding, OwnerClaimView)`.
Returning from `StartManaged` invalidates the per-call binding; it is never a
Flow field and cannot affect another invocation.

Immediately after `Owner.PrepareStart` returns a successful authentic
preparation and before Load, Build, or Launcher work, Flow passes one immutable
Owner claim view containing:

```text
WorkspaceID
ConfigurationID
RuntimeInstanceID
LaunchAttemptID
TargetConfigurationVersionID
```

The exact attempt and version come from the Owner-issued LoadRequest. The
continuation obtains expected revision, generation, authorization tuple, linked
execution identity, and rendezvous only from the same per-call
`StartExecutionBinding`. Cross-call or scope mismatch is fail-closed. Neither
value carries a preparation token, Host, Snapshot, context cancellation
authority, parent/phase/Stop permit, or mutable Owner state.

## 15. Continuation Outcomes

The continuation returns exactly one of:

- `Continue`: exact durable attempt membership and same-generation binding are
  confirmed, and the final Stop gate released Flow;
- `StopConverged`: the original pending Stop claimant invoked exact Stop and
  converged the claimed attempt; Flow begins no Load;
- `BindingFailed`: coherent exact evidence proves that attempt publication or
  generation binding did not commit without conflict and no external
  preparation began; Flow submits the failure through the authentic Owner
  preparation;
- `Blocked`: permit/signal loss, stale or conflicting revision, different
  generation, unavailable/unknown facts, unproven Stop convergence, or
  indeterminate publication; Flow begins no Load and the linked set remains
  unresolved.

The continuation does not publish a lifecycle, phase, or parent terminal
outcome. Only exact Owner results followed by DP-014 and DP-015 conditional
publication may terminalize them.

## 16. Attempt Publication and Generation Binding

The sole in-process claim remains `Owner.PrepareStart`. The continuation then,
before Load:

1. verifies the Owner-issued identities against the authorized target;
2. resolves the already-admitted pending-Stop rendezvous, if any, through the
   original Stop claimant; `StopConverged` exits without publishing attempt
   membership or generation binding, while `Blocked` leaves the linked set
   unresolved;
3. only after definitive proof that no pending Stop remains, conditionally
   publishes that exact Launch Attempt and version pin in
   DP-014 at the expected aggregate revision;
4. conditionally binds the exact attempt to the composition-owned execution
   generation;
5. inspects exact aggregate/attempt facts after any indeterminate result;
6. executes the final Stop-versus-Continue gate for a Stop admitted after the
   early rendezvous check.

This clarification preserves DP-016 ordering: command/phase claim precedes
Owner claim, durable attempt membership and generation binding precede all
external preparation, and Running remains impossible before Host readiness.
Neither the continuation nor persistence becomes a lifecycle owner.

Exact same-attempt/same-version membership or same-generation binding may
converge idempotently. Different attempt, version, generation, stale revision,
inactive fact, or unknown state is never repaired or replaced automatically.

## 17. Pending Stop Rendezvous

A Stop that occupies the tracked-Start exception before Owner claim remains on
its original synchronous claiming stack with its private permit. The
continuation signals only `OwnerClaimed` with the exact attempt identity.

That claimant alone rechecks its cancellation gate, invokes exact private
DP-013 Stop once, publishes its outcome, and signals the result. The
continuation receives no permit and performs no Stop on its behalf. A Start
path that definitively ends before Owner claim signals `StartNoClaim`; the
original Stop path may terminalize satisfied without lifecycle invocation.

This rendezvous completes before DP-014 attempt publication or generation
binding. `StopConverged` returns from the continuation without either write;
only a definitive no-pending-Stop result permits the publication/binding path
in section 16. A Stop admitted after that early check is ordered by the final
Stop-versus-Continue gate.

Lost signal, caller return without definitive publication, abandoned permit,
or unproven convergence yields `Blocked`, never implicit Continue.

## 18. Cancellation and Failure

Cancellation before external authorization or command claim performs zero
mutation. After parent/phase claim, cancellation cannot delete or transfer the
claim. DP-010 and DP-011 gates remain authoritative for lifecycle invocation.

Definitive pre-invocation failure may terminalize the exact parent/phase with a
zero-mutation outcome. Once mutation may have occurred, absence of exact
terminal publication is unresolved. Panic and `runtime.Goexit` run capability
expiry cleanup; they never restore a permit or authorize retry.

No failure resurrects the old Host, infers release, selects another version,
or starts automatic rollback.

## 19. Concurrency and Locks

One Runtime Instance uses one DP-015 admission boundary for parent, phase,
independent Stop, replay, and unresolved-barrier decisions. DP-014 retains its
own conditional aggregate revision boundary; Owner retains lifecycle
serialization.

No command-admission, aggregate, Directory, Flow, or Owner lock is held across:

- authorization callback;
- DP-014 storage operation or exact inspection;
- continuation signal or wait;
- DP-013/Owner lifecycle invocation;
- Load, Build, Launcher, Host, or external I/O.

Different Runtime Instances share no orchestration lock and may progress
independently.

## 20. Ownership

| Fact or capability | Owner |
| --- | --- |
| External authorization policy | composition; borrowed for each submission |
| Parent/phase records and live permits | DP-015 boundary |
| Runtime Instance/attempt/generation facts | DP-014 boundary |
| Lifecycle decision and live Host | Runtime Lifecycle Owner |
| Exact scope routing/private invoker | DP-013 composition |
| Synchronous preparation sequence | DP-011 Flow call stack |
| Continue/pending-Stop rendezvous | bounded orchestration execution |

No row transfers ownership to the orchestrator.

## 21. Acceptance Proofs

A prerequisite implementation must prove at minimum:

1. exact orchestration authorization on initial and replay submissions;
2. denial/failure/cancellation before claim causes zero mutation;
3. one parent claim and immutable intent under concurrency;
4. only legal phase order and derived identities are accepted;
5. parent permit cannot invoke lifecycle work;
6. each phase permit invokes at most one exact scope operation;
7. lost/abandoned parent or phase capability leaves a durable barrier;
8. Continue gate has one winner against independent Stop;
9. pending Stop permit remains on and is used only by its original stack;
10. Owner claim view matches exact target/attempt/version;
11. durable attempt membership and generation binding complete before Load;
12. stale/different/unknown facts return Blocked without preparation;
13. definitive binding absence converges through authentic Owner outcome;
14. panic, error, cancellation, and `runtime.Goexit` cannot duplicate work;
15. reconstruction preserves records but restores no live permit;
16. different Instances progress independently;
17. unmanaged Flow cannot be used by production activation composition;
18. no public/private routing bypass exists before authorization;
19. EN/RU and linked DP semantics remain aligned.

These proofs do not prove DP-016 orchestration itself.

## 22. Implementation Boundary

Implementation Status remains Planned overall. Package
`internal/runtimecommandidempotency` now implements in isolation exact
Replace/Rollback intent, durable parent and derived `StopOld`/`StartTarget`
records, generation-bound callback capabilities, strict optional-StopOld-then-
StartTarget ordering, phase replay, parent terminal gating, unresolved
barriers, reconstruction invalidation, the non-bypassable StartTarget Continue
gate, and the synchronous pending-Stop rendezvous with immutable signal cause.

TASK-031/TASK-032/TASK-035 implement the exact authorization and binding values,
primitive managed adapter, and managed Flow/OwnerClaimView seam in isolation.
TASK-037 implements the managed parent/StartTarget adapter, common managed
gates, concrete stateless continuation, attempt membership/generation binding
sequence, and managed Flow outcome mapping, implemented and independently
accepted in isolation. The repository still lacks the concrete private scoped
invoker, later terminal publication and command/phase terminalization,
activation orchestrator, and production composition audit required by the
complete design.

TASK-026 therefore remains Blocked. Successor tasks must implement and
independently verify the remaining prerequisites before TASK-026 may be
reconsidered against the complete unmodified DP-016 proofs. The focused
readiness decomposition of those prerequisites is recorded in the mirrored
[DP-020](DP-020-runtime-orchestration-binding-sequence-readiness.md), with
Design Status Draft and Implementation Status Planned overall, with Slice 3
implemented and independently accepted in isolation.

## 23. Consequences

Positive:

- the blocking integration surfaces become finite and independently testable;
- DP-016 ordering remains unchanged;
- authorization and lifecycle ownership remain explicit;
- permit loss and process loss fail closed.

Costs:

- at least one prerequisite implementation task precedes TASK-026;
- the synchronous pending-Stop rendezvous can block its callers;
- process restart still requires Planned DP-017 implementation;
- production integration still requires external durability and composition
  audit.

## 24. Decision

UWP will implement activation orchestration prerequisites as one focused
internal contract: exact intent authorization, bounded DP-015 parent/phase
claims, and a synchronous DP-011/DP-013 Start-claim continuation that publishes
the Owner-issued attempt and generation binding before Load. It will not
approximate DP-016 with an adapter, add replacement/rollback operations to the
Lifecycle Owner, transfer permits, or treat planned behavior as implemented.

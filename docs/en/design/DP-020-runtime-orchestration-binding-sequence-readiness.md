# DP-020: Runtime Orchestration Binding Sequence Readiness

[Russian version](../../ru/design/DP-020-runtime-orchestration-binding-sequence-readiness.md)

## 1. Status

- **Design Status:** Draft
- **Implementation Status:** Planned

Implementation progress: TASK-031 and TASK-032 produced Coordinator-Accepted
isolated partial implementations of Slices 1 and 2, and TASK-034 defined their
required conformance repair. TASK-035 implements Slice 2R in isolation: the
six-field authorization and complete binding now live in the dependency-leaf
package, primitive managed claims use the sole `ExecuteManagedStart` adapter,
and command-owned rendezvous identities are unique and callback-scoped. The
concrete DP-013 composition-private invoker and production wiring remain
absent. Slice 3 is the next Planned, unactivated slice; Slice 4 has not started.
The overall status remains Planned.

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

## 8. Deferred Decision: Private Managed Invoker and Managed Flow Seam

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

The binding and rendezvous lookup authority expire when the callback/permit
returns, panics, executes `runtime.Goexit`, or loses its Boundary generation.
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
only on that synchronous call stack, invoked at most once, and invalidated on
return; it is never stored as a Flow field.

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

### 9.2 Closed continuation outcome

`StartClaimContinuation.AfterOwnerClaim(StartExecutionBinding, OwnerClaimView)`
returns exactly one closed outcome:

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

The continuation never publishes a lifecycle, phase, or parent terminal
outcome itself; only the exact Owner result followed by the DP-014 and DP-015
conditional publications may terminalize them.

### 9.3 Binding-sequence order and revision threading

The fixed order, after the sole Owner claim and before Load, is:

1. resolve the already-admitted pending-Stop rendezvous through the original
   Stop claimant only; `StopConverged` exits without either write, and
   `Blocked` leaves the linked set unresolved;
2. only after definitive proof that no pending Stop remains, conditionally
   publish the exact Launch Attempt membership and version pin in DP-014 at the
   expected aggregate revision;
3. read the committed revision returned by that write and conditionally bind
   the exact active attempt to the composition-owned execution generation at
   that new expected revision;
4. after any indeterminate result, inspect exact aggregate facts by
   `ReadRuntimeInstance` and `ReadLaunchAttemptHistory` and converge to an
   exact existing terminal outcome or return `Blocked`;
5. execute the final Stop-versus-Continue gate for a Stop admitted after the
   early rendezvous check, then release `Continue` only under confirmed, exact
   same-generation binding.

Same-attempt/same-version membership and same-generation binding already
present are zero-mutation satisfied convergence observations. Different
attempt, different version, different generation, stale revision, inactive
fact, or unknown state is never auto-repaired or auto-replaced.

## 10. Activation Adaptation Invariant

Initial activation uses the existing primitive Start path unchanged:

- `Boundary.Execute` remains the sole command admission for the primitive
  Start submission;
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

Current slice status: partial isolated implementation accepted historically by
TASK-032; TASK-035 Slice 2R supplies the complete authoritative binding repair
without rewriting TASK-032 acceptance history.

- Add the managed construction and `StartManaged` per-call seam and the
  `StartExecutionBinding` / `OwnerClaimView` immutable values, the opaque
  `StartRendezvous` handle, and the failed private-invocation error contract,
  using Slice 1.
- Prove: validate-before-Owner-mutation, invoke-at-most-once,
  invalidate-on-return, never-stored binding; unchanged unmanaged `New` and
  `Start`; no goroutine, registry entry, or detached callback.
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

Current slice status: next Planned slice after accepted TASK-035 Slice 2R; not
activated.

- Implement `StartClaimContinuation.AfterOwnerClaim` using the existing
  `runtimeidentity.Store` conditional publication/binding operations, the
  existing pending-Stop rendezvous, and the final Stop-versus-Continue gate,
  using Slices 1, 2, and 2R.
- Prove: attempt membership and same-generation binding before Load,
  stale/different/unknown facts yield `Blocked` without preparation,
  definitive binding absence converges through the authentic Owner outcome,
  and a Stop admitted after the early check is ordered by the final gate.

### Slice 4 — DP-016 orchestrator readiness re-assessment

Current slice status: not started; it remains gated by Slice 3 implementation
and independent acceptance.

- Only after Slices 1–3, including Slice 2R, are implemented and independently
  accepted, re-assess
  whether TASK-026 can be unblocked against the unmodified DP-016 §25 proofs.
- This slice is not started by TASK-030 and may conclude that TASK-026 remains
  Blocked.

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
TASK-034 identified the remaining conformance gap, and TASK-035 implements and
independently accepts its Slice 2R repair in isolation. The
repository still lacks Slice 3 attempt publication/binding composition, the
concrete private composition invoker, activation orchestrator, external
persistence, API, recovery worker, and production wiring. TASK-026 therefore
remains Blocked; Slice 3 is the next unactivated slice, and later prerequisites
still require separate implementation and
acceptance before readiness may be reconsidered against unchanged DP-016.

## 15. Consequences

Positive:

- each remaining DP-019 prerequisite becomes independently testable and
  independently acceptable;
- authorization, lifecycle ownership, durability ordering, and permit
  ownership stay explicit;
- no production or public contract is decided here.

Costs:

- independent Slice 2R acceptance and Slice 3 still precede any DP-016
  readiness re-assessment;
- the synchronous pending-Stop rendezvous may block callers;
- process restart still requires Planned DP-017 implementation;
- production integration still requires external durability and a composition
  audit.

## 16. Decision

UWP records the readiness decomposition of the remaining Approved DP-019
prerequisites in this Draft/Planned proposal and implements each slice only
through a separate, individually reviewed task. Slices 1 and 2 remain
historically accepted partial implementations; TASK-035 implements and
independently accepts Slice 2R in isolation, and Slice 3 becomes the next
Planned, unactivated slice. This
proposal does not approximate DP-016
with an adapter, does not add replacement/rollback operations to the Owner,
does not transfer permits, does not change any Approved status or semantic,
and does not treat planned capability as implemented.

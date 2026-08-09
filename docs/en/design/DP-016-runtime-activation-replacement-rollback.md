# DP-016: Runtime Activation, Replacement, and Rollback

[Russian version](../../ru/design/DP-016-runtime-activation-replacement-rollback.md)

## 1. Status

- **Design Status:** Approved
- **Implementation Status:** Planned

This approved design defines a planned ordering
contract for activation, replacement, and explicit rollback of one Runtime
Instance. No activation/replacement orchestrator or its workflow persistence,
API, recovery worker, or production wiring exists as a result of this document.
Implementation is architecture-blocked until the Approved/Planned DP-019
parent/phase, authorization, and Start-claim continuation prerequisites are
implemented and independently accepted.

## 2. Purpose

Define how one exact Published ConfigurationVersion becomes the source of a
Runtime Instance execution, how an active execution is replaced without Host
overlap, and how a caller explicitly rolls back to another exact Published
version without reusing history or inventing automatic fallback policy.

## 3. Authority

This proposal refines, without overriding:

- [ARCH-004](../architecture/ARCH-004-runtime-deployment-and-identity-model.md),
  especially section 19(4);
- [DP-013](DP-013-runtime-management-routing.md) for exact routing and
  authorization-before-mutation;
- [DP-014](DP-014-runtime-operational-identity-persistence.md) for aggregate,
  Launch Attempt, revision, and lifecycle publication semantics;
- [DP-015](DP-015-runtime-management-command-idempotency.md) for durable
  command identity, execution permits, replay, and unresolved barriers.
- [DP-011](DP-011-runtime-launch-pipeline-integration.md) for the Owner-owned
  Start claim and the planned private claim-continuation extension.
- [DP-017](DP-017-runtime-recovery-reconciliation.md) for the recovery boundary
  that requires and consumes the DP-014-owned execution-generation binding.
- [DP-019](DP-019-runtime-activation-orchestration-prerequisites.md) for the
  exact internal API that joins authorization, parent/phase admission, Owner
  claim, durable attempt publication, generation binding, and Continue.

Accepted ADRs and Active or Frozen architecture remain authoritative. DP-013
remains Draft and is implemented in isolation. Approved DP-014 and the
primitive DP-015 boundary are implemented in isolation; the DP-019 parent/phase
extension and Approved DP-016 and DP-017 remain Planned.

## 4. Scope

The design covers initial activation, already-satisfied activation, ordered
replacement, explicit rollback, exact target-version selection, Stop races,
caller cancellation, failure cut points, and indeterminate outcomes.

It does not define public commands, HTTP resources, storage schema, zero-
downtime replacement, overlapping Hosts, in-place reload, automatic rollback,
restart policy, recovery, reconciliation, reporting, or production activation.

## 5. Terms

**Activation target** is the exact Published ConfigurationVersion selected by
one authorized immutable command intent for one Runtime Instance.

**Activation** is the ordered creation and start of one new Launch Attempt from
an exact target when no conflicting active execution must first be released.

**Replacement** is one orchestration intent that releases a different active
or starting attempt before activating a new attempt pinned to the target.

**Rollback** is an explicit replacement whose target is an exact earlier
Published ConfigurationVersion. It is not reversal of history or resurrection
of an old Launch Attempt.

**Proven release** means the Lifecycle Owner has confirmed that the preceding
attempt owns no Host resources, as required by ARCH-004 and DP-014.

**Continue gate** is the per-Instance linearization boundary between an
accepted Stop intent and claim of the orchestration's linked Start phase.

**Parent orchestration command** is the durable replacement or rollback intent.
Its process-local permit may advance only the finite linked phases defined here
and never invokes DP-013 directly.

**Linked phase claim** is a durable, parent-derived `Stop old` or `Start target`
fact. Its newly issued phase permit authorizes at most one exact DP-013
invocation. It has no caller-selected key and cannot be reused across phases.

## 6. Responsibility Boundary

Runtime Lifecycle Owner remains the sole lifecycle decision maker and owner of
the live Host reference. It performs or converges Start and Stop for exact
attempts. The orchestration contract sequences existing responsibilities; it
does not become a second lifecycle owner.

DP-014 persistence records truthful aggregate and attempt facts. DP-015 owns
command claim, execution permit, replay, and unresolved barriers. DP-016 owns
only the bounded parent/phase shape, order, and preconditions between those
boundaries. A parent permit is not lifecycle authority; only a linked phase
permit reaches DP-013, whose Lifecycle Owner remains authoritative. Runtime
Host does not read management state, select versions, or perform replacement.

## 7. Supported Intents

This design defines semantic intents, not a public API:

- activate an exact target when no different active attempt must be released;
- replace the current exact attempt with a new attempt pinned to another exact
  target;
- explicitly roll back by replacing with a caller-selected exact Published
  version.

Observe and Stop retain their existing semantics. Create, Delete, Restart,
automatic rollback, latest-version activation, and policy-driven deployment are
not introduced.

## 8. Target Preconditions

Before durable command claim or lifecycle mutation, the boundary validates that
the Target addresses the exact Workspace, Configuration, and Runtime Instance,
and that the requested ConfigurationVersion is exact, non-zero, Published, and
belongs to that Configuration.

Authorization occurs for the current caller, exact action, Target, and target
version before command mutation. No latest, previous, recommended, fallback,
or inferred version is selected. Failed validation, lookup, or authorization
performs zero command and lifecycle mutation.

## 9. Coherent Starting Observation

Ordering begins from one coherent observation of the Runtime Instance
aggregate and current Owner facts. It includes expected aggregate revision,
desired and actual facts, active-attempt identity when present, exact version
pin, and whether the current process owns the corresponding live operation.

A persisted Running or Stopping fact after loss of the Owner is not liveness
proof. If live ownership cannot be established, no activation, replacement, or
rollback proceeds; the DP-015 unresolved barrier remains authoritative, and
Approved [DP-017](DP-017-runtime-recovery-reconciliation.md) defines the planned
section 19(5) recovery contract.

## 10. Initial Activation

From confirmed no-active-attempt `Stopped` or an eligible resource-free
`Failed` state, activation follows this order:

```text
authorize exact target
    -> claim durable command and execution permit
    -> revalidate aggregate identity and revision
    -> claim one new Launch Attempt with exact target pin
    -> confirm the DP-014-owned execution-generation binding required by DP-017
    -> Load -> Build -> Start through the existing Flow and Owner
    -> publish Running only after Host readiness
    -> publish replay-equivalent command outcome
```

Loader, Builder, Launcher, or Host work never begins before confirmed command
and Launch Attempt claims. At most one attempt is accepted. Startup failure is
retained as that attempt's truthful outcome and never publishes Running.
Initial activation is one primitive Start command under DP-015; it is not a
two-phase parent orchestration.

## 11. Already-Satisfied and In-Progress Targets

If the Runtime Instance is confirmed Running on the exact target version with
desired Running, activation is already satisfied and performs zero lifecycle
and aggregate mutation. The command may publish a terminal satisfied outcome.

If the exact target is already Starting under a tracked execution, no new
attempt is created. Same-command replay observes the existing command; a
different activation intent receives a non-mutating in-progress outcome.

Running or Starting on a different target is never silently treated as
satisfied. It requires ordered replacement.

## 12. Replacement Ordering

Replacement preserves the Runtime Instance identity and uses this strict order:

```text
claim exact parent replacement command and orchestration permit
    -> claim linked Stop-old phase and its one DP-013 Stop permit
    -> invoke or converge Stop for the old active attempt
    -> prove release of all old Host resources
    -> publish old attempt terminal, aggregate Stopped, and Stop-phase outcome
    -> pass the Continue gate
    -> claim linked Start-target phase and its one DP-013 Start permit
    -> invoke DP-013 Start; Owner claims the exact new Launch Attempt
    -> resolve pending Stop at the private Start-claim continuation before Load
    -> confirm the DP-014-owned execution-generation binding required by DP-017 before Load
    -> publish Start-phase and parent outcomes truthfully
```

The new attempt has a fresh identity and immutable target pin. It never reuses
the old attempt, Host, Snapshot, context, Listener, or readiness. A service gap
between proven release and new readiness is accepted by this initial single-
node contract.

## 13. Replacement During Starting

If the old active attempt is still Starting on a different target,
command admission atomically claims the replacement parent and its linked
Stop-old phase in the single DP-015 tracked-Start Stop exception. That phase's
permit invokes Stop, which captures the same attempt, prevents Running
publication, and waits for startup rollback or Host shutdown convergence. No
independent Stop can occupy the exception at the same time.

No replacement attempt is claimed until the old attempt is historical and
resource release is proven. If the old attempt is Starting on the exact target,
section 11 applies instead and replacement creates nothing.

## 14. Existing Stop and Stopping State

A replacement or rollback does not bypass an already claimed independent Stop.
While the active attempt is Stopping under another command, replacement or
rollback cannot claim its parent or a new attempt and receives a non-mutating
in-progress outcome under DP-015 admission. Same-parent replay only observes
the exact tracked parent and linked phase.

Only confirmed terminal Stop and proven release allow evaluation of the
Continue gate. Stop failure or unproven cleanup retains the active attempt and
blocks replacement.

## 15. Continue Gate and Concurrent Stop

Immediately before a replacement or rollback continues, one per-Instance
atomic boundary orders a newly authorized independent Stop claim against the
parent's linked Start-target phase claim:

- if Stop linearizes first, orchestration records a stopped/cancelled terminal
  outcome and neither Start phase nor new attempt is claimed;
- if the Start-phase claim linearizes first, its phase permit is the only
  authority for one DP-013 Start; exactly one later Stop may claim the existing
  tracked-Start exception and becomes pending until Owner claim;
- concurrent observations cannot see both “Stop won with no attempt” and an
  accepted Start phase.

The same rule closes the gap after old release and before the Start-phase claim.
A definitive phase-claim failure leaves the parent terminal and the Instance
Stopped; an indeterminate claim makes the linked command set unresolved. DP-013
then invokes existing Flow, and only Owner may claim the Launch Attempt. This
gate never pre-creates an attempt. The planned private continuation below is a
claim-observation coordination seam, not an ownership handoff.

The planned private DP-011/DP-013 Start-claim continuation runs synchronously
after Owner claim and before Load. If a pending Stop exists, the continuation
signals the original blocked Stop claimant. That call stack retains its permit,
checks cancellation, alone invokes exact DP-013 Stop, publishes its outcome,
and signals the result back. The continuation never receives the Stop permit or
caller context.

If no pending Stop remains, the continuation conditionally publishes the exact
DP-014 execution-generation binding. Exact same-generation presence proceeds.
Only absence proven by a coherent exact read for the exact still-active attempt
at the expected revision enters a final Stop-ordering gate and may return
`BindingFailed` without terminal publication. A different generation, stale
revision, conflicting or inactive state, unavailable store, or unknown result
requires an exact re-read and then either convergence to an exact terminal
outcome or `Blocked`; none of those cases becomes `BindingFailed`. Flow submits
that failure through existing
Owner.Start with the authentic preparation token. Owner's mutex orders failure
acceptance against a later Stop, and only the exact returned outcome may drive
DP-014 and command/phase terminal publication. After confirmed
binding, a final per-Instance gate orders a new Stop claim against `Continue`.
Stop winning converges before Load; `Continue` winning releases Flow, and a
later Stop reaches the claimed attempt normally. No admission or Owner lock is
held across persistence, wait, or Stop convergence. The current isolated Flow
implements neither this continuation nor the binding gate, so DP-016 remains
Planned.

## 16. Explicit Rollback

Rollback is accepted only as a new authorized and idempotent intent naming one
exact Published ConfigurationVersion of the same Configuration. The target may
have been used by historical attempts, but rollback creates a fresh command
identity and fresh Launch Attempt identity.

No previous-version pointer, stack, latest-minus-one rule, timestamp ordering,
or automatic candidate is inferred. If the exact rollback target is already
Running, the result is satisfied with zero mutation. Otherwise rollback follows
the same replacement ordering and failure rules as any other target.

## 17. Desired and Actual Facts

This contract introduces no `Replacing`, `RollingBack`, `Activating`, or other
new operational state. During release of the old attempt, existing Stop
semantics publish desired Stopped and actual Stopping, then actual Stopped only
after proven release.

Claim of the new attempt publishes desired Running and actual Starting through
DP-014. Running is published only after readiness. The immutable command intent
retains the overall target across phases; it does not falsify the aggregate's
current desired or actual fact.

## 18. Linearization Points

The required semantic linearization points are:

1. durable primitive command claim or parent orchestration claim and permit;
2. linked old-attempt Stop phase claim and phase permit when active attempt
   exists;
3. old-attempt terminal publication with proven release;
4. Continue gate ordering independent Stop claim against linked Start-phase
   claim;
5. one DP-013 invocation by the newly issued Start-phase permit, followed by
   Owner's exact new Launch Attempt claim/version pin;
6. DP-014 conditional binding of that exact attempt to the composition-owned
   execution generation, including exact inspection after indeterminate;
7. final Start-claim continuation gate ordering pending Stop against release of
   confirmed binding to external preparation;
8. Running or terminal failure publication;
9. linked phase and replay-equivalent parent terminal publication.

Each DP-014 publication uses exact expected aggregate revision. Long Load,
Build, Start, Stop, or wait work never executes under a command-admission or
aggregate lock.

## 19. Failure Matrix

| Failure cut | Truthful result | Forbidden consequence |
| --- | --- | --- |
| validation/lookup/authorization | zero mutation | command claim or lifecycle call |
| command claim definitive failure | zero lifecycle mutation | detached retry |
| linked Stop phase claim definitive failure | parent becomes terminal without Stop | lifecycle invocation |
| old Stop failure or cleanup unproven | old attempt remains active/Stopping or Failed truthfully | new attempt claim |
| old Stop indeterminate | command remains unresolved | assume release or continue |
| proven release, Start-phase claim definitive failure | aggregate remains Stopped with old attempt historical | resurrect old Host |
| Start-phase claim indeterminate | linked command set remains unresolved | Start, Stop exception, or another claim |
| phase claimed, DP-013 Start cancelled before Owner claim | parent and phase terminalize without new attempt | fabricate Starting or Running |
| pending Stop claimant cancellation before delegation | Stop terminal no-mutation; continuation may proceed to binding | transfer or invoke its permit |
| pending Stop caller return or permit loss | linked command set remains unresolved before Load | second permit or preparation work |
| Owner claim, coherent exact binding absence | final gate may return BindingFailed; Flow converges exact token through Owner.Start | continuation publishes lifecycle/command facts or begins Load |
| different generation, stale/conflicting/inactive, unavailable, or unknown binding facts | exact re-read; converge to an exact terminal outcome or Blocked | BindingFailed or resource-free inference |
| Owner claim, binding or inspection indeterminate | linked command set remains unresolved before Load | preparation work, new generation, or blind retry |
| BindingFailed, Owner failure acceptance wins | Owner returns preparation failure; then persist Failed and command/phase outcome | publish before Owner outcome |
| BindingFailed, later Stop wins Owner mutex | Owner returns stopped-before-running; persist that exact outcome | overwrite with binding failure |
| binding confirmed, final Stop gate wins | original Stop claimant converges exact attempt before Load | return Continue or transfer Stop permit |
| binding confirmed, final release gate wins | Flow may begin Load; later Stop uses ordinary tracked route | bypass exact attempt or second permit |
| Owner/durable terminal publication indeterminate | linked command set remains unresolved before Load | terminal command or preparation work |
| new startup failure | new attempt becomes historical failure when resources are released | automatic rollback or Running |
| Running/terminal publication indeterminate | exact command remains unresolved | another replacement or fabricated success |
| command terminal publication indeterminate | inspect exact command and aggregate facts | re-execute with another key |

A definitive failure after old release may leave the Runtime Instance Stopped.
That is a truthful, supported failure outcome rather than permission to revive
the old execution.

## 20. Caller Cancellation

Cancellation visible before command claim performs zero mutation. Between
parent claim and linked old-Stop phase claim it may terminalize the parent
without lifecycle mutation if the existing gate confirms cancellation first.

After the linked old-Stop phase claim wins, DP-010 Stop convergence is authoritative. Caller
cancellation may interrupt only the waiter; without a definitive outcome the
parent and phase become unresolved and cannot proceed to a new attempt. After
proven release, cancellation is checked again before the Continue gate; if it
wins, the parent terminalizes and the Runtime Instance remains truthfully
Stopped.

After the new Start Caller Cancellation Gate wins, DP-011 synchronous wait is
authoritative and caller cancellation no longer shortens the operation.
If cancellation wins before Owner claim while a Stop is pending, Start and the
parent terminalize without an attempt and the Stop terminalizes satisfied; its
permit never invokes DP-013.

Pending Stop caller cancellation before the Owner-claim signal is published by
that original claimant as terminal no-mutation and lets the continuation proceed
to binding and its final gate. After the signal, ordinary DP-010 cancellation ordering applies. A
caller return, permit loss, unproven convergence, or indeterminate result sends
`Blocked`; Flow starts no Load and the linked command set remains unresolved.
If linked Start definitively returns before Owner claim, its path signals
`StartNoClaim`; the original Stop claimant consumes its permit as terminal
satisfied without DP-013. Lost or indeterminate signaling is `Blocked`.

## 21. Concurrency and Command Admission

DP-015 same-key replay and different-intent conflict remain unchanged. One
per-Instance admission boundary serializes replacement/rollback commands and
the Continue gate. A tracked parent may advance only its next immutable linked
phase. It may coexist only with the one exact Stop exception described in
section 13. At section 15, Stop either wins before the linked Start phase or,
after that phase wins, exactly one Stop claims the tracked-Start exception. A
pre-claim Stop remains pending until the private continuation observes Owner
claim; a post-continuation Stop delegates immediately to the same Owner. An
unresolved parent, phase, or pending Stop permits no exception.

Different Runtime Instances progress independently. Commands never acquire a
cross-Instance global lock. Concurrent publication of a newer Configuration
version does not retarget an already claimed command or Launch Attempt.

## 22. Indeterminate and Recovery Boundary

Any loss of an exact parent or phase permit, Control Service termination, or
indeterminate parent/phase/attempt publication leaves the linked command set
unresolved. No new
activation, replacement, or rollback may pass admission for that Runtime
Instance until coherent inspection and an approved section 19(5) recovery
contract resolve the parent, every phase, and live-resource truth. DP-017
is Approved and defines fail-closed exact-fact reconciliation without
lifecycle replay; its Planned implementation remains absent.

This design does not hydrate an Owner, inspect a process, probe a socket,
reconcile persisted Running, or choose whether an orphan operation completed.
Those are mandatory recovery decisions, not hidden ordering mechanics.

## 23. Security and Redaction

Validation, observation, command claim, replay, and each continuation occur
inside the exact authorized Workspace, Configuration, Runtime Instance, action,
and target-version scope. A retry is authorized again; an earlier authorization
result is not durable authority.

Outcomes expose only opaque identities and redacted semantic categories. They
contain no credential, Secret, raw Configuration payload, Snapshot, internal
error, stack trace, Host pointer, process-local permit, or cross-scope state.
Concrete reporting and redaction remain section 19(6).

## 24. Technology Neutrality

This contract does not require a transaction product, database, workflow
engine, queue, distributed lock, saga framework, clock, or identifier format.
Terms such as claim, gate, revision, and publication describe observable
semantic ordering.

Generic deployment engines, universal command buses, dynamic registries, and
service locators are not required or authorized. Private implementation
mechanics must prove the contract without expanding it.

## 25. Acceptance Proofs

A future implementation must prove at minimum:

1. initial activation creates one exact version-pinned attempt;
2. exact Running target returns satisfied with zero mutation;
3. different Running target cannot change in place;
4. replacement never owns old and new Hosts simultaneously;
5. Stop during old Starting captures that same attempt;
6. new claim occurs only after old proven release;
7. Continue gate has one winner between linked Start phase and concurrent Stop;
8. Stop winning before new claim prevents any new attempt;
9. Start phase winning first permits exactly one pending Stop, which invokes
   DP-013 from its original claiming path only after Owner claim and before Load;
10. Stop after continuation reaches the exact tracked attempt and prevents or
    terminates Running under Owner rules;
11. Stop failure or unproven cleanup prevents new claim;
12. startup failure never resurrects old Host or triggers automatic rollback;
13. explicit rollback uses exact Published target and fresh attempt identity;
14. same-target rollback/activation is zero-mutation satisfied;
15. cancellation at every phase preserves truthful state and ownership;
16. indeterminate outcomes close DP-015 admission until recovery;
17. an exact DP-014-owned execution-generation binding required and consumed
    by DP-017 commits after attempt claim and before Load, or preparation does
    not begin;
18. different Instances progress independently;
19. EN/RU contract, failure matrix, gates, and planned status remain aligned.

Proofs include technically available concurrency, race, failure injection, and
storage-client-restart scenarios. They do not authorize production activation.

## 26. Formal and Downstream ARCH-004 Section 19 Gates

This Approved design closes the focused architecture design gate for ARCH-004
section 19(4). Approved DP-014, DP-015, DP-017, and DP-018 close the other
focused design gates in sections 19(2), 19(3), 19(5), and 19(6). The complete
approved set defines ordering. Isolated process-local DP-014/DP-015 stores
exist, but no activation orchestrator, external durable workflow persistence,
recovery, reporting, integration, or Production Activation exists.

## 27. Explicit Deferrals

Deferred to focused designs or implementation tasks:

- public Activate, Replace, Rollback, Restart, or deployment API;
- storage schema, workflow persistence, adapters, and migrations;
- process-restart recovery, hydration, orphan resolution, and reconciliation;
- diagnostics, audit, metrics, alerting, and redaction policy;
- zero-downtime replacement, Listener transfer, connection draining policy;
- automatic rollback/restart, retry/backoff, scheduling, and supervision;
- safe command retention, deletion, and Production Activation.

## 28. Implementation Boundary

Implementation Status is Planned. The repository contains isolated Lifecycle
Owner, launch flow, source adapter, Draft DP-013 routing, Approved DP-014
aggregate storage, and Approved DP-015 command storage. DP-016 through DP-018
remain Planned. It contains no activation/replacement orchestrator, external
durable command/aggregate/workflow storage, public management API, recovery
executor, or production wiring.

Approval closes the section 19(4) design gate but does not implement or wire
the contract. TASK-026 remains Blocked until DP-019 is implemented; no reduced
DP-016 slice is permitted.

## 29. Decision

UWP will activate one Runtime Instance only from an exact authorized Published
ConfigurationVersion. Replacement and explicit rollback preserve the Runtime
Instance identity but always create a fresh Launch Attempt after the prior Host
has proven release. Overlapping Hosts, in-place reload, inferred previous/latest
selection, and automatic rollback are forbidden.

The Continue gate atomically orders concurrent Stop against claim of the linked
Start phase. Stop either prevents that phase or occupies the one tracked-Start
exception; the private continuation signals its original claimant only after
Owner claim and before Load. Only that path uses the Stop permit, and only Owner
claims the eventual exact attempt. Failures remain
truthful and may leave the Instance Stopped or Failed; indeterminate outcomes
close command admission until separate recovery.

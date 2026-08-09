# DP-015: Runtime Management Command Idempotency

[Russian version](../../ru/design/DP-015-runtime-management-command-idempotency.md)

## 1. Status

- **Design Status:** Approved
- **Implementation Status:** Primitive Start/Stop boundary, Approved DP-019
  parent/phase sequential core, and command-boundary Continue/pending-Stop
  rendezvous implemented in isolation; the complete DP-019 extension remains
  Planned

This approved design defines the durable idempotency boundary for state-changing
Runtime management commands. Package `internal/runtimecommandidempotency`
implements the boundary in isolation over process-local in-memory storage,
without management integration, external schema, API, recovery worker, or
production wiring.

## 2. Purpose

Define how repeated, concurrent, and ambiguously completed submissions of one
authorized state-changing Runtime management intent converge on one durable
command execution and one replayable outcome without creating a second
lifecycle mutation or Launch Attempt.

## 3. Authority

This proposal refines, without overriding:

- [ARCH-004](../architecture/ARCH-004-runtime-deployment-and-identity-model.md),
  especially section 19(3);
- [DP-013](DP-013-runtime-management-routing.md) for exact routing and
  authorization-before-mutation;
- [DP-014](DP-014-runtime-operational-identity-persistence.md) for durable
  aggregate identity, conditional revision, and lifecycle-fact publication.
- [DP-019](DP-019-runtime-activation-orchestration-prerequisites.md) for the
  exact planned callback-scoped parent/phase API and authorization integration.

Accepted ADRs and Active or Frozen architecture remain authoritative. DP-013
remains Draft; Approved DP-014, the primitive DP-015 boundary, and the partial
DP-019 parent/phase sequential core and command-boundary Continue/pending-Stop
rendezvous are implemented in isolation. Managed continuation, binding, and
Approved DP-016 retain Implementation Status Planned.

## 4. Scope

The design covers:

- opaque client-supplied command identity;
- immutable command intent binding;
- authorization, validation, durable claim, and lifecycle-delegation order;
- same-key replay and different-intent conflict;
- concurrent submission, in-progress observation, and terminal replay;
- lifecycle outcome observation and unresolved-command barrier;
- definitive and indeterminate outcomes;
- retention, isolation, and redaction constraints.

The design does not define HTTP headers, DTOs, status codes, SDK behavior,
database schema, storage technology, activation, replacement, rollback,
recovery, reconciliation, or operational reporting.

## 5. Terms

**Command key** is an opaque value supplied by a caller to identify one
state-changing management intent within one exact command scope.

**Command scope** is the tuple of operational management domain, Workspace,
Configuration, Runtime Instance, and management operation kind.

**Immutable intent** is the complete normalized semantic input that can affect
the authorized lifecycle mutation. It excludes transport representation,
credentials, request context, tracing metadata, and mutable observation data.

**Command record** is the durable idempotency fact binding one command key and
scope to one immutable intent, command state, observed lifecycle facts when
known, and replayable outcome when terminal.

**Execution permit** is the one non-transferable process-local capability
returned only to the path that commits a new claim. It authorizes one exact
DP-013 invocation. It is not durable identity, lifecycle ownership, or proof
that mutation began.

A focused downstream orchestration design may define one **parent
orchestration permit** and durably linked **phase claims**. The parent permit
authorizes only the finite phase-claim transitions declared by that design; it
authorizes no DP-013 invocation. Each newly committed phase claim returns its
own non-transferable phase permit for at most one exact DP-013 invocation. A
phase identity is derived from the parent command identity and immutable phase
kind/ordinal, is not caller-selectable, and cannot be replayed as another phase.

**Replay-equivalent outcome** is the stable semantic result needed to answer a
same-intent repeat without delegating lifecycle work again. It is not a stored
HTTP response or raw internal error.

## 6. Responsibility Boundary

The command idempotency boundary owns only command claim, intent equality,
command state, the unresolved-command barrier, observed outcome facts, and
replay facts. It does not own Runtime lifecycle decisions, live Host resources,
authorization policy, transport mapping, or recovery.

Runtime Lifecycle Owner remains the only lifecycle decision maker and live
Host owner. DP-014 remains the owner of durable Runtime Instance and
Launch Attempt facts. The idempotency boundary cannot infer current liveness,
select a version, retry lifecycle work, or become a service locator.

## 7. Covered Commands

This contract applies only to state-changing management operations whose
semantics are defined by an authoritative design. In the current isolated
DP-013 surface these are Start and Stop. Observe is read-only and does not
create a durable command record.

The contract does not introduce Create, Delete, Restart, Replace, Rollback, or
any other operation. A future operation must first define its lifecycle
semantics and immutable intent before using this boundary.

## 8. Command Identity and Namespace

A command identity is the pair:

```text
(CommandScope, CommandKey)
```

CommandKey representation and allocation are implementation decisions. The
key is never a RuntimeInstanceID, LaunchAttemptID, aggregate revision,
ConfigurationVersionID, Principal, credential, timestamp, PID, pointer, or
transport request identity.

The same raw key may exist in a different command scope without collision.
Within one scope a committed key is never rebound to another intent. Cross-
scope lookup or replay is forbidden.

## 9. Immutable Intent

Intent equality is semantic and exact. It includes the verified Target and
every operation input that can change the lifecycle mutation:

- Start includes the exact Published ConfigurationVersion identity requested
  by the command;
- Stop includes the exact Target and no inferred version or latest attempt;
- future commands include only inputs defined by their approved lifecycle
  contract.

Intent excludes Principal, credential, authorization result, deadline,
cancellation state, trace ID, transport encoding, retry count, and aggregate
observations made after submission. Canonicalization or fingerprint mechanics
must be deterministic and collision-safe for equality, but this design does
not require hashing or a wire format.

## 10. Validation and Authorization Ordering

Each submission follows this order:

1. validate the command key, exact Target, operation, and operation inputs;
2. resolve the exact Runtime Instance scope without mutation;
3. authorize the current caller for the exact action and Target;
4. perform a final caller-cancellation gate required by the command contract;
5. inspect or claim the durable command record;
6. only a newly committed claim may proceed toward lifecycle delegation.

Invalid, missing, mismatched, canceled-before-claim, denied, or failed
authorization performs zero command and lifecycle mutation. Authorization is
executed for every submission, including replay. Its result is neither stored
as durable authority nor reused for another caller.

## 11. Durable Claim

A new command claim atomically publishes:

- exact command identity;
- immutable intent binding;
- state `Claimed`;
- no terminal outcome;
- no fabricated lifecycle mutation;
- a monotonic command revision or equivalent conditional token.

The complete claim commits before any Flow or Owner lifecycle method is
called. Definitive claim failure delegates nothing. Allocation of a candidate
key or private intent representation before commit does not prove that a
command exists.

## 12. Same-Key Decision Matrix

After validation and authorization, inspection of one command identity has
exactly these semantic outcomes:

| Existing record | Submitted intent | Result | Lifecycle delegation |
| --- | --- | --- | --- |
| absent | valid intent | atomically claim new command | permitted once after confirmed claim |
| non-terminal | same intent | return truthful non-terminal observation | forbidden |
| terminal | same intent | return replay-equivalent terminal outcome | forbidden |
| any | different intent | command-key conflict with zero mutation | forbidden |

A same-intent repeat never refreshes authority, changes intent, advances a
lifecycle phase, creates a Launch Attempt, or waits implicitly for completion.
Waiting or polling behavior is a future API concern.

## 13. Concurrency and Serialization

Concurrent submissions of the same command identity use one conditional
serialization boundary. At most one submission creates the claim. Every other
same-intent submission observes that claim or its later state; every different-
intent submission receives conflict.

Commands for different Runtime Instances may progress independently. For one
Runtime Instance, evaluation of every non-terminal record, the permitted
tracked-Start exception, and insertion of a new claim have one atomic command-
admission linearization point. Concurrent different keys cannot both pass a
stale barrier check.

A Claimed record with its exact live execution permit is **tracked** in the
current process. While tracked Start is executing, exactly one distinct Stop
command may claim its own permit and delegate to the same DP-013 scope. This is
the required ARCH-004 Stop-during-Starting convergence; the Owner captures the
same Launch Attempt and remains authoritative. Another Start, or another Stop
after a tracked Stop exists, receives a non-mutating in-progress conflict.

The focused DP-016 contract refines that same single exception for a tracked
replacement or rollback parent; it does not create a general orchestration
escape from the barrier. If replacement/rollback is accepted while an earlier
Start is tracked, the parent claim and its first linked Stop phase claim occupy
the one Stop exception atomically. No independent Stop can also occupy it. Once
the old attempt is released, the tracked parent may claim only its declared
linked Start phase. At the DP-016 Continue gate, one independent Stop is ordered
against that Start-phase claim: Stop either terminalizes the parent before the
phase exists, or the phase wins. After the phase wins, exactly one distinct Stop
may claim the tracked-Start exception. Before Owner claim it is recorded as
pending and its permit cannot invoke DP-013 Stop; the DP-011/DP-013 Start-claim
continuation first signals Owner claim to the original Stop claiming path. That
same blocked call stack retains the non-transferable permit, checks its own
cancellation, performs its one DP-013 Stop invocation, publishes the outcome,
and signals the continuation. The continuation never receives or invokes that
permit.

If no pending Stop remains, the continuation coordinates the exact DP-014
execution-generation binding. The same per-Instance admission boundary then
atomically orders a final Stop claim against either `Continue` for confirmed
same-generation binding or `BindingFailed` for coherently proven absence on the
exact still-active attempt/expected revision. A different generation,
stale/conflicting/inactive facts, unavailable state, or unknown is re-read and
converges to an exact terminal outcome or remains `Blocked`; it never receives a
permit or BindingFailed. Stop winning is converged by
its original claimant; `Continue` winning permits preparation, after which a
later Stop may claim the ordinary tracked-Start exception. `BindingFailed`
terminalizes no command directly: Flow converges the
exact token through Owner.Start with `FailedPreparation`, and the command/phase
terminalizes only from the exact Owner outcome after confirmed DP-014 terminal
publication. A Stop winning Owner's mutex yields the stopped-before-running
outcome instead. Binding, Owner convergence, or terminal-publication
indeterminacy is `Blocked` and unresolved.
Another Stop or lifecycle command receives a non-mutating in-progress conflict.

The pending claimant waits for exactly one process-local signal:
`OwnerClaimed`, or `StartNoClaim` when the linked Start path definitively returns
before Owner claim. `StartNoClaim` lets that same Stop path consume its permit
as terminal satisfied without DP-013 Stop. Signal loss is unresolved, never an
implicit choice.

A Claimed primitive, parent, or phase record without its exact live permit is
**unresolved**. An unresolved parent or any one of its phases is a durable
barrier against every new state-changing command and every further phase until
an approved recovery contract makes the linked command set Terminal. Approved
[DP-017](DP-017-runtime-recovery-reconciliation.md) defines fail-closed exact-
fact resolution; its Planned implementation remains absent. Observe remains
read-only. No
tracked exception applies after process restart, after loss of the claiming
call stack, or when a claim or terminal publication is indeterminate.

## 14. Lifecycle Delegation

Only the execution path that confirmed creation of a new durable claim may
delegate the command to the exact DP-013 scope. It delegates at most once in
that process execution and preserves existing Flow and Owner cancellation,
outcome, and failure semantics.

A parent orchestration path never delegates to DP-013 directly. It may advance
only by conditionally committing the next immutable linked phase declared by
the focused orchestration contract. Only the path holding that phase's newly
issued permit may perform its one exact DP-013 invocation. Parent and phase
outcomes advance monotonically; a missing or indeterminate phase outcome closes
the parent's barrier rather than authorizing another phase or permit.

A pending Stop claiming path remains synchronously blocked while retaining its
own permit. An Owner-claim signal only opens that path's downstream gate; it
does not transfer authority. The same path alone invokes DP-013 Stop and then
publishes and signals one definitive no-mutation, converged, failed, or
indeterminate outcome. Returning without such an outcome loses the live permit
and makes the pending Stop and parent unresolved.

A claimed command is not proof that lifecycle mutation began. A caller
cancellation after claim does not erase the record or permit another request
to delegate it again. The existing Flow or Owner gate decides whether mutation
begins; the command outcome must truthfully distinguish no mutation from a
started or completed mutation.

## 15. Lifecycle Outcome and Unresolved Barrier

This design does not change the exact DP-013 Start or Stop surface and does not
pass command identity into Flow, Owner, or DP-014 aggregate publication. After
the durable command claim, only the path holding its execution permit invokes
the exact DP-013 operation once. Same-key replay never receives that permit.

If that synchronous invocation returns a definitive outcome, the command
boundary may publish its replay-equivalent Terminal outcome. Any Launch
Attempt, version, Stop origin, or aggregate fact included in that outcome is
recorded only as an observed immutable fact; the boundary never fabricates an
identity that the existing outcome does not expose.

There is deliberately no promise of an atomic commit spanning the current
DP-013 call and the command record. If lifecycle mutation may have occurred but
terminal command publication is absent or indeterminate, the record remains
Claimed. Once its execution permit is gone, it is unresolved and closes the
per-Instance barrier. No retry or different key may delegate lifecycle work.
Approved DP-017 defines the section 19(5) contract for exact command,
lifecycle, and execution-evidence inspection and truthful barrier resolution;
its Planned implementation remains absent.

## 16. Command States

The minimum semantic states are:

```text
Claimed -> Terminal
```

`Claimed` means durable command ownership exists, but a replay-equivalent
terminal outcome is not durable. A matching live execution permit distinguishes
tracked execution from unresolved Claim without changing durable identity.
Lifecycle mutation may be absent, in progress, completed, or indeterminate;
Claim alone does not choose among them. `Terminal` means one replay-equivalent
outcome is durable and the per-Instance barrier may open for a later command.

Implementations may use private substates but cannot regress, skip required
truth, or present Claimed as successful terminal completion.

## 17. Terminal Outcome

Terminal publication conditionally verifies command identity, immutable
intent, current command state, command revision, and the definitive operation
outcome. It then stores one immutable replay-equivalent outcome and advances
command state to Terminal once.

The outcome records stable domain categories and identities required for
semantic replay. It does not store credentials, Principal, raw internal error,
stack trace, Host pointer, context, transport response, or mutable live
observation. Replaying a terminal outcome performs zero lifecycle and
aggregate mutation.

## 18. Definitive Failures

Validation, lookup, authorization, or pre-claim cancellation failures create
no command record. A definitive claim failure creates nothing and delegates
nothing.

After a claim exists, a definitive no-mutation lifecycle rejection may be
published as a terminal command outcome. A definitive committed lifecycle
outcome must be linked and then published terminal. Failure wording or
transport mapping may change; the stored semantic category and identity facts
must remain replay-equivalent.

## 19. Indeterminate Outcomes

If claim, lifecycle invocation, command observation, or terminal publication
has an indeterminate outcome, the caller must not:

- create a replacement command key for the same intent;
- re-delegate lifecycle work;
- create another Launch Attempt;
- assume the command is absent, failed, or terminal;
- fabricate a replay result from stale in-memory state.

The caller inspects the exact command identity and available exact Runtime
Instance, revision, Observation, and Launch Attempt facts. A coherent read may
establish absent, tracked Claimed, unresolved Claimed, Terminal, or still
unknown in the current process. Durable state alone never fabricates a live
permit. An unresolved record blocks every new state-changing command for that
Runtime Instance. Restart-time resolution and orphan convergence belong to
ARCH-004 section 19(5), not this design.

## 20. Caller Cancellation and Retry

Cancellation visible before durable claim causes zero mutation. Between claim
and the downstream Flow or Owner gate, visible cancellation may still win that
existing gate and produce a definitive no-lifecycle-mutation result.

For Start, once the DP-011 Caller Cancellation Gate wins, the same synchronous
Flow invocation no longer checks caller context and waits for one exact Owner
outcome or operation error. The idempotency boundary cannot return early or
detach that work. For Stop, the DP-010 locked cancellation gate decides whether
mutation begins; after a nil check and locked mutation win, later cancellation
may interrupt only that caller's wait while Owner-owned convergence continues.
If no definitive terminal outcome is available and the Stop caller returns,
its permit is gone; the command remains unresolved Claimed and keeps the per-
Instance barrier closed. While a Start permit remains live, the separately
claimed Stop exception remains available and reaches the same Owner.

For a DP-016 pending Stop, cancellation before the Owner-claim signal is checked
and terminally published by the original claiming path. If it definitively wins
before delegation, the permit is consumed without DP-013 mutation and the
Start continuation may proceed to execution binding and its final gate.
Cancellation visible only after the
Owner-claim signal is governed by the ordinary DP-010 Stop gate. If the pending
caller returns, loses its permit, cannot prove no mutation, or cannot publish a
definitive outcome, it signals `Blocked`; Flow begins no Load and the linked set
remains unresolved. No admission or Owner lock is held while either call stack
waits for the signal or result.

Cancellation never deletes the command, transfers command ownership to a
retry, or authorizes duplicate delegation.

A retry must use the same command identity and immutable intent. A new key is a
new command, not a retry, and remains subject to current lifecycle
preconditions. SDK retry counts, backoff, deadlines, polling, and transport
status are outside this design.

## 21. Retention and Key Reuse

Safe forgetting depends on caller retry horizons, indeterminate outcomes,
audit requirements, and Approved recovery semantics whose Planned
implementation remains absent. Therefore no command record may be deleted and
no command identity may be reused under this Approved contract.

A future focused retention contract may permit bounded expiry only if it
proves that an expired key cannot cause an old intent to be executed again or
an old terminal result to be confused with a new command. Time-to-live alone
is not such proof.

## 22. Security and Redaction

Every inspection, claim, conflict, in-progress observation, and replay occurs
only after current authorization for the exact Target and action. Responses do
not disclose whether the same raw key exists in another scope or for an
unauthorized target.

Durable command records contain opaque identity, normalized intent facts,
state, bounded observed lifecycle facts, and redacted semantic outcomes. They contain no credential,
Secret, authentication material, raw Configuration payload, Snapshot, raw
error, stack trace, Host reference, or process-local capability. Concrete
reporting and redaction policy remains required by section 19(6).

## 23. Technology Neutrality

Command key encoding, intent comparison, revision representation, storage,
locking, transaction mechanics, and serialization are implementation choices.
The observable requirements are durable claim-before-delegation, exact intent
binding, one accepted claim, atomic per-Instance command admission, one non-
transferable permit per accepted primitive or phase claim, no lifecycle
delegation by a parent permit, the bounded DP-016 phase/pending-Stop exceptions,
the Start-claim continuation before external preparation work, an
unresolved-command-set barrier, at-most-once phase delegation, monotonic
command state, truthful inspection, and replay without mutation.

Generic CRUD repositories, distributed lock services, universal command buses,
dynamic registries, and service locators are neither required nor authorized.

## 24. Acceptance Proofs

An implementation must prove at minimum:

1. one same-key/same-intent claim under concurrent submission;
2. same-key/different-intent conflict with zero mutation;
3. authorization on every initial and replay submission;
4. no command claim or lifecycle mutation before authorization;
5. durable claim before lifecycle delegation;
6. at-most-once delegation from concurrent and repeated submissions;
7. barrier evaluation and a different-key claim have one per-Instance atomic
   linearization point;
8. tracked Start permits exactly one distinct Stop claim and delegates it to
   the same Owner, while unresolved Claim blocks all new mutation;
9. non-terminal replay never reports terminal success;
10. terminal replay returns the same semantic outcome with zero mutation;
11. caller cancellation after claim does not permit duplicate delegation;
12. definitive failures preserve the specified zero-mutation boundary;
13. indeterminate outcomes force exact inspection and prohibit blind retry;
14. different command identities still obey one-Instance lifecycle
    serialization;
15. restart of a storage client preserves claim and terminal replay facts;
16. domain isolation and redaction prevent cross-scope disclosure.

Proofs include technically available concurrency, race, failure-injection,
durability, and storage-client-restart scenarios. They do not authorize
Production Activation.

## 25. Formal and Downstream ARCH-004 Section 19 Gates

This Approved design closes the focused architecture design gate for ARCH-004
section 19(3). Approved DP-014 and DP-016 through DP-018 close the other
focused design gates in sections 19(2) and 19(4)–(6). The complete approved
set defines dependency ordering. The status decisions alone create no
implementation; the current isolated package provides process-local in-memory
command storage, while external persistence, recovery, reporting, integration,
and Production Activation remain absent.

## 26. Explicit Deferrals

Deferred beyond the isolated command implementation:

- transport idempotency field, DTO, status code, and client SDK behavior;
- external durable command schema, migration, storage adapter, and production
  integration API;
- activation, replacement, rollback, and version-selection policy;
- process-restart recovery, orphan command resolution, and reconciliation;
- diagnostic taxonomy, reporting, audit, metrics, and redaction policy;
- safe command retention and deletion;
- concrete authorization policy and Production Activation.

## 27. Implementation Boundary

The primitive Start/Stop Implementation Status is Implemented in isolation.
The DP-019 parent/phase sequential core is also Implemented in isolation, while
the complete extension remains Planned. Package
`internal/runtimecommandidempotency` implements exact Scope/CommandKey identity,
immutable Start/Stop intent, authorization-before-claim, atomic per-Instance
admission, claim-before-delegation, a one-shot process-local execution permit,
the tracked-Start Stop exception, an unresolved barrier, and terminal semantic
replay. A separate `MemoryStorage` preserves claim/replay facts across
`Boundary` reconstruction but promises no persistence across process restart
and restores no live permits. `Boundary.Execute` keeps the primitive permit
private on the synchronous claiming call stack, so caller code cannot abandon
it between claim and delegation. Client-generation transition is atomically
serialized with admission; a stale Boundary cannot create a new Claim.
`ExecuteParent` adds exact Replace/Rollback intent, durable parent/derived-phase
records, generation-bound callback capability, strict optional `StopOld` then
`StartTarget` order, phase replay, parent terminal gating, and the same
unresolved barrier. `ContinueOrExecuteStartTarget` adds the non-bypassable
pre-phase Continue gate and synchronous pending-Stop rendezvous, with immutable
signal cause and fail-closed callback/reconstruction behavior.

External durable storage/schema, API, DP-016 orchestration, DP-017 recovery,
the DP-019 exact authorization, private managed continuation, Owner claim view,
and DP-014 binding prerequisites, management wiring, and
Production Activation remain absent. The isolated
package changes no lifecycle contract and is not connected to the DP-013
Directory.

## 28. Decision

UWP will identify one state-changing Runtime management intent by an opaque
command key inside an exact authorized command scope. A complete durable claim
binds that identity to one immutable intent before lifecycle delegation.
Same-intent repeats observe or replay the same command without delegation;
different intent under the same key conflicts with zero mutation.

One immutable replay-equivalent terminal outcome is retained. A live execution
permit tracks its claiming path and preserves the mandatory Stop-during-Start
exception. A Claimed record without that exact permit is unresolved and blocks
every new state-changing command for the same Runtime Instance. Cancellation
or an indeterminate outcome never authorizes blind re-execution. Runtime
Lifecycle Owner remains the sole lifecycle decision maker, while truthful
barrier resolution, recovery, retention, API mapping, and production wiring
remain separate work.

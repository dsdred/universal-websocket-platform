# DP-003: Runtime Session Manager

[Russian version](../../ru/design/DP-003-runtime-session-manager.md)

## 1. Status

**Status:** Draft

This revision simplifies the lifecycle after the second independent architecture review. It separates shutdown transition from waiting, reduces completion to one atomic mutation, and defines continuous ownership for every registration transaction. It does not define concrete Go APIs or neighboring Runtime subsystems.

## 2. Problem Statement

Runtime can create and run a Session after successful Handshake, but it has no authoritative registry or complete shutdown accounting for handed-off Sessions. The current HTTP handler synchronously owns Session `Start/Run/Stop`. Runtime cannot yet prove that every successful handoff is tracked until terminal completion or that shutdown cannot miss an in-flight registration.

[DP-001](DP-001-runtime-handshake-pipeline.md) requires a Session to enter Runtime shutdown tracking before handoff completes. [ARCH-002](../architecture/ARCH-002-runtime-foundation-freeze.md) leaves Session ownership and the complete shutdown wait set open. This proposal defines only that lifecycle foundation while preserving frozen Host, Admission Gate, Runtime context, startup, rollback, and Listener lifecycle semantics.

## 3. Goals

- Preserve one WebSocket owner through Handshake and Session handoff.
- Define reservation and committed-registration ownership without an untracked gap.
- Establish one linearization point for registration and one for completion.
- Separate nonblocking shutdown transition from context-bounded waiting.
- Make `Reserve/Register/Complete/BeginShutdown/Wait` races deterministic.
- Provide identity-safe lookup metadata without exposing Session operations.
- Keep Manager limited to registry and shutdown accounting.

## 4. Non Goals

This document does not design Router, Delivery, Presence, Groups, Persistence, cluster ownership, Plugins, Metrics, diagnostics aggregation, Session limits policy, transport protocol, public API, Configuration fields, concrete Go interfaces, a generic Session framework, or a supervision framework.

## 5. Ownership Boundaries

Four ownership concerns remain separate.

### WebSocket Transport Ownership

Upgrade boundary owns WebSocket and child connection context until Session explicitly accepts them. Before acceptance, Upgrade boundary is the sole closer. After acceptance, Session is the sole closer, including when later registration fails.

### Registration Transaction Ownership

Successful Reserve returns one conceptual reservation handle. The calling Handshake/execution-setup control flow owns that handle until exactly one terminal operation:

- `Commit`, which consumes the reservation and creates a committed record; or
- `Abort`, which removes the reservation without registry visibility.

The normative control-flow pattern is:

```text
reservation := Manager.Reserve(SessionID)
defer reservation.AbortUnlessCommitted()
```

This notation describes an ownership obligation, not a required Go API. Every recoverable return, rejection, error, or panic before Commit must execute Abort. Manager accounts for the reservation but does not own the goroutine responsible for Commit or Abort.

### Session Execution Ownership

A narrow per-Session execution owner is created before registration Commit and lives at least until completion accounting finishes. It owns Session `Start/Run/Stop`, invokes Complete through a terminal defer on all recoverable paths, and exposes a stable non-owning Stop-request capability for Runtime shutdown. It is not a generic supervisor.

### Manager Tracking Ownership

Manager owns only reservation accounting, committed registry membership, immutable lookup views, shutdown transition, and completion accounting. It does not own Session execution or WebSocket transport.

## 6. Session Manager Responsibilities

Manager contains only responsibilities tied by one invariant: Runtime is `Closed` only after every reserved transaction is resolved and every committed registration is completed.

Manager is responsible for:

- reservation accounting;
- committed registry membership;
- RegistrationID allocation and identity-safe completion;
- immutable lookup by SessionID;
- atomic `Open -> Closing` transition;
- committed shutdown snapshot;
- atomic completion accounting;
- context-bounded waiting for empty accounting.

Manager does not execute Session, own WebSocket, route or send messages, publish Presence events, retain terminal history, aggregate diagnostics, apply Session-limit policy, or know Router, Delivery, Presence, Persistence, and Groups.

Registry may remain internal in the first version because reservation, commit, lookup, and completion share one atomic invariant. The boundary must permit later extraction without changing these semantics.

## 7. Registration Transaction

The normative sequence is:

```text
Reserve
    -> create Session
    -> create execution owner
    -> Session accepts transport ownership
    -> Commit registration with RegistrationID and Stop-request capability
    -> execution owner performs Start/Run
    -> deferred Complete
```

### Reserve

Reserve is allowed only while Manager is `Open`. It validates SessionID and creates:

- a reservation handle with exclusive Commit-or-Abort obligation;
- a RegistrationID unique for the lifetime of this Manager.

Reservation is counted for Manager closure but is invisible to Lookup and is not a committed Session in the shutdown registry.

### Commit

Commit consumes its reservation and atomically creates one committed record containing immutable identity metadata and a stable Stop-request capability. This mutation is the single registration linearization point: registry visibility and committed wait-set membership appear together.

### Abort

Abort atomically removes an uncommitted reservation and wakes Wait callers when accounting may now be empty. Abort never creates lookup visibility or a committed record. Repeated Abort after Commit or Abort has no second accounting effect.

### Transport Handoff

Session accepts transport before Commit. If any failure occurs before acceptance, Upgrade boundary closes the transport and the reservation owner aborts. If failure occurs after acceptance but before Commit, Session closes the transport through execution owner and the reservation owner aborts. There is always exactly one transport owner.

## 8. Reserve/Commit/BeginShutdown Linearization

This proposal chooses the strict shutdown model.

At one Manager synchronization boundary:

- a Commit completed before `BeginShutdown` becomes a committed record and belongs to the shutdown snapshot; or
- `BeginShutdown` wins, Manager becomes `Closing`, and every uncommitted reservation loses the right to Commit.

New Reserve is rejected after `Closing`. A pre-existing reservation whose Commit lost must execute Abort through its owner. If Session already accepted transport, execution owner requests Stop before Abort completes. Manager does not forcibly abort a handle because it does not own that control flow.

There is no late Commit in `Closing` and no partially visible registration. Manager cannot become `Closed` until every invalidated reservation has reached Abort.

## 9. Registration Identity

Each reservation receives a RegistrationID with these normative properties:

- opaque and immutable;
- unique for the entire lifetime of one Manager;
- never reused, including after Abort or Complete;
- bound to exactly one reservation and at most one committed record;
- unknown, completed, or stale RegistrationID cannot change accounting.

Manager may allocate this identity without retaining removed records, provided allocation never reuses a value during its lifetime.

SessionID is:

- required and non-empty;
- at most 255 bytes;
- opaque and immutable;
- compared byte-for-byte without normalization;
- unique among current reservations and committed records.

SessionID may be reused after Abort or Complete. The identity of one registration is the pair `(SessionID, RegistrationID)`. Manager does not automatically log raw SessionID; safe diagnostics require explicit encoding.

SessionID is scoped to one Runtime instance and is not durable cross-Runtime identity.

## 10. Session and Manager Record Lifecycles

DP-003 does not add states to the existing Session state machine.

Session remains the execution source of truth:

```text
Created -> Running -> Stopping -> Stopped
```

Manager record is only registration state:

```text
Reserved -> Registered -> Removed
       \-> Aborted
```

- `Reserved` is invisible and carries Commit-or-Abort obligation.
- `Registered` is visible and counted in the committed shutdown registry.
- `Removed` follows the first valid Complete and is immediately invisible and uncounted.
- `Aborted` never becomes visible or committed.

There is no normative `Completing` state. Manager state never claims that Session is Running; Session state never claims registry membership.

## 11. Execution Owner Lifetime

Execution owner is created before Commit. Commit records its stable non-owning Stop-request capability. Execution owner remains alive until:

- its Session execution has reached a recoverable terminal path;
- it has attempted Complete;
- every Stop request already issued through its stable capability has returned.

Execution owner invokes `Start`, then `Run`, and invokes `Stop` according to the existing Session contract. A terminal defer attempts Complete exactly once for every recoverable return, error, or panic inside the owned execution goroutine.

If BeginShutdown wins after Commit but before Start, the Stop-request capability may cancel startup or stop Session from `Created`. Start may then be skipped or fail as a normal terminal path. The defer still attempts Complete, so no committed record remains stranded.

Complete does not destroy a capability while a previously issued Stop request is in progress. The shutdown snapshot owns the capability reference until that invocation returns; it does not own Session or extend registry visibility.

## 12. Completion Linearization

Complete identifies a committed record by RegistrationID. The first valid Complete performs one atomic mutation:

```text
Registered -> Removed
```

At that single linearization point Manager:

- removes lookup visibility;
- removes the committed record;
- decrements committed wait-set accounting;
- notifies Wait callers that accounting may be empty.

Repeated Complete for the same RegistrationID is a no-op or one stable already-completed/unknown-registration contract result. It never decrements accounting again. Because RegistrationID is never reused, stale completion cannot affect a later registration reusing SessionID.

Complete accounts for lifecycle termination only. Session and execution owner remain responsible for resource cleanup. Manager does not retain Session terminal result after Complete and does not aggregate it into Manager shutdown.

## 13. Manager Lifecycle

Manager lifecycle is:

```text
Open -> Closing -> Closed
```

Restart is forbidden.

### Open

- Reserve and Lookup are allowed.
- Commit and Abort are allowed for owned reservations.
- Complete is allowed for committed records.
- BeginShutdown is allowed.

### Closing

- Reserve and Commit are rejected.
- Abort and Complete remain allowed.
- Lookup returns only current Registered views.
- BeginShutdown is idempotent and refers to the same shutdown cycle.
- Wait is allowed.
- transition to `Closed` occurs automatically when reservation and committed accounting both become empty.

### Closed

- reservation set and committed registry are empty;
- Reserve and Commit are rejected;
- Lookup returns absence;
- BeginShutdown is idempotent and exposes no active Stop capabilities;
- Wait returns nil;
- Abort and Complete have no accounting effect.

## 14. BeginShutdown

`BeginShutdown` is a conceptual nonblocking transition contract. It:

- atomically changes `Open` to `Closing` once;
- forbids new Reserve and every uncommitted Commit;
- marks existing reservations as requiring Abort;
- creates or exposes the stable shutdown snapshot of all records committed before the transition;
- performs no network I/O;
- invokes no Session operation;
- waits for nothing;
- is idempotent and joins the same shutdown cycle on repetition.

The shutdown snapshot contains only stable non-owning Stop-request capabilities. It contains no raw Session, WebSocket, Handler, or mutable registry record.

Removal by Complete does not invalidate a capability already issued for the current shutdown orchestration. Its lifetime continues until the corresponding Stop request returns. Repeated BeginShutdown does not create a new cycle or duplicate accounting.

## 15. Stop-Request Capabilities

A Stop-request capability represents permission to request termination from one execution owner without acquiring Session ownership.

Each request must be nonblocking or bounded by its caller context and must not wait for full Session completion. Stop requests for different registrations are independent: one blocked or slow request cannot serialize requests to all other Sessions through Manager.

Runtime shutdown orchestration owns invocation of the snapshot capabilities. Manager only produces the stable snapshot and later observes Complete mutations. No generic executor pool, worker framework, or Session supervisor is introduced.

## 16. Wait

`Wait(ctx)` observes the one Manager shutdown cycle. It does not initiate shutdown and does not perform Stop requests.

Wait returns nil when both are empty:

- reservation accounting;
- committed registry accounting.

The mutation that removes the final reservation or committed record automatically changes `Closing` to `Closed` and notifies all Wait callers.

If caller context ends first:

- Wait returns `ctx.Err()` to that caller;
- Manager remains `Closing` while accounting is nonempty;
- no record is falsely removed;
- a later Wait continues observing the same cycle.

Manager has no independent operational failure model in this proposal. Session errors are neither aggregated nor retained. BeginShutdown returns only contract validation results, and Wait after `Closed` returns nil rather than an undefined terminal Manager error.

## 17. Runtime Shutdown Ordering

The executable Runtime ordering is:

```text
Host closes Admission Gate
    -> Manager.BeginShutdown
    -> Host cancels root Runtime context
    -> Runtime shutdown orchestration invokes stable Stop-request capabilities independently
    -> Host invokes existing Listener Stop
    -> active Session executions terminate and call Complete
    -> dependent HTTP handlers finish
    -> Listener Stop returns
    -> Manager.Wait
    -> Manager Closed
    -> remaining Runtime components stop
```

`BeginShutdown` cannot block Host before root cancellation. Manager Wait occurs after Listener Stop, while Complete happens on Session execution terminal paths before their handlers return. Therefore Listener may wait for handlers without waiting for Manager Wait, and Manager Wait observes accounting already converging independently. No lifecycle lock is held across Stop requests, Listener Stop, Session execution, or Wait.

This preserves [ARCH-002](../architecture/ARCH-002-runtime-foundation-freeze.md): Host closes admission and cancels root Runtime context before invoking existing Listener Stop. It adds no new Host state and does not change Listener ownership.

## 18. Panic, Failure, and Abandonment Limits

Execution owner guarantees a deferred Complete attempt for all recoverable returns, errors, and panics inside its owned goroutine. Reservation owner guarantees deferred Abort for all recoverable exits before Commit.

Completion is not guaranteed after:

- process termination;
- unrecoverable runtime failure;
- a permanently blocked goroutine;
- violation of the reservation or execution-owner contract by an external component.

An abandoned reservation or execution may keep Manager in `Closing` indefinitely. This is an accepted limitation of truthful shutdown accounting: Manager does not invent completion and does not force goroutine termination.

## 19. Lookup Contract

First-version Lookup is exact by SessionID and returns an immutable RegistrationView containing only:

- SessionID;
- RegistrationID;
- manager-visible state, which is `Registered` for every visible record in this version.

Lookup never returns raw Session, execution owner, Stop capability, Send operation, or mutable record.

Lookup semantics:

- reservation is invisible;
- Complete removes visibility at its single linearization point;
- lookup does not extend lifetime;
- returned view may become stale immediately;
- `(SessionID, RegistrationID)` distinguishes reuse of SessionID;
- during `Closing`, only records still Registered are visible;
- after `Closed`, lookup returns absence.

RegistrationView is not a Router, Delivery, Presence, lease, or targeting contract.

## 20. Runtime Context

Manager does not create, own, replace, or cancel root Runtime context. Host remains its sole owner under [ADR-0004](../adr/0004-handshake-runtime-dependencies.md). Root cancellation is one shutdown signal to execution owners; only Complete changes Manager accounting.

## 21. Architectural Invariants

- WebSocket always has exactly one owner.
- Reservation handle has exactly one terminal Commit-or-Abort accounting effect.
- Registration Commit is the sole registry-visibility linearization point.
- BeginShutdown is the sole `Open -> Closing` linearization point.
- No Commit succeeds after BeginShutdown.
- Complete is one atomic `Registered -> Removed` mutation.
- RegistrationID is never reused during Manager lifetime.
- Stale Complete cannot affect another registration.
- Manager becomes `Closed` only with empty reservation and committed accounting.
- BeginShutdown performs no I/O and Wait performs no Stop requests.
- Stop capabilities are stable for in-progress shutdown invocation and never expose raw Session.
- Lookup returns immutable identity metadata only and never extends lifetime.
- Manager does not execute Session, own transport, route, deliver, publish Presence, store history, aggregate diagnostics, or apply limits.

## 22. Future Integration and Limits

DP-003 guarantees only ownership-safe registration and truthful shutdown tracking.

- Router requires a separate targeting contract.
- Delivery requires a separate lifetime-aware capability.
- Presence requires a separate event or snapshot contract.
- Persistence requires Runtime instance identity.
- Groups are outside this proposal.

Session limits remain outside DP-003. Manager lifecycle provides a possible future registration-admission observation point, but policy, configured limits, and rejection semantics require separate Configuration validation and focused design. The MASTER_PLAN Session Manager epic is therefore split conceptually into this lifecycle foundation and a future limits capability.

## 23. God-Object Prevention

Manager contains only reservation accounting, committed registry, identity-safe lookup views, shutdown transition, and completion accounting. These responsibilities share one atomic emptiness invariant.

Manager explicitly does not execute Session, own WebSocket, route or send messages, publish domain events, retain terminal history, aggregate diagnostics, or apply limits policy. Future systems cannot add those responsibilities merely because Manager exposes registration metadata.

## 24. Alternatives

### Alternative A — One Blocking Manager Stop

**Advantage:** one lifecycle call.

**Disadvantage:** Host cannot deterministically close registration, cancel Runtime, invoke Listener Stop, and only then wait without hidden asynchronous behavior.

**Decision:** rejected in favor of separate BeginShutdown and Wait.

### Alternative B — Allow Reserved Transactions to Commit in Closing

**Advantage:** fewer aborted handoffs during concurrent shutdown.

**Disadvantage:** shutdown snapshot remains open to late committed records and requires dynamic Stop-capability discovery.

**Decision:** rejected for the first version. Strict BeginShutdown invalidates every uncommitted reservation.

### Alternative C — Return Raw Session from Lookup

**Advantage:** immediately useful to callers.

**Disadvantage:** exposes mutable lifecycle and creates stale-reference and ownership ambiguity.

**Decision:** rejected; Lookup returns immutable RegistrationView only.

### Alternative D — Retain Completion History

**Advantage:** distinguishes repeated completion from unknown identity and aids diagnostics.

**Disadvantage:** turns Manager into history storage and grows state beyond live accounting.

**Decision:** rejected. Unknown and already-completed RegistrationID have the same no-accounting-effect semantics.

### Alternative E — Separate Registry Component Immediately

**Advantage:** isolates indexing.

**Disadvantage:** splits one registration/completion invariant before independent consumers exist.

**Decision:** rejected for the first version; internal boundaries must still permit later extraction.

## 25. Migration Strategy

The internal lifecycle changes substantially.

Current model:

```text
HTTP handler synchronously owns Session Start/Run/Stop
```

Proposed model:

```text
Handshake owns reservation
    -> creates Session and execution owner
    -> Session accepts transport
    -> registration commits
    -> execution owner runs Session and defers Complete
```

External WebSocket behavior remains unchanged, but internal concurrency, ownership, registration, and shutdown semantics change. Migration must therefore be treated as lifecycle work, not as a behavior-neutral refactor, and must be race-reviewed at every boundary.

## 26. Open Questions

- What concrete non-owning Stop-request capability shape satisfies the bounded-return contract?
- How will terminal Session results reach a future diagnostics boundary without Manager aggregation?
- What future design owns configured Session limits?
- What Runtime instance identity will future Persistence combine with SessionID?
- What operational response is appropriate for an accepted permanently blocked reservation or execution?

These questions do not change the linearization points or Manager accounting defined here.

## 27. Architecture Review Traceability

| Finding | Status | Resolution |
| --- | --- | --- |
| F-01 | Resolved | BeginShutdown performs the nonblocking transition; Wait is a separate context-bounded observation. |
| F-02 | Resolved | Complete is one atomic `Registered -> Removed` mutation and the sole visibility/accounting linearization point. |
| F-03 | Resolved | Reservation handle gives Handshake/execution setup continuous Commit-or-Abort ownership. |
| F-04 | Resolved | Execution owner exists before Commit; shutdown between Commit and Start is a normal terminal path ending in Complete. |
| F-05 | Resolved | Shutdown snapshot contains stable Stop capabilities whose invocation lifetime survives Complete. |
| F-06 | Resolved | Stop requests are independent and nonblocking or caller-context bounded; Manager invokes none of them. |
| F-07 | Resolved | Runtime shutdown orchestration owns Stop requests; Manager progresses through Abort/Complete mutations without a waiting caller. |
| F-08 | Resolved | RegistrationView includes never-reused RegistrationID, distinguishing SessionID reuse. |
| F-09 | Resolved | RegistrationID is unique for Manager lifetime and never reused. |
| F-10 | Resolved | Undefined Manager terminal error was removed; Wait after Closed returns nil. |
| F-11 | Clarified | Deferred completion is guaranteed only for recoverable owned-goroutine paths; nonrecoverable cases are accepted limitations. |
| F-12 | Clarified | Lookup is intentionally metadata-only and is not an operational Session capability. |
| F-13 | Resolved | Lookup returns immutable RegistrationView, never raw Session or mutable execution capability. |
| F-14 | Resolved | Normative `Completing` state was removed. |
| F-15 | Resolved | Execution owner is created before Commit and survives completion plus outstanding Stop invocations. |
| F-16 | Clarified | SessionID is opaque, byte-for-byte, at most 255 bytes, and not automatically logged raw. |
| F-17 | Clarified | Migration explicitly acknowledges substantial internal lifecycle and concurrency changes. |
| F-18 | Clarified | Manager responsibilities are restricted by explicit negative invariants. |
| F-19 | Clarified | Exactly-once describes effective accounting mutation, not notification delivery. |
| F-20 | Accepted limitation | Permanent abandonment can retain truthful accounting and keep Manager Closing. |
| F-21 | Deferred | Limits are split into a future focused Configuration/admission capability. |
| F-22 | Resolved | BeginShutdown is the explicit nonblocking `Open -> Closing` contract. |
| F-23 | Resolved | Complete atomically removes visibility and wait-set accounting. |
| F-24 | Resolved | Reservation handle is the continuous owner before Commit or Abort. |
| F-25 | Resolved | Shutdown snapshot retains Stop-capability lifetime independently of registry visibility. |
| F-26 | Resolved | Manager has no terminal error aggregation; Wait returns nil or caller `ctx.Err()`. |
| F-27 | Resolved | Immutable view carries RegistrationID, so stale and reused SessionID observations are distinguishable. |

## 28. Approval Readiness

Second-review Blockers F-01 through F-03 are resolved by split shutdown contracts, atomic Complete, and continuously owned reservation handles. High findings F-04 through F-10 are resolved through execution-owner lifetime, stable Stop capabilities, independent shutdown orchestration, identity-safe views, and removal of undefined Manager terminal errors.

Accepted limitations are explicit: process termination, unrecoverable failure, permanently blocked goroutines, or external contract violation may prevent accounting from becoming empty. Manager remains truthfully `Closing` rather than declaring false completion.

Remaining questions concern concrete capability shape, future diagnostics delivery, configured limits, durable Runtime identity, and operational handling of permanently blocked work. They do not alter the core lifecycle.

**Approval decision candidate:** Approved with Findings.

The document remains Draft until independent review confirms the simplified BeginShutdown/Wait ordering, atomic completion, reservation ownership, and stable capability lifetime.

## 29. References

- [ARCH-001: Runtime Architectural Pattern](../architecture/ARCH-001-runtime-architectural-pattern.md)
- [ARCH-002: Runtime Foundation Freeze](../architecture/ARCH-002-runtime-foundation-freeze.md)
- [DP-001: Runtime Handshake Pipeline](DP-001-runtime-handshake-pipeline.md)
- [DP-002: Runtime Host Composition Root](DP-002-runtime-host-composition-root.md)
- [ADR-0004: Handshake Runtime Dependency Boundary](../adr/0004-handshake-runtime-dependencies.md)
- [Master Engineering Plan](../roadmap/MASTER_PLAN.md)

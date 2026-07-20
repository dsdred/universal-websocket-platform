# DP-003: Runtime Session Manager

[Russian version](../../ru/design/DP-003-runtime-session-manager.md)

## 1. Status

**Status:** Approved

This approved design defines the Session Manager registration, identity, lookup, shutdown-accounting, and shutdown-snapshot contracts. Detailed per-Session execution ownership is defined normatively by the approved [DP-004](DP-004-per-session-execution-boundary.md) and is not duplicated here.

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
- Define the Manager-side integration boundary for the per-Session execution contract in DP-004.
- Keep Manager limited to registry and shutdown accounting.

## 4. Non Goals

This document does not design Router, Delivery, Presence, Groups, Persistence, cluster ownership, Plugins, Metrics, a diagnostics framework, Session limits policy, transport protocol, public API, Configuration fields, a generic Session framework, or a supervision framework.

## 5. Ownership Boundaries

Four ownership concerns remain separate.

### WebSocket Transport Ownership

Upgrade boundary owns WebSocket and child connection context until the integrated Commit boundary succeeds. Before Commit, Upgrade boundary is the sole closer. Successful Commit simultaneously publishes Registration and transfers ownership to Session; no later registration failure exists. After Commit, Session is the sole closer.

### Registration Transaction Ownership

Successful Reserve returns one conceptual reservation handle. The Session handoff flow retains continuous ownership of that handle until exactly one terminal operation:

- `Commit`, which consumes the reservation and creates a committed record; or
- `Abort`, which removes the reservation without registry visibility.

The normative control-flow pattern is:

```text
reservation := Manager.Reserve(SessionID)
defer reservation.AbortUnlessCommitted()
```

This notation describes an ownership obligation, not a required Go API. Every recoverable return, rejection, error, or panic before Commit must execute Abort. Manager accounts for the reservation but does not own the control flow responsible for Commit or Abort. Handshake never owns the reservation and does not import Session Manager; it delegates post-Upgrade preparation to the Session handoff boundary. The exact transfer from this transaction into execution ownership is defined by DP-004.

### Session Execution Ownership

Session Manager does not define or own Session execution. [DP-004](DP-004-per-session-execution-boundary.md) is the normative source for the creator boundary, provisional Session model, ownership activation, Start/Run/Stop ordering, connection cancellation, terminal observation, panic handling, and Stop-request lifetime.

Manager participates only through opaque registration identity, Commit/Complete accounting, immutable shutdown identity, Owner Lifetime Lease accounting, and the remaining narrow integration contracts explicitly assigned to it by DP-004. It never receives Session, WebSocket, HTTP request, connection context, or execution-control methods.

### Manager Tracking Ownership

Manager owns only reservation accounting, committed registry membership, immutable lookup views, shutdown transition, completion accounting, and opaque Owner Lifetime Lease accounting. It does not own Session execution or WebSocket transport.

## 6. Session Manager Responsibilities

Manager contains only responsibilities tied by one invariant: Runtime is `Closed` only after every reserved transaction is resolved, every committed registration is completed, and every opaque Owner Lifetime Lease is released. Full DP-004 integration binds that implemented execution-lifetime accounting to the Owner terminal contract without changing reservation or registration mutations.

Manager is responsible for:

- reservation accounting;
- committed registry membership;
- RegistrationID allocation and identity-safe completion;
- immutable lookup by SessionID;
- atomic `Open -> Closing` transition;
- committed shutdown snapshot;
- atomic completion accounting;
- opaque Owner Lifetime Lease accounting;
- context-bounded waiting for empty accounting.

Manager does not execute Session, own WebSocket, route or send messages, publish Presence events, retain terminal history, aggregate diagnostics, apply Session-limit policy, or know Router, Delivery, Presence, Persistence, and Groups.

Registry may remain internal in the first version because reservation, commit, lookup, and completion share one atomic invariant. The boundary must permit later extraction without changing these semantics.

## 7. Registration Transaction

The Manager-visible sequence is:

```text
Upgrade boundary calls Session handoff/Dispatcher
    -> Reserve SessionID
    -> Commit registration
    -> expose immutable registration identity
    -> execution proceeds under DP-004
    -> Complete
```

### Reserve

Reserve is allowed only while Manager is `Open`. It validates SessionID and creates:

- a reservation handle with exclusive Commit-or-Abort obligation;
- a RegistrationID unique for the lifetime of this Manager.

Reservation is counted for Manager closure but is invisible to Lookup and is not a committed Session in the shutdown registry.

### Commit

Commit consumes its reservation and atomically creates one committed record containing immutable identity metadata. This mutation is the single registration linearization point: registry visibility and committed wait-set membership appear together. The current lifecycle foundation returns the opaque RegistrationID and its bound Owner Lifetime Lease to the committing control flow.

The integrated DP-004 contract extends that same atomic success result with one RegistrationID-bound Completion Adapter, one Owner Lifetime Lease, one Stop-publication binding, and one narrow one-shot execution-publication binding prepared before Commit. Dispatcher creates exactly one dormant execution path before calling Commit. Under the same Manager synchronization boundary that excludes Lookup and BeginShutdown, Commit publishes the complete committed record and owner-scoped bundle and invokes the binding's single publication operation with the `Committed` outcome. The synchronization boundary is released only after both mutations are complete.

The execution-publication binding is not a passive value, an Execution Owner reference, or a general lifecycle capability. It gives Manager only the one-shot right to publish one already prepared outcome. Publishing `Committed` makes exactly one dormant path eligible; publishing the non-committed outcome terminates that path without execution. Repeated publication returns the existing outcome and has no second effect. Manager does not construct, retain, start, stop, or observe the owner. Dispatcher remains responsible for creating exactly one dormant path and supplying exactly one binding.

Either the complete committed record, owner-scoped bundle, and committed execution binding become observable together or no committed state is created. This extension does not introduce a second Commit phase or change the registration linearization point.

Every potentially fallible computation and allocation for the integrated bundle completes before Manager enters the publication critical section. That section invokes no external code and performs only bounded Manager mutations plus the prepared panic-free one-shot publication operation. Recoverable Commit therefore has only non-committed failure before observable mutation or complete successful publication. Process termination or unrecoverable Go runtime failure is outside this recoverable atomicity guarantee.

Commit is irreversible. Commit does not return success until the owner path is eligible to execute committed lifecycle. After successful Commit, Registration is externally observable through Lookup, belongs to shutdown accounting, is eligible for capture by BeginShutdown, and already has exactly one committed execution path. It may disappear only through the normal `Complete` mutation. An already captured Snapshot remains immutable.

Repeated Commit while the same Registration exists returns the same logical RegistrationID-bound bundle: the same RegistrationID, bound Completion Adapter, Owner Lifetime Lease identity, Stop-publication binding, and committed execution outcome. It does not invoke publication again, cause another Runtime-callback installation, create another lease, Stop capability, or goroutine, or change accounting. After Complete, the existing terminal-handle semantics report that the Registration was removed.

### Abort

Abort atomically removes an uncommitted reservation and wakes Wait callers when accounting may now be empty. Abort never creates lookup visibility or a committed record. Repeated Abort after Commit or Abort has no second accounting effect.

### Transport Handoff

Session Manager does not define transport handoff and never owns transport. DP-004 prepares Session, transport attachment, owner control, lease identity, Stop publication, and exactly one dormant execution path before Commit without publishing Registration or transferring ownership. Successful Commit is the sole irreversible publication point and the handoff acceptance point: before its synchronization boundary is released, the complete Registration bundle and the one-shot execution binding are committed together. A failed Commit leaves no Registration or lease, retires the dormant path through pre-Commit disposal, requires Abort, and leaves transport cleanup with Dispatcher. No recoverable post-Commit failure returns `accepted=false`.

## 8. Reserve/Commit/BeginShutdown Linearization

This proposal chooses the strict shutdown model.

At one Manager synchronization boundary:

- a Commit completed before `BeginShutdown` becomes a committed record and belongs to the shutdown snapshot; or
- `BeginShutdown` wins, Manager becomes `Closing`, and every uncommitted reservation loses the right to Commit.

New Reserve is rejected after `Closing`. A pre-existing reservation whose Commit lost must execute Abort through its current transaction owner. Complete is not invoked because no committed record exists. Manager does not forcibly abort a handle because it does not own that control flow.

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

## 11. Per-Session Execution Boundary

The detailed Execution Owner model has moved to [DP-004](DP-004-per-session-execution-boundary.md). DP-003 remains authoritative only for Reservation, Commit, RegistrationID, Complete, Lookup, Manager lifecycle, shutdown accounting, and the implemented identity-only Shutdown Snapshot.

DP-004 adds narrow integration around committed records: the complete atomic owner-scoped Commit bundle, Stop-request capability storage, and a compatible capability accessor on the shutdown snapshot. It retains the implemented Owner Lifetime Lease accounting and binds its release to the full Execution Owner terminal contract. These additions must not alter the existing meaning or linearization points of Reserve, Commit, Abort, Complete, Lookup, or BeginShutdown, and identity-only Snapshot callers retain their existing behavior.

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
- transition to `Closed` occurs automatically when reservation, committed, and Owner Lifetime Lease accounting all become empty.

### Closed

- reservation set and committed registry are empty;
- Owner Lifetime Lease accounting is empty;
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

The currently implemented shutdown snapshot contains detached immutable Registration identity only. It contains no raw Session, WebSocket, Handler, Execution Owner, callback, or mutable registry record.

DP-004 defines the remaining narrow Stop-request integration and binds the implemented owner-lifetime foundation to full Execution Owner lifetime. That integration must preserve atomic first-snapshot capture, reservation exclusion, Commit/BeginShutdown linearization, repeated BeginShutdown semantics, and the independence of the captured identity snapshot from later Complete.

## 15. Stop-Request Integration

The implemented Manager exposes no operational Stop capability. The future capability contract, its lifetime, and its relationship to shutdown orchestration are defined by DP-004. Manager must not expose raw Session, Context, Runtime, WebSocket, Send, Stop, or arbitrary callback capability through its public identity views.

## 16. Wait

`Wait(ctx)` observes the one Manager shutdown cycle. It does not initiate shutdown and does not perform Stop requests.

In the current identity-only Snapshot implementation, Wait returns nil when all three accounting sets are empty:

- reservation accounting;
- committed registry accounting;
- Owner Lifetime Lease accounting.

DP-004 retains one opaque Owner Lifetime Lease per successful execution handoff and extends Commit with the bound Stop capability and one-shot execution binding at the same irreversible point. Before Commit none is visible or eligible; after Commit none is rolled back. Full integration preserves the existing rule that Wait returns nil only when reservation accounting, committed registry accounting, and owner-lifetime accounting are all empty. The lease does not change Complete, Lookup, or Snapshot identity semantics.

The mutation that removes the final outstanding accounting item automatically changes `Closing` to `Closed` and notifies all Wait callers.

If caller context ends first:

- Wait returns `ctx.Err()` to that caller;
- Manager remains `Closing` while accounting is nonempty;
- no record is falsely removed;
- a later Wait continues observing the same cycle.

Manager has no independent operational failure model in this proposal. Session errors are neither aggregated nor retained. BeginShutdown returns only contract validation results, and Wait after `Closed` returns nil rather than an undefined terminal Manager error.

## 17. Runtime Shutdown Ordering

Manager contributes two distinct operations to Runtime shutdown: nonblocking `BeginShutdown` and context-bounded `Wait`. The complete ordering with per-Session Stop requests, Runtime cancellation, Listener shutdown, terminal observation, and owner-lifetime accounting is normative in DP-004.

DP-003 requires only that `BeginShutdown` cannot block Runtime cancellation and that `Wait` never performs Session I/O or invents completion.

## 18. Failure and Abandonment Limits

Execution panic handling and terminal-result observation are defined by DP-004. Manager itself neither executes recoverable work nor aggregates operational failures.

Manager accounting convergence is not guaranteed after:

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

Manager does not create, own, replace, or cancel root Runtime context. Host remains its sole owner under [ADR-0004](../adr/0004-handshake-runtime-dependencies.md). DP-004 defines how per-Session execution derives and owns child cancellation. Only Manager operations change Manager accounting.

## 21. Architectural Invariants

- WebSocket always has exactly one owner.
- Reservation handle has exactly one terminal Commit-or-Abort accounting effect.
- Registration Commit is the sole registry-visibility linearization point.
- Successful Commit is irreversible; only Complete removes a committed Registration.
- BeginShutdown is the sole `Open -> Closing` linearization point.
- No Commit succeeds after BeginShutdown.
- Complete is one atomic `Registered -> Removed` mutation.
- RegistrationID is never reused during Manager lifetime.
- Stale Complete cannot affect another registration.
- Manager becomes `Closed` only with empty reservation, committed, and Owner Lifetime Lease accounting; full DP-004 integration preserves this rule.
- BeginShutdown performs no I/O and Wait performs no Stop requests.
- Lookup returns immutable identity metadata only and never extends lifetime.
- Manager does not execute Session, own transport, route, deliver, publish Presence, store history, aggregate diagnostics, or apply limits.
- Per-Session execution invariants are defined by DP-004 and must not be inferred from Manager records.

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

Proposed Manager-side model:

```text
Handshake retains Upgrade ownership
    -> Session handoff reserves SessionID
    -> registration commits
    -> per-Session execution proceeds under DP-004
    -> terminal execution calls Complete
```

The current Dispatcher has no Runtime-wide registration integration. Adding Manager registration changes internal accounting and must preserve the existing transaction linearization points. DP-004 defines the separate execution migration.

## 26. Open Questions

- What future design owns configured Session limits?
- What Runtime instance identity will future Persistence combine with SessionID?
- What operational response is appropriate for an accepted permanently blocked reservation or execution?

Execution-specific open questions are maintained in DP-004. They do not change the Manager transaction and accounting invariants defined here.

## 27. Architecture Review Traceability

| Finding | Status | Resolution |
| --- | --- | --- |
| F-01 | Resolved | BeginShutdown performs the nonblocking transition; Wait is a separate context-bounded observation. |
| F-02 | Resolved | Complete is one atomic `Registered -> Removed` mutation and the sole visibility/accounting linearization point. |
| F-03 | Resolved | Reservation ownership remains continuous through Commit or Abort; the execution-side transfer is defined by DP-004. |
| F-04 | Resolved by DP-004 | Dispatcher prepares exactly one dormant launch path; the integrated Commit boundary atomically publishes Registration and its one-shot execution binding before either becomes observable. |
| F-05 | Delegated | Stop-capability lifetime is defined by DP-004; the implemented Snapshot remains identity-only. |
| F-06 | Delegated | Stop-request concurrency is defined by DP-004 and remains outside Manager execution. |
| F-07 | Clarified | Runtime shutdown orchestration is outside Manager; DP-004 defines its execution-side ordering. |
| F-08 | Resolved | RegistrationView includes never-reused RegistrationID, distinguishing SessionID reuse. |
| F-09 | Resolved | RegistrationID is unique for Manager lifetime and never reused. |
| F-10 | Resolved | Undefined Manager terminal error was removed; Wait after Closed returns nil. |
| F-11 | Clarified | Deferred completion is guaranteed only for recoverable owned-goroutine paths; nonrecoverable cases are accepted limitations. |
| F-12 | Clarified | Lookup is intentionally metadata-only and is not an operational Session capability. |
| F-13 | Resolved | Lookup returns immutable RegistrationView, never raw Session or mutable execution capability. |
| F-14 | Resolved | Normative `Completing` state was removed. |
| F-15 | Delegated | Execution Owner creation and lifetime are defined by DP-004. |
| F-16 | Clarified | SessionID is opaque, byte-for-byte, at most 255 bytes, and not automatically logged raw. |
| F-17 | Clarified | Migration explicitly acknowledges substantial internal lifecycle and concurrency changes. |
| F-18 | Clarified | Manager responsibilities are restricted by explicit negative invariants. |
| F-19 | Clarified | Exactly-once describes effective accounting mutation, not notification delivery. |
| F-20 | Accepted limitation | Permanent abandonment can retain truthful accounting and keep Manager Closing. |
| F-21 | Deferred | Limits are split into a future focused Configuration/admission capability. |
| F-22 | Resolved | BeginShutdown is the explicit nonblocking `Open -> Closing` contract. |
| F-23 | Resolved | Complete atomically removes visibility and wait-set accounting. |
| F-24 | Resolved | Reservation handle is the continuous owner before Commit or Abort. |
| F-25 | Delegated | The implemented identity Snapshot remains detached; future Stop-capability lifetime is defined by DP-004. |
| F-26 | Resolved | Manager has no terminal error aggregation; Wait returns nil or caller `ctx.Err()`. |
| F-27 | Resolved | Immutable view carries RegistrationID, so stale and reused SessionID observations are distinguishable. |

## 28. Approval Readiness

Session Manager lifecycle, registration identity, transaction linearization, lookup, shutdown accounting, and identity-snapshot semantics are fully specified here. Execution ownership findings are delegated normatively to DP-004 rather than duplicated.

Accepted limitations are explicit: process termination, unrecoverable failure, permanently blocked goroutines, or external contract violation may prevent accounting from becoming empty. Manager remains truthfully `Closing` rather than declaring false completion.

**Approval decision:** Approved.

Approval is based on the resolved and delegated findings recorded in Section 27 and the independently reviewed execution contract in approved DP-004. F-20 remains an accepted truthful-accounting limitation, and F-21 remains explicitly deferred Session-limits design; neither changes this proposal's Manager invariants. No Blocker or High architectural finding remains open.

## 29. References

- [ARCH-001: Runtime Architectural Pattern](../architecture/ARCH-001-runtime-architectural-pattern.md)
- [ARCH-002: Runtime Foundation Freeze](../architecture/ARCH-002-runtime-foundation-freeze.md)
- [DP-001: Runtime Handshake Pipeline](DP-001-runtime-handshake-pipeline.md)
- [DP-002: Runtime Host Composition Root](DP-002-runtime-host-composition-root.md)
- [DP-004: Per-Session Execution Boundary](DP-004-per-session-execution-boundary.md)
- [ADR-0004: Handshake Runtime Dependency Boundary](../adr/0004-handshake-runtime-dependencies.md)
- [Master Engineering Plan](../roadmap/MASTER_PLAN.md)

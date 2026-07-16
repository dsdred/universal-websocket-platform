# DP-003: Runtime Session Manager

[Russian version](../../ru/design/DP-003-runtime-session-manager.md)

## 1. Status

**Status:** Draft

This revision simplifies the lifecycle after the second independent architecture review. It separates shutdown transition from waiting, reduces completion to one atomic mutation, defines continuous ownership for every registration transaction, and specifies the conceptual Execution Owner contract required for implementation. Exact Go names remain implementation details; the production boundary, ownership, ordering, and concurrency semantics are normative.

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
- Define one per-Session Execution Owner, its launch and terminal-completion contract, and a stable non-owning Stop-request capability.
- Keep Manager limited to registry and shutdown accounting.

## 4. Non Goals

This document does not design Router, Delivery, Presence, Groups, Persistence, cluster ownership, Plugins, Metrics, a diagnostics framework, Session limits policy, transport protocol, public API, Configuration fields, a generic Session framework, or a supervision framework.

## 5. Ownership Boundaries

Four ownership concerns remain separate.

### WebSocket Transport Ownership

Upgrade boundary owns WebSocket and child connection context until Session explicitly accepts them. Before acceptance, Upgrade boundary is the sole closer. After acceptance, Session is the sole closer, including when later registration fails.

### Registration Transaction Ownership

Successful Reserve returns one conceptual reservation handle. The Session handoff/Dispatcher control flow owns that handle until it transfers the handle to the per-Session Execution Owner. The Execution Owner then owns it until exactly one terminal operation:

- `Commit`, which consumes the reservation and creates a committed record; or
- `Abort`, which removes the reservation without registry visibility.

The normative control-flow pattern is:

```text
reservation := Manager.Reserve(SessionID)
defer reservation.AbortUnlessCommitted()
```

This notation describes an ownership obligation, not a required Go API. Every recoverable return, rejection, error, or panic before Commit must execute Abort. Manager accounts for the reservation but does not own the goroutine responsible for Commit or Abort. Handshake never owns the reservation and does not import Session Manager; it delegates post-Upgrade preparation to the Session handoff boundary.

### Session Execution Ownership

A narrow per-Session Execution Owner is a separate production object created by the Session handoff/Dispatcher after Session construction and successful Reserve, but before registration Commit. The Dispatcher is its only creator in the production composition path. It transfers the reservation handle, one Session, a Runtime-derived execution context, a narrow completion capability, and a narrow terminal-result observer to the owner.

The owner is inactive before Commit and owns no committed accounting. It owns the reservation transaction after transfer. Successful Commit is the single point at which it acquires the committed RegistrationID and activates terminal completion. The immediately following serialized owner activation is the single execution-ownership point: it transfers Session execution and transport-cleanup responsibility before a successful handoff result is observable. The owner then owns Session `Start/Run/Stop` in one owned goroutine until terminal cleanup and completion accounting finish.

The Execution Owner depends only on the Session lifecycle contract, opaque registration identity, its owned context, and the narrow completion and terminal-observer contracts. It must not import or retain concrete Manager, Runtime Host, Listener, Handshake, HTTP, WebSocket, Shutdown Snapshot, Router, Delivery, Presence, or Persistence. It uses no service locator, `context.Value` dependency injection, generic supervisor, or generic worker framework.

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
Upgrade boundary calls Session handoff/Dispatcher
    -> create provisional Session without transferring transport ownership
    -> Reserve SessionID
    -> create inactive per-Session Execution Owner
    -> transfer Reservation ownership to Execution Owner
    -> owned goroutine installs terminal guard
    -> Commit registration with stable Stop-request capability
    -> acquire RegistrationID and activate completion obligation
    -> accept Session execution and transport-cleanup ownership
    -> report successful handoff
    -> Start
    -> Run
    -> Stop
    -> Complete
    -> report one terminal result
```

### Reserve

Reserve is allowed only while Manager is `Open`. It validates SessionID and creates:

- a reservation handle with exclusive Commit-or-Abort obligation;
- a RegistrationID unique for the lifetime of this Manager.

Reservation is counted for Manager closure but is invisible to Lookup and is not a committed Session in the shutdown registry.

### Commit

Commit consumes its reservation and atomically creates one committed record containing immutable identity metadata and a stable Stop-request capability. This mutation is the single registration linearization point: registry visibility and committed wait-set membership appear together. A successful Commit returns the opaque RegistrationID to the same owned goroutine. The owner activates its completion obligation before Start or a successful handoff result can become observable.

### Abort

Abort atomically removes an uncommitted reservation and wakes Wait callers when accounting may now be empty. Abort never creates lookup visibility or a committed record. Repeated Abort after Commit or Abort has no second accounting effect.

### Transport Handoff

Session construction is provisional and does not transfer transport ownership. The Upgrade boundary remains the sole transport closer through Session construction, Reserve, Execution Owner construction, and a failed Commit. On any such pre-Commit failure, the Execution Owner performs Abort, performs no Complete, and returns a failed handoff; the Upgrade boundary closes the transport.

Successful Commit activates the owner's terminal completion obligation. The controlled handoff then performs one serialized owner activation that transfers Session execution and transport-cleanup ownership to the Execution Owner/Session boundary before the Dispatcher reports successful acceptance. A failure after Commit but before this activation performs Complete while the Upgrade boundary still closes the transport. From activation onward, every failure is owner-managed: the owner stops Session, performs Complete, and prevents the Upgrade boundary from closing the transport again. The state “transport accepted but Commit failed” is intentionally unreachable. There is always exactly one transport owner.

## 8. Reserve/Commit/BeginShutdown Linearization

This proposal chooses the strict shutdown model.

At one Manager synchronization boundary:

- a Commit completed before `BeginShutdown` becomes a committed record and belongs to the shutdown snapshot; or
- `BeginShutdown` wins, Manager becomes `Closing`, and every uncommitted reservation loses the right to Commit.

New Reserve is rejected after `Closing`. A pre-existing reservation whose Commit lost must execute Abort through its Execution Owner. Transport ownership has not transferred at that point, so Complete and Session Stop are not invoked; the Upgrade boundary performs transport cleanup. Manager does not forcibly abort a handle because it does not own that control flow.

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

The Session handoff/Dispatcher creates exactly one Execution Owner for one provisional Session and transfers one Reservation handle to it. The owner creates exactly one owned execution goroutine. The handoff call waits only until that goroutine reports either pre-Commit failure or successful committed ownership acceptance; it does not wait for Session termination.

Before calling Commit, the owned goroutine installs one terminal guard. Before Commit succeeds, that guard may only Abort the Reservation. The successful Commit return is the owner's pre-Commit-to-committed linearization point: it supplies RegistrationID, makes Abort a no-op obligation, and activates terminal Complete. A subsequent serialized activation is the execution-ownership linearization point. No Start or successful ownership acceptance is observable before activation.

The normal committed algorithm is normative:

```text
activate completion obligation
    -> accept execution and transport-cleanup ownership
    -> Start(Session, execution context)
    -> if Start succeeds and no Stop request won, Run(Session, execution context)
    -> attempt Stop(Session, owner-controlled cleanup context)
    -> invoke Complete(RegistrationID) once
    -> publish one terminal result to the injected observer
```

Stop is attempted once on every committed terminal path, including Start error, Run return, Run error, cancellation, and recovered panic. A Stop error is retained in the terminal result but never suppresses Complete. Complete is attempted after the Stop attempt; a permanently blocked Stop may therefore prevent completion and remains an accepted abandonment limitation. The cleanup context is owned by the Execution Owner and is not the already-canceled execution context.

The owner remains alive until its execution reached a recoverable terminal path, its single Complete invocation returned, its terminal result was offered once to the observer, and every Stop-request invocation that linearized before terminal completion returned. The terminal observer is a narrow non-owning sink supplied through Runtime composition to the Dispatcher; it does not receive Session or transport. Observer failure or panic is isolated and cannot prevent the Complete attempt. The concrete terminal-result type and diagnostics backend remain implementation details, but Start, Run, Stop, recovered-panic, and completion-anomaly categories must remain distinguishable.

### Stop Before Start

The stable Stop-request capability records one termination request and cancels the owner-controlled execution context. If that request linearizes before the owner marks Start as begun, Start and Run are skipped. The owned goroutine calls Session Stop from `Created`, then performs Complete. There is no alternative “Start may still run” behavior.

### Concurrent Start and Stop

The Execution Owner, not Session Manager, serializes its control state. The Start linearization point is the owner's transition that marks Start as begun after observing no accepted Stop request. The Stop-request linearization point is the first atomic recording of termination intent.

- If Stop wins, Start and Run are forbidden.
- If Start wins, the request cancels the execution context; after Start returns, Run begins only if no Stop request is pending and the context remains active.
- If Run is active, cancellation causes it to return according to the Session contract; the owner then invokes Stop.
- Repeated and concurrent Stop requests do not create additional Stop or Complete obligations.
- No lifecycle lock is held while Start, Run, Stop, Complete, or the terminal observer executes.

### RegistrationID and Completion Capability

An Execution Owner may exist before it has a committed RegistrationID, but only in its inactive pre-Commit phase. The owned goroutine receives RegistrationID only from successful Commit. Failed Commit leaves the owner without an active completion obligation: it aborts the Reservation, reports failed handoff, and never calls Complete.

The owner depends on a consumer-oriented completion capability with the conceptual operation `Complete(RegistrationID) bool`, not on concrete Manager. Runtime composition supplies a Manager-backed implementation to the Session handoff/Dispatcher, which injects it into each owner. The capability becomes active only after successful Commit. Each owner invokes it once. `true` means this invocation performed the effective `Registered -> Removed` mutation; `false` means no known committed record was mutated and is reported as a terminal accounting anomaly. It is never retried. Manager still guarantees at most one effective mutation for the RegistrationID.

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

The target shutdown snapshot contains immutable Registration identity and one stable non-owning Stop-request capability for each committed record. It contains no raw Session, WebSocket, Handler, Execution Owner, or mutable registry record. The currently implemented identity-only Snapshot intentionally remains unchanged until the Execution Owner and capability integration task.

Removal by Complete does not invalidate a capability already issued for the current shutdown orchestration. Its lifetime continues until the corresponding Stop request returns. Repeated BeginShutdown does not create a new cycle or duplicate accounting.

## 15. Stop-Request Capabilities

A Stop-request capability represents permission to request termination from one Execution Owner without acquiring Session ownership. Its conceptual operation is `RequestStop() bool`; it accepts no context because it is strictly nonblocking and performs no Session I/O in the caller.

The first call that linearizes before terminal completion records termination intent, cancels the owner-controlled execution context once, and returns `true`. Repeated or concurrent calls, and calls that linearize after terminal completion, are stable no-ops that return `false`. The operation does not wait for Start, Run, Stop, Complete, or full Session termination and returns no Session result.

The capability never exposes Session, Send, Context cancellation, Runtime, Listener, WebSocket, callback registration, or mutable owner state. Stop requests for different registrations are independent and cannot serialize through Manager. Capability invocation does not acquire a Manager lifecycle lock.

Each invocation holds the control cell alive until that invocation returns. Complete may run while a previously entered RequestStop is returning, but it cannot invalidate that invocation. At terminal completion the control cell detaches from Session execution state; later calls remain safe `false` no-ops and do not retain or reacquire Session ownership. No new Stop request is accepted after terminal completion.

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

Execution Owner recovers a panic that occurs inside its owned execution boundary. Recovery does not classify the execution as successful and does not re-panic from the owned goroutine. The terminal guard cancels the execution context, attempts Session Stop if committed ownership was activated, invokes Complete once when RegistrationID is committed, and publishes one sanitized panic-category terminal result to the injected observer. Before successful Commit, the same guard performs Abort and does not call Complete or Session Stop because transport ownership has not transferred.

A panic or error from Stop or the terminal observer cannot suppress the already-established Complete attempt; each terminal step is guarded independently. The observer receives no panic payload that could expose credentials or transport metadata. This rule defines ownership cleanup only and does not create a diagnostics framework.

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

Manager does not create, own, replace, or cancel root Runtime context. Host remains its sole owner under [ADR-0004](../adr/0004-handshake-runtime-dependencies.md). Session handoff derives each owner's execution context from the Runtime-owned connection context. The owner controls only its child cancellation; root cancellation and an accepted Stop request both cancel execution. Only Complete changes Manager accounting.

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
- Session handoff/Dispatcher is the sole production creator of one Execution Owner per committed Session.
- Successful Commit activates exactly one owner completion obligation before Start or successful handoff is observable.
- The Execution Owner is the sole caller of Session Start, Run, and Stop after committed ownership acceptance.
- A Stop request that wins before Start forbids Start and Run.
- One owned goroutine performs one Stop attempt and one Complete invocation on every recoverable committed terminal path.
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
Handshake retains Upgrade ownership
    -> Session handoff/Dispatcher creates provisional Session
    -> Dispatcher reserves SessionID
    -> Dispatcher creates per-Session Execution Owner
    -> owner receives Reservation and installs terminal guard
    -> registration commits with stable Stop-request capability
    -> owner accepts execution and transport-cleanup ownership
    -> Dispatcher reports successful handoff
    -> owner goroutine performs Start/Run/Stop and Complete
```

The current Dispatcher responsibilities change explicitly: it keeps Session construction and handoff coordination, gains Reserve and owner creation, and delegates all post-Commit Start/Run/Stop execution and terminal Complete to the per-Session owner. It no longer blocks synchronously in Session Run and no longer calls Session Stop directly after successful ownership transfer.

External WebSocket behavior remains unchanged, but internal concurrency, ownership, registration, and shutdown semantics change. Migration must therefore be treated as lifecycle work, not as a behavior-neutral refactor, and must be race-reviewed at every boundary.

## 26. Open Questions

- What exact Go names and package-private interfaces represent the normative consumer-oriented capabilities?
- What terminal-result value shape preserves phase and cleanup errors for the narrow observer?
- What cleanup deadline policy should bound Session Stop without changing existing Session semantics?
- What future design owns configured Session limits?
- What Runtime instance identity will future Persistence combine with SessionID?
- What operational response is appropriate for an accepted permanently blocked reservation or execution?

The creator, ownership start, launch model, pre-Commit cleanup, Stop-before-Start behavior, concurrent Start/Stop ordering, completion dependency, panic cleanup, Stop-request operation, and owner lifetime are no longer open questions. Remaining questions are implementation-level representation or explicitly separate subsystem policy; they do not change the linearization points or Manager accounting defined here.

## 27. Architecture Review Traceability

| Finding | Status | Resolution |
| --- | --- | --- |
| F-01 | Resolved | BeginShutdown performs the nonblocking transition; Wait is a separate context-bounded observation. |
| F-02 | Resolved | Complete is one atomic `Registered -> Removed` mutation and the sole visibility/accounting linearization point. |
| F-03 | Resolved | Reservation handle moves once from Session handoff/Dispatcher to Execution Owner and retains continuous Commit-or-Abort ownership. |
| F-04 | Resolved | Execution Owner exists before Commit; shutdown between Commit and Start is a normal terminal path ending in Stop and Complete. |
| F-05 | Resolved | Shutdown snapshot contains stable Stop capabilities whose invocation lifetime survives Complete. |
| F-06 | Resolved | Stop requests are independent and strictly nonblocking; Manager invokes none of them. |
| F-07 | Resolved | Runtime shutdown orchestration owns Stop requests; Manager progresses through Abort/Complete mutations without a waiting caller. |
| F-08 | Resolved | RegistrationView includes never-reused RegistrationID, distinguishing SessionID reuse. |
| F-09 | Resolved | RegistrationID is unique for Manager lifetime and never reused. |
| F-10 | Resolved | Undefined Manager terminal error was removed; Wait after Closed returns nil. |
| F-11 | Clarified | Deferred completion is guaranteed only for recoverable owned-goroutine paths; nonrecoverable cases are accepted limitations. |
| F-12 | Clarified | Lookup is intentionally metadata-only and is not an operational Session capability. |
| F-13 | Resolved | Lookup returns immutable RegistrationView, never raw Session or mutable execution capability. |
| F-14 | Resolved | Normative `Completing` state was removed. |
| F-15 | Resolved | Execution Owner is created by Session handoff before Commit and survives completion plus outstanding Stop invocations. |
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

## 28. TASK-B4-007B Blocker Traceability

| Finding | Status | Resolution |
| --- | --- | --- |
| B-01 — Creator boundary | Resolved | Session handoff/Dispatcher is the sole production creator of one per-Session Execution Owner. |
| B-02 — Ownership start | Resolved | Successful Commit activates completion; the immediately following serialized owner activation is the single execution-ownership point before Start or successful handoff is observable. |
| B-03 — Pre-Commit cleanup | Resolved | Owner Abort resolves Reservation; Upgrade boundary retains and closes transport; Stop and Complete are not invoked. |
| B-04 — Launch model | Resolved | One owned goroutine performs Commit activation and the normative Start/Run/Stop lifecycle; handoff waits only for commit/acceptance outcome. |
| B-05 — Completion activation | Resolved | Successful Commit returns RegistrationID and activates exactly one terminal Complete obligation; failed Commit never activates it. |
| B-06 — Panic semantics | Resolved | Recoverable owned-boundary panic is recovered, cleanup and Complete are attempted, one sanitized terminal result is observed, and panic is not rethrown. |
| B-07 — Stop before Start | Resolved | A Stop request winning before Start causes Stop from Created and forbids Start and Run. |
| B-08 — Concurrent Start/Stop | Resolved | Owner serializes control state; first Stop records termination once, Start wins only through its explicit linearization point, and Run cannot begin after accepted Stop. |
| B-09 — RegistrationID handoff | Resolved | Owner exists inactive without RegistrationID and receives the opaque identity only from successful Commit. |
| B-10 — Completion capability | Resolved | Owner uses injected `Complete(RegistrationID) bool` semantics and never depends on concrete Manager. |
| B-11 — Stop-request capability | Resolved | `RequestStop() bool` is nonblocking, idempotent, context-free, Session-free, and has stable terminal no-op behavior. |
| B-12 — Owner lifetime | Resolved | Owner lasts through terminal cleanup, one Complete invocation, one terminal observation, and all Stop invocations entered before terminal completion. |

No TASK-B4-007B blocker finding is Deferred.

## 29. Approval Readiness

Second-review Blockers F-01 through F-03 are resolved by split shutdown contracts, atomic Complete, and continuously owned reservation handles. High findings F-04 through F-10 are resolved through execution-owner lifetime, stable Stop capabilities, independent shutdown orchestration, identity-safe views, and removal of undefined Manager terminal errors.

Accepted limitations are explicit: process termination, unrecoverable failure, permanently blocked goroutines, or external contract violation may prevent accounting from becoming empty. Manager remains truthfully `Closing` rather than declaring false completion.

All TASK-B4-007B blocker decisions are resolved normatively. Exact Go names, private package placement within the Session handoff subsystem, terminal-result representation, and cleanup deadline mechanics remain implementation-level choices constrained by this contract. Configured limits, durable Runtime identity, and operational handling of permanently blocked work remain separate subsystem questions and do not alter this lifecycle.

**Approval decision candidate:** Approved with Findings.

The document remains Draft and the Execution Owner contract is ready for targeted independent review. That review must confirm the creator boundary, commit-to-ownership transition, owned-goroutine launch, Stop ordering, completion capability, recoverable-panic cleanup, and stable capability lifetime.

## 30. References

- [ARCH-001: Runtime Architectural Pattern](../architecture/ARCH-001-runtime-architectural-pattern.md)
- [ARCH-002: Runtime Foundation Freeze](../architecture/ARCH-002-runtime-foundation-freeze.md)
- [DP-001: Runtime Handshake Pipeline](DP-001-runtime-handshake-pipeline.md)
- [DP-002: Runtime Host Composition Root](DP-002-runtime-host-composition-root.md)
- [ADR-0004: Handshake Runtime Dependency Boundary](../adr/0004-handshake-runtime-dependencies.md)
- [Master Engineering Plan](../roadmap/MASTER_PLAN.md)

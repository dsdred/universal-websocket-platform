# DP-003: Runtime Session Manager

[Russian version](../../ru/design/DP-003-runtime-session-manager.md)

## 1. Status

**Status:** Draft

## 2. Problem Statement

The Runtime can create and run a Session after a successful Handshake, but it has no component that owns the set of live Sessions. A Session is currently executed synchronously by the handoff path and is not registered in a Runtime-wide shutdown wait set. Runtime therefore has no single place to answer whether a Session is still active, prevent an admitted connection from escaping shutdown tracking, or wait for every handed-off Session to finish.

[DP-001](DP-001-runtime-handshake-pipeline.md) requires a Session to enter the Runtime shutdown wait set before ownership handoff completes. [ARCH-002](../architecture/ARCH-002-runtime-foundation-freeze.md) deliberately leaves Session ownership in that wait set open. This proposal defines that missing boundary without changing the frozen Runtime Host, Admission Gate, Runtime context, or startup transaction.

## 3. Goals

- Define one lifecycle owner for the set of Sessions in one Runtime instance.
- Make registration the exact boundary at which a Session becomes Runtime-visible and joins the shutdown wait set.
- Make removal deterministic and exactly once.
- Support lookup by the minimum stable identity required by the Runtime.
- Coordinate graceful Runtime shutdown without taking transport ownership away from Session.
- Preserve explicit ownership, bounded lifecycle, and ordinary dependency injection.

## 4. Non Goals

This document does not design:

- Router or message dispatch policy;
- Delivery, queues, retries, or backpressure;
- Persistence or Session recovery;
- Presence, Groups, Topics, or Broadcast;
- cluster membership, federation, or distributed Session ownership;
- Plugin SDK or Plugin ABI;
- Metrics or an operational diagnostics framework;
- public Control Plane or HTTP API;
- concrete Go interfaces or storage structures.

Those systems may integrate with Session lifecycle later, but they do not belong to Session Manager.

## 5. Responsibilities

Session Manager is responsible only for:

- accepting lifecycle ownership of a successfully constructed Session;
- registering and deregistering that Session;
- maintaining the authoritative set of registered Sessions for one Runtime instance;
- exact lookup of a registered Session by its stable SessionID;
- coordinating Session closure and completion during Runtime shutdown;
- enforcing the Runtime Session lifecycle invariants defined here.

Session Manager does not authenticate clients, accept WebSockets, read messages, send messages, route traffic, persist state, evaluate policy, or own the root Runtime context.

## 6. Ownership Model

The ownership chain is:

```text
Runtime Host
    -> Session Manager
        -> Session
            -> WebSocket connection
```

The chain describes lifecycle authority, not shared transport ownership:

- Runtime Host composes Session Manager and invokes its shutdown coordination. Host does not store or close individual Sessions.
- Session Manager owns the authoritative membership and completion tracking of registered Sessions. It may request that a Session stop, but it does not close the WebSocket directly.
- Session exclusively owns its WebSocket after the handoff defined by DP-001. Session performs the actual transport close and releases its per-connection resources.

Before registration, the Upgrade boundary owns the candidate Session, derived connection context, and WebSocket. Successful registration is one atomic conceptual linearization point: the Session enters the Manager's registry and shutdown wait set, and Manager accepts lifecycle ownership. Only then may the Upgrade boundary report successful handoff.

If registration fails, ownership does not transfer and the Upgrade boundary closes the connection. After registration succeeds, an upper layer must not close the WebSocket or deregister the Session. Session completion follows the Manager-owned removal path.

## 7. Runtime Session Lifecycle

The conceptual lifecycle is:

```text
Created
    -> Registering
    -> Registered
    -> Running
    -> Closing
    -> Closed
    -> Removed
```

These are architectural states, not a required Go enum.

- **Created:** the Upgrade boundary has constructed a Session candidate and still owns it.
- **Registering:** Manager is attempting the atomic ownership and membership transition. This state is not externally visible through lookup.
- **Registered:** ownership has transferred, the Session is visible to Runtime lookup, and it belongs to the shutdown wait set.
- **Running:** Session executes its connection and message lifecycle. It remains registered.
- **Closing:** Session is terminating because of peer closure, an error, its own lifecycle, or Runtime shutdown. It remains in the wait set.
- **Closed:** Session transport and per-connection work have finished. It remains tracked only until removal completes.
- **Removed:** Manager has removed it exactly once. It is no longer visible and no longer belongs to the wait set.

A failure after successful registration does not bypass `Closing`, `Closed`, or `Removed`. A failure before registration leaves cleanup with the Upgrade boundary. Restarting or re-registering a removed Session is not supported.

## 8. Registration

Registration is the single operation that makes a Session exist in the Runtime management domain. It must atomically establish all of the following:

- SessionID is valid and not already registered;
- Session becomes visible through lookup;
- Session joins the Runtime shutdown wait set;
- Session Manager accepts lifecycle ownership;
- a concurrent shutdown will either observe the Session or reject registration.

No successful handoff may refer to an unregistered Session. Registration is accepted at most once for a SessionID during the lifetime of one Manager. Registration after Manager has closed admission for shutdown is rejected; the Upgrade boundary remains responsible for cleanup.

Registration must not start network I/O while holding Manager synchronization. The Manager must not call arbitrary Session behavior under its registry lock.

## 9. Removal

Removal occurs only after Session transport closure and Session execution have completed. Manager owns the removal transition, even when completion was caused by the peer or by a Session error.

Removal must:

- happen exactly once;
- remove all lookup visibility atomically;
- remove the Session from the shutdown wait set;
- wake shutdown coordination waiting for the set to become empty;
- preserve the original Session termination result for the future diagnostics boundary without turning the registry into that boundary.

Session must not mutate Manager storage directly. A Manager-owned execution/completion path observes terminal Session completion and performs removal. Repeated completion notifications are harmless and cannot decrement tracking twice. No tombstone is required in the first implementation; after removal, lookup reports absence.

## 10. Shutdown Coordination

Session Manager participates in, but does not own, Runtime shutdown. The frozen Host remains the lifecycle coordinator and owns Admission Gate closure and root Runtime context cancellation.

The conceptual shutdown sequence is:

```text
Host closes Admission Gate
    -> no new Handshake admission commit may begin
    -> Session Manager closes registration
    -> in-flight registration resolves as accepted-and-tracked or rejected
    -> Runtime context cancellation reaches Sessions
    -> Manager requests closure of every registered Session
    -> Manager waits until every registered Session is Removed
    -> Runtime Stop may complete
```

The shutdown wait set is exactly the set of Sessions for which registration succeeded and removal has not completed. A Session enters the set at the registration linearization point, before handoff succeeds, and leaves only at removal.

Closing Manager registration and registering a Session must be mutually ordered. A registration linearized before closure is included and awaited. A registration linearized after closure fails and remains owned by the Upgrade boundary. This rule closes the gap between the final Admission Gate check and ownership handoff without moving the Gate into Session Manager.

Shutdown waiting must honor the caller's shutdown boundary when that policy is defined. This proposal does not select forced-close escalation or timeout policy. Regardless of timeout outcome, Manager must not falsely report an unremoved Session as removed.

## 11. Lookup

The mandatory first lookup key is **SessionID**. It is stable, unique within one Runtime instance, already belongs to Session metadata, and does not introduce identity or routing semantics.

Lookup observes only registered, not-yet-removed Sessions. It must not expose Manager's mutable registry or transfer Session ownership. A caller receives only the capability needed by its future integration and remains unable to deregister or close transport behind Manager's lifecycle.

The following keys are deferred:

- **ConnectionID:** no separate stable Runtime meaning is currently established.
- **Principal ID:** one Principal may own multiple Sessions and anonymous Sessions need explicit semantics.
- **UserID:** no User domain exists in the Runtime contracts.
- enumeration and secondary indexes: requirements depend on Delivery, Presence, and operational use cases.

These deferred lookup paths require focused design when real consumers exist.

## 12. Runtime Context

Session Manager does not create, replace, or cancel the root Runtime context. Host remains its sole owner under [ADR-0004](../adr/0004-handshake-runtime-dependencies.md).

The Upgrade boundary creates a connection context derived from the active Runtime context. Before registration, its cancellation belongs to the Upgrade boundary. At successful registration and ownership handoff, per-connection cancellation becomes part of Session lifecycle. Manager tracks and coordinates that lifecycle but never receives the root cancellation function.

Root Runtime cancellation signals every derived Session context during shutdown. Manager still waits for actual Session completion and removal; context cancellation alone is not evidence that cleanup has finished.

## 13. Architectural Invariants

- Every handed-off Session is registered in exactly one Session Manager.
- Every registered Session belongs to the Runtime shutdown wait set.
- Successful registration is the single lifecycle ownership transfer point.
- Session is the sole owner and closer of its WebSocket after transfer.
- A SessionID is registered at most once during one Manager lifetime.
- Removal happens exactly once and only after Session completion.
- Lookup never returns a removed or not-yet-registered Session.
- Closing registration is ordered atomically with concurrent registration.
- Shutdown cannot complete while the shutdown wait set is non-empty.
- Manager does not hold registry synchronization while invoking Session network or lifecycle work.
- Manager does not own Authentication, routing, message handling, or Runtime root context cancellation.
- Host does not become a Session registry or close individual Session transports.
- Failures cannot silently abandon a registered Session outside both lookup and shutdown tracking.

## 14. Future Integration

Future systems integrate at the Session Manager boundary without changing ownership:

- **Router** may target a currently registered Session but does not own its lifecycle.
- **Delivery** may resolve a target by a future lookup/index contract; queueing and retry semantics remain separate.
- **Presence** may derive state from registration and removal events; Presence is not stored by Manager.
- **Groups** may index SessionIDs without owning Sessions or transport.
- **Persistence** may record durable metadata or events; a persisted record never becomes ownership of a live Session.

This section names integration seams only. It does not design those subsystems or their APIs.

## 15. Alternatives

### Alternative A — No Session Manager

**Advantages:** preserves the current small handoff path and adds no component.

**Disadvantages:** there is no authoritative live set, lookup, exactly-once removal, or complete shutdown wait set. DP-001 handoff requirements remain unsatisfied.

**Decision:** rejected because Runtime cannot prove ownership or shutdown completion.

### Alternative B — Runtime Host Owns Sessions Directly

**Advantages:** fewer top-level components and direct access during Stop.

**Disadvantages:** Host becomes a mutable Session registry and gains per-connection lifecycle logic. This conflicts with its frozen composition and lifecycle-coordination role and moves it toward a god object.

**Decision:** rejected. Host composes and coordinates Manager instead.

### Alternative C — Listener Owns Sessions

**Advantages:** Listener is near accepted connections and transport shutdown.

**Disadvantages:** it couples HTTP/WebSocket transport to Runtime Session identity, lookup, and future consumers. It also obscures the ownership boundary between Upgrade and Session.

**Decision:** rejected because Listener must remain a transport component.

### Alternative D — Distributed Session Ownership

**Advantages:** could provide cross-instance lookup and coordination for a cluster.

**Disadvantages:** requires membership, failure detection, consistency, remote addressing, and partition semantics before the single-instance lifecycle is settled.

**Decision:** rejected for the current Runtime. Cluster and federation are deferred beyond this proposal.

### Alternative E — Separate Registry and Manager

**Advantages:** separates indexing from lifecycle coordination and could allow specialized indexes later.

**Disadvantages:** the first implementation would split the atomic registration/removal invariant across two owners without an independent use case. It increases coordination and failure modes.

**Decision:** rejected for now. Extraction may be reconsidered after multiple real lookup or indexing consumers exist.

## 16. Explicitly Out of Scope

- Router, Delivery, queues, backpressure, and persistence;
- Presence, Groups, Topics, Broadcast, and fan-out;
- cluster, federation, horizontal scaling, and remote lookup;
- Plugin contracts or extension loading;
- Metrics, tracing, logging policy, and diagnostics sinks;
- Authentication and Authorization decisions;
- Session transport protocol and message loop behavior;
- public API, Configuration DSL additions, and database schema;
- concrete concurrency primitives or Go method signatures.

## 17. Migration Strategy

Migration is incremental and must preserve current behavior:

1. Compose one Session Manager per Runtime instance through Runtime Host without changing frozen Host lifecycle states.
2. Route successful post-Upgrade Session construction through Manager registration.
3. Make registration success the DP-001 handoff acceptance point; retain Upgrade-boundary cleanup on failure.
4. Run existing Session lifecycle under Manager completion tracking and exactly-once removal.
5. Include Manager's wait set in Runtime shutdown coordination after Admission Gate closure.
6. Add SessionID lookup only when registration, removal, and shutdown tests prove the ownership invariants.

No migration step changes Authentication, Listener transport behavior, Message handling, or Configuration Snapshot. Each step must retain a single closer for the WebSocket and must be independently testable under concurrent handoff and shutdown.

## 18. Open Questions

- What bounded escalation applies when a Session does not finish before the shutdown caller's deadline?
- How should a Session start failure after successful registration be categorized for operational diagnostics?
- Does a future ConnectionID have semantics distinct from SessionID?
- Which capability shape should a lookup return without exposing lifecycle mutation?
- When real consumers require enumeration, what consistency guarantee is necessary during concurrent removal?
- Which later component publishes Session lifecycle events without making Manager a diagnostics framework?

These questions do not block agreement on ownership, registration, removal, or the shutdown wait-set invariant. They require separate implementation design or focused proposals when their consumers are known.

## 19. Review Readiness

This Draft is ready for architecture review when reviewers can verify that:

- registration is an unambiguous ownership and shutdown-tracking linearization point;
- every successful DP-001 handoff is either tracked or cleaned up by the Upgrade boundary;
- Session remains the sole WebSocket owner after handoff;
- shutdown cannot miss a concurrent registration;
- removal and lookup rules are deterministic under concurrency;
- Manager does not absorb Host, Listener, Authentication, Router, Delivery, or diagnostics responsibilities;
- migration can occur without changing frozen Runtime Foundation semantics.

Implementation should not begin until the ownership transfer, shutdown ordering, and minimum lookup decision receive architecture approval.

## 20. References

- [ARCH-001: Runtime Architectural Pattern](../architecture/ARCH-001-runtime-architectural-pattern.md)
- [ARCH-002: Runtime Foundation Freeze](../architecture/ARCH-002-runtime-foundation-freeze.md)
- [DP-001: Runtime Handshake Pipeline](DP-001-runtime-handshake-pipeline.md)
- [DP-002: Runtime Host Composition Root](DP-002-runtime-host-composition-root.md)
- [ADR-0004: Handshake Runtime Dependency Boundary](../adr/0004-handshake-runtime-dependencies.md)
- [Master Engineering Plan](../roadmap/MASTER_PLAN.md)


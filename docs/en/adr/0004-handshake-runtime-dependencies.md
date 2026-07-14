# ADR 0004: Handshake Runtime Dependency Boundary

[Russian version](../../ru/adr/0004-handshake-runtime-dependencies.md)

## Status

Accepted.

## Context

The Runtime foundation now has a Host-owned Admission Gate and a Host-owned Runtime context. [ARCH-002](../architecture/ARCH-002-runtime-foundation-freeze.md) freezes their ownership and lifecycle semantics. [DP-001](../design/DP-001-runtime-handshake-pipeline.md) requires pre-Upgrade Authentication, a final admission check before `websocket.Accept`, and a Session context whose lifetime does not depend on `http.Request.Context()`.

The current composition cannot satisfy those requirements without an explicit dependency boundary:

- Handshake does not receive the Host-owned admission permission;
- Handshake does not receive the active Runtime context;
- passing the concrete Host would invert dependency direction and expose unrelated lifecycle operations;
- creating another Gate or Runtime context in Listener would duplicate mutable state and ownership;
- Listener must not import `internal/runtime`, and Runtime must remain the composition root rather than a dependency of transport code.

The decision must connect the frozen Runtime foundation to the open Handshake architecture without changing Host lifecycle semantics, exposing cancellation, or introducing a general dependency container.

## Decision

Runtime composition passes two minimal, live, read-only Runtime capabilities to Handshake through explicit constructor injection:

1. **Admission permission capability** — observes whether the Host-owned Admission Gate currently permits entry into admission commit.
2. **Runtime context capability** — returns the currently published Host-owned Runtime context without exposing its cancellation function.

The contracts are consumer-oriented and narrow. Handshake depends on capability semantics, not on the concrete Host, Runtime Container, lifecycle state enum, Admission Gate implementation, or context holder implementation. The exact Go names and package placement are implementation details, provided the dependency direction remains acyclic and Listener does not import `internal/runtime`.

The capabilities are live views over Host-owned state. They do not copy the Gate state or Runtime context into Handshake. Handshake cannot open or close admission, publish or replace the Runtime context, cancel the root context, start or stop Host, or locate any other Runtime dependency.

### Admission Semantics

Handshake consults admission at two points:

1. before Authentication, so a closed Runtime does not invoke Authentication;
2. after an Allow decision, as the last operation before `websocket.Accept` begins.

The successful final check is the linearization point for entry into the conceptual `Committing` state. If Gate closes before that check, Upgrade does not start. If Gate closes after that successful check, the commit is already in progress and follows DP-001 ownership and shutdown rules: it must complete handoff into Runtime tracking or close the upgraded connection before shutdown completes.

An Allow decision is not cached as admission permission. Cancellation or Gate closure before the final check invalidates the decision for execution.

### Runtime Context Semantics

The Runtime context capability returns no active context before successful startup commit. Handshake must not cache an unavailable result during construction.

After successful admission and Upgrade, the Upgrade boundary derives a connection/Session context from the active Runtime context. It does not derive Session lifetime from `http.Request.Context()`. Host remains the sole owner of root Runtime context cancellation. A derived connection context may have its own cancellation function for connection termination; before handoff that function belongs to the Upgrade boundary, and after successful ownership acceptance it belongs to Session.

If no active Runtime context is available at commit time, Handshake fails closed and does not create a Session.

## Composition Order

Composition follows this order without a Runtime-to-Listener cycle:

```text
Snapshot
    -> Host-owned lifecycle primitives
    -> read-only Runtime capabilities
    -> Authentication
    -> Session handoff
    -> Handshake
    -> Listener
    -> Host startup transaction
```

The lifecycle primitives, stable capability implementations, and inactive Runtime context holder are created when Host is created. `Build` does not activate them, open admission, publish a Runtime context, or start network resources.

During `Start`, Runtime composition injects the same stable capabilities into Handshake, constructs Listener, and starts Listener inside the existing startup transaction. Admission remains closed while construction and Listener startup are in progress. The active Runtime context is created and published only as part of successful startup commit; the Gate then opens under the existing Host lifecycle boundary. A request observed before commit fails the initial admission check and does not run Authentication.

On startup failure, no active Runtime context is published and Gate remains closed. On Stop, Gate closure and root context cancellation are immediately visible through the same capabilities, while the existing Host shutdown order remains unchanged.

The capability contracts should be owned by the consuming Handshake boundary or by a narrowly neutral contract package. Runtime composition supplies implementations or adapters. Handshake, Listener, Authentication, and Session do not import the concrete Runtime Host.

## Ownership Model

| Resource or state | Owner | Handshake access |
|---|---|---|
| Admission Gate and its mutable state | Runtime Host | Live read-only admission capability |
| Root Runtime context and its cancellation | Runtime Host | Live read-only context capability; no `CancelFunc` |
| HTTP request, ResponseWriter, and accepted pre-Upgrade transport | `net/http` under Listener lifecycle | Borrowed synchronously for Handshake execution |
| Authentication result and admission decision | Handshake for one request | Owned until terminal rejection, abort, or commit |
| WebSocket after successful `websocket.Accept` and before handoff | Upgrade boundary | Exclusive ownership and close-on-failure responsibility |
| Derived connection context before handoff | Upgrade boundary | Created from the Runtime context and canceled on pre-handoff failure |
| WebSocket and derived connection context after successful handoff | Session | Exclusive lifecycle and closure responsibility |
| Runtime capability object | Runtime Host/composition | Owns no WebSocket, Session, Principal, or request state |

Ownership transfers exactly once. A Session creation or acceptance failure leaves ownership with the Upgrade boundary, which closes the WebSocket and cancels the derived context. After explicit successful handoff, upper layers do not close the WebSocket again.

## Alternatives Considered

### Alternative A — Pass the Whole Host

Handshake could receive a concrete Host reference and call its readiness, admission, and context methods.

This is rejected because it couples transport execution to Runtime lifecycle implementation, exposes unrelated operations such as Start and Stop, encourages Host to become a god object, complicates focused tests, and risks an `internal/runtime` to Listener dependency cycle. It violates the narrow dependency boundaries established by ARCH-001 and ADR-0003.

### Alternative B — Create a Second Gate in Listener

Listener could maintain its own admission state and synchronize it with Host.

This is rejected because two mutable sources of truth can diverge during startup, rollback, concurrent Stop, or future failure handling. It violates Host ownership frozen by ARCH-002 and makes correctness depend on synchronization between two Gates.

### Alternative C — Pass Read-Only Capabilities

Runtime composition passes only live admission permission and Runtime context access through narrow contracts or an immutable set of callbacks.

This is selected. It preserves Host ownership, keeps Handshake independent of concrete Runtime, supports controlled test doubles, exposes no cancellation or lifecycle mutation, and makes dependency wiring explicit. The capability set is closed to these two responsibilities for this decision; adding unrelated dependencies requires a separate focused design.

### Alternative D — Pass an Immutable Runtime Execution Environment

Host could create one read-only environment object containing admission and Runtime context access.

This can satisfy immediate behavior, but it creates an attractive place to accumulate Resolver, configuration, diagnostics, registries, and other services. That growth would turn the object into a service locator. A bundled value may be an implementation detail only if it remains a fixed carrier for the two selected narrow capabilities; it is not accepted as a generic Runtime environment abstraction.

### Alternative E — Move Gate Ownership to Listener

Listener could own Gate because it executes the transport commit.

This is rejected. It would change the ownership frozen by ARCH-002, split Runtime readiness from admission lifecycle, and require Host-to-Listener mutation during startup and shutdown. Such a direction would require an explicit revision of the Runtime Foundation architecture and offers no advantage over injecting the Host-owned permission capability.

## Frozen Foundation Impact

This decision is a compatible extension of the composition boundary described by ARCH-002, not a change to frozen lifecycle semantics.

The following frozen invariants remain unchanged:

- Host remains the production composition root and sole owner of lifecycle coordination;
- Host remains the sole owner of Admission Gate state;
- Host remains the sole owner of root Runtime context cancellation;
- `Build` opens no network resources and activates neither capability;
- active Runtime context, Running, readiness, and open admission are published only after successful startup commit;
- startup rollback leaves readiness and admission closed;
- Stop closes admission and cancels Runtime context using the existing order;
- Container does not become a service locator;
- restart and reload semantics do not change.

ARCH-002 explicitly leaves Handshake and application of Gate at admission commit open. Therefore this limited constructor-injection bridge does not require changing the freeze document or Host state machine. An implementation that changes Gate ownership, lifecycle states, context publication timing, cancellation ownership, startup commit, rollback, readiness, or shutdown ordering would exceed this ADR and require a new DP or ADR.

## Consequences

### Benefits

- Handshake can enforce Host admission without depending on Host.
- Session lifetime can derive from Runtime lifecycle rather than HTTP request lifetime.
- Gate closure and Runtime cancellation are observed without duplicated mutable state.
- Listener remains independent from `internal/runtime`.
- Composition stays explicit and test doubles can replace either capability independently.
- No global state, `context.Value` dependency injection, service locator, or second Gate is introduced.

### Drawbacks

- Runtime composition gains two additional explicit dependencies to wire.
- A stable context holder or equivalent adapter must exist before an active Runtime context is published.
- Handshake must not cache admission or context results across requests or lifecycle transitions.
- The final admission check must remain immediately adjacent to the Upgrade commit to preserve its linearization semantics.

## Follow-up Documentation

After this ADR is accepted:

- DP-001 should be revised to state that Host owns Gate while Handshake owns only the admission capability and commit execution.
- DP-002 should reference the stable capability bridge in its composition graph and construction order.
- ARCH-002 does not require an invariant change; after implementation it may reference this ADR as evidence of the permitted Handshake integration.
- `spec/current-state.md` should continue to distinguish this accepted decision from implemented pre-Upgrade Authentication until code and tests exist.
- MASTER_PLAN should be updated only when implementation materially changes Beta completion state.

## References

- [ADR-0003: Runtime Architecture](0003-runtime-architecture.md)
- [ARCH-001: Runtime Architectural Pattern](../architecture/ARCH-001-runtime-architectural-pattern.md)
- [ARCH-002: Runtime Foundation Freeze](../architecture/ARCH-002-runtime-foundation-freeze.md)
- [DP-001: Runtime Handshake Pipeline](../design/DP-001-runtime-handshake-pipeline.md)
- [DP-002: Runtime Host Composition Root](../design/DP-002-runtime-host-composition-root.md)
- [Master Engineering Plan](../roadmap/MASTER_PLAN.md)


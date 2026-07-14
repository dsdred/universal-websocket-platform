# ARCH-002: Runtime Foundation Freeze

[Russian version](../../ru/architecture/ARCH-002-runtime-foundation-freeze.md)

**Status:** Active

**Scope:** Implemented Runtime foundation

**Stability:** Frozen

## 1. Purpose

ARCH-002 records that the implemented Runtime foundation is architecturally stable. It does not propose a new architecture. It identifies the components and invariants that have been implemented, verified by tests, and may now be treated as a stable base for the next Beta work.

This document complements [ARCH-001](ARCH-001-runtime-architectural-pattern.md), [DP-001](../design/DP-001-runtime-handshake-pipeline.md), [DP-002](../design/DP-002-runtime-host-composition-root.md), the [Master Engineering Plan](../roadmap/MASTER_PLAN.md), and the recorded [current state](../../../spec/current-state.md). It freezes only the implemented subset of those designs.

## 2. Scope

The freeze covers Runtime Host composition and lifecycle infrastructure:

- Runtime Host;
- the production composition root;
- Host lifecycle coordination;
- Runtime root context;
- startup transaction and rollback;
- Runtime readiness;
- the lifecycle-only Admission Gate.

The freeze applies to architectural responsibilities, ownership, ordering, and observable lifecycle semantics. It does not freeze every private implementation detail or require current private types to become public contracts.

## 3. Frozen Runtime Components

### Runtime Host

Runtime Host owns an independent Snapshot copy, its Container, the composed Listener reference, Runtime context, lifecycle state, readiness, and Admission Gate. It coordinates construction, startup, rollback, and shutdown without absorbing Authentication, Session, Message, or transport business logic.

### Composition Root

Runtime Bootstrap creates and builds Host. Host is the production composition root for one Runtime instance. During Start, explicit constructors assemble the supported Authentication, connection dispatch, Session handoff, Message Handler, and Listener graph. No service locator, reflection, DI framework, or generic component registry is used.

The frozen property is the existence of one explicit production composition root and its dependency direction. The future Handshake order and future subsystem graph are not frozen by this document.

### Lifecycle

Host owns a single, thread-safe lifecycle and does not support restart or in-place reload. Long-running Listener operations execute outside the Host lifecycle mutex.

### Runtime Context

Host owns one Runtime context for a successfully started instance. Callers can observe it but cannot obtain its cancellation function.

### Startup Transaction and Rollback

Start acquires the composed Listener as a startup resource, starts it, commits only after success, and rolls back acquired resources on failure. Startup and rollback errors remain distinguishable through normal Go error wrapping and joining.

### Runtime Readiness

Readiness is a Host lifecycle fact. It becomes true only when startup commits and Host enters `Running`.

### Admission Gate

Host owns a small, thread-safe, lifecycle-only Admission Gate. It answers only whether the current Host lifecycle permits new admission. It contains no Authentication, Origin, rate-limit, maintenance, or configuration policy.

## 4. Frozen Architectural Invariants

The following implemented invariants are frozen:

- Host is the production composition root and lifecycle coordinator, not a business-logic service.
- Dependencies are connected explicitly through constructors and focused Bootstrap components.
- Container remains a Snapshot holder rather than a service locator.
- Host stores and returns independent Snapshot copies.
- `Build` prepares lifecycle state without opening network resources.
- Runtime dependencies are composed during `Start`; Listener is the externally visible component started by Host.
- Startup publishes Runtime resources only after successful Listener startup.
- Failed startup does not leave Host Running, Ready, or open for admission.
- Acquired startup resources are rolled back after startup failure.
- Shutdown clears readiness and closes admission before the potentially long Listener stop operation.
- Host lifecycle methods and state accessors are safe for concurrent use.
- Restart and in-place reload are not supported by a Host instance.

These invariants describe existing behavior. They do not define future Handshake, Session Manager, diagnostics, or supervision contracts.

## 5. Guaranteed Runtime Lifecycle

The implemented Host lifecycle is:

```text
Created
    -> Built
    -> Starting
    -> Running
    -> Stopping
    -> Stopped
```

- `Build` performs the one-time `Created -> Built` transition.
- `Start` is valid only from `Built`.
- A startup failure returns Host to `Built` after rollback, allowing a corrected startup attempt.
- Successful startup commits `Running` exactly once.
- `Stop` from `Running`, or concurrently with `Starting`, coordinates one shutdown operation.
- Concurrent and repeated Stop calls observe the same terminal shutdown result.
- Stop before Start is a no-op and does not prevent the first Start.
- `Stopped` is terminal; restart is not supported.

No `Initialized` or `Failed` state is part of the frozen implementation. Those states remain proposed in DP-002 rather than guaranteed here.

## 6. Guaranteed Startup Semantics

Startup has the following implemented semantics:

1. `NewHost` validates required Host inputs and stores an independent Snapshot through Container.
2. `Build` marks Host as prepared without acquiring or starting Runtime components.
3. `Start` enters `Starting` and keeps readiness and admission closed.
4. The supported component graph is assembled explicitly.
5. Listener is constructed and started.
6. If composition or Listener startup fails, acquired resources are rolled back and the original failure is preserved.
7. If Listener startup succeeds, Host publishes the Listener, creates the Runtime context, enters `Running`, opens admission, and becomes Ready as one lifecycle commit.

The context passed to `Start` bounds startup work that observes it. It does not become the lifetime context of a successfully running Runtime.

## 7. Guaranteed Runtime Context Semantics

- `RuntimeContext()` is nil before successful startup commit.
- Host creates the Runtime context only after Listener startup succeeds.
- The Runtime context is independent from the context passed to `Start`.
- Canceling the startup context after successful Start does not stop Runtime.
- Host owns the Runtime cancellation function and never exposes it.
- Stop cancels the Runtime context before calling Listener Stop.
- The same canceled context remains observable after Stop.
- Repeated Stop does not create or cancel another Runtime context.

This freeze does not define future child contexts for Handshake or Session.

## 8. Guaranteed Readiness Semantics

`Ready()` is true exactly while Host is in `Running`.

Readiness is false:

- after `NewHost` and `Build`;
- throughout `Starting`;
- during composition failure, Listener startup failure, and rollback;
- as soon as Stop begins;
- throughout `Stopping` and after `Stopped`.

Concurrent readers observe the lifecycle boundary safely. Readiness does not currently represent dependency health, traffic probes, supervision, or a Control Service health endpoint.

## 9. Guaranteed Admission Gate Semantics

`CanAccept()` returns a boolean lifecycle decision. Admission is open only when Host is both `Running` and Ready.

The Gate is closed:

- after Host creation and Build;
- throughout Starting;
- during composition failure and rollback;
- before startup commit;
- at the beginning of Stop, before Listener Stop is called;
- throughout a long Listener Stop;
- after Stop and after a rejected restart.

The Gate creates no goroutine and performs no network I/O. It is currently a Host-owned lifecycle boundary. Applying it atomically inside the future pre-Upgrade Handshake commit remains part of the open Handshake architecture.

## 10. Architecture Freeze

The Runtime Host, production composition-root boundary, implemented lifecycle, Runtime context, startup transaction and rollback, readiness, and lifecycle-only Admission Gate are considered stable.

Their architectural responsibilities, ownership, and lifecycle semantics may change only through a new focused Design Proposal or a new Architecture Decision. Ordinary bug fixes, test strengthening, and internal refactoring are allowed when they preserve these frozen invariants and observable semantics.

The freeze is not a production-readiness statement and does not promote Draft DP content that has not been implemented.

## 11. Open Architecture

The following areas remain open and are not frozen by ARCH-002:

- Handshake;
- Authentication Pipeline, including pre-Upgrade Authentication;
- Session ownership across Handshake handoff and Runtime shutdown tracking;
- Router;
- Delivery;
- Persistence;
- Operational Diagnostics;
- Runtime supervision.

These areas require focused design and implementation evidence. ARCH-002 does not define their APIs or internal component models.

## 12. Explicitly Out of Scope

This document does not freeze or design:

- Handshake or WebSocket admission commit;
- Router or route configuration;
- Delivery guarantees, addressing, groups, or topics;
- Persistence or database contracts;
- Plugin ABI or plugin lifecycle;
- Session Manager;
- TLS execution and effective Listener timeout behavior;
- operational diagnostics, monitoring, or supervision;
- reload, restart, failover, or process-level Runtime replacement.

It introduces no new API, interface, lifecycle state, configuration field, or Runtime requirement.

## 13. Change Policy

A proposed change must first be classified:

- a compatible implementation correction may proceed with tests that preserve this freeze;
- a subsystem addition outside the frozen scope requires its own focused design when contracts are not stable;
- a change to a frozen responsibility, ownership boundary, startup commit, rollback, Runtime context, readiness, Admission Gate, or lifecycle semantic requires a new DP or ADR before implementation.

ARCH-002 itself should be updated only to reflect an explicitly accepted architectural change or stronger implementation evidence. It must not be used as a shortcut around DP or ADR review.

## 14. Next Beta Phase

The frozen Runtime foundation is the base for the next Beta phase. The next work proceeds into open architecture, beginning with the Handshake boundary described by Draft DP-001 and its integration with Runtime lifecycle and admission.

Router, Delivery, Persistence, Plugins, and Session Manager remain later focused epics in the [Master Engineering Plan](../roadmap/MASTER_PLAN.md). Their future designs must build on the frozen foundation without turning Host into a business-logic component or Container into a service locator.

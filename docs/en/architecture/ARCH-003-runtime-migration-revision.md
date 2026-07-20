# ARCH-003: Runtime Foundation Migration Revision

[Russian version](../../ru/architecture/ARCH-003-runtime-migration-revision.md)

**Status:** Active

**Scope:** Runtime Foundation migration architecture

**Stability:** Approved migration sequence

## 1. Purpose

ARCH-003 records the revised implementation sequence for completing the Runtime Foundation defined by [DP-003](../design/DP-003-runtime-session-manager.md) and [DP-004](../design/DP-004-per-session-execution-boundary.md).

The target architecture remains unchanged. This document changes only migration architecture: the order and boundaries by which the repository can reach the approved target without temporarily introducing invalid ownership, publication, or shutdown behavior.

This document complements [ARCH-002](ARCH-002-runtime-foundation-freeze.md), the [Master Engineering Plan](../roadmap/MASTER_PLAN.md), and the recorded [current state](../../../spec/current-state.md). It does not replace DP-003 or DP-004.

For migration sequencing only, ARCH-003 supersedes the sequence in DP-004 section 29. Every normative target-architecture contract in DP-003 and DP-004 remains unchanged.

## 2. Decision Context

The previous remaining sequence treated Dormant Execution, Execution Binding integration, Dispatcher migration, Runtime callback integration, and Runtime shutdown as separately activatable steps. That sequence was not safely executable even though the approved target architecture was consistent.

The defect was in migration sequencing, not in DP-003 or DP-004. Several primitives can be implemented and tested independently, but their production ownership and publication effects are inseparable.

## 3. Completed Migration Tasks

The following migration tasks are complete:

1. Session Core;
2. Provisional Session;
3. Cleanup Acknowledgement;
4. Pre-Commit Session Bundle.

These tasks preserve the synchronous production Dispatcher. The private pre-Commit bundle is a structurally complete Session-side prepared object graph, not the normative Manager Commit result. No ownership transfer, execution publication, or Runtime integration has occurred.

## 4. Primitive and Production Responsibility

An independently implementable **primitive** has one local invariant that can be fully tested without activating incomplete production behavior. Existing examples include the one-shot Execution Binding, the bound Completion Adapter, and Owner Lifetime Lease accounting.

An inseparable **production responsibility** combines effects that must become observable together. It cannot be divided into production steps when doing so would create an ownerless resource, partial Commit, untracked Session, incomplete terminal path, or false shutdown convergence.

Independent implementation of a primitive does not authorize its isolated production activation.

## 5. Inseparable Migration Boundaries

### Dormant Handoff

Dormant Execution cannot be separated from:

- one Execution Binding integration;
- Dispatcher pre-Commit ownership;
- one `NotCommitted` publication for every non-committed outcome;
- dormant-path return before Dispatcher proceeds;
- prepared-value disposal;
- Reservation abort.

The dormant launch path is not a passive value or an Execution Owner alias. Without its Binding and Dispatcher-owned disposal sequence, the path could be orphaned or remain blocked after a recoverable pre-Commit failure.

### Atomic Commit Publication

Successful Manager Commit publication cannot be separated from:

- Registration publication;
- the Registration-bound Completion capability;
- Owner Lifetime Lease publication;
- Stop-publication capability;
- the `Committed` execution outcome;
- irreversible ownership transfer.

These effects share one synchronization boundary. Failure publishes none of them. No observable Registration may exist without one already eligible execution path.

### Complete Owner Lifecycle

Production handoff must not occur before the complete terminal Owner lifecycle exists. Releasing a committed path that stops at `Terminalizing` would leave Cleanup, Complete, Observer, callback drain, Terminal, and lease release obligations unresolved. Manager accounting could not converge truthfully.

### Production Cutover and Shutdown

Production Runtime activation must be combined with truthful shutdown integration. Runtime must not begin producing Manager-tracked Sessions until Host shutdown can request Stop through the captured Snapshot, cancel the root Runtime context, drain Listener handlers, and wait for Manager accounting.

## 6. Revised Remaining Migration Roadmap

### Task 5: Complete Atomic Commit Publication Foundation

**Objective:** Complete the Manager-side publication contract and compatible capability-bearing Shutdown Snapshot without activating production Session handoff.

**Architectural invariant:** A successful Manager Commit publishes one complete logical Commit result and the `Committed` execution outcome under one synchronization boundary. Failure publishes no committed state.

**Dependencies:** Completed Tasks 1–4 and the existing Manager, Execution Binding, Completion Adapter, Lifetime Lease, and Owner control foundations.

**Out of scope:** Dormant launch-path creation, production Dispatcher selection, Runtime callback lifecycle, complete Owner terminal execution, and Runtime shutdown cutover.

### Task 6: Runtime-Cancellation and Control-Call Lifecycle

**Objective:** Complete the Owner-local Runtime-cancellation observation and control-call admission, accounting, unregister, and drain contracts.

**Architectural invariant:** Explicit Stop and Runtime cancellation share one first-writer causal state, while callback registration and drain have one bounded lifecycle.

**Dependencies:** The required Owner control capability contracts established by the existing foundations and Task 5.

**Out of scope:** Session Start or Run migration, Manager Commit invocation, Terminal Result publication, dormant execution, production Dispatcher selection, and Runtime shutdown activation.

### Task 7: Terminal Result and Observer Contracts

**Objective:** Add the immutable terminal outcome and synchronous Observer boundaries without activating production handoff.

**Architectural invariant:** One committed Owner produces at most one immutable Terminal Result, and one synchronous Observer invocation receives no lifecycle ownership.

**Dependencies:** The required committed capability contracts established by Task 5.

**Out of scope:** Runtime callback implementation, Session execution migration, complete terminal orchestration, dormant execution, Dispatcher migration, and Runtime composition.

Tasks 6 and 7 are logically parallel. They should normally be committed sequentially so each concurrency and ownership contract receives a focused review.

### Task 8: Complete Execution Owner Terminal Lifecycle

**Objective:** Extend the existing Execution Owner skeleton through the complete committed terminal sequence.

**Architectural invariant:** Every committed execution follows:

```text
Cleanup
    -> Complete
    -> Terminal Result
    -> Observer
    -> UnregisterAndDrain
    -> seal
    -> Terminal
    -> conditional Lifetime Lease release
```

**Dependencies:** Tasks 6 and 7, Cleanup Acknowledgement from Task 3, and committed capabilities from Task 5.

**Out of scope:** Dormant launch-path creation, production Dispatcher selection, Runtime composition, Listener shutdown, and Manager Wait orchestration.

### Task 9: Transactional Dispatcher and Dormant Handoff

**Objective:** Build and fully test the complete pre-Commit transaction, dormant launch path, Commit handoff, and accepted-result boundary.

**Architectural invariant:** Exactly one dormant launch path and one Binding belong to one Dispatcher pre-Commit transaction. Every non-committed path performs:

```text
NotCommitted publication
    -> dormant path return
    -> prepared-value disposal
    -> Reservation abort
    -> accepted=false
```

Successful Commit makes exactly one execution path eligible and transfers ownership irreversibly.

**Dependencies:** Tasks 5 and 8 and the completed pre-Commit Session bundle.

**Out of scope:** Selection by production Runtime composition, Host shutdown ordering, Listener shutdown orchestration, and production Manager Wait.

A complete transaction-capable Dispatcher may temporarily exist off the production Runtime path. It must be complete and fully tested; it is not a partially active alternative ownership model.

### Task 10: Atomic Runtime Composition and Shutdown Cutover

**Objective:** Select the transaction-capable Dispatcher in production composition and activate truthful Runtime-wide shutdown accounting in the same production cutover.

**Architectural invariant:** Every production accepted Session is Manager-tracked from Commit through completion and Lifetime Lease release. Shutdown follows:

```text
close Admission
    -> BeginShutdown
    -> Snapshot RequestStop
    -> cancel root Runtime context
    -> Listener Stop
    -> Manager Wait
```

**Dependencies:** Task 9 and the frozen Host, Admission Gate, Runtime context, Listener, and startup-rollback contracts from ARCH-002.

**Out of scope:** Changes to the approved target architecture, Router, Delivery, Persistence, Plugins, Metrics, diagnostics backends, restart, and reload.

## 7. Dependency Graph

```text
Completed:
Task 1: Session Core
    -> Task 2: Provisional Session
    -> Task 3: Cleanup Acknowledgement
    -> Task 4: Pre-Commit Session Bundle
                |
                v
Task 5: Complete Atomic Commit Publication Foundation
                |
                +-------------------+
                |                   |
                v                   v
Task 6: Runtime-Cancellation   Task 7: Terminal Result
and Control-Call Lifecycle     and Observer Contracts
                |                   |
                +---------+---------+
                          |
                          v
Task 8: Complete Execution Owner Terminal Lifecycle
                          |
                          v
Task 9: Transactional Dispatcher and Dormant Handoff
                          |
                          v
Task 10: Atomic Runtime Composition and Shutdown Cutover
```

## 8. Retired Remaining Sequence

The following standalone remaining task boundaries are retired:

- Dormant Execution without Binding integration and Dispatcher ownership;
- a remaining standalone Execution Binding task;
- Runtime callback as an isolated production integration step;
- Dispatcher migration without dormant-path ownership and disposal;
- Runtime shutdown separated from production activation.

Historical completed-task information remains valid. The existing Execution Binding primitive remains implemented and tested; only its integration moves into the atomic publication and handoff tasks.

## 9. Compatibility and Production Status

This revision introduces no Runtime capability and changes no production behavior. The current production Dispatcher remains synchronous, and Runtime composition does not yet construct or coordinate the Session Manager.

DP-003 and DP-004 remain the normative target architecture. ARCH-002 remains unchanged: production cutover must preserve its frozen Host lifecycle, readiness, Admission Gate, Runtime context, startup rollback, and Listener ordering.

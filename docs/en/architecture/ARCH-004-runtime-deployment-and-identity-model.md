# ARCH-004: Runtime Deployment and Identity Model

[Russian version](../../ru/architecture/ARCH-004-runtime-deployment-and-identity-model.md)

**Status:** Active

**Scope:** Single-node Runtime operational identity, ownership, lifecycle, and management boundary

**Stability:** Approved architectural model

## 1. Purpose

ARCH-004 defines the operational identity and ownership model between a Published ConfigurationVersion and one running Runtime Host.

The document closes the identity and lifecycle gap identified by TASK-M11-DISCOVERY-001. It establishes the stable concepts required before Control Service can create, start, stop, replace, or observe Runtime instances. It does not define implementation APIs, persistence schemas, Configuration loading, or Snapshot construction.

This model complements [ADR-0002](../adr/0002-configuration-dsl.md), [ADR-0003](../adr/0003-runtime-architecture.md), [ARCH-002](ARCH-002-runtime-foundation-freeze.md), and the Runtime foundation defined by [DP-003](../design/DP-003-runtime-session-manager.md), [DP-004](../design/DP-004-per-session-execution-boundary.md), and [DP-006](../design/DP-006-runtime-production-integration.md). It does not change their Runtime-internal ownership or lifecycle contracts.

## 2. Context

The repository contains two independently implemented boundaries:

```text
Configuration
    -> ConfigurationVersion
    -> Published ConfigurationVersion
```

and:

```text
Immutable Runtime Snapshot
    -> Runtime Bootstrap
    -> Runtime Host
```

There is no production owner for the transition between them. ConfigurationVersion cannot safely serve as the identity of both declarative configuration and operational execution because one Configuration may be launched repeatedly, may have more than one independently managed Runtime instance, and may move between Published versions without changing the identity of the managed server.

Runtime Host is intentionally limited to one execution lifetime. It owns Runtime resources but does not own Control Plane identity, desired state, persistence, process supervision, or Configuration selection.

## 3. Scope

This document defines:

- Runtime Instance identity;
- Launch Attempt identity;
- the relationship between Workspace, Configuration, ConfigurationVersion, Runtime Instance, Launch Attempt, and Runtime Host;
- operational ownership before, during, and after a launch;
- desired and actual lifecycle state;
- the management boundary used to start, stop, and replace Runtime execution;
- concurrency and identity invariants;
- the initial single-node execution topology.

This document does not define:

- Go types or public HTTP endpoints;
- repository or PostgreSQL schemas;
- Configuration Loader or immutable Snapshot fields;
- Secret resolution;
- authorization for management operations;
- automatic restart, failover, scheduling, clustering, or federation;
- zero-downtime replacement;
- in-place reload;
- deployment-specific Configuration overrides;
- metrics, diagnostics storage, or retention.

## 4. Architectural Terms

### Configuration

Configuration is the stable declarative parent inside one Workspace. It owns ConfigurationVersion history and contains no Runtime operational state.

### ConfigurationVersion

ConfigurationVersion is one immutable declarative payload after publication. It is a source of Runtime behavior, not a Runtime identity, launch record, or process record.

### Runtime Instance

Runtime Instance is the stable operational identity of one independently managed WebSocket server. It belongs to exactly one Workspace and exactly one Configuration.

Runtime Instance survives individual starts, stops, failed starts, and replacements. It contains or references operational desired and actual state, but it does not duplicate the Configuration payload.

Deletion of a Runtime Instance, deletion of its Launch Attempt history, and retention policy are intentionally outside this document. A focused Design Proposal must define those semantics before deletion is implemented.

### Launch Attempt

Launch Attempt is one immutable execution identity created whenever the lifecycle owner attempts to start a Runtime Instance. A Launch Attempt belongs to exactly one Runtime Instance and pins exactly one Published ConfigurationVersion as its source.

A failed startup remains a distinct Launch Attempt. A restart or replacement creates a new Launch Attempt; it never reuses the identity of an earlier attempt.

### Runtime Host

Runtime Host is the Runtime-internal composition and lifecycle owner for one Launch Attempt. It owns the immutable Snapshot copy, Listener, Runtime context, Session Manager, and the component graph defined by the existing Runtime architecture.

Host identity is not Control Plane identity. A pointer, process address, goroutine, context, socket, or PID must not be used as Runtime Instance or Launch Attempt identity.

### Runtime Lifecycle Owner

Runtime Lifecycle Owner is the Control Service-side orchestration responsibility that serializes management operations for one Runtime Instance, creates Launch Attempts, invokes the Runtime Launcher, owns the active Host reference, and records truthful actual state.

This responsibility may be implemented by focused components. It is a role, not permission to turn Control Service or Runtime Host into a universal manager.

## 5. Identity Model

The normative identity graph is:

```text
Workspace
    -> Configuration
        -> Runtime Instance
            -> Launch Attempt
                -> Runtime Host

Configuration
    -> ConfigurationVersion history

Launch Attempt
    -> exactly one Published ConfigurationVersion
```

The following cardinality rules apply:

- one Workspace may own many Configurations;
- one Configuration may own many ConfigurationVersions;
- one Configuration may have many Runtime Instances;
- one Runtime Instance belongs to one Configuration for its entire lifetime;
- one Runtime Instance may have many historical Launch Attempts;
- one Runtime Instance may have at most one active Launch Attempt;
- one Launch Attempt has at most one Runtime Host;
- one Launch Attempt pins one exact Published ConfigurationVersion;
- many Runtime Instances may use the same ConfigurationVersion when their effective Listener resources do not conflict.

Runtime Instance ID and Launch Attempt ID are stable, opaque identifiers. Their concrete representation and allocation strategy are implementation decisions. PID is optional observed process metadata and never domain identity.

An active Launch Attempt remains active for the entire period in which Runtime Lifecycle Owner owns its associated Runtime Host, including startup and shutdown. After the owner observes terminal completion and releases the Host reference, the Launch Attempt is historical. A startup attempt that produces no owned Host becomes historical when its failure is recorded. A historical Launch Attempt never becomes active again. These terms describe ownership duration and do not introduce a separate Launch Attempt state machine.

## 6. Configuration Binding

Runtime Instance is bound to a Configuration, not permanently to one ConfigurationVersion. This preserves stable operational identity while the Configuration evolves through new immutable versions.

Each Launch Attempt records the exact source version selected before Runtime construction. Publication of a newer ConfigurationVersion does not mutate an existing Launch Attempt, Snapshot, Host, or running Runtime.

The initial architecture permits no Runtime Instance override of behavior-affecting Configuration fields. Listener host and port remain part of ConfigurationVersion. Supporting placement or binding overrides requires a separate architectural decision because such overrides would otherwise become a second source of Runtime behavior.

## 7. Operational State

Desired state and actual state are separate.

The initial desired states are:

- **Stopped:** no Runtime execution is requested;
- **Running:** one Runtime execution is requested.

Adding another desired state, including concepts such as Maintenance, Paused, or Draining, requires a separate approved Design Proposal. Implementations must not infer additional desired states from actual state or management commands.

The initial actual states are:

- **Stopped:** no active Launch Attempt owns Runtime resources;
- **Starting:** one Launch Attempt is acquiring and starting Runtime resources;
- **Running:** the active Host completed startup and is Ready;
- **Stopping:** the lifecycle owner is waiting for the active Host to release its resources;
- **Failed:** the latest Launch Attempt failed to start or terminated without satisfying the requested lifecycle outcome.

This is an operational lifecycle outside Runtime Host. It does not change the Host lifecycle frozen by ARCH-002 and does not introduce new states into Runtime, Session Manager, Session, or Execution Owner.

## 8. Lifecycle Transitions

The supported operational transitions are:

```text
Stopped
    -> Starting
    -> Running
    -> Stopping
    -> Stopped

Starting
    -> Failed

Starting
    -> Stopping
    -> Stopped

Running
    -> Failed

Failed
    -> Starting

Failed
    -> Stopped
```

Rules:

- Start requests desired `Running`.
- A Start request while the same Runtime Instance is already `Starting` or `Running` must not create another Launch Attempt.
- Startup success is linearized only after Host reports successful Start and Runtime readiness is open.
- Startup failure publishes no Running state and leaves no active Host ownership.
- A Stop request during `Starting` claims the same Launch Attempt, prevents Running publication, and waits for Host startup rollback or shutdown before publishing `Stopped`.
- Stop requests desired `Stopped` and applies only to the currently active Launch Attempt.
- Stop completion is published only after Host has completed its owned shutdown contract.
- Repeated or concurrent Stop must converge on the same active Launch Attempt and must not create another shutdown owner.
- Restart is not a Host operation. If later exposed, it is an orchestration of Stop followed by a new Launch Attempt.
- Replacement is not in-place reload. It requires a new Launch Attempt with a new immutable Snapshot.

Automatic restart after failure is not part of the initial model. A future policy may react to desired/actual divergence only after its retry, backoff, and failure semantics are approved.

## 9. Ownership Model

| Object or state | Creator | Owner before start | Owner while active | Terminal responsibility |
|---|---|---|---|---|
| Workspace | Control Plane | Control Plane | Control Plane | Control Plane persistence policy |
| Configuration | Control Plane | Control Plane | Control Plane | Control Plane persistence policy |
| ConfigurationVersion | Control Plane | Control Plane | Control Plane; borrowed as immutable source | ConfigurationVersion lifecycle |
| Runtime Instance identity | Control Plane management boundary | Runtime Lifecycle Owner | Runtime Lifecycle Owner | Remains until explicit domain deletion |
| Desired state | Management command boundary | Runtime Lifecycle Owner | Runtime Lifecycle Owner | Retained as operational intent |
| Launch Attempt identity | Runtime Lifecycle Owner | Runtime Lifecycle Owner | Runtime Lifecycle Owner | Retained as launch history according to future persistence policy |
| Immutable Snapshot | Runtime boundary | Launch preparation | Runtime Host | Value lifetime ends with Host |
| Runtime Host | Runtime Launcher | Runtime Lifecycle Owner during construction | Runtime Host owns Runtime resources; Lifecycle Owner owns the Host reference | Lifecycle Owner waits for Host completion and releases the reference |
| Actual state | Runtime Lifecycle Owner | Runtime Lifecycle Owner | Runtime Lifecycle Owner from observed Host outcomes | Must remain truthful after success or failure |
| PID or worker metadata | Execution adapter, if applicable | None | Runtime Lifecycle Owner observes it | Cleared or retained as non-identity history |

Configuration services never own Runtime resources. Runtime Host never owns Runtime Instance identity, desired state, repository state, or Configuration publication.

## 10. Management Boundary

Management operations target Runtime Instance ID. They never target a Host pointer, PID, Listener address, Session, or ConfigurationVersion as an execution identity.

The management boundary is responsible for:

- verifying Workspace and Configuration ownership;
- serializing state-changing operations for one Runtime Instance;
- loading the selected Published ConfigurationVersion through the future Loader boundary;
- creating one Launch Attempt;
- asking Runtime Launcher to construct and start one Host;
- publishing actual state from observed outcomes;
- stopping only the active Host;
- preserving identity and state across caller retries.

Runtime Launcher accepts prepared launch input and returns lifecycle ownership of one Host or a startup failure. It does not select Configuration, mutate desired state, read management repositories, or decide retry policy.

Runtime Launcher is a stateless construction boundary. It does not own Runtime Instance, retain lifecycle state, contain a Runtime registry, make management decisions, or become a second Runtime Lifecycle Owner.

Exact commands, HTTP resources, error contracts, authorization, and repository mutations require focused Design Proposals.

## 11. Initial Execution Topology

The initial Beta topology is single-node and in-process: Control Service owns the Runtime Lifecycle Owner, which launches Runtime Host through an explicit Runtime Launcher boundary.

The identity model does not equate in-process execution with domain identity. A future approved execution adapter may place Runtime Host in a child process without changing Runtime Instance or Launch Attempt identity.

Domain identity is fully independent of execution topology: changing between single-process and multi-process execution does not change the Runtime Instance or Launch Attempt model.

Separate process supervision, PID persistence, crash recovery after Control Service restart, remote workers, scheduling, and clustering are out of scope. They must not be simulated through hidden state in Runtime Host.

## 12. Concurrency and Linearization

All state-changing management operations for one Runtime Instance must share one serialization boundary.

The architecture requires distinct linearization points for:

- **Launch claim:** creation of the only active Launch Attempt from a state that permits Start;
- **Running publication:** observation of successful Host startup and readiness;
- **Stop claim:** transfer of shutdown responsibility for the active Launch Attempt to one lifecycle operation;
- **Stopped publication:** observation that Host shutdown completed and no active resources remain;
- **Failure publication:** recording that the Launch Attempt cannot become or remain Running.

No observer may see:

- two active Launch Attempts for one Runtime Instance;
- actual `Running` before Host readiness;
- actual `Stopped` while an owned Host still holds resources;
- an active Host without one Launch Attempt identity;
- one Launch Attempt associated with different ConfigurationVersions;
- a reused Launch Attempt identity after Stop or failure.

Operations on different Runtime Instances may proceed independently, subject to external resource conflicts such as Listener address ownership.

## 13. Publication, Replacement, and Rollback

Publishing a new ConfigurationVersion changes Configuration state only. It does not automatically:

- start a Runtime Instance;
- stop a Runtime Instance;
- mutate a running Snapshot;
- replace a Host;
- change the source identity of an active Launch Attempt.

Configuration rollback and Runtime replacement are separate responsibilities:

- Configuration rollback selects an immutable version for future launch preparation;
- Runtime replacement creates a new Launch Attempt from an explicitly selected version;
- an existing Launch Attempt continues with its pinned Snapshot until its Host stops.

The ordering, availability, and failure semantics of replacement and rollback require a focused DP. No in-place reload is permitted by this model.

## 14. Failure and Recovery Boundaries

A failed start produces a failed Launch Attempt and truthful actual `Failed`. It must not publish Runtime readiness or retain unowned Host resources.

An unexpected active-runtime termination produces actual `Failed` after Runtime-owned cleanup is observed. It does not silently create a replacement Launch Attempt.

Caller cancellation bounds the caller's wait only when the focused command contract says so. It must not transfer active Host ownership to an untracked goroutine or cause actual state to claim cleanup that has not happened.

Recovery after Control Service process termination depends on future persistent Runtime Instance and Launch Attempt state plus an approved reconciliation contract. This document does not infer a running Host from stale PID data.

## 15. Security and Workspace Boundary

Runtime Instance belongs to the same Workspace as its Configuration. A management operation must not bind an existing Runtime Instance to a Configuration in another Workspace.

Operational identity contains references and lifecycle state, not Secret values. Runtime Instance and Launch Attempt records must not become an alternate Configuration or credential store.

Authorization policy for creating or controlling Runtime Instances is outside scope, but it must be evaluated at the management boundary before lifecycle mutation.

## 16. Compatibility with Existing Runtime Architecture

ARCH-004 does not change:

- Published ConfigurationVersion as the behavior source;
- immutable Runtime Snapshot ownership;
- Runtime Host as the composition root for one execution lifetime;
- Host startup, readiness, Admission Gate, rollback, or shutdown;
- Runtime Context;
- Transactional Dispatcher;
- Session Manager accounting;
- Execution Owner lifecycle;
- Listener, Router, Authentication, or Secret Resolver contracts.

The current repository has no production Runtime Lifecycle Owner, Runtime Launcher integration, Runtime Instance repository, management API, or process supervision. This document defines architecture only and declares none of those capabilities implemented.

## 17. Required Invariants

1. Runtime Instance is operational identity and ConfigurationVersion is declarative identity.
2. Runtime Instance belongs permanently to one Workspace and one Configuration.
3. One Runtime Instance has at most one active Launch Attempt.
4. Every Launch Attempt pins exactly one Published ConfigurationVersion.
5. Every Runtime Host belongs to exactly one Launch Attempt.
6. Start, Stop, and future replacement operations target Runtime Instance ID.
7. PID, socket address, Host pointer, context, goroutine, and Session ID are not Runtime identity.
8. Actual `Running` is impossible before Host readiness.
9. Actual `Stopped` is impossible before all Host-owned resources are released.
10. Publication does not mutate or restart an active Runtime.
11. Restart and replacement create a new Launch Attempt and Snapshot.
12. Runtime Host does not read Control Plane repositories or mutate desired state.
13. Control Plane does not take ownership of Runtime-internal Session or execution lifecycle.
14. Operational state does not become hidden Configuration.
15. No behavior-affecting deployment override exists without a separate approved decision.

## 18. Consequences

### Benefits

- Runtime execution receives stable operational identity without corrupting Configuration identity.
- Start retries, restarts, replacements, and failures are distinguishable.
- One Configuration may support independently managed Runtime Instances.
- Desired and actual state can remain truthful without entering Runtime Host.
- Snapshot provenance can refer to a concrete Launch Attempt.
- A future process adapter does not require changing domain identity.

### Costs

- Control Service requires a new domain model and lifecycle owner.
- Production persistence must preserve identity and state transition invariants.
- Management concurrency requires explicit serialization and idempotency.
- Replacement, rollback, and recovery require focused contracts rather than implicit side effects.

## 19. Follow-up Architecture

Implementation must not begin before focused design resolves at least:

1. Configuration Loader, Snapshot provenance, and schema compatibility;
2. Runtime Instance and Launch Attempt persistence contracts;
3. management command and idempotency contracts;
4. activation, replacement, and rollback ordering;
5. recovery and reconciliation after Control Service termination;
6. operational error reporting and redaction.

Process isolation, automatic restart, scheduling, and clustering remain later decisions and are not prerequisites for the initial in-process single-node implementation.

## 20. Decision

UWP adopts a separate Runtime Instance as the stable operational identity between Configuration and Runtime execution.

Runtime Instance belongs to one Workspace and one Configuration. Every start creates a new immutable Launch Attempt that pins one Published ConfigurationVersion and owns at most one Runtime Host execution. The Control Service-side Runtime Lifecycle Owner serializes management operations, owns desired and actual state, and holds the active Host reference. Runtime Host remains the existing composition and resource owner for one execution lifetime and does not acquire Control Plane responsibilities.

The initial topology is single-node and in-process behind an explicit Runtime Launcher boundary. Publication, restart, replacement, rollback, and process recovery never mutate an active Snapshot and require explicit lifecycle operations or later focused designs.

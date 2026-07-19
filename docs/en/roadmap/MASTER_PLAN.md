# Universal WebSocket Platform Master Engineering Plan

[Russian version](../../ru/roadmap/MASTER_PLAN.md)

**Status:** Living engineering plan

This document describes the intended engineering evolution of Universal WebSocket Platform. It defines maturity stages, architectural epics, and release criteria without assigning dates, performance targets, or future subsystem APIs. It is neither a marketing roadmap, a release schedule, nor a task backlog.

## 1. Project Vision

Universal WebSocket Platform is an open-source platform for creating, configuring, deploying, and operating independent WebSocket servers without repeatedly writing infrastructure code.

Users describe server behavior through understandable Configuration and manage it through stable APIs. Runtime executes that Configuration predictably, with explicit Provider boundaries, isolated ownership of resources, and behavior that operators can explain before deployment.

The full product intent is defined in the [Project Vision](../../../spec/00-product/vision.md). This plan translates that intent into engineering maturity stages; it does not replace the Vision.

## 2. Current State

The repository currently contains an Alpha foundation rather than a production-ready platform. The implemented state is recorded in [`spec/current-state.md`](../../../spec/current-state.md) and assessed by the [Runtime Alpha Architecture Review](../reviews/runtime-alpha-review.md).

### Configuration

- Workspace, Configuration, and ConfigurationVersion have in-memory Control Service APIs.
- ConfigurationVersion supports create, publish, and archive lifecycle operations.
- Listener, TLS, timeout, and Authentication metadata are represented in the Configuration DSL.
- Authentication metadata includes API Key, JWT, and Basic settings, with validation separated into a dedicated component.

### Snapshot

- `runtimeconfig.Builder` accepts only Published ConfigurationVersion.
- Runtime Snapshot contains Listener and Authentication data.
- Nested Provider and JWT collections are copied so later Configuration changes do not alter an existing Snapshot.

### Listener and Connection

- Listener opens a TCP socket and runs an HTTP server.
- `GET /ws` performs an RFC 6455 Upgrade and transfers the connection through a Dispatcher.
- TLS and Listener timeout metadata are not yet fully enforced by Runtime.

### Authentication

- Transport-neutral request, result, Principal, Provider, Factory, Registry, Service, and Bootstrap boundaries exist.
- Production API Key and HMAC JWT Providers and Factories exist.
- Secret values are resolved per Authentication attempt through Secret References.
- Authentication currently occurs after WebSocket Upgrade; Basic and asymmetric JWT verification are absent.

### Session, Message, and Echo

- Session owns an authenticated WebSocket connection, one read loop, and serialized writes.
- Runtime Message is transport-neutral and copies text or binary payload.
- Message Handler is independent of WebSocket transport.
- EchoHandler demonstrates inbound Message to Handler to `Session.Send` flow.

### Architecture

- [ADR-0002](../adr/0002-configuration-dsl.md) defines ConfigurationVersion as the Configuration DSL and Published source of truth.
- [ADR-0003](../adr/0003-runtime-architecture.md) defines the component Runtime model and explicit dependency injection.
- [ARCH-001](../architecture/ARCH-001-runtime-architectural-pattern.md) records the confirmed `Context -> Evaluation -> Decision -> Execution` pattern, ownership, lifecycle, and Boring Core.
- The Runtime Alpha Review verdict is **Ready with findings**. Runtime Host is not yet a production composition root, and lifecycle, effective Listener settings, and operational diagnostics require further work.

## 3. Engineering Principles

The primary design guidance is [ARCH-001: Runtime Architectural Pattern](../architecture/ARCH-001-runtime-architectural-pattern.md).

### Configuration First

Configuration defines Runtime behaviour. Published ConfigurationVersion is the source of truth, and Runtime executes an independently copied Snapshot. Every startup-critical capability must be executable or explicitly rejected before Runtime becomes Ready. Runtime capabilities assigned by the approved architecture to later roadmap gates may remain configured but inactive, and that status must be explicit.

### Ownership

Every mutable resource has one current owner. Ownership transfer is explicit, including success, rejection, cancellation, and partial failure paths. Components may accept resources created elsewhere, but the contract must say who closes them.

### Lifecycle

Components with resources have explicit lifecycle and concurrency semantics. Every goroutine has an owner and termination path. Context cancellation participates in shutdown, long operations do not execute under lifecycle locks, and shutdown errors remain observable.

### Boring Core

The core remains small. New behavior should first fit an appropriate existing seam such as Handler or Provider. A new Matcher, Policy, Middleware, or Plugin seam is introduced only when a concrete subsystem requires it.

### No Magic

Dependencies are passed explicitly. Runtime does not discover Control Plane repositories, use globals or `context.Value` as service locators, create hidden Configuration, or start detached work without an owner.

## 4. Milestones

Milestones describe engineering maturity. They do not specify calendar dates or guarantee public release timing.

### Alpha — Prove the Foundation

**Goal:** Demonstrate that Configuration can become an isolated Runtime Snapshot and drive a minimal WebSocket vertical with explicit component boundaries.

**Completion criteria:**

- ConfigurationVersion lifecycle and core Listener/Authentication metadata exist.
- Published Configuration builds a copied Runtime Snapshot.
- Listener, Connection Dispatcher, Authentication, Session, Message, Handler, and Send form a tested vertical.
- Secret values remain outside Configuration and Snapshot.
- Architecture review identifies stable decisions, limitations, and debt.
- ARCH-001 captures only patterns supported by implementation evidence.

**Main epics:** Configuration DSL, Snapshot, Listener, Authentication contracts and Providers, Session, Runtime Message, Echo vertical, architecture review, and architecture guidance.

### Beta — Complete the Single-Node Runtime

**Goal:** Turn the Alpha components into one coherent, configurable, observable, and safely operated single-node Runtime.

**Completion criteria:**

- Runtime Host is the production composition root and owns ordered startup, rollback, and shutdown.
- Authentication and Origin evaluation occur at the correct pre-Upgrade Handshake boundary.
- Router deterministically selects configured behavior without transport leakage.
- Session ownership and coordinated management are defined and tested.
- Delivery and the minimum required Persistence behavior have explicit semantics.
- TLS, timeouts, and all accepted Published settings are applied or rejected at startup.
- Metrics and operational diagnostics expose failures without exposing credentials or Secrets.
- High findings from the Alpha review are closed and no unsupported capability is represented as active.

**Main epics:** Handshake, Runtime Host, Configuration validation, Router, Session Manager, Delivery, Persistence, TLS, Metrics, operational diagnostics, and initial Plugin contracts.

### RC — Stabilize the Product Contract

**Goal:** Stabilize the supported Configuration, API, Runtime lifecycle, and operational contract before 1.0.

**Completion criteria:**

- No unresolved Blocker or High architecture findings remain.
- Supported Configuration and API compatibility rules are documented and tested.
- Upgrade, migration, restart, failure recovery, and shutdown paths are exercised end to end.
- Security and operational reviews cover Handshake, Secrets, TLS, diagnostics, and persistence boundaries.
- Unsupported Configuration is rejected with actionable, non-sensitive diagnostics.
- Documentation describes actual behavior, limitations, and recovery procedures.

**Main epics:** contract hardening, compatibility, failure recovery, security review, operational review, migration verification, and release qualification.

### 1.0 — Stable Single-Node Platform

**Goal:** Provide the first stable, supportable implementation of the core UWP model for independent WebSocket servers on a single Runtime node.

**Completion criteria:**

- All mandatory 1.0 criteria in Section 6 are satisfied.
- Published Configuration produces reproducible Runtime behavior.
- The full ownership and lifecycle chain is testable and observable.
- Public contracts have a documented backward-compatibility policy.
- The platform fails explicitly when Configuration or operational dependencies cannot be honored.

**Main epics:** closure of RC findings, stable documentation, compatibility guarantees, and supportable operational behavior. New core features are not an objective of this milestone.

### 2.0+ — Evidence-Driven Expansion

**Goal:** Consider capabilities that change deployment topology, compatibility boundaries, or extension isolation only after stable single-node use provides evidence.

**Completion criteria:**

- Each capability has concrete use cases and a focused DP.
- Consequential compatibility or topology decisions have an ADR.
- Existing 1.x behavior has an explicit migration strategy.
- Distributed ownership, consistency, and failure semantics are defined before implementation.

**Main epics:** possible clustering, federation, horizontal coordination, stronger plugin isolation, and broader integrations. Inclusion is not promised by this plan.

## 5. Beta Roadmap

Beta work is organized as architectural epics. Each epic requires focused design where contracts are not already stable.

### Foundation Gates

1. **Handshake:** move Authentication and future Origin evaluation before Upgrade, with explicit allow/reject and transport error semantics. A separate DP must define the pipeline.
2. **Runtime Host:** compose existing components without becoming a god object; own startup ordering, partial-start rollback, and shutdown ordering.
3. **Lifecycle hardening:** make cancellation and concurrent Stop semantics consistent, remove lifecycle locks from network I/O, and retain shutdown errors.
4. **Configuration validation:** ensure every startup-critical capability is executable or explicitly rejected before Runtime becomes Ready; capabilities assigned to later gates may remain configured but explicitly inactive.

These gates precede Router because Router must not inherit an unstable security, composition, or shutdown boundary.

### Message Processing

5. **Router:** select a Handler or destination from Runtime Message Context without accessing WebSocket transport. A dedicated DP defines routing semantics, not this plan.
6. **Session Manager:** coordinate Session registration, removal, limits, and shutdown without taking over Session transport ownership.
7. **Delivery:** define recipient selection, ordering, failure, and backpressure semantics before adding Groups, Topics, or broadcast features.
8. **Persistence:** define what is persisted, when storage is optional or required, and how failures affect Message processing. Storage technology and API remain undecided.

### Operational Completion

9. **TLS and Listener settings:** resolve certificate references safely and apply TLS, handshake, read, write, and idle limits.
10. **Metrics:** expose bounded, non-sensitive component and lifecycle measurements with controlled label cardinality.
11. **Operational diagnostics:** report server, Dispatcher, Provider, Handler, and shutdown failures while preserving Secret and credential redaction.
12. **Plugin contracts:** identify the smallest extension contracts supported by multiple real use cases. Dynamic loading and ABI are not implied.

This ordering expresses architectural dependencies, not dates or a task queue. Epics may be split by future DPs and reviews.

## 6. Release Criteria

### Mandatory for 1.0

- A production composition root builds and owns the complete supported Runtime vertical.
- Published Configuration is the only behavior source; all supported fields are effective and unsupported fields fail explicitly.
- Handshake performs required Authentication and connection admission before WebSocket Upgrade.
- Router behavior is deterministic, explainable, and transport-neutral at the Message boundary.
- Session lifecycle, coordinated shutdown, writes, and resource ownership are race-tested.
- Required delivery and persistence semantics survive defined restart and failure scenarios.
- TLS, Origin policy, timeouts, Secret resolution, and credential redaction have end-to-end security tests.
- Metrics and diagnostics expose component health and failures without leaking sensitive data.
- Control and Runtime API contracts have documented compatibility and validation behavior.
- Recovery, migration, shutdown, and operational procedures are documented and tested.

### Eligible for 1.x

- Additional Authentication Providers and algorithms beyond the supported 1.0 set.
- Additional Secret Storage adapters.
- Advanced routing and delivery policies based on demonstrated use cases.
- Import/export improvements and operational automation.
- Admin UI capabilities that consume stable APIs.
- A supported plugin development surface after extension contracts stabilize.

### Reserved for 2.0 or Later Evaluation

- Cluster coordination and distributed Runtime ownership.
- Federation between independent platform domains.
- Horizontal scaling semantics that require distributed Session or delivery state.
- Compatibility-breaking Plugin ABI or process isolation.
- Broad enterprise and cloud integration families.

Placement after 1.0 is not a commitment to implement a feature. Every item still requires evidence, design, and prioritization.

## 7. Deferred Features

The following features are intentionally deferred because they introduce distributed ownership, external compatibility, or operational complexity beyond the single-node core:

- Cluster operation;
- Federation;
- horizontal scaling;
- distributed Session and delivery coordination;
- enterprise plugin families;
- broad cloud integrations;
- dynamic Plugin ABI and isolation;
- cross-node Message ordering and recovery.

Deferred features must not shape the current core through speculative abstractions. They can be reconsidered when stable use produces concrete requirements.

## 8. Technical Debt

The Runtime Alpha Review identifies implementation debt that must be tracked independently from new functionality:

- Listener stores TLS and timeout metadata without fully enforcing it.
- Some shutdown waits are not bounded by the caller's context.
- Concurrent Stop semantics differ between components.
- Session currently holds a lifecycle read lock across WebSocket write.
- HTTP server and Dispatcher errors lack an operational reporting path.
- `runtimeconfig.Builder` is the explicit ConfigurationVersion adapter inside the Runtime model package and must not accumulate Repository concerns.
- Basic and asymmetric JWT Runtime coverage is incomplete.
- Origin behavior relies on library defaults rather than explicit Configuration.

Technical debt is closed through tests and implementation changes, not by relabeling current limitations as supported behavior.

## 9. Architectural Debt

Architectural debt concerns boundaries that are unresolved or not yet represented by production composition:

- **Authentication before Upgrade:** current post-Upgrade Authentication conflicts with the intended Handshake security boundary.
- **Runtime Host:** the Host does not yet assemble or own the implemented Runtime vertical.
- **Effective Listener Configuration:** TLS and timeout metadata can reach Snapshot without complete execution or explicit rejection.
- **Shutdown semantics:** cancellation, concurrent Stop completion, error propagation, and long-running Handler behavior need a uniform contract.
- **Operational diagnostics:** error ownership and redaction must cross component boundaries without coupling components to one logging implementation.
- **Extension boundaries:** Router is ready to use Handler as a seam, while Session Manager, Persistence, Delivery, and Plugin contracts still require focused design.

Architectural debt is resolved through DP, ADR when consequential, implementation, and follow-up review. MASTER_PLAN does not settle those contracts itself.

## 10. Things We Will Not Do

UWP will not:

- become an enterprise service bus or a general integration platform;
- introduce a second programmable language or an executable general-purpose DSL around the declarative ConfigurationVersion model;
- turn Runtime into a cluster orchestrator;
- create one universal Policy Engine for unrelated subsystems;
- create one global Decision type for cosmetic uniformity;
- introduce a generic framework before multiple real use cases require the same abstraction;
- let Runtime read Control Plane repositories or mutate Published Configuration;
- hide dependencies in globals, service locators, or `context.Value`;
- give extensions unrestricted access to internal Runtime state;
- ignore unsupported Configuration silently;
- change Core solely to accommodate one concrete integration;
- add distributed complexity to the single-node core without evidence.

These boundaries protect the Configuration-first product model and the Boring Core principle. They do not prevent focused components or extensions when real requirements justify them.

## 11. Success Definition

UWP 1.0 is successful when it has these engineering properties:

- **Predictable:** the same Published Configuration and explicit dependencies produce explainable Runtime behavior.
- **Configuration-driven:** operational behavior is traceable to validated Configuration rather than hidden defaults.
- **Isolated:** independent configured servers have clear resource, lifecycle, and failure boundaries.
- **Safe:** credentials and Secrets remain outside public Configuration, and connection admission occurs at a defined security boundary.
- **Owned:** every mutable resource and goroutine has an owner and a termination path.
- **Observable:** operators can distinguish rejection, configuration failure, dependency failure, and internal failure without sensitive data exposure.
- **Extensible:** supported Providers and Handlers can be added through narrow contracts without changing unrelated Core components.
- **Recoverable:** restart, shutdown, partial startup, and supported failure scenarios have deterministic behavior.
- **Compatible:** public Configuration and API evolution follows documented compatibility and migration rules.
- **Boring:** composition is explicit, dependencies are visible, and abstraction follows demonstrated use cases.

Success is not defined by a feature count, benchmark headline, deployment topology, or integration catalog.

## 12. Living Document

MASTER_PLAN is updated when implemented state, architecture reviews, milestone exit criteria, or engineering priorities materially change. Updates must distinguish completed work, current debt, and future intent.

The document does not replace focused DPs, ADRs, `spec/current-state.md`, or task planning. Future subsystem APIs are designed in their own DPs. Consequential decisions are recorded in ADRs. Actual capability remains documented in current state and verified by reviews.

Documentation changes at different rates:

- MASTER_PLAN is reviewed regularly as engineering evidence changes.
- ARCH documents change rarely and only with architectural justification.
- Vision changes only when the product's fundamental purpose changes.

No revision of MASTER_PLAN alone makes a proposed capability implemented or committed for release.

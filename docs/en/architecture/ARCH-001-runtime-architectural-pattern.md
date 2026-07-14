# ARCH-001: Runtime Architectural Pattern

[Russian version](../../ru/architecture/ARCH-001-runtime-architectural-pattern.md)

**Status:** Active

**Scope:** Runtime architecture

**Stability:**

- This document describes an architectural pattern confirmed by the implemented Alpha vertical.
- It does not freeze future Router, Middleware, Plugin, or Policy contracts.
- It may be revised only when there is an explicit architectural justification supported by implementation experience or subsystem design.

## 1. Purpose

ARCH-001 records a shared way to reason about Runtime components. It provides design questions and boundaries that can be reused when Handshake, Router, Delivery, Persistence, and Plugins are designed, without prescribing their future Go APIs.

The project uses several document types for different purposes:

- **ARCH** describes an architecture-wide design pattern and vocabulary supported by several implemented boundaries.
- **ADR** records a decision about one consequential architecture problem, its context, and its trade-offs.
- **DP** designs one concrete subsystem before its contracts or behavior become stable.
- **Review** assesses the actual implementation against accepted decisions and proposed designs.

ARCH-001 does not replace ADRs or DPs. A consequential choice still requires an ADR, and a new subsystem still requires focused design when its contracts are not established.

## 2. Evidence from Alpha

This pattern is extracted from the implemented Alpha vertical, not from a hypothetical framework:

```text
Published ConfigurationVersion
    -> runtimeconfig.Snapshot
    -> Runtime components
    -> Listener
    -> Connection
    -> Authentication
    -> Session
    -> Runtime Message
    -> Handler
    -> Send
```

The current implementation and [Runtime Alpha Architecture Review](../reviews/runtime-alpha-review.md) confirm these principles:

- Snapshot separates Control Plane data from effective Runtime data and copies nested Provider configuration.
- A Factory converts an `AuthenticationProviderSnapshot` into Provider-specific Runtime configuration before constructing a Provider.
- Listener completes its transport acceptance responsibility by transferring an upgraded connection through `Dispatch`.
- Authentication Dispatcher owns the upgraded connection until it rejects, fails, or explicitly transfers it downstream.
- Session owns the WebSocket connection after successful Authentication.
- Runtime Message contains copied text or binary payload and knows nothing about WebSocket.
- Network resources and application goroutines have identifiable owners and termination paths.
- Host, Listener, and Session lifecycles are explicit, although their contracts are not identical.

This evidence does not mean Runtime Host is already a production composition root. It currently owns a Snapshot and Container and performs state transitions only. Production composition remains an Alpha review finding.

## 3. Core Architectural Pattern

The conceptual pattern is:

```text
Context
    -> Evaluation
    -> Decision
    -> Execution
```

- **Context** provides the minimum information required by a subsystem.
- **Evaluation** determines what action is permitted or required.
- **Decision** expresses the result without performing an unrelated effect.
- **Execution** applies the result through a component that owns the effect and its resources.

Evaluation is intentionally neutral. Depending on the subsystem, it may consist of policies, rules, matchers, Providers, or Handlers. Each subsystem selects the smallest model suited to its responsibility.

ARCH-001 does not require a universal Policy Engine. It also does not require every subsystem to manufacture four separate Go types. The pattern is a design tool for identifying boundaries, not a mandatory framework or generic processing pipeline.

## 4. Context

Context contains only data needed to evaluate one operation.

A well-defined Context:

- contains a minimal data set;
- excludes unrelated transport references;
- performs no business logic;
- passes mutable resources only when ownership and transfer semantics are explicit;
- does not use `context.Value` as a hidden dependency injection container.

Possible contexts include handshake metadata, an authenticated Principal, a Runtime Message, or a future delivery target set. These are conceptual examples, not declarations of required Go types.

The current Authentication adapter demonstrates normalization: it copies selected handshake Headers, Query values, RemoteAddress, and a transport name into `AuthenticationRequest`; Providers do not receive `http.Request` or a WebSocket connection. Conversely, `ConnectionContext` intentionally carries mutable transport references because it represents an explicit ownership handoff, not immutable business data.

## 5. Evaluation

Evaluation answers which action is allowed or required under the effective Configuration and current Context.

Depending on the subsystem, Evaluation may use:

- Authentication Providers;
- future route matchers;
- future origin policies;
- future delivery rules;
- future persistence conditions.

Evaluation must not execute another component's responsibility. Authentication verifies credentials but does not perform WebSocket Upgrade. A future Router may choose a destination or Handler but must not own Session merely because it selected that Session. A Policy must not open sockets, start goroutines, or acquire unrelated mutable resources.

Evaluation may be sequential, compositional, or specialized. The chosen form must follow the subsystem's real use cases rather than a project-wide generic engine.

## 6. Decision

Decision is the result of Evaluation. Examples include:

- allow or reject;
- dispatch to a Handler or destination;
- drop;
- persist;
- select recipients.

Decision does not have to be one global type. Each subsystem may define a small result model that expresses only its own outcomes. A universal Decision object must not be introduced merely to make unrelated subsystems look uniform.

A negative decision and an execution error are different concepts. Rejected credentials, no matching route, or a deliberate persistence skip may be valid outcomes. Resolver unavailability, failed storage, or an unexpected transport failure are operational errors and need separate semantics.

## 7. Execution

An Executor applies a Decision and owns the corresponding effect. Existing or future examples include:

- Listener performs an HTTP rejection or a WebSocket Upgrade;
- Session reads and sends messages;
- Handler processes a Runtime Message;
- a future Storage component persists a Message.

Execution requires:

- an explicit owner;
- a clear lifecycle where resources are involved;
- defined error semantics;
- a graceful completion or shutdown path.

The component evaluating a rule may also execute its own narrowly scoped effect when that remains its single responsibility. ARCH-001 does not require an artificial Executor interface between every function call.

## 8. Configuration First

The governing formulation is:

> Configuration defines Runtime behaviour.
> Published ConfigurationVersion is the source of truth.
> Runtime executes an immutable Snapshot.

Runtime makes operational decisions within rules obtained from Published Configuration. Connection acceptance, Provider results, routing selection, and shutdown behavior are Runtime decisions bounded by effective Configuration and explicit operational dependencies.

The following invariants apply:

- Runtime does not read Control Plane repositories.
- Runtime does not modify Published Configuration.
- Snapshot is an independently copied, stable Runtime view; mutable transport resources are governed by ownership rather than called immutable.
- Unsupported Published settings must not be ignored silently.
- A setting is either supported, or Snapshot construction/Bootstrap/startup rejects it explicitly.

The last invariant directly incorporates the Alpha finding that Listener currently stores TLS and timeout metadata without applying all of it.

## 9. Dependency Direction

Dependency direction is a set of verifiable boundaries, not one mandatory linear package stack.

Required invariants are:

- Runtime packages do not depend on Control Plane repositories.
- Listener does not depend on concrete Authentication Providers.
- Authentication does not depend on `net/http` or WebSocket.
- Session does not depend on Listener or HTTP.
- Message does not depend on Session or WebSocket.
- Extensions do not receive arbitrary access to internal Runtime state.

Conceptual layer diagrams help explain responsibility, but a dependency is acceptable only when its contract and responsibility justify it. For example, `runtimeconfig.Builder` is an explicit adapter to ConfigurationVersion, while the resulting Snapshot has no Repository or HTTP dependency. Session imports the WebSocket library because it owns the authenticated connection; Message does not.

## 10. Ownership

Every mutable resource has one current owner, and ownership transfer must be explicit.

The Alpha vertical provides concrete examples:

- Listener owns the TCP listener and HTTP server.
- Connection Dispatcher receives and transfers an upgraded connection.
- Authentication Dispatcher owns the connection until successful downstream transfer or closure on failure.
- Session owns the WebSocket connection after Authentication.
- Session owns its read loop and serialization of writes.
- Secret Resolver owns stored copies of secret bytes; a successful resolve returns a separate caller-owned copy.

A component may accept ownership of a resource created by another component when the transfer is defined by the contract. Ownership is therefore not limited to resources a component created itself.

An ownership contract must answer who closes the resource on success, rejection, cancellation, construction failure, and partial transfer. Two components must not simultaneously assume that the other will close the same resource.

## 11. Lifecycle and Concurrency

Runtime components follow these confirmed requirements where their responsibility includes lifecycle or concurrency:

- lifecycle is explicit;
- Stop is idempotent;
- restart is not assumed automatically;
- network I/O and long waits do not run while a lifecycle mutex is held;
- every goroutine has an owner and a termination path;
- context cancellation is part of lifecycle;
- concurrent Start, Stop, Run, and Send have defined semantics when those operations exist;
- shutdown errors are not silently lost.

Not every component needs the same complete state machine. A stateless Matcher may have no Start or Stop, while Listener and Session require richer lifecycle contracts. The component contract determines the necessary states.

The Alpha review identified incomplete deadline handling, concurrent Stop semantics, a lifecycle lock held during Session write, and discarded server/Dispatcher errors. These are findings to correct, not behaviors endorsed by this pattern.

## 12. Boring Core

The Runtime core remains compact and predictable.

- A new capability is first considered as a Handler, Provider, Matcher, Policy, Middleware, or Plugin appropriate to its actual responsibility.
- Core changes only when existing contracts are fundamentally insufficient.
- Universal abstractions are not created before several real use cases demonstrate the same requirement.
- A recurring conceptual pattern does not have to become one generic framework.
- One integration does not justify broad access to Runtime internals.

“Boring” means explicit ownership, narrow contracts, and ordinary composition. It does not mean hiding complexity or ignoring necessary lifecycle and error behavior.

## 13. Applying the Pattern

The following examples are cautious applications. They are future design inputs, not implemented contracts.

### Handshake

```text
Handshake metadata
    -> Authentication and Origin Evaluation
    -> allow or reject
    -> HTTP rejection or WebSocket Upgrade
```

The Handshake pipeline requires a future DP. It is not implemented in this form: current Authentication occurs after Upgrade, as recorded by the Alpha review.

### Routing

```text
Message Context
    -> route matchers
    -> selected Handler or destination
    -> Handler execution
```

This illustrates responsibility only. It does not fix a Router API, route model, matcher interface, or error policy.

### Persistence

```text
Message Context
    -> persistence rule
    -> store or skip
    -> storage execution
```

This does not define a Storage contract, transaction model, delivery guarantee, retry rule, or database.

### Delivery

```text
Message and target Context
    -> addressing and delivery rules
    -> recipient set
    -> Session delivery
```

This does not design Groups, Topics, fan-out, backpressure, or Session Manager behavior.

## 14. Architectural Questions

Every new Runtime subsystem must answer:

1. What is its single responsibility?
2. Which data forms its minimal Context?
3. What performs Evaluation?
4. How is Decision expressed?
5. Who executes the effect?
6. Who owns the resources?
7. How does the component start and stop?
8. Can the capability use an existing extension point?
9. Does Control Plane state leak into Runtime?
10. Are we creating a universal abstraction before real use cases exist?

Answers should be concrete enough to test and review. “The framework handles it” is not an ownership, lifecycle, or error-semantics answer.

## 15. Anti-Patterns

The following directions are prohibited or strongly discouraged:

- a universal Policy Engine for every subsystem;
- one global Decision type;
- a god-object Runtime Host;
- Session that knows `http.Request`;
- Message that knows `websocket.Conn`;
- Provider that knows Snapshot;
- Listener containing JWT- or API Key-specific logic;
- hidden dependencies through `context.Value`;
- detached goroutines without an owner and termination path;
- silent ignoring of unsupported Configuration;
- a premature generic Module framework;
- changing Core for one concrete integration.

Runtime Host as composition root remains responsible for wiring and lifecycle coordination, but it must delegate subsystem behavior rather than absorb it.

## 16. Relationship to Future Documents

- Handshake Pipeline will be designed in a separate DP.
- Router will receive its own DP.
- A Plugin ABI requires a separate DP and likely an ADR because compatibility and isolation create long-term constraints.
- ARCH-001 does not define any of those APIs.
- A future DP must explain how its subsystem follows ARCH-001 or why a justified exception is necessary.
- If several valid DPs cannot fit ARCH-001, the architecture guide must be reviewed explicitly instead of being bypassed through hidden exceptions.

Future documents should cite concrete evidence and preserve the distinction between proposed contracts and implemented behavior.

## 17. Final Statement

> Configuration defines Runtime behaviour.
> Runtime evaluates context, makes bounded decisions, and executes them through explicit owners.

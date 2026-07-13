# ADR 0003: Runtime Architecture

## Status

Accepted.

## Context

ADR-0002 established ConfigurationVersion as the declarative Configuration DSL. The next architectural boundary is the execution model for that DSL.

Runtime executes a Published ConfigurationVersion. It is not a source of Configuration and must not develop a second configuration model, hidden defaults, or dependencies on the management API and its persistence implementation.

The architecture must also allow Authentication and future capabilities to evolve as replaceable components without coupling them to a transport, Repository, or concrete Provider implementation.

## Decision

Runtime is composed from independent components with one responsibility each. Components receive their dependencies explicitly through dependency injection and communicate through stable contracts.

The Runtime core does not depend on the HTTP API or Repository implementations. Integration adapters and the composition root provide the Published ConfigurationVersion and concrete dependencies at startup.

### Runtime Pipeline

Runtime uses the following conceptual pipeline:

```text
Configuration Loader
        |
        v
Configuration Snapshot
        |
        v
Secret Resolver
        |
        v
Authentication Service
        |
        v
Authentication Provider Registry
        |
        v
Authentication Providers
        |
        v
Principal
        |
        v
Authorization
        |
        v
Routing
        |
        v
Storage
        |
        v
Monitoring
```

This pipeline defines component boundaries and composition order. It does not require every future request or operation to pass through every component, and it does not define transport behavior.

### Runtime Principles

- **Single Responsibility:** each component has one explicit purpose.
- **Composition over inheritance:** Runtime behavior is assembled from collaborating components rather than type hierarchies.
- **Dependency Injection:** dependencies are supplied explicitly at composition time.
- **Stateless services where possible:** operational state is introduced only where the responsibility requires it.
- **Immutable Configuration Snapshot:** Runtime Services observe one stable view of Configuration.
- **No hidden configuration:** every behavior-affecting setting originates from the Published ConfigurationVersion or an explicit operational dependency.

### Configuration Snapshot

Runtime never reads a Draft ConfigurationVersion. It starts only from a Published ConfigurationVersion supplied through the Configuration Loader boundary.

After loading, Runtime creates an immutable Configuration Snapshot. All Runtime Services consume only that Snapshot and cannot mutate it. Reload, replacement, and process lifecycle semantics require separate decisions; they must not change a Snapshot in place.

The Configuration Loader is an abstraction at the Runtime boundary. It does not give Runtime Services access to a Repository or the HTTP API.

### Secret Resolver

Configuration and Configuration Snapshot contain only Secret References, never real Secret values. The Secret Resolver resolves required references while Runtime is starting.

Resolved Secrets exist only in process memory and are never written back to Configuration or the Snapshot. They must not be logged, persisted, exported, or included in error details. Their lifecycle must be limited to the components that require them.

Reference syntax, storage separation, and security rules are defined by [DP-002: Secret References](../proposals/DP-002-secret-references.md).

### Provider Registry

Runtime does not know concrete Provider implementations. An Authentication Provider Registry maps effective Configuration to Provider construction and supplies composed Providers to Authentication Service.

The Registry owns Provider selection and construction. Authentication Service depends on Provider contracts, not concrete JWT, API Key, Basic, or future Plugin Provider types.

### Authentication

Authentication uses the transport-neutral contracts proposed by [DP-004: Authentication Runtime Contracts](../proposals/DP-004-authentication-runtime-contracts.md).

Authentication does not know the transport and does not depend on WebSocket. Transport adapters normalize input before invoking Authentication. Authentication returns an explicit result and immutable Principal for later Authorization.

### Dependency Injection

The composition root creates Runtime components, resolves their dependencies, and connects the pipeline. Services do not locate dependencies through globals, service locators, Repository access, or the HTTP API.

This decision requires dependency injection as an architectural technique. It does not select a dependency injection framework.

### Future Runtime Components

The architecture reserves independent component boundaries for:

- Authorization;
- Routing;
- Storage;
- Monitoring;
- Logging;
- Plugins.

Their detailed contracts and behavior require separate proposals or decisions before implementation.

## Consequences

### Benefits

- Components can be tested independently with explicit dependencies.
- Implementations can be replaced without changing unrelated services.
- Composition and dependency injection remain straightforward.
- Coupling between Runtime, transport, persistence, and Providers is reduced.
- New components and Provider implementations can be added behind stable contracts.
- The structure supports a future Plugin Architecture.
- Runtime behavior remains traceable to an immutable Published ConfigurationVersion.

### Drawbacks

- The architecture introduces more interfaces and components.
- Composition and lifecycle management require explicit design.
- Contract compatibility must be maintained rigorously.
- Diagnostics must preserve component boundaries while still explaining failures across the pipeline.

## References

- [ADR-0001: Bootstrap Control Service](0001-bootstrap-control-service.md)
- [ADR-0002: Configuration DSL](0002-configuration-dsl.md)
- [DP-001: Authentication](../proposals/DP-001-authentication.md)
- [DP-002: Secret References](../proposals/DP-002-secret-references.md)
- [DP-003: JWT Provider](../proposals/DP-003-jwt-provider.md)
- [DP-004: Authentication Runtime Contracts](../proposals/DP-004-authentication-runtime-contracts.md)


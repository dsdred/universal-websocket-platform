# ADR 0002: Configuration DSL

## Status

Accepted.

## Context

The project has developed a declarative ConfigurationVersion model through the previous milestones. Listener, TLS, timeout, and Authentication metadata are represented as explicit sections that can be created, validated, versioned, published, and inspected independently of their future Runtime behavior.

This model is becoming a public contract shared by management APIs, persistence, import and export, and future tooling. The architectural role of ConfigurationVersion must therefore be explicit before Runtime implementation begins.

## Decision

ConfigurationVersion is the declarative domain-specific language (DSL) for describing a WebSocket server.

Runtime does not construct or maintain a separate configuration model. Runtime executes a Published ConfigurationVersion.

### Configuration First

Every new platform capability is represented in the Configuration DSL before its Runtime behavior is implemented. Runtime work must not introduce configuration that is absent from the public ConfigurationVersion model.

### Metadata Before Behavior

A capability is developed in this order:

1. metadata design;
2. validation;
3. persistence;
4. Runtime behavior.

This sequence keeps the public model explainable and reviewable before operational behavior depends on it.

### Published Configuration

A Published ConfigurationVersion is the single source of truth for Runtime. Runtime never modifies a Published ConfigurationVersion.

Changes are prepared in a Draft and become available to Runtime only through the ConfigurationVersion lifecycle. Runtime-derived state does not become hidden Configuration.

### Independent Sections

Every top-level ConfigurationVersion section is independent. Expected sections include:

- Listener;
- Authentication;
- Authorization;
- Routing;
- Storage;
- Monitoring;
- Logging.

Changing or extending one section should not require unrelated schema changes in the others. Cross-section validation may express explicit invariants, but must not merge independent concerns into one hidden model.

### Backward Compatibility

The public Configuration DSL schema evolves in a backward-compatible way. Existing published documents and API clients must remain interpretable when optional capabilities are added.

An incompatible schema change requires a new ADR that explains the migration and compatibility strategy.

### Secrets

Configuration stores only Secret References and never stores real secret values. Secret values, private keys, passwords, tokens, and similar sensitive material are outside the Configuration DSL.

The reference rules and future resolution boundary are defined by [DP-002: Secret References](../proposals/DP-002-secret-references.md).

### Runtime

Runtime is an executor of the Configuration DSL. It does not contain hidden configuration, alternate defaults that contradict the public model, or a second independently evolving schema.

Runtime may produce operational state and diagnostics, but those are not Configuration and must not mutate the Published ConfigurationVersion.

## Consequences

### Benefits

- One model for the REST API.
- One model for YAML representation.
- One model for Import and Export.
- One model for PostgreSQL persistence.
- One model for a future Admin UI.
- One model for a future Terraform Provider.
- One model for a future Kubernetes Operator.
- Runtime behavior remains traceable to an explicit Published ConfigurationVersion.

### Drawbacks

- The Configuration DSL becomes a public contract.
- Schema changes require strong design, validation, compatibility, and migration discipline.
- Runtime implementation may need to wait until the relevant metadata and lifecycle semantics are stable.

## References

- [ADR-0001: Bootstrap Control Service](0001-bootstrap-control-service.md)
- [DP-001: Authentication](../proposals/DP-001-authentication.md)
- [DP-002: Secret References](../proposals/DP-002-secret-references.md)
- [DP-003: JWT Provider](../proposals/DP-003-jwt-provider.md)

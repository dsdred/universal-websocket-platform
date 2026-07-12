# DP-002: Secret References

**Status:** Proposed

This document is the official English translation. The Russian version is the primary engineering document.

## Motivation

ConfigurationVersion is versioned, exposed through management APIs, and may later be exported or presented in Admin UI. Storing secret values inside Configuration would allow them to leak into Git, JSON or YAML exports, logs, version history, and Admin UI responses. Once copied into those systems, a secret cannot be reliably removed.

Embedding secret values also prevents safe rotation and conflicts with immutable Published ConfigurationVersion entities: changing a credential would require either mutating an immutable version or publishing a configuration change unrelated to its structure.

ConfigurationVersion therefore MUST NOT contain real API keys, JWT symmetric secrets, private keys, PEM certificate contents, OAuth2 client secrets, passwords, or any other secret values.

ConfigurationVersion MAY contain only references such as `secretRef`, `certificateRef`, `privateKeyRef`, and other typed references to a future Secret Storage.

## Goals

- Separate configuration metadata from secret values.
- Keep the public JSON model stable and free of secret material.
- Allow Secret Storage implementations to change without changing the Configuration schema.
- Support future secret rotation without embedding new values in ConfigurationVersion.
- Permit future backends including in-memory storage, PostgreSQL, filesystem, environment variables, HashiCorp Vault, Kubernetes Secrets, and cloud secret managers.

## Non Goals

This proposal does not design:

- a concrete Secret Storage API;
- encryption at rest;
- Secret Storage RBAC;
- rotation workflows;
- caching;
- HashiCorp Vault integration;
- Kubernetes integration;
- UI for secret management.

## Terminology

- **Secret** is sensitive data used to authenticate, sign, encrypt, decrypt, or otherwise establish trust. Examples include passwords, API keys, tokens, and private keys.
- **Secret Reference** is a stable, non-secret string that identifies a Secret or another protected object without containing its value.
- **Secret Storage** is a backend responsible for storing and protecting Secret values.
- **Secret Resolver** is a component that resolves a Secret Reference against Secret Storage and returns the Secret to an authorized Runtime operation.
- **Secret Type** identifies the expected purpose or representation of a referenced Secret, such as an API key, JWT symmetric key, certificate, or private key.

## Reference Format

A Secret Reference:

- is a string;
- does not contain the Secret value;
- remains stable while the referenced value may change;
- is resolved by Runtime or a separate Secret Resolver;
- does not have to be verified for existence while a Draft is edited;
- is fully verified before publication or before Runtime starts.

This proposal intentionally does not define one URI-like reference format. A common syntax may be selected after concrete Secret Storage and namespace requirements are known.

Valid examples:

```text
secrets/api-keys/internal
secrets/jwt/main
certificates/default/server
workspace/main/oauth/client-secret
```

Invalid examples:

```text
actual-secret-value
-----BEGIN PRIVATE KEY-----
C:\certs\server.key
https://example.com/secret
```

The first invalid example represents a value rather than a reference. The remaining examples expose secret material, couple Configuration to a local filesystem path, or introduce a remote location format that has not been accepted.

## Runtime Behavior

The future resolution flow is:

```text
ConfigurationVersion
        ↓
Secret Reference
        ↓
Secret Resolver
        ↓
Secret Storage
        ↓
Resolved Secret in memory
```

Runtime MUST NOT:

- write a resolved Secret to logs;
- return a resolved Secret through the management API;
- save a resolved Secret back into Configuration;
- include a Secret value in errors.

Resolution must occur only for an authorized operation and as close as practical to the point of use.

## Versioning

A Published ConfigurationVersion remains immutable. Its `secretRef` remains unchanged, while the Secret value stored behind that reference MAY change. This separation allows rotation without changing the public JSON schema or rewriting published configuration history.

Whether Runtime observes a rotated value through hot reload, resolves it only at startup, or pins a specific Secret version remains an open question.

## Security Principles

- Apply least privilege to Secret resolution and access.
- No Secret values in Configuration.
- No Secret values in logs.
- No Secret values in exports.
- No Secret values in error messages.
- Keep resolved Secret values in memory only as long as needed.

## Existing Model Compatibility

The existing `certificateRef` and `privateKeyRef` fields in TLSSettings conform to this proposal: they store references and do not contain PEM data, private-key contents, or filesystem locations.

Future API Key, JWT, and OAuth2 settings must use `secretRef` or another typed Secret Reference. They must not introduce fields that hold the underlying credential value.

## Open Questions

- Is one common reference format required?
- Should references include a Provider namespace?
- At which lifecycle stage should Secret existence be verified?
- How should rotation be performed and propagated?
- Is Secret version pinning required?
- Which component or identity is allowed to resolve references?
- How should resolved Secrets be cached?
- How should temporary Secret Storage unavailability be handled?

# DP-004: Authentication Runtime Contracts

**Status:** Proposed

This document is the official English translation. The Russian version is the primary engineering document.

## Motivation

The Configuration DSL already describes Authentication metadata, but the platform does not yet have transport-neutral contracts for applying that policy. These contracts must separate Authentication orchestration and Provider implementations from transport adapters and future Runtime internals.

This proposal defines logical boundaries and data exchanged during Authentication. It does not design or implement Runtime behavior.

## Goals

- Define a transport-neutral AuthenticationRequest.
- Define an immutable Principal representing a successfully authenticated or explicitly anonymous client.
- Define AuthenticationResult and a structured failure model.
- Define the boundary between AuthenticationService and AuthenticationProvider implementations.
- Keep future Provider implementations independent of transport, persistence, and ConfigurationVersion internals.

## Design Principles

The contracts follow ADR-0002:

- Configuration describes Authentication policy before behavior is implemented.
- AuthenticationService consumes an effective Configuration but does not expose ConfigurationVersion to Providers.
- Provider contracts contain no transport-specific request or connection types.
- Authentication output is explicit and can be passed to Authorization without hidden state.

The models in this document are logical contracts. Exact language types, serialization formats, and package boundaries will be decided by a future implementation task.

## AuthenticationRequest

AuthenticationRequest is a normalized view of client input. A transport adapter constructs it before Authentication begins.

| Field | Purpose |
| --- | --- |
| `Headers` | Case-insensitive named values supplied by the client. Multiple values per name are preserved. |
| `Query` | Named request parameters. Multiple values per name are preserved. |
| `RemoteAddress` | The client address as observed by the trusted transport boundary. |
| `Transport` | A stable transport identifier, not a transport implementation object. |
| `Attributes` | Trusted, transport-neutral input attributes added by platform adapters. |
| `Body` | Optional opaque input for Providers that explicitly require it. Its absence is distinct from an empty value. |
| `RequestContext` | Cancellation, deadline, correlation, and request-scoped control information without exposing a transport request object. |

AuthenticationRequest MUST NOT contain an HTTP request, a transport connection, a session, or Runtime implementation state. Normalization gives every Provider the same input shape, prevents transport APIs from leaking into Provider code, and makes contracts independently testable.

Headers, Query values, Body, and Attributes are untrusted unless a platform boundary explicitly marks their origin as trusted. Implementations must not log credential-bearing values by default.

## Principal

Principal represents the identity produced by successful Authentication or the explicit anonymous identity used when Authentication is disabled.

| Field | Purpose |
| --- | --- |
| `ID` | Stable identity identifier within the applicable identity domain. |
| `Name` | Human-readable identity name when one is available. |
| `AuthenticationType` | Authentication mechanism that established the identity. |
| `Claims` | Normalized identity assertions produced by Authentication. |
| `Roles` | Role names supplied as identity context for later Authorization decisions. |
| `Attributes` | Additional normalized identity attributes. |
| `Anonymous` | Indicates the explicit anonymous Principal. |
| `Authenticated` | Indicates that credentials were successfully authenticated. |
| `Metadata` | Non-secret diagnostic and provenance metadata about identity creation. |

After successful Authentication, Principal is immutable. AuthenticationService, Providers, and downstream consumers MUST NOT modify it in place. This prevents identity from changing between Authentication and Authorization and makes concurrent use predictable.

`Anonymous` and `Authenticated` express different facts and must remain consistent. A credential-authenticated Principal has `Authenticated` set and is not anonymous. An explicit anonymous Principal is anonymous and is not credential-authenticated. Exact anonymous defaults remain an open question.

Claims, Roles, Attributes, and Metadata are normalized contract data, not access decisions. Their presence does not grant permission; Authorization remains responsible for interpreting them.

## AuthenticationResult

AuthenticationResult records the outcome of one Provider attempt or the final AuthenticationService decision.

| Field | Purpose |
| --- | --- |
| `Success` | Indicates whether Authentication produced an acceptable Principal. |
| `Principal` | The immutable Principal on success; absent on failure. |
| `FailureReason` | Structured, non-secret failure category and safe diagnostic information. |
| `ProviderName` | Configuration name of the Provider responsible for the result. |
| `ProviderType` | Provider type associated with the result. |
| `Metadata` | Non-secret result metadata for diagnostics, audit integration, and observability. |

When `Success` is true, Principal MUST be present and FailureReason MUST be absent. When `Success` is false, an authenticated Principal MUST NOT be returned. Metadata must never contain credentials, resolved Secrets, tokens, passwords, or private key material.

## AuthenticationProvider

AuthenticationProvider has one conceptual operation:

```text
Authenticate(AuthenticationRequest) -> AuthenticationResult
```

AuthenticationService supplies normalized input and receives an explicit result. A Provider implements one Authentication mechanism but does not decide access rights.

A Provider does not know about:

- Runtime;
- WebSocket or another transport protocol;
- Storage or Repository implementations;
- ConfigurationVersion domain objects;
- Authorization policy.

Provider-specific settings are transformed into an effective Provider during composition. They are not read from repositories during `Authenticate`.

## AuthenticationService

AuthenticationService is the orchestration boundary. It receives the effective Authentication policy and ordered Providers, invokes Providers according to that policy, interprets their results, and returns the final AuthenticationResult and Principal when successful.

AuthenticationService does not contain JWT, API Key, or Basic verification logic. It does not grant permissions and does not create transport sessions.

## Authentication Pipeline

```text
Configuration
      |
      v
AuthenticationService
      |
      v
AuthenticationProvider(s)
      |
      v
AuthenticationResult
      |
      v
Principal
      |
      v
Authorization
```

Configuration selects and orders Authentication policy. AuthenticationService orchestrates the configured Providers. Providers evaluate normalized input and return AuthenticationResult. A successful result exposes an immutable Principal to Authorization.

This sequence describes contract boundaries only. It does not define Runtime construction, lifecycle, or transport behavior.

## Failure Model

AuthenticationResult must distinguish these outcomes:

| Outcome | Meaning |
| --- | --- |
| `Success` | Credentials were accepted, or the configured anonymous policy produced an explicit Principal. |
| `Rejected credentials` | Input was understood but did not prove an acceptable identity. |
| `Provider error` | A Provider could not complete its operation because of a Provider-specific operational failure. |
| `Configuration error` | Effective Authentication policy or Provider setup is invalid or unusable. |
| `Internal error` | An unexpected platform failure prevented a reliable decision. |

Rejected credentials are an expected Authentication decision and must not be conflated with operational failures. Provider, Configuration, and Internal errors must remain distinguishable for safe diagnostics and future policy decisions. FailureReason exposed outside trusted boundaries must not reveal credential or Secret details.

This proposal does not define status codes, transport responses, retry policy, or whether the pipeline continues after each failure category.

## Explicitly Out of Scope

- SecretResolver;
- JWT verification;
- Basic verification;
- API Key verification;
- Runtime design or implementation;
- Authorization;
- Session and Connection models;
- rate limiting;
- caching.

## Relationship to Existing Documents

- [ADR-0002](../adr/0002-configuration-dsl.md) defines ConfigurationVersion as the declarative Configuration DSL and the source of truth for future execution.
- [DP-001](DP-001-authentication.md) defines the high-level Authentication architecture and Provider ordering.
- [DP-002](DP-002-secret-references.md) defines the boundary between Configuration and secret values.
- [DP-003](DP-003-jwt-provider.md) defines proposed JWT Provider Configuration metadata.

DP-004 refines the future runtime-facing data contracts without changing the Configuration models proposed or implemented by those documents.

## Open Questions

- Should Provider timeout be global, per Provider, or both?
- How should cancellation be propagated and represented by RequestContext?
- Will asynchronous Providers be supported?
- Which Provider metrics are required, and which labels are safe?
- Is distributed Authentication orchestration required?
- Which audit events belong to AuthenticationService and which belong to Providers?
- How should challenge support be represented without coupling results to a transport?
- What are the required fields and values of the anonymous Principal?


# DP-001: Authentication

**Status:** Proposed

## Motivation

Authentication establishes the identity of a client before a WebSocket connection is opened. It answers the question: “Who is this client?”

Authentication does not determine what the authenticated client is allowed to do. Access decisions belong to Authorization and must remain a separate concern.

## Goals

The initial Authentication design should support:

- API Key
- JWT
- Basic

OAuth2 may be added in a future version.

## Non Goals

The initial scope does not include:

- OpenID Connect
- LDAP
- Kerberos
- SAML
- mTLS Authentication
- External Secret Manager integration

## High Level Architecture

```text
Client
  ↓
Handshake
  ↓
Authentication
  ↓
Principal
  ↓
Authorization
  ↓
WebSocket Session
```

- **Client** initiates a connection and supplies credentials through supported handshake inputs, such as headers.
- **Handshake** receives the initial connection request and creates a transport-neutral AuthenticationRequest.
- **Authentication** evaluates the supplied credentials through configured Authentication Providers.
- **Principal** represents the established client identity and the attributes produced by Authentication.
- **Authorization** evaluates whether the Principal may perform the requested action. It is separate from Authentication.
- **WebSocket Session** is created only after Authentication and Authorization succeed.

The connection must not be upgraded to a WebSocket Session before the required identity and access checks complete.

## AuthenticationRequest

Authentication must not depend on `net/http`. A transport-neutral request model is proposed:

```go
type AuthenticationRequest struct {
    Headers       map[string][]string
    RemoteAddress string
    TLS           bool
}
```

- `Headers` contains the handshake header values needed by Authentication Providers.
- `RemoteAddress` identifies the client network address as observed by Runtime.
- `TLS` reports whether the incoming transport is protected by TLS.

Runtime should not pass `http.Request` directly because that would couple Authentication contracts and Provider implementations to one transport adapter. A dedicated model keeps Authentication testable, prevents Providers from depending on unrelated HTTP state, and allows future Runtime transports to reuse the same Authentication pipeline.

Runtime is responsible for copying only the required request data into AuthenticationRequest. Sensitive values must not be logged as part of this conversion.

## Principal

The authenticated identity is represented by:

```go
type Principal struct {
    ID                     string
    AuthenticationProvider string
    Roles                  []string
    Claims                 map[string]string
}
```

- `ID` is the stable identity assigned to the authenticated client.
- `AuthenticationProvider` is the name of the Provider that established the identity, such as `jwt`, `api-key`, or `basic`.
- `Roles` contains role names supplied by Authentication for later Authorization evaluation. Authentication does not interpret access rules for these roles.
- `Claims` contains additional normalized identity attributes that downstream components may inspect.

`Claims` uses `map[string]string` to keep the core Principal deterministic, transport-neutral, and independent of Provider-specific dynamic value types. Providers must normalize selected claims into strings instead of exposing raw JWT structures or arbitrary nested data. This deliberately limits the initial claim model; richer typed claims may be considered later if concrete requirements justify them.

## Authentication Provider

Each credential mechanism is implemented behind a common interface:

```go
type AuthenticationProvider interface {
    Name() string

    Authenticate(
        ctx context.Context,
        request AuthenticationRequest,
    ) (*Principal, error)
}
```

`Name` returns the stable Provider identifier used for configuration, diagnostics, and Principal.AuthenticationProvider.

`Authenticate` evaluates one request and returns a Principal only when that Provider successfully establishes an identity. Provider errors must not expose credentials or secret material.

Runtime depends on this interface and AuthenticationService. It must not contain JWT parsing, API Key lookup, or Basic credential verification logic.

## AuthenticationService

AuthenticationService coordinates Authentication Providers. It:

- receives the configured Provider list;
- orders Providers by priority;
- invokes Providers in that order;
- returns the first successful Principal;
- stops evaluation after the first success;
- returns an Authentication error if no Provider succeeds.

AuthenticationService owns orchestration only. It does not contain JWT validation, API Key verification, Basic credential checks, or Provider-specific configuration rules.

The service should validate Provider names and priorities when ConfigurationVersion is prepared for use. Duplicate or unavailable Provider definitions must produce an explainable configuration error rather than an unpredictable runtime order.

## Provider Ordering

Multiple Providers may be enabled at the same time. Their order is declared in Configuration:

```yaml
authentication:
  providers:
    - type: jwt
      priority: 10
    - type: api-key
      priority: 20
    - type: basic
      priority: 30
```

A lower numeric priority executes first. In this example, JWT is evaluated before API Key, and API Key is evaluated before Basic.

Evaluation stops after the first successful result. Providers after the successful Provider are not invoked. The behavior for equal priorities must be defined before implementation; Configuration should not rely on declaration order unless that rule is explicitly accepted.

## Disabled Authentication

Authentication may be disabled explicitly:

```yaml
authentication:
  enabled: false
```

When disabled, Runtime receives a special Principal with an ID such as `anonymous`:

```text
ID = "anonymous"
```

Using an anonymous Principal instead of `nil` gives downstream Authorization and session components a consistent identity object. It avoids repeated nil handling, makes anonymous behavior explicit in logs and decisions, and allows Authorization to apply policies to unauthenticated clients without bypassing its normal evaluation path.

The anonymous Principal must be distinguishable from identities returned by configured Providers and must not receive implicit privileges.

## Error Handling

If no configured Provider can authenticate the client, the handshake returns:

```text
401 Unauthorized
```

The handshake then terminates and the WebSocket connection is not opened.

Client responses must use a stable, non-sensitive error message. Detailed Provider diagnostics may be recorded for operators, but logs must not contain credentials, API Keys, tokens, passwords, or secret values.

The implementation must distinguish an expected credential rejection from an internal Provider failure for diagnostics, while preventing internal errors and stack traces from reaching the client.

## Configuration

Authentication is a separate section of ConfigurationVersion:

```yaml
authentication:
  enabled: true
  providers:
    - type: jwt
      priority: 10
```

Authentication is not part of ListenerSettings. ListenerSettings describes how a future listener accepts connections; Authentication describes how client identity is established during the handshake.

Provider-specific settings should remain inside the corresponding Provider configuration and must not leak into Runtime orchestration contracts.

## Future Extensions

Possible future extensions include:

- OAuth2
- OpenID Connect
- LDAP
- Kerberos
- SAML
- Plugin Providers
- External Secret Managers

This proposal does not define their implementation.

## Open Questions

- Should refresh tokens be supported, and which component would own their lifecycle?
- Is token introspection required for opaque tokens?
- Is a distributed Authentication cache required across Runtime instances?
- How should credential and token revocation be propagated?
- How should secret rotation affect active and new handshakes?
- How should AuthenticationService distinguish “credentials not applicable” from “credentials rejected” when evaluating multiple Providers?
- Should equal Provider priorities be rejected or resolved by a deterministic secondary rule?
- Which claims should be normalized into Principal.Claims, and how should claim-name collisions be handled?
- What identity and Provider values should be assigned to the anonymous Principal?

## Out of Scope

This document does not specify:

- Authorization
- Session Management
- Runtime implementation
- WebSocket Protocol behavior
- Secret Storage

Those concerns require separate specifications and decisions.

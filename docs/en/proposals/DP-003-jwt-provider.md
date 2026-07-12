# DP-003: JWT Provider

**Status:** Proposed

This document is the official English translation. The Russian version is the primary engineering document.

## Motivation

A JWT Provider needs a declarative Configuration model that describes which JSON Web Tokens are acceptable. Configuration defines verification policy: trusted Signing Keys, permitted algorithms, expected issuers and audiences, required Claims, and allowed clock skew.

Configuration does not define the internal implementation of JWT parsing, cryptographic verification, Claims processing, or Runtime orchestration.

## Goals

- Define Configuration metadata for a future JWT Provider.
- Support multiple trusted Signing Keys.
- Support multiple allowed algorithms, issuers, and audiences.
- Declare required Standard Claims and Custom Claims.
- Keep secret material outside ConfigurationVersion.
- Keep the model independent of a particular JWT library or Runtime implementation.

## Proposed Model

The future JWT Provider Configuration is expected to contain a JWTSettings section with this logical shape:

```text
JWTSettings
├── SigningKeys[]
├── AllowedAlgorithms[]
├── AllowedIssuers[]
├── AllowedAudiences[]
├── RequiredClaims[]
└── ClockSkewSeconds
```

- `SigningKeys` identifies one or more trusted signing keys through Secret References.
- `AllowedAlgorithms` lists every signing algorithm accepted by policy.
- `AllowedIssuers` lists accepted values of the standard `iss` Claim.
- `AllowedAudiences` lists accepted values of the standard `aud` Claim.
- `RequiredClaims` declares Claims that must be present and may include declarative value constraints.
- `ClockSkewSeconds` defines the permitted time tolerance for time-based Standard Claims.

All collections support multiple values. The exact Go and JSON representation, required fields, defaults, duplicate handling, and empty-list semantics must be fixed by an implementation task before the model becomes a stable API contract.

## Signing Keys

ConfigurationVersion MUST store Signing Keys only as Secret References. It MUST NOT contain:

- PEM content;
- embedded JWK documents;
- HMAC secret values.

A logical Signing Key entry contains a `secretRef` and may later contain non-secret metadata needed for deterministic key selection. The Secret value is resolved outside Configuration according to DP-002.

Multiple Signing Keys must be supported so that a policy can trust more than one active key and can prepare for key rotation. This proposal does not define how keys are loaded, cached, selected, or rotated.

## Algorithms

The initial Configuration model should recognize these algorithm identifiers:

- `HS256`, `HS384`, `HS512`;
- `RS256`, `RS384`, `RS512`;
- `ES256`, `ES384`, `ES512`;
- `PS256`, `PS384`, `PS512`.

`AllowedAlgorithms` is an explicit allowlist. Runtime must not infer an algorithm from key material or accept an algorithm merely because a JWT library supports it. Algorithm negotiation behavior is not defined by this proposal.

## Claims

JWT Claims are divided into two policy groups.

### Standard Claims

Standard Claims are registered JWT Claims with defined meanings, including `iss`, `sub`, `aud`, `exp`, `nbf`, `iat`, and `jti`. Dedicated fields such as `AllowedIssuers`, `AllowedAudiences`, and `ClockSkewSeconds` express policy for the Standard Claims that require specialized handling.

The future model must define which time-based Claims are mandatory by default and which must be named explicitly in `RequiredClaims`.

### Custom Claims

Custom Claims are application- or organization-specific Claims outside the registered set. Configuration must allow them to be required declaratively without embedding executable expressions or Provider implementation details.

A Required Claim policy should identify a Claim name and may later define non-secret constraints such as expected string values or allowed value sets. The exact constraint model remains open and must preserve a stable, explainable JSON representation.

## Validation Policy

Configuration must be able to declare:

- multiple issuers;
- multiple audiences;
- multiple Signing Keys;
- multiple allowed algorithms;
- multiple required Standard Claims;
- multiple required Custom Claims.

Validation of the Configuration model is distinct from validation of an incoming token. Before publication or Runtime startup, the platform should be able to reject internally inconsistent policy, unsupported algorithm identifiers, malformed Secret References, and ambiguous duplicate entries without resolving or exposing Secret values.

The matching semantics for multiple issuers and audiences, and whether all or any configured values must match, require an explicit decision before implementation.

## Runtime Notes

This proposal does not design Runtime. The conceptual future processing sequence is included only to show where the Configuration policy applies:

```text
Token
  ↓
Signature verification
  ↓
Claims validation
  ↓
Principal
  ↓
Authorization
```

Signature verification uses configured Signing Keys and AllowedAlgorithms. Claims validation applies issuer, audience, required-Claim, and time policies. Principal construction and Authorization are separate concerns and are not specified here.

## Explicitly Out of Scope

This proposal does not design:

- OpenID Connect;
- JWKS Discovery;
- OAuth2;
- Refresh Token handling;
- token Introspection;
- token Revocation;
- a user database;
- Authorization;
- Principal;
- Runtime implementation.

## Open Questions

- Should embedded or referenced JWK be supported in a future version?
- Should a JWKS URL be supported, and how would it comply with Secret Reference rules?
- How should `kid` select among multiple Signing Keys?
- Is any algorithm negotiation allowed, or must policy selection always be explicit?
- What should the default clock skew be?
- Should nested JWT be supported?
- Should encrypted JWT be supported?
- How are multiple active keys represented and selected?
- What rotation strategy should be used for Signing Keys?

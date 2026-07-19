# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- ListenerSettings metadata with validated Host and Port for Draft ConfigurationVersion entities.
- TLSSettings metadata with certificate references and minimum TLS version for Draft ConfigurationVersion entities.
- Listener TimeoutSettings metadata for handshake, read, write, and idle limits in seconds.
- Authentication Domain Model metadata with configurable JWT, API Key, and Basic Provider entries for Draft ConfigurationVersion entities.
- API Key Provider metadata with validated HTTP header names and Secret References.
- JWT Provider metadata with Signing Keys, allowed algorithms, issuers, audiences, required Claims, and clock skew policy.
- Basic Authentication Provider metadata with Realm and Secret Reference.
- Dedicated AuthenticationValidator component for Authentication metadata business validation.
- Immutable Runtime Configuration Snapshot model and Builder for Published ConfigurationVersion entities.
- Runtime dependency Container that owns and exposes independent Configuration Snapshot copies.
- Thread-safe Runtime Host composition root with immutable Snapshot ownership and Created, Running, and Stopped lifecycle states.
- Listener Bootstrap with immutable TLS configuration, a thread-safe lifecycle, and an HTTP layer that returns `501 Not Implemented` for every request.
- RFC 6455 WebSocket Upgrade endpoint at `GET /ws` with an immediate normal closure after a successful handshake.
- Transport-only ConnectionContext and injectable Connection Dispatcher between WebSocket Upgrade and future Runtime components.
- Authentication Dispatcher that converts handshake metadata into transport-neutral Authentication requests and forwards successful connections with an immutable Principal context.
- Minimal WebSocket Session and Session Dispatcher with cryptographically random identifiers, safe transport metadata, immutable Principal copies, and explicit connection ownership after Authentication.
- Transport-neutral immutable Runtime Message model and blocking Session read loop for text and binary WebSocket application messages.
- Thread-safe Session outbound writer that accepts only Runtime Message values and serializes text and binary WebSocket writes.
- Transport-neutral Runtime Message Handler contract and Echo Handler that returns incoming text and binary messages through Session Send.
- Storage-neutral Secret Resolver contract and thread-safe in-memory implementation for tests and local development.
- Extensible runtime Authentication Provider and Factory contracts with transport-neutral request, result, and Principal models.
- Thread-safe Authentication Provider Registry with Factory registration and delegated Provider creation.
- Runtime API Key Authentication Provider with case-insensitive header lookup and constant-time credential comparison.
- Sequential Authentication Service that evaluates Providers in order and stops after the first successful result.
- Authentication Bootstrap that assembles an ordered Service from Authentication Snapshot, Provider Registry, and Secret Resolver.
- Production API Key Factory that converts Authentication Provider Snapshot metadata into Provider-local runtime configuration.
- Runtime JWT Authentication Provider with HS256, HS384, and HS512 verification, declarative Claim policy, and per-request Secret resolution.
- Production JWT Factory that deeply converts Authentication Provider Snapshot metadata into Provider-local runtime configuration.
- Session lifecycle hardening that keeps WebSocket writes outside the lifecycle mutex while preserving serialized writes and deterministic Send-versus-Stop admission.
- Listener shutdown result sharing for concurrent and repeated Stop calls, context-bounded secondary waits, and independent HTTP Shutdown and TCP Close failures preserved through `errors.Is`.
- Runtime startup capability validation before Listener construction, with explicit TLS rejection and safe classified errors.
- Configured pre-Upgrade Handshake deadline propagated through Authentication and final admission validation.
- Authentication composition of enabled Providers in ascending configured Priority, while disabled Provider metadata remains inactive.

## [0.1.0-alpha] - 2026-07-12

### Added

- Initial project structure and architecture documentation.
- Go-based Control Service.
- `GET /health` endpoint.
- Startup configuration through environment variables.
- Structured logging with `log/slog`.
- Graceful shutdown on operating-system signals.
- In-memory CRUD API for Workspace.
- In-memory CRUD API for Configuration.
- ConfigurationVersion creation.
- Automatic ConfigurationVersion numbering.
- ConfigurationVersion publishing.
- Automatic archiving of the previous Published Version.
- Manual ConfigurationVersion archiving.
- Unit and HTTP tests.

### Known limitations

- Data is stored only in memory and is lost when the process stops.
- PostgreSQL is not connected yet.
- Runtime is not implemented yet.
- WebSocket listener is not implemented yet.
- Admin UI is not implemented yet.
- Validation, Rollback, and Snapshot are not implemented yet.
- The race detector was not run in the current Windows environment without CGO.

[Unreleased]: https://github.com/dsdred/universal-websocket-platform/compare/v0.1.0-alpha...HEAD
[0.1.0-alpha]: https://github.com/dsdred/universal-websocket-platform/releases/tag/v0.1.0-alpha

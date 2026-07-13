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
- Storage-neutral Secret Resolver contract and thread-safe in-memory implementation for tests and local development.
- Extensible runtime Authentication Provider and Factory contracts with transport-neutral request, result, and Principal models.
- Thread-safe Authentication Provider Registry with Factory registration and delegated Provider creation.

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

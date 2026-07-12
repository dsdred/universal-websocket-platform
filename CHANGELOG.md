# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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

[0.1.0-alpha]: https://github.com/dsdred/universal-websocket-platform/releases/tag/v0.1.0-alpha

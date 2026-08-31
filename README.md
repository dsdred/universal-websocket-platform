# Universal WebSocket Platform

[Russian version](README.ru.md)

Universal WebSocket Platform is an open-source platform for creating, configuring, deploying, and operating independent WebSocket servers without writing infrastructure code.

## Status

The project is in the **Beta — Complete the Single-Node Runtime** engineering
milestone and is not production-ready. The repository contains the Control
Service, in-memory domain APIs, and a production-composed single-node Runtime
vertical with pre-Upgrade Authentication, deterministic routing,
transactional Session handoff, and Manager-aware shutdown.

The Configuration Loader, Snapshot Builder, Runtime Bootstrap and Launcher,
Lifecycle Owner, management routing, operational identity, command
idempotency, and orchestration prerequisites also exist in isolation. They are
not wired into a Control Service Runtime-management path: Production
Activation, external persistence, recovery and reporting implementations,
complete TLS and Listener-settings execution, and the Control Service Runtime
API remain absent.

## Current release

**Version:** `v0.1.0-alpha`

**Release maturity:** alpha

This tagged release includes the Control Service and the basic lifecycle for
Workspace, Configuration, and ConfigurationVersion. The current repository
contains later, unreleased Beta engineering progress described above. See the
[release notes](docs/en/releases/v0.1.0-alpha.md) and
[`CHANGELOG.md`](CHANGELOG.md) for release details.

## Project principles

- Configuration over Code
- Runtime Isolation
- API First
- Technology Neutrality
- Provider-based architecture
- Explainability
- Predictability
- Keep MVP Simple

## Documentation

- [Documentation home](docs/en/README.md)
- [Architecture guides](docs/en/architecture/README.md)
- [Runtime design documents](docs/en/design/README.md)
- [Architecture Decision Records](docs/en/adr/README.md)
- [Engineering roadmap](docs/en/roadmap/README.md)
- [Architecture reviews](docs/en/reviews/README.md)
- [Current implementation state](spec/current-state.md)
- [Engineering Wiki](wiki/README.md)
- [Release Notes](docs/en/releases/)
- [Internal specifications](spec/README.md)
- [AI agent entry point](AGENTS.md)
- [Internal engineering process](docs/engineering/AGENT.md)
- [Task records and synchronization report](docs/tasks/README.md)

## Contributing

The project is in an active Beta engineering stage. Read the English
documentation and the factual implementation state before proposing changes.
Architecture choices should be recorded before they become implementation
constraints.

## License

See [`LICENSE`](LICENSE).

# Universal WebSocket Platform

[Russian version](README.ru.md)

Universal WebSocket Platform is an open-source platform for creating, configuring, deploying, and operating independent WebSocket servers without writing infrastructure code.

## Status

The project is in early alpha and is not production-ready. The repository contains the Control Service, in-memory domain APIs, and an implemented single-node Runtime vertical whose Manager-aware shutdown integration is still in progress.

## Current release

**Version:** `v0.1.0-alpha`

**Status:** early alpha

This release includes the Control Service and the basic lifecycle for Workspace, Configuration, and ConfigurationVersion. See the [release notes](docs/en/releases/v0.1.0-alpha.md) and [`CHANGELOG.md`](CHANGELOG.md) for details.

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

## Contributing

The project is at an early stage. Read the English documentation before proposing changes. Architecture choices should be recorded before they become implementation constraints.

## License

See [`LICENSE`](LICENSE).

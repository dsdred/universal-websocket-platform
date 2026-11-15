# Universal WebSocket Platform

[Russian version](README.ru.md)

Universal WebSocket Platform is an open-source platform for creating, configuring, deploying, and operating independent WebSocket servers without writing infrastructure code.

## Status

The project is in the **Beta — Complete the Single-Node Runtime** engineering
milestone and is not production-ready. The repository contains the Control
Service, in-memory domain APIs, and a production-composed single-node Runtime
vertical with pre-Upgrade Authentication, deterministic routing,
transactional Session handoff, and Manager-aware shutdown.

The current repository also contains a built-in **Local Configuration
Studio** at the Control Service root. It is a trusted local-development UI for
one existing-API-only flow: create a Workspace, Configuration and Draft, set
Listener host/port, publish, and inspect the returned state/version plus a
preliminary WebSocket URL. `Published` does not mean `Running`; data remains
in memory and is lost on restart; the URL is not a reachability check. The
Studio is not a production Admin UI and does not start or manage Runtime.

The embedded **UWP Control Dashboard** at `/dashboard` adds a Slate Blue shell
shared with Studio and GET-only browsing of existing Workspaces, Configurations
and Versions. It shows Control Service availability, version state, listener
host/port, available resource update times and browser fetch times. Runtime,
Connections and Diagnostics are explicitly unavailable because no HTTP GET API
is connected. Refresh is manual; there is no polling or WebSocket probing.
Unsafe JavaScript integer IDs produce an unsupported-response state before
rendering or nested requests. Raw JSON and secret references are not displayed.
The unauthenticated service is for trusted localhost/same-origin use only.
Dashboard does not make the Studio or existing API read-only.

Run locally with Go, without npm or a frontend build:

```powershell
$env:UWP_HTTP_HOST="127.0.0.1"
$env:UWP_HTTP_PORT="18081"
go run ./cmd/control-service
```

Open [Dashboard](http://127.0.0.1:18081/dashboard) or
[Configuration Studio](http://127.0.0.1:18081/). Without environment overrides,
the default address is `127.0.0.1:8080`. A fresh service has empty lists; use
Studio to create and publish a configuration, then return to Dashboard and
select its Workspace and Configuration. Data disappears on service restart.

The Configuration Loader, Snapshot Builder, Runtime Bootstrap and Launcher,
Lifecycle Owner, management routing, operational identity, command
idempotency, and orchestration prerequisites also exist in isolation. They are
not wired into a Control Service Runtime-management path: Production
Activation, external persistence, recovery and reporting implementations,
and the Control Service Runtime API remain absent. The Runtime vertical supports
one startup-loaded TLS certificate/key pair and configured Session timeouts;
certificate rotation/reload and mTLS are not implemented.

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

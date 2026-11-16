# Local Quick Start

[Russian version](../ru/quick-start.md) · [Documentation home](README.md) · [Project README](../../README.md)

Run UWP Control in a trusted local development environment. The Control Service
is unauthenticated; keep it on loopback. This guide uses the current repository,
which includes later work than the tagged `v0.1.0-alpha` release. It is not a
production deployment guide.

## Requirements

- A checkout of this repository and a Go toolchain compatible with its
  `go.mod` (`go 1.25.0`). Go 1.26.5 on Windows was used for this guide's smoke test.
- A free local TCP port. The examples use `18085`.
- No npm, frontend build, CDN or additional frontend dependencies.

## Start the Control Service

From the repository root, in PowerShell:

```powershell
$env:UWP_HTTP_HOST="127.0.0.1"
$env:UWP_HTTP_PORT="18085"
go run ./cmd/control-service
```

Or in a POSIX shell:

```sh
UWP_HTTP_HOST=127.0.0.1 UWP_HTTP_PORT=18085 go run ./cmd/control-service
```

Keep the process running while using the UI; stop it with `Ctrl+C`. Without
environment overrides, the service defaults to `127.0.0.1:8080`. If the port is
already in use, choose another `UWP_HTTP_PORT` and use it in the URLs below.
The startup log reports the actual address. These variables configure the
Control Service HTTP address, not a Runtime WebSocket listener.

These addresses point to your own computer. Dashboard and Configuration Studio
are available only after you start UWP locally as described above.

| Route | Purpose |
|---|---|
| [Dashboard](http://127.0.0.1:18085/dashboard) (`/dashboard`) | Read-only GET browsing; manual Refresh |
| [Configuration Studio](http://127.0.0.1:18085/) (`/`) | Existing create/publish flow |
| [Health](http://127.0.0.1:18085/health) (`/health`) | Control Service availability only |

A fresh service has no Workspaces. Empty lists are expected, not a failure.

## Create and publish a configuration

1. Open Configuration Studio at `/`.
2. Enter a Workspace name and Configuration name, for example `Local example`
   and `Echo listener`.
3. Keep Listener host `127.0.0.1` and port `8080`, or enter your intended
   configuration values. These are saved settings; they do not bind a listener.
4. Select **Create and publish**. Studio creates a Workspace, Configuration and
   Draft, sets the listener and publishes that version through the existing API.
5. Check the returned **Published** state, version and preliminary WebSocket URL.
   `Published` does not mean `Running`; the URL is not a reachability check.
6. Open Dashboard, choose the Workspace and Configuration, and inspect Versions,
   state, listener host/port and available timestamps. **Refresh** reloads the
   lists and Control Service health; there is no automatic polling.

Dashboard is read-only. Studio uses the existing mutation API; the shared shell
does not make the entire service read-only. The example names are illustrative
local data, not seeded product records or fabricated metrics.

## Screenshots

Actual running local UI, captured after the example workflow above. The Viking
branding comes from the accepted asset set; these are not mockups.

![Dashboard with a published local example](../images/dashboard.jpg)

![Configuration Studio after publish](../images/configuration-studio.jpg)

## Product limitations

- **Published != Running.** Neither Studio nor Dashboard starts or manages Runtime.
- **`/health` covers Control Service only.** It does not confirm Runtime status,
  WebSocket reachability, connection counts or uptime.
- **Control Service data is in memory.** Workspaces, Configurations and Versions
  are lost when the service restarts; this flow does not provide persistence.
- **Runtime, Connections and Diagnostics have no connected HTTP GET API.**
  Dashboard displays them as unavailable, not zero or healthy.
- **Trusted local environment only.** Control Service has no authentication;
  keep the loopback/same-origin boundary and do not expose it publicly.

Resource update timestamps come from the API; browser checked/fetched times are
separate. Dashboard does not display raw JSON or secret references. See the
[current implementation state](../../spec/current-state.md) for implemented,
isolated and absent capabilities, and the [release notes](releases/v0.1.0-alpha.md)
for the earlier tagged release. Stage B and production activation remain inactive.

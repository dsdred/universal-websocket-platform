# ADR 0001: Bootstrap Control Service

## Status

Accepted.

## Context

After the documentation bootstrap, the project needs its first executable component: a minimal HTTP Control Service with a health-check endpoint.

## Decision

- Use Go 1.25.
- Use Chi Router for HTTP request routing.
- Place the entry point in `cmd/control-service`.
- Place the internal HTTP Server, Configuration, and logging packages in `internal/`.
- Use `slog` for logging.
- Construct dependencies explicitly, without third-party DI libraries or global variables.

## Consequences

- `go run ./cmd/control-service` starts the HTTP Server on port `8080` by default.
- The port can be set through the `PORT` environment variable.
- `GET /health` returns HTTP 200 and `{"status":"ok"}`.
- Go and Chi become accepted technologies for the current Control Service.

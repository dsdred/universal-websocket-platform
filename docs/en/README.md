# Universal WebSocket Platform Documentation

[Russian version](../ru/README.md)

## Start Here

Choose the route that matches what you need:

- **What works in the current repository?** Read the root
  [project status](../../README.md#status), then use the factual
  [current implementation state](../../spec/current-state.md) for the detailed
  implemented, isolated, planned, and absent boundaries.
- **What is included in the published release?** Read the
  [`v0.1.0-alpha` release notes](releases/v0.1.0-alpha.md). The tagged alpha
  release is an earlier snapshot than the current repository.
- **Want to contribute or follow the engineering direction?** Start with the
  root [contribution guidance](../../README.md#contributing), then use the
  [Engineering Roadmap](roadmap/README.md) and [Engineering Wiki](../../wiki/README.md)
  for plans, principles, process, and knowledge ownership.

The current **Beta — Complete the Single-Node Runtime** milestone describes
engineering progress; it is not a production-readiness claim. A production
installation/operator quick start is not available: Control Service Production
Activation and the complete operator workflow remain absent. Use the current
state for repository facts and the release notes for tagged-release facts.

## Architecture Guides

Active architecture-wide patterns, including the complete ARCH document list, are maintained in the [Architecture Guides index](architecture/README.md).

## Engineering Roadmap

Living engineering maturity plans are maintained in the [Engineering Roadmap index](roadmap/README.md).

## Runtime Design Documents

Focused Runtime subsystem designs, their document statuses, and the complete DP list are maintained in the [Runtime Design Documents index](design/README.md).

## Architecture Decision Records

Accepted architecture decisions and ADR conventions are maintained in the [Architecture Decision Records index](adr/README.md).

## Authentication Design Proposals

Authentication proposals that predate the separate Runtime design-document series are available in [`proposals/`](proposals/).

- [DP-001: Authentication](proposals/DP-001-authentication.md)
- [DP-002: Secret References](proposals/DP-002-secret-references.md)
- [DP-003: JWT Provider](proposals/DP-003-jwt-provider.md)
- [DP-004: Authentication Runtime Contracts](proposals/DP-004-authentication-runtime-contracts.md)

## Release Notes

Release-specific summaries are available in [`releases/`](releases/).

## Architecture Reviews

Evidence-based implementation reviews are maintained in the [Architecture Reviews index](reviews/README.md).

## Engineering Process

- [LLM Development Guide](process/LLM_DEVELOPMENT_GUIDE.md)

## Retrospectives

- [DP-005 Runtime Message Router Retrospective](retrospectives/DP-005.md)

## Internal Specifications

The team's working specifications are maintained in [`spec/`](../../spec/). The factual implementation record is [`spec/current-state.md`](../../spec/current-state.md). Internal specifications are currently written predominantly in Russian and may describe work in progress rather than stable public contracts.

## Internal Agent Workflow

AI agents start at the repository root [`AGENTS.md`](../../AGENTS.md), then read
the internal [agent contract](../engineering/AGENT.md). Operational task records
and synchronization reports are indexed in [`docs/tasks/`](../tasks/README.md).
These internal Russian-language documents do not require public EN/RU mirrors.

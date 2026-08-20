# Runtime Design Documents

[Russian version](../../ru/design/README.md)

This directory contains focused Runtime design documents. A design document explores a subsystem and proposed architecture before implementation. It does not replace accepted ADRs, active architecture guides, or evidence-based reviews.

| Document | Status |
| --- | --- |
| [DP-001: Runtime Handshake Pipeline](DP-001-runtime-handshake-pipeline.md) | Draft; partially implemented |
| [DP-002: Runtime Host Composition Root](DP-002-runtime-host-composition-root.md) | Draft; partially implemented |
| [DP-003: Runtime Session Manager](DP-003-runtime-session-manager.md) | Approved |
| [DP-004: Per-Session Execution Boundary](DP-004-per-session-execution-boundary.md) | Approved |
| [DP-005: Runtime Message Router](DP-005-runtime-message-router.md) | Approved |
| [DP-006: Runtime Production Integration](DP-006-runtime-production-integration.md) | Draft; partially implemented |
| [DP-007: Configuration Loader Contract](DP-007-configuration-loader-contract.md) | Draft; implemented in isolation |
| [DP-008: Snapshot Builder Contract](DP-008-snapshot-builder-contract.md) | Draft; implemented in isolation |
| [DP-009: Runtime Bootstrap Contract](DP-009-runtime-bootstrap-contract.md) | Draft; implemented in isolation |
| [DP-010: Runtime Lifecycle Owner Contract](DP-010-runtime-lifecycle-owner-contract.md) | Draft; base and expected-attempt Stop extension implemented in isolation |
| [DP-011: Runtime Launch Pipeline Integration](DP-011-runtime-launch-pipeline-integration.md) | Draft; base and managed continuation implemented in isolation; composition invoker planned |
| [DP-012: Runtime Source Composition](DP-012-runtime-source-composition.md) | Draft; implemented in isolation |
| [DP-013: Runtime Management Routing](DP-013-runtime-management-routing.md) | Draft; implemented in isolation; integration blocked |
| [DP-014: Runtime Operational Identity Persistence](DP-014-runtime-operational-identity-persistence.md) | Approved; implemented in isolation |
| [DP-015: Runtime Management Command Idempotency](DP-015-runtime-management-command-idempotency.md) | Approved; primitive/parent managed gates, parent/phase core, Continue/pending-Stop, and Slice 3 continuation implemented in isolation; complete extension planned |
| [DP-016: Runtime Activation, Replacement, and Rollback](DP-016-runtime-activation-replacement-rollback.md) | Approved; planned and architecture-blocked by remaining DP-019 prerequisites |
| [DP-017: Runtime Recovery and Reconciliation](DP-017-runtime-recovery-reconciliation.md) | Approved; planned |
| [DP-018: Runtime Operational Error Reporting and Redaction](DP-018-runtime-operational-error-reporting-redaction.md) | Approved; planned |
| [DP-019: Runtime Activation Orchestration Prerequisites](DP-019-runtime-activation-orchestration-prerequisites.md) | Approved; planned overall; managed command/Flow/continuation and attempt/generation binding seams implemented in isolation |
| [DP-020: Runtime Orchestration Binding Sequence Readiness](DP-020-runtime-orchestration-binding-sequence-readiness.md) | Draft; planned overall; Slice 3 implemented and independently accepted in isolation |
| [DP-021: Private Exact-Scope Managed Start Invoker](DP-021-private-exact-scope-managed-start-invoker.md) | Draft; planned |

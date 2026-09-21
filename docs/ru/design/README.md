# Документы проектирования Runtime

[English version](../../en/design/README.md)

Каталог содержит сфокусированные документы проектирования Runtime. Design document исследует подсистему и предлагаемую архитектуру до реализации. Он не заменяет принятые ADR, активные архитектурные руководства или основанные на фактах reviews.

| Документ | Статус |
| --- | --- |
| [DP-001: Runtime Handshake Pipeline](DP-001-runtime-handshake-pipeline.md) | Draft; реализован частично |
| [DP-002: Runtime Host Composition Root](DP-002-runtime-host-composition-root.md) | Draft; реализован частично |
| [DP-003: Runtime Session Manager](DP-003-runtime-session-manager.md) | Approved |
| [DP-004: Per-Session Execution Boundary](DP-004-per-session-execution-boundary.md) | Approved |
| [DP-005: Маршрутизатор Runtime-сообщений](DP-005-runtime-message-router.md) | Approved |
| [DP-006: Production-интеграция Runtime](DP-006-runtime-production-integration.md) | Draft; реализован частично |
| [DP-007: Configuration Loader Contract](DP-007-configuration-loader-contract.md) | Draft; реализован изолированно |
| [DP-008: Snapshot Builder Contract](DP-008-snapshot-builder-contract.md) | Draft; реализован изолированно |
| [DP-009: Runtime Bootstrap Contract](DP-009-runtime-bootstrap-contract.md) | Draft; реализован изолированно |
| [DP-010: Контракт Runtime Lifecycle Owner](DP-010-runtime-lifecycle-owner-contract.md) | Draft; base и extension expected-attempt Stop реализованы изолированно |
| [DP-011: Интеграция Runtime Launch Pipeline](DP-011-runtime-launch-pipeline-integration.md) | Draft; base, managed continuation и exact invoker реализованы изолированно; callback/production composition запланированы |
| [DP-012: Композиция Runtime Source](DP-012-runtime-source-composition.md) | Draft; реализован изолированно |
| [DP-013: Маршрутизация управления Runtime](DP-013-runtime-management-routing.md) | Draft; реализован изолированно; integration blocked |
| [DP-014: Персистентность operational identity Runtime](DP-014-runtime-operational-identity-persistence.md) | Approved; реализован изолированно |
| [DP-015: Идемпотентность management commands Runtime](DP-015-runtime-management-command-idempotency.md) | Approved; Partial implementation: TASK-062 parent-terminalization repair принята/опубликована; TASK-064 durable primitive Satisfied-outcome repair независимо принята и опубликована через PR #68 |
| [DP-016: Activation, replacement и rollback Runtime](DP-016-runtime-activation-replacement-rollback.md) | Approved; реализован изолированно Coordinator-Accepted TASK-026; Tester, PROCESS-002, Scope Audit 31/0/0 и final Reviewer проходят |
| [DP-017: Восстановление и сверка Runtime](DP-017-runtime-recovery-reconciliation.md) | Approved; запланирован |
| [DP-018: Operational error reporting и redaction Runtime](DP-018-runtime-operational-error-reporting-redaction.md) | Approved; запланирован |
| [DP-019: Предпосылки оркестрации активации Runtime](DP-019-runtime-activation-orchestration-prerequisites.md) | Approved; planned overall; accepted/published prerequisites TASK-062 и TASK-064 поддерживают Coordinator-Accepted isolated orchestrator TASK-026 |
| [DP-020: Готовность последовательности связывания оркестрации Runtime](DP-020-runtime-orchestration-binding-sequence-readiness.md) | Draft; planned overall; historical readiness TASK-061 сохраняется; accepted/published TASK-062/TASK-064 и readiness TASK-063 поддерживают Coordinator-Accepted isolated orchestrator TASK-026 |
| [DP-021: Private Exact-Scope Managed Start Invoker](DP-021-private-exact-scope-managed-start-invoker.md) | Draft; partial, реализован изолированно TASK-043; invoker boundary неизменна; Coordinator-Accepted TASK-026 использует её только внутри isolated orchestrator |
| [DP-022: Граница containment исполнения Runtime и evidence](DP-022-runtime-execution-containment-and-evidence.md) | Approved; запланирован; определяет containment/evidence prerequisite, требуемый DP-017 §11; containment capability, ledger и evidence adapter отсутствуют, поэтому реализация DP-017 остаётся неактивированной |
| [DP-023: Bootstrap process-containment Runtime](DP-023-runtime-process-containment-bootstrap.md) | Approved; запланирован; определяет safety-atomic bootstrap capability/ledger/generation и первый implementation slice; implementation и downstream activation отсутствуют |

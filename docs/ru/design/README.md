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
| [DP-010: Контракт Runtime Lifecycle Owner](DP-010-runtime-lifecycle-owner-contract.md) | Draft; base реализован изолированно, extension expected-attempt Stop запланирован |
| [DP-011: Интеграция Runtime Launch Pipeline](DP-011-runtime-launch-pipeline-integration.md) | Draft; base реализован изолированно, continuation/binding gate DP-019 запланирован |
| [DP-012: Композиция Runtime Source](DP-012-runtime-source-composition.md) | Draft; реализован изолированно |
| [DP-013: Маршрутизация управления Runtime](DP-013-runtime-management-routing.md) | Draft; реализован изолированно; integration blocked |
| [DP-014: Персистентность operational identity Runtime](DP-014-runtime-operational-identity-persistence.md) | Approved; реализован изолированно |
| [DP-015: Идемпотентность management commands Runtime](DP-015-runtime-management-command-idempotency.md) | Approved; primitive boundary, parent/phase sequential core DP-019 и command-boundary Continue/pending-Stop реализованы изолированно |
| [DP-016: Activation, replacement и rollback Runtime](DP-016-runtime-activation-replacement-rollback.md) | Approved; запланирован и architecture-blocked оставшимися prerequisites DP-019 |
| [DP-017: Восстановление и сверка Runtime](DP-017-runtime-recovery-reconciliation.md) | Approved; запланирован |
| [DP-018: Operational error reporting и redaction Runtime](DP-018-runtime-operational-error-reporting-redaction.md) | Approved; запланирован |
| [DP-019: Предпосылки оркестрации активации Runtime](DP-019-runtime-activation-orchestration-prerequisites.md) | Approved; в целом запланирован, parent/phase core и command-boundary Continue/pending-Stop реализованы изолированно |
| [DP-020: Готовность последовательности связывания оркестрации Runtime](DP-020-runtime-orchestration-binding-sequence-readiness.md) | Draft; статус реализации Planned |

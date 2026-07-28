# Контекст проекта

## Основная информация

**Проект:** Universal WebSocket Platform

**Миссия:** Open-source платформа для создания, настройки, развертывания и эксплуатации независимых WebSocket-серверов без написания инфраструктурного кода.

## Текущее состояние

- Текущая веха: **Beta — Complete the Single-Node Runtime**
- Статус реализации: **Control Service, single-node Runtime vertical, Configuration Loader boundary, DP-008 Snapshot Builder, DP-009 Runtime Bootstrap и stateless Runtime Launcher реализованы изолированно; Runtime Lifecycle Owner и production launch pipeline не реализованы**
- Последняя завершённая development task: **TASK-007 — Draft DP-010 Runtime Lifecycle Owner prerequisites; Completed — Coordinator Accepted**
- Последняя завершённая operational task: **TASK-007 — Draft DP-010 Runtime Lifecycle Owner prerequisites; Completed — Coordinator Accepted**
- Текущая operational task: **не назначена; принятый восьмифайловый diff TASK-007 остаётся в attributed dirty worktree ветки `docs/task-007-runtime-lifecycle-owner-prerequisites` до отдельно разрешённого commit**
- Trusted baseline TASK-007: **TASK-006 merged через PR #7 в clean `main` commit `e791482`; TASK-007 начата от этого baseline в `docs/task-007-runtime-lifecycle-owner-prerequisites`, task record создан первым content change**
- Verification TASK-001: **targeted tests PASS 3/3; full `go test ./... -count=1` PASS 2/2; `go vet ./...`, `gofmt -d` и `git diff --check` PASS; race detector недоступен без CGO/gcc**
- Operational entry после принятия TASK-002: **точная команда `Продолжай проект.` запускает repository-native selection и полный PROCESS-001 cycle без неявных commit, push или merge**
- Verification TASK-002: **Tester PASS; Reviewer Approved after rework; scope audit accepted — 8 Required, 0 Questionable, 0 Removable**
- Результат TASK-003: **implementation prerequisites Draft DP-009 завершены; Tester PASS after rework; Reviewer final closure Approved; Coordinator scope audit accepted — 6 Required, 0 Questionable, 0 Removable**
- Результат TASK-004: **isolated Runtime Bootstrap реализован и принят; Tester PASS; Reviewer Approved with Findings, blocking 0; scope audit accepted — 12 Required, 0 Questionable, 0 Removable**
- Verification TASK-004: **targeted и full tests, vet, gofmt, documentation links/parity и diff checks PASS; race detector недоступен без CGO/gcc**
- Результат TASK-005: **planned in-process `func Launch(request *BootstrapRequest) BootstrapOutcome` contract зеркально уточнён в Draft DP-009: borrowed same pointer, ровно одна delegation в реализованный Bootstrap, unchanged outcome/Host/failure identities и cause chain, ownership handoff будущему Lifecycle Owner, zero policy/cleanup/state и AP-003/AP-011 local-vs-integration proof split; Completed — Coordinator Accepted; production code отсутствует**
- Verification TASK-005: **Tester PASS; PROCESS-002 Synchronized; Final Reviewer Approved, 0 blocking и 0 nonblocking findings; Coordinator Scope Audit accepted — 6 Required, 0 Questionable, 0 Removable; Coordinator Acceptance получена**
- Результат TASK-006: **`internal/runtime.Launch` реализован как exact stateless `return Bootstrap(request)` без adapter, state, validation, wrapping, cleanup, retry или lifecycle policy; Lifecycle Owner и production wiring отсутствуют**
- Verification TASK-006: **targeted и full Go tests, `go vet`, `gofmt -d`, EN/RU structure/status parity и diff checks PASS; race detector недоступен при `CGO_ENABLED=0` и отсутствующем `gcc`; Final Reviewer Approved, 0 blocking и 0 nonblocking findings; Scope Audit accepted — 8 Required, 0 Questionable, 0 Removable**
- Результат TASK-007: **зеркальный Draft DP-010 с Implementation Status Planned фиксирует минимальный `internal/runtimelifecycle` contract: Owner-bound Workspace/Configuration/Runtime Instance, Owner-issued Launch Attempt и exact ConfigurationVersion pin в `PrepareStart` до Loader/Builder, closed `PreparationResult`, first-valid-result-wins, origin-sensitive Stop, truthful immutable outcomes/observation и local-vs-integration proofs; production code отсутствует**
- Verification TASK-007: **после rework B-001/B-002, R-001/R-002 и project-state correction F-001 Final Reviewer выдал Approved, 0 blocking и 0 nonblocking findings; Final Tester PASS; PROCESS-002 Synchronized; Coordinator Scope Audit accepted — 8 Required, 0 Questionable, 0 Removable; Coordinator Acceptance получена**
- Следующий разрешённый шаг: **только отдельно разрешённый commit принятого diff TASK-007; stage, commit, push и merge не выполняются неявно**
- Следующая рекомендуемая work после closure TASK-007: **отдельная изолированная production implementation минимального in-process Runtime Lifecycle Owner по reviewed Draft DP-010; следующая task/branch не созданы и implementation не начата**
- Readiness boundary следующей work: **только локальный Owner package и proof tests DP-010 могут стать bounded implementation candidate; Loader/Builder production wiring, persistence, management API, retry/reconciliation и supervision остаются отдельной неготовой work**
- Stage 2 task-before-work ordering выполнен для TASK-003, TASK-004, TASK-005, TASK-006 и TASK-007: task record создан первым content change, а task index обновлён только после initial gate
- Design Status DP-009 остаётся **Draft**; Bootstrap и Runtime Launcher Implementation Status — **Implemented in isolation**; Runtime Lifecycle Owner и production launch pipeline не реализованы; AP-003 и AP-011 остаются integration-gated
- Design Status DP-010 остаётся **Draft**, Implementation Status — **Planned**; reviewed contract не является реализацией Owner или production pipeline
- Design Status DP-008 остаётся **Draft**, Implementation Status — **Implemented in isolation**
- Содержимое репозитория: документация, спецификации, инженерные соглашения, исполняемый Control Service и изолированные Runtime-компоненты с тестами

## Архитектурные принципы

1. Configuration over Code
2. Runtime Isolation
3. API First
4. Technology Neutrality
5. Provider-based architecture
6. Explainability
7. Predictability
8. Keep MVP Simple

## Источники истины

- `spec/00-product/vision.md` определяет замысел продукта.
- `spec/01-principles/architecture-principles.md` определяет архитектурные ориентиры.
- `spec/current-state.md` фиксирует текущее состояние проекта.
- `spec/decisions.md` содержит перечень принятых и ожидающих принятия решений.
- `docs/en/adr/` и `docs/ru/adr/` содержат публичные записи об архитектурных решениях.

Не делайте вывод о реализованных возможностях на основании миссии или спецификаций, описывающих будущее состояние. Перед изменением репозитория начните с корневого `AGENTS.md`, затем следуйте `docs/engineering/AGENT.md` и сверяйтесь с `spec/current-state.md`.

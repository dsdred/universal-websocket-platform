# Контекст проекта

## Основная информация

**Проект:** Universal WebSocket Platform

**Миссия:** Open-source платформа для создания, настройки, развертывания и эксплуатации независимых WebSocket-серверов без написания инфраструктурного кода.

## Текущее состояние

- Текущая веха: **Beta — Complete the Single-Node Runtime**
- Статус реализации: **Control Service, single-node Runtime vertical,
  Configuration Loader boundary, DP-008 Snapshot Builder, DP-009 Runtime
  Bootstrap, stateless Runtime Launcher и DP-010 Runtime Lifecycle Owner
  реализованы изолированно; production launch pipeline не реализован**
- Последняя завершённая development task: **TASK-009 — Runtime Lifecycle Owner implementation; Completed — Coordinator Accepted**
- Последняя завершённая operational task: **TASK-008 — Publisher pipeline governance; Completed — Coordinator Accepted**
- Текущая operational task: **отсутствует**
- Trusted baseline TASK-009: **clean synchronized
  `main@63b961eeb59af9205c3c3d0b68d3f4bd7b8ac25c`; локальная ветка
  `feature/task-009-runtime-lifecycle-owner`; task record создан первым
  content change**
- Trusted baseline TASK-008: **TASK-007 task commit `2e6d221` опубликован через PR #8 и merged в clean `main` commit `802760a`; TASK-008 начата от этого baseline, task record создан первым content change**
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
- Operational governance TASK-008: **точная команда `Разрешаю публиковать.` документирована как единое разрешение P0–P10 для одного immutable task target; Initial P0 отделён от phase-aware Resume Reconstruction Guard, push/merge являются checkpoints, external blocker сохраняет authority, post-P6 resume остаётся на `main`, No CI допускается только при `MERGEABLE / CLEAN`, cleanup и terminal payload обязательны; R-001/R-002 устранены, Final Reviewer Approved 0/0, Tester PASS, PROCESS-002 Synchronized, Coordinator Scope Audit accepted — 14 Required, 0 Questionable, 0 Removable; Coordinator Acceptance получена**
- Product impact TASK-008: **отсутствует; production code/tests, `.github`, ADR/ARCH/DP, product capability и Runtime implementation не изменены**
- Результат TASK-009: **добавлен изолированный
  `internal/runtimelifecycle`: Owner-bound identities, Owner-issued Launch
  Attempt и exact version pin через `PrepareStart`, closed preparation result,
  single tracked Launcher/Host Stop operations, same-token convergence,
  origin-sensitive truthful outcomes, cancellation-only caller waits и
  coherent observation; Bootstrap, Launcher, Host и production wiring не
  изменены**
- Verification TASK-009: **targeted и full `go test ./... -count=1`,
  stress `-count=100`, `go vet ./...`, `go fmt ./...`,
  `git diff --check`, EN/RU parity и link validation PASS; race detector
  недоступен при `CGO_ENABLED=0` и отсутствующем `gcc`; independent Tester
  PASS; Scope Audit accepted — 14 Required, 0 Questionable, 0 Removable;
  Final Reviewer Approved, 0 blocking и 0 nonblocking findings; Coordinator
  Acceptance получена**
- Closure publication state TASK-008: **на момент closure stage, commit и publication не выполнялись; это historical fact, а не live gate. Любая последующая разрешённая публикация reconstruct-ит фактическое состояние из Git/GitHub и не хранит transient dirty/push/PR/cleanup state в project context**
- Следующая work после TASK-009: **не активирована; production wiring
  Loader-to-Builder-to-Launcher, persistence, management API,
  retry/reconciliation и supervision остаются за границей TASK-009 и требуют
  отдельного readiness/contract решения**
- Stage 2 task-before-work ordering выполнен для TASK-003, TASK-004, TASK-005, TASK-006 и TASK-007: task record создан первым content change, а task index обновлён только после initial gate
- Publication history: **TASK-005 commit `99e0d3d`, TASK-006 commit `fd0f80a` и TASK-007 commit `2e6d221` merged через PR #6/#7/#8; transient pre-commit/Publisher blockers не являются durable project-state instructions**
- Design Status DP-009 остаётся **Draft**; Bootstrap и Runtime Launcher
  Implementation Status — **Implemented in isolation**; production launch
  pipeline не реализован, AP-003 и AP-011 остаются integration-gated
- Design Status DP-010 остаётся **Draft**, Implementation Status —
  **Implemented in isolation**; status не утверждает production wiring,
  persistence или management capability
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

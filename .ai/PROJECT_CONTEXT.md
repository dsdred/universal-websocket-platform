# Контекст проекта

## Основная информация

**Проект:** Universal WebSocket Platform

**Миссия:** Open-source платформа для создания, настройки, развертывания и эксплуатации независимых WebSocket-серверов без написания инфраструктурного кода.

## Текущее состояние

- Текущая веха: **Beta — Complete the Single-Node Runtime**
- Статус реализации: **Control Service, single-node Runtime vertical, Configuration Loader boundary, DP-008 Snapshot Builder и DP-009 Runtime Bootstrap реализованы изолированно; Runtime Launcher, Runtime Lifecycle Owner и production launch pipeline не реализованы**
- Последняя завершённая development task: **TASK-005 — documentation-only refinement implementation prerequisites in-process Runtime Launcher; Completed — Coordinator Accepted**
- Последняя завершённая operational task: **TASK-005 — documentation-only refinement implementation prerequisites in-process Runtime Launcher; Completed — Coordinator Accepted**
- Текущая operational task: **не назначена; принятые изменения TASK-005 остаются в attributed dirty worktree ветки `docs/task-005-runtime-launcher-prerequisites` до отдельно разрешённого commit**
- Trusted baseline TASK-005: **TASK-004 merged в `main` commit `7d614c4`; TASK-005 начата от clean baseline этого commit**
- Verification TASK-001: **targeted tests PASS 3/3; full `go test ./... -count=1` PASS 2/2; `go vet ./...`, `gofmt -d` и `git diff --check` PASS; race detector недоступен без CGO/gcc**
- Operational entry после принятия TASK-002: **точная команда `Продолжай проект.` запускает repository-native selection и полный PROCESS-001 cycle без неявных commit, push или merge**
- Verification TASK-002: **Tester PASS; Reviewer Approved after rework; scope audit accepted — 8 Required, 0 Questionable, 0 Removable**
- Результат TASK-003: **implementation prerequisites Draft DP-009 завершены; Tester PASS after rework; Reviewer final closure Approved; Coordinator scope audit accepted — 6 Required, 0 Questionable, 0 Removable**
- Результат TASK-004: **isolated Runtime Bootstrap реализован и принят; Tester PASS; Reviewer Approved with Findings, blocking 0; scope audit accepted — 12 Required, 0 Questionable, 0 Removable**
- Verification TASK-004: **targeted и full tests, vet, gofmt, documentation links/parity и diff checks PASS; race detector недоступен без CGO/gcc**
- Результат TASK-005: **planned in-process `func Launch(request *BootstrapRequest) BootstrapOutcome` contract зеркально уточнён в Draft DP-009: borrowed same pointer, ровно одна delegation в реализованный Bootstrap, unchanged outcome/Host/failure identities и cause chain, ownership handoff будущему Lifecycle Owner, zero policy/cleanup/state и AP-003/AP-011 local-vs-integration proof split; Completed — Coordinator Accepted; production code отсутствует**
- Verification TASK-005: **Tester PASS; PROCESS-002 Synchronized; Final Reviewer Approved, 0 blocking и 0 nonblocking findings; Coordinator Scope Audit accepted — 6 Required, 0 Questionable, 0 Removable; Coordinator Acceptance получена**
- Следующий разрешённый шаг: **только отдельно разрешённый commit TASK-005; commit, push и merge не выполнялись и не выполняются без разрешения**
- Следующая рекомендуемая Ready work после closure TASK-005: **isolated implementation in-process stateless Runtime Launcher строго по уточнённому contract; task и branch не созданы и работа не активирована**
- Readiness boundary следующей work: **implementation может начаться только после closure TASK-005 и clean trusted baseline; Runtime Lifecycle Owner и production Loader-to-Builder-to-Launcher pipeline остаются отдельной неготовой работой**
- Stage 2 task-before-work ordering выполнен для TASK-003, TASK-004 и TASK-005: task record создан первым content change, а task index обновлён только после initial gate
- Design Status DP-009 остаётся **Draft**; Bootstrap Implementation Status — **Implemented in isolation**; Runtime Launcher Implementation Status — **Planned**; Runtime Lifecycle Owner и production launch pipeline не реализованы; AP-003 и AP-011 остаются integration-gated
- Design Status DP-008 остаётся **Draft**
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

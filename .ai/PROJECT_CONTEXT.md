# Контекст проекта

## Основная информация

**Проект:** Universal WebSocket Platform

**Миссия:** Open-source платформа для создания, настройки, развертывания и эксплуатации независимых WebSocket-серверов без написания инфраструктурного кода.

## Текущее состояние

- Текущая веха: **Beta — Complete the Single-Node Runtime**
- Статус реализации: **Control Service, single-node Runtime vertical, Configuration Loader boundary, DP-008 Snapshot Builder и DP-009 Runtime Bootstrap реализованы изолированно; Runtime Launcher, Runtime Lifecycle Owner и production launch pipeline не реализованы**
- Последняя завершённая development task: **TASK-004 — изолированная реализация Runtime Bootstrap DP-009; Completed — Coordinator Accepted**
- Последняя завершённая operational task: **TASK-004 — изолированная реализация Runtime Bootstrap DP-009; Completed — Coordinator Accepted**
- Текущая operational task: **не назначена; TASK-004 принята Coordinator, её изменения остаются в attributed dirty worktree ветки `feature/task-004-dp009-runtime-bootstrap` до отдельно разрешённого commit**
- Verification TASK-001: **targeted tests PASS 3/3; full `go test ./... -count=1` PASS 2/2; `go vet ./...`, `gofmt -d` и `git diff --check` PASS; race detector недоступен без CGO/gcc**
- Operational entry после принятия TASK-002: **точная команда `Продолжай проект.` запускает repository-native selection и полный PROCESS-001 cycle без неявных commit, push или merge**
- Verification TASK-002: **Tester PASS; Reviewer Approved after rework; scope audit accepted — 8 Required, 0 Questionable, 0 Removable**
- Результат TASK-003: **implementation prerequisites Draft DP-009 завершены; Tester PASS after rework; Reviewer final closure Approved; Coordinator scope audit accepted — 6 Required, 0 Questionable, 0 Removable**
- Результат TASK-004: **isolated Runtime Bootstrap реализован и принят; Tester PASS; Reviewer Approved with Findings, blocking 0; scope audit accepted — 12 Required, 0 Questionable, 0 Removable**
- Verification TASK-004: **targeted и full tests, vet, gofmt, documentation links/parity и diff checks PASS; race detector недоступен без CGO/gcc**
- Следующий разрешённый шаг: **только отдельно разрешённый commit TASK-004; commit, push и merge не выполнялись и не выполняются без разрешения**
- Следующая рекомендуемая Ready work: **focused architecture/documentation refinement implementation prerequisites in-process Runtime Launcher boundary: concrete input/output, точная delegation в реализованный Bootstrap, ownership handoff будущему Lifecycle Owner, failure passthrough и AP-003/AP-011 proof boundary; task и branch не созданы**
- Readiness boundary следующей work: **ARCH-004 не определяет implementation APIs, а DP-009 §22 откладывает Launcher implementation; Launcher code, Runtime Lifecycle Owner и production pipeline не Ready и не начинаются**
- Stage 2 task-before-work ordering выполнен для TASK-003 и TASK-004: task record создан первым content change, а task index обновлён вторым
- Design Status DP-009 остаётся **Draft**, Implementation Status — **Implemented in isolation**; Runtime Launcher и production launch pipeline не реализованы
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

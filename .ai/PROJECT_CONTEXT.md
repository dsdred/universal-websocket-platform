# Контекст проекта

## Основная информация

**Проект:** Universal WebSocket Platform

**Миссия:** Open-source платформа для создания, настройки, развертывания и эксплуатации независимых WebSocket-серверов без написания инфраструктурного кода.

## Текущее состояние

- Текущая веха: **Beta — Complete the Single-Node Runtime**
- Статус реализации: **Control Service, single-node Runtime vertical, Configuration Loader boundary и DP-008 Snapshot Builder реализованы изолированно; production launch pipeline не завершён**
- Последняя завершённая development task: **TASK-001 — Draft DP-008 Snapshot Builder поверх neutral `DetachedLoadResult`**
- Последняя завершённая operational task: **TASK-003 — уточнение implementation prerequisites Draft DP-009; Completed — Coordinator Accepted**
- Текущая operational task: **не назначена; закрытые изменения TASK-003 остаются в attributed dirty worktree ветки `docs/task-003-dp009-prerequisites-refinement`**
- Verification TASK-001: **targeted tests PASS 3/3; full `go test ./... -count=1` PASS 2/2; `go vet ./...`, `gofmt -d` и `git diff --check` PASS; race detector недоступен без CGO/gcc**
- Operational entry после принятия TASK-002: **точная команда `Продолжай проект.` запускает repository-native selection и полный PROCESS-001 cycle без неявных commit, push или merge**
- Verification TASK-002: **Tester PASS; Reviewer Approved after rework; scope audit accepted — 8 Required, 0 Questionable, 0 Removable**
- Результат TASK-003: **implementation prerequisites Draft DP-009 завершены; Tester PASS after rework; Reviewer final closure Approved; Coordinator scope audit accepted — 6 Required, 0 Questionable, 0 Removable**
- Следующий разрешённый шаг: **отдельно разрешённый commit TASK-003; commit, push и merge не выполняются без разрешения**
- Следующая рекомендуемая Ready work после clean trusted baseline: **изолированная реализация Runtime Bootstrap DP-009; task и branch не созданы, AP-003/AP-011 и production integration явно отложены**
- Stage 2 task-before-work ordering выполнен: TASK-003 создан первым content change, а task index обновлён вторым
- Design Status DP-009 остаётся **Draft**, Implementation Status — **Planned**; Bootstrap и production launch pipeline не реализованы
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

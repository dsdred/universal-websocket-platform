# Контекст проекта

## Основная информация

**Проект:** Universal WebSocket Platform

**Миссия:** Open-source платформа для создания, настройки, развертывания и эксплуатации независимых WebSocket-серверов без написания инфраструктурного кода.

## Текущее состояние

- Текущая веха: **Beta — Complete the Single-Node Runtime**
- Статус реализации: **Control Service, single-node Runtime vertical и Configuration Loader boundary реализованы; production launch pipeline не завершён**
- Последняя завершённая development task: **реализация Configuration Loader contract DP-007**
- Текущая development task: **TASK-001 — Draft DP-008 Snapshot Builder поверх neutral `DetachedLoadResult`; architecture refinement Approved, implementation not started**
- Следующий разрешённый шаг: **Developer реализует полный уточнённый Draft DP-008 contract: exact schema `uwp.configuration` v1, private immutable detached Snapshot с полным ARCH-005 provenance и exhaustive blocking Diagnostics; Builder не подключается к production launch pipeline**
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

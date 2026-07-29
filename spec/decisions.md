# Перечень решений

В этом файле перечислены архитектурные решения и открытые вопросы. Подробное описание решений размещается в `adr/`.

## Принятые решения

- [`ADR 0001: Базовая реализация Control Service`](../docs/ru/adr/0001-bootstrap-control-service.md)
- [`ADR 0002: Configuration DSL`](../docs/ru/adr/0002-configuration-dsl.md)
- [`ADR 0003: Runtime Architecture`](../docs/ru/adr/0003-runtime-architecture.md)
- [`ADR 0004: Handshake Runtime Dependency Boundary`](../docs/ru/adr/0004-handshake-runtime-dependencies.md)

## Установленные ограничения проекта

Следующие исходные условия не заменяют ADR, Active/Frozen ARCH или Approved DP:

- Миссия проекта определена в `00-product/vision.md`.
- Архитектурные принципы определены в `01-principles/architecture-principles.md`.
- Первый Control Service реализуется на Go 1.25 с Chi Router и `slog`.

## Определённые архитектурные границы

- ADR-0003 определяет component boundaries Runtime и Provider-based composition.
- ARCH-004 определяет Runtime Instance, Launch Attempt и deployment identity
  model; минимальный in-process Runtime Lifecycle Owner реализован
  изолированно, а persistence и production routing operational сущностей
  отсутствуют.
- ARCH-005 определяет Configuration Loader, Snapshot provenance и loading boundary.
- Draft DP-007, DP-008, DP-009, DP-010 и DP-011 являются implementation
  contracts и не повышаются до нормативного статуса реализацией или commit.

## Ожидающие отдельного решения

Delivery, Persistence, Plugin ABI, production deployment adapters, operational
diagnostics и supervision требуют сфокусированных решений в соответствующих
вехах. Их отсутствие не отменяет уже определённые component, configuration и
deployment boundaries.

Draft DP-011 определяет in-process integration
`PrepareStart -> Load -> Build -> Start`. Concrete Source composition,
management API/authorization, persistence operational identities,
recovery/reconciliation и Production Activation остаются ожидающими отдельных
решений.

Package `internal/runtimelaunchflow` реализует DP-011 изолированно без
изменения этих ожидающих решения production boundaries.

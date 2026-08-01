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
- Draft DP-007–DP-015 являются implementation/design contracts и не
  повышаются до нормативного статуса реализацией или commit. DP-012 реализован
  изолированно; DP-013, DP-014 и DP-015 сохраняют Implementation Status
  Planned.

## Ожидающие отдельного решения

Delivery, Persistence, Plugin ABI, production deployment adapters, operational
diagnostics и supervision требуют сфокусированных решений в соответствующих
вехах. Их отсутствие не отменяет уже определённые component, configuration и
deployment boundaries.

Draft DP-011 определяет in-process integration
`PrepareStart -> Load -> Build -> Start`. Draft DP-012 определяет
repository-backed Source composition и реализован изолированно. Draft DP-013
определяет planned process-local management routing и
authorization-before-mutation. Non-normative Draft DP-014 предлагает candidate
focused contract ARCH-004 §19(2): durable aggregate Runtime Instance,
append-only membership Launch Attempt с immutable parent/ID/version pin и
monotonic child lifecycle facts, opaque identity namespaces, conditional
revision, atomic phase-sensitive lifecycle publication и
inspect-after-indeterminate boundary. Он не создаёт persistence
implementation, schema, API, recovery или production wiring и не снимает
formal gate §19(2) до отдельного approval/status decision.

DP-013 не разрешает isolated implementation: Active ARCH-004 §19 требует до
любого management package approval/status decision candidate contract §19(2)
и focused designs durable management idempotency,
activation/replacement/rollback, recovery/reconciliation и operational
reporting/redaction (§19(3)–(6)). Concrete authorization policy, management
implementation/API и Production Activation также отсутствуют.

Non-normative Draft DP-015 предлагает candidate focused contract ARCH-004
§19(3): opaque command identity в exact authorized scope, immutable intent,
durable claim до lifecycle delegation, same-intent replay без mutation,
per-Instance barrier для unresolved command, mandatory tracked-Start Stop и
truthful indeterminate outcome. Он не создаёт command store, schema, API,
recovery или production wiring и не снимает
formal gates §19(2) или §19(3) до отдельных approval/status decisions.

По dependency ordering следующей рекомендуемой design work после DP-015 может
быть activation/replacement/rollback ordering ARCH-004 §19(4). Рекомендация не
активирует task или implementation.

Package `internal/runtimelaunchflow` реализует DP-011 изолированно без
изменения этих ожидающих решения production boundaries.

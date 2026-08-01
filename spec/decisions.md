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
- DP-007–DP-013 сохраняют Design Status Draft; DP-014–DP-018 имеют Design
  Status Approved. Статус не повышается реализацией или commit. DP-012
  реализован изолированно; DP-013–DP-018 сохраняют Implementation Status
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
authorization-before-mutation; он Ready для bounded isolated implementation
slice. Approved DP-014 определяет focused contract ARCH-004 §19(2): durable aggregate Runtime Instance,
append-only membership Launch Attempt с immutable parent/ID/version pin и
monotonic child lifecycle facts, opaque identity namespaces, conditional
revision, atomic phase-sensitive lifecycle publication и
inspect-after-indeterminate boundary. Он не создаёт persistence
implementation, schema, API, recovery или production wiring. Design gate
§19(2) закрыт.

DP-013 разрешает только bounded isolated implementation slice. Approved
DP-014–DP-018 закрывают focused design gates ARCH-004 §19(2)–(6), но concrete
authorization policy, persistence и command implementations, private
Start-claim continuation, execution-generation binding/load gate, management
integration/API и Production Activation отсутствуют.

Approved DP-015 определяет focused contract ARCH-004
§19(3): opaque command identity в exact authorized scope, immutable intent,
durable claim до lifecycle delegation, same-intent replay без mutation,
per-Instance barrier для unresolved command, mandatory tracked-Start Stop и
truthful indeterminate outcome. Он не создаёт command store, schema, API,
recovery или production wiring. Design gate §19(3) закрыт.

Approved DP-016 определяет focused contract ARCH-004
§19(4): exact-version activation, ordered replacement через
Stop-to-proven-release, fresh-attempt explicit rollback, zero Host overlap и
phase-specific concurrency/cancellation. Для обязательного Stop-during-Starting
он требует planned private Start-claim continuation DP-011/DP-013 после sole
Owner claim и до Load; current isolated Flow этот seam не реализует. DP-016 не
создаёт lifecycle implementation, API, recovery или production wiring.
Approved DP-016 закрывает design gate §19(4); implementation остаётся
отсутствующей.

Approved DP-017 определяет focused contract ARCH-004
§19(5): exact fail-closed restart assessment, durable recovery claim,
DP-014-owned execution binding после attempt claim и до Load, разделение
last-confirmed durable facts и attempt/generation-bound execution evidence,
phase-sensitive reconciliation primitive/linked commands без lifecycle replay
и reopening admission только после coherent fully terminal verification.
Resource absence даёт Failed/interrupted; Stopped требует exact proof
Host-owned shutdown completion. Persisted Running, PID, address, time или probe
отдельно не восстанавливают ownership. DP-017 не создаёт recovery store/schema, adapter,
executor, API, automatic restart или production wiring. Design gate §19(5)
закрыт.

Approved DP-018 определяет focused contract ARCH-004
§19(6): exact failure ownership сохраняется у component boundary; только
authoritative fact проецируется в scoped allowlist operator report; valid
negative outcomes не relabel как errors; exact owner/phase precedence задаёт
stable category; unknown content fails closed; replay scoped projection version;
delivery failure не меняет lifecycle/command truth и не повторяет work. DP-018
не создаёт report model/projector/adapter, API или production wiring. Approved
DP-018 закрывает design gate §19(6), а Approved predecessor set DP-014–DP-017
уже закрывает design gates §19(2)–(5); implementation остаётся отсутствующей.

TASK-021 завершена со статусом `Completed — Coordinator Accepted`. Architect
verdict закрывает design gates ARCH-004 §19(2)–(6) и делает DP-013 Ready для
bounded isolated implementation slice; implementation автоматически не
активирована. Следующий Ready, но не активированный candidate — отдельная
Documentation-only correction root README mirrors для pre-existing
Loader-to-Builder implemented-state drift; после неё рекомендуется bounded
isolated DP-013 implementation, также не активированная.

Package `internal/runtimelaunchflow` реализует base DP-011 изолированно без
private Start-claim continuation DP-016 и без изменения этих ожидающих решения
production boundaries.

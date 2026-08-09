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
  model; минимальный in-process Runtime Lifecycle Owner и process-local
  isolated operational identity/command stores реализованы, а external durable
  persistence и production routing operational сущностей отсутствуют.
- ARCH-005 определяет Configuration Loader, Snapshot provenance и loading boundary.
- DP-007–DP-013 сохраняют Design Status Draft; DP-014–DP-019 имеют Design
  Status Approved. Статус не повышается реализацией или commit. DP-012 и
  DP-013 реализованы изолированно; DP-014, primitive boundary DP-015 и partial
  DP-019 parent/phase sequential core и command-boundary Continue/pending-Stop
  rendezvous реализованы изолированно. Полный DP-015/DP-019 extension и
  DP-016–DP-019 сохраняют Implementation Status Planned overall.

## Ожидающие отдельного решения

Delivery, Message Persistence, Plugin ABI, production deployment adapters, operational
diagnostics и supervision требуют сфокусированных решений в соответствующих
вехах. Их отсутствие не отменяет уже определённые component, configuration и
deployment boundaries.

Draft DP-011 определяет in-process integration
`PrepareStart -> Load -> Build -> Start`. Draft DP-012 определяет
repository-backed Source composition и реализован изолированно. Draft DP-013
определяет process-local management routing и authorization-before-mutation;
package `internal/runtimemanagement` реализует этот contract изолированно.
Approved DP-014 определяет focused contract ARCH-004 §19(2): durable aggregate Runtime Instance,
append-only membership Launch Attempt с immutable parent/ID/version pin и
monotonic child lifecycle facts, opaque identity namespaces, conditional
revision, atomic phase-sensitive lifecycle publication и
inspect-after-indeterminate boundary. Он не создаёт persistence
implementation сам по себе; package `internal/runtimeidentity` реализует этот
contract изолированно на process-local in-memory storage. External durable
schema, API, recovery и production wiring отсутствуют. Design gate §19(2)
закрыт.

Изолированная реализация DP-013 не разрешает integration. Approved
DP-014–DP-018 закрывают focused design gates ARCH-004 §19(2)–(6). DP-014 и
primitive boundary DP-015, partial DP-019 parent/phase sequential core и
command-boundary Continue/pending-Stop rendezvous реализованы изолированно, но
concrete authorization policy/private invoker, external persistence, managed
Start-claim continuation/OwnerClaimView, execution-generation binding/load gate, management
integration/API и Production Activation отсутствуют.

Approved DP-015 определяет focused contract ARCH-004
§19(3): opaque command identity в exact authorized scope, immutable intent,
durable claim до lifecycle delegation, same-intent replay без mutation,
per-Instance barrier для unresolved command, mandatory tracked-Start Stop и
truthful indeterminate outcome. Package `internal/runtimecommandidempotency`
реализует claim/replay store изолированно на process-local in-memory storage;
external schema, API, recovery и production wiring отсутствуют. Design gate
§19(3) закрыт.

Approved DP-016 определяет focused contract ARCH-004
§19(4): exact-version activation, ordered replacement через
Stop-to-proven-release, fresh-attempt explicit rollback, zero Host overlap и
phase-specific concurrency/cancellation. Для обязательного Stop-during-Starting
он требует planned private Start-claim continuation DP-011/DP-013 после sole
Owner claim и до Load; current isolated Flow этот seam не реализует. DP-016 не
создаёт lifecycle implementation, API, recovery или production wiring.
Approved DP-016 закрывает design gate §19(4); implementation остаётся
отсутствующей и architecture-blocked prerequisites DP-019.

Approved DP-019 определяет focused internal integration contract, необходимый
для реализации DP-016 без ослабления proofs: exact authorization tuple
OperationalDomain/Workspace/Configuration/Runtime Instance/action/target
version; callback-scoped DP-015 parent/phase claims для replacement/rollback;
private DP-011/DP-013 Start-claim continuation; publication exact Owner-issued
attempt и execution-generation binding до Load. Implementation Status DP-019 —
Planned overall; durable parent/derived-phase storage, callback capability и
strict sequential core реализованы изолированно в TASK-028, а command-boundary
Continue/pending-Stop rendezvous — в TASK-029. Он не меняет Owner
lifecycle semantics, не создаёт orchestrator/API и
не снимает TASK-026 blocker до отдельной implementation/acceptance.

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

TASK-021 закрыла design gates ARCH-004 §19(2)–(6), TASK-022 исправила root
README drift, TASK-023 реализовала bounded isolated DP-013 package, а TASK-024
реализовала bounded isolated DP-014 package. TASK-025 реализовала DP-015
package изолированно и завершена как `Completed — Coordinator Accepted`.
Concrete policy, external
persistence, management integration/API и Production Activation не
активированы и остаются отсутствующими.

Package `internal/runtimelaunchflow` реализует base DP-011 изолированно без
private Start-claim continuation DP-016 и без изменения этих ожидающих решения
production boundaries.

TASK-026 зафиксирована как `Blocked by Architecture`; упрощённый adapter
Variant B отклонён, Coordinator Acceptance/commit/publication запрещены.
Design-only TASK-027 устраняет только design ambiguity через DP-019; следующая
implementation prerequisites не активируется автоматически.
TASK-027 завершена как `Completed — Coordinator Accepted` после независимого
review `Approved` с blocking findings 0. DP-019 остаётся Approved/Planned;
Acceptance design task не снимает TASK-026 blocker и не авторизует
implementation. На момент Coordinator closure commit и publication ещё не
выполнялись; впоследствии task commit
`7ac0a6b372d9e54c73d024703e6d3ee4b06e15cd` опубликован через PR #27 и merged
как `2c017aace7e56a4747d3cecbe8ff3f6cf53e009f`.

TASK-028 реализует только independently testable partial DP-019 core в
`internal/runtimecommandidempotency`: exact Replace/Rollback intent, durable
parent/derived-phase records, callback-scoped generation-bound capabilities,
strict optional StopOld затем StartTarget order, replay, unresolved barriers и
reconstruction invalidation. На closure-time StartTarget foundation оставался
package-private; Continue/pending-Stop, private managed-Flow continuation,
attempt binding и production composition оставались Planned. TASK-029
впоследствии реализует только command-boundary Continue/pending-Stop;
остальные prerequisites остаются Planned. TASK-026 остаётся Blocked.
Repeat Independent Review TASK-028 — `Approved`, blocking/non-blocking 0;
Coordinator Closure Audit — PASS, Coordinator Acceptance — `Accepted`. На
момент closure commit и publication ещё не выполнялись; впоследствии task
commit `d28efa4e88e02ef528c78c3ca88b3f91945069ce` опубликован через PR #28 и
merged как `ba75e54e00c3cf1d0d87ca2a985acc9699698efd`.

TASK-029 завершена как `Completed — Coordinator Accepted`. Она реализует и
верифицирует только bounded isolated Continue/pending-Stop prerequisite внутри
DP-015 command boundary. Architecture PASS/0, Size Guard ACCEPT (`DO NOT
SPLIT`, net production `+680`), Independent Tester PASS WITH ENVIRONMENT
LIMITATION 0/0 и Independent Review `APPROVED` 0/0; PROCESS-002/status/parity,
links 886/0 и Scope Audit 25/0/0 — PASS. Coordinator Closure Audit — PASS,
Acceptance — `Accepted`; branch baseline
`ba75e54e00c3cf1d0d87ca2a985acc9699698efd`; commit, push и publication
TASK-029 не выполнялись. DP-019 остаётся Planned overall, а TASK-026 —
`Blocked by Architecture` до реализации и независимой приёмки exact
authorization/private managed invocation, OwnerClaim-to-DP-014 binding,
orchestrator и остальных remaining prerequisites.

Следующая рекомендация — отдельный bounded readiness/intake для lowest
remaining DP-019 prerequisite: exact orchestration authorization, private
managed invocation и OwnerClaim-to-DP-014 binding sequence. Она не активирована.

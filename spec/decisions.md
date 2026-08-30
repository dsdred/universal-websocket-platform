# Перечень решений

В этом файле перечислены архитектурные решения и открытые вопросы. Подробное описание решений размещается в `adr/`.

## Принятые решения

- TASK-049 — завершённая и Coordinator-Accepted (2026-08-28) design-only
  refinement в Draft DP-020 и Approved DP-015 для replay-first orchestration
  admission и late generation allocation; отдельная isolated implementation
  prerequisite остаётся Planned и Not Activated без Task ID, а TASK-026 —
  Blocked.

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
command-boundary Continue/pending-Stop rendezvous, policy-neutral orchestration
authorization surface, managed Flow/Start-claim continuation, OwnerClaimView,
execution-generation binding/load sequence и concrete composition-private
invoker TASK-043 реализованы изолированно. Authorization policy и external
persistence, management integration/API и Production Activation отсутствуют.

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
он требует private Start-claim continuation DP-011/DP-013 после sole Owner
claim и до Load; managed Flow/continuation, binding sequence и concrete
composition-private invoker TASK-043 реализованы изолированно. DP-016 не
создаёт lifecycle implementation, API, recovery или production wiring.
Approved DP-016 закрывает design gate §19(4); implementation остаётся
отсутствующей. Historical TASK-044 `UNBLOCK TASK-026` superseded recheck:
DP-015 tracked-Start managed-parent плюс preclaimed `StopOld` admission
prerequisite с historical matrix 7 Direct / 9 Compositional / 2 Missing core /
1 Missing prerequisite / 0 Deferred. TASK-046 фиксирует additive contract, а
TASK-047 реализует его изолированно. Fresh reassessment принимает `READY —
UNBLOCK TASK-026` с matrix 7/10/2/0/0/0 как historical readiness evidence.
DP-016 остаётся Approved/Planned; текущий implementation cycle TASK-026 теперь
Blocked после repeat Architecture `NEEDS DECISION` / `SPLIT REQUIRED`: текущий
admission DP-015/DP-020 не обеспечивает replay-first inspection и late
generation allocation. TASK-049 — завершённая и Coordinator-Accepted
design-only refinement; её отдельная isolated implementation prerequisite
остаётся Not Activated без Task ID.

Approved DP-019 определяет focused internal integration contract, необходимый
для реализации DP-016 без ослабления proofs: exact authorization tuple
OperationalDomain/Workspace/Configuration/Runtime Instance/action/target
version; callback-scoped DP-015 parent/phase claims для replacement/rollback;
private DP-011/DP-013 Start-claim continuation; publication exact Owner-issued
attempt и execution-generation binding до Load. Implementation Status DP-019 —
Planned overall; durable parent/derived-phase storage, callback capability и
strict sequential core реализованы изолированно в TASK-028, command-boundary
Continue/pending-Stop rendezvous — в TASK-029, а managed gates, continuation и
binding sequence — в TASK-037. Он не меняет Owner
lifecycle semantics, не создаёт orchestrator/API и
не реализует TASK-026 автоматически; DP-019 остаётся Planned overall, а
reactivation требует отдельного Coordinator selection.

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

Package `internal/runtimelaunchflow` реализует base DP-011 и additive managed
Start-claim continuation изолированно; TASK-043 добавляет concrete
composition-private invoker в `internal/runtimemanagement` без изменения
ожидающих решения production boundaries.

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
managed invocation и OwnerClaim-to-DP-014 binding sequence. Она активирована
как design-only TASK-030, создавшая зеркальный Draft/Planned
[DP-020: Готовность последовательности связывания оркестрации Runtime](../docs/ru/design/DP-020-runtime-orchestration-binding-sequence-readiness.md).
DP-020 фиксирует упорядоченное implementable разложение и закрывает отложенные
design-решения. Deferred slice 1 (orchestration authorizer surface) и deferred
slice 2 (private managed invoker plus managed Flow/OwnerClaimView continuation)
получили исторически принятые partial isolated реализации TASK-031/TASK-032.
TASK-034 не меняет их acceptance или Approved DP-019, но исправляет live
conformance interpretation. TASK-035 реализует Slice 2R изолированно: exact
six-field authorization, dependency-leaf binding values, all-or-none linked
identity и unique callback-scoped command-owned rendezvous identity; sole
primitive `Boundary.ExecuteManagedStart` adapter не допускает синтез binding
после legacy claim. TASK-035 независимо принята. TASK-036 затем фиксирует
managed protocol Среза 3, а TASK-037 реализует и независимо принимает Slice 3
OwnerClaim-to-DP-014 изолированно. TASK-043 реализует concrete private
composition invoker изолированно; production wiring остаётся отсутствующим.
TASK-044 впоследствии исторически фиксирует `UNBLOCK TASK-026`; superseding
recheck TASK-026 определяет DP-015 tracked-Start managed-parent плюс preclaimed
`StopOld` admission prerequisite. TASK-046 определяет atomic parent плюс
preclaimed `StopOld` admission, а TASK-047 реализует его изолированно. TASK-026
остаётся заблокированной до отдельной readiness reassessment.

TASK-036 завершена и принята как bounded design-update перед Slice 3 после
read-only readiness inspection. Draft DP-020 теперь фиксирует единый primitive/linked
managed protocol: managed parent/StartTarget adapter, command-owned opaque
early/final gates, stateless `runtimeorchestrationcontinuation`, четыре закрытых
исхода Flow и DP-014 revision-sandwich после неоднозначного результата. Approved
DP-014/DP-015/DP-019 semantics и statuses не меняются; production code, Slice 3
implementation и TASK-026 activation не выполняются.
Independent Review — Approved 0/0; PROCESS-002, Verification Matrix, Size Guard
и Scope Audit 13/0/0 — PASS. Slice 3 остаётся следующим Planned,
неактивированным implementation-срезом.

TASK-037 активирует и реализует DP-020 Slice 3 изолированно: managed
primitive/linked parent/StartTarget paths используют общий command-owned
early/final/no-claim rendezvous protocol; новый stateless
`runtimeorchestrationcontinuation` выполняет pre-mutation aggregate
scope/revision proof, DP-014 conditional attempt claim, exact returned-revision
threading в generation bind и fresh revision-sandwich convergence; managed Flow
отображает четыре закрытых исхода и после Owner claim использует
`context.WithoutCancel`. Independent code proof — PASS 0/0; Independent
Reviewer — `APPROVED` 0/0; Coordinator Acceptance — `Accepted`, поэтому
TASK-037 завершена как `Completed — Coordinator Accepted`. Approved
DP-014/DP-015/DP-019 semantics не меняются. TASK-043 реализует concrete private
composition invoker изолированно; future callback, terminal publication,
DP-015 terminalization, orchestration и production wiring отсутствуют.

TASK-038 активирует только design/readiness reassessment Среза 4. Матрица всех
19 proofs §25 DP-016 после Reviewer B-001/B-002 rework классифицирована как
7 Direct, 10 Compositional, 2 Missing, 0 Deferred; exact verdict —
`TASK-026 REMAINS BLOCKED`. Первый bounded prerequisite candidate — отдельный
design-only atomic expected-attempt Owner Stop contract; TASK-038 не
финализировала его API. TASK-039 завершена как `Completed — Coordinator
Accepted` и фиксирует принятый Draft DP-010 design:
`StopExpectedAttempt`, `StopAttemptMismatch` и
`ErrInvalidExpectedAttempt`, active-before-last selection, mismatch без
mutation и exact ordinary-Stop convergence. TASK-040 завершена как `Completed —
Coordinator Accepted`: extension реализован и верифицирован изолированно,
repeat final Reviewer `APPROVED` 0/0. Documentation-only TASK-041 завершена как
`Completed — Coordinator Accepted (2026-08-20)` после синхронизации critical
live status drift; final Reviewer `APPROVED` 0/0, commit и publication не
авторизованы и не выполнялись. Private exact-scope composition invoker design
завершён как `Completed — Coordinator Accepted (2026-08-20)` TASK-042: Tester
`PASS` 0/0, repeat Reviewer `APPROVED` 0/0, Scope Audit 17/0/0 и PROCESS-002
Synchronized. Draft DP-021 фиксирует exact contract; на closure implementation
отсутствовала. Впоследствии task commit `ebf4421` опубликован через PR #42 и
merged как `ded3aa0`. TASK-043 реализует exact invoker изолированно и завершена
как `Completed — Coordinator Accepted (2026-08-21)` после final Reviewer
`APPROVED` 0/0, Scope Audit 21/0/0 и PROCESS-002 Synchronized; DP-021 имеет
Implementation Status Partial. TASK-044 завершает отдельную repository-first
readiness reassessment remaining TASK-026 terminal/orchestrator work и
исторически фиксирует `UNBLOCK TASK-026` с matrix 7 Direct / 10 Compositional /
2 Missing core / 0 Missing external / 0 Deferred. Superseding recheck TASK-026
подтверждает missing DP-015 tracked-Start managed-parent плюс preclaimed
`StopOld` admission prerequisite и corrected matrix 7 Direct / 9 Compositional
/ 2 Missing core / 1 Missing prerequisite / 0 Deferred. DP-016 сохраняет Design
Status Approved и Implementation Status Planned;
DP-019/DP-020/DP-021 сохраняют свои статусы. TASK-044 — `Completed —
Coordinator Accepted (2026-08-24)`, repeat Reviewer `APPROVED` 0/0, Scope Audit
16/0/0 и PROCESS-002 Synchronized. Design refinement
завершён как TASK-046 `Completed — Coordinator Accepted (2026-08-25)`; он
сохраняет DP-015 Approved и фиксирует atomic tracked-Start managed-parent плюс
preclaimed `StopOld` admission. TASK-047 реализует contract изолированно.
Fresh TASK-026 Design-only reassessment принимает `READY — UNBLOCK TASK-026` с
matrix 7/10/2/0/0/0 как historical readiness evidence. Последующий
implementation cycle теперь Blocked repeat Architecture `NEEDS DECISION` /
`SPLIT REQUIRED` на DP-015/DP-020 refinement replay-first admission и late
generation. Design refinement завершена как TASK-049 и принята Coordinator
2026-08-28; её отдельная isolated implementation prerequisite остаётся Not
Activated без Task ID; implementation Acceptance/Completion отсутствуют.

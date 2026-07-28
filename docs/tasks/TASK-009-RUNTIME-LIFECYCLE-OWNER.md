# TASK-009 — Runtime Lifecycle Owner implementation

## Status

**Completed — Coordinator Accepted**

## Objective

Реализовать изолированный пакет `internal/runtimelifecycle` как минимального
in-process владельца жизненного цикла ровно одного Runtime Instance согласно
точному контракту Draft DP-010.

Результат должен реализовать двухфазный `PrepareStart -> Start`, Owner-issued
Launch Attempt, per-Instance serialization, truthful Start/Stop outcomes,
caller-cancellation и ownership Host без production wiring или дополнительной
policy.

## Selection Evidence

- Пользователь явно назначил TASK-009 и задал bounded implementation scope.
- Baseline: clean синхронизированный `main` commit
  `63b961eeb59af9205c3c3d0b68d3f4bd7b8ac25c`, содержащий завершённые
  TASK-007 и TASK-008.
- `.ai/PROJECT_CONTEXT.md` и `spec/current-state.md` рекомендуют ровно
  изолированную implementation минимального Owner по reviewed DP-010.
- DP-010 Section 31 разрешает первым code slice только
  `internal/runtimelifecycle` и local proof tests.
- Prerequisites подтверждены: neutral Loader handoff, Snapshot Builder,
  Runtime Bootstrap и stateless Runtime Launcher реализованы изолированно.
- Отклонённые alternatives:
  - Loader/Builder production wiring — future integration proof;
  - Control Service routing и management API — требуют отдельного контракта;
  - persistence operational identities — отдельно deferred;
  - retry, restart, replacement, reconciliation и supervision — прямо
    запрещены DP-010;
  - изменение Bootstrap, Host или Launcher contract — не требуется.

## Scope

- добавить только production package `internal/runtimelifecycle`;
- реализовать exact exported declarations и sentinel errors DP-010;
- реализовать immutable Owner scope, StartRequest, LaunchPreparation,
  PreparationResult, outcomes, AttemptFact и Observation;
- реализовать `PrepareStart`, `Start`, `Stop`, `Observe` и одну short-held
  synchronization boundary;
- вызывать Runtime только через существующий stateless `runtime.Launch`;
- использовать package-private immutable launch seam только для локальных
  proof tests;
- добавить local proof tests DP-010, включая race-sensitive concurrency и
  cancellation scenarios;
- синхронизировать Implementation Status DP-010 EN/RU, project state и task
  record только по фактически подтверждённой реализации;
- выполнить PROCESS-002, Tester, Scope Audit и independent Final Reviewer.

## Non-Goals

- Loader/Builder adapter или production launch pipeline;
- изменение Bootstrap, Host, Runtime Launcher или их архитектуры;
- Runtime state machine сверх exact public facts DP-010;
- restart, replacement, retry, rollback policy, recovery или reconciliation;
- supervisor, watchdog, monitoring, diagnostics policy или background workers;
- process registry, generic manager, service locator или mutable package
  registry;
- persistence, repository, HTTP API, authorization, command DTO или durable
  idempotency;
- `.github`, CI или governance changes;
- следующая integration task.

## Sources of Truth

Приоритет источников соответствует PROCESS-001:

- ADR-0002 Configuration DSL EN/RU — Accepted;
- ADR-0003 Runtime Architecture EN/RU — Accepted;
- ARCH-002 Runtime Foundation Freeze EN/RU — Active/Frozen;
- ARCH-004 Runtime Deployment and Identity Model EN/RU — Active/Approved;
- ARCH-005 Runtime Configuration Snapshot and Loading Model EN/RU —
  Active/Approved;
- DP-007 Configuration Loader Contract EN/RU — Draft, implemented in
  isolation;
- DP-008 Snapshot Builder Contract EN/RU — Draft, implemented in isolation;
- DP-009 Runtime Bootstrap Contract EN/RU — Draft, Bootstrap и Launcher
  implemented in isolation;
- DP-010 Runtime Lifecycle Owner Contract EN/RU — Draft, exact implementation
  contract этой task;
- `.ai/PROJECT_CONTEXT.md`, `spec/current-state.md`,
  `spec/decisions.md`, MASTER_PLAN EN/RU;
- existing `internal/runtime`, `internal/runtimeconfig`,
  `internal/runtimeconfigload` и tests как factual implementation evidence.

DP-010 не переопределяет источники более высокого статуса.

## Roles

- **Coordinator:** intake, task/branch gate, handoffs, scope audit, acceptance
  и closure.
- **Architect:** подтверждает существующий DP-010 contract и constraints; не
  изменяет архитектуру.
- **Documentation Agent:** baseline и финальная синхронизация PROCESS-002.
- **Developer:** минимальная production implementation и proof tests.
- **Tester:** независимо выполняет обязательные проверки и сопоставляет proof
  matrix с DP-010.
- **Reviewer:** независимо проверяет полный diff без опоры на verdict других
  ролей.
- **Publisher:** неприменим; commit и публикация запрещены этой task до
  отдельной команды пользователя.

## Branch

- исходный trusted baseline:
  `main@63b961eeb59af9205c3c3d0b68d3f4bd7b8ac25c`;
- task branch: `feature/task-009-runtime-lifecycle-owner`;
- branch создана безопасно до content changes;
- этот task record является первым content change;
- запрещены commit, push, merge, rebase, branch deletion, fetch, pull и remote
  mutation.

## Constraints

- exact public surface и semantics задаёт только DP-010;
- один Owner навсегда bound к одному Workspace, Configuration и Runtime
  Instance;
- Owner создаёт Launch Attempt и pins exact ConfigurationVersion до external
  preparation;
- один mutex сериализует claims и state publication; ID source, Launcher,
  Host Stop и waits выполняются вне mutex;
- один accepted Snapshot вызывает ровно один `runtime.Launch`;
- Host reference публикуется только после successful ready outcome;
- Stop не утверждает `ActualStopped`, пока cleanup не доказан nil outcome;
- caller cancellation после locked claim прекращает только ожидание caller;
- exact failures и outcome identity сохраняются согласно DP-010;
- different Owners не разделяют lifecycle state;
- запрещены speculative abstractions и возможности вне DP-010.

## Stop Conditions

- для реализации требуется изменить Approved/Active/Frozen architecture;
- DP-010 или связанный higher-authority source оставляет обязательную
  lifecycle semantic неоднозначной;
- требуется изменить Bootstrap, Host или exported Launcher contract;
- требуется production Loader/Builder wiring, persistence, management API,
  retry или supervision;
- baseline получает неатрибутированные изменения;
- обязательная проверка падает либо independent Reviewer выдаёт blocking
  finding.

## Acceptance Criteria

1. Task record остаётся первым content change ветки.
2. Architecture Confirmation фиксирует отсутствие изменения архитектуры и
   exact DP-010 boundary.
3. Package экспортирует только declarations и sentinels DP-010.
4. `PrepareStart` валидирует scope, выделяет ID вне lock, commits не более
   одного attempt и создаёт exact five-identity `LoadRequest`.
5. `Start` authenticates preparation, принимает first valid result, вызывает
   Launcher не более одного раза и обеспечивает same-token convergence.
6. `Stop` сериализуется и truthfully обрабатывает Preparing, Launching,
   Running, Stopping и Failed без retry cleanup.
7. Conflicting lifecycle operations не выполняются одновременно как
   независимые owners; resource calls и waits отсутствуют под state mutex.
8. Outcome, AttemptFact, Observation и exact failure semantics соответствуют
   DP-010.
9. Caller cancellation следует locked winner semantics и не отменяет
   Owner-owned work после claim.
10. Proof tests покрывают минимум successful PrepareStart/Start/Stop,
    duplicate и concurrent Start/Stop, Start while Stop, Stop while Start,
    Bootstrap/startup failure propagation, ownership release, launch attempt
    lifecycle, outcomes и cancellation.
11. `go fmt ./...`, `go test ./...`, `go vet ./...` и `git diff --check`
    успешны; `go test -race ./...` выполняется при поддержке toolchain либо
    причина недоступности фиксируется.
12. DP-010 Design Status остаётся Draft; Implementation Status и project state
    отражают только фактически доказанную isolated implementation.
13. PROCESS-002 возвращает `Synchronized`.
14. Tester возвращает PASS.
15. Scope Audit классифицирует каждый changed file как Required, Questionable
    или Removable, без unresolved Questionable/Removable.
16. Independent Final Reviewer возвращает Approved без blocking findings.
17. Commit, push и merge не выполняются.

## Verification

- targeted: `go test ./internal/runtimelifecycle -count=1`;
- formatter: `go fmt ./...`;
- full tests: `go test ./... -count=1`;
- vet: `go vet ./...`;
- race: `go test -race ./...` при поддержке toolchain;
- repository: `git diff --check`, conflict markers, unexpected files;
- documentation: EN/RU DP-010 status parity, project-state truthfulness, links;
- independent: Tester и Final Reviewer.

## Architecture Confirmation

**Confirmed — existing contract, no architecture change.**

- Accepted ADR-0002/0003, Active/Frozen ARCH-002 и Active/Approved
  ARCH-004/005 совместимы с exact DP-010 implementation surface.
- DP-010 уже определяет Owner scope, identities, lifecycle facts,
  linearization, cancellation, outcome accessors и local proof boundary;
  новой design decision для implementation не требуется.
- Package-private immutable launch seam допустим только для proof scheduling
  и результатов. Production constructor неизменно связывает его с exact
  `runtime.Launch`.
- Runtime Host остаётся единственным owner operational graph/startup/rollback;
  Owner хранит только Host reference и вызывает `Host.Stop` после handoff.
- Design Status DP-010 остаётся Draft; implementation не повышает его.

## Documentation Baseline

**Synchronized for implementation start; critical drift отсутствует.**

- Полностью прочитаны DP-010 EN/RU, PROCESS-001/002, PROJECT_CONTEXT,
  current-state и связанные authoritative ADR/ARCH/DP.
- DP-010 EN/RU имеют одинаковые 33 headings, Draft/Planned status и
  эквивалентный normative contract.
- Все связанные EN/RU document pairs существуют; их heading counts совпадают:
  ADR-0002 15, ADR-0003 16, ARCH-002 22, ARCH-004 29, ARCH-005 25,
  DP-007 53, DP-008 74, DP-009 54.
- Implementation evidence подтверждает существующие exact abstractions:
  `runtimeconfigload.LoadRequest`, `runtimeconfig.Snapshot`,
  `runtime.BootstrapRequest`, `runtime.BootstrapOutcome`,
  `runtime.DependencyBindings`, `runtime.Host` и `runtime.Launch`.
- Project state правдиво отделяет planned Owner от implemented
  Loader/Builder/Bootstrap/Launcher; обновление потребуется только после
  успешной реализации и verification.

## Implementation Handoff

**Developer implementation complete; independent gates pending.**

Production files created:

- `internal/runtimelifecycle/types.go` — exact DP-010 public declarations,
  closed results, immutable outcomes/facts/observation и sentinels;
- `internal/runtimelifecycle/owner.go` — Owner claim/state boundary,
  `PrepareStart`, `Start`, `Stop`, `Observe`, tracked launch/stop operations и
  exact production delegation в `runtime.Launch`.

Proof file created:

- `internal/runtimelifecycle/owner_test.go` — constructor/request/ID
  validation, five-identity pin, success/failure outcomes, duplicate и
  concurrent calls, both crossing Start/Stop races, cancellation,
  before/after-Running Stop, retained failed Host, ownership release,
  independent Owners и no-resource-call-under-mutex proofs.

Implementation constraints confirmed:

- Bootstrap, Host, Launcher, Loader и Builder не изменены;
- package-private launch seam immutable per Owner и не является exported
  interface, package state, registry или policy;
- production seam выполняет exact `runtime.Launch(request)` и сохраняет exact
  `BootstrapOutcome`;
- per-operation goroutines существуют только для accepted tracked launch или
  Host Stop, чтобы caller cancellation не отменяла Owner-owned duty;
- retry, restart, supervisor, watchdog, monitoring, persistence, registry и
  production wiring отсутствуют.

## Tester

**PASS — independent Tester; 0 blocking, 0 nonblocking findings.**

- Exact DP-010 exported surface и sentinels проверены через source и
  `go doc`; дополнительных lifecycle abstractions нет.
- Обязательный scenario matrix покрыт targeted proof tests: successful,
  duplicate и concurrent PrepareStart/Start/Stop, оба пересечения Start/Stop,
  preparation/Bootstrap/startup/Stop failures, ownership release, attempt
  lifecycle, outcome accessors и cancellation.
- `go test ./internal/runtimelifecycle -count=1` — PASS.
- `go test ./internal/runtimelifecycle -count=20` — PASS.
- concurrency/cancellation subset `-count=100` — PASS.
- affected Runtime regression и полный `go test ./... -count=1` — PASS.
- `go vet ./...`, package `gofmt -d`, `git diff --check`, whitespace и
  conflict checks — PASS.
- EN/RU parity и 107 relative links, 0 broken — PASS.
- Race detector недоступен: Windows/amd64, `CGO_ENABLED=0`, default run
  возвращает `-race requires cgo`; при forced CGO отсутствуют usable race
  runtime/toolchain и `gcc`.
- Tester файлов не изменял.
- После FR-001/FR-002 repeat Tester независимо подтвердил `PASS`, 0 blocking
  и 0 nonblocking findings: focused retention proofs, full targeted, stress
  concurrency/cancellation `-count=100`, full tests, vet, formatting и
  diff-check прошли; same-token convergence не регрессировала.

## PROCESS-002

**Synchronized pending independent confirmation.**

- DP-010 EN/RU: Design Status остаётся Draft, Implementation Status изменён на
  `Implemented in isolation`; production wiring остаётся явно отсутствующим.
- Design indexes EN/RU и MASTER_PLAN debt зеркально отражают isolated
  components без production pipeline.
- `.ai/PROJECT_CONTEXT.md`, `spec/current-state.md` и `spec/decisions.md`
  отражают фактическую isolated implementation и сохраняют deferred
  operational entities/integration.
- Root README и CHANGELOG неприменимы: exported public product/API capability
  и release artifact не появились.
- ADR и ARCH не изменены: архитектурное решение не менялось; утверждение
  ARCH-004 об отсутствии production Owner/integration остаётся правдивым.
- DP-007/008/009 не изменены: их собственные implementation boundaries и
  integration-gated claims не менялись.
- EN/RU DP-010 structure: 33/33 headings, 6/6 fences.
- Relative links checked: 107; broken: 0.

## Scope Audit

**Accepted independently: 14 Required, 0 Questionable, 0 Removable.**

Required production/test:

- `internal/runtimelifecycle/types.go` — exact declarations/outcomes/facts;
- `internal/runtimelifecycle/owner.go` — DP-010 Owner behavior;
- `internal/runtimelifecycle/owner_test.go` — mandatory local proof matrix.

Required documentation/process:

- DP-010 EN/RU и design indexes EN/RU — factual implementation status/parity;
- MASTER_PLAN EN/RU — isolated components versus remaining pipeline debt;
- `.ai/PROJECT_CONTEXT.md`, `spec/current-state.md`, `spec/decisions.md` —
  implemented/planned boundary;
- `docs/tasks/README.md` и этот TASK-009 record — process navigation/evidence.

Forbidden-scope audit:

- Bootstrap, Launcher, Host, Loader, Builder, Control Service, `.github`/CI,
  governance и Publisher не изменены;
- production integration, persistence, retry/reconciliation, supervision,
  registry и policy отсутствуют;
- generated, accidental и formatting-only files отсутствуют;
- следующая task не начата.

Questionable/Removable disposition не требуется: findings отсутствуют.

После FR-001/FR-002 repeat Scope Audit подтвердил тот же полный состав:
14 Required, 0 Questionable, 0 Removable; retention fixes/proofs относятся
непосредственно к ownership acceptance criteria, новых путей и forbidden scope
не появилось.

## Final Reviewer

**Initial verdict: Needs Revision — two blocking findings.**

- **FR-001:** `attemptState` сохранял accepted `runtimeconfig.Snapshot` после
  завершения Launcher, хотя convergence требует только stored outcomes.
  Поскольку authentic historical attempts сохраняются на lifetime Owner, это
  удерживало полный Snapshot дольше Host lifetime и конфликтовало с ownership
  ARCH-004/DP-009.
- **FR-002:** successful `runtime.BootstrapOutcome` также сохранялся в
  `attemptState` и косвенно удерживал Host после successful Stop.
- **Rework:** retained Snapshot field и assignment удалены; Snapshot теперь
  живёт только в tracked launch operation/request до завершения
  `runtime.Launch`, после чего на success независимое значение owned Host, а
  Owner хранит только active Host reference/outcomes/facts. Launch outcome
  сохраняется в attempt только для failure; successful outcome с Host не
  удерживается.
- **Proof:** `TestAttemptStateDoesNotRetainSnapshot` запрещает Snapshot field в
  retained attempt state; successful Stop дополнительно проверяет отсутствие
  direct Host и Host через retained BootstrapOutcome; existing convergence
  tests доказывают, что stored results не требуют этих references.
- **Documentation rework:** DP-010 EN/RU зеркально уточнён: accepted Snapshot
  хранится только tracked launch operation до возврата Launcher, historical
  convergence использует stored outcome; exact preparation failure остаётся
  retained. Ownership table и local proof #10 синхронизированы.
- **Первый repeat verdict:** `Needs Revision`. DR-001 blocking подтвердил
  оставшуюся на момент review старую формулировку DP-010; DR-002 nonblocking
  указал неверный count initial findings.
- **DR rework:** DR-001 устранён зеркально в Section 14, ownership table и
  local proof #10; DR-002 исправлен на `two blocking findings`. DP-010
  structure остаётся 33/33 headings и 6/6 fences; `git diff --check` PASS.
- **Final repeat verdict:** `Approved`, 0 blocking и 0 nonblocking findings.
  Reviewer независимо перепроверил полный 14-file scope, DP-010 EN/RU,
  ownership retention/release, concurrency и cancellation semantics, proofs,
  PROCESS-002 state и отсутствие forbidden integration/`.github` changes.

## Handoff

- implementation, proof tests, documentation synchronization и независимые
  quality gates завершены;
- diff готов к отдельно разрешаемому commit;
- commit, push и merge не выполнялись.

## Publication

- на момент closure stage, commit и publication не выполнялись; это
  historical closure fact, а не durable запрет на отдельно разрешённое
  последующее действие;
- commit authority и Publisher authority проверяются по актуальной команде
  пользователя, а не по сохранённому transient state;
- publication не разрешена этой task автоматически.

## Next Candidate

- следующая Ready work не активирована;
- production integration этой task не начинается автоматически.

## Closure

- Final status: `Completed — Coordinator Accepted`.
- Closed by: Coordinator после Tester PASS, Scope Audit
  `14 Required / 0 Questionable / 0 Removable` и Final Reviewer
  `Approved / 0 blocking / 0 nonblocking`.
- Date: 2026-07-28.

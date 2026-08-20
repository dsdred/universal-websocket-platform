# TASK-043 — Private Exact-Scope Managed Start Invoker Implementation

## Status

`Completed — Coordinator Accepted (2026-08-21)`.

## Task Contract

### Task Mode

`Implementation`. Реализовать только изолированный private exact-scope managed
Start invoker, определённый Draft/Planned DP-021, внутри существующего package
`internal/runtimemanagement`, без orchestration, terminal work или production
wiring.

### Why Now

- TASK-042 завершена, опубликована и зафиксировала единственный bounded
  implementation contract DP-021;
- prerequisites TASK-035, TASK-037 и TASK-040 реализованы и независимо приняты;
- project-state sources одинаково рекомендуют отдельный intake exact invoker;
- TASK-026 остаётся Blocked и не может возобновляться до принятия remaining
  prerequisites;
- invoker является наименьшим независимо проверяемым slice перед будущей
  callback/orchestrator integration.

### Definition of Done

1. `internal/runtimemanagement` содержит immutable scope-bound managed Start
   invoker с constructor validation domain/Target/non-nil preconstructed
   `ManagedFlow` и exact stable sentinels DP-021.
2. Единственная invocation operation валидирует receiver, non-nil context,
   `StartRequest`, exact six-field authorization tuple, Target, closed
   primitive/linked binding shape до Owner mutation.
3. Успешный вызов синхронно делегирует stored `ManagedFlow.StartManaged` ровно
   один раз с unchanged context, request и binding и возвращает exact
   `StartOutcome`/error без wrapping или terminal mapping.
4. Already-cancelled non-nil context не отклоняется invoker и достигает Flow;
   nil context и mismatch дают zero Owner/Load/Build/Launcher work.
5. Invoker не создаёт Flow, не вызывает command/identity boundaries, не хранит
   per-call binding/rendezvous и не добавляет accessor, registry, goroutine,
   fallback либо public management route.
6. Focused proof/regression tests покрывают применимые DP-021 acceptance rows;
   непроверяемые без future callback/orchestrator composition строки явно
   остаются integration limitations, а isolated capability не представляется
   как Production Activation.
7. Focused/full/stress/race-or-limitation, vet, formatting, diff, dependency,
   documentation parity/link и independent review gates завершены.
8. PROCESS-002 правдиво отражает isolated implementation, сохраняет DP-021
   Draft Design Status и TASK-026 Blocked.

### Out of Scope

- future orchestrator-owned callback closure и result mapping;
- DP-014 terminal publication и DP-015 command/phase terminalization;
- activation/replacement/rollback orchestrator TASK-026;
- изменение `Directory` public behavior, transport, policy, persistence,
  recovery, reporting, supervision или production wiring;
- изменение Approved DP-014–DP-019 semantics/status;
- новый package, dependency/module change, legacy/unmanaged fallback;
- stage, commit, push, PR, merge, publication или branch cleanup.

### Verification Plan

- до test edits зафиксировать Existing Coverage Report и mapping DP-021 rows;
- focused package tests с coverage и targeted validation/delegation/cancellation
  proofs, shuffled/stress repetitions и full `go test ./... -count=1`;
- focused race detector; при недоступности — точная environment limitation и
  substitute stress evidence;
- `go vet ./...`, `gofmt`, `go mod tidy` no-diff check, exported GoDoc/API audit
  и dependency-direction inspection;
- `git diff --check`, exact file-set, conflict/whitespace/generated checks;
- PROCESS-002 EN/RU parity, links, statuses и planned-vs-implemented audit;
- independent Tester и Reviewer verdicts после реализации.

## Objective

Добавить минимальный изолированный DP-021 invoker, который проверяет exact
immutable scope и structural binding до lifecycle mutation и делегирует одному
preconstructed managed Flow без принятия orchestration или terminal ownership.

## Selection Evidence

Clean synchronized baseline `main@ded3aa0` не содержит active task. TASK-042,
`.ai/PROJECT_CONTEXT.md`, `spec/current-state.md`, `spec/decisions.md` и
MASTER_PLAN EN/RU одинаково называют следующим неактивированным candidate
отдельный implementation/readiness intake exact DP-021 invoker. DP-021 уже
задаёт ownership, validation, dependency и failure boundaries; prerequisite
implementations TASK-035/037/040 находятся в baseline.

Отклонены: resume TASK-026 (всё ещё Blocked), combined invoker+terminal work
(несколько independently shipped behaviors), production composition/API,
recovery/reporting и новая design task (contract уже записан DP-021).

## Scope

Initial production/test scope:

- новый invoker implementation file в `internal/runtimemanagement`;
- один focused invoker test file в том же package;
- этот task record;
- только документы, доказанно необходимые final PROCESS-002.

Расширение production scope за существующий package или изменение Directory,
managed Flow/binding contracts требует stop и Coordinator reassessment.

## Sources of Truth

- PROCESS-001 и PROCESS-002;
- ADR-0003, Frozen ARCH-002 и Active ARCH-004;
- DP-011, DP-013, Approved DP-014–DP-019, DP-020 и Draft DP-021 EN/RU;
- TASK-035, TASK-037, TASK-040 и TASK-042 closure evidence;
- фактические packages `runtimemanagement`, `runtimelaunchflow`,
  `runtimelifecycle` и `runtimeorchestrationbinding`;
- project-state sources, task/design indexes и MASTER_PLAN EN/RU.

## Roles

- Coordinator: selection, gates, scope audit, acceptance и next recommendation;
- Architect: explicit conformance/readiness verdict до implementation;
- Documentation Agent: baseline и final PROCESS-002 synchronization;
- Developer: только confirmed invoker production/test scope;
- Tester: Existing Coverage Report и independent verification;
- Reviewer: independent final review и deletion test;
- Publisher: не применяется без отдельных user gates.

## Branch

- trusted baseline: clean synchronized `main@ded3aa0`;
- task branch: `feature/task-043-private-exact-scope-invoker`;
- branch создан безопасно; этот record — первый content change;
- git history/remote mutations запрещены.

## Constraints and Stop Conditions

- остановиться при противоречии DP-021 источнику более высокого уровня;
- остановиться, если exact invoker нельзя реализовать без callback/orchestrator,
  terminal ownership, second authorization или public bypass;
- остановиться при critical documentation drift, новом package/dependency,
  Size Guard trigger, mandatory check failure или blocking review finding;
- structural binding validity не объявляется live authority;
- downstream outcome/error identities сохраняются unchanged;
- следующая task не активируется автоматически.

## Existing Coverage Report — Intake

- **Existing Coverage:** текущие tests доказывают Target/Binding/Directory
  validation и routing, managed Flow request/binding matching, pre-claim
  cancellation/`StartNoClaim`, post-claim continuation и primitive/linked
  `StartExecutionBinding.Valid()` semantics.
- **Coverage Gap:** concrete invoker, constructor errors, exact cross-scope
  validation order, one-call unchanged delegation и no-storage/no-bypass
  surface отсутствуют.
- **Added Proof Tests:** focused constructor, invalid receiver/context/request/
  binding/scope, primitive/linked unchanged delegation, already-cancelled
  context и exact downstream outcome/error identity tests добавлены в
  `managed_start_invoker_test.go`; new functions покрыты на 100%.
- **Added Regression Tests:** package regression, shuffled/stress repetitions,
  relevant-package и full-suite runs подтверждают отсутствие изменения
  existing Directory/Flow/binding behavior.
- **Remaining Limitations:** future callback custody, DP-015 replay admission,
  terminal mapping/publication/terminalization, orchestrator и production
  composition не могут быть интеграционно доказаны этим isolated slice.

## Size Guard

Final re-reassessment: **`DO NOT SPLIT`**. Exact set содержит 2 Go files и 19
обязательных документов PROCESS-002. Это один isolated behavior в одном
existing package, zero new package/contract/dependency; остальные файлы —
неразделимые EN/RU mirrors, indexes и durable-state synchronization. Split
оставил бы implementation и authoritative status заведомо рассинхронизированными.
Production diff далёк от 500 строк и второго independently shipped behavior
нет. DP-016 EN/RU добавлены только после blocking final-review drift finding;
их mirror pair обязательна и не меняет Approved/Planned overall semantics.
Любой двадцать второй файл либо изменение Approved/Frozen semantics требует
новой остановки и reassessment.

## Documentation Baseline

Initial Documentation Agent verdict: **`Drift Detected — bounded, non-critical;
implementation not blocked`**. EN/RU trees имели 46/46 документов без
unmatched paths; DP-013/019/020/021, design indexes и MASTER_PLAN сохраняли
structural/status parity; 248 scoped relative links были valid, 0 broken.
Baseline correctly recorded Planned/absent invoker and Blocked TASK-026, но
TASK-043 ещё не была отражена в indexes/live state, а live TASK-042 publication
wording устарело после PR #42. Эти exact drift входят в final sync ниже.

## Architecture Confirmation

Architect verdict: **`PASS`**, blocking findings `0`.

Mapping к Draft DP-021:

- `ManagedStartInvoker` хранит только copied operational domain/Target и
  borrowed preconstructed `ManagedFlow`;
- `NewManagedStartInvoker` валидирует domain, Target и non-nil Flow exact
  sentinel `ErrInvalidManagedStartInvoker`;
- `InvokeManagedStart` валидирует receiver, non-nil context, request,
  structurally valid binding и exact six-field scope/request tuple до Flow;
- successful invocation один раз синхронно делегирует unchanged context,
  request и binding в `StartManaged` и возвращает exact outcome/error;
- already-cancelled non-nil context достигает Flow/`StartNoClaim`; command,
  identity, callback, terminal и public routing ownership не добавлены.

## Developer Handoff

Developer изменил только:

- `internal/runtimemanagement/managed_start_invoker.go` — isolated invoker и
  два stable sentinel;
- `internal/runtimemanagement/managed_start_invoker_test.go` — focused proof и
  regression tests.

Новый package, module dependency, Directory/API route, goroutine, registry,
legacy fallback, callback, terminal mapping/publication или production wiring
не добавлены.

## Verification, Review, Scope Audit and Documentation Sync

Independent Tester verdict: **`PASS WITH ENVIRONMENT / DECLARED INTEGRATION
LIMITATIONS`**, blocking findings `0`, non-blocking findings `0`.

Exact evidence:

- focused `go test ./internal/runtimemanagement -count=1 -cover` PASS,
  package coverage 96.3%, new functions 100%;
- focused shuffled/stress `-count=50` PASS;
- relevant packages `-count=1` и shuffled `-count=10` PASS;
- full `go test ./... -count=1` PASS;
- `go vet ./...`, `gofmt -d`, `go mod tidy -diff`, unchanged `go.mod`/`go.sum`,
  dependency/API/GoDoc и diff checks PASS;
- default race unavailable: `go: -race requires cgo; enable cgo by setting
  CGO_ENABLED=1`; retry with `CGO_ENABLED=1` failed before tests with
  `runtime/race: package testmain: cannot find package`;
- substitute focused `-count=50`, relevant shuffled `-count=10` и full suite
  PASS.

Declared integration limitations: future DP-015 callback custody/replay,
terminal mapping/publication/terminalization, TASK-026 orchestrator and
production composition/wiring remain unimplemented and unproved by this
isolated slice.

Initial independent Reviewer verdict: **`Needs Revision`**, one blocking
finding B-001. Separate `managedStartBindingShapeValid` helper/condition
duplicated the closed-shape guarantee already enforced by
`StartExecutionBinding.Valid()` and was `Removable` under the deletion test.
Developer removed only that duplicate. Repeat independent Reviewer verdict:
**`APPROVED`**, blocking findings `0`, non-blocking findings `0`.

Final independent Reviewer затем выдал **`Needs Revision`** с blocking B-001:
Approved DP-016 EN/RU в трёх cited live implementation-status sites каждого
mirror всё ещё называл private exact-scope/concrete composition invoker
отсутствующим, противореча TASK-043 и Partial DP-021; downstream boundary также
сохранял stale later-prerequisite wording. Bounded documentation rework
добавляет только обязательную mirror pair DP-016 и исправляет это factual
implementation wording; Design Status Approved и Implementation Status Planned
overall не меняются. Repeat final Reviewer после rework вынес **`APPROVED`**,
blocking findings `0`, non-blocking findings `0`.

Final Documentation Agent sync updates only the exact 19-document set. DP-021
remains Draft and becomes Partial / implemented in isolation; DP-013 remains
Draft/Implemented in isolation; DP-019 remains Approved/Planned overall;
DP-020 remains Draft/Planned overall; DP-016 remains Approved/Planned overall.
Concrete invoker now exists, while future callback/terminal/orchestrator/
production composition remains absent and TASK-026 remains Blocked.
Final checks, Scope Audit и Coordinator Acceptance завершены.

### Final Coordinator Scope Audit

Result: **21 Required / 0 Questionable / 0 Removable**.

Disposition полного exact set:

1. `internal/runtimemanagement/managed_start_invoker.go` — `Required`; напрямую
   обеспечивает Definition of Done 1–5.
2. `internal/runtimemanagement/managed_start_invoker_test.go` — `Required`;
   обеспечивает proof/regression evidence Definition of Done 6–7.
3. Этот TASK-043 record — `Required`; task record и evidence обязательны
   всегда.
4. `docs/tasks/README.md` — `Required`; navigation и live task status.
5–6. DP-013 EN/RU — 2 `Required`; factual ownership/implementation boundary и
   обязательная semantic parity.
7–8. DP-019 EN/RU — 2 `Required`; factual prerequisite status, remaining
   limitations и semantic parity без изменения Approved design.
9–10. DP-020 EN/RU — 2 `Required`; factual slice/dependency status, remaining
   limitations и semantic parity.
11–12. DP-021 EN/RU — 2 `Required`; Draft/Partial implementation boundary,
   isolated proofs/limitations и semantic parity.
13–14. DP-016 EN/RU — 2 `Required`; factual invoker implementation status,
    preserved Approved/Planned overall semantics и обязательная mirror parity.
15–16. Design indexes EN/RU — 2 `Required`; navigation/status parity.
17–19. `.ai/PROJECT_CONTEXT.md`, `spec/current-state.md`, `spec/decisions.md` —
    3 `Required`; active task/capability durable truth и устранение live
    publication drift TASK-042.
20–21. MASTER_PLAN EN/RU — 2 `Required`; durable dependency/roadmap state и
    mirror parity.

Deletion test: ни один файл нельзя удалить, сохранив Definition of Done,
executable proof, обязательную parity/navigation и отсутствие live
contradiction. Questionable и Removable changes отсутствуют.

Audit также подтверждает: следующая task преждевременно не начата; unrelated
behavior/refactoring, generated, formatting-only и planned-as-implemented
changes отсутствуют; staged files `0`, generated files `0`. Итог:
**21 Required / 0 Questionable / 0 Removable**; audit `PASS`.

### Final Checks and Acceptance

- final independent Reviewer: **`APPROVED`**, blocking `0`, non-blocking `0`;
- focused/full/stress substitutes, vet, formatting, module, dependency/API/
  GoDoc и diff checks: `PASS`; race сохраняет точно описанную environment
  limitation;
- EN/RU parity, statuses, planned-vs-implemented boundary и links: `PASS`;
- stale live task/publication/invoker claims, conflict markers, trailing
  whitespace, staged, generated и unexpected files: `0`;
- PROCESS-002: **`Synchronized`**;
- Coordinator Closure Audit: **`PASS`**;
- Coordinator Acceptance: **`Accepted (2026-08-21)`**.

Process Health Review: `Not triggered` — это не десятая completed task trigger
в текущем cycle; rollback, escaped defect и repeating Publisher failure не
было; task возвращалась ровно два раза, а trigger требует более двух возвратов.

### PROCESS-002 Applicability

- task record: applicable, always; synchronized with implementation, evidence,
  review, final checks and closure;
- `spec/current-state.md`: applicable — factual capability/task boundary
  changed;
- MASTER_PLAN EN/RU: applicable — durable prerequisite status and next
  recommendation changed;
- DP-013/016/019/020/021 EN/RU: applicable — linked implementation status and
  absent-invoker claims changed without Design Status promotion or change to
  DP-016/019 Planned-overall semantics;
- `.ai/PROJECT_CONTEXT.md`: applicable — current task, fundamental capability
  boundary and TASK-042 terminal publication state changed;
- `spec/decisions.md`: applicable — live implementation boundary and TASK-042
  publication evidence changed;
- task/design indexes EN/RU: applicable — navigation/status changed;
- root README: `Not applicable` — no public product or entry-point behavior;
- ARCH/ADR and their indexes: `Not applicable` — no architecture decision or
  Approved/Frozen semantic changed;
- `spec/README.md`: `Not applicable` — specification navigation set unchanged;
- EN/RU root and roadmap indexes: `Not applicable` — no indexed document set
  changed;
- `CHANGELOG.md`: `Not applicable` — isolated internal capability, no
  user-facing or release behavior.

## Commit Gate and Publication

Not authorized. Staged files `0`; commit, push, PR, merge и publication не
выполнялись.

## Next Candidate

Не активирован. После TASK-043 требуется отдельный repository-first bounded
readiness intake remaining TASK-026 terminal/orchestrator work по authoritative
dependency order. Его `Ready` status этой task не доказан; TASK-026 остаётся
Blocked до отдельного intake и устранения всех remaining blockers.

## Closure

- Final status: `Completed — Coordinator Accepted`;
- Closed by: Coordinator;
- Date: `2026-08-21`;
- Scope Audit: `21 Required / 0 Questionable / 0 Removable`;
- Review: final independent Reviewer `APPROVED`, blocking/non-blocking `0/0`;
- Documentation: PROCESS-002 `Synchronized`;
- Publication: commit/push/PR/merge/publication not authorized and not
  performed;
- Next candidate: not activated; readiness not asserted; TASK-026 remains
  Blocked.

# TASK-007 — Implementation prerequisites Runtime Lifecycle Owner

## Status

**Completed — Coordinator Accepted**

## Objective

Зеркально определить минимальный concrete implementation contract начального
in-process Runtime Lifecycle Owner по ARCH-004 до production-реализации или
Loader-to-Builder-to-Launcher wiring.

Результат должен сделать следующий изолированный implementation slice
однозначным, не проектируя persistence, management HTTP API, automatic retry,
replacement, reconciliation или production pipeline целиком.

## Selection Evidence

- Autonomous entry: точная bare-команда `Продолжай проект.`.
- Active tasks отсутствуют; TASK-006 имеет статус
  `Completed — Coordinator Accepted` и merged в `main` через PR #7.
- Baseline: clean `main` на `e791482`; `origin/main` указывает на тот же commit.
- TASK-006, `.ai/PROJECT_CONTEXT.md` и `spec/current-state.md` однозначно
  рекомендуют focused Lifecycle Owner prerequisite refinement.
- ARCH-004 требует focused design до реализации и уже фиксирует ownership,
  per-instance serialization, Launch Attempt, desired/actual state и Host
  handoff.
- Loader, Builder, Bootstrap и Launcher реализованы изолированно; production
  implementation Owner и pipeline ещё не Ready без concrete Owner API,
  serialization/outcome и local/integration proof contract.
- Slice является prerequisite текущей Beta milestone, documentation-only и не
  требует изменения Approved ADR либо Active/Frozen architecture.
- Отклонённые alternatives:
  - production Runtime Lifecycle Owner — contract ещё недостаточно конкретен;
  - Loader-to-Builder-to-Launcher wiring — зависит от Owner contract;
  - persistence Runtime Instance/Launch Attempt — отдельный focused design;
  - management HTTP API, authorization и idempotency — отдельный command
    contract;
  - retry, replacement, rollback и reconciliation — явно deferred ARCH-004;
  - operational diagnostics и supervision — отдельные roadmap gaps.
- Post-merge drift: `.ai/PROJECT_CONTEXT.md` всё ещё описывает accepted
  TASK-006 как dirty task-ветку и разрешает только её commit, хотя Git history
  уже содержит merge; design index также ошибочно называет реализованный
  DP-008 planned. Оба расхождения подлежат PROCESS-002 в этой
  documentation-only task и не меняют selection.

## Scope

- создать зеркальный Draft DP-010 с concrete минимальным контрактом
  in-process Runtime Lifecycle Owner;
- определить package placement, construction/input surface, single-instance
  state ownership и read-only observation;
- определить serialization и linearization Start/Stop для одного Runtime
  Instance;
- определить Launch Attempt identity/provenance input, вызов только
  `runtime.Launch`, успешный Host handoff и truthful failure publication;
- определить caller-cancellation, concurrent/repeated call и Host Stop
  ownership semantics в минимальной локальной границе;
- разделить local proof и будущие integration proofs;
- обновить design indexes EN/RU, project state и task index;
- устранить ограниченный factual post-merge drift TASK-006 и stale DP-008
  index status;
- выполнить documentation checks, PROCESS-002, scope audit и независимый
  review.

## Non-Goals

- production Go code или тесты Runtime Lifecycle Owner;
- Loader-to-Builder-to-Launcher production wiring;
- repository/persistence schema Runtime Instance или Launch Attempt;
- management HTTP API, authorization, durable idempotency или command DTO;
- automatic retry, restart, replacement, rollback, recovery или
  reconciliation;
- cross-process execution, PID supervision, scheduling или clustering;
- изменение Host, Bootstrap, Launcher, Loader, Builder или Control Service;
- повышение Design Status DP-010, DP-009, DP-008 или DP-007;
- изменение Approved ADR, Active/Frozen ARCH либо unrelated documentation.

## Sources of Truth

- `docs/ru/adr/0002-configuration-dsl.md` и EN mirror;
- `docs/ru/adr/0003-runtime-architecture.md` и EN mirror;
- `docs/ru/architecture/ARCH-002-runtime-foundation-freeze.md` и EN mirror;
- `docs/ru/architecture/ARCH-004-runtime-deployment-and-identity-model.md` и EN
  mirror;
- `docs/ru/architecture/ARCH-005-runtime-configuration-snapshot-and-loading-model.md`
  и EN mirror;
- `docs/ru/design/DP-007-configuration-loader-contract.md`,
  `DP-008-snapshot-builder-contract.md`, `DP-009-runtime-bootstrap-contract.md`
  и EN mirrors;
- `internal/runtime/bootstrap.go`, `internal/runtime/launcher.go`,
  `internal/runtime/host.go` и их tests как implementation evidence;
- TASK-005 и TASK-006;
- `.ai/PROJECT_CONTEXT.md`, `spec/current-state.md`, `spec/decisions.md` и
  MASTER_PLAN EN/RU.

Draft DP-010 не переопределяет источники более высокого статуса.

## Roles

- **Coordinator:** selection, task/branch gate, handoffs, scope audit,
  acceptance, project-state update и next recommendation.
- **Architect:** определяет bounded DP-010 contract, constraints, acceptance
  proofs и implementation boundary; production code не меняет.
- **Documentation Agent:** выполняет baseline audit, зеркально записывает
  утверждённое Architect решение и проводит PROCESS-002.
- **Tester:** проверяет EN/RU parity, links, structure, status honesty и
  repository checks; production tests не добавляет.
- **Reviewer:** независимо проверяет sources, architecture, completeness,
  parity и scope; автор documentation changes не выполняет final review.
- **Developer:** неприменим, поскольку production code и tests вне scope.

## Branch

- trusted baseline: clean `main`, commit `e791482`;
- task branch: `docs/task-007-runtime-lifecycle-owner-prerequisites`;
- branch создана безопасно до content changes;
- этот task record является первым content change;
- запрещены stage, commit, push, merge, rebase, branch deletion, fetch, pull и
  remote mutation без отдельного разрешения.

## Constraints

- Owner остаётся Control Service-side orchestration responsibility и не
  переносит Control Plane authority в Runtime Host;
- один Owner scope управляет ровно одним Runtime Instance;
- все state-changing operations этого Instance используют одну serialization
  boundary без network/resource wait под state lock;
- каждый accepted Start создаёт новую Launch Attempt identity и один Snapshot
  provenance; identity не генерируется Launcher или Host;
- Owner вызывает только stateless `runtime.Launch`, не `Bootstrap` или
  `Host.Start()` напрямую;
- active Host reference публикуется только после успешного Launch outcome;
- actual state остаётся правдивым относительно Host readiness и terminal
  cleanup;
- configuration loading, Snapshot build и persistence остаются внешними
  dependencies/boundaries, а не скрытыми обязанностями Owner;
- API должен оставаться минимальным и in-process, без speculative generic
  manager, repository или policy framework.

## Stop Conditions

- contract требует изменения Active ARCH-004 или Frozen ARCH-002;
- невозможно отделить минимальный in-memory Owner от persistence/management
  command semantics;
- требуется выбрать public HTTP API, repository schema, retry/replacement или
  recovery policy;
- authoritative sources конфликтуют;
- baseline получает неатрибутированные изменения;
- EN/RU parity невозможно сохранить;
- обязательная проверка падает или independent Reviewer возвращает blocking
  finding;
- scope требует production code или materially расширяется.

## Acceptance Criteria

1. Этот task record является первым content change task-ветки.
2. Documentation baseline отделяет factual drift от architecture work и не
   обнаруживает неустранённого critical drift.
3. Architect явно подтверждает bounded DP-010 contract и его совместимость с
   ADR-0003, ARCH-002, ARCH-004, ARCH-005 и DP-009.
4. DP-010 EN/RU зеркально определяет concrete минимальную in-process surface,
   ownership, states, Start/Stop serialization и linearization.
5. Contract однозначно определяет Launch Attempt identity/provenance input,
   единственный вызов Launcher, successful Host handoff и failure outcome.
6. Contract определяет concurrent/repeated Start/Stop, caller cancellation и
   отсутствие resource wait под serialization lock.
7. Contract запрещает persistence, HTTP API, retry/replacement,
   reconciliation, diagnostics policy и pipeline wiring в первом
   implementation slice.
8. Local acceptance proofs отделены от будущих Loader/Builder/management и
   production integration proofs.
9. Design Status DP-010 остаётся Draft, Implementation Status — Planned.
10. Design indexes, `.ai/PROJECT_CONTEXT.md`, `spec/current-state.md` и task
    index синхронизированы без planned-as-implemented claims.
11. PROCESS-002 имеет итог `Synchronized`.
12. Documentation verification и independent final review успешны.
13. Scope Audit содержит только `Required`, 0 unresolved `Questionable` и 0
    `Removable`.
14. Closure рекомендует ровно один следующий Ready bounded slice либо честно
    фиксирует отсутствие Ready work; следующая task не запускается.

## Verification

- зафиксировать task-before-work chronology через Git diff;
- проверить EN/RU links и mirror structure DP-010 и design indexes;
- проверить status, normative keywords, headings, tables и fenced blocks;
- проверить отсутствие production/test changes;
- проверить references на stale TASK-006 branch/commit gate и DP-008 planned
  status;
- выполнить `git diff --check` и поиск conflict markers;
- выполнить независимый architecture/documentation review.

## PROCESS-002 Applicability

Обязателен: task создаёт новый mirrored design contract, меняет design indexes
и project-state navigation, а также устраняет factual post-merge drift.

## Handoff

### Coordinator Intake

- **Task:** TASK-007.
- **Selection:** deterministic explicit next candidate после TASK-006.
- **Current stage:** Documentation Baseline, затем explicit Architecture
  Confirmation.
- **Required next action:** собрать точный inventory drift и сформировать
  Architect handoff для bounded DP-010.
- **Forbidden:** production implementation, pipeline wiring и deferred
  management/persistence policy.

### Documentation Baseline

- **Status:** `Ready for Architect handoff`; critical documentation drift не
  обнаружен, architecture work может продолжаться в пределах task contract.
- **Inventory authoritative architecture:** ADR-0002 и ADR-0003 имеют
  `Accepted` и зеркально сохраняют Configuration-first, immutable Published
  source и explicit Runtime dependency boundaries; ARCH-002 имеет
  `Active`/`Frozen`, а ARCH-004 и ARCH-005 — `Active`/`Approved architectural
  model`. Их EN/RU mirrors согласованы по статусам, структуре и нормативному
  смыслу. ARCH-002 оставляет внешнюю launch boundary вне freeze; ARCH-004
  назначает Lifecycle Owner Control Service-side orchestration responsibility;
  ARCH-005 закрепляет Owner-to-Loader-to-Builder-to-Launcher flow и provenance
  Runtime Instance/Launch Attempt.
- **Inventory implementation contracts:** DP-007, DP-008 и DP-009 сохраняют
  Design Status `Draft`. Сами DP зеркально и правдиво отмечают Loader, Builder,
  Bootstrap и Runtime Launcher как implemented in isolation, а Lifecycle Owner
  и production pipeline — как planned/unimplemented. DP-009 AP-003/AP-011
  остаются integration-gated.
- **Inventory implementation evidence:** `internal/runtime/bootstrap.go`,
  `launcher.go`, `host.go` и tests подтверждают только isolated Bootstrap,
  exact stateless Launcher и Host lifecycle/composition behavior. Runtime
  Lifecycle Owner, operational Runtime Instance/Launch Attempt entities и
  production wiring отсутствуют; documented architecture отделена от
  implemented state.
- **Inventory project state and navigation:** `.ai/PROJECT_CONTEXT.md`,
  `spec/current-state.md`, `spec/decisions.md`, MASTER_PLAN EN/RU, design
  indexes EN/RU, TASK-005, TASK-006 и task index проверены. `spec/decisions.md`
  согласован с ADR/ARCH; MASTER_PLAN зеркально и правдиво сохраняет production
  launch pipeline как architectural debt.
- **Parity evidence:** для перечисленных ADR, ARCH, DP, MASTER_PLAN и design
  indexes совпадают количества headings и fenced blocks в EN/RU; mirror links
  и statuses присутствуют. Новый DP на baseline stage не создан.
- **Drift F-001 — factual operational state, non-critical:** `.ai/PROJECT_CONTEXT.md`
  и `spec/current-state.md` сохраняют pre-merge TASK-006 dirty-worktree и
  commit-only gate, тогда как Git history подтверждает merge PR #7 в clean
  `main` `e791482`, а TASK-007 уже активна в своей documentation branch.
  Причина однозначна; correction назначена final PROCESS-002 этой task и не
  меняет architecture selection.
- **Drift F-002 — factual implementation status, non-critical:** design
  indexes EN/RU называют DP-008 `planned`, хотя DP-008, MASTER_PLAN,
  `spec/current-state.md` и production evidence согласованно подтверждают
  Builder `implemented in isolation`. Correction назначена зеркальному
  pre-implementation documentation/PROCESS-002 update; Design Status `Draft`
  повышать запрещено.
- **Drift F-003 — navigation, resolved in baseline:** task index не содержал
  TASK-007 после initial task-record gate. Добавлена ссылка на существующий
  record; task-before-work chronology сохранена, поскольку TASK-007 остаётся
  первым content change ветки.
- **No-change disposition:** ADR, ARCH, DP-007/008/009, MASTER_PLAN,
  `spec/decisions.md` и production code не требуют baseline correction.
  F-001/F-002 не являются критическими и не разрешают выдавать planned Owner
  contract за approved или implemented.
- **Open findings and risks:** architecture blocker отсутствует; до
  documentation synchronization остаётся риск stale operational navigation,
  ограниченный F-001/F-002. Любое требование выбрать persistence, HTTP API,
  retry/replacement, recovery policy или изменить ARCH-002/004/005 является
  stop condition.
- **Required next action:** Architect должен подтвердить bounded DP-010
  contract либо вернуть blocker; Documentation Agent только после этого может
  зеркально записать утверждённое решение без status promotion.

### Intermediate Reviewer Finding

- **Verdict:** `Needs Revision`.
- **B-001 — blocking, authority/ordering:** DP-010 вынес allocation Launch
  Attempt ID за Owner и начал local surface с уже построенного Snapshot.
  ARCH-004 назначает создание Launch Attempt Runtime Lifecycle Owner, а
  ARCH-005 требует Owner сначала создать attempt и pin exact Published
  ConfigurationVersion до Loader/Builder. Contract должен сохранить эту
  ordering и ownership без необходимости менять surface при будущей
  integration.
- **B-002 — blocking, concrete surface/concurrency:** DP-010 назвал
  `StartOutcome`, `StopOutcome` и `Observation`, но не зафиксировал exported
  state/category declarations и accessors. Также не определены concrete
  operation/terminal facts, same-attempt Start в `Stopping` и linearization
  гонки already-cancelled caller context с claim.
- **Confirmed passes:** Draft/Planned honesty, local/integration proof split,
  Host terminal-signal limitation, EN/RU structure/status/index parity,
  DP-008 index correction, отсутствие production/test changes и
  `git diff --check`.
- **Coordinator disposition:** оба blocking finding приняты. Architecture
  rework возвращён Architect; PROCESS-002 closure и следующий implementation
  slice запрещены до зеркальной correction и повторного independent review.

### Architecture Confirmation — Revised after B-001/B-002/R-001/R-002

- **Verdict:** `Ready`; B-001, B-002, R-001 и R-002 разрешены без изменения
  источников более высокого статуса. Все предыдущие варианты contract
  полностью superseded.
- **B-001 disposition:** Owner создаёт Launch Attempt и pin точной
  ConfigurationVersion до Loader/Builder. Two-phase surface возвращает
  immutable preparation с готовым neutral `runtimeconfigload.LoadRequest`;
  production wiring Loader/Builder остаётся вне первого implementation slice.
- **Exact surface:** Documentation Agent должен использовать следующие
  declarations без расширения:

```go
package runtimelifecycle

type DesiredState string
const (
    DesiredStopped DesiredState = "stopped"
    DesiredRunning DesiredState = "running"
)

type ActualState string
const (
    ActualStopped  ActualState = "stopped"
    ActualStarting ActualState = "starting"
    ActualRunning  ActualState = "running"
    ActualStopping ActualState = "stopping"
    ActualFailed   ActualState = "failed"
)

type LaunchAttemptIDSource func() (runtimeconfigload.LaunchAttemptID, error)

type StartRequest struct { /* immutable Workspace/Configuration/version */ }
func NewStartRequest(workspaceID, configurationID, configurationVersionID uint64) StartRequest
func (r StartRequest) WorkspaceID() uint64
func (r StartRequest) ConfigurationID() uint64
func (r StartRequest) ConfigurationVersionID() uint64

type LaunchPreparation struct { /* opaque owner/claim token */ }
func (p LaunchPreparation) LoadRequest() runtimeconfigload.LoadRequest
func (p LaunchPreparation) Context() context.Context

type PreparationResultKind string
const (
    PreparationSnapshot PreparationResultKind = "snapshot"
    PreparationFailure  PreparationResultKind = "failure"
)

type PreparationResult struct { /* closed success/failure union */ }
func PreparedSnapshot(snapshot runtimeconfig.Snapshot) PreparationResult
func FailedPreparation(cause error) PreparationResult
func (r PreparationResult) Kind() PreparationResultKind
func (r PreparationResult) Snapshot() (runtimeconfig.Snapshot, bool)
func (r PreparationResult) Failure() (error, bool)

type StartOutcomeKind string
const (
    StartRunning              StartOutcomeKind = "running"
    StartPreparationFailed    StartOutcomeKind = "preparation-failed"
    StartLaunchFailed         StartOutcomeKind = "launch-failed"
    StartStoppedBeforeRunning StartOutcomeKind = "stopped-before-running"
)

type StartOutcome struct { /* immutable */ }
func (r StartOutcome) Kind() StartOutcomeKind
func (r StartOutcome) Attempt() AttemptFact
func (r StartOutcome) PreparationFailure() (error, bool)
func (r StartOutcome) LaunchOutcome() (runtime.BootstrapOutcome, bool)

type StopOutcomeKind string
const (
    StopStopped StopOutcomeKind = "stopped"
    StopFailed  StopOutcomeKind = "stop-failed"
)

type StopOutcome struct { /* immutable */ }
func (r StopOutcome) Kind() StopOutcomeKind
func (r StopOutcome) Attempt() (AttemptFact, bool)
func (r StopOutcome) Failure() (error, bool)

type AttemptPhase string
const (
    AttemptPreparing  AttemptPhase = "preparing"
    AttemptLaunching  AttemptPhase = "launching"
    AttemptRunning    AttemptPhase = "running"
    AttemptStopping   AttemptPhase = "stopping"
    AttemptHistorical AttemptPhase = "historical"
)

type StopOrigin string
const (
    StopNotClaimed    StopOrigin = ""
    StopBeforeRunning StopOrigin = "before-running"
    StopAfterRunning  StopOrigin = "after-running"
)

type AttemptTerminalKind string
const (
    AttemptNotTerminal          AttemptTerminalKind = ""
    AttemptPreparationFailed    AttemptTerminalKind = "preparation-failed"
    AttemptLaunchFailed         AttemptTerminalKind = "launch-failed"
    AttemptStoppedBeforeRunning AttemptTerminalKind = "stopped-before-running"
    AttemptStopped              AttemptTerminalKind = "stopped"
    AttemptStopFailed           AttemptTerminalKind = "stop-failed"
)

type AttemptFact struct { /* immutable identities/phase/category, no cause */ }
func (f AttemptFact) WorkspaceID() uint64
func (f AttemptFact) ConfigurationID() uint64
func (f AttemptFact) ConfigurationVersionID() uint64
func (f AttemptFact) RuntimeInstanceID() runtimeconfigload.RuntimeInstanceID
func (f AttemptFact) LaunchAttemptID() runtimeconfigload.LaunchAttemptID
func (f AttemptFact) Phase() AttemptPhase
func (f AttemptFact) StopOrigin() StopOrigin
func (f AttemptFact) RunningPublished() bool
func (f AttemptFact) TerminalKind() AttemptTerminalKind

type Observation struct { /* immutable */ }
func (s Observation) RuntimeInstanceID() runtimeconfigload.RuntimeInstanceID
func (s Observation) WorkspaceID() uint64
func (s Observation) ConfigurationID() uint64
func (s Observation) DesiredState() DesiredState
func (s Observation) ActualState() ActualState
func (s Observation) ActiveAttempt() (AttemptFact, bool)
func (s Observation) LastAttempt() (AttemptFact, bool)

type Owner struct { /* unexported synchronized state */ }
func NewOwner(
    workspaceID uint64,
    configurationID uint64,
    instanceID runtimeconfigload.RuntimeInstanceID,
    nextAttemptID LaunchAttemptIDSource,
    dependencies *runtime.DependencyBindings,
) (*Owner, error)
func (o *Owner) PrepareStart(request StartRequest) (LaunchPreparation, error)
func (o *Owner) Start(
    ctx context.Context,
    preparation LaunchPreparation,
    result PreparationResult,
) (StartOutcome, error)
func (o *Owner) Stop(ctx context.Context) (StopOutcome, error)
func (o *Owner) Observe() Observation
```

- **Sentinel errors:** exact names `ErrInvalidOwner`,
  `ErrInvalidStartRequest`, `ErrAttemptIDSourceFailed` с exact wrapped source
  error, `ErrAttemptIDReused`, `ErrStartConflict`,
  `ErrPreparationNotOwned` и `ErrInvalidPreparationResult`.
  `ErrPreparationConsumed` намеренно отсутствует: same-token convergence не
  сравнивает повторные result values. Conflict/validation errors не меняют
  lifecycle state.
- **Construction:** Owner навсегда bound к одному Workspace, Configuration и
  Runtime Instance. Пустые identities, nil ID source и nil bindings pointer
  отклоняются. Bindings заимствуются stable. ID source лишь выделяет opaque
  value; именно Owner создаёт attempt и pin. Source вызывается вне mutex;
  failure, empty или duplicate ID не создаёт attempt. Losing concurrent
  allocations могут остаться unused, но Launch Attempt становится только ID,
  committed на claim.
- **PrepareStart:** request содержит exact ConfigurationVersion identity без
  attempt ID/Snapshot и должен совпасть с Owner Workspace/Configuration. После
  allocation launch-claim LP под mutex допустим только из `Stopped` либо
  `Failed` без Host/active attempt. LP создаёт attempt, pin version, строит
  `runtimeconfigload.NewLoadRequest`, создаёт Owner-owned preparation/start
  context, отмечает ID used, публикует desired `Running`, actual `Starting`,
  phase `Preparing`. Opaque preparation non-forgeable между Owners/claims;
  `LoadRequest()` является единственным Loader input, `Context()` —
  read-only.
- **Start acceptance:** принимает только exact live preparation этого Owner и
  closed `PreparationResult`. Prepared Snapshot до mutation обязан совпасть со
  всеми пятью identity pinned LoadRequest. Mismatch возвращает
  `ErrInvalidPreparationResult`, не consume preparation и не вызывает Launch.
  Для live unaccepted token первый valid result под mutex выигрывает и
  сохраняется exact; acceptance LP переводит `Preparing -> Launching` либо
  terminalizes exact preparation failure. После acceptance каждый later Start
  с тем же authentic token полностью игнорирует supplied result, даже zero или
  different, и attach к stored operation/outcome без equality, deep,
  comparability или semantic-equivalence check. Snapshot path после unlock
  ровно один раз вызывает fixed `runtime.Launch`.
- **Outcomes:** preparation failure без Stop сохраняет exact error, очищает
  active attempt, actual `Failed`, desired `Running`, terminal
  `PreparationFailed`, Launch отсутствует. Launch failure аналогично сохраняет
  exact BootstrapOutcome с terminal `LaunchFailed`. Launch success публикует
  Host/Running только после readiness success.
- **Stop/convergence:** claim Stop сохраняет immutable `StopOrigin` и
  `RunningPublished`. Stop до publication Running использует
  `StopBeforeRunning`; same-token Start возвращает
  `StartStoppedBeforeRunning` при success или failure Stop, а late Host
  получает ровно один Stop без publication Running. Stop после publication
  Running использует `StopAfterRunning`; immutable stored `StartRunning`
  возвращается same-token Start во время Stopping и после success или failure
  Stop. Historical PreparationFailed/LaunchFailed возвращают stored exact
  outcomes. Forged token — `ErrPreparationNotOwned`; authentic consumed token
  всегда игнорирует новый result argument.
- **Stop terminal proof:** Host.Stop использует Owner-owned
  `context.Background()` вне mutex. Nil является единственным proof для
  Stopped и очистки Host. Non-nil публикует Failed/desired Stopped, сохраняет
  Host/attempt и exact error; repeated Stop сходится на stored failure без
  retry, новый Start запрещён.
- **B-002 context-race disposition:** `PrepareStart` не имеет caller context.
  `Start` и `Stop` под mutex непосредственно перед acceptance/claim проверяют
  `ctx.Err()`. Non-nil возвращается без mutation/attachment. Nil и следующая
  mutation под тем же lock выигрывают race; последующая cancellation влияет
  только на ожидание caller, Owner operation продолжается. Runtime calls,
  channel/resource/context waits отсутствуют под mutex.
- **Observation:** coherent desired/actual и active/last `AttemptFact`, включая
  Stop origin, Running publication и terminal category; не раскрывает Host,
  cancellation, dependencies или raw causes. Exact failures доступны только
  convergent operation outcomes.
- **Truthfulness:** unexpected `Running -> Failed` остаётся integration-gated
  до Host terminal signal; polling `Running()` запрещён.
- **Local proofs:** constructor/source/identity; PrepareStart ordering и exact
  LoadRequest; concurrent single claim; provenance mismatch zero Launch;
  first-valid-result-wins для non-comparable Snapshot/error без equality;
  preparation/launch exact failures; origin-sensitive Start truthfulness при
  Stop до и после Running; same-token convergence; one Host.Stop; locked
  context race; caller cancellation wait-only; conservative Stop failure; no
  waits under lock; coherent capability-safe observation; independent Owners;
  race tests.
- **Integration/deferred:** фактический Loader/Builder use, production
  Owner-to-Launcher routing, one Owner per Instance, authorization, durable
  allocation/history/idempotency/recovery и Host supervision остаются
  integration-gated. Persistence, HTTP, retry/replacement/reconciliation,
  diagnostics, Stop policy, process supervision, Host changes и generic
  framework остаются deferred. DP-010 — Draft/Planned.

### First Documentation Rework — B-001/B-002

- **Historical disposition:** B-001 исправлен Owner-issued attempt/version pin
  в `PrepareStart` до Loader/Builder; B-002 исправлен exact surface и locked
  `ctx.Err()` claim rule.
- **Preserved evidence:** one-phase external-attempt/prebuilt-Snapshot contract
  удалён, two-phase ownership и concrete declarations сохранены во втором
  rework.
- **Superseded detail:** первоначальная формулировка same-token convergence и
  unconditional outcome в Stopping оказалась недостаточной и заменена
  финальным Architect handoff ниже.

### Repeat Reviewer Finding

- **Verdict:** `Needs Revision`.
- **R-001 — blocking, origin-sensitive truthfulness:** прежний contract
  безусловно превращал same-token Start в `StartStoppedBeforeRunning` во время
  Stopping. Это ретроактивно отменяло уже опубликованный successful
  `StartRunning`, когда Stop был claimed после Running.
- **R-002 — blocking, undefined result equivalence:** формулировка
  `materially identical` и `ErrPreparationConsumed` требовала несуществующего
  сравнения Snapshot/error values. Snapshot или errors могут быть
  non-comparable, а pointer/deep/text/`errors.Is` equivalence не определяет
  identity accepted preparation result.
- **Confirmed passes:** B-001 two-phase authority/order, B-002 exact base
  surface и locked context race, Draft/Planned honesty, proof split, indexes,
  ссылки и отсутствие production changes.
- **Coordinator disposition:** оба finding приняты; final PROCESS-002 и
  implementation запрещены до второго зеркального rework и repeat review.

### Second Documentation Rework Handoff — R-001/R-002

- **Status:** `Ready for repeat independent review`; R-001/R-002 исправлены
  зеркально, но closure требует Reviewer confirmation.
- **R-001 correction:** добавлены exact `StopOrigin` (`StopNotClaimed`,
  `StopBeforeRunning`, `StopAfterRunning`) и
  `AttemptFact.RunningPublished()`. Facts фиксируются при Stop claim и не
  регрессируют. Before-Running origin всегда возвращает
  `StartStoppedBeforeRunning`; after-Running origin во время Stopping и после
  success/failure Stop сохраняет immutable stored `StartRunning`.
- **R-001 outcome honesty:** `LaunchOutcome()` доступен только для primary
  `StartLaunchFailed`. Secondary late Launch failure при
  `StartStoppedBeforeRunning` сохраняется internal и не смешивается с primary
  Start outcome; raw Stop error раскрывает только `StopOutcome`.
- **R-002 correction:** добавлены `PreparationResultKind`,
  `PreparationSnapshot`, `PreparationFailure` и `Kind()`. Для live unaccepted
  preparation первый valid result, accepted под mutex, выигрывает и сохраняется
  exact.
- **R-002 convergence:** после first acceptance каждый later Start с тем же
  authentic token игнорирует supplied result полностью, включая zero или
  different non-comparable Snapshot/error, и attach к stored
  operation/outcome. Equality, pointer, deep, text, `errors.Is`, chain,
  comparability и semantic-equivalence checks запрещены.
- **Removed surface:** `ErrPreparationConsumed` удалён из DP-010 и task
  contract; foreign/forged token по-прежнему получает
  `ErrPreparationNotOwned`, invalid first result —
  `ErrInvalidPreparationResult`.
- **Preserved corrections:** Owner-bound Workspace/Configuration/Instance,
  Owner-issued attempt/version pin, exact `LoadRequest`, closed preparation,
  one-call Launcher, locked context race, conservative Host retention и
  integration deferrals не изменены.
- **Status and scope:** DP-010 остаётся `Draft`/`Planned`; Loader/Builder
  adapter, project-state synchronization и production code не изменялись.
- **Verification:** DP-010 EN/RU — 33/33 headings, 6/6 fenced markers и 21/21
  table rows; 3/3 Go blocks совпадают дословно; declarations
  `PreparationResultKind`, `StopOrigin`, `RunningPublished` и first-result
  semantics присутствуют зеркально. `ErrPreparationConsumed`, `materially
  identical/different` и mixed `LaunchOutcome` отсутствуют в normative DP;
  их упоминания в TASK-007 ограничены historical finding/disposition. Relative
  links, trailing whitespace, conflict markers и `git diff --check` — PASS;
  production/project-state changes отсутствуют.
- **Open findings:** blocking finding со стороны Documentation Agent
  отсутствует; R-001/R-002 ожидают repeat Reviewer confirmation.
- **Required next action:** Reviewer повторно проверяет origin-sensitive
  outcomes, first-result-wins semantics, exact surface, EN/RU parity и
  отсутствие stale equality/consumption terms. Только после `Approved`
  разрешён final PROCESS-002.

### Repeat Independent Reviewer Approval

- **Verdict:** `Approved`.
- **Blocking findings:** 0.
- **Nonblocking findings:** 0.
- **B-001/B-002:** Owner-issued attempt/version pin до Loader/Builder, exact
  exported surface и locked caller-context claim rule подтверждены.
- **R-001:** immutable `StopOrigin`/`RunningPublished` сохраняют
  `StartRunning` после Running-origin Stop и
  `StartStoppedBeforeRunning` после before-Running Stop при любом Stop result.
- **R-002:** first-valid-result-wins и unconditional ignore позднего
  same-token result исключают undefined equality/comparability и повторную
  lifecycle mutation.
- **Status/scope:** DP-010 остаётся `Draft`/`Planned`; local proof и future
  integration gates разделены; Owner, Loader/Builder adapter и production
  pipeline не представлены реализованными.
- **Coordinator disposition:** review handoff принят как gate для final
  PROCESS-002. Task остаётся `In Progress`; Tester, Scope Audit, final checks и
  Coordinator Acceptance ещё обязательны.

### Final PROCESS-002 Handoff

- **Status:** `Synchronized`; последующие Tester/Scope Audit/final checks и
  Coordinator Acceptance pending.
- **DP-010 and indexes:** зеркальные DP-010 сохраняют reviewed two-phase
  contract со статусами Design `Draft` и Implementation `Planned`; design
  indexes EN/RU отражают DP-010 planned и factual DP-008 implemented in
  isolation.
- **Project state:** `.ai/PROJECT_CONTEXT.md` и `spec/current-state.md`
  синхронизированы с Git baseline `e791482`, active TASK-007 branch/stage,
  approved review evidence и truthful отсутствием Owner/production pipeline.
  Stale TASK-006 dirty-worktree/commit-only gate удалён.
- **Next candidate honesty:** отдельная isolated implementation минимального
  Owner package по reviewed Draft DP-010 указана только как рекомендация после
  closure TASK-007. Следующая task/branch не созданы, implementation не
  начата; production Loader/Builder wiring и deferred boundaries не стали
  Ready автоматически.
- **Explicit no-change decisions:**
  - `spec/decisions.md` — новый Accepted/Approved architecture decision или
    status transition отсутствует;
  - MASTER_PLAN EN/RU — production launch pipeline по-прежнему отсутствует,
    milestone/debt ordering не изменились;
  - root README EN/RU и `CHANGELOG.md` — Draft/Planned internal contract не
    создаёт user-visible или release capability;
  - ADR, ARCH и DP-007/008/009 — normative contracts и statuses не менялись.
- **Navigation:** task index уже содержит TASK-007; task-before-work chronology
  сохранена.
- **Implementation boundary:** production code/tests не изменялись. Runtime
  Lifecycle Owner, management routing, persistence, Loader/Builder adapter,
  retry/reconciliation и Host supervision остаются unimplemented/deferred.
- **Changed in final PROCESS-002:** `.ai/PROJECT_CONTEXT.md`,
  `spec/current-state.md` и этот task record. MASTER_PLAN, decisions, root
  README, CHANGELOG, ADR/ARCH/other DP и production files не изменялись.
- **Verification:** DP-010 EN/RU structure остаётся 33/33 headings, 6/6 fenced
  markers, 21/21 table rows и 3/3 exact Go blocks; relative links — PASS;
  stale TASK-006 operational/commit gate и stale PROCESS-002 pending claims в
  project-state files отсутствуют; trailing whitespace, conflict markers и
  `git diff --check` — PASS; MASTER_PLAN и `spec/decisions.md` unchanged.
- **Task status:** TASK-007 намеренно не помечена Completed; closure принадлежит
  Coordinator после оставшихся gates.
- **Required next action:** Tester выполняет final documentation verification,
  затем Coordinator проводит Scope Audit/final checks и acceptance.

### Final Tester Handoff

- **Verdict:** `PASS`; blocking findings отсутствуют.
- **Scope:** 8 changed documentation files, 5 tracked modified и 3 untracked;
  0 out-of-scope и 0 production/test changes.
- **Parity:** DP-010 EN/RU — 33/33 headings, 6/6 fence markers, 21/21 table
  rows и 3/3 exact Go blocks.
- **Links/status/contract:** 81 relative links checked, 0 broken; Design
  `Draft`, Implementation `Planned`; 23/23 required contract tokens в каждом
  mirror; normative stale equality/consumption terms отсутствуют.
- **Project state:** stale TASK-006 gate отсутствует; TASK-007 правдиво
  находится в Final Verification; Owner package и production pipeline
  отсутствуют.
- **Repository checks:** trailing whitespace 0, conflict markers 0,
  `git diff --check` PASS. Отдельный markdown toolchain в репозитории и
  окружении отсутствует.
- **Required next action:** Coordinator Scope Audit, final independent review
  и acceptance.

### Scope Audit

| Файл | Классификация | Доказательство необходимости |
| --- | --- | --- |
| `docs/tasks/TASK-007-RUNTIME-LIFECYCLE-OWNER-PREREQUISITES.md` | `Required` | Task contract, role handoffs, rework evidence, verification, audit и closure PROCESS-001. |
| `docs/tasks/README.md` | `Required` | Навигация к active task после initial task-record gate. |
| `docs/en/design/DP-010-runtime-lifecycle-owner-contract.md` | `Required` | EN mirror нового bounded implementation contract AC-004–AC-009. |
| `docs/ru/design/DP-010-runtime-lifecycle-owner-contract.md` | `Required` | RU mirror нового bounded implementation contract AC-004–AC-009. |
| `docs/en/design/README.md` | `Required` | Навигация/status DP-010 и factual correction DP-008. |
| `docs/ru/design/README.md` | `Required` | Зеркальная навигация/status DP-010 и factual correction DP-008. |
| `.ai/PROJECT_CONTEXT.md` | `Required` | PROCESS-002: active task, reviewed planned contract, truthful next gate и удаление stale TASK-006 state. |
| `spec/current-state.md` | `Required` | PROCESS-002: то же factual project state и planned/implemented split. |

- **Questionable:** 0.
- **Removable:** 0.
- **Production/test/generated files:** отсутствуют.
- **Premature next-task work:** отсутствует; package
  `internal/runtimelifecycle`, следующая task и branch не созданы.
- **Unrelated behavior/refactoring/architecture:** отсутствуют. Approved ADR,
  Active/Frozen ARCH, production code и tests не изменялись.
- **Status honesty:** DP-010 остаётся Draft/Planned; Owner, Loader/Builder
  adapter, management routing и pipeline не представлены реализованными.
- **Coordinator disposition:** 8 Required, 0 Questionable, 0 Removable. Audit
  принят; следующий gate — final independent review полного diff.

### Final Reviewer Finding

- **Verdict:** `Needs Revision`.
- **F-001 — blocking, project-state gate:** `.ai/PROJECT_CONTEXT.md` и
  `spec/current-state.md` сохраняли инструкции выполнить Tester и Scope Audit,
  хотя Final Tester Handoff уже имел `PASS`, а Coordinator Scope Audit был
  принят с 8 Required, 0 Questionable и 0 Removable. Новый агент мог ошибочно
  повторить завершённые gates.
- **Scope:** finding ограничен factual operational state. DP-010 contract,
  Draft/Planned honesty, EN/RU parity, links, proof boundary и восьмифайловый
  scope audit новых findings не получили.
- **Required correction:** project-state должен фиксировать Tester PASS и
  accepted Scope Audit, оставляя pending только repeat final Reviewer,
  Coordinator Acceptance и отдельно разрешённый commit.
- **Coordinator disposition:** F-001 принят и возвращён Documentation Agent;
  TASK-007 остаётся `In Progress`.

### Final Reviewer Rework Handoff — F-001

- **Status:** `Ready for repeat final review`; TASK-007 не помечена Completed.
- **Corrected project-state files:** `.ai/PROJECT_CONTEXT.md` и
  `spec/current-state.md` теперь фиксируют Final Tester `PASS` и Coordinator
  Scope Audit accepted: 8 Required, 0 Questionable, 0 Removable.
- **Changed in F-001 rework:** `.ai/PROJECT_CONTEXT.md`,
  `spec/current-state.md` и этот task record.
- **Remaining gates:** только repeat final Reviewer полного diff и Coordinator
  Acceptance. Commit остаётся отдельным разрешённым действием; push/merge не
  выполняются неявно.
- **Preserved state:** reviewed DP-010 остаётся `Draft`/`Planned`; Owner,
  Loader/Builder adapter, management routing и production pipeline остаются
  unimplemented. Next candidate не начат.
- **Unchanged:** DP-010 EN/RU, indexes, task index, MASTER_PLAN, decisions,
  production code и tests в рамках F-001 rework не менялись.
- **Verification:** stale pending Tester/Scope Audit instructions в обоих
  project-state files отсутствуют; TASK-007 status остаётся
  `In Progress — Final Verification`; trailing whitespace и conflict markers
  отсутствуют; `git diff --check` — PASS.
- **Required next action:** repeat final Reviewer подтверждает устранение
  stale gate; затем Coordinator выполняет acceptance.

### Final Reviewer Approval

- **Verdict:** `Approved`.
- **Blocking findings:** 0.
- **Nonblocking findings:** 0.
- **F-001:** project-state правдиво фиксирует Final Tester PASS и accepted
  Scope Audit 8/0/0; завершённые gates больше не представлены pending.
- **Contract:** two-phase Owner-issued attempt/version pin, exact surface,
  first-valid-result-wins, origin-sensitive Stop, truthful outcomes,
  cancellation race и local/integration proof boundary соответствуют final
  Architect handoff.
- **Documentation:** DP-010 EN/RU parity, Draft/Planned status, indexes,
  project state, links и scope подтверждены; Owner и production pipeline не
  представлены реализованными.
- **Required next action:** Coordinator Acceptance.

### Coordinator Closure

- **Final status:** `Completed — Coordinator Accepted`.
- **Completed scope:** создан зеркальный reviewed Draft DP-010 с
  Implementation Status `Planned`; определены bounded two-phase prerequisites
  минимального in-process Runtime Lifecycle Owner; design/task indexes и
  project state синхронизированы; factual DP-008 index drift и stale TASK-006
  operational gate устранены.
- **Changed files — 8 Required:**
  1. `docs/tasks/TASK-007-RUNTIME-LIFECYCLE-OWNER-PREREQUISITES.md`;
  2. `docs/tasks/README.md`;
  3. `docs/en/design/DP-010-runtime-lifecycle-owner-contract.md`;
  4. `docs/ru/design/DP-010-runtime-lifecycle-owner-contract.md`;
  5. `docs/en/design/README.md`;
  6. `docs/ru/design/README.md`;
  7. `.ai/PROJECT_CONTEXT.md`;
  8. `spec/current-state.md`.
- **Architecture alignment:** ADR-0002/0003, Frozen ARCH-002, Active
  ARCH-004/005 и Draft DP-007/008/009 не изменены. DP-010 реализует только
  implementation contract утверждённых boundaries и не повышает собственный
  Design Status.
- **Verification:** Final Tester `PASS`; 81 relative links checked, 0 broken;
  DP-010 parity — 33/33 headings, 6/6 fenced markers, 21/21 table rows и 3/3
  exact Go blocks; required contract tokens — 23/23 в каждом mirror; trailing
  whitespace и conflict markers — 0; `git diff --check` — PASS.
- **Review:** Final Reviewer `Approved`, 0 blocking и 0 nonblocking findings
  после rework B-001/B-002, R-001/R-002 и F-001.
- **PROCESS-002:** `Synchronized`.
- **Scope Audit:** 8 Required, 0 Questionable, 0 Removable; production, test,
  generated, unrelated и premature next-task changes отсутствуют.
- **Known limitations:** Runtime Lifecycle Owner package и tests, actual
  Loader/Builder adapter, management routing/API, persistence, durable
  identity/idempotency, retry/reconciliation, diagnostics и Host supervision
  не реализованы. DP-010 остаётся Draft/Planned; AP-003/AP-011 остаются
  integration-gated.
- **Commit readiness:** принятый восьмифайловый documentation diff готов к
  отдельно разрешённому commit. Stage, commit, push и merge не выполнялись.
- **Next recommended work:** отдельная изолированная production implementation
  минимального `internal/runtimelifecycle` Owner и local proof tests строго по
  reviewed Draft DP-010.
- **Not activated:** следующая task/branch/code не созданы и не начаты.
- **Closed by:** Coordinator.
- **Date:** 2026-07-28.

## Next Candidate

- **Рекомендуемая Ready work после closure:** отдельная изолированная
  production implementation минимального `internal/runtimelifecycle` Owner и
  local proof tests строго по reviewed Draft DP-010.
- **Boundary:** без Loader/Builder production wiring, persistence, management
  HTTP API, retry/replacement/reconciliation, diagnostics, supervision или
  Host changes.
- **Явно не начата:** task/branch/code следующей work не созданы.

# DP-010: Контракт Runtime Lifecycle Owner

[English version](../../en/design/DP-010-runtime-lifecycle-owner-contract.md)

## 1. Статус

**Design Status:** Draft

**Implementation Status:** base Lifecycle Owner и расширение expected-attempt
Stop реализованы изолированно

**Статус архитектуры:** implementation contract утверждённой модели
operational identity из
[ARCH-004](../architecture/ARCH-004-runtime-deployment-and-identity-model.md)
и модели loading из
[ARCH-005](../architecture/ARCH-005-runtime-configuration-snapshot-and-loading-model.md).

Runtime Lifecycle Owner реализован изолированно в
`internal/runtimelifecycle`. Production wiring
Loader-to-Builder-to-Launcher не реализован. Контракт
`StopExpectedAttempt`, добавленный TASK-039, реализован и верифицирован
изолированно завершённой и Coordinator-Accepted TASK-040; его private invoker,
integration и production
wiring отсутствуют. Этот Draft не пересматривает утверждённую архитектуру.

## 2. Назначение

Определить минимальный in-process Control Service-side owner, который:

- владеет scope ровно одного Workspace, Configuration и Runtime Instance;
- создаёт каждый Launch Attempt и pin его exact ConfigurationVersion до работы
  Loader или Builder;
- сериализует lifecycle operations этого Instance;
- вызывает только существующий stateless Runtime Launcher;
- владеет active Host reference; и
- публикует правдивые desired, actual, attempt и terminal facts.

Первая реализация изолирована. Persistence, management API и production wiring
остаются deferred.

## 3. Источники полномочий

Контракт ограничен следующими источниками:

- [ADR-0002](../adr/0002-configuration-dsl.md): Published
  ConfigurationVersion является immutable источником поведения;
- [ADR-0003](../adr/0003-runtime-architecture.md): зависимости явны, а Runtime
  не получает доступ к Repository или management API;
- [ARCH-002](../architecture/ARCH-002-runtime-foundation-freeze.md): Host
  является production composition root и владеет startup и shutdown;
- [ARCH-004](../architecture/ARCH-004-runtime-deployment-and-identity-model.md):
  Lifecycle Owner создаёт Launch Attempts и владеет per-Instance orchestration;
- [ARCH-005](../architecture/ARCH-005-runtime-configuration-snapshot-and-loading-model.md):
  Owner pin exact version до Loader и Builder и передаёт полный operational
  provenance;
- [DP-007](DP-007-configuration-loader-contract.md): Loader принимает neutral
  exact-version `LoadRequest`;
- [DP-008](DP-008-snapshot-builder-contract.md): Builder создаёт Snapshot с
  matching five-part provenance;
- [DP-009](DP-009-runtime-bootstrap-contract.md): `runtime.Launch` является
  единственной stateless launch boundary.

Источник более высокого статуса имеет приоритет над этим Draft.

## 4. Область и boundary

DP-010 определяет:

- точные exported local declarations и sentinel errors;
- construction и immutable identity binding;
- two-phase `PrepareStart`, затем `Start`;
- Owner-issued attempt identity и exact version pin;
- semantics preparation, launch, Start, Stop и observation;
- одну synchronization boundary и linearization points;
- caller cancellation и same-token convergence;
- truthful failure retention;
- local proof requirements и будущие integration gates.

Он не реализует и не определяет adapter, вызывающий Loader или Builder.

## 5. Package и ответственность

Package — `internal/runtimelifecycle`, а не `internal/runtime`.

Один `Owner` постоянно bound ровно к одному Workspace, Configuration и
`runtimeconfigload.RuntimeInstanceID`. Он не является Host, repository, generic
manager, registry, service locator или policy engine.

Разные Owners не разделяют lifecycle state и могут продвигаться независимо.

## 6. Точные exported declarations

Base implementation использует declarations ниже. TASK-039 спроектировала
expected-attempt declarations, а TASK-040 реализует их без добавления другой
public lifecycle abstraction:

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

type StartRequest struct { /* unexported identities */ }

func NewStartRequest(workspaceID, configurationID, configurationVersionID uint64) StartRequest
func (r StartRequest) WorkspaceID() uint64
func (r StartRequest) ConfigurationID() uint64
func (r StartRequest) ConfigurationVersionID() uint64

type LaunchPreparation struct { /* opaque owner/claim token, immutable LoadRequest, read-only context */ }

func (p LaunchPreparation) LoadRequest() runtimeconfigload.LoadRequest
func (p LaunchPreparation) Context() context.Context

type PreparationResultKind string

const (
    PreparationSnapshot PreparationResultKind = "snapshot"
    PreparationFailure  PreparationResultKind = "failure"
)

type PreparationResult struct { /* closed union, fields unexported */ }

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
    StopStopped         StopOutcomeKind = "stopped"
    StopFailed          StopOutcomeKind = "stop-failed"
    StopAttemptMismatch StopOutcomeKind = "attempt-mismatch"
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

type AttemptFact struct { /* immutable identities/facts only; no Host/raw causes */ }

func (f AttemptFact) WorkspaceID() uint64
func (f AttemptFact) ConfigurationID() uint64
func (f AttemptFact) ConfigurationVersionID() uint64
func (f AttemptFact) RuntimeInstanceID() runtimeconfigload.RuntimeInstanceID
func (f AttemptFact) LaunchAttemptID() runtimeconfigload.LaunchAttemptID
func (f AttemptFact) Phase() AttemptPhase
func (f AttemptFact) StopOrigin() StopOrigin
func (f AttemptFact) RunningPublished() bool
func (f AttemptFact) TerminalKind() AttemptTerminalKind

type Observation struct { /* immutable coherent snapshot */ }

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
func (o *Owner) PrepareStart(
    request StartRequest,
) (LaunchPreparation, error)
func (o *Owner) Start(ctx context.Context, preparation LaunchPreparation, result PreparationResult) (StartOutcome, error)
func (o *Owner) Stop(ctx context.Context) (StopOutcome, error)
func (o *Owner) StopExpectedAttempt(ctx context.Context, expectedAttemptID runtimeconfigload.LaunchAttemptID) (StopOutcome, error)
func (o *Owner) Observe() Observation
```

Exported Launcher interface, Loader adapter, Builder adapter, repository или
generic manager не входят в package.

## 7. Sentinel errors

Package раскрывает stable sentinels с точными именами:

```go
var (
    ErrInvalidOwner             error
    ErrInvalidStartRequest      error
    ErrAttemptIDSourceFailed    error
    ErrAttemptIDReused          error
    ErrStartConflict            error
    ErrPreparationNotOwned      error
    ErrInvalidPreparationResult error
    ErrInvalidExpectedAttempt   error
)
```

Callers различают их через `errors.Is`. Ошибка source attempt ID оборачивает и
`ErrAttemptIDSourceFailed`, и exact source error, чтобы оба оставались
discoverable. Validation и conflict errors не изменяют lifecycle state.
`ErrInvalidExpectedAttempt` отклоняет empty identity ожидаемого Launch
Attempt. Вызов реализованного `StopExpectedAttempt` для nil Owner возвращает
`ErrInvalidOwner`. Ни один validation outcome не изменяет lifecycle state.

## 8. Construction

`NewOwner` отклоняет с `ErrInvalidOwner`:

- zero Workspace ID;
- zero Configuration ID;
- empty Runtime Instance ID;
- nil `LaunchAttemptIDSource`;
- nil pointer `DependencyBindings`.

Owner заимствует один stable bindings envelope на весь lifetime и не изменяет
и не закрывает его. Bootstrap DP-009 сохраняет typed-nil и semantic dependency
validation.

Начальные desired и actual states — `DesiredStopped` и `ActualStopped`.
Construction не выполняет ID allocation, loading, building, launch или
resource acquisition.

## 9. Start request

`StartRequest` содержит ровно identities Workspace, Configuration и
ConfigurationVersion. Он не содержит Launch Attempt ID, Snapshot, Host,
capability Loader или Builder.

`PrepareStart` отклоняет zero request identities или Workspace/Configuration,
отличные от Owner, с `ErrInvalidStartRequest`, без allocation или mutation
state.

Identity version закрепляется в attempt, созданном Owner. Последующая
publication не перенаправляет эту preparation.

## 10. Allocation Attempt ID

Owner вызывает `LaunchAttemptIDSource` вне state mutex. Source только выделяет
opaque candidate value; он не создаёт Launch Attempt.

- Source error возвращает error, оборачивающую `ErrAttemptIDSourceFailed` и
  exact source error.
- Empty candidate возвращает `ErrAttemptIDSourceFailed`.
- Candidate, уже committed во время lifetime Owner, возвращает
  `ErrAttemptIDReused`.
- Ни один из этих outcomes не создаёт attempt и не изменяет lifecycle state.

Concurrent `PrepareStart` могут выделить несколько candidates вне mutex.
Candidates, проигравшие последующий claim, остаются unused. Launch Attempt
существует только после commit одного candidate со стороны Owner в claim
linearization point.

## 11. Claim PrepareStart

После validation request и allocation candidate `PrepareStart` блокирует
Owner. Claim допустим только из:

- `ActualStopped`; или
- `ActualFailed` без retained Host и active attempt.

Любая active preparation, launch, Host, stop или retained failed-cleanup Host
возвращает `ErrStartConflict` без claim.

Claim linearization point является одной locked mutation, которая:

1. создаёт Launch Attempt, owned Owner;
2. commits candidate ID в lifetime used-ID set;
3. pins identities Workspace, Configuration, exact ConfigurationVersion,
   Runtime Instance и Launch Attempt;
4. создаёт exact neutral `runtimeconfigload.LoadRequest`;
5. создаёт Owner-owned preparation/start context;
6. публикует desired `DesiredRunning` и actual `ActualStarting`;
7. устанавливает phase `AttemptPreparing`;
8. создаёт один opaque non-forgeable `LaunchPreparation`.

`LoadRequest()` является единственным Loader input этого contract.
`Context()` read-only и позволяет внешней preparation work наблюдать Stop.

## 12. Ownership Launch Preparation

`LaunchPreparation` является opaque token, bound ровно к одному claim Owner.
Его zero value, token другого Owner и forged или stale token возвращают
`ErrPreparationNotOwned`.

Token не раскрывает mutation, cancellation, Host, dependencies или operation
record. Он содержит exact `LoadRequest` и Owner-owned context через read-only
accessors.

Одна preparation принимает не более одного result. Её stored result и
operation остаются convergence target для каждого последующего Start с тем же
authentic token. Token не является durable idempotency key и не переживает
process restart.

## 13. Closed Preparation Result

`PreparationResult` содержит ровно одно:

- полный Snapshot, созданный через `PreparedSnapshot`; или
- exact preparation error, созданную через `FailedPreparation`.

`Kind()` сообщает `PreparationSnapshot` или `PreparationFailure`. Zero value,
nil failure, сочетание accessors, не соответствующее declared kind, или другой
malformed union возвращает `ErrInvalidPreparationResult`, а preparation
остаётся unaccepted.

До mutation Start prepared Snapshot должен совпасть со всеми пятью identities,
pinned в `LaunchPreparation.LoadRequest()`:

1. Workspace;
2. Configuration;
3. exact ConfigurationVersion;
4. Runtime Instance;
5. Launch Attempt.

Mismatch возвращает `ErrInvalidPreparationResult`, не accepts preparation и
выполняет zero Launcher calls. Semantics Loader и Builder не повторяются сверх
этой identity handoff check.

## 14. Acceptance Start

`Start` сначала authenticates preparation token этого Owner и claim. Zero,
unknown или other-Owner tokens возвращают `ErrPreparationNotOwned`.

Непосредственно перед acceptance или attachment под state mutex Start
проверяет `ctx.Err()`:

- non-nil возвращает exact context error без mutation, acceptance или
  attachment;
- nil и следующая acceptance или attachment под тем же lock выигрывают гонку
  с последующей cancellation.

Для live unaccepted preparation первый valid result, accepted под mutex,
выигрывает:

- failure result terminalizes preparation без Launcher call;
- Snapshot result переводит `AttemptPreparing -> AttemptLaunching`, создаёт
  одну tracked operation и после unlock планирует ровно один Launcher call.

Owner принимает этот exact Snapshot value или error interface identity как
единственный preparation result attempt. Tracked launch operation сохраняет
Snapshot только до возврата Launcher; Owner не сохраняет Snapshot в
historical attempt. Exact identity preparation error сохраняется в convergent
outcome. Owner не выполняет structural, pointer, deep, text, `errors.Is`,
chain, comparability или semantic-equivalence comparison между results.

После первой acceptance каждый последующий Start с тем же authentic token
полностью игнорирует переданный `PreparationResult`, включая zero или different
value, и attaches к stored operation или возвращает stored outcome. Поздние
arguments не могут изменить attempt и намеренно не проходят validation.

Все runtime calls и waits выполняются вне mutex.

## 15. Fixed Launcher call

Для accepted Snapshot result production path точен:

```go
runtime.Launch(&runtime.BootstrapRequest{
    Snapshot:       snapshot,
    StartupContext: ownerPreparationContext,
    Dependencies:   owner.dependencies,
})
```

Для attempt существует ровно один вызов `runtime.Launch`. Owner никогда не
вызывает напрямую `runtime.Bootstrap` или `Host.Start`.

Package-private immutable test seam может доказывать scheduling и results. Он
не должен становиться exported, mutable package state, registry или production
policy boundary.

## 16. Preparation failure

Accepted `FailedPreparation(cause)`:

- сохраняет exact error identity и cause chain;
- выполняет zero Launcher calls;
- помечает phase `AttemptHistorical`;
- устанавливает terminal kind `AttemptPreparationFailed`;
- очищает active attempt;
- публикует actual `ActualFailed`;
- оставляет desired `DesiredRunning`;
- завершает Start с kind `StartPreparationFailed`.

Exact cause доступна из convergent `StartOutcome`, а не `Observation`.

Если Stop уже terminalized preparation, последующий Start с тем же token
возвращает stored `StartStoppedBeforeRunning` без acceptance нового failure
или Launcher call.

## 17. Success и failure Launch

Launch failure без claim Stop:

- сохраняет exact `runtime.BootstrapOutcome`, failure identity и cause chain;
- не сохраняет Host;
- помечает attempt historical с `AttemptLaunchFailed`;
- публикует actual `ActualFailed`, а desired остаётся `DesiredRunning`;
- завершает Start с `StartLaunchFailed`.

Launch success публикует Host, phase `AttemptRunning`, actual `ActualRunning`,
`RunningPublished() == true` и один immutable `StartRunning` outcome только
если Stop не claimed attempt и desired остаётся `DesiredRunning`. Success
DP-009 уже доказывает readiness Host. После publication Running fact и stored
Start outcome никогда не регрессируют.

## 18. Claims Stop по phase

Stop всегда проверяет `ctx.Err()` под mutex непосредственно перед claim или
attachment. Non-nil error выигрывает без mutation или attachment. Nil check и
следующая locked mutation выигрывают у последующей caller cancellation.

| Phase attempt или state | Обязательное поведение Stop |
| --- | --- |
| Нет active attempt, `ActualStopped` | Возврат idempotent `StopStopped`. |
| `AttemptPreparing` | Установка `StopBeforeRunning`, сохранение Running unpublished, desired Stopped, cancellation Owner context после unlock, terminalization `AttemptStoppedBeforeRunning`, очистка active attempt, publication ActualStopped. |
| `AttemptLaunching` | Установка `StopBeforeRunning`, сохранение Running unpublished, desired Stopped и ActualStopping, phase AttemptStopping, cancellation context после unlock, attach к той же combined operation. |
| `AttemptRunning` | Установка `StopAfterRunning`, сохранение Running published, desired Stopped и ActualStopping, phase AttemptStopping, создание ровно одной tracked Host Stop operation. |
| `AttemptStopping` | Attach к существующей operation без создания второго shutdown owner. |
| `ActualFailed` без Host | Немедленная publication ActualStopped с сохранением LastAttempt. |
| `ActualFailed` с retained Host | Возврат stored `StopFailed` без retry cleanup. |

Concurrent operations следуют порядку mutex claims. `StopOrigin()` и
`RunningPublished()` записываются при claim Stop и никогда не регрессируют.

### 18.1 Расширение atomic expected-attempt Stop

`StopExpectedAttempt(ctx, expectedAttemptID)` — реализованная изолированная
atomic operation для будущего private orchestration caller, который должен
остановить один exact Owner-issued Launch Attempt. Expected identity должна
быть non-empty; nil Owner возвращает
`ErrInvalidOwner`, а empty identity — `ErrInvalidExpectedAttempt`, без
lifecycle mutation.

После validation operation под mutex Owner проверяет `ctx.Err()`, затем
выбирает один relevant attempt: active attempt, если он существует, иначе
retained last attempt, если он существует, иначе none. Active attempt всегда
имеет приоритет над last attempt. Поэтому old last attempt A не может совпасть,
пока существует active successor B.

Если relevant identity отсутствует или отличается от expected identity,
operation возвращает `StopAttemptMismatch` с nil error. Это valid negative
outcome, а не conflict или lifecycle failure. Он выполняет zero mutation,
attachment, cancellation, Host call и wait. `StopOutcome.Attempt()` раскрывает
captured relevant immutable fact, если он существует, и отсутствует при none;
`Failure()` отсутствует.

Когда совпадает exact active attempt, operation использует ordinary Stop state
machine section 18 без phase-specific divergence:

- Preparing terminalizes как `AttemptStoppedBeforeRunning` и планирует
  cancellation Owner context только после unlock;
- Launching claims ту же combined operation, планирует cancellation после
  unlock и ждёт её exact result;
- Running claims ровно одну Host Stop operation для этого attempt;
- Stopping только attaches к уже tracked operation;
- retained active `AttemptStopFailed` возвращает exact stored `StopFailed`
  outcome без retry.

Когда active attempt отсутствует, а retained last attempt совпадает:

- `AttemptStopped` и `AttemptStoppedBeforeRunning` replay attempt-specific
  outcome `StopStopped` без mutation;
- `AttemptPreparationFailed` и `AttemptLaunchFailed` используют тот же
  существующий resource-free transition Failed-to-Stopped, что и ordinary
  Stop: desired становится `DesiredStopped`, actual становится
  `ActualStopped`, exact last attempt сохраняется, а operation возвращает
  attempt-specific `StopStopped`;
- любое impossible matched state возвращает `ErrStartConflict` без mutation.

Match, claim или attachment является cancellation linearization point. Context
error, видимая при locked check, выигрывает без mutation или attachment. После
locked match и claim последующая caller cancellation прерывает только wait
этого caller; Owner-owned cancellation, launch convergence, Host Stop и cleanup
продолжаются. Callers с одинаковой identity сходятся на одном tracked или
retained outcome. Другая identity никогда не attaches к этой work.

Реализация направляет generic `Stop(ctx)` и
`StopExpectedAttempt` через один private helper ordinary Stop, чтобы их phase
behavior не расходился. Expected path нельзя реализовывать как `Observe()` с
последующим `Stop()`. Mutex Owner освобождается до context cancellation,
`runtime.Launch`, `Host.Stop`, callbacks, external storage, work Flow, I/O и
любого channel, context, resource или caller wait.

## 19. Stop во время preparation или launch

Stop в `AttemptPreparing` завершается без Host. Последующий Start с тем же
preparation token сходится на stored `StartStoppedBeforeRunning` и выполняет
zero Launch calls.

Stop в `AttemptLaunching` отменяет Owner-owned context вне mutex. Если Launch
позже возвращает Host:

- Running никогда не публикуется;
- Owner сохраняет Host в `AttemptStopping`;
- `Host.Stop` вызывается ровно один раз.

Если Launch возвращает failure, exact Launch outcome сохраняется internal как
secondary attempt fact, а Start kind равен `StartStoppedBeforeRunning`.
`LaunchOutcome()` остаётся false, поскольку primary Start kind не равен
`StartLaunchFailed`; Host Stop не требуется.

При success или failure late Host Stop before-Running Start остаётся
`StartStoppedBeforeRunning`. Stop failure сообщает только `StopOutcome`.

## 20. Host Stop и terminal proof

`Host.Stop` получает Owner-owned non-caller `context.Background()` и
выполняется вне mutex. DP-010 не определяет timeout, force или retry policy.

Пока он блокируется, actual state остаётся `ActualStopping`.

Nil является единственным proof, позволяющим Owner:

- очистить references Host и active attempt;
- пометить attempt historical с `AttemptStopped`;
- опубликовать actual `ActualStopped`;
- завершить Stop с `StopStopped`.

Non-nil error не доказывает отсутствие ресурсов:

- actual становится `ActualFailed`;
- desired остаётся `DesiredStopped`;
- retained active attempt остаётся в `AttemptStopping` с terminal kind
  `AttemptStopFailed`;
- exact error identity и cause chain сохраняются;
- references Host и attempt сохраняются;
- Stop завершается с `StopFailed`;
- repeated Stop возвращает stored failure без retry;
- новый `PrepareStart` возвращает `ErrStartConflict`.

Для `StopBeforeRunning` Start остаётся `StartStoppedBeforeRunning` после success
или failure Stop. Для `StopAfterRunning` уже stored `StartRunning` остаётся
неизменным во время Stopping и после любого Stop result. Stop никогда
retroactively не изменяет факт publication Running.

## 21. Same-token convergence

Для exact preparation token:

- в `AttemptPreparing` может быть принят первый valid Start result;
- после первой acceptance result каждый repeated Start игнорирует supplied
  result без comparison или validation;
- в `AttemptLaunching` repeated Start attach к одной launch operation;
- в `AttemptRunning` repeated Start возвращает stored `StartRunning`;
- в `AttemptStopping` с `StopBeforeRunning` repeated Start attach к combined
  operation и возвращает `StartStoppedBeforeRunning` после любого Stop result;
- в `AttemptStopping` с `StopAfterRunning` repeated Start возвращает immutable
  stored `StartRunning` во время и после Stop;
- historical preparation или launch failure возвращает его stored exact
  `StartOutcome`;
- preparation, остановленная до Start, возвращает stored
  `StartStoppedBeforeRunning`.

Foreign или forged token возвращает `ErrPreparationNotOwned`. Equality или
equivalence comparison поздних Snapshots или errors не выполняется.
Convergence никогда не вызывает Launcher или Host Stop дважды.

## 22. Caller cancellation

`PrepareStart` не имеет caller context.

Для Start и Stop locked check `ctx.Err()` непосредственно перед acceptance,
claim или attachment определяет гонку:

- cancellation, уже видимая на check, возвращает context error и ничего не
  изменяет;
- nil check вместе с непосредственно следующей locked mutation выигрывает;
- cancellation после этой точки прерывает только ожидание этого caller.

Owner-owned tracked work, context и cleanup duties продолжаются после возврата
waiter. Callers могут повторить call с тем же token или использовать `Observe`
для получения истины. Caller cancellation не является terminal lifecycle
outcome.

## 23. Outcomes и access

`StartOutcome` immutable и имеет ровно один объявленный `StartOutcomeKind`. Он
всегда раскрывает `AttemptFact`.

- `PreparationFailure()` успешен только для
  `StartPreparationFailed`.
- `LaunchOutcome()` успешен только для `StartLaunchFailed`.
- `StartRunning` и `StartStoppedBeforeRunning` не раскрывают failure accessor.
- Stop failure никогда не дублируется как raw Start failure.

`StopOutcome` immutable и имеет ровно один объявленный `StopOutcomeKind`.
`Attempt()` отсутствует только для idempotent Stop без applicable attempt или
outcome `StopAttemptMismatch`, когда relevant attempt отсутствует.
`Failure()` успешен только для `StopFailed` и возвращает exact Host Stop error.

Для expected-attempt extension `StopAttemptMismatch` также
является declared immutable kind. Его `Attempt()` сообщает relevant fact, если
он существовал, а `Failure()` всегда отсутствует. Он возвращается с nil
method-level error и не представляет accepted lifecycle mutation.

Method-level errors представляют validation, conflict, ownership, ID-source
или wait этого caller. Они не заменяют accepted lifecycle outcomes.

## 24. Observation

`Observe` возвращает одно detached immutable coherent value, построенное под
state mutex.

Оно раскрывает identities Workspace, Configuration и Runtime Instance Owner;
desired и actual state; optional active `AttemptFact`; optional last
`AttemptFact`. Attempt facts включают immutable Stop origin, был ли Running
когда-либо published, и terminal category.

`Observation` и `AttemptFact` не раскрывают Host, dependency, context,
cancellation function, mutable operation или raw cause. Exact failures доступны
только через convergent operation outcomes до отдельного contract diagnostics
и redaction.

## 25. Owned state и locking

Один short-held mutex защищает:

- bound identities и desired/actual state;
- used attempt IDs;
- active и last attempt facts;
- live opaque preparation token и consumption fact;
- tracked launch и stop operations;
- Owner-owned preparation context и cancellation;
- active или conservatively retained Host;
- exact terminal outcomes для convergence.

Mutex никогда не удерживается во время ID-source calls, Loader/Builder work,
`runtime.Launch`, `Host.Stop`, resource/channel waits, context waits или caller
waits.

## 26. Ownership

| Значение или ресурс | Ownership |
| --- | --- |
| Scope Workspace, Configuration, Runtime Instance | Immutable bound к Owner при construction. |
| Identity и record Launch Attempt | Candidate выделяется injected source; attempt создаётся и owned Owner при claim. |
| Exact pin ConfigurationVersion | Выбирается StartRequest и pin Owner до Loader/Builder. |
| LoadRequest | Создаётся Owner и раскрывается read-only через preparation. |
| Snapshot или preparation failure | Создаётся снаружи и возвращается через closed PreparationResult. Snapshot сохраняется только tracked launch operation до возврата Launcher; exact failure сохраняется в convergent outcome. |
| Dependency bindings | Внешний owner; stable borrowed без изменений Owner. |
| Preparation/start context | Создаётся, отменяется и owned Owner. |
| Runtime Launcher | Stateless и не владеет lifecycle state. |
| Runtime Host reference | Owned Owner после Launch до доказанного cleanup. |
| Runtime graph и resources | Owned исключительно Host по ARCH-002. |

## 27. Ограничение truthfulness

Текущий Host не имеет completion или supervision signal для unexpected
termination после publication Running. Owner не должен выводить `Running ->
Failed` через polling `Running()` или предположение по unrelated facts.

Unexpected termination Runtime остаётся integration-gated до появления
Host-owned terminal signal.

## 28. Local acceptance proofs

Изолированная реализация должна доказать:

1. validation constructor, request, ID-source и identity;
2. Owner создаёт attempt и pin exact version до external preparation;
3. `LoadRequest` содержит пять exact identities;
4. concurrent preparation commits не более одного claim;
5. losing ID candidates никогда не становятся attempts;
6. foreign, stale и reused preparations используют declared errors;
7. invalid или mismatched preparation выполняет zero Launch calls и остаётся
   unaccepted;
8. первый valid concurrent result выигрывает, поздние same-token arguments
   игнорируются, а non-comparable Snapshots или errors не приводят к equality
   call или panic;
9. exact preparation и Launch failures сходятся unchanged;
10. accepted Snapshot приводит ровно к одному Launcher call, а Owner
    освобождает свою копию Snapshot после возврата Launcher;
11. Running публикуется только после возврата successful ready Host;
12. Stop работает в Preparing, Launching, Running, Stopping и обеих формах
    Failed;
13. before-Running Stop сохраняет `StartStoppedBeforeRunning` при success и
    failure Stop;
14. after-Running Stop сохраняет exact stored `StartRunning` во время и после
    success или failure Stop;
15. same-token Start сходится во всех phases без validation поздних result
    arguments;
16. concurrent Stop приводит к одному Host Stop и одному exact result;
17. locked гонка `ctx.Err()` имеет заданные winner semantics;
18. caller cancellation после claim влияет только на waiting;
19. Stop failure сохраняет Host и блокирует новую preparation;
20. resource call или wait отсутствует под mutex;
21. Observation coherent, capability-safe и сообщает immutable Stop origin и
    publication Running;
22. отдельные Owners продвигаются независимо;
23. package race tests и применимые Runtime regression tests проходят при
    поддержке toolchain.

### Acceptance proofs расширения expected-attempt

Изолированная реализация `StopExpectedAttempt` дополнительно доказывает:

1. validation nil Owner и empty expected ID использует declared sentinels и
   ничего не изменяет;
2. отсутствие relevant attempt и отличающийся relevant attempt возвращают
   `StopAttemptMismatch` с exact optional fact, nil failure и zero attachment,
   cancellation, Host call или wait;
3. выбор active attempt предшествует выбору retained last attempt, включая
   successor race old A/new active B, и никогда не останавливает B для expected
   A;
4. exact attempts Preparing, Launching, Running и Stopping сохраняют ordinary
   Stop phase behavior и exact outcomes;
5. callers с одинаковым ID сходятся, а callers с разными ID никогда не
   attaches;
6. retained active Stop failure сохраняет exact identity error и не выполняет
   cleanup retry;
7. retained stopped forms replay `StopStopped`, а matching historical failures
   preparation и launch выполняют resource-free transition Failed-to-Stopped
   для этого exact attempt;
8. cancellation, видимая при locked check, выигрывает без mutation, а
   cancellation после match или claim освобождает только этого waiter;
9. regression proofs generic `Stop` подтверждают неизменность semantics shared
   helper;
10. lock/lifetime, independent-Owner, race, vet, formatting и exported GoDoc
    checks проходят при поддержке toolchain.

## 29. Будущие integration proofs

Будущим production evidence остаются:

- Loader использует ровно `LaunchPreparation.LoadRequest`;
- Builder возвращает Snapshot provenance, matching той же preparation;
- каждый production path проходит Owner-to-`runtime.Launch` без обхода через
  Bootstrap или `Host.Start`;
- для каждого Runtime Instance маршрутизируется ровно один Owner scope;
- management authorization выполняется до mutation;
- dependency composition стабильна и не содержит hidden registry;
- durable allocation, history, idempotency и recovery переживают restart;
- terminal supervision Host позволяет truthful observation unexpected
  failure.

AP-003 и AP-011 DP-009 остаются integration-gated.

## 30. Явно отложено

DP-010 не определяет:

- actual Loader/Builder adapter или production wiring;
- persistence schema, transaction или durable history;
- HTTP API, authorization, command DTO или durable idempotency;
- concrete distributed или persistent ID-allocation strategy;
- retry, restart, replacement, rollback, recovery, reconciliation или reload;
- PID, worker, process supervision, scheduling или clustering;
- diagnostics, logging, redaction, metrics или alerting;
- timeout, force или cleanup-retry policy Stop;
- изменения Host API или polling unexpected termination;
- generic manager, registry, service locator или policy framework.

Он также не определяет public expected-attempt command DP-013, composition
invoker или orchestration policy. Они остаются отдельной работой после
изолированного расширения Owner.

## 31. Implementation boundary

Реализованный первый code slice добавляет только
`internal/runtimelifecycle` и local proof tests этого contract. Он использует
fakes вокруг package-private immutable launch seam и external preparation
boundary.

Declarations и semantics expected-attempt в sections 6, 7, 18.1, 23 и 28
реализованы и верифицированы изолированно завершённой и
Coordinator-Accepted TASK-040. Это не заявляет private invocation, integration
или production capability. Repeat final Reviewer verdict — `APPROVED`,
blocking/non-blocking findings 0/0.

Он не подключает Loader, Builder, HTTP handlers Control Service, repositories,
persistence или production routing. Implementation не повышает design status.

## 32. Итог

Runtime Lifecycle Owner сначала создаёт и pin один Launch Attempt через
`PrepareStart`, раскрывая exact neutral LoadRequest, требуемый Loader и
Builder. Затем `Start` accepts один closed preparation result и вызывает
только stateless Runtime Launcher.

Одна short-held state boundary сохраняет per-Instance serialization,
same-token convergence, caller-cancellation rules, conservative Host
ownership и truthful operational state без преждевременного проектирования
persistence, management, policy, supervision или production wiring.

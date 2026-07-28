# DP-011: Интеграция Runtime Launch Pipeline

[English version](../../en/design/DP-011-runtime-launch-pipeline-integration.md)

## 1. Статус

**Design Status:** Draft

**Implementation Status:** Planned

**Статус архитектуры:** сфокусированный integration contract поверх
утверждённых ARCH-004 и ARCH-005 и существующих Draft DP-007–DP-010.

Документ не объявляет production launch capability реализованной и не повышает
статусы связанных Draft DP.

## 2. Назначение

Определить минимальный in-process flow, который соединяет один уже созданный
Runtime Lifecycle Owner с Configuration Loader, Snapshot Builder и stateless
Runtime Launcher.

Flow должен:

- начинать Launch Attempt только через `Owner.PrepareStart`;
- передавать Loader ровно выданный Owner `LoadRequest`;
- передавать Builder ровно успешный `DetachedLoadResult`;
- преобразовывать Loader failure или Diagnostics Builder в закрытый
  `PreparationResult`;
- передавать результат обратно тому же Owner через `Owner.Start`; и
- сохранять ownership, cancellation и concurrency contracts DP-007–DP-010.

## 3. Источники полномочий

Контракт ограничен следующими источниками:

- [ADR-0002](../adr/0002-configuration-dsl.md): Published
  ConfigurationVersion является immutable declarative source;
- [ADR-0003](../adr/0003-runtime-architecture.md): Runtime dependencies явны,
  а Runtime не читает Control Plane repositories;
- [ARCH-002](../architecture/ARCH-002-runtime-foundation-freeze.md): Host
  владеет production composition, startup, rollback и shutdown;
- [ARCH-004](../architecture/ARCH-004-runtime-deployment-and-identity-model.md):
  Lifecycle Owner владеет Runtime Instance, Launch Attempt и lifecycle
  serialization;
- [ARCH-005](../architecture/ARCH-005-runtime-configuration-snapshot-and-loading-model.md):
  Launch Attempt владеет построением detached Snapshot;
- [DP-007](DP-007-configuration-loader-contract.md): Loader принимает exact
  five-identity `LoadRequest` и возвращает один detached result или failure;
- [DP-008](DP-008-snapshot-builder-contract.md): Builder возвращает один
  полный Snapshot или non-empty exhaustive Diagnostics;
- [DP-009](DP-009-runtime-bootstrap-contract.md): каждый launch проходит через
  stateless `runtime.Launch`;
- [DP-010](DP-010-runtime-lifecycle-owner-contract.md): Owner создаёт
  preparation, принимает closed result и единолично вызывает Launcher.

Источник более высокого статуса имеет приоритет над этим Draft.

## 4. Область

DP-011 определяет:

- один package-level integration boundary;
- immutable binding одного Flow к одному Owner и одному Loader;
- exact `PrepareStart -> Load -> Build -> Start` flow;
- representation Builder Diagnostics как preparation failure;
- synchronous attempt operation и caller cancellation gate;
- Stop/cancellation races;
- dependency direction, ownership и lifetime;
- local и future production activation proofs.

DP-011 не определяет:

- HTTP endpoint, management command DTO или authorization;
- repository adapter либо выбор concrete `configurationloader.Source`;
- persistence Runtime Instance, Launch Attempt, desired или actual state;
- durable idempotency, retry, restart, replacement или reconciliation;
- supervision, terminal Host signal, diagnostics transport или metrics;
- новые Loader, Builder, Launcher или Owner semantics.

## 5. Термины

### Runtime Launch Flow

Immutable in-process integration object, который заимствует один Owner и один
Loader и для каждого принятого Start request выполняет единственную
последовательность preparation.

### Start Operation

Ровно одна synchronous call-stack operation, связанная с одним invocation
`Flow.Start`. После успешного `PrepareStart` та же goroutine caller
последовательно выполняет Load, Build и передачу result Owner. Flow не создаёт
goroutine, detached work или background worker.

### Build Failure

Immutable error, содержащая полный non-empty набор blocking Diagnostics одного
Builder invocation. Она переводит result Builder в существующий
`FailedPreparation`, не превращая Diagnostics в Loader или Bootstrap failure.

### Production Activation

Подключение Flow к единственной авторизованной management boundary, concrete
Source composition и operational state storage. Создание package Flow само по
себе не является Production Activation.

## 6. Решение

Принимается один flow:

```text
authorized caller
    -> Runtime Launch Flow.Start
        -> Owner.PrepareStart
        -> Configuration Loader.Load
        -> Snapshot Builder.Build
        -> Owner.Start
            -> runtime.Launch
                -> Runtime Bootstrap
                    -> Runtime Host
```

Flow не создаёт Launch Attempt, Snapshot, Bootstrap request или Host
самостоятельно. Он только последовательно соединяет существующих owners этих
объектов.

## 7. Package и responsibility

Первый implementation slice располагается в
`internal/runtimelaunchflow`.

Package отвечает только за:

- binding одного Owner и одного configured Loader;
- вызов существующего stateless Builder;
- преобразование полного Diagnostics set в Build Failure;
- synchronous выполнение одной Start Operation;
- передачу closed preparation result исходному Owner.

Package не является Lifecycle Owner, source adapter, repository, generic
orchestrator, registry, supervisor или policy engine.

## 8. Точная первая public surface

Первая реализация ограничена следующим conceptual Go contract:

```go
package runtimelaunchflow

var (
    ErrInvalidFlow         error
    ErrInvalidStartContext error
)

type Flow struct {
    // package-private immutable dependencies
}

func New(
    owner *runtimelifecycle.Owner,
    loader *configurationloader.Loader,
) (*Flow, error)

func (f *Flow) Start(
    ctx context.Context,
    request runtimelifecycle.StartRequest,
) (runtimelifecycle.StartOutcome, error)

type BuildFailure struct {
    // package-private immutable diagnostics
}

func (f *BuildFailure) Error() string
func (f *BuildFailure) Diagnostics() []runtimeconfig.Diagnostic
```

Дополнительный exported interface, callback, registry, option, stage enum,
result union или lifecycle state не вводится.

`New` отклоняет nil Flow dependencies с `ErrInvalidFlow`. `Start` отклоняет
nil context с `ErrInvalidStartContext`. Ошибки Owner, Loader и context не
переклассифицируются.

## 9. Construction и dependency binding

`Flow` permanently bound к одному `*runtimelifecycle.Owner` и одному
`*configurationloader.Loader`.

Constructor:

- не создаёт Runtime Instance или Launch Attempt;
- не вызывает Loader, Builder, Launcher или Host;
- не читает repository;
- не создаёт goroutine;
- не копирует и не изменяет dependency state.

Builder создаётся как stateless `runtimeconfig.NewBuilder()` внутри package
или хранится как value без mutable state. Production constructor не принимает
заменяемый Builder или Launcher seam.

Concrete Source выбирается и передаётся при construction Loader за пределами
Flow. Поэтому Flow не знает transport, repository или deployment adapter.

## 10. Claim и startup request

`Flow.Start` до lifecycle mutation проверяет non-nil context и ровно один раз
читает `ctx.Err()`. Этот read является **Caller Cancellation Gate**
linearization point. Затем Flow вызывает:

```go
preparation, err := owner.PrepareStart(request)
```

Flow не создаёт и не исправляет identity. Validation Start request, allocation
Launch Attempt ID, conflict detection и claim linearization остаются только у
Owner.

Если cancellation конкурирует с Gate:

- cancellation, наблюдаемая Gate, выигрывает и запрещает claim;
- nil, наблюдаемый Gate, разрешает попытку claim, и последующая cancellation
  caller не отменяет эту Start Operation, даже если возникает до внутренней
  claim linearization Owner.

Если `PrepareStart` возвращает error, Flow:

- не начинает Load или Build;
- не вызывает Loader, Builder, Owner.Start или Launcher;
- возвращает exact error caller.

## 11. Synchronous operation и caller lifetime

После successful claim тот же `Flow.Start` invocation синхронно выполняет
Load, Build и `Owner.Start`. Flow не создаёт goroutine, channel, detached work
или package-owned operation state.

Caller context после победы Gate больше не управляет lifetime operation и не
проверяется повторно. `Flow.Start` возвращается только после одного outcome
Owner либо exact operation error. Это сохраняет одного явного владельца
blocking call stack и не оставляет unawaited preparation work.

Если configured Source блокируется, `Flow.Start` и goroutine caller остаются
заблокированными до возврата Source. Текущий контракт не обещает timeout или
force cancellation. Такая ограниченность должна быть видима Production
Activation и не может маскироваться detached worker.

## 12. Exact Loader handoff

Start Operation вызывает Loader ровно один раз и только с:

```go
preparation.LoadRequest()
```

Запрещено:

- строить второй `LoadRequest`;
- заменять любую из пяти identities;
- выбирать latest или replacement ConfigurationVersion;
- выполнять fallback, retry или второй Load;
- передавать Loader Owner, Snapshot dependencies или management authority.

При Loader success ровно полученный `DetachedLoadResult` передаётся Builder.
Flow не нормализует, не исправляет и не перечитывает source material.

## 13. Loader failure

При Loader error:

1. Builder не вызывается;
2. operation создаёт `runtimelifecycle.FailedPreparation` с exact error
   interface;
3. тот же Owner получает result через `Owner.Start`;
4. Launcher, Bootstrap и Host не вызываются.

Flow не оборачивает и не редактирует Loader error. Sentinel identity и cause
chain остаются доступными через `StartOutcome.PreparationFailure()`.

## 14. Exact Builder handoff

При Loader success operation вызывает один stateless Builder ровно один раз:

```go
snapshot, diagnostics := runtimeconfig.NewBuilder().Build(loadResult)
```

Flow не выполняет semantic validation, normalization или provenance
construction самостоятельно.

Результат обрабатывается только как closed contract DP-008:

- empty Diagnostics означает один complete Snapshot;
- non-empty Diagnostics означает отсутствие Snapshot и Build Failure.

Malformed output невозможен от concrete Builder и не создаёт отдельной
production policy boundary.

## 15. Build Failure

Для non-empty Diagnostics operation создаёт один `*BuildFailure`.

`BuildFailure`:

- сохраняет порядок полного Diagnostics set;
- владеет detached копией slice;
- возвращает detached copy из `Diagnostics()`;
- имеет exact constant `Error()` `Runtime Snapshot build failed` без
  concatenation locations, values или source material;
- не реализует retry, severity mapping, HTTP mapping, logging или redaction
  policy.

Exact pointer Build Failure передаётся в
`runtimelifecycle.FailedPreparation`. Builder Diagnostics остаются
Builder-owned semantic category и не становятся Loader, Bootstrap или Startup
failure.

## 16. Snapshot success

При успешном Build operation создаёт только:

```go
runtimelifecycle.PreparedSnapshot(snapshot)
```

и передаёт result тому же Owner с исходным opaque preparation token.

Owner повторно проверяет matching five-part provenance до acceptance. Flow не
обходит эту validation и не вызывает `runtime.Launch`, `runtime.Bootstrap` или
`Host.Start` напрямую.

После вызова `Owner.Start` local copies `DetachedLoadResult`, Snapshot и
Diagnostics не сохраняются Flow.

## 17. Stop и preparation cancellation

`LaunchPreparation.Context()` является единственным signal от Owner к
synchronous Start Operation. Flow не получает cancellation authority.

Operation проверяет этот context:

- перед Load;
- после возврата Load и до Build;
- после возврата Build и до передачи result Owner.

Текущий synchronous Loader contract не принимает context. Поэтому Stop не
обещает прервать уже выполняющийся `Source.LoadExact`; после его возврата
operation не начинает следующий stage, если Owner уже отменил preparation.
Timeout, force cancellation и изменение Source API не вводятся этим Draft.

Если Stop terminalized AttemptPreparing, operation не публикует второй failure
и не вызывает Launcher. После наблюдения cancelled preparation context
operation вызывает `Owner.Start` с исходным token, zero result и внутренним
non-cancelled wait context. По same-token convergence DP-010 поздний result не
валидируется, и operation получает stored `StartStoppedBeforeRunning`, не
изменяя terminal fact Stop.

Если Stop конкурирует с `Owner.Start`, winner semantics, Host cleanup и
`StartStoppedBeforeRunning` остаются ровно DP-010.

## 18. Owner.Start wait context

Caller context не передаётся как preparation, Owner wait или Runtime startup
authority после победы Caller Cancellation Gate. Для передачи результата
operation использует non-caller-cancelled wait context; Owner-owned
`LaunchPreparation.Context()` по-прежнему становится
`BootstrapRequest.StartupContext` внутри DP-010.

Это разделение необходимо, чтобы cancellation caller после Gate не оставляла
AttemptPreparing без terminal outcome. Оно не скрывает Stop: Stop меняет
Owner state и отменяет preparation context, а Owner convergence определяет
итог.

Flow не создаёт новый root Runtime context и не получает `CancelFunc` Host.

## 19. Concurrency и linearization

Один `Flow` может получать concurrent `Start` calls. Он не сериализует их
собственным mutex:

- Owner остаётся единственной per-Instance serialization boundary;
- не более одного `PrepareStart` claim становится active;
- losing calls не начинают Load или Build;
- один successful claim выполняет ровно одну synchronous Start Operation;
- одна operation выполняет не более одного Load, Build и Owner.Start;
- только Owner может вызвать Launcher;
- concurrent Flows, ошибочно bound к одному Owner, всё равно сходятся через
  Owner, но production composition обязана создавать ровно один Flow на Owner.

Разные Owner/Flow pairs не разделяют mutable state и могут выполняться
независимо.

## 20. Ownership и lifetime

| Объект | Owner |
| --- | --- |
| Runtime Instance и Launch Attempt | Runtime Lifecycle Owner |
| Configured source access | Configuration Loader |
| Load operation/source material до detachment | Loader |
| Detached Load Result между Load и Build | Synchronous Start Operation |
| Diagnostics до Build Failure | Synchronous Start Operation |
| Build Failure после acceptance | Owner outcome |
| Snapshot до acceptance | Synchronous Start Operation |
| Accepted Snapshot и launch operation | Runtime Lifecycle Owner |
| Bootstrap request и construction inputs | Bootstrap на время вызова |
| Runtime resources | Runtime Host |
| Caller wait | Один `Flow.Start` invocation |

Flow не владеет Host reference, lifecycle state, repository transaction,
Snapshot cache или retry state.

## 21. Dependency rules

Разрешённое направление production dependencies:

```text
management composition
    -> runtimelaunchflow
        -> runtimelifecycle
        -> configurationloader
        -> runtimeconfig

runtimelifecycle
    -> runtimeconfigload
    -> runtime
```

Запрещены reverse imports из `runtime`, `runtimeconfig`,
`configurationloader` или `runtimelifecycle` в `runtimelaunchflow`.

Flow не передаётся Host, Bootstrap, Builder или Loader как arbitrary
capability. Registry, global singleton, reflection и service locator
запрещены.

## 22. Failure matrix

| Stage | Observed result | Subsequent work |
| --- | --- | --- |
| Invalid context или cancellation, наблюдаемая Caller Cancellation Gate | `ErrInvalidStartContext` или exact context error | no claim, Load, Build, Start или Launch |
| Nil, наблюдаемый Gate; cancellation до или после Owner claim | cancellation не переигрывает Gate | synchronous operation продолжается до Owner outcome |
| Owner claim failure | exact Owner error | no Load, Build, Start или Launch |
| Loader failure | `StartPreparationFailed` с exact Loader error | no Build или Launch |
| Builder Diagnostics | `StartPreparationFailed` с `*BuildFailure` | no Launch |
| Snapshot provenance mismatch | exact `ErrInvalidPreparationResult` Owner | no Launch; blocking contract defect |
| Bootstrap/Startup failure | `StartLaunchFailed` с unchanged `BootstrapOutcome` | Owner records launch failure |
| Runtime success | `StartRunning` | Owner owns Host reference |
| Stop во время preparation | DP-010 `StartStoppedBeforeRunning` convergence | no new stage after observed preparation cancellation |

Flow не классифицирует management, authorization, persistence или recovery
failures.

## 23. Acceptance proofs

Первая implementation task должна доказать:

1. constructor и context validation выполняют zero lifecycle mutation;
2. ровно `LaunchPreparation.LoadRequest()` достигает Loader;
3. Loader failure вызывает zero Build и zero Launch;
4. Loader success передаётся Builder unchanged;
5. Builder вызывается не более одного раза;
6. полный Diagnostics set доступен через immutable `BuildFailure`;
7. Builder failure вызывает zero Launch;
8. Snapshot достигает того же Owner и проходит five-identity validation;
9. один accepted Snapshot приводит ровно к одному
   Owner-to-`runtime.Launch` path;
10. production flow не импортирует и не вызывает `runtime.Bootstrap` или
    `Host.Start`;
11. Caller Cancellation Gate имеет exact winner semantics, а cancellation
    после победы Gate не прерывает synchronous operation;
12. Stop до, во время и после Load/Build не создаёт второй result, operation
    или Launcher call;
13. concurrent Start одного Flow начинает не более одной Start Operation;
14. разные Owner/Flow pairs продвигаются независимо;
15. Flow, Loader, Builder и Launcher не разделяют package-global mutable state;
16. no resource call или wait выполняется под новым Flow mutex, поскольку
    такого mutex нет;
17. package и full repository tests проходят, включая race tests при
    поддержке toolchain.

Эти proofs закрывают integration часть AP-003 и AP-011 DP-009 только для
введённого in-process Flow. Полная Production Activation дополнительно требует
доказать, что единственная management start boundary использует этот Flow и
что обходных production paths нет.

## 24. Production Activation gates

До объявления capability production-integrated отдельные tasks должны
определить и проверить:

- concrete Source composition;
- единственную management command boundary и authorization до mutation;
- routing ровно одного Owner/Flow scope на Runtime Instance;
- persistence Runtime Instance, Launch Attempt, desired/actual facts и
  idempotency согласно ARCH-004;
- recovery/reconciliation либо явный startup rejection при отсутствии этих
  contracts;
- operational reporting и redaction preparation/launch failures.

DP-011 не выбирает порядок или API этих tasks. Planned package implementation
может быть отдельно проверяемым prerequisite, но сама по себе не меняет
`spec/current-state.md` на «управление Runtime из Control Service реализовано».

## 25. Намеренно отложено

Отложены:

- Source adapter in-memory repositories, PostgreSQL, YAML или remote transport;
- HTTP/CLI/API surface и authorization;
- persistence schema и transactions;
- durable command/result identity;
- retry, backoff, restart, replacement, rollback policy и reconciliation;
- terminal supervision Host и unexpected failure;
- timeout/force policy для blocking Source;
- diagnostics transport, metrics, audit и redaction policy;
- process isolation, scheduling, clustering и federation.

Ни один из этих вопросов не может быть реализован как hidden behavior Flow.

## 26. Implementation boundary

Следующий допустимый code slice после принятия этого Draft ограничен:

- `internal/runtimelaunchflow`;
- local proof tests;
- factual documentation synchronization.

Он не изменяет DP-007–DP-010 packages, Control Service, repositories,
management API, persistence, Host или Bootstrap.

Design review этого Draft предшествует любой implementation. Implementation не
повышает Design Status автоматически.

## 27. Последствия

Положительные:

- закрывается единственный явный gap между реализованными isolated contracts;
- Owner остаётся lifecycle authority;
- Loader и Builder responsibilities не смешиваются;
- AP-003/AP-011 получают конкретную integration proof boundary;
- caller cancellation имеет одну явную Gate linearization и не оставляет
  claimed attempt без synchronous owner duty;
- production activation gates остаются видимыми.

Издержки:

- caller после победы Gate может ждать blocking Source без cancellation;
- synchronous Source не может быть force-cancelled, а detached workaround
  запрещён;
- package implementation ещё не создаёт пользовательскую management
  capability;
- persistence и authorization по-прежнему требуют отдельных designs.

## 28. Итог

Runtime Launch Flow является узким immutable integration boundary. Он
последовательно использует Owner-issued preparation, exact Loader handoff,
stateless Builder и Owner acceptance, после чего только Owner вызывает
stateless Runtime Launcher.

Flow не становится вторым lifecycle owner и не скрывает product, persistence
или recovery policy. До отдельной Production Activation repository сохраняет
честный статус: integration contract определён, но управление Runtime из
Control Service не реализовано.

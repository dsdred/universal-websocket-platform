# DP-020: Готовность последовательности связывания оркестрации Runtime

[English version](../../en/design/DP-020-runtime-orchestration-binding-sequence-readiness.md)

## 1. Статус

- **Статус проектирования:** Draft
- **Статус реализации:** Planned overall; Срез 3 реализован и независимо
  принят изолированно

Прогресс реализации: TASK-031 и TASK-032 создали изолированные частичные
реализации Срезов 1 и 2, исторически принятые Coordinator, а TASK-034 определила
обязательный repair соответствия. TASK-035 реализует Срез 2R изолированно:
six-field authorization и полный binding теперь принадлежат dependency-leaf
package, primitive managed claims используют sole `ExecuteManagedStart`
adapter, а command-owned rendezvous identities уникальны и callback-scoped.
Concrete DP-013 composition-private invoker и production wiring отсутствуют.
TASK-036 устраняет оставшуюся неоднозначность command-gate и continuation API
Среза 3, а TASK-037 реализует этот протокол Среза 3 изолированно: managed
primitive и linked gates, stateless continuation OwnerClaim-to-DP-014, exact
threading revision и адаптацию outcomes managed Flow. TASK-037 независимо
принята. Срез 4 завершён и Coordinator Accepted как TASK-038 с verdict
`TASK-026 REMAINS BLOCKED`; его первый design-only следующий candidate не
активировался автоматически, а затем завершён как Coordinator-Accepted
TASK-039. Завершённая и Coordinator-Accepted TASK-040 реализует и верифицирует
принятые semantics Draft DP-010 изолированно; repeat final Reviewer
`APPROVED` 0/0. Общий статус остаётся Planned.

Этот focused design разделяет оставшиеся prerequisites Approved DP-019 — точную
авторизацию оркестрации, private managed invocation и связывание
OwnerClaim-to-DP-014 — на упорядоченные, независимо проверяемые
implementation-срезы и закрывает отложенные решения проектирования, которые им
нужны.

Он не определяет production-код и не меняет ни одну принятую семантику.
TASK-026 остаётся `Blocked by Architecture`.

## 2. Назначение

Approved DP-019 фиксирует инварианты ordering, ownership, concurrency, permit и
fail-closed для prerequisites оркестрации активации, но описывает tuple
авторизации, private managed invoker, managed Flow seam и последовательность
связывания только концептуально. Этот proposal фиксирует focused,
implementable разложение именно этих seam, чтобы каждый мог быть реализован и
независимо принят до повторного рассмотрения оркестратора DP-016.

## 3. Источники полномочий

Этот proposal уточняет, не переопределяя:

- [ARCH-004](../architecture/ARCH-004-runtime-deployment-and-identity-model.md),
  section 19;
- [DP-010](DP-010-runtime-lifecycle-owner-contract.md) — Owner остаётся
  единственным lifecycle authority и claimant попытки запуска;
- [DP-011](DP-011-runtime-launch-pipeline-integration.md) — синхронный путь
  `PrepareStart -> Load -> Build -> Start` и место private continuation;
- [DP-013](DP-013-runtime-management-routing.md) — exact routing,
  authorization-before-mutation и planned private seam;
- [DP-014](DP-014-runtime-operational-identity-persistence.md) — conditional
  aggregate revision и граница membership Launch Attempt / связывания
  execution-generation;
- [DP-015](DP-015-runtime-management-command-idempotency.md) — граница claim /
  replay / permit / unresolved-barrier, включая parent/phase core TASK-028 и
  Continue gate и rendezvous pending-Stop TASK-029;
- [DP-016](DP-016-runtime-activation-replacement-rollback.md) — неизменный
  ordering activation / replacement / rollback и acceptance proofs;
- [DP-017](DP-017-runtime-recovery-reconciliation.md) — fail-closed recovery;
- [DP-019](DP-019-runtime-activation-orchestration-prerequisites.md) — Approved
  prerequisite contract, который этот proposal разбивает.

Accepted ADR и Active/Frozen architecture остаются authoritative. Этот Draft
никогда не переопределяет Approved-документ и не повышает и не понижает ни один
существующий Design или Implementation Status.

## 4. Область действия

Дизайн покрывает только focused разложение концептуальных seam Approved DP-019
§7–§8 и §14–§16:

- representation и ownership policy-neutral orchestration authorizer, validated
  значение authorization request, exact failed-authorization error contract и
  адаптация DP-013 authorization-before-mutation;
- package split и направление invocation private managed invoker против
  managed Flow seam, managed construction и per-call managed Start surfaces,
  immutable per-invocation lifecycle `StartExecutionBinding` и exact
  failed private-invocation error contract;
- immutable tuple `OwnerClaimView` и последовательность связывания
  OwnerClaim-to-DP-014: composition inputs, closed четырёх-outcome результат
  continuation, порядок conditional публикации membership попытки и
  same-generation binding относительно существующего rendezvous pending-Stop,
  и размещение final gate Stop-versus-Continue;
- доказуемый инвариант адаптации активации;
- план доказательства lock/ownership, требуемый перед реализацией;
- упорядоченный список bounded implementation-срезов и порядок их зависимостей.

## 5. Не-цели

Дизайн не определяет:

- production-реализацию любого среза, public route HTTP/CLI/DTO/API или wiring;
- concrete authorization policy, модель Principal, отображение credentials или
  transport;
- storage schema, migration, external persistence, cross-process lease или
  выполнение recovery;
- реализацию оркестратора DP-016, реализацию recovery DP-017, реализацию
  reporting DP-018, automatic rollback/restart, supervision, scheduling или
  Production Activation;
- любое изменение контракта lifecycle DP-010, базового поведения Flow DP-011,
  public surfaces DP-013/DP-015 или принятых семантик
  DP-014/DP-015/DP-016/DP-017/DP-019;
- любую новую концепцию domain Runtime Instance или domain развёртывания.

## 6. Краткое решение

Оставшиеся prerequisites DP-019 фиксируются как три исходных упорядоченных
implementation-среза, один repair-срез соответствия перед Срезом 3 и одна
gated переоценка:

1. exact поверхность orchestration authorizer на существующей границе команд
   DP-013 / DP-019;
2. private managed invoker и managed Flow per-call seam с per-invocation
   `StartExecutionBinding`;
3. Срез 2R — repair соответствия authoritative binding и primitive managed
   adapter;
4. conditional публикация попытки OwnerClaim-to-DP-014 и same-generation binding
   до Load, интегрированные с существующим rendezvous pending-Stop;
5. переоценка готовности оркестратора DP-016 после того, как Срезы 1, 2, 2R и
   3 реализованы и независимо приняты.

Ни один срез не обходит другой. Срез 1 может существовать независимо; Срез 2
зависит от Среза 1; Срез 2R исправляет принятые partial Срезы 1 и 2; Срез 3
зависит от Среза 2R; Срез 4 зависит от независимой приёмки Среза 3.

## 7. Отложенное решение: Представления Orchestration Authorizer и Request

### 7.1 Значение authorization request

Одно immutable, validated значение является единственным входом авторизации:

```text
OrchestrationAuthorizationRequest {
    OperationalDomain           string
    WorkspaceID                  uint64
    ConfigurationID              uint64
    RuntimeInstanceID            runtimeconfigload.RuntimeInstanceID
    Action                       OrchestrationAction
    TargetConfigurationVersionID uint64
}
```

Validation требует, чтобы каждая identity была ненулевой и exact, а целевая
версия была действительной Published identity, принадлежащей этой Configuration.
Validation failure, denial/failure/panic авторизации, absent scope или pre-claim
cancellation выполняют нулевую mutation command, aggregate и lifecycle.

### 7.2 Сохранение поля `OperationalDomain`

Approved DP-019 §7 требует `OperationalDomain`; этот Draft не может его
удалить. Значение является exact непустым opaque domain, уже присутствующим в
принятом command `Scope` DP-015. Оно переносится в каждый initial и replay
authorization request и в тот же per-invocation binding. Оно не выводится из
Workspace, Configuration, Runtime Instance, process locality или default
константы. `runtimemanagement.Target` остаётся неизменным; private managed
invoker получает domain как отдельный immutable composition input и валидирует
его вместе с Target.

### 7.3 Представление authorizer

Seam — это один policy-neutral synchronous named function type, принадлежащий
будущему composition-facing orchestration package (не хранящийся в
`runtimemanagement.Directory` и не заменяющий существующий параметр DP-015
`Authorize(...) error`). Он валидируется при каждой submission, никогда не
кэшируется и повторно вычисляется при каждой initial и replay submission.

Соответствующий authorizer возвращает либо отсутствие ошибки, либо одну exact
ошибку. Denial, unavailability, invalid scope, panic и cancellation каждый
производят truthful, specific исход ошибки и нулевую mutation. Существующая
exported поверхность авторизации DP-013 для `Directory.Start/Stop/Observe`
остаётся неизменной; orchestration actions никогда не проходят через эту
поверхность без дополнительной exact orchestration authorization.

## 8. Закрытая Design Boundary: Private Managed Invoker и Managed Flow Seam

TASK-042 закрывает только оставшуюся design ambiguity concrete invoker через
Draft/Planned [DP-021](DP-021-private-exact-scope-managed-start-invoker.md).
Следующее существующее decomposition остаётся authoritative context: DP-021
фиксирует ownership `runtimemanagement`, custody preconstructed Flow, единственную
operation `InvokeManagedStart`, cancellation delegation, capability custody,
failure behavior и отсутствие legacy fallback. Future orchestrator-owned
callback closure DP-015 задачи TASK-026 вызывает этот invoker как sole
lifecycle subcall и владеет mapping `TerminalOutcome`, publication и
terminalization вне DP-021; сам invoker не является callback. Implementation
или orchestrator не активируется.

### 8.1 Package split и направление invocation

- Private managed invoker находится на стороне composition DP-013 и валидирует
  binding против своего immutable scope и `StartRequest` перед вызовом stored
  Flow.
- Authoritative immutable cross-layer значения находятся в dependency-leaf
  package `internal/runtimeorchestrationbinding`. Он может импортировать типы
  identity `runtimeconfigload`, но не импортирует `runtimecommandidempotency`,
  `runtimeidentity`, `runtimemanagement` или `runtimelaunchflow`.
- Managed Flow seam находится в `internal/runtimelaunchflow`. Существующий
  unmanaged конструктор `New(owner, loader)` и его семантика
  `Start(ctx, request)` остаются неизменными.
- Invocation — один synchronous per-invocation вызов
  `Flow.StartManaged(context, StartRequest, StartExecutionBinding)`.
- Четыре вышестоящих package могут зависеть от neutral binding package;
  neutral package не зависит ни от одного из них. `runtimelaunchflow` не
  импортирует `internal/runtimecommandidempotency` или `internal/runtimeidentity`.

#### 8.1.1 Seam primitive managed adapter

`runtimecommandidempotency.Boundary` владеет одним additive internal-repository
primitive adapter, концептуально:

```text
ExecuteManagedStart(
    context,
    Start Scope,
    CommandKey,
    Start Intent,
    ExpectedAggregateRevision,
    ExecutionGeneration,
    AuthorizeOrchestration,
    invoke(StartExecutionBinding) -> TerminalOutcome, error,
) -> Admission, error
```

Он принимает только exact primitive Start scope/intent. Он валидирует каждый
input, выводит шестипольный authorization request `ActivateExactTarget`,
выполняет authorization и pre-claim cancellation gate и только затем входит в
ту же command admission linearization point, что и `Execute`. Authorization
выполняется для каждой initial, in-progress и replay submission до inspection
команды; denial, panic, cancellation или invalid input дают zero mutation.

Callback получает только newly committed primitive claim. В той же locked
claim transaction, до освобождения admission locks, Boundary коммитит command
record и live permit, выделяет уникальную generation-bound identity
`StartRendezvous`, устанавливает её private lookup entry и определяет полный
primitive `StartExecutionBinding`. Все caller-supplied binding facts
валидируются до claim, поэтому deterministic construction binding не может
сломаться после command mutation. Затем callback выполняется один раз,
синхронно и вне всех command locks, без parent/phase identity.

In-progress и replay submission возвращают обычный Admission и не получают
callback, binding, allocation rendezvous или permit. Record, изначально claimed
через legacy `Execute`, также только наблюдается; managed seam никогда не
перенимает и не пересоздаёт его execution authority.

Permit, rendezvous lookup и callback authority истекают при возврате
callback/permit, panic, `runtime.Goexit` или потере generation Boundary. Это не
mutate и не invalidate structurally valid value binding; его no-reuse зависит
от callback custody и отсутствия bypass. Valid terminal outcome
публикуется существующими primitive permit rules. Callback error, panic,
invalid outcome, missing terminal publication или indeterminate return
оставляет команду Claimed/unresolved, прекращает live authority, блокирует
resolution rendezvous и возвращает существующую truthful indeterminate error;
callback не повторяется и не отделяется.

Существующий `Boundary.Execute` остаётся неизменным isolated legacy primitive
surface для текущих non-orchestration callers и тестов. Production activation
orchestration обязана использовать `ExecuteManagedStart`; она не может вызвать
`Execute`, затем синтезировать binding, вызвать private managed invoker напрямую
или fallback на legacy execution после managed failure. Public
`Directory.Start/Stop/Observe` DP-013 и существующие surfaces DP-015 остаются
неизменными; новый метод является repository-internal и не создаёт transport/API
path.

#### 8.1.2 Seam managed parent и StartTarget adapter

Replacement и rollback используют один additive managed parent path.
Концептуальные surfaces:

```text
Boundary.ExecuteManagedParent(
    context, Replace|Rollback Scope, CommandKey, matching Intent,
    AuthorizeOrchestration,
    invoke(*ManagedParentExecution) error,
) -> ParentAdmission, error

ManagedParentExecution.ContinueOrExecuteManagedStartTarget(
    context, ExpectedAggregateRevision, ExecutionGeneration,
    invoke(StartExecutionBinding) -> TerminalOutcome, error,
) -> PhaseAdmission, prevented bool, error
```

`ExecuteManagedParent` принимает только exact
Replace или Rollback, использует `AuthorizeOrchestration` для каждой initial,
in-progress и replay submission и является единственным источником
`ManagedParentExecution`. Parent, admitted через legacy `ExecuteParent`, нельзя
adopt или upgrade.

`ManagedParentExecution.ContinueOrExecuteManagedStartTarget(...)` сохраняет
существующий pre-phase Continue ordering. Callback получает только newly
committed phase `StartTarget`. В transaction claim phase метод выводит exact
parent identity и ordinal-one identity StartTarget, коммитит phase и её permit,
выделяет и индексирует уникальную generation-bound rendezvous identity,
конструирует полный linked `StartExecutionBinding`, а затем вызывает callback
один раз вне command locks. In-progress и replay observations не получают
binding или execution authority.

Существующие `ExecuteParent`, `ParentExecution` и
`ContinueOrExecuteStartTarget` остаются неизменными compatibility surfaces.
Managed path не синтезирует authority из их records.

### 8.2 Immutable per-invocation `StartExecutionBinding`

`StartExecutionBinding` — единственное authoritative immutable значение во
владении `runtimeorchestrationbinding`, сконструированное permit-holding stack
и переданное внутрь. Оно содержит validated шестипольный
`OrchestrationAuthorizationRequest`, expected aggregate revision,
composition-owned exact `ExecutionGeneration`, identity parent и phase, когда
применимо, и opaque `StartRendezvous` для этого live primitive или phase
execution. Оно не содержит ни primitive, parent, phase или Stop permit, ни
preparation token, ни Host или Snapshot, ни полномочие cancellation context, ни
mutable состояние Owner. Оно валидируется до любой mutation Owner, удерживается
только на том synchronous call stack и вызывается не более одного раза; оно
никогда не сохраняется как поле Flow. Возврат callback прекращает live
authority без mutation или invalidation structurally valid value binding.

Linked execution identity является явным all-or-none вариантом: primitive
Start не имеет ни parent, ни phase; `StartTarget` Replace/Rollback имеет и
exact parent command identity, и его command-boundary-derived phase identity
`StartTarget` с ordinal один. Одинокий parent, одинокая phase,
caller-selected phase или несовпадение action/variant невалидны. Store-owned
`Revision` и `ExecutionGeneration` остаются authoritative persistence
concepts; leaf package переносит validated lossless значения, явно
конвертируемые на границе `runtimeidentity`.

Managed construction связывает существующий stateless private
`StartClaimContinuation` ровно один раз. Binding не создаёт запись Registry,
mutable slot, goroutine, detached callback или новое состояние lifecycle.

### 8.3 Механизм opaque `StartRendezvous` через границу package

Существующий primitive rendezvous pending-Stop в
`internal/runtimecommandidempotency` остаётся единственным владельцем своих
сигналов и своих lock, согласно DP-019 §13 и §17. Seam через границу package —
opaque identity value без методов signal или wait, определённый в
`runtimeorchestrationbinding`, поэтому:

- граница команд DP-015 выделяет одну collision-safe identity для каждого live
  primitive Start или `StartTarget` execution и индексирует свой private
  concrete rendezvous этой identity вместе с активной generation Boundary;
- наружу в binding она выставляет только opaque identity;
- `StartExecutionBinding` хранит opaque handle;
- `runtimemanagement` и `runtimelaunchflow` передают его, не импортируя package
  границы команд.

Handle не несёт pointer, channel, function, capability, permit или mutable
состояние. Только `runtimecommandidempotency` разрешает эту identity и только
пока живы исходный callback-scoped execution и generation Boundary. Missing,
forged, reused, cross-generation или identity-mismatched handle дают
`Blocked`; они никогда не означают отсутствие pending Stop. Возврат callback и
замена Boundary прекращают resolution authority, но не стирают durable
unresolved facts. Global registry rendezvous вне command boundary отсутствует.

Command boundary выставляет только три repository-internal operation над
полным binding; их типы результата закрытые, не boolean:

```text
ResolveManagedStartEarly(binding, launchAttemptID)
    -> GateClear | GateStopConverged | GateBlocked

ResolveManagedStartFinal(binding, FinalContinue | FinalBindingFailed)
    -> FinalContinue | FinalBindingFailed | GateStopConverged | GateBlocked

SignalManagedStartNoClaim(binding, Cancelled | Rejected | Failed)
```

`ManagedStartGateOutcome`, `ManagedStartFinalDisposition` и их constants
принадлежат `runtimecommandidempotency`. Neutral immutable
`StartNoClaimCause` и три его constants принадлежат
`runtimeorchestrationbinding`; существующие имена command package могут
остаться aliases для compatibility. `StartClaimOutcome` и четыре его constants
принадлежат `runtimelaunchflow`. Такое размещение избегает каждого reverse
import.

Каждая operation разрешает opaque handle вместе с полным authorization tuple,
primitive-or-linked identity, exact command или phase, live permit и active
generation Boundary. Caller не передаёт command key: private lookup entry
владеет exact command identity. Missing, forged, reused, mismatched,
cross-generation или expired handles возвращают `GateBlocked` с существующей
indeterminate-execution error; они никогда не означают clear.

Один managed primitive или StartTarget rendezvous начинается в `PreOwner`.
Stop, admitted в этом состоянии, занимает single tracked-Start exception на
своём original stack `Boundary.Execute` и ждёт. Early operation сохраняет exact
attempt Owner и переходит в `Binding`; уже pending Stop получает signal, и
только его original claimant может вызвать Stop и опубликовать результат. Exact
convergence возвращает `GateStopConverged`; cancellation до delegation очищает
gate; lost или conflicting evidence блокирует его. Во время `Binding` один
более поздний Stop может claim и ждать. Final operation является единственной
admission linearization point между этим Stop и `FinalContinue` или
`FinalBindingFailed`. Stop-first обязан converge на своём original stack;
disposition-first запечатывает результат. После Continue более поздний Stop
использует ordinary tracked-Start behavior. Второй Stop или unrelated lifecycle
mutation не обходят ни один gate.

Возврат callback, panic, `runtime.Goexit` или замена generation Boundary
истекают handle и будят всех waiters как Blocked, пока durable command или phase
facts остаются Claimed. Command locks удерживаются только для validation,
transition и захвата notification; ни один command lock не удерживается во
время wait или вызова identity, continuation, Flow, Owner или external work.

### 8.4 Exact failed private-invocation error contract

Validation failure binding до mutation Owner возвращает exact, distinguishable
ошибку (fail-closed) и не выполняет mutation Owner или aggregate. Ошибки Owner
или continuation возвращаются неизменными и без обёртки; на этом seam нет ни
reclassification, ни wrapping, ни recovery.

## 9. Отложенное решение: OwnerClaimView, последовательность binding и исходы

### 9.1 `OwnerClaimView` и accessor `LaunchAttemptID`

Сразу после успешного authentic `Owner.PrepareStart` и до любой работы Load,
Build или Launcher Flow собирает один immutable `OwnerClaimView`:

```text
OwnerClaimView {
    WorkspaceID                  uint64
    ConfigurationID              uint64
    RuntimeInstanceID            runtimeconfigload.RuntimeInstanceID
    LaunchAttemptID              runtimeconfigload.LaunchAttemptID
    TargetConfigurationVersionID uint64
}
```

Доказательство accessor LaunchAttemptID присутствует уже сегодня: authentic
`LaunchPreparation.LoadRequest()` возвращает `runtimeconfigload.LoadRequest`,
который несёт `LaunchAttemptID()`. Claim view источён из этого Owner-issued
`LoadRequest`; изменение формы `runtimelifecycle.StartRequest` не требуется.

### 9.2 Закрытый contract и исход continuation

`runtimelaunchflow` владеет interface и lifecycle adaptation, поэтому сохраняет
отсутствие dependency на packages command или identity:

```text
StartNoClaim(context, StartExecutionBinding, StartNoClaimCause) error
AfterOwnerClaim(context, StartExecutionBinding, OwnerClaimView)
    -> StartClaimOutcome, error
```

Новый focused package `internal/runtimeorchestrationcontinuation` владеет
long-lived stateless implementation. Он зависит от
`runtimeorchestrationbinding`, `runtimecommandidempotency`, `runtimeidentity` и
`runtimelaunchflow`; ни один из них не зависит обратно от него. Он хранит только
command-boundary и narrow identity-boundary dependencies, но никогда не хранит
per-call binding, Owner preparation, permit, mutable rendezvous, goroutine или
registry entry.

`AfterOwnerClaim` возвращает ровно один закрытый исход:

- `Continue`: already-admitted rendezvous pending-Stop окончательно отсутствует;
  exact durable membership попытки и same-generation binding закоммичены при
  своих expected revision; final gate Stop-versus-Continue отпустил Flow.
- `StopConverged`: original pending Stop claimant вызвал exact Stop и свёл
  заявленную попытку; Flow не начинает Load и не выполняет публикацию или
  binding DP-014.
- `BindingFailed`: coherent exact evidence доказывает, что публикация membership
  попытки или binding generation не закоммичены и конфликта нет, и внешняя
  подготовка не начиналась; Flow подаёт failure через authentic подготовку
  Owner.
- `Blocked`: потеря permit или signal, stale или conflicting revision, другое
  generation, unavailable или unknown facts, недоказанная convergence Stop или
  indeterminate публикация; Flow не начинает Load, и связанный набор остаётся
  unresolved.

`Continue` и `StopConverged` требуют nil error. `BindingFailed` и `Blocked`
требуют одну exact non-nil cause. Invalid combination или panic даёт `Blocked`
с managed continuation validation error.

Caller cancellation наблюдается только в существующем gate DP-011 до claim
Owner. После успешного `PrepareStart` Flow выводит non-cancelable continuation
context (эквивалент `context.WithoutCancel(ctx)`, сохраняющий values, но не
cancellation или deadline) и использует его для `AfterOwnerClaim`. Каждый вызов
convergence Owner с той же preparation также использует local non-cancelable
context. `StartNoClaim` после definitive pre-claim outcome также получает local
non-cancelable context, поэтому caller cancellation не может стереть command
signal. Dependency failure или unavailable состояние identity/gate остаётся
Blocked и не переклассифицируется в caller cancellation.

Flow отображает исходы, не передавая continuation свой preparation token:
Continue вызывает существующий prepared-start path; StopConverged вызывает
`Owner.Start` с той же authentic preparation и empty result, получая сохранённый
outcome `StartStoppedBeforeRunning` без Load; BindingFailed
передаёт `FailedPreparation(cause)` через authentic Owner preparation и
возвращает semantic Owner outcome; Blocked также локально сводит in-memory
Owner claim, чтобы не оставить leaked preparation, но возвращает exact Blocked
cause, поэтому command или phase остаётся unresolved. Failure convergence Owner
имеет приоритет над continuation cause. Definitive pre-Owner cancellation,
rejection или failure вызывает `StartNoClaim`; failure signal делает execution
Blocked.

Continuation никогда сам не публикует terminal исход lifecycle, phase или
parent; terminal их могут только exact результат Owner, за которым следуют
conditional публикации DP-014 и DP-015.

### 9.3 Порядок последовательности binding и threading revision

Фиксированный порядок, после единственного claim Owner и до Load:

1. разрешить already-admitted rendezvous pending-Stop только через original Stop
   claimant; `StopConverged` выходит без обеих записей, а `Blocked` оставляет
   связанный набор unresolved;
2. только после definitive доказательства отсутствия pending Stop прочитать
   exact Runtime Instance и до любой mutation доказать workspace,
   configuration, instance identity и expected revision; read failure или
   любое несоответствие дают `Blocked` с нулём записей;
3. после этого pre-mutation proof conditionally
   опубликовать exact membership Launch Attempt и pin версии в DP-014 при
   expected aggregate revision;
4. прочитать committed revision, возвращённый этой записью, и conditionally
   связать exact active попытку с composition-owned execution generation при
   этом новом expected revision;
5. после любого indeterminate результата проверить exact aggregate facts через
   `ReadRuntimeInstance` и `ReadLaunchAttemptHistory` и свести к exact
   существующему terminal исходу или вернуть `Blocked`;
6. выполнить final gate Stop-versus-Continue для Stop, admitted после ранней
   проверки rendezvous, затем выпустить `Continue` только при confirmed, exact
   same-generation binding.

Continuation зависит от narrow interface `IdentityStore`, соответствующего
существующим operations Store, чтобы unavailable и indeterminate behavior были
тестируемы; сам `runtimeidentity.Store` остаётся неизменным. Конверсия на этой
границе явная и lossless:
`runtimeidentity.Revision(uint64(binding.ExpectedAggregateRevision()))` и
`runtimeidentity.ExecutionGeneration(string(binding.ExecutionGeneration()))`.
Zero или round-trip mismatch дают Blocked; значение не выделяется и не
выводится.

Revision, возвращённый committed claim attempt, является единственным expected
revision для binding generation. Stale write никогда не retry. После любой
error или non-commit exact inspection использует revision sandwich:
`ReadRuntimeInstance A -> ReadLaunchAttemptHistory -> ReadRuntimeInstance B`.
Pre-mutation scope/revision read не является observation A для sandwich;
inspection начинается со свежего чтения после ambiguous operation.
Результат coherent только когда A и B имеют равные revision и одинаковые
immutable identity и active-attempt facts. Exact active attempt, version,
Claimed phase и generation могут доказать satisfied convergence. Coherent
absence при всё ещё current relevant expected revision, без conflicting
history, active attempt или generation и без external preparation, может дать
BindingFailed. Stale revision, changed sandwich, different или reused attempt,
different version или generation, inactive или terminal attempt, read failure,
unavailability или unknown facts дают Blocked. Ни один исход не разрешает blind
retry или repair.

Interface содержит ровно operations, нужные этой последовательности:
`ConditionalClaimLaunchAttempt`, `ConditionalBindExecutionGeneration`,
`ReadRuntimeInstance` и `ReadLaunchAttemptHistory` с существующими parameter и
result types `runtimeidentity`. У него нет terminal-publication, recovery,
allocation или mutation method кроме claim и bind.

Уже присутствующие same-attempt/same-version membership и same-generation
binding являются zero-mutation satisfied наблюдениями convergence. Другая
попытка, другая версия, другой generation, stale revision, inactive факт или
unknown состояние никогда не auto-repaired или auto-replaced.

## 10. Инвариант адаптации активации

Initial activation сохраняет primitive identity и intent Start:

- orchestration использует `Boundary.ExecuteManagedStart` как свой единственный
  primitive managed admission; legacy `Boundary.Execute` остаётся неизменным
  только для isolated compatibility callers и не может быть adopted
  orchestration;
- `runtimelifecycle.StartRequest` остаётся immutable существующей формой
  `(WorkspaceID, ConfigurationID, ConfigurationVersionID)` и не несёт authority
  parent, phase или authorization;
- composition-facing authorization adapter помечает exact authorization action
  submission Start как `ActivateExactTarget`;
- parent с одной phase не создаётся ради uniformity API; формы parent
  используются только для `Replace` и `Rollback`.

Replacement и rollback отображаются на `ReplaceWithExactTarget` и
`RollbackToExactTarget` через существующий путь parent `ExecuteParent`,
представленный TASK-028 и завершённый TASK-029. Нет fallback между actions и нет
кэшированного результата авторизации.

## 11. План доказательства Lock и Ownership

Перед принятием любого implementation среза ответственные роли должны доказать
focused тестами и статическим обзором каждого сайта `Lock`/`Unlock`, что ни один
из этих lock не удерживается через названные операции:

- ledger admission команд DP-015 и lock клиента хранилища;
- lock aggregate store DP-014 и lock per-aggregate;
- состояние Directory DP-013 (которое намеренно не имеет process lock);
- managed Flow DP-011 (который намеренно не имеет process lock);
- mutex Runtime Lifecycle Owner;
- любую синхронизацию private для нового среза.

Ни один не удерживается через: (1) callback authorizer; (2) операцию storage
DP-014 или exact inspection; (3) signal или wait continuation; (4) invocation
lifecycle DP-013 или Owner; (5) Load, Build, Launcher, Host или external I/O.

Ownership остаётся: external authorization policy принадлежит composition и
заимствуется на submission; записи parent/phase DP-015 и live permits остаются в
границе команд; факты aggregate и attempt DP-014 остаются в identity store;
Lifecycle Owner остаётся единственным принимающим решения lifecycle и owner live
Host; DP-013 остаётся exact composition routing/private-invocation; DP-011
остаётся synchronous последовательностью подготовки; rendezvous pending-Stop
остаётся в bounded исполнении оркестрации. Ни одна строка не передаёт ownership
оркестратору.

## 12. Упорядоченные Implementation-срезы (выход готовности)

### Срез 1 — поверхность orchestration authorizer

Текущий статус среза: partial isolated implementation, исторически принятая
TASK-031; отсутствующий `OperationalDomain` исправлен изолированно Срезом 2R
TASK-035.

- Ввести validated значение `OrchestrationAuthorizationRequest`, named
  policy-neutral тип функции `AuthorizeOrchestration`, набор
  `OrchestrationAction` и exact failed-authorization error contract на
  существующей границе команд, не меняя public поверхности DP-013 / DP-015.
- Доказать: exact initial/replay авторизация, нулевая mutation при
  denial/failure/cancellation, zero-path активации на immutable primitive
  submission Start, и отсутствие public или private bypass.
- Никакого binding DP-014, никакого изменения Flow, никакого оркестратора.

### Срез 2 — private managed invoker и managed Flow seam

Текущий статус среза: partial isolated implementation, исторически принятая
TASK-032; Срез 2R TASK-035 предоставляет полный authoritative binding repair,
не переписывая историю acceptance TASK-032.

- Добавить managed construction и per-call seam `StartManaged` и immutable
  значения `StartExecutionBinding` / `OwnerClaimView`, opaque handle
  `StartRendezvous` и exact failed private-invocation error contract, используя
  Slice 1.
- Доказать: validate-before-Owner-mutation, invoke-at-most-once,
  callback-scoped expiry authority, custody-based no-reuse и never-stored
  binding без mutation structural value; неизменные unmanaged `New` и `Start`;
  никакого goroutine, записи registry или detached callback.
- Ещё никакой логики binding DP-014.

### Срез 2R — repair соответствия managed binding

Текущий статус среза: реализован и независимо принят изолированно TASK-035.
Production composition/private-invoker wiring не входит в этот срез.

- Ввести dependency-leaf authoritative binding values, восстановить
  `OperationalDomain`, перенести полный authorization tuple и all-or-none
  linked identity, заменить constant token уникальной command-owned identity
  rendezvous, удалить import Flow-to-`runtimeidentity` и предоставить путь
  validation composition-private invoker, а также добавить
  `Boundary.ExecuteManagedStart` как единственный primitive orchestration
  claim-to-binding adapter при неизменном `Execute`.
- Доказать constructor и cross-field validation, exact command-to-binding и
  Owner-request mapping, отклонение unique/stale/forged rendezvous,
  synchronous call-stack lifetime, неизменность unmanaged Flow и public
  поведения DP-013, acyclic imports и zero mutation при каждом validation
  failure.
- Доказать authorization-before-inspection для initial/replay, callback только
  для new claim, atomic installation claim/permit/rendezvous, отсутствие replay
  adoption, expiry callback lifetime, unresolved при panic/error/indeterminate
  и отсутствие orchestration fallback на legacy `Execute`.
- Не реализовывать publication/binding DP-014, continuation outcomes,
  orchestrator, policy, persistence, API или production wiring.

### Срез 3 — последовательность связывания OwnerClaim-to-DP-014

Текущий статус среза: реализован и независимо принят изолированно TASK-037.

- Реализовать managed parent/StartTarget adapter, общие command-owned early/final
  rendezvous gates и stateless `StartClaimContinuation.AfterOwnerClaim` через
  narrow identity-store boundary над существующими operations
  `runtimeidentity.Store`, используя Срезы 1, 2 и 2R.
- Доказать: membership попытки и same-generation binding до Load; stale/different/
  unknown facts дают `Blocked` без подготовки; definitive отсутствие binding
  сходится через authentic исход Owner; Stop, admitted после ранней проверки,
  упорядочен final gate для primitive и linked execution.
- Не заявлять terminal publication DP-014 или terminalization command/phase
  DP-015: более поздний orchestration callback может опубликовать их только
  после exact Owner outcome. Изолированные Blocked и BindingFailed paths могут
  намеренно оставить command или phase Claimed.

### Срез 4 — переоценка готовности оркестратора DP-016

Текущий статус среза: завершён и Coordinator Accepted как TASK-038. Его reassessment из 19 строк
фиксирует `TASK-026 REMAINS BLOCKED`: 7 Direct, 10 Compositional, 2 Missing и
0 Deferred proofs после Reviewer rework и repeat Reviewer APPROVED 0/0.
Принятый verdict не активирует implementation и не меняет статус TASK-026 автоматически.

- Только после того, как Slices 1–3, включая Срез 2R, реализованы и независимо
  приняты, переоценить,
  может ли TASK-026 быть разблокирована против неизменных proofs §25 DP-016.
- TASK-038 определяет design-only contract атомарного expected-attempt Owner
  Stop как первую bounded prerequisite. TASK-039 завершена с Coordinator
  Acceptance после фиксации этого design в Draft DP-010; завершённая и
  Coordinator-Accepted TASK-040 реализует и верифицирует isolated Owner
  extension, repeat final Reviewer `APPROVED` 0/0. Design private exact-scope
  composition invoker теперь зафиксирован Draft DP-021, тогда как его
  implementation остаётся последующей и неактивированной.

Каждый срез требует собственного intake задачи, Existing Coverage Report,
Verification Matrix, Independent Review, PROCESS-002 и Coordinator Acceptance.

## 13. Acceptance Proofs

Этот дизайн сам доказывает:

1. выбранное разложение сохраняет каждое обязательство acceptance proof Approved
   DP-019 §21 и отображает каждое оставшееся proof ровно на один срез;
2. существующие поверхности DP-013, DP-014, DP-015 и DP-016 не требуют semantic
   изменения для размещения трёх исходных implementation-срезов и repair
   соответствия Среза 2R;
3. активация остаётся на primitive immutable command path Start, не использует
   synthetic parent/phase identity и не может обойти тот же managed
   continuation prerequisite, который требует Approved DP-019;
4. locks доказуемо не удерживаются через forbidden границы;
5. зеркала EN и RU семантически равны, с равными заголовками и равными количествами
   code-fence, и каждая relative ссылка разрешается.

Эти proofs проверяются проверками документации, parity, ссылок, статуса и
regression, а также свежим Independent Review этого предложения.

## 14. Граница реализации

Implementation Status остаётся Planned overall. Сама design-задача не
реализовала ни один срез; successor TASK-031 и TASK-032 создали исторически
принятые Coordinator частичные изолированные реализации Срезов 1 и 2. TASK-034
обнаружила оставшийся gap соответствия, TASK-035 реализует и независимо
принимает его repair Среза 2R изолированно, а TASK-036 устраняет оставшуюся
неоднозначность command-gate и continuation API Среза 3. TASK-037 реализует и
независимо принимает Срез 3 изолированно. Репозиторий содержит принятый Draft
design и завершённую TASK-040 isolated implementation atomic expected-attempt Owner
Stop, но всё ещё не содержит concrete private exact-scope composition invoker,
определённый Draft DP-021,
оркестратор активации, external persistence,
API, worker recovery и production wiring. Последующая
terminal publication DP-014 и terminalization command/phase DP-015 после
результата Owner принадлежат orchestrator TASK-026, а не отдельной prerequisite.
Поэтому TASK-026 остаётся Blocked; Срез 4 завершён и принят как TASK-038.

## 15. Последствия

Положительные:

- каждый оставшийся prerequisite DP-019 становится независимо проверяемым и
  независимо приемлемым;
- authorization, lifecycle ownership, durability ordering и permit ownership
  остаются explicit;
- здесь не принимается никакого production или public контракта.

Стоимость:

- TASK-038 выполняет разрешённую переоценку readiness, а завершённая TASK-039
  фиксирует только её design-only candidate atomic expected-attempt Stop;
- synchronous rendezvous pending-Stop может блокировать callers;
- restart процесса по-прежнему требует Planned реализации DP-017;
- production integration по-прежнему требует external durability и аудита
  composition.

## 16. Решение

UWP фиксирует разложение готовности оставшихся prerequisites Approved DP-019 в
этом Draft/Planned предложении и реализует каждый срез только через отдельную,
индивидуально reviewed задачу. Срезы 1 и 2 остаются исторически принятыми
частичными реализациями; TASK-035 реализует и независимо принимает Срез 2R
изолированно; TASK-036 фиксирует exact протокол Среза 3; а TASK-037 реализует и
независимо принимает Срез 3 изолированно. Срез 4 завершён и принят как TASK-038
с remains-Blocked verdict. Proposal не
approximates DP-016 adapter-ом,
не добавляет операции replacement/rollback Owner, не передаёт permits, не меняет
ни один Approved статус или семантику и не выдаёт planned capability за
реализованную.

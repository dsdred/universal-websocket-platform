# DP-019: Предпосылки оркестрации активации Runtime

[English version](../../en/design/DP-019-runtime-activation-orchestration-prerequisites.md)

## 1. Статус

- **Статус проектирования:** Approved
- **Статус реализации:** Planned в целом; parent/phase storage, callback
  capability, sequential phase core и command-boundary Continue/pending-Stop
  rendezvous реализованы изолированно; managed gates и continuation Среза 3
  реализованы и независимо приняты изолированно

Этот focused design закрывает только неоднозначность integration-contract,
обнаруженную TASK-026. Репозиторий реализует isolated parent/phase core DP-015,
command-boundary Continue/pending-Stop rendezvous, exact authorization и
dependency-leaf binding values, primitive managed adapter и seam managed
Flow/OwnerClaimView. TASK-037 реализует изолированно managed
parent/StartTarget adapter, общие managed gates, concrete stateless
continuation, binding sequence attempt/generation DP-014 и exact адаптацию
outcomes managed Flow, реализованные и независимо принятые изолированно.
TASK-038 определяет factual readiness gap: реализованный Stop DP-010 не может
атомарно выбрать ожидаемый Launch Attempt. Завершённая и Coordinator-Accepted
TASK-039 фиксирует принятый design этой operation в Draft DP-010, но до concrete private exact-scope composition
invoker всё ещё требуется implementation. Это утверждение не изменяет данный
Approved decision. Terminal publication после
Owner относится к самому последующему orchestrator TASK-026; этот orchestrator,
external persistence, API, recovery worker и production wiring отсутствуют.
TASK-026 остаётся Blocked до реализации design TASK-039 и независимой приёмки
этой реализации, последующего private exact-scope composition invoker и последующей переоценки
readiness.

## 2. Назначение

Определить точные внутренние contracts, необходимые для реализации Approved
DP-016 без ослабления его ordering или acceptance proofs:

- bounded DP-015 parent/phase claim surface для replacement и rollback;
- точную авторизацию intents activation, replacement и rollback;
- private DP-011/DP-013 continuation после единственного Owner claim и до
  Load;
- durable publication этого exact attempt и его execution generation до
  внешней preparation work.

## 3. Источники полномочий

Proposal уточняет, не переопределяя:

- [ARCH-004](../architecture/ARCH-004-runtime-deployment-and-identity-model.md),
  особенно section 19(4);
- [DP-010](DP-010-runtime-lifecycle-owner-contract.md), чей Owner остаётся
  единственным lifecycle authority;
- [DP-011](DP-011-runtime-launch-pipeline-integration.md) для synchronous
  launch path и места private continuation;
- [DP-013](DP-013-runtime-management-routing.md) для exact routing и
  authorization-before-mutation;
- [DP-014](DP-014-runtime-operational-identity-persistence.md) для aggregate,
  attempt, revision и execution-generation facts;
- [DP-015](DP-015-runtime-management-command-idempotency.md) для claim,
  replay, permits и unresolved barriers;
- [DP-016](DP-016-runtime-activation-replacement-rollback.md) для полного
  activation/replacement/rollback ordering;
- [DP-017](DP-017-runtime-recovery-reconciliation.md) для fail-closed recovery.

Accepted ADR и Active/Frozen architecture остаются authoritative.

## 4. Область действия

DP-019 определяет один internal integration contract, содержащий:

- orchestration actions и exact authorization input;
- replacement/rollback parent identity и два конечных linked phase kinds;
- callback-scoped parent и phase execution capabilities;
- phase admission, replay, terminalization и unresolved behavior;
- private exact-scope lifecycle invocation после authorization;
- construction, input, outcomes и ordering Start-claim continuation;
- exact attempt publication и generation binding до Load;
- cancellation, lock, ownership и failure boundaries.

Он не реализует эти contracts и не меняет Runtime lifecycle behavior.

## 5. Не-цели

Design не определяет:

- public Activate, Replace или Rollback API, DTO, route или status code;
- Principal representation или concrete policy;
- generic workflow engine, saga framework, command bus или registry;
- новые Owner states, operations или ownership;
- storage schema, transaction product, migration или cross-process lease;
- recovery execution, reporting, automatic rollback/restart, zero-downtime
  replacement, supervision, scheduling или Production Activation.

## 6. Краткое решение

UWP выбирает путь DP-015 parent/phase API вместо новых lifecycle operations.

- Initial activation остаётся одним primitive exact-version Start command.
- Replacement и rollback являются разными parent operation kinds.
- Каждый accepted parent имеет не более `StopOld`, затем `StartTarget`.
- Parent permit никогда не вызывает lifecycle work.
- Только permit новой committed phase вызывает одну exact scoped lifecycle
  operation и остаётся на исходном synchronous call stack.
- Только Owner claim/converge Launch Attempt.
- Private continuation публикует exact Owner claim в DP-014 и связывает его с
  composition-owned generation до Load.

Ни один acceptance proof DP-016 не откладывается и не ослабляется.

## 7. Точная модель авторизации

Orchestration boundary определяет caller actions:

```text
ActivateExactTarget
ReplaceWithExactTarget
RollbackToExactTarget
```

Каждая initial submission и replay авторизуется до inspection или claim
команды. Immutable authorization input равен:

```text
OperationalDomain
WorkspaceID
ConfigurationID
RuntimeInstanceID
Action
TargetConfigurationVersionID
```

Planned policy-neutral seam является одной synchronous function, conceptually
equivalent:

```text
AuthorizeOrchestration(context, AuthorizationRequest) -> error
```

`AuthorizationRequest` — validated immutable value, содержащий ровно tuple
выше. Initial activation адаптирует существующую primitive DP-015 Start
submission к `ActivateExactTarget`; `ExecuteParent` использует
`ReplaceWithExactTarget` или `RollbackToExactTarget`. Fallback между actions и
cached authorization result отсутствуют.

Все identities ненулевые и exact. Target version должна быть Published и
принадлежать той же Configuration. Rollback отличается caller intent и
historical-target precondition DP-016; previous/latest version не выводится.
Principal, credential, deadline, trace data, current aggregate state и
inferred source version не являются durable intent facts.

Authorization denial/failure/panic, invalid target, absent scope или
pre-claim cancellation создают zero command, aggregate и lifecycle mutation.
Replay авторизуется снова; результат не сохраняется как durable authority.

## 8. Авторизация и полномочия фаз

Authorization относится к полному external orchestration intent. Linked
phases не являются новыми caller submissions и не переопределяются на другой
Principal, Target, version или policy result. Их authority происходит только
из accepted immutable parent и ограничена объявленной phase.

Exact management scope после authorization предоставляет private
composition-only lifecycle invoker. Он route только к уже связанным
Flow/Owner и не является public Directory bypass. Его planned Start path
conceptually равен:

```text
InvokeManagedStart(context, StartRequest, StartExecutionBinding) -> exact Owner outcome
```

`StartExecutionBinding` — per-invocation immutable value с validated
`AuthorizationRequest`, expected aggregate revision, exact execution
generation, parent/phase identity при применимости и opaque
`StartRendezvous` этого live primitive/phase execution. Он не содержит
primitive, parent, phase или Stop permit. Call stack с exact live permit
создаёт binding и synchronously один раз вызывает operation; он не может
получить другой scope, изменить target или передать permit.

Private invoker проверяет binding против immutable scope и StartRequest, затем
вызывает planned managed Start surface одного stored Flow с тем же binding. Он
не создаёт Flow, registry entry, mutable current-operation slot, goroutine или
detached callback.

Существующее public DP-013 Start/Stop/Observe authorization behavior не
меняется для primitive submissions. Production composition должна доказать
отсутствие unauthenticated transport path к private invoker.

## 9. Parent identity и immutable intent

Parent command identity остаётся `(CommandScope, CommandKey)`. Operation
parent scope равна точно `Replace` или `Rollback`. Immutable intent содержит
exact authorization tuple и target version.

Coherent source observation, old attempt identity, old pinned version,
expected aggregate revision и execution generation являются bound execution
preconditions. После parent claim они могут только уточняться exact
conditional reads и не могут retarget immutable caller intent.

Initial activation использует существующие primitive Start scope и intent. Он
не создаёт однофазный parent только ради API uniformity.

## 10. Конечный parent/phase автомат

Допустимы только linked shapes:

```text
Replace parent: [StopOld when required] -> StartTarget
Rollback parent: [StopOld when required] -> StartTarget
```

При отсутствии active attempt `StopOld` пропускается. Если exact target уже
Running, parent может terminalize Satisfied без phase. Другие phase kind,
повтор, reorder, branch, loop, retry phase или caller-selected phase identity
недопустимы.

Каждая phase identity collision-safe производна от parent identity, phase
kind и fixed ordinal. Она immutable, не выбирается caller, не используется
другим parent и хранится вместе с parent.

## 11. Parent/phase claim API

DP-015 extension callback-scoped и сохраняет existing non-transferable permit
invariant. Durable parent/phase storage, generation-bound callback capability и
strict sequential core реализованы изолированно; command-boundary
Continue/pending-Stop surface также реализован там изолированно. Managed
continuation и binding остаются Planned.

Conceptually он предоставляет:

```text
ExecuteParent(submission, authorize, invokeParent) -> admission/result

invokeParent(parentExecution):
    InspectOrClaimPhase(StopOld, invokeExactStop)
    ContinueOrClaimPhase(StartTarget, invokeManagedStartWithBinding)
    PublishParentTerminal(exact linked outcomes)
```

`ExecuteParent` выполняет validation, exact authorization, cancellation gate,
same-key decision, per-Instance barrier evaluation и parent claim под existing
admission boundary. Только call, committed новый parent, получает
`parentExecution`.

`parentExecution` действителен только во время synchronous callback и для exact
storage-client generation. Retain, возврат callback, panic, `runtime.Goexit`
или generation replacement expire все неопубликованные live capabilities и
оставляют committed non-terminal parent/phase unresolved.

## 12. Семантика claim фаз

Phase claim атомарно проверяет:

- exact parent identity, intent, state, revision и live parent execution;
- expected next phase kind и ordinal;
- per-Instance barrier и единственное DP-016 Stop exception;
- отсутствие уже committed conflicting phase;
- current storage-client generation.

Новая committed phase создаёт один durable linked record и один private live
phase permit. Callback выполняет не более одной exact lifecycle invocation.
Same-phase replay возвращает in-progress или terminal facts без permit.
Definitive claim failure выполняет zero lifecycle mutation. Indeterminate
claim, lost callback, invalid outcome, panic или publication uncertainty
оставляет linked set unresolved.

Parent terminal publication разрешена только после definitive durable
terminal outcome каждой required phase либо definitive zero-mutation parent
outcome до следующей phase. Parent не фабрикует phase result только из
aggregate observation.

## 13. Continue Gate

`ContinueOrClaimPhase(StartTarget, ...)` разделяет одну per-Instance admission
linearization point с independent Stop claim:

- Stop first: parent terminalize Stopped/Cancelled, Start phase отсутствует;
- Start phase first: существует ровно один Start phase permit, и один later
  Stop может занять tracked-Start exception;
- indeterminate: ни одна сторона не выводит winner; linked set unresolved.

Gate не создаёт Launch Attempt заранее и не выполняет lifecycle/storage
callbacks под admission lock.

## 14. Private Start-Claim Continuation

Один long-lived managed Flow immutably создаётся с одним stateless private
synchronous service `StartClaimContinuation`. Per-operation facts никогда не
хранятся в Flow и не bind при construction. Existing unmanaged Flow constructor
и isolated semantics не меняются. Production activation может использовать
только managed construction/invocation path.

Для каждого primitive Start или phase `StartTarget` original permit-holding
stack вызывает:

```text
Flow.StartManaged(context, StartRequest, StartExecutionBinding)
```

`StartManaged` проверяет per-call binding до Owner mutation и удерживает его
только на synchronous call stack. После `Owner.PrepareStart` он вызывает
`StartClaimContinuation.AfterOwnerClaim(StartExecutionBinding,
OwnerClaimView)`. Возврат `StartManaged` invalidates per-call binding; он
никогда не является Flow field и не влияет на другой invocation.

Сразу после успешного authentic `Owner.PrepareStart` и до Load, Build или
Launcher Flow передаёт один immutable Owner claim view:

```text
WorkspaceID
ConfigurationID
RuntimeInstanceID
LaunchAttemptID
TargetConfigurationVersionID
```

Exact attempt и version происходят из Owner-issued LoadRequest. Continuation
получает expected revision, generation, authorization tuple, linked execution
identity и rendezvous только из того же per-call `StartExecutionBinding`.
Cross-call/scope mismatch fail-closed. Ни одно value не содержит preparation
token, Host, Snapshot, context cancellation authority, parent/phase/Stop
permit или mutable Owner state.

## 15. Outcomes continuation

Continuation возвращает ровно одно:

- `Continue`: exact durable attempt membership и same-generation binding
  подтверждены, final Stop gate отпустил Flow;
- `StopConverged`: original pending Stop claimant вызвал exact Stop и converged
  claimed attempt; Flow не начинает Load;
- `BindingFailed`: coherent exact evidence доказывает, что attempt publication
  или generation binding не committed без conflict и external preparation не
  начиналась; Flow проводит failure через authentic Owner preparation;
- `Blocked`: permit/signal loss, stale/conflicting revision, different
  generation, unavailable/unknown facts, unproven Stop convergence или
  indeterminate publication; Flow не начинает Load, linked set unresolved.

Continuation не публикует lifecycle, phase или parent terminal outcome. Их
terminalize только exact Owner results и последующие DP-014/DP-015 conditional
publications.

## 16. Publication attempt и binding generation

Единственным in-process claim остаётся `Owner.PrepareStart`. Затем до Load
continuation:

1. проверяет Owner-issued identities против authorized target;
2. разрешает уже admitted pending-Stop rendezvous, если он есть, только через
   original Stop claimant; `StopConverged` завершает путь без publication
   attempt membership и generation binding, а `Blocked` оставляет linked set
   unresolved;
3. только после definitive proof отсутствия pending Stop conditionally
   публикует exact Launch Attempt и version pin в DP-014 при
   expected aggregate revision;
4. conditionally bind exact attempt к composition-owned execution generation;
5. inspect exact aggregate/attempt facts после любого indeterminate result;
6. выполняет final Stop-versus-Continue gate для Stop, admitted после early
   rendezvous check.

Это уточнение сохраняет DP-016 ordering: command/phase claim предшествует Owner
claim, durable attempt membership и generation binding предшествуют всей
external preparation, Running невозможен до Host readiness. Ни continuation,
ни persistence не становится lifecycle owner.

Exact same-attempt/same-version membership или same-generation binding могут
converge idempotently. Different attempt/version/generation, stale revision,
inactive fact или unknown state автоматически не repair/replace.

## 17. Rendezvous pending Stop

Stop, занявший tracked-Start exception до Owner claim, остаётся на исходном
synchronous claiming stack с private permit. Continuation signal только
`OwnerClaimed` с exact attempt identity.

Только этот claimant повторно проверяет cancellation gate, один раз вызывает
exact private DP-013 Stop, публикует outcome и signal result. Continuation не
получает permit и не выполняет Stop вместо него. Start path, definitive
завершившийся до Owner claim, signal `StartNoClaim`; original Stop path может
terminalize satisfied без lifecycle invocation.

Этот rendezvous завершается до DP-014 attempt publication или generation
binding. `StopConverged` возвращает continuation без обеих записей; только
definitive no-pending-Stop result разрешает publication/binding path раздела 16.
Stop, admitted после early check, упорядочивается final Stop-versus-Continue
gate.

Lost signal, caller return без definitive publication, abandoned permit или
unproven convergence дают `Blocked`, никогда implicit Continue.

## 18. Cancellation и failure

Cancellation до external authorization или command claim создаёт zero
mutation. После parent/phase claim cancellation не удаляет и не передаёт
claim. DP-010/DP-011 gates остаются authoritative для lifecycle invocation.

Definitive pre-invocation failure может terminalize exact parent/phase с
zero-mutation outcome. После возможной mutation отсутствие exact terminal
publication означает unresolved. Panic и `runtime.Goexit` выполняют capability
expiry cleanup; они не восстанавливают permit и не разрешают retry.

Никакой failure не resurrect old Host, не infer release, не выбирает другую
version и не запускает automatic rollback.

## 19. Concurrency и locks

Один Runtime Instance использует одну DP-015 admission boundary для parent,
phase, independent Stop, replay и unresolved-barrier decisions. DP-014
сохраняет conditional aggregate revision boundary; Owner сохраняет lifecycle
serialization.

Ни command-admission, aggregate, Directory, Flow, ни Owner lock не удерживается
через authorization callback, DP-014 storage/inspection, continuation
signal/wait, DP-013/Owner invocation, Load, Build, Launcher, Host или external
I/O.

Разные Runtime Instances не разделяют orchestration lock и прогрессируют
независимо.

## 20. Ownership

| Fact или capability | Owner |
| --- | --- |
| External authorization policy | composition; borrowed на submission |
| Parent/phase records и live permits | DP-015 boundary |
| Runtime Instance/attempt/generation facts | DP-014 boundary |
| Lifecycle decision и live Host | Runtime Lifecycle Owner |
| Exact scope routing/private invoker | DP-013 composition |
| Synchronous preparation sequence | DP-011 Flow call stack |
| Continue/pending-Stop rendezvous | bounded orchestration execution |

Ни одна строка не передаёт ownership orchestrator.

## 21. Acceptance Proofs

Prerequisite implementation обязана доказать минимум:

1. exact orchestration authorization initial/replay submissions;
2. denial/failure/cancellation до claim создаёт zero mutation;
3. один parent claim и immutable intent при concurrency;
4. принимаются только legal phase order и derived identities;
5. parent permit не может вызвать lifecycle work;
6. каждый phase permit вызывает не более одной exact scope operation;
7. lost/abandoned parent или phase capability оставляет durable barrier;
8. Continue gate имеет одного winner против independent Stop;
9. pending Stop permit остаётся и используется только original stack;
10. Owner claim view совпадает с exact target/attempt/version;
11. durable attempt membership и generation binding завершаются до Load;
12. stale/different/unknown facts дают Blocked без preparation;
13. definitive binding absence converges через authentic Owner outcome;
14. panic/error/cancellation/`runtime.Goexit` не дублируют work;
15. reconstruction сохраняет records, но не восстанавливает live permit;
16. different Instances progress independently;
17. unmanaged Flow не используется production activation composition;
18. public/private routing bypass до authorization отсутствует;
19. EN/RU и linked DP semantics aligned.

Эти proofs не доказывают сам DP-016 orchestrator.

## 22. Граница реализации

Implementation Status остаётся Planned в целом. Package
`internal/runtimecommandidempotency` теперь реализует изолированно exact
Replace/Rollback intent, durable parent и derived `StopOld`/`StartTarget`
records, generation-bound callback capabilities, strict порядок optional
StopOld затем StartTarget, phase replay, parent terminal gating, unresolved
barriers, reconstruction invalidation, non-bypassable StartTarget Continue gate
и synchronous pending-Stop rendezvous с immutable signal cause.

TASK-031/TASK-032/TASK-035 реализуют изолированно exact authorization и binding
values, primitive managed adapter и seam managed Flow/OwnerClaimView. TASK-037
реализует изолированно managed parent/StartTarget adapter, общие managed gates,
concrete stateless continuation, sequence membership/generation binding
попытки и отображение outcomes managed Flow, реализованные и независимо
принятые изолированно. TASK-038 дополнительно устанавливает, что atomic
expected-attempt Owner Stop должен предшествовать concrete private scoped
invoker. Завершённая и Coordinator-Accepted TASK-039 фиксирует его принятый
design в Draft DP-010; implementation всё ещё отсутствует. Репозиторий также не содержит этот invoker,
последующую terminal publication и terminalization command/phase, activation
orchestrator и production composition audit полного design.

Поэтому TASK-026 остаётся Blocked. Следующие tasks должны реализовать и
независимо проверить оставшиеся prerequisites. Только после этого TASK-026
может быть повторно оценена против полного неизменённого набора DP-016 proofs.
Focused readiness decomposition этих prerequisites зафиксирована в зеркальном
[DP-020](DP-020-runtime-orchestration-binding-sequence-readiness.md), со
статусом Design Status Draft и Implementation Status Planned overall, где
Срез 3 реализован и независимо принят изолированно.

## 23. Последствия

Положительные:

- blocking integration surfaces становятся конечными и независимо testable;
- DP-016 ordering не меняется;
- authorization и lifecycle ownership остаются explicit;
- permit/process loss fail closed.

Стоимость:

- минимум одна prerequisite design task и её implementation предшествуют
  TASK-026;
- synchronous pending-Stop rendezvous может блокировать callers;
- process restart всё ещё требует Planned DP-017 implementation;
- production integration всё ещё требует external durability и composition
  audit.

## 24. Решение

UWP реализует prerequisites activation orchestration как один focused internal
contract: exact intent authorization, bounded DP-015 parent/phase claims и
synchronous DP-011/DP-013 Start-claim continuation, публикующий Owner-issued
attempt и generation binding до Load. UWP не approximates DP-016 adapter,
не добавляет replacement/rollback operations Lifecycle Owner, не передаёт
permits и не выдаёт planned behavior за implemented.

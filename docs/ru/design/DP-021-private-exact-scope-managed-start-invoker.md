# DP-021: Private Exact-Scope Managed Start Invoker

[English version](../../en/design/DP-021-private-exact-scope-managed-start-invoker.md)

## 1. Статус

- **Design Status:** Draft
- **Implementation Status:** Partial — реализован изолированно

Этот focused proposal определяет один composition-private invocation contract
DP-013. TASK-043 реализует этот object изолированно в существующем package; он
не добавляет public API, transport, policy, persistence или production wiring.
Approved DP-014–DP-019 остаются неизменными, а TASK-026 остаётся
нереализованной. Завершённая и Coordinator-Accepted TASK-044 (2026-08-24)
исторически фиксирует `UNBLOCK TASK-026`. Superseding recheck TASK-026
определяет missing DP-015 prerequisite tracked-Start managed-parent плюс
preclaimed `StopOld` admission; TASK-047 реализует её изолированно. Fresh
reassessment TASK-026 принимает `READY — UNBLOCK TASK-026` с matrix
7/10/2/0/0/0. Текущий цикл supersedes эту readiness для live execution: repeat
Architecture Confirmation вернула `NEEDS DECISION` / `SPLIT REQUIRED`, потому
что текущий admission DP-015/DP-020 не обеспечивает replay-first inspection и
late generation allocation. TASK-026 заблокирована. TASK-049 — завершённая и
Coordinator-Accepted design-only DP-015/DP-020 refinement; её отдельная
isolated implementation prerequisite остаётся `Not Activated` без Task ID.
Этот DP остаётся Draft/Partial.

## 2. Назначение

Определить exact internal object, соединяющий immutable management scope
DP-013 с одним already-constructed scope-bound managed Flow и принятыми seams
authorization/command binding как sole lifecycle subcall future
orchestrator-owned callback closure TASK-026, не делая сам invoker callback и
без создания ещё одного Flow, lifecycle, authorization, command или
orchestration owner.

## 3. Источники полномочий

Proposal уточняет, не переопределяя:

- [ARCH-004](../architecture/ARCH-004-runtime-deployment-and-identity-model.md)
  для management authorization-before-mutation и lifecycle ownership;
- [DP-011](DP-011-runtime-launch-pipeline-integration.md) для synchronous
  managed Flow execution;
- [DP-013](DP-013-runtime-management-routing.md) для exact Target ownership и
  composition-private routing;
- [DP-014](DP-014-runtime-operational-identity-persistence.md) для facts attempt
  и execution generation;
- [DP-015](DP-015-runtime-management-command-idempotency.md) для authorization,
  claim, live callback capability, replay и unresolved barriers;
- [DP-016](DP-016-runtime-activation-replacement-rollback.md) для неизменного
  ordering activation, replacement и rollback;
- [DP-019](DP-019-runtime-activation-orchestration-prerequisites.md) для
  Approved obligations exact-scope private invocation;
- [DP-020](DP-020-runtime-orchestration-binding-sequence-readiness.md) для
  dependency-neutral binding и managed primitive/linked adapters.

Статусы и semantics источников более высокого уровня не меняются.

## 4. Область действия

Design охватывает только:

- construction одного immutable invoker вокруг одного already-constructed
  managed Flow для одного exact scope Runtime Instance;
- exact invocation validation до Owner mutation;
- synchronous delegation одному stored managed Flow;
- upstream mapping primitive Start и linked parent/`StartTarget`;
- type-correct boundary между future orchestrator-owned callback closure
  DP-015 задачи TASK-026, `StartOutcome` invoker и mapping/publication terminal
  result вне DP-021;
- custody и lifetime per-call capability;
- cancellation, failure, panic, `runtime.Goexit` и indeterminate behavior;
- направление dependencies и отсутствие legacy fallback.

## 5. Не-цели

Proposal не определяет:

- production code, exported product API, HTTP/CLI/DTO routes или transport;
- Principal model или concrete authorization policy;
- terminal publication DP-014 или terminalization command/phase DP-015;
- mapping возвращённого Owner `StartOutcome` в DP-015 `TerminalOutcome`;
- orchestrator DP-016 или selection policy activation/replacement/rollback;
- persistence, recovery, reporting, supervision или production composition;
- изменение public DP-013 `Directory.Start`, `Stop` или `Observe`;
- adoption legacy command records или unmanaged Flow execution.

## 6. Ownership и Construction

Invoker принадлежит существующей internal composition boundary
`runtimemanagement`. Его conceptual construction contract:

```text
NewManagedStartInvoker(
    OperationalDomain,
    runtimemanagement.Target,
    alreadyConstructed *runtimelaunchflow.ManagedFlow,
) -> ManagedStartInvoker, error
```

Construction проверяет:

1. `OperationalDomain` является непустым и opaque;
2. `Target` является одним valid immutable scope DP-013;
3. `alreadyConstructed` не nil.

До этого constructor composition вызывает `runtimelaunchflow.NewManaged`
ровно один раз с Owner и Loader того же Binding плюс accepted
`StartClaimContinuation`. Composition audit доказывает, что copied Target,
Owner/Loader Binding и preconstructed Flow принадлежат одному immutable scope.
Invoker не повторяет и не заменяет это construction.

При успехе invoker хранит только:

- exact immutable `OperationalDomain`;
- копию exact immutable DP-013 `Target`;
- одну borrowed immutable reference на already-constructed scope-bound managed
  Flow.

Он не принимает и не удерживает Binding, Owner, Loader, continuation,
authorization policy, command/parent/phase record, command key, permit,
`StartExecutionBinding`, rendezvous, caller context, current-operation slot,
Owner claim view, Host, Snapshot или terminal publisher. Он создаёт zero Flow,
registry entry, goroutine, detached callback или второй Flow для каждого
вызова.

Для proof construction alignment не добавляется introspection Flow, accessor,
scope token или новая identity. Этот proof принадлежит composition audit над
original inputs Binding и identities objects.

Conceptual stable errors — `ErrInvalidManagedStartInvoker` для failed
validation domain/Target/non-nil Flow при construction и
`ErrInvalidManagedStartInvocation` для validation receiver, context, request
или binding при invocation. Upstream error `NewManaged` возникает до invoker
construction и возвращается composition unchanged. Exact downstream error Flow
или Owner не конвертируется ни в один sentinel.

## 7. Единственный Invocation Contract

Invoker предоставляет одну composition-private operation:

```text
InvokeManagedStart(
    context.Context,
    runtimelifecycle.StartRequest,
    runtimeorchestrationbinding.StartExecutionBinding,
) -> runtimelifecycle.StartOutcome, error
```

Новая operation Stop, Observe, parent, phase, terminal publication или generic
Execute не добавляется.

Invoker не является callback DP-015 и никогда не возвращает DP-015
`TerminalOutcome`. Future composition TASK-026 владеет callback closure,
переданным managed primitive или linked adapter DP-015. Этот closure делает
`InvokeManagedStart` своим sole lifecycle subcall, получает exact
`StartOutcome` и error, а любое последующее mapping, terminal publication
DP-014 и terminalization command/phase DP-015 выполняет вне DP-021.

Invoker никогда не вызывает DP-015 `Boundary`, не создаёт Flow/command, не
inspect command, не map terminal result, не публикует identity и не terminalize
command/phase.

## 8. Exact Validation до Owner Mutation

До delegation stored managed Flow invocation проверяет по порядку:

1. receiver создан валидно;
2. context не nil;
3. `StartRequest` содержит ненулевые identities Workspace, Configuration и
   target ConfigurationVersion;
4. `StartExecutionBinding.Valid()` структурно равен true;
5. `OperationalDomain` authorization binding равен stored domain;
6. Workspace, Configuration и Runtime Instance binding равны stored Target;
7. Workspace, Configuration и target version `StartRequest` равны tuple
   authorization binding;
8. closed shape binding является либо primitive `ActivateExactTarget` без
   linked identity, либо exact Replace/Rollback с all-or-none derived identity
   `StartTarget`.

Любое observable mismatch возвращает exact invocation-validation error до
`Owner.PrepareStart`, Load, Build, Launcher, Host, write DP-014 или иной
lifecycle mutation. Validation не inspect mutable aggregate/command state и не
может retarget invocation.

Invoker не может и не должен introspect preconstructed Flow для повторного
proof его Owner или Loader. Exact alignment Flow-to-Target является preceding
composition audit obligation section 6; добавление accessor или validation
token расширило бы contract managed Flow и запрещено.

Structural `Valid()` доказывает только well-formed immutable fields и их
cross-field shape. Он не является proof live permit, rendezvous,
callback-authority, generation или custody. Invoker не может отличить fresh
callback-delivered binding от retained structurally valid value.

## 9. Linearization Cancellation

Nil context невалиден и даёт zero mutation.

Non-nil context никогда не отклоняется invoker только потому, что
`ctx.Err()` уже не nil. После входа DP-015 в newly claimed future
orchestrator-owned callback closure TASK-026 и его sole lifecycle subcall
invoker обязан вызвать managed Flow с этим context. Flow владеет pre-Owner cancellation
check и обязательным exact signal `StartNoClaim` command-owned rendezvous.
Early return invoker потерял бы этот signal и мог бы оставить independently
admitted Stop заблокированным.

После authentic Owner claim существующий managed Flow сохраняет context values,
игнорируя caller cancellation/deadlines через `context.WithoutCancel`; invoker
не добавляет cancellation authority и не меняет это behavior.

## 10. Authorization и Command Ordering

Authorization не принадлежит invoker.

Для каждой initial, in-progress и replay submission DP-015 вычисляет exact
six-field request `AuthorizeOrchestration` до inspection или mutation команды.
Только newly committed claim передаёт future orchestrator-owned callback
closure полный binding. Closure вызывает invoker один раз как sole lifecycle
subcall. Invoker проверяет принятый tuple против stored scope; он не вызывает policy второй раз,
не cache authority, не интерпретирует denial заново и не авторизует linked
phase независимо от accepted parent.

Поэтому authorization-before-mutation является compositional:

```text
DP-015 validate + authorize + pre-claim cancellation
    -> claim and orchestrator-owned callback closure, only for a new claim
        -> DP-021 exact scope/request validation, sole lifecycle subcall
        -> ManagedFlow.StartManaged
        -> Owner.PrepareStart
        -> exact StartOutcome/error back to the closure
    -> closure-owned mapping/publication/terminalization outside DP-021
```

Replay и in-progress observations не получают callback closure и поэтому не
могут вызвать invoker через этот protocol.

## 11. Primitive и Linked Call Paths

### 11.1 Primitive Start

Future composition TASK-026 передаёт orchestrator-owned callback closure в
`Boundary.ExecuteManagedStart`. DP-015 создаёт primitive binding с
`ActivateExactTarget`, без identity parent/phase, с одной exact aggregate
revision, execution generation и live rendezvous. Closure вызывает один
scope-bound invoker ровно один раз как sole lifecycle subcall, получает exact
`StartOutcome`/error и выполняет terminal mapping/publication вне DP-021.

### 11.2 Replace/Rollback StartTarget

Composition сначала входит в `Boundary.ExecuteManagedParent`, затем использует
только его callback-scoped `ManagedParentExecution` и
`ContinueOrExecuteManagedStartTarget`. DP-015 выводит exact identities parent и
ordinal-one `StartTarget` и создаёт linked binding. Future orchestrator-owned
callback closure этой phase вызывает ту же scope-bound operation
`InvokeManagedStart` как sole lifecycle subcall, затем владеет result
mapping/publication вне DP-021.

Invoker не содержит ветки primitive-versus-parent, кроме validation closed
shape binding. Оба path converge к одному contract exact Target, Start request,
binding и managed Flow.

Legacy `Execute`, `ExecuteParent`, `ContinueOrExecuteStartTarget`, public
`Directory.Start` и unmanaged `Flow.Start` нельзя adopt, upgrade или
использовать как fallback.

## 12. Lifecycle Per-call Capability

`StartExecutionBinding` является immutable structural value, переданным DP-015
future orchestrator-owned callback closure на original synchronous
permit-holding stack. Closure lends его invoker для sole lifecycle subcall;
invoker передаёт value unchanged в `StartManaged`.

Invoker:

- никогда не хранит, не копирует в long-lived state, не cache, не index, не
  заменяет и не reuse binding или rendezvous;
- не может manufacture permit или resolve opaque rendezvous;
- возвращается до expiry callback capability DP-015;
- не раскрывает accessor binding, Flow, continuation, Owner или Target,
  расширяющий custody.

Stored Flow reference borrowed immutably на lifetime одной scope composition.
Invoker не owns, close, reconstruct, swap или duplicate её. Composition
сохраняет preconstructed Flow live на весь lifetime invoker и retire их вместе
с immutable scope.

Return, panic, `runtime.Goexit`, loss generation Boundary или callback expiry
expires live permit, rendezvous lookup и callback authority во владении DP-015.
Это не mutate и не invalidate immutable value `StartExecutionBinding`; value
может оставаться structurally `Valid()`. Поэтому no-reuse является invariant
callback custody/no-bypass, а не invoker-enforced liveness check. Caller,
retaining value, нарушает contract, и invoker не может отличить его от fresh
structurally valid binding.

## 13. Направление Dependencies

Направление construction:

```text
runtimemanagement composition Binding(Target, Owner, Loader)
    -> runtimelaunchflow.NewManaged(Owner, Loader, continuation), exactly once
    -> composition audit of the same Binding/Target/object identities
    -> NewManagedStartInvoker(domain, Target, preconstructed ManagedFlow)
```

Направление invocation:

```text
runtimecommandidempotency managed adapter
    -> future TASK-026 orchestrator-owned callback closure
        -> runtimemanagement private invoker, sole lifecycle subcall
            -> runtimelaunchflow.ManagedFlow
            -> exact StartOutcome/error back to closure
        -> mapping/publication/terminalization outside DP-021
```

`runtimemanagement` может зависеть от `runtimelaunchflow`,
`runtimelifecycle` и dependency-leaf `runtimeorchestrationbinding`. Для
реализации invoker он не импортирует `runtimecommandidempotency`,
`runtimeidentity`, transport или future orchestrator. Invoker никогда не
вызывает вверх command boundary. Dependency-leaf package binding продолжает не
зависеть ни от одного higher package. `runtimelaunchflow` не зависит обратно от
`runtimemanagement`; preconstructed reference не создаёт cycle.

## 14. Results, Failures, Panic и Indeterminate Outcomes

Design различает:

- **upstream managed Flow construction error** — возвращается composition
  unchanged до вызова constructor invoker;
- **invoker construction error** — invalid domain, Target или nil
  preconstructed managed Flow; invoker не возвращается и Flow не создаётся;
- **invocation-validation error** — nil context или mismatch exact
  scope/request/binding; zero Owner mutation;
- **managed Flow outcome/error** — invoker возвращает без изменения;
- **panic validation dependencies или Flow** — не конвертируется в success;
  он проходит через future orchestrator-owned closure до существующей
  panic-safe callback boundary DP-015, оставляющей command/phase unresolved и
  expiring permit/rendezvous/callback authority;
- **`runtime.Goexit`** — unwinds synchronous closure; deferred expiry DP-015
  удаляет live permit/rendezvous/callback authority и оставляет durable work
  unresolved;
- **indeterminate callback return или missing terminal publication** — остаётся
  indeterminate в DP-015 без retry, adoption или fallback.

Invoker не wrap/relabel exact downstream Owner outcome/error, не возвращает
`TerminalOutcome` и не выводит terminal command truth. Он возвращает exact
`StartOutcome`/error future orchestrator-owned closure TASK-026. Этот closure
владеет mapping в terminal outcome DP-015, terminal publication DP-014 и
terminalization command/phase DP-015 вне DP-021.

## 15. Ограничение Privacy и Custody

Repository visibility `internal` и composition-private constructor являются
encapsulation, а не authentication boundary. Invoker не может доказать, что
caller имеет current callback authority DP-015, потому что structural validity
binding намеренно не несёт proof live permit. Поэтому security и at-most-once
proof зависят от custody:

- production composition создаёт один managed Flow ровно один раз из scope
  Binding, audit тот же Owner/Loader/Target, затем создаёт invoker с этим
  preconstructed Flow;
- только future orchestrator-owned callback closure TASK-026, получающий fresh
  binding DP-015, может вызвать invoker ровно один раз как sole lifecycle
  subcall;
- ни transport, public method Directory, registry, service locator, long-lived
  command record, ни recovery reconstruction не раскрывает invoker;
- reconstruction восстанавливает durable records, но не callback authority для
  invocation invoker.

Если production composition не может доказать эту custody и отсутствие bypass,
integration fail closed и TASK-026 остаётся blocked.

## 16. Acceptance Proofs

Последующая implementation должна доказать все 18 строк:

1. composition вызывает `NewManaged` ровно один раз с тем же Binding
   Owner/Loader/Target, возвращает error unchanged и не создаёт duplicate или
   unmanaged Flow до construction invoker;
2. empty domain, invalid Target или nil preconstructed Flow возвращает
   invoker-construction error; success хранит только copied domain/Target и
   одну borrowed immutable Flow reference без per-call fact;
3. nil context или любой structural mismatch domain, Target, request,
   authorization, action или linked identity fail до Owner mutation и даёт
   zero Load, Build, Launcher, Host, DP-014 или terminal mutation;
4. `StartExecutionBinding.Valid()` является только structural и никогда не
   считается proof live permit, rendezvous, callback, generation или custody;
5. DP-015 авторизует каждую initial/replay submission до inspection, передаёт
   только newly committed claim future callback closure TASK-026, а invoker не
   делает second policy call;
6. primitive adapter вызывает future orchestrator-owned closure, sole
   lifecycle subcall которого достигает stored Flow один раз через invoker с
   unchanged request и binding;
7. linked parent/`StartTarget` adapter вызывает future orchestrator-owned phase
   closure, sole lifecycle subcall которого достигает того же invoker/Flow один
   раз с unchanged request и binding;
8. invoker никогда не является callback DP-015, не вызывает Boundary и
   возвращает `StartOutcome`/error, а не `TerminalOutcome`;
9. exact `StartOutcome` и error identities Flow возвращаются unchanged через
   invoker future orchestrator-owned closure;
10. mapping в `TerminalOutcome`, terminal publication DP-014 и terminalization
    command/phase DP-015 выполняются только этим closure вне DP-021;
11. already-cancelled non-nil context достигает Flow и создаёт exact
    `StartNoClaim`; post-Owner-claim cancellation сохраняет existing semantics
    continuation и Owner convergence;
12. in-progress, replay, legacy и reconstructed records не получают callback
    closure и поэтому не получают authorized invocation invoker;
13. closure lends immutable binding на один synchronous lifecycle subcall;
    invoker никогда не хранит, не index и не раскрывает его или rendezvous;
14. callback return, panic, `runtime.Goexit` или generation loss expires live
    permit, rendezvous lookup и callback authority, но не mutate/invalidate
    structurally valid binding value;
15. reuse retained structurally valid binding запрещён callback custody и
    no-bypass composition; invoker не может detect liveness или отличить
    retained от fresh structural value;
16. panic, `runtime.Goexit`, error и indeterminate outcomes не становятся
    success, duplicate work, retry, adoption или legacy/unmanaged fallback и
    оставляют unresolved work по существующим rules DP-015;
17. imports construction/invocation остаются acyclic; invoker не создаёт Flow и
    напрямую не вызывает DP-015, DP-014, transport или orchestrator;
18. не добавляется public/private bypass, accessor/token, second policy check,
    terminal owner или production capability, а EN/RU mirrors остаются
    семантически равными.

## 17. Граница Implementation

Implementation Status — Partial, реализован изолированно. TASK-043 добавляет
concrete `ManagedStartInvoker`, `NewManagedStartInvoker`,
`InvokeManagedStart`, `ErrInvalidManagedStartInvoker` и
`ErrInvalidManagedStartInvocation` в существующий package
`internal/runtimemanagement`. Focused proofs покрывают validation constructor и
invocation, exact primitive/linked delegation, передачу already-cancelled
context и unchanged identity downstream outcome/error. Invoker хранит только
copied domain/Target и один borrowed preconstructed managed Flow и не дублирует
construction Flow.

Эта isolated implementation не предоставляет future callback closure DP-015,
integration callback custody/replay, terminal result mapping, terminal
publication DP-014, terminalization command/phase DP-015, orchestrator
DP-016/TASK-026, production composition audit или production wiring.

TASK-043 завершена как `Completed — Coordinator Accepted (2026-08-21)` и не
активирует следующую task. TASK-044 исторически фиксирует `UNBLOCK TASK-026`;
superseding recheck TASK-026 подтверждает missing DP-015 prerequisite
tracked-Start managed-parent плюс preclaimed `StopOld` admission. TASK-046 позже
фиксирует этот contract, TASK-047 реализует его изолированно, а fresh
reassessment принимает `READY — UNBLOCK TASK-026` как historical evidence.
Repeat Architecture Confirmation теперь блокирует TASK-026 отдельной
DP-015/DP-020 refinement replay-first admission и late generation. Design
refinement завершена как TASK-049 и принята Coordinator 2026-08-28; её
отдельная isolated implementation prerequisite остаётся `Not Activated` без
Task ID; implementation TASK-026
здесь не утверждается.

## 18. Решение

UWP использует один immutable scope-bound managed Start invoker во владении
существующей internal management composition DP-013. Composition заранее
создаёт один managed Flow ровно один раз и передаёт его borrowed immutable
reference вместе с copied domain/Target в invoker. Invoker создаёт zero Flow,
валидирует exact domain, Target, request и structural binding, затем синхронно
один раз делегирует этому Flow как sole lifecycle subcall future
orchestrator-owned callback closure TASK-026. Exact `StartOutcome` и error
возвращаются этому closure; mapping, terminal publication и terminalization
остаются вне DP-021. Invoker не является callback DP-015 и не может доказать
live callback authority по structurally valid binding. DP-015 сохраняет
ownership authorization, command, replay, permit, rendezvous и callback
authority; expiry callback не mutate и не invalidate value binding. Owner
сохраняет lifecycle authority; no-reuse опирается на custody closure и
no-bypass composition.

В contract отсутствуют second policy check, command owner, lifecycle owner,
public route, detached work, terminal publication и legacy fallback.

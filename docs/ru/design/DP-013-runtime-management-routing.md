# DP-013: Маршрутизация управления Runtime

[English version](../../en/design/DP-013-runtime-management-routing.md)

## 1. Статус

- **Design Status:** Draft
- **Implementation Status:** Start/Stop/Observe implemented in isolation; C2-Exposure раздел 32 Planned
- **Integration Readiness:** Blocked; full integration и Production Activation
  требуют Approved/Planned prerequisites DP-019 и wiring ниже

До утверждения proposal остаётся ненормативным. Его изолированная in-process
management command boundary реализована в `internal/runtimemanagement`.
Production routing Control Service, concrete authorization policy,
external/process-restart persistence, recovery и Production Activation
остаются отсутствующими.

## 2. Назначение

Определить минимальную boundary, маршрутизирующую авторизованные команды
Start, Stop и Observe ровно в один immutable scope Runtime Instance,
содержащий ровно один Runtime Lifecycle Owner и один Runtime Launch Flow.

## 3. Источники полномочий

Proposal уточняет, но не переопределяет:

- [ADR-0002](../adr/0002-configuration-dsl.md);
- [ADR-0003](../adr/0003-runtime-architecture.md);
- [ADR-0004](../adr/0004-handshake-runtime-dependencies.md);
- [ARCH-002](../architecture/ARCH-002-runtime-foundation-freeze.md);
- [ARCH-004](../architecture/ARCH-004-runtime-deployment-and-identity-model.md);
- [ARCH-005](../architecture/ARCH-005-runtime-configuration-snapshot-and-loading-model.md);
- [DP-010](DP-010-runtime-lifecycle-owner-contract.md);
- [DP-011](DP-011-runtime-launch-pipeline-integration.md);
- [DP-012](DP-012-runtime-source-composition.md).
- [DP-021](DP-021-private-exact-scope-managed-start-invoker.md) для exact
  composition-private managed Start invocation contract и partial isolated
  implementation.

Accepted ADR, Frozen foundation и Active architecture сохраняют приоритет.

## 4. Область

Design охватывает один process-local command directory, immutable bindings
Runtime Instance, exact identity routing, policy-neutral authorization seam,
порядок authorization-before-mutation, concurrency, cancellation, сохранение
failures, composition audit и isolated proof evidence.

Он не определяет transport resources или data transfer objects.

## 5. Package и ответственность

Изолированный package — `internal/runtimemanagement`.

Он владеет только:

- validation management targets;
- exact lookup одного immutable binding Runtime Instance;
- authorization admission для Start, Stop и Observe;
- delegation exact Flow или Owner этого binding.

Он не владеет Runtime lifecycle state, загрузкой Configuration,
конструированием Snapshot, Host resources, authorization policy, persistence,
recovery или transport mapping.

## 6. Термины

### Target

Immutable assertion identities Workspace, Configuration и Runtime Instance.

### Binding

Opaque immutable construction input, содержащий один Target, один Owner и один
Loader. Caller никогда не передаёт Flow.

### Scope

Private accepted entry Directory, содержащая один Target, Owner из Binding и
один Flow, сконструированный Directory из того же Owner и Loader.

### Directory

Единственная focused management command boundary. Её private map immutable
после construction и разрешает только scopes Runtime Instance. Она не
является dynamic registry, generic resolver или service locator.

## 7. Точная public surface

```go
package runtimemanagement

var (
    ErrInvalidBinding          error
    ErrInvalidDirectory        error
    ErrInvalidContext          error
    ErrInvalidTarget           error
    ErrRuntimeInstanceNotFound error
)

type Action string

const (
    ActionStart   Action = "start"
    ActionStop    Action = "stop"
    ActionObserve Action = "observe"
)

type Target struct {
    // private immutable identities
}

func NewTarget(
    workspaceID uint64,
    configurationID uint64,
    runtimeInstanceID runtimeconfigload.RuntimeInstanceID,
) (Target, error)

func (t Target) WorkspaceID() uint64
func (t Target) ConfigurationID() uint64
func (t Target) RuntimeInstanceID() runtimeconfigload.RuntimeInstanceID

type Authorize func(
    context.Context,
    Action,
    Target,
    uint64,
) error

type Binding struct {
    // private immutable Target, Owner, and Loader
}

func NewBinding(
    target Target,
    owner *runtimelifecycle.Owner,
    loader *configurationloader.Loader,
) (Binding, error)

type Directory struct {
    // private immutable Runtime Instance scope map
}

func NewDirectory(
    authorize Authorize,
    bindings ...Binding,
) (*Directory, error)

func (d *Directory) Start(
    ctx context.Context,
    target Target,
    configurationVersionID uint64,
) (runtimelifecycle.StartOutcome, error)

func (d *Directory) Stop(
    ctx context.Context,
    target Target,
) (runtimelifecycle.StopOutcome, error)

func (d *Directory) Observe(
    ctx context.Context,
    target Target,
) (runtimelifecycle.Observation, error)
```

Новые exported command DTO, interface, option, mutable scope, registration
method, list operation, replacement operation или lifecycle declaration не
добавляются.

## 8. Sentinel identities

Точные sentinels и strings:

```go
var (
    ErrInvalidBinding = errors.New(
        "invalid Runtime management binding",
    )
    ErrInvalidDirectory = errors.New(
        "invalid Runtime management directory",
    )
    ErrInvalidContext = errors.New(
        "invalid Runtime management context",
    )
    ErrInvalidTarget = errors.New(
        "invalid Runtime management target",
    )
    ErrRuntimeInstanceNotFound = errors.New(
        "Runtime Instance not found",
    )
)
```

Callers используют `errors.Is` для этих identities. Boundary не добавляет
authorization, persistence, retry, transport или diagnostics sentinel.

## 9. Target

`NewTarget` требует ненулевые IDs Workspace и Configuration и непустой ID
Runtime Instance. Failure возвращает bare `ErrInvalidTarget`.

Поля Target private и раскрываются только value accessors. Zero value
недействителен. Target не содержит ConfigurationVersion, Launch Attempt,
Host, PID, Listener, Principal, policy или transport identity.

## 10. Policy-neutral authorization seam

`Authorize` является named function type, а не interface. Nil function
непосредственно обнаруживается без reflection или typed-nil detector.

Аргументы:

- caller context;
- exact `ActionStart`, `ActionStop` или `ActionObserve`;
- verified Target;
- exact ID ConfigurationVersion для Start либо zero для Stop и Observe.

Non-nil authorization error возвращается как тот же error interface без
изменения и wrapping. Directory не выполняет text inspection, normalization,
logging, `errors.Join`, transport mapping или recovery.

Panic из non-nil authorization function является composition или policy
defect. Directory не выполняет recover. Lifecycle mutation ещё не началась,
поскольку authorization предшествует delegation.

Этот seam определяет ordering, а не concrete Principal model или policy.

Directory borrow-ит одно immutable value function `Authorize` на весь свой
lifetime. Function и каждая captured dependency должны быть безопасны для
concurrent invocation. Directory может вызывать её concurrently для одного
или разных Targets и не добавляет synchronization вокруг неё.

Conforming Authorize разрешает разным Targets Runtime Instance выполняться
независимо. Он может кратко синхронизировать internal policy state, но не
удерживает cross-scope или global lock во время external I/O, blocking waits
или callback execution. Вызовы одного Target могут пересекаться; каждый
result принадлежит только своему invocation и не кэшируется как lifecycle
authority.

## 11. Construction Binding

`NewBinding` принимает Target, Owner и Loader, но никогда Flow.

Перед принятием Binding он:

1. проверяет Target;
2. отклоняет nil Owner или Loader;
3. читает `Owner.Observe()` без mutation;
4. проверяет exact identities Workspace, Configuration и Runtime Instance.

Любой failure возвращает bare `ErrInvalidBinding`. Construction Binding не
выполняет construction Flow, loading, building, launch, Stop, authorization,
создание goroutine или repository access.

## 12. Static binding Owner-to-Flow

`NewDirectory` сначала проверяет полный набор Binding. Только после успешной
проверки каждого Binding он вызывает:

```go
runtimelaunchflow.New(binding.owner, binding.loader)
```

ровно один раз для каждого принятого Binding. Полученный Flow сохраняется в
том же private scope, что и exact Owner.

Поскольку callers не могут передать Flow, public API не может связать Flow с
другим Owner. Flow introspection, новый Flow accessor и изменения существующих
packages не требуются.

Constructor может предотвратить duplicate IDs Runtime Instance и duplicate
Owner pointers внутри одного Directory. Он не может обнаружить отдельно
сконструированный bypass Flow или другой Directory в иной composition. Это
обязанности Composition Audit, а не причины добавлять global state.

## 13. Construction Directory

`NewDirectory` отклоняет с bare `ErrInvalidDirectory`:

- nil `Authorize`;
- zero Bindings;
- zero или invalid Binding;
- duplicate ID Runtime Instance;
- reuse одного Owner pointer более чем в одном Binding;
- невозможный failure construction Flow после validation.

Все Bindings проверяются до construction любого Flow. Partial Directory не
возвращается. Construction не создаёт lifecycle state, command, Launch
Attempt, Runtime resource, goroutine, cache или background work.

Constructor копирует принятые routing entries в private map, которая больше
не изменяется.

Принятое value Authorize также immutable после construction. Directory не
заменяет, не оборачивает, не сериализует и иначе не управляет им.

## 14. Exact command targets

Start принимает Target и один non-zero exact ID ConfigurationVersion.

Stop и Observe принимают только Target. Они не принимают и не выводят version.

Commands никогда не используют ConfigurationVersion как execution identity,
Host pointer, PID, Listener address, Session, goroutine, context или transport
resource. Latest/current selection, replacement identity и fallback scope
отсутствуют.

## 15. Validation и error precedence

Каждая command применяет следующий precedence:

1. nil или invalid receiver Directory возвращает bare
   `ErrInvalidDirectory`;
2. nil context возвращает bare `ErrInvalidContext`;
3. invalid Target, а для Start также zero version, возвращает bare
   `ErrInvalidTarget`;
4. уже видимый `ctx.Err()` возвращает этот exact context error;
5. exact lookup использует ID Runtime Instance;
6. отсутствующий Runtime Instance или mismatch Workspace/Configuration
   возвращает один и тот же bare `ErrRuntimeInstanceNotFound`;
7. `Authorize` вызывается ровно один раз;
8. только nil authorization result разрешает downstream delegation.

Lookup и identity comparison являются read-only и не выполняют lifecycle
mutation. Нормализация absence и mismatch не создаёт отдельное различие
identity oracle в этой boundary.

## 16. Start

После успешных validation, routing и authorization Start вызывает только:

```go
scope.flow.Start(
    ctx,
    runtimelifecycle.NewStartRequest(
        scope.target.WorkspaceID(),
        scope.target.ConfigurationID(),
        configurationVersionID,
    ),
)
```

Exact requested version сохраняется. Directory не вызывает напрямую
`Owner.PrepareStart`, `Owner.Start`, Loader, Builder, `runtime.Launch`,
`runtime.Bootstrap` или `Host.Start`.

Возвращённые `StartOutcome` и error возвращаются без изменения.

Реализованная изолированно orchestration DP-016 и planned recovery DP-017
сохраняют exported surface `Directory.Start` без изменения и используют один
private management-only seam Start-claim continuation в stored Flow. Exact
management composition снабжает
continuation borrowed capabilities coordination pending Stop DP-015 и
conditional execution binding DP-014, а также opaque Control Service execution
generation. Directory не allocate generation, не обращается прямо к
persistence и не exposes capabilities.

После claim exact attempt Owner и до Load Flow seam:

1. сначала resolves already admitted pending Stop;
2. иначе conditionally bind exact attempt/expected aggregate revision к exact
   generation через DP-014;
3. после confirmed binding атомарно упорядочивает final Stop claim и release
   `Continue` в Flow.

Pending Stop сигнализирует Owner claim original blocked Stop call stack; только
тот claimant вызывает exact `scope.owner.Stop` этого binding, публикует result
и возвращает `StopConverged`. Confirmed binding с final release возвращает
`Continue`. Только coherently proven absence binding для exact still-active
attempt на expected revision входит в тот же final gate против pending Stop и
может вернуть `BindingFailed` без publication lifecycle/command fact. Different
generation, stale/conflicting/inactive facts, unavailable state или unknown
перечитывается и converge к exact existing terminal outcome либо возвращает
`Blocked`; это никогда не становится BindingFailed. Flow затем с authentic
preparation token вызывает existing Owner.Start с
`FailedPreparation`; mutex Owner упорядочивает failure и later Stop. Только
exact returned Owner outcome и confirmed terminal publication DP-014 разрешают
terminalization command/phase. Permit loss, unproven Stop convergence,
indeterminate binding/terminal publication или unknown exact inspection
возвращает `Blocked`, не разрешает preparation work и закрывает linked barrier
DP-015.

Command-admission decisions происходят до wait обоих call stacks. Continuation
не несёт command/recovery permit или caller context, и admission/Owner lock не
удерживается во время persistence, signal, result wait или convergence Owner.
Stop, проигравший final release gate, использует обычный route section 17 и
достигает already-claimed attempt. Binding DP-013 передаёт internal-package-
callable `StartClaimContinuation` при construction Flow; seam не добавляет
exported operation Directory/Replace/Rollback, не передаёт mutable
`LaunchPreparation`. Managed continuation реализована и независимо принята
изолированно за пределами этого package.

Approved [DP-019](DP-019-runtime-activation-orchestration-prerequisites.md)
закрывает design ambiguity этой seam. Он добавляет отдельный exact
orchestration authorization tuple и private composition-only lifecycle invoker,
используемый только original primitive/phase permit holder. Invoker передаёт
один per-call immutable `StartExecutionBinding` в managed Start surface одного
уже stored Flow; revision, generation, authorization, linked execution и
rendezvous facts никогда не являются Flow-construction state или mutable slot
Directory. Existing public `Start`, `Stop`, `Observe` и их authorization
behavior не меняются. Managed command, binding, Flow и continuation surfaces
реализованы в своих focused internal packages; TASK-043 теперь реализует
соединяющий scope-bound Flow surface composition-private invoker DP-013
изолированно.
Draft [DP-021](DP-021-private-exact-scope-managed-start-invoker.md) теперь
фиксирует ownership и exact contract этого invoker: existing
composition `runtimemanagement` создаёт один scope-bound managed Flow ровно
один раз, затем передаёт invoker copied domain/Target и borrowed immutable
reference на этот preconstructed Flow. Invoker раскрывает только synchronous
`InvokeManagedStart`, создаёт zero Flow, не
выполняет second policy authorization, не хранит per-call capability, никогда
не вызывает command Boundary и не добавляет public operation Directory.
TASK-043 реализует этот exact invoker изолированно. TASK-026 теперь реализует
callback custody, terminal work и orchestration изолированно; production
composition и wiring остаются отсутствующими.

Если linked path `Directory.Start` возвращает definitive cancellation/error до
claim Owner, он сигнализирует `StartNoClaim` original pending Stop call stack.
Только тот claimant terminalizes свой Stop satisfied без invocation section 17.
Indeterminate return или lost signal даёт `Blocked`, а не `StartNoClaim`.

## 17. Stop

После успешных validation, routing и authorization Stop вызывает только:

```go
scope.owner.Stop(ctx)
```

Он не вызывает Flow, не выбирает attempt, не повторяет cleanup, не добавляет
timeout и не становится другим shutdown owner. Возвращённые `StopOutcome` и
error возвращаются без изменения.

## 18. Observe

Observe проходит authorization gate, хотя не выполняет lifecycle mutation.
После authorization он выполняет одну финальную проверку `ctx.Err()`. При nil
он вызывает:

```go
scope.owner.Observe()
```

ровно один раз и возвращает coherent immutable Observation. Он не раскрывает
Host, context, cancellation, raw failure, payload Configuration или Secret.

## 19. Authorization-before-mutation

Ни один lifecycle method Flow или Owner не вызывается до успешной
authorization.

Invalid input, missing или mismatched identity, context cancellation,
authorization denial и authorization failure поэтому приводят к zero
lifecycle mutation. Directory не кэширует authorization result и не
превращает его в продолжающуюся lifecycle authority.

Policy revocation, authorization audit storage и concealment в конкретном
transport являются отдельными contracts.

Nil authorization return является только per-command admission. Он не
является lifecycle linearization point: Flow и Owner сохраняют все
linearization claim, conflict, operation и outcome.

## 20. Concurrency и linearization

Directory не добавляет per-Instance mutex. Owner остаётся единственной
serialization boundary для state-changing operations одного Runtime Instance.

- concurrent Start commands достигают только того же Flow; его Owner
  арбитрирует их, принимает не более одного claim и operation и сохраняет
  существующие conflict или outcome semantics для проигравших calls;
- concurrent Stop commands достигают только того же Owner;
- Observe читает один coherent snapshot Owner;
- вызовы Authorize для одного или разных Targets могут пересекаться;
- разные immutable scopes выполняются независимо, когда borrowed Authorize
  соответствует contract cross-scope progress.

Private map Directory read-only после construction. Ни один lock Directory не
удерживается во время authorization, Flow, Owner, resource work или waiting.
Directory не содержит mutex, queue, semaphore или authorization goroutine.
Один blocked authorization call блокирует только своего caller и не мешает
другому scope войти в Directory или conforming Authorize.

## 21. Cancellation

Cancellation, видимая до authorization, возвращает exact context error без
lifecycle mutation. Authorization получает original context.

Conforming Authorize наблюдает cancellation для своей blocking work и
возвращается без detached work. Directory не может принудительно прервать
function, игнорирующую context, и не скрывает это ограничение за goroutine или
timeout.

После authorization Start передаёт тот же context Flow, чей Caller
Cancellation Gate остаётся authoritative. Directory не прерывает, не
отделяет и не переопределяет Start operation после победы этого gate.

Stop передаёт тот же context Owner, чьи locked cancellation gate и wait
semantics остаются authoritative.

Observe выполняет описанную выше финальную post-authorization проверку
context. Directory не создаёт substitute context, timeout, goroutine, channel
или detached operation.

Если cancellation конкурирует с nil authorization return, original context
достигает существующего gate Flow или Owner; Observe использует финальную
проверку. Gate, наблюдающий cancellation, определяет возможность начала
lifecycle mutation точно по DP-010 и DP-011.

## 22. Outcomes и failure boundaries

Directory владеет только объявленными validation и routing errors.
Authorization errors проходят без изменения. Существующие method errors Flow
и Owner, Start outcomes, Stop outcomes и values Observation также проходят
без изменения.

Authorization error влияет только на свой invocation и приводит к zero
downstream calls. Panic является policy или composition defect до lifecycle
delegation: Directory не выполняет recovery, panic распространяется в caller
goroutine, а methods Flow или Owner не вызываются.

Directory не объединяет и не переклассифицирует:

- preparation, Loader или Builder failure;
- Bootstrap или startup failure;
- Start conflict или attempt-ID failure;
- Stop failure или retained Host ownership;
- caller cancellation после downstream gate.

Невозможная post-routing identity divergence является composition defect. Она
не запускает fallback к другому scope.

## 23. Ownership и lifetime

| Value или resource | Owner |
| --- | --- |
| Binding declaration | Composition root до construction Directory |
| Private routing map и scopes | Directory |
| Authorization policy/function | External composition; borrowed Directory |
| Runtime Instance и Launch Attempts | Owner scope |
| Start pipeline | Flow scope |
| Loader и Source access | Configured Loader |
| Runtime Host reference | Owner scope |
| Runtime resources | Runtime Host |
| Caller wait | Individual command invocation |

Owner, Loader, Authorize и все dependencies, captured Authorize, должны жить
дольше Directory. Captured policy state остаётся во внешнем владении и
сохраняет contract concurrent safety и cross-scope progress. Directory не
владеет shutdown hook и не передаёт Runtime ownership.

## 24. Composition Audit

До construction Directory и повторно до Production Activation composition
evidence должна доказать:

1. один immutable scope на Runtime Instance;
2. один Owner на scope и отсутствие reuse Owner между scopes;
3. один Flow, сконструированный Directory из Owner и Loader этого scope;
4. отсутствие отдельно сконструированного или bypass Flow;
5. все вызовы Start используют Directory и stored Flow;
6. все вызовы Stop и Observe используют Directory и stored Owner;
7. authorization является единственным admission path до delegation;
8. Authorize и его captured state concurrent-safe и не содержат cross-scope
   lock, удерживаемый во время blocking work;
9. отсутствуют package-global directory, dynamic writer, service locator,
   importer или alternate management path.

Audit является evidence reference graph и ownership. Он не является runtime
detector, registry, reflection check, global lock или persistence substitute.

## 25. Dependency direction

Разрешённое направление:

```text
future management composition
    -> runtimemanagement
        ├── runtimelaunchflow
        │   ├── runtimelifecycle
        │   └── configurationloader
        ├── runtimelifecycle
        ├── configurationloader
        └── runtimeconfigload
```

Нижележащие packages Runtime, Flow, Owner, Loader, Source и Host не импортируют
`runtimemanagement`. Directory никогда не передаётся Runtime и не используется
как generic dependency container.

## 26. Обязательные prerequisites implementation

Draft design contract реализован в bounded isolated slice. Approved
DP-014–DP-018 закрывают все focused design gates ARCH-004 sections 19(2)–(6).
Full integration остаётся blocked до появления required implementations
external durability, activation, recovery и reporting и композиции isolated
routing, identity и command packages через explicit Control Service boundary.

Loader, provenance Snapshot и schema compatibility уже разрешены ARCH-005 и
DP-007 — DP-012. Прецедент isolated implementation DP-010 — DP-012 не создаёт
исключение из gate ARCH-004 с более высоким status. Approved DP-014–DP-018
определяют dependency-ordered contracts. DP-014–DP-016 теперь предоставляют
isolated process-local packages; external adapters, schemas, wiring и
implementations DP-017/DP-018 отсутствуют.

## 27. Proofs изолированной реализации

Bounded isolated implementation доказывает следующий local contract. Full
integration должна повторить применимые composition proofs после реализации
каждой required dependency section 26:

1. exact exported surface и sentinel strings;
2. validation Target и immutable accessors;
3. nil authorization отклоняется без reflection;
4. Binding отклоняет nil dependencies и mismatch identity Owner;
5. Directory отклоняет zero, duplicate и reused-Owner bindings до
   construction Flow;
6. Directory конструирует ровно один Flow из каждого принятого Owner и Loader;
7. invalid, missing, mismatched, denied, failed или cancelled commands
   выполняют zero downstream lifecycle calls;
8. authorization происходит ровно один раз до delegation Start, Stop и
   Observe;
9. Start сохраняет exact version и ровно один раз вызывает только exact stored
   Flow;
10. Stop ровно один раз вызывает только exact stored Owner;
11. Observe авторизован, coherent и non-mutating;
12. authorization и downstream errors сохраняют exact interface identity и
    cause chains, а каждый authorization error влияет только на один
    invocation;
13. panic authorization распространяется без recovery и выполняет zero
    downstream calls;
14. concurrent calls Authorize одного и разных Targets race-safe; один blocked
    scope не мешает другому conforming scope выполнить authorization и
    delegation;
15. cancellation до и во время Authorize не создаёт detached work, а nil
    return, конкурирующий с cancellation, остаётся под управлением
    существующего gate Flow/Owner или финального gate Observe;
16. lifecycle concurrency одного scope остаётся под управлением Owner, а
    разные conforming scopes выполняются независимо;
17. отсутствуют fallback, latest selection, retry, re-read, detached work,
    новый mutex, dynamic registration или package-global state;
18. Composition Audit доказывает отсутствие bypass path или cross-scope
    serialization authorization;
19. targeted, stress, race, dependency и full repository checks проходят при
    технической доступности.

Tests могут использовать local fakes или package-private seams. Они не
разрешают wiring Control Service или Runtime activation.

## 28. Явно отложено

Focused design gates section 19 закрыты. Для operational identity DP-014 и
command idempotency DP-015 теперь существуют isolated process-local in-memory
implementations. До integration требуются следующие production dependencies:

- external durable operational identity storage и его integration adapter;
- external durable management command storage и его integration adapter;
- orchestration activation, replacement и rollback;
- recovery и reconciliation;
- operational error reporting и redaction.

Дополнительные concerns, отложенные за пределы этого design:

- HTTP paths, JSON DTO, status codes и transport error concealment;
- authentication transport, representation Principal и concrete
  authorization policy;
- create, delete, list или rebinding Runtime Instance;
- implementation persistence, schema, transactions и hydration после её
  focused design;
- retry, automatic restart, reload и process supervision сверх обязательных
  focused contracts;
- metrics, audit storage и alerting сверх обязательного reporting и redaction
  design;
- application production wiring и Production Activation.

Ничто из этого не скрывается внутри Directory, Binding, Authorize, Owner или
Flow.

## 29. Implementation boundary

Implementation Status — Implemented in isolation. Package
`internal/runtimemanagement` предоставляет exact Target, Binding, Directory,
authorization seam, routing и local proof tests этого design. Binding Control
Service, endpoint, concrete policy, persistence adapter, recovery path и
activation path отсутствуют.

Approved DP-014–DP-018 закрывают focused design gates ARCH-004 sections
19(2)–(6). Packages `internal/runtimeidentity` и
`internal/runtimecommandidempotency` реализуют DP-014 и DP-015 изолированно на
process-local in-memory storage. External durability, activation, recovery,
reporting, integration Control Service и Production Activation отсутствуют.
Любой integration slice сначала должен предоставить exact required
dependencies, включая composition существующих isolated private Start-claim
continuation и execution-generation binding/load gate, определённых Approved
DP-019. Concrete composition-private invoker точно определён Draft DP-021, а
его Implementation Status — Partial, реализован изолированно TASK-043. TASK-026
теперь реализует callback/terminal/orchestrator integration изолированно;
production composition остаётся отсутствующей.

## 30. Решение

Runtime Management Directory является одной immutable
process-local command boundary. Он проверяет и маршрутизирует exact identity
Runtime Instance, авторизует Start, Stop и Observe до lifecycle delegation,
конструирует один Flow из того же Owner и Loader для каждого принятого scope и
сохраняет все downstream lifecycle outcomes.

Он не добавляет второй lifecycle owner, dynamic registry, service locator,
симуляцию persistence, retry, transport API или Production Activation.

Draft design реализован изолированно. Full integration и Production Activation
остаются blocked отсутствующими dependencies и wiring, перечисленными выше.

## 31. Boundary consume admission для C2 diagnostics (TASK-121)

Этот раздел сохраняет TASK-121 boundary C2-Binding. Раздел 32 задаёт selected
exposure design TASK-123; ни один не добавляет current Directory method, endpoint
или production capability.

`ActionObserve` разрешает только detached lifecycle observation и никогда не даёт
право destructive diagnostics Poll. Authority command/replay DP-018 также неприменима.
C2-Exposure design раздела 32 выбирает отдельный diagnostics-consume authorization
contract с exact `Target` и expected non-empty Launch Attempt identity во входе.
Перегружать version parameter текущего `Authorize` запрещено.

Mandatory order каждого future request:

1. validate context, syntactic Target, expected attempt identity и Poll budget;
2. вызвать dedicated consume authorizer до target lookup, Owner access или Poll;
3. при denial, cancellation или invalid input вернуть результат без Directory lookup,
   раскрывающего existence, и с zero mailbox read;
4. после authorization выбрать exact immutable Directory scope; absent/mismatched
   targets используют тот же externally safe unavailable class;
5. вызвать expected-attempt diagnostics operation Owner, которая atomically match-ит
   active attempt и non-retired private lifetime до копирования View;
6. выполнить не более одного bounded `View.Poll` вне Directory/Owner locks.

Победившая current security authorization разрешает одну admitted invocation; grant
revocation после cut влияет на subsequent requests. Независимо от этого Owner lifetime
retirement до locked View claim предотвращает Poll, а после claim не retract-ит Poll.
Раздел 32.2 задаёт эти независимые cuts. Successor attempt никогда не
удовлетворяет старый expected-attempt request. Unauthorized, nonexistent, retired и
mismatched requests не consume-ят event и не раскрывают конкретное condition.

Раздел 32 и DP-027 раздел 37 задают concrete policy, bounds rate/concurrency/pending
requests, transport, output rendering, retention и redaction как Draft design. Он не добавляет unbounded
target/handle registry, background worker или lifecycle wait. Directory хранит только
existing Owner scope; raw Reader/View handle остаётся private у Owner.

## 32. Consume contract C2-Exposure (TASK-123)

Это выбранный **Draft / Planned** exposure design, не реализованный route или
production authorization. Раздел 31 сохраняет prerequisite TASK-121; этот раздел
задаёт его concrete consume policy и admission contract. Existing Start/Stop/Observe
не меняются. Resource/result/retention/proof requirements канонически заданы в
[DP-027, раздел 37](DP-027-runtime-diagnostics-v1.md#37-resources-results-и-proofs-c2-exposure-task-123).

### 32.1 Topology и trust

Единственный выбранный deployment — будущий single Control Service process с одним
immutable Directory, одним Lifecycle Owner на Runtime Instance и его in-process Host.
Current cmd/control-service их не создаёт; его HTTP server и stdout logger не получают
diagnostics authority из этого design. Host остаётся sole operational graph/lifecycle
owner. Reader/View остаются private у Owner; Producer проходит только через existing
launch-private binding, никогда через external API. Нет пятого Bootstrap binding,
mutable dependency rebind, worker, logger fallback, dynamic target registry или
cross-process capability transfer.

Planned wire adapter использует **TLS 1.3 mutually authenticated HTTP/1.1**, HTTP/2
отключён в initial contract. Trusted application authentication сопоставляет verified
configured client identity с opaque construction-issued principal handle. Body,
headers, query, arbitrary context values, unverified forwarded identities и
client-asserted grants не могут создать handle. Cookie/bearer/ambient-browser,
redirect или CORS authority отсутствуют.

Composition заранее задаёт максимум **32 principal slots** и **64 immutable Target
slots**. Unknown/zero/unconfigured handles fail closed. Requests не создают principal,
Target или attempt slots. Future adapter обязан обеспечить bounded ingress DP-027
раздела 37 до exposure route; current server этим не меняется.

### 32.2 Exact current consume authority

Dedicated action — `runtime.diagnostics.consume`, отдельно от ActionObserve,
Start/Stop, current Authorize version parameter и reporting/replay DP-018 либо durable
command authority. Одна evaluation получает verified principal handle, этот action,
exact Workspace/Configuration/Runtime Instance Target, nonempty expected Launch Attempt
ID, requested Poll budget и current caller context.

Concrete policy — **explicit current exact-grant membership**. Trusted grant snapshot
должен разрешать именно principal/action/full Target/expected-attempt tuple. Wildcard,
admin/role inheritance, cross-Workspace fallback, latest-attempt selection или
client-asserted grant consumption не разрешают. Snapshot содержит максимум **2,048
grants** (32 principals x 64 Targets), не более одного current expected-attempt grant
на principal/Target; grant history и diagnostic batch не хранит.

Trusted policy provider атомарно предоставляет current snapshot. Missing provider,
unavailable evaluation, malformed/unknown snapshot, absent/disabled/stale exact grant
или canceled evaluation означают deny. Missing production policy нельзя заменить
allow-all. Endpoint не provision/update grants. Authenticated trusted application grant
provisioning и revocation остаются отдельным **production activation prerequisite**;
существование backend этим не заявляется. Provider обязан соблюдать trusted prompt,
concurrent-safe и nonpanicking callback contract; remote policy I/O, detached evaluation
или timeout goroutine не прячутся в этой boundary.

Successful current evaluation — **security admission linearization point** ровно одной
invocation. Result не сериализуется, не replay/cache и не передаётся другому request.
Security-grant revocation до cut даёт deny; после cut влияет на последующие requests,
не отзывая уже admitted invocation. Независимо от этого Owner lifetime retirement до
locked View capture запрещает Poll, а после capture уже admitted Poll не отзывает.
External-policy check под Owner mutex и atomic multi-lock security/lifetime cut не
обещаются. Existing DP-010 seam сохраняется.

### 32.3 Admission и exact routing order

1. Попытаться взять process request slot без ожидания; closed admission отклонить до
   lookup. Ingress bounds действуют ещё до parsing или principal resolution.
2. Validate method/content type/body bound и один fixed DTO. Unknown/duplicate fields,
   compression и noncanonical identities отклонить без target lookup.
3. Resolve already-authenticated trusted principal handle, sample cancellation,
   charge process/principal rate buckets и acquire principal slot независимо от
   requested Target и его existence. Unknown handles fail closed.
4. Ровно один раз evaluate dedicated current consume authorizer на full request.
5. Recheck cancellation. Denial/cancellation до routing дают **zero Directory lookup,
   Owner call и mailbox read**.
6. Только после authorization выбрать existing immutable Directory scope и сравнить
   full Target. Не маршрутизировать по ConfigurationVersion/PID/Sequence/Host pointer,
   не fallback в другой Workspace или successor/latest attempt.
7. Попытаться взять configured Target slot и charge его rate bucket без ожидания.
   Absent/mismatched binding и Target quota refusal дают одинаковый unavailable result.
8. Один раз вызвать Owner.PollDiagnosticsExpectedAttempt с original expected attempt,
   context и budget. Owner атомарно matches active attempt/lifetime под mutex, captures
   View, unlocks и выполняет максимум один bounded Poll/Drain.
9. Validate detached result и построить только bounded allowlisted response DP-027
   раздела 37. Очистить temporary copies на ordinary return/unwind.
10. Освободить Target/principal/process in-flight slots после response handling на
    каждом returning/unwinding path. Никаких queue/retry/reinsert/detached work.

Admission/counter updates используют короткие nonblocking critical sections; contention
отказывает admission вместо скрытой waiting queue. Adapter/Directory/quota lock не
охватывает authorizer callback, Owner/Poll, serialization или network I/O. Contexts
сохраняют existing prompt/nonpanicking precondition. Deadlines не гарантируют завершение
hostile/nonreturning trusted callbacks. Existing Directory.admit, делающий lookup до
management authorization, не переиспользуется как consume path. Ни d.admit, ни
per-request d.valid/Observe/scope iteration не выполняются до consume authorization;
construction-time immutable binding validation допустима. После authorization новый
path может прямо выбрать existing package-private immutable map.

### 32.4 Planned request и failure API

Planned resource — `POST /runtime-diagnostics/v1/poll`, Content-Type application/json,
один request/response. Нет GET consumption, streaming, WebSocket, pagination, cursor,
compression, automatic retry или idempotency/replay key. Design не добавляет route/handler.

```json
{"workspaceId":"1","configurationId":"2","runtimeInstanceId":"runtime-opaque-id","expectedLaunchAttemptId":"attempt-opaque-id","budget":16}
```

Workspace/Configuration IDs — canonical nonzero uint64 decimal strings: 1..20 digits
в пределах uint64, без leading zero/sign/whitespace. Runtime Instance и expected-attempt
IDs — exact opaque 1..128 ASCII characters `[A-Za-z0-9_-]`. Trim/case-fold/normalization
отсутствуют. Wire grammar не меняет internal domain types; другие internal identities
недоступны через эту API version. Budget — required JSON integer 1..64. Expected
attempt берётся из отдельно authorized management knowledge; endpoint его не discovers.

| Condition | Exact external behavior |
| --- | --- |
| TLS authentication failure | TLS handshake failure; target-aware HTTP response отсутствует |
| Wrong method | 405, Allow: POST, fixed unavailable body; без lookup |
| Wrong content type, malformed/oversized/noncanonical DTO, duplicate/unknown field или compression | 400, fixed body `{"error":"invalid_diagnostics_request"}`; без lookup/read |
| Process/principal prelookup capacity/rate refusal | 429, fixed unavailable body, независимо от requested Target |
| Process shutting down | 503, fixed unavailable body, независимо от Target |
| Unknown trusted principal; consume deny/failure/unavailable; absent/mismatched Target; successor/retired lifetime; Target quota refusal; Poll Busy или invalid result | 404, identical bytes `{"error":"diagnostics_unavailable"}` |
| Valid detached result | 200, только response DP-027 раздела 37 |
| Caller disconnect/cancellation | Остановить invocation; без usable connection response нет; без retry/replay/reinsert |

Fixed unavailable body для 405/429/503 также равен
`{"error":"diagnostics_unavailable"}`. Все responses используют application/json и
Cache-Control: no-store; ETag/Last-Modified/shared cache/Retry-After отсутствуют.
Error body не echo Target/attempt/principal, grant epoch/reason, quotas, existence,
raw causes или internal status. Target-specific Busy/rate refusal — **404**, никогда
не target-aware 429. Anti-enumeration гарантирует same external class/body и zero
denied lookup/read, не exact microtiming equality. После Poll admission failure уже
может означать consumption; его body не обещает zero consumption или retraction.

### 32.5 Implementation split и activation gates

Первый candidate — isolated diagnostics admission/resource adapter только в
internal/runtimemanagement/diagnostics.go и diagnostics_test.go с existing Directory
package-private immutable scopes. Existing NewDirectory/Start/Stop/Observe/Authorize
APIs менять не требуется. Новый dedicated adapter API может выдавать opaque principal
handles при trusted construction; raw Owner/Reader/View он не exposes. Construction
rejects missing evaluator, invalid clock/configuration, более 32 principals/64 Targets
и duplicate/reused bindings или principal slots внутри supplied configuration. Copies одного adapter разделяют quota state. Constructor не обнаруживает отдельно
созданные adapters/Directories без нового global registry. Один sanctioned adapter на
process и отсутствие bypass — обязательные Composition Audit/activation proofs, не
вымышленный global enforcement конструктора. Production authentication backend/route/handler не создаёт.
Test grants не являются production policy.

Core — candidate **Ready только для отдельного implementation intake после final design
Acceptance**, exact scope/coverage confirmation и отдельного Owner permission. Следующий
отдельный candidate — bounded mTLS transport/serialization с fake core. Ни один здесь
не активируется. Production integration **NOT READY** до trusted mTLS configuration,
identity-to-handle mapping, real current exact-grant provisioning/revocation, bounded
ingress, единственной sanctioned adapter/Directory/Owner composition, применимых
ARCH-004 section 19 dependencies и integrated P01-P08 proofs. Unavailable fail-closed
backend не является useful production consume authority. Isolated allow-all fixture
через этот gate не проходит. C2/C3 Not Activated; Gate 11 OPEN. Frozen/Approved
contract не изменяется.

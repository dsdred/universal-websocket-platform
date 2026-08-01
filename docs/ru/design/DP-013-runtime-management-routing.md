# DP-013: Маршрутизация управления Runtime

[English version](../../en/design/DP-013-runtime-management-routing.md)

## 1. Статус

- **Design Status:** Draft
- **Implementation Status:** Planned
- **Implementation Readiness:** Blocked оставшимися обязательными focused
  designs ARCH-004 section 19

До утверждения proposal не является нормативным. Он определяет планируемую
изолированную in-process management command boundary. Package, production
routing Control Service, authorization policy, persistence, recovery и
Production Activation не появляются в результате этого документа.

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

Accepted ADR, Frozen foundation и Active architecture сохраняют приоритет.

## 4. Область

Design охватывает один process-local command directory, immutable bindings
Runtime Instance, exact identity routing, policy-neutral authorization seam,
порядок authorization-before-mutation, concurrency, cancellation, сохранение
failures, composition audit и требования к будущим изолированным proofs.

Он не определяет transport resources или data transfer objects.

## 5. Package и ответственность

Планируемый package — `internal/runtimemanagement`.

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

## 7. Точная планируемая public surface

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

Future implementation DP-016/DP-017 сохраняет exported surface
`Directory.Start` без изменения, но требует один private management-only seam
Start-claim continuation в stored Flow. Exact management composition снабжает
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
`LaunchPreparation` и является Planned, а не implemented.

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

Design contract валиден, но Implementation Readiness заблокирована.
ARCH-004 section 19 имеет Active status и приоритет над этим Draft. Package
`internal/runtimemanagement` и local proof implementation не начинаются, пока
focused designs не разрешат все оставшиеся обязательные prerequisites:

1. approval/status decision для candidate persistence contract section 19(2);
2. durable idempotency management command;
3. ordering activation, replacement и rollback;
4. recovery и reconciliation после termination Control Service;
5. operational error reporting и redaction.

Loader, provenance Snapshot и schema compatibility уже разрешены ARCH-005 и
DP-007 — DP-012. Прецедент isolated implementation DP-010 — DP-012 не создаёт
исключение из gate ARCH-004 с более высоким status.

Draft [DP-014](DP-014-runtime-operational-identity-persistence.md) предлагает
candidate contract persistence Runtime Instance и Launch Attempt, требуемый
ARCH-004 section 19(2). Он определяет durable aggregate, identity, history,
conditional revision и indeterminate-outcome semantics без создания
implementation persistence. Поскольку он остаётся non-normative Draft,
section 19(2) остаётся formal implementation blocker до отдельного
approval/status decision.

Draft [DP-015](DP-015-runtime-management-command-idempotency.md) теперь
предлагает candidate contract durable idempotency management commands для
section 19(3). Он связывает одну authorized command identity с immutable intent
до lifecycle delegation и определяет non-mutating replay. Как non-normative
Draft он не снимает gates section 19(2) или 19(3) и не активирует
implementation.

Draft [DP-016](DP-016-runtime-activation-replacement-rollback.md) теперь
предлагает candidate contract ordering activation, replacement и explicit
rollback для section 19(4). Он сохраняет exact-version attempts и
Stop-to-proven-release перед любым replacement или rollback Start. Как
non-normative Draft он не снимает gates section 19(2), 19(3) или 19(4) и не
активирует implementation.

## 27. Будущие implementation proofs

После разрешения каждого prerequisite section 26 будущая implementation task
должна доказать:

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

Оставшиеся блокирующие architecture prerequisites до любой implementation:

- approval/status decision для candidate persistence contract section 19(2);
- durable idempotency management command;
- ordering activation, replacement и rollback;
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

Implementation Status остаётся Planned. Package
`internal/runtimemanagement`, test, binding Control Service, endpoint, policy,
persistence adapter или activation path не появляются в результате этого
документа.

Implementation Readiness — Blocked. Ни isolated package, ни local proof code
не разрешены, пока не существуют все обязательные focused designs section 26.
DP-015, DP-016 и [DP-017](DP-017-runtime-recovery-reconciliation.md) являются
candidate designs sections 19(3), 19(4) и 19(5), а не implementation tasks;
approval gates sections 19(2)–(5) сохраняются. Draft
[DP-018](DP-018-runtime-operational-error-reporting-redaction.md) теперь
предлагает candidate contract reporting/redaction section 19(6). Он не снимает
ни один gate section 19 и не разрешает implementation до отдельных status
decisions.

## 30. Решение

Планируемый Runtime Management Directory является одной immutable
process-local command boundary. Он проверяет и маршрутизирует exact identity
Runtime Instance, авторизует Start, Stop и Observe до lifecycle delegation,
конструирует один Flow из того же Owner и Loader для каждого принятого scope и
сохраняет все downstream lifecycle outcomes.

Он не добавляет второй lifecycle owner, dynamic registry, service locator,
симуляцию persistence, retry, transport API или Production Activation.

Design готов к review и возможному принятию как Draft/Planned. Implementation
остаётся заблокированной ARCH-004 section 19(2)-(6).

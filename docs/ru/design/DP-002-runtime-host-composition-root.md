# DP-002: Runtime Host Composition Root

[English version](../../en/design/DP-002-runtime-host-composition-root.md)

## 1. Статус

**Статус:** Draft

Этот документ предлагает production-границу composition для одного экземпляра Runtime. Он определяет сборку компонентов, readiness, координацию lifecycle, startup rollback и порядок shutdown. Документ не определяет Go API или business behavior подсистем.

## 2. Контекст

Alpha Runtime содержит компоненты рабочей вертикали, но ни один production path не собирает их в единую систему и не владеет ею.

Сегодня:

- `runtime.DefaultHost` хранит копии Snapshot и Container и выполняет только переходы `Created -> Running -> Stopped`;
- `runtime.Container` предоставляет только копию Snapshot;
- Authentication Bootstrap строит Service из Authentication Snapshot, Registry, Factories и Secret Resolver;
- Listener Bootstrap строит Listener из Listener Snapshot и внедрённого Dispatcher;
- Session Dispatcher, Authentication Dispatcher, Message Handler, Registry, Resolver и Listener вручную связываются callers и integration tests;
- Host не запускает Listener, не выполняет rollback частичной сборки, не координирует shutdown Session, не владеет Runtime context и не предоставляет содержательную readiness.

Эти границы полезны и остаются узкими, но subsystem Bootstrap больше недостаточно для Beta. Каждый Bootstrap умеет создать одну подсистему; ни один не владеет полным dependency graph, startup transaction, cleanup в обратном порядке или общим failure state Runtime. Расширение Container до набора services скрыло бы проблему, а не решило её.

[Runtime Alpha Architecture Review](../reviews/runtime-alpha-review.md) фиксирует отсутствие production composition root как High finding F-02. [DP-001: Runtime Handshake Pipeline](DP-001-runtime-handshake-pipeline.md) также требует startup readiness, Runtime-owned context, admission gate и участия Session в Runtime shutdown wait set.

## 3. Цели

- Сделать Runtime Host единственной production composition root одного экземпляра Runtime.
- Собирать поддерживаемый Runtime graph из одного immutable Published Snapshot и явных operational dependencies.
- Проверять всё поддерживаемое effective behavior до начала обслуживания traffic.
- Определить детерминированные construction, startup, rollback, shutdown и failure ordering.
- Предоставить Runtime один owned root context и одну readiness boundary.
- Сохранить узкие subsystem contracts и dependency direction.
- Сделать создание компонентов явным, обычным и удобным для review.
- Позволить тестам заменять узкие dependencies без production service locator.

## 4. Не входит в цели

Runtime Host не:

- маршрутизирует Runtime Messages;
- аутентифицирует clients и не реализует Provider logic;
- владеет WebSocket transport после handoff в Session;
- хранит, ищет, группирует или адресует Sessions;
- реализует Delivery или Persistence;
- записывает application data в database;
- самостоятельно разрешает credentials request;
- определяет Plugin ABI или динамически обнаруживает Plugins;
- интерпретирует business rules и не принимает Decisions подсистем;
- выполняет reload или restart экземпляра Runtime;
- проектирует загрузку Control Plane или supervision процесса.

Host принимает только решения о composition и lifecycle. Всё поведение подсистем он делегирует компоненту, владеющему соответствующей ответственностью.

## 5. Ответственности

Runtime Host отвечает за:

- приём immutable Runtime Snapshot и явных operational dependencies;
- проверку совместимости Snapshot и поддержки effective Runtime;
- создание concrete infrastructure dependencies через явные constructors;
- регистрацию явно поддерживаемых Authentication Factories;
- вызов subsystem Bootstrap в порядке dependencies;
- wiring Message Handler, Session handoff, Handshake, Authentication и Listener boundaries;
- создание и cancellation root Runtime context;
- запуск externally visible components только после успешной внутренней сборки;
- rollback частично initialized или started components;
- перевод Runtime в Ready только после работоспособности всех обязательных компонентов;
- закрытие admission до shutdown cancellation;
- остановку и cleanup owned components в обратном порядке dependencies;
- сохранение primary и cleanup failures для diagnostics.

Host не является general-purpose dependency container. Его fields и construction steps соответствуют только поддерживаемым production components. Добавление компонента требует явного изменения composition, обоснованного design соответствующей подсистемы.

## 6. Dependency Graph

### Construction Graph

```text
Published ConfigurationVersion
    -> runtimeconfig.Builder
    -> immutable Snapshot
    -> Runtime Host
         |
         +-> Runtime root context и lifecycle state
         |
         +-> concrete Secret Resolver
         |
         +-> Authentication Registry
         |      +-> explicit API Key Factory
         |      +-> explicit JWT Factory
         |
         +-> Authentication Bootstrap
         |      +-> Authentication Service
         |             +-> ordered Providers
         |                    +-> Secret Resolver
         |
         +-> Message Handler
         +-> Session handoff/Dispatcher
         +-> Handshake executor и admission gate
         |      +-> Authentication Service
         |      +-> Session handoff
         |
         +-> Listener Bootstrap
                +-> Listener
                       +-> Handshake executor
```

Graph описывает explicit production composition, а не обязательную package structure или generic construction API. Существующие contracts Authentication Factory и Registry остаются subsystem-specific; Host не вводит universal Factory или Registry.

### Runtime Call Direction

```text
Listener
    -> Handshake executor
         -> Authentication Service
              -> Provider
                   -> Secret Resolver
         -> Session handoff
              -> Session
                   -> Message Handler
```

Обратные dependencies запрещены:

- Listener не знает Host, concrete Providers, внутреннее устройство Session, Router или Persistence.
- Authentication не знает Host, Listener, HTTP, WebSocket, Session или repositories.
- Session не знает Host, Listener, Snapshot, Control Plane или repositories.
- Secret Resolver не знает Host, поведение Authentication или ConfigurationVersion.
- Message Handler не знает Host или WebSocket transport.
- ни один Runtime component не ищет другой компонент через Container, globals, reflection или `context.Value`.

`runtime.Container` остаётся holder Snapshot, пока отдельное сфокусированное решение не изменит это. Proposal не превращает его в service locator.

## 7. Startup Pipeline

Startup является одной упорядоченной transaction:

```text
Immutable Snapshot
    -> validation совместимости и effective settings
    -> создание root Runtime context
    -> создание и readiness Secret Resolver
    -> resolution обязательных startup Secrets
    -> Authentication Registry и регистрация Factory
    -> создание Authentication Service
    -> создание Message Handler и Session handoff
    -> создание Handshake executor и admission gate
    -> создание Listener без открытия socket
    -> Listener Start
    -> открытие admission gate
    -> Runtime Ready
```

Подробный порядок:

1. Host копирует и проверяет identity Snapshot, Listener metadata, timeouts, поддержку TLS, Authentication metadata и каждый active setting, который Runtime заявляет как исполняемый.
2. Host создаёт Runtime-owned root context. Startup context caller ограничивает startup, но не становится lifetime успешно запущенного Runtime.
3. Host явно создаёт выбранный Secret Resolver и проверяет его startup readiness.
4. Secret material, необходимый для запуска компонента, например будущий TLS key material, разрешается до Listener Start и передаётся только его владельцу. Authentication Providers, намеренно разрешающие credentials для каждого request, сохраняют этот контракт; Host проверяет их references и construction, но не хранит их secret values заранее.
5. Host создаёт Authentication Registry и регистрирует только production Factories, явно поддерживаемые этой сборкой.
6. Authentication Bootstrap создаёт упорядоченные Service и Providers из Authentication Snapshot и Resolver.
7. Host создаёт выбранные Message Handler, Session handoff и Handshake executor из DP-001 вместе с admission gate.
8. Listener Bootstrap проверяет Listener Snapshot и создаёт Listener без network side effects.
9. Listener запускается последним. Network traffic не принимается до существования всех обязательных downstream components.
10. Host открывает admission gate и переходит в `Running`. Успешный Start означает, что Runtime Ready.

Construction детерминирован. Map iteration, reflection, package discovery, dynamic registration order и generic component factories не определяют behavior.

## 8. Shutdown Pipeline

Shutdown обращает порядок externally visible ownership и dependencies:

```text
Runtime Ready = false
    -> закрытие admission gate
    -> остановка Listener acceptance и HTTP admission
    -> cancellation root Runtime context
    -> завершение или abort активных Handshakes и commits
    -> ожидание handed-off Sessions из Runtime shutdown set
    -> остановка владельцев Session handoff и message processing
    -> освобождение Authentication components
    -> освобождение startup-resolved Secrets и ресурсов Resolver
    -> Runtime Stopped
```

Гарантии:

- readiness становится false до начала shutdown work;
- после закрытия gate не начинается новый admission commit;
- уже выполняющийся commit следует DP-001: входит в Runtime shutdown tracking через успешный handoff либо закрывается до завершения shutdown;
- каждый owned component останавливается или освобождается не более одного раза;
- cleanup выполняется в порядке, обратном успешному acquisition/start;
- Stop идемпотентен, concurrent callers наблюдают одинаковые completion и terminal result;
- shutdown errors накапливаются без потери первой causal failure;
- deadline caller ограничивает ожидание этого caller, а не ownership obligation Host;
- если cleanup продолжается после deadline caller, он остаётся Host-owned в `Stopping` и имеет явный completion path;
- dependency, игнорирующую cancellation, нельзя принудительно остановить; Runtime сообщает timeout и сохраняет ответственность за eventual cleanup, а не объявляет ложный state `Stopped`.

Ни один component не может отделить cleanup work без владельца Runtime.

## 9. Ownership

| Resource или component | Creator | Lifecycle owner | Ответственность за cleanup |
|---|---|---|---|
| Published ConfigurationVersion | Граница Control Plane/Loader | Вне Runtime | Не сохраняется Runtime services |
| Копия immutable Snapshot | Builder, затем копируется Host | Host | Lifetime значения; никогда не мутируется |
| State Host и root Runtime context | Host | Host | Cancel при shutdown, rollback или terminal failure |
| Secret Resolver | Host через explicit concrete constructor или explicit owned dependency | Host | Close/release, если его contract содержит lifecycle |
| Startup-resolved secret material | Resolver с передачей consuming component | Consuming component | Clear/release согласно contracts компонента и Resolver |
| Authentication Registry | Host | Host | Cleanup значения; без dynamic production mutation после initialization |
| Authentication Providers и Service | Authentication Bootstrap в composition Host | Host для lifecycle компонентов; Provider владеет per-call secret copies | Освобождение component resources; очистка per-call Secrets рядом с использованием |
| Message Handler | Host через explicit subsystem constructor | Host или будущий subsystem owner | Stop, только если focused contract определяет lifecycle |
| Handshake executor и admission gate | Host | Host | Закрытие gate, cancellation context, tracking terminal Handshake |
| Listener | Listener Bootstrap в composition Host | Host координирует; Listener владеет socket/server | Host вызывает Listener Stop; Listener закрывает и ожидает свои resources |
| Upgraded WebSocket до handoff | Upgrade boundary из DP-001 | Upgrade boundary | Close при failure до handoff |
| Session после ownership acceptance | Session | Session; Runtime отслеживает completion shutdown | Session закрывает WebSocket; Host никогда не выполняет double close |

Borrowed operational dependencies должны явно обозначаться при composition. Production defaults предпочитают Host-owned lifecycle dependencies. Host никогда не хранит raw credential values ради их глобальной доступности.

## 10. Lifecycle

Концептуальный lifecycle Host:

```text
Created
    -> Initialized
    -> Starting
    -> Running
    -> Stopping
    -> Stopped

Created / Initialized / Starting / Running / Stopping
    -> Failed
```

- **Created:** Host владеет копией input, но ещё не имеет полного component graph.
- **Initialized:** Snapshot и readiness checks успешны; все обязательные components созданы без externally visible traffic.
- **Starting:** lifecycle-bearing components запускаются по порядку; admission остаётся закрытым.
- **Running:** все обязательные components работоспособны, admission открыт, Runtime Ready.
- **Stopping:** readiness и admission закрыты; выполняются cancellation, waits и cleanup в обратном порядке.
- **Stopped:** все owned resources завершили cleanup. State является terminal.
- **Failed:** initialization, startup, runtime supervision или cleanup завершились ошибкой. Causal и cleanup errors остаются наблюдаемыми. Cleanup всё равно выполняется до освобождения ownership resources.

Restart и in-place reload не поддерживаются. Новый Published Snapshot создаёт новый экземпляр Runtime Host.

Stop до Running выполняет reverse cleanup всего уже acquired. Normal explicit Stop достигает `Stopped`; startup failure, startup cancellation или unexpected component termination достигают `Failed` после rollback. Cleanup failure также оставляет terminal state `Failed`.

## 11. Failure Model

### Invalid Snapshot

Initialization завершается ошибкой до запуска network resource. Host фиксирует invalid Runtime configuration, освобождает ранее owned values и переходит в `Failed`.

### Secret Resolution Failure

Failure Secret, обязательного при startup, предотвращает Listener Start. Частично resolved material освобождается, initialized components откатываются в обратном порядке. Per-request Provider resolution failure остаётся Authentication operational error и не изменяет Runtime readiness задним числом, пока отдельная focused health policy не определит иное.

### Authentication Construction Failure

Missing Factory, unsupported enabled Provider, invalid Provider runtime configuration или unavailable mandatory Authentication dependency предотвращают Listener Start. Registry, Resolver и другие более ранние components очищаются.

### Listener Start Failure

Runtime не становится Ready. Host останавливает частично started state Listener, затем освобождает Handshake, Session handoff, Authentication и Resolver resources в обратном порядке.

### Startup Cancellation

Startup учитывает cancellation caller и root context Host. Cancellation предотвращает admission, выполняет rollback acquired resources и завершается в `Failed` с категорией cancellation/deadline.

### Unexpected Failure While Running

Неожиданное завершение обязательного lifecycle component немедленно сбрасывает readiness, закрывает admission, запускает coordinated shutdown и приводит к `Failed`. Runtime не может оставаться Ready без обязательного component.

### Rollback Failure

Rollback продолжается после отдельной cleanup error. Primary startup или runtime failure и все cleanup failures остаются различимыми. Host никогда не сообщает successful Start или clean Stop при незавершённом ownership resources.

## 12. Runtime Readiness

Runtime Ready, только если одновременно выполнены условия:

- state Host равен `Running`;
- Snapshot и все active settings прошли support validation;
- обязательные startup Secrets разрешены и переданы владельцам;
- Authentication Service и все enabled Providers созданы;
- Handshake executor и admission gate работоспособны;
- Listener успешно запущен и принимает traffic;
- root Runtime context активен;
- каждый обязательный lifecycle component находится под supervision Host.

До Ready:

- admission gate остаётся закрытым;
- Listener не должен принимать externally usable WebSocket Sessions;
- успешные construction или `Initialized` не считаются успешным startup Runtime;
- Runtime не должен объявляться доступным.

Readiness становится false атомарно в начале shutdown или terminal failure. Readiness является lifecycle fact Host; документ не проектирует HTTP health endpoint, service-discovery protocol или monitoring API.

## 13. Runtime Context

Host создаёт и владеет одним root Runtime context во время initialization/startup. Он отделён от context, переданного caller в Start:

- startup context управляет длительностью ожидания startup caller;
- root Runtime context управляет lifetime работающего экземпляра;
- cancellation Host завершает Runtime-owned work при rollback, Stop или terminal failure.

Components получают root context или узко производные child contexts через explicit construction или lifecycle calls. Context values не переносят services или configuration.

Handshake context из DP-001 учитывает request cancellation, настроенный Handshake timeout и Runtime shutdown. После successful admission Session получает отдельный Runtime-owned connection context, производный от Runtime lifecycle, а не от `http.Request.Context()`. Host shutdown cancellation достигает этого connection context, при этом после handoff владельцем WebSocket остаётся Session.

Закрытие admission gate предшествует root cancellation, поэтому Evaluation, завершившаяся во время shutdown, не может начать новый Upgrade commit.

## 14. Extension Points

Позднее Host может компоновать дополнительные focused components, но не определяет их API:

- **Metrics:** diagnostics/metrics component может наблюдать bounded lifecycle и subsystem events, не становясь dependency lookup mechanism.
- **Persistence:** focused Storage component может быть создан до Handler или Delivery component, который от него зависит, и остановлен после своих dependents.
- **Router:** спроектированный Router может заменить текущий explicit выбор Message Handler без изменения Listener или Authentication.
- **Plugins:** поддерживаемые Plugin implementations могут предоставляться через subsystem-specific contracts после design compatibility и isolation; Host не выполняет scanning и не предоставляет universal Plugin registry.
- **Delivery:** focused Delivery component может быть связан между Router decisions и Session capabilities после design его semantics.

Каждое extension должно определить dependencies, ownership, readiness, startup failure и shutdown order в собственном DP. Host изменяется только для добавления explicit production wiring. Ни одно extension не получает произвольный доступ к Host или Container.

## 15. Миграция

Миграция выполняется небольшими проверяемыми шагами:

1. Сохранить characterization tests текущих Host, Bootstrap, Listener, Authentication, Session и Echo vertical.
2. Ввести расширенный lifecycle Host и readiness semantics без запуска network components.
3. Перенести support validation Snapshot и rejection effective settings в pre-start composition phase.
4. Предоставить Host owned Runtime context и deterministic rollback ledger для явно acquired components.
5. Создать Resolver, Registry и production Authentication Factories через explicit wiring; построить Authentication Service через существующий Bootstrap.
6. Скомпоновать Message Handler и Session handoff через существующие узкие contracts.
7. Скомпоновать Handshake executor и admission gate из DP-001 после реализации этой boundary.
8. Создать Listener последним и сделать его единственным externally visible component, запускаемым Host.
9. Добавить failure-injection tests для каждой initialization/start boundary и каждого reverse cleanup step.
10. Заменить manual production assembly на Host, сохранив direct component construction в unit и focused integration tests.
11. Добавить supervision tests для unexpected Listener termination, concurrent Stop, deadlines и partial cleanup failure.

Ни один migration step не превращает Container в service locator, не вводит reflection и не требует general component framework.

## 16. Связь с другими документами

### ARCH-001

Proposal применяет [ARCH-001](../architecture/ARCH-001-runtime-architectural-pattern.md): dependencies явны, у каждого resource один owner, lifecycle и cancellation видимы, а Host делегирует Evaluation и Execution сфокусированным components. Composition root намеренно boring и не является god object.

### MASTER_PLAN

[Master Engineering Plan](../roadmap/MASTER_PLAN.md) определяет production Runtime Host, startup rollback, shutdown ordering, lifecycle hardening и Configuration validation как Beta foundation gates перед Router. Proposal определяет их composition boundary, не становясь task backlog.

### DP-001

[DP-001: Runtime Handshake Pipeline](DP-001-runtime-handshake-pipeline.md) определяет pre-Upgrade Authentication, admission gate, Runtime-owned Session context и требования shutdown handoff. Host создаёт и владеет dependencies, необходимыми для соблюдения этих инвариантов, но не реализует Handshake decisions.

### Принятые ADR

- [ADR-0002](../adr/0002-configuration-dsl.md) делает Published ConfigurationVersion источником истины и требует явно отклонять unsupported behavior.
- [ADR-0003](../adr/0003-runtime-architecture.md) требует независимых components, explicit dependency injection, immutable Snapshot и production composition root без выбора DI framework.

## 17. Открытые вопросы

- Какой concrete production Secret Resolver выбирается для первой Beta environment?
- Какая startup readiness operation, если она нужна, может проверять remote Secret backend без чтения request-scoped credentials?
- Как представляются и публикуются multiple cleanup errors без потери primary failure?
- Какой supervision signal представляет unexpected termination Listener или другого mandatory component?
- Какой component владеет Runtime shutdown wait set до появления focused design Session Manager?
- Какой diagnostics sink принимает lifecycle, startup, rollback и terminal component errors?
- Какой process-level owner создаёт и заменяет экземпляры Host при активации нового Published Snapshot?

Reload, restart, Router behavior, Delivery semantics, Plugin ABI и service-discovery APIs находятся вне scope, а не являются open questions этого proposal.

## 18. Решение

Runtime Host является единственной production composition root одного экземпляра Runtime.

Он получает один immutable Snapshot и явные operational inputs, создаёт поддерживаемый component graph через обычные constructors и существующие subsystem Bootstrap, проверяет readiness до открытия traffic, запускает Listener последним, владеет root Runtime context и выполняет rollback или shutdown в обратном порядке acquisition.

Выбранный design централизует composition, но не business behavior. Host не маршрутизирует, не выполняет Authentication, не хранит Sessions, не сохраняет данные и не предоставляет доступ к internal dependencies. Он не использует service locator, DI framework, reflection, generic component factories или Universal Component Registry. Explicit wiring выбрано потому, что сохраняет dependency direction, ownership, failure и cleanup видимыми в коде и review.

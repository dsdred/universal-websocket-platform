# DP-001: Runtime Handshake Pipeline

[English version](../../en/design/DP-001-runtime-handshake-pipeline.md)

## 1. Статус

**Статус:** Draft

Этот документ предлагает архитектуру Runtime Handshake. Он определяет концептуальные границы и инварианты, а не Go-интерфейсы, пакеты или переиспользуемый policy framework.

## 2. Контекст

Alpha Runtime выполняет Upgrade HTTP request до Authentication. Полученный WebSocket затем проходит через Connection Dispatcher, Authentication Dispatcher и Session Dispatcher. Поэтому клиент с отклонёнными credentials сначала получает `101 Switching Protocols`, после чего сервер закрывает соединение.

[Runtime Alpha Architecture Review](../reviews/runtime-alpha-review.md) фиксирует это как High finding F-01. [DP-001: Authentication](../proposals/DP-001-authentication.md) уже требует выполнить identity evaluation до допуска WebSocket Session.

Текущий путь:

```text
TCP accept
    -> HTTP request
    -> WebSocket Accept
    -> Authentication
    -> reject через WebSocket close
       или
       Session
```

Предложение меняет порядок admission, не перенося Provider logic в Listener и не связывая Authentication с HTTP или WebSocket.

## 3. Постановка проблемы

Authentication после Upgrade пересекает границу ownership транспорта до того, как известно admission decision. Это выделяет WebSocket-ресурсы отклонённым клиентам, исключает HTTP authentication rejection, усложняет shutdown и смешивает policy rejection с ошибками после Upgrade.

Design также должен устранить пробелы ownership и lifecycle из независимого Architecture Stress Review: request context не должен становиться Session context, shutdown должен атомарно закрывать admission gate, точка handoff WebSocket должна быть однозначной, а error model должен учитывать реальное поведение `websocket.Accept`.

## 4. Существующая архитектура

Необходимо сохранить следующие текущие ответственности:

- Listener владеет listening socket и lifecycle HTTP server.
- `net/http` владеет принятыми HTTP connections до успешного hijack или Upgrade.
- Authentication Service и Providers работают с transport-neutral input.
- Runtime Snapshot предоставляет effective Configuration без доступа к Repository.
- Session владеет допущенным WebSocket, его read loop и сериализованной записью.

Текущий порядок и передача context не сохраняются. Authentication переносится до Upgrade, а Session получает Runtime-owned lifecycle context вместо `http.Request.Context()`.

## 5. Цели

- Завершить настроенную Authentication до WebSocket Upgrade.
- Сохранить transport-neutral Authentication Service и Providers.
- Определить единственного владельца каждого transport resource на каждом этапе.
- Определить admission state machine и её переходы при shutdown.
- Разделить request-scoped Handshake context и connection-scoped Session context.
- Различать rejection, cancellation, dependency, configuration, protocol, internal и handoff failures.
- Исключить чтение Snapshot и Control Plane repositories при обработке request.
- Ограничить первую реализацию pre-Upgrade Authentication.

## 6. Не входит в цели

Документ не проектирует:

- Router или Delivery;
- API Session Manager;
- Middleware или Universal Policy Engine;
- Plugin ABI;
- Origin Policy за пределами текущего поведения библиотеки;
- rate limiting, maintenance, IP filtering или общие admission rules;
- другой транспорт, QUIC или HTTP/3;
- Provider-specific credential verification;
- детали композиции Runtime Host.

## 7. Ограничения

Design следует [ARCH-001: Runtime Architectural Pattern](../architecture/ARCH-001-runtime-architectural-pattern.md).

### Configuration First

Effective behavior строится из Published Configuration Snapshot до запуска Listener. Unsupported active behavior отклоняется, а не игнорируется.

### Ownership

У каждого mutable transport resource ровно один текущий владелец. Передача ownership явная и имеет одну концептуальную точку принятия.

### Lifecycle

Handshake и Session имеют отдельные contexts и lifetimes. Network I/O и вызовы dependencies не выполняются при удержании lifecycle lock.

### No Magic

Dependencies компонуются явно. Request path не обнаруживает Providers, не читает repositories или Snapshot, не использует global state или `context.Value` как service locator.

### Boring Core

Первая реализация выполняет только настроенную Authentication до Upgrade. Концептуальная последовательность не является generic evaluator chain и не обещает будущие policies.

## 8. Рассмотренные альтернативы

### Alternative A — Сохранить Authentication после Upgrade

Отклонённые клиенты продолжат получать `101`, занимать upgraded resources и входить в connection lifecycle до admission. Это сохраняет дефект, поэтому альтернатива отклонена.

### Alternative B — Перенести Provider Logic в Listener

Listener сможет выполнить Authentication до Upgrade, но получит знания о Providers, Secrets и проверке credentials. Это нарушает transport neutrality и single responsibility, поэтому альтернатива отклонена.

### Alternative C — Выполнять Authentication в Session

Provisional Session будет существовать без допустимого Principal и сможет отклонить клиента только через WebSocket closure. Это нарушает границу authenticated Session, поэтому альтернатива отклонена.

### Alternative D — Выделенная граница Handshake

Концептуальная граница Handshake нормализует request metadata, выполняет настроенную Authentication, создаёт одноразовый admission Decision и только затем разрешает transport executor начать commit Upgrade.

Эта альтернатива выбрана. «Pipeline» описывает упорядоченный lifecycle одного request. Он не определяет generic pipeline API, registry evaluators или universal policy model.

## 9. Предлагаемая архитектура

```text
Published Snapshot
    -> Runtime Host bootstrap и readiness validation
    -> Host-owned Admission Gate и Runtime context holder
    -> read-only Runtime capabilities
    -> composition bridge
    -> composed Handshake, Listener, Authentication Service, timeout и dependencies
    -> Listener Start

HTTP request
    -> начальная проверка admission
    -> Handshake Context
    -> pre-Upgrade Authentication
    -> one-use Decision
       -> Reject или operational failure: HTTP handling, без WebSocket
       -> Allow: финальная проверка admission непосредственно перед commit
    -> websocket.Accept
       -> Accept rejection: библиотека может выполнить commit HTTP response
       -> success: Upgrade boundary владеет WebSocket
    -> Runtime-owned connection context
    -> Session зарегистрирована в Runtime shutdown wait set
    -> explicit Session ownership acceptance
    -> Session lifecycle
```

Runtime composition связывает Host-owned lifecycle state с Handshake через живые read-only capabilities. Handshake не зависит от Runtime Host и не может изменять Admission Gate или отменять Runtime context. Listener остаётся ответственным за transport execution, но не содержит Provider-specific logic. Authentication остаётся transport-neutral. Session допускается только с допустимым Principal и никогда не получает исходный HTTP request или его context как lifecycle state.

## 10. Модель Context

### Handshake Context

Handshake executor владеет одним request-scoped context. Его lifetime завершается в `Rejected`, `Aborted` или после успешного `HandedOff`.

Effective deadline является самым ранним из:

- cancellation `http.Request.Context()`;
- cancellation Runtime или Listener shutdown;
- настроенного Handshake timeout из effective Runtime configuration.

В transport-neutral Authentication input копируются только необходимые request metadata. HTTP request, Body, Headers, URL, ResponseWriter и request context не передаются в Session.

### Runtime-Owned Connection Context

После успешного Accept и до handoff Upgrade boundary создаёт отдельный connection/session context от активного Runtime context, предоставленного read-only Runtime context capability. Он не является дочерним context, lifetime которого зависит от `http.Request.Context()`.

Путь cancellation принадлежит Runtime lifecycle. Context отменяется при завершении connection или Session, а также при Runtime shutdown. Будущий Session Manager может участвовать в lifecycle, но его API находится вне документа.

После успешного handoff HTTP handler может завершиться, не отменяя Session.

## 11. Инварианты Decision

Admission Decision следует этим инвариантам:

- Allow содержит допустимый Principal.
- При включённой Authentication Principal представляет успешную Authentication.
- При отключённой Authentication Principal является explicit anonymous Principal, определённым ниже.
- Reject является ожидаемым отрицательным результатом, а не Go error.
- Operational failure отличается от Reject.
- Decision теряет силу, если его context отменён или admission gate закрыт до начала commit.
- Decision может быть применён не более одного раза.
- Только Allow может начать admission commit.
- Execution error не переписывает и не переклассифицирует исходный Authentication result.

Точное представление в Go остаётся implementation detail.

## 12. Admission State Machine

Концептуальные состояния:

```text
Evaluating
    -> Allowed
    -> Rejected
    -> Aborted

Allowed
    -> Committing
    -> Aborted

Committing
    -> Upgraded
    -> Rejected     (Accept выполняет protocol или Origin rejection)
    -> Aborted      (cancellation или transport/internal failure)

Upgraded
    -> HandedOff
    -> Aborted
```

`Rejected`, `Aborted` и `HandedOff` являются terminal states Handshake. Эта модель концептуальна и не требует Go enum.

### Переходы при Shutdown

Runtime Host владеет Admission Gate. Handshake получает через composition bridge только живую read-only admission capability; он не может открыть или закрыть Gate и не зависит от Runtime Host.

Handshake проверяет эту capability до Authentication, поэтому request, поступивший при закрытом admission, не запускает проверку credentials. Та же capability проверяется повторно непосредственно перед `websocket.Accept`; успешная финальная проверка admission является linearization point входа в `Committing`.

В начале shutdown Host закрывает Gate до ожидания активной работы:

- `Evaluating` отменяется и завершается как `Aborted`; Evaluation, вернувшая Allow после этого, не может выполнить commit.
- `Allowed` не может перейти в `Committing` после закрытия gate и становится `Aborted`.
- Вход в `Committing` требует атомарной успешной финальной проверки admission непосредственно перед `websocket.Accept`.
- Работа, уже находящаяся в `Committing`, должна завершиться ошибкой и выполнить cleanup либо достичь `Upgraded` и завершить контролируемый handoff.
- Upgraded connection должна войти в Runtime shutdown wait set до завершения handoff; иначе Upgrade boundary закрывает её до завершения shutdown.
- Работа в `HandedOff` принадлежит Session lifecycle и участвует в Runtime shutdown.

Таким образом, shutdown не может завершиться, пока admitted connection находится вне Upgrade boundary и одновременно вне Runtime shutdown tracking.

## 13. Модель Ownership

### Listening Socket

Listener единолично владеет listening socket от успешного `net.Listen` до его закрытия в Listener Stop.

### Accepted HTTP Connection

Transport `net/http` владеет каждым accepted HTTP connection до успешного hijack или Upgrade. Handshake Evaluation не владеет и не закрывает raw TCP connection. При HTTP rejection переиспользование или закрытие определяется семантикой `net/http`.

### HTTP Request и ResponseWriter

HTTP server владеет lifecycle request. Handshake синхронно заимствует request и ResponseWriter. Evaluation никогда не записывает response.

До Accept transport/Handshake executor может применить один HTTP rejection или failure response. После commit любого response верхний слой не записывает другой response и не пытается выполнить Upgrade.

### Upgrade Boundary

После успешного `websocket.Accept` Upgrade boundary единолично владеет WebSocket. Ownership сохраняется, пока создаётся Runtime connection context и Session подготавливается к acceptance.

### Session Ownership Acceptance

Концептуальная точка handoff наступает только после выполнения всех условий:

1. Session успешно создана.
2. Runtime-owned connection context существует.
3. Session внесена в Runtime shutdown wait set.
4. Session явно приняла ответственность за WebSocket и его закрытие.

До этой точки любой failure закрывается Upgrade boundary. В этой точке ownership ровно один раз переходит к Session. После неё Session является единственным closer. Downstream error после acceptance является Session failure и никогда не вызывает повторный Close верхним уровнем.

[DP-004](DP-004-per-session-execution-boundary.md) является нормативным уточнением этой точки. Его единственный synchronous atomic handoff одновременно публикует construction Session, Runtime-owned connection context, shutdown accounting и acceptance ownership transport. `accepted=false` означает, что transfer не произошёл и Upgrade boundary очищает WebSocket; `accepted=true` означает, что transfer произошёл и Upgrade boundary никогда не очищает его.

Общий ownership contract `AuthenticatedDispatcher` позволяет реализации вернуть error при любом ownership result: только `accepted` определяет ownership cleanup, и Handshake никогда не возвращает transport после `accepted=true`. Поэтому generic Handshake tests для `accepted=true, error!=nil` остаются корректными tests ownership boundary для произвольных реализаций этого interface.

Целевой production Session Dispatcher нормативно сужен в [DP-004](DP-004-per-session-execution-boundary.md): каждый pre-Commit failure возвращает `accepted=false` с безопасной error, successful Commit возвращает только `accepted=true, nil`, а post-Commit failures принадлежат terminal accounting и, если известны до построения immutable Terminal Result, его единственному invocation Terminal Observer. Целевая реализация не создаёт `accepted=true, error!=nil`; это production-ограничение не меняет общий ownership contract Handshake.

### Authentication Result и Principal

Authentication result data ограничены request и не владеют transport. Principal копируется через commit boundary, чтобы downstream mutation не могла изменить evaluated identity. Credentials и resolved Secrets никогда не включаются.

## 14. Семантика `websocket.Accept`

`websocket.Accept` не является пассивным transport conversion. Библиотека:

- проверяет требования WebSocket transport и protocol;
- может применять default Origin validation;
- при rejection может записать и выполнить commit HTTP response;
- создаёт WebSocket только после успешного handshake.

Executor различает три фазы:

### Pre-Accept Rejection

Authentication Reject, cancellation, нарушение readiness invariant или другой pre-commit failure происходит до Accept. Executor может записать один подходящий HTTP response. WebSocket не существует.

### Accept Rejection

Accept отклоняет protocol, method, headers или default Origin conditions и может уже записать HTTP response. Executor фиксирует terminal category и не выполняет вторую HTTP-запись.

### Post-Accept Failure

Accept завершился успешно и WebSocket существует, но создание connection context, регистрация shutdown, создание Session или handoff завершились ошибкой. HTTP semantics больше недоступны. Текущий владелец WebSocket закрывает connection согласно ownership rules.

В первой реализации Origin остаётся transport/library concern. Перенос явной настраиваемой Origin Policy до Accept требует отдельного сфокусированного DP.

## 15. Отключённая Authentication

Отключённая Authentication создаёт explicit anonymous Principal. Principal остаётся обязательным для каждого Allow Decision и каждой Session.

Anonymous Principal явно отмечен как anonymous и unauthenticated и не содержит вымышленных credentials, roles или claims. Это сохраняет intent Configuration: `authentication.enabled: false` допускает клиентов — и предоставляет downstream-компонентам единую форму identity. Также это исключает optional identity checks во всех Session, Router и будущем delivery code.

Текущей реализации Session может потребоваться внутренняя адаптация для принятия explicit anonymous identity; документ не определяет новый публичный API.

## 16. Timeout и Cancellation

Handshake имеет bounded lifecycle. Настроенное значение `HandshakeSeconds` задаёт максимальную длительность, а request cancellation или Runtime shutdown могут завершить её раньше.

Runtime readiness validation отклоняет отсутствующий, нулевой, unsupported или находящийся вне диапазона effective Handshake timeout до Listener Start. Unbounded fallback отсутствует, request processing не придумывает default.

Cancellation распространяется в Authentication Service, Providers и их dependencies. Dependency, игнорирующую context, нельзя принудительно остановить средствами Go context cancellation. Её операция может пережить caller; Runtime не должен классифицировать это ограничение как successful admission, а bounded shutdown для такой dependency нельзя гарантировать без dependency-specific isolation mechanism вне этого DP.

## 17. Error Model

Следующие минимальные категории должны оставаться различимыми внутри Runtime:

- **rejection** — ожидаемый отрицательный Authentication result, а не operational error;
- **cancellation или deadline** — request, настроенный Handshake timeout или Runtime shutdown завершили Evaluation;
- **dependency unavailable** — dependency Provider не смогла обслужить request;
- **invalid Runtime configuration** — нарушен readiness invariant;
- **protocol или Upgrade rejection** — Accept отклонил HTTP/WebSocket handshake;
- **internal failure** — неожиданный дефект Handshake или transport executor;
- **handoff или Session failure** — ошибка после успешного Accept при подготовке или работе Session.

Документ не предписывает Go error types или точные client response bodies. Client-facing responses остаются обобщёнными и никогда не содержат credentials или internal details.

Execution сохраняет исходный Authentication outcome. Например, transport failure после Allow не становится credential rejection, а Accept Origin rejection не становится dependency failure.

## 18. Ownership Operational Errors

Transport/Handshake executor является единственным владельцем terminal outcome одного Handshake. Он получает или наблюдает итоговый результат Evaluation, response, Accept и pre-handoff execution и сообщает одну безопасную terminal category.

После Session ownership acceptance последующими failures владеет Session lifecycle. Listener lifecycle владеет terminal failures HTTP server `Serve`, которые не являются normal shutdown.

Diagnostics sink, event schema, metrics и logging integration требуют отдельного design. До этого terminal Dispatcher, Handshake и неожиданные `Serve` errors запрещено молча игнорировать. Компоненты могут оборачивать или распространять errors, но не должны независимо создавать дублирующие terminal reports для одного Handshake.

## 19. Configuration и Readiness

До Listener Start Runtime bootstrap проверяет и компонует:

- совместимость Published Snapshot;
- настроенный Handshake timeout;
- Authentication Service;
- enabled Provider factories и Provider metadata;
- SecretResolver и другие обязательные dependencies;
- Host-owned Admission Gate, read-only Runtime capabilities, composition bridge и transport/Handshake executor.

Ошибка предотвращает Listener Start. Request-time configuration failure не является штатной моделью; если invariant всё же нарушен, это internal или invalid Runtime configuration failure, приводящий к fail closed.

Request path использует только уже скомпонованные dependencies и effective values. Он не читает Runtime Snapshot, ConfigurationVersion, состояние management API или Control Plane repositories.

Composition bridge следует [ADR-0004](../adr/0004-handshake-runtime-dependencies.md): Runtime composition предоставляет Handshake живой read-only доступ к admission permission и Runtime context. Handshake не импортирует Runtime Host и не управляет им. Этот DP определяет использование capabilities, не проектируя Host APIs.

## 20. Связь с Session Manager

Будущий Session Manager не участвует в Authentication Evaluation. Его значимая архитектурная обязанность находится в handoff: новая Session должна войти в Runtime shutdown wait set до завершения Session ownership acceptance.

После acceptance lifecycle Session, включая shutdown cancellation и completion, не зависит от HTTP request. Документ не определяет интерфейсы Session Manager, storage, lookup, grouping или routing behavior.

## 21. Объём первой реализации

Первая реализация включает только:

- нормализацию request metadata, необходимую существующей Authentication;
- bounded pre-Upgrade Authentication;
- одноразовый allow, reject или operational outcome;
- admission gate и координацию shutdown;
- контролируемый Accept и ownership handoff;
- explicit anonymous Principal при отключённой Authentication;
- безопасное распространение terminal errors.

Она не добавляет общий список evaluators. Origin остаётся под управлением WebSocket-библиотеки. Rate Limit, Maintenance, IP rules и другие admission checks не являются гарантированными extension points текущего контракта.

Общий policy framework отложен до появления нескольких реальных use cases, подтверждающих общие ordering, data, ownership и failure semantics. Каждая будущая возможность следует Configuration First и требует focused design.

## 22. Стратегия миграции

1. Сохранить characterization coverage текущей обработки method, Upgrade, rejection, cleanup и продолжения работы Listener.
2. Добавить startup readiness checks для timeout, Authentication, Providers и dependencies до Listener Start.
3. Внедрить Host-owned admission и Runtime context capabilities через composition bridge, затем ввести request-scoped Handshake context без изменения Provider logic.
4. Выполнять существующую Authentication до Accept и создавать одноразовый Decision.
5. Реализовать explicit anonymous Principal при отключённой Authentication.
6. Применять pre-Accept rejection ровно один раз и учитывать responses, committed внутри Accept.
7. Создавать Runtime-owned connection context после успешного Accept из Runtime context capability.
8. Регистрировать будущую Session в Runtime shutdown tracking до explicit ownership acceptance.
9. Удалить устаревший post-Upgrade Authentication path, чтобы каждый request проходил Authentication не более одного раза.
10. Сохранять и распространять terminal Handshake и неожиданные Serve errors без logging sensitive input.

Каждый шаг сохраняет сборку вертикали и не вводит Router, API Session Manager, concrete Host dependency в Handshake или generic policy framework.

## 23. Оставшиеся риски

- Dependencies, игнорирующие context, могут пережить bounded Handshake caller.
- Предположения о proxy и trusted address остаются неопределёнными.
- Неправильное отслеживание response commit всё ещё может вызвать попытку второй записи.
- Неправильная реализация admission gate может допустить работу во время shutdown.
- Shutdown tracking Session зависит от отдельных эпиков Runtime Host и Session Manager.
- Operational diagnostics остаются неполными до принятия focused diagnostics design.

## 24. Открытые вопросы

- Какие точные HTTP statuses, bodies и challenge headers применяются для каждого authentication rejection?
- Как нормализуются trusted proxy data, remote address, TLS state и forwarded headers?
- Какие Configuration и matching semantics будет использовать будущая explicit Origin Policy?
- Нужен ли Provider отдельный результат «not applicable» помимо rejection?
- Какие bounded Handshake events и metrics необходимы и какие fields безопасны?
- Какой diagnostics sink принимает terminal Runtime errors?
- Какое поведение переиспользования HTTP connection ожидается после каждого pre-Accept rejection?

Anonymous identity, общий Handshake timeout, startup readiness, shutdown admission и ownership transfer больше не являются открытыми вопросами этого DP.

## 25. Трассировка Findings

| Finding | Статус | Разрешение |
|---|---|---|
| F-01 — Конфликт request context и lifecycle Session | Resolved | Session использует отдельный Runtime-owned context; request context завершается вместе с Handshake. |
| F-02 — Shutdown не атомарен относительно admission | Resolved | Host владеет Gate; Handshake использует его read-only capability до Authentication и непосредственно перед Accept. Commits после shutdown запрещены, а in-flight commits входят в wait set или выполняют cleanup. |
| F-03 — Для TCP connection указаны два владельца | Resolved | Listening socket, accepted HTTP connection, Upgrade boundary и Session ownership разделены. |
| F-04 — Классификация Accept противоречит поведению библиотеки | Resolved | Разделены pre-Accept rejection, Accept rejection и post-Accept failure; committed библиотекой responses не переписываются. |
| F-05 — Точка handoff не определена | Resolved | Ownership acceptance наступает только после construction, создания Runtime context и регистрации shutdown. |
| F-06 — Disabled Authentication не создаёт Session | Resolved | Disabled Authentication создаёт explicit anonymous Principal. |
| F-07 — Handshake timeout не определён | Resolved | Effective Handshake timeout обязателен, проверяется до start и объединяется с request и shutdown cancellation. |
| F-08 — Инварианты Decision неполны | Resolved | Явно определены Allow, Reject, operational failure, cancellation validity и one-use semantics. |
| F-09 — Категории ошибок не гарантированы | Clarified | Минимальные категории должны оставаться различимыми без определения Go error types. |
| F-10 — У terminal operational error нет владельца | Clarified | Handshake executor владеет terminal outcome Handshake; Listener и Session владеют errors после своих границ. |
| F-11 — Будущая расширяемость не доказана | Deferred | Первый scope ограничен Authentication; Origin и другие policies требуют focused DP и не обещаны как extension points. |
| F-12 — Владелец Configuration validation неоднозначен | Resolved | Runtime bootstrap readiness предшествует Listener Start; request path не читает Snapshot или Control Plane. |
| F-13 — Session Manager зависит от неопределённого handoff | Clarified | Session входит в Runtime shutdown wait set до ownership acceptance; API Session Manager не проектируется. |
| F-14 — Pipeline может стать преждевременным policy framework | Deferred | Framework evaluators не вводится; общая policy-механика ждёт нескольких подтверждённых use cases. |

## 26. Связанные документы

### ARCH-001

Design применяет [ARCH-001](../architecture/ARCH-001-runtime-architectural-pattern.md): request metadata образуют Context, Authentication выполняет Evaluation, admission создаёт Decision, а transport/Handshake executor выполняет effect. Ownership, cancellation и lifecycle явны.

### Runtime Alpha Review

Design устраняет границу pre-Upgrade Authentication, отмеченную в [Runtime Alpha Architecture Review](../reviews/runtime-alpha-review.md). Композиция Runtime Host и более широкое lifecycle hardening остаются отдельной работой.

### MASTER_PLAN

[Master Engineering Plan](../roadmap/MASTER_PLAN.md) определяет Handshake как первый Beta foundation gate перед Router. Этот DP определяет gate, не превращаясь в implementation backlog.

### Существующие Decisions и Proposals

- [ADR-0002](../adr/0002-configuration-dsl.md) требует Configuration First и явного отклонения unsupported effective behavior.
- [ADR-0003](../adr/0003-runtime-architecture.md) требует component responsibility, dependency injection и transport-neutral Authentication.
- [ADR-0004](../adr/0004-handshake-runtime-dependencies.md) определяет read-only capability bridge от Host-owned admission и Runtime context state к Handshake.
- [DP-001: Authentication](../proposals/DP-001-authentication.md) определяет identity evaluation до admission и HTTP rejection.
- [DP-004](../proposals/DP-004-authentication-runtime-contracts.md) определяет направление transport-neutral Authentication contracts.

## 27. Решение

Выбранным направлением остаётся **Alternative D: Dedicated Handshake Boundary**.

Handshake получает только живые read-only Runtime capabilities и не зависит от Runtime Host. Он проверяет Host-owned Admission Gate до Authentication и выполняет финальную проверку admission непосредственно перед `websocket.Accept`. Настроенная Authentication выполняется до WebSocket Accept. Одноразовый Allow Decision с допустимым authenticated или explicit anonymous Principal может войти в admission commit только после успешной финальной проверки. Библиотека может отклонить request и выполнить commit HTTP response во время Accept. После успешного Accept Upgrade boundary владеет WebSocket, пока Session не зарегистрирована для Runtime shutdown и явно не приняла ownership. Session создаётся с Runtime-owned context, полученным через Runtime context capability и не зависящим от HTTP request.

Это остаётся концептуальной последовательностью подсистемы, а не Universal Policy Engine, generic framework, объявлением package или фиксированным Go API. Первая реализация поддерживает только Authentication. Будущее admission behavior требует собственной Configuration и focused design.

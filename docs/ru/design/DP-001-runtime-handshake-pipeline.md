# DP-001: Runtime Handshake Pipeline

[English version](../../en/design/DP-001-runtime-handshake-pipeline.md)

## 1. Status

**Статус:** Draft

Документ предлагает архитектуру Runtime Handshake. Он не содержит implementation contract, Go interface, объявления package или обязательства по API будущих policies.

## 2. Background

Alpha Runtime имеет работающую WebSocket-вертикаль. Listener владеет TCP listener и HTTP server, обрабатывает `GET /ws`, вызывает WebSocket library для принятия Upgrade, а затем передает upgraded connection через Dispatcher. AuthenticationDispatcher преобразует выбранные metadata HTTP request в transport-neutral `AuthenticationRequest`, вызывает Authentication Service и либо закрывает upgraded connection, либо передает его вместе с Principal в Session Dispatcher. После этого Session владеет WebSocket connection и запускает read loop.

Текущий путь запроса:

```text
TCP accept
    -> HTTP request
    -> GET /ws handler
    -> WebSocket Upgrade (101)
    -> ConnectionContext with WebSocket connection
    -> AuthenticationDispatcher
    -> Authentication Service and Providers
    -> reject with WebSocket close
       or
       AuthenticatedContext
    -> Session creation and lifecycle
```

Этот путь подтвержден реализацией и integration tests. [Runtime Alpha Architecture Review](../reviews/runtime-alpha-review.md) фиксирует Authentication после Upgrade как High finding F-01. Ревью также выделяет composition Runtime Host и lifecycle hardening как связанную работу, которую нельзя скрывать внутри Handshake.

[DP-001: Authentication](../proposals/DP-001-authentication.md) уже устанавливает, что обязательная проверка identity завершается до открытия WebSocket Session, а отклоненные credentials приводят к HTTP `401`. Реализация пока не соответствует этому design.

## 3. Problem Statement

Authentication после Upgrade является архитектурной проблемой, а не только ограничением HTTP status.

### Ownership

После успешного `websocket.Accept` HTTP exchange завершен и существует upgraded mutable connection. AuthenticationDispatcher должен владеть и закрывать это соединение даже для клиентов, которых не следовало допускать. Таким образом, security decision принимается после перехода transport ownership через admission boundary.

### Transport Semantics

До Upgrade Runtime может выразить отказ через HTTP response. После Upgrade тот же outcome приходится кодировать как WebSocket close. Клиент наблюдает успешный `101` с последующим закрытием, а не отклоненный handshake. Semantics Authentication начинают зависеть от момента преобразования transport.

### Lifecycle

Evaluation после Upgrade создает WebSocket resources до завершения admission. Cancellation, Provider failure, downstream construction failure и Listener shutdown должны закрывать соединение, которое не требовалось создавать. Connection-specific lifecycle начинается слишком рано.

### Observability

Принятый transport с последующим policy rejection сложнее классифицировать. Upgrade success, Authentication rejection, Provider outage и Session startup failure могут смешиваться, если каждый последующий close интерпретируется только через скрытое знание pipeline.

### Extensibility

Будущие решения Origin, maintenance, rate limit или IP filtering столкнутся с тем же выбором: выполняться после Upgrade с лишним выделением ресурсов либо встраиваться непосредственно в Listener. Ни один вариант не создает стабильную admission boundary.

### Future Policies

Размещение каждой policy на отдельной transport stage сделает ordering, short-circuit behavior, ownership и error mapping непредсказуемыми. Нужна концептуальная граница Handshake, но она не должна превращаться в universal Policy Engine.

## 4. Existing Architecture

```text
Listener
  owns TCP listener and HTTP server
        |
        v
HTTP /ws Handler
  validates method
        |
        v
websocket.Accept
  commits HTTP 101 and creates WebSocket connection
        |
        v
Connection Dispatcher
  transfers upgraded transport ownership
        |
        v
Authentication Dispatcher
  copies HTTP metadata
  invokes transport-neutral Authentication
  closes on rejection/error
        |
        v
Authenticated Dispatcher
  transfers connection and Principal
        |
        v
Session
  owns WebSocket, read loop, and serialized writes
```

Следующие обязанности необходимо сохранить:

- Listener владеет network acceptance и выполнением HTTP/WebSocket.
- Authentication владеет credential evaluation и остается независимой от `net/http` и WebSocket.
- Connection handoff делает передачу mutable transport ownership явной.
- Session создается только для аутентифицированного Principal и после этого владеет принятым WebSocket.
- Runtime Snapshot предоставляет effective Configuration без доступа к Repository.

Измениться должен порядок Upgrade и Authentication.

## 5. Goals

- Завершать обязательную Authentication до WebSocket Upgrade.
- Сохранить transport-neutral AuthenticationRequest, Authentication Service и Providers.
- Сохранить независимость Runtime от Control Plane repositories и management HTTP API.
- Предоставить ясное концептуальное место для будущей Evaluation на этапе handshake.
- Сделать ownership HTTP и WebSocket resources явным для каждого outcome.
- Сохранить предсказуемые connection и Session lifecycles при rejection, cancellation и failure.
- Различать отрицательные admission decisions, internal failures и dependency failures.
- Сохранить ответственность Listener за transport execution без Provider-specific logic.
- Остаться совместимым с текущей вертикалью Snapshot, Authentication и Session.

## 6. Non Goals

Документ не проектирует:

- Router;
- Session Manager;
- Middleware;
- Plugin ABI;
- другой transport или transport abstraction;
- QUIC;
- HTTP/3;
- OAuth;
- реализацию mTLS;
- Provider-specific credential verification;
- API Origin, rate limit, maintenance или IP filtering;
- детали composition Runtime Host.

## 7. Constraints

Design следует [ARCH-001: Runtime Architectural Pattern](../architecture/ARCH-001-runtime-architectural-pattern.md).

### Configuration First

Поведение Handshake происходит из Published Configuration через immutable Runtime Snapshot. Unsupported active settings явно отклоняются, а не игнорируются.

### Ownership

У каждого mutable transport resource есть один текущий владелец. Передача явна, а failure до или во время передачи имеет определенного владельца закрытия.

### Lifecycle

Evaluation учитывает cancellation запроса. Длительная работа и network I/O не выполняются под lifecycle locks. Handshake завершается до начала lifecycle Session.

### No Magic

Зависимости компонуются явно. Handshake не обнаруживает Providers, не читает repositories, не использует global state или `context.Value` как service locator.

### Boring Core

Design повторно использует существующие границы Authentication и transport. Он не создает generic pipeline framework, universal Policy Engine или общую Decision model для несвязанных подсистем.

## 8. Design Alternatives

### Alternative A — Keep the Current Architecture

Authentication остается после `websocket.Accept` в AuthenticationDispatcher.

**Преимущества:**

- Не требуется изменение реализации.
- Существующие integration tests и ownership transfer сохраняются.
- Authentication имеет доступ к нормализованным request metadata, уже связанным с WebSocket connection.

**Недостатки:**

- Отклоненные credentials получают `101` до закрытия.
- Неаутентифицированные клиенты потребляют ресурсы upgraded connection.
- Admission ownership начинается до admission decision.
- Будущие handshake policies также будут выполняться слишком поздно или потребуют отдельного размещения.
- Вариант противоречит существующему Authentication proposal и finding Alpha review.

**Причины отклонения:** Вариант сохраняет именно ту границу security и lifecycle, которую должна исправить TASK-BETA-003.

### Alternative B — Put Authentication Logic Inside Listener

Listener напрямую читает Configuration Authentication, выбирает Providers, проверяет credentials и решает, вызывать ли Upgrade.

**Преимущества:**

- Authentication может выполняться до Upgrade.
- HTTP rejection реализуется напрямую.
- В request path меньше видимых компонентов.

**Недостатки:**

- Listener получает знания о JWT, API Key, Basic, Registry и оркестрации Provider.
- Authentication связывается с поведением HTTP и WebSocket transport.
- Расширение Provider требует изменения Listener.
- Теряются unit boundaries и dependency direction из ADR-003.

**Причины отклонения:** Вариант исправляет ordering ценой нарушения single responsibility, transport-neutral Authentication и Boring Core.

### Alternative C — Put Authentication Inside Session

Listener выполняет Upgrade и создает provisional Session, которая проводит Authentication до обычной обработки сообщений.

**Преимущества:**

- Session могла бы централизовать lifecycle соединения после Upgrade.
- Authentication и последующая обработка Message имели бы одного владельца connection.
- Listener оставался бы независимым от Providers.

**Недостатки:**

- Authentication по-прежнему происходит после Upgrade.
- Session существовала бы до валидного Principal, нарушая текущий инвариант.
- Session потребовались бы HTTP handshake metadata или другой скрытый transport bridge.
- Rejection остается WebSocket close и не может быть HTTP response.
- Ответственность Session расширяется до admission policy.

**Причины отклонения:** Вариант переносит проблему глубже в Runtime и нарушает установленную границу authenticated Session.

### Alternative D — Dedicated Handshake Pipeline

Концептуальная граница Handshake нормализует request metadata, выполняет configured Evaluation, формирует admission Decision и только затем просит владельца transport отклонить запрос или выполнить Upgrade.

**Преимущества:**

- Authentication завершается до выделения WebSocket.
- Authentication Service и Providers остаются transport-neutral.
- Listener сохраняет выполнение HTTP/WebSocket без Provider-specific logic.
- Ownership transfer происходит только после allow Decision.
- Будущие evaluations handshake имеют одну упорядоченную границу.
- Отрицательные decisions и operational errors могут иметь разную HTTP semantics.

**Недостатки:**

- Request path получает явную orchestration stage.
- Migration должна исключить duplicate Authentication и двойное владение.
- Ordering и failure semantics будущих evaluations требуют сфокусированного design.
- Медленные evaluations удерживают ресурсы HTTP request до завершения.

**Причины выбора:** Вариант исправляет security boundary, сохраняя текущие обязанности и extension seams. Термин Pipeline обозначает концептуальную последовательность, а не generic framework или объявление Go interfaces.

## 9. Proposed Architecture

```text
Runtime Snapshot and composed dependencies
                 |
                 v
Listener / HTTP transport owner
  accepts request and owns ResponseWriter
                 |
                 v
Handshake Context normalization
  copies only required metadata
                 |
                 v
Handshake Evaluation
  required method/path checks
  configured Authentication
  future handshake-scoped checks
                 |
                 v
Handshake Decision
          /                 \
         v                   v
reject or fail              allow
HTTP response                 |
no WebSocket                  v
                        WebSocket Upgrade
                              |
                              v
                    authenticated handoff
                              |
                              v
                           Session
```

Архитектура концептуальна. Она не предписывает Handshake package, interface, method или универсальный список evaluators.

Listener остается transport executor. Он предоставляет request metadata для нормализации и применяет итоговый transport effect. Authentication остается существующим transport-neutral service. Успешный результат предоставляет данные Principal для handoff после Upgrade. Назначение Session не меняется: она создается только после успешного admission и владеет WebSocket connection.

Будущие Handshake policies могут включаться в Evaluation stage только после появления их Configuration и сфокусированного design. Они не становятся generic Runtime policies автоматически.

## 10. Handshake Stages

```text
Transport
    -> Handshake Context
    -> Evaluation
    -> Decision
    -> Upgrade
    -> Session
```

### Transport

Listener и HTTP server принимают request. Они владеют transport execution и раскрывают только request data, необходимые Handshake.

### Handshake Context

Необходимые metadata копируются из HTTP request в узкий input для Evaluation. Значения, содержащие credentials, остаются request-scoped и не сохраняются в Session или Snapshot.

### Evaluation

Evaluation применяет настроенные admission checks. Authentication использует существующие Authentication Service и Providers. Evaluation — архитектурный термин ARCH-001, а не обязательный framework, engine или Go abstraction.

### Decision

Результат различает allow, expected reject, Configuration failure, dependency failure, cancellation и internal failure в объеме, необходимом подсистеме. Документ не определяет единый global Decision type или его точное представление.

### Upgrade

Только allow Decision разрешает WebSocket Upgrade. Transport operation выполняет Listener. Failure во время Upgrade является transport failure, а не Authentication rejection.

### Session

После успешного Upgrade соединение и скопированный аутентифицированный Principal передаются следующему компоненту. Затем создается Session и начинается lifecycle connection.

## 11. Ownership Model

### HTTP Request

HTTP server владеет lifecycle request. Handshake path синхронно заимствует request и копирует только metadata, необходимые Evaluation. Он не должен сохранять `http.Request`, Body, Headers, URL или cancellation context после завершения handler. Authentication Providers по-прежнему получают нормализованные данные, а не объект request.

### ResponseWriter

HTTP handler заимствует ResponseWriter у HTTP server и является единственным компонентом Handshake, которому разрешено применять HTTP Decision. Evaluation не пишет responses. До Upgrade handler может отправить один rejection или failure response. После commit response выполнять Upgrade запрещено.

После успешного Upgrade HTTP response semantics больше недоступны. Handler не должен пытаться выполнить еще одну HTTP write.

### TCP Connection

Listener и HTTP server владеют принятым TCP connection до Upgrade. Handshake Evaluation не получает доступ к raw connection и не закрывает его. При HTTP rejection обычная semantics HTTP server определяет возможность повторного использования или закрытия transport. При shutdown Listener server остается ответственным за активные requests до Upgrade.

### WebSocket Connection

Во время Evaluation WebSocket connection не существует. Успешный Upgrade создает его на границе transport execution. Эта граница временно владеет соединением и должна:

- ровно один раз передать его authenticated downstream path; либо
- закрыть его, если handoff или создание Session завершается ошибкой.

Authentication rejection не может владеть или закрывать WebSocket, поскольку он не был создан.

### Authentication Result

Authentication Service формирует result data во время Evaluation. Handshake path владеет результатом на протяжении request. Значения credentials и resolved Secrets не входят в result. При успехе Principal копируется в allow Decision или authenticated handoff, чтобы downstream mutation не могла изменить оцененную identity.

Result не владеет transport resource. Отрицательный результат приводит к HTTP rejection без передачи connection.

### Session

Session не существует до allow Decision и успешного Upgrade. После успешного создания Session становится единственным владельцем WebSocket connection, read loop и serialization writes. Если создание Session завершается ошибкой, connection закрывает владелец handoff до Session.

Session хранит Principal и минимальные connection metadata, необходимые ее контракту. Она не сохраняет HTTP request, ResponseWriter, credential Headers, Query или Handshake Context.

### Ownership Timeline

```text
HTTP server owns TCP and request
        |
        | Handshake borrows request/ResponseWriter
        | Evaluation owns copied metadata and result values
        v
allow Decision
        |
        | transport boundary creates and temporarily owns WebSocket
        v
authenticated handoff
        |
        | Session construction succeeds
        v
Session owns WebSocket until Stop
```

Каждый failure path завершается до следующей передачи или закрывает ресурс, принадлежащий текущему владельцу.

## 12. Error Model

### Before Upgrade

До Upgrade failures используют HTTP semantics и не создают WebSocket connection.

- Invalid request shape или unsupported method является ошибкой HTTP request.
- Rejected credentials являются ожидаемым отрицательным Authentication decision и приводят к безопасному authentication rejection согласно существующему Authentication proposal.
- Будущий rejection Origin или admission policy является ожидаемым отрицательным decision с подходящей HTTP semantics.
- Cancellation останавливает Evaluation и не представляется успешным admission.
- Provider или dependency failure является operational, observable и отличается от rejected credentials.
- Internal failure возвращает обобщенный response без implementation details.
- Configuration failure обычно должен предотвращать startup Listener. Если он обнаружен на request, система fail closed и сообщает об ошибке operationally.

Точные response bodies, challenge headers и status mapping сверх уже принятого поведения требуют implementation-focused design. Sensitive values никогда не включаются в response.

### During Upgrade

Failure при Upgrade является transport failure и не переклассифицируется в Authentication rejection. Listener сохраняет ответственность за безопасную диагностику и cleanup согласно контракту WebSocket library.

### After Upgrade

HTTP errors больше невозможны. Failures во время authenticated handoff, создания Session или выполнения Session используют WebSocket close либо немедленный transport cleanup согласно lifecycle компонента. Они не должны представляться как policy rejection до Upgrade.

### Error Separation

Expected negative Decision, Configuration failure, dependency failure, cancellation, internal failure, Upgrade failure и Session failure остаются различимыми для observability, даже если client-facing details намеренно обобщены.

## 13. Lifecycle

Общие Runtime и Listener уже находятся в Running до поступления request в Handshake. Handshake является ограниченным per-request lifecycle внутри HTTP handler:

```text
request accepted
    -> context normalized
    -> Evaluation started
    -> Decision produced
       -> rejected/failed: HTTP response, Handshake ends
       -> allowed: Upgrade attempted
          -> failed: transport cleanup, Handshake ends
          -> succeeded: authenticated handoff
             -> Session created and started
             -> Handshake ownership ends
```

HTTP rejection остается возможным до commit response или Upgrade. Evaluation не должна выполнять ни то ни другое.

Connection-specific lifecycle Session начинается только после allow, успешного Upgrade и успешной передачи. Read loop Session никогда не выполняется одновременно с незавершенной Handshake Evaluation для того же connection.

Shutdown Listener отменяет contexts активных Handshake и предотвращает новые admissions. Evaluation должна распространять cancellation в Authentication Providers. Ожидания shutdown и error propagation следуют lifecycle requirements ARCH-001 и debt, зафиксированному Alpha finding F-04.

## 14. Security

### Origin

Текущая обработка WebSocket origin использует defaults библиотеки. Будущая явная Origin Policy располагается до Upgrade в Handshake Evaluation и следует Configuration First. Документ не определяет ее schema, proxy model или matching rules.

### Credentials

В AuthenticationRequest копируются только необходимые request fields с credentials. Providers разрешают Secret References близко к моменту использования. Raw credentials, tokens, API keys, passwords и resolved Secrets не сохраняются после Evaluation.

### Logging and Observability

Logs, metrics, traces и errors не должны содержать credential Headers, tokens, credentials из query, значения Secret, private key material или полный request dump. Operational signals различают rejection и failure через безопасные categories и bounded labels.

### Sensitive Data Lifetime

Handshake metadata и Authentication results ограничены request. Ownership secret bytes следует контрактам SecretResolver и cleanup Provider. Session получает identity data Principal, а не входные credentials.

### Configuration

Handshake управляется только данными Published Snapshot. Runtime не читает Control Plane repositories. Unsupported Authentication или будущие admission settings по возможности приводят к явной ошибке до приема traffic.

## 15. Future Extension Points

Предлагаемая граница в будущем может включать сфокусированные настроенные evaluations:

- Origin Policy;
- maintenance admission;
- rate limiting;
- IP filtering;
- enterprise-specific admission extensions.

Это будущие возможности, а не реализованные функции или обязательства по API. Каждая требует Configuration metadata, validation, явных ordering и failure semantics и сфокусированного design. Их нельзя объединять в Universal Policy Engine.

Extensions получают только минимальный Handshake Context для своей ответственности. Они не получают произвольный Runtime state, ownership raw transport, доступ к Repository или право самостоятельно выполнять Upgrade.

## 16. Migration Strategy

Migration следует выполнять небольшими проверяемыми шагами:

1. Добавить characterization tests для текущей обработки HTTP method, успешного Upgrade, Authentication rejection после Upgrade, cleanup connection и продолжения работы Listener.
2. Ввести pre-Upgrade orchestration seam в HTTP request path без изменения Authentication Service, Providers, Session или Snapshot.
3. Нормализовать существующий AuthenticationRequest из HTTP metadata до `websocket.Accept` и проверить copying, cancellation и sensitive-data handling.
4. Вызывать существующий Authentication Service до Upgrade и преобразовывать expected rejection в безопасный HTTP rejection.
5. Передавать скопированный успешный Principal через Upgrade в существующий authenticated handoff Session.
6. Поручить границе Upgrade закрывать connection при ошибке handoff после Upgrade.
7. Удалить устаревший путь Authentication после Upgrade, чтобы каждый request оценивался ровно один раз.
8. Добавить integration tests для rejection без `101`, Provider failure, cancellation, успешного Upgrade, startup Session, concurrent requests и shutdown Listener.
9. Добавить безопасный operational reporting этапов Handshake без логирования credentials.
10. Рассматривать дополнительные admission evaluations только в отдельных задачах после стабилизации пути Authentication.

Каждый шаг сохраняет сборку существующей вертикали и исключает одновременное перепроектирование Router, Session Manager, Runtime Host или Provider contracts.

## 17. Risks

- HTTP response может быть committed до завершения всей необходимой Evaluation, после чего корректно применить Decision невозможно.
- Migration может случайно выполнять Authentication дважды или оставить активными старый и новый пути.
- Principal или request metadata могут ссылаться на mutable input через границу Upgrade.
- Неоднозначный ownership может привести к утечке connection или double close при ошибке handoff.
- Медленные или недоступные Providers могут удерживать ресурсы HTTP request до timeout или cancellation.
- Предположения о proxy и trusted address могут сделать Evaluation Origin или IP некорректной.
- Error mapping может раскрыть sensitive Provider details или скрыть operational failure как credential rejection.
- Shutdown Listener может конфликтовать с Evaluation или Upgrade без явных cancellation и transfer.
- Будущие checks могут добавляться без детерминированного ordering, снова создавая hidden policy behavior.
- Semantics disabled Authentication может остаться несовместимой с предложенной моделью anonymous Principal.

## 18. Open Questions

- Какой именно Principal создается при отключенной Authentication?
- Какие HTTP status, body и challenge behavior обязательны для каждой категории rejection?
- Какая timeout и cancellation policy применяется к отдельным Providers и общей Handshake Evaluation?
- Как нормализуются trusted proxy data, remote address, TLS state и forwarded headers?
- Какая Configuration model и matching semantics определяют Origin Policy?
- В каком порядке выполняются maintenance, rate limiting, IP filtering, Origin и Authentication?
- Может ли ожидаемый результат Provider «not applicable» продолжить проверку следующим Provider и как он отличается от rejection?
- Какие Handshake events и metrics необходимы и какие fields безопасны?
- На какой startup stage отклоняются недоступные Providers и unsupported admission settings?
- Какое поведение повторного использования HTTP connection ожидается после каждого rejection до Upgrade?

Эти вопросы намеренно отложены до implementation tasks или сфокусированных DP. Они не блокируют выбор архитектурной границы до Upgrade.

## 19. Relationship

### ARCH-001

Design применяет [ARCH-001](../architecture/ARCH-001-runtime-architectural-pattern.md): HTTP metadata образуют Context, настроенные admission checks выполняют Evaluation, allow/reject является Decision, а Listener как владелец transport выполняет HTTP rejection или Upgrade. Ownership и lifecycle остаются явными. Паттерн не превращается в generic framework.

### Runtime Alpha Review

Design напрямую отвечает на [High finding F-01](../reviews/runtime-alpha-review.md). Он сохраняет стабильные решения review и не заявляет, что F-02 composition Runtime Host или F-04 lifecycle hardening устранены одной документацией.

### MASTER_PLAN

[Master Engineering Plan](../roadmap/MASTER_PLAN.md) определяет Handshake как первый Beta foundation gate до Router. Этот DP задает архитектурное направление эпика, не меняя milestone criteria и не превращаясь в implementation backlog.

### Existing Decisions and Proposals

- [ADR-0002](../adr/0002-configuration-dsl.md) требует Configuration First и явного отклонения unsupported effective behavior.
- [ADR-0003](../adr/0003-runtime-architecture.md) требует component responsibility, dependency injection и transport-neutral Authentication.
- [DP-001: Authentication](../proposals/DP-001-authentication.md) устанавливает intent identity evaluation до соединения и HTTP rejection.
- [DP-004](../proposals/DP-004-authentication-runtime-contracts.md) задает направление transport-neutral Authentication contracts, сохраненное здесь.

## 20. Decision

Выбрано направление **Alternative D: Dedicated Handshake Pipeline**.

Обязательная настроенная Authentication выполняется до WebSocket Upgrade. Listener остается ответственным за transport effects HTTP и WebSocket, но не реализует Provider logic. Authentication остается transport-neutral. Allow Decision переносит скопированные identity data через Upgrade, после чего существующий authenticated handoff создает Session и передает ownership WebSocket.

Handshake Pipeline является концептуальной последовательностью подсистемы на основе ARCH-001. Это не Universal Policy Engine, generic framework, объявление нового package или фиксированный набор Go interfaces. Будущие admission checks требуют собственной Configuration и design. Implementation details остаются отдельной задачей.

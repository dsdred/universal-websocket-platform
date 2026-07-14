# Runtime Alpha Architecture Review

[English version](../../en/reviews/runtime-alpha-review.md)

**Статус:** Завершено  
**Дата ревью:** 2026-07-14  
**Оценка:** Ready with findings

## 1. Scope

Это ревью оценивает реализованную архитектуру Runtime перед разработкой Router. В область проверки входит production-код и тесты пакетов:

- `internal/runtimeconfig`;
- `internal/runtime`;
- `internal/listener`;
- `internal/connection`;
- `internal/authentication`;
- `internal/session`;
- `internal/message`;
- `internal/secretresolver`.

Реализация сопоставлена с [ADR-0002: Configuration DSL](../adr/0002-configuration-dsl.md), [ADR-0003: Runtime Architecture](../adr/0003-runtime-architecture.md), а также [DP-001](../proposals/DP-001-authentication.md), [DP-002](../proposals/DP-002-secret-references.md), [DP-003](../proposals/DP-003-jwt-provider.md) и [DP-004](../proposals/DP-004-authentication-runtime-contracts.md). Поведение Control Service, persistence и предлагаемое поведение Router находятся вне области ревью, кроме случаев, когда они задают границу Runtime.

Доказательная база — состояние репозитория на дату ревью, фактический граф импортов Go и полный набор тестов. Запланированное поведение не считается реализованным.

## 2. Current Runtime Vertical

Реализованный путь Configuration:

```text
Published ConfigurationVersion
    -> runtimeconfig.Builder
    -> immutable Runtime Snapshot
```

Реализованный путь соединения при сборке компонентов тестами или вызывающим кодом:

```text
TCP Listener
    -> HTTP Server
    -> GET /ws
    -> RFC 6455 Upgrade (101)
    -> ConnectionContext
    -> AuthenticationDispatcher
    -> Authentication Service
    -> ordered Provider(s)
    -> AuthenticatedContext
    -> Session Dispatcher
    -> Session read loop
    -> transport-neutral Message
    -> Message Handler (сейчас EchoHandler)
    -> Session.Send
    -> WebSocket frame
```

Эта вертикаль работает в integration tests, но не собирается production composition root. Сейчас `runtime.DefaultHost` хранит Snapshot и Container и только меняет lifecycle state. Он не строит Listener, Authentication, Dispatcher или зависимости Session и не владеет ими.

Authentication сейчас выполняется после WebSocket handshake. Поэтому при отклонении credentials уже открытое соединение закрывается с `PolicyViolation`, а HTTP-ответ `401` не возвращается.

## 3. Package Responsibilities

| Пакет | Чем владеет | Чем не должен владеть |
|---|---|---|
| `runtimeconfig` | Runtime-only Snapshot models, проверка Published Version и deep-copy преобразование из ConfigurationVersion | Доступ к Repository, HTTP, значения Secret, живые Runtime Services |
| `runtime` | Snapshot Container и оболочка lifecycle Host | Детали HTTP, Repository, поведение конкретных Provider, обработка соединений |
| `listener` | TCP socket, HTTP server, `/ws` Upgrade, server contexts, координация handler и serve goroutine | Правила Authentication, поведение Session, message routing, разрешение Secret |
| `connection` | Передача transport, временный transport context, адаптер Authentication, передача аутентифицированного соединения | Provider-specific verification, внутреннее устройство Session, Repository, routing |
| `authentication` | Transport-neutral contracts, упорядоченный Service, Bootstrap, Registry, Factory adapters, API Key и JWT Providers | Listener, WebSocket, типы HTTP request, Repository, ConfigurationVersion |
| `session` | Одно аутентифицированное WebSocket-соединение, копия Principal, lifecycle, read loop, сериализованные записи, вызов Handler | Метаданные HTTP request, lifecycle Listener, routing policy, persistence |
| `message` | Immutable text/binary Message, минимальные Sender и Handler contracts, EchoHandler | WebSocket, HTTP, реализация Session, Authentication |
| `secretresolver` | Валидация Secret Reference, контракт Resolver, принадлежащие вызывающему коду secret bytes, in-memory реализация | Provider logic, HTTP, ConfigurationVersion, logging |

Фактические импорты подтверждают эти обязанности. Циклов импортов Go нет. Runtime packages не импортируют Control Plane Repository. `message` не импортирует `session` или WebSocket, `session` не импортирует `listener` или `net/http`, а `authentication` не импортирует пакеты Listener или WebSocket.

Сохраняется одна намеренная зависимость адаптера: `runtimeconfig.Builder` импортирует доменную модель ConfigurationVersion. Полученный Snapshot не хранит эту модель или ссылку на Repository.

## 4. Dependency Direction

Фактическое направление зависимостей:

```text
runtimeconfig.Builder -> configurationversion

runtime -> runtimeconfig
listener -> runtimeconfig, connection
connection -> authentication, WebSocket transport
session -> connection, authentication, message, WebSocket transport
authentication -> runtimeconfig, secretresolver
message -> standard library only
secretresolver -> standard library only
```

В сетевом пути dependency injection используется на основных границах: Listener принимает `connection.Dispatcher`; AuthenticationDispatcher принимает Authentication `Service` и следующий authenticated dispatcher; Session принимает `message.Handler`; Bootstrap использует Registry и Resolver; Registry хранит интерфейсы Factory.

Направление зависимостей в целом соответствует ADR-003. Из-за отсутствия production-сборки оно подтверждено на уровне отдельных компонентов и integration tests, но еще не lifecycle исполняемого Runtime.

## 5. Resource Ownership

| Ресурс | Владелец | Получение | Освобождение и ожидание |
|---|---|---|---|
| Данные Published Configuration | Потребители `runtimeconfig.Snapshot` | Builder глубоко копирует Published ConfigurationVersion | Жизненный цикл значения; Container и Host возвращают независимые копии |
| TCP socket | `listener.DefaultListener` | `net.Listen` в `Start` | `http.Server.Shutdown`, `Close` listener, `WaitGroup` serve |
| HTTP server и base context | `listener.DefaultListener` | Создаются в `Start` | Отмена base context, `Shutdown`, ожидание tracked handler |
| Upgraded WebSocket до Authentication | `connection.AuthenticationDispatcher` | Получает владение от WebSocket handler | Закрывает при отказе/ошибке или передает дальше |
| Аутентифицированный WebSocket | `session.DefaultSession` | Получает через Session Dispatcher | Normal close и `CloseNow`; ожидает завершения read loop |
| Session read loop | Goroutine вызывающего HTTP handler | `Session.Run` блокирует вызывающий код | Возврат при закрытии peer, отмене context, Stop, ошибке Handler или чтения |
| Concurrent writes | Вызывающие goroutine, сериализованные Session | `Session.Send` | Каждый вызов возвращается после одной записи или ошибки |
| Разрешенные Secret bytes | Вызов Authentication Provider | Resolver возвращает принадлежащую вызывающему коду копию | API Key и JWT Providers очищают полученные byte slices после использования |

Listener владеет своей serve goroutine и отслеживает активные HTTP handlers. Session Dispatcher не отделяет работу Session: он последовательно выполняет `Start`, блокирующий `Run` и `Stop` в goroutine запроса. Authentication, Registry, Providers, Message handlers и SecretResolver не создают фоновые goroutine.

Передача владения на основных границах явная. Главное слабое место — ограниченность времени shutdown: часть финальных ожиданий не управляется context вызывающей стороны, что зафиксировано в F-04.

## 6. Lifecycle Review

### Runtime Host

Переход `Created -> Running -> Stopped` потокобезопасен, Restart запрещен. Start и Stop только меняют состояние и игнорируют переданные contexts. Ресурсы Runtime пока не подключены к этому lifecycle.

### Listener

Переход `Created -> Running -> Stopping -> Stopped` защищен `sync.RWMutex`. Сетевые операции и ожидание выполняются после освобождения lifecycle mutex. Stop отменяет server base context, вызывает `http.Server.Shutdown`, закрывает TCP Listener и ожидает завершения server и handlers. Конкурентный Stop, увидевший `Stopping`, возвращается сразу, не ожидая завершения первого Stop.

### Session

Переход `Created -> Running -> Stopping -> Stopped` защищен lifecycle mutex. Разрешен только один блокирующий read loop. Stop закрывает WebSocket вне lifecycle lock, а конкурентные вызовы Stop могут ожидать `stopDone`. При этом первый Stop ожидает read loop без выбора по своему context, а Send удерживает lifecycle read lock во время сетевой записи.

### Cancellation and goroutines

Handlers Listener получают отменяемый server context. Authentication и Provider calls распространяют отмену request. Чтение и запись Session принимают contexts. Намеренно отделенных application goroutine нет. Обычные реализованные пути завершаются корректно, однако custom downstream components, игнорирующие cancellation, могут продлить shutdown Listener или Session сверх заданного deadline.

## 7. Security Boundaries

- Configuration и Snapshot models содержат Secret References, а не значения секретов.
- API Key и JWT Providers разрешают secrets при каждой попытке Authentication. Constructors и Factories не разрешают их.
- MemoryResolver копирует значения на входе и выходе. Providers очищают возвращенные byte slices после использования.
- Сравнение API Key использует `crypto/subtle.ConstantTimeCompare`.
- Runtime packages не содержат логирования credentials. Ошибки Authentication dispatcher имеют общее сообщение, сохраняя обернутую причину для `errors.Is`.
- `AuthenticationRequest` временно содержит скопированные Headers и Query из handshake. Session не хранит ни этот request, ни исходный `http.Request`; остаются только копия Principal, RemoteAddress, ID и время создания.
- Текущая граница выполняет Authentication после `101 Switching Protocols`. Это противоречит модели DP-001 с Authentication до соединения и HTTP `401` и увеличивает расход ресурсов на неаутентифицированных клиентов.
- Обработка WebSocket Origin использует стандартное поведение библиотеки. Явной per-Configuration origin policy пока нет.
- Runtime-проверка JWT поддерживает только HMAC algorithms HS256, HS384 и HS512. Asymmetric algorithms остаются metadata DSL, но отклоняются Runtime Provider constructor.
- TLS metadata и timeout metadata Listener присутствуют в Snapshot, но текущая реализация Listener их не применяет.

## 8. Extensibility Review

| Расширение | Существующая точка расширения | Оценка |
|---|---|---|
| Router | `message.Handler`, внедренный в Session Dispatcher | Router может реализовать Handler без изменения Session или Listener. Routing contracts и failure semantics еще предстоит спроектировать. |
| Middleware | Обертки Handler могут композиционно включать другой Handler | Технически возможно без изменения ядра, но ordering, configuration и lifecycle contracts не определены. |
| Basic Provider | Factory и Provider Registry по типу Provider | Конкретные Basic Provider и Factory можно добавить без изменения Registry или Service. |
| Session Manager | Контракт регистрации и удаления отсутствует | Управление, limits или coordinated shutdown потребуют новой границы вокруг создания и завершения Session. |
| Message persistence | Handler или декоратор Handler | Storage можно вызвать за Handler, но delivery guarantees, retries, backpressure и error policy не определены. |
| Plugins | Регистрация Factory и внедрение Handler | Compile-time расширение возможно. Discovery, compatibility, isolation и dynamic plugin loading отсутствуют. |

Архитектура предоставляет полезные узкие границы и не использует type switch в Registry, Service и Bootstrap. Router — наиболее подготовленная следующая точка расширения. Session Manager и dynamic plugins пока нельзя добавить без новых контрактов, что допустимо на alpha-этапе.

## 9. Findings

### F-01 — Authentication выполняется после WebSocket Upgrade

**Severity:** High  
**Observation:** `websocket.Accept` отправляет `101 Switching Protocols` до запуска AuthenticationDispatcher. Отказ выражается WebSocket close frame, а не HTTP `401`.  
**Risk:** Неаутентифицированные клиенты занимают ресурсы upgraded connection, HTTP-клиент не получает документированный контракт ошибки Authentication, а последующее исправление изменит границу Listener-to-Dispatcher.  
**Recommendation:** Определить и реализовать границу Authentication до Upgrade либо официально заменить DP-001 принятым решением и явной threat model для Authentication после Upgrade.  
**Timing:** Необходимо решить до Router.

### F-02 — Runtime Host не является production composition root

**Severity:** High  
**Observation:** Рабочая вертикаль Listener -> Authentication -> Session -> Handler собирается только вызывающим кодом и integration tests. Host хранит Snapshot и Container, но не запускает компоненты.  
**Risk:** Нет production lifecycle, который подтверждает порядок зависимостей, rollback частичного запуска, порядок shutdown и единую цепочку владения из ADR-003.  
**Recommendation:** На основе существующих интерфейсов поручить Host сборку и владение Resolver, Provider Registry/Factories, Authentication Bootstrap, dispatchers, Handler и Listener; добавить rollback частичного запуска.  
**Timing:** Необходимо решить до Router.

### F-03 — Configuration Listener применяется частично

**Severity:** Medium  
**Observation:** Listener валидирует и хранит TLS metadata, но открывает plain TCP. Значения timeout из Snapshot не применяются к `http.Server` или WebSocket operations.  
**Risk:** Валидный Published Snapshot может обещать transport protection и limits, которые Runtime не обеспечивает. Оператор может ошибочно считать Configuration активной.  
**Recommendation:** Реализовать настроенные TLS и timeout либо отклонять неподдерживаемую активную Configuration при build/start Runtime с явной ошибкой.  
**Timing:** Можно выполнить после Router, но обязательно до beta.

### F-04 — Deadline shutdown не является единообразно обязательным

**Severity:** Medium  
**Observation:** Listener ожидает tracked handlers после `Shutdown(ctx)` без deadline-aware wait; конкурентный Listener Stop возвращается во время `Stopping`; первый Session Stop ожидает read loop без учета `ctx`; Session Send удерживает lifecycle read lock во время WebSocket I/O.  
**Risk:** Ошибочный или медленный Handler/write может задержать shutdown сверх deadline вызывающей стороны, а конкурентные lifecycle callers не всегда одинаково наблюдают завершение.  
**Recommendation:** Определить единый lifecycle contract, сделать все ожидания завершения context-aware, добавить общий completion signal для Listener Stop и не удерживать lifecycle locks во время network I/O.  
**Timing:** Необходимо решить до Router, поскольку Router добавит произвольную работу Handler.

### F-05 — Для ошибок Runtime нет operational reporting path

**Severity:** Medium  
**Observation:** Listener отбрасывает ошибки `http.Server.Serve`, а WebSocket handler — ошибки Dispatcher. Runtime components намеренно не логируют sensitive input, но нет безопасного event, metric или injected logger для ошибок.  
**Risk:** Provider outages, ошибки Handler и неожиданное завершение server могут остаться незаметными в production и затруднить диагностику.  
**Recommendation:** Добавить transport-neutral внедряемую границу observability/error reporting с документированной redaction policy; сохранить обобщенные client-facing errors.  
**Timing:** Можно выполнить после Router, но обязательно до beta.

### F-06 — Snapshot adapter находится в пакете Runtime model

**Severity:** Low  
**Observation:** `runtimeconfig.Builder` импортирует `configurationversion`, хотя сами типы Snapshot не зависят от Control Plane.  
**Risk:** Будущие удобные изменения могут втянуть дополнительные понятия Control Plane в пакет, импортируемый всем Runtime.  
**Recommendation:** Сохранить этот код единственным явным adapter либо перенести преобразование в отдельный composition/adapter package при появлении production Loader. Никогда не добавлять доступ к Repository в `runtimeconfig`.  
**Timing:** Backlog; вернуться к вопросу вместе с Loader и до v1.0.

### F-07 — Покрытие Provider намеренно неполное

**Severity:** Accepted limitation  
**Observation:** Реализованы API Key и HMAC JWT; отсутствуют Basic, asymmetric JWT, JWKS, revocation и distributed Secret backends.  
**Risk:** Поддерживаемые deployment scenarios пока ограничены, но текущий код отклоняет неподдерживаемые JWT algorithms, а не принимает их молча.  
**Recommendation:** Явно отражать ограничения в документации и постепенно добавлять Providers через регистрацию Factory, не расширяя Service или Registry.  
**Timing:** Определяется продуктом; Basic и необходимые JWT algorithms нужно выбрать до beta.

### F-08 — Origin policy полагается на стандартное поведение библиотеки

**Severity:** Accepted limitation  
**Observation:** WebSocket Accept использует стандартное поведение Origin библиотеки без Configuration policy.  
**Risk:** Deployment не может выразить более строгие или proxy-aware origin rules через DSL.  
**Recommendation:** Спроектировать явную Configuration-first origin policy до использования Runtime вне контролируемого окружения.  
**Timing:** До beta.

### F-09 — Высокоуровневые Runtime services отсутствуют

**Severity:** Accepted limitation  
**Observation:** Router, Middleware contracts, Session Manager, persistence, backpressure, broadcast, monitoring и dynamic plugins не реализованы.  
**Risk:** Текущая вертикаль подходит для проверки архитектуры и Echo behavior, но не для production workloads.  
**Recommendation:** Добавлять только следующие необходимые contracts, начав с Router после обязательных исправлений, и сохранять transport-neutral границы Message и Handler.  
**Timing:** Milestone backlog; это не дефект alpha foundation.

## 10. Stable Decisions

Следующие решения достаточно подтверждены кодом и тестами, чтобы считать их стабильными для следующей вехи:

- Published ConfigurationVersion — единственный вход, принимаемый Snapshot Builder.
- Потребители Snapshot получают deep copies; изменяемые коллекции Provider и JWT не связаны с ConfigurationVersion или значениями вызывающего кода.
- Runtime отделен от HTTP API Control Service и реализаций Repository.
- Зависимости передаются явно через constructors и узкие интерфейсы.
- Authentication Service упорядочен и не зависит от конкретных Provider; Factory и Registry изолируют создание конкретных Provider.
- Secret values разрешаются близко к месту использования и не хранятся в Configuration, Snapshot, Factory или Provider configuration.
- Контракты Runtime Message и Handler transport-neutral.
- После передачи Session владеет аутентифицированным WebSocket, использует один read loop и сериализует writes.
- Router можно добавить как `message.Handler` без раскрытия WebSocket transport.

## 11. Decisions Not Yet Stable

- Должна ли Authentication выполняться до Upgrade, как требует DP-001, или после него.
- Production composition Host, rollback запуска и порядок shutdown.
- Единые semantics lifecycle cancellation и concurrent Stop.
- Фактическое поведение TLS, timeout и origin policy.
- Контракты Runtime observability, audit, metrics и redaction.
- Таксономия ошибок Authentication сверх success, rejection и возвращенной error.
- Объем Basic и asymmetric JWT Providers.
- Контракты Router, Middleware, Session Manager, persistence, backpressure и plugins.
- Владение Loader и долгосрочное расположение adapter ConfigurationVersion-to-Snapshot.

## 12. Milestone Alpha Assessment

**Verdict: Ready with findings.**

Реализация подтверждает основное направление ADR-003: immutable Snapshot data, явный dependency injection, transport-neutral Authentication и Message contracts, расширяемость Registry/Factory, понятную передачу соединения и тестируемые компоненты без import cycles или связи с Repository.

Проект не готов к production. Разработку Router следует начинать только после устранения F-01, F-02 и F-04. Эти findings затрагивают security boundary, реальный composition root и lifecycle, которые иначе унаследует любая реализация Router. Medium findings F-03 и F-05 необходимо закрыть до beta. Accepted limitations должны оставаться явными и не представляться реализованными возможностями.

## 13. Recommended Next Tasks

### Необходимо решить до Router

1. Перенести Authentication на границу до Upgrade либо зафиксировать принятое заменяющее решение с явной security model и client failure contract.
2. Сделать Runtime Host production composition root для существующих Snapshot, Authentication, Dispatcher, Session Handler и Listener, включая rollback запуска и упорядоченный shutdown.
3. Усилить lifecycle contracts: context-aware waits, общий completion signal для Stop, отсутствие lifecycle locks во время network I/O и тесты с блокирующими/ошибающимися downstream handlers.

### Можно выполнить после Router, но до beta

4. Применить настройки TLS и timeout из Listener Snapshot либо завершать startup ошибкой при неподдерживаемой активной Configuration.
5. Добавить внедряемые и безопасные для Secret Runtime error reporting, metrics и lifecycle events.
6. Добавить Configuration-first WebSocket Origin policy и integration tests для разрешенных, запрещенных, отсутствующих и proxy-mediated значений Origin.

### Backlog

7. Добавлять контракты Session Manager, persistence и plugins только после появления конкретных use cases, определяющих ownership, delivery guarantees и compatibility requirements.

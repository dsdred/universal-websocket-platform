# ARCH-001: Runtime Architectural Pattern

[English version](../../en/architecture/ARCH-001-runtime-architectural-pattern.md)

**Статус:** Active

**Область:** Runtime architecture

**Стабильность:**

- Документ описывает архитектурный паттерн, подтвержденный реализованной Alpha-вертикалью.
- Он не замораживает будущие контракты Router, Middleware, Plugin или Policy.
- Его можно пересматривать только при наличии явного архитектурного обоснования, подтвержденного опытом реализации или проектированием подсистемы.

## 1. Purpose

ARCH-001 фиксирует общий способ рассуждать о компонентах Runtime. Он предоставляет вопросы и границы проектирования, которые можно повторно использовать при проектировании Handshake, Router, Delivery, Persistence и Plugins, не предписывая их будущие Go API.

Проект использует несколько видов документов с разным назначением:

- **ARCH** описывает общий архитектурный паттерн и терминологию, подтвержденные несколькими реализованными границами.
- **ADR** фиксирует решение одной значимой архитектурной проблемы, ее контекст и компромиссы.
- **DP** проектирует конкретную подсистему до стабилизации ее контрактов или поведения.
- **Review** оценивает фактическую реализацию относительно принятых решений и предложенных designs.

ARCH-001 не заменяет ADR или DP. Значимый выбор по-прежнему требует ADR, а новая подсистема — сфокусированного проектирования, если ее контракты еще не установлены.

## 2. Evidence from Alpha

Паттерн извлечен из реализованной Alpha-вертикали, а не из гипотетического framework:

```text
Published ConfigurationVersion
    -> runtimeconfig.Snapshot
    -> Runtime components
    -> Listener
    -> Connection
    -> Authentication
    -> Session
    -> Runtime Message
    -> Handler
    -> Send
```

Текущая реализация и [Runtime Alpha Architecture Review](../reviews/runtime-alpha-review.md) подтверждают следующие принципы:

- Snapshot отделяет данные Control Plane от effective Runtime data и копирует вложенную Configuration Provider.
- Factory преобразует `AuthenticationProviderSnapshot` в runtime-specific Configuration Provider до создания Provider.
- Listener завершает свою ответственность по приему transport, передавая upgraded connection через `Dispatch`.
- Authentication Dispatcher владеет upgraded connection до отказа, ошибки или явной передачи следующему компоненту.
- Session владеет WebSocket connection после успешной Authentication.
- Runtime Message содержит скопированный text или binary payload и ничего не знает о WebSocket.
- У сетевых ресурсов и application goroutine есть определяемые владельцы и пути завершения.
- Lifecycles Host, Listener и Session явны, хотя их контракты не идентичны.

Эти факты не означают, что Runtime Host уже является production composition root. Сейчас он владеет Snapshot и Container и только меняет lifecycle state. Production composition остается finding из Alpha review.

## 3. Core Architectural Pattern

Концептуальный паттерн:

```text
Context
    -> Evaluation
    -> Decision
    -> Execution
```

- **Context** предоставляет минимальную информацию, необходимую подсистеме.
- **Evaluation** определяет допустимое или требуемое действие.
- **Decision** выражает результат, не выполняя несвязанный эффект.
- **Execution** применяет результат через компонент, владеющий эффектом и его ресурсами.

Evaluation — намеренно нейтральный термин. В зависимости от подсистемы Evaluation может состоять из policies, rules, matchers, Providers или Handlers. Каждая подсистема выбирает минимальную модель, соответствующую ее ответственности.

ARCH-001 не требует универсального Policy Engine. Он также не требует, чтобы каждая подсистема искусственно создавала четыре отдельных Go-типа. Паттерн является инструментом поиска границ, а не обязательным framework или generic processing pipeline.

## 4. Context

Context содержит только данные, необходимые для оценки одной операции.

Хорошо определенный Context:

- содержит минимальный набор данных;
- исключает несвязанные transport references;
- не выполняет business logic;
- передает mutable resources только при явных semantics владения и передачи;
- не использует `context.Value` как скрытый dependency injection container.

Возможными Context являются handshake metadata, аутентифицированный Principal, Runtime Message или будущий delivery target set. Это концептуальные примеры, а не объявления обязательных Go-типов.

Текущий Authentication adapter демонстрирует нормализацию: он копирует выбранные handshake Headers, Query values, RemoteAddress и имя transport в `AuthenticationRequest`; Providers не получают `http.Request` или WebSocket connection. Напротив, `ConnectionContext` намеренно передает mutable transport references, поскольку представляет явную передачу владения, а не immutable business data.

## 5. Evaluation

Evaluation отвечает на вопрос, какое действие разрешено или требуется при effective Configuration и текущем Context.

В зависимости от подсистемы Evaluation может использовать:

- Authentication Providers;
- будущие route matchers;
- будущие origin policies;
- будущие delivery rules;
- будущие persistence conditions.

Evaluation не должна исполнять ответственность другого компонента. Authentication проверяет credentials, но не выполняет WebSocket Upgrade. Будущий Router может выбрать destination или Handler, но не должен владеть Session только потому, что выбрал эту Session. Policy не должна открывать sockets, запускать goroutines или получать несвязанные mutable resources.

Evaluation может быть последовательной, композиционной или специализированной. Ее форма должна следовать реальным use cases подсистемы, а не общепроектному generic engine.

## 6. Decision

Decision является результатом Evaluation. Примеры:

- allow или reject;
- dispatch в Handler или destination;
- drop;
- persist;
- select recipients.

Decision не обязана быть единым глобальным типом. Каждая подсистема может определить небольшую модель результата, выражающую только ее outcomes. Универсальный Decision object нельзя вводить только ради внешнего единообразия несвязанных подсистем.

Отрицательное решение и execution error — разные понятия. Rejected credentials, отсутствие подходящего route или намеренный persistence skip могут быть корректными outcomes. Недоступность Resolver, ошибка Storage или неожиданный transport failure являются operational errors и требуют отдельной семантики.

## 7. Execution

Executor применяет Decision и владеет соответствующим эффектом. Существующие и будущие примеры:

- Listener выполняет HTTP rejection или WebSocket Upgrade;
- Session читает и отправляет messages;
- Handler обрабатывает Runtime Message;
- будущий компонент Storage сохраняет Message.

Execution требует:

- явного владельца;
- понятного lifecycle при наличии ресурсов;
- определенной error semantics;
- пути корректного завершения или shutdown.

Компонент, выполняющий Evaluation, может также исполнить собственный узкий эффект, если это остается его единственной ответственностью. ARCH-001 не требует искусственного интерфейса Executor между каждым вызовом функций.

## 8. Configuration First

Определяющая формулировка:

> Configuration определяет поведение Runtime.
> Published ConfigurationVersion является источником истины.
> Runtime исполняет immutable Snapshot.

Runtime принимает операционные решения в рамках правил, полученных из Published Configuration. Прием соединения, результаты Provider, выбор routing и поведение shutdown являются решениями Runtime, ограниченными effective Configuration и явными операционными зависимостями.

Действуют следующие инварианты:

- Runtime не читает Control Plane repositories.
- Runtime не изменяет Published Configuration.
- Snapshot является независимо скопированным стабильным представлением Runtime; mutable transport resources описываются через ownership, а не называются immutable.
- Unsupported Published settings нельзя молча игнорировать.
- Настройка либо поддерживается, либо Snapshot construction/Bootstrap/startup явно ее отклоняет.

Последний инвариант непосредственно учитывает Alpha finding о том, что Listener сейчас хранит TLS и timeout metadata, не применяя их полностью.

## 9. Dependency Direction

Dependency direction является набором проверяемых границ, а не одной обязательной линейной иерархией packages.

Обязательные инварианты:

- Runtime packages не зависят от Control Plane repositories.
- Listener не зависит от concrete Authentication Providers.
- Authentication не зависит от `net/http` или WebSocket.
- Session не зависит от Listener или HTTP.
- Message не зависит от Session или WebSocket.
- Extensions не получают произвольный доступ к внутреннему Runtime state.

Концептуальные layer diagrams помогают объяснять ответственность, но допустимость зависимости определяется только обоснованным контрактом и ответственностью. Например, `runtimeconfig.Builder` является явным adapter к ConfigurationVersion, при этом полученный Snapshot не зависит от Repository или HTTP. Session импортирует WebSocket library, потому что владеет аутентифицированным соединением; Message ее не импортирует.

## 10. Ownership

У каждого mutable resource есть один текущий владелец, а передача владения должна быть явной.

Alpha-вертикаль предоставляет конкретные примеры:

- Listener владеет TCP listener и HTTP server.
- Connection Dispatcher получает и передает upgraded connection.
- Authentication Dispatcher владеет connection до успешной передачи следующему компоненту или закрытия при ошибке.
- Session владеет WebSocket connection после Authentication.
- Session владеет своим read loop и сериализацией writes.
- Secret Resolver владеет сохраненными копиями secret bytes; успешный Resolve возвращает отдельную копию, принадлежащую вызывающему коду.

Компонент может принять владение ресурсом, созданным другим компонентом, если передача определена контрактом. Владение не ограничивается только ресурсами, созданными самим компонентом.

Контракт владения должен отвечать, кто закрывает ресурс при успехе, отказе, cancellation, ошибке construction и частичной передаче. Два компонента не должны одновременно предполагать, что один и тот же ресурс закроет другой.

## 11. Lifecycle and Concurrency

Runtime components следуют подтвержденным требованиям в случаях, когда их ответственность включает lifecycle или concurrency:

- lifecycle explicit;
- Stop идемпотентен;
- restart не предполагается автоматически;
- network I/O и длительные waits не выполняются под lifecycle mutex;
- у каждой goroutine есть владелец и путь завершения;
- context cancellation является частью lifecycle;
- concurrent Start, Stop, Run и Send имеют определенную семантику, если такие операции существуют;
- ошибки shutdown не теряются молча.

Не каждому компоненту нужна одинаковая полная state machine. Stateless Matcher может не иметь Start или Stop, тогда как Listener и Session требуют более подробных lifecycle contracts. Необходимые states определяются контрактом компонента.

Alpha review выявил неполную обработку deadline, semantics concurrent Stop, удержание lifecycle lock во время Session write и потерю ошибок server/Dispatcher. Это findings, которые необходимо исправить, а не поведение, одобренное данным паттерном.

## 12. Boring Core

Ядро Runtime остается компактным и предсказуемым.

- Новая возможность сначала рассматривается как Handler, Provider, Matcher, Policy, Middleware или Plugin в соответствии с ее фактической ответственностью.
- Core меняется только тогда, когда существующие контракты принципиально недостаточны.
- Универсальные абстракции не создаются до того, как несколько реальных use cases подтвердят одинаковую потребность.
- Повторяющийся концептуальный паттерн не обязан становиться единым generic framework.
- Одна интеграция не является основанием для широкого доступа к внутреннему состоянию Runtime.

«Boring» означает явное владение, узкие контракты и обычную композицию. Это не означает сокрытие сложности или игнорирование необходимых lifecycle и error behavior.

## 13. Applying the Pattern

Следующие примеры являются осторожным применением паттерна. Это входные данные для будущего проектирования, а не реализованные контракты.

### Handshake

```text
Handshake metadata
    -> Authentication and Origin Evaluation
    -> allow or reject
    -> HTTP rejection or WebSocket Upgrade
```

Handshake pipeline требует отдельного будущего DP. В таком виде он еще не реализован: текущая Authentication происходит после Upgrade, что зафиксировано Alpha review.

### Routing

```text
Message Context
    -> route matchers
    -> selected Handler or destination
    -> Handler execution
```

Пример показывает только распределение ответственности. Он не фиксирует Router API, route model, matcher interface или error policy.

### Persistence

```text
Message Context
    -> persistence rule
    -> store or skip
    -> storage execution
```

Пример не определяет Storage contract, transaction model, delivery guarantee, retry rule или database.

### Delivery

```text
Message and target Context
    -> addressing and delivery rules
    -> recipient set
    -> Session delivery
```

Пример не проектирует Groups, Topics, fan-out, backpressure или поведение Session Manager.

## 14. Architectural Questions

Для каждой новой подсистемы Runtime необходимо ответить:

1. Какова ее единственная ответственность?
2. Какие данные образуют минимальный Context?
3. Что выполняет Evaluation?
4. Как выражается Decision?
5. Кто исполняет эффект?
6. Кто владеет ресурсами?
7. Как компонент запускается и останавливается?
8. Можно ли реализовать возможность через существующую точку расширения?
9. Не проникает ли Control Plane state в Runtime?
10. Не создается ли универсальная абстракция до появления реальных use cases?

Ответы должны быть достаточно конкретными для тестирования и ревью. Формулировка «это делает framework» не определяет ownership, lifecycle или error semantics.

## 15. Anti-Patterns

Следующие направления запрещены или настоятельно не рекомендуются:

- universal Policy Engine для всех подсистем;
- единый глобальный Decision type;
- god-object Runtime Host;
- Session, знающая `http.Request`;
- Message, знающий `websocket.Conn`;
- Provider, знающий Snapshot;
- Listener, содержащий JWT- или API Key-specific logic;
- hidden dependencies через `context.Value`;
- detached goroutines без владельца и пути завершения;
- silent ignoring of unsupported Configuration;
- premature generic Module framework;
- изменение Core ради одной конкретной интеграции.

Runtime Host как composition root отвечает за wiring и координацию lifecycle, но должен делегировать поведение подсистем, а не поглощать его.

## 16. Relationship to Future Documents

- Handshake Pipeline будет спроектирован в отдельном DP.
- Router получит собственный DP.
- Plugin ABI требует отдельного DP и, вероятно, ADR, поскольку compatibility и isolation создают долгосрочные ограничения.
- ARCH-001 не определяет ни один из этих API.
- Будущий DP должен объяснять, как его подсистема следует ARCH-001, либо почему необходимо обоснованное исключение.
- Если несколько обоснованных DP не укладываются в ARCH-001, архитектурное руководство необходимо явно пересмотреть, а не обходить скрытыми исключениями.

Будущие документы должны ссылаться на конкретные evidence и сохранять различие между proposed contracts и реализованным поведением.

## 17. Final Statement

> Configuration определяет поведение Runtime.
> Runtime оценивает контекст, принимает ограниченные правилами решения и исполняет их через явно определённых владельцев ресурсов.

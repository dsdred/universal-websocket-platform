# Universal WebSocket Platform Master Engineering Plan

[English version](../../en/roadmap/MASTER_PLAN.md)

**Статус:** Living engineering plan

Документ описывает предполагаемое инженерное развитие Universal WebSocket Platform. Он определяет стадии зрелости, архитектурные эпики и критерии релиза без назначения дат, целевых показателей производительности или API будущих подсистем. Это не маркетинговый roadmap, не расписание релизов и не backlog задач.

## 1. Project Vision

Universal WebSocket Platform — open-source платформа для создания, настройки, развертывания и эксплуатации независимых WebSocket-серверов без повторного написания инфраструктурного кода.

Пользователи описывают поведение серверов через понятную Configuration и управляют ими через стабильные API. Runtime предсказуемо исполняет Configuration с явными границами Provider, изолированным владением ресурсами и поведением, которое оператор может понять до развертывания.

Полный замысел продукта определен в [Vision проекта](../../../spec/00-product/vision.md). Этот план переводит замысел в стадии инженерной зрелости и не заменяет Vision.

## 2. Current State

Сейчас репозиторий содержит Alpha foundation, а не готовую к production платформу. Реализованное состояние зафиксировано в [`spec/current-state.md`](../../../spec/current-state.md) и оценено в [Runtime Alpha Architecture Review](../reviews/runtime-alpha-review.md).

### Configuration

- Workspace, Configuration и ConfigurationVersion имеют in-memory API Control Service.
- ConfigurationVersion поддерживает lifecycle create, publish и archive.
- Metadata Listener, TLS, timeout и Authentication представлены в Configuration DSL.
- Metadata Authentication включает настройки API Key, JWT и Basic; validation выделена в отдельный компонент.

### Snapshot

- `runtimeconfig.Builder` принимает только Published ConfigurationVersion.
- Runtime Snapshot содержит данные Listener и Authentication.
- Вложенные коллекции Provider и JWT копируются, поэтому последующее изменение Configuration не меняет существующий Snapshot.

### Listener and Connection

- Listener открывает TCP socket и запускает HTTP server.
- `GET /ws` выполняет RFC 6455 Upgrade и передает соединение через Dispatcher.
- Metadata TLS и timeout Listener еще не полностью применяется Runtime.

### Authentication

- Существуют transport-neutral границы request, result, Principal, Provider, Factory, Registry, Service и Bootstrap.
- Реализованы production API Key и HMAC JWT Providers и Factories.
- Значения Secret разрешаются по Secret References для каждой попытки Authentication.
- Authentication сейчас выполняется после WebSocket Upgrade; Basic и asymmetric JWT verification отсутствуют.

### Session, Message, and Echo

- Session владеет аутентифицированным WebSocket connection, одним read loop и сериализованными writes.
- Runtime Message не зависит от transport и копирует text или binary payload.
- Message Handler не зависит от WebSocket transport.
- EchoHandler демонстрирует поток от входящего Message через Handler к `Session.Send`.

### Architecture

- [ADR-0002](../adr/0002-configuration-dsl.md) определяет ConfigurationVersion как Configuration DSL и Published source of truth.
- [ADR-0003](../adr/0003-runtime-architecture.md) определяет компонентную модель Runtime и явный dependency injection.
- [ARCH-001](../architecture/ARCH-001-runtime-architectural-pattern.md) фиксирует подтвержденный паттерн `Context -> Evaluation -> Decision -> Execution`, ownership, lifecycle и Boring Core.
- Вердикт Runtime Alpha Review — **Ready with findings**. Runtime Host еще не является production composition root; lifecycle, effective Listener settings и operational diagnostics требуют дальнейшей работы.

## 3. Engineering Principles

Основным руководством по проектированию является [ARCH-001: Runtime Architectural Pattern](../architecture/ARCH-001-runtime-architectural-pattern.md).

### Configuration First

Configuration определяет поведение Runtime. Published ConfigurationVersion является источником истины, а Runtime исполняет независимо скопированный Snapshot. Published setting либо поддерживается, либо явно отклоняется; молча игнорировать его нельзя.

### Ownership

У каждого mutable resource есть один текущий владелец. Передача владения явна для success, rejection, cancellation и partial failure. Компонент может принять ресурс, созданный в другом месте, но контракт должен определять, кто его закрывает.

### Lifecycle

Компоненты с ресурсами имеют явные semantics lifecycle и concurrency. У каждой goroutine есть владелец и путь завершения. Context cancellation участвует в shutdown, длительные операции не выполняются под lifecycle locks, а ошибки shutdown остаются наблюдаемыми.

### Boring Core

Core остается компактным. Новое поведение сначала следует пытаться реализовать через подходящую существующую границу, например Handler или Provider. Новая граница Matcher, Policy, Middleware или Plugin вводится только при наличии требований конкретной подсистемы.

### No Magic

Зависимости передаются явно. Runtime не обнаруживает Control Plane repositories, не использует globals или `context.Value` как service locators, не создает скрытую Configuration и не запускает detached work без владельца.

## 4. Milestones

Milestones описывают инженерную зрелость. Они не задают календарные даты и не гарантируют время публичного релиза.

### Alpha — Prove the Foundation

**Цель:** Доказать, что Configuration можно преобразовать в изолированный Runtime Snapshot и использовать в минимальной WebSocket-вертикали с явными границами компонентов.

**Критерии завершения:**

- Существуют lifecycle ConfigurationVersion и основные metadata Listener/Authentication.
- Published Configuration преобразуется в скопированный Runtime Snapshot.
- Listener, Connection Dispatcher, Authentication, Session, Message, Handler и Send образуют протестированную вертикаль.
- Значения Secret остаются вне Configuration и Snapshot.
- Архитектурное ревью фиксирует стабильные решения, ограничения и debt.
- ARCH-001 содержит только паттерны, подтвержденные реализацией.

**Основные эпики:** Configuration DSL, Snapshot, Listener, контракты и Providers Authentication, Session, Runtime Message, Echo vertical, архитектурное ревью и архитектурное руководство.

### Beta — Complete the Single-Node Runtime

**Цель:** Собрать Alpha-компоненты в единый, конфигурируемый, наблюдаемый и безопасно эксплуатируемый single-node Runtime.

**Критерии завершения:**

- Runtime Host является production composition root и владеет упорядоченными startup, rollback и shutdown.
- Evaluation Authentication и Origin выполняется на корректной границе Handshake до Upgrade.
- Router детерминированно выбирает настроенное поведение без утечки transport.
- Владение Session и coordinated management определены и протестированы.
- Delivery и минимально необходимое поведение Persistence имеют явную семантику.
- TLS, timeout и все принятые Published settings применяются либо отклоняются при startup.
- Metrics и operational diagnostics показывают ошибки, не раскрывая credentials или Secrets.
- High findings из Alpha review закрыты, и неподдерживаемая возможность не представляется активной.

**Основные эпики:** Handshake, Runtime Host, Configuration validation, Router, Session Manager, Delivery, Persistence, TLS, Metrics, operational diagnostics и начальные Plugin contracts.

### RC — Stabilize the Product Contract

**Цель:** Стабилизировать поддерживаемые Configuration, API, Runtime lifecycle и operational contract перед 1.0.

**Критерии завершения:**

- Не остается нерешенных Blocker или High architecture findings.
- Правила compatibility поддерживаемых Configuration и API документированы и протестированы.
- Пути upgrade, migration, restart, failure recovery и shutdown проверены end to end.
- Security и operational reviews охватывают границы Handshake, Secrets, TLS, diagnostics и persistence.
- Unsupported Configuration отклоняется с понятной диагностикой без sensitive data.
- Документация описывает фактическое поведение, ограничения и процедуры recovery.

**Основные эпики:** contract hardening, compatibility, failure recovery, security review, operational review, migration verification и release qualification.

### 1.0 — Stable Single-Node Platform

**Цель:** Предоставить первую стабильную и поддерживаемую реализацию основной модели UWP для независимых WebSocket-серверов на одном Runtime node.

**Критерии завершения:**

- Выполнены все обязательные критерии 1.0 из раздела 6.
- Published Configuration порождает воспроизводимое поведение Runtime.
- Полная цепочка ownership и lifecycle тестируема и наблюдаема.
- Для публичных контрактов документирована backward-compatibility policy.
- Платформа явно завершается ошибкой, если Configuration или operational dependencies невозможно исполнить.

**Основные эпики:** закрытие findings RC, стабильная документация, compatibility guarantees и поддерживаемое operational behavior. Новые core features не являются целью этой стадии.

### 2.0+ — Evidence-Driven Expansion

**Цель:** Рассматривать возможности, меняющие deployment topology, compatibility boundaries или isolation расширений, только после появления evidence от стабильного single-node использования.

**Критерии завершения:**

- У каждой возможности есть конкретные use cases и сфокусированный DP.
- Значимые решения compatibility или topology имеют ADR.
- Для существующего поведения 1.x существует явная migration strategy.
- Distributed ownership, consistency и failure semantics определены до реализации.

**Основные эпики:** возможные clustering, federation, horizontal coordination, более сильная isolation plugins и более широкие integrations. Этот план не обещает их включение.

## 5. Beta Roadmap

Beta организована как набор архитектурных эпиков. Для каждого эпика требуется сфокусированное проектирование, если его контракты еще не стабильны.

### Foundation Gates

1. **Handshake:** перенести Authentication и будущую Origin evaluation до Upgrade с явными allow/reject и transport error semantics. Pipeline должен быть определен отдельным DP.
2. **Runtime Host:** скомпоновать существующие компоненты без превращения Host в god object; владеть порядком startup, rollback частичного запуска и порядком shutdown.
3. **Lifecycle hardening:** согласовать cancellation и semantics concurrent Stop, убрать lifecycle locks из network I/O и сохранять ошибки shutdown.
4. **Configuration validation:** гарантировать, что каждый Published setting либо исполняется, либо отклоняется до приема traffic.

Эти gates предшествуют Router, чтобы Router не унаследовал нестабильную границу security, composition или shutdown.

### Message Processing

5. **Router:** выбирать Handler или destination из Context Runtime Message без доступа к WebSocket transport. Routing semantics определяются отдельным DP, а не этим планом.
6. **Session Manager:** координировать регистрацию, удаление, limits и shutdown Session, не забирая владение transport у Session.
7. **Delivery:** определить semantics recipient selection, ordering, failure и backpressure до добавления Groups, Topics или broadcast.
8. **Persistence:** определить, что сохраняется, когда Storage optional или required и как его ошибки влияют на обработку Message. Технология и API Storage не выбраны.

### Operational Completion

9. **TLS and Listener settings:** безопасно разрешать certificate references и применять TLS, handshake, read, write и idle limits.
10. **Metrics:** предоставлять ограниченные и не содержащие sensitive data measurements компонентов и lifecycle с контролируемой cardinality labels.
11. **Operational diagnostics:** сообщать об ошибках server, Dispatcher, Provider, Handler и shutdown, сохраняя redaction Secrets и credentials.
12. **Plugin contracts:** определить минимальные extension contracts, подтвержденные несколькими реальными use cases. Dynamic loading и ABI не подразумеваются.

Последовательность отражает архитектурные зависимости, а не даты или очередь задач. Эпики могут быть разделены будущими DP и reviews.

## 6. Release Criteria

### Mandatory for 1.0

- Production composition root строит полную поддерживаемую Runtime-вертикаль и владеет ею.
- Published Configuration является единственным источником поведения; все поддерживаемые поля действуют, а неподдерживаемые явно приводят к ошибке.
- Handshake выполняет обязательную Authentication и admission соединения до WebSocket Upgrade.
- Поведение Router детерминировано, объяснимо и transport-neutral на границе Message.
- Lifecycle Session, coordinated shutdown, writes и ownership ресурсов проверены race tests.
- Необходимые semantics delivery и persistence выдерживают определенные сценарии restart и failure.
- TLS, Origin policy, timeout, Secret resolution и redaction credentials имеют end-to-end security tests.
- Metrics и diagnostics показывают health и failures компонентов без утечки sensitive data.
- Контракты API Control и Runtime имеют документированное поведение compatibility и validation.
- Процедуры recovery, migration, shutdown и эксплуатации документированы и протестированы.

### Eligible for 1.x

- Дополнительные Authentication Providers и algorithms сверх поддерживаемого набора 1.0.
- Дополнительные adapters Secret Storage.
- Advanced routing и delivery policies, основанные на подтвержденных use cases.
- Улучшения Import/Export и operational automation.
- Возможности Admin UI, использующие стабильные API.
- Поддерживаемая поверхность plugin development после стабилизации extension contracts.

### Reserved for 2.0 or Later Evaluation

- Cluster coordination и distributed Runtime ownership.
- Federation между независимыми domains платформы.
- Semantics horizontal scaling, требующие distributed Session или delivery state.
- Нарушающий compatibility Plugin ABI или process isolation.
- Широкие семейства enterprise и cloud integrations.

Расположение после 1.0 не является обязательством реализовать возможность. Для каждого пункта по-прежнему необходимы evidence, design и prioritization.

## 7. Deferred Features

Следующие возможности намеренно отложены, поскольку вводят distributed ownership, external compatibility или operational complexity сверх single-node core:

- Cluster operation;
- Federation;
- horizontal scaling;
- distributed Session и delivery coordination;
- enterprise plugin families;
- широкие cloud integrations;
- dynamic Plugin ABI и isolation;
- cross-node Message ordering и recovery.

Deferred features не должны формировать текущий Core через speculative abstractions. К ним можно вернуться, когда стабильное использование сформирует конкретные требования.

## 8. Technical Debt

Runtime Alpha Review фиксирует implementation debt, который необходимо учитывать отдельно от новой функциональности:

- Listener хранит TLS и timeout metadata, не применяя их полностью.
- Некоторые waits shutdown не ограничиваются context вызывающей стороны.
- Semantics concurrent Stop различаются между компонентами.
- Session сейчас удерживает lifecycle read lock во время WebSocket write.
- У ошибок HTTP server и Dispatcher нет operational reporting path.
- `runtimeconfig.Builder` является явным adapter ConfigurationVersion внутри пакета Runtime model и не должен накапливать concerns Repository.
- Покрытие Runtime для Basic и asymmetric JWT неполное.
- Поведение Origin полагается на defaults библиотеки, а не на явную Configuration.

Technical debt закрывается тестами и изменениями реализации, а не переименованием текущих ограничений в поддерживаемое поведение.

## 9. Architectural Debt

Architectural debt относится к границам, которые еще не разрешены или не представлены production composition:

- **Authentication before Upgrade:** текущая Authentication после Upgrade противоречит ожидаемой security boundary Handshake.
- **Runtime Host:** Host еще не собирает реализованную Runtime-вертикаль и не владеет ею.
- **Effective Listener Configuration:** metadata TLS и timeout может попасть в Snapshot без полного исполнения или явного отклонения.
- **Shutdown semantics:** cancellation, завершение concurrent Stop, error propagation и поведение длительного Handler требуют единого контракта.
- **Operational diagnostics:** ownership ошибок и redaction должны пересекать границы компонентов без привязки компонентов к одной реализации logging.
- **Extension boundaries:** Router готов использовать Handler как seam, а Session Manager, Persistence, Delivery и Plugin contracts еще требуют сфокусированного проектирования.

Architectural debt устраняется через DP, при значимых последствиях ADR, реализацию и последующее review. MASTER_PLAN не определяет эти контракты самостоятельно.

## 10. Things We Will Not Do

UWP не будет:

- превращаться в enterprise service bus или общую integration platform;
- вводить второй программируемый язык или executable general-purpose DSL вокруг декларативной модели ConfigurationVersion;
- превращать Runtime в cluster orchestrator;
- создавать единый universal Policy Engine для несвязанных подсистем;
- создавать единый global Decision type ради внешнего единообразия;
- вводить generic framework до появления нескольких реальных use cases для одной абстракции;
- позволять Runtime читать Control Plane repositories или изменять Published Configuration;
- скрывать зависимости в globals, service locators или `context.Value`;
- предоставлять extensions неограниченный доступ к внутреннему Runtime state;
- молча игнорировать unsupported Configuration;
- изменять Core только ради одной конкретной integration;
- добавлять distributed complexity в single-node core без evidence.

Эти границы защищают Configuration-first модель продукта и принцип Boring Core. Они не запрещают сфокусированные компоненты или extensions при наличии реальных требований.

## 11. Success Definition

UWP 1.0 считается успешным, если обладает следующими инженерными свойствами:

- **Predictable:** одинаковые Published Configuration и явные зависимости приводят к объяснимому поведению Runtime.
- **Configuration-driven:** operational behavior прослеживается до validated Configuration, а не скрытых defaults.
- **Isolated:** независимо настроенные серверы имеют явные границы ресурсов, lifecycle и failure.
- **Safe:** credentials и Secrets остаются вне публичной Configuration, а admission соединения выполняется на определенной security boundary.
- **Owned:** у каждого mutable resource и goroutine есть владелец и путь завершения.
- **Observable:** оператор может различать rejection, configuration failure, dependency failure и internal failure без раскрытия sensitive data.
- **Extensible:** поддерживаемые Providers и Handlers добавляются через узкие контракты без изменения несвязанных компонентов Core.
- **Recoverable:** restart, shutdown, partial startup и поддерживаемые failure scenarios имеют детерминированное поведение.
- **Compatible:** развитие публичных Configuration и API следует документированным правилам compatibility и migration.
- **Boring:** composition явна, зависимости видимы, а abstraction следует подтвержденным use cases.

Success не определяется количеством функций, громким benchmark, deployment topology или каталогом integrations.

## 12. Living Document

MASTER_PLAN обновляется при существенном изменении реализованного состояния, архитектурных reviews, milestone exit criteria или инженерных приоритетов. Изменения должны различать завершенную работу, текущий debt и будущие намерения.

Документ не заменяет сфокусированные DP, ADR, `spec/current-state.md` или планирование задач. API будущих подсистем проектируются в отдельных DP. Значимые решения фиксируются в ADR. Фактические возможности остаются описанными в current state и проверяются reviews.

Разные документы меняются с разной частотой:

- MASTER_PLAN регулярно пересматривается по мере изменения engineering evidence.
- ARCH меняется редко и только при архитектурном обосновании.
- Vision меняется только при изменении фундаментального назначения продукта.

Само по себе изменение MASTER_PLAN не делает proposed capability реализованной или обязательной для релиза.

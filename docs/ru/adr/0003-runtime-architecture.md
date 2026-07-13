# ADR 0003: Runtime Architecture

## Статус

Принято (Accepted).

## Контекст

ADR-0002 закрепил ConfigurationVersion как декларативный Configuration DSL. Следующая архитектурная граница — модель исполнения этого DSL.

Runtime исполняет Published ConfigurationVersion. Он не является источником Configuration и не должен создавать вторую Configuration-модель, скрытые defaults или зависимости от management API и его реализации persistence.

Архитектура также должна позволять Authentication и будущим возможностям развиваться как заменяемым компонентам без привязки к transport, Repository или конкретной реализации Provider.

## Решение

Runtime компонуется из независимых компонентов с единственной ответственностью. Компоненты явно получают зависимости через dependency injection и взаимодействуют через стабильные контракты.

Ядро Runtime не зависит от HTTP API или реализаций Repository. Integration adapters и composition root передают Published ConfigurationVersion и конкретные зависимости во время запуска.

### Runtime Pipeline

Runtime использует следующий концептуальный pipeline:

```text
Configuration Loader
        |
        v
Configuration Snapshot
        |
        v
Secret Resolver
        |
        v
Authentication Service
        |
        v
Authentication Provider Registry
        |
        v
Authentication Providers
        |
        v
Principal
        |
        v
Authorization
        |
        v
Routing
        |
        v
Storage
        |
        v
Monitoring
```

Этот pipeline определяет границы компонентов и порядок композиции. Он не требует, чтобы каждый будущий запрос или операция проходили через каждый компонент, и не определяет поведение transport.

### Принципы Runtime

- **Single Responsibility:** каждый компонент имеет одно явное назначение.
- **Composition over inheritance:** поведение Runtime собирается из взаимодействующих компонентов, а не из иерархий типов.
- **Dependency Injection:** зависимости передаются явно во время композиции.
- **Stateless services where possible:** эксплуатационное состояние вводится только там, где этого требует ответственность компонента.
- **Immutable Configuration Snapshot:** Runtime Services наблюдают одно стабильное представление Configuration.
- **No hidden configuration:** каждая настройка, влияющая на поведение, происходит из Published ConfigurationVersion или явной эксплуатационной зависимости.

### Configuration Snapshot

Runtime никогда не читает Draft ConfigurationVersion. Он запускается только из Published ConfigurationVersion, переданной через границу Configuration Loader.

После загрузки Runtime создает immutable Configuration Snapshot. Все Runtime Services используют только этот Snapshot и не могут его изменять. Семантика reload, replacement и lifecycle процесса требует отдельных решений; она не должна изменять существующий Snapshot на месте.

Configuration Loader является абстракцией на границе Runtime. Он не предоставляет Runtime Services доступ к Repository или HTTP API.

### Secret Resolver

Configuration и Configuration Snapshot содержат только Secret References и никогда не содержат реальные значения Secret. Secret Resolver разрешает необходимые ссылки во время запуска Runtime.

Разрешенные Secrets существуют только в памяти процесса и никогда не записываются обратно в Configuration или Snapshot. Их нельзя записывать в журнал, сохранять, экспортировать или включать в детали ошибок. Их lifecycle должен быть ограничен компонентами, которым они необходимы.

Синтаксис ссылок, разделение со Storage и правила безопасности определены в [DP-002: Secret References](../proposals/DP-002-secret-references.md).

### Provider Registry

Runtime не знает конкретных реализаций Provider. Authentication Provider Registry сопоставляет effective Configuration с созданием Provider и передает скомпонованные Providers в Authentication Service.

Registry отвечает за выбор и создание Provider. Authentication Service зависит от контрактов Provider, а не от конкретных типов JWT, API Key, Basic или будущих Plugin Provider.

### Authentication

Authentication использует transport-neutral контракты, предложенные в [DP-004: Authentication Runtime Contracts](../proposals/DP-004-authentication-runtime-contracts.md).

Authentication не знает transport и не зависит от WebSocket. Transport adapters нормализуют входные данные до вызова Authentication. Authentication возвращает явный результат и immutable Principal для последующей Authorization.

### Dependency Injection

Composition root создает компоненты Runtime, разрешает их зависимости и связывает pipeline. Services не ищут зависимости через globals, service locators, доступ к Repository или HTTP API.

Решение требует dependency injection как архитектурный прием. Оно не выбирает dependency injection framework.

### Будущие компоненты Runtime

Архитектура резервирует независимые границы компонентов для:

- Authorization;
- Routing;
- Storage;
- Monitoring;
- Logging;
- Plugins.

Их подробные контракты и поведение требуют отдельных Proposal или решений до реализации.

## Последствия

### Преимущества

- Компоненты можно независимо тестировать с явными зависимостями.
- Реализации можно заменять без изменения несвязанных Services.
- Композиция и dependency injection остаются простыми.
- Связанность между Runtime, transport, persistence и Providers снижается.
- Новые компоненты и реализации Provider можно добавлять за стабильными контрактами.
- Структура поддерживает будущую Plugin Architecture.
- Поведение Runtime остается прослеживаемым до immutable Published ConfigurationVersion.

### Недостатки

- Архитектура вводит больше interfaces и компонентов.
- Композиция и управление lifecycle требуют явного проектирования.
- Совместимость контрактов необходимо строго поддерживать.
- Диагностика должна сохранять границы компонентов и при этом объяснять ошибки по всему pipeline.

## Ссылки

- [ADR-0001: Базовая реализация Control Service](0001-bootstrap-control-service.md)
- [ADR-0002: Configuration DSL](0002-configuration-dsl.md)
- [DP-001: Authentication](../proposals/DP-001-authentication.md)
- [DP-002: Secret References](../proposals/DP-002-secret-references.md)
- [DP-003: JWT Provider](../proposals/DP-003-jwt-provider.md)
- [DP-004: Authentication Runtime Contracts](../proposals/DP-004-authentication-runtime-contracts.md)


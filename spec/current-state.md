# Текущее состояние

**Веха:** M3 Listener Settings
**Статус реализации:** Control Service предоставляет in-memory API для Workspace, Configuration, Configuration Version и ListenerSettings.
**Release:** v0.1.0-alpha
**Architecture Review:** AR-001 — PASS

## Архитектурные решения

- ADR-001 закрепляет базовую реализацию Control Service.
- ADR-002 закрепляет ConfigurationVersion как декларативный Configuration DSL и единственный источник истины для будущего Runtime.
- Published ConfigurationVersion является immutable; Runtime исполняет ее без скрытой или альтернативной Configuration.
- Публичная схема Configuration DSL развивается обратно совместимо; несовместимые изменения требуют нового ADR.
- ADR-003 закрепляет компонентную архитектуру будущего Runtime, dependency injection и независимость от HTTP API и Repository.
- Runtime использует только immutable Configuration Snapshot, созданный из Published ConfigurationVersion.

## Состояние релиза

- Workspace CRUD завершен.
- Configuration CRUD завершен.
- ConfigurationVersion create, publish и archive завершены.

## Что существует

- Миссия и видение продукта
- Архитектурные принципы
- Структура спецификаций
- Соглашения по оформлению ADR
- Руководства для участников и агентов
- Правила исключения файлов из репозитория
- Go module для Go 1.25
- Исполняемый Control Service
- HTTP Server на Chi Router с endpoint `GET /health`
- Configuration адреса и уровня журнала через `UWP_HTTP_HOST`, `UWP_HTTP_PORT` и `UWP_LOG_LEVEL`
- Безопасные значения по умолчанию: `127.0.0.1:8080` и уровень журнала `info`
- Валидация Configuration до запуска сервиса
- Структурированное логирование с настраиваемым уровнем через `slog`
- HTTP timeout и graceful shutdown по `os.Interrupt` и `SIGTERM`
- Тесты загрузки Configuration и endpoint `GET /health`
- Доменная сущность Workspace с полями ID, Name, Description, CreatedAt и UpdatedAt
- Потокобезопасный in-memory Workspace repository с последовательными ID
- Service-слой с доменной валидацией Workspace и управлением временными метками
- HTTP CRUD API `/api/v1/workspaces` с единым форматом ошибок и строгой обработкой JSON
- Unit-тесты repository, service и HTTP handler Workspace
- Доменная сущность Configuration с обязательной принадлежностью существующему Workspace
- Потокобезопасный in-memory Configuration repository с последовательными ID
- Service-слой с Unicode-валидацией, UTC-временем и проверкой существования Workspace
- Вложенный HTTP CRUD API `/api/v1/workspaces/{workspaceID}/configurations`
- Запрет удаления Workspace, содержащего Configuration
- Unit-тесты repository, service и HTTP handler Configuration
- Доменная сущность ConfigurationVersion с последовательной нумерацией внутри Configuration
- Потокобезопасный in-memory ConfigurationVersion repository
- Создание Draft Version и получение списка через вложенный API `/api/v1/workspaces/{workspaceID}/configurations/{configurationID}/versions`
- Проверка существования Configuration перед созданием и чтением Version
- Unit-тесты repository, service и HTTP handler ConfigurationVersion
- Публикация Draft Version через endpoint `/api/v1/workspaces/{workspaceID}/configurations/{configurationID}/versions/{versionID}/publish`
- Атомарное архивирование предыдущей Published Version при публикации новой
- Инвариант единственной Published Version внутри Configuration
- Ручное архивирование Draft, Validated и Published Version через endpoint `/api/v1/workspaces/{workspaceID}/configurations/{configurationID}/versions/{versionID}/archive`
- Архивирование Published Version без автоматической публикации замены
- ListenerSettings с Host и Port для ConfigurationVersion
- Значения ListenerSettings по умолчанию `127.0.0.1:8080`
- Редактирование ListenerSettings только для Draft Version
- Валидация IP-адреса или hostname без DNS lookup и диапазона Port `1..65535`
- TLSSettings с Enabled, CertificateRef, PrivateKeyRef и MinVersion для ConfigurationVersion
- Редактирование TLSSettings только для Draft Version
- Ссылки на сертификат и закрытый ключ без хранения PEM или чтения файлов
- Поддержка минимальных версий TLS `1.2` и `1.3`
- TimeoutSettings с handshake, read, write и idle timeout для ConfigurationVersion
- Значения timeout задаются в секундах и редактируются только для Draft Version
- Значение `0` отключает только read и idle timeout; handshake и write требуют положительного значения

## Authentication Domain Model

- AuthenticationSettings как отдельная metadata-секция ConfigurationVersion
- Настройки Authentication с флагом Enabled и упорядочиваемыми по Priority Provider типа `jwt`, `api-key` и `basic`
- Полная замена AuthenticationSettings только для Draft Version через endpoint `/api/v1/workspaces/{workspaceID}/configurations/{configurationID}/versions/{versionID}/authentication`
- Валидация уникальности Name и Priority Provider при допустимом повторении Type
- API Key Provider metadata с Header и SecretRef внутри AuthenticationSettings
- Default Header `X-API-Key` и строгая валидация HTTP header field name
- Проверка формата SecretRef без разрешения ссылки и проверки существования Secret
- JWT Provider metadata с SigningKeys, AllowedAlgorithms, AllowedIssuers, AllowedAudiences, RequiredClaims и ClockSkewSeconds
- Signing Keys представлены SecretRef; поддерживаются algorithms HS, RS, ES и PS семейств с размерами 256, 384 и 512
- Default ClockSkewSeconds равен `60`; JWT metadata редактируется через общую секцию Authentication только для Draft Version
- Basic Authentication Provider metadata с Realm и SecretRef
- Default Realm `Universal WebSocket Platform`; SecretRef хранит только ссылку на будущие credentials
- AuthenticationValidator отделяет cross-provider и provider-specific business validation от ConfigurationVersion Service
- DefaultAuthenticationValidator не зависит от Repository, HTTP, Runtime или Persistence
- При включенной Authentication требуется минимум один enabled Provider; при выключенной Authentication настроенные Providers сохраняются и могут быть проигнорированы будущим Runtime
- Configuration domain не выполняет Authentication; Runtime API Key Provider описан ниже

## Secret References

- Принято направление Secret References: ConfigurationVersion хранит только ссылки на секреты, а не секретные значения
- Существующие CertificateRef и PrivateKeyRef соответствуют этому направлению
- Создан storage-neutral интерфейс Secret Resolver с общей валидацией и нормализацией Secret Reference
- Добавлена потокобезопасная in-memory реализация для тестирования и будущей локальной разработки
- Реальные Secret Storage backend еще не реализованы
- Resolver используется API Key и JWT Provider, но пока не подключен к Runtime Container или Authentication Pipeline

## JWT Provider Design

- DP-003 предлагает Configuration-модель JWT Provider с несколькими Signing Keys, algorithms, issuers, audiences и Required Claims
- Signing Keys представлены только Secret References без хранения PEM, JWK или HMAC secret в ConfigurationVersion
- JWT Provider metadata и Runtime Provider реализованы; Runtime поддерживает только HS256, HS384 и HS512, а полный pipeline отсутствует

## Authentication Runtime Contracts Design

- DP-004 предлагает transport-neutral контракты AuthenticationRequest, Principal, AuthenticationResult и AuthenticationProvider
- Предлагаемые контракты отделяют AuthenticationService и Provider от transport, Repository, Storage и внутреннего устройства ConfigurationVersion
- Модель ошибок различает rejected credentials, Provider error, Configuration error и Internal error
- Principal после успешной Authentication предлагается сделать immutable перед передачей в Authorization
- Созданы минимальные transport-neutral модели AuthenticationRequest, AuthenticationResult и Principal
- Созданы расширяемые интерфейсы Authentication Provider и Factory, принимающие AuthenticationProviderSnapshot и Secret Resolver
- Создан потокобезопасный Authentication Provider Registry с регистрацией Factory по Provider Type
- Registry делегирует создание Provider соответствующей Factory и не выполняет Authentication
- Реализован первый Runtime Authentication Provider для API Key с case-insensitive поиском Header
- API Key Provider разрешает Secret Reference при каждом Authenticate и сравнивает credentials через constant-time operation
- Реализован Authentication Service, последовательно вызывающий Provider в заданном порядке и завершающийся после первого успешного результата
- Реализован Authentication Bootstrap, собирающий Service из Authentication Snapshot через Provider Registry и Secret Resolver
- Реализован production API Key Factory, изолирующий преобразование AuthenticationProviderSnapshot в локальную runtime-конфигурацию API Key Provider
- Реализован Runtime JWT Provider с проверкой signature, exp, nbf, issuer, audience и Required Claims
- JWT Provider разрешает Signing Key через Secret Resolver при каждом Authenticate и поддерживает rotation без хранения Secret
- Реализован production JWT Factory, глубоко копирующий JWT metadata из AuthenticationProviderSnapshot в локальную runtime-конфигурацию Provider
- Интеграция Authentication Bootstrap в Runtime и Basic Provider по-прежнему не реализованы

## Runtime Architecture

- Принята последовательность компонентов от Configuration Loader и Configuration Snapshot до Monitoring
- Secret Resolver разрешает Secret References только при запуске Runtime; значения Secret остаются только в памяти процесса
- Authentication Provider Registry отделяет Runtime и Authentication Service от конкретных реализаций Provider
- Authentication использует transport-neutral контракты DP-004 и не зависит от WebSocket
- Реализована immutable Runtime Configuration Snapshot-модель для Listener и Authentication
- Builder принимает только Published ConfigurationVersion и глубоко копирует все Provider и JWT collections
- Snapshot не зависит от HTTP API, Repository или исходного ConfigurationVersion после создания
- Runtime Container хранит собственную глубокую копию Snapshot и возвращает новую копию через единственный метод `Snapshot()`
- Container пока не содержит других зависимостей и самостоятельно не управляет запуском, остановкой или reload Runtime
- Реализован потокобезопасный Runtime Host, владеющий независимой копией Snapshot и Container
- Host поддерживает однократный lifecycle `Created -> Running -> Stopped`; Restart и Reload отсутствуют
- Host пока не запускает Listener, Authentication или другие Runtime components
- Реализован Listener Bootstrap, создающий потокобезопасный Listener из ListenerSnapshot
- Listener хранит локальную копию Host, Port и TLS configuration и поддерживает lifecycle `Created -> Running -> Stopped`
- Listener открывает TCP socket и запускает HTTP Server с единым ответом `501 Not Implemented` для любого запроса
- Listener корректно завершает HTTP Server, accept loop и связанные goroutine через graceful shutdown
- Listener выполняет RFC 6455 WebSocket Upgrade через endpoint `GET /ws` и передает соединение Connection Dispatcher
- Immutable ConnectionContext содержит только context.Context, WebSocket connection и исходный HTTP request
- DefaultDispatcher сразу завершает переданное WebSocket-соединение с normal closure; Bootstrap позволяет внедрить другую реализацию Dispatcher
- Authentication подключена к WebSocket connection pipeline через AuthenticationDispatcher без зависимости Listener от пакета Authentication
- AuthenticationDispatcher преобразует HTTP handshake metadata в transport-neutral AuthenticationRequest и передает успешное соединение следующему AuthenticatedDispatcher вместе с immutable Principal context
- Отказ Authentication пока происходит после WebSocket Upgrade через close frame `PolicyViolation`, а системные ошибки используют `InternalError`
- Реализована минимальная WebSocket Session, которая после Authentication владеет соединением, хранит криптографически случайный ID, глубокую копию Principal, RemoteAddress и время создания
- Session Dispatcher создает Session из AuthenticatedContext и в текущей goroutine последовательно вызывает Start, блокирующий Run и завершающий Stop
- Session не хранит исходный HTTP Request, Headers, Query, credentials, AuthenticationRequest или transport context wrappers
- Добавлена immutable transport-neutral Runtime Message модель для text и binary application messages с копированием payload и UTC-временем получения
- Session удерживает WebSocket-соединение открытым и выполняет единственный блокирующий read loop до закрытия клиента, отмены context, Stop или ошибки чтения
- Session предоставляет потокобезопасный `Send(context.Context, message.Message)` для сериализованной отправки text и binary Runtime Message без raw `[]byte` API
- Прочитанные сообщения пока отбрасываются; Echo, Message Handler, Message Queue, Session Manager и Routing отсутствуют
- Архитектура Runtime принята в ADR-003, но Loader, подключение Resolver к Runtime Container и остальные компоненты pipeline еще не реализованы

## Чего не существует

- Персистентного хранения Workspace
- Персистентного хранения Configuration
- Validation, Rollback и lifecycle Snapshot для Configuration Version
- PostgreSQL
- Управления WebSocket-серверами
- Поведения Runtime для WebSocket-серверов
- WebSocket listener
- Реальный TLS listener и другие сетевые параметры Listener
- Применение Listener TimeoutSettings в Runtime
- Интеграция Authentication Bootstrap в Runtime и полный Authentication Pipeline
- Проверка Basic credentials
- Асимметричные JWT algorithms, JWKS, OIDC и token revocation
- Реальные Secret Storage backend и подключение Resolver к Runtime Container еще не реализованы
- Инфраструктуры развертывания
- Инфраструктуры хранения данных
- Admin UI

Этот файл описывает реализованное состояние репозитория, а не запланированные возможности продукта. Обновляйте его только при существенном изменении этого состояния.

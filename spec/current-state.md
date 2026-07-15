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
- ADR-004 закрепляет передачу в Handshake минимальных live read-only capabilities для Host-owned Admission Gate и Runtime context без зависимости от concrete Host.
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
- Resolver используется API Key и JWT Provider и подключен к production Authentication Pipeline через Runtime Host composition

## JWT Provider Design

- DP-003 предлагает Configuration-модель JWT Provider с несколькими Signing Keys, algorithms, issuers, audiences и Required Claims
- Signing Keys представлены только Secret References без хранения PEM, JWK или HMAC secret в ConfigurationVersion
- JWT Provider metadata и Runtime Provider реализованы; Runtime поддерживает только HS256, HS384 и HS512 и выполняет Provider через pre-Upgrade Authentication Pipeline

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
- Basic Provider по-прежнему не реализован

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
- Реализован потокобезопасный Runtime Host, являющийся production composition root и владеющий независимой копией Snapshot и Container
- Host поддерживает lifecycle `Created -> Built -> Starting -> Running -> Stopping -> Stopped`; Restart и Reload отсутствуют
- Runtime Bootstrap создает Built Host, а Host во время Start явно собирает Authentication, connection dispatch, Session handoff, Message Handler и Listener без service locator или DI framework
- Startup transaction публикует Listener только после успешного запуска и выполняет rollback полученного ресурса при ошибке, сохраняя исходную и rollback errors
- Host создает независимый root Runtime context после успешного запуска Listener; startup context не становится lifecycle context запущенного Runtime
- Runtime readiness становится true только после startup commit и сбрасывается в false в начале Stop
- Host владеет lifecycle-only Admission Gate, который открывается только в Running и закрывается до вызова Listener Stop
- Реализован Listener Bootstrap, создающий потокобезопасный Listener из ListenerSnapshot
- Listener хранит локальную копию Host, Port и TLS configuration и поддерживает lifecycle `Created -> Running -> Stopped`
- Listener открывает TCP socket и запускает HTTP Server с единым ответом `501 Not Implemented` для любого запроса
- Listener корректно завершает HTTP Server, accept loop и связанные goroutine через graceful shutdown
- Listener передает `GET /ws` выделенному Handshake Handler; `websocket.Accept` выполняется только после начальной проверки Admission Gate, Authentication Allow Decision и финальной проверки Gate
- Immutable ConnectionContext содержит derived Runtime context, WebSocket connection и исходный HTTP request, используемый только синхронно при handoff
- DefaultDispatcher сразу завершает переданное WebSocket-соединение с normal closure; Bootstrap позволяет внедрить другую реализацию Dispatcher
- Production composition передает Handshake только read-only Admission capability и Runtime Context Provider; concrete Runtime Host в Handshake не передается
- Handshake преобразует HTTP metadata в transport-neutral AuthenticationRequest и выполняет Authentication до `websocket.Accept`
- Authentication Reject и operational error предотвращают Upgrade и возвращаются как HTTP rejection; Session создается только после успешного Upgrade
- Runtime composition явно передает Handshake и Listener минимальный callback для terminal operational errors без diagnostics registry, event bus или глобального состояния
- Handshake сохраняет через `errors.Is` причину Session handoff failure в безопасной error-категории; Listener аналогично передает unexpected `http.Server.Serve` failure
- Штатные `http.ErrServerClosed` и `net.ErrClosed` при Listener shutdown не создают ложные terminal error reports
- Disabled Authentication формирует explicit anonymous Principal без запуска Provider
- Реализована минимальная WebSocket Session, которая после Authentication владеет соединением, хранит криптографически случайный ID, глубокую копию Principal, RemoteAddress и время создания
- Session Dispatcher создает Session из AuthenticatedContext и в текущей goroutine последовательно вызывает Start, блокирующий Run и завершающий Stop
- Создан независимый пакет `internal/sessionmanager` с потокобезопасным lifecycle skeleton `Open -> Closing -> Closed`
- Session Manager предоставляет только неблокирующий идемпотентный `BeginShutdown`, явный `Wait` и read-only наблюдение состояния; в `Open` метод `Wait` возвращает sentinel error и не начинает shutdown
- Базовые типы `SessionID`, `RegistrationID` и `RegistrationView` не содержат Session и пока не подключены к registry или Runtime composition
- Session не хранит исходный HTTP Request, Headers, Query, credentials, AuthenticationRequest или transport context wrappers
- Добавлена immutable transport-neutral Runtime Message модель для text и binary application messages с копированием payload и UTC-временем получения
- Session удерживает WebSocket-соединение открытым и выполняет единственный блокирующий read loop до закрытия клиента, отмены context, Stop или ошибки чтения
- Session предоставляет потокобезопасный `Send(context.Context, message.Message)` для сериализованной отправки text и binary Runtime Message без raw `[]byte` API
- Добавлен transport-neutral Runtime Message Handler contract; Session передает ему каждое прочитанное Message, а при nil Handler сохраняет discard-поведение
- Реализован EchoHandler, возвращающий неизмененные text и binary Runtime Message исключительно через Session Send без доступа к WebSocket transport
- Router, Middleware, Message Queue, Broadcast, Session Manager registry/accounting и Persistence отсутствуют
- Архитектура Runtime принята в ADR-003; pre-Upgrade Handshake реализован в объеме Authentication, а Configuration Loader, полный Session shutdown tracking, operational diagnostics и supervision еще отсутствуют

## Чего не существует

- Персистентного хранения Workspace
- Персистентного хранения Configuration
- Validation, Rollback и lifecycle Snapshot для Configuration Version
- PostgreSQL
- Управления WebSocket-серверами
- Поведения Runtime для WebSocket-серверов
- Реальный TLS listener и другие сетевые параметры Listener
- Применение Listener TimeoutSettings в Runtime
- Полный Handshake Pipeline за пределами Authentication: configured timeout enforcement, Session shutdown wait set и operational diagnostics
- Проверка Basic credentials
- Асимметричные JWT algorithms, JWKS, OIDC и token revocation
- Реальные Secret Storage backend и подключение Resolver к Runtime Container еще не реализованы
- Инфраструктуры развертывания
- Инфраструктуры хранения данных
- Admin UI

Этот файл описывает реализованное состояние репозитория, а не запланированные возможности продукта. Обновляйте его только при существенном изменении этого состояния.

## Runtime Alpha Architecture Review

- 2026-07-14 выполнено двуязычное [Runtime Alpha Architecture Review](../docs/ru/reviews/runtime-alpha-review.md) ([English version](../docs/en/reviews/runtime-alpha-review.md)).
- Итоговая оценка: `Ready with findings`.
- Подтверждены immutable Snapshot, явный dependency injection, отсутствие import cycles и зависимости Runtime от Control Plane Repository, transport-neutral границы Authentication и Message, а также явная передача владения WebSocket-соединением.
- Authentication после WebSocket Upgrade и отсутствие production composition в Runtime Host устранены; неполная ограниченность lifecycle shutdown по context остается открытым finding.
- Проект остается alpha foundation и не заявляется как production-ready.

## Runtime Architectural Pattern

- Создано двуязычное активное архитектурное руководство [ARCH-001: Runtime Architectural Pattern](../docs/ru/architecture/ARCH-001-runtime-architectural-pattern.md) ([English version](../docs/en/architecture/ARCH-001-runtime-architectural-pattern.md)).
- ARCH-001 обобщает подтвержденный Alpha-вертикалью паттерн `Context -> Evaluation -> Decision -> Execution` без создания универсального Policy Engine или новых обязательных Go-контрактов.
- Зафиксированы Configuration First, проверяемые границы зависимостей, явная передача владения mutable resources, lifecycle и concurrency requirements, а также принцип Boring Core.
- Handshake Pipeline, Router, Delivery, Persistence и Plugin ABI остаются предметом будущих DP и при необходимости ADR; ARCH-001 не определяет их API.

## Master Engineering Plan

- Создан двуязычный живой инженерный [MASTER_PLAN](../docs/ru/roadmap/MASTER_PLAN.md) ([English version](../docs/en/roadmap/MASTER_PLAN.md)).
- План разделяет стадии зрелости Alpha, Beta, RC, 1.0 и 2.0+ без календарных сроков, performance promises или проектирования API будущих подсистем.
- Для Beta выделены эпики Handshake, Runtime Host, lifecycle hardening, Configuration validation, Router, Session Manager, Delivery, Persistence, TLS, Metrics, operational diagnostics и Plugin contracts.
- Обязательные свойства 1.0 отделены от возможностей 1.x и отложенных distributed-возможностей 2.0+.
- MASTER_PLAN не является release schedule, backlog или заменой DP, ADR, current state и архитектурных reviews.

## Runtime Handshake Pipeline Design

- Создан двуязычный Draft design [DP-001: Runtime Handshake Pipeline](../docs/ru/design/DP-001-runtime-handshake-pipeline.md) ([English version](../docs/en/design/DP-001-runtime-handshake-pipeline.md)).
- Выбрана концептуальная последовательность `Transport -> Handshake Context -> Evaluation -> Decision -> Upgrade -> Session`.
- Design переносит обязательную Authentication до WebSocket Upgrade, сохраняя transport-neutral Authentication Service и ownership Session после успешной передачи.
- Listener остается владельцем HTTP/WebSocket transport effects и не получает Provider-specific logic.
- Реализация следует основному порядку DP-001: Admission Gate, Authentication, Allow Decision, финальная проверка Gate, Upgrade и Session handoff.
- Origin Policy, rate limiting, maintenance, IP filtering, Router, Session Manager и Plugin ABI остаются future work без зафиксированных API.

## Runtime Host Composition Root Design

- Создан двуязычный Draft design [DP-002: Runtime Host Composition Root](../docs/ru/design/DP-002-runtime-host-composition-root.md) ([English version](../docs/en/design/DP-002-runtime-host-composition-root.md)).
- Runtime Host предложен как единственная production composition root одного экземпляра Runtime с явными dependency graph, startup rollback, shutdown ordering и readiness boundary.
- Определён lifecycle `Created -> Initialized -> Starting -> Running -> Stopping -> Stopped` с terminal state `Failed` и запретом Restart и in-place Reload.
- Host владеет root Runtime context, запускает Listener последним и закрывает admission до cancellation и cleanup в обратном порядке.
- Container не превращается в service locator; DI framework, reflection, generic component factories и Universal Component Registry запрещены.
- После публикации DP-002 реализована его фундаментальная часть: Host стал production composition root, получил startup transaction, root Runtime context, readiness и lifecycle-only Admission Gate; `Failed`, supervision и полный shutdown wait set пока отсутствуют.

## Runtime Session Manager Design

- Создан двуязычный Draft design [DP-003: Runtime Session Manager](../docs/ru/design/DP-003-runtime-session-manager.md) ([English version](../docs/en/design/DP-003-runtime-session-manager.md)).
- DP-003 разделяет ownership WebSocket, registration transaction, выполнение Session и tracking Manager; Session сохраняет единоличный ownership WebSocket после transport handoff.
- `BeginShutdown` и `Wait` разделяют неблокирующий transition shutdown и ожидание, а атомарный `Complete` предлагается как единственная linearization point удаления будущей committed registration.
- Runtime Host остается владельцем Admission Gate и корневого Runtime context; Listener, Authentication, Router, Delivery, Persistence и diagnostics не входят в ответственность Session Manager.
- Реализован только минимальный lifecycle skeleton Manager без интеграции с Runtime Host; Reserve, Commit, Abort, Complete, registry, lookup, Session accounting, execution owner, Stop capabilities и полный shutdown wait set отсутствуют.
- Текущий Session Dispatcher по-прежнему синхронно выполняет отдельную Session без Runtime-wide registration и tracking.

## Runtime Foundation Freeze

- Создан двуязычный [ARCH-002: Runtime Foundation Freeze](../docs/ru/architecture/ARCH-002-runtime-foundation-freeze.md) ([English version](../docs/en/architecture/ARCH-002-runtime-foundation-freeze.md)).
- Архитектурно стабильными признаны реализованные Runtime Host, production composition root, lifecycle, root Runtime context, startup transaction и rollback, readiness и lifecycle-only Admission Gate.
- Freeze фиксирует фактический lifecycle `Created -> Built -> Starting -> Running -> Stopping -> Stopped` и не объявляет реализованными предложенные в Draft DP-002 состояния `Initialized` или `Failed`.
- Session ownership в полном Runtime shutdown wait set, Router, Delivery, Persistence, Operational Diagnostics и supervision остаются открытой архитектурой.
- Изменение замороженных архитектурных обязанностей, ownership или lifecycle-семантики требует нового сфокусированного DP или ADR.

## Handshake Runtime Dependency Boundary

- Принят двуязычный [ADR-0004: Handshake Runtime Dependency Boundary](../docs/ru/adr/0004-handshake-runtime-dependencies.md) ([English version](../docs/en/adr/0004-handshake-runtime-dependencies.md)).
- Host остается единственным владельцем Admission Gate и cancellation корневого Runtime context; Handshake получает только живые read-only capabilities через явную constructor injection.
- Draft DP-001 и DP-002 синхронизированы с ADR-0004: composition bridge передает Handshake admission permission и Runtime context access без зависимости от concrete Host.
- Handshake должен проверять admission до Authentication и повторно непосредственно перед `websocket.Accept`; Runtime context holder создается вместе с Host и активируется только при успешном startup commit.
- Финальная проверка admission непосредственно перед `websocket.Accept` является linearization point входа в admission commit; закрытие Gate до нее запрещает Upgrade.
- Session context должен создаваться как дочерний от активного Runtime context, а не от `http.Request.Context()`; root `CancelFunc` Handshake не раскрывается.
- ADR-0004 реализован минимальным capability bridge: Handshake не зависит от concrete Host, а Session context создается от активного Runtime context.

# Текущее состояние

**Веха:** M3 Listener Settings
**Статус реализации:** Control Service предоставляет in-memory API для Workspace, Configuration, Configuration Version и ListenerSettings.
**Release:** v0.1.0-alpha
**Architecture Review:** AR-001 — PASS

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
- Реальная Authentication и выполнение Provider не реализованы

## Secret References

- Принято направление Secret References: ConfigurationVersion хранит только ссылки на секреты, а не секретные значения
- Существующие CertificateRef и PrivateKeyRef соответствуют этому направлению
- Secret Storage и Secret Resolver еще не реализованы

## JWT Provider Design

- DP-003 предлагает Configuration-модель JWT Provider с несколькими Signing Keys, algorithms, issuers, audiences и Required Claims
- Signing Keys представлены только Secret References без хранения PEM, JWK или HMAC secret в ConfigurationVersion
- JWT Provider metadata реализована; проверка token и Runtime pipeline отсутствуют

## Чего не существует

- Персистентного хранения Workspace
- Персистентного хранения Configuration
- Validation, Rollback и Snapshot для Configuration Version
- PostgreSQL
- Управления WebSocket-серверами
- Поведения Runtime для WebSocket-серверов
- WebSocket listener и запуск TCP listener
- Реальный TLS listener и другие сетевые параметры Listener
- Применение Listener TimeoutSettings в Runtime
- Реальная Authentication, проверка JWT, API Key и Basic credentials
- AuthenticationService, AuthenticationRequest, Principal и выполнение Provider
- Secret Storage и Secret Resolver еще не реализованы
- Инфраструктуры развертывания
- Инфраструктуры хранения данных
- Admin UI

Этот файл описывает реализованное состояние репозитория, а не запланированные возможности продукта. Обновляйте его только при существенном изменении этого состояния.

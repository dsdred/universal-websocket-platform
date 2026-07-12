# Текущее состояние

**Веха:** M2 Configuration Version
**Статус реализации:** Control Service предоставляет in-memory API для Workspace, Configuration и метаданных Configuration Version.
**Release:** v0.1.0-alpha
**Architecture Review:** AR-001 — PASS
**Следующая веха:** M3 Listener Settings

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

## Чего не существует

- Персистентного хранения Workspace
- Персистентного хранения Configuration
- Validation, Rollback и Snapshot для Configuration Version
- PostgreSQL
- Управления WebSocket-серверами
- Поведения Runtime для WebSocket-серверов
- Инфраструктуры развертывания
- Инфраструктуры хранения данных
- Admin UI

Этот файл описывает реализованное состояние репозитория, а не запланированные возможности продукта. Обновляйте его только при существенном изменении этого состояния.

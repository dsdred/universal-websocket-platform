# Universal WebSocket Platform

[English version](README.md)

Universal WebSocket Platform — open-source платформа для создания, настройки, развертывания и эксплуатации независимых WebSocket-серверов без написания инфраструктурного кода.

## Статус

Проект находится на инженерной вехе **Beta — Complete the Single-Node
Runtime** и не готов к production-эксплуатации. Репозиторий содержит Control
Service, in-memory доменные API и собранную в production single-node Runtime
vertical с Authentication до Upgrade, детерминированной маршрутизацией,
транзакционной передачей Session и Manager-aware shutdown.

Configuration Loader, Snapshot Builder, Runtime Bootstrap и Launcher,
Lifecycle Owner, management routing, operational identity, command
idempotency и orchestration prerequisites также существуют изолированно. Они
не подключены к пути управления Runtime из Control Service: Production
Activation, реализации external persistence, recovery и reporting, полное
исполнение TLS и Listener settings и Control Service Runtime API отсутствуют.

## Текущий релиз

**Версия:** `v0.1.0-alpha`

**Зрелость релиза:** alpha

Этот опубликованный релиз содержит Control Service и базовый жизненный цикл
Workspace, Configuration и ConfigurationVersion. Текущий репозиторий содержит
более поздний, ещё не выпущенный инженерный прогресс Beta, описанный выше.
Подробности релиза приведены в
[заметках к релизу](docs/ru/releases/v0.1.0-alpha.md).

## Принципы проекта

- Configuration over Code
- Runtime Isolation
- API First
- Technology Neutrality
- Provider-based architecture
- Explainability
- Predictability
- Keep MVP Simple

## Документация

- [Главная страница документации](docs/ru/README.md)
- [Архитектурные руководства](docs/ru/architecture/README.md)
- [Документы проектирования Runtime](docs/ru/design/README.md)
- [Архитектурные решения](docs/ru/adr/README.md)
- [Инженерный план развития](docs/ru/roadmap/README.md)
- [Архитектурные ревью](docs/ru/reviews/README.md)
- [Текущее состояние реализации](spec/current-state.md)
- [Инженерная Wiki](wiki/README.md)
- [Заметки к релизам](docs/ru/releases/)
- [Внутренние спецификации](spec/README.md)
- [Точка входа AI-агентов](AGENTS.md)
- [Внутренний инженерный процесс](docs/engineering/AGENT.md)
- [Задачи и отчёт о синхронизации](docs/tasks/README.md)

## Участие в разработке

Проект находится на активной инженерной стадии Beta. Перед тем как предлагать
изменения, прочитайте русскоязычную документацию и фактическое состояние
реализации. Архитектурные решения следует фиксировать до того, как они
превратятся в ограничения реализации.

## Лицензия

Условия лицензии приведены в файле `LICENSE`.

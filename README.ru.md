# Universal WebSocket Platform

[English version](README.md)

Universal WebSocket Platform — open-source платформа для создания, настройки, развертывания и эксплуатации независимых WebSocket-серверов без написания инфраструктурного кода.

## Статус

Проект находится на ранней alpha-стадии и не готов к production-эксплуатации. Репозиторий содержит Control Service, in-memory доменные API и реализованную single-node Runtime vertical, для которой ещё завершается Manager-aware shutdown integration.

## Текущий релиз

**Версия:** `v0.1.0-alpha`

**Статус:** early alpha

Релиз содержит Control Service и базовый жизненный цикл Workspace, Configuration и ConfigurationVersion. Подробности приведены в [заметках к релизу](docs/ru/releases/v0.1.0-alpha.md).

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

## Участие в разработке

Проект находится на ранней стадии. Перед тем как предлагать изменения, прочитайте русскоязычную документацию и внутренние спецификации. Архитектурные решения следует фиксировать до того, как они превратятся в ограничения реализации.

## Лицензия

Условия лицензии приведены в файле `LICENSE`.

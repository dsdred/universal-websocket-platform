# Universal WebSocket Platform

Universal WebSocket Platform — open-source платформа для создания, настройки, развертывания и эксплуатации независимых WebSocket-серверов без написания инфраструктурного кода.

## Статус

Проект находится на ранней alpha-стадии. Репозиторий содержит инженерную основу, Control Service и первые in-memory доменные API.

## Текущий релиз

**Версия:** `v0.1.0-alpha`

**Статус:** early alpha

Релиз содержит Control Service и базовый жизненный цикл Workspace, Configuration и ConfigurationVersion. Подробности приведены в [`CHANGELOG.md`](CHANGELOG.md).

## Принципы проекта

- Configuration over Code
- Runtime Isolation
- API First
- Technology Neutrality
- Provider-based architecture
- Explainability
- Predictability
- Keep MVP Simple

## Структура репозитория

- [`spec/`](spec/) — спецификации продукта, архитектуры, решений и текущего состояния.
- [`.ai/PROJECT_CONTEXT.md`](.ai/PROJECT_CONTEXT.md) — краткий контекст для работы с помощью AI.
- [`AGENTS.md`](AGENTS.md) — инструкции по участию в проекте для автоматизированных агентов разработки.
- [`CONTRIBUTING.md`](CONTRIBUTING.md) — процесс и правила участия в разработке.

## Участие в разработке

Проект находится на ранней стадии. Перед тем как предлагать изменения, прочитайте [`CONTRIBUTING.md`](CONTRIBUTING.md) и спецификации. Архитектурные решения следует фиксировать до того, как они превратятся в ограничения реализации.

## Лицензия

См. [`LICENSE`](LICENSE).

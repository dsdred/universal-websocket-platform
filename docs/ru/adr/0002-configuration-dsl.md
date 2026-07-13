# ADR 0002: Configuration DSL

## Статус

Принято (Accepted).

## Контекст

В ходе предыдущих этапов в проекте сформировалась декларативная модель ConfigurationVersion. Metadata Listener, TLS, timeout и Authentication представлены явными разделами, которые можно создавать, валидировать, версионировать, публиковать и анализировать независимо от их будущего поведения в Runtime.

Эта модель становится публичным контрактом для management API, persistence, Import/Export и будущих инструментов. Поэтому до начала реализации Runtime необходимо явно закрепить архитектурную роль ConfigurationVersion.

## Решение

ConfigurationVersion является декларативным предметно-ориентированным языком (DSL) для описания WebSocket-сервера.

Runtime не создает и не поддерживает отдельную Configuration-модель. Runtime исполняет Published ConfigurationVersion.

### Configuration First

Любая новая возможность платформы сначала представляется в Configuration DSL и только после этого получает поведение в Runtime. Реализация Runtime не должна вводить Configuration, отсутствующую в публичной модели ConfigurationVersion.

### Metadata Before Behavior

Возможность разрабатывается в следующем порядке:

1. проектирование metadata;
2. validation;
3. persistence;
4. поведение Runtime.

Такая последовательность позволяет проверить и объяснить публичную модель до того, как от нее начнет зависеть эксплуатационное поведение.

### Published Configuration

Published ConfigurationVersion является единственным источником истины для Runtime. Runtime никогда не изменяет Published ConfigurationVersion.

Изменения подготавливаются в Draft и становятся доступны Runtime только через lifecycle ConfigurationVersion. Производное состояние Runtime не превращается в скрытую Configuration.

### Independent Sections

Каждый верхнеуровневый раздел ConfigurationVersion независим. Ожидаемые разделы включают:

- Listener;
- Authentication;
- Authorization;
- Routing;
- Storage;
- Monitoring;
- Logging.

Изменение или расширение одного раздела не должно требовать несвязанных изменений схемы остальных разделов. Cross-section validation может задавать явные инварианты, но не должна объединять независимые области в скрытую модель.

### Backward Compatibility

Публичная схема Configuration DSL развивается обратно совместимым способом. Существующие опубликованные документы и API clients должны оставаться интерпретируемыми при добавлении необязательных возможностей.

Несовместимое изменение схемы требует нового ADR с описанием migration и compatibility strategy.

### Secrets

Configuration хранит только Secret References и никогда не хранит реальные секретные значения. Значения Secret, private keys, passwords, tokens и другие чувствительные данные находятся вне Configuration DSL.

Правила ссылок и будущая граница их разрешения определены в [DP-002: Secret References](../proposals/DP-002-secret-references.md).

### Runtime

Runtime является исполнителем Configuration DSL. Он не содержит скрытую Configuration, альтернативные defaults, противоречащие публичной модели, или вторую независимо развивающуюся схему.

Runtime может формировать эксплуатационное состояние и диагностику, но они не являются Configuration и не должны изменять Published ConfigurationVersion.

## Последствия

### Преимущества

- Единая модель для REST API.
- Единая модель для представления YAML.
- Единая модель для Import и Export.
- Единая модель для persistence в PostgreSQL.
- Единая модель для будущего Admin UI.
- Единая модель для будущего Terraform Provider.
- Единая модель для будущего Kubernetes Operator.
- Поведение Runtime остается прослеживаемым до явной Published ConfigurationVersion.

### Недостатки

- Configuration DSL становится публичным контрактом.
- Изменения схемы требуют высокой дисциплины проектирования, validation, compatibility и migration.
- Реализация Runtime может ожидать стабилизации соответствующих metadata и lifecycle semantics.

## Ссылки

- [ADR-0001: Базовая реализация Control Service](0001-bootstrap-control-service.md)
- [DP-001: Authentication](../proposals/DP-001-authentication.md)
- [DP-002: Secret References](../proposals/DP-002-secret-references.md)
- [DP-003: JWT Provider](../proposals/DP-003-jwt-provider.md)

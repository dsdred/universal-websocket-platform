# DP-002: Secret References

**Статус:** Proposed

Русская версия является основной инженерной версией. Английская версия — официальный перевод.

## Мотивация

ConfigurationVersion хранится с историей версий, доступна через management API, а в будущем может экспортироваться или отображаться в Admin UI. Если хранить секретные значения внутри Configuration, они могут попасть в Git, экспорт JSON или YAML, журналы, историю версий и ответы Admin UI. После распространения по этим системам надежно удалить секрет уже невозможно.

Встроенные секретные значения также мешают безопасной ротации и противоречат неизменяемости Published ConfigurationVersion: для замены credentials пришлось бы либо изменять immutable-версию, либо публиковать Configuration, структура которой фактически не менялась.

ConfigurationVersion MUST NOT содержать реальные API keys, JWT symmetric secrets, private keys, PEM-содержимое сертификатов, OAuth2 client secrets, пароли и любые другие секретные значения.

ConfigurationVersion MAY содержать только ссылки: `secretRef`, `certificateRef`, `privateKeyRef` и другие типизированные ссылки на будущий Secret Storage.

## Цели

- Отделить metadata Configuration от секретных значений.
- Сохранить стабильную публичную JSON-модель без секретных данных.
- Позволить менять реализацию Secret Storage без изменения схемы Configuration.
- Поддержать будущую ротацию без встраивания новых значений в ConfigurationVersion.
- Допустить разные будущие backend: in-memory, PostgreSQL, filesystem, environment variables, HashiCorp Vault, Kubernetes Secrets и cloud secret managers.

## Вне целей

Документ не проектирует:

- конкретный API Secret Storage;
- шифрование at rest;
- RBAC для Secret Storage;
- процесс ротации;
- кеширование;
- интеграцию с HashiCorp Vault;
- интеграцию с Kubernetes;
- UI для управления секретами.

## Терминология

- **Secret** — чувствительные данные, используемые для аутентификации, подписи, шифрования, расшифрования или иного установления доверия. Например, passwords, API keys, tokens и private keys.
- **Secret Reference** — стабильная строка без секретного значения, которая идентифицирует Secret или другой защищенный объект.
- **Secret Storage** — backend, отвечающий за хранение и защиту значений Secret.
- **Secret Resolver** — компонент, который разрешает Secret Reference через Secret Storage и передает Secret авторизованной операции Runtime.
- **Secret Type** — ожидаемое назначение или представление Secret, например API key, JWT symmetric key, certificate или private key.

## Формат ссылок

Secret Reference:

- является строкой;
- не содержит значение Secret;
- остается стабильной, даже если значение по ссылке меняется;
- разрешается Runtime или отдельным Secret Resolver;
- не обязана проверяться на существование во время редактирования Draft;
- полностью проверяется перед публикацией или запуском Runtime.

Этот документ намеренно не фиксирует единый URI-like формат. Общий синтаксис можно выбрать после появления конкретных требований к Secret Storage и namespaces.

Допустимые примеры:

```text
secrets/api-keys/internal
secrets/jwt/main
certificates/default/server
workspace/main/oauth/client-secret
```

Недопустимые примеры:

```text
actual-secret-value
-----BEGIN PRIVATE KEY-----
C:\certs\server.key
https://example.com/secret
```

Первый недопустимый пример является значением, а не ссылкой. Остальные раскрывают секретный материал, связывают Configuration с локальным filesystem path или вводят еще не принятый формат удаленного расположения.

## Поведение Runtime

Будущий поток разрешения ссылки:

```text
ConfigurationVersion
        ↓
Secret Reference
        ↓
Secret Resolver
        ↓
Secret Storage
        ↓
Resolved Secret in memory
```

Runtime MUST NOT:

- записывать resolved Secret в журнал;
- возвращать resolved Secret через management API;
- сохранять resolved Secret обратно в Configuration;
- включать значение Secret в ошибки.

Разрешать ссылку следует только для авторизованной операции и как можно ближе к моменту использования.

## Версионирование

Published ConfigurationVersion остается immutable. Ее `secretRef` не меняется, при этом значение Secret в Secret Storage MAY изменяться. Такое разделение позволяет выполнять ротацию без изменения публичной JSON schema и без переписывания истории опубликованных Configuration.

Будет ли Runtime получать новое значение через hot reload, разрешать ссылку только при запуске или закреплять конкретную версию Secret, остается открытым вопросом.

## Принципы безопасности

- Применять least privilege при разрешении ссылок и доступе к Secret.
- Не хранить значения Secret в Configuration.
- Не записывать значения Secret в журналы.
- Не включать значения Secret в экспорт.
- Не включать значения Secret в сообщения об ошибках.
- Хранить resolved Secret в памяти только в течение необходимого времени.

## Совместимость с существующей моделью

Существующие поля `certificateRef` и `privateKeyRef` в TLSSettings соответствуют этому решению: они хранят ссылки и не содержат PEM, private-key contents или filesystem locations.

Будущие настройки API Key, JWT и OAuth2 должны использовать `secretRef` или другую типизированную Secret Reference. Они не должны вводить поля, хранящие само значение credentials.

## Открытые вопросы

- Нужен ли единый формат reference?
- Должны ли ссылки включать Provider namespace?
- На каком этапе lifecycle следует проверять существование Secret?
- Как выполнять и распространять rotation?
- Нужен ли version pinning для Secret?
- Какой компонент или identity имеет право разрешать ссылки?
- Как кешировать resolved Secrets?
- Как обрабатывать временную недоступность Secret Storage?

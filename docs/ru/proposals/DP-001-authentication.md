# DP-001: Authentication

**Статус:** Proposed

## Мотивация

Authentication устанавливает личность клиента до открытия WebSocket-соединения. Она отвечает на вопрос: «Кто этот клиент?»

Authentication не определяет, какие действия разрешены аутентифицированному клиенту. Решения о правах доступа относятся к Authorization и должны оставаться отдельной задачей.

## Цели

Начальная архитектура Authentication должна поддерживать:

- API Key
- JWT
- Basic

В будущей версии может быть добавлен OAuth2.

## Не входит в цели

В начальный объем не входят:

- OpenID Connect
- LDAP
- Kerberos
- SAML
- mTLS Authentication
- интеграция с External Secret Manager

## Высокоуровневая архитектура

```text
Client
  ↓
Handshake
  ↓
Authentication
  ↓
Principal
  ↓
Authorization
  ↓
WebSocket Session
```

- **Client** инициирует соединение и передает credentials через поддерживаемые входные данные handshake, например заголовки.
- **Handshake** принимает первоначальный запрос на соединение и создает не зависящий от транспорта AuthenticationRequest.
- **Authentication** проверяет переданные credentials с помощью настроенных Authentication Providers.
- **Principal** представляет установленную личность клиента и атрибуты, полученные в результате Authentication.
- **Authorization** определяет, может ли Principal выполнить запрошенное действие. Это отдельный от Authentication этап.
- **WebSocket Session** создается только после успешного завершения Authentication и Authorization.

Соединение нельзя переводить в WebSocket Session до завершения обязательных проверок личности и прав доступа.

## AuthenticationRequest

Authentication не должна зависеть от `net/http`. Предлагается модель запроса, не привязанная к транспорту:

```go
type AuthenticationRequest struct {
    Headers       map[string][]string
    RemoteAddress string
    TLS           bool
}
```

- `Headers` содержит значения заголовков handshake, необходимые Authentication Providers.
- `RemoteAddress` содержит сетевой адрес клиента в том виде, в котором его наблюдает Runtime.
- `TLS` указывает, защищен ли входящий транспорт с помощью TLS.

Runtime не следует передавать `http.Request` напрямую: это связало бы контракты Authentication и реализации Provider с одним транспортным адаптером. Отдельная модель упрощает тестирование Authentication, не позволяет Provider зависеть от постороннего HTTP-состояния и дает будущим транспортам Runtime возможность использовать тот же pipeline Authentication.

Runtime отвечает за копирование в AuthenticationRequest только необходимых данных запроса. В ходе этого преобразования нельзя записывать чувствительные значения в журнал.

## Principal

Аутентифицированная личность представляется моделью:

```go
type Principal struct {
    ID                     string
    AuthenticationProvider string
    Roles                  []string
    Claims                 map[string]string
}
```

- `ID` — стабильный идентификатор аутентифицированного клиента.
- `AuthenticationProvider` — имя Provider, который установил личность, например `jwt`, `api-key` или `basic`.
- `Roles` содержит имена ролей, переданные Authentication для последующей проверки в Authorization. Authentication не интерпретирует правила доступа этих ролей.
- `Claims` содержит дополнительные нормализованные атрибуты личности, доступные последующим компонентам.

Для `Claims` используется `map[string]string`, чтобы основная модель Principal оставалась предсказуемой, независимой от транспорта и от динамических типов конкретного Provider. Provider должен нормализовать выбранные claims в строки, а не раскрывать исходные структуры JWT или произвольные вложенные данные. Начальная модель намеренно ограничена; более богатые типизированные claims можно рассмотреть позже при появлении конкретных требований.

## Authentication Provider

Каждый механизм credentials реализуется через общий интерфейс:

```go
type AuthenticationProvider interface {
    Name() string

    Authenticate(
        ctx context.Context,
        request AuthenticationRequest,
    ) (*Principal, error)
}
```

`Name` возвращает стабильный идентификатор Provider, используемый в Configuration, диагностике и Principal.AuthenticationProvider.

`Authenticate` проверяет один запрос и возвращает Principal только тогда, когда Provider успешно установил личность. Ошибки Provider не должны раскрывать credentials или секретные данные.

Runtime зависит от этого интерфейса и AuthenticationService. Он не должен содержать логику разбора JWT, поиска API Key или проверки Basic credentials.

## AuthenticationService

AuthenticationService координирует Authentication Providers. Он:

- получает настроенный список Provider;
- упорядочивает Provider по приоритету;
- вызывает Provider в этом порядке;
- возвращает первый успешно полученный Principal;
- прекращает проверку после первого успешного результата;
- возвращает ошибку Authentication, если ни один Provider не сработал.

AuthenticationService отвечает только за оркестрацию. Он не содержит проверку JWT, API Key, Basic credentials или правил Configuration конкретных Provider.

Имена и приоритеты Provider следует проверять при подготовке ConfigurationVersion к использованию. Дублирующиеся или недоступные Provider должны приводить к понятной ошибке Configuration, а не к непредсказуемому порядку выполнения.

## Порядок Provider

Одновременно может быть включено несколько Provider. Их порядок задается в Configuration:

```yaml
authentication:
  providers:
    - type: jwt
      priority: 10
    - type: api-key
      priority: 20
    - type: basic
      priority: 30
```

Provider с меньшим числовым приоритетом выполняется первым. В этом примере JWT проверяется перед API Key, а API Key — перед Basic.

Проверка прекращается после первого успешного результата. Provider, следующие за успешным Provider, не вызываются. Поведение при одинаковых приоритетах необходимо определить до реализации; Configuration не должна полагаться на порядок объявления, пока такое правило явно не принято.

## Отключенная Authentication

Authentication можно явно отключить:

```yaml
authentication:
  enabled: false
```

В этом случае Runtime получает специальный Principal с ID, например `anonymous`:

```text
ID = "anonymous"
```

Anonymous Principal вместо `nil` дает Authorization и компонентам сессии единый объект личности. Это устраняет повторяющиеся проверки на nil, явно отражает анонимный режим в журналах и решениях и позволяет Authorization применять политики к неаутентифицированным клиентам через обычный pipeline.

Anonymous Principal должен отличаться от личностей, возвращенных настроенными Provider, и не должен получать неявных привилегий.

## Обработка ошибок

Если ни один настроенный Provider не смог аутентифицировать клиента, handshake возвращает:

```text
401 Unauthorized
```

После этого handshake завершается, а WebSocket-соединение не открывается.

Ответ клиенту должен содержать стабильное сообщение без чувствительных данных. Подробную диагностику Provider можно записывать для операторов, но журнал не должен содержать credentials, API Keys, tokens, passwords или другие секретные значения.

Для диагностики реализация должна отличать ожидаемый отказ credentials от внутреннего сбоя Provider, не возвращая клиенту внутренние ошибки и stack trace.

## Configuration

Authentication является отдельным разделом ConfigurationVersion:

```yaml
authentication:
  enabled: true
  providers:
    - type: jwt
      priority: 10
```

Authentication не входит в ListenerSettings. ListenerSettings описывает, как будущий Listener принимает соединения; Authentication определяет, как устанавливается личность клиента во время handshake.

Настройки конкретного Provider должны находиться внутри его Configuration и не должны проникать в контракты оркестрации Runtime.

## Будущие расширения

В будущем могут быть добавлены:

- OAuth2
- OpenID Connect
- LDAP
- Kerberos
- SAML
- Plugin Providers
- External Secret Managers

Это предложение не описывает их реализацию.

## Открытые вопросы

- Нужна ли поддержка refresh tokens и какой компонент должен управлять их жизненным циклом?
- Требуется ли token introspection для opaque tokens?
- Нужен ли распределенный кэш Authentication между экземплярами Runtime?
- Как распространять отзыв credentials и tokens?
- Как rotation секретов должна влиять на текущие и новые handshakes?
- Как AuthenticationService должен различать «credentials неприменимы» и «credentials отклонены» при проверке нескольких Provider?
- Следует ли запрещать одинаковые приоритеты Provider или разрешать их через дополнительное детерминированное правило?
- Какие claims следует нормализовать в Principal.Claims и как обрабатывать конфликты имен?
- Какие значения identity и Provider следует назначить anonymous Principal?

## Вне области документа

Этот документ не описывает:

- Authorization
- Session Management
- реализацию Runtime
- поведение WebSocket Protocol
- Secret Storage

Для этих областей нужны отдельные спецификации и решения.

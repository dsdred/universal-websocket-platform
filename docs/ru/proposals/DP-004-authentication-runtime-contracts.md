# DP-004: Authentication Runtime Contracts

**Статус:** Proposed

Русская версия является основной инженерной версией. Английская версия — официальный перевод.

## Мотивация

Configuration DSL уже описывает metadata Authentication, но в платформе еще нет transport-neutral контрактов для применения этой политики. Эти контракты должны отделить оркестрацию Authentication и реализации Provider от transport adapters и внутреннего устройства будущего Runtime.

Документ определяет логические границы и данные, передаваемые во время Authentication. Он не проектирует и не реализует поведение Runtime.

## Цели

- Определить transport-neutral AuthenticationRequest.
- Определить immutable Principal, представляющий успешно аутентифицированного или явно анонимного клиента.
- Определить AuthenticationResult и структурированную модель ошибок.
- Определить границу между AuthenticationService и реализациями AuthenticationProvider.
- Сохранить независимость будущих реализаций Provider от transport, persistence и внутреннего устройства ConfigurationVersion.

## Принципы проектирования

Контракты следуют ADR-0002:

- Configuration описывает политику Authentication до реализации поведения.
- AuthenticationService получает effective Configuration, но не передает ConfigurationVersion в Provider.
- Контракты Provider не содержат transport-specific типы запросов или соединений.
- Результат Authentication является явным и передается в Authorization без скрытого состояния.

Модели в документе являются логическими контрактами. Точные типы языка, форматы сериализации и границы пакетов будут определены будущей задачей реализации.

## AuthenticationRequest

AuthenticationRequest — нормализованное представление входных данных клиента. Transport adapter формирует его до начала Authentication.

| Поле | Назначение |
| --- | --- |
| `Headers` | Регистронезависимые именованные значения, переданные клиентом. Несколько значений одного имени сохраняются. |
| `Query` | Именованные параметры запроса. Несколько значений одного имени сохраняются. |
| `RemoteAddress` | Адрес клиента в представлении доверенной transport boundary. |
| `Transport` | Стабильный идентификатор transport, а не объект реализации transport. |
| `Attributes` | Доверенные transport-neutral входные атрибуты, добавленные adapters платформы. |
| `Body` | Необязательные непрозрачные входные данные для Provider, которым они явно нужны. Отсутствие отличается от пустого значения. |
| `RequestContext` | Данные отмены, deadline, correlation и request-scoped управления без раскрытия объекта transport request. |

AuthenticationRequest MUST NOT содержать HTTP request, transport connection, session или состояние реализации Runtime. Нормализация дает всем Provider одинаковую форму входных данных, не позволяет transport API проникать в код Provider и делает контракты независимо тестируемыми.

Headers, значения Query, Body и Attributes считаются недоверенными, если platform boundary явно не обозначила их происхождение как доверенное. Реализации не должны по умолчанию записывать в журнал значения, которые могут содержать credentials.

## Principal

Principal представляет identity, полученную при успешной Authentication, либо явную anonymous identity, используемую при отключенной Authentication.

| Поле | Назначение |
| --- | --- |
| `ID` | Стабильный идентификатор identity в применимой identity domain. |
| `Name` | Человекочитаемое имя identity, если оно доступно. |
| `AuthenticationType` | Механизм Authentication, установивший identity. |
| `Claims` | Нормализованные identity assertions, полученные при Authentication. |
| `Roles` | Имена ролей, передаваемые как identity context для последующих решений Authorization. |
| `Attributes` | Дополнительные нормализованные атрибуты identity. |
| `Anonymous` | Признак явного anonymous Principal. |
| `Authenticated` | Признак успешной Authentication credentials. |
| `Metadata` | Несекретные диагностические metadata и сведения о происхождении identity. |

После успешной Authentication Principal является immutable. AuthenticationService, Providers и последующие потребители MUST NOT изменять его на месте. Это не позволяет identity измениться между Authentication и Authorization и делает конкурентное использование предсказуемым.

`Anonymous` и `Authenticated` выражают разные факты и должны оставаться согласованными. Principal, аутентифицированный по credentials, имеет установленный `Authenticated` и не является anonymous. Явный anonymous Principal является anonymous и не считается аутентифицированным по credentials. Точные defaults для anonymous остаются открытым вопросом.

Claims, Roles, Attributes и Metadata являются нормализованными данными контракта, а не решениями о доступе. Их наличие не предоставляет разрешений; за их интерпретацию отвечает Authorization.

## AuthenticationResult

AuthenticationResult фиксирует результат одной попытки Provider или окончательное решение AuthenticationService.

| Поле | Назначение |
| --- | --- |
| `Success` | Показывает, создала ли Authentication допустимый Principal. |
| `Principal` | Immutable Principal при успехе; отсутствует при ошибке. |
| `FailureReason` | Структурированная несекретная категория ошибки и безопасная диагностическая информация. |
| `ProviderName` | Configuration name Provider, ответственного за результат. |
| `ProviderType` | Тип Provider, связанный с результатом. |
| `Metadata` | Несекретные metadata результата для диагностики, audit integration и observability. |

Если `Success` равен true, Principal MUST присутствовать, а FailureReason MUST отсутствовать. Если `Success` равен false, аутентифицированный Principal MUST NOT возвращаться. Metadata никогда не должны содержать credentials, разрешенные Secrets, tokens, passwords или private key material.

## AuthenticationProvider

AuthenticationProvider имеет одну концептуальную операцию:

```text
Authenticate(AuthenticationRequest) -> AuthenticationResult
```

AuthenticationService передает нормализованные входные данные и получает явный результат. Provider реализует один механизм Authentication, но не принимает решений о правах доступа.

Provider ничего не знает о:

- Runtime;
- WebSocket или другом transport protocol;
- реализациях Storage или Repository;
- доменных объектах ConfigurationVersion;
- политике Authorization.

Provider-specific settings преобразуются в effective Provider во время composition. Метод `Authenticate` не читает их из repositories.

## AuthenticationService

AuthenticationService является границей оркестрации. Он получает effective Authentication policy и упорядоченные Providers, вызывает Providers согласно этой политике, интерпретирует результаты и возвращает окончательный AuthenticationResult, а при успехе — Principal.

AuthenticationService не содержит логику проверки JWT, API Key или Basic. Он не предоставляет права и не создает transport sessions.

## Authentication Pipeline

```text
Configuration
      |
      v
AuthenticationService
      |
      v
AuthenticationProvider(s)
      |
      v
AuthenticationResult
      |
      v
Principal
      |
      v
Authorization
```

Configuration выбирает и упорядочивает политику Authentication. AuthenticationService оркестрирует настроенные Providers. Providers оценивают нормализованные входные данные и возвращают AuthenticationResult. Успешный результат предоставляет immutable Principal подсистеме Authorization.

Эта последовательность описывает только границы контрактов. Она не определяет создание Runtime, его lifecycle или поведение transport.

## Модель ошибок

AuthenticationResult должен различать следующие результаты:

| Результат | Значение |
| --- | --- |
| `Success` | Credentials приняты либо настроенная anonymous policy создала явный Principal. |
| `Rejected credentials` | Входные данные распознаны, но не подтверждают допустимую identity. |
| `Provider error` | Provider не смог завершить операцию из-за специфической для него операционной ошибки. |
| `Configuration error` | Effective Authentication policy или настройка Provider недопустима либо неработоспособна. |
| `Internal error` | Неожиданная ошибка платформы помешала принять достоверное решение. |

Rejected credentials являются ожидаемым решением Authentication и не должны смешиваться с операционными ошибками. Provider, Configuration и Internal errors должны оставаться различимыми для безопасной диагностики и будущих policy decisions. FailureReason, раскрываемый за пределами доверенной границы, не должен содержать сведения о credentials или Secrets.

Документ не определяет status codes, transport responses, retry policy или необходимость продолжать pipeline после каждой категории ошибки.

## Явно вне области документа

- SecretResolver;
- проверка JWT;
- проверка Basic;
- проверка API Key;
- проектирование или реализация Runtime;
- Authorization;
- модели Session и Connection;
- rate limiting;
- caching.

## Связь с существующими документами

- [ADR-0002](../adr/0002-configuration-dsl.md) определяет ConfigurationVersion как декларативный Configuration DSL и источник истины для будущего исполнения.
- [DP-001](DP-001-authentication.md) определяет высокоуровневую архитектуру Authentication и порядок Provider.
- [DP-002](DP-002-secret-references.md) определяет границу между Configuration и значениями секретов.
- [DP-003](DP-003-jwt-provider.md) определяет предлагаемую Configuration metadata JWT Provider.

DP-004 уточняет будущие runtime-facing контракты данных, не изменяя Configuration-модели, предложенные или реализованные этими документами.

## Открытые вопросы

- Должен ли Provider timeout быть общим, индивидуальным для Provider или поддерживать оба уровня?
- Как передавать cancellation и представлять его в RequestContext?
- Будут ли поддерживаться asynchronous Providers?
- Какие Provider metrics необходимы и какие labels безопасны?
- Нужна ли distributed Authentication orchestration?
- Какие audit events относятся к AuthenticationService, а какие — к Providers?
- Как представить challenge support без привязки результата к transport?
- Какие поля и значения обязательны для anonymous Principal?


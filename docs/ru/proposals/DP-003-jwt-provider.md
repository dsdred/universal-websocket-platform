# DP-003: JWT Provider

**Статус:** Proposed

Русская версия является основной инженерной версией. Английская версия — официальный перевод.

## Мотивация

JWT Provider нужна декларативная Configuration-модель, описывающая допустимые JSON Web Tokens. Configuration задает политику проверки: доверенные Signing Keys, разрешенные algorithms, ожидаемые issuers и audiences, обязательные Claims и допустимый clock skew.

Configuration не определяет внутреннюю реализацию разбора JWT, криптографической проверки, обработки Claims или оркестрации Runtime.

## Цели

- Определить Configuration metadata будущего JWT Provider.
- Поддержать несколько доверенных Signing Keys.
- Поддержать несколько разрешенных algorithms, issuers и audiences.
- Декларативно задавать обязательные Standard Claims и Custom Claims.
- Не хранить секретные данные внутри ConfigurationVersion.
- Сохранить независимость модели от конкретной JWT library и реализации Runtime.

## Предлагаемая модель

Будущая Configuration JWT Provider должна содержать секцию JWTSettings следующего логического вида:

```text
JWTSettings
├── SigningKeys[]
├── AllowedAlgorithms[]
├── AllowedIssuers[]
├── AllowedAudiences[]
├── RequiredClaims[]
└── ClockSkewSeconds
```

- `SigningKeys` определяет один или несколько доверенных ключей подписи через Secret References.
- `AllowedAlgorithms` перечисляет все algorithms подписи, разрешенные политикой.
- `AllowedIssuers` содержит допустимые значения стандартного Claim `iss`.
- `AllowedAudiences` содержит допустимые значения стандартного Claim `aud`.
- `RequiredClaims` декларативно задает Claims, которые должны присутствовать, и в будущем может содержать ограничения значений.
- `ClockSkewSeconds` определяет допустимое отклонение времени при проверке временных Standard Claims.

Все коллекции допускают несколько значений. Точное представление в Go и JSON, обязательные поля, defaults, правила обработки duplicates и семантика пустых списков должны быть зафиксированы задачей реализации до того, как модель станет стабильным API-контрактом.

## Signing Keys

ConfigurationVersion MUST хранить Signing Keys только как Secret References. Она MUST NOT содержать:

- PEM-содержимое;
- встроенные документы JWK;
- значения HMAC secret.

Логическая запись Signing Key содержит `secretRef` и в будущем может включать несекретные metadata, необходимые для детерминированного выбора ключа. Значение Secret разрешается вне Configuration согласно DP-002.

Необходимо поддержать несколько Signing Keys, чтобы политика могла доверять нескольким активным ключам и позволяла подготовить rotation. Документ не определяет загрузку, кеширование, выбор или ротацию ключей.

## Algorithms

Начальная Configuration-модель должна распознавать следующие идентификаторы algorithms:

- `HS256`, `HS384`, `HS512`;
- `RS256`, `RS384`, `RS512`;
- `ES256`, `ES384`, `ES512`;
- `PS256`, `PS384`, `PS512`.

`AllowedAlgorithms` является явным allowlist. Runtime не должен определять algorithm по ключевому материалу или принимать algorithm только потому, что его поддерживает JWT library. Поведение algorithm negotiation этим документом не определяется.

## Claims

Claims JWT разделяются на две группы политики.

### Standard Claims

Standard Claims — зарегистрированные JWT Claims с определенной семантикой, включая `iss`, `sub`, `aud`, `exp`, `nbf`, `iat` и `jti`. Специализированные поля `AllowedIssuers`, `AllowedAudiences` и `ClockSkewSeconds` задают политику для Standard Claims, требующих особой обработки.

Будущая модель должна определить, какие временные Claims обязательны по умолчанию, а какие нужно явно перечислять в `RequiredClaims`.

### Custom Claims

Custom Claims — специфичные для приложения или организации Claims, не входящие в зарегистрированный набор. Configuration должна позволять декларативно требовать их наличия без исполняемых выражений и деталей реализации Provider.

Политика Required Claim должна идентифицировать имя Claim и в будущем может задавать несекретные ограничения, например ожидаемое строковое значение или набор допустимых значений. Точная модель ограничений остается открытым вопросом и должна сохранять стабильное и объяснимое JSON-представление.

## Политика валидации

Configuration должна позволять задавать:

- несколько issuers;
- несколько audiences;
- несколько Signing Keys;
- несколько разрешенных algorithms;
- несколько обязательных Standard Claims;
- несколько обязательных Custom Claims.

Валидация Configuration-модели отличается от валидации входящего token. Перед публикацией или запуском Runtime платформа должна уметь отклонять внутренне противоречивую политику, неподдерживаемые identifiers algorithms, некорректные Secret References и неоднозначные duplicates без разрешения или раскрытия значений Secret.

Семантика сопоставления нескольких issuers и audiences, включая выбор между «совпадает любое» и «совпадают все», требует явного решения до реализации.

## Примечания о Runtime

Документ не проектирует Runtime. Концептуальная будущая последовательность приведена только для обозначения места применения Configuration policy:

```text
Token
  ↓
Signature verification
  ↓
Claims validation
  ↓
Principal
  ↓
Authorization
```

Signature verification использует настроенные Signing Keys и AllowedAlgorithms. Claims validation применяет политики issuer, audience, обязательных Claims и времени. Создание Principal и Authorization являются отдельными областями и здесь не определяются.

## Явно вне области документа

Документ не проектирует:

- OpenID Connect;
- JWKS Discovery;
- OAuth2;
- обработку Refresh Token;
- Introspection token;
- Revocation token;
- базу данных пользователей;
- Authorization;
- Principal;
- реализацию Runtime.

## Открытые вопросы

- Следует ли в будущей версии поддержать встроенный JWK или ссылку на JWK?
- Следует ли поддержать JWKS URL и как согласовать его с правилами Secret References?
- Как `kid` должен выбирать ключ из нескольких Signing Keys?
- Допустим ли algorithm negotiation или policy всегда должна задавать выбор явно?
- Какое значение clock skew должно использоваться по умолчанию?
- Следует ли поддерживать nested JWT?
- Следует ли поддерживать encrypted JWT?
- Как представлять и выбирать несколько активных ключей?
- Какую rotation strategy использовать для Signing Keys?

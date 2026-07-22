# ARCH-005: Runtime Configuration Snapshot and Loading Model

[English version](../../en/architecture/ARCH-005-runtime-configuration-snapshot-and-loading-model.md)

**Статус:** Active

**Область:** Runtime Configuration Snapshot и граница загрузки между Control Plane и Runtime

**Стабильность:** Approved architectural model

## 1. Назначение

ARCH-005 определяет, что именно запускает Runtime: один immutable Runtime Configuration Snapshot, построенный из точной Published ConfigurationVersion, закреплённой за одним Launch Attempt.

Документ устанавливает границы между источниками Configuration, Configuration Loader, Builder, Snapshot, Runtime Bootstrap, Runtime Host и Runtime Services. Он определяет архитектуру и ownership, но не реализацию Loader.

Эта модель дополняет [ADR-0002](../adr/0002-configuration-dsl.md), [ADR-0003](../adr/0003-runtime-architecture.md), [ARCH-002](ARCH-002-runtime-foundation-freeze.md) и [ARCH-004](ARCH-004-runtime-deployment-and-identity-model.md). Она не изменяет Runtime Foundation или модель operational identity.

## 2. Контекст и область

Control Plane владеет декларативной Configuration и lifecycle её версий. Runtime владеет исполнением на основе стабильного входа. Прямая зависимость Runtime Services от Configuration repositories, HTTP-моделей, YAML-представлений или состояния публикации объединила бы эти ответственности и поставила бы поведение запущенного Runtime в зависимость от mutable management state.

Необходимая граница:

```text
Configuration source
    -> Published ConfigurationVersion
    -> Configuration Loader
    -> Builder
    -> immutable Runtime Configuration Snapshot
    -> Runtime Bootstrap
    -> Runtime Host
    -> Runtime Services
```

Документ определяет эту границу, content и provenance Snapshot, построение и ownership, ответственности validation и normalization, schema compatibility, concurrency и semantics отказов.

Документ не определяет:

- Go API или interfaces;
- repository, PostgreSQL или HTTP schemas;
- caching, polling или retry policies;
- persistence Runtime Instance;
- реализацию Runtime Launcher;
- reload, replacement или reconciliation;
- backend хранилищ секретов;
- migration Configuration schema или version negotiation.

## 3. Runtime Configuration Snapshot

Runtime Configuration Snapshot — immutable, detached и execution-ready вход для одного Runtime Host.

Snapshot не является:

- Configuration или ConfigurationVersion;
- entity repository Control Plane;
- HTTP DTO;
- моделью YAML-документа;
- записью persistence;
- Runtime service или mutable Runtime state.

Snapshot содержит effective configuration, необходимую поддерживаемому графу Runtime-компонентов, и provenance, необходимую для определения его декларативного и operational происхождения. Он не содержит ни management history, ни mutable operational state.

Один Snapshot принадлежит одному Launch Attempt. Runtime запускается из этого Snapshot и больше не обращается к исходной ConfigurationVersion на протяжении lifetime данного Host.

## 4. Граница Configuration-to-Runtime

Ответственность Control Plane заканчивается после предоставления точной Published ConfigurationVersion через границу Configuration Loader. Ответственность Runtime начинается с построения detached Snapshot из выбранного источника.

Через границу разрешена передача immutable data. Через неё не передаются:

- доступ к repository;
- authority публикации;
- mutation lifecycle Configuration;
- история ConfigurationVersion;
- ownership сервисов Control Plane;
- source-specific parsing или storage behavior.

Runtime не должен получать Configuration через запросы к management services. Control Plane не должен изменять Snapshot или Runtime-owned state после handoff.

## 5. Модель источников Configuration

Источник Configuration — адаптер, способный предоставить декларативную модель, необходимую границе Loader. PostgreSQL, YAML, tests, in-memory repositories, imports и будущие адаптеры являются возможными источниками.

Тип источника относится к реализации и deployment. Он не должен создавать второй язык Configuration или менять Runtime semantics. Каждый источник должен представлять одну и ту же domain model ConfigurationVersion и сохранять её identity, lifecycle state, schema identity и значения configuration.

Сейчас repository реализует in-memory storage Control Plane и прямые значения в тестах. PostgreSQL и YAML являются архитектурными представлениями, а не production loading capabilities, объявленными этим документом.

## 6. Published ConfigurationVersion

В качестве декларативного источника нового Snapshot может быть выбрана только точная Published ConfigurationVersion.

Версии Draft, Validated и Archived не могут быть выбраны для нового Launch Attempt. Loader обязан установить Published state и точную identity как одно согласованное наблюдение источника. Последующая публикация или архивация не изменяет уже выбранную версию, её Snapshot или активный Launch Attempt.

ConfigurationVersion остаётся во владении Control Plane. Runtime заимствует её immutable content только на время, необходимое для построения detached Snapshot. Runtime никогда не меняет её состояние и не записывает в неё derived values.

## 7. Граница Configuration Loader

Configuration Loader — архитектурная граница, которая получает точную Published ConfigurationVersion, выбранную для одного Launch Attempt, и предоставляет её для построения Snapshot.

Loader отвечает за:

- сохранение связи запрошенных identity Workspace, Configuration, ConfigurationVersion, Runtime Instance и Launch Attempt;
- возврат только согласованно наблюдаемого Published source;
- сохранение identity Configuration schema;
- сообщение об ошибках загрузки, identity, lifecycle state и целостности источника до построения Runtime;
- недопущение передачи Runtime-компонентам доступа к repository или management API.

Loader не:

- выбирает desired state или принимает решения о запуске;
- выбирает произвольную более новую версию после закрепления версии за Launch Attempt;
- строит Runtime-компоненты;
- нормализует Runtime configuration;
- разрешает Secret values;
- сохраняет ownership Snapshot;
- определяет caching, polling, persistence или retry policy.

ARCH-005 определяет эту ответственность, не определяя interface Loader, его методы или конкретные адаптеры.

## 8. Граница Builder

Builder преобразует одну загруженную Published ConfigurationVersion вместе с уже установленной operational provenance в один validated, normalized и detached Snapshot.

Builder отвечает за:

- defensive validation входа, необходимого для построения Snapshot;
- deterministic normalization согласно semantics домена Configuration;
- полное преобразование поддерживаемой behavior-affecting configuration;
- глубокое отделение от всей caller-owned mutable memory;
- построение обязательной provenance;
- отклонение unsupported или malformed input без публикации частичного Snapshot.

Builder не является Loader. Он не знает о repositories, HTTP, YAML, PostgreSQL, истории публикаций или management commands. Он не выбирает Configuration и не решает, следует ли запускать Runtime. Он не сохраняет source или построенный Snapshot после возврата.

## 9. Ownership построения Snapshot

Построением Snapshot владеет launch preparation на границе Runtime. Runtime Lifecycle Owner сначала создаёт Launch Attempt и закрепляет за ним точную Published ConfigurationVersion. Loader получает этот source, а Builder строит Snapshot для данного attempt.

Поток построения:

```text
Runtime Lifecycle Owner
    -> pins Launch Attempt and Published ConfigurationVersion
    -> invokes Configuration Loader
    -> supplies loaded source and operational provenance to Builder
    -> receives complete immutable Snapshot
    -> passes Snapshot through Runtime Launcher to Bootstrap
```

Snapshot не является архитектурно видимым до полного успешного завершения построения. Builder создаёт значение, но не владеет решением о запуске или lifetime полученного результата.

## 10. Модель содержимого Snapshot

Snapshot содержит ровно две архитектурные категории информации:

1. effective, behavior-affecting Runtime configuration, представленную выбранной Configuration schema;
2. stable provenance, идентифицирующую декларативный source и execution identity, для которой построен Snapshot.

Snapshot должен содержать каждое поддерживаемое значение, необходимое для composition и работы Runtime без возвращения в Control Plane. Он может сохранять настроенные capabilities, исполнение которых относится к утверждённому более позднему этапу, если существующие правила startup capabilities явно разрешают им оставаться неактивными.

Snapshot не должен содержать:

- Secret values;
- repositories, services, loaders или реализации resolver;
- mutable Runtime state;
- desired или actual lifecycle state;
- историю ConfigurationVersion;
- тип source adapter;
- timestamps загрузки, cache metadata или transport metadata;
- deployment overrides, не представленные утверждённой архитектурой.

Конкретный набор полей относится к реализации и здесь не определяется.

## 11. Provenance Snapshot

Каждый Snapshot должен нести достаточно immutable provenance для установления всех следующих связей:

- Workspace, владеющий Configuration;
- identity Configuration;
- точная identity и номер ConfigurationVersion;
- identity и версия Configuration schema, представленной Snapshot;
- Runtime Instance, для которого запрошен запуск;
- Launch Attempt, для которого построен Snapshot.

Provenance является identity, а не telemetry. Тип source adapter, время загрузки, длительность построения, identity процесса, Host pointer, PID и адрес socket не являются provenance Snapshot.

Конкретное представление и имена полей требуют focused implementation design. Их semantic identity и согласованность обязательны.

## 12. Модель schema compatibility

Snapshot представляет одну конкретную Configuration schema. Его provenance идентифицирует эту schema, чтобы граница Runtime могла доказать, что выбранная configuration интерпретируема текущими Builder и графом Runtime-компонентов.

Unsupported, unknown или incompatible schema должна приводить к отказу launch preparation до получения Runtime Host ресурсов или перехода в Ready. Runtime не должен угадывать отсутствующие semantics, молча понижать версию input или переинтерпретировать source через hidden defaults.

Документ не определяет schema negotiation, migration, compatibility ranges или versioning адаптеров. Эти механизмы требуют focused Design Proposal.

## 13. Слои validation

Validation является многоуровневой и не передаёт ownership между компонентами:

1. Адаптеры источников Configuration проверяют целостность представления и сохраняют domain model.
2. Control Plane проверяет Configuration domain и правила публикации.
3. Loader проверяет запрошенную identity, допустимость Published, согласованное наблюдение источника и его целостность.
4. Builder defensively проверяет полноту, поддерживаемую schema, cross-field инварианты Snapshot и deterministic conversion.
5. Runtime Bootstrap проверяет startup-critical capabilities до получения externally visible resources и до readiness.

Validation на более раннем слое не позволяет следующей trust boundary отказаться от defensive checks. Слои не должны противоречить acceptance semantics домена Configuration. Ошибка останавливает launch preparation или startup на слое, который владеет соответствующей ответственностью.

## 14. Граница normalization

Правила домена Configuration определяют semantic equality и canonical values. Адаптеры источников должны сохранять эти semantics и могут выполнять только representation-level decoding.

Builder является authoritative шагом Runtime-boundary normalization. Он создаёт один canonical Snapshot без изменения загруженной ConfigurationVersion. Эквивалентный valid input из разных источников должен создавать семантически эквивалентные значения Snapshot.

Loader не нормализует Runtime configuration. Bootstrap и Runtime Services не повторяют normalization Configuration и не устанавливают source-specific defaults. Defensive validation на нескольких слоях допускается, но её acceptance semantics должны оставаться эквивалентными.

## 15. Immutability Snapshot

Snapshot становится immutable при успешном завершении construction. С этого момента:

- ни один owner не может изменять его in place;
- caller-owned mutable memory не может быть alias его logical content;
- readers получают immutable views или detached copies;
- Runtime Services могут выполнять concurrent read без synchronization для mutation Snapshot;
- derived operational state должен храниться вне Snapshot;
- публикация новой ConfigurationVersion не может изменить его.

Immutability является гарантией ownership, а не просто соглашением о том, что callers не должны выполнять mutation.

## 16. Ownership и lifetime Snapshot

Ownership передаётся только в одном направлении:

| Этап | Ownership и разрешённое действие |
|---|---|
| ConfigurationVersion | Control Plane владеет immutable published source |
| Loader | Временно владеет операцией загрузки и detached loaded material; никогда не владеет lifetime Snapshot |
| Builder | Владеет только construction; не сохраняет ни source, ни result |
| Launch preparation | Владеет полным Snapshot до построения Runtime |
| Runtime Bootstrap | Принимает Snapshot для построения одного Host и передаёт Host независимое immutable value |
| Runtime Host | Владеет Snapshot весь lifetime Host |
| Runtime Services | Только читают; не владеют, не заменяют и не изменяют Snapshot |

Если Bootstrap завершается ошибкой до полного ownership transfer, launch preparation освобождает свои values и Host-visible Snapshot не возникает. После успешного построения Host сохраняет Snapshot во время startup, Running, shutdown и terminal completion. Lifetime его значения заканчивается, когда terminal Host и все разрешённые readers становятся недостижимы.

Snapshot не имеет явного destruction protocol, поскольку не содержит Secret values или независимо owned Runtime resources.

## 17. Граница Secret Reference

ConfigurationVersion и Snapshot содержат только Secret References. Loader и Builder сохраняют и проверяют references, но никогда не разрешают и не встраивают Secret values.

Secret values появляются только через явно composed Secret Resolver после построения Snapshot и остаются во владении consuming Runtime capability. Они не должны записываться обратно в ConfigurationVersion или Snapshot, включаться в provenance, логироваться, сохраняться или раскрываться через inspection Snapshot.

ARCH-005 не меняет существующие Authentication semantics. ADR-0003 описывает startup resolution обязательных references, тогда как некоторые текущие provider contracts разрешают references во время обработки запроса. Точное время resolution и lifetime значений для каждой consuming capability остаются предметом focused architectural follow-up; документ не разрешает это различие ни переносом Secret values в Snapshot, ни неявным изменением поведения provider.

## 18. Интеграция Launch Attempt

Runtime Lifecycle Owner создаёт один Launch Attempt и закрепляет одну точную Published ConfigurationVersion до начала загрузки. Построение Snapshot использует эту фиксированную пару и записывает provenance Runtime Instance и Launch Attempt.

Один Launch Attempt может создать не более одного успешного Snapshot и не более одного Runtime Host. Ошибка построения не оставляет runnable Host и регистрируется как failed historical Launch Attempt согласно ARCH-004.

Retry или replacement создаёт новый Launch Attempt и новый Snapshot. Повторное использование Snapshot разными Launch Attempts запрещено, даже если оба attempt выбирают одну ConfigurationVersion, поскольку их operational provenance различается.

## 19. Согласованность публикации и concurrency

Linearization point выбора source является согласованное наблюдение того, что точная версия, закреплённая за Launch Attempt, находится в состоянии Published. Всё до этой точки является выбором и validation; всё после неё использует эту immutable version identity.

Архитектура требует:

- ни Draft, ни Validated, ни уже Archived version не может пересечь границу выбора для нового attempt;
- публикация другой версии после выбора не перенаправляет attempt;
- concurrent publication не может создать Snapshot, составленный из нескольких версий;
- concurrent starts одного Runtime Instance остаются под правилом единственного active Launch Attempt из ARCH-004;
- starts разных Runtime Instances могут независимо выбрать одну Published ConfigurationVersion;
- running Snapshot никогда не заменяется, не refresh, не patch и не нормализуется повторно;
- readers наблюдают только полный построенный Snapshot, а не частичный construction.

Publication consistency относится к выбранному декларативному source. Она не вводит механизм in-place reload.

## 20. Границы отказов

Ошибка загрузки или построения возникает до ownership Runtime Host и до получения Runtime resources. Такая ошибка:

- не публикует Snapshot в Runtime Services;
- не запускает Listener и не открывает admission;
- не оставляет частично composed Runtime graph;
- не изменяет исходную ConfigurationVersion;
- не выбирает fallback или более новую версию неявно;
- создаёт truthful failed Launch Attempt через lifecycle owner ARCH-004.

Ошибки Bootstrap после передачи Snapshot в Runtime регулируются startup transaction и rollback из ARCH-002. ARCH-005 не определяет retry, backoff, fallback или recovery policy.

## 21. Эквивалентность источников

Все conforming sources должны создавать семантически эквивалентные Snapshots для одинаковых identity ConfigurationVersion, schema, Runtime Instance и Launch Attempt.

Эквивалентность включает effective Runtime configuration, normalization, ordering, semantics absence-versus-presence, Secret References и provenance. Source-specific formatting, storage metadata, retrieval path и тип adapter не должны влиять на результат.

Tests и in-memory sources подчиняются той же границе. Они не должны получать privileged fields, ослабленную validation или alternate defaults, недоступные production sources.

## 22. Обязательные инварианты

1. Runtime запускается ровно из одного полного immutable Snapshot.
2. Каждый Snapshot происходит ровно из одной Published ConfigurationVersion, выбранной для нового Launch Attempt.
3. Версии Draft, Validated и Archived не могут стать source нового Snapshot.
4. Один Snapshot принадлежит ровно одному Runtime Instance и одному Launch Attempt.
5. Один Launch Attempt создаёт не более одного успешного Snapshot и одного Runtime Host.
6. Snapshot идентифицирует provenance Workspace, Configuration, точной ConfigurationVersion, Configuration schema, Runtime Instance и Launch Attempt.
7. Snapshot содержит только effective Runtime configuration и provenance.
8. Snapshot содержит Secret References и никогда не содержит Secret values.
9. Loader и Builder никогда не раскрывают Control Plane repositories или management API в Runtime.
10. Loader не выбирает Configuration, не принимает lifecycle decisions, не нормализует Runtime values и не строит Runtime-компоненты.
11. Builder не загружает, не сохраняет, не выбирает, не публикует и не изменяет ConfigurationVersion.
12. Успешное построение полностью отделяет Snapshot от caller-owned mutable memory.
13. Snapshot никогда не изменяется после построения и никогда не reload in place.
14. Runtime Services читают Snapshot и никогда не владеют, не заменяют и не изменяют его.
15. Публикация новой ConfigurationVersion не влияет на существующие Snapshot или Host.
16. Эквивалентные valid представления source создают семантически эквивалентные Snapshots.
17. Unsupported schema или malformed input приводят к отказу до получения Runtime resources и readiness.
18. Ни один observer не может получить доступ к частично построенному Snapshot.
19. Bootstrap и Host получают Snapshot без Loader, repository или publication authority.
20. Lifetime Snapshot ограничен его Host и разрешёнными readers и не владеет независимо уничтожаемым resource.
21. Тип source adapter, timestamp загрузки, PID, Host pointer и identity socket не становятся provenance Snapshot.
22. Retry, replacement и restart требуют нового Launch Attempt и нового Snapshot.

## 23. Вопросы для последующих Design Proposal

Focused Design Proposals должны определить implementation contracts для:

- commands, errors и adapter integration Configuration Loader;
- конкретного представления provenance Snapshot;
- compatibility ranges, negotiation и migration Configuration schema;
- persistence Runtime Instance и Launch Attempt;
- времени secret resolution и lifetime значений каждой consuming capability там, где текущие contracts расходятся с общей startup-моделью ADR-0003;
- operational diagnostics и redaction ошибок загрузки и построения;
- policies replacement, rollback, reconciliation и retry.

Эти вопросы не должны решаться через hidden behavior Loader или Builder.

## 24. Решение

UWP принимает Runtime Configuration Snapshot как единственный immutable Runtime input между Published ConfigurationVersion и Runtime Host.

Runtime Lifecycle Owner закрепляет одну точную Published ConfigurationVersion за одним Launch Attempt. Configuration Loader получает этот точный published source, не раскрывая persistence или management infrastructure. Builder defensively validates, normalizes, detaches и строит один Snapshot с effective Runtime configuration и обязательной declarative и operational provenance. Bootstrap передаёт полный Snapshot одному Host, который владеет им весь lifetime; Runtime Services являются read-only consumers.

Все источники Configuration должны быть семантически эквивалентны на этой границе. Snapshot никогда не содержит Secret values, никогда не изменяется in place и никогда не перенаправляется последующей публикацией. Ошибки загрузки, построения, несовместимости schema и validation предотвращают получение Runtime resources и не создают частичный Runtime input.

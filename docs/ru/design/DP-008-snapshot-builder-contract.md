# DP-008: Контракт Snapshot Builder

[English version](../../en/design/DP-008-snapshot-builder-contract.md)

## 1. Статус

**Статус:** Draft

**Статус реализации:** запланирован; текущий Builder всё ещё принимает
ConfigurationVersion напрямую и не строит полный provenance ARCH-005 или
blocking Diagnostics

**Статус архитектуры:** Контракт реализации утверждённой модели из
[ARCH-004](../architecture/ARCH-004-runtime-deployment-and-identity-model.md)
и
[ARCH-005](../architecture/ARCH-005-runtime-configuration-snapshot-and-loading-model.md)

Этот proposal не вводит и не пересматривает архитектуру. Он определяет
инженерный контракт, согласно которому Builder преобразует один полный
`DetachedLoadResult` в один полный immutable Runtime Snapshot либо возвращает
блокирующие Diagnostics без Snapshot.

## 2. Назначение

[DP-007](DP-007-configuration-loader-contract.md) определяет, как Runtime
Lifecycle Owner получает точную Published ConfigurationVersion, закреплённую
за одним Launch Attempt, и принимает detached, не зависящий от source
результат. ARCH-005 определяет Runtime Snapshot как единственный immutable
configuration input для Runtime Bootstrap и Runtime Host.

DP-008 определяет отсутствующий контракт Builder между этими границами:

```text
Configuration Loader
    -> Detached Load Result
    -> Builder
    -> immutable Runtime Snapshot
    -> Runtime Bootstrap
```

Proposal определяет одну operation Builder, границы её input и output,
semantic validation, normalization, Diagnostics, ownership, determinism,
правила dependencies, инварианты Runtime Snapshot и обязательные acceptance
proofs.

## 3. Источники архитектурных решений

Нормативной архитектурой остаются:

- [ADR-0002: Configuration DSL](../adr/0002-configuration-dsl.md);
- [ADR-0003: Runtime Architecture](../adr/0003-runtime-architecture.md);
- [ARCH-002: Runtime Foundation Freeze](../architecture/ARCH-002-runtime-foundation-freeze.md);
- [ARCH-004: Runtime Deployment and Identity Model](../architecture/ARCH-004-runtime-deployment-and-identity-model.md);
- [ARCH-005: Runtime Configuration Snapshot and Loading Model](../architecture/ARCH-005-runtime-configuration-snapshot-and-loading-model.md);
- [DP-007: Configuration Loader Contract](DP-007-configuration-loader-contract.md).

Если этот implementation proposal допускает более широкое толкование, чем
указанные документы, применяется более узкий утверждённый архитектурный
контракт.

## 4. Область

DP-008 определяет:

- одну operation Runtime Snapshot Builder;
- handoff от `DetachedLoadResult` к Builder;
- границу между полнотой representation и semantic completeness;
- полную semantic validation правил, принадлежащих Builder;
- детерминированную normalization в каноническую Runtime model;
- создание одного полного immutable Runtime Snapshot;
- сохранение declarative и operational provenance;
- создание блокирующих Diagnostics;
- взаимоисключающий контракт результата success-or-failure;
- правила ownership, detachment, immutability, atomicity и determinism;
- правила dependencies между Builder, контрактом Loader, Bootstrap и Runtime;
- инварианты Runtime Snapshot;
- acceptance proofs, обязательные до утверждения реализации.

DP-008 не определяет:

- загрузку Configuration, выбор source или source adapters;
- поведение Repository, PostgreSQL, HTTP или YAML;
- публикацию ConfigurationVersion или изменение её lifecycle;
- Runtime Lifecycle Owner, Runtime Launcher или management commands;
- реализацию Runtime Bootstrap или запуск Runtime;
- изменения lifecycle, readiness, rollback или shutdown Runtime Host;
- создание Runtime resources;
- разрешение Secret или lifetime значения Secret;
- persistence, caching, serialization или inspection API для Snapshot;
- hot reload, replacement, reconciliation, retry или fallback;
- Go interfaces, сигнатуры методов, конкретные structs, packages или
  exported APIs.

## 5. Термины контракта

### Detached Load Result

Detached Load Result — это полный, immutable за счёт ownership output Loader,
определённый в DP-007. Он содержит точный declarative payload, проверенный факт
Published, identity схемы и declarative и operational identities, необходимые
для provenance Snapshot.

Builder заимствует это значение на время одной operation Build. Он не получает
полномочий Loader, Source, Repository, publication или lifecycle.

### Runtime Snapshot

Runtime Snapshot — это полное, immutable, detached, готовое к исполнению
значение, определённое ARCH-005 для одного Runtime Host и одного Launch Attempt.

Runtime Snapshot является Runtime model. Он не является ConfigurationVersion,
Configuration, Workspace, сущностью Repository, Control Plane DTO или
обёрткой над любым из этих объектов.

### Semantic Validation

Semantic Validation определяет, удовлетворяют ли detached declarative values и
provenance поддерживаемой схеме, domain rules, cross-field rules и инвариантам
полного Runtime Snapshot, принадлежащим Builder.

Semantic Validation не повторяет доступ к source, выбор публикации,
representation decoding или validation ресурсов при запуске.

### Normalization

Normalization — это детерминированное преобразование семантически корректных
declarative values в одно каноническое Runtime representation. Оно может
вычислять чистые derived values, необходимые Runtime model, но не создаёт
Runtime resources или mutable operational state.

### Diagnostic

Diagnostic — одно блокирующее semantic violation, обнаруженное Builder. Оно
структурировано для последующего представления через CLI, Web UI или API без
необходимости для этих consumers заново интерпретировать Configuration domain.

### Semantic Equivalence

Два Detached Load Result семантически эквивалентны для Builder, когда они
удовлетворяют Semantic Equivalence из DP-007 и содержат равные declarative
values, schema facts, факт Published и provenance identities, относящиеся к
созданию Snapshot. Это определение применяется независимо от того, являются ли
эти значения семантически корректными для создания Snapshot.

Два Runtime Snapshot семантически эквивалентны, когда равны их effective
Runtime configuration, canonical ordering, семантика
absence-versus-presence, Secret References, чистые derived structures и полный
provenance. Allocation layout и адрес объекта не влияют на эквивалентность.

Два Diagnostics Set семантически эквивалентны, когда они содержат одинаковые
блокирующие semantic violations, определённые одинаковыми rule identities и
logical locations. Presentation ordering, allocation layout и адрес объекта
не влияют на эквивалентность Diagnostics Set.

## 6. Обзор контракта

Наблюдаемая успешная operation:

```text
получить полный Detached Load Result
    -> проверить handoff, schema, semantic и normalization obligations
    -> создать полную Runtime model и provenance
    -> отделить все значения, принадлежащие Snapshot
    -> вернуть immutable Runtime Snapshot
```

Наблюдаемая неуспешная operation:

```text
получить Detached Load Result
    -> проверить каждое применимое правило, принадлежащее Builder
    -> собрать детерминированный набор блокирующих Diagnostics
    -> вернуть Diagnostics
    -> не публиковать Runtime Snapshot
```

У Builder есть ровно два результата:

1. один полный Runtime Snapshot без Diagnostics; либо
2. одна непустая коллекция Diagnostics без Runtime Snapshot.

Runtime Snapshot вместе с Diagnostics, partial Snapshot, recoverable Snapshot
и Snapshot с отложенной semantic validation запрещены.

Эти flows определяют наблюдаемые обязательства и результаты, а не внутренний
алгоритм. Builder может чередовать, повторять или совместно использовать любые
чистые промежуточные вычисления, необходимые для semantic validation,
normalization, Diagnostics или создания Snapshot. Ни одно промежуточное
вычисление не становится наблюдаемым Snapshot, Diagnostic или Runtime resource.

## 7. Operation Builder

Единственная архитектурная operation — **Build Runtime Snapshot**.

У operation есть следующие наблюдаемые обязанности:

- получить один полный `DetachedLoadResult`;
- защитно проверить факты handoff, требуемые DP-007 и ARCH-005;
- проверить, что Configuration schema поддерживается этим Builder;
- проверить каждое применимое semantic, cross-field и normalization rule;
- собрать все применимые блокирующие Diagnostics, не останавливаясь после
  первого независимого violation;
- вернуть полный Diagnostics Set без Snapshot при наличии любого Diagnostic;
- иначе создать одну каноническую Runtime model, включая допустимые чистые
  derived structures и полный provenance;
- отделить всё логическое содержимое Snapshot от mutable memory input;
- опубликовать один полный immutable Runtime Snapshot.

Этот список не предписывает порядок фаз или внутренний control flow. Builder
может вычислять предварительные канонические значения или чистые derived
structures тогда, когда они нужны для определения semantic validity. Такие
вычисления локальны для operation и не образуют partial Snapshot.

Успешный return является linearization point создания Snapshot. До него
Snapshot архитектурно невидим. После него полный Snapshot immutable, а Builder
не сохраняет ни input, ни output.

Operation не:

- загружает и не выбирает Configuration;
- вызывает Loader, Source, Repository или management API;
- изменяет ConfigurationVersion или `DetachedLoadResult`;
- принимает решения о launch, retry, replacement или lifecycle;
- вызывает Bootstrap и не создаёт Runtime components;
- разрешает значения Secret;
- получает sockets или иные Runtime resources;
- запускает goroutines, callbacks или background work.

## 8. Input Builder

Builder получает ровно один полный `DetachedLoadResult` через neutral handoff
contract, определённый после DP-007.

Input предоставляет:

- полный declarative payload точной ConfigurationVersion;
- identity Workspace;
- identity Configuration;
- точную identity и номер ConfigurationVersion;
- факт Published, наблюдавшийся Loader;
- identity и версию Configuration schema;
- identity Runtime Instance;
- identity Launch Attempt.

Builder должен защитно проверить факты input, необходимые ему для успешного
создания, включая:

- наличие и внутреннюю согласованность обязательного provenance;
- факт Published, переданный через handoff;
- поддерживаемую Configuration schema;
- полноту, необходимую для semantic validation и создания Snapshot.

Builder не перечитывает и не устанавливает заново:

- существование source;
- repository ownership между Workspace и Configuration;
- Consistent Observation Loader;
- остаётся ли эта version текущей Published version;
- persistence state Runtime Instance или Launch Attempt.

Гарантия Loader не разрешает опускать защитную handoff validation. Защитная
validation не должна превращать Builder во второй Loader.

## 9. Output Runtime Snapshot

При успехе Builder возвращает один Runtime Snapshot, содержащий ровно
архитектурные категории, утверждённые ARCH-005:

1. полную effective behavior-affecting Runtime configuration для
   поддерживаемого component graph; и
2. полный стабильный declarative и operational provenance.

Runtime Snapshot должен быть:

- полным;
- immutable после успешного создания;
- структурно независимым от Configuration domain model;
- detached от всей mutable memory, принадлежащей input;
- каноническим для семантически эквивалентного input;
- достаточным для Bootstrap и Runtime Services без доступа к source;
- свободным от независимо принадлежащих Runtime resources.

Runtime Snapshot не должен содержать:

- объекты ConfigurationVersion, Configuration или Workspace;
- сущности Repository, Source adapters, Loader или management services;
- HTTP, YAML или persistence DTO;
- историю ConfigurationVersion или полномочия publication;
- значения Secret или Secret Resolver;
- Runtime components, callbacks, contexts, goroutines, channels, locks,
  sockets, timers или mutable operational state;
- desired или actual lifecycle state;
- identity Snapshot, identity Build, source metadata или telemetry.

Конкретная структура полей Runtime Snapshot остаётся вопросом implementation
design. Она не должна ослаблять инварианты этого proposal.

## 10. Контракт Provenance

У Runtime Snapshot нет независимой business identity. Builder не должен
создавать Snapshot ID, Build ID или другую identity.

Snapshot сохраняет уже установленное замыкание identities для выбранного
launch:

- identity Workspace;
- identity Configuration;
- точную identity и номер ConfigurationVersion;
- identity и версию Configuration schema;
- identity Runtime Instance;
- identity Launch Attempt.

Provenance содержит значения, а не Control Plane entities. Он не содержит тип
source adapter, время load, время build, duration, PID, указатель Host, socket
address, process identity или другую telemetry.

Builder может проверять и копировать provenance. Он не выделяет, не заменяет и
не переинтерпретирует identities, переданные через handoff.

## 11. Граница Semantic Validation

Builder является авторитетной границей semantic validation для создания
Runtime Snapshot.

Builder должен проверять:

- поддерживаемые identity и version Configuration schema;
- полноту, требуемую поддерживаемой Runtime model;
- domain semantics поддерживаемых behavior-affecting sections;
- cross-field и cross-section инварианты Runtime Snapshot;
- правила uniqueness, ordering, reference и absence-versus-presence,
  принадлежащие поддерживаемой Configuration semantics;
- возможность создать каждое обязательное каноническое Runtime value и чистую
  derived structure;
- сохранение и внутреннюю согласованность обязательного provenance;
- факт Published в handoff, требуемый DP-007.

Builder должен проверить каждое применимое semantic rule и собрать все
обнаруженные блокирующие violations. Он не должен останавливаться на первом
независимом violation.

Semantic validation может использовать любую чистую operation-local
normalization или derived calculation, необходимую для проверки rule. Различие
между validation и normalization определяет ответственность и наблюдаемые
результаты; оно не требует отдельных внутренних фаз.

Builder не должен проверять:

- доступность source или storage representation;
- repository ownership посредством другого чтения source;
- историю publication или выбор current version;
- persistence state Runtime Instance или Launch Attempt;
- management authorization, desired state или launch eligibility;
- существование Secret или значения Secret;
- возможность создать Listener, TLS, Authentication или другой Runtime
  resource в текущем environment;
- Runtime lifecycle, readiness, rollback или shutdown.

Control Plane validation и Loader validation не отменяют защитную semantic
ответственность Builder. Acceptance semantics Builder не должны противоречить
утверждённому Configuration domain.

`Host.Start()` сохраняет startup-critical capability validation, требуемую
ARCH-002 и ARCH-005. Эта validation относится к исполняемой полностью
собранной Runtime configuration и созданию operational resources, а не к
semantic validity Configuration. Bootstrap может проверять только свой static
construction input.

## 12. Normalization и чистые Derived Structures

Builder является единственным authority normalization на границе Runtime.

Normalization должна:

- быть чистой функцией полного `DetachedLoadResult`;
- сохранять утверждённую Configuration semantics;
- создавать канонические Runtime values;
- сохранять значимый ordering и различия absence-versus-presence;
- устанавливать детерминированный ordering там, где Configuration semantics
  определяет порядок независимо от source representation;
- разрешать declarative references только тогда, когда resolution является
  чистой operation над detached input;
- вычислять lookup tables, indices, inheritance results и другие derived
  structures только тогда, когда они являются immutable values, полностью
  определяемыми input;
- оставлять input неизменным.

В рамках normalization Builder не должен создавать:

- goroutines;
- channels;
- mutexes или другие synchronization primitives;
- sockets или network clients;
- timers;
- объекты TLS configuration;
- compiled regular expressions;
- JWT validators;
- Authentication Providers;
- экземпляры Router;
- Listener, Session Manager или другие Runtime objects.

Эти объекты принадлежат `Host.Start()` или специализированному Runtime
component, создание которого координирует Host.

Bootstrap, запуск Host и Runtime Services не должны повторять semantic
normalization, создавать source-specific defaults или переинтерпретировать
корректный Snapshot. Bootstrap может только проверять static construction input
и связывать dependencies. `Host.Start()` выполняет startup-critical checks и
создаёт Runtime resources.

## 13. Контракт Diagnostics

Diagnostics Builder содержат только блокирующие semantic violations. Warning и
informational diagnostics находятся вне этого proposal.

Каждый Diagnostic должен:

- определять одно нарушенное semantic rule в machine-readable виде;
- определять затронутое logical configuration location, если violation
  относится к конкретному location;
- предоставлять presentation text, подходящий для человека-оператора;
- не содержать значения Secret, объекта source, Runtime resource или mutable
  authority;
- оставаться осмысленным без доступа к Repository, Loader, Bootstrap или
  Runtime.

Коллекция Diagnostics должна:

- содержать каждое блокирующее violation, обнаруживаемое проверкой всех
  применимых semantic rules относительно переданного detached input;
- не содержать Snapshot или partial Runtime model;
- не зависеть от traversal timing, source adapter и allocation layout;
- подходить для детерминированного rendering слоями CLI, Web UI и API;
- оставаться detached от mutable memory, принадлежащей input.

### Применимость Diagnostics

Semantic rule **применимо**, когда:

- rule принадлежит поддерживаемой schema и проверяемой configured section;
- каждый prerequisite fact, необходимый для проверки rule, присутствует и
  семантически пригоден; и
- проверка rule не требует предположений о факте, уже отклонённом другим
  применимым prerequisite rule.

**Prerequisite rule** устанавливает факт, необходимый одному или нескольким
dependent rules. Нарушение prerequisite rule:

- создаёт собственный блокирующий Diagnostic;
- делает неприменимыми только rules, которым необходим отклонённый факт;
- не подавляет независимые rules, чьи prerequisites остаются выполненными.

Builder должен подавлять cascading Diagnostic, когда тот сообщает только о
следствии уже сообщённого нарушения prerequisite и не может определить
дополнительное независимо проверяемое violation. Builder не должен использовать
cascade suppression для пропуска независимого violation.

Для одного input и одного утверждённого набора semantic rules Builder должен
вернуть один детерминированный Diagnostics Set:

- каждое нарушенное применимое rule добавляет один Diagnostic для каждого
  отдельного logical location, где существует это violation;
- одинаковые identity rule и logical location встречаются не более одного
  раза;
- семантически эквивалентные некорректные inputs создают семантически
  эквивалентные Diagnostics Sets;
- traversal order, allocation и presentation ordering не изменяют состав Set;
- rules, ставшие неприменимыми из-за нарушенных prerequisites, не добавляют
  Diagnostic.

Builder не должен логировать Diagnostics, публиковать их во внешнюю систему,
записывать failure Launch Attempt или выбирать management response. После
возврата Builder ответственность за результат failed launch остаётся у Runtime
Lifecycle Owner.

Точное namespace кодов Diagnostic, грамматика logical-location, canonical
ordering, localization policy и представление redaction требуют
implementation contract до начала разработки.

**TODO:** Определить эти детали representation Diagnostics, не изменяя
установленный здесь semantic contract только блокирующих нарушений.

## 14. Атомарность результата и семантика Failure

Builder публикует ровно одно из следующего:

- один полный immutable Runtime Snapshot; либо
- одну непустую полную коллекцию Diagnostics.

Запрещены следующие состояния:

- одновременный возврат Snapshot и Diagnostics;
- partial Snapshot;
- recoverable или degraded Snapshot;
- Snapshot с некорректными semantic sections, которые должны быть проверены
  при запуске;
- Snapshot, завершаемый асинхронно после return;
- success с последующим failure, принадлежащим Builder.

При failure Builder:

- Snapshot не возвращается и не публикуется;
- Builder не вызывает Bootstrap или Runtime Host;
- operation Build не создаёт Runtime resources;
- `DetachedLoadResult` остаётся неизменным;
- fallback, retry, replacement или downgrade schema не выполняются;
- Builder не сохраняет input или mutable reference, принадлежащую Diagnostics;
- ответственность за правдивую запись failed Launch Attempt остаётся у Runtime
  Lifecycle Owner.

Публикация Diagnostics caller является linearization point failure. Ни один
observer не может получить частично проверенный или частично нормализованный
Snapshot.

## 15. Инварианты Snapshot

### Инварианты Identity

- У Runtime Snapshot нет независимой identity.
- Snapshot сохраняет identities Workspace, Configuration, точной
  ConfigurationVersion, Runtime Instance и Launch Attempt.
- Snapshot сохраняет номер ConfigurationVersion, а также identity и version
  Configuration schema.
- Builder не создаёт, не заменяет и не перепривязывает ни одну из этих
  identities.

### Структурные инварианты

- Snapshot содержит Runtime model, а не объекты Configuration domain.
- Snapshot содержит только effective Runtime configuration и provenance.
- Каждая поддерживаемая Runtime section, требуемая для исполнения, структурно
  полна.
- Из Snapshot недостижимы Repository, Source, Loader, DTO, history или mutable
  authority.

### Семантические инварианты

- Каждое значение Snapshot удовлетворяет поддерживаемым правилам Configuration
  и cross-field rules, принадлежащим Builder.
- Unsupported schema и семантически некорректный input не могут создать
  Snapshot.
- Secret References могут сохраняться; значения Secret не могут попадать в
  Snapshot.
- Bootstrap и Host не получают неразрешённых semantic violations.

### Инварианты Normalization

- Snapshot использует одно каноническое representation для семантически
  эквивалентного input.
- Значимый ordering и различия absence-versus-presence сохраняются.
- Derived structures immutable и определяются исключительно input.
- Никакие source-specific или скрытые Runtime defaults не изменяют
  Configuration semantics.

### Инварианты Ownership

- Builder заимствует input и после return не сохраняет ссылок.
- Успешный Snapshot не имеет mutable alias на логическое содержимое input.
- После return Builder не сохраняет ownership Snapshot.
- Readers Snapshot не могут изменить logical view другого reader.

### Независимость Runtime

- Создание Snapshot не получает Runtime resources.
- Snapshot не содержит Runtime object или mutable operational state.
- Для интерпретации Snapshot Runtime Services никогда не нуждаются в Loader,
  Source, Repository или истории Configuration.

### Детерминизм

- Семантически эквивалентные корректные inputs создают семантически
  эквивалентные Snapshots.
- Повторные operations Build не зависят от traversal timing, allocation,
  process state, типа source или внешнего mutable state.

### Полнота

- Каждое поддерживаемое behavior-affecting значение input либо представлено в
  Snapshot согласно утверждённой semantics, либо вызывает блокирующие
  Diagnostics.
- Snapshot достаточен для Bootstrap и Host без дополнительного чтения
  declarative source.
- Ни одно поддерживаемое semantic obligation не откладывается до Runtime
  Services.

### Атомарность

- Snapshot становится видимым только после полного успешного создания.
- Failure публикует Diagnostics и не публикует Snapshot.
- Ни один observer не может получить промежуточное состояние validation,
  normalization или construction.

## 16. Ownership, Immutability и Lifetime

Ownership передаётся в одном направлении:

| Этап | Ownership |
| --- | --- |
| Runtime Lifecycle Owner | Владеет подготовкой launch и Detached Load Result |
| Builder | Заимствует input и владеет только синхронной operation Build |
| Runtime Lifecycle Owner после success | Владеет полным Runtime Snapshot |
| Runtime Launcher / Bootstrap | Заимствует или принимает Snapshot согласно передаче ownership из ARCH-005 |
| Runtime Host | Владеет своим независимым immutable Snapshot в течение lifetime Host |
| Runtime Services | Только чтение |

Builder не должен сохранять:

- `DetachedLoadResult`;
- любую mutable reference, принадлежащую input;
- Diagnostics после возврата failure;
- Runtime Snapshot после возврата success;
- cache или registry предыдущих operations Build.

Snapshot становится immutable при успешном создании. Lifetime его значения
продолжается согласно ARCH-005 через подготовку launch, Bootstrap, Host и
разрешённых Runtime readers. У него нет cleanup protocol, поскольку он не
содержит значений Secret или независимо принадлежащих Runtime resources.

## 17. Детерминизм и Concurrency

Output Builder должен зависеть только от полного input и утверждённых semantic
rules.

Правила determinism:

- семантически эквивалентные корректные inputs создают семантически
  эквивалентные Snapshots;
- повторные builds эквивалентных inputs не накапливают state;
- изменения publication или source после success Loader не могут повлиять на
  Build;
- Builder не читает clock, random source, environment, global registry,
  Repository, network service или Runtime state;
- чистые derived structures вычисляются одинаково для эквивалентного input;
- ordering input сохраняется или канонизируется только согласно требованиям
  утверждённой Configuration semantics.

Builder не содержит mutable operation state, переживающего Build. Независимые
operations Build не разделяют частично созданные состояния Snapshot или
Diagnostics. Контракт не вводит synchronization object или background
execution.

Точные concurrency API и семантика ожидания caller находятся вне этого
proposal. Никакая реализация не может ослаблять детерминированную изоляцию
между независимыми operations Build.

## 18. Правила Dependencies

Требуемое направление dependencies:

```text
Runtime Lifecycle Owner
    -> neutral Loader handoff contract
        -> Builder
            -> immutable Runtime Snapshot

Runtime Lifecycle Owner
    -> Runtime Launcher
        -> Runtime Bootstrap
            -> immutable Runtime Snapshot
            -> Host.Start()
```

Builder использует только neutral handoff contract Loader. Он не импортирует
конкретную реализацию Configuration Loader и не зависит от неё. Текущий
repository реализует этот neutral contract в `internal/runtimeconfigload`;
этот path является фактом реализации, а не нормативным требованием к package
layout из DP-008.

Обязательные правила:

- Builder не знает реализацию Loader или Source adapters.
- Builder не знает Repository, PostgreSQL, HTTP, YAML или Configuration
  services.
- Builder не знает Runtime Bootstrap, Runtime Host, Listener, Session Manager,
  реализацию Router или Runtime Services.
- Builder не знает Secret Resolver.
- Loader не импортирует и не вызывает Builder.
- Runtime Bootstrap принимает Runtime Snapshot, связывает construction inputs,
  создаёт Host и вызывает `Host.Start()`; он не получает
  `DetachedLoadResult`.
- Runtime Host и Runtime Services не получают полномочий Builder или Loader.
- Никакой dependency cycle не может соединять создание Runtime Snapshot
  обратно с repositories Control Plane.

Этот proposal определяет ответственности dependencies, а не package layout.

## 19. Последовательность взаимодействия

```mermaid
sequenceDiagram
    participant O as Runtime Lifecycle Owner
    participant L as Configuration Loader
    participant B as Builder
    participant X as Runtime Launcher
    participant R as Runtime Bootstrap
    participant H as Runtime Host
    L-->>O: Complete Detached Load Result
    O->>B: Build from Detached Load Result
    B->>B: Evaluate applicable semantics and canonical form
    alt Diagnostics exist
        B-->>O: Complete blocking Diagnostics, no Snapshot
        O->>O: Record failed Launch Attempt
    else Input is valid
        B->>B: Normalize and construct complete Snapshot
        B-->>O: Immutable Runtime Snapshot
        O->>X: Launch with Snapshot
        X->>R: Invoke Bootstrap
        R->>H: Bind inputs, create Host, invoke Host.Start()
        H-->>R: Active Host or startup failure after rollback
        R-->>X: Launch outcome
        X-->>O: Launch outcome
    end
```

Builder не вызывает ни Loader, ни Bootstrap. Runtime Lifecycle Owner сохраняет
утверждённый pipeline и владеет всеми management outcomes.

## 20. Acceptance Proofs

Следующие свойства реализации обязательны. Они не являются именами тестов и не
предписывают конкретную технику тестирования.

### AP-001: Детерминизм Builder

Семантически эквивалентные корректные значения `DetachedLoadResult` создают
семантически эквивалентные Runtime Snapshots независимо от allocation, порядка
вызовов, source adapter и случайного process-local state.

### AP-002: Независимость Runtime Model

Из Runtime Snapshot недостижимы ConfigurationVersion, Configuration,
Workspace, сущность Repository, Control Plane DTO или source-specific
representation.

### AP-003: Граница Semantic Validation

Каждое поддерживаемое semantic и cross-field rule, принадлежащее Builder,
проверяется до success, а доступ к source, разрешение Secret, создание ресурсов
и lifecycle decisions остаются вне Builder.

### AP-004: Полнота Snapshot

Каждый успешный Runtime Snapshot содержит всю поддерживаемую effective Runtime
configuration и обязательный provenance, требуемые Bootstrap без
дополнительного чтения declarative source.

### AP-005: Отсутствие Partial Snapshot

Любое блокирующее semantic violation не создаёт Runtime Snapshot, и ни один
observer не может получить промежуточное или recoverable состояние Snapshot.

### AP-006: Полнота Diagnostics

Для некорректного input возвращённые Diagnostics определяют каждое блокирующее
violation каждого применимого semantic rule, принадлежащего Builder, подавляют
только cascading consequences нарушенных prerequisite rules и содержат каждое
независимое violation ровно один раз для каждого отдельного logical location.

### AP-007: Immutability Snapshot

Никакое изменение caller-owned input после Build не может изменить успешный
Snapshot, и ни один reader Snapshot не может изменить logical view другого
reader.

### AP-008: Изоляция Ownership

После return Builder не сохраняет ни input, ни Diagnostics, ни Snapshot, и
независимые результаты Build не разделяют mutable logical content.

### AP-009: Полнота Normalization

Каждый успешный Snapshot представляет весь поддерживаемый input в одной
канонической Runtime form, сохраняя значимый order и absence semantics и не
применяя source-specific или скрытые defaults.

### AP-010: Независимость Runtime

Один только success Build не создаёт Runtime component, goroutine, channel,
lock, socket, timer, TLS configuration, validator, context, readiness state
или другой Runtime resource.

### AP-011: Ответственность за Startup

Builder возвращает только семантически полные значения Snapshot; Bootstrap
выполняет только validation static construction-input и связывание
dependencies, а `Host.Start()` выполняет startup-critical capability
validation и создание ресурсов. Ни один из них не повторяет semantic
validation или normalization Configuration.

### AP-012: Сохранение Identity

Каждый успешный Snapshot сохраняет полный declarative и operational provenance
из `DetachedLoadResult`, не создавая identity Snapshot или Build.

### AP-013: Semantic Equivalence

Эквивалентные source representations, удовлетворяющие Semantic Equivalence из
DP-007, создают семантически эквивалентные значения Snapshot либо семантически
эквивалентные детерминированные Diagnostics Sets для одинаковой invalid
semantics.

### AP-014: Эквивалентность повторных Build

Повторные успешные Builds семантически эквивалентного input создают
семантически эквивалентные, независимо принадлежащие Snapshots, не изменяют
input и не сохраняют operation state.

### AP-015: Полномочия Snapshot

Builder является единственным authority создания Runtime Snapshot; Loader,
Bootstrap, `Host.Start()` и Runtime Services не дополняют, не нормализуют, не
исправляют и не изменяют Snapshot.

## 21. Совместимость с архитектурой

### ADR-0002

ConfigurationVersion остаётся единственным declarative source поведения
Runtime. Builder создаёт detached Runtime representation, а не второй язык
Configuration, и не вводит скрытый behavior-affecting input.

### ADR-0003

Runtime продолжает использовать один immutable Configuration Snapshot через
явную composition. Builder не получает dependencies HTTP или Repository и не
создаёт Runtime service.

### ARCH-002

DP-008 заканчивается до Runtime Bootstrap. Он не изменяет принадлежащие Host
composition, startup transaction, readiness, Admission Gate, rollback,
shutdown или lifecycle.

### ARCH-004

Builder не создаёт Runtime Instance, Launch Attempt, desired state, actual
state или management decision. Он сохраняет identities, уже принадлежащие
Runtime Lifecycle Owner.

### ARCH-005

Builder остаётся единственной границей semantic validation, normalization,
detachment и создания Snapshot. Snapshot остаётся immutable Runtime input,
принадлежащим согласно утверждённому однонаправленному lifetime.

### DP-007

Builder использует полный neutral `DetachedLoadResult`, повторяет только
назначенные ему защитные handoff checks, не обращается к source и возвращает
либо полный Snapshot, либо принадлежащие Builder Diagnostics.

## 22. Намеренно отложенные вопросы

Следующие вопросы остаются вне DP-008:

- диапазоны compatibility, negotiation и migration Configuration schema;
- реализация Runtime Bootstrap и Runtime Launcher;
- контракт Runtime startup и startup error;
- момент разрешения Secret и lifetime значения Secret;
- persistence, cache, serialization, inspection и transport Snapshot;
- hot reload, replacement, reconciliation, retry, fallback и recovery;
- Warning и informational Diagnostics;
- operational logging, metrics и хранение diagnostics;
- представление результатов Build в management API.

Ничто из перечисленного не может быть реализовано как скрытое поведение
Builder.

## 23. Предварительные условия реализации

Утверждённая архитектура пока не определяет:

1. точные identity поддерживаемой Configuration schema и compatibility rule;
2. полную структуру полей Runtime Snapshot и section-specific правила
   normalization;
3. точное structured representation Diagnostics и canonical ordering.

**TODO:** Определить эти три implementation contracts посредством
сфокусированного review до начала разработки Builder. Их определение должно
уточнить этот proposal, не изменяя модель ownership, validation, atomicity или
Runtime independence.

## 24. Решение

UWP создаёт Runtime Snapshot посредством одной operation Builder над neutral
`DetachedLoadResult`.

Builder является единственным authority semantic validation, normalization и
создания Runtime Snapshot. Он проверяет все применимые semantic violations,
возвращает либо один полный immutable Runtime Snapshot, либо полные блокирующие
Diagnostics, сохраняет обязательный provenance без создания новой identity и
не создаёт Runtime resources.

Runtime Snapshot — detached Runtime model, а не ConfigurationVersion или
Control Plane entity. Он содержит только полную effective Runtime configuration
и provenance, остаётся immutable после успешного создания и детерминирован для
семантически эквивалентного input. Bootstrap получает только семантически
полный Snapshot, проверяет static construction input, связывает dependencies,
создаёт Host и вызывает `Host.Start()`. Только Host владеет startup-critical
capability validation, operational composition, созданием ресурсов, rollback
и окончательным результатом startup.

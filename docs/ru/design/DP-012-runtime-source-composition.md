# DP-012: Композиция Runtime Source

[English version](../../en/design/DP-012-runtime-source-composition.md)

## 1. Статус

- **Design Status:** Draft
- **Implementation Status:** Planned

До утверждения proposal не является нормативным. Он определяет планируемый
repository-backed `configurationloader.Source`, но не утверждает наличие
adapter, management composition или Production Activation.

## 2. Назначение

Определить минимальную concrete Source composition, адаптирующую существующие
in-memory repositories Configuration и ConfigurationVersion к
`configurationloader.Source`, чтобы будущая application composition root могла
сконструировать `Source -> Loader -> Flow` без запуска Runtime.

## 3. Источники полномочий

Proposal уточняет, но не переопределяет:

- [ADR-0002](../adr/0002-configuration-dsl.md);
- [ADR-0003](../adr/0003-runtime-architecture.md);
- [ARCH-004](../architecture/ARCH-004-runtime-deployment-and-identity-model.md);
- [ARCH-005](../architecture/ARCH-005-runtime-configuration-snapshot-and-loading-model.md);
- [DP-007](DP-007-configuration-loader-contract.md) —
  [DP-011](DP-011-runtime-launch-pipeline-integration.md).

Approved ADR и Active architecture сохраняют приоритет.

## 4. Область

Design охватывает один in-process adapter поверх существующих concrete
in-memory repositories, exact lookup, consistency confinement, schema facts,
detachment, error classification, construction, dependency direction,
concurrency, lifetime и будущие implementation proofs.

## 5. Package и ответственность

Планируемый package — `internal/configurationloadsource`.

Он владеет:

- exact source addressing и confined consistency proof;
- статическими source representation facts;
- defensive detachment;
- классификацией repository outcomes в существующие Loader source errors.

Он не владеет validation Loader, semantic validation Builder, orchestration
Flow, lifecycle, routing или activation.

## 6. Точный планируемый API

```go
type MemorySource struct {
    // private borrowed dependencies
}

func NewMemorySource(
    configurations *configuration.MemoryConfigurationRepository,
    versions *configurationversion.MemoryConfigurationVersionRepository,
) *MemorySource

func (s *MemorySource) LoadExact(
    workspaceID uint64,
    configurationID uint64,
    configurationVersionID uint64,
) (configurationloader.SourceObservation, error)
```

`*MemorySource` реализует существующий `configurationloader.Source`.
Новые error, interface, option, context или lifecycle declaration не
добавляются.

## 7. Constructor

`NewMemorySource` только сохраняет borrowed repository references. Он не
выполняет read, write, создание goroutine, захват lock, validation,
инициализацию cache или lifecycle action. Nil receiver или любая nil dependency
приводит к `configurationloader.ErrSourceUnavailable` из `LoadExact`.

## 8. Exact lookup algorithm

Для одного вызова Source после проверки binding должен:

1. ровно один раз вызвать `versions.Get(configurationVersionID)`;
2. нормализовать возвращённую ошибку;
3. проверить возвращённые Version ID и Configuration ID;
4. отклонить известное не-Published state, неизвестное state или нулевой Number;
5. сохранить detached exact Version value;
6. ровно один раз вызвать `configurations.Get(configurationID)`;
7. нормализовать возвращённую ошибку;
8. проверить возвращённые Configuration ID и Workspace ID;
9. вернуть новый полный observation.

Запрошенный Workspace ID не используется для выбора другой entity; это
identity assertion для exact parent.

## 9. Запрещённый выбор

Source не вызывает `GetPublished`, не перечисляет versions, не выбирает
latest/current, не повторяет read, не заменяет запрошенную identity и не
делает fallback к другой ConfigurationVersion. Failure exact chain является
terminal для этого load.

## 10. Published validation

`Draft`, `Validated` и `Archived` отображаются в
`configurationloader.ErrVersionNotPublished`. Неизвестное lifecycle state или
нулевой Version Number отображаются в
`configurationloader.ErrSourceIntegrity`. Допустима только exact observed
`Published` Version.

## 11. Consistency linearization

Успешный Version `Get` является linearization point **L**. Последующий
Configuration `Get` родителя, **C**, доказывает существование родителя в L
только при single-instance mutation topology и immutability invariants из
следующего раздела. Единственный Version Service сериализует lifecycle
operations вокруг L; единственный Configuration Service не может изменить
parent identity или воскресить удалённого parent. Два независимых repository
lock сами по себе не создают cross-repository snapshot.

## 12. Обязательный composition confinement

Одна application composition root должна владеть одной repository pair и
сконструировать поверх неё ровно один `*configuration.Service` и ровно один
`*configurationversion.Service`. Эти два instances являются единственными
mutation authorities. Handlers получают только эти exact Service instances;
repository references не раскрываются.

После bootstrap/setup запрещены прямые repository
`Create`/`Update`/`UpdateBatch`/`Delete`, второй Service instance, importer,
migration или alternate writer. MemorySource выполняет только `Get`.

Единственный ConfigurationVersion Service использует один `lifecycleMu` для
сериализации всех Create, update, publish и archive operations. Это исключает
перезапись Published value устаревшим Draft update. Единственный Configuration
Service не имеет дополнительного mutex, но его Update меняет только Name,
Description и UpdatedAt: ID и Workspace ID остаются immutable. Update,
конкурирующий с Delete, не может воскресить parent, потому что repository
Update после удаления возвращает Not Found.

Composition обязана сохранять все факты:

- IDs монотонны и не переиспользуются;
- parent Configuration создаётся до своих Versions;
- Configuration ID и Workspace ID immutable;
- Version ID, Configuration ID, Number и Published payload immutable;
- каждый record `Get` и lifecycle transition atomic;
- каждый `Get` возвращает detached value.

Эти факты устанавливает обязательный Composition Audit до construction
MemorySource и повторно до Production Activation. Constructor не может
интроспектировать или доказать topology. Если Audit не доказывает topology или
обнаруживает второй Service/direct writer, Source не конструируется и
activation блокируется. Нарушение, добавленное после construction, является
programming/composition contract violation, а не обнаруживаемым runtime Source
failure. Ни два lock, ни retry не могут создать требуемый snapshot.

## 13. Parent races

Если parent удалён до C, Source консервативно возвращает Not Found. Если
удаление произошло после C, success остаётся truthful observation в L при
confinement.

## 14. Version lifecycle races

Archive до L возвращает Version Not Published. Archive после L не меняет уже
observed в L detached exact Published Version, поэтому load может завершиться
успешно при confinement.

## 15. Source observation

Success возвращает новый `configurationloader.SourceObservation`:

| Поле | Значение |
|---|---|
| `WorkspaceID` | запрошенный и проверенный Workspace ID |
| `Configuration` | detached exact parent Configuration |
| `ConfigurationVersion` | detached exact Published Version |
| `SchemaIdentity` | literal `uwp.configuration` |
| `SchemaVersion` | literal `1` |
| `RepresentationComplete` | `true` |

Schema facts являются adapter-owned static representation facts. Adapter не
импортирует `runtimeconfig`, не согласовывает schema и не выполняет semantic
validation.

## 16. Detachment

Observation не разделяет mutable logical content ни с repository, ни с
caller. Defense-in-depth копирует все nested collections Authentication
providers, pointers API Key/Basic/JWT, JWT slices, Routing, Routes и Matchers.
Изменение repository input после вызова не меняет observation, а изменение
возвращённого observation не меняет repository state или следующий
observation.

## 17. Failure matrix

| Условие | Точная возвращаемая identity |
|---|---|
| nil receiver или dependency | `ErrSourceUnavailable` |
| Version `Get` not found | `ErrSourceNotFound` |
| Configuration `Get` not found | `ErrSourceNotFound` |
| любая другая ошибка `Get` | `ErrSourceUnavailable` |
| mismatch возвращённых Version ID или Configuration ID | `ErrIdentityMismatch` |
| mismatch parent ID или Workspace ID | `ErrIdentityMismatch` |
| exact Version в `Draft`, `Validated` или `Archived` | `ErrVersionNotPublished` |
| неизвестное state или нулевой Version Number | `ErrSourceIntegrity` |
| неполные static schema или representation facts | `ErrSourceIntegrity` |

Mapping использует `errors.Is`. Source не раскрывает raw repository error,
wrapped detail, configuration data или transport diagnostic.

`ErrInconsistentSourceObservation` остаётся частью generic Loader contract для
других Source implementations. MemorySource никогда его не синтезирует. Его
runtime outcomes ограничены Not Found, Identity Mismatch, Version Not
Published, Source Integrity, Source Unavailable или success.

## 18. Defense-in-depth validation

Source классифицирует repository и representation facts. Loader сохраняет
ответственность за независимую повторную проверку request completeness,
ownership chain, lifecycle state, schema completeness и closed failure
contract. Source validation не ослабляет и не заменяет Loader validation.

## 19. Dependency graph

```text
future cmd composition root
    ├── configuration.MemoryConfigurationRepository
    ├── configurationversion.MemoryConfigurationVersionRepository
    └── configurationloadsource.MemorySource
            │ implements
            ▼
        configurationloader.Source
            ▼
        configurationloader.Loader
            ▼
        runtimelaunchflow.Flow

Flow.Start / Runtime Host не вызываются construction.
```

Adapter может зависеть от `configuration`, `configurationversion` и
`configurationloader`. Loader зависит только от Source boundary. Runtime
configuration, lifecycle, Flow и Runtime packages не зависят от adapter.
Dependency cycle запрещён.

## 20. Ownership

Repositories остаются во владении application composition root. MemorySource
borrow-ит их и не владеет entity или lifecycle. Loader borrow-ит Source; Flow
borrow-ит Loader и Owner. Construction этой цепочки не передаёт resource
ownership и не выполняет activation.

## 21. Lifetime

Оба repositories должны жить дольше MemorySource; MemorySource — дольше
каждого заимствующего его Loader; Loader — дольше каждого заимствующего его
Flow. Shutdown ordering принадлежит будущей application composition и не
добавляет adapter hook.

## 22. Concurrency

После construction MemorySource stateless и безопасен для concurrent calls:
он не добавляет mutable state и полагается на существующие concurrent-safe,
detaching repository reads. Он не добавляет mutex, cache, goroutine, retry или
background work.

## 23. Будущая composition

Будущая `cmd` composition root может сконструировать repositories, services,
MemorySource, Loader, Owner и Flow. Proposal разрешает только dependency
chain; он не разрешает routing management request, вызов `Flow.Start`,
публикацию Host или иную Runtime activation.

## 24. Будущие implementation proofs

Implementation task должна доказать:

1. compile-time conformance интерфейсу Source;
2. отсутствие constructor side effects и nil-binding behavior;
3. ровно один Version `Get` и не более одного Configuration `Get`;
4. отсутствие list, `GetPublished`, fallback, retry и re-read;
5. exhaustive identity, state, schema и error mapping;
6. глубокое two-way detachment;
7. эквивалентные repeated и concurrent loads;
8. outcomes publish/archive/delete races в L и C;
9. ровно один Configuration Service и один ConfigurationVersion Service над
   одной repository pair;
10. отклонение Composition Audit двух Version Service instances, любого
    direct writer, importer или migration;
11. regression proof, что stale Draft не перезаписывает Published через
    единственный Version Service;
12. single-Service serialization update/publish;
13. invariants Configuration update/delete, identity и non-resurrection;
14. завершение Composition Audit до Source construction и activation;
15. отсутствие test, ожидающего
    `ErrInconsistentSourceObservation` от MemorySource;
16. Loader integration с реальным adapter;
17. изолированный construction `Source -> Loader -> Flow` без Start или Host;
18. отсутствие detector, registry, global lock, retry, repository extension,
    dependency cycle, cache или goroutine;
19. targeted tests, stress и race detector при технической доступности.

## 25. Activation gate

До Source construction и повторно до Production Activation Composition Audit
должен доказать exact reference graph: одна repository pair, по одному Service
instance каждого типа, handler references только на эти instances, read-only
доступ MemorySource, service-only mutations и отсутствие direct или alternate
writers.

Если любой пункт не доказан, Source не конструируется и activation
блокируется. Audit является composition evidence, а не runtime detector,
registry, global lock, retry loop или repository extension.

## 26. Non-goals

- management HTTP routing, authentication или authorization;
- persistence или хранение Runtime identities;
- Production Activation;
- recovery, reconciliation, retry, supervision или caching;
- diagnostics transport;
- schema migration или negotiation;
- repository interfaces, extensions или redesign;
- lifecycle hooks или background work.

## 27. Implementation boundary

Implementation Status остаётся Planned. Package, constructor, methods, tests и
application wiring не появляются из-за этого документа. Реализация требует
отдельной task после design review и acceptance.

## 28. Решение

Планируемый минимальный Source — stateless adapter поверх двух существующих
concrete in-memory repositories. Он читает exact Version до exact parent,
считает Version read точкой L при audited single-instance mutation topology,
возвращает глубоко detached observation `uwp.configuration` v1 и отображает
runtime failures в применимую часть существующего Loader error set, не
синтезируя inconsistent-observation errors. Он позволяет future construction,
не разрешая management routing, persistence или Production Activation.

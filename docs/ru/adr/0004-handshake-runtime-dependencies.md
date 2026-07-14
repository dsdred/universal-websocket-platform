# ADR 0004: Handshake Runtime Dependency Boundary

[English version](../../en/adr/0004-handshake-runtime-dependencies.md)

## Статус

Принято.

## Контекст

Фундамент Runtime теперь содержит Admission Gate и Runtime context, принадлежащие Host. [ARCH-002](../architecture/ARCH-002-runtime-foundation-freeze.md) замораживает их ownership и lifecycle-семантику. [DP-001](../design/DP-001-runtime-handshake-pipeline.md) требует выполнять Authentication до Upgrade, проводить финальную проверку admission перед `websocket.Accept` и создавать Session context, lifecycle которого не зависит от `http.Request.Context()`.

Текущая композиция не может выполнить эти требования без явной dependency boundary:

- Handshake не получает Host-owned admission permission;
- Handshake не получает активный Runtime context;
- передача concrete Host инвертирует направление зависимостей и раскрывает посторонние lifecycle-операции;
- создание другого Gate или Runtime context в Listener дублирует mutable state и ownership;
- Listener не должен импортировать `internal/runtime`, а Runtime должен оставаться composition root, а не зависимостью transport-кода.

Решение должно соединить замороженный фундамент Runtime с открытой архитектурой Handshake, не меняя lifecycle-семантику Host, не раскрывая cancellation и не вводя общий dependency container.

## Решение

Runtime composition передает в Handshake две минимальные живые read-only Runtime capabilities через явную constructor injection:

1. **Admission permission capability** — наблюдает, разрешает ли Host-owned Admission Gate вход в admission commit в текущий момент.
2. **Runtime context capability** — возвращает опубликованный в текущий момент Host-owned Runtime context, не раскрывая его cancellation function.

Контракты ориентированы на потребителя и остаются узкими. Handshake зависит от семантики capabilities, а не от concrete Host, Runtime Container, enum lifecycle-состояний, реализации Admission Gate или реализации context holder. Точные Go-имена и расположение пакетов являются деталями реализации при условии, что направление зависимостей остается ациклическим и Listener не импортирует `internal/runtime`.

Capabilities являются живыми представлениями Host-owned state. Они не копируют состояние Gate или Runtime context в Handshake. Handshake не может открыть или закрыть admission, опубликовать или заменить Runtime context, отменить root context, запустить или остановить Host либо найти другую Runtime dependency.

### Admission Semantics

Handshake проверяет admission в двух точках:

1. до Authentication, чтобы закрытый Runtime не вызывал Authentication;
2. после Allow decision, последней операцией перед началом `websocket.Accept`.

Успешная финальная проверка является linearization point входа в концептуальное состояние `Committing`. Если Gate закрывается до этой проверки, Upgrade не начинается. Если Gate закрывается после успешной проверки, commit уже выполняется и следует правилам ownership и shutdown из DP-001: он должен завершить handoff в Runtime tracking или закрыть upgraded connection до завершения shutdown.

Allow decision не кэшируется как admission permission. Cancellation или закрытие Gate до финальной проверки лишает decision права на execution.

### Runtime Context Semantics

Runtime context capability не возвращает активный context до успешного startup commit. Handshake не должен кэшировать недоступный результат во время construction.

После успешных admission и Upgrade граница Upgrade создает дочерний connection/Session context от активного Runtime context. Lifecycle Session не выводится из `http.Request.Context()`. Host остается единственным владельцем cancellation корневого Runtime context. Дочерний connection context может иметь собственную cancellation function для завершения соединения; до handoff она принадлежит границе Upgrade, а после успешного принятия ownership — Session.

Если активный Runtime context недоступен во время commit, Handshake завершается с запретом и не создает Session.

## Порядок композиции

Композиция выполняется в следующем порядке без цикла Runtime-to-Listener:

```text
Snapshot
    -> Host-owned lifecycle primitives
    -> read-only Runtime capabilities
    -> Authentication
    -> Session handoff
    -> Handshake
    -> Listener
    -> Host startup transaction
```

Lifecycle primitives, стабильные реализации capabilities и неактивный Runtime context holder создаются при создании Host. `Build` не активирует их, не открывает admission, не публикует Runtime context и не запускает сетевые ресурсы.

Во время `Start` Runtime composition внедряет те же стабильные capabilities в Handshake, создает Listener и запускает его внутри существующей startup transaction. Admission остается закрытым во время construction и запуска Listener. Активный Runtime context создается и публикуется только в рамках успешного startup commit; затем Gate открывается под существующей lifecycle boundary Host. Запрос, поступивший до commit, не проходит начальную проверку admission и не запускает Authentication.

При ошибке startup активный Runtime context не публикуется, а Gate остается закрытым. При Stop закрытие Gate и cancellation root context немедленно наблюдаются через те же capabilities, при этом существующий порядок shutdown Host не меняется.

Контракты capabilities должны принадлежать потребляющей границе Handshake или узкому нейтральному пакету контрактов. Runtime composition предоставляет реализации или adapters. Handshake, Listener, Authentication и Session не импортируют concrete Runtime Host.

## Модель владения

| Ресурс или состояние | Владелец | Доступ Handshake |
|---|---|---|
| Admission Gate и его mutable state | Runtime Host | Живая read-only admission capability |
| Корневой Runtime context и его cancellation | Runtime Host | Живая read-only context capability без `CancelFunc` |
| HTTP request, ResponseWriter и принятый pre-Upgrade transport | `net/http` под lifecycle Listener | Синхронно заимствуются для выполнения Handshake |
| Authentication result и admission decision | Handshake одного запроса | Принадлежат до terminal rejection, abort или commit |
| WebSocket после успешного `websocket.Accept` и до handoff | Граница Upgrade | Exclusive ownership и обязанность закрыть при ошибке |
| Дочерний connection context до handoff | Граница Upgrade | Создается от Runtime context и отменяется при ошибке до handoff |
| WebSocket и дочерний connection context после успешного handoff | Session | Exclusive lifecycle и ответственность за закрытие |
| Runtime capability object | Runtime Host/composition | Не владеет WebSocket, Session, Principal или request state |

Ownership передается ровно один раз. Ошибка создания или принятия Session оставляет ownership у границы Upgrade, которая закрывает WebSocket и отменяет дочерний context. После явного успешного handoff верхние слои больше не закрывают WebSocket.

## Рассмотренные альтернативы

### Alternative A — передать Host целиком

Handshake может получить ссылку на concrete Host и вызывать его методы readiness, admission и context.

Альтернатива отклонена, потому что она связывает transport execution с реализацией lifecycle Runtime, раскрывает посторонние операции Start и Stop, способствует превращению Host в god object, усложняет сфокусированные тесты и создает риск цикла зависимостей `internal/runtime` и Listener. Она нарушает узкие dependency boundaries из ARCH-001 и ADR-0003.

### Alternative B — создать второй Gate в Listener

Listener может хранить собственное состояние admission и синхронизировать его с Host.

Альтернатива отклонена, потому что два mutable источника истины могут разойтись во время startup, rollback, concurrent Stop или будущей обработки failure. Она нарушает ownership Host, замороженный ARCH-002, и ставит корректность в зависимость от синхронизации двух Gate.

### Alternative C — передать read-only capabilities

Runtime composition передает только живой admission permission и доступ к Runtime context через узкие контракты или immutable набор callbacks.

Альтернатива выбрана. Она сохраняет ownership Host, оставляет Handshake независимым от concrete Runtime, поддерживает управляемые test doubles, не раскрывает cancellation или lifecycle mutation и делает wiring зависимостей явным. В рамках этого решения capability set закрыт этими двумя обязанностями; добавление несвязанных зависимостей требует отдельного сфокусированного design.

### Alternative D — передать immutable Runtime execution environment

Host может создать единый read-only environment object с доступом к admission и Runtime context.

Такой вариант способен обеспечить нужное поведение, но создает привлекательное место для накопления Resolver, configuration, diagnostics, registries и других services. Это превратит объект в service locator. Объединенное значение допустимо только как деталь реализации, если оно остается фиксированным carrier двух выбранных узких capabilities; оно не принимается как общий Runtime environment abstraction.

### Alternative E — перенести ownership Gate в Listener

Listener может владеть Gate, поскольку он выполняет transport commit.

Альтернатива отклонена. Она меняет ownership, замороженный ARCH-002, отделяет admission lifecycle от Runtime readiness и требует mutation Listener со стороны Host во время startup и shutdown. Такое направление потребовало бы явной ревизии архитектуры Runtime Foundation и не дает преимуществ перед внедрением Host-owned permission capability.

## Влияние на Frozen Foundation

Это решение является совместимым расширением composition boundary, описанной ARCH-002, а не изменением замороженной lifecycle-семантики.

Следующие frozen invariants остаются неизменными:

- Host остается production composition root и единственным владельцем lifecycle coordination;
- Host остается единственным владельцем состояния Admission Gate;
- Host остается единственным владельцем cancellation корневого Runtime context;
- `Build` не открывает сетевые ресурсы и не активирует ни одну capability;
- активный Runtime context, Running, readiness и открытый admission публикуются только после успешного startup commit;
- startup rollback оставляет readiness и admission закрытыми;
- Stop закрывает admission и отменяет Runtime context в существующем порядке;
- Container не превращается в service locator;
- семантика restart и reload не меняется.

ARCH-002 явно оставляет Handshake и применение Gate в admission commit открытыми. Поэтому ограниченный bridge через constructor injection не требует изменения freeze-документа или state machine Host. Реализация, которая меняет ownership Gate, lifecycle states, момент публикации context, ownership cancellation, startup commit, rollback, readiness или порядок shutdown, выйдет за рамки этого ADR и потребует нового DP или ADR.

## Последствия

### Преимущества

- Handshake может применять admission Host, не завися от Host.
- Lifecycle Session может происходить от Runtime lifecycle, а не от lifecycle HTTP request.
- Закрытие Gate и Runtime cancellation наблюдаются без дублирования mutable state.
- Listener остается независимым от `internal/runtime`.
- Composition остается явной, а test doubles могут независимо заменять каждую capability.
- Не появляются global state, dependency injection через `context.Value`, service locator или второй Gate.

### Недостатки

- Runtime composition должна явно передавать две дополнительные зависимости.
- До публикации активного Runtime context должен существовать стабильный context holder или эквивалентный adapter.
- Handshake не должен кэшировать результаты admission или context между запросами или lifecycle transitions.
- Финальная проверка admission должна оставаться непосредственно рядом с Upgrade commit для сохранения linearization semantics.

## Последующие изменения документации

После принятия этого ADR:

- DP-001 следует уточнить: Host владеет Gate, а Handshake владеет только admission capability и выполнением commit.
- DP-002 следует добавить ссылку на стабильный capability bridge в composition graph и construction order.
- ARCH-002 не требует изменения инвариантов; после реализации в него можно добавить ссылку на этот ADR как evidence разрешенной Handshake integration.
- `spec/current-state.md` должен отличать это принятое решение от реализованной pre-Upgrade Authentication, пока не появятся код и тесты.
- MASTER_PLAN следует обновлять только после существенного изменения состояния реализации Beta.

## Ссылки

- [ADR-0003: Runtime Architecture](0003-runtime-architecture.md)
- [ARCH-001: Runtime Architectural Pattern](../architecture/ARCH-001-runtime-architectural-pattern.md)
- [ARCH-002: Runtime Foundation Freeze](../architecture/ARCH-002-runtime-foundation-freeze.md)
- [DP-001: Runtime Handshake Pipeline](../design/DP-001-runtime-handshake-pipeline.md)
- [DP-002: Runtime Host Composition Root](../design/DP-002-runtime-host-composition-root.md)
- [Master Engineering Plan](../roadmap/MASTER_PLAN.md)


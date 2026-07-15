# DP-003: Runtime Session Manager

[English version](../../en/design/DP-003-runtime-session-manager.md)

## 1. Статус

**Статус:** Draft

Эта редакция упрощает lifecycle после второго независимого архитектурного ревью. Она отделяет shutdown transition от ожидания, сводит completion к одной атомарной mutation и определяет непрерывный ownership каждой registration transaction. Конкретные Go API и соседние подсистемы Runtime не проектируются.

## 2. Постановка проблемы

Runtime умеет создавать и запускать Session после успешного Handshake, но в нем нет авторитетного registry и полного shutdown accounting переданных Session. Сейчас HTTP handler синхронно владеет Session `Start/Run/Stop`. Runtime пока не может доказать, что каждый успешный handoff отслеживается до terminal completion или что shutdown не пропустит незавершенную регистрацию.

[DP-001](DP-001-runtime-handshake-pipeline.md) требует включить Session в Runtime shutdown tracking до завершения handoff. [ARCH-002](../architecture/ARCH-002-runtime-foundation-freeze.md) оставляет ownership Session и полный shutdown wait set открытой архитектурой. Этот документ определяет только такой lifecycle foundation, сохраняя замороженные semantics Host, Admission Gate, Runtime context, startup, rollback и Listener lifecycle.

## 3. Цели

- Сохранить единственного owner WebSocket на всем пути Handshake и handoff Session.
- Определить ownership reservation и committed registration без untracked gap.
- Установить одну linearization point регистрации и одну для completion.
- Отделить неблокирующий shutdown transition от ожидания, ограниченного context.
- Сделать race `Reserve/Register/Complete/BeginShutdown/Wait` детерминированными.
- Предоставить identity-safe lookup metadata без раскрытия операций Session.
- Ограничить Manager registry и shutdown accounting.

## 4. Не-цели

Документ не проектирует Router, Delivery, Presence, Groups, Persistence, cluster ownership, Plugins, Metrics, diagnostics aggregation, policy лимитов Session, transport protocol, публичный API, Configuration fields, конкретные Go-интерфейсы, generic Session framework или supervision framework.

## 5. Границы ownership

Четыре области ownership остаются разделенными.

### Ownership WebSocket Transport

Upgrade boundary владеет WebSocket и дочерним connection context, пока Session явно их не примет. До принятия Upgrade boundary является единственным closer. После принятия Session становится единственным closer, в том числе при ошибке последующей регистрации.

### Ownership Registration Transaction

Успешный Reserve возвращает один концептуальный reservation handle. Вызывающий control flow Handshake/execution setup владеет этим handle до ровно одной terminal operation:

- `Commit`, который поглощает reservation и создает committed record;
- `Abort`, который удаляет reservation без видимости registry.

Нормативный pattern control flow:

```text
reservation := Manager.Reserve(SessionID)
defer reservation.AbortUnlessCommitted()
```

Эта запись описывает ownership obligation, а не обязательный Go API. Каждый recoverable return, rejection, error или panic до Commit обязан выполнить Abort. Manager учитывает reservation, но не владеет goroutine, обязанной выполнить Commit или Abort.

### Ownership выполнения Session

Узкий владелец выполнения отдельной Session создается до Commit регистрации и существует как минимум до завершения completion accounting. Он владеет Session `Start/Run/Stop`, вызывает Complete через terminal defer на всех recoverable paths и предоставляет стабильную non-owning capability запроса Stop для shutdown Runtime. Это не generic supervisor.

### Ownership Tracking Manager

Manager владеет только reservation accounting, committed membership registry, immutable lookup views, shutdown transition и completion accounting. Он не владеет выполнением Session или WebSocket transport.

## 6. Ответственность Session Manager

Manager содержит только ответственность, связанную одним инвариантом: Runtime становится `Closed` только после разрешения каждой reserved transaction и completion каждой committed registration.

Manager отвечает за:

- reservation accounting;
- committed membership registry;
- выделение RegistrationID и identity-safe completion;
- immutable lookup по SessionID;
- атомарный transition `Open -> Closing`;
- committed shutdown snapshot;
- атомарный completion accounting;
- ограниченное context ожидание пустого accounting.

Manager не выполняет Session, не владеет WebSocket, не маршрутизирует и не отправляет сообщения, не публикует Presence events, не хранит terminal history, не агрегирует diagnostics, не применяет policy лимитов Session и не знает Router, Delivery, Presence, Persistence или Groups.

Registry может оставаться внутренним в первой версии, поскольку reservation, commit, lookup и completion разделяют один атомарный invariant. Граница должна допускать последующее выделение без изменения этой семантики.

## 7. Registration Transaction

Нормативная последовательность:

```text
Reserve
    -> создать Session
    -> создать execution owner
    -> Session принимает transport ownership
    -> Commit registration с RegistrationID и Stop-request capability
    -> execution owner выполняет Start/Run
    -> deferred Complete
```

### Reserve

Reserve разрешен только пока Manager находится в `Open`. Он валидирует SessionID и создает:

- reservation handle с исключительным обязательством Commit-or-Abort;
- RegistrationID, уникальный за весь срок жизни этого Manager.

Reservation учитывается для закрытия Manager, но невидима Lookup и не является committed Session в shutdown registry.

### Commit

Commit поглощает reservation и атомарно создает один committed record с immutable identity metadata и стабильной Stop-request capability. Эта mutation является единственной registration linearization point: видимость registry и membership committed wait set появляются вместе.

### Abort

Abort атомарно удаляет uncommitted reservation и уведомляет Wait callers, если accounting теперь может быть пуст. Abort никогда не создает lookup visibility или committed record. Повторный Abort после Commit или Abort не имеет второго accounting effect.

### Transport Handoff

Session принимает transport до Commit. При ошибке до принятия Upgrade boundary закрывает транспорт, а owner reservation выполняет Abort. При ошибке после принятия, но до Commit, Session закрывает transport через execution owner, а owner reservation выполняет Abort. В каждый момент существует ровно один transport owner.

## 8. Linearization Reserve/Commit/BeginShutdown

Документ выбирает строгую shutdown model.

На одной synchronization boundary Manager:

- либо Commit завершается до `BeginShutdown`, становится committed record и входит в shutdown snapshot;
- либо побеждает `BeginShutdown`, Manager переходит в `Closing`, а каждая uncommitted reservation теряет право на Commit.

Новый Reserve после `Closing` отклоняется. Существующая reservation, чей Commit проиграл, должна выполнить Abort через своего owner. Если Session уже приняла transport, execution owner запрашивает Stop до завершения Abort. Manager не выполняет forced Abort handle, поскольку не владеет этим control flow.

Late Commit в `Closing` и partially visible registration отсутствуют. Manager не может стать `Closed`, пока каждая invalidated reservation не достигла Abort.

## 9. Registration Identity

Каждая reservation получает RegistrationID со следующими нормативными свойствами:

- opaque и immutable;
- уникален за весь срок жизни одного Manager;
- никогда не используется повторно, включая состояния после Abort или Complete;
- связан ровно с одной reservation и не более чем одним committed record;
- неизвестный, completed или stale RegistrationID не изменяет accounting.

Manager может выделять identity без хранения removed records, если значения никогда не повторяются за срок его жизни.

SessionID:

- обязателен и не пуст;
- имеет длину не более 255 bytes;
- opaque и immutable;
- сравнивается byte-for-byte без normalization;
- уникален среди текущих reservations и committed records.

SessionID можно использовать повторно после Abort или Complete. Identity одной регистрации — пара `(SessionID, RegistrationID)`. Manager не логирует raw SessionID автоматически; безопасная diagnostics требует явного encoding.

SessionID относится к одному экземпляру Runtime и не является durable cross-Runtime identity.

## 10. Lifecycle Session и Record Manager

DP-003 не добавляет состояния в существующую state machine Session.

Session остается источником истины execution:

```text
Created -> Running -> Stopping -> Stopped
```

Record Manager отражает только состояние регистрации:

```text
Reserved -> Registered -> Removed
       \-> Aborted
```

- `Reserved` невидим и несет обязательство Commit-or-Abort.
- `Registered` видим и учитывается в committed shutdown registry.
- `Removed` следует за первым valid Complete и немедленно становится невидимым и неучитываемым.
- `Aborted` никогда не становится видимым или committed.

Нормативного состояния `Completing` нет. Состояние Manager никогда не утверждает, что Session находится в Running; состояние Session никогда не утверждает membership registry.

## 11. Lifetime Execution Owner

Execution owner создается до Commit. Commit записывает его стабильную non-owning Stop-request capability. Execution owner остается жив, пока:

- выполнение Session не достигло recoverable terminal path;
- он не попытался выполнить Complete;
- каждый уже выданный через стабильную capability Stop request не вернул управление.

Execution owner вызывает `Start`, затем `Run` и вызывает `Stop` согласно существующему контракту Session. Terminal defer пытается выполнить Complete ровно один раз для каждого recoverable return, error или panic внутри owned execution goroutine.

Если BeginShutdown побеждает после Commit, но до Start, Stop-request capability может отменить startup или остановить Session из состояния `Created`. Start после этого пропускается или завершается ошибкой как штатный terminal path. Defer все равно пытается выполнить Complete, поэтому committed record не зависает.

Complete не уничтожает capability, пока ранее выданный Stop request выполняется. Shutdown snapshot владеет ссылкой capability до возврата этого вызова; он не владеет Session и не продлевает видимость registry.

## 12. Linearization Completion

Complete идентифицирует committed record по RegistrationID. Первый valid Complete выполняет одну атомарную mutation:

```text
Registered -> Removed
```

В этой единственной linearization point Manager:

- удаляет lookup visibility;
- удаляет committed record;
- уменьшает accounting committed wait set;
- уведомляет Wait callers, что accounting может быть пуст.

Повторный Complete для того же RegistrationID является no-op или возвращает один стабильный already-completed/unknown-registration contract result. Он никогда не уменьшает accounting повторно. Поскольку RegistrationID никогда не используется повторно, stale completion не может повлиять на более позднюю регистрацию с повторно используемым SessionID.

Complete учитывает только завершение lifecycle. Session и execution owner остаются ответственными за cleanup ресурсов. Manager не сохраняет terminal result Session после Complete и не агрегирует его в shutdown Manager.

## 13. Lifecycle Manager

Lifecycle Manager:

```text
Open -> Closing -> Closed
```

Restart запрещен.

### Open

- Reserve и Lookup разрешены.
- Commit и Abort разрешены для owned reservations.
- Complete разрешен для committed records.
- BeginShutdown разрешен.

### Closing

- Reserve и Commit отклоняются.
- Abort и Complete остаются разрешены.
- Lookup возвращает только текущие Registered views.
- BeginShutdown идемпотентен и относится к тому же shutdown cycle.
- Wait разрешен.
- transition в `Closed` происходит автоматически, когда reservation и committed accounting становятся пустыми.

### Closed

- reservation set и committed registry пусты;
- Reserve и Commit отклоняются;
- Lookup возвращает отсутствие;
- BeginShutdown идемпотентен и не предоставляет active Stop capabilities;
- Wait возвращает nil;
- Abort и Complete не имеют accounting effect.

## 14. BeginShutdown

`BeginShutdown` — концептуальный неблокирующий transition contract. Он:

- атомарно один раз переводит `Open` в `Closing`;
- запрещает новый Reserve и каждый uncommitted Commit;
- помечает существующие reservations как требующие Abort;
- создает или предоставляет стабильный shutdown snapshot всех records, committed до transition;
- не выполняет network I/O;
- не вызывает операции Session;
- ничего не ожидает;
- идемпотентен и при повторе присоединяется к тому же shutdown cycle.

Shutdown snapshot содержит только стабильные non-owning Stop-request capabilities. В нем нет raw Session, WebSocket, Handler или mutable registry record.

Removal через Complete не инвалидирует capability, уже выданную текущей shutdown orchestration. Ее lifetime продолжается до возврата соответствующего Stop request. Повторный BeginShutdown не создает новый cycle и не дублирует accounting.

## 15. Stop-Request Capabilities

Stop-request capability представляет разрешение запросить завершение у одного execution owner без получения ownership Session.

Каждый request должен быть неблокирующим или ограниченным context caller и не должен ждать полного completion Session. Stop requests разных registrations независимы: один blocked или slow request не может сериализовать запросы ко всем остальным Session через Manager.

Runtime shutdown orchestration владеет вызовом capabilities snapshot. Manager только создает стабильный snapshot и позже наблюдает mutations Complete. Generic executor pool, worker framework или Session supervisor не вводятся.

## 16. Wait

`Wait(ctx)` наблюдает единственный shutdown cycle Manager. Он не инициирует shutdown и не выполняет Stop requests.

Wait возвращает nil, когда одновременно пусты:

- reservation accounting;
- committed registry accounting.

Mutation, удаляющая последнюю reservation или committed record, автоматически переводит `Closing` в `Closed` и уведомляет всех Wait callers.

Если первым завершается context caller:

- Wait возвращает `ctx.Err()` этому caller;
- Manager остается в `Closing`, пока accounting не пуст;
- ни один record не удаляется ложно;
- последующий Wait продолжает наблюдать тот же cycle.

В этом proposal у Manager нет собственной operational failure model. Session errors не агрегируются и не сохраняются. BeginShutdown возвращает только результаты проверки контракта, а Wait после `Closed` возвращает nil вместо неопределенного terminal error Manager.

## 17. Порядок Shutdown Runtime

Исполнимый порядок Runtime:

```text
Host закрывает Admission Gate
    -> Manager.BeginShutdown
    -> Host отменяет корневой Runtime context
    -> Runtime shutdown orchestration независимо вызывает стабильные Stop-request capabilities
    -> Host вызывает существующий Listener Stop
    -> active executions Session завершаются и вызывают Complete
    -> зависимые HTTP handlers завершаются
    -> Listener Stop возвращает управление
    -> Manager.Wait
    -> Manager Closed
    -> остальные компоненты Runtime останавливаются
```

`BeginShutdown` не может заблокировать Host до cancellation корневого context. Manager Wait выполняется после Listener Stop, а Complete происходит на terminal paths выполнения Session до возврата их handlers. Поэтому Listener может ожидать handlers, не ожидая Manager Wait, а Manager Wait наблюдает accounting, независимо стремящийся к завершению. Ни один lifecycle lock не удерживается во время Stop requests, Listener Stop, выполнения Session или Wait.

Этот порядок сохраняет [ARCH-002](../architecture/ARCH-002-runtime-foundation-freeze.md): Host закрывает admission и отменяет корневой Runtime context до вызова существующего Listener Stop. Новые состояния Host не добавляются, ownership Listener не меняется.

## 18. Ограничения Panic, Failure и Abandonment

Execution owner гарантирует deferred попытку Complete для всех recoverable returns, errors и panics внутри owned goroutine. Owner reservation гарантирует deferred Abort для всех recoverable exits до Commit.

Completion не гарантируется после:

- завершения процесса;
- unrecoverable runtime failure;
- навсегда заблокированной goroutine;
- нарушения контракта reservation или execution owner внешним компонентом.

Abandoned reservation или execution может оставить Manager в `Closing` навсегда. Это accepted limitation правдивого shutdown accounting: Manager не выдумывает completion и не завершает goroutine принудительно.

## 19. Контракт Lookup

Lookup первой версии выполняет точный поиск по SessionID и возвращает immutable RegistrationView только с:

- SessionID;
- RegistrationID;
- manager-visible state, который в этой версии равен `Registered` для каждой видимой записи.

Lookup никогда не возвращает raw Session, execution owner, Stop capability, операцию Send или mutable record.

Семантика Lookup:

- reservation невидима;
- Complete удаляет видимость в своей единственной linearization point;
- lookup не продлевает lifetime;
- возвращенный view может устареть немедленно;
- `(SessionID, RegistrationID)` различает повторное использование SessionID;
- во время `Closing` видны только records, все еще находящиеся в Registered;
- после `Closed` lookup возвращает отсутствие.

RegistrationView не является контрактом Router, Delivery, Presence, lease или targeting.

## 20. Runtime Context

Manager не создает, не владеет, не заменяет и не отменяет корневой Runtime context. Host остается его единственным owner согласно [ADR-0004](../adr/0004-handshake-runtime-dependencies.md). Root cancellation является одним shutdown signal для execution owners; accounting Manager меняет только Complete.

## 21. Архитектурные инварианты

- У WebSocket всегда ровно один owner.
- Reservation handle имеет ровно один terminal accounting effect Commit-or-Abort.
- Commit регистрации — единственная linearization point видимости registry.
- BeginShutdown — единственная linearization point `Open -> Closing`.
- После BeginShutdown ни один Commit не завершается успешно.
- Complete — одна атомарная mutation `Registered -> Removed`.
- RegistrationID никогда не повторяется за срок жизни Manager.
- Stale Complete не может повлиять на другую регистрацию.
- Manager становится `Closed` только при пустых reservation и committed accounting.
- BeginShutdown не выполняет I/O, а Wait не выполняет Stop requests.
- Stop capabilities стабильны во время текущего shutdown invocation и никогда не раскрывают raw Session.
- Lookup возвращает только immutable identity metadata и никогда не продлевает lifetime.
- Manager не выполняет Session, не владеет transport, не маршрутизирует, не доставляет, не публикует Presence, не хранит history, не агрегирует diagnostics и не применяет limits.

## 22. Будущая интеграция и Limits

DP-003 гарантирует только ownership-safe registration и правдивый shutdown tracking.

- Router требует отдельного targeting contract.
- Delivery требует отдельной lifetime-aware capability.
- Presence требует отдельного event или snapshot contract.
- Persistence требует identity экземпляра Runtime.
- Groups не входят в proposal.

Лимиты Session остаются вне DP-003. Lifecycle Manager предоставляет возможную будущую точку наблюдения registration admission, но policy, configured limits и rejection semantics требуют отдельной Configuration validation и сфокусированного design. Поэтому эпик Session Manager из MASTER_PLAN концептуально разделяется на этот lifecycle foundation и будущую limits capability.

## 23. Предотвращение God Object

Manager содержит только reservation accounting, committed registry, identity-safe lookup views, shutdown transition и completion accounting. Эти обязанности связаны единым атомарным инвариантом пустоты.

Manager явно не выполняет Session, не владеет WebSocket, не маршрутизирует и не отправляет сообщения, не публикует domain events, не хранит terminal history, не агрегирует diagnostics и не применяет limits policy. Будущие системы не могут добавлять эти обязанности только потому, что Manager раскрывает registration metadata.

## 24. Альтернативы

### Альтернатива A — Один блокирующий Manager Stop

**Преимущество:** один lifecycle call.

**Недостаток:** Host не может детерминированно закрыть регистрацию, отменить Runtime, вызвать Listener Stop и только затем ждать без скрытого asynchronous behavior.

**Решение:** отклонено в пользу отдельных BeginShutdown и Wait.

### Альтернатива B — Разрешить Reserved Transactions выполнить Commit в Closing

**Преимущество:** меньше aborted handoffs во время конкурентного shutdown.

**Недостаток:** shutdown snapshot остается открытым для late committed records и требует динамического обнаружения Stop capabilities.

**Решение:** отклонено для первой версии. Строгий BeginShutdown инвалидирует каждую uncommitted reservation.

### Альтернатива C — Возвращать Raw Session из Lookup

**Преимущество:** результат сразу полезен callers.

**Недостаток:** раскрывает mutable lifecycle и создает неоднозначность stale reference и ownership.

**Решение:** отклонено; Lookup возвращает только immutable RegistrationView.

### Альтернатива D — Хранить Completion History

**Преимущество:** отличает repeated completion от unknown identity и помогает diagnostics.

**Недостаток:** превращает Manager в history storage и увеличивает state за пределы live accounting.

**Решение:** отклонено. Unknown и already-completed RegistrationID имеют одинаковую семантику отсутствия accounting effect.

### Альтернатива E — Немедленно выделить Registry Component

**Преимущество:** изолирует indexing.

**Недостаток:** разделяет единый invariant registration/completion до появления независимых consumers.

**Решение:** отклонено для первой версии; внутренняя граница все равно должна допускать последующее выделение.

## 25. Стратегия миграции

Внутренний lifecycle существенно меняется.

Текущая модель:

```text
HTTP handler синхронно владеет Session Start/Run/Stop
```

Предлагаемая модель:

```text
Handshake владеет reservation
    -> создает Session и execution owner
    -> Session принимает transport
    -> registration выполняет Commit
    -> execution owner запускает Session и откладывает Complete
```

Внешнее WebSocket behavior сохраняется, но внутренние concurrency, ownership, registration и shutdown semantics меняются. Поэтому миграцию следует считать lifecycle work, а не behavior-neutral refactor, и проводить race review каждой границы.

## 26. Открытые вопросы

- Какая конкретная форма non-owning Stop-request capability соответствует контракту bounded return?
- Как terminal results Session достигнут будущей diagnostics boundary без aggregation в Manager?
- Какой будущий design владеет configured Session limits?
- Какую identity экземпляра Runtime будущая Persistence объединит с SessionID?
- Какова operational response на accepted permanently blocked reservation или execution?

Эти вопросы не изменяют определенные здесь linearization points или accounting Manager.

## 27. Трассировка архитектурного ревью

| Finding | Статус | Решение |
| --- | --- | --- |
| F-01 | Resolved | BeginShutdown выполняет неблокирующий transition; Wait является отдельным context-bounded наблюдением. |
| F-02 | Resolved | Complete — одна атомарная mutation `Registered -> Removed` и единственная linearization point visibility/accounting. |
| F-03 | Resolved | Reservation handle дает Handshake/execution setup непрерывный ownership Commit-or-Abort. |
| F-04 | Resolved | Execution owner существует до Commit; shutdown между Commit и Start — штатный terminal path, завершающийся Complete. |
| F-05 | Resolved | Shutdown snapshot содержит стабильные Stop capabilities, lifetime вызова которых переживает Complete. |
| F-06 | Resolved | Stop requests независимы и неблокирующие либо ограничены context caller; Manager их не вызывает. |
| F-07 | Resolved | Runtime shutdown orchestration владеет Stop requests; Manager прогрессирует через Abort/Complete mutations без ожидающего caller. |
| F-08 | Resolved | RegistrationView включает никогда не повторяющийся RegistrationID и различает повторное использование SessionID. |
| F-09 | Resolved | RegistrationID уникален за срок жизни Manager и никогда не повторяется. |
| F-10 | Resolved | Неопределенный terminal error Manager удален; Wait после Closed возвращает nil. |
| F-11 | Clarified | Deferred completion гарантируется только для recoverable paths owned goroutine; nonrecoverable cases являются accepted limitations. |
| F-12 | Clarified | Lookup намеренно возвращает только metadata и не является operational capability Session. |
| F-13 | Resolved | Lookup возвращает immutable RegistrationView, но не raw Session или mutable execution capability. |
| F-14 | Resolved | Нормативное состояние `Completing` удалено. |
| F-15 | Resolved | Execution owner создается до Commit и переживает completion и outstanding Stop invocations. |
| F-16 | Clarified | SessionID opaque, сравнивается byte-for-byte, имеет длину до 255 bytes и не логируется raw автоматически. |
| F-17 | Clarified | Migration явно признает существенные изменения внутреннего lifecycle и concurrency. |
| F-18 | Clarified | Ответственность Manager ограничена явными negative invariants. |
| F-19 | Clarified | Exactly-once описывает effective accounting mutation, а не доставку notification. |
| F-20 | Accepted limitation | Permanent abandonment может сохранять правдивый accounting и удерживать Manager в Closing. |
| F-21 | Deferred | Limits выделены в будущую сфокусированную Configuration/admission capability. |
| F-22 | Resolved | BeginShutdown — явный неблокирующий contract `Open -> Closing`. |
| F-23 | Resolved | Complete атомарно удаляет visibility и accounting wait set. |
| F-24 | Resolved | Reservation handle является непрерывным owner до Commit или Abort. |
| F-25 | Resolved | Shutdown snapshot удерживает lifetime Stop capability независимо от visibility registry. |
| F-26 | Resolved | Manager не агрегирует terminal errors; Wait возвращает nil или caller `ctx.Err()`. |
| F-27 | Resolved | Immutable view содержит RegistrationID, поэтому stale observation и повторное использование SessionID различимы. |

## 28. Готовность к утверждению

Blocker findings второго ревью F-01–F-03 закрыты разделением shutdown contracts, атомарным Complete и reservation handles с непрерывным ownership. High findings F-04–F-10 закрыты lifecycle execution owner, стабильными Stop capabilities, независимой shutdown orchestration, identity-safe views и удалением неопределенных terminal errors Manager.

Accepted limitations определены явно: завершение процесса, unrecoverable failure, permanently blocked goroutine или нарушение внешнего контракта могут не позволить accounting стать пустым. Manager правдиво остается в `Closing`, а не объявляет ложный completion.

Оставшиеся вопросы касаются конкретной формы capability, будущей доставки diagnostics, configured limits, durable Runtime identity и operational обработки permanently blocked work. Они не изменяют core lifecycle.

**Кандидат решения:** Approved with Findings.

Документ сохраняет статус Draft, пока независимое ревью не подтвердит упрощенные ordering BeginShutdown/Wait, атомарный completion, ownership reservation и lifetime стабильной capability.

## 29. Ссылки

- [ARCH-001: Runtime Architectural Pattern](../architecture/ARCH-001-runtime-architectural-pattern.md)
- [ARCH-002: Runtime Foundation Freeze](../architecture/ARCH-002-runtime-foundation-freeze.md)
- [DP-001: Runtime Handshake Pipeline](DP-001-runtime-handshake-pipeline.md)
- [DP-002: Runtime Host Composition Root](DP-002-runtime-host-composition-root.md)
- [ADR-0004: Handshake Runtime Dependency Boundary](../adr/0004-handshake-runtime-dependencies.md)
- [Master Engineering Plan](../roadmap/MASTER_PLAN.md)

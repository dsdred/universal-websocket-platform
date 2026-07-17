# DP-003: Runtime Session Manager

[English version](../../en/design/DP-003-runtime-session-manager.md)

## 1. Статус

**Статус:** Draft

Эта редакция определяет контракты Session Manager для registration, identity, lookup, shutdown accounting и shutdown snapshot. Детальный ownership per-Session execution нормативно определён в [DP-004](DP-004-per-session-execution-boundary.md) и здесь не дублируется.

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
- Определить Manager-side integration boundary для per-Session execution contract из DP-004.
- Ограничить Manager registry и shutdown accounting.

## 4. Не-цели

Документ не проектирует Router, Delivery, Presence, Groups, Persistence, cluster ownership, Plugins, Metrics, diagnostics framework, policy лимитов Session, transport protocol, публичный API, Configuration fields, generic Session framework или supervision framework.

## 5. Границы ownership

Четыре области ownership остаются разделенными.

### Ownership WebSocket Transport

Upgrade boundary владеет WebSocket и дочерним connection context до successful integrated boundary Commit. До Commit Upgrade boundary является единственным closer. Successful Commit одновременно публикует Registration и передаёт ownership Session; последующего failure регистрации не существует. После Commit Session является единственным closer.

### Ownership Registration Transaction

Успешный Reserve возвращает один концептуальный reservation handle. Control flow Session handoff сохраняет непрерывный ownership этого handle до ровно одной terminal operation:

- `Commit`, который поглощает reservation и создает committed record;
- `Abort`, который удаляет reservation без видимости registry.

Нормативный pattern control flow:

```text
reservation := Manager.Reserve(SessionID)
defer reservation.AbortUnlessCommitted()
```

Эта запись описывает ownership obligation, а не обязательный Go API. Каждый recoverable return, rejection, error или panic до Commit обязан выполнить Abort. Manager учитывает reservation, но не владеет control flow, обязанным выполнить Commit или Abort. Handshake никогда не владеет reservation и не импортирует Session Manager; post-Upgrade preparation делегируется границе Session handoff. Точная передача от этой transaction к execution ownership определяется DP-004.

### Ownership выполнения Session

Session Manager не определяет и не владеет execution Session. [DP-004](DP-004-per-session-execution-boundary.md) является нормативным источником для creator boundary, модели provisional Session, активации ownership, порядка Start/Run/Stop, connection cancellation, terminal observation, обработки panic и lifetime Stop request.

Manager участвует только через opaque registration identity, accounting Commit/Complete, immutable shutdown identity и будущие узкие integration contracts, явно назначенные ему DP-004. Он никогда не получает Session, WebSocket, HTTP request, connection context или методы управления execution.

### Ownership Tracking Manager

Manager владеет только reservation accounting, committed membership registry, immutable lookup views, shutdown transition и completion accounting. Он не владеет выполнением Session или WebSocket transport.

## 6. Ответственность Session Manager

Manager содержит только ответственность, связанную одним инвариантом: Runtime становится `Closed` только после разрешения каждой reserved transaction и completion каждой committed registration. При наличии интеграции DP-004 та же wait boundary дополнительно требует release каждого opaque Owner Lifetime Lease; это additive accounting execution lifetime, не изменяющий mutations reservation или registration.

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

Последовательность, видимая Manager:

```text
Upgrade boundary вызывает Session handoff/Dispatcher
    -> Reserve SessionID
    -> Commit registration
    -> раскрыть immutable registration identity
    -> execution выполняется согласно DP-004
    -> Complete
```

### Reserve

Reserve разрешен только пока Manager находится в `Open`. Он валидирует SessionID и создает:

- reservation handle с исключительным обязательством Commit-or-Abort;
- RegistrationID, уникальный за весь срок жизни этого Manager.

Reservation учитывается для закрытия Manager, но невидима Lookup и не является committed Session в shutdown registry.

### Commit

Commit поглощает reservation и атомарно создаёт один committed record с immutable identity metadata. Эта mutation является единственной registration linearization point: видимость registry и membership committed wait set появляются вместе. В текущей реализации успешный Commit возвращает opaque RegistrationID выполняющему Commit control flow.

Интегрированный contract DP-004 расширяет тот же атомарный успешный результат одним Completion Adapter, привязанным к RegistrationID, одним Owner Lifetime Lease, одним Stop-publication binding и одним узким one-shot execution-publication binding, подготовленными до Commit. Dispatcher создаёт ровно один dormant execution path до вызова Commit. Под той же synchronization boundary Manager, которая исключает Lookup и BeginShutdown, Commit публикует complete committed record и owner-scoped bundle и вызывает единственную publication operation binding с outcome `Committed`. Synchronization boundary освобождается только после завершения обеих mutations.

Execution-publication binding не является пассивным value, ссылкой на Execution Owner или общей lifecycle capability. Он предоставляет Manager только one-shot право опубликовать один уже подготовленный outcome. Publication `Committed` делает ровно один dormant path eligible; publication non-committed outcome завершает этот path без execution. Repeated publication возвращает существующий outcome и не имеет второго effect. Manager не создаёт, не хранит, не запускает, не останавливает и не наблюдает owner. Dispatcher отвечает за создание ровно одного dormant path и передачу ровно одного binding.

Либо complete committed record, owner-scoped bundle и committed execution binding становятся observable вместе, либо committed state не создаётся. Это расширение не вводит вторую фазу Commit и не изменяет registration linearization point.

Каждое потенциально fallible computation и allocation integrated bundle завершается до входа Manager в publication critical section. Эта section не вызывает внешний код и выполняет только bounded mutations Manager и подготовленную panic-free one-shot publication operation. Поэтому recoverable Commit имеет только non-committed failure до observable mutation либо полную successful publication. Process termination или unrecoverable failure Go runtime находятся вне этой recoverable atomicity guarantee.

Commit необратим. Commit не возвращает success, пока path owner не стал eligible для выполнения committed lifecycle. После successful Commit Registration externally observable через Lookup, входит в shutdown accounting, может быть зафиксирована BeginShutdown и уже имеет ровно один committed execution path. Она может исчезнуть только через normal mutation `Complete`. Уже зафиксированный Snapshot остаётся immutable.

Repeated Commit, пока та же Registration существует, возвращает тот же logical bundle, привязанный к RegistrationID: те же RegistrationID, bound Completion Adapter, identity Owner Lifetime Lease, Stop-publication binding и committed execution outcome. Он не вызывает publication повторно, не вызывает повторную установку Runtime callback, не создаёт другой lease, Stop capability или goroutine и не меняет accounting. После Complete сохраняются существующие semantics terminal handle, сообщающие об удалённой Registration.

### Abort

Abort атомарно удаляет uncommitted reservation и уведомляет Wait callers, если accounting теперь может быть пуст. Abort никогда не создает lookup visibility или committed record. Повторный Abort после Commit или Abort не имеет второго accounting effect.

### Transport Handoff

Session Manager не определяет transport handoff и никогда не владеет transport. DP-004 до Commit подготавливает Session, attachment transport, control owner, identity lease, Stop publication и ровно один dormant execution path без публикации Registration или transfer ownership. Successful Commit является единственной irreversible publication point и точкой acceptance handoff: до освобождения его synchronization boundary complete bundle Registration и one-shot execution binding committed вместе. Failed Commit не оставляет Registration или lease, завершает dormant path через pre-Commit disposal, требует Abort и сохраняет cleanup transport у Dispatcher. Ни один recoverable post-Commit failure не возвращает `accepted=false`.

## 8. Linearization Reserve/Commit/BeginShutdown

Документ выбирает строгую shutdown model.

На одной synchronization boundary Manager:

- либо Commit завершается до `BeginShutdown`, становится committed record и входит в shutdown snapshot;
- либо побеждает `BeginShutdown`, Manager переходит в `Closing`, а каждая uncommitted reservation теряет право на Commit.

Новый Reserve после `Closing` отклоняется. Существующая reservation, чей Commit проиграл, должна выполнить Abort через текущего owner transaction. Complete не вызывается, поскольку committed record не существует. Manager не выполняет forced Abort handle, поскольку не владеет этим control flow.

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

## 11. Граница Per-Session Execution

Детальная модель Execution Owner перенесена в [DP-004](DP-004-per-session-execution-boundary.md). DP-003 остаётся нормативным только для Reservation, Commit, RegistrationID, Complete, Lookup, lifecycle Manager, shutdown accounting и реализованного identity-only Shutdown Snapshot.

DP-004 добавляет узкую интеграцию вокруг committed records: атомарный owner-scoped Commit bundle, хранение Stop-request capability, accounting lifetime owner и совместимый accessor capability в shutdown snapshot. Эти дополнения не должны изменять существующий смысл или linearization points Reserve, Commit, Abort, Complete, Lookup или BeginShutdown, а identity-only callers Snapshot сохраняют существующее поведение.

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
- после интеграции DP-004 `Closed` также требует пустого accounting Owner Lifetime Lease.

### Closed

- reservation set и committed registry пусты;
- после интеграции DP-004 accounting Owner Lifetime Lease пуст;
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

Текущий реализованный shutdown snapshot содержит только detached immutable identity Registration. В нём нет raw Session, WebSocket, Handler, Execution Owner, callback или mutable registry record.

DP-004 определяет будущую узкую интеграцию Stop request и owner lifetime. Эта будущая интеграция должна сохранить atomic capture первого snapshot, исключение reservation, linearization Commit/BeginShutdown, semantics повторного BeginShutdown и независимость captured identity snapshot от последующего Complete.

## 15. Интеграция Stop Request

Реализованный Manager не раскрывает operational Stop capability. Будущий capability contract, его lifetime и связь с shutdown orchestration определены DP-004. Manager не должен раскрывать raw Session, Context, Runtime, WebSocket, Send, Stop или произвольную callback capability через публичные identity views.

## 16. Wait

`Wait(ctx)` наблюдает единственный shutdown cycle Manager. Он не инициирует shutdown и не выполняет Stop requests.

В текущей identity-only реализации Wait возвращает nil, когда одновременно пусты:

- reservation accounting;
- committed registry accounting.

DP-004 добавляет один opaque Owner Lifetime Lease для каждого successful execution handoff. Commit атомарно публикует Registration, accounting lease, bound Stop capability и one-shot execution binding в одной irreversible point. До Commit ни одна из них не видима или eligible; после Commit ни одна не откатывается. После интеграции Wait возвращает nil только при пустых accounting reservation, committed registry и owner lifetime. Lease не изменяет semantics Complete, Lookup или Snapshot identity.

Mutation, удаляющая последний outstanding accounting item, автоматически переводит `Closing` в `Closed` и уведомляет всех Wait callers.

Если первым завершается context caller:

- Wait возвращает `ctx.Err()` этому caller;
- Manager остается в `Closing`, пока accounting не пуст;
- ни один record не удаляется ложно;
- последующий Wait продолжает наблюдать тот же cycle.

В этом proposal у Manager нет собственной operational failure model. Session errors не агрегируются и не сохраняются. BeginShutdown возвращает только результаты проверки контракта, а Wait после `Closed` возвращает nil вместо неопределенного terminal error Manager.

## 17. Порядок Shutdown Runtime

Manager предоставляет Runtime shutdown две разные операции: неблокирующий `BeginShutdown` и context-bounded `Wait`. Полный порядок со Stop requests per-Session, cancellation Runtime, shutdown Listener, terminal observation и accounting owner lifetime нормативно определён в DP-004.

DP-003 требует только, чтобы `BeginShutdown` не блокировал cancellation Runtime, а `Wait` никогда не выполнял Session I/O и не создавал фиктивный completion.

## 18. Ограничения Failure и Abandonment

Обработка execution panic и observation terminal result определены DP-004. Сам Manager не исполняет recoverable work и не агрегирует operational failures.

Сходимость accounting Manager не гарантируется после:

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

Manager не создаёт, не владеет, не заменяет и не отменяет корневой Runtime context. Host остаётся его единственным owner согласно [ADR-0004](../adr/0004-handshake-runtime-dependencies.md). DP-004 определяет, как per-Session execution производит и контролирует child cancellation. Только операции Manager изменяют accounting Manager.

## 21. Архитектурные инварианты

- У WebSocket всегда ровно один owner.
- Reservation handle имеет ровно один terminal accounting effect Commit-or-Abort.
- Commit регистрации — единственная linearization point видимости registry.
- Successful Commit необратим; committed Registration удаляет только Complete.
- BeginShutdown — единственная linearization point `Open -> Closing`.
- После BeginShutdown ни один Commit не завершается успешно.
- Complete — одна атомарная mutation `Registered -> Removed`.
- RegistrationID никогда не повторяется за срок жизни Manager.
- Stale Complete не может повлиять на другую регистрацию.
- Manager становится `Closed` только при пустых reservation и committed accounting; после интеграции DP-004 accounting owner lifetime также должен быть пуст.
- BeginShutdown не выполняет I/O, а Wait не выполняет Stop requests.
- Lookup возвращает только immutable identity metadata и никогда не продлевает lifetime.
- Manager не выполняет Session, не владеет transport, не маршрутизирует, не доставляет, не публикует Presence, не хранит history, не агрегирует diagnostics и не применяет limits.
- Инварианты per-Session execution определены DP-004 и не выводятся из records Manager.

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

Предлагаемая Manager-side модель:

```text
Handshake сохраняет ownership Upgrade
    -> Session handoff выполняет Reserve SessionID
    -> registration выполняет Commit
    -> per-Session execution выполняется согласно DP-004
    -> terminal execution вызывает Complete
```

Текущий Dispatcher не имеет Runtime-wide integration регистрации. Добавление регистрации Manager меняет внутренний accounting и обязано сохранить существующие transaction linearization points. DP-004 определяет отдельную migration execution.

## 26. Открытые вопросы

- Какой будущий design владеет configured Session limits?
- Какую identity экземпляра Runtime будущая Persistence объединит с SessionID?
- Какова operational response на accepted permanently blocked reservation или execution?

Open questions, специфичные для execution, поддерживаются в DP-004. Они не изменяют инварианты transaction и accounting Manager, определённые здесь.

## 27. Трассировка архитектурного ревью

| Finding | Статус | Решение |
| --- | --- | --- |
| F-01 | Resolved | BeginShutdown выполняет неблокирующий transition; Wait является отдельным context-bounded наблюдением. |
| F-02 | Resolved | Complete — одна атомарная mutation `Registered -> Removed` и единственная linearization point visibility/accounting. |
| F-03 | Resolved | Ownership Reservation остаётся непрерывным до Commit или Abort; execution-side transfer определён DP-004. |
| F-04 | Resolved в DP-004 | Dispatcher подготавливает ровно один dormant launch path; integrated boundary Commit атомарно публикует Registration и её one-shot execution binding до появления observable state. |
| F-05 | Delegated | Lifetime Stop capability определён DP-004; реализованный Snapshot остаётся identity-only. |
| F-06 | Delegated | Concurrency Stop request определён DP-004 и остаётся вне execution Manager. |
| F-07 | Clarified | Runtime shutdown orchestration находится вне Manager; DP-004 определяет её execution-side ordering. |
| F-08 | Resolved | RegistrationView включает никогда не повторяющийся RegistrationID и различает повторное использование SessionID. |
| F-09 | Resolved | RegistrationID уникален за срок жизни Manager и никогда не повторяется. |
| F-10 | Resolved | Неопределенный terminal error Manager удален; Wait после Closed возвращает nil. |
| F-11 | Clarified | Deferred completion гарантируется только для recoverable paths owned goroutine; nonrecoverable cases являются accepted limitations. |
| F-12 | Clarified | Lookup намеренно возвращает только metadata и не является operational capability Session. |
| F-13 | Resolved | Lookup возвращает immutable RegistrationView, но не raw Session или mutable execution capability. |
| F-14 | Resolved | Нормативное состояние `Completing` удалено. |
| F-15 | Delegated | Создание и lifetime Execution Owner определены DP-004. |
| F-16 | Clarified | SessionID opaque, сравнивается byte-for-byte, имеет длину до 255 bytes и не логируется raw автоматически. |
| F-17 | Clarified | Migration явно признает существенные изменения внутреннего lifecycle и concurrency. |
| F-18 | Clarified | Ответственность Manager ограничена явными negative invariants. |
| F-19 | Clarified | Exactly-once описывает effective accounting mutation, а не доставку notification. |
| F-20 | Accepted limitation | Permanent abandonment может сохранять правдивый accounting и удерживать Manager в Closing. |
| F-21 | Deferred | Limits выделены в будущую сфокусированную Configuration/admission capability. |
| F-22 | Resolved | BeginShutdown — явный неблокирующий contract `Open -> Closing`. |
| F-23 | Resolved | Complete атомарно удаляет visibility и accounting wait set. |
| F-24 | Resolved | Reservation handle является непрерывным owner до Commit или Abort. |
| F-25 | Delegated | Реализованный identity Snapshot остаётся detached; будущий lifetime Stop capability определён DP-004. |
| F-26 | Resolved | Manager не агрегирует terminal errors; Wait возвращает nil или caller `ctx.Err()`. |
| F-27 | Resolved | Immutable view содержит RegistrationID, поэтому stale observation и повторное использование SessionID различимы. |

## 28. Готовность к утверждению

Lifecycle Session Manager, registration identity, transaction linearization, lookup, shutdown accounting и semantics identity snapshot полностью определены здесь. Findings execution ownership нормативно делегированы DP-004, а не дублируются.

Accepted limitations определены явно: завершение процесса, unrecoverable failure, permanently blocked goroutine или нарушение внешнего контракта могут не позволить accounting стать пустым. Manager правдиво остается в `Closing`, а не объявляет ложный completion.

**Кандидат решения:** Approved with Findings.

Документ сохраняет статус Draft. DP-004 должен пройти независимое ревью до начала реализации Execution Owner.

## 29. Ссылки

- [ARCH-001: Runtime Architectural Pattern](../architecture/ARCH-001-runtime-architectural-pattern.md)
- [ARCH-002: Runtime Foundation Freeze](../architecture/ARCH-002-runtime-foundation-freeze.md)
- [DP-001: Runtime Handshake Pipeline](DP-001-runtime-handshake-pipeline.md)
- [DP-002: Runtime Host Composition Root](DP-002-runtime-host-composition-root.md)
- [DP-004: Per-Session Execution Boundary](DP-004-per-session-execution-boundary.md)
- [ADR-0004: Handshake Runtime Dependency Boundary](../adr/0004-handshake-runtime-dependencies.md)
- [Master Engineering Plan](../roadmap/MASTER_PLAN.md)

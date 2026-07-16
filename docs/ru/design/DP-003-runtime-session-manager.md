# DP-003: Runtime Session Manager

[English version](../../en/design/DP-003-runtime-session-manager.md)

## 1. Статус

**Статус:** Draft

Эта редакция упрощает lifecycle после второго независимого архитектурного ревью. Она отделяет shutdown transition от ожидания, сводит completion к одной атомарной mutation, определяет непрерывный ownership каждой registration transaction и задаёт концептуальный контракт Execution Owner, необходимый для реализации. Точные имена Go остаются деталями реализации; production boundary, ownership, ordering и concurrency semantics являются нормативными.

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
- Определить один per-Session Execution Owner, его контракт launch и terminal completion, а также стабильную non-owning Stop-request capability.
- Ограничить Manager registry и shutdown accounting.

## 4. Не-цели

Документ не проектирует Router, Delivery, Presence, Groups, Persistence, cluster ownership, Plugins, Metrics, diagnostics framework, policy лимитов Session, transport protocol, публичный API, Configuration fields, generic Session framework или supervision framework.

## 5. Границы ownership

Четыре области ownership остаются разделенными.

### Ownership WebSocket Transport

Upgrade boundary владеет WebSocket и дочерним connection context, пока Session явно их не примет. До принятия Upgrade boundary является единственным closer. После принятия Session становится единственным closer, в том числе при ошибке последующей регистрации.

### Ownership Registration Transaction

Успешный Reserve возвращает один концептуальный reservation handle. Control flow Session handoff/Dispatcher владеет этим handle до передачи per-Session Execution Owner. Затем Execution Owner владеет им до ровно одной terminal operation:

- `Commit`, который поглощает reservation и создает committed record;
- `Abort`, который удаляет reservation без видимости registry.

Нормативный pattern control flow:

```text
reservation := Manager.Reserve(SessionID)
defer reservation.AbortUnlessCommitted()
```

Эта запись описывает ownership obligation, а не обязательный Go API. Каждый recoverable return, rejection, error или panic до Commit обязан выполнить Abort. Manager учитывает reservation, но не владеет goroutine, обязанной выполнить Commit или Abort. Handshake никогда не владеет reservation и не импортирует Session Manager; post-Upgrade preparation делегируется границе Session handoff.

### Ownership выполнения Session

Узкий per-Session Execution Owner является отдельным production object, который Session handoff/Dispatcher создаёт после создания Session и успешного Reserve, но до Commit регистрации. Dispatcher является его единственным creator в production composition path. Он передаёт owner reservation handle, одну Session, execution context, производный от Runtime, узкую completion capability и узкий observer terminal result.

До Commit owner неактивен и не владеет committed accounting. После передачи он владеет reservation transaction. Успешный Commit является единственной точкой, где owner получает committed RegistrationID и активирует terminal completion. Следующая непосредственно за ним сериализованная активация owner является единственной execution-ownership point: она передаёт ответственность за execution Session и cleanup transport до наблюдаемого успешного результата handoff. Затем owner владеет Session `Start/Run/Stop` в одной owned goroutine до завершения terminal cleanup и completion accounting.

Execution Owner зависит только от lifecycle contract Session, opaque registration identity, собственного context и узких contracts completion и terminal observer. Он не должен импортировать или удерживать concrete Manager, Runtime Host, Listener, Handshake, HTTP, WebSocket, Shutdown Snapshot, Router, Delivery, Presence или Persistence. Он не использует service locator, DI через `context.Value`, generic supervisor или generic worker framework.

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
Upgrade boundary вызывает Session handoff/Dispatcher
    -> создать provisional Session без передачи transport ownership
    -> Reserve SessionID
    -> создать неактивный per-Session Execution Owner
    -> передать ownership Reservation Execution Owner
    -> owned goroutine устанавливает terminal guard
    -> Commit registration со стабильной Stop-request capability
    -> получить RegistrationID и активировать completion obligation
    -> принять ownership execution Session и cleanup transport
    -> сообщить об успешном handoff
    -> Start
    -> Run
    -> Stop
    -> Complete
    -> передать один terminal result
```

### Reserve

Reserve разрешен только пока Manager находится в `Open`. Он валидирует SessionID и создает:

- reservation handle с исключительным обязательством Commit-or-Abort;
- RegistrationID, уникальный за весь срок жизни этого Manager.

Reservation учитывается для закрытия Manager, но невидима Lookup и не является committed Session в shutdown registry.

### Commit

Commit поглощает reservation и атомарно создаёт один committed record с immutable identity metadata и стабильной Stop-request capability. Эта mutation является единственной registration linearization point: видимость registry и membership committed wait set появляются вместе. Успешный Commit возвращает opaque RegistrationID той же owned goroutine. Owner активирует completion obligation до того, как Start или успешный результат handoff могут стать наблюдаемыми.

### Abort

Abort атомарно удаляет uncommitted reservation и уведомляет Wait callers, если accounting теперь может быть пуст. Abort никогда не создает lookup visibility или committed record. Повторный Abort после Commit или Abort не имеет второго accounting effect.

### Transport Handoff

Создание Session является provisional и не передаёт transport ownership. Upgrade boundary остаётся единственным closer transport во время создания Session, Reserve, создания Execution Owner и неуспешного Commit. При любой такой pre-Commit ошибке Execution Owner выполняет Abort, не вызывает Complete и возвращает failed handoff; Upgrade boundary закрывает transport.

Успешный Commit активирует terminal completion obligation owner. Затем controlled handoff выполняет одну сериализованную активацию owner, которая передаёт ownership execution Session и cleanup transport границе Execution Owner/Session до того, как Dispatcher сообщит об успешном принятии. Ошибка после Commit, но до этой активации, приводит к Complete, пока transport по-прежнему закрывает Upgrade boundary. Начиная с активации каждая ошибка обрабатывается owner: он останавливает Session, выполняет Complete и не позволяет Upgrade boundary повторно закрыть transport. Состояние «transport принят, но Commit завершился ошибкой» намеренно недостижимо. В каждый момент существует ровно один transport owner.

## 8. Linearization Reserve/Commit/BeginShutdown

Документ выбирает строгую shutdown model.

На одной synchronization boundary Manager:

- либо Commit завершается до `BeginShutdown`, становится committed record и входит в shutdown snapshot;
- либо побеждает `BeginShutdown`, Manager переходит в `Closing`, а каждая uncommitted reservation теряет право на Commit.

Новый Reserve после `Closing` отклоняется. Существующая reservation, чей Commit проиграл, должна выполнить Abort через Execution Owner. Transport ownership в этой точке ещё не передан, поэтому Complete и Session Stop не вызываются; cleanup transport выполняет Upgrade boundary. Manager не выполняет forced Abort handle, поскольку не владеет этим control flow.

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

Session handoff/Dispatcher создаёт ровно один Execution Owner для одной provisional Session и передаёт ему один Reservation handle. Owner создаёт ровно одну owned execution goroutine. Вызов handoff ожидает только сообщения этой goroutine о pre-Commit failure или успешном committed ownership acceptance; он не ожидает завершения Session.

Перед вызовом Commit owned goroutine устанавливает один terminal guard. До успешного Commit этот guard может только выполнить Abort Reservation. Успешный возврат Commit является linearization point перехода owner из pre-Commit в committed: он предоставляет RegistrationID, делает обязательство Abort no-op и активирует terminal Complete. Следующая сериализованная активация является execution-ownership linearization point. До активации Start и успешное ownership acceptance не наблюдаются.

Нормативный committed algorithm:

```text
активировать completion obligation
    -> принять ownership execution и cleanup transport
    -> Start(Session, execution context)
    -> если Start успешен и Stop request не победил, Run(Session, execution context)
    -> попытаться выполнить Stop(Session, owner-controlled cleanup context)
    -> один раз вызвать Complete(RegistrationID)
    -> передать один terminal result внедрённому observer
```

Stop предпринимается ровно один раз на каждом committed terminal path, включая ошибку Start, возврат Run, ошибку Run, cancellation и recovered panic. Ошибка Stop сохраняется в terminal result, но никогда не подавляет Complete. Complete предпринимается после попытки Stop; поэтому навсегда заблокированный Stop может помешать completion и остаётся accepted abandonment limitation. Cleanup context принадлежит Execution Owner и не является уже отменённым execution context.

Owner остаётся жив, пока execution не достигнет recoverable terminal path, единственный вызов Complete не вернёт управление, terminal result не будет один раз предложен observer и каждый Stop-request invocation, линеаризованный до terminal completion, не вернёт управление. Terminal observer является узким non-owning sink, который Runtime composition передаёт Dispatcher; он не получает Session или transport. Ошибка или panic observer изолируются и не могут помешать попытке Complete. Конкретный тип terminal result и diagnostics backend остаются деталями реализации, но категории Start, Run, Stop, recovered panic и completion anomaly должны оставаться различимыми.

### Stop до Start

Стабильная Stop-request capability записывает один termination request и отменяет owner-controlled execution context. Если request линеаризуется до того, как owner пометит Start начавшимся, Start и Run пропускаются. Owned goroutine вызывает Session Stop из `Created`, затем выполняет Complete. Альтернативного поведения «Start всё же может выполниться» нет.

### Concurrent Start и Stop

Control state сериализует Execution Owner, а не Session Manager. Linearization point Start — transition owner, помечающий Start начавшимся после проверки отсутствия принятого Stop request. Linearization point Stop request — первая атомарная запись termination intent.

- Если побеждает Stop, Start и Run запрещены.
- Если побеждает Start, request отменяет execution context; после возврата Start выполнение Run начинается только при отсутствии Stop request и активном context.
- Если Run активен, cancellation завершает его согласно контракту Session; затем owner вызывает Stop.
- Повторные и concurrent Stop requests не создают дополнительных обязательств Stop или Complete.
- Lifecycle lock не удерживается во время Start, Run, Stop, Complete или вызова terminal observer.

### RegistrationID и Completion Capability

Execution Owner может существовать без committed RegistrationID, но только в неактивной pre-Commit фазе. Owned goroutine получает RegistrationID только от успешного Commit. Failed Commit не активирует completion obligation: owner выполняет Abort Reservation, сообщает о failed handoff и никогда не вызывает Complete.

Owner зависит от consumer-oriented completion capability с концептуальной операцией `Complete(RegistrationID) bool`, а не от concrete Manager. Runtime composition предоставляет Manager-backed реализацию Session handoff/Dispatcher, который внедряет её в каждый owner. Capability активируется только после успешного Commit. Каждый owner вызывает её один раз. `true` означает, что этот вызов выполнил effective mutation `Registered -> Removed`; `false` означает, что известный committed record не был изменён, и передаётся как terminal accounting anomaly. Повторной попытки нет. Manager по-прежнему гарантирует не более одной effective mutation для RegistrationID.

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

Целевой shutdown snapshot содержит immutable identity Registration и одну стабильную non-owning Stop-request capability для каждого committed record. В нём нет raw Session, WebSocket, Handler, Execution Owner или mutable registry record. Текущий реализованный identity-only Snapshot намеренно остаётся без изменений до задачи интеграции Execution Owner и capability.

Removal через Complete не инвалидирует capability, уже выданную текущей shutdown orchestration. Ее lifetime продолжается до возврата соответствующего Stop request. Повторный BeginShutdown не создает новый cycle и не дублирует accounting.

## 15. Stop-Request Capabilities

Stop-request capability представляет разрешение запросить завершение у одного Execution Owner без получения ownership Session. Её концептуальная операция — `RequestStop() bool`; она не принимает context, поскольку строго неблокирующая и не выполняет Session I/O в goroutine caller.

Первый вызов, линеаризованный до terminal completion, записывает termination intent, один раз отменяет owner-controlled execution context и возвращает `true`. Повторные или concurrent вызовы, а также вызовы, линеаризованные после terminal completion, являются стабильными no-op и возвращают `false`. Операция не ожидает Start, Run, Stop, Complete или полного завершения Session и не возвращает Session result.

Capability никогда не раскрывает Session, Send, cancellation Context, Runtime, Listener, WebSocket, регистрацию callback или mutable state owner. Stop requests разных registrations независимы и не могут сериализоваться через Manager. Вызов capability не получает lifecycle lock Manager.

Каждый invocation удерживает control cell живой до своего возврата. Complete может выполняться, пока ранее вошедший RequestStop возвращает управление, но не может инвалидировать этот invocation. При terminal completion control cell отсоединяется от execution state Session; последующие вызовы остаются безопасными no-op с результатом `false` и не удерживают или повторно не получают ownership Session. После terminal completion новые Stop requests не принимаются.

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

Execution Owner выполняет recover для panic внутри owned execution boundary. Recovery не классифицирует execution как успешный и не выполняет повторный panic из owned goroutine. Terminal guard отменяет execution context, предпринимает Session Stop при активированном committed ownership, один раз вызывает Complete при committed RegistrationID и передаёт один sanitized terminal result категории panic внедрённому observer. До успешного Commit тот же guard выполняет Abort и не вызывает Complete или Session Stop, поскольку transport ownership ещё не передан.

Panic или error из Stop либо terminal observer не могут подавить уже установленную попытку Complete; каждый terminal step защищён независимо. Observer не получает panic payload, способный раскрыть credentials или transport metadata. Это правило определяет только ownership cleanup и не создаёт diagnostics framework.

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

Manager не создаёт, не владеет, не заменяет и не отменяет корневой Runtime context. Host остаётся его единственным owner согласно [ADR-0004](../adr/0004-handshake-runtime-dependencies.md). Session handoff создаёт execution context каждого owner из Runtime-owned connection context. Owner управляет только cancellation своего child context; root cancellation и принятый Stop request отменяют execution. Accounting Manager меняет только Complete.

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
- Session handoff/Dispatcher является единственным production creator одного Execution Owner для каждой committed Session.
- Успешный Commit активирует ровно одно completion obligation owner до наблюдаемого Start или успешного handoff.
- Execution Owner является единственным caller Session Start, Run и Stop после принятия committed ownership.
- Stop request, победивший до Start, запрещает Start и Run.
- Одна owned goroutine выполняет одну попытку Stop и один вызов Complete на каждом recoverable committed terminal path.
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
Handshake сохраняет ownership Upgrade
    -> Session handoff/Dispatcher создаёт provisional Session
    -> Dispatcher выполняет Reserve SessionID
    -> Dispatcher создаёт per-Session Execution Owner
    -> owner получает Reservation и устанавливает terminal guard
    -> registration выполняет Commit со стабильной Stop-request capability
    -> owner принимает ownership execution и cleanup transport
    -> Dispatcher сообщает об успешном handoff
    -> goroutine owner выполняет Start/Run/Stop и Complete
```

Обязанности текущего Dispatcher меняются явно: он сохраняет создание Session и координацию handoff, получает Reserve и создание owner, а все post-Commit Start/Run/Stop и terminal Complete делегирует per-Session owner. После успешной передачи ownership он больше не блокируется синхронно в Session Run и не вызывает Session Stop напрямую.

Внешнее WebSocket behavior сохраняется, но внутренние concurrency, ownership, registration и shutdown semantics меняются. Поэтому миграцию следует считать lifecycle work, а не behavior-neutral refactor, и проводить race review каждой границы.

## 26. Открытые вопросы

- Какие точные Go-имена и package-private interfaces представляют нормативные consumer-oriented capabilities?
- Какая форма terminal-result value сохраняет phase и cleanup errors для узкого observer?
- Какая policy cleanup deadline должна ограничивать Session Stop без изменения существующей семантики Session?
- Какой будущий design владеет configured Session limits?
- Какую identity экземпляра Runtime будущая Persistence объединит с SessionID?
- Какова operational response на accepted permanently blocked reservation или execution?

Creator, начало ownership, launch model, pre-Commit cleanup, поведение Stop-before-Start, ordering concurrent Start/Stop, completion dependency, cleanup panic, операция Stop request и lifetime owner больше не являются открытыми вопросами. Оставшиеся вопросы относятся к representation уровня реализации или явно отдельной subsystem policy; они не изменяют определённые здесь linearization points или accounting Manager.

## 27. Трассировка архитектурного ревью

| Finding | Статус | Решение |
| --- | --- | --- |
| F-01 | Resolved | BeginShutdown выполняет неблокирующий transition; Wait является отдельным context-bounded наблюдением. |
| F-02 | Resolved | Complete — одна атомарная mutation `Registered -> Removed` и единственная linearization point visibility/accounting. |
| F-03 | Resolved | Reservation handle один раз переходит от Session handoff/Dispatcher к Execution Owner и сохраняет непрерывный ownership Commit-or-Abort. |
| F-04 | Resolved | Execution Owner существует до Commit; shutdown между Commit и Start — штатный terminal path, завершающийся Stop и Complete. |
| F-05 | Resolved | Shutdown snapshot содержит стабильные Stop capabilities, lifetime вызова которых переживает Complete. |
| F-06 | Resolved | Stop requests независимы и строго неблокирующие; Manager их не вызывает. |
| F-07 | Resolved | Runtime shutdown orchestration владеет Stop requests; Manager прогрессирует через Abort/Complete mutations без ожидающего caller. |
| F-08 | Resolved | RegistrationView включает никогда не повторяющийся RegistrationID и различает повторное использование SessionID. |
| F-09 | Resolved | RegistrationID уникален за срок жизни Manager и никогда не повторяется. |
| F-10 | Resolved | Неопределенный terminal error Manager удален; Wait после Closed возвращает nil. |
| F-11 | Clarified | Deferred completion гарантируется только для recoverable paths owned goroutine; nonrecoverable cases являются accepted limitations. |
| F-12 | Clarified | Lookup намеренно возвращает только metadata и не является operational capability Session. |
| F-13 | Resolved | Lookup возвращает immutable RegistrationView, но не raw Session или mutable execution capability. |
| F-14 | Resolved | Нормативное состояние `Completing` удалено. |
| F-15 | Resolved | Execution Owner создаётся Session handoff до Commit и переживает completion и outstanding Stop invocations. |
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

## 28. Трассировка blocker findings TASK-B4-007B

| Finding | Статус | Решение |
| --- | --- | --- |
| B-01 — Creator boundary | Resolved | Session handoff/Dispatcher является единственным production creator одного per-Session Execution Owner. |
| B-02 — Начало ownership | Resolved | Успешный Commit активирует completion; следующая сериализованная активация owner является единственной execution-ownership point до наблюдаемого Start или успешного handoff. |
| B-03 — Pre-Commit cleanup | Resolved | Owner выполняет Abort Reservation; Upgrade boundary сохраняет и закрывает transport; Stop и Complete не вызываются. |
| B-04 — Launch model | Resolved | Одна owned goroutine выполняет Commit activation и нормативный lifecycle Start/Run/Stop; handoff ожидает только outcome commit/acceptance. |
| B-05 — Активация completion | Resolved | Успешный Commit возвращает RegistrationID и активирует ровно одно terminal obligation Complete; failed Commit его не активирует. |
| B-06 — Panic semantics | Resolved | Recoverable panic owned boundary восстанавливается, предпринимаются cleanup и Complete, наблюдается один sanitized terminal result, повторный panic не выполняется. |
| B-07 — Stop до Start | Resolved | Stop request, победивший до Start, вызывает Stop из Created и запрещает Start и Run. |
| B-08 — Concurrent Start/Stop | Resolved | Owner сериализует control state; первый Stop один раз записывает termination, Start побеждает только в явной linearization point, Run не начинается после принятого Stop. |
| B-09 — Передача RegistrationID | Resolved | Owner существует неактивным без RegistrationID и получает opaque identity только от успешного Commit. |
| B-10 — Completion capability | Resolved | Owner использует внедрённую семантику `Complete(RegistrationID) bool` и никогда не зависит от concrete Manager. |
| B-11 — Stop-request capability | Resolved | `RequestStop() bool` является неблокирующей, идемпотентной, не принимает context, не раскрывает Session и имеет стабильное terminal no-op поведение. |
| B-12 — Lifetime owner | Resolved | Owner живёт до terminal cleanup, одного вызова Complete, одного terminal observation и возврата всех Stop invocations, вошедших до terminal completion. |

Ни один blocker finding TASK-B4-007B не оставлен Deferred.

## 29. Готовность к утверждению

Blocker findings второго ревью F-01–F-03 закрыты разделением shutdown contracts, атомарным Complete и reservation handles с непрерывным ownership. High findings F-04–F-10 закрыты lifecycle execution owner, стабильными Stop capabilities, независимой shutdown orchestration, identity-safe views и удалением неопределенных terminal errors Manager.

Accepted limitations определены явно: завершение процесса, unrecoverable failure, permanently blocked goroutine или нарушение внешнего контракта могут не позволить accounting стать пустым. Manager правдиво остается в `Closing`, а не объявляет ложный completion.

Все blocker decisions TASK-B4-007B закрыты нормативно. Точные Go-имена, private package placement внутри подсистемы Session handoff, representation terminal result и механика cleanup deadline остаются implementation-level choices, ограниченными этим контрактом. Configured limits, durable Runtime identity и operational обработка permanently blocked work остаются вопросами отдельных подсистем и не изменяют этот lifecycle.

**Кандидат решения:** Approved with Findings.

Документ сохраняет статус Draft, а контракт Execution Owner готов к targeted independent review. Ревью должно подтвердить creator boundary, transition Commit-to-ownership, launch в owned goroutine, ordering Stop, completion capability, cleanup recoverable panic и lifetime стабильной capability.

## 30. Ссылки

- [ARCH-001: Runtime Architectural Pattern](../architecture/ARCH-001-runtime-architectural-pattern.md)
- [ARCH-002: Runtime Foundation Freeze](../architecture/ARCH-002-runtime-foundation-freeze.md)
- [DP-001: Runtime Handshake Pipeline](DP-001-runtime-handshake-pipeline.md)
- [DP-002: Runtime Host Composition Root](DP-002-runtime-host-composition-root.md)
- [ADR-0004: Handshake Runtime Dependency Boundary](../adr/0004-handshake-runtime-dependencies.md)
- [Master Engineering Plan](../roadmap/MASTER_PLAN.md)

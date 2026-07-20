# DP-004: Per-Session Execution Boundary

[English version](../../en/design/DP-004-per-session-execution-boundary.md)

## 1. Статус

**Статус:** Approved

Этот утверждённый design определяет per-Session execution boundary после успешного WebSocket Upgrade. Он разделяет синхронную activation, ownership transport, асинхронное execution Session, terminal cleanup и shutdown accounting Runtime без изменения замороженного lifecycle Runtime Host.

## 2. Постановка проблемы

Текущий Session Dispatcher создаёт Session с уже переданным WebSocket и синхронно выполняет `Start`, `Run` и `Stop` в goroutine HTTP handler. Такая модель не поддерживает asynchronous execution, принадлежащее Runtime:

- handler остаётся заблокированным весь lifetime Session;
- у Runtime нет стабильной per-Session Stop-request capability;
- Session Manager не может учитывать execution до terminal observation;
- cancellation connection context принадлежит defer Dispatcher, который исчезает после перехода Dispatcher к asynchronous model;
- completion регистрации может произойти раньше завершения Runtime-owned work.

[DP-003](DP-003-runtime-session-manager.md) определяет registration identity и accounting. Этот документ задаёт execution boundary и её узкую интеграцию с этими контрактами.

## 3. Цели

- Сохранить ровно одного owner WebSocket и одного owner connection cancellation.
- Определить одного синхронного activation executor.
- Определить один per-Session Execution Owner и одну owned goroutine.
- Сделать Start, Run, termination, cleanup и completion линейризуемыми.
- Предоставить shutdown Runtime стабильную non-owning Stop-request capability.
- Не передавать Session Manager знания о Session и transport behavior.
- Гарантировать покрытие успешным Manager Wait полного tracked lifetime owner.

## 4. Scope

Документ проектирует только:

- ownership после WebSocket Upgrade;
- transport-independent provisional construction;
- package-private attachment transport;
- создание, launch, activation и lifecycle Execution Owner;
- Session Start, Run и terminal cleanup;
- acknowledgement cancellation connection context;
- completion, terminal observation, Stop requests, интеграцию Snapshot capability и accounting lifetime owner.

Документ не проектирует Router, Delivery, Persistence, Presence, Groups, Diagnostics infrastructure, Session limits, Cluster behavior, Configuration fields, публичный management API или generic supervision framework.

## 5. Существующие ограничения

Design сохраняет:

- принципы ownership и explicit execution из [ARCH-001](../architecture/ARCH-001-runtime-architectural-pattern.md);
- замороженные lifecycle Host, Runtime context, Admission Gate, startup, rollback, readiness и ownership Listener из [ARCH-002](../architecture/ARCH-002-runtime-foundation-freeze.md);
- pre-Upgrade Authentication и ownership Upgrade boundary из [DP-001](DP-001-runtime-handshake-pipeline.md);
- направление composition из [DP-002](DP-002-runtime-host-composition-root.md);
- read-only Runtime capability boundary из [ADR-0004](../adr/0004-handshake-runtime-dependencies.md);
- semantics DP-003 Reserve, Commit, Complete, Lookup, registration accounting и identity Snapshot.

Текущая реализация Session требует migration, поскольку её constructor требует WebSocket, а state `Created` всегда владеет transport. Эта implementation detail не считается целевым provisional contract.

## 6. Выбранная архитектура

Dispatcher является единственным handoff executor:

```text
Upgrade boundary входит в Dispatcher
    -> Reserve SessionID
    -> создать transport-independent Session Core
    -> создать Execution Owner
    -> создать ровно одну dormant launch goroutine, ожидающую свой one-shot Commit gate
    -> provisionally attach transport и cancellation
    -> активировать owner в PreCommit без запуска execution Session
    -> Commit
       -> под одной synchronization boundary Manager:
          -> опубликовать Registration, accounting lease и Stop capability
          -> передать ownership Session, WebSocket, cancellation и execution
          -> перевести one-shot Commit gate из PreCommit в Committed
       -> освободить synchronization boundary только после завершения всей publication
    -> немедленно вернуть accepted=true

Dormant launch goroutine
    -> не может пройти Commit gate до successful Commit
    -> становится owned execution path, когда gate достигает Committed
    -> установить observation Runtime cancellation с race-safe проверкой после регистрации
    -> Start
    -> Run, если допустим
    -> Terminalizing
    -> Session Cleanup
    -> Complete
    -> Terminal Observer
    -> закрыть admission control calls
    -> UnregisterAndDrain callback Runtime cancellation и вошедшие control calls
       -> если отсутствие callback не подтверждено: остаться в Terminalizing с активным lease
    -> запечатать detached Stop control cell
    -> Terminal
    -> освободить Owner Lifetime Lease
    -> вернуться без Runtime-owned epilogue
```

Commit является единственной irreversible publication point как Registration, так и eligibility execution. Каждая operation до Commit provisional и может завершиться через Abort и cleanup Dispatcher. Ни одна recoverable operation после Commit не может вернуть `accepted=false` или восстановить pre-Commit state.

«Owner started» имеет одно нормативное значение: one-shot Commit gate достиг `Committed`, поэтому уже существующая dormant launch goroutine получила право войти в lifecycle owner. Scheduler execution, `Session.Start` и `Session.Run` являются последующими effects и не являются publication points.

## 7. Домены Ownership

### Upgrade Boundary

Ownership начинается после успешного `websocket.Accept`. Dispatcher, действуя от имени Upgrade boundary, исключительно владеет WebSocket, connection cancellation и pre-handoff cleanup.

Ownership заканчивается только при successful Commit. Commit является единственной linearization point ownership transport и publication Registration. До него Dispatcher является единственным closer и owner cancellation. После него Dispatcher никогда не закрывает и не отменяет эти resources.

### Session Core и Session

Session Core владеет только immutable identity, Principal, metadata и Message Handler. Он не является Session, не раскрывает `Start`, `Run`, `Send` или `Stop` и не владеет transport.

До Commit один Core может быть provisionally сформирован в полностью configured Session и связан с dormant owner и одной dormant launch goroutine, но Dispatcher сохраняет ownership transport, provisional launch obligation и cleanup. Goroutine может наблюдать только свой private Commit gate и не может вызывать lifecycle Session. Successful Commit публикует committed identity, передаёт подготовленные Session, WebSocket, connection cancellation и execution responsibility и делает ту же goroutine eligible. До Commit Runtime-owned execution Session отсутствует.

### Reservation и Registration

Dispatcher владеет obligation Reservation до Commit или Abort. До Commit Registration не существует, Lookup не может её наблюдать, Snapshot не может её содержать, а accounting owner lifetime её не учитывает.

Commit потребляет Reservation и необратимо публикует committed identity, owner-only completion, owner lifetime, Stop capability Snapshot и acceptance ownership. После Commit Registration может исчезнуть только через normal Complete.

### Execution Owner

Construction и pre-Commit activation owner не получают ownership Manager или transport. Owner остаётся dormant в `PreCommit`; Session Start и Run запрещены. Dispatcher создаёт ровно одну launch goroutine и ровно один one-shot Commit gate для этого owner. Successful Commit передаёт owner attached Session, committed bundle и execution responsibility, переводя этот gate в `Committed` внутри той же publication boundary. От Commit до Terminal этот единственный path является sole production caller Session `Start`, `Run` и terminal cleanup.

### Session Manager

Manager владеет accounting reservation, identity и visibility registration, публикацией Snapshot, mutation completion и accounting Owner Lifetime Lease. Он не владеет Session, WebSocket, Context, goroutine или Terminal Result.

### Terminal Result

Owner создаёт один immutable категоризированный Terminal Result и синхронно передаёт его одному вызову observer. Observer не владеет resource или lifecycle capability.

## 8. Connection Cancellation и Cleanup Acknowledgement

Activation оборачивает derived connection context в одну узкую идемпотентную cancellation cell, принадлежащую вместе с WebSocket. Cell использует фактический Runtime-derived context и его private cancellation function; ни одно из этих значений не раскрывается owner или observer.

Терминология cancellation:

- **invocation** — попытка вызвать cancellation cell; repeated invocation допустим;
- **effective cancellation** — единственный transition, после которого derived context наблюдаемо отменён;
- **acknowledgement** — immutable output cleanup, подтверждающий effective canceled state.

Первая effective cancellation один раз изменяет state context. Repeated invocations являются безопасными no-op. Terminal requirement относится к effective canceled state, а не к ровно одному invocation.

Attached Session раскрывает owner одну синхронную terminal operation:

```text
Cleanup(cleanupContext) CleanupResult
```

`CleanupResult` является immutable и содержит категоризированный outcome закрытия transport, acknowledgement effective cancellation и категорию cleanup panic. Он не содержит raw transport, Context, callback или arbitrary error.

Session Cleanup сам является panic-safe boundary. Внутренние steps close, join и cancellation могут panic, но wrapper Cleanup recover-ит каждый внутренний panic и всегда возвращает `CleanupResult`. Outward panic из contract Cleanup запрещён.

Session Cleanup:

1. выполняет существующую работу Stop/закрытия transport;
2. всегда вызывает cancellation cell через внутренний final guard;
3. наблюдает derived context в canceled state;
4. возвращает acknowledgement только после такого наблюдения;
5. записывает sanitized categories для каждой recovered internal anomaly.

Repeated Cleanup возвращает тот же detached terminal result и не выполняет второе effective close или cancellation. Cleanup не раскрывает partial mutable state.

Cancellation cell создаётся только из стандартной Runtime-derived cancellation primitive. Safe invocation изолирует internal cancellation panic, выполняет private cancellation operation и подтверждает canceled state. Невозможность подтвердить cancellation является architecture anomaly в `CleanupResult`; owner не освобождает lease до возврата Cleanup. Owner доказывает canceled state только по возвращённому acknowledgement и никогда не обращается к private cancellation cell.

## 9. Модель Provisional Session

Выбрана модель **transport-independent Session Core с provisional формированием Session до Commit**.

Core не является lifecycle object. Поэтому:

- Start, Run, Send, Stop и Cleanup до provisional formation невозможны по type boundary;
- transport-free state `Created` отсутствует;
- ни одна операция не может обратиться к nil transport;
- disposal unattached Core не требует lifecycle cleanup.

Package-private preparation operation концептуально выглядит так:

```text
prepareOwner(
    core,
    WebSocket,
    ConnectionCancellation,
) (PreCommitOwner, error)
```

Её может вызывать только Dispatcher. Production composition допускает ровно одного caller и запрещает concurrent invocation. Operation не является публичным API attachment или transfer ownership.

Contract:

- один Core может быть подготовлен один раз;
- формирование Session, установка control cell owner и подготовка binding Stop завершаются до Commit;
- Dispatcher создаёт ровно одну dormant launch goroutine и связывает её ровно с одним one-shot Commit gate;
- Dispatcher сохраняет ownership WebSocket и cancellation на протяжении preparation;
- prepared owner остаётся в `PreCommit`, а launch goroutine не может пройти свой gate или выполнять lifecycle work Session;
- race со Start, Run или Cleanup до Commit невозможен;
- operation выполняет bounded in-memory construction без network I/O;
- validation failure или recovered panic оставляет Registration отсутствующей, а оба переданных resource — у Dispatcher;
- каждый terminal pre-Commit outcome ровно один раз публикует non-committed outcome gate, ожидает возврата dormant goroutine, освобождает owner-local prepared values и затем выполняет Abort до cleanup transport;
- до Commit registration callback Runtime cancellation не существует;
- до Commit lease и capability Snapshot не существуют;
- Start, Run, Stop и Cleanup могут начаться только после successful Commit.

Prepared Session содержит values, необходимые существующему transport-owning state `Created`, но ownership и lifecycle authority остаются у Dispatcher до Commit. При successful Commit Session становится Runtime-owned в `Created`; затем применяются существующие semantics `Created -> Running -> Stopping -> Stopped`.

## 10. Creator и Composition Boundary

Runtime Host остаётся production composition root. Он создаёт:

- Session Manager;
- узкий activation/registration adapter;
- factory Execution Owner;
- Terminal Observer;
- Dispatcher.

Execution Owner получает только:

- owner control cell, Runtime-derived execution context и read-only input observation root Runtime context;
- lifecycle-only Session после activation;
- owner-scoped Completion Adapter;
- owner-scoped Lifetime Lease;
- один Terminal Observer;
- immutable RegistrationID.

Owner не импортирует concrete Runtime Host, Manager, Listener, Handshake, HTTP, WebSocket library, Runtime Snapshot, Router, Delivery, Presence или Persistence.

Service locator, global registry, reflection, DI через `context.Value`, generic executor pool и generic supervisor не вводятся.

## 11. Preparation Owner и Start Boundary

Construction owner синхронен. Dispatcher сохраняет ownership Reservation, WebSocket, cancellation и cleanup.

Pre-Commit activation подготавливает:

1. полностью сформированную, но ещё не Runtime-owned Session;
2. control cell owner и causal termination state;
3. binding Stop-publication, необходимый Commit;
4. immutable read-only input observation root Runtime context, необходимый будущему owned execution path;
5. ровно одну dormant launch goroutine и её private one-shot Commit gate.

После этого owner находится в `PreCommit`. Launch goroutine существует, но остаётся provisional control flow во владении Dispatcher и не может пройти Commit gate. Она не вызывает Session Start, Run, Stop, Cleanup, Complete, observer или release lease.

Dispatcher владеет всей pre-Commit transaction через единую panic-safe orchestration boundary, начинающуюся не позднее создания dormant path. Обычная error, context cancellation, проигрыш Commit операции BeginShutdown, panic attachment или preparation и panic непосредственно перед Commit используют одну terminal sequence:

1. ровно один раз опубликовать non-committed outcome gate;
2. дождаться, пока dormant path наблюдает outcome и вернётся;
3. освободить все owner-local prepared values;
4. выполнить Abort;
5. сохранить ownership transport и connection cancellation у Dispatcher;
6. вернуть `accepted=false` с безопасной error, чтобы Handshake выполнил cleanup transport.

Abort не является механизмом completion gate. Dispatcher никогда не возвращает pre-Commit outcome до возврата dormant path. Dormant path никогда не вызывает committed Complete, observer, observation Runtime cancellation или lease operations. При failed Commit отсутствует callback, который требовалось бы unregister или join. Orchestration boundary преобразует recoverable panic в тот же безопасный error result и не распространяет panic через transport contract.

Successful Commit является start boundary owner и единственной execution publication point. Внутри той же synchronization boundary Manager он переводит gate `PreCommit -> Committed`, передаёт ownership, публикует owner-scoped bundle и делает eligible ровно один уже существующий path. Operation не возвращает success и не освобождает synchronization boundary до завершения всех этих effects.

Второй launch call, acknowledgement launch или fallible post-Commit activation step отсутствуют. Попадание goroutine в scheduler не входит в Commit; `Session.Start`, Run, Cleanup, Complete, invocation observer и release lease также в него не входят.

Commit gate не является ещё одним lifecycle, coordinator, supervisor или owner. Это узкий one-shot publication binding, представляющий существующий transition `PreCommit -> Committed`. Dispatcher создаёт его, а Manager получает только право вызвать единственную publication operation. `Committed` делает единственный dormant path eligible. Non-committed outcome завершает этот path без execution. Repeated publication возвращает уже зафиксированный outcome и никогда не освобождает второй path. После publication binding не предоставляет lifecycle control.

## 12. Commit Bundle и Identity Lease

Commit является единственной publication point Manager и handoff. Он либо завершается failure без publication committed state, либо необратимо публикует полный owner bundle:

```text
RegistrationID
CompletionAdapter
OwnerLifetimeLease
StopPublicationBinding
ExecutionPublicationBinding
```

Все потенциально fallible validation, construction и allocation, необходимые bundle, завершаются до входа в publication critical section. Critical section выполняет только bounded mutations state Manager и заранее подготовленного panic-free one-shot publication binding. Под lock Manager не вызываются внешний код, Session или owner methods, callback, goroutine scheduler operation или blocking channel send.

Свойства bundle:

- RegistrationID не повторяется за lifetime Manager;
- identity Lease связана one-to-one с RegistrationID и одной owner control cell;
- для committed Registration существует ровно один lease;
- Completion Adapter не принимает RegistrationID caller и может завершить только связанную Registration;
- Lifetime Lease может освободить только связанный lease;
- Stop publication связывает capability Snapshot с той же уже подготовленной control cell owner;
- Execution publication связывает ровно одну dormant launch goroutine с тем же RegistrationID и один раз публикует eligibility через panic-free state mutation;
- failure возврата полного bundle означает, что Commit не произошёл;
- частично committed или частично возвращённого bundle не существует.

Panic-contained интегрированная operation Commit состоит только из:

1. validation, что Reservation и state Manager всё ещё допускают Commit;
2. consumption Reservation и publication identity Registration;
3. publication Completion Adapter, Owner Lifetime Lease и binding Stop;
4. вызова узкого one-shot execution-publication binding с `Committed`;
5. освобождения общей synchronization boundary Manager;
6. возврата полного bound result.

Steps 2–4 externally inseparable. Lookup и BeginShutdown используют ту же synchronization boundary и не могут наблюдать Registration до committed execution binding. Manager хранит только узкое publication right; он не владеет path binding и не вызывает lifecycle Session.

Commit имеет ровно два recoverable outcomes:

- **non-committed failure:** validation завершается failure до первой observable mutation; Registration, lease, Stop binding и committed execution outcome отсутствуют, а Dispatcher публикует non-committed outcome gate;
- **successful Commit:** Registration и полный bundle опубликованы, outcome gate равен `Committed`; rollback запрещён.

Recoverable panic между partial publication mutations отсутствует. Там не вызывается внешний или arbitrary code, а каждая подготовленная publication mutation является panic-free по contract. Process termination, unrecoverable failure Go runtime, memory corruption и эквивалентные failures являются accepted unrecoverable limitations, а не recoverable branch Commit panic.

Commit не включает scheduling goroutine, `Session.Start`, `Session.Run`, Cleanup, Complete, invocation observer, release lease или возврат Dispatcher в Handshake.

До Commit:

- Registration, accounting lease и capability Snapshot не существуют;
- Lookup не может наблюдать prepared owner;
- BeginShutdown не может его зафиксировать;
- dormant launch goroutine не может войти в lifecycle owner;
- Dispatcher сохраняет ownership transport и cancellation.

После Commit:

- Registration немедленно видима Lookup и shutdown accounting;
- BeginShutdown может зафиксировать immutable identity и bound Stop capability;
- ровно один committed launch path уже eligible для входа в lifecycle owner;
- Session, WebSocket, cancellation и execution responsibility принадлежат owner;
- Registration никогда не откатывается и может быть удалена только Complete;
- Dispatcher обязан вернуть `accepted=true` и не выполнять cleanup.

Post-Commit failure branch Dispatcher отсутствует. Panic, cancellation, failure Start или любое другое recoverable event после Commit является normal terminal cause owner и следует ordering Cleanup, Complete, observer и release lease.

Между возвратом Commit и фактическим scheduler execution launch goroutine может существовать интервал. Он безопасен и не является orphan interval: Registration, lease, Stop capability, ownership и уже eligible execution path существуют вместе. Shutdown может запросить Stop в этом интервале; owner наблюдает termination до Start и входит в normal terminal path.

Repeated Commit для той же ещё существующей Registration возвращает тот же logical bound publication result: те же RegistrationID, Completion Adapter, identity Owner Lifetime Lease, Stop-publication binding и committed execution outcome. Он не вызывает publication повторно, не вызывает повторную установку callback, не создаёт lease, Stop capability или goroutine и не меняет accounting. Commit после Complete сохраняет semantics `ErrRegistrationRemoved`, определённые DP-003.

Release lease раскрывается только через owner-bound panic-safe adapter с explicit outcome. Release owner имеет не более одного effective effect. Repeated release возвращает стабильный anomaly result. Unknown, stale или foreign release не может повлиять на другой lease. Never-reused RegistrationID вместе с owner-bound lease identity предотвращает ABA.

Panic adapter или unsuccessful release оставляет accounting этого lease активным. Он не может опустошить другой lease или позволить Wait сообщить success.

## 13. Contract Handoff Dispatcher

Dispatcher выполняет один serialized control flow:

1. Reserve SessionID;
2. создать Session Core;
3. создать Execution Owner в `PreCommit`;
4. создать ровно одну dormant launch goroutine, ожидающую Commit gate owner;
5. provisionally attach WebSocket и connection cancellation;
6. активировать все control structures owner без запуска execution Session;
7. проверить termination и eligibility shutdown;
8. выполнить Commit prepared binding owner.

Commit является единственной linearization point publication Registration, publication execution, transfer ownership и success handoff. Его внутренняя mutation Manager и transition Commit gate защищены одной synchronization boundary и externally indivisible. Последующая mutation activation отсутствует, как и recoverable state «Commit успешен, но publication execution не состоялась».

Dispatcher гарантирует exactly-once execution, создавая одну launch goroutine до Commit и передавая один one-shot gate в binding Commit. Manager гарантирует atomic visibility opaque binding вместе с Registration, но не создаёт и не владеет goroutine. После возврата successful Commit у Dispatcher отсутствует obligation launch, cleanup, Session или cancellation; он только возвращает `accepted=true`.

Observable semantics `DispatchAuthenticated(...)(accepted, error)`:

- `accepted=false, error=nil` — transfer ownership не произошёл; Handshake/Upgrade boundary сохраняет и очищает transport;
- `accepted=false, error!=nil` — transfer ownership не произошёл; Handshake/Upgrade boundary сохраняет и очищает transport после фиксации operational failure;
- `accepted=true, error=nil` — Commit успешен, transfer ownership завершён, Dispatcher немедленно возвращает управление, а owner продолжает независимо.

Production Dispatcher не возвращает `accepted=true` вместе с error. Recoverable post-Commit execution outcomes, зафиксированные до построения Terminal Result, передаются через Terminal Observer. Callback outcomes, возникающие после этого построения, принадлежат только существующему terminal-accounting path и никогда не возвращаются через Dispatcher. Request cancellation до Commit приводит к Abort и `accepted=false`; cancellation после Commit является normal termination source owner.

## 14. Lifecycle Execution Owner

Концептуальная state machine:

```text
PreCommit
    -> Committed

Committed
    -> Starting
    -> Terminalizing

Starting
    -> Running
    -> Terminalizing

Running
    -> Terminalizing

Terminalizing
    -> Terminal
```

`PreCommit` означает, что owner, values Session, control cell, binding Stop, read-only input observation root Runtime context и одна dormant launch goroutine подготовлены, пока Dispatcher владеет Reservation, transport и provisional launch control. Callback Runtime cancellation отсутствует. Goroutine ожидает Commit gate, поэтому execution Session не началось.

`Committed` начинается внутри successful Commit до освобождения общей synchronization boundary. Registration, lease, Stop capability, ownership Session и eligibility execution публикуются вместе, а Start ещё не линеаризован. Фактическое scheduling может произойти позже без создания ещё одного lifecycle state.

`Starting` и `Running` являются явными linearization states Start и Run.

`Terminalizing` обязателен для каждого committed terminal path. Он исполняет panic-safe terminal chain и закрывает admission control calls на определённом post-observer step. Ни один post-Commit path не переходит напрямую в Terminal.

`Terminal` означает, что execution Session и work observer завершены, `UnregisterAndDrain` подтвердил отсутствие будущей или вошедшей callback work, admission control calls закрыт, каждый entered control call вернулся, а control cell sealed. Остаются только conditional attempt release lease и немедленный technical return. Неподтверждённый lifetime callback остаётся в `Terminalizing` и никогда не заявляет Terminal. Restart запрещён.

Dormant goroutine принадлежит `PreCommit` и не требует другого lifecycle state. Non-committed outcome gate завершает этот path без transition в `Committed` и без Complete, committed observer, callback entry или accounting owner lifetime Manager. Каждый post-Commit terminal outcome входит в `Terminalizing` до единого определения Terminal.

## 15. Linearization Start и Run

После handoff и до linearization Start owner устанавливает observation cancellation Host-owned root Runtime context. Этот root context является единственным нормативным источником observation; callback никогда не наблюдает derived execution context. Создание и ownership callback принадлежат исключительно owner; Dispatcher и Manager не создают и не сохраняют его. Callback получает только узкую operation causal cell и не получает Session, Manager, WebSocket, lifecycle control или completion capability.

Установка использует один race-safe contract: сначала зарегистрировать observation, затем синхронно проверить root Runtime context, либо применить эквивалентный primitive с той же гарантией. Поэтому root cancellation до или во время регистрации либо доставляется механизмом context, либо наблюдается проверкой после регистрации. Оба пути пытаются выполнить одну и ту же first-writer mutation `RuntimeCanceled`, поэтому cancellation effective не более одного раза. Повторный callback invocation безопасен.

Если установка callback возвращает error или panic, wrapper owner фиксирует sanitized anomaly установки callback, пытается установить termination intent `ExecutionFailure` и входит в `Terminalizing`. Registration и ownership остаются committed; rollback запрещён. Если root Runtime context наблюдается canceled, вместо этого предпринимается попытка `RuntimeCanceled`. Cleanup callback всё равно выполняется через общий terminal contract, в том числе когда установка не создала registration.

После установки owner проверяет unified termination state под control lock.

- При уже существующем termination он входит в `Terminalizing`; Start и Run пропускаются.
- Иначе `Committed -> Starting` является Start linearization point.

Lock освобождается до `Session.Start`. Start вызывается не более одного раза. Error или panic ведёт в `Terminalizing`; Run не вызывается.

После successful Start owner повторно проверяет termination state:

- существующий termination ведёт в `Terminalizing`;
- иначе `Starting -> Running` является Run linearization point.

Lock освобождается до `Session.Run`.

Если termination линеаризуется первым, Run запрещён. Если первым линеаризуется `Running`, Run официально начался, даже если cancellation происходит до следующей machine instruction с вызовом Run. Вызов Run с уже canceled execution context является допустимым running path.

Lifecycle lock не удерживается во время Start, Run или Cleanup.

## 16. Unified Termination Intent

Все источники termination конкурируют через одну owner control cell и одну first-writer linearization point.

Источники termination:

- `ExplicitStop`;
- `RuntimeCanceled`;
- `NaturalCompletion`;
- `ExecutionFailure`;
- `RecoveredPanic`.

Первый source, записавший пустую causal cell, становится primary termination source. `ExplicitStop`, `RuntimeCanceled`, `NaturalCompletion`, `ExecutionFailure` и `RecoveredPanic` используют одну mutation. Simultaneous attempts имеют ровно одного winner под control lock.

`RequestStop() bool` возвращает `true` только если этот invocation первым установил termination intent. После ранее принятой Runtime cancellation, explicit Stop, Terminalizing или Terminal возвращается `false`.

Timing Runtime cancellation имеет одну нормативную модель. До Commit registration callback отсутствует. После observation `Committed` dormant owner устанавливает observation root Runtime context до linearization Start через описанный race-safe contract регистрации и проверки. Root context, отменённый до Commit, синхронно отклоняется Dispatcher и никогда не создаёт callback. Root context, отменённый после Commit, но до установки, наблюдается во время установки и конкурирует через causal cell как `RuntimeCanceled`. Session Cleanup отменяет только derived execution context через его private cancellation cell; поскольку callback наблюдает отдельный root Runtime context, normal Session Cleanup не может самостоятельно создать `RuntimeCanceled`.

Owner владеет registration callback от создания до cleanup. Callback invocation и explicit RequestStop используют одинаковые admission и accounting outstanding control calls. Wrapper callback panic-safe: outward panic запрещён, panic становится sanitized anomaly callback, termination intent всё равно предпринимается, а accounting entered call уменьшается в final guard. Callback не выполняет Session I/O и не ожидает Cleanup, Complete, observer или release lease.

Последующие signals сводятся к bounded secondary categories. Они не заменяют first cause, не меняют исходный authentication result, не влияют на уже возвращённое значение `RequestStop` и не создают второе obligation Stop, Complete, observer или lease.

## 17. Contract RequestStop

`RequestStop`:

- thread-safe;
- не ждёт Session I/O;
- не ждёт Start, Run, Cleanup, Complete, observer или terminal completion;
- выполняет только bounded owner-state mutation и локальную context cancellation;
- не обещает жёсткий time bound стандартной context cancellation;
- возвращает `false` для repeated, concurrent losing, post-Terminalizing и post-Terminal calls.

До Start он запрещает Start и Run. Во время Start отменяет execution и запрещает Run, если Run ещё не линеаризован. Во время Run отменяет execution context. Он никогда не вызывает Session Cleanup в goroutine caller.

Stable capability содержит только detached control cell. Она не раскрывает Session, WebSocket, Context, CancelFunc, Send operation, callback registration, Runtime, Listener или record Manager.

До release lease explicit RequestStop и callback Runtime входят через одну termination operation. Terminalization отклоняет новые entries и выполняет drain уже вошедших calls. Непосредственно перед invocation adapter release lease detached cell запечатывается: каждый последующий invocation capability возвращает `false` без чтения или mutation освобождённого state owner.

## 18. Semantics Session Cleanup

Owner вызывает Session Cleanup один раз на каждом activated terminal path:

- termination до Start;
- Start error или panic;
- Run return, error или panic;
- explicit Stop;
- Runtime cancellation.

Execution Owner создаёт cleanup context. Он независим от execution context, не отменяется автоматически root Runtime cancellation и существует только для cooperative cleanup и join operations. Expiration этого context не разрешает Complete, release lease или successful Wait, пока Cleanup фактически не вернул acknowledgement.

Cleanup включает существующие semantics Session Stop:

- закрытие transport;
- ожидание read loop при его наличии;
- идемпотентное repeated observation;
- стабильный terminal result.

Текущий primary Stop/read-loop wait может блокироваться бесконечно независимо от cancellation cleanup context. Это accepted limitation. Пока Cleanup заблокирован, lease остаётся активным, а Manager Wait ожидает.

`CleanupResult` имеет два cancellation outcomes:

- `Confirmed` разрешает terminal chain проверить eligibility release lease;
- `Anomaly` означает, что effective connection cancellation доказать не удалось.

При cancellation `Anomaly` owner всё равно следует единому terminal order: предпринимает Complete, строит Terminal Result, вызывает observer, закрывает control admission, успешно unregister-ит и drain-ит Runtime callback, seal-ит control cell и достигает Terminal. Owner не освобождает lease, поскольку effective cancellation не подтверждена. Поэтому Manager Wait остаётся заблокированным. Если cleanup callback также не подтверждён, owner вместо этого остаётся в `Terminalizing` с активным lease. Ни один incomplete-shutdown outcome не имеет hidden retry. Operational intervention находится вне этого proposal.

## 19. Exclusive Ownership Completion

Execution Owner является единственным production terminal completion caller для своей Registration.

Owner не получает общий `Manager.Complete(RegistrationID)`. Он получает один owner-scoped Completion Adapter, созданный Runtime composition из Commit bundle:

```text
CompleteBoundRegistration() CompleteOutcome
```

Adapter не может адресовать другую Registration. Handshake, Dispatcher после activation, Host, Listener, Session, Snapshot и observer его не получают.

Технически exported internal `Manager.Complete` может остаться low-level механизмом adapter, но production composition не должна вызывать или распространять его вне construction bound adapter.

Dependency и composition tests обязаны доказать:

- только package adapter импортирует low-level completion method для production wiring;
- owner получает bound adapter, а не Manager;
- другие production constructors не получают completion capability;
- один invocation owner создаёт не более одной effective mutation Manager.

`CompleteOutcome` категоризирован как `Completed` или `AccountingAnomaly`. False/unknown removal становится anomaly category и не повторяется.

## 20. Полный Lifetime Owner Lease

Owner Lifetime Lease создаётся и публикуется атомарно Commit. Он остаётся активным весь Runtime-owned lifetime.

Единый terminal order для каждого committed path:

```text
войти в Terminalizing
    -> Session Cleanup возвращает управление
    -> Complete bound Registration
    -> создать immutable Terminal Result
    -> synchronous Terminal Observer возвращает управление
    -> вызвать panic-safe UnregisterAndDrain
       -> закрыть admission entries RequestStop и Runtime callback
       -> запретить будущий callback entry
       -> unregister callback Runtime cancellation
       -> дождаться уже вошедших control calls
       -> вернуть immutable CallbackCleanupResult
    -> если отсутствие callback не подтверждено: остаться в Terminalizing с активным lease
    -> запечатать и отделить control cell
    -> перевести owner в Terminal
    -> если выполнены все release conditions, вызвать panic-safe adapter release Lease
    -> не выполнять дальнейшую Runtime-owned work
    -> owned goroutine возвращает управление
```

Release lease разрешён только при выполнении всех условий:

- Cleanup вернулся и подтвердил effective cancellation;
- invocation bound Complete вернулся;
- Terminal Observer вернулся;
- lifecycle установки и cleanup callback завершён;
- `UnregisterAndDrain` подтвердил отсутствие будущей или вошедшей callback work;
- каждый вошедший control call вернулся;
- admission control calls закрыт;
- causal cell sealed.
- owner достиг Terminal.

Если cleanup callback не может подтвердить отсутствие будущей и вошедшей callback work, owner остаётся в `Terminalizing`, lease не освобождается, а Manager Wait не может завершиться успешно. Если owner достиг Terminal, но отсутствует другое release condition, включая acknowledgement effective cancellation, lease также остаётся активным. Successful release lease является последней Runtime-owned operation launch wrapper и linearization point, в которой Manager может удалить accounting lease. После возврата goroutine не выполняет Runtime-owned work и немедленно возвращает управление.

Manager Wait гарантирует отсутствие оставшейся Runtime-owned work; он не утверждает scheduler-level observation физического исчезновения stack goroutine.

Blocked mandatory step сохраняет lease активным. Panic mandatory step изолируется и записывается; последующие mandatory steps всё равно выполняются, когда их prerequisites остаются доказуемыми. Normal completion, explicit Stop, Runtime cancellation, failure установки callback, panic invocation callback, failure Start или Run, recovered execution panic, cancellation anomaly, Complete anomaly и observer anomaly используют тот же порядок. Только неподтверждённый cleanup callback запрещает transition `Terminalizing -> Terminal`.

## 21. Panic-Safe Terminal Chain

Launch wrapper владеет одним outer recover boundary. Каждая terminal operation также использует независимую safe-invocation boundary, чтобы panic не пропустил последующие obligations.

Правила safe invocation:

| Operation | Обработка panic | Последующие obligations |
| --- | --- | --- |
| Internals Session Cleanup | Wrapper Cleanup записывает sanitized category и возвращает `CleanupResult`; outward panic запрещён | Complete, observer и cleanup callback продолжаются; Terminal по-прежнему требует Confirmed cleanup callback, а release требует confirmed cancellation |
| Internals cancellation cell | Wrapper Cleanup recover-ит panic, вызывает private cancellation primitive и записывает `CancellationPanic`; невозможность подтвердить cancellation становится `Anomaly` | Остальная chain продолжается; release пропускается при `Anomaly` |
| Установка callback Runtime | Recover-ить panic или error как sanitized anomaly установки; попытаться установить соответствующий termination source | Начинается общая terminal chain; rollback и изменение result Dispatcher отсутствуют |
| Invocation callback Runtime | Final guard уменьшает accounting entered call; outward panic преобразуется в sanitized anomaly callback, а termination intent всё равно предпринимается | Callback возвращает управление без Session I/O или terminal waits |
| Completion Adapter | Записать `CompletePanic` и accounting anomaly | Observer и cleanup callback продолжаются; Terminal и release остаются conditional |
| Adapter Terminal Observer | Изолировать panic как local operational anomaly вне Terminal Result; не вызывать повторно | Cleanup callback продолжается; Terminal и release остаются conditional |
| `UnregisterAndDrain` | Закрыть admission entry, recover-ить internal panic, unregister и дождаться entered calls; при завершении вернуть immutable `CallbackCleanupResult` | Confirmed result разрешает seal и Terminal; unconfirmed result оставляет owner в Terminalizing с активным lease без retry |
| Adapter release lease | Recover panic и вернуть explicit unsuccessful release outcome вне Terminal Result; не выполнять re-panic | Hidden retry отсутствует; accounting lease остаётся активным, и Wait не может завершиться успешно |

Panic Start и Run восстанавливается owned-goroutine boundary, категоризируется и направляется в Terminalizing.

Panic не выполняется повторно из owned goroutine. Diagnostics backend не определяется. Process termination, unrecoverable Go runtime failure, memory corruption и permanently blocked operation находятся вне recoverable guarantees.

## 22. Outstanding Termination Calls

RequestStop и post-Commit callback Runtime cancellation используют одну control cell.

Invocation:

1. входит только при открытом admission termination;
2. увеличивает invocation count;
3. пытается выполнить first-writer termination mutation;
4. выполняет локальную cancellation execution context, если победил;
5. уменьшает count до возврата.

Вход в `Terminalizing` фиксирует направление lifecycle, но сам по себе не закрывает admission control calls. После возврата observer owner вызывает идемпотентный `UnregisterAndDrain() CallbackCleanupResult`. Operation закрывает admission под control lock, запрещает будущий callback entry, выполняет panic-safe unregister и вне locks Session, lifecycle и Manager ждёт возврата уже вошедших calls. Если operation завершается, она возвращает ровно один immutable result и никогда не распространяет panic. Permanently blocked entered call сохраняет operation in progress, owner в Terminalizing, lease активным, а Manager Wait — blocked.

`Confirmed` доказывает, что registration не может создать будущую callback work и что каждый вошедший callback/control call вернулся. После этого owner может seal-ить cell и достичь Terminal. `Unconfirmed` фиксирует anomaly lifetime callback; owner остаётся в Terminalizing, lease остаётся активным, Manager Wait остаётся blocked, а automatic retry отсутствует. Repeated cleanup возвращает тот же detached result и не имеет второго unregister или accounting effect.

Detached Stop capabilities могут пережить Terminal. Непосредственно перед release lease их control cell запечатывается и отделяется от state owner; последующие calls всегда возвращают `false`, включая случай unsuccessful outcome adapter release.

## 23. Terminal Result и Observer

Terminal Observer синхронен и вызывается один раз. Его возврат покрывается Owner Lifetime Lease.

Terminal Result является immutable value только с bounded enums и booleans:

- категория Start: `NotAttempted`, `Succeeded`, `Failed`, `Panicked`;
- категория Run: `NotStarted`, `Returned`, `Failed`, `Panicked`;
- категория Cleanup: `NotRequired`, `Succeeded`, `Failed`, `Panicked`, `Blocked` только пока result ещё не опубликован;
- категория cancellation: `Confirmed`, `Anomaly`;
- категория Complete: `Completed`, `AccountingAnomaly`, `Panicked`;
- категория phase recovered execution panic без raw panic value;
- primary termination source;
- bounded secondary termination categories.

Result не хранит arbitrary raw error, credentials, headers, request, WebSocket, Context, callback, Session или mutable collection. Он намеренно представляет только execution-lifecycle outcomes, известные в момент его построения до invocation observer: уже наблюдавшиеся к этому моменту anomaly установки или invocation Runtime callback, Start, Run, Cleanup/Stop, recovered execution panic, Complete и termination source.

Callback invocation outcomes, впервые возникающие после построения Terminal Result, outcome `UnregisterAndDrain`, outcome invocation observer, outcome release lease и outcome возврата goroutine намеренно находятся вне Terminal Result, поскольку возникают после построения value. Они никогда не изменяют опубликованный result и не вызывают Observer второй раз. Late callback outcomes остаются bounded local facts terminal accounting; только cleanup callback определяет, разрешает ли lifetime callback состояние Terminal. Panic observer изолируется adapter и не препятствует последующему cleanup callback или eligible release lease. Anomaly release lease также не изменяет Terminal Result. Это local operational anomalies; diagnostics backend остаётся вне scope.

Blocking observer сохраняет lease активным, а Wait — ожидающим. Observer не может вызвать Complete, release lease, RequestStop или сохранить execution capability через свой contract.

## 24. Capability-Bearing Shutdown Snapshot

Текущий identity-only internal public API Snapshot расширяется совместимо.

Каждая immutable Shutdown Registration сохраняет существующие:

- `SessionID`;
- `RegistrationID`.

Она получает один read-only accessor, возвращающий stable non-owning Stop-request capability, связанную с committed owner control cell. Существующие identity accessors и identity-only callers сохраняют behavior.

Сохраняются rules Snapshot:

- первый BeginShutdown атомарно фиксирует membership;
- reservations исключены;
- Commit, выигравший до BeginShutdown, присутствует; проигравший отсутствует;
- repeated BeginShutdown возвращает тот же logical Snapshot;
- Complete не изменяет captured Snapshot;
- capability безопасна после Complete и Terminal;
- capability безопасна после release lease и возвращает `false` без доступа к state owner;
- entry не раскрывает Session, Context, Runtime, Send, callback registration или mutable owner state.

Existing Snapshot tests сохраняют все identity assertions и добавляют assertions lifetime/concurrency capability.

## 25. Failure Matrix

Каждая pre-Commit строка использует panic-safe sequence: publication non-committed, ожидание возврата dormant path, освобождение owner-local values, Abort и возврат `accepted=false`. До Commit callback отсутствует. Каждая строка после successful Commit имеет один eligible path и использует единый terminal order.

| Failure или race | Outcome dormant path | Owner transport | Owner cancellation | Outcome Reservation | Outcome Registration | accepted/error | Complete | Observer | Outcome callback | Final state owner | Outcome lease | Outcome Manager.Wait |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| Failure construction owner | Не создан | Dispatcher | Dispatcher | Aborted | Нет | `false`, safe construction error | Не вызывается | Не вызывается | Никогда не создан | Owner отсутствует | Нет | Не затрагивается после Abort |
| Failure readiness dormant goroutine | Non-committed path возвращается до Dispatcher | Dispatcher | Dispatcher | Aborted | Нет | `false`, safe readiness error | Не вызывается | Не вызывается | Никогда не создан | `PreCommit` disposed | Нет | Не затрагивается после Abort |
| Panic Dispatcher после создания dormant path | Non-committed опубликован один раз; path joined | Dispatcher | Dispatcher | Aborted | Нет | `false`, sanitized panic error | Не вызывается | Не вызывается | Никогда не создан | `PreCommit` disposed | Нет | Не затрагивается после Abort |
| Panic attachment | Non-committed опубликован один раз; path joined | Dispatcher | Dispatcher | Aborted | Нет | `false`, sanitized attachment error | Не вызывается | Не вызывается | Никогда не создан | `PreCommit` disposed | Нет | Не затрагивается после Abort |
| Runtime context canceled до Commit | Non-committed опубликован один раз; path joined | Dispatcher | Dispatcher | Aborted | Нет | `false`, cancellation error | Не вызывается | Не вызывается | Никогда не создан | `PreCommit` disposed | Нет | Не затрагивается после Abort |
| BeginShutdown выигрывает Commit | Non-committed опубликован один раз; path joined | Dispatcher | Dispatcher | Aborted | Нет; Snapshot исключает | `false`, rejection Commit | Не вызывается | Не вызывается | Никогда не создан | `PreCommit` disposed | Нет | Success после Abort при пустом остальном accounting |
| Recoverable validation failure Commit | Non-committed опубликован один раз; path joined | Dispatcher | Dispatcher | Aborted | Нет | `false`, validation error | Не вызывается | Не вызывается | Никогда не создан | `PreCommit` disposed | Нет | Не затрагивается после Abort |
| Successful Commit | Committed один раз; один path eligible | Owner | Owner | Consumed | Registered с полным bundle | `true`, nil | Один раз при terminalization | Один раз | Устанавливается только owner | `Committed`, затем normal lifecycle | Active до eligible release | Success только после очистки Registration и lease |
| Repeated Commit до Complete | Возвращается существующий committed outcome | Owner | Owner | Уже consumed | Те же Registration и logical bundle | Тот же prior successful result | Новый invocation отсутствует | Новый invocation отсутствует | Вторая установка отсутствует | Существующий state не меняется | Тот же lease; accounting не меняется | Без изменений |
| Runtime context canceled после Commit до установки callback | Owner наблюдает cancellation во время race-safe установки | Owner | Control cell owner | Consumed | Registered до Complete | Prior `true`, nil | Один раз | Один раз с RuntimeCanceled | Registration или проверка после регистрации фиксирует RuntimeCanceled один раз | `Committed -> Terminalizing -> Terminal` после Confirmed cleanup | Release при выполнении всех conditions | Pending до Complete и release |
| Установка callback успешна | Owner продолжает после race-safe проверки | Owner | Control cell owner | Consumed | Registered до Complete | Prior `true`, nil | Один раз при последующей terminalization | Один раз | Одна registration во владении owner | `Committed -> Starting` либо `Terminalizing`, если termination победил | Active до eligible release | Pending до очистки terminal accounting |
| Panic или error установки callback | Owner входит в общий terminal path | Owner | Control cell owner | Consumed | Registered до Complete | Prior `true`, nil | Один раз | Один раз с anomaly установки | Partial или отсутствующая registration обрабатывается cleanup contract | Terminal только после Confirmed cleanup | Release только после Terminal и всех conditions | Permanently blocked при Unconfirmed cleanup |
| RequestStop до scheduling owner | Committed path наблюдает `ExplicitStop` | Owner | Control cell owner | Consumed | Registered до Complete | Prior `true`, nil | Один раз | Один раз с ExplicitStop | Установка и cleanup остаются obligations owner | `Committed -> Terminalizing -> Terminal` после Confirmed cleanup | Release при выполнении всех conditions | Pending до Complete и release |
| Entry callback во время Starting | RuntimeCanceled конкурирует до linearization Run | Owner | Control cell owner | Consumed | Registered до Complete | Prior `true`, nil | Один раз | Один раз | Entry учтён; Run запрещён, если callback победил | `Starting -> Terminalizing -> Terminal` после Confirmed cleanup | Release при выполнении всех conditions | Pending до Complete и release |
| Entry callback во время Running | RuntimeCanceled отменяет execution | Owner | Control cell owner | Consumed | Registered до Complete | Prior `true`, nil | Один раз | Один раз | Entry учтён и возвращается без Session I/O | `Running -> Terminalizing -> Terminal` после Confirmed cleanup | Release при выполнении всех conditions | Pending до Complete и release |
| Panic wrapper callback | Panic sanitized; termination attempted; entry count уменьшен | Owner | Control cell owner | Consumed | Registered до Complete | Prior `true`, nil | Один раз | Один раз с anomaly callback, только если она известна до построения Terminal Result; для поздней anomaly второго invocation нет | Wrapper возвращается; поздняя anomaly остаётся bounded fact terminal accounting; Confirmed cleanup всё ещё обязателен | Terminal после Confirmed cleanup; иначе остаётся Terminalizing | Release при выполнении всех conditions; иначе остаётся active | Pending до Complete и release; blocked, пока cleanup Unconfirmed |
| Normal completion | Committed path выполняется normally | Owner | Owner | Consumed | Removed через Complete | Prior `true`, nil | Один раз, effective once | Один раз с NaturalCompletion | `UnregisterAndDrain` Confirmed | `Running -> Terminalizing -> Terminal` | Released once | Success после release |
| Callback уже вошёл во время terminalization | Owner ждёт вне locks lifecycle и Manager | Owner | Вошедший callback до возврата | Consumed | Complete уже attempted | Prior `true`, nil | Один раз | Вернулся один раз | Drain ждёт entry; будущий entry отсутствует | Terminalizing до drain, затем Terminal | Active до Terminal | Blocked до возврата entry и release lease |
| Unregister success | Confirmed cleanup result | Owner | Sealed cell owner после drain | Consumed | Complete уже attempted | Prior `true`, nil | Один раз | Вернулся один раз | Будущая или вошедшая callback work отсутствует | `Terminalizing -> Terminal` | Eligible с учётом остальных conditions | Pending до release |
| Panic или internal anomaly unregister | Unconfirmed cleanup result; outward panic отсутствует | Owner | Control cell owner | Consumed | Complete уже attempted | Prior `true`, nil | Один раз | Вернулся один раз | Отсутствие callback work не доказано | Остаётся Terminalizing | Retained; retry отсутствует | Blocked |
| Unregister не может доказать non-entry | Unconfirmed cleanup result | Owner | Control cell owner | Consumed | Complete уже attempted | Prior `true`, nil | Один раз | Вернулся один раз | Future или entered work нельзя исключить | Остаётся Terminalizing | Retained; retry отсутствует | Blocked |
| Repeated unregister | Возвращается существующий immutable cleanup result | Owner | Как в первом outcome | Consumed | Без изменений | Prior `true`, nil | Новый invocation отсутствует | Новый invocation отсутствует | Второй unregister или drain отсутствует | Без изменений | Accounting не меняется | Без изменений |
| Cleanup cancellation anomaly | Committed path terminalizes | Owner | Wrapper Cleanup | Consumed | Complete attempted once | Prior `true`, nil | Один раз | Один раз с Anomaly | Confirmed cleanup callback всё ещё обязателен | Terminal после Confirmed cleanup callback | Retained, поскольку cancellation не подтверждена | Blocked |
| Complete false | Committed path продолжается | Owner | Control cell owner | Consumed | Accounting anomaly; этот call ничего не удаляет | Prior `true`, nil | Один раз, AccountingAnomaly | Один раз с anomaly Complete | Confirmed cleanup всё ещё обязателен | Terminal после Confirmed cleanup callback | Release только при выполнении всех conditions | Registration или lease предотвращает false success |
| Complete panic | Panic изолирован; path продолжается | Owner | Control cell owner | Consumed | Accounting outcome может остаться активным | Prior `true`, nil | Один раз, Panicked | Один раз с CompletePanic | Confirmed cleanup всё ещё обязателен | Terminal после Confirmed cleanup callback | Release только при выполнении всех conditions | Оставшийся accounting предотвращает false success |
| Observer blocks | Path заблокирован в observer | Owner | Owner | Consumed | Complete уже attempted | Prior `true`, nil | Один раз | In progress | Cleanup ещё не достигнут в terminal order | Terminalizing | Active | Blocked |
| Observer panic | Adapter возвращает sanitized anomaly | Owner | Control cell owner | Consumed | Complete уже attempted | Prior `true`, nil | Один раз | Один раз; panic изолирован | Далее выполняется `UnregisterAndDrain` | Terminal после Confirmed cleanup | Eligible с учётом остальных conditions | Pending до release |
| Anomaly release lease | Все предыдущие steps завершены | Runtime-owned transport work отсутствует | Sealed cell | Consumed | Complete уже attempted | Prior `true`, nil | Один раз | Вернулся один раз | Confirmed; callback work отсутствует | Terminal | Retained; retry отсутствует | Blocked |

Unrecoverable low-level panic completion может оставить accounting Registration активным после exit owner lease. Неподтверждённый lease release оставляет accounting lease активным. В обоих случаях Manager Wait остаётся ожидающим и не сообщает false success.

## 26. Гарантия Manager Wait

После успешного `Manager.Wait`:

- set Reservation пуст;
- set Registration пуст;
- set Owner Lifetime Lease пуст;
- каждый tracked owner достиг Terminal;
- каждая registration callback Runtime cancellation удалена;
- все вошедшие control calls вернулись;
- каждый Session Cleanup вернул immutable acknowledgement;
- каждый tracked connection context effectively canceled;
- каждый synchronous Terminal Observer вернулся;
- после successful release lease не осталось Runtime-owned work.

Если Cleanup, observer, drain control calls, removal Registration или release lease не завершается как требуется, accounting остаётся активным и Wait не завершается успешно. Deadline caller может завершить только Wait этого caller без изменения правдивого state Manager.

Wait не обещает scheduler-level доказательство физического возврата stack owned goroutine после её последней non-Runtime instruction.

## 27. Shutdown Ordering

Нормативный порядок:

```text
Host закрывает Admission Gate
    -> Manager.BeginShutdown фиксирует capability-bearing Snapshot
    -> вызываются capabilities RequestStop Snapshot
    -> отменяется root Runtime context
    -> инициируется Listener Stop
       |-> drain handlers Listener -> Listener Stop возвращает управление
       |-> owners terminalize -> eligible leases освобождаются
    -> после возврата Listener Stop выполняется Manager Wait
    -> останавливаются оставшиеся Runtime components
```

Ветки drain handlers и drain owners выполняются параллельно; ни одна не упорядочена перед другой. Manager Wait начинается только после возврата Listener Stop. Owners, committed до BeginShutdown, присутствуют в Snapshot. In-flight Reservation, чей Commit проиграл, выполняет Abort, а Dispatcher очищает transport. Commit winner необратимо публикуется, входит в Snapshot, передаёт ownership, возвращает `accepted=true` и следует normal terminal lifecycle owner.

Explicit Stop и root cancellation используют один first-writer state, поэтому их ordering не создаёт второе obligation Stop.

Это сохраняет ARCH-002: admission закрывается первым, root Runtime context отменяется до Listener Stop.

## 28. Анализ Deadlock

- Pre-Commit proof: `panic-safe boundary Dispatcher -> publication non-committed -> return dormant path -> disposal owner-local values -> Abort -> accepted=false`. Callback до Commit отсутствует. Dispatcher не может вернуться, пока recoverable pre-Commit dormant path остаётся blocked.
- Commit proof: `precompute полного bundle -> panic-free atomic publication под lock Manager -> committed outcome gate -> unlock -> один path owner eligible`.
- Post-Commit proof: `path owner -> install callback -> lifecycle -> Terminalizing -> Cleanup -> Complete -> observer -> UnregisterAndDrain -> Confirmed -> seal -> Terminal -> outcome lease`.
- Dispatcher не ждёт owner после successful Commit, а owner не ждёт Dispatcher после observation `Committed`.
- Manager никогда не ждёт dormant path. Dormant path ждёт только one-shot outcome и не ждёт Manager после появления outcome.
- Commit выполняет только bounded panic-free in-memory publication и не вызывает callback, channel send, scheduling goroutine, I/O Session или owner method под locks Manager.
- One-shot publication сохраняет один terminal outcome до wake единственного waiter; repeated publication наблюдает существующий outcome, поэтому lost wakeup и double release невозможны по contract.
- Owner Start, Run и Cleanup выполняются без locks owner или Manager.
- Entry callback выполняет только mutation causal cell и возвращается; он не ждёт terminal operation owner.
- RequestStop и Runtime cancellation не ждут Session operations.
- После observer owner не удерживает lock lifecycle или Manager, пока `UnregisterAndDrain` ждёт вошедшие calls. Callback никогда не ждёт этот cleanup и поэтому не может сформировать lock cycle.
- Completion Adapter выполняет только mutation Manager и не ждёт Manager Wait.
- Observer по contract не может вызвать Complete, release lease или ждать owner.
- Release lease не ждёт Manager Wait.
- Listener Stop может ждать HTTP handlers, пока owners terminalize параллельно; owner не ждёт Listener Stop.
- Manager Wait ждёт только convergence accounting и не удерживает lock во время ожидания.

Circular wait между Dispatcher и prepared owner отсутствует, и recoverable panic Dispatcher не может orphan-ить dormant path. Manager Wait никогда не участвует в cleanup callback, а после Commit у Dispatcher отсутствует obligation callback или owner. Permanently blocked Session Cleanup, observer, entry callback или неподтверждённый cleanup callback может задержать правдивый shutdown, но по этим contracts не создаёт lock cycle.

## 29. Migration с текущего Dispatcher

Migration выполняется последовательно:

1. ввести transport-independent Core без изменения текущей Session;
2. добавить package-private provisional preparation Session и owner;
3. добавить cleanup acknowledgement вокруг текущих Stop и connection cancellation;
4. добавить Commit bundle Manager, opaque one-shot execution binding и совместимый accessor capability Snapshot;
5. добавить tests dormant launch-path preparation и boundary Commit;
6. добавить установку Runtime callback только owner после Commit и proof `UnregisterAndDrain`;
7. перенести Start/Run/Cleanup в owner после successful Commit;
8. подключить shutdown ordering Manager через Runtime composition.

На каждом шаге Handshake продолжает видеть `DispatchAuthenticated(...)(accepted, error)`. `accepted=true` возвращается только после successful Commit. `accepted=false` всегда означает, что Commit не произошёл, Dispatcher сохраняет ownership cleanup, а Registration, lease и capability Snapshot отсутствуют.

## 30. Dependency Boundaries и Контроль God Object

Execution Owner координирует lifecycle только одной Session. Каждая дополнительная сущность имеет отдельный invariant:

- cancellation cell доказывает effective connection cancellation;
- Completion Adapter запрещает completion чужой registration;
- Lifetime Lease покрывает lifetime goroutine после удаления visibility registration;
- Terminal Result передаёт immutable terminal categories;
- Terminal Observer потребляет ровно один result;
- Stop capability запрашивает termination без ownership Session.

Ни одна сущность не является service locator или generic coordination framework. Owner не получает responsibilities routing, delivery, persistence, presence, limits или supervision.

## 31. Архитектурные инварианты

- Dispatcher является единственным coordinator handoff.
- Dispatcher не получает ambiguous outcome: failed Commit означает отсутствие transfer; successful Commit означает irreversible transfer.
- Core не является Session и не владеет transport.
- Dispatcher создаёт ровно одну dormant launch goroutine и один one-shot Commit gate для каждого prospective owner.
- Dispatcher владеет одной panic-safe pre-Commit orchestration boundary, не создаёт Runtime callback и не может вернуться, пока non-committed dormant path не вернулся и prepared values не освобождены.
- Commit является единственной publication point Registration, lease, Stop capability, ownership и execution.
- Publication Commit не содержит fallible external operation и имеет только полный non-committed либо полный committed outcome.
- Commit не возвращает success и не освобождает synchronization boundary, пока execution binding не достиг `Committed`.
- Observable Registration без уже eligible execution path не существует.
- Complete является единственной removal point Registration.
- Preparation owner предшествует Commit; execution Session начинается только после Commit.
- Каждый committed path входит в Terminalizing до Terminal.
- Start и Run имеют по одной linearization point.
- Только owner устанавливает observation только root Runtime context после Commit и до linearization Start через race-safe registration-and-check contract.
- Explicit Stop и Runtime cancellation используют один first-writer termination state.
- Session Cleanup синхронно подтверждает effective canceled state.
- Только owner получает bound capabilities completion и lease.
- Terminal Observer синхронен и вызывается один раз.
- `UnregisterAndDrain` должен подтвердить отсутствие будущей и вошедшей callback work до Terminal; неподтверждённый cleanup остаётся Terminalizing.
- Release lease является последней Runtime-owned operation.
- Успешный Manager Wait покрывает Terminal и отсутствие оставшейся Runtime-owned work, а не scheduler-level возврат goroutine.
- Capability Snapshot не раскрывает Session или mutable execution state.

## 32. Альтернативы

### Synchronous Dispatcher

Отклонён, поскольку не предоставляет independent execution Runtime или shutdown capability.

### Owner выполняет Attachment

Отклонён, поскольку требует transport-transfer capability у owner и разделяет synchronous cleanup boundary Upgrade.

### Dispatcher как Activation Executor

Выбран, поскольку Dispatcher уже владеет resources Upgrade, может сделать сам Commit point ownership acceptance и publication execution и возвращает один явный outcome handoff.

### Provisional Session в `Created`

Отклонена, поскольку один state представлял бы и отсутствующий, и принадлежащий transport.

### Transport-Independent Core

Выбран, поскольку до attachment lifecycle Session не существует.

### Complete как Final Accounting

Отклонён, поскольку observer и полный lifetime goroutine не представлялись бы после removal registration.

### Global Supervisor

Отклонён, поскольку один per-Session invariant не оправдывает generic framework.

## 33. Трассировка TASK-REV-004

| Finding | Статус | Решение |
| --- | --- | --- |
| F-01 | Resolved | Session Cleanup возвращает acknowledgement effective cancellation. |
| F-02 | Resolved | Release lease является последней Runtime-owned operation перед немедленным возвратом goroutine. |
| F-03 | Resolved | Run линеаризуется в `Starting -> Running`. |
| F-04 | Resolved | Core не является Session; attachment создаёт transport-owning Session. |
| F-05 | Resolved | Один shutdown order сохраняет ARCH-002. |
| F-06 | Resolved | Owner получает только bound Completion Adapter. |
| F-07 | Resolved | Observer синхронный, one-shot, categorized и покрыт lease. |
| F-08 | Accepted limitation | Permanently blocked Cleanup сохраняет accounting активным. |

## 34. Трассировка TASK-REV-005

| Finding | Статус | Решение |
| --- | --- | --- |
| F-01 | Resolved | CleanupResult синхронно подтверждает effective cancellation. |
| F-02 | Resolved | Release lease является последней Runtime-owned operation без Runtime epilogue. |
| F-03 | Resolved | Каждый cleanup step имеет independent safe-invocation semantics. |
| F-04 | Resolved | Dispatcher не может abandon полную activation transaction. |
| F-05 | Resolved | Dispatcher является единственным executor attachment/activation. |
| F-06 | Resolved | Transport-independent Core не является Session в `Created`. |
| F-07 | Resolved | Attachment package-private, atomic, one-use и concurrency-defined. |
| F-08 | Resolved | Reservation остаётся у Dispatcher до readiness owner. |
| F-09 | Resolved | Explicit Stop и Runtime cancellation используют first-writer intent. |
| F-10 | Resolved | Каждый committed terminal path входит в Terminalizing. |
| F-11 | Clarified | Invocation count и effective cancellation различаются. |
| F-12 | Clarified | RequestStop обещает отсутствие Session wait, а не hard time bound. |
| F-13 | Resolved | Owner получает только bound Completion Adapter registration. |
| F-14 | Resolved | Commit атомарно возвращает один полный identity-bound bundle. |
| F-15 | Clarified | Snapshot получает совместимый immutable capability accessor. |
| F-16 | Resolved | Terminal Result содержит bounded categories без raw error. |

Ни один Blocker или High finding TASK-REV-005 не оставлен Deferred.

## 35. Трассировка TASK-REV-006

| Finding | Статус | Решение |
| --- | --- | --- |
| Ownership attachment и acknowledgement | Resolved | Successful Commit одновременно передаёт transport, cancellation, Session и execution ownership. |
| Linearization activation | Resolved | Commit возвращает либо irreversible publication, либо отсутствие committed state. |
| Ordering Terminal и lease | Resolved | Owner закрывает admission callback, выполняет unregister, drain entered calls, seal cell, достигает Terminal и затем выполняет final operation release lease; Wait гарантирует отсутствие Runtime-owned work, а не scheduler-level исчезновение goroutine. |
| Acknowledgement при panic Cleanup | Resolved | Cleanup является outward panic-safe wrapper и всегда возвращает immutable acknowledgement cancellation. |
| Late outcomes panic observer и lease | Resolved | Они изолируются как operational anomalies вне уже построенного Terminal Result. |
| Lifetime callback Runtime | Resolved by TASK-REV-011 | Owner создаёт observation только после Commit; confirmed `UnregisterAndDrain` обязателен до Terminal. |
| Observation pre-Commit | Resolved | Failure возвращается через Dispatcher/Handshake как обычная error; committed Terminal Observer не вызывается. |
| Shutdown ordering DP-002 | Resolved | DP-002 синхронизирован с совместимым с ARCH-002 ordering active Session, определённым здесь. |
| Ownership cleanup context | Resolved | Owner создаёт независимый cooperative cleanup context; его expiry не разрешает false completion. |

Ни один Blocker или High finding TASK-REV-006 не оставлен Deferred.

## 36. Трассировка TASK-REV-007

| Finding | Статус | Решение |
| --- | --- | --- |
| F-01 | Resolved | Commit является единственной publication point ownership и Registration. |
| F-02 | Resolved | Cancellation `Anomaly` достигает Terminal, но сохраняет lease активным и Wait заблокированным без retry. |
| F-03 | Resolved | Drain handlers Listener и drain owners выполняются параллельно после initiation Listener Stop. |
| F-04 | Resolved | Все источники termination конкурируют через одну causal cell; последующие sources являются bounded secondary categories. |
| F-05 | Resolved | Failure matrix явно задаёт outcomes transport, cancellation, Reservation, Registration, owner, lease и Wait. |
| F-06 | Clarified | Точные release conditions lease и boundary отсутствия Runtime work после release определены нормативно. |

Ни один Blocker или High finding TASK-REV-007 не оставлен Deferred.

## 37. Трассировка TASK-REV-008

| Finding | Статус | Решение |
| --- | --- | --- |
| F-01 — Visibility Commit до success handoff | Resolved | Commit сам является success handoff и единственной irreversible publication point. |
| F-02 — BeginShutdown фиксирует незавершённый handoff | Resolved | Commit, выигравший BeginShutdown, уже является successful handoff; проигравший Commit ничего не публикует и требует Abort. |
| F-03 — Неопределённая post-Commit lifecycle branch | Resolved | Branch удалена. Каждый post-Commit path использует normal lifecycle `Committed -> Terminalizing -> Terminal`. |
| F-04 — Отсутствующий outcome Commit/BeginShutdown в matrix | Resolved | Matrix содержит явные строки для обоих winners race и их outcomes Snapshot, ownership, lease и Wait. |
| F-05 — `accepted=true` с synchronous error | Resolved в TASK-REV-010 | Semantics generic и target production Dispatcher теперь явно разделены. |
| F-06 — Wording ordering anomaly | Resolved в TASK-REV-010 | Один Terminal order применяется к каждому committed path. |

Эта revision закрывает только F-01–F-04.

## 38. Трассировка TASK-REV-009 Commit-to-Execution

| Finding | Статус | Решение |
| --- | --- | --- |
| F-01 — Boundary execution Commit-to-owner | Resolved | Dispatcher создаёт ровно одну dormant launch goroutine. Integrated operation Commit публикует Registration и переводит её one-shot execution binding в `Committed` под той же synchronization boundary до возврата success. |
| F-02 — Timing callback Runtime cancellation | Superseded by TASK-REV-011 | Финальная модель удаляет pre-Commit registration и передаёт post-Commit установку owner. |
| F-03 — Terminal ordering cancellation anomaly | Resolved в TASK-REV-010 | Cancellation anomaly следует единому общему Terminal order. |
| F-04 — Последующий failure регистрации после transfer ownership в DP-003 | Resolved | DP-003 теперь устанавливает, что сам Commit одновременно публикует Registration и передаёт ownership Session; последующего failure регистрации не существует. |
| F-05 — Repeated Commit и bundle owner | Resolved | Repeated Commit для существующей Registration возвращает то же bound publication observation и не может создать или освободить второй execution path. |
| F-06 — Неполный outcome execution в failure matrix | Clarified | Matrix теперь задаёт общие outcomes accepted и launch path до и после Commit; несвязанные terminal columns не изменены. |
| F-07 — `accepted=true, error!=nil` общего interface | Resolved в TASK-REV-010 | Ownership semantics generic interface сохраняются, а target Dispatcher использует только `true,nil` после Commit. |

Закрыты только findings F-01, F-04 и F-05, связанные с boundary Commit-to-execution. F-06 уточнён, но не заявлен как полный redesign matrix.

## 39. Трассировка TASK-REV-010

| Finding | Статус | Решение |
| --- | --- | --- |
| F-01 — pre-Commit terminal guarantee | Resolved | Одна panic-safe boundary Dispatcher публикует non-committed один раз, join-ит dormant path, освобождает owner-local values, выполняет Abort и возвращает `accepted=false`; callback до Commit отсутствует. |
| F-02 — panic atomicity Commit | Resolved | Fallible work завершается до critical section; publication не вызывает внешний код и имеет только полный non-committed либо полный committed outcome. |
| F-03 — timing callback Runtime cancellation | Superseded by TASK-REV-011 | Owner устанавливает observation только после Commit и до Start через один race-safe contract регистрации и проверки. |
| F-04 — ordering Terminal | Resolved by TASK-REV-011 | Каждый committed terminal cause использует один порядок, и только confirmed cleanup callback разрешает Terminal. |
| F-05 — неполная Failure Matrix | Resolved by TASK-REV-011 | Matrix теперь разделяет установку, invocation и cleanup callback, Complete, Observer, state owner, lease и outcomes Wait. |
| F-06 — terminology execution binding | Clarified | Binding является узкой one-shot publication capability, а не passive value или общим lifecycle control. |
| F-07 — contract accepted/error | Resolved | Generic interface допускает true вместе с error как ownership result; target production Dispatcher создаёт только pre-Commit false с error либо successful Commit true с nil. |

Ни один Blocker или High finding TASK-REV-010 не оставлен Deferred.

## 40. Трассировка TASK-REV-011

| Finding | Статус | Решение |
| --- | --- | --- |
| F-01 — anomaly unregister против Terminal | Resolved | `UnregisterAndDrain` обязан подтвердить отсутствие будущей и вошедшей callback work до Terminal; unconfirmed cleanup остаётся Terminalizing с активными lease и Wait. |
| F-02 — join callback до Commit | Resolved | Callback Runtime до Commit отсутствует; при non-committed outcome требуется join только dormant path owner. |
| F-03 — два порядка disposal до Commit | Resolved | Единственный порядок: publication non-committed, возврат dormant path, disposal owner-local values, Abort и `accepted=false`. |
| F-04 — неполная Failure Matrix | Resolved | Отдельные строки покрывают cancellation до/после Commit, установку, invocation и cleanup callback, Complete, Observer и anomalies lease. |

Ни один Blocker или High finding TASK-REV-011 не оставлен Deferred.

## 41. Влияние на DP-003

DP-003 остаётся авторитетным для Reserve, Abort, registration identity, Lookup, mutation Complete, lifecycle Manager, BeginShutdown, Wait и semantics identity Snapshot.

Этот документ добавляет нормативные integrated Commit bundle, accounting owner-lifetime lease и совместимый accessor capability-bearing Snapshot. Эти additions не изменяют существующие linearization points identity или completion.

DP-003 не дублирует execution state, attachment, cleanup или termination behavior.

## 42. Открытые вопросы

- Точные Go-имена и private package placement Core, Commit bundle, adapters и categorized values.
- Точные enum names terminal categories.
- Test-only instrumentation для доказательства отсутствия Runtime-owned work после successful release lease.

Оставшиеся вопросы относятся к representation choices реализации и proof instrumentation. Timing callback, semantics race установки, cleanup callback, ordering Terminal, ownership semantics accepted/error и boundary publication Commit-to-execution больше не являются открытыми.

## 43. Самопроверка сложности

- Conceptual states owner: шесть — один pre-Commit state и пять committed states.
- Committed accounting dimensions: два — Registration и Owner Lifetime Lease. Reservation остаётся pre-Commit transaction accounting.
- Ownership-transfer points: одна.
- Execution-publication points: одна, внутри Commit.
- Dormant launch paths на prospective owner: один.
- Invocations Terminal Observer: один.
- Callers completion на Registration: один bound owner.
- Введённые generic supervisors или policy frameworks: ноль.

## 44. Готовность к ревью

Все Blocker и High findings TASK-REV-010 и TASK-REV-011 закрыты нормативно. Externally observable committed Registration не может существовать без одного уже eligible execution path, каждый recoverable pre-Commit path завершается без создания callback, а установка callback только owner имеет одну race-safe модель. Terminal достижим только после confirmed cleanup callback.

Accepted limitations: process termination или unrecoverable failure Go runtime, scheduler starvation, permanently blocked Session Cleanup, observer или вошедший callback, unconfirmed cleanup callback, anomaly acknowledgement cancellation и unsuccessful release lease. Anomaly lifetime callback сохраняет owner в Terminalizing; anomaly acknowledgement cancellation может достичь Terminal, но оба варианта сохраняют accounting lease и Wait активными вместо false completion.

**Решение об утверждении:** Approved.

Утверждение основано на завершённом review trail из разделов 33–40. Timing callback имеет одну модель, pre-Commit registration callback отсутствует, а Terminal имеет одно доказуемое определение. TASK-REV-013 Codex завершён решением Approved с одним неблокирующим clarity finding, TASK-REV-013 Kiro — решением Approved, а TASK-DOC-016 устранил оставшиеся clarity и synchronization findings без изменения архитектуры. Пункты раздела 42 относятся к representation реализации и proof instrumentation, а не к нерешённым архитектурным решениям. Открытых архитектурных findings уровня Blocker или High не осталось.

## 45. Ссылки

- [ARCH-001: Runtime Architectural Pattern](../architecture/ARCH-001-runtime-architectural-pattern.md)
- [ARCH-002: Runtime Foundation Freeze](../architecture/ARCH-002-runtime-foundation-freeze.md)
- [DP-001: Runtime Handshake Pipeline](DP-001-runtime-handshake-pipeline.md)
- [DP-002: Runtime Host Composition Root](DP-002-runtime-host-composition-root.md)
- [DP-003: Runtime Session Manager](DP-003-runtime-session-manager.md)
- [ADR-0004: Handshake Runtime Dependency Boundary](../adr/0004-handshake-runtime-dependencies.md)

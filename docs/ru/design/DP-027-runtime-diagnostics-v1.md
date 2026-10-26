# DP-027: Runtime Diagnostics v1

[English version](../../en/design/DP-027-runtime-diagnostics-v1.md)

## 1. Статус и scope

- **Design Status:** Draft.
- **Implementation Status:** Partial — mailbox (TASK-110), component hooks (TASK-112), optional Host/Bootstrap wiring (TASK-113), isolated consumer core C1 (TASK-115); production sink/exposure Planned.
- **Целевая веха:** Beta, Gate 11; этот документ не закрывает Gate 11.
- **Задача:** [TASK-109](../../tasks/TASK-109-RUNTIME-DIAGNOSTICS-V1-DESIGN.md).
- **Source baseline:** `67ca1b200d42fda5be8b6dc8741d8f0bc24bd2a6`.

Исходный deliverable TASK-109 — design/readiness evidence без нормативного Approval.
TASK-110 опубликовала isolated `internal/runtimediagnostics` mailbox.
[TASK-112](../../tasks/TASK-112-COMPONENT-LOCAL-DIAGNOSTICS.md) реализует два
component-local source hooks и copied Producer construction по отдельной Owner
authorization: unexpected Listener Serve return и Handshake Provider operational
failure. Package schema/fixed64 loss contract неизменен.

[TASK-113](../../tasks/TASK-113-RUNTIME-DIAGNOSTICS-WIRING.md) реализует optional
copied DP-009 Producer binding и private Host/Bootstrap passage в оба существующих
component constructors. Zero отключает diagnostics. Одна external mailbox pair
принадлежит одному Host lifetime; Reader внешний при rollback/retry/Stop.
Production sink, transport, Metrics и durable delivery отсутствуют. Design остаётся
Draft/Partial; Gate11 OPEN. TASK-112 даёт I01–I12, TASK-113 — W01–W05 wiring proofs
с exact limitations в evidence envelope. I06 narrow offer/timeout-callback
coexistence остаётся compositional source-order proof.

Consumer/delivery refinement: [TASK-114](../../tasks/TASK-114-RUNTIME-DIAGNOSTICS-CONSUMER-DESIGN.md),
разделы20–26 ниже, сохраняют historical design. TASK-115 отдельно активирует
и реализует только C1; раздел27 фиксирует текущие implementation boundaries.

Integration design: [TASK-111](../../tasks/TASK-111-RUNTIME-DIAGNOSTICS-INTEGRATION-DESIGN.md),
baseline `060cf22cf040a2e84d49c52095af9ed4ab9390a7`. Раздел15 сохраняет historical
source inventory; разделы16–19 отделяют implemented component APIs от later wiring.

## 2. Authority и invariants

[ADR-0002](../adr/0002-configuration-dsl.md) и
[ADR-0003](../adr/0003-runtime-architecture.md) сохраняют Configuration First,
явные dependencies и transport-neutral Authentication.
[ADR-0004](../adr/0004-handshake-runtime-dependencies.md) ограничивает Host capability
bundle admission и Runtime context. Отдельная diagnostic dependency не добавляется
в этот bundle и не становится service locator.
[ARCH-002](../architecture/ARCH-002-runtime-foundation-freeze.md) сохраняет Host
startup/rollback, readiness, admission, root cancellation и Stop.
[ARCH-003](../architecture/ARCH-003-runtime-migration-revision.md) и Approved
[DP-003](DP-003-runtime-session-manager.md)/[DP-004](DP-004-per-session-execution-boundary.md)
сохраняют Session ownership, terminal ordering, Observer и Manager accounting.

Reporting наблюдает source fact. Он не может останавливать/перезапускать Runtime,
менять readiness, открывать/закрывать admission, создавать Session, забирать
переданное соединение, публиковать Stopped или подтверждать cleanup completion.
Approved/Frozen amendment не предлагается; требующая его реализация должна
остановиться для решения Owner.

## 3. Source-backed inventory baseline

| Path / boundary | Текущий факт | Diagnostic classification / coverage |
| --- | --- | --- |
| [Listener.serve](../../../internal/listener/listener.go), после возврата `serveHTTP` | nil, `errors.Is(http.ErrServerClosed)` и `errors.Is(net.ErrClosed)` подавляются | Normal/closed Serve return; нового report нет |
| Тот же метод, неожиданный nonnil return | Один вызов `reportListenerTerminalError` до отложенного `wg.Done` | Eligible `ListenerServeFailed`; unexpected failure остаётся eligible при concurrent Stop |
| `listenerServeError`, тот же файл | Фиксированный Error text, но Unwrap сохраняет exact raw cause | Trusted legacy identity, не operator-safe report |
| [Listener Bootstrap](../../../internal/listener/bootstrap.go) | Injected callback синхронный, обязан быстро возвращаться; panic изолирован | Существующий seam не ограничивает arbitrary blocking sink |
| [Handshake.ServeHTTP](../../../internal/handshake/handler.go), initial admission/context checks | Closed admission или unavailable/canceled Runtime context дают 503 до Authentication | Admission/cancellation decision, не Provider operational failure |
| [Authentication Service](../../../internal/authentication/service.go) | Ordered Providers, первый success или error; error возвращается неизменным | Handshake владеет одним request outcome; дублирование reporting каждым Provider запрещено |
| [API Key](../../../internal/authentication/apikey_provider.go) | Missing/wrong credentials дают reject с nil error; Resolver error возвращает wrapped cause | Credential rejection отдельно от operational dependency failure |
| [JWT](../../../internal/authentication/jwt_provider.go) | Invalid token/signature/claims дают reject с nil error; key resolution error возвращает error | То же разделение; token, claim и key details не входят в reporting |
| Handshake Authentication error branch | Вызывает `rejectCanceledHandshake`; generic 503, нет Provider error report | Coverage gap для Provider failure при active context |
| Handshake false/invalid successful result | Существующий 401; нет Upgrade и Session | Credential/decision rejection, нового operational report нет |
| `rejectCanceledHandshake` | Только configured timeout cause вызывает legacy timeout callback; прочая cancellation без callback | Cancellation/timeout отдельно; новый Provider event не заменяет timeout callback |
| Accept и handoff paths | Accept error возвращается; handoff error может вызвать callback; failed handoff очищает owned connection | Отложенная traffic/Dispatcher coverage; ownership неизменно |
| [Host.NewHost](../../../internal/runtime/host.go) | Default reporter nil; private composition передаёт reporter | Default construction не является active safe reporting sink |
| [Bootstrap bindings](../../../internal/runtime/bootstrap.go) | `DependencyBindings.TerminalErrorReporter` разрешает explicit legacy injection | Injection уже существует; raw callback не является новым safe sink adapter |
| [Runtime composition](../../../internal/runtime/composition.go) | Один callback передаётся Listener/Handshake; `terminalObserver.Observe` пустой | Session terminal observation вне этого slice |

Legacy callbacks перехватывают panic, но могут блокироваться. Зависший Listener
reporter задерживает serve return/Stop wait; Handshake reporter — handler cleanup.
Новый safe mailbox не делает arbitrary legacy callbacks неблокирующими задним числом.

## 4. Eligibility и cancellation precedence

Listener точно сохраняет существующий closed-error filter. Для eligible unexpected
Serve return владелец делает не более одного safe offer после возврата и до выхода
serve goroutine. Начало Stop не превращает unexpected return в normal shutdown.
Выводы о liveness или supervision из этого не следуют.

Непосредственно после возврата `Authenticate` Handshake ровно один раз наблюдает
`ctx.Err()` для diagnostic eligibility до существующего rejection path:

| Sample / result | Предлагаемый новый report | Сохраняемый outcome |
| --- | --- | --- |
| sampled context error nonnil, любой returned result/error | Нет: cancellation/timeout имеют precedence | Existing cancellation/503 path и отдельное legacy timeout behavior |
| sampled context active, returned error nonnil | Один offer `AuthenticationProviderFailed` | Existing generic 503, нет Upgrade/Session |
| sampled context active, nil error, rejected/invalid successful result | Нет | Existing 401 с сохранением последующих cancellation checks |
| sampled context active, nil error, valid success | Нет | Existing admission/cancellation/Upgrade/handoff checks |

Если Provider возвращает `context.Canceled` или `DeadlineExceeded` при active
boundary context, это operational failure вызова; один error sentinel не доказывает
caller cancellation. Не нужны обход Provider cause chain и разбор message: eligibility
определяют только nil/non-nil error и context sample. Cancellation после sample
не отзывает eligible fact. Sample не заменяет lifecycle/admission checks и не
меняет HTTP rules.

Cancellation precedence намеренно консервативна: независимый Provider failure,
вернувшийся после boundary cancellation, опускается, а не считается доказанно
отсутствующим или успешным. Полная late-failure causality отложена. Это lossy
diagnostics, а не исчерпывающий audit всех Provider failures.

## 5. Safe report и bounded identity

Каждый copied report содержит ровно три scalar fields:

| Field | Допустимые значения / смысл |
| --- | --- |
| SchemaVersion | Фиксированное `1` |
| Code | Closed enum: `ListenerServeFailed`, `AuthenticationProviderFailed` |
| Sequence | Positive uint64, уникальный только внутри mailbox lifetime; accepted insertion order |

Severity при отображении consumer — константа `Error`, полученная из Code, а не
дополнительное free-form field. Code — safe semantic identity, не raw Go error:
нет `Unwrap`, `errors.Is` chain и cause reference. Исходные errors неизменны на
существующих trusted owner boundaries; projection их не форматирует и не хранит.
Существующий Listener closed filtering остаётся owner-local.

В records нет strings, maps, slices, interfaces, pointers, error/context,
request/response, Principal и capability. Исключены Provider name/type/index,
route/session/Workspace IDs, Snapshot/provenance, timestamps, host/address/PID и
user-supplied labels. Arbitrary labels и key/value bags отсутствуют. У Code ровно
два значения; Sequence никогда не становится metric label. Unknown schema/code
и invalid record отклоняются без raw fallback text.

## 6. Mailbox и nonwaiting operations

Один inert internal mailbox принадлежит одному explicit composition lifetime.
Он хранит fixed ring из **64** safe report values и bounded scalar state. Capacity —
фиксированная v1 implementation bound, а не новая user Configuration setting.
Producer capability предлагает только два closed codes; reader удаляет safe copies.
Ни одна capability не даёт Runtime lifecycle access или cross-mailbox lookup.

Unbound/nil producer отключён и не делает offer; unbound/nil reader возвращает
no-item. Mailbox state не копируется после использования; capabilities относятся
только к исходному mailbox. Initial sequence state равен нулю; первая insertion — 1.

Каждый offer/read защищён одной попыткой `sync.Mutex.TryLock`. Busy означает
немедленный drop/no-item; нет lock wait, spin, retry и I/O. Под lock обрабатывается
только один fixed-size record и scalar indexes/count/sequence. Там не выполняются
callback, formatting, raw error operation, allocation, external code и blocking
operation. Empty read возвращает no-item; full offer отбрасывает новый record и
сохраняет имеющийся FIFO. Invalid code отбрасывается без расхода sequence.

Sequence назначается только при accepted insertion под тем же lock. После
`MaxUint64` offers fail closed как exhausted; нет wrap, reset и reuse. Удалённый
slot очищается. Successful read возвращает одно value без сохранения caller
memory. Capacity 64 ограничивает retained reports; worker goroutines, timers и
retry queues не создаются. Nonwaiting описывает algorithmic dependency, а не
hard real-time scheduling guarantee. Contention допускает starvation/loss.

## 7. Ownership, timing, ordering и deduplication

Source owner определяет eligibility и делает один offer на source occurrence.
Provider, Authentication Service, Listener callback adapter и Handshake не могут
каждый выдавать копию одного occurrence. Scope Listener — один Serve invocation;
scope Handshake — один `Authenticate` return одного request. Source integration
обязана обеспечить это структурно; mailbox не выполняет deduplication.

Accepted insertion — publication point safe observation. FIFO removal и Sequence
упорядочивают только accepted insertions, а не wall-clock chronology failures,
HTTP response delivery или domain revisions. Иного global ordering concurrent
occurrences нет. Два failures с одним raw error остаются разными occurrences.

Несколько readers деструктивно делят records. Это не fanout и не durable replay.
Consumer может использовать `(exact mailbox capability, Sequence)` только для
local duplicate recognition; globally comparable ID не предоставляется. Busy,
full, invalid и exhausted offers означают loss, а не разрешение retry/replay
Authentication/Serve. Не вводятся drop counters, reliable delivery, at-least-once
или end-to-end exactly-once guarantee.

## 8. Sink behavior и trust boundary

Runtime producers только предлагают safe values concrete mailbox. Они никогда не
вызывают arbitrary sink, logger, callback, serializer, channel consumer или external
delivery interface. Consumer сначала удаляет copied value, освобождает mailbox
lock и только затем может вызвать собственный sink. Raw cause не пересекает boundary.

Slow, blocking, failing или panicking sink влияет только на caller consumer.
Без reads накапливается максимум 64 records, затем новые offers отбрасываются.
Consumed record после failure автоматически не возвращается. Sink retry, recursion,
recursive diagnostic failure reports и delivery workers не предоставляются.
External consumer владеет собственным execution/resources; этот design их не
создаёт и не обещает termination arbitrary sink code. Будущая sink integration
требует собственного lifecycle/authorization/delivery design до production wiring.

## 9. Cancellation, Stop и Session compatibility

Mailbox construction не открывает socket и не запускает goroutine. Failed startup
может отбросить mailbox и queued records. Host Stop не закрывает, не flush/drain и
не join-ит mailbox; новый Stop hook, observer wait и shutdown error не добавляются.
Explicitly retained reader может читать после Stop. Source return после начала
Stop всё ещё может предложить eligible historical failure; это не доказывает
current liveness или completion Host/Listener/Session cleanup.

Две live Host capabilities ADR-0004 неизменны; будущая mailbox dependency отдельная,
явная и narrowly observation-only. Sink не вызывается под Host, Manager, Session
и execution-owner locks. Synchronous Approved DP-004 Observer не переиспользуется:
его blocking/lease accounting остаётся точно прежним. Session terminal observation
и shutdown diagnostics отложены. Existing legacy callbacks остаются trusted
prompt-return interfaces; блокирующие их consumers вне новой producer-isolation
guarantee.

## 10. Связь с DP-018

[DP-018](DP-018-runtime-operational-error-reporting-redaction.md) — Approved / Planned
для management, activation и recovery facts. Его scoped authorization,
owner/phase precedence, durable fact correlation и projection-version replay keys
не становятся traffic semantics автоматически. Failed HTTP Serve не доказывает
`ExecutionLost`, `StartupFailed` или `ShutdownFailed`; Provider dependency failure
не является management `AuthorizationUnavailable`.

Общие только совместимые принципы: сохранить owner truth, конструировать по
allowlist, исключить raw causes/credentials и не давать delivery менять domain state.
Design использует два отдельных traffic/runtime codes и mailbox-local occurrence
ordering, а не DP-018 report keys/categories. Не подразумеваются DP-018 implementation/
status change и authorization bypass. Reader capability — trusted internal
dependency, а не public management/API authorization model.

## 11. Verification Matrix будущей реализации

Это требования, **не выполненные тесты и не implementation evidence**.

| ID | Proof scenario | Требуемый результат |
| --- | --- | --- |
| V01 | Listener nil и wrapped closed sentinels | Нет нового report; existing filter сохранён |
| V02 | Unexpected Serve return, включая concurrent Stop | Один offer на invocation; lifecycle/readiness неизменны |
| V03 | Active context с arbitrary Provider error или context-shaped error | Один Provider code, generic 503, ноль Upgrade/handoff |
| V04 | Missing/invalid credentials и invalid successful result | Нет operational report; existing rejection behavior |
| V05 | Cancellation/timeout до sample, late result/error; cancellation после sample | Точная precedence раздела 4, без retraction и duplicate timeout Provider code |
| V06 | Error с panicking/blocking Error/Unwrap/Is; hostile request/ProviderName | Provider projection их не вызывает; records без sensitive input; Listener сохраняет только pre-existing owner filter |
| V07 | Empty, capacity 64, full и forced contention | Immediate explicit loss/no-item; FIFO сохранён, нет retry/spin |
| V08 | Concurrent producers/readers, reader mutation, independent mailboxes | Race-free copies, unique positive insertion sequence, нет cross-mailbox state |
| V09 | Sequence на MaxUint64 и следующий offer | Последнее value допустимо, следующие offers drop; нет wrap/reset/reuse |
| V10 | Consumer перестаёт читать, блокируется, panic/error после removal | Producer progress/cleanup независимы; <=64 retained; нет worker/reinsert/recursion |
| V11 | Construct/discard, failed startup, Stop с pending/late records | Нет goroutine/I/O/flush/join/new lifecycle result; post-Stop read допустим |
| V12 | Legacy callbacks и Approved Session Observer regressions | Existing owner/cause/accounting contracts неизменны; новый safe sink туда не подключён |
| V13 | Unknown enum/schema и bounds inspection | Closed-model violation отклоняется; нет raw fallback/arbitrary labels |
| V14 | Architecture/source/import/scope checks | Нет service locator, Control Plane dependency, production wiring и Approved/Frozen amendment |

Применимая future verification: unit/concurrency tests, `go test -race`, stress и
module/build/vet checks exact authorized slice. V01–V06 и Host integration parts
V11–V12 относятся к последующей source integration, а не к proof из isolated
mailbox test. Недоступный test нельзя считать PASS.

## 12. Изолированный substrate prerequisite (TASK-110)

Исторический prerequisite опубликован как TASK110; текущий next slice — раздел19.

Рекомендован **один isolated package `internal/runtimediagnostics` mailbox с
тестами**, без production imports/call sites/source wiring. Readiness обусловлена
fresh exact-baseline intake и отдельной Owner authorization, принимающей этот
Draft design как требования bounded task; acceptance здесь не является permission.

Acceptance criteria:

1. Three-scalar safe report и ровно два codes; нет raw-input constructor и sink
   callback; ненужный exported surface остаётся private.
2. Fixed 64-slot ring; один try-lock, нет allocation/external work в critical
   section, waiting/retries; slot clearing и truthful loss/no-item outcomes.
3. Accepted-insertion FIFO sequence и saturation из разделов 6–7.
4. Ноль worker/timer/I/O/lifecycle hook; copied reader results; mailboxes изолированы;
   sink может выполняться только после удаления value consumer.
5. Доказать V07–V10, V13–V14 и package-local V11/V12 focused tests, race/stress;
   никаких production/test modifications вне package и обычных task documents.

Это prerequisite, а не operational reporting delivery. Listener и Handshake binding —
последующая отдельно авторизуемая работа; её tasks доказывают source classification
и сохраняют legacy callbacks, не обещая sink safety для них.

## 13. Deferred coverage и non-goals

Явные gaps: Dispatcher; Handler; Session terminal anomalies (включая late
Observer/lease anomalies); shutdown diagnostics; Metrics integration. Также отложены:
configured Handshake/TLS timeout reporting, Accept/TLS peer rejection taxonomy,
полная late Provider failure causality, production composition/wiring, sink
adapters и consumer lifecycle, public API, access policy, durable audit/replay,
sampling/drop counters, supervision, restart и automatic remediation.

Не входят production code/test/PoC, TASK-102 build/run/change, Architecture B
activation, Approved/Frozen amendment, Secret backend, новая Configuration или
deployment feature. Gate 9 closure неизменен и не означает production readiness.

## 14. Existing evidence и design readiness

Изучены test source, без повторного запуска в этой design task:

- [Listener tests](../../../internal/listener/serve_error_test.go):
  `TestListenerExpectedServeShutdownIsNotReported`,
  `TestListenerExpectedClosedServeErrorsAreNotReported`,
  `TestListenerUnexpectedServeErrorIsReportedOnce`,
  `TestListenerConcurrentStopAndServeFailureReportsOnce`,
  `TestListenerServeErrorReporterPanicDoesNotChangeLifecycle`.
- [Handshake tests](../../../internal/handshake/handler_test.go):
  `TestHandlerAuthenticationRejectSkipsUpgradeAndSession`,
  `TestHandlerAuthenticationErrorSkipsUpgradeAndSession`,
  `TestHandlerConfiguredDeadlinePreventsUpgrade`,
  `TestHandlerReporterPanicDoesNotChangeHandoffCleanup`.
- [Runtime tests](../../../internal/runtime/bootstrap_test.go):
  `TestBootstrapHostAuthenticationErrorPreventsUpgrade`,
  `TestBootstrapTerminalObserverConsumesSessionError`.

TASK109 проверила первоначальный design. TASK110 затем реализовала/опубликовала
isolated mailbox через PR110 на integration baseline; PR/post-merge Linux CI
прошли race/tidy/module guards. Эти historical facts не доказывают новую
integration. TASK111 проверила только design разделов15–19; на том baseline source integration, consumer/delivery и Gate11 closure отсутствовали. Текущий component status TASK112 указан в разделе1.

## 15. Source integration inventory на baseline TASK-111

Это только Design-update / Readiness на `060cf22cf040a2e84d49c52095af9ed4ab9390a7`.
Source и test code не меняются. Mailbox опубликован; integrations ниже Planned.
Source anchors относятся к этому exact baseline.

| Source boundary | Текущий факт | Exact planned producer point |
| --- | --- | --- |
| [Listener](../../../internal/listener/listener.go), serve, строки 212–218 | Один Serve call, closed-error filter215, legacy callback216, deferred wg.Done213 | Внутри неизменённой eligible branch один Offer(ListenerServeFailed) перед legacy callback |
| Тот же файл, wrapper/reporter220–242 и Stop185–186 | Unwrap сохраняет raw cause; synchronous panic-isolated callback до Serve completion; Stop ждёт Serve/handlers | Сохранить filter/wrapper/callback/waits; cause не передаётся Offer |
| [Listener Bootstrap](../../../internal/listener/bootstrap.go), 16–97 | Constructors связывают dependencies; Build копирует их без socket | Скопировать concrete Producer value до Start; Build не делает offer |
| [Handshake](../../../internal/handshake/handler.go), 107–115 | Один Authenticate; error ведёт к generic rejection; последующий context check отдельный | Сразу после Authenticate один diagnostic-only handshakeContext.Err sample; nonnil error при active sample допускает один Provider offer |
| Тот же файл, 116–163 | Уже есть decision/final admission/context/Accept/handoff/cleanup и configured-timeout callback | Сохранить все response/cancellation/ownership operations |
| [Authentication Service](../../../internal/authentication/service.go), 32–46 | Ordered Providers; первый error возвращается неизменным; первый success завершает evaluation | Service/Providers не report-ят; Handshake владеет одной opportunity на Authenticate return |
| [API Key](../../../internal/authentication/apikey_provider.go) / [JWT](../../../internal/authentication/jwt_provider.go) | Credential rejection возвращает nil error; resolver/key failure — operational error | Nil-error rejection не emits; active context + nonnil error eligible без обхода cause |
| [Runtime composition](../../../internal/runtime/composition.go), 153–181 | Host создаёт обе boundaries с одним legacy callback; terminalObserver34 no-op | Позднее передать Producer отдельно от двух ADR004 capabilities; Session Observer не использовать |

Listen/build/TLS preparation errors, expected closed Serve returns, TLS peer
rejections, Stop errors, Accept/handoff failures и Session outcomes не добавляют
ни одного code. Management categories/report keys DP018 не переиспользуются.

## 16. Exact value dependency и construction design

TASK-113 implementation note: полные value passage/lifetime requirements ниже
реализованы по отдельной Owner authorization. Prospective wording и W-row labels
«Later wiring» сохраняют хронологию design TASK-111, а не отсутствие Host wiring.
Исходный scope TASK-112 — historical. Source hooks/delivery capability/normative
status здесь не повышаются.


Следующие component APIs реализованы TASK112; later Host passage остаётся Planned:

- `DefaultBootstrap.WithDiagnosticsProducer(producer runtimediagnostics.Producer) DefaultBootstrap` возвращает копию Listener Bootstrap с private Producer value. Existing constructors/Build interface неизменны; Build копирует value в DefaultListener без offer и новой validation.
- `handshake.NewHandlerWithDiagnostics(admission AdmissionCapability, runtimeContext RuntimeContextProvider, authenticationService authentication.Service, handoff connection.AuthenticatedDispatcher, reportError func(error), producer runtimediagnostics.Producer) (*Handler, error)` хранит private Producer value. Existing constructors сохраняют signatures и порядок required-dependency validation, делегируя с zero Producer.

Zero отключает observation; fallback mailbox не создаётся. Caller один раз
разыменовывает nonnil Producer pointer из mailbox New на construction boundary.
Копируется только opaque capability, не mailbox/mutex state. Последующее присвоение
caller Producer object zero/другой capability не перенаправляет созданный component.
Concurrent mutation caller во время construction copy остаётся unsupported misuse.
Не добавляются runtime setter, diagnostic interface/callback или raw-error constructor.

Полный будущий passage явный и **Planned**, не входит в следующий slice:

1. External composition owner создаёт одну mailbox pair на один Host construction/lifetime и удерживает Reader снаружи. Разные Host lifetimes требуют отдельных mailboxes; нет implicit global sharing или Host identity field.
2. Будущий fixed `DependencyBindings.DiagnosticsProducer runtimediagnostics.Producer` optional, zero absent. Bootstrap один раз snapshot-ит value; не создаёт/не владеет mailbox и не вызывает Offer/TryRead. Caller не меняет borrowed request/dependencies, пока Bootstrap их читает.
3. Private Host-construction core получает/хранит value и передаёт его через private production composition в оба component constructors. Public Host interface/NewHost signature и existing zero-producer behavior неизменны.
4. Host остаётся единственным operational composition root во время Start. Reader/getter/consumer не попадает в Host, Bootstrap outcomes, Snapshot, Configuration, launch identity или ADR004 admission/context carrier. Новая dependency импортирует только leaf mailbox package.

[Draft DP009](DP-009-runtime-bootstrap-contract.md) §11 явно отделяет этот planned
четвёртый binding от трёх текущих implemented bindings. Static validation/failure
precedence, fixed factory ownership, one Host/Build/Start, wrapping causes и cleanup
ownership неизменны. Approved/Frozen amendment не нужен; потребовавшая его
implementation останавливается для Owner decision.

## 17. Exact source algorithm и compatibility

Listener добавляет ровно один concrete Offer в existing nonnil/unexpected Serve-error
branch перед reportListenerTerminalError; bool игнорируется. Existing errors.Is
filter/raw-cause wrapper сохраняются. Concurrent Stop не меняет classification.
Blocked legacy callback не мешает предшествующей offer attempt, но по-прежнему
задерживает wg.Done/Stop. Existing filter сам может вызвать hostile Is/Unwrap,
block или panic; новая guarantee против этого не вводится.

Handshake добавляет ровно один diagnostic-only context sample сразу после возврата
Authenticate. Если returned error nonnil и sample nil, одна попытка
Offer(AuthenticationProviderFailed), bool игнорируется. Затем existing error branch,
последующие context checks, decision, admission, Accept и handoff выполняются
неизменно. Sample никогда не заменяет эти checks.

| Diagnostic sample / Authenticate return | Новая Provider opportunity | Сохраняемый existing flow |
| --- | --- | --- |
| Cancelled/timed out sample; любой result/error | Нет, включая independent late dependency failure | Existing cancellation/error rejection и legacy configured-timeout handling |
| Active sample; nonnil error, включая direct/wrapped context sentinels | Один offer; не вызывать Error/Unwrap/Is | Existing generic503; нет Upgrade/handoff |
| Active sample; nil error и reject/invalid success | Нет | Existing401, если subsequent existing cancellation checks не выиграли |
| Active sample; nil error и valid success | Нет | Все existing admission/context/Upgrade/handoff checks |

Cancellation после active sample не отзывает eligible offer. Если configured timeout
позднее виден existing rejection helper, его legacy timeout callback может сработать;
это не duplicate Provider code. Cancellation до sample намеренно теряет late failure,
не доказывая его отсутствие/success. Provider error sentinel не доказывает caller
cancellation.

Ни один source не передаёт raw errors, request/Provider metadata, credentials,
Principal или labels. Mailbox получает только fixed Code и назначает version и
accepted-insertion sequence. Service, Providers, callback adapters и generic legacy
report helpers не дублируют offers. На occurrence не более одной attempt; reliable
delivery/exactly-once acceptance не гарантируются. Loss не вызывает retry/replay,
recursion/fallback reporting и не меняет response/lifecycle.

## 18. Lifetime, startup и Stop

Будущий mailbox lifetime — один Host lifetime, включая failed Start/rollback/retry
того же Host. Не reset/replace/drain/flush при retry; accepted historical events
сохраняют sequence. Failed lifetime может отбросить mailbox целиком. Retained Reader
может читать после rollback/Stop; event не доказывает attempt identity, readiness,
startup success, current liveness или cleanup.

Static binding/Build, failed Listen/TLS preparation не emits codes. Если реальный
Serve invocation fails до startup commit или при rollback/Stop, eligibility
определяет обычный source filter. Pre-commit unavailable Runtime context/admission
по-прежнему предотвращают Authentication и Provider reporting.

Не добавляются resource/worker/sink/timer/Start/Stop/Close/reset/flush/drain/join или
shutdown error. Source offer не получает Host/Listener lifecycle, Manager, Session
или execution-owner lock; mailbox TryLock завершается до legacy callbacks и дальнейшей
handler work. Blocked external consumer влияет лишь на consumer. Existing arbitrary
legacy callbacks/Providers всё ещё могут блокировать cleanup; новая hard termination
или scheduling latency guarantee отсутствует. Approved Session Observer/lease/
accounting, Host startup/rollback/Stop order/readiness/admission и errors.Is-compatible
cause chains сохраняются.

## 19. Exact future verification и минимальный следующий slice

Это требования будущих proofs, **не tests, выполненные TASK111**. Component slice
использует два focused diagnostic test files Listener/Handshake; existing tests
остаются regression coverage. W rows относятся только к later separately selected
wiring slice. I rows уточняют source portions V01–V14; existing mailbox proofs не
заменяют source proofs.

| ID / proposed proof | Exact required outcome | Slice |
| --- | --- | --- |
| I01 TestListenerDiagnosticsExpectedServeReturns | nil, direct/wrapped http.ErrServerClosed/net.ErrClosed: нет event, legacy behavior unchanged; wrapped cases дополняют direct-sentinel coverage | Component |
| I02 TestListenerDiagnosticsUnexpectedServeOnce | Один unexpected return, включая concurrent Stop: одна opportunity/code, safe schema/positive sequence; legacy callback once, errors.Is(callbackErr,rawCause) true | Component |
| I03 TestListenerDiagnosticsBeforeLegacyReporter | Block controllable legacy reporter: prior event removable; release до cleanup assertions. Reporter panic isolated; blocked legacy не обещает Stop completion | Component |
| I04 TestHandshakeDiagnosticsOperationalFailure | Active sample + arbitrary/direct/wrapped context-shaped error: один Provider code, generic503, zero Upgrade/handoff; ordered multi-Provider service даёт одну boundary opportunity | Component |
| I05 TestHandshakeDiagnosticsCredentialRejection | Nil-error credential reject/invalid success: нет code, existing401/cancellation precedence; valid success тоже без code | Component |
| I06 TestHandshakeDiagnosticsCancellationPrecedence | Cancel/timeout до sample + late result/error: нет Provider event, existing503/timeout callback. Cancellation после active sample не retracts; legacy timeout callback может coexist. Deterministic phase-controlled fixtures, не timing sleeps | Component |
| I07 TestHandshakeDiagnosticsDoNotInspectProviderError | Hostile Error/Unwrap/Is и request/Provider metadata не inspected projection; ровно три safe scalars. Listener existing filter исключён из non-invocation guarantee | Component |
| I08 TestDiagnosticsSaturationPreservesSourceBehavior | Prefill real mailbox64, каждый eligible source: FIFO/HTTP/legacy behavior сохраняются при full loss, без retry; zero Producer сохраняет legacy behavior | Component |
| I09 TestDiagnosticsCaptureProducerValue | Bind P1 value, overwrite caller object P2/zero, trigger: только original Reader1; отдельные mailbox instances не пересекаются | Component |
| I10 TestDiagnosticsConcurrentSourcesAndReaders | Valid copies, positive unique accepted sequence/per-reader FIFO, accepted count <= eligible occurrences, loss/no duplicates/cross-mailbox; future race/stress и legacy regressions | Component |
| I11 TestDiagnosticsDoNotAddOutcomeKinds | Accept/handoff/Session/startup/Stop failures не добавляют codes; existing handoff/reporter cleanup/errors.Is behavior unchanged | Component |
| I12 source/import/API audit + existing mailbox proofs | Один concrete Offer на eligible branch, bool ignored; нет raw input/adapter/counter/retry/sink/worker. Busy/exhaustion compatibility compositional: mailbox white-box proofs + source structure. Нет mailbox test APIs и ложного direct forcing inaccessible state | Component |
| W01 TestBootstrapDiagnosticsBindingSnapshot | Optional value copied once; request/validation precedence unchanged; no Offer/TryRead during binding; caller reassignment не redirects Host | Later wiring |
| W02 TestHostDiagnosticsSharedProducerIsolation | Один Host использует один mailbox для обоих sources; independent Hosts separate pairs; Reader не хранится/не exposed Host | Later wiring |
| W03 TestHostDiagnosticsRollbackRetryLifetime | Composition/startup/rollback/retry сохраняют lifecycle/causes и одну sequence; old events retained, no reset/flush/new resource | Later wiring |
| W04 TestHostDiagnosticsStopIndependence | При nil/prompt legacy reporter saturated mailbox/blocked external consumer не меняет Host Stop order; retained reader/late eligible Listener event работают; legacy blocking exclusions сохраняются | Later wiring |
| W05 Host/Session contract audit + existing regressions | Host interface/NewHost unchanged; ADR004 carrier2; Observer/accounting unchanged; no ControlPlane cycle/sink/Metrics/API/durable transport/Configuration field | Later wiring |

Future regression включает пять Listener serve_error tests, Handshake rejection/
error/deadline/admission/handoff/panic/concurrency tests, Host rollback/retry/shutdown,
Bootstrap optional-binding/cause и Approved Session/Manager tests. Implementation
verification требует formatting/full test/build/vet/module checks, доступные race
и stress; недоступные checks — явные limitations. Linux race не квалифицирует
Windows/native TASK102.

**Component slice TASK112:** реализованный scope — только component-local
hooks/construction: listener.go, listener/bootstrap.go, handshake/handler.go, два
focused package test files и необходимые task/status docs. Без runtime,
Authentication, Session/Manager, mailbox/module edits. Production composition
сохраняет legacy constructors/zero Producer; Host/Bootstrap wiring не активируется.

Later value-binding/Host/private composition slice требует independently accepted
prerequisite, fresh readiness/Size Guard и отдельную Owner authorization. Ни один
slice не вводит sink/consumer lifecycle. Acceptance этого design не является
нормативным DP Approval, implementation permission или Gate11 closure.
**STOP BEFORE COMMIT GATE AND IMPLEMENTATION AUTHORIZATION.**

## 20. Consumer / delivery readiness (TASK-114)

[TASK-114](../../tasks/TASK-114-RUNTIME-DIAGNOSTICS-CONSUMER-DESIGN.md) —
design/readiness addendum на baseline `fd739a87a602229293c41f50c23a7854b2b5e90f`.
На closure TASK114 разделы20–26 были **proposed / Planned**, не действующий API,
implementation permission, normative Approval или выполненный consumer proof.
TASK115 отдельно разрешает C1 и реализует разделы21–24 без amendment; раздел27
и task envelope задают current proofs. Design остаётся Draft/Partial.
Historical component/wiring scopes и evidence выше сохранены.

MASTER_PLAN §5 Gate11 требует reporting server, Dispatcher, Provider, Handler и
shutdown failures с Secret/credential redaction. Конкретный built-in logger или
transport не предписан; §9 требует не связывать компоненты с одним logging
implementation. Bounded core достаточен как следующий независимо проверяемый
consumer prerequisite без production sink в этом slice. Isolated consumer с test
sinks **недостаточен для полного Gate11**: отсутствуют actual sanctioned operator
exposure, scheduling/access/lifecycle evidence и Dispatcher/Handler/shutdown
coverage. Поздний exposure design может выбрать caller-supplied sink; built-in
production logger не является обязательной формой. Сокращённый критерий завершения
Gate11-v1 не вводится неявно. Gate11 OPEN.

## 21. Proposed consumer API и ownership

Будущий leaf-layer package `internal/runtimediagnosticsconsumer` импортирует только
стандартные `context`/`sync` и существующий `internal/runtimediagnostics`. Sink
execution вне mailbox package сохраняет его текущий package godoc и все
существующие source files. Новый module/dependency не предлагается.

Предлагаемые signatures, не текущий Go API:

```go
type Sink func(runtimediagnostics.Event) error
func NewConsumer(reader runtimediagnostics.Reader, sink Sink) Consumer
func (consumer Consumer) Drain(ctx context.Context, budget uint8) DrainResult
```

Consumer — opaque copyable value с private pointer на shared state: copied Reader,
captured Sink function и один private TryLock guard. Копии разделяют state/guard;
mailbox/mutex state не копируется. Construction не выполняет read/offer/callback/
I/O/timer/worker/lifecycle operation. Zero Reader (comparable concrete value) или
nil Sink даёт zero/disabled Consumer. Nil Reader не представим в value signature;
caller проверяет pointer до dereference. Последующие Reader/Sink reassignment не
перенаправляют созданный Consumer. Concurrent mutation construction inputs —
unsupported caller misuse.

Один trusted external composition owner сохраняет Reader/Consumer/Sink и scheduling.
External означает вне Host/producer lifecycle, а не public management access.
Один logical Reader owner на mailbox — precondition ordering. Дополнительные Reader
copies или отдельно созданные Consumers могут destructively partition records;
это не fanout, а guard нового Consumer не разделяется с другими instances.
Global registry/enforcement cross-instance ownership отсутствуют.

Один TryLock attempt допускает Drain; concurrent/reentrant calls того же Consumer
или копий возвращают Busy без ожидания/read/sink. Consumer guard удерживается во
время synchronous sink call и освобождается defer на возвращающихся путях. Mailbox
lock уже отпущен; Host/Manager/Session locks не берутся. Blocked sink задерживает
только своего caller и делает этот Consumer Busy.

## 22. Exact turn algorithm и scalar receipt

Budget — 1…64 **TryRead attempts**, а не нужное число success и не безграничный
drain-until-empty. Одновременно удерживается максимум одна Event copy, без batch queue.

DrainResult содержит только bounded scalar fields:

| Field | Proposed type / exact meaning |
| --- | --- |
| ReadAttempts | uint8; фактические TryRead calls этого turn, не больше budget |
| Observed | uint8; successful valid TryRead results, не все physical removals |
| SinkSucceeded | uint8; sink invocations этого turn с nil return |
| SinkFailed | uint8; nonnil sink return, максимум один до остановки |
| SinkPanicked | uint8; recovered abnormal sink return, максимум один до остановки |
| Discarded | uint8; known valid copy, пропущенная при cancellation after read/before sink |
| Stop | closed DrainStop enum ниже, не raw cause/free-form message |

DrainStop names: Disabled, InvalidInput, Busy, NoItem, Budget, Canceled,
SinkFailed, SinkPanicked. Это local API outcomes, не новая wire schema.
Receipt не содержит mutable caller labels, strings, errors, maps или global totals.
Для каждого нормально возвращённого receipt:
`Observed = SinkSucceeded + SinkFailed + SinkPanicked + Discarded`;
все counts не больше64, ReadAttempts не меньше Observed.

Упорядоченный algorithm / precedence:

1. Disabled Consumer возвращает Disabled с нулевыми counts; context не читается.
2. Nil context или budget0/>64 возвращает InvalidInput без read/sink. Typed-nil или
   malformed context вне caller precondition; его обнаружение не обещается.
3. Одна попытка consumer guard. При неуспехе Busy, zero counts, без context sample/
   read/sink. Input validation поэтому предшествует Busy.
4. Перед каждой read attempt sample `ctx.Err()`. Nonnil останавливает Canceled без
   нового read. Увеличить ReadAttempts, затем ровно один TryRead.
5. False немедленно останавливает NoItem без retry. True увеличивает Observed.
6. Снова sample cancellation перед sink invocation. Nonnil увеличивает Discarded
   и останавливает Canceled; removed copy никогда не возвращается в mailbox.
7. Иначе вызвать Sink один раз с detached Event value. Cancellation после этого
   sample не предотвращает и не прерывает call.
8. Nonnil return увеличивает SinkFailed и останавливает SinkFailed. Recoverable
   panic увеличивает SinkPanicked и останавливает SinkPanicked. Эти факты имеют
   приоритет над более поздней cancellation. Нет retry/recursive diagnostic.
9. Nil return увеличивает SinkSucceeded, затем sample cancellation: Canceled,
   если видна; иначе продолжить или Budget после исчерпания attempts.

Caller передаёт nonnil well-formed Context: Err/Done соблюдают Context contract,
возвращаются promptly и без panic. Well-formed forwarding context допустим;
typed-nil/hostile implementation не получает safety/termination guarantee.
Sink получает только Event, без context/cancellation authority/Host capability.

Panic adapter отслеживает normal return (flag устанавливается только после Sink
return), использует deferred recover и отбрасывает payload. Nil recovered payload
не приравнивается к normal success. Запрещены Error/Unwrap/Is, formatting и retention
error/panic payload. Nonreturning callback, Goexit, fatal runtime/process termination/
resource exhaustion не получают receipt или hard termination guarantee.
Goroutine-per-call timeout/preemption wrapper не вводится.

## 23. Delivery, ordering и честное наблюдение loss

Delivery destructive и lossy: максимум один sink invocation для каждой valid copy,
полученной этим turn. SinkSucceeded означает только nil return этого вызова, а не
external observation, persistence, acknowledgment или durable delivery. Нет
at-least-once, global exactly-once, retry/reinsert/replay/reliable-delivery guarantee.
Sink failure/panic или post-read cancellation теряют эту copy; первый такой outcome
останавливает turn. Arbitrary sink side effects — ответственность его owner;
consumer не может их rollback.

При одном logical Reader owner и shared consumer guard порядок sink invocations
следует accepted-insertion FIFO. Нет global/wall-clock/source-failure/cross-mailbox
order или гарантии для asynchronously retained sink effects. Полученные Sink
copies не могут менять queued events или будущие receipts.

NoItem **не означает** empty/drained/flushed/complete/successful shutdown. Existing
TryRead false объединяет disabled/busy/empty/invalid-record cases; invalid slot
может физически удаляться/очищаться до false. Constructor-zero Reader вместо этого
даёт Disabled. Observed считает только valid successful reads; ReadAttempts не
раскрывает exact physical removals или remaining queue length.

Producer busy/full/invalid/exhaustion loss остаётся **unknown** для consumer.
Sequence присваивается только accepted insertion: source drops не создают gaps.
Gaps с несколькими readers могут означать partitioning, не измеренные drops;
отсутствие gaps не доказывает отсутствие loss. Receipt показывает только attempts,
valid copies, local sink outcomes и known post-read discard этого turn. Не вводятся
producer-drop estimate, cumulative counter, telemetry label, sampling или metrics.
False/Busy не разрешает internal retry loop; позднее scheduling — новая внешняя
opportunity, без implicit retry consumed event.

## 24. Shutdown, backpressure, redaction и compatibility

External owner прекращает scheduling при собственном shutdown. У Consumer нет
Start/Stop/Close/flush/reset/drain-to-completion/join lifecycle. Host Stop/rollback/
cleanup никогда не ждут Consumer, его guard/result/Sink. Retained Reader/Consumer
может выполнить поздний bounded turn с active caller context после Host Stop;
late eligible source facts допустимы, поэтому final emptiness/attempt/liveness/
cleanup completion из этого не выводятся.

Запрещён Drain внутри Listener/Handshake legacy callbacks, Approved Session
Observer/lease accounting, Host lifecycle methods или под Runtime locks. Это
ownership restriction необходимо: composition, сама ожидающая arbitrary sink code,
не может заявлять isolation. Existing blocking legacy callbacks/Providers и hostile
Listener error filters сохраняют текущие limitations.

Slow/blocked/failing/panicking Sink не создаёт producer backpressure. Без progress
existing mailbox capacity64 ограничивает retained records, поздние offers drops;
consumer state/receipt constant size, не новая retry queue. Sink-owned allocations,
retained copies/resources внешние, их объём этот API не ограничивает.

Sink видит только существующие allowlisted SchemaVersion/Code/Sequence Event:
не добавляются raw cause/request/credential/Principal/provider metadata/timestamp/
Host/arbitrary labels. Consumer не вводит public endpoint/authorization model,
serializer/fallback logger/raw-error adapter. Поздний exposure design должен явно
определить caller access и retention. DP018 management categories/replay keys
отдельны. Не меняются ARCH002 startup/rollback/Stop, ADR004 carrier2, DP003/004
Observer/accounting/leases, Host/NewHost, configuration, DP009 bindings,
source eligibility/cancellation и errors.Is.

## 25. Minimal slices и acceptance boundaries

| Slice | Readiness / exact boundary |
| --- | --- |
| C1 isolated consumer core | Activated отдельной Owner authorization TASK-115; implemented in isolation. Один новый internal/runtimediagnosticsconsumer package, consumer.go и consumer_test.go; без existing source/test/mailbox edits, worker, Runtime wiring или adapter. Acceptance требует разделы21–24 и C01–C16 proofs |
| C2 operational exposure/adapter | Not Ready / Not Activated. Сначала separate design выбора external owner/scheduling/access/resources/deployment и actual sink/exposure contract, затем отдельно разрешённая реализация. C1 Acceptance не активирует C2 |
| C3 remaining Gate11 coverage | Not Ready / Not Activated. Отдельные Dispatcher/Handler/shutdown source-category designs/readiness; нельзя переиспользовать DP018 management categories или объявлять существующие два codes полным Gate11 coverage |

Successor Task IDs/branches не назначены. TASK114 Acceptance не даёт design
Approval/promotion, production adapter readiness, Gate11 closure или implementation
permission.

## 26. Exact future consumer verification matrix

Это **future requirements**, не tests/code, выполненные или добавленные TASK114.

| ID | Proposed proof / required result |
| --- | --- |
| C01 | Inert construction; zero Consumer/Reader/nil Sink => Disabled; без read/offer/callback/resources |
| C02 | Reader/Sink reassignment isolation; Consumer copies разделяют guard, mailboxes независимы |
| C03 | Precedence Disabled, затем nil-context/budget0/>64, затем Busy; invalid input ничего не потребляет |
| C04 | Concurrent/reentrant same Consumer => immediate Busy, zero read/sink; без waiting/spin |
| C05 | Budget1/64 ограничивает actual TryRead attempts при continuous producer refill; без drain-until-empty loop |
| C06 | Enabled Reader false => NoItem, не empty/drained; forced busy/invalid cases COMPOSITIONAL через existing mailbox white-box tests и consumer source-order audit; без unsafe/fake Reader/new production hooks |
| C07 | Single-owner FIFO sink invocation, detached safe copies, independent mailboxes; extra readers partition без global order |
| C08 | Before-read против after-read/pre-sink cancellation: exact attempts/valid counts/one known discard. Deterministic well-formed forwarding Context может вызвать standard CancelFunc на выбранном Err sample, сохраняя Done/Err; оставшиеся scheduling races — explicit source-order proof |
| C09 | Post-entry cancellation не отменяет sink call; nil return учитывается перед Canceled; sink error/panic имеет приоритет над поздней cancellation |
| C10 | Hostile error/panic payload (включая nil где поддерживается) не inspected/formatted/retained; normal-return flag классифицирует panic; один local failure, stop, без reinsert/retry |
| C11 | Controllably blocked sink: producer/independent Reader progress, same Consumer Busy; fixtures освобождаются до cleanup; без hard callback termination claim |
| C12 | Full mailbox/abandoned consumer сохраняет capacity64/FIFO/loss; без upstream wait/retry/new worker |
| C13 | Per-turn receipt equation/bounds; physical-removal/producer-drop counts unknown, без sequence-gap inference/cumulative metrics |
| C14 | Retained capability/sequence при rollback/retry/Stop; без reset/flush/join/lifecycle mutation; late source eligibility compositional с existing component proofs |
| C15 | Existing source HTTP/cancellation/errors.Is/legacy/Session/Manager regressions неизменны; без Runtime dependency/Observer repurposing |
| C16 | API/godoc/import/schema/scope audit; zero workers/timers/I/O/production adapter/Configuration/Host API changes |

Будущий C1 требует focused/full tests, build/vet, available race/stress и module/
tidy guards. Недоступные checks остаются explicit limitations; Linux race не
квалифицирует Windows/native TASK102. Текущая design verification — source-backed
manual decision traces, authority/parity/link/manifest/protected-path checks и
independent review, не executable consumer/delivery evidence.

## 27. Isolated C1 implementation (TASK-115)

[TASK-115](../../tasks/TASK-115-RUNTIME-DIAGNOSTICS-CONSUMER-CORE.md) реализует только
`internal/runtimediagnosticsconsumer/consumer.go` и новые external-package tests.
Ранее proposed signatures, scalar fields и восемь local Stop names разделов21–24
теперь являются isolated C1 API. Один captured concrete Reader/Sink и shared TryLock
guard; budget1..64 actual read attempts, без retry/reinsert/worker. Nil/error/
recoverable-panic outcomes и cancellation сохраняют exact documented precedence;
panic adapter отбрасывает payload и отслеживает normal return. Returned error только
сравнивается с nil; Error/Unwrap/Is/formatting не вызываются.

Task сопоставляет C01–C16 с executable consumer proofs, existing regressions и explicit
source-order/compositional portions. C06 forced mailbox busy/invalid false использует
unchanged mailbox white-box tests; C14 actual Host rollback/retry/Stop и late source
eligibility — unchanged wiring/component proofs. Runtime/Host consumer integration или
production exposure test не заявляются; C2/C3 остаются DEFERRED, Not Ready / Not Activated.
Exact verification results/limitations определяются task envelope. Матрица раздела26
остаётся requirements, впервые заданными TASK114, а не claim executable tests TASK114.

Consumer запускается только explicit external Drain caller. Нет sink implementation,
нового Host binding, lifecycle ownership, startup/Stop join, metrics, transport или
durable delivery. Producer drops unknown; NoItem != proven empty; нет retry,
durable/lossless guarantee или hard sink termination. Gate11 OPEN. Draft design и
higher-authority contracts неизменны; C1 не активирует C2/C3.

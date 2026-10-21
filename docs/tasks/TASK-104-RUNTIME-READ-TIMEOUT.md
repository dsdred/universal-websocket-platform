# TASK-104 — Runtime read-timeout enforcement

## Status

In Progress. Текущие gates и closure определяются только newest valid matching
Recovery Evidence Envelope после независимого восстановления subject manifest.
Stage, commit и publication не разрешены.

## Task Contract

- Task Mode: Implementation, documentation-first; bounded применение уже
  существующего ReadSeconds в Go Runtime.
- Why Now: Owner разрешил предложенный полный bounded engineering cycle TASK-104.
  Gate 9 MASTER_PLAN оставляет ReadSeconds inactive; опубликованный TASK-103
  реализует WriteSeconds. Read — следующий независимо проверяемый slice, не
  зависящий от native toolchain; Idle требует иных activity/reset semantics,
  TLS имеет более широкий security scope.
- Repository: universal-websocket-platform; task branch
  `feature/task-104-runtime-read-timeout`; trusted baseline main
  `aed7cc8f1693cf8bb48d71532fe543f09d2d1191`.
- Preserved sources: TASK-102 `8bfa2e22a4b6310a71f6d36c8724b1adedd727ad`,
  source-only Accepted, committed/unpublished на отдельной ветке; native build/link
  NOT RUN, toolchain NOT QUALIFIED. TASK-103 source
  `a2742ab927f7a23fd10a92db986ce36ff6a987e8`, опубликован PR #103, merge
  `aed7cc8f1693cf8bb48d71532fe543f09d2d1191`; его immutable record не меняется.
- Definition of Done: focused design EN/RU и собственный Architecture Confirmation
  PASS до кода; immutable pre-Commit propagation и read enforcement; acceptance
  matrix, focused/regression/concurrency checks, vet/formatting, race при доступном
  tooling; независимые Tester и non-author Review; PROCESS-002, Scope Audit,
  canonical committable subject, Coordinator Acceptance; STOP AT COMMIT GATE.
- Out of Scope: IdleSeconds, TLS, WriteSeconds semantics, новая Configuration,
  lifecycle/termination categories, retries, native acquisition/installation,
  downloads, host mutations, Architecture B, PoC, изменение Approved/Frozen
  contracts, TASK-102 или historical TASK-103 record, stage/commit/publication.
- Verification Plan: offline Go focused tests и regression, real WebSocket proof,
  lifecycle/concurrency stress, vet/formatting; GOTOOLCHAIN=local, GOPROXY=off,
  GOSUMDB=off. Race только существующим пригодным tooling; недоступность явно
  фиксируется как limitation. Missing dependency/download/host mutation — STOP.
  Проверить EN/RU parity, ссылки, diff и exact raw content identities.

## Objective и bounded scope

Собственная Architecture Confirmation PASS предшествовала реализации;
implemented candidate реализует следующее поведение (Acceptance ещё не заявлен):
ReadSeconds 0 отключает configured read timeout; 1..86400 задаёт новый бюджет
каждого чтения полного application message, включая ожидание и body/fragment
reading. Более ранний caller deadline/cancellation сохраняется. Handler получает
исходный Run context; время Handler не входит в read budget. Expiry завершает Run
ошибкой через существующие Owner cleanup и Manager accounting, без retry.

Intended production boundary: Session, transactional Dispatcher, Runtime composition
и executable Snapshot validation; необходимые тесты и документация. Нет нового
package/dependency, watchdog, sliding per-fragment reset, handler limit или нового
transport ownership. Архитектурные детали и exact failure semantics определяет
Architect до implementation, а не Documentation Agent.

## Sources и documentation baseline

AGENT, PROCESS-001/002 и применимые role contracts; PROJECT_CONTEXT,
spec/README/current-state/decisions, MASTER_PLAN EN/RU, ADR-0003, Frozen ARCH-002,
Active ARCH-003, Approved DP-003/004 и DP-008; Session/Dispatcher/Runtime и их
существующие тесты. Приоритет Approved/Frozen источников сохраняется.

ARCH-002 исключает effective Listener timeout behavior из freeze; DP-008 задаёт
ReadSeconds 0..86400; DP-004 допускает полностью сконфигурированную provisional
Session до Commit. Это baseline analysis, не Architecture Confirmation TASK-104
и не перенос TASK-103 PASS. Навигация ещё содержит pre-publication TASK-103
wording; bounded PROCESS-002 synchronization исправит durable outcome и current
task, не перепишет historical acceptance и не создаст Publisher authority.

## Roles и ordered workflow

Coordinator /root; Documentation Agent /root/docs104; Architect /root/architect104;
Developer /root/developer104; Tester /root/tester104. Отдельный non-author Reviewer
назначается Coordinator; автор не
выполняет independent final review собственного изменения. Publisher неприменим:
publication не разрешена.

Intake → Documentation Baseline → Architect analysis → focused EN/RU note →
exact-note Architecture Confirmation → bounded Developer implementation → Tester
verification → independent review → PROCESS-002 → Scope Audit/Size Guard → final
checks/review → Coordinator Acceptance → STOP AT COMMIT GATE. Ни один planned
stage не считается завершённым по наличию записи в этом списке.

## Existing Coverage Report — до изменения тестов

- Existing Coverage: Session Run text/binary/multiple messages, cancellation,
  unexpected read error, normal close, concurrent Run и Stop interrupt; TASK-103
  write-budget/queue/Stop tests; transactional Run failure, Owner cleanup и Manager
  convergence; Runtime Snapshot bounds и production wiring.
- Coverage Gap: configured per-Read 0/1/86400; один full-body budget, partial
  fragment timeout; earlier caller deadline/cancellation; fresh next-message
  budget; Handler original context и uncharged handler time; timeout error
  categorization, concurrent Stop/root cancellation и accounting convergence;
  separate Builder rejection versus Runtime defensive guard coverage.
- Added Proof Tests: 13 top-level TestReadTimeout tests с subtables: zero original
  context; configured deadline; earlier caller deadline/cancellation; fresh budget,
  original Handler context/uncharged Handler и child cleanup; error/normal-close
  classification до cleanup; deadline priority; success/expiry race; panic cleanup;
  dispatcher bounds/legacy compatibility; обе durations до Commit; Owner failure
  и concurrent Stop/root convergence; real fragmented-body/control-frame proof.
  Runtime proof проверяет Published Snapshot propagation 0/1/86400, detached copy,
  real Bootstrap text/binary echo, timeout Session без изменения Runtime readiness
  и subsequent Host Stop. Bounds tests различают Builder и Runtime direct guard.
- Added Regression Tests: existing write wiring proof адаптирован к combined
  constructor seam без изменения WriteSeconds contract; existing lifecycle,
  handoff, accounting и write suites сохранены. Focused stress повторён 25 раз.
- Remaining Limitations: fresh TASK-104 race probe NOT RUN — CGO_ENABLED=0,
  gcc/clang отсутствуют. Stress не доказывает race freedom. Real-time deadline
  tests имеют bounded scheduling tolerance; успешный Read может выиграть expiry
  race. TASK-103 race NOT RUN и R-103-N1 nonblocking сохраняются как historical
  limitations, не выдаются за fresh TASK-104 verification.

## Acceptance criteria

1. Runtime принимает 0/1/86400 и отвергает 86401/overflow до ресурсов; validation
   evidence различает Builder и непосредственный defensive Runtime guard.
2. Production Snapshot duration назначается Session до Commit и не меняется
   post-Commit; ownership/registration/Owner lifetime contracts сохранены.
3. Configured timeout ограничивает полное Read, включая blocked/partial body;
   новый message получает новый бюджет; 0 сохраняет исходный context behavior.
4. Caller cancellation/deadline не теряется; Handler получает исходный Run
   context без read deadline. Handler time не расходует следующий read budget.
5. Error/lifecycle, text/binary/echo, normal close, Stop/root cancellation и
   однократный cleanup/accounting подтверждены без retry/false success.
6. Existing WriteSeconds behavior проходит regression; independent review не
   содержит blocking findings; documentation соответствует verified code.

## PROCESS-002 applicability и scope guard

Task record/index, current-state и PROJECT_CONTEXT required для current task и
factual capability. Focused design EN/RU required после Architect handoff.
MASTER_PLAN EN/RU синхронизированы для Gate 9: read/write/handshake effective,
TLS/idle open; CHANGELOG описывает user-facing read enforcement. Historical write
design EN/RU синхронизирован только по publication wording, семантика не менялась.
Approved/Frozen/ADR amendments — Not applicable: architecture совместима,
normative source/status не меняется. spec/README, general design indexes и root
README — Not applicable: новая подсистема/API/навигационный раздел не вводятся,
focused note доступна из task/current-state и имеет EN/RU cross-links.
spec/decisions и остальные DP — Not applicable: нового долгосрочного решения
или design/status amendment нет. TASK-103 immutable record не меняется.

Size Guard: остановить и переоценить при >15 paths, >500 production lines,
>1 new package/contract или independently deliverable behavior; целостность slice
должна быть доказана до дальнейшей работы, а не неявно перенесена из TASK-103.
Prospective reassessment Coordinator: ожидается 17–19 paths (4 production Go,
4 test, 9–11 documentation вследствие обязательных зеркал и synchronization),
production delta <500 lines, без нового package. Один ReadSeconds behavior
требует wiring, lifecycle proof и согласованной документации; отдельная поставка
этих частей оставила бы неполный enforcement или ложный inactive/current-task
статус. TASK-103 publication wording — только metadata sync, без изменения его
record/behavior. Итоговые counts и exact Required/Questionable/Removable inventory
перепроверяются перед final review; расширение этой границы требует STOP.

## Scope Audit — candidate inventory

Все 19 paths Required для одного ReadSeconds slice; Questionable 0, Removable 0.
Новых packages/dependencies нет; production delta <500 lines. Exact counts,
raw identities и canonical manifest пересчитываются перед final review.
При raw-diff audit обе historical Write notes восстановлены к LF baseline Git
blob вместо checkout CRLF; retained changes — только publication status metadata
(3 additions/1 deletion в каждой). Format-only paths/churn не сохраняются;
остальные documentation line endings не менялись.

| Path | Required reason |
|---|---|
| internal/session/session.go | Per-read budget, deferred cleanup и distinct terminal result |
| internal/session/transactional_dispatcher.go | Bounded combined constructor и immutable pre-Commit durations |
| internal/runtime/composition.go | Published Snapshot read/write wiring |
| internal/runtime/configuration_validation.go | Actual defensive read range guard до ресурсов |
| internal/session/read_timeout_test.go | Read, transport, lifecycle, pre-Commit и concurrency proof |
| internal/runtime/read_timeout_test.go | Production Snapshot/bootstrap/echo/read-timeout integration |
| internal/runtime/configuration_validation_test.go | Separate Builder/direct Runtime guard coverage и capability inventory |
| internal/runtime/write_timeout_test.go | Existing write propagation proof сохраняется с combined factory arity |
| docs/en/design/runtime-read-timeout.md | Approved focused implementation note |
| docs/ru/design/runtime-read-timeout.md | Mandatory semantic mirror |
| docs/en/design/runtime-write-timeout.md | Удаление stale candidate/publication status TASK-103 |
| docs/ru/design/runtime-write-timeout.md | Mandatory mirror только исторического status sync |
| docs/en/roadmap/MASTER_PLAN.md | Gate 9 delivered read slice, TLS/idle remain open |
| docs/ru/roadmap/MASTER_PLAN.md | Mandatory roadmap mirror |
| spec/current-state.md | Actual read capability и current task |
| .ai/PROJECT_CONTEXT.md | Current TASK-104 и durable TASK-103 publication facts |
| docs/tasks/README.md | Current task navigation и publication drift repair |
| CHANGELOG.md | User-facing ReadSeconds enforcement |
| docs/tasks/TASK-104-RUNTIME-READ-TIMEOUT.md | Contract, handoffs, scope и recovery evidence |

## Verification handoff — candidate

Все Go commands offline: GOTOOLCHAIN=local, GOPROXY=off, GOSUMDB=off,
существующий `C:\Program Files\Go\bin\go.exe`; production/tests frozen перед
documentation synchronization. Developer focused Runtime command
`go test ./internal/runtime -run 'ReadTimeout|ReadSeconds|RuntimeWriteTimeout'
-count=1 -timeout=30s` в trusted context: exit 0, PASS, 3.629s.
Первый default-sandbox invocation был interrupted с exit 1 без output;
это не test verdict, результат не переименован в PASS.

Tester focused `go test ./internal/session -run '^TestReadTimeout' -count=1`
exit 0 PASS, 3.075s; `-count=25` exit 0 PASS, 9.509s: 13 top-level tests × 25 =
325, плюс subtables. Frozen Session test raw blob
`4302fab79b7aa7c3d5de5035529daa41e9b0e96c`. Первый partial/Ping test выявил
недостаток test harness до добавления client read loop; исходный fail не является
passing evidence, засчитан только последующий исправленный proof.

Architect actual four-production-file compatibility PASS, без Approved/Frozen
amendments; точные финальные subject identities относятся к downstream manifest.
Полный regression Coordinator: `go test -count=1 -timeout=120s ./...`, exit 0,
все Go packages PASS (Runtime 15.208s, Session 4.615s, runtimecontainment 25.949s).
`go vet ./...` exit 0, marker GO_VET_PASS_EXIT0. `gofmt -l` для всех восьми
changed Go paths не вывел имён, exit 0. Go version 1.26.5, без downloads.

`go mod tidy -diff` exit 1: только 24 existing go.sum lines CRLF→LF с теми же
module/hash strings; go.mod/go.sum не менялись. Это не tidy PASS и не dependency
change. Raw-byte-preserving whitespace check
`git -c core.autocrlf=false -c core.whitespace=cr-at-eol diff --check` exit 0.
Index пуст; HEAD/main/cached origin/main остались baseline aed7; TASK-102 остался
8bfa. Общий verification: PASS WITH LIMITATION, race NOT RUN. Native artifacts,
host tooling и immutable task records не изменялись.

Independent final review и Coordinator Acceptance ещё не завершены; latest
matching terminal envelope должен предоставить canonical rows/manifest,
фактические verdicts/findings и post-decision integrity, не предполагая их PASS.

## Branch, recovery и stop boundaries

Coordinator создал task branch из clean main до первой content mutation; этот
record является первой content mutation TASK-104. Source baseline и отдельный
TASK-102 сохраняются. No stage/commit/publication, native workstream или transfer
Publisher ownership не разрешены task cycle.

Для будущего отдельно разрешённого Commit Gate reviewed raw bytes сохраняются
без filter/CRLF normalization. Сейчас core.autocrlf=true; staging должен применять
per-invocation `git -c core.autocrlf=false add -- <exact paths>` при необходимости
и сравнивать staged blob bytes с approved subject. Это требование подготовки,
не stage authority; конфигурация Git и index в этом cycle не меняются.

STOP: Architecture Confirmation FAIL; required Approved/Frozen amendment;
unattributed diff/baseline mismatch; scope expansion; blocking review finding;
missing dependency, download или host mutation; готовый candidate в Commit Gate.
После interruption сначала inspect/reconstruct actual bytes/checkpoints, без
replay completed mutations. Следующая task/recommendation — Not Activated.

## Recovery Evidence Envelope

### E-104-001 — Intake anchor

Owner разрешил bounded TASK-104 engineering cycle; Coordinator передал
Documentation Agent создание только этого record после branch preparation.
Baseline: `aed7cc8f1693cf8bb48d71532fe543f09d2d1191`; branch:
`feature/task-104-runtime-read-timeout`. Record не attests собственные final bytes.
First incomplete checkpoint: TASK-104 Architecture Analysis/Confirmation и
pre-implementation focused design. Implementation, tests, independent review,
Acceptance, commit/publication ещё не выполнены; approval не подразумевается.

### E-104-002 — Architecture analysis handoff и focused design

Independent Architect /root/architect104 передал совместимое решение: Runtime
ReadSeconds guard 0..86400; immutable pre-Commit read/write durations через
combined constructor; lexical per-Read child budget с classification до deferred
cleanup и различением raw failure/normal close; исходный Handler context;
неизменные Owner/Manager и causal termination rules. Focused notes созданы в
docs/en/design/runtime-read-timeout.md и docs/ru/design/runtime-read-timeout.md
с Implementation status Planned. Exact-note Architecture Confirmation pending;
до её PASS implementation не разрешена. TASK-103 PASS не переносится.

### E-104-003 — Exact-note Architecture Confirmation

Independent Architect /root/architect104 выполнил exact-byte readback и EN/RU
semantic parity: Architecture Confirmation PASS, blocking findings 0;
Approved/Frozen amendment не требуется. Approved note identities до status-only
synchronization:
- EN raw blob `5301400c2f7dd0142788ca810ac0270f4ca8e647`, SHA-256
  `CA9423E9300C0C65217FB6BEF85EA5A175389A58D18EA46DBBA56CE2A7C3CD81`.
- RU raw blob `5b6b47692d512ad467e7981ccbce47e3233867e5`, SHA-256
  `99F83AF9E54D5FE0D1AA90A4C968BFF27E65AD8CE574E412D3637F7EBCD66BCF`.

По явному Architect handoff Design status обеих notes синхронизирован на
Architecture Confirmation PASS; Implementation остаётся Planned. Это status-only
изменение не меняет approved semantics; final subject будет пересчитан позднее.
Implementation разрешена только в границе этих notes. First incomplete checkpoint:
bounded Developer implementation; Verification/Review/Acceptance ещё не пройдены.

Tester104 baseline handoff: GOTOOLCHAIN=local, GOPROXY=off, GOSUMDB=off;
`C:\Program Files\Go\bin\go.exe test ./internal/session ./internal/runtime
./internal/listener ./internal/runtimeconfig ./internal/executionowner
./internal/sessionmanager` завершился exit 0, PASS (3.467s/3.536s/2.787s/
0.574s/0.617s/0.686s соответственно), до implementation.
Fresh capability probe `go test -race -run '^$' ./internal/session` завершился
exit 1: `go: -race requires cgo; enable cgo by setting CGO_ENABLED=1`.
CGO_ENABLED=0, host amd64, gcc/clang отсутствуют; race NOT RUN, baseline verification
PASS WITH LIMITATION. Это отсутствие tooling, не race failure и не доказательство
race freedom. Новые инструменты/host changes не предпринимались.

### E-104-004 — Candidate Identity

Repository: universal-websocket-platform; branch
`feature/task-104-runtime-read-timeout`; trusted baseline и current HEAD
`aed7cc8f1693cf8bb48d71532fe543f09d2d1191`; object format sha1.
Documentation Agent независимо восстановил 19 present paths, mode 100644.
Canonical subject manifest: `0dd86392366f9e1d7fe75df63f3864ec69b6d37b`.
Task projected blob: `1e696f9236e4cd511848feca4048533bc1c5d85c`.

Воспроизведение: exact set — changed tracked paths относительно baseline плюс
untracked TASK-104 paths; ascending unsigned UTF-8 path-byte order (все paths
ASCII). Для present full rows читать raw File.ReadAllBytes и передавать bytes в
`git hash-object --stdin`, без -w, filters или normalization. Для task-record-v1
только по PROCESS-001: BOF через actual Status line terminator + UTF-8
STATUS-EVIDENCE-EXCLUDED и один NUL + raw Task Contract bytes до envelope heading.
Проверить unique/ordered headings и terminal envelope. Manifest stream — для
каждой ordered row `path\0projection\0state\0mode\0oid\0`, без newline между rows;
`git hash-object --stdin` возвращает указанный manifest. Strongly typed
List[string].Sort(StringComparer.Ordinal) корректен для этих ASCII paths;
неоднозначная Array.Sort overload была отвергнута до записи как tooling error.
Эти identities не attest собственные raw final envelope bytes.

| Path | Projection | State | Mode | OID |
|---|---|---|---|---|
| .ai/PROJECT_CONTEXT.md | full | present | 100644 | 11ca566f53cf74acb8351498a45b1466352d1544 |
| CHANGELOG.md | full | present | 100644 | bdd879cf9600b08fba3d6704c8fe59f33905a46d |
| docs/en/design/runtime-read-timeout.md | full | present | 100644 | fe76132947af7662712782cedc965e7ec7aec360 |
| docs/en/design/runtime-write-timeout.md | full | present | 100644 | 9b03f3ad3c3affcb6db29ab81069634617548873 |
| docs/en/roadmap/MASTER_PLAN.md | full | present | 100644 | 4266b5e724f65186a322755bd889c978672abea0 |
| docs/ru/design/runtime-read-timeout.md | full | present | 100644 | 578a0837fcf65749bac33f88dbd3139d4a31dfd0 |
| docs/ru/design/runtime-write-timeout.md | full | present | 100644 | 80c2972b793a7f04ffa2dac5e1fe87aa4f5a9594 |
| docs/ru/roadmap/MASTER_PLAN.md | full | present | 100644 | b7e18616c06df58044376c11dd75be2169c93291 |
| docs/tasks/README.md | full | present | 100644 | c07537eb9279db3af97a1ab62a93a45130a732dc |
| docs/tasks/TASK-104-RUNTIME-READ-TIMEOUT.md | task-record-v1 | present | 100644 | 1e696f9236e4cd511848feca4048533bc1c5d85c |
| internal/runtime/composition.go | full | present | 100644 | f414940d99ac2a13559b6d42a0a1c605477ed541 |
| internal/runtime/configuration_validation.go | full | present | 100644 | 643bd6f45f123846142d690bdc3df95f221f5c86 |
| internal/runtime/configuration_validation_test.go | full | present | 100644 | e4205a2f352048b9e19f83628159f465a75f72e4 |
| internal/runtime/read_timeout_test.go | full | present | 100644 | abc274eddd919fce7035a6d8e3275f5e8fdb73c4 |
| internal/runtime/write_timeout_test.go | full | present | 100644 | d7996bdff132fc7ef5abdcaa4740d0fe924c1dcf |
| internal/session/read_timeout_test.go | full | present | 100644 | 4302fab79b7aa7c3d5de5035529daa41e9b0e96c |
| internal/session/session.go | full | present | 100644 | 8f4c8c9045a6ec6a4eafb05db45934ae0c484cb5 |
| internal/session/transactional_dispatcher.go | full | present | 100644 | 37e6ae9ed11ee9d7386274a4249089c16e0322bf |
| spec/current-state.md | full | present | 100644 | d1803366364b1cce29283d0ea2892afde5b6212d |

Documentation synchronization: EN/RU parity и local links checked; root сообщил
218 valid relative links в 11 changed Markdown files. Raw CR-aware diff check
PASS, index empty, eight changed Go files gofmt clean; refs остаются baseline,
TASK-102 source неизменён. Fresh Architect/Tester handoffs и limitations описаны
выше; independent final Review и Coordinator Acceptance pending. Эта запись не
является Acceptance, commit или publication authority.

### E-104-005 — Exact-subject Tester и Independent Review

Subject: manifest `0dd86392366f9e1d7fe75df63f3864ec69b6d37b`, task projection
`1e696f9236e4cd511848feca4048533bc1c5d85c`, exact 19 rows E-104-004,
baseline/current HEAD `aed7cc8f1693cf8bb48d71532fe543f09d2d1191`, branch
`feature/task-104-runtime-read-timeout`.

Tester /root/tester104 final exact-subject handoff: PASS WITH LIMITATION.
Session focused 13 × 25 = 325 top-level executions PASS (9.509s), Runtime focused
PASS (3.629s); полный `go test -count=1 -timeout=120s ./...` exit 0 и
`go vet ./...` exit 0. Детали исходных команд, test harness correction,
coverage и layer distinctions сохранены в Verification handoff выше.
Race NOT RUN: CGO/tooling unavailable; stress не является race-freedom proof.
`go mod tidy -diff` exit 1, только existing go.sum CRLF/LF при одинаковых
module/hash strings; это не tidy PASS, module files не изменены.

Independent non-author Reviewer /root/reviewer104: APPROVED, blocking findings 0,
nonblocking findings 0. Reviewer независимо подтвердил exact manifest/projection,
19 Required paths, architecture conformance, PROCESS-002 EN/RU parity и Scope
Audit/Size Guard justification. Source/evidence bytes frozen; Reviewer не менял
файлы и не выполнял product execution. main/cached origin/main остаются baseline,
TASK-102 сохранён отдельно, index пуст.

Первый незавершённый checkpoint: Coordinator Acceptance и post-decision integrity.
Эта запись не self-attests собственные final raw bytes и не разрешает stage,
commit или publication.

### E-104-006 — Coordinator Acceptance R-104-001

Coordinator /root: ACCEPT TASK-104, decision R-104-001, exact manifest
`0dd86392366f9e1d7fe75df63f3864ec69b6d37b`, task projection
`1e696f9236e4cd511848feca4048533bc1c5d85c`; branch
`feature/task-104-runtime-read-timeout`, baseline/current HEAD
`aed7cc8f1693cf8bb48d71532fe543f09d2d1191`. Acceptance criteria 1–6 выполнены
по recorded proofs: Architecture Confirmation PASS, Tester PASS WITH LIMITATION,
independent Reviewer APPROVED (0 blocking/0 nonblocking), PROCESS-002 parity PASS,
Scope Audit 19 Required/0 Questionable/0 Removable. Size Guard: один behavior,
4 production + 4 test + 11 documentation paths; production delta +53/−9.

Приняты limitations: race NOT RUN (недоступный CGO/tooling) и existing
`go mod tidy -diff` CRLF-only exit 1, не tidy PASS. Source/tooling не изменялись
для обхода ограничений. Latest task disposition: Coordinator Accepted, READY для
отдельно разрешаемого Commit Gate, uncommitted/unpublished. Projected In Progress
разрешается этим newest valid matching envelope, а не дублированием verdict в body.

Первый незавершённый checkpoint: post-decision integrity, затем отдельный Owner
Commit Gate. Ни stage, ни commit/publication не разрешены. Следующая рекомендация:
независимый IdleSeconds design/intake только после terminal task gates; Not
Activated, новый Task ID не зарезервирован. Эта запись не attests собственные
final raw bytes; post-decision readback должен выполняться независимо.

### E-104-007 — Post-decision integrity

Independent Reviewer /root/reviewer104 после E-104-006: post-decision integrity
PASS. Canonical 19-path subject manifest
`0dd86392366f9e1d7fe75df63f3864ec69b6d37b` и task projection
`1e696f9236e4cd511848feca4048533bc1c5d85c` неизменны. R-104-001 является
правильным latest Coordinator Accepted disposition; candidate остаётся
uncommitted/unpublished. Independent review APPROVED, 0 blocking/0 nonblocking;
race NOT RUN и existing tidy CRLF-only exit 1 (не tidy PASS) сохранены.

Coordinator fresh readback независимо подтвердил тот же subject и exact blobs:
branch `feature/task-104-runtime-read-timeout`; HEAD/main/cached origin/main
`aed7cc8f1693cf8bb48d71532fe543f09d2d1191`; TASK-102
`8bfa2e22a4b6310a71f6d36c8724b1adedd727ad` сохранён отдельно; index пуст.
Перед этой append raw task SHA-256 был
`B1355C6E24703103BAFDDF97F06025706A82D440AAA6D1E43EF617F7E9D36F1D`;
это observation предыдущих bytes, не self-attestation final envelope.

Первый незавершённый gate: Commit Gate, требующий отдельного Owner authorization.
STOP AT COMMIT GATE. Stage, commit и publication не выполнялись и не разрешены;
новая реализация, Architectures B/PoC и следующая task не активированы.

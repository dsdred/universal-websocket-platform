# TASK-123 — Runtime Diagnostics C2-Exposure Design/Readiness

## Status

`In Progress`. Owner разрешил Design-only после READY intake, Architecture
Confirmation PASS и Independent Intake Review Approved (0/0). Production C2/C3
и C2-Exposure Not Activated; Gate 11 OPEN. Commit/Publication Gates не разрешены.

## Task Contract

### Task Mode

`Design-only`: documentation checkpoint, затем exact C2-Exposure consume/API/resource
contract в Draft DP-013 и DP-027 EN/RU и независимая приёмка одного frozen subject.
Ранее завершённый intake является prerequisite, не acceptance текущего design.

### Why Now

- TASK-122 опубликовала isolated C2-Binding, PR #123 MERGED; P10/P10R-122-001 PASS.
- Требуется оценить следующий C2-Exposure design slice по действующим contracts.
- Owner отдельно разрешил TASK-123 Design-only в принятом twelve-path scope.

### Definition of Done

1. Проверены full baseline и terminal TASK-122; семь project-state paths синхронизированы
   с EN/RU parity до первого normative edit.
2. Architect определил dedicated consume authority до lookup, anti-enumeration,
   exact Instance/Attempt, bounded Poll/concurrency/rate/resources, concrete transport/API,
   redaction/retention/data classification, Stop/revocation/in-flight semantics и trust boundaries.
3. DP-013/027 EN/RU зеркально фиксируют design и проверяемые acceptance proofs,
   dependency/activation gates и smallest separate implementation scope.
4. Architecture Conformance, Independent Verification, Independent Final Review,
   PROCESS-002, Scope Audit, final-byte integrity и EOL staging projection PASS
   для одного exact frozen twelve-path subject.
5. Coordinator Acceptance design записана без production activation или Commit/Publication Gate.

### Out of Scope

- Production/test/module/workflow changes, реализация новых API/endpoints/workers/registry.
  Planned API/transport contract в разрешённых Draft DP входит в Design-only scope.
- Изменения Frozen/Approved contracts без отдельного Architecture Gate.
- TASK-102, Issue #121 implementation, production C2/C3, Gate 11 closure.
- C3 schema/source design, durable/lossless guarantees, flush/final emptiness/join.
- Real-index stage/intent-to-add, commit, push, PR, merge, fetch/pull, reset/rebase
  и удаление веток. Исключение для проверки: explicitly authorized EOL projection
  использует отдельный scratch index, не меняющий real index/refs/worktree/config.
- Изменение historical accepted TASK-122 record или ранее опубликованного subject.

### Verification Plan

- Git history/refs/status, live read-only GitHub PR/origin facts и durable native P10/P10R.
- Source-backed mapping каждого requested aspect к ARCH/DP и current source seams.
- Architect design/conformance; Independent Verification и read-only Independent
  Final Review полного frozen twelve-path design subject, не только intake record.
- Проверка local links, exact changed/untracked set, whitespace и final-byte identity.
- Runtime tests N/A для Design-only: executable behavior не меняется; future runtime
  acceptance proofs описываются, но не объявляются выполненными.
- Raw-byte/EOL preservation, temporary-index staging projection при пустом real index,
  exact canonical task-record-v1 manifest, independent frozen-subject recomputation.

## Scope

Exact accepted scope — 12 documentation paths:

- `.ai/PROJECT_CONTEXT.md`;
- `spec/current-state.md`;
- `docs/tasks/README.md`;
- `docs/en/roadmap/MASTER_PLAN.md`;
- `docs/ru/roadmap/MASTER_PLAN.md`;
- `docs/en/design/README.md`;
- `docs/ru/design/README.md`;
- `docs/en/design/DP-013-runtime-management-routing.md`;
- `docs/ru/design/DP-013-runtime-management-routing.md`;
- `docs/en/design/DP-027-runtime-diagnostics-v1.md`;
- `docs/ru/design/DP-027-runtime-diagnostics-v1.md`;
- `docs/tasks/TASK-123-RUNTIME-DIAGNOSTICS-C2-EXPOSURE-READINESS.md`.

Иных путей не менять. DP-010/009/011/012, ARCH, DP-018 и historical task records
остаются read-only. При необходимости нового path — STOP и отдельная авторизация scope.

## Sources of Truth

- [AGENT](../engineering/AGENT.md), [PROCESS-001](../engineering/PROCESS-001-AI-DEVELOPMENT-WORKFLOW.md),
  [PROCESS-002](../engineering/PROCESS-002-DOCUMENTATION-SYNCHRONIZATION.md).
- [ARCH-002](../en/architecture/ARCH-002-runtime-foundation-freeze.md),
  [ARCH-004](../en/architecture/ARCH-004-runtime-deployment-and-identity-model.md).
- Draft DP-010/011/012/013/027; Approved DP-018 только в собственной management boundary.
- [TASK-121](TASK-121-RUNTIME-DIAGNOSTICS-C2-BINDING-DESIGN.md),
  [TASK-122](TASK-122-RUNTIME-DIAGNOSTICS-C2-BINDING.md); actual source и project state.

## Roles

- Coordinator: /root, intake contract, evidence, scope и final verdict.
- Architect: `/root/exposure_architect`, design decisions/handoff и conformance;
  файловые edits выполняет Documentation Agent после checkpoint.
- Documentation Agent: /root отдельно назначен компилировать факты intake в эту запись;
  не принимает архитектурных решений.
- Verifier/Reviewer: отдельные independent роли на новом frozen design subject;
  intake Approved не переносится на текущий design.
- Developer/Tester runtime cycle/Publisher: N/A; код и executable behavior не меняются,
  Commit/Publication Gates не разрешены.

## Branch and Recovery Anchor

- Repository: `dsdred/universal-websocket-platform`.
- Trusted baseline: clean `main == origin/main == live origin/main` at
  `6fb5b02bbff3186db86e26cb946be4b24095edfd`.
- Baseline tree: `4a39c852b2467ac4e2071387eeaab821fc213c63`.
- Intake branch: `docs/task-123-c2-exposure-readiness`, создана без изменения main.
- На старте design attributed dirty state — ровно один intake record с final raw SHA256
  `85ec2784546052b12d524d9fbcf4df3bddac43b50bdac53160911d00e463bff0`;
  остальные 11 scope paths совпадают с baseline, real index пустой.
- Branch preparation deviation: первая sandbox-команда switch завершилась отказом
  записи ref; её exit не был проверен до создания untracked record. До дальнейшей
  работы отдельно проверен exact baseline/единственный untracked path и создана
  docs-ветка через supported trusted runner. Main ref и tracked content не менялись.
- До mutation: read-only inventory → TASK-122 terminal reconstruction → task contract.
- После interruption не считать started checks completed; сверить refs/bytes/subject
  и продолжить первый недоказанный intake checkpoint.

## TASK-122 Terminal Prerequisite

Fresh read-only Git/GitHub inspection 2026-11-11 подтверждает:

- Source: `7881d66cfe7e4bcd1c5eba7e8e8a5ae50385ae54`.
- Base: `0bf83bd394071f4ca9a6c102201a2f117acec819`.
- PR [#123](https://github.com/dsdred/universal-websocket-platform/pull/123) MERGED;
  merge `6fb5b02bbff3186db86e26cb946be4b24095edfd`.
- Source/merge tree: `4a39c852b2467ac4e2071387eeaab821fc213c63`.
- Subject M122-002; manifest `568f836c96ef5adad922444f732f97500d10d8c2`, 15 paths.
- Publication P0–P10 PASS; native root P10
  `msg_0674951bff14c618016af437a653d487d2b0550772d5e014b3`, thread
  `01a1c9d6-854b-71c3-a014-96a4fdba208c`, ordinal 504,
  payload SHA256 `d5a060fc3dbdd952f3ca12dddcce9a0ca1f1c2f5e4f492d269e6f40f19921b97`.
- Independent P10R-122-001 PASS, findings 0/0; original reader thread
  `01a1c9ed-8306-79a2-91d4-354d172dd8cf`, item
  `msg_0168cac747d81031016af43877980087d2acd3de17f349da87`, ordinal 311,
  payload SHA256 `ddc14ce15455f98c6476fde9990435406423694be293873d24292784b091d2bd`.
- PR CI `42996024425` и post-merge CI `42996472196` SUCCESS, все семь checks,
  включая hosted Linux race, tidy/module guards.
- Publication terminal: `Consumed(P10) / NoneTerminal / Unissued`, transfer ID none.
- Перед intake main synchronized/clean; task refs отсутствуют. Это terminal isolated
  implementation prerequisite, не production C2 admission или Gate 11 proof.
- Issue #121 fresh API: OPEN; в project scope остаётся Not Activated.

## Stop Conditions

- Baseline/terminal evidence conflict, необъяснённые изменения или другая active work.
- Необходимость изменить Frozen/Approved source без отдельного Architecture Gate.
- Расширение exact twelve-path scope без отдельного Owner decision.
- Отсутствующие decisions трактуются как design inputs, а не invented implementation readiness.

## Selection and Existing Coverage

Выбран один bounded slice: readiness к отдельно разрешённому C2-Exposure design.
DP-027 §32 ставит exposure после C2-Pull и C2-Binding; TASK-120/121/122 эти isolated
prerequisites опубликовали. Scope можно оценить без изменения Frozen/Approved sources.
Production implementation отклонена: P03–P05/P07/P08 не закрыты. C3 — отдельный
неактивированный track; TASK-102 и Issue #121 не prerequisites этого design intake.

Existing coverage: TASK-120 O01–O16 и TASK-122 B01–B08 подтверждают fixed64 Poll,
private per-attempt lifetime, isolation и revocation races. Hosted PR/post-merge
TASK-122 CI подтверждает regression/race/module checks на exact accepted tree.
Это не tests production policy, Directory consume path, API или operator exposure.
Intake не создаёт tests, поскольку executable behavior остаётся тем же.

## Source Inventory and Readiness Matrix

Опорные источники:
[DP-027 §§29–36](../en/design/DP-027-runtime-diagnostics-v1.md),
[DP-010 §33](../en/design/DP-010-runtime-lifecycle-owner-contract.md),
[DP-013 §31](../en/design/DP-013-runtime-management-routing.md),
[DP-018](../en/design/DP-018-runtime-operational-error-reporting-redaction.md).
EN/RU counterparts проверяются Architect; normative edits в этом intake отсутствуют.

| Aspect | Доказанный prerequisite / обязательная граница | Что должен решить последующий design |
| --- | --- | --- |
| Production owner/admission | DP-027 §36.1 выбирает per-Instance Lifecycle Owner в будущем single Control Service process. `cmd/control-service/main.go` сейчас не создаёт Owner/Directory/Runtime graph. | Явная application composition, deployment и единственный sanctioned consume path; current binary не получает capability автоматически. |
| Consume authority | DP-027 §36.3 и DP-013 §31 требуют dedicated authorizer до lookup/Owner/Poll. Existing `ActionObserve` и DP-018 authority не подходят. | Exact request authority с Target и expected Launch Attempt; policy source, current authorization/revocation и fail-closed outcomes. |
| Exact Runtime Instance lookup | ARCH-004 identity graph: Workspace → Configuration → Runtime Instance → Launch Attempt → Host. Directory имеет immutable exact binding; Owner seam проверяет expected active attempt. | После consume authorization сверить весь Target и exact lifetime; не искать по ConfigurationVersion, Sequence, PID/Host pointer и не подменять successor attempt. |
| Anti-enumeration | Current Directory.admit (`directory.go:232–237`) сначала lookup, потом authorizer: это existing management contract, не готовый C2 adapter. | Distinct consume admission до lookup; denied/absent/mismatched/retired responses не раскрывают existence и не читают mailbox. Exact external status/body/timing/logging disclosure contract ещё не выбран. |
| Concurrency/rate/resources | View.Poll имеет общий TryLock для copies, fixed64 result и один Drain при admission. Это не per-caller/global rate limit и не pending queue limit. | Bounded caller/target/global concurrency, pending policy, rate/budget limits, storage bounds и rejection/cancellation semantics; конкретные значения не выдумываются intake. |
| Bounded Poll | `diagnostics.go:51–79`: exact active attempt/lifetime claim под Owner mutex, затем unlock и один Poll. Budget 1..64; View detached result; Busy без read, без retry/reinsert/worker. | External request validation и safe mapping receipt; NoItem не final-empty/flush/ACK, Sequence не identity/drop count. Budget bound не hard wall-clock guarantee. |
| Transport/API | Current Control Service регистрирует Configuration/Workspace/Studio routes, diagnostics route отсутствует; DP-013 не задаёт DTO/transport. | Выбрать конкретный transport, request/response/version/status contract, cancellation, serializer и authorized target context вне Event. HTTP/CLI/streaming не выбираются этим intake. |
| Redaction/retention | DP-027: только safe Event SchemaVersion/Code/Sequence и scalar receipt; нет raw error/request/Principal/credential/provider metadata/identity labels. Returned batch принадлежит caller. | Allowlist presentation/serializer, disclosure of separately authorized target, retention/expiry/resource policy и proof отсутствия утечек; DP-018 categories/replay не переносить на diagnostics. |
| Shutdown/revocation | TASK-122: failed launch/successful Stop retire, failed Stop retains pair; caller cancellation не retire. Revocation до locked claim запрещает Poll; уже admitted Poll может завершиться. | Adapter прекращает scheduling независимо, не joins Host/Manager; blocked/failed transport не меняет lifecycle. Явные late/unread loss и discard semantics без final read/flush/delivery guarantee. |

## Minimal Project-State Reconciliation Scope

Следующие paragraphs сохраняют historical intake scope rationale. Current Owner
authorization разрешила перечисленные семь existing paths и четыре normative paths;
актуальный contract находится выше. Execution checkpoint фиксируется в envelope.

Найдены **семь обязательных existing paths** для следующего отдельно разрешённого
reconciliation/design work; вместе с текущим task record это восемь documentation
paths. В этом intake изменяется только task record; таблица является scope proposal.

| Path | Минимальное изменение |
| --- | --- |
| `.ai/PROJECT_CONTEXT.md` | TASK-122 terminal source/PR/merge/CI/P10R facts вместо stale active/pre-publication текста; current intake и неактивированный следующий design. |
| `spec/current-state.md` | Те же durable facts, current isolated-versus-production boundary. |
| `docs/tasks/README.md` | Terminal TASK-122, navigation к TASK-123 и truthful intake/design status. |
| `docs/en/roadmap/MASTER_PLAN.md` | Удалить stale TASK-122 acceptance-in-progress, записать terminal prerequisite; C2-Exposure остаётся design work, production Not Activated. |
| `docs/ru/roadmap/MASTER_PLAN.md` | Зеркально EN roadmap, без изменения Gate 11 criteria. |
| `docs/en/design/README.md` | DP-027 row: заменить stale TASK-122 acceptance-in-progress на published isolated binding. |
| `docs/ru/design/README.md` | Зеркально EN design index. |

Исторический TASK-122 record не переписывается. DP-010/DP-027 уже правдиво обозначают
isolated implementation и отсутствие production composition; их headers/status
не требуют изменения только из-за merge. P08 требует production composition/race
и source-to-operator proofs: hosted race isolated TASK-122 это требование не закрывает.
Не следует механически переписывать historical TASK-116/121 prose или повышать DP status.

PROCESS-002 applicability:

- Task record — Required, текущий единственный content path.
- Семь current-state/navigation paths выше — Required для будущей reconciliation;
  deferred до отдельного разрешения scope, поскольку Owner запросил определить scope.
- `spec/decisions.md` — N/A: новый architectural/product decision ещё не выбран.
- Root README EN/RU и CHANGELOG — N/A: нет нового user-facing capability/release.
- DP-010/027, architecture, API docs — read-only sources в этом intake; design edits
  допускаются только в будущем отдельно разрешённом design scope.
- Production/tests/modules/CI — N/A: поведение не меняется.

## Proposed Next Design Contract

Historical intake proposal ниже принят current Owner authorization как Design-only
work этой TASK-123; это по-прежнему не implementation/activation permission.

На момент intake это было proposal, не implementation permission.
Следующий separately authorized Design-only slice должен закрыть P03–P05/P07/P08
на уровне contract и дать проверяемые future proof obligations; P01/P02/P06
сохраняют accepted binding design и isolated evidence, не превращаясь в production proof.
Предварительные normative sources: DP-027 и DP-013 EN/RU; менять DP-010/009/012
только если Architect докажет необходимость изменения уже существующего seam.
Семь reconciliation paths учитываются отдельно. При подтверждении только DP-027/013
и текущего task record суммарный proposed documentation scope — 12 paths;
точный scope/Size Guard фиксируется перед full design, не расширяется автоматически.

Обязательные design deliverables:

1. Exact consume authority/target/attempt/revocation model и authorization-before-lookup trace.
2. Anti-enumerating denial/not-found/mismatch/retired contract и zero-read proof cases.
3. Explicit bounded rate/concurrency/pending/resources/retention policy.
4. Concrete transport/API/serialization/redaction/cancellation contract.
5. Scheduling/shutdown/retirement traces и no-lifecycle-join proofs, включая failed Stop.
6. P01–P08 traceability и smallest separately authorized implementation slices.

Production authority policy и transport должны получить явное design решение.
Никакие их отсутствующие параметры не интерпретируются как допустимые defaults.

## Architecture Confirmation and Intake Verdict

Это historical intake verdict. Текущий full design проходит новый independent
Verification/Review/Acceptance; прежние manifest и Approval не переносятся.

Architect `/root/exposure_architect` независимо проверил source-backed matrix,
ARCH-002/004 и действующие DP boundaries. Architecture Confirmation:
**PASS для отдельно разрешённого C2-Exposure Design-only slice**, blocking
architecture conflicts **0**. Нового architecture/product decision этот verdict не создаёт.

Совместимость сохраняется при следующих ограничениях: Host остаётся sole operational
graph/lifecycle owner; lifecycle Admission Gate не становится auth/rate gate;
Runtime Instance/Launch Attempt остаются exact identity, не PID/Sequence/Host pointer;
Reader/View не экспортируются; пятого Bootstrap binding нет; base dependencies не
мутируются; single-process scope не превращается в multi-process capability.
ARCH-004 §19 production prerequisites сохраняются отдельно и не снимаются C2 readiness.

Уточнение failed Stop: тот же Host/pair удерживается; repeated same-attempt Stop
возвращает stored failure **без cleanup retry**. Intake не обещает повторную очистку
или усиленные delivery/retirement guarantees.

| Gate | Verdict |
| --- | --- |
| TASK-122 terminal prerequisite | PASS, exact Git/GitHub/native P10/P10R evidence |
| C2-Exposure Design/Readiness | READY к отдельному Owner-authorized Design-only scope |
| Architecture Confirmation | PASS, blocking architecture conflicts 0 |
| Full exposure contract | Ещё не разработан/не принят; policy, API и bounds остаются decisions следующего design |
| Implementation / production C2 | NOT READY / Not Activated |
| Production C3 / C2-Exposure activation | Not Activated |
| TASK-102 / Issue #121 | Не затронуты / Not Activated |
| Gate 11 | OPEN |

Readiness не является Coordinator Acceptance ещё не выполненного full design,
permission на implementation или production activation. После независимого review
этого intake — STOP; следующий Design-only этап не запускается автоматически.

## Selected Full Design and Readiness Boundary

Architect выбрал future single Control Service process с существующими immutable
Directory/per-Instance Owner/in-process Host. Planned transport — HTTP/1.1 over
TLS 1.3 mutual TLS, `POST /runtime-diagnostics/v1/poll`; никаких current routes,
handlers, TLS configuration, principal mapping или authority backend не добавлено.
Нормативный источник consume/admission/API/core-slice — DP-013 §32; ресурсных,
wire/redaction/retention/shutdown/proof решений — DP-027 §37 (EN/RU).
Copies одного adapter разделяют admission/quota state; отдельные instances — нет.
Constructor проверяет duplicates/reused Owner/inconsistent mapping только внутри
переданной immutable configuration. Process-global bounds требуют Composition Audit:
ровно один sanctioned adapter, без direct Owner/alternate Directory/route bypass.
Global registry/lock или обнаружение неизвестных других adapters не подразумеваются.

Dedicated action `runtime.diagnostics.consume` проверяет current exact-grant
membership principal + Workspace/Configuration/RuntimeInstance + expected attempt
до lookup. Grant success — security admission cut ровно одного invocation;
политический revoke до него deny, после — для subsequent requests. Отдельный Owner
lifetime capture cut не меняется: retirement до capture исключает Poll, после
допущенный Poll может завершиться. Никакого external-policy callback под Owner mutex.

Выбраны конечные conservative initial budgets: 32 configured principals,
64 immutable Targets, максимум 2048 exact grants; process/principal/Target active
8/1/1, waiting queue 0; token buckets 8/s burst8, 2/s burst2, 1/s burst1;
budget1..64, body1024B, response16KiB; handler2s, body-read1s, write1s.
Ingress32 connections/8 handshakes, handshake2s/header1s/header8KiB/idle5s.
Это bounds work/memory, не benchmark/throughput или hard termination guarantee.
Refill lazy monotonic, без worker; нет request-created slots/history и reset quotas
через смену attempt/grant. Saturation discards excess whole credit; backwards/overflow
refuse без изменения counter/time.

Target-specific denial/missing/retired/mismatch/Busy/rate и invalid detached result
дают одну fixed404 unavailable response без identity. Только source-independent
prelookup process/principal admission даёт429; policy и quotas не дают enumeration API.
Нет обещания exact timing equality. Post-admission failure может потерять consumed
данные: error response не обещает zero consumption после Poll.
Success — existing closed Event/schema/uint64 Sequence и bounded receipt, без raw
Reader/View/Producer/error/Principal. Response cache/replay/retained history нет;
server copies живут один request, client erasure не обещается. Host/Manager не joins
authorizer/Poll/renderer/network; failed Stop сохраняет stored failure без cleanup retry.

Design acceptance proofs E01–E16 покрывают authority, zero-read denial, exact lookup,
две revocation cuts, budget/receipt, finite concurrency/refill/state, mTLS/DTO/output,
retention, shutdown и production composition. Они Specified by Design; executable
production proof/P01–P08 PASS этот deliverable не создаёт.
Минимальный следующий candidate после принятого design и отдельного intake —
isolated admission/resource core, только `internal/runtimemanagement/diagnostics.go`
и `diagnostics_test.go`, без существующих ctor/API changes, HTTP или production backend.
Новый opaque adapter constructor возможен только при отдельной implementation
authorization. Transport/fake-core и production integration — отдельные gates;
real mTLS/provisioning/revocation, bounded ingress, ARCH-004 §19 и integrated proofs
отсутствуют, поэтому production C2/C3 Not Activated, full production NOT READY.

## Current PROCESS-002 and Scope Audit

Это current full-design applicability; historical intake proposal выше его не заменяет.

| Path group | Classification / applicability |
| --- | --- |
| TASK-123 record (1) | Required: authorization, contract, decisions, traceability, exact manifests, independent handoffs/Acceptance и recovery anchor. |
| `.ai/PROJECT_CONTEXT.md`, `spec/current-state.md`, `docs/tasks/README.md` (3) | Required: TASK-122 terminal publication и truthful current TASK-123 projected state/navigation. |
| MASTER_PLAN EN/RU (2) | Required: same terminal prerequisite, active Design-only boundary и unchanged Gate 11. |
| Design indexes EN/RU (2) | Required: исправить stale acceptance-in-progress и отразить existing isolated vs planned design portions без status promotion. |
| DP-013 EN/RU (2) | Required: canonical dedicated authority/order/target/API/core implementation boundary и исправление ambiguity security/lifetime revocation. |
| DP-027 EN/RU (2) | Required: canonical exposure budgets/wire/redaction/retention/shutdown/proofs и design/implementation/activation split. |

Итого 12 Required / 0 Questionable / 0 Removable; удалить любой path нельзя без потери
DoD или обязательного зеркала/state synchronization. Size Guard: 12 paths, bounded
documentation-only subject; production/test files 0, modules/dependencies 0.
Coordinator переоценил признаки authority/core/transport как потенциально нескольких
contracts/independently deliverable behaviors: здесь выбран один end-to-end destructive
consume contract. Удаление authority, wire anti-enumeration, resource или lifetime
части оставляет DoD небезопасно неопределённым; общий E01–E16/P01–P08 проверяет их
совместно. Поэтому Design-only slice сохраняется целым, а implementation явно
разбивается на отдельные core/transport/integration gates, не запускаемые этой task.
`spec/decisions.md` N/A: новый Approved architecture/ADR или изменение Frozen model
не вводится; selected policy/API остаются локальным Draft design под этим task contract.
Root README EN/RU и CHANGELOG N/A: user-facing implemented/release capability не меняется.
DP-010/009/011/012 и DP-018, ARCH-002/004 — read-only conformance/applicability sources;
их API/ownership/Approved contracts не меняются. Historical TASK-119–122 не переписываются.

Runtime build/tests/race/vet/module checks N/A для текущего documentation-only diff;
future runtime E/P proofs описываются, не объявляются исполненными. Обязательные current
checks — EN/RU parity, source/contract/scenario conformance, local links, whitespace,
scope, raw representation, temporary-index EOL staging projection, exact final bytes
и independent Verification/Final Review. Real index/worktree/config остаются неизменны
во время staging projection; temp artifacts только вне repository.

Intended representation exception после R123-EOL001: DP-013 EN/RU сохраняют LF
representation Git baseline, а не CRLF clean checkout. Lossless premutation capture
остаётся неизменным historical evidence. Для этих двух paths equal-content EOL
сверяется с Git HEAD; все новые/изменённые строки также LF. Для остальных paths
reference — исходные raw bytes capture. Raw/staged blob equality обязательна для всех;
дополнительно normal/ignore-EOL staged-tree numstats должны совпадать с baseline,
чтобы Git filter/checkout representation не создавала EOL-only scope migration.

## Recovery Evidence Envelope

### E123-001 — Intake started

Current Owner request 2026-11-11 разрешает только intake/readiness.
Local HEAD/index/worktree и fresh origin/PR facts проверены до создания record;
branch preparation deviation и её bounded correction описаны в recovery anchor.
Task ID TASK-123 свободен. TASK-122 не возобновляется: terminal outcome реконструирован.
Первый незавершённый checkpoint: Architecture/Readiness assessment и source inventory.

### E123-002 — Architect handoff

Architect source-backed assessment получен: READY отдельно разрешённому design,
NOT READY implementation/production; 0 blocking architecture conflicts. Проверены
все requested aspects; minimum future normative scope DP-027/013 EN/RU, семь
project-state paths + record. ARCH-002/004, DP-010/011/012/018 остаются read-only.
Сам Architect файлы не менял. Coordinator назначил Documentation Agent /root
зафиксировать этот handoff, scope proposal и verdict выше.
Первый незавершённый checkpoint: validation exact record и independent final intake review.

### E123-003 — Frozen intake subject and validation

- Anchor: `6fb5b02bbff3186db86e26cb946be4b24095edfd`; object format SHA-1.
- Единственная manifest row: path
  `docs/tasks/TASK-123-RUNTIME-DIAGNOSTICS-C2-EXPOSURE-READINESS.md`, projection
  `task-record-v1`, state `present`, intended mode `100644`, projected OID
  `4972b52e7a2f37afc6567fe28bea4b8d719f7d61`.
- Canonical PROCESS-001 NUL-separated manifest OID:
  `7755798fc52fdb5a0e8ad9bfe4a59d38d60e643d`.
- Raw file SHA256 at review boundary:
  `8a5991c5edf765f4f3fa7cc2a5902302e9bd55d3f4b2a1380575ea52e8f12c16`.
  Это исторический pre-append digest, не self-hash final envelope bytes.
- Validation PASS: 11 local links; UTF-8/BOM/NUL/whitespace; один untracked record,
  пустой index, zero tracked diff; main/origin baseline unchanged; TASK-102 ref unchanged.
- Scope Audit: 1 Required / 0 Questionable / 0 Removable. Без record теряются durable
  terminal prerequisite, scope и intake verdict. Size Guard: один documentation path.
- Runtime tests N/A: executable behavior не менялся; previous hosted CI указан
  только как exact TASK-122 prerequisite evidence, не новый intake test run.

### E123-004 — Independent review and Coordinator intake closure

Независимый `/root/intake_reviewer`: **Approved**, findings **0/0** на exact
projected subject/manifest E123-003. Reviewer независимо подтвердил identity,
11 links, representation/whitespace, branch/baseline/index/scope, unchanged TASK-102
и original native P10R receipt с recorded SHA256. Source-backed requested matrix,
authority-before-lookup gap, failed-Stop и revocation boundaries признаны корректными.
Seven-path reconciliation и separate future design scope подтверждены. Branch-prep
deviation раскрыта; unresolved ref/content mutation отсутствует.

Coordinator принимает завершение **только intake/readiness deliverable**:
**READY FOR SEPARATELY AUTHORIZED C2-EXPOSURE DESIGN**.
Architecture Confirmation PASS; blocking architecture conflicts 0. Full design
не выполнен/не принят; implementation и production C2 NOT READY / Not Activated.
Дополнение E123-003/004 находится только в excluded evidence envelope;
projected intake subject остаётся invariant и повторно проверяется после append.

Следующий checkpoint — отдельное Owner решение о Design-only scope, включая
seven-path project-state reconciliation. Эта запись не активирует следующий этап.
Текущие Commit/Publication Gates отсутствуют. TASK-102 не затронута; Issue #121,
production C2/C3 и C2-Exposure Not Activated; Gate 11 OPEN.

STOP AFTER INTAKE/READINESS VERDICT.

### E123-005 — Owner authorization and Design-only admission

Owner принял READY intake и отдельно разрешил TASK-123 Design-only в exact 12 paths.
Full SHA baseline проверен read-only: HEAD/main/origin/main/live origin main равны
`6fb5b02bbff3186db86e26cb946be4b24095edfd`. Current branch остаётся существующей
`docs/task-123-c2-exposure-readiness`; index пустой, единственный dirty path — attributed
intake record. Intake final bytes подтверждаются raw SHA256 в recovery anchor;
старый one-path frozen intake subject исторический, не текущий design subject.
Premutation lossless capture всех 12 paths сохранён вне repository, SHA256
`a9a98d25bcf7e188ed75a311fee9a1e562db6b039242790031bd9d271e7f926a`; после записи identity будет дополнена в envelope.
Порядок: contract update → seven-path documentation sync/parity checkpoint →
Architect design handoff → DP-013/027 mirrored design → freeze → Independent
Verification → Independent Final Review → Coordinator Acceptance → STOP.
Без code/transport handlers/routes/workers/endpoints, activation, Commit/Publication.
Первый незавершённый checkpoint: terminal TASK-122 documentation synchronization.

### E123-006 — Documentation checkpoint before normative changes

Documentation Agent `/root/state_writer` завершил только seven-path sync; root
независимо проверил checkpoint до DP edits: **PASS**. TASK-122 source/PR/merge/tree,
durable P10R-122-001, оба CI/all-seven-success и terminal consumption синхронизированы.
TASK-123 current projected In Progress и envelope-resolution сохраняют verification
stability; никакое live Acceptance не дублируется в state surfaces.
EN/RU semantics/token parity PASS; 342 local links resolved; UTF-8/BOM/NUL/whitespace
PASS; equal-content raw EOL preservation PASS; normal/ignore-EOL numstats identical.
Два изменения EOL у content-identical blank lines были обнаружены и механически
восстановлены из premutation capture до checkpoint; semantic изменений при этом нет.
Все четыре DP-013/027 EN/RU byte-identical capture, index пустой; historical TASK-122,
production/tests/modules и Frozen/Approved files не изменены.
Root `git -c core.whitespace=cr-at-eol diff --check` exit 0.
Scope checkpoint: ровно семь tracked state paths и attributed untracked task record.
Первый незавершённый checkpoint: Architect full design handoff, затем mirrored DP edits.

### E123-007 — Architect decisions and mirrored design handoff

Architect `/root/exposure_architect` выбрал full contract, ограниченный DP-013 §32
и DP-027 §37. Дополнительно подтвердил exact refill/clock-refusal, fixed HTTP
body/header mapping, body-read/context deadlines и Composition Audit ограничения.
Documentation Agent `/root/design_writer` записал четыре normative mirrors только
после E123-006. Существующие Start/Stop/Observe contracts и DP-010 seam не меняются.
Clarifications: Reader/View остаются Owner-private; Producer идёт только через
existing immutable launch-private binding; raw capabilities не попадают во внешний API.
DP-013 design-index rows дополнительно уточняют implemented management vs Planned exposure.

Writer checks PASS: 108 local links в четырёх DP, UTF-8/BOM/NUL/whitespace,
content-identical raw EOL, identical EN/RU JSON/resource rows/E01–E16/P01–P08 IDs,
manual semantics parity, normal/ignore-EOL numstats, CR-at-EOL-aware diff check.
Root preliminary twelve-path check PASS: 461 local links, exact scope, real index
пустой, content-identical EOL changes 0. Это prefreeze checks, не independent final
verdict. Source-backed NoItem/Budget receipt equations уточнены до freeze.
Runtime proof tests added 0; regression tests added 0; executable behavior не менялся.
Remaining limitations: design-only, реальные trust/provisioning/transport/integrated
P01–P08 proofs отсутствуют и не объявляются PASS. Первый незавершённый checkpoint:
actual-DP Architecture Conformance, final scope/representation/staging checks и freeze.

### E123-008 — Recovery after usage limit and fresh preflight

Current Owner явно возобновил тот же TASK-123 Design-only, без расширения scope.
Read-only reconstruction: branch `docs/task-123-c2-exposure-readiness`, HEAD/main/
origin/main `6fb5b02bbff3186db86e26cb946be4b24095edfd`, real index пустой;
11 tracked documentation changes + только attributed untracked TASK-123 record.
Exact set совпадает с 12-path contract; production/Frozen/Approved files не изменены.

Native assignment inventory независимо прочитан через read_thread: Writer thread
`01a1ca40-17b6-7032-b590-ba9331ccb369` завершён; final editorial item
`msg_0ec3af7949dbdfbb016af44bfa37f887d2a3f729e2b2d8e71e` содержит final OIDs,
совпадающие с actual DP bytes. Verifier thread `01a1ca42-d711-7b61-8e3f-cc467ebaf5f3`
завершился usage-limit error без final Verification verdict. Actual-DP Architect
conformance turn `01a1ca48-6688-7620-bf24-11083977387c` также failed без verdict;
ранние design handoffs не заменяют conformance actual bytes. Назначен fresh Architect.

Результат прерванной staging-команды не переносится. Fresh read-only checker exit0:
12 paths, 461 local links, UTF-8/whitespace, content-identical EOL changes 0;
normal/ignore-EOL numstats identical; EN/RU JSON/proof IDs/key bounds PASS.
Fresh temporary-index preflight завершён exit0 (native command output chunk `7d92ad`):
`core.autocrlf=false`, все staged raw blobs exact; raw staged tree
`3b22d5a9f8b85f71d34addcc89b60f68efcf7351`, projected tree
`87cc501f27749488b7b246281d1dcedee5188642`. Temporary index удалён; real-index bytes,
worktree, config и attributes byte-identical before/after; реального staging нет.
TASK-102 ref по-прежнему `8bfa2e22a4b6310a71f6d36c8724b1adedd727ad`.

Current candidate manifest `887360990f0f7c204cbf011b5481e6850e69d6b7`, projected
record `4cf936251846c4687c0a035c38b15b556dea0754`. Это prefreeze candidate,
не Verification/Acceptance. Первые недоказанные gates: actual-DP Architecture
Conformance, затем final freeze, Independent Verification и Independent Final Review.

### E123-009 — Architecture Conformance and frozen M123-001

Fresh Architect `/root/recovery_architect`: **Architecture Conformance PASS**,
blocking conflicts **0**, на actual четырех DP raw OIDs, listed in manifest below.
Проверены ARCH-002/004, DP-010 §33, DP-018, actual Owner/C1/Directory/current binary;
security/lifetime cuts реализуемы без изменения Owner seam, finite resources,
authority-before-any-lookup и anti-enumeration, receipt invariants, mirrored contracts,
planned-only production boundaries соответствуют source. Protected sources unchanged.

PROCESS-002 applicability и per-path Scope Audit из current sections подтверждены:
12 Required / 0 Questionable / 0 Removable; mirrors и state synchronization обязательны,
Size Guard cohesion rationale сохраняется. Root final prefreeze check exit0:
exact scope, 461 local links, EOL/representation/whitespace, real index empty.
EOL staging projection E123-008 PASS относится к тем же projected bytes;
последующие record appends только внутри excluded envelope.

Frozen subject: **M123-001**. Anchor HEAD/main/origin/main:
`6fb5b02bbff3186db86e26cb946be4b24095edfd`; object format SHA-1.
Canonical ascending unsigned UTF-8 path order; NUL-separated
`path\0projection\0state\0mode\0oid\0` rows; Git blob hash of manifest stream.
Manifest OID: `887360990f0f7c204cbf011b5481e6850e69d6b7`.
Task-record-v1 projected OID: `4cf936251846c4687c0a035c38b15b556dea0754`.
Projected temporary-index tree: `87cc501f27749488b7b246281d1dcedee5188642`.
All 12 paths present, intended mode 100644; record untracked until separate Commit Gate.

| Path | Projection | State | Mode | OID |
| --- | --- | --- | --- | --- |
| `.ai/PROJECT_CONTEXT.md` | full | present | 100644 | `9549ffa2d7c30b2fc6c415bc7da10ac6f64c0027` |
| `docs/en/design/DP-013-runtime-management-routing.md` | full | present | 100644 | `fbc8d29c6d9f4c93952bfe31e1aed3a6608f023b` |
| `docs/en/design/DP-027-runtime-diagnostics-v1.md` | full | present | 100644 | `f2a5a4b09dd8dd18203430d28fc4823e1d3c9d91` |
| `docs/en/design/README.md` | full | present | 100644 | `7a68b49bfb43063fececae42d48e65861010b8fe` |
| `docs/en/roadmap/MASTER_PLAN.md` | full | present | 100644 | `8848312a79fa567bfdc297429c8188ef1c16dfeb` |
| `docs/ru/design/DP-013-runtime-management-routing.md` | full | present | 100644 | `259152abfca23ea4b9d9166a25fca1fa3c527487` |
| `docs/ru/design/DP-027-runtime-diagnostics-v1.md` | full | present | 100644 | `31fce4d8865af9d83297cc1d600fb73fabc071b2` |
| `docs/ru/design/README.md` | full | present | 100644 | `ad183303a8af43162291dc155069031a93183602` |
| `docs/ru/roadmap/MASTER_PLAN.md` | full | present | 100644 | `9f153af82fc0298e92ed7a284c06621caa71f184` |
| `docs/tasks/README.md` | full | present | 100644 | `02cdb0cf9bcb1a3178d0e6bcdad2d471306899bd` |
| `docs/tasks/TASK-123-RUNTIME-DIAGNOSTICS-C2-EXPOSURE-READINESS.md` | task-record-v1 | present | 100644 | `4cf936251846c4687c0a035c38b15b556dea0754` |
| `spec/current-state.md` | full | present | 100644 | `a8c508b965282af0dd8b0b5e98da4cfd8bf1219c` |

После freeze допускаются только append-only evidence в terminal envelope.
Любое projected/non-record изменение invalidates M123-001 и downstream gates.
Ранние intake manifest/Approval и preliminary Verifier notes не заменяют новый verdict.
Первый незавершённый checkpoint: Independent Verification exact M123-001.

### E123-010 — Independent representation finding and bounded rework

Independent Verifier `/root/recovery_verifier`: M123-001 **FAIL**, Major
`R123-EOL001`. Exact subject manifest `887360990f0f7c204cbf011b5481e6850e69d6b7`
и все его downstream gates **INVALIDATED**, не использовать для Acceptance/Commit.
Raw/staged equality E123-008 была действительно PASS, но не доказывала отсутствие
EOL migration относительно Git baseline: для DP-013 EN staged numstat 913/746,
RU 909/747 против ignore-EOL 176/9 и 171/9. Clean-checkout capture CRLF скрывал
Git HEAD LF representation; capture-only preservation было недостаточно.

Coordinator направил ограниченный mechanical rework Documentation Agent /root:
только DP-013 EN/RU восстановлены в Git baseline LF, включая новые строки;
UTF-8 semantic lines unchanged. Исходный lossless capture сохранён; Git config/
attributes/real index не менялись. Task contract явно фиксирует intended Git-LF
reference exception, остальные 10 paths не получают normalization.
Это исправление в пределах authorized 12 paths, не scope expansion/design decision.
Необходимо заново построить M123-002, повторить affected Architecture Conformance,
baseline-aware staging/EOL, Independent Verification и затем Independent Final Review.

### E123-011 — Corrected staging, fresh conformance and M123-002 freeze

Affected Architecture Conformance `/root/recovery_architect`: **PASS**, conflicts 0
на новых DP-013 raw OIDs. Независимо подтверждены LF baseline и semantic-line equality;
DP-027 byte-identical прежнему handoff, Frozen/Approved/production sources unchanged.
Root corrected preflight exit0 (output chunk `d9ab9f`): raw/staged blobs exact,
baseline-aware normal/ignore-EOL staged-tree numstats identical; unintended migration 0.
Real-index/worktree/config/attributes bytes unchanged, temp index удалён.
Warnings о будущем checkout conversion не являются mutation; current accepted bytes LF.
Raw staged tree до нового envelope append: `2723c154cdb584a57456d867c6a60cb6f82641d4`.
Projected tree: `963efd05c9e661040b73306562bee63b217e6836`.

PROCESS-002/Scope Audit повторно проверены: 12 Required / 0 Questionable / 0 Removable;
461 local links, UTF-8/whitespace, raw representation и mirror semantics сохранены.
R123-EOL001 исправлен в source representation; independent closure ещё требует
Verification нового subject. M123-001 остаётся INVALIDATED навсегда.

Frozen subject: **M123-002**. Anchor HEAD/main/origin/main:
`6fb5b02bbff3186db86e26cb946be4b24095edfd`; object format SHA-1.
Canonical ascending unsigned UTF-8 paths; NUL-separated
`path\0projection\0state\0mode\0oid\0` rows, Git blob hash.
Manifest OID: `f45315701c06bbc8ae18ad777fd47e679937f3a1`.
Task-record-v1 projection: `9e7ef5e72680435e25c9f7e495490ad0cccf46a0`.

| Path | Projection | State | Mode | OID |
| --- | --- | --- | --- | --- |
| `.ai/PROJECT_CONTEXT.md` | full | present | 100644 | `9549ffa2d7c30b2fc6c415bc7da10ac6f64c0027` |
| `docs/en/design/DP-013-runtime-management-routing.md` | full | present | 100644 | `6a2f8a86c936ca2504e903d45fdea0cc24b1648e` |
| `docs/en/design/DP-027-runtime-diagnostics-v1.md` | full | present | 100644 | `f2a5a4b09dd8dd18203430d28fc4823e1d3c9d91` |
| `docs/en/design/README.md` | full | present | 100644 | `7a68b49bfb43063fececae42d48e65861010b8fe` |
| `docs/en/roadmap/MASTER_PLAN.md` | full | present | 100644 | `8848312a79fa567bfdc297429c8188ef1c16dfeb` |
| `docs/ru/design/DP-013-runtime-management-routing.md` | full | present | 100644 | `fa53939e798e37d215764383a04babcad863554a` |
| `docs/ru/design/DP-027-runtime-diagnostics-v1.md` | full | present | 100644 | `31fce4d8865af9d83297cc1d600fb73fabc071b2` |
| `docs/ru/design/README.md` | full | present | 100644 | `ad183303a8af43162291dc155069031a93183602` |
| `docs/ru/roadmap/MASTER_PLAN.md` | full | present | 100644 | `9f153af82fc0298e92ed7a284c06621caa71f184` |
| `docs/tasks/README.md` | full | present | 100644 | `02cdb0cf9bcb1a3178d0e6bcdad2d471306899bd` |
| `docs/tasks/TASK-123-RUNTIME-DIAGNOSTICS-C2-EXPOSURE-READINESS.md` | task-record-v1 | present | 100644 | `9e7ef5e72680435e25c9f7e495490ad0cccf46a0` |
| `spec/current-state.md` | full | present | 100644 | `a8c508b965282af0dd8b0b5e98da4cfd8bf1219c` |

Только append-only excluded envelope допускается после freeze.
Первый незавершённый checkpoint: fresh Independent Verification M123-002;
после PASS назначается отдельный Independent Final Reviewer.

### E123-012 — Independent Verification M123-002

Independent Verifier `/root/recovery_verifier`: **PASS**, findings **0/0**.
Fresh independent command exit0, output chunk `f13341`; не использован прежний
capture-only вывод M123-001. Независимый hashlib Git-blob reconstruction дал
manifest `f45315701c06bbc8ae18ad777fd47e679937f3a1`, record projection
`9e7ef5e72680435e25c9f7e495490ad0cccf46a0`; все 12 rows/modes соответствуют freeze.

R123-EOL001 **CLOSED**: DP-013 EN/RU LF соответствует Git baseline, semantic lines
M123-001 сохранены. Corrected staged tree `2723c154cdb584a57456d867c6a60cb6f82641d4`
имеет identical normal/ignore-EOL numstats на всех 12 paths; non-record raw staged
blobs и projected tree `963efd05c9e661040b73306562bee63b217e6836` exact.
461 local links, четыре новых cross-anchors, mirror JSON/resource tables/E01–E16/
P01–P08 IDs, UTF-8/whitespace/representation PASS. Scope/base/real index/TASK-102 PASS.

Source-backed scenario verification PASS: authority-before-any-lookup, exact grant/
attempt, fixed anti-enumerating errors, finite quotas/ingress, single-adapter audit,
C1 stop-specific equations, security/lifetime cuts, retention/redaction/nonjoining
shutdown и production gates. Это design verification, не исполнение future E/P tests.
Runtime tests N/A для documentation-only diff; production/future integration не
активированы. Verifier файлы не менял. M123-001 остаётся INVALIDATED.
Первый незавершённый checkpoint: Independent Final Review M123-002.

### E123-013 — Independent Final Review M123-002

Independent non-author Reviewer `/root/final_design_reviewer`: **Approved**,
findings **0/0**, exact manifest `f45315701c06bbc8ae18ad777fd47e679937f3a1`,
record projection `9e7ef5e72680435e25c9f7e495490ad0cccf46a0`.
Fresh independent commands exit0, output chunks `a79847` и `c3af2d`: hashlib
manifest, exact 12 paths/modes/base/empty real index, raw/projected tree equality,
baseline-aware EOL, whitespace, 461 links и четыре новых anchors, EN/RU JSON/proof
parity и unchanged TASK-102. Envelope E123-012 не изменил frozen subject.

Reviewer подтвердил R123-EOL001 closure и все requested contract aspects against
ARCH-002/004, DP-010/018 и actual Owner/C1/Directory APIs. Security admission и
lifetime retirement независимы; нового Owner seam или production capability нет.
PROCESS-002/Scope Audit: **12 Required / 0 Questionable / 0 Removable**; каждый path
необходим contract/mirror/terminal state synchronization/durable evidence. Size Guard
cohesion обоснована. Runtime tests N/A; future E01–E16/P01–P08 не исполнены.
Production C2/C3 и C2-Exposure Not Activated; Gate 11 OPEN. Reviewer файлы не менял.
Первый незавершённый checkpoint: Coordinator Acceptance M123-002 после final integrity.

### E123-014 — Coordinator Acceptance and Design-only closure

Acceptance subject: **M123-002**, manifest
`f45315701c06bbc8ae18ad777fd47e679937f3a1`, task-record-v1 projection
`9e7ef5e72680435e25c9f7e495490ad0cccf46a0`; exact 12 rows E123-011, anchor
`6fb5b02bbff3186db86e26cb946be4b24095edfd`.

Coordinator final gate: Architecture Conformance PASS on corrected raw OIDs;
Independent Verification PASS 0/0; Independent Final Review **Approved** 0/0;
R123-EOL001 independently CLOSED, M123-001 remains INVALIDATED. PROCESS-002, EN/RU
parity/JSON/proof obligations, 461 links/four new anchors, exact scope/Size Guard
and Scope Audit **12 Required / 0 Questionable / 0 Removable** PASS.
Final pre-acceptance command exit0, chunk `998ee9`: raw/staged equality, baseline-aware
normal/ignore-EOL staged numstats identical, corrected projected tree
`963efd05c9e661040b73306562bee63b217e6836`; real index/worktree/config/attributes
unchanged. Raw staged tree with then-current evidence was
`0af2270ba67de9a9c686813c2ea5aaec5d5bda19`; it is a historical staging proof,
не OID будущего commit и не self-attestation final envelope bytes.
Final manifest invariant after E123-012/013; Acceptance append excluded by projection.

Definition of Done выполнен для **Design-only**: DP-013/027 EN/RU задают dedicated
consume authority до lookup, exact grant/attempt binding, anti-enumeration, planned
mTLS/HTTP POST, finite process/principal/Target budgets и single-adapter audit,
C1 receipt invariants, retirement/revocation races, redaction/retention/trust boundaries
и E01–E16/P01–P08 acceptance obligations. Seven-path TASK-122 terminal/project-state
reconciliation выполнена до normative changes. Frozen/Approved sources не менялись.

> ACCEPTED — COORDINATOR ACCEPTANCE OF TASK-123 DESIGN M123-002

DP Design Status остаётся **Draft**; task Acceptance не является нормативным Approval
Frozen/Approved architecture или production implementation. Runtime tests N/A для
documentation-only diff; future E/P executable proofs не объявляются выполненными.
Следующий recommended candidate — отдельный intake isolated diagnostics admission/
resource core (`diagnostics.go`, `diagnostics_test.go` в runtimemanagement), **Not Activated**;
его implementation требует отдельной authorization и scope/coverage проверки.
Transport и real production integration/provisioning также не активированы.

Current branch `docs/task-123-c2-exposure-readiness`; HEAD/main/origin/main baseline
сохранён; real index пустой, 11 modified tracked docs + один attributed untracked
TASK-123 record. Ни Commit Gate, ни Publication Gate не выдавались; stage в real index,
commit, push, PR и merge не выполнялись. TASK-102 не затронута; Issue #121 Not Activated;
production C2/C3 и C2-Exposure Not Activated; Gate 11 OPEN.

STOP AFTER DESIGN AND ACCEPTANCE VERDICT.

### E123-015 — Commit Gate authorization and pre-commit verification

Owner exact command `Разрешаю коммит.` разрешила ровно один Accepted Task commit
из M123-002 после E123-014. Publication/push/PR/merge не разрешены.
Fresh Commit Gate reconstruction: branch `docs/task-123-c2-exposure-readiness`,
HEAD/main/origin/main `6fb5b02bbff3186db86e26cb946be4b24095edfd`, real index пустой,
exact 12-path set без посторонних/generated/temp files. Accepted final record bytes
до этого append совпали с сохранённым final SHA256; все остальные 11 raw OIDs
и canonical manifest `f45315701c06bbc8ae18ad777fd47e679937f3a1` повторно exact.
Fresh checker exit0: 461 local links, representation/UTF-8/EOL, scope/index PASS;
`git -c core.whitespace=cr-at-eol diff --check` exit0. Runtime tests N/A — docs only,
accepted independent Verification/Review остаются привязаны к тем же projected bytes.

Commit message policy CONTRIBUTING (imperative + scope):
`docs(diagnostics): define C2 exposure admission contract`.
Real staging выполняется только теперь, с per-invocation `core.autocrlf=false`,
по exact 12 paths; это сохраняет accepted raw blobs и DP-013 Git-baseline LF.
Перед commit обязательны raw/staged blob equality, exact staged scope/modes,
baseline-aware normal/ignore-EOL equality и staged whitespace check. Config и
attributes не меняются; GPG/DCO/sign-off не добавляются новой policy.
Этот append только evidence envelope, M123-002 projection invariant.
Commit/tree OIDs после выполнения реконструируются из Git, не записываются
самоссылочным изменением этого commit. После одного commit — STOP без Publication Gate.

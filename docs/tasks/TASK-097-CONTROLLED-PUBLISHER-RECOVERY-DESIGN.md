# TASK-097 — Controlled Publisher Recovery Design

## Status

Completed — Coordinator Accepted (replacement), 2026-10-18, R-097-005.
Accepted replacement manifest 5a583be8e6e935c1f556f4bc4bdb58b8c6c975b6.
Original E001–E008 approvals historical only; newest matching R097 entry resolves current closure.
DP-025 Draft / Planned; no operational activation or publication authority.

## Task Contract

### Replacement/Rework Contract — current

2026-10-18. Current explicit user authority — только bounded replacement/rework
preparation TASK-097. Original source e4603d86cc0179fc29f5f26d0a0b4485e8280215,
parent82cf49627e54881599bb60a55f28c8e9c6b0c4db,
treefe23714e4b1b2163f6053c67bba2cf358696f4e6,
manifest523bc07b8145e6c39be42852cf8a4b76f0bf10b4 immutable;
original branch docs/task-097-controlled-publisher-recovery-design сохраняется.
Current anchor/base c8047fd08caa79480c2bc73b5c382f4aa5079db7 — fresh qualified
remote main, treec6c5e6a78140cf66e42a35c474504bd8e55fb39c;
current branch docs/task-097-controlled-publisher-recovery-isolated.
Reason: original main..source включал TASK-096; новый Target должен исключать
его commit/record и содержать только TASK-097 changes.

Current DoD поверх сохранённого design contract:
1. Independent clean baseline/intake, main provenance и original ref preservation;
   task record первым content change, ровно прежние8scope paths.
2. Перенос raw source files: indexes unchanged; DP mirrors меняют только один
   TASK-096 link каждый; три nav paths только current097 branch/base.
3. Три relative links заменить проверенным immutable GitHub commit82cf/blob
   locator; historical evidence не объявляется publication/ownership authority.
4. Current branch/base/recovery anchors и новый rework contract отделены от
   historical original intake; original E001–E008 сохранены raw append-only prefix.
5. Raw equivalence/delta ledger, fresh manifest; повторные Architecture impact/
   final Confirmation, independent Tester, EN/RU/link checks, PROCESS002,
   ScopeAudit, non-author Review, новая Acceptance и post-decision integrity.
6. Future replacement parent только main; base не содержит096commit/record.
   До separate Commit Gate commit не существует; readiness qualified только
   для подготовки commit, не для publication.

A/B/C, I01 и Draft/Planned не меняются; C operative. Normative contracts/tooling,
oldbranches/records и operational state не изменяются. Нет stage/commit/push/
PR/merge/fetch/pull/cleanup, operational ownership events или recovery.
Commit требует нового отдельного user gate; прежнее one-commit permission
использовано original commit. Old approvals historical only; current verdict
берётся только из newest valid matching R097 envelope. Next candidate Not Activated.
STOP при provenance mismatch, unqualified baseline, scope expansion или
непройденном обязательном gate. Publication/implementation/migration не начинаются.


### Task Mode

Documentation-only / Design. Новое требование владельца проекта:
исследовать controlled succession после невосстановимой потери original Publisher.
Confirmed governance gap не установлен. Эта TASK не изменяет контракты.

### Why Now

Explicit limited Coordinator intake разрешён 2026-10-18 после READ-ONLY
Pre-Change Plan Gate. TASK-096 до merge и TASK-095 после merge используются
только как bounded forensic fixtures, без выполнения их recovery.

### Definition of Done

1. Независимо подтвердить ordinary authoring intake, exact clean baseline,
   свободный ID, безопасную ветку и task record первым content change.
2. Получить explicit Architect Confirmation bounded зеркального proposal.
3. Сравнить A Direct Publisher с доказуемым fencing, B Fenced Mutation Gateway
   и C действующий fail-closed STOP по всем заданным safety boundaries.
4. Разделить действующие контракты, proposed guarantees и Unknown provider
   capabilities; определить evidence, authority, successor admission, CAS/fencing,
   in-flight effects, crash recovery, competing successors и lost responses.
5. Описать candidate state machine, adversarial traces, prospective activation
   и ограничения historical migration без активации.
6. Завершить independent Verification/Tester, PROCESS-002, Scope Audit,
   distinct non-author Review, Coordinator Acceptance и post-decision integrity.

### Out of Scope

PROCESS-001/002, Publisher/Coordinator и другие role contracts, tooling,
production code, tests, ADR/ARCH, existing DP contents и старые task records
не изменяются. Provider setup, credentials, control store и gateway не создаются.
Не выполнять InitialDispatch, ownership/recovery/terminal events, native history
edits, resume/reconnect, publication, push, merge, fetch, pull или cleanup.
Обязательства и permissions TASK-095/096 не передаются. TASK-092 не затрагивается.
Commit и publication требуют отдельных explicit user permissions.

### Verification Plan

Existing coverage: PROCESS-001 interruption/initial-owner/handoff/terminal
semantics; Publisher scenarios S-027,S-033–S-040,S-056–S-082; TASK-096 assessment.
Это normative/manual trace sources, не provider capability certificates.
Проверить canonical raw-byte manifest, task-record-v1, eight-path diff,
protected baseline equality, index и старые refs, relative links, mirror
section/table parity, claim-to-evidence boundaries и A/B/C adversarial traces.
Executable tests N/A: documentation-only comparison, runtime behavior не меняется.
Live provider mutation и recovery tests запрещены.

## Original Intake, Baseline and Scope (historical source e460)

Repository E:/wikiPRJ/universal-websocket-platform; object format SHA1.
Anchor HEAD 82cf49627e54881599bb60a55f28c8e9c6b0c4db;
tree 83c8d22bf920275f9c701389fc5b9b7cb1c73d24.
Ветка docs/task-097-controlled-publisher-recovery-design от exact HEAD.
Подготовительные cleanup, checkout, pull и synchronization не выполнялись.
Новая ветка создана после admission; checkout hooks отключены только для этой
invocation, permanent Git configuration не менялась.

Independent intake /root/intake097: ALLOWED. TASK-096 E-096-007/E-096-008:
ordinary authoring Completed, accepted manifest
1e065935ded73cd69c89b7f60a8359488509697d; четыре committed/index/worktree
raw paths совпали. Full record SHA256
d365b915631aa644f7b98db2ee1d516145b10f1a9a7edad57aabe6dbb495f21b.
Blocked ordinary publications не означают незавершённую authoring автоматически.
ND/BCC sealing barriers не расширяются на ordinary Accepted class.
Local main/cached origin/main 8cc432c3b6c751daa3e2daf4d791e54d99a74c75 —
понятное historical lag, не синхронизируется этой TASK.
TASK-095 ref 6a35b76f7678a06f65c24eb06de9bc8e0f6149bd сохранён;
TASK-096 ref остаётся anchor HEAD. Conflicting active authoring не обнаружена;
historical projected In Progress entries resolve через matching envelopes.
TASK-097, DP-025 и новая ветка были свободны в проверенном inventory.

Exact authorized manifest; все остальные пути protected:

| Path | Projection | Necessity |
| --- | --- | --- |
| .ai/PROJECT_CONTEXT.md | full | Stable current authoring scope |
| docs/en/design/DP-025-controlled-publisher-recovery.md | full | English proposal |
| docs/en/design/README.md | full | Separate Publisher-governance discovery |
| docs/ru/design/DP-025-controlled-publisher-recovery.md | full | Russian mirror |
| docs/ru/design/README.md | full | Mirrored discovery |
| docs/tasks/README.md | full | Current task navigation |
| docs/tasks/TASK-097-CONTROLLED-PUBLISHER-RECOVERY-DESIGN.md | task-record-v1 | Contract, evidence and closure |
| spec/current-state.md | full | Stable design/implementation boundary |

Size Guard: восемь documentation paths, один inseparable mirrored A/B/C
comparison, ноль production lines/packages/activated operational contracts.
Индексы сохраняют Runtime classifications; root guides не меняются.
Projected navigation использует verification-stable In Progress/envelope resolver.

## Sources, Roles and Documentation Applicability

Repository first: [AGENT](../engineering/AGENT.md),
[PROCESS-001](../engineering/PROCESS-001-AI-DEVELOPMENT-WORKFLOW.md),
[PROCESS-002](../engineering/PROCESS-002-DOCUMENTATION-SYNCHRONIZATION.md),
[Publisher](../engineering/agents/publisher.md),
[TASK-096 immutable source](https://github.com/dsdred/universal-websocket-platform/blob/82cf49627e54881599bb60a55f28c8e9c6b0c4db/docs/tasks/TASK-096-FAILED-OWNER-RECOVERY-ASSESSMENT.md).
TASK-095 fixtures anchored через TASK-096 F096-01–F096-12.
Publication observations TASK-096 — historical qualified native evidence,
не fresh remote outcome и не authority; exact anchors фиксируются в DP.

Coordinator /root: intake/acceptance. Independent Architect определяет design
boundaries; Documentation author фиксирует их. Independent Verifier/Tester и
distinct non-author Reviewer attests exact content identity.
Task record был первым content change; Architecture Confirmation на intake
ожидался. Действительные последующие handoffs находятся в append-only envelope.

PROCESS-002: task, DP mirrors, design indexes, current-state, task navigation
и PROJECT_CONTEXT Required. MASTER_PLAN N/A: milestone/dependency не активируется.
spec/decisions и ADR/ARCH N/A: Draft comparison не принимает durable architecture
choice. CHANGELOG N/A: user-facing release отсутствует. Normative contracts,
старые records и tooling protected.

## STOP Conditions and Next Gates

Unknown provider capability — explicit result, не доказательство availability.
STOP при authority/task conflict, unqualified baseline, scope expansion,
непройденном mandatory role gate или требовании impersonation/history rewrite/
recovery. Current STOP действует до отдельно согласованной normative activation.
Следующий candidate только Not Activated: provider evidence/architecture decision
или contract amendment; implementation/migration автоматически не начинаются.
Commit/publication — отдельные gates.

## Recovery Evidence Envelope

### E-097-001 — Independent Intake and First Content Change

2026-10-18. Independent /root/intake097 ALLOWED after read-only raw manifest,
closure, free ID/branch and ordinary authoring checks. Corrected independent
read command6823d7 reproduced the 44495-byte predecision SHA256
014114734d3da3e0beae1b8632dda7e7e66b83170d9fe447a9cc18ec1f83ff70.
Root rechecked clean HEAD, four-path accepted subject and raw baseline of
494 tracked paths. Index raw SHA256 before intake:
f8a99a284e98b930549be27f638766a9fd2e457a398ba545f38b9affca9924f5.
Branch command chunk1e225e exit0 created only new TASK-097 ref from anchor.
This record is first content change; no stage/commit/publication/recovery.


### E-097-002 — Explicit Architecture Confirmation перед Writer

2026-10-18. Independent /root/architecture097 CONFIRMED exact eight-path Draft
DP-025 scope. Design Status Draft; Implementation Planned / not implemented.
A требует complete enforcement всех local/remote/legacy paths; B только conditional
research preference с mediation, durable journal, serialized queue, fenced gateway
и reconciliation accepted requests. C остаётся operative contract. CAS, native
readback, credential fencing и external effects — разные domains; shared
atomicity/exactly-once не установлены. Successor только own attributed actor;
old Release/InitialDispatch/run не фабрикуются. Prospective activation и legacy
migration требуют отдельных gates; Unknown сохраняет STOP.

Writer /root/writer097 назначен после CONFIRMED; готовит семь non-task paths.
Final exact-subject Architecture Confirmation следует после freeze. Canonical
русский Task Contract уточнён до subject freeze; исходный first-content-change
record SHA256 a2c9db52e4d49fb9a92c1dd40dc286e4b28a61235ed2e61e919cf0e9206d2231
сохранён как chronology anchor; E001 raw bytes не переписаны. Protected489 raw
aggregate e7afa7bbc454bd50861473977640e0317fa689e0ad6e6db7e0415cebd9ac1fe1.

### E-097-003 — Writer Handoff, Frozen Subject и PROCESS-002

2026-10-18. Documentation author /root/writer097 завершил семь назначенных
non-task paths после E002; record остался Coordinator-owned. DP-025 EN/RU
Draft / Planned (not implemented), 267 lines,12sections,30traces T01–T30,
8invariants I01–I08,10evidence rows E01–E10; mirrored statuses/structure.
Writer checks chunkdb9b02 exit0: links/parity/historical suffix PASS.
Предварительный distinct Reviewer потребовал явно указать exclusive gateway-owned
Git workspace/index/refs и P6–P9 mediation; Writer включил это в обе mirrors.
A/B остаются proposed guarantees с Unknown provider qualification; C operative.
Permission carry-over — OPEN policy decision; migration/activation не разрешены.

Root audit chunk996be1 exit0 на frozen eight-path raw subject:
SHA1manifest a2c08c84e179c1777d52c718164bf58d05c0d2c5,784canonicalNULbytes.
Anchor HEAD82cf49627e54881599bb60a55f28c8e9c6b0c4db,branch097.
Ordered rows (path | projection | state | mode | projected blob):
.ai/PROJECT_CONTEXT.md | full | present | 100644 | 80065770a9cc294fa3ec3f1a3cb59a1d1eee43b1
docs/en/design/DP-025-controlled-publisher-recovery.md | full | present | 100644 | fa23fdd3e147ff837fce6c111726fc5eb263266f
docs/en/design/README.md | full | present | 100644 | 235a358a50791d980fae690a738ae59b2a142d01
docs/ru/design/DP-025-controlled-publisher-recovery.md | full | present | 100644 | b974ea5a3d582bebb9ed38ac5e7aaf2551d47ab6
docs/ru/design/README.md | full | present | 100644 | eb856d72aeb19b48784acd34d2e3d117ed282f94
docs/tasks/README.md | full | present | 100644 | fb7d292c458c620fd684839072110fbcf04e3441
docs/tasks/TASK-097-CONTROLLED-PUBLISHER-RECOVERY-DESIGN.md | task-record-v1 | present | 100644 | 4ed718f2d6aff5928903bf918bf0f330c54dd8f9
spec/current-state.md | full | present | 100644 | af81b7b1222a26d298893663e9cc577e16ef4eb9

Root checks:218relative links/0broken; whitespace/conflicts0; protected489raw
SHA256e7afa7bbc454bd50861473977640e0317fa689e0ad6e6db7e0415cebd9ac1fe1;
index rawSHAf8a99a284e98b930549be27f638766a9fd2e457a398ba545f38b9affca9924f5
unchanged/no stage. HEAD/oldlocalrefs/main/cachedmain unchanged; new097branch only.
Historical navigation suffix beginning at paragraph TASK-095 preserved raw;
initial broad substring matcher corrected to paragraph anchor before PASS.
E001 raw721-byte prefix preserved SHA256
720f17eb245b6f74c3cdbab674bc9359ad3bec1f2dcf68b3652ffb512dd6a4d4.
Нормативные contracts/tooling/oldtasks не менялись; no remote/native mutations.

PROCESS-002 SYNCHRONIZED для bounded Draft: current097stableInProgress/envelope
resolver, historical096Accepted отделён от publication completion; public indexes
с отдельным Publisher-governance subsection сохраняют Runtime entries.
Applicability и Size Guard — как в contract. ScopeAudit8Required/0Questionable/
0Removable: task contract/evidence; one DP mirror pair; two required discovery
indexes; three mandatory navigation/context paths. Ни один путь не removable
без потери DoD или обязательной synchronization.
Independent final Architecture/Verification/Review pending; Acceptance не выдан.

### E-097-004 — Architecture Rework и Revised Subject

2026-10-18. Independent Architect /root/architecture097 withheld final
confirmation старого a2c08... subject: I01 должен гарантировать одного owner
для transaction across all epochs. Это bounded wording rework, не изменение
scope/новый operational contract. Старые downstream verdicts не переносятся.
Writer chunkd8ad8b exit0 изменил только I01 в EN/RU; reverse-substitution raw
equality проверена. Новый I01 требует enforced rejection всех prior epochs и
reconciliation/drain их in-flight operations до admission нового epoch.
Reviewer предварительно считал остальные fencing clauses достаточными;
Architect tightening принят Writer до final gates.

Revised raw subject SHA1manifest 523bc07b8145e6c39be42852cf8a4b76f0bf10b4,784NULbytes/8rows:
.ai/PROJECT_CONTEXT.md | full | present | 100644 | 80065770a9cc294fa3ec3f1a3cb59a1d1eee43b1
docs/en/design/DP-025-controlled-publisher-recovery.md | full | present | 100644 | e2ebeb68d0d703a2b2e9de2c16056645d9cfbb48
docs/en/design/README.md | full | present | 100644 | 235a358a50791d980fae690a738ae59b2a142d01
docs/ru/design/DP-025-controlled-publisher-recovery.md | full | present | 100644 | 4eb73a022fdffbaece6ec3d70c1636aa04f78450
docs/ru/design/README.md | full | present | 100644 | eb856d72aeb19b48784acd34d2e3d117ed282f94
docs/tasks/README.md | full | present | 100644 | fb7d292c458c620fd684839072110fbcf04e3441
docs/tasks/TASK-097-CONTROLLED-PUBLISHER-RECOVERY-DESIGN.md | task-record-v1 | present | 100644 | 4ed718f2d6aff5928903bf918bf0f330c54dd8f9
spec/current-state.md | full | present | 100644 | af81b7b1222a26d298893663e9cc577e16ef4eb9

Root chunk6ce1d8 exit0 повторил full audit:218links/0broken,conflicts/trailing0,
protected489SHAe7afa7bbc454bd50861473977640e0317fa689e0ad6e6db7e0415cebd9ac1fe1,
unchangedrawindex/HEAD/main/cachedmain/oldrefs и eight-path ScopeAudit8/0/0.
PROCESS-002 applicability unchanged. Final Architect, Tester и distinct
non-author Reviewer проверяют этот manifest заново; Acceptance pending.

### E-097-005 — Final Architecture, Independent Verification и Scope Audit

2026-10-18. Independent /root/architecture097: Final Architecture Confirmation
CONFIRMED на SHA1manifest523bc07b8145e6c39be42852cf8a4b76f0bf10b4/8rows784B,
independent raw/task-record-v1 command4d1d16 exit0; RU readback f6d6d8 exit0.
I01 resolved across all epochs; остальные handoff constraints выполнены.
Это подтверждение bounded Draft, не Approved operational contract/provider
guarantees/activation. A/B не активированы, C operative.

Independent /root/intake097 в отдельно назначенной роли Verifier/Tester:
FINAL PASS,0blocking/0actionable, тот же exact eight-path manifest E004.
Command881772 exit0: Python stdlib raw SHA1 Gitblob/task-record-v1/canonical NUL
с independent git hash-object --stdin crosscheck на каждом subject; все8rows
совпали E004. Taskprojection4ed718f2d6aff5928903bf918bf0f330c54dd8f9,
ENe2ebeb68d0d703a2b2e9de2c16056645d9cfbb48,
RU4eb73a022fdffbaece6ec3d70c1636aa04f78450.
Полный reproduction: взять PATHS/ordered rows E004, raw read каждого present
path, применить PROCESS-001 task-record-v1 только к task record; остальные
full; canonical path/projection/present/100644/OID NUL fields и trailing NUL,
git hash-object --stdin без -w. Hash object stream независимо равен523bc...
Никаких stage или newline filters. Source/proof artifacts — actual eight raw
files, exact rows E004 и protected tracked inventory excluding этих8paths.

881772 также:218relative links/0broken;12matching mirror sections,30matching
traceIDs;conflicts0/trailingwhitespace0;HEAD/oldrefs/main/cachedmain/index
unchanged. Protected489rawaggregate
e7afa7bbc454bd50861473977640e0317fa689e0ad6e6db7e0415cebd9ac1fe1.
Aggregate recipe: UTF8-byte-sorted tracked paths excluding scope; concatenate
path,NUL,SHA256(rawfile) lowercasehex,NUL; SHA256 итогового stream.
Commandaf2612 exit0: полный RU semantic read и
git -c core.whitespace=cr-at-eol diff --check PASS.
Command2a8d39 exit0:3historicalnavigation suffixes raw identical to anchor,
2Runtime index prefixes unchanged;8IIDs/10EIDs mirror parity.
Root313ad2 exit0 подтвердил diff --check и отсутствие staged paths.

Manual T01–T30 PASS30/30 как DESIGNTRACE, не provider execution/model guarantee.
Проверены competing successors, prior/staleepoch, credential/workspace bypass,
partial fence/reservation/admission crashes, lostCAS/API results, accepted
in-flight/revoke races, immutable Target, prospective/legacy limits, terminal
readback и обе forensic phasefixtures без replay. I01 по всемepochs, Bexclusive
workspace/P6–P9 и fullmediation явно заданы; capabilityUnknown не повышен.
Existing coverage — cited normative suites/fixtures, новое покрытие30designtraces.
Executable/live tests N/A по documentation-only scope; provider behavior,
external atomicity и actual recovery capability не сертифицированы.

Independent ScopeAudit8Required/0Questionable/0Removable: one DP mirroredpair,
taskrecord, two discoveryindexes, three mandatory navigation/context files.
PROCESS-002 SYNCHRONIZED; sourceprecedence/Draftstatus/applicability/SizeGuard
сохранены. Provider qualification/permissioncarryover/implementation/migration
остаются explicit future requirements, не скрытыми documentary failures.
Distinct non-author final Review next; Coordinator Acceptance pending.

### E-097-006 — Final Independent Non-Author Review

2026-10-18. Distinct non-author /root/reviewer097: APPROVED,
0blocking/0nonblocking. Exact reviewed manifest
523bc07b8145e6c39be42852cf8a4b76f0bf10b4,8rows784B; taskprojection4ed718...,
ENe2ebeb68d0d703a2b2e9de2c16056645d9cfbb48,
RU4eb73a022fdffbaece6ec3d70c1636aa04f78450.
Independent raw reproduction44fa28 exit0; final post-E005 reproduction7eb5e4
exit0; durable E004/E005 readback4dece1 exit0. Preliminary interrupted
review не используется как final verdict.

Reviewer подтвердил A complete direct/legacy/subprocess/queued fencing;
B credential плюс exclusive workspace/index/ref mediation, serial dispatch,
fenced leadership, durable journal/crash recovery; C operative. CAS/enforcement/
external atomicity разделены; I01 across all epochs; actual recoveryauthority/
successor attribution/admission/revoke races/lostresponses/immutableTarget/
readback описаны. Mirrors12sections/30traces/8invariants/10evidence rows
совпадают структурно и семантически. Historical facts не выдаются за current
provider/remote/source eligibility; нет approved recovery contract/activation.

Independent ScopeAudit8Required/0Questionable/0Removable с явным removability
ответом No для каждого пути: task contract/evidence; ENDP principal deliverable;
RU mandatory mirror; EN/RUindexes discovery; tasksREADME current navigation;
PROJECT_CONTEXT current authoring context; current-state design/capability boundary.
Ни одна из этих правок не удаляется без потери DoD или обязательной sync.
PROCESS-002/SizeGuard/applicability согласованы, next candidate Not Activated.

Ограничения: designtraces30/30 не live/executable recovery/provider proofs.
Capabilities/permissioncarryover/enforcement deployment/legacy migration требуют
отдельных gates. Reviewer не редактировал files/operational state; protected/
refs/index equality подтверждена independent Tester/root evidence.
Coordinator может выполнить Acceptance/post-decision integrity только exact
subject; commit/publication/recovery эта Review не разрешает.
Root850651 exit0 после E005 повторно подтвердил весь eight-path subject,
218links,protected489,index/HEAD и все oldheads/cachedremote refs; добавлен
только новый097ref отanchor82cf. Нормативные contracts/tooling unchanged.

### E-097-007 — Coordinator Acceptance и STOP Boundary

2026-10-18. Coordinator /root явно принимает TASK-097:
**Coordinator Acceptance: ACCEPTED; Completed — Coordinator Accepted**.
Принят только Documentation-only / Design exact eight-path subject:
SHA1manifest523bc07b8145e6c39be42852cf8a4b76f0bf10b4,8rows784canonicalNULbytes,
все ordered rows E004 неизменны; taskprojection4ed718f2d6aff5928903bf918bf0f330c54dd8f9.
AnchorHEAD82cf49627e54881599bb60a55f28c8e9c6b0c4db;
branch docs/task-097-controlled-publisher-recovery-design.

Все шесть DoD выполнены: independent ordinary authoring intake ALLOWED;
firstcontentrecord/freeID/cleanexactbaseline/safebranch; explicit Architect
Confirmation до Writer и final exactsubjectCONFIRMED; полное A/B/C comparison
и Unknown capability/evidence boundaries; candidate state machine/adversarial
traces/prospectiveactivation/historicalmigration; independentTesterPASS30/30
manualdesigntraces,218links/0broken,mirrors/489protected/index/refs; PROCESS002
SYNCHRONIZED; ScopeAudit8Required/0Questionable/0Removable; distinct non-author
Review APPROVED0/0. I01 rework устранён, affected gates повторены на523bc...
Root858f40 exit0 после durable E006 снова подтвердил current exactsubject.

DP-025 остаётся **Draft / Planned (not implemented)**. Принят результат design
TASK, не operational protocol и не утверждение architectureA/B. Coperative.
Provider CAS/fencing/externalatomicity не сертифицированы; new requirement
не объявлен confirmed governance gap. Permissioncarryover — OPEN policy
decision. Никакой source/successorowner, Release/Route/Accept, terminal event,
P10/NoneTerminal или recovery eligibility эта Acceptance не создаёт.
TASK-095/096 operational obligations/permissions/history не изменены;
TASK-092 не затронута. Runtime/tooling/normativecontracts protected.

Pre-Acceptance raw record 22103bytes SHA256
3c930416ebadb29416bc838509e18d84d3ccd342553250094f1035b584bd7dc6.
Status-excluded full stream through E006 21937bytes SHA256
9bd18bb0514c93f7593b49dc6670459f8336a726689a398900abbd2d0e76e2dd.
Post-decision integrity обязан воспроизвести этот prefix; E007 не self-attests
свои bytes. Statusbody исключён по task-record-v1; только его closure content
и append-only envelope меняются, reviewed subject остаётся тем же.

Следующий candidate Not Activated: scoped provider evidence research либо
отдельное policy/architecture decision с последующим согласованным normative
этапом. Новая TASK/branch не назначается; implementation/activation/migration
не начинаются. Commit и publication требуют отдельных explicit user gates.
После read-only post-decision integrity — STOP.

### E-097-008 — Post-Decision Integrity и Final STOP

2026-10-18. Root read-only post-Acceptance audit chunk9e568f exit0:
manifest523bc07b8145e6c39be42852cf8a4b76f0bf10b4/8rows784B unchanged;
status-excluded full stream through E00621937bytes reproduced SHA256
9bd18bb0514c93f7593b49dc6670459f8336a726689a398900abbd2d0e76e2dd.
Предыдущие evidence entries не переписаны; E007 изменил excluded Statusbody
и append-only envelope. Protected489rawSHAe7afa7bbc454bd50861473977640e0317fa689e0ad6e6db7e0415cebd9ac1fe1
совпал;218links/0broken;conflicts/trailing0;index unstaged/rawunchanged.
AnchorHEAD82cf49627e54881599bb60a55f28c8e9c6b0c4db, branch097; oldheads/
cachedremote refs/main unchanged, только new097branch. Никакой Git synchronization,
stage, commit, push, PR, merge, cleanup или operational control event не выполнен.

TASK-097 Completed — Coordinator Accepted. DP-025 Draft / Planned остаётся
bounded comparison нового требования. Current C fail-closed действует;
A/B provider capabilities и permissioncarryover не превращены в guarantee.
TASK-095/096 recovery/publication и TASK-092 не затронуты.
Дальнейший intake/implementation/activation/migration не начинается.
Commit/publication — только отдельные explicit user gates. STOP.

### R-097-001 — Replacement Intake и First Content Change

2026-10-18. User bounded rework scope; original E001–E008 выше —
historical source evidence, не current replacement approvals.
Independent /root/replacement097_intake ALLOWED; independent Architecture impact
/root/replacement097_architecture CONFIRMED до Writer с exact eight paths.
Root99d735 exit0: fresh REST main ref/commit и exact-origin ls-remote наc804,
treec6c5...,parents8cc+6a. Root4d137f clean source/8rawfiles/branchfree PASS.
Branch7e69b1 exit0 создал только isolated ref отmain; original097e460 и prior
refs/main/cachedmain сохранены. Этот task record — first content change.
Предыдущая inline Python попытка завершилась SyntaxError до исполнения;
worktree оставался clean, file не создавался; исправлена до actual first write.

Original record raw26656bytes/blob41073bd3ae8b909ccd84cf7a4213c3ca41e10070,
SHA256e0f7b2a4da82c902a99c56a2286af8d6ac5c027eab0108a30998d5a58f396986.
Original envelope18382bytes SHA256
f592c9356cda868fdcced3d7d1a30a72d99f7ea0c93c9d33571253643c957b32
losslessly preserved prefix; full source contract remains immutable in e460.
Root e3ec31 base-protected488rawaggregate
89a64009d47ef9f343267e53a04750d4d597a4501e40311c94afa632268d1cae.
Current index baseline фиксируется проверкой first write.

Immutable TASK-096 evidence locator:
https://github.com/dsdred/universal-websocket-platform/blob/82cf49627e54881599bb60a55f28c8e9c6b0c4db/docs/tasks/TASK-096-FAILED-OWNER-RECOVERY-ASSESSMENT.md
Roota74b17 exit0: GitHub contents?ref=82cf decoded50122bytes equals local Gitblob
6f71ae2a45a559e5ce6e564e76dc56ddd15de4a7, SHA256
d365b915631aa644f7b98db2ee1d516145b10f1a9a7edad57aabe6dbb495f21b;
html_url exactcommit qualified. Это forensic source, не публикация096.
Task Sources link исправлен; два DP mirror links исправит Writer. Current
anchors/contract выделены; historical original intake ниже в subject явно
помечен. Fresh roles/subject pending; old523bc не replacement Acceptance.

### R-097-002 — Transplant, Equivalence Ledger и Replacement Freeze

2026-10-18. /root/replacement097_writer восстановил семь non-task paths raw
из Sourcee460, без filters/normalization; taskrecord восстановлен root первым.
Writer13951c exit0: exact reversible substitutions, indexes raw identical,
index unstaged/unchanged,096record отсутствует. Root5d17ad exit0 независимо
подтвердил все8paths/allowed deltas, originalrefs/base/Source и oldenvprefix.
Новый subject SHA1manifest 5a583be8e6e935c1f556f4bc4bdb58b8c6c975b6,784canonicalNULbytes:
.ai/PROJECT_CONTEXT.md | full | present | 100644 | e5dbaa24920e8f3661d5e01aa68ac8b4587ed657
docs/en/design/DP-025-controlled-publisher-recovery.md | full | present | 100644 | 28390832e1d9af9f918f82ecf0c367c58ab46cbc
docs/en/design/README.md | full | present | 100644 | 235a358a50791d980fae690a738ae59b2a142d01
docs/ru/design/DP-025-controlled-publisher-recovery.md | full | present | 100644 | 401a1ac8078439feb26d0d753c0cbca9007e5016
docs/ru/design/README.md | full | present | 100644 | eb856d72aeb19b48784acd34d2e3d117ed282f94
docs/tasks/README.md | full | present | 100644 | ae58374b5816488752131f9b582f73da9bf47c86
docs/tasks/TASK-097-CONTROLLED-PUBLISHER-RECOVERY-DESIGN.md | task-record-v1 | present | 100644 | 1760e60efb5f8de71c7468d1c28d1c907939a48a
spec/current-state.md | full | present | 100644 | 9bba8361cef46e2a108034eb0f7ea00b00571c29

File-by-file equivalence ledger SOURCE raw OID -> replacement raw OID:
.ai/PROJECT_CONTEXT.md | 80065770a9cc294fa3ec3f1a3cb59a1d1eee43b1 -> e5dbaa24920e8f3661d5e01aa68ac8b4587ed657 | only current097 branch/base
docs/en/design/DP-025-controlled-publisher-recovery.md | e2ebeb68d0d703a2b2e9de2c16056645d9cfbb48 -> 28390832e1d9af9f918f82ecf0c367c58ab46cbc | one immutable evidence URL
docs/en/design/README.md | 235a358a50791d980fae690a738ae59b2a142d01 -> 235a358a50791d980fae690a738ae59b2a142d01 | raw identical
docs/ru/design/DP-025-controlled-publisher-recovery.md | 4eb73a022fdffbaece6ec3d70c1636aa04f78450 -> 401a1ac8078439feb26d0d753c0cbca9007e5016 | one immutable evidence URL
docs/ru/design/README.md | eb856d72aeb19b48784acd34d2e3d117ed282f94 -> eb856d72aeb19b48784acd34d2e3d117ed282f94 | raw identical
docs/tasks/README.md | fb7d292c458c620fd684839072110fbcf04e3441 -> ae58374b5816488752131f9b582f73da9bf47c86 | only current097 branch/base
docs/tasks/TASK-097-CONTROLLED-PUBLISHER-RECOVERY-DESIGN.md | 41073bd3ae8b909ccd84cf7a4213c3ca41e10070 -> 04744a370b9f35b980cdc1aac7277bdec1813c8f (PRE-LEDGER raw snapshot; stable projection1760e60efb5f8de71c7468d1c28d1c907939a48a) | current rework contract/anchors + historical intake heading + evidence URL; status/envelope metadata
spec/current-state.md | af81b7b1222a26d298893663e9cc577e16ef4eb9 -> 9bba8361cef46e2a108034eb0f7ea00b00571c29 | only current097 branch/base

Task raw snapshot04744... — actual bytes перед добавлением этого R002,
не self-hash final record; source raw41073... сохраняется immutable in e460.
Final full raw OID будет independently reported вне собственных bytes после
closure/при новом Commit Gate. Manifest task row всегда task-record-v1.
Task projected exact deltas: current Rework Contract, explicit historical
original Intake heading и immutable Sources link; Status и append-only Rmetadata
excluded. Original E001–E008 raw18382bytes prefix SHAf592... сохранён.
DP URL-only и nav two-token-only reversibility подтверждены root/Writer;
design A/B/C/I01/statuses и two Runtimeindexes unchanged fromSOURCE.

Root current checks:215relative links/0broken;3external evidence links exact
qualified immutableURL;conflicts/trailing0;protected488rawSHA256
89a64009d47ef9f343267e53a04750d4d597a4501e40311c94afa632268d1cae.
BaseHEADc8047fd08caa79480c2bc73b5c382f4aa5079db7 excludes096ancestor;
record096 notpresent. Original097ref=e460; alloldrefs/maincached8cc unchanged,
newisolatedref=baseM only; rawindexSHA
f2a6e27377563d9cac6793955d8a6433b37597a993ac8f0f40dce530182afcde.
No stage/commit/pub or operationalownership effects.

PROCESS-002 replacement applicability: same8paths Required; old currentanchor
tokens corrected, historicalsources unchanged; mirrors Draft/Planned,Coperative.
MASTER_PLAN/specdecisions/ADR/ARCH/CHANGELOG N/A unchanged boundedreference
rework, no durable architecture activation. ScopeAudit8Required/0Questionable/
0Removable: exactsame delivered DPmirror/index/task/nav set; no otherpath.
Fresh final Architecture, independent Tester и non-author Review pending;
old523bc approvals remain historical only. New Acceptance not issued.

### R-097-003 — Fresh Architecture, Independent Verification и Scope Audit

2026-10-18. Fresh independent /root/replacement097_architecture:
Final Architecture Confirmation CONFIRMED наmanifest
5a583be8e6e935c1f556f4bc4bdb58b8c6c975b6/8rows784B; rawreproductionee8897exit0.
0blocking/0actionable; exact DP URL-only, indexes unchanged, navtoken-only,
task currentcontract/historicalanchors и lossless oldenv подтверждены.
A/B/C/I01/Draft/Planned/Coperative unchanged. Не переносится oldapproval;
новое подтверждение bounded replacement impact, не operational authority.

Fresh independent Verifier/Tester /root/replacement097_intake PASS на тот же
manifest, taskprojection1760e60efb5f8de71c7468d1c28d1c907939a48a.
Actual commands: Python через PowerShell stdin, Git --no-optional-locks show/
ls-tree/status/diff и hash-object --stdin без -w. Chunk7d64de exit0 —
independent SHA1 blob/canonical manifest versus Git no-write crosscheck.
Chunk31bb40 exit0 — fresh full summary head/manifest/protected/index/links/
oldenv/scope. Native command/result locators, independently read by root API:
exec-ccf54267-ae2f-48f0-83bb-2c5d9a2317d4 exit0 protected/rawindex/refs;
exec-fd41a9d8-d6dc-4a9b-b260-d9a8c1d782d0 exit0 manifest/equivalence/links;
exec-20a7fe13-346a-4079-b0f3-9dae3b4b9a49 exit0 mirrors/diffcheck;
exec-9fe9ea2a-b9bb-42cb-9657-1cc3545648e3 exit0 rawGit no-write crosscheck.
Verifier earlier rawCRLF-vs-GitLF helper assertion corrected; applicable
rawcapture/index-tree checks repeated PASS, no content finding or mutation.

Reproduction: use exact PATHS/rows R002; task-record-v1 per PROCESS001, other
files full raw; path/projection/present/100644/OID NUL stream with trailingNUL.
Independent recomputation yields5a583.../784B. Equivalence: compare each current
non-task file to git show e460:path after only specified reversible replacement;
for task inspect projected delta and original18382-byte envelope prefix hashf592...
Protected recipe: UTF8-byte-sorted trackedpaths outside8scope; concatenate
path,NUL,SHA256(raw),NUL; expected488rawaggregate
89a64009d47ef9f343267e53a04750d4d597a4501e40311c94afa632268d1cae.
IndexSHAf2a6e27377563d9cac6793955d8a6433b37597a993ac8f0f40dce530182afcde
unchanged/0staged; originalrefs preserved, HEADqualifiedmainc804; no096ancestor/
record.215relative links/0broken,3exact immutable82URLs qualified root a74b17;
conflicts/trailing0, git -c core.whitespace=cr-at-eol diff --check exit0.
12mirrorsections/30traces/8invariants/10evidence/103tablerows identical structure;
full semantic inspection confirms URL-only DP delta. Fresh T01–T30 manual
designtraces PASS30/30 with unchanged safety outcomes. Runtime/live tests N/A:
documentation-only rework, not provider behavior or recovery certification.

Fresh ScopeAudit8Required/0Questionable/0Removable; same8Required paths and
PROCESS002applicability. Source-mode/raw equivalence ledger R002 valid; old
approvals historical only. Future exact replacement commit must have soleparent
qualifiedmain and onlyaccepted8paths, but commit/futureliveReadiness notcertified.
Rootfresh b6a010exit0: remote main stillc804, isolatedremoteRefabsent,matchingPR0.
No TASK095096recovery/permission or DP025activation. Non-author finalReview next.

### R-097-004 — Fresh Independent Non-Author Review

2026-10-18. Distinct /root/replacement097_review: APPROVED,
0blocking/0nonblocking findings, exactmanifest
5a583be8e6e935c1f556f4bc4bdb58b8c6c975b6/8rows784B,
taskprojection1760e60efb5f8de71c7468d1c28d1c907939a48a.
Independent0fd764exit0 raw/equivalence/refs/no096/noStage; a9ec5aexit0
protected488SHA89a640.../indexf2a6.../diffcheck;7e5b37exit0 links215/0broken,
3immutable82URLs,12sections/30T/8I/10E mirrorparity.
e332ffexit0 полностью прочитан durable fresh R003 Architect/Tester/Scope;
f7f1f5exit0 post-R003 taskprojection unchanged/no staged paths.
Предыдущие ошибочные read-only helper rawCRLF/GitLF и encoding assumptions
исправлены до final completed checks; content findings не обнаружены.

ScopeAudit8Required/0Questionable/0Removable. Explicit removability No для
каждого: taskrecord contract/provenance/gates; ENDP deliverable; RUDP mirror;
twoindexes discovery; three navcontext/currentauthoring/base synchronization.
Twoindexes source-byte-identical; DP oneURL each; nav currentbranch/base only;
task projected contract/historyheading/link bounded. Original18382B envelope
SHAf592c9356cda868fdcced3d7d1a30a72d99f7ea0c93c9d33571253643c957b32
сохранён losslessly. Current/historical approvals, PROCESS002 applicability
и prospectiveCommit/livePublicationReadiness separation согласованы.

Draft/Planned/Coperative, A/Bconditional/providerUnknown/provenance unchanged.
Никаких normative activation, owner events или old095096operational effects.
Reviewer не менял files/operationalstate. Coordinator может выполнить fresh
Acceptance/post-decision integrity на current subject; commit не разрешён.

### R-097-005 — New Coordinator Acceptance (Replacement)

2026-10-18. Coordinator /root явно принимает current bounded replacement/rework:
**Coordinator Acceptance: ACCEPTED; TASK-097 Completed — replacement accepted**.
Exact SHA1manifest5a583be8e6e935c1f556f4bc4bdb58b8c6c975b6/8rows784B;
ordered rows R002 unchanged,taskprojection1760e60efb5f8de71c7468d1c28d1c907939a48a.
Current branch docs/task-097-controlled-publisher-recovery-isolated;
anchor/prospective parent c8047fd08caa79480c2bc73b5c382f4aa5079db7,
qualified main treec6c5e6a78140cf66e42a35c474504bd8e55fb39c.

Current six ReworkDoD выполнены: fresh baseline/provenance и separatebranch/
firstcontentrecord/source preservation; ровно8transplanted paths, onlyapproved
deltas; three qualifiedimmutable evidence links, no brokenrelative links;
file-by-file raw ledger/source comparison и lossless original18382B envelope;
freshArchitectureCONFIRMED/TesterPASS30of30designtraces/ReviewAPPROVED0of0,
PROCESS002SYNCHRONIZED/ScopeAudit8Required0Questionable0Removable; isolated
prospectiveparentM excludes096ancestor and no096record, actualdiffscope8.
No old approval carried over: original523bc/e460 remains historical source;
это новая explicit Acceptance нового subject и changed baseline/branch tuple.
Root117eb7exit0 после durableR004 повторил exactsubject/protected/index/refs,
source preservation/allowed deltas/links; fresh main/ref/PR read b6a010 qualified.

Original branch docs/task-097-controlled-publisher-recovery-design=e460 retained.
No rebase/reset/cherry-pick, source rewrite, extra record or normative/tooling
change. DP-025 Draft/Planned, C operative, A/B capabilitiesUnknown remain.
TASK-095/096 operational permissions/ownership/effects не решаются; TASK-092
не затронута. Это authoring Acceptance, не Publisher event/terminal closure.

Pre-Acceptance current raw record 41544bytes SHA256
60e5b9625426aadbe111e9de1e63df6231321ac210f2fa519f9ad1b378b0d4a7.
Current full status-excluded stream through R004 41420bytes SHA256
28330a68bc5edcaf2fc237cf97d00e0f56d197cb32845d84a9d84bf2dc247584.
Post-decision integrity воспроизводит этот prefix и oldenv18382 independently.
R005 не self-hashes finalbytes; excluded Statusbody reconciled только с новой
authoring closure, appendmetadata не меняет reviewed projected subject.

Readiness: prepared для отдельного Commit Gate нового одного replacement
commit с exact parentM и только8accepted paths. Actual commitSHA/tree/range и
live PublicationReadiness будут проверены после отдельного commit permission.
Сейчас no stage/commit/push/PR/merge/publication/recovery. Старый commit gate
не используется второй раз. Следующий gate только отдельное user commit
authorization; дальнейший implementation/migration/activation Not Activated.
После post-decision integrity — STOP.

### R-097-006 — Replacement Post-Decision Integrity и STOP

2026-10-18. Root684a20exit0 fresh post-R005 audit PASS:
manifest5a583be8e6e935c1f556f4bc4bdb58b8c6c975b6/8rows784B unchanged;
current through-R004 full status-excluded41420byte prefix SHA256
28330a68bc5edcaf2fc237cf97d00e0f56d197cb32845d84a9d84bf2dc247584 reproduced.
Original E001–E008 envelope18382B/f592... still raw exact prefix.
R005 only reconciled excluded current Status and appended Acceptance;
reviewed projection1760e60e... and seven full content blobs unchanged.

Protected488rawSHA89a64009d47ef9f343267e53a04750d4d597a4501e40311c94afa632268d1cae,
rawindexf2a6e27377563d9cac6793955d8a6433b37597a993ac8f0f40dce530182afcde,
originalSource/e460branch и все priorrefs/maincached preserved. CurrentHEADc804,
isolatedbranchonly; 215relative links/0broken;exact8scope;no096ancestor/record.
Новый replacement subject Coordinator Accepted; independentfreshrolespass,
old approvals remain historical. Equivalence/diff ledger R002 плюс final
external raw readback, без self-hash финального record.

Prepared для отдельного Commit Gate; не stage/commit/publication ready event.
ReplacementcommitSHA отсутствует; prospective actualtree будет read-only
calculated/reported вне своих bytes, затем verified actual stagedtree при
отдельном permission. Требуется один новый commit exact8paths onparentc804,
затем fresh PublicationReadiness/userpublicationgate по действующимконтрактам.
DP025 Draft/Planned, Coperative; noactivation/recovery/ownershiptransition.
TASK095096operationalstate/TASK092 untouched. Никакой дальнейший scope,
cleanup, provider setup или publication не начинается. STOP.

### R-097-007 — Explicit Replacement Coordinator Acceptance Gate

2026-10-18. По новому explicit user request Coordinator /root повторно проверил
replacement subject и выдаёт **Coordinator Acceptance: ACCEPTED** для exact
manifest5a583be8e6e935c1f556f4bc4bdb58b8c6c975b6/8orderedrows784NULbytes,
taskprojection1760e60efb5f8de71c7468d1c28d1c907939a48a.
Branch docs/task-097-controlled-publisher-recovery-isolated;
qualified base и actual HEAD c8047fd08caa79480c2bc73b5c382f4aa5079db7.
Это новый explicit authoring acceptance checkpoint того же replacement
subject; R005/R006 остаются preserved prior checkpoints, не переписываются.

Root96c2ac exit0: current eightfiles/rawmanifest/full diff/equivalence ledger,
sourcepreservation,refs/index/488protected,oldenv/currentpostdecisionprefix —
PASS; actual rawTask доэтогоappend совпал previous final blob726c2... SHA cfa53...
Доказан только scope8:3added/5modified относительно base; никакого096record,
096commit не ancestorHEAD. Future replacementcommit ещё не создан.
Root9ef98e exit0 fresh RESTmain+exactoriginls agreec804; no stalecachedmainproof.

11prerequisites:
1. Exactbranch/base/HEAD verified.
2. Полный eightpathdiff/orderedR002rows/newmanifest recomputed unchanged.
3. OriginalE001–E00818382rawbytes SHA256
f592c9356cda868fdcced3d7d1a30a72d99f7ea0c93c9d33571253643c957b32 preserved.
4. DPoneURL each, navcurrentbranch/base only, indexesbyteidentical; task bounded
currentreworkcontract/historyheading/link/status/envelope deltas approved.
5. Sourcee460/oldbranch/tree/manifest/provenance retained; immutable82URL
availability/raw6f71blob50122bytes SHAd365 qualifieda74b17; historicalsource only.
6. Fresh Architecture CONFIRMED actual item
msg_0aa4f864029acf3d016ad3ddb34cd88191919b1167b43aa537,
thread01a14b95-ff6c-7c33-b6ab-3c8a62acdb43,turn01a14b99-5cde-7ed1-8227-9c73f0ba6450.
7. Independent Tester PASS actual item
msg_0426726c8b97358e016ad3de69fa70819199a8004069e7a68f,
thread01a14b95-43a4-7382-9edc-4eaebeae79f4,turn01a14b97-fa77-70c3-b265-c4d5f8cca80a.
8. Non-author Review APPROVED actual item
msg_057578e530853b54016ad3dea07c28819185e749e62be9a634,
thread01a14b99-9104-7840-9df9-fd310cbd7a70,turn01a14b99-92f5-7a13-b580-e566d6a05dfa.
Root read_thread independently reread completed actual role messages thisgate;
eachverdict cites5a583... exactreplacement,0blocking,not original523bc approval.
9. ScopeAudit8Required/0Questionable/0Removable; PROCESS002SYNCHRONIZED.
10. No096commit/record in publication scope; no095096operationalauthority.
11. ExistingpostdecisionR006 prefix reproduced; newappend integrity follows.

Acceptedpaths exactly R002 eightorderedrows, task-record-v1 onlyTASK097;
other7full raw. 215relativelinks0broken,3qualifiedimmutableURLs;30designtraces,
mirror12sections/8I/10E. Providerabilities notcertified; Draft/Planned/Coperative.
Commit, push, publication/release/recovery/ownershiptransition не разрешены
этой Acceptance. TASK092 untouched. No phase effects performed.
Pre-this-gate rawrecord46590bytes SHA256
cfa53e72cc1886f4ad52dc5f3b8100ad83abaa2f9792d18c1d03c7091021d29a;
full status-excludedthroughR00646291bytes SHA256
8cbede14255da857e4be19e2b3b5b9e722bbf9a7e6394b3fdb99c77d7d64600e.
Postreadback must reproduce thisprefix; thisentrydoesnotselfhash.
No change to reviewed projectedsubject; only append-only current evidence.
STOP after required post-decision readback.

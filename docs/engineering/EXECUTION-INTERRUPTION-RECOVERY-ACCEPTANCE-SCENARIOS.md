# Execution Interruption Recovery Acceptance Scenarios

## Purpose

Сценарии проверяют общий Execution Interruption Recovery contract
[PROCESS-001](PROCESS-001-AI-DEVELOPMENT-WORKFLOW.md). Publisher-specific
P0–P10 cases остаются в
[Publisher Acceptance Scenarios](PUBLISHER-ACCEPTANCE-SCENARIOS.md) и не
дублируются либо не ослабляются здесь.

Во всех сценариях repository и independently reproducible evidence имеют
приоритет над памятью session. Recovery сначала read-only и не создаёт verdict,
status, checkpoint или permission.

## R-001 — Interruption не является результатом

Given agent был прерван model/time limit, network loss, session/process crash,
host restart/reboot либо tool timeout/failure, when новый агент reconstruct-ит
pipeline, then interruption не классифицируется как `PASS`, `FAIL`, `Approved`,
`Accepted`, `Completed` или checkpoint. Started stage остаётся первым
незавершённым, пока completion не доказан.

## R-002 — Новый агент без session history

Given новый агент не имеет предыдущего chat/tool context, when task record
содержит Task ID/status, exact branch/baseline, scope, roles, ordered stages и
content identity, then агент независимо сверяет Git/diff/evidence и продолжает
первый checkpoint без proven completion. Chat history не требуется и не
используется как recovery state.

Given exact task уже `In Progress` и current user input явно просит
продолжить/resume её, then исходная task-cycle authority сохраняется до STOP в
неизменном scope. Без current resume/continue input task record сам execution
не запускает.

## R-003 — Interruption до side effect

Given precondition inspect доказывает, что mutation и её postcondition
отсутствуют, when recovery классифицирует operation как `Proven Not Started`,
then operation может начаться только при действующих scope/prerequisites и
применимой task-cycle/operation permission. Отсутствие side effect не
доказывает, что gated permission уже было.

## R-004 — Interruption после side effect

Given exact postcondition и mandatory completion evidence доказаны, when
recovery классифицирует operation как `Proven Completed`, then operation не
повторяется; pipeline продолжает следующий checkpoint либо отправляет
пропущенный user report.

## R-005 — Неизвестный момент относительно side effect

Given pre/postcondition не позволяют доказать момент interruption, when
outcome `Unknown`, then blind retry запрещён. Агент выполняет
operation-specific reconciliation; если факт всё ещё неоднозначен, stage
останавливается без relabel в Failed.

## R-006 — Local side effect, remote отсутствует

Given exact local commit существует, но remote ref не содержит его OID, then
commit checkpoint может быть completed, а push остаётся `Proven Not Started`
или `Unknown` по remote evidence. Commit не повторяется, publication без gate
не начинается.

## R-007 — Remote side effect, local state устарело

Given exact remote ref/PR/merge outcome доказан, но local refs отстают, then
remote operation не replay-ится. Агент продолжает phase-appropriate
synchronization; stale local state не превращает completed remote mutation в
failed operation.

## R-008 — Interruption до permission gate

Given repository доказывает readiness, но current user input не содержит
required permission, then Commit/Publisher gate остаётся непройденным. Task
status, target или существующая ветка permission не создают.

## R-009 — Permission было дано, operation доказанно не выполнялась

Given one-shot commit permission находилось только в потерянной session и
reconstruction доказывает отсутствие commit, then новый агент запрашивает
`Разрешаю коммит.` повторно для того же неизменного accepted tuple.

Given immutable Publisher Target reconstructable и current user input явно
ссылается на ранее разрешённую publication, then Publisher применяет свой
resume contract без новой publish permission. Без explicit resume reference
обычный publish gate требуется снова.

## R-010 — Permission было дано, outcome operation неизвестен

Given permission известно либо заново подтверждено, но outcome stage/commit/
remote mutation неизвестен, then reconciliation выполняется до любого retry.
Permission не превращает unknown outcome в safe retry.

## R-011 — Operation выполнена, report не отправлен

Given exact operation outcome доказан, но terminal/user report отсутствует,
then агент не повторяет operation и отправляет truthful report либо продолжает
оставшиеся обязательные checkpoints. Для Publisher merge-only report всё ещё
не является P10.

## R-012 — Partial file mutation

Given interruption произошёл во время edit/apply/generation, then агент
сравнивает actual bytes и полный diff с expected pre/postcondition. Он
продолжает только отсутствующую часть и не overwrites unrelated/attributed
changes повторным полным patch.

## R-013 — Partial или ambiguous stage

Given index содержит часть accepted/certified file set либо unexpected path,
then stage не является commit checkpoint. Coordinator inspect-ит index и
worktree; unexpected/ambiguous state останавливает gate, exact partial state
reconciled только в пределах действующего tuple.

## R-014 — Ambiguous commit

Given commit command/tool response потерян, then Coordinator inspect-ит HEAD,
parents, tree, message, log/reflog и accepted/certified subject-manifest identity.
Exact existing commit считается completed; duplicate commit запрещён. Если
commit доказанно отсутствует, применяется permission rule R-009.

## R-015 — Implementation interruption

Given Developer handoff отсутствует либо не связан с current content identity,
then Implementation не completed. Фактический diff сохраняется, scope
reconstructed, и работа продолжается от первого missing change; следующий
stage не получает ложный handoff.

## R-016 — Verification Matrix interruption

Given check начал выполняться, но exact exit/result не сохранён, then он не
PASS и не FAIL. На unchanged content его выполняют заново после reconciliation
возможных material artifacts; completed unaffected checks с reproducible
evidence не повторяются без необходимости.

## R-017 — Tester interruption

Given Tester был прерван до explicit report, then verdict отсутствует. Given
report существует, но content identity изменился, then verdict stale и
затронутые checks повторяются.

## R-018 — Independent Review interruption

Given Reviewer был прерван во время анализа, then Approval отсутствует даже
при partial notes. Review completed только explicit verdict/findings для exact
file set/subject-manifest identity.

Given verdict/evidence envelope хранит subject-manifest OID, then envelope не
self-attest-ит свои bytes. Task record projection использует exact
`task-record-v1` raw-byte algorithm: status body и terminal envelope исключены,
фиксированный NUL marker не зависит от newline style, а missing/duplicate/
out-of-order headings дают `Inconsistent`. До commit недоказанная envelope
mutation требует Repeat Independent Review; после commit final bytes доказывает
tree OID.

## R-019 — Rework interruption и repeat review

Given rework частично изменил diff, then affected Verification, Scope Audit,
Review и Acceptance invalidated. После reconciliation rework завершается в
исходном scope, затем выполняются repeat Verification и Repeat Independent
Review; старый Approved не переносится.

## R-020 — Между Review и Coordinator Acceptance

Given exact Independent Review completed, но explicit Coordinator Acceptance
не записана, then Acceptance непройдена. Recovery не выводит её из Approved и
не переходит к Commit Gate.

## R-021 — После Coordinator Acceptance

Given explicit Acceptance tuple существует, when current subject/subject-manifest
identity совпадает, then Acceptance proven completed и не повторяется без
необходимости. Любой content change invalidates affected Acceptance evidence и
возвращает pipeline в applicable verification/review.

## R-022 — До, во время и после Commit Gate

Given interruption до gate, permission отсутствует. Given interruption во
время stage/commit, index и commit reconciled по R-013/R-014. Given exact
commit proven completed, gate не replay-ится и task не получает второй commit;
publication остаётся отдельным permission.

## R-023 — Push, PR и merge unknown outcomes

Given push/create/merge response потерян из-за network/GitHub/auth/tool
failure, then exact remote ref, exact head/base PR и PR `MERGED`/merge OID
inspect-ятся до retry. Matching outcome завершает checkpoint; ambiguous state
останавливает его. Publisher scenarios S-004, S-011 и S-012 остаются
обязательными.

## R-024 — Branch deletion unknown outcome

Given delete response потерян, then exact local/remote ref existence и OID
inspect-ятся. Already absent может доказать completion; moved/recreated ref не
удаляется. Merge не откатывается и не повторяется.

## R-025 — Documentation/status transition interruption

Given task/status/index/mirror mutation прервалась, then actual bytes,
source-precedence и required evidence сверяются. Partial transition не
повышает status; ложный `Completed` возвращается в documentation rework и
Independent Review.

## R-026 — Publisher P0–P10 interruption

Given interruption в любом P0–P10, then общий Recovery Reconstruction Gate
передаёт управление phase-aware Publisher Resume Reconstruction Guard. Guard
сохраняет immutable Target, P6 phase boundary, inspect-first ambiguous
mutations, cleanup и P10; общий contract не replay-ит P-step и не ослабляет
S-001–S-040 Publisher scenarios.

## R-027 — Permission target изменился

Given diff, branch, commit/head, base, scope, publication class либо certified
tuple изменились после permission, then старое permission invalidated. Retry,
reconciliation или existing partial side effect не переносит authority на
новый target.

## R-028 — Canonical path order не зависит от platform collation

Given subject содержит paths, для которых case-insensitive/locale collation и
unsigned UTF-8 byte ordering дают разную последовательность (например,
uppercase filename и lowercase directory в одном parent), when canonical
subject manifest строится на разных hosts/tools, then записи сортируются по
ascending unsigned UTF-8 path bytes, case-sensitive и locale-independent.
Одинаковые repository bytes дают один manifest OID; platform-default sort не
является допустимым evidence.

## R-029 — Blocked closure manifest не self-referential

Given blocked-closure evidence set зафиксирован Coordinator, when canonical
subject manifest строится, then task record входит с projection
`task-record-v1`, каждый другой present path — с `full`, deleted path — с
baseline mode и OID `-`, а rows имеют exact NUL-separated форму
`path\0projection\0state\0mode\0oid\0` и unsigned UTF-8 path order. Terminal
Recovery Evidence Envelope является metadata и исключается из task-record
projection. Append-only запись tuple в envelope не меняет manifest OID;
mutation projected subject вне envelope invalidates downstream gates. Изменение
исключённого status evidence body не меняет manifest OID, но требует
status/contract reconciliation. Raw или normalized diff digest не является
допустимой заменой.

## R-030 — Blocked closure Tester handoff durable

Given Tester выдал verdict для blocked-closure subject, then repository
handoff содержит exact tested subject/manifest identity, ordered path set,
commands с exit/results, limitations, scope/coverage counts и
reproducible proof artifacts. Reviewer проверяет эти сведения без chat
history; отсутствующий, неполный или stale identity не даёт certification и
требует повторной Verification/Review.

## R-031 — Publisher execution context не наследует user capability

Given normal interactive identity проходит GitHub API и Git remote probes, но
agent sandbox identity не проходит их, then recovery не выводит capability из
общего `USERPROFILE`, account/helper metadata или user vault. Exact sandbox
context остаётся blocked до successful dual proof либо trusted-context handoff.

## R-032 — Interruption во время Release/Accept Handoff

Given interruption произошёл после Release Handoff, но до доказанного Accept
Handoff, then source остаётся observation-only, destination не считается
owner, side effects запрещены. Recovery reconstruct-ит exact handoff/Target и
его unique non-secret transfer ID, повторяет только недоказанные read-only inspections/probes; `Started !=
Completed` применяется к acceptance ownership.

## R-033 — Accepted destination и duplicate source resume

Given exact Accept Handoff доказан, then destination является единственным
procedural owner remaining P0-P10. Source или второй destination, пытающийся
resume, останавливается до mutation; отсутствие machine lock не разрешает
concurrent ownership. Accept обязан ссылаться на exact unique transfer ID и
Target; ID сам не является secret, permission или lock.

## R-034 — Mismatched handoff или reconciled remote effect

Given factual immutable Target mismatch, then publication authorization
становится `InvalidatedByTargetChange`, ownership `NoneTerminal`, связанные
attempts закрываются `Closed(TargetChanged)`. Given Target unchanged, но
acceptance использует unknown/reused/mismatched/duplicate/already-accepted
transfer ID либо два destinations, then отклоняется только эта acceptance
attempt: proven exact owner сохраняется, а ambiguous/conflicting ownership даёт
`Unknown` и STOP. Given Target unchanged и remote inspect доказал push/PR/merge
после previously unknown outcome, then completed checkpoint не replay-ится;
destination принимает ownership только через exact valid Accept и продолжает
первый незавершённый step после dual capability proof.

## R-035 — Operational handoff record утрачен или partial

Given interruption произошёл в любом handoff state, recovery читает append-only
operational record вне immutable Target/project-state и сверяет UUIDv4 transfer
ID, Target, source/destination identities, Release snapshot и event chain.
Каждый event обязан цитировать predecessor ID/digest/tail и resulting
authorization, ownership, attempt/reason. Release без Accept означает
`Active/InTransitNone/Released`; exact route/Accept —
`Active/Owned(destination)/Accepted`; closed event обязан доказать reason и
authorization/owner disposition. Partial/conflicting chain либо более одного
state/owner даёт ownership `Unknown`, STOP; started output не является durable
transition.

## R-036 — Recovery terminal/return disposition

Given interruption после cancellation request, reverse Release, Target
mismatch, user revoke или P10, recovery не применяет generic terminal state. Он
доказывает exact event: valid `CancelledBeforeAccept` возвращает recorded
releasing source при `Active`; reverse Release оставляет `InTransitNone` до
fresh Accept; mismatch даёт `InvalidatedByTargetChange/NoneTerminal`; revoke —
`RevokedByUser/NoneTerminal`. P10 transition допустим только если predecessor
доказывает `Active/Owned(execution-context)` у exact actor; `Released/
InTransitNone` запрещает P10 до exact Accept либо valid
`CancelledBeforeAccept`, вернувшего owner. Proven P10 даёт
`Consumed(P10)/NoneTerminal` и report-only; no-handoff сохраняет `Unissued` с
publication-level no-ID event, current `Accepted` закрывает exact ID, а already
`Closed(reason)` сохраняет reason с отдельным P10 event только при доказанном
current owner. Missing predecessor, owner, reason/disposition либо fabricated
transfer ID даёт ownership `Unknown` и STOP.

## Coverage Matrix

### New-record bootstrap decision scenarios

R-037–R-048 trace to the numbered clauses of Attributed New-Record Bootstrap
Recovery in PROCESS-001 and PROCESS-002 Durable Blocked-Closure Evidence.
These are process proof/regression scenarios, not product implementation tests.

| ID | Given / When | Required result | Trace |
|---|---|---|---|
| R-037 | One active task; trusted branch/base; original-scope record absent from baseline; independent exact-byte observation before blocker; complete capture and all fresh checks/status reconciliation succeed | Coordinator may certify Blocked closure without stage; no product Acceptance or new task | 1–3, 6–7 |
| R-038 | Extra/unowned, product, temporary, generated or unlisted untracked path | Eligibility fails; naming it evidence, hiding/deleting it or stage/intent-to-add cannot bypass intake | 1, 4, 7 |
| R-039 | Current ownership and snapshot bytes match, but chronology has only branch reflog, timestamps and record/user assertions | Chronology Not Proven; certification STOP; new capture/review cannot create historical event evidence | 2–3, 6 |
| R-040 | Record/anchor was created after blocker discovery | Not eligible; backdating or redefining blocker as bootstrap start rejected | 1–2, 4 |
| R-041 | Lossless original capture has encoding, size and raw/projection/manifest identities; authorized current contract edit occurs | Reconstruct all original identities; compute distinct current subject; historical snapshot is not replacement acceptance or chronology proof | 3–4, 6 |
| R-042 | Snapshot loses/changes a byte or LF/CRLF is normalized before hashing | Preservation/identity fails; certification STOP despite similar rendered text | 3, 6, 8; canonical identity |
| R-043 | No explicit current bootstrap authorization, or proposal adds product work, published-subject acceptance, task-specific exception or expanded commit/publication authority | Process mutation STOP; precedent/task-local contract grants no exemption | 4–5, 7 |
| R-044 | General amendment and preservation pass, chronology remains Not Proven | Separate process result; PROCESS-002 Blocked; no Acceptance/certification or automatic task status/intake transition | 5–6; PROCESS-002 |
| R-045 | Mandatory verification fails or Reviewer rejects current subject | Not Certified; bounded rework repeats affected checks; old PASS/Approved cannot transfer | 6, 8 |
| R-046 | Capture/edit started but outcome unknown after interruption | Inspect actual original/current bytes and inventory; missing capture blocks dependent mutation; partial edit reconciled without blind replay | 3, 8 |
| R-047 | After review/certification projected current subject changed, or excluded status/envelope integrity is unproven | Invalidate affected gates for projected change; separately reconcile excluded metadata; no inferred certification | 6, 8 |
| R-048 | Certified evidence or checkpoint exists, but terminal publication/clean-main proof absent | No prerequisite activation; separate gates, full P0–P10 and exact sealed admission remain required | 7; ordinary intake |

| Required area | Scenarios |
|---|---|
| Model/time, network, session/process crash, host restart, tool timeout/failure, GitHub outage/auth failure | R-001, R-005, R-016, R-023, R-026 |
| Implementation, Verification Matrix, Tester, Review, Rework | R-015–R-019 |
| Review-to-Acceptance, post-Acceptance, Commit Gate | R-020–R-022 |
| Publisher P0–P10 and unknown remote outcome | R-006, R-007, R-023, R-024, R-026 |
| File, stage, commit, push, PR, merge, branch delete, documentation/status | R-012–R-014, R-023–R-025 |
| Before/after/unknown side effect; local/remote split | R-003–R-007 |
| New agent without history | R-002 |
| Permission before/after, granted-not-run, run-not-reported | R-008–R-011, R-027 |
| Cross-platform canonical identity | R-028 |
| Blocked closure canonical subject and durable Tester evidence | R-029–R-030 |
| Exact-context capability and user/sandbox identity separation | R-031 |
| Trusted-context handoff interruption, ownership, return and terminal disposition | R-032–R-036 |
| New-record eligibility, chronology, preservation, scope and certification | R-037–R-045 |
| Interrupted bootstrap and unchanged sealed-intake gates | R-046–R-048 |

## Acceptance

### Negative disposition decision scenarios

R-049–R-073 exercise the distinct general ND-1–ND-5 PROCESS-001 class C.
Each row is a reproducible governance decision trace, not a runtime test or
evidence that publication was executed. Earlier A/B and interruption scenarios
remain mandatory in their original scope.

| ID | Given / When | Required result | Trace |
|---|---|---|---|
| R-049 | Normal Accepted task can use A | C reject; use ordinary A gate | ND-1.1 |
| R-050 | Task is eligible for BCC | C reject; use B, not a negative substitute | ND-1.1 |
| R-051 | Not Proven; known feasible recovery route remains | C reject/STOP to recover evidence | ND-1.3 |
| R-052 | Not Proven; bounded required inventory exhausted, all other eligibility proven | C candidate only; general Approval does not decide application | ND-1, ND-2 |
| R-053 | Scoped chronology Disproven, ownership/capture proven, recovery and other gates pass | Explicit Disproven negative candidate; never relabel Not Proven or positive proof | ND-1.2 |
| R-054 | Ownership not proven or another task owns record | C reject, not a provenance waiver | ND-1.1, ND-1.4 |
| R-055 | Unexplained production/test/module diff | C reject; no hiding failed implementation | ND-1.5 |
| R-056 | Reviewer has unresolved blocking finding on disposition subject | Decision and commit reject | ND-1.6, ND-2 |
| R-057 | All reviews pass but Coordinator decision absent | Commit Gate reject; Approved is not disposition | ND-2–ND-3 |
| R-058 | Valid decision, user commit permission absent | Stage/commit reject | ND-3 |
| R-059 | Exact projected subject/tuple changes after decision | Affected eligibility/decision/permission invalid; repeat affected gates | ND-2, ND-5 |
| R-060 | Exact checkpoint exists, no publication permission | Publication reject; commit not replayed | ND-3, ND-5 |
| R-061 | Publication permission binds a different target or A/B class | C reject; no authority transfer | ND-3 |
| R-062 | Interruption in any P0-P10 | Inspect/Reconstruct/Reconcile/Resume exact first unfinished phase; no blind side effects | ND-5; Publisher Guard |
| R-063 | Exact P10 proven; merged ancestry/refs/clean synchronized main match | Sealed Negative Disposition removes only active-task barrier; next separate ordinary intake checks readiness | ND-4 |
| R-064 | Decision/commit/push/merge exists but unpublished, dirty or no P10 | Intake reject; projected In Progress never resumes original work after valid decision | ND-4 |
| R-065 | Downstream proof cites C as accepted result or repaired prerequisite | Reject; C only establishes negative disposition | ND-1, ND-4 |
| R-066 | Delete/stash/move/reset record to make baseline clean | Reject; preserve original history | ND-1.4 |
| R-067 | Ownership or capture integrity is Disproven | Reject, unlike scoped provenance Disproven | ND-1.2 |
| R-068 | Snapshot altered or LF/CRLF treated equivalent; staged bytes differ | Preservation/identity/Commit Gate reject; no normalization rule | ND-1.4, ND-2–ND-3 |
| R-069 | Interruption before decision or after decision before commit | Before: no decision; after: verify tuple/integrity, separate one-shot permission still required | ND-5 |
| R-070 | Commit response unknown; publish permission absent | Inspect exact Git tree/parent/OID before any retry; proven commit not repeated, publication STOP | ND-5 |
| R-071 | Publish permission received, interruption before P0 | Reconstruct exact C Target/authority/owner; P0 still first unfinished | ND-3, ND-5 |
| R-072 | P10 proven completed; new historical evidence then appears | Report/reconcile only; no replay/retroactive rewrite; new evidence requires separate authorized normal intake | ND-4–ND-5 |
| R-073 | New concrete provenance pointer/proof after decision before commit, or during publication before P10 | Hold first remaining mutation; independently re-evaluate ND-1; failed eligibility forbids remaining effects/P10/intake even unchanged Git target; preserve effects/authority axes, no forced cleanup; resume only proven original tuple validity | ND-5 |

Additional boundary for R-051/R-052: an inaccessible mandatory source, an
unknown inspection outcome or an uninspected concrete pointer is Not Exhausted.
No hypothetical-archive search is required, but known sources cannot be omitted.

### Immutable published-subject prospective acceptance scenarios

R-074–R-091 exercise IPSPA PROCESS-001. They create no historical equivalence,
publication permission or downstream activation.

| ID | Given / When | Required result |
|---|---|---|
| R-074 | Historical accepted manifest equals exact published source manifest | IPSPA is N/A; existing proven equivalence is not duplicated |
| R-075 | Historical equivalence is Not Proven and immutable published objects exist | New event may become Candidate; historical value remains Not Proven |
| R-076 | Historical equivalence is Disproven | New Candidate may proceed, but Disproven is preserved |
| R-077 | Normal exact source, fresh gates and decision complete | Only Prospective Event becomes Accepted for named claims |
| R-078 | Commit/tree/path/base/row/source-manifest mismatch | Reject/STOP; no approximate equivalence |
| R-079 | Authoritative object bytes or publication observation unavailable | Reject/STOP; working-tree or copied bytes cannot substitute |
| R-080 | Source tuple changes during the event | Old event invalidated; changed source requires fresh UUIDv4 event |
| R-081 | Evidence Record occurs in its own Source Subject | Reject as self-attestation |
| R-082 | Historical Tester/Reviewer/Acceptance/merge/P10 offered as fresh gate | Reject; run fresh applicable gates on exact source |
| R-083 | Fresh Verifier/Tester refers to another tree or manifest | Reject stale/mismatched handoff |
| R-084 | Independent Reviewer has unresolved finding | Coordinator decision prohibited |
| R-085 | Interruption before prospective decision | Reconstruct `S/E`; no decision exists from started work |
| R-086 | Interruption after decision before evidence commit | Require exact unchanged decision tuple/post-decision integrity and current commit permission |
| R-087 | Publication/ref observation changes but immutable object tuple is unchanged | Reconstruct observation; refs alone neither mutate source nor prove publication |
| R-088 | Repository identity, UUID, source object or Publisher Target is ambiguous | Fail closed; affected event/authority cannot transfer |
| R-089 | Checkout/filter/EOL/decoded/archive/diff bytes offered as source | Reject; object API and full-only rows are mandatory |
| R-090 | Projected Evidence Record changes versus append-only envelope update | Projected change invalidates affected gates; proven envelope append keeps projection but requires metadata reconciliation |
| R-091 | Accepted event is cited for unrelated prerequisite or automatic task activation | Reject; exact named downstream contract and separate intake are mandatory |

These rows preserve Accepted Task, Blocked Evidence Recovery and Negative
Disposition semantics and all earlier interruption/user-gate scenarios.

Contract passes only if every scenario is traceable to normative PROCESS-001,
PROCESS-002 or role text; negative assertions (`no verdict`, `no blind retry`,
`no inferred permission`, `no replay completed checkpoint`) are preserved; and
Publisher S-001–S-040 remain valid.

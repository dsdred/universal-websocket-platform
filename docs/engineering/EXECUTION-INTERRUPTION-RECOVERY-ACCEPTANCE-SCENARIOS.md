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
S-001–S-025 Publisher scenarios.

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

## Coverage Matrix

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

## Acceptance

Contract passes only if every scenario is traceable to normative PROCESS-001,
PROCESS-002 or role text; negative assertions (`no verdict`, `no blind retry`,
`no inferred permission`, `no replay completed checkpoint`) are preserved; and
Publisher S-001–S-025 remain valid.

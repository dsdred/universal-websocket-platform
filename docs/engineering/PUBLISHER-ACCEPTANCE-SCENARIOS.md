# Publisher Acceptance Scenarios

## Purpose

Эти process scenarios проверяют контракт Publisher из
[PROCESS-001](PROCESS-001-AI-DEVELOPMENT-WORKFLOW.md) и
[Publisher role](agents/publisher.md). Если scenario не проверяет отсутствие
checkpoint, admissible accepted task commit, blocked evidence checkpoint либо
negative disposition checkpoint уже создан по отдельной команде
`Разрешаю коммит.`.

Общий pre-publication и cross-pipeline interruption contract проверяется
[Execution Interruption Recovery Acceptance Scenarios](EXECUTION-INTERRUPTION-RECOVERY-ACCEPTANCE-SCENARIOS.md).
Эти scenarios сохраняют более строгие immutable Target и phase-aware P0–P10
rules.

## S-001 — Полностью успешный pipeline

Given clean exact task branch/OID и команда `Разрешаю публиковать.`, Publisher
выполняет P0–P10 без дополнительных permission prompts. Terminal report
содержит PR number/URL, task и merge OID, checks state, `MERGEABLE / CLEAN`,
явную repository-approved merge strategy, `git fetch --prune`, обе branch
deletions, `main == origin/main`, clean worktree и STOP.

## S-002 — Push является checkpoint

Given P0 прошёл и P1 подтвердил remote OID, Publisher немедленно выполняет P2:
inspect и create/discover PR. Он не выводит запрос `Создать Pull Request`, не
останавливается и не просит новое разрешение.

## S-003 — SSH preflight failure до push

Given noninteractive SSH/origin proof не прошёл, mutation отсутствует, P1
не attempted, P0 Initial Preflight является первым незавершённым шагом, а
ordered completed pipeline steps отсутствуют. Blocker report может перечислить
успешные P0 subchecks, сохраняет branch/worktree и подтверждает действующее
publish authority; resume повторяет P0.

## S-004 — `gh` auth failure после push

Given P1 подтверждён, а `gh` authentication недоступна перед P2, Publisher не
повторяет push. P1 остаётся завершённым, P2 — первым незавершённым; remote
branch/OID сохраняются.

## S-005 — Resume без повторного разрешения

Given blocker S-004 устранён и пользователь сообщает
`Авторизация готова. Продолжай ранее разрешённую публикацию.`, task-branch
phase-aware Resume Reconstruction Guard reconstruct-ит P1 из remote OID и
продолжает с P2.

Given P6 уже confirmed, current branch clean `main`, а blocker остановил P7,
P8 или P9, main-phase Resume Reconstruction Guard не требует current task
branch/HEAD target OID, не recreates её и продолжает exact cleanup step. Новая
команда `Разрешаю публиковать.` в обоих случаях не требуется.

## S-006 — CI отсутствует

Given `.github/workflows` отсутствует или зарегистрировано zero checks,
Publisher фиксирует `No CI`. При PR gate `MERGEABLE / CLEAN` P3 не блокирует
pipeline; P4 merge и P5–P10 выполняются.

## S-007 — Required checks pending

Given required checks имеют `Pending`, P3 остаётся первым незавершённым шагом.
Merge и cleanup не выполняются; PR и refs сохраняются; resume повторно
inspect-ит checks без нового publish permission.

## S-008 — Required checks failed

Given required check имеет `Failed`, P3 блокируется. Publisher не bypass-ит
check и не merge-ит PR; blocker report содержит failed check state.

## S-009 — PR перестал быть mergeable перед P4

Given P3 ранее завершён, но непосредственный P4 recheck обнаружил новый
conflict, `UNKNOWN`, non-clean gate или branch-protection refusal, P4 остаётся
первым незавершённым. PR и refs сохраняются; Publisher не выполняет
force/rebase/bypass. Conflict, обнаруженный initial P3 после push, относится к
S-015 и оставляет P3 незавершённым.

## S-010 — Cleanup failure после merge

Given PR уже подтверждён `MERGED` и merge OID сохранён, failure на P5, P6, P7,
P8 или P9 останавливает exact первый незавершённый cleanup step. Publisher не
откатывает и не повторяет merge. После confirmed P6 resume P7/P8/P9 происходит
на clean current `main`, не требует task branch/HEAD, не recreates branch и
учитывает, что `main` может отставать до P7, а equality требуется только P9.

## S-011 — Ambiguous PR creation

Given create-PR response ambiguous, Publisher inspect-ит existing PR с exact
head branch/OID и base `main` до retry. Duplicate PR не создаётся; при
невозможности доказать state P2 остаётся незавершённым.

## S-012 — Ambiguous merge response

Given merge response ambiguous, Publisher inspect-ит PR state. P4 считается
завершённым только после `MERGED` и capture merge OID; blind re-merge
запрещён.

## S-013 — Remote branch recreated or moved

Given после merge remote task ref указывает не на authorized target head OID,
Publisher не удаляет ref. P5 остаётся незавершённым, текущее состояние
сохраняется и сообщается как cleanup blocker.

## S-014 — Target invalidation

Given publication class, branch, ordered commit target/head OID, base или
accepted/certified/negative-disposition scope изменились, read-only
preflight останавливает pipeline до mutation. Это invalidation exact authority,
а не внешний blocker; старое разрешение не применяется к новому target.

## S-015 — Conflict появился после push

Given P1 завершён, но после push PR стал conflicting, P3 фиксирует conflict и
остаётся первым незавершённым checkpoint. P4 не начинается, Publisher не
выполняет rebase/force и blocker report сохраняет exact PR/head state.

## S-016 — Gate изменился перед merge

Given P3 ранее наблюдал success, непосредственно перед P4 Publisher повторно
проверяет CI, exact head/base OID и mergeability. Новый pending/failed check,
conflict или non-clean gate блокирует P4; stale P3 result не используется.

## S-017 — Repository merge strategy

Given repository policy разрешает конкретную merge strategy, Publisher явно
называет и применяет её. Если policy недоступна или разрешённая strategy
неоднозначна, P4 блокируется; Publisher не изобретает policy.

## S-018 — Pruned cleanup evidence

Given P5 удалил exact remote branch, Publisher выполняет `git fetch --prune`,
затем P6–P9. P9 подтверждает отсутствие remote-tracking и local task refs,
`main == origin/main` и clean worktree. Failure prune фиксируется как первый
незавершённый cleanup action без replay merge.

## S-019 — Certified blocked closure не является Acceptance

Given task имеет `Blocked Closure Certified` и отдельно созданный `Blocked
Evidence Checkpoint`, Publisher target использует class `Blocked Evidence
Recovery`. P0 отклоняет любой tuple, где task названа Accepted или Completed.

## S-020 — Ordered recovery-chain

Given между base OID и evidence checkpoint существует ровно один явно
authorized process-amendment commit, P0 подтверждает exact ordered range и
разрешает P1 только для его head OID. Любой дополнительный commit invalidates
target.

## S-021 — Dirty evidence после checkpoint

Given после evidence checkpoint остались staged, unstaged или untracked
changes, initial P0 сообщает safety failure без mutation. Publish authority не
используется для stash, reset, amendment или нового commit.

## S-022 — Blocked publication выполняет полный P0–P10

Given valid blocked recovery target и `Разрешаю публиковать.`, Publisher не
сокращает pipeline: push, PR, checks, merge, cleanup, synchronized `main` и
terminal report выполняются по тем же P0–P10 gates.

## S-023 — Prerequisite не активируется публикацией

Given blocked recovery merged и P9 подтвердил clean synchronized `main`, P10
сохраняет task `Blocked` и prerequisite `Not Activated`. Только отдельная
последующая `Продолжай проект.` может через узкое sealed-evidence exception
запустить обычный deterministic intake exact prerequisite. Без P10 exception
не действует.

## S-024 — Certification tuple изменился

Given evidence file set, canonical evidence digest, blocker identity или verification/review
results отличаются от certification tuple, Commit Gate либо Publisher P0
останавливается. Ранее выданное permission не применяется к изменённому target.

## S-025 — Новый Publisher без истории session

Given новый Publisher не имеет прежнего chat/tool context, он сначала
reconstruct-ит immutable Target и completed P-checkpoints из Git/GitHub. Если
current user input явно ссылается на ранее разрешённую exact publication,
Resume Reconstruction Guard продолжает первый незавершённый P-step без новой
команды `Разрешаю публиковать.`. Если explicit resume reference либо exact
Target отсутствуют, существующий ref/PR/merge не выдаёт permission: уже
completed effect только сообщается, а не начатая publication требует обычного
gate.

## S-026 (A) — Windows user capable, sandbox identity incapable

Given normal Windows identity успешно проходит GitHub API и Git remote probes,
но exact sandbox identity не имеет доступа к user credential vault и проваливает
обязательные probes, then capability sandbox не доказана, P0 `BLOCKED`, P1 не
attempted и side effects отсутствуют. Blocker классифицируется `Blocked by
Publisher Execution Environment — GitHub credentials unavailable to execution
identity`; общий `USERPROFILE` не меняет verdict.

## S-027 (B) — Trusted-context handoff после blocker

Given source выпустил non-secret Release Handoff и стал observation-only, а
handoff содержит новый unique non-secret transfer ID, пользователь явно
маршрутизировал exact ID плюс Target ранее разрешённой publication в trusted
destination, then destination выполняет `Inspect -> Reconstruct -> Reconcile`,
доказывает unchanged Target и оба capability probe, фиксирует Accept Handoff с
тем же ID/Target и
продолжает первый незавершённый step. Authorization сохраняется; новый publish
gate, Commit Gate и Coordinator Acceptance не требуются.

## S-028 (C) — Destination обнаруживает другой HEAD

Given destination при reconstruction видит HEAD, ordered range или иной
immutable Target field, отличный от Release Handoff, then authorization
становится `InvalidatedByTargetChange`, ownership `NoneTerminal`, attempts
закрываются `Closed(TargetChanged)`. Given Target unchanged, но ID unknown/
reused/duplicate/already-accepted/mismatched либо user route подменяет
destination, then отклоняется только acceptance attempt: proven Released state
либо proven exact owner сохраняется; ambiguous/conflicting ownership становится
`Unknown` и STOP. Ни один случай не является resumable credential blocker.

## S-029 (D) — Push найден после unknown outcome

Given Release Handoff классифицировал P1 outcome `Unknown`, а destination
read-only inspect обнаружил exact remote branch at authorized head OID, then P1
reconciled как completed и push не повторяется. Destination продолжает P2 только
после полного Accept Handoff.

## S-030 (E) — API PASS, Git remote FAIL

Given GitHub API probe успешен, но exact-origin Git remote authentication/read
неуспешен, then execution capability не доказана. P0 остаётся незавершённым и
side effects запрещены.

## S-031 (F) — Git remote PASS, API FAIL

Given exact-origin Git remote probe успешен, но GitHub API authentication или
repository probe неуспешен, then execution capability не доказана. P0 остаётся
незавершённым и side effects запрещены.

## S-032 (G) — User vault недоступен sandbox SID

Given credentials существуют только в vault normal user SID, а exact sandbox
SID видит тот же profile/configuration, но не может использовать vault, then
profile/config/helper metadata является diagnostic evidence, не capability
proof. Secret не копируется и login/elevation workaround не выполняется.

## S-033 (H) — Source пытается продолжить после handoff

Given source выпустил Release Handoff или destination уже выпустил Accept
Handoff, when source пытается resume Publisher, then ownership check останавливает
source до side effect. Только accepted destination может продолжать; duplicate
publication attempt запрещён даже без machine-enforced lock. Transfer ID не
является lock, secret или permission.

## S-034 — Destination interrupted до Accept Handoff

Given destination прерван во время inspect или capability probes до Accept
Handoff, then он не владеет publication и не выполняет mutation. Source также
остаётся observation-only; resume reconstruct-ит handoff и повторяет только
недоказанные read-only checks. Если exact attempt всё ещё доказан как open
`Released` с неизменными ID/Target/route и непротиворечивым predecessor chain,
recovery продолжает explicit route/Accept с тем же exact transfer ID. Fresh ID
требуется только для новой, повторной или reverse attempt; любой `Closed` ID
нельзя route, accept или reuse. `CancelledBeforeAccept` остаётся отдельным
explicit transition, который закрывает текущий ID.

## S-035 — Duplicate или mismatched acceptance

Given два releases/destination пытаются использовать один transfer ID, Accept
Handoff уже существует либо acceptance ссылается на unknown, reused,
mismatched, duplicate, already-accepted ID или другой Target, then дальнейшая
mutation останавливается. При unchanged Target доказанный exact owner
сохраняется; factual Target mismatch применяет
`InvalidatedByTargetChange/NoneTerminal`; ambiguous ownership становится
`Unknown` и не разрешается предположением.

## S-036 — Durable transfer identity и state recovery

Given candidate ID не имеет durable Release, then attempt `Unissued`,
authorization `Active`, owner остаётся source. Given Release event cites Target,
ID, actor и predecessor/tail, then attempt `Released`, ownership
`InTransitNone`. Exact route/Accept даёт `Accepted/Owned(destination)`. Every
event records all three axes and terminal reason/disposition. Если append-only
record недоступен, conflicting или допускает два predecessor/owner, ownership
`Unknown` и все mutations STOP; session memory не является evidence.

## S-037 — CancelledBeforeAccept возвращает releasing source

Given exact attempt `Released`, explicit user directive называет его ID, and
reconciliation доказывает отсутствие Accept и destination side effect, then
`Closed(CancelledBeforeAccept)` сохраняет authorization `Active` и возвращает
`Owned(recorded-release-source)`. Source повторно доказывает capability перед
mutation; старый ID нельзя route/accept/reuse. Без exact directive или
reconciliation cancellation запрещена и ownership остаётся `InTransitNone`.

## S-038 — Reverse после Accepted

Given destination имеет `Active/Owned(destination)/Accepted`, only этот
destination может выпустить fresh reverse Release. Release немедленно даёт
`InTransitNone/Released`; fresh route/Accept назначает нового owner. Если этот
attempt получает valid `CancelledBeforeAccept`, ownership возвращается
recorded releasing destination, не первоначальному source. Machine lock не
утверждается.

## S-039 — Target mismatch invalidates authorization

Given factual reconstruction обнаружил изменение любого immutable Target
field, then authorization становится `InvalidatedByTargetChange`, ownership
`NoneTerminal`, все связанные attempts — `Closed(TargetChanged)`. Ни source,
ни destination не используют прежнюю authorization; resumable handoff и
mutation запрещены.

## S-040 — User revoke и proven P10 различаются

Given explicit user revoke, then authorization `RevokedByUser`, ownership
`NoneTerminal`, open attempt `Closed(UserRevoked)` и более поздняя publication
требует нового exact gate. Given P10 independently proven and its predecessor
chain proves `Active/Owned(execution-context)` for the exact actor, then
authorization `Consumed(P10)`, ownership `NoneTerminal` и recovery выполняет
только reconciliation/report. If handoff не выпускался, attempt остаётся
`Unissued`, а publication-level P10 event имеет `transfer ID: none`; if current
attempt `Accepted`, exact accepting owner закрывает exact ID как
`Closed(CompletedP10)`; if attempt уже `Closed(reason)`, reason сохраняется и
P10 добавляется отдельным publication-level event only when current ownership
is still proven. Given attempt `Released` and ownership `InTransitNone`, P10 is
forbidden until exact Accept establishes the destination owner or valid
`CancelledBeforeAccept` restores the releasing owner. `NoneTerminal`, `Unknown`
or missing owner proof cannot create P10. Каждый terminal event содержит reason
и authorization/owner disposition; fabricated ID запрещён.

## Negative Disposition — S-041–S-049

These documentation decision scenarios extend, not replace, S-001–S-040.
Trace: ND-3–ND-5 PROCESS-001 and Publisher Authorization/P0/Terminal Success.
They do not execute publication.

| ID | Given / When | Required result |
|---|---|---|
| S-041 | Valid negative decision and separately authorized exact checkpoint, no publication permission | P0-P10 side effects forbidden; new exact publication gate required |
| S-042 | Permission for Negative Disposition exact single checkpoint/base/tuple; all P0 proofs succeed | Full unchanged P0-P10; no A/B relabeling or extra commit range |
| S-043 | Negative tuple changed, wrong class, extra commit, changed parent/base or stale manifest | Target invalidation; no reuse of permission |
| S-044 | C P0 dirty, API-only/Git-only capability or unknown/released owner | STOP before mutation; all ordinary dual probes/ownership required |
| S-045 | Interruption at each P0-P10 with known/unknown outcomes | Inspect exact remote/PR/merge/ref facts first; before P6 task phase, after P6 main phase, no replay completed mutation |
| S-046 | C handoff/revoke/target mismatch/P10 | Same three-axis ownership/authorization protocol S-026–S-040; no new authority from negative class |
| S-047 | Exact negative P10 plus MERGED PR/ancestry, both refs absent, clean synchronized main | Report negative-only publication and Sealed Negative Disposition; no original Acceptance/BCC/Completed or automatic next task |
| S-048 | Negative merge/P9 but no terminal P10, dirty main, missing terminal evidence, or downstream positive use | No sealing/intake release; positive proof rejected; preserve original Not Proven/Disproven |
| S-049 | New concrete provenance pointer arrives after decision or mid-publication, including after merge before P10 | ND-5 holds first remaining mutation; independently revalidate eligibility, preserve/reconstruct prior effects; no forced cleanup/P10, no fabricated TargetChanged if tuple unchanged; after failed eligibility STOP |

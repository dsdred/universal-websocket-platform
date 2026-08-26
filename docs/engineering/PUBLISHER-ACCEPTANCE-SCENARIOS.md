# Publisher Acceptance Scenarios

## Purpose

Эти process scenarios проверяют контракт Publisher из
[PROCESS-001](PROCESS-001-AI-DEVELOPMENT-WORKFLOW.md) и
[Publisher role](agents/publisher.md). Во всех сценариях accepted task commit
либо blocked evidence checkpoint уже создан по отдельной команде
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
accepted/certified scope изменились, read-only
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

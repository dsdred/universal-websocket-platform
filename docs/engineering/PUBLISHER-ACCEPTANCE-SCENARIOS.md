# Publisher Acceptance Scenarios

## Purpose

Эти process scenarios проверяют контракт Publisher из
[PROCESS-001](PROCESS-001-AI-DEVELOPMENT-WORKFLOW.md) и
[Publisher role](agents/publisher.md). Во всех сценариях task commit уже
создан по отдельному разрешению.

## S-001 — Полностью успешный pipeline

Given clean exact task branch/OID и команда `Разрешаю публиковать.`, Publisher
выполняет P0–P10 без дополнительных permission prompts. Terminal report
содержит PR number/URL, task и merge OID, checks state, `MERGEABLE / CLEAN`,
обе branch deletions, `main == origin/main`, clean worktree и STOP.

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
branch/HEAD task OID, не recreates её и продолжает exact cleanup step. Новая
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

## S-009 — PR не mergeable

Given conflict, `UNKNOWN`, non-clean gate или branch-protection refusal, P4
остаётся первым незавершённым. PR и refs сохраняются; Publisher не выполняет
force/rebase/bypass.

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

Given после merge remote task ref указывает не на authorized task OID,
Publisher не удаляет ref. P5 остаётся незавершённым, текущее состояние
сохраняется и сообщается как cleanup blocker.

## S-014 — Target invalidation

Given branch, task OID, base или accepted scope изменились, read-only
preflight останавливает pipeline до mutation. Это invalidation exact authority,
а не внешний blocker; старое разрешение не применяется к новому target.

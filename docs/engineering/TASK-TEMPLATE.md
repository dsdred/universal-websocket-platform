# TASK-NNN — Название задачи

## Status

`Planned`, `In Progress`, `Blocked` или `Completed`.

## Task Contract

### Task Mode

`Design-only`, `Design-update`, `Implementation` или `Documentation-only` с
обоснованием.

### Why Now

- почему выбрана именно эта task;
- какие зависимости выполнены;
- что task разблокирует;
- почему scope целен.

### Definition of Done

Конкретные проверяемые критерии завершения.

### Out of Scope

Явно запрещённые изменения.

### Verification Plan

Существующие tests/checks, Coverage Gap и risk-based дополнительные проверки.

## Objective

Однозначный проверяемый результат задачи.

## Selection Evidence

- откуда выбран candidate;
- подтверждённые prerequisites;
- применённые readiness и ranking rules;
- рассмотренные и отклонённые alternatives;
- причина остановки, если однозначный выбор невозможен.

## Scope

- разрешённые подсистемы и файлы;
- обязательные deliverables;
- явно исключённая работа.

## Non-Goals

- следующая task, которая не начинается автоматически;
- преждевременная integration или capability;
- unrelated refactoring и speculative work.

## Sources of Truth

- применимые Approved ADR;
- Active/Frozen ARCH;
- Approved/Accepted DP;
- factual implementation evidence;
- связанные task records.

## Roles

- Coordinator:
- Architect:
- Documentation Agent:
- Developer:
- Tester:
- Reviewer:
- Publisher:

Неприменимый stage должен иметь явное обоснование.

## Branch

- исходный trusted baseline:
- task branch:
- branch action:
- запрещённые git actions:

## Constraints

- запрещённые изменения;
- compatibility и ownership invariants;
- commit policy.

## Stop Conditions

- архитектурный или продуктовый blocker;
- conflicting sources;
- dirty, unattributed или diverged baseline;
- materially different candidates либо scope expansion.

## Acceptance Criteria

1. Проверяемое требование.
2. Требуемое evidence.

## Verification

- Existing Coverage Report:
  - Existing Coverage;
  - Coverage Gap;
  - Added Proof Tests;
  - Added Regression Tests;
  - Remaining Limitations;
- Verification Matrix:
  - concurrency/lifecycle/shared state;
  - API/CLI/UI/configuration/production wiring;
  - dependencies;
  - public API;
  - documentation;
- formatter/lint:
- tests:
- race/vet:
- documentation structure:
- independent review:

## Scope Audit

Для каждого изменённого production, test, documentation и generated-файла:

- классификация: `Required`, `Questionable` или `Removable`;
- связь с acceptance criteria;
- меняет ли поведение либо является механической migration;
- disposition для `Questionable` и `Removable`.

Для каждого `Questionable`: пункт Definition of Done, почему task без change
некорректна и почему change нельзя вынести. Final Reviewer отвечает, можно ли
удалить change и сохранить Definition of Done.

Отдельно проверить premature next-task/pipeline work, unrelated refactoring,
generated, formatting-only и незадокументированное planned behavior.

## Size Guard

- сработавшие признаки: >15 files, >500 production lines, >1 new package,
  >1 architecture contract, >1 independently shipped behavior;
- решение: доказательство целостности либо split до дальнейшей work.

## Documentation Sync

- task record:
- current-state:
- MASTER_PLAN EN/RU:
- связанные Design Proposal:
- PROJECT_CONTEXT:
- CHANGELOG только для user-facing/release change:
- parity, links и contradictions:

## Interruption Recovery

- persistent anchor: repository, Task ID/status, branch, baseline/HEAD OID,
  scope, roles и ordered applicable stages;
- current evidence subject/exclusions: exact file set in ascending unsigned
  UTF-8 path-byte order (case-sensitive, locale-independent), regions and
  allowed evidence envelope per PROCESS-001;
- canonical subject-manifest rows/object format/OID;
- pre-commit envelope provenance доказан либо affected Review/Acceptance
  повторены; post-commit final tree/commit OID reconstruct-ится из Git
  refs/history и не записывается внутрь того же commit;
- proven completed checkpoints и independently reproducible evidence;
- first checkpoint without proven completion;
- unknown/inconsistent operations и required reconciliation;
- permission state: current explicit command/resume signal либо `not proven`;
  task record не является permission;
- operation reconciliation: file mutation, stage, commit, push, PR, merge,
  branch deletion и documentation/status transitions;
- downstream evidence invalidated by rework/content change;
- новый агент без chat history способен продолжить: да/нет и evidence.

## Commit Gate

- exact command `Разрешаю коммит.` получена: да/нет;
- gate class: `Coordinator Accepted` / `Blocked Closure Certified` / not ready;
- commit message policy:
- exact file set:
- post-acceptance/certification diff:
- temporary/generated/unrelated files:
- final checks:

### Blocked Evidence Checkpoint (если применимо)

- task status остаётся `Blocked`:
- Coordinator Acceptance: `не пройден`;
- exact blocker и missing prerequisite:
- prerequisite: `Not Activated`;
- evidence-only scope и подтверждение отсутствия product implementation:
- certification tuple: repository, Task ID, branch, base/OID, HEAD OID, exact
  file set in PROCESS-001 ascending unsigned UTF-8 path-byte order,
  staging-invariant canonical evidence
  digest command/certified HEAD/object format/OID, blocker identity,
  verification/review results;
- `Blocked Closure Certified`: да/нет, кем и когда;
- exact checkpoint commit OID либо `not authorized/not created`:

## Process Health

- trigger применим: да/нет и причина;
- bounded findings либо отсутствие process change;

## Handoff

- выполненный scope;
- изменённые файлы;
- результаты проверок;
- открытые findings и risks;
- следующий разрешённый шаг.

## Publication

- publication readiness отдельно от completion;
- publication class: `Accepted Task` / `Blocked Evidence Recovery`;
- repository:
- exact branch:
- ordered commit target и head OID:
- base `main`:
- accepted/certified verification и scope:
- Publisher P0–P10 state или `not authorized`;
- при blocker: completed steps, exact first unfinished step, preserved state и
  phase (`task branch` до P6 либо `main` после P6), known PR/merge OID и
  confirmation, что permission остаётся действительным;
- при terminal success: PR, task/merge commits, checks, merge gate, обе branch
  deletions, `main == origin/main`, clean worktree и STOP.
- для blocked recovery: reconstructable terminal P10/merged PR/OID/refs state
  и exact prerequisite, который остаётся `Not Activated` до отдельного normal
  intake; durable facts синхронизируются при следующем применимом PROCESS-002;

Post-commit/PR/merge OID в этой секции является handoff/reconstruction evidence
из Git/GitHub. Он не требует и не разрешает mutating update того же immutable
task commit; durable post-publication запись выполняется только отдельным
применимым synchronization transition.

## Next Candidate

- рекомендуемая Ready work:
- readiness evidence:
- явно не начата:

## Closure

- Final status:
- closure class: `Coordinator Accepted` / `Blocked Closure Certified`;
- при blocked closure: подтверждение, что task не Accepted/Completed и
  prerequisite не активирован;
- Closed by:
- Date:

## Recovery Evidence Envelope

Эта секция обязана быть последним top-level `##` heading task record и следует
projection `task-record-v1` PROCESS-001. Status evidence, role verdicts,
Acceptance tuple, subject-manifest rows/OID, post-acceptance integrity и
pre-commit readiness фиксируются здесь. Секция не self-attest-ит свои bytes;
до commit недоказанная после interruption envelope mutation возвращает
affected gate. После commit final bytes доказывает reconstructable Git tree
OID, который не записывается внутрь того же commit.

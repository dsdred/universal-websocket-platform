# TASK-NNN — Название задачи

## Status

`Planned`, `In Progress`, `Blocked` или `Completed`.

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

Отдельно проверить premature next-task/pipeline work, unrelated refactoring,
generated, formatting-only и незадокументированное planned behavior.

## Handoff

- выполненный scope;
- изменённые файлы;
- результаты проверок;
- открытые findings и risks;
- следующий разрешённый шаг.

## Next Candidate

- рекомендуемая Ready work:
- readiness evidence:
- явно не начата:

## Closure

- Final status:
- Closed by:
- Date:

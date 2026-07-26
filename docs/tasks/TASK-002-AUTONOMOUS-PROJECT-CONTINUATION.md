# TASK-002 — Упрощённое автономное продолжение проекта

## Status

**Completed — Coordinator Accepted**

## Objective

Внедрить минимальный repository-native процесс, при котором точная команда
`Продолжай проект.` позволяет Coordinator самостоятельно выбрать одну Ready
work, подготовить task и безопасную локальную branch, провести полный
PROCESS-001 cycle, выполнить scope audit, принять результат и рекомендовать
следующую работу без неявных commit, push или merge.

## Selection Evidence

- Источник задачи: явный пользовательский запрос на первую версию упрощённого
  автономного процесса.
- Current milestone: Beta — Complete the Single-Node Runtime.
- Existing process: PROCESS-001 уже определял роли и полный development cycle,
  но не определял bare-command routing, детерминированный выбор следующей
  работы, branch authority и обязательный scope audit.
- Ready slice: Stage 1 является documentation-only evolution существующего
  operational process, не требует изменения product или Runtime architecture
  и независимо проверяется repository documentation review.
- Отклонены:
  - scripts, bots и background automation — преждевременны без практического
    evidence Stage 1;
  - новый PROCESS — создавал бы конкурирующий источник истины;
  - запуск DP-009 — отдельная следующая работа, не часть governance task.

## Evolutionary Plan

1. **Stage 1 — Repository-native governance.** Точная команда, deterministic
   selection, safe local branch authority, task evidence, полный role cycle,
   scope audit, acceptance, project-state update и next recommendation.
2. **Stage 2 — Practical validation.** Применить Stage 1 к следующей UWP task,
   зафиксировать реальные ambiguities и минимально уточнить существующие
   contracts.
3. **Stage 3 — Evidence-driven tooling.** Только при подтверждённой
   повторяемости рассмотреть узкие локальные checks или templates; не вводить
   bots, network automation или background execution без отдельной task.

Stage 1 достаточен для практической проверки процесса на следующей задаче.
Stages 2 и 3 этой task не начинаются.

## Scope

- `docs/engineering/AGENT.md`: exact bare-command routing и ограниченные
  полномочия;
- `docs/engineering/PROCESS-001-AI-DEVELOPMENT-WORKFLOW.md`: selection,
  readiness, ranking, branch, full cycle и scope audit;
- `docs/engineering/agents/coordinator.md`: исполняемый Coordinator algorithm,
  stop conditions и Final Report;
- `docs/engineering/TASK-TEMPLATE.md`: постоянное evidence;
- этот task record и `docs/tasks/README.md`;
- operational state в `.ai/PROJECT_CONTEXT.md` и `spec/current-state.md`.

## Non-Goals

- новый PROCESS, ADR, ARCH, DP или изменение MASTER_PLAN;
- изменение production code, tests или product architecture;
- scripts, bots, GitHub automation, background agents или network operations;
- изменение контрактов ролей, кроме уточнения orchestration Coordinator;
- продуктовая приоритизация;
- автоматические stage, commit, push, merge, branch deletion, fetch или pull;
- запуск DP-009, повышение его Design/Implementation Status либо production
  Loader-to-Builder-to-Launcher pipeline.

## Sources of Truth

- [Agent Contract](../engineering/AGENT.md);
- [PROCESS-001](../engineering/PROCESS-001-AI-DEVELOPMENT-WORKFLOW.md);
- [PROCESS-002](../engineering/PROCESS-002-DOCUMENTATION-SYNCHRONIZATION.md);
- [Coordinator contract](../engineering/agents/coordinator.md);
- [MASTER_PLAN RU](../ru/roadmap/MASTER_PLAN.md) и
  [EN mirror](../en/roadmap/MASTER_PLAN.md);
- [current state](../../spec/current-state.md);
- [decisions](../../spec/decisions.md);
- [Draft DP-009 RU](../ru/design/DP-009-runtime-bootstrap-contract.md) и
  [EN mirror](../en/design/DP-009-runtime-bootstrap-contract.md).

## Roles

- **Coordinator:** intake, назначение ролей, acceptance и next recommendation.
- **Architect:** подтвердил boundaries Stage 1, deterministic selection,
  branch authority и DP-009 dry-run result.
- **Documentation Agent:** реализует documentation-only Stage 1.
- **Developer:** не применяется; production code не меняется.
- **Tester:** независимо выполнил read-only verification documentation-only
  diff; Go tests неприменимы. Первичный verdict `FAIL` вернул три factual
  documentation findings; после ограниченного rework повторная verification
  завершилась итоговым verdict `PASS`, findings resolved.
- **Reviewer:** независимо проверяет process safety, determinism, scope и
  documentation consistency. Первичный review вернул `Needs Revision` с тремя
  governance-consistency findings; после ограниченного rework итоговый verdict
  — `Approved`.

## Branch

- **Исходный trusted baseline:** merge TASK-001 в `main`, commit `e13b693`.
- **Task branch:** `docs/autonomous-project-governance`.
- **Branch action:** Coordinator создал и переключил
  `docs/autonomous-project-governance` из clean `main` на `e13b693` до того,
  как правило Task ID для новых branch slugs было финализировано в этой же
  documentation-only task. Поэтому текущая branch является grandfathered
  exception и не переименовывается.
- **Запрещено:** stage, commit, push, merge, delete, fetch, pull и изменение
  `main` без отдельного разрешения пользователя.

## Constraints

- Сохранить PROCESS-001/002 и существующие role boundaries.
- Не превращать MASTER_PLAN в task queue.
- Не документировать planned behavior как implemented product capability.
- Bare-команда даёт только read-only intake и безопасную локальную branch
  preparation; все необратимые или remote git actions требуют разрешения.

## Stop Conditions

- противоречие с Approved/Frozen architecture;
- необходимость product prioritization;
- materially different Ready candidates;
- любой dirty baseline при выборе или подготовке новой task;
- unattributed dirty worktree всегда;
- attributed dirty worktree допускается только для resume единственной active
  task, когда branch и изменения однозначно соответствуют её record;
- diverged либо непонятный baseline;
- необходимость scope expansion за Stage 1.

## Acceptance Criteria

1. Exact bare-команда маршрутизируется однозначно и не расширяет git authority.
2. Selection имеет deterministic resume/readiness/ranking/tie/stop rules.
3. Task record создаётся до work и хранит selection, branch, scope, non-goals,
   verification, scope audit и next candidate.
4. Full cycle сохраняет roles, independent review, PROCESS-002 и rework.
5. Scope audit классифицирует каждый файл и обнаруживает premature,
   unrelated, generated и formatting-only changes.
6. Coordinator acceptance обновляет project state и рекомендует, но не
   запускает следующую task.
7. Документы достаточны для нового агента без истории чата.

## Current-Repository Dry Run

Dry run разделяет фактические состояния репозитория:

1. **Во время выполнения TASK-002:** worktree содержал однозначно
   атрибутированные изменения, а TASK-002 имела статус `In Progress`.
   Bare-команда должна была возобновить TASK-002; новую task выбирать было
   нельзя.
2. **Сейчас, после closure, но до отдельно разрешённого commit:** завершённый
   diff всё ещё оставляет baseline dirty. Bare-команда должна остановиться;
   следующая task и branch не создаются.
3. **После отдельно разрешённого commit и появления clean trusted baseline со
   Stage 1:** next-candidate simulation пересекает roadmap, current-state gap
   и Draft DP-009 §23. Она рекомендует:

**сфокусированное уточнение implementation prerequisites Draft DP-009**.

Этот будущий candidate является ограниченным architecture/documentation slice:
concrete Bootstrap input, dependency bindings и failure representation.
Реализация DP-009, Runtime Launcher, Runtime Lifecycle Owner и production
pipeline пока не Ready. Simulation не создаёт следующую task, не создаёт
branch, не изменяет DP-009 и не повышает его статусы.

## Stage 1 Self-Hosting Evidence

Stage 1 определяет строгий invariant: task record является первым content
change после подготовки branch. Однако сама Stage 1 не доказывает его
соблюдение для TASK-002.

Filesystem evidence, зафиксированное при первичной реализации, показывает:

- `docs/engineering/TASK-TEMPLATE.md` был изменён в `00:48:56`;
- `docs/tasks/TASK-002-AUTONOMOUS-PROJECT-CONTINUATION.md` был создан в
  `00:48:57`.

Следовательно, как минимум один operational content change предшествовал
созданию task record. Эту chronology невозможно исправить ретроактивно.
Отклонение фиксируется честно и не ослабляет invariant для новых tasks.
Практическое доказательство task-before-work ordering обязательно для Stage 2.

## Verification

- repository Markdown links: PASS, 97 Markdown files;
- Markdown fences: PASS, 97 Markdown files;
- conflict markers: PASS;
- trailing whitespace: PASS для всех восьми изменённых файлов; repository-wide
  scan также обнаруживает существующий вне scope whitespace в зеркалах
  `runtime-alpha-review.md`, которые этой task не изменялись;
- required EN/RU parity applicability: N/A — изменены только internal
  operational documents без обязательного EN mirror; публичные EN/RU trees не
  изменены;
- forbidden scope: PASS — production, tests, ADR, ARCH, DP, MASTER_PLAN,
  CHANGELOG и generated files не изменены;
- Go formatter, tests, race и vet: N/A — production code и tests отсутствуют
  в documentation-only diff;
- `git diff --check`: PASS;
- independent Tester: initial `FAIL` — три factual documentation findings;
  после ограниченного rework итоговый verdict `PASS`, findings resolved;
- independent Reviewer: initial `Needs Revision` — три
  governance-consistency findings; после rework итоговый verdict `Approved`.

## Scope Audit

Все восемь файлов классифицированы `Required`:

- `docs/engineering/AGENT.md` — exact entry routing и ограничение authority;
- `docs/engineering/PROCESS-001-AI-DEVELOPMENT-WORKFLOW.md` — selection,
  branch, full cycle и scope audit;
- `docs/engineering/agents/coordinator.md` — role algorithm и Final Report;
- `docs/engineering/TASK-TEMPLATE.md` — воспроизводимое task evidence;
- `docs/tasks/TASK-002-AUTONOMOUS-PROJECT-CONTINUATION.md` — contract,
  verification и permanent handoff;
- `docs/tasks/README.md` — navigation;
- `.ai/PROJECT_CONTEXT.md` и `spec/current-state.md` — operational state и
  честная следующая рекомендация.

`Questionable` и `Removable` changes отсутствуют. Production, tests, ADR,
ARCH, DP, MASTER_PLAN, CHANGELOG, generated и formatting-only files не
изменены. DP-009 refinement не начат, production pipeline не подключён, а
planned product behavior не представлено как implemented.

Coordinator принял итоговый scope audit: 8 `Required`, 0 `Questionable`, 0
`Removable`; forbidden scope отсутствует.

## Handoff

- **Stage 1 documentation implementation:** завершена.
- **Изменённые файлы:** восемь `Required` operational documents из Scope
  Audit; production и public architecture отсутствуют.
- **Verification:** Markdown links/fences, conflict markers, trailing
  whitespace, forbidden scope и `git diff --check` — PASS.
- **Findings и risks:** implementation findings отсутствуют; практическая
  проверка selection rules и task-before-work ordering отложена до Stage 2.
  Три factual Tester findings исправлены, итоговый Tester verdict — `PASS`.
  Три Reviewer findings исправлены: dirty semantics, factual Tester stage и
  self-hosting chronology; итоговый Reviewer verdict — `Approved`. Existing
  trailing whitespace в EN/RU
  `runtime-alpha-review.md` не создан этой task и требует отдельной
  repository-maintenance оценки.
- **Следующий разрешённый шаг:** отдельно разрешённый commit закрытой TASK-002.
  До него bare-команда распознаёт закрытую task и attributed dirty worktree,
  останавливается и не выбирает новую task.

## Next Candidate

- **Рекомендация после clean trusted Stage 1 baseline:** focused DP-009
  implementation-prerequisites refinement.
- **Readiness evidence:** DP-009 §23, текущий pipeline gap и завершённые
  изолированные DP-007/DP-008 prerequisites.
- **Статус:** simulation only; явно не начата, task и branch не созданы.

## Closure

- **Final status:** Completed — Coordinator Accepted.
- **Tester:** PASS after factual documentation rework.
- **Reviewer:** Approved after governance-consistency rework.
- **Scope audit:** Accepted — 8 Required, 0 Questionable, 0 Removable.
- **Stage 1 availability:** доступен в текущем attributed dirty worktree ветки
  `docs/autonomous-project-governance`; новая task запрещена до clean trusted
  baseline.
- **Unresolved risk:** Stage 1 self-hosting deviation сохранён; практическое
  доказательство task-before-work ordering обязательно в Stage 2.
- **Closed by:** Coordinator.
- **Date:** 2026-07-27.

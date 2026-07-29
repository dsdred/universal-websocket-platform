# TASK-012 — Engineering Process Hardening

## Status

**Completed — Coordinator Accepted**

## Task Contract

### Task Mode

**Documentation-only.**

Изменение уточняет только operational workflow Coordinator/Publisher и
связанные process records. Product architecture, production code, tests,
Runtime Launch Flow и Production Activation не меняются. `Design-update`
неприменим: новый product или architecture contract не создаётся.

### Why Now

- TASK-011 завершена со статусом `Completed — Coordinator Accepted`.
- PROCESS-001/002 и Publisher governance проверены серией TASK-002–TASK-011 и
  выявили повторяемые gaps в intake, coverage accounting, risk-based
  verification, scope control, commit gate, documentation closure и Publisher
  resume/cleanup evidence.
- Эти gaps должны быть устранены до отдельного анализа Production Activation,
  чтобы следующая product task начиналась с устойчивого процесса.
- Scope целен: все изменения относятся к одному Coordinator/Publisher
  workflow и проверяются как единый непротиворечивый documentation-only slice.

### Definition of Done

1. Новые правила встроены в существующие нормативные process sources без
   создания альтернативной методологии.
2. Пользовательские permission/gate-команды `Продолжай проект.`,
   `Разрешаю коммит.` и
   `Разрешаю публиковать.` сохраняют прежний смысл и остаются единственными
   пользовательскими gates полного сценария; resume-сигнал Publisher не
   выдаёт нового разрешения.
3. Coordinator фиксирует Task Contract до реализации.
4. Добавлен Existing Coverage Report до создания или изменения тестов.
5. Добавлена risk-oriented Verification Matrix.
6. Scope Audit требует доказательства для каждого `Questionable` change и
   явного removable-question Final Reviewer.
7. Добавлен Size Guard без жёсткого лимита строк или файлов.
8. Закреплена обязательная Documentation Sync с явной применимостью task
   record, current-state, MASTER_PLAN, Design Proposal, PROJECT_CONTEXT и
   CHANGELOG.
9. Добавлен Commit Gate без GPG, DCO или обязательного sign-off.
10. Publisher повторно проверяет CI/mergeability перед merge, конфликт после
    push, repository merge strategy, cleanup с `git fetch --prune`, exact
    checkpoint/blocker report, обе task branches, `main == origin/main` и
    clean worktree.
11. Добавлен лёгкий Process Health Review с заданными triggers и минимальными
    показателями без production metric tooling.
12. Внутренние process documents не требуют искусственного EN mirror; все
    затронутые публичные EN/RU документы синхронизированы.
13. TASK-012, current-state и зеркальные MASTER_PLAN отражают completion.
14. Documentation checks, Tester, Scope Audit и независимый Final Reviewer
    проходят.

### Out of Scope

- Production-код UWP и product tests.
- Production Activation и её readiness/design analysis.
- Runtime Launch Flow, новые API и зависимости.
- Изменение смысла трёх пользовательских команд.
- Альтернативный PROCESS или новая методология.
- Fast path для малых задач.
- GPG, DCO и обязательный sign-off.
- Сложная система метрик, telemetry или production analytics процесса.
- Commit, push, PR, merge и удаление веток.

### Verification Plan

- Сначала инвентаризировать существующие process checks и acceptance
  scenarios, затем зафиксировать Existing Coverage и Coverage Gap.
- Проверить command semantics и normative consistency между AGENT,
  PROCESS-001/002, Coordinator, Publisher, task template и acceptance
  scenarios.
- Применить Verification Matrix к documentation-only diff: EN/RU parity,
  relative links, headings/fences при зеркалах, противоречия нормативных
  источников, whitespace/conflict markers и `git diff --check`.
- Проверить, что production, test, API, dependency и `.github` scope не
  изменены.
- Выполнить независимые Tester и Final Reviewer passes; все blocking findings
  направить в ограниченный Rework Loop.

## Selection Evidence

- Работа явно назначена пользователем как следующая отдельная
  documentation/process-only task.
- Clean baseline: `main == origin/main`, worktree clean, active task
  отсутствовала.
- TASK-011 завершена; Production Activation явно запрещена в этой task.
- TASK-002 и TASK-008 являются историческими records внедрения autonomous
  Coordinator и Publisher governance; PROCESS-001/002 и role contracts
  являются текущими нормативными sources.
- Отклонены:
  - рекомендованный после TASK-011 Production Activation analysis — явно
    запрещён пользователем;
  - новый PROCESS-003 — создавал бы конкурирующий источник истины;
  - executable process tooling — не требуется для минимального изменения;
  - production implementation — вне scope.

## Scope

- `docs/engineering/AGENT.md`;
- `docs/engineering/PROCESS-001-AI-DEVELOPMENT-WORKFLOW.md`;
- `docs/engineering/PROCESS-002-DOCUMENTATION-SYNCHRONIZATION.md`;
- Coordinator, Documentation Agent, Tester, Reviewer и Publisher role
  contracts;
- `docs/engineering/TASK-TEMPLATE.md`;
- Publisher acceptance scenarios, если для новых Publisher gates требуется
  proof;
- task index и этот task record;
- `.ai/PROJECT_CONTEXT.md`, `spec/current-state.md`;
- зеркальные `docs/en/roadmap/MASTER_PLAN.md` и
  `docs/ru/roadmap/MASTER_PLAN.md`.

## Non-Goals

- изменения ADR, ARCH, DP, product specifications и implementation status;
- запуск следующей task;
- переписывание существующего workflow целиком;
- дублирование полного набора правил во всех role documents;
- автоматизация Process Health Review.

## Sources of Truth

- корневой `AGENTS.md` и `docs/engineering/AGENT.md`;
- PROCESS-001 и PROCESS-002;
- Coordinator, Documentation Agent, Tester, Reviewer и Publisher contracts;
- Publisher acceptance scenarios;
- TASK-002, TASK-008 и TASK-011;
- task template и task index;
- `.ai/PROJECT_CONTEXT.md`, `spec/current-state.md`,
  `spec/decisions.md`;
- зеркальные MASTER_PLAN.

## Roles

- **Coordinator:** Task Contract, sequencing, handoffs, Size Guard, Scope
  Audit, acceptance и project-state synchronization.
- **Documentation Agent:** минимальные изменения нормативных документов.
- **Tester:** Existing Coverage Report и независимая documentation
  verification.
- **Reviewer:** независимый Final Review и removable-question.
- **Publisher:** contract audit only; publication не выполняется.
- **Architect:** неприменим; product architecture не меняется.
- **Developer:** неприменим; production code и executable tests не меняются.

## Branch

- исходный trusted baseline: clean synchronized `main`;
- task branch: `docs/task-012-process-hardening`;
- branch action: создана до content changes;
- этот task record является первым content change;
- stage, commit, push, PR, merge, fetch/pull и branch deletion запрещены.

## Constraints

- Минимальный нормативный набор: PROCESS-001 владеет общим workflow,
  PROCESS-002 — documentation synchronization, role documents — только
  role-specific obligations, acceptance scenarios — executable examples.
- Ссылки из entry points и шаблона могут кратко направлять к нормативному
  правилу, но не должны создавать противоречащую копию.
- Канонический язык `docs/engineering/` и `docs/tasks/` — русский; EN/RU parity
  применяется к публичным зеркальным деревьям.
- CHANGELOG меняется только при user-facing или release change.

## Stop Conditions

- изменение требует product/architecture решения;
- обнаружено противоречие Approved/Frozen источнику;
- Production Activation входит в diff;
- production code, tests, dependencies или `.github` входят в diff;
- новые правила требуют четвёртой пользовательской команды;
- нормативные источники невозможно согласовать без альтернативной
  методологии;
- blocking Tester/Reviewer finding не устранён.

## Existing Coverage Report

### Existing Coverage

- PROCESS-001 уже определяет task-before-work record, полный role cycle,
  risk-agnostic verification, Scope Audit, documentation synchronization,
  closure и commit only by explicit permission.
- PROCESS-002 уже определяет drift detection, project-state applicability и
  stable-vs-ephemeral Publisher state.
- Publisher contract и acceptance scenarios уже покрывают P0–P10,
  immutable target, resume reconstruction, CI classes, merge gate, cleanup и
  terminal evidence.
- TASK template уже содержит scope, verification, audit, handoff, publication
  и closure sections.

### Coverage Gap

- Task Contract не стандартизирует Task Mode, Why Now, DoD, Out of Scope и
  Verification Plan как обязательный pre-work block.
- Нет обязательного Existing Coverage Report перед изменением тестов.
- Verification не выбирается через явную risk matrix.
- Нет Size Guard, усиленного `Questionable` proof и removable-question.
- Documentation Sync и Commit Gate не сведены в явные завершительные gates.
- Publisher не фиксирует все запрошенные recheck/conflict/strategy/prune
  детали.
- Process Health Review отсутствует.

### Added Proof Tests

- S-015 доказывает post-push conflict gate на P3.
- S-016 доказывает повторный CI/head/base/mergeability gate перед P4.
- S-017 доказывает выбор только repository-approved merge strategy.
- S-018 доказывает `git fetch --prune` и exact cleanup evidence.

### Added Regression Tests

- exact command scan отделяет три permission/gate-команды от невыдающего
  authority Publisher resume-сигнала;
- S-009/S-015 разделяют conflict, впервые обнаруженный на P3, и conflict,
  появившийся при P4 recheck;
- changed-document relative links, conflict markers, whitespace,
  `git diff --check`, forbidden-scope scan и MASTER_PLAN EN/RU structure
  проверяются после каждого rework.

### Remaining Limitations

- Documentation process проверяется repository evidence и scenarios, а не
  production telemetry.
- Наличие CI, race detector и executable smoke зависит от характера будущей
  task; недоступность должна давать `PASS WITH LIMITATION`, а не молчаливый
  PASS.

## Size Guard

**Triggered and accepted after reassessment.** Фактический diff содержит
16 documentation/project-state files, то есть превышает порог 15 на один
файл. Production-код равен 0 строк; новых packages, architecture contracts и
независимо поставляемых поведений нет. Scope остаётся единым:

- PROCESS-001/002 владеют общими нормативными правилами;
- entry point, template и пять role contracts содержат только routing или
  role-specific obligations;
- Publisher contract и scenarios должны меняться совместно;
- task record/index и четыре обязательных project-state/EN-RU mirror files
  обеспечивают traceability и completion sync.

Удаление любого из этих классов оставляет невыполненным соответствующий пункт
Definition of Done; split создал бы временное противоречие нормативных
источников.

## Verification Matrix

| Risk class | Applicability | Required evidence |
|---|---|---|
| Concurrency/lifecycle/shared state | Не применяется к diff | Подтвердить отсутствие production/test changes; race не запускать |
| API/CLI/UI/config/production wiring | Только три permission/gate-команды | Scenario/manual executable proof сохранения exact semantics |
| Dependencies | Не применяется | `go.mod`/`go.sum` отсутствуют в diff |
| Public API | Не применяется | Exported Go surface отсутствует в diff |
| Documentation | Применяется | EN/RU parity, links, contradictions, headings/fences where mirrored, diff checks |

## Scope Audit

**Accepted candidate — 16 Required, 0 Questionable, 0 Removable.**

- Normative workflow: `AGENT.md`, PROCESS-001 и PROCESS-002 — Required для
  DoD 1–11.
- Role-specific obligations: Coordinator, Documentation Agent, Tester,
  Reviewer и Publisher contracts — Required для DoD 3–10.
- Reusable evidence: task template и Publisher acceptance scenarios —
  Required для DoD 3–11.
- Traceability: TASK-012 и task index — Required для DoD 1, 13 и 14.
- Project state: PROJECT_CONTEXT, current-state и MASTER_PLAN EN/RU —
  Required для DoD 12–14 и явного требования пользователя.
- Production, tests, API, dependencies, `.github`, ADR/ARCH/DP, generated и
  unrelated files отсутствуют.
- Следующая Production Activation task не начата и не анализировалась.

Questionable/Removable отсутствуют; disposition не требуется. Final Reviewer
должен независимо ответить на removable-question для каждого класса.

## Documentation Sync

**PROCESS-002 status: Synchronized.**

- task record и task index обновлены;
- current-state и PROJECT_CONTEXT отражают TASK-012 как последнюю завершённую
  operational task и отсутствие active operational task;
- MASTER_PLAN EN/RU имеют зеркальные Engineering Process sections;
- Design Proposal — `Not applicable`, product design/status не менялись;
- CHANGELOG — `Not applicable`, user-facing/release change отсутствует;
- links broken 0, MASTER_PLAN headings 36/36 и fences 0/0;
- critical drift и нормативные противоречия после Tester rework не известны.

## Commit Gate

Commit не разрешён этой task. До возможного будущего commit Coordinator обязан
повторно проверить policy сообщения, exact file set, отсутствие post-acceptance
изменений и временных/generated/unrelated файлов. GPG, DCO и sign-off не
добавляются.

## Handoff

- Documentation implementation: complete.
- Initial Tester: `FAIL`, findings B-001–B-005 направлены в bounded rework.
- Repeat Tester: B-001/B-002 resolved; closure/state findings возвращены в
  rework.
- Final repeat Tester: `PASS`, 0 blocking и 0 nonblocking findings; все
  B-001–B-005 resolved.
- Focused post-review Tester: `PASS`, Coordinator coverage gate и полный
  16-file/five-role accounting подтверждены.
- Changed files: 16 Required documentation/project-state files; production,
  tests и generated files отсутствуют.
- Verification: `git diff --check`, links, conflict markers, forbidden scope,
  scenario count и MASTER_PLAN parity — PASS.
- Independent Final Reviewer: initial `Needs Revision`, два blocking findings
  устранены; repeat verdict `Approved`, 0 blocking и 0 nonblocking findings.
- Coordinator Acceptance: получена.

## Process Health Review

**Triggered.** TASK-002–TASK-011 образуют десять завершённых task после
введения autonomous workflow.

- Scope Audit records: documented `Questionable`/`Removable` — 0.
- Defects after primary verification: TASK-002 и TASK-007 records фиксируют
  bounded rework findings до final acceptance; post-merge product defects не
  зафиксированы.
- CI failures after local PASS и post-merge fixes: repository evidence не
  зафиксировало.
- Unavailable checks: race detector повторно был недоступен в TASK-001,
  TASK-004, TASK-006, TASK-009 и TASK-011.
- Repeated extra work: task-contract/closure synchronization и Publisher
  checkpoint ambiguity являются подтверждёнными источниками; TASK-012
  адресует их без metrics tooling.

Результат review — текущий bounded process hardening; отдельная analytics
system не требуется.

## Publication

Не авторизована. Commit отсутствует; Publisher P0–P10 не запускается.

## Independent Final Review

**Approved — 0 blocking and 0 nonblocking findings.**

- Coordinator нормативно владеет Existing Coverage pre-test gate.
- Все пять изменённых role contracts входят в declared scope и Size Guard.
- Removable-question: ни один из пяти Scope Audit classes нельзя удалить с
  сохранением Definition of Done; forbidden scope отсутствует.
- Exact diff содержит 16 Required, 0 Questionable, 0 Removable files.
- Production Activation не анализировалась и не изменялась.

## Next Candidate

Не активирован. После closure может быть повторно рассмотрена отдельная
documentation-first readiness/design task Production Activation, но эта task
не выполняет её анализ.

## Closure

- Final status: Completed — Coordinator Accepted.
- Tester: PASS, 0 blocking и 0 nonblocking findings.
- Scope Audit: Accepted — 16 Required, 0 Questionable, 0 Removable.
- Final Reviewer: Approved, 0 blocking и 0 nonblocking findings.
- Commit/push/PR/publication: not performed, not authorized.
- Closed by: Coordinator.
- Date: 2026-07-29.

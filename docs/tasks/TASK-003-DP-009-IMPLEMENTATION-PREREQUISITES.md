# TASK-003 — Уточнение implementation prerequisites Draft DP-009

## Status

**Completed — Coordinator Accepted**

## Objective

Уточнить в зеркальных EN/RU документах Draft DP-009 минимальный,
однозначный и независимо проверяемый контракт implementation prerequisites:
конкретный вход Runtime Bootstrap, dependency bindings и представление
failures без переноса operational startup responsibility из Runtime Host.

Результат этой documentation-only task должен сделать последующую отдельную
implementation task планируемой без скрытых архитектурных решений. Design
Status и Implementation Status DP-009 не повышаются.

## Selection Evidence

- Autonomous entry: точная bare-команда `Продолжай проект.`.
- Active tasks: отсутствуют; TASK-002 закрыта со статусом
  `Completed — Coordinator Accepted`.
- Явный next candidate: closure TASK-002 и project-state documents рекомендуют
  focused refinement implementation prerequisites Draft DP-009 после clean
  trusted Stage 1 baseline.
- Current milestone: **Beta — Complete the Single-Node Runtime**.
- Dependency evidence: Loader DP-007 и Snapshot Builder DP-008 реализованы
  изолированно, а production Loader-to-Builder-to-Launcher pipeline остаётся
  ранним незакрытым Beta architectural gap.
- Readiness evidence: Draft DP-009 §23 прямо требует до implementation
  определить concrete Bootstrap input, dependency bindings и failure
  representation. Поэтому DP-009 implementation ещё не является Ready.
- Ranking:
  1. работа относится к dependency текущей Beta milestone;
  2. закрывает prerequisite до любой реализации DP-009;
  3. является наименьшим независимо проверяемым architecture/documentation
     slice;
  4. не требует изменения Approved/Frozen architecture или продуктовой
     приоритизации.
- Отклонённые alternatives:
  - реализация DP-009 — не Ready до фиксации §23;
  - Runtime Launcher, Runtime Lifecycle Owner и production launch pipeline —
    последующие отдельные задачи с более широкими lifecycle и ownership
    последствиями;
  - Listener/TLS, Metrics, operational diagnostics, Delivery, Persistence и
    Plugin contracts — иные Beta epics; их выбор сейчас нарушил бы
    prerequisite order либо расширил scope;
  - operational diagnostics/redaction, retry, replacement, reconciliation,
    persistence, Secret resolution timing, process topology и remote launch
    transport — явно отложены DP-009 §22.

## Scope

- архитектурный анализ соответствия Draft DP-009 действующим ADR, ARCH,
  DP-007, DP-008 и фактическим Runtime boundaries;
- уточнение только зеркальных:
  - `docs/en/design/DP-009-runtime-bootstrap-contract.md`;
  - `docs/ru/design/DP-009-runtime-bootstrap-contract.md`;
- определение конкретного Bootstrap input, dependency bindings и failure
  representation, достаточных для будущей реализации;
- сохранение Host как единственного владельца operational startup;
- синхронизация только действительно затронутых task, navigation и
  project-state documents по PROCESS-002;
- независимая documentation verification, review и полный scope audit.

## Non-Goals

- production code, tests, package layout или конкретные Go interfaces;
- реализация DP-009, Runtime Bootstrap, Runtime Launcher, Runtime Lifecycle
  Owner либо production Loader-to-Builder-to-Launcher pipeline;
- повышение Design Status или Implementation Status DP-009;
- изменение Approved ADR, Active/Frozen ARCH, DP-007 или DP-008;
- operational diagnostics/redaction, retry, replacement, reconciliation,
  persistence, Secret resolution timing, process topology или remote launch
  transport;
- Listener/TLS, Metrics, Delivery, Persistence, Plugin contracts и другие
  Beta epics;
- следующая implementation task: она не создаётся и не начинается
  автоматически;
- unrelated refactoring, formatting-only changes, scripts, generated files
  или automation tooling.

## Sources of Truth

- [ADR-0002 RU](../ru/adr/0002-configuration-dsl.md) и
  [EN mirror](../en/adr/0002-configuration-dsl.md);
- [ADR-0003 RU](../ru/adr/0003-runtime-architecture.md) и
  [EN mirror](../en/adr/0003-runtime-architecture.md);
- [ARCH-002 RU](../ru/architecture/ARCH-002-runtime-foundation-freeze.md) и
  [EN mirror](../en/architecture/ARCH-002-runtime-foundation-freeze.md);
- [ARCH-004 RU](../ru/architecture/ARCH-004-runtime-deployment-and-identity-model.md)
  и
  [EN mirror](../en/architecture/ARCH-004-runtime-deployment-and-identity-model.md);
- [ARCH-005 RU](../ru/architecture/ARCH-005-runtime-configuration-snapshot-and-loading-model.md)
  и
  [EN mirror](../en/architecture/ARCH-005-runtime-configuration-snapshot-and-loading-model.md);
- [Draft DP-007 RU](../ru/design/DP-007-configuration-loader-contract.md) и
  [EN mirror](../en/design/DP-007-configuration-loader-contract.md);
- [Draft DP-008 RU](../ru/design/DP-008-snapshot-builder-contract.md) и
  [EN mirror](../en/design/DP-008-snapshot-builder-contract.md);
- [Draft DP-009 RU](../ru/design/DP-009-runtime-bootstrap-contract.md) и
  [EN mirror](../en/design/DP-009-runtime-bootstrap-contract.md);
- [MASTER_PLAN RU](../ru/roadmap/MASTER_PLAN.md) и
  [EN mirror](../en/roadmap/MASTER_PLAN.md);
- [current state](../../spec/current-state.md);
- [decisions](../../spec/decisions.md);
- [TASK-002](TASK-002-AUTONOMOUS-PROJECT-CONTINUATION.md).

## Roles

- **Coordinator:** выполнил autonomous preflight, deterministic selection,
  final scope audit, acceptance, project-state update и next recommendation.
- **Documentation Agent:** создал task record как первый content change,
  обновил navigation вторым change, зеркально зафиксировал утверждённое
  Architect уточнение Draft DP-009 и выполнил финальную синхронизацию
  PROCESS-002.
- **Architect:** подтвердил совместимость с authoritative boundaries и
  передал concrete implementation prerequisites Draft DP-009; verdict
  `Ready`, blocker отсутствует.
- **Developer:** не применяется, поскольку production code и tests запрещены
  scope этой architecture/documentation task.
- **Tester:** initial verdict `FAIL` — два mirrored clarity findings:
  преждевременная normative Go/package ссылка Snapshot и смешение static
  validations с execution failure points constructor/Build. Documentation
  rework выполнен; итоговый independent retest verdict — `PASS`.
- **Reviewer:** независимо сверил task, authoritative architecture, DP-009
  refinement, verification evidence и scope; итоговый verdict — `Approved`.

## Branch

- **Исходный trusted baseline:** clean `main`, commit `ef79d3e`.
- **Task branch:** `docs/task-003-dp009-prerequisites-refinement`.
- **Branch action:** Coordinator создал и переключил локальную
  documentation-only task branch до content changes.
- **Первый content change:** создание этого task record; обновление task index
  и любая архитектурная работа выполняются только после него.
- **Bare-command authority:** точная команда `Продолжай проект.` разрешила
  read-only intake, deterministic selection, создание task record и безопасной
  локальной task branch, но не разрешила stage, commit, push, merge, branch
  deletion, fetch, pull, remote mutation или изменение `main`.
- **Запрещённые git actions:** stage, commit, push, merge, удаление ветки,
  fetch, pull, изменение remote и изменение `main` без отдельного разрешения
  пользователя.

## Constraints

- Репозиторий является единственным источником истины.
- Не менять Approved/Frozen architecture.
- Не принимать архитектурные решения ролью Documentation Agent.
- Сохранить DP-009 как Draft и его implementation как Planned.
- Не переносить в Bootstrap validation/composition/acquisition/startup,
  rollback, readiness, admission или lifecycle ownership Runtime Host.
- Не давать Bootstrap полномочия Loader, Builder, Repository, publication,
  Runtime Lifecycle Owner или Control Plane.
- EN/RU DP-009 должны оставаться структурно и семантически зеркальными.
- Planned contract нельзя описывать как implemented behavior.
- Commit, push и merge запрещены без отдельного разрешения пользователя.

## Stop Conditions

- authoritative documents противоречат друг другу;
- уточнение требует изменения Approved ADR либо Active/Frozen ARCH;
- concrete input, bindings или failure representation требуют отсутствующего
  продуктового решения;
- границы нельзя определить без реализации DP-009, Launcher, Lifecycle Owner
  или production pipeline;
- возникает scope expansion в §22 deferred questions или другие Beta epics;
- обнаружен критический documentation drift;
- baseline становится dirty неатрибутированными изменениями;
- обязательная проверка или независимый review возвращает blocking finding.

## Acceptance Criteria

1. Task-before-work invariant доказан: этот task record является первым
   content change на task branch, а task index изменён только после него.
2. Architect подтвердил совместимость уточнения с authoritative ADR, ARCH,
   DP-007, DP-008 и существующими Runtime boundaries либо вернул явный blocker.
3. EN/RU Draft DP-009 зеркально определяют concrete Bootstrap input без
   полномочий Loader, Builder, Repository, publication или Lifecycle Owner.
4. EN/RU Draft DP-009 зеркально определяют минимальные dependency bindings,
   их ownership и передачу в Host без service locator, global registry или
   скрытого operational work.
5. EN/RU Draft DP-009 зеркально определяют mutually exclusive success,
   Bootstrap Failure и Startup Failure representation, сохраняя failure
   boundary и rollback responsibility Host.
6. Уточнение не меняет Design Status и Implementation Status DP-009 и не
   начинает implementation, Launcher, Lifecycle Owner или production
   pipeline.
7. PROCESS-002 отражает только фактически уточнённый planned contract и
   обновляет лишь необходимые navigation/project-state documents.
8. Documentation verification и независимый review проходят без blocking
   findings.
9. Scope audit классифицирует каждый изменённый файл и подтверждает отсутствие
   premature next-task/pipeline work, unrelated, generated и formatting-only
   changes.

## Verification

- проверить filesystem chronology task record и последующего index change;
- проверить links для всех repository Markdown files;
- проверить Markdown fences, conflict markers и trailing whitespace;
- проверить зеркальную структуру, headings, statuses и normative meaning
  DP-009 EN/RU;
- проверить точное сохранение Design Status `Draft` и Implementation Status
  `Planned`;
- проверить diff на отсутствие production code, tests, ADR, ARCH, DP-007,
  DP-008, generated и unrelated files;
- выполнить `git diff --check`;
- получить независимый Tester verdict;
- получить независимый Reviewer verdict после финальных documentation changes
  и scope audit;
- Go formatter, tests, race и vet неприменимы, пока diff остаётся
  documentation-only.

## PROCESS-002 Applicability

| Документ | Решение | Обоснование |
| --- | --- | --- |
| `AGENTS.md`, `docs/engineering/AGENT.md`, PROCESS-001/002 и role contracts | Не изменять | Task применяет существующий процесс и границы ролей, но не меняет их |
| `docs/tasks/README.md` | Изменён | TASK-003 требует navigation entry |
| `.ai/PROJECT_CONTEXT.md` | Изменён | Нужен factual active-task и gate state для repository-first continuation |
| `spec/current-state.md` | Изменён | Нужен factual operational state; product capability не заявляется |
| `spec/decisions.md` | Не изменять | Новый Approved decision и status transition отсутствуют; DP-009 остаётся Draft |
| MASTER_PLAN EN/RU | Не изменять | Milestone, dependency ordering и architectural debt не изменились; implementation не начата |
| `CHANGELOG.md` | Не изменять | Реализованная capability, release behavior и public API не изменились |
| root/project README и public documentation indices | Не изменять | DP-009 уже доступен через существующую navigation structure; новый public document не создан |

DP-009 mirrors проверены после Reviewer approval; реального contract drift,
требующего дополнительного изменения на final synchronization, не обнаружено.

## Scope Audit

Для каждого изменённого production, test, documentation и generated-файла:

- классификация: `Required`, `Questionable` или `Removable`;
- связь с acceptance criteria;
- меняет ли файл planned architecture, factual state или только navigation;
- disposition для `Questionable` и `Removable`.

Отдельно проверить:

- не началась ли реализация DP-009, Runtime Launcher, Runtime Lifecycle Owner
  либо production launch pipeline;
- не изменены ли Approved ADR, Active/Frozen ARCH, DP-007 или DP-008;
- не затронуты ли иные Beta epics;
- не описано ли planned поведение как implemented;
- нет ли generated, formatting-only, accidental или unrelated changes.

### Final Scope Assessment

Reviewer предложил, а Coordinator подтвердил следующую классификацию:

| Файл | Классификация | Связь со scope | Тип изменения |
| --- | --- | --- | --- |
| `.ai/PROJECT_CONTEXT.md` | `Required` | AC-007: factual active-task и gate state | Project-state synchronization |
| `docs/en/design/DP-009-runtime-bootstrap-contract.md` | `Required` | AC-003–AC-006: planned Bootstrap contract | Planned architecture refinement |
| `docs/ru/design/DP-009-runtime-bootstrap-contract.md` | `Required` | AC-003–AC-006 и EN/RU parity | Planned architecture refinement |
| `docs/tasks/README.md` | `Required` | AC-007: navigation к TASK-003 | Navigation only |
| `docs/tasks/TASK-003-DP-009-IMPLEMENTATION-PREREQUISITES.md` | `Required` | AC-001, AC-002, AC-007–AC-009 | Operational task evidence |
| `spec/current-state.md` | `Required` | AC-007: factual active-task и gate state | Project-state synchronization |

Итого: **6 `Required`, 0 `Questionable`, 0 `Removable`**. Final scope audit
принят Coordinator.

Production code, tests, ADR, ARCH, DP-007, DP-008, MASTER_PLAN, CHANGELOG,
generated и formatting-only files отсутствуют. Реализация DP-009, Runtime
Launcher, Runtime Lifecycle Owner и production pipeline не началась.

## Handoff

- **Architect verdict:** `Ready`; противоречия с ADR-0002, ADR-0003, Frozen
  ARCH-002, ARCH-004, ARCH-005, DP-007 или DP-008 не обнаружены. Новый ADR,
  ARCH и изменение статуса не требуются.
- **Architecture boundaries:** concrete request содержит Snapshot by value,
  обязательный startup context и fixed typed bindings без дублирования
  identity; Bootstrap выполняет только static validation, конструирует и
  строит не более одного Host и вызывает Start не более одного раза; Host
  сохраняет ownership operational startup и rollback.
- **Failure contract:** зафиксированы exclusive Success, Bootstrap Failure и
  Startup Failure; пять ordered Stage/Code pairs, cause preservation и запрет
  post-Start cleanup, Stop, retry и reclassification.
- **Documentation scope:** зеркально изменены Draft DP-009 EN/RU; Design Status
  `Draft` и Implementation Status `Planned` сохранены. Task record и index
  фиксируют task identity, evidence и navigation; `.ai/PROJECT_CONTEXT.md` и
  `spec/current-state.md` отражают factual closure, Tester `PASS`, Reviewer
  final closure `Approved` и Coordinator acceptance.
- **Tester rework:** из §8 удалена normative ссылка на concrete
  `runtimeconfig.Snapshot`; §10 теперь содержит ровно три static validation
  items и отдельно определяет constructor и Build как реальные четвёртую и
  пятую execution failure points общего fail-fast precedence. Порядок
  validate request → validate bindings → construct → Build → Start сохранён
  зеркально.
- **Task-before-work evidence:** до изменения index `git status --short`
  показывал только новый TASK-003; filesystem CreationTime task record —
  `2026-07-26T20:43:25.4111467Z`, LastWriteTime index после второго change —
  `2026-07-26T20:43:40.5679903Z`.
- **Verification:** initial Tester verdict `FAIL` с двумя clarity findings;
  ограниченный mirrored rework выполнен, итоговый retest verdict — `PASS`;
  Reviewer final closure verdict — `Approved`. Targeted проверки затрагиваемых
  пакетов `internal/runtime` и `internal/runtimeconfig` — PASS.
  Post-rework проверки premature package references, EN/RU numbered headings
  и AP identifiers, Draft/Planned statuses, пять Stage/Code pairs, links для
  99 Markdown files с hidden/untracked, fences, conflict markers, changed-file
  trailing whitespace, `git diff --check` и six-file scope guard — PASS.
  Repository-wide trailing-whitespace scan повторно обнаруживает 76
  существующих строк только в неизменённых EN/RU
  `runtime-alpha-review.md`; этот pre-existing drift не создан TASK-003.
- **Scope audit:** Coordinator accepted — 6 `Required`, 0 `Questionable`, 0
  `Removable`.
- **Открытые findings и risks:** blocking findings отсутствуют; implementation
  и production integration не начаты; AP-003 и AP-011 остаются
  integration-gated.
- **Следующий разрешённый шаг:** только отдельно разрешённый commit закрытой
  TASK-003; push и merge требуют отдельных разрешений.

## Next Candidate

- **Рекомендуемая Ready work после clean trusted baseline:** изолированная
  реализация уточнённого Runtime Bootstrap contract DP-009.
- **Readiness evidence:** implementation prerequisites §23 завершены и приняты;
  concrete request, fixed bindings, ordered failure contract и ownership
  boundaries определены; Architect verdict `Ready`, Tester `PASS`, Reviewer
  `Approved`, Coordinator acceptance получен.
- **Обязательные boundaries следующей task:** только isolated Bootstrap;
  Runtime Launcher, Runtime Lifecycle Owner, Control Service wiring и
  production Loader-to-Builder-to-Launcher pipeline не входят в scope.
  Acceptance proofs AP-003 и AP-011 остаются отложенными до отдельной
  production integration.
- **Явно не начата:** следующая task и branch не созданы; implementation
  DP-009, Runtime Launcher, Runtime Lifecycle Owner и production pipeline
  отсутствуют.

## Closure

- **Final status:** Completed — Coordinator Accepted.
- **Architecture:** Draft DP-009 implementation-prerequisites refinement
  завершён; Design Status `Draft`, Implementation Status `Planned`.
- **Verification:** Tester `PASS` after rework; Reviewer final closure
  `Approved`; targeted Runtime/runtimeconfig и documentation checks — PASS.
- **Scope audit:** Accepted — 6 `Required`, 0 `Questionable`, 0 `Removable`.
- **Stage 2 proof:** task record создан первым content change, task index
  изменён вторым; task-before-work invariant practically verified.
- **Remaining gaps:** DP-009 implementation и production pipeline не начаты;
  AP-003 и AP-011 integration-gated.
- **Git authority:** commit, push и merge не разрешены этой closure.
- **Closed by:** Coordinator.
- **Date:** 2026-07-27.

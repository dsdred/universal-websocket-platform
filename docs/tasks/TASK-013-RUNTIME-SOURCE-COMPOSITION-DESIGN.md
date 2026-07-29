# TASK-013 — Runtime Source Composition Design

## Status

**Completed — Coordinator Accepted**

## Task Contract

### Task Mode

**Design-only.**

Задача фиксирует минимальный concrete Source composition contract, который
адаптирует существующие in-memory Configuration и ConfigurationVersion
repository boundaries к `configurationloader.Source` и
`configurationloader.Loader` и позволяет будущей management composition
сконструировать `Loader` и `runtimelaunchflow.Flow`. Production-код и tests в
этой task не меняются; Production Activation не запускается.

### Why Now

- TASK-011 завершила изолированную реализацию Runtime Launch Flow, а TASK-012
  завершила hardening инженерного процесса.
- Draft DP-011 требует отдельной concrete Source composition до Production
  Activation; текущий `configurationloader.Source` существует только как
  boundary и не имеет repository-backed composition contract.
- Source composition предшествует management routing: routing не может
  корректно сконструировать `Loader`/`Flow`, пока не определены dependency,
  ownership, lookup и failure semantics Source.
- Source composition предшествует persistence boundary: минимальный slice
  использует уже существующие in-memory repositories, тогда как durable
  persistence затрагивает более широкую модель хранения и lifecycle.
- Scope целен: один Source composition contract, один зеркальный Design
  Proposal и ноль production changes.

### Definition of Done

1. Создан зеркальный `DP-012` со статусами `Design Status: Draft` и
   `Implementation Status: Planned`.
2. DP-012 однозначно определяет владельца Source adapter и границу его
   ответственности относительно repositories, Loader, Flow и будущей
   management composition.
3. Зафиксирован exact lookup только по запрошенной цепочке
   `Workspace ID + Configuration ID + ConfigurationVersion ID`; latest,
   current, replacement и `GetPublished(configurationID)` fallback запрещены.
4. Зафиксирована проверка parent identity и `Published` state exact найденной
   версии без выбора другой версии.
5. Зафиксирована detached `SourceObservation` и отсутствие передачи
   repository-owned mutable state через Source boundary.
6. Зафиксирована полная и непротиворечивая mapping-таблица repository/source
   failures в существующие Loader source errors без transport diagnostics.
7. Зафиксированы constructor shape, dependency direction и место будущего
   composition, позволяющие создать `configurationloader.Loader`, а затем
   `runtimelaunchflow.Flow`, не активируя Runtime.
8. Зафиксированы concurrency и lifetime rules для adapter и borrowed
   repositories без нового cache, retry, lock ownership или background work.
9. Явно запрещены fallback/latest selection, retry, cache, persistence,
   management API/routing, Production Activation и изменение repository
   contract в рамках этой task.
10. Зеркала DP-012 имеют reciprocal language links, одинаковые структуру,
    статусы и нормативный смысл; design indexes, task index, current state,
    MASTER_PLAN и PROJECT_CONTEXT синхронизированы в этой task.
11. Documentation verification, независимые Tester и Reviewer passes, Scope
    Audit и Coordinator Acceptance завершены без blocking findings.

### Out of Scope

- Любой production-код или executable tests.
- Redesign или расширение Configuration/ConfigurationVersion repositories.
- HTTP endpoints, management routing, authentication или authorization.
- Runtime Instance и Launch Attempt persistence.
- Любая durable persistence technology или migration.
- Production Activation, Runtime start/stop wiring и Control Service launch.
- Recovery, reconciliation, retry, supervision и caching.
- Diagnostics transport или новый public error protocol.
- Commit, push, PR, merge, publication и удаление веток.

### Verification Plan

- Проверить normative consistency DP-012 с ADR-0002/0003, ARCH-004/005 и
  DP-007–DP-011.
- Сопоставить каждый declaration и invariant с фактическими
  `configuration`, `configurationversion`, `runtimeconfigload`,
  `configurationloader` и `runtimelaunchflow` boundaries, не меняя код.
- Проверить exact lookup, Published validation, detachment, error mapping,
  constructor/dependency direction и concurrency/lifetime через reviewable
  contract tables и сценарии.
- Проверить EN/RU headings/fences/status parity, relative links,
  conflict markers, trailing whitespace и `git diff --check`.
- Подтвердить exact file set: не более 10 documentation files, 0 production
  files, 0 test files.
- Выполнить независимую documentation verification и независимый Final
  Review; blocking finding направить в bounded Rework Loop.

## Objective

Определить минимальный concrete repository-backed Source composition contract,
который выдаёт Loader одну exact detached Published ConfigurationVersion и
оставляет будущей management composition только construction
`Source -> Loader -> Flow`, без Runtime activation и без расширения
persistence или management scope.

## Selection Evidence

- `.ai/PROJECT_CONTEXT.md`, `spec/current-state.md`, `spec/decisions.md` и
  зеркальные MASTER_PLAN фиксируют отсутствие concrete Source composition и
  Production Activation после TASK-011.
- TASK-011 подтверждает, что изолированный Flow уже принимает готовый
  `*configurationloader.Loader`; следовательно, Source composition является
  ближайшим отсутствующим prerequisite, а не частью Flow.
- `internal/configurationloader` определяет `Source.LoadExact` и сохраняет
  точные source error identities; существующие in-memory repositories
  предоставляют factual read boundaries, но composition contract не
  документирован.
- Candidate удовлетворяет readiness/ranking PROCESS-001: prerequisites
  реализованы, это наименьший независимо проверяемый slice, и он не требует
  изменения Approved/Frozen источника.
- Отклонена **management routing**: она зависит от способа construction
  Source/Loader/Flow и преждевременно добавила бы HTTP/auth/lifecycle scope.
- Отклонена **persistence boundary**: она шире минимального in-memory adapter,
  требует отдельного storage/lifecycle решения и не нужна для доказательства
  Source-to-Loader composition.
- Отклонена **Production Activation**: DP-011 требует сначала concrete Source
  composition; activation смешала бы design, routing и runtime side effects.

## Scope

Разрешены только:

- этот task record и task index;
- зеркальные
  `docs/en/design/DP-012-runtime-source-composition.md` и
  `docs/ru/design/DP-012-runtime-source-composition.md`;
- зеркальные `docs/en/design/README.md` и `docs/ru/design/README.md`;
- `spec/current-state.md`;
- зеркальные `docs/en/roadmap/MASTER_PLAN.md` и
  `docs/ru/roadmap/MASTER_PLAN.md`;
- `.ai/PROJECT_CONTEXT.md`.

Exact scope — 10 documentation files; production files и test files — 0.

## Non-Goals

- Реализация approved Source adapter является следующей возможной task и не
  начинается автоматически.
- Конструирование management route, Handler или application composition root.
- Запуск `Flow.Start`, создание Runtime Instance/Launch Attempt или Host.
- Speculative abstraction для будущих persistent sources.
- Unrelated refactoring, formatting или status promotion DP-007–DP-011.

## Sources of Truth

- Approved ADR-0002 и ADR-0003, зеркально:
  `docs/en/adr/0002-configuration-dsl.md`,
  `docs/ru/adr/0002-configuration-dsl.md`,
  `docs/en/adr/0003-runtime-architecture.md`,
  `docs/ru/adr/0003-runtime-architecture.md`;
- Active ARCH-004 и ARCH-005, зеркально;
- Draft DP-007–DP-011 как применимые implementation contracts без повышения
  их статуса;
- factual implementation evidence:
  `internal/configuration`,
  `internal/configurationversion`,
  `internal/runtimeconfigload`,
  `internal/configurationloader`,
  `internal/runtimelaunchflow`;
- TASK-011 и TASK-012;
- `.ai/PROJECT_CONTEXT.md`, `spec/current-state.md`,
  `spec/decisions.md` и зеркальные MASTER_PLAN;
- PROCESS-001, PROCESS-002 и task template.

## Roles

- **Coordinator:** orchestration, contract, handoffs, Scope Audit, acceptance
  и project-state synchronization.
- **Architect:** определяет Source composition contract, ownership,
  validation, failure, dependency и lifetime boundaries.
- **Documentation Agent:** фиксирует только переданное Architect решение в
  зеркальном Draft DP-012 и синхронизирует разрешённые документы.
- **Developer:** неприменим; production-код и tests запрещены.
- **Tester:** независимая documentation verification по Verification Plan.
- **Reviewer:** независимый architecture/documentation review и
  removable-question.
- **Publisher:** не авторизован; publication не выполняется.

## Branch

- исходный trusted baseline:
  clean synchronized `main@8666eea3ae2a546047e55f283c58acd7c8a68804`;
- task branch: `docs/task-013-runtime-source-composition-design`;
- branch action: ветка уже создана Coordinator до content changes;
- этот task record является первым content change;
- stage, commit, push, PR, merge, fetch/pull, rebase, reset и branch deletion
  запрещены.

## Constraints

- Documentation Agent не выбирает архитектурную форму adapter и не разрешает
  неясности самостоятельно; точный design передаёт Architect.
- Repository остаётся factual dependency, но не становится dependency Runtime
  или Flow; Loader продолжает зависеть только от своего `Source` boundary.
- Exact pinned identity не может быть заменена latest/current/published
  selection по одному Configuration ID.
- Planned DP-012 не должен описываться как implemented capability.
- Нельзя вводить fallback, retry, cache, persistence, background work,
  management API или Production Activation.
- Commit и publication не разрешены.

## Stop Conditions

- Architect не может однозначно определить ownership, constructor/dependency
  direction или границу между Source, Loader и future composition root.
- Текущие repository reads не позволяют однозначно сформировать consistent
  observation exact parent/version chain без repository redesign.
- Schema facts, parent Configuration lookup, detachment либо mapping ошибок
  требуют решения за пределами одного Source composition contract.
- Draft DP-012 конфликтует с Approved ADR, Active ARCH или более приоритетным
  source.
- Требуется изменение production code, tests, public API, repository contract
  или более 10 documentation files.
- Management routing, persistence или Production Activation входят в diff.
- EN/RU normative parity невозможно сохранить либо обязательная verification
  завершается ошибкой.
- Blocking Tester/Reviewer finding не устранён.
- Worktree становится dirty неатрибутированными изменениями или baseline
  оказывается diverged.

## Acceptance Criteria

1. Architect handoff закрывает все пункты Definition of Done 2–9 без
   предположений Documentation Agent.
2. DP-012 явно отделяет Source adapter ownership от Loader validation и future
   management composition.
3. Exact identity, Published eligibility, detachment и error mapping выражены
   исчерпывающими tables/scenarios без fallback.
4. Constructor/dependency graph допускает construction Loader/Flow, но не
   выполняет Production Activation.
5. Concurrency/lifetime contract не добавляет shared mutable state, retry,
   cache или фоновые операции.
6. DP-012 EN/RU parity, links, project-state applicability и exact scope
   подтверждены воспроизводимыми проверками.
7. Independent Tester — PASS, Final Reviewer — Approved, blocking findings 0.

**Acceptance evidence:** mirrored Draft/Planned DP-012 закрывает DoD 1–9;
reciprocal links и design/task indexes закрывают DoD 10; final Tester PASS,
Final Reviewer Approved 0/0, accepted Scope Audit 10/0/0 и checks из
Verification закрывают DoD 11.

## Existing Coverage Report

### Existing Coverage

- DP-007 и `configurationloader.Loader` определяют neutral `Source.LoadExact`,
  completeness/identity/Published validation и closed source failures.
- `runtimeconfigload` создаёт detached immutable Loader-to-Builder result.
- In-memory Configuration и ConfigurationVersion repositories применяют
  concurrent-safe reads и возвращают detached entity copies.
- DP-011 и `runtimelaunchflow.Flow` принимают готовый Loader и не владеют
  Source selection.

### Coverage Gap

- Concrete repository-backed Source owner и constructor не определены.
- Не закреплены exact parent/version read sequence, consistency boundary,
  schema fact source и repository-to-source error mapping.
- Не закреплено место будущего construction `Source -> Loader -> Flow`.
- Нет зеркального DP, отделяющего этот Planned contract от management routing,
  persistence и Production Activation.

### Added Proof Tests

`Not applicable` на этапе task intake: production/executable tests запрещены.
DP-012 должен определить будущие proof scenarios для отдельной implementation
task.

### Added Regression Tests

`Not applicable` на этапе task intake: код и tests не меняются. Documentation
checks должны доказать parity, links, exact scope и отсутствие forbidden
capability claims.

### Remaining Limitations

- Design-only review не доказывает executable adapter behavior.
- Consistent observation зависит от обязательного composition confinement;
  если future composition не докажет shared repositories, service-only
  mutation path и отсутствие alternate writers, activation блокируется.
- Production composition остаётся отдельной integration work.

## Architecture Confirmation

**Architect verdict: Ready with mandatory confinement.**

- Утверждён package `internal/configurationloadsource` и concrete
  `MemorySource`, borrowing ровно два существующих in-memory repositories.
- Утверждены constructor без side effects и существующий
  `configurationloader.Source.LoadExact` surface без новых interfaces/errors.
- Утверждены Version-first exact lookup, linearization point L, subsequent
  parent read C, Published/schema/identity validation, deep detachment и
  closed failure mapping.
- Consistent observation допустим только при audited single-instance mutation
  topology: одна repository pair, ровно по одному Service каждого типа,
  handlers получают только эти instances, post-bootstrap mutations только
  через них, direct/alternate writers отсутствуют. Если pre-construction Audit
  не доказан, Source не создаётся и activation блокируется.
- MemorySource не синтезирует `ErrInconsistentSourceObservation`; generic
  Loader error остаётся доступным другим adapters. Violation topology после
  construction является programming/composition contract violation, а не
  runtime detector outcome.
- Management routing, persistence и Production Activation не разрешены.

## Size Guard

- Actual: 10 documentation files, 0 production lines, 0 test lines,
  0 new packages и 0 independently shipped behavior.
- Size Guard не сработал: diff остаётся ниже порога >15 files; единственный
  Draft architecture contract не создаёт implemented capability.
- При превышении 10 documentation files или необходимости второго
  architecture contract Coordinator обязан остановить работу и split/reassess
  scope до дальнейших changes.

## Verification Matrix

| Risk class | Applicability | Required evidence |
|---|---|---|
| Concurrency/lifecycle/shared state | Design применим | Architect contract и independent review lifetime/borrowed repository rules; race не запускается без кода |
| API/CLI/UI/config/production wiring | Composition design применим | Contract scenarios доказывают construction без activation; executable smoke не применяется |
| Dependencies | Documentation-only | `go.mod`/`go.sum` и production imports отсутствуют в diff |
| Public API | Не применяется | Exported Go declarations отсутствуют |
| Documentation | Применяется | EN/RU parity, links, status/meaning, contradictions, whitespace и exact diff |

## Verification

- Initial Tester: **FAIL**, blocking findings B-001/B-002.
- B-001/B-002 rework уточнил single-instance mutation topology и Composition
  Audit и удалил ошибочный runtime
  `ErrInconsistentSourceObservation` outcome MemorySource.
- Repeat Tester: **FAIL**, blocking finding B-003.
- B-003 correction заменила несуществующую Source qualification на
  существующий `configurationloader.Source`.
- Final repeat Tester: **PASS**, 0 blocking и 0 nonblocking findings.
- Final Reviewer: **Approved**, 0 blocking и 0 nonblocking findings.
- Evidence:
  - exact changed files: 10 documentation, 0 production/test/generated;
  - DP-012 EN/RU headings: 29/29; numbered sections: 28/28; fences: 4/4;
  - broken relative links: 0;
  - obsolete package-qualified Source occurrences: 0;
  - conflict markers: 0; trailing whitespace: 0;
  - `git diff --check`: PASS, line-ending warnings only.

## Scope Audit

**Accepted — 10 Required, 0 Questionable, 0 Removable.**

- DP-012 EN/RU — 2 Required: DoD 1–9, exact mirrored design contract.
- Design indexes EN/RU — 2 Required: discoverability, reciprocal navigation и
  DoD 10.
- TASK-013 record — 1 Required: process contract, findings, verification и
  traceability.
- Task index — 1 Required: operational navigation.
- `spec/current-state.md` и `.ai/PROJECT_CONTEXT.md` — 2 Required: TASK-013
  как latest completed architecture task, отсутствие current architecture
  task, неактивированный next candidate и truthful Draft/Planned project
  state.
- MASTER_PLAN EN/RU — 2 Required: durable architectural-debt/prerequisite
  status и mirror parity.
- Production, tests, generated, premature next-task и unrelated changes
  отсутствуют.

Questionable и Removable отсутствуют. Final Reviewer подтвердил, что удаление
любого перечисленного class нарушит соответствующий DoD, navigation,
traceability либо truthful project state.

## Documentation Sync

- PROCESS-002 status: **Synchronized**.
- task record: **Completed — Coordinator Accepted**, создан первым content
  change;
- task index: обновлён ссылкой на TASK-013;
- `spec/current-state.md`: отражает TASK-013 как latest completed architecture
  task, отсутствие current task, неактивированный next candidate и
  Draft/Planned boundary без заявления новой capability;
- MASTER_PLAN EN/RU: синхронизированы prerequisite и Draft/Planned status без
  изменения milestone boundary;
- связанный Design Proposal: mirrored Draft DP-012 / Planned создан по
  Architect handoff;
- design indexes EN/RU: обновлены Draft/Planned entry и reciprocal
  discoverability;
- `.ai/PROJECT_CONTEXT.md`: отражает TASK-013 как latest completed
  architecture task, отсутствие current task, неактивированный next candidate
  и planned boundary;
- `spec/decisions.md`: **Not applicable**; перечень ожидающих решений уже
  содержит concrete Source composition и не требует нового ADR/status;
- root `README.md`: **Not applicable**; public capability не появилась;
- `CHANGELOG.md`: **Not applicable**; user-facing/release change отсутствует;
- parity, links и contradictions: Final Tester PASS; Final Reviewer Approved
  0/0.

## Commit Gate

- exact command `Разрешаю коммит.` получена: **нет**;
- commit message policy: не применялась, commit не авторизован;
- exact file set: только разрешённые documentation files;
- post-acceptance diff: bounded administrative closure sync в TASK-013 record,
  `.ai/PROJECT_CONTEXT.md` и `spec/current-state.md` после initial Final
  Reviewer; correction завершена; Repeat closure audit **Approved**, 0
  blocking и 0 nonblocking findings; DP и scope не изменены, exact set —
  10 documentation files;
- temporary/generated/unrelated files: отсутствуют;
- final checks: PASS.

## Process Health

- trigger применим: **нет**; TASK-012 уже выполнила review после десяти task,
  а rollback, escaped defect, repeated Publisher failure или >2 review returns
  для TASK-013 отсутствуют;
- bounded process changes: не планируются.

## Handoff

- Architecture Analysis: **Ready with mandatory confinement**; утверждённый
  revised Architect handoff полностью зафиксирован в mirrored DP-012.
- Initial Tester: **FAIL**, blocking findings B-001/B-002 — single-instance
  mutation topology и ошибочный runtime inconsistent-observation claim.
- Первый bounded rework B-001/B-002: **complete**.
- Repeat Tester: **FAIL**, blocking finding B-003 — task contract ошибочно
  приписывал Source package `runtimeconfigload`.
- Bounded correction B-003: **complete**.
- Final repeat Tester: **PASS**, 0 blocking и 0 nonblocking findings.
- Final Reviewer: **Approved**, 0 blocking и 0 nonblocking findings.
- Focused post-closure audit: **Needs Revision**, 1 blocking finding —
  stale active/current task wording и неточный post-acceptance diff record.
- Focused closure correction: **complete**.
- Repeat closure audit: **Approved**, 0 blocking и 0 nonblocking findings;
  DP и scope не изменены, exact set — 10 documentation files.
- Documentation change set accepted — ровно 10
  разрешённых files:
  - `.ai/PROJECT_CONTEXT.md`;
  - `docs/en/design/DP-012-runtime-source-composition.md`;
  - `docs/ru/design/DP-012-runtime-source-composition.md`;
  - `docs/en/design/README.md`;
  - `docs/ru/design/README.md`;
  - `docs/en/roadmap/MASTER_PLAN.md`;
  - `docs/ru/roadmap/MASTER_PLAN.md`;
  - `docs/tasks/README.md`;
  - `docs/tasks/TASK-013-RUNTIME-SOURCE-COMPOSITION-DESIGN.md`;
  - `spec/current-state.md`.
- Production code, tests, `spec/decisions.md`, root README и CHANGELOG не
  изменены.
- Scope Audit: Accepted — 10 Required, 0 Questionable, 0 Removable.
- PROCESS-002: Synchronized.
- Coordinator Acceptance: получена.

## Publication

Не авторизована. Commit отсутствует; Publisher P0–P10 не запускается.

## Next Candidate

- рекомендуемая Ready work: отдельная implementation task
  `MemorySource`, local proof tests и изолированный construction proof;
- readiness evidence: Coordinator-accepted design task и Draft DP-012 с
  Implementation Status Planned; новый intake всё ещё обязателен;
- явно не начата: да; implementation, management routing, persistence и
  Production Activation отсутствуют.

## Closure

- Final status: Completed — Coordinator Accepted.
- Design Status DP-012: Draft.
- Implementation Status DP-012: Planned; implementation отсутствует.
- Tester: PASS, 0 blocking и 0 nonblocking findings.
- Final Reviewer: Approved, 0 blocking и 0 nonblocking findings.
- Repeat closure audit: Approved, 0 blocking и 0 nonblocking findings; DP и
  scope unchanged, exact set — 10 documentation files.
- Scope Audit: Accepted — 10 Required, 0 Questionable, 0 Removable.
- PROCESS-002: Synchronized.
- Commit/push/PR/merge/publication: not performed, not authorized.
- Closed by: Coordinator.
- Date: 2026-07-29.

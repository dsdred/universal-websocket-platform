# TASK-016 — Runtime Operational Identity Persistence Design

## Status

`Completed — Coordinator Accepted`.

## Task Contract

### Task Mode

`Design-only`. Задача определяет один focused persistence contract для
operational identity Runtime Instance и Launch Attempt до любой реализации
management package, repository, schema или production wiring.

### Why Now

- TASK-015, `.ai/PROJECT_CONTEXT.md`, `spec/decisions.md` и зеркальные
  MASTER_PLAN независимо рекомендуют этот slice следующим;
- ARCH-004 §19(1) закрыт существующими Loader и Snapshot contracts, а §19(2)
  является первым остающимся prerequisite management implementation;
- DP-013 имеет Design READY/valid, но Implementation Readiness Blocked до
  focused designs §19(2)–(6);
- отдельный persistence contract является наименьшим независимо проверяемым
  architecture slice и не смешивает durable idempotency, activation,
  recovery или reporting.

### Definition of Done

1. Зеркальный Draft DP-014 EN/RU определяет ownership и durable boundaries
   Runtime Instance, Launch Attempt, desired/actual facts и opaque ID
   allocation без выбора storage technology или schema.
2. Контракт фиксирует atomicity, uniqueness, history, concurrency, failure и
   redaction invariants в соответствии с Active ARCH-004.
3. Planned design явно отделён от реализованного in-memory Lifecycle Owner и
   не объявляет persistence, repository, API, recovery или production wiring
   существующими.
4. Связанные индексы и durable project-state документы синхронизированы
   зеркально и без изменения статуса ARCH-004.
5. Documentation verification, PROCESS-002, Scope Audit и независимый final
   review завершены без blocking findings.

### Out of Scope

- production-код, тесты, package, repository или database schema;
- выбор PostgreSQL, SQL, ORM, migration framework или ID algorithm;
- durable management idempotency ARCH-004 §19(3);
- activation, replacement, rollback, recovery, reconciliation,
  diagnostics/redaction contracts §19(4)–(6);
- HTTP/API/DTO, authorization policy, Production Activation и Control Service
  wiring;
- deletion, retention, automatic restart, scheduling, clustering и process
  isolation.

### Verification Plan

- проверить status/source precedence Active ARCH-004, DP-010 и DP-013;
- сравнить EN/RU headings, normative meaning, links и code fences;
- проверить отсутствие schema/technology/API/implementation claims;
- проверить ссылки, conflict markers, trailing whitespace и
  `git diff --check`;
- существующие Go tests не изменяются; полный `go test ./...` и `go vet ./...`
  применимы как regression evidence только если final diff не затронет code;
- независимый Reviewer проверяет atomicity, ownership, failure boundaries,
  scope и возможность удалить каждый change без потери Definition of Done.

## Objective

Создать один проверяемый candidate design contract для prerequisite ARCH-004
§19(2): durable operational identity и history boundary Runtime
Instance/Launch Attempt. До отдельного approval/status decision Draft не
снимает formal implementation gate §19(2).

## Selection Evidence

- trusted baseline: clean synchronized `main` на merge commit `49c7973`;
- active `In Progress` или `Blocked` task отсутствует;
- explicit next candidate: TASK-015 `Next Candidate`;
- durable corroboration: `.ai/PROJECT_CONTEXT.md`, `spec/decisions.md`,
  `spec/current-state.md` и зеркальные MASTER_PLAN;
- readiness: architecture refinement разрешён PROCESS-001 до production
  implementation; prerequisites identity graph и ownership уже определены
  Active ARCH-004 и изолированным DP-010;
- ranking: current Beta dependency, первый unresolved prerequisite §19,
  smallest independently verifiable design, lowest unresolved risk, first
  authoritative ordering;
- отклонены:
  - DP-013 implementation — Blocked §19(2)–(6);
  - durable command idempotency — следующий §19(3), зависит от persistence
    transaction boundary;
  - activation/replacement/rollback — более поздний §19(4);
  - generic Persistence epic для Messages — materially different product
    capability, не prerequisite management routing.

## Scope

- новый зеркальный DP-014 EN/RU;
- bounded synchronization зеркального DP-013: candidate contract §19(2),
  сохранённый formal gate §19(2), downstream §19(3)–(6) и next design
  recommendation;
- зеркальные design indexes и MASTER_PLAN;
- task record и task index;
- применимые `.ai/PROJECT_CONTEXT.md`, `spec/current-state.md` и
  `spec/decisions.md`;
- только необходимые cross-links к существующим architecture/design sources.

## Non-Goals

- не начинать следующий prerequisite §19(3);
- не создавать executable persistence capability;
- не менять Active/Frozen architecture, public API или runtime behavior;
- не выполнять unrelated cleanup или formatting.

## Sources of Truth

- ADR-0002, ADR-0003 и ADR-0004;
- Frozen ARCH-002;
- Active ARCH-004 и ARCH-005;
- Approved DP-003, DP-004 и DP-005 в пределах их scope;
- Draft DP-007–DP-013 только как subordinate design и factual implementation
  boundary;
- `internal/runtimelifecycle` и его tests как evidence изолированного
  in-memory state, не как authority для нового persistence design;
- TASK-015 и durable project-state documents.

## Roles

- Coordinator: intake, role assignments, gates, Scope Audit и acceptance;
- Architect: persistence design, invariants, acceptance constraints;
- Documentation Agent: baseline, mirrored DP и project-state synchronization;
- Developer: Not applicable — production code запрещён;
- Tester: documentation verification и regression checks;
- Reviewer: независимый initial/final review; не автор изменений;
- Publisher: Not applicable — publication не разрешена этой командой.

## Branch

- исходный trusted baseline: `main` at `49c7973`, clean,
  `main == origin/main`;
- task branch:
  `docs/task-016-runtime-operational-identity-persistence-design`;
- branch action: создана и активирована до первого content change;
- запрещены stage, commit, push, merge, fetch, pull, rebase, reset и удаление
  веток без соответствующего explicit gate.

## Constraints

- ровно один новый architecture contract;
- Technology Neutrality и opaque identifiers;
- Runtime Host не читает Control Plane repositories и не владеет desired
  state;
- один Runtime Instance постоянно принадлежит одному Workspace и одной
  Configuration;
- не более одного active Launch Attempt, immutable exact Published
  ConfigurationVersion per attempt;
- historical attempt не активируется повторно;
- planned и implemented state разделяются;
- commit только после точной команды `Разрешаю коммит.`.

## Stop Conditions

- конфликт с Approved ADR или Active/Frozen ARCH;
- необходимость включить §19(3)–(6) в candidate contract §19(2);
- невозможность сохранить technology/schema neutrality;
- более одного нового architecture contract или >15 changed files;
- critical documentation drift;
- materially different Ready solution без authoritative ordering.

## Acceptance Criteria

1. Durable aggregate boundaries и ownership не создают второго Lifecycle
   Owner или скрытого service locator.
2. Runtime Instance creation/identity binding и Launch Attempt claim/history
   имеют однозначные atomicity и uniqueness invariants.
3. Desired/actual facts сохраняются правдиво без превращения persistence в
   authority над live Host resources.
4. Failure, retry и concurrency не допускают duplicate active attempt,
   identity reuse или partial history membership; retained active
   `AttemptStopping` при stop failure/cleanup-unproven сохраняется правдиво.
5. Recovery, command idempotency, activation и reporting остаются explicit
   deferred contracts, а не неявные promises.
6. EN/RU parity, navigation, applicability record и repository checks
   подтверждены evidence.

## Verification

- Existing Coverage Report:
  - Existing Coverage: ARCH-004 фиксирует identity graph, ownership,
    lifecycle и prerequisite §19(2); DP-010 и implementation доказывают только
    process-local Owner state/attempt allocation; TASK-015 фиксирует ordering;
  - Coverage Gap: durable aggregate, atomic persistence boundary, allocation,
    history and failure invariants ещё не определены;
  - Added Proof Tests: Not applicable до production implementation;
  - Added Regression Tests: Not applicable — code/test changes запрещены;
  - Remaining Limitations: executable persistence proofs появятся только в
    отдельной implementation task после закрытия всех применимых prerequisites.
- Verification Matrix:
  - concurrency/lifecycle/shared state: design invariants и independent review;
    race не применим без code changes;
  - API/CLI/UI/configuration/production wiring: Not applicable, запрещены;
  - dependencies: Not applicable, module files не меняются;
  - public API: Not applicable;
  - documentation: EN/RU parity, links, contradictions, whitespace и diff
    checks обязательны.
- Initial Tester: `PASS`;
- formatter/lint: `git diff --check` PASS;
- tests: preceding full `go test ./...` PASS reused after documentation-only
  rework; production code и tests не менялись;
- race: Not applicable — documentation-only task без concurrency code changes;
- vet: preceding full `go vet ./...` PASS reused after documentation-only
  rework;
- documentation structure:
  - exact changed file set: 13 documentation files;
  - DP-014 EN/RU: 27/27 headings including title, 4/4 code fences;
  - DP-013 EN/RU: 35/35 headings;
  - MASTER_PLAN EN/RU: 36/36 headings;
  - changed-scope relative links: 152 checked, 0 missing;
  - repository relative links: 753 checked, 0 missing;
- initial independent review: `Needs Revision`;
  - R-001: Draft преждевременно представлял §19(2) resolved;
  - R-002: atomicity wording запрещал допустимый retained active
    `AttemptStopping`;
  - R-003: append-only membership и mutable monotonic lifecycle facts не были
    разделены явно;
  - N-001: active stage project-state требовал синхронизации;
- Architect/Documentation rework: R-001/R-002/R-003/N-001 addressed;
- Repeat Tester: `PASS`;
- Repeat Reviewer: `Approved`, 0 blocking и 0 nonblocking findings.
- Final Reviewer: `Approved`, 0 blocking и 0 nonblocking findings;
- Coordinator Acceptance: granted for exact 13-file design diff.

## Documentation Baseline

- Status: `Drift Detected — non-critical`, не Blocked.
- Critical drift: 0.
- EN/RU baseline: design tree 14/14 filenames; ARCH-004 headings 29/29,
  DP-010 33/33, DP-013 35/35, design indexes 1/1, MASTER_PLAN 36/36.
- Links: 159 relative links checked, 0 missing.
- Status truth: ARCH-004 Active; DP-010 Draft/Implemented in isolation; DP-013
  Draft/Planned/Implementation Readiness Blocked; persistence отсутствует.
- Expected task-stage drift: task index и project-state ещё не отражают
  active TASK-016.
- Scope finding D-001: первоначальные 11 files не включали DP-013 EN/RU.
  После activation TASK-016 их “recommended, not activated” wording устарело.
  DP-014 добавляет candidate contract, но §19(2) остаётся unresolved formal
  gate до отдельного approval/status decision. Disposition: обе версии DP-013
  добавлены в Required scope; Design/Implementation status DP-013 не меняется.
- Exact expected final file set: 13 documentation files:
  `.ai/PROJECT_CONTEXT.md`, DP-014 EN/RU, DP-013 EN/RU, design indexes EN/RU,
  MASTER_PLAN EN/RU, task index, этот task record, `spec/current-state.md`,
  `spec/decisions.md`.

## Architecture Analysis

- Verdict: `Ready`; blockers 0.
- DP-014 предлагает отдельно проверяемый candidate contract ARCH-004 §19(2):
  durable aggregate и conditional commit semantics не определяют
  client-command deduplication, activation, restart reconciliation или
  reporting. Non-normative Draft не снимает formal gate §19(2) до отдельного
  approval/status decision.
- Aggregate root: Runtime Instance с immutable Workspace/Configuration/Instance
  binding, durable desired/actual facts, optional active-attempt link,
  monotonic revision и complete append-only membership attempt history.
- Launch Attempt: owned child, а не отдельный aggregate; durable key
  `(RuntimeInstanceID, LaunchAttemptID)`, exact immutable Published
  ConfigurationVersion pin и immutable parent/child identity. Membership
  append-only; phase/outcome facts conditionally и монотонно обновляются внутри
  того же child без regression или rewrite immutable facts. Attempt ID не
  переиспользуется в lifetime/history одного Instance; global uniqueness между
  Owners не требуется.
- Runtime Instance ID уникален в одном operational persistence/management
  domain, поскольку DP-013 сначала маршрутизирует по нему.
- Lifecycle Owner остаётся единственным lifecycle decision maker и владельцем
  live Host reference. Persistence только conditional commit boundary и не
  становится Owner, registry или service locator.
- Required atomic publications: initial Instance; one active Attempt claim;
  Running; Stop claim; phase-sensitive terminal result. Initial desired/actual
  равны `Stopped`. `Preparing` без созданного Host может завершиться
  stopped-before-running; `Launching`/`Running` сначала публикуют same-attempt
  Stop claim и actual `Stopping`. Active-attempt association очищается только
  при proof отсутствия Host resources; stop failure или unproven cleanup
  сохраняет active `AttemptStopping` association и правдивый failure/cleanup
  fact; это не terminal historical attempt. Partial aggregate и history
  membership publication запрещена.
- Definitive commit failure публикует nothing. Indeterminate outcome запрещает
  blind retry с новой identity и требует coherent inspection exact identity
  и revision; это не command idempotency §19(3) и не recovery §19(5).
- Persisted actual state — последний Owner-confirmed fact. После потери Owner
  он не доказывает текущую liveness до отдельного reconciliation contract.
- Conceptual operations остаются technology-neutral и не становятся Go/API:
  allocate candidate, create/read aggregate, read history, conditional claim
  attempt, publish Running, claim Stop, publish terminal outcomes.
- Formal implementation blockers: §19(2) до отдельного approval/status
  decision и downstream durable management idempotency,
  activation/replacement/rollback, recovery/reconciliation,
  reporting/redaction §19(3)–(6).

## Scope Audit

Accepted result: **13 Required / 0 Questionable / 0 Removable**.

- DP-014 EN/RU — `Required`: acceptance criteria 1–5, один зеркальный
  candidate contract §19(2), ownership/atomicity/failure/status truth.
- DP-013 EN/RU — `Required`: formal §19(2) gate pending separate
  approval/status decision, remaining §19(3)–(6) design work и next
  recommendation; без этих changes DP-013 противоречил бы новому Draft.
- design indexes EN/RU и task index — `Required`: navigation и отсутствие
  orphan DP/task record.
- MASTER_PLAN EN/RU — `Required`: durable architecture-dependency state и
  truthful next-design ordering.
- `.ai/PROJECT_CONTEXT.md`, `spec/current-state.md`, `spec/decisions.md` —
  `Required`: active task, Draft/Planned/Blocked boundary, isolated
  implementation truth и формальный gate.
- этот task record — `Required`: Task Contract, role handoffs, verification,
  rework, applicability, audit и final gates.

Ни один из 13 changes нельзя удалить, сохранив Definition of Done,
PROCESS-002, EN/RU parity и repository-only continuation context.
Premature §19(3) design, implementation, persistence package/schema/API,
Production Activation, code/tests/modules, generated, formatting-only и
unrelated changes отсутствуют.

## Size Guard

- ожидается 13 documentation files, 0 production lines, 0 packages, 1 новый
  architecture contract, 0 shipped behaviors;
- фактически 13 documentation files, 0 production lines, 0 packages, 1 новый
  architecture contract, 0 shipped behaviors;
- признаки Size Guard не сработали; D-001 расширил обязательную синхронизацию
  на DP-013 mirrors, но не добавил второй contract или behavior.

## Documentation Sync

- PROCESS-002 status: `Synchronized`;
- mandatory applicability:
  - task record: Required — active stage, verification, review/rework и
    PROCESS-002 evidence;
  - `spec/current-state.md`: Required — active task/design state и
    planned/implemented boundary;
  - MASTER_PLAN EN/RU: Required — durable dependency и roadmap status
    section 19(2)–(6);
  - DP-014 EN/RU: Required — новый candidate persistence contract;
  - DP-013 EN/RU: Required — formal §19(2) gate и next design recommendation;
  - `.ai/PROJECT_CONTEXT.md`: Required — current task, stage и durable
    operational handoff;
  - design indexes EN/RU: Required — navigation нового DP-014;
  - task index: Required — navigation TASK-016;
  - `spec/decisions.md`: Required — formal gate и unresolved dependency
    ordering;
  - `CHANGELOG.md`: Not applicable — user-facing/release change отсутствует;
  - root `README.md`: Not applicable — public usage/capability не изменены;
  - `spec/README.md`: Not applicable — specification tree/navigation не
    изменены;
  - ARCH-004 EN/RU: Not applicable — Active architecture contract и status не
    изменены; Draft DP-014 не повышает status;
  - DP-010 EN/RU: Not applicable — Draft/Implemented-in-isolation status и
    process-local Owner contract не изменены;
- EN/RU parity, links, planned/implemented truth и contradictions: PASS;
- formal §19(2) implementation gate сохранён до отдельного approval/status
  decision; downstream §19(3)–(6) также остаются blockers.

## Commit Gate

- exact command `Разрешаю коммит.` получена: да, после Coordinator Acceptance;
- commit message policy: Conventional Commits, proposal
  `docs(runtime): define operational identity persistence`;
- exact accepted file set: 13 documentation files из Scope Audit;
- post-acceptance diff: только bounded administrative closure в task record,
  `.ai/PROJECT_CONTEXT.md` и `spec/current-state.md`;
- temporary/generated/unrelated files: запрещены;
- final checks: documentation structure/status/parity, links и
  `git diff --check` PASS; preceding full `go test ./...` и `go vet ./...`
  PASS reused после documentation-only rework;
- один exact task commit авторизован; push, PR, merge и publication не
  авторизованы и не выполняются этой командой.

## Process Health

- trigger применим: нет;
- причина: TASK-016 не является десятой completed task после последнего review,
  rollback/escaped defect/repeated Publisher failure/>2 review returns
  отсутствуют.

## Handoff

- Task Intake зафиксирован;
- Documentation Baseline: complete, non-critical drift D-001 confined;
- Architecture Analysis: Ready, blockers 0;
- Pre-Implementation Documentation: Documentation Agent создал зеркальный
  Draft/Planned DP-014 и синхронизировал exact 13-file scope; status DP-013
  сохранён Draft/Planned/Blocked;
- Initial Reviewer: `Needs Revision`, R-001/R-002/R-003/N-001;
- bounded Documentation rework: §19(2) formal gate восстановлен, atomicity и
  append-only semantics уточнены, active stage синхронизирован;
- Repeat Tester: `PASS`;
- Repeat Reviewer: `Approved`, 0 blocking и 0 nonblocking findings;
- Final Reviewer: `Approved`, 0 blocking и 0 nonblocking findings;
- PROCESS-002: `Synchronized`;
- Scope Audit accepted: 13 Required / 0 Questionable / 0 Removable;
- Coordinator Acceptance: granted;
- следующий этап: создать один exact task commit после успешного Commit Gate;
- production/test changes не разрешены;
- open risk: persistence boundary не должна преждевременно определять
  idempotency, recovery или storage schema.

## Publication

- publication readiness: отсутствует до отдельного разрешённого task commit;
- repository: `dsdred/universal-websocket-platform`;
- accepted task branch:
  `docs/task-016-runtime-operational-identity-persistence-design`;
- exact task commit: отсутствует;
- base `main`: `49c7973`;
- accepted verification/scope: Final Reviewer Approved 0/0, Scope Audit
  13/0/0, PROCESS-002 Synchronized, exact 13-file diff;
- Publisher P0–P10: not authorized.

## Next Candidate

- рекомендуемая следующая Design-only work: focused durable management command
  idempotency ARCH-004 §19(3);
- readiness evidence: dependency ordering после candidate contract §19(2);
  formal §19(2) implementation gate остаётся до отдельного approval/status
  decision, поэтому рекомендация не разрешает management implementation;
- явно не начата: да.

## Closure

- Final status: `Completed — Coordinator Accepted`;
- Closed by: Coordinator;
- Date: 2026-07-29;
- commit: authorized after Coordinator Acceptance, not yet performed;
- publication: not authorized, not performed.

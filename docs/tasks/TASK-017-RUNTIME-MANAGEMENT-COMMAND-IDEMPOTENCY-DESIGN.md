# TASK-017 — Runtime Management Command Idempotency Design

## Status

`Completed — Coordinator Accepted`.

## Task Contract

### Task Mode

`Design-only`. Задача определяет один focused contract долговечной
идемпотентности state-changing management commands для Runtime Instance до
любой реализации management package, persistence adapter, API или production
wiring.

### Why Now

- TASK-016, `.ai/PROJECT_CONTEXT.md`, `spec/current-state.md`,
  `spec/decisions.md` и зеркальные MASTER_PLAN независимо рекомендуют этот
  slice следующим;
- Active ARCH-004 §19(3) является следующим обязательным prerequisite после
  candidate persistence boundary §19(2);
- Draft DP-014 отделяет conditional aggregate revision и
  inspect-after-indeterminate от client-command idempotency и тем самым задаёт
  необходимую нижнюю boundary;
- отдельный idempotency contract является наименьшим независимо проверяемым
  architecture slice и не смешивает activation, recovery или reporting.

### Definition of Done

1. Зеркальный Draft DP-015 EN/RU определяет command identity, durable claim,
   ownership, deduplication, replay и concurrency semantics state-changing
   Runtime management commands без выбора API, storage technology или schema.
2. Контракт однозначно связывает command outcome с exact management scope,
   Runtime Instance, operation kind и immutable request intent; повтор не
   создаёт вторую lifecycle mutation или Launch Attempt.
3. Definitive, in-progress и indeterminate outcomes, per-Instance unresolved-
   command barrier, retry/inspection boundary, retention prerequisite и
   redaction invariants согласованы с Active ARCH-004 и Draft DP-014 без
   подмены aggregate revision command idempotency.
4. Activation/replacement/rollback, recovery/reconciliation и operational
   reporting/redaction остаются явными последующими designs §19(4)–(6).
5. Связанные зеркальные документы, индексы и durable project state
   синхронизированы; documentation verification, PROCESS-002, Scope Audit и
   независимый final review завершаются без blocking findings.

### Out of Scope

- production-код, тесты, package, repository, database schema и migrations;
- HTTP paths, headers, DTO, status codes, SDK retry policy и concrete command
  key representation;
- выбор storage technology, transaction mechanism, clock или ID algorithm;
- approval/status decision DP-014 и снятие formal gate ARCH-004 §19(2);
- activation, replacement и rollback ordering §19(4);
- recovery/reconciliation §19(5), operational reporting/redaction §19(6);
- concrete authorization policy, Production Activation, automatic restart,
  scheduling, clustering и process isolation.

### Verification Plan

- проверить precedence Active ARCH-004, Draft DP-013 и Draft DP-014;
- проверить, что DP-015 не обещает API/schema/implementation и не повышает
  status других документов;
- сопоставить EN/RU headings, normative meaning, links и code fences;
- проверить command identity, same-key/same-intent replay,
  same-key/different-intent conflict, concurrent claim, durable terminal replay,
  in-progress observation и indeterminate outcome;
- проверить отсутствие conflict markers, trailing whitespace и broken links;
- выполнить `git diff --check`; полный `go test ./...` и `go vet ./...`
  использовать как regression evidence для documentation-only diff;
- независимый Reviewer проверяет архитектурные границы и возможность удалить
  каждый change без потери Definition of Done.

## Objective

Создать один проверяемый candidate design contract для prerequisite ARCH-004
§19(3): durable idempotency management commands. До отдельных approval/status
решений Draft не снимает formal gate §19(2), не удовлетворяет нормативно
§19(3) и не разрешает management implementation.

## Selection Evidence

- trusted baseline: clean synchronized `main` на merge commit `ec002e5`;
- active `In Progress` или `Blocked` task отсутствует;
- explicit next candidate: TASK-016 `Next Candidate`;
- durable corroboration: `.ai/PROJECT_CONTEXT.md`, `spec/current-state.md`,
  `spec/decisions.md` и зеркальные MASTER_PLAN;
- readiness: architecture refinement разрешён до production implementation;
  Active ARCH-004 задаёт command concurrency requirement, а Draft DP-014 —
  candidate durable conditional publication boundary;
- ranking: current Beta dependency, следующий unresolved prerequisite §19(3),
  smallest independently verifiable design, lowest unresolved risk и первый
  authoritative ordering;
- отклонены:
  - management implementation — Blocked formal §19(2) и designs §19(3)–(6);
  - approval DP-014 — materially different status decision, не design §19(3);
  - activation/replacement/rollback — последующий §19(4);
  - recovery/reconciliation — последующий §19(5);
  - operational reporting/redaction — последующий §19(6).

## Scope

- новый зеркальный DP-015 EN/RU;
- bounded synchronization DP-013/DP-014 EN/RU только для нового candidate
  contract, сохранённых gates и следующего prerequisite;
- зеркальные design indexes и MASTER_PLAN;
- task record и task index;
- применимые `.ai/PROJECT_CONTEXT.md`, `spec/current-state.md` и
  `spec/decisions.md`;
- только необходимые cross-links к существующим architecture/design sources.

## Non-Goals

- не начинать §19(4) или implementation;
- не определять transport command contract;
- не менять Active/Frozen architecture, public API или runtime behavior;
- не выполнять unrelated cleanup или formatting.

## Sources of Truth

- ADR-0002, ADR-0003 и ADR-0004;
- Frozen ARCH-002;
- Active ARCH-004 и ARCH-005;
- Draft DP-013 и DP-014 как subordinate candidate designs;
- production code/tests только как evidence отсутствия management
  implementation;
- TASK-016 и durable project-state documents.

## Roles

- Coordinator: intake, role assignments, gates, Scope Audit и acceptance;
- Architect: idempotency design, invariants и acceptance constraints;
- Documentation Agent: baseline, mirrored DP и project-state synchronization;
- Developer: Not applicable — production code запрещён;
- Tester: documentation verification и regression checks;
- Reviewer: независимый initial/final review; не автор изменений;
- Publisher: Not applicable — publication не разрешена этой командой.

## Branch

- исходный trusted baseline: `main` at `ec002e5`, clean,
  `main == origin/main`;
- task branch:
  `docs/task-017-runtime-management-command-idempotency-design`;
- branch action: создана и активирована до первого content change;
- запрещены stage, commit, push, merge, fetch, pull, rebase, reset и удаление
  веток без соответствующего explicit gate.

## Constraints

- ровно один новый architecture contract;
- Technology Neutrality и opaque command identity;
- authorization предшествует lifecycle mutation, но authorization result не
  становится долговечным command authority;
- Runtime Lifecycle Owner остаётся единственным lifecycle decision maker;
- один command identity не может представлять разные immutable intents;
- replay не создаёт новую aggregate revision, lifecycle mutation или Launch
  Attempt;
- planned и implemented state разделяются;
- commit только после точной команды `Разрешаю коммит.`.

## Stop Conditions

- конфликт с Approved ADR или Active/Frozen ARCH;
- необходимость включить §19(4)–(6) в candidate contract §19(3);
- невозможность сохранить technology/schema/API neutrality;
- более одного нового architecture contract или >15 changed files;
- critical documentation drift;
- materially different Ready solution без authoritative ordering.

## Acceptance Criteria

1. Command identity и immutable intent однозначно scoped и не становятся
   identity Runtime Instance, Launch Attempt или aggregate revision.
2. Same-key/same-intent concurrent/repeated submissions сходятся на одной
   durable command execution и одном replayable outcome; different intent с
   тем же key выполняет zero mutation.
3. Durable claim предшествует lifecycle mutation; terminal outcome сохраняется
   до replay, unresolved command сохраняет per-Instance barrier и не выдаётся
   за успешный или terminal, а tracked Start допускает обязательный Stop.
4. Indeterminate persistence/lifecycle boundaries запрещают blind re-execution
   и требуют exact inspection без создания второй mutation.
5. Retention и redaction constraints не раскрывают cross-scope identity и не
   позволяют забыть key, пока безопасный replay/re-execution не доказан.
6. Subsequent §19(4)–(6), formal §19(2) gate, EN/RU parity, navigation и
   repository checks отражены правдиво.

## Existing Coverage Report

- Existing Coverage: ARCH-004 требует explicit serialization и idempotency;
  DP-013 фиксирует authorization-before-mutation и exact routing; DP-014
  определяет conditional revision и inspect-after-indeterminate persistence
  boundary.
- Coverage Gap: command identity, immutable intent binding, durable claim,
  replay, concurrency, retention и command-level indeterminate semantics не
  определены.
- Added Proof Tests: Not applicable до production implementation.
- Added Regression Tests: Not applicable — code/test changes запрещены.
- Remaining Limitations: executable durability/concurrency/restart proofs
  появятся только после нормативного закрытия prerequisites и отдельной
  implementation task.

## Documentation Baseline

- Status: `Synchronized with expected task-stage drift`; critical drift 0.
- EN/RU baseline: design tree 14/14 filenames before DP-015; ARCH-004,
  DP-013 и DP-014 имеют mirror и одинаковые Design/Implementation statuses.
- Factual boundary: Runtime Lifecycle Owner, Flow и Source реализованы только
  изолированно; management routing, persistence и idempotency существуют
  только как planned designs; package/schema/API/recovery/wiring отсутствуют.
- Expected drift: task index и durable project-state документы ещё не отражали
  active TASK-017; disposition — Required synchronization в этой task.
- Critical conflicts с ADR, Frozen ARCH-002 или Active ARCH-004: отсутствуют.
- Exact expected final file set: 15 documentation files — task record/index,
  DP-015 EN/RU, DP-013 EN/RU, DP-014 EN/RU, design indexes EN/RU,
  MASTER_PLAN EN/RU, `.ai/PROJECT_CONTEXT.md`, `spec/current-state.md` и
  `spec/decisions.md`.

## Architecture Analysis

- Verdict: `Ready`; blockers 0.
- Command identity: `(CommandScope, CommandKey)`, где scope содержит exact
  operational domain, Workspace, Configuration, Runtime Instance и operation
  kind; key не переиспользует domain identities или aggregate revision.
- Intent: exact semantic operation input; Start связывает exact Published
  ConfigurationVersion, Stop не выводит version/latest attempt. Principal,
  credentials, context, transport и observation исключены.
- Ordering: validate -> exact lookup -> current authorization -> cancellation
  gate -> durable inspect/claim -> lifecycle delegation только newly committed
  claim. Authorization повторяется для replay и не сохраняется как authority.
- Same key/same intent сходится на одном non-terminal observation или terminal
  replay без delegation; same key/different intent даёт conflict с zero
  mutation.
- Exact surface DP-013 не меняется и command identity не передаётся в
  Flow/Owner/DP-014 publication. Claiming path вызывает exact DP-013 operation
  не более одного раза; definitive outcome сохраняется для replay.
- Claim commit возвращает один non-transferable process-local execution permit
  только claiming path. `Claimed -> Terminal` запрещает выдавать non-terminal
  state за success; Claim без matching live permit является unresolved и
  закрывает durable per-Instance barrier.
- Barrier evaluation, exception tracked Start и insertion нового claim имеют
  одну per-Instance atomic command-admission linearization point. Ровно один
  distinct Stop может получить собственный permit рядом с tracked Start и
  обязан достичь того же Owner; другой Start/Stop получает non-mutating
  in-progress conflict.
- Caller cancellation между claim и downstream gate может предотвратить
  lifecycle mutation. После победы Start gate synchronous wait продолжается
  до Owner outcome; для Stop поздняя cancellation может прервать только caller
  wait. Ни один вариант не удаляет record и не разрешает duplicate delegation.
- Indeterminate outcome или потеря exact live permit требуют inspection и
  запрещают replacement key, blind re-delegation, другой state-changing command
  и второй Launch Attempt. Restart convergence и truthful opening barrier
  остаются deferred §19(5).
- До focused retention contract command record не удаляется и identity не
  переиспользуется; TTL сам по себе недостаточен.
- Runtime Lifecycle Owner сохраняет sole lifecycle/live Host ownership;
  idempotency boundary не становится recovery executor, command bus, registry
  или service locator.
- DP-015 остаётся non-normative Draft/Planned; gates §19(2) и §19(3) и
  downstream §19(4)–(6) сохраняются.

## Verification

- Existing Coverage Report: заполнен до test changes; code/tests не менялись,
  Added Proof/Regression Tests — Not applicable, executable durability,
  concurrency и restart proofs остаются future implementation work.
- Verification Matrix: concurrency/lifecycle проверены design invariants и
  independent review; race, API, dependencies и public API не применимы к
  documentation-only diff; documentation checks применены полностью.
- Full regression: `go test ./... -count=1` — PASS; `go vet ./...` — PASS.
- Structure: DP-015 EN/RU 29/29 headings и 4/4 fences; DP-013 35/35 и
  14/14; DP-014 27/27 и 4/4; MASTER_PLAN 36/36 и 0/0.
- Links: changed scope 169/0 missing; repository 770/0 missing.
- `git diff --check`: PASS; conflict markers: 0.
- Initial Reviewer: `Needs Revision` — B-001 impossible atomic linkage through
  unchanged DP-013, B-002 inaccurate cancellation, N-001 bookkeeping.
- First rework removed cross-boundary atomicity promise and made cancellation
  phase-specific. Repeat Reviewer: `Needs Revision` — B-003 barrier blocked
  mandatory Stop-during-Starting and different-key admission was not atomic.
- Second rework introduced one non-transferable Execution Permit, atomic
  per-Instance command admission, tracked-Start distinct Stop exception and
  unresolved Claim classification.
- Final architecture Reviewer: `Approved with Findings`, blocking 0; B-001,
  B-002 и B-003 resolved. После closure bookkeeping terminal Reviewer:
  `Approved`, 0 blocking и 0 nonblocking findings; N-001 resolved.

## Scope Audit

Accepted: **15 Required / 0 Questionable / 0 Removable**.

- DP-015 EN/RU — Required: candidate contract §19(3) и acceptance criteria.
- DP-013/DP-014 EN/RU — Required: cross-links, formal gates и next ordering.
- Design indexes EN/RU и task index — Required: navigation.
- MASTER_PLAN EN/RU — Required: durable dependency state.
- `.ai/PROJECT_CONTEXT.md`, `spec/current-state.md`, `spec/decisions.md` —
  Required: repository-only continuation, factual gaps и gates.
- Task record — Required: contract, evidence, rework, verification и closure.

Production, tests, modules, API, schema, generated, formatting-only и unrelated
changes отсутствуют. Size Guard: 15 documentation files, 0 production lines,
0 packages, 1 architecture contract, 0 shipped behaviors; trigger `>15 files`
не сработал.

## Documentation Sync

- PROCESS-002: `Synchronized`.
- Required: task record/index, `spec/current-state.md`, MASTER_PLAN EN/RU,
  DP-015 EN/RU, bounded DP-013/DP-014 EN/RU, `.ai/PROJECT_CONTEXT.md`, design
  indexes EN/RU и `spec/decisions.md`.
- Not applicable: `CHANGELOG.md` — user-facing/release change отсутствует;
  root/spec README — capability/tree не менялись; ADR/ARCH — authoritative
  status/contract не менялся.
- EN/RU parity, planned/implemented truth, links и contradictions: PASS.

## Process Health

- trigger применим: нет; TASK-017 не является десятой completed task после
  TASK-012, rollback/escaped defect/repeated Publisher failure и >2 review
  returns отсутствуют.

## Handoff

- Task Intake first-content-change, Documentation Baseline и Architecture
  Analysis complete; blockers 0.
- Documentation Agent: exact 15-file mirrored design/state diff complete.
- Developer: Not applicable; production code запрещён и не изменён.
- Tester: PASS; regression и documentation checks complete.
- Independent terminal Reviewer: `Approved`, 0 blocking и 0 nonblocking
  findings; все contract findings и closure bookkeeping resolved.
- PROCESS-002 `Synchronized`; Scope Audit 15/0/0.
- Coordinator Acceptance granted for exact 15-file documentation diff.
- Commit readiness: только после точной команды `Разрешаю коммит.`; stage,
  commit, push и publication не выполнялись.

## Commit Gate

- exact command `Разрешаю коммит.` получена после Coordinator Acceptance: да;
- commit message policy: Conventional Commits;
- selected message: `docs(runtime): define management command idempotency`;
- exact accepted file set: 15 documentation files из Scope Audit;
- post-acceptance changes: только bounded administrative closure и этот
  permission record; architecture contract и scope не изменены;
- temporary/generated/unrelated files: отсутствуют;
- final checks: full Go regression, vet, EN/RU parity, links, exact file set и
  `git diff --check` — PASS;
- разрешён ровно один local task commit; push, PR, merge и publication не
  разрешены.

## Publication

- publication readiness: отсутствует до отдельного разрешённого task commit;
- Publisher P0–P10: not authorized.

## Next Candidate

- рекомендуемая Design-only work: focused activation, replacement и rollback
  ordering ARCH-004 §19(4);
- readiness: dependency ordering после candidate §19(2)/§19(3); их formal
  approval/status gates остаются открытыми;
- явно не активирован: да.

## Closure

- Final status: `Completed — Coordinator Accepted`;
- Closed by: Coordinator;
- Date: 2026-08-01;
- commit: authorized after Coordinator Acceptance, not yet performed;
- publication: not authorized, not performed.

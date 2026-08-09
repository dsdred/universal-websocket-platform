# TASK-027 — Runtime Activation Orchestration Prerequisites Design

## Status

`Completed — Coordinator Accepted`.

## Task Contract

### Task Mode

`Design-only`. Определить один focused integration contract, который делает
Approved DP-016 реализуемым без упрощения его acceptance proofs и без
production-кода.

### Why Now

- TASK-026 обнаружила воспроизводимый Architecture Blocking Discovery и
  переведена в `Blocked by Architecture`;
- текущая реализация DP-015 поддерживает только primitive Start/Stop и не
  предоставляет parent/phase claim API;
- DP-011/DP-013 описывают planned private Start-claim continuation, но текущие
  package surfaces её не предоставляют;
- DP-013 authorization surface не различает activation, replacement и
  rollback как exact caller intents;
- упрощённый adapter не способен доказать DP-016 §25 и явно отклонён
  Coordinator.

### Definition of Done

1. TASK-026 остаётся `Blocked by Architecture`; Variant B, implementation,
   Acceptance, commit и publication для неё не выполняются.
2. Зеркальный DP-019 EN/RU определяет один согласованный prerequisite contract:
   parent/phase claim API поверх DP-015, private Start-claim continuation
   DP-011/DP-013 и exact orchestration authorization scope.
3. Contract сохраняет существующие ownership/lifecycle semantics: Owner остаётся
   единственным lifecycle authority; activation остаётся primitive Start;
   replacement/rollback используют только bounded parent с linked Stop/Start
   phases; permits не передаются.
4. DP-011, DP-013, DP-015 и DP-016 получают только необходимые ссылки,
   implementation-boundary и status clarifications; ранее принятые semantics
   вне prerequisite scope не меняются.
5. Verification Matrix, PROCESS-002, Status Consistency Validation, Scope Audit
   и Independent Review завершены без blocking findings.

### Out of Scope

- production-код, tests, package/API implementation и wiring;
- реализация `internal/runtimeactivation` или изменение кода существующих
  packages;
- упрощение, исключение или deferred classification любого DP-016 §25 proof;
- HTTP/CLI/DTO, concrete authorization policy, Principal model;
- external storage/schema/migrations, DP-017 recovery executor, DP-018 report
  projector и Production Activation;
- automatic rollback/restart, zero-downtime replacement, scheduling,
  supervision и unrelated cleanup;
- commit, push, PR, merge или publication.

### Verification Plan

- semantic trace DP-019 → DP-011/DP-013/DP-015/DP-016 and ARCH-004 §19(4);
- EN/RU heading, status and normative parity;
- repository-wide status/absence sweep for DP-011, DP-013, DP-015, DP-016,
  TASK-026 and TASK-027;
- relative-link validation, conflict-marker scan and `git diff --check`;
- `go test ./... -count=1` and `go vet ./...` as no-code regression evidence;
- PROCESS-002 applicability record and exact per-file Scope Audit;
- fresh Independent Reviewer verdict; Coordinator Acceptance only after
  `Approved` with zero blocking findings.

## Objective

Закрыть архитектурные prerequisites DP-016 одним technology-neutral
integration contract, не реализуя orchestration и не меняя approved ordering:
точно определить authorization, durable parent/phase admission и synchronous
private continuation после Owner claim и до Load.

## Selection Evidence

- baseline: `main@751577e839cdea3a0f35032b1339d1d9f74d28ec`, равный
  `origin/main` после PR #26;
- единственная active task TASK-026 имела только атрибутированный Task
  Record/index diff и была заблокирована до implementation;
- Coordinator подтвердил blocker и запретил Variant B или scope workaround;
- DP-016 implementation не является Ready; Design refinement разрешён
  PROCESS-001 и является первым prerequisite по dependency order;
- выбран один cohesive integration contract вместо трёх несогласованных API
  edits;
- отклонены: упрощённый DP-016 slice; production implementation; изменение
  lifecycle ownership; отдельная concrete policy task; DP-017/DP-018 work.

## Scope

- TASK-026 blocker record и task index;
- новый зеркальный DP-019 EN/RU и design indexes;
- минимальные DP-011, DP-013, DP-014, DP-015, DP-016 и DP-017 EN/RU
  clarifications/links/status consistency;
- зеркальные MASTER_PLAN и обязательные project-state mirrors;
- TASK-027 record и exact PROCESS-002 evidence.

## Sources of Truth

- ADR-0002, ADR-0003, ADR-0004;
- Frozen ARCH-002; Active ARCH-004 and ARCH-005;
- DP-010 lifecycle ownership;
- Draft DP-011 and DP-013 factual isolated surfaces;
- Approved DP-014, DP-015, DP-016 and DP-017;
- implementation/tests only as evidence of missing surfaces;
- PROCESS-001 and PROCESS-002.

## Roles

- Coordinator: contract, gates, Scope Audit, Closure Audit and Acceptance;
- Architect: focused DP-019 decision and constraints;
- Documentation Agent: mirrored design/project-state synchronization;
- Developer/Tester code changes: Not applicable — production/tests forbidden;
- Verifier: documentation matrix and no-code regressions;
- Reviewer: fresh independent final review without author conclusions;
- Publisher: Not applicable — publication forbidden.

## Branch

- branch: `docs/task-027-runtime-activation-prerequisites-design`;
- trusted history baseline remains exact `main@751577e`;
- TASK-026 commit was explicitly forbidden, therefore its two-file,
  coordinator-attributed Blocked governance diff was carried into this
  documentation branch and is part of TASK-027 prerequisite-state scope;
- no implementation diff was carried; stage/commit/push/merge/publish remain
  forbidden.

## Constraints

- exactly one new architecture contract;
- no second lifecycle owner, no command bus, no transferable permit;
- no admission/Owner lock across storage, callback, wait or lifecycle work;
- Workspace, Configuration, Runtime Instance and exact target version are
  authorization and immutable-intent facts;
- parent and phase identities are derived, finite and non-reusable;
- Owner claim precedes private continuation; durable attempt publication and
  execution binding complete before Load or other external preparation;
- unknown/indeterminate state remains unresolved and fail-closed.

## Stop Conditions

- proposed contract changes Active/Frozen ownership or weakens DP-016 proofs;
- more than one competing architecture contract is required;
- implementation becomes necessary to decide semantics;
- EN/RU or project-state contradiction remains;
- Independent Reviewer reports a blocking finding.

## Existing Coverage Report

- **Existing Coverage:** DP-011/DP-013 specify base launch/routing and a planned
  continuation semantically; DP-015 specifies primitive claims and conceptual
  parent/phase rules; DP-016 specifies complete ordering/proofs.
- **Coverage Gap:** no exact implementable parent/phase surface, no exact
  orchestration authorization actions, and no complete construction/decision
  contract joining command admission to the Start-claim continuation.
- **Added Proof Tests:** Not applicable — design-only, tests cannot prove a
  planned API implementation.
- **Added Regression Tests:** Not applicable — code/test changes prohibited.
- **Remaining Limitations:** all executable, concurrency/race and production
  composition proofs remain for successor implementation tasks.

## Size Guard

- expected: documentation only, one new DP, no production lines/packages;
- more than 15 files is expected because six linked DP pairs, indexes,
  roadmap and project-state mirrors must remain consistent;
- the slice remains cohesive because removing any linked mirror would leave
  the prerequisite contract ambiguous or stale; final exact disposition is
  deferred to Scope Audit.

## Architecture Decision

Architect decision: DP-019 `Approved`, Implementation Status `Planned`;
architecture blockers inside the design task 0. Independent Review and
Coordinator Acceptance are complete.

- выбран parent/phase API path DP-015; новые lifecycle operations Owner не
  добавляются;
- initial activation остаётся primitive Start; parent kinds ограничены
  Replace/Rollback, linked phases — optional StopOld и single StartTarget;
- authorization input фиксирует Operational Domain, Workspace,
  Configuration, Runtime Instance, action и exact target version до любого
  command claim; replay авторизуется снова;
- linked phase authority производна от immutable parent и не переносится;
  private scoped invoker недоступен transport caller;
- callback-scoped parent/phase executions expire при return/panic/Goexit или
  client reconstruction; committed nonterminal record становится unresolved;
- managed Flow получает construction-bound stateless continuation; после sole
  Owner claim и до Load continuation сначала разрешает already-admitted pending
  Stop; `StopConverged` выходит без attempt publication/binding, а только
  no-pending path публикует exact attempt/version в DP-014, связывает generation
  и проходит final Continue gate;
- stale/different/unknown/indeterminate facts fail closed как Blocked;
- DP-016 ordering, ownership и полный §25 proof set сохранены без limitation.

## Documentation Baseline and Design Changes

- исходный live status: DP-011/DP-013 base implemented in isolation with
  continuation Planned; primitive DP-015 implemented in isolation; DP-016
  Approved/Planned; TASK-026 had no production/test changes;
- drift: TASK-026 одновременно содержала valid blocker и invalid self-
  resolution through a reduced proof set; record rewritten as Blocked-only;
- added: mirrored Approved/Planned DP-019 and design navigation;
- synchronized: DP-011, DP-013, DP-014, DP-015, DP-016, DP-017 EN/RU,
  MASTER_PLAN EN/RU,
  task index, `.ai/PROJECT_CONTEXT.md`, `spec/current-state.md` and
  `spec/decisions.md`;
- implementation truth remains unchanged: all DP-019 surfaces and DP-016
  orchestrator are absent.

## Verification

- Verification Matrix:
  - documentation/design: semantic trace and mirror checks PASS;
  - concurrency/lifecycle/shared state: design invariants reviewed; race and
    stress Not applicable because code/tests are unchanged and implementation
    remains Planned;
  - API/production wiring: Not applicable — public/production surface is
    explicitly forbidden; planned internal surface reviewed as design;
  - dependencies/modules: no `go.mod`/`go.sum` change;
  - public API/godoc: Not applicable — no identifiers implemented;
- `go test ./... -count=1`: PASS;
- `go vet ./...`: PASS;
- repository-source relative links: 878 checked, 0 broken (generated `.cache`
  dependency Markdown excluded);
- EN/RU parity: DP-011 33/33 headings and 16/16 fences; DP-013 35/35 and
  14/14; DP-014 28/28 and 4/4; DP-015 29/29 and 4/4; DP-016 30/30 and 4/4;
  DP-017 30/30 and 2/2; DP-019 25/25 and 16/16;
  indexes 1/1; MASTER_PLAN 36/36;
- status assertions: 16/16 PASS;
- conflict markers: 0; `git diff --check`: PASS;
- exact scope: 24 expected/24 actual, 0 missing, 0 extra;
- staged changes: 0; production/test/module changes: 0.

## PROCESS-002 Applicability

Status: `Synchronized`.

- TASK-026 record: Required — Blocked decision and invalid Variant B removal;
- TASK-027 record: Required — contract, handoffs, verification and closure;
- DP-019 EN/RU: Required — one new prerequisite architecture contract;
- DP-011/DP-013/DP-014/DP-015/DP-016/DP-017 EN/RU: Required — exact planned
  seam/status and DP-019 linkage without changing implemented behavior;
- design indexes EN/RU: Required — navigation and live statuses;
- MASTER_PLAN EN/RU: Required — dependency ordering and TASK-026 blocker;
- `spec/current-state.md`: Required — factual task/capability state and TASK-025
  terminal publication evidence;
- `spec/decisions.md`: Required — Approved/Planned DP-019 and durable
  architecture decision;
- `.ai/PROJECT_CONTEXT.md`: Required — current tasks, status matrix and durable
  continuation context;
- task index: Required — TASK-026/TASK-027 navigation/status;
- `CHANGELOG.md`: Not applicable — no user-facing or release capability;
- root README mirrors: Not applicable — product usage/capability unchanged;
- `spec/README.md`: Not applicable — specification tree unchanged;
- ARCH-004 EN/RU: Not applicable — §19(4), status and ownership unchanged;
- DP-010 EN/RU: inspected; Not applicable — lifecycle ownership is referenced
  but unchanged by this integration clarification;
- production/tests/modules: Not applicable — forbidden and unchanged.

## Scope Audit

Final result: **24 Required / 0 Questionable /
0 Removable**.

- 2 DP-019 mirrors: new contract and semantic parity;
- 12 DP-011/013/014/015/016/017 mirrors: exact prerequisite linkage and truthful
  implementation boundary;
- 2 design indexes and 2 MASTER_PLAN mirrors: navigation/status/dependency
  synchronization;
- TASK-026, TASK-027 and task index: Blocked handoff, active contract and
  navigation;
- `.ai/PROJECT_CONTEXT.md`, `spec/current-state.md`, `spec/decisions.md`:
  durable project-state and decision synchronization.

Removing any changed file would leave an orphan, one language mirror stale, an
authoritative dependency ambiguous, TASK-026 falsely resolvable, or repository-
only continuation incomplete. No implementation, test, module, generated,
temporary, formatting-only, next-task implementation or unrelated file exists.

Size Guard reassessment: 24 files exceed the 15-file trigger, but the slice has
0 production lines, 0 packages, 1 architecture contract and 0 shipped
behaviors. Splitting it would knowingly leave EN/RU, six linked DPs,
navigation, roadmap or project-state inconsistent, so the cohesive
documentation scope is retained.

## Independent Review

First independent review verdict: **Needs Revision**.

- B-001: the original DP-019 text construction-bound per-operation facts to a
  long-lived DP-013 Flow without defining a private per-call data path;
- B-002: live DP-015/current-state/task status wording did not consistently
  split implemented primitive claims from the planned parent/phase extension;
- B-003: recorded fence/status verification counts were stale.

Bounded rework completed without production/test changes: DP-019 now defines
the immutable per-call `StartExecutionBinding` and synchronous private
`InvokeManagedStart`/`Flow.StartManaged` seam; all live linked status mirrors
were swept and synchronized; verification counts and exact scope were rebuilt
from repository state. Repeat Independent Review is pending.

During the repeat review, the Reviewer found one remaining ordering ambiguity:
DP-019 listed attempt publication/generation binding before the mandatory early
pending-Stop rendezvous required by DP-011/DP-013/DP-016. The EN/RU contract
and this record were corrected in the same bounded design slice: an
already-admitted pending Stop is resolved first; `StopConverged` exits without
attempt publication or binding; only the definitive no-pending path may publish
and bind before the final gate.

Final repeat review verdict: **Approved**. Blocking findings: **0**;
non-blocking findings: **0**. The Reviewer independently confirmed the
implementable one-long-lived-Flow/per-call binding seam, exact authorization and
pending-Stop ordering, unchanged DP-016 §25 proofs, PROCESS-002, status
consistency, Verification Matrix, and exact Scope Audit 24/0/0. The review was
read-only and changed no file or Git state.

## Coordinator Closure Audit

**PASS — Coordinator Acceptance completed.**

- Task Contract: PASS; the one design-only prerequisite contract is complete,
  Variant B remains rejected, and no production/test implementation exists;
- architecture: PASS; DP-019 is Approved/Planned, sole Owner authority and all
  DP-016 §25 proofs are preserved, and the private seam is implementable;
- Verification Matrix: PASS; tests, vet, 878/0 repository-source links,
  EN/RU parity, status assertions, diff and conflict checks are closed;
- PROCESS-002 and Status Consistency Validation: PASS;
- Independent Review: Approved, blocking findings 0, non-blocking findings 0;
- Scope Audit: 24 Required / 0 Questionable / 0 Removable;
- repository state: exactly the attributed 24-file documentation/governance
  diff, staged 0, production/test/module changes 0;
- TASK-026 remains `Blocked by Architecture`; its Acceptance, commit and
  publication were not performed;
- commit, push, PR, merge and publication for TASK-027 were not performed.

Coordinator decision: **TASK-027 Accepted**.

## Next Candidate

After acceptance, the next candidate is a bounded implementation of the
DP-019 prerequisites only. TASK-026 remains Blocked until that implementation
is accepted; it does not become active automatically.

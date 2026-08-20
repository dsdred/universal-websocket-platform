# TASK-041 — Managed Continuation Documentation Baseline Reconciliation

## Status

`Completed — Coordinator Accepted (2026-08-20)`.

## Task Contract

### Task Mode

`Documentation-only`. Устранить critical implemented-state drift в live
описаниях уже реализованных managed Flow, command-gate, continuation и
attempt/generation binding seams до проектирования private exact-scope invoker.

### Why Now

- autonomous intake выбрал следующий private exact-scope invoker candidate из
  closure TASK-040 и durable project-state sources;
- обязательный Documentation Baseline обнаружил, что DP-011, DP-013, DP-016,
  DP-017, design indexes, `.ai/PROJECT_CONTEXT.md` и `spec/decisions.md`
  ошибочно называют уже реализованные TASK-032/TASK-037 seams Planned/absent;
- PROCESS-001 запрещает начинать новый design при critical drift;
- correction и invoker contract являются двумя независимо поставляемыми
  результатами и вместе требуют около 20 файлов, поэтому Size Guard требует
  split до архитектурного проектирования.

### Definition of Done

1. DP-011 EN/RU отделяет implemented-in-isolation managed Flow и
   StartClaimContinuation от отсутствующего concrete invoker, terminal
   publication, orchestrator и production wiring.
2. DP-013 EN/RU не смешивает реализованные managed command/binding/continuation
   surfaces с отсутствующим DP-013 composition-private invoker.
3. DP-016 и DP-017 EN/RU признают принятые Slices 1–3 без изменения Approved
   semantics, Design Status или Planned overall implementation boundary.
4. Design indexes EN/RU, `.ai/PROJECT_CONTEXT.md`, `spec/current-state.md` и
   `spec/decisions.md` согласованы с TASK-032/TASK-037 и current code.
5. Historical TASK-018/TASK-026 baseline statements не переписываются как live
   facts.
6. Concrete invoker остаётся absent/unactivated, TASK-026 остаётся
   `Blocked by Architecture`, product capability не заявляется.
7. EN/RU parity, links, contradiction sweeps, focused Go regression,
   PROCESS-002, Scope Audit и independent review проходят до Acceptance.

### Out of Scope

- новый invoker design, production code или tests;
- изменение Approved DP-014–DP-019 semantics/status;
- terminal publication, terminalization или TASK-026 orchestrator;
- public API, persistence, recovery, reporting, wiring, Production Activation;
- исторические bulk rewrites; stage, commit, push, PR, merge или publication.

### Verification Plan

- compare EN/RU headings, fences, status and normative meaning;
- validate affected and repository-relative links;
- sweep stale live Planned/absent claims while preserving invoker absence and
  TASK-026 Blocked state;
- rerun focused regression for the seven managed packages;
- run `git diff --check`, conflict-marker, whitespace and exact-file checks;
- perform PROCESS-002, Scope Audit, independent review and bounded rework.

## Intake Reassessment and Selection Evidence

Trusted preflight: clean synchronized
`main@2a9f5ca2d7034b127aa837558552c71441a8c788`, no In Progress task,
TASK-026 still Blocked. TASK-040 named the invoker next, so TASK-041 was first
created as its design intake.

Documentation Baseline then found critical live drift. Architect blocked the
new design gate; Documentation Agent proved a standalone exact 15-file repair
versus an approximately 20-file combined repair/design scope. Coordinator
therefore re-contracted this same uncommitted first content change before any
other edit. Invoker design remains the next candidate.

Rejected: design over contradictory sources; combined repair/design; rewriting
historical facts; resuming TASK-026; starting terminal publication.

## Documentation Baseline

Verdict: `Drift Detected — critical; bounded repair Ready`.

- DP-011 EN/RU: managed Flow/StartClaimContinuation incorrectly Planned/absent;
- DP-013 EN/RU: implemented managed surfaces conflated with absent invoker;
- DP-016/017 EN/RU: accepted managed continuation/binding called Planned;
- design indexes EN/RU: stale DP-011/015/019/020 summaries;
- `.ai/PROJECT_CONTEXT.md`: current paragraphs contradict TASK-037 evidence;
- `spec/decisions.md`: stale Flow/continuation/execution-binding statements;
- `spec/current-state.md`: current capability text is materially correct; only
  active-task/closure synchronization is Required.

Affected mirror headings/fences align; 374 scoped links valid / 0 broken.
ARCH-004/005 and current DP-010/014/015/019/020 invoker-absence statements have
no blocking drift.

## Sources of Truth

Approved DP-014–DP-019; Active ARCH-004/005; Draft DP-011/013/020; TASK-032,
TASK-035, TASK-037, TASK-038 and TASK-040; current managed package code/tests;
PROCESS-001 and PROCESS-002.

## Roles and Handoffs

- Coordinator: re-contract, gates, scope audit and closure;
- Architect: new-design gate Blocked; repair must not change Approved semantics;
- Documentation Agent: exact reconciliation and PROCESS-002;
- Tester: focused regression and documentation verification;
- independent Reviewer: truth, parity, historical/live boundary and deletion test;
- Developer: Not applicable; code/test edits are forbidden.

## Branch Decision

- branch `docs/task-041-private-exact-scope-invoker-design` from clean
  `main@2a9f5ca2d7034b127aa837558552c71441a8c788`;
- retained after re-contract to avoid unnecessary branch mutation;
- this record was the first content change; the current bounded worktree
  contains the exact 15-file documentation repair recorded below;
- bare continuation permits no commit or publication.

## Size Guard

Exact ceiling: **15 files**, zero production/test lines, zero new package or
architecture contract, one factual repair. Decision: `DO NOT SPLIT` further.
Invoker design is split into the next task and forbidden here.

## Existing Coverage Report

### Existing Coverage

Tests prove six-field authorization and primitive/linked binding; managed
primitive and parent/StartTarget admission; early/final/no-claim rendezvous;
revision-sandwich continuation; managed Flow validation/four outcomes; DP-014
claim/bind; ten expected-attempt Stop groups; public Directory routing/isolation.
Fresh focused baseline across seven relevant packages: PASS.

### Coverage Gap

No DP-013 composition-private invoker or end-to-end invoker proof exists. This
intentional gap belongs to later design/implementation.

### Added Proof Tests / Added Regression Tests

Not applicable: test edits are forbidden; existing regression is rerun.

### Remaining Limitations

Invoker, terminal publication, orchestrator and production wiring remain absent.
TASK-026 remains Blocked.

## PROCESS-002 Applicability

- task record/index; DP-011/013/016/017 EN/RU; design indexes;
  `.ai/PROJECT_CONTEXT.md`, `spec/current-state.md`, `spec/decisions.md`:
  **Required**;
- MASTER_PLAN EN/RU: **Not applicable**; no milestone/dependency/roadmap change;
- root README and `CHANGELOG.md`: **Not applicable**; no user-facing change;
- other ADR/ARCH/DP/task/roadmap/review docs: **Not applicable** unless a direct
  live contradiction is found;
- code/tests/modules/generated artifacts: **Not applicable** and forbidden.

## Documentation Agent Handoff

- stage: bounded factual reconciliation only; no invoker design or Approved
  semantic/status change;
- repaired live sources: DP-011/013/016/017 EN/RU, design indexes,
  `.ai/PROJECT_CONTEXT.md`, `spec/current-state.md` and `spec/decisions.md`;
- preserved boundary: managed Flow, command gates, continuation and
  attempt/generation binding are implemented and accepted in isolation;
  concrete composition-private invoker, terminal publication, orchestrator and
  production wiring remain absent;
- preserved history: TASK-018/TASK-026 baseline descriptions remain historical
  evidence and are not rewritten as current facts;
- next role: Tester verifies existing managed packages and documentation
  structure/status/links; independent Reviewer then checks truth, parity and
  deletion scope before Coordinator Acceptance.

## PROCESS-002 Result

Status: `Synchronized`.

- language scope remains mirrored; Draft/Approved Design Status values are
  unchanged;
- no factual product capability, public API or production integration is
  claimed;
- task/index and durable project state are synchronized to completed TASK-041,
  no current documentation task, and the unactivated next design candidate;
- MASTER_PLAN, root README and CHANGELOG remain explicitly Not applicable for
  this factual repair;
- concrete invoker remains absent/unactivated and TASK-026 remains Blocked;
- mirror, status, link, focused regression and diff evidence is complete;
  Coordinator Scope Audit is recorded below and repeat final review follows.

## Documentation Verification Evidence

- focused existing regression:
  `go test ./internal/runtimelaunchflow ./internal/runtimecommandidempotency
  ./internal/runtimeorchestrationbinding
  ./internal/runtimeorchestrationcontinuation ./internal/runtimeidentity
  ./internal/runtimelifecycle ./internal/runtimemanagement -count=1` — PASS,
  seven packages;
- EN/RU structure — PASS: DP-011 headings 33/33, fences 16/16; DP-013
  35/35 and 14/14; DP-016 30/30 and 4/4; DP-017 30/30 and 2/2; design indexes
  headings 1/1;
- Design Status — unchanged: DP-011/013 Draft, DP-016/017 Approved; DP-016/017
  Implementation Status remains Planned overall;
- repository-wide relative Markdown links — 924 valid / 0 broken by the final
  independent full-repository parser; the earlier 891 count used a narrower
  pre-rework method and is superseded;
- initial independent Tester verdict — `FAIL`: residual live wording in
  DP-016 authority/section 15 EN/RU and DP-017 section 8 EN/RU still called the
  implemented managed continuation Planned or absent;
- bounded rework updates only those live paragraphs and preserves Approved /
  Planned-overall statuses, absent concrete invoker and historical facts;
- repeat stale live continuation/binding sweep — 0 in DP-016/017 EN/RU;
- rework mirror structure — PASS: DP-016 headings 30/30, fences 4/4;
  DP-017 headings 30/30, fences 2/2;
- rework status assertions — PASS: DP-016/017 remain Approved with
  Implementation Status Planned overall; concrete invoker and production
  composition remain absent;
- rework `git diff --check` and conflict-marker checks — PASS; only
  informational LF-to-CRLF working-copy warnings were emitted;
- repeat independent Tester verdict — `FAIL`: Branch Decision still called the
  task record the first and only current content change after the full 15-file
  repair existed;
- bookkeeping rework preserves the truthful first-content-change history and
  records the current exact 15-file worktree; no design or product fact changed;
- final independent Tester verdict after bookkeeping rework — `PASS`, blocking
  findings 0; focused regression, mirror/status/stale-live/link/diff and exact
  file-set evidence confirmed;
- concrete invoker absent/unactivated and TASK-026 Blocked assertions remain
  present in live project-state/design sources;
- exact file set — 15 expected / 0 unexpected / 0 missing; code/tests,
  MASTER_PLAN, root README and CHANGELOG diffs — 0;
- `git diff --check`, conflict-marker and unexpected-file checks — PASS; only
  informational LF-to-CRLF working-copy warnings were emitted.

## Coordinator Scope Audit

Final classification: **15 Required / 0 Questionable / 0 Removable**.

- TASK-041 record: **Required** for contract, baseline discovery, rework,
  verification, applicability and closure evidence;
- task index: **Required** for navigation and completed-task truth;
- DP-011 EN/RU: **Required** to remove false absent/planned managed Flow and
  continuation claims while preserving mirror parity;
- DP-013 EN/RU: **Required** to distinguish implemented managed seams from the
  absent concrete composition-private invoker;
- DP-016 EN/RU: **Required** to reconcile live authority/Stop-during-Starting
  wording with accepted Slice 3 without changing Approved/Planned-overall state;
- DP-017 EN/RU: **Required** to remove the same stale continuation premise from
  recovery design while preserving Approved/Planned-overall state;
- design indexes EN/RU: **Required** so navigation summaries match the corrected
  mirrored documents and implemented/planned boundary;
- `.ai/PROJECT_CONTEXT.md`: **Required** to remove its internal current-state
  contradiction and record completed TASK-041;
- `spec/current-state.md`: **Required** for closure and next-candidate truth;
- `spec/decisions.md`: **Required** to remove live contradiction in the durable
  decision inventory and preserve the absent-invoker boundary.

Deletion test: removing any DP mirror or index restores factual/parity drift;
removing a project-state source restores a live contradiction or loses active
task/next-candidate truth; removing either task document loses mandatory record
or navigation. No file can be deleted while preserving Definition of Done.
Production, test, generated, MASTER_PLAN, root README and CHANGELOG files are
absent from the diff. Next-task design work has not started.

## Independent Review and Rework

- initial final Reviewer verdict — `Needs Revision`, blocking B-001/B-002,
  non-blocking findings 0;
- B-001: final Scope Audit was not yet recorded before review; resolved by the
  Coordinator audit above with 15/0/0 classification and deletion mapping;
- B-002: the task record retained the narrower 891-link count; resolved by the
  fresh repository-wide 924/0 result and explicit supersession of the old
  method;
- repeat final Reviewer verdict — `APPROVED`, blocking findings 0,
  non-blocking findings 0;
- Reviewer confirmed no semantic, parity, status, scope-file, history-boundary,
  invoker-activation or TASK-026-status defect remains.

## Coordinator Acceptance and Closure

- acceptance: `Granted — 2026-08-20`;
- completed scope: exact 15-file documentation-only reconciliation; no code,
  test, architecture-status, MASTER_PLAN, public API or product capability
  change;
- PROCESS-002: `Synchronized`; EN/RU parity, links, status, stale-live sweep,
  focused regression, Scope Audit and final independent review passed;
- known limitations: concrete composition-private invoker, terminal
  publication, orchestrator and production wiring remain absent; TASK-026
  remains `Blocked by Architecture`;
- current documentation task after closure: absent;
- next candidate: bounded private exact-scope composition invoker design,
  unactivated;
- commit gate: not authorized and not performed; task commit does not exist;
- push, PR, merge, publication and branch cleanup: not authorized and not
  performed.

## Stop Conditions

Approved semantic/status change; more than 15 files; historical bulk rewrite;
claiming invoker designed/implemented; code/test/API/wiring diff; remaining
critical drift; failed mandatory check; blocking review; unexplained change.

## Next Candidate

After accepted repair, a separate bounded design-update for the private
exact-scope composition invoker. Not activated here. TASK-026 terminal
publication remains core orchestrator work.

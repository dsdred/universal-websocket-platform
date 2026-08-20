# TASK-039 — Atomic Expected-Attempt Runtime Owner Stop Design

## Status

`Completed — Coordinator Accepted`.

## Task Contract

### Task Mode

`Design-update`. Define the smallest conceptual DP-010 lifecycle contract that
allows a private orchestration caller to stop one exact expected Launch Attempt
atomically, without implementing the operation or changing existing public
management behavior.

### Why Now

- TASK-038 is completed and Coordinator Accepted with verdict
  `TASK-026 REMAINS BLOCKED`;
- its 19-row DP-016 proof reassessment identifies the missing atomic
  expected-attempt Owner Stop contract as the first bounded prerequisite;
- current `Owner.Observe()` followed by `Owner.Stop(ctx)` is TOCTOU because
  `Stop` selects the active attempt only after acquiring the Owner mutex and
  accepts no expected identity;
- the later private exact-scope composition invoker cannot close that race
  outside the Owner boundary and remains a separate subsequent candidate;
- a design decision must precede any lifecycle implementation.

### Definition of Done

1. DP-010 EN/RU mirrors define one atomic expected-attempt Stop concept that
   accepts a non-zero expected Launch Attempt identity.
2. Under the Owner mutex, the contract validates cancellation, compares the
   exact active or retained terminal attempt, and either claims/attaches to and
   converges that same attempt or returns a dedicated mismatch outcome with
   zero lifecycle mutation.
3. The contract never stops a different or newer attempt and releases the
   Owner mutex before cancellation, Host Stop, or waiting.
4. Existing `Owner.Stop(ctx)` and public DP-013 Start/Stop/Observe behavior
   remain unchanged.
5. Method naming, result/failure semantics, visibility, lifecycle
   linearization, replay/terminal behavior, and implementation constraints are
   explicit enough for a later independently testable implementation slice.
6. TASK-026 remains Blocked; no implementation, status promotion, production
   wiring, or next task is activated.
7. Required navigation and project-state documents are synchronized with
   EN/RU parity, links, scope audit, Size Guard, verification, and independent
   review passing before Coordinator Acceptance.

### Out of Scope

- production code or test changes;
- implementation of the expected-attempt Stop operation;
- concrete private composition invoker or DP-016 orchestrator;
- DP-014 terminal publication or DP-015 terminalization;
- change to Approved DP-014, DP-015, DP-016, or DP-019 semantics or status;
- public DP-013 management API changes;
- external persistence, recovery, reporting, production wiring, or Production
  Activation;
- reactivation, acceptance, commit, or publication of TASK-026;
- stage, commit, push, PR, merge, publication, or branch cleanup.

### Verification Plan

- inspect DP-010 EN/RU and current `runtimelifecycle` Owner code/tests for the
  exact mutex, attempt retention, cancellation, Stop convergence, and terminal
  replay boundaries;
- map the proposed contract back to TASK-038 findings B-001/B-002 and DP-016
  §25 proofs without weakening any Approved requirement;
- verify explicit outcomes for zero identity, cancellation before claim,
  identity mismatch, active Starting/Running/Stopping attempt, retained
  terminal attempt, newer attempt, concurrent Stop, and indeterminate caller
  cancellation;
- verify DP-010 EN/RU structural and semantic parity, design indexes, task
  index, project-state sources, MASTER_PLAN mirrors, relative links, status
  wording, conflict markers, whitespace, and `git diff --check`;
- obtain independent architecture/documentation review and repeat it after any
  rework before Coordinator Acceptance.

## Objective

Close only the first missing architecture prerequisite found by TASK-038: an
atomic Runtime Owner operation that targets exactly one expected Launch Attempt
and cannot accidentally stop its successor.

## Selection Evidence

Selected deterministically from the accepted TASK-038 Architecture Handoff,
DP-020 Slice 4 status, `spec/current-state.md`, `spec/decisions.md`, and both
MASTER_PLAN mirrors. At selection time, all sources named the same first
unactivated candidate; TASK-039 then activated only its design-update scope.

Rejected alternatives:

- resume TASK-026 — it remains architecture-blocked by the missing contract;
- implement the operation immediately — design must precede lifecycle code;
- build the private invoker first — it cannot atomically exclude other Owner
  callers or hold an external lock across Stop convergence;
- change `Owner.Stop(ctx)` semantics — existing public management behavior must
  remain compatible;
- combine design and implementation — they are independently reviewable and
  the design has unresolved API/result choices;
- start terminal publication, recovery, API, or production wiring — later
  concerns outside the first prerequisite.

## Scope

- this task record as the first content change;
- read-only inspection of current Owner code/tests and authoritative designs;
- mirrored DP-010 design update for the atomic expected-attempt Stop contract;
- only navigation and durable project-state synchronization required by
  PROCESS-002;
- no production or test files.

## Non-Goals

- no executable capability or exported API;
- no lifecycle behavior claim beyond an approved design contract;
- no orchestration implementation or product integration;
- no weakening or reclassification of DP-016 §25 proofs;
- no automatic activation of the next implementation slice.

## Sources of Truth

- PROCESS-001 and PROCESS-002;
- Approved ADR-0003 and Frozen ARCH-002 lifecycle ownership boundaries;
- DP-010 EN/RU current lifecycle Owner contract;
- Approved DP-016, especially §§8–25 and its 19 acceptance proofs;
- Approved DP-019 and Draft DP-020 prerequisite ordering;
- TASK-038 Architecture Handoff and repeat-review findings;
- current `internal/runtimelifecycle` code and tests as implementation evidence;
- `.ai/PROJECT_CONTEXT.md`, `spec/current-state.md`, `spec/decisions.md`, task
  index, design indexes, and MASTER_PLAN EN/RU.

## Roles

- Coordinator: selection, gates, Size Guard, scope audit, acceptance, and next
  recommendation;
- Architect: define the conceptual operation, ownership, linearization,
  outcomes, and implementation constraints;
- Documentation Agent: record the accepted design in DP-010 mirrors and
  synchronize only required durable status sources;
- Developer: not applicable; production changes are prohibited;
- Tester: static contract matrix, parity, links, contradiction and diff checks;
- Reviewer: independent review of architecture, parity, scope, and status;
- Publisher: not applicable without later explicit authorization.

## Branch

- trusted baseline: clean synchronized
  `main@e53c6debfaca864d51785b9e16512435dda57894`;
- task branch: `docs/task-039-expected-attempt-owner-stop-design`;
- branch action: created safely from the trusted baseline;
- this task record is the first content change;
- forbidden: stage, commit, push, merge, branch deletion, or mutation of `main`
  without the corresponding explicit gate.

## Size Guard

Expected scope is one architecture contract mirrored EN/RU plus task,
navigation, and durable project-state synchronization. No production code,
tests, package, or independently deliverable second behavior is allowed.
Crossing 15 files, introducing more than this one contract, or requiring an
Approved/Frozen semantic change triggers reassessment.

Documentation inventory found six mirrored live DP-016/DP-019/DP-020 status
locations that explicitly called this design absent or unactivated. Leaving
them unchanged after accepting TASK-039 would create project-state drift.
The initial 17-document scope therefore triggered Coordinator Size Guard
reassessment; it remains one indivisible design decision plus mandatory EN/RU
and durable-status synchronization, with zero production/test files and no
second deliverable behavior.

Reviewer rework adds one required correction to the live TASK-026 readiness
paragraph, producing an exact 18-document scope. Coordinator Size Guard
decision remains **`DO NOT SPLIT`**. All 18 documents belong
to one accepted design baseline and its mandatory mirrored and durable
contradiction synchronization. Splitting would temporarily leave live sources
calling the same design absent or unactivated. The exact scope contains zero
production/test files, zero new package, and zero second independently
deliverable behavior.

## Existing Coverage Report

- Existing Coverage: current Owner tests prove generic Stop convergence,
  cancellation, concurrent callers, retained terminal observation, and attempt
  sequencing; TASK-038 statically proves the missing expected-identity guard.
- Coverage Gap: no current API or test can atomically bind Stop to a caller's
  expected Launch Attempt.
- Added Proof Tests: not applicable; this is design-only and test changes are
  prohibited.
- Added Regression Tests: not applicable.
- Remaining Limitations: executable proof belongs to the later implementation
  task; TASK-026 remains Blocked until that implementation and the remaining
  prerequisite are independently accepted.

## Documentation Baseline

Required inventory before design editing:

- DP-010 EN/RU mirrors and design indexes;
- DP-016, DP-019, and DP-020 EN/RU references;
- task index and TASK-026/TASK-038 records;
- `.ai/PROJECT_CONTEXT.md`, `spec/current-state.md`, `spec/decisions.md`, and
  MASTER_PLAN EN/RU.

At intake all durable sources agree that the atomic expected-attempt Owner Stop
design is unactivated, TASK-026 remains Blocked, the private invoker follows,
and production integration is absent. No critical drift was found.

During task execution, architecture acceptance required a truthful transition:
TASK-039 was active and `In Progress`, Draft DP-010 contains an accepted but
Planned expected-attempt extension, the base lifecycle implementation remains
implemented in isolation, and no production or test capability has changed.
DP-016, DP-019, and DP-020 mirrors require narrow status corrections because
their live text otherwise continues to call this exact design absent or
unactivated.

## Architecture Handoff

Accepted design baseline:

- add `ErrInvalidExpectedAttempt`;
- add `StopAttemptMismatch StopOutcomeKind = "attempt-mismatch"`;
- add
  `StopExpectedAttempt(ctx context.Context, expectedAttemptID runtimeconfigload.LaunchAttemptID) (StopOutcome, error)`;
- nil Owner returns `ErrInvalidOwner`; empty expected identity returns
  `ErrInvalidExpectedAttempt`, with zero lifecycle mutation;
- after validation, under the Owner mutex, check `ctx.Err()` and select the
  relevant attempt as active when present, otherwise retained last when
  present, otherwise none; active always precedes last;
- no relevant attempt or a different relevant identity returns the valid
  negative `StopAttemptMismatch` outcome with nil error and zero mutation,
  attachment, cancellation, Host call, or wait; `Attempt()` exposes the
  captured relevant fact when present and `Failure()` is absent;
- an exact active match uses the same ordinary Stop phase semantics: Preparing
  terminalizes and cancels after unlock, Launching claims/cancels/converges,
  Running claims one Host Stop, Stopping attaches, and retained active
  StopFailed returns the exact stored failure without retry;
- with no active attempt, a matching retained Stopped or
  StoppedBeforeRunning attempt replays attempt-specific `StopStopped`; matching
  PreparationFailed or LaunchFailed performs the existing resource-free
  Failed-to-DesiredStopped/ActualStopped transition and returns exact-attempt
  `StopStopped`; an impossible matching state returns `ErrStartConflict`
  without mutation;
- cancellation visible at the locked check wins. After exact match/claim,
  caller cancellation ends only that wait while Owner-owned work converges;
- same-ID callers converge and different IDs never attach. An old retained A
  cannot match while active successor B exists;
- implementation must use one private ordinary-Stop helper shared with
  existing `Stop(ctx)`, never `Observe()` followed by `Stop()`, and release the
  mutex before cancellation, Host Stop, callback, external storage, Flow, I/O,
  or any wait;
- existing `Owner.Stop(ctx)` and public DP-013 Start/Stop/Observe semantics are
  unchanged.

Compatibility and deferrals: this task changes no Approved status or
semantics, production/test file, DP-013 public management API, private
composition invoker, orchestrator, terminal publication, persistence,
recovery, reporting, or production wiring. The base DP-010 implementation
remains implemented in isolation; only the expected-attempt extension is
Planned. TASK-026 remains `Blocked by Architecture`.

Later implementation acceptance must prove sentinel validation, mismatch with
and without a relevant attempt, active-before-last successor races, every
ordinary Stop phase, same-ID convergence and different-ID non-attachment,
exact retained StopFailed identity with no retry, retained historical terminal
cases, waiter cancellation linearization, generic Stop regression, lock and
Owner isolation, race, vet, formatting, and exported GoDoc.

## Documentation Result

Pre-implementation synchronization records the accepted design in mirrored
Draft DP-010 and preserves semantic/structural parity. Design indexes now
separate the implemented base Owner from the Planned expected-attempt
extension. Task index and durable project-state sources now identify TASK-039
as `Completed — Coordinator Accepted`, while TASK-026 remains Blocked and
implementation is explicitly absent. Narrow DP-016/DP-019/DP-020 mirror corrections remove stale
“absent/unactivated design” wording without changing their Design Status,
Approved semantics, proof matrix, or implementation status. Reviewer rework
also corrects the live TASK-026 readiness paragraph while preserving its
TASK-038 closure history.

The initial documentation handoff did not claim verification, review, Scope
Audit, Acceptance, commit, or publication. Subsequent Coordinator and Tester
evidence is recorded below. Independent final review and Coordinator
Acceptance remain open; commit and publication remain prohibited.

## Verification Matrix

- Documentation and EN/RU parity: **PASS**. DP-010 headings 35/35, fences 6/6,
  and Go blocks 5717/5717; DP-016 headings 30/30 and fences 4/4; DP-019
  headings 25/25 and fences 16/16; DP-020 headings 34/34 and fences 12/12;
  design indexes headings 1/1; MASTER_PLAN headings 36/36.
- Links: **PASS**. Changed-document links 235 valid / 0 missing;
  repository-wide links 916 valid / 0 missing.
- Repository hygiene: **PASS**. Stale live wording/status contradictions 0,
  conflict markers 0, trailing-whitespace findings 0, and `git diff --check`
  PASS with line-ending LF-to-CRLF warnings only.
- Existing lifecycle regression: `go test ./internal/runtimelifecycle -count=1`
  **PASS**; `go test ./...` **PASS**; `go vet ./...` **PASS**.
- Race detector: unavailable because `CGO_ENABLED=0`; not applicable to this
  no-code design slice and remains mandatory for the later concurrency
  implementation task.
- Public API executable proof and GoDoc: not applicable to this design-only
  task because no exported production declaration is implemented; the later
  implementation acceptance proofs require both.
- Production wiring/manual smoke, dependency/module checks, generated-artifact
  checks: not applicable because no production, test, dependency, module, or
  generated file changes exist.

Repeat Independent Tester result: **PASS**, blocking findings 0, non-blocking
findings 0. B-001 and B-002 are verified resolved. This evidence does not
replace the still-required repeat independent Reviewer verdict or Coordinator
Acceptance.

## PROCESS-002 Applicability

- this task record and task index: **Required**;
- TASK-026 live readiness paragraph: **Required** to preserve its historical
  TASK-038 context while recording subsequent TASK-039 design activation and
  the still-absent implementation;
- DP-010 EN/RU mirrors: **Required** to record the accepted design while
  distinguishing the implemented base from the Planned extension;
- design indexes EN/RU: **Required** for the same implementation-status
  boundary;
- DP-016, DP-019, and DP-020 EN/RU mirrors: **Required** to remove live
  absent/unactivated-design contradictions without changing their Approved or
  Draft semantics/statuses;
- `.ai/PROJECT_CONTEXT.md`, `spec/current-state.md`, `spec/decisions.md`, and
  MASTER_PLAN EN/RU: **Required** to record TASK-039 completed and Coordinator
  Accepted, accepted design not implemented, and TASK-026 still Blocked;
- root README and `CHANGELOG.md`: **Not applicable**; no user-facing, release,
  or root-level capability changed;
- ADR and ARCH: **Not applicable**; no Approved/Frozen architecture semantics
  or status changed;
- production code, tests, dependencies, generated artifacts: **Not
  applicable**; this is a documentation-only design-update slice.

Final PROCESS-002 result: **Synchronized** after repeat Independent Reviewer
`APPROVED` 0/0 and Coordinator Acceptance.

## Scope Audit

Coordinator deletion test after Reviewer rework: **18 Required / 0 Questionable / 0 Removable**. Exact scope: 18 documents, 17 tracked plus 1 untracked.
Repeat Tester pre-closure diff evidence: tracked 356 additions / 98 deletions; task record 399 lines; combined 755 additions / 98 deletions.

Required exact file groups:

- task record and navigation: this file and `docs/tasks/README.md`;
- active blocked-task status: TASK-026;
- accepted lifecycle design and navigation: DP-010 EN/RU and design indexes
  EN/RU;
- contradiction-free prerequisite status: DP-016, DP-019, and DP-020 EN/RU;
- durable project state: `.ai/PROJECT_CONTEXT.md`, `spec/current-state.md`,
  `spec/decisions.md`, and MASTER_PLAN EN/RU.

Deletion-test rationale: removing either DP-010 mirror loses the accepted
contract or EN/RU parity; removing an index or durable-state update misstates
implementation/task status; removing any DP-016/019/020 correction restores a
live claim that the accepted TASK-039 design is absent or unactivated; removing
the TASK-026 correction leaves the active blocked task on stale pre-TASK-039
readiness wording; removing the task record or task index loses process
traceability/navigation. No file can be removed while preserving Definition of
Done and contradiction-free PROCESS-002 synchronization.

## Independent Review

- initial verdict: **`Needs Revision`**;
- B-001: the live TASK-026 readiness paragraph and DP-019 status sentence still
  described the expected-attempt design as pending a separate design/intake —
  resolved by preserving TASK-038 as historical context while recording that
  TASK-039 subsequently activated the Draft DP-010 design and implementation
  remains absent;
- B-002: the task index did not mark the TASK-038 wording as closure-time
  history, and TASK-039 scope/process bookkeeping omitted the newly required
  TASK-026 correction — resolved by explicit historical wording and the exact
  18 Required / 0 Questionable / 0 Removable file set; Repeat Tester verified
  B-001 and B-002 resolved;
- repeat independent review: **`APPROVED`**, blocking findings 0,
  non-blocking findings 0;
- Coordinator Acceptance: **`Accepted` on 2026-08-20**.

## Closure

- Final status: `Completed — Coordinator Accepted`;
- accepted scope: one design-only expected-attempt Owner Stop contract and
  mandatory 18 Required / 0 Questionable / 0 Removable documentation sync;
- architecture: Draft DP-010 records the accepted Planned extension; the base
  remains implemented in isolation and the extension remains unimplemented;
- verification: Repeat Tester PASS 0/0, repeat Independent Reviewer
  `APPROVED` 0/0, PROCESS-002 final `Synchronized`;
- known limitations: expected-attempt Stop implementation and independent
  acceptance remain absent; private exact-scope invoker, orchestrator,
  terminal publication, persistence, recovery, reporting, and production
  wiring remain later work; TASK-026 remains `Blocked by Architecture`;
- closed by: Coordinator on 2026-08-20.

## Commit Gate

- exact command `Разрешаю коммит.` received for TASK-039: no;
- stage, commit, push, PR, merge, publication, and branch cleanup were not
  performed and remain prohibited without their explicit gates.

## Stop Conditions

- the operation requires changing Approved DP-016 or Frozen ownership rules;
- exact attempt comparison cannot be linearized under the existing Owner mutex;
- retained terminal attempts or concurrent Stop cannot be specified without a
  second architectural contract;
- existing public `Owner.Stop(ctx)` or DP-013 behavior would change;
- the task expands into code, tests, invoker, orchestrator, terminal
  publication, or production wiring;
- EN/RU sources cannot be made semantically consistent;
- unexpected worktree changes or an unresolved blocking review finding appears.

## Next Candidate

A separate bounded implementation task for the accepted expected-attempt Owner
Stop contract. It is a recommendation only and is not active. The concrete
private exact-scope composition invoker remains later in prerequisite order.

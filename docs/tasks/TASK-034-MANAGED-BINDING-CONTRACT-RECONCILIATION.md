# TASK-034 — Managed Binding Contract Reconciliation

## Status

`Completed — Coordinator Accepted`.

## Task Contract

### Task Mode

`Design-update`. This task resolves a prerequisite contract contradiction and
records an implementable repair boundary before any DP-020 Slice-3 production
work. It does not alter Approved DP-019 semantics or production code.

### Why Now

- clean synchronized baseline `main@5082a0b` has no active task;
- TASK-033 removed the previous closure drift and leaves DP-020 deferred
  Slice 3 as the explicit next recommendation;
- readiness inspection found that accepted TASK-032 implementation does not
  carry the validated orchestration authorization tuple or parent/phase
  identity required by its own Definition of Done and Approved DP-019 §14;
- Draft DP-020 §8.1 says `runtimelaunchflow` imports neither
  `runtimecommandidempotency` nor `runtimeidentity`, while the accepted
  implementation imports `runtimeidentity` for revision/generation types;
- Slice 3 cannot be implemented safely until package ownership, neutral value
  placement, and the exact binding content are unambiguous.

### Definition of Done

1. Architecture determines the authoritative package/dependency direction for
   the complete per-invocation managed binding without weakening DP-019.
2. The binding contract explicitly includes or maps every required fact:
   validated authorization tuple, expected aggregate revision, exact execution
   generation, optional parent/phase identity, and opaque Start rendezvous.
3. DP-020 EN/RU distinguishes the currently implemented partial Slice-2 seam
   from the remaining repair work and keeps Slice 3 unactivated.
4. A smallest independently testable implementation repair slice is defined
   with compatibility, ownership, failure, and verification constraints.
5. TASK-032 historical acceptance evidence remains unchanged, but current
   project state no longer represents its partial binding as the complete
   prerequisite.
6. EN/RU parity, links, status consistency, PROCESS-002, Scope Audit, and
   independent review pass with zero blocking findings.

### Out of Scope

- production code or test changes;
- DP-020 Slice-3 attempt publication/generation binding implementation;
- DP-016 orchestrator implementation or TASK-026 reactivation;
- changing Approved DP-019, DP-016, DP-014, or DP-015 semantics/statuses;
- concrete authorization policy, API, persistence, recovery, or production
  wiring;
- commit, push, PR, merge, publication, or branch cleanup.

### Verification Plan

- static comparison of Approved DP-019 §§14–17, Draft DP-020 §§8–12,
  TASK-032 contract/closure, and current package surfaces/tests;
- EN/RU structural and semantic parity checks for every changed mirror;
- repository-relative link and status-consistency validation;
- conflict-marker, trailing-whitespace, and `git diff --check` checks;
- independent architecture/documentation review.

## Objective

Produce one coherent, dependency-safe managed-binding contract and a bounded
repair handoff that restores readiness for DP-020 Slice 3.

## Selection Evidence

Selected instead of the recommended Slice 3 because readiness requires all
prerequisites to be both implemented and consistent with higher-order sources.
The contradiction is directly on Slice 3's input boundary, so its resolution
is the smallest prerequisite work.

Rejected alternatives:

- implement Slice 3 against the current partial binding — violates DP-019 and
  hides an accepted-contract gap;
- silently update documentation to match code — forbidden because
  implementation does not override Approved architecture;
- modify Approved DP-019 — no evidence requires changing its semantics;
- resume TASK-026 — it remains `Blocked by Architecture`;
- combine design reconciliation and production repair — more than one
  independently verifiable behavior.

## Scope

- TASK-034 record and task index;
- DP-020 EN/RU managed-binding/package direction and ordered-slice wording;
- project-state descriptions necessary to distinguish partial Slice 2 from
  complete prerequisite readiness;
- explicit implementation-repair next candidate.

## Non-Goals

- no production behavior, exported surface, module, or test change;
- no speculative orchestration framework;
- no activation of the repair implementation or Slice 3.

## Sources of Truth

- PROCESS-001 and PROCESS-002;
- Approved/Planned DP-019 §§14–17 and proof obligations;
- Draft/Planned DP-020 §§8–12;
- TASK-032 Task Contract, rework, review, and closure evidence;
- current `internal/runtimelaunchflow`, `runtimeconfigload`,
  `runtimecommandidempotency`, and `runtimeidentity` surfaces and tests;
- current project-state sources and MASTER_PLAN EN/RU.

## Roles

- Coordinator: selection, gates, scope audit, acceptance;
- Architect: resolve dependency/ownership contract and define repair slice;
- Documentation Agent: mirror the accepted clarification and synchronize state;
- Developer: not applicable — production changes are out of scope;
- Tester: documentation/static verification only;
- Reviewer: independent final review;
- Publisher: not applicable until separately authorized after acceptance and
  commit.

## Branch

- trusted baseline: clean `main@5082a0b` equal to `origin/main`;
- task branch: `docs/task-034-managed-binding-contract-reconciliation`;
- branch action: created safely from the trusted baseline;
- forbidden: stage, commit, push, merge, branch deletion, fetch, pull, or
  mutation of `main` without the corresponding explicit gate.

## Constraints

- Approved DP-019 semantics are preserved;
- command permits and mutable rendezvous state remain command-boundary owned;
- Flow receives no permit, mutable aggregate state, or cancellation authority;
- package dependencies remain acyclic and explicit;
- planned work is not described as implemented.

## Stop Conditions

- reconciliation requires changing Approved DP-019 semantics;
- no dependency-safe placement can carry the required immutable facts;
- materially different repair contracts remain after architecture analysis;
- unexpected worktree changes or documentation contradictions appear;
- independent review leaves a blocking finding unresolved.

## Acceptance Criteria

1. One architecture handoff closes the contradiction without production work.
2. Mirrored DP-020 and project-state wording truthfully describe the partial
   implementation boundary.
3. The next repair slice is bounded, testable, and explicitly not started.
4. Verification and independent review pass.

## Existing Coverage Report

- Existing Coverage: TASK-032 tests prove validation of the current
  revision/generation/rendezvous value, OwnerClaimView mapping, managed Flow
  call-stack lifetime, failure convergence, and unchanged unmanaged behavior.
- Coverage Gap: no test can prove authorization-tuple or parent/phase identity
  binding because those facts are absent; current tests permit the
  `runtimeidentity` import that DP-020 §8.1 forbids.
- Added Proof Tests: not applicable in this design-update task.
- Added Regression Tests: not applicable in this design-update task.
- Remaining Limitations: implementation conformance remains pending a separate
  repair task and Slice 3 stays blocked.

## Documentation Baseline

Status before synchronization: `Drift Detected`.

- Approved DP-019 requires the exact authorization tuple including
  `OperationalDomain`, a per-invocation binding carrying authorization and
  linked identity, and command-owned rendezvous semantics.
- TASK-031 and TASK-032 records truthfully preserve their closure-time
  contracts, reviews, acceptance, and publication evidence; those
  time-scoped historical statements are not rewritten.
- Live DP-020, project-state, and roadmap wording incorrectly treated the two
  accepted partial seams as the complete prerequisites and named Slice 3 as
  next without a conformance repair.
- DP-020 EN/RU had matching Draft/Planned statuses and mirror links, but §7.2,
  §8.1, and the live implementation boundary contradicted Approved DP-019 and
  current package dependencies.

## Architecture Handoff

Architect verdict: `PASS`, blocking findings `0` after focused reconciliation.

- The single authoritative immutable cross-layer contract belongs in new
  dependency-leaf `internal/runtimeorchestrationbinding`, which may depend only
  on `runtimeconfigload` identity types. Command/idempotency storage,
  management composition, identity conversion, and Flow may consume it;
  `runtimelaunchflow` imports neither `runtimecommandidempotency` nor
  `runtimeidentity`.
- `OrchestrationAuthorizationRequest` retains all six Approved DP-019 fields,
  including exact non-empty opaque `OperationalDomain`; no default or inference
  is allowed.
- `StartExecutionBinding` carries that authorization, neutral lossless expected
  revision and execution generation, an all-or-none parent plus derived
  `StartTarget` phase identity for linked execution, and one unique opaque
  rendezvous identity. Primitive Start carries neither parent nor phase.
- The DP-015 command boundary owns the generation-bound mapping from rendezvous
  identity to concrete mutable state. Missing, forged, reused, mismatched, or
  cross-generation identities yield `Blocked`; the binding carries no permit,
  channel, callback, pointer, or mutable state.
- Continuation outcomes remain exactly `Continue`, `StopConverged`,
  `BindingFailed`, and `Blocked`; Approved DP-019 semantics and statuses do not
  change.
- The next implementation candidate is Slice 2R only. It does not implement
  DP-014 publication/binding, continuation outcome behavior, orchestration,
  policy, persistence, API, or production wiring. Slice 3 remains blocked by
  independent implementation and acceptance of Slice 2R.

### Architecture Rework — Independent Review B-001 / N-001

- `B-001` is resolved by one exact additive primitive adapter owned by
  `runtimecommandidempotency.Boundary`:
  `ExecuteManagedStart(ctx, Start Scope, CommandKey, Start Intent,
  ExpectedAggregateRevision, ExecutionGeneration, AuthorizeOrchestration,
  invoke(StartExecutionBinding) (TerminalOutcome, error)) (Admission, error)`.
- It validates all inputs, derives and authorizes the exact six-field
  `ActivateExactTarget` request, and checks cancellation before the shared
  primitive admission point. Authorization is repeated before inspection for
  initial, in-progress, and replay submissions.
- Only a new primitive claim atomically installs the command record, live
  permit, unique generation-bound rendezvous identity/private lookup, and
  complete primitive binding under the admission lock. The callback runs once
  synchronously outside locks. Replay/in-progress, including a record created
  by legacy `Execute`, receives no callback, binding, new rendezvous, permit, or
  adoption of execution authority.
- Callback/permit return, panic, `runtime.Goexit`, or generation loss expires
  binding and lookup authority. Error, panic, invalid outcome, missing
  publication, or indeterminacy preserves Claimed/unresolved truth and returns
  the existing indeterminate error; no retry or detached work exists.
- Existing `Boundary.Execute` remains byte-for-behavior compatible for isolated
  legacy callers. Production orchestration must use `ExecuteManagedStart` and
  has no fallback through `Execute`, synthesized binding, or direct private-
  invoker bypass. DP-013 public and existing DP-015 surfaces remain unchanged.
- `N-001` is resolved by listing Slice 2R explicitly in DP-020 §6 EN/RU and
  fixing the dependency order to `Slice 1 -> Slice 2 -> Slice 2R -> Slice 3 ->
  Slice 4`.
- Approved DP-019 remains unchanged.

## Size Guard

- projected files: 9 documentation files;
- production lines: 0; new packages: 0; architecture contracts: 1 refined
  Draft DP-020 contract; independently shipped behaviors: 1 design outcome;
- decision: `ACCEPT`, no split required.

## Documentation Sync

Status: `Synchronized` after architecture rework B-001/N-001.

Resolved drift:

- DP-020 EN/RU now preserve Draft/Planned overall status while distinguishing
  historically accepted partial Slices 1–2, Planned/unactivated Slice 2R, and
  Planned Slice 3 blocked by Slice 2R;
- both mirrors record the dependency-leaf binding ownership, exact six-field
  authorization, all-or-none linked identity, and command-owned unique
  rendezvous identity without changing Approved DP-019;
- architecture rework disposition is synchronized: B-001 adds the sole
  primitive `Boundary.ExecuteManagedStart` claim-to-binding adapter with no
  legacy adoption/fallback, and N-001 makes the order explicitly
  `1 -> 2 -> 2R -> 3 -> 4`; both findings are resolved in DP-020 EN/RU;
- `.ai/PROJECT_CONTEXT.md`, `spec/current-state.md`, `spec/decisions.md`, and
  MASTER_PLAN EN/RU use the same live prerequisite and next-candidate boundary;
- TASK-031/TASK-032 historical closure and acceptance statements remain
  unchanged; the task index labels them as historically accepted partial
  prerequisites rather than complete current readiness.

PROCESS-002 applicability:

- TASK-034 record and `docs/tasks/README.md`: required and updated for active
  task navigation, handoff, applicability, and verification evidence;
- DP-020 EN/RU: required and synchronized because this task refines its Draft
  contract and current implementation boundary; Design Status remains Draft
  and Implementation Status remains Planned overall;
- `spec/current-state.md`, `spec/decisions.md`, and `.ai/PROJECT_CONTEXT.md`:
  required and synchronized because live capability, blocker, active-task, and
  next-candidate statements changed;
- MASTER_PLAN EN/RU: required and synchronized because the durable milestone
  dependency now includes Slice 2R before Slice 3;
- DP-019 EN/RU: inspected, not edited; Approved semantics/status are unchanged;
- DP-013–DP-018, design indexes, root README EN/RU, and `CHANGELOG.md`: `Not
  applicable` because no status, user-facing/release capability, index entry,
  or their contract changed;
- production code, tests, `go.mod`, and `go.sum`: `Not applicable`; this task is
  documentation-only in effect and changes no executable artifact.

## Verification

Documentation/static evidence completed:

- exact changed-file set contains the nine provisional Required documentation
  files and no production, test, module, generated, or unrelated artifact;
- repository-relative links from the exact nine-file scope: `0` missing;
- DP-020 EN/RU structural parity after rework: `33/33` headings and `6/6`
  code-fence lines;
- `ExecuteManagedStart` occurs `3/3` times in EN/RU, and the ordered dependency
  statement explicitly contains Slices `1, 2, 2R, 3, 4` in both mirrors;
- DP-020 status parity: Design Status `Draft`, Implementation Status `Planned`
  overall; Slice 2R Planned/unactivated and Slice 3 Planned/blocked in both;
- conflict-marker scan: `0`; trailing-whitespace scan: `0`;
- targeted live-state scan confirms TASK-031/TASK-032 historical partial
  acceptance, Slice 2R as next unactivated repair, Slice 3 blocked by Slice 2R,
  TASK-026 `Blocked by Architecture`, and no live state bypasses the managed
  adapter through legacy `Execute`;
- `git diff --check`: `PASS`; only configured LF-to-CRLF working-copy warnings
  were emitted;
- executable tests, race, vet, formatter, and module checks: `Not applicable`
  because no production code, tests, or module files changed;
- repeat Independent Review after architecture rework: `Approved`, blocking
  findings `0`, non-blocking findings `0`; B-001 and N-001 are resolved;
- the Documentation Agent did not perform the independent review.

## Scope Audit

Final Coordinator classification after repeat Independent Review:

| # | File | Classification | Rationale |
|---|---|---|---|
| 1 | `docs/tasks/TASK-034-MANAGED-BINDING-CONTRACT-RECONCILIATION.md` | `Required` | task contract, architecture handoff, sync and verification evidence |
| 2 | `docs/tasks/README.md` | `Required` | active task and corrected partial-prerequisite navigation |
| 3 | `docs/en/design/DP-020-runtime-orchestration-binding-sequence-readiness.md` | `Required` | authoritative Draft clarification and Slice 2R boundary |
| 4 | `docs/ru/design/DP-020-runtime-orchestration-binding-sequence-readiness.md` | `Required` | semantic mirror of DP-020 clarification |
| 5 | `.ai/PROJECT_CONTEXT.md` | `Required` | live task/capability/next-candidate state |
| 6 | `spec/current-state.md` | `Required` | factual current boundary and active documentation task |
| 7 | `spec/decisions.md` | `Required` | live summary of Approved DP-019 versus Draft DP-020 implementation readiness |
| 8 | `docs/en/roadmap/MASTER_PLAN.md` | `Required` | durable milestone dependency ordering |
| 9 | `docs/ru/roadmap/MASTER_PLAN.md` | `Required` | mirror of milestone dependency ordering |

Final totals: `9 Required / 0 Questionable / 0 Removable`.

## Process Health

- trigger: yes — post-acceptance evidence exposed a contract/conformance gap;
- finding: TASK-032 review and closure failed to map the implementation against
  every Definition-of-Done field, so incomplete conformance reached `main`;
- PROCESS-001 already requires traceability, scope verification, and final
  independent review against the Task Contract. This is an execution finding,
  not evidence for a process-contract change; no process document is edited.

## Commit Gate

- exact command `Разрешаю коммит.` received: no;
- commit, stage, push, PR, merge, and publication are not authorized.

## Next Candidate

- DP-020 Slice 2R implementation repair of the complete dependency-safe
  managed binding and sole primitive `Boundary.ExecuteManagedStart` adapter;
- Slice 2R is recommended but not activated by this closure;
- DP-020 Slice 3 follows only after Slice 2R is independently accepted and
  remains blocked and unactivated.

## Closure

- Final status: `Completed — Coordinator Accepted`.
- Architecture rework findings B-001/N-001: resolved.
- Repeat Independent Review: `Approved`, blocking `0`, non-blocking `0`.
- Verification, PROCESS-002, EN/RU parity, link validation, status consistency,
  and final Scope Audit `9/0/0`: `PASS`.
- Process Health: execution finding recorded; no PROCESS-001/PROCESS-002
  contract change required.
- TASK-026 remains `Blocked by Architecture`; Slice 2R is the next
  unactivated candidate and Slice 3 remains blocked by it.
- Closed by: Coordinator.
- Date: 2026-08-11.

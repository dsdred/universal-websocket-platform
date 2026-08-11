# TASK-036 — Slice 3 Readiness Reconciliation

## Status

`Completed — Coordinator Accepted`.

## Task Contract

### Task Mode

`Design-update`. This task resolves the concrete cross-package rendezvous and
continuation protocol needed before DP-020 Slice 3 can be implemented. It does
not change production code or activate Slice 3 implementation.

### Why Now

- TASK-035 is published and independently accepted, so DP-020 names Slice 3 as
  the next Planned prerequisite;
- readiness inspection found that the Slice-2R opaque rendezvous lookup proves
  only callback liveness and exposes no closed early/final gate operation;
- the existing pending-Stop rendezvous and its `OwnerClaimed` signal belong to
  the parent `StartTarget` phase path, while `ExecuteManagedStart` currently
  creates only primitive Start bindings and rejects a pending phase rendezvous;
- DP-020 requires both primitive Start and linked `StartTarget` to use the same
  four-outcome continuation without transferring command permits or mutable
  rendezvous state across package boundaries;
- implementing Slice 3 before fixing this ownership/API ambiguity would require
  the Developer to invent architecture.

### Definition of Done

1. Architecture specifies one exact command-owned early rendezvous resolution
   and final Stop-versus-Continue gate for primitive and linked managed Start.
2. The private continuation owner, construction inputs, package dependencies,
   closed outcome representation, error mapping, and callback lifetime are
   unambiguous and compatible with DP-014, DP-015, and DP-019.
3. The design states how primitive and linked command paths construct complete
   bindings and how the opaque rendezvous resolves only the exact live permit
   and Boundary generation.
4. The fixed sequence distinguishes definitive `BindingFailed` from stale,
   conflicting, unavailable, unknown, or indeterminate `Blocked` evidence and
   defines exact revision/generation conversion at the identity-store boundary.
5. A smallest independently testable Slice-3 implementation handoff is defined
   without adding an orchestrator, policy, API, recovery, or production wiring.
6. EN/RU parity, project-state consistency, PROCESS-002, Scope Audit, and
   independent review pass with zero blocking findings.

### Out of Scope

- production code or test changes;
- DP-020 Slice 3 or Slice 4 implementation;
- TASK-026 reactivation;
- changes to Approved DP-019, DP-015, or DP-014 semantics;
- concrete authorization policy, DP-016 orchestrator, API, persistence,
  recovery, or production wiring;
- stage, commit, push, PR, merge, publication, or branch cleanup.

### Verification Plan

- compare Approved DP-019 §§14–18, DP-015 pending-Stop rules, DP-014
  conditional operations, Draft DP-020 §§8–12, and current package surfaces;
- verify EN/RU structural and semantic parity for every changed mirror;
- validate relative links, status wording, conflict markers, whitespace, and
  `git diff --check`;
- perform independent architecture/documentation review.

## Objective

Produce an implementable, dependency-safe Slice-3 continuation and rendezvous
contract that removes the remaining need for production-code design decisions.

## Selection Evidence

The candidate comes from the explicit post-TASK-035 recommendation: DP-020
Slice 3 OwnerClaim-to-DP-014 binding. It is not directly Ready because its
required command-owned early/final rendezvous protocol is absent from both the
implemented surfaces and the exact design API.

Selected this bounded readiness reconciliation instead of implementation because
it is the lowest unresolved prerequisite. Rejected alternatives:

- implement identity-store writes only — would omit mandatory Stop ordering;
- expose `StartTargetExecution` or a permit to Flow — violates ownership and
  dependency direction;
- treat a live opaque handle as proof that no Stop is pending — violates the
  fail-closed contract;
- invent the full DP-016 orchestrator or production wiring — out of sequence;
- resume TASK-026 — it remains `Blocked by Architecture`.

## Scope

- this task record and task index;
- DP-020 EN/RU exact Slice-3 command/continuation protocol;
- narrow factual implementation-boundary synchronization in DP-015 and DP-019
  EN/RU; their Approved semantics and statuses remain unchanged;
- project context, current state, decision log, and EN/RU roadmaps necessary to
  record the clarified implementation boundary and next candidate;
- no production or test files.

## Non-Goals

- no executable capability or public API;
- no change to historical TASK-031 through TASK-035 acceptance evidence;
- no automatic start of the resulting implementation task;
- no speculative general-purpose coordination framework.

## Sources of Truth

- PROCESS-001 and PROCESS-002;
- Approved DP-019 §§14–18;
- Approved DP-015 pending-Stop ownership and ordering;
- Approved DP-014 conditional attempt/generation operations;
- Draft DP-020 §§8–12;
- accepted TASK-035 implementation and tests;
- current `runtimeorchestrationbinding`, `runtimecommandidempotency`,
  `runtimeidentity`, and `runtimelaunchflow` surfaces.

## Roles

- Coordinator: selection, gates, scope audit, acceptance;
- Architect: exact ownership/API/failure handoff;
- Documentation Agent: mirrored design and state synchronization;
- Developer: not applicable; production changes are out of scope;
- Tester: static, parity, link, and contradiction verification;
- Reviewer: independent final review;
- Publisher: not applicable without later explicit authorization.

## Branch

- trusted baseline: clean synchronized
  `main@b687174c64b7638daafca61b08b9c29ea129764f`;
- task branch: `docs/task-036-slice-3-readiness-reconciliation`;
- branch action: created safely from the trusted baseline;
- forbidden: stage, commit, push, merge, branch deletion, or mutation of `main`
  without the corresponding explicit gate.

## Constraints

- command permits, mutable rendezvous state, and Boundary generation stay in
  `runtimecommandidempotency`;
- `runtimelaunchflow` imports neither command idempotency nor runtime identity;
- the continuation is synchronous, stateless between invocations, and receives
  no lifecycle preparation token or mutable Owner state;
- DP-014 remains the sole owner of aggregate/attempt facts;
- planned work is never described as implemented.

## Stop Conditions

- reconciliation would require changing Approved DP-019 semantics;
- primitive and linked paths require materially different architecture rather
  than one bounded protocol;
- a dependency-safe closed protocol cannot be specified without production
  composition decisions outside the current sources;
- unexpected worktree changes or conflicting authoritative sources appear;
- independent review leaves a blocking finding unresolved.

## Acceptance Criteria

1. One architecture handoff closes the command/continuation ambiguity.
2. Mirrored design text defines an implementation-ready Slice 3 and truthful
   current-state boundary.
3. Verification, PROCESS-002, Scope Audit, and independent review pass.

## Architecture Handoff

Architect verdict: `PASS`, blocking findings `0`. Approved DP-014, DP-015, and
DP-019 permit one common primitive/linked protocol; no Approved semantic or
status change is required.

- `runtimecommandidempotency` remains the sole owner of primitive/phase
  permits, Boundary generation, mutable rendezvous, handle resolution, and
  early/final gates; it performs no DP-014 writes.
- A future focused `internal/runtimeorchestrationcontinuation` package owns the
  stateless long-lived continuation implementation and imports the neutral
  binding, command, identity, and launch-flow packages; none imports back.
- Primitive orchestration keeps `ExecuteManagedStart`. Replacement/rollback
  adds a managed parent adapter and managed StartTarget method that alone
  constructs a complete linked binding; legacy records cannot be adopted.
- The command boundary exposes closed early resolution, final disposition, and
  no-claim signaling over the complete binding. It resolves the opaque handle
  internally to exact command/phase identity and live generation; it exposes no
  key, permit, channel, pointer, or wait object.
- One shared `PreOwner -> Binding -> sealed final disposition` rendezvous
  orders the original Stop claimant before Owner claim and at the final
  Continue-versus-BindingFailed gate for primitive and linked execution.
- `runtimelaunchflow` retains the interface and maps exactly Continue,
  StopConverged, BindingFailed, and Blocked without importing command or
  identity packages or passing its authentic preparation token outward.
- Caller cancellation is consumed only before Owner claim. Continuation,
  no-claim signalling, and same-preparation Owner convergence use a local
  non-cancelable context after that gate, preserving exact outcome convergence.
- DP-014 claim success threads only its returned revision into generation
  binding. Error/noncommit inspection uses
  `ReadInstance A -> history -> ReadInstance B`; changed or conflicting facts
  are Blocked, never retried or repaired.
- Slice 3 ends at the continuation decision and local Flow convergence. It does
  not claim later DP-014 terminal publication or DP-015 command/phase
  terminalization.
- No command or identity lock crosses a wait, the other store, continuation,
  Flow, Owner, Load/Build/Host, or external work. Callback return, panic,
  Goexit, or generation loss expires authority and wakes waiters Blocked.

## Existing Coverage Report

- Existing Coverage: TASK-035 tests prove six-field and primitive/linked leaf
  validation, primitive managed claim-to-binding, unique callback-scoped opaque
  lookup, authorization/replay/legacy non-adoption, expiry/failure paths, exact
  Flow request mapping, and one authentic PrepareStart. Parent/rendezvous tests
  prove the legacy linked `OwnerClaimed`/`StartNoClaim` protocol. Identity-store
  tests prove isolated conditional claim, revision threading, same-generation
  satisfied observation, coherent reads, rejection, and concurrency behavior.
- Coverage Gap: no managed linked adapter, common primitive/linked early gate,
  final Stop-versus-disposition gate, concrete four-outcome continuation,
  integrated OwnerClaim-to-DP-014 sequence, indeterminate-store seam, or
  outcome-specific Flow proof exists. These executable proofs belong to the
  successor implementation task.
- Added Proof Tests: not applicable to this design-update.
- Added Regression Tests: not applicable to this design-update.
- Remaining Limitations: Slice 3, concrete private composition invoker, and
  production wiring remain absent.

## Documentation Baseline

Status at intake: `Drift Detected`.

- DP-020 §10 incorrectly named legacy `Boundary.Execute` as the sole primitive
  admission despite accepted `ExecuteManagedStart`.
- DP-020 lacked the managed linked adapter, exact early/final gate operations,
  closed result/error mapping, continuation package owner, identity-store test
  seam, and revision-sandwich convergence rule.
- Approved DP-015 and DP-019 retained stale factual statements that the exact
  authorization/binding and managed Flow/OwnerClaimView seams were wholly
  absent. Their normative semantics/statuses remain correct and unchanged.
- Project-state sources correctly kept Slice 3 Planned but did not record the
  readiness gap or active reconciliation.

## Verification

- Existing Coverage Report: complete above;
- Verification Matrix: DP authority/contradiction table; EN/RU semantic and
  structural parity; relative links and mirror backlinks; stale-state grep;
  static code-name/import audit; exact documentation-only file set; conflict,
  whitespace, and diff checks;
- documentation structure: DP-020 EN/RU headings 34/34 and fences 12/12;
  changed-document relative links 0 broken; conflict markers and trailing
  whitespace 0; `git diff --check` PASS;
- regression: focused baseline PASS from Tester; full `go test ./...` PASS;
  `go vet ./...` PASS;
- static scope: 13 documentation files, zero `.go`, test, `go.mod`, or `go.sum`
  changes; no historical TASK-031–TASK-035 record edited;
- race: not applicable to a documentation-only diff; existing concurrency
  behavior is unchanged;
- independent review: `Approved`, blocking findings 0, non-blocking findings 0.

## Scope Audit

Final classification: 13 `Required`, 0 `Questionable`, 0 `Removable`.

- task record/index are required for PROCESS-001 discoverability;
- DP-020 EN/RU are required for the architecture handoff;
- DP-015/DP-019 EN/RU are required to remove factual implementation-boundary
  drift found while using those documents as authority;
- PROJECT_CONTEXT, current-state, decisions, and MASTER_PLAN EN/RU are required
  to keep active-task, durable dependency, and next-candidate state consistent.

Deletion test: removing any pair or state record reintroduces a mirrored
contract, factual-boundary, task-discovery, or roadmap contradiction. No file
can be removed while preserving the Definition of Done.

## Size Guard

- actual: 13 documentation files, zero production lines, zero new package, one
  clarified architecture contract, zero shipped behavior;
- decision: no Size Guard trigger; the design update remains bounded.

## Documentation Sync

- task record: synchronized; created first;
- `spec/current-state.md`: synchronized for active task and next dependency;
- MASTER_PLAN EN/RU: synchronized for the durable Slice-3 dependency;
- DP-020 EN/RU: synchronized with the exact handoff;
- DP-015/DP-019 EN/RU: synchronized narrowly for factual implementation
  boundary; Approved semantics/statuses unchanged;
- `.ai/PROJECT_CONTEXT.md`: synchronized for active task and next candidate;
- task index and `spec/decisions.md`: synchronized;
- CHANGELOG: not applicable because no user-facing or release behavior changes.
- PROCESS-002 status: `Synchronized`; independent review confirmed parity,
  links, status consistency, scope, and factual boundaries.

## Commit Gate

- exact command `Разрешаю коммит.` received for this task: yes, after
  Coordinator Acceptance on 2026-08-12;
- commit message: `docs: reconcile TASK-036 slice 3 readiness`;
- exact file set: accepted 13 documentation files recorded by Scope Audit;
- post-acceptance diff: only this Commit Gate bookkeeping update;
- temporary/generated/unrelated files: none;
- final checks: full tests, vet, links, parity, status consistency, and
  `git diff --check` PASS;
- authorized action: stage the exact accepted file set and create one task
  commit; push, PR, merge, publication, and branch cleanup remain prohibited.

## Process Health

- trigger applicable: no process defect observed; the issue is task-specific
  architecture readiness;
- process changes: none planned.

## Handoff

- completed scope: exact Slice-3 primitive/linked command-gate, continuation,
  Flow outcome, cancellation, and identity-convergence contract;
- changed files: 13 documentation files; no production or test changes;
- verification: full tests, vet, parity, links, hygiene, status consistency,
  PROCESS-002, Size Guard, and Scope Audit PASS;
- review: Approved 0/0;
- open risks: Slice 3 remains unimplemented; concrete private composition
  invoker and production wiring remain absent;
- next permitted step: a new bounded implementation intake for Slice 3.

## Publication

- publication readiness is separate from completion;
- Publisher P0–P10: not authorized.

## Next Candidate

- recommended Ready work: bounded Slice-3 implementation defined by this task;
- readiness evidence: Architect PASS 0 blockers, Tester coverage matrix,
  Independent Review Approved 0/0, PROCESS-002 and Scope Audit PASS;
- explicitly not started.

## Closure

- Final status: `Completed — Coordinator Accepted`;
- Closed by: Coordinator after independent Approved review;
- Date: 2026-08-12.

# TASK-037 — Runtime Owner Claim Binding

## Status

`Completed — Coordinator Accepted`.

## Task Contract

### Task Mode

`Implementation`. Implement DP-020 Slice 3 end to end in isolation: managed
primitive and linked command gates, the concrete stateless OwnerClaim-to-DP-014
continuation, and exact managed Flow outcome adaptation before Load.

### Why Now

- TASK-035 Slice 2R and TASK-036 readiness reconciliation are completed,
  independently accepted, committed, and published;
- DP-020 identifies Slice 3 as the single next Planned implementation slice;
- TASK-036 fixed exact package ownership, adapter/gate surfaces, lifecycle
  mapping, neutral values, revision threading, inspection, and failure rules;
- the current repository still has only a primitive live-handle proof, a legacy
  linked rendezvous capability, an error-only continuation interface, and no
  OwnerClaim-to-DP-014 composition;
- this slice is the lowest remaining DP-019 prerequisite before Slice 4
  orchestration-readiness reassessment.

### Definition of Done

1. Primitive `ExecuteManagedStart` and additive managed parent/StartTarget paths
   create complete primitive/linked bindings and one exact command-owned live
   rendezvous without adopting legacy records.
2. Closed early, final, and no-claim command operations resolve only the exact
   opaque handle, binding identity, permit, and active Boundary generation.
3. A new stateless continuation performs early Stop ordering, exact DP-014
   attempt claim and generation bind with revision threading, fail-closed
   revision-sandwich convergence, and final Stop-versus-disposition ordering.
4. Managed Flow validates primitive and linked bindings, consumes caller
   cancellation only before Owner claim, and maps Continue, StopConverged,
   BindingFailed, and Blocked exactly without Load on non-Continue outcomes.
5. Command/identity/lifecycle locks never cross callbacks, waits, other stores,
   Flow/Owner, Load/Build/Host, or external work; callback expiry wakes waiters
   Blocked and leaves unresolved durable claims.
6. Existing public DP-013, runtimeidentity Store, unmanaged Flow, and legacy
   Execute/ExecuteParent/ContinueOrExecuteStartTarget behavior remain unchanged.
7. Focused, full, stress, race where available, vet, dependency, documentation,
   PROCESS-002, Scope Audit, Size Guard, and independent review pass with zero
   blocking findings.

### Out of Scope

- DP-014 terminal publication and DP-015 parent/phase terminalization after the
  exact Owner result;
- DP-020 Slice 4, DP-016 orchestrator, TASK-026 reactivation, production
  composition/private invoker wiring, concrete authorization policy, API,
  external persistence, recovery, reporting, or deployment activation;
- changes to Approved DP-014/015/019 semantics or public DP-013 surfaces;
- stage, commit, push, PR, merge, publication, or branch cleanup.

### Verification Plan

- establish Existing Coverage Report before changing tests;
- add proof matrices for primitive/linked construction, early/final gates,
  no-claim, DP-014 claim/bind/inspection, Flow outcomes, expiry, and races;
- preserve all legacy command, identity, lifecycle, management, and unmanaged
  Flow tests;
- run focused tests with stress/shuffle, full tests, vet, race when available,
  gofmt, dependency/import audit, module diff, and documentation checks.

## Objective

Implement one fail-closed synchronous OwnerClaim-to-DP-014 binding sequence
that releases Load only after exact same-generation binding and final command
gate confirmation for primitive and linked managed Start.

## Selection Evidence

Selected from the explicit accepted TASK-036 Next Candidate and DP-020 Slice 3
status. Prerequisites TASK-031, TASK-032, TASK-035, and TASK-036 are accepted;
the baseline is clean and synchronized.

Rejected alternatives:

- start Slice 4 or resume TASK-026 — both remain gated by Slice 3 acceptance;
- implement identity writes without command gates — violates Stop ordering;
- implement gates without the continuation — produces no usable/provable
  OwnerClaim-to-Load behavior and a false-ready intermediate;
- add production composition or policy — later scope;
- modify runtimeidentity semantics — existing conditional operations are the
  authoritative boundary and already sufficient.

## Scope

- `internal/runtimeorchestrationbinding`: neutral no-claim cause only if needed
  by the approved dependency direction;
- `internal/runtimecommandidempotency`: managed parent/phase adapter, shared
  managed rendezvous state and closed gate operations, preserving legacy paths;
- new focused `internal/runtimeorchestrationcontinuation`: stateless concrete
  continuation and narrow four-method identity boundary;
- `internal/runtimelaunchflow`: closed continuation result and exact outcome /
  non-cancelable convergence mapping;
- focused tests in those packages plus unchanged-regression proof;
- task/state/design/roadmap documentation required by PROCESS-002.

## Non-Goals

- no transport/API/public product capability;
- no generic orchestration framework;
- no global registry, detached goroutine, permit transfer, retry, repair, or
  inferred identity/generation;
- no automatic activation of the next task.

## Sources of Truth

- PROCESS-001 and PROCESS-002;
- Approved DP-014 conditional persistence semantics;
- Approved DP-015 command/permit/pending-Stop semantics;
- Approved DP-019 §§14–18 lifecycle and ordering contract;
- Draft DP-020 §§8–12 as reconciled and accepted by TASK-036;
- TASK-036 Architecture Handoff and Existing Coverage Report;
- current package code and tests.

## Roles

- Coordinator: selection, gates, scope audit, acceptance;
- Architect: confirm exact implementation mapping before code;
- Developer: implement only the confirmed plan and tests;
- Tester: Existing Coverage Report and independent verification matrix;
- Documentation Agent: PROCESS-002 synchronization after proof;
- Reviewer: independent final review;
- Publisher: not applicable without later explicit authorization.

## Branch

- trusted baseline: clean synchronized
  `main@b8f27439e06821eb6704de81ce75716628ce428e`;
- task branch: `feature/task-037-runtime-owner-claim-binding`;
- branch action: created safely from trusted baseline;
- forbidden: stage, commit, push, merge, branch deletion, or mutation of `main`
  without the corresponding explicit gate.

## Constraints

- no command or identity lock across a callback, wait, other store, Owner, or
  external work;
- command owns mutable rendezvous and generation; identity owns durable facts;
  Owner remains sole lifecycle authority; Flow holds authentic preparation;
- `runtimelaunchflow` imports neither command nor identity;
- bindings contain no permit, pointer, channel, callback, preparation token, or
  mutable state and remain callback-scoped;
- legacy records are observed only, never adopted;
- caller cancellation after Owner claim cannot suppress exact convergence;
- no blind retry, repair, fallback, or reclassification of unknown evidence.

## Stop Conditions

- implementation requires an architectural choice not fixed by TASK-036;
- current Store cannot satisfy the documented narrow boundary without semantic
  change;
- primitive and linked protocols cannot share the confirmed state machine;
- baseline becomes dirty from unattributed work or scope must materially expand;
- required tests fail or independent review leaves a blocking finding.

## Acceptance Criteria

1. Exact primitive and linked bindings reach one concrete continuation only on
   newly committed managed authority.
2. Early/final gate races produce only Clear/Continue, StopConverged,
   BindingFailed, or Blocked under the documented ordering.
3. DP-014 facts are committed before Load with exact revision/generation proof.
4. Every non-Continue Flow path performs zero Load and truthful Owner
   convergence without falsely terminalizing the command.
5. All compatibility, concurrency, documentation, and process gates pass.

## Existing Coverage Report

Baseline focused tests and a 20-run stress baseline pass. Existing tests prove
the immutable primitive/linked binding values, primitive managed admission and
handle lifetime, legacy parent/StartTarget rendezvous ordering, isolated
DP-014 claim/bind/read semantics, authentic Owner preparation, and the current
primitive managed Flow path.

The uncovered Slice-3 behavior is the managed parent/StartTarget adapter, the
shared primitive/linked managed rendezvous state, closed early/final/no-claim
gates, the concrete four-method identity continuation and revision sandwich,
the four managed Flow outcomes, and post-Owner-claim cancellation suppression.
The implementation tests must cover both binding variants, exact authorization
and non-adoption, Stop races at every gate, no-claim causes, forged/stale/expired
handles, exact claim-result revision threading, adversarial coherent reads,
non-Continue zero-Load behavior, callback expiry, legacy compatibility, lock
boundaries, instance independence, stress, and race where available.

## Architecture Handoff

`PASS`, zero blocking findings. The implementation is one indivisible behavior:
splitting the command gates from the continuation would create a false-ready
intermediate, while the continuation cannot be proved without those gates.

- `runtimeorchestrationbinding` owns the closed neutral Cancelled, Rejected,
  and Failed no-claim causes; historical command names remain aliases.
- `runtimecommandidempotency` owns an extended shared managed rendezvous with
  `PreOwner -> Binding -> sealed/blocked` state, exact opaque lookup, primitive
  and linked live permits, and closed early/final/no-claim operations. Invalid
  input is rejected as invalid submission; stale Boundary generation as
  boundary expiry; forged, reused, mismatched, lost, or unproven authority as
  indeterminate execution. A blocked gate always carries a non-nil error.
- `ExecuteManagedParent` authorizes exact Replace/Rollback before every
  inspection and is the only source of `ManagedParentExecution`. Its managed
  StartTarget operation derives the exact parent plus ordinal-1 linked binding
  and invokes only for a newly committed managed phase. Legacy parent records
  and capabilities cannot be upgraded.
- the new stateless `runtimeorchestrationcontinuation` depends on Boundary plus
  a narrow identity interface containing only conditional claim, conditional
  generation bind, instance read, and attempt-history read. It validates and
  converts neutral values losslessly, resolves the early gate, threads only a
  committed claim result revision into bind, never retries, classifies uncertain
  results through an A/history/B revision sandwich, then resolves the final
  Continue or BindingFailed disposition.
- exact same attempt/version/generation evidence may converge Continue;
  coherent absence at the relevant unchanged revision may produce
  BindingFailed; empty generation after an indeterminate claim, stale or
  conflicting facts, changed sandwich evidence, unavailable reads, and unknown
  states are Blocked.
- `runtimelaunchflow` owns the closed Continue, StopConverged, BindingFailed,
  and Blocked result contract. Continue and StopConverged require nil error;
  BindingFailed and Blocked require a cause. Invalid pairs and recovered
  continuation panics become Blocked with `ErrInvalidContinuation`.
- caller cancellation before `PrepareStart` maps to Cancelled; definitive Owner
  rejection maps to Rejected; other pre-claim dependency/failure paths map to
  Failed. After authentic preparation, continuation and Owner convergence use
  `context.WithoutCancel` so later cancellation cannot hide the exact result.
  Blocked converges the local Owner claim with `FailedPreparation(cause)` but
  returns the original cause, and Owner convergence errors take precedence.
- no command/identity locks cross callbacks, waits, stores, lifecycle or I/O;
  no runtimeidentity production change, DP-014 terminal publication, DP-015
  terminalization, production wiring, or legacy/public behavior change is in
  scope.

## Verification

- Verification Matrix: accepted before implementation; it covers neutral value
  and alias proof, managed parent and linked construction, lookup validation,
  every early/no-claim/final race, identity call traces and adversarial
  sandwiches, all Flow outcomes and error precedence, cancellation, expiry,
  legacy regression, static dependency direction, stress, and full gates;
- independent code proof: `PASS`, blocking findings 0, non-blocking findings 0;
- focused tests for `runtimecommandidempotency`,
  `runtimeorchestrationcontinuation`, `runtimeorchestrationbinding`, and
  `runtimelaunchflow`: PASS;
- full `go test ./... -count=1` and `go vet ./...`: PASS;
- `go mod tidy -diff` and `git diff --check`: PASS; `gofmt -d` reported only
  the repository CRLF-to-LF representation difference and no semantic
  formatting delta;
- race detector: environment-limited because `CGO_ENABLED=0` and no C compiler
  is available; the accepted focused concurrency and stress proof is covered
  by the independent code proof;
- dependency direction, unchanged public/legacy surfaces, four-method identity
  boundary, and unchanged `runtimeidentity` production semantics: PASS;
- PROCESS-002 documentation checks: PASS; DP-015 headings/fences 29/29 and
  4/4, DP-019 25/25 and 16/16, DP-020 34/34 and 12/12, MASTER_PLAN
  headings 36/36; changed-document links 163/0, repository links 920/0, and
  final diff check PASS;
- independent final Reviewer: `APPROVED`, blocking findings 0, non-blocking
  findings 0;
- Coordinator closure audit and Acceptance: PASS; implementation accepted on
  2026-08-12.

## Scope Audit

Final Coordinator-accepted classification: 27 Required, 0 Questionable,
0 Removable.

- Required production (9): `managed_start.go`, `managed_parent.go`,
  `managed_rendezvous.go`, `rendezvous.go`, and `store.go` in
  `runtimecommandidempotency`; `flow.go` and `managed_flow.go` in
  `runtimelaunchflow`; `binding.go` in `runtimeorchestrationbinding`; and the
  new `runtimeorchestrationcontinuation/continuation.go`;
- Required tests (5): `managed_start_test.go`, `managed_slice3_test.go`,
  `managed_flow_test.go`, `binding_test.go`, and
  `runtimeorchestrationcontinuation/continuation_test.go`;
- Required documentation (13): this task record, the task index, DP-015,
  DP-019, and DP-020 EN/RU, `.ai/PROJECT_CONTEXT.md`, `spec/current-state.md`,
  `spec/decisions.md`, and MASTER_PLAN EN/RU.

Deletion test: each production file owns a distinct required protocol seam;
each test file supplies direct proof for its owning seam; each documentation
file is required by PROCESS-002 applicability or EN/RU parity. No file can be
removed without losing implementation, proof, status truth, or mirror parity.

## Size Guard

- triggers accepted: the actual scope is 27 files (9 production, 5 tests, and
  13 documentation files) and exceeds 500 production lines, but adds only one
  focused package, preserves one architecture contract, and ships one
  independently correct behavior;
- decision: `ACCEPT — DO NOT SPLIT`. Reassess only if a second new package,
  runtimeidentity semantic change, DP-014 terminal publication, or production
  wiring enters scope.

## Documentation Sync

- applicable and synchronized in this pass: task record and task index;
  DP-020 implementation boundary/status EN/RU; DP-015 and DP-019 factual
  implementation boundaries EN/RU; `.ai/PROJECT_CONTEXT.md`,
  `spec/current-state.md`, `spec/decisions.md`, and MASTER_PLAN EN/RU;
- semantic boundary: Slice 3 is implemented and accepted only in isolation;
  Slice 4 is not activated, TASK-026 remains
  Blocked, and no production wiring or terminal publication is claimed;
- CHANGELOG: not applicable because TASK-037 adds no user-facing or releasable
  capability.

## Independent Review

- verdict: `APPROVED`;
- blocking findings: 0;
- non-blocking findings: 0;
- Coordinator Acceptance: `Accepted` on 2026-08-12.

## Commit Gate

- exact command `Разрешаю коммит.` received for this task: no;
- stage, commit, push, and publication remain prohibited.

## Process Health

- trigger applicable only if the accepted Slice-3 handoff proves impossible or
  the workflow cannot classify this large single behavior;
- no process change planned at intake.

## Handoff

Implementation and independent code proof are complete with zero findings.
PROCESS-002 synchronization and its checks are complete.
Independent Review is `APPROVED` 0/0 and Coordinator Acceptance is complete.
Slice 4 remains a separate, explicitly unactivated candidate.

## Publication

- publication readiness is separate from completion;
- Publisher P0–P10: not authorized.

## Next Candidate

- DP-020 Slice 4 orchestration-readiness reassessment is now eligible for a
  separate intake after TASK-037 acceptance;
- explicitly not started.

## Closure

- Final status: `Completed — Coordinator Accepted`;
- Closed by: Coordinator after Independent Reviewer `APPROVED` 0/0;
- Date: 2026-08-12.

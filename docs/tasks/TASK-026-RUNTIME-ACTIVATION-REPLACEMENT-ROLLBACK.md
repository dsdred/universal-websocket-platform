# TASK-026 — Runtime Activation, Replacement, and Rollback Implementation

## Status

`Blocked by Architecture`.

Coordinator Acceptance, implementation, commit and publication are forbidden
until the prerequisite design and implementation identified below are complete.

## Original Task Contract

### Task Mode

`Implementation`: implement the complete Approved DP-016 bounded isolated
orchestrator without production wiring, HTTP API, external persistence or
changes to existing package contracts.

### Definition of Done

The implementation would have to satisfy every DP-016 §25 acceptance proof,
including parent/phase admission, Stop-during-Starting continuation, exact
attempt/generation binding before Load, replacement/rollback authorization,
truthful cancellation and indeterminate outcomes.

### Out of Scope

- weakening or deferring any DP-016 acceptance proof;
- changing DP-016 ordering to fit current package surfaces;
- production wiring, HTTP/API, external storage, recovery/reporting or
  automatic rollback;
- commit or publication while Blocked.

## Architecture Blocking Discovery

Repository evidence proves that the original implementation scope is not
Ready:

1. DP-015 describes bounded parent/phase semantics but the implemented
   `internal/runtimecommandidempotency` surface exposes only primitive
   Start/Stop execution and no parent/phase claim API.
2. DP-011/DP-013 describe a planned private Start-claim continuation after
   Owner claim and before Load, but current Flow/Directory surfaces do not
   implement it.
3. DP-016 requires exact DP-014 attempt publication and execution-generation
   binding before Load; the current integration surfaces cannot coordinate
   those facts with the Owner-issued attempt.
4. Existing authorization actions cover Start/Stop/Observe but do not define
   the complete Workspace/Configuration/Runtime Instance/target-version
   authorization intent for replacement and rollback.

These are missing architecture/API prerequisites, not implementation details.
An adapter over current public surfaces cannot satisfy DP-016 §25.

## Coordinator Decision

- Architecture Blocking Discovery: `Confirmed`.
- Variant B (a simplified bounded adapter with limited/deferred proofs):
  `Rejected`.
- Original TASK-026 implementation scope: unchanged.
- Coordinator Acceptance: forbidden while Blocked.
- Commit/publish: forbidden.
- Required next work: design-only TASK-027 and subsequent prerequisite
  implementation before TASK-026 readiness may be reassessed.

## Rejected Resolution

The earlier proposal to map activation/replacement/rollback onto existing
primitive Start/Stop calls without new contracts is invalid. DP-015 conceptual
text is not an implemented parent/phase API, and documenting DP-016 §25 proof 9
as a limitation would violate the original Definition of Done. No production
code or tests were created from that proposal.

## TASK-038 Readiness Reassessment

Accepted DP-020 Slices 1–3 now implement the original command authorization,
parent/phase, rendezvous, managed Flow/Owner-claim continuation, and DP-014
binding seams in isolation. Completed and Coordinator-Accepted TASK-038 records
the exact verdict
`TASK-026 REMAINS BLOCKED`: the atomic expected-attempt Owner Stop prerequisite
is now implemented and independently accepted in isolation, but the later
private exact-scope composition invoker and subsequent readiness/orchestrator
boundaries remain absent.

At TASK-038 closure, the first bounded prerequisite candidate was a separate
unactivated design-only DP-010 atomic expected-attempt Stop contract, and
TASK-038 did not finalize its API or change Approved sources. TASK-039
subsequently activated and recorded that accepted design in Draft DP-010.
Completed and Coordinator-Accepted TASK-040 implements and verifies the
extension in isolation, with repeat final Reviewer `APPROVED` 0/0. TASK-026
remains Blocked by the later private exact-scope composition invoker and
subsequent readiness/orchestrator boundaries.
Post-Owner DP-014 terminal publication and DP-015 command/phase terminalization
remain core TASK-026 orchestrator work, not a separate prerequisite. External
API, persistence, recovery, reporting, production wiring, and Production
Activation remain outside the bounded reassessment. This record stays
`Blocked by Architecture`; no automatic reactivation or acceptance occurred.

## Sources of Truth

- Active ARCH-004 §19(4);
- Draft DP-011 and DP-013, including their Planned continuation;
- Approved DP-014, DP-015, DP-016 and DP-017;
- current package surfaces and tests as factual implementation evidence;
- PROCESS-001 and PROCESS-002.

## Branch and Repository State

- trusted baseline: `main@751577e839cdea3a0f35032b1339d1d9f74d28ec`;
- original branch: `feature/task-026-runtime-activation-orchestration`;
- no production/test changes, stage, commit, push or publication were made;
- Blocked governance record is carried as Required prerequisite-state evidence
  in TASK-027 because a separate TASK-026 commit is forbidden.

## Verification and Scope

- implementation Verification Matrix: not run; implementation never became
  Ready and no code/test diff exists;
- architecture inspection: FAIL/Blocked for the four missing contracts above;
- exact TASK-026-only diff before handoff: this Task Record and task index;
- classification: 2 Required / 0 Questionable / 0 Removable;
- PROCESS-002: Blocked-state synchronization continues in TASK-027 project-
  state scope.

## Handoff

TASK-027 must define one coherent prerequisite contract covering:

- the agreed DP-015 parent/phase claim API rather than new Owner lifecycle
  operations;
- the private DP-011/DP-013 Start-claim continuation;
- exact activation/replacement/rollback authorization scope;
- explicit linkage to unchanged DP-016 ordering and proofs.

TASK-026 remains Blocked after design acceptance. It may be reconsidered only
after the prerequisite contract is implemented and independently accepted.

## Closure

- Final status: `Blocked by Architecture`;
- Coordinator Acceptance: not performed;
- commit: not performed and forbidden;
- publication: not performed and forbidden.

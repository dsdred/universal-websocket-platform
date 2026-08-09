# TASK-029 — Runtime Command Continue and Pending-Stop Prerequisite

## Status

`Completed — Coordinator Accepted`.

## Task Contract

### Task Mode

`Implementation`.

### Objective

Implement one bounded, isolated remainder of Approved DP-019 inside the
DP-015 command boundary: the per-Instance Continue-versus-independent-Stop
linearization and the synchronous pending-Stop rendezvous for a parent
`StartTarget` phase. The slice must retain every live permit on its original
call stack, preserve the partial parent/phase core accepted in TASK-028, and
remain independently testable without managed Flow, DP-014 binding, or the
DP-016 orchestrator.

### Why Now

- `main@ba75e54e00c3cf1d0d87ca2a985acc9699698efd` is clean and synchronized
  with `origin/main` after PR #28;
- at deterministic selection no other development, architecture,
  documentation, or operational task was active; TASK-029 is now the sole
  active development task;
- TASK-026 remains `Blocked by Architecture` and cannot be resumed until all
  DP-019 prerequisites are implemented and independently accepted;
- TASK-028 implemented the lower-level durable parent/derived-phase storage,
  callback capability, and sequential phase core, while deliberately keeping
  `StartTarget` package-private until the Continue gate exists;
- DP-019 sections 13 and 17 require the Continue gate and pending-Stop
  rendezvous before managed Flow continuation, attempt publication, and
  generation binding can be implemented without inventing a bypass.

### Definition of Done

1. One Runtime Instance uses the existing DP-015 admission boundary to choose
   exactly one pre-Start winner between `StartTarget` Continue and an
   independent Stop.
2. If Stop wins before `StartTarget`, no StartTarget phase record or permit is
   created and the parent can converge only through a definitive stopped or
   cancelled result.
3. If `StartTarget` wins, exactly one derived phase record and one private live
   phase permit exist; at most one later independent Stop can occupy the
   tracked-Start exception.
4. A Stop occupying that exception retains its private permit on its original
   synchronous stack and does not invoke lifecycle work before rendezvous.
5. The Start-side rendezvous can signal only exact `OwnerClaimed` or
   `StartNoClaim`; only the original Stop claimant can recheck cancellation,
   invoke its exact callback once, publish its result, and signal convergence.
6. Definitive Stop convergence before Owner claim prevents Start work;
   definitive no-pending-Stop permits continuation; indeterminate, abandoned,
   stale-generation, panic, error, cancellation-after-claim, and
   `runtime.Goexit` paths fail closed and leave durable unresolved facts.
7. No callback, signal wait, or lifecycle invocation runs while a command
   admission/storage-client lock is held; different Instances progress
   independently.
8. Same-key replay/in-progress, reconstruction invalidation, redaction, and
   existing primitive and parent/phase behavior remain correct.
9. The final StartTarget surface is exported only if it cannot bypass the
   Continue gate; every exported identifier has contract-accurate GoDoc.
10. Applicable DP-019 proofs 8, 9, 14–16 and the command-boundary portion of
    proof 19 have focused executable evidence. Proofs requiring Owner claim
    views, DP-014 attempt/generation binding, managed Flow, private routing,
    or production composition remain explicitly deferred.
11. DP-019 remains `Implementation Status: Planned overall`; documentation
    reports only the isolated Continue/pending-Stop command prerequisite as
    implemented.
12. TASK-026 remains `Blocked by Architecture`; its scope, acceptance state,
    code, commit, and publication are unchanged.

### Out of Scope

- TASK-026 implementation, Coordinator Acceptance, commit, or publication;
- DP-016 activation/replacement/rollback orchestrator;
- concrete orchestration authorization policy or external Principal model;
- DP-013 private lifecycle invoker or Directory integration;
- DP-011 managed Flow constructor, `StartManaged`, or Owner claim view;
- DP-014 Launch Attempt publication, expected-revision handling, execution
  generation binding, or exact fact inspection;
- Load, Build, Launcher, Host, Control Service, HTTP/API, production wiring,
  external persistence, recovery, reporting, supervision, or automatic retry;
- changes to DP-010 lifecycle ownership or accepted DP-016 ordering;
- unrelated refactoring or new packages.

### Verification Plan

- complete the Existing Coverage Report before changing tests;
- map focused proofs to DP-019 acceptance proofs 8, 9, 14–16 and applicable
  DP-015 invariants;
- run focused package tests, adversarial cancellation/panic/`runtime.Goexit`
  proofs, stress `-count=100`, shuffled stress, full `go test ./... -count=1`,
  and `go vet ./...`;
- run race detector when technically available; otherwise record exact
  limitation and substitute stress evidence as `PASS WITH LIMITATION`;
- run `gofmt`/diff checks, `go mod tidy -diff`, module-diff validation,
  conflict-marker scan, link validation, EN/RU parity, exported-GoDoc audit,
  Status Consistency Validation, PROCESS-002, and full Scope Audit;
- require a fresh Independent Review by an agent that did not author the
  implementation; blocking findings return the task to rework.

## Selection Evidence

Selected: the DP-019 Continue/pending-Stop command-boundary prerequisite.

Rejected for this task:

- resume TASK-026 — forbidden while DP-019 prerequisites remain incomplete;
- implement all remaining DP-019 cross-package work — combines command
  rendezvous, policy, routing, Flow, aggregate binding, and composition into
  multiple independently deliverable behaviors;
- implement managed Flow first — it requires a final non-bypassable
  StartTarget Continue surface and rendezvous supplied by this slice;
- implement authorization or DP-014 binding first — those are higher-level
  per-invocation inputs and effects, not the missing command linearization;
- reduce DP-016 or use Variant B — explicitly rejected and still forbidden.

## Sources of Truth

- ADR-0002, ADR-0003, and ADR-0004;
- Frozen ARCH-002; Active ARCH-004 and ARCH-005;
- Draft/implemented-in-isolation DP-010, DP-011, and DP-013;
- Approved DP-014, DP-015, DP-016, DP-017, and DP-019;
- TASK-026, TASK-027, and TASK-028 records;
- `internal/runtimecommandidempotency` implementation and tests as factual
  baseline;
- PROCESS-001 and PROCESS-002.

## Roles and Handoffs

- Coordinator: deterministic selection, Task Contract, gates, Scope Audit,
  Closure Audit, and Acceptance;
- Architect: independently confirm implementability, ownership, lifetime,
  lock order, exact public/private surface, and proof applicability without
  changing Approved semantics;
- Developer: implement only the Architect-confirmed bounded slice;
- Tester/Verifier: Existing Coverage Report and Verification Matrix;
- Documentation Agent: PROCESS-002 applicability and status synchronization;
- Independent Reviewer: fresh adversarial final review; must not rely on the
  author's conclusions.

## Architecture Handoff

Verdict: `READY`, blocking findings: `0`.

The bounded slice is implementable without changing Approved DP-015, DP-016,
or DP-019 semantics, subject to these constraints:

- the existing per-Instance DP-015 admission boundary is the sole
  linearization point for the pre-Start Stop-versus-`StartTarget` decision;
- after `StartTarget` wins, at most one independent Stop occupies the existing
  tracked-Start exception; its non-transferable permit remains on the original
  synchronous claiming stack;
- the rendezvous carries only exact `OwnerClaimed` or `StartNoClaim` signals
  and convergence facts; it never carries or invokes a Stop permit;
- only the original Stop claimant rechecks cancellation, invokes its exact
  callback once when permitted, publishes the definitive result, and signals
  convergence;
- callback, wait, signal, and lifecycle work execute without an admission or
  storage-client lock; different Runtime Instances remain independent;
- cancellation, panic, error, `runtime.Goexit`, generation replacement,
  abandoned capability, lost signal, or indeterminate publication fails
  closed and preserves durable unresolved facts;
- any exported `StartTarget` operation must include the Continue gate and must
  not expose a phase permit or a bypassable phase-claim primitive.

Focused new proof applicability is DP-019 proofs 8 and 9, proofs 14–16 as
regression/lifetime evidence, and the command-boundary portion of proof 19.
Applicable DP-015 §24 invariants include per-Instance atomic admission,
tracked-Start Stop behavior, truthful non-terminal replay, post-claim
cancellation, one-Instance serialization, reconstruction invalidation, and
domain isolation/redaction. Existing TASK-028 proofs 1–7 remain required
regression evidence rather than a new contract.

DP-019 proofs 10–13 and 17–18, plus the Owner-view, DP-014 binding, managed
Flow, private-routing, and production-composition portions of proof 19, remain
deferred. No deferred proof may be reported as satisfied by this slice.

## Documentation Baseline

Initial verdict: `Drift Detected`.

- DOC-BL-001: live project-state sources lacked the subsequent terminal
  publication facts for TASK-027 and TASK-028, although their task records
  truthfully preserve the closure-time fact that publication had not yet
  occurred;
- DOC-BL-002: the task index and project-state sources did not yet identify
  TASK-029 as the sole active development task.

Bounded pre-implementation resolution: TASK-027 task commit
`7ac0a6b372d9e54c73d024703e6d3ee4b06e15cd` was published through PR #27 and
merged as `2c017aace7e56a4747d3cecbe8ff3f6cf53e009f`; TASK-028 task commit
`d28efa4e88e02ef528c78c3ca88b3f91945069ce` was published through PR #28 and
merged as `ba75e54e00c3cf1d0d87ca2a985acc9699698efd`. The task index,
`spec/current-state.md`, `spec/decisions.md`, and `.ai/PROJECT_CONTEXT.md` are
synchronized to those durable facts and active TASK-029. Historical task
records are not rewritten. TASK-026 remains `Blocked by Architecture`.

## Branch Decision

- branch: `feature/task-029-runtime-command-continue-rendezvous`;
- baseline: clean synchronized
  `main@ba75e54e00c3cf1d0d87ca2a985acc9699698efd`;
- this Task Record is the first content change;
- stage, commit, push, merge, publication, remote mutation, and `main`
  mutation are not authorized by `Продолжай проект.`.

## Size Guard

Initial expectation is one existing package, focused proof tests, this task
record/index, linked DP status mirrors, and applicable project-state mirrors.
The slice contains one independently deliverable concurrency behavior and no
new package or architecture contract. If implementation exceeds 500
production lines, changes more than 15 files without documentation-sync
necessity, or requires managed Flow/DP-014 integration, Coordinator must stop
and reassess rather than silently expand scope.

## Size Guard Reassessment

Decision: `DO NOT SPLIT`, conditionally cohesive but not accepted.

The current implementation is 566 production LOC, exceeding the 500-line
reassessment trigger. The diff still represents one independently deliverable
concurrency behavior: the pre-Start Continue-versus-Stop gate and the
post-Start pending-Stop rendezvous share one per-Instance admission state,
one original-stack permit lifetime, and one fail-closed convergence boundary.
Splitting storage/admission from signaling/convergence would create an
intermediate bypass or an unprovable permit-transfer boundary. This cohesion
justifies keeping the slice together for rework; it does not satisfy the Task
Contract or authorize Acceptance.

Blocking findings:

- `SG-B-001`: cancellation must linearize under the ledger lock only while the
  rendezvous signal is `None`. Cancellation-first must publish a definitive
  no-mutation result so a later `OwnerClaimed` observes no pending Stop and may
  allow Start; `OwnerClaimed`-first must invoke the exact Stop callback despite
  later caller-context cancellation.
- `SG-B-002`: only `OutcomeSucceeded` proves `StopConverged` after Owner claim.
  `OutcomeFailed` or `OutcomeRejected` must produce `Blocked` and leave the
  linked set unresolved, although the terminal Stop record may remain durable.
- `SG-B-003`: remove the raw `inspectOrExecuteStartTarget` seam and migrate its
  tests to the non-bypassable surface. `StartNoClaim` is legal only for an
  explicit, provable definitive Start end before Owner claim.
- `SG-B-004`: `ContinueOrExecuteStartTarget` must accept the exact context and
  atomically check cancellation in the same admission gate before phase claim.
  A cancellation winner creates no StartTarget phase and permits only a
  Cancelled parent outcome.

Required rework and proofs:

- cover both SG-B-001 cancellation orders and a repeated adversarial race;
- prove Failed/Rejected Stop outcomes never report convergence;
- prove no raw StartTarget phase seam remains callable;
- cover parent cancellation-versus-Continue gate ordering and no-phase
  cancellation outcome;
- retain reconstruction, `runtime.Goexit`, and different-Instance progress
  proofs while applying the bounded fixes.

Focused rerun clarification: with `GOCACHE` set to a dedicated temporary
directory outside the repository,
`go test ./internal/runtimecommandidempotency -count=1 -timeout=20s` passed;
the Go package test duration was `1.409s`, while cold command wall time was
`43.4s`. The earlier apparent hang is therefore classified as cold-cache
uncertainty, not a reproduced test deadlock. The attributable in-repository
`.tmp-go-cache-task029/` was removed only after exact path verification.

Current disposition: `Needs Revision`; Coordinator Acceptance, Commit Gate,
commit, push, and publication are not permitted. After SG-B-001–SG-B-004
rework, rerun the focused proofs and full Verification Matrix, then obtain a
fresh Independent Review.

## Existing Coverage Report

### Existing Coverage

The current `internal/runtimecommandidempotency` tests already prove the
primitive DP-015 boundary and TASK-028 sequential parent/phase core:

- same-key claim/replay/conflict, authorization-before-mutation and on replay,
  claim-before-delegation, private synchronous permits, and at-most-once
  callbacks under concurrent submission;
- primitive post-claim cancellation, error, panic, invalid/lost execution,
  `runtime.Goexit`, one tracked-Start Stop exception, unresolved barriers,
  one-Instance different-key serialization, scope isolation, redaction,
  storage-client reconstruction, and stale-Boundary invalidation;
- parent authorization/replay, immutable Replace/Rollback intent, derived
  `StopOld -> StartTarget` order, one phase callback, phase replay, parent
  terminal gate, abandoned/retained/copied capability rejection, in-flight
  callback join, cancellation, error, panic, `runtime.Goexit`, reconstruction,
  redaction, and different-Instance progress;
- callbacks run after the current admission locks are released; concurrent
  observers and commands for another Instance can progress while an existing
  callback is blocked.

This is regression evidence for applicable DP-015 section 24 invariants and
DP-019 proofs 3–7 and 14–16 only within the already implemented sequential
core. It is not evidence for the missing Continue/pending-Stop protocol.

### Coverage Gaps

The repository has no executable proof yet for:

- the single pre-Start Continue-versus-independent-Stop winner;
- Stop-first convergence without a `StartTarget` record, permit, or callback;
- Start-first creation of exactly one phase permit and admission of at most one
  later pending Stop;
- original-stack retention of the pending Stop permit, zero lifecycle work
  before rendezvous, and exactly-once work after `OwnerClaimed`;
- exact `OwnerClaimed`/`StartNoClaim` signaling, invalid or duplicate signal
  rejection, and truthful convergence;
- cancellation, callback error, panic, invalid outcome, `runtime.Goexit`, lost
  signal, abandoned capability, and stale-generation behavior while a
  rendezvous exists;
- reconstruction while a Stop is waiting, no restored permits, and no wait or
  callback under admission/storage-client locks;
- replay, redaction, barrier reopening, and primitive/parent cross-kind
  serialization after the admission rules are extended;
- a final exported non-bypassable StartTarget surface and exact exported
  GoDoc.

Therefore DP-019 proofs 8 and 9 and the new-path portions of proofs 14–16 are
open; the command-boundary documentation portion of proof 19 is also pending.

### Added Proof Tests — Planned

Focused tests must prove:

1. Stop-first creates no StartTarget fact and permits only definitive
   Stopped/Cancelled parent convergence.
2. Start-first creates one StartTarget phase/permit; concurrent Stop keys admit
   at most one pending claimant.
3. An adversarial simultaneous Continue-versus-many-Stop race has exactly one
   legal winner on every iteration.
4. The claiming Stop call remains synchronously blocked with its private permit;
   its callback is not invoked before a signal, is invoked at most once after
   exact `OwnerClaimed`, and replay receives no callback.
5. `StartNoClaim` consumes the pending permit without Stop lifecycle work;
   invalid, opposite, or repeated signals fail closed without duplicate work.
6. Pre-signal and post-claim cancellation, callback error/panic/invalid result,
   and separate Start-side and Stop-side `runtime.Goexit` paths unblock peers
   only with a definitive result or `Blocked`, preserve unresolved facts, and
   never imply Continue.
7. Boundary reconstruction before Continue and during a pending wait preserves
   records, restores no capability, rejects stale signal/publication/work, and
   cannot deadlock on a retained storage-client lock.
8. Same-key inspection remains responsive during rendezvous and a different
   Runtime Instance completes independently.

### Added Regression Tests — Planned

- retain the primitive tracked-Start one-Stop exception and all unresolved
  primitive/parent/phase barriers;
- race primitive versus parent and different parent keys through the same
  per-Instance admission point;
- retain phase order, phase/parent replay, terminal reopening, redaction,
  callback-copy/post-return/in-flight behavior, invalid-outcome handling, and
  stale-Boundary rejection;
- prove the final exported surface has no callable StartTarget bypass and that
  every new exported identifier has contract-accurate GoDoc.

### Remaining Limitations

- Owner claim views, DP-014 attempt publication/generation binding, the Load
  gate, managed Flow, private DP-013 routing, concrete authorization, DP-016
  orchestration, and production composition remain explicitly deferred.
- `MemoryStorage` proves same-process client reconstruction, not process-restart
  durability or DP-017 recovery.
- Go exposes no supported goroutine identity. Original-stack ownership must be
  proved structurally and synchronously: no permit is exposed, the claiming
  call cannot return before use/expiry, the continuation only signals, and a
  replay cannot invoke work.
- Scheduling tests prove required safety and progress under controlled
  interleavings, not fairness; fairness is not part of the contract.

### Baseline Verification Matrix

| Check | Baseline result |
| --- | --- |
| `go test ./internal/runtimecommandidempotency -count=1 -cover` | PASS; 85.9% statements |
| `go test ./internal/runtimecommandidempotency -count=100` | PASS |
| `go test ./internal/runtimecommandidempotency -count=20 -shuffle=on` | PASS; final changed code still requires shuffled `-count=100` |
| `go test ./... -count=1` | PASS |
| `go vet ./...` | PASS |
| `go test -race ./internal/runtimecommandidempotency -count=1` | unavailable: `-race requires cgo`; `CGO_ENABLED=0`, configured `CC=gcc`, executable `gcc` absent; final status requires `PASS WITH LIMITATION` plus focused adversarial stress substitutes |
| formatting/module/diff safety | normalized `gofmt` PASS; `go mod tidy -diff` empty; `go.mod`/`go.sum` diff empty; `git diff --check` PASS; conflict markers absent |
| exported GoDoc, links, EN/RU parity, status consistency, PROCESS-002, Scope Audit | pending implementation and final documentation sync |

The report is complete before any TASK-029 test creation or modification.

## Post-Rework Architecture and Verification

### Size Guard

Decision remains `DO NOT SPLIT`; current production diff is net `+586` LOC and
is conditionally cohesive, but it is not accepted. The admission winner,
immutable Stop-slot consumption, original-stack rendezvous, and fail-closed
terminal compatibility still form one independently testable concurrency
behavior. Splitting them would create an intermediate state in which the
single Stop exception or its permit lifetime could not be proved. The larger
diff therefore proceeds only through bounded rework and a new review; it does
not waive the Size Guard or authorize Acceptance.

### Architect Reassessment

Verdict: `NEEDS REVISION`; blocking findings: `3`.

- `PSG-B-001` — once one independent Stop wins the tracked-Start exception,
  that Stop slot is consumed immutably even if its caller later cancels. A
  second distinct Stop must never be admitted for the same Start execution.
- `PSG-B-002` — Stop-first parent convergence permits only exact `Stopped` or
  `Cancelled`; `StartNoClaim`, signal state, Stop record outcome, and parent
  terminal outcome must remain truthfully compatible.
- `PSG-B-003` — exported contracts must state the actual synchronous
  capability lifetime, cancellation ordering, legal outcomes, and terminal
  preconditions without implying permit transfer, retry, or generic
  convergence.

### Independent Tester

Verdict: `FAIL`; blocking findings: `4`.

- `T-B-001` — missing proof that many distinct concurrent Stop submissions
  cannot reuse a consumed slot, including after cancellation of its first
  claimant;
- `T-B-002` — pending Stop error, panic, invalid outcome, Failed, and Rejected
  paths lack complete proof that they remain Blocked/unresolved rather than
  reporting Stop convergence;
- `T-B-003` — Start-side error, panic, and `runtime.Goexit` with a pending peer,
  plus stale or late signaling, require explicit no-deadlock and fail-closed
  proofs;
- `T-B-004` — exported GoDoc and new-path replay/redaction evidence are
  incomplete, so the public contract and DP-019 proof boundary are not yet
  demonstrated.

### Verification Matrix

| Check | Post-rework result |
| --- | --- |
| `go test ./internal/runtimecommandidempotency -count=1 -cover` | PASS; 84.0% statements |
| SG focused proof set, `-count=100` | PASS |
| `go test ./internal/runtimecommandidempotency -count=100` | PASS |
| `go test ./internal/runtimecommandidempotency -count=100 -shuffle=on` | PASS |
| `go test ./... -count=1` | PASS |
| `go vet ./...` | PASS |
| `gofmt` check | PASS |
| `go mod tidy -diff` and `go.mod`/`go.sum` diff | PASS; empty |
| `git diff --check` | PASS |
| `go test -race ./internal/runtimecommandidempotency -count=1` | unavailable: `-race requires cgo`; `CGO_ENABLED=0`, configured `CC=gcc`, executable `gcc` absent; status remains `PASS WITH LIMITATION` only after blocking findings are resolved and substitute stress evidence is rerun |
| exported GoDoc audit | FAIL; `Execute`, `ExecuteParent`, `ContinueOrExecuteStartTarget`, and `PublishTerminal` do not yet fully describe their actual contracts |

### Required Second Rework

The bounded rework must:

1. make the consumed independent-Stop slot immutable after its first claim,
   including cancellation paths;
2. restrict Stop-first parent terminalization to exact `Stopped` or
   `Cancelled` and enforce truthful `StartNoClaim`/Stop/parent outcome
   compatibility;
3. correct exported GoDoc for `Execute`, `ExecuteParent`,
   `ContinueOrExecuteStartTarget`, and `PublishTerminal`;
4. add a many-distinct-Stop concurrency proof, including first-claimant
   cancellation and repeated adversarial races;
5. add pending Stop error, panic, invalid-outcome, Failed, and Rejected proofs;
6. add Start-side error, panic, and `runtime.Goexit` proofs while a pending Stop
   peer waits;
7. add stale/late-signal rejection and new-path replay/redaction proofs while
   retaining reconstruction, Goexit, and different-Instance coverage.

Status remains `In Progress — Needs Revision`. Coordinator Acceptance, Commit
Gate, commit, push, and publication are prohibited. A fresh Architect check,
complete independent Tester PASS, full Verification Matrix, documentation
sync, and Independent Review are required after the second rework.

## Second Repeat Audit / Third Rework Handoff

### Size Guard

Decision remains `DO NOT SPLIT`; the production diff is now net `+601` LOC,
conditionally cohesive, and not accepted. The remaining findings all concern
one indivisible semantic boundary: when the Start phase exists, when the one
Stop slot is consumed, and which exact signal/outcome can release or block the
same parent execution. Splitting this state machine would weaken the atomic
proof rather than reduce scope. The increased size requires another bounded
rework and fresh gates; it grants no Acceptance exception.

### Architect Second Repeat Audit

Verdict: `NEEDS REVISION`; blocking findings: `2`.

- `PS2-B-001` — pre-phase Stop cancellation and post-phase pending-Stop
  cancellation are distinct. A pre-phase cancellation winner must prevent the
  StartTarget phase and allow only a Cancelled parent; cancellation after the
  phase exists consumes the Stop slot but leaves `NoPending` so Start may
  continue.
- `PS2-B-002` — `StartNoClaim` requires an empty attempt identity and a
  truthful definitive no-attempt outcome. It cannot accept Succeeded; only
  contract-compatible Rejected or Failed is legal. After Start exists, any
  exact converged Stop permits only truthful `ParentStopped`; `Cancelled` is
  reserved for the cancellation winner.

### Independent Tester Repeat Audit

Verdict: `FAIL`; blocking findings: `3`.

- `RT-B-001` — the suite does not yet prove both pre-phase and post-phase Stop
  cancellation semantics together with exact parent terminal compatibility;
- `RT-B-002` — terminal replay with zero callback, rendezvous redaction, and a
  stale retained signal after reconstruction lack focused executable proof;
- `RT-B-003` — exported GoDoc does not yet state the corrected cancellation,
  `StartNoClaim`, converged-Stop, replay, and stale-capability contracts.

### Verification Matrix

| Check | Second repeat result |
| --- | --- |
| `go test ./internal/runtimecommandidempotency -count=1 -cover` | PASS; 85.0% statements |
| adversarial focused proof set, `-count=100` | PASS |
| `go test ./internal/runtimecommandidempotency -count=100` | PASS |
| `go test ./internal/runtimecommandidempotency -count=100 -shuffle=on` | PASS |
| `go test ./... -count=1` | PASS |
| `go vet ./...` | PASS |
| `git diff --check` | PASS |
| `go test -race ./internal/runtimecommandidempotency -count=1` | unavailable: `-race requires cgo`; `CGO_ENABLED=0`, configured `CC=gcc`, executable `gcc` absent; final status remains `PASS WITH LIMITATION` only after blockers are resolved and substitute stress is rerun |

### Required Third Rework

The bounded rework must:

1. distinguish pre-phase Stop cancellation so Continue creates no phase and
   the parent may terminalize only Cancelled, while post-phase cancellation
   retains `NoPending` and permits Start continuation;
2. require `StartNoClaim` to carry an empty attempt identity and reject
   Succeeded, accepting only truthful definitive Rejected/Failed no-attempt
   outcomes allowed by the contract;
3. permit only `ParentStopped` after any exact converged Stop once Start exists;
   `ParentCancelled` remains exclusive to the cancellation winner;
4. add terminal replay proof with zero callback and rendezvous redaction proof;
5. prove a stale retained signal returns `ErrBoundaryExpired`, exposes no live
   capability after reconstruction, and cannot publish or invoke work;
6. update exported GoDoc to match the corrected phase/cancellation,
   `StartNoClaim`, convergence, replay, and stale-capability semantics.

Status remains `In Progress — Needs Revision`. Coordinator Acceptance, Commit
Gate, commit, push, and publication remain prohibited. Third rework must be
followed by fresh Architecture, Tester, full Verification Matrix,
documentation synchronization, and Independent Review gates.

## Terminal Audit / Fourth Rework Handoff

### Size Guard

Decision remains `DO NOT SPLIT`; the production diff is now net `+620` LOC,
conditionally cohesive, and not accepted. The remaining defect is inside the
same rendezvous state machine and terminal-cause projection; extracting it
would separate the signal discriminator from the only code that can publish a
truthful parent outcome. The slice remains bounded for one more rework, without
waiving Size Guard or any acceptance gate.

### Architect Terminal Audit

Verdict: `NEEDS REVISION`; blocking findings: `1`.

- `TA-B-001` — a post-phase `StartNoClaim` that satisfies the pending Stop is
  not Stop convergence. The parent must preserve the exact Start-phase cause:
  Rejected with no attempt may map to `ParentCancelled` only when it truthfully
  represents caller cancellation; Failed with no attempt maps to
  `ParentFailed`. This path must never fabricate `ParentStopped`.
  `OwnerClaimed` followed by exact Stop convergence and the separate pre-phase
  Stop-first winner continue to map truthfully to `ParentStopped`.

### Independent Tester Terminal Audit

Revised verdict after Source-of-Truth check: `FAIL`; blocking findings: `1`.
Tester independently confirmed `TA-B-001`. The mechanical Verification Matrix
remains PASS with the race-detector environment limitation, but those results
do not satisfy the contradicted architectural outcome contract.

### Verification Matrix

| Check | Terminal-audit result |
| --- | --- |
| `go test ./internal/runtimecommandidempotency -count=1 -cover` | PASS; 85.9% statements |
| specific rendezvous/adversarial proof set, `-count=100` | PASS |
| `go test ./internal/runtimecommandidempotency -count=100` | PASS |
| `go test ./internal/runtimecommandidempotency -count=100 -shuffle=on` | PASS |
| `go test ./... -count=1` | PASS |
| `go vet ./...` | PASS |
| documentation and exported-GoDoc checks | PASS for the audited pre-TA-B-001 contract; fourth rework must update and recheck the corrected discriminator |
| formatting and `git diff --check` | PASS |
| `go test -race ./internal/runtimecommandidempotency -count=1` | unavailable: `-race requires cgo`; `CGO_ENABLED=0`, configured `CC=gcc`, executable `gcc` absent; substitute focused/package/shuffled stress passed |

### Required Fourth Rework

The bounded rework must:

1. introduce or correct the internal state discriminator so post-phase
   `StartNoClaim` satisfaction is distinct from exact Stop convergence;
2. preserve Rejected/no-attempt cancellation as `ParentCancelled` only when
   the exact cause is caller cancellation, and map Failed/no-attempt to
   `ParentFailed`;
3. retain `ParentStopped` only for the pre-phase Stop-first winner or
   `OwnerClaimed` followed by exact converged Stop;
4. add pending-Stop plus `StartNoClaim` proofs for Rejected/cancelled and for
   Failed/`ParentFailed`;
5. add negative mappings for Succeeded, Satisfied, and Stopped outcomes where
   they are incompatible with the no-attempt `StartNoClaim` path, proving no
   fabricated `ParentStopped`;
6. update exported GoDoc and state-discriminator documentation to state the
   exact cause-to-parent mapping, then rerun the focused and documentation
   checks.

Status remains `In Progress — Needs Revision`. Coordinator Acceptance, Commit
Gate, commit, push, and publication remain prohibited. Fourth rework requires
fresh Architecture confirmation and applicable verification/review gates.

## Final Architecture Audit / Fifth Rework Handoff

### Size Guard

Decision remains `DO NOT SPLIT`; the production diff is net `+652` LOC and is
conditionally accepted only after `FA-B-001` is fixed and reverified. The
remaining change is cause preservation inside the same StartNoClaim signal,
callback, and parent-publication path. Separating it would introduce a second
source of terminal truth. No current Acceptance follows from this conditional
Size Guard disposition.

### Architect Final Audit

Verdict: `NEEDS REVISION`; blocking findings: `1`.

- `FA-B-001` — the implementation can inspect `ctx.Err()` after the Start-side
  callback has already chosen a definitive non-cancellation Rejected outcome.
  Cancellation arriving in that interval can incorrectly relabel the committed
  Rejected cause as Cancelled. Parent outcome must derive only from the exact
  immutable StartNoClaim cause captured at the signal/callback boundary, never
  from a later caller-context observation.

### Independent Tester

Verdict: `PASS WITH ENVIRONMENT LIMITATION`; blocking and non-blocking test
findings: `0`. This result covers the tested implementation and does not
override the architectural cause-race finding `FA-B-001`.

### Verification Matrix

| Check | Final-audit result |
| --- | --- |
| `go test ./internal/runtimecommandidempotency -count=1 -cover` | PASS; 86.1% statements |
| focused cause/rendezvous proof set, `-count=100` | PASS |
| `go test ./internal/runtimecommandidempotency -count=100` | PASS |
| `go test ./internal/runtimecommandidempotency -count=100 -shuffle=on` | PASS |
| `go test ./... -count=1` | PASS |
| `go vet ./...` | PASS |
| `go test -race ./internal/runtimecommandidempotency -count=1` | unavailable: `-race requires cgo`; `CGO_ENABLED=0`, configured `CC=gcc`, executable `gcc` absent; substitute focused/package/shuffled stress passed |

### Required Fifth Rework

The cause-only rework must:

1. capture one immutable explicit `StartNoClaim` cause at the signal/callback
   boundary: Cancelled, Rejected, or Failed;
2. derive the parent terminal outcome only from that captured cause and never
   infer or overwrite it from `ctx.Err()` after the callback;
3. add a barrier/order proof in which the callback commits the Rejected choice,
   caller context cancels before phase publication, and the parent remains
   `ParentRejected`;
4. retain a separate explicit-cancellation proof yielding
   `ParentCancelled`;
5. update exported GoDoc and state-discriminator documentation to describe the
   immutable cause and late-cancellation rule.

Status remains `In Progress — Needs Revision`. Coordinator Acceptance, Commit
Gate, commit, push, and publication remain prohibited. Fifth rework requires
fresh Architecture confirmation and applicable verification/review gates.

## Final Architecture and Verification Handoff

### Developer Handoff

The bounded implementation changes exactly six code/test files:

- production: `parent_store.go`, `store.go`, `types.go`, and new
  `rendezvous.go`, net `+680` LOC;
- tests: `parent_store_test.go` and new `rendezvous_test.go`, net `+1387` LOC.

The production change implements the DP-015 parent/phase command-boundary
`ContinueOrExecuteStartTarget` gate and synchronous pending-Stop rendezvous in
isolation. It preserves the original Stop claimant stack and immutable Stop
slot, records an exact immutable `StartNoClaim` cause, fail-closes callback
failure and reconstruction, and keeps terminal replay callback-free. It does
not implement orchestration authorization, a private DP-013 invoker, managed
Flow/`OwnerClaimView`, DP-014 attempt/generation binding, a DP-016 orchestrator,
external durability, API, recovery, or production composition.

### Added Proof Tests

- Stop-first creates no StartTarget phase; pre-phase cancellation creates only
  a cancelled parent;
- Start-first signals the original pending-Stop stack exactly once, including
  both cancellation orders and repeated legal-winner races;
- non-successful, invalid, panic, error, and `runtime.Goexit` Stop paths remain
  blocked and never imply convergence;
- `StartNoClaim` requires an empty attempt and an exact immutable Cancelled,
  Rejected, or Failed cause, including late cancellation after a committed
  Rejected cause;
- Start callback error, panic, and `runtime.Goexit` wake a pending Stop as
  blocked;
- many distinct Stops consume one immutable slot; terminal replay executes zero
  callbacks and preserves exact facts;
- reconstruction, retained-signal expiry/redaction, and different-Instance
  progress remain fail-closed and capability-safe.

### Added Regression Tests

- primitive command replay/conflict and parent/phase sequential proofs continue
  to use only the exported, non-bypassable Continue surface;
- cancellation cannot fabricate `ParentStopped`, relabel a definitive
  Rejected cause, or create a phase after a pre-phase winner;
- Failed/Rejected Stop outcomes and satisfied StartNoClaim rendezvous are not
  treated as exact Stop convergence.

### Final Architecture Audit and Size Guard

- Architect verdict: `PASS`;
- blocking findings: `0`;
- Size Guard: `ACCEPT`; `DO NOT SPLIT`; net production delta `+680` LOC is one
  indivisible concurrency protocol whose admission, signal, callback, and
  terminal-cause state must share one source of truth.

### Independent Tester Handoff

- verdict: `PASS WITH ENVIRONMENT LIMITATION`;
- blocking findings: `0`;
- non-blocking findings: `0`.

| Check | Result |
| --- | --- |
| `go test ./internal/runtimecommandidempotency -count=1 -cover` | PASS; 85.9% statements |
| focused rendezvous/adversarial set, `-count=100` | PASS |
| `go test ./internal/runtimecommandidempotency -count=100` | PASS |
| `go test ./internal/runtimecommandidempotency -count=100 -shuffle=on` | PASS |
| `go test ./... -count=1` | PASS |
| `go vet ./...` | PASS |
| GoDoc/exported surface, `gofmt`, tidy/diff checks | PASS |
| `go test -race ./internal/runtimecommandidempotency -count=1` | unavailable: `-race requires cgo`; `CGO_ENABLED=0`, configured `CC=gcc`, and `gcc` is absent; focused, package, and shuffled `-count=100` stress are the recorded substitutes |

### Lightweight Process Health Review

Trigger: more than two review returns occurred before the final Architecture
PASS.

- Questionable/Removable: no production or test file is removable without
  splitting the single admission/rendezvous/cause protocol or deleting an
  acceptance proof; documentation changes are limited by PROCESS-002
  applicability below.
- Defects after verification: successive audits found cancellation
  linearization, convergence classification, stale-capability, exported-GoDoc,
  and immutable terminal-cause gaps. All were found before Acceptance and are
  covered by the final proof set.
- CI/post-merge: no task commit, CI publication, merge, or post-merge repair has
  occurred; this handoff does not claim those gates.
- Unavailable checks/repeated work: only the race build is unavailable. Rework
  repeatedly revisited signal timing and cause-to-terminal compatibility.
- Bounded finding `PHR-029-001`: future concurrency Architecture Handoffs should
  state an explicit signal-timing and cause-to-terminal compatibility matrix
  before implementation. Existing PROCESS-001/PROCESS-002 review gates caught
  the defects, so this single task does not justify changing engineering
  process documents.

### Final PROCESS-002 Applicability

Required and synchronized in this documentation handoff:

- this TASK-029 record and `docs/tasks/README.md`;
- DP-014, DP-015, DP-016, DP-017, and DP-019 EN/RU;
- EN/RU design indexes and MASTER_PLAN;
- `.ai/PROJECT_CONTEXT.md`, `spec/current-state.md`, and
  `spec/decisions.md`.

Inspected and Not applicable:

- root README mirrors and `CHANGELOG.md`: no user-facing, release, integrated,
  or production capability;
- DP-011 and DP-013 EN/RU: managed Flow/`OwnerClaimView` and private routing
  remain Planned and unchanged;
- DP-018 EN/RU: reporting/redaction contract and status are unchanged;
- `spec/README.md`, documentation-home and roadmap indexes: no navigation tree
  change beyond the task and design indexes;
- ADR/ARCH mirrors: no accepted architecture semantics changed;
- `go.mod` and `go.sum`: no dependency change.

DP-019 remains Approved with Implementation Status Planned overall. Only its
DP-015 parent/phase core and command-boundary Continue/pending-Stop rendezvous
are implemented in isolation. TASK-026 remains `Blocked by Architecture`.
TASK-029 proceeded from this handoff to Independent Review and Coordinator
Closure Audit as recorded below.

Final synchronization validation:

- PROCESS-002: `PASS`;
- link validation: `886` local Markdown targets checked, `0` broken;
- EN/RU parity: DP-014 `26/26` headings and `4/4` fences; DP-015 `28/28`
  and `4/4`; DP-016 `29/29` and `4/4`; DP-017 `29/29` and `2/2`; DP-019
  `24/24` and `16/16`; design filename sets identical;
- Status Consistency Validation: `PASS`; DP-019 is Approved/Planned overall,
  the isolated implemented subset is named consistently, and TASK-026 remains
  Blocked in every live mirror;
- Scope Audit: `25 Required / 0 Questionable / 0 Removable` (six bounded
  code/test files plus nineteen Required documentation files);
- conflict-marker, trailing-whitespace, and `git diff --check`: `PASS`.

## Independent Review Report

- verdict: `APPROVED`;
- blocking findings: `0`;
- non-blocking findings: `0`;
- the independent review confirmed the Task Contract, isolated DP-015/DP-019
  implementation boundary, concurrency proofs, documentation synchronization,
  and absence of scope expansion.

## Coordinator Closure Audit and Acceptance

Coordinator Closure Audit: `PASS`.

- Task Contract: complete;
- Architecture: `PASS`, blocking findings `0`;
- Size Guard: `ACCEPT`, `DO NOT SPLIT`, net production delta `+680` LOC;
- Independent Tester: `PASS WITH ENVIRONMENT LIMITATION`, blocking/non-blocking
  findings `0/0`;
- Independent Review: `APPROVED`, blocking/non-blocking findings `0/0`;
- Verification Matrix: complete; focused coverage `85.9%`, applicable stress,
  full tests, vet, GoDoc, formatting, module, and diff checks PASS; race build
  unavailable without CGO/gcc and recorded with substitute stress evidence;
- PROCESS-002 and Status Consistency Validation: `PASS`;
- link validation: `886/0`; EN/RU parity: `PASS`;
- Scope Audit: `25 Required / 0 Questionable / 0 Removable`;
- staged changes: `0`; unexpected changes: `0`;
- branch baseline: `ba75e54e00c3cf1d0d87ca2a985acc9699698efd`;
- TASK-026 remains `Blocked by Architecture` because exact orchestration
  authorization/private managed invocation, OwnerClaim-to-DP-014
  attempt/generation binding, the DP-016 orchestrator, and production
  composition remain Planned.

Coordinator Acceptance decision: `Accepted`.

No task commit, push, publication, PR, or merge has been performed for
TASK-029.

## Stop Conditions

- Architect finds that Continue and pending-Stop cannot be implemented as one
  independently correct command-boundary slice;
- the slice requires a change to Approved DP-015/DP-016/DP-019 semantics;
- permit ownership would move away from the original synchronous call stack;
- a callback or signal wait would require holding an admission/storage lock;
- implementation requires managed Flow, DP-014 binding, or production wiring;
- critical documentation drift, failed mandatory verification, or a blocking
  Independent Review finding remains unresolved.

## PROCESS-002 Applicability

Historical pre-implementation applicability (the Required-after-implementation
items below are completed by the Final PROCESS-002 Applicability section):

- TASK-029 record: Required — contract, Architecture Handoff, documentation
  baseline, verification, review, applicability, scope, and closure evidence;
- task index: Required — sole active TASK-029 and still-Blocked TASK-026;
- `spec/current-state.md`, `spec/decisions.md`, `.ai/PROJECT_CONTEXT.md`:
  Required — active-task state and subsequent TASK-027/TASK-028 publication
  facts; synchronized in this bounded pre-implementation rework;
- DP-015 and DP-019 EN/RU: Required after implementation — exact isolated
  Continue/pending-Stop capability and truthful Planned-overall boundary;
- DP-014, DP-016, and DP-017 EN/RU: Required after implementation — linked
  live status wording without changing their accepted semantics;
- design indexes and MASTER_PLAN EN/RU: Required after implementation —
  navigation and durable dependency progress while TASK-026 remains Blocked;
- DP-011 and DP-013 EN/RU: inspected, Not applicable — managed continuation
  and private routing remain absent;
- DP-018 EN/RU: Not applicable — reporting contract/status is unchanged;
- root README EN/RU and `CHANGELOG.md`: Not applicable — no user-facing,
  release, integrated, or production capability;
- `spec/README.md`, documentation-home and roadmap indexes, ADR/ARCH mirrors,
  `go.mod`, and `go.sum`: Not applicable — no tree, architecture, or dependency
  change.

At this pre-implementation checkpoint, final PROCESS-002 remained pending until
implementation, verification, and the required post-implementation mirrors.

## Next Candidate

The next recommendation is a separate bounded readiness/intake task for the
lowest remaining DP-019 prerequisite: the exact orchestration authorization,
private managed invocation, and OwnerClaim-to-DP-014 binding sequence. It is a
recommendation only and is not activated.
TASK-026 remains Blocked until every prerequisite is implemented and accepted.

# TASK-028 — Runtime Command Parent/Phase Prerequisites Implementation

## Status

`Completed — Coordinator Accepted`.

## Task Contract

### Objective

Implement one bounded, isolated prerequisite slice of Approved DP-019: the
DP-015 parent/phase durable storage, callback capability, and strict sequential
phase core for exact replacement and rollback intent. The result must be
independently testable without adding the
DP-016 orchestrator, managed Flow continuation, DP-013 private invoker, DP-014
attempt publication/binding composition, public API, or production wiring.

### Why This Task Is Ready

- `main@2c017aace7e56a4747d3cecbe8ff3f6cf53e009f` is clean and synchronized
  with `origin/main` after PR #27;
- TASK-027 is `Completed — Coordinator Accepted` and DP-019 is
  Approved/Planned;
- TASK-026 remains `Blocked by Architecture` and cannot be resumed until all
  DP-019 prerequisites are implemented and independently accepted;
- at the selection baseline parent/phase admission was the lowest-level missing
  dependency: it owns the
  durable linked records, finite phase order, callback-scoped capabilities,
  replay, and unresolved barriers on which the later Continue/pending-Stop
  slice depends;
- the existing `internal/runtimecommandidempotency` package already owns the
  primitive command ledger, per-Instance linearization point, storage-client
  generation, non-transferable execution permits, and reconstruction facts.

### In Scope

- extend `internal/runtimecommandidempotency` with exact Replace/Rollback
  parent intent and finite `StopOld` / `StartTarget` linked phase identities;
- add a synchronous `ExecuteParent` boundary that validates, authorizes on
  every submission, checks cancellation, atomically inspects or claims the
  parent, and exposes a private callback-scoped parent execution only to the
  new claimant;
- add callback-scoped phase inspection/claim foundations and parent terminal
  publication that enforce the legal finite order, derived identities, one
  lifecycle callback per newly claimed phase, and terminalization only from
  exact linked facts; the StartTarget claim helper remains package-private
  until a final Continue gate exists;
- preserve one per-Instance barrier and one linearization point against all
  independent primitive commands while a parent or phase is nonterminal;
- preserve records across `Boundary` reconstruction while invalidating every
  parent/phase capability from an earlier client generation;
- add focused and adversarial tests for all applicable acceptance proofs;
- synchronize implementation status and project-state documentation according
  to PROCESS-002 without claiming the rest of DP-019 is implemented.

### Out of Scope

- TASK-026 implementation or Coordinator Acceptance;
- DP-016 activation/replacement/rollback orchestration;
- DP-011 managed Flow constructor or `StartManaged`;
- DP-013 private lifecycle invoker or Directory integration;
- the pre-Start Continue gate and its Stop-versus-Continue winner;
- independent Stop coexistence with a nonterminal parent;
- the tracked-Start parent exception and post-Start Stop admission;
- pending Stop permit admission, holding, rendezvous, or lifecycle delegation;
- pending-Stop rendezvous after Owner claim;
- DP-014 attempt publication, execution-generation binding, or integration;
- concrete authorization policy, Principal model, HTTP/CLI/DTO/API;
- external persistence/schema/migrations, recovery, reporting, production
  composition, automatic rollback/restart, or zero-downtime behavior;
- commit, push, PR, merge, publication, or branch cleanup.

### Acceptance Criteria

1. Replace and Rollback parent scopes bind immutable exact non-zero
   target-version intent and reject structurally invalid/cross-operation
   combinations before authorization. Published-version and
   same-Configuration validation remain an upstream prerequisite.
2. Authorization executes on initial, in-progress, and terminal replay
   submissions before any command inspection or mutation.
3. Authorization denial/failure, validation failure, and pre-claim
   cancellation perform zero mutation.
4. Concurrent same-key/same-intent submissions create one parent claim;
   different intent conflicts without mutation.
5. Only a newly claimed parent receives one callback-scoped parent execution;
   returning, error, panic, `runtime.Goexit`, or client reconstruction expires
   unpublished live capabilities and leaves committed nonterminal facts
   unresolved.
6. Only legal derived phase identities and order are accepted:
   optional `StopOld` ordinal 0, then exactly one `StartTarget` ordinal 1;
   no caller-selected identity, repetition, branch, loop, or reorder.
7. A parent execution cannot directly invoke lifecycle work. Only a newly
   claimed phase callback can run one exact operation at most once.
8. Same-phase replay returns in-progress or terminal facts without a permit;
   conflicting phase intent/order fails without lifecycle mutation.
9. Parent terminal publication requires every required claimed phase to have
   an exact definitive terminal outcome, or a definitive zero-phase parent
   outcome before a phase is claimed; aggregate observation cannot fabricate
   a phase result.
10. An unresolved parent/phase blocks every unrelated primitive or parent
    command. This slice does not admit the DP-019 tracked-Start Stop exception.
11. Reconstruction preserves parent/phase records and terminal replay but
    restores no live capability; stale clients cannot claim or publish.
12. Different Runtime Instances progress independently and no callback runs
    under command-admission/storage-client locks.
13. Stored records contain no authorization capability, caller context, raw
    callback error, lifecycle token, Host, Snapshot, or mutable preparation.
14. Existing primitive DP-015 Start/Stop behavior and tests remain unchanged
    and passing.
15. DP-019 remains Implementation Status Planned overall; documentation states
    only that its parent/phase storage, callback capability, and sequential
    phase core are implemented in isolation, while Continue and pending-Stop
    semantics remain Planned.

### Verification Plan

- focused tests for parent validation, authorization, claim/replay/conflict,
  legal/illegal phase order, phase at-most-once, parent terminalization,
  abandoned capability, `runtime.Goexit`, stale-client reconstruction,
  unresolved independent-command barriers, primitive regression, isolation,
  and redaction;
- stress and shuffled runs for concurrent same-parent, phase, barrier, and
  different-Instance scenarios;
- `go test ./... -count=1`, `go vet ./...`, `gofmt -d`, race when available,
  module/diff/conflict checks;
- PROCESS-002 applicability, EN/RU parity, status consistency, links, exact
  Scope Audit, and fresh Independent Review.

## Selection Evidence

Selected: isolated DP-019/DP-015 parent-and-phase admission boundary.

Rejected for this task:

- resume TASK-026 — forbidden while prerequisites remain unimplemented;
- implement the full DP-019 cross-package continuation in one change — would
  combine command storage, management routing, Flow, aggregate binding, and
  rendezvous into a size-guard-heavy slice before its lowest-level admission
  primitive exists;
- implement only managed Flow continuation first — it requires a live
  primitive/phase execution binding and Continue gate not yet implemented;
- add replacement/rollback methods to Lifecycle Owner — rejected by DP-019;
- reduced Variant B — explicitly rejected by Coordinator and TASK-026.

## Sources of Truth

- ADR-0002, ADR-0003, ADR-0004;
- Frozen ARCH-002; Active ARCH-004 and ARCH-005;
- DP-010 lifecycle ownership;
- DP-011 and DP-013 existing isolated integration/routing contracts;
- Approved DP-014, DP-015, DP-016, DP-017, and DP-019;
- `internal/runtimecommandidempotency` implementation/tests as factual baseline;
- PROCESS-001 and PROCESS-002.

## Roles and Handoffs

- Coordinator: selection, Task Contract, gates, Scope/Closure Audit;
- Architect: confirm the bounded mapping from DP-019 to the existing DP-015
  ledger without changing accepted semantics;
- Developer: implementation and focused tests after architecture handoff;
- Tester/Verifier: Existing Coverage Report and Verification Matrix;
- Documentation Agent: PROCESS-002 and status synchronization;
- independent Reviewer: fresh final review after Scope Audit;
- Publisher: Not applicable; no publication authorization.

## Existing Coverage Report

### Existing Coverage

- Existing primitive coverage proves exact validation/authorization ordering,
  claim-before-delegation, replay/conflict, one tracked Start Stop exception,
  unresolved barriers, reconstruction, `runtime.Goexit` cleanup, instance
  isolation, and stored-fact redaction.
- No parent record, linked phase identity, phase state machine, parent/phase
  execution capability, Continue gate, or parent terminalization test exists.
  Continue/pending-Stop behavior is explicitly deferred rather than treated as
  missing coverage for this bounded slice.
- Existing primitive tests are mandatory regression evidence and must not be
  weakened or reclassified.

### Added Proof Tests

- structural parent intent validation, authorization-before-inspection on
  initial/in-progress/terminal submissions, pre-claim cancellation, immutable
  intent conflict, and zero-mutation failures;
- one parent and one phase delegation under concurrency, exact derived
  ordinals/order, replay, parent terminal gating, per-Instance barriers, and
  different-Instance independence;
- post-claim cancellation, indeterminate parent/phase outcomes, reconstruction,
  generation expiry, parent/phase `runtime.Goexit`, and durable redacted facts;
- zero-phase parent terminal replay and phase-bearing exact-fact publication.

### Added Regression Tests

- callback return/error/panic cannot leave a live reusable parent capability;
- retained pointers and value copies reject phase claim and parent publication
  after callback return;
- callback return atomically closes new authority and waits for a phase that
  already began, so lifecycle work cannot outlive `ExecuteParent`;
- stale storage clients and post-claim cancellation cannot duplicate parent or
  phase work.

### Remaining Limitations

- race detector is unavailable in the current environment (`CGO_ENABLED=0` and
  no `gcc`); repeated focused and shuffled stress are the recorded substitute;
- the Continue gate, independent Stop coexistence, tracked-Start parent
  exception, pending-Stop rendezvous, private managed Flow continuation, DP-014
  binding, external durability, recovery, API, and production wiring remain
  deliberately unimplemented and unproved in this task.

## Size Guard

Expected production scope is one existing package and one cohesive behavior.
If production changes exceed 500 lines, tests exceed 1000 lines, or total scope
exceeds 15 files, Coordinator must explicitly reassess whether the slice can
remain independently correct if split. No split may leave a reusable live
permit, unproved barrier, or incomplete phase state machine.

Reassessment after rework: production changes are 659 added physical lines
including comments (592 in the two new parent files and 67 in the two existing
files), tests are 574 lines, and exact repository scope is 24 files.
The triggers are exceeded, but a split is not independently correct. Parent
types without storage are inert; storage without callback lifetime and barrier
integration is unsafe; phase publication without parent terminal gating leaves
an incomplete state machine. The remaining 19 files are the proof file and
mandatory EN/RU/navigation/roadmap/project-state mirrors. The cohesive slice is
therefore retained with explicit Scope Audit and independent review.

## Architecture Analysis

Architect handoff: **READY only as a partial DP-019 parent/phase core**.

The original full-boundary wording was rejected. Approved DP-015/DP-019 lets
one Stop race after `StartTarget`, but before Owner claim that Stop must retain
its permit on the original stack without invoking lifecycle work until a later
rendezvous. The current primitive boundary invokes a newly claimed Stop
synchronously, while this task excludes that rendezvous. A one-call Continue
helper also cannot expose the required pre-Start Stop winner without a durable
`continue-open` substate. Claiming full Continue/pending-Stop semantics here
would therefore be false or would expand the task into the next prerequisite.

Accepted bounded mapping:

- add exact Replace/Rollback intent and a separate parent admission path;
- keep primitive `Boundary.Execute` restricted to Start/Stop;
- store parent and derived phase records in the same per-Instance ledger, with
  fixed `StopOld` ordinal 0 and `StartTarget` ordinal 1;
- expose no caller-selected phase identity and no reusable phase capability;
- allow linked phase work only through callback-scoped parent execution;
- keep the StartTarget claim foundation package-private until a future
  `ContinueOrExecuteStartTarget` gate can be implemented without a bypass;
- allow `StartTarget` only when `StopOld` is absent or terminal, reject
  `StopOld` after `StartTarget`, and never fabricate phase facts at parent
  publication;
- allow parent terminal publication only before any phase for a definitive
  zero-phase outcome, or after every existing linked phase is terminal;
- make every independent primitive or parent command observe the unresolved
  parent/phase barrier; do not consume the tracked-Start Stop exception here;
- bind live parent/phase capabilities to storage-client generation and expected
  record revision; return, error, panic, `runtime.Goexit`, or reconstruction
  removes liveness but preserves durable unresolved claims;
- use lock order `clientMu.R -> storage.mu (ledger lookup only) -> ledger.mu`,
  never the reverse, and invoke authorization/lifecycle callbacks without
  locks;
- store a separate bounded parent terminal outcome and detached views without
  authorization, context, raw errors, lifecycle tokens, Host, or Snapshot;
- validate only structural identity, non-zero version, and operation/intent
  agreement locally; Published/same-Configuration validation is upstream.

Applicable DP-019 proofs for this slice are 1-7, 14-16, and the truthful
status/parity portion of 19. Proofs 8-13 and 17-18 remain deferred where they
depend on Continue, pending Stop, private managed-Flow continuation, attempt
binding, or routing. TASK-026 remains Blocked by Architecture.

## Implementation

Implemented in `internal/runtimecommandidempotency` without production wiring:

- `OperationReplace` / `OperationRollback` and exact non-zero target-version
  intents extend structural Scope/Intent validation; primitive `Execute`
  remains Start/Stop-only;
- `ExecuteParent` performs validation, fresh authorization, cancellation,
  same-key decision, barrier evaluation, claim, and synchronous callback-
  scoped delegation;
- the per-Instance ledger owns separate durable parent and derived phase maps,
  generation/revision-bound live capabilities, fixed `StopOld` ordinal 0 and
  `StartTarget` ordinal 1, detached replay views, and redacted bounded parent
  outcomes;
- `InspectOrExecuteStopOld` is the only exported phase executor in this slice;
  the StartTarget foundation is package-private until the required Continue
  gate can be added without an internal bypass;
- phase publication is at-most-once and parent terminal publication accepts no
  caller-supplied phase facts; a phase-bearing parent requires terminal
  StartTarget and every existing phase;
- primitive and new-parent admission both observe unresolved parent/phase
  barriers; no tracked-Start parent exception is introduced;
- return, error, panic, `runtime.Goexit`, and storage-client reconstruction
  remove live capability without deleting durable unresolved claims;
- callback return atomically closes a shared copy-safe authority state and
  joins every phase already in flight; retained pointers or value copies cannot
  begin work after closure;
- callbacks and authorization run without ledger/client locks; lock order
  remains `clientMu.R -> storage.mu for lookup -> ledger.mu`.

No DP-016 orchestrator, DP-013/DP-011 continuation, DP-014 binding, concrete
policy, external persistence, public API, module change, or production
composition was added.

## Verification

Verification Matrix before independent review: **PASS WITH ENVIRONMENT
LIMITATION**.

- focused parent/phase suite: PASS;
- focused parent/phase stress `-count=100`: PASS;
- complete package shuffled stress `-shuffle=on -count=100`: PASS;
- package coverage: 85.9% statements;
- full `go test ./... -count=1`: PASS;
- `go vet ./...`: PASS;
- changed-file `gofmt -d`: PASS, zero diff;
- `go mod tidy -diff` and `go.mod` / `go.sum` diff: PASS, zero change;
- race detector: unavailable in the environment — default `CGO_ENABLED=0`
  rejects `-race`; explicit CGO retry cannot find `gcc`. Focused and shuffled
  stress are the recorded substitute, not a claim that race executed;
- repository-source relative links: 885 checked, 0 broken (879 inline local
  targets plus 6 reference-style uses);
- EN/RU headings/fences: DP-014 28/28 and 4/4; DP-015 29/29 and 4/4;
  DP-016 30/30 and 4/4; DP-017 30/30 and 2/2; DP-019 25/25 and 16/16;
  indexes 1/1 and 0/0; MASTER_PLAN 36/36 and 0/0;
- status consistency: 24/24 assertions PASS across 22 live sources; stale
  contradiction scan 0;
- `git diff --check`, conflict-marker scan, module check: PASS;
- exact worktree scope: 24 expected / 24 actual; staged changes 0.

## PROCESS-002 Applicability

Status: `Synchronized`; validation is included in the Verification Matrix.

- TASK-028 record: Required — contract, architecture blocker, implementation,
  proofs, scope, review, and closure;
- task index: Required — active TASK-028 and still-Blocked TASK-026;
- DP-015 and DP-019 EN/RU: Required — implemented partial core and truthful
  Planned-overall boundary;
- DP-014, DP-016, and DP-017 EN/RU: Required — live linked status statements
  distinguish the partial core from remaining Continue/pending-Stop work;
- design indexes EN/RU: Required — live navigation/status mirror;
- MASTER_PLAN EN/RU: Required — durable dependency progress while TASK-026
  remains Blocked;
- `spec/current-state.md`, `spec/decisions.md`, `.ai/PROJECT_CONTEXT.md`:
  Required — factual capability, task, status, and continuation state;
- root README EN/RU: Not applicable — no user-facing or production capability;
- `CHANGELOG.md`: Not applicable — no release/user-facing change;
- `spec/README.md`: Not applicable — specification tree unchanged;
- ARCH-004/ARCH-005 and ADR mirrors: Not applicable — accepted ownership and
  architecture semantics are unchanged;
- DP-011 and DP-013 EN/RU: inspected; Not applicable — their still-Planned
  managed continuation/current-package statements remain true;
- DP-018 EN/RU: Not applicable — reporting contract/status unchanged;
- `go.mod` / `go.sum`: Not applicable — no dependency change.

## Scope Audit

Final exact result before independent review: **24 Required / 0
Questionable / 0 Removable**.

- 5 code/proof files: primitive ledger/type integration plus the cohesive new
  parent types, storage/capability implementation, and focused proofs;
- 10 linked DP mirrors: DP-014/015/016/017/019 EN/RU;
- 4 navigation/roadmap mirrors: design indexes and MASTER_PLAN EN/RU;
- TASK-028 and task index;
- 3 project-state mirrors: `.ai/PROJECT_CONTEXT.md`, `spec/current-state.md`,
  and `spec/decisions.md`.

No existing unrelated package, test, module, API, generated file, temporary
artifact, TASK-026 implementation, or future Continue/rendezvous code changed.

## Independent Review

Initial Independent Review: `Needs Revision`, 4 blocking findings, 0
non-blocking findings.

- IR-B-001: parent capability could race callback return or let an in-flight
  phase outlive the synchronous boundary;
- IR-B-002: missing explicit in-progress authorization, post-claim
  cancellation, redaction, and capability-escape proofs plus incomplete
  PROCESS-001 coverage-report fields;
- IR-B-003: one duplicate `spec/current-state.md` task-phase mirror was stale;
- IR-B-004: exported Intent/parent outcome/view GoDoc was stale or incomplete.

Rework resolves all four findings within the existing scope: shared atomic
callback closure with in-flight join, pointer/copy escape regressions, missing
proof tests and coverage report, project-state synchronization, and complete
exported contract documentation. Full focused stress/shuffle/regression/vet,
coverage, formatting/module/diff, documentation, and status verification pass;
Repeat Independent Review is required before Closure Audit or Acceptance.

Repeat Independent Review: `Approved`; blocking findings 0, non-blocking
findings 0. IR-B-001–IR-B-004 are independently confirmed resolved. Applicable
DP-019 proofs 1–7, 14–16 and the status/parity portion of 19 pass; proofs 8–13
and 17–18 remain truthfully deferred with no forbidden implementation.

## Coordinator Closure Audit

Result: **PASS**.

- Task Contract: all 15 acceptance criteria satisfied within the Architect-
  approved partial-core boundary;
- Architecture: original full Continue/Stop claim rejected; no semantic
  weakening or out-of-scope substitute introduced;
- Verification Matrix: PASS WITH ENVIRONMENT LIMITATION; race unavailable only
  because CGO is disabled and `gcc` is absent;
- PROCESS-002 and Status Consistency Validation: PASS;
- Independent Review: repeat `Approved`, blocking 0, non-blocking 0;
- Scope Audit: 24 Required / 0 Questionable / 0 Removable; staged 0;
- repository baseline: task branch at clean synchronized
  `main == origin/main == 2c017aace7e56a4747d3cecbe8ff3f6cf53e009f`
  plus exactly the 24 planned unstaged task files;
- TASK-026 remains `Blocked by Architecture`; Continue/pending-Stop, managed
  continuation, authorization policy, binding and production work remain
  Planned.

Coordinator Acceptance: **Accepted**. Commit Gate, commit, push, PR, merge,
publication, and branch cleanup were not authorized and were not performed.

## Next Candidate

If TASK-028 is accepted, the next candidate is the remaining DP-019 Continue,
pending-Stop rendezvous, managed Flow/private-invoker, and attempt-binding
continuation prerequisite. TASK-026 stays Blocked and is not activated
automatically.

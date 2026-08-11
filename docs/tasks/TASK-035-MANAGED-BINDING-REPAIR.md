# TASK-035 — Managed Binding Repair

## Status

`Completed — Coordinator Accepted`.

## Task Contract

### Task Mode

`Implementation`. This task implements only DP-020 Slice 2R: the complete
dependency-safe managed Start binding and the sole primitive
`Boundary.ExecuteManagedStart` claim-to-binding adapter.

### Why Now

- the trusted baseline `main@9ec1979` is clean and equal to `origin/main`;
- no active task exists;
- TASK-034 is the latest accepted architecture task and names Slice 2R as the
  single next unactivated candidate;
- Slice 3 and TASK-026 remain blocked until this repair is independently
  implemented and accepted.

### Definition of Done

1. New dependency-leaf `internal/runtimeorchestrationbinding` owns immutable,
   validated authorization, neutral revision/generation, linked identity, and
   opaque rendezvous values without importing command, identity, or Flow
   packages.
2. `runtimelaunchflow` consumes the complete binding and no longer imports
   `runtimecommandidempotency` or `runtimeidentity`; its existing synchronous
   managed lifecycle behavior remains compatible.
3. `runtimecommandidempotency.Boundary.ExecuteManagedStart` is the only
   primitive claim-to-binding adapter and implements the TASK-034/DP-020
   authorization, cancellation, claim, replay, permit, rendezvous, and
   indeterminate-failure rules.
4. Existing `Boundary.Execute`, DP-013 public behavior, parent/phase behavior,
   and unmanaged Flow behavior remain compatible.
5. Proof and regression tests cover validation, exact mapping, new claim,
   replay/in-progress, legacy-record non-adoption, denial/cancellation,
   callback failure/panic/`runtime.Goexit`, generation loss, binding expiry,
   and concurrency-sensitive invariants.
6. Documentation truthfully marks Slice 2R implemented in isolation only after
   verification; Slice 3 and TASK-026 remain blocked and unactivated.
7. Verification Matrix, PROCESS-002, Scope Audit, and independent review pass
   with no unresolved blocking findings.

### Out of Scope

- DP-020 Slice 3 attempt publication or DP-014 generation binding;
- continuation outcome implementation or orchestration sequencing;
- parent/phase adapter changes beyond compatibility proofs;
- concrete authorization policy, management/API wiring, persistence schema,
  recovery, reporting, or Production Activation;
- changes to Approved DP-019, DP-014, DP-015, or DP-016 semantics;
- commit, push, PR, merge, publication, or branch cleanup.

### Verification Plan

- focused tests for the new binding package, managed Flow, and primitive
  adapter, including stress and shuffled execution where applicable;
- full `go test ./...`, `go test -race ./...`, `go vet ./...`, formatter,
  module-tidy diff, and `git diff --check`;
- public/exported-surface and package-dependency audit;
- EN/RU parity, link, status, and project-state consistency checks;
- independent technical review after final Scope Audit.

## Selection Evidence

Selected TASK-035 because TASK-034 and both MASTER_PLAN mirrors identify Slice
2R as the earliest unmet dependency in the current Beta milestone. The task is
bounded to one independently testable behavior and all required architecture
decisions exist in Approved DP-019 plus reconciled Draft DP-020.

Rejected alternatives:

- Slice 3 — explicitly blocked by Slice 2R;
- resume TASK-026 — remains `Blocked by Architecture`;
- combine Slice 2R with Slice 3 — violates the bounded-slice and Size Guard
  intent;
- concrete authorization policy or production wiring — later dependencies and
  outside the accepted repair contract.

## Scope

- new neutral binding package and its tests;
- managed Flow binding migration and compatibility tests;
- primitive command-boundary managed Start adapter and proofs;
- task navigation and only the documentation/project-state files required by
  factual Slice 2R completion.

## Sources of Truth

- PROCESS-001 and PROCESS-002;
- Approved/Planned DP-019 §§7–17;
- Draft/Planned DP-020 §§6–12 as reconciled by TASK-034;
- TASK-034 architecture handoff and closure;
- existing `runtimecommandidempotency`, `runtimelaunchflow`,
  `runtimeidentity`, and `runtimeconfigload` implementation and tests;
- current project-state sources and MASTER_PLAN EN/RU.

## Roles

- Coordinator: intake, gates, scope audit, acceptance;
- Architect: confirm exact package/API mapping before implementation;
- Developer: implement the accepted Slice 2R contract and tests;
- Tester: coverage report and risk-oriented verification;
- Documentation Agent: PROCESS-002 synchronization after executable proof;
- Reviewer: independent final review;
- Publisher: not applicable without later explicit commit and publication
  commands.

## Branch

- trusted baseline: clean `main@9ec1979`, equal to `origin/main`;
- task branch: `feature/task-035-managed-binding-repair`;
- branch action: created safely from the trusted baseline;
- forbidden: stage, commit, push, merge, branch deletion, fetch, pull, or
  mutation of `main` without the corresponding explicit gate.

## Constraints

- Approved DP-019 semantics are preserved;
- command permits and mutable rendezvous state remain command-boundary owned;
- no permit, channel, callback, pointer, mutable aggregate state, or
  cancellation authority crosses in the immutable binding;
- callbacks run synchronously outside locks and cannot transfer execution
  authority;
- package dependencies remain acyclic and explicit;
- planned behavior is not described as implemented before proof.

## Stop Conditions

- implementation requires changing Approved DP-019 semantics;
- the adapter cannot preserve legacy `Execute` byte-for-behavior compatibility;
- binding authority cannot be expired on every callback exit path;
- required race or failure semantics remain indeterminate without a new
  architecture decision;
- unexpected worktree changes or a blocking independent-review finding cannot
  be resolved inside this scope.

## Existing Coverage Report

- Existing Coverage: TASK-031 tests prove exact authorization tuple validation
  except for the missing `OperationalDomain`; TASK-032 tests prove the partial
  revision/generation/rendezvous binding, call-stack lifetime, pre-Owner
  validation, failure convergence, and unmanaged compatibility; DP-015 tests
  prove primitive claim/replay, per-instance barriers, synchronous permits,
  panic/`runtime.Goexit` cleanup, generation invalidation, and pending-Stop
  rendezvous behavior.
- Coverage Gap: no current proof covers the complete six-field authorization
  binding, neutral package direction, all-or-none linked identity, unique
  opaque command-owned rendezvous identity, or atomic primitive
  claim-to-binding adapter behavior and non-adoption of legacy records.
- Added Proof Tests: complete leaf constructor/cross-field tests, six-field
  command-to-binding mapping, two distinct new claims receiving unique non-zero
  rendezvous identities, live exact rendezvous lookup, concurrent same-key
  single callback, replay/legacy non-adoption, validation/denial/cancellation,
  error/panic/invalid outcome/generation loss/`runtime.Goexit`, and Flow
  request/instance validation proofs.
- Added Regression Tests: real successful managed lifecycle completion through
  the authentic first preparation, unchanged public command signatures/import
  direction, unmanaged Flow, DP-013, parent/phase, and legacy primitive suites.
- Remaining Limitations: race detector availability depends on the local Go
  toolchain and C compiler; any unavailability will be reported exactly as
  `PASS WITH LIMITATION` with substitute stress evidence.

## Architecture Confirmation

Architect verdict: `PASS`, blocking findings `0`.

- new dependency-leaf `internal/runtimeorchestrationbinding` owns the exact
  action/request, neutral aggregate revision and execution generation,
  all-or-none linked identity, opaque comparable rendezvous identity, and
  immutable `StartExecutionBinding`; it imports only `runtimeconfigload`;
- existing command-package orchestration names remain source-compatible through
  aliases/wrappers, while the authoritative values move to the leaf package;
- `Boundary.ExecuteManagedStart` accepts only primitive exact Start, derives
  the six-field `ActivateExactTarget` request, authorizes and checks
  cancellation before shared admission, and allocates record, permit,
  generation-bound rendezvous lookup, and complete binding atomically only for
  a new claim;
- replay/in-progress and legacy-created records are observed only and receive
  no callback, binding, rendezvous, permit, or adopted authority;
- callback error, panic, `runtime.Goexit`, invalid outcome, or generation loss
  expires lookup/permit, blocks the rendezvous, preserves Claimed truth, and
  uses the existing indeterminate error contract;
- managed Flow consumes the leaf binding, verifies primitive Activate
  authorization against StartRequest before mutation, verifies the exact
  Runtime Instance through the authentic post-claim OwnerClaimView, retains
  existing failure convergence, and removes its `runtimeidentity` dependency;
- current `StartManaged` performs a second `PrepareStart` indirectly through
  `Flow.Start`; the repair must factor a private prepared-start helper used by
  unmanaged and managed paths so the managed path continues the authentic
  first preparation and real success is executable;
- the old constant `runtimeconfigload.StartRendezvous` is removed; no
  command-owned signaling state crosses the leaf boundary;
- concrete DP-013 composition-private managed invoker is factually absent.
  Inventing production wiring is outside this task; this slice proves the sole
  command-adapter callback path plus Flow request/binding validation and keeps
  composition wiring Planned.

Non-blocking representation gap: DP-020 intentionally does not prescribe Go
identifier spellings for neutral revision/generation/linked identity values.
The implementation may choose the smallest constructor-only immutable API
that enforces the recorded invariants without changing semantics.

## Size Guard

- final files: 21 — 7 production, 5 test, and 9 documentation files;
- production additions exceed the 500-line review trigger when the complete
  constructor-only leaf value model, adapter, and migration are counted;
- new packages: exactly one dependency-leaf value package;
- architecture contracts: no new contract; implementation of one reconciled
  Draft DP-020 slice;
- independently shipped behaviors: one managed-binding repair;
- reassessment decision: `ACCEPT`, do not split. The threshold is crossed by
  one indivisible behavior: the leaf binding has no executable authority
  without the atomic adapter, and the adapter cannot compile or be proven
  without the binding and Flow migration. There remains exactly one new
  package, one architecture contract implementation, and one independently
  delivered behavior; tests and required mirror/state synchronization account
  for most of the file count.

## Implementation Handoff

Developer result: Slice 2R implemented within the confirmed architecture.

- `runtimeorchestrationbinding` now owns the immutable six-field authorization,
  lossless neutral revision/generation, constructor-derived all-or-none linked
  identity, opaque comparable rendezvous identity, and primitive/linked
  `StartExecutionBinding` variants;
- command authorization aliases preserve the accepted internal surface while
  restoring exact `OperationalDomain` mapping;
- `Boundary.ExecuteManagedStart` authorizes before inspection on every call,
  atomically creates binding/permit/private lookup only for a new primitive
  Start claim, invokes once outside locks, and never adopts legacy execution;
- defer-based cleanup expires permit and rendezvous lookup on every callback
  exit, including panic and `runtime.Goexit`; generation loss retains the
  existing `ErrBoundaryExpired` distinction;
- managed Flow validates binding against StartRequest before Owner mutation,
  validates Runtime Instance against the authentic OwnerClaimView, and uses a
  shared prepared-start helper to avoid the previously hidden second
  `PrepareStart` conflict;
- the old constant `runtimeconfigload.StartRendezvous` was removed;
- concrete DP-013 private invoker, Slice 3, identity-store binding, policy, and
  production wiring remain absent and Planned.

## Verification

Tester result before independent review: `PASS WITH ENVIRONMENT LIMITATION`.

- baseline focused suites before implementation: PASS; 20× stress PASS;
- focused Slice 2R packages after implementation: PASS;
- focused dependency packages with `-count=10 -shuffle=on`: PASS;
- full `go test ./...`: PASS;
- `go vet ./...`: PASS;
- `go mod tidy`: PASS with no `go.mod` or `go.sum` diff;
- dependency audit: `runtimeorchestrationbinding` imports only
  `runtimeconfigload`; `runtimelaunchflow` imports neither
  `runtimecommandidempotency` nor `runtimeidentity`;
- race detector: unavailable because CGO compilation cannot find `gcc`;
  exact failure: `cgo: C compiler "gcc" not found`; substitute shuffled stress
  and full tests pass;
- `gofmt` for every changed Go file, conflict-marker scan, and
  `git diff --check`: PASS; only configured LF-to-CRLF working-copy warnings
  are emitted. A broader package formatter scan lists six untouched pre-existing
  DP-015 files; they are outside this task and no changed file is listed;
- module files, generated artifacts, secrets, and external dependencies:
  unchanged/absent.

## Documentation Synchronization

PROCESS-002 final status: `Synchronized`.

- DP-020 EN/RU retain Draft/Planned overall status and now distinguish Slice
  2R implemented and independently accepted in isolation from absent private-
  invoker/production wiring;
- `.ai/PROJECT_CONTEXT.md`, `spec/current-state.md`, `spec/decisions.md`, and
  MASTER_PLAN EN/RU record TASK-035 acceptance and Slice 3 as the next
  unactivated candidate;
- TASK-035 and task index are updated;
- Approved DP-019/DP-014–DP-016: inspected, not edited; semantics/statuses did
  not change;
- design indexes, root README EN/RU, and `CHANGELOG.md`: Not applicable because
  no design status, user-facing/release surface, or index membership changed;
- historical TASK-031/TASK-032 acceptance evidence remains unchanged.

## Scope Audit

- production — all `Required`:
  `internal/runtimeorchestrationbinding/binding.go`,
  `internal/runtimecommandidempotency/orchestration_authorization.go`,
  `internal/runtimecommandidempotency/managed_start.go`,
  `internal/runtimecommandidempotency/store.go`,
  `internal/runtimeconfigload/contract.go`,
  `internal/runtimelaunchflow/flow.go`, and
  `internal/runtimelaunchflow/managed_flow.go` for the leaf values, sole
  adapter, exact cleanup/lookup ownership, old-handle removal, and authentic
  prepared path;
- tests — all `Required`:
  `internal/runtimeorchestrationbinding/binding_test.go`,
  `internal/runtimecommandidempotency/orchestration_authorization_test.go`,
  `internal/runtimecommandidempotency/managed_start_test.go`,
  `internal/runtimelaunchflow/flow_test.go`, and
  `internal/runtimelaunchflow/managed_flow_test.go` for proof/regression and
  import compatibility;
- documentation — all `Required`: this task record, `docs/tasks/README.md`,
  DP-020 EN/RU, `.ai/PROJECT_CONTEXT.md`, `spec/current-state.md`,
  `spec/decisions.md`, and MASTER_PLAN EN/RU for task evidence, mirrors,
  roadmap, and current project state;
- totals: `21 Required / 0 Questionable / 0 Removable`;
- deletion test: removing any production file breaks the accepted binding or
  adapter behavior; removing any test loses a Definition-of-Done proof;
  removing any documentation file creates task, parity, dependency, or live
  status drift.

## Independent Review

Initial final verdict: `Approved with Findings`, blocking `0`, non-blocking `1`.

- N-001: adapter implementation was collision-safe, but the proof set did not
  directly compare rendezvous identities from two distinct new claims.
- rework: added `TestExecuteManagedStartAllocatesUniqueRendezvousPerNewClaim`;
  focused `ExecuteManagedStart` suite with `-count=20 -shuffle=on`: PASS.
- focused repeat verdict: `Approved`, blocking `0`, non-blocking `0`; N-001
  resolved; prior DoD, Size Guard, Scope Audit, and out-of-scope conclusions
  remain PASS.

## Commit Gate

- exact command `Разрешаю коммит.` received: no;
- stage, commit, push, PR, merge, publication, and branch cleanup are not
  authorized.

## Closure

- Final status: `Completed — Coordinator Accepted`.
- Definition of Done: PASS 7/7.
- Verification: `PASS WITH ENVIRONMENT LIMITATION`; race unavailable because
  `gcc` is absent, with focused shuffled stress, full tests, and vet PASS.
- Independent Review: `Approved`, blocking `0`, non-blocking `0` after N-001
  proof rework.
- PROCESS-002, EN/RU parity, links, status consistency, and Scope Audit
  `21/0/0`: PASS.
- TASK-026 remains `Blocked by Architecture`; concrete private invoker,
  Slice 3 DP-014 binding, policy, orchestration, persistence, and production
  wiring remain absent.
- Closed by: Coordinator.
- Date: 2026-08-12.

## Next Candidate

DP-020 Slice 3 is the next recommendation after TASK-035 independent
acceptance. It is not activated by this closure.

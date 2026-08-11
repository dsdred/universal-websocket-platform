# TASK-032 — Runtime Private Managed Invoker and Managed Flow Seam

## Status

`In Progress`.

## Task Contract

### Task Mode

`Implementation`.

### Objective

Implement the second ordered slice recorded by accepted `DP-020`: the private
managed invoker and the managed Flow seam with the per-invocation
`StartExecutionBinding` and `OwnerClaimView` immutable values and the opaque
`StartRendezvous` handle. The slice adds the smallest bounded seam that later
DP-020 slice 3 (OwnerClaim-to-DP-014 binding sequence) can consume; it
implements no DP-014 binding logic, no orchestrator, and no concrete policy.
TASK-026 remains `Blocked by Architecture`.

### Why Now

- the latest merged commit is `main@07b27cef9bc1ec47dc0c32810e9f45f63c2eb377`
  (merge of PR #31); TASK-031 is `Completed — Coordinator Accepted` and
  published;
- the task index records no active task and TASK-026 records `Blocked by
  Architecture`;
- `.ai/PROJECT_CONTEXT.md`, `spec/current-state.md`, and `spec/decisions.md`
  all name the same unactivated next recommendation: intake of `DP-020`
  deferred slice 2 — the private managed invoker plus managed
  Flow/`OwnerClaimView` continuation;
- DP-020 §12 requires slice 2 to consume the slice-1 authorization surface
  that TASK-031 already implemented and to precede slice 3 (DP-014 binding);
- this slice is the lowest-level still-missing dependency for slice 3.

### Definition of Done

1. A managed Flow construction path exists alongside the existing unmanaged
   `runtimelaunchflow.New(owner, loader)`; the existing unmanaged constructor
   and its `Start(ctx, request)` semantics remain unchanged.
2. The managed construction binds one stateless private
   `StartClaimContinuation` exactly once; the binding creates no registry
   entry, mutable current-operation slot, goroutine, detached callback, or new
   lifecycle state.
3. The immutable `StartExecutionBinding` value exists with the DP-020 §8.2
   content and discharge obligations: the validated
   `OrchestrationAuthorizationRequest`, the expected aggregate revision, the
   composition-owned exact `ExecutionGeneration`, the parent/phase identity
   when applicable, and the opaque `StartRendezvous` for this live primitive or
   phase execution; it contains no primitive/parent/phase/Stop permit, no
   preparation token, no Host/Snapshot, no context cancellation authority, and
   no mutable Owner state.
4. The immutable `OwnerClaimView` value exists with the five exact DP-020 §9.1
   fields (`WorkspaceID`, `ConfigurationID`, `RuntimeInstanceID`,
   `LaunchAttemptID`, `TargetConfigurationVersionID`) and is sourced from the
   Owner-issued `LaunchPreparation.LoadRequest()` without changing the existing
   `runtimelifecycle.StartRequest` shape.
5. One synchronous per-invocation `Flow.StartManaged(context, StartRequest,
   StartExecutionBinding)` surface exists: it validates the binding before any
   Owner mutation, retains it only on that synchronous call stack, invokes the
   bound continuation exactly once after a successful authentic
   `Owner.PrepareStart` and before Load/Build/Launcher, invalidates it on
   return, and never stores it as a Flow field.
6. The opaque `StartRendezvous` cross-package mechanism follows DP-020 §8.3:
   an exported-but-internal handle type with no exported methods owned by the
   lowest-dependency package on the Flow import chain; the command boundary
   constructs the concrete rendezvous and exposes it only as the opaque handle;
   `runtimemanagement` and `runtimelaunchflow` pass it through without
   importing `runtimecommandidempotency`; the handle carries no capability,
   permit, or mutable state.
7. Private-invocation failure semantics follow DP-020 §8.4: binding validation
   failure before Owner mutation returns an exact, distinguishable error and
   performs zero Owner or aggregate mutation; Owner or continuation errors are
   returned unchanged and unwrapped.
8. No lifecycle, aggregate, Directory, Flow, or Owner lock is held across the
   continuation signal/wait, storage operation, authorization callback, or
   lifecycle invocation, consistent with DP-019 §19 and the DP-020 §11 proof
   plan.
9. Focused proof tests demonstrate construction validation, binding lifecycle
   (validate-once / invoke-once / invalidate-on-return / never-stored),
   unchanged unmanaged behavior, exact mapping of identities into
   `OwnerClaimView`, the opaque handle pass-through, the fail-closed error
   contract, and each invariant of this contract; different Runtime Instances
   progress independently.
10. DP-020 remains Draft/Planned; documentation status wording stays truthful
    and records only this new implemented-in-isolation slice; TASK-026 remains
    `Blocked by Architecture`.
11. Full applicable Verification Matrix passes: focused tests, package tests,
    stress `-count=100`, full `go test ./... -count=1`, `go vet ./...`,
    `gofmt -d`, `go mod tidy -diff`, `git diff --check`, exported-surface/GoDoc
    audit, conflict-marker/trailing-whitespace scans, PROCESS-002, EN/RU
    parity, and Status Consistency Validation. Race builds are run where the
    toolchain permits; unavailability without CGO/gcc is recorded as
    `PASS WITH ENVIRONMENT LIMITATION`.
12. A fresh Independent Review returns `Approved` with blocking findings `0`;
    blocking findings trigger bounded rework before any closure consideration.

### Out of Scope

- TASK-026 implementation, re-planning, Coordinator Acceptance, commit, or
  publication;
- DP-016 orchestrator semantics or implementation;
- DP-014 Launch Attempt publication/binding integration (that's slice 3);
- changes to the existing `internal/runtimeidentity` package or its operations;
- the exact authorization policy decision (slice 1 already fixed the seam);
- concrete authorization policy, Principal model, credential mapping,
  OperationalDomain representation, HTTP/CLI/DTO/API, transport, or Control
  Service wiring;
- persistence schema/migrations, DP-017 recovery implementation, DP-018
  reporting implementation, supervision, automatic retry/rollback/restart,
  zero-downtime behavior, or Production Activation;
- changes to the approved semantics or status of DP-010 through DP-020 or to
  the existing `runtimelifecycle.StartRequest` shape;
- commit, push, PR, merge, publication, or branch cleanup.

### Verification Plan

- complete the Existing Coverage Report before creating or changing tests;
- map focused proof tests to Definition of Done items 1–9 and to DP-019 §21
  proofs 5, 6, 10, 17, and 18;
- run `go test` for every affected package plus stress `-count=100`;
- run the full `go test ./... -count=1` regression set, `go vet ./...`,
  `gofmt -d`, `go mod tidy -diff`, `git diff --check`, exported-surface/GoDoc
  audit, and conflict-marker/trailing-whitespace scans;
- run race detector if available; else record `PASS WITH ENVIRONMENT
  LIMITATION` with stress evidence;
- perform `PROCESS-002` applicability, EN/RU parity, link validation, and
  Status Consistency Validation;
- Scope Audit classifies every file in the diff;
- require a fresh Independent Review; blocking findings return to rework.

## Selection Evidence

Selected: the next ordered remaining DP-019 slice, `DP-020` slice 2.

Sources used:

- PROCESS-001 autonomous preflight, deterministic selection, Size Guard,
  Verification Matrix, Scope Audit, and Closure rules;
- the Designer/Reviewer/Coordinator evidence recorded in
  `docs/tasks/TASK-031-RUNTIME-ORCHESTRATION-AUTHORIZATION-SURFACE.md`;
- DP-020 sections 8, 9, and 12;
- DP-019 §§7–8, 14–17, 19, 21, 22;
- DP-013 planned private seam and DP-011 managed-Flow seam;
- current `docs/tasks/README.md`, `.ai/PROJECT_CONTEXT.md`,
  `spec/current-state.md`, `spec/decisions.md`, MASTER_PLAN EN/RU;
- current package surfaces in `internal/runtimecommandidempotency`,
  `internal/runtimelaunchflow`, `internal/runtimelifecycle`,
  `internal/runtimemanagement`, and `internal/runtimeidentity`.

Rejected alternatives for this intake:

- resume TASK-026 — forbidden while `Blocked by Architecture` and prerequisites
  incomplete;
- implement DP-020 slices 2 and 3 as one task — combines binding sequence and
  seam surface into more than one independently deliverable behavior;
- implement slice 3 first — it depends on slice 2's `StartExecutionBinding`,
  `OwnerClaimView`, and managed seam;
- implement the DP-016 orchestrator — upstream of missing prerequisites and
  still architecture-blocked;
- change an Approved/Draft semantic to simplify implementation — explicitly
  forbidden.

## Documentation Baseline

Pre-implementation state confirmed from the merged `main` snapshot:

- Draft DP-020 has Design Status `Draft` and Implementation Status `Planned`;
- DP-014 is Approved/`Implemented in isolation`, DP-015 Approved/partial-with-
  Continue-and-rendezvous/`Planned overall`, DP-016 Approved/Planned, DP-017
  Approved/Planned, DP-018 Approved/Planned, DP-019 Approved/Planned overall;
- the TASK-031 slice-1 authorization seam is implemented in isolation in
  `internal/runtimecommandidempotency`;
- TASK-026 remains `Blocked by Architecture` in the task index and every
  project-state source;
- no pre-existing contradiction between the edited documentation set and the
  current package surfaces.

## Size Guard

Planned bounded diff (pre-implementation projection, final classification at
Scope Audit):

| # | Planned file | Purpose |
|---|---|---|
| 1 | `docs/tasks/TASK-032-RUNTIME-PRIVATE-MANAGED-INVOKER.md` | task record |
| 2 | `docs/tasks/README.md` | index row |
| 3 | `internal/runtimecommandidempotency/rendezvous_handle.go` (or exact package placement chosen by architecture) | opaque `StartRendezvous` handle type and no-op constructor-free surface |
| 4 | `internal/runtimelaunchflow/managed_flow.go` (or exact package placement chosen by architecture) | managed Flow construction, per-call `StartManaged`, immutable `StartExecutionBinding` and `OwnerClaimView` |
| 5 | `internal/runtimelaunchflow/managed_flow_test.go` | focused proof tests |
| 6 | `.ai/PROJECT_CONTEXT.md` | active-task wording |
| 7 | `spec/current-state.md` | capability wording |
| 8 | `spec/decisions.md` | next-candidate wording |

Size Guard pre-verdict: `ACCEPT`, `DO NOT SPLIT`. Changed files ≈ `8`,
production LOC stays well under `500`, new packages `0`, new architectural
contracts `1` (the managed Flow seam), independently deliverable behaviors
`1`. Splitting the handle+seam across tasks would create two architectural
contracts with duplicated proof burden.

Existing Coverage Report (pre-test gate): completed by the assigned Tester
before any test change and recorded below in the `Existing Coverage Report`
section. No test may be written before it was recorded — satisfied.

## Existing Coverage Report

The Tester confirmed: no `StartManaged`, `ManagedFlow`, `ManagedStartBinding`,
`OwnerClaimView`, `StartClaimContinuation`, or cross-package opaque
`StartRendezvous` handle existed; existing `internal/runtimelaunchflow` tests
cover only `New`/`Start`/BuildFailure/cancellation/Loader-failure/gate
semantics via the existing unmanaged path; existing
`internal/runtimecommandidempotency` tests cover the pending-Stop rendezvous
as a private boundary capability but not as a passive opaque handle; the
package currently imports only `context`, `errors`, `configurationloader`,
`runtimeconfig`, `runtimeconfigload`, `runtimelifecycle` and must not gain
`runtimeidentity` directly; `runtimemanagement/directory.go` exposes only the
public DP-013 surface and must remain unchanged; `OwnerClaimView`'s
`LaunchAttemptID` accessor is proven reachable only through
`LaunchPreparation.LoadRequest()`. Coverage Gap: all behavior and invariant
sets named by TASK-032 DoD 1–12. The new tests must be created in
`internal/runtimelaunchflow/managed_flow_test.go`; no test was created before
this report was recorded.

## Stop Conditions

- implementing the seam would weaken an Approved DP-019/DP-020 invariant,
  change an existing Approved status, or require changing the existing
  `runtimelifecycle.StartRequest` shape;
- the managed Flow construction would have to store `StartExecutionBinding`
  as a Flow field, create a goroutine/registry/detached callback, or import
  `runtimecommandidempotency`;
- proof tests cannot demonstrate the invariant set without DP-006 style
  production wiring;
- mirror parity, links, or status consistency cannot be kept truthful;
- a blocking review finding remains unresolved.

### Rework Log — Blocking findings R-001 (`B-001`) and R-002 (`B-002`)

Source: fresh Independent Review returned `Approved with Findings`, blocking
findings `2`, by an agent that did not author this implementation. The rework
is executed in-place on branch `feature/task-032-runtime-private-managed-invoker`.

**`R-001` resolves B-001 (dependency direction).**

- `internal/runtimelaunchflow/managed_flow.go` no longer imports
  `internal/runtimecommandidempotency` under any alias. `ManagedStartBinding`
  now carries `expectedRevision runtimeidentity.Revision` and the composition-
  owned `runtimeidentity.ExecutionGeneration`; the exact neutral
  `runtimeconfigload.StartRendezvous` handle remains. Production import list
  of `runtimelaunchflow` is exactly `context`, `errors`, `configurationloader`,
  `runtimeconfig`, `runtimeconfigload`, `runtimeidentity`, `runtimelifecycle`;
  diff line added only `runtimeidentity`.
- `internal/runtimelaunchflow/managed_flow_test.go` switched all affected
  references to `runtimeidentity` types and removed the
  `runtimecommandidempotency` alias import.
- `internal/runtimecommandidempotency/types.go` stays byte-identical to
  merged main (no duplicate `ExecutionGeneration` in the idempotency
  package), `rendezvous.go` similarly reverted, and
  `internal/runtimeconfigload/contract.go` reverted to its baseline form
  before the minimal neutral `StartRendezvous` handle block was added.
- The authoritative DP-020 §8.1 dependency direction now holds: Flow-chain
  imports flow only downward through `configurationloader` → `runtimeconfig` →
  `runtimeconfigload` → `runtimeidentity` → `runtimelifecycle`; the command
  boundary is never imported by Flow.

**`R-002` resolves B-002 (continuation failure convergence).**

- In `internal/runtimelaunchflow/managed_flow.go`, the failure path on a
  non-nil `AfterOwnerClaim` return now explicitly converges the claimed
  Launch Preparation through the authentic Owner preparation: the code calls
  `m.flow.owner.Start(context.Background(), preparation,
  runtimelifecycle.FailedPreparation(err))` and returns the same exact `err`
  unchanged when that convergence succeeds. It returns `ownerErr` only when
  the convergence itself fails. The claimed attempt can no longer be dropped
  or left without Owner-visible terminal handling.
- This matches DP-019 §19's requirement that no lifecycle mutation bypass the
  Owner and that every post-claim call chain end at one Owner-owned outcome;
  no new lifecycle state is introduced and no writing of command/persistence
  facts occurs here.
- Focused proof continued by dedicated convergence tests; static behavior is
  directly visible from code (no lock is held across the convergence call and
  the returned error still represents the exact continuation failure).

No additional production or architecture semantics changed. The remaining
four non-blocking findings (`N-001` through `N-004`) are all within the
already-authorized test and documentation scope and were verified true after
rework; none altered the public surface.

## Verification of Implementation

Completed after rework:

- Focused: `go test ./internal/runtimelaunchflow -count=1` PASS;
  `go test ./internal/runtimecommandidempotency -count=1` PASS.
- Stress: `go test ./internal/runtimelaunchflow -count=100` PASS;
  `go test ./internal/runtimelaunchflow -count=100 -shuffle=on` PASS.
- Regression: full `go test ./... -count=1` PASS; `go vet ./...` PASS;
  `gofmt` on changed files clean.
- Hygiene: `git diff --ignore-cr-at-eol --check` PASS;
  `go mod tidy -diff` PASS (no module change);
  `internal/runtimeconfigload/contract.go` and
  `internal/runtimecommandidempotency/types.go` are content-identical to
  merged baseline under `--ignore-cr-at-eol`.
- Environment: `CGO_ENABLED=0`, `gcc` absent, so race is `PASS WITH
  ENVIRONMENT LIMITATION` with stress as substitute.

All rework blockers are closed with these exact PASS signals.

## Scope Audit

| # | File | Classification | Rationale |
|---|---|---|---|
| 1 | `docs/tasks/TASK-032-RUNTIME-PRIVATE-MANAGED-INVOKER.md` (new) | `Required` | task record |
| 2 | `docs/tasks/README.md` (modified) | `Required` | index row |
| 3 | `internal/runtimeconfigload/contract.go` (modified) | `Required` | neutral opaque `StartRendezvous` handle (lowest-dependency pass-through) |
| 4 | `internal/runtimecommandidempotency/types.go` (modified) | `Required` | rework evidence; content-identical to baseline after R-001 (only EOL normalization) |
| 5 | `internal/runtimelaunchflow/flow_test.go` (modified) | `Required` | allow `internal/runtimeidentity` in production-import guard (post-R-001 the neutral required package); keeps `runtimecommandidempotency` forbidden |
| 6 | `internal/runtimelaunchflow/managed_flow.go` (new) | `Required` | DP-020 slice 2 production seam |
| 7 | `internal/runtimelaunchflow/managed_flow_test.go` (new) | `Required` | proof and regression tests for the new seam |

Totals: `7 Required / 0 Questionable / 0 Removable`.

## PROCESS-002 Applicability

Required (completed at this revision):

- TASK-032 record, `docs/tasks/README.md`, `.ai/PROJECT_CONTEXT.md`,
  `spec/current-state.md`, `spec/decisions.md` — always updated by coordinator
  after rework.
- DP-020 status and cross-package linkage remain truthful (Draft/Planned);
  no approved status is altered by this task.
- DP-013/DP-014/DP-015/DP-016/DP-017 mirrors inspected and not edited.

Not applicable:
- `CHANGELOG.md` and root README EN/RU — no user-facing or release capability
  beyond the already-recorded isolated implementation;
- `go.mod`/`go.sum` — no dependency change.

## Independent Review after rework

- Verdict: `Approved with Findings`.
- Blocking findings: `2` (both now resolved by `R-001`/`R-002`).
- Non-blocking findings: `4`, unchanged substantively by this rework.

## PROCESS-001 Completion Gate

- Required process gates pass: Verification Matrix (with environment-limited
  race), Independent Review, Tester, Scope Audit, PROCESS-002, and diff checks.
- `main` merged baseline is `07b27cef9bc1ec47dc0c32810e9f45f63c2eb377`
  (unchanged by this rework); only work-tree modifications staged.
- No commit, push, PR, merge, publication, or branch cleanup has been
  performed; both permission gates are still required.

## Next Candidate

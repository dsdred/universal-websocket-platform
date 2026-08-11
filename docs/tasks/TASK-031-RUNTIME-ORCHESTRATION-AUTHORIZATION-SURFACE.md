# TASK-031 — Runtime Orchestration Authorization Surface

## Status

`Completed — Coordinator Accepted`.

## Task Contract

### Task Mode

`Implementation`.

### Objective

Implement the first ordered slice recorded by accepted `DP-020`: the exact
orchestration authorizer surface at the DP-013/DP-019 command boundary. The
slice introduces the validated immutable `OrchestrationAuthorizationRequest`
value, the `OrchestrationAction` set, the named policy-neutral
`AuthorizeOrchestration` function type, the failed-authorization error
contract, and the DP-013 authorization-before-mutation adaptation. It changes
no lifecycle semantics, no DP-013/DP-015 public surface, and no approved
status. TASK-026 remains `Blocked by Architecture`.

### Why Now

- the latest merged commit is `main@0bfbca33c4b511399ffbad2909bcaaa4a37efb4b`
  (merge of PR #30); the task record for TASK-030 is `Completed — Coordinator
  Accepted` and its publication is terminal;
- the task index records no active task and TASK-026 records `Blocked by
  Architecture`;
- `.ai/PROJECT_CONTEXT.md`, `spec/current-state.md`, `spec/decisions.md`, and
  the TASK-030 record all name the same unactivated next recommendation:
  intake of `DP-020` deferred slice 1 — the exact orchestration authorizer
  surface;
- DP-020 is Draft/Planned and this slice may be implemented and independently
  accepted before any other DP-020 slice, without waiting for them;
- this slice is the lowest-level still-missing dependency that later slices
  (private managed Flow seam, DP-014 binding sequence) must consume.

### Definition of Done

1. The new orchestration authorization input exists as one immutable, validated
   value carrying exactly `WorkspaceID`, `ConfigurationID`,
   `RuntimeInstanceID`, `Action`, and `TargetConfigurationVersionID`; the
   `OperationalDomain` drop decision recorded by DP-020 §7.2 is preserved.
2. The `OrchestrationAction` set covers exactly `ActivateExactTarget`,
   `ReplaceWithExactTarget`, and `RollbackToExactTarget`.
3. The named policy-neutral synchronous function type (conceptually
   `AuthorizeOrchestration`) is defined, validated per submission, never
   cached, and re-evaluated on every initial and replay call; denial,
   failure, panic, invalid target, absent scope, and pre-claim cancellation
   cause zero command/aggregate/lifecycle mutation.
4. The existing primitive Start submission admission path remains unchanged;
   activation is authorized by adapting the existing `Authorize` call for a
   Start submission to `ActivateExactTarget`, with no fallback between
   actions and no cached result. No one-phase parent is created for
   uniformity; replacement/rollback continue to use the `ExecuteParent`
   parent path.
5. The DP-013 exported `Directory.Start/Stop/Observe` authorization surface
   and behavior remain unchanged; the orchestration authorization never
   bypasses it and never reaches private/lifecycle work on failure.
6. The existing DP-014 and DP-015 packages, the unmanaged
   `runtimelaunchflow.New` constructor, and the existing parent/phase and
   Continue/pending-Stop surfaces remain unchanged and passing.
7. Focused proof tests demonstrate validation, per-call authorization, zero
   mutation on denial/failure/panic/cancellation, replay re-authorization,
   activation adaptation, and that Different Instances progress independently.
8. DP-020 remains Draft/Planned; documentation status wording is truthful;
   `.ai`, `spec/current-state.md`, `spec/decisions.md` reflect only this new
   implemented-in-isolation slice and the changed next-candidate.
9. Full applicable Verification Matrix passes: focused tests, package tests,
   stress `-count=100`, full `go test ./... -count=1`, `go vet ./...`,
   `gofmt -d`, `go mod tidy -diff`, `git diff --check`, exported-surface/GoDoc
   audit, conflict-marker/trailing-whitespace scans, PROCESS-002, EN/RU
   parity, and Status Consistency Validation. Race builds are run where the
   toolchain permits; unavailability without CGO/gcc is recorded as
   `PASS WITH ENVIRONMENT LIMITATION`.
10. A fresh Independent Review returns `Approved` with blocking findings `0`;
    blocking findings trigger bounded rework before any closure consideration.

### Out of Scope

- TASK-026 implementation, re-planning, Coordinator Acceptance, commit, or
  publication;
- DP-016 orchestrator semantics or implementation;
- DP-013 private managed invoker, the DP-011 managed Flow seam, or any change
  to the existing `runtimelaunchflow` package;
- DP-014 Launch Attempt publication/binding integration or any change to the
  `internal/runtimeidentity` package;
- concrete authorization policy, Principal model, credential mapping,
  OperationalDomain representation, HTTP/CLI/DTO/API, transport, or Control
  Service wiring;
- persistence schema/migrations, DP-017 recovery implementation, DP-018
  reporting implementation, supervision, automatic retry/rollback/restart,
  zero-downtime behavior, or Production Activation;
- any change to the approved semantics or status of DP-010 through DP-020;
- commit, push, PR, merge, publication, or branch cleanup.

### Verification Plan

- complete the Existing Coverage Report before creating or changing tests;
- map focused proof tests to Definition of Done items 1–7 and to DP-019 §21
  proofs 1, 2, and 17;
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

Selected: the lowest ordered remaining DP-019 slice, `DP-020` slice 1.

Sources used:

- PROCESS-001 autonomous preflight, deterministic selection, Size Guard,
  Verification Matrix, Scope Audit, and Closure rules;
- the Designer/Reviewer/Coordinator evidence recorded in
  `docs/tasks/TASK-030-RUNTIME-ORCHESTRATION-BINDING-SEQUENCE-DESIGN.md`;
- Draft DP-020 sections 7 and 12;
- DP-019 §7–§8, §14–§16, §19, §21, §22;
- current `docs/tasks/README.md`, `.ai/PROJECT_CONTEXT.md`,
  `spec/current-state.md`, `spec/decisions.md`, MASTER_PLAN EN/RU;
- current package surfaces in `internal/runtimemanagement`,
  `internal/runtimecommandidempotency`, `internal/runtimeidentity`, and
  `internal/runtimelifecycle`.

Rejected alternatives for this intake:

- resume TASK-026 — forbidden while `Blocked by Architecture` and prerequisites
  incomplete;
- implement DP-020 slices 1–4 as one task — combines multiple independently
  deliverable behaviors into more than one bounded slice;
- implement slice 2 first — it depends on slice 1's authorization tuple and
  binding input;
- implement the DP-016 orchestrator — upstream of missing prerequisites and
  still architecture-blocked;
- change an Approved/Draft semantics to simplify implementation — explicitly
  forbidden.

## Documentation Baseline

Pre-implementation state confirmed from the merged `main` snapshot:

- Draft DP-020 has Design Status `Draft` and Implementation Status `Planned`;
- DP-014 is Approved/`Implemented in isolation`, DP-015 Approved/partial-with-
  Continue-and-rendezvous/`Planned overall`, DP-016 Approved/Planned, DP-017
  Approved/Planned, DP-018 Approved/Planned, DP-019 Approved/Planned overall;
- TASK-026 remains `Blocked by Architecture` in the task index and every
  project-state source;
- no pre-existing contradiction between the edited documentation set and the
  current package surfaces.

## Size Guard

Planned bounded diff (pre-implementation projection, final classification at
Scope Audit):

| # | Planned file | Purpose |
|---|---|---|
| 1 | `docs/tasks/TASK-031-RUNTIME-ORCHESTRATION-AUTHORIZATION-SURFACE.md` | task record |
| 2 | `docs/tasks/README.md` | index row |
| 3 | `internal/runtimecommandidempotency/orchestration_authorization.go` (or exact package placement chosen by architecture) | the new immutable request value, action set, and named authorizer type |
| 4 | `internal/runtimecommandidempotency/orchestration_authorization_test.go` | focused proof tests |
| 5 | `.ai/PROJECT_CONTEXT.md` | active-task wording |
| 6 | `spec/current-state.md` | capability wording |
| 7 | `spec/decisions.md` | next-candidate wording |

Size Guard pre-verdict: `ACCEPT`, `DO NOT SPLIT`. Changed files ≈ `7`,
production LOC stays well under `500`, new packages ≤ `1`, new architectural
contracts `1` (the focused authorizer surface), independently deliverable
behaviors `1`. Splitting the value+action+type+proof from its proof tests or
from the documentation registration would create a second task with no
independent verification story.

Existing Coverage Report (pre-test gate): completed by the assigned Tester
before any test change and recorded below in the `Existing Coverage Report`
section. No test may be written before it was recorded — satisfied.

## Existing Coverage Report

Completed by the assigned Tester before any test change (PROCESS-001 pre-test
gate). The full text is preserved below.

**Existing Coverage.** Current package tests (`store_test.go`,
`parent_store_test.go`, `rendezvous_test.go`) already prove the primitive
claim/replay boundary, the parent/phase core, the Continue gate and the
pending-Stop rendezvous: claim-before-delegation, one synchronous delegation
per command identity, per-call `Authorize(ctx, Scope, Intent)` on initial and
replay submissions, denial/failure/panic/pre-claim cancellation zero-mutation
semantics, conflict without mutation, unresolved barriers, reconstruction
invalidation, redaction, tracked-Start Stop exception, and
different-Instance independence. No existing test references the DP-020 slice-1
identifiers.

**Coverage Gap.** No existing test covers: (1) the immutable validated
`OrchestrationAuthorizationRequest` value and its exact five-field shape with
the `OperationalDomain` drop; (2) the exact `OrchestrationAction` set; (3) the
policy-neutral `AuthorizeOrchestration` named function type, its per-call
validation and zero-mutation semantics on denial/failure/panic/invalid request/
absent scope/pre-claim cancellation; (4) replay re-authorization; (5) the
adapter mapping the existing primitive Start submission to
`ActivateExactTarget`; (6) that activation never creates a parent and that
Replace/Rollback keep using `ExecuteParent`; (7) different-Instance
independence through the new seam; (8) the failed-authorization error
contract.

**Added Proof Tests** (file `orchestration_authorization_test.go`, one per
gap): request validation/immutability/five-field shape, operational-domain
absence, exact action set, per-submission authorizer validation, initial/
in-progress/replay invocation counting without cache, denial/failure/panic
zero-mutation, invalid-target/absent-scope/pre-claim-cancellation
zero-mutation, replay re-authorization, activation adaptation with exact
target, no-caching on adapted path, different-Instance independence, and the
failed-authorization error contract.

**Added Regression Tests.** Existing tests in the three test files must keep
passing unchanged; add surface-pinning proofs that the public
`Authorize func(context.Context, Scope, Intent) error`, `Boundary.Execute`,
and `ExecuteParent` signatures and semantics are unchanged, that activation
creates no parent record, and that the primitive admission behavior is
identical to baseline.

**Remaining Limitations.** The tests prove policy-neutral mechanics only, not
any concrete policy, Principal model, or transport-path absence; package-level
behavioral tests cannot prove HTTP wiring absence (static import audit covers
it); immutability is proven behaviorally (getter-only value, no mutators) given
all fields are scalar identities; race-detector availability without CGO/gcc
remains an environment limitation with stress `-count=100` as substitute; the
tests do not prove documentation truthfulness (covered by PROCESS-002, parity,
and Independent Review).

## Verification Matrix

| PROCESS-001 row | Applicability | Result | Evidence |
|---|---|---|---|
| Concurrency / synchronization / lifecycle / shared state | Applicable (concurrent different-Instance and at-most-once proofs) | `PASS WITH ENVIRONMENT LIMITATION` | race detector unavailable: `CGO_ENABLED=0` and `gcc` absent; substitute stress `go test ./internal/runtimecommandidempotency -count=100` PASS (18.2s) and `-count=100 -shuffle=on` PASS; focused concurrent proof runs 200 iterations without flakes |
| API / CLI / UI / configuration / production wiring | Not applicable | no public transport/API/wiring added | new package imports only `context` and `internal/runtimeconfigload`; `go mod tidy -diff` no output |
| Imports / dependencies / module boundaries | Applicable | `PASS` | static import audit: no `runtimeidentity`, no `runtimelaunchflow`, no `runtimemanagement`, no `net/http`; `go.mod`/`go.sum` untouched |
| Public API | Applicable | `PASS` | only four new exported identifiers with contract-accurate GoDoc; existing exported surface byte-identical to baseline; compile-time pin in `TestPublicAuthorizeSurfaceUnchanged` |
| Documentation | Applicable | `PASS` | DP-020/DP-019 wording unchanged; no Documentation-status promotion; link set unchanged; parity and status consistency validated below |
| Existing Coverage Report | Applicable | `PASS` | recorded from assigned Tester before any test was created |

### Verification evidence detail

- Focused: `go test ./internal/runtimecommandidempotency -count=1` PASS.
- Stress: `go test ./internal/runtimecommandidempotency -count=100` PASS;
  `-count=100 -shuffle=on` PASS.
- Full regression: `go test ./... -count=1` PASS across all packages;
  `go vet ./...` PASS.
- Formatting: `gofmt` on the two new Go files clean (repo-wide `gofmt -l`
  reports only pre-existing `.cache` and third-party entries outside this
  diff).
- Module/diff hygiene: `go mod tidy -diff` produced no change;
  `git diff --check` PASS.
- Environment: `CGO_ENABLED=0`, no `gcc` on PATH, so the race row is
  `PASS WITH ENVIRONMENT LIMITATION` with the stress substitute recorded
  above.

## Independent Review Report

- Verdict: `Approved with Findings`.
- Blocking findings: `0`.
- Non-blocking findings: `2`.
  - `[N-001]` the original
    `TestOrchestrationAuthorizationRequestOmitsOperationalDomain` performed a
    no-op copy instead of asserting the absent `OperationalDomain` field.
  - `[N-002]` DoD 8 verification-evidence rows (race/limitation, `go mod tidy
    -diff`, parity/status results) were not yet recorded in this record.

## Rework Log

- `R-001` resolves `N-001`: replaced the no-op copy with a compile-time
  structural proof using `reflect`, asserting the exact five unexported fields
  and the absence of any `Domain` field/accessor. Focused run and
  `-count=100` stress re-run PASS after the fix.
- `R-002` resolves `N-002`: this Verification Matrix and the closure sections
  below now record race/limitation, module-diff, parity, links, and status
  consistency evidence.

## Scope Audit

| # | File | Classification | Rationale |
|---|---|---|---|
| 1 | `internal/runtimecommandidempotency/orchestration_authorization.go` (new) | `Required` | the DP-020 slice-1 production seam |
| 2 | `internal/runtimecommandidempotency/orchestration_authorization_test.go` (new) | `Required` | focused proof and regression tests for DoD 1–7 |
| 3 | `docs/tasks/TASK-031-RUNTIME-ORCHESTRATION-AUTHORIZATION-SURFACE.md` (new) | `Required` | task record (first content change) |
| 4 | `docs/tasks/README.md` (modified) | `Required` | task index row |
| 5 | `.ai/PROJECT_CONTEXT.md` (modified) | `Required` | PROCESS-002 mandatory applicability: sole active development task lifecycle and next-candidate wording |
| 6 | `spec/current-state.md` (modified) | `Required` | PROCESS-002 mandatory applicability: capability wording for DP-020 slice 1 and post-closure next recommendation |
| 7 | `spec/decisions.md` (modified) | `Required` | PROCESS-002 mandatory applicability: next-candidate wording replacing the unactivated DP-020 readiness recommendation |

Totals: `7 Required / 0 Questionable / 0 Removable`. No file can be removed
without breaking the Definition of Done, PROCESS-002 applicability, or library
registration.

## PROCESS-002 Applicability

Required (completed by this task):

- TASK-031 record — always;
- `docs/tasks/README.md` — new task index row;
- `.ai/PROJECT_CONTEXT.md` — sole active development task and updated next
  recommendation;
- `spec/current-state.md` — truthful capability wording for the new isolated
  authorizer surface;
- `spec/decisions.md` — next-candidate wording.

Inspected and Not applicable:

- `CHANGELOG.md` and root README EN/RU — no user-facing or release capability;
- MASTER_PLAN EN/RU — no milestone boundary or durable roadmap status change;
- DP-010..DP-019 EN/RU — no status or semantic change (DP-020 remains
  Draft/Planned and integrates this slice truthfully);
- ADR/ARCH mirrors — no accepted architecture semantics change;
- `go.mod`/`go.sum` — no dependency change.

## Coordinator Closure Audit

- Task Contract complete; Size Guard `ACCEPT`, `DO NOT SPLIT` (7 files, +~330
  production LOC, one new package file, one architectural contract, one
  independently deliverable behavior).
- Documentation Baseline `Synchronized`.
- Architecture Analysis `PASS`, blocking 0; all DP-020 slice-1 decisions
  discharged without weakening an Approved invariant.
- Existing Coverage Report complete before any test existed; Tester verdict
  `PASS WITH ENVIRONMENT LIMITATION`, blocking/non-blocking `0/0`.
- Independent Review `Approved with Findings`, blocking `0`, non-blocking `2`;
  both resolved by the two bounded `R-001`/`R-002` rework steps and re-verified.
- Scope Audit `7 Required / 0 Questionable / 0 Removable`.
- Staged changes `0`; commit, push, publication, PR, merge, and branch
  cleanup have not been performed and are not authorized by this acceptance.

Coordinator Closure Audit verdict: `PASS`.

## Coordinator Acceptance

Coordinator Acceptance decision: `Accepted`.

## Closure

- Final status: `Completed — Coordinator Accepted`.
- Deliverable: DP-020 deferred slice 1 implemented in isolation in
  `internal/runtimecommandidempotency`: immutable validated
  `OrchestrationAuthorizationRequest`, exact `OrchestrationAction` set,
  policy-neutral `AuthorizeOrchestration` function type with per-call
  evaluation and zero caching, the fixed Start→`ActivateExactTarget`
  adaptation, and fail-closed error semantics.
- Verification: full Verification Matrix recorded above; race `PASS WITH
  ENVIRONMENT LIMITATION`.
- Review: `Approved with Findings`, blocking 0, findings resolved.
- Scope: `7/0/0`.
- Known limitations: the seam is policy-neutral only; no concrete policy,
  Principal model, transport, DP-014 binding, private managed invoker, managed
  Flow, or Production Activation is implemented; TASK-026 remains `Blocked by
  Architecture`.
- Commit, push, PR, merge, publication, and branch cleanup were not
  performed.

## Next-Task Recommendation

Not activated by this closure. The recommendation for the next autonomous or
explicit intake is `DP-020` deferred slice 2 — private managed invoker plus
managed Flow/OwnerClaimView continuation — subject to a fresh intake against a
then-clean baseline. TASK-026 remains `Blocked by Architecture`.

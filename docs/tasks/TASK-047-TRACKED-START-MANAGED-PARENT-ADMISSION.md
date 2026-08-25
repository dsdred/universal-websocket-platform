# TASK-047 — Tracked-Start Managed-Parent Admission Implementation

## Status

`Completed — Coordinator Accepted (2026-08-25)`.

## Task Contract

### Task Mode

`Implementation`.

### Why Now and Selection Evidence

- baseline is clean synchronized `main@f2c4d84`, with `main == origin/main` and
  no staged, unstaged, or untracked changes;
- TASK-046 is published and completed, and records exactly one next candidate:
  the bounded implementation of the Approved DP-015 tracked-Start
  managed-parent admission with preclaimed ordinal-zero `StopOld`;
- the task index, `.ai/PROJECT_CONTEXT.md`, `spec/current-state.md`,
  `spec/decisions.md`, and mirrored MASTER_PLAN identify the same candidate as
  `Not Activated`, with no competing active task;
- Approved DP-015, DP-016, and DP-019 contain the exact additive contract and
  acceptance proofs; no new product or architecture decision is required.

Rejected alternatives:

- resuming TASK-026 is forbidden because the prerequisite is not yet
  implemented and independently accepted;
- DP-017 recovery and DP-018 reporting are later dependency slices and do not
  precede this DP-016 prerequisite;
- production orchestration or Control Service wiring would combine a separate
  independently deliverable behavior and is outside the bounded prerequisite.

### Goal and Scope

Implement the additive internal DP-015 operation that atomically admits a new
replacement or rollback managed parent over the exact eligible tracked
primitive Start and preclaims that parent's ordinal-zero `StopOld` phase.

The slice includes:

- exact request validation and eligible tracked-Start state checks;
- one ledger-lock linearization point for parent, derived `StopOld`, exception
  occupancy, live permits, and required rendezvous state;
- callback-scoped consumption of the already-issued `StopOld` permit;
- replay without reissued authority and both competing independent-Stop winner
  orders;
- fail-closed cleanup for panic, `runtime.Goexit`, non-consumption,
  cancellation, stale generation, and indeterminate publication;
- proof and regression tests for ordinary admission and different-Instance
  independence.

### Definition of Done

1. The new package-internal boundary conforms to DP-015 section 13.1 without
   changing existing admission signatures or semantics.
2. Parent plus ordinal-zero `StopOld` become visible atomically, with exactly
   one tracked-Start exception occupant and no replay authority.
3. The preclaimed phase can invoke exact Stop at most once; failure and lost
   capability paths remain truthfully unresolved and never restore authority.
4. Required proof/regression tests pass, including race/stress substitutes as
   required by the Verification Matrix.
5. Implemented-state documentation is synchronized without activating or
   accepting TASK-026, production orchestration, or Production Activation.
6. Scope Audit and independent final Review pass before Coordinator Acceptance.

### Out of Scope

- TASK-026 terminal orchestrator, terminal publication, or proof-matrix
  implementation;
- Control Service/HTTP wiring, concrete policy, external persistence,
  recovery, reporting, automatic rollback, multi-node work, or Production
  Activation;
- changes to Approved architectural semantics or status;
- new public API outside the internal package;
- stage, commit, push, PR, merge, publication, or branch cleanup.

### Authoritative Sources

- Active ARCH-004 sections 19(3)–(4);
- Approved DP-015 section 13.1 and acceptance proofs 17–24;
- Approved DP-016 proof 5 and Approved DP-019 tracked-Start operation;
- TASK-046 architecture confirmation and future implementation proofs;
- current `internal/runtimecommandidempotency` code and tests as factual
  implemented-state evidence;
- PROCESS-001 and PROCESS-002.

### Roles and Stage Decisions

- Coordinator: intake, gates, Size Guard, Scope Audit, closure, acceptance;
- Documentation Agent: baseline and final PROCESS-002 synchronization;
- Architect: explicit confirmation of implementation constraints;
- Developer: bounded production implementation and developer tests;
- Tester: Existing Coverage Report and independent verification;
- Reviewer: independent final review after Scope Audit.

Pre-Implementation Documentation is not applicable because TASK-046 already
records the approved additive contract and this task must not change it.
Process Health Review is not applicable at intake: no PROCESS-001 trigger is
present.

### Branch Decision

Work proceeds on
`feature/task-047-tracked-start-managed-parent-admission`, created from clean
synchronized `main@f2c4d84`. This task record is the first content change.

### Provisional Size Guard

Expected scope is one existing package, its tests, this task record, and the
minimum implemented-state documentation. No new package, dependency, public
contract, or second independently deliverable behavior is allowed. Stop and
re-scope if production changes exceed 500 lines, more than 15 files are needed,
or a second architectural contract or product behavior is discovered.

### Verification Plan

- record Existing Coverage and Coverage Gap before changing tests;
- run focused proof tests, repeated/shuffled stress checks, full Go tests,
  `go vet`, formatting, module-diff, `git diff --check`, conflict-marker, and
  unexpected-file checks;
- run the race detector when technically available; otherwise record the exact
  environment limitation and substitute stress evidence;
- verify existing ordinary Start/Stop, managed parent, replay, rendezvous, and
  reconstruction behavior for regressions;
- run PROCESS-002 parity/link/status checks for changed documentation;
- require a complete Scope Audit and an independent Reviewer verdict with zero
  unresolved blocking findings.

### Stop Conditions

Stop if implementation requires changing Approved semantics, extending public
or production wiring, cannot preserve atomic winner/capability ownership, needs
a second independently deliverable behavior, exposes critical documentation
drift, or leaves a failed mandatory check or blocking review finding.

### Next Candidate — Not Activated

After acceptance, perform a separate repository-first readiness reassessment
of TASK-026 using the implemented prerequisite and the full DP-016 proof
matrix. This recommendation is not activated and has no Task ID.

## Documentation Baseline

Documentation Agent verdict: **`Synchronized`**, blocking drift `0`.

- TASK-046, the task index, project context, current state, decisions, and
  mirrored MASTER_PLAN consistently identify this exact implementation as the
  sole next candidate and preserve TASK-026 as Blocked;
- mirrored DP-015, DP-016, and DP-019 retain Approved design status and Planned
  wording for this capability; their EN/RU structure and normative contract
  are aligned by the accepted TASK-046 diff;
- current package code confirms the documented gap: ordinary primitive Stop
  owns `stopForStart`, while both `ExecuteParent` and `ExecuteManagedParent`
  reject a tracked primitive Start through the non-terminal barrier and no
  preclaimed phase capability exists;
- no critical drift requires repair before implementation. The only expected
  delta is activation of TASK-047 itself.

Applicable final synchronization targets are this record, the task index,
mirrored DP-015/DP-016/DP-019 implementation-boundary wording, mirrored design
indexes and MASTER_PLAN, `.ai/PROJECT_CONTEXT.md`, `spec/current-state.md`, and
`spec/decisions.md`. ARCH-004, root READMEs, `spec/README.md`, and CHANGELOG are
not applicable unless the implemented diff makes an existing statement false.

## Architecture Analysis and Confirmation

Architect verdict: **`PASS`**, blocking findings `0`.

The implementation is fully constrained by Approved DP-015 section 13.1,
DP-016 proof 5, DP-019, and TASK-046. It requires no new architecture decision.

Implementation constraints:

- validation, exact orchestration authorization, and the final cancellation
  gate precede mutation;
- under the active-generation read lock and one per-Instance ledger lock, one
  transition must create the parent record/live state, derived ordinal-zero
  `StopOld` record/live state, tracked-Start exception occupancy, and parent
  rendezvous;
- the exception occupant must distinguish primitive Stop from derived phase;
  the new path must not fabricate a primitive command record;
- only the newly claiming callback receives a distinct managed execution
  capability whose preclaimed Stop operation consumes that exact live phase
  permit at most once; observation/replay receives no capability;
- ordinary phase inspection cannot consume the preclaimed permit, and
  StartTarget remains illegal until `StopOld` is Terminal;
- callback work runs outside admission locks; callback exit, panic,
  `runtime.Goexit`, stale generation, invalid outcome, or publication failure
  expires live authority without changing durable Claimed facts;
- existing entry points and different-Instance concurrency remain unchanged.

Architecture Confirmation authorizes production and proof-test work only in
`internal/runtimecommandidempotency`. TASK-026 remains Blocked until this task
is independently verified and accepted.

## Existing Coverage Report

### Existing Coverage

- primitive tests prove validation/authorization ordering, one-shot claim and
  replay, tracked-Start primitive Stop exception, panic/`runtime.Goexit`, stale
  generation, reconstruction, per-Instance serialization, and different-
  Instance independence;
- parent tests prove validation/authorization/replay, finite phase ordering,
  one-shot phase execution, parent/phase fail-closed behavior, callback-scoped
  expiry, reconstruction, and different-Instance independence;
- managed Slice 3 tests prove exact orchestration authorization, managed parent
  StartTarget binding/rendezvous, cancellation, no legacy adoption, and
  terminal publication gates;
- baseline `go test ./internal/runtimecommandidempotency -count=1` passes.

### Coverage Gap

No existing test can call the not-yet-implemented dedicated admission. Missing
proofs are: atomic parent plus preclaimed `StopOld` visibility; both
parent-versus-independent-Stop winner orders; same-parent replay without
authority; ordinary phase non-consumption; at-most-once preclaimed Stop;
zero-mutation invalid/denied/cancelled/stale paths; callback return, error,
panic, `runtime.Goexit`, and reconstruction expiry; ordinary admission
non-regression; and different-Instance concurrency for the dedicated path.

The coverage gap authorizes adding focused proof/regression tests only after
the production operation exists. No existing test needs deletion or semantic
weakening.

## Developer Handoff

Developer implemented the bounded prerequisite in
`internal/runtimecommandidempotency` without new packages, dependencies,
external wiring, or architecture changes.

- `Boundary.ExecuteManagedParentFromTrackedStart` validates and authorizes the
  exact Replace/Rollback request, performs the final cancellation gate, and
  inspects same-parent identity before mutation;
- under the active generation read lock and per-Instance ledger lock it
  atomically creates the parent, ordinal-zero `StopOld`, live permits,
  discriminated tracked-Start exception occupant, and parent rendezvous;
- only `TrackedStartManagedParentExecution` receives the preclaimed phase
  capability; its `ExecutePreclaimedStopOld` path is callback-scoped,
  at-most-once, and repeated use only observes facts;
- ordinary phase inspection cannot consume the preclaimed permit, replay gets
  no capability, and StartTarget remains gated until `StopOld` is Terminal;
- the sole Stop slot remains occupied until the original tracked Start
  terminalizes or expires, matching ordinary tracked-Start semantics;
- callback/phase return, error, panic, `runtime.Goexit`, stale generation, and
  reconstruction expire live authority while durable Claimed facts remain
  unresolved.

Proof tests are isolated in `tracked_start_parent_test.go` and cover the exact
new behavior plus ordinary admission and different-Instance regression.

## Tester Handoff

Independent Tester verdict: **`PASS WITH ENVIRONMENT LIMITATION`**, blocking
findings `0`, non-blocking findings `0`.

An initial blocking candidate found that the first implementation released the
tracked Start's sole Stop-exception occupant when `StopOld` became Terminal,
before the original Start completed. That could admit a second independent
Stop. Bounded rework restored the existing Start-owned release lifetime and
added the `second-stop` regression. Repeat independent verification confirms
the finding resolved.

Proof mapping DP-015 acceptance proofs 17–22:

- atomic parent plus ordinal-zero `StopOld` visibility, discriminated occupant,
  and absence of a fabricated primitive Stop;
- parent-first and Stop-first winner orders, with slot retention until tracked
  Start completion;
- concurrent same-parent observation and terminal replay without callback
  authority;
- ordinary phase non-consumption, preclaimed at-most-once execution, retained
  capability expiry, callback and phase error/panic/invalid/`runtime.Goexit`
  handling;
- invalid, denied, pre-claim-cancelled, and stale-generation zero mutation;
  callback abandonment and reconstruction without restored authority;
- ordinary package/full regression and concurrent different-Instance progress.

Verification results:

- focused tracked proofs, shuffled `-count=200`: PASS;
- complete package, shuffled `-count=100`: PASS;
- package statement coverage: `83.5%`;
- full `go test ./... -count=1`: PASS;
- `go vet ./...`: PASS;
- gofmt, module diff, `git diff --check`, conflict-marker, and unexpected-file
  checks: PASS;
- race detector: unavailable because `CGO_ENABLED=1 go test -race` cannot find
  `gcc` in PATH; focused x200 and whole-package x100 stress are the recorded
  substitutes.

Remaining declared limitations are out of scope: external persistence/process
restart, production orchestration/wiring/terminal publication, and injectable
indeterminate storage-publication failure.

## Size Guard Reassessment

Decision: **`ACCEPT — DO NOT SPLIT`**.

Production scope remains one existing package and approximately `+238/-3`
tracked production lines, below the 500-line trigger. The final file count is
expected to exceed 15 only because one implemented behavior requires mirrored
DP-015/DP-016/DP-019 parity, mirrored indexes/roadmap, and mandatory project-
state synchronization. These documentation files cannot be removed without
leaving a live false claim that the prerequisite is unimplemented. No second
package, dependency, public production integration, architecture contract, or
independently deliverable behavior is present.

## PROCESS-002 Applicability Draft

- task record and task index: **Applicable** — activation, handoffs, findings,
  verification, closure, and navigation must be durable;
- `spec/current-state.md`: **Applicable** — isolated implementation and current
  task state changed;
- mirrored DP-015/DP-016/DP-019: **Applicable** — implementation boundary and
  TASK-026 relationship changed without status promotion;
- mirrored design indexes and MASTER_PLAN: **Applicable** — live summaries can
  no longer call the prerequisite unimplemented or Not Activated;
- `.ai/PROJECT_CONTEXT.md` and `spec/decisions.md`: **Applicable** — current
  implementation boundary and next recommendation changed;
- ARCH-004, DP-020/DP-021, root READMEs, docs root/roadmap indexes, and
  `spec/README.md`: **Not applicable** — architecture status, public capability,
  inventory, and navigation target did not change;
- `CHANGELOG.md`: **Not applicable** — no user-facing or release change.

Final PROCESS-002 status, Scope Audit, independent Reviewer verdict, and
Coordinator Acceptance remain pending.

## Final PROCESS-002 Audit

Documentation Agent verdict: **`Synchronized`**.

- exact semantic file set: `18` Required files, `0` unexpected content-diff
  files; `parent_store.go` has no content diff and is excluded from scope;
- mirrored target structure: DP-015 headings/fences `30/4` EN and `30/4` RU;
  DP-016 `30/4` and `30/4`; DP-019 `25/16` and `25/16`; design indexes `1/0`
  and `1/0`; MASTER_PLAN `36/0` and `36/0`;
- repository Markdown links: `958` relative links checked, broken `0`;
- ARCH-004 remains Active; DP-015, DP-016, and DP-019 remain Approved; DP-016
  remains Planned and DP-019 Planned overall;
- implemented wording is limited to the isolated TASK-047 DP-015 prerequisite;
  TASK-026 remains Blocked and the corrected matrix is explicitly historical
  pending a separate reassessment;
- `git diff --check` PASS; conflict markers `0`; no generated, dependency,
  module, root README, CHANGELOG, or production-wiring change exists.

Mandatory applicability is resolved exactly as drafted. Status:
**`Synchronized`**.

## Scope Audit

Decision: **`PASS` — 18 Required / 0 Questionable / 0 Removable**.

Required production and test files:

- `internal/runtimecommandidempotency/managed_parent.go` — implements the only
  new bounded admission and callback-scoped capability;
- `internal/runtimecommandidempotency/store.go` — discriminates primitive Stop
  versus derived phase ownership of the existing sole tracked-Start exception;
- `internal/runtimecommandidempotency/tracked_start_parent_test.go` — supplies
  DP-015 proofs 17–22 and regression evidence.

Required operational and project-state files:

- `docs/tasks/TASK-047-TRACKED-START-MANAGED-PARENT-ADMISSION.md` — Task
  Contract, role handoffs, evidence, PROCESS-002, Scope Audit, and closure;
- `docs/tasks/README.md` — active task navigation;
- `.ai/PROJECT_CONTEXT.md`, `spec/current-state.md`, and `spec/decisions.md` —
  live isolated implementation boundary, Blocked TASK-026 state, and next
  recommendation;
- `docs/en/design/DP-015-runtime-management-command-idempotency.md` and
  `docs/ru/design/DP-015-runtime-management-command-idempotency.md` — normative
  implementation boundary of the approved contract;
- `docs/en/design/DP-016-runtime-activation-replacement-rollback.md` and
  `docs/ru/design/DP-016-runtime-activation-replacement-rollback.md` — exact
  downstream traceability without activation;
- `docs/en/design/DP-019-runtime-activation-orchestration-prerequisites.md` and
  `docs/ru/design/DP-019-runtime-activation-orchestration-prerequisites.md` —
  parent/phase implementation-boundary traceability;
- `docs/en/design/README.md` and `docs/ru/design/README.md` — mirrored live
  design navigation summaries;
- `docs/en/roadmap/MASTER_PLAN.md` and `docs/ru/roadmap/MASTER_PLAN.md` — mirrored
  dependency-state summary.

None can be removed while retaining both the Definition of Done and mandatory
parity/project-state truth. No next-task implementation, refactoring,
production composition, generated artifact, dependency change, or unrelated
formatting content is present. Final Reviewer must explicitly validate this
disposition and the answer to the deletion question for every group.

## Independent Final Review

Reviewer verdict: **`APPROVED`**, blocking findings `0`, non-blocking findings
`0`.

The Reviewer confirms:

- architecture, ownership, and concurrency conform to Approved DP-015 section
  13.1, DP-016 proof 5, and DP-019;
- one generation plus per-Instance ledger linearization commits the complete
  parent/phase/occupant/rendezvous fact set before callbacks or waits;
- replay and observation receive no authority, the sole Stop exception remains
  Start-owned until original tracked Start completion/expiry, and different
  Instances remain independent;
- focused shuffled x100, full tests, vet, gofmt diff, module diff,
  `git diff --check`, and conflict-marker checks pass;
- the race limitation is reproduced exactly (`gcc` absent) and the recorded
  stress substitutes are proportionate;
- PROCESS-002, status/parity, Size Guard, and the complete `18/0/0` Scope Audit
  are sound.

Deletion question: every production/test file and every operational/project-
state/mirrored documentation group classified Required is necessary for the
Definition of Done. Removing any one would remove the implementation, a proof,
task traceability, mandatory factual state, or EN/RU parity. `parent_store.go`
has no byte/content diff and is not part of the semantic file set or future
exact commit set.

Unresolved risks are only the declared limitations: unavailable local race
detector; external persistence/process restart; production orchestration,
terminal publication, and wiring; injectable indeterminate storage publication.

### Closure Bookkeeping Rework

The first bounded repeat closure review returned `Needs Revision`, blocking
`1`, non-blocking `0`: B-001 found one live present-tense phrase in
`spec/current-state.md` saying the accepted TASK-047 contract “is being
implemented,” while all other closure sources already said Completed. The
bounded documentation rework changed it to the factual completed wording
“implemented in isolation in TASK-047.”

Repeat bounded closure review verdict: **`APPROVED`**, blocking `0`,
non-blocking `0`. B-001 is resolved; closure statuses, Blocked TASK-026, the
Not Activated/no-Task-ID reassessment, and absent commit/publication state are
consistent with no scope drift.

## Coordinator Closure and Acceptance

Coordinator Closure Audit: **`PASS`**.

- Task Contract and Definition of Done: satisfied;
- Architecture Confirmation: PASS, blocking `0`;
- Existing Coverage Report and Added Proof/Regression coverage: complete;
- independent Tester: PASS WITH ENVIRONMENT LIMITATION, `0/0` after resolved
  initial Stop-slot-lifetime finding;
- independent final Reviewer: APPROVED, `0/0`;
- Verification Matrix: PASS WITH the recorded race environment limitation;
- PROCESS-002: Synchronized;
- Size Guard: ACCEPT — DO NOT SPLIT;
- Scope Audit: `18 Required / 0 Questionable / 0 Removable`;
- staged changes: `0`; generated/unexpected content-diff files: `0`;
- TASK-026 remains Blocked; no orchestrator, terminal publication, production
  wiring, next-task implementation, commit, push, or publication occurred.

Coordinator Acceptance: **`Accepted`**.

Final status: **`Completed — Coordinator Accepted (2026-08-25)`**.

Commit readiness: the accepted exact semantic diff is ready for Commit Gate
only after the separate exact command `Разрешаю коммит.`. That command has not
been received; no stage or commit is performed.

Next recommendation: a separate bounded repository-first readiness
reassessment of TASK-026 against the complete DP-016 proof matrix after the
accepted TASK-047 prerequisite. It is **`Not Activated`**, has no Task ID, and
is not started by this closure.

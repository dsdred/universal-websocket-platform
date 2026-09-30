# TASK-071 — Runtime Owner Shutdown Provenance and Exact Attempt Snapshot

## Status

`Completed — Coordinator Accepted (2026-09-30)`. The bounded Design-update and
first DP-014 implementation slice passed independent Architecture Confirmation,
verification, three-round final Review, Scope Audit and Coordinator Acceptance.
No commit, publication or downstream activation is authorized or claimed.

## Task Contract

### Task Mode

`Implementation` with one mandatory pre-code Design-update gate in the same
task. The design stage amends only existing Approved DP-014, DP-016, DP-017,
DP-022 and DP-023 for the confirmed Owner shutdown-provenance defect. After
independent confirmation of the actual mirrored bytes, the implementation
stage may add the DP-014 provenance fact and a private coherent exact-attempt
snapshot/read with revision revalidation and adversarial tests.

This is not a readiness task and creates no new DP.

### Why Now

- TASK-070 is Coordinator-Accepted and published in synchronized
  `main@c058da69f2296e52a8e32cc25e195190389dbca7`; its task commit
  `82b7cce29ea9bca350b7945ac51a2531135d1a21` is the head of merged PR #75;
- a repository-first adversarial reconciliation of DP-014, DP-016, DP-017,
  DP-022, DP-023 and TASK-066..070 returned
  `CONFIRMED — EXISTING DESIGN DEFECT`: Approved DP-022 requires an
  Owner-produced shutdown-completion fact while the current DP-014 model stores
  the same terminal phase for Owner, no-Host, and future recovery projections;
- the same reconciliation disproved coherent-read mechanics and the
  composition/persistence boundary as new blockers. They remain governed by
  DP-014 coherent reads and DP-022 read-only composition;
- the user explicitly authorized a bounded amendment of the five existing DPs
  and direct transition to this first implementation slice after independent
  Architecture Confirmation.

### Definition of Done

1. Mirrored DP-014, DP-016, DP-017, DP-022 and DP-023 define one immutable
   terminal completion basis with distinct `OwnerShutdownCompleted`,
   Owner-confirmed `NoHostProduced`, and `RecoveryReconciled` meanings.
2. Only an exact successful Runtime Lifecycle Owner shutdown path may publish
   `OwnerShutdownCompleted`; a no-Host Owner outcome cannot be upgraded to it,
   and recovery/reconciliation can publish only `RecoveryReconciled`.
3. DP-022 positive `HostShutdownCompleted` evidence requires the exact
   `OwnerShutdownCompleted` basis for the same attempt/generation,
   `GenerationTerminated`, and fresh aggregate-revision plus generation-
   authority revalidation. DP-015 outcome is downstream/corroborating command
   truth, not a mandatory second evidence producer.
4. DP-023 places this provenance/read prerequisite before full shutdown-
   completion composition without treating coherent-read mechanism or
   composition/persistence ownership as new architecture blockers.
5. Independent Architect reviews the actual EN/RU amended bytes and returns
   `APPROVED` with zero blocking findings and no new prerequisite before any
   production or test edit.
6. `internal/runtimeidentity` stores immutable terminal completion basis,
   exposes it through detached attempt facts, and rejects invalid or attempted
   provenance upgrades without mutation.
7. A private repository-internal exact-attempt snapshot helper composes the
   existing DP-014 coherent reads, returns exact Workspace, Configuration,
   Runtime Instance, Launch Attempt, ConfigurationVersion, execution
   generation, phase, completion basis and aggregate revision, and fails
   closed on missing, duplicate, foreign or revision-incoherent facts.
8. Fresh revision revalidation rejects concurrent aggregate change. Tests
   cover Owner shutdown, no-Host terminalization, recovery projection, loss
   between DP-014 and DP-015 terminal publications, and concurrent revision
   change.
9. Focused/full tests, stress, vet, formatting, race or explicit environment
   limitation, documentation parity/links, PROCESS-002, Scope Audit and
   independent final Review pass before Coordinator Acceptance.

### Out of Scope

- full DP-022 `HostShutdownCompleted` composer or use-once evidence handle;
- production durable DP-014/DP-015 adapter, schema, migration or provisioning;
- containment provider/admission wiring or Control Service integration;
- DP-017 recovery assessment, claim, barrier, terminal publications or
  reconciliation implementation;
- DP-018 reporting, public API/DTO, Production Activation, child/remote
  topology, adoption or supervision;
- a new DP, a separate readiness task, or reopening coherent-read and
  composition/persistence ownership decisions already settled by Approved
  contracts.

### Verification Plan

- before code/test mutation: EN/RU structure, source precedence and focused
  semantic checks for the five amended DPs; independent Architecture
  Confirmation of their actual raw bytes;
- Existing Coverage Report below is the test-mutation gate;
- focused `go test ./internal/runtimeidentity -count=1` and stress;
- affected `internal/runtimeactivation` and continuation tests;
- full `go test ./... -count=1`, `go vet ./...`, `gofmt -l`,
  `go test -race` when supported or `PASS WITH LIMITATION`, and
  `go mod tidy -diff` without accepting unrelated line-ending changes;
- adversarial fake-reader tests for duplicate/foreign/missing/incoherent
  snapshot facts and stale revalidation;
- EN/RU semantic parity, relative-link validation, status/source-precedence
  assertions, `git diff --check`, independent final Review and full scope
  audit.

## Objective

Repair the existing Approved-design provenance defect and implement the first
bounded DP-014 prerequisite that can later support truthful
`HostShutdownCompleted` composition without implementing that composition.

## Selection Evidence

- exact user direction selects the confirmed defect and forbids a separate
  readiness task or new DP;
- TASK-070 is present in local `main` ancestry and `main == origin/main ==
  c058da69f2296e52a8e32cc25e195190389dbca7` at intake;
- TASK-070 newest matching envelope proves Coordinator Acceptance, while its
  publication is reconstructed from the merge commit and clean synchronized
  baseline;
- DP-023 section 19 orders shutdown-completion composition after the accepted
  generation reader, but fresh prerequisite inspection proves that immutable
  Owner provenance must be added before a truthful positive composition;
- rejected alternatives: a new DP/readiness task is unnecessary; implementing
  the full composer would cross the confirmed bounded slice; treating DP-015
  terminal outcome as mandatory evidence would preserve the Approved ambiguity
  at the DP-014/DP-015 crash cut.

## Scope

- mirrored focused normative amendments to DP-014, DP-016, DP-017, DP-022 and
  DP-023 only for Owner shutdown provenance and its dependency order;
- `internal/runtimeidentity` terminal provenance fact, validation, coherent
  exact-attempt snapshot helper and revision revalidation;
- minimal `internal/runtimeactivation` call-site adaptation needed to publish
  the correct Owner basis; no orchestration behavior expansion;
- focused proof/regression tests in affected existing packages;
- task record plus mandatory project-state/navigation synchronization,
  including factual TASK-070 publication and TASK-071 active state.

## Non-Goals

- any positive full-tuple shutdown evidence result;
- new storage technology, package dependency direction or external schema;
- recovery-owned terminal mutation implementation;
- refactoring unrelated identity, activation, lifecycle or containment code.

## Sources of Truth

- ADR-0003; Active ARCH-002, ARCH-004 and ARCH-005;
- Approved DP-014, DP-015, DP-016, DP-017, DP-022 and DP-023;
- TASK-066..070 accepted records and immutable Git ancestry;
- `internal/runtimeidentity`, `internal/runtimeactivation`,
  `internal/runtimeorchestrationcontinuation` and their tests.

## Roles

- Coordinator: intake, task boundary, gates, Size Guard, Scope Audit and
  Acceptance;
- Architect: initial amendment specification and independent confirmation of
  actual mirrored bytes before code;
- Documentation Agent: mirrored amendment and PROCESS-002 synchronization;
- Developer: implement only the independently confirmed DP-014 slice and
  minimal activation call-site mapping;
- Tester: Existing Coverage Report, adversarial tests and verification;
- Reviewer: independent final review; implementation/documentation author may
  not provide the final verdict;
- Publisher: not applicable without later explicit commit/publication gates.

## Branch and Recovery Anchor

- repository: `E:/wikiPRJ/universal-websocket-platform`;
- trusted baseline: `c058da69f2296e52a8e32cc25e195190389dbca7`;
- branch: `feature/task-071-owner-shutdown-provenance`;
- branch was created from clean synchronized `main` before content mutation;
  this task record is the first content change;
- permitted operations: in-scope documentation, production and test edits
  after their respective gates, plus read-only verification;
- forbidden without separate exact user command: stage, commit, push, PR,
  merge, fetch/pull, branch deletion, rebase, reset or remote mutation.

## Constraints

- terminal provenance is immutable and monotonic; no terminal or recovery path
  can upgrade one basis into another;
- `OwnerShutdownCompleted` is not inferred from `Stopped`, resource absence,
  process termination, DP-015 success, time, PID, port or probe;
- DP-015 remains command truth and follows lifecycle publication; losing its
  publication cannot erase an already committed Owner completion fact;
- snapshot read remains observation-only and creates no durable composition
  record; current process-local storage may support isolated proofs but does
  not claim restart durability;
- implementation stops on any new architecture prerequisite.

## Stop Conditions

- independent Architecture Confirmation finds a new prerequisite or cannot
  approve exact mirrored bytes;
- correct provenance requires a new store, schema, DP, public API or full
  composer inside this slice;
- implementation cannot distinguish Owner/no-Host/recovery without weakening
  Approved ownership;
- unexpected dirty/unowned change, mandatory check failure or blocking final
  review finding.

## Acceptance Criteria

1. Two terminal attempts with the same phase but different provenance remain
   observably distinct and immutable.
2. Only the Owner Stop-success call site publishes
   `OwnerShutdownCompleted`; Owner pre-Host terminal outcomes publish
   `NoHostProduced`.
3. A recovery publication surface cannot publish or upgrade to
   `OwnerShutdownCompleted`.
4. Exact-attempt snapshot facts correspond to one revision and stale use is
   rejected by fresh revision revalidation.
5. A committed DP-014 Owner completion remains identifiable when the following
   DP-015 command-outcome publication is absent or fails.
6. Documentation preserves the absent full composer, durable production
   adapter, recovery implementation and Production Activation boundaries.

## Existing Coverage Report

- **Existing Coverage:** `internal/runtimeidentity/store_test.go` covers
  atomic attempt claim, generation binding, Running/Stop/terminal publication,
  stale revision, append-only history, coherent aggregate reads, concurrent
  same-Instance mutation and different-Instance progress. Activation tests
  cover Owner Start/Stop outcomes and current DP-014-before-DP-015 publication
  ordering. Continuation tests already exercise the aggregate/history/
  aggregate revision sandwich.
- **Coverage Gap:** no stored terminal producer/basis exists; `Stopped` merges
  Owner shutdown with stopped-before-running. No recovery-only publication
  category or attempted-upgrade proof exists. No reusable exact-attempt
  snapshot helper exposes provenance or proves stale revision revalidation.
  No focused regression proves that a lost DP-015 publication leaves the
  DP-014 Owner basis intact.
- **Added Proof Tests planned:** basis validation/immutability, exact Owner
  shutdown and no-Host mapping, recovery projection and no-upgrade, coherent
  exact snapshot, missing/duplicate/foreign/incoherent input, stale revision
  revalidation and concurrent revision change.
- **Added Regression Tests planned:** DP-014 terminal commit followed by
  absent/failing DP-015 terminal publication; existing start/replace/rollback
  behavior and all repository tests.
- **Remaining Limitations:** external durability and restart persistence are
  intentionally absent; full composer/use-once handle and DP-017 recovery are
  later slices. Race-tool availability depends on the current Windows Go/CGO
  environment and must be reported truthfully.

## Size Guard

Provisional verdict: `ACCEPT — BOUNDED CROSS-DOCUMENT CONTRACT REPAIR`.
Five mirrored Approved DPs change because the confirmed defect crosses their
existing ownership chain, but only one semantic decision and one independently
shippable implementation behavior are added. Expected production change is
well below 500 lines, no new package or dependency is planned, and full
composition remains split out. The >15-file trigger, if reached solely through
required mirrors, task/project-state sync and focused tests, requires explicit
final re-evaluation rather than an automatic split.

## Ordered Stages

1. Task Intake and Documentation Baseline.
2. Architect specification of the bounded amendment.
3. Documentation Agent writes mirrored DP bytes.
4. Independent Architecture Confirmation of actual bytes.
5. Developer implementation and focused tests.
6. Verification and rework.
7. PROCESS-002 and project-state synchronization.
8. Scope Audit, final checks and independent final Review.
9. Coordinator Acceptance and next-task recommendation.

## Documentation Baseline

- DP-014, DP-016, DP-017, DP-022 and DP-023 are Approved; their EN/RU mirrors
  exist and statuses are aligned;
- DP-014 defines coherent aggregate reads and Owner-only lifecycle
  publications but stores no producer/completion basis;
- DP-016 orders Owner terminal publication before DP-015 command/phase
  outcome; DP-017 permits recovery terminal projections while forbidding
  inference of successful Stop from resource absence;
- DP-022 requires Owner-produced shutdown-completion evidence and rejects
  recovery-produced facts, but does not define a durable discriminator or the
  normative role of DP-015 at the crash cut;
- DP-023 orders the full shutdown-completion composition after TASK-070's exact
  generation reader and requires fresh prerequisite checks;
- repository code confirms the defect: `LaunchAttemptRecord` stores only phase
  and generation, and `ConditionalPublishTerminal` uses a Boolean terminal
  distinction. No critical higher-precedence conflict was found;
- navigation/project-state files still project TASK-070 as `In Progress` even
  though its newest matching envelope and Git ancestry prove Acceptance and
  publication. This stable post-publication drift is required synchronization
  inside TASK-071, not a new architecture prerequisite.

## Architecture Confirmation

Initial Architect specification: `SPECIFICATION COMPLETE — NO NEW
PREREQUISITE`. It authorizes only the bounded mirrored documentation update and
requires one closed `TerminalCompletionBasis` domain with
`OwnerShutdownCompleted`, `NoHostProduced`, and `RecoveryReconciled`; immutable
absent-to-one publication; Owner-only Owner bases; recovery-only
`RecoveryReconciled`; exact same-generation shutdown evidence; and private
coherent snapshot/revalidation mechanics. It also confirms that DP-015 is
downstream/corroborating command truth, not a mandatory second evidence source.

Independent Architecture Confirmation of the actual amended EN/RU bytes:
`APPROVED — NO NEW PREREQUISITE`. The independent Architect reviewed the full
working-tree diff, reported zero blocking findings, confirmed semantic mirror
parity, unchanged statuses and implementation boundary, and opened Developer
handoff only for the bounded DP-014 provenance/snapshot slice.

## Verification

- focused `go test ./internal/runtimeidentity ./internal/runtimeactivation
  ./internal/runtimeorchestrationcontinuation`: PASS;
- full `go test ./... -count=1`: PASS;
- affected `runtimeactivation`, `runtimeidentity`, `runtimelifecycle` and
  `runtimeorchestrationcontinuation` tests with `-count=25`: PASS;
- full `go vet ./...`: PASS;
- `gofmt -l` for all changed Go files: clean;
- `git diff --check`: PASS;
- `go mod tidy -diff`: no semantic dependency change; output is only the
  baseline CRLF/LF representation difference in `go.sum`, which was not
  accepted as a repository mutation;
- race detector: unavailable in the current Windows toolchain. The default run
  reports that `-race` requires CGO; explicit `CGO_ENABLED=1` reports missing C
  compiler `gcc`. This is an environment limitation, not a test failure.
- independent final Review round 1: `NEEDS REVISION`, three blockers fixed;
- repeat round 2: `NEEDS REVISION`, one Starting revision-order blocker fixed;
- repeat round 3: `APPROVED`, blocking findings 0, exact reviewed manifest
  SHA-256 `f1a0d267b2bc352ba0c4e8ded94bd2440966bf5b7dafb8914148652a3d219f86`.

## Scope Audit

Final audit: `21 Required / 0 Questionable / 0 Removable` paths.

- 10 required EN/RU DP mirrors;
- 1 required task record and 2 required project-state files;
- 4 bounded production files (`runtimeidentity` plus minimal
  `runtimeactivation` mapping);
- 4 focused regression/adversarial test files.

The >15-file Size Guard trigger is accepted on re-evaluation: mirror and
governance bytes account for 13 files; production additions are 365 lines,
below the 500-line threshold; there is one semantic behavior and no second
package, dependency, composer, adapter, recovery workflow, or wiring surface.

## Documentation Synchronization

`Synchronized` for the accepted implementation: DP-014/016/017/022/023 EN/RU
implementation boundaries, `spec/current-state.md`, and
`.ai/PROJECT_CONTEXT.md` reflect TASK-070 publication, TASK-071 provenance/read
implementation, exact remaining deferrals, and unchanged Design/Implementation
statuses. Historical task records were not rewritten.

## Handoff

Developer, Tester, Reviewer and Coordinator gates are complete. Publisher
handoff for publication is not open. The user has explicitly authorized one
task commit; the Commit Gate is in progress. Push, PR, merge and publication
remain unauthorized.

## Next Candidate

After this task is accepted and separately published, the next candidate is the
private full `HostShutdownCompleted` evidence composer with its invocation-
scoped use-once handle. It is `Not Activated`; external durable adapter,
provider/admission wiring, DP-017 recovery and Production Activation remain
later prerequisites.

## Closure

`Completed — Coordinator Accepted (2026-09-30)`. Architecture Confirmation,
Developer implementation, verification, PROCESS-002 synchronization, Scope
Audit and independent final Review are complete. Commit and publication do not
exist and were not authorized. The full composer and every later prerequisite
remain `Not Activated`.

## Recovery Evidence Envelope

### 2026-09-29 — Intake

- branch/status/history preflight: clean synchronized
  `main == origin/main == c058da69f2296e52a8e32cc25e195190389dbca7`;
  TASK-070 task commit is in ancestry and no local/remote TASK-070 task branch
  is required for this intake;
- branch `feature/task-071-owner-shutdown-provenance` was created from that
  exact baseline before content mutation; this record is the first content
  change;
- exact current user direction authorizes one integrated Implementation task
  with a bounded pre-code amendment of existing DPs, forbids a separate
  readiness task/new DP, and requires STOP on a new prerequisite;
- Task Contract, Documentation Baseline and Existing Coverage Report are
  recorded before architecture, production or test mutation;
- first incomplete checkpoint: mirrored Design-update, followed by independent
  Architecture Confirmation of actual bytes;
- no stage, commit, push, PR, merge or publication permission exists.

### 2026-09-29 — Bounded Design-update authored

- the initial Architect returned `SPECIFICATION COMPLETE — NO NEW
  PREREQUISITE` and authorized mirrored documentation authoring only;
- DP-014, DP-016, DP-017, DP-022 and DP-023 EN/RU mirrors were amended only for
  immutable Owner shutdown provenance, exact-attempt observation/revalidation,
  DP-015 corroboration, recovery non-upgrade, and dependency order;
- no new DP, readiness task, store, record type, public contract, recovery
  implementation, full evidence composer, provider/admission wiring or
  Production Activation was introduced;
- first incomplete checkpoint: independent Architecture Confirmation of the
  actual amended bytes before production or test mutation.

### 2026-09-30 — Independent Architecture Confirmation

- a different agent from the specification author reviewed the actual ten DP
  mirror bytes, TASK-071 and the complete working-tree diff without editing;
- verdict: `APPROVED — NO NEW PREREQUISITE`, zero blocking findings;
- confirmed: closed immutable three-value basis; Owner/recovery authority
  separation; exact same-generation DP-022 predicate and fresh revalidation;
  DP-015 corroborating role; DP-023 dependency order; existing coherent-read
  and persistence boundaries; EN/RU parity and unchanged statuses;
- Developer handoff is open only for the bounded runtimeidentity provenance,
  private exact-attempt snapshot/revalidation, minimal runtimeactivation mapping
  and focused adversarial tests;
- first incomplete checkpoint: implementation and focused verification.

### 2026-09-30 — Implementation and verification candidate

- `internal/runtimeidentity` now stores one immutable terminal completion basis
  atomically with terminal phase and exposes disjoint narrow Owner and recovery
  terminal-publication capabilities; no caller-selected provenance enum is
  exposed;
- `ConditionalClaimStop` never terminalizes an attempt or selects provenance:
  even a claim from pre-Host `Claimed` moves to `Stopping`, retains the active
  attempt and awaits an exact Owner `NoHostProduced` outcome;
- a private repository-internal exact-attempt snapshot composes
  aggregate/history/aggregate reads, validates exact identity/binding/phase/
  basis coherence, and revalidates exact aggregate revision before use;
- minimal `runtimeactivation` mapping uses exact Owner terminal kinds, while a
  fault-injection test proves committed DP-014 Owner basis survives loss before
  DP-015 terminal publication;
- adversarial tests cover same-phase Owner/no-Host distinction, recovery
  projection and prohibited upgrade, malformed/foreign/duplicate/partial
  snapshots, concurrent revision change, and existing orchestration regression;
- focused and full tests, focused vet and diff check pass; race is unavailable
  because the environment has no C compiler;
- PROCESS-002 candidate synchronization and Size Guard re-evaluation are
  recorded; first incomplete checkpoint: independent final Review.

### 2026-09-30 — Independent Review round 1 and bounded rework

- independent Reviewer verdict: `NEEDS REVISION`; Scope Audit remained
  `21 required / 0 justified / 0 unrelated` and the Size Guard remained
  accepted;
- blocking finding 1: the first candidate let a stop claim from `Claimed`
  create `Stopped/NoHostProduced` without an exact Owner outcome;
- blocking finding 2: Owner and recovery terminal methods were both exported
  directly on `Store`, so possession of the same store did not enforce
  authority separation;
- blocking finding 3: the final DP-022 Decision still described DP-014 and
  DP-015 jointly as durable shutdown-completion evidence;
- bounded rework makes every Stop claim non-terminal, requires the subsequent
  exact Owner result, gives activation only the narrow Owner publication
  capability, gives recovery a disjoint method set, and corrects the mirrored
  DP-022 Decision to make DP-014 authoritative while DP-015 remains optional
  corroborating command truth;
- focused tests pass after rework; the first incomplete checkpoint is full
  verification followed by repeat independent Review of the latest bytes.

### 2026-09-30 — Independent Review round 2 and bounded rework

- repeat Reviewer confirmed all three round-1 blockers fixed, but returned
  `NEEDS REVISION` for one Starting-path regression;
- the intended `facts.running || facts.starting` Stop-claim condition had been
  applied to an unreachable parent-candidate branch rather than `stopOld`, so
  an Owner that created and stopped a Host before Running publication returned
  `AttemptStopped` or `AttemptStopFailed` while DP-014 remained `Claimed`;
- bounded correction restores the parent revision calculation and makes
  `stopOld` publish the nonterminal DP-014 Stop claim for both Running and
  Starting before invoking Owner;
- the tracked-Starting no-Host integration remains `NoHostProduced`; a focused
  regression proves an unavailable exact Owner outcome after the Starting
  claim leaves active `Stopping` with no basis. Existing real-Owner lifecycle
  tests cover Host-created-before-Running Stop success/failure, while DP-014
  capability tests cover their `OwnerShutdownCompleted` and provenance-free
  stop-failure publications;
- no new prerequisite or scope path was introduced; first incomplete
  checkpoint: verification and another repeat independent Review.

### 2026-09-30 — Final Review and Coordinator Acceptance

- latest verification: full tests PASS; four affected packages at `-count=25`
  PASS; full vet PASS; formatting and diff check clean; race remains unavailable
  because the Windows toolchain has no C compiler; tidy diff is line endings
  only and no dependency bytes were changed;
- independent final Review round 3: `APPROVED`, blocking findings 0, no new
  prerequisite, Scope Audit `21 Required / 0 Questionable / 0 Removable`;
- accepted reviewed manifest SHA-256:
  `f1a0d267b2bc352ba0c4e8ded94bd2440966bf5b7dafb8914148652a3d219f86`;
- Coordinator accepts the bounded Design-update and isolated DP-014 Owner
  provenance/exact-attempt snapshot implementation. Full
  `HostShutdownCompleted` composition, durable adapter, provider/admission
  wiring, DP-017 recovery and Production Activation remain `Not Activated`;
- no stage, commit, push, PR, merge or publication operation was performed.

### 2026-09-30 — Commit authorization

- the user sent the exact repository Commit Entry command, `Разрешаю коммит.`;
- this authorizes one accepted TASK-071 commit after the PROCESS-001 Commit
  Gate verifies the exact path set, accepted subject plus expected closure
  synchronization, absence of unrelated changes, and applicable checks;
- it does not authorize push, PR, merge, or publication.

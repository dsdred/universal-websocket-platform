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

### 2026-09-30 — R001 bounded conflict rework intake

- Current explicit user authority confirms this existing TASK-071 container;
  no new task or design decision is created. Content rework is limited to
  `.ai/PROJECT_CONTEXT.md` and `spec/current-state.md`. This task record is
  the sole additional evidence path and may only receive append-only records.
- Historical accepted source: branch
  `feature/task-071-owner-shutdown-provenance`, HEAD
  `b653337dc93eb9419648246e447813dc168a16ea`, fixed publication base
  `c058da69f2296e52a8e32cc25e195190389dbca7`, historical accepted scope SHA-256
  `f1a0d267b2bc352ba0c4e8ded94bd2440966bf5b7dafb8914148652a3d219f86`.
  Current integration source is immutable
  `main@fb6341e48f6abdea0f2c2b23ea4c949a16ed3a70`, containing published
  TASK-072 commit `60ee9bec4dde3df08620f1a9d54b28da3f5b4c62` / PR #76.
- Governing contracts are the published PROCESS-001/002 at that main OID,
  not the older copies in the historical TASK-071 checkout. Their raw bytes
  are inspected through Git objects; this rework does not change contracts.
- Original task record prefix: 26,988 bytes, raw Git blob
  `11f27e8a26a4f6cbdc68508aa268b5e95c1c0557`, raw SHA-256
  `7c98a6ebb20a6f919fe76998e0f31a664340b2ceeaaeb5b18ec387fa687ecc062`.
  All original records, including the historical Acceptance, remain unchanged.
- Coordinator assigns the primary agent sequential Documentation and Tester
  roles; independent Reviewer is `process_proposal_review`, not the author.
  Ordered stages: reconstruct both sources; documentation reconciliation;
  fresh Verification / PROCESS-002; Scope Audit / Size Guard; independent
  actual-byte Review; new exact-subject Coordinator Acceptance; STOP.
- Existing coverage: no code or tests change; TASK-071 already contains Owner,
  no-host, recovery, publication-gap and concurrent-revision adversarial tests.
  Fresh verification plan: full tests, four affected packages at `-count=25`,
  full vet, formatting, diff checks, raw-prefix and protected-path equality,
  canonical manifest reproduction and cross-source documentation checks.
  Race remains conditional on the available C toolchain; no dependency change.
- No stage, commit, branch creation, merge, rebase, amend, push, PR update or
  deletion is permitted in this rework. PR #77 and both immutable source
  commits are untouched. Historical publication authorization and transfer
  `87ea13f0-4607-47e4-9669-58bc88714b30` apply only to their unchanged Target,
  not to a prospective successor. Next implementation remains Not Activated.
- Rework subject is the three authorized paths; task-record-v1 excludes this
  terminal envelope. Every other path must equal historical HEAD. A later
  integration commit must additionally preserve the unchanged published
  TASK-072 files from current main; this preparation creates no commit/tree.
- First incomplete checkpoint: documentation reconciliation. Any unexpected
  path, historical-prefix change, unresolved source contradiction or failed
  mandatory gate requires STOP before Acceptance.

### 2026-09-30 — R002 actual reconciliation and exact subject

- Inspect -> Reconstruct -> Reconcile read both source deltas against common
  ancestor `c058da69f2296e52a8e32cc25e195190389dbca7`. TASK-071 contributes
  accepted provenance/snapshot capability; current main contributes published
  TASK-072 governance. Neither whole-side selection preserves both facts.
- Both project-state files now distinguish current bounded rework, historical
  accepted TASK-071 source and published TASK-072 process repair. TASK-070
  publication and TASK-069 bootstrap history are retained. Stale commit-only
  instructions are removed, not replaced by transient Publisher state. All
  lower TASK-071 capability sections remain byte-for-byte unchanged.
- This is a prospective metadata reconciliation, not retroactive Acceptance
  of historical implementation and not a TASK-071 merge/publication claim.
- Reproducible exact diffs: `git diff
  b653337dc93eb9419648246e447813dc168a16ea -- .ai/PROJECT_CONTEXT.md
  spec/current-state.md docs/tasks/TASK-071-RUNTIME-OWNER-SHUTDOWN-PROVENANCE.md`
  and the same command with base
  `fb6341e48f6abdea0f2c2b23ea4c949a16ed3a70`. Against current main, lower
  implementation paragraphs are inherited TASK-071 changes, not new rework.
- Canonical subject per published PROCESS-001: object format `sha1`, anchor
  HEAD `b653337dc93eb9419648246e447813dc168a16ea`, historical base
  `c058da69f2296e52a8e32cc25e195190389dbca7`, integration source
  `fb6341e48f6abdea0f2c2b23ea4c949a16ed3a70`.
  Ordered rows (unsigned UTF-8 path bytes; every mode `100644`):

| Path | Projection | State | Mode | Projected blob OID |
| --- | --- | --- | --- | --- |
| `.ai/PROJECT_CONTEXT.md` | full | present | 100644 | `230430ee783ac7d10421c9f1a0eec48e6a2d2acc` |
| `docs/tasks/TASK-071-RUNTIME-OWNER-SHUTDOWN-PROVENANCE.md` | task-record-v1 | present | 100644 | `37d9625eb8e61429be7a27d98a0ca9a7416aefb5` |
| `spec/current-state.md` | full | present | 100644 | `ba032e270b21a4cb664fde3e6ec22e3308a084d4` |

- Manifest OID: `c4e1d09ea8bddf047eea7781c82a6bbcfc51995c`.
  Hash each full file using `git hash-object --no-filters`; project task record
  exactly as PROCESS-001 task-record-v1, without newline normalization.
  Concatenate each row's five fields and trailing NUL, then hash raw stream
  with `git hash-object --stdin` without `-w`. The terminal envelope is excluded;
  these entries do not self-attest their own final raw bytes.

### 2026-09-30 — R003 fresh Verification and PROCESS-002

- Tested subject is exactly R002 manifest
  `c4e1d09ea8bddf047eea7781c82a6bbcfc51995c`, not historical scope SHA-256.
- `go test ./... -count=1`: completed exit 0, PASS (all packages; no cached
  test result substituted for this fresh regression run).
- `go test ./internal/runtimeidentity ./internal/runtimeactivation
  ./internal/runtimelifecycle ./internal/runtimeorchestrationcontinuation
  -count=25`: completed exit 0, all four packages PASS. Additional containment
  stress at `-count=25` completed exit 0; it is supporting, not a new slice.
- `go vet ./...`: completed exit 0, PASS. `gofmt -l` on the eight historical
  TASK-071 Go paths: empty, exit 0. No Go path changes in this rework.
- Broad exploratory `gofmt -l` reports 45 pre-existing paths outside the
  rework/source-code scope; the initial broad formatting assertion failed.
  No formatting rewrite performed. This is not claimed as a full-format PASS.
- Race: default `go test -race ./internal/runtimeidentity` is unavailable
  with `CGO_ENABLED=0`; explicit `CGO_ENABLED=1` fails because `gcc` is missing.
  Verdict `PASS WITH LIMITATION` for the race row; fresh regression, stress and
  vet are substitutes, not proof of race-detector execution.
- `git diff --check`: completed exit 0. Exact changed set is three authorized
  paths; index empty, HEAD/source refs unchanged. Original 26,988-byte task
  prefix equals original Git blob byte-for-byte. Lower project-state sections
  beginning at TASK-068 equal historical HEAD byte-for-byte.
- Protected source inventory: all 18 other historical TASK-071 paths remain
  unchanged by rework. Four DP checkout files already have CRLF-only differences
  from Git blobs; 14 paths, including all eight Go files, equal raw Git blobs.
  CRLF normalization was used only to diagnose those existing checkout
  representations, never to attest subject identity or authorize a commit.
- PROCESS-002 applicability: `.ai` and current-state Required/Synchronized;
  this task record Required/append-only evidence. Architecture/code/tests,
  dependency/module files, external API, EN/RU DP mirrors, process contracts,
  README/task index and CHANGELOG: no new change applicable. Published process
  contracts are read at fixed main OID; no copied stale contract claims.
  The restored relative DP-023 link resolves. No new public mirrored content.
- No test/dependency bytes, generated artifacts, credentials or operational
  handoff records added. Verification does not claim integration mergeability
  or production activation; actual integration has not been executed.

### 2026-09-30 — R004 Scope Audit and Size Guard

- Exact audit: `3 Required / 0 Questionable / 0 Removable`.
  `.ai/PROJECT_CONTEXT.md` and `spec/current-state.md` are independently
  required conflict/project-state consumers; removing either loses one side's
  accepted facts. The append-only task record is required durable evidence for
  fresh gates; removing it loses reconstructable Review/Acceptance.
- Rework is a single documentation reconciliation: no implementation, new
  design contract, task, dependency, API or activation. Existing original
  21-path implementation scope is not silently broadened or reaccepted.
- Size Guard re-evaluation: three changed paths, two small content patches
  confined to the state headings plus the authorized evidence envelope.
  No >15-file trigger; no decomposition is needed for this rework.
- First incomplete checkpoint: independent actual-byte Review of R002 subject.
- Proposed later integration path, not performed or authorized here: preserve
  both immutable commits with a new integration merge, intended parents
  `[b653337dc93eb9419648246e447813dc168a16ea,
  fb6341e48f6abdea0f2c2b23ea4c949a16ed3a70]`, incorporating these resolutions.
  Published PROCESS-001 Commit Gate explicitly does not authorize merge.
  Separate explicit integration authority and full composite-tree integrity /
  applicable Review / Acceptance are required before a separately authorized
  commit. Three-path Acceptance does not attest the extra imported main paths.
- Future staged bytes must match the exact accepted raw subject; current
  `core.autocrlf=true` is not permission for LF/CRLF equivalence. Proposed new
  head OID is undefined until the separately gated commit exists. No ordinary
  metadata successor is asserted to resolve PR #77 merge conflicts by itself.

### 2026-09-30 — R005 navigation finding rework / final tested subject

- Independent Reviewer identified one low navigation ambiguity before final
  verdict: `.ai` called TASK-070 the latest completed task below newer facts.
  Bounded fix relabels it `Previous published product task`. No other content
  path or capability changes. R002's manifest and R003's subject binding are
  superseded for this candidate, not rewritten or reused as final Approval.
- Final exact subject, object format `sha1`, anchor HEAD
  `b653337dc93eb9419648246e447813dc168a16ea`, historical base
  `c058da69f2296e52a8e32cc25e195190389dbca7`, integration source
  `fb6341e48f6abdea0f2c2b23ea4c949a16ed3a70`:

| Path | Projection | State | Mode | Projected blob OID |
| --- | --- | --- | --- | --- |
| `.ai/PROJECT_CONTEXT.md` | full | present | 100644 | `802263aa4e319b210358fe32e51376b88fb92d44` |
| `docs/tasks/TASK-071-RUNTIME-OWNER-SHUTDOWN-PROVENANCE.md` | task-record-v1 | present | 100644 | `37d9625eb8e61429be7a27d98a0ca9a7416aefb5` |
| `spec/current-state.md` | full | present | 100644 | `ba032e270b21a4cb664fde3e6ec22e3308a084d4` |

- Canonical manifest OID `86de55dfe803f42f815bc46fd4fbb899dbf3592e`, using
  the same raw-byte algorithm and path ordering as R002, without normalization.
- Fresh repeat `go test ./... -count=1`: completed exit 0, all packages PASS.
  Fresh repeat four-package `-count=25` command from R003: completed exit 0,
  all four PASS. Fresh repeat `go vet ./...`: completed exit 0, PASS.
  Race/environment and exploratory baseline-format limitations remain R003;
  this label rework changes no code, test, dependency or toolchain bytes.
- Repeat applicable documentation/scope checks completed: diff check exit 0;
  three-path set unchanged, no staged bytes; original task prefix identical;
  both lower document sections identical; protected 18-path inventory
  unchanged (14 raw-equal / 4 existing DP checkout EOL representations);
  restored DP-023 link resolves; final manifest independently reproducible.
  PROCESS-002 Synchronized; Scope Audit `3/0/0`; Size Guard unchanged.
- First incomplete checkpoint: final independent Review on this exact subject.

- Independent prefix readback also identified a transcription typo in R001's
  SHA-256 claim (65 hex characters). The authoritative 26,988-byte Git source
  and retained prefix both hash to the 64-character SHA-256
  `7c98a6ebb20a6f919fe76998e0f31a664340b2ceaaeb5b18ec387fa687ecc062`.
  This correction supersedes that R001 digest claim append-only; original
  source blob, historical records and prefix bytes are unchanged. It does not
  change the projected subject or attest final envelope bytes.

### 2026-09-30 — R006 independent actual-byte Review

- Independent Reviewer `process_proposal_review` completed actual-byte
  Architecture/Process and final Review: `APPROVED`, blocking findings 0,
  unresolved findings 0. Exact reviewed subject is R005's ordered three rows,
  canonical manifest `86de55dfe803f42f815bc46fd4fbb899dbf3592e`, format sha1,
  HEAD `b653337dc93eb9419648246e447813dc168a16ea`, historical base
  `c058da69f2296e52a8e32cc25e195190389dbca7`, integration source
  `fb6341e48f6abdea0f2c2b23ea4c949a16ed3a70`.
- Reviewer independently recomputed manifest from raw bytes after reading
  R005, inspected both complete project-state diffs against HEAD and main,
  and checked R001-R005 evidence. TASK-071 implementation and TASK-072
  published process facts coexist without false integration/activation claim.
- Independent preservation proof: original 26,988-byte task prefix equals
  original Git blob exactly; corrected 64-character SHA-256 matches R005.
  Historical contract/Acceptance/evidence retained. Protected inventory is
  14 raw-equal / 4 baseline CRLF-only DP checkout representations, with all
  eight Go source/test paths raw-equal. EOL diagnosis is not subject equivalence.
- Finding dispositions: competing TASK-070 latest label fixed in actual `.ai`
  bytes; R001 digest typo corrected append-only in R005. Neither remains open.
  Reviewer confirmed no conflict markers, diff check exit 0, restored link,
  unchanged refs, empty staged set, and exact authorized changed paths.
- Independent Scope Audit `3 Required / 0 Questionable / 0 Removable`:
  each state consumer and durable task evidence is necessary to the bounded
  DoD. Size Guard accepts one bounded documentation behavior. Fresh R005
  Tester commands/results reviewed; race and broad-format limitations retained.
- No new architecture prerequisite. Approval does not establish future merge
  readiness or accept imported main files. R004's explicit integration,
  composite-tree Review/Acceptance and separate Commit Gate remain required.
  No historical authorization or Accepted Handoff carries to a new Target.
- Next gate: bounded Coordinator Acceptance followed by append integrity.
  Reviewer performed no edits, Git mutation, auth/network probe or publication.

### 2026-09-30 — R007 new bounded Coordinator Acceptance / STOP

- Coordinator decision `ACCEPT`: prospective three-path project-state
  conflict reconciliation and append-only rework evidence, exactly R005 rows
  and manifest `86de55dfe803f42f815bc46fd4fbb899dbf3592e`, format sha1.
  Anchor HEAD `b653337dc93eb9419648246e447813dc168a16ea`, fixed historical
  base `c058da69f2296e52a8e32cc25e195190389dbca7`, current integration source
  `fb6341e48f6abdea0f2c2b23ea4c949a16ed3a70`. This is a new bounded
  Acceptance; original implementation Acceptance is neither rewritten nor
  reinterpreted. It does not accept a not-yet-created composite Git tree.
- Fresh Verification / PROCESS-002 Synchronized, Scope Audit `3/0/0`, Size
  Guard and independent exact-subject R006 `APPROVED` satisfy affected gates.
  No blocking findings or new architecture prerequisite. Runtime/composer,
  adapter, wiring, recovery and Production Activation scope remains unchanged.
- Bounded rework checkpoint `Completed — Coordinator Accepted`. Stable
  projected `In Progress` fields resolve through this newest matching envelope,
  not through a copied mutable closure claim. These metadata records do not
  self-hash final envelope bytes; final commit/tree does not yet exist.
- Proposed commit path remains R004's separately authorized integration merge
  preserving parents b653/main fb634 and unchanged source payloads. New head
  OID is undefined. Before that path: explicit local integration authority,
  exact complete composite-tree/parents/scope checks and affected fresh Review /
  Coordinator Acceptance, then separate exact `Разрешаю коммит.` Commit Gate.
- A subsequent immutable Target must include actual branch/head/ordered target,
  current fixed publication base and its accepted composite scope. It requires
  a new publication gate. PROCESS-001 factual Target change requires
  `InvalidatedByTargetChange / NoneTerminal` and related old attempts
  `Closed(TargetChanged)`; old Target authorization/Accept cannot be reused.
  That operational disposition belongs in the canonical Publisher store,
  not a fabricated Release or update to historical immutable task bytes.
- If trusted-context handoff is needed for the new Target, use its own valid
  source Release, actual user Route, destination reconstruction/fresh P0,
  durable Accept and readback; never reuse the old TASK-071 transfer. Until
  new gates are satisfied no publication mutations are permitted.
- STOP after post-decision integrity. No stage, commit, merge, rebase, amend,
  push, PR #77 update, deletion or publication was performed in this rework.

### 2026-09-30 — R008 post-decision integrity receipt

- Independent Reviewer reread actual R006/R007: `POST-DECISION INTEGRITY —
  PASS`, blocking findings 0. R006 faithfully records Review; R007 accepts
  only the three-path reconciliation/evidence, not a future integration tree.
- Independently recomputed after those appends: manifest
  `86de55dfe803f42f815bc46fd4fbb899dbf3592e` and every R005 row unchanged;
  original 26,988-byte prefix and corrected SHA-256 unchanged; no post-review
  projected-content mutation, only excluded evidence appends. Changed set
  remains the three authorized paths, index empty, diff check exit 0;
  HEAD b653 and main/origin-main fb634 unchanged.
- This receipt records that completed readback, not a digest of its own bytes.
  Final Coordinator recomputation checks this permitted envelope append without
  changing accepted subject or historical prefix. Acceptance STOP remains in
  effect; all later integration/commit/new-Target publication gates remain open.

### 2026-09-30 — I001 integration preparation intake

- Current explicit user authority permits only preparation and verification of
  the integration result. Intended ordered commit parents are
  `b653337dc93eb9419648246e447813dc168a16ea` and
  `fb6341e48f6abdea0f2c2b23ea4c949a16ed3a70`; historical common base remains
  `c058da69f2296e52a8e32cc25e195190389dbca7`. Accepted three-path rework
  manifest `86de55dfe803f42f815bc46fd4fbb899dbf3592e` must remain unchanged.
- Confirm existing TASK-071 container and original contract; do not create a
  task, alter historical Acceptance or add implementation. Primary agent takes
  sequential Documentation/Tester roles; independent composite Reviewer is
  `process_proposal_review`. Governing contracts are fixed published main.
- Preparation uses only Git blob/tree objects: current-main tree plus exact
  TASK-071-only delta, overriding only the three accepted rework paths. No
  commit object, real-index staging, merge command, ref movement or branch
  creation. Trees do not contain parents; the independently verified intended
  parent tuple is bound to the result and must be enforced at a later Commit Gate.
- Composite subject is the 33-path union of the two source deltas: 18 unchanged
  TASK-071 payload paths, 12 unchanged published-main payload paths and three
  accepted reconciliation/evidence paths. Only the current TASK-071 record
  uses task-record-v1; the immutable published TASK-072 record uses full bytes.
  Every other tree entry must equal the fixed current-main entry exactly.
- Existing coverage is unchanged; no test additions are needed. Verification
  plan: export exact tree without checkout filters to an isolated temporary
  snapshot; full fresh tests, four-package stress, full vet, source formatting,
  both-parent blob/mode preservation, no-extra-change proof, canonical manifest,
  relative links/mirror/scenario composition and whitespace checks. Record
  unavailable race separately. Repeat affected gates on actual composite bytes.
- Ordered gates: build/prove tree; fresh Verification/PROCESS-002; resulting
  Scope Audit/Size Guard; fresh independent composite Review; new composite
  Coordinator Acceptance/post-decision integrity; STOP before Commit Gate.
- No final commit, amend, rebase, force-push, PR #77 update, push, external merge,
  deletion, auth probe or publication. Old authorization/transfer are not
  inherited. Any source, accepted rework, unrelated-path or evidence mismatch
  requires STOP, not silent normalization or whole-side conflict selection.

### 2026-09-30 — I002 composite construction / exact subject

- Independently readable source commits remain b653/main fb634 from I001;
  source trees are respectively `fd04de9402a2aad826f37e1d681d6d7314a2d0da`
  and `e746b7969565efc40a2d31d9f5f643e7f45847b6`. Common ancestor verified
  by `git merge-base` is c058 from I001. No commit/ref/index was changed.
- Object-only construction: enumerate full main tree; overlay each TASK-071
  delta path from the exact b653 tree except the two state consumers; replace
  the three accepted override paths with raw, unfiltered current bytes. Build
  nested tree objects with `git mktree -z`. Mode/type/OID entries, not checkout
  representations or normalized text, select the unchanged source payloads.
- Initial preparation tree `8065d20e37fe20ab5cf8545fecc50c8cf1787249` has
  433 entries; delta vs first parent 15 paths, delta vs main 21 paths.
  This initial raw tree contains I001 but predates this I002 evidence append.
  Later raw candidate trees may differ only by authorized task-envelope appends;
  final raw tree identity must be reported externally after Acceptance.
- Exact composite subject: 33 rows below in ascending unsigned UTF-8 path order,
  all `present`, mode `100644`, object format `sha1`. All rows use `full`
  except the current TASK-071 record's `task-record-v1`. Published TASK-072
  remains full immutable source evidence, not an excluded active task record.

| Path | Projection | Projected blob OID |
| --- | --- | --- |
| `.ai/PROJECT_CONTEXT.md` | full | `802263aa4e319b210358fe32e51376b88fb92d44` |
| `docs/en/design/DP-014-runtime-operational-identity-persistence.md` | full | `3ac9c684489271727528e468aba270cc693d48c4` |
| `docs/en/design/DP-016-runtime-activation-replacement-rollback.md` | full | `d4150dd181414ff35a1f9fe253ce1a81304882ea` |
| `docs/en/design/DP-017-runtime-recovery-reconciliation.md` | full | `e826f561b8c40c606127617679881f657e957cbd` |
| `docs/en/design/DP-022-runtime-execution-containment-and-evidence.md` | full | `016136b160512f6b469581497219d4dce18fb607` |
| `docs/en/design/DP-023-runtime-process-containment-bootstrap.md` | full | `714fb528095365c0d61394aa2790c260f955e2bc` |
| `docs/en/process/LLM_DEVELOPMENT_GUIDE.md` | full | `fd8e38f9c8868131326c860b2c3c231d72269f5f` |
| `docs/engineering/AGENT.md` | full | `f03b77624f5ebca3647a7b2dc7a4d88fff2d647e` |
| `docs/engineering/EXECUTION-INTERRUPTION-RECOVERY-ACCEPTANCE-SCENARIOS.md` | full | `b42d596feeb2da07233f37f658c02a69def00241` |
| `docs/engineering/PROCESS-001-AI-DEVELOPMENT-WORKFLOW.md` | full | `846ed56d9358e2b2f5e4263e5499dd5e1fffc782` |
| `docs/engineering/PROCESS-002-DOCUMENTATION-SYNCHRONIZATION.md` | full | `3045c90a36896a61f67d9a0fba67220428ba6785` |
| `docs/engineering/PUBLISHER-ACCEPTANCE-SCENARIOS.md` | full | `9da5ec7fc1bc5df81bb05bc76a91baf819633f6b` |
| `docs/engineering/TASK-TEMPLATE.md` | full | `1c59f9c277b7f98d4fea19c14e2f8414b3a72f32` |
| `docs/engineering/agents/coordinator.md` | full | `fcf152558f390eb75bae162eef96387d8e211115` |
| `docs/engineering/agents/publisher.md` | full | `772657d1aba378ea0c29888ca9e48f6b9967d085` |
| `docs/ru/design/DP-014-runtime-operational-identity-persistence.md` | full | `9bd5c26e04f1927e2b73438d51068ff904ce4351` |
| `docs/ru/design/DP-016-runtime-activation-replacement-rollback.md` | full | `bea988e0724b7a9478a272fe5159133dd21893d8` |
| `docs/ru/design/DP-017-runtime-recovery-reconciliation.md` | full | `5a457268a61c6d237d2aae148c832bc6e9127a6d` |
| `docs/ru/design/DP-022-runtime-execution-containment-and-evidence.md` | full | `15664f0b036b759fc91f7a8402e636715e50b8a7` |
| `docs/ru/design/DP-023-runtime-process-containment-bootstrap.md` | full | `cf60c2d64271cf5098efbfb1f39d23cae8b5c221` |
| `docs/ru/process/LLM_DEVELOPMENT_GUIDE.md` | full | `7881668e5421004bca0265dd2b755441804ea997` |
| `docs/tasks/README.md` | full | `52cca365f478ad71afef649956584a029e5c4d63` |
| `docs/tasks/TASK-071-RUNTIME-OWNER-SHUTDOWN-PROVENANCE.md` | task-record-v1 | `37d9625eb8e61429be7a27d98a0ca9a7416aefb5` |
| `docs/tasks/TASK-072-PUBLISHER-HANDOFF-STORE-RECOVERY.md` | full | `bb029549f51b8e2b61c7bff10acae4fff1dd377f` |
| `internal/runtimeactivation/activation.go` | full | `96166c353fa4ace4880584bc0b9cca02cffc670a` |
| `internal/runtimeactivation/activation_test.go` | full | `1cf316a5dda128053696e39a13435e44853c058e` |
| `internal/runtimeidentity/exact_attempt_snapshot.go` | full | `6fe87f6e50f9328e5396a5524d57e78156757354` |
| `internal/runtimeidentity/exact_attempt_snapshot_test.go` | full | `7328d18d4849feedc5c9c4453670ab1ce5f4cb94` |
| `internal/runtimeidentity/store.go` | full | `f81ee6b36f6ae93cbaad7a6198f97d4a6d07a7d8` |
| `internal/runtimeidentity/store_test.go` | full | `02b51a7142823657933ead25e7a43b369083564f` |
| `internal/runtimeidentity/types.go` | full | `d3c7aa94ec167e4f53b16c152ba22f5f5bf578cc` |
| `internal/runtimeorchestrationcontinuation/continuation_test.go` | full | `6db60f89cf1d75508ef3c0d105342854bcb83624` |
| `spec/current-state.md` | full | `ba032e270b21a4cb664fde3e6ec22e3308a084d4` |

- Composite manifest OID `e76b688192a1bfb0e7c207b5bb4e71e4dbee1045`:
  concatenate each row's path/projection/present/mode/OID fields and trailing
  NUL in that order, then `git hash-object --stdin` without `-w`. Current
  TASK-071 projection is exactly PROCESS-001 task-record-v1, no normalization.
- Bounded three-path identity remains separately
  `86de55dfe803f42f815bc46fd4fbb899dbf3592e`. Existing project-state resolvers
  still address the bounded rework; they do not themselves attest composite
  integration. Composite checkpoints resolve through entries explicitly bound
  to this new 33-path manifest and I001's verified intended parent tuple.
- First incomplete checkpoint: fresh tests/checks of an exact exported candidate
  tree, whole-tree preservation proof and composite documentation reconciliation.

### 2026-09-30 — I003 fresh composite Verification / documentation finding

- Tested actual Git tree `5656b2bbc505fb6ceacc20478d4c5af87d373ed6`;
  composite projected manifest `e76b688192a1bfb0e7c207b5bb4e71e4dbee1045`,
  bounded manifest `86de55dfe803f42f815bc46fd4fbb899dbf3592e`, fixed parent
  tuple and canonical rows I001/I002. This fresh evidence does not reuse the
  historical implementation or bounded rework verdict as composite Approval.
- Raw test snapshot:
  `C:\Users\dsdred\AppData\Local\Temp\uwp-task-071-integration-5697accb05114fb8b76706eaf9c7500b\snapshot-raw`.
  Export command `git -c core.autocrlf=false archive --format=tar` with this
  exact tree, extracted to a fresh directory. Independently recomputed raw Git
  blob OIDs match all 433 materialized files, including after tests. No extra
  files, deletion, byte/mode normalization or generated repository artifact.
- The first default archive was rejected: `.gitignore` exported 370 bytes /
  34 CRLF rather than the 336-byte LF Git blob, due to `core.autocrlf=true`.
  A tool-result delivery error also required inspect-first temporary inventory.
  Neither first attempt is claimed PASS; no tests used that converted snapshot.
  Command-scoped autocrlf=false produced the verified raw snapshot, without
  Git configuration, index, ref or source-file changes.
- Fresh offline commands in the raw snapshot, `GOPROXY=off/GOSUMDB=off`:
  `go test ./... -count=1` completed exit 0, all packages PASS;
  `go test ./internal/runtimeidentity ./internal/runtimeactivation
  ./internal/runtimelifecycle ./internal/runtimeorchestrationcontinuation
  -count=25` completed exit 0, all four PASS; `go vet ./...` completed exit 0.
- Formatting on the eight TASK-071 Go paths is clean. Broad raw-snapshot
  `gofmt -l` reports two unchanged common-base files,
  `internal/runtimecommandidempotency/orchestration_admission.go` and its test,
  with existing CRLF representations. They are outside the integration delta
  and remain exact source Git blobs. The initial absolute-argument formatting
  invocation had a PowerShell launcher failure; the relative-argument repeat
  completed, and the broad check is not misreported as globally clean.
- Race availability: default command exit 2 requires CGO; explicit
  `CGO_ENABLED=1` command exit 1 reports missing `gcc`. Race row is unavailable
  (`PASS WITH LIMITATION`), not an executed race-test PASS. Fresh full/stress/vet
  checks are substitutes, not proof of race-detector execution.
- Whole-tree proof: 18 TASK-071-only payload entries exactly match first-parent
  blob/mode identities; 12 main-only entries exactly match published main;
  all 400 entries outside the 33-path union equal both parents and common base.
  Exact path inventory includes additions/deletions, with no extra change.
  Current task prefix is append-only; real index is empty and refs unchanged.
  `git diff --check` against both parents completed exit 0.
- Documentation mechanics: 238 relative links, broken 0; conflict markers 0;
  process mirror headings 45/45. All ten DP mirror and twelve published process
  payload identities are preserved. Nine handoff scenarios S-056-S-064 and
  recovery R-092-R-097 remain exact published bytes; integration grants no
  actual qualification, source Release, Route, Accept or publication capability.
- Potential composite documentation blocker D001: imported
  `docs/tasks/README.md:6-11` still declares TASK-072 the current process task
  and routes latest gates through its recomputed current-subject envelope.
  TASK-072's fixed source manifest is
  `00108732096fc68ad06391dab709edd962c1ebf1`; recomputing its 14-path subject
  in this composite yields `41f26b13dd90d85fd0418f6abe1b20a018eaf835`, because
  the two accepted project-state files changed. No current TASK-072 envelope
  matches that composite identity. Historical TASK-072 Acceptance remains valid
  for its published immutable source; it must not be reinterpreted as current
  integration Acceptance. D001 awaits independent composite Review adjudication.
- PROCESS-002 is not claimed Synchronized until D001 is resolved. No index,
  published TASK-072 record or accepted rework bytes were edited to mask the
  issue. The current authority requires preservation of main payload plus only
  accepted bounded rework, not a new metadata scope expansion.

### 2026-09-30 — I004 resulting scope inventory / independent Review handoff

- Composite payload audit `33 Required / 0 Questionable / 0 Removable`: 18
  unchanged accepted implementation/design payload paths from b653, 12 unchanged
  governance/index/evidence payload paths from fb634, and three already accepted
  reconciliation/evidence paths. Against first parent delta is 15 paths; against
  current main delta is 21. Actual authored content remains the accepted two
  state consumers and append-only TASK-071 envelope; no code/design rework.
- Removal test: dropping TASK-071 paths loses accepted implementation/design;
  dropping main paths loses published process repair or its required contracts,
  mirrors/scenarios/index/evidence; dropping either state consumer or task
  evidence loses required reconciliation or reconstructable gates. Required
  preservation alone does not resolve D001's live-projection semantics.
- Size Guard: >15-path composite assessment is an indivisible integration of
  two fixed already-approved source payloads, not new 33-path implementation.
  Splitting sources would violate the requested parent/preservation tuple;
  no new feature or discretionary change is included. Independent Reviewer
  must confirm this assessment and D001 before any composite Acceptance.
- Fresh Review subject is exactly I002's 33-path manifest, parent tuple and
  an actual candidate Git tree whose only changes since the tested tree are
  these authorized task-envelope appends. It is not the three-path Acceptance.
- First incomplete checkpoint: independent actual-byte composite Review;
  unresolved mandatory documentation finding forbids Coordinator Acceptance.
  STOP before additional scope authority, Commit Gate or publication mutation.

### 2026-09-30 — I005 interruption recovery and authorized D001 rework

- Current explicit user input resumes integration and authorizes exactly one
  additional content path, `docs/tasks/README.md`, only for live routing/current
  state needed by TASK-071's exact composite subject. Existing append-only task
  evidence authority remains; no new task or implementation/design scope.
- Inspect -> Reconstruct -> Reconcile found the record ended at I004. The
  attempted I005/I006 append was not executed: automatic approval review failed
  due to a usage limit, explicitly not a determination of unsafe action.
  No rejected record, disposition or Acceptance is treated as completed.
  The current input resolves continuation; no approval-check bypass is used.
- HEAD remains `b653337dc93eb9419648246e447813dc168a16ea`, main/origin-main
  remain `fb6341e48f6abdea0f2c2b23ea4c949a16ed3a70`; real index empty.
  Prior candidate tree `f164832d34d1bbdaba70c59fa4a6e1857fe3413e` exists.
  Bounded three-path manifest independently recomputed unchanged:
  `86de55dfe803f42f815bc46fd4fbb899dbf3592e`. Original task prefix and lower
  state sections remain byte-identical; no publication side effect occurred.
- Prior independent actual-byte composite Review returned `NEEDS REVISION`
  on f164/e76b688192a1bfb0e7c207b5bb4e71e4dbee1045 with one blocker D001,
  also explicitly confirmed by the current user. This recovered finding is
  recorded now; it is not a claim that the rejected append previously existed.
  Historical TASK-072 Acceptance/source remain valid; the defect is the
  imported index's current TASK-072 routing/current-manifest mismatch in I003.
- Bounded rework remains the same three-path identity. README is a separate
  integration reconciliation path; the 33-path composite set already contains
  it, so only its full row will change. Eleven other main payload paths and
  all 18 TASK-071 payloads must remain exact source blob/mode identities.
- Primary agent sequential Documentation/Tester roles and independent Reviewer
  `process_proposal_review` remain assigned. Ordered next stages: README-only
  metadata repair; raw object-only tree reconstruction; fresh applicable full
  tests/vet/docs/source/scope checks; independent composite Review; exact
  composite Coordinator Acceptance/integrity; STOP before Commit Gate.
- No commit, staging, Git merge command, ref movement, PR #77 update, push,
  external merge/deletion/publication, old-authorization reuse or transfer reuse.
  Existing history is append-only; no amendment/rebase/force-push.

### 2026-09-30 — I006 D001 actual metadata rework / new composite subject

- Actual README live routing now names TASK-071 integration, stable In Progress,
  exact intended parent tuple and a current 33-path composite-envelope resolver.
  It explicitly distinguishes the unchanged three-path bounded identity.
- TASK-072 is described only as historical published process repair, with
  unchanged branch/base/source commit, PR #76/main OID and historical source
  manifest. Its record and other published process bytes are not rewritten.
  TASK-070 published facts are retained; its pre-publication candidate list
  projection receives an explicit source-snapshot/not-current-routing qualifier,
  without changing the historical candidate description or other task entries.
- New exact composite subject: all I002's 33 ordered rows/projections/states/
  modes/OIDs remain unchanged except row 22, `docs/tasks/README.md`:
  `full | present | 100644 | dc849582b9b6a76fcab22c44b09f2aa663b89147`.
  Object format sha1, intended parents/common base I001. Canonical manifest
  `52dfbaebad7b6804c3b9a3710129f017a5f3292a`, calculated using the same raw
  NUL-separated algorithm and task-record-v1, no normalization.
- Bounded identity remains `86de55dfe803f42f815bc46fd4fbb899dbf3592e`.
  Historical blocked composite e76b688192a1bfb0e7c207b5bb4e71e4dbee1045
  is not reused as current verification/Approval/Acceptance.
- Initial repaired candidate object `551cb57714539eebd444f1d22084d9eed0d15692`
  contains I005 and the corrected index; it predates this I006 evidence append.
  It has 433 entries; source-preservation groups are now 18 unchanged TASK-071,
  11 unchanged main and four authorized reconciliation/evidence paths, with
  400 entries outside the union. Real index remains empty, refs unchanged.
- First incomplete checkpoint: raw snapshot/materialization proof and fresh
  applicable full tests/vet/documentation/scope checks, then independent Review
  of this exact repaired composite. No composite Acceptance yet exists.

### 2026-09-30 — I007 fresh repaired-composite Verification / Scope Audit

- Exact tested tree `6a37fa5df4722f719d3471cbc01a6b7a09f1517b`, object format
  sha1, composite manifest `52dfbaebad7b6804c3b9a3710129f017a5f3292a` with
  I006's ordered row replacement and I001's intended parents/common base.
  Raw export snapshot is the I003 temporary root's `snapshot-d001-raw` directory.
  Command-scoped autocrlf=false archive/extraction and raw hash verification
  prove all 433 files before and after checks; no normalization equivalence.
- Fresh offline full `go test ./... -count=1`: completed exit 0, all packages
  PASS. Fresh four-package `-count=25` command from I003: completed exit 0,
  all four PASS. Fresh full `go vet ./...`: completed exit 0, PASS.
  GOPROXY/GOSUMDB off; no repository dependency or test-file mutation.
- All eight TASK-071 Go paths have clean `gofmt -l`. Full raw-snapshot scan
  still reports only the same two unchanged CRLF common-base Go files from
  I003. They remain exact source blobs; no broad-format clean claim or fix.
  Current CGO_ENABLED is 0 and gcc is unavailable. I003's race limitation
  remains explicit, not a claim of a successful race-detector execution.
- Repeated docs checks: 239 relative links, broken 0; conflict markers 0;
  process mirror headings 45/45. Both-parent `git diff --check` exit 0.
  D001 live routing is repaired; current index resolves only TASK-071 composite
  gates. Historical TASK-072 source is not a current-subject resolver. Its
  full blob `bb029549f51b8e2b61c7bff10acae4fff1dd377f` remains unchanged.
- Preservation audit: 18 TASK-071 payloads exact first-parent blob/mode;
  11 main payloads exact second-parent blob/mode; 400 other entries exact both
  parents/common base. README is the sole additional main-byte exception and
  only authorized live metadata changes. Its introduction and remaining raw
  historical bytes equal main after removing the explicit TASK-070 qualifier.
  The entire 60,136-byte pre-interruption task-record prefix from f164 remains
  byte-identical, including all historical and prior rework evidence.
- Bounded three-path manifest independently recomputed unchanged 86de55 from
  I005; no .ai/spec content changes. Four current authored/evidence paths only;
  real index empty, HEAD/main/origin-main unchanged. Candidate object changes
  create neither a commit nor an immutable publication Target.
- PROCESS-002 applicability: task record, .ai/spec and task-index routing
  Synchronized for the distinct bounded/composite subjects. Other main
  governance/mirrors/scenarios and historical TASK-072 record are unchanged
  source payloads; DP/code/test/dependency/API/roadmap/CHANGELOG mutations N/A.
  No new product/architecture/readiness/production activation is claimed.
- Resulting-tree Scope Audit: `33 Required / 0 Questionable / 0 Removable`
  (18 TASK-071, 11 main, four reconciled/evidence paths); actual authored
  worktree audit `4/0/0`. Every preserved source and each independent state
  consumer/routing/evidence is necessary; no optional/removable change.
- Size Guard retains one indivisible fixed-parent integration, not new
  33-path implementation. The one additional index repair resolves its live
  routing defect without new behavior. Fresh independent Reviewer must confirm
  scope/removal test and D001 closure on this exact subject.
- First incomplete checkpoint: fresh independent repaired-composite Review;
  no new Coordinator Acceptance, Commit Gate or publication permission yet.

### 2026-09-30 — I008 fresh independent composite Review completed

- Independent Reviewer `/root/process_proposal_review`: APPROVED, blocking
  findings 0, unresolved findings 0; fresh actual-byte Architecture/Process
  Review and Final Review, not reuse of the earlier NEEDS REVISION verdict.
- Reviewed raw Git tree `e9cf717f7660584113b7a356a2d16d429486b979`;
  independently recomputed exact 33-path manifest
  `52dfbaebad7b6804c3b9a3710129f017a5f3292a`. Complete ordered rows are I002
  with only I006's row-22 replacement; object format sha1, all modes 100644.
  Intended ordered parents are b653337dc93eb9419648246e447813dc168a16ea,
  then fb6341e48f6abdea0f2c2b23ea4c949a16ed3a70; verified common ancestor
  c058da69f2296e52a8e32cc25e195190389dbca7. A tree itself has no parents.
- D001 CLOSED prospectively: actual README routes the current TASK-071
  composite through its matching envelope, separately identifies bounded
  Acceptance and historical published TASK-072, and qualifies the unchanged
  historical TASK-070 candidate description. Candidate PROCESS-001 sections
  at lines 378-385 and PROCESS-002 lines 145-146, 230-235 are satisfied.
  Historical TASK-072 Acceptance is neither rewritten nor used for this result.
- Independent raw proof: 433 entries; 18 TASK-071 payloads and 11 main payloads
  retain exact source mode/type/OID; 400 other entries equal both parents and
  common base. TASK-072 full blob remains
  `bb029549f51b8e2b61c7bff10acae4fff1dd377f`; all 60,136 pre-interruption
  record bytes remain identical. README historical bytes pass I007's exact
  preservation comparison. Bound three-path manifest remains
  `86de55dfe803f42f815bc46fd4fbb899dbf3592e`.
- All 433 raw snapshot files independently match tested tree 6a37fa5d from
  I007. Tested-to-reviewed difference is only the excluded TASK-071 evidence
  envelope. Fresh full tests, four-package stress and vet PASS are verified;
  unavailable race detector and two unchanged common-base formatting findings
  remain explicit limitations, not successful checks.
- Removal-test/Scope Audit confirmed `33 Required / 0 Questionable / 0
  Removable`; actual authored paths `4/0/0`. The fixed-source integration
  justifies Size Guard reassessment; no new implementation/design work.
  Index empty; refs unchanged; both-parent whitespace checks PASS.
- Reviewer confirmed the rejected append did not execute and is not credited.
  Current I005-I007 evidence is actual rework; no fabricated prior closure.
- Next checkpoint: exact-subject composite Coordinator Acceptance and
  independent post-decision append-integrity verification. This Review grants
  no staging, commit, changed-Target authorization, handoff reuse or publication.

### 2026-09-30 — I009 exact-subject composite Coordinator Acceptance

- Coordinator: ACCEPTED integration result, subject object format sha1,
  canonical 33-path manifest `52dfbaebad7b6804c3b9a3710129f017a5f3292a`;
  exact ordered paths/projections/states/modes/OIDs are I002 plus I006's single
  README replacement. This is the current composite Acceptance resolver;
  it does not replace or reinterpret the historical three-path Acceptance.
- Intended ordered commit parents: TASK-071
  `b653337dc93eb9419648246e447813dc168a16ea`, then published main
  `fb6341e48f6abdea0f2c2b23ea4c949a16ed3a70`; common ancestor/fixed historical
  publication base `c058da69f2296e52a8e32cc25e195190389dbca7`. Original commits
  and bounded manifest `86de55dfe803f42f815bc46fd4fbb899dbf3592e` are unchanged.
- Acceptance binds the actual resulting bytes reviewed as e9cf717f from I008,
  I007 completed Verification and I008 APPROVED with zero unresolved findings.
  D001 is closed. Final raw tree will be rebuilt/reported externally after
  authorized append-only gate evidence; task-record-v1 excludes only Status
  evidence/envelope, so this metadata does not alter the accepted subject.
  No raw-tree self-hash or normalization-based equivalence is asserted.
- Scope Audit ACCEPTED: full composite `33 Required / 0 Questionable / 0
  Removable` = 18 unchanged TASK-071 payloads + 11 unchanged main payloads +
  four authorized reconciliation/evidence paths. All 400 outside-union entries
  equal both parents/common base; all 433 entries have accounted provenance.
  Actual authored worktree scope is only .ai/PROJECT_CONTEXT.md,
  spec/current-state.md, append-only TASK-071 record and docs/tasks/README.md
  live metadata. Prospective delta is 15 paths against first parent and 22
  against main, not permission to author 33 paths. No other main/TASK-072
  bytes, implementation, design, dependency or activation are changed.
- Required gates completed: Inspect/Reconstruct/Reconcile, D001 authorized
  metadata rework, raw resulting-tree/snapshot proof, full tests/stress/vet,
  applicable documentation checks, exact-subject independent composite
  Architecture/Process/Final Review and resulting-tree Scope Audit.
  Race execution remains unavailable and unchanged base formatting is explicit.
- This decision accepts the prospective integration result only. No commit,
  staging, PR #77 update, push, merge, deletion or publication occurred here.
  Old publication authorization/Accepted Handoff are not carried to a future
  changed immutable Target; old operational transfer is not modified.
- Next required user gate after post-decision integrity and STOP is exactly
  `Разрешаю коммит.`. It authorizes one checked integration commit only;
  Commit Gate must prove actual ordered parents, entire raw staged-tree match,
  canonical manifest, authorized final evidence, file set and final checks.
  LF/CRLF substitution is not equivalence. No amend/rebase/force-push.
  A resulting commit becomes a new immutable Target and requires its separate
  publication authorization and applicable ownership/P0/handoff gates.

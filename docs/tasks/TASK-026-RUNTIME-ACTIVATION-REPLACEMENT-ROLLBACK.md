# TASK-026 — Runtime Activation, Replacement, and Rollback Implementation

## Status

`Blocked — missing DP-015/DP-020 orchestration-admission refinement
(2026-08-27)`.

The separately accepted and published readiness verdict `READY — UNBLOCK
TASK-026`, with matrix 7 Direct / 10 Compositional / 2 Missing core / 0 Missing
prerequisite / 0 Missing external / 0 Deferred, is preserved as historical
evidence. The current implementation cycle supersedes that readiness for live
execution: repeat Architecture Confirmation returned `NEEDS DECISION` and
`SPLIT REQUIRED` because the current DP-015/DP-020 eager-generation and
combined inspect/claim boundaries cannot satisfy exact replay-first admission
and late generation allocation. The rejected implementation was removed, so
production and test diff is zero. DP-016 remains Approved/Planned. A separate
narrow DP-015/DP-020 design and isolated implementation prerequisite is `Not
Activated` and has no Task ID.

## Task Contract

### Task Mode

`Implementation`: bounded isolated DP-016 activation/replacement/rollback
orchestrator and its proof/regression tests.

### Why Now

- the Design-only TASK-026 reassessment was Coordinator Accepted, committed as
  `e4c0cf8`, and published through PR #50 into synchronized
  `main@65af65c3154d70db10de52483149960a42dcfb9a`;
- its fresh Architect verdict is `READY — UNBLOCK TASK-026`, and all mandatory
  prerequisites are present in isolation;
- its recovery handoff names this exact Implementation contract transition as
  the sole next candidate and leaves it `Not Activated` until a separate bare
  continuation command;
- the current bare continuation command activates that candidate without
  authorizing commit or publication.

### Definition of Done

1. One isolated `internal/runtimeactivation` package implements the exact
   activation, replacement, and explicit rollback composition authorized by
   DP-016 and the accepted readiness boundary.
2. Exact-version validation and same-target satisfied outcomes occur before
   command/lifecycle mutation; replacement proves old release before Continue
   and fresh claim; rollback uses an exact Published target and a fresh attempt.
3. The orchestrator preserves DP-014/DP-015/Owner truth, replay, cancellation,
   unresolved/indeterminate outcomes, zero Host overlap, and per-Instance
   independence without holding locks across repository, lifecycle, wait, or
   storage calls.
4. Proof and regression tests cover all nineteen DP-016 §25 rows, including the
   two previously Missing-core same-target decisions and the ten compositional
   rows, with current Direct prerequisites protected against regression.
5. Applicable focused/stress/full tests, vet, formatting, module, documentation,
   diff, scope, and independent-review gates pass; PROCESS-002 truthfully
   synchronizes the isolated implementation boundary.

### Out of Scope

- changing Approved/Frozen semantics, DP-016 acceptance proofs, or existing
  prerequisite contracts;
- generic orchestration engines, registries, service locators, exported public
  API, recovery hooks, new lifecycle abstractions, or changes to existing seams;
- production wiring, Control Service/API, persistence, recovery, reporting,
  authorization policy, automatic rollback, or multi-node supervision;
- stage, commit, push, PR, merge, publication, branch deletion, or mutation of
  `main`.

### Verification Plan

- complete the repository-first Documentation Baseline and explicit Architect
  confirmation before production/test mutation;
- preserve the Existing Coverage Report below before adding tests;
- implement the smallest closed package-local dependency surface and proof
  tests mapped to DP-016 §25 rows 1–19;
- run focused tests and repeated shuffled stress for the affected packages,
  full `go test ./... -count=1`, `go vet ./...`, `gofmt -d`,
  `go mod tidy -diff`, and `git diff --check`; attempt race verification and
  record an exact environment limitation if it remains unavailable;
- validate exported-surface confinement, EN/RU parity, links, live status
  claims, exact file set, Size Guard, Scope Audit, and independent final review.

### Selection Evidence

- repository preflight: clean synchronized
  `main@65af65c3154d70db10de52483149960a42dcfb9a` with
  `main == origin/main` at branch creation;
- current task branch:
  `feature/task-026-runtime-activation-orchestration-implementation` at that
  trusted baseline;
- no other active task or unexplained staged, unstaged, or untracked state was
  present at intake;
- TASK-026 is the sole explicit next candidate; its accepted readiness record
  provides the exact one-package implementation boundary;
- rejected alternative: a new prerequisite or architecture task, because the
  current matrix has zero Missing prerequisite and the Accepted contracts are
  sufficient;
- rejected alternative: production wiring, because it is a materially larger
  later integration slice and explicitly outside the readiness boundary.

### Scope, Sources, and Branch

- repository: `E:\wikiPRJ\universal-websocket-platform`;
- trusted baseline and intake HEAD:
  `65af65c3154d70db10de52483149960a42dcfb9a`;
- task branch:
  `feature/task-026-runtime-activation-orchestration-implementation`;
- first content change: this Implementation task-contract transition in the
  existing task record; historical sections remain evidence;
- mutation permitted before current Architecture Confirmation: this task
  record only;
- sources of truth: Active ARCH-004 §19(4), Approved DP-014–DP-019, applicable
  DP-020/DP-021 contracts, TASK-046/TASK-047 accepted evidence, current package
  surfaces and tests, PROCESS-001, and PROCESS-002;
- forbidden git actions: stage, commit, push, merge, rebase, reset, branch
  deletion, remote mutation, or modification of `main`.

### Roles and Ordered Stages

1. Coordinator — recovery reconstruction, deterministic selection, intake,
   gates, Size Guard, scope audit, and closure.
2. Documentation Agent — repository-first baseline and later PROCESS-002
   synchronization without overstating isolated implementation.
3. Architect — explicit confirmation that the current contract fits the
   Accepted DP-016 boundary and requires no new decision.
4. Developer — production implementation and scoped tests only after the
   Architecture Confirmation.
5. Tester — independent Verification Matrix execution and coverage audit.
6. Reviewer — independent review; any blocking finding returns through the
   bounded Rework Loop and invalidates affected downstream evidence.
7. Documentation Agent — final applicability/status synchronization.
8. Coordinator — final Scope Audit, final independent review, Acceptance, and
   next-task recommendation.
9. Publisher — not applicable and not authorized.

### Stop Conditions

Stop and return to Coordinator on dirty/unattributed or diverged repository
state, source-precedence conflict, ambiguous ownership/order, any newly found
prerequisite, required Approved/Frozen contract change, critical documentation
drift, public API or production-wiring need, more than one new package, more
than 500 net production lines, more than 15 total task files, more than one
independently deliverable behavior, or failed mandatory verification/review.

### Existing Coverage Report — Current Read-Only Gate

`Existing Coverage`: accepted TASK-047 evidence proves the isolated DP-015
tracked-Start managed-parent plus preclaimed `StopOld` prerequisite. The fresh
Architect/Tester handoffs below map the current seams and existing tests as
**7 Direct / 10 Compositional / 2 Missing core / 0 Missing prerequisite / 0
Missing external / 0 Deferred**.

`Coverage Gap`: proofs 2 and 14 are the only Missing-core rows and belong to the
new orchestrator's authorized pre-claim zero-mutation same-target decisions.
The ten Compositional rows still require TASK-026 end-to-end implementation and
proof tests; existing isolated seams do not prove that future composition.

`Added Proof Tests`: `internal/runtimeactivation/activation_test.go` now maps
all nineteen DP-016 §25 proofs, including real-Host successful replacement and
rollback, exact Running same-target zero mutation, same-address zero Host
overlap, release-before-fresh-claim, fresh rollback attempt, and binding-before-
Load order.

`Added Regression Tests`: replay without reinvocation, Stop failure,
cancellation, unresolved barrier, startup failure without resurrection or
automatic rollback, and different-Instance progress are executable; existing
tracked-Start/rendezvous/managed-continuation suites remain green.

`Remaining Limitations`: production wiring, external persistence, recovery,
reporting, concrete authorization policy, and public management API remain out
of scope. Race execution was unavailable at readiness time because `gcc` was
missing; the implementation gate must attempt it again and record the current
fact rather than inherit the historical limitation.

### Recovery Checkpoints

1. Recovery reconstruction — `Proven Completed`: clean synchronized
   `main@65af65c3154d70db10de52483149960a42dcfb9a`, published readiness history,
   active/next-task state, and absence of competing work were inspected.
2. Implementation intake and safe branch preparation — `Proven Completed`:
   current branch is the exact task branch above and this task record is the
   first content change.
3. Documentation Baseline and bounded pre-implementation synchronization —
   `Proven Completed`: the critical stale recovery/navigation drift recorded
   below is repaired in the exact 19-path documentation subject.
4. Architecture Confirmation — `Proven Completed`: independent Architect
   verdict `PASS`, exact one-package boundary and ordering constraints are
   recorded below.
5. Initial Developer implementation — `Proven Completed`: one new package, 499
   production lines, no existing-seam change; handoff recorded below.
6. Initial Independent Tester verification — `Proven Completed`: `PASS WITH
   ENVIRONMENT LIMITATION`, findings `0/0`; the later Reviewer invalidated the
   claimed complete proof mapping.
7. Initial Independent Review — `Proven Completed`: `Needs Revision`, blocking
   findings `6`, non-blocking findings `0`; exact findings and reviewed identity
   are recorded below.
8. Developer rework attempt — `Proven Stopped Before Mutation`: B-001 and B-003
   cannot be corrected through the Architecture-confirmed existing seams.
9. Repeat Architecture Confirmation and Size Guard decision — `Proven
   Completed`: `NEEDS DECISION`; the required DP-015/DP-020 refinement is a
   separate prerequisite behavior and cannot enter this task.
10. Rejected implementation cleanup — `Proven Completed`: the two
    `internal/runtimeactivation` files were removed; production, test, module,
    dependency, generated, staged, and untracked diff is zero.
11. Blocked-handoff PROCESS-002 — `Proven Completed`: the exact 19-path
    documentation set records the superseding blocker and unactivated
    prerequisite without changing any Design/Implementation Status.
12. Final Tester, independent Reviewer, Scope Audit, and blocked closure —
    first unfinished checkpoint.

### Documentation Baseline — Implementation Intake

#### Inventory and Status Evidence

All **26/26** scoped files are present: TASK-026, TASK-047, task index,
ARCH-004 EN/RU, DP-014–DP-021 EN/RU, `spec/current-state.md`,
`spec/decisions.md`, `.ai/PROJECT_CONTEXT.md`, and MASTER_PLAN EN/RU.

- TASK-026 at this baseline checkpoint: `In Progress — Implementation intake
  (2026-08-27)`; the later repeat Architecture verdict supersedes it with the
  live Blocked status at the top of this record;
- TASK-047: `Completed — Coordinator Accepted (2026-08-25)`;
- ARCH-004: `Active`, Approved stability in both mirrors;
- DP-014: Approved / Implemented in isolation;
- DP-015: Approved; isolated primitive/parent-phase/rendezvous/managed-gate/
  tracked-Start admission slices implemented, complete extension still Planned;
- DP-016, DP-017, and DP-018: Approved / Planned;
- DP-019: Approved / Planned overall with recorded isolated partial seams;
- DP-020: Draft / Planned overall with Slice 3 implemented in isolation;
- DP-021: Draft / Partial, implemented in isolation.

No scoped document presents the DP-016 orchestrator, terminal publication,
production wiring, recovery, or reporting as implemented. TASK-047 evidence is
consistently isolated from the still-unproven TASK-026 composition.

#### Parity, Structure, and Links

All **10/10** scoped EN/RU pairs are present. Heading/fence counts match per
pair: ARCH-004 `29/29`, `8/8`; DP-014 `28/28`, `4/4`; DP-015 `30/30`,
`4/4`; DP-016 `30/30`, `4/4`; DP-017 `30/30`, `2/2`; DP-018 `27/27`,
`2/2`; DP-019 `25/25`, `16/16`; DP-020 `34/34`, `12/12`; DP-021
`21/21`, `10/10`; MASTER_PLAN `36/36`, `0/0`. Scoped relative-link check:
**294 valid / 0 broken**.

#### Drift Verdict

Initial Documentation Baseline verdict: **`Drift Detected — Critical`**.
Approved/Active status, EN/RU normative meaning, and the
Planned-versus-Implemented boundary were mutually consistent, but the intake
record retained the terminal Design-only recovery anchor and stale baseline,
while live navigation, project state, roadmap, and related proposal summaries
still described the separately activated Implementation transition as `Not
Activated`. The EN/RU design indexes and the DP-021 implementation-boundary
conclusion also retained a superseded Blocked/pending-reassessment statement.

Bounded repair verdict: **`Synchronized for Architecture Confirmation`**.
The exact 19-path update makes current task activation and recovery facts
consistent without claiming implementation, changing any Design or
Implementation Status, or rewriting historical closure chronology. DP-016
remains Approved/Planned and no production/test mutation is present.

#### Applicability

- this TASK-026 record: **Required and updated** for current recovery evidence;
- task index, `spec/current-state.md`, `spec/decisions.md`, and
  `.ai/PROJECT_CONTEXT.md`: **Required and updated** for current task
  navigation;
- MASTER_PLAN EN/RU and design indexes EN/RU: **Required and updated** for the
  durable active-work and design-navigation boundary;
- DP-015, DP-016, DP-019, DP-020, and DP-021 EN/RU: **Required and updated**
  only for live TASK-026 activation/boundary wording; all statuses and contract
  semantics are preserved;
- ARCH-004, DP-014, DP-017, and DP-018 EN/RU: **Not applicable at baseline**;
  no contract, status, or implementation fact changed;
- root README EN/RU and `CHANGELOG.md`: **Not applicable**; no user-facing,
  release, or production capability changed;
- production, tests, modules, dependencies, generated artifacts: **Not
  applicable and forbidden** before Architecture Confirmation.

#### Reproducible Commands and Results

- repository reconstruction:
  `git branch --show-current; git rev-parse HEAD; git rev-parse main; git status
  --short; git log -1 --oneline --decorate` → branch
  `feature/task-026-runtime-activation-orchestration-implementation`, HEAD/main
  `65af65c3154d70db10de52483149960a42dcfb9a`, clean pre-intake worktree and
  one attributed task-record change at baseline review;
- inventory/status/structure: PowerShell `Test-Path`, `Get-Content`, and
  `Select-String -Pattern '^#{1,6} |^```'` over the exact 26-file inventory
  above → `26 present / 0 missing`, pair counts recorded above;
- factual drift search: scoped `rg -n` searches for TASK-026, `Not Activated`,
  Blocked, and current-task summaries → stale live matches repaired in the
  exact applicability set; explicit historical matches preserved;
- relative-link validation: PowerShell Markdown-link extraction plus
  `Test-Path -LiteralPath` over the exact inventory → `294 valid / 0 broken`;
- task-record diff hygiene: `git diff --check` → `PASS` (Git emits only the
  repository line-ending advisory; no whitespace error).

The current later boundary is a separate production composition/integration
intake after accepted isolated implementation. It remains **`Not Activated`**;
Control Service/API wiring, recovery, reporting, and Production Activation are
outside this task and are not assigned a Task ID here.

### Independent Architect Readiness Handoff

Fresh Architect verdict: **`READY — UNBLOCK TASK-026`**. This supersedes the
2026-08-25 blocker only for current readiness and does not alter historical
evidence below. The current matrix is **7 Direct / 10 Compositional / 2 Missing
core / 0 Missing prerequisite / 0 Missing external / 0 Deferred**.

| DP-016 §25 proof | Class | Concise current evidence |
|---:|---|---|
| 1 | Compositional | Exact repository `Get` plus DP-014 attempt claim/binding; orchestrator composition absent. |
| 2 | Missing core | Orchestrator-owned exact-Running same-target zero-mutation satisfied decision. |
| 3 | Compositional | Exact different-target replacement selection composes accepted seams. |
| 4 | Compositional | Proven old release must precede fresh claim; existing Owner/aggregate seams support the order. |
| 5 | Compositional | TASK-047 `ExecuteManagedParentFromTrackedStart`, preclaimed `StopOld`, and `Owner.StopExpectedAttempt`. |
| 6 | Compositional | Accepted continuation is callable only after proven old release. |
| 7 | Direct | Continue/Stop atomic winner semantics are implemented and tested. |
| 8 | Direct | Stop-first prevents linked claim and Load. |
| 9 | Direct | Start-first permits one pending Stop from the original claiming stack. |
| 10 | Compositional | Exact late-Stop outcome must be mapped by the orchestrator callback. |
| 11 | Compositional | Failed/unproven Stop retains ownership and the unresolved admission barrier. |
| 12 | Compositional | Startup failure composes accepted Owner truth without resurrection/automatic rollback. |
| 13 | Compositional | Exact Published rollback validation plus fresh attempt identity compose existing seams. |
| 14 | Missing core | Orchestrator-owned same-target activation/rollback zero-mutation satisfied decision. |
| 15 | Compositional | Phase cancellation truth must be preserved across the end-to-end callback order. |
| 16 | Direct | Panic, `runtime.Goexit`, generation loss, and reconstruction retain unresolved truth. |
| 17 | Direct | DP-014 exact claim/generation bind occurs before Load or preparation does not begin. |
| 18 | Direct | Per-Instance admission and lifecycle progress remain independent. |
| 19 | Direct | EN/RU contract, gates, failure matrix, and Planned status are aligned. |

#### Ownership and Order

- `runtimelifecycle.Owner` remains the sole lifecycle/Host owner;
- `runtimeidentity.Store` owns aggregate revision, exact attempt/generation,
  Running, release, and terminal facts;
- `runtimecommandidempotency.Boundary` owns submission authorization, claims,
  permits, replay, tracked-Start/managed-parent admission, rendezvous, and
  unresolved barriers;
- accepted continuation, managed Flow, and the TASK-043 invoker retain their
  narrow exact-scope claim/bind/preparation responsibilities;
- the TASK-026 orchestrator may own only exact version reads/decisions,
  callback closures, exact Owner-outcome mapping, and ordering of DP-014
  publication before DP-015 phase/parent/command terminalization.

Only `ConfigurationVersionRepository.Get(exactID)` may run before command or
lifecycle mutation; returned/requested IDs, Configuration ownership, and exact
`Published` state must match. Replacement publishes proven old release before
Continue and fresh claim. Managed Start uses the TASK-043 invoker; linked Stop
uses `StopExpectedAttempt` through TASK-047 preclaimed authority. Indeterminate
DP-014 truth leaves the linked DP-015 set unresolved. No command, aggregate, or
Owner lock spans repository, lifecycle, wait, or storage calls.

#### Implementation Boundary

The Ready boundary is one new isolated `internal/runtimeactivation` package and
one immutable activation/replacement/rollback orchestrator with package-local
narrow dependency interfaces, a validating constructor, closed entry points,
and closed result categories. No generic engine, registry, service locator,
public API, recovery hook, lifecycle abstraction, existing seam change, or
production wiring is authorized. Crossing one package, 500 net production
lines, 15 total task files, or one independently deliverable behavior reopens
the Size Guard before implementation.

### Independent Tester Readiness Handoff

Tester verdict: **`PASS WITH ENVIRONMENT LIMITATION`**, blocking findings `0`,
non-blocking findings `0`. The handoff verifies the current read-only coverage
mapping and accepted prerequisite seams; it does not verify an absent TASK-026
implementation.

Exact retained commands/results:

- `go version; go env GOOS GOARCH CGO_ENABLED; gcc --version` → Go `1.26.5`,
  `windows/amd64`, `CGO_ENABLED=0`; `gcc` missing;
- default-cache `go test` over the eight focused packages
  `configurationversion`, `runtimeidentity`, `runtimelifecycle`,
  `runtimecommandidempotency`, `runtimeorchestrationbinding`,
  `runtimeorchestrationcontinuation`, `runtimelaunchflow`, and
  `runtimemanagement` with `-count=1` → exit `1`: six packages pass and two
  fail only on default-cache `Access is denied`; environment reconciled before
  retry;
- the same eight-package command with workspace-local `TEMP` and `GOCACHE` →
  exit `0`, all pass;
- `go test ./internal/runtimecommandidempotency -run 'TestTrackedStart'
  -shuffle=on -count=50` with the same temporary environment → exit `0`;
- `go test ./internal/runtimecommandidempotency -shuffle=on -count=50` with
  the same temporary environment → exit `0`;
- `go test ./... -count=1` with the same temporary environment → exit `0`;
- `go vet ./...` with the same temporary environment → exit `0`;
- `CGO_ENABLED=1 go test -race ./internal/runtimecommandidempotency` with
  temporary cache → exit `1` before tests: `gcc` not found, race did not run;
- Git diff/status/name/HEAD inspection and `git diff --check` → exit `0`, only
  TASK-026 differs, HEAD `55821d4a0c17f5cce5658087b38ba3f15973b695`.

The race limitation is environmental, explicit, and non-blocking for this
Design-only reassessment. Focused and full-package shuffled `-count=50` runs
are the recorded concurrency substitutes; implementation verification must
attempt race again and retain `PASS WITH LIMITATION` if the environment remains
unchanged.

### Coordinator Readiness Decision

Coordinator accepts the Design-only reassessment boundary:
**`READY — UNBLOCK TASK-026`**. The previously missing prerequisite is now
implemented and independently accepted in isolation; no separate prerequisite
remains in the current 19-row matrix.

This decision is readiness only. It is not Coordinator Acceptance or
Completion of DP-016 implementation, does not change DP-016 Approved/Planned
status, and does not authorize production/test mutation under this contract.
The exact Implementation contract transition remains **`Not Activated`** until
a separate bare `Продолжай проект.` normal intake/resume transition records
Implementation mode, current scope, coverage gaps, Size Guard, and verification
plan. Independent review of this Design-only reassessment is the next stage.

### PROCESS-002 Readiness Synchronization

Verdict: **`Synchronized`**, subject to final independent review of the exact
post-synchronization content. The durable repository truth is now consistent:
fresh `READY — UNBLOCK TASK-026`, matrix 7 Direct / 10 Compositional / 2 Missing
core / 0 Missing prerequisite / 0 Missing external / 0 Deferred; DP-016 remains
Approved/Planned; the TASK-026 Implementation transition is `Not Activated`;
orchestrator code/tests, implementation Acceptance/Completion, and production
wiring remain absent.

#### Applicability and Exact Required Set

The exact changed set is **17 Required / 0 Questionable / 0 Removable**:

- this TASK-026 record and `docs/tasks/README.md`;
- `.ai/PROJECT_CONTEXT.md`, `spec/current-state.md`, and `spec/decisions.md`;
- MASTER_PLAN EN/RU;
- DP-015, DP-016, DP-019, DP-020, and DP-021 EN/RU.

Each file is Required to remove one live false Blocked/pending-reassessment
summary, preserve task discovery, or maintain mandatory EN/RU parity for the
same readiness decision. Historical blocker/closure statements remain where
their chronology is explicit. ARCH-004, DP-014, DP-017, DP-018, their indexes,
root README EN/RU, and `CHANGELOG.md` are Not applicable: no architecture
semantic/status, adjacent design boundary, user-facing/release capability, or
implemented behavior changed. Production, tests, modules, dependencies, and
generated artifacts remain absent and Not applicable.

#### Provisional Scope Audit and Size Guard

Provisional Scope Audit: **17 Required / 0 Questionable / 0 Removable**.
Deletion of any changed file would leave active-task navigation, a scoped live
readiness summary, or its required mirror false. No next-task implementation,
architecture change, unrelated refactor, generated artifact, formatting-only
change, or planned-as-implemented claim is present.

Size Guard triggers on more than 15 files. Decision: **`DO NOT SPLIT`**. The 17
files are one atomic documentation synchronization for one accepted readiness
fact; splitting would knowingly leave cross-document contradictions or EN/RU
drift. Production lines, packages, architecture contracts, and independently
shippable implementation behaviors changed: `0/0/0/0`.

The initial Reviewer `Approved 0/0` covers the pre-synchronization one-file
manifest only. PROCESS-002 changes invalidate it for the expanded exact subject;
final independent review, final scope confirmation, and Coordinator Design-only
closure remain unclaimed.

#### Synchronization Verification

- exact changed set: `17 expected / 17 actual`; production, test, module,
  dependency, generated, staged, and untracked files: `0`;
- changed EN/RU heading/fence parity: MASTER_PLAN `36/36`, `0/0`; DP-015
  `30/30`, `4/4`; DP-016 `30/30`, `4/4`; DP-019 `25/25`, `16/16`; DP-020
  `34/34`, `12/12`; DP-021 `21/21`, `10/10`;
- scoped relative links: `214 valid / 0 broken`;
- TASK-026 `task-record-v1` headings: one `## Status`, one `## Task Contract`,
  one terminal `## Recovery Evidence Envelope`, in order;
- scoped stale-status scan: `19` residual Blocked-form matches, all explicit
  historical task-closure facts or the conditional DP-021 fail-closed rule;
  live contradictions `0`;
- `git diff --check`: `PASS`; Git reports only CRLF conversion advisories, not
  whitespace errors.

### Current Implementation Architecture Confirmation

Independent Architect verdict: **`PASS`**. Active ARCH-004 §19(4), Approved
DP-014/DP-015/DP-016/DP-019, and the accepted isolated DP-020/DP-021 seams are
sufficient for this exact Implementation contract; no new design decision or
existing-seam change is required. The current proof matrix remains **7 Direct /
10 Compositional / 2 Missing core / 0 Missing prerequisite / 0 Missing external
/ 0 Deferred**. Proofs 2 and 14 are intentionally orchestrator-owned
same-target decisions.

Authorized shape: one new `internal/runtimeactivation` package with a validating
constructor, immutable exact-scope request, closed result categories, and exact
`ActivateExact`, `ReplaceExact`, and `RollbackExact` entry points. Package-local
interfaces may mirror only exact version `Get`, DP-014 read/history/conditional
publication, Owner `Observe`/`StopExpectedAttempt`, existing DP-015 managed
admission/callback capabilities, DP-021 `InvokeManagedStart`, and the exact
execution-generation source. `Unresolved` is non-terminal and must never be
published as a terminal command outcome.

Mandatory order is exact target lookup/Published and scope validation, coherent
DP-014/Owner observation, same-target zero mutation, managed primitive or parent
admission, proven old release and DP-014 terminal publication before Continue,
managed Start invocation, DP-014 Running/terminal publication, then DP-015
phase and parent terminal publication. Tracked Starting replacement must use
only the TASK-047 atomic parent plus preclaimed `StopOld` path. Rollback uses an
exact Published target and fresh attempt. Indeterminate reads, calls,
publication, panic, `runtime.Goexit`, expired capability, or generation loss
remain unresolved without retry or inferred terminal truth.

Architecture stop conditions: any Approved/Frozen contract change, missing
prerequisite, ambiguous ownership/order, inability to reconcile aggregate and
Owner facts, existing-seam/public-API/production-wiring need, inability to
publish DP-014 truth before DP-015 terminalization, assumed release, blind
retry, authority reuse, Host overlap, or Size Guard breach. Size Guard remains
one new package, at most 500 net production lines, at most 15 total task files,
and one independently deliverable behavior.

### Developer Implementation Handoff

Developer added only `internal/runtimeactivation/activation.go` and
`internal/runtimeactivation/activation_test.go`; no existing seam, public API,
production wiring, module, dependency, or documentation file was changed by
Developer. The package implements a validating constructor, immutable exact-
scope request, closed result categories, and `ActivateExact`, `ReplaceExact`,
and `RollbackExact`. It performs exact Published-version validation, coherent
DP-014/Owner observation, authorized same-target decisions, primitive or
managed-parent admission, exact expected-attempt Stop, proven-release
publication before Continue, fresh generation/managed Start, DP-014 terminal or
Running publication, and DP-015 phase/parent terminalization. `Unresolved`
never becomes a terminal durable outcome.

Developer focused, shuffled stress, affected-package, full, vet, module,
formatting, and diff checks passed. Initial production size 494 grew to 527
during authorization/coherent-observation correction; Developer stopped on the
guard and reduced only redundant comments/blank lines/accessors, preserving the
same behavior, to exactly **499 physical production lines** after `gofmt`.

### Independent Tester Handoff

Tester verdict: **`PASS WITH ENVIRONMENT LIMITATION`**, blocking findings `0`,
non-blocking findings `0`, production defects `0`. Tester changed only
`internal/runtimeactivation/activation_test.go`, now **584 lines**; production
remains 499 lines. All DP-016 §25 proofs 1–19 have executable new or retained
prerequisite evidence. New real-Host scenarios prove successful replacement and
rollback, same-address zero overlap, old release before fresh claim, exact
Published rollback with fresh attempt, same-target activation/rollback zero
mutation, replay, stop failure/cancellation fail-closed behavior, unresolved
barrier, startup failure without resurrection, binding-before-Load, and
different-Instance progress.

Checks: focused PASS; focused shuffled `-count=20` PASS; eight affected packages
shuffled `-count=5` PASS; full `go test ./... -count=1` PASS; `go vet ./...`
PASS; `go mod tidy -diff` PASS; scoped `gofmt -d` empty; `git diff --check`
PASS. Default race attempt requires CGO; explicit `CGO_ENABLED=1` race build
completed exit 1 because `gcc` is absent. Stress/full/vet substitutes pass, so
the verdict retains that exact environment limitation.

### Initial Independent Reviewer Handoff

Reviewer verdict: **`Needs Revision`**, blocking findings `6`, non-blocking
findings `0`. The blocking areas are: exact DP-015 replay is bypassed by
aggregate-derived preflight decisions; cancellation after proven old release
and definitive pre-Owner no-claim outcomes are not terminalized; generation is
allocated before authorization/admission or the cancellation gate; the claimed
nineteen-row proof set omits tracked-Starting replacement, post-release and
no-claim cancellation, post-continuation Stop, authorization/allocation order,
and shared-infrastructure different-Instance proof; all new exported
identifiers lack required GoDoc; and the persistent Recovery Evidence Envelope
was stale.

The reviewed 21-path subject used SHA-1 object format and unsigned UTF-8
path-byte ordering, with `task-record-v1` projection for this record. Its
canonical manifest was `37a6092603e7ba53b646fec23dfcf1f216de8bba`;
projected task-record blob `6c0edf570733d55aff18ec4f9f94826cffb14eac`,
production blob `b984df83494af1a4b6e679627da60e0a0e2ad217`, and test
blob `e1bc321f79599d71868692b5c37b8a2df7cb7212`. Those identities are
historical after this recovery correction and subsequent Developer rework.
Reviewer checks passed for focused tests (`-count=20`), full tests, vet,
formatting, module tidiness, and diff hygiene; race remained unavailable
because `gcc` is absent. PROCESS-002 and final Scope Audit remain forbidden
until rework, repeat independent verification, and repeat review succeed.

### Developer Rework Stop Handoff

Developer performed read-only composition analysis and stopped before mutation.
The current DP-015 managed-Start API requires an expected aggregate revision
and non-empty generation as eager arguments, validates them before
authorization, and only then inspects/replays/claims. Parent APIs likewise
combine inspection and claim. Consequently the Approved requirements for exact
replay before aggregate-derived same-target decisions and generation allocation
only after authorization, cancellation, and confirmed new claim cannot be
composed through the Architecture-confirmed existing seams. An authorized
inspect-only/admission-decision mechanism and late generation provider are
candidate additive seams, but Developer is forbidden to invent them. Repeat
Architecture Confirmation and Size Guard reevaluation are active; no code or
test byte was changed by this rework attempt.

### Repeat Architecture Confirmation and Split Decision

Independent Architect verdict: **`NEEDS DECISION`**. B-001 is confirmed:
current DP-015 APIs combine inspection and claim and cannot provide exact replay
before aggregate-derived same-target decisions. A bare inspect-only shortcut
would itself violate DP-015 atomic absent-intent claim semantics. B-003 is also
confirmed: eager non-empty generation is required before authorization/replay
by the current APIs, while moving generation after a winning claim changes the
Approved DP-020 binding timing and failure semantics.

The minimum correction is a distinct DP-015/DP-020 refinement owned by
`internal/runtimecommandidempotency`: replay-first authorized orchestration
admission, a closed orchestrator decision for absent intents, atomic recheck and
claim, and late generation allocation only after the winning primitive or
StartTarget-phase claim and cancellation winner. Provider failure, empty value,
panic, `runtime.Goexit`, replacement, or binding-install uncertainty must leave
unresolved durable truth, exhaust the capability, and never retry or call
Owner/Load. Exact replay, concurrency linearization, same-target semantics,
tracked-Starting replacement, pending/post-continuation Stop, cross-Instance
progress, and reconstruction require independent proof.

This refines Approved DP-015 behavior and Draft DP-020 binding timing, affects
an existing second production package, adds production lines beyond the task's
exact 499-line stop condition, and is independently deliverable. The earlier
21-path `DO NOT SPLIT` decision
does not waive it. Decision: **`SPLIT REQUIRED`**. Further runtimeactivation
changes, repeat implementation testing/review, PROCESS-002 completion, and
Acceptance are forbidden. TASK-026 must close Blocked with no implementation
diff; the prerequisite remains a separate Not Activated candidate with no Task
ID until a future intake.

### Implementation Size Guard Re-evaluation

The implementation attempt reached **21 files**: 19 Required documentation
paths plus two files in one new package. The repeat Architecture verdict then
proved that a correction would require a refinement of Approved DP-015 and
Draft DP-020, an existing second production package, additional lines beyond the
499-line stop boundary, and a separately deliverable prerequisite. Decision:
**`SPLIT REQUIRED`**. The rejected two-file implementation was removed. The
remaining exact subject is **19 documentation files**, still above the 15-file
review trigger; Coordinator decision for that evidence-only synchronization is
**`DO NOT SPLIT`** because the task/navigation/state documents and mandatory
EN/RU mirrors form one indivisible blocker handoff. This does not activate or
design the prerequisite and does not waive final verification, Scope Audit, or
independent review.

### Bounded Certification Process-Recovery Contract (2026-08-27)

This recovery cycle is evidence-only and may change process/documentation
contracts, not product behavior. Its sole objective is to resolve the three
certification defects found by the latest independent Reviewer without
changing TASK-026's blocker, status, implementation scope, or prerequisite
activation.

The minimum cross-cutting amendment scope is: `docs/engineering/PROCESS-001-
AI-DEVELOPMENT-WORKFLOW.md`, `docs/engineering/PROCESS-002-DOCUMENTATION-
SYNCHRONIZATION.md`, `docs/engineering/TASK-TEMPLATE.md`,
`docs/engineering/EXECUTION-INTERRUPTION-RECOVERY-ACCEPTANCE-SCENARIOS.md`,
`docs/engineering/agents/coordinator.md`, `docs/engineering/agents/tester.md`,
`docs/engineering/agents/reviewer.md`, `.ai/PROJECT_CONTEXT.md`, and this
task record. No production, test, module, dependency, generated, Publisher,
or next-prerequisite path is in scope.

The general process convention is now recorded by the bounded amendment to
PROCESS-001, PROCESS-002, TASK-TEMPLATE, the interruption scenarios, and the
Coordinator/Tester/Reviewer contracts. A canonical ordered subject manifest
is the only certification identity: the TASK record is included with
`task-record-v1` projection; every other present evidence path uses `full`;
deleted paths use baseline mode and OID `-`; paths sort by ascending unsigned
UTF-8 bytes; each row is `path\\0projection\\0state\\0mode\\0oid\\0`; and the
manifest OID is the Git blob identity of that complete NUL-separated stream.
`task-record-v1` excludes the Status evidence body and terminal Recovery
Evidence Envelope, so the tuple can be appended envelope-only without
changing the certified subject identity. Raw `git diff` is not a canonical
certification digest and is not replaced by an 18-file, normalized, or
projected-diff workaround.

The envelope must durably record the exact ordered path/projection set, rows,
repository/Task/branch/base/current HEAD, object format, manifest OID, blocker
identity, and a fresh Tester handoff containing exact commands, exit/results,
limitations, scope/coverage counts, reproducible proof artifacts, and tested
identity. Independent Reviewer and Scope Audit must verify that tuple; any
mutation of projected subject outside the envelope invalidates the subject and
all downstream gates. Status evidence body remains excluded from manifest
identity but requires separate status/contract reconciliation. This amendment
is process-only: Commit, Publisher, Coordinator
Acceptance, Completion, product implementation, and prerequisite activation
remain forbidden.

The ordered pipeline is: process amendment → PROCESS-002 → Verification
Matrix → durable Tester handoff → Independent Reviewer/rework as needed →
Scope Audit → recomputed projected manifest → envelope-only tuple append and
identity recheck → `Blocked Closure Certified`. The task remains Blocked, and
the prior non-certified digest is never reused.

### Process-Contract Amendment Documentation Handoff (2026-08-27)

Documentation Agent completed only the bounded process/documentation
amendment. The seven amended operational contracts are
`docs/engineering/PROCESS-001-AI-DEVELOPMENT-WORKFLOW.md`,
`docs/engineering/PROCESS-002-DOCUMENTATION-SYNCHRONIZATION.md`,
`docs/engineering/TASK-TEMPLATE.md`,
`docs/engineering/EXECUTION-INTERRUPTION-RECOVERY-ACCEPTANCE-SCENARIOS.md`,
`docs/engineering/agents/coordinator.md`,
`docs/engineering/agents/tester.md`, and
`docs/engineering/agents/reviewer.md`; `.ai/PROJECT_CONTEXT.md` and this task
record carry the durable governance/task facts. The resulting current
documentation subject is 26 paths: the prior 19 blocked-handoff paths plus
these seven operational contracts. `git diff --check` passes; no production,
test, module, dependency, generated, staged, or untracked path is present.
The amendment preserves TASK-026 `Blocked`, DP-016 `Approved/Planned`, the
unactivated prerequisite, and all permission/Acceptance/Completion gates.
Tester, Independent Reviewer, Scope Audit, canonical manifest calculation,
and `Blocked Closure Certified` remain pending and are not claimed by this
handoff.

### Blocked-Handoff PROCESS-002 Synchronization (2026-08-27)

Documentation Agent verdict: **`Synchronized for blocked-handoff
verification`**. The current repository truth is aligned across the exact 19
already-modified documentation paths: TASK-026 is Blocked by the missing
DP-015/DP-020 replay-first admission and late-generation refinement; the
rejected `internal/runtimeactivation` implementation/test diff is absent;
DP-015, DP-016, DP-019, DP-020, and DP-021 retain their existing Design and
Implementation Status values; no TASK-026 implementation, Acceptance, or
Completion is claimed.

The next candidate is a separate narrow design plus isolated implementation
prerequisite for DP-015/DP-020. It remains **`Not Activated`**, has no assigned
Task ID, and may be selected only by a later normal intake after this blocked
evidence reaches its required closure path. This synchronization neither
defines the new contract nor starts its implementation.

PROCESS-002 applicability:

- this task record and `docs/tasks/README.md`: **Required and updated** for the
  live Blocked status, chronology, recovery handoff, and next-candidate boundary;
- `.ai/PROJECT_CONTEXT.md`, `spec/current-state.md`, and `spec/decisions.md`:
  **Required and updated** for current task/capability/decision truth;
- MASTER_PLAN EN/RU and design indexes EN/RU: **Required and updated** for the
  durable dependency and navigation boundary;
- DP-015, DP-016, DP-019, DP-020, and DP-021 EN/RU: **Required and updated**
  only for the superseding blocker and missing-prerequisite wording; normative
  semantics and statuses are unchanged;
- ARCH/ADR, DP-014, DP-017, DP-018, root README EN/RU, `spec/README.md`, and
  `CHANGELOG.md`: **Not applicable** because no approved semantic, adjacent
  status, user-facing/release capability, or specification inventory changed;
- production code, tests, modules, dependencies, generated artifacts, and the
  next prerequisite: **Not applicable and absent**.

Documentation verification: exact changed set `19 expected / 19 actual`;
production, test, module, dependency, generated, staged, and untracked paths
`0`; changed EN/RU heading/fence parity aligned for DP-015 `30/30, 4/4`,
DP-016 `30/30, 4/4`, DP-019 `25/25, 16/16`, DP-020 `34/34, 12/12`, DP-021
`21/21, 10/10`, design indexes `1/1, 0/0`, and MASTER_PLAN `36/36, 0/0`;
changed-file relative links `258 valid / 0 broken`; live stale
In-Progress/Architecture-PASS activation matches `0`; conflict markers `0`;
task-record-v1 has one `## Status`, one `## Task Contract`, and one terminal
`## Recovery Evidence Envelope` in order; `git diff --check` PASS with only
line-ending advisories.

Final Tester, independent Reviewer, Scope Audit, canonical evidence identity,
`Blocked Closure Certified`, Coordinator closure, commit, and publication are
not claimed by this Documentation Agent handoff.

## Reactivation Task Contract (2026-08-25)

### Task Mode

`Implementation`.

### Why Now and Selection Evidence

- clean synchronized baseline `main@00b05a5` (`main == origin/main`), with no
  staged, unstaged, or untracked changes;
- no other active task exists;
- at intake, TASK-044 recorded the accepted Architect verdict `UNBLOCK TASK-026`, and
  TASK-045 synchronizes every live readiness summary;
- the task index, `.ai/PROJECT_CONTEXT.md`, `spec/current-state.md`,
  `spec/decisions.md`, and both MASTER_PLAN mirrors identify TASK-026 as the
  next Ready-to-Reactivate work;
- the retained historical branch points to ancestor `751577e` and contains no
  implementation diff, so reactivation uses a new branch rather than moving or
  rewriting that historical ref.

Rejected alternatives: another prerequisite slice is not selected because
TASK-044 proves that none remains; DP-017 recovery and DP-018 reporting
implementation are later dependency work and cannot precede the DP-016 core.

### Branch Decision

Work continues on
`feature/task-026-runtime-activation-orchestration-reactivation`, created from
the clean current `main@00b05a5`. This task-record update is the first content
change. The historical
`feature/task-026-runtime-activation-orchestration@751577e` ref is preserved
unchanged.

### Definition of Done

1. One bounded isolated orchestrator implements every unchanged DP-016 §25
   proof, including exact-version validation, zero-mutation satisfied outcomes,
   replacement/rollback sequencing, no Host overlap, truthful cancellation and
   indeterminate outcomes, exact callback publication, and DP-015
   terminalization.
2. The implementation composes only already accepted seams and does not change
   Approved/Frozen contracts or add production wiring.
3. Proof and regression tests cover all nineteen DP-016 §25 rows and the
   failure/cancellation/concurrency boundaries identified by TASK-044.
4. Required formatter, focused/full/stress/race-or-explicit-limitation, vet,
   module, smoke/proof, and diff checks pass.
5. PROCESS-002, full scope audit, independent final review, Coordinator
   Acceptance, and durable project-state synchronization complete without
   presenting planned external capabilities as implemented.

### Out of Scope

- production Control Service or HTTP/API wiring;
- external persistence, recovery/reconciliation, operational reporting or
  redaction;
- authorization policy, automatic rollback, multi-node supervision, or
  Production Activation;
- changes to Approved DP-014–DP-019 semantics or new public package contracts;
- stage, commit, push, PR, merge, publication, or branch cleanup.

### Verification Plan

- before any test edit, record an Existing Coverage Report mapping current
  package tests to all nineteen DP-016 §25 proofs and identifying exact gaps;
- run focused package tests plus shuffled and repeated stress checks for
  lifecycle, cancellation, rendezvous, shared-state, replay, and phase order;
- run `go test ./... -count=1`, `go test -race` for affected packages when the
  environment supports it (otherwise record `PASS WITH LIMITATION` and
  substitute stress evidence), `go vet ./...`, `go mod tidy` with zero
  unexplained module diff, `gofmt` checks, executable isolated proof/smoke
  scenarios, `git diff --check`, and conflict-marker/unexpected-file checks;
- require independent architecture confirmation before implementation,
  independent Tester verification, and an independent Reviewer verdict after
  PROCESS-002 and the Scope Audit.

### Roles and Stage Decisions

- Coordinator: intake, gates, Size Guard, scope audit, closure and acceptance;
- Documentation Agent: pre-implementation baseline and final PROCESS-002;
- Architect: explicit confirmation of TASK-044 constraints and all nineteen
  DP-016 §25 acceptance mappings; no new design is authorized;
- Developer: bounded implementation and proof tests after the coverage gate;
- Tester: independent verification and limitation accounting;
- Reviewer: independent review; the implementation author cannot perform the
  final review.

Pre-Implementation Documentation is provisionally `Not applicable`: the task
implements already Approved contracts and must stop if implementation requires
contract changes. This decision is rechecked after Architecture Analysis.

### Provisional Size Guard

The slice intentionally delivers one inseparable behavior: the DP-016
activation/replacement/rollback orchestrator whose terminal publication and
command finalization must be atomic with its ordered callbacks. TASK-044
classifies all remaining work as one coherent bounded core. Intake nevertheless
stops for re-scoping before implementation if the concrete plan exceeds 15
changed files, 500 net production lines, one new package, one new contract, or
reveals independently deliverable behavior without breaking the nineteen-proof
Definition of Done.

### Stop Conditions

Stop and return to Coordinator if architecture confirmation finds a missing
callable seam, Existing Coverage cannot map every §25 proof, implementation
would change an Approved/Frozen contract or public API, exact lifecycle/
authorization ownership is ambiguous, critical documentation drift exists, or
an obligatory verification/review gate fails.

### Initial Next Candidate — Superseded by Blocker

At intake, after hypothetical TASK-026 acceptance, the next dependency
candidate was a separate bounded readiness/intake for DP-017 recovery and
reconciliation implementation. The superseding blocker verdict prevents that
sequence; the current next candidate is the unactivated DP-015 conformance
prerequisite recorded below.

## Existing Coverage Report (2026-08-25, Pre-Test Gate)

Tester performed a read-only mapping before any production or test edit.
Focused baseline across `configurationversion`, `runtimeidentity`,
`runtimelifecycle`, `runtimecommandidempotency`,
`runtimeorchestrationbinding`, `runtimeorchestrationcontinuation`,
`runtimelaunchflow`, and `runtimemanagement` passed `8/8` packages. A separate
full baseline `go test ./... -count=1` also passed.

### Existing Coverage

The accepted package tests prove the prerequisite seams summarized by
TASK-044: **7 Direct / 10 Compositional / 2 Missing core / 0 Missing external
/ 0 Deferred**. Direct evidence exists for the Continue/pending-Stop gates,
indeterminate DP-015 barriers, exact generation binding before Load,
cross-Instance independence, and mirrored contract checks. Compositional
evidence exists for exact attempt claims, expected-attempt Stop, Host release,
fresh attempt identity, exact authorization, Owner outcomes, DP-014
publication primitives, and managed Flow ordering. The two missing-core rows
are the orchestrator-owned exact-Running/same-target zero-mutation satisfied
decisions (DP-016 §25 proofs 2 and 14).

No existing test proves the full callback -> exact Owner outcome -> DP-014
publication -> DP-015 primitive/phase/parent terminalization path.

### Coverage Gap

- pre-command exact Published-version validation through only the exact-ID
  repository `Get` seam, with zero mutation for zero/missing/mismatched/
  foreign/non-Published targets;
- initial activation, ordered replacement and explicit rollback composition;
- exact-Running/same-target satisfied and in-progress decisions;
- old release before Continue/new claim and zero Host overlap;
- callback-owned mapping of exact Start/Stop outcomes into ordered DP-014 and
  DP-015 terminal facts;
- end-to-end replay, failure, cancellation, concurrency and indeterminate
  publication behavior;
- one isolated executable smoke scenario for the complete bounded core.

### Required Added Proof and Regression Tests

The implementation must add a focused package test matrix covering: exact
Published initial activation and binding-before-Load; zero-mutation satisfied
activation/rollback; replacement release-before-fresh-claim; replacement
during Starting with Stop-first, pending-Stop and late-Stop winners; Stop
failure/unproven cleanup; startup failure without resurrection/automatic
rollback; exact Published rollback with fresh attempt; phase-by-phase
cancellation; indeterminate identity/command publication; exact outcome
mapping; same-key replay/conflict; stale revision/foreign scope; callback
panic/`runtime.Goexit`; many concurrent Stops; and independent Instances.
DP-016 §25 proof 19 remains a PROCESS-002 parity/status/link proof rather than
a Go test.

### Remaining Limitations

External durable persistence, process-restart recovery, DP-017 recovery
execution, DP-018 reporting, public API, production wiring and Production
Activation remain out of scope. Storage-client reconstruction is the only
available restart-like proof. Manual API/UI smoke is not applicable; the task
requires an executable isolated proof. Race evidence is mandatory when the
environment supports it; otherwise the result must be `PASS WITH LIMITATION`
with the exact failure and repeated/shuffled stress substitutes.

### Coverage-Based File Estimate

Expected scope is one new bounded package (`2–3` production files), `3–4`
proof/fixture files, no prerequisite-package changes, this record, and
approximately `7–9` final synchronization documents: projected `12–16` files.
Crossing 15 files, 500 net production lines, one new package, or any accepted
seam change reopens the Size Guard before implementation proceeds.

## Architecture Confirmation (2026-08-25)

Architect verdict: **`PASS — implementation may proceed after the
Documentation Baseline gate`**. The review confirms TASK-044 without creating
or changing an architecture decision: DP-016 remains Approved/Planned, every
prerequisite seam is present and accepted in isolation, and the remaining
callback/terminal/orchestrator behavior is one coherent TASK-026 core.

The unchanged §25 matrix is **7 Direct / 10 Compositional / 2 Missing core /
0 Missing external / 0 Deferred**. Proofs 2 and 14 are the only Missing-core
rows and belong to the orchestrator's authorized pre-claim, zero-mutation
same-target satisfied decisions. All other non-documentary rows must compose
the existing exact-ID repository read, DP-015 managed primitive/parent/phase
admission, DP-014 aggregate operations, DP-010 expected-attempt Stop, managed
continuation/Flow and TASK-043 invoker without weakening their contracts.

### Confirmed Ownership and Ordering

- `runtimelifecycle.Owner` remains the sole lifecycle and Host owner;
- `runtimeidentity.Store` owns aggregate revision, attempt/generation and
  Running/terminal facts;
- `runtimecommandidempotency.Boundary` owns authorization-at-submission,
  claims, permits, replay, rendezvous and unresolved admission;
- continuation, managed Flow and managed invoker retain their accepted narrow
  claim/bind, preparation-order and exact-scope responsibilities;
- the new orchestrator owns only exact reads/decisions, callback closures,
  exact Owner-outcome mapping and the order DP-014 facts precede DP-015
  terminalization.

Before command or lifecycle mutation, only
`ConfigurationVersionRepository.Get(exactID)` may be borrowed; requested and
returned IDs must match, Configuration ownership must match and state must be
exactly `Published`. Latest/current/list/fallback selection is forbidden.
Replacement must publish proven old release before Continue and new claim.
Managed Start must use the TASK-043 invoker; Stop must use
`StopExpectedAttempt`. Any indeterminate DP-014 truth leaves the linked DP-015
set unresolved. No command, aggregate or Owner lock may span lookup,
lifecycle work, waits or storage calls.

### Implementation Boundary and Size Decision

The approved implementation location is one new package,
`internal/runtimeactivation`, with one immutable orchestrator, a validating
constructor, closed exact activation/replacement/rollback entries, closed
result categories, and package-local narrow dependency interfaces. No generic
engine, registry, service locator, public API, recovery hook or lifecycle
abstraction is authorized.

Architect estimate is `2–3` production and `3–5` focused test files, with
`350–500` net production lines. Coordinator retains the provisional Size Guard:
one package and one inseparable behavior are accepted; exceeding 500 net
production lines, 15 total task files, or any existing seam/package contract
change stops work for re-evaluation. Splitting terminal publication from the
orchestration is not acceptable because it would break the DP-016 ordering
proof.

Pre-Implementation Architecture Documentation is **Not applicable** because
no contract changes. A separate PROCESS-002 baseline repair is required before
implementation because the Documentation Agent found factual live status
drift; that repair must not alter this architecture verdict.

### Size Guard Re-evaluation after Documentation Baseline

**Triggered by projected file count; decision: `DO NOT SPLIT`.** The
Documentation Agent identified approximately 25 mandatory final PROCESS-002
files because one factual implementation transition must update EN/RU mirrors,
design indexes, roadmap mirrors, task navigation and durable state sources.
Those files do not represent additional deliverable behaviors or contracts.
Splitting them from the implementation would leave the repository in known
documentation drift, while splitting callback mapping from terminal
publication would break DP-016 ordering and acceptance proofs.

The accepted integrity envelope remains exactly one new
`internal/runtimeactivation` package and one independently deliverable
activation/replacement/rollback behavior. No prerequisite package contract,
public API, production wiring or second new package may change. More than 500
net production lines, a second package/contract, or implementation behavior
not necessary for the nineteen proofs remains a stop-and-rescope condition.
Every final file must still pass the per-file Scope Audit deletion test; this
decision does not pre-approve the projected documentation set.

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

## TASK-038 and TASK-044 Readiness Reassessments

Accepted DP-020 Slices 1–3 now implement the original command authorization,
parent/phase, rendezvous, managed Flow/Owner-claim continuation, and DP-014
binding seams in isolation. Completed and Coordinator-Accepted TASK-038 records
the exact verdict
`TASK-026 REMAINS BLOCKED`: the atomic expected-attempt Owner Stop prerequisite
was then the first missing prerequisite; the later private exact-scope
composition invoker and subsequent readiness/orchestrator boundaries were also
absent at that historical verdict.

At TASK-038 closure, the first bounded prerequisite candidate was a separate
unactivated design-only DP-010 atomic expected-attempt Stop contract, and
TASK-038 did not finalize its API or change Approved sources. TASK-039
subsequently activated and recorded that accepted design in Draft DP-010.
Completed and Coordinator-Accepted TASK-040 implements and verifies the
extension in isolation, with repeat final Reviewer `APPROVED` 0/0. TASK-026
then remained Blocked by the later private exact-scope composition invoker and
subsequent readiness/orchestrator boundaries. Completed and Coordinator-
Accepted TASK-043 subsequently implements the invoker in isolation.
Post-Owner DP-014 terminal publication and DP-015 command/phase terminalization
remain core TASK-026 orchestrator work, not a separate prerequisite. External
API, persistence, recovery, reporting, production wiring, and Production
Activation remain outside the bounded reassessment.

TASK-044 remaps every unchanged DP-016 §25 proof after TASK-040 and TASK-043:
**7 Direct / 10 Compositional / 2 Missing core / 0 Missing external / 0
Deferred**. Its historical Architect verdict is `UNBLOCK TASK-026`. The two Missing-core
rows are TASK-026-owned zero-mutation satisfied decisions, while the ten
Compositional rows have all required callable seams and remain end-to-end
TASK-026 implementation/proof work. No separate prerequisite remains.
TASK-044 completed with Coordinator Acceptance (2026-08-24), repeat Reviewer
`APPROVED` 0/0, Scope Audit 16/0/0, and PROCESS-002 Synchronized.

At TASK-044 closure the status was `Ready to Reactivate — Not Activated`. No
automatic branch resume, implementation, test edit, acceptance, commit, or
publication occurred at that historical checkpoint.

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

Historical TASK-027 handoff is satisfied by the subsequently accepted
prerequisite sequence through TASK-043 and the TASK-044 readiness verdict.
TASK-026 may be reactivated only by a later explicit Coordinator task
selection after accepted TASK-044; this record does not perform that selection.

## Closure

- Historical closure status: `Blocked by Architecture`;
- Historical readiness status after TASK-044 architecture handoff: `Ready to
  Reactivate — Not Activated`;
- TASK-026 Coordinator Acceptance: not performed;
- commit: not performed and forbidden;
- publication: not performed and forbidden.

## Documentation Baseline and Pre-Implementation Drift Repair (2026-08-25)

Documentation Agent completed the baseline before implementation. The
public/project mirror inventory is `46 EN / 46 RU` with no unmatched paths,
scoped ARCH/DP headings and code fences are aligned, and scoped links resolve.
Planned and implemented state remain separated: TASK-043 implements the
concrete exact-scope composition-private invoker in isolation, while the
DP-016 callback/terminal orchestrator and production wiring remain absent.

The baseline found critical live factual drift in mirrored Draft DP-011 and
Approved DP-015: both still described the accepted TASK-043 concrete invoker
as Planned or absent. The bounded repair updates only those live
implementation-boundary statements in EN/RU. It preserves every Design and
Implementation Status, historical chronology, Approved semantic, and the
Planned TASK-026 orchestrator/terminal boundary. It makes no architecture
decision, production/test change, TASK-026 implementation claim, or final
project-state synchronization.

Earlier `Ready to Reactivate — Not Activated` and blocked-state statements in
the Original Task Contract, historical handoffs, and Closure above describe
their recorded historical baselines. The current task status and authority are
defined by the top-level Reactivation Task Contract.

Documentation Agent handoff: critical pre-implementation drift is repaired;
verification covers EN/RU semantic parity, headings/fences, scoped links, live
contradiction search, exact file set, and `git diff --check`. Final PROCESS-002
applicability and implementation-status/project-state synchronization remain
required after verified implementation and are not performed here.

## Reactivation Architecture Blocking Discovery (2026-08-25)

The initial reactivation Architecture Confirmation `PASS` is **superseded** by
a repeat Architect verdict: **`CONFIRMED BLOCKER`**. Developer inspection and
an independent read-only recheck proved that the current DP-015 managed-parent
surface does not implement one mandatory contract already stated by DP-015,
DP-016 and DP-019.

### Reproduction Evidence

- `runtimecommandidempotency.Boundary.ExecuteManagedParent` rejects a new
  parent whenever `ledger.hasAnyNonterminalLocked()` is true;
- a tracked primitive Start remains a non-terminal record throughout its
  callback and Starting lifetime;
- the existing `mayClaimLocked` tracked-Start exception is used by primitive
  admission and permits only an independent `OperationStop`;
- managed parent admission neither uses that exception nor atomically claims a
  replacement/rollback parent with its derived ordinal-zero `StopOld` phase;
- existing rendezvous tests begin after parent acceptance and therefore do not
  prove DP-016 §13 / §25 proof 5 admission over a tracked primitive Start.

An independent primitive Stop is not a valid workaround: it mutates lifecycle
before the durable parent claim, has a different identity/intent, introduces a
race before later parent admission, cannot truthfully replay as the parent's
`StopOld` phase, and loses the required atomic winner/callback custody.

### Corrected Readiness Matrix

The superseding matrix is **7 Direct / 9 Compositional / 2 Missing core / 1
Missing prerequisite / 0 Deferred**. Proof 5 moves from Compositional to
Missing prerequisite. Proofs 3, 4 and 6 retain their downstream compositional
classification but cannot be completed until proof 5 admission exists.

### Developer and Size Guard Handoff

Developer built an exploratory one-package partial composition and confirmed
focused validation, exact activation/replacement/rollback/replay, Stop
publication and Instance-isolation paths. It could not cover replacement while
the old primitive Start is tracked. The partial production envelope also
reached 548 added lines and triggered the independent 500-line Size Guard.
Full, race, vet, module and final verification were not completed after the
blocker.

The exploratory `internal/runtimeactivation` production/test files were
classified `Removable` and removed by Coordinator: blocker evidence is wholly
in the existing command-admission code and Approved contracts. The temporary
workspace-local Go build cache was also removed. No prerequisite production or
test file was changed, and no generated residue is retained.

### Smallest Required Prerequisite — Not Activated

The next candidate is a separate bounded DP-015 conformance extension for
atomic tracked primitive Start -> replacement/rollback managed parent plus
preclaimed `StopOld` admission. It must define callback-scoped consumption of
the already-issued phase permit, competing independent-Stop winner semantics,
replay without capability, panic/`runtime.Goexit`/generation-loss unresolved
behavior, ordinary admission non-regression and different-Instance
independence. Exact capability shape must be recorded before implementation;
Approved DP-015/DP-016 semantics need no weakening.

This prerequisite is recommended but **not activated**. TASK-026
implementation, test expansion, Coordinator Acceptance, commit and publication
are forbidden while this blocker remains.

## Coordinator Scope Audit and PROCESS-002 Applicability (2026-08-25)

Full diff classification: **21 Required / 0 Questionable / 0 Removable** after
removal of the exploratory package and temporary cache.

- TASK-026 record and task index: Required for the reactivation contract,
  superseding blocker evidence, truthful current status and next-candidate
  boundary;
- DP-011 and DP-015 EN/RU: Required to repair the pre-existing stale TASK-043
  invoker status and record the exact remaining DP-015 prerequisite;
- DP-016, DP-019, DP-020 and DP-021 EN/RU: Required to supersede live
  Ready/no-prerequisite wording, preserve statuses, and record the corrected
  matrix and dependency boundary;
- design indexes EN/RU: Required navigation/live status summaries for those
  mirrored proposals;
- MASTER_PLAN EN/RU: Required durable current-milestone dependency correction;
- `.ai/PROJECT_CONTEXT.md`, `spec/current-state.md`, and
  `spec/decisions.md`: Required durable current task, capability gap,
  superseded readiness verdict and next recommendation;
- root README EN/RU: Not applicable — no user-facing or production capability
  changed;
- `CHANGELOG.md`: Not applicable — no release/user-facing change;
- `spec/README.md`, ADR/ARCH documents and indexes: Not applicable — no
  specification set, Approved/Frozen architecture semantic or status changed;
- production, tests, modules, dependencies and generated artifacts: zero final
  files; Not applicable after Removable exploratory/generated cleanup.

Deletion test: removing any of the 21 files would leave task navigation,
mandatory EN/RU parity, a live false Ready/no-prerequisite claim, or a durable
project-state contradiction. No next-task content, prerequisite implementation,
public API, production wiring or unrelated refactor is present.

Size Guard remains triggered by documentation file count and resolved `DO NOT
SPLIT`: the exact set is mandatory synchronization for one blocker verdict,
not multiple independently deliverable behaviors. `git diff --check`, conflict
marker scan, exact file-set check and unexpected/untracked-file check pass.
PROCESS-002 verdict: **`Synchronized`**.

## Final Blocked-Handoff Verification and Review (2026-08-25)

Independent Tester verdict: **`PASS`**, blocking/non-blocking/limitations
`0/0/0`. Exact diff is `21/21` documentation files with missing/extra,
staged/untracked and production/test/module/generated files all `0`. Focused
eight-package baseline, `go test ./... -count=1`, `go vet ./...`,
`go mod tidy -diff`, repository-wide `gofmt -d`, focused tracked-Start/
managed-parent semantic tests, mirror inventory `46/46`, changed-pair
headings/fences, links `277/0`, live status assertions, conflict scan and
`git diff --check` all pass. Race is Not applicable to the final
documentation-only diff; no environment limitation remains.

Initial final Reviewer verdict was `Needs Revision` with one blocking hygiene
finding: the exploratory package files were gone, but their empty directory
remained invisible to Git. Coordinator verified the path was inside the
workspace and empty, removed it, and repeated the hygiene gate.

Repeat Independent Reviewer verdict: **`Approved`**, blocking/non-blocking
`0/0`. The Reviewer confirms the superseding blocker, corrected matrix,
historical/live separation, unchanged statuses, next prerequisite Not
Activated/no Task ID, root README and CHANGELOG non-applicability, absence of
hidden architecture/API/implementation/wiring/next-task work, and exact
documentation verification. Deletion-test answer: all 21 changes are Required;
none can be removed while preserving the blocked-handoff Definition of Done.

## Coordinator Blocked Closure

- Final status: **`Blocked by missing command-admission prerequisite
  (2026-08-25)`**;
- Architecture Analysis: initial PASS superseded; repeat verdict
  **`CONFIRMED BLOCKER`**;
- corrected readiness: **7 Direct / 9 Compositional / 2 Missing core / 1
  Missing prerequisite / 0 Deferred**;
- implementation acceptance: not performed and forbidden; final production,
  test, module, dependency and generated diff is zero;
- Size Guard: documentation-count `DO NOT SPLIT` accepted for one synchronized
  blocker verdict; exploratory production overage removed;
- PROCESS-002: **`Synchronized`**;
- Scope Audit: **21 Required / 0 Questionable / 0 Removable**;
- Independent Tester: **`PASS`**, `0/0/0`;
- repeat Independent Reviewer: **`Approved`**, `0/0`;
- Coordinator Acceptance: **not performed** because TASK-026 is Blocked;
- commit, push, PR, merge, publication and branch cleanup: not authorized and
  not performed;
- next recommended work: separate bounded DP-015 tracked-Start managed-parent
  plus preclaimed `StopOld` conformance prerequisite; **Not Activated**, no
  Task ID assigned.

## PROCESS-002 Blocker Synchronization (2026-08-25)

Documentation Agent synchronized the superseding `CONFIRMED BLOCKER` verdict
without changing Approved semantics or any Design/Implementation Status. Live
task navigation, project context/current state/decisions, mirrored roadmap,
design indexes, and DP-015/DP-016/DP-019/DP-020/DP-021 mirrors now distinguish
the historical TASK-044 `UNBLOCK TASK-026` checkpoint from the corrected 7
Direct / 9 Compositional / 2 Missing core / 1 Missing prerequisite / 0 Deferred
matrix. Historical task chronology remains intact.

The next candidate is one separate bounded DP-015 tracked-Start managed-parent
plus preclaimed `StopOld` admission conformance prerequisite. It is **Not
Activated** and has no assigned Task ID. TASK-026 remains Blocked;
TASK-043 remains implemented in isolation; DP-016 remains Approved/Planned.

Applicability: this task record, task index, `.ai/PROJECT_CONTEXT.md`,
`spec/current-state.md`, `spec/decisions.md`, MASTER_PLAN EN/RU, design indexes
EN/RU, and affected DP mirrors are Required and updated. Root README EN/RU,
`CHANGELOG.md`, ADR/ARCH, release notes, production code, tests, modules,
dependencies, and generated artifacts are Not applicable because no
user-facing/release capability, architecture contract, executable behavior, or
dependency changed. PROCESS-002 verdict: **Synchronized**, subject to the
verification evidence reported by Documentation Agent.

Documentation verification: exact changed set `21 expected / 21 actual`, with
production/test/generated files `0`; public/project mirrors `46 EN / 46 RU`
with unmatched paths `0/0`; all changed DP, design-index, and MASTER_PLAN
heading/fence counts aligned; changed-file relative links `277 valid / 0
broken`; all live TASK-044 `UNBLOCK` occurrences are historical/superseded and
the current blocker/next-candidate boundary is present; conflict markers `0`;
`git diff --check` PASS. Root README and CHANGELOG remain unchanged as Not
applicable.

### Historical Design-only Recovery Evidence Envelope (published readiness cycle)

- projection: `task-record-v1`; this is the final top-level heading and its
  bytes through EOF are excluded from the projected evidence subject;
- repository: `E:\wikiPRJ\universal-websocket-platform`;
- Task/status: `TASK-026` / `Ready for Implementation — Design-only
  reassessment Coordinator Accepted (2026-08-26); Implementation Not
  Activated`;
- branch: `feature/task-026-runtime-activation-orchestration-resume`;
- anchor baseline/HEAD before the first mutation:
  `55821d4a0c17f5cce5658087b38ba3f15973b695`;
- evidence subject at intake:
  `docs/tasks/TASK-026-RUNTIME-ACTIVATION-REPLACEMENT-ROLLBACK.md` with
  projection `task-record-v1`; no canonical subject-manifest OID or role verdict
  is claimed before post-mutation verification;
- exclusions: this envelope, status-evidence body per `task-record-v1`, staging,
  commit/publication state, chat history, and any unrecorded command output;
- file mutation reconciliation: baseline was clean and this task record is the
  first and only intended content change at intake; exact diff scope is one
  task-record file and `git diff --check` is `PASS`;
- Documentation Baseline evidence: `26 present / 0 missing`, `10/10` EN/RU
  pairs with aligned heading/fence counts, scoped links `294 valid / 0 broken`,
  critical drift `0`, bounded intake-status drift pending post-verdict
  PROCESS-002;
- Architect evidence: fresh `READY — UNBLOCK TASK-026`, matrix `7 Direct / 10
  Compositional / 2 Missing core / 0 Missing prerequisite / 0 Missing external /
  0 Deferred`, exact ownership/order and one-package implementation boundary;
- Tester evidence: `PASS WITH ENVIRONMENT LIMITATION`, findings `0/0`; focused,
  stress, full, and vet checks pass after workspace-local cache reconciliation;
  race did not run because `gcc` is missing;
- Coordinator readiness: Design-only READY boundary accepted; no implementation
  Acceptance/Completion and Implementation transition remains `Not Activated`;
- initial Independent Reviewer verdict: **`Approved`**, blocking findings `0`,
  non-blocking findings `0`; reviewed subject is the single TASK-026
  `task-record-v1` projection, mode `100644`, raw/projected bytes
  `57847/54210`, projected blob
  `6ff74067d4260bca5b70d0a79a7b8d3a7ecea934`, canonical subject-manifest
  `b41d567c3c9290330dd76c7eb7546d7a130a9697`, object format `SHA-1`;
- initial Reviewer subject invalidation: subsequent PROCESS-002 changes expand
  the exact subject to 17 files; the initial verdict remains historical
  evidence and does not attest the synchronized bytes;
- PROCESS-002 evidence: 17 Required files synchronize fresh READY matrix
  7/10/2/0/0/0 while preserving DP-016 Approved/Planned and implementation Not
  Activated; provisional Scope Audit `17/0/0`, Size Guard `DO NOT SPLIT`;
- synchronization verification: exact set `17/17`, changed EN/RU heading/fence
  parity aligned, scoped links `214/0`, live status contradictions `0`,
  `task-record-v1` heading order valid, and `git diff --check` `PASS`;
- stage/commit/push/PR/merge/branch deletion: `Proven Not Started` and not
  authorized;
- final Independent Reviewer verdict: **`Approved`**, blocking findings `0`,
  non-blocking findings `0`, re-attested for the exact current synchronized
  17-path subject bound to corrected task projection
  `be6d5e0d66a8621594dd6ba82390e7ca66730ea1` and canonical manifest
  `1462e5e5d775fef2e3d29841f6561a1a0fd13772`;
- final Scope Audit: **`17 Required / 0 Questionable / 0 Removable`**;
- Size Guard: **`DO NOT SPLIT`** for one atomic readiness synchronization and
  mandatory EN/RU mirrors;
- PROCESS-002: **`Synchronized`**;
- Coordinator Acceptance: **accepted only the Design-only readiness
  reassessment** on 2026-08-26; DP-016 implementation Acceptance/Completion is
  not claimed and implementation remains `Not Activated`;
- completed checkpoints: Recovery Reconstruction, Task Intake, Documentation
  Baseline, Architect readiness recheck, Existing Coverage verification,
  Coordinator readiness decision, initial review, PROCESS-002 synchronization,
  final Scope Audit, final Independent Review, and Design-only Coordinator
  Acceptance/closure;
- first checkpoint without proven completion: none; Design-only reassessment
  closure is terminal and the active cycle stops;
- permission state: the bare continuation authority for this exact Design-only
  cycle is consumed by terminal closure; commit and publication are not
  authorized;
- commit: `not authorized / not created`;
- next candidate: exact TASK-026 Implementation contract transition after a
  separate bare continuation/intake transition; `Not Activated`;
- final evidence anchor: HEAD
  `55821d4a0c17f5cce5658087b38ba3f15973b695`, object format `SHA-1`, canonical
  17-path subject-manifest `1462e5e5d775fef2e3d29841f6561a1a0fd13772`,
  TASK-026 `task-record-v1` projected blob
  `be6d5e0d66a8621594dd6ba82390e7ca66730ea1`, raw/projected bytes
  `64504/57988`;
- superseded invalid calculator evidence:
  `77601536839934edfe42d7cead6fc9489866b15c` /
  `207f781b6b371a299336e6a95a4001e2ba827d10`; the invalid calculator used
  UTF-16 character offsets as UTF-8 byte offsets and is not evidence for the
  current subject;
- exact subject paths in ascending unsigned UTF-8 path-byte order:
  `.ai/PROJECT_CONTEXT.md`,
  `docs/en/design/DP-015-runtime-management-command-idempotency.md`,
  `docs/en/design/DP-016-runtime-activation-replacement-rollback.md`,
  `docs/en/design/DP-019-runtime-activation-orchestration-prerequisites.md`,
  `docs/en/design/DP-020-runtime-orchestration-binding-sequence-readiness.md`,
  `docs/en/design/DP-021-private-exact-scope-managed-start-invoker.md`,
  `docs/en/roadmap/MASTER_PLAN.md`,
  `docs/ru/design/DP-015-runtime-management-command-idempotency.md`,
  `docs/ru/design/DP-016-runtime-activation-replacement-rollback.md`,
  `docs/ru/design/DP-019-runtime-activation-orchestration-prerequisites.md`,
  `docs/ru/design/DP-020-runtime-orchestration-binding-sequence-readiness.md`,
  `docs/ru/design/DP-021-private-exact-scope-managed-start-invoker.md`,
  `docs/ru/roadmap/MASTER_PLAN.md`, `docs/tasks/README.md`,
  `docs/tasks/TASK-026-RUNTIME-ACTIVATION-REPLACEMENT-ROLLBACK.md`,
  `spec/current-state.md`, `spec/decisions.md`;
- recovery handoff: this Design-only task is terminal. A new agent may begin
  only the exact Not Activated Implementation candidate through a separate bare
  continuation/intake transition and fresh recovery reconstruction.

## Recovery Evidence Envelope

- projection: `task-record-v1`; this is the final top-level heading and its
  bytes through EOF are excluded from the projected evidence subject;
- repository: `E:\wikiPRJ\universal-websocket-platform`;
- Task/status: `TASK-026` / `Blocked — missing DP-015/DP-020
  orchestration-admission refinement (2026-08-27)`;
- branch: `feature/task-026-runtime-activation-orchestration-implementation`;
- trusted baseline and current HEAD before the first intake mutation:
  `65af65c3154d70db10de52483149960a42dcfb9a`;
- current exact subject: the 19 Required documentation paths recorded in the
  current Documentation Baseline; production, test, module, dependency,
  generated, staged, and untracked paths are absent;
- recovery reconstruction and Implementation intake: `Proven Completed`;
  synchronized `main`, accepted/published readiness history, exact branch,
  attributed first task-record change, and absence of competing work were
  inspected read-only;
- Documentation Baseline and pre-implementation synchronization: `Proven
  Completed`; initial verdict `Drift Detected — Critical`, repaired verdict
  `Synchronized for Architecture Confirmation`; inventory `26/26`, public
  mirrors `46/46` with unmatched `0/0`, scoped pair structure aligned, links
  `294/0`, and exact applicability `19 Required / 0 Questionable / 0
  Removable`;
- Architecture Confirmation: `Proven Completed`; independent Architect verdict
  `PASS`, one-package boundary, exact ordering, proof mapping, Size Guard, and
  stop conditions are recorded in the current Task Contract;
- initial Developer implementation and Tester verification: `Proven Completed`;
  their handoffs are historical inputs, not acceptance evidence, because the
  subsequent independent review found six blocking defects;
- initial Independent Review: `Proven Completed`, verdict `Needs Revision`
  `6/0`, reviewed canonical manifest
  `37a6092603e7ba53b646fec23dfcf1f216de8bba`; its exact findings and identity
  are recorded in the handoff above;
- initial Developer rework attempt: `Proven Stopped Before Mutation`; existing
  seams cannot satisfy B-001/B-003 and the exact stop handoff is recorded above;
- repeat Architecture Confirmation: `Proven Completed`, verdict `NEEDS
  DECISION`; `SPLIT REQUIRED` because the minimum correction changes Approved
  DP-015 behavior and Draft DP-020 binding timing, touches a second production
  package, and breaches the
  500-line stop condition;
- rejected implementation cleanup: `Proven Completed`; the two
  `internal/runtimeactivation` files and their empty directory are absent, and
  production/test diff is zero;
- blocked-handoff PROCESS-002: `Proven Completed`; exact applicability and
  the synchronized 19-path handoff are recorded above;
- final Tester verification: `Proven Completed` by cleanup and documentation
  closure checks; race is not applicable because production/test diff is zero;
- prior independent final Review and Scope Audit: `Proven Completed` for the
  prior envelope bytes, verdict `Approved` `0/0`, Scope Audit `19/0/0`,
  canonical manifest `6332ae0cc8edfc8a2ddf0e0851a08d25280d939b`, projected
  task-record blob `fb87e0bab803f0349157ccacb4f7b8957b7383a6`; that evidence is
  superseded by the current revalidation and does not certify current bytes;
- previously recorded `Blocked Closure Certified` / Coordinator closure is
  superseded for the current revalidation and must not be treated as a current
  certification; no ordinary or blocked Coordinator Acceptance of TASK-026 is
  performed;
- status boundary: DP-016 remains Design Status `Approved`, Implementation
  Status `Planned`; implementation behavior, Acceptance, and Completion are not
  claimed by this synchronization;
- exclusions: historical status evidence, the envelope itself, chat-only
  assertions, staging, commit/publication authority, and unrecorded command
  output;
- stage/commit/push/PR/merge/branch deletion: not authorized and not performed;
- current-cycle next candidate: separate narrow DP-015/DP-020 design plus
  isolated implementation prerequisite for replay-first orchestration
  admission and late generation allocation; `Not Activated`, no Task ID
  assigned; production composition/integration remains later and outside this
  task;
- current handoff: TASK-026 remains Blocked; Coordinator Acceptance and Blocked
  Closure Certification are not performed for the current bytes. No
  implementation Acceptance or Completion may be claimed. A future normal
  intake must first design and independently accept the separate DP-015/DP-020
  replay-first/late-generation prerequisite.
- latest Blocked Closure Certification revalidation (2026-08-27): **`Not
  Certified — STOP`**. Independent Tester returned `PASS WITH LIMITATION`; the
  subsequent independent Reviewer returned `Needs Revision`, blocking findings
  `3`, non-blocking findings `0`. Blocking findings are: the raw evidence-diff
  digest/command/object-format tuple is absent; the current envelope does not
  enumerate the exact 19 paths; and the mandated raw digest includes this task
  record, so appending the missing tuple would change the digest and violate
  the non-self-attestation rule. The current Tester handoff is not durably
  recorded with exact current identity and commands, and `Coordinator
  Acceptance: not performed` is not explicit in the prior summary.
- observed (not certified) raw diff digest for the pre-revalidation bytes:
  `ae8935eb3e2d643f3ad34d3353187765062d973f`; it must not be reused after this
  envelope append, replaced by an 18-file digest, or substituted with a
  projected/normalized diff. The current exact 19-path subject-manifest and
  projected blob remain historical observations only until a non-self-referential
  convention is resolved and the affected Tester/Reviewer gates are rerun.
- certification eligibility is therefore **not confirmed**. No baseline
  cleanup, blocker-scope change, local implementation, stage, commit, push,
  publication, ordinary Acceptance, or prerequisite activation is authorized;
  the next permitted step is process clarification or a future recovery cycle
  that reconstructs this changed envelope, reconciles identity, and reruns the
  affected Tester → Independent Reviewer → Scope Audit → digest gates.

### Certification Tuple Draft Superseded (2026-08-27)

The tuple below was drafted before the durable Tester handoff was appended to
this terminal envelope. It is retained as historical recovery evidence only and
is **not** a current `Blocked Closure Certified` result. The final certification
tuple appears after the durable handoff and its identity recheck. Certification
does not change TASK-026's status, does not accept DP-016 implementation, and
does not activate the prerequisite.

Immutable certification tuple:

- repository: `E:\wikiPRJ\universal-websocket-platform`;
- Task/status: `TASK-026` / `Blocked — missing DP-015/DP-020
  orchestration-admission refinement`;
- branch: `feature/task-026-runtime-activation-orchestration-implementation`;
- base branch/OID: `main` /
  `65af65c3154d70db10de52483149960a42dcfb9a`;
- certified HEAD/OID: `65af65c3154d70db10de52483149960a42dcfb9a`;
- exact evidence set: the 26 present `100644` documentation paths listed in
  the durable Tester handoff above, in ascending unsigned UTF-8 path-byte
  order; the TASK record uses `task-record-v1`, all other present paths use
  `full`, and there are no deleted paths;
- canonical identity: object format `sha1`; task-record projection
  `83,477` bytes, Git blob
  `feabbed62af1c957bfde57becc16186d94a76548`; ordered NUL-separated manifest
  `2,873` bytes, Git blob
  `161a6d10f0355dc313b2290d1111bc1f75b2ad4e`;
- manifest command/semantics: build exact rows
  `path\\0projection\\0state\\0mode\\0oid\\0` using raw current bytes for
  `full`, the exact `task-record-v1` projection for the TASK record, and
  unsigned UTF-8 path order, then hash the complete stream with
  `git hash-object --stdin`; terminal envelope bytes are excluded by the
  projection and do not self-attest;
- blocker identity: repeat Architecture `NEEDS DECISION / SPLIT REQUIRED` —
  current DP-015/DP-020 admission cannot provide replay-first inspection and
  late generation allocation; the separate refinement remains `Not Activated`
  with no Task ID;
- Verification Matrix/PROCESS-002: Tester `PASS`, exact 26/26 subject,
  commands/results and limitation durable above; process amendment and
  R-029/R-030 checks pass;
- independent Reviewer: `Approved`, blocking/non-blocking findings `0/0`;
  Scope Audit `26 Required / 0 Questionable / 0 Removable`;
- repository residue: production/test/module/dependency/generated/temp/
  staged/untracked paths `0`; `internal/runtimeactivation` absent;
- prior observed raw diff digest `ae8935eb3e2d643f3ad34d3353187765062d973f`
  is not part of this tuple and was not reused;
- Coordinator Acceptance: **not performed**; Completion and product
  Definition of Done are **not claimed**;
- commit, Blocked Evidence Checkpoint, push, PR, merge, publication and
  prerequisite activation: **not authorized / not performed**;
- certification class: superseded evidence-only tuple draft; it must not be
  used as certification. Any mutation outside this terminal envelope, any
  change to the exact set or tuple, or any change to the projected manifest
  requires the affected gates to be rerun.

### Current Bounded Recovery Tester Handoff (2026-08-27)

- role/verdict: **Tester — `PASS`** for the current evidence-only bounded
  recovery subject. This is a verification handoff only; it is not Scope Audit,
  Independent Review, Coordinator Acceptance, or `Blocked Closure Certified`.
- repository: `E:\wikiPRJ\universal-websocket-platform`;
  Task/status: `TASK-026` / `Blocked — missing DP-015/DP-020
  orchestration-admission refinement (2026-08-27)`;
  branch: `feature/task-026-runtime-activation-orchestration-implementation`;
  trusted base `main` and current `HEAD` before this envelope append:
  `65af65c3154d70db10de52483149960a42dcfb9a` / `65af65c3154d70db10de52483149960a42dcfb9a`.
- exact tested subject: 26 present documentation paths below, with the task
  record represented by `task-record-v1` and every other present path by
  `full`. Missing paths, extra paths, production code, test code, module or
  dependency files, generated/temporary files, staged paths, and untracked
  paths are all `0`. The exact diff is `26 expected / 26 actual`, with
  `0 missing / 0 extra`.
- canonical identity calculation: all path comparisons use ascending unsigned
  UTF-8 bytes (case-sensitive, locale-independent); present full bytes are
  hashed with `git hash-object --no-filters`, the task record projection is
  hashed with `git hash-object --stdin`, and the complete NUL-separated rows
  below are hashed with `git hash-object --stdin`. The task-record projection
  uses the exact PROCESS-001 `task-record-v1` rule: raw bytes through the
  `## Status` line, `STATUS-EVIDENCE-EXCLUDED`, one NUL, then raw bytes from
  `## Task Contract` through the byte before `## Recovery Evidence Envelope`.
  Object format is `sha1`; projected task-record bytes/OID are
  `83477` / `feabbed62af1c957bfde57becc16186d94a76548`; canonical manifest
  bytes/OID are `2873` / `161a6d10f0355dc313b2290d1111bc1f75b2ad4e`.
- exact ordered canonical rows (`<NUL>` denotes one NUL byte):

  ```text
  .ai/PROJECT_CONTEXT.md<NUL>full<NUL>present<NUL>100644<NUL>69e295cff446489a29aea4dd631e76c7bd17e867<NUL>
  docs/en/design/DP-015-runtime-management-command-idempotency.md<NUL>full<NUL>present<NUL>100644<NUL>caf0fec1e6b4ae28ca3523948028cea520849236<NUL>
  docs/en/design/DP-016-runtime-activation-replacement-rollback.md<NUL>full<NUL>present<NUL>100644<NUL>7c752d8b37c51445b3ae264f3aaefc2e6384768c<NUL>
  docs/en/design/DP-019-runtime-activation-orchestration-prerequisites.md<NUL>full<NUL>present<NUL>100644<NUL>1091d956bd8938ee39d2e040aa4ffa5aae04096e<NUL>
  docs/en/design/DP-020-runtime-orchestration-binding-sequence-readiness.md<NUL>full<NUL>present<NUL>100644<NUL>90286bdc9cffc97a354b4d89491ee98f18ba85da<NUL>
  docs/en/design/DP-021-private-exact-scope-managed-start-invoker.md<NUL>full<NUL>present<NUL>100644<NUL>29f289c41047d07682e6ef5402a60da5e3dd1bf5<NUL>
  docs/en/design/README.md<NUL>full<NUL>present<NUL>100644<NUL>b177cc044569c2ba36cd9cdff6679074eae6de75<NUL>
  docs/en/roadmap/MASTER_PLAN.md<NUL>full<NUL>present<NUL>100644<NUL>286197478dd8aa35c853eb55d9419fdebd611059<NUL>
  docs/engineering/EXECUTION-INTERRUPTION-RECOVERY-ACCEPTANCE-SCENARIOS.md<NUL>full<NUL>present<NUL>100644<NUL>ac8414d7fa8bb0b73d80d97a968ab7c5a1e9c298<NUL>
  docs/engineering/PROCESS-001-AI-DEVELOPMENT-WORKFLOW.md<NUL>full<NUL>present<NUL>100644<NUL>590cab13e3ca2b2c192a787aa58be651d9723ab3<NUL>
  docs/engineering/PROCESS-002-DOCUMENTATION-SYNCHRONIZATION.md<NUL>full<NUL>present<NUL>100644<NUL>49b8cb942b19ff61b37ad89666559fb9004a9bff<NUL>
  docs/engineering/TASK-TEMPLATE.md<NUL>full<NUL>present<NUL>100644<NUL>9a62a2f2b78bc146eba4490700350138a4ce4a6c<NUL>
  docs/engineering/agents/coordinator.md<NUL>full<NUL>present<NUL>100644<NUL>22146e3e98f68c0163c899355dc268b7261e5858<NUL>
  docs/engineering/agents/reviewer.md<NUL>full<NUL>present<NUL>100644<NUL>439c5b84216b0d84c2de99232a5ca5ca1f1be569<NUL>
  docs/engineering/agents/tester.md<NUL>full<NUL>present<NUL>100644<NUL>1fdc6a7a8b47a46163589c7b7ea034c948ba898b<NUL>
  docs/ru/design/DP-015-runtime-management-command-idempotency.md<NUL>full<NUL>present<NUL>100644<NUL>ad0c88c619db1e4760ba951cda228351e61837c5<NUL>
  docs/ru/design/DP-016-runtime-activation-replacement-rollback.md<NUL>full<NUL>present<NUL>100644<NUL>3312c3974ff06e3dec179b1a4ff07cc49abf908d<NUL>
  docs/ru/design/DP-019-runtime-activation-orchestration-prerequisites.md<NUL>full<NUL>present<NUL>100644<NUL>1a81224da0549d5ace3cb5a8a4d979ea54d427ce<NUL>
  docs/ru/design/DP-020-runtime-orchestration-binding-sequence-readiness.md<NUL>full<NUL>present<NUL>100644<NUL>7c22c6b1477527511f1aac2e3ba67f253e90c655<NUL>
  docs/ru/design/DP-021-private-exact-scope-managed-start-invoker.md<NUL>full<NUL>present<NUL>100644<NUL>bf756ad3aad4cc27441762dae9f646bb0870277c<NUL>
  docs/ru/design/README.md<NUL>full<NUL>present<NUL>100644<NUL>b8ebd7f2042992ff07afdd82c14e48508f409c0e<NUL>
  docs/ru/roadmap/MASTER_PLAN.md<NUL>full<NUL>present<NUL>100644<NUL>86d6cd86a31b3ca6b2d066d3ff5ae000f1971f15<NUL>
  docs/tasks/README.md<NUL>full<NUL>present<NUL>100644<NUL>f0955fbdaead5afc4b4cc8bcbdf541602b6e6f8a<NUL>
  docs/tasks/TASK-026-RUNTIME-ACTIVATION-REPLACEMENT-ROLLBACK.md<NUL>task-record-v1<NUL>present<NUL>100644<NUL>feabbed62af1c957bfde57becc16186d94a76548<NUL>
  spec/current-state.md<NUL>full<NUL>present<NUL>100644<NUL>b5d365a6879f3d7738a12a1815270617f442b478<NUL>
  spec/decisions.md<NUL>full<NUL>present<NUL>100644<NUL>fe4a25057b24859a6c5e6807ff62f43a25e6e2ae<NUL>
  ```
- exact verification commands and results (all commands were run against the
  subject above before this envelope-only append):

  | Command | Exit | Result |
  |---|---:|---|
  | `git status --short` | `0` | 26 expected modified tracked documentation paths; no staged or untracked paths |
  | `git diff --name-only` | `0` | exact 26-path set; no missing or extra path |
  | `git diff --cached --name-only` | `0` | empty |
  | `git ls-files --others --exclude-standard` | `0` | empty |
  | `git diff --check` | `0` | PASS; only Git line-ending advisories, no whitespace errors |
  | `go test ./... -count=1` | `0` | all repository packages passed; no implementation package is present |
  | `go vet ./...` | `0` | PASS |
  | `go mod tidy -diff` | `0` | no module diff |

- Verification Matrix results: documentation parity is `7/7` changed EN/RU
  pairs with aligned heading/fence counts — DP-015 `30/30, 4/4`, DP-016
  `30/30, 4/4`, DP-019 `25/25, 16/16`, DP-020 `34/34, 12/12`, DP-021
  `21/21, 10/10`, design indexes `1/1, 0/0`, and MASTER_PLAN `36/36, 0/0`;
  changed-file links are `271 valid / 0 broken`; conflict markers are `0`;
  task-record-v1 has one Status, one Task Contract, and one terminal Recovery
  Evidence Envelope heading in the required order. No product API,
  configuration, dependency, production wiring, or concurrency behavior was
  changed; race testing is **not applicable** to this zero-code subject.
- process-contract verification: the amended PROCESS-001, PROCESS-002,
  TASK-TEMPLATE, interruption scenarios R-029/R-030, and Coordinator/Tester/
  Reviewer contracts consistently require task-record-v1 projection, unsigned
  UTF-8 ordered rows, envelope exclusion/non-self-attestation, exact durable
  Tester evidence, and repeat review after projected rework. TASK-026 remains
  `Blocked`; DP-016 remains `Approved / Planned`; the DP-015/DP-020
  prerequisite remains `Not Activated` with no Task ID. The prior observed
  digest `ae8935eb3e2d643f3ad34d3353187765062d973f` is historical and was not
  reused.
- coverage and limitations: `26/26` subject paths verified; `26/26` required
  documentation paths present; code/test/module/dependency/generated/temp/
  staged/untracked counts `0`; no executable proof artifact was introduced.
  Race is not run because no production or test path exists in this subject;
  this is the only non-applicable matrix item. No other limitation remains for
  this evidence-only handoff. Scope Audit, Independent Reviewer, exact
  evidence-set recheck, and Blocked Closure Certification remain pending at
  handoff time; the subsequent identity recheck, Reviewer approval, Scope Audit
  and final certification below supersede that pending-state note.
- reproducibility: an independent agent can reconstruct this handoff from the
  Repository by reading the seven amended contracts and this envelope, running
  the commands above, applying the stated byte-exact projection, and verifying
  every ordered row and both OIDs. Chat history, partial output, raw/normalized
  diff digests, and the terminal envelope bytes are not used as subject
  identity. Post-append projection/manifest identity recheck follows as a
  separate envelope-only recovery step.

### Tester Envelope-Only Identity Recheck (2026-08-27)

- the append above was reconciled read-only immediately after mutation;
  `task-record-v1` heading/order invariants remain `1/1/1` (Status/Task
  Contract/terminal Envelope), projected bytes remain `83477`, projected blob
  remains `feabbed62af1c957bfde57becc16186d94a76548`, and the exact 26-path
  manifest remains `2873` bytes with OID
  `161a6d10f0355dc313b2290d1111bc1f75b2ad4e` under object format `sha1`.
- the exact path count remains `26`, staged count `0`, and untracked count `0`.
  Therefore the envelope-only append did not alter the certified subject,
  projected task-record identity, ordered rows, or manifest identity.

### Final Blocked Closure Certification (2026-08-27)

The amended bounded recovery pipeline is complete and the current evidence-only
subject is **`Blocked Closure Certified`**. This certification is not ordinary
Coordinator Acceptance, does not mark TASK-026 Completed, and does not authorize
implementation, staging, commit, push, publication, or prerequisite activation.

Immutable certification tuple:

- repository: `E:\wikiPRJ\universal-websocket-platform`;
- Task/status: `TASK-026` / `Blocked — missing DP-015/DP-020
  orchestration-admission refinement (2026-08-27)`;
- branch: `feature/task-026-runtime-activation-orchestration-implementation`;
- trusted base and certified HEAD: `main` /
  `65af65c3154d70db10de52483149960a42dcfb9a`;
- exact evidence set: the 26 present `100644` documentation paths in the
  durable Tester handoff immediately above, in ascending unsigned UTF-8 path
  order; the TASK record uses `task-record-v1`, all other present paths use
  `full`, and no deleted paths are present;
- canonical identity: Git object format `sha1`; task-record-v1 projection is
  `83,477` bytes with blob OID
  `feabbed62af1c957bfde57becc16186d94a76548`; the ordered NUL-separated
  manifest is `2,873` bytes with blob OID
  `161a6d10f0355dc313b2290d1111bc1f75b2ad4e`;
- canonical rows are exactly `path\0projection\0state\0mode\0oid\0`, sorted
  by unsigned UTF-8 path bytes; `full` rows hash raw current bytes and the TASK
  row hashes the stated task-record projection. The terminal envelope is
  excluded from that projection, so this tuple is non-self-referential;
- Verification Matrix and PROCESS-002: Tester `PASS`, exact `26/26`,
  missing/extra `0/0`, all commands and results durably recorded above; the
  sole non-applicable item is race testing because the subject has no
  production or test path;
- Independent Reviewer: `Approved`, blocking/non-blocking findings `0/0`;
  Scope Audit `26 Required / 0 Questionable / 0 Removable`;
- repository residue: production/test/module/dependency/generated/temp,
  staged, and untracked counts `0`; `internal/runtimeactivation` is absent;
- prior raw digest `ae8935eb3e2d643f3ad34d3353187765062d973f` is historical and
  was not reused;
- Coordinator Acceptance, Completion, and DP-016 implementation acceptance:
  **not performed / not claimed**;
- commit, Blocked Evidence Checkpoint, push, PR, merge, publication, and
  activation of the separate DP-015/DP-020 prerequisite: **not authorized /
  not performed**.

This immutable tuple is reproducible from the Repository using the amended
PROCESS-001/PROCESS-002 convention, the exact durable Tester handoff, and the
independent Reviewer record above. Any change to the projected subject, exact
path set, ordered rows, or tuple invalidates this certification and requires the
affected gates to be rerun.

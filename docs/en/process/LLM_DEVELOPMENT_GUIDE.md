# LLM Development Guide

## Purpose

This guide defines the engineering workflow for development assisted by language models. It is a living project standard for contributors, maintainers, the Project Owner, and engineers implementing approved Design Proposals.

The project uses a structured workflow instead of ad hoc code generation to provide:

- deterministic development;
- traceable architecture decisions;
- reproducible implementation;
- independent verification;
- long-term maintainability.

Assistance does not transfer engineering accountability. Every change remains subject to the same architecture, scope, testing, review, and ownership requirements as any other project contribution.

## Core Principles

### Architecture Before Code

The intended component model, responsibilities, ownership, lifecycle, invariants, and boundaries must be established before implementation begins.

### Design Proposal as the Source of Truth

An approved Design Proposal is the normative source for the behavior and architecture in its scope. Implementation must conform to it and must not silently reinterpret it.

### Compile Once

Configuration-dependent validation, normalization, ordering, and resolution should occur once at the appropriate construction boundary. Published runtime structures should be ready for direct execution without defensive recompilation on operational paths.

### Immutable by Default

Published configuration, compiled structures, identity snapshots, and message context values should be immutable. Mutable state requires explicit ownership and synchronization contracts.

### Small, Independently Reviewable Steps

Implementation must be decomposed into the smallest coherent steps that compile, preserve existing behavior, and can be reviewed independently.

### One Completed Idea per Commit

Each commit should contain one complete engineering idea, its tests, and any documentation required to describe that idea. Partial or unrelated changes must not be combined.

### Performance Requirements Belong to Architecture

Allocation limits, bounded work, compilation boundaries, and concurrency properties must be specified before they are relied upon by implementation.

### Independent Verification

Implementation and review are distinct responsibilities. Review must inspect the repository state and evidence directly rather than rely on the implementation report.

### Blockers Are Preferable to Incorrect Implementations

Implementation must stop when the approved architecture is ambiguous, contradictory, or insufficient. Missing decisions must be resolved explicitly before coding continues.

### Hot Paths Require Explicit Performance Review

Message processing, routing, admission, and other frequently executed paths require deliberate review of allocations, synchronization, data copying, boundedness, and hidden work.

## Roles

### Architecture

The architecture role:

- defines component responsibilities and boundaries;
- assigns ownership and lifecycle authority;
- specifies observable behavior and invariants;
- records material decisions in the appropriate architectural document;
- resolves ambiguity before implementation begins;
- approves or rejects proposed architectural changes.

### Implementation

The implementation role:

- follows the approved Design Proposal and implementation scope;
- inspects existing code and tests before making changes;
- introduces no hidden architectural decisions;
- preserves compatibility unless a change is explicitly approved;
- implements deterministic proof tests;
- runs all required verification;
- reports the resulting repository state accurately.

### Review

The review role:

- independently compares code and tests with the approved architecture;
- challenges ownership, lifecycle, concurrency, immutability, and boundary claims;
- verifies that tests prove the required invariants;
- classifies findings by impact;
- distinguishes architectural defects from implementation defects;
- withholds approval when evidence is incomplete.

### Project Owner

The Project Owner:

- controls project scope and roadmap priority;
- approves architecture and material changes to it;
- resolves product-level trade-offs;
- decides whether findings block progress;
- authorizes commits and integration;
- ensures that project records reflect completed work.

No role may assume the authority assigned to another role merely to keep implementation moving.

### Publisher

The Publisher integrates one admissible accepted or evidence target after explicit publication
authorization. The exact command `Разрешаю публиковать.` authorizes the whole
immutable-target pipeline from read-only preflight through push, Pull Request,
checks, merge, branch cleanup, synchronized local `main`, terminal report, and
STOP. Push and merge are checkpoints, not terminal outcomes. Commit creation
remains separately authorized.

## Standard Workflow

```text
MASTER PLAN
    |
    v
Design Proposal
    |
    v
Architecture Review
    |
    v
Implementation Prompt
    |
    v
Implementation
    |
    v
Independent Review
    |
    v
Architecture Fix, if required
    |
    v
Commit
    |
    v
Publication
    |
    v
Post-Implementation Architecture Review
    |
    v
CHANGELOG
```

### Master Plan

The Master Plan identifies the engineering sequence and milestone intent. It selects the problem area but does not replace a Design Proposal or prescribe future APIs.

### Design Proposal

The Design Proposal defines the scoped architecture, contracts, invariants, failure model, compatibility requirements, and excluded work. It must contain enough information for implementation without inventing new decisions.

### Architecture Review

Independent architecture review attempts to disprove the proposal. Blocking ambiguity, ownership gaps, invalid lifecycle transitions, and unproven concurrency semantics must be resolved before implementation.

### Implementation Prompt

The implementation prompt converts the approved design into one bounded engineering step. It states the permitted files and behavior, exclusions, required tests, verification commands, and expected report.

### Implementation

Implementation changes only the approved scope. It includes proof tests and preserves the repository in a compiling, reviewable state.

### Independent Review

Independent review verifies actual code and tests against the Design Proposal and implementation prompt. It does not treat a successful build or implementation report as proof of architectural correctness.

### Architecture Fix

If review exposes an architectural defect, implementation pauses. The model is analyzed first, the required architectural change is approved, and the correction is delivered separately from unrelated work. Implementation-only defects do not require an architecture change.

### Commit

A commit is created only after the scoped implementation and required verification are complete and review findings have been resolved or explicitly accepted. The commit records one completed idea.

For evidence-only paths, the separate exact command `Разрешаю коммит.` may
instead authorize one Blocked Evidence Checkpoint after Blocked Closure
Certified, or one Negative Disposition Checkpoint after Negative Disposition
Recorded and post-decision integrity. These are not implementation Acceptance.
The complete gates in PROCESS-001 remain mandatory, including exact staged-tree
matching; LF/CRLF mismatch is not equivalence. No decision grants permission.

### Publication

Publication authorization is bound to the exact class, task/repository/branch,
ordered commit target, base `main`, and class-specific scope:

- `Accepted Task`: accepted task commit/scope;
- `Blocked Evidence Recovery`: certified checkpoint/recovery-chain and scope;
- `Negative Disposition`: one exact Negative Disposition Checkpoint directly
  above the fixed base, its disposition tuple and negative scope.

Negative Disposition follows ND-1–ND-5 in
[PROCESS-001](../../engineering/PROCESS-001-AI-DEVELOPMENT-WORKFLOW.md).
It requires proven ownership/preservation, an exact mandatory provenance
blocker, unavailable normal Acceptance/BCC, independently demonstrated bounded
Required Recovery Exhausted, full governance testing/review, PROCESS-002 and
Scope Audit. A known feasible recovery route, unknown required-source outcome,
product/test changes or unresolved blocking finding rejects this path.
Not Proven remains uncertainty; Disproven retains its explicit refutation.
Neither is positive downstream evidence or successful implementation.

The separate command `Разрешаю публиковать.` remains required after the
separately authorized checkpoint. All classes execute the same complete pipeline:

```text
P0 read-only preflight
    -> P1 push exact branch/commit
    -> P2 inspect then create or discover one PR to main
    -> P3 inspect required checks
    -> P4 reconfirm MERGEABLE / CLEAN and merge
    -> P5 delete the exact remote task branch
    -> P6 checkout main
    -> P7 pull --ff-only
    -> P8 safely delete the exact local task branch
    -> P9 verify main == origin/main and a clean worktree
    -> P10 full terminal report and STOP
```

Initial P0 verifies a clean staged/unstaged/untracked state, current exact task
branch and commit, immutable Target, origin and noninteractive SSH access,
`gh auth status`, and access to the current GitHub repository/default branch.
An auth, transport, or repository failure inside P0 leaves P0 first unfinished,
zero completed pipeline steps, and P1 not attempted. A successful push
proceeds directly to Pull Request discovery/creation without another
permission or a `Create Pull Request` prompt.

The exact execution context that will perform publication must prove usable
GitHub capability with two successful read-only probes: a decisive GitHub API
user/repository operation and Git remote authentication/read for the exact
origin. Both are mandatory; `gh auth status` is supporting diagnostics only.
A user profile path, account or helper
configuration, keyring reference, installed credential tool, or successful
probe from another Windows identity is not capability evidence. Non-secret
evidence identifies the context and results; credentials, tokens, headers, and
credential payloads must never enter prompts, logs, task records, or the
repository.

If the source context lacks capability, it may issue a Release Handoff with a
unique opaque non-secret transfer ID for one release instance and the unchanged
immutable Target, then become observation-only. The user must explicitly route
that exact ID and Target for the previously authorized publication to a
trusted destination. The destination performs `Inspect -> Reconstruct ->
Reconcile`, proves the Target and all local/remote checkpoints, passes both
probes from its exact context, and then records Accept Handoff citing the same
ID and Target. Accept Handoff
is the procedural ownership linearization point: only the destination may
resume P0-P10; the source and duplicate destinations remain observation-only.
This is an exclusivity contract, not a machine lock; the transfer ID is also
not a secret, credential, or permission.

The normative Transfer Identity is the immutable tuple `{transfer ID, Target,
source execution identity, Release checkpoint snapshot}`. The ID is a fresh
canonical lowercase UUIDv4 absent from all available operational handoff
records for that publication; it is opaque and is not derived from the Target,
identity, or time. The Target contains publication class, Task ID,
repository/origin, exact branch, ordered range/head OID, base OID, and scope
identity. The snapshot contains P0-P10 classifications, known refs/PR/merge
OIDs, and the first unfinished step. The explicit user route and Accept event
bind the exact destination identity to those unchanged fields.

The state model has three independent axes: authorization `Active |
Consumed(P10) | RevokedByUser | InvalidatedByTargetChange`; mutation ownership
`Owned(context) | InTransitNone | NoneTerminal | Unknown`; and transfer attempt
`Unissued | Released | Accepted | Closed(reason)`. Release sets
`Active/InTransitNone/Released`; Accept sets
`Active/Owned(destination)/Accepted`.

`CancelledBeforeAccept` requires an explicit user directive naming the exact
ID and reconciliation proving no Accept or destination side effect. Its event
closes the ID, keeps authorization `Active`, returns ownership to the recorded
releasing source, and requires that context to re-prove capability. After
Accept, only the current destination-owner may reverse-transfer by releasing a
fresh ID and losing ownership at Release; cancellation returns that releasing
destination. A factual Target mismatch sets
`InvalidatedByTargetChange/NoneTerminal` and closes all attempts; user revoke
sets `RevokedByUser/NoneTerminal` and a later run needs a new gate. P10 may be
terminalized only when its predecessor chain proves
`Active/Owned(execution-context)` for the exact actor. `Released/InTransitNone`
forbids P10 until an exact Accept establishes the destination owner or a valid
`CancelledBeforeAccept` restores the releasing owner. Proven P10 then always
sets `Consumed(P10)/NoneTerminal` and recovery is report-only. With no handoff,
the attempt remains `Unissued` and a publication-level P10 event uses
`transfer ID: none`; current `Accepted` closes that exact attempt; an already
`Closed(reason)` attempt retains its reason and P10 is a separate event only
while current ownership remains proven. `NoneTerminal`, `Unknown`, or missing
owner proof cannot create P10.

Release, user route, Accept, and closed events form an append-only non-secret
operational record outside the immutable Target and project-state documents.
Each event cites Target, actor identity, predecessor event ID or digest/tail,
resulting three-axis state, owner, and terminal reason/disposition. Transfer
events cite the exact ID; a publication-level P10 event explicitly cites
`transfer ID: none` and never fabricates an attempt.
The record must survive interruption and be independently inspectable by
source, destination, and Coordinator. If it is unavailable, ambiguous, or
conflicting, ownership is `Unknown` and all publication mutations STOP. This
is procedural exclusivity, not a machine or distributed lock.

Projected live-state sources remain verification-stable `In Progress`. The
exact latest verdict, identity, and first incomplete checkpoint come only from
the newest valid terminal envelope entry matching an independently recomputed
canonical manifest. Missing, stale, conflicting, or mismatched evidence means
STOP; it does not imply Acceptance, commit, or publication.

The handoff preserves authorization only for the unchanged publication class,
Task ID, repository, branch, ordered commit range/head OID, base, and scope. It
does not create a new permission, Commit Gate, or Coordinator Acceptance.
Different release/destination, unknown/reused/mismatched/duplicate/
already-accepted ID, ambiguous ownership, missing explicit user routing, or
interruption before Accept Handoff forbids side effects. Repeat, reverse, and
cancellation follow only the exact transitions above; every new attempt uses a
fresh ID. A completed remote effect found
during reconciliation is not blindly repeated.

Credential unavailability to the exact execution identity is classified
separately from invalid/expired credentials, repository permission denial,
network/transport failure, GitHub outage, and tool/session failure. Login,
secret transfer, authentication bypass, and undocumented elevation are not
handoff mechanisms.

Required checks must succeed. No workflows or zero registered checks is
reported as `No CI` and is non-blocking only when the merge gate is
`MERGEABLE / CLEAN`. Pending or failed required checks, an unproven or
non-mergeable PR, conflicts, and branch protection refusal block merge and
must not be bypassed.

External blockers do not consume publication authorization. The blocker
report lists completed steps, the exact first unfinished step, factual state,
preserved refs/worktree, and the required unblock action. After
`Авторизация готова. Продолжай ранее разрешённую публикацию.` the Publisher
uses a non-checkpoint, phase-aware Resume Reconstruction Guard and resumes
without another publication command. Before confirmed P6 it normally expects
the clean task branch/commit phase. After confirmed P6 it requires clean
current `main`, never requires or recreates the task branch, permits `main` to
be behind until P7, and requires equality only at P9. A changed target commit,
branch, base, or scope instead invalidates the exact authorization.

After confirmed merge, remote/local branch cleanup and synchronized clean
`main` are mandatory. Cleanup uses the exact branch, verifies the remote ref
still points to the authorized commit, uses only fast-forward pull and safe
local deletion, and never uses force, reset, or rebase. Terminal success
reports PR number/URL, task and merge commits, checks state, observed
`MERGEABLE / CLEAN`, both branch deletions, `main == origin/main`, clean
worktree/current `main`, and then STOP.

For Negative Disposition, P0 additionally verifies the exact checkpoint/base,
decision tuple, negative facts and absence of Acceptance/BCC/Completed claims.
All capability, ownership/handoff, invalidation and phase-aware recovery rules
above apply unchanged. P10 proves only publication of negative evidence.
Negative Disposition Recorded stops original work but keeps the active-task
barrier. Only reconstructed terminal P10, exact merged PR/ancestry, removed task
refs and clean synchronized main yield Sealed Negative Disposition. A later
separate ordinary intake still needs its own readiness; publication does not
activate it. No immutable checkpoint is edited to record its publication.
Before/after decision or commit, missing authority never follows from a status;
unknown effects are inspected before retry, and changed targets invalidate it.
Before proven P10, new concrete provenance evidence/pointers halt the next
mutation for independent eligibility revalidation. Active authority is not
sufficient; failed eligibility forbids remaining effects/P10/intake even with
an unchanged Git target. Do not force cleanup to obtain a clean baseline.
Only an actual tuple/target change triggers TargetChanged. After proven P10,
new evidence instead requires a separate authorized normal intake.

### Prospective Acceptance of an Immutable Published Subject

IPSPA is a new evidence event for already published immutable Git bytes. It
never repairs historical Acceptance: Historical Equivalence (`Proven`, `Not
Proven`, or `Disproven`) remains independent from the prospective event.

The source is an exact repository, full commit/tree OID, optional fixed
deletion base, ordered path set, and full-only canonical manifest read through
Git object APIs. Working-tree, checkout/filter, normalized, decoded, archive,
or diff bytes are not authoritative. A separate Evidence Record cites the
source but cannot occur in its source tree, path set, or manifest.

Fresh independent source verification, applicable testing, documentation
synchronization, Scope Audit, Independent Review, explicit Coordinator
Prospective Acceptance, and post-decision integrity are mandatory. Historical
verdicts do not transfer. Source or evidence mutation invalidates affected
gates and is recovered inspect-first. An accepted event may satisfy only a
downstream contract that explicitly names its exact source and claims, followed
by a separate repository-first reassessment. It creates no fourth publication
class and does not activate work automatically.

### Post-Implementation Architecture Review

The completed implementation may be reviewed against the broader component model to confirm that local correctness did not introduce boundary drift or invalidate downstream assumptions.

### Changelog

The changelog is updated only when the repository contains the completed capability. It records the implemented state, not intended or partially implemented behavior.

## Design Proposal Rules

- Implementation must not change architecture.
- The proposal must distinguish normative requirements from examples and future work.
- Any ambiguity that affects ownership, lifecycle, behavior, concurrency, failure handling, or public contracts stops implementation.
- Architecture changes require explicit review and approval in the appropriate document.
- An implementation convenience is not sufficient justification for changing an approved invariant.
- Deferred behavior must remain absent; placeholder APIs must not imply unsupported capability.
- Code is evidence of implementation state, not a substitute for the approved architectural model.

## Implementation Rules

- Work must remain within one explicitly bounded scope.
- Each component and change must have a single stated responsibility.
- Existing behavior must remain compatible unless the prompt authorizes a change.
- Public APIs must not be added for speculative future use.
- Dependencies must follow approved package and ownership boundaries.
- Operational paths must not acquire hidden compilation, normalization, I/O, concurrency, or allocation costs.
- Tests must prove the specified success, failure, lifecycle, concurrency, and compatibility invariants.
- Synchronization tests must be deterministic and must not use sleep-based ordering.
- All started goroutines must be deterministically released and joined by tests.
- Required formatting, tests, static analysis, and diff checks must pass before completion is reported.
- The final report must describe actual changes and verification results, including checks that could not be performed.

## Review Rules

Reviewers must verify the following where applicable:

- dependency direction and package boundaries;
- immutable ownership and absence of aliasing;
- resource, connection, context, and lifecycle ownership;
- legal state transitions and terminal semantics;
- concurrency linearization points and exactly-once guarantees;
- bounded execution and algorithmic cost;
- hot-path allocations, copying, locking, and hidden work;
- error identity and observable behavior;
- backward compatibility and default behavior;
- consistency between architecture, implementation, tests, and current-state documentation;
- compliance with the explicit scope and exclusions.

Review must cite concrete documents, files, and tests. Successful formatting, compilation, or tests do not by themselves prove ownership or architectural correctness. Findings must identify the violated invariant and its impact.

## Prompt Standard

Every implementation prompt should contain the following sections.

### Objective

Defines the single outcome of the task and prevents unrelated goals from entering the change.

### Scope

Identifies permitted behavior, packages, files, and architectural responsibilities. It establishes the authority available to implementation.

### Out of Scope

Lists adjacent capabilities that must remain absent. This prevents speculative APIs, hidden integration, and premature subsystem work.

### Architecture Constraints

Restates the applicable source documents, ownership rules, lifecycle invariants, compatibility requirements, dependency boundaries, and performance constraints.

### Tests

Specifies the properties that require proof, including failure and concurrency cases. Tests should demonstrate invariants rather than only exercise lines of code.

### Verification

Lists the exact formatting, test, static-analysis, race, allocation, documentation, and diff checks required for the task. Unavailable checks must be reported precisely and must not be represented as completed.

### Final Report

Defines the evidence required for handoff: changed files, resulting contracts, tests, verification results, limitations, risks, and a suggested commit message. It must describe the resulting state rather than the development narrative.

## Lessons Learned

Lessons are objective engineering knowledge extracted from completed Design Proposals and their implementation. They complement the governing architecture but do not alter it.

DP-005 provides the following examples:

- Independent review is necessary even when compilation and tests pass; it verifies whether implementation preserves the intended model.
- Immutable compiled runtime structures separate validation and preparation from message-time execution.
- Configuration normalization, Handler resolution, and route ordering belong at the construction boundary rather than the routing hot path.
- Priority ordering must be explicit in both the compiled representation and its proof tests; declaration order must not become an accidental runtime rule.
- Hot-path behavior benefits from focused allocation and concurrency proofs in addition to functional tests.
- Architecture corrections should be isolated from feature implementation so that their intent, evidence, and rollback boundary remain clear.

New lessons should record reusable engineering outcomes. They must not replace Design Proposals, decision records, implementation reports, or project history.

## Future Evolution

This guide evolves with the project. Future Design Proposals, implementation reviews, and demonstrated engineering constraints may extend the workflow. Changes to this guide must remain consistent with the project's architecture, decision process, and repository verification standards.

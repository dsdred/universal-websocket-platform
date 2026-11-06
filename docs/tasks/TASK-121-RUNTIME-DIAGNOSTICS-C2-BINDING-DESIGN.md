# TASK-121 — Runtime Diagnostics C2-Binding design

## Status

`In Progress`. Design-only; production C2/C3 remain Not Activated, Gate 11 OPEN.
No Commit or Publication Gate has been authorized.

## Task Contract

### Task Mode

Design-only. The task may update only the necessary Draft design sources and
project-state documentation. It must not change production code, tests, modules,
generated artifacts or production wiring.

### Why Now

- TASK-120 published the isolated C2-Pull prerequisite through PR #120 and P10R PASS.
- DP-027 section 32 places concrete Host-lifetime C2-Binding before C2-Exposure.
- The repository still lacks a production mailbox-pair allocation, ownership,
  retirement and authorized access contract.
- Owner authorization accepted the repository-first intake and its Architecture
  Confirmation PASS for this exact design-only slice.

### Definition of Done

- terminal TASK-120 publication state is synchronized in every applicable current
  project-state surface;
- actual construction, launch, management and deployment seams are inventoried;
- a concrete production owner and process/deployment scope are selected from
  repository evidence, or the task stops with a blocking finding;
- immutable per-Host-lifetime allocation, binding, retention through same-Host
  rollback/retry, retirement, revocation and lifetime-handle ownership are defined;
- destructive consume authorization and the C2-Binding/C2-Exposure boundary are
  explicit without activating either production capability;
- lifecycle traces and a P01-P08 traceability matrix classify every proof as
  `Satisfied by Design`, `Deferred` or `Blocking`;
- EN/RU parity, Architecture Confirmation, independent Verification, PROCESS-002,
  Scope Audit, independent Final Review, final-byte integrity and Coordinator
  Acceptance are complete for one frozen documentation subject.

### Out of Scope

- production or test code, Host/Bootstrap/Control Service wiring and activation;
- C2-Exposure implementation, endpoint, serializer, sink, worker or transport;
- C3 schema/source work or implementation;
- TASK-102, Windows/native qualification and Gate 11 closure;
- retries, replay, durable/lossless delivery, flush, final-emptiness or lifecycle
  join guarantees;
- changes to Frozen/Approved contracts without a separate Owner decision;
- stage, intent-to-add, commit, push, PR, merge or publication.

### Verification Plan

- source-backed inventory of production constructors/call sites and the relevant
  Runtime, lifecycle, launch-flow and management contracts;
- mirrored semantic comparison for every changed EN/RU design/roadmap document;
- local-link, UTF-8/BOM/NUL, whitespace and repository-scope checks;
- canonical exact-path subject identity after design freeze;
- independent Verification and independent non-author Final Review on that subject;
- documentation-only PROCESS-002 and per-path removability Scope Audit.

## Objective

Define the smallest repository-conforming production C2-Binding contract for one
diagnostics mailbox pair per concrete Runtime Host lifetime, while leaving production
C2/C3 inactive and preserving all existing Runtime lifecycle and security boundaries.

## Selection Evidence

- Baseline `main`/`origin/main`/HEAD:
  `da0ea8c8f8c0c5f412171bbc59c2cb35cbb3c63b`, clean and synchronized at intake.
- HEAD is the PR #120 merge; TASK-120 source commit is
  `40e90d82599fd42ef9bf7fa967a83575829db603`.
- TASK-120 publication is terminal per current Owner statement: PR #120 MERGED,
  P10R PASS. Its projected pre-publication state requires current PROCESS-002 sync.
- DP-027 sections 30, 32 and 33 identify C2-Binding and P01-P08 as the next C2
  prerequisite after isolated Pull.
- C2-Exposure was rejected as premature because it depends on accepted binding,
  consume authorization, scheduling/resources and transport/retention design.
- C3 was rejected because it is a separate schema/source track and is not activated.

## Scope

- New task record and applicable project-state synchronization.
- Draft DP-027 and only the Draft lifecycle/launch/management design sources needed
  by the selected owner and access boundary, with mirrored EN/RU changes.
- Exact lifecycle traces for allocation, bind, launch, rollback/retry, Stop,
  failed startup, retirement, revocation and next attempt.
- P01-P08 design traceability and an explicit future implementation split.

## Non-Goals

- No successor implementation task is activated or assigned automatically.
- No public operator capability, management route or observable delivery claim.
- No reclassification of isolated O01-O16 as production P01-P08 evidence.
- No historical task-record rewrite.

## Sources of Truth

- `AGENTS.md`, AGENT, PROCESS-001, PROCESS-002 and applicable role contracts;
- ADR-0002/0003/0004, Frozen ARCH-002, Active ARCH-003/004/005;
- Draft DP-009/010/011/012/013 and DP-027;
- TASK-116, TASK-119 and TASK-120;
- current Runtime Bootstrap/composition, lifecycle Owner, launch Flow, management
  Directory and process entry-point source.

## Roles

- Coordinator: repository intake, scope control, acceptance and handoff.
- Architect: source-backed owner/binding/access design; no implementation.
- Documentation Agent: record accepted design in Draft EN/RU sources and state docs.
- Developer: N/A; production and test changes are prohibited.
- Tester: N/A for executable product behavior; documentation/source-conformance
  verification is performed by the Verifier.
- Reviewer: independent exact-subject design and Scope Audit review.
- Publisher: inactive; no Publication Gate.

## Branch

- trusted baseline: `main` at `da0ea8c8f8c0c5f412171bbc59c2cb35cbb3c63b`;
- task branch: none created; current worktree remains on `main`;
- branch action: none authorized or required for this design-only worktree;
- prohibited: stage, commit, push, PR, merge, fetch/pull and remote/ref mutation.

## Constraints

- one concrete Host lifetime has exactly one mailbox pair;
- rollback/retry on that same Host retains the pair; a new Host/attempt gets a new pair;
- `Owner.dependencies` is never mutated or rebound; no Owner-level shared mailbox;
- no fifth Bootstrap binding or public Host capability;
- `ActionObserve` and DP-018 authority never authorize destructive Poll;
- lossy semantics remain explicit; no flush, final emptiness, delivery acknowledgment
  or lifecycle join is introduced.

## Source-backed Design Decision

### Production owner and scope

The repository provides a justified owner. Active ARCH-004 assigns the
Control Service-side Runtime Lifecycle Owner responsibility for one Runtime Instance,
Launch Attempt creation, the only Runtime Launcher invocation and retention of the
active Host. Draft DP-010 and `internal/runtimelifecycle/owner.go` implement that
isolated owner and its single `Owner -> runtime.Launch` path. Therefore:

- the per-Runtime-Instance `runtimelifecycle.Owner` owns one private diagnostics
  lifetime per accepted Host launch;
- the lifetime belongs to the active attempt/Host execution, not to Owner as a
  reusable dependency;
- deployment scope is the future single Control Service process containing Owner and
  its in-process Host; current `cmd/control-service` has no Runtime composition and
  remains inactive;
- a future multi-process deployment requires a separate decision because Reader/View
  capabilities are process-local.

This is not a fabricated new application owner: it follows the existing ARCH-004
ownership and exact implemented launch seam. Production Activation remains absent.

### Immutable binding

After `Owner.Start` accepts a valid Snapshot and commits `AttemptLaunching`, but before
the sole Launcher call, Owner creates one mailbox pair and one C2-Pull View. It stores
the View in private attempt state, copies the stable base `DependencyBindings`, writes
Producer only into the copy's existing optional field and passes that private copy to
the Launcher. `Owner.dependencies` remains borrowed and byte-for-byte unchanged.
The C2-bound base Producer must be zero/disabled; a nonzero value fails binding
admission before allocation/Launch and is never silently overwritten or reused.

No pair is created for Load/Build/preparation failure. Rollback/retry inside the same
Host launch and overlapping Stop retain the pair. Launch failure without a retained
Host or successful Host Stop retires it; Stop failure retains Host and pair. A later
accepted attempt creates a new pair and sequence domain. There is no shared Owner
mailbox, fifth Bootstrap binding, Host capability or identity label.

### Lifetime handle, retirement and access

The private attempt state exclusively owns an opaque handle containing exact Launch
Attempt identity, View and `active/retired` state. Raw Reader/View never leaves Owner.
A future management consume operation must authorize exact Target plus expected Launch
Attempt before lookup/access/read, then ask Owner to atomically match the active,
non-retired lifetime. Owner copies View under its mutex and Polls only after unlock.

Locked retirement revokes later access. An already admitted Poll may complete; Stop
does not join or cancel it. Unread and racing late facts may be lost. No retirement
path drains, flushes, acknowledges delivery or proves final emptiness.

`ActionObserve`, the current Authorize version parameter and DP-018 authority are not
consume authority. Concrete policy, rate/concurrency/pending bounds, transport,
rendering, retention and redaction remain C2-Exposure.

### P01-P08 disposition

| Proof | Disposition | Reason |
| --- | --- | --- |
| P01 | Satisfied by Design | Lifecycle Owner, single-process scope and private exact-attempt access seam are selected. |
| P02 | Satisfied by Design | One pair per accepted Host launch, same-Host retention, later-attempt isolation and immutable dependency-copy rules are exact. |
| P03 | Deferred | Ordering/revocation/anti-enumeration are specified; concrete authorization/API belongs to C2-Exposure. |
| P04 | Deferred | Scheduling, concurrency, pending, rate and retention limits require C2-Exposure. |
| P05 | Deferred | Lifecycle non-join is specified; real bridge/rendering overlap proof requires production composition. |
| P06 | Satisfied by Design | Retirement, failed Stop, admitted Poll and late/unread lossy semantics are exact. |
| P07 | Deferred | Allowlist remains; operator rendering/redaction requires C2-Exposure. |
| P08 | Deferred | Code composition, compatibility and executable production proof require separately authorized implementation/integration. |

No design row is Blocking. These classifications do not claim executable P proofs.
Production C2/C3 remain Not Activated and Gate 11 remains OPEN.

## Stop Conditions

- no concrete production owner can be justified from repository evidence;
- the proposal conflicts with a Frozen/Approved source or needs its amendment;
- owner selection requires product prioritization outside the authorized slice;
- scope expands into implementation, C2-Exposure, C3, TASK-102 or Gate 11 closure;
- baseline/diff attribution becomes inconsistent or an independent blocking finding
  remains unresolved.

## Acceptance Criteria

1. TASK-120 terminal publication facts replace stale current-task projections in
   all applicable project-state documents without rewriting TASK-120 history.
2. A repository-backed production owner and deployment scope are named exactly.
3. Allocation and lifecycle traces preserve every mandatory invariant and distinguish
   one Host lifetime from Runtime Instance and Launch Attempt identity.
4. Lifetime handle ownership, retirement/revocation and late-fact loss semantics are
   explicit and do not add Host or Bootstrap public capabilities.
5. Consume authorization occurs before any destructive Poll and is separate from
   `ActionObserve` and DP-018; implementation remains deferred to C2-Exposure.
6. P01-P08 has one explicit `Satisfied by Design`, `Deferred` or `Blocking`
   disposition per row, without converting design evidence into executable proof.
7. Necessary Draft contracts and project-state surfaces are synchronized with EN/RU
   parity; Gate 11 remains OPEN and C2/C3 remain Not Activated.
8. Independent Verification and Final Review report no unresolved blocking findings;
   PROCESS-002, Scope Audit and final-byte integrity pass on one frozen subject.

## Verification

- Existing Coverage Report:
  - Existing Coverage: TASK-120 isolated C2-Pull O01-O16; DP-010 lifecycle Owner,
    DP-011 Flow, DP-012 composition Source and DP-013 Directory exist in isolation.
  - Coverage Gap: no production composition, binding code, consume policy, exposure,
    transport or executable P01-P08 evidence.
  - Added Proof Tests: none; design-only task.
  - Added Regression Tests: none; production/test files unchanged.
  - Remaining Limitations: Windows race/sandbox/CRLF TASK-120 limitations remain
    historical and do not qualify production binding.
- Verification Matrix:
  - concurrency/lifecycle/shared state: design traces in DP-010/011/027;
  - API/configuration/production wiring: explicitly deferred and inactive;
  - dependencies: private copy of existing binding only; no new module/package;
  - public API: no implementation or Host/Bootstrap capability change;
  - documentation: DP-010/011/012/013/027 EN/RU plus current-state surfaces;
- formatter/lint: documentation checks only, pending.
- tests/race/vet: N/A to design-only diff; no runtime result will be claimed.
- documentation structure: primary check PASS — 18 UTF-8 paths, no BOM/NUL,
  local links 0 missing, `git diff --check` no errors; independent check pending.
- independent review: pending.

## Architecture Confirmation

Independent Architect `/root/architecture121`: **PASS**, blocking findings 0,
non-blocking findings 0 after one corrected coherence finding. Owner selection is
justified by ARCH-004 and the actual sole `Owner -> runtime.Launch` seam. DP-010
sections 15/26/30 now distinguish current implementation from the future C2-bound
private dependency copy, define base/copy/handle ownership and fail closed on a
nonzero base Producer. ARCH-002 Host ownership, DP-009 binding shape, Approved/Frozen
contracts, production inactivity and EN/RU parity are preserved.

## Scope Audit

Primary classification: **18 Required / 0 Questionable / 0 Removable**, pending
independent exact-subject audit.

- task record: selection, design, evidence and recovery anchor;
- PROJECT_CONTEXT/current-state/task index: current task and TASK-120 terminal truth;
- MASTER_PLAN EN/RU and design indexes EN/RU: current roadmap/navigation truth;
- DP-027 EN/RU: central owner/lifecycle/P01-P08 decision and TASK-120 publication;
- DP-010 EN/RU: normative Owner allocation, private binding, handle and retirement;
- DP-011 EN/RU: exact launch ordering and Flow non-ownership;
- DP-012 EN/RU: application composition confinement and audit;
- DP-013 EN/RU: consume authorization ordering and C2-Exposure boundary.

Removing any path leaves PROCESS-002 state, an EN/RU mirror or one of the four Draft
seams explicitly required by DP-027 contradictory/incomplete. There is no generated,
formatting-only, source/test/module, TASK-102 or premature implementation path.

## Size Guard

Triggered: final candidate affects 18 documentation paths, exceeding the 15-path
signal. The task remains one cohesive contract: ten paths are mandatory task/current-
state/publication synchronization, while eight are the four mirrored Draft contracts
explicitly identified by DP-027 section 32 after owner selection. There are zero
production/test packages, zero executable behaviors and one architecture decision.
Splitting would leave either project state or the binding contract knowingly
contradictory. Size Guard decision: provisionally ACCEPT / DO NOT SPLIT, subject to
independent removability review.

## Documentation Sync

- task record: created first as this task's persistent anchor;
- current-state / PROJECT_CONTEXT / task index: required;
- MASTER_PLAN EN/RU: Runtime Diagnostics paragraph only, if needed for current truth;
- DP-027 EN/RU: required;
- DP-010/011/012/013 EN/RU: only where the accepted owner/binding boundary requires it;
- CHANGELOG/root README/Wiki: N/A absent a user-facing or released capability;
- parity, links and contradictions: primary PASS; independent verification pending.

## Interruption Recovery

- persistent anchor: this record; TASK-121; `In Progress`; baseline/HEAD above;
- current subject: 18 documentation paths pending exact freeze;
- completed checkpoints: clean synchronized baseline; first task-record change;
  TASK-120 terminal project-state synchronization; source-backed owner decision and
  mirrored Draft design;
- first incomplete checkpoint: Architecture Confirmation on the design bytes;
- mutation/permission state: design documentation edits authorized; Commit and
  Publication permissions absent;
- downstream evidence is invalidated by any material design rework.

## Commit Gate

- exact command `Разрешаю коммит.`: not received for TASK-121;
- gate: not ready;
- stage/commit: prohibited.

## Process Health

- trigger: no process-contract defect identified; no process change in scope.

## Handoff

Pending design and verification.

## Publication

Not authorized; Publisher inactive.

## Next Candidate

No implementation candidate is activated. A later C2-Binding implementation intake
is possible only after accepted design, fresh repository-first readiness and separate
Owner authorization.

## Closure

Pending Design Acceptance Verdict.

## Recovery Evidence Envelope

### A121-001 — Baseline, publication sync and Architecture Confirmation

Baseline/HEAD/main/origin-main are
`da0ea8c8f8c0c5f412171bbc59c2cb35cbb3c63b`; worktree began clean. TASK-120 source
`40e90d82599fd42ef9bf7fa967a83575829db603` and merge HEAD share tree
`fb7c36f7909787f7b47fc5a144048f78aaf1b16f`; P10R-120-001 PASS and terminal
`Consumed(P10) / NoneTerminal / Unissued` are synchronized. A transient malformed
OID line introduced during initial sync was removed before any design gate; exhaustive
search now finds none.

Independent Architecture Confirmation initially reported one DP-010 coherence blocker:
old sections 15/26/30 did not distinguish the implemented stable dependency pointer
from the future private attempt copy. EN/RU corrections now preserve current truth,
define the private copy/handle and make nonzero base Producer fail closed. Recheck PASS,
blocking/non-blocking findings 0/0. No production/test/module/Approved/Frozen change,
production C2/C3 activation or Gate 11 closure occurred.

### M121-001 — Frozen design subject

HEAD/baseline `da0ea8c8f8c0c5f412171bbc59c2cb35cbb3c63b`, branch `main`, Git object
format SHA-1. Canonical staging-independent manifest uses ascending ordinal UTF-8
path order and rows `path\0projection\0present\0100644\0oid\0`. Raw bytes are used
without filters; task-record-v1 is the exact prefix before the Recovery Evidence
Envelope heading. Manifest length 1889 bytes, Git blob OID
`5a441941fa8f4f0294991e43487b946c4790c006`.

| Path | Projection | Mode | OID |
| --- | --- | --- | --- |
| `.ai/PROJECT_CONTEXT.md` | full | 100644 | `f769616f9422a35af17b5d5ced02b8de3dccae12` |
| `docs/en/design/DP-010-runtime-lifecycle-owner-contract.md` | full | 100644 | `def9ac1d17d6507fa741e96f40b2ca11b58fb6db` |
| `docs/en/design/DP-011-runtime-launch-pipeline-integration.md` | full | 100644 | `dc3fe7486ef78d9652d0ccc32c63fa189cd8d13a` |
| `docs/en/design/DP-012-runtime-source-composition.md` | full | 100644 | `8575dbd8ebf22852ebc063ea7c470dbb041cc8de` |
| `docs/en/design/DP-013-runtime-management-routing.md` | full | 100644 | `9cbce1e5f1d5deef78937b9e759cd939127bd28e` |
| `docs/en/design/DP-027-runtime-diagnostics-v1.md` | full | 100644 | `e6dd56aae02feba4974a2e01886aded89ca99b76` |
| `docs/en/design/README.md` | full | 100644 | `670b8e1fa404237ec94cd7957304638ac0834220` |
| `docs/en/roadmap/MASTER_PLAN.md` | full | 100644 | `ec64b2a4e1113888ed1da75d94f6776156e94e64` |
| `docs/ru/design/DP-010-runtime-lifecycle-owner-contract.md` | full | 100644 | `d1cf5b1c517169d85b20c50ac97ca8384f329f8a` |
| `docs/ru/design/DP-011-runtime-launch-pipeline-integration.md` | full | 100644 | `b687afd697b87c29d2029c20471fff06ec3d060d` |
| `docs/ru/design/DP-012-runtime-source-composition.md` | full | 100644 | `a6d0eb6a81689abded42ffd10b740a9dec913e9c` |
| `docs/ru/design/DP-013-runtime-management-routing.md` | full | 100644 | `9f7a0a599650cda12ef94dc92dce729b7c7b94bf` |
| `docs/ru/design/DP-027-runtime-diagnostics-v1.md` | full | 100644 | `f1ae8bf1af55f98253ab426d79bbe4d232d693b6` |
| `docs/ru/design/README.md` | full | 100644 | `7fe6f6039b96c4fc2dcae7b7f65d81b340758acb` |
| `docs/ru/roadmap/MASTER_PLAN.md` | full | 100644 | `0efbf564d67f9a4ee3d8854b86527872659784d2` |
| `docs/tasks/README.md` | full | 100644 | `2696758adb16bab8ad49985cf43536e97492ab0a` |
| `docs/tasks/TASK-121-RUNTIME-DIAGNOSTICS-C2-BINDING-DESIGN.md` | task-record-v1 | 100644 | `78efa0731a359be362649e5ae08d17e6aa0e97ef` |
| `spec/current-state.md` | full | 100644 | `dc7fe5075f5bee64bf413a8b3c9800454617ca23` |

No path is staged. After this freeze only append-only evidence below this heading is
allowed without fresh affected gates; any change to the projected prefix or another
full row invalidates Verification, Review and Acceptance.

### M121-002 — Canonical task-record-v1 correction; M121-001 superseded

M121-001 used only the prefix before the envelope and therefore omitted the mandatory
PROCESS-001 Status-body exclusion marker. Its manifest `5a441941...` is invalid and
must not be used for Verification, Review or Acceptance. No projected design byte or
other full path changed; this entry corrects only the append-only evidence identity.

Canonical byte reconstruction finds `## Status`, `## Task Contract` and the terminal
`## Recovery Evidence Envelope` exactly once and in order, at raw task-record byte
offsets 54, 198 and 17597. `task-record-v1` is:

1. raw BOF through the actual LF terminating `## Status`;
2. exact UTF-8 `STATUS-EVIDENCE-EXCLUDED` followed by one NUL;
3. raw bytes from `## Task Contract` through the byte before the envelope heading.

The corrected projection is 17488 bytes with blob OID
`a91c9d7e96027fb98e035ba641c6d14ebff13d2c`. The other 17 full row OIDs and the
18-path ordinal order recorded in M121-001 are unchanged. Correct canonical manifest
length remains 1889 bytes; authoritative Git blob OID is
**`c5ed9f2c076efc10a666a8ffc701954f71fb1610`**.

Independent Verification and Final Review against M121-001 are invalidated and must
recompute M121-002. This correction is inside the excluded append-only envelope and
does not mutate the design projection.

### M121-003 — Stable project-state projection; M121-002 superseded

Before Acceptance, three current-state full rows were corrected from the transient
claim that gates were still unfinished to the stable rule that actual Verification,
Review and Acceptance resolve only from the newest valid matching TASK-121 envelope.
This prevents the projected `In Progress` state from becoming false after a later
decision. No design contract, task-record-v1 projection or other full row changed.

Changed full rows:

| Path | Projection | Mode | OID |
| --- | --- | --- | --- |
| `.ai/PROJECT_CONTEXT.md` | full | 100644 | `29ad3b8cbfecf75e5430a648dcbddad3a7f7ed46` |
| `docs/tasks/README.md` | full | 100644 | `8b1bd81ef609081e6fd511e3ba1049b6f276f4f9` |
| `spec/current-state.md` | full | 100644 | `7863c4d7fe50ab5595845e053fab4fd1d36918b9` |

The canonical task projection remains 17488 bytes/OID
`a91c9d7e96027fb98e035ba641c6d14ebff13d2c`; the other 14 full rows remain exactly
M121-001's recorded OIDs. Correct M121-003 manifest is 18 paths/1889 bytes with
authoritative OID **`b89c606c0e9cca2d7a57b43767474210fa55762d`**. M121-002 is superseded for
downstream gates. Verification and Review must use M121-003.

### V121-003 — Independent Verification and PROCESS-002

Independent Verifier `/root/verify121`: **PASS WITH LIMITATIONS**, blocking and
non-blocking findings 0/0 on exact M121-003
`b89c606c0e9cca2d7a57b43767474210fa55762d` (18 paths/1889 bytes), canonical
task projection 17488 bytes/OID `a91c9d7e96027fb98e035ba641c6d14ebff13d2c`.
The three stable current-state rows and all other rows matched independently.

Source conformance PASS: ARCH-004 and the sole current `Owner -> runtime.Launch` seam
justify Lifecycle Owner; current `owner.dependencies` truth, future private copy,
fail-closed nonzero base Producer, lifecycle/retirement/revocation, authorization
separation and P01-P08 dispositions are coherent. PROCESS-002 PASS: UTF-8/BOM/NUL,
local links, EN/RU parity, project state, stale/malformed identity, diff/staging,
source/module and TASK-102 guards all pass.

Limitations: P10R-120-001 is synchronized accepted terminal evidence, not a new native
replay; product build/test/vet/race/tidy are N/A and not claimed; checkout autocrlf
messages are informational; no executable P proof, production wiring or activation is
inferred.

### R121-003 / S121-003 — Independent Final Review and Scope Audit

Independent non-author Reviewer `/root/review121`: **APPROVED**, blocking and
non-blocking findings 0/0 on exact M121-003. Scope Audit:
**18 Required / 0 Questionable / 0 Removable**. All three changed current-state rows,
task projection and remaining rows matched independently; M121-001/M121-002 are
correctly superseded.

Reviewer confirms the owner, immutable binding, fail-closed base, lifecycle/Stop/
revocation/loss and authorization boundaries; honest P dispositions; no C2/C3
activation or Gate 11 closure. The >15-path Size Guard is accepted because the ten
task/state/navigation paths and eight mirrored Draft seam paths form one indivisible
design/PROCESS-002 subject; removing any path leaves a contradiction or breaks parity.

### CA121-001 — Coordinator Acceptance Verdict

Coordinator `/root`: **ACCEPTED — DESIGN-ONLY C2-BINDING, PRODUCTION NOT ACTIVATED**,
blocking findings 0, exact M121-003
`b89c606c0e9cca2d7a57b43767474210fa55762d` (18 paths/1889 bytes).

Accepted design selects the existing per-Runtime-Instance Lifecycle Owner as the
private owner of one diagnostics lifetime per accepted Host launch in the future
single-process Control Service topology. It requires zero base Producer admission,
one attempt-private bindings copy, same-Host/attempt retention, distinct later pairs,
private exact-attempt handle, truthful retirement/revocation and no lifecycle join.
`ActionObserve` and DP-018 are not consume authority.

P01/P02/P06 are Satisfied by Design. P03/P04/P05/P07/P08 remain Deferred and require
C2-Exposure and/or executable production composition. Acceptance supplies no code,
Host/Bootstrap/Control Service wiring, endpoint, transport, worker, retry, durable or
lossless guarantee. C2/C3 remain Not Activated, Gate 11 OPEN and TASK-102 untouched.

Commit Gate and Publication Gate are **NOT AUTHORIZED / NOT PERFORMED**. A later
isolated C2-Binding implementation intake is the recommended candidate, Ready only
for separate repository-first intake and Owner authorization; it is not activated by
this decision. Independent post-decision final-byte integrity remains required before
handoff.

### M121-004 — Minimal EOL rebase candidate; M121-003 preserved

Owner authorized a bounded documentation-only EOL rebase after Commit Readiness
proved that exact-byte staging of M121-003 would introduce EOL-only changes in the
eight DP-010--013 EN/RU mirrors. M121-003 and its Acceptance evidence above remain
unaltered historical evidence; this entry does not reinterpret or overwrite them.

Only CR bytes belonging to CRLF pairs were removed from those eight mirrors. Their
LF-normalized semantic bytes were identical before and after the rewrite. Design
text, scope, acceptance criteria, task-record-v1 projection, the other ten subject
paths, repository configuration, `.gitattributes` policy and production code were
not changed.

Changed full rows:

| Path | Projection | Mode | OID |
| --- | --- | --- | --- |
| `docs/en/design/DP-010-runtime-lifecycle-owner-contract.md` | full | 100644 | `6c04e293de7f0c95e3a8142e0fee8cab975a170e` |
| `docs/en/design/DP-011-runtime-launch-pipeline-integration.md` | full | 100644 | `665ce5579ea390ff8816d72cd94ec3b6c41bbc16` |
| `docs/en/design/DP-012-runtime-source-composition.md` | full | 100644 | `a30f6a43c87018412a88f1e9ce08e26e7f99d23d` |
| `docs/en/design/DP-013-runtime-management-routing.md` | full | 100644 | `8db6a1efd64746d024e7f097016ab684e045c8c0` |
| `docs/ru/design/DP-010-runtime-lifecycle-owner-contract.md` | full | 100644 | `09dd3ec1d05640a28fe1455c8244ec545060dca3` |
| `docs/ru/design/DP-011-runtime-launch-pipeline-integration.md` | full | 100644 | `b30786b552122125f35aaf1dbbe5e8e72ce3f0c2` |
| `docs/ru/design/DP-012-runtime-source-composition.md` | full | 100644 | `e7bb0473729a93dc3b2833ed8bb441918c26d778` |
| `docs/ru/design/DP-013-runtime-management-routing.md` | full | 100644 | `0b67df7465c6288ce5e60c3323b88e885e5ec8c7` |

The other nine full rows remain byte-identical to M121-003. The canonical task
projection remains 17488 bytes/OID
`a91c9d7e96027fb98e035ba641c6d14ebff13d2c`. Candidate M121-004 is exactly 18
paths/1889 bytes with canonical manifest OID
**`aecd71fbac793ce389006dd9ac54c763b679367d`**. No gate result or Acceptance is
claimed by this freeze; all affected gates must use M121-004.

### A121-004 — Architecture Conformance after minimal EOL rebase

Independent Architect `/root/architecture121`: **PASS**, blocking and non-blocking
findings 0/0 on exact M121-004
`aecd71fbac793ce389006dd9ac54c763b679367d` (18 paths/1889 bytes), task projection
17488 bytes/OID `a91c9d7e96027fb98e035ba641c6d14ebff13d2c`.

All 18 rows matched independently. The eight LF-rebased DP-010--013 EN/RU mirrors
retain the accepted architecture: Lifecycle Owner, zero-base fail-closed admission,
attempt-private binding copy and handle, same-Host retention, distinct later pair,
retirement/revocation/loss, no fifth Bootstrap binding or public Host capability,
ActionObserve/DP-018 separation and unchanged P01--P08 dispositions. No design/scope
drift, production/source/module/Frozen/Approved mutation, C2/C3 activation, Gate 11
closure or TASK-102 effect was found.

### E121-004 — Exact-byte temporary-index proof

Coordinator simulated `git -c core.autocrlf=false add -- <exact 18 paths>` in an
isolated temporary index and object directory initialized from HEAD. The simulated
index contained exactly 18 changed paths; every staged blob was byte-identical to its
worktree file, all 18 projection-aware rows matched M121-004, and the reconstructed
manifest was 1889 bytes/OID
`aecd71fbac793ce389006dd9ac54c763b679367d`. The resulting pre-evidence tree was
`18854c2a3f1e2e5f59ae14eb828a87f79154bc88`.

Normal and `--ignore-space-at-eol` numstat were byte-identical, proving no EOL-only
migration in the candidate tree. Temporary state was removed. Real index, worktree,
HEAD, branch, repository config and absent `.gitattributes` remained unchanged. A
fresh proof is required after subsequent allowed envelope appends and before
Acceptance; this entry does not self-attest later task-record bytes.

### V121-004 — Independent Verification and PROCESS-002

Independent Verifier `/root/verify121`: **PASS WITH LIMITATIONS**, blocking and
non-blocking findings 0/0 on exact M121-004. The verifier independently reconstructed
18 rows, 1889 manifest bytes/OID
`aecd71fbac793ce389006dd9ac54c763b679367d`, task-record-v1 17488 bytes/OID
`a91c9d7e96027fb98e035ba641c6d14ebff13d2c` and heading offsets 54/198/17597.

The eight rebased rows are LF-only with no CR, match M121-004 exactly and have equal
raw and clean-filter OIDs. M121-003 Acceptance remains preserved historically; the
task projection and other nine full rows are unchanged. PROCESS-002 **SYNCHRONIZED**:
all 18 paths are strict UTF-8 without BOM/NUL, 494 local links have zero missing,
five mirrored design pairs have equal heading structure, EN/RU semantics and
project-state/TASK-120 terminal facts remain aligned, and `git diff --check` has no
errors. There are no staged, source, module, TASK-102, malformed-OID or stale
publication changes.

Limitations remain truthful: P10R-120-001 is synchronized terminal evidence, not a
new native replay; product build/test/vet/race/tidy are N/A for this design-only
rebase; future-checkout autocrlf warnings are informational; no executable P proof,
wiring or production activation is inferred.

### S121-004 — Primary Scope Audit after rebase

Coordinator classification remains **18 Required / 0 Questionable / 0 Removable**.
The rebase changes no path set or semantic responsibility: ten task/state/navigation
paths preserve recovery, publication and roadmap truth; eight mirrored Draft paths
form the four production ownership seams required by DP-027. Only those eight
mirrors changed raw OID, solely to restore their HEAD-compatible LF representation.
Removing a path still breaks recovery, project-state synchronization, EN/RU parity
or one required seam. No generated, temporary, code, test, module, policy or
TASK-102 path is present. The Size Guard remains one indivisible documentation design
subject and awaits independent Final Review.

### R121-004 — Independent Final Review and Scope Audit

Independent non-author Reviewer `/root/review121`: **APPROVED**, blocking and
non-blocking findings 0/0 on exact M121-004
`aecd71fbac793ce389006dd9ac54c763b679367d` (18 paths/1889 bytes), task projection
17488 bytes/OID `a91c9d7e96027fb98e035ba641c6d14ebff13d2c`.

The Reviewer independently confirmed all eight mirrors are LF-only (`CR=0`), each
raw OID equals its clean-filter OID and M121-004 row, `git ls-files --eol` reports
`i/lf w/lf`, normal and ignore-EOL numstat are identical, and `git diff --check`
has no errors. Architecture semantics, P01--P08 dispositions, PROCESS-002 and EN/RU
parity remain unchanged; M121-003 historical evidence is preserved.

Independent Scope Audit: **18 Required / 0 Questionable / 0 Removable**. The >15
Size Guard remains justified as one indivisible task/state/navigation plus mirrored
design subject. No production, module, TASK-102, generated, temporary, staged or
policy path is present. Production C2/C3 remain Not Activated and Gate 11 OPEN.

### E121-005 — Final pre-Acceptance staging proof

After all Architecture, Verification, PROCESS-002, Scope and Review evidence above,
Coordinator repeated the isolated exact-byte staging simulation on the current task
record. PASS: 18 changed paths and exact path set; every staged blob byte-identical to
worktree; canonical manifest 1889 bytes/OID
`aecd71fbac793ce389006dd9ac54c763b679367d`; task full blob
`adc72aeae7c4b9d96e2098bf113d61035fbf22bc`; task projection
`a91c9d7e96027fb98e035ba641c6d14ebff13d2c`; resulting tree
`b110ba56476d073ab79ffd50d5edd8568ac6cff4`. Normal and ignore-EOL numstat were
identical. Real index, worktree, config, absent `.gitattributes`, HEAD and branch were
unchanged; temporary state was removed.

### CA121-002 — Coordinator Re-Acceptance after minimal EOL rebase

Coordinator `/root`: **ACCEPTED — DESIGN-ONLY C2-BINDING, GIT-CLEAN EOL REBASE,
PRODUCTION NOT ACTIVATED**, blocking findings 0 on exact M121-004
`aecd71fbac793ce389006dd9ac54c763b679367d` (18 paths/1889 bytes), task projection
17488 bytes/OID `a91c9d7e96027fb98e035ba641c6d14ebff13d2c`.

M121-003 remains the preserved historical accepted subject. M121-004 changes only the
raw EOL representation of eight DP-010--013 EN/RU mirrors, with unchanged design,
scope and acceptance criteria. Architecture Conformance PASS 0/0, Independent
Verification PASS 0/0, PROCESS-002 Synchronized, EN/RU parity PASS, Scope Audit
18/0/0, Independent Final Review APPROVED 0/0 and final pre-Acceptance staging proof
PASS all resolve on M121-004.

Acceptance adds no production code, Host/Bootstrap/Control Service wiring, exposure,
C3 capability, executable P proof, durable/lossless guarantee or lifecycle join.
Production C2/C3 remain Not Activated, Gate 11 OPEN and TASK-102 untouched. Commit
and Publication Gates are **NOT AUTHORIZED / NOT PERFORMED**. Final post-decision
byte integrity remains required and cannot self-attest this envelope append.

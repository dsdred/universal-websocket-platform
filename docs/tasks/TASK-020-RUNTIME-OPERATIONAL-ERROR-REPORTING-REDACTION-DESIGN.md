# TASK-020 — Runtime Operational Error Reporting and Redaction Design

**Status:** Completed — Coordinator Accepted

## Task Contract

- **Task Mode:** Design-only.
- **Why Now:** TASK-019 completed the dependency-ordered ARCH-004 section
  19(5) recovery candidate. The authoritative project state and both
  MASTER_PLAN mirrors identify section 19(6), operational error reporting and
  redaction, as the next bounded design prerequisite. Management
  implementation remains blocked until sections 19(2)–(6) have focused
  contracts and separate status decisions.
- **Definition of Done:** a mirrored, non-normative Draft/Planned DP defines a
  technology-neutral reporting boundary for lifecycle, command, activation,
  and recovery failures; stable operator-safe categories and correlations;
  separation of durable domain truth from report delivery; mandatory
  redaction and scope isolation; replay, cancellation, concurrency, and
  delivery-failure behavior; acceptance proofs; formal gate honesty; updated
  indexes, roadmap, task index, and project-state documents; applicable
  verification, PROCESS-002, Scope Audit, and final review complete.
- **Out of Scope:** production code or tests; public HTTP/API/DTO mapping;
  logging, metrics, tracing, audit, alerting, or storage product selection;
  concrete authorization policy; retention/deletion; secret classification
  outside existing boundaries; approval/status promotion of DP-014–DP-018;
  persistence schema, recovery executor, management wiring, Production
  Activation, automatic restart, scheduling, or supervision.
- **Verification Plan:** Existing Coverage Report before any test change;
  EN/RU heading and code-fence parity; semantic review of failure ownership,
  redaction, scope isolation, replay, and indeterminate delivery; changed-link
  validation; full Go regression and vet as documentation-only safety checks;
  conflict-marker, whitespace, generated-file, and exact-scope checks;
  independent final review required by PROCESS-001.

## Selection Evidence

- baseline: clean synchronized `main@a3b931e`; no active `In Progress` or
  `Blocked` task was found in the task index or project-state documents;
- TASK-019 is published through PR #20 and merged at `a3b931e`;
- `.ai/PROJECT_CONTEXT.md`, `spec/current-state.md`, `spec/decisions.md`, and
  both MASTER_PLAN mirrors converge on ARCH-004 section 19(6) as the next
  dependency-ordered candidate;
- selected slice: one focused design contract for operational error reporting
  and redaction, with no implementation;
- rejected alternative — management implementation: not Ready while the
  section 19(6) design and section 19(2)–(6) status decisions are absent;
- rejected alternative — Production Activation/API wiring: downstream and
  materially larger than the smallest independently reviewable slice;
- rejected alternative — metrics/audit/alerting backends: product and
  technology choices explicitly deferred by ARCH-004 and the predecessor DPs.

## Sources of Truth

- `docs/engineering/PROCESS-001-AI-DEVELOPMENT-WORKFLOW.md`;
- `docs/engineering/PROCESS-002-DOCUMENTATION-SYNCHRONIZATION.md`;
- `docs/en|ru/architecture/ARCH-001-runtime-architectural-pattern.md`;
- `docs/en|ru/architecture/ARCH-002-runtime-foundation-freeze.md`;
- `docs/en|ru/architecture/ARCH-004-runtime-deployment-and-identity-model.md`;
- `docs/en|ru/architecture/ARCH-005-runtime-configuration-snapshot-and-loading-model.md`;
- `docs/en|ru/design/DP-011-runtime-launch-pipeline-integration.md`;
- `docs/en|ru/design/DP-013-runtime-management-routing.md`;
- `docs/en|ru/design/DP-014-runtime-operational-identity-persistence.md`;
- `docs/en|ru/design/DP-015-runtime-management-command-idempotency.md`;
- `docs/en|ru/design/DP-016-runtime-activation-replacement-rollback.md`;
- `docs/en|ru/design/DP-017-runtime-recovery-reconciliation.md`;
- `.ai/PROJECT_CONTEXT.md`, `spec/current-state.md`, `spec/decisions.md`, and
  both MASTER_PLAN mirrors.

## Roles and Handoffs

- **Coordinator:** selection, contract, gates, Scope Audit, acceptance, and
  project-state closure.
- **Documentation Baseline:** document inventory, status/parity/link drift, and
  implemented-versus-planned boundary.
- **Architect:** define the focused section 19(6) candidate and confirm that it
  preserves approved ownership and lifecycle invariants.
- **Documentation Agent:** create mirrored DP-018 and synchronize navigation,
  predecessor readiness wording, roadmap, and project state.
- **Tester:** verify parity, links, regression safety, diff quality, and the
  risk-based matrix; no test creation is authorized by this contract.
- **Reviewer:** independently assess source precedence, completeness,
  redaction, failure truth, scope, and removability.

## Branch Decision and Stop Conditions

- branch: `docs/task-020-runtime-operational-error-reporting-redaction-design`;
- this task record is the first content change on the branch;
- stop on conflict with Approved/Active architecture, missing error ownership
  boundary, ambiguous public disclosure policy, critical EN/RU drift, failed
  mandatory verification, blocking review finding, or required production
  implementation.

## Size Guard

- expected scope may update the mirrored DP, predecessor readiness/gate
  wording, indexes, roadmap, and project-state documents;
- no production code, test, package, or independently shippable behavior;
- if the diff exceeds 15 files, Coordinator must re-prove that every file is
  mandatory contract/navigation/state synchronization or split the work.

**Final re-proof:** triggered at 21 documentation files. The slice remains one
independently reviewable behavior: the mirrored section 19(6) candidate.
Splitting would leave predecessor readiness, EN/RU navigation, roadmap, task
index, or mandatory project state knowingly stale. Production lines/packages,
tests, generated files, API contracts, and shippable behavior are zero.

## Existing Coverage Report

- **Existing Coverage:** repository Go tests cover implemented isolated
  runtime packages; documentation has mirrored indexes, stable heading/fence
  checks used by prior design tasks, link validation, and diff checks.
- **Coverage Gap:** no current normative or Draft document fully defines the
  section 19(6) reporting/redaction contract or its cross-DP acceptance proofs.
- **Added Proof Tests:** none planned; this task changes no executable
  behavior.
- **Added Regression Tests:** none planned; full existing Go regression is a
  safety check only.
- **Remaining Limitations:** design verification cannot prove future storage,
  delivery, transport, or redaction implementation.

## Documentation Baseline

- ARCH-004 is Active and lists reporting/redaction as the final focused
  section 19 prerequisite without selecting a diagnostics backend;
- DP-013–DP-017 are mirrored Draft/Planned candidates and consistently defer
  concrete reporting/redaction to section 19(6);
- ARCH-001 distinguishes valid negative decisions from operational errors and
  forbids silently discarded lifecycle failures; ARCH-002 preserves startup
  and rollback causes; ARCH-005 forbids Secret values in configuration
  provenance or diagnostics;
- design indexes are EN/RU aligned through DP-017; task index and project-state
  documents consistently recommend section 19(6);
- implemented state remains isolated runtime components only; no reporting,
  persistence, management, recovery, or Production Activation implementation
  exists;
- baseline result: `Synchronized`; no critical drift blocks design work.

## Architecture Confirmation

- **Result:** Design READY; blockers 0.
- The reporting boundary may consume immutable, authorized domain facts and an
  ephemeral internal cause, but it must not mutate lifecycle/command truth,
  reopen admission, or create a second durable outcome.
- Stable operator-visible categories are a deliberately smaller projection of
  domain outcomes; they are not universal component error types and never
  replace exact internal error identity or cause chains at their owners.
- A report is correlated by opaque scoped identities and operation/phase, not
  by payload, address, PID, Host pointer, permit, stack trace, or raw error
  text.
- Redaction is allowlist projection at the trust boundary. Unknown fields and
  unclassified causes fail closed; secret values, raw configuration/Snapshot,
  authorization detail, cross-scope existence, and unrestricted process
  metadata are never report content.
- Report publication occurs only after the corresponding domain fact is
  committed or authoritatively observed. Delivery failure cannot change that
  fact, cause lifecycle replay, or block command admission; it remains a
  separately observable reporting failure.
- Metrics, traces, logs, audit retention, alert routing, transport mapping,
  storage technology, and public API shapes remain adapters or later focused
  decisions.

## Current Handoff

Architecture and Documentation work, Tester verification, PROCESS-002, Scope
Audit, terminal design/closure review, project-state synchronization, and
Coordinator Acceptance are complete. Commit and publication remain outside
this command.

## Next Candidate

Not activated. After TASK-020, the repository must reassess formal status
decisions for the section 19(2)–(6) candidate set before selecting any
management implementation slice.

## Verification Matrix

| Risk | Required proof | Result |
| --- | --- | --- |
| valid negative outcome relabeled as error | explicit exclusion and owner mapping EN/RU | PASS after B-001/B-003 rework |
| overlapping category breaks replay | exact owner/phase precedence and overlap cuts | PASS after B-005 rework |
| report leaks Secret/payload/cross-scope fact | allowlist construction, forbidden-content list, authorization rules | PASS |
| replay changes across occurrence or schema | no occurrence marker; invariant scoped to fact + projection version | PASS after B-002/B-004 rework |
| delivery failure mutates domain truth | downstream-only delivery and non-recursion rules | PASS |
| recovery report fabricates liveness/release | DP-017 evidence boundary and blocked classification | PASS |
| EN/RU drift | headings, fences, matrix and semantic parity | PASS, 27/27 and 2/2 |
| code regression from documentation-only diff | full Go tests and vet | PASS |
| broken navigation or malformed diff | repository links, conflict and diff checks | PASS |

## Verification Results

- `go test ./... -count=1` — PASS, all packages;
- `go vet ./...` — PASS;
- DP-018 headings EN/RU — 27/27;
- DP-018 code fences EN/RU — 2/2;
- changed/new and repository relative links — 0 failures;
- `git diff --check` — PASS; platform LF-to-CRLF notices informational;
- conflict markers — 0;
- production code/tests/generated files — unchanged/absent;
- race/stress — Not applicable: concurrency implementation and tests are
  unchanged; design acceptance proofs require them for future implementation.

## Review and Rework History

- Initial independent design review — blocking B-001/B-002: valid negatives
  leaked into error reporting and occurrence state contradicted deterministic
  replay;
- first rework — separated authorization denial/malformed/not-found/conflict
  from operational failures and removed first/replay semantic content;
- Repeat review — blocking B-003/B-004: residual validation/cancellation
  ambiguity and unpinned cross-schema replay promise;
- second rework — ordinary validation/cancellation explicitly excluded;
  deterministic identity scoped to source fact plus projection version;
- Repeat review — blocking B-005: overlapping Source/Preparation,
  persistence/indeterminate, and recovery/indeterminate categories;
- third rework — exact owner/phase precedence, overlap rules, matrix, and proof;
- final matrix clarity — added `PersistenceUnavailable` row;
- Terminal Design Reviewer — `Approved`, 0 blocking / 0 nonblocking;
- initial terminal Closure Review — Needs Revision: premature project-state
  completion claim;
- process-truth rework — state restored to factual In Progress; residual stale
  transition wording corrected;
- Repeat terminal Closure Review — `Approved`, 0 blocking / 0 nonblocking.

## PROCESS-002 and Applicability

- DP-018 mirrors, predecessor gate wording, design indexes, MASTER_PLAN
  mirrors, task index, `.ai/PROJECT_CONTEXT.md`, `spec/current-state.md`, and
  `spec/decisions.md` — Required and synchronized;
- Design/Implementation statuses remain Draft/Planned; no source represents
  candidate architecture as approved or implemented;
- root README — Not applicable: project entry, install, and shipped capability
  did not change;
- `CHANGELOG.md` — Not applicable: no user-facing or release behavior;
- production/API documentation — Not applicable: no implementation or public
  contract;
- result: `Synchronized`; next agent can reconstruct scope and gates from the
  repository alone.

## Scope Audit

- **Required — 21:** task record (1); DP-018 mirrors (2); DP-013–DP-017 gate
  mirrors (10); design indexes (2); MASTER_PLAN mirrors (2); task index (1);
  `.ai/PROJECT_CONTEXT.md`, `spec/current-state.md`, `spec/decisions.md` (3).
- **Questionable — 0.**
- **Removable — 0.**
- Exact-scope proof: all files are documentation; each is the candidate
  contract, mandatory EN/RU/navigation/predecessor synchronization, or durable
  task/project-state evidence. No unrelated, temporary, generated, staged, or
  hidden file is included.

## Closure Handoff

- Definition of Done mapped: mirrored candidate, ownership/classification,
  correlation/redaction, replay/delivery, acceptance proofs, formal gates,
  navigation/state synchronization, verification, and Scope Audit complete;
- production/test changes: none;
- unresolved design findings: none;
- remaining limitation: future implementation must provide executable
  redaction corpus, concurrency, replay, failure-injection, and adapter-failure
  proofs;
- Coordinator Acceptance: granted after Repeat terminal Closure Review
  `Approved` 0/0; Definition of Done satisfied and Scope Audit accepted 21/0/0;
- commit readiness: exact accepted documentation diff is ready only for a
  separate `Разрешаю коммит.` gate; stage/commit/push/PR/merge were not
  performed.

## Commit Gate

- exact command `Разрешаю коммит.` received after Coordinator Acceptance: yes;
- commit message policy: Conventional Commits;
- selected message: `docs(runtime): define operational error reporting`;
- exact accepted file set: 21 documentation files from Scope Audit;
- post-acceptance changes: only this bounded permission/Commit Gate record;
  design contract, project-state semantics, and scope are unchanged;
- temporary/generated/unrelated files: absent;
- one task commit is authorized; push, PR, merge, and publication are not
  authorized by this command.

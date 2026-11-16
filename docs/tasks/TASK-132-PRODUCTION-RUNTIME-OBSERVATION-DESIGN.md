# TASK-132 — Production Runtime Observation Design & Architecture Admission

## Status

In Progress. Exact verdict, subject and first incomplete gate resolve only from
the newest matching Recovery Evidence Envelope and recomputed canonical manifest.
Design-only; no implementation or protected-stage activation.

## Task Contract

- Task Mode: Design-only, explicit Owner direction 2026-11-16.
- Why Now: TASK-131 read-only admission identified absent production observation
  composition and no post-Running Host terminal signal. This task reconstructs
  those facts from repository sources; the prior chat report is not authority.
- Repository: E:/wikiPRJ/universal-websocket-platform.
- Branch: docs/task-132-runtime-observation-design.
- Trusted baseline/current HEAD: 88fb333453ab9563285d65cd44a8b0a4019f447a.
- Definition of Done: reproducible forensic evidence; three architecture options;
  proposed observation contract, ten failure scenarios, explicit ADMISSIBLE or
  BLOCKED implementation admission with Owner options; mirrored draft ADR;
  minimal conditional roadmap; independent review; exact documentation manifest.
- Out of Scope: production/test code, endpoints, lifecycle changes, wiring,
  Dashboard/branding/README, TASK-102, TASK-131 activation, Stage B/B1/B2,
  Production C2/C3, C2-Exposure, Gate 11 closure; stage/commit/push/PR/merge.
- Gate 11 OPEN; protected stages Not Activated. No automatic successor task.
- Verification Plan: raw baseline/source/line evidence, contract contradiction
  audit, EN/RU semantic parity and links, failure/coverage matrix, exact diff,
  unchanged-line EOL integrity, protected paths/refs, canonical manifest and
  independent final review. No service launch or executable qualification.

## Intake / Ownership / Roles

Clean main and origin/main locally equal baseline; Owner identifies TASK-130
terminal merge at that OID. No fresh remote/P10 reconstruction is claimed here.
Local safe branch created before this first content change. Current root is
Coordinator and is explicitly assigned sequential Documentation Agent and
documentation verification responsibilities. Separate Architect performs design;
separate non-author Reviewer performs independent review. No Publisher authority.
TASK-102 protected branch is 8bfa2e22a4b6310a71f6d36c8724b1adedd727ad;
only git-show inspection allowed, no checkout/build/run/edit of its sources.

Order: intake -> documentation baseline -> Architect forensic/design admission
-> documentation of proposed decision -> documentation verification -> PROCESS-002
-> scope audit -> independent review/rework -> Coordinator design disposition
-> Pre-Commit STOP. If proposed implementation is BLOCKED, stop with evidence
and Owner options; do not create implementation task or grant product Acceptance.

## Sources / Existing Coverage / Size Guard

AGENTS/AGENT, PROCESS-001/002 and role contracts; .ai/PROJECT_CONTEXT,
spec README/current-state/decisions; ADR-0002/0003/0004, ARCH-002/004/005,
DP-010/013/017/022/027, TASK-102/126/127/130; Runtime Host/Owner/Directory,
Control composition, configuration Publish, HTTP/Studio/Dashboard and tests.
Existing coverage: lifecycle/Directory unit tests prove isolated snapshots,
authorization and launch/stop ordering; existing Dashboard tests keep publication
and health separate. Gap: no production observation graph or current-liveness
proof. Added proof/regression tests: none planned (design-only); conceptual
acceptance scenarios do not claim executed production proofs. Go/race/browser
tests N/A for documentation-only changes; no implementation validation claimed.
Size Guard: one proposed architecture contract, zero production LOC/packages;
expected <=10 documentation paths. Reassess above 15 files or expanded contract.

## Planned Documentation Scope / Applicability

Required: this task record, mirrored ADR-0005 and ADR indexes, task index,
.ai context and spec/current-state (current design task). Potentially
spec/decisions for the explicitly pending proposal. MASTER_PLAN mirrors N/A
unless dependency/status is actually changed; this task proposes conditional
roadmap without activation. Existing DP/ARCH/ADR bodies remain authoritative
and unchanged: new ADR stays Draft, cannot override them. CHANGELOG and root
README N/A: no user-facing capability or release change. Historical records
including TASK-102/130 remain untouched. No new status document asserts mutable
review/Acceptance outside the envelope.

## Repository Forensic Report (baseline above)

Line anchors below refer to unchanged baseline sources; ranges identify the
relevant implementation, not proof of behavior outside that source boundary.

| Evidence | Symbol / relevant lines | Verified fact |
|---|---|---|
| `internal/runtimelifecycle/types.go` | ActualState 34–41; Observation 224–247 | Five lifecycle states and detached per-Instance identity/attempt facts; no observation timestamp, expiry or process-health proof. |
| `internal/runtimelifecycle/owner.go` | NewOwner 63–108; Observe 619–642 | One process-local Owner, initially stopped; coherent mutex snapshot. Constructing a new Owner cannot discover an already running Runtime. |
| Same | runLaunch 504–574; runHostStop 576–615 | Running follows successful launch; Stop success/failure retains existing ordering and outcomes. No post-launch terminal watcher is installed. |
| `internal/runtimemanagement/directory.go` | Target 40–78; Observe 198–210; admit 213–243 | Exact Workspace/Configuration/Instance targeting; authorization before Owner access, but internal scope lookup precedes authorization. No production authentication/discovery API. |
| `internal/runtime/host.go` | Host 30–40; Start 212–261; Stop 300–379; Running/Ready 384–396 | Host owns lifecycle/readiness/context; interface has no terminal completion capability. Accessors read host.state, not a traffic probe. |
| `internal/listener/listener.go` | Running 78–83; serve 215–223 | Listener accessor reads stored state; Serve return offers a diagnostic and calls a legacy reporter only for unexpected nonnil errors. It does not publish Host/Owner termination. |
| `internal/runtime/startup_transaction.go` | acquire/commit/rollback 14–43 | Existing rollback ownership and reverse ordering; observation must not run or wait inside this path. |
| `internal/configurationversion/service.go` | Publish 275–320 | Publication/archive updates only; no Runtime launch. |
| `cmd/control-service/main.go` | run 40–57 | Only memory workspace/configuration/version services and Studio/CRUD handlers; no Runtime Owner/Directory graph. |
| `internal/http/server.go` | New 19–35; health 37–41 | Control health is a fixed JSON response, not Runtime inspection. |
| `internal/configurationstudio/assets/dashboard.html` | availability 30–36 | Runtime is already Unavailable; publication and health are explicitly separate. |
| `internal/runtimediagnosticspull/pull.go` | PollResult/View 15–30 | Destructive lossy diagnostics; NoItem does not prove complete absence of errors or liveness. |
| `internal/runtimemanagement/diagnostics.go` | DiagnosticsConsume 18–19 | Consume authority is distinct from Observe. |
| `docs/ru/architecture/ARCH-002-runtime-foundation-freeze.md` | sections 2/4/8/11/13 | Frozen lifecycle/readiness unchanged; focused additive supervision design is possible, not implicitly implemented or approved. |
| `docs/ru/architecture/ARCH-004-runtime-deployment-and-identity-model.md` | sections 6–8/12/14/17 | Instance/attempt identity distinct from version/PID; initial topology in-process; shutdown proof distinct from absence of execution. |
| `docs/ru/design/DP-010-runtime-lifecycle-owner-contract.md` | section 27, 740–747; section 29 | Unexpected termination remains integration-gated; polling Running cannot manufacture Failed. |
| `docs/ru/design/DP-013-runtime-management-routing.md` | sections 18/28–30 | Isolated Observe exists, HTTP/auth/instance discovery and production wiring are deferred. |
| `docs/ru/design/DP-022-runtime-execution-containment-and-evidence.md` | section 17, 571–580 | Default undeclared guarantee gives Unknown; containment evidence is not Runtime readiness. |
| `docs/ru/design/DP-027-runtime-diagnostics-v1.md` | section 38, 1463–1473 | Stage B/B1/B2/production C2/C3/exposure stay unactivated; transport is not lifecycle observation. |
| `docs/ru/roadmap/MASTER_PLAN.md` | section 5, Gate 11 | Operational diagnostics requirements remain open; observation design does not close the gate. |

TASK-102 source: `git show feature/task-102-lpac-process-admission:docs/tasks/TASK-102-LPAC-PROCESS-ADMISSION-SUBSTRATE.md`,
Task Contract lines 10–30. Native LPAC experimental source-only scope forbids
build/run and does not supply a Runtime supervisor. Protected ref remains the
intake OID. No inference is made from its verification-stable Status projection.

Existing executable coverage inspected, not rerun: Host lifecycle/readiness/context
tests; `TestObserveChecksCancellationAfterAuthorization` and
`TestStopAndObserveAuthorizeExactScope`; Listener diagnostics/serve-error tests,
including expected returns, concurrent Stop/failure and reporter panic/blocking.
These prove isolated historical obligations only, not the proposed contract.
Additional restart evidence: `Owner.usedAttempts` in owner.go 107/143/179 is
process-local; cross-restart uniqueness of equal Target/Attempt strings cannot
be assumed. Observation requires separately qualified source provenance.

## Architecture Proposal / Admission / Owner Options

Architect A132-001 selected B+C: retained Host-owned completion evidence plus
read-only validity adapter; full production implementation admission BLOCKED.
The mirrored [ADR-0005](../en/adr/0005-production-runtime-observation.md)
and [RU version](../ru/adr/0005-production-runtime-observation.md) record the
eleven-dimension A/B/C comparison, state/proof-level contract, bounded capture,
identity/restart/expiry rules and all ten Owner-required failure scenarios.

Authoritative actual state remains in Owner. Proposed source facts cannot change
Ready, admission, context, startup/rollback/Stop, resource ownership or publish
Failed/Stopped. Additive evidence may be compatible with Frozen ARCH-002;
absence of a ready production graph/security contract still blocks the full
Dashboard implementation. A supervisor or Stage B is not a universal prerequisite.
Unconsumed diagnostics or an active marker never prove continuous execution.

Owner choices: separately approve or revise Draft ADR; choose sampled lifecycle
with execution Unknown versus a separate stronger supervision/progress design;
choose sanctioned real managed target graph and concrete authentication/policy
boundary. Proposed R1–R5 roadmap is conditional, has no successor Task IDs and
does not assert implementation readiness. TASK-131 remains stopped.

No approved source is amended. Design evidence acceptance, if later recorded,
accepts only this documentation deliverable; ADR remains Draft / Not Implemented
and production admission remains BLOCKED. No BCC or product Completion claim.

R132-01 review rework: Architect A132-003 confirms mandatory opaque
observationSourceIncarnation per source process lifetime. Qualified composition
issues it without cross-restart reuse; it is provenance, not Runtime identity,
execution generation, authority or liveness proof. Authenticated bootstrap and
expectedObservationSourceIncarnation pin participate in capture/response/client
binding. Missing/unqualified identity gives Unknown; mismatch never silently
rebinds. Exact issuer/nonreuse/bootstrap qualification is an explicit additional
production-admission blocker. Mirrored acceptance cases cover reused Target/
Attempt strings, delayed bootstrap/old response, missing issuer, wrong pin,
duplicate source identity and replacement during capture. No code implemented.

## PROCESS-002 / Exact Scope Audit

Nine Required paths; zero Questionable/Removable. Removing any path loses either
Owner design deliverable, mirror/navigation or mandatory recovery/state sync:

1. `.ai/PROJECT_CONTEXT.md`: current design task and local TASK-130 Git facts.
2. `docs/en/adr/0005-production-runtime-observation.md`: proposed contract/design.
3. `docs/en/adr/README.md`: discoverable new Draft ADR.
4. `docs/ru/adr/0005-production-runtime-observation.md`: semantic mirror.
5. `docs/ru/adr/README.md`: mirrored navigation.
6. `docs/tasks/README.md`: current task navigation and historical TASK-130 fact.
7. `docs/tasks/TASK-132-PRODUCTION-RUNTIME-OBSERVATION-DESIGN.md`: contract,
   forensic report, role handoffs and exact recovery evidence.
8. `spec/current-state.md`: current design task; no new capability claim.
9. `spec/decisions.md`: pending proposal clearly separate from accepted decisions.

Existing ARCH/DP/ADR bodies N/A for edits: no design/implementation status or
normative semantics changed; contradiction audit includes ARCH-002/004/005,
ADR-0002/0003/0004, DP-010/013/017/022/027. MASTER_PLAN mirrors N/A: gate,
milestone and active dependency boundaries unchanged; conditional R1–R5 is not
an activated roadmap commitment. Root README/CHANGELOG N/A: no user-facing
feature or release. Public home indexes N/A: existing ADR directory navigation
still resolves; ADR subindexes updated. TASK-102/126/127/130 and all other
historical records untouched. TASK-131 has no repository implementation task;
its findings are independently reconstructed here rather than fabricated closure.

PROCESS-002 synchronization is limited to this design slice. Baseline navigation
still referring to TASK-130 is replaced with current TASK-132 projection; local
merge/source facts are distinguished from Owner-reported P10 and fresh remote
verification. No claim of newly performed Publisher/CI qualification.

## Verification / Pre-Commit Report

Documentation-only result: forensic report, two mirrored Draft ADR files,
comparison/recommendation/contract/failure analysis, BLOCKED production admission
and conditional implementation/test plan. No production/API/UI/test changes.
Size Guard: nine Markdown paths, zero production LOC/packages, one proposed
architecture contract; DO NOT SPLIT this mirrored design evidence deliverable.

Reproducible external checker: `audit132.py` in the Codex visualizations directory
for root chat 01a1e297-4f2e-75d1-8459-1e12a73bddf0. It checks exact nine-path
set including untracked files, baseline/branch/protected ref, empty staged set,
relative links, conflict markers/whitespace, byte equality for unchanged lines
and EOL churn; creates canonical manifest JSON. It hashes raw bytes with
git hash-object --stdin, no filters/normalization. Results and identities belong
only to the matching envelope below; edits require rerun and affected review.

Manual evidence trace and EN/RU semantic parity compare all eleven option rows,
seven state meanings, capture algorithm, bounds, closed reasons, ten failure
rows, five conditional slices and Draft/BLOCKED/nonactivation boundaries.
Existing coverage is documented above. Added proof tests 0; added regression
tests 0. Go/race/build/vet/browser acceptance NOT RUN / N/A for design-only
documentation; no live behavior or implementation qualification claimed.
Future required tests are acceptance criteria, not successful executions.

Commit Gate unentered. No staging/commit/publication authority. STOP after
independent review/disposition with exact documentation manifest and Owner
options; no automatic TASK-131 or successor activation.

## Recovery Evidence Envelope

Intake only. Architecture, verification, review and disposition not yet complete.

### M132-001 / V132-001 / A132-002

Repository/branch/base/HEAD: E:/wikiPRJ/universal-websocket-platform, docs/task-132-runtime-observation-design, 88fb333453ab9563285d65cd44a8b0a4019f447a. Object format sha1.
Canonical raw subject manifest M132-001: `7b98ee35111ba2a26d91288053c921e96e5b624e`.
All rows present/mode100644; unsigned UTF8 path-byte order:

| Path | Projection | Blob OID |
|---|---|---|
| .ai/PROJECT_CONTEXT.md | full | 1fd444955f182995e99718f335cc10e3e190115f |
| docs/en/adr/0005-production-runtime-observation.md | full | e064c49b862ce8390981b932d4e5fa8bf9a3729f |
| docs/en/adr/README.md | full | 0d32620447f443de9d6d4575741c48bf859929f4 |
| docs/ru/adr/0005-production-runtime-observation.md | full | bcab3ff58cb793cc9c712dde9b6738983c44bbfe |
| docs/ru/adr/README.md | full | 7954a9a9985a0a2f82843154fcb74c207a85b944 |
| docs/tasks/README.md | full | c1ca17e7d3576ac94cf2ea86defa707f79565e66 |
| docs/tasks/TASK-132-PRODUCTION-RUNTIME-OBSERVATION-DESIGN.md | task-record-v1 | 98470f5bfba30df471a0107e39b507f31508f78c |
| spec/current-state.md | full | dd7ee7ac827c0b5f18f75816e8631b40b089231c |
| spec/decisions.md | full | 8436d9bcd61bc00fd11291659a4cd4d676b94a5a |

Reproduce: for full rows hash raw bytes via git hash-object --no-filters;
for task-record-v1 use exact PROCESS-001 projection excluding status body and
terminal envelope, with STATUS-EVIDENCE-EXCLUDED plus NUL. Concatenate each
path NUL projection NUL present NUL 100644 NUL oid NUL in unsigned UTF8 order;
git hash-object --stdin gives the manifest. No stage or normalization.

V132-001 documentation verification: audit132.py exit0; nine exact paths,
288 relative references valid, 4337 unchanged raw lines byte-exact, EOL churn0,
whitespace/conflict markers0, staged paths0. HEAD and protected TASK-102 ref
match intake. No code/tests/API/assets/generated artifacts changed. EN/RU
manual semantic comparison covers all eleven option dimensions, state meanings,
identity/capture/bounds/errors/freshness, ten failure cases and R1–R5. Go/race/
browser NOT RUN / N/A for design-only; no implementation PASS claim.
PROCESS-002 Synchronized within bounded design slice; Scope Audit9 Required/
0Questionable/0Removable, Size Guard PASS (zero production LOC/packages).

A132-001 Architect proposal: B+C selected, full production implementation
admission BLOCKED. A132-002 Architecture fidelity CONFIRMED for authored EN/RU;
material discrepancies0. This is not independent final review or ADR Approval.
Draft/Not Implemented, no lifecycle/protected-stage changes.
Independent final Review is the first incomplete design checkpoint.

### T132-REV-001 / R132-01 / A132-003 (historical M132-001)

Independent Reviewer /root/admission_review: Needs Revision on M132-001,
one Major blocking R132-01, no other findings. Existing process-local Attempt
anti-reuse did not prove observation-source identity across restart. Own raw
projection and canonical SHA1 reconstruction matched M132-001; audit exit0.
All nine files Required; all other design/parity/scope boundaries verified.
Architect A132-003 adjudicated R132-01 Valid and supplied mandatory qualified
source incarnation, expected pin, immutable binding, mismatch/Unknown handling,
bootstrap invalidation and restart/reuse acceptance cases. Documentation Agent
applied mirrored rework; old review/verification subject is superseded below.
This adds a production provenance qualification blocker, not a second execution
generation allocator or a newly implemented source.

### M132-002 / V132-002

Same repository/branch/base/HEAD as M132-001; canonical raw manifest: `ac7ffbe45fe51e38e20ad88870df85e83dc5bb3a`. All rows present/mode100644.

| Path | Projection | Blob OID |
|---|---|---|
| .ai/PROJECT_CONTEXT.md | full | 1fd444955f182995e99718f335cc10e3e190115f |
| docs/en/adr/0005-production-runtime-observation.md | full | cbf7347772851a2a5c79815d6dc078413c4a1148 |
| docs/en/adr/README.md | full | 0d32620447f443de9d6d4575741c48bf859929f4 |
| docs/ru/adr/0005-production-runtime-observation.md | full | 4907583f08b5e7b3dc46f3ed0b45d3f9ddf37722 |
| docs/ru/adr/README.md | full | 7954a9a9985a0a2f82843154fcb74c207a85b944 |
| docs/tasks/README.md | full | c1ca17e7d3576ac94cf2ea86defa707f79565e66 |
| docs/tasks/TASK-132-PRODUCTION-RUNTIME-OBSERVATION-DESIGN.md | task-record-v1 | 05c04766ad1b5c4fa4c1cf6c516572179b9e7cf5 |
| spec/current-state.md | full | dd7ee7ac827c0b5f18f75816e8631b40b089231c |
| spec/decisions.md | full | 8436d9bcd61bc00fd11291659a4cd4d676b94a5a |

V132-002: audit132.py exit0 after mirrored R132-01 rework; exact9 paths,
288 valid relative references, 4337 raw unchanged lines equal, EOL churn0,
staged0, no production or protected-ref change. Same canonical algorithm.
Mirror structure/protocol comparison remains applicable; manually checked new
mandatory source incarnation, request pin, immutable binding, bootstrap failure,
closed mismatch reason, source nonreuse qualification and added R3/R4 cases.
PROCESS-002 and Scope Audit remain9 Required/0Questionable/0Removable; no new
paths or statuses. Production admission BLOCKED; ADR Draft/Not Implemented.
First incomplete checkpoint: repeat independent final Review on M132-002.

### T132-REV-002 / CA132-001 / Pre-Commit STOP

Independent /root/admission_review T132-REV-002: Approved exact M132-002
ac7ffbe45fe51e38e20ad88870df85e83dc5bb3a; open findings0/blocking0.
R132-01 resolved: mandatory qualified source incarnation, bootstrap/pin,
immutable binding, capture recheck, mismatch/Unknown handling, old-response
invalidation, acceptance cases and explicit production provenance blocker.
Reviewer independently reconstructed raw task-record-v1 projection and Git blob
SHA1 manifest (exit0), in addition to audit132.py exit0:9 paths/288 links/
4337 unchanged raw lines/EOL0/staged0. EN/RU parity confirmed; Scope Audit
9 Required/0Questionable/0Removable confirmed. Existing source/failure/TOCTOU
review remains applicable; no new code or qualification claim.

CA132-001 Coordinator accepts the bounded TASK-132 design-only documentation
and forensic deliverable for exact M132-002 after independent Approved review.
This accepts a design report with full production admission BLOCKED; it does
not approve ADR-0005 (still Draft / Not Implemented), provide product
Acceptance/BCC, authorize implementation, or resolve the listed production
prerequisites. No successor task/ID, TASK-131 or protected-stage activation.

Final Owner options: separately decide sampled lifecycle plus unknown execution
versus stronger supervision/progress semantics; approve/revise focused ADR;
choose a real managed production graph, target/security boundary and qualified
source-incarnation issuer/bootstrap with nonreuse proof. R1–R5 remain conditional
recommendations, Not Activated. TASK-102 unchanged; Gate11 OPEN; StageB/B1/B2,
ProductionC2/C3 and C2-Exposure Not Activated.

Exact documentation diff manifest is the nine M132-002 rows above. No generated
files; no production/test/API/UI/assets/branding/root README changes. Go/race/
browser NOT RUN / N/A for design-only documentation. Pre-Commit Report is the
body section above with this final review/disposition. First incomplete gate:
Commit Gate NOT ENTERED, no stage/commit/publication permission; STOP. No push,
PR or merge performed. Post-disposition subject integrity check follows.

Post-disposition integrity V132-003: audit132.py exit0 after CA132-001; independently recomputed M132-002 remains ac7ffbe45fe51e38e20ad88870df85e83dc5bb3a, all9 rows unchanged, links288/raw-equal4337/EOL0/staged0. This append is evidence-envelope metadata only; no claim to hash its own final bytes. STOP at Pre-Commit; full production admission BLOCKED.

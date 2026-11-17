# TASK-135 — Configuration Studio: Draft in Existing Configuration

## Status

In Progress. No commit or publication authorized. Exact checkpoints resolve
from the matching Recovery Evidence Envelope below.

## Task Contract

- Task Mode: Implementation, explicitly selected by Owner.
- Why Now: allow reuse of existing Workspace/Configuration without duplicate
  resource creation; Runtime work is deferred by current Owner instruction.
- Branch: feature/task-135-studio-existing-configuration, created from clean
  local main f408dc2735589e53c009662c4fc2110f6c083bd7.
- TASK-134 remains on its separate local branch at
  c12a3f531f497475062a1672c0abb66872aab9fc; it is not a dependency or part of
  this diff. Main is not modified. TASK-131 DEFERRED, no liveness work.
- Scope: existing-resource selectors and explicit creation modes in Studio;
  same-origin existing APIs only, no backend/domain/API/security changes.
- Existing new Workspace -> new Configuration -> Draft -> listener update ->
  publish flow remains available. Existing Workspace + new Configuration uses
  existing version/listener/publish flow. Existing Configuration creates only
  a new Draft: no implicit publish or listener overwrite in this mode.
- Out of Scope: editing existing Draft, cloning Published, timeouts, archive,
  Runtime Status/Liveness, Dashboard mutations, dependencies/frameworks,
  protected-stage activation, commit/push/PR/merge/tag/release.
- Roles: root Coordinator/Documentation/Tester; independent Architect confirms
  existing boundaries; delegated Developer owns UI/tests; independent Reviewer
  must not author implementation.
- Order: intake/coverage -> Architecture Confirmation -> implementation ->
  focused/regression verification -> documentation sync/scope -> independent
  review -> Pre-Commit Report -> STOP. No automatic successor.
- Stop conditions: new domain/API contract, changed scope/auth/lifecycle,
  unexpected paths or code blocker. Size Guard reassess >15 paths or >500
  production lines or independently deliverable behavior.

## Acceptance Criteria

1. Preserve the complete existing create-and-publish scenario.
2. Select an existing Workspace and only a Configuration within that Workspace;
   create Draft via existing POST versions without parent creation.
3. Create a new Configuration within a selected existing Workspace, without
   creating that Workspace again.
4. Workspace changes clear Configuration selection and dependent data;
   Configuration/mode changes clear results and stale operation data.
5. Lists have loading/empty/error/retry states; old asynchronous responses
   cannot replace the current selection. Invalid/unsafe IDs and mismatched
   response scopes are rejected before downstream requests.
6. One mutation flow at a time; repeated submit cannot duplicate resources.
   No automatic retries of mutations. Unknown mutation outcomes require state
   inspection before a new attempt, not optimistic success or automatic replay.
7. Deleted/unavailable resources and Draft errors never produce success or
   fallback parent creation. Explicit scoped recovery permits a fresh action.
8. Preserve trusted-local/same-origin boundary, CSP, safe text rendering and
   Published != Running/in-memory notices. No secrets or new Runtime claims.

## Existing Coverage Report / Verification Plan

Existing Coverage: configurationstudio/handler_test.go covers embedded assets,
exact routes/security headers/branding/notices and static flow guards;
dashboard_test.js exercises asynchronous selection and scope handling in Node.
Workspace/Configuration/Version handler and service tests cover existing HTTP
creation/list/update, Draft restrictions, listener settings and publish rules.
TASK-076 proves the original guided UI; TASK-127 proves GET-only Dashboard.
Coverage Gap: no executable Studio test for selected existing resources,
selection races, malformed IDs/scope, duplicate submit or partial failure.
Added Proof plan: Node VM user-flow tests for all three creation modes,
out-of-order list responses, invalid scope/IDs, empty/error lists, missing
resources, repeated submit, Draft failure/unknown outcome and recovery; focused
Go embedded/API tests and full Go regression/build/vet; existing Dashboard Node
tests. Browser/manual check when an available tool permits; otherwise record
the precise limitation. No tests claimed PASS unless executed. No need to repeat
TASK-134 race; this diff does not change Go concurrency. Race qualification of
this new diff remains distinct if unavailable locally.

## Expected Changed Files / Documentation Applicability

- internal/configurationstudio/assets/index.html — mode/select controls.
- internal/configurationstudio/assets/studio.js — selected-resource flows.
- internal/configurationstudio/assets/studio.css — native select/hidden styles
  only if needed.
- internal/configurationstudio/handler_test.go — adjust existing UI guards.
- internal/configurationstudio/studio_test.js — executable interaction proofs.
- This task record, docs/tasks/README.md, .ai/PROJECT_CONTEXT.md,
  spec/current-state.md — mandatory task/capability synchronization.
- Mirrored docs/en/roadmap/MASTER_PLAN.md and docs/ru/roadmap/MASTER_PLAN.md —
  bounded factual Studio capability note, no milestone/dependency changes.
- User documentation/CHANGELOG applicability to be resolved after inventory;
  additions require scope justification before modification.
- spec/decisions.md, ADR/DP/ARCH: N/A unless a factual mismatch is discovered;
  no architectural decisions or implementation statuses promoted.
- Historical task records including TASK-134 remain unchanged.

## Admission and Documentation Baseline

A135-001 independent Architect: APPROVED — IMPLEMENTATION READY, blockers0.
Existing Configuration creates fresh default Draft only; other modes retain
listener/publish sequence. Existing API and same-origin embedded boundary
suffice; no new ADR/DP/domain contract. Bounded list GET timeout compatible;
no polling or mutation replay. Root Documentation/Tester roles assigned.

Selected directly by Owner; metadata editor, timeout editor and archive are
rejected alternatives. No automatic next candidate activation. TASK-076 and
TASK-127 are published historical records, not active tasks despite historical
Status projections. TASK-134 remains separate; current Owner explicitly selects
this task and defers TASK-131. Old TASK-132 current navigation is stale after
its merge into this baseline; repair only current navigation.

Additional scope justification BEFORE changes: docs/en/quick-start.md and its
RU mirror describe new modes and retain historical screenshot qualification;
CHANGELOG.md gets a bounded user-facing Unreleased entry. Expected total14
paths (5 UI/tests,9 documentation), no new Go package or backend source.
README/screenshots N/A: still illustrate preserved default flow; no new
screenshot claimed. spec/decisions and ADR/DP/ARCH N/A: no architecture changes.
MASTER_PLAN mirrors update factual capability only, no milestone/dependency
change. EN/RU user guidance must remain aligned.

## Implementation and Added Proof Inventory

Implemented three explicit Studio modes using only existing APIs. Existing
Configuration mode performs one POST versions and displays Draft, with listener
inputs and preliminary URL hidden. New-parent modes retain listener/publish
sequence, omitting Workspace creation when it already exists. Selected IDs must
be safe positive integers and members of the selected scope. Mode/Workspace/
Configuration changes invalidate dependent data; generation guards reject old
list responses. Reload clears both selections and refreshes Workspace list.
One flow at a time; controls stay locked after an attempted mutation until
explicit new-action acknowledgement. No automatic mutation retry. Client request
budgets never imply server operation cancellation or a known mutation outcome.
No Dashboard/backend/domain/security/Runtime/lifecycle code changed.

Added Proof: 23 executable Studio Node tests cover all3 modes and exact mutation
sequences, stale list responses, scope/unsafe IDs, empty/error lists and bounded
GET timeout, deleted-parent reload, double submit, Draft errors/malformed scope,
unknown outcome acknowledgement and text-only rendering. Existing15 Dashboard
Node tests remain unchanged. Go asset guards retain allowed-API/security/notices
checks while allowing bounded request timers (no polling). CSS keeps hidden
controls hidden and native selects readable in the shared dark theme.

## Verification Matrix / Remaining Limitations

Frozen Developer five-file raw OIDs match root verification before/after SHA256
and actual browser-tested source. Successful checks were not repeated after
code freeze; documentation changes do not alter executable inputs.

| Check | Actual result |
|---|---|
| node --test internal/configurationstudio/studio_test.js internal/configurationstudio/dashboard_test.js | Developer final run: 38 PASS (23 Studio +15 Dashboard),0fail |
| go test -mod=readonly -count=1 -timeout=3m ./internal/configurationstudio ./internal/workspace ./internal/configuration ./internal/configurationversion | Root exit0 on frozen files |
| go test -mod=readonly -count=1 -timeout=3m ./... | Initial sandbox exit1: loopback access denied; native authorized repeat exit0 |
| go build -mod=readonly ./... | Root exit0 |
| go vet -mod=readonly ./... | Root exit0 |
| Real Chrome headless + actual local Control Service | PASS all3 modes, exact mutation counts5/1/4, no duplicate parents, Draft not published, no JS exceptions |
| Browser responsive/layout | Desktop1280 and mobile390, mobile scrollWidth390, hidden listener/URL rows and dark select colors verified; screenshots visually inspected |
| gofmt / diff / module guards | Final checks recorded below; module files/backend/workflows remain unchanged |
| Go Race Detector | NOT RUN for TASK135; Windows toolchain has no configured C compiler. Only HTML/CSS/JS production changes; no Go concurrency code changed. No TASK134 CI result transferred to this diff. |

Browser smoke used isolated temporary profile/data and test-only in-memory
resources, without installation or system configuration changes. Background
processes created for smoke were stopped. External artifacts: task135-browser.json,
task135-desktop-draft.png, task135-mobile-select.png, task135-verification.json,
task135-{focused,regression,regression-native,build,vet}.log in this chat's
visualizations directory. Initial sandbox failures are retained, not hidden.

Residual limits: trusted-local unauthenticated service, in-memory resource loss
on restart, no transaction across the multi-request creation flow, and no backend
idempotency. Unknown outcomes require inspection; earlier resources can remain.
No existing Draft editor, cloning, timeout settings, archive or Runtime feature.
The browser smoke covers successful live flows; failure/interleaving proofs use
controlled Node fixtures and existing backend tests. Race NOT RUN is explicit;
no claim of a new full Go race qualification is made.

## Scope Audit / Pre-Commit Report

Exactly14 paths, all Required, Questionable0/Removable0: three UI assets provide
controls/flow/readability; two test paths prove interactions and embedded scope;
task record/index/context/current-state provide durable handoff; mirrored Quick
Start explains new actions; mirrored MASTER_PLAN qualifies factual capability;
CHANGELOG records the user-facing enhancement. No new package/dependency,
architecture decision or milestone boundary. Root documentation synchronization
preserves baseline raw bytes on unchanged lines; no incidental EOL normalization.
Final canonical manifest and independent review resolve from envelope below.

Commit Gate recommendation: eligible for separate Owner consideration after
matching final independent review; no commit permission requested or inferred.
Record race limitation separately from successful UI/API/regression verification.
Main, TASK134 local commit and protected TASK102 ref remain unchanged. TASK131
DEFERRED; ADR0005 Draft; Gate11 OPEN; protected stages Not Activated. STOP at
Pre-Commit, no automatic successor or publication.

## Recovery Evidence Envelope

Intake baseline and clean branch verified. Task record is first content change.
Architecture Confirmation and implementation not started. Gate11 OPEN;
TASK-102 untouched; Stage B/B1/B2, Production C2/C3 and C2-Exposure Not Activated.

### M135-001 / exact reviewed subject and Tester handoff

Branch feature/task-135-studio-existing-configuration; baseline/current HEAD
f408dc2735589e53c009662c4fc2110f6c083bd7; object format sha1.
Canonical manifest M135-001: 4765c1ed9bcca69463ea4c4c3a8c52bb06addc7d.
All rows present/mode100644; unsigned UTF8 path order, raw no-filter hashing,
task-record-v1 Status-body/terminal-envelope exclusion with
STATUS-EVIDENCE-EXCLUDED+NUL; NUL-delimited path/projection/state/mode/OID rows
including trailing NUL hashed as Git blob. Envelope is not self-attested.

| Path | Projection | Blob OID |
|---|---|---|
| .ai/PROJECT_CONTEXT.md | full | 450d598c09fe985b830a6773711905deb94865a0 |
| CHANGELOG.md | full | fc38bdec2a21d6694f9f131a48f1e0be9bc5c45b |
| docs/en/quick-start.md | full | 09dba912ad646eee02081937e2a8fc9069afa4fb |
| docs/en/roadmap/MASTER_PLAN.md | full | 713043c6a0e66e88970dc9183d929b0bfcabd766 |
| docs/ru/quick-start.md | full | 8d1862f32c25d9a2cbb8afd538fe4135e5635348 |
| docs/ru/roadmap/MASTER_PLAN.md | full | 030da699a8a5bf66dc609a25b63d562122aa5e5a |
| docs/tasks/README.md | full | e5b1fa342f7c69a1bc631189bab0d7f029af90b2 |
| docs/tasks/TASK-135-STUDIO-EXISTING-CONFIGURATION.md | task-record-v1 | f47a704894f1804c6160eb6e0d2c757c19618700 |
| internal/configurationstudio/assets/index.html | full | f8a43cb953453489e3e07c9afcdd803ed7beba47 |
| internal/configurationstudio/assets/studio.css | full | ca2a87b6c060177f89a64aecd2b79b5e24b82e06 |
| internal/configurationstudio/assets/studio.js | full | a7c2171122a2519a3c02c2da3faf9a231af1f76f |
| internal/configurationstudio/handler_test.go | full | 9f13864b324dfe629f107a2c9f0a236d4c457fa1 |
| internal/configurationstudio/studio_test.js | full | 25c3355695efb99e5aa43855323dba3cfa07c778 |
| spec/current-state.md | full | 025d20e8241436b71489c18f0157b0fe227ac777 |

Root audit135.py exit0: exact14paths,329 relative links,6082 unchanged raw lines
byte-exact,EOLchurn0,staged0. ScopeAudit14Required/0Questionable/0Removable.
SizeGuardPASS: one selected-resource creation behavior,252 added production
lines across3 assets,0newGoPackages. PROCESS002 Synchronized; mirrored user
guidance and roadmap factual notes consistent. Developer frozen5 OIDs above
match root before/after SHA256 and browser inputs.

Root focused4packages/build/vet exit0; native regression40packages exit0 after
recorded sandbox loopback failure. Developer final Node38PASS; root actual Chrome
3modes/5-1-4mutations/no duplicates/hiddenDraftURL/desktop1280/mobile390/noJSerrors
PASS. gofmt empty; diff--check exit0; go.mod/go.sum/workflows/backend unchanged.
Race NOT RUN for this UI-only diff; no Go production/concurrency changes.
Final independent Review pending M135-001, no commit/publication authorization.

### T135-REV-001 / CA135-001 — final review and Pre-Commit disposition

Independent Reviewer /root/reviewer135 APPROVED exact M135-001
4765c1ed9bcca69463ea4c4c3a8c52bb06addc7d; open findings0. Preliminary
R135-01 deleted-parent Reload fixed with regression; dark-select issue fixed
and visually checked. Own raw Gitblob/task-record-v1/NUL reconstruction matched
all14 rows; exact14/staged0/main+TASK134+TASK102 refs unchanged;329links,
6082 raw-equal unchanged lines,EOLchurn0. Five tested before/after SHA256 match.
Reviewer inspected existing focused/build/vet/native40package regression logs,
browser3mode5/1/4mutation proof, actual resources and desktop/mobile screenshots;
reran no tests. Documentation mirrors and scope consistent;14Required,
0Questionable/0Removable. No backend/domain/lifecycle/API/security changes.

Coordinator CA135-001 accepts the bounded UI/API slice with the explicit
verification limitation: Go Race Detector NOT RUN for TASK135 (no configured
C compiler; no Go production or concurrency changes). No complete new Go race
qualification and no transfer of TASK134 CI evidence. All executed final UI/API,
focused/regression/build/vet/browser/format/scope checks PASS; initial sandbox
loopback failures retained with successful native substitute separately.
Commit Gate recommendation: ready for separate exact Owner review/authorization
with this disclosed limitation. No stage/commit/publication performed or granted.
Current HEAD/main f408dc2735589e53c009662c4fc2110f6c083bd7; TASK134 separate
local commit c12a3f531f497475062a1672c0abb66872aab9fc remains preserved.
TASK131 DEFERRED; ADR0005 Draft; TASK102 untouched; Gate11 OPEN;
protected stages Not Activated. STOP at Pre-Commit; no successor activated.

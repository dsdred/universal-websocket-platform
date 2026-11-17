# TASK-136 — Existing Draft Management in Studio

## Status

In Progress. No commit/publication permission. Final checkpoints resolve from
the newest matching Recovery Evidence Envelope.

## Task Contract

- Task Mode: Implementation, subject to independent Architecture Confirmation.
- Why Now: Owner selected discover/open/edit/save an existing Configuration
  Draft, preserving all three published TASK-135 creation flows.
- Branch: feature/task-136-existing-draft-management.
- Trusted baseline: clean main 3683b7734a692a97ea8b423cc6bba9376a9976ef;
  TASK-135 PR134 terminal publication confirmed in the original native record.
- Protected local refs: TASK134 c12a3f531f497475062a1672c0abb66872aab9fc
  outside main; TASK102 8bfa2e22a4b6310a71f6d36c8724b1adedd727ad.
- Smallest workflow: select existing Workspace/Configuration, list Drafts,
  open selected Draft without POST, edit Listener host/port, explicitly save
  via existing scoped PUT listener. Other configuration sections stay unchanged.
  This completes the existing listener editing surface; it is not a whole-DSL
  editor, clone, new backend contract or automatic publication.
- Out of Scope: backend/domain/API changes, TLS/authentication/routing/timeout
  editors, publish action in existing-Draft mode, archive, Runtime/health,
  lifecycle changes, TASK131 activation, protected stages, tools/dependencies,
  commit/push/PR/merge/tag/release. TASK131 DEFERRED; Gate11 OPEN.
- Roles: root Coordinator/Documentation/Tester; independent Architect confirms
  admission; Developer owns UI/tests; independent non-author Reviewer.
- Order: discovery/coverage -> bounded product contract/admission -> UI/tests
  -> focused/browser/regression/build/vet -> documentation/scope -> independent
  review -> Pre-Commit Report -> STOP. No automatic successor.
- Stop if existing APIs cannot safely support this workflow, new semantics or
  security boundary needed, unexpected paths, or unresolved blocking finding.

## Capability Inventory (Phase 0)

| Operation | Classification | Exact repository evidence |
|---|---|---|
| Select Workspace/Configuration | EXISTS AND WORKS | configurationstudio/assets/studio.js:loadLists/updateMode; workspace/handler.go:RegisterRoutes:38; configuration/handler.go:RegisterRoutes:29 |
| List complete scoped versions | EXISTS BUT NOT EXPOSED IN UI for editor | configurationversion/handler.go:RegisterRoutes:91; service.go:List:353; memory_repository.go:ListByConfiguration:48 |
| Open one existing Draft via scoped list | EXISTS BUT NOT EXPOSED IN UI | List returns detached version data; memory_repository.go:ListByConfiguration:48 |
| GET version-by-ID / whole-version PATCH | MISSING, unnecessary here | handler.go:RegisterRoutes:89-98 lists actual routes; repository.Get is not an HTTP endpoint |
| Save existing Draft host/port | EXISTS BUT NOT EXPOSED IN UI | service.go:UpdateListener:241; handler.go:updateListener:282; PUT route:94 |
| Published immutability and field validation | EXISTS AND WORKS | service.go:UpdateListener:249-271 checks exact configuration, Draft state, validateListener; repository.Update clones |
| Save preserves other settings | EXISTS AND WORKS | service.go:UpdateListener:269-272 assigns only Host,Port,UpdatedAt |
| TLS/timeouts/authentication/routing section updates | EXISTS, outside this UI slice | handler.go:RegisterRoutes:95-98; service.go:UpdateTLS/UpdateTimeouts/UpdateAuthentication/UpdateRouting |
| Explicit publish | EXISTS AND WORKS, not used by editor | service.go:Publish:276-319; separate POST route:92 |
| Atomic conditional save / revision / ETag conflict prevention | MISSING, not promised | UpdateListener has no expected revision; repository.Update replaces by ID |

## Acceptance Criteria / Product Contract

1. Preserve three TASK135 creation modes; editor is a fourth explicit action.
2. Discover only exact-Configuration Drafts. Opening/switching/reloading editor
   uses GET versions, never POST/create or publish. No hidden version cloning.
3. Edit existing listener host/port only; successful Save is one scoped PUT to
   the pinned existing version. Verify returned ID/configuration/state/settings.
4. Dirty state visible. Mode/Workspace/Configuration/Draft changes, reload,
   reset and page unload must not silently discard edits. Explicit discard
   confirmation; Cancel restores previous choices and input values.
5. Generation/abort guards prevent late lists/open responses from rebinding
   selected Draft. Parent changes clear dependent selections after discard.
6. Duplicate Save prevented; edit/selection controls locked while Save pending.
   Save errors retain input; no automatic replay. Unknown outcomes require an
   explicit reconciliation action, preserving edits until acknowledged.
7. If externally published/archived/deleted after open, existing backend denies
   save; UI shows error, retains input and never falls back to new Draft.
8. Validate safe IDs and exact scopes; use text-only labels, no raw JSON/Secret
   references exposed. Existing same-origin/trusted-local/CSP unchanged.
9. Desktop/mobile actual browser flows and meaningful error/race tests pass.
   No cross-client lost-update prevention claim without conditional-write API.

## Existing Coverage Report / Verification Plan

Existing: TASK135 Studio23 Node and Dashboard15 Node tests prove creation modes,
scope/safe IDs, list races, duplicate actions and unknown outcomes. Embedded Go
tests prove route/assets/CSP/notices/API scope. ConfigurationVersion handler and
service tests cover list/create/listener validation, state/scope restrictions,
publish/archive; repository tests cover detached copies and ordered lists.
Gap: existing-Draft discovery/open/save, dirty navigation/unload, stale opens,
save conflicts/errors and editor interaction regression are not covered.
Add focused Node proofs for gap, plus actual Chrome+Control Service round trips
and desktop/mobile screenshots; preserve all TASK135 tests. Run focused Go,
full native Go regression, build/vet/gofmt/diff/link checks. Race is required if
production concurrency code changes; otherwise record NOT RUN/N/A precisely.

## Expected Changed Files / Documentation Applicability

- Studio assets/index.html, assets/studio.js, assets/studio.css as needed.
- Existing configurationstudio/studio_test.js and handler_test.go as needed;
  a separate private editor script/test may be justified before modification.
- This task record, docs/tasks/README.md, .ai/PROJECT_CONTEXT.md,
  spec/current-state.md: mandatory recovery/task/capability synchronization.
- Mirrored Quick Start and MASTER_PLAN: user instructions and factual capability
  only, no milestone/architecture promotion. CHANGELOG: user-facing enhancement.
- README/screenshots N/A if original illustrated flow remains valid; no fake
  new screenshot claim. spec/decisions/ADR/DP/ARCH N/A without actual decision.
- Historical TASK134/135 records untouched. Backend and workflows unchanged.
- Size Guard reassess >15 paths/>500 production lines/new package/multiple
  contracts. Task is one bounded discover/open/save existing-listener workflow.

## Architecture Confirmation / Admission

A136-001 independent Architect: A — IMPLEMENTATION ADMISSIBLE, blockers0.
ADR0002 permits independent sections; existing list GET + listener PUT support
one complete bounded existing-Draft host/port workflow. Other section editors
are excluded, not missing prerequisites or a partial speculative delivery.
No new ADR/DP/API/domain/security/lifecycle contract required.

Opening uses fresh scoped GET. Store only safe field projections, because full
version responses include Secret References. Error/unknown save retains inputs
and pinned identity; unknown results require explicit same-version GET across
all states before a new intentional Save. Matching current server fields proves
observed values, not attribution to the prior request. Missing/non-Draft stays
unsavable. Dirty/pending/unknown guards include unload and cancel restores prior
choices. No other-section equality check: backend preserves their current values.
No CAS guarantee invented; concurrent listener writers can overwrite each other.

Documentation applicability confirmed before modifications: exact14 expected
paths =5UI/tests +9docs (record/index/context/state,QuickStartEN/RU,
MASTER_PLANEN/RU,CHANGELOG). Historical TASK135 record unchanged; source commit
and terminal publication facts belong in current navigation. README/screenshots
still illustrate preserved creation flow; no new exported screenshot claim.
No milestone/engineering dependency or architectural decision changes.

## Implementation Summary / Added Proofs

Admission A136-001 implemented without backend or contract expansion. Fourth
Studio action Edit existing Draft discovers versions using existing scoped GET,
projects only allowlisted metadata, filters Draft choices and freshly reads
before opening. Open/switch never creates a version. Save listener pins existing
Workspace/Configuration/version and sends one PUT host/port; success validates
identity, Draft state and returned fields. Other sections remain backend-owned.

Dirty/pending/unknown state guards changes of mode/parents/Draft, reload/reset
and beforeunload. Cancel restores prior selections and input. Controls serialize
Save; generation+Abort guards reject old reads. Failed Save keeps typed values.
Unknown transport/5xx/malformed outcomes block replay until explicit confirmed
Reload Draft and fresh GET of the same version across all states. Successful
reload replaces local fields only after reading; it reports observed/opened
state, never attribution of an uncertain save. Failed/missing/non-Draft reload
keeps old fields and unknown lock. Concurrent clients have no CAS protection.
All three TASK135 creation flows remain intact. Editor has no publish/create
fallback, no TLS/auth/routing/timeout editor and no Runtime work.

Added proofs: 36 Studio cases extend the unchanged23 creation proofs, including
all parent/Draft/mode dirty-cancel guards, reset/unload, fresh/open races, pinned
PUT, repeated actions, validation/HTTP errors, unknown-response reconciliation,
missing/non-Draft, malformed IDs/scope/state and projection/redaction. Existing15
Dashboard tests remain unchanged. Go asset surface assertions and focus styling
are bounded support for this same workflow. Preliminary R136-pre-01 mojibake
was corrected before final freeze; final source UTF8 punctuation and LF checked.

## Final Verification Matrix / Limitations

Exact five frozen UI/test OIDs resolve in the final envelope. Root's completed
Go runner before/after SHA256 and current files match; final Node and real-browser
checks use these same inputs. No production edit after corrected freeze.

| Check | Actual result |
|---|---|
| node --test internal/configurationstudio/studio_test.js internal/configurationstudio/dashboard_test.js | Root final exit0:74PASS,59Studio (23old+36new)+15Dashboard,0fail/skip/cancel |
| go test -mod=readonly -count=1 -timeout=3m ./internal/configurationstudio ./internal/workspace ./internal/configuration ./internal/configurationversion | exit0,5.578s |
| go test -mod=readonly -count=1 -timeout=3m ./... | Native exit0,33.177s,40 tested packages |
| go build -mod=readonly ./... | exit0,2.750s |
| go vet -mod=readonly ./... | exit0,2.069s |
| gofmt -l internal/configurationstudio/handler_test.go | exit0,empty |
| Chrome + actual isolated local Control Service | PASS three creation modes and13 editor assertions; actual native confirm dialogs exercised |
| Browser save/reopen | Same version ID, onePUT despite repeat action, saved host/port reread; other TLS/timeouts/auth/routing settings preserved |
| Browser unknown outcome | Actual PUT response deliberately dropped at response stage; local edits retained, Save blocked, explicit confirmed freshGET reconciliation succeeds |
| Browser lifecycle race | External fixture publishes opened Draft; UI PUT rejected, typed fields retained, stored Published fields unchanged |
| Browser data/appearance | No editor-created version, Published not offered, references not rendered;1280desktop/390mobile screenshots inspected,scrollWidth390,noJSexceptions |
| Race Detector | NOT RUN for TASK136; production changes only HTML/CSS/JS, no Go concurrency source change, so conditional race requirement does not apply. No prior CI/race result transferred. |

Native Go commands completed before the Owner resume; recovery matched their
before/after/current inputs and reused successful results. Browser proof was
actually executed after resume. External artifacts in this chat's visualizations
folder: task136-verification.json,task136-node.log,task136-focused.log,
task136-regression.log,task136-build.log,task136-vet.log,task136-browser.json,
browser136.cjs and task136-desktop-saved/mobile-dirty.png. Isolated test service
and headless browser were stopped; no installs or persistent environment changes.

Remaining limits: only host/port editing; no revision/ETag/conditional update,
so concurrent clients can overwrite each other. In-memory data, trusted-local
unauthenticated deployment and existing backend semantics remain. Unknown Save
is reconciled by explicit confirmed discard/load, not automatically retried or
attributed to the original request. beforeunload handler was exercised by event;
no guarantee of native warning on browser crash/forced termination/mobile close.
No new health, activation, persistence or cross-client locking claim.

## Documentation / Scope Audit / Pre-Commit Report

Exact14 paths all Required,Questionable0/Removable0:3Studioassets implement
controls/flow/focus; existing Node/Go test files prove new behavior and surface;
record/index/context/current-state anchor recovery and publication facts;
mirrored QuickStart documents editor and limits; mirrored MASTER_PLAN adds only
factual capability; CHANGELOG records user-facing behavior. No new package,
dependency, domain/API/architecture/milestone change. SizeGuard: one bounded
existing-Draft editing workflow,14paths,well below500production additions.
README/screenshots/ADR/DP/ARCH/spec-decisions N/A as justified above. TASK134/135
historical records untouched; no unrelated backend/workflow/module changes.
PROCESS002 synchronized,EN/RU parity checked. Final manifest and review verdict
resolve only from the matching terminal envelope; pending review is not PASS.

Pre-Commit recommendation depends on final independent review of exact subject.
No stage/commit/publication authorized; STOP at Pre-Commit. TASK131 DEFERRED;
ADR0005 Draft; TASK102 untouched; TASK134 separate local commit retained;
Gate11 OPEN; Stage B/B1/B2,Production C2/C3,C2-Exposure Not Activated.

## Recovery Evidence Envelope

Intake: baseline/clean main/protected refs verified; safe task branch created;
this record is first content change. Architecture Confirmation pending;
production implementation not started. Gate11 OPEN; protected stages inactive.

### REC136-001 — Owner final-verification resume

Current Owner resumes TASK136 final verification only, no new implementation.
HEAD/main baseline3683b7734a692a97ea8b423cc6bba9376a9976ef, taskbranch and
exact14 attributed paths reconstructed; staging0, TASK134/TASK102 unchanged.
A136-001 admission completed, implementation frozen in5UI/test paths. Reviewer
preliminary R136-pre-01 UTF8 punctuation issue fixed; final author Node74PASS
(59Studio including23old+36new,15Dashboard). Root matched all5 current SHA256
with completed verification runner before/after identities. Focused4 Go packages,
full regression, build and vet actually completed exit0; no need to repeat.
Real browser script prepared but NOT RUN at this checkpoint. Final independent
review NOT completed: reviewer turn errored on usage limit after preliminary
code assessment. Formal review must resume/reconstruct after final browser/docs
and exactmanifest. No review PASS inferred from preliminary inspection.
First incomplete checkpoint: actual browser proof, then documentation/scope/
final Review/Acceptance. No commit/publication authorized.

### M136-001 / final subject and Tester handoff

Branch feature/task-136-existing-draft-management; HEAD/base/main
3683b7734a692a97ea8b423cc6bba9376a9976ef; sha1.
Canonical M136-001 25de7f83b8ab6b0c6cda255a16e124b2cdcd7da6.
Raw no-filter blob hashing; task-record-v1 retains BOF through Status heading
newline, STATUS-EVIDENCE-EXCLUDED+NUL, then Task Contract through before terminal
envelope. Unsigned UTF8 ordered path/projection/present/mode/OID rows joined
with NUL, including trailing NUL, hashed as Git blob. Envelope excluded, not
self-attested. All rows present/mode100644.

| Path | Projection | Blob OID |
|---|---|---|
| .ai/PROJECT_CONTEXT.md | full | 60f63f16508d914ee02e74dbda9025dab1c4fd4c |
| CHANGELOG.md | full | a5228b3f75c2f22c3cac38bf541da753a9a75051 |
| docs/en/quick-start.md | full | c23e5e91002536a07972b1fa46ec80f5b295e401 |
| docs/en/roadmap/MASTER_PLAN.md | full | 09e5d944a94d06c855a828c007c90fa352a72b2d |
| docs/ru/quick-start.md | full | 80d084bdf4839c4ec899465b7fdc4786375833ac |
| docs/ru/roadmap/MASTER_PLAN.md | full | 2c540a99bcf949dd4e87315bbd715a4be94090e3 |
| docs/tasks/README.md | full | 0489aae95aeaa5429138e638f6e89f09d2e0bc7a |
| docs/tasks/TASK-136-EXISTING-DRAFT-MANAGEMENT.md | task-record-v1 | 3d85c60e40691f054b5fd69164ef2164a33558cd |
| internal/configurationstudio/assets/index.html | full | 20ee3728111a2a37535b390c2cb32864d87e0201 |
| internal/configurationstudio/assets/studio.css | full | 1bcb98e51d1f4beb87dafd3630d72b0951c5c763 |
| internal/configurationstudio/assets/studio.js | full | 50ef3b4456d389ee5efba7889bc6f675783b049d |
| internal/configurationstudio/handler_test.go | full | 6b8877423f1b756789c645bbe29827d9c925169d |
| internal/configurationstudio/studio_test.js | full | d864fa19f3489be69b897cc8d05e7eb9678169b6 |
| spec/current-state.md | full | 00765151cbb95c67004cff521a9c5df3f8bd1e58 |

Root audit136.py exit0:14paths,332links,6586 unchanged raw lines byte-exact,
EOLchurn0,staged0. All14 Required,0Questionable/0Removable; SizeGuardPASS.
Corrected frozen5 OIDs match Go before/after/current SHA256 and actual browser
inputs. Actual Root Node74PASS, focused4packages/full40packages/build/vet/gofmt
PASS; Chrome3oldflows+13editorassertions/desktop1280/mobile390/noJSerrors PASS.
Browser simulated response loss after real PUT execution; exact sameversion
reload and externalpublish409 preserve truthful localstate. Production Go
concurrency unchanged; RaceNOTRUN/conditional requirement notapplicable.
No priorCI/race evidence transferred. Scope/domain/module/workflowguards clean;
main/TASK134/TASK102 refs unchanged. Final independent Review pending.

### M136-002 / exact evidence-anchor correction

Supersedes pending M136-001 for final review: two capability-inventory anchors
correct ListByConfiguration declaration line49 to48. No executable/test/user
documentation byte changes. Completed test inputs remain identical; only
documentation subject identity changed. M136-001 is historical, not current.

Canonical M136-002 38490957940a904a2399a44017a110f2879ae739; same base/branch,
raw Gitblob/task-record-v1/NUL algorithm. Exact ordered rows:

| Path | Projection | Blob OID |
|---|---|---|
| .ai/PROJECT_CONTEXT.md | full | 60f63f16508d914ee02e74dbda9025dab1c4fd4c |
| CHANGELOG.md | full | a5228b3f75c2f22c3cac38bf541da753a9a75051 |
| docs/en/quick-start.md | full | c23e5e91002536a07972b1fa46ec80f5b295e401 |
| docs/en/roadmap/MASTER_PLAN.md | full | 09e5d944a94d06c855a828c007c90fa352a72b2d |
| docs/ru/quick-start.md | full | 80d084bdf4839c4ec899465b7fdc4786375833ac |
| docs/ru/roadmap/MASTER_PLAN.md | full | 2c540a99bcf949dd4e87315bbd715a4be94090e3 |
| docs/tasks/README.md | full | 0489aae95aeaa5429138e638f6e89f09d2e0bc7a |
| docs/tasks/TASK-136-EXISTING-DRAFT-MANAGEMENT.md | task-record-v1 | 18d754e363b8722f45f3b9e2bd8558bd544a1519 |
| internal/configurationstudio/assets/index.html | full | 20ee3728111a2a37535b390c2cb32864d87e0201 |
| internal/configurationstudio/assets/studio.css | full | 1bcb98e51d1f4beb87dafd3630d72b0951c5c763 |
| internal/configurationstudio/assets/studio.js | full | 50ef3b4456d389ee5efba7889bc6f675783b049d |
| internal/configurationstudio/handler_test.go | full | 6b8877423f1b756789c645bbe29827d9c925169d |
| internal/configurationstudio/studio_test.js | full | d864fa19f3489be69b897cc8d05e7eb9678169b6 |
| spec/current-state.md | full | 00765151cbb95c67004cff521a9c5df3f8bd1e58 |

audit136.py exit0:14paths,332links,6586 unchanged raw lines,EOLchurn0,staged0;
all14Required/0Questionable/0Removable. Completed final verification remains
exact5input-qualified. Final Review now targets M136-002; still pending.

### R136-001 / CA136-001 — final independent review and acceptance

Independent Reviewer /root/reviewer136: Approved exact M136-002
38490957940a904a2399a44017a110f2879ae739; unresolved findings0. Own raw14path
Gitblob/task-record-v1/NUL reconstruction matched all rows; task projection
18d754e363b8722f45f3b9e2bd8558bd544a1519. Baseline/branch/main/staged0 and
protected refs unchanged. Preliminary punctuation finding resolved. No code or
test changes since corrected freeze. Reviewer inspected existing actual test
logs, matching executable before/after/current hashes, real browser traces,
response loss after serverwrite, screenshots and EN/RU documentation; reran no
tests. All14 paths Required,0Questionable/0Removable. No scope/architecture/
backend/API/protected-stage violation or false success/liveness assertion.

Coordinator CA136-001 accepts this bounded existing-Draft host/port editor for
exact M136-002. Executed UI/API/browser/focused/regression/build/vet/format/scope
verification PASS. Go race NOT RUN: conditional requirement not applicable,
production concurrency code unchanged; no full Go race qualification claimed.
No revision/ETag/CAS or cross-client lost-update guarantee; no other-section
editor. TASK131 DEFERRED; TASK134 separate localcommit and TASK102 unchanged;
Gate11 OPEN; ADR0005 Draft; protected stages Not Activated.

Commit Gate recommendation: READY for separate exact Owner authorization.
Current task has no stage/commit/push/PR/merge/tag/release permission. STOP at
Pre-Commit Gate; no successor task activated. This append-only metadata does
not alter the accepted subject or self-attest its own final bytes.

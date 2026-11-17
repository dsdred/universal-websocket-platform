# TASK-137 — Extended Draft Editing in Studio

## Status

In Progress; timeout slice admitted. Exact current checkpoint resolves from the
matching terminal Recovery Evidence Envelope, not this projected status.

## Task Contract

- Status: In Progress; Task Mode: Implementation.
- Repository: universal-websocket-platform; branch:
  `feature/task-137-extended-draft-editing`.
- Trusted baseline: clean main/origin/main
  `0dfd13bf49db2a3bd63d03111f46cedcb07261d7`, PR135 merge; Owner reports
  TASK136 terminal PASS. Local Git ancestry confirms merge/source; no new
  Publisher transaction or fresh remote terminal certification is claimed.
- Why Now: TASK136 provides existing-Draft host/port editing; Owner directs
  inventory and one additional independently verifiable metadata editor.
- Definition of Done: repository capability inventory with exact symbols;
  explicit architecture admission; smallest useful existing-contract slice;
  same-version save/load/reopen, validation, preservation and failure guards;
  Node/Go/browser verification, independent non-author review, PROCESS002,
  exact scope audit and Pre-Commit Report.
- Candidate: four listener timeout values through existing timeout PUT;
  admission must decide independent section Save versus multi-request Save.
- Out of Scope: new backend contracts, lifecycle/runtime/security boundary
  changes, CAS/ETag guarantees, whole-DSL editor, TLS/auth/routing editors,
  TASK134 modification, TASK131 activation, TASK102, protected stages,
  stage/commit/push/PR/merge. Gate11 OPEN; protected stages Not Activated.
- Roles assigned by Coordinator: root Coordinator/Documentation/Tester;
  independent Architect for confirmation; separate Developer for UI/tests;
  independent non-author Reviewer for final exact diff.
- Ordered stages: read-only inventory/coverage -> architecture confirmation ->
  bounded implementation -> tests/browser -> documentation/scope -> independent
  review/rework -> Coordinator decision -> Pre-Commit STOP.
- Stop: missing safe existing API/validation, architectural conflict, unrelated
  dirty state, failed required verification or unresolved blocking review.
- Next candidate: further Draft editing only after separate Owner direction;
  Not Activated. TASK131 DEFERRED; TASK134 remains local outside main.

## Existing Coverage Report / Verification Plan (before test mutation)

Existing: Studio59 Node cases (TASK135 creation and TASK136 listener editor),
Dashboard15; Go embedded asset/CSP surface; configurationversion service and
handler timeout range/default/state/scope tests, detached repository copies.
Gap: Studio timeout load/save, section isolation, bounds, unknown result,
dirty/cancel/stale/duplicate guards and actual desktop/mobile persistence.
Plan: preserve existing tests; add focused timeout proofs; use real local
Control Service and browser for same-ID/count/state/unrelated-field checks;
run Node, focused/full Go tests, build, vet, format/diff and doc checks.
Race: assess applicability; explicitly record NOT RUN if UI-only changes.
Added Proof Tests / Added Regression Tests: 46 added Studio cases, preserving
the existing 59 Studio and 15 Dashboard cases (120 total); Go asset/API surface
assertions admit only the new timeout endpoint and accessible field surface.
Remaining Limitations: no revision/ETag/CAS, trusted-local unauthenticated
service, in-memory data; no stronger guarantees admitted.

## Size Guard / Documentation Baseline

One timeout section editor, no new package/API/dependency; target <=15 paths,
<500 production lines. Reassess if exceeded. Expected UI/test files plus task
record/index/context/current-state, mirrored Quick Start/MASTER_PLAN, CHANGELOG.
PROCESS002 applicability: those documents require factual task/capability
synchronization; ADR/ARCH/DP/spec-decisions N/A unless admission finds contract
change. README screenshots remain valid for preserved creation flow.
Published TASK136 projected status is historical; this record will synchronize
current navigation without modifying historical TASK134/136 records.

## Capability Inventory (Phase 0, baseline)

Source paths below are relative to repository root. For every existing domain
property in rows 1–12: **model YES; GET YES** via
`internal/configurationversion/handler.go:Handler.list/RegisterRoutes` ->
`service.go:Service.List` -> `memory_repository.go:ListByConfiguration`.
The scoped GET returns complete detached ConfigurationVersion values, including
references; current Studio deliberately projects host/port only. No single-ID
GET exists. All existing section PUTs scope-check and reject non-Draft states.
**Protected-stage changes NO** for metadata editing in every row. Effective
Runtime/security/trust/lifecycle expansion is excluded and would need separate
admission. READY means existing safe backend capability, not active UI scope.

| Property (exact model symbol in internal/configurationversion/model.go) | Update API / validation evidence in same package | Safe Draft edit? | Baseline Studio? | Classification |
|---|---|---|---|---|
| ListenerSettings.Host, Port | PUT /listener; Service.UpdateListener; validateListener in service.go | Yes, address pair | Yes TASK136 | READY, preserved |
| TimeoutSettings.HandshakeSeconds | PUT /listener/timeouts; Service.UpdateTimeouts; validateTimeouts: 1..300 | Yes, complete timeout section | No | READY, selected |
| TimeoutSettings.ReadSeconds | same; 0..86400 | Yes, complete section | No | READY, selected |
| TimeoutSettings.WriteSeconds | same; 1..300 | Yes, complete section | No | READY, selected |
| TimeoutSettings.IdleSeconds | same; 0..86400 | Yes, complete section | No | READY, selected |
| TLSSettings.Enabled, CertificateRef, PrivateKeyRef, MinVersion | PUT /listener/tls; Service.UpdateTLS; validateTLS/validTLSReference in service.go | Yes as metadata with reference-aware UI; no secret values | No | OUT OF SCOPE |
| AuthenticationSettings.Enabled, Providers; AuthenticationProvider.Name, Type, Enabled, Priority | PUT /authentication; Service.UpdateAuthentication; DefaultAuthenticationValidator.Validate/validateAuthenticationProvider in authentication_validator.go | Yes as complete section; dedicated security-aware UX needed | No | OUT OF SCOPE |
| APIKeySettings.Header, SecretRef; BasicSettings.Realm, SecretRef | same; validateAPIKey/validateBasic/validateSecretReference | Yes metadata only, never credentials | No | OUT OF SCOPE |
| JWTSettings.SigningKeys (JWTSigningKey.Name, SecretRef), AllowedAlgorithms, AllowedIssuers, AllowedAudiences, RequiredClaims (JWTRequiredClaim.Name, Value), ClockSkewSeconds | same; validateJWT and nested validation in authentication_validator.go; request defaults in handler.go:authenticationSettingsRequest.settings | Yes metadata only; replacement/reference semantics need separate editor | No | OUT OF SCOPE |
| RoutingSettings.Routes, DefaultHandlerRef | PUT /routing; Service.UpdateRouting; DefaultRoutingValidator.Validate in routing_validator.go | Yes, complete nullable section | No | OUT OF SCOPE |
| Route.ID, Enabled, Priority, Matchers, HandlerRef; Matcher.Type, Value | same; Validate/validateMatcher; uniqueness, ordering, canonicalization | Yes as structured routing metadata | No | OUT OF SCOPE |
| ConfigurationVersion.ID, ConfigurationID, Number, State, CreatedAt, UpdatedAt | No arbitrary metadata update; Create/Publish/Archive own identities/state; server owns times | No arbitrary edits | Selection/status only | OUT OF SCOPE |
| Configuration.Name, Description (internal/configuration/model.go); Workspace.Name, Description (internal/workspace/model.go) | Parent resource services/PUTs, not ConfigurationVersion fields | Parent edits are not Draft edits | Creation names only | OUT OF SCOPE |
| Authorization, Storage, Monitoring, Logging sections anticipated by ADR0002 | Absent from actual ConfigurationVersion model/GET/section PUT/validation | No existing capability | No | BACKEND MISSING |

All section routes are enumerated by `Handler.RegisterRoutes` in handler.go;
timeout request `timeoutSettingsRequest.missingValue` requires all four values.
`MemoryConfigurationVersionRepository.Update` preserves the ID; section services
read current version and assign only their section plus UpdatedAt. No whole-DSL
PUT/PATCH, revision, ETag or CAS exists. No inventory item is falsely labelled
SECURITY BLOCKED: reference editing warrants separate UX admission, while no
repository protected-stage prerequisite prevents the selected timeout metadata.
No raw secret values are admitted by ADR0002/DP002. TLS/auth/routing are existing
backend capabilities, not missing APIs. Inventory does not authorize their UI.

## Architecture Confirmation / Selected Slice

A137-001 independent Architect: APPROVED, blockers 0. Accepted ADR0002 permits
independent metadata sections. Existing `Service.UpdateTimeouts` checks scope,
Draft state and `validateTimeouts`, then changes only Listener.Timeouts/UpdatedAt.
`Handler.updateTimeouts` requires all four fields. No new architecture/API or
Runtime contract. Four timeout values form one independently testable section.
TLS/auth involve reference/security presentation decisions and routing is a
larger structured editor; neither is needed for this slice.

Product contract: add explicit Listener address / Timeouts selector in existing
Draft mode. Default remains listener address. Each Save sends exactly one PUT
for the active section to the pinned version ID; never a multi-request Save.
Switching section uses the same dirty discard/Cancel guard as other navigation.
Display all four current values, validate whole numbers: handshake/write 1..300,
read/idle 0..86400 seconds (zero disables read/idle timeout). No blank-to-zero
coercion. Successful section Save must verify ID/configuration/number/Draft and
the submitted section. A response must not silently change another section's
baseline. Unknown result blocks retry until explicit confirmed fresh same-ID
read; failed reload retains inputs and unknown lock. Duplicate Save, stale reads,
unload warning and failed-input preservation apply equally to both sections.
Only allowlisted values enter UI state; no raw JSON or secret refs rendered.
Published/Validated/Archived cannot be edited. Concurrent writers have no CAS.

## Implementation / Role Handoff

Developer implemented only the admitted section editor. Existing-Draft mode
defaults to Listener address; Timeouts projects four validated numbers from the
fresh scoped GET. Section Save pins the existing identity, sends all required
section fields in one PUT, checks the returned identity/state/number/values,
and acknowledges only the saved section. Other section baselines are not
silently adopted from responses. Server-side section assignments preserve
unrelated configuration values. Section switching restores opened values only
after the discard guard accepts; Cancel keeps section, selection and text.

Inputs and navigation lock during Save; duplicate actions cannot redirect or
repeat the request. Unknown transport/5xx/malformed outcomes retain inputs and
block Save across section switches until explicit confirmed same-ID reload.
Failed/missing/non-Draft reload keeps input and unknown lock. Existing generation
and abort guards reject stale reads. No create/publish fallback exists in the
editor. TASK135 creation flows remain unchanged; TASK136 address Save persists.

Architect A137-001 APPROVED; Developer four-file frozen handoff; root Tester
independently ran Node, native Go regression/build/vet and real Chrome proofs.
Independent final Reviewer verdict resolves from the matching terminal envelope.
Initial sandbox branch creation denied before product edits; read-only status
confirmed only the new task record and unchanged main OID. Supported escalation
created the intended local task branch with that first content change preserved.
No main history mutation, staging, commit or publication occurred.

## Final Test Matrix

| Check / required scenario | Actual evidence |
|---|---|
| Load existing values | Node timeout projection proofs; Chrome loads 17/23/29/31 from real scoped GET |
| Edit/Save/reopen/persistence | Chrome changes all four values; explicit reload reads 41/0/300/86400 |
| Same ID/no duplicate/no accidental publish | Chrome checks ID/number/CreatedAt/state and exactly one version; one timeout PUT even with repeated submit |
| Published rejection | Real API publishes opened fixture externally, subsequent UI Save fails, typed 44 retained, Published stored 43 unchanged; Node HTTP409/response-state cases |
| Unrelated fields | Chrome preserves host/port/TLS/auth/routing/CreatedAt; real TASK136 address Save preserves timeout section; Node verifies response-baseline isolation both directions |
| Invalid input | Node lower/upper limits, blank, fractions, invalid GET values; Chrome blank read rejects with zero mutations; range messages shown |
| API/network errors | Node 400/404/409 retain selected Draft/input; network/500/invalid JSON/ID/scope/number/state/settings become unknown |
| Unknown Save result | Chrome deliberately drops real PUT response after server commits; input retained, Save locked; confirmed fresh GET reconciles same ID without attribution/retry |
| Dirty/Cancel/unload | Node section/mode/parent/Draft/reload/reset Cancel and beforeunload; Chrome native Cancel retains timeout section/value, actual reload discard dialog exercised |
| Stale/duplicate requests | Node stale opens/reloads and pending navigation/duplicate submit; real repeated submit yields one PUT |
| Existing TASK135/136 | All previous 59 Studio tests retained; actual Chrome three creation modes plus address Save pass |
| Desktop/mobile | Chrome1280 desktop and390 mobile, screenshot inspection; mobile Save persists same Draft; scrollWidth<=390, fields usable, no JS exceptions |
| Node | `node --test internal/configurationstudio/studio_test.js internal/configurationstudio/dashboard_test.js`: exit0,120PASS (105Studio+15Dashboard),0fail/skip/cancel |
| Focused Go | Developer `go test -mod=readonly -count=1 -timeout=3m ./internal/configurationstudio ./internal/configurationversion`: exit0 |
| Full Go regression | Native `go test -mod=readonly -count=1 -timeout=3m ./...`: exit0,48.625s |
| Build | Native `go build -mod=readonly ./...`: exit0,3.338s; browser service separately built from same frozen UI |
| Vet | Native `go vet -mod=readonly ./...`: exit0,5.494s |
| Format | `gofmt -l internal/configurationstudio/handler_test.go`: empty; `git diff --check`: exit0 |
| Race | NOT RUN: production diff is HTML/JS only, no Go concurrency/synchronization/lifecycle source change; no previous race PASS transferred |
| Modules | tidy N/A: no imports/dependencies/module boundaries changed; go.mod/go.sum unchanged; all Go commands use -mod=readonly |

Initial sandbox regression exit1 denied loopback sockets in existing session
tests; preserved sandbox log is environment evidence, not a product failure
waiver. Full native rerun above passed. Sandbox Chrome did not start; native
Chrome succeeded. A subsequent harness rerun first encountered its stale Chrome
debugging-port file before fixture creation; removing that exact stale file and
rerunning gave final21/21 checks PASS. Product files were unchanged throughout.

Reproducible external artifacts in this chat's visualizations directory:
`browser137.cjs`, `task137-browser.json`, `task137-desktop.png`,
`task137-mobile.png`, `task137-node.log`, `task137-verification.json`,
`task137-regression.log`, `task137-build.log`, `task137-vet.log`,
`task137-regression-sandbox.log`, `task137-verification-sandbox.json`.
Harness starts isolated loopback Control Service/headless Chrome and closes
both. No installs or persistent project/environment changes. Browser fixture
publication tests are isolated in-memory data, not repository publication.

## Documentation Synchronization / Exact Changed-File Manifest

All13 paths Required; Questionable0/Removable0. Size Guard PASS: one independently
deliverable timeout section editor, no package/dependency/new contract and fewer
than500 production additions. Each path is necessary for behavior, proof or
mandatory recovery/user-facing documentation; no unrelated refactoring.

| Exact path | Reason / classification |
|---|---|
| internal/configurationstudio/assets/index.html | Required: selector, timeout fields/labels/help |
| internal/configurationstudio/assets/studio.js | Required: projection, validation, guarded section Save |
| internal/configurationstudio/studio_test.js | Required:46 new proofs plus full-model fixtures |
| internal/configurationstudio/handler_test.go | Required: embedded surface and admitted endpoint assertions |
| docs/tasks/TASK-137-EXTENDED-DRAFT-EDITING.md | Required: contract/inventory/verification/recovery/Pre-Commit |
| docs/tasks/README.md | Required: current task navigation/publication baseline |
| .ai/PROJECT_CONTEXT.md | Required: current task and durable baseline |
| spec/current-state.md | Required: actual capability/task state |
| docs/en/quick-start.md | Required: user instructions and limits |
| docs/ru/quick-start.md | Required: semantic mirror |
| docs/en/roadmap/MASTER_PLAN.md | Required: factual capability only, no milestone promotion |
| docs/ru/roadmap/MASTER_PLAN.md | Required: semantic mirror |
| CHANGELOG.md | Required: user-facing timeout editor |

PROCESS002 applicability: task/index/current-state/context synchronized;
QuickStart and MASTER_PLAN EN/RU preserve structure and meaning. Related
ADR/ARCH/DP/spec-decisions N/A: no design/status/architecture decision changed.
Root README/screenshots N/A: preserved illustrated creation flow remains valid;
no new product screenshot claim. Historical task records untouched, including
TASK134/136. TASK102/TASK131/protected-stage paths absent from diff. Baseline
project-state TASK136 projection reconciled to known local PR135 merge plus
explicitly attributed Owner terminal report, without fresh remote/CI claims.
Link/parity/scope results and exact subject rows resolve from terminal envelope.

## Remaining Limitations / Pre-Commit Report

Only four timeout values added; TLS/authentication/routing remain separate
future UI tasks. Metadata edit is not Runtime activation or configuration reload.
No revision/ETag/CAS, so concurrent clients can overwrite a saved section.
In-memory/trusted-local unauthenticated boundary unchanged. Unknown-result
reload observes server state and cannot attribute an earlier operation. Native
beforeunload warnings cannot be guaranteed on crash/forced close/mobile process
termination; event guard has executable coverage. Mobile checks use real Chrome
device metrics, not a physical phone or Safari. Race NOT RUN as justified above.

Pre-Commit gate requires independent final review and Coordinator Acceptance of
the exact canonical subject. This record's projected In Progress is stable;
latest matching append-only envelope carries actual verdict/acceptance. No
stage/commit authorized. No push/PR/merge authorized. STOP after that gate.
TASK134 untouched outside main; TASK131 DEFERRED; TASK102 untouched; Gate11 OPEN;
protected stages Not Activated. Next candidate: a separately admitted routing
metadata editor, Not Activated; no work started and no backend change proposed.

## Recovery Evidence Envelope

### V137-001 — Tester / PROCESS002 / Scope Freeze

Target: TASK137, feature/task-137-extended-draft-editing, HEAD/base
`0dfd13bf49db2a3bd63d03111f46cedcb07261d7`. Object format sha1.
Canonical subject-manifest `a6afe0451ad84252d632451b5039a30aa50b4678`.
Independent recomputation: external `manifest137.py`, raw object convention
PROCESS001; no staging. Ordered rows (path | projection | state | mode | OID):

```text
.ai/PROJECT_CONTEXT.md | full | present | 100644 | 225509cbe0545c312029e54598a57c1806b0e634
CHANGELOG.md | full | present | 100644 | 9b8768548e6e62d9b83af18bb3a5c5f834697e29
docs/en/quick-start.md | full | present | 100644 | d770d01a08214ac5cce3930ba53610f9f9b143de
docs/en/roadmap/MASTER_PLAN.md | full | present | 100644 | b7b98b91292fd2950d9151badb99853074f73f90
docs/ru/quick-start.md | full | present | 100644 | 2330d6aa6b31c79e460859d489aea08465db1429
docs/ru/roadmap/MASTER_PLAN.md | full | present | 100644 | 697c44bc8e05460ce2e61fc6f12b51ad81cbc016
docs/tasks/README.md | full | present | 100644 | b352bbc7b528bb7e3527bde4158fd6f6dbce18a5
docs/tasks/TASK-137-EXTENDED-DRAFT-EDITING.md | task-record-v1 | present | 100644 | dac7ab5be948d19e90764f2624a27f509efb633d
internal/configurationstudio/assets/index.html | full | present | 100644 | c3afce889ee2852da300b0e3b456ff986280b87b
internal/configurationstudio/assets/studio.js | full | present | 100644 | 4469bab2a301fc84998872262de4c8fe56978e09
internal/configurationstudio/handler_test.go | full | present | 100644 | 88878903de5ee603b605d67f09655e2b337e78b8
internal/configurationstudio/studio_test.js | full | present | 100644 | 4b9955e15dd72adcaee202baebbe30303fdb5ac7
spec/current-state.md | full | present | 100644 | fbd62eba93e93a8bbccb06cf90279f83a9765df9
```

Tester: PASS for exact four frozen UI/test blobs in rows. Node120/120 exit0;
focused Go exit0; full native regression/build/vet exit0, exact commands/timings
and all scenario results in Final Test Matrix. Browser21/21 checks PASS, no JS
exceptions. Root viewed desktop/mobile images; labels, field flow and single
column mobile layout legible, no horizontal overflow. Race NOT RUN/N/A for
HTML/JS-only production changes. Sandbox denied socket/cache/Chrome operations
preserved as limitations; successful native rerun replaces no historical fact.

PROCESS002: Synchronized.335 local Markdown targets resolve; QuickStart8/8 and
MASTER_PLAN36/36 heading levels match; EN/RU section-save instructions and limits
semantically aligned. UTF8/conflict checks pass. Scope13 Required/0Questionable/
0Removable. Protected TASK134 branch remains c12a3f531f497475062a1672c0abb66872aab9fc;
main/origin/main remain trusted baseline; no staged paths. Four raw source OIDs
match Developer freeze and exact tested bytes. No production changes afterward.

Independent final Review and Coordinator Acceptance: Proven Not Started at this
entry. Next stage: independent review of this exact projected subject; no commit
or publication authorization.

### R137-001 — Independent Final Review

Independent non-author Reviewer: Approved; blocking0/unresolved0. Exact subject
`a6afe0451ad84252d632451b5039a30aa50b4678`, same13 ordered rows as V137-001,
branch feature/task-137-extended-draft-editing, HEAD/base
0dfd13bf49db2a3bd63d03111f46cedcb07261d7. Reviewer independently rebuilt all
rows from raw bytes, verified tracked100644 modes and task-record-v1 projection,
reproduced manifest, reran Node120PASS and diff-check0, inspected native Go and
browser21PASS evidence. Mapping, data loss, version duplication, unintended
publication, security/reference exposure, unrelated overwrites, stale state,
API compatibility and protected-stage boundaries reviewed without findings.
Dirty/Cancel, pending locks, unknown reconciliation and section baselines meet
admitted contract. EN/RU meaning and architecture statuses preserved.

Removal assessment: no listed change can be removed while retaining complete
Definition of Done. All13 path purposes in the manifest table are Required;
Questionable0/Removable0. No backend/API/dependency/Runtime/protected-stage code
changes; TASK134/131/102 outside diff. Disclosed no-CAS, in-memory trusted-local,
mobile-emulation and unload-warning limitations accepted as accurate. Review
allows Coordinator Acceptance only; no commit/publication authorization.

### CA137-001 — Coordinator Acceptance / Pre-Commit STOP

Coordinator accepts the implemented bounded timeout metadata editor, exact
subject `a6afe0451ad84252d632451b5039a30aa50b4678` with V137-001's13 ordered
rows, HEAD/base0dfd13bf49db2a3bd63d03111f46cedcb07261d7 and branch
feature/task-137-extended-draft-editing. A137-001 admission, Developer handoff,
V137-001 verification/PROCESS002/scope and R137-001 independent review are
complete; blockers0. No pre-implementation architecture-document stage needed:
existing metadata API/ADR0002 contract confirmed without change. No Publisher
stage: separate Owner gate absent. No new task or successor activated.

Post-review integrity: canonical subject recomputed unchanged before this
append; only append-only terminal envelope evidence added. Acceptance does not
self-attest final envelope bytes. Project-state remains verification-stable
projected In Progress per PROCESS001; this matching latest entry is the actual
accepted checkpoint. Final independent Review and Coordinator Acceptance are
Proven Completed. First incomplete gate: separate exact Owner Commit Gate.

Pre-Commit Report: READY for separate Owner review/authorization of one task
commit of this exact accepted13-file subject. Nothing staged; HEAD unchanged;
main/origin/main trusted baseline; TASK134 branch retains original OID. No
commit, push, PR or merge performed. Publication requires its own later exact
Owner Publication Gate after a separately authorized commit. No authorization
is inferred here. STOP at Pre-Commit Gate. Gate11 OPEN; protected stages Not
Activated; TASK131 DEFERRED; TASK102/TASK134 untouched. Routing editor remains
only a recommendation, Not Activated.

### CG137-001 — Owner Commit Gate

Current exact Owner command `Разрешаю коммит.` received after CA137-001.
Gate class: Coordinator Accepted. Target: one TASK137 commit on
feature/task-137-extended-draft-editing, parent
0dfd13bf49db2a3bd63d03111f46cedcb07261d7, accepted subject
`a6afe0451ad84252d632451b5039a30aa50b4678`, exact13 V137-001 paths.
Current manifest independently recomputed unchanged; no staged files or
post-acceptance subject changes, temporary/generated/unrelated paths absent.
Only append-only terminal metadata is added for this gate. Node120PASS,
diff-check0, gofmt empty and335 links/parity checks reconfirmed. Existing native
Go/browser evidence remains applicable to unchanged source/test blobs.
Commit message: `feat(studio): edit existing draft timeouts`, matching repository
feat(studio) history; no additional message policy found. No GPG/DCO/sign-off
added. Stage must preserve exact raw bytes (no LF/CRLF equivalence); compare
all staged modes/blobs and projected manifest before commit. Unknown commit
outcome must be reconciled via HEAD/history, never blindly repeated.
Authorization covers exactly one commit, no push/PR/merge. Commit OID and final
tree will be inspected externally after creation, not embedded into their own
commit. Next: stage exact subject -> verify -> commit once -> inspect -> STOP.

### CG137-002 — Exact Staged Bytes / EOL and Size Guard Audit

Staging preserved every accepted raw byte, mode and canonical subject; no
LF/CRLF equivalence substituted. Three accepted raw files are CRLF while their
HEAD blobs are LF: assets/index.html142 CRLF lines, assets/studio.js453,
studio_test.js566. All staged bytes equal current accepted working bytes.
Default cached diff-check reports these CR-at-EOL bytes as whitespace; explicit
`git -c core.whitespace=cr-at-eol diff --cached --check` passes without changing
bytes, repository configuration or suppressing actual trailing spaces. No bare
CR or repository LF-only requirement found. Initial whitespace gate stopped
before commit; commit was Proven Not Started. No commit retry occurred.

Size Guard explicitly reassessed: raw production additions595 cross the500
trigger because exact accepted line endings differ from HEAD. EOL-aware diff
has21 HTML and70 JavaScript additions, one timeout editor,13 Required paths,
no new package/contract/dependency or second independently deliverable behavior.
Coordinator confirms bounded coherence; no split needed. Three files contain
required behavioral changes, not removable formatting-only files. Converting
them to LF at this gate would change the accepted raw identity and is not done.
Independent Reviewer reproduced canonical subject and exact staged raw equality;
final gate verdict follows. This audit is append-only envelope metadata only.
Restage this record's envelope bytes, verify all13 exact staged bytes/modes and
subject again, then commit once under the current exact Owner gate if approved.

### CG-R137-001 — Independent Commit Gate Audit

Independent non-author Reviewer: PASS after explicit Size Guard/EOL audit,
blockers0. All13 staged raw files/modes independently checked; canonical subject
remains a6afe0451ad84252d632451b5039a30aa50b4678. Valid CRLF, no bare CR;
cr-at-eol whitespace check passes with no normalization or identity substitution.
No LF-only repository rule.595 raw production additions require explicit audit;
91 semantic additions, one behavior and no added package/API/architecture
contract justify the bounded slice. CG137-002 records Coordinator reaffirmation.
Re-stage envelope and repeat exact staged checks before one commit. Existing
verification/review/acceptance and current Owner gate remain valid because
accepted subject bytes are unchanged. No new authorization inferred or requested.

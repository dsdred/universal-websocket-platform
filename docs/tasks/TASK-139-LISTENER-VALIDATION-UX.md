# TASK-139 — Listener Host/Port Validation UX

## Status

In Progress. Owner approved the bounded host contract amendment. Final review and acceptance resolve from the terminal evidence envelope.

## Task Contract

- Mode: Implementation, conditional on Phase 0/1/3 admission. Separate branch `feature/task-139-listener-validation-ux` in an isolated worktree at published main `a41e1569710b0245de584919cea2ed1d610d87ad` (PR #136).
- Why now: Owner found host/port feedback gaps during open TASK-138 acceptance.
- Definition of Done: reproduce both examples in both workflows, determine backend behavior, then implement and independently review field validation only if it can honor existing contracts and the Owner's invalid-input requirements. If a required backend contract change is discovered, stop for Owner decision with evidence.
- Sources: [ADR-0002](../en/adr/0002-configuration-dsl.md), [TASK-137](TASK-137-EXTENDED-DRAFT-EDITING.md), Quick Start, Studio DOM and JS, `configurationversion` public API, validators and tests.
- Existing coverage: TASK-137 has 105 Studio Node tests and 15 Dashboard tests, including a local invalid listener input case and retained input for Draft HTTP errors. Go tests exercise Listener validation and state/scope. Gaps: both exact Owner probes through real browser input/blur/submit and API, field associations/focus, lexical numeric handling and correction.
- Verification plan: actual isolated Control Service and Chrome baseline replay, direct Listener API round trips and published/Draft checks; if admitted, focused Node and browser regressions, full Go regression/build/vet/format, race if applicable, independent review. Coordinator and Tester handle admission and evidence; independent Reviewer checks findings and any final diff.
- Out of scope: network semantics, protected Runtime stages, new backend contract without Owner decision, TASK-138 worktree or test data changes, TASK-134, TASK-131 activation, TASK-102, commit or publication. Gate 11 Open; protected stages Not Activated.
- Existing TASK-138 worktree and service are separate and must remain available. No successor is activated. Size Guard: one bounded validation UX task.

## Approved amendment and role handoff

The Owner approved a write-time rejection of ASCII dotted numeric-only hosts
that fail bare IPv4 parsing, including numeric trailing dots. Existing valid
IPv4, IPv6, bind addresses and other supported hostnames retain their syntax.
No DNS lookup, connectivity check, Runtime change or stored data rewrite.
Historical out-of-contract values stay readable and can be corrected on the
next address write. Independent admission review identified this as the narrow
contract change. The Owner's current direction resolves that blocker.

Studio uses the same accepted address syntax, plain decimal port input bounded
to 1–65535 before conversion, inline controls with inline/submit correction and
accessible first-error focus. Listener API validation details are projected to
the appropriate field. Create validates before any mutation and reports known
created resources if a subsequent Listener write fails. The existing multi-step
flow, retained input, and unknown-outcome guards remain; no rollback.

Coordinator assigns the Studio and backend patches plus their regression tests
to separate Developer contributors under this existing contract. Coordinator
coordinates verification and documentation, followed by independent non-author
review. The TASK-138 service remains outside this test scope.

## Contract findings

| Setting | Existing public behavior | Evidence |
|---|---|---|
| Host | Trims surrounding Go whitespace, rejects empty and whitespace-only, and checks the trimmed limit of 255 characters. Accepts Go `net.ParseIP` literals or the repository's ASCII hostname grammar. | [Service.validateListener](../../internal/configurationversion/service.go#L371) |
| IPv4, IPv6 and bind addresses | Accepts bare IPv4/IPv6, including `0.0.0.0` and `::`. Bracketed IPv6 and zone suffixes are rejected by this metadata contract. IPv6 brackets are used by Studio for a rendered URL. | [Service.validateListener](../../internal/configurationversion/service.go#L371), [Studio.preliminaryURL](../../internal/configurationstudio/assets/studio.js#L250), actual API cases |
| Hostnames | Allows ASCII letters, digits and internal hyphens, labels of 1–63 bytes, optional terminal dot, and total hostname length up to 255 bytes. Numeric labels are allowed. The fallback does not require a letter in any label or the last label. The Owner's dotted host is accepted by this path even though IP parsing fails. No DNS lookup or bind is performed. | [validHostname](../../internal/configurationversion/service.go#L389), [ADR-0002](../en/adr/0002-configuration-dsl.md), actual API cases |
| Port | JSON numeric integer in 1–65535. The `uint16` decoder rejects overflow, negative, fractional and string input as HTTP 400 `invalid_request`. Zero is representable but explicitly rejected as HTTP 400 `validation_failed`. Omitted/null port is zero after decoding and is rejected for an otherwise valid host. | [ListenerSettings](../../internal/configurationversion/model.go#L125), [updateListener](../../internal/configurationversion/handler.go#L282), [validateListener](../../internal/configurationversion/service.go#L371) |
| Studio input | Shared host text input uses `maxlength=255`; port is a number input with min/max/step. The form is `novalidate`. Create uses `inputValues` to check nonempty trimmed host and numeric port range. Draft uses its own nonempty/length/numeric checks. Both call JavaScript `Number` before validating the port; neither checks the backend hostname grammar or renders a field error on input/blur. | [Studio markup](../../internal/configurationstudio/assets/index.html#L89), [inputValues](../../internal/configurationstudio/assets/studio.js#L229), [saveDraft](../../internal/configurationstudio/assets/studio.js#L381) |
| API errors | Backend emits `error.code` and `error.message`; validation failure messages identify the field. Studio `requestJSON` throws the status without reading that error body. Create and Save display generic status below the form. | [writeServiceError](../../internal/configurationversion/handler.go#L396), [WriteError](../../internal/http/json.go#L44), [requestJSON](../../internal/configurationstudio/assets/studio.js#L118) |

## Reproduction on published TASK-137

Built `./cmd/control-service` from the isolated clean published checkout. Actual Chrome CDP inserted the exact text through Chrome's input system, blurred the field, clicked the real Create/Save control, and captured network mutations and status. The dedicated service bound `127.0.0.1:18139` with fresh in-memory data. The TASK-138 service on port 18138 was not used. All cases lacked a field association or inline error during input and blur.

| Input and flow | During input / blur | On Create or Save | Backend outcome |
|---|---|---|---|
| Port `8777777777777777777777`, Configuration creation | Text preserved; Chrome validity reports `rangeOverflow=true`, with general status still “Ready.” | General status “Listener port must be a whole number between 1 and 65535.” Zero HTTP mutations. | Direct PUT with the exact unrounded JSON numeric text returns HTTP 400 `invalid_request`. |
| Same port, existing Draft | Text preserved; same browser overflow, general status still shows opened Draft. | General status “Enter a listener host and whole port between 1 and 65535.” Zero HTTP mutations. | Same direct API rejection. |
| Host `127.0.0.1.4.5.6.7.8.9.0`, Configuration creation | Text preserved. No field error. | Five normal create-flow writes: Workspace, Configuration, Draft, listener PUT, publish. | Fresh public GET confirmed Published version ID 2 with the exact host. |
| Same host, existing Draft | Text preserved. No field error. | One listener PUT. Studio reported successful Save to the existing Draft. | Fresh GET confirmed same version ID 1 still Draft with the exact host. |
| Control host `http://localhost`, Configuration creation | No field error. | Four attempted writes. Workspace, Configuration and Draft were created first. Listener request failed; no publish request. Generic HTTP 400 and inspect-Dashboard advice. | Direct PUT HTTP 400 `validation_failed`, message `host: must be a valid IP address or hostname`. |
| Same control host, existing Draft | No field error. | One failed PUT. Generic HTTP 400, unsaved input retained. | Same server validation rejection. |

Nineteen direct API cases passed their expected outcomes, including both examples, empty host, zero, 1, 65535, 65536, negative/fractional/empty ports, IPv4, IPv6, wildcard binds, DNS name with terminal dot, bracket/zone rejection and valid host whitespace trimming. The full browser captures and HTTP payloads are in `task139-repro.json`. The first harness run stalled at a native beforeunload prompt while changing pages. Identified and stopped only its isolated processes. The second run explicitly handled the prompt and passed. The reproduction service and headless browser close at the end of the run.

The port defect is missing inline feedback while typing or blurring and generic error presentation on Draft submission. Range validation already exists on submit and the API independently enforces it. For actually malformed hostnames, frontend format validation and field-specific backend feedback are missing. For the exact Owner host, there is a product-policy gap: both Studio and the API treat it as valid. A frontend-only block would no longer match the API.

## Initial admission decision and alternatives

**Initial verdict: STOP pending Owner decision.** The requested rejection of the exact Owner host would change the allowed host set. The current code does perform backend validation, but does not enforce the requested interpretation of dotted numeric hosts. Phase 3 therefore requires a separate Owner decision.

Recommended minimal contract change if this host must be rejected: after trimming, preserve any valid IP literal first. If a remaining host contains a dot and consists only of ASCII digits and dots, reject it as an invalid dotted numeric address rather than falling back to hostname validation. Apply the same rule in Studio and the existing backend validator. This would newly reject some hostname strings accepted today, including dotted numeric names with a trailing dot. Single numeric labels, other supported hostname patterns, IPv6 and special bind addresses would retain their current semantics. Existing stored values should not be rewritten; compatibility for Drafts already carrying the newly rejected values must be documented with the approved contract.

Alternative if the existing accepted host set takes priority: preserve the exact example as a hostname, and improve inline validation and API-error presentation for existing port and host constraints. This would not make the exact example invalid. No network or Runtime semantics change is needed for either choice.

## Initial admission report

At the initial checkpoint, Pre-Commit was not reached. Product code and repository tests are unchanged. No implementation diff, Coordinator Acceptance or commit-ready verdict exists. Changed-file manifest contains only `docs/tasks/TASK-139-LISTENER-VALIDATION-UX.md`, the admission and recovery record. TASK-138 checklist and navigation edits are outside this worktree. No staging, commit, push, PR or merge. TASK-134 and TASK-102 refs preserved, TASK-131 Deferred, Gate 11 Open, protected stages Not Activated.

PROCESS-002 applicability: only the TASK-139 admission record is required. No capability or project-state transition is being published; public Quick Start, mirrors, roadmap, changelog and architecture decisions remain as-is. This is a blocked admission report, not a certified Blocked Evidence Checkpoint.

## Implemented behavior

- The existing Listener address PUT rejects a dotted ASCII numeric-only host
  when IP parsing fails. The request shape and field-specific HTTP 400 shape
  remain unchanged. The API checks before mutating the version.
- Both Studio address forms share lexical host and port validation. The browser
  uses strict IPv4 octets, lexical IPv6 group checks, the existing hostname grammar
  and Go-compatible whitespace trimming. Port uses a text field with numeric
  keyboard hint, checks decimal syntax and range before JS conversion, and
  preserves the original text. Input and blur render errors beside their
  fields. Submission focuses the first invalid field before mutation.
- The listener PUT client reads the existing 400 validation envelope, maps
  only address field errors, and retains input. In Draft mode the form unlocks
  and focuses that field. In creation mode, the result states which parent
  and Draft writes already succeeded and that Studio did not publish the Draft
  after a definite Listener validation failure. It asks the user to inspect
  Dashboard before retry. Existing unknown-outcome handling remains.
- Address Save uses a single PUT for the pinned existing Draft and leaves
  other sections to their existing contracts. Opening and editing timeouts on
  Drafts with historical addresses still works. A later Listener address
  write must correct a newly invalid host.

Architecture confirmation: this is the Owner-approved write validation rule
over the existing ADR-0002 configuration sections. JSON structure, state
transitions and interpretation of stored Published versions are unchanged.
The rule does not run during Publish, read, or a different section update.
No schema migration, new architecture decision, or DNS/Runtime work is needed.

## Verification and limitations

The existing 120 Node cases are retained. Twenty-three additional cases test
exact Owner examples, accepted addresses, malformed and boundary ports, pasted
text and correction, keyboard focus, field-specific backend rejection, both
creation scopes, unchanged Draft outcome, and legacy read/timeout behavior.
The Go embedded test verifies actual HTML associations and input types. Direct
handler tests add accepted host/port pairs and rejected exact request bodies
with unchanged GET snapshots after failures.

In headless Chrome on isolated port 18139, 22 input/blur/submit probes covered the
four bad hosts and seven bad ports in both workflows. Invalid probes sent
zero writes and preserved exact text. A corrected create flow published its
new version. Existing Draft Save sent one PUT and reopened as the same Draft
ID with IPv6 and port 65535. Nine direct API probes returned HTTP 400 and left
the full version snapshot unchanged. For late backend rejection, the harness
altered only the outgoing Listener body to reach the real rejection path after
valid Studio inputs. It confirmed a field-specific error and retained input
on Draft Save, and three earlier successful creation writes with no Publish
on Create. The returned server message and IDs were visible in Studio and
verified through public GET. Mobile Chrome had no horizontal page overflow
and both errors were visible beside their inputs. No scenario used TASK-138.

Creation remains a sequence of requests, so a late response can leave earlier
resources in place. The UI does not roll back or automatically retry. Older
stored addresses are retained, and new validation is enforced at Listener
address write time. There is no new revision check, transactional create,
connectivity validation or Runtime behavior.

## Exact changed-file manifest and Pre-Commit Report

| Path | Scope / necessity |
|---|---|
| `internal/configurationversion/service.go` | Required. Add the approved write-time host guard. |
| `internal/configurationversion/handler_test.go` | Required. Direct HTTP rejection/acceptance and no-change proof. |
| `internal/configurationstudio/assets/index.html` | Required. Accessible error associations and text entry. |
| `internal/configurationstudio/assets/studio.css` | Required. Visible inline and invalid-field treatments. |
| `internal/configurationstudio/assets/studio.js` | Required. Shared validation, field errors, and partial outcomes. |
| `internal/configurationstudio/handler_test.go` | Required. Proof the served markup exposes the promised controls. |
| `internal/configurationstudio/studio_test.js` | Required. Regression coverage for both workflows and failure paths. |
| `docs/tasks/TASK-139-LISTENER-VALIDATION-UX.md` | Required. Contract, evidence, recovery and gate anchor. |
| `docs/tasks/README.md` | Required. Task navigation and published baseline. |
| `.ai/PROJECT_CONTEXT.md` | Required. Current task and published baseline. |
| `spec/current-state.md` | Required. Existing write behavior and validation UX. |
| `docs/en/quick-start.md` | Required. User guidance and partial-write limitations. |
| `docs/ru/quick-start.md` | Required. Semantic mirror. |
| `CHANGELOG.md` | Required. User-visible validation and write contract announcement. |

PROCESS-002 applies to the task record, task index, current state, project
context, EN/RU Quick Start and changelog. Master plan/milestones, README,
ADR/ARCH/DP and decision index are not affected. No dependency, package,
architecture, or protected-stage change. Size Guard: 14 paths and one
independently verifiable Listener address validation change. No unrelated
format rewrites. Existing CRLF UI blobs stay CRLF; changed Go and CSS blobs use LF.
No TASK-138 checklist edits are included. Terminal review/verification/acceptance
must match the exact projected manifest below. No staging or commit, and no
publication authorized. Stop at the accepted Pre-Commit boundary.

## Recovery Evidence Envelope

Baseline HEAD `a41e1569710b0245de584919cea2ed1d610d87ad`, branch `feature/task-139-listener-validation-ux`. Initial pending checkpoint: explicit Owner host-policy decision, resolved by the approved amendment above. Exact browser/API evidence: `repro139.cjs`, `task139-repro.json`, `task139-repro-service.log` in the task artifacts.

### V139-001 — Baseline verification

- Chrome exact input/blur/submit and HTTP evidence: six browser probes passed their expected outcomes. Nineteen direct API cases passed. Reproduction JSON SHA-256 `a5d42f226c6df3f1499356834517f4551b69c05dbc5080d72f42ef666940e77b`.
- `node --test internal/configurationstudio/studio_test.js internal/configurationstudio/dashboard_test.js`: 120 passed, exit 0.
- Native `go test -mod=readonly -count=1 -timeout=3m ./...`, `go build -mod=readonly ./...`, and `go vet -mod=readonly ./...`: all exit 0. Logs are under `task139-baseline-*.log`.
- Native worktree `gofmt -l` reports the existing Windows CRLF checkout representation. A byte-exact formatter comparison against all 291 committed Go blobs reported zero mismatches. No Go source was edited. Race NOT RUN because there is no production, lifecycle or concurrency diff. Implementation regressions, paste/correction behavior and accessible field feedback remain pending admission.
- TASK-138 still responded HTTP 200 on port 18138 under original PID 37772 after reproduction. Port 18139 was closed.

### R139-001 — Independent admission review

Independent non-author reviewer confirmed STOP pending Owner decision. Runtime and network changes are unnecessary, but rejecting the exact host in both layers requires changing the write contract. Reviewer corroborated the direct 200, Published create, one same-ID Draft update, both zero-write port submissions, the direct overflow 400, and the missing field-specific UI for a genuine backend rejection. This is a review of admission evidence, not implementation review or Coordinator Acceptance. The one-file documentation scope is required for handoff. Pre-Commit Gate is not reached.

### V139-002 — Final implemented verification and scope audit

Target: TASK-139, feature/task-139-listener-validation-ux, HEAD/base
a41e1569710b0245de584919cea2ed1d610d87ad. Object format sha1. Canonical subject
c2a571ef148b2c9d3b1ea23bba11857f73d038cb, calculated with manifest139.py.
Ordered paths, projections, states, modes and raw projected Git blob OIDs:

```text
.ai/PROJECT_CONTEXT.md | full | present | 100644 | 0966955225d3f375db717d9632c18ac97daaecad
CHANGELOG.md | full | present | 100644 | 3aa9ff6c08dda94e0f50288bf74367fb968d9340
docs/en/quick-start.md | full | present | 100644 | fa79533e89ee9c74c6daab0852796852940263ed
docs/ru/quick-start.md | full | present | 100644 | a57ce846ae54a220dbcf75f3c5ca75ee62d73d7f
docs/tasks/README.md | full | present | 100644 | be6e87204916550dbf0b38688633a012c08a5065
docs/tasks/TASK-139-LISTENER-VALIDATION-UX.md | task-record-v1 | present | 100644 | f430bd5542e5ed4fc24a9a195f96bae3bfb858fd
internal/configurationstudio/assets/index.html | full | present | 100644 | 5ae09228e59f38bef6fb28fd7eb3142302231e68
internal/configurationstudio/assets/studio.css | full | present | 100644 | 0963856059c077c9c2c6040e156dcd1e1495e81e
internal/configurationstudio/assets/studio.js | full | present | 100644 | a4b325a603f7a673f752c546ff134ffd24729b92
internal/configurationstudio/handler_test.go | full | present | 100644 | cabf3c5c32068fd764b917dc58e60fc938d4e6ba
internal/configurationstudio/studio_test.js | full | present | 100644 | ef2aff1272e890c144fb4154e495f501618c766b
internal/configurationversion/handler_test.go | full | present | 100644 | 801e3ea50b0093ff2d548f5fa4681f6a885dca7e
internal/configurationversion/service.go | full | present | 100644 | e91f179af05ff1e58dc3c5914fcf5381072bde69
spec/current-state.md | full | present | 100644 | 65f583dbe97fa479d651079e96d5f4460293c6ef
```

- Final Node combined test command: 143 passed, 0 failed/skipped/cancelled
  (128 Studio, 15 Dashboard). 23 new cases. Focused Go in Studio and
  configurationversion passed, including served field associations and
  direct HTTP rejection with full unchanged Draft snapshots.
- Final native `go test -mod=readonly -count=1 -timeout=3m ./...`,
  `go build -mod=readonly ./...`, `go vet -mod=readonly ./...`: exit 0.
  Changed Go formatter list empty. No dependency/module changes.
- Initial regression failed the existing Studio no-literal-http:// embedded
  asset check. Replaced URL-based local IPv6 parsing with lexical group
  validation. The final suite passed with the existing guard intact.
- Final Chrome/browser harness exit 0: 22 zero-write invalid attempts,
  nine direct HTTP 400 probes with unchanged persisted versions, successful
  Publish flow, same-ID Draft address Save, server-side field error on each
  flow, true partial-write readback, and mobile inline screenshot. Only the
  outgoing Listener body was perturbed to exercise later server rejection.
- Race NOT RUN: validation and rendering changes contain no new shared-state,
  lifecycle, synchronization or concurrent server path. Existing Go tests ran
  natively. No test result is transferred from a prior publication.
- PROCESS-002 synchronized: 176 local links checked and EN/RU Quick Start
  structural parity 9/9, semantic validation and partial-creation guidance
  checked. Scope 14 Required, 0 Questionable, 0 Removable. Git whitespace
  check used `git -c core.autocrlf=false -c core.whitespace=cr-at-eol diff --check`
  because three published UI blobs use CRLF. No normalization substitutes for
  raw identity. No generated/unrelated file or protected path in the diff.
- Logs: task139-final-{node,test,build,vet}.log; prior failed attempt retained
  as task139-initial-regression.log; final Chrome/script artifacts
  task139-implemented-browser.json, verify139.cjs and
  task139-mobile-validation.png. Build and fixtures outside the repository.

Tester verdict PASS on these exact source bytes. Final independent review
and Coordinator Acceptance are separate entries.

### R139-002 — Independent final implementation review

Approved. Zero findings or blockers. Exact canonical subject
c2a571ef148b2c9d3b1ea23bba11857f73d038cb, same 14 ordered V139-002 rows and
same published HEAD. The independent non-author recomputed all raw/projected
rows, verified an empty index and reran Node 143/143. Reviewed host and port
parity, invalid submission bypasses, parsing, retained input, partial Create,
actual API 400, accessible feedback, one same-ID Draft PUT, and preserved
creation/publish, dirty/Cancel, timeout, legacy and protected-stage behavior.
All 14 paths are Required for behavior, verification or documentation. None
can be removed while retaining the Definition of Done.

### CA139-001 — Coordinator Acceptance / Pre-Commit STOP

Accepted exact subject c2a571ef148b2c9d3b1ea23bba11857f73d038cb with V139-002's
14 ordered rows, branch feature/task-139-listener-validation-ux, HEAD/base
a41e1569710b0245de584919cea2ed1d610d87ad. Owner amendment resolves R139-001.
Implementation, verification, PROCESS-002, scope and R139-002 are complete.
No architecture migration, Runtime activation or publication stage applies.
Pre-Commit Report: READY for a separate exact Owner Commit Gate. Commit and
publication remain unauthorized. The worktree and index are preserved without
staging. TASK-138 acceptance remains open. No successor is activated. The
Owner may resume TASK-138 manual testing against its dedicated service when
ready; review does not certify Owner acceptance. STOP.

### CG139-001 — Owner Commit Gate

Exact Owner commit command received for the accepted TASK-139 subject
c2a571ef148b2c9d3b1ea23bba11857f73d038cb. Working-tree manifest and exact
14-path scope match CA139-001. No generated, unexpected or staged content at
gate entry. Node 143/143 reconfirmed. Existing Go and browser evidence uses
the same accepted source bytes. EOL-aware diff check passed while preserving
the reviewed CRLF and LF bytes. Commit message follows repository history:
`fix(studio): validate listener host and port`.

Stage only these 14 paths and verify raw blobs, modes, projected manifest and
index tree before the one commit. Final commit identifiers are verified
after creation. Authorization covers this commit only.

# TASK-106 — Runtime idle-timeout enforcement

## Status

In Progress. Verification-stable projection; actual gates resolve from the newest
matching Recovery Evidence Envelope. No staging, commit or publication authorized.

## Task Contract

- Task Mode: Implementation, documentation-first, bounded existing IdleSeconds.
- Why Now: explicit Owner authorization; TASK-103/104 timeout work and TASK-105
  CI are published. IdleSeconds remains configured but inactive in Gate 9.
- Repository: universal-websocket-platform; branch
  `codex/task-106-runtime-idle-timeout`; trusted baseline/current HEAD
  `48a6c359ab5604add3c309248ba39dfaff7a0656`.
- Definition of Done: reconcile timeout/publication documentation drift; focused
  EN/RU design and independent Architecture Confirmation PASS before code;
  immutable Snapshot-to-Session propagation, idle validation/enforcement and
  concurrency/lifecycle proofs; independent Tester/non-author Reviewer;
  PROCESS-002, Scope Audit, exact manifest, Coordinator Acceptance and independent
  post-decision integrity. STOP AT COMMIT GATE.
- Out of Scope: Approved/Frozen amendments, new Configuration fields,
  dependencies, lifecycle states or terminal categories; TLS/native toolchain,
  Architecture B, PoC, TASK-102 changes, stage/commit/push/publication/settings.
- Stop conditions: baseline/ID conflict, Architecture FAIL, unresolved contract
  ambiguity, blocking review finding, unexpected scope expansion, missing required
  evidence or candidate ready for commit. No intermediate Owner gate within scope.
- Verification Plan: existing offline Go tooling (GOTOOLCHAIN=local, GOPROXY=off,
  GOSUMDB=off); focused tests, stress, real WebSocket smoke, full build/test/vet,
  tracked formatting, tidy/module guards, race if existing tooling supports it.
  Record unavailable checks explicitly; no acquisition/installation workaround.
  EN/RU parity, links, exact raw/projected manifest and diff integrity.
- Next candidate: separate Commit Gate after acceptance; Not Activated. Future
  publication requires separate Owner authorization and required go-ci SUCCESS
  on the exact PR head. TLS and native workstreams remain Not Activated.

## Sources, roles and ordered stages

Repository entry contracts, PROCESS-001/002, Coordinator/Architect/Developer/
Tester/Reviewer/Documentation roles; PROJECT_CONTEXT, spec README/current-state/
decisions, MASTER_PLAN EN/RU; ADR-0003, Frozen ARCH-002, Active ARCH-003,
Approved DP-003/004, implemented Snapshot boundary, TASK-103/104/105 and focused
read/write timeout notes. DP-008 is Draft, not independently normative authority.

Coordinator /root assigns itself Documentation Agent and later Developer as
nonoverlapping roles; separate Architect, independent Tester and non-author
Reviewer. Publisher N/A. Intake -> Documentation Baseline/reconciliation ->
focused design -> Architecture Confirmation -> implementation -> Verification ->
PROCESS-002 -> Scope Audit/Size Guard -> independent final Review -> Acceptance ->
independent post-decision integrity -> STOP. Architect never implements code.

## Existing Coverage Report

- Existing Coverage: Session text/binary/copy/serialized Send and Run/Stop;
  transactional pre-Commit handoff; Owner cleanup and Manager convergence;
  TASK-103 shared write budget and queue cancellation; TASK-104 full-message read
  budget, fragments/control frames, caller cancellation, Handler exclusion;
  Runtime Snapshot propagation, direct validation and real WebSocket bootstrap.
- Coverage Gap: idle policy disabled/bounds; idle-start/reset events and Handler
  interaction; activity/expiry/Stop races; successful versus partial/failed I/O;
  simultaneous read/write budgets; timer shutdown and Manager convergence.
- Added Proof Tests: 20 Session top-level idle tests plus 2 Runtime tests and
  Runtime scalar/Builder bounds table. Public behavior proofs cover idle start,
  reset events, Handler context/time, read/write interaction, late success,
  expiry/close contention, timer join, Commit wiring and Owner/Manager convergence.
  B-106-001 adds raw transport plus idle errors.Is identities and negative
  classification controls for parent, ReadSeconds and Stop cancellation.
- Added Regression Tests: full existing text/binary/echo, read/write, handshake,
  Owner and Manager suites PASS; existing Runtime wiring proofs adapted only for
  the explicit third duration. Final rework Session idle stress 20 tests x25 PASS.
- Remaining Limitations: fresh race probe exits 1 (-race requires cgo),
  CGO_ENABLED=0 and gcc/clang absent. No race-freedom claim. Native checkout tidy
  exits 1 for existing go.sum CRLF/LF only; exact Git-object source overlay tidy
  exits 0. Baseline hosted CI is not candidate/Windows native qualification.

## Documentation baseline and applicability

Required: task/index/current-state/PROJECT_CONTEXT, focused design EN/RU,
MASTER_PLAN EN/RU Gate 9 partial status and user-facing CHANGELOG. Before code,
correct current-state's stale absent read/write timeout claim and reconcile
published TASK-105 navigation/state facts against Git/GitHub evidence. Historical
TASK-103/104/105 records remain unchanged. ADR/ARCH/Approved DP amendments forbidden;
other indexes/README/spec decisions assessed for applicability, no automatic edits.
Size Guard reassessment required at >15 paths or >500 production lines; retain
one behavior, no unrelated formatting/refactoring or new package/dependency.

## Initial implementation and pre-rework verification handoff

This section preserves the initial cycle at pre-rework subject9a7999ecd6f37ed73fa881feed961bf81d5f2cfb;
final B-106-001 rework evidence is separately recorded below and in the envelope.

Independent Architect106 confirmed exact EN/RU notes before any product edits;
E-106-003 records raw pre-code note identities. Targeted actual-code conformance
PASS / no new prerequisite, blocking 0, for five production raw OIDs below.
Only later status wording changed Planned -> implemented candidate; design
semantics remained unchanged. Read notes received additive Handler idle-scope
qualification; existing Approved/Frozen sources and historical task records
have no Git delta and were not edited. TASK-103 checkout retains existing CRLF
while its Git blob has LF; no raw checkout/object byte equivalence is claimed.

DefaultSession uses one admitted-Run idleTracker and private cancellation-cause
child; only complete text/binary Read and successful Write reset active idle.
Late success cannot revive expiry/close. ReadSeconds remains a separate lexical
budget; Send retains its caller/write budget. Stop and Run idempotently join the
same watchdog outside lifecycle locks before transport shutdown/readLoopDone.
No idle-owned work survives Manager successful closure. Cooperative limits remain.

Commands used existing Go1.26.5 windows/amd64, GOTOOLCHAIN=local/GOPROXY=off/
GOSUMDB=off; no downloads or toolchain changes:
- Independent Tester final `go test -mod=readonly -count=25 -timeout=120s
  -run '^TestIdleTimeout' ./internal/session`: exit0,24.924s; 18 top-level tests,
  450 executions plus subtests. Focused Session/Runtime exit0,3.212s/3.085s.
- Root `go test -mod=readonly -count=1 -timeout=120s ./...`: exit0, all packages;
  Session5.068s, Runtime7.182s, existing Windows containment24.071s. Actual
  production text/binary echo then 1s idle expiry preserves root/readiness and
  Host Stop joins Manager/Owner. Fragment/control proof emits non-final wire
  data (installed websocket v1.8.15 writer source independently inspected).
- `go build -mod=readonly ./...` and `go vet -mod=readonly ./...`: exit0.
  Build emitted a denied optional module stat-cache write in sandbox; build
  completed successfully. No repository module mutation occurred.
- Changed Go formatting: 10 files gofmt clean. Exact source fixture uses
  `git -c core.autocrlf=false archive HEAD`, verifies each extracted source blob
  against baseline ls-tree SHA1, overlays current changed/new source bytes,
  then `gofmt -l cmd internal`:269 files clean; `go mod tidy -diff`:exit0.
  This fixture does not claim raw checkout tidy PASS or content normalization.
- Native checkout `go mod tidy -diff`:exit1, only 24 go.sum CRLF-to-LF lines;
  normalized module/hash strings identical to baseline; no module file edits.
  `git diff --exit-code -- go.mod go.sum`:exit0.
- `go test -race -run '^$' ./internal/session`:exit1, requires cgo;
  CGO_ENABLED=0 and no gcc/clang. Race NOT RUN, stress is not race freedom.
- EN/RU semantic parity and 223 local Markdown links:PASS,missing0.
  `git -c core.whitespace=cr-at-eol diff --check`:exit0, preserving existing
  mixed documentation endings. No global Git config changes.

Initial sandbox tests failed loopback connectex access restrictions, not accepted
as proof; supported escalated execution supplied decisive socket/subprocess
results. Two initial test assertions assumed adjacent Windows time.Now calls
advance; Tester inserted elapsed time before measuring activity, then reran
stable focused/stress checks. No product change or semantic workaround followed.
Temporary verification fixtures remain outside repository and are not candidate
artifacts. Initial fixture tooling failures (restricted .ai directory, Windows
command-length and archive autocrlf) were reconciled before successful final
raw-object-verified fixture; no repository normalization or hidden tidy repair.

## B-106-001 rework and final verification handoff

Explicit Owner rework admission reconstructed the exact pre-rework21-path
manifest9a7999ecd6f37ed73fa881feed961bf81d5f2cfb. Architect confirmed the
minimal correction uses the existing focused contract, without amendment.
Historical Review/STOP evidence is retained, not treated as current Acceptance.

Root cause: sessionReadError prioritizes cancelled context and had discarded
an independent raw Read error before Run joined the idle DeadlineExceeded cause.
Fix: pass the exact Run-local tracker to private readApplicationMessage; retain
existing classification, read the irreversible expired latch under tracker mutex,
and join the wrapped raw Read error only when that exact idle scope expired.
No joins on success; disabled/parent-only/read-only/Stop-only classifications,
Send budgets, timer ownership, cleanup and Manager accounting are unchanged.

Rework changes only session.go, existing idle_timeout_test.go and this task
record versus pre-rework candidate. The original18-test prefix remains byte-
identical; two appended top-level tests bring total20 and add six subcases:
both error identities with original parent live under zero/hour read budgets,
and disabled-parent/enabled-parent/ReadSeconds/Stop negative controls.

Fresh offline verification, existing Go1.26.5 windows/amd64, GOTOOLCHAIN=local,
GOPROXY=off/GOSUMDB=off, no downloads/installations:
- Deterministic `go test -mod=readonly -count=1 -timeout=30s -run
  '^TestIdleTimeoutReadPreserves' ./internal/session`:exit0,2.234s.
- Original external Reviewer overlay reproducer, unchanged command/source from
  E-106-006:exit0,2.441s, now preserves both errors.Is identities.
- Independent Tester `go test -mod=readonly -count=1 -timeout=90s -run
  'TestIdleTimeout|TestReadTimeout|TestWriteTimeout|TestRuntimeIdleTimeout|TestBootstrapIdleTimeout'
  ./internal/session ./internal/runtime`:exit0,3.355s/3.288s; includes real
  partial/control proof and production text/binary echo/idle/Host survival smoke.
- Independent Tester `go test -mod=readonly -count=25 -timeout=120s -run
  '^TestIdleTimeout' ./internal/session`:exit0,20.299s,500 top-level executions
  plus subtests. Tester PASS WITH LIMITATION, findings0/0, no rework edits.
- Root full `go build -mod=readonly ./...`, `go test -mod=readonly -count=1
  -timeout=120s ./...`, `go vet -mod=readonly ./...`:exit0. All packages PASS;
  Session4.674s, Runtime11.869s, Windows containment54.259s. Supported escalated
  socket/subprocess checks, no host/toolchain mutation.
- Ten changed Go files gofmt clean; exact baseline-object-verified raw source
  overlay269 Go files formatting clean and `go mod tidy -diff` exit0.
  Module guard and CR-aware diff check exit0; go.mod/go.sum raw identities
  unchanged (aa82a60c0769256f13c86b31382a8994cf6e73e4 /
  f1912accf367972c7c6278b7a8d7de478028139b). No dependency edits.
- Fresh race probe rejected with '-race requires cgo'; CGO/compiler situation
  unchanged, race NOT RUN. Native checkout tidy limitation remains existing
  go.sum CRLF-only, not a raw tidy PASS. 223 local links valid, EN/RU unchanged.

Full source identity is frozen only after this handoff; new rows/manifest below
replace the pre-rework candidate. Independent final Review, Coordinator decision
and post-decision integrity must use the new subject. Old Review/Acceptance
cannot transfer. Until an Approved repeat Review there is no commit readiness.

## PROCESS-002 and Scope Audit

Synchronized: task/index/current-state/PROJECT_CONTEXT, focused EN/RU idle design,
additive read-note qualification, MASTER_PLAN EN/RU partial Gate9 capability,
CHANGELOG including now-effective default60s. Published TASK-103/104/105 durable
facts reconciled; historical records preserved. Gate9 still incomplete (TLS).
Root README/spec README/decisions/general indexes:Not applicable, no new product
positioning, subsystem, normative decision or public document family. Focused
notes discoverable through current-state/task and mirror links. ADR/ARCH/Approved
DP: no changes permitted/needed; Draft DP-008 is correctly treated as Draft.

Scope Audit Required21 / Questionable0 / Removable0. Mandatory removal question:
- five production paths: removing any loses idle enforcement, pre-resource
  validation, immutable wiring or joined timer ownership;
- five test paths: new Session/Runtime proofs, bounds/support matrix and two
  existing wiring signatures are necessary for verified production composition;
- eleven documentation paths: task recovery/coverage/evidence, three navigation/
  state facts, four EN/RU design notes, two roadmap mirrors and user-facing
  CHANGELOG are necessary for DoD and PROCESS-002.
Size Guard ACCEPT after >15-path reassessment: one behavior, five production
files under500 added lines, no new package/dependency/architecture contract.
Splitting propagation, timer ownership or required documentation leaves this DoD
incomplete. No formatting-only, generated or unrelated changes retained.

## Candidate identity and byte preservation

Candidate manifest is staging-invariant: all changed/new paths relative to HEAD,
unsigned UTF-8 path sort, raw full bytes; task record uses task-record-v1 (raw
through Status line terminator + STATUS-EVIDENCE-EXCLUDED NUL + raw Task Contract
through before terminal envelope). Rows are path NUL projection NUL state NUL
mode NUL oid NUL; hash with git hash-object --stdin, no writes/staging. All paths
present mode100644; object format sha1. Ordered rows and OID are appended below.
After identity freeze, only append-only envelope entries are permitted; verdicts
do not attest their own bytes. Existing mixed CRLF docs are not normalized.
Future separately authorized staging must preserve reviewed raw bytes using
per-command core.autocrlf=false and verify exact staged blobs; no permission now.

## Recovery Evidence Envelope

### E-106-001 — Intake and preflight

Owner authorized the bounded engineering cycle. Clean local main/HEAD/cached
origin/main and live remote main all equal the trusted baseline. TASK-106 absent
from records/hidden state, tracked paths, available all-ref history/local refs,
live remote heads and GitHub issue/PR search (0). No other worktree has that task.
TASK-102 ref remains `8bfa2e22a4b6310a71f6d36c8724b1adedd727ad`, outside main ancestry.
First branch creation was refused by sandbox with no ref/index/worktree effect;
read-only reconstruction confirmed no ref or lock. Supported escalated retry
created the exact authorized branch. This record is the first content change.
No product/test changes, Architecture verdict, acceptance or staged paths yet.
First incomplete gate: documentation reconciliation and Architecture Confirmation.

### E-106-002 — Recovery and documentation baseline

Current explicit Owner continuation restored the same branch and exact baseline;
index empty, only attributed six documentation paths changed/untracked. No
production/test changes or unknown side effects. E-106-001/Task Contract/Coverage
Report preserved. Navigation/state reconciliation is localized; historical task
records untouched. Architect106 read-only assessment found compatible design,
no Approved/Frozen amendment or new Owner decision; exact EN/RU confirmation
was explicitly pending, not PASS. Notes amended to match the handoff: successful
Read/expiry race, joined error preservation, Stop joins same watchdog outside
locks, additive read-note Handler-context qualification. First incomplete gate:
independent exact-note Architecture Confirmation before product implementation.

### E-106-003 - Architecture Confirmation PASS before code

Independent Architect /root/architect106: PASS / APPROVED, blocking findings 0;
no unresolved ambiguity, Owner decision or Approved/Frozen amendment required.
Exact raw notes independently read back: EN idle eaf39be25c8194a7aed39d93913f6c991de17683;
RU idle 3875b489cad671b67523658bf4ad82392de0de7d; EN read
f4962fa1dfce6beebbc8cb329b5753cc00e21d11; RU read
c0e00a08e93326a792c6383724541af9eb7a13f2. EN/RU semantic parity confirmed.
Private Run child/watchdog, irreversible expiry, original Send budgets and
watchdog stop/join preserve ARCH-002 and DP-003/004 ownership/first-writer/lease
boundaries. All Owner architecture points resolved. No product/test changes
preceded this confirmation. Implementation admitted within exact focused notes.
First incomplete gate: implementation and fresh verification.

### E-106-004 - Frozen candidate and verification subject

Anchor HEAD/base 48a6c359ab5604add3c309248ba39dfaff7a0656; branch codex/task-106-runtime-idle-timeout.
Canonical manifest `5739fbc2cd4cced432055ef1140e1877b44fd654`, object format sha1; 21 ordered
present mode100644 rows, raw/task-record-v1 algorithm in projected handoff.
Independent Tester /root/tester106 PASS WITH LIMITATION, findings0/0;
all five production OIDs and both new test OIDs independently recorded,
matching rows below. Commands/results/450 stress executions and race/tidy
limitations are in the projected handoff. Full build/test/vet and source
fixture guards complete. PROCESS-002 Synchronized; Scope Audit21/0/0,
Size Guard ACCEPT. Index empty; no changes to baseline/main/TASK-102.
No acceptance yet; first incomplete gate: independent final Review.

| Path | Projection | State | Mode | OID |
|---|---|---|---|---|
| .ai/PROJECT_CONTEXT.md | full | present | 100644 | 04cead25a98000acb321526af5454d6b5da5a8dd |
| CHANGELOG.md | full | present | 100644 | 4c805f25560b9e1a44adbdf900471d92c0c51016 |
| docs/en/design/runtime-idle-timeout.md | full | present | 100644 | a43632302fd62654551240e98289545967123033 |
| docs/en/design/runtime-read-timeout.md | full | present | 100644 | f4962fa1dfce6beebbc8cb329b5753cc00e21d11 |
| docs/en/roadmap/MASTER_PLAN.md | full | present | 100644 | e80753bdec52138ff64a3513ee5d236d8aeefe49 |
| docs/ru/design/runtime-idle-timeout.md | full | present | 100644 | 7c391097ffdd00d814c958e68f65bc44e693a979 |
| docs/ru/design/runtime-read-timeout.md | full | present | 100644 | c0e00a08e93326a792c6383724541af9eb7a13f2 |
| docs/ru/roadmap/MASTER_PLAN.md | full | present | 100644 | 0d6a48254537a7e1b8b02219f77f41a8da466c6b |
| docs/tasks/README.md | full | present | 100644 | dc96d44ad1621c85971d493748d17304eb4576a1 |
| docs/tasks/TASK-106-RUNTIME-IDLE-TIMEOUT.md | task-record-v1 | present | 100644 | bb0f33253a344ca6d277991dd7109e19173ed148 |
| internal/runtime/composition.go | full | present | 100644 | e8547eb8a88eeb219f8b7949673b9fa1e54ff166 |
| internal/runtime/configuration_validation.go | full | present | 100644 | b5d45090587e82d0c360597e583220971cd02887 |
| internal/runtime/configuration_validation_test.go | full | present | 100644 | 9b602bb3ba111a761cf9562d576a8934dea27ba1 |
| internal/runtime/idle_timeout_test.go | full | present | 100644 | b1e59c52e06cb0bea9d94136622c69f0cd342d08 |
| internal/runtime/read_timeout_test.go | full | present | 100644 | a7dd64b087be63a60fd9d57168e5e1fb3acb7cec |
| internal/runtime/write_timeout_test.go | full | present | 100644 | fbd82dccd795fdae79a90dd96914396036f31ce2 |
| internal/session/idle_timeout.go | full | present | 100644 | 7c90100c639f8e985f21d4f31b7fee4132927c44 |
| internal/session/idle_timeout_test.go | full | present | 100644 | 75b94a01198b841d9bf609e00b69b95eb127cb32 |
| internal/session/session.go | full | present | 100644 | 74372be22dac5739a16a542aa03486636cfa6443 |
| internal/session/transactional_dispatcher.go | full | present | 100644 | 3d54cd65c9861db5169da24b64740a839ac0b59e |
| spec/current-state.md | full | present | 100644 | 5502163b7bf390471ee5a80de5a7d011c897a753 |

### E-106-005 - Replacement frozen candidate, documentation precision

E-106-004 is superseded for candidate identity only. A projected handoff
sentence now distinguishes unchanged historical Git paths from pre-existing
checkout CRLF. No product/test/design semantic changes; no final Review or
Acceptance existed. Reconstruct fresh identity; prior test OIDs unchanged.

Anchor HEAD/base 48a6c359ab5604add3c309248ba39dfaff7a0656; branch codex/task-106-runtime-idle-timeout.
Canonical manifest `9a7999ecd6f37ed73fa881feed961bf81d5f2cfb`, object format sha1; 21 ordered
present mode100644 rows, raw/task-record-v1 algorithm in projected handoff.
Independent Tester /root/tester106 PASS WITH LIMITATION, findings0/0;
all five production OIDs and both new test OIDs independently recorded,
matching rows below. Commands/results/450 stress executions and race/tidy
limitations are in the projected handoff. Full build/test/vet and source
fixture guards complete. PROCESS-002 Synchronized; Scope Audit21/0/0,
Size Guard ACCEPT. Index empty; no changes to baseline/main/TASK-102.
No acceptance yet; first incomplete gate: independent final Review.

| Path | Projection | State | Mode | OID |
|---|---|---|---|---|
| .ai/PROJECT_CONTEXT.md | full | present | 100644 | 04cead25a98000acb321526af5454d6b5da5a8dd |
| CHANGELOG.md | full | present | 100644 | 4c805f25560b9e1a44adbdf900471d92c0c51016 |
| docs/en/design/runtime-idle-timeout.md | full | present | 100644 | a43632302fd62654551240e98289545967123033 |
| docs/en/design/runtime-read-timeout.md | full | present | 100644 | f4962fa1dfce6beebbc8cb329b5753cc00e21d11 |
| docs/en/roadmap/MASTER_PLAN.md | full | present | 100644 | e80753bdec52138ff64a3513ee5d236d8aeefe49 |
| docs/ru/design/runtime-idle-timeout.md | full | present | 100644 | 7c391097ffdd00d814c958e68f65bc44e693a979 |
| docs/ru/design/runtime-read-timeout.md | full | present | 100644 | c0e00a08e93326a792c6383724541af9eb7a13f2 |
| docs/ru/roadmap/MASTER_PLAN.md | full | present | 100644 | 0d6a48254537a7e1b8b02219f77f41a8da466c6b |
| docs/tasks/README.md | full | present | 100644 | dc96d44ad1621c85971d493748d17304eb4576a1 |
| docs/tasks/TASK-106-RUNTIME-IDLE-TIMEOUT.md | task-record-v1 | present | 100644 | 31a3aaf7bc45ab2d58eeb11410d1e37ac72d8879 |
| internal/runtime/composition.go | full | present | 100644 | e8547eb8a88eeb219f8b7949673b9fa1e54ff166 |
| internal/runtime/configuration_validation.go | full | present | 100644 | b5d45090587e82d0c360597e583220971cd02887 |
| internal/runtime/configuration_validation_test.go | full | present | 100644 | 9b602bb3ba111a761cf9562d576a8934dea27ba1 |
| internal/runtime/idle_timeout_test.go | full | present | 100644 | b1e59c52e06cb0bea9d94136622c69f0cd342d08 |
| internal/runtime/read_timeout_test.go | full | present | 100644 | a7dd64b087be63a60fd9d57168e5e1fb3acb7cec |
| internal/runtime/write_timeout_test.go | full | present | 100644 | fbd82dccd795fdae79a90dd96914396036f31ce2 |
| internal/session/idle_timeout.go | full | present | 100644 | 7c90100c639f8e985f21d4f31b7fee4132927c44 |
| internal/session/idle_timeout_test.go | full | present | 100644 | 75b94a01198b841d9bf609e00b69b95eb127cb32 |
| internal/session/session.go | full | present | 100644 | 74372be22dac5739a16a542aa03486636cfa6443 |
| internal/session/transactional_dispatcher.go | full | present | 100644 | 3d54cd65c9861db5169da24b64740a839ac0b59e |
| spec/current-state.md | full | present | 100644 | 5502163b7bf390471ee5a80de5a7d011c897a753 |

### E-106-006 - Independent Review R-106-001, mandatory STOP

Independent non-author Reviewer /root/reviewer106: Needs Revision, blocking1,
nonblocking0. Independently reconstructed with separate SHA1/NUL implementation:
manifest `9a7999ecd6f37ed73fa881feed961bf81d5f2cfb`, task projection
`31a3aaf7bc45ab2d58eeb11410d1e37ac72d8879`; all21 ordered rows E-106-005 match.
Independent Tester also confirmed this replacement identity and unchanged20
other rows; prior positive tests remain bounded coverage, not final Acceptance.
Scope Audit21 Required/0 Questionable/0 Removable and Size Guard justified.

B-106-001 BLOCKING: Session readApplicationMessage passes raw Read failure to
sessionReadError (session.go:304); with private idle child cancelled but original
parent live, sessionReadError returns contextErr (session.go:408), dropping an
independent transport sentinel. idle_timeout.go:121 only adds DeadlineExceeded
and cannot recover the lost cause. Confirmed focused design requires both
transport cause and idle cause to survive errors.Is. Architect106 explicitly
confirmed this requirement; its earlier targeted conformance PASS missed this
detail and does not prove this obligation. No contract reinterpretation or
new Owner decision is needed for the requirement itself.

Reviewer deterministic external Go overlay regression (no repository changes):
`go test -mod=readonly -overlay=C:/Users/dsdred/.codex/visualizations/2026/10/21/01a15fae-9d66-7f40-abae-087229e4c274/review106-overlay.json -count=1 -timeout=20s -run '^TestReviewerIdleReadPreservesTransportSentinel$' ./internal/session`
exit1; test0.02s/package2.345s. Output: transport cause missing after idle
cancellation: context canceled / context deadline exceeded. Proof returns an
independent Read sentinel after private child Done, keeping original parent live;
expected errors.Is(sentinel) and errors.Is(DeadlineExceeded) are both required.

Coordinator disposition: STOP under explicit Owner mandatory blocking-review
condition. TASK-106 is not Accepted, not Completed and not ready for commit.
No implementation/test rework, stage, commit, push or publication after finding.
Existing candidate/design/coverage/evidence preserved; projected In Progress
resolves to this newest matching STOP envelope. Task remains on exact branch
and baseline, index empty, TASK-102 unchanged/unpublished; B/PoC not activated.

Remaining engineering work requires separate explicit Owner rework authorization:
correct B-106-001 within confirmed scope, add deterministic missing regression,
repeat affected verification and fresh exact-subject independent Review, then
Coordinator Acceptance/post-decision integrity if approved. Prior passing tests,
Scope Audit and Architecture design PASS do not bypass this failed gate.
Append-only stop evidence does not change candidate manifest/projection and does
not self-attest its final raw bytes. First incomplete gate: blocking Review
resolution (not Commit Gate). No commit or publication permission requested.

### E-106-007 - Independent STOP evidence integrity

Reviewer /root/reviewer106 read back E-106-006: PASS for STOP evidence integrity
only. Needs Revision,1 blocking/0 nonblocking, B-106-001 and no Acceptance or
commit-readiness claim are accurate. Independently recomputed manifest
9a7999ecd6f37ed73fa881feed961bf81d5f2cfb, task projection
31a3aaf7bc45ab2d58eeb11410d1e37ac72d8879,21 ordered rows, index empty.
No tests, rework or repository edits by Reviewer. This append is excluded
evidence only; it does not self-attest final raw envelope bytes. STOP persists.

### E-106-008 - Explicit B-106-001 rework intake

Current Owner explicitly authorized bounded rework of B-106-001 only, on the
same branch/base and pre-rework manifest9a7999ecd6f37ed73fa881feed961bf81d5f2cfb.
Read-only reconstruction reproduced exact21 rows/task projection31a3aaf7bc45ab2d58eeb11410d1e37ac72d8879,
empty index, baseline48a6c359ab5604add3c309248ba39dfaff7a0656 and unchanged
TASK-102 ref8bfa2e22a4b6310a71f6d36c8724b1adedd727ad. B-106-001 evidence,
external deterministic overlay and Architect requirement checked. Old manifest
is pre-rework evidence only, not final candidate or Acceptance. Prior historical
STOP/Review records preserved; this authorized cycle starts with implementation
correction, then deterministic regression, affected and full checks, independent
Review, PROCESS-002/Scope Audit, Acceptance/integrity only if approved.

Existing Coverage Report before test mutation: prior18 Session idle tests and
2 Runtime tests cover lifecycle/reset/wiring but not raw transport sentinel
during private idle expiry. Coverage Gap: both errors.Is identities plus
negative controls preserving disabled-idle, parent cancellation and ReadSeconds
classification. Added proofs planned in existing Session idle test file only.
No new files/packages/dependencies/contracts or ownership changes required.
Architect read-only scope confirmation pending; no product/test edits yet.
Regression failure, non-Approved Review, new blocking finding, contract amendment
or scope expansion requires STOP. No stage/commit/publication authorization.

### E-106-009 - B-106-001 implementation handoff

Independent Architect106 confirmed proposed boundary conforms to unchanged
Architecture contract, no amendment/new prerequisite before rework. Source edit
passes exact Run-local idle tracker to private Read helper and retains existing
sessionReadError classification; only latched idle expiry joins wrapped raw Read
error. Successful Reads, disabled idle, parent-only, ReadSeconds-only and Stop-only
classification unchanged. No ownership/lifecycle/accounting edits.
Changed production raw OID session.go54e6c0eddbdb94dfd84bb1f08bb8af7e64c47d4b.
Existing idle test file adds2 top-level tests (20 total) with live-parent sentinel
plus DeadlineExceeded under zero/hour read budgets and4 negative controls; raw
OID68daa3184fefb169b82921413b45d55730217f1f. All other product/test files unchanged.
Fresh deterministic regression pending; independent Tester verifying. No PASS,
final candidate or Acceptance claimed yet. Old identities remain historical.

### E-106-010 - Rework frozen candidate and verification subject

New canonical manifest `28a408a9e246ebd602310c4f8f5381bf4a21032f`, task-record-v1
projection `f2e0a7d3433c75da08222f8b708bf7e922174377`, sha1,21 rows.
Same branch/base48a6c359ab5604add3c309248ba39dfaff7a0656, index empty.
Pre-rework9a7999ecd6f37ed73fa881feed961bf81d5f2cfb remains historical only.
Exactly3 subject rows differ: session.go, existing idle_timeout_test.go, task
projection; other18 rows unchanged. Fresh deterministic and original external
reproducer PASS, full build/test/vet PASS, independent Tester PASS WITH LIMITATION
with500 stress executions and Runtime smoke. Existing race/checkout tidy limits
remain explicit. Projected rework handoff contains commands/results and scope.
Independent repeat Review pending; no Acceptance or commit readiness yet.

| Path | Projection | State | Mode | OID |
|---|---|---|---|---|
| .ai/PROJECT_CONTEXT.md | full | present | 100644 | 04cead25a98000acb321526af5454d6b5da5a8dd |
| CHANGELOG.md | full | present | 100644 | 4c805f25560b9e1a44adbdf900471d92c0c51016 |
| docs/en/design/runtime-idle-timeout.md | full | present | 100644 | a43632302fd62654551240e98289545967123033 |
| docs/en/design/runtime-read-timeout.md | full | present | 100644 | f4962fa1dfce6beebbc8cb329b5753cc00e21d11 |
| docs/en/roadmap/MASTER_PLAN.md | full | present | 100644 | e80753bdec52138ff64a3513ee5d236d8aeefe49 |
| docs/ru/design/runtime-idle-timeout.md | full | present | 100644 | 7c391097ffdd00d814c958e68f65bc44e693a979 |
| docs/ru/design/runtime-read-timeout.md | full | present | 100644 | c0e00a08e93326a792c6383724541af9eb7a13f2 |
| docs/ru/roadmap/MASTER_PLAN.md | full | present | 100644 | 0d6a48254537a7e1b8b02219f77f41a8da466c6b |
| docs/tasks/README.md | full | present | 100644 | dc96d44ad1621c85971d493748d17304eb4576a1 |
| docs/tasks/TASK-106-RUNTIME-IDLE-TIMEOUT.md | task-record-v1 | present | 100644 | f2e0a7d3433c75da08222f8b708bf7e922174377 |
| internal/runtime/composition.go | full | present | 100644 | e8547eb8a88eeb219f8b7949673b9fa1e54ff166 |
| internal/runtime/configuration_validation.go | full | present | 100644 | b5d45090587e82d0c360597e583220971cd02887 |
| internal/runtime/configuration_validation_test.go | full | present | 100644 | 9b602bb3ba111a761cf9562d576a8934dea27ba1 |
| internal/runtime/idle_timeout_test.go | full | present | 100644 | b1e59c52e06cb0bea9d94136622c69f0cd342d08 |
| internal/runtime/read_timeout_test.go | full | present | 100644 | a7dd64b087be63a60fd9d57168e5e1fb3acb7cec |
| internal/runtime/write_timeout_test.go | full | present | 100644 | fbd82dccd795fdae79a90dd96914396036f31ce2 |
| internal/session/idle_timeout.go | full | present | 100644 | 7c90100c639f8e985f21d4f31b7fee4132927c44 |
| internal/session/idle_timeout_test.go | full | present | 100644 | 68daa3184fefb169b82921413b45d55730217f1f |
| internal/session/session.go | full | present | 100644 | 54e6c0eddbdb94dfd84bb1f08bb8af7e64c47d4b |
| internal/session/transactional_dispatcher.go | full | present | 100644 | 3d54cd65c9861db5169da24b64740a839ac0b59e |
| spec/current-state.md | full | present | 100644 | 5502163b7bf390471ee5a80de5a7d011c897a753 |

### E-106-011 - Exact-subject Tester and Independent Repeat Review

Independent Tester /root/tester106 reproduced manifest
28a408a9e246ebd602310c4f8f5381bf4a21032f, task projection
f2e0a7d3433c75da08222f8b708bf7e922174377, all21 E-106-010 rows. Previously
tested source/test raw identities unchanged;500 stress executions, affected
suites and Runtime smoke apply. PASS WITH LIMITATION, findings0/0; no rerun needed.

Independent non-author Reviewer /root/reviewer106: R-106-002 Approved,
blocking0/nonblocking0; B-106-001 resolved. Independently reconstructed same
manifest/projection/all21rows and verified exactly3 rework rows versus pre9a.
Original18-test prefix byte-identical,2 tests appended. Exact Run-local latch
conditional join preserves raw cause plus idle identity without unrelated
classification/Send/watchdog/ownership/accounting changes. ScopeAudit21/0/0,
SizeGuard justified, PROCESS-002/EN-RU adequate. No edits by Tester/Reviewer.

### E-106-012 - Final PROCESS-002, Scope Audit and integrity before decision

After Approved R-106-002 Coordinator/Documentation Agent completed PROCESS-002
applicability: task coverage/rework/verification handoff reflects current source;
index/current-state/PROJECT_CONTEXT retain In Progress with matching-envelope
resolution and durable TASK-103/104/105 publication facts. Focused EN/RU notes
already require preserved error causes, so no contract/status amendment needed.
MASTER_PLAN mirrors keep Gate9 partial and TLS open; CHANGELOG already documents
idle behavior/default. Historical task/evidence records preserved. Root README,
spec README/decisions/general indexes remain N/A for this bounded error fix.
Rechecked223 local Markdown links,missing0; EN/RU bytes unchanged from reviewed
candidate with confirmed semantic parity. No further projected/full mutations.

Coordinator Scope Audit agrees Required21/Questionable0/Removable0, per-path
removal rationale in projected handoff and R-106-002. Rework delta has only2
product/test paths plus task record, no new independent behavior/package,
dependency, state or contract. SizeGuard ACCEPT; total production under500 lines.
Root fresh reconstruction matched all21 reviewed rows, manifest/projection,
empty index, unchanged HEAD/main/TASK-102 and no unexpected candidate paths.

### E-106-013 - Coordinator Acceptance A-106-001, Commit Gate STOP

Coordinator ACCEPT TASK-106 authoring candidate after bounded B-106-001 rework.
Decision A-106-001 binds exact manifest
28a408a9e246ebd602310c4f8f5381bf4a21032f, task-record-v1
f2e0a7d3433c75da08222f8b708bf7e922174377, all21 ordered E-106-010 rows,
branch codex/task-106-runtime-idle-timeout, baseline/current HEAD
48a6c359ab5604add3c309248ba39dfaff7a0656. Pre-rework9a7999... remains historical
and is not final candidate or Acceptance. Original NeedsRevision/STOP evidence
was not rewritten or reused as approval.

DoD: confirmed Architecture unchanged, immutable propagation and bounds,
idle activity/expiry/Handler/read-write interactions, concurrent cancellation,
joined watchdog/Owner cleanup/Manager convergence and echo regressions have
recorded proof. Deterministic B-106-001 regression and original Reviewer overlay
now PASS, both errors.Is identities retained with live parent; negative controls
PASS. Fresh focused/stress500/full build-test-vet/format/module checks complete.
Independent R-106-002 Approved0/0; Tester PASS WITH LIMITATION0/0. PROCESS-002
Synchronized, ScopeAudit21/0/0 and SizeGuard ACCEPT. No pending blocking findings.

Accepted limitations: race NOT RUN (CGO/compiler unavailable); native checkout
tidy existing go.sum CRLF-only is not PASS, separately exact-object source
fixture formatting/tidy PASS. Cooperative Handler/transport limits remain.
No TLS/native/B/PoC/TASK-102 changes or new architecture/dependencies/categories.

Candidate ready for separately authorized Commit Gate after independent
post-decision integrity. No stage, commit, push, publication or settings changes.
Future publication requires separate Owner gate and successful required go-ci
on the exact PR head; baseline CI supplies no fresh target authority. This
excluded evidence append does not self-attest its final raw bytes. First remaining
checkpoint: independent post-decision integrity, then STOP AT COMMIT GATE.

### E-106-014 - Independent post-decision integrity PASS

Independent Reviewer /root/reviewer106 read back E-106-011/012/013 after
Coordinator Acceptance A-106-001: PASS, blocking0/nonblocking0. Reconstructed
manifest28a408a9e246ebd602310c4f8f5381bf4a21032f, task projection
f2e0a7d3433c75da08222f8b708bf7e922174377 and all21 E-106-010 rows unchanged.
No post-review/post-acceptance projected/full mutations. Branch unchanged,
HEAD/main48a6c359ab5604add3c309248ba39dfaff7a0656, index empty, TASK-102 unchanged.
Acceptance accurately follows Approved R-106-0020/0; B-106-001 resolved and race,
checkout-tidy/cooperative limitations retained. Historical STOP separate.
Reviewer performed no tests/edits/staging/commit/publication in this integrity
step. This append is excluded metadata, without self-attesting final raw bytes.
Final disposition: Coordinator Accepted, candidate ready for separately
authorized Commit Gate. STOP AT COMMIT GATE. No stage/commit/publication.

# TASK-120 — Isolated C2-Pull

## Status

Verification-stable `In Progress`. Isolated source/test implementation prepared;
Acceptance not implied. Actual gates/closure resolve only from newest valid matching
Recovery Evidence Envelope after independent canonical manifest reconstruction.
Production C2/C3 Not Activated; Gate 11 OPEN. No Commit/Publication permission.

## Task Contract

- Task Mode: Implementation. Owner separately authorized isolated implementation
  2026-11-05, then minimal task/project-state/EN-RU documentation synchronization.
- Repository: E:/wikiPRJ/universal-websocket-platform; branch
  `feature/task-120-runtime-diagnostics-pull`; trusted clean baseline
  `ac021e9402f37d60ab826f100534c88415c08a21`.
- Why Now: TASK-119 published intake; exact DP-027 §29/§32–34 prerequisite is Ready,
  Architecture Confirmation PASS accepted by Owner, current execution admission.
- DoD: exact New(Reader)/View.Poll/fixed64/shared outer TryLock/one C1 Drain/detached
  batch/guarded clearing; O01–O16 with explicit automated/compositional/source-audit
  evidence; source conformance; focused/full build/test/vet/race/stress/regression,
  tidy/module guards with actual platform limitations; PROCESS-002, Scope Audit,
  independent Verification/Review, Coordinator Acceptance and final-byte integrity.
- Product scope: only new `internal/runtimediagnosticspull/pull.go` and
  `pull_test.go`. Existing production/tests/modules unchanged.
- Non-goals: Host/Bootstrap/Control Service wiring, Dispatcher/Handler/shutdown
  sources, production exposure, workers/timers, retry/reinsert/durable/lossless
  guarantees, C3, TASK-102, Gate11 closure, Approved/Frozen changes and git
  stage/intent-to-add/commit/push/PR/merge/publication.
- Verification Plan: O01–O08/O10–O12 automated plus source audit; O09/O13 explicitly
  compositional unchanged mailbox/C1/Host/component evidence; O14–O16 API/schema/
  imports/godoc audit and executable checks. Commands/results below in envelope.

## Selection, baseline and architecture

MASTER_PLAN §5/§8 EN/RU, task index, TASK-116/119 and DP-027 §29/32–34 choose
exact isolated C2-Pull after published C1. Production C2 rejected: sanctioned
owner/per-Host allocation-retirement/consume authorization/scheduling/deployment/
retention missing. C3 requires separate closed schema/source decisions and admission.
Metrics/Plugin contracts do not displace explicit Owner Runtime Diagnostics priority.
No successor activated; roadmap is dependency ordering, not task queue.

Fresh preceding intake proved main/origin/live remote at baseline, clean worktree;
TASK119 PR119 MERGED, source `4623505b62ab126ebf2fe029de754f42faee1bb4`, source/merge
tree `3e2cbd14e653e617af538af79d24b770fddd3dad`. Original native terminal P10 and
final P10R PASS read back from thread `01a1a7cb-d082-7b10-b467-023072bb9a3f`.
P10 payload `msg_0010b23aae93f079016aeb81bd8d7c87d2851c3be2fbf7eabe` independently
matched original backing SHA256
`06fa7e483abcbcee261a26fb71cec553dd8ff1429793d084d39a583654eb583d`.
Consumed(P10)/NoneTerminal/Unissued, no transfer; exact task refs absent. Those
historical gates create no current commit/publication authority.

DP-027 Draft/Partial, DP-009 Draft, ADR-0004, Frozen ARCH-002, Active ARCH-003/004/005
preserved. Opaque copyable View captures Reader value/private C1 Sink. Disabled →
InvalidInput → Busy; guard covers initial clear, exactly one Drain, detached copy,
deferred clear-before-unlock. Count == SinkSucceeded <=64; unused slots zero;
pre-sink discard absent; no stale replay. Well-formed prompt nonpanic Context remains
precondition. No hard wall-clock bound. NoItem != empty; Sequence not Host identity/
drop metric. Reader/View not fifth binding/Host capability/ADR004 carrier member;
Host sole lifecycle owner; no Drain/sink in callbacks/Observer/Complete/lease/Stop.
No architecture amendment needed; any new decision/out-of-scope change => STOP.

## Existing Coverage and proof obligations

Existing Coverage before test mutation: mailbox schema/fixed64/FIFO/copies/busy/
invalid/exhaustion/loss/concurrency; C1 C01–C16 receipt/bounds/cancel/guard/partition;
Runtime W01–W05 copied Producer/pair isolation/rollback-retry/Stop independence.
Gap: no View, outer capture guard, detached batch/zeroing/Poll composition.
Premutation external anchor `TASK-120-contract.md` was written before product/test
changes; repository record added only after explicit supplementary docs permission.
Added Proof Tests: new-package O01–O08/O10–O12, O14 and local O09 NoItem scenario.
Added Regression Tests: none to existing suites; unchanged suites run as regressions.
Remaining Limitations: O09/O13 compositional; actual Host binding not proved by View;
Windows CGO0/no C compiler, historical race/CRLF/parallel/native limits not removed.

| Proof | Required disposition |
| --- | --- |
| O01–O03 | Inert/zero construction; Reader capture/copies/isolation; input/Busy precedence and no rejected work |
| O04–O05 | Deterministic concurrent/reentrant Busy, outer TryLock audit; budgets1/64/refill; exactly-one Drain audit + C1 bound |
| O06–O08 | Receipt equations/discard absent; FIFO detached batch/mutation isolation/zero unused slots; initial/final guarded clearing/no replay |
| O09 | COMPOSITIONAL mailbox busy/invalid + C1 false handling; local NoItem scenario, no fake Reader |
| O10–O12 | Three exact cancel samples; externally joined caller presentation fixture/independent progress; destructive partition and separate Views/mailboxes |
| O13 | COMPOSITIONAL retained capability with unchanged Host/component/C1 rollback/Stop proofs; no flush/reset/join/final-empty claim |
| O14–O16 | Safe fixed64/scalar API/no loss inference; godoc/import/scope audits; focused/full checks with explicit unavailable limits |

## Roles, scope and recovery

Coordinator root sequentially assigns itself Developer/Documentation recorder;
tests120 authors only new tests; independent architecture_intake supplies read-only
conformance; verify120 independent Verifier; review120 non-author Reviewer/Scope
Auditor; independent final-byte reader assigned after decision. No Publisher.
Ordered stages: intake/accepted Architecture Confirmation → bounded implementation →
source conformance/Verification → PROCESS-002 → subject freeze → independent final
Verification/Scope Audit/Review → Coordinator Acceptance → independent integrity → STOP.
Affected/downstream gates repeat after rework. No self-review as independent Review.

Size Guard: one package/two product files/one behavior. Reassess >15 changed paths,
>500 production lines or >1 package/contract/independent behavior.
Stop on unexplained baseline/diff, absent decision, outside source/test change,
blocking finding, unresolved documentation drift or unknown evidence.
Exact subject paths/manifest and reproducible checkpoints belong to envelope;
append-only entries after freeze, no self-hash of final evidence bytes.

## PROCESS-002 applicability

Required: this record; task index; PROJECT_CONTEXT/current-state leading current-task
and published TASK119 facts; EN/RU DP027 implementation note; EN/RU MASTER_PLAN only
diagnostics candidate/debt paragraph. DP027 historical proposed §29–34 retained as
design baseline, current note separates prepared source from pending Acceptance.
DP009 N/A: Reader external/fourth Producer only unchanged. ADR/ARCH/spec decisions
N/A: no decision changed. Root README/Wiki/public indexes/CHANGELOG N/A: isolated
internal leaf, no operator exposure or release capability. Historical task records
immutable. Scope Audit must include every supplementary docs path.

## Recovery Evidence Envelope

### R120-001 — Admission and implementation preparation

HEAD/status rechecked clean on accepted baseline. Sandbox branch creation refused
without effects; supported escalation created only exact task branch. External
premutation anchor preceded both product mutations. No stage/commit/publication.
Independent Architecture Confirmation already PASS, accepted by Owner. Source-backed
Architecture Conformance PASS/0 findings from architecture_intake on pull.go SHA256
`17779ce3ff3769833a5c60b3ec913cd4d67d147ce2a2ab4f76e3ac213555cffd`.
Tests author reports focused count1 and shuffle/count100 PASS; full/independent
checks pending. No Acceptance claimed by this entry.

### V120-PRIMARY-001 — Implementation and primary verification

Final source SHA256 unchanged from Architecture Conformance. Final tests SHA256
`7df0da67581607489fbf309e9de6b2d888493bf5db697f2706250234c3099b6c`.
R120-TEST-01 preliminary review finding: O11 failed assertion consumed result before
Cleanup receive and could hang; Tester corrected O11 and analogous O04 to close-only
completion channels with retained results. Initial historical test hash/evidence
superseded; corrected focused count1 and count100/shuffle/timeout90s PASS exit0.
Independent final Review remains required; primary author does not self-review.

External artifact root for this cycle:
`C:/Users/dsdred/.codex/visualizations/2026/11/04/01a1a7f6-58f2-7121-80f4-17d6f82a85e6`.
`execution120-primary.json`, `checks120.json` and `check120-*.log` retain commands,
exits and limitations. `go build -mod=readonly ./...` PASS in sandbox and ordinary
Windows context. `go test -mod=readonly -count=1 ./...` sandbox exit1: loopback
connectex access denials/downstream timeouts and containment rename AccessDenied;
historical failure preserved, no assertion every failure individually diagnosed.
Same full command on final bytes in ordinary Windows (`require_escalated`, native
session5535) PASS exit0 all packages. `go vet -mod=readonly ./...` PASS exit0;
independent final-hash focused vet PASS. No existing test/source was changed.

Race `go test -mod=readonly -race -count=1 ./...` NOT RUN, exit2: requires cgo;
Go1.26.5 windows/amd64, CGO_ENABLED=0, no gcc/clang found. This is no race PASS or
Windows/native TASK102 qualification. `go mod tidy -diff` checkout exit1 from
CRLF-only go.sum diff; exact same root sources with external LF module copies via
`-modfile=.../tidy120/go.mod` PASS exit0 (native session72988). Original module files
unchanged (`git diff --exit-code HEAD -- go.mod go.sum` exit0); normalized equality
independently checked. Source/module content proof does not erase checkout limit.
gofmt -l new files empty; git diff --check PASS.

O01–O08/O10–O12/O14 executable tests + source audit; O09 COMPOSITIONAL unchanged
mailbox busy/invalid and C1 false handling; O13 COMPOSITIONAL retained View Reader/C1
with unchanged Runtime W01–W05/reader-outlives-producer/retained C1 tests. Separate
independent Runtime targeted ordinary-Windows run PASS; sandbox denial retained.
O15 independent Architecture Conformance/source/import/API audit PASS. O16 executable
checks above, PASS WITH LIMITATION pending independent full-subject verification;
no O proof is production deployment/binding evidence. P01–P08/S01–S12 DEFERRED.

### D120-001 — PROCESS-002 and primary Scope Audit

Owner explicitly authorized minimal supplementary docs after product admission.
Ten paths: two new product files, new task record, context/current-state/task index,
DP027 EN/RU current implementation note and diagnostics-only MASTER_PLAN EN/RU
candidate paragraph. Scope primary **10 Required / 0 Questionable / 0 Removable**:
product/test needed for exact API/proofs; record needed for recovery/evidence;
three navigation/state files needed to remove stale active-task/nonexistent-source
claims and record TASK119 publication; DP mirrors separate historical proposed
design from current authorized source state; roadmap mirrors remove stale successor
admission while preserving all other items/criteria. Each omission leaves DoD or
mandatory applicability incomplete. No generated/foreign files in repository diff.

`python verify120.py docs`: exit0, 339 local links/0 missing, strict UTF8/no BOM/NUL
for changed docs and git diff-check PASS; semantic EN/RU current notes aligned.
Early mixed-newline documentation normalization was corrected before freeze:
unchanged historical regions restored from baseline objects; final diff bounded.
Historical task records, existing source/tests/modules and TASK102 branch untouched.
DP027 Draft/Partial, DP009 Draft, Gate11 OPEN, production C2/C3 inactive.
Independent PROCESS-002/Scope Audit/final Review and Acceptance not supplied by
this primary entry. Freeze next uses all ten exact paths, task-record-v1 projection;
after freeze only append-only envelope metadata is allowed without fresh affected
gates. No Commit/Publication Gate or successor activation.

### M120-001 — Frozen implementation subject

HEAD/baseline `ac021e9402f37d60ab826f100534c88415c08a21`, branch
`feature/task-120-runtime-diagnostics-pull`, Git object format sha1. Canonical
PROCESS-001 manifest: ascending unsigned UTF8 path-byte order, each row
`path\0projection\0present\0mode\0oid\0`; 977 bytes; Git blob manifest OID
**e411f30c11ac8dbc6bef120898b3a1124ba79272**. New/untracked paths included without
stage. Raw no-filter full bytes; task-record-v1 excludes status evidence body and
terminal envelope only. Reproduce external `python verify120.py manifest`;
subject120.json contains ordered rows/raw hash snapshot, not self-attestation.

| Path | Projection | Mode | Projected blob OID |
| --- | --- | --- | --- |
| .ai/PROJECT_CONTEXT.md | full | 100644 | 40400610b95828ba8305ae0b4738b0452c9ae52c |
| docs/en/design/DP-027-runtime-diagnostics-v1.md | full | 100644 | 7e551b9f3ed14c2c852c8899005106c20244802f |
| docs/en/roadmap/MASTER_PLAN.md | full | 100644 | 83f8bc361f9ba09d0e3300d2d6260b99ec0b68b3 |
| docs/ru/design/DP-027-runtime-diagnostics-v1.md | full | 100644 | ca8234cae1dcf5bd1571e048ba870a57eebce3c6 |
| docs/ru/roadmap/MASTER_PLAN.md | full | 100644 | dede147523f21989657f10ce6827a0870d616507 |
| docs/tasks/README.md | full | 100644 | 17e61e8c5e27c9150c5c27cb6e3009fd914d42bf |
| docs/tasks/TASK-120-RUNTIME-DIAGNOSTICS-PULL.md | task-record-v1 | 100644 | b9b78b7c9a7219addcbfd592ec6e50559035031f |
| internal/runtimediagnosticspull/pull.go | full | 100644 | 0e62cf8b2faef524ebbae764455f3ae04504795f |
| internal/runtimediagnosticspull/pull_test.go | full | 100644 | ebabfd44c205114f9421d060de5c604f993eef3a |
| spec/current-state.md | full | 100644 | 9048b156f7496951003ea1c8b65d1ef3c66062eb |

Independent full-subject Verification/PROCESS-002/Scope Audit/final Review required
on this identity before Coordinator Acceptance. After freeze append-only envelope
metadata only; any other subject change invalidates affected/downstream gates.

### V120-002 — Independent final Verification

Independent Verifier `/root/verify120_final`: **PASS WITH LIMITATION**, blocking
findings 0. It independently reconstructed exact M120-001: ten paths, 977 bytes,
manifest `e411f30c11ac8dbc6bef120898b3a1124ba79272`; final source/test SHA256 match.
Report `verification120-final.json`, SHA256
`466e54334279aff48b75a304e447eccb840d921993bc27699aa98ad2c6eda6ff`.

O01–O15 PASS; O09/O13 explicitly COMPOSITIONAL. O16 PASS WITH LIMITATION.
Verifier independently reran corrected O11 count100, package count100 shuffled stress,
focused vet and exact mailbox/C1/Runtime/View O09/O13 suite: exit0. Source conformance
DP027 §29/§32–34 PASS. PROCESS-002 PASS: 339 links/0 missing, strict encoding,
EN/RU parity, project-state/applicability. Scope support 10/0/0.

Limitations remain evidence, not successful checks: Windows race **NOT RUN** because
CGO_ENABLED=0/no C compiler and `-race` refused; no Windows/native TASK102
qualification. Sandbox full regression failure remains recorded with loopback
connectex restrictions, downstream timeouts and containment rename AccessDenied;
ordinary Windows full regression on final test bytes PASS does not relabel sandbox
failure or prove one root cause for every downstream failure. Checkout tidy exit1
remains CRLF-only discrepancy; normalized module equality and external LF modfile
tidy PASS prove source/module content only; original module files unchanged.

### R120-002 / S120-001 — Independent Final Review and Scope Audit

Independent non-author Reviewer `/root/review120`: **APPROVED**, blocking/unresolved
findings 0. Exact M120-001 independently reconstructed twice. Report
`review120-independent-final.json`, SHA256
`6a392bb6a14bc435b18b8966bb7edf6798f6fae3417a7fb35c3c84d0a66a5b01`.
R120-TEST-01 RESOLVED on final test hash: O11/O04 use close-only completion channels
and retained results, so failure paths join without a second blocking result receive.
PROCESS-002/manual EN/RU parity PASS. Limitation classifications preserved.

Scope Audit: **10 Required / 0 Questionable / 0 Removable**. Reviewer answered the
removability question per path: each Go file supplies the only authorized behavior/
proofs; task record is required recovery/evidence; three state/navigation paths
remove stale current-task/nonexistent-source facts and record TASK119 terminal
publication; DP/roadmap mirrors preserve current execution vs historical planned
state and dependency/Gate11 boundaries. Removing any path makes product DoD or
mandatory PROCESS-002 incomplete. No foreign/generated repository path, existing
source/test/module change, TASK102 effect or next-task activation observed.

### CA120-001 — Coordinator Acceptance Verdict

Coordinator `/root`: **ACCEPTED — ISOLATED C2-PULL, WITH VERIFIED LIMITATIONS**,
blocking findings 0, exact M120-001
`e411f30c11ac8dbc6bef120898b3a1124ba79272` (ten paths/977 bytes).
Decision uses independently reconstructed V120-002 and R120-002/S120-001, accepted
Architecture Confirmation and source-backed Architecture Conformance PASS.

Accepted capability is only the isolated process-local View.Poll leaf: New captures
Reader by value; opaque copies share one outer TryLock; one admitted Poll makes
exactly one C1 Drain, returns a detached fixed64 batch and clears temporary storage
before unlock. O01–O16 acceptance dispositions above are complete within stated
limits. O09/O13 remain compositional; this decision does not claim production Host
binding, operator exposure, race evidence or native qualification.

Windows race stays NOT RUN; sandbox failed run and checkout CRLF tidy discrepancy
remain historical limitations exactly classified by V120-002. These do not conflict
with PROCESS-001 technical-unavailability path because available stress, focused,
compositional, full ordinary-Windows regression, build/vet and module-content checks
PASS and no normative source requires unavailable race to be fabricated. DP027 stays
Draft/Partial, DP009 Draft; Gate11 OPEN; production C2/C3 Not Activated; P/S rows
deferred; TASK102 untouched.

Implementation task result is Coordinator Accepted in isolation. Projected sources
remain verification-stable `In Progress`; this exact envelope decision resolves the
actual gate after independent manifest reconstruction. Commit and Publication Gates
remain **NOT AUTHORIZED / NOT PERFORMED**. No successor is activated.

Separate independent post-decision final-byte integrity is REQUIRED. Pre-CA append
task raw SHA256 was
`f2d3259291aacfbdc26e1554e9ea515b7568a406d6c027304302b97316421748`;
accepted projection and all other nine full rows must remain M120-001. After this
append no repository mutation is permitted before final actual-byte readback.

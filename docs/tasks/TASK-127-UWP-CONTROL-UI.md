# TASK-127 — UWP Control UI v0.1

## Status

In Progress. Exact gates, subject and first incomplete stage resolve through the
newest matching Recovery Evidence Envelope. Commit and Publication not authorized.

## Task Contract

- Repository: `dsdred/universal-websocket-platform`.
- Task Mode: Implementation; Owner explicitly selected UWP Control / Slate Blue.
- Branch: `feature/task-127-uwp-control-ui`.
- Trusted baseline: `e5975d5a36cd4c1e431ad76d0629275f17ceab81`.
- Why Now: Owner prioritizes a usable local UI after terminal TASK-126.
- Definition of Done: embedded GET-only `/dashboard`, common responsive shell,
  actual health and Workspace -> Configuration -> Version browsing, listener
  host/port and available timestamps; loading/empty/offline/error/unsupported
  states and manual Refresh; preserved Studio create/publish flow; focused,
  regression, build/vet, browser/network/responsive checks; independent review,
  PROCESS-002, scope audit and reviewable Pre-Commit Report.
- Out of Scope: API changes, npm/dependencies/CDN, polling, probes, Runtime,
  Connections or Diagnostics API, fabricated metrics/activity/uptime; Stage B,
  TASK-102, B1/B2, production C2/C3, C2-Exposure, activation, commit/publication.
  Gate 11 remains OPEN. Published != Running; health confirms only Control Service.
- Security: trusted localhost/same-origin, unauthenticated service unchanged;
  retain CSP/nosniff; textContent allowlist, no raw JSON/secret references;
  reject unsafe uint64 IDs before rendering/navigation, never round them.

## Admission and Native Ownership

Clean main/HEAD/origin-main matched baseline before task preparation. TASK-127
record/content/ref absent. TASK-102 ref is
`8bfa2e22a4b6310a71f6d36c8724b1adedd727ad`. No other current active implementation
was found. Historical projected In Progress TASK-126 is terminal, not active.
Native read_thread confirmed root `01a1d052-268c-7532-a69d-c795044c889c` terminal
P10R recovery turn `01a1de6d-bcdd-7e43-bbd0-28083dd7ebb5`, and independent receipt
`msg_0b9d820147bf1cc9016af97587a44c87d2aaa49d9d6a8e6c9c` in observer thread
`01a1de6e-f126-7981-86a6-c5f21119698b`: P10R-126-001 PASS, exact baseline merge,
Consumed(P10)/NoneTerminal/Unissued. Previous UI intake turn
`01a1de76-464c-7a60-b43f-591dbe27a5d2` ended read-only STOP without task/branch.
Current native root `01a1df3b-0f9c-7a02-80b9-20447eca4254` owns only current
Owner-authorized TASK-127 work; no Publisher/InitialDispatch established.
Branch creation first failed without ref creation; inspect confirmed absence;
supported escalated retry succeeded. No stage/commit/fetch/push/PR/merge.

Reference exists at `C:/Users/dsdred/Downloads/uwp-control-slate.html`, SHA256
`c636545e69210e16fd096a0b90f278319f08440b1d51fac2fa90583c4fd47f5d`.
Used only as visual reference; its fictional metrics and inline scripts excluded.

## Sources, Roles and Ordered Stages

AGENTS/AGENT, PROCESS-001/002, role contracts, project context/specs, ADR-0001,
Frozen ARCH-002/Active ARCH-004/005, existing TASK-076 Studio and actual HTTP
handlers define boundaries. Owner direction does not alter architecture/API.
Root Coordinator assigns Developer and Documentation roles sequentially to root;
read-only Architect and independent Reviewer/Verifier delegated by mandatory
Coordinator role contract. Author cannot provide independent final review.
Stages: admission/record -> Architecture Confirmation -> implementation ->
verification -> PROCESS-002/scope -> exact manifest -> independent final review ->
Coordinator acceptance/Pre-Commit Report -> STOP. Rework repeats affected gates.

## Size Guard — SG127-001

Initial cohesive slice: 0 new packages/contracts, 1 browsable read-only overview
with shared shell, estimate <=500 added production LOC. Up to 17 paths expected
because preserving Studio needs its HTML wrapper plus shared CSS, and mandatory
project-state mirrors require documentation. The >15-path trigger is explicitly
reassessed: these shell/assets/tests and status/user-guide paths jointly satisfy
DoD, not separate behaviors. No extra screens/API/refactoring. Reassess actual
LOC/path set before acceptance; STOP for new contract/independent behavior.

Intended paths: task record; configurationstudio handler.go/handler_test.go;
assets index.html/dashboard.html/control.css/dashboard.js; optional existing
studio.css only if required; focused JS behavioral tests; .ai/PROJECT_CONTEXT.md,
spec/current-state.md, task index, MASTER_PLAN EN/RU, README.md/README.ru.md,
CHANGELOG.md. No other code paths authorized.

## Existing Coverage Report and Verification Plan

Existing: configurationstudio handler tests prove embedded bytes, exact routes,
CSP/nosniff, local dependency/security notices and Studio mutation sequence.
Workspace/Configuration/Version handler/service tests cover actual GET models
and create/publish. Gap: Dashboard behavior, selection races, invalid responses,
safe IDs, secret omission, browser read-only network and responsive shell.
Planned proof: add exact asset/method tests and JS behavior harness exercising
fetch/DOM behavior; real browser smoke with empty/populated data, navigation,
Refresh, narrow viewport, Studio validation and successful publish; GET-only
network audit scoped to Dashboard. Regression go test ./..., go build ./...,
go vet ./...; race N/A if no concurrent Go changes. No dependency changes/tidy
required. Record exact exits/limitations. Test fixtures never shipped as UI data.

## PROCESS-002 Applicability and Stop Conditions

Required: task/context/current-state/index/MASTER_PLAN mirrors (current task and
terminal TASK-126 reconciliation), root README mirrors and CHANGELOG (user UI).
Design indexes/DPs/spec-decisions N/A: no design/status/API change; retain Stage B
Not Activated. Historical TASK-126 record immutable. Documentation drift in seven
navigation paths is stale terminal publication state, reconciled within scope;
design indexes assessed separately, no UI design-status promotion.
Stop on unknown ownership, unrelated dirty bytes, unavailable reference, missing
architectural prerequisite, new scope, failed required checks or blocking review.
No next candidate activated. Future Runtime/Connections views remain Not Activated.

## Implementation Result and Pre-Commit Report

Реализован один GET-only Dashboard `/dashboard` в existing Go embed package.
Общая оболочка UWP Control / Slate Blue использует Owner reference tokens,
боковую навигацию на desktop и постоянно видимое переносимое меню на mobile.
Studio получил только wrapper/theme; `studio.js` и `studio.css` byte-exact baseline.
Реальные данные загружаются последовательно Workspace -> Configuration -> Versions;
health независим. Manual Refresh сохраняет существующие selections и очищает
устаревшие данные. Abort/generation guards исключают stale response rendering;
15s deadlines ограничивают ожидание, не создавая polling/retry requests.
Показаны только allowlisted names/IDs, version number/state, host/port и updatedAt.
Все IDs проверяются до display/URL; unsafe IDs означают Unsupported response.

Exact changed files — 16 Required / 0 Questionable / 0 Removable:

1. `.ai/PROJECT_CONTEXT.md` — current task и terminal TASK-126.
2. `CHANGELOG.md` — unreleased user-facing UI.
3. `README.md` — UI behavior и local run.
4. `README.ru.md` — semantic mirror.
5. `docs/en/roadmap/MASTER_PLAN.md` — capability/current task/terminal TASK-126.
6. `docs/ru/roadmap/MASTER_PLAN.md` — semantic mirror.
7. `docs/tasks/README.md` — navigation/current task/terminal TASK-126.
8. `docs/tasks/TASK-127-UWP-CONTROL-UI.md` — contract/recovery/report.
9. `internal/configurationstudio/assets/control.css` — shared responsive theme.
10. `internal/configurationstudio/assets/dashboard.html` — Dashboard/read-only controls.
11. `internal/configurationstudio/assets/dashboard.js` — GET loading/rendering/states.
12. `internal/configurationstudio/assets/index.html` — minimal Studio shell wrapper.
13. `internal/configurationstudio/dashboard_test.js` — executable browser-logic proofs.
14. `internal/configurationstudio/handler.go` — exact embedded asset routes.
15. `internal/configurationstudio/handler_test.go` — routing/security regression.
16. `spec/current-state.md` — factual capability/current task/terminal TASK-126.

SG127-002 reassessment: PASS for 16 paths (>15 trigger explicitly evaluated),
341 added production LOC, 0 removed production LOC, 0 new packages/contracts,
one cohesive behavior. Seven code/test paths plus task record/eight required
documentation paths; removing any loses DoD behavior, proof, recovery or mandatory
synchronization. No separate Runtime/Studio rewrite/refactoring behavior added.
Production LOC counts all physical HTML/CSS/JS and Go additions, including blanks.

Tests/results on final production bytes:

- `node --test internal/configurationstudio/dashboard_test.js`: exit 0, 15/15 PASS.
  Loading/empty, selections/Refresh, GET-only/no polling, unsupported shape/JSON/
  content-type, offline/HTTP error, duplicates/unsafe IDs/parent mismatch,
  secret omission/text rendering, stale requests/deadlines, health/data independence.
- `go test ./internal/configurationstudio`: exit 0 PASS, latest independent run 1.644s.
- `go test ./...`: exit 0 PASS, all packages, including existing Studio/API regressions.
- `go build ./...`: exit 0 PASS; `go vet ./...`: exit 0 PASS.
- Browser smoke on final embedded assets: localhost 18081, empty lists; Studio
  empty-field validation; real create/configure/publish success (`Published v1`,
  listener 127.0.0.1:19090); Dashboard selection/Versions/Refresh retained IDs;
  Workspaces/Configurations anchors; unavailable Runtime/Connections/Diagnostics.
- Real offline check: stopped exact smoke service, Refresh showed Offline and
  cleared stale lists/versions; restarted service and Refresh restored Available
  with genuinely empty lists (in-memory data lost). No probe of listener port.
- Browser GET-only network audit through temporary localhost proxy18082->18081:
  first 11 requests cover Dashboard/assets/health/lists/selection/Refresh; all GET.
  Subsequent 4 Studio navigation/assets requests also GET. No mutation/probe/polling.
  Studio create/publish was tested separately on18081, outside Dashboard audit.
- Responsive: Dashboard375x812: document/body360px (15px scrollbar), visible
  wrapping menu, table internal scrolling. Studio320x740: document/body305px,
  all navigation links visible. Temporary viewport override reset.
- No Go concurrency changes: race N/A. No module/import/dependency changes:
  tidy N/A. Tests use only Node built-ins; Node is a test tool, not runtime/build dependency.

Screenshots and audit are outside Git, in
`C:/Users/dsdred/.codex/visualizations/2026/11/15/01a1df3b-0f9c-7a02-80b9-20447eca4254/`:
`dashboard-desktop.jpg`, `dashboard-mobile.jpg`, `dashboard-offline.jpg`,
`studio-published.jpg`, `studio-mobile.jpg`, `dashboard-network.jsonl`.
Temporary audit proxy stopped. Owner preview: `http://127.0.0.1:18081/dashboard`;
smoke service left running on localhost only. Reproduce with README commands.

Known limitations: IDs above Number.MAX_SAFE_INTEGER unsupported rather than
rounded; lists are non-atomic process-local snapshots loaded on demand, with no
polling/pagination/global aggregates. Resource updatedAt is shown for Versions;
browser fetch/check times are clearly distinct. Failed health request cannot
distinguish network failure from a stopped service. Runtime/Connections/Diagnostics
unavailable, Published != Running. Control Service remains unauthenticated and
existing mutation APIs/Studio remain usable; this is only trusted local development.
VM proofs cover malformed-response injection; real browser covers live data,
offline, responsiveness and Studio, not every injected error fixture.

PROCESS-002: eight documentation paths synchronized; root README and MASTER_PLAN
mirrors semantically aligned, new local links valid, unchanged bytes preserved.
spec/decisions, ADR/ARCH/DP and design indexes N/A: design/status/API unchanged;
historical TASK-126 projected record/index references still resolve via its envelope,
while current state/navigation now explicitly records terminal publication.
No Stage B/production activation and no TASK-102 edits. Review must confirm every
Required path: removing it would lose behavior, verification or mandatory evidence.
Independent final review and Coordinator Acceptance pending; no commit/publication.

## Recovery Evidence Envelope

Bootstrap: admission complete, branch/record prepared. Architecture confirmation
pending. No implementation/test/review/acceptance PASS claimed yet.

A127-001 Architect: APPROVED — IMPLEMENTATION READY, blocking findings0, independent
`/root/architecture`; required boundaries incorporated without new contracts.
T127-VERIFY-001 Tester: PASS15/15 JS and focused Go, blocking findings0,
`/root/verification`; exact production dashboard.js SHA256
`da3f933cba6fe2ec8201a2b4bebaa8a925482aac4bf44f966e3ecfb70917e4db`.
Documentation role `/root/documentation`: eight Required paths, links/mirrors/
preservation/diff-check PASS; no design/API promotion. Full production regression,
build/vet and browser evidence recorded above. Final canonical subject/review pending.

M127-001 frozen for independent final review: anchor HEAD/base
`e5975d5a36cd4c1e431ad76d0629275f17ceab81`, object format sha1;
canonical manifest `cb1a357c6173f08940030cace04b1a0f0c232c5c`.
Ascending unsigned UTF-8 path-byte order; each row encodes path, projection,
present,100644,OID with NUL separators and final NUL. `full` uses exact raw bytes,
not EOL normalization; task-record-v1 excludes status body/envelope per PROCESS-001.

```text
.ai/PROJECT_CONTEXT.md full present 100644 de9618710a2c094f2fb21b276d5898d2d32d7b07
CHANGELOG.md full present 100644 6f0451c425439e7674dada8727314624709aa0f5
README.md full present 100644 15c0ac1b79beeb59bd99f02e1b1a242f0c0ec169
README.ru.md full present 100644 ff7206ac56a9606ac3c787e3ee3ade998164f129
docs/en/roadmap/MASTER_PLAN.md full present 100644 7d229490f110b09c9bfc7bcad8ff6fae02e64b28
docs/ru/roadmap/MASTER_PLAN.md full present 100644 d9edc8cfadc11f6a2330c02c8df4dc7e3abc5213
docs/tasks/README.md full present 100644 759e7961b42b287936dcbbfbc6880b7c7ebc5064
docs/tasks/TASK-127-UWP-CONTROL-UI.md task-record-v1 present 100644 b57f759532155b0a7f34461fc851cb0a87a0810b
internal/configurationstudio/assets/control.css full present 100644 7758d463e511f91f3b3a0a5aa6a27cf1942902d9
internal/configurationstudio/assets/dashboard.html full present 100644 4ea7a26f1172c48fa564a5916faf0b9e53087e78
internal/configurationstudio/assets/dashboard.js full present 100644 dbf06d4a2692afd32fdb29fb22da5d546870abb6
internal/configurationstudio/assets/index.html full present 100644 d7256a1fb82eeb0a78366efb5993a895a1a9864b
internal/configurationstudio/dashboard_test.js full present 100644 49c66720cfc7a2a66bbf63857505bb950e278074
internal/configurationstudio/handler.go full present 100644 7116ce216baf1b2497e07f4e29fec0959d8576fd
internal/configurationstudio/handler_test.go full present 100644 2bb94af95a7551a7705e383ce344b39312a737fb
spec/current-state.md full present 100644 34eefd6720f8477f9c013dfa5709ca94d7bb9085
```

R127-003 — Approved, Critical0/Major0/Minor0/blocking0. Independent Reviewer
recomputed manifest `567957cd183af39df1e0fc9e8fff77543a0cb25d`, confirmed14
projected rows exactly M127-001 and only handler.go/index.html CRLF->LF, original
byte reconstruction matches accepted OIDs. Genuine gofmt/ref guards PASS; same
behavior/security/API/docs meaning and prior executable evidence. Existing M2
index explicitly not approved; exact M3 raw staged gate remains mandatory.

CA127-003 — Coordinator Accepted final M127-003 / sha1 manifest
`567957cd183af39df1e0fc9e8fff77543a0cb25d`,2026-11-15, same baseline/branch/16paths.
This supersedes CA127-002 for the one current Owner-authorized commit. Scope
16 Required/0 Questionable/0 Removable and341 added production LOC unchanged;
eight documentation raw files restored, only two shell files normalized.
PROCESS-002/Size Guard/review recheck PASS. Full original behavioral proofs and
fresh JS15/15/focusedGo PASS remain applicable via exact content preservation.
Commit message and one-local-commit/no-publication boundary remain CG127-002.
Current M2 index verified as attributed, no unexpected staged edit; replace only
the same16 paths, verify raw bytes and compact staged diff, then commit once.

Root verification final checks: all16 paths attributed, real index empty;
HEAD/origin-main unchanged; TASK-102 ref unchanged. Studio script/styles and
historical TASK-126 record unchanged; `git diff --check` exit0. No generated
repository artifacts, staged files, commit, push, PR, merge or publication.
Next gate: independent Final Review of M127-001, then final-byte integrity and
explicit Coordinator Acceptance. Pre-Commit STOP only after those gates.

R127-001 — Approved, Critical0/Major0/Minor0/blocking0. Independent Reviewer
`/root/architecture` recomputed M127-001 with a separate raw-byte hashlib Git-blob
algorithm: all16 ordered rows and manifest match. Independent reruns:15/15 JS
PASS; focused Go PASS1.625s; diff-check, normalized-content gofmt, navigation,
index/ref guards PASS. Reviewed full code/docs diff, screenshot and network
artifacts, security/unsafe IDs/async lifetimes/Studio preservation and removal
question.16 Required/0 Questionable/0 Removable; SG127-002 PASS. No mutations.
This receipt attests projected subject only; envelope bytes excluded.

Final-byte integrity after R127-001: manifest independently regenerated as
`cb1a357c6173f08940030cace04b1a0f0c232c5c`, exact16 paths; HEAD/base and branch
unchanged, real index empty, TASK-102 ref unchanged, diff-check exit0. Changed Go
content equals gofmt after preserving existing CRLF convention. An additional
raw-local-vs-Git comparison on untouched Studio files exposed expected checkout
CRLF versus Git LF; corrected check uses configured checkout filters and proves
baseline Git content equality. No product rework/normalization was required.

CA127-001 — Coordinator Accepted, 2026-11-15, exact M127-001 and manifest above.
Architecture Confirmation, Implementation, applicable Verification, independent
Final Review, PROCESS-002 and Scope Audit complete. Definition of Done satisfied
with documented safe-ID/process-local/unavailable-API limitations. Pre-Commit
Report is the Implementation Result section above plus this latest envelope.
Commit Gate and Publication Gate remain separate, unauthorized Owner gates;
no stage/commit/push/PR/merge/publication performed. Projected In Progress fields
are verification-stable projections resolved by this matching terminal envelope.
No next task activated. Stage B, B1/B2, production C2/C3 and C2-Exposure remain
Not Activated; Gate 11 OPEN; TASK-102 untouched.
STOP AT PRE-COMMIT GATE. Only later explicit Owner authorization may open Commit Gate.

CG127-001 — Owner requested one local commit. Inspect-first gate found raw accepted
CRLF/mixed-EOL bytes differ from normal staging under core.autocrlf=true in ten
already authorized paths. No stage/commit attempted. No LF/CRLF equivalence is
claimed. Coordinator reopens only bounded EOL rework: convert CRLF to LF in those
ten paths, prove all other content identical, regenerate manifest and repeat
affected checks/independent Review/Acceptance before one task commit. M127-001
remains historical and is not used for the new staged tree. No new paths or
product behavior authorized. Owner's current one-commit direction remains pending;
no publication authorization is inferred.

M127-002 — bounded EOL rework subject. Same anchor/base/branch, sha1 and exact16
ordered paths. Canonical raw-byte manifest:
`e9bb8a78794b7551652bfbe82bba2a543a89f858`.
The ten old/new raw OID pairs are independently reproducible in external artifact
`commit-eol-rework.json`; transformation is only CRLF -> LF. No contents, runtime
behavior, path set, production LOC, dependency, API or Size Guard change.
Six other projected rows remain exactly M127-001. Genuine gofmt on both changed
Go files is now clean without normalization; git diff --check PASS. Historical
raw-byte-preservation assertions above refer to pre-rework checkpoints, not an
equivalence claim for this new subject. Prior browser/full regression/build/vet
evidence applies to identical executable content; affected focused tests repeated.
New independent review/Acceptance pending; M127-001 is not the commit subject.
Rows use the same NUL encoding/projections as M127-001:

```text
.ai/PROJECT_CONTEXT.md full present 100644 69fddf7ff2fd214bc1721b150433cb7fb52f2fd2
CHANGELOG.md full present 100644 830a6a7821e358a32adab77de79570c9075b190a
README.md full present 100644 3e943f8ca60e32f66f49e3db46e5c54b693fb178
README.ru.md full present 100644 5e26f6d581cbeca7feeed9e53e8a85c7106c42b2
docs/en/roadmap/MASTER_PLAN.md full present 100644 abe463737534676380f7ca30c5ac727cef2b0c90
docs/ru/roadmap/MASTER_PLAN.md full present 100644 51ce2b56fea6545cdd1973c6ba7033eaa67460c9
docs/tasks/README.md full present 100644 959c260dcfdbe76049f3317e973d7db2ee5330e5
docs/tasks/TASK-127-UWP-CONTROL-UI.md task-record-v1 present 100644 b57f759532155b0a7f34461fc851cb0a87a0810b
internal/configurationstudio/assets/control.css full present 100644 7758d463e511f91f3b3a0a5aa6a27cf1942902d9
internal/configurationstudio/assets/dashboard.html full present 100644 4ea7a26f1172c48fa564a5916faf0b9e53087e78
internal/configurationstudio/assets/dashboard.js full present 100644 dbf06d4a2692afd32fdb29fb22da5d546870abb6
internal/configurationstudio/assets/index.html full present 100644 6e1d83c5aba42983b7f4c92d5f06677acf5adb77
internal/configurationstudio/dashboard_test.js full present 100644 49c66720cfc7a2a66bbf63857505bb950e278074
internal/configurationstudio/handler.go full present 100644 bea464c6a0795270cce7edfe96ae0fde0a7b3607
internal/configurationstudio/handler_test.go full present 100644 2bb94af95a7551a7705e383ce344b39312a737fb
spec/current-state.md full present 100644 3f0c6df90f2371b0eb50177482f7e2583346dd23
```

R127-002 — Approved, Critical0/Major0/Minor0/blocking0. Independent Reviewer
`/root/architecture` independently recomputed M127-002 and proved all ten changed
rows are CRLF -> LF only against M127-001 OIDs (eight original blobs available;
two reconstructed CRLF byte streams matched exact old OIDs). Six other projected
rows unchanged. Genuine gofmt empty, diff-check/exact16 paths/index/ref guards
PASS; independent JS15/15 PASS; root focused Go exit0 PASS1.587s. R127-001's
behavior assessment applies through this exact transformation proof, but its
old manifest is not used. No reviewer mutation or self-attestation of envelope.

CA127-002 — Coordinator Accepted, 2026-11-15, exact M127-002 / sha1 manifest
`e9bb8a78794b7551652bfbe82bba2a543a89f858`, baseline
`e5975d5a36cd4c1e431ad76d0629275f17ceab81`, branch
`feature/task-127-uwp-control-ui`. Rework/affected verification/independent review,
PROCESS-002 and scope recheck complete. Same16 Required,341 added production LOC;
no new behavior/path/dependency/contract. M127-001/CA127-001 remain historical.

CG127-002 — current Owner request authorizes exactly one local TASK-127 task
commit after gate validation; no push/PR/merge/publication. Proposed message:
`feat(control-ui): add read-only dashboard and shared shell` conforms to
CONTRIBUTING imperative/scoped policy and existing task conventions. No GPG/DCO/
sign-off required or added. Existing post-commit/post-checkout hooks contain only
shebangs. Final pre-stage check independently recomputed exact raw M127-002,
all16 paths, empty real index, unchanged baseline/branch/TASK-102 and diff-check.
Staging must match every full raw blob and projected task record exactly, plus
current append-only envelope, before commit. Any mismatch means STOP, not
LF/CRLF equivalence. Commit result/immutable OID will be reported outside this
commit's bytes after independently reading Git; no self-referential OID appended.
All nonactivated boundaries preserved. Publication Gate remains unauthorized.

CG127-003 — staged scope audit rejects M127-002 as commit target: normalization
of eight baseline CRLF documentation paths produces3410 removed lines of EOL-only
churn in staged diff, contrary to compact-slice intent. No commit attempted.
The raw staging comparison passed; subsequent restricted git write-tree failed
on index.lock permission, not a commit. HEAD remains baseline and index is known.
Coordinator narrows EOL rework: restore those eight docs from exact M127-001
raw blobs; keep LF only in handler.go and Studio index.html. Stage with transient
core.autocrlf=false to preserve exact accepted bytes without repository-config
changes. M127-002/CA127-002 remain historical; fresh M127-003 Review/Acceptance
required before updating staging and committing. No LF/CRLF equivalence claimed.

M127-003 — final compact commit subject, same anchor/base/branch and sha1.
Canonical manifest: `567957cd183af39df1e0fc9e8fff77543a0cb25d`. Fourteen projected rows are
byte-exact M127-001; only handler.go and Studio index.html use LF. The eight
documentation files retain their originally accepted raw bytes; new scope is
not activated. Genuine gofmt/previous focused tests remain applicable. Existing
M127-002 staging is attributed and must be replaced only after fresh acceptance.
Ordered rows (same NUL encoding and projection convention):

```text
.ai/PROJECT_CONTEXT.md full present 100644 de9618710a2c094f2fb21b276d5898d2d32d7b07
CHANGELOG.md full present 100644 6f0451c425439e7674dada8727314624709aa0f5
README.md full present 100644 15c0ac1b79beeb59bd99f02e1b1a242f0c0ec169
README.ru.md full present 100644 ff7206ac56a9606ac3c787e3ee3ade998164f129
docs/en/roadmap/MASTER_PLAN.md full present 100644 7d229490f110b09c9bfc7bcad8ff6fae02e64b28
docs/ru/roadmap/MASTER_PLAN.md full present 100644 d9edc8cfadc11f6a2330c02c8df4dc7e3abc5213
docs/tasks/README.md full present 100644 759e7961b42b287936dcbbfbc6880b7c7ebc5064
docs/tasks/TASK-127-UWP-CONTROL-UI.md task-record-v1 present 100644 b57f759532155b0a7f34461fc851cb0a87a0810b
internal/configurationstudio/assets/control.css full present 100644 7758d463e511f91f3b3a0a5aa6a27cf1942902d9
internal/configurationstudio/assets/dashboard.html full present 100644 4ea7a26f1172c48fa564a5916faf0b9e53087e78
internal/configurationstudio/assets/dashboard.js full present 100644 dbf06d4a2692afd32fdb29fb22da5d546870abb6
internal/configurationstudio/assets/index.html full present 100644 6e1d83c5aba42983b7f4c92d5f06677acf5adb77
internal/configurationstudio/dashboard_test.js full present 100644 49c66720cfc7a2a66bbf63857505bb950e278074
internal/configurationstudio/handler.go full present 100644 bea464c6a0795270cce7edfe96ae0fde0a7b3607
internal/configurationstudio/handler_test.go full present 100644 2bb94af95a7551a7705e383ce344b39312a737fb
spec/current-state.md full present 100644 34eefd6720f8477f9c013dfa5709ca94d7bb9085
```

Envelope assembly reconciliation: the R127-003/CA127-003 text above was inserted
by an ambiguous repeated-table patch anchor after M127-001 rows. It was written
now, after M127-003 review, not at that historical position. It is not treated as
ordered historical authority. Native tool history preserves the assembly error;
no earlier checkpoint or reviewer verdict is rewritten. Subject projection and
M127-003 remain unchanged. This final append supplies the current ordered gate.

CA127-004 — Coordinator Accepted, 2026-11-15, final M127-003 / sha1 manifest
`567957cd183af39df1e0fc9e8fff77543a0cb25d`, after independent R127-003 Approved
0/0/0 and exact 14 unchanged / 2 pure-EOL proof. This is the latest authoritative
matching acceptance. Earlier CA127-003 placement is superseded metadata, not
an alternate acceptance chain. Same baseline/branch,16 Required paths,341 added
production LOC; PROCESS-002/Size Guard/affected verification PASS. One current
Owner-authorized local commit only; publication remains unauthorized.

CG127-004 — all16 staged raw blobs/current task envelope matched M127-003;
compact staged diff has40 removed lines, eliminating M2's3410-line EOL churn.
No unstaged/untracked changes, no other file or behavior, baseline HEAD and
TASK-102 intact. After staging this final append, repeat raw-byte staged gate
and diff-check, then create the one approved scoped commit. No result OID is
embedded here; independent Git readback will bind immutable source/tree.

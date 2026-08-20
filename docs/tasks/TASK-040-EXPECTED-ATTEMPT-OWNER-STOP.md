# TASK-040 — Expected-Attempt Runtime Owner Stop Implementation

## Status

`Completed — Coordinator Accepted` (2026-08-20).

## Task Contract

### Task Mode

`Implementation`. Реализовать только принятый в TASK-039 и запланированный в
Draft DP-010 atomic `StopExpectedAttempt` contract внутри изолированного
`internal/runtimelifecycle`, сохранив без изменений generic `Owner.Stop` и
внешние integration boundaries.

### Why Now

- TASK-039 завершена, Coordinator Accepted и опубликована в trusted baseline;
- её единственная следующая неактивированная рекомендация — bounded
  implementation принятого `StopExpectedAttempt` contract;
- DP-010 уже фиксирует declarations, lifecycle linearization, mismatch и
  cancellation semantics, shared-helper constraint и десять обязательных
  proof groups, поэтому отдельное новое архитектурное решение не требуется;
- текущая реализация имеет только generic `Owner.Stop(ctx)` и не может
  атомарно связать Stop с ожидаемой Owner-issued Launch Attempt identity;
- private exact-scope composition invoker остаётся следующим prerequisite и
  не может корректно закрыть Owner-local TOCTOU до этой реализации.

### Definition of Done

1. `internal/runtimelifecycle` реализует экспортированные DP-010 declarations:
   `ErrInvalidExpectedAttempt`, `StopAttemptMismatch` и
   `(*Owner).StopExpectedAttempt(ctx, expectedAttemptID)` с необходимым GoDoc.
2. Nil Owner и empty expected Attempt ID возвращают exact declared sentinels и
   не изменяют lifecycle state.
3. Под Owner mutex операция после locked cancellation check выбирает active
   attempt раньше retained last; отсутствие relevant attempt и differing
   relevant identity возвращают valid-negative `StopAttemptMismatch` с exact
   optional `AttemptFact`, nil failure и без mutation, attachment,
   cancellation, Host call или wait.
4. Exact active Preparing, Launching, Running и Stopping attempt использует
   неизменённую ordinary Stop phase semantics и exact outcomes; generic
   `Stop(ctx)` и expected-attempt path разделяют один private ordinary-Stop
   helper и не расходятся по lifecycle behavior.
5. Same-ID callers сходятся на одном tracked/retained результате; different-ID
   callers не attach к работе. Old retained A при active successor B никогда
   не выбирается и expected A не останавливает B.
6. Retained active `AttemptStopFailed` возвращает stored `StopFailed` с exact
   error identity без cleanup retry.
7. Matching retained `AttemptStopped` и `AttemptStoppedBeforeRunning`
   replay-ят attempt-specific `StopStopped`; matching historical
   `AttemptPreparationFailed` и `AttemptLaunchFailed` выполняют существующий
   resource-free Failed-to-Stopped transition для exact attempt; impossible
   matching state возвращает `ErrStartConflict` без mutation.
8. Cancellation, видимая в locked check до match/claim/attachment, выигрывает
   без mutation; после linearization caller cancellation прекращает только
   ожидание этого caller, а Owner-owned work продолжает convergence.
9. Generic Stop regression coverage подтверждает неизменность существующих
   `Owner.Stop`, Start/Stop convergence, failure retention и attempt sequencing
   semantics.
10. Lock/lifetime и independent-Owner proofs подтверждают отсутствие resource
    call, cancellation или wait под mutex; focused/full/stress/race, vet,
    formatting, diff, documentation parity/link и exported GoDoc checks дают
    требуемое воспроизводимое evidence либо явно оформленное environment
    limitation по PROCESS-001.
11. TASK-026 остаётся `Blocked by Architecture`; реализация не активирует
    invoker, orchestrator, terminal publication или production wiring и не
    повышает Design Status либо иной документальный статус автоматически.
12. Independent Tester и Reviewer подтверждают десять DP-010 proof groups,
    exact scope и отсутствие нового публичного management behavior до
    Coordinator Acceptance.

### Out of Scope

- concrete private exact-scope composition invoker и orchestration composition;
- изменение `internal/runtimelaunchflow`,
  `internal/runtimeorchestrationcontinuation`, DP-013 management routing или
  public management API;
- DP-014 terminal publication, DP-015 terminalization и core work TASK-026;
- изменение Approved DP-014, DP-015, DP-016 или DP-019 semantics/status;
- persistence, external command storage/schema, recovery, reconciliation,
  reporting, redaction, supervision, HTTP/API, adapters или production wiring;
- retry, timeout, force-stop либо cleanup policy;
- status promotion любого DP, reactivation или acceptance TASK-026;
- stage, commit, push, PR, merge, publication, branch cleanup или mutation
  `main`.

### Verification Plan

- до test edits сохранить этот Existing Coverage Report и сопоставить каждый
  новый test с одним или несколькими десятью proof groups DP-010 §28;
- применить `gofmt` к изменённым Go files и проверить отсутствие formatting
  diff;
- выполнить focused tests пакета `internal/runtimelifecycle`, включая
  targeted proof names и повторные shuffled/stress runs;
- выполнить `go test ./... -count=1` и применимые cross-package regression
  tests;
- выполнить `go test -race` для focused package и релевантные stress checks;
  если race технически недоступен, записать точную причину и результат `PASS
  WITH LIMITATION`, не подменяя его обычным PASS;
- выполнить `go vet ./...`, exported GoDoc inspection и проверить отсутствие
  неожиданных dependency/module changes;
- выполнить `git diff --check`, exact file-set/deletion audit, conflict-marker,
  whitespace, generated-artifact и unexpected-file checks;
- после реализации выполнить PROCESS-002: EN/RU semantic/status parity,
  применимость DP-010 и durable project-state sources, task/design navigation,
  relative links и отсутствие planned-as-implemented wording;
- получить независимые Tester и Reviewer verdicts и повторить применимые
  проверки после rework до Coordinator Acceptance.

## Objective

Добавить минимальную независимо проверяемую Owner-local capability, которая
атомарно останавливает только exact expected Launch Attempt и никогда не
останавливает её successor, без integration либо orchestration expansion.

## Deterministic Selection Evidence

TASK-040 выбрана прямо из `Next Candidate` завершённой TASK-039. С этой
рекомендацией согласованы `.ai/PROJECT_CONTEXT.md`, `spec/current-state.md`,
`spec/decisions.md`, DP-010 EN/RU и MASTER_PLAN EN/RU: первым remaining bounded
prerequisite является implementation уже принятого expected-attempt Owner Stop,
а private exact-scope composition invoker следует после него.

Отклонённые альтернативы:

- resume TASK-026 — она остаётся Blocked до реализации и независимой приёмки
  remaining prerequisites;
- private invoker — следующий по prerequisite order и зависит от atomic Owner
  primitive;
- combined Owner + invoker slice — содержит два независимо поставляемых
  поведения и нарушает bounded selection/Size Guard;
- orchestrator, public management, terminal publication, persistence,
  recovery, reporting или production wiring — более поздние boundaries вне
  первого ready slice;
- новая design task — TASK-039 уже приняла implementable DP-010 contract.

## Scope

Ожидаемый production/test scope ограничен:

- `internal/runtimelifecycle/types.go` — exact result kind и sentinel;
- `internal/runtimelifecycle/owner.go` — method и shared private Stop helper;
- `internal/runtimelifecycle/owner_test.go` — десять proof groups и generic
  regression coverage.

Evidence может обосновать меньший file set. Расширение за эти три Go files до
Architecture Confirmation и Coordinator scope reassessment запрещено. После
implementation допускаются только task record и документы, строго необходимые
PROCESS-002; их exact set должен быть доказан applicability и deletion test.

## Non-Goals

- не создавать второй lifecycle или orchestration contract;
- не добавлять management command, adapter, registry либо service locator;
- не менять generic Stop observable semantics;
- не представлять isolated primitive как production integration;
- не начинать автоматически следующий candidate.

## Sources of Truth

- `docs/engineering/PROCESS-001-AI-DEVELOPMENT-WORKFLOW.md`;
- `docs/engineering/PROCESS-002-DOCUMENTATION-SYNCHRONIZATION.md`;
- Approved ADR-0003 и Frozen ARCH-002 lifecycle ownership boundaries;
- Draft DP-010 EN/RU, особенно sections 6, 7, 18.1, 22, 23, 25, 28 и 31;
- TASK-039 Architecture Handoff, Closure и Next Candidate;
- Approved DP-016 и DP-019 и Draft DP-020 только как ordering/non-expansion
  constraints;
- фактические `internal/runtimelifecycle` code/tests;
- `.ai/PROJECT_CONTEXT.md`, `spec/README.md`, `spec/current-state.md`,
  `spec/decisions.md`, task index, design indexes и MASTER_PLAN EN/RU.

## Roles

- Coordinator: trusted-baseline/selection gates, Size Guard, exact scope,
  acceptance и next recommendation;
- Documentation Agent: этот intake как первый content change, documentation
  baseline и последующий PROCESS-002 sync только утверждённых/реализованных
  фактов;
- Architect: explicit confirmation существующего DP-010 contract и
  implementation constraints до Developer work;
- Developer: реализация только подтверждённого three-file-or-less Go scope;
- Tester: Existing Coverage gate, ten-group proof mapping, risk-based
  verification и ограничения среды;
- Reviewer: независимая проверка code, tests, architecture, scope и docs;
- Publisher: не применяется без отдельных commit/publication gates.

## Constraints

- Repository First; chat history не заменяет указанные sources;
- implementation не принимает новых архитектурных решений и останавливается
  при невозможности реализовать DP-010 exact semantics;
- mutex освобождается перед cancellation, `runtime.Launch`, `Host.Stop`,
  callback, external storage, Flow/I/O и любым wait;
- expected path не реализуется как `Observe()` затем `Stop()`;
- exported declarations имеют GoDoc и proof через public package behavior;
- no dependency/module/package expansion;
- unrelated refactoring и formatting-only files запрещены;
- commit и publication запрещены до соответствующих exact user commands.

## Branch and Trusted Baseline

- trusted baseline: clean synchronized
  `main@d8ea6107720cd764e38d10c92216563835609ca5`;
- task branch: `feature/task-040-expected-attempt-owner-stop`;
- branch создан безопасно от trusted baseline;
- этот task record обязан быть и является первым content change;
- stage, commit, push, merge, branch deletion и mutation `main` не разрешены.

## Size Guard

Initial implementation slice содержит одно независимо поставляемое поведение,
один существующий package, ожидаемо не более трёх Go files и без нового
архитектурного контракта. Порог reassessment наступает при любом из условий:

- требуется более 15 total files после обязательного PROCESS-002 sync;
- production diff приближается к 500 lines;
- требуется новый package, dependency или module change;
- требуется более одного нового contract/behavior;
- требуется менять Approved/Frozen semantics либо публичный management API.

При threshold Coordinator либо доказывает indivisibility через DoD и
verification, либо делит работу до дальнейших изменений.

PROCESS-002 inventory расширяет total expected scope до **21 files**: exact
three-file Go implementation/proofs плюс 18 required documents, которые
иначе противоречиво называли extension Planned/absent или не фиксировали
active TASK-040. Size Guard поэтому triggered. Final Coordinator decision:
**`DO NOT SPLIT`** — это одно independently deliverable behavior и
его обязательная EN/RU/durable-state synchronization, без нового package,
dependency, architecture contract или второго behavior. Разделение оставило
бы live sources в противоречии с executable declarations. Final deletion test
и Coordinator Scope Audit подтверждают решение.

## Existing Coverage Report — Before Test Edits

### Existing Coverage

Текущий `owner_test.go` уже доказывает generic Stop behavior:

- success convergence, idempotent repeated Stop и освобождение ownership;
- concurrent Stop с ровно одним Host Stop;
- Stop в Preparing, Launching, Running и Stopping/concurrent attachment paths;
- before-Running success/failure и сохранение Start outcome;
- Failed-without-Host resource-free transition;
- retained Stop failure с exact error и без retry;
- locked pre-claim cancellation и post-claim waiter-only cancellation;
- attempt sequencing/reuse, foreign preparations, independent Owners;
- Runtime/Host calls вне Owner mutex и отсутствие Snapshot retention.

### Coverage Gap

В `types.go` отсутствуют `StopAttemptMismatch` и
`ErrInvalidExpectedAttempt`; в `owner.go` отсутствует
`StopExpectedAttempt`. Ни один executable test не может атомарно подтвердить
expected-ID validation, relevant-attempt mismatch, active-before-last successor
selection, same-ID/different-ID attachment isolation или exact retained
historical behavior expected path. Generic `Observe()` + `Stop()` остаётся
TOCTOU и не закрывает этот gap.

### Added Proof Tests

`Added and independently verified`: 390 test lines in
`internal/runtimelifecycle/owner_test.go` cover all ten DP-010 extension proof
groups through focused validation/mismatch, active-before-last, active phase,
retained outcome, cancellation and lock/Owner-isolation tests.

### Added Regression Tests

Existing generic `TestStop*` coverage is rerun together with all
`TestStopExpectedAttempt*` tests. Focused, stress/shuffled and full regression
runs pass; no generic Stop semantic change is reported.

### Remaining Limitations

Даже после TASK-040 изолированный Owner primitive не создаёт private
composition invoker, DP-014/DP-015 terminal integration, orchestrator или
production wiring. TASK-026 останется Blocked.

## Ten Proof Groups

1. Sentinel validation: nil Owner и empty expected ID, zero mutation.
2. Mismatch: no relevant/different relevant attempt, exact optional fact, nil
   failure, zero attachment/cancellation/Host call/wait.
3. Active-before-last: old A/new active B successor race never stops B for A.
4. Exact active phases: Preparing, Launching, Running, Stopping preserve
   ordinary Stop results.
5. Convergence isolation: same ID attaches/converges; different ID never does.
6. Retained Stop failure: exact stored error identity, no retry.
7. Retained historical cases: stopped replay and exact failed-to-stopped
   transition.
8. Cancellation linearization: locked-check winner and waiter-only late
   cancellation.
9. Generic Stop regression: shared helper causes no semantic drift.
10. Lock/lifetime/tooling: no work/wait under lock, Owner isolation, race,
    vet, gofmt and exported GoDoc.

## Documentation Baseline

Inventory: TASK-039 and task index; DP-010 EN/RU and design indexes; related
DP-016/DP-019/DP-020 status references; `.ai/PROJECT_CONTEXT.md`,
`spec/current-state.md`, `spec/decisions.md` and MASTER_PLAN EN/RU; current
Owner code/tests.

At intake the sources agree: base DP-010 Owner is implemented in isolation;
expected-attempt declarations/semantics are accepted but Planned and absent
from production code/tests; TASK-026 remains Blocked; private exact-scope
invoker follows this implementation. EN/RU DP-010 planned declarations and ten
proof groups are semantically aligned. No critical drift blocking intake is
identified. This record alone changes active-task operational state; no
product capability is claimed yet.

## Architecture Confirmation

**`PASS`**. Architect confirms that TASK-039 and Draft DP-010 provide a
sufficient, implementable and unchanged contract with no conflict against an
Approved ADR/DP or Active/Frozen architecture source. No new architecture
decision or status promotion is required.

Implementation handoff:

- add the exact exported declarations `ErrInvalidExpectedAttempt`,
  `StopAttemptMismatch StopOutcomeKind = "attempt-mismatch"` and
  `(*Owner).StopExpectedAttempt(ctx context.Context, expectedAttemptID
  runtimeconfigload.LaunchAttemptID) (StopOutcome, error)` with applicable
  GoDoc;
- validation order is nil Owner -> `ErrInvalidOwner`, then empty expected ID
  -> `ErrInvalidExpectedAttempt`, then the locked `ctx.Err()` check; every
  rejected validation/cancellation path performs zero lifecycle mutation;
- generic `Stop(ctx)` and `StopExpectedAttempt` must use one shared private
  ordinary-Stop helper, preserving the current phase state machine and keeping
  all cancellation, Host work and waits outside the Owner mutex;
- while holding the mutex, the expected path selects active attempt first,
  otherwise retained last, otherwise none; an old retained A is never selected
  while active successor B exists;
- no relevant attempt or differing relevant ID returns
  `StopAttemptMismatch` with nil method error, absent failure and the exact
  optional relevant immutable `AttemptFact`; it performs zero mutation,
  attachment, cancellation, Host call or wait;
- exact active Preparing, Launching, Running and Stopping cases retain ordinary
  Stop claim/attachment/convergence behavior; same-ID callers converge and a
  different ID never attaches;
- exact retained active `AttemptStopFailed` returns the stored `StopFailed`
  including exact error identity and never retries cleanup;
- exact retained `AttemptStopped` and `AttemptStoppedBeforeRunning` replay
  attempt-specific `StopStopped`; exact retained `AttemptPreparationFailed`
  and `AttemptLaunchFailed` perform only the existing resource-free
  Failed-to-Stopped transition; an impossible exact retained state returns
  `ErrStartConflict` without mutation;
- the locked match/claim/attachment is the cancellation linearization point:
  cancellation visible at the locked check wins without work; later caller
  cancellation releases only that waiter while Owner-owned convergence
  continues;
- existing generic `Owner.Stop`, Start outcomes, DP-013 public management
  behavior and all integration boundaries remain compatible and unchanged;
- production/test scope is at most `internal/runtimelifecycle/types.go`,
  `internal/runtimelifecycle/owner.go` and
  `internal/runtimelifecycle/owner_test.go`; evidence may require fewer files,
  while any expansion requires Coordinator reassessment;
- tests must map explicitly to all ten proof groups in this record and DP-010
  §28, including generic Stop regression and lock/lifetime/tooling proofs;
- TASK-026 remains `Blocked by Architecture`; private exact-scope invoker,
  orchestrator, terminal publication and production wiring remain outside
  TASK-040 and unactivated.

The Existing Coverage Report was recorded before any test edit. Architecture
Confirmation PASS therefore authorizes Developer production edits and Tester/
Developer test edits only within the confirmed three-Go-file-or-less scope. No
implementation, test addition or verification result is claimed by this
handoff.

## Developer Handoff

- exact production scope: `internal/runtimelifecycle/types.go` +6,
  `internal/runtimelifecycle/owner.go` +66; total **+72 production lines**;
- exact test scope: `internal/runtimelifecycle/owner_test.go` **+390 test
  lines**;
- added `ErrInvalidExpectedAttempt`, `StopAttemptMismatch` and documented
  `StopExpectedAttempt` declarations;
- extracted one private `stopLocked` helper shared by generic and
  expected-attempt Stop;
- expected path validates nil Owner, empty ID and locked cancellation, selects
  active before retained last, returns zero-work mismatch, and preserves exact
  active/retained ordinary Stop behavior;
- no package, dependency, module, public management, invoker, integration,
  persistence or wiring change;
- implementation is isolated on the TASK-040 branch; final documentation,
  scope gate and Coordinator Acceptance subsequently completed.

## Tester Handoff

Verdict: **`PASS WITH ENVIRONMENT LIMITATION`**, blocking findings 0,
non-blocking findings 0.

Exact evidence:

- `gofmt -d internal/runtimelifecycle/types.go internal/runtimelifecycle/owner.go internal/runtimelifecycle/owner_test.go`
  — PASS, no output;
- `go test ./internal/runtimelifecycle -run 'TestStopExpectedAttempt|TestStop' -count=1 -v`
  — PASS;
- `go test ./internal/runtimelifecycle -run 'TestStopExpectedAttempt|TestStop' -count=100 -shuffle=on`
  — PASS in 2.610s;
- `go test ./... -count=1` — PASS;
- `go vet ./...` — PASS;
- `git diff --check` — PASS;
- exported GoDoc for all three declarations — PASS;
- repository relative links at Tester handoff — 922 valid / 0 broken.

Race evidence and limitation:

- default `go test -race ./internal/runtimelifecycle -count=1` fails before
  tests with `go: -race requires cgo; enable cgo by setting CGO_ENABLED=1`;
- retry with `CGO_ENABLED=1` fails during race runtime/testmain resolution with
  `runtime/race: package testmain: cannot find package`;
- environment: `GOOS=windows`, `GOARCH=amd64`, default `CGO_ENABLED=0`,
  `CC=gcc`, while neither `gcc` nor `clang` is installed;
- focused count-100 shuffled stress and full regression are successful
  substitutes, but do not relabel unavailable race as PASS.

Proof mapping:

1. validation and no mutation —
   `TestStopExpectedAttemptValidationAndMismatchDoNotMutate`;
2. no/different relevant attempt mismatch, optional fact and zero work — same
   validation/mismatch test plus different-ID active-phase assertion;
3. active-before-last successor race —
   `TestStopExpectedAttemptSelectsActiveBeforeRetainedLast`;
4. exact Preparing/Launching/Running/Stopping semantics —
   `TestStopExpectedAttemptPreservesActivePhaseSemantics`;
5. same-ID convergence/different-ID non-attachment — Running/Stopping
   convergence subtest;
6. retained exact Stop failure/no retry —
   `TestStopExpectedAttemptRetainedOutcomes/Stop_failure`;
7. retained stopped/failure/impossible forms — remaining retained-outcome
   subtests;
8. locked-check/late-waiter cancellation —
   `TestStopExpectedAttemptCancellationLinearization`;
9. generic shared-helper compatibility — focused `TestStopExpectedAttempt|TestStop`
   runs and full regression;
10. outside-mutex work, independent Owners and tooling —
    `TestStopExpectedAttemptRunsWorkOutsideMutexAndOwnersRemainIndependent`,
    stress, vet, gofmt and GoDoc; race has the environment limitation above.

## Initial Independent Review

Verdict before PROCESS-002: **`APPROVED`**, blocking findings 0, non-blocking
findings 0. Reviewer additionally ran
`go test ./internal/runtimelifecycle -count=20 -shuffle=on` — PASS in 2.093s,
and `git diff --check` — PASS. Pre-documentation deletion audit classified the
task record and three Go files as **4 Required / 0 Questionable / 0
Removable**. This verdict does not replace final post-PROCESS-002 review.

## Final Independent Review — Rework B-001

- initial final-review verdict: **`Needs Revision`**, one blocking finding
  B-001, non-blocking findings 0;
- exact finding: mirrored DP-010 §6 still said TASK-039 added explicitly marked
  planned declarations, and §7 still called `StopExpectedAttempt` planned,
  contradicting the same documents' implemented TASK-040 status;
- resolution: DP-010 EN/RU now state that TASK-039 designed and TASK-040
  implements the expected-attempt declarations; the nil-Owner sentence calls
  the method implemented. No code, behavior, design status or other document
  changed in this rework;
- reverified DP-010 parity: headings 35/35, fences 6/6 and Go declaration lines
  67/67; stale exact-extension `planned` wording 0; repository links 923/0;
  `git diff --check` PASS with only LF-to-CRLF working-copy warnings;
- repeat final independent review: **`APPROVED`**, blocking findings 0,
  non-blocking findings 0; B-001 confirmed resolved.

## Process Health Review — TASK-031–TASK-040 Cadence

Ten-task cadence trigger is reached for TASK-031 through TASK-040.

- Questionable/Removable scope findings observed in accepted task audits and
  current preliminary audit: 0/0; no recurring removable work identified;
- defects found after primary verification in this slice: 0; earlier rework in
  the cadence remained bounded inside its originating task and produced no
  known defect in `main`;
- CI failures and post-merge fixes: none recorded for this cadence; repository
  publication has repeatedly reported `No CI`, not a failed registered check;
- unavailable checks: race remains repeatedly unavailable on this Windows
  environment because default CGO is disabled and no C compiler is installed;
  tasks consistently preserve `PASS WITH ENVIRONMENT LIMITATION` and use
  focused stress/full regression substitutes;
- repeated source of documentation work: live prerequisite/status paragraphs
  across DP-016/019/020 and durable state must change together when a planned
  prerequisite becomes isolated implementation; this is mandatory PROCESS-002
  synchronization, not removable scope.

Bounded finding: the persistent local race-toolchain limitation remains an
environment capability gap, already truthfully handled by Verification Matrix
policy. Final Process Health result: no repository process change is justified
by current evidence.

## Final Scope Audit

Final exact total: **21 Required / 0 Questionable / 0 Removable**.

- 3 Go files: exact declarations, shared Owner implementation and ten-group
  proof/regression tests;
- TASK-040 record and task index: completed contract/evidence and navigation;
- DP-010 EN/RU plus design indexes: executable declarations/proofs and mirrored
  implementation-status boundary;
- DP-016/DP-019/DP-020 EN/RU: narrow removal of live absent-extension wording
  while preserving their Approved/Approved/Draft Design Status and Planned
  overall implementation boundaries;
- TASK-026: live blocker now names completed TASK-040 and the later invoker,
  without changing `Blocked by Architecture`;
- `.ai/PROJECT_CONTEXT.md`, `spec/current-state.md`, `spec/decisions.md` and
  MASTER_PLAN EN/RU: durable completed-task and implemented/planned boundary.

Deletion mapping: removing a Go file loses declaration, behavior or proof;
removing a DP mirror/index breaks semantic parity or leaves the extension
Planned; removing a related live DP/TASK-026 paragraph restores a false absent
prerequisite; removing a durable-state/task navigation update loses completed
TASK-040 truth. Final diff classification, exact counts and independent review
pass.

## Stop Conditions

- Architecture Confirmation is not explicit or reports unresolved ambiguity;
- exact comparison cannot be linearized under the existing Owner mutex;
- implementation would require `Observe()` + `Stop()` or resource work/wait
  under lock;
- shared-helper extraction would change generic Stop semantics;
- retained terminal/concurrent behavior requires a new design decision;
- scope expands beyond expected three Go files before reassessment, or into
  invoker/orchestrator/public management/persistence/wiring/status promotion;
- source conflict, critical documentation drift, failed mandatory check,
  blocking review finding или unexplained worktree change appears.

## PROCESS-002 Applicability

- this task record and task index: **Required** for completed contract, evidence
  and navigation;
- DP-010 EN/RU and design indexes: **Required** to remove planned-only comments
  and report executable isolated declarations/proofs while retaining Design
  Status Draft and absent production integration;
- DP-016/DP-019/DP-020 EN/RU: **Required** for narrow live prerequisite wording;
  their Design Status remains Approved/Approved/Draft and their overall
  orchestrator/readiness implementation status remains Planned;
- active TASK-026: **Required** to retain Blocked status while replacing the
  stale absent-extension prerequisite with completed TASK-040 and the later
  invoker;
- `.ai/PROJECT_CONTEXT.md`, `spec/current-state.md`, `spec/decisions.md` and
  MASTER_PLAN EN/RU: **Required** for completed TASK-040, isolated
  implementation, final gates and unactivated next candidate;
- root README and `CHANGELOG.md`: **Not applicable**; no user-facing, release,
  root-level or production-integrated capability changed;
- ADR/ARCH: **Not applicable**; no Approved/Frozen semantics or status changed;
- other DP, task, roadmap and review documents: **Not applicable**; they do not
  contain a live contradiction requiring task scope;
- dependencies/modules/generated artifacts: **Not applicable**; none changed.

Final Documentation Agent PROCESS-002 result: **`Synchronized`**. Repeat final
Reviewer APPROVED 0/0, final Scope Audit and Coordinator Acceptance completed;
no design-status promotion is claimed.

## Documentation Verification

- EN/RU structure: DP-010 headings 35/35, fences 6/6; DP-016 30/30 and 4/4;
  DP-019 25/25 and 16/16; DP-020 34/34 and 12/12; design indexes 1/1 and 0/0;
  MASTER_PLAN 36/36 and 0/0 — PASS;
- semantic/status parity: DP-010 remains Draft with base and expected-attempt
  extension implemented in isolation; DP-016/DP-019 remain Approved and
  Planned overall; DP-020 remains Draft and Planned overall — PASS;
- links: repository-wide 923 valid / 0 broken — PASS;
- stale live absent/planned-only wording for the exact expected-attempt
  extension in synchronized sources: 0 after final-review B-001 rework and
  repeat stale sweep; TASK-039 closure-only history remains explicitly
  historical — PASS;
- private invoker/integration/wiring absent, next candidate unactivated,
  TASK-026 Blocked, TASK-040 Completed and Coordinator Accepted — PASS;
- conflict-marker and trailing-whitespace inspection plus `git diff --check` —
  PASS, with only Git LF-to-CRLF working-copy warnings;
- exact worktree file set: 21 files, consisting of 3 Go files and 18 Required
  documentation files; unexpected files 0.

## Commit and Publication Gate

- commit permission for TASK-040: **not received**;
- publish permission for TASK-040: **not received**;
- stage, commit, push, PR, merge, publication и cleanup запрещены;
- bare `Продолжай проект.` authorizes task execution only through acceptance
  and then requires STOP.

## Next Candidate

После реализации и независимой приёмки TASK-040 следующий рекомендуемый
candidate — отдельный private exact-scope composition invoker. Он не активирован
этим record и не входит в TASK-040. TASK-026 terminal publication остаётся core
work, а не частью candidate.

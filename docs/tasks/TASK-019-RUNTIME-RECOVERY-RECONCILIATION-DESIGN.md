# TASK-019 — Runtime Recovery and Reconciliation Design

## Status

`Completed — Coordinator Accepted`.

## Task Contract

### Task Mode

`Design-only`. Задача определяет один focused recovery/reconciliation contract
после termination или restart Control Service для prerequisite ARCH-004 §19(5)
без recovery implementation, persistence adapter, API или production wiring.

### Why Now

- TASK-018, `.ai/PROJECT_CONTEXT.md`, `spec/current-state.md`,
  `spec/decisions.md` и зеркальные MASTER_PLAN независимо рекомендуют этот
  slice следующим;
- Active ARCH-004 §19(5) является следующим обязательным prerequisite после
  candidate contracts persistence §19(2), command idempotency §19(3) и
  activation/replacement/rollback §19(4);
- Draft DP-014 сохраняет last-confirmed lifecycle facts, Draft DP-015 закрывает
  command admission при потере permit, а Draft DP-016 оставляет linked command
  set unresolved после потери process-local orchestration truth;
- отдельный recovery contract является наименьшим независимо проверяемым
  slice и не смешивает operational reporting/redaction §19(6).

### Definition of Done

1. Зеркальный non-normative Draft DP-017 EN/RU определяет technology-neutral
   recovery и reconciliation одного Runtime Instance после потери Control
   Service process-local ownership.
2. Contract разделяет durable last-confirmed facts, current execution evidence
   и process-local capabilities; сохранённый `Running`, PID, socket или address
   по отдельности не доказывают live Host ownership.
3. Recovery fail-closed классифицирует exact aggregate, active attempt,
   primitive/parent/phase command set и доступное execution evidence, затем
   публикует только доказуемый reconciled outcome.
4. Unresolved barrier открывается только после terminal resolution всех
   связанных command/lifecycle facts; blind retry, duplicate delegation,
   Launch Attempt reuse, hidden adoption и automatic restart запрещены.
5. Formal gates §19(2)–(5), последующий §19(6), planned/implemented truth,
   зеркальные документы, PROCESS-002, Scope Audit и independent final review
   завершены без blocking findings.

### Out of Scope

- production-код, tests, package, repository, schema, migrations и adapters;
- HTTP paths, DTO, status codes, SDK behavior и concrete authorization policy;
- approval/status decisions DP-014, DP-015, DP-016 или DP-017;
- process supervision, external worker protocol, PID registry и socket probing
  implementation;
- automatic restart, rollback, retry, backoff, scheduling или policy engine;
- operational reporting/redaction §19(6), retention и Production Activation;
- multi-node placement, failover, clustering и remote ownership transfer.

### Verification Plan

- проверить precedence Active ARCH-004 и boundaries Draft DP-013–DP-016;
- проверить restart cuts до/после command claim, attempt claim, readiness,
  Stop claim, proven release и terminal publications;
- проверить orphan/no-orphan, stale evidence, unavailable evidence,
  contradictory facts, concurrent recovery и new-command admission;
- проверить отсутствие automatic execution, inferred liveness, API/schema/
  technology и implementation claims;
- сопоставить EN/RU headings, normative meaning, links и code fences;
- выполнить full `go test ./... -count=1`, `go vet ./...`,
  `git diff --check`, conflict-marker и relative-link checks;
- independent Reviewer проверяет ownership, evidence hierarchy, failure matrix,
  scope и removability каждого change.

## Objective

Создать один проверяемый candidate design contract ARCH-004 §19(5), который
после потери Control Service сверяет долговечные operational/command facts с
доступным execution evidence и безопасно разрешает либо сохраняет unresolved
barrier. Draft не снимает formal gates и не разрешает implementation.

## Selection Evidence

- trusted baseline: clean synchronized `main` на merge commit `d083957`;
- active `In Progress` или `Blocked` task отсутствует;
- explicit next candidate: TASK-018 `Next Candidate`;
- corroboration: `.ai/PROJECT_CONTEXT.md`, `spec/current-state.md`,
  `spec/decisions.md`, ARCH-004 §19 и зеркальные MASTER_PLAN;
- ranking: current Beta dependency, следующий unresolved prerequisite §19(5),
  smallest independent design и highest remaining lifecycle-truth risk;
- отклонены:
  - management implementation — Blocked formal §19(2)–(6);
  - status approval DP-014–DP-016 — materially different governance decision;
  - reporting/redaction — последующий §19(6);
  - automatic restart/supervision — не prerequisite initial in-process model;
  - Production Activation — требует всех gates и concrete composition.

## Scope

- новый зеркальный DP-017 EN/RU;
- bounded synchronization DP-011, DP-013, DP-014, DP-015 и DP-016 EN/RU:
  continuation/binding capability, recovery authority, resolved barrier,
  prerequisite execution binding и preserved formal gates;
- design indexes и MASTER_PLAN EN/RU;
- task record/index;
- `.ai/PROJECT_CONTEXT.md`, `spec/current-state.md`, `spec/decisions.md`;
- только необходимые cross-links и status/dependency statements.

## Non-Goals

- не начинать §19(6) или implementation;
- не добавлять recovery operation в public management surface DP-013;
- не менять Active/Frozen architecture, Runtime Host lifecycle или code;
- не принимать stale PID/socket/Running как ownership proof;
- не выполнять unrelated cleanup или formatting.

## Sources of Truth

- ADR-0002, ADR-0003 и ADR-0004;
- Frozen ARCH-002;
- Active ARCH-004 и ARCH-005;
- Draft DP-013, DP-014, DP-015 и DP-016 только как subordinate candidate
  designs;
- DP-010/DP-011 и code/tests как evidence process-local lifecycle behavior;
- TASK-018 и durable project-state documents.

## Roles

- Coordinator: intake, gates, Scope Audit и acceptance;
- Architect: recovery authority, evidence hierarchy, reconciliation outcomes и
  acceptance constraints;
- Documentation Agent: baseline, mirrored DP и project-state synchronization;
- Developer: Not applicable — production code запрещён;
- Tester: documentation verification и full regression checks;
- Reviewer: independent initial/final review; не автор changes;
- Publisher: Not applicable — publication не разрешена этой командой.

## Branch

- исходный baseline: clean `main@d083957`, `main == origin/main`;
- task branch: `docs/task-019-runtime-recovery-reconciliation-design`;
- branch создана до первого content change;
- stage, commit, push, merge, fetch, pull, rebase, reset и branch deletion
  запрещены без соответствующего explicit gate.

## Constraints

- ровно один новый architecture contract;
- single-node, initially in-process и Technology Neutrality;
- durable actual state является last-confirmed fact, не liveness proof;
- Runtime Lifecycle Owner остаётся sole lifecycle/live Host owner; recovery не
  создаёт второй Owner и не переносит process-local permit;
- никакой новый attempt, lifecycle delegation или Load до resolved barrier;
- existing desired/actual states не расширяются;
- commit только после точной команды `Разрешаю коммит.`.

## Stop Conditions

- конфликт с Approved ADR или Active/Frozen ARCH;
- необходимость автоматического restart/failover или adoption неизвестного
  Host без доказуемого ownership protocol;
- необходимость включить reporting/redaction §19(6);
- невозможность сохранить API/schema/technology neutrality;
- более одного architecture contract или critical documentation drift;
- materially different Ready solution без authoritative ordering.

## Acceptance Criteria

1. После process loss все surviving Claimed primitive/parent/phase commands и
   non-terminal active attempt рассматриваются как unresolved до coherent
   exact-scope reconciliation.
2. Recovery использует одно per-Instance conditional authority и не допускает
   конкурирующих recovery owners, lifecycle commands или stale publication.
3. Persisted `Running`, PID, Listener address, elapsed time или command Claim
   отдельно не доказывают живой Host, завершённую mutation или право adoption.
4. Доказанное отсутствие Host resources позволяет только phase-sensitive
   Failed/interrupted reconciliation; `Stopped`/stopped-satisfied требует exact
   proof completion Host-owned shutdown contract. Unknown сохраняет barrier.
5. Definitively absent execution binding после Owner claim разрешает
   resource-free Failed для уже Starting attempt; indeterminate binding
   сохраняет active association и запрещает Load.
6. Доказанная live execution может быть только наблюдаемым evidence в рамках
   будущего approved execution adapter protocol; текущий in-process restart не
   фабрикует Owner и не объявляет Runtime управляемым.
7. Command set terminalizes только из exact aggregate/attempt/evidence facts;
   contradictory, unavailable или indeterminate evidence сохраняет unresolved
   status и выполняет zero new lifecycle mutation.
8. Recovery не создаёт new attempt, не повторяет Start/Stop, не выбирает target,
   не выполняет rollback/restart и не открывает admission частично.
9. EN/RU parity, formal gates, navigation, project state и verification
   подтверждены evidence.

## Existing Coverage Report

- Existing Coverage: ARCH-004 запрещает liveness inference из stale PID;
  DP-014 хранит coherent last-confirmed aggregate facts; DP-015 определяет
  unresolved command barrier; DP-016 перечисляет process-loss cut points.
- Coverage Gap: recovery authority, evidence hierarchy, restart classification,
  phase-sensitive reconciliation outcomes и exact barrier reopening отсутствуют.
- Added Proof Tests: Not applicable до future implementation.
- Added Regression Tests: Not applicable — code/test changes запрещены.
- Remaining Limitations: executable failure/restart proofs появятся после
  normative status decisions и отдельной implementation task.

## Size Guard

- expected scope: 21 documentation files, 0 production lines, 0 packages,
  1 new architecture contract, 0 shipped behaviors;
- trigger `>15 files` сработал и scope переоценён;
- split отвергнут: пять predecessor DP mirrors содержат непосредственные
  readiness, unresolved/recovery и ordering boundaries, а indexes, roadmap,
  task и durable state обязаны синхронно отражать новый candidate contract;
- 21-file slice остаётся одним independently reviewable behavior contract,
  каждый файл будет повторно проверен Scope Audit.

## Documentation Baseline

- Status: `Synchronized with expected task-stage drift`; critical drift 0.
- Factual boundary: process-local Lifecycle Owner/Flow/Source реализованы только
  изолированно; management routing, persistence, idempotency, activation и
  recovery существуют только как planned designs либо отсутствуют.
- Expected drift: task index и durable project-state документы ещё не отражают
  active TASK-019; disposition — Required synchronization в этой task.
- Critical conflicts с ADR, Frozen ARCH-002 или Active ARCH-004: отсутствуют.
- Exact expected final file set: 21 documentation files — task record/index,
  DP-017 EN/RU, bounded DP-011/DP-013/DP-014/DP-015/DP-016 EN/RU, design indexes
  EN/RU, MASTER_PLAN EN/RU, `.ai/PROJECT_CONTEXT.md`, `spec/current-state.md` и
  `spec/decisions.md`.

## Architecture Analysis

- Architecture Confirmation: `Approved`; blockers 0;
  initial blockers B-001–B-003, repeat blocker B-004, terminal-repeat blocker
  B-005 и его residual acceptance-proof wording addressed.
- Recovery authority: один durable conditional recovery claim закрывает
  per-Instance admission и выдаёт ровно один non-transferable process-local
  permit только для reconciliation publications, не для lifecycle calls.
- Truth model: durable aggregate/command records являются last-confirmed facts;
  current capabilities не восстанавливаются; execution evidence обязано иметь
  exact Attempt/Generation binding и approved containment semantics.
- Correlation prerequisite: Control Service composition создаёт generation,
  DP-014 владеет conditional durable binding, DP-011 continuation координирует
  binding/pending Stop/final Load gate; только coherent exact no-binding proof
  у still-active attempt/expected revision возвращается Flow как
  `BindingFailed` и converges exact token через existing Owner.Start/
  FailedPreparation, где mutex Owner упорядочивает later Stop. Только Owner
  outcome публикуется durably и terminalizes command.
- Initial topology: proven termination exact in-process generation доказывает
  отсутствие runnable Host этой generation, но не graceful cleanup, Stop или
  readiness; resource absence публикует Failed/interrupted, а Stopped требует
  exact proof Host-owned shutdown completion; adoption/hydration запрещены.
- Reconciliation: aggregate/attempt terminal fact публикуется первым, затем
  exact primitive/phases, затем parent; barrier открывается только после
  coherent fully terminal verification и conditional recovery release.
- Crash safety: cross-store transaction не предполагается; каждый step
  conditional/idempotent, permit loss и indeterminate commit сохраняют barrier
  closed, следующий pass resume из durable truth без lifecycle replay.
- Result boundary: desired/actual state set не расширяется; automatic restart,
  rollback, retry, phase creation, DP-013 invocation и Host adoption отсутствуют.

## Verification Matrix

| Risk | Required proof | Current result |
| --- | --- | --- |
| stale Running/PID becomes false liveness | explicit evidence hierarchy and negative rules EN/RU | PASS |
| attempt exists but execution never began | unbound Starting -> resource-free Failed; indeterminate binding unresolved | PASS after B-001/B-002 rework |
| live binding failure bypasses sole Owner | BindingFailed -> exact-token Owner.Start; Stop race under Owner mutex; durable/command outcome after Owner | PASS after B-004 rework |
| stale/conflicting binding rejection is misclassified as absence | BindingFailed only after coherent exact no-binding proof for still-active attempt/expected revision; otherwise exact terminal convergence or Blocked | PASS after B-005 rework |
| two recovery owners or command race | shared per-Instance claim/admission linearization | PASS |
| recovery crash replays lifecycle | monotonic publication order under closed barrier | PASS |
| partial linked-set resolution opens admission | fully terminal coherent-set release gate | PASS |
| Stop/replacement outcome fabricated | Stopped only from Host-shutdown proof, not resource absence | PASS after B-003 rework |
| EN/RU drift | 29/29 headings, 2/2 fences and semantic review | PASS |
| code regression | `go test ./... -count=1`; `go vet ./...` | PASS |
| broken navigation or malformed diff | changed/new links 0 failures; `git diff --check` | PASS |

## Verification

- `go test ./... -count=1` — PASS, все packages;
- `go vet ./...` — PASS;
- DP-017 heading parity EN/RU — 29/29;
- DP-017 code-fence parity EN/RU — 2/2;
- changed/new relative-link validation — 0 failures;
- `git diff --check` — PASS; platform LF-to-CRLF notices informational;
- production code/tests — unchanged;
- Initial Reviewer — `Needs Revision`, B-001/B-002/B-003 Major, nonblocking 0;
- first bounded rework — выполнен;
- Repeat Reviewer — `Needs Revision`, B-004 Major, nonblocking 0;
- second bounded rework — выполнен;
- Terminal Repeat Reviewer — `Needs Revision`, B-005 Major, nonblocking 0;
- third bounded rework — выполнен;
- Final Reviewer — `Needs Revision`, residual B-005 acceptance-proof wording,
  1 Major / 0 nonblocking; остальные проверки, B-001–B-004 и Scope Audit
  21/0/0 — PASS;
- fourth bounded rework — выполнен;
- Final Confirmation Reviewer — `Approved`, 0 blocking / 0 nonblocking;
- Terminal Closure Reviewer — `Approved`, 0 blocking / 0 nonblocking.

## Scope Audit

- final accepted: 21 Required / 0 Questionable / 0 Removable;
- DP-017 mirrors — core candidate contract;
- DP-011/DP-013–DP-016 mirrors — exact continuation/binding capability,
  readiness, unresolved barrier, recovery link и ordering, без которых
  predecessor truth или implementability drift немедленно возникнет;
- indexes/roadmap/task/state — mandatory navigation, applicability и live
  project-state synchronization;
- terminal removability audit подтверждён independent Reviewer.

## Documentation Sync

- PROCESS-002 applicability: `Required`, так как task design-only и изменяет
  architecture knowledge, roadmap, task/state truth и mirrored navigation;
- synchronized: DP mirrors, design indexes, MASTER_PLAN mirrors, task index,
  `.ai/PROJECT_CONTEXT.md`, `spec/current-state.md`, `spec/decisions.md`;
- CHANGELOG: Not applicable — shipped/user-facing behavior отсутствует;
- final status: `Synchronized`; closure bookkeeping завершён.

## Process Health

- task-before-work: PASS — task branch создана до content, task record был
  первым и единственным initial content change;
- size guard: triggered, reassessed after review, split rejected with exact
  21-file proof;
- commit/publication gates: preserved; staged files 0;
- initial independent review: Needs Revision B-001–B-003; first rework done;
- repeat independent review: Needs Revision B-004; second rework done;
- terminal repeat review: Needs Revision B-005; third rework done;
- final review: Needs Revision, residual B-005 acceptance-proof wording;
  fourth rework done;
- final confirmation review: Approved 0/0;
- terminal closure review: Approved 0/0;
- corrective process change: не требуется на текущем этапе.

## Handoff

- Architect/Documentation work завершена;
- Tester evidence — PASS;
- Final Confirmation и Terminal Closure Reviewer — Approved 0/0;
- Coordinator acceptance и final Scope Audit 21/0/0 завершены;
- следующий рекомендуемый candidate — ARCH-004 §19(6), не активирован.

## Commit Gate

- exact command `Разрешаю коммит.` получена после Coordinator Acceptance: да;
- commit message policy: Conventional Commits;
- selected message: `docs(runtime): define recovery reconciliation`;
- exact accepted file set: 21 documentation file из Scope Audit;
- post-acceptance changes: только bounded closure bookkeeping и этот permission
  record; architecture contract и scope после terminal approvals не изменены;
- temporary/generated/unrelated files: отсутствуют;
- final checks: full Go regression, vet, EN/RU parity, 247/0 links, conflict
  scan, exact file set и `git diff --check` — PASS;
- разрешён ровно один task commit; push/PR/merge/publication не разрешены.

## Publication

- не разрешена текущей командой;
- push/PR/merge и cleanup требуют отдельной точной команды
  `Разрешаю публиковать.` после commit gate.

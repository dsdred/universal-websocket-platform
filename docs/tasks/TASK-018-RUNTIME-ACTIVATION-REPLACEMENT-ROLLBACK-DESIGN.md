# TASK-018 — Runtime Activation, Replacement, and Rollback Design

## Status

`Completed — Coordinator Accepted`.

## Task Contract

### Task Mode

`Design-only`. Задача определяет один focused ordering contract activation,
replacement и rollback Runtime Instance для prerequisite ARCH-004 §19(4) без
management implementation, API, persistence adapter или production wiring.

### Why Now

- TASK-017, `.ai/PROJECT_CONTEXT.md`, `spec/current-state.md`,
  `spec/decisions.md` и зеркальные MASTER_PLAN независимо рекомендуют этот
  slice следующим;
- Active ARCH-004 §19(4) является следующим обязательным prerequisite после
  candidate contracts persistence §19(2) и command idempotency §19(3);
- Draft DP-014 задаёт durable lifecycle publications, а Draft DP-015 — exact
  command intent, at-most-once delegation и unresolved barrier;
- отдельный ordering contract является наименьшим независимо проверяемым
  slice и не смешивает process recovery или operational reporting.

### Definition of Done

1. Зеркальный non-normative Draft DP-016 EN/RU определяет technology-neutral
   ordering initial activation, replacement и explicit rollback одного Runtime
   Instance без изменения Active ARCH-004.
2. Contract сохраняет single active Launch Attempt, exact Published
   ConfigurationVersion pin, Runtime Lifecycle Owner ownership и обязательный
   Stop-to-proven-release перед новым replacement/rollback attempt.
3. Failure и indeterminate outcomes не допускают overlapping Hosts,
   automatic fallback, hidden retry, in-place reload или publication Running
   для stale target.
4. Rollback является отдельным authorized/idempotent intent к exact Published
   version и создаёт новый Launch Attempt; previous/latest не выводится
   автоматически.
5. Formal gates §19(2)–(4), последующие §19(5)–(6), planned/implemented truth,
   зеркальные документы, PROCESS-002, Scope Audit и independent final review
   завершены без blocking findings.

### Out of Scope

- production-код, tests, package, repository, schema, migrations и adapters;
- HTTP paths, DTO, status codes, SDK behavior и concrete authorization policy;
- approval/status decisions DP-014, DP-015 или DP-016;
- zero-downtime/overlapping replacement, in-place reload и listener handoff;
- automatic rollback, restart, retry, backoff, scheduling или policy engine;
- process recovery/reconciliation §19(5), reporting/redaction §19(6);
- multi-node placement, workers, clustering, supervision и Production
  Activation.

### Verification Plan

- проверить precedence Active ARCH-004 и boundaries Draft DP-013–DP-015;
- проверить initial activation, running same-version no-op, replacement
  Stop-before-Start, Stop-during-Starting, stop failure, target startup failure,
  explicit rollback и indeterminate outcomes;
- проверить отсутствие новых desired/actual states, API/schema/technology и
  implementation claims;
- сопоставить EN/RU headings, normative meaning, links и code fences;
- выполнить full `go test ./... -count=1`, `go vet ./...`,
  `git diff --check`, conflict-marker и relative-link checks;
- independent Reviewer проверяет ownership, ordering, failure matrix, scope и
  removability каждого change.

## Objective

Создать один проверяемый candidate design contract ARCH-004 §19(4), который
определяет safe single-node activation/replacement/rollback ordering поверх
существующих planned persistence и idempotency boundaries. Draft не снимает
formal gates и не разрешает implementation.

## Selection Evidence

- trusted baseline: clean synchronized `main` на merge commit `2f10442`;
- active `In Progress` или `Blocked` task отсутствует;
- explicit next candidate: TASK-017 `Next Candidate`;
- corroboration: `.ai/PROJECT_CONTEXT.md`, `spec/current-state.md`,
  `spec/decisions.md`, DP-013–DP-015 и зеркальные MASTER_PLAN;
- ranking: current Beta dependency, следующий unresolved prerequisite §19(4),
  smallest independent design, lowest unresolved risk и first authoritative
  ordering;
- отклонены:
  - management implementation — Blocked formal §19(2)–(6);
  - status approval DP-014/DP-015 — materially different governance decision;
  - recovery/reconciliation — последующий §19(5);
  - reporting/redaction — последующий §19(6);
  - Production Activation — требует всех gates и concrete composition.

## Scope

- новый зеркальный DP-016 EN/RU;
- bounded synchronization DP-011, DP-013, DP-014 и DP-015 EN/RU: private
  Start-claim continuation, candidate §19(4), preserved formal gates и next
  §19(5) recommendation;
- design indexes и MASTER_PLAN EN/RU;
- task record/index;
- `.ai/PROJECT_CONTEXT.md`, `spec/current-state.md`, `spec/decisions.md`;
- только необходимые cross-links и status/dependency statements.

## Non-Goals

- не начинать §19(5), §19(6) или implementation;
- не добавлять Replace/Rollback public operation в exact surface DP-013;
- не менять Active/Frozen architecture, Runtime Host lifecycle или code;
- не выполнять unrelated cleanup или formatting.

## Sources of Truth

- ADR-0002, ADR-0003 и ADR-0004;
- Frozen ARCH-002;
- Active ARCH-004 и ARCH-005;
- Draft DP-013, DP-014 и DP-015 только как subordinate candidate designs;
- DP-010/DP-011 и code/tests как evidence process-local lifecycle behavior;
- TASK-017 и durable project-state documents.

## Roles

- Coordinator: intake, gates, Scope Audit и acceptance;
- Architect: ordering, ownership, failure boundaries и acceptance constraints;
- Documentation Agent: baseline, mirrored DP и project-state synchronization;
- Developer: Not applicable — production code запрещён;
- Tester: documentation verification и full regression checks;
- Reviewer: independent initial/final review; не автор changes;
- Publisher: Not applicable — publication не разрешена этой командой.

## Branch

- исходный baseline: clean `main@2f10442`, `main == origin/main`;
- task branch:
  `docs/task-018-runtime-activation-replacement-rollback-design`;
- branch создана до первого content change;
- stage, commit, push, merge, fetch, pull, rebase, reset и branch deletion
  запрещены без соответствующего explicit gate.

## Constraints

- ровно один новый architecture contract;
- single-node, in-process и Technology Neutrality;
- replacement/rollback никогда не имеют два owned Host одновременно;
- новый attempt начинается только после confirmed release предыдущего Host;
- rollback всегда explicit и exact-version; automatic previous/latest forbidden;
- existing desired `Stopped|Running` и actual states не расширяются;
- Runtime Lifecycle Owner остаётся sole lifecycle/live Host owner;
- commit только после точной команды `Разрешаю коммит.`.

## Stop Conditions

- конфликт с Approved ADR или Active/Frozen ARCH;
- необходимость zero-downtime overlap или нового operational state;
- необходимость включить recovery/reporting §19(5)–(6);
- невозможность сохранить API/schema/technology neutrality;
- более одного architecture contract или critical documentation drift;
- materially different Ready solution без authoritative ordering.

## Acceptance Criteria

1. Initial activation из отсутствия active Host создаёт не более одного exact
   version-pinned Launch Attempt и публикует Running только после readiness.
2. Replacement/rollback сохраняет один Runtime Instance identity и выполняет
   `Stop old -> proven release -> claim new attempt -> Start new`; overlap и
   reuse attempt identity запрещены.
3. Running на exact target version является zero-mutation satisfied outcome;
   conflicting target требует ordered replacement, не in-place change.
4. Stop failure или unproven cleanup сохраняет active attempt и запрещает new
   attempt; target startup failure не resurrect old Host и не запускает
   automatic rollback.
5. Explicit rollback выбирает exact Published version, использует новый
   command/attempt identity и проходит те же authorization/idempotency/order
   rules, что replacement.
6. Caller cancellation, concurrency и indeterminate outcomes сохраняют
   truthful desired/actual/command facts и не обходят DP-015 barrier.
7. EN/RU parity, formal gates, navigation, project state и verification
   подтверждены evidence.

## Existing Coverage Report

- Existing Coverage: ARCH-004 определяет identity, single active attempt,
  Stop-during-Starting, replacement as new attempt и отсутствие in-place reload;
  DP-014 задаёт durable publications; DP-015 — command identity/barrier.
- Coverage Gap: exact activation/replacement/rollback ordering, version-target
  truth, failure cut points и explicit rollback semantics не определены.
- Added Proof Tests: Not applicable до future implementation.
- Added Regression Tests: Not applicable — code/test changes запрещены.
- Remaining Limitations: executable ordering/failure/restart proofs появятся
  после normative status decisions и отдельной implementation task.

## Size Guard

- expected scope: 19 documentation files, 0 production lines, 0 packages,
  1 new architecture contract, 0 shipped behaviors;
- trigger `>15 files` сработал и scope переоценён;
- split отвергнут: predecessor DP mirrors содержат current-next/gate statements,
  а DP-011 mirror обязан фиксировать planned claim-continuation extension,
  выявленный обязательным review; omission создаст immediate EN/RU/project-state
  drift или оставит ordering неimplementable;
- 19-file slice остаётся одним independently reviewable behavior contract,
  каждый файл будет повторно проверен Scope Audit.

## Documentation Baseline

- Status: `Synchronized with expected task-stage drift`; critical drift 0.
- EN/RU baseline: design tree 15/15 filenames before DP-016; ARCH-004 и
  DP-011/DP-013/DP-014/DP-015 имеют mirrors и согласованные statuses.
- Factual boundary: Runtime Lifecycle Owner, base Flow и Source реализованы
  только изолированно; private claim continuation, management routing,
  persistence, idempotency и activation ordering существуют только как planned
  designs; package/schema/API/recovery/wiring отсутствуют.
- Expected drift: task index и durable project-state документы ещё не отражали
  active TASK-018; disposition — Required synchronization в этой task.
- Critical conflicts с ADR, Frozen ARCH-002 или Active ARCH-004: отсутствуют.
- Exact expected final file set: 19 documentation files — task record/index,
  DP-016 EN/RU, bounded DP-011/DP-013/DP-014/DP-015 EN/RU, design indexes EN/RU,
  MASTER_PLAN EN/RU, `.ai/PROJECT_CONTEXT.md`, `spec/current-state.md` и
  `spec/decisions.md`.

## Architecture Analysis

- Verdict: `Ready for independent review`; authored blockers 0.
- Target: каждый intent содержит exact Published ConfigurationVersion; latest,
  previous и вывод target из history запрещены.
- Initial activation: отсутствие owned Host позволяет claim ровно одного fresh
  exact-version Launch Attempt; Running публикуется только после readiness.
- Already Running на exact target — satisfied outcome с zero mutation;
  conflicting target требует replacement, а не in-place reload.
- Replacement order: claim command -> Stop/converge old attempt -> prove Host
  release -> publish old terminal/Stopped facts -> Continue gate -> claim linked
  Start phase -> Owner claims fresh attempt -> Start target. Service gap
  разрешён, overlap двух Hosts запрещён.
- Replacement/rollback command является bounded parent orchestration: parent
  permit не вызывает DP-013; durable linked Stop/Start phase claims получают
  собственный non-transferable permit ровно для одного exact invocation. Это
  сохраняет at-most-once DP-015 вместо двух lifecycle calls по одному permit.
- Stop-during-Starting захватывает exact old attempt; новый attempt не может
  быть claimed до proven release. Parent и linked Stop phase атомарно занимают
  единственный tracked-Start Stop exception; independent Stop одновременно не
  допускается.
- Continue gate атомарно упорядочивает concurrent Stop и linked Start-phase
  claim: Stop first запрещает phase и новый attempt; Start phase first даёт
  ровно одному Stop pending tracked-Start exception; после Owner claim private
  continuation направляет Stop к exact attempt до Load.
- Explicit rollback является новым authorized/idempotent exact-version intent
  и fresh Launch Attempt; history не обращается, old Host не воскресает,
  automatic fallback отсутствует.
- Stop failure или unproven cleanup блокирует новый attempt; target startup
  failure оставляет old Host released и не запускает automatic rollback.
- Desired `Stopped|Running` и существующие actual states сохраняются;
  промежуточные/terminal facts публикуются правдиво без нового operational state.
- Cancellation зависит от phase и никогда не отменяет уже доказанные facts,
  не возвращает ownership released Host и не обходит DP-015 barrier.
- Planned private Start-claim continuation DP-011/DP-013 выполняется после sole
  Owner claim и до Load: continuation только сигнализирует original blocked Stop
  claimant; тот сохраняет non-transferable permit, единолично достигает той же
  attempt и возвращает durable outcome. После `Continue` обычный tracked-Start
  Stop достигает already-claimed attempt. Current isolated Flow seam не
  реализует.
- Indeterminate outcome закрывает дальнейший admission по DP-015 до отдельного
  recovery/reconciliation contract §19(5).
- DP-016 остаётся non-normative Draft/Planned; exact public surface DP-013 не
  расширен; gates §19(2)–(4) и downstream §19(5)–(6) сохраняются.

## Verification

- Existing Coverage Report заполнен до changes; production/test changes не
  выполнялись, Added Proof/Regression Tests — Not applicable.
- Verification Matrix: ownership/concurrency/cancellation/failure проверены
  design invariants и независимыми review; race/API/dependency checks не
  применимы к documentation-only diff; repository regression применена.
- Full regression: `go test ./... -count=1` — PASS; `go vet ./...` — PASS.
- Первый final full regression получил transient timeout
  `TestRuntimeProductionCompositionUsesTransactionalDispatcher` при Host Stop;
  targeted repeat — PASS, немедленный full repeat — PASS. Production diff
  отсутствует; устойчивый defect не воспроизведён.
- Structure: DP-011 28/28 headings и 14/14 fences; DP-013 30/30 и 14/14;
  DP-014 26/26 и 4/4; DP-015 28/28 и 4/4; DP-016 29/29 и 4/4;
  MASTER_PLAN 12/12 и 0/0.
- Links: repository relative links 799/0 missing — PASS.
- `git diff --check`: PASS; conflict markers: 0.
- Independent reviews: initial `Needs Revision` B-001/B-002; repeat
  `Needs Revision` B-003; third `Needs Revision` B-004; fourth
  `Needs Revision` B-005; terminal `Approved`, 0 blocking и 0 nonblocking.

## Scope Audit

Accepted: **19 Required / 0 Questionable / 0 Removable**.

- DP-016 EN/RU — Required: candidate ordering contract §19(4).
- DP-011 EN/RU — Required: implementable internal Start-claim continuation и
  current-vs-planned status truth.
- DP-013/DP-014/DP-015 EN/RU — Required: exact routing/permit/persistence
  integration, formal gates и next ordering.
- Design indexes EN/RU и task index — Required: navigation/status discovery.
- MASTER_PLAN EN/RU — Required: durable dependency/implementation gap.
- `.ai/PROJECT_CONTEXT.md`, `spec/current-state.md`, `spec/decisions.md` —
  Required: repository-only continuation, factual gaps и gates.
- Task record — Required: contract, review/rework, Process Health, verification
  и closure evidence.

Production, tests, modules, public API, schema, generated, formatting-only и
unrelated changes отсутствуют. Size Guard: 19 documentation files, 0 production
lines, 0 packages, 1 new architecture contract, 0 shipped behaviors; trigger
`>15 files` сработал, scope reassessed и сохранён по documented rationale.

## Documentation Sync

- PROCESS-002: `Synchronized`.
- Required: task record/index, DP-016 EN/RU, bounded DP-011/DP-013/DP-014/
  DP-015 EN/RU, design indexes EN/RU, MASTER_PLAN EN/RU,
  `.ai/PROJECT_CONTEXT.md`, `spec/current-state.md`, `spec/decisions.md`.
- Not applicable: `CHANGELOG.md` — shipped/release capability отсутствует;
  root/spec README — capability/tree не менялись; ADR/ARCH — authoritative
  status/contract не менялся.
- EN/RU parity, planned/implemented truth, relative links и contradictions:
  PASS.

## Process Health

`Triggered`: TASK-018 возвращена с review более двух раз.

- Scope Audit на текущем exact 19-file diff: Questionable 0, Removable 0;
  terminal confirmation ожидает final review.
- Defects после primary verification: architecture findings B-001–B-005
  обнаружены до acceptance/commit; один local final regression timeout не
  воспроизвёлся на targeted и full repeat. Product/code defect и escaped defect
  в `main` отсутствуют.
- CI failures после local PASS: 0; CI/publication не запускались.
- Post-merge fixes: 0; task не merged.
- Недоступные проверки: 0 применимых; race/performance не применимы к
  documentation-only diff, full Go regression и vet доступны.
- Повторяющийся источник лишней работы: composite lifecycle ordering было
  сначала описано без end-to-end trace permit ownership, call stacks, sole Owner
  claim и cancellation на каждом межконтрактном gate.
- Bounded process finding: future composite lifecycle design intake должен до
  initial review составлять explicit trace `durable claim -> live permit owner
  -> exact call stack -> lifecycle linearization -> cancellation/permit-loss
  outcome` для каждой phase и concurrency exception. Изменение PROCESS-001,
  новый fast path, tooling или пользовательская команда не требуются.

## Handoff

- Task Intake зафиксирован первым content change;
- Documentation Baseline и Architecture Analysis завершены; authored blockers
  0;
- mirrored 19-file design/state diff завершён и независимо reviewed;
- Developer: Not applicable; production code запрещён и не изменён;
- следующий этап: independent architecture review, rework при findings,
  verification, Scope Audit и Coordinator Acceptance — завершён.

### Review History

- Initial Reviewer: `Needs Revision`, blocking B-001/B-002, nonblocking 0.
- B-001: один DP-015 permit не мог легально выполнить Stop и Start parent
  replacement/rollback; rework ввёл parent permit без lifecycle authority и
  отдельные durable linked phase claims/permits.
- B-002: admission DP-015 разрешал только distinct Stop рядом с tracked Start и
  не разрешал Stop рядом с replacement; rework ограничил оба race одним
  атомарным parent/phase exception и Continue gate без обхода unresolved
  barrier.
- Repeat Reviewer: `Needs Revision`, blocking B-003, nonblocking 0. B-001/B-002
  закрыты.
- B-003: combined Start-phase/Launch-Attempt claim нарушал sole Owner claim
  authority DP-010/DP-011 и оставлял cancellation gap. Rework оставил Continue
  gate только между Stop и Start-phase claim; phase permit вызывает обычный
  DP-013 Start, Owner единолично claim attempt, а Stop после победы phase
  получает non-mutating in-progress до Terminal parent.
- Third Reviewer: `Needs Revision`, blocking B-004, nonblocking 0; B-003 закрыт.
- B-004: временная блокировка Stop после Start-phase claim нарушала обязательный
  Stop-during-Starting ARCH-004. Rework добавил planned private Start-claim
  continuation DP-011/DP-013: один Stop может быть pending до Owner claim,
  затем достигает той же attempt до Load; post-continuation Stop использует
  обычный Owner route.
- Fourth Reviewer: `Needs Revision`, blocking B-005, nonblocking 0; B-004 закрыт.
- B-005: continuation не могла transfer non-transferable Stop permit из
  original Stop call stack. Rework оставил permit у synchronously blocked
  claimant: continuation только сигнализирует Owner claim и ждёт
  `Continue|StopConverged|Blocked`; caller cancellation, return/permit loss,
  definitive no-mutation, convergence и indeterminate outcomes определены без
  admission/Owner lock через wait.
- Terminal architecture Reviewer: `Approved`, 0 blocking и 0 nonblocking;
  B-001–B-005 закрыты, deadlock/signal/cancellation audit PASS, removable
  question — 19 Required / 0 Questionable / 0 Removable.
- Initial administrative closure Reviewer: `Needs Revision`, blocking F-001,
  nonblocking 0: top-level PROJECT_CONTEXT/current-state summaries сохраняли
  stale `active` label после Completion. Bounded fix заменил только live summary
  wording на `Completed`; historical activation evidence сохранено.
- Final administrative closure Reviewer: `Approved`, 0 blocking и 0
  nonblocking; F-001 закрыт, architecture unchanged, exact scope/status/gates/
  verification/PROCESS-002/Process Health/commit-publication facts aligned.

### Coordinator Acceptance

- Documentation Agent: exact mirrored 19-file design/state diff complete.
- Developer: Not applicable; production code запрещён и не изменён.
- Tester: PASS; full regression, vet и documentation checks complete.
- Independent terminal Reviewer: `Approved`, 0/0 findings.
- PROCESS-002 `Synchronized`; Scope Audit 19/0/0; Process Health Review complete.
- Coordinator Acceptance granted для exact 19-file documentation diff.
- Commit readiness: только после новой точной команды `Разрешаю коммит.`;
  stage, commit, push и publication не выполнялись.

## Commit Gate

- exact command `Разрешаю коммит.` получена после Coordinator Acceptance: да;
- commit message policy: Conventional Commits;
- selected message:
  `docs(runtime): define activation replacement rollback ordering`;
- exact accepted file set: 19 documentation files из Scope Audit;
- post-acceptance changes: только bounded administrative closure, F-001 fix и
  этот permission record; architecture contract и scope не изменены после
  terminal approvals;
- temporary/generated/unrelated files: отсутствуют;
- final checks: full Go regression после transient repeat, vet, EN/RU parity,
  links 799/0, exact file set и `git diff --check` — PASS;
- разрешён ровно один local task commit; push, PR, merge и publication не
  разрешены.

## Publication

- Coordinator Acceptance получен; publication readiness отсутствует до
  отдельного разрешённого task commit и последующей команды публикации;
- Publisher P0–P10: not authorized.

## Next Candidate

- предварительно: focused recovery/reconciliation после Control Service
  termination ARCH-004 §19(5);
- не активирован.

## Closure

- Final status: `Completed — Coordinator Accepted`;
- Closed by: Coordinator;
- Date: 2026-08-01;
- commit: authorized after Coordinator Acceptance, not yet performed;
- publication: not authorized, not performed.

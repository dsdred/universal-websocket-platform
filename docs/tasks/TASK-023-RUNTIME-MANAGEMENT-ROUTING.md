# TASK-023 — Runtime Management Routing Implementation

## Status

`Completed — Coordinator Accepted`.

## Task Contract

### Task Mode

`Implementation`: реализовать только Ready bounded isolated slice Draft DP-013
без production wiring и без расширения management capability.

### Why Now

- TASK-021 подтвердила закрытие design gates ARCH-004 §19(2)–(6) и Ready
  status DP-013 для bounded isolated implementation;
- TASK-022 устранила обязательный pre-existing root README drift;
- Owner, Loader и Flow уже реализованы и независимо проверены;
- это первый dependency-ready slice текущей milestone и prerequisite будущей
  management integration, которая в этой task не начинается.

### Definition of Done

1. `internal/runtimemanagement` реализует точную public surface и error
   precedence DP-013 для Target, Binding, Directory и Start/Stop/Observe.
2. Construction связывает один Target, Owner и ровно один Flow, отклоняя
   invalid/duplicate bindings до создания Flow.
3. Authorization выполняется ровно один раз до lifecycle delegation; exact
   downstream outcomes и errors сохраняются.
4. Proof и regression tests покрывают validation, routing, authorization,
   concurrency/cancellation и отсутствие cross-scope mutation.
5. Documentation synchronization, Scope Audit, независимый review и все
   применимые проверки завершены без blocking findings.

### Out of Scope

- Control Service, HTTP/API/CLI transport, production composition или
  Production Activation;
- persistence, command idempotency, recovery, reporting/redaction и concrete
  authorization policy;
- DP-016 private Start-claim continuation и изменения Owner, Loader или Flow;
- изменение архитектурного смысла или Design Status DP-013;
- commit, push, PR, merge или publication.

### Verification Plan

- до test changes: inventory существующих Owner/Flow/Loader tests и exact
  DP-013 coverage gaps;
- focused unit/proof tests нового package, включая race-oriented concurrent
  scopes и authorization ordering;
- `gofmt`, focused tests, `go test ./... -count=1`, `go test -race` для нового
  package, `go vet ./...`, link/parity checks и `git diff --check`;
- независимый semantic, API, lifecycle, scope и documentation review.

## Objective

Добавить изолированную process-local boundary, которая после точной validation
и authorization маршрутизирует Start/Stop/Observe ровно к immutable scope
Runtime Instance.

## Selection Evidence

- baseline: clean synchronized `main@200dce8`, равный `origin/main`;
- active tasks отсутствовали; TASK-022 завершена и опубликована через PR #23;
- `spec/decisions.md`, `spec/current-state.md`, PROJECT_CONTEXT и TASK-021/022
  однозначно рекомендуют DP-013 implementation после README correction;
- prerequisites: `runtimelifecycle.Owner`, `configurationloader.Loader` и
  `runtimelaunchflow.Flow` существуют и имеют proof tests;
- отклонены production integration и DP-014–DP-018 implementation как
  dependent/broader slices; product prioritization не требуется.

## Scope

- новый package `internal/runtimemanagement` и его focused tests;
- DP-013 EN/RU implementation-status synchronization без повышения Design
  Status;
- task record/index и обязательные project-state документы;
- только минимальные изменения, необходимые для exact DP-013 surface.

## Non-Goals

- не добавлять dynamic registry, list/register/replace operations или DTO;
- не менять contracts существующих packages;
- не начинать management integration, persistence или Production Activation;
- не выполнять unrelated refactoring.

## Sources of Truth

- Accepted ADR-0002, ADR-0003 и ADR-0004;
- Frozen ARCH-002, Active ARCH-004 и ARCH-005;
- Draft/Ready DP-013 как утверждённый task-level implementation contract;
- DP-010–DP-012 и фактические Owner/Loader/Flow implementations;
- TASK-015, TASK-021 и TASK-022;
- PROCESS-001 и PROCESS-002.

## Roles

- **Coordinator:** intake, gates, Scope Audit, acceptance и closure.
- **Architect:** подтверждение existing DP-013 без нового решения.
- **Documentation Agent:** baseline и synchronized factual status.
- **Developer:** минимальная реализация exact approved task scope.
- **Tester:** Existing Coverage Report, proof/regression tests и verification.
- **Reviewer:** независимый final review после реализации.
- **Publisher:** Not applicable — publication не разрешена.

## Branch

- trusted baseline: clean synchronized `main@200dce8`;
- task branch: `feature/task-023-runtime-management-routing`;
- branch action: создана и активирована до первого content change;
- stage, commit, push, merge, fetch, pull, rebase, reset и branch deletion
  запрещены текущей командой.

## Constraints

- exact exported identifiers и sentinel strings следуют DP-013;
- Directory остаётся immutable и не добавляет lifecycle synchronization;
- Owner остаётся единственной serialization boundary одного Runtime Instance;
- authorization error и downstream outcome/error возвращаются без wrapping;
- generated, temporary, environment-specific и unrelated files запрещены.

## Stop Conditions

- реализация требует изменения higher-status architecture или existing package
  contract;
- невозможно сохранить authorization-before-mutation либо static Owner/Flow
  binding;
- scope расширяется до production wiring, persistence или следующей task;
- baseline/diff становится неатрибутированным;
- обязательная verification или independent review даёт blocking finding.

## Acceptance Criteria

1. Exact identity routing не создаёт identity oracle для mismatch/absence.
2. Invalid/cancelled/denied commands производят zero lifecycle mutation.
3. Start делегируется только Flow; Stop и Observe — только exact Owner.
4. Разные scopes остаются независимыми при concurrent invocation.
5. Изменён только Required scope, а planned integration не заявлена
   implemented.

## Existing Coverage Report

- **Existing Coverage:** Owner tests покрывают lifecycle arbitration,
  cancellation и coherent observation; Flow tests покрывают exact
  PrepareStart/Load/Build/Start delegation и preservation outcomes; Loader
  tests покрывают configuration identity validation.
- **Coverage Gap:** package Directory, Target/Binding validation, duplicate
  rejection, exact lookup, authorization ordering и cross-scope routing ещё
  отсутствуют.
- **Added Proof Tests:** planned focused tests всех DP-013 command и
  construction invariants.
- **Added Regression Tests:** planned denial/cancellation/mismatch tests,
  доказывающие zero mutation и exact error identity.
- **Remaining Limitations:** tests не доказывают отсутствующие transport,
  persistence, recovery, policy или Production Activation.

## Size Guard

- triggered by 23 changed files, после чего scope повторно оценён;
- production scope остаётся одним package, менее 500 production lines, одним
  independently verifiable behavior и не создаёт новый architecture contract;
- 21 documentation files образуют обязательную цельную synchronization set:
  task/navigation, DP-013 и dependent status mirrors, root status, MASTER_PLAN
  mirrors и три project-state sources;
- split оставил бы нормативные status contradictions, поэтому 23-file scope
  принят как один целостный slice; Questionable и Removable отсутствуют.

## Documentation Baseline

- DP-013 EN/RU согласованы как Draft/Planned/Ready и описывают exact target;
- project-state документы отмечают implementation как следующий Ready, но не
  started;
- production capability, application wiring и Production Activation явно
  отсутствуют;
- critical drift после TASK-022 не обнаружен.

## Architecture Confirmation

`Confirmed`: task реализует существующий DP-013 bounded isolated contract и не
изменяет ADR/ARCH, ownership, lifecycle или production composition.

## Verification

- **Tester verdict:** `PASS WITH LIMITATION`, 0 blocking и 0 nonblocking
  findings.
- `go test ./internal/runtimemanagement -count=100`: PASS;
- `go test ./... -count=1`: PASS;
- `go vet ./...`: PASS;
- `go test -race ./internal/runtimemanagement -count=1`: unavailable — default
  `CGO_ENABLED=0`; при `CGO_ENABLED=1` отсутствует `gcc`; substitute focused
  stress `-count=100` PASS;
- Existing Coverage Report: Added Proof/Regression Tests покрывают exact
  surface, sentinel precedence, immutable Target/Binding, duplicate rejection,
  authorization-before-mutation, error identity, Start/Stop/Observe routing,
  panic, cancellation и independent scope progress;
- dependency direction: production package импортирует только `context`,
  `errors`, Loader, runtimeconfigload, Flow и Owner packages;
- documentation: DP-013 headings 34/34 EN/RU; 263 relative links в 21 changed
  Markdown files, broken 0; stale current-status scan PASS;
- `git diff --check`, conflict-marker и trailing-whitespace checks: PASS.

## Scope Audit

**Accepted — 23 Required, 0 Questionable, 0 Removable.**

- `internal/runtimemanagement/directory.go` — Required production behavior;
- `internal/runtimemanagement/directory_test.go` — Required proof/regression
  evidence;
- task record/index — Required governance и discoverability;
- DP-013 EN/RU и design indexes — Required factual implementation status;
- DP-014/015/016/018 EN/RU — Required removal of current dependent-boundary
  contradictions; DP-017 не содержал stale DP-013 implementation claim;
- root README EN/RU — Required user-facing current capability boundary;
- MASTER_PLAN EN/RU — Required milestone/debt status;
- `.ai/PROJECT_CONTEXT.md`, `spec/current-state.md`, `spec/decisions.md` —
  Required durable continuation state;
- existing packages, dependencies, API transports, generated и unrelated files
  не изменены.

## Documentation Sync

`Synchronized`.

- Design Status DP-013 остаётся Draft, Implementation Status изменён только на
  factual `Implemented in isolation`;
- EN/RU DP mirrors, indexes, root README и MASTER_PLAN сохраняют semantic
  parity;
- project-state sources отражают завершённую TASK-023 и implemented isolated
  boundary;
- DP-014–DP-018 остаются Approved/Planned; private Start-claim continuation,
  persistence, concrete policy, integration и Production Activation остаются
  absent/blocked;
- `CHANGELOG.md` Not applicable: release/production capability не меняется.

## Independent Review

**Approved — 0 blocking и 0 nonblocking findings.**

- initial coverage finding по downstream error identity, concurrent same/different
  Target authorization и cancellation during Authorize устранён bounded
  test-only rework;
- repeat review подтвердил exact DP-013 surface, sentinels, precedence,
  construction, routing, lifecycle/cancellation/concurrency invariants и
  dependency direction;
- EN/RU parity, 23/0/0 Scope Audit, removable-question и отсутствие premature
  integration подтверждены;
- Reviewer checks: focused stress, full tests, vet, gofmt, `go mod tidy -diff`
  и `git diff --check` PASS; race limitation подтверждена.

## Coordinator Acceptance

**Accepted.**

- Definition of Done и Acceptance Criteria выполнены;
- Tester `PASS WITH LIMITATION`, independent Reviewer `Approved` 0/0,
  PROCESS-002 `Synchronized`, Scope Audit 23/0/0;
- DP-013 реализован только изолированно; Design Status остаётся Draft;
- concrete policy, persistence, integration и Production Activation не
  добавлены.

## Commit Gate

- exact command `Разрешаю коммит.` получена: да;
- commit message policy: Conventional Commits;
- selected message: `feat(runtime): add management routing`;
- exact accepted file set: 23 Required files из принятого Scope Audit;
- post-acceptance changes: только эта bounded Commit Gate/closure запись и
  исправление stale `In Progress` wording внутри task record; accepted behavior,
  documentation semantics и scope не изменены;
- temporary, generated и unrelated files: отсутствуют;
- разрешён ровно один локальный task commit; push, PR, merge и publication не
  разрешены этой командой.

## Process Health

- **Triggered as overdue ten-task checkpoint:** после TASK-012 завершены ровно
  TASK-013–TASK-022; TASK-022 ошибочно записала trigger как неприменимый.
- Scope Audit этих records не фиксирует accepted Questionable/Removable;
  post-merge product defects и fixes не обнаружены;
- race detector повторно недоступен из-за отсутствующего C compiler;
  substitute stress checks используются явно;
- TASK-021 потребовала два bounded review/rework pass, но не более двух returns;
- bounded finding: ten-task counter должен считаться от последнего completed
  Process Health Review; process text менять не требуется.

## Handoff

Implementation, verification, PROCESS-002, Scope Audit, independent review и
Coordinator Acceptance завершены. Изменения готовы только к отдельному Commit
Gate; permission получена и применяется к exact accepted file set.

## Next Candidate

- Ready для отдельного bounded readiness/intake: implementation prerequisites
  Approved/Planned DP-014 durable operational identity persistence;
- package/API/storage technology и verification scope должны быть однозначно
  определены отдельным Task Contract до code changes;
- candidate не активирован.

## Closure

- final status: `Completed — Coordinator Accepted`;
- closure date: `2026-08-02`;
- changed files: 23 Required, 0 Questionable, 0 Removable;
- known limitation: race detector unavailable без C compiler; focused stress
  PASS;
- до Commit Gate stage, commit, push, PR, merge и publication не выполнялись;
- exact permission выдана только на один локальный task commit; push, PR,
  merge и publication остаются запрещены.

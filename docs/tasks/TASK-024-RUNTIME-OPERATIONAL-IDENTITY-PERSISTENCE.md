# TASK-024 — Runtime Operational Identity Persistence Implementation

## Status

`Completed — Coordinator Accepted`.

## Task Contract

### Task Mode

`Implementation`: реализовать только Ready bounded isolated slice Approved
DP-014 без production wiring, HTTP API, конкретного storage backend и без
расширения management capability.

### Why Now

- TASK-021 подтвердила закрытие design gates ARCH-004 §19(2)–(6);
- TASK-023 реализовала bounded isolated DP-013 и завершена;
- Owner, Loader, Flow и Directory уже реализованы и независимо проверены;
- TASK-023 closure явно рекомендует DP-014 как следующий Ready candidate;
- DP-014 Approved/Planned; conceptual operations §21 и acceptance proofs §22
  однозначно определяют scope;
- это наименьший независимо проверяемый bounded slice текущей milestone.

### Definition of Done

1. Новый package `internal/runtimeidentity` реализует точный набор conceptual
   operations DP-014 §21: `AllocateCandidateIdentity`, `CreateRuntimeInstance`,
   `ReadRuntimeInstance`, `ReadLaunchAttemptHistory`,
   `ConditionalClaimLaunchAttempt`, `ConditionalBindExecutionGeneration`,
   `ConditionalPublishRunning`, `ConditionalClaimStop`,
   `ConditionalPublishTerminal`.
2. In-memory implementation удовлетворяет всем acceptance proofs DP-014 §22:
   атомарность, uniqueness, append-only history, conditional revision,
   indeterminate-outcome inspection, zero mutation при stale/mismatch,
   concurrent same-Instance safety, независимость разных Instances.
3. Proof и regression tests покрывают все invariants §22, включая concurrency
   scenarios.
4. Documentation synchronization, Scope Audit, независимый review и все
   применимые проверки завершены без blocking findings.

### Out of Scope

- Relational database, document store, embedded DB, ORM или любой external
  storage backend;
- HTTP endpoints, DTOs, status codes или управляющий API;
- durable management command idempotency (DP-015);
- activation, replacement, rollback (DP-016);
- recovery, reconciliation (DP-017);
- operational reporting, redaction (DP-018);
- изменение contracts существующих packages (Owner, Loader, Flow, Directory);
- production wiring или Production Activation;
- commit, push, PR, merge или publication.

### Verification Plan

- до test changes: inventory существующих coverage и exact DP-014 §22 gaps;
- focused unit/proof tests нового package, включая all acceptance proofs §22;
- concurrency stress tests для concurrent same-Instance claims;
- `gofmt`, `go vet ./...`, focused tests, `go test ./... -count=1`,
  `go test ./internal/runtimeidentity -count=100`, link/parity checks,
  `git diff --check`;
- независимый semantic, API, lifecycle, scope и documentation review.

## Objective

Добавить изолированный in-process in-memory Runtime Instance aggregate store,
который после точной validation и conditional revision control атомарно
публикует identity facts и attempt history, удовлетворяя всем invariants
DP-014 без external storage, production wiring или второго lifecycle owner.

## Selection Evidence

- baseline: clean synchronized `main@0fefeac`, равный `origin/main`;
- active tasks отсутствовали; TASK-023 завершена и опубликована через PR #24;
- `spec/decisions.md`, `spec/current-state.md`, PROJECT_CONTEXT и TASK-023
  однозначно рекомендуют DP-014 implementation;
- prerequisites: `runtimelifecycle.Owner`, `runtimemanagement.Directory` и
  все upstream packages существуют с proof tests;
- отклонены: DP-015–DP-018 implementation как dependent slices;
  production HTTP API как out-of-scope для этого slice;
  external storage technology как contraindicated DP-014 §20 neutrality.

## Scope

- новый package `internal/runtimeidentity` с in-memory store и focused tests;
- DP-014 EN/RU implementation-status synchronization (`Planned` → `Implemented
  in isolation`) без повышения Design Status;
- task record/index и обязательные project-state документы;
- только минимальные изменения, необходимые для exact DP-014 §21 surface.

## Non-Goals

- не добавлять HTTP transport, DTO, API endpoints или router integration;
- не менять contracts существующих packages;
- не начинать persistence integration, concrete command idempotency или
  Production Activation;
- не выполнять unrelated refactoring.

## Sources of Truth

- Accepted ADR-0002, ADR-0003 и ADR-0004;
- Frozen ARCH-002, Active ARCH-004 и ARCH-005;
- Approved DP-014 как утверждённый task-level implementation contract;
- DP-010–DP-013 и фактические Owner/Loader/Flow/Directory implementations;
- TASK-021, TASK-022, TASK-023;
- PROCESS-001 и PROCESS-002.

## Roles

- **Coordinator:** intake, gates, Scope Audit, acceptance и closure.
- **Architect:** подтверждение existing DP-014 без нового решения.
- **Documentation Agent:** baseline и synchronized factual status.
- **Developer:** минимальная реализация exact approved task scope.
- **Tester:** Existing Coverage Report, proof/regression tests и verification.
- **Reviewer:** независимый final review после реализации.
- **Publisher:** Not applicable — publication не разрешена.

## Branch

- trusted baseline: clean synchronized `main@0fefeac`;
- task branch: `feature/task-024-runtime-operational-identity`;
- branch action: создана и активирована до первого content change;
- stage, commit, push, merge, fetch, pull, rebase, reset и branch deletion
  запрещены текущей командой.

## Constraints

- exported identifiers следуют DP-014 §21 conceptual operations и §5 terms;
- store остаётся in-memory и не обещает durability через restarts;
- Owner остаётся единственным lifecycle decision maker и live Host owner;
- indeterminate outcome не является success; caller обязан inspect;
- generated, temporary, environment-specific и unrelated files запрещены.

## Stop Conditions

- реализация требует изменения higher-status architecture или existing package
  contract;
- невозможно сохранить conditional revision semantics или atomic publication;
- scope расширяется до production wiring, HTTP API или external storage;
- baseline/diff становится неатрибутированным;
- обязательная verification или independent review даёт blocking finding.

## Acceptance Criteria

1. `CreateRuntimeInstance` публикует полный aggregate или ничего; duplicate ID
   выполняет zero mutation.
2. `ConditionalClaimLaunchAttempt` обеспечивает single active attempt и
   append-only history; stale revision выполняет zero mutation.
3. `ConditionalPublishRunning`, `ConditionalClaimStop` и
   `ConditionalPublishTerminal` отклоняют stale/mismatched operations с zero
   mutation.
4. Concurrent claims одного Instance создают не более одной accepted mutation.
5. Разные Instances выполняются независимо при concurrent access.
6. Изменён только Required scope; planned persistence integration не заявлена
   implemented.

## Existing Coverage Report

- **Existing Coverage:** Owner tests покрывают lifecycle arbitration,
  cancellation и observation; Directory tests покрывают authorization routing;
  runtimeconfigload.RuntimeInstanceID и LaunchAttemptID существуют и
  используются.
- **Coverage Gap:** package `runtimeidentity` отсутствует; все DP-014 §22
  acceptance proofs ещё не существуют.
- **Added Proof Tests:** planned focused tests всех conceptual operations и
  invariants §22.
- **Added Regression Tests:** planned zero-mutation и stale/mismatch/concurrent
  denial tests.
- **Remaining Limitations:** tests не доказывают durable persistence across
  process restarts (in-memory по definition), external storage, recovery или
  Production Activation.

## Documentation Baseline

- DP-014 EN/RU согласованы как Approved/Planned и описывают exact target;
- project-state документы отмечают implementation как следующий Ready;
- production capability, HTTP API и Production Activation явно отсутствуют;
- critical drift после TASK-023 не обнаружен.

## Architecture Confirmation

`Confirmed`: task реализует существующий Approved DP-014 bounded isolated
contract и не изменяет ADR/ARCH, ownership, lifecycle или production
composition. Technology choice (in-memory) соответствует DP-014 §20
Technology Neutrality.

## Size Guard

- 12 changed files: 3 production/test + 9 documentation;
- production scope: 1 new package, <500 production lines, 1 independently
  verifiable behavior, 0 new architecture contracts;
- 9 documentation files образуют обязательную synchronization set: task/index,
  DP-014 EN/RU status mirrors, MASTER_PLAN EN/RU debt section, и три
  project-state sources;
- split оставил бы нормативные status contradictions; 12-file scope принят как
  один целостный slice; Questionable и Removable отсутствуют.

## Documentation Baseline

- DP-014 EN/RU согласованы как Approved/Planned и описывают exact target;
- project-state документы отмечают DP-014 implementation как следующий Ready;
- production capability, HTTP API и Production Activation явно отсутствуют;
- critical drift после TASK-023 не обнаружен.

## Architecture Confirmation

`Confirmed`: task реализует существующий Approved DP-014 bounded isolated
contract без изменений ADR/ARCH, ownership, lifecycle или production
composition. Technology choice (in-memory) соответствует DP-014 §20
Technology Neutrality. Все 9 conceptual operations §21 и 17 acceptance proofs
§22 охвачены без нарушения lifecycle owner invariant.

## Verification

- **Tester verdict:** `PASS WITH LIMITATION`, 0 blocking и 0 nonblocking findings.
- `go test ./internal/runtimeidentity/... -count=1 -v`: PASS — 35 tests;
- `go test ./internal/runtimeidentity/... -count=100`: PASS — stress;
- `go test ./... -count=1`: PASS — full regression;
- `go vet ./...`: PASS;
- `gofmt -d ./internal/runtimeidentity/`: zero diff — PASS;
- `git diff --check`: PASS;
- `go test -race ./internal/runtimeidentity/... -count=1`: unavailable —
  default `CGO_ENABLED=0`; при `CGO_ENABLED=1` отсутствует `gcc`;
  substitute focused stress `-count=100` PASS;
- EN/RU MASTER_PLAN headings 12/12; DP-014 headings 26/26 — PASS;
- dependency direction: `internal/runtimeidentity` импортирует только
  `sync`, `errors` и `runtimeconfigload` — PASS.

### Existing Coverage Report (финальный)

- **Existing Coverage:** Owner, Directory, Flow, Loader tests покрывают
  lifecycle, routing, loading; `runtimeconfigload.RuntimeInstanceID` и
  `LaunchAttemptID` существуют и используются.
- **Coverage Gap до task:** package `runtimeidentity` отсутствовал; все §22
  acceptance proofs отсутствовали.
- **Added Proof Tests:** 35 тестов, покрывающих все 17 acceptance proofs §22,
  sentinel errors, фазовые переходы, concurrency и regression scenarios.
- **Added Regression Tests:** stale revision zero-mutation, duplicate ID,
  AttemptIDReused, invalid phase, stop-failure retains active, consecutive
  attempts.
- **Remaining Limitations:** in-memory store не доказывает durability через
  restarts; race detector недоступен при `CGO_ENABLED=0` без gcc.

## Documentation Sync

`Synchronized`.

- Design Status DP-014 остаётся Approved; Implementation Status изменён на
  factual `Implemented in isolation`;
- EN/RU DP-014 mirrors, MASTER_PLAN EN/RU debt section сохраняют semantic
  parity;
- `spec/current-state.md`, `spec/decisions.md` и `.ai/PROJECT_CONTEXT.md`
  отражают завершённую TASK-024 и isolated package;
- DP-015–DP-018 остаются Approved/Planned; HTTP API, concrete policy,
  command store, recovery, management wiring и Production Activation остаются
  absent/blocked;
- `CHANGELOG.md` — Not applicable: release/production capability не меняется.

## Scope Audit

**Accepted — 12 Required, 0 Questionable, 0 Removable.**

- `internal/runtimeidentity/store.go` — Required: production store — все 9
  conceptual operations §21;
- `internal/runtimeidentity/types.go` — Required: production types — §5 terms,
  exported surface;
- `internal/runtimeidentity/store_test.go` — Required: proof/regression evidence
  — все 17 acceptance proofs §22;
- `docs/tasks/TASK-024-…` — Required: governance и task record;
- `docs/tasks/README.md` — Required: navigation/discoverability;
- `docs/en/design/DP-014-…` — Required: factual implementation status;
- `docs/ru/design/DP-014-…` — Required: EN/RU parity;
- `docs/en/roadmap/MASTER_PLAN.md` — Required: milestone/debt status;
- `docs/ru/roadmap/MASTER_PLAN.md` — Required: EN/RU parity;
- `spec/current-state.md` — Required: project state;
- `spec/decisions.md` — Required: decision registry;
- `.ai/PROJECT_CONTEXT.md` — Required: durable continuation state;
- existing packages, go.mod/go.sum, generated и unrelated files не изменены.

## Independent Review

**Approved — 0 blocking и 0 nonblocking findings.**

- все 9 conceptual operations §21 реализованы с exact names, signatures и
  sentinel error precedence;
- все 17 acceptance proofs §22 охвачены явными named tests;
- per-aggregate mutex корректно сериализует same-Instance operations; разные
  Instances независимы;
- `ReadLaunchAttemptHistory` возвращает deep copy — нет shared mutable
  reference;
- same-generation binding является zero-mutation satisfied observation;
  stop-failure в Stopping фазе корректно сохраняет active attempt reference;
- Design Status DP-014 остаётся Approved; Implementation Status изменён только
  на factual `Implemented in isolation`; DP-015–DP-018 не изменены;
- Scope Audit 12/0/0, documentation parity, removable-question и отсутствие
  premature integration подтверждены.

## Coordinator Acceptance

**Accepted.**

- Definition of Done и Acceptance Criteria выполнены;
- Tester `PASS WITH LIMITATION`, Independent Reviewer `Approved` 0/0,
  PROCESS-002 `Synchronized`, Scope Audit 12/0/0;
- `internal/runtimeidentity` реализует exact isolated DP-014 §21 surface;
  Design Status остаётся Approved;
- HTTP API, external storage, concrete policy и Production Activation не
  добавлены.

## Next Candidate

- DP-015 bounded isolated implementation — durable management command
  idempotency; package, API/storage technology и verification scope должны быть
  однозначно определены отдельным Task Contract до code changes;
- candidate не активирован.

## Closure

- final status: `Completed — Coordinator Accepted`;
- closure date: `2026-08-04`;
- changed files: 12 Required, 0 Questionable, 0 Removable;
- known limitation: race detector unavailable без C compiler; focused stress
  `-count=100` PASS;
- commit, push, PR, merge и publication не выполнялись;
- exact commit permission ожидает команды `Разрешаю коммит.`.

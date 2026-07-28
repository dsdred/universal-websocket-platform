# TASK-010 — Production Launch Pipeline Design

## Status

**Completed — Coordinator Accepted**

## Objective

Подготовить сфокусированный зеркальный Draft DP-011, определяющий минимальный
production launch flow от Runtime Lifecycle Owner через Configuration Loader и
Snapshot Builder к stateless Runtime Launcher.

Задача должна закрыть отсутствующий integration contract между уже
реализованными изолированно DP-007, DP-008, DP-009 и DP-010, не приступая к
production implementation.

## Selection Evidence

- Baseline: clean synchronized `main@d4f3413`, совпадающий с `origin/main`.
- Active task records отсутствуют; TASK-009 имеет статус
  `Completed — Coordinator Accepted`.
- `.ai/PROJECT_CONTEXT.md` и `spec/current-state.md` называют следующим
  незапущенным разрывом production wiring
  `Loader-to-Builder-to-Launcher`.
- MASTER_PLAN относит отсутствие production launch flow к architectural debt и
  прямо требует устранять такой debt через отдельный DP.
- DP-009 AP-003/AP-011 и DP-010 future integration proofs остаются
  integration-gated до определения и проверки production composition.
- Candidate является наименьшим independently reviewable prerequisite:
  сначала design contract, затем отдельная implementation task.
- Отклонённые alternatives:
  - немедленная production implementation — запрещена до определения
    integration contract;
  - persistence operational identities — отдельный deferred contract;
  - Control Service routing и management API — требуют отдельной product/API
    task;
  - retry, restart, reconciliation и supervision — не определены и исключены
    DP-010;
  - Delivery, TLS, Metrics и operational diagnostics — более поздние либо
    независимые Beta epics и не закрывают текущий dependency gap.

## Scope

- определить ровно один production launch orchestration path;
- сохранить Lifecycle Owner единственной lifecycle authority Launch Attempt;
- определить adapter/orchestrator между `PrepareStart`, Loader, Builder и
  `Start`;
- определить exact identity/provenance propagation и immutable handoff;
- определить mapping Loader и Builder failures в закрытый `PreparationResult`;
- определить ownership входов, Snapshot, Bootstrap request и Host handoff;
- определить cancellation, concurrency и no-hidden-state constraints;
- определить local и integration acceptance proofs, включая DP-009
  AP-003/AP-011 и DP-010 future integration proofs;
- создать зеркальные EN/RU DP-011 и обновить только необходимые индексы и
  project-state документы;
- выполнить PROCESS-002, documentation verification, scope audit и независимый
  review.

## Non-Goals

- production code и tests;
- изменение существующих exported contracts DP-007–DP-010;
- Control Service endpoint, command DTO, authorization или routing;
- persistence и durable operational identities;
- retry, restart, replacement, reconciliation, recovery или supervision;
- background worker, process registry, generic manager или service locator;
- Delivery, TLS, Metrics, diagnostics или Plugin contracts;
- повышение Design Status существующих Draft DP;
- следующая implementation task.

## Sources of Truth

- PROCESS-001 и PROCESS-002;
- ADR-0002 Configuration DSL EN/RU — Accepted;
- ADR-0003 Runtime Architecture EN/RU — Accepted;
- ARCH-002 Runtime Foundation Freeze EN/RU — Active/Frozen;
- ARCH-004 Runtime Deployment and Identity Model EN/RU — Active/Approved;
- ARCH-005 Runtime Configuration Snapshot and Loading Model EN/RU —
  Active/Approved;
- DP-007 Configuration Loader Contract EN/RU — Draft, implemented in
  isolation;
- DP-008 Snapshot Builder Contract EN/RU — Draft, implemented in isolation;
- DP-009 Runtime Bootstrap Contract EN/RU — Draft, Bootstrap и Launcher
  implemented in isolation;
- DP-010 Runtime Lifecycle Owner Contract EN/RU — Draft, implemented in
  isolation;
- `.ai/PROJECT_CONTEXT.md`, `spec/current-state.md`, `spec/decisions.md` и
  MASTER_PLAN EN/RU;
- production declarations и tests как factual evidence, но не как источник,
  переопределяющий architecture.

## Roles

- **Coordinator:** selection, gates, handoffs, scope audit и acceptance.
- **Architect:** DP-011 decision, boundaries, constraints и acceptance proofs.
- **Documentation Agent:** task record, зеркальная фиксация design и
  PROCESS-002.
- **Developer:** неприменим; production implementation запрещена.
- **Tester:** неприменим к production behavior; документационные и
  repository checks выполняет Documentation Agent.
- **Reviewer:** независимо проверяет architecture, parity, truthfulness и
  scope.
- **Publisher:** неприменим; commit и publication не разрешены.

## Branch

- trusted baseline: `main@d4f3413`;
- task branch: `docs/task-010-production-launch-pipeline-design`;
- branch создана безопасно до content changes;
- этот task record является первым content change;
- запрещены stage, commit, push, merge, rebase, branch deletion, fetch, pull и
  remote mutation.

## Constraints

- DP-011 не переопределяет ADR, Active/Frozen ARCH или существующие DP;
- orchestration не становится новой lifecycle authority;
- Loader остаётся owner source observation, Builder — semantic validation и
  detached Snapshot construction, Launcher — stateless delegation, Owner —
  lifecycle serialization и Host ownership;
- exact identities одного Launch Attempt не могут быть заменены между
  `PrepareStart` и `Start`;
- orchestration не хранит durable state и не вводит retry/policy;
- planned architecture не документируется как implemented;
- EN/RU documents имеют одинаковую структуру, статусы и нормативный смысл.

## Stop Conditions

- для решения требуется изменить Accepted ADR или Active/Frozen ARCH;
- существующие DP оставляют несовместимые обязательные semantics;
- выбор production caller требует продуктового/API решения;
- contract невозможно определить без persistence, retry или supervision;
- baseline получает неатрибутированные изменения;
- independent Reviewer выдаёт blocking finding.

## Acceptance Criteria

1. Task record остаётся первым content change ветки.
2. Documentation Baseline подтверждает отсутствие blocking drift.
3. DP-011 EN/RU имеют статус Draft и Implementation Status Planned.
4. DP-011 определяет один orchestration flow
   `PrepareStart -> Load -> Build -> Start`.
5. Ownership, identities, provenance, validation и failure boundaries
   совместимы с ARCH-004/005 и DP-007–DP-010.
6. Loader/Builder rejection преобразуется в authentic closed
   `PreparationResult` без вызова Launcher.
7. Только accepted Snapshot достигает Owner `Start`; каждый accepted attempt
   вызывает существующий stateless Launcher не более одного раза.
8. Cancellation и concurrency не создают второй lifecycle owner, hidden retry
   или detached background work.
9. Proof matrix закрывает production presence/statelessness AP-003/AP-011
   DP-009 и перечисленные integration proofs DP-010, не утверждая их
   выполненными.
10. Scope явно не выбирает Control Service API, persistence или management
    policy.
11. EN/RU parity, links, formatting и `git diff --check` проходят.
12. PROCESS-002 возвращает `Synchronized`.
13. Scope Audit не содержит unresolved `Questionable` или `Removable`.
14. Independent Reviewer возвращает `Approved` без blocking findings.
15. Production implementation, tests, commit, push и merge не выполняются.

## Verification

- structural EN/RU heading and status parity;
- relative-link validation;
- terminology and source-reference audit;
- `git diff --check`;
- full diff scope audit;
- independent architecture/documentation review.

## Architecture Handoff

**Complete — new focused Draft required; implementation prohibited.**

- ARCH-004/005 требуют, чтобы Lifecycle Owner создавал Launch Attempt до
  Loader/Builder и оставался единственной lifecycle authority.
- DP-007–DP-010 определяют изолированные exact boundaries, но не определяют
  owner-bound adapter, Builder Diagnostics failure representation или caller
  cancellation semantics полного flow.
- Создан зеркальный Draft DP-011 с planned package
  `internal/runtimelaunchflow`, immutable Owner/Loader binding и exact
  `PrepareStart -> Load -> Build -> Start`.
- Flow не создаёт goroutine или worker: тот же synchronous `Start` invocation
  выполняет Load/Build/Owner.Start. Caller Cancellation Gate перед
  `PrepareStart` задаёт winner semantics; после победы Gate cancellation не
  прерывает accepted operation, а Stop остаётся Owner-owned.
- Loader failures сохраняются unchanged; exhaustive Builder Diagnostics
  становятся immutable `BuildFailure`; только matching Snapshot достигает
  Owner `Start`.
- DP-011 не меняет DP-007–DP-010, не выбирает Source adapter, persistence,
  management API, authorization, retry, reconciliation или supervision.
- Production Activation отдельно gated; planned package implementation не
  будет означать управление Runtime из Control Service.

## Documentation Baseline

**Synchronized for design; blocking drift отсутствует.**

- Прочитаны PROCESS-001/002, role contracts, PROJECT_CONTEXT, current-state,
  decisions и MASTER_PLAN EN/RU.
- Проверены Active/Frozen ARCH-002, Active/Approved ARCH-004/005 и связанные
  DP-007–DP-010.
- Factual Go surfaces подтверждают exact existing boundaries:
  `configurationloader.Loader.Load`, `runtimeconfig.Builder.Build`,
  `runtimelifecycle.Owner.PrepareStart/Start` и `runtime.Launch`.
- Production use `runtimelifecycle` отсутствует; package реализован только
  изолированно.
- DP-011 EN/RU имеют одинаковые 33 headings, 14 fences, Draft/Planned statuses
  и эквивалентный normative structure.
- Проверены links изменённых navigation/design/roadmap documents: broken 0.
- Planned design нигде не представлен как implemented capability.

## PROCESS-002

**Synchronized.**

- DP-011 EN/RU созданы зеркально со статусами Draft/Planned.
- Design indexes EN/RU содержат DP-011 с одинаковым статусом.
- MASTER_PLAN EN/RU связывает существующий architectural debt с DP-011 и
  сохраняет implementation/Production Activation отсутствующими.
- `.ai/PROJECT_CONTEXT.md`, `spec/current-state.md` и `spec/decisions.md`
  отделяют planned contract от implemented state.
- README и CHANGELOG неприменимы: public product/API capability и release
  artifact не появились.
- ADR/ARCH и DP-007–DP-010 не изменены: higher-authority architecture и
  существующие contracts не пересматривались.
- Production code/tests не менялись; Developer и production Tester stages
  обоснованно неприменимы.

## Scope Audit

**Accepted — 11 Required, 0 Questionable, 0 Removable.**

Required design/process:

- `docs/ru/design/DP-011-runtime-launch-pipeline-integration.md`;
- `docs/en/design/DP-011-runtime-launch-pipeline-integration.md`;
- `docs/tasks/TASK-010-PRODUCTION-LAUNCH-PIPELINE-DESIGN.md`;
- `docs/tasks/README.md`.

Required navigation/project state:

- `docs/ru/design/README.md`;
- `docs/en/design/README.md`;
- `docs/ru/roadmap/MASTER_PLAN.md`;
- `docs/en/roadmap/MASTER_PLAN.md`;
- `.ai/PROJECT_CONTEXT.md`;
- `spec/current-state.md`;
- `spec/decisions.md`.

Production, test, generated, CI и unrelated paths: 0.

## Independent Review

**Approved after rework — 0 blocking, 0 nonblocking findings.**

Initial Reviewer verdict `Needs Revision`:

- B-001: detached Attempt Worker не имел гарантированного termination при
  blocking `Source.LoadExact`;
- B-002: обещание cancellation до Owner claim не имело общей linearization
  boundary.

Rework:

- Worker, goroutine и channel удалены; Flow выполняет одну synchronous
  call-stack Start Operation, а blocking Source правдиво блокирует caller;
- введён один Caller Cancellation Gate как exact `ctx.Err()` read до
  `PrepareStart`; nil-at-Gate выигрывает последующую гонку cancellation.

Repeat Reviewer подтвердил совместимость Stop и same-token convergence с
DP-010, EN/RU parity, links и `git diff --check`; verdict `Approved`.

## Coordinator Acceptance

**Accepted.**

- Все acceptance criteria выполнены.
- Design-only scope сохранён; implementation не начата.
- Architecture Handoff, Documentation Baseline, PROCESS-002, Scope Audit и
  independent Review завершены.
- Task готова к commit только после отдельного разрешения пользователя.
- Stage, commit, push и merge не выполнялись.

## Next Recommendation

Не активирована. Следующий candidate — отдельная минимальная implementation
`internal/runtimelaunchflow` и local proof tests по DP-011, без Source adapter,
Control Service routing, persistence или Production Activation. Readiness
должен быть подтверждён новым task intake.

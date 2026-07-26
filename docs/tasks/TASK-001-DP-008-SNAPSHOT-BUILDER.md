# TASK-001 — DP-008 Snapshot Builder поверх DetachedLoadResult

## Status

**Blocked for Architecture**

**Branch:** `feature/dp-008-snapshot-builder`

## Objective

Реализовать минимальный полный контракт Draft DP-008 Snapshot Builder поверх
neutral `DetachedLoadResult`, включая обязательный provenance ARCH-005 и
полные блокирующие Diagnostics, без подключения Builder к production launch
pipeline.

## Scope

После устранения архитектурного blocker задача включает:

- вход Builder только через полный neutral `DetachedLoadResult`;
- defensive handoff, schema и identity checks на границе Builder;
- полный provenance Workspace, Configuration, ConfigurationVersion ID и
  version number, Configuration schema, Runtime Instance и Launch Attempt;
- semantic validation всех применимых Listener, TLS, Timeout, Authentication
  и Routing rules;
- canonical normalization и создание полного immutable Runtime Snapshot;
- полную, дедуплицированную и детерминированную коллекцию blocking Diagnostics;
- proof tests для success, failure, provenance, immutability, detachment и
  determinism.

## Non-Goals

- изменение утверждённой или Frozen архитектуры;
- повышение Design Status DP-008;
- DP-009;
- Runtime Launcher, Runtime Lifecycle Owner или Control Service launch
  pipeline;
- создание Runtime resources, source access, lifecycle decisions, retry,
  fallback, persistence или operational diagnostics;
- несвязанный рефакторинг и возможности вне task и authoritative documents.

## Sources of Truth

- [PROCESS-001](../engineering/PROCESS-001-AI-DEVELOPMENT-WORKFLOW.md);
- [PROCESS-002](../engineering/PROCESS-002-DOCUMENTATION-SYNCHRONIZATION.md);
- [ADR-0002: Configuration DSL](../en/adr/0002-configuration-dsl.md);
- [ADR-0003: Runtime Architecture](../en/adr/0003-runtime-architecture.md);
- [ARCH-002: Runtime Foundation Freeze](../en/architecture/ARCH-002-runtime-foundation-freeze.md);
- [ARCH-004: Runtime Deployment and Identity Model](../en/architecture/ARCH-004-runtime-deployment-and-identity-model.md);
- [ARCH-005: Runtime Configuration Snapshot and Loading Model](../en/architecture/ARCH-005-runtime-configuration-snapshot-and-loading-model.md);
- [DP-007: Configuration Loader Contract](../en/design/DP-007-configuration-loader-contract.md),
  Design Status `Draft`, implementation в изоляции;
- [DP-008: Snapshot Builder Contract](../en/design/DP-008-snapshot-builder-contract.md)
  и его [RU-зеркало](../ru/design/DP-008-snapshot-builder-contract.md), Design
  Status `Draft`;
- [текущее состояние проекта](../../spec/current-state.md);
- production-код и тесты только как evidence фактически реализованного
  состояния.

Draft DP-008 является implementation contract и не переопределяет Approved,
Active или Frozen источники.

## Roles and Stage State

- **Coordinator:** intake, task contract, назначение ролей и обработка blocker.
- **Architect:** анализ завершён; возвращён blocker до начала coding.
- **Documentation Agent:** создан этот operational record и обновлена
  навигация задач; выполнена blocker-state synchronization в
  `.ai/PROJECT_CONTEXT.md`, `spec/current-state.md` и MASTER_PLAN EN/RU,
  отражающая статус `Blocked for Architecture` и следующий разрешённый
  refinement. Архитектурные и implementation-документы не изменялись.
- **Developer:** не начинал работу из-за stop condition.
- **Tester:** не начинал verification implementation из-за stop condition.
- **Reviewer:** implementation review не начинался из-за отсутствия
  реализации.
- **Post-Implementation Documentation Synchronization:** не выполнялась;
  реализации не было и фактически реализованная capability не изменилась.

## Constraints

- Production-код изменяет только Developer.
- Coordinator и Documentation Agent не пишут production-код.
- Immutable и detached boundaries сохраняются.
- Builder не подключается к production launch pipeline.
- Builder не получает source, resource, lifecycle или management authority.
- Работа выполняется только в текущей feature-ветке; `main` не изменяется.
- Commit и push запрещены.

## Architecture Handoff and Blocker

Architect сверил задачу с ADR-0002, ADR-0003, ARCH-002, ARCH-005, DP-007 и
Draft DP-008. Архитектурного разрешения для начала реализации нет: DP-008 §13
и §23 явно требуют до coding определить три implementation contracts:

1. точные identity поддерживаемой Configuration schema и compatibility rule;
2. полную структуру полей Runtime Snapshot, section-specific normalization и
   immutable reader surface;
3. точное structured representation Diagnostics, включая namespace codes,
   logical-location grammar, canonical ordering и applicability.

Репозиторий этих контрактов не содержит. Используемое в тестах
`uwp.configuration` v1 является fixture, а не authoritative schema contract.
Выбор этих деталей в Developer stage был бы новым архитектурным решением.

## Acceptance Criteria

После устранения blocker реализация должна доказать:

1. Builder принимает `DetachedLoadResult`, не `ConfigurationVersion` и не
   source-specific input.
2. Defensive checks отклоняют некорректные handoff, schema и identity без
   доступа к source.
3. Успешный Snapshot содержит полный provenance: Workspace, Configuration,
   ConfigurationVersion ID и version number, schema identity/version, Runtime
   Instance и Launch Attempt.
4. Проверяются все применимые Listener, TLS, Timeout, Authentication и Routing
   semantic rules, установленные уточнённым contract.
5. При любом blocking violation возвращается полная дедуплицированная
   Diagnostics collection и не возвращается Snapshot или partial Snapshot.
6. Diagnostics имеют утверждённые structured codes и locations, canonical
   ordering и корректную prerequisite applicability без cascading duplicates.
7. Эквивалентные валидные inputs дают canonical, semantically equivalent
   Snapshots; эквивалентные невалидные inputs дают equivalent Diagnostics.
8. Snapshot immutable и глубоко detached от input; Diagnostics также detached
   от input-owned mutable memory.
9. Повторный и concurrent-safe Build детерминирован и не сохраняет mutable
   state между вызовами.
10. Builder не создаёт Runtime resources, не обращается к source, не принимает
    lifecycle decisions и не реализует launch pipeline или DP-009.

## Mandatory Verification

После реализации обязательны:

- `gofmt` для всех изменённых Go-файлов;
- `go test` затронутых пакетов минимум три раза с `-count=1`;
- `go test ./... -count=1` минимум два раза;
- `go vet ./...`;
- `go test -race` для затронутых пакетов, если среда поддерживает;
- `git diff --check`;
- независимый review task, authoritative architecture, Draft DP-008, кода и
  tests с verdict `Approved` либо rework по обязательным findings;
- PROCESS-002 после утверждённой реализации.

Эти implementation-команды пока не выполнялись: Architecture Analysis
завершился обязательной stop condition до Developer stage.

## Next Allowed Action

Единственный разрешённый следующий шаг — focused Architect refinement Draft
DP-008 в EN/RU, который определит три отсутствующих implementation contracts,
сохранит Design Status `Draft` и не изменит ownership, validation, atomicity
или Runtime-independence model. Уточнение должно пройти независимый review.
Только после его утверждения Coordinator может открыть Developer stage.

## Handoff

- **Completed scope:** зафиксированы task contract, implementation boundaries,
  acceptance criteria, verification contract и точный architecture blocker;
  project-state документы синхронизированы со статусом blocker и следующим
  разрешённым Architect refinement.
- **Production files:** не изменялись.
- **Test files:** не изменялись.
- **Architecture/design files:** не изменялись.
- **Operational documentation:** этот task record и `docs/tasks/README.md`.
- **Project-state documentation:** `.ai/PROJECT_CONTEXT.md`,
  `spec/current-state.md`, `docs/en/roadmap/MASTER_PLAN.md` и
  `docs/ru/roadmap/MASTER_PLAN.md`.
- **Documentation evidence:** local Markdown links и fences проверены;
  MASTER_PLAN EN/RU сохраняют semantic parity и одинаковую heading hierarchy;
  conflict markers и trailing whitespace отсутствуют; `git diff --check`
  выполнен успешно.
- **Post-implementation synchronization:** не применима до реализации и не
  объявляется выполненной.
- **Finding:** три implementation contracts отсутствуют; coding запрещён.
- **Unresolved risk:** до focused refinement невозможно реализовать Builder
  без скрытого schema, Snapshot API или Diagnostics design decision.
- **Final status:** Blocked for Architecture.

## Closure

- **Final status:** Blocked for Architecture
- **Reason:** DP-008 §23 implementation prerequisites не определены.
- **Commit:** не создавался.

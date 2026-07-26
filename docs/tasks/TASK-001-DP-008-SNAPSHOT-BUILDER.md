# TASK-001 — DP-008 Snapshot Builder поверх DetachedLoadResult

## Status

**Completed — Coordinator Accepted**

**Branch:** `feature/task-001-dp008-snapshot-builder`

## Objective

Реализовать минимальный полный контракт Draft DP-008 Snapshot Builder поверх
neutral `DetachedLoadResult`, включая обязательный provenance ARCH-005 и
полные блокирующие Diagnostics, без подключения Builder к production launch
pipeline.

## Scope

Implementation scope:

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
- **Architect:** focused refinement завершён. Определены точная schema v1,
  полный Snapshot и detached-reader contract, все section semantics,
  исчерпывающий Diagnostics registry и правила applicability, duplicate
  anchoring, deduplication и ordering. Approved/Frozen architecture не
  изменялась.
- **Documentation Agent:** утверждённый Architect handoff зеркально
  зафиксирован в Draft DP-008 EN/RU; после реализации выполнен PROCESS-002 и
  project-state documents синхронизированы с фактическим состоянием
  изолированного Builder.
- **Developer:** реализовал полный контракт Builder поверх neutral
  `DetachedLoadResult`, private Snapshot/readers, полный provenance,
  exhaustive Diagnostics и механически перевёл существующих Runtime
  consumers на reader surface.
- **Tester:** независимо проверил coverage и обязательные команды; после
  Reviewer findings добавлены только недостающие proof tests и повторно
  выполнены целевые проверки.
- **Reviewer архитектурного refinement:** `Approved`; обязательные findings
  отсутствуют.
- **Reviewer implementation:** после двух test-only rework cycles выдал
  итоговый verdict `Approved`; production, architecture и scope findings
  отсутствуют.
- **Post-Implementation Documentation Synchronization:** выполнена по
  PROCESS-002; Design Status DP-008 сохранён `Draft`, Implementation Status
  изменён на `Implemented`.

## Constraints

- Production-код изменяет только Developer.
- Coordinator и Documentation Agent не пишут production-код.
- Immutable и detached boundaries сохраняются.
- Builder не подключается к production launch pipeline.
- Builder не получает source, resource, lifecycle или management authority.
- Работа выполняется только в текущей feature-ветке; `main` не изменяется.
- Commit и push запрещены.

## Architecture Refinement Handoff

Architect сверил задачу с ADR-0002, ADR-0003, ARCH-002, ARCH-005, DP-007 и
Draft DP-008. Первоначальный blocker DP-008 §13 и §23 требовал до coding
определить три implementation contracts:

1. точные identity поддерживаемой Configuration schema и compatibility rule;
2. полную структуру полей Runtime Snapshot, section-specific normalization и
   immutable reader surface;
3. точное structured representation Diagnostics, включая namespace codes,
   logical-location grammar, canonical ordering и applicability.

Focused Architect refinement определил и independent Reviewer утвердил:

1. ровно schema pair `uwp.configuration` / version `1` без negotiation,
   migration, downgrade или fallback;
2. полную private Snapshot model, observable optionality, recursively detached
   reader behavior и section-specific validation/normalization;
3. immutable Diagnostics tuple, exhaustive Code/Location/fixed English Message
   registry, applicability, precedence, duplicate anchoring, deduplication и
   canonical ordering.

Контракт зафиксирован в DP-008 EN/RU с semantic и heading parity. Design Status
DP-008 остаётся `Draft`. Architecture blocker был снят до coding; текущая
реализация прошла verification и независимый implementation review, поэтому
Implementation Status — `Implemented`.

## Acceptance Criteria

Реализация должна доказать:

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

Результаты verification:

- `gofmt` и итоговый `gofmt -d` для изменённых Go-файлов — PASS, formatter
  drift отсутствует;
- целевые `go test` затронутых пакетов с `-count=1` — PASS, 3 из 3 запусков
  обязательного gate; после test-only rework целевые тесты также повторно
  прошли;
- `go test ./... -count=1` — PASS, 2 из 2 запусков обязательного gate;
- `go vet ./...` — PASS;
- `git diff --check` — PASS;
- Diagnostics registry DP-008 EN/RU — 93 из 93 строк с точным parity;
- race detector в текущей Windows-среде недоступен: при `CGO_ENABLED=0`
  команда сообщает `-race requires cgo`, а при `CGO_ENABLED=1` отсутствует
  `gcc` в `PATH`.

После финальной documentation synchronization Coordinator выполнил применимый
repository gate: targeted tests PASS 3/3, full repository tests PASS 2/2,
`go vet ./...`, `gofmt -d` и `git diff --check` — PASS. Race gap остаётся
только ограничением среды.

## Next Allowed Action

Следующий разрешённый шаг — commit закрытой TASK-001 только после отдельного
разрешения пользователя. Следующая development task не выбрана. DP-009,
Runtime Launcher, Runtime Lifecycle Owner и production launch pipeline не
являются следующим неявным шагом.

## Handoff

- **Architect handoff:** `Ready for Implementation`; конфликтов с ADR-0002,
  ADR-0003, ARCH-002, ARCH-004, ARCH-005, DP-007 и Draft DP-008 не найдено.
  Builder ограничен neutral handoff, pure validation/normalization и
  immutable Snapshot; Runtime lifecycle и production launch pipeline
  запрещены.
- **Developer handoff:** реализован exclusive result contract
  `Snapshot/no Diagnostics` либо `no Snapshot/non-empty Diagnostics`, exact
  schema `uwp.configuration` v1, полный ARCH-005 provenance, private detached
  readers, все 93 blocking Diagnostics и механическая миграция существующих
  consumers без подключения production pipeline.
- **Tester handoff:** обязательный gate прошёл; race detector недоступен по
  причине toolchain environment, а не test failure.
- **Reviewer handoff:** initial `Needs Revision` содержал пять test coverage
  findings: exhaustive Runtime support matrix, три именованных Runtime proofs,
  AP-004 full model, AP-005 zero result и recursive detachment. Re-review
  оставил два test findings: `availablePort` для TLS proof и package-local
  cross-Build backing isolation. Все findings исправлены только тестами.
  Итоговый verdict — `Approved`.
- **Production files:** `internal/configurationloader/loader.go`,
  `internal/router/router.go`, `internal/runtime/composition.go`,
  `internal/runtime/configuration_validation.go`,
  `internal/runtime/container.go`, `internal/runtime/host.go`,
  `internal/runtimeconfig/builder.go`, `internal/runtimeconfig/diagnostic.go`,
  `internal/runtimeconfig/model.go`, `internal/runtimeconfig/validation.go`,
  `internal/runtimeconfigload/contract.go`.
- **Test files:** `internal/configurationloader/loader_test.go`,
  `internal/router/router_test.go`, `internal/runtime/bootstrap_test.go`,
  `internal/runtime/configuration_validation_test.go`,
  `internal/runtime/container_test.go`,
  `internal/runtime/handshake_shutdown_test.go`,
  `internal/runtime/host_test.go`,
  `internal/runtime/router_integration_test.go`,
  `internal/runtimeconfig/builder_test.go`,
  `internal/runtimeconfig/registry_test.go`,
  `internal/runtimeconfig/routing_builder_test.go`.
- **Architecture/design files:** зеркала Draft DP-008 синхронизированы по
  факту implementation status; в зеркалах Draft DP-007 исправлено устаревшее
  утверждение, что neutral handoff не используется Builder. Design Status
  обоих DP не повышался; ADR, ARCH и DP-009 не изменялись.
- **Operational documentation:** этот task record; `docs/tasks/README.md`
  проверен, изменение навигации не требуется.
- **Project-state documentation:** `.ai/PROJECT_CONTEXT.md`,
  `spec/current-state.md`, `docs/en/roadmap/MASTER_PLAN.md` и
  `docs/ru/roadmap/MASTER_PLAN.md`.
- **CHANGELOG:** факт изолированной реализации Builder добавлен без заявления
  о production launch pipeline.
- **Documentation evidence:** `git diff --check`, trailing whitespace и
  conflict-marker checks — PASS; repository Markdown links и fences — PASS
  для 97 tracked файлов; DP-008 heading hierarchy — 74/74 с учётом H1; все 93
  Code/Location/Message rows Diagnostics byte-identical в EN/RU; MASTER_PLAN
  heading hierarchy — 35/35; `markdownlint` недоступен в среде.
- **Post-implementation synchronization:** выполнена; Design Status и
  Implementation Status разделены.
- **Unresolved risk:** race detector не запущен из-за отсутствия CGO/gcc в
  среде. Production launch pipeline, operational identities как сущности,
  Runtime Launcher, Runtime Lifecycle Owner и DP-009 остаются вне scope и не
  реализованы.
- **Current status:** Completed — Coordinator Accepted.

## Closure

- **Architecture refinement status:** Approved; blocker resolved.
- **Implementation review status:** Approved after test-only rework.
- **Documentation synchronization status:** Synchronized and independently
  re-reviewed.
- **Coordinator acceptance:** Accepted.
- **Task closure:** Completed; обязательный repository gate завершён успешно.
- **Commit:** не создаётся без отдельного явного разрешения.

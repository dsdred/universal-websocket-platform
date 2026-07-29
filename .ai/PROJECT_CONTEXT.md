# Контекст проекта

## Основная информация

**Проект:** Universal WebSocket Platform

**Миссия:** Open-source платформа для создания, настройки, развертывания и эксплуатации независимых WebSocket-серверов без написания инфраструктурного кода.

## Текущее состояние

- Текущая веха: **Beta — Complete the Single-Node Runtime**
- Статус реализации: **Control Service, single-node Runtime vertical,
  Configuration Loader boundary, DP-008 Snapshot Builder, DP-009 Runtime
  Bootstrap, stateless Runtime Launcher и DP-010 Runtime Lifecycle Owner
  реализованы изолированно; Draft DP-011 и `internal/runtimelaunchflow`
  реализуют integration contract изолированно; Draft DP-012 и
  `internal/configurationloadsource.MemorySource` реализуют concrete Source
  adapter изолированно; Completed Design-only TASK-015 и Draft/Planned DP-013
  определяют management routing contract; Completed Design-only TASK-016 и
  non-normative Draft/Planned DP-014 предлагают durable operational identity
  persistence candidate contract. Implementation, HTTP, persistence
  package/schema, management wiring и Production Activation отсутствуют**
- Последняя завершённая development task: **TASK-014 — Runtime Source
  Implementation; Completed — Coordinator Accepted**
- Последняя завершённая operational task: **TASK-012 — Engineering Process
  Hardening; Completed — Coordinator Accepted**
- Текущая operational task: **отсутствует**
- Последняя завершённая architecture task: **TASK-016 — Runtime Operational
  Identity Persistence Design; Completed — Coordinator Accepted**
- Текущая architecture task: **отсутствует**
- Текущая development task: **отсутствует**
- Trusted baseline TASK-009: **clean synchronized
  `main@63b961eeb59af9205c3c3d0b68d3f4bd7b8ac25c`; локальная ветка
  `feature/task-009-runtime-lifecycle-owner`; task record создан первым
  content change**
- Trusted baseline TASK-008: **TASK-007 task commit `2e6d221` опубликован через PR #8 и merged в clean `main` commit `802760a`; TASK-008 начата от этого baseline, task record создан первым content change**
- Verification TASK-001: **targeted tests PASS 3/3; full `go test ./... -count=1` PASS 2/2; `go vet ./...`, `gofmt -d` и `git diff --check` PASS; race detector недоступен без CGO/gcc**
- Operational entry после принятия TASK-002: **точная команда `Продолжай проект.` запускает repository-native selection и полный PROCESS-001 cycle без неявных commit, push или merge**
- Verification TASK-002: **Tester PASS; Reviewer Approved after rework; scope audit accepted — 8 Required, 0 Questionable, 0 Removable**
- Результат TASK-003: **implementation prerequisites Draft DP-009 завершены; Tester PASS after rework; Reviewer final closure Approved; Coordinator scope audit accepted — 6 Required, 0 Questionable, 0 Removable**
- Результат TASK-004: **isolated Runtime Bootstrap реализован и принят; Tester PASS; Reviewer Approved with Findings, blocking 0; scope audit accepted — 12 Required, 0 Questionable, 0 Removable**
- Verification TASK-004: **targeted и full tests, vet, gofmt, documentation links/parity и diff checks PASS; race detector недоступен без CGO/gcc**
- Результат TASK-005: **planned in-process `func Launch(request *BootstrapRequest) BootstrapOutcome` contract зеркально уточнён в Draft DP-009: borrowed same pointer, ровно одна delegation в реализованный Bootstrap, unchanged outcome/Host/failure identities и cause chain, ownership handoff будущему Lifecycle Owner, zero policy/cleanup/state и AP-003/AP-011 local-vs-integration proof split; Completed — Coordinator Accepted; production code отсутствует**
- Verification TASK-005: **Tester PASS; PROCESS-002 Synchronized; Final Reviewer Approved, 0 blocking и 0 nonblocking findings; Coordinator Scope Audit accepted — 6 Required, 0 Questionable, 0 Removable; Coordinator Acceptance получена**
- Результат TASK-006: **`internal/runtime.Launch` реализован как exact stateless `return Bootstrap(request)` без adapter, state, validation, wrapping, cleanup, retry или lifecycle policy; Lifecycle Owner и production wiring отсутствуют**
- Verification TASK-006: **targeted и full Go tests, `go vet`, `gofmt -d`, EN/RU structure/status parity и diff checks PASS; race detector недоступен при `CGO_ENABLED=0` и отсутствующем `gcc`; Final Reviewer Approved, 0 blocking и 0 nonblocking findings; Scope Audit accepted — 8 Required, 0 Questionable, 0 Removable**
- Результат TASK-007: **зеркальный Draft DP-010 с Implementation Status Planned фиксирует минимальный `internal/runtimelifecycle` contract: Owner-bound Workspace/Configuration/Runtime Instance, Owner-issued Launch Attempt и exact ConfigurationVersion pin в `PrepareStart` до Loader/Builder, closed `PreparationResult`, first-valid-result-wins, origin-sensitive Stop, truthful immutable outcomes/observation и local-vs-integration proofs; production code отсутствует**
- Verification TASK-007: **после rework B-001/B-002, R-001/R-002 и project-state correction F-001 Final Reviewer выдал Approved, 0 blocking и 0 nonblocking findings; Final Tester PASS; PROCESS-002 Synchronized; Coordinator Scope Audit accepted — 8 Required, 0 Questionable, 0 Removable; Coordinator Acceptance получена**
- Operational governance TASK-008: **точная команда `Разрешаю публиковать.` документирована как единое разрешение P0–P10 для одного immutable task target; Initial P0 отделён от phase-aware Resume Reconstruction Guard, push/merge являются checkpoints, external blocker сохраняет authority, post-P6 resume остаётся на `main`, No CI допускается только при `MERGEABLE / CLEAN`, cleanup и terminal payload обязательны; R-001/R-002 устранены, Final Reviewer Approved 0/0, Tester PASS, PROCESS-002 Synchronized, Coordinator Scope Audit accepted — 14 Required, 0 Questionable, 0 Removable; Coordinator Acceptance получена**
- Product impact TASK-008: **отсутствует; production code/tests, `.github`, ADR/ARCH/DP, product capability и Runtime implementation не изменены**
- Результат TASK-009: **добавлен изолированный
  `internal/runtimelifecycle`: Owner-bound identities, Owner-issued Launch
  Attempt и exact version pin через `PrepareStart`, closed preparation result,
  single tracked Launcher/Host Stop operations, same-token convergence,
  origin-sensitive truthful outcomes, cancellation-only caller waits и
  coherent observation; Bootstrap, Launcher, Host и production wiring не
  изменены**
- Verification TASK-009: **targeted и full `go test ./... -count=1`,
  stress `-count=100`, `go vet ./...`, `go fmt ./...`,
  `git diff --check`, EN/RU parity и link validation PASS; race detector
  недоступен при `CGO_ENABLED=0` и отсутствующем `gcc`; independent Tester
  PASS; Scope Audit accepted — 14 Required, 0 Questionable, 0 Removable;
  Final Reviewer Approved, 0 blocking и 0 nonblocking findings; Coordinator
  Acceptance получена**
- Closure publication state TASK-008: **на момент closure stage, commit и publication не выполнялись; это historical fact, а не live gate. Любая последующая разрешённая публикация reconstruct-ит фактическое состояние из Git/GitHub и не хранит transient dirty/push/PR/cleanup state в project context**
- Следующая work после TASK-009: **активирована как documentation-first
  TASK-010; production implementation Loader-to-Builder-to-Launcher,
  persistence, management API, retry/reconciliation и supervision остаются
  отдельной work и требуют собственного readiness/contract решения**
- Результат TASK-010: **создан и независимо reviewed зеркальный Draft DP-011
  с Implementation Status Planned; он определяет immutable
  `internal/runtimelaunchflow`, synchronous Start Operation, Caller
  Cancellation Gate и точный `PrepareStart -> Load -> Build -> Start`
  contract без production code, Source composition, management API,
  persistence или Production Activation**
- Verification TASK-010: **EN/RU 33/33 headings и 14/14 fences; broken links
  0; `git diff --check` PASS; Scope Audit 11 Required, 0 Questionable,
  0 Removable; repeat Independent Reviewer Approved 0/0**
- Следующий candidate после TASK-010: **не активирован; минимальная
  implementation `internal/runtimelaunchflow` и local proof tests требуют
  нового task intake, а Source adapter, Control Service routing, persistence
  и Production Activation остаются вне slice**
- TASK-011: **isolated `internal/runtimelaunchflow` реализует immutable
  Owner/Loader binding, Caller Cancellation Gate, synchronous
  `PrepareStart -> Load -> Build -> Start`, exact Loader failures, immutable
  Build Failure и Stop convergence; Completed — Coordinator Accepted,
  Production Activation отсутствует**
- Verification TASK-011: **independent Tester PASS и Final Reviewer Approved,
  0 blocking и 0 nonblocking findings; targeted stress `-count=100`, affected
  и full tests, vet, exported-surface, formatting, diff, EN/RU parity и links
  PASS; PROCESS-002 Synchronized; Scope Audit accepted — 13 Required,
  0 Questionable, 0 Removable; race detector недоступен при `CGO_ENABLED=0`
  и отсутствующем `gcc`**
- Operational governance TASK-012: **Task Contract, Existing Coverage Report,
  risk-oriented Verification Matrix, Size Guard, strengthened Scope Audit,
  mandatory Documentation Sync, exact Commit Gate, Publisher hardening и
  lightweight Process Health Review встроены в существующий
  Coordinator/Publisher workflow; три permission/gate-команды не изменены,
  production code и Production Activation не затронуты; independent Tester
  PASS, Final Reviewer Approved, Scope Audit 16 Required / 0 Questionable /
  0 Removable; Completed — Coordinator Accepted**
- Следующий candidate после TASK-011: **не активирован; рекомендуется отдельная
  documentation-first readiness/design task, выбирающая один prerequisite
  Production Activation между concrete Source composition, management routing
  и persistence boundary**
- TASK-013: **design-only task завершена и принята Coordinator; зеркальный
  Draft DP-012 с
  Implementation Status Planned определяет только planned
  `internal/configurationloadsource.MemorySource`, exact repository lookup,
  mandatory composition confinement, detachment, error mapping и будущий
  construction `Source -> Loader -> Flow`; production code, management
  routing, persistence и Production Activation не изменены; final Tester PASS,
  Final Reviewer Approved 0/0, Scope Audit accepted 10/0/0, PROCESS-002
  Synchronized**
- Candidate после TASK-013: **последовательно активирован как TASK-014 и
  завершён; management routing, persistence и Production Activation остаются
  более поздней work**
- TASK-014: **`internal/configurationloadsource.MemorySource` реализован
  изолированно с exact Version-first lookup, closed Loader errors, static
  schema facts, deep detachment, local proof tests, Loader integration и
  construction proof без Start/Host; после B-001 rework repeat Tester —
  PASS WITH LIMITATION, 0 findings; race недоступен при CGO=0 и отсутствии
  gcc, substitute stress PASS; после R-001 test-only rework Repeat Final
  Reviewer Approved 0/0; Scope Audit accepted 12/0/0; PROCESS-002
  Synchronized; Completed — Coordinator Accepted**
- Candidate после TASK-014: **активирован и завершён как Design-only TASK-015
  — Runtime Management Routing Design**
- TASK-015: **Completed — Coordinator Accepted; Architecture Confirmation
  Design READY/valid, design blockers 0; зеркальный Draft
  DP-013 с Implementation Status Planned определяет immutable process-local
  `internal/runtimemanagement.Directory`, exact Target routing,
  policy-neutral named `Authorize`, authorization-before-mutation и static
  Owner-to-Flow binding. Design contract READY/valid, но Implementation
  Readiness BLOCKED обязательными focused designs ARCH-004 §19(2)–(6):
  operational identity persistence, durable management idempotency,
  activation/replacement/rollback, recovery/reconciliation и
  reporting/redaction. Production package, HTTP API/DTO, concrete policy,
  persistence, recovery, application wiring и Production Activation
  отсутствуют; Initial Tester FAIL B-001/B-002 -> bounded rework -> Repeat
  PASS 0/0; Initial Final Reviewer Needs Revision R-001/R-002 -> bounded
  Architect/Documentation rework -> Repeat Approved 0/0; Scope Audit
  11 Required / 0 Questionable / 0 Removable; PROCESS-002 Synchronized; code,
  commit и publication отсутствуют**
- Candidate после TASK-015: **активирован как Design-only TASK-016 — Runtime
  Operational Identity Persistence Design**
- TASK-016: **Completed — Coordinator Accepted; Architecture Analysis Ready,
  blockers 0;
  зеркальный non-normative Draft DP-014 с Implementation Status Planned
  предлагает candidate contract ARCH-004 §19(2): immutable aggregate Runtime
  Instance, append-only membership Launch Attempt с monotonic child facts,
  opaque identity namespaces, conditional revision, atomic phase-sensitive
  publication и inspect-after-indeterminate boundary. Initial Reviewer —
  Needs Revision: R-001/R-002/R-003/N-001; bounded Architect/Documentation
  rework завершён; Repeat Tester PASS; Repeat Reviewer и Final Reviewer
  Approved 0/0; Scope Audit accepted 13/0/0; PROCESS-002 Synchronized. Exact
  scope 13 files; DP-014 27/27 headings и 4/4 fences; DP-013 35/35;
  MASTER_PLAN 36/36; changed links 152/0, repository links 753/0; diff check
  PASS; preceding `go test ./...` и `go vet ./...` PASS reused после doc-only
  rework. §19(2) остаётся formal implementation blocker до отдельного
  approval/status decision вместе с §19(3)–(6). Persistence
  package/schema/API, recovery и production wiring отсутствуют; commit и
  publication не авторизованы и не выполнялись**
- Следующая рекомендация после TASK-016: **не активирована; отдельный
  Design-only durable management command idempotency contract ARCH-004
  §19(3). Он не снимает formal implementation gate §19(2)**
- Stage 2 task-before-work ordering выполнен для TASK-003, TASK-004, TASK-005, TASK-006 и TASK-007: task record создан первым content change, а task index обновлён только после initial gate
- Publication history: **TASK-005 commit `99e0d3d`, TASK-006 commit `fd0f80a` и TASK-007 commit `2e6d221` merged через PR #6/#7/#8; transient pre-commit/Publisher blockers не являются durable project-state instructions**
- Design Status DP-009 остаётся **Draft**; Bootstrap и Runtime Launcher
  Implementation Status — **Implemented in isolation**; production launch
  pipeline не реализован, AP-003 и AP-011 остаются integration-gated
- Design Status DP-010 остаётся **Draft**, Implementation Status —
  **Implemented in isolation**; status не утверждает production wiring,
  persistence или management capability
- Design Status DP-011 остаётся **Draft**, Implementation Status —
  **Implemented in isolation**; Flow package не утверждает concrete Source
  composition, management routing или Production Activation
- Design Status DP-012 — **Draft**, Implementation Status — **Implemented in
  isolation**; repository-backed Source adapter реализован, production
  composition отсутствует
- Design Status DP-013 — **Draft**, Implementation Status — **Planned**;
  management routing определён только на design level; Implementation
  Readiness **Blocked** prerequisites ARCH-004 §19(2)–(6),
  implementation и production wiring отсутствуют
- Design Status DP-014 — **Draft**, Implementation Status — **Planned**;
  durable operational identity определена только на design level; repository,
  schema, API, recovery и production wiring отсутствуют
- Design Status DP-008 остаётся **Draft**, Implementation Status — **Implemented in isolation**
- Содержимое репозитория: документация, спецификации, инженерные соглашения, исполняемый Control Service и изолированные Runtime-компоненты с тестами

## Архитектурные принципы

1. Configuration over Code
2. Runtime Isolation
3. API First
4. Technology Neutrality
5. Provider-based architecture
6. Explainability
7. Predictability
8. Keep MVP Simple

## Источники истины

- `spec/00-product/vision.md` определяет замысел продукта.
- `spec/01-principles/architecture-principles.md` определяет архитектурные ориентиры.
- `spec/current-state.md` фиксирует текущее состояние проекта.
- `spec/decisions.md` содержит перечень принятых и ожидающих принятия решений.
- `docs/en/adr/` и `docs/ru/adr/` содержат публичные записи об архитектурных решениях.

Не делайте вывод о реализованных возможностях на основании миссии или спецификаций, описывающих будущее состояние. Перед изменением репозитория начните с корневого `AGENTS.md`, затем следуйте `docs/engineering/AGENT.md` и сверяйтесь с `spec/current-state.md`.

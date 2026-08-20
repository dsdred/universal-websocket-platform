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
  реализуют base integration contract и managed Start-claim continuation
  изолированно; TASK-037 реализует managed gates и execution-binding/load
  sequence изолированно; TASK-043 реализует concrete exact-scope composition
  invoker изолированно;
  Draft DP-012 и
  `internal/configurationloadsource.MemorySource` реализуют concrete Source
  adapter изолированно; Draft DP-013 и `internal/runtimemanagement` реализуют
  management routing изолированно; Approved DP-014 и `internal/runtimeidentity`
  реализуют in-memory Runtime Instance aggregate store изолированно; Approved
  DP-015 и `internal/runtimecommandidempotency` реализуют primitive command
  claim/replay boundary изолированно. Approved/Planned DP-016–DP-018 закрывают focused
  design gates ARCH-004 §19(4)–(6) для activation/replacement/rollback,
  recovery/reconciliation и operational error reporting/redaction. Approved/
  Planned overall DP-019 определяет parent/phase, authorization и private
  Start-claim continuation prerequisites DP-016. TASK-028 реализует partial
  parent/phase sequential core, а TASK-029 — command-boundary Continue и
  pending-Stop rendezvous изолированно; TASK-031 и TASK-032 дали исторически
  принятые partial isolated authorization и managed Flow/OwnerClaimView seams.
  TASK-034 определила managed-binding conformance repair, а TASK-035 реализует
  Slice 2R изолированно: `OperationalDomain`, полный binding/linked identity,
  unique callback-scoped command-owned rendezvous identity, dependency-leaf
  package и sole primitive `Boundary.ExecuteManagedStart` adapter; TASK-035
  независимо принята. TASK-036
  уточнила единый primitive/linked command-gate и continuation protocol.
  TASK-037 реализует Slice 3 изолированно: managed parent/StartTarget path,
  общие early/final/no-claim gates, stateless continuation, DP-014 conditional
  claim/bind с revision threading и managed Flow outcomes; independent code
  proof PASS 0/0, Independent Reviewer APPROVED 0/0, Coordinator Acceptance
  завершена. TASK-043 реализует concrete private composition invoker
  изолированно; future callback/terminal publication, orchestration и
  production wiring отсутствуют. Slice 4 активирован как design/readiness TASK-038; его verdict
  `TASK-026 REMAINS BLOCKED` после Reviewer B-001/B-002 rework принят
  Coordinator; repeat Reviewer APPROVED 0/0. Первый bounded candidate
  завершён как design-update TASK-039 с Coordinator Acceptance: принятый Draft
  DP-010 contract atomic expected-attempt Owner Stop зафиксирован. TASK-040
  завершена как `Completed — Coordinator Accepted`: contract реализован и
  верифицирован изолированно, repeat final Reviewer APPROVED 0/0. TASK-041
  завершена как `Completed — Coordinator Accepted (2026-08-20)`: live status
  managed continuation/binding синхронизирован, final Reviewer APPROVED 0/0.
  Design-only TASK-042 завершена как `Completed — Coordinator Accepted
  (2026-08-20)` и фиксирует Draft DP-021 для private exact-scope invoker;
  TASK-043 реализует его изолированно и завершена как `Completed — Coordinator
  Accepted (2026-08-21)` после final Reviewer APPROVED 0/0 и Scope Audit
  21/0/0.
  HTTP, concrete policy, external command storage, recovery/reporting
  package/schema, management wiring и Production Activation отсутствуют**
- Последняя завершённая development task: **TASK-043 — Private Exact-Scope
  Managed Start Invoker Implementation; Completed — Coordinator Accepted
  (2026-08-21); final Reviewer APPROVED 0/0; Scope Audit 21/0/0; PROCESS-002
  Synchronized**
- Последняя завершённая operational task: **TASK-012 — Engineering Process
  Hardening; Completed — Coordinator Accepted**
- Текущая operational task: **отсутствует**
- Последняя завершённая architecture task: **TASK-042 — Private Exact-Scope
  Composition Invoker Design; Completed — Coordinator Accepted (2026-08-20);
  Tester PASS 0/0; repeat Reviewer APPROVED 0/0; Scope Audit 17/0/0;
  PROCESS-002 Synchronized; Draft DP-021 зафиксирован, а TASK-043 реализует
  его concrete invoker изолированно и завершена/принята**
- Текущая architecture task: **отсутствует**
- TASK-027 acceptance evidence: **DP-019 Approved/Planned; implementable
  parent/phase, exact authorization and private per-call managed Start seams
  defined; repeat Independent Review Approved with blocking 0 and non-blocking
  0; Verification Matrix, PROCESS-002, status consistency and Scope Audit
  24/0/0 PASS; TASK-026 remains Blocked by Architecture; at Coordinator
  closure commit and publication had not yet been performed. Subsequently task
  commit `7ac0a6b372d9e54c73d024703e6d3ee4b06e15cd` was published through PR #27
  and merged as `2c017aace7e56a4747d3cecbe8ff3f6cf53e009f`**
- TASK-025 acceptance evidence: **Runtime Management Command Idempotency
  Implementation; Completed — Coordinator Accepted;
  initial independent
  Reviewer: 4 blocking findings — abandoned permit tracking, stale Boundary
  claim after reconstruction, critical DP-014 §25 drift и incomplete
  post-claim/lost-permit/stale-client proofs; rework B-001–B-004 завершён:
  synchronous private permit, atomic stale-client admission, DP-014 EN/RU sync
  и новые regression proofs; repeat Reviewer подтвердил B-001–B-004 resolved,
  но вернул RR-B-001–RR-B-003 по stale DP-013 EN/RU, exported permit godoc и
  отсутствующей README applicability record; Acceptance/closure запрещены до
  second rework и нового Approved independent review; second rework завершён:
  DP-013 EN/RU synchronized, godoc corrected, README applicability recorded,
  Verification/PROCESS-002 PASS WITH LIMITATION и provisional Scope Audit
  16/0/0; third Independent Reviewer подтвердил code/proofs, но вернул Critical
  IR3-B-001 по residual stale status contradictions в earlier sections
  DP-013/DP-014/DP-015 EN/RU и `spec/decisions.md`, плюс Low IR3-N-001 grammar
  MASTER_PLAN EN; final documentation cleanup синхронизировал все live
  DP-013/014/015 EN/RU sections, MASTER_PLAN и project-state sources, исправил
  grammar и добавил актуальную DP-015 summary; links 852/0, parity/status,
  full/stress/vet/gofmt/module/diff checks и Scope Audit 16/0/0 PASS; race
  unavailable без `gcc`; fourth Independent Reviewer вынес Needs Revision:
  FIR-B-001 — `runtime.Goexit` сохраняет lost permit falsely tracked,
  FIR-B-002 — stale design indexes и live DP-016/DP-017 status wording вне
  16-file sync, FIR-B-003 — residual `spec/decisions.md` contradiction;
  fifth rework устранил FIR-B-001 defer-based cleanup и добавил
  `runtime.Goexit` proof; repository-wide sweep 32 documents/22 live/10
  historical синхронизировал design indexes, DP-013/016/017/018 EN/RU,
  `spec/decisions.md` и bookkeeping; Verification/PROCESS-002/status assertions
  PASS; два interrupted read-only reviews нашли и same-slice rework устранил
  generic drift в `spec/current-state` и DP-011 EN/RU; current Scope Audit
  26/0/0, race limited без `gcc`; задача передаётся новому independent
  Reviewer вынес Approved, blocking 0: FIR-B-001, DP-015 proofs, 32/37-document
  status sweep, PROCESS-002, Verification Matrix и Scope Audit 26/0/0 PASS;
  задача передана Coordinator для отдельного Closure Audit / Acceptance;
  Coordinator Closure Audit PASS; Task Contract, exact scope 26/0/0,
  Verification Matrix, PROCESS-002, status consistency и repository-state
  audit подтверждены; Commit Gate, commit, push и publication не выполнялись**
- Текущая development task: **отсутствует. Последняя завершённая development
  task — TASK-043 — Private Exact-Scope Managed Start Invoker Implementation;
  Completed — Coordinator Accepted (2026-08-21). Concrete invoker DP-021
  реализован изолированно; final Reviewer APPROVED 0/0; Scope Audit 21/0/0;
  PROCESS-002 Synchronized. Future callback/terminal publication, orchestrator
  и production wiring отсутствуют; TASK-026 остаётся Blocked by Architecture**
- TASK-032 acceptance evidence: **Completed — Coordinator Accepted после
  rework; DP-020 deferred slice 2 реализован изолированно в
  `internal/runtimelaunchflow`: ManagedFlow/NewManaged/StartManaged, immutable
  ManagedStartBinding и OwnerClaimView, stateless StartClaimContinuation и
  neutral opaque `runtimeconfigload.StartRendezvous` handle. Rework устранил
  blocking findings: (R-001) убран импорт `runtimecommandidempotency` в пользу
  `runtimeidentity` Revision/ExecutionGeneration, теперь DP-020 §8.1 соблюдён;
  (R-002) при ошибке continuation claimed LaunchPreparation корректно
  конвергируется через `Owner.Start(FailedPreparation(err))`, и exact error
  возвращается unchanged. Independent Review rework: Approved 0 blocking /
  4 non-blocking. Verification Matrix (focused/full/stress/shuffled, vet,
  mod-tidy, diff-check) — PASS; race ограничен отсутствием CGO/gcc.
  Scope Audit — 7 Required / 0 Questionable / 0 Removable. On branch
  `feature/task-032-runtime-private-managed-invoker` от baseline
  `main@07b27ce`; на момент Coordinator closure commit, push и publication не
  выполнялись. Впоследствии task commit
  `577e1ced0a984952396238cc94bdcbec80c2a6d4` опубликован через PR #32 и merged
  как `74e55a6d9a14502f134cbf20eb53359fd9abc995`. Closure-time рекомендация
  Slice 3 сохранена как исторический факт; TASK-034 позднее установила, что до
  неё требуется отдельный Slice 2R conformance repair**
- TASK-031 acceptance evidence: **Completed — Coordinator Accepted; DP-020
  deferred slice 1 реализован изолированно в
  `internal/runtimecommandidempotency`: immutable validated
  `OrchestrationAuthorizationRequest`, exact `OrchestrationAction` set и
  policy-neutral `AuthorizeOrchestration` function type с per-call evaluation
  без кэша, fixed Start→`ActivateExactTarget` adaptation и fail-closed
  error semantics. Independent Tester — PASS WITH ENVIRONMENT LIMITATION
  (race недоступен без CGO/gcc, substitute stress `-count=100` PASS);
  Independent Review — Approved with Findings 0 blocking / 2 non-blocking
  (resolved by R-001/R-002). Verification Matrix, PROCESS-002 и Scope Audit
  7 Required / 0 Questionable / 0 Removable — PASS. On branch
  `feature/task-031-runtime-orchestration-authorization-surface` от baseline
  `main@0bfbca3`; commit, push, publication и branch cleanup не выполнялись.
  Следующая рекомендация — intake DP-020 deferred slice 2 (private managed
  invoker plus managed Flow/OwnerClaimView continuation); не активирована**
- TASK-030 acceptance evidence: **Completed — Coordinator Accepted; design-only
  task создала зеркальный Draft/Planned DP-020, закрыла отложенные design
  decisions, не меняла Approved статусы/семантики и не реализовала ни один
  implementation slice. Independent Review — Approved 0 blocking / 1
  non-blocking (resolved by R-001). Verification Matrix — PASS по всем
  применимым строкам; race Not applicable (no concurrency code). Scope Audit —
  11 Required / 0 Questionable / 0 Removable. On branch
  `docs/task-030-runtime-orchestration-binding-readiness` от baseline
  `main@4a66fa3`; commit, push, publication и branch cleanup не выполнялись.
  Следующая рекомендация — intake DP-020 deferred slice 1 (orchestration
  authorizer surface); не активирована**
- TASK-029 acceptance evidence: **Completed — Coordinator Accepted;
  Architecture PASS, blocking 0; Size Guard ACCEPT,
  `DO NOT SPLIT`, net production `+680`; Independent Tester PASS WITH
  ENVIRONMENT LIMITATION, blocking/non-blocking 0; focused coverage 85.9%,
  focused/package/shuffled `-count=100`, full tests, vet, GoDoc и diff checks
  PASS; race build unavailable без CGO/gcc; Independent Review APPROVED 0/0;
  PROCESS-002/status/parity PASS, links 886/0, Scope Audit 25/0/0,
  staged/unexpected 0; Coordinator Closure Audit PASS and Acceptance
  `Accepted`; branch baseline
  `ba75e54e00c3cf1d0d87ca2a985acc9699698efd`; no task commit, push, or
  publication. Exact authorization/private
  invoker, managed Flow/OwnerClaimView, DP-014 binding, orchestrator и
  production composition остаются Planned; TASK-026 remains Blocked**
- TASK-041 acceptance evidence: **Managed Continuation Documentation Baseline
  Reconciliation; Completed — Coordinator Accepted (2026-08-20); PROCESS-002
  Synchronized; final Reviewer APPROVED 0/0; Scope Audit 15/0/0; commit и
  publication не авторизованы и не выполнялись**
- TASK-042 acceptance evidence: **Private Exact-Scope Composition Invoker
  Design; Completed — Coordinator Accepted (2026-08-20); Tester PASS 0/0;
  repeat Reviewer APPROVED 0/0; Scope Audit 17/0/0; PROCESS-002 Synchronized;
  на момент closure commit и publication не выполнялись; впоследствии task
  commit `ebf4421` опубликован через PR #42 и merged как `ded3aa0`**
- TASK-043 acceptance evidence: **Private Exact-Scope Managed Start Invoker
  Implementation; Completed — Coordinator Accepted (2026-08-21); Draft DP-021
  имеет Implementation Status Partial/implemented in isolation; Tester PASS
  WITH ENVIRONMENT / DECLARED INTEGRATION LIMITATIONS 0/0; final Reviewer
  APPROVED 0/0; Scope Audit 21/0/0; PROCESS-002 Synchronized; commit и
  publication не авторизованы и не выполнялись**
- Следующая рекомендация после TASK-043: **не активирована; отдельный bounded
  repository-first readiness intake remaining TASK-026 terminal/orchestrator
  work. Ready status не доказан; TASK-026 остаётся Blocked**
- TASK-028 acceptance evidence: **partial DP-019 durable parent/derived-phase
  storage, callback capability и sequential phase core реализованы
  изолированно; Repeat Independent Review Approved, blocking/non-blocking 0;
  Verification Matrix PASS WITH ENVIRONMENT LIMITATION, PROCESS-002/status
  consistency PASS, Scope Audit 24/0/0; на closure-time Continue/pending-Stop,
  managed Flow continuation, binding и production wiring оставались Planned;
  TASK-029 впоследствии реализует только первый prerequisite; at Coordinator
  closure commit и publication ещё не выполнялись. Subsequently task commit
  `d28efa4e88e02ef528c78c3ca88b3f91945069ce` was published through PR #28
  and merged as `ba75e54e00c3cf1d0d87ca2a985acc9699698efd`**
- Последняя завершённая documentation task: **TASK-041 — Managed Continuation
  Documentation Baseline Reconciliation; Completed — Coordinator Accepted
  (2026-08-20); PROCESS-002 Synchronized; final Reviewer APPROVED 0/0; Scope
  Audit 15/0/0**
- Текущая documentation task: **отсутствует**
- Текущая architecture task: **отсутствует; TASK-026 остаётся Blocked**
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
- Candidate после TASK-016: **активирован как Design-only TASK-017 — Runtime
  Management Command Idempotency Design**
- TASK-017: **Completed — Coordinator Accepted; зеркальный non-normative Draft
  DP-015 с
  Implementation Status Planned предлагает candidate contract ARCH-004
  §19(3): exact authorized command scope, opaque key, immutable intent, durable
  claim до lifecycle delegation, same-intent non-mutating replay, durable
  per-Instance barrier при unresolved outcome, mandatory tracked-Start Stop и
  truthful indeterminate outcomes. Initial Reviewer Needs Revision B-001/B-002;
  first rework; Repeat Reviewer Needs Revision B-003; second rework; Final
  Reviewer Approved with Findings, blocking 0; closure bookkeeping; terminal
  Reviewer Approved 0/0. Full `go test ./... -count=1`, `go vet ./...`, EN/RU
  29/29 headings, links 770/0 и diff checks PASS; Scope Audit 15/0/0;
  PROCESS-002 Synchronized. Formal gates §19(2) и §19(3),
  downstream designs §19(4)–(6), implementation и production wiring остаются
  открытыми**
- Candidate после TASK-017: **активирован как Design-only TASK-018 — Runtime
  Activation, Replacement, and Rollback Design**
- TASK-018: **Completed — Coordinator Accepted; зеркальный
  non-normative Draft/Planned DP-016 предлагает candidate contract ARCH-004
  §19(4): exact-version activation, Stop-to-proven-release replacement,
  fresh-attempt explicit rollback, zero Host overlap и phase-specific
  concurrency/cancellation с planned private Start-claim continuation
  DP-011/DP-013 после Owner claim и до Load. Current Flow seam не реализует.
  Review rework B-001–B-005 завершён; terminal Reviewer Approved 0/0; full Go
  regression, vet, parity, links и diff checks PASS; Scope Audit 19/0/0;
  PROCESS-002 Synchronized; Process Health Review complete. Formal gates
  §19(2)–(4), downstream §19(5)–(6), implementation и Production Activation
  остаются открытыми**
- Следующая рекомендация после TASK-018: **предварительно §19(5) recovery и
  reconciliation после termination Control Service; активирована как
  Design-only TASK-019**
- TASK-018 publication: **task commit `64e1fe7` опубликован через PR #19 и
  merged в clean `main` commit `d083957`; task branch удалена после exact OID
  verification**
- TASK-019: **Completed — Coordinator Accepted; зеркальный non-normative
  Draft/Planned DP-017
  предлагает candidate contract ARCH-004 §19(5): exact fail-closed assessment,
  durable recovery claim, DP-014-owned attempt/generation binding до Load,
  attempt/generation-bound execution evidence, phase-sensitive reconciliation
  без lifecycle replay и reopening admission только для coherent fully terminal
  set. Resource absence даёт Failed/interrupted; Stopped требует exact Host
  shutdown-completion proof. Review rework B-001–B-005 и residual wording
  завершён; Final Confirmation Reviewer Approved 0/0; full Go regression, vet,
  parity, 247/0 links и diff checks PASS; Scope Audit accepted 21/0/0;
  PROCESS-002 Synchronized; на момент Coordinator closure commit и publication
  не выполнялись; последующая commit permission фиксируется task record**
- Следующая рекомендация после TASK-019: **operational error reporting и
  redaction ARCH-004 §19(6); активирована и завершена как Design-only
  TASK-020**
- TASK-020: **Completed — Coordinator Accepted; зеркальный non-normative
  Draft/Planned DP-018 предлагает candidate contract ARCH-004 §19(6): exact
  failure ownership, authoritative-fact-first projection, valid-negative/error
  separation, owner/phase category precedence, scoped opaque correlation,
  fail-closed allowlist redaction, projection-version replay и downstream
  delivery failure isolation. Review B-001–B-005 и matrix clarity rework
  завершён; Terminal Design Reviewer Approved 0/0; final regression, vet,
  parity, links и diff checks PASS; Scope Audit accepted 21/0/0; Repeat terminal
  Closure Review Approved 0/0; PROCESS-002 Synchronized. Reporting
  model/projector/adapter, management implementation и Production Activation
  отсутствуют**
- TASK-021: **Completed — Coordinator Accepted; DP-014–DP-018 имеют Design
  Status Approved и Implementation Status Planned; focused design gates
  ARCH-004 §19(2)–(6) закрыты; Draft/Planned DP-013 Ready для bounded isolated
  implementation slice. Tester PASS; independent Reviewer Approved 0/0;
  PROCESS-002 Synchronized; Scope Audit 21/0/0. Commit и publication не
  выполнялись**
- TASK-022: **Completed — Coordinator Accepted; root `README.md` и
  `README.ru.md` теперь правдиво отражают связь Configuration Loader со
  Snapshot Builder через isolated in-process Runtime Launch Flow и
  доказанную isolated конструкцию `Source -> Loader -> Flow` поверх concrete
  in-memory Source adapter. Application/Control Service wiring и Production
  Activation остаются отсутствующими. Tester PASS 0 findings; independent
  Reviewer Approved 0/0; PROCESS-002 Synchronized; Scope Audit 6/0/0. На
  момент closure commit и publication не выполнялись. Bounded isolated DP-013
  implementation Ready/recommended, но не активирована**
- TASK-023: **Completed — Coordinator Accepted;
  `internal/runtimemanagement` реализует exact
  Target/Binding/Directory surface Draft DP-013, authorization-before-mutation
  и immutable routing Start к Flow, а Stop/Observe к exact Owner. Focused/full
  tests, focused stress и vet PASS; race detector недоступен без C compiler;
  independent Reviewer Approved 0/0; PROCESS-002 Synchronized; Scope Audit
  23/0/0. Production Control Service routing, concrete policy, persistence,
  recovery, management wiring и Production Activation отсутствуют. Следующий
  DP-014 readiness candidate не активирован**
- TASK-024: **Completed — Coordinator Accepted;
  `internal/runtimeidentity` реализует все девять conceptual operations
  Approved DP-014 §21 и удовлетворяет всем семнадцати acceptance proofs §22
  как isolated in-process in-memory store; 35 proof/regression tests и focused
  stress -count=100 PASS; race detector недоступен без C compiler; full
  regression, vet, gofmt и diff checks PASS; PROCESS-002 Synchronized; Scope
  Audit 12/0/0. External storage, HTTP API, production wiring и Production
  Activation отсутствуют**
- Stage 2 task-before-work ordering выполнен для TASK-003, TASK-004, TASK-005, TASK-006 и TASK-007: task record создан первым content change, а task index обновлён только после initial gate
- Publication history: **TASK-005 commit `99e0d3d`, TASK-006 commit `fd0f80a` и TASK-007 commit `2e6d221` merged через PR #6/#7/#8; TASK-025 commit `06c80265` merged через PR #26 в `751577e8`; transient pre-commit/Publisher blockers не являются durable project-state instructions**
- Design Status DP-009 остаётся **Draft**; Bootstrap и Runtime Launcher
  Implementation Status — **Implemented in isolation**; production launch
  pipeline не реализован, AP-003 и AP-011 остаются integration-gated
- Design Status DP-010 остаётся **Draft**, Implementation Status —
  **Implemented in isolation**; status не утверждает production wiring,
  persistence или management capability
- Design Status DP-011 остаётся **Draft**, Implementation Status —
  **base Flow и managed Start-claim continuation реализованы изолированно**;
  concrete composition invoker, production composition, management routing и
  Production Activation отсутствуют
- Design Status DP-012 — **Draft**, Implementation Status — **Implemented in
  isolation**; repository-backed Source adapter реализован, production
  composition отсутствует
- Design Status DP-013 — **Draft**, Implementation Status — **Implemented in
  isolation**; exact management routing package и local proofs существуют,
  concrete policy, full integration и production wiring отсутствуют
- Design Status DP-014 — **Approved**, Implementation Status — **Implemented in
  isolation**; in-memory Runtime Instance aggregate store `internal/runtimeidentity`
  реализован изолированно; external storage, HTTP API, recovery и production
  wiring отсутствуют
- Design Status DP-015 — **Approved**, primitive Start/Stop boundary, partial
  parent/phase sequential core DP-019 и command-boundary Continue/pending-Stop
  rendezvous — **Implemented in isolation**, полный extension — **Planned**;
  process-local `internal/runtimecommandidempotency` реализует claim/replay
  storage и one-shot permits; external schema, API, recovery, integration и
  production wiring отсутствуют
- Design Status DP-016 — **Approved**, Implementation Status — **Planned**;
  activation/replacement/rollback ordering определён только на design level;
  implementation architecture-blocked unimplemented prerequisites DP-019;
  API, recovery и production wiring отсутствуют
- Design Status DP-017 — **Approved**, Implementation Status — **Planned**;
  recovery/reconciliation определены только на design level; recovery store,
  execution-evidence adapter, executor, API и production wiring отсутствуют
- Design Status DP-018 — **Approved**, Implementation Status — **Planned**;
  operational reporting/redaction определены только на design level; report
  model, projector, delivery adapter, API и production wiring отсутствуют
- Design Status DP-019 — **Approved**, Implementation Status — **Planned
  overall**; parent/phase durable storage, callback capability, sequential core
  и command-boundary Continue/pending-Stop rendezvous реализованы изолированно,
  exact orchestration authorization surface, private invoker и managed Flow/
  OwnerClaimView seams частично реализованы изолированно TASK-031/TASK-032;
  полный managed binding и Slice 2R repair реализованы и независимо приняты
  изолированно TASK-035; TASK-037 реализует managed gates, continuation,
  DP-014 attempt/generation binding sequence и managed Flow outcomes
  изолированно и независимо принята; concrete private invoker,
  policy, terminal publication, orchestrator и production wiring не реализованы
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

# TASK-015 — Runtime Management Routing Design

## Status

**Completed — Coordinator Accepted**

## Task Contract

### Task Mode

Design-only.

### Why Now

TASK-014 завершила concrete in-memory Source и изолированный proof
`Source -> Loader -> Flow`. `spec/current-state.md`,
`.ai/PROJECT_CONTEXT.md`, зеркальные MASTER_PLAN и closure TASK-014
однозначно рекомендуют следующим отдельный management routing
design/readiness slice. Активные task отсутствуют; baseline — чистый
синхронизированный `main@6a03fa91562e080dcc2e316be0462767ae4f88d3`.

### Definition of Done

1. Зеркальный Draft DP определяет ровно одну in-process management command
   boundary для Start, Stop и Observe по Runtime Instance identity.
2. Design фиксирует routing ровно одного immutable Owner/Flow scope на один
   Runtime Instance без service locator, package-global registry или второго
   lifecycle owner.
3. Workspace/Configuration/Runtime Instance identity validation,
   authorization-before-mutation ordering, concurrency, cancellation,
   error/outcome mapping и ownership boundaries заданы однозначно.
4. Design сохраняет точные контракты ARCH-004/ARCH-005 и DP-010–DP-012 и не
   объявляет persistence, recovery, HTTP API, application wiring либо
   Production Activation реализованными.
5. Readiness следующего минимального implementation slice либо обязательный
   blocker доказаны repository evidence.
6. EN/RU parity, navigation, project state, ссылки, whitespace и полный
   documentation diff независимо проверены.

### Out of Scope

- production code и Go tests;
- HTTP paths, JSON DTO и transport-specific status mapping;
- конкретная authentication или authorization policy;
- Runtime Instance/Launch Attempt persistence и durable idempotency;
- recovery, reconciliation, retry, restart, replacement или rollback;
- application composition, вызов `Flow.Start`, Host creation и Production
  Activation;
- operational diagnostics backend, metrics, audit storage или supervision;
- изменение Approved ADR, Active/Frozen ARCH либо существующих package API.

### Verification Plan

- inventory authoritative architecture, related DP, current code surfaces и
  existing tests;
- semantic EN/RU structure/status parity;
- relative-link validation;
- explicit conflict check против ARCH-004/ARCH-005 и DP-010–DP-012;
- `git diff --check`, conflict-marker и trailing-whitespace checks;
- full repository test/vet smoke как regression evidence, хотя production code
  не меняется;
- independent documentation verification и independent Final Review;
- PROCESS-002 и per-file Scope Audit.

## Selection Evidence

- pre-selection task records with current status: отсутствуют;
- current operational/development/architecture task: отсутствует;
- latest completed development task: TASK-014;
- explicit next candidate в TASK-014, `.ai/PROJECT_CONTEXT.md` и
  `spec/current-state.md`: management routing design/readiness;
- MASTER_PLAN prerequisite order: Source composition завершена изолированно,
  management routing остаётся ближайшим незакрытым Production Activation
  gate;
- baseline: clean `main`, `main == origin/main`,
  `6a03fa91562e080dcc2e316be0462767ae4f88d3`.

## Rejected Alternatives

- Persistence design не включён в TASK-015, потому что это второй отдельный
  architecture contract. После фиксации DP-013 он является первым
  детерминированным следующим prerequisite по ARCH-004 §19(2), а не parallel
  implementation work.
- Production Activation implementation отклонена: ARCH-004 требует отдельные
  persistence, idempotency, recovery и reporting contracts.
- HTTP API implementation отклонена: transport и authorization policy ещё не
  спроектированы в пределах bounded management boundary.
- Изменение Runtime Lifecycle Owner, Flow, Source, Loader, Builder, Bootstrap
  или Host отклонено: существующие contracts уже задают необходимые
  downstream boundaries, а task является Design-only.

## Scope

Ожидаемый scope:

- новый зеркальный Draft DP-013;
- зеркальные design indexes;
- task record и task index;
- применимые `.ai/PROJECT_CONTEXT.md`, `spec/current-state.md` и зеркальные
  MASTER_PLAN после фактического результата.

Точный file set уточняется Documentation Baseline и Architecture
Confirmation. Production, test и generated files запрещены.

## Sources of Truth

- ADR-0002, ADR-0003 и ADR-0004;
- ARCH-002, ARCH-004 и ARCH-005;
- DP-007–DP-012, особенно DP-010–DP-012;
- `spec/decisions.md`, `spec/current-state.md`;
- зеркальные MASTER_PLAN;
- фактические `internal/runtimelifecycle`,
  `internal/runtimelaunchflow`, `internal/configurationloader` и
  `internal/configurationloadsource`;
- существующая Control Service composition и HTTP boundary как factual
  evidence, не как design authority.

## Roles

- Coordinator: task selection, gates, Size Guard, Scope Audit и Acceptance.
- Architect: management routing contract, ownership и readiness verdict.
- Documentation Agent: baseline, зеркальный DP и PROCESS-002.
- Tester: documentation verification и regression smoke.
- Reviewer: независимый architecture/documentation review.
- Developer: Not applicable — production code запрещён.
- Publisher: Not applicable до отдельных commit/publication permissions.

## Branch Decision

- branch: `docs/task-015-runtime-management-routing-design`;
- создана от clean synchronized `main`;
- task record является первым content change;
- stage, commit и publication не разрешены bare-командой.

## Size Guard

Предварительно не превышается: один новый архитектурный contract, ноль
production packages и ожидаемо не более 10–12 documentation files. При
выходе за 15 files либо появлении второго независимо поставляемого поведения
scope должен быть переоценён до продолжения.

## Documentation Baseline

- Authoritative inventory проверен: Accepted ADR-0002/0003/0004, Frozen
  ARCH-002, Active ARCH-004/ARCH-005, Draft DP-010/DP-011/DP-012,
  `spec/decisions.md`, `spec/current-state.md`, зеркальные MASTER_PLAN и
  фактические packages Owner/Flow/Loader/Source.
- Factual Control Service composition содержит только Workspace,
  Configuration и ConfigurationVersion management; Runtime management routing
  и authorization отсутствуют.
- DP-010–DP-012 реализованы изолированно и не являются production management
  wiring.
- Обнаружен non-critical drift `spec/decisions.md`: DP-012 отсутствовал в
  перечне implementation contracts, а concrete Source composition всё ещё
  называлась ожидающим решением после её изолированной реализации. Drift
  разрешён в pre-implementation synchronization.
- Critical drift, конфликт статусов, EN/RU orphan или blocker до design не
  обнаружены.

## Architecture Confirmation

- Design verdict: **READY / valid**.
- Design blockers: **0**.
- Implementation Readiness: **BLOCKED**.
- Минимальный contract: один planned
  `internal/runtimemanagement.Directory` над immutable process-local bindings
  `{Target, Owner, Loader}`, который сам конструирует ровно один Flow из того
  же Owner и Loader и маршрутизирует Start/Stop/Observe по exact Runtime
  Instance identity.
- Frozen public surface включает `Action`, immutable `Target`, named function
  `Authorize`, opaque `Binding`, immutable `Directory`, пять exact sentinels и
  методы Start/Stop/Observe. Named function исключает typed-nil interface
  ambiguity без reflection.
- Ordering frozen как validation -> visible context cancellation -> exact
  lookup/Workspace-Configuration assertion -> authorization -> delegation.
  Missing identity и mismatch возвращают один
  `ErrRuntimeInstanceNotFound`; authorization и downstream errors проходят
  unchanged.
- Owner остаётся единственной per-Instance serialization boundary. Directory
  не добавляет mutex: Owner арбитрирует concurrent Start, принимает не более
  одного claim/operation и сохраняет существующие conflict/outcome semantics
  проигравших calls.
- Active ARCH-004 §19 имеет приоритет над Draft DP-013. Любая package
  implementation, включая isolated package/local proofs, заблокирована до
  focused designs: persistence Runtime Instance/Launch Attempt и
  desired/actual facts (§19(2)), durable management idempotency (§19(3)),
  activation/replacement/rollback ordering (§19(4)),
  recovery/reconciliation (§19(5)) и operational reporting/redaction
  (§19(6)).
- Borrowed immutable `Authorize` и его captured dependencies обязаны быть
  concurrent-safe. Directory не сериализует authorization, не добавляет
  mutex/queue/goroutine; conforming Authorize сохраняет cross-scope progress,
  exact error isolation, no-recover panic boundary и original-context
  cancellation semantics.
- Запрещены Flow introspection и existing API changes, caller-supplied Flow,
  dynamic registry, service locator, package-global state, second Owner,
  retry, detached work и production bypass.

## Pre-Implementation Documentation Result

- Созданы зеркальные
  `docs/en/design/DP-013-runtime-management-routing.md` и
  `docs/ru/design/DP-013-runtime-management-routing.md`.
- Design Status: `Draft`; Implementation Status: `Planned`.
- Exact planned Go surface, sentinel strings, validation/error precedence,
  authorization-before-mutation, static Owner-to-Flow binding, branching
  dependency graph, concurrency/cancellation, ownership, Composition Audit,
  persistence boundary, proofs и deferrals записаны зеркально.
- Design indexes EN/RU содержат DP-013 как Draft/Planned.
- Production code, Go tests, HTTP API/DTO, concrete authorization policy,
  persistence, recovery, application wiring и Production Activation не
  изменены и не объявлены реализованными.
- Design contract READY, но Implementation Readiness BLOCKED обязательными
  prerequisites ARCH-004 §19(2)–(6). DP-013 не разрешает isolated
  implementation по precedent более низкостатусных Draft.
- Documentation writing stage, independent verification, review, Scope Audit
  и Coordinator Acceptance complete.

## Existing Coverage Report

- Existing Coverage: DP-010 local proofs покрывают Owner serialization;
  DP-011 — Flow orchestration и production activation gates; DP-012 —
  Source composition; текущие tests доказывают isolated components.
- Coverage Gap: design gap single management command boundary и exact
  per-Instance Owner/Flow routing закрыт зеркальным Draft DP-013. Executable
  proof, package implementation, production Control Service routing и
  activation отсутствуют и не готовы до обязательных focused designs
  ARCH-004 §19(2)–(6).
- Added Proof Tests: Not applicable для Design-only task.
- Added Regression Tests: Not applicable; production behavior не меняется.
- Remaining Limitations: executable production routing, persistence,
  authorization policy, recovery и activation остаются future evidence.

## Verification Matrix

| Risk class | Applicability | Required evidence |
| --- | --- | --- |
| Concurrency, lifecycle, cancellation, shared state | Применяется к design semantics Owner arbitration, immutable Directory, borrowed concurrent-safe Authorize и existing Flow/Owner gates; production state не добавляется | Explicit conflict check с ARCH-004/ARCH-005 и DP-010/DP-011, Authorize error/panic/cancellation/cross-scope contract, semantic EN/RU parity и independent architecture review. Новый executable race target отсутствует, потому что code/tests не меняются |
| API, configuration, production wiring | Применяется только как planned Go surface и future construction boundary | Exact Draft API, validation/error precedence, authorization-before-mutation и deferrals проверяются documentation review. Implementation Readiness BLOCKED ARCH-004 §19(2)–(6); executable capability, HTTP/API/config и production wiring отсутствуют |
| Imports, dependencies, module boundaries | Применяется только к planned dependency graph | Проверка branching dependency direction и отсутствия design cycle. Production imports, `go.mod` и `go.sum` не меняются; `go mod tidy` не применим |
| Public API | Применяется только к planned exported identifiers | Проверка минимальности и godoc-ready semantics frozen surface. Exported Go declarations не созданы; executable public behavior отсутствует |
| Documentation | Применяется полностью | EN/RU structure/status/semantic parity, relative links, Draft/Planned truthfulness, contradiction check с authoritative sources, conflict markers, trailing whitespace, exact 11-file diff и `git diff --check` |
| Repository regression smoke | Применяется как защита от случайного repository drift, хотя production code не меняется | Полные `go test ./... -count=1` и `go vet ./...`; результат фиксирует Tester, недоступность обязательной проверки не считается PASS |

## Documentation Synchronization

Mandatory Applicability Record PROCESS-002:

- `docs/tasks/TASK-015-RUNTIME-MANAGEMENT-ROUTING-DESIGN.md`: synchronized,
  task record обязателен всегда; хранит completed contract, architecture
  handoff, verification и closure.
- `spec/current-state.md`: synchronized, потому что task lifecycle и factual
  Draft/Planned design state изменились; implementation и activation явно
  отсутствуют, Implementation Readiness BLOCKED.
- зеркальные `docs/en/roadmap/MASTER_PLAN.md` и
  `docs/ru/roadmap/MASTER_PLAN.md`: synchronized, потому что durable
  Production Activation prerequisite получил отдельный planned contract;
  milestone completion не заявлен.
- зеркальные `docs/en/design/DP-013-runtime-management-routing.md` и
  `docs/ru/design/DP-013-runtime-management-routing.md`: synchronized как
  основной новый design contract; Design Status Draft, Implementation Status
  Planned.
- `.ai/PROJECT_CONTEXT.md`: synchronized, потому что current architecture task
  и fundamental planned boundary/implementation blocker изменились.
- `spec/decisions.md`: synchronized, чтобы устранить stale DP-012 gap и
  перечислить planned DP-013 без нормативного повышения Draft.
- зеркальные design indexes: synchronized, DP-013 добавлен как Draft/Planned.
- `docs/tasks/README.md`: synchronized, lifecycle завершённой TASK-015 отражён
  в task index.
- root `README.md`: Not applicable — user-facing capability, usage,
  installation и release surface не изменились.
- `CHANGELOG.md`: Not applicable — Design-only task не создаёт user-facing или
  release change.
- Accepted ADR и Active/Frozen ARCH: Not applicable — frozen Architecture
  Confirmation не изменяет их status, ownership или normative invariants.

## Stop Conditions

- конфликт с Approved ADR либо Active/Frozen ARCH;
- необходимость определить persistence transaction или product HTTP API для
  завершения management routing contract;
- невозможность отделить authorization ordering от конкретной policy;
- более одного нового архитектурного контракта;
- критический documentation drift;
- materially different Ready designs без authoritative ordering.

## Next Candidate

Не активирован. Детерминированная следующая рекомендация после принятия
TASK-015 — отдельный Design-only Runtime Operational Identity Persistence
contract для Runtime Instance, Launch Attempt, desired/actual facts, durable
opaque ID allocation и atomic history/ownership invariants. Это первый
unresolved prerequisite ARCH-004 §19(2); implementation, persistence code и
Production Activation не начинаются автоматически.

## Handoff

- Task Intake и Existing Coverage Report: complete.
- Documentation Baseline: complete; critical drift 0.
- Architecture Confirmation: READY, blockers 0.
- Pre-Implementation Documentation: complete.
- Initial Tester: FAIL, findings B-001/B-002 — отсутствовали explicit
  Verification Matrix и Mandatory Applicability Record, а Coverage Gap не
  отражал закрытый DP-013 design gap.
- Bounded Tester rework: complete только в task record; Verification Matrix,
  Documentation Synchronization и corrected Coverage Gap добавлены.
- Repeat Tester до architecture rework: PASS.
- Initial Final Reviewer: Needs Revision, R-001/R-002 — DP-013 ошибочно
  разрешал isolated implementation вопреки Active ARCH-004 §19 и не фиксировал
  concurrency contract borrowed Authorize.
- Bounded Architect rework: complete; Design READY/valid отделён от
  Implementation Readiness BLOCKED, следующий persistence design определён,
  Authorize concurrency/error/panic/cancellation contract frozen.
- Bounded Documentation rework: complete в exact authorized 8-file scope.
- Repeat Tester verification после architecture rework: PASS, 0 blocking и
  0 nonblocking findings.
- Exact verified task diff: 11 files —
  `.ai/PROJECT_CONTEXT.md`,
  `docs/en/design/DP-013-runtime-management-routing.md`,
  `docs/en/design/README.md`,
  `docs/en/roadmap/MASTER_PLAN.md`,
  `docs/ru/design/DP-013-runtime-management-routing.md`,
  `docs/ru/design/README.md`,
  `docs/ru/roadmap/MASTER_PLAN.md`,
  `docs/tasks/README.md`,
  `docs/tasks/TASK-015-RUNTIME-MANAGEMENT-ROUTING-DESIGN.md`,
  `spec/current-state.md`, `spec/decisions.md`.
- Repeat Tester evidence: EN/RU numbered headings 30/30, code fences 14/14,
  relative links 123 checked / 0 missing, `git diff --check`, conflict-marker
  и trailing-whitespace checks PASS. Полные `go test ./... -count=1` и
  `go vet ./...` PASS переиспользованы, поскольку bounded rework меняла только
  documentation и не создала нового executable target.
- Repeat Final Reviewer: Approved, 0 blocking и 0 nonblocking findings.
- Coordinator Acceptance: granted.
- Current status: Completed — Coordinator Accepted.
- Следующее действие: commit только после exact permission; implementation не
  запускается этой task.

## Scope Audit

Result: **11 Required / 0 Questionable / 0 Removable**.

Required groups and disposition:

- design contract: mirrored DP-013 EN/RU — exact planned boundary,
  concurrency, prerequisites, proofs and deferrals; both required;
- navigation: mirrored design indexes and task index — DP/task discoverability;
  all three required;
- durable project state: `.ai/PROJECT_CONTEXT.md`, `spec/current-state.md`,
  mirrored MASTER_PLAN and `spec/decisions.md` — completed task, absence of a
  current architecture task, Draft/Planned/Blocked truth, prerequisite ordering
  and drift correction; all five required;
- operational record: this TASK-015 record — completed-task contract, evidence,
  rework, applicability, audit and closure; required.

No code, Go test, generated, HTTP, persistence, activation, unrelated
refactoring or next-task implementation file exists in the diff. Removing any
Required group would break Definition of Done, navigation, PROCESS-002 or
repository truthfulness.

## Process Health

- Trigger: Not applicable.
- Reason: TASK-015 is not the tenth completed task since the applicable review
  and had no rollback, escaped defect, repeated Publisher failure or more than
  two review returns.
- Bounded process finding/change: none.

## Commit Gate

- exact command `Разрешаю коммит.` received: yes, after Coordinator
  Acceptance;
- commit message policy: Conventional Commits,
  `docs(runtime): define management routing`;
- one task commit is authorized; push/PR/merge/publication remain
  unauthorized by this command;
- exact accepted task diff: 11 documentation files;
- post-review administrative closure diff: bounded to this task record,
  `.ai/PROJECT_CONTEXT.md` and `spec/current-state.md`;
- initial focused closure audit of that bounded three-file diff: Needs Revision,
  C-001 — stale lifecycle wording remained in this task record;
- bounded C-001 correction: complete, limited to this task record;
- repeat focused closure audit of that bounded three-file diff: Approved,
  0 blocking / 0 non-blocking findings;
- generated/temporary/unrelated files: none expected;
- final `git diff --check`: PASS before this administrative update and repeated
  after it.

## Closure

- Final status: **Completed — Coordinator Accepted**.
- Design Status: **Draft**.
- Implementation Status: **Planned**.
- Implementation Readiness: **Blocked** by mandatory focused designs
  ARCH-004 §19(2)–(6).
- Completed scope: mirrored DP-013, navigation, project-state synchronization,
  verification/review rework and exact Scope Audit.
- Verification: Repeat Tester PASS 0/0; EN/RU headings 30/30, fences 14/14,
  links 123 checked / 0 missing, contradiction/conflict/trailing/diff checks
  PASS; full `go test ./... -count=1` and `go vet ./...` PASS reused because
  all subsequent changes were documentation-only.
- Review: Initial Final Reviewer Needs Revision R-001/R-002; bounded rework;
  Repeat Final Reviewer Approved 0/0.
- PROCESS-002: **Synchronized**.
- Known limitations: management package and executable proofs do not exist;
  persistence, durable idempotency, activation/replacement/rollback,
  recovery/reconciliation, reporting/redaction, concrete authorization,
  HTTP/API and Production Activation remain absent.
- Code/tests changed: no.
- Commit/publication: not performed or authorized.
- Next recommended candidate: separate Design-only Runtime Operational
  Identity Persistence contract, explicitly not activated.
- Administrative note: this accepted closure introduces only the bounded
  three-file closure diff named in Commit Gate; initial focused closure audit
  returned Needs Revision C-001, bounded task-record correction is complete,
  and repeat focused closure audit is Approved with 0 blocking / 0 non-blocking
  findings; commit requires exact permission.

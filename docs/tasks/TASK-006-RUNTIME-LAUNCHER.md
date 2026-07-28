# TASK-006 — Изолированная реализация Runtime Launcher

## Status

**Completed — Coordinator Accepted**

## Objective

Реализовать в `internal/runtime` минимальную stateless-операцию
`func Launch(request *BootstrapRequest) BootstrapOutcome` строго по
подтверждённому TASK-005 контракту: один синхронный вызов `Bootstrap` с тем же
borrowed pointer и неизменённый возврат результата без дополнительной policy,
state или lifecycle authority.

## Selection Evidence

- Autonomous entry: точная bare-команда `Продолжай проект.`.
- Active tasks отсутствуют; TASK-005 имеет статус
  `Completed — Coordinator Accepted`.
- Baseline: clean `main` на merge commit `249634a`, который содержит merged
  PR #6 и принятый diff TASK-005.
- TASK-005, `.ai/PROJECT_CONTEXT.md` и `spec/current-state.md` однозначно
  рекомендуют isolated Runtime Launcher implementation как следующий Ready
  slice.
- DP-009 зеркально фиксирует concrete signature, exact delegation, ownership,
  failure passthrough и local/integration proof boundary.
- Slice является prerequisite текущей Beta milestone, меньше Lifecycle Owner
  и production pipeline и не требует изменения Approved ADR или
  Active/Frozen ARCH.
- Отклонённые alternatives:
  - Runtime Lifecycle Owner — требует отдельных persistence, serialization,
    management и operational-state contracts ARCH-004;
  - production Loader-to-Builder-to-Launcher wiring — зависит от Lifecycle
    Owner и остаётся отдельной integration work;
  - Runtime Instance/Launch Attempt persistence, retry, replacement,
    reconciliation и management API — отдельная архитектура;
  - изменения Bootstrap или Host — существующая boundary достаточна;
  - operational diagnostics и supervision — отдельные roadmap gaps.
- Post-merge drift объяснён: project-state документы были закрыты до commit,
  но фактическая история уже содержит merge TASK-005. Drift должен быть
  устранён PROCESS-002 до final review и не меняет выбранный candidate.

## Scope

- добавить минимальную production surface `internal/runtime.Launch`;
- добавить локальные proof tests для успешного результата, failure passthrough,
  отсутствия Launcher-owned lifecycle действий и независимых concurrent calls;
- обновить DP-009 EN/RU только в части фактического Implementation Status;
- синхронизировать `.ai/PROJECT_CONTEXT.md`, `spec/current-state.md`, task index
  и этот task record;
- выполнить применимые Go и documentation checks, scope audit и независимый
  review.

## Non-Goals

- Runtime Lifecycle Owner;
- production Loader-to-Builder-to-Launcher pipeline и полная AP-003/AP-011;
- Runtime Instance/Launch Attempt model, persistence или management API;
- retry, replacement, reconciliation, logging policy или diagnostics;
- изменение Bootstrap, Host, Loader, Builder, Approved ADR, Active/Frozen ARCH
  либо публичного API;
- unrelated refactoring и запуск следующей task.

## Sources of Truth

- `docs/ru/adr/0003-runtime-architecture.md` и EN mirror;
- `docs/ru/architecture/ARCH-002-runtime-foundation-freeze.md` и EN mirror;
- `docs/ru/architecture/ARCH-004-runtime-deployment-and-identity-model.md` и EN
  mirror;
- `docs/ru/architecture/ARCH-005-runtime-configuration-snapshot-and-loading-model.md`
  и EN mirror;
- `docs/ru/design/DP-009-runtime-bootstrap-contract.md` и EN mirror;
- `docs/tasks/TASK-004-DP-009-RUNTIME-BOOTSTRAP.md`;
- `docs/tasks/TASK-005-RUNTIME-LAUNCHER-PREREQUISITES.md`;
- `internal/runtime/bootstrap.go` и `internal/runtime/bootstrap_test.go`;
- `.ai/PROJECT_CONTEXT.md`, `spec/current-state.md`, `spec/decisions.md` и
  MASTER_PLAN EN/RU.

Draft DP-009 не переопределяет источники более высокого статуса; его concrete
Launcher contract уже получил Architect `Ready` в TASK-005.

## Roles

- **Coordinator:** intake, task/branch gate, handoffs, scope audit, acceptance
  и next recommendation.
- **Architect:** подтверждает неизменность TASK-005 contract и отсутствие
  нового архитектурного решения; production code не меняет.
- **Documentation Agent:** выполняет baseline audit и PROCESS-002, обновляет
  только фактический implementation status и project state.
- **Developer:** реализует минимальный Launcher и proof tests без расширения
  architecture.
- **Tester:** выполняет targeted/full tests, formatter, vet, race availability
  check и documentation verification.
- **Reviewer:** независимо проверяет architecture, exact delegation, tests,
  documentation и scope; автор изменений не выполняет final review.

Pre-Implementation Documentation неприменима: TASK-005 уже зафиксировала
полный implementation contract, а TASK-006 его не меняет.

## Branch

- исходный trusted baseline: clean `main`, commit `249634a`;
- task branch: `feature/task-006-runtime-launcher`;
- branch action: создана безопасно до content changes;
- запрещены stage, commit, push, merge, rebase, branch deletion, fetch, pull и
  remote mutation без отдельного разрешения.

## Constraints

- production implementation эквивалентна `return Bootstrap(request)`;
- тот же pointer передаётся ровно в один Bootstrap call;
- outcome, Host, failure identities и cause chain не изменяются;
- Launcher не валидирует, не оборачивает ошибки, не вызывает `Host.Stop()`, не
  сохраняет state и не создаёт goroutine, cache, registry или adapter;
- успешный Host передаётся будущему Lifecycle Owner; Launcher им не владеет;
- Design Status DP-009 остаётся `Draft`; полные AP-003/AP-011 остаются
  integration-gated;
- commit policy: commit только по отдельному разрешению пользователя.

## Stop Conditions

- реализация требует adapter/seam, mutable package state или изменения
  Bootstrap/Host;
- возникает необходимость определить Lifecycle Owner или production wiring;
- authoritative sources конфликтуют либо нужен новый архитектурный выбор;
- baseline получает неатрибутированные изменения;
- обязательная проверка падает или independent Reviewer возвращает blocking
  finding;
- scope требует неразрешённого файла или materially расширяется.

## Acceptance Criteria

1. Этот task record является первым content change task-ветки.
2. Architect подтверждает TASK-005 contract как `Ready` без нового design.
3. `internal/runtime.Launch` имеет точную signature DP-009 и реализацию,
   эквивалентную `return Bootstrap(request)`.
4. Launcher не добавляет state, validation, wrapping, cleanup, retry, policy,
   lifecycle action или background work.
5. Proof tests покрывают success, Bootstrap Failure, Startup Failure,
   cause/Host passthrough и независимые concurrent calls в доступной локальной
   границе.
6. Existing Bootstrap/Host regression и полный Go suite проходят.
7. DP-009 EN/RU честно отражает Launcher как implemented in isolation, сохраняя
   Design Status `Draft` и integration gate AP-003/AP-011.
8. PROCESS-002 устраняет stale post-merge TASK-005 gate и синхронизирует
   project state без user-visible capability claim.
9. Scope audit не содержит неразрешённых `Questionable` или `Removable`.
10. Independent Reviewer не имеет blocking findings.

## Verification

- `gofmt -d` для затронутых Go files;
- targeted `go test ./internal/runtime -count=1`;
- full `go test ./... -count=1`;
- `go vet ./...`;
- `go test -race` при доступном CGO toolchain либо точная причина
  недоступности;
- EN/RU headings, fences, links, statuses и semantic parity;
- conflict markers, trailing whitespace и `git diff --check`;
- independent review полного diff.

## Scope Audit

**Coordinator disposition:** audit принят перед final review.

- **Required:** 8.
- **Questionable:** 0.
- **Removable:** 0.
- **Production files:** 1.
- **Test files:** 1.
- **Documentation files:** 6.
- **Generated files:** 0.

| Файл | Классификация | Связь со scope |
| --- | --- | --- |
| `internal/runtime/launcher.go` | `Required` | AC-003/004: exact stateless production delegation |
| `internal/runtime/launcher_test.go` | `Required` | AC-005/006: local outcome и concurrency proofs |
| `docs/en/design/DP-009-runtime-bootstrap-contract.md` | `Required` | AC-007: фактический EN implementation status |
| `docs/ru/design/DP-009-runtime-bootstrap-contract.md` | `Required` | AC-007: зеркальный RU implementation status |
| `.ai/PROJECT_CONTEXT.md` | `Required` | AC-008: operational state и post-merge drift correction |
| `spec/current-state.md` | `Required` | AC-008: factual implemented state |
| `docs/tasks/README.md` | `Required` | AC-001: navigation после initial task gate |
| `docs/tasks/TASK-006-RUNTIME-LAUNCHER.md` | `Required` | task contract, handoffs, audit и closure evidence |

- Premature Lifecycle Owner/pipeline work отсутствует.
- Bootstrap, Host, Loader, Builder, ADR, ARCH и другие DP не изменены.
- Generated, formatting-only и unrelated changes отсутствуют.
- Planned-state honesty сохранена: только Launcher local boundary реализована;
  полные AP-003/AP-011 остаются integration-gated.

## Handoff

### Coordinator Intake

- **Task:** TASK-006, isolated stateless Runtime Launcher.
- **Accepted contract:** TASK-005/DP-009 exact synchronous delegation.
- **Forbidden work:** Lifecycle Owner, pipeline, persistence, management,
  Bootstrap/Host changes и новая policy.
- **Baseline:** clean `main` `249634a`; branch
  `feature/task-006-runtime-launcher`.
- **Known drift:** stale pre-merge TASK-005 commit gate в project-state files;
  объяснён merge history и назначен PROCESS-002.
- **Required next action:** Documentation baseline и explicit Architect
  confirmation до Developer.

### Documentation Baseline

- **Status:** `Synchronized with one explained factual drift pending
  correction`.
- **Inventory:** DP-009 EN/RU, ARCH-002/004/005 EN/RU, task records
  TASK-004/005, task index, `.ai/PROJECT_CONTEXT.md`,
  `spec/current-state.md`, `spec/decisions.md`, MASTER_PLAN EN/RU, root README
  EN/RU и `CHANGELOG.md`.
- **Finding:** `.ai/PROJECT_CONTEXT.md` и `spec/current-state.md` сохраняют
  pre-commit gate TASK-005, хотя clean `main` `249634a` уже содержит merge PR
  #6. История однозначно объясняет drift; correction назначена PROCESS-002
  TASK-006.
- **No conflict:** Design Status DP-009 `Draft`, Launcher `Planned` и
  integration gates AP-003/AP-011 соответствуют коду до implementation.
- **No-change baseline:** decisions, MASTER_PLAN, README и CHANGELOG не требуют
  pre-implementation изменения; user-visible capability не появилась.
- **Gate:** критического архитектурного drift нет; Developer может начать
  после Architect confirmation.

### Architect Confirmation

- **Verdict:** `Ready`; blocker и новое архитектурное решение отсутствуют.
- **Confirmed operation:** `func Launch(request *BootstrapRequest)
  BootstrapOutcome` в `internal/runtime`.
- **Implementation constraint:** тело эквивалентно
  `return Bootstrap(request)`; тот же pointer, ровно один synchronous call,
  неизменённый outcome.
- **Ownership:** Launcher заимствует request только на время вызова и не
  сохраняет request, Snapshot, dependencies, outcome или Host. Успешная ссылка
  Host передаётся будущему Lifecycle Owner.
- **Forbidden:** adapter/test seam, validation, wrapping, cleanup, Stop, retry,
  policy, registry, cache, goroutine и mutable package state.
- **Proof boundary:** исходный код доказывает exact delegation; локальные tests
  проверяют observable success/failure/cause и concurrent independence. Полные
  AP-003/AP-011 остаются integration-gated до Lifecycle Owner и production
  wiring.
- **Documentation disposition:** после verification зеркально изменить только
  фактический Launcher Implementation Status и project state; Design Status
  `Draft` не повышать.

### Developer Handoff

- **Implemented:** `internal/runtime.Launch` с точной DP-009 signature и
  единственным `return Bootstrap(request)`.
- **Tests added:** Success, ordered Bootstrap Failure, Startup Failure cause и
  четыре независимых concurrent launch calls.
- **Unchanged boundaries:** Bootstrap, Host, Loader, Builder, Lifecycle Owner и
  production wiring.
- **State/authority:** fields, adapter, mutable package state, goroutine,
  validation, wrapping, cleanup, Stop, retry и policy не добавлены.
- **Initial verification:** targeted Runtime suite — PASS.
- **Open findings:** отсутствуют.

### Tester Handoff

- **Verdict:** `PASS`; blocking findings — 0.
- **Targeted:** `go test ./internal/runtime -count=1` — PASS.
- **Full regression:** `go test ./... -count=1` — PASS.
- **Static checks:** `go vet ./...`, `gofmt -d` и `git diff --check` — PASS.
- **Contract evidence:** production body непосредственно доказывает one-call
  same-pointer delegation и unchanged value return; tests подтверждают
  observable Success, Bootstrap Failure Stage/Code, Startup Failure cause,
  отсутствие partial Host и concurrent independence.
- **Race detector:** недоступен — `go env CGO_ENABLED` возвращает `0`, команда
  `gcc` отсутствует.
- **Documentation:** DP-009 EN/RU имеют 54/54 headings и 10/10 fences,
  одинаковые statuses и integration gates; stale Launcher absence claims
  отсутствуют.
- **Open findings:** отсутствуют.

### PROCESS-002 Handoff

- **Status:** `Synchronized`; final independent review pending.
- **DP-009:** Design Status сохранён `Draft`; Bootstrap и Runtime Launcher
  отмечены `Implemented in isolation`; Lifecycle Owner и production pipeline
  остаются planned/unimplemented; AP-003/AP-011 integration-gated.
- **Project state:** stale TASK-005 dirty/commit-only gate заменён factual merge
  baseline `249634a`, active TASK-006, verification evidence и текущим final
  review gate.
- **Explicit no-change decisions:**
  - `spec/decisions.md` — новый Approved decision или status transition
    отсутствует;
  - MASTER_PLAN EN/RU — production Loader-to-Builder-to-Launcher flow всё ещё
    отсутствует;
  - root README EN/RU и `CHANGELOG.md` — isolated internal boundary не создаёт
    user-visible/release capability;
  - design indexes — `Draft; implemented in isolation` остаётся точным;
  - ADR, ARCH, DP-007 и DP-008 — contracts не менялись.
- **Navigation:** task index обновлён после доказанного initial task-record
  gate.
- **Documentation checks:** EN/RU structure/status parity, links, fences,
  conflict markers, trailing whitespace и diff check — PASS.
- **Next gate:** final independent review, затем Coordinator acceptance.

### Final Reviewer Handoff

- **Verdict:** `Approved`.
- **Blocking findings:** 0.
- **Nonblocking findings:** 0.
- **Architecture:** exact same-pointer one-call delegation, unchanged outcome,
  absence adapter/state/policy и ownership boundaries соответствуют
  TASK-005/DP-009 и ARCH-002/004/005.
- **Tests:** Success, Bootstrap Failure, Startup Failure cause/no partial Host
  и concurrent independence достаточны в local proof boundary; production body
  непосредственно доказывает exact delegation без запрещённого seam.
- **Documentation:** DP-009 EN/RU status и semantic parity подтверждены;
  Lifecycle Owner/pipeline не представлены реализованными; AP-003/AP-011
  integration-gated.
- **Scope:** 8 Required, 0 Questionable, 0 Removable; premature, unrelated и
  generated changes отсутствуют.
- **Independent checks:** targeted/full Go tests, vet, gofmt, diff check, links,
  54/54 headings и 10/10 fences — PASS.
- **Limitation:** race detector недоступен при `CGO_ENABLED=0` и отсутствии
  `gcc`.
- **Coordinator disposition:** handoff принят; Coordinator Acceptance получена.

## Next Candidate

- **Рекомендуемая Ready work:** focused architecture/documentation refinement
  минимальных implementation prerequisites in-process Runtime Lifecycle Owner
  по ARCH-004 до его реализации или production pipeline wiring.
- **Readiness evidence:**
  - ARCH-004 уже определяет ownership, serialization responsibility,
    Launch Attempt, desired/actual state и Host handoff;
  - Loader, Builder, Bootstrap и Launcher boundaries реализованы изолированно;
  - factual dependency gap MASTER_PLAN — отсутствующий production
    Loader-to-Builder-to-Launcher flow;
  - smallest safe next slice должен сначала определить concrete Owner API,
    launch/stop serialization, outcome publication и proof boundary без
    persistence, management HTTP API, retry/replacement или wiring;
  - production implementation Owner не Ready до такого refinement.
- **Явно не начата:** task/branch Lifecycle Owner не созданы; code, persistence,
  management API и pipeline wiring не изменялись.

## Closure

- **Final status:** `Completed — Coordinator Accepted`.
- **Completed scope:** exact stateless `internal/runtime.Launch`, local proof
  tests, truthful DP-009 EN/RU implementation status и synchronized project
  state.
- **Changed files:** `internal/runtime/launcher.go`,
  `internal/runtime/launcher_test.go`, DP-009 EN/RU, `.ai/PROJECT_CONTEXT.md`,
  `spec/current-state.md`, task index и этот record.
- **Verification:** targeted/full Go tests, `go vet`, `gofmt -d`, links,
  headings/fences parity, whitespace/conflict markers и `git diff --check` —
  PASS; race detector unavailable due toolchain.
- **Review:** independent Reviewer `Approved`, 0 blocking и 0 nonblocking
  findings.
- **PROCESS-002:** `Synchronized`.
- **Scope Audit:** 8 Required, 0 Questionable, 0 Removable.
- **Known limitations:** Runtime Lifecycle Owner, operational identities,
  persistence, management API и production pipeline отсутствуют; AP-003/AP-011
  остаются integration-gated.
- **Commit readiness:** accepted eight-file diff готов к отдельно разрешённому
  commit; stage, commit, push и merge не выполнялись.
- **Next recommended work:** focused Lifecycle Owner prerequisite refinement;
  не активирована.
- **Closed by:** Coordinator.
- **Date:** 2026-07-28.

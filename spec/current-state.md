# Текущее состояние

**Веха:** Beta — Complete the Single-Node Runtime
**Статус реализации:** DP-005 Router и Runtime Foundation Tasks 1–10
реализованы; TASK-M10-002 добавил полный Manager-aware production shutdown
pipeline. Configuration Loader contract DP-007, Snapshot Builder contract
DP-008, Runtime Bootstrap contract DP-009, stateless Runtime Launcher и
Runtime Lifecycle Owner DP-010 реализованы изолированно. Production pipeline
Loader-to-Builder-to-Launcher и production persistent operational entities не
реализованы. Draft DP-011 и package `internal/runtimelaunchflow`
реализуют base integration contract этого pipeline изолированно; private
Managed Start-claim continuation и execution-binding/load gate DP-016/DP-017
определены Approved/Planned DP-019, но не реализованы. Draft DP-012 и
package `internal/configurationloadsource` реализуют concrete Source adapter
изолированно. Draft DP-013 и package `internal/runtimemanagement` реализуют
management routing изолированно. Together Approved DP-014–DP-018 закрывают
focused design gates ARCH-004 §19(2)–(6): DP-014 — §19(2), DP-015 — §19(3),
DP-016 — §19(4), DP-017 — §19(5), DP-018 — §19(6). DP-014 и primitive
Start/Stop boundary DP-015 реализованы изолированно packages `internal/runtimeidentity` и
`internal/runtimecommandidempotency`; partial parent/phase sequential core и
command-boundary Continue/pending-Stop rendezvous DP-019 также реализованы там
изолированно, а полный extension остаётся Planned;
DP-016–DP-019 имеют Implementation Status Planned overall. Packages
Dedicated DP-016 orchestration, DP-017 recovery, DP-018 reporting и DP-019
continuation/binding packages, external schema/HTTP API/persistence,
orchestration/recovery/reporting implementation, concrete authorization policy,
management wiring и Control Service activation отсутствуют.
**Release:** v0.1.0-alpha
**Architecture Review:** Findings TASK-ARCH-REVIEW-010 реализованы в TASK-M10-002; DP-001, DP-002 и DP-006 сохраняют Draft до отдельного status review

**Последняя завершённая development task:** TASK-029 — Runtime Command
Continue and Pending-Stop Prerequisite; `Completed — Coordinator Accepted`.

**Последняя завершённая operational task:** TASK-012 — Engineering Process
Hardening; `Completed — Coordinator Accepted`.

**Текущая operational task:** отсутствует.

**Последняя завершённая architecture task:** TASK-027 — Runtime Activation
Orchestration Prerequisites Design; `Completed — Coordinator Accepted`.

**Текущая architecture task:** отсутствует.

TASK-027 закрыла design ambiguity DP-016 через Approved/Planned DP-019.
Independent Review — `Approved`, blocking и non-blocking findings 0;
Verification Matrix, PROCESS-002, status consistency и Scope Audit 24/0/0 —
PASS. Acceptance не реализует prerequisites и не снимает TASK-026 blocker. На
момент Coordinator closure commit и publication ещё не выполнялись;
впоследствии task commit `7ac0a6b372d9e54c73d024703e6d3ee4b06e15cd`
опубликован через PR #27 и merged как
`2c017aace7e56a4747d3cecbe8ff3f6cf53e009f`.

**Текущая development task:** отсутствует; TASK-031 — Runtime Orchestration
Authorization Surface завершена `Completed — Coordinator Accepted`. TASK-026
остаётся `Blocked by
Architecture`; Coordinator Acceptance, commit и publication TASK-026 запрещены.

**TASK-031:** `Completed — Coordinator Accepted`. Bounded isolated DP-020
deferred slice 1 реализован в `internal/runtimecommandidempotency` на branch
`feature/task-031-runtime-orchestration-authorization-surface` от clean
synchronized `main@0bfbca33c4b511399ffbad2909bcaaa4a37efb4b` (merge PR #30).
Добавлены immutable validated `OrchestrationAuthorizationRequest`, exact
`OrchestrationAction` set {ActivateExactTarget, ReplaceWithExactTarget,
RollbackToExactTarget} и policy-neutral `AuthorizeOrchestration` function type с
per-call evaluation без кэша; зафиксирована Start→`ActivateExactTarget`
adaptation. Independent Tester PASS WITH ENVIRONMENT LIMITATION (race без
CGO/gcc, substitute stress `-count=100` PASS); Independent Review Approved with
Findings 0/2 (resolved); Verification Matrix, PROCESS-002, Scope Audit 7/0/0
PASS. Commit, push, PR, merge и publication не выполнялись.

**Последняя завершённая documentation task:** TASK-022 — Root README Runtime
Status Synchronization; `Completed — Coordinator Accepted`.

**Текущая documentation task:** отсутствует.

**Trusted baseline TASK-009:** clean synchronized
`main@63b961eeb59af9205c3c3d0b68d3f4bd7b8ac25c`; локальная ветка
`feature/task-009-runtime-lifecycle-owner`; task record создан первым content
change.

**Trusted baseline TASK-008:** task commit TASK-007 `2e6d221` опубликован
через PR #8 и merged в clean `main` commit `802760a`. TASK-008 начата от этого
baseline; task record создан первым content change.

**Verification TASK-001:** targeted tests PASS 3/3; full
`go test ./... -count=1` PASS 2/2; `go vet ./...`, `gofmt -d` и
`git diff --check` PASS. Race detector недоступен в текущей среде без CGO/gcc.

**Результат TASK-003:** implementation prerequisites Draft DP-009 завершены:
concrete Bootstrap request, fixed dependency bindings и structured failure
representation зеркально уточнены. На момент closure TASK-003 Design Status
оставался Draft, а Implementation Status — Planned. Последующая TASK-004
реализовала Bootstrap изолированно; последующая TASK-006 реализовала Launcher
изолированно. Runtime Lifecycle Owner и production
Loader-to-Builder-to-Launcher pipeline по-прежнему не реализованы.

**Verification TASK-004:** targeted Bootstrap, затрагиваемые Runtime boundaries,
Host lifecycle/rollback/admission и полный `go test ./... -count=1` — PASS;
`go vet ./...`, `gofmt -d` и `git diff --check` — PASS. Blocking Tester findings
отсутствуют. Race detector недоступен: `CGO_ENABLED=0`, а при включении CGO
отсутствует `gcc`.

**Результат TASK-005:** planned in-process
`func Launch(request *BootstrapRequest) BootstrapOutcome` contract зеркально
уточнён в Draft DP-009. Launcher заимствует тот же pointer request, ровно один
раз делегирует в реализованный Bootstrap и возвращает unchanged
outcome/Host/failure identities и cause chain. Success передаёт Host reference
будущему Runtime Lifecycle Owner; Launcher не добавляет policy, cleanup или
state. AP-003/AP-011 разделены на local и future integration proof. Production
code отсутствует. TASK-005 завершена со статусом
`Completed — Coordinator Accepted`.

**Verification TASK-005:** Tester verdict — PASS. PROCESS-002 — Synchronized.
Final Reviewer verdict — Approved, 0 blocking и 0 nonblocking findings.
Coordinator Scope Audit accepted: 6 Required, 0 Questionable, 0 Removable.
Coordinator Acceptance получена.

**Результат TASK-006:** `internal/runtime.Launch` реализован как exact stateless
`return Bootstrap(request)`. Launcher не добавляет adapter, state, validation,
wrapping, cleanup, retry или lifecycle policy и не сохраняет Host.

**Verification TASK-006:** targeted `go test ./internal/runtime -count=1`,
полный `go test ./... -count=1`, `go vet ./...`, `gofmt -d`, EN/RU
structure/status parity и diff checks — PASS. Race detector недоступен:
`CGO_ENABLED=0`, `gcc` отсутствует. Final Reviewer verdict — Approved,
0 blocking и 0 nonblocking findings. Coordinator Scope Audit accepted:
8 Required, 0 Questionable, 0 Removable.

**Результат TASK-007:** зеркальный Draft DP-010 с
Implementation Status `Planned` определяет минимальный
`internal/runtimelifecycle` contract. Owner immutable bound к Workspace,
Configuration и Runtime Instance, создаёт Launch Attempt и pin exact
ConfigurationVersion в `PrepareStart` до Loader/Builder, принимает closed
`PreparationResult` с first-valid-result-wins, сохраняет origin-sensitive
truthful Start/Stop outcomes и отделяет local proofs от production integration.
Runtime Lifecycle Owner и production wiring не реализованы.

**Review TASK-007:** initial findings B-001/B-002 и repeat findings R-001/R-002
устранены зеркально. Independent Reviewer verdict — `Approved`, 0 blocking и
0 nonblocking findings. Final PROCESS-002 — `Synchronized`; Final Tester —
`PASS`; Coordinator Scope Audit accepted: 8 Required, 0 Questionable,
0 Removable. Final Reviewer finding F-001 относился только к stale
project-state instructions и исправлен. Repeat Final Reviewer verdict —
`Approved`, 0 blocking и 0 nonblocking findings; Coordinator Acceptance
получена. TASK-007 завершена.

**Operational governance TASK-008:** документировано единое разрешение команды
`Разрешаю публиковать.` на полный P0–P10 pipeline одного immutable task
target. Initial P0 отделён от phase-aware Resume Reconstruction Guard;
push/merge являются checkpoints, external blocker сохраняет authority,
post-P6 resume остаётся на `main`, No CI допускается только при
`MERGEABLE / CLEAN`, cleanup и terminal payload обязательны. R-001/R-002
устранены; Repeat Reviewer verdict — `Approved`, 0 blocking и 0 nonblocking
findings; Tester verdict — `PASS`; PROCESS-002 — `Synchronized`. Commit
остаётся отдельно разрешаемым действием. Final Reviewer verdict — `Approved`,
0 blocking и 0 nonblocking findings; Coordinator Scope Audit accepted:
14 Required, 0 Questionable, 0 Removable; Coordinator Acceptance получена.
TASK-008 завершена.

**Product impact TASK-008:** отсутствует. Production code/tests, `.github`,
ADR/ARCH/DP, product capability и Runtime implementation не изменены.

**Результат TASK-009:** добавлен изолированный package
`internal/runtimelifecycle`. Один Owner permanently bound к Workspace,
Configuration и Runtime Instance, создаёт и pin Launch Attempt/version через
`PrepareStart`, accepts один closed `PreparationResult`, вызывает только
stateless `runtime.Launch`, владеет active Host reference и сериализует
truthful Start/Stop/Observe semantics без production wiring, persistence,
retry или supervision.

**Verification TASK-009:** targeted `go test
./internal/runtimelifecycle -count=1`, полный `go test ./... -count=1`,
stress `-count=100`, `go vet ./...`, `go fmt ./...`, `git diff --check`,
EN/RU parity и link validation — PASS. Race detector недоступен при
`CGO_ENABLED=0` и отсутствующем `gcc`. Independent Tester verdict — `PASS`.
Coordinator Scope Audit accepted: 14 Required, 0 Questionable, 0 Removable.
Final Reviewer verdict — `Approved`, 0 blocking и 0 nonblocking findings;
Coordinator Acceptance получена.

**Closure publication state TASK-008:** на момент closure stage, commit и
publication не выполнялись. Это historical fact, а не live operational gate.
Любая последующая разрешённая публикация reconstruct-ит фактическое состояние
из Git/GitHub; transient dirty-branch, push, PR, checks или cleanup state здесь
не хранится.

**Следующая work после TASK-009:** активирована как documentation-first
TASK-010. Production implementation Loader-to-Builder-to-Launcher,
persistence, management API, retry/reconciliation и supervision остаются
отдельной work и требуют собственного readiness/contract решения.

**Результат TASK-010:** создан и независимо reviewed зеркальный Draft DP-011.
Он определяет immutable `internal/runtimelaunchflow` boundary, synchronous
Start Operation, Caller Cancellation Gate и exact
`PrepareStart -> Load -> Build -> Start` contract. Implementation, Source
composition, management routing/authorization, persistence и Production
Activation не начаты.

**Verification TASK-010:** EN/RU имеют 33/33 headings, 14/14 fences и
эквивалентный нормативный смысл; links broken 0; `git diff --check` PASS;
Scope Audit accepted — 11 Required, 0 Questionable, 0 Removable; repeat
Independent Reviewer verdict `Approved`, 0 blocking и 0 nonblocking findings.

**Следующий candidate после TASK-010:** не активирован. Минимальная
implementation `internal/runtimelaunchflow` и local proof tests по DP-011
допустимы только через новый task intake; Source adapter, Control Service
routing, persistence и Production Activation остаются вне этого slice.

**TASK-011:** реализован isolated package `internal/runtimelaunchflow`:
immutable Owner/Loader binding, Caller Cancellation Gate, synchronous
`PrepareStart -> Load -> Build -> Start`, exact Loader failure preservation,
immutable Build Failure и Stop convergence. TASK-011 завершена со статусом
`Completed — Coordinator Accepted`; Source composition и Production Activation
отсутствуют.

**Verification TASK-011:** independent Tester verdict — `PASS`; Final Reviewer
verdict — `Approved`, 0 blocking и 0 nonblocking findings. Targeted stress
`-count=100`, affected packages, полный `go test ./... -count=1`, `go vet
./...`, exported-surface, formatting, diff, EN/RU parity и link checks — PASS.
PROCESS-002 — `Synchronized`; Scope Audit accepted: 13 Required,
0 Questionable, 0 Removable. Race detector недоступен при `CGO_ENABLED=0` и
отсутствующем `gcc`.

**Operational governance TASK-012:** в существующий Coordinator/Publisher
workflow встроены обязательный pre-work Task Contract, Existing Coverage
Report, risk-oriented Verification Matrix, Size Guard, усиленный Scope Audit,
mandatory Documentation Sync, exact Commit Gate, Publisher rechecks/cleanup
evidence и lightweight Process Health Review. Три permission/gate-команды
сохранили смысл; product architecture, production code, Runtime Launch Flow и
Production Activation не изменены. Independent Tester — `PASS`; Final
Reviewer — `Approved`; Scope Audit — 16 Required, 0 Questionable,
0 Removable. TASK-012 завершена со статусом
`Completed — Coordinator Accepted`.

**Следующий candidate после TASK-011:** не активирован. Рекомендуется отдельная
documentation-first readiness/design task, которая выберет один минимальный
prerequisite Production Activation между concrete Source composition,
management routing и persistence boundary до следующей implementation.

**TASK-013:** design-only task завершена и принята Coordinator. Зеркальный
Draft DP-012 с Implementation Status `Planned` определяет планируемый
`internal/configurationloadsource.MemorySource`: exact Version-first lookup,
mandatory composition confinement, static `uwp.configuration` v1 facts,
deep detachment, closed Loader error mapping и будущий construction
`Source -> Loader -> Flow`. Production code, repositories, management routing,
persistence и Production Activation не изменены. Final Tester — `PASS`, Final
Reviewer — `Approved` 0/0, Scope Audit accepted 10/0/0, PROCESS-002 —
`Synchronized`.

**Candidate после TASK-013:** последовательно активирован как TASK-014 и
завершён. Management routing, persistence и Production Activation остаются
более поздней work.

**TASK-014:** `internal/configurationloadsource.MemorySource` реализован
изолированно: exact Version-first repository lookup, short-circuit validation,
closed Loader error mapping, static `uwp.configuration` v1 facts, deep
detachment, repeated/concurrent loads, Loader integration и construction proof
`Source -> Loader -> Flow` без Start/Host. Initial Tester выявил B-001 по
nil-vs-non-nil empty slices; bounded rework завершён. Repeat Tester —
`PASS WITH LIMITATION`, 0 blocking и 0 nonblocking findings. Race detector
недоступен при `CGO_ENABLED=0` и отсутствии `gcc`; substitute stress
`-count=100` — PASS. После R-001 test-only rework Repeat Final Reviewer —
`Approved`, 0 blocking и 0 nonblocking findings. Scope Audit accepted 12/0/0,
PROCESS-002 — `Synchronized`; TASK-014 — `Completed — Coordinator Accepted`.
Management routing, persistence, application wiring и Production Activation
отсутствуют.

**Candidate после TASK-014:** активирован и завершён как Design-only TASK-015
— Runtime Management Routing Design.

**TASK-015:** `Completed — Coordinator Accepted`. Architecture Confirmation —
Design `READY / valid`, design blockers 0. Зеркальный Draft DP-013 с
Implementation Status `Planned` определяет один immutable process-local
`internal/runtimemanagement.Directory`, exact routing Target, policy-neutral
named function `Authorize`, authorization-before-mutation, static construction
одного Flow из exact Owner/Loader binding и сохранение существующих Owner/Flow
outcomes. Implementation Readiness `Blocked` обязательными focused designs
ARCH-004 §19(2)–(6): operational identity persistence, durable management
idempotency, activation/replacement/rollback, recovery/reconciliation и
reporting/redaction. Initial Tester `FAIL` B-001/B-002, bounded rework, Repeat
Tester `PASS` 0/0; Initial Final Reviewer `Needs Revision` R-001/R-002,
bounded Architect/Documentation rework, Repeat Final Reviewer `Approved` 0/0.
Scope Audit — 11 Required / 0 Questionable / 0 Removable; PROCESS-002 —
`Synchronized`. Package, Go tests, HTTP API/DTO, concrete authorization
policy, persistence, recovery, application wiring и Production Activation
отсутствуют. Commit и publication не выполнялись.

**Candidate после TASK-015:** активирован как Design-only TASK-016 — Runtime
Operational Identity Persistence Design.

**TASK-016:** `Completed — Coordinator Accepted`. Architecture Analysis — `Ready`,
blockers 0. Зеркальный non-normative Draft DP-014 с Implementation Status
`Planned` предлагает candidate contract ARCH-004 §19(2): один durable
aggregate Runtime Instance с immutable Workspace/Configuration/Instance
binding, monotonic conditional revision, последними подтверждёнными Owner
desired/actual facts, не более чем одним active Launch Attempt и append-only
membership history. Attempt является owned child с key
`(RuntimeInstanceID, LaunchAttemptID)`; его parent/ID/exact Published
ConfigurationVersion pin immutable, а lifecycle phase/outcome facts
conditionally и монотонно обновляются внутри того же child. Atomic
phase-sensitive publications разрешают retained active `AttemptStopping` при
stop failure или cleanup-unproven; association очищается только после proof
отсутствия Host resources. Stale operations выполняют zero mutation, а
indeterminate outcome требует inspection exact identity/revision без blind
retry с новым ID. Persisted actual является последним подтверждённым fact, а
не liveness proof после потери Owner. Lifecycle Owner остаётся единственным
lifecycle decision maker и owner live Host. Initial Tester — `PASS`. Initial
Reviewer — `Needs Revision`, findings R-001/R-002/R-003/N-001; bounded
Architect/Documentation rework завершён. Repeat Tester — `PASS`; Repeat
Reviewer и Final Reviewer — `Approved`, 0 blocking и 0 nonblocking findings.
Scope Audit accepted — 13 Required / 0 Questionable / 0 Removable.
PROCESS-002 — `Synchronized`. Exact scope — 13 files; DP-014 27/27 headings и
4/4 fences, DP-013 35/35, MASTER_PLAN 36/36; changed links 152/0, repository
links 753/0; diff check PASS; preceding full `go test ./...` и
`go vet ./...` PASS reused после documentation-only rework. §19(2) остаётся
formal implementation blocker до отдельного approval/status decision вместе
с §19(3)–(6). Persistence package, schema, API, recovery и production wiring
отсутствуют. Commit и publication не авторизованы и не выполнялись.

**Candidate после TASK-016:** активирован как Design-only TASK-017 — Runtime
Management Command Idempotency Design.

**TASK-017:** `Completed — Coordinator Accepted`. Зеркальный non-normative
Draft DP-015 с
Implementation Status `Planned` предлагает candidate contract ARCH-004
§19(3): opaque command key внутри exact authorized command scope, immutable
intent, durable claim до lifecycle delegation, same-intent non-mutating replay,
durable per-Instance barrier при unresolved outcome, mandatory tracked-Start
Stop и truthful indeterminate outcome. Initial Reviewer — `Needs Revision`
B-001/B-002; first bounded rework; Repeat Reviewer — `Needs Revision` B-003;
second bounded rework; Final Architecture Reviewer — `Approved with Findings`,
blocking 0; closure bookkeeping; Terminal Reviewer — `Approved`, 0 blocking и
0 nonblocking findings. Full `go test ./... -count=1`, `go vet ./...`, EN/RU
29/29 headings, repository links 770/0 и diff checks — PASS. Scope Audit —
15 Required / 0 Questionable / 0 Removable; PROCESS-002 — `Synchronized`.
Formal gates §19(2) и §19(3),
downstream designs §19(4)–(6), command store, schema, API, recovery, management
wiring и Production Activation остаются открытыми.

**Candidate после TASK-017:** активирован как Design-only TASK-018 — Runtime
Activation, Replacement, and Rollback Design.

**TASK-018:** `Completed — Coordinator Accepted`. Зеркальный
non-normative Draft DP-016 с Implementation Status `Planned` предлагает
candidate contract ARCH-004 §19(4): exact Published ConfigurationVersion,
fresh Launch Attempt для activation/replacement/rollback,
Stop-to-proven-release до нового Start, zero Host overlap, explicit rollback
без automatic fallback и phase-specific concurrency/cancellation. Обязательный
Stop-during-Starting использует planned private Start-claim continuation
DP-011/DP-013 после Owner claim и до Load; current Flow seam не реализует.
Review rework B-001–B-005 завершён; terminal Reviewer — `Approved`, 0 blocking
и 0 nonblocking; full Go regression, vet, parity, links и diff checks — PASS;
Scope Audit 19/0/0; PROCESS-002 `Synchronized`; Process Health Review complete.
Formal gates §19(2)–(4), downstream §19(5)–(6), implementation и Production
Activation остаются открытыми.

**Следующая рекомендация после TASK-018:** предварительно focused recovery и
reconciliation после termination Control Service ARCH-004 §19(5); активирована
как Design-only TASK-019.

**Publication TASK-018:** task commit `64e1fe7` опубликован через PR #19 и
merged в clean `main` commit `d083957`; task branch удалена после exact OID
verification.

**TASK-019:** `Completed — Coordinator Accepted`. Зеркальный non-normative Draft DP-017 с
Implementation Status `Planned` предлагает candidate contract ARCH-004 §19(5):
exact fail-closed restart assessment, один durable recovery claim,
DP-014-owned attempt/generation binding до Load, attempt/generation-bound
execution evidence, phase-sensitive reconciliation primitive/linked commands
без lifecycle replay и reopening admission только после coherent fully terminal
verification. Resource absence даёт Failed/interrupted; Stopped требует exact
Host shutdown-completion proof. Review rework B-001–B-005 и residual wording
завершён; Final Confirmation Reviewer — `Approved`, 0 blocking и 0
nonblocking; full Go regression, vet, EN/RU parity, 247/0 links и diff checks —
PASS; Scope Audit accepted 21/0/0; PROCESS-002 `Synchronized`. Formal gates
§19(2)–(5) и последующий §19(6) остаются открытыми. Recovery store/schema,
execution-evidence adapter, executor, API и production wiring отсутствуют.

**Следующая рекомендация после TASK-019:** operational error reporting и
redaction ARCH-004 §19(6); активирована и завершена как Design-only TASK-020.

**TASK-020:** `Completed — Coordinator Accepted`. Зеркальный non-normative
Draft DP-018 с Implementation Status `Planned` предлагает candidate contract
ARCH-004 §19(6): exact owner сохраняет raw error/cause, а report projected
только после authoritative fact; valid negative outcomes не становятся errors;
exact owner/phase precedence выбирает одну stable category; correlation opaque
и authorized; redaction строится fail-closed allowlist; replay детерминирован в
projection version; delivery failure не меняет domain truth и не повторяет
lifecycle work. Review B-001–B-005 и matrix clarity rework завершён; Terminal
Design Reviewer — `Approved`, 0 blocking и 0 nonblocking; final Go regression,
vet, EN/RU parity, repository links и diff checks — PASS; Scope Audit accepted
21/0/0; Repeat terminal Closure Review — `Approved`, 0/0; PROCESS-002
`Synchronized`. Formal gates §19(2)–(6) остаются до
отдельных status decisions. Report model/projector/adapter, public API,
management implementation и Production Activation отсутствуют.

**TASK-021:** `Completed — Coordinator Accepted`. DP-014–DP-018 имеют Design
Status `Approved` и Implementation Status `Planned`; focused design gates
ARCH-004 §19(2)–(6) закрыты. Draft/Planned DP-013 имеет Implementation
Readiness `Ready for a bounded isolated implementation slice`; full integration
и Production Activation остаются blocked. Tester PASS; independent Reviewer
Approved 0/0; PROCESS-002 `Synchronized`; Scope Audit 21/0/0. Commit и
publication не выполнялись.

**TASK-022:** `Completed — Coordinator Accepted`. Root `README.md` и
`README.ru.md` теперь правдиво отражают связь Configuration Loader со Snapshot
Builder через isolated in-process Runtime Launch Flow и доказанную isolated
конструкцию `Source -> Loader -> Flow` поверх concrete in-memory Source
adapter. Application/Control Service wiring и Production Activation остаются
отсутствующими. Tester PASS, 0 findings; independent Reviewer Approved 0/0;
PROCESS-002 `Synchronized`; Scope Audit 6/0/0. На момент closure commit и
publication не выполнялись. Bounded isolated DP-013 implementation
Ready/recommended, но не активирована.

**TASK-023:** `Completed — Coordinator Accepted`. Bounded isolated package
`internal/runtimemanagement` реализует exact Target/Binding/Directory surface
Draft DP-013, authorization-before-mutation и immutable routing Start к Flow,
а Stop/Observe к exact Owner. Focused и full repository tests и vet проходят;
race detector технически недоступен из-за отсутствующего C compiler, focused
stress PASS. Independent Reviewer Approved 0/0; PROCESS-002 Synchronized;
Scope Audit 23/0/0. DP-013 остаётся Draft, production Control Service routing,
concrete policy, persistence, recovery, management wiring и Production
Activation отсутствуют. Следующий candidate — отдельный bounded readiness
intake implementation prerequisites Approved/Planned DP-014; не активирован.

**TASK-024:** `Completed — Coordinator Accepted`. Bounded isolated package
`internal/runtimeidentity` реализует все девять conceptual operations
Approved DP-014 §21 — `AllocateCandidateIdentity`, `CreateRuntimeInstance`,
`ReadRuntimeInstance`, `ReadLaunchAttemptHistory`,
`ConditionalClaimLaunchAttempt`, `ConditionalBindExecutionGeneration`,
`ConditionalPublishRunning`, `ConditionalClaimStop`,
`ConditionalPublishTerminal` — и удовлетворяет всем семнадцати acceptance
proofs §22 как isolated in-process in-memory store. 35 proof/regression tests
и focused stress `-count=100` PASS; race detector недоступен без C compiler.
Full regression, vet, gofmt и diff checks PASS; Independent Reviewer Approved
0/0; PROCESS-002 Synchronized; Scope Audit 12/0/0. Design Status DP-014
остаётся Approved. External storage, HTTP API, production wiring и Production
Activation отсутствуют. Следующий candidate — bounded isolated DP-015
implementation; активирован как TASK-025.

**TASK-025:** `Completed — Coordinator Accepted`. Isolated package
`internal/runtimecommandidempotency` реализует exact Scope/CommandKey identity,
immutable Start/Stop intent, authorization-before-claim, atomic per-Instance
admission, claim-before-delegation, one-shot process-local execution permit,
tracked-Start Stop exception, unresolved barrier и terminal semantic replay.
`MemoryStorage` сохраняет claim/replay facts при reconstruction `Boundary`, но
не обещает persistence через restart процесса и не восстанавливает live
permits. Focused stress, full regression, vet, gofmt и diff checks PASS; race
detector недоступен без `gcc`. Independent review и Coordinator Acceptance ещё
не выполнены. Independent Reviewer вернул 4 blocking findings: abandoned
permit ошибочно остаётся tracked; stale Boundary может commit Claim после
client reconstruction; pre-existing DP-014 §25 EN/RU противоречит factual
Implementation Status; post-claim cancellation/lost-permit/stale-client proofs
неполны. Rework устранил returned-permit gap через synchronous private permit,
атомарно сериализовал stale-client admission, исправил DP-014 §25 EN/RU и
добавил post-claim cancellation/abandoned-permit/stale-client proofs. Repeat
Verification PASS WITH LIMITATION (`gcc`/race unavailable); repeat independent
review ожидается. Integration, API, recovery и Production Activation
отсутствуют. Repeat Reviewer подтвердил исходные B-001–B-004 resolved, но
вернул новые blocking RR-B-001–RR-B-003: stale DP-013 EN/RU, exported godoc о
returned permit и отсутствующий README applicability record. Acceptance и
closure запрещены до second rework и нового Approved independent review.
Second rework синхронизировал DP-013 EN/RU, исправил exported private-permit
godoc и добавил root README applicability record. Verification/PROCESS-002 и
16/0/0 provisional Scope Audit завершены. Third Independent Reviewer подтвердил
code/proofs, но вернул Critical IR3-B-001: earlier normative sections
DP-013/DP-014/DP-015 EN/RU и `spec/decisions.md` содержат residual stale status
contradictions; Low IR3-N-001 отмечает grammar MASTER_PLAN EN. Acceptance и
closure запрещены до documentation rework и нового Approved review. Final
documentation cleanup устранил live contradictions во всех DP-013/014/015
EN/RU sections, MASTER_PLAN и project-state sources, исправил grammar и
дополнил актуальную сводку DP-015. Links 852/0, parity, status validation,
full/stress tests, vet, gofmt, module/diff checks и Scope Audit 16/0/0 PASS;
race остаётся недоступен без `gcc`. Задача передана новому независимому
Reviewer; Acceptance остаётся запрещённым до verdict Approved.
Fourth Independent Reviewer вынес `Needs Revision`: FIR-B-001 обнаружил
`runtime.Goexit` path, сохраняющий lost permit falsely tracked; FIR-B-002
обнаружил stale design indexes и live DP-016/DP-017 status wording вне
16-file sync; FIR-B-003 подтвердил residual contradiction в
`spec/decisions.md`. Defensible rework scope — 22 Required files. Coordinator
Acceptance, Closure, commit и publication запрещены до bounded rework и нового
Approved independent review.
Fifth rework устранил FIR-B-001 defer-based cleanup для `runtime.Goexit` и
добавил regression proof. Полный status sweep классифицировал 32 документа как
22 live и 10 historical, синхронизировал design indexes, DP-013/016/017/018
EN/RU, `spec/decisions.md` и project bookkeeping. Verification Matrix,
PROCESS-002 и 19 status assertions PASS; два interrupted read-only reviews
дополнительно нашли generic drift в sections `Что существует`/`Чего не
существует` и DP-011 EN/RU. Same-slice corrections устранили его, exact Scope
Audit расширен до 26/0/0; race ограничен отсутствием `gcc`. Задача передаётся
новому independent Reviewer; Acceptance остаётся запрещённым до Approved
verdict.
Post-terminal Independent Reviewer вынес `Approved`, blocking findings 0:
FIR-B-001, DP-015 §24 isolated proofs, 32/37-document status sweep,
PROCESS-002, 25/25 status assertions, Verification Matrix и Scope Audit 26/0/0
PASS. Задача передана Coordinator для отдельного Closure Audit / Coordinator
Acceptance; Commit Gate, commit, push и publication не выполнялись.
Coordinator Closure Audit повторно подтвердил Task Contract, exact scope
26/0/0, Verification Matrix, PROCESS-002, status consistency, Approved review
0 blocking и отсутствие unexpected/staged changes. Coordinator Acceptance —
`Accepted`. Task commit `06c80265f262b654c0a4fd71db6466b4a3c5d644`
опубликован и merged через PR #26 в merge commit
`751577e839cdea3a0f35032b1339d1d9f74d28ec`; branches удалены, synchronized
`main` подтверждён.

**TASK-026:** `Blocked by Architecture`. Architecture Blocking Discovery на
своём baseline подтвердил отсутствие implementable DP-015 parent/phase API, private
DP-011/DP-013 Start-claim continuation, coordination exact Owner-issued attempt
с DP-014 publication/binding и authorization tuple для replacement/rollback.
Simplified Variant B отклонён; implementation scope DP-016 не уменьшен;
production/test changes, Coordinator Acceptance, commit и publication не
выполнялись и запрещены до prerequisite implementation.

**TASK-027:** `Completed — Coordinator Accepted`. Approved/Planned DP-019 фиксирует
единый prerequisite contract: exact orchestration authorization tuple,
callback-scoped parent/phase admission, private scoped lifecycle invoker и
synchronous continuation после Owner claim и до Load. Independent Review —
Approved, blocking и non-blocking findings 0; Verification Matrix,
PROCESS-002, Status Consistency Validation и Scope Audit 24/0/0 — PASS.
TASK-026 остаётся Blocked после design acceptance — до отдельной реализации
DP-019. На момент Coordinator closure TASK-027 commit и publication ещё не
выполнялись; впоследствии task commit
`7ac0a6b372d9e54c73d024703e6d3ee4b06e15cd` опубликован через PR #27 и merged
как `2c017aace7e56a4747d3cecbe8ff3f6cf53e009f`.

**TASK-028:** `Completed — Coordinator Accepted`. После Architect handoff исходное
утверждение полного parent/phase admission boundary с Continue/Stop race
отклонено как нецелостное без pending-Stop rendezvous. Принят independently
testable partial core: durable parent/derived-phase records, callback-scoped
capabilities, strict optional `StopOld` -> `StartTarget` order, replay,
unresolved barriers и reconstruction invalidation. На closure-time Continue
gate, independent Stop coexistence, tracked-Start exception, pending-Stop
rendezvous, private managed-Flow continuation и attempt binding оставались
отдельными prerequisites.
Repeat Independent Review подтвердил `Approved`, blocking/non-blocking 0;
Verification Matrix, PROCESS-002, Status Consistency Validation и Scope Audit
24/0/0 — PASS. TASK-026 остаётся Blocked. На момент Coordinator closure
TASK-028 commit и publication ещё не выполнялись; впоследствии task commit
`d28efa4e88e02ef528c78c3ca88b3f91945069ce` опубликован через PR #28 и merged
как `ba75e54e00c3cf1d0d87ca2a985acc9699698efd`.

**TASK-029:** `Completed — Coordinator Accepted`. Bounded isolated DP-015
command-boundary Continue gate и synchronous pending-Stop rendezvous
реализованы и прошли Architecture PASS с blocking 0, Size Guard ACCEPT
(`DO NOT SPLIT`, net production `+680`) и Independent Tester PASS WITH
ENVIRONMENT LIMITATION с blocking/non-blocking 0. Focused coverage — 85.9%;
focused/package/shuffled stress `-count=100`, full tests, vet, GoDoc и diff
checks — PASS. Race build недоступен без CGO/gcc. Independent Review —
`APPROVED`, blocking/non-blocking 0/0; PROCESS-002/status/parity PASS, links
886/0, Scope Audit 25/0/0, staged/unexpected changes 0. Coordinator Closure
Audit — PASS; Coordinator Acceptance — `Accepted`; branch baseline
`ba75e54e00c3cf1d0d87ca2a985acc9699698efd`. Commit, push и publication
TASK-029 не выполнялись. Exact authorization/private
invoker, managed Flow/OwnerClaimView continuation, DP-014 attempt/generation
binding, DP-016 orchestrator и production composition остаются Planned.
TASK-026 остаётся Blocked.

Следующая рекомендация — отдельный bounded readiness/intake для lowest
remaining DP-019 prerequisite: exact orchestration authorization, private
managed invocation и OwnerClaim-to-DP-014 binding sequence. Она активирована
как design-only TASK-030, создавшая зеркальный Draft/Planned
[DP-020](../docs/ru/design/DP-020-runtime-orchestration-binding-sequence-readiness.md).
DP-020 фиксирует упорядоченное implementable разложение и закрывает отложенные
design-решения, но не реализует ни один срез. Следующая рекомендация после
TASK-030 — отдельный intake для deferred slice 1 (orchestration authorizer
surface) — активирована и завершена как TASK-031. Следующая после TASK-031 —
аналогичная intake deferred slice 2 — активирована и завершена как TASK-032
после rework; slice 2 реализован. Следующая рекомендация — deferred slice 3
(OwnerClaim-to-DP-014 conditional binding sequence); не активирована. TASK-026
остаётся Blocked до реализации и независимой приёмки оставшихся срезов.

**Stage 2 verification completed:** для TASK-003, TASK-004, TASK-005, TASK-006
и TASK-007 соответствующий task record создан как первый content change на task
branch, а task index обновлён только после initial gate; task-before-work
ordering доказан без ослабления invariant.

**Publication history:** TASK-005 commit `99e0d3d`, TASK-006 commit
`fd0f80a` и TASK-007 commit `2e6d221` merged через PR #6, PR #7 и PR #8.
Transient pre-commit или Publisher blocker state не является durable
project-state инструкцией и реконструируется из Git/GitHub.

Builder не подключён к production launch pipeline. Design Status DP-008
остаётся Draft, Implementation Status — Implemented.

Runtime Bootstrap DP-009 реализован и проверен изолированно. Design Status
остаётся Draft, Bootstrap Implementation Status — Implemented in isolation,
Runtime Launcher Implementation Status — Implemented in isolation. Production
launch pipeline не реализован; AP-003 и AP-011 остаются integration-gated.

Runtime Lifecycle Owner DP-010 зеркально спроектирован и независимо reviewed.
Design Status остаётся Draft, Implementation Status — Implemented in
isolation. Two-phase contract `PrepareStart -> external preparation -> Start`
реализован в локальном Owner package, но не означает, что Loader/Builder
adapter, management routing или production pipeline реализованы.

Runtime Launch Flow DP-011 реализован изолированно в
`internal/runtimelaunchflow`. Design Status остаётся Draft, Implementation
Status — Implemented in isolation. Flow синхронно соединяет Owner, Loader и
Builder, но не выбирает Source, management route или Production Activation.

Runtime Source Composition DP-012 имеет Design Status Draft и Implementation
Status Implemented in isolation. In-memory Source adapter, local proofs,
Loader integration и construction proof существуют; external Source
persistence и production composition отсутствуют.

Runtime Management Routing DP-013 имеет Design Status Draft и Implementation
Status Implemented in isolation. Immutable Directory, exact Target/Binding,
policy-neutral authorization seam и local proofs существуют; concrete policy,
HTTP API, external/process-restart persistence, application wiring и Production
Activation отсутствуют.

Runtime Operational Identity Persistence DP-014 имеет Design Status Approved и
Implementation Status Implemented in isolation. Package `internal/runtimeidentity`
реализует все девять conceptual operations §21 и удовлетворяет всем acceptance
proofs §22 как in-memory store изолированно; external storage, HTTP API,
production wiring и Production Activation отсутствуют.

Runtime Management Command Idempotency DP-015 имеет Design Status Approved;
primitive Start/Stop boundary, partial parent/phase sequential core DP-019 и
command-boundary Continue/pending-Stop rendezvous имеют Implementation Status
Implemented in isolation, полный extension DP-019 — Planned. Package
`internal/runtimecommandidempotency` реализует process-local in-memory
claim/replay boundary и применимые isolated acceptance proofs §24; external
durable storage, management integration/API, recovery, reporting и Production
Activation отсутствуют.

## Архитектурные решения

- ADR-001 закрепляет базовую реализацию Control Service.
- ADR-002 закрепляет ConfigurationVersion как декларативный Configuration DSL и единственный источник истины для будущего Runtime.
- Published ConfigurationVersion является immutable; Runtime исполняет ее без скрытой или альтернативной Configuration.
- Публичная схема Configuration DSL развивается обратно совместимо; несовместимые изменения требуют нового ADR.
- ADR-003 закрепляет компонентную архитектуру будущего Runtime, dependency injection и независимость от HTTP API и Repository.
- ADR-004 закрепляет передачу в Handshake минимальных live read-only capabilities для Host-owned Admission Gate и Runtime context без зависимости от concrete Host.
- Runtime использует только immutable Configuration Snapshot, созданный из Published ConfigurationVersion.

## Состояние релиза

- Workspace CRUD завершен.
- Configuration CRUD завершен.
- ConfigurationVersion create, publish и archive завершены.

## Что существует

- Миссия и видение продукта
- Архитектурные принципы
- Структура спецификаций
- Соглашения по оформлению ADR
- Руководства для участников и агентов
- Правила исключения файлов из репозитория
- Go module для Go 1.25
- Исполняемый Control Service
- HTTP Server на Chi Router с endpoint `GET /health`
- Configuration адреса и уровня журнала через `UWP_HTTP_HOST`, `UWP_HTTP_PORT` и `UWP_LOG_LEVEL`
- Безопасные значения по умолчанию: `127.0.0.1:8080` и уровень журнала `info`
- Валидация Configuration до запуска сервиса
- Структурированное логирование с настраиваемым уровнем через `slog`
- HTTP timeout и graceful shutdown по `os.Interrupt` и `SIGTERM`
- Тесты загрузки Configuration и endpoint `GET /health`
- Доменная сущность Workspace с полями ID, Name, Description, CreatedAt и UpdatedAt
- Потокобезопасный in-memory Workspace repository с последовательными ID
- Service-слой с доменной валидацией Workspace и управлением временными метками
- HTTP CRUD API `/api/v1/workspaces` с единым форматом ошибок и строгой обработкой JSON
- Unit-тесты repository, service и HTTP handler Workspace
- Доменная сущность Configuration с обязательной принадлежностью существующему Workspace
- Потокобезопасный in-memory Configuration repository с последовательными ID
- Service-слой с Unicode-валидацией, UTC-временем и проверкой существования Workspace
- Вложенный HTTP CRUD API `/api/v1/workspaces/{workspaceID}/configurations`
- Запрет удаления Workspace, содержащего Configuration
- Unit-тесты repository, service и HTTP handler Configuration
- Доменная сущность ConfigurationVersion с последовательной нумерацией внутри Configuration
- Потокобезопасный in-memory ConfigurationVersion repository
- Создание Draft Version и получение списка через вложенный API `/api/v1/workspaces/{workspaceID}/configurations/{configurationID}/versions`
- Проверка существования Configuration перед созданием и чтением Version
- Unit-тесты repository, service и HTTP handler ConfigurationVersion
- Публикация Draft Version через endpoint `/api/v1/workspaces/{workspaceID}/configurations/{configurationID}/versions/{versionID}/publish`
- Атомарное архивирование предыдущей Published Version при публикации новой
- Инвариант единственной Published Version внутри Configuration
- Ручное архивирование Draft, Validated и Published Version через endpoint `/api/v1/workspaces/{workspaceID}/configurations/{configurationID}/versions/{versionID}/archive`
- Архивирование Published Version без автоматической публикации замены
- ListenerSettings с Host и Port для ConfigurationVersion
- Значения ListenerSettings по умолчанию `127.0.0.1:8080`
- Редактирование ListenerSettings только для Draft Version
- Валидация IP-адреса или hostname без DNS lookup и диапазона Port `1..65535`
- TLSSettings с Enabled, CertificateRef, PrivateKeyRef и MinVersion для ConfigurationVersion
- Редактирование TLSSettings только для Draft Version
- Ссылки на сертификат и закрытый ключ без хранения PEM или чтения файлов
- Поддержка минимальных версий TLS `1.2` и `1.3`
- TimeoutSettings с handshake, read, write и idle timeout для ConfigurationVersion
- Значения timeout задаются в секундах и редактируются только для Draft Version
- Значение `0` отключает только read и idle timeout; handshake и write требуют положительного значения

## Authentication Domain Model

- AuthenticationSettings как отдельная metadata-секция ConfigurationVersion
- Настройки Authentication с флагом Enabled и упорядочиваемыми по Priority Provider типа `jwt`, `api-key` и `basic`
- Полная замена AuthenticationSettings только для Draft Version через endpoint `/api/v1/workspaces/{workspaceID}/configurations/{configurationID}/versions/{versionID}/authentication`
- Валидация уникальности Name и Priority Provider при допустимом повторении Type
- API Key Provider metadata с Header и SecretRef внутри AuthenticationSettings
- Default Header `X-API-Key` и строгая валидация HTTP header field name
- Проверка формата SecretRef без разрешения ссылки и проверки существования Secret
- JWT Provider metadata с SigningKeys, AllowedAlgorithms, AllowedIssuers, AllowedAudiences, RequiredClaims и ClockSkewSeconds
- Signing Keys представлены SecretRef; поддерживаются algorithms HS, RS, ES и PS семейств с размерами 256, 384 и 512
- Default ClockSkewSeconds равен `60`; JWT metadata редактируется через общую секцию Authentication только для Draft Version
- Basic Authentication Provider metadata с Realm и SecretRef
- Default Realm `Universal WebSocket Platform`; SecretRef хранит только ссылку на будущие credentials
- AuthenticationValidator отделяет cross-provider и provider-specific business validation от ConfigurationVersion Service
- DefaultAuthenticationValidator не зависит от Repository, HTTP, Runtime или Persistence
- При включенной Authentication требуется минимум один enabled Provider; при выключенной Authentication настроенные Providers сохраняются и могут быть проигнорированы будущим Runtime
- Configuration domain не выполняет Authentication; Runtime API Key Provider описан ниже

## Secret References

- Принято направление Secret References: ConfigurationVersion хранит только ссылки на секреты, а не секретные значения
- Существующие CertificateRef и PrivateKeyRef соответствуют этому направлению
- Создан storage-neutral интерфейс Secret Resolver с общей валидацией и нормализацией Secret Reference
- Добавлена потокобезопасная in-memory реализация для тестирования и будущей локальной разработки
- Реальные Secret Storage backend еще не реализованы
- Resolver используется API Key и JWT Provider и подключен к production Authentication Pipeline через Runtime Host composition

## JWT Provider Design

- [Authentication proposal DP-003: JWT Provider](../docs/ru/proposals/DP-003-jwt-provider.md) предлагает Configuration-модель JWT Provider с несколькими Signing Keys, algorithms, issuers, audiences и Required Claims
- Signing Keys представлены только Secret References без хранения PEM, JWK или HMAC secret в ConfigurationVersion
- JWT Provider metadata и Runtime Provider реализованы; Runtime поддерживает только HS256, HS384 и HS512 и выполняет Provider через pre-Upgrade Authentication Pipeline

## Authentication Runtime Contracts Design

- [Authentication proposal DP-004: Authentication Runtime Contracts](../docs/ru/proposals/DP-004-authentication-runtime-contracts.md) предлагает transport-neutral контракты AuthenticationRequest, Principal, AuthenticationResult и AuthenticationProvider
- Предлагаемые контракты отделяют AuthenticationService и Provider от transport, Repository, Storage и внутреннего устройства ConfigurationVersion
- Модель ошибок различает rejected credentials, Provider error, Configuration error и Internal error
- Principal после успешной Authentication предлагается сделать immutable перед передачей в Authorization
- Созданы минимальные transport-neutral модели AuthenticationRequest, AuthenticationResult и Principal
- Созданы расширяемые интерфейсы Authentication Provider и Factory, принимающие AuthenticationProviderSnapshot и Secret Resolver
- Создан потокобезопасный Authentication Provider Registry с регистрацией Factory по Provider Type
- Registry делегирует создание Provider соответствующей Factory и не выполняет Authentication
- Реализован первый Runtime Authentication Provider для API Key с case-insensitive поиском Header
- API Key Provider разрешает Secret Reference при каждом Authenticate и сравнивает credentials через constant-time operation
- Реализован Authentication Service, последовательно вызывающий Provider в заданном порядке и завершающийся после первого успешного результата
- Реализован Authentication Bootstrap, собирающий Service из Authentication Snapshot через Provider Registry и Secret Resolver
- Реализован production API Key Factory, изолирующий преобразование AuthenticationProviderSnapshot в локальную runtime-конфигурацию API Key Provider
- Реализован Runtime JWT Provider с проверкой signature, exp, nbf, issuer, audience и Required Claims
- JWT Provider разрешает Signing Key через Secret Resolver при каждом Authenticate и поддерживает rotation без хранения Secret
- Реализован production JWT Factory, глубоко копирующий JWT metadata из AuthenticationProviderSnapshot в локальную runtime-конфигурацию Provider
- Basic Provider по-прежнему не реализован

## Runtime Architecture

- Принята последовательность компонентов от Configuration Loader и Configuration Snapshot до Monitoring
- Secret Resolver разрешает Secret References только при запуске Runtime; значения Secret остаются только в памяти процесса
- Authentication Provider Registry отделяет Runtime и Authentication Service от конкретных реализаций Provider
- Authentication использует transport-neutral контракты [Authentication proposal DP-004](../docs/ru/proposals/DP-004-authentication-runtime-contracts.md) и не зависит от WebSocket
- Реализована immutable Runtime Configuration Snapshot-модель для Listener, Authentication и optional Routing
- Реализован neutral immutable `runtimeconfigload` handoff: `LoadRequest` и `DetachedLoadResult` сохраняют declarative и operational identities, schema facts и detached ConfigurationVersion
- Реализован Configuration Loader, который загружает ровно одну pinned Published ConfigurationVersion через source boundary, проверяет completeness, identity chain, lifecycle state и schema facts и возвращает detached result
- Loader и neutral handoff покрыты unit-тестами, но не подключены к production launch pipeline Control Service или Runtime
- Реализован Snapshot Builder поверх neutral `DetachedLoadResult`: он проверяет exact schema `uwp.configuration` v1, handoff identity и все применимые Listener, TLS, Timeout, Authentication и Routing semantics
- Builder возвращает исключительно полный Snapshot без Diagnostics либо полные дедуплицированные blocking Diagnostics без Snapshot; registry содержит 93 детерминированно упорядоченных Code/Location/fixed Message rules
- Snapshot хранит полный provenance Workspace, Configuration, ConfigurationVersion ID и number, schema identity/version, Runtime Instance и Launch Attempt из ARCH-005
- Snapshot имеет private storage и detached readers для Listener, Authentication и optional Routing; вложенные collections и повторные Build не разделяют mutable logical content
- Snapshot не зависит от HTTP API, Repository или исходного ConfigurationVersion после создания
- Runtime Container хранит immutable Snapshot value и возвращает его by value через единственный метод `Snapshot()`; independent ownership сохраняется private storage и detached readers Snapshot без mutable logical aliases
- Container пока не содержит других зависимостей и самостоятельно не управляет запуском, остановкой или reload Runtime
- Реализован потокобезопасный Runtime Host, являющийся production composition root и владеющий независимой копией Snapshot и Container
- Host поддерживает lifecycle `Created -> Built -> Starting -> Running -> Stopping -> Stopped`; Restart и Reload отсутствуют
- Runtime Bootstrap принимает concrete request со Snapshot by value, startup context и fixed typed bindings, выполняет три ordered static validations, конструирует и строит не более одного Host и синхронно вызывает `Host.Start()` не более одного раза
- Runtime Bootstrap возвращает взаимоисключающий Success, structured Bootstrap Failure либо сохраняющий cause Startup Failure; Host остаётся владельцем operational composition, startup transaction и rollback
- Отдельный `PreparedRuntime` handoff был исключён принятым Architect rewrite
  DP-009 и не является target implementation
- Startup transaction публикует Listener только после успешного запуска и выполняет rollback полученного ресурса при ошибке, сохраняя исходную и rollback errors
- Host создает независимый root Runtime context после успешного запуска Listener; startup context не становится lifecycle context запущенного Runtime
- Runtime readiness становится true только после startup commit и сбрасывается в false в начале Stop
- Host владеет lifecycle-only Admission Gate, который открывается только в Running и закрывается до вызова Listener Stop
- Первый Host Stop выполняет единый production shutdown pipeline: закрывает Admission, вызывает `SessionManager.BeginShutdown`, запрашивает Stop для immutable Snapshot, отменяет root Runtime context, останавливает Listener и затем ожидает `SessionManager.Wait`
- Listener Stop и Manager Wait выполняются без удержания lifecycle mutex; их ошибки сохраняются вместе, а concurrent и repeated Stop получают один сохранённый terminal result
- Успешный Host Stop подтверждает `SessionManager.StateClosed` и пустое Reservation, Registration и Owner Lifetime Lease accounting; context-bounded Wait failure оставляет Manager правдиво в `Closing`
- Production composition до создания Listener проверяет startup-critical поля Snapshot: Runtime identity, Listener binding metadata, поддержку TLS и bounded Handshake timeout
- Включённый TLS явно отклоняется как unsupported runtime capability до открытия TCP socket; CertificateRef и PrivateKeyRef при этом не разрешаются и не включаются в текст ошибки
- `HandshakeSeconds` применяется как deadline всей pre-Upgrade evaluation; истёкшее решение не может перейти к `websocket.Accept`
- `ReadSeconds`, `WriteSeconds` и `IdleSeconds` сохраняются в immutable Snapshot как configured-but-inactive Runtime capabilities до отдельного эпика TLS and Listener settings; default Published Configuration остаётся исполнимой
- Реализован Listener Bootstrap, создающий потокобезопасный Listener из ListenerSnapshot
- Listener хранит локальную копию Host, Port и TLS configuration и поддерживает lifecycle `Created -> Running -> Stopping -> Stopped`
- Listener открывает TCP socket и запускает HTTP Server; `GET /ws` передаётся
  Handshake Handler, а fallback для остальных неподдерживаемых путей возвращает
  `501 Not Implemented`
- Listener корректно завершает HTTP Server, accept loop и связанные goroutine через graceful shutdown
- Listener передает `GET /ws` выделенному Handshake Handler; `websocket.Accept` выполняется только после начальной проверки Admission Gate, Authentication Allow Decision и финальной проверки Gate
- Immutable ConnectionContext содержит derived Runtime context, WebSocket connection и исходный HTTP request, используемый только синхронно при handoff
- Legacy `DefaultDispatcher` сохраняется для изолированных compatibility-тестов, но недостижим из production Runtime composition
- Production composition передает Handshake только read-only Admission capability и Runtime Context Provider; concrete Runtime Host в Handshake не передается
- Handshake преобразует HTTP metadata в transport-neutral AuthenticationRequest и выполняет Authentication до `websocket.Accept`
- Authentication Reject и operational error предотвращают Upgrade и возвращаются как HTTP rejection; Session создается только после успешного Upgrade
- Runtime composition явно передаёт Handshake и Listener минимальный callback для pre-Commit и transport terminal operational errors без diagnostics registry, event bus или глобального состояния
- Handshake сохраняет через `errors.Is` причину pre-Commit Session handoff failure в безопасной error-категории; Listener аналогично передаёт unexpected `http.Server.Serve` failure. Post-Commit outcomes проходят через synchronous Terminal Observer; DP-006 пока не определяет для него diagnostics backend
- Штатные `http.ErrServerClosed` и `net.ErrClosed` при Listener shutdown не создают ложные terminal error reports
- Первый Listener Stop выполняет shutdown, конкурентные Stop ожидают тот же terminal result с учетом cancellation context ожидающего caller, а повторный Stop возвращает сохраненный результат; независимые ошибки HTTP Shutdown и TCP Close сохраняются через `errors.Join`
- Disabled Authentication формирует explicit anonymous Principal без запуска Provider
- При включённой Authentication Bootstrap создаёт только enabled Providers и упорядочивает их по возрастанию `Priority`; активные Basic и asymmetric JWT configurations продолжают явно отклоняться до Listener Start
- Реализована минимальная WebSocket Session, которая после Authentication владеет соединением, хранит криптографически случайный ID, глубокую копию Principal, RemoteAddress и время создания
- Private transport-independent Session Core создаёт и хранит stable ID, deep-copied Principal, creation metadata и Handler до формирования transport-bound Session; Core не владеет WebSocket или lifecycle operations
- Package-private provisional preparation формирует из существующего Core один transport-bound Session в `Created` и один prospective Execution Owner в `PreCommit` как единый transaction-local unit; TransactionalDispatcher использует этот путь до Commit без запуска lifecycle, передачи ownership или publication Registration
- Provisional unit содержит private Cleanup machinery: один synchronous Cleanup выполняет Session Stop, затем panic-safe cancellation и наблюдение derived connection context, после чего возвращает один stable immutable categorized acknowledgement; repeated и concurrent calls разделяют execution/result, а committed Execution Owner вызывает Cleanup в terminal lifecycle
- Pre-Commit Session Bundle является одним structurally complete private Session-side object graph с фиксированными identities Core, Session, Owner, Cleanup и cancellation cell; TransactionalDispatcher владеет им до Commit, после которого ownership необратимо переходит Execution Owner
- Production TransactionalDispatcher подготавливает Session и dormant execution path, выполняет Reserve/Commit transaction и после успешного Commit возвращает `accepted=true` без post-Commit ownership; legacy synchronous Dispatcher остаётся только для изолированных тестов
- Создан независимый пакет `internal/sessionmanager` с потокобезопасным lifecycle skeleton `Open -> Closing -> Closed`
- Session Manager предоставляет неблокирующий идемпотентный `BeginShutdown`, context-bounded `Wait` и read-only наблюдение состояния; `Wait` не меняет accounting, а `Closed` достижим только при пустых Reservation, Registration и Owner Lifetime Lease sets
- Реализована первая полная граница Reservation transaction: `Reserve` создает уникальный за lifetime Manager `RegistrationID`, запрещает резервировать `SessionID`, уже занятый Reservation или committed Registration, и возвращает единственный Handle
- Abort атомарно удаляет Reservation, после чего ее `SessionID` можно использовать повторно; stale и concurrent Abort не имеют повторного accounting effect
- `Commit` является единственной linearization point появления Registration: он атомарно завершает Reservation, сохраняет тот же `RegistrationID` и публикует committed Registration ровно один раз; retry возвращает тот же ID только пока record существует, а после Complete сообщает `ErrRegistrationRemoved`
- `Complete(RegistrationID)` является единственной linearization point удаления Registration: первая валидная completion атомарно удаляет committed record и освобождает `SessionID`, а repeated, unknown и stale completion ничего не изменяют
- Reservation и committed Registration содержат только identity metadata, не хранят Session, WebSocket, Context или Runtime-компоненты и участвуют в shutdown accounting; Commit переносит одну accounting entry без изменения общего количества, Abort и Complete удаляют ее
- Committed registrations хранятся внутри Manager; `Lookup(SessionID)` возвращает только detached immutable `RegistrationView` с `SessionID`, `RegistrationID` и нормативным `StateRegistered`, не раскрывая Session или lifecycle capabilities
- Первый `BeginShutdown` атомарно фиксирует immutable capability-bearing `ShutdownSnapshot` только из committed registrations; Snapshot содержит `SessionID`, `RegistrationID` и Manager-bound Stop capability, не раскрывает Session или Owner, не меняется после Complete и одинаково возвращается повторными BeginShutdown
- Session не хранит исходный HTTP Request, Headers, Query, credentials, AuthenticationRequest или transport context wrappers
- Добавлена immutable transport-neutral Runtime Message модель для text и binary application messages с копированием payload и UTC-временем получения
- Session удерживает WebSocket-соединение открытым и выполняет единственный блокирующий read loop до закрытия клиента, отмены context, Stop или ошибки чтения
- Session предоставляет потокобезопасный `Send(context.Context, message.Message)` для сериализованной отправки text и binary Runtime Message без raw `[]byte` API; lifecycle mutex не удерживается во время WebSocket Write, допущенный до Stop write завершается с transport outcome, а новые writes после начала Stop отклоняются
- Добавлены immutable transport-neutral Runtime Message Context и Handler contract; Session создает отдельный Context для каждого прочитанного Message, не раскрывая HTTP или WebSocket transport, а при nil Handler сохраняет discard-поведение
- Реализован EchoHandler, возвращающий неизмененные text и binary Runtime Message исключительно через Session Send без доступа к WebSocket transport
- DP-005 Runtime Message Router завершён: optional Routing metadata проходит нормализацию и валидацию в ConfigurationVersion и `runtimeconfig.Builder`, после чего Runtime до Listener Start создаёт один immutable Router
- При наличии Routing Runtime использует strict compilation с единственным initial Handler reference `legacy`; при отсутствии Routing создаётся отдельный compatibility Router, сохраняющий прежнее Handler- или nil-discard-поведение
- Compiled Router хранит только enabled Routes, сортирует их один раз по возрастанию Priority, применяет exact case-sensitive Matchers и синхронно вызывает ровно один выбранный Handler
- Default Handler используется только после отсутствия explicit match; No Match не вызывает Handler, возвращает nil и позволяет Session продолжить read loop без legacy fallback для явно заданной Routing-секции
- Router переиспользуется всеми Session как единый immutable `message.Handler`; route compilation, sorting, normalization и Handler resolution на message hot path отсутствуют
- Middleware, Message Queue, Broadcast, публичный Session Manager Registry API и Message Persistence отсутствуют
- Архитектура Runtime принята в ADR-003; pre-Upgrade Handshake, transactional production Session handoff, Manager-aware Runtime shutdown и изолированный Configuration Loader реализованы, а production launch pipeline, operational diagnostics и supervision ещё отсутствуют
- Изолированный `internal/runtimelifecycle` реализует DP-010
  `PrepareStart`/`Start`/`Stop`/`Observe`, Owner-issued Launch Attempt,
  per-Instance serialization и truthful Host ownership без management или
  production integration
- Изолированный `internal/runtimemanagement` реализует DP-013 exact routing и
  authorization-before-mutation без Control Service integration
- Изолированный `internal/runtimeidentity` реализует DP-014 process-local
  in-memory Runtime Instance aggregate и append-only Launch Attempt history
- Изолированный `internal/runtimecommandidempotency` реализует DP-015
  process-local command claim/replay store и unresolved admission barriers

## Чего не существует

- Персистентного хранения Workspace
- Персистентного хранения Configuration
- Validation, Rollback и lifecycle Snapshot для Configuration Version
- PostgreSQL
- User-facing/Control Service API управления WebSocket-серверами
- Control Plane lifecycle управления экземплярами Runtime
- Production/external-durable Runtime Instance и Launch Attempt operational
  entities; process-local isolated DP-014 store существует
- Production integration Runtime Lifecycle Owner в Control Service
- Интеграция Configuration Loader в production launch pipeline
- Запуск Runtime и управление им из Control Service
- Реальный TLS listener и другие сетевые параметры Listener
- Применение read, write и idle Listener TimeoutSettings в Runtime
- Operational diagnostics и supervision полного Handshake Pipeline за
  пределами реализованных Authentication, configured timeout enforcement и
  Session shutdown wait set
- Проверка Basic credentials
- Асимметричные JWT algorithms, JWKS, OIDC и token revocation
- Реальные Secret Storage backend и подключение Resolver к Runtime Container еще не реализованы
- Инфраструктуры развертывания
- Инфраструктуры хранения данных
- Admin UI

Этот файл описывает реализованное состояние репозитория, а не запланированные возможности продукта. Обновляйте его только при существенном изменении этого состояния.

## Runtime Alpha Architecture Review

- 2026-07-14 выполнено двуязычное [Runtime Alpha Architecture Review](../docs/ru/reviews/runtime-alpha-review.md) ([English version](../docs/en/reviews/runtime-alpha-review.md)).
- Итоговая оценка: `Ready with findings`.
- Подтверждены immutable Snapshot, явный dependency injection, отсутствие import cycles и зависимости Runtime от Control Plane Repository, transport-neutral границы Authentication и Message, а также явная передача владения WebSocket-соединением.
- Authentication после WebSocket Upgrade, отсутствие production composition в Runtime Host, lifecycle lock во время Session Write и несогласованный результат concurrent Listener Stop устранены; неполная ограниченность lifecycle shutdown по context остается открытым finding.
- Проект остается alpha foundation и не заявляется как production-ready.

## Runtime Architectural Pattern

- Создано двуязычное активное архитектурное руководство [ARCH-001: Runtime Architectural Pattern](../docs/ru/architecture/ARCH-001-runtime-architectural-pattern.md) ([English version](../docs/en/architecture/ARCH-001-runtime-architectural-pattern.md)).
- ARCH-001 обобщает подтвержденный Alpha-вертикалью паттерн `Context -> Evaluation -> Decision -> Execution` без создания универсального Policy Engine или новых обязательных Go-контрактов.
- Зафиксированы Configuration First, проверяемые границы зависимостей, явная передача владения mutable resources, lifecycle и concurrency requirements, а также принцип Boring Core.
- Router определён и реализован по одобренному [DP-005: Runtime Message Router](../docs/ru/design/DP-005-runtime-message-router.md) ([English version](../docs/en/design/DP-005-runtime-message-router.md)); Delivery, Persistence и Plugin ABI остаются предметом будущих DP и при необходимости ADR.

## Master Engineering Plan

- Создан двуязычный живой инженерный [MASTER_PLAN](../docs/ru/roadmap/MASTER_PLAN.md) ([English version](../docs/en/roadmap/MASTER_PLAN.md)).
- План разделяет стадии зрелости Alpha, Beta, RC, 1.0 и 2.0+ без календарных сроков, performance promises или проектирования API будущих подсистем.
- Для Beta выделены эпики Handshake, Runtime Host, lifecycle hardening, Configuration validation, Router, Session Manager, Delivery, Persistence, TLS, Metrics, operational diagnostics и Plugin contracts.
- Обязательные свойства 1.0 отделены от возможностей 1.x и отложенных distributed-возможностей 2.0+.
- MASTER_PLAN не является release schedule, backlog или заменой DP, ADR, current state и архитектурных reviews.

## Runtime Handshake Pipeline Design

- Создан двуязычный Draft design [DP-001: Runtime Handshake Pipeline](../docs/ru/design/DP-001-runtime-handshake-pipeline.md) ([English version](../docs/en/design/DP-001-runtime-handshake-pipeline.md)).
- Выбрана концептуальная последовательность `Transport -> Handshake Context -> Evaluation -> Decision -> Upgrade -> Session`.
- Design переносит обязательную Authentication до WebSocket Upgrade, сохраняя transport-neutral Authentication Service и ownership Session после успешной передачи.
- Listener остается владельцем HTTP/WebSocket transport effects и не получает Provider-specific logic.
- Реализация следует основному порядку DP-001: Admission Gate, Authentication, Allow Decision, финальная проверка Gate, Upgrade и Session handoff.
- TASK-ARCH-REVIEW-010 подтвердил production-реализацию pre-Upgrade Authentication, bounded Handshake, Runtime-owned connection context и transactional ownership handoff. TASK-M10-002 добавил полный Runtime shutdown wait set; operational diagnostics/supervision остаются незавершёнными, поэтому status review DP-001 ещё требуется.
- Origin Policy, rate limiting, maintenance, IP filtering и Plugin ABI остаются future work без зафиксированных API.

## Runtime Host Composition Root Design

- Создан двуязычный Draft design [DP-002: Runtime Host Composition Root](../docs/ru/design/DP-002-runtime-host-composition-root.md) ([English version](../docs/en/design/DP-002-runtime-host-composition-root.md)).
- Runtime Host предложен как единственная production composition root одного экземпляра Runtime с явными dependency graph, startup rollback, shutdown ordering и readiness boundary.
- Определён lifecycle `Created -> Initialized -> Starting -> Running -> Stopping -> Stopped` с terminal state `Failed` и запретом Restart и in-place Reload.
- Host владеет root Runtime context, запускает Listener последним и закрывает admission до cancellation и cleanup в обратном порядке.
- Container не превращается в service locator; DI framework, reflection, generic component factories и Universal Component Registry запрещены.
- После публикации DP-002 Host стал production composition root и получил startup transaction, root Runtime context, readiness, lifecycle-only Admission Gate и полный Manager-aware shutdown wait set; `Failed` и supervision пока отсутствуют.
- TASK-M10-002 закрыл выявленный TASK-ARCH-REVIEW-010 разрыв Manager-aware shutdown. Terminal-состояние `Failed`, mandatory-component supervision и продолжение Host-owned cleanup после истечения context caller остаются невыполненными; статус DP-002 остаётся Draft до отдельной проверки.

## Runtime Session Manager Design

- Утверждён двуязычный design [DP-003: Runtime Session Manager](../docs/ru/design/DP-003-runtime-session-manager.md) ([English version](../docs/en/design/DP-003-runtime-session-manager.md)).
- DP-003 сохраняет нормативные контракты registration transaction, identity, Lookup, lifecycle Manager, shutdown accounting и реализованного capability-bearing Shutdown Snapshot; детальная модель execution из него удалена.
- Утверждён двуязычный design [DP-004: Per-Session Execution Boundary](../docs/ru/design/DP-004-per-session-execution-boundary.md) ([English version](../docs/en/design/DP-004-per-session-execution-boundary.md)).
- DP-004 определяет transport-independent Session Core и provisional preparation Session/Execution Owner до Commit без transfer ownership или visibility Registration.
- Commit является единственной irreversible publication point: Dispatcher заранее создаёт ровно один `CommitHandoff` и один dormant execution path и владеет всей pre-Commit transaction через panic-safe boundary; любой recoverable pre-Commit outcome публикует `NotCommitted` один раз, ждёт возврата path, освобождает owner-local values, выполняет Abort и возвращает `accepted=false`. Callback Runtime cancellation до Commit не существует.
- Integrated Commit под одной synchronization boundary создаёт Manager-bound immutable `CommitResult` из RegistrationID, Completion Adapter и Owner Lifetime Lease, публикует Registration, lease accounting и Stop capability Snapshot, сохраняет identity `CommitHandoff` и публикует через него `Committed` с полным result. Dormant path наблюдает тот же logical result, который возвращает Commit; post-Commit activation и capability delivery отсутствуют, rollback запрещён, а Registration удаляется только normal Complete.
- После Commit только Execution Owner устанавливает callback observation Host-owned root Runtime context до Start через race-safe registration-and-check contract; derived execution context callback не наблюдает, поэтому Session Cleanup не создаёт `RuntimeCanceled`. Все post-Commit causes используют единый order `Cleanup -> Complete -> Terminal Result -> Observer -> UnregisterAndDrain -> seal -> Terminal -> lease outcome`.
- Lifecycle Execution Owner упрощён до `PreCommit -> Committed -> Starting -> Running -> Terminalizing -> Terminal`; после Commit используется только normal terminal lifecycle.
- `ExplicitStop`, `RuntimeCanceled`, `NaturalCompletion`, `ExecutionFailure` и `RecoveredPanic` конкурируют через одну causal cell; первый source становится primary, последующие сохраняются только как bounded secondary categories.
- Terminal Result является immutable снимком execution-lifecycle outcomes, известных до Observer; поздние callback и cleanup outcomes относятся только к terminal accounting и не изменяют result. Panic-safe Session Cleanup возвращает immutable cancellation outcome. После confirmed cleanup callback `Cancellation Anomaly` допускает lifecycle до `Terminal`, но запрещает release lease; неподтверждённый lifetime callback оставляет owner в `Terminalizing`. Оба outcome оставляют `Manager.Wait` заблокированным без hidden retry.
- Release lease разрешён только после confirmed cancellation, возврата Complete и observer, confirmed `UnregisterAndDrain`, достижения `Terminal` и seal causal cell; release является последней Runtime-owned operation.
- После initiation Listener Stop drain HTTP handlers и terminalization owners выполняются параллельно; `Manager.Wait` начинается после возврата Listener Stop и не завершается до правдивой convergence Registration и owner-lifetime accounting.
- `BeginShutdown` и `Wait` разделяют неблокирующий transition shutdown и ожидание, а атомарный `Complete` реализован как единственная linearization point удаления committed Registration.
- Runtime Host остается владельцем Admission Gate и корневого Runtime context; Listener, Authentication, Router, Delivery, Persistence и diagnostics не входят в ответственность Session Manager.
- Реализованы transport-independent Session Core, lifecycle Manager, identity-safe registration transaction, read-only Lookup, immutable capability-bearing Shutdown Snapshot, Owner Lifetime Lease accounting, one-shot Execution Binding, bound Completion Adapter, Control Cell, Runtime Execution Owner и полный Owner-local terminal lifecycle до conditional Lifetime Lease release.
- Активное двуязычное руководство [ARCH-003: Runtime Foundation Migration Revision](../docs/ru/architecture/ARCH-003-runtime-migration-revision.md) ([English version](../docs/en/architecture/ARCH-003-runtime-migration-revision.md)) фиксирует завершённые Tasks 1–8 и пересмотренную последовательность Tasks 9–10; target architecture DP-003/DP-004 и production behavior не изменены.
- Текущая Go-реализация Task 9 завершает atomic Commit-to-Execution publication через domain-specific `CommitHandoff`. Successful Commit под одной Manager synchronization boundary публикует Registration, Registration-bound Completion Adapter, Owner Lifetime Lease accounting, Snapshot Stop capability и полный immutable `CommitResult`; тот же logical result получает dormant execution path. Repeated Commit проверяет identity handoff и не создаёт повторную публикацию или accounting.
- `CommitHandoffPublisher` является непрозрачной Manager-facing capability без доступных внешнему package committed-side операций. Только `ReservationHandle.Commit` может опубликовать `Committed`; Dispatcher сохраняет отдельную `NotCommittedPublisher` для pre-Commit failure path.
- Task 6 завершает Owner-local Runtime-cancellation и control-call primitives: explicit Stop и Runtime cancellation используют одну first-writer causal state; root Runtime context наблюдается только через явно устанавливаемую read-only dependency; control-call admission, outstanding accounting, panic-safe callback, idempotent unregister-and-drain result и seal после confirmed drain реализованы без production integration.
- Task 7 добавляет immutable, нативно сравнимый Terminal Result с validating construction и синхронный Terminal Observer contract.
- Task 8 завершает Owner-local terminal lifecycle: каждый claimed committed execution path проходит `Terminalizing -> Cleanup -> Completion -> Terminal Result -> Observer -> UnregisterAndDrain -> seal -> Terminal -> conditional lease release`; panic Completion и Observer изолируются, unconfirmed callback drain оставляет Owner в `Terminalizing`, а cancellation anomaly запрещает release lease после `Terminal`.
- TASK-M10-001 переключил production Runtime composition на TransactionalDispatcher: один Session Manager и synchronous composition-local Terminal Observer создаются на экземпляр Runtime, Router передаётся как `message.Handler`, а Handshake использует только transactional handoff. Live read-only input корневого Runtime context устанавливает post-Commit Runtime-cancellation observation без передачи Host-owned cancellation authority.
- Создан новый двуязычный Draft design [DP-006: Runtime Production Integration](../docs/ru/design/DP-006-runtime-production-integration.md) ([English version](../docs/en/design/DP-006-runtime-production-integration.md)), который фиксирует только production composition и shutdown cutover Task 10 без изменения DP-003, DP-004 или ARCH-003.
- TASK-M10-002 завершил production shutdown cutover: Host Stop один раз выполняет `BeginShutdown`, фиксирует capability-bearing Snapshot, вызывает каждый `RequestStop`, отменяет root Runtime context, останавливает Listener и после его возврата вызывает `Manager.Wait`.
- Реализация TASK-M10-002 закрывает acceptance criteria DP-006 8-11: Commit/BeginShutdown сохраняют одну Manager linearization boundary, успешный Stop требует закрытого Manager accounting, а Listener и context-bounded Wait errors не скрывают друг друга. Формальный статус DP-006 остаётся Draft до независимой полной проверки criteria 1-13.
- TASK-REV-013 Codex утвердил DP-003/DP-004 с одним неблокирующим clarity finding, TASK-REV-013 Kiro утвердил их без findings; TASK-DOC-016 синхронизировал Failure Matrix, composition root dependency и generic/production `accepted,error` semantics. DP-003 и DP-004 имеют статус Approved и интегрированы в production Session handoff; TASK-M10-002 завершил Runtime shutdown integration.

## Runtime Foundation Freeze

- Создан двуязычный [ARCH-002: Runtime Foundation Freeze](../docs/ru/architecture/ARCH-002-runtime-foundation-freeze.md) ([English version](../docs/en/architecture/ARCH-002-runtime-foundation-freeze.md)).
- Архитектурно стабильными признаны реализованные Runtime Host, production composition root, lifecycle, root Runtime context, startup transaction и rollback, readiness и lifecycle-only Admission Gate.
- Freeze фиксирует фактический lifecycle `Created -> Built -> Starting -> Running -> Stopping -> Stopped` и не объявляет реализованными предложенные в Draft DP-002 состояния `Initialized` или `Failed`.
- ARCH-002 оставил Router и Session ownership в полном Runtime shutdown wait
  set открытыми на момент freeze; впоследствии Router реализован по DP-005, а
  shutdown wait set — в TASK-M10-002 без изменения замороженных Runtime
  Foundation contracts. Delivery, Persistence, Operational Diagnostics и
  supervision остаются открытой архитектурой.
- Изменение замороженных архитектурных обязанностей, ownership или lifecycle-семантики требует нового сфокусированного DP или ADR.

## Handshake Runtime Dependency Boundary

- Принят двуязычный [ADR-0004: Handshake Runtime Dependency Boundary](../docs/ru/adr/0004-handshake-runtime-dependencies.md) ([English version](../docs/en/adr/0004-handshake-runtime-dependencies.md)).
- Host остается единственным владельцем Admission Gate и cancellation корневого Runtime context; Handshake получает только живые read-only capabilities через явную constructor injection.
- Draft DP-001 и DP-002 синхронизированы с ADR-0004: composition bridge передает Handshake admission permission и Runtime context access без зависимости от concrete Host.
- Handshake должен проверять admission до Authentication и повторно непосредственно перед `websocket.Accept`; Runtime context holder создается вместе с Host и активируется только при успешном startup commit.
- Финальная проверка admission непосредственно перед `websocket.Accept` является linearization point входа в admission commit; закрытие Gate до нее запрещает Upgrade.
- Session context должен создаваться как дочерний от активного Runtime context, а не от `http.Request.Context()`; root `CancelFunc` Handshake не раскрывается.
- ADR-0004 реализован минимальным capability bridge: Handshake не зависит от concrete Host, а Session context создается от активного Runtime context.

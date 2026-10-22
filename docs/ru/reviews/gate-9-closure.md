# Техническое закрытие Gate 9 и evidence

[English version](../../en/reviews/gate-9-closure.md)

**Решение:** технически закрыт явным решением Owner 2026-10-22.
**Baseline оценки:** `af61085843fd1666aec68d635daef944f2993972`.
**Documentation task:** [TASK-108](../../tasks/TASK-108-GATE-9-TECHNICAL-CLOSURE.md).

## Scope и authority

[MASTER_PLAN Gate 9](../roadmap/MASTER_PLAN.md#operational-completion) требует
безопасного разрешения certificate references и применения TLS, handshake, read,
write и idle limits. Эта запись декомпозирует требование и проверяет совместимость
с существующими Authentication, startup/rollback/shutdown и Session/Manager
contracts. Она фиксирует решение Owner после полной evidence assessment;
timeout delivery или TASK-107 Acceptance отдельно gate не закрывали.

Закрытие охватывает существующий single-node Runtime и TLS execution v1. Оно не
означает production readiness, завершение всех Beta gates, Windows/native
qualification, новые cancellation/shutdown guarantees или Approved/Frozen amendment.
Architecture B, PoC и Control Service Runtime activation остаются вне scope.

Приоритет источников сохраняется по [PROCESS-001](../../engineering/PROCESS-001-AI-DEVELOPMENT-WORKFLOW.md):
accepted ADR, Active/Frozen ARCH и Approved DP выше implementation/navigation.
[ARCH-002](../architecture/ARCH-002-runtime-foundation-freeze.md) фиксирует Host
ownership/lifecycle и явно исключает TLS и effective timeout behavior.
Active ARCH-003/005 и Approved DP-003/004 не меняются.
[DP-008](../design/DP-008-snapshot-builder-contract.md) — **Draft**, с реализованной
Snapshot boundary; закрытие не повышает его статус до Approved.

## Closure Evidence Matrix

Все implementation/test links ниже указывают на bytes assessment baseline.
Тесты названы exact function names. Evidence их выполнения находится в
исторических task records и exact-SHA hosted runs ниже; TASK-108 не заявляет
новый product test run. `SATISFIED` ограничен существующим контрактом,
а не безусловной hard-stop, delivery или race-freedom guarantee.

| Критерий | Классификация | Exact implementation и verification evidence |
| --- | --- | --- |
| Безопасное разрешение certificate/key | SATISFIED | [PrepareTLS](../../../internal/listener/tls.go): одна пара на Start attempt, exact Start context, disabled nonresolution, reference/key-pair/leaf validation, очистка raw buffers и safe error identity. [TLS tests](../../../internal/listener/tls_test.go): `TestTLSPreparationValidationAndRawBufferCleanup`, `TestTLSPreparationPreservesIdentityWithoutExposingCause`, `TestTLSPreparationRejectsLeafPolicy`, `TestTLSPreparationCancellationBoundaries`. TASK-107 E-107-003/004/006. |
| Effective TLS/WSS, отсутствие plaintext application fallback | SATISFIED | [Listener.Start](../../../internal/listener/listener.go), [TLS preparation](../../../internal/listener/tls.go): configured minimum 1.2/1.3, maximum 1.3, HTTP1 only. `TestTLSListenerVersionAndProtocolNegotiation`, `TestTLSListenerPlaintextNeverReachesApplication` в Listener TLS tests; `TestTLSRuntimeWSSAuthenticationAndEcho` в [Runtime TLS tests](../../../internal/runtime/tls_test.go). |
| HandshakeSeconds | SATISFIED | [Composition](../../../internal/runtime/composition.go) применяет request/pre-Upgrade H; TLS Listener задаёт ReadHeaderTimeout H. `TestHandshakeTimeoutHandlerUsesConfiguredDuration` в [configuration tests](../../../internal/runtime/configuration_validation_test.go); `TestTLSListenerHandshakeBudgetAndStop` доказывает silent/partial TLS expiry и Stop. TLS handshake, HTTP headers и request phases имеют отдельные H budgets, а не общий accept-to-Upgrade deadline. |
| ReadSeconds | SATISFIED | [Session.readApplicationMessage](../../../internal/session/session.go), [read note](../design/runtime-read-timeout.md): zero disabled; 1..86400, единый full-message budget с fragments/control frames, без Handler. [Read tests](../../../internal/session/read_timeout_test.go): `TestReadTimeoutFreshBudgetExcludesHandlerAndCancelsChildren`, `TestReadTimeoutPartialMessageAndControlFramesDoNotResetBudget`, cancellation/error/panic cases. TASK-104 E-104-005/006/007, 325 stress executions. |
| WriteSeconds | SATISFIED | Session.Send и [write note](../design/runtime-write-timeout.md): 1..300, queue и network Write разделяют один budget; caller deadline/cancellation сохранены; Stop независим от gate. [Write tests](../../../internal/session/write_timeout_test.go): `TestWriteTimeoutQueueAndNetworkUseOneBudget`, `TestWriteTimeoutQueuedDeadlineDoesNotWrite`, `TestWriteTimeoutConcurrentStopDoesNotWaitForAdmittedWrite`. Runtime TLS `TestTLSRuntimeWriteTimeoutOnNonreadingWSSPeer` использует 16 MiB reply и errors.Is DeadlineExceeded. TASK-103 E-103-002/003/004/005. |
| IdleSeconds | SATISFIED | [idleTracker](../../../internal/session/idle_timeout.go), [idle note](../design/runtime-idle-timeout.md): zero disabled; 1..86400 от Run/последнего успешного полного application Read/Write; Handler включён; partial/control/failed I/O не обновляют deadline. [Idle tests](../../../internal/session/idle_timeout_test.go): successful Read/Write refresh, fragments/control, overdue activity, stale wake и contention; `TestIdleTimeoutReadPreservesTransportCause` проверяет обе errors.Is identities после B-106-001. TASK-106 E-106-011/013/014, 500 rework stress executions. |
| Immutable Snapshot propagation и validation | SATISFIED | Composition передаёт read/write/idle durations; [transactional Dispatcher](../../../internal/session/transactional_dispatcher.go) назначает Session fields до dormant launch/Commit. `TestRuntimeWriteTimeoutFromPublishedSnapshot`, `TestRuntimeReadTimeoutFromPublishedSnapshot`, `TestRuntimeIdleTimeoutFromPublishedSnapshot` в соответствующих Runtime timeout files; bounds и read/idle direct guards в configuration tests; Session pre-Commit assignment tests. Ограничение R-103-N1 сохранено ниже. |
| Authentication и message compatibility | SATISFIED | Pre-Upgrade Authentication и Session handoff не изменены. `TestTLSRuntimeWSSAuthenticationAndEcho` доказывает missing credentials -> 401 до Upgrade, API Key/HMAC JWT success и real trusted WSS text/binary echo. `TestTLSRuntimeReadAndIdleTimeoutsPreserveRoot` и blocked-write test сохраняют Session budgets поверх WSS. TASK-107 E-107-006. Basic/asymmetric JWT не реализуются заново. |
| Startup/rollback и Frozen Host compatibility | SATISFIED | [Host](../../../internal/runtime/host.go), [composition](../../../internal/runtime/composition.go), [startup transaction](../../../internal/runtime/startup_transaction.go): validation/material preparation до socket, readiness/admission false при failure, независимый successful Runtime lifetime. `TestTLSRuntimePreparationFailureBeforeSocketAndRedactedIdentity`, `TestTLSRuntimeStartCancellationWhileResolverBlockedAndStopWaits`, `TestTLSRuntimeDisabledDoesNotResolveAndStartupContextIsNotLifetime`; [startup transaction regressions](../../../internal/runtime/startup_transaction_test.go). TASK-107 Architecture conformance PASS без normative amendment. |
| Shutdown, single cleanup и Manager convergence | SATISFIED | Сохранён порядок Admission close -> Manager snapshot RequestStop -> root cancellation -> Listener Stop -> Manager.Wait. Run/Stop join-ят idle watchdog; Owner сохраняет terminal cleanup. `TestIdleTimeoutStopJoinsBeforeTransportClose`, `TestIdleTimeoutOwnerCleanupAndCancellationConverge`, `TestTLSRuntimeConcurrentWSSAndStopConverge`, `TestTLSListenerHandshakeBudgetAndStop`; [Host shutdown tests](../../../internal/runtime/shutdown_orchestration_test.go). Deadline-returning Stop не доказывает join всех net/http goroutines. |
| Concurrency, regressions и independent engineering gates | SATISFIED | TASK-103: 250 stress executions, Review Approved with Findings, blocking0; TASK-104: 325, Review APPROVED0/0; TASK-106: 500 после rework, R-106-002 Approved0/0, B-106-001 resolved; TASK-107: 125 TLS stress executions, R-107-001 APPROVED0/0. Каждая task фиксирует Architecture, Tester, PROCESS-002, Scope Audit, Acceptance и integrity. Exact final merge Linux go-ci проходит full test/race/build/vet/format/tidy/module checks. |

В bounded Gate 9 scope нет технических строк PARTIALLY SATISFIED, NOT SATISFIED
или NEEDS FORENSIC. Documentation synchronization — deliverable TASK-108;
её Review/Acceptance определяются только matching task envelope,
а не самоутверждаются этой матрицей.

## Exact publication evidence

Read-only GitHub verification 2026-10-22 подтверждает все четыре PR MERGED в
main. Local Git доказывает ancestry каждого source в assessment baseline и
exact equality trees внутри каждой source/merge pair; equality между разными
tasks не заявляется. Parents TASK-107 merge — TASK-106 merge и TASK-107 source.

| Task / PR | Exact source SHA | Exact merge SHA | Одинаковый source/merge tree |
| --- | --- | --- | --- |
| [TASK-103 / PR103](https://github.com/dsdred/universal-websocket-platform/pull/103) | `a2742ab927f7a23fd10a92db986ce36ff6a987e8` | `aed7cc8f1693cf8bb48d71532fe543f09d2d1191` | `25795d4bcdc5c722f1a16fe69197ca913f0a123f` |
| [TASK-104 / PR104](https://github.com/dsdred/universal-websocket-platform/pull/104) | `63d99906924195115cb9dc260762ca050cb675b8` | `fb916df77e697fb7a45dc7bb6e8d0150411f2d82` | `59afd6cdc78c14a96326c8d73e38625761de70e3` |
| [TASK-106 / PR106](https://github.com/dsdred/universal-websocket-platform/pull/106) | `909abc7f67a36f6709ba343d0cf5b63bc3509c85` | `9151db9e32e0b351d403087e4f1361b31b134fac` | `dd11a05d473683ea53faf67d197eb65362f8d8c4` |
| [TASK-107 / PR107](https://github.com/dsdred/universal-websocket-platform/pull/107) | `8f97a9c8ca9ad6831911c6b7fe133c16aa588ee9` | `af61085843fd1666aec68d635daef944f2993972` | `512840af96c547e312fd146aa0cf4069e5f322e0` |

| Task | Exact source pull_request run | Exact merge push run | Результат |
| --- | --- | --- | --- |
| 103 / 104 | Не обнаружены | Не обнаружены | GitHub Actions runs и check-runs равны нулю на обоих exact source и merge SHA. CI введён позже TASK-105; historical CI PASS не выдумывается. |
| 106 | [39728769463](https://github.com/dsdred/universal-websocket-platform/actions/runs/39728769463) | [39729079012](https://github.com/dsdred/universal-websocket-platform/actions/runs/39729079012) | completed / SUCCESS |
| 107 | [39773286256](https://github.com/dsdred/universal-websocket-platform/actions/runs/39773286256) | [39773713007](https://github.com/dsdred/universal-websocket-platform/actions/runs/39773713007) | completed / SUCCESS |

Все четыре наблюдавшихся runs имеют successful `go-ci` steps: Check tracked Go
formatting, Build, Test, Vet, Race test, Check module tidiness, Check unchanged
module files. [Workflow](../../../.github/workflows/go-ci.yml): Ubuntu24.04,
Go1.26.5, race CGO_ENABLED=1. Final merge CI проверяет совокупную Gate 9
implementation, включая ранние read/write suites. Это baseline product evidence,
а не будущая TASK-108 publication authority; отдельно разрешённая publication
по-прежнему требует successful required go-ci на своём exact PR head.

Historical engineering evidence сохранено неизменным:
[TASK-103](../../tasks/TASK-103-RUNTIME-WRITE-TIMEOUT.md),
[TASK-104](../../tasks/TASK-104-RUNTIME-READ-TIMEOUT.md),
[TASK-106](../../tasks/TASK-106-RUNTIME-IDLE-TIMEOUT.md),
[TASK-107](../../tasks/TASK-107-RUNTIME-TLS-ENFORCEMENT.md).
Их closure-time no-commit/publication wording историческое и не является live
инструкцией после доказанной последующей publication.

## Сохранённые ограничения и другие gates

- Local race остаётся **NOT RUN**: CGO/compiler недоступны. Stress не доказывает
  race freedom. Hosted Linux race **PASS** не квалифицирует Windows/native TASK-102.
- Checkout tidy сохраняет existing CRLF-only limitation, это не local tidy PASS;
  exact-object source-fixture tidy **PASS**, hosted Linux tidy **PASS**. Module
  identities не менялись в TASK-107 или этом documentation cycle.
- Resolver, Handler и transport cancellation остаются cooperative. Stop во время
  Starting ждёт Start; новой startup cancellation authority или hard-kill
  guarantee нет. Deadline-returning Stop не доказывает все net/http internal joins.
  Manager accounting не должен заявлять ложный terminal completion.
- TLS v1 содержит только одну startup-loaded pair: без rotation/reload, ACME,
  SNI certificate selection или mTLS. Нового Secret backend нет. Client trust,
  hostname и revocation validation остаются ответственностью клиента; parsed Go
  crypto objects не имеют secure-zeroization guarantee.
- TASK-102 остаётся source-only Accepted, committed/unpublished на
  `8bfa2e22a4b6310a71f6d36c8724b1adedd727ad`; native build/link NOT RUN,
  toolchain NOT QUALIFIED. Её branch/content не затронуты и исключены из main.
- R-103-N1 сохраняется как **nonblocking**: WriteSeconds invalid-input tests
  проверяют Builder rejection, а не прямой Runtime defensive guard rejection.
  Code inspection, immutable Snapshot boundary и constructor tests поддерживают
  принятый результат; direct Runtime-guard branch coverage не заявляется.
- Другие Beta gates и release criteria остаются независимыми. Delivery/Persistence,
  Metrics, operational diagnostics, Origin policy, более широкие Authentication
  или activation work этим закрытием не завершаются и не активируются.

## Documentation synchronization и follow-up

TASK-108 сверяет MASTER_PLAN/debt, current-state, PROJECT_CONTEXT, task index,
review/design navigation и focused ReadSeconds/TLS implementation statuses.
Published write/idle status notes уже соответствуют фактам и не требуют
переписывания истории. Root README capability descriptions остаются верными.
CHANGELOG неприменим: закрытие не меняет user-facing behavior или release.

Для этого technical closure новый product prerequisite не требуется. Optional
direct WriteSeconds defensive-guard test strengthening и отдельная TASK-102
native qualification остаются bounded future work, не назначаются и не
активируются. Commit/publication документации требуют отдельных Owner gates.

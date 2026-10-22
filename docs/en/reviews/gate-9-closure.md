# Gate 9 technical closure and evidence

[Russian version](../../ru/reviews/gate-9-closure.md)

**Decision:** technically closed by explicit Owner decision on 2026-10-22.
**Assessment baseline:** `af61085843fd1666aec68d635daef944f2993972`.
**Documentation task:** [TASK-108](../../tasks/TASK-108-GATE-9-TECHNICAL-CLOSURE.md).

## Scope and authority

[MASTER_PLAN Gate 9](../roadmap/MASTER_PLAN.md#operational-completion) requires
safe certificate-reference resolution and effective TLS, handshake, read, write
and idle limits. This record decomposes that requirement and audits compatibility
with existing Authentication, startup/rollback/shutdown and Session/Manager
contracts. It records the Owner decision after complete evidence assessment;
neither timeout delivery nor TASK-107 Acceptance alone closed the gate.

Closure covers the existing single-node Runtime and TLS execution v1. It does not
declare production readiness, completion of all Beta gates, Windows/native
qualification, new cancellation/shutdown guarantees or an Approved/Frozen amendment.
Architecture B, PoC and Control Service Runtime activation remain outside scope.

Source precedence remains [PROCESS-001](../../engineering/PROCESS-001-AI-DEVELOPMENT-WORKFLOW.md):
accepted ADRs, Active/Frozen ARCH and Approved DP precede implementation and
navigation. [ARCH-002](../architecture/ARCH-002-runtime-foundation-freeze.md)
freezes Host ownership/lifecycle, explicitly excluding TLS and effective timeout
behavior. Active ARCH-003/005 and Approved DP-003/004 remain unchanged.
[DP-008](../design/DP-008-snapshot-builder-contract.md) is **Draft**, with an
implemented Snapshot boundary; it is not promoted to Approved by this closure.

## Closure Evidence Matrix

Every implementation/test link below refers to bytes present at the assessment
baseline. Tests are identified by exact function names. Their executed evidence
is in the historical task records and exact-SHA hosted runs below; TASK-108 does
not claim a new product test run. `SATISFIED` is scoped to the existing contract,
not an unlimited hard-stop, delivery or race-freedom guarantee.

| Criterion | Classification | Exact implementation and verification evidence |
| --- | --- | --- |
| Safe certificate/key resolution | SATISFIED | [PrepareTLS](../../../internal/listener/tls.go): one pair per Start attempt, exact Start context, disabled nonresolution, reference/key-pair/leaf validation, raw buffer clearing and safe error identity. [TLS tests](../../../internal/listener/tls_test.go): `TestTLSPreparationValidationAndRawBufferCleanup`, `TestTLSPreparationPreservesIdentityWithoutExposingCause`, `TestTLSPreparationRejectsLeafPolicy`, `TestTLSPreparationCancellationBoundaries`. TASK-107 E-107-003/004/006. |
| Effective TLS/WSS, no plaintext application fallback | SATISFIED | [Listener.Start](../../../internal/listener/listener.go), [TLS preparation](../../../internal/listener/tls.go): configured minimum 1.2/1.3, maximum 1.3, HTTP1 only. `TestTLSListenerVersionAndProtocolNegotiation`, `TestTLSListenerPlaintextNeverReachesApplication` in Listener TLS tests; `TestTLSRuntimeWSSAuthenticationAndEcho` in [Runtime TLS tests](../../../internal/runtime/tls_test.go). |
| HandshakeSeconds | SATISFIED | [Composition](../../../internal/runtime/composition.go) applies request/pre-Upgrade H; TLS Listener sets ReadHeaderTimeout H. `TestHandshakeTimeoutHandlerUsesConfiguredDuration` in [configuration tests](../../../internal/runtime/configuration_validation_test.go); `TestTLSListenerHandshakeBudgetAndStop` proves silent/partial TLS expiry and Stop. TLS handshake, HTTP headers and request phases have separate H budgets, not one accept-to-Upgrade deadline. |
| ReadSeconds | SATISFIED | [Session.readApplicationMessage](../../../internal/session/session.go), [read note](../design/runtime-read-timeout.md): zero disabled; 1..86400, one full-message budget including fragments/control frames, excluding Handler. [Read tests](../../../internal/session/read_timeout_test.go): `TestReadTimeoutFreshBudgetExcludesHandlerAndCancelsChildren`, `TestReadTimeoutPartialMessageAndControlFramesDoNotResetBudget`, cancellation/error/panic cases. TASK-104 E-104-005/006/007, 325 stress executions. |
| WriteSeconds | SATISFIED | Session.Send and [write note](../design/runtime-write-timeout.md): 1..300, queue plus network Write share one budget; caller deadline/cancellation preserved; Stop independent of gate. [Write tests](../../../internal/session/write_timeout_test.go): `TestWriteTimeoutQueueAndNetworkUseOneBudget`, `TestWriteTimeoutQueuedDeadlineDoesNotWrite`, `TestWriteTimeoutConcurrentStopDoesNotWaitForAdmittedWrite`. Runtime TLS `TestTLSRuntimeWriteTimeoutOnNonreadingWSSPeer` uses 16 MiB reply and errors.Is DeadlineExceeded. TASK-103 E-103-002/003/004/005. |
| IdleSeconds | SATISFIED | [idleTracker](../../../internal/session/idle_timeout.go), [idle note](../design/runtime-idle-timeout.md): zero disabled; 1..86400 since Run/last successful full application Read/Write; Handler included; partial/control/failed I/O never refresh. [Idle tests](../../../internal/session/idle_timeout_test.go): successful Read/Write refresh, fragments/control, overdue activity, stale wake and contention; `TestIdleTimeoutReadPreservesTransportCause` verifies both errors.Is identities after B-106-001. TASK-106 E-106-011/013/014, 500 rework stress executions. |
| Immutable Snapshot propagation and validation | SATISFIED | Composition passes read/write/idle durations; [transactional Dispatcher](../../../internal/session/transactional_dispatcher.go) assigns Session fields before dormant launch/Commit. `TestRuntimeWriteTimeoutFromPublishedSnapshot`, `TestRuntimeReadTimeoutFromPublishedSnapshot`, `TestRuntimeIdleTimeoutFromPublishedSnapshot` in corresponding Runtime timeout files; bounds and read/idle direct guards in configuration tests; Session pre-Commit assignment tests. R-103-N1 limitation remains below. |
| Authentication and message compatibility | SATISFIED | Unchanged pre-Upgrade Authentication and Session handoff. `TestTLSRuntimeWSSAuthenticationAndEcho` proves missing credentials reject with 401 before Upgrade, API Key/HMAC JWT success and real trusted WSS text/binary echo. `TestTLSRuntimeReadAndIdleTimeoutsPreserveRoot` and blocked-write test preserve Session budgets over WSS. TASK-107 E-107-006. Basic/asymmetric JWT are not newly implemented. |
| Startup/rollback and Frozen Host compatibility | SATISFIED | [Host](../../../internal/runtime/host.go), [composition](../../../internal/runtime/composition.go), [startup transaction](../../../internal/runtime/startup_transaction.go): validation/material preparation before socket, readiness/admission false on failure, independent successful Runtime lifetime. `TestTLSRuntimePreparationFailureBeforeSocketAndRedactedIdentity`, `TestTLSRuntimeStartCancellationWhileResolverBlockedAndStopWaits`, `TestTLSRuntimeDisabledDoesNotResolveAndStartupContextIsNotLifetime`; [startup transaction regressions](../../../internal/runtime/startup_transaction_test.go). TASK-107 Architecture conformance PASS, no normative amendment. |
| Shutdown, single cleanup and Manager convergence | SATISFIED | Existing Admission close -> Manager snapshot RequestStop -> root cancellation -> Listener Stop -> Manager.Wait order retained. Idle watchdog joined by Run/Stop; Owner retains terminal cleanup. `TestIdleTimeoutStopJoinsBeforeTransportClose`, `TestIdleTimeoutOwnerCleanupAndCancellationConverge`, `TestTLSRuntimeConcurrentWSSAndStopConverge`, `TestTLSListenerHandshakeBudgetAndStop`; [Host shutdown tests](../../../internal/runtime/shutdown_orchestration_test.go). Deadline-returning Stop does not prove every net/http goroutine joined. |
| Concurrency, regressions and independent engineering gates | SATISFIED | TASK-103: 250 stress executions, Review Approved with Findings, blocking0; TASK-104: 325, Review APPROVED0/0; TASK-106: 500 after rework, R-106-002 Approved0/0, B-106-001 resolved; TASK-107: 125 TLS stress executions, R-107-001 APPROVED0/0. Each task records Architecture, Tester, PROCESS-002, Scope Audit, Acceptance and integrity. Exact final merge Linux go-ci passes full test/race/build/vet/format/tidy/module checks. |

No technical row is PARTIALLY SATISFIED, NOT SATISFIED or NEEDS FORENSIC within
this bounded Gate 9 scope. Documentation synchronization is the TASK-108
deliverable; its Review/Acceptance are resolved only from the matching task
envelope, not self-certified by this matrix.

## Exact publication evidence

Read-only GitHub verification on 2026-10-22 confirms all four PRs MERGED into
main. Local Git proves each source is an ancestor of the assessment baseline and
each source/merge pair has exactly equal trees; no cross-task tree equality is
claimed. TASK-107 merge parents are the TASK-106 merge and TASK-107 source.

| Task / PR | Exact source SHA | Exact merge SHA | Equal source/merge tree |
| --- | --- | --- | --- |
| [TASK-103 / PR103](https://github.com/dsdred/universal-websocket-platform/pull/103) | `a2742ab927f7a23fd10a92db986ce36ff6a987e8` | `aed7cc8f1693cf8bb48d71532fe543f09d2d1191` | `25795d4bcdc5c722f1a16fe69197ca913f0a123f` |
| [TASK-104 / PR104](https://github.com/dsdred/universal-websocket-platform/pull/104) | `63d99906924195115cb9dc260762ca050cb675b8` | `fb916df77e697fb7a45dc7bb6e8d0150411f2d82` | `59afd6cdc78c14a96326c8d73e38625761de70e3` |
| [TASK-106 / PR106](https://github.com/dsdred/universal-websocket-platform/pull/106) | `909abc7f67a36f6709ba343d0cf5b63bc3509c85` | `9151db9e32e0b351d403087e4f1361b31b134fac` | `dd11a05d473683ea53faf67d197eb65362f8d8c4` |
| [TASK-107 / PR107](https://github.com/dsdred/universal-websocket-platform/pull/107) | `8f97a9c8ca9ad6831911c6b7fe133c16aa588ee9` | `af61085843fd1666aec68d635daef944f2993972` | `512840af96c547e312fd146aa0cf4069e5f322e0` |

| Task | Exact source pull_request run | Exact merge push run | Result |
| --- | --- | --- | --- |
| 103 / 104 | None observed | None observed | GitHub Actions runs and check-runs are zero on both exact source and merge SHAs. CI was introduced later by TASK-105; no historical CI PASS is fabricated. |
| 106 | [39728769463](https://github.com/dsdred/universal-websocket-platform/actions/runs/39728769463) | [39729079012](https://github.com/dsdred/universal-websocket-platform/actions/runs/39729079012) | completed / SUCCESS |
| 107 | [39773286256](https://github.com/dsdred/universal-websocket-platform/actions/runs/39773286256) | [39773713007](https://github.com/dsdred/universal-websocket-platform/actions/runs/39773713007) | completed / SUCCESS |

All four observed runs have successful `go-ci` steps: Check tracked Go formatting,
Build, Test, Vet, Race test, Check module tidiness, Check unchanged module files.
[Workflow](../../../.github/workflows/go-ci.yml): Ubuntu24.04, Go1.26.5, race
CGO_ENABLED=1. Final merge CI tests the cumulative Gate 9 implementation, including
the earlier read/write suites. This is baseline product evidence, not future
TASK-108 publication authority; separately authorized publication still requires
successful required go-ci on its exact PR head.

Historical engineering evidence is preserved unchanged:
[TASK-103](../../tasks/TASK-103-RUNTIME-WRITE-TIMEOUT.md),
[TASK-104](../../tasks/TASK-104-RUNTIME-READ-TIMEOUT.md),
[TASK-106](../../tasks/TASK-106-RUNTIME-IDLE-TIMEOUT.md),
[TASK-107](../../tasks/TASK-107-RUNTIME-TLS-ENFORCEMENT.md).
Their closure-time no-commit/publication wording is historical, not a live
instruction after the proven later publication.

## Retained limitations and other gates

- Local race remains **NOT RUN**: CGO/compiler unavailable. Stress is not race
  freedom. Hosted Linux race **PASS** does not qualify Windows/native TASK-102.
- Checkout tidy retains its existing CRLF-only limitation, not local tidy PASS;
  exact-object source-fixture tidy **PASS**, hosted Linux tidy **PASS**. Module
  identities are unchanged by TASK-107 or this documentation cycle.
- Resolver, Handler and transport cancellation remain cooperative. Stop during
  Starting waits for Start; no new startup cancellation authority or hard-kill
  guarantee. Deadline-returning Stop does not prove all net/http internal joins.
  Manager accounting must not report false terminal completion.
- TLS v1 has one startup-loaded pair only: no rotation/reload, ACME, SNI
  certificate selection or mTLS. No new Secret backend. Client trust, hostname
  and revocation validation remain client responsibilities; parsed Go crypto
  objects have no secure-zeroization guarantee.
- TASK-102 remains source-only Accepted, committed/unpublished at
  `8bfa2e22a4b6310a71f6d36c8724b1adedd727ad`; native build/link NOT RUN,
  toolchain NOT QUALIFIED. Its branch/content is untouched and excluded from main.
- R-103-N1 remains **nonblocking**: WriteSeconds invalid-input tests exercise
  Builder rejection rather than direct Runtime defensive guard rejection. Code
  inspection, immutable Snapshot boundary and constructor tests support the
  accepted result; no direct Runtime-guard branch coverage is claimed.
- Other Beta gates and release criteria remain independent. Delivery/Persistence,
  Metrics, operational diagnostics, Origin policy and broader Authentication or
  activation work are not completed or activated by this closure.

## Documentation synchronization and follow-up

TASK-108 reconciles MASTER_PLAN/debt, current-state, PROJECT_CONTEXT, task index,
review/design navigation and focused ReadSeconds/TLS implementation statuses.
Published write/idle status notes already agree and need no historical rewrite.
Root README capability descriptions remain valid. CHANGELOG is not applicable:
closure changes no user-facing behavior or release.

No new product prerequisite is required for this technical closure. Optional
direct WriteSeconds defensive-guard test strengthening and separate TASK-102
native qualification remain bounded future work, not assigned or activated here.
Commit/publication of this documentation requires separate Owner gates.

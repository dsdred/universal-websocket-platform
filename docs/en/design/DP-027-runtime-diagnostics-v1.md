# DP-027: Runtime Diagnostics v1

[Russian version](../../ru/design/DP-027-runtime-diagnostics-v1.md)

## 1. Status and scope

- **Design Status:** Draft.
- **Implementation Status:** Planned.
- **Target Milestone:** Beta, Gate 11; this document does not close Gate 11.
- **Task:** [TASK-109](../../tasks/TASK-109-RUNTIME-DIAGNOSTICS-V1-DESIGN.md).
- **Source baseline:** `67ca1b200d42fda5be8b6dc8741d8f0bc24bd2a6`.

The accepted deliverable may be design/readiness evidence without normative
Approval or implementation authorization. No new behavior is implemented here.
The only proposed source kinds are unexpected Listener Serve return and
Handshake Authentication Provider operational failure. Existing behavior,
proposed eligibility and future implementation proofs are distinguished below.

## 2. Authority and invariants

[ADR-0002](../adr/0002-configuration-dsl.md) and
[ADR-0003](../adr/0003-runtime-architecture.md) preserve Configuration First,
explicit dependencies and transport-neutral Authentication.
[ADR-0004](../adr/0004-handshake-runtime-dependencies.md) keeps the Host capability
bundle closed to admission and Runtime context. A separate diagnostic dependency
must not be added to that bundle or become a service locator.
[ARCH-002](../architecture/ARCH-002-runtime-foundation-freeze.md) keeps Host
startup/rollback, readiness, admission, root cancellation and Stop unchanged.
[ARCH-003](../architecture/ARCH-003-runtime-migration-revision.md) and Approved
[DP-003](DP-003-runtime-session-manager.md)/[DP-004](DP-004-per-session-execution-boundary.md)
preserve Session ownership, terminal ordering, Observer and Manager accounting.

Reporting observes a source fact. It cannot stop/restart Runtime, change readiness,
open/close admission, create a Session, reclaim a handed-off connection, publish
Stopped or claim cleanup completion. No Approved/Frozen amendment is proposed;
any implementation needing one must stop for an Owner decision.

## 3. Baseline source-backed inventory

| Path / boundary | Current fact | Diagnostic classification / coverage |
| --- | --- | --- |
| [Listener.serve](../../../internal/listener/listener.go), after `serveHTTP` returns | nil, `errors.Is(http.ErrServerClosed)` and `errors.Is(net.ErrClosed)` are suppressed | Normal/closed Serve return; no new report |
| Same method, unexpected nonnil return | Calls `reportListenerTerminalError` once, before deferred `wg.Done` | Eligible `ListenerServeFailed`; unexpected failure remains eligible during concurrent Stop |
| `listenerServeError`, same file | Fixed Error text, but Unwrap retains exact raw cause | Trusted legacy identity, not an operator-safe report |
| [Listener Bootstrap](../../../internal/listener/bootstrap.go) | Injected callback is synchronous, must return promptly; panic isolated | Existing seam cannot contain an arbitrary blocking sink |
| [Handshake.ServeHTTP](../../../internal/handshake/handler.go), initial admission/context checks | Closed admission or unavailable/canceled Runtime context returns 503 before Authentication | Admission/cancellation decision, not Provider operational failure |
| [Authentication Service](../../../internal/authentication/service.go) | Ordered Providers, first success or first error; error returned unchanged | Handshake owns one overall request outcome; no duplicate per-Provider reporting |
| [API Key](../../../internal/authentication/apikey_provider.go) | Missing/wrong credentials reject with nil error; Resolver error returns wrapped cause | Credential rejection versus operational dependency failure |
| [JWT](../../../internal/authentication/jwt_provider.go) | Invalid token/signature/claims reject with nil error; key resolution error returns error | Same distinction; no token, claim or key detail enters reporting |
| Handshake Authentication error branch | Calls `rejectCanceledHandshake`; generic 503, no Provider error report | Coverage gap for an active-context Provider failure |
| Handshake false/invalid successful result | Existing 401; no Upgrade or Session | Credential/decision rejection, no new operational report |
| `rejectCanceledHandshake` | Only configured timeout cause emits legacy timeout callback; other cancellation has no callback | Cancellation/timeout is distinct; new Provider event does not replace timeout callback |
| Accept and handoff paths | Accept error returns; handoff error may call callback; failed handoff cleans owned connection | Deferred traffic/Dispatcher coverage; ownership unchanged |
| [Host.NewHost](../../../internal/runtime/host.go) | Default reporter nil; private composition threads reporter | Default construction is not an active safe reporting sink |
| [Bootstrap bindings](../../../internal/runtime/bootstrap.go) | `DependencyBindings.TerminalErrorReporter` permits explicit legacy injection | Existing injection exists; raw callback is not a new safe sink adapter |
| [Runtime composition](../../../internal/runtime/composition.go) | Same callback reaches Listener/Handshake; `terminalObserver.Observe` is empty | Session terminal observation remains outside this slice |

The legacy callbacks catch panic but can block. A blocked Listener reporter delays
serve return/Stop wait; a blocked Handshake reporter delays handler cleanup. A new
safe mailbox does not retroactively make arbitrary legacy callbacks nonblocking.

## 4. Eligibility and cancellation precedence

Listener preserves its existing closed-error filter exactly. For an eligible
unexpected Serve return, its owner makes at most one safe offer after return and
before the serve goroutine exits. Stop beginning does not reclassify an unexpected
return as normal shutdown. No liveness or supervision inference follows.

Immediately after `Authenticate` returns, Handshake takes exactly one `ctx.Err()`
sample for diagnostic eligibility, before its existing rejection path:

| Sample / result | Proposed new report | Existing outcome retained |
| --- | --- | --- |
| sampled context error nonnil, any returned result/error | None: cancellation/timeout takes precedence | Existing cancellation/503 path and separate legacy timeout behavior |
| sampled context active, returned error nonnil | One `AuthenticationProviderFailed` offer | Existing generic 503, no Upgrade/Session |
| sampled context active, nil error, rejected/invalid successful result | None | Existing 401, subject to existing later cancellation checks |
| sampled context active, nil error, valid success | None | Existing admission/cancellation/Upgrade/handoff checks |

A Provider returning `context.Canceled` or `DeadlineExceeded` while the boundary
context is active is an operational failure of that call; a bare error sentinel
does not prove caller cancellation. No Provider cause traversal or message parsing
is needed: only nil/non-nil error and the context sample determine eligibility.
Cancellation racing after the sample does not retract an eligible fact. The
sample does not replace any existing lifecycle/admission check or alter HTTP rules.

Cancellation precedence is deliberately conservative: an independent Provider
failure returned after boundary cancellation is omitted, not proven absent or
reclassified as successful. Complete late-failure causality is deferred. This is
lossy diagnostics, not an exhaustive audit of every Provider failure.

## 5. Safe report and bounded identity

Each copied report contains exactly three scalar fields:

| Field | Allowed values / meaning |
| --- | --- |
| SchemaVersion | Fixed `1` |
| Code | Closed enum: `ListenerServeFailed`, `AuthenticationProviderFailed` |
| Sequence | Positive uint64, unique only within one mailbox lifetime; accepted insertion order |

Severity, if rendered by a consumer, is the constant `Error` derived from Code,
not an extra free-form field. Code is safe semantic identity. It is not a raw Go
error and exposes no `Unwrap`, `errors.Is` chain or cause reference. Original
errors remain unchanged at existing trusted owner boundaries; projection neither
formats nor stores them. Existing Listener closed filtering remains owner-local.

Records contain no strings, maps, slices, interfaces, pointers, error/context,
request/response, Principal or capability. Provider name/type/index, route/session/
Workspace IDs, Snapshot/provenance, timestamps, host/address/PID and user-supplied
labels are excluded. No arbitrary labels or key/value bags exist. Code has exactly
two values; Sequence must never become a metric label. Unknown schema/code or an
invalid record is rejected without raw fallback text.

## 6. Mailbox and nonwaiting operations

One inert internal mailbox belongs to one explicit composition lifetime. It holds
a fixed ring of **64** safe report values and bounded scalar state. Capacity is a
fixed v1 implementation bound, not a new user Configuration setting. A producer
capability can offer only the two closed codes; a reader can remove safe copies.
Neither grants Runtime lifecycle access or cross-mailbox lookup.

An unbound/nil producer is disabled and performs no offer; an unbound/nil reader
returns no item. Mailbox state is never copied after use; capabilities refer only
to their original mailbox. Initial sequence state is zero, so first insertion is 1.

One `sync.Mutex.TryLock` attempt protects each offer/read. Busy means immediate
drop/no-item; there is no lock wait, spin, retry or I/O. Under the lock, only one
fixed-size record and scalar indexes/count/sequence are handled. No callback,
formatting, raw error operation, allocation, external code or blocking operation
runs there. Empty read returns no item; full offer drops the new record and keeps
the existing FIFO contents. Invalid code drops without consuming sequence.

Sequence is assigned only on accepted insertion under the same lock. After
`MaxUint64`, offers fail closed as exhausted; no wrap, reset or reuse. A removed
slot is cleared. A successful read returns one value and does not retain caller
memory. Capacity 64 bounds retained reports; zero worker goroutines, timers or
retry queues are created. Nonwaiting describes algorithmic dependency, not a
hard real-time scheduling guarantee. Contention may cause starvation/loss.

## 7. Ownership, timing, ordering and deduplication

The source owner decides eligibility and attempts one offer per source occurrence.
No Provider, Authentication Service, Listener callback adapter and Handshake layer
may each emit another copy of the same occurrence. Listener scope is one Serve
invocation; Handshake scope is one `Authenticate` return for one request. Source
integration must structurally guarantee this; the mailbox does not deduplicate.

Accepted insertion is the publication point for safe observation. FIFO removal
and Sequence order accepted insertions only, not wall-clock failure chronology,
HTTP response delivery or domain revisions. Concurrent occurrences have no further
global ordering. Two failures with the same raw error remain distinct occurrences.

Multiple readers destructively partition records. This is neither fanout nor
durable replay. A consumer may use `(exact mailbox capability, Sequence)` for
local duplicate recognition only; no globally comparable ID is exposed. Busy,
full, invalid or exhausted offers are loss, not permission to retry or replay
Authentication/Serve. No drop counters, reliable delivery, at-least-once or
end-to-end exactly-once guarantee is introduced.

## 8. Sink behavior and trust boundary

Runtime producers only offer safe values into the concrete mailbox. They never
invoke an arbitrary sink, logger, callback, serializer, channel consumer or
external delivery interface. Consumer code first removes a copied value, releases
the mailbox lock, then may invoke its own sink. No raw cause crosses this boundary.

A slow, blocking, failing or panicking sink affects that consumer's caller only.
Without reads, at most 64 records accumulate, then new offers drop. A consumed
record is not reinserted automatically after failure. Sink retry, recursion,
recursive diagnostic failure reports and delivery workers are not provided.
An external consumer owns its own execution and resources; this design creates
none and promises no termination of arbitrary sink code. Future sink integration
requires its own lifecycle/authorization/delivery design before production wiring.

## 9. Cancellation, Stop and Session compatibility

Mailbox construction opens no socket and starts no goroutine. Failed startup may
discard its mailbox and queued records. Host Stop does not close, flush, drain or
join the mailbox; no new Stop hook, observer wait or shutdown error is added.
An explicitly retained reader may drain after Stop. A source return after Stop
begins may still offer its eligible historical failure; this proves neither
current liveness nor completion of Host/Listener/Session cleanup.

The two live Host capabilities from ADR-0004 remain unchanged; any future mailbox
dependency is separately explicit and narrowly observation-only. No sink runs
under Host, Manager, Session or execution-owner locks. The synchronous Approved
DP-004 Observer is not repurposed: its blocking/lease accounting remains exactly
as specified. Session terminal observation and shutdown diagnostics are deferred.
Existing legacy callbacks remain trusted prompt-return interfaces; consumers that
block those callbacks are outside the new producer-isolation guarantee.

## 10. Relationship with DP-018

[DP-018](DP-018-runtime-operational-error-reporting-redaction.md) is Approved / Planned
for management, activation and recovery facts. Its scoped authorization,
owner/phase precedence, durable fact correlation and projection-version replay
keys are not automatically traffic semantics. A failed HTTP Serve is not evidence
of `ExecutionLost`, `StartupFailed` or `ShutdownFailed`; a Provider dependency
failure is not management `AuthorizationUnavailable`.

This design shares only compatible principles: preserve owner truth, construct by
allowlist, avoid raw causes/credentials and prevent delivery from mutating domain
state. It uses two separate traffic/runtime codes and mailbox-local occurrence
ordering, not DP-018 report keys or categories. No DP-018 implementation/status
change or authorization bypass is implied. The reader capability is a trusted
internal dependency, not a public management/API authorization model.

## 11. Verification Matrix for future implementation

These are requirements, **not executed tests or implementation evidence**.

| ID | Proof scenario | Required result |
| --- | --- | --- |
| V01 | Listener nil and wrapped closed sentinels | No new report; existing filter preserved |
| V02 | Unexpected Serve return, including concurrent Stop | One offer per invocation; no lifecycle/readiness mutation |
| V03 | Active context with arbitrary Provider error or context-shaped error | One Provider code, generic 503, zero Upgrade/handoff |
| V04 | Missing/invalid credentials and invalid successful result | No operational report; existing rejection behavior |
| V05 | Cancellation/timeout before sample, late result/error; cancellation after sample | Exact section 4 precedence, no retraction or duplicate timeout Provider code |
| V06 | Error whose Error/Unwrap/Is panics or blocks; hostile request/ProviderName | Provider projection calls none of them; records retain no sensitive input; Listener preserves only its pre-existing owner filter |
| V07 | Empty, capacity 64, full and forced contention | Immediate explicit loss/no-item; FIFO existing contents, no retry/spin |
| V08 | Concurrent producers/readers, reader mutation, independent mailboxes | Race-free copies, unique positive insertion sequence, no cross-mailbox state |
| V09 | Sequence at MaxUint64 and next offer | Last value allowed, later offers drop; no wrap/reset/reuse |
| V10 | Consumer stops reading, blocks, panics or errors after removal | Producer progress/cleanup independent; <=64 retained; no worker/reinsert/recursion |
| V11 | Construct/discard, failed startup, Stop with pending/late records | No goroutine/I/O/flush/join/new lifecycle result; post-Stop read allowed |
| V12 | Legacy callbacks and Approved Session Observer regressions | Existing owner/cause/accounting contracts unchanged; no new safe sink attached there |
| V13 | Unknown enum/schema and bounds inspection | Reject closed-model violation; no raw fallback or arbitrary labels |
| V14 | Architecture/source/import/scope checks | No service locator, Control Plane dependency, production wiring or Approved/Frozen amendment |

Applicable future verification includes unit/concurrency tests, `go test -race`,
stress and module/build/vet checks for the exact authorized slice. V01–V06 and
Host integration portions of V11–V12 belong to later source integration, not
proof supplied by an isolated mailbox test. No unavailable test may be called PASS.

## 12. Minimal next implementation slice

Recommend **one isolated `internal/runtimediagnostics` mailbox package and its
tests**, no production imports/call sites or source wiring. Readiness is conditional
on fresh exact-baseline intake and separate Owner authorization adopting this
Draft design as that bounded task's requirements; acceptance here is not permission.

Acceptance criteria:

1. Three-scalar safe report and exactly two codes, no raw-input constructor or
   sink callback; unnecessary exported surface remains private.
2. Fixed 64-slot ring; one try-lock, no allocation/external work in the critical
   section, no waiting/retries, slot clearing and truthful loss/no-item outcomes.
3. Accepted-insertion FIFO sequence and saturation behavior from sections 6–7.
4. Zero worker/timer/I/O/lifecycle hook; copied reader results; separate mailboxes
   isolated; sink can run only after a consumer removes a value.
5. Prove V07–V10, V13–V14 and package-local V11/V12 by focused tests, race/stress,
   no production/test modification outside this package and normal task documents.

This is a prerequisite, not operational reporting delivery. Listener and Handshake
binding are later separately authorized work; their future tasks must prove source
classification and preserve legacy callbacks without claiming sink safety for them.

## 13. Deferred coverage and non-goals

Explicit gaps: Dispatcher; Handler; Session terminal anomalies (including late
Observer/lease anomalies); shutdown diagnostics; Metrics integration. Also deferred:
configured Handshake/TLS timeout reporting, Accept/TLS peer rejection taxonomy,
complete late Provider failure causality, production composition/wiring, sink
adapters and consumer lifecycle, public API, access policy, durable audit/replay,
sampling/drop counters, supervision, restart and automatic remediation.

No production code/test/PoC, TASK-102 build/run/change, Architecture B activation,
Approved/Frozen amendment, Secret backend, new Configuration or deployment feature
is included. Gate 9 closure is unchanged and does not imply production readiness.

## 14. Existing evidence and design readiness

Inspected test source (not rerun in this design task):

- [Listener tests](../../../internal/listener/serve_error_test.go):
  `TestListenerExpectedServeShutdownIsNotReported`,
  `TestListenerExpectedClosedServeErrorsAreNotReported`,
  `TestListenerUnexpectedServeErrorIsReportedOnce`,
  `TestListenerConcurrentStopAndServeFailureReportsOnce`,
  `TestListenerServeErrorReporterPanicDoesNotChangeLifecycle`.
- [Handshake tests](../../../internal/handshake/handler_test.go):
  `TestHandlerAuthenticationRejectSkipsUpgradeAndSession`,
  `TestHandlerAuthenticationErrorSkipsUpgradeAndSession`,
  `TestHandlerConfiguredDeadlinePreventsUpgrade`,
  `TestHandlerReporterPanicDoesNotChangeHandoffCleanup`.
- [Runtime tests](../../../internal/runtime/bootstrap_test.go):
  `TestBootstrapHostAuthenticationErrorPreventsUpgrade`,
  `TestBootstrapTerminalObserverConsumesSessionError`.

These prove the existence of tests for current boundaries, not execution success
or new mailbox behavior. Independent actual-byte design gates and exact candidate
identity resolve from the TASK-109 envelope. Proposed mailbox implementation is
bounded and independently testable; full production diagnostics remains not ready
for activation within this task. STOP before implementation authorization.

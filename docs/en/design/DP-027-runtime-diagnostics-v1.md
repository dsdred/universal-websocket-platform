# DP-027: Runtime Diagnostics v1

[Russian version](../../ru/design/DP-027-runtime-diagnostics-v1.md)

## 1. Status and scope

- **Design Status:** Draft.
- **Implementation Status:** Partial — mailbox (TASK-110), component hooks (TASK-112), optional Host/Bootstrap wiring (TASK-113); production sink/delivery Planned.
- **Target Milestone:** Beta, Gate 11; this document does not close Gate 11.
- **Task:** [TASK-109](../../tasks/TASK-109-RUNTIME-DIAGNOSTICS-V1-DESIGN.md).
- **Source baseline:** `67ca1b200d42fda5be8b6dc8741d8f0bc24bd2a6`.

The original TASK-109 deliverable was design/readiness evidence, without normative
Approval. TASK-110 published the isolated `internal/runtimediagnostics` mailbox.
[TASK-112](../../tasks/TASK-112-COMPONENT-LOCAL-DIAGNOSTICS.md) implements the two
component-local source hooks and copied Producer construction under separate Owner
authorization: unexpected Listener Serve return and Handshake Provider operational
failure. The package schema/fixed64 loss contract is unchanged.

[TASK-113](../../tasks/TASK-113-RUNTIME-DIAGNOSTICS-WIRING.md) implements the
optional copied DP-009 Producer binding and private Host/Bootstrap passage into
both existing component constructors. Zero remains disabled. One external mailbox
pair belongs to one Host lifetime; Reader stays external through rollback/retry/Stop.
No production sink, transport, Metrics or durable delivery exists. Design remains
Draft/Partial; Gate11 OPEN. TASK-112 supplies I01–I12; TASK-113 supplies W01–W05
wiring proofs and their exact limitations in its evidence envelope. I06 narrow
offer/timeout-callback coexistence remains compositional source-order proof.

Integration design: [TASK-111](../../tasks/TASK-111-RUNTIME-DIAGNOSTICS-INTEGRATION-DESIGN.md),
baseline `060cf22cf040a2e84d49c52095af9ed4ab9390a7`. Section 15 retains that historical
source inventory; sections 16–19 separate implemented component APIs from later wiring.

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

## 12. Isolated substrate prerequisite (TASK-110)

Historical prerequisite, now published as TASK-110; the current next slice is section 19.

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

TASK-109 verified the original design only. TASK-110 subsequently implemented and
published the isolated mailbox through PR110 at this integration baseline; its
PR/post-merge Linux CI passed race/tidy/module guards. These historical facts do
not prove new integration. TASK-111 verifies only the design in sections 15–19;
at that design baseline source integration, consumer/delivery and Gate11 closure were absent. Current TASK-112 component status is stated in section 1.

## 15. Source integration inventory at TASK-111 baseline

This is Design-update / Readiness only at `060cf22cf040a2e84d49c52095af9ed4ab9390a7`.
No source or test changes are made. The mailbox is published; the integrations below
are Planned. Source anchors refer to this exact baseline.

| Source boundary | Existing fact | Exact planned producer point |
| --- | --- | --- |
| [Listener](../../../internal/listener/listener.go), `serve`, lines 212–218 | One Serve call, closed-error filter at 215, legacy callback at 216, deferred wg.Done at 213 | Inside the unchanged eligible branch, one `Offer(ListenerServeFailed)` before the legacy callback |
| Same file, wrapper/reporter 220–242 and Stop 185–186 | Unwrap retains raw cause; synchronous panic-isolated callback precedes Serve completion; Stop waits for Serve/handlers | Preserve filter, wrapper, callback and waits; do not pass a cause to Offer |
| [Listener Bootstrap](../../../internal/listener/bootstrap.go), 16–97 | Constructors bind dependencies; Build copies them without opening a socket | Copy one concrete Producer value before Start; no event during Build |
| [Handshake](../../../internal/handshake/handler.go), 107–115 | One Authenticate; error goes to existing generic rejection; later context check remains separate | Immediately after Authenticate, take one diagnostic-only `handshakeContext.Err()` sample; nonnil error plus active sample permits one Provider offer |
| Same file, 116–163 | Decision, final admission/context, Accept, handoff/cleanup and configured-timeout callback already exist | Preserve all existing response, cancellation and ownership operations |
| [Authentication Service](../../../internal/authentication/service.go), 32–46 | Ordered Providers; first error is returned unchanged; first success stops evaluation | No reporting in Service/Providers; Handshake owns the one opportunity per Authenticate return |
| [API Key](../../../internal/authentication/apikey_provider.go) / [JWT](../../../internal/authentication/jwt_provider.go) | Credential rejection returns nil error; resolver/key dependency failure returns operational error | Nil-error rejection emits nothing; active-context nonnil error is eligible without cause inspection |
| [Runtime composition](../../../internal/runtime/composition.go), 153–181 | Host builds both boundaries with one legacy callback; terminalObserver at 34 is a no-op | Later pass Producer separately from the two ADR-0004 capabilities; no Session Observer reuse |

Listen/build/TLS preparation errors, expected closed Serve returns, TLS peer
rejections, Stop errors, Accept/handoff failures and Session outcomes add neither
code. DP-018 management categories and report keys are not reused.

## 16. Exact value dependency and construction design

TASK-113 implementation note: the complete value passage and lifetime requirements
below are now implemented under separate Owner authorization. Prospective wording
and “Later wiring” W-row labels retain the TASK-111 design chronology; they no
longer describe absent Host wiring. The original TASK-112 scope remains historical.
No source hook, delivery capability or normative status is promoted here.

The following component APIs are implemented by TASK-112; later Host passage remains Planned:

- `DefaultBootstrap.WithDiagnosticsProducer(producer runtimediagnostics.Producer) DefaultBootstrap` returns a copy of the Listener Bootstrap with a private Producer value. Existing constructors and the Build interface stay unchanged; Build copies the value to DefaultListener without offering or new validation.
- `handshake.NewHandlerWithDiagnostics(admission AdmissionCapability, runtimeContext RuntimeContextProvider, authenticationService authentication.Service, handoff connection.AuthenticatedDispatcher, reportError func(error), producer runtimediagnostics.Producer) (*Handler, error)` stores a private Producer value. Existing constructors keep signatures and required-dependency validation order, delegating with zero Producer.

Zero disables observation; no fallback mailbox is constructed. A caller dereferences
the nonnil Producer pointer returned by mailbox New once at its construction
boundary. Only the opaque capability is copied, never mailbox/mutex state. Later
assignment of the caller's Producer object to zero or another capability cannot
redirect an already-constructed component. Concurrent caller mutation during the
construction copy remains unsupported caller misuse. No runtime setter, diagnostic
interface/callback or raw-error constructor is added.

Complete later passage is explicit and **Planned**, not part of the next slice:

1. External composition owner creates one mailbox pair for one Host construction/lifetime and retains Reader externally. Separate Host lifetimes require separate mailboxes; there is no implicit global sharing or Host identity field.
2. Future fixed `DependencyBindings.DiagnosticsProducer runtimediagnostics.Producer` is optional, zero absent. Bootstrap snapshots this value once; it never creates/owns the mailbox or calls Offer/TryRead. Caller must not mutate the borrowed request/dependencies while Bootstrap reads them.
3. A private Host-construction core accepts/stores the value and threads it through private production composition into the two component constructors. Public Host interface, NewHost signature and existing zero-producer behavior remain unchanged.
4. Host remains the only operational composition root during Start. No Reader/getter/consumer enters Host, Bootstrap outcomes, Snapshot, Configuration, launch identity or the ADR-0004 admission/context carrier. The new dependency imports only the leaf mailbox package.

[Draft DP-009](DP-009-runtime-bootstrap-contract.md) section 11 explicitly separates
this planned fourth binding from its three currently implemented bindings. Static
validation/failure precedence, fixed factory ownership, one Host/Build/Start call,
cause wrapping and cleanup ownership remain unchanged. No Approved/Frozen amendment
is needed; an implementation requiring one must stop for Owner decision.

## 17. Exact source algorithm and compatibility

Listener adds exactly one concrete Offer call inside the existing nonnil/unexpected
Serve-error branch, before `reportListenerTerminalError`. Ignore the bool. Keep the
existing errors.Is filter and raw-cause wrapper intact. Concurrent Stop does not
reclassify an unexpected return. A blocked legacy callback cannot prevent the prior
offer attempt, but still delays wg.Done/Stop. The pre-existing filter can itself
invoke hostile Is/Unwrap and block or panic; no stronger guarantee is introduced.

Handshake adds exactly one diagnostic-only context sample immediately after the
Authenticate return. If returned error is nonnil and that sample is nil, attempt
one `Offer(AuthenticationProviderFailed)`, ignoring bool. Then execute the existing
error branch, later context checks, decision, admission, Accept and handoff logic
unchanged. The sample never replaces those checks.

| Diagnostic sample / Authenticate return | New Provider opportunity | Existing flow retained |
| --- | --- | --- |
| Cancelled/timed out sample; any result/error | None, including independent late dependency failure | Existing cancellation/error rejection and legacy configured-timeout handling |
| Active sample; nonnil error, including direct/wrapped context sentinels | One offer; do not call Error/Unwrap/Is | Existing generic 503; no Upgrade/handoff |
| Active sample; nil error and rejected/invalid success | None | Existing 401 unless subsequent existing cancellation checks win |
| Active sample; nil error and valid success | None | All existing admission/context/Upgrade/handoff checks |

Cancellation after an active sample does not retract an eligible offer. If a
configured timeout becomes visible to the existing rejection helper afterward,
its legacy timeout callback may still run; that is not a duplicate Provider code.
Cancellation before the sample deliberately loses late failures, not proves their
absence or success. Provider error sentinels do not establish caller cancellation.

Neither source passes raw errors, request/Provider metadata, credentials, Principal
or labels to the mailbox. It receives only a fixed Code and assigns version and
accepted-insertion sequence. Service, Providers, callback adapters and generic
legacy report helpers emit no duplicate offers. One occurrence has at most one
attempt, not a reliable-delivery/exactly-once acceptance guarantee. Loss never
causes retry, replay, recursion, fallback reporting, response or lifecycle changes.

## 18. Lifetime, startup and Stop

The later mailbox lifetime is one Host lifetime, including failed Start/rollback
and retry on that same Host. Do not reset, replace, drain or flush it on retry;
accepted historical events retain their sequence. A failed lifetime may discard
its entire mailbox. A retained Reader may consume after rollback or Stop; no event
proves attempt identity, readiness, startup success, current liveness or cleanup.

Static binding/Build, failed Listen or TLS preparation emit neither code. If an
actual Serve invocation fails before startup commit or during rollback/Stop, its
normal source filter still decides eligibility. Pre-commit unavailable Runtime
context/admission still prevents Authentication and hence Provider reporting.

No new resource, worker, sink, timer, Start/Stop/Close/reset/flush/drain/join or
shutdown error is added. Source offers acquire no Host/Listener lifecycle,
Manager, Session or execution-owner lock; mailbox TryLock ends before legacy
callbacks or further handler work. A blocked external consumer affects only that
consumer. Existing arbitrary legacy callbacks/Providers can still block cleanup;
the design supplies no new hard termination or scheduling-latency guarantee.
Approved Session Observer/lease/accounting and Host startup/rollback/Stop order,
readiness, admission and errors.Is-compatible cause chains remain unchanged.

## 19. Exact future verification and minimum next slice

These are required future proofs, **not tests executed by TASK-111**. The component
slice uses two focused diagnostic test files in Listener/Handshake; existing tests
remain regression coverage. W rows belong only to a later separately selected
wiring slice. I rows refine the source portions of V01–V14; existing mailbox proofs
remain substrate evidence, not a replacement for source proofs.

| ID / proposed proof | Exact required outcome | Slice |
| --- | --- | --- |
| I01 TestListenerDiagnosticsExpectedServeReturns | nil, direct and wrapped http.ErrServerClosed/net.ErrClosed: no new event; legacy behavior unchanged; wrapped cases extend the current direct-sentinel coverage | Component |
| I02 TestListenerDiagnosticsUnexpectedServeOnce | One unexpected return, including concurrent Stop: one opportunity/code, safe schema/positive sequence; legacy callback once and errors.Is(callbackErr,rawCause) true | Component |
| I03 TestListenerDiagnosticsBeforeLegacyReporter | Block controllable legacy reporter: prior event is removable; release reporter before asserting cleanup. Reporter panic remains isolated; no claim that blocked legacy permits Stop completion | Component |
| I04 TestHandshakeDiagnosticsOperationalFailure | Active sample + arbitrary/direct/wrapped context-shaped error: one Provider code, generic 503, zero Upgrade/handoff; ordered multi-Provider service still has one boundary opportunity | Component |
| I05 TestHandshakeDiagnosticsCredentialRejection | Nil-error credential reject/invalid success: no code, existing 401/cancellation precedence; valid success also no code | Component |
| I06 TestHandshakeDiagnosticsCancellationPrecedence | Cancel/timeout before sample + late result/error: no Provider event, existing 503/timeout callback. Cancellation after active sample does not retract; legacy timeout callback may coexist. Deterministic phase-controlled fixtures, not timing sleeps | Component |
| I07 TestHandshakeDiagnosticsDoNotInspectProviderError | Hostile Error/Unwrap/Is and request/Provider metadata never inspected by projection; exactly three safe scalars. Exclude Listener pre-existing filter from hostile-method non-invocation guarantee | Component |
| I08 TestDiagnosticsSaturationPreservesSourceBehavior | Prefill real mailbox64; trigger each eligible source: preserve FIFO/HTTP/legacy behavior under full loss, no retry; zero Producer preserves legacy behavior | Component |
| I09 TestDiagnosticsCaptureProducerValue | Bind P1 value, overwrite caller object with P2/zero, trigger: only original Reader1 receives; separate mailbox instances do not cross | Component |
| I10 TestDiagnosticsConcurrentSourcesAndReaders | Valid copies, positive unique accepted sequence/per-reader FIFO, accepted count <= eligible occurrences, allowed loss/no source duplicates/cross-mailbox events; future race/stress and legacy regressions | Component |
| I11 TestDiagnosticsDoNotAddOutcomeKinds | Accept/handoff/Session/startup/Stop failures add neither code; existing handoff/reporter cleanup and errors.Is behavior unchanged | Component |
| I12 source/import/API audit + existing mailbox proofs | One concrete Offer per eligible branch, ignored bool; no raw input/adapter/counter/retry/sink/worker. Busy/exhaustion compatibility is compositional: mailbox white-box proofs plus source structure. Do not add mailbox test APIs or claim directly forced inaccessible state | Component |
| W01 TestBootstrapDiagnosticsBindingSnapshot | Optional value copied once; unchanged request/validation precedence; no Offer/TryRead during binding; caller reassignment cannot redirect Host | Later wiring |
| W02 TestHostDiagnosticsSharedProducerIsolation | One Host uses same mailbox for both sources; independent Hosts use separate pairs; Reader never stored/exposed by Host | Later wiring |
| W03 TestHostDiagnosticsRollbackRetryLifetime | Composition/startup/rollback/retry preserve lifecycle/causes and one sequence; old events retained, no reset/flush/new resource | Later wiring |
| W04 TestHostDiagnosticsStopIndependence | With nil/prompt legacy reporter, saturated mailbox or blocked external consumer does not change Host Stop order; retained reader/late eligible Listener event work; legacy blocking exclusions remain | Later wiring |
| W05 Host/Session contract audit + existing regressions | Host interface/NewHost unchanged; ADR004 carrier2; Observer/accounting unchanged; no ControlPlane cycle, sink/Metrics/API/durable transport/Configuration field | Later wiring |

Future regression includes the five Listener serve_error tests, Handshake rejection/
error/deadline/admission/handoff/panic/concurrency tests, Host rollback/retry/shutdown
and Bootstrap optional-binding/cause tests, and Approved Session/Manager tests.
Implementation verification requires formatting, full test/build/vet/module checks,
available race and stress; unavailable checks remain explicit limitations. Linux
race does not qualify Windows/native TASK-102.

**TASK-112 component slice:** the implemented scope is only component-local
hooks/construction: listener.go, listener/bootstrap.go, handshake/handler.go, two
focused package test files and necessary task/status docs. No runtime, Authentication,
Session/Manager, mailbox or module edits. Existing production composition keeps its
legacy constructors/zero Producer; this does not activate Host/Bootstrap wiring.

The later value-binding/Host/private composition slice requires this prerequisite
independently accepted, fresh readiness/Size Guard and separate Owner authorization.
Neither slice introduces a sink or consumer lifecycle. This design acceptance is
not normative DP Approval, implementation permission or Gate11 closure.
**STOP BEFORE COMMIT GATE AND IMPLEMENTATION AUTHORIZATION.**

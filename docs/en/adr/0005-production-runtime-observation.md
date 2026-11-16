# ADR 0005: Production Runtime Observation

[Русская версия](../../ru/adr/0005-production-runtime-observation.md)

## Status

**Design Status: Draft. Implementation Status: Not Implemented.**

Architect proposal A132-001, [TASK-132](../../tasks/TASK-132-PRODUCTION-RUNTIME-OBSERVATION-DESIGN.md).
Production implementation admission: **BLOCKED** pending the decisions below.
Design review does not approve this ADR, change existing contracts or activate
TASK-131, Stage B/B1/B2, Production C2/C3 or C2-Exposure. Gate 11 remains OPEN.

## Context and repository evidence

At baseline `88fb333453ab9563285d65cd44a8b0a4019f447a`,
[`Owner.Observe`](../../../internal/runtimelifecycle/owner.go) is a coherent
mutex-protected snapshot of one Instance's lifecycle; authorized
[`Directory.Observe`](../../../internal/runtimemanagement/directory.go) exists
in isolation. Actual states are stopped/starting/running/stopping/failed.
[`Control Service`](../../../cmd/control-service/main.go) creates configuration
services and HTTP/Studio routes, with no Owner/Directory/managed Runtime graph.
[`Publish`](../../../internal/configurationversion/service.go) updates version
records; [`health`](../../../internal/http/server.go) returns Control health.
Neither proves Runtime execution. Dashboard already displays Runtime Unavailable.

[`Host.Running/Ready`](../../../internal/runtime/host.go) read host.state.
[`Listener.serve`](../../../internal/listener/listener.go) reports unexpected
Serve errors but does not update Host or Owner lifecycle on Serve return.
[DP-010 section 27](../design/DP-010-runtime-lifecycle-owner-contract.md) forbids
inferring Running -> Failed by polling Running(). Process containment evidence
can establish only its exact declared generation facts, not Host readiness.
The task forensic table identifies symbols, baseline lines and inspected tests.

[ARCH-002](../architecture/ARCH-002-runtime-foundation-freeze.md) freezes
readiness, lifecycle ownership, startup/rollback and shutdown order. It leaves
focused supervision designs open. [ARCH-004](../architecture/ARCH-004-runtime-deployment-and-identity-model.md)
separates Instance, Attempt, version and process identity; initial topology is
in-process. [DP-013](../design/DP-013-runtime-management-routing.md) defers
production composition, HTTP and concrete policy. This Draft overrides none of
these sources or [ADR-0002](0002-configuration-dsl.md)/[ADR-0003](0003-runtime-architecture.md).

## Options comparison

| Dimension | A: existing supervisor observation | B: source-owned evidence | C: read-only validity adapter |
|---|---|---|---|
| Required changes | Separately admit real containment/evidence composition; no existing production supervisor | Add retained per-Host/Listener completion capability | Exact-target capture, identity, validity and transport boundary |
| Source of Truth | Existing exact-generation authority | Owner lifecycle; Host-owned component facts | Borrows Owner/B/A, owns no Runtime state |
| Failure detection | Generation termination only within declared guarantee | Underlying Serve return; not whole-process death | None independently |
| Concurrency | Exact-generation authority revalidation | One-shot retained fact, no lifecycle callbacks | Revision and attempt recheck; fail closed on overlap |
| Stale state | Durable binding is not fresh Host health | Completion remains observable without queue loss | Explicit expiry withdraws current presentation |
| Performance | Platform/authority I/O may need separate bounds | O(1) bounded local publication/read | Bounded local capture; no polling worker |
| API implications | Recovery evidence must not become a health API | Private read-only capability, no lifecycle commands | Versioned allowlisted DTO, no-store, concealed denial |
| Testability | Generation/unknown/platform qualification | Deterministic Serve/start/stop/reporter races | Identity, age, timeout, out-of-order, policy tests |
| Compatibility | DP-017/022 guarantees unchanged | No change to Ready, Owner state, cleanup or Stop | Existing Start/Stop/Observe signatures retained |
| Protected dependencies | Own production admission; TASK-102 not automatically required | No diagnostics C2/C3 or Stage B | Separate observation path need not use Stage B |
| Limitations | Current production supervisor absent; process alive is not Host readiness | Serve return is not complete Host shutdown or deadlock detection | Cannot manufacture missing source evidence |

A is a possible future supplementary generation source, not a ready supervisor.
C alone can display lifecycle facts with execution Unknown; it cannot solve
unexpected execution loss. Reliability requires source evidence and explicit
validity rather than choosing an inexpensive endpoint wrapper.

## Proposed decision: B plus C

Keep lifecycle authority in the existing Owner. Add a bounded retained source
fact and a read-only adapter. Do not add a second state repository, event-sourced
lifecycle registry, diagnostics consumer, background supervisor or automatic
terminalization. All requirements below describe a proposal, not current API.

### State and evidence semantics

| Value | Meaning and source |
|---|---|
| Running | Owner lifecycle actual running: successful startup was published. Not continuous liveness, traffic health or deadlock freedom. |
| Starting / Stopping | Existing Owner launch/shutdown work; do not infer progress or bounded completion. |
| Stopped | Existing authoritative Owner stopped outcome, with its resource-ownership semantics; not missing data or missing PID. |
| Failed | Existing Owner failure outcome; observation cannot create this transition from a component event. |
| Unknown | Request reached the observation boundary but evidence is insufficient, stale, mismatched or contradictory. Not a new Owner state. |
| Unavailable | Observation/source cannot be reached or is not configured/authorized; no assertion about target execution. Not Stopped. |

The result separates `lifecycleState`, `desiredState`, `executionEvidence` and
`validity`. `executionEvidence` is unknown or an allowlisted fact such as
`listenerServeReturned`, with explicit proof scope. Never emit `liveConfirmed`:
absence of a completion marker does not prove useful execution. Lifecycle
running can coexist with Serve-return evidence; presentation must expose the
contradiction in evidence level and withhold a live-success badge.

Process liveness, Host lifecycle readiness, configuration publication and
Control availability are separate. This initial design does not measure
application progress or continuous health. Even a future process-alive proof
would not prove Listener usability. Published != Running; Health != Runtime
liveness; Snapshot != current liveness; Unknown != Stopped.

### Identity, source ownership and restart

Target is exact `(workspaceID, configurationID, runtimeInstanceID)`.
An observation identifies active/no-active attempt explicitly; a last attempt
is historical and never substituted for active. Version belongs to its pinned
Attempt. Multiple Instances may execute one version. No latest-version inference,
implicit list/discovery, arbitrary PID/address probe or dummy stopped Owner.

Include `launchAttemptID` and any execution generation only when exact binding
exists. Generation is borrowed from sanctioned execution composition, never
allocated by observation. Existing DP-017/022 generation authority is distinct
from a Host/Attempt and is not health evidence. No bound generation means no
generation claim. A Control restart invalidates process-local observations;
recovery/reconciliation is never run by an observation request. Old responses
cannot bind to a new Attempt even when PID/version/address is reused.

Every observation, including no-active-attempt results, additionally requires
`observationSourceIncarnation`: a trusted composition-issued identity for one
source process lifetime, never reused across source replacement. It is an
observation provenance epoch, not Runtime identity, execution/containment
generation, recovery authority or proof of alive/ready. Observation consumes the
qualified identity; it does not mint an independent execution generation.
The issuer's uniqueness/nonreuse and authenticated current-source bootstrap
must be proven at production admission; PID, clock or an unqualified random
identifier assertion is insufficient. No issuer or unverifiable provenance
means Unknown with no current-success presentation.

Requests carry `expectedObservationSourceIncarnation` from authenticated current-source
bootstrap; bootstrap itself grants no successful Runtime-status claim.
Capture and every response bind the same identity. A mismatch gives
Unknown, never silent rebinding. Client binding is `(source incarnation, exact
target, attempt/no-attempt, request generation)`; new incarnation or bootstrap
failure clears old presentation before rebinding. Old-incarnation responses are
discarded even if Target/Attempt strings are identical. Unknown current-source
identity forbids a current claim. This establishes provenance at bootstrap and
capture, not a guarantee that the source survives until display.
Source binding is immutable within an adapter. Incarnation change invalidates
pending requests; a new bootstrap requires a new pinned request. Delayed bootstrap
responses cannot replace the current binding outside their request generation.

B retains one typed completion fact per Listener/Host lifetime. Publish once
immediately after **every** underlying Serve return (including nil/closed),
before diagnostics or arbitrary legacy reporters. Publication is bounded,
non-consuming and performs no callbacks, I/O, cancellation, Stop or Owner state
publication. Host binds the capability to its actual component graph. Missing
coverage yields Unknown; a blocked/panicking reporter cannot hide publication.
Serve completion does not prove handlers drained, Host cleanup or resources
released. Retention lasts with the same owned Host reference; no replay/history
store or cross-Attempt capability reuse.

### Capture and concurrency

1. Authenticate and authorize exact observation before target-aware external
   disclosure. This is a proposed transport gate: current Directory looks up
   its internal scope before calling Authorize, though Owner access is gated.
2. Pin the qualified source incarnation; capture Owner snapshot and proposed private monotonic observation revision
   under its mutex. Increment revision at existing relevant lifecycle/attempt
   publications, not in response to reads. Overflow fails closed.
3. Sample the exact bound Host evidence outside the Owner mutex. No nested
   cross-owner locks, external callbacks, network I/O or lifecycle waits.
4. Recheck source incarnation, Owner revision/Attempt and source revision. Any change, absent binding
   or ambiguous ordering gives Unknown(`ConcurrentChange`/`CoverageMissing` etc.)
   without retry/spin. A stable sample covers a bounded capture interval, not a
   global atomic process/Owner cut or a promise about future execution.
5. Cancellation, policy failure or timeout discards the sample; never reuse a
   previous successful result. Commands, state repair and shutdown remain absent.

### Proposed read-only boundary and errors

Before an endpoint is implemented, separately approve its exact route/schema,
target provisioning/discovery, real authentication/authorization and ingress
bounds. Existing trusted-local unauthenticated Studio does not automatically
authorize operational target disclosure. Minimal conceptual response fields:
schema version, mandatory source incarnation, exact target/attempt binding, optional bound generation,
lifecycle/desired state, evidence/proof scope, validity/closed reason,
`observedAt`, capture duration and maximum presentation age. No raw Host,
cancellation capability, Configuration payload, Secret, PID, path or raw error.

Proposed request budget is 2 seconds; at most 8 admitted requests globally and
1 per authenticated principal, with immediate refusal on excess. These logical
bounds do not bound pre-handler ingress or a hostile backend: network limits,
shutdown and prompt local callbacks require separate production proofs. No
unbounded goroutine wrapper around a blocking source. No retry, queue or worker.

Closed reasons: NotConfigured, Unavailable, Cancelled, Busy, ConcurrentChange,
ScopeMismatch, GenerationMismatch, SourceIncarnationMismatch, CoverageMissing, Stale. Unauthorized and
nonexistent targets have the same generic external failure; concrete HTTP
codes/DTO/resource limits remain production-admission blockers, not an endpoint
implemented by this ADR. No error returns an old successful sample.

### Time and presentation validity

`observedAt` is server UTC at capture completion; capture duration uses server
monotonic time. UTC is informational, never a generation/lease or authority.
Proposed maximum presentation age: 5000 ms, a UI policy rather than proof of
liveness. Do not compare browser/server wall clocks to establish freshness.
The client conservatively measures age from monotonic request start, including
the entire round trip; an over-age response cannot become current. Source
validation and response serialization must remain inside the request budget.

Each refresh/target selection has a client request generation; obsolete responses
are discarded. New request, errors, timeout, lost coverage, identity change,
tab resume and expiry withdraw current presentation. After expiry show Unknown
and optional explicitly historical last sample; manual Refresh is sufficient.
Client timers cannot guarantee behavior while a page is suspended: withdraw on
resume before redisplay. Even an unexpired response means only “Lifecycle
running — observed at …”, with execution proof level, never “alive now”. Exit
between sampling and display is inherently possible.

## Failure-mode analysis

| Scenario | Required truthful behavior |
|---|---|
| Unexpected Runtime exit | Whole in-process death loses the observer/API. Serve return records source completion, never automatic Failed/Stopped or cleanup success. |
| Supervisor unavailable | Optional authority missing: execution Unknown/Unavailable; independently available lifecycle fact stays separately labelled. |
| Control restart | New qualified source incarnation; same Target/Attempt strings cannot restore old binding. No snapshot hydration or implicit DP-017 recovery; old samples historical. |
| Runtime restart | New Attempt/Host, old completion/response discarded; no in-place Host restart introduced. |
| Stale snapshot | Age expiry withdraws current presentation regardless of UTC label or previous HTTP success. |
| Observation lost | Retained cell avoids lossy queue delivery; missing capability/coverage is Unknown, never default alive. |
| Exit between read and display | No absolute guarantee; identify sampled facts and age rather than promise current liveness. |
| Concurrent start/stop/exit | Existing lifecycle arbitration unchanged; mismatch gives Unknown, never second Stop or fabricated ordering. |
| PID reuse | PID is never identity/routing; exact Instance/Attempt and available generation binding required. |
| Dashboard retains Running | Generation guard, no-store and expiry; error/refresh/resume clears current claim; historical sample labelled separately. |

## Admission, Owner decisions and conditional roadmap

**BLOCKED for full production implementation.** Additive B+C can preserve
ARCH-002: no change to Ready == Host Running, startup commit, Gate/context,
rollback, shutdown, resource ownership or Owner lifecycle transitions. A focused
read-only capability is possible without a blanket frozen-foundation amendment.
However this Draft is not an approved implementation contract; actual production
graph, target provisioning and concrete security/transport admission are missing.

Owner decisions required: choose sampled lifecycle plus explicitly unknown
execution semantics or separately request stronger supervision; separately
approve the focused design; select real managed deployment/target graph and
authentication/authorization boundary. “Failed immediately after exit” or
“Running means healthy now” needs a separate lifecycle/progress design.
No universal prerequisite on TASK-102, Stage B or diagnostics C2/C3 is established.
This observation proposal does not satisfy or bypass Gate 11 diagnostics.

| Conditional slice (no Task ID or activation) | Acceptance criteria before downstream use |
|---|---|
| R1: design approval and production boundary decisions | Exact proof vocabulary, frozen compatibility, real graph/security owners named; no implicit implementation permission. |
| R2: isolated retained source capability | Deterministic all-Serve-return, pre/post-start, Stop overlap, blocked/panicking reporter and retention proofs; no lifecycle/cancellation side effects; full Go/race/stress/frozen regressions. |
| R3: isolated capture adapter | Exact tuple/revision/generation/source-incarnation checks, overflow, cancellation, capacity, redaction and missing evidence proofs; same Target/Attempt after source restart, delayed bootstrap/old response, missing issuer, wrong pin, duplicate/reused identity and source replacement during capture; no retry/worker/new state store. |
| R4: separate production composition/transport admission | Real Owner graph/grants, qualified nonreused source-incarnation issuer and authenticated bootstrap, approved route/DTO/ingress/shutdown limits, security/API/source-to-operator proofs; STOP without real graph or provenance. R2/R3 cannot substitute. |
| R5: Dashboard after R4 | Browser actual-contract/expiry/resume/error/old-response tests, publication/health separation and Studio regression; no controls. |

Each slice requires fresh intake, coverage report and Size Guard; split if >500
production lines, >15 paths or multiple independent contracts. Roadmap is a
conditional dependency plan, not Ready implementation tasks or a successor queue.

## Consequences

Authoritative lifecycle remains single-owned, source completion cannot be lost
through diagnostic consumption, and stale data cannot silently become a new
live claim. Costs are an extra focused capability, revision/recheck discipline
and explicit UI evidence vocabulary. This deliberately leaves progress/health,
whole-process supervision, durable recovery and automatic terminalization open.
Existing Approved/Active/Frozen sources and accepted ADRs remain unchanged.
Independent review and exact documentation manifest are recorded in TASK-132;
review acceptance of design evidence does not promote this Draft or activate work.

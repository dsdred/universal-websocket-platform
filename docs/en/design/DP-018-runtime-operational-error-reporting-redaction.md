# DP-018: Runtime Operational Error Reporting and Redaction

[Russian version](../../ru/design/DP-018-runtime-operational-error-reporting-redaction.md)

## 1. Status

- **Design Status:** Draft
- **Implementation Status:** Planned
- **Target Milestone:** Beta — Complete the Single-Node Runtime
- **Scope:** candidate ARCH-004 section 19(6) contract for safe operational
  failure projection across Runtime management, activation, and recovery

This document is non-normative until separately approved. It proposes a
focused contract and does not authorize management implementation or
Production Activation.

## 2. Purpose

Runtime ownership contracts already preserve exact internal failures and
durable lifecycle and command outcomes. Operators also need a stable, safe way
to understand those failures without receiving Secrets, payloads, raw error
text, process-local authority, or cross-scope facts.

This design defines the boundary between authoritative domain truth and an
operator-safe operational report. It preserves failure ownership, makes
redaction fail closed, keeps replay deterministic, and leaves logs, metrics,
traces, audit storage, alerting, and public transports as replaceable adapters.

## 3. Authority

This Draft is subordinate to:

- [ARCH-001](../architecture/ARCH-001-runtime-architectural-pattern.md), which
  distinguishes valid decisions from operational errors and forbids discarded
  lifecycle failures;
- [ARCH-002](../architecture/ARCH-002-runtime-foundation-freeze.md), which
  preserves startup and rollback cause identity;
- [ARCH-004](../architecture/ARCH-004-runtime-deployment-and-identity-model.md),
  especially section 19(6);
- [ARCH-005](../architecture/ARCH-005-runtime-configuration-snapshot-and-loading-model.md),
  which excludes Secret values from Snapshot and provenance;
- [DP-013](DP-013-runtime-management-routing.md), [DP-014](DP-014-runtime-operational-identity-persistence.md),
  [DP-015](DP-015-runtime-management-command-idempotency.md),
  [DP-016](DP-016-runtime-activation-replacement-rollback.md), and
  [DP-017](DP-017-runtime-recovery-reconciliation.md).

If this Draft conflicts with those sources, the higher source wins.

## 4. Scope

This design defines:

- failure ownership and the reporting projection boundary;
- stable operator-safe classification and severity;
- exact scoped correlation and anti-enumeration behavior;
- allowlist redaction and forbidden content;
- report creation, publication, replay, and delivery-failure semantics;
- cancellation, concurrency, and recovery interaction;
- acceptance proofs and implementation gates.

It does not define an HTTP error schema, logging library, metric names, trace
format, audit database, retention period, alert policy, UI wording, localization
system, or concrete authorization policy.

## 5. Terms

**Authoritative fact** is an immutable domain or lifecycle fact owned by its
existing boundary: validation result, Owner outcome, durable command outcome,
attempt state, recovery classification, or conditional publication result.

**Internal cause** is the original component-owned error identity and cause
chain. It remains available only inside the trusted execution boundary that
already owns it.

**Operational report** is an immutable, operator-safe projection of one
authoritative fact. It is evidence for observation, not lifecycle or command
truth.

**Report key** is a stable opaque correlation tuple inside one authorized
scope. It identifies the source fact and projection version without exposing
payload or process-local authority.

**Category** is a stable semantic class safe for operator decisions. It is not
a global replacement for component errors.

**Public detail** is a bounded allowlisted value whose type and disclosure are
defined by this contract. Arbitrary strings are not public detail.

**Delivery adapter** exports an already-safe report to a log, metric, trace,
audit sink, alert path, or future API mapper. It receives no raw cause.

## 6. Responsibility Boundary

The component that owns a failure keeps its exact error identity, wrapping,
joining, and cleanup behavior. The reporting boundary does not normalize
errors before their owner makes the authoritative decision.

Only a trusted projector may combine an authoritative fact, its exact scoped
correlation, and the internal cause long enough to select an allowlisted
category and detail set. The projector emits a safe value and discards any
reference to the cause. Delivery adapters receive only that safe value.

The reporting boundary never:

- changes desired or actual state;
- terminalizes or reopens a command;
- claims lifecycle, recovery, or delivery work;
- converts an unknown outcome into success or failure;
- returns report delivery as the component's domain outcome;
- becomes a universal error or policy engine.

## 7. Report Model

Each report contains only:

- projection schema version;
- opaque report key;
- authorized operational domain and opaque Workspace, Configuration, Runtime
  Instance, command, attempt, and generation correlations when applicable;
- operation and phase from a closed vocabulary;
- category, severity, and resolution state from closed vocabularies;
- bounded allowlisted public details;
- authoritative fact revision or immutable outcome identity when available;
- no delivery-attempt or first-versus-replay marker in semantic content.

Absence is explicit. A report does not use an empty string to mean hidden,
unknown, not applicable, and absent simultaneously.

Time, host name, PID, address, stack, source file, and goroutine are not core
report fields. A delivery adapter may add its own envelope time or sink
metadata, but that metadata is not authoritative Runtime truth and cannot
affect correlation or replay.

## 8. Stable Categories

The initial closed category set is:

- `AuthorizationUnavailable` — authorization evaluation failed operationally,
  with no policy reason or target-existence disclosure;
- `SourceUnavailable` — an authorized required source failed operationally;
- `PreparationFailed` — Loader, Builder, binding, or pre-Host preparation
  failed;
- `StartupFailed` — Host startup or readiness failed after Host ownership
  began;
- `ShutdownFailed` — exact shutdown completion was not achieved or proved;
- `ExecutionLost` — a bound execution terminated or became unavailable without
  graceful completion proof;
- `RecoveryBlocked` — recovery evidence is missing, contradictory, stale, or
  indeterminate and admission remains closed;
- `PersistenceUnavailable` — an authoritative persistence operation failed
  definitively before its effect;
- `OutcomeIndeterminate` — an authoritative mutation may have committed and
  exact inspection is required;
- `InternalFailure` — a trusted owner produced a failure not safely classified
  by a more specific rule.

Authorization denial, malformed input, not-found, immutable-intent or revision
conflict, other valid negative decisions, and satisfied/idempotent outcomes
are not error reports. They may be represented by separate operational events
in a future design, but must not be mislabeled as failures.

## 9. Severity and Resolution

Severity is derived from the authoritative fact, never from raw error text:

- `Info` is not used for an operational error;
- `Warning` means the requested operation did not complete but authoritative
  state remains coherent and no safety barrier is closed;
- `Error` means an operation or execution failed and operator attention may be
  required;
- `Critical` means truth is unresolved, a safety barrier remains closed, or
  contradiction prevents safe progress.

Resolution is one of `Terminal`, `RetryableAfterChange`, `InspectionRequired`,
or `Blocked`. `RetryableAfterChange` never means automatic retry permission.
Category alone does not promise retryability.

## 10. Classification Rules

Classification uses the authoritative boundary and phase, stable error
identity where defined, and proven durable state. It never parses raw error
messages.

Primary category is selected by authoritative owner and phase in this exact
precedence:

1. authorization-evaluator operational failure -> `AuthorizationUnavailable`;
2. any aggregate, command, attempt, phase, binding, or recovery publication
   that may have committed -> `OutcomeIndeterminate`;
3. recovery assessment/evidence that is missing, stale, contradictory, or
   unavailable without an indeterminate mutation -> `RecoveryBlocked`;
4. a definite no-effect aggregate/command/recovery-store operation failure ->
   `PersistenceUnavailable`;
5. operational acquisition failure in the Configuration Source before a
   detached load result exists -> `SourceUnavailable`;
6. Loader identity/schema/semantic failure after source acquisition, Builder
   failure, or pre-Load execution-binding failure -> `PreparationFailed`;
7. Host Start/readiness/rollback failure -> `StartupFailed`;
8. claimed Stop/shutdown-completion failure -> `ShutdownFailed`;
9. proven loss of a previously bound execution -> `ExecutionLost`;
10. every other trusted operational cause -> fail-closed `InternalFailure`.

An error from a persistence-backed Configuration Source is classified by its
Source role as `SourceUnavailable`, not by storage technology. Conversely, a
durable command-store error is `PersistenceUnavailable`; if its commit is
uncertain it is `OutcomeIndeterminate`. Recovery evidence uncertainty is
`RecoveryBlocked` unless a conditional mutation may have committed, in which
case `OutcomeIndeterminate` wins.

One source fact produces one primary category. Cleanup or rollback causes that
remain distinguishable in the internal joined chain may produce bounded
secondary safe facets only if each has an explicit allowlist mapping. They do
not replace the primary category or expose raw text.

## 11. Correlation and Scope Isolation

Projection occurs only after current authorization permits observation of the
exact operational domain, Workspace, Configuration, Runtime Instance, action,
and target version needed for the source fact. Stored authorization is never
replayed as authority. A caller that is not authorized to observe that exact
fact receives no operational report about it; the separate request outcome
must preserve anti-enumeration semantics.

Opaque correlations are included only when the report consumer is authorized
to observe that exact object. An operational failure of authorization
evaluation may be reported only at a parent scope already authorized for that
consumer, without a narrower correlation or target-existence disclosure.

The same opaque command key in a different command scope is unrelated. Report
lookup, replay, aggregation, and delivery must never join across scopes from a
raw identifier alone.

## 12. Redaction Policy

Redaction is construction by allowlist, not removal from a completed raw
message. Every field has a fixed safe type, maximum size, and disclosure rule.
Unrecognized fields, categories, owner types, and causes are omitted and
classified fail closed.

The following are always forbidden report content:

- Secret values, credentials, tokens, authorization headers, cookies, keys,
  certificates, and private material;
- raw ConfigurationVersion, Snapshot, request, response, message, header,
  query, environment, or repository payload;
- raw internal error text, formatting, cause chain, stack trace, source path,
  SQL, storage key, or vendor response;
- Host pointers, contexts, goroutines, process-local permits, memory addresses,
  unrestricted PID/process metadata, or socket endpoints;
- existence or state of an unauthorized or cross-Workspace object;
- user-controlled labels or identifiers unless a separate field rule proves a
  bounded normalized safe representation.

Opaque domain identifiers are correlation, not Secrets, but remain subject to
exact scope authorization.

## 13. Public Details

The initial allowed detail types are closed enums and bounded facts already
safe in authoritative state, such as operation, lifecycle phase, expected
versus observed revision relation, selected schema version, and whether exact
inspection is required.

A public detail never contains a configuration field value, Listener address,
secret reference name, storage location, provider message, or arbitrary error
parameter. If an operator needs deeper diagnosis, privileged internal tooling
may inspect the original owner boundary under a separate authorization and
audit design; it does not weaken this report.

## 14. Publication Ordering

A report is projected only after its source fact is authoritative:

- authorization-evaluation and source-access reports after their operational
  failure is authoritative, never after an ordinary denial or not-found;
- lifecycle reports after the exact Owner outcome;
- durable command reports after the corresponding immutable command outcome;
- activation phase reports after that phase publication;
- recovery reports after the exact conditional reconciliation publication or
  a proven blocked assessment.

No report may announce Running, Stopped, Failed, command completion, recovery
release, or cleanup before the owning boundary publishes or proves it. Report
creation is downstream observation and never a linearization point for domain
state.

## 15. Replay and Deduplication

The report key is derived from exact scope, source fact identity/revision,
projection schema version, operation, and phase. Under the same projection
schema version, replaying the same immutable command outcome or recovery fact
produces a byte-equivalent semantic report and the same report key. A newer
projection schema creates a distinct projection and key; it never rewrites the
older projection or claims sink-level exactly-once identity across upgrades.

At-least-once delivery is permitted. A sink may deduplicate by report key.
Duplicate delivery is not duplicate lifecycle work, a new failure, or a new
command outcome. A changed authoritative revision or distinct failure phase
has a distinct key.

The contract does not require exactly-once delivery or a global report store.

## 16. Delivery Failure

Delivery occurs after safe projection. A delivery adapter cannot receive the
raw cause and cannot request lifecycle replay.

A definitive or indeterminate delivery failure:

- does not change the source domain fact;
- does not fail an otherwise completed management command;
- does not reopen or close DP-015 admission;
- does not claim a new attempt, phase, or recovery pass;
- may be exposed through a bounded local health signal that contains sink
  identity category and report key only, never the original raw cause.

Recursive reporting into the failing adapter is forbidden. Adapter health,
buffering, retry, backpressure, retention, and loss policy require a concrete
delivery design.

## 17. Cancellation and Concurrency

Cancellation before an authoritative operational failure produces no error
report for an operation that never occurred. Cancellation after an
authoritative fact may end the caller's wait but cannot retract the fact or its
report eligibility. An ordinary caller-cancellation outcome is not relabeled
as an operational error by this design.

Concurrent projectors for the same fact converge on the same report key and
semantic content. They do not share lifecycle locks or hold aggregate/command
locks across delivery. Different Runtime Instances progress independently.

A report observed before a later state change remains truthful history and is
not rewritten. The later authoritative fact creates its own report.

## 18. Recovery Interaction

DP-017 recovery may report only assessed or conditionally published facts. A
stale Running record, PID, port, probe, timeout, or process absence is never
reported as current liveness or graceful shutdown proof.

`RecoveryBlocked` distinguishes unresolved truth from a terminal execution
failure. It reveals no contradictory evidence detail beyond safe phase and
inspection requirement. Report delivery never releases recovery claim or
command admission.

Crash-resumed reconciliation under the same projection schema reprojects the
same already-published facts with the same report keys and never repeats
lifecycle work. After a schema upgrade, a new-version projection has a distinct
key and still performs no lifecycle work.

## 19. Failure Matrix

| Source condition | Safe report | Forbidden projection |
| --- | --- | --- |
| authorization denial | no operational error report | relabel valid decision as failure |
| authorization evaluation failure | `AuthorizationUnavailable` | reason, target existence, policy internals |
| malformed, not-found, or conflict outcome | no operational error report | relabel valid decision as failure |
| operational Source acquisition failure before detached result | `SourceUnavailable` | source location, vendor response, payload |
| Loader semantic/identity, Builder, or pre-Load binding failure | `PreparationFailed` | Configuration/Snapshot/cause text |
| definite no-effect durable-store failure | `PersistenceUnavailable` | storage location, vendor response, retry permission |
| Host Start/rollback failure | `StartupFailed` with safe rollback facet | joined raw causes |
| unproved shutdown completion | `ShutdownFailed` | `Stopped` or graceful claim |
| proven process loss | `ExecutionLost` | liveness, adoption, automatic restart |
| recovery evidence unresolved, no mutation uncertainty | `RecoveryBlocked` | guessed terminal state |
| any domain mutation effect unknown | `OutcomeIndeterminate` | success/failure inference or retry permission |
| unknown owner cause | `InternalFailure` | raw fallback message |
| report sink failure | separate bounded adapter health | mutation of original outcome or recursion |

## 20. Technology Neutrality

Projection, report key, category, safe detail, and delivery adapter are
semantic requirements. They do not require structured logging, OpenTelemetry,
Prometheus, a message broker, database, outbox, queue, audit product, error
registry library, or a global event bus.

Implementations must not introduce a universal diagnostics service with domain
mutation authority, a service locator, or a dependency from Runtime internals
to one observability product.

## 21. Conceptual Operations

The design requires capability equivalent to:

```text
ClassifyAuthoritativeFailure
ProjectSafeOperationalReport
DeriveScopedReportKey
DeliverSafeReport
ObserveDeliveryHealth
```

These are explanatory capabilities, not API or Go interface names. Only
projection receives the internal cause; delivery receives the safe report.

## 22. Acceptance Proofs

A future implementation must prove at minimum:

1. component error identity and cause chains remain unchanged at their owners;
2. valid negative and idempotent outcomes are not operational errors;
3. every report is downstream of one authoritative fact;
4. owner-and-phase precedence maps every fact to one primary stable category
   without message parsing, including Source/Preparation, persistence, and
   recovery overlap cuts;
5. unknown causes fail closed to `InternalFailure` with no raw text;
6. authorization denial and other valid negative outcomes create no error
   report, while authorization evaluation failure reveals neither target
   existence nor policy detail;
7. cross-scope identifiers cannot correlate or retrieve a report;
8. Secrets, payloads, raw causes, stacks, paths, process authority, and
   unrestricted metadata never enter the safe report;
9. allowlisted details are bounded and typed;
10. startup and rollback failures remain internally distinguishable while the
    report exposes only safe primary and optional mapped facets;
11. Stopped is never reported without exact shutdown-completion proof;
12. indeterminate mutation never becomes success, failure, or retry permission;
13. recovery-blocked reporting neither opens admission nor guesses liveness;
14. same immutable fact and projection-version replay produces the same
    semantic report and key; a new projection version produces a distinct key;
15. duplicate delivery performs no lifecycle or command mutation;
16. delivery failure does not change the authoritative outcome or recurse;
17. cancellation after fact publication cannot retract report eligibility;
18. concurrent projection converges without holding domain locks across I/O;
19. different Runtime Instances progress independently;
20. EN/RU contract, matrices, gates, and Planned status remain aligned.

Proofs include technically available redaction corpus, failure-injection,
concurrency, replay, cross-scope, and adapter-failure scenarios. They do not
authorize production activation.

## 23. ARCH-004 Section 19 Gates

This Draft proposes the candidate focused contract for section 19(6). Because
it is non-normative, section 19(6) remains a formal blocker until a separate
approval/status decision. Sections 19(2)–(5) also remain blocked pending their
own DP-014–DP-017 status decisions.

The dependency-ordered candidate design set is now documented, but no
management implementation is Ready merely because this Draft exists. A
separate task must assess and record formal status decisions for the complete
set, then select the smallest implementation slice from repository evidence.

## 24. Explicit Deferrals

Deferred to focused decisions or implementation tasks:

- public error DTOs, HTTP status codes, UI text, localization, and client
  compatibility;
- concrete authorization and privileged diagnostic access;
- log/metric/trace/audit/alert schemas, sinks, sampling, buffering, retry,
  backpressure, loss, retention, and deletion;
- storage schema, outbox, transaction mechanics, migrations, and adapters;
- operator identity, audit evidence, compliance policy, and data residency;
- automated remediation, restart, rollback, retry/backoff, supervision, and
  Production Activation.

## 25. Implementation Boundary

Implementation Status is Planned. The repository contains no report model,
projector, redaction implementation, delivery adapter, management API,
durable management store, recovery executor, or production wiring.

This Draft does not satisfy section 19(6) formally, does not promote any
predecessor DP, and removes no section 19 implementation gate.

## 26. Decision

UWP will preserve exact failures at their existing owners and expose only an
immutable, scoped, allowlist-constructed operational projection after the
corresponding authoritative fact. Reports use stable semantic categories,
opaque authorized correlations, bounded typed details, and
projection-version-scoped deterministic replay keys. Unknown content fails
closed.

Raw errors, Secrets, payloads, stacks, process-local authority, and cross-scope
facts never cross the reporting trust boundary. Report delivery is downstream
observation: its failure cannot change lifecycle or command truth, authorize
retry, reopen admission, or repeat work. Concrete observability and transport
products remain replaceable adapters and later decisions.

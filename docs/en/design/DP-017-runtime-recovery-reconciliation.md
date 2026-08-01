# DP-017: Runtime Recovery and Reconciliation

[Russian version](../../ru/design/DP-017-runtime-recovery-reconciliation.md)

## 1. Status

- **Design Status:** Approved
- **Implementation Status:** Planned

This approved design defines a planned
recovery and reconciliation boundary after loss of Control Service
process-local Runtime ownership. No recovery package, store, schema, execution
adapter, API, scanner, or production wiring exists as a result of this
document.

## 2. Purpose

Define how one Runtime Instance becomes safe to manage after Control Service
termination or equivalent loss of every live lifecycle and command permit. The
contract compares exact durable facts with authoritative execution evidence,
publishes only proven outcomes, and opens command admission only after the
whole linked lifecycle and command set is terminal and coherent.

## 3. Authority

This proposal refines, without overriding:

- [ARCH-004](../architecture/ARCH-004-runtime-deployment-and-identity-model.md),
  especially sections 9, 12, 14, and 19(5);
- [DP-013](DP-013-runtime-management-routing.md) for exact Target routing and
  authorization-before-mutation;
- [DP-014](DP-014-runtime-operational-identity-persistence.md) for coherent
  aggregate revision, attempt history, and conditional lifecycle publication;
- [DP-015](DP-015-runtime-management-command-idempotency.md) for command facts,
  non-transferable permits, linked command sets, and unresolved admission;
- [DP-016](DP-016-runtime-activation-replacement-rollback.md) for phase order,
  proven release, and process-loss cut points.

Accepted ADRs and Active or Frozen architecture remain authoritative. DP-013
remains Draft; Approved DP-014 through DP-017 do not implement their contracts.

## 4. Scope

The design covers:

- exact per-Instance restart assessment;
- one durable recovery claim and one non-transferable recovery permit;
- separation of durable facts, live capabilities, and execution evidence;
- evidence classification for the initial in-process topology;
- reconciliation of active Launch Attempt and primitive or linked commands;
- conditional publication order and crash-resumable convergence;
- command-admission reopening and unresolved outcomes;
- concurrency, cancellation, security, and technology-neutrality constraints.

The design does not define discovery scans, a database schema, process
supervision, PID files, socket probes, child-worker protocol, public recovery
API, transport mapping, automatic restart, reporting, or retention.

## 5. Terms

**Restart assessment** is a coherent read that determines whether one Runtime
Instance is already clean or requires recovery before any state-changing
command may be admitted.

**Recovery claim** is a durable internal coordination fact bound to one exact
Runtime Instance, its starting aggregate revision, and the exact non-terminal
command revisions observed at claim time. It is not a management command,
Launch Attempt, lease by elapsed time, or authority to invoke lifecycle work.

**Recovery permit** is the non-transferable process-local capability returned
only to the path that confirms a new or resumable recovery claim. It authorizes
conditional reconciliation publications for that claim. It never authorizes
Start, Stop, Load, Build, Launch, Host adoption, or command replay.

**Execution generation** is an opaque identity for one execution-containment
boundary. In the initial topology it identifies one Control Service process
generation. It is not a PID, timestamp, liveness lease, Runtime identity, or
command identity.

**Execution binding** is the durable immutable correlation of one exact Launch
Attempt to the execution generation allowed to prepare or own its Host. It is
confirmed after attempt claim and before Load. It proves correlation, not
liveness or that preparation began.

**Execution evidence** is an observation from the execution boundary whose
contract can bind the observation to the exact Launch Attempt and execution
generation and can distinguish proven termination, proven live execution,
proven Host-owned shutdown completion where supported, and unknown. PID,
address, time, stored actual state, or a successful connection is not such
evidence by itself.

**Clean set** means no active attempt, no non-terminal command or phase, no
unresolved recovery claim, and coherent terminal facts at one verified set of
revisions.

**Unresolved set** means any exact aggregate, attempt, primitive command,
parent, phase, recovery claim, or execution fact needed for a truthful outcome
is missing, contradictory, stale, unavailable, or indeterminate.

## 6. Responsibility Boundary

Recovery owns assessment, exclusive reconciliation coordination, evidence
classification, conditional repair publication, and the final decision to
keep or open the per-Instance admission barrier. It does not own Runtime
lifecycle policy, live Host resources, version selection, authorization
policy, transport, diagnostics presentation, or automatic execution.

Runtime Lifecycle Owner remains the only normal lifecycle decision maker and
live Host owner. Recovery never hydrates an Owner from stored fields and never
transfers an old execution or phase permit. Persistence validates facts;
execution evidence reports execution truth. Neither becomes a second Owner.

The exact Control Service composition creates the opaque execution generation
and owns its containment/termination proof. DP-014 owns conditional durable
execution binding inside the Runtime Instance aggregate. The planned DP-011
Start-claim continuation coordinates that borrowed binding capability and the
DP-015 pending-Stop gate. DP-017 consumes the binding during assessment but
allocates neither generation nor binding.

## 7. Recovery Trigger and Assessment

Assessment is required before a state-changing command whenever the current
Control Service process cannot prove process-local ownership corresponding to
durable non-terminal facts. This includes process start, loss of an Owner or
command call stack, an indeterminate lifecycle/command publication, and an
existing unresolved recovery claim.

One coherent assessment reads:

- exact Runtime Instance identity, binding, revision, desired and actual facts;
- active-attempt identity and its exact phase, outcome, and version pin;
- every non-terminal primitive, parent, and phase command for the Instance;
- exact command revisions, immutable intents, and linked identities;
- the immutable execution binding, if one committed for the active attempt;
- any existing recovery claim and its observed-revision set;
- available execution evidence bound to the exact attempt and generation.

A clean set requires no recovery mutation. Read-only Observe remains available
subject to normal authorization and truthful stale-fact labeling. Assessment
does not mutate lifecycle state or prove liveness.

## 8. Fail-Closed Admission

From the first observation of a non-clean set until verified release of the
recovery claim, every new state-changing command and every further parent
phase is rejected without mutation. No tracked-Start Stop exception survives
process loss. Same-key observation may report non-terminal status but cannot
receive a permit or delegate work.

Admission and recovery-claim creation share one per-Instance atomic ordering
boundary. A command cannot pass from a stale clean observation while recovery
claims the Instance, and two recovery paths cannot both obtain permits.

## 9. Durable Recovery Claim

Claiming recovery atomically verifies the exact Instance and command revisions,
confirms the set is non-clean, establishes one recovery claim, closes command
admission, and returns one live recovery permit. A definitive conflict or stale
revision performs zero mutation and requires reassessment.

An existing claim may be resumed only by a new conditional recovery step that
re-reads all exact facts and establishes that the old permit cannot survive.
It does not reuse or recreate that permit. Loss of a recovery permit leaves the
claim unresolved and admission closed; a later recovery pass resumes from
durable publications rather than repeating lifecycle work.

Recovery identity representation, storage mechanism, and claim layout are
implementation decisions. A wall-clock timeout alone never proves permit loss,
execution termination, or safe takeover.

## 10. Evidence Hierarchy

Recovery evaluates three separate truth classes:

1. DP-014 and DP-015 durable facts prove only what was committed at exact
   revisions;
2. current process-local capabilities prove only ownership created in the
   current process and cannot be reconstructed from durable identity;
3. execution evidence proves only the liveness or termination property
   guaranteed by its approved adapter contract.

No class substitutes for another. Persisted `Running`, `Stopping`, a command
`Claimed`, PID, Listener address, port-bind result, elapsed time, log record,
health response, or configuration version does not alone prove present Host
ownership, resource release, or lifecycle completion.

Contradictory evidence is unresolved. Recovery must not select the most recent,
most convenient, or majority observation.

Before any Load, Build, Launcher, or Host work, the planned DP-011 continuation
must confirm DP-014 publication of one execution binding for the exact already-
claimed attempt and expected aggregate revision. Only a coherent exact read
proving no binding for that exact still-active attempt at the expected revision
may produce `BindingFailed`; it performs no external preparation but does not
erase the already committed attempt/Starting mutation. A different generation,
stale revision, conflicting or inactive state, unavailable store, or unknown
result requires an exact re-read and then exact terminal convergence or
`Blocked`, never `BindingFailed`. The continuation's final per-Instance gate
then orders Stop against release to Load. This prerequisite refines planned
DP-016 ordering without authorizing implementation.

While Owner is still live, that coherently proven exact binding absence is not
recovery and does not authorize direct durable terminalization. The continuation returns
`BindingFailed`; Flow submits `FailedPreparation` with the authentic token to
existing Owner.Start. Owner's mutex orders that failure against Stop. Only the
exact Owner outcome may be published durably and then terminalize the command.
DP-017 handles an unbound attempt only after that Owner convergence is lost.

## 11. Initial In-Process Topology

In the current ARCH-004 single-node in-process topology, a Runtime Host cannot
survive termination of the process that owned it. An approved containment
boundary must first establish a unique current generation and prove termination
of the exact prior generation named by the attempt's execution binding. That
proof establishes that no Host from the prior generation remains owned or
runnable. It does not prove graceful Runtime cleanup, successful Stop,
readiness, or the terminal command publication that was lost.

The replacement Control Service must not fabricate a Host reference, hydrate
an Owner, probe a port and call it Running, or adopt any execution. It may use
proven generation termination only to publish phase-sensitive process-loss
failure facts and to clear an active-attempt association when all required
exact facts are coherent. Resource absence alone never proves successful Stop
or Host shutdown completion.

If a future child-process or remote adapter can report a live orphan, that
execution remains non-manageable and admission remains closed until a separate
approved adoption or termination protocol exists. DP-017 does not authorize
that protocol.

## 12. Recovery Classification

Assessment or a claimed recovery path classifies the exact set as one of:

- **Clean:** all attempts and commands are terminal and no recovery claim
  exists; assessment creates no claim or reconciliation mutation;
- **Release-only:** all lifecycle and command facts are terminal and coherent,
  and only the exact recovery claim remains to be conditionally released;
- **Command only:** a command was claimed, but exact aggregate/attempt facts
  prove that Owner never claimed a Launch Attempt and no lifecycle mutation
  began;
- **Unbound attempt:** Owner claimed the exact attempt and published Starting,
  but an exact conditional read proves that execution binding never committed;
  this proves no external preparation, not no lifecycle mutation;
- **Execution terminated:** an active attempt existed and authoritative
  evidence proves its exact execution generation terminated;
- **Resource absence:** a Stop or replacement release phase was in progress and
  evidence plus exact aggregate facts prove no old Host resources remain, but
  do not prove Host shutdown completion;
- **Shutdown completed:** exact authoritative evidence proves the Host completed
  its owned shutdown contract for the exact attempt;
- **Live orphan:** authoritative evidence reports execution that the current
  process does not own;
- **Unknown:** evidence is absent, stale, contradictory, indeterminate, or
  cannot bind to the exact attempt and generation.

Clean returns before claim. Release-only, Command only, Unbound attempt,
Execution terminated, and evidence-backed shutdown completion can progress
under the initial topology. Resource absence without shutdown-completion proof,
Live orphan, and Unknown preserve truthful failure/unresolved outcomes and never
fabricate Stopped.

## 13. Phase-Sensitive Attempt Reconciliation

Recovery never creates or reuses a Launch Attempt. A definitively unbound
active attempt is already a durable lifecycle mutation in actual `Starting`.
Because the DP-011 gate forbids external preparation before binding, recovery
may conditionally publish that exact attempt historical `Failed` with a stable
unprepared-process-loss category and clear it as resource-free. If binding
publication is indeterminate or cannot be read coherently, the attempt remains
active and unresolved.

For a proven terminated bound generation:

- an attempt claimed for desired `Running` that did not reach a confirmed
  resource-free terminal fact becomes historical `Failed` with a stable
  process-loss category;
- persisted `Running` becomes actual `Failed`, never restored Running;
- `Preparing` or `Launching` becomes failed even when no readiness was
  published; recovery does not infer how far startup progressed;
- a Stop-claimed attempt becomes resource-free `Failed`/interrupted when
  generation termination proves only resource absence;
- historical Stopped is allowed only when exact authoritative evidence proves
  that Host completed its owned shutdown contract for that attempt, not merely
  that the process or resources disappeared;
- otherwise the active association and truthful non-terminal or Failed fact
  remain and the set stays unresolved.

Clearing the active-attempt reference uses DP-014 exact revision and is allowed
only with proven resource absence. Historical identity and version pin never
change.

## 14. Primitive Command Reconciliation

A Claimed primitive command terminalizes only from exact durable lifecycle
facts:

- if Owner never claimed an attempt, publish a stable recovered-no-mutation
  outcome;
- if Owner claimed an exact attempt but binding is definitively absent,
  terminalize that resource-free attempt Failed first, then publish a stable
  failed-before-preparation command outcome;
- if Start owns the exact attempt terminated by process loss, publish a stable
  failed outcome linked to that attempt after its terminal publication;
- if Stop targets the exact attempt and only generation termination/resource
  absence is proven, publish a stable interrupted/process-loss failure after
  terminal Failed publication;
- publish stopped/satisfied only when exact evidence proves completion of the
  Host-owned shutdown contract;
- if the command-to-attempt relationship or outcome cannot be proved, retain
  Claimed and keep the barrier closed.

Recovery does not invoke DP-013, call Flow or Owner, replay the command, or
manufacture identities missing from the existing outcome contract.

## 15. Parent and Phase Reconciliation

DP-016 parent orchestration is reconciled from its immutable linked phase set:

- every claimed phase is resolved before its parent;
- an unclaimed later phase remains absent and is never created by recovery;
- old-Stop with exact Host-shutdown completion and no Start phase leaves the
  Instance Stopped and terminalizes the parent as interrupted after release;
- old-Stop with only process termination/resource absence leaves the old
  attempt and Instance Failed and terminalizes the phase/parent as interrupted
  by process loss; no Start phase is created;
- a claimed Start phase with a terminated exact attempt terminalizes failed
  after the attempt and phase facts;
- a pending Stop without a surviving permit is resolved only when its exact
  no-mutation, Host-shutdown-complete, or process-loss failure outcome is proven;
- any missing phase link, indeterminate publication, or contradictory ordering
  keeps the entire linked set unresolved.

The parent becomes Terminal only after every existing phase is terminal and
their outcomes form one valid DP-016 ordering. Recovery never continues to a
later phase.

## 16. Reconciliation Publication Order

No distributed transaction across aggregate and command stores is assumed.
While the durable recovery barrier remains closed, publications proceed
monotonically:

1. re-read and verify exact aggregate, attempt, command, and recovery revisions;
2. conditionally publish the phase-sensitive attempt/aggregate terminal fact;
3. conditionally terminalize primitive or linked phase commands from that fact;
4. conditionally terminalize the parent after all existing phases;
5. coherently verify the entire set at the resulting revisions;
6. release the recovery claim and open admission atomically against that exact
   verified set.

Each step is idempotent by exact identity, immutable outcome, and conditional
revision. A crash or indeterminate publication at any step leaves admission
closed. The next pass inspects and resumes; it never repeats lifecycle work.

## 17. Barrier Reopening

The barrier opens only when one coherent verification proves:

- no active attempt remains;
- every primitive command, parent, and existing phase is Terminal;
- no contradictory or unknown execution evidence remains relevant;
- aggregate, attempt, command, and recovery revisions match the verified set;
- recovery-claim release commits.

Partial command terminalization, a cleared attempt alone, or a successful
publication response whose commit is indeterminate does not open admission.
After reopening, a later explicitly authorized Start may create a fresh
attempt through ordinary DP-015/DP-016 ordering. Recovery itself never does.

## 18. Desired and Actual Facts

Recovery introduces no new public desired or actual state. Desired remains the
last accepted management intent. Actual is reconciled using existing
`Stopped|Starting|Running|Stopping|Failed` facts:

- proven process loss while desired Running publishes actual Failed;
- exact proof that Host completed its owned shutdown contract for accepted
  desired Stopped may publish actual Stopped;
- process termination or resource absence without that proof publishes actual
  Failed/interrupted, never Stopped;
- unknown release never publishes Stopped;
- persisted Running is never refreshed or preserved as live solely because it
  was durable.

Recovery categories and claim status are internal coordination/audit facts,
not additional Runtime lifecycle states.

## 19. Concurrency and Linearization

The required semantic linearization points are:

1. assessment versus new command admission;
2. one recovery claim and one newly issued recovery permit;
3. each exact conditional attempt or command reconciliation publication;
4. final coherent-set verification and recovery release versus new admission.

Operations on different Runtime Instances may progress independently. Long
evidence collection does not hold aggregate or command locks, but revisions
must be revalidated before every publication. A concurrent observer never
receives mutation authority.

## 20. Failure and Indeterminate Matrix

| Failure cut | Truthful result | Forbidden consequence |
| --- | --- | --- |
| clean assessment | zero recovery mutation | create recovery claim or lifecycle work |
| stale assessment/claim conflict | reassess exact revisions | overwrite newer facts |
| recovery permit loss | recovery claim unresolved | recreate permit by timeout |
| command claimed, no attempt exists | recovered no-lifecycle-mutation terminal outcome | create attempt or invoke lifecycle |
| active attempt, binding definitively absent | resource-free Failed attempt and failed-before-preparation command | call claim no mutation or begin Load |
| binding publication/inspection indeterminate | active attempt and linked set unresolved | infer absence or bind a new generation |
| execution generation termination proven | reconcile exact attempt phase | claim graceful cleanup occurred |
| Stop claimed, only resource absence proven | Failed/interrupted process-loss outcome | publish Stopped or stopped/satisfied |
| exact Host shutdown completion proven | phase-sensitive Stopped may publish | infer proof from process termination alone |
| live orphan observed | barrier remains closed | hydrate Owner or adopt Host |
| evidence unavailable or contradictory | unresolved | infer release from PID, port, or time |
| attempt publication indeterminate | inspect exact revision | terminalize commands from assumption |
| phase terminal, parent publication lost | resume from durable phase | rerun Stop or Start |
| all commands terminal, release indeterminate | barrier remains closed | admit a new command |
| recovery cancellation | caller may stop waiting; claim remains | erase claim or transfer permit |

## 21. Caller Cancellation

Cancellation before recovery claim performs zero mutation. After claim it may
end only the current caller's wait. It does not delete the recovery claim,
prove an execution outcome, open admission, or transfer the recovery permit.

If cancellation wins before a conditional publication, the path may return
with the claim unresolved. If a publication may have committed, exact
inspection is mandatory. A later pass resumes from durable truth.

## 22. Security and Scope Isolation

Assessment and reconciliation are scoped to one exact operational management
domain, Workspace, Configuration, Runtime Instance, attempt, and linked command
set. Recovery cannot use cross-Workspace evidence or disclose whether an
unauthorized identity exists.

Evidence and outcomes contain opaque identities and stable semantic categories,
not credentials, Secrets, Configuration/Snapshot payloads, raw internal errors,
stack traces, Host pointers, process-local permits, or unrestricted process
metadata. Concrete operator reporting and redaction remain section 19(6).

## 23. Technology Neutrality

Recovery claim, revision, generation, evidence, conditional publication, and
barrier are semantic requirements. They do not require a database product,
transaction coordinator, lease service, workflow engine, queue, PID file,
signal mechanism, process supervisor, clock, or identifier format.

An implementation may add private mechanics only if it proves exact binding,
single recovery authority, crash-resumable publication, and fail-closed
admission without becoming a generic registry or service locator.

## 24. Conceptual Operations

The design requires capability equivalent to:

```text
AssessRuntimeRecovery
ConditionalClaimRecovery
ReadExactExecutionEvidence
ReadExactExecutionBinding
ConditionalReconcileUnboundAttempt
ConditionalReconcileAttempt
ConditionalReconcileCommandOrPhase
ConditionalReconcileParent
ConditionalReleaseRecovery
```

These are explanatory semantic capabilities, not API or Go interface names.
None invokes Runtime lifecycle work.

## 25. Acceptance Proofs

A future implementation must prove at minimum:

1. clean Stopped Instance performs zero recovery mutation;
2. non-clean assessment atomically excludes new command admission;
3. concurrent recovery paths issue at most one live recovery permit;
4. persisted Running, PID, port, time, and health response never alone prove
   ownership or liveness;
5. execution binding commits before any Load and never proves liveness;
6. coherently proven exact binding absence for the exact still-active attempt
   at the expected revision converges through Owner.Start and races Stop under
   Owner's mutex before durable command terminalization;
7. exact in-process generation termination permits no Host adoption;
8. process loss before Owner claim resolves as no lifecycle mutation;
9. definitively absent binding after Owner claim resolves the exact already-
   Starting attempt as resource-free Failed, never as no mutation;
10. indeterminate binding remains active and unresolved with no Load;
11. process loss during Starting terminalizes the exact attempt Failed only
   after resource absence is proven;
12. process loss after Running never preserves Running as current truth;
13. Stop-in-progress publishes Stopped only from exact Host-owned shutdown
    completion proof, never resource absence alone;
14. primitive command outcome follows the exact reconciled attempt fact;
15. every existing phase terminalizes before its DP-016 parent;
16. recovery never claims a missing later phase or invokes DP-013;
17. crash after each publication resumes without lifecycle replay;
18. stale revisions and contradictory evidence perform no mutation;
19. barrier opens only for one coherent fully terminal set;
20. cancellation and indeterminate outcomes leave admission closed;
21. different Instances recover independently;
22. EN/RU contract, matrices, gates, and Planned status remain aligned.

Proofs include technically available concurrency, race, failure-injection,
durability, process-restart, and recovery-restart scenarios. They do not
authorize production activation.

## 26. Formal and Downstream ARCH-004 Section 19 Gates

This Approved design closes the focused architecture design gate for ARCH-004
section 19(5). Approved DP-014 through DP-016 and DP-018 close the other
focused design gates in sections 19(2)–(4) and 19(6). The complete approved
set defines fail-closed recovery ordering but creates no store, execution
adapter, recovery executor, reporting, integration, or Production Activation.

## 27. Explicit Deferrals

Deferred to focused designs or implementation tasks:

- startup discovery, enumeration, scheduling, and recovery worker lifecycle;
- storage schema, transaction mechanics, migrations, and adapters;
- child/remote process protocol, supervision, adoption, and forced termination;
- automatic restart, rollback, retry/backoff, failover, and policy evaluation;
- public recovery controls, API/DTO/status mapping, and concrete authorization;
- operational error taxonomy, reporting, audit, metrics, alerting, and
  redaction policy;
- retention/deletion and Production Activation.

## 28. Implementation Boundary

Implementation Status is Planned. The repository contains no durable Runtime
aggregate or command store, recovery claim, execution-evidence adapter,
recovery executor, public management API, or production wiring.

The current in-process Runtime components do not survive Control Service
process termination and expose no restart-time recovery capability. Creating
approval closes the section 19(5) design gate but does not implement or wire
the contract.

## 29. Decision

UWP will recover one Runtime Instance through an exact, durable, fail-closed
reconciliation claim. Durable lifecycle and command facts remain
last-confirmed history; current liveness or release requires authoritative
attempt-and-generation-bound execution evidence. No stored Running fact, PID,
address, clock, or probe result alone recreates ownership.

Recovery performs no Start, Stop, Load, adoption, retry, or new Launch Attempt.
It conditionally terminalizes the exact attempt and linked command set from
proven facts, remains crash-resumable under a closed admission barrier, and
opens that barrier only after one coherent fully terminal verification. Unknown
or contradictory truth remains unresolved. Automatic restart and operational
reporting remain separate decisions.

# DP-014: Runtime Operational Identity Persistence

[Russian version](../../ru/design/DP-014-runtime-operational-identity-persistence.md)

## 1. Status

- **Design Status:** Approved
- **Implementation Status:** Implemented in isolation

This approved design defines a durable boundary for Runtime Instance and
Launch Attempt identity and history. Package `internal/runtimeidentity`
implements all nine conceptual operations from §21 and satisfies all
acceptance proofs from §22 as an isolated in-process in-memory store. No
external storage, HTTP API, production wiring, or second lifecycle owner
exists as a result of this implementation.

## 2. Purpose

Define the smallest persistence contract that preserves the operational
identity model of ARCH-004 across process lifetime without creating a second
Runtime Lifecycle Owner.

The contract makes Runtime Instance binding, Launch Attempt history, and the
last Owner-confirmed desired and actual facts durable. It does not make stored
facts proof of current process liveness.

## 3. Authority

This proposal refines, without overriding:

- [ADR-0002](../adr/0002-configuration-dsl.md);
- [ADR-0003](../adr/0003-runtime-architecture.md);
- [ADR-0004](../adr/0004-handshake-runtime-dependencies.md);
- [ARCH-002](../architecture/ARCH-002-runtime-foundation-freeze.md);
- [ARCH-004](../architecture/ARCH-004-runtime-deployment-and-identity-model.md);
- [ARCH-005](../architecture/ARCH-005-runtime-configuration-snapshot-and-loading-model.md);
- [DP-010](DP-010-runtime-lifecycle-owner-contract.md);
- [DP-011](DP-011-runtime-launch-pipeline-integration.md);
- [DP-012](DP-012-runtime-source-composition.md);
- [DP-013](DP-013-runtime-management-routing.md).

Active ARCH-004 takes precedence over this Approved design. DP-010 remains the contract
for the process-local Owner and live Host ownership. DP-013 remains
Draft/Planned and Ready only for bounded isolated implementation.

## 4. Scope and Non-goals

This design defines:

- one durable Runtime Instance aggregate;
- immutable Workspace, Configuration, and Runtime Instance binding;
- append-only Launch Attempt history;
- durable opaque identity allocation boundaries;
- atomic and conditional publication of lifecycle facts;
- coherent aggregate and history reads;
- failure and indeterminate-outcome rules;
- concurrency, uniqueness, security, and redaction invariants.

This design does not define:

- a database, repository product, table, document, index, migration, or ORM;
- Go interfaces, packages, HTTP endpoints, DTOs, or status codes;
- durable management command idempotency;
- activation, replacement, rollback, recovery, or reconciliation;
- operational reporting, logging, metrics, audit retention, or alerting;
- deletion, retention, automatic restart, scheduling, clustering, or process
  isolation;
- Configuration payload or Secret storage.

## 5. Terms

**Operational management domain** is one authority boundary in which
Runtime Instance IDs are unique and management routing resolves an ID to one
aggregate.

**Runtime Instance aggregate** is the durable consistency boundary rooted at
one Runtime Instance identity.

**Aggregate revision** is an opaque, monotonically advancing value identifying
one committed aggregate state. It is a concurrency token, not a timestamp or
business identity.

**Attempt history** is the complete append-only membership of committed Launch
Attempt children owned by one Runtime Instance. Membership, parent identity,
child identity, and exact version pin are immutable. Lifecycle phase and
outcome facts may advance conditionally within that same child.

**Last confirmed fact** is a desired or actual lifecycle fact explicitly
published by the Runtime Lifecycle Owner at a defined linearization point.

**Indeterminate outcome** means the caller cannot determine whether the
requested atomic publication committed.

## 6. Ownership and Aggregate Boundary

Runtime Instance is the aggregate root. Its immutable identity binding,
revision, desired and actual facts, optional active-attempt reference, and
complete Launch Attempt history form one consistency boundary.

A Launch Attempt is an owned child of exactly one Runtime Instance aggregate.
It is not an independently mutable aggregate and cannot move between Runtime
Instances.

The Runtime Lifecycle Owner remains the only lifecycle decision maker and the
only owner of the live Runtime Host reference. Persistence validates and
conditionally publishes facts; it does not start or stop Host, select
Configuration, own resources, route arbitrary services, or become a registry
or service locator.

## 7. Durable Facts

The aggregate preserves only the facts required by ARCH-004:

- immutable Workspace, Configuration, and Runtime Instance identity binding;
- current aggregate revision;
- last Owner-confirmed desired state;
- last Owner-confirmed actual state;
- optional identity of the active Launch Attempt;
- complete append-only Launch Attempt history;
- for each attempt, immutable identity and exact Published
  ConfigurationVersion pin;
- for a claimed attempt that may enter external preparation, the optional
  immutable opaque execution-generation binding required by Approved
  [DP-017](DP-017-runtime-recovery-reconciliation.md);
- committed phase and terminal outcome facts required to distinguish claimed,
  running, stop-claimed, stopped, and failed attempts.

The persistence contract does not copy the Configuration payload, Snapshot,
Secret values, Host pointer, context, goroutine, PID, socket, or Session state.

## 8. Opaque IDs and Namespaces

RuntimeInstanceID and LaunchAttemptID are opaque, non-zero identities. Their
representation and generation algorithm are outside this design.

RuntimeInstanceID is unique within one operational management domain. A
candidate ID that already identifies any current or historical Runtime
Instance in that domain is rejected.

LaunchAttemptID is unique only within the complete history of its owning
Runtime Instance. Its durable child key is:

```text
(RuntimeInstanceID, LaunchAttemptID)
```

Global LaunchAttemptID uniqueness across different Runtime Instances or
different operational management domains is not required. An attempt ID is
never reused within one Runtime Instance history, including after failure,
stop, replacement, or any future retention transition that preserves the
aggregate identity.

## 9. Runtime Instance Creation

Creation atomically publishes:

- one new RuntimeInstanceID;
- exact immutable WorkspaceID and ConfigurationID binding;
- initial desired `Stopped` and actual `Stopped` facts;
- no active attempt;
- empty attempt history;
- the initial aggregate revision.

Creation either publishes the complete initial aggregate or publishes
nothing. Existing identity, invalid identity, invalid binding, or a stale
creation precondition performs zero mutation. Rebinding an existing Runtime
Instance to another Workspace or Configuration is forbidden.

Candidate identity allocation may occur before publication, but allocation
alone does not prove that an aggregate exists.

## 10. Launch Attempt Claim

Claiming a launch atomically:

- validates the expected aggregate identity and revision;
- validates that current lifecycle facts permit Start;
- validates that no active Launch Attempt exists;
- validates that the candidate LaunchAttemptID is absent from complete
  history;
- appends one new Launch Attempt with the exact immutable Published
  ConfigurationVersion pin;
- establishes it as the only active attempt;
- publishes the corresponding desired/actual claim facts;
- advances aggregate revision once.

No Loader, Builder, Launcher, or Host work begins from a claim that was not
confirmed committed. A retry or replacement creates a distinct candidate
attempt identity only after the outcome of any preceding claim is known.

### Execution Generation Binding

The Control Service composition allocates one opaque execution generation for
its process-containment boundary. Persistence does not allocate it or infer it
from PID, time, address, or caller identity.

After launch claim and before any Load, Build, Launcher, or Host work, the
original tracked Start path through the planned DP-011 continuation may
conditionally bind the exact active attempt to that generation. Publication:

- validates aggregate identity and exact expected revision;
- validates the exact active non-terminal attempt and its immutable version pin;
- validates that no different generation is already bound;
- stores the immutable attempt-to-generation correlation;
- advances aggregate revision once.

An exact already-present same-generation binding is a zero-mutation satisfied
observation. Any conditional rejection performs zero mutation and permits no
external preparation, but rejection alone does not prove binding absence. A
different generation, stale revision, inactive/terminal attempt, conflicting
fact, or unavailable store requires a coherent exact re-read. An exact existing
terminal outcome is converged; a different binding or unresolved conflict is
`Blocked` and must never enter the resource-free failure path.

Only a coherent read proving no binding for the exact still-active attempt at
the expected revision permits `BindingFailed`. The attempt claim itself remains
a durable lifecycle mutation. The same Start path must then converge the
process-local attempt through existing Owner.Start with the authentic token and
`FailedPreparation(bindingFailure)`. Owner's mutex orders that failure against
concurrent Stop. Only the exact returned Owner outcome may request conditional
durable terminal publication: preparation failure publishes resource-free
Failed; a Stop-winning outcome publishes the exact Owner-confirmed stopped-
before-running fact. Command/phase terminalization follows only after that
durable publication is confirmed.

After an indeterminate binding publication, the path coherently inspects the
exact aggregate, attempt, candidate generation, and expected/new revision.
Exact same-generation presence confirms binding; coherently proven absence for
the exact still-active attempt/expected revision permits the Owner-owned
failure-convergence path, not direct persistence mutation, a new generation, or
blind binding retry; different generation, stale/conflicting/inactive facts,
unavailable state, or unknown remains unresolved unless an exact existing
terminal outcome can be converged. A concurrent Stop
is ordered first by the DP-011/DP-016 final continuation gate and then, if the
binding-failure path wins, by Owner's existing mutex against failure acceptance.
An absent or indeterminate Owner/durable terminal outcome remains unresolved.

The binding is retained with attempt history after terminalization. It proves
correlation only and never proves liveness, readiness, preparation, ownership,
or graceful shutdown.

## 11. Running Publication

Only the Runtime Lifecycle Owner may request Running publication after the
exact claimed attempt has completed Host startup and readiness.

Running publication atomically validates aggregate identity, expected
revision, active attempt identity, and allowed prior attempt phase; publishes
the attempt and aggregate actual `Running` fact; preserves the exact version
pin; and advances revision once.

It cannot create an attempt, switch the active attempt, repair missing
history, or publish Running for a stale, terminal, or different attempt.

## 12. Stop and Terminal Publication

Stop claim atomically records transfer of shutdown responsibility for the
exact active attempt, publishes the corresponding desired and phase facts,
and advances revision. Repeated or stale stop claims do not create another
shutdown owner.

Phase-sensitive stop publication follows these rules:

- an attempt still in `Preparing` that produced no owned Host may atomically
  publish desired `Stopped`, actual `Stopped`, and a historical
  stopped-before-running outcome;
- an attempt in `Launching` or `Running` first publishes a same-attempt Stop
  claim and actual `Stopping`, then publishes a terminal outcome only from the
  confirmed shutdown result;
- a stop failure or unproven cleanup preserves the active-attempt association
  and publishes only truthful `Failed` or `Stopping` facts that do not claim
  resource release.

Confirmed terminal publication atomically:

- validates exact aggregate, revision, and attempt identity;
- records the phase-sensitive stopped or failed outcome;
- publishes the last Owner-confirmed actual fact;
- clears the active-attempt reference only when the Owner proves that no Host
  resources remain, or that startup produced no owned Host;
- retains the complete immutable attempt history;
- advances revision once.

Actual `Stopped` is published only after the Owner confirms that Host-owned
resources are released or that startup produced none. A stop-operation failure
or cleanup-unproven fact is not a terminal historical attempt outcome: the
attempt remains active in `AttemptStopping` and its association remains
retained. A terminal historical `Failed` outcome may be published only when
the phase-sensitive contract also proves that no Host resources remain or that
startup produced no owned Host.

## 13. Desired, Actual, and Liveness

Desired and actual state remain distinct. Desired state records the last
accepted management intent at a defined lifecycle publication. Actual state
records the last lifecycle fact confirmed by the Owner.

Persisted actual `Running`, `Stopping`, `Stopped`, or `Failed` is historical
operational knowledge. After loss of the Owner or Control Service process it
is not proof of present Host, process, socket, or resource liveness.

Only an approved recovery and reconciliation contract may compare durable
facts with external execution evidence and publish a reconciled state. Approved
[DP-017](DP-017-runtime-recovery-reconciliation.md) defines that recovery
contract; its Planned implementation remains absent. This design does not infer
liveness from stored PID, address, time, or an earlier Running fact.

## 14. Atomicity

The following are separate required atomic publications:

1. initial Runtime Instance aggregate;
2. one active Launch Attempt claim and history append;
3. `ConditionalBindExecutionGeneration` for the exact claimed attempt;
4. Running publication;
5. Stop claim;
6. phase-sensitive terminal publication.

Each publication commits all specified aggregate and attempt facts with one
new revision or commits none. No observer may see a partial identity binding,
attempt without history, history entry without its exact version pin, two
active attempts, a historical terminal or cleared attempt still active, or a
lifecycle fact without the revision that committed it. A retained active
`AttemptStopping` with a stop-failure or cleanup-unproven fact is explicitly
permitted and is not a historical terminal attempt.

Atomicity is a semantic requirement and does not prescribe transaction,
locking, consensus, or storage implementation.

## 15. Uniqueness and History

The durable boundary enforces:

1. one immutable aggregate for each RuntimeInstanceID in a management domain;
2. one permanent Workspace and Configuration binding for that aggregate;
3. at most one active Launch Attempt;
4. one immutable version pin for each attempt;
5. no reuse of a LaunchAttemptID within the aggregate history;
6. append-only history membership: a committed child is never removed or
   replaced;
7. immutable parent identity, LaunchAttemptID, and exact Published
   ConfigurationVersion pin for every child;
8. lifecycle phase and outcome facts advance only conditionally within the
   same child, never regress, and never rewrite its immutable facts;
9. no transition of a terminal historical attempt back to active.

Deletion and retention are deferred. Until a focused contract defines them,
the design assumes identity and attempt history remain available for
uniqueness and inspection.

## 16. Concurrency and Conditional Revisions

All mutations of one Runtime Instance aggregate use one serialization boundary
and an exact expected revision. A successful publication advances revision
monotonically. Concurrent operations on different Runtime Instances may
progress independently.

A stale revision, wrong active attempt, wrong phase, or mismatched immutable
binding is rejected with zero mutation. The persistence boundary must not
silently re-read and reinterpret an operation against newer state, select a
latest attempt, merge conflicting lifecycle intents, or retry detached from
the caller.

Revision representation, increment size, and storage mechanism are not
specified. The only observable contract is equality for the expected revision
and monotonic change after each successful publication.

## 17. Failure and Indeterminate Outcomes

A definitive rejection or definitive commit failure publishes nothing and
returns enough category information for the caller to distinguish invalid,
missing, stale, conflicting, or unavailable outcomes without exposing
sensitive data.

For an indeterminate outcome, the caller must not:

- retry blindly with a new RuntimeInstanceID or LaunchAttemptID;
- assume that no mutation occurred;
- launch Host from an unconfirmed claim;
- publish a later phase before resolving the earlier publication.

The caller first performs a coherent inspection of the exact aggregate
identity, candidate child identity, and expected/new revision. It then
determines whether the requested publication is present, absent, or still
unknown. This inspect-after-indeterminate rule prevents duplicate identity and
history; it does not define command deduplication from ARCH-004 section 19(3)
or process recovery from section 19(5).

## 18. Coherent Reads

A read of one Runtime Instance returns one coherent aggregate revision:
immutable binding, desired and actual facts, active-attempt reference, and the
attempt facts exposed with that view cannot come from mutually incompatible
revisions.

History reads must preserve every committed attempt and its exact version pin.
Pagination, streaming, snapshots, and read consistency mechanisms are
implementation choices, but they must not fabricate ordering, omit committed
entries while claiming completeness, or present a child under another
aggregate.

Reads are observation only. They do not claim lifecycle ownership, refresh
liveness, repair state, or advance revision.

## 19. Security and Redaction

Every operation is scoped to the exact operational management domain and
Runtime Instance identity. Workspace and Configuration binding cannot be
altered to cross an authorization boundary.

The durable model contains opaque identities and lifecycle facts, not:

- Secret or credential values;
- full Configuration or Snapshot payloads;
- authentication material;
- raw internal errors or stack traces;
- Host pointers or process-local capabilities.

Errors and inspection results must not disclose the existence or state of
another unauthorized aggregate. Concrete authorization and operational
reporting/redaction policy remain separate required designs.

## 20. Technology Neutrality

This contract can be implemented by any storage technology that proves its
atomicity, conditional revision, uniqueness, coherent-read, durability, and
failure requirements.

Terms such as aggregate, append, conditional publication, and revision are
semantic. They do not require relational tables, document storage, event
sourcing, compare-and-swap instructions, distributed consensus, UUIDs, or a
specific transaction isolation level.

An implementation may add private mechanics, but it must not expose them as
new architecture or weaken the contract.

## 21. Conceptual Operations

The design requires capability equivalent to these conceptual operations:

```text
AllocateCandidateIdentity
CreateRuntimeInstance
ReadRuntimeInstance
ReadLaunchAttemptHistory
ConditionalClaimLaunchAttempt
ConditionalBindExecutionGeneration
ConditionalPublishRunning
ConditionalClaimStop
ConditionalPublishTerminal
```

These names are explanatory, not API or Go interface definitions. Operations
take exact identity and expected revision where applicable and return a
committed revision or a truthful definitive/indeterminate outcome.

No generic CRUD repository, dynamic entity registry, universal transaction
manager, or service locator is authorized.

## 22. Acceptance Proofs

A future implementation must prove at minimum:

1. atomic complete Instance creation and immutable binding;
2. RuntimeInstanceID uniqueness within the management domain;
3. atomic single-active-attempt claim with exact version pin;
4. child-key uniqueness and non-reuse within complete Instance history;
5. append-only history across start failure, stop, and later attempts;
6. exact immutable execution-generation binding after claim and before Load;
7. binding mismatch, stale revision, or indeterminate inspection permits no
   external preparation;
8. exact conditional Running, Stop, and terminal publications;
9. stale and mismatched operations perform zero mutation;
10. concurrent same-Instance claims produce at most one accepted mutation;
11. different Instances can progress independently;
12. coherent reads correspond to one committed revision;
13. definitive failure publishes nothing;
14. indeterminate outcomes are resolved by exact identity/revision inspection,
    without blind new-ID retry;
15. persisted actual state or execution binding is never used as liveness proof
    after Owner loss;
16. redaction and domain isolation prevent cross-scope disclosure;
17. no second lifecycle owner, Host ownership, schema promise, or hidden
    service locator appears.

Proofs must include technically available concurrency, race, failure-injection,
durability, and restart-of-storage-client scenarios. They do not authorize
production activation.

## 23. Formal and Downstream ARCH-004 Section 19 Gates

This Approved design closes the focused architecture design gate for ARCH-004
section 19(2). Approved DP-015 through DP-018 close the downstream focused
design gates in sections 19(3)–(6). These status decisions authorize no
implementation by themselves.

Conditional aggregate revision prevents stale mutation but is not client
command idempotency. Inspecting an indeterminate write is not recovery or
reconciliation. Terminal fact storage is not operational reporting. Approved
[DP-015](DP-015-runtime-management-command-idempotency.md),
[DP-016](DP-016-runtime-activation-replacement-rollback.md),
[DP-017](DP-017-runtime-recovery-reconciliation.md), and
[DP-018](DP-018-runtime-operational-error-reporting-redaction.md) define those
separate responsibilities without implementing them.

## 24. Explicit Deferrals

Deferred to focused designs or implementation tasks:

- command keys, deduplication windows, replay results, and caller retry policy;
- choosing, activating, replacing, or rolling back a version;
- hydration of Owner after restart and reconciliation with execution evidence;
- diagnostic taxonomy, redaction policy, audit, metrics, and alerting;
- create/list/delete management APIs and concrete authorization;
- storage product, schema, migrations, serialization, backup, and retention;
- automatic restart, scheduling, process supervision, and clustering.

None may be smuggled into identity allocation, aggregate revision, history, or
conceptual operations.

## 25. Implementation Boundary

Implementation Status is Planned. The repository has the isolated process-local
Lifecycle Owner, launch flow/source components, and DP-013 management routing.

No durable operational identity repository, schema, package, adapter, API,
hydration, recovery, management wiring, or Production Activation exists.
Approval does not create or authorize any of them. The isolated DP-013 package
does not change this boundary; integration remains blocked until the required
persistence and downstream dependencies exist.

## 26. Decision

UWP will persist operational identity as one Runtime Instance aggregate with
immutable Workspace/Configuration/Instance binding, a monotonic conditional
revision, last Owner-confirmed desired and actual facts, at most one active
attempt, and complete append-only Launch Attempt history.

Each Launch Attempt is an owned child keyed by
`(RuntimeInstanceID, LaunchAttemptID)`, pins exactly one Published
ConfigurationVersion, and never reuses its child identity within the Instance
history. RuntimeInstanceID is unique within the operational management domain.

Before external preparation, a claimed attempt may conditionally receive one
immutable opaque execution-generation binding owned by DP-014. The planned
DP-011 continuation coordinates the binding capability, DP-016 defines the
final binding/load gate, and DP-017 consumes the binding during recovery. The
binding is retained as correlation history and never proves liveness or
shutdown.

The Runtime Lifecycle Owner remains the sole lifecycle and live Host owner.
Atomic persistence records truthful facts but does not prove liveness after
Owner loss. Stale operations perform zero mutation; indeterminate outcomes
require exact identity and revision inspection and never blind new-ID retry.
The design chooses no schema, technology, API, or implementation.

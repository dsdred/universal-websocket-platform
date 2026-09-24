# DP-022: Runtime Execution Containment and Evidence

[Russian version](../../ru/design/DP-022-runtime-execution-containment-and-evidence.md)

## 1. Status

- **Design Status:** Approved
- **Implementation Status:** Planned

This proposal defines the execution-containment and evidence boundary that
Approved DP-017 section 11 requires before recovery reconciliation can be
implemented. It is Approved as that boundary, so the DP-017 section 11
prerequisite for an approved containment boundary is satisfied at design level.
That says nothing about implementation, which remains absent: DP-017
implementation stays unactivated.

[DP-023](DP-023-runtime-process-containment-bootstrap.md) is the Approved,
Implemented-in-isolation boundary for the initial capability/ledger/generation
bootstrap. TASK-069 implements that Windows-only substrate in its worktree.
The latest verification, review and Acceptance checkpoint and subject identity
resolve only from TASK-069's newest valid matching envelope and are not
duplicated here. The slice does not implement this proposal's evidence outcomes.

No evidence adapter, scanner, supervisor, production wiring, or composed
runtime behavior exists. Nothing here states or implies that Control Service
can already observe process termination.

## 2. Purpose

Define one bounded, technology-neutral answer to a question that Approved
DP-017 section 11 asks and no authoritative source answered before this
document: what makes an execution generation unique and current, and what
authority may prove that the exact generation named by an attempt's durable
execution binding has terminated.

The design must let a later replacement Control Service distinguish, for one
exact Runtime Instance, between execution that is still live, execution whose
containing generation is proven terminated, resources that are absent, a Host
shutdown contract that is proven complete, and truth it does not know. It must
also state precisely which observations are insufficient, so that no
implementation can substitute a probe, a clock, or a stored state for proof.

## 3. Authority

This proposal refines, without overriding:

- [ARCH-004](../architecture/ARCH-004-runtime-deployment-and-identity-model.md)
  sections 9, 11, 12, 14, 17, and 19 — single-node in-process topology, Host
  ownership, publication of `Stopped`, process-loss semantics, and the ban on
  process metadata as identity;
- [ADR-0003](../adr/0003-runtime-architecture.md) and the
  [ARCH-002](../architecture/ARCH-002-runtime-foundation-freeze.md) freeze for
  composition-root ownership and the absence of Host restart or in-place reload;
- [DP-011](DP-011-runtime-launch-pipeline-integration.md) and
  [DP-013](DP-013-runtime-management-routing.md) for the rule that the exact
  Control Service composition creates the opaque execution generation and that
  Directory, Flow, Owner, and Host neither allocate it nor persist its binding;
- [DP-014](DP-014-runtime-operational-identity-persistence.md) sections 10 and
  13 for durable attempt-to-generation binding and its explicit inability to
  prove liveness;
- [DP-016](DP-016-runtime-activation-replacement-rollback.md) sections 5, 10,
  12, and 22 for proven release, phase order, and the deferral of process
  inspection to recovery;
- [DP-017](DP-017-runtime-recovery-reconciliation.md) sections 5, 6, 10, 11,
  12, and 20 for the definitions of execution generation, execution binding,
  and execution evidence, for the fail-closed ordering, and for the authority to
  classify a reconciled set;
- [DP-019](DP-019-runtime-activation-orchestration-prerequisites.md),
  [DP-020](DP-020-runtime-orchestration-binding-sequence-readiness.md) section
  8.5, and [DP-021](DP-021-private-exact-scope-managed-start-invoker.md) for the
  existing request-exactly-once generation provider seam, binding sequence, and
  invocation custody.

DP-022 consumes those seams and never reallocates, redefines, or replaces them.
DP-017 section 12 remains the only authority that combines facts into a recovery
classification; DP-022 supplies facts and their guarantees. ARCH-004 section
19(5) is closed by Approved DP-017 and is not claimed or re-opened by this
document. Approved sources keep their status; a Draft cannot override them.

## 4. Scope

The design defines:

- the containment boundary, the resource classes it covers, and the containment
  guarantee it can claim in the initial single-node in-process topology;
- the containment capability that establishes one unique current execution
  generation per containment domain;
- generation issuance rules, permitted consumers, and reuse prohibition;
- the durable containment ledger and the facts it may and may not carry;
- the exact proof obligation that establishes termination of a named prior
  generation, and what that proof never establishes;
- the separate authority for Host-owned shutdown completion;
- a closed evidence result model, its `Unknown` variants, and their precedence;
- the list of observations that are not evidence and the forbidden inferences
  built from them;
- adapter guarantee levels, trust, scope isolation, security, cancellation, and
  concurrency boundaries;
- explicit deferrals and acceptance proofs.

## 5. Non-goals

The design does not define:

- recovery claim, permit, assessment, reconciliation, publication order, barrier
  reopening, or any DP-017 behavior;
- a storage engine, schema, transaction, migration, lock primitive, file layout,
  named object, signal, process table query, or identifier format;
- child-process or remote-worker protocols, supervision, adoption, forced
  termination, live-orphan handling, or cross-node quorum;
- restart, retry, backoff, failover, scheduling, watchdog, or self-replacement;
- public or internal API, DTO, transport mapping, health endpoint, or operator
  reporting and redaction, which remain DP-018 and ARCH-004 section 19(6);
- Control Service startup sequencing, deployment, process management, or
  Production Activation;
- any new Runtime lifecycle state, desired state, actual state, or category.

## 6. Terms

**Containment domain** is the durable scope inside which execution-generation
containment is decided. In the initial topology it is the operational management
domain served by one Control Service together with the durable identity state it
owns. A domain is not a Runtime Instance, Workspace, Configuration, or host
machine label.

For the initial process boundary, that durable scope includes one immutable
authoritative storage root and a pre-provisioned Domain-to-storage authority
binding. DP-023 section 8.1 defines its bootstrap and deployment trust
contract. Absence of storage is not evidence that a Domain is new.

**Containment capability** is the exclusively held authority to be the live
generation of one containment domain. It is acquired by one Control Service
process exactly once, held for that process's authoritative lifetime, and never
transferred, renewed, timed out, or shared.

**Containment boundary** is the set of execution resources whose lifetime is
bounded by the holder of the containment capability. Resources outside that set
are outside the containment guarantee.

**Execution generation** keeps its Approved DP-017 section 5 meaning: an opaque
identity for one execution-containment boundary, which in the initial topology
identifies one Control Service process generation. It is not a PID, timestamp,
liveness lease, Runtime identity, or command identity.

**Generation authority** is the composition-owned responsibility that holds the
containment capability, issued the current generation identity, and answers
closed evidence questions about generations of its own domain. It is not a
lifecycle owner, not a persistence component, and not a second source of attempt
truth.

**Containment ledger** is the durable, append-only record of which opaque
generation identities have held the containment capability of one domain and
which have been superseded by a later exclusive acquisition.

**Termination proof** is the derived, evidence-bearing conclusion that an exact
named generation other than the current one no longer holds the containment
capability and therefore has terminated.

**Shutdown-completion evidence** is the coherently read durable fact, owned by
DP-014 and DP-015, that the Runtime Lifecycle Owner completed the Host-owned
shutdown contract for one exact Launch Attempt within one exact generation.

**Unbound observation** is any signal that cannot be correlated to the exact
`(domain, generation, attempt)` tuple. It is not execution evidence.

**Unknown** is the family of results that prove nothing: absent, unavailable,
stale, scope-mismatched, contradictory, indeterminate, cancelled, or produced by
an adapter whose declared guarantees do not cover the case.

## 7. Containment boundary model

In the initial ARCH-004 topology the Runtime Host is composed inside the Control
Service process and cannot survive that process. The boundary is therefore the
process, and containment is a property of resource classes, not of a component
that watches processes.

1. **Covered class.** Resources whose lifetime the platform ends with the
   process: in-process memory and goroutines, in-process session state,
   listeners and sockets opened by that process, and handles it holds. For this
   class, proven termination of a generation is proof that those resources are
   no longer held or reachable.
2. **Uncovered class.** Anything that can outlive the process: child processes,
   resources another process holds on its behalf, or registrations in external
   systems. The initial topology defines no containment guarantee for this class
   and no adoption, termination, or cleanup protocol for it.

A Runtime component that acquires an uncovered resource while DP-017's live
orphan path is unapproved obtains evidence that the containment boundary cannot
interpret. The design therefore treats acquisition of uncovered resources as a
boundary violation: evidence for that generation resolves to `Unknown`, the
admission barrier stays closed, and the situation is escalated as an
architecture gap rather than repaired by inference. ARCH-004 section 11
prohibits simulating supervision through hidden state in the Runtime Host, and
this boundary does not create such state.

The containment guarantee is exclusively about existence and reachability of
resources. It never claims that a terminated generation cleaned up gracefully,
completed a stop, finished a command, or reached any lifecycle outcome.

## 8. Containment capability and one live generation

1. One containment domain has at most one live containment capability holder.
   Acquisition is exclusive, atomic, and either succeeds or fails; there is no
   shared, partial, or "probably held" acquisition, and acquisition is not
   re-attempted by a process that already holds or already lost it.
2. The composition acquires the capability once, before it opens command
   admission, before it accepts a management command, before any Launch Attempt
   is claimed, and before any execution binding is written. Ordering is
   normative: no generation exists and nothing may be bound before acquisition.
3. The holder never releases the capability while it continues to act as
   authoritative, and never re-acquires it in place. Release is performed only
   by the platform as part of process termination, and an adapter may declare
   that guarantee only if the release cannot be lost, delayed, or granted twice.
4. The capability is not a lease. It has no expiry, no renewal, no heartbeat, no
   clock, and no elapsed-time inference. No wall-clock observation contributes
   to any conclusion derived from it.
5. Failed, ambiguous, lost, or revoked acquisition fails closed: the process
   allocates no generation, persists no binding, opens no admission, invokes no
   lifecycle work, and reports the domain as unavailable. Loss during life
   closes admission and forbids further binding until a new process performs a
   new exclusive acquisition. Recovery assessment under DP-017 then governs;
   this design authorizes no automatic restart, self-replacement, or takeover.

Exclusive re-acquisition is the only mechanism by which this boundary establishes
that a prior generation is gone, and it is why the proof does not depend on
timing.

## 9. Generation issuance

1. Exactly one opaque execution generation identity is issued per successful
   capability acquisition, by the generation authority in the Control Service
   composition, consistent with DP-011, DP-013, and DP-014 section 10.
2. The identity is non-empty, unguessable, opaque, and immutable for the
   process's lifetime. It is never derived from, equal to, or composed of a PID,
   process name, executable or working path, address, port, clock value, boot
   identifier, host name, configuration version, Runtime Instance identity,
   Launch Attempt identity, or command identity.
3. An identity is never reused inside a domain while any durable record of it
   exists. The containment ledger enforces uniqueness: issuance of an identity
   already present in the ledger is a defect that resolves to `Unknown` and
   closes admission rather than silently superseding a generation.
4. The value offered to the DP-020 section 8.5 provider seam is exactly the
   current generation identity. The provider may be consulted at most once per
   winning claim, as DP-020 already requires; it must return the current
   identity or fail. It must never return another generation, a remembered
   generation, a reconstructed value, or a value minted after loss.
5. No consumer may cache, substitute, merge, or replace a generation. After the
   capability is lost, the process has no authoritative generation and issues no
   new one without a new acquisition, which this design forbids in place.

### 9.1 Permitted consumers

The generation authority supplies the identity only to the paths that already
own it: the DP-015/DP-019/DP-020 binding sequence, the DP-014 conditional
binding operation, and closed evidence reads. Flow, Lifecycle Owner, DP-013
Directory, Runtime Host, and recovery code never allocate it, never persist it
outside DP-014, and never expose it.

## 10. Durable attempt-to-generation correlation

DP-014 exclusively owns the durable immutable correlation of one exact Launch
Attempt to one execution generation, published conditionally inside the Runtime
Instance aggregate before Load, as DP-014 section 10 and DP-016 section 10
require. DP-022 adds no new durable correlation, no parallel record, and no
second write path. It only fixes what the correlation may and may not mean:

1. a binding may name only the generation the acquiring process currently holds;
2. a binding proves correlation only, exactly as DP-014 states, and never
   liveness, preparation, readiness, ownership, release, or shutdown;
3. a durable binding that names a generation with no containment-ledger entry in
   the same domain is a scope mismatch, not evidence, and resolves to `Unknown`;
4. a binding is the input that names the generation whose termination must be
   proven; it is never itself that proof;
5. ledger and binding must be read as one coherent domain-scoped set; if they
   disagree, or if either read is stale, partial, or contradictory, no
   termination conclusion is admissible.

## 11. Containment ledger

The ledger is the durable, append-only containment record of one domain. It
carries, per generation: its opaque identity, that it held the containment
capability, and that a later exclusive acquisition superseded it. It carries
nothing else.

1. Ownership: the generation authority is the only writer. DP-014 persistence
   remains the only durable store of aggregate, attempt, and binding facts.
   Recovery is a reader and never a writer.
2. Order: the ledger needs no timestamp, sequence clock, or elapsed-time
   comparison. Supersession is recorded as an explicit containment fact at
   acquisition time, so termination reasoning is purely relational: the current
   holder plus a recorded superseded entry for a different identity.
3. Prohibited content: lifecycle or desired/actual state, command state,
   outcome, attempt history, PID, address, port, path, credential, configuration
   or Snapshot payload, diagnostic text, or any field that would make the
   ledger a competing source of Runtime truth.
4. Durability and scope: ledger and attempt state must belong to the same
   containment domain. A ledger readable from more than one independent durable
   state, or a domain whose two candidate stores disagree, yields `Unknown`; the
   design does not resolve cross-store ambiguity by preference or recency.
5. Retention is not defined here; deletion or compaction that could erase the
   supersession fact of a generation still named by a durable binding is not
   permitted.
6. The ledger is the containment fact set that DP-017 section 11 requires in
   order to name an exact prior generation. It is not process supervision, PID
   persistence, scheduling, clustering, or crash recovery itself, all of which
   ARCH-004 section 11 keeps out of scope, and it stores no process identifier,
   address, or path of any kind.

For initial bootstrap, `ProvisionedEmpty` is an existing, validated
pre-provisioned anchor and ledger with no generation entry, as specified by
DP-023 section 8.1. A missing, replaced, relocated, stale, or alternate store
is not an empty ledger and grants no authority. The anchor is storage provenance
metadata outside generation records; it does not add Runtime facts to the
ledger. The trusted deployment/storage boundary prevents undetectable joint
anchor-and-ledger rollback or cloning; runtime validation alone cannot prove
that a consistent copy is original.

## 12. Termination proof

For the exact prior generation named by an attempt's execution binding,
`GenerationTerminated` is proven when and only when all of the following hold
for one coherent read of one domain:

1. the reading process currently holds that domain's containment capability,
   exclusively acquired in this process generation;
2. the ledger records the named prior generation as having held that capability;
3. the named prior generation differs from the current generation identity;
4. the capability in use cannot be held by two live generations and cannot be
   released except by process termination, as declared by the adapter guarantee
   level in section 17;
5. the generation authority reports no loss or revocation of its own
   acquisition.

This proof establishes only that no execution contained by that generation
remains live or reachable within the covered resource class. It does not
establish:

- when, why, or how cleanly the process ended, nor any crash classification;
- that any Host stopped, closed listeners, drained sessions, or ran a shutdown
  contract;
- that a Stop, replacement phase, or command completed or was satisfied;
- that resources of the uncovered class are absent;
- that the attempt may be terminalized, admitted, replayed, adopted, or
  restarted, which remain DP-017 and DP-016 decisions.

Self-proof is excluded: the current generation can never report itself
terminated, and no process may report termination for a generation it does not
record in this domain. Absence of a ledger entry is never a termination proof.

## 13. Shutdown-completion evidence

Host-owned shutdown completion is a different fact with a different producer.

1. It is established only by a durable terminal fact that the Runtime Lifecycle
   Owner itself published for the exact attempt within the exact bound
   generation, after the Host completed its owned shutdown contract, as DP-016
   and DP-017 section 13 already require.
2. Its admissibility as recovery evidence additionally requires `GenerationTerminated`
   for that same generation, so that the record cannot be read while the Host it
   describes may still be live.
3. No new store, record type, or second publication path is introduced. The
   durable attempt and command facts owned by DP-014 and DP-015 are the fact;
   DP-022 defines only when reading them is legitimate evidence.
4. A fact written by a recovery or reconciliation path is never shutdown-
   completion evidence. Recovery output cannot certify the thing it is asked to
   determine, so this design forbids that circular use explicitly.
5. Absence of the record is not evidence of non-completion. Termination plus no
   admissible shutdown-completion fact yields only the truthful
   resource-absence and interrupted outcomes DP-017 already prescribes, never
   `Stopped`.

Consequently `resource absence != Host-owned shutdown completion` is structural
here: the two conclusions have distinct producers, distinct proofs, and no
derivation between them.

## 14. Evidence query contract

Evidence is produced by one read-only capability equivalent to DP-017 section 24
`ReadExactExecutionEvidence`. Its semantic contract:

- input is the exact tuple `(containment domain, Runtime Instance, Launch
  Attempt, execution generation)`; an incomplete or partially resolved tuple is
  rejected as `Unknown(ScopeMismatch)` before any observation;
- output is exactly one result from the closed model in section 15;
- the operation performs no mutation, allocates or binds nothing, opens nothing,
  and grants no authority beyond reading;
- the result is bound to the tuple that produced it and is single-use: a later
  consumer must re-read, because a cached or replayed result is `Unknown(Stale)`;
- the query may be answered only by the generation authority of the domain it
  names, or by an adapter whose declared guarantee level covers that case;
- concurrent reads are permitted and independent; long reads hold no aggregate,
  command, Owner, or admission lock, and any downstream conditional publication
  still revalidates revisions under DP-017 section 16;
- there is no batch, scan, discovery, enumeration, or "nearest match" operation.

## 15. Closed evidence outcomes

| Result | Proves | Never proves |
| --- | --- | --- |
| `GenerationLive` | the queried generation is the reader's own current generation and its capability is held | that any earlier generation terminated; any lifecycle completion |
| `GenerationTerminated` | the exact named prior generation is superseded by an exclusive re-acquisition in this domain | graceful cleanup, Stop success, command outcome, shutdown completion |
| `CoveredResourcesAbsent` | no covered-class resource of the terminated generation remains held or reachable | uncovered-class absence; successful release by the application |
| `HostShutdownCompleted` | the Owner's durable terminal fact proves the Host's owned shutdown contract completed for the exact attempt in the exact terminated generation | that recovery may reopen admission, that a later phase may run, or any earlier readiness claim |
| `LiveUnownedExecution` | an adapter whose approved level reports it observed execution the current process does not own | manageability, adoption rights, identity of the owner, or a safe termination action |
| `Unknown(reason)` | nothing | any of the conclusions above |

`reason` is closed and one of: `Absent`, `Unavailable`, `Stale`,
`ScopeMismatch`, `Contradictory`, `Indeterminate`, `Cancelled`,
`UnsupportedTopology`, `GuaranteeNotDeclared`.

Precedence is deliberately minimal:

1. an unbound observation is not an input at all;
2. any tuple, ledger, or binding mismatch resolves to `Unknown(ScopeMismatch)`;
3. any two inputs that disagree resolve to `Unknown(Contradictory)`;
4. no ordering, recency, majority, confidence, or cost rule may select a
   winner between disagreeing facts;
5. `Unknown` is terminal for that read: it is never degraded into a guess and
   never upgraded by repetition.

`LiveUnownedExecution` is unreachable in the initial in-process topology, which
defines no adapter level able to report it; it is named so that a future adapter
cannot redefine containment silently, and it keeps the barrier closed per DP-017
section 11.

## 16. Insufficient observations

None of the following, alone or combined, is execution evidence or any part of a
proof: process identifier or process table entry, process or service name,
executable or working directory, listening address or port, successful bind,
successful connect, refused connect, elapsed or wall-clock time, clock skew,
boot identifier, host name, log line, health or readiness response, memory
dump, stored desired or actual `Running` or `Stopping`, persisted `Running`
refreshed at read time, command `Claimed`, in-flight callback, configuration or
Snapshot version, aggregate revision, Runtime Instance or Launch Attempt
identity, or the absence of any of these.

Forbidden inferences:

| Inference | Status |
| --- | --- |
| connect failed, therefore the prior generation terminated | forbidden |
| port is free, therefore Host resources are released and Stop succeeded | forbidden |
| process gone from the table, therefore shutdown contract completed | forbidden |
| elapsed time exceeded a threshold, therefore a permit, capability, or execution is lost | forbidden |
| durable binding exists, therefore execution is or was live | forbidden |
| durable state says `Running`, therefore the Host is manageable now | forbidden |
| no shutdown-completion record, therefore the Host did not complete shutdown | forbidden |
| a recovery publication exists, therefore its precondition was proven | forbidden |
| evidence is stale but favorable, therefore use the favorable reading | forbidden |

## 17. Adapter trust and guarantee levels

An adapter answers evidence queries only at a declared guarantee level, and may
not exceed it.

| Level | Covers | May report | Guarantees required |
| --- | --- | --- | --- |
| `None` | no adapter exists or it cannot bind the tuple | `Unknown` only | none; the default state of the repository |
| `ProcessContainment` | the initial single-node in-process boundary | `GenerationLive`, `GenerationTerminated`, `CoveredResourcesAbsent`, `HostShutdownCompleted` read through DP-014/DP-015 facts | exclusive single-holder acquisition; release only on process termination with no loss or double grant; no clock use; ledger locality and pre-provisioned storage authority per section 11 and DP-023 section 8.1 |
| `ExecutionIsolation` | a future approved child or remote boundary | additionally `LiveUnownedExecution` | all of `ProcessContainment` plus an approved adoption and termination protocol, which does not exist |

Every adapter must declare, per level: the domain it binds, how a result binds
to the exact tuple, what it can and cannot observe, the basis of its
release-on-termination guarantee, and its behavior under unavailability,
contradiction, partial read, panic, and cancellation. Any undeclared or
unverifiable property makes the level unusable and yields
`Unknown(GuaranteeNotDeclared)`. An adapter never mutates, never closes a
barrier, never classifies a recovery set, and never becomes an owner; the
generation authority that consumes it stays inside the Control Service
composition, so ADR-0003's single composition root and ARCH-002's freeze are
preserved.

## 18. Scope isolation and security

Evidence and ledger reads are scoped to exactly one containment domain — one
operational management domain served by one Control Service together with the
durable identity state it owns — and, within it, one Workspace, Configuration,
Runtime Instance, Launch Attempt, and execution generation. Cross-domain
evidence is prohibited even when it is available and favorable.

Results carry only opaque identities and closed semantic categories. They never
carry credentials, Secrets, configuration or Snapshot payloads, raw internal
errors, stack traces, host pointers, process-local permits, command callbacks,
or unrestricted process metadata. A reader that cannot prove the tuple binding
discards the input instead of narrowing it. Existence of an execution belonging
to another tenant is never disclosed, and concrete operator reporting and
redaction remain DP-018 and ARCH-004 section 19(6).

## 19. Cancellation and concurrency

1. Cancellation of an evidence read yields `Unknown(Cancelled)` and performs no
   mutation. It never releases a capability, never proves termination, never
   resolves a contradiction, and never authorizes a caller to proceed.
2. Cancellation of a lifecycle path is governed by DP-016 and DP-017; it does
   not change containment facts.
3. Acquisition is serialized by exclusivity itself: concurrent acquisitions in
   one domain yield exactly one holder and one failure, deterministically and
   without clocks or priorities.
4. Concurrent evidence reads are independent and never block admission; they
   never hold locks across the read, and any downstream conditional publication
   revalidates revisions under DP-017 section 16.
5. A concurrent reader never obtains mutation authority, and a writer that
   changes the tuple, ledger, or binding between read and use invalidates the
   read.

## 20. Failure matrix

| Situation | Truthful result | Forbidden consequence |
| --- | --- | --- |
| acquisition fails or is ambiguous | no generation, no binding, no admission, domain unavailable | proceed with an assumed generation |
| acquisition succeeds, ledger write indeterminate | `Unknown(Indeterminate)` | bind or admit on partial containment state |
| capability lost or revoked during life | admission closed, no new binding, no in-place re-acquisition | continue serving, self-heal, or restart automatically |
| named generation has no ledger entry | `Unknown(ScopeMismatch)` | treat absence as termination |
| ledger entry equals the current generation | `GenerationLive` | report self-termination |
| ledger and current holder differ and exclusivity guarantees hold | `GenerationTerminated` | infer cleanup or Stop success |
| adapter cannot prove release-on-termination | `Unknown(GuaranteeNotDeclared)` | fall back to clock, PID, or probe |
| two reads disagree | `Unknown(Contradictory)` | pick recent, majority, or convenient |
| cached read reused later | `Unknown(Stale)` | reuse as current truth |
| query cancelled | `Unknown(Cancelled)` | default to favorable outcome |
| binding names a generation of another domain | `Unknown(ScopeMismatch)` | cross-domain inference |
| Host absent, no shutdown-completion fact | covered-resource absence only | publish or infer `Stopped` |
| Owner durable terminal fact plus termination proof | `HostShutdownCompleted` | skip either half of the proof |
| durable terminal fact written by recovery | not evidence | circular certification |
| future adapter reports unowned live execution | `LiveUnownedExecution` | adoption, replay, or forced termination |
| uncovered-class resource may exist | `Unknown` and barrier stays closed | assume containment covers it |
| process table says the old PID is gone | `Unknown` | termination proof |
| port refuses connections | `Unknown` | resource release or successful Stop |

## 21. Ownership summary

| Fact | Owner | DP-022 role |
| --- | --- | --- |
| containment capability and its exclusivity | Control Service composition (generation authority) | defines semantics and failure behavior |
| current generation identity | generation authority | defines issuance and reuse rules |
| durable attempt-to-generation binding | DP-014 | constrains admissible values and meaning |
| containment ledger | generation authority as writer, recovery as reader | defines content and prohibitions |
| live Host ownership and lifecycle decisions | Runtime Lifecycle Owner (ARCH-004 section 9) | untouched; evidence never transfers it |
| Host resources | Runtime Host | untouched; containment only bounds their lifetime |
| recovery classification, claim, permit, barrier | DP-017 | supplies facts and guarantees only |
| command and attempt durable outcomes | DP-014, DP-015, DP-016 | defines when they count as evidence |
| operator reporting and redaction | DP-018, ARCH-004 section 19(6) | explicitly out of scope |

## 22. Technology neutrality

Containment capability, generation identity, ledger, supersession, and closed
evidence results are semantic requirements. The design does not select or
require a database, file lock, PID file, named object, semaphore, lease service,
consensus or quorum mechanism, clock, boot identifier, process supervisor,
operating-system primitive, container or VM facility, vendor, or identifier
format. It does not authorize a generic registry, service locator, process
manager, or background scanner, and it adds no dependency or package.

An implementation may choose any mechanism that proves exclusive single-holder
acquisition, release only on process termination, durable append-only
supersession, exact tuple binding, and fail-closed `Unknown` behavior. A
mechanism that cannot prove one of these is not a partial implementation of this
design; it yields `Unknown` by definition.

## 23. Explicit deferrals

Deferred to separate approved designs or implementation tasks:

- child-process, remote-worker, and container adapters, including any live-
  orphan detection, adoption, and termination protocol;
- process supervision, scheduling, clustering, cross-node quorum, and any
  multi-holder domain;
- concrete storage engine, schema encoding, migrations, retention, compaction,
  and deployment layout beyond the guarantee class and atomic bootstrap
  approved by DP-023;
- Control Service startup sequencing, signal handling, shutdown orchestration,
  production wiring, and Production Activation;
- recovery claim, permit, assessment, reconciliation, and barrier mechanics,
  which DP-017 owns;
- public or internal API, DTO, status mapping, authorization policy, health
  endpoint, and operator reporting and redaction;
- metrics, logging, tracing, alerting, and retention;
- automatic restart, retry, backoff, failover, and remediation policy.

## 24. Acceptance proofs

A future implementation of this boundary must prove at minimum:

1. at most one live Control Service process holds the capability for one domain,
   under concurrent acquisition;
2. the current generation is issued exactly once per successful acquisition,
   before admission, attempt claim, and binding;
3. no generation identity is derived from or equal to PID, address, port, clock,
   host name, configuration version, or any Runtime or Attempt identity;
4. identity reuse inside a domain is detected and resolves to `Unknown`, with
   admission closed;
5. acquisition failure or ambiguity produces no binding, no admission, and no
   lifecycle call;
6. capability loss closes admission and cannot be repaired in place;
7. no durable fact, probe, log, port result, or elapsed time alone yields any
   positive result of section 15;
8. a terminated prior generation is proven only through exclusive
   re-acquisition plus a matching ledger entry;
9. termination proof never yields `HostShutdownCompleted`, cleanup, Stop
   success, or a command outcome;
10. covered-class absence is reported without claiming uncovered-class absence;
11. `HostShutdownCompleted` requires the Owner's durable terminal fact and the
    termination proof of the same exact generation;
12. a recovery-written durable fact is never accepted as shutdown evidence;
13. a binding naming a generation absent from the domain ledger yields
    `Unknown(ScopeMismatch)`;
14. contradictory, stale, cancelled, unavailable, and unsupported-topology
    inputs each yield their exact `Unknown` reason and no mutation;
15. evidence reads perform no mutation, open no admission, and grant no
    authority;
16. concurrent readers never serialize or bypass DP-017 revision revalidation;
17. no adapter below `ProcessContainment` can produce a positive result;
18. the ledger carries none of the prohibited fields of section 11;
19. cross-domain and cross-Workspace evidence is never used, and results
    disclose no prohibited payload;
20. EN/RU contract, matrices, outcome names, guarantee levels, and Planned
    status remain aligned.

## 25. Implementation boundary

Implemented in isolation today: an opaque
`ExecutionGeneration` identity type, the DP-014 conditional
attempt-to-generation binding operation, the DP-020 request-exactly-once
provider seam, and TASK-069's Windows-only DP-023 bootstrap package slice. The
latter establishes its private capability, ledger, and generation authority
only inside that package. The latest verification, review and Acceptance
checkpoint resolves only from TASK-069's newest valid matching envelope. There
is no evidence adapter, production composition wiring, or code path that can
expose termination evidence to Control Service.

DP-023 is Approved/Implemented in isolation by explicit Coordinator status
decision through TASK-069. Mutable role verdicts and identities resolve from
that task's newest valid matching envelope. Every evidence and downstream gate
remains unchanged; DP-022 itself stays Planned.

This document is an Approved design boundary, so DP-017 section 11 now has an
authoritative containment boundary to consume; DP-017 itself stays
Approved/Planned and unactivated. The status came from an explicit decision
through the project's design status process; Documentation, Tester, Reviewer, or
Coordinator acceptance of a task does not raise this document's Design Status
and never raises Implementation Status. An isolated bootstrap candidate does
not activate evidence: no evidence adapter or composition exists, so the exact
prior-generation termination proof that DP-017 section 11 requires still cannot
be consumed by any component, and DP-017 recovery, DP-018 reporting, production
integration, and Production Activation remain `Not Activated` and absent, and
downstream consumption of containment evidence is a later, separately approved
boundary. No ARCH-004 section 19 gate is claimed or re-opened here.

## 26. Decision

UWP will make execution containment an exclusivity property of one Control
Service process per containment domain. The composition acquires one exclusive
containment capability exactly once, issues one opaque execution generation
under it, and never releases that capability while authoritative. Termination of
the exact prior generation named by a durable attempt binding is proven only by
exclusive re-acquisition recorded as supersession in a durable containment
ledger — never by a clock, a lease, a process table entry, a PID, an address, a
port, a probe, or a stored lifecycle state.

Host-owned shutdown completion is proved by a different fact, produced by the
Runtime Lifecycle Owner and durable through DP-014 and DP-015, and admissible as
evidence only together with the termination proof of that same generation.
Resource absence therefore never proves shutdown completion, and absence of a
record never proves non-completion. Every other case is `Unknown`, and `Unknown`
adopts nothing, replays nothing, admits nothing, and invents no terminal truth.
Child-process and remote containment, adoption, supervision, and production
wiring remain separate approved decisions.

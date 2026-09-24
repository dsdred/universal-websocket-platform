# DP-023: Runtime Process-Containment Bootstrap

[Russian version](../../ru/design/DP-023-runtime-process-containment-bootstrap.md)

## 1. Status

- **Design Status:** Approved
- **Implementation Status:** Implemented in isolation

TASK-068 approves this implementation boundary and decomposition. Approval makes
one later code slice Ready for a separate intake; it does not create a package,
adapter, ledger, generation authority, positive evidence, or production
capability.

TASK-069 implements the isolated Windows-only first slice in the worktree:
independent trusted descriptor validation,
existing-only pre-provisioned anchor/bbolt ledger, a non-inheritable Windows
`Global\\` Event, exact successor append/inspection, identity checks, and
fail-closed tests. Coordinator explicitly set this implementation status
through TASK-069. Mutable Tester/Reviewer verdicts and subject identities
resolve only from that task's newest valid matching envelope and are not
duplicated here; the latest verification, review and Acceptance checkpoint is
envelope-owned. The package is not wired into Control Service; evidence reading, recovery,
reporting, provisioning, and Production Activation remain absent.

## 2. Purpose

Define the smallest implementable bootstrap that can honestly claim the
`ProcessContainment` guarantee required by DP-022 and later consumed by DP-017.
The bootstrap must establish one authoritative process, one fresh opaque
execution generation, and one durable supersession transition as a single
safety behavior before any Runtime management work is admitted.

## 3. Authority and precedence

This proposal refines, but does not replace:

- ARCH-004 ownership, lifecycle, serialization, and activation gates;
- ADR-0003 fail-closed lifecycle semantics;
- DP-014 attempt-to-generation binding;
- DP-015 command truth and DP-016 orchestration;
- DP-017 recovery ownership and ordering;
- DP-022 containment and exact-evidence outcomes.

If this document conflicts with an Accepted ADR, Active architecture guide, or
an Approved ownership rule above, that higher-precedence source wins and the
implementation stops for a new decision.

## 4. Scope

This design owns only:

- the initial local process-containment guarantee class;
- one safety-atomic acquisition, generation, and ledger bootstrap;
- capability retention and fatal fencing rules;
- the private package and adapter boundary;
- executable conformance requirements for the first implementation slice.

It defines no evidence reader, recovery action, public API, production wiring,
or activation.

## 5. Non-goals

- choosing a storage vendor, database product, or named operating-system lock;
- release, unlock, transfer, lease, renewal, expiry, or in-place reacquisition;
- PID, process-name, address, port, health, wall-clock, or elapsed-time probes;
- attempt binding, lifecycle mutation, command mutation, or recovery mutation;
- `ReadExactExecutionEvidence` or shutdown-completion evidence composition;
- child, remote, container, cluster, adoption, or supervision protocols;
- ledger retention, compaction, migration, deletion, or arbitrary repair;
- implementation of production domain provisioning or production startup wiring;
- Control Service integration, public DTO, authorization, or Production
  Activation.

## 6. Initial guarantee source

The initial implementation guarantee is:

> A stable-domain, host-kernel-enforced, process-scoped exclusive capability,
> retained privately for the authoritative Control Service lifetime and
> automatically released only by process termination, paired with a
> crash-consistent, single-writer, durable append-only containment ledger
> anchored in the same stable domain namespace.

The guarantee requires all of the following:

1. exclusivity is enforced outside ordinary application memory;
2. a second live process cannot acquire the same domain capability;
3. application code cannot explicitly release or transfer the capability;
4. retention is strong until process exit and does not rely on a finalizer or a
   caller-owned handle;
5. the namespace cannot silently resolve to two independent authorities;
6. ledger commit is crash-consistent and inspectable as exactly committed or
   absent;
7. the capability holder is the only configured ledger writer for the domain;
8. guarantee, namespace, storage, corruption, or inspection uncertainty grants
   no authority and no positive evidence.
9. the domain's authoritative storage root and pre-provisioned anchor are
   immutable for that domain; loss or replacement cannot be mistaken for first
   use.

A database session lock, file marker, PID file, expiring lease, or process-local
mutex is insufficient unless its concrete contract and subprocess tests prove
every property above.

## 7. Ownership

| Fact or authority | Sole owner |
| --- | --- |
| Exclusive containment capability | `runtimecontainment` generation authority |
| Current execution generation | `runtimecontainment` generation authority |
| Append-only generation/supersession ledger | same authority, as sole writer |
| Attempt-to-generation binding | DP-014 / `runtimeidentity` |
| Host and lifecycle decisions | Runtime Lifecycle Owner |
| Command truth | DP-015 command boundary |
| Recovery classification, claim, and barrier | DP-017 |
| Operator projection | DP-018 |

The ledger is a separate containment fact set. It is not a second Runtime
aggregate, lifecycle store, command store, attempt store, or recovery store.

## 8. Containment domain and identities

A containment domain is a stable operational namespace for one initial Control
Service authority. Domain identity and execution-generation identity are opaque,
non-zero, non-reusable values. Neither may be derived from PID, time, address,
hostname, process name, or mutable Runtime state.

The same normalized domain identity selects both the exclusive capability and
its ledger. Namespace mismatch, aliasing, replacement, or uncertain resolution
is fatal; implementation must not continue under a newly resolved namespace.

### 8.1 Provisioned domain and storage authority

An authoritative immutable storage root is part of provisioning the containment
domain. Trusted deployment provisioning binds one normalized Domain to one
storage authority and an existing domain anchor/ledger in that root before
Runtime bootstrap. The anchor records the Domain and an opaque, non-reused
storage-authority identity as namespace/provenance metadata; it is not a
generation record or a second source of Runtime truth. A canonical path derived
from Domain inside the root is only a locator, not proof that the located
storage is the original authority. Changing the root or reprovisioning the same
Domain does not create legitimate first use.

Runtime bootstrap receives an immutable trusted provisioning descriptor from
the deployment boundary, independently of the candidate root, anchor, ledger,
or any bytes stored beneath that root. The descriptor binds the normalized
Domain to the expected canonical root, the expected physical root identity, and
the expected opaque storage-authority identity. Candidate storage supplies only
observed values; it must never supply, derive, override, or repair any expected
descriptor value and cannot attest its own authority. Bootstrap validates the
observed canonical and physical root identities, anchor Domain and
storage-authority identity, and ledger binding against the descriptor before
`ProvisionedEmpty` or authority is possible. An alternate or copied store fails
closed even when its anchor and ledger are internally consistent.

`ProvisionedEmpty` means that the trusted anchor and ledger already exist and
validate as the provisioned authority, while the ledger has no generation
entry. It is the sole state from which a first generation with no predecessor
may be committed. A missing anchor, ledger, or directory is never
`ProvisionedEmpty`. Runtime opens the provisioned storage existing-only; it
does not create, repair, migrate, relocate, replace, or rebind a Domain, anchor,
root, or ledger.

The deployment/provisioning authority must preserve the Domain-to-root and
storage-authority binding across restarts and prevent unauthorized deletion,
replacement, cloning, relocation, and rollback of the anchor and ledger as one
unit. It must never reissue the same storage-authority identity or treat loss
of a previously provisioned Domain as a new Domain. Bootstrap may validate
observable root, anchor, Domain, authority-identity, and ledger mismatches; it
cannot detect a perfectly consistent joint rollback or clone by reading only
the rolled-back/duplicated storage. The declared `ProcessContainment` guarantee
therefore depends on the trusted deployment/storage boundary preventing that
undetectable case. An environment unable to establish this boundary cannot
declare the guarantee or expose authority. Production provisioning and wiring
are outside the first implementation slice.

Missing, mismatched, replaced, relocated, stale, or alternate candidate
storage, and uncertainty about which root is authoritative, fail closed. A
pre-acquisition failure grants no authority; after capability acquisition the
domain is `FatalFenced` and the process terminates. Neither case permits a
fallback root or automatic reprovisioning.

## 9. Safety-atomic bootstrap

Capability acquisition, fresh generation creation, and durable ledger
transition are one indivisible safety behavior even though kernel capability
and durable append need not share a physical transaction.

Authority exists only after all conditions hold:

1. the process exclusively holds the domain capability;
2. the existing provisioned anchor and ledger have been opened and validated
   under that capability in the authoritative storage root;
3. one fresh opaque generation candidate has been created;
4. the exact successor entry is durably committed while capability remains
   held;
5. exact inspection confirms that candidate as the current ledger tail.

The durable successor commit is the logical linearization point. Raw capability
acquisition and a private candidate are provisional and grant no generation
authority.

## 10. Ledger transition protocol

For previous tail `Gprev` and fresh candidate `Gnew`, the only successful
transition is:

```text
expected tail = Gprev
append current generation = Gnew
record predecessor = Gprev
thereby record Gprev superseded
```

The first generation has no predecessor only when the expected state is the
validated `ProvisionedEmpty` state. Append compares the exact expected tail or
that explicit empty state. A committed generation is never rewritten, removed,
reordered, or reused.
Only the exact candidate may be retried after an inspected definite absence.

## 11. Crash and indeterminate cuts

| Cut | Required result |
| --- | --- |
| before exclusive acquisition | no mutation, generation, or authority |
| acquisition fails or is ambiguous | domain unavailable; no ledger write or downstream work |
| provisioned anchor/ledger absent, mismatched, replaced, relocated, stale, alternate, or uninspectable | no first-use inference; fail closed, and fatal-fence if capability is held |
| capability held before candidate creation | provisional only; crash leaves no durable generation |
| candidate created before append | candidate remains private and unbound |
| definite append failure and inspected tail unchanged | retry the exact candidate while capability remains held |
| append outcome indeterminate | inspect first under held capability; never mint another candidate |
| exact candidate is current tail | converge to success |
| candidate absent and exact prior tail unchanged | retry only the exact append and candidate |
| different tail, partial/corrupt ledger, or unavailable inspection | fatal fence for the domain and terminate |
| crash after commit before authority is returned | next process may append a successor; no attempt was bound |
| crash after authority is returned | process termination releases capability; next acquisition appends a successor |
| live capability loss, revocation, or namespace uncertainty | close authority, never reacquire in place, and terminate |
| caller cancellation after provisional acquisition | reconcile the exact transition or fatal-fence; never release or transfer |

A post-acquisition fatal path must not return the process to ordinary service.

## 12. Capability retention and fatal fencing

The authoritative process retains the capability in private strongly reachable
state for its lifetime. No consumer receives the raw handle. There is no normal
cleanup operation that releases it.

Loss, revocation, guarantee downgrade, namespace ambiguity, unreconciled append,
storage-authority uncertainty, or corruption permanently changes the
process-domain state to `FatalFenced` after acquisition.
Fencing closes admission, disables generation provision and positive evidence,
and requires process termination. The process cannot reacquire authority in
place.

## 13. Ledger model

Each record contains only:

- the containment domain identity;
- the opaque current generation identity;
- the optional exact predecessor identity.

It contains no PID, time, address, Host state, lifecycle phase, command,
attempt, recovery claim, operator annotation, payload, or user data. The ledger
is append-only, single-writer, domain-isolated, and durable across process
termination. Loss, partial state, duplicate identity, impossible predecessor,
or unverified tail is unavailable/fatal—not `ProvisionedEmpty` and not positive
termination evidence.

## 14. Conceptual API

The public package semantics are deliberately narrow:

```text
AcquireProcessContainment(domain, trustedProvisioningDescriptor) -> ActiveAuthority | Unavailable | FatalFenced

ActiveAuthority.CurrentGeneration() -> exact committed current generation
ActiveAuthority.GuaranteeLevel() -> ProcessContainment
ActiveAuthority.IsAuthoritative() -> true only while unfenced
```

Private driver semantics are:

```text
AcquireExclusiveProcessLifetime(domain)
OpenExistingProvisionedAnchor(heldCapability, domain, trustedProvisioningDescriptor)
ReadLedgerTail(heldCapability, domain)
AppendSuccessor(heldCapability, domain, expectedTail, exactCandidate)
InspectExactAppend(heldCapability, domain, exactCandidate)
```

No surface exposes release, unlock, transfer, renew, expiry, arbitrary ledger
mutation, PID lookup, evidence classification, attempt binding, lifecycle,
command, or recovery mutation.

## 15. Package and dependency boundary

The first implementation target is:

```text
internal/runtimecontainment
```

It owns opaque domain/generation types, invalid-zero semantics, the capability
state machine, existing-only anchor validation, ledger records and
expected-tail append, candidate creation, guarantee identity, fatal fencing,
one concrete initial local adapter, and its subprocess/crash conformance
harness.

Dependency direction is:

```text
OS and durable adapter
        ↓
internal/runtimecontainment
        ↓ later composition adapter
existing ProvideExecutionGeneration seam
        ↓
runtimeactivation and DP-014 binding path
```

`runtimecontainment` imports no runtimeactivation, command, lifecycle, recovery,
HTTP, or reporting package. Downstream packages neither allocate nor replace its
generation.

## 16. Admission and composition boundary

Before authoritative bootstrap completion, composition must not open management
admission, accept a state-changing command, provide a generation, bind an
attempt, load a Runtime, or invoke lifecycle work. Wiring the committed
generation into the existing provider seam and proving no bypass are a later
slice, not part of the bootstrap package.

Consequently, Slice 1 proves local authority semantics but does not claim that
the current Control Service composition already enforces them.

## 17. First implementation slice

The first implementation slice is **Runtime Process-Containment Bootstrap over
a Pre-Provisioned Domain Anchor**:

- one `internal/runtimecontainment` package;
- one real local adapter satisfying section 6;
- existing-only open and validation of the pre-provisioned domain anchor and
  ledger under the trusted storage authority in section 8.1;
- the safety-atomic protocol in sections 9–11;
- fatal fencing from section 12;
- subprocess, restart, crash-cut, concurrency, durability, identity-reuse, and
  corruption proofs, including missing/deleted anchor, alternate root,
  replacement, relocation, stale copy, and two candidate stores.

Evidence reading, DP-014 changes, provider wiring, production provisioner,
recovery, reporting, public API, production integration, and activation remain
explicit non-goals.

## 18. Proof matrix

| Requirement | First-slice coverage |
| --- | --- |
| DP-022 proof 1: at most one live holder | direct concurrent-subprocess proof |
| proof 2: one generation per acquisition | bootstrap direct; admission/binding deferred to composition |
| proofs 3–4: opaque unique identity and reuse fencing | direct structural and restart proofs |
| proof 5: failure grants no authority | direct; downstream no-work proof deferred to composition |
| proof 6: loss closes authority with no in-place recovery | direct |
| proof 7: probes/time cannot establish evidence | direct API absence; positive evidence deferred |
| proof 8: termination substrate | direct exclusive reacquisition plus ledger; classification deferred |
| proofs 9–16 | deferred to evidence and recovery slices |
| proof 17: lower guarantee is never positive | direct adapter conformance |
| proof 18: forbidden ledger fields absent | direct structural proof |
| proof 19: domain isolation | direct for ledger; cross-Workspace evidence deferred |
| DP-017 section 11 and proofs 7, 11–13 | compositional prerequisite only; no recovery claim |

Required scenarios include N concurrent processes for one domain with exactly
one winner, independent domains, forced termination, one successor per restart,
every crash cut, indeterminate append inspection, committed-but-unacknowledged
append convergence, corruption and namespace mismatch, identity reuse rejection,
absence of release/reacquire API, race checks, and repository regression tests.
The harness provisions the anchor before invoking bootstrap and proves a first
generation only from validated `ProvisionedEmpty`; missing/deleted anchor,
alternate root, mismatched identity, replacement, relocation, stale copy, and
two candidate stores fail closed. Tests of observable mismatch do not claim to
detect an undetectable joint rollback/clone; the adapter must declare and
verify its deployment/storage trust precondition before claiming the guarantee.
Tests also provide expected canonical root, physical root identity, and
storage-authority identity only through the independent trusted provisioning
descriptor. A candidate anchor or store that copies or self-reports matching
expected values cannot replace that descriptor; alternate/copied storage fails
closed against it even when internally consistent.

## 19. Ordered downstream decomposition

1. DP-023 process-containment bootstrap.
2. DP-022 exact generation evidence reader.
3. DP-022 shutdown-completion evidence composition.
4. containment composition/admission/provider gate.
5. DP-017 read-only recovery assessment.
6. DP-017 durable recovery claim and admission barrier.
7. DP-017 attempt and primitive-command reconciliation.
8. DP-017 linked phase and parent reconciliation.
9. DP-017 coherent release and barrier reopening.
10. DP-018 reporting, then production integration and Production Activation.

Every later slice requires a fresh intake and prerequisite check. This order
activates none of them.

## 20. Size Guard

Verdict: **ACCEPT — ONE INDIVISIBLE ATOMIC IMPLEMENTATION BEHAVIOR**.

The target is one package and one independently shipped behavior. Capability,
ledger, and generation issuance are inseparable internal parts of the same
safety transition. Evidence reading and every consumer are split out. A second
adapter family, evidence reader, wiring, or recovery behavior requires a split;
crossing the normal production-line threshold requires re-evaluation, not
exposure of a partial authority.

## 21. Implementation boundary

Implementation Status is `Implemented in isolation`. The TASK-069 worktree
contains the Windows-only `internal/runtimecontainment` slice with a conforming
capability, pre-provisioned anchor/bbolt ledger, authoritative generation
bootstrap, unsupported-platform fail-closed stub, and focused proofs. It is not
composed into Control Service and has no positive evidence reader. Therefore
current production composition behavior and DP-017 are unchanged; evidence,
recovery, reporting, provisioning, integration, and Production Activation
remain later work.

## 22. Decision

UWP will establish initial `ProcessContainment` through one process-lifetime
exclusive capability and one same-domain durable append-only ledger in a
trusted, pre-provisioned, immutable storage authority. The first implementation
slice must acquire the capability, validate the existing anchor and ledger,
create one opaque generation, and durably append and inspect its exact successor
record before exposing authority. The first generation requires validated
`ProvisionedEmpty`; missing or uncertain storage is never first use. Any
ambiguous guarantee or unreconciled state fails closed; after acquisition it
fatal-fences and terminates the process. Evidence, composition, recovery,
reporting, provisioning implementation, and activation remain later work.

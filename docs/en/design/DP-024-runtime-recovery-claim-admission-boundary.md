# DP-024: Runtime Recovery Claim and Admission Boundary

[Russian version](../../ru/design/DP-024-runtime-recovery-claim-admission-boundary.md)

## 1. Status

- **Design Status:** Draft
- **Implementation Status:** Planned

TASK-079 records a proposed private implementation boundary. Architecture
Confirmation authorizes documenting this proposal, not Approval or code.
DP-017 remains Approved / Planned overall with only TASK-075 read-only assessment
implemented in isolation. No durable claim, recovery permit or barrier exists.

## 2. Purpose and authority

Resolve the four TASK-078 gaps for DP-023 section 19 item 7: joint ordering,
authentic prior-authority loss, durable claim/readback/resume, and claim-aware
assessment. The proposal refines [DP-017](DP-017-runtime-recovery-reconciliation.md)
sections 7–9, 16–19 and 23 under
[ARCH-004](../architecture/ARCH-004-runtime-deployment-and-identity-model.md).
[DP-014](DP-014-runtime-operational-identity-persistence.md) owns aggregate and
attempt truth; [DP-015](DP-015-runtime-management-command-idempotency.md) owns
command/parent/phase truth. [DP-022](DP-022-runtime-execution-containment-and-evidence.md)
and [DP-023](DP-023-runtime-process-containment-bootstrap.md) own containment
and generation evidence. This Draft overrides none of those sources.

No Approved-source amendment or ADR is proposed: this is a replaceable private
realization of existing ordering and ownership requirements, without a permanent
technology choice. Changing those requirements needs separate architecture work.

## 3. Scope and non-goals

The proposed initial realization is single-node, in-process and restart-based.
It defines coordination semantics, owner participation, private provenance,
closed admission, claim issuance/resume and read-only observation. It does not
select a DB, schema, adapter, identifier format or production provisioner.
It implements no lifecycle/command repair, claim release, API, discovery,
reporting, automatic restart, production integration or activation.

Same-process loss of a client, Owner, stack or permit stays unresolved. A future
revocation-and-drain protocol must independently prove every former capability
cannot survive before same-process recovery can be supported. NewBoundary,
cancellation, timeouts or an empty live map are insufficient.

## 4. Ownership and exact scope

The exact scope is operational/containment domain, Workspace, Configuration and
Runtime Instance. One composition-private per-Instance ordering participant is
explicitly supplied to DP-014, DP-015 and recovery owners. Composition must supply
the same participant for all writers of that scope, never create independent
gates for the same Instance. Different Instances have independent participants.

The participant owns only ordering and a local admission fence. DP-014 and
DP-015 keep their records, validation and conditional publications. Recovery
owns claim/barrier coordination and its private non-transferable permit.
Composition supplies current containment authority and prior-process-loss proof.
No raw map/lock, generic transaction registry, service locator, arbitrary callback
or ordinary lifecycle capability crosses this boundary.

## 5. Joint ordering protocol

Every DP-014 conditional writer and every DP-015 primitive/managed claim,
candidate-to-claim path, tracked-Start Stop exception, parent with preclaimed
StopOld, later-phase creation/delegation gate and ordinary terminal publication
enters the same short region. Recovery claim/resume and future release also
participate. One bypass invalidates the joint-ordering proof.

Acquisition order is outer exact-scope participant, then owner-local locks;
reverse entry is forbidden. Only closed owner-defined bounded validation and
publication operations execute inside. Authorization, evidence queries and
Owner/Flow/Load/Host work run outside. No long lifecycle or evidence operation
holds the region. A persistent publication may use only the qualified bounded
owner operation; unknown storage completion closes/fences rather than allowing
another writer to reinterpret a partial result.

Before committing, validate exact aggregate/history identity and revision,
complete primitive/parent/phase membership and revisions (including absence and
links), claim revision/absence, coordination revision and current authority.
Detached reads and evidence are collected outside and revalidated inside.
Stale/conflicting/unknown validation changes no durable fact and issues no
permit; it requires reassessment. It may retain a closed local fence.

The current DP-014 aggregate mutex and DP-015 client/ledger locks are separate;
stable rereads alone do not implement this protocol. Cross-process writer
exclusion requires validated current DP-023 authority and a conforming storage
boundary. No cross-store distributed transaction is assumed: owner publications
remain separate, all valid writers share ordering, and restart admission starts
closed. Physical layout remains undecided.

## 6. Admission fence before claim

Startup admission defaults closed before reading coordination, aggregate or
command truth. For a required recovery assessment under DP-017 section 7
(matching live ownership cannot be proved or an unresolved claim exists), on
first non-clean or unknown observation, enter the participant
and close its sticky local fence before returning that observation or attempting
claim. If a command won the region first, the assessment is stale and must read
the entire set again. A stale claim rejection has zero durable mutation while
the local fence stays closed. Observation does not itself create a claim.

A crash before claim cannot reopen admission because successor startup is
closed. Only fresh authoritative fully clean no-claim verification under joint
ordering permits normal admission without recovery mutation. Missing storage is
not a qualified no-claim read. Existing unresolved claim stays closed until
future DP-023 item 10 release. All state-changing and later-phase paths obey the
fence; the tracked-Start Stop exception does not survive ownership loss.
Authorized same-key observation/terminal replay remains read-only with no permit.
An ordinary observation of a normally managed live Starting/Running Instance is
not this recovery trigger and does not disable the normal tracked-Start Stop path.

## 7. Authentic origin and loss of prior authority

Before any normal command/phase permit issuance or delegation, durable private
coordination provenance binds exact scope, command/phase identity, admission
revision and parent links to the authoritative process execution generation.
DP-014 lifecycle producers also bind operation kind, candidate attempt when
applicable, expected/resulting aggregate revision and related command/phase to
their authentic producer generation. Recovery issuer generation and an opaque
incarnation are persisted before recovery permit issuance.

Provenance is coordination metadata, not another lifecycle/command truth store,
attempt execution binding or containment-ledger extension. It must commit with
the owner publication or as a write-ahead certificate before that publication;
either path requires exact inspection before capability release. An orphan
certificate alone is neither admission nor claim truth. Proven publication
absence at unchanged expected revision permits retry of the same candidate;
exact matching publication correlates provenance with owner facts. Unknown,
newer, foreign or ambiguous state stays closed and grants no capability.
No certificate deletion or unconditional reinterpretation is authorized.

Validated current containment authority plus exact prior-generation supersession
may prove process-wide loss only if every relevant non-terminal producer and
normal/recovery capability origin maps to a proven terminated generation.
This covers CommandOnly and UnboundAttempt without assuming attempt binding.
Missing/legacy provenance is unresolved; it cannot be fabricated from current
generation, client generation, absent maps, PID or time. Relevant current-generation
provenance prevents takeover under this restart-only protocol.

This proof is distinct from exact attempt-bound resource absence and Host shutdown
completion. Producer provenance grants neither execution binding nor Stopped.
Outside-region evidence is an immutable exact supersession observation; inside,
revalidate its provenance binding and the local current-authority identity and
unfenced state before commit and issuance. No evidence query callback runs there.
Authority loss follows DP-023 fatal fencing/termination, never in-place takeover.

## 8. Durable coordination record and storage obligations

The abstract record retains immutable scope and non-reused opaque claim identity,
claim revision, starting aggregate revision and attempt/history correlation,
complete captured non-terminal primitive/parent/phase identities/revisions with
membership/absence proof, creation generation, current issuer generation and
opaque incarnation, observed producer-provenance set, and unresolved/closed
barrier state. Resume preserves original claim identity and observations and
conditionally appends current monotonic observations and a fresh issuer; it never
rewrites original observations as if no reconciliation occurred.

The replaceable persistent boundary must provide complete atomic record CAS,
monotonic revisions, exact detached readback, inspectable committed/absent/unknown
results, process-restart retention and existing-only original storage-authority
validation. Missing, replaced, rolled-back, corrupt or uncertain storage is not
empty and grants no authority. Provisioning must prevent undetectable joint
rollback/clone; candidate storage cannot attest its own expected authority.
Only qualified authoritative absence proves no claim.

DP-014 and DP-015 truth must also survive restart. Persisting only claim while
aggregate/command stores disappear cannot satisfy item 7. Their current Store
and MemoryStorage remain process-local. The containment ledger remains only
generation/supersession facts. No raw permits, Host pointers, payloads, Secrets,
PID or wall-clock takeover metadata belong in coordination records.

## 9. Claim, confirmation and single issuance

Under the joint region, fresh exact non-clean facts plus complete prior-authority
loss proof allow one conditional persistent claim and closed recovery barrier at
one logical publication point. The claim commits its exact captured revisions,
scope, provenance and candidate issuer. Stale/conflicting input has zero durable
mutation. All valid normal writers are excluded by the fence and claim.

Durable commit alone returns no permit. Exact candidate readback must confirm the
whole record. Then, under the same participant with current authority and revisions
revalidated, record one local issuance disposition for exact claim revision and
issuer incarnation. Only the synchronous path owning the candidate receives one
private non-transferable permit; observers never do. Issuance is not reconstructed
from stored identity. Recovery authority cannot invoke Start, Stop, Flow, Load,
Build, Launcher or adopt a Host.

Unknown commit/readback leaves admission closed and issues none. Inspect that
same candidate first; confirmed absence with unchanged expected state permits
only exact-candidate retry. A different tail/claim or unknown authority cannot be
overwritten. If local issuance may have occurred without proven disposition,
never issue a duplicate; same-generation reissue/adoption is unsupported.
Cancellation/loss keeps claim unresolved. A claim committed before subsequent
fatal fencing remains for successor inspection; fenced authority cannot use a
permit or delegate work.

## 10. Conditional resume

A successor first proves prior issuer and all relevant normal-capability origins
terminated, reads the full current aggregate/command/claim/provenance set, and
revalidates exact revisions under joint ordering. It retains claim identity and
original observations, conditionally advances claim revision with fresh issuer
generation/incarnation and current observations, confirms exact readback, then
uses the single-issuance protocol. It never reconstructs the old permit.

Current-generation issuer, incomplete origin set, stale facts, missing original
authority or uncertainty whether a prior issuer/capability can still survive
prevents resume. Proven prior-generation termination resolves historical possible
issuance; current-generation uncertain issuance never permits reissue. Inspect-first convergence
after every unknown cut precedes retry. Later items 8–10 alone perform monotonic
reconciliation and release; claim/resume cannot clear an attempt, terminalize a
command/parent, manufacture an Owner completion basis or reopen a claimed barrier.

## 11. Claim-aware read-only assessment

Extend the private assessment with a detached exact claim/provenance observation
and revision, supplied explicitly by recovery's reader. Stable rereads include
the claim/coordination revision around evidence. Missing/foreign/stale/unknown
reads return Unknown, not absent claim. Assessment exposes no mutation capability.

Clean requires no active attempt, no non-terminal primitive/parent/existing phase,
coherent terminal aggregate/command facts and qualified absence
of an unresolved claim. The same terminal facts with an exact unresolved claim
classify ReleaseOnly, never Clean; this classification alone performs no release
or reopening. Nonterminal classifications retain exact claim observation and
closed admission. Current TASK-075 has no claim reader or ReleaseOnly result;
those additions are planned, not existing capability.

## 12. Risk-oriented proof obligations

| ID | Future executable proof | Source |
| --- | --- | --- |
| P1 | All owner writers share one exact participant; no admission/phase bypass | DP-017 §§8,19 |
| P2 | Command-vs-observation/claim race; stale full-set/absence revisions cause zero durable mutation | DP-014 §16; DP-017 §§8–9 |
| P3 | First non-clean/unknown observation and restart retain closed fence | DP-017 §§7–9,20 |
| P4 | Concurrent claim paths produce one committed claim and at most one live permit | DP-017 §9 |
| P5 | CommandOnly/unbound origins prove old process loss without synthetic binding | DP-017 §§10–11 |
| P6 | Unlocked live callback, client expiry, same-generation issuer or missing provenance prevent takeover | DP-015 permits; DP-017 §9 |
| P7 | Crash at provenance/publication/readback/issuance cuts; exact inspection, no duplicate capability | DP-014 §17; DP-017 §§9,20 |
| P8 | Persistent aggregate, complete commands and claim survive real process restart | DP-017 §§9,17,23 |
| P9 | Missing/replaced/rollback/clone/corrupt authority never becomes empty | DP-023 §8.1; DP-017 §23 |
| P10 | Fresh conditional successor resume preserves original observations; stale/foreign input mutates nothing | DP-017 §§9,16 |
| P11 | Clean vs ReleaseOnly; stable detached claim-aware reads give no permit | DP-017 §§7,12 |
| P12 | Cancellation/unknown result/fatal fence keep admission closed and disable issuance | DP-017 §§20–21; DP-023 §12 |
| P13 | Independent Instances, fixed lock order, evidence/lifecycle outside region, race/stress | DP-017 §19 |
| P14 | No lifecycle/reconciliation/release/Owner-basis promotion or new attempt from claim | DP-017 §§13–17 |

All fourteen rows are planned obligations, not PASS. In-memory simulations can
prove local ordering/issuance mechanics only. They cannot prove P8/P9 durable
conformance, process-origin production coverage or production recovery readiness.

## 13. Prerequisites and next decision

First candidate: separate design-only DP-024 Design Status and conformance
decision against Approved sources and all proof obligations. It is Not Activated;
this Draft's acceptance cannot authorize implementation. That decision must
bound the first isolated mechanics slice and its exact owner seams/proof coverage,
then distinguish it from persistent DP-014/DP-015/recovery storage qualification.
Concrete durable adapter/schema/provisioning remains a separate prerequisite.

Only a later fresh intake with approved mechanics, exact owner integration,
authentic provenance and qualified durable stores can claim item 7 completion.
Items 8–10 reconciliation/release and item 11 reporting/integration remain later.
One ordering/claim behavior is kept intact; evidence scanning, a second adapter,
production wiring and unrelated functionality must be split. Production-line
and file-count Size Guard requires fresh evaluation at implementation intake.

## 14. Decision boundary

Record this proposed private restart-based protocol as Draft / Planned. Existing
Approved statuses and delivered capabilities remain unchanged. No code, durable
store or production recovery has been delivered by TASK-079. Unknown provenance,
storage, authority or commit outcome always leaves admission closed.

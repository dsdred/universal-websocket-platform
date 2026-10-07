# DP-024: Runtime Recovery Claim and Admission Boundary

[Russian version](../../ru/design/DP-024-runtime-recovery-claim-admission-boundary.md)

## 1. Status

- **Design Status:** Approved
- **Implementation Status:** Planned

TASK-079 introduced the proposed private boundary. TASK-080 independent Architect
Confirmation approves the refined design recorded here; Implementation remains
Planned. Approval delivers no code or storage qualification.
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
and generation evidence. This Approved refinement overrides none of those sources.

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

Exact scope is operational/containment domain, Workspace, Configuration and Runtime
Instance. Composition explicitly supplies one identical private per-Instance
ordering participant to DP-014, DP-015 and recovery owners. Different Instances
have independent participants. The participant owns ordering and a local fence;
owner records, validation and conditional publications remain with their owners.
Recovery owns claim/barrier coordination and its non-transferable permit;
composition supplies current containment authority and prior-process-loss proof.
No raw map/lock, generic transaction registry, service locator, arbitrary callback
or ordinary lifecycle capability crosses this boundary.

These are present owners and prospective participation seams, not delivered APIs:

| Owner/source | Required participation |
| --- | --- |
| `runtimeidentity/store.go` | Instance creation/candidate identity allocation where absence/identity is affected; attempt claim/history append, execution binding, Running publication, Stop claim, all authority-specific terminal publishers including recovery |
| `runtimecommandidempotency/store.go` | Primitive inspect/claim, tracked-Start Stop admission, permit consumption/delegation gates, terminal publication |
| `managed_start.go`, `managed_parent.go`, `orchestration_admission.go` | Managed candidate revalidation, tracked parent/preclaimed StopOld atomic admission, binding/rendezvous continuation, Satisfied terminal publication |
| `parent_store.go` | Parent admission, StopOld/StartTarget creation/delegation, phase/parent terminal publication, omission/order decisions |
| `assessment_snapshot.go` and DP-014 detached readers | Complete coherent fact collection; pending/unknown coordination cannot be qualified no-claim/Clean; reader snapshots expose no participant, lock, map or authority |
| `MemoryStorage.nextGeneration` / `NewBoundary` | Client-epoch invalidation, never process-termination proof |
| `runtimerecoveryassessment` | Planned detached claim/provenance reader and ReleaseOnly, no mutation/issuance |
| `runtimecontainmentcomposition` | Explicit same participant supply, current containment authority and immutable supersession observation; no synthetic producer provenance or containment-ledger extension |
| Future recovery owner/store | Claim CAS, exact readback, conditional resume, local single issuance; future release participates but is not activated |

All writers and delegation gates must participate before integrated ordering can
be claimed. Ordinary terminal completion cannot bypass the participant. A live
callback may run outside locks; takeover still requires proven termination of
every authentic relevant process origin.

## 5. Joint ordering protocol

The outer exact-scope serialization token and owner-local locks are different
boundaries. Every seam in section 4 participates; recovery claim/resume and
future release share ordering. Acquisition is outer participant, then the fixed
owner-local lock order; reverse entry is forbidden. DP-014 aggregate locks,
DP-015 active-client-generation/ledger locks and Runtime Owner locks protect
only short internal transitions. Release every owner-local lock before external
persistence/readback, authorization, evidence queries, waits or lifecycle work.

Recovery may retain the outer token during qualified bounded CAS/readback;
this is never permission to hold any owner-local lock across external I/O.
Different Instances remain independent. Unknown completion retains closed
admission and unresolved disposition, even after the physical token is released.
No lifecycle/evidence callback runs under the token.

DP-015 section 13.1 remains authoritative: parent, preclaimed StopOld, sole Stop
exception occupant, rendezvous and private live phase permit form one atomic
internal transition under active-generation/ledger locks. This design neither
splits that transition nor moves ordinary parent admission into recovery storage.
A future persistent DP-015 adapter must separately demonstrate compatibility.
No new ordinary-command publication protocol or adapter is approved here.

Collect detached facts/evidence outside owner-local locks. Under the participant
briefly lock and validate exact aggregate/history, complete command membership,
absence/links/revisions, claim and coordination revisions, authentic provenance,
current unfenced containment authority and local client/admission epoch. Record
only the closed candidate reservation described in section 9, then unlock.
Stale/conflicting validation performs zero durable mutation and issues no permit.

All conforming writers remain excluded by the token while recovery publication
is pending; validated DP-023 authority excludes another authorized process writer.
Captured owner revisions therefore cannot change through a conforming writer
between validation and recovery CAS. A storage adapter lacking this guarantee
is unqualified. Separate current owner mutexes/stable rereads do not implement
this protocol. No cross-store distributed transaction or physical layout is chosen.

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
attempt execution binding or containment-ledger extension. The lock separation
and reservation/CAS/readback discipline of sections 5 and 9 applies; no external
provenance publication or inspection runs under owner-local locks. It must commit with
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

After detached collection and section 5 validation, close/retain the sticky
local admission fence and record a private candidate reservation containing the
exact validated facts. Release every owner-local lock. This local transition is
only fence/reservation: no committed claim, success, reopened admission, usable
permit or delegation. The reservation is not recovery truth.

Retain the outer token and perform qualified recovery-store conditional atomic
publication. Its successful durable CAS is the claim commit/publication point:
one complete record contains immutable claim identity, captured facts/provenance,
issuer generation/incarnation, monotonic revision and unresolved closed barrier.
Exact candidate readback occurs outside owner-local locks. Then briefly reacquire
local locks under the participant and revalidate reservation, epoch, exact owner
revisions, claim revision and current unfenced authority. Record one local issuance
disposition. Release locks and participant before synchronous delivery to the
original candidate-owning path. Observers never receive a permit. No issuance is
reconstructed from stored identity; the private permit is non-transferable and
cannot invoke Start, Stop, Flow, Load, Build, Launcher or adopt a Host.

Every issued and delivered permit is permanently bound to exact claim revision,
issuer incarnation, local client/admission epoch and current containment authority.
After releasing the issuance locks/token, synchronous delivery enters a new short
closed internal acceptance gate atomically ordered with epoch replacement/fatal
invalidation. It performs only bounded local validation/acceptance, no external I/O,
lifecycle or arbitrary callback under owner-local locks. Invalidation winning before
acceptance suppresses successful delivery and leaves issuance exhausted/unresolved.
If acceptance wins first, subsequent invalidation immediately disables the handle;
retaining or returning it is no continuing authority. Delivery linearizes at this
gate, not at physical return. Every later permitted conditional recovery use enters
a fresh exact-epoch/current-unfenced-authority gate under the same participant,
atomically ordered with invalidation. Stale handles reject without a new authorized
mutation or delegation. External persistence remains outside all owner-local locks
using sections 5 and 9 reservation/publication/confirmation. An already in-flight
uncertain publication is inspected as possibly committed, never reported successful
from invalidated authority. No gate retries or recreates an issued permit; epoch
invalidation is not proof that an old process/callback terminated.

Durable commit and live issuance are distinct. DP-017's atomic semantic boundary
is preserved: normal admission/another issuer cannot interleave, reservation
grants no authority, interrupted completion remains unresolved and closed.
Client-generation replacement or fatal fencing invalidates pending reservation
and disables issuance even during persistence. Invalidation never waits for
external storage under owner-local locks or acquires locks in reverse order.
It neither rewrites captured owner revisions nor erases a possibly committed
claim, nor proves prior-process termination. Inspection may finish, but the
invalidated candidate cannot successfully return a permit.

Unknown publication/readback stays closed. Inspect the same candidate first.
Retry requires proven exact absence, unchanged expected durable/owner facts,
valid original storage authority and fresh current authorization; retry uses the
same candidate. Another tail/claim is not overwritten. If issuance may already
have occurred, same-generation reissue/adoption is forbidden. Cancellation/loss
retains unresolved claim; a fenced issuer cannot continue or delegate.

| Cut | Required recovery behavior |
| --- | --- |
| Before reservation/fence | No claim/permit; successor admission starts closed |
| Local fence/reservation, before durable CAS | No committed claim/success; successor inspects qualified storage, reservation never reconstructs a permit |
| Durable publication uncertain | Closed admission; exact-candidate inspection before retry |
| Claim committed, before local confirmation | Authoritative unresolved claim retained; successor proves prior origins terminated before conditional resume |
| Confirmed claim, before issuance | Stored identity does not prove issuance; same-generation uncertainty fails closed |
| Local issuance recorded, before/uncertain delivery | Invalidation before acceptance suppresses success; after acceptance disables handle; no duplicate in that generation/incarnation |
| Permit delivered, cancellation/client-epoch replacement/fatal fence | Claim unresolved; invalidated handle cannot continue/delegate |
| Restart | No old permit revived; fresh conditional resume requires all relevant origins terminated and preserves original observations |

## 10. Conditional resume

A successor first proves prior issuer and all relevant normal-capability origins
terminated, reads the full current aggregate/command/claim/provenance set, and
revalidates exact revisions under joint ordering. It retains claim identity and
original observations, conditionally advances claim revision with fresh issuer
generation/incarnation and current observations, confirms exact readback, then
uses sections 5 and 9 reservation, unlocked CAS/readback and single-issuance
protocol. No owner-local lock crosses external I/O. It never reconstructs the old permit.

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
reads return Unknown, not absent claim. Pending reservations/unknown disposition
never qualify as no-claim or Clean. Assessment exposes no mutation capability.

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
| P7 | Crash at provenance/publication/readback/issuance and issuance→unlock→delivery invalidation cuts; exact inspection, no duplicate capability | DP-014 §17; DP-017 §§9,20 |
| P8 | Persistent aggregate, complete commands and claim survive real process restart | DP-017 §§9,17,23 |
| P9 | Missing/replaced/rollback/clone/corrupt authority never becomes empty | DP-023 §8.1; DP-017 §23 |
| P10 | Fresh conditional successor resume preserves original observations; stale/foreign input mutates nothing | DP-017 §§9,16 |
| P11 | Clean vs ReleaseOnly; stable detached claim-aware reads give no permit | DP-017 §§7,12 |
| P12 | Cancellation/unknown result/fatal fence keep admission closed; epoch replacement revokes issued and delivered permits | DP-017 §§20–21; DP-023 §12 |
| P13 | Independent Instances, fixed lock order, evidence/lifecycle outside region, race/stress | DP-017 §19 |
| P14 | No lifecycle/reconciliation/release/Owner-basis promotion or new attempt from claim | DP-017 §§13–17 |

All fourteen rows are planned obligations, not PASS. In-memory simulations can
prove local ordering/issuance mechanics only. They cannot prove P8/P9 durable
conformance, process-origin production coverage or production recovery readiness.


### First foundation slice coverage

| Proof | Planned first foundation coverage | Remaining prerequisite |
| --- | --- | --- |
| P1 | Shared identity/no bypass among test participants only | Every real writer/delegation gate |
| P2 | Local race/reservation/revision rejection | Real complete owners and durable conditional publication |
| P3 | Local startup/sticky fence | Real restart with qualified owner/claim stores |
| P4 | Reservations grant no authority | Durable one-claim/single live issuance |
| P5 | Reject missing/current synthetic origins | Authentic durable producer provenance |
| P6 | Client expiry is not takeover proof | Complete origins/surviving real callbacks |
| P7 | Local cut/invalidation model only | Real provenance/CAS/readback/issuance crash cuts |
| P8 | Deferred | Persistent aggregate/complete commands/claim restart |
| P9 | Deferred | Existing-only authority, anti-clone/rollback provisioning |
| P10 | No resume authority exposed | Conditional durable successor resume |
| P11 | Pending/unknown never qualified Clean in model | Actual claim-aware assessment/ReleaseOnly |
| P12 | Local cancellation/epoch/fence negative gates | Actual fatal authority/uncertain storage cuts |
| P13 | Independent scopes/fixed order/no owner locks during simulated I/O | Real integration, race/stress/persistent I/O |
| P14 | No lifecycle/reconciliation/release capability | Continued integrated negative proofs |

This table is planned partial model coverage, never PASS or real-owner/durability
qualification. The eight cuts in section 9 require future executable evidence.

## 13. Prerequisites and next decision

The next recommendation is a separate isolated private ordering/fence foundation,
Not Activated. Fresh independent intake must bound scope and Size Guard. It may
contain exact-scope participant identity supplied explicitly, startup-closed/sticky
fence, closed typed reservation/epoch invalidation, fixed lock order and test
owner participants. It validates negative dispositions without claim-store mutation,
provenance publication, live recovery permit, lifecycle delegation or production
composition. No package/API/file-count promise is made before intake.

Test owner ports prove local semantics only, not participation of every real
DP-014/DP-015 writer. Real-owner and generation-transition integration is a separate
bounded slice. Persistent aggregate/complete command/recovery adapter and original
storage-authority/provisioning qualification remain distinct prerequisites; DB,
schema, adapter and provisioner choices are not made here.
Only later fresh intake with integrated owners, authentic durable provenance and
qualified stores can claim item 7 completion. Items 8–10 reconciliation/release
and item 11 reporting/integration remain later. No next task is started here.

## 14. Decision boundary

TASK-080 independent Architect explicitly approves this refined private
restart-based design: Approved / Planned. DP-015 internal atomic transitions and
external-I/O prohibition remain unchanged. No code, durable store, recovery permit
or production recovery has been delivered by TASK-079/TASK-080. All proof rows
remain planned. Unknown provenance, storage, authority or publication outcome
leaves admission closed; design Approval is no implementation activation.

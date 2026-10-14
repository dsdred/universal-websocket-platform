# DP-024: Runtime Recovery Claim and Admission Boundary

[Russian version](../../ru/design/DP-024-runtime-recovery-claim-admission-boundary.md)

## 1. Status

- **Design Status:** Approved
- **Implementation Status:** Partial (isolated local test-owner foundation, closed DP-014/DP-015 participation and client replacement)

TASK-079 introduced the private boundary. TASK-080 independent Architect
Confirmation approves the refined design recorded here. TASK-081 implements only
`internal/runtimerecoverycoordination`: exact-scope explicit participant identity,
closed aggregate/command test-owner models, detached complete membership/revisions,
startup-closed/sticky fence, local reservation and client/fatal epoch invalidation.
Outer serialization and short state/aggregate/command locks remain separate;
reservation sessions retain only the outer token during simulated external I/O.
One local revalidation disposition and idempotent disposal create no claim or permit.

TASK-084 adds the minimal dependency leaf `internal/runtimecoordination` and
complete closed participation of the real DP-014 owner: construction before
allocation, every direct/narrow writer, ordered detached readers, coherent complete
history and allocation/absence stamps. All coordinated writes reject; missing
commands, current authority and qualified no-claim cannot open admission. Legacy
constructor regressions retain successful isolated publication behavior.

TASK-085 adds complete closed real DP-015 participation, detached command
membership/link/presence stamps and storage-wide client replacement. The explicitly
supplied inert `AuthorityFence` and storage-owned `ClientEpoch` provide only local
ordering/invalidation; actual containment authority and fatal coupling are absent.
Authorized replay/read remains available, every coordinated mutation rejects,
and successful fixtures do not prove coordinated startup.

Implementation Status remains Partial for TASK-081 models and TASK-084/TASK-085
closed local slices. Each task's gates/closure resolve only from its newest valid
matching Recovery Evidence Envelope. Integrated P1–P14 remain Planned and item7
unproven. Authentic producer provenance, persistent stores, durable claim/readback/
resume and live recovery permits remain absent. DP-017 stays Approved / Planned
overall with only TASK-075 isolated read-only assessment. No durable recovery
barrier or production recovery exists.

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

This is the complete participation inventory. TASK-084/TASK-085 deliver closed
DP-014/DP-015 participation and client replacement; recovery and actual containment
seams remain planned:

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

| Proof | TASK-081 isolated local model coverage | Remaining prerequisite |
| --- | --- | --- |
| P1 | Explicit same participant and exact owner identity among test models | Every real writer/delegation gate |
| P2 | Detached complete membership/revisions and stale/foreign rejection | Real complete owners and durable conditional publication |
| P3 | Startup closed; fresh clean-only opening and sticky unknown/recovery fence | Real restart with qualified owner/claim stores |
| P4 | Reservation/revalidation is only local data/disposition, no authority | Durable one-claim/single live issuance |
| P5 | Deferred; no origin input or loss proof | Authentic durable producer provenance |
| P6 | Client epoch invalidation supplies no takeover authority | Complete origins/surviving real callbacks |
| P7 | Pending local candidate invalidation during simulated external I/O | Real provenance/CAS/readback/issuance crash cuts |
| P8 | Deferred | Persistent aggregate/complete commands/claim restart |
| P9 | Deferred | Existing-only authority, anti-clone/rollback provisioning |
| P10 | No resume operation or authority | Conditional durable successor resume |
| P11 | Unknown/incomplete/claim-present model facts cannot open admission | Actual claim-aware assessment/ReleaseOnly |
| P12 | Local epoch/fatal invalidation closes admission and rejects old candidate | Actual fatal authority/uncertain storage and delivered-permit revocation |
| P13 | Independent Instances, fixed owner lock order, owner locks released during simulated I/O | Real integration, race/stress/persistent I/O |
| P14 | No lifecycle/reconciliation/release capability | Continued integrated negative proofs |

This is partial local model coverage, never full P1–P14 PASS, real-owner participation
or durability qualification. The eight cuts in section 9 still require integrated
executable evidence. Reservations and successful local revalidation grant no authority.

## 13. Prerequisites and next decision

TASK-081 supplies only isolated closed test-owner models. TASK-082 completes
real-owner/generation readiness: full integration is SPLIT REQUIRED / NOT READY
AS ONE BOUNDED TASK. TASK-083 records the concrete private realization in section 15.
TASK-084 implements the minimal leaf and complete closed DP-014 local owner
coverage; TASK-085 adds complete closed DP-015 participation, membership stamps
and storage-wide client replacement described in section 15. The next recommendation
is fresh bounded actual containment-source fatal coupling and explicit composition
supply readiness, Not Activated, with fresh Size Guard and independent review.
Missing current authority/qualified no-claim keeps coordinated writes closed;
inert root construction is not qualification of containment authority.

Durable provenance, persistent aggregate/complete command/recovery storage and
original storage-authority/provisioning qualification remain separate prerequisites;
DB, schema, adapter and provisioner are not selected. Only later fresh intake with
integrated owners, authentic durable origins and qualified stores can claim item7.
Items 8–10 reconciliation/release and item11 reporting/integration remain later.
No implementation or successor is activated by this design task.

## 14. Decision boundary

TASK-080 independent Architect explicitly approves this refined private
restart-based design: Approved; Implementation remains Partial for TASK-081 local models and closed DP-014/DP-015 participation plus client replacement in TASK-084/TASK-085. DP-015 internal atomic transitions and
external-I/O prohibition remain unchanged. No code, durable store, recovery permit
or production recovery has been delivered by TASK-079/TASK-080. All integrated proof rows remain planned. Unknown provenance, storage, authority or publication outcome
leaves admission closed; design Approval is no implementation activation.

## 15. Concrete local owner participation plan

TASK-083 records this private realization of the existing Approved semantics.
TASK-084 implements the minimal typed leaf, closed real DP-014 participation
and coherent detached aggregate observations. TASK-085 adds complete closed real
DP-015 participation, membership observations and local client replacement.
Actual fatal-source coupling, reservation-session and integrated recovery protocol
remain planned. Design Status stays Approved; Implementation Status stays Partial.
These slices grant no claim, permit, takeover authority or production startup
qualification.

### 15.1 Typed binding and explicit coverage

The dependency leaf `internal/runtimecoordination` imports no aggregate, command,
containment, lifecycle or recovery owner. TASK-084 implements immutable validated
`Scope`, copy-safe `Participant`, non-rebinding `AggregateAttachment` and short
`ObservationRegion`; its transition region always rejects. TASK-085 implements
copy-safe `CommandAttachment`, `ClientEpoch`, inert `AuthorityFence` and bounded
registration/replacement regions. Reservation-session handles remain planned.
Composition supplies the identical participant explicitly. Attachments bind exact
owner identity and scope once; duplicate, foreign and rebinding attempts reject.
Complete-owner markers belong to conforming owner construction, never caller claims.
No raw locks, generic callbacks, service discovery or ordinary lifecycle capability
cross the seam. The leaf latch does not query or replace containment authority.

Implemented `runtimeidentity.NewCoordinatedStore(scope, participant)` binds one
Instance before allocation: historical `AllocateCandidateIdentity(id)` cannot
convey domain/Workspace/Configuration. Every direct method, including narrow
terminal publishers, rejects foreign identity and enters its own conforming region.
Historical `NewStore`, `NewMemoryStorage` and uncoordinated composition stay
supported but do not count toward DP-024 conformance. TASK-085
`NewCoordinatedMemoryStorage(authority)` uses typed `AttachInstance(scope, participant)`
before capturing ledger absence. Staged registration commits only after short owner
catalog checks; rejected unattached calls disclose no absence and do not poison attachment. Attachment is independent of aggregate or ledger existence.

Complete DP-014 coverage includes allocation/create, attempt claim/history append,
execution binding, Running, Stop and all authority-specific terminal publishers,
aggregate/history readers and exact-attempt capture/revalidation. TASK-085 complete closed DP-015
coverage includes primitive inspect/claim and tracked Stop; managed admission,
binding and live gates; parent/phase admission, creation, delegation, terminal and
expiry; tracked parent/preclaimed StopOld/slot/rendezvous/private permit atomic
admission and consumption; replay-first orchestration; pending Stop and managed
rendezvous transitions; complete assessment snapshots and client replacement. Actual
containment source fatal paths must participate in the later separate slice. A wrapper cannot prove direct-call coverage.

### 15.2 Complete detached observations and revisions

Detached stamps contain exact scope, opaque attachment identity, aggregate presence/
allocation revision, aggregate/history revision, command-ledger presence, command
membership revision, client epoch and participant epoch. Capture and revalidation
change no domain fact and return no handle or authority. Allocation/create advance
presence revision; successful aggregate publications advance their existing
aggregate revision. Ledger creation, primitive/parent/phase membership/publication,
live-permit consumption/expiry, Stop-slot and rendezvous changes advance a separate
coordination membership revision. Existing domain record revisions keep their semantics.

Capture all terminal and non-terminal members and links; an empty set includes its
explicit absence stamp. Missing, stale, foreign or partial coverage cannot open
admission. Create-after-absence must invalidate the stamp. Counters never wrap:
exhaustion closes admission and rejects future authorized mutation.

TASK-084 `CaptureCoordinationSnapshot` and `RevalidateCoordinationSnapshot`
implement only the DP-014 part: exact scope, an isolated opaque identity with no
owner/lock links, presence revision, allocation/aggregate presence, aggregate view
and every terminal/nonterminal history member. Capture verifies complete index,
immutable binding, active/terminal and desired/actual coherence; nil, incomplete
or contradictory owner facts reject without a partial result. Coherent unbound
early facts and retained StopFailure remain valid local observations; the separate
exact execution reader still rejects an unbound selected attempt. No command,
participant/client epoch or qualified no-claim stamp is fabricated. Presence and
domain counters reject exhaustion before mutation. Create-after-absence and
successful-publication observations use explicitly labelled package-local fixtures;
they do not prove successful coordinated mutation.

TASK-085 command capture/revalidation copies every terminal/nonterminal member,
live/slot/rendezvous link and explicit ledger absence under one ordered region.
Opaque snapshot identity has no owner/lock links; mutable permit/channel/bridge
references are removed. Exact scope, intent/authorization, derived phase order and
retained resolved/blocked diagnostic bridges are checked without treating retained
tokens as live authority. Private value/identity accounting advances membership
only for actual changes; reads do not fence or change facts. New coordination
counters reject exhaustion before mutation; existing domain revision semantics
remain unchanged, with covered pre-mutation exhaustion guards. This is no global
legacy counter or authority guarantee. Successful changes use labelled local
fixtures or separate legacy regressions, not coordinated mutation evidence.

### 15.3 Fixed short-region order

The complete planned protocol acquires `Instance outer -> AuthorityFence read gate ->
ClientEpoch read gate -> participant state -> owner-local locks`. Owner catalog
lookup/registration precedes aggregate locking. DP-015 retains its existing
`clientMu.RLock -> ledger.mu` atomic transition, including parent, preclaimed
StopOld, sole occupant, rendezvous and private live phase permit.

TASK-085 transparently upgrades previously attached aggregate observations to
the root order when command registration commits. A rootless short observation
retains participant state until Close so registration cannot overlap its owner
read; it is not an external-I/O reservation. Rooted command regions retain the
fixed root/state gates before `clientMu.RLock` and ledger locks. Public wrappers
delegate to private short-region helpers, releasing each region before nested
entry or callbacks/waits. Missing current authority/qualified no-claim keeps
transitions closed. Shared root/attachment/region copies retain one disposition;
zero/nil Close is safe. No raw mutex or arbitrary callback is exposed.

Aggregate and command facts may be copied sequentially while the same outer token
excludes writers; raw owner locks never cross packages. Release every authority,
client, participant-state and owner gate before authorization, evidence/storage/
provider calls, lifecycle work or waits. A reservation may retain only outer during
qualified external work. Direct owner calls and narrow publishers acquire their
own conforming region; composition wrapping alone is insufficient. Reverse entry
from an owner/client lock to the outer token is forbidden.

### 15.4 Storage-wide client replacement

`ClientEpoch` is a separate bounded local gate; it does not replace DP-015
`clientMu`. Replacement takes its exclusive gate, closes/invalidates every typed
attached participant under each participant state lock, then publishes the new
generation under `clientMu.Lock`. It takes no outer or owner ledger/aggregate lock
and waits for no external I/O. Another Instance's retained outer/I/O cannot block
this bounded fan-out.

Registration orders `AuthorityFence.R -> ClientEpoch.W -> participant.state`,
with short owner registration afterward. It cannot escape concurrent replacement
or inherit prior opening. Overflow permanently fences the client root and attached
participants and prevents visibility of a new generation. Fan-out waits only for
bounded local gates; it is not process-loss proof or permission to recover.

TASK-085 implements this local registration/replacement order, copy-safe disposal
and permanent exhaustion fencing before new client generation visibility. Its
inert root does not prove prior-process loss, current authority or fatal coupling.
Fixture-only outer retention demonstrates that fan-out takes no outer token;
short observation regions are released before external work.

### 15.5 Fatal coupling at every actual source

The containment keeper owns an explicitly associated `AuthorityFence`. Every
source fatal-publication path couples existing `keeper.fenced` publication and
complete attached-participant invalidation under the authority root exclusive
gate before returning the fatal result. Registration and internal acceptance use
the same root read gate: fatal-first suppresses acceptance; acceptance-first is
immediately invalidated afterward. Terminal completion, expiry and later phases
obey the resulting fence; authorized replay stays read-only.

External `IsAuthoritative`, provisioning validation and `ReadGeneration` storage
work remain outside every short gate. Roots expose only bounded local invalidation,
no evidence callback or termination claim. The leaf neither replaces containment
truth nor changes DP-023 fatal shutdown obligations.

### 15.6 Smallest independent sequence and executable obligations

TASK-084 implements the first slice: leaf binding/local observation regions and
complete closed DP-014 direct-writer, narrow-publisher, reader and absence coverage.
Its 13 proof tests and two regression tests cover copy/zero handles, attachment
before allocation/absence, direct/narrow rejection, detached/stale/foreign/partial
facts, counter exhaustion, independent Instances and released owner locks.
Success fixtures are package-local; historical constructor regressions retain
existing successful publication behavior. Focused stress, full regression and vet
pass; race verification is unavailable in the current CGO-disabled/no-C-compiler
environment. This evidence proves only the local slice, not integrated P1–P14.

TASK-085 implements the next closed slice with 28 added tests: 25 proof and
three regression tests. They cover complete direct/callback rejection, authorized
primitive/managed/replay-first terminal replay with zero work callbacks, detached
membership/absence invalidation, exact contradictory-link rejection, retained
diagnostic bridges, copy/zero/abort handles, attachment/replacement races and
overflow without new generation visibility. Successful mutation fixtures remain
package-local; actual legacy Stop-first/atomicity/counter regressions stay distinct.
Full supported regression/vet and serial focused stress pass. Race verification
is unavailable with CGO disabled; an earlier concurrent legacy bounded-yield
failure followed by serial passes is retained as a limitation, not diagnosed fact.
These proofs establish closed local mechanics only, not integrated P1–P14/item7.

Next candidate: fresh bounded actual containment-source fatal coupling and explicit
composition supply readiness, Not Activated. Each successor needs fresh readiness,
Size Guard and independent review; no successor is activated here.

Missing command attachment, current authority or qualified no-claim verification
keeps every coordinated write closed. No production test-opening API exists.
Synthetic opening in existing closed model tests remains model-only; successful
real-owner startup needs later complete qualification.

Remaining integrated executable proofs must cover every command/live/
rendezvous change invalidating stamps; preserved tracked-parent atomicity; two
Instances and racing attachment during replacement; overflow without new-generation
visibility; fatal before/after acceptance; closed-fence authorized replay with
callback count zero; and external work outside all short gates. These prove only
local participation mechanics, not integrated P1–P14 or item7 completion.

Durable producer provenance, persistent aggregate/complete commands/claim,
independently trusted original storage authority, anti-clone/rollback qualification,
claim-aware assessment, conditional claim/resume and live recovery issuance remain
separate full item7 prerequisites. Reconciliation, release, reporting, production
wiring and activation remain out of this local sequence.

## 16. Source-bound fatal coupling and closed composition readiness

TASK-086 records a planned private realization of sections 4–9 and 15.5. It
supersedes the historical TASK-085 next-readiness recommendation in sections
13/15 without changing those sections, Design Status or Implementation Status.
This design supplies no runtime capability, current-authority qualification,
qualified no-claim, recovery permit or item7 proof.

### 16.1 Actual source inventory and gap

At the TASK-086 baseline, seven production fatal publications exist:

| Actual source | Fatal condition |
| --- | --- |
| `containment.go`, `IsAuthoritative` | Invalid held capability, nil DB or failed provisioning revalidation |
| `containment.go`, `AcquireProcessContainment`, first acquired guard | Acquired capability is invalid |
| `containment.go`, `AcquireProcessContainment`, existing-store guard | Existing provisioned store open fails or held capability is invalid |
| `containment.go`, `AcquireProcessContainment`, successor guard | Generation creation, successor bootstrap, post-append provisioning revalidation or held-capability validation fails |
| `evidence.go`, `ReadGeneration` | Ledger tail contradicts the current generation |
| `evidence.go`, `validateForEvidence` | Held capability is invalid or DB is unavailable |
| `evidence.go`, `fenceEvidenceError` | Provisioning/ledger/storage inspection fails |

Direct test stores are fixtures, not source coverage. The current authority root
contains only an RWMutex; the client owns its participant registry. Participant
roots appear only after command registration commits, leaving aggregate-only
participants rootless. Existing `runtimecontainmentcomposition.New` accepts
prebuilt dependencies and supplies no shared participant or coordinated stores.
Existing authority/evidence checks do not implement source-local invalidation.

### 16.2 Source association before owner attachment

The keeper privately creates and retains one immutable domain-bound authority
root before its first acquired-capability fatal branch. It accepts no caller
inert root as source-associated. Copies retain identity; source/domain rebinding
is forbidden. Private composition obtains that exact keeper-associated root
through the exact active authority and existing external validation. The root
itself never attests containment authority.

Additive source-bound participant construction registers the exact-domain
participant under `Authority.R -> participant.state` before any owner attachment
or absence capture. The authority registry includes every participant, including
aggregate-only and failed partial construction, and every associated ClientEpoch.
No detach/rebind or removal-as-absence is permitted. Root/client construction,
participant registration and staged command registration reject sticky fatal
under the same authority read gate. Aborting command registration adds no command
fact and cannot remove an already source-registered participant. Historical
rootless/inert construction remains supported but unqualified.

### 16.3 Typed source-local publication and complete invalidation

One private keeper helper covers every actual fatal branch. It acquires an opaque
copy-safe fatal-publication region retaining `Authority.W`, publishes existing
`keeper.fenced=true`, invokes the bounded typed `Invalidate()` operation and only
then closes the region and returns. No raw lock or arbitrary callback is exposed.
A handle grants no source association, evidence, opening, generation, release or
recovery authority. Nil/zero handles are safe; copies share one idempotent
invalidation/close disposition.

Every source-associated fast return that observes `keeper.fenced=true` joins
completed publication through a bounded typed same-root completion read before
returning. The read acquires Authority.R outside owner/outer gates and after
external inspection, so in-flight Authority.W must finish flag publication,
complete invalidation and release first. Repeated fatal observations are safe;
the atomic flag alone does not prove completed coupling. This includes an
IsAuthoritative final false result caused by concurrent fencing. Nil/no-source
caller errors fabricate no root or authority.

Fan-out follows `Authority.W -> ClientEpoch.W` (one client at a time) then
`participant.state`, and also covers separately registered partial participants.
It marks root/clients/participants permanently terminal, with MaxUint64 epochs.
Deduplication or idempotent terminal assignments prevent repeated registry
associations from skipping participants or wrapping counters. It takes no
Instance outer, aggregate/ledger/clientMu lock and invokes no callback,
storage/evidence/lifecycle work or external wait. It waits only for bounded local
gates. Existing replacement publication and fatal fan-out serialize through each
client gate; replacement cannot clear terminal state or resurrect a participant.

Authority write orders against registration/observation/internal-acceptance read
gates: fatal-first suppresses acceptance; acceptance-first is invalidated before
fatal returns. Actual live recovery issuance is absent; region ordering or labelled
fixtures cannot prove delivered-permit revocation. Authorized replay remains
read-only with zero work callbacks and all coordinated writes remain closed.

`IsAuthoritative`, provisioning validation and `ReadGeneration` fault detection
remain outside every short gate. Source results/reasons, caller-error nonfencing,
fatal-over-cancellation precedence, retained capability, no reacquisition and
DP-023 process-termination obligations remain unchanged. Local invalidation is
never process-loss evidence or termination proof.

### 16.4 Explicit opaque closed assembly

An additive private assembly obtains the exact keeper-associated root outside
short gates and checks canonical operational/containment domain and exact target
scope. It registers one source-bound participant before creating its coordinated
DP-014 Store and coordinated DP-015 MemoryStorage/attachment, supplying the same
participant and source root to both owners. Only a complete closed assembly is
returned. Partial owners/absence are not exposed; no participant/root, caller
completeness assertion, storage adoption or unqualified owner injection crosses
its surface.

This separate opaque local assembly owns its coordinated dependencies and exposes
only detached read observations. It adds no activation methods or Flow/Owner/
private-invoker adoption. Legacy activation/evidence `Composition` and `New` remain
supported but unqualified. No ordinary runtime owner creation, successful startup,
production wiring or activation is delivered by this assembly.

### 16.5 Next implementation admission and proof obligations

Assessment is Ready. After independent acceptance of this plan, the smallest
implementation candidate is Ready to fresh intake / Not Activated: source-bound
leaf registry/fatal handle, complete actual containment fatal coverage and explicit
closed composition supply. Three existing packages, zero new packages, estimated
300–500 production lines and 13–17 total paths including documentation require
fresh Size Guard; more than 500 production lines or 15 paths triggers reassessment
before further work. Source publication cannot be separated from complete source
association/fan-out/supply without admitting a bypass.

| Future obligation | Required executable proof and limitation |
| --- | --- |
| All actual sources | Static inventory plus fault injection at all seven sites, including bootstrap before active authority exposure |
| Registration ordering | Fatal-first rejection, registration-first invalidation, aggregate-only and aborted command attachment coverage |
| Completed fatal return | Pause after keeper flag publication and before fan-out: repeated-fatal/fast-return readers wait for complete invalidation before false/Unknown return |
| Complete fan-out | Two clients/two Instances, partial participants, exact domain/root mismatch, copy/zero handles |
| Sticky terminal state | Fatal versus replacement cannot resurrect state or wrap epochs; exhaustion exposes no new client generation |
| Bounded gates | Retained outer does not block fatal; no fan-out owner locks; external authority/evidence work outside short gates |
| Explicit supply | Identical source/participant/root for both real owners, foreign scope rejection, no partial exposure or absence leak |
| Closed compatibility | Direct writers reject, authorized replay has zero callbacks; separate legacy constructor regressions |

These are planned proofs, not PASS. Integrated P1–P14 remain Planned and item7
unproven; full integration is Split Required / Not Ready. Authentic durable
producer provenance, persistent aggregate/complete commands/claim, independently
trusted original storage authority, anti-clone/rollback qualification, qualified
current authority/no-claim, claim-aware assessment, conditional claim/readback/
resume and live issuance remain separate prerequisites. Reservation, reconciliation,
release, reporting, production wiring and activation remain later work. No successor
is activated by this design or its acceptance.

## 17. TASK-087 isolated source-bound implementation

TASK-087 adds the local implementation of §16. This appendix temporally supersedes
historical assertions in §§1–16 that actual source fatal coupling and explicit
closed supply are absent. Their original bytes and Design/Implementation Status
remain unchanged: DP-024 stays Approved / Partial. The approved plan is unchanged;
this appendix reports implemented local mechanics, not integrated recovery readiness.

### 17.1 Source association and completed fatal publication

The containment keeper creates and privately retains one immutable domain-bound
`NewSourceAuthorityFence` before its first acquired-capability fatal branch.
`CoordinationRoot` returns only that keeper's same-domain root after external
validation; the root itself grants no current-authority proof. `NewSourceParticipant`
registers before owner attachment/absence, with permanent partial/aborted membership.
The authority registry also retains every associated client. Registration rejects
foreign source/domain and sticky fatal; legacy inert/rootless construction stays
supported and unqualified.

`FatalPublication` provides bounded `BeginFatalPublication`, `Invalidate` and
`Close`, with copy-safe shared disposition and safe zero/nil handling. The keeper's
single helper publishes its existing atomic flag while holding Authority.W,
invalidates clients/participants and completes publication before fatal return.
All seven actual containment publications use it, including bootstrap before
ActiveAuthority exposure. Flag-observing fast returns join completed same-root
publication through `JoinFatalPublication`, including the final false
`IsAuthoritative` path. The atomic flag alone is not completed invalidation.

Fan-out uses Authority.W, then one ClientEpoch.W at a time, then participant state;
it also covers aggregate-only and partially constructed participants. Terminal
state is sticky and epochs become MaxUint64. Replacement cannot resurrect it.
Fan-out takes no Instance outer, aggregate/ledger/clientMu lock and performs no
callback, storage/evidence/lifecycle work or external wait. External source fault
detection remains outside short gates. Existing reasons, caller-error nonfencing,
fatal-over-cancellation, retained capability/no reacquisition and DP-023 process
termination obligations remain unchanged. Local invalidation is not process-loss
or termination evidence, and no delivered recovery permit exists to revoke.

### 17.2 Explicit closed supply and remaining boundary

Separate `NewClosedAssembly` validates canonical operational/containment domain
and exact target, obtains the keeper root outside short gates and supplies one
source-registered participant to its own coordinated DP-014 Store and coordinated
DP-015 MemoryStorage/attachment. Only a complete opaque `ClosedAssembly` returns;
failed partial registrations remain source-retained. Its only read operation is
`Capture(scope)`, returning a `ClosedObservation` with detached `Identities()` and
`Commands()` snapshots. Capture reads the owners sequentially; it is not an atomic
joint snapshot, current-authority qualification or qualified no-claim.

Owners, root/participant, boundary, adoption, activation and recovery capabilities
are not exposed. All coordinated writers remain closed; authorized replay remains
read-only with zero work callbacks. Existing owner implementations and legacy
activation/evidence `Composition.New` remain supported and unqualified.

Proofs and actual check outcomes are recorded in TASK-087's matching Recovery
Evidence Envelope; this appendix does not independently assert PASS or Acceptance.
Source-level tests distinguish actual fatal paths from local ordering fixtures;
ordering fixtures cannot prove live-permit revocation. Integrated P1–P14 remain
Planned, item7 unproven and full integration Split Required / Not Ready. Qualified
current authority/no-claim, authentic durable provenance/stores, original storage
qualification, claim/readback/resume/issuance, reservation, reconciliation/release,
reporting and production wiring/activation remain separate work. Next recommendation
is bounded readiness/decomposition of those prerequisites, Not Activated, without
a readiness verdict or automatic successor activation.

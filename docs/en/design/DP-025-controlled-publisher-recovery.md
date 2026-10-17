# DP-025: Controlled Publisher Recovery

[Russian version](../../ru/design/DP-025-controlled-publisher-recovery.md)

## 1. Status and authority

- **Design Status:** Draft
- **Implementation Status:** Planned (not implemented)
- **Task:** [TASK-097](../../tasks/TASK-097-CONTROLLED-PUBLISHER-RECOVERY-DESIGN.md)

This proposal studies a new project-owner requirement, not a confirmed governance gap.
Architect Confirmation authorizes this bounded comparison; it does not approve an
operational recovery contract. Current PROCESS-001/002 and Publisher rules remain effective.
No provider capability, ownership transition, permission carry-over or migration is activated.
Words such as requires below specify candidate proof obligations, not current permissions.

## 2. Problem, boundaries and decision

Study unfinished publication after irrecoverable loss of its original owning execution run.
Separate authorization, mutation ownership, recovery claim, operation outcomes and terminal state.
An unavailable run is not proof that its subprocesses, credentials or accepted requests are inert.
Safety means no unauthorized/concurrent mutation, no duplicated proven effects and no fabricated provenance.
Liveness means eventual remaining-phase completion under explicitly stated, proven assumptions.
Permanent uncertainty may legitimately prevent liveness.

Compare A Direct Publisher with proven fencing; B Fenced Mutation Gateway; C current fail-closed STOP.
Recommendation: retain C as the only active policy; research B conditionally; admit A only with complete proof.
This is a recommendation for separate design decisions, not approval of B or rejection of current contracts.
No implementation, provider setup, live fixture exercise or normative amendment belongs to this task.

## 3. Existing semantics and evidence ledger

Repository contracts outrank this Draft. Historical evidence remains bounded to its observed cut.
Native API and provider backing are views of one store, not two independent ownership authorities.
Reading a transcript does not restore an owning execution context.

| Evidence | Source and exact scope | Consequence / limit |
| --- | --- | --- |
| E01 | [PROCESS-001](../../engineering/PROCESS-001-AI-DEVELOPMENT-WORKFLOW.md), owning-context tuple and per-call linkage, lines 1208–1223 | Same SID, conversation or role alone does not prove continuity |
| E02 | PROCESS-001, Trusted-Context Release Handoff, line 1484 onward | Only proven source issues Release; exact user Route and durable destination Accept follow |
| E03 | PROCESS-001, canonical store/source emission, lines 1606–1670 | Reader receipt cannot fabricate source event; inaccessible provenance stays Unknown/STOP |
| E04 | PROCESS-001, terminal semantics, lines 1714–1736 | P10 requires immediately preceding Active/Owned(exact actor); protocol is procedural, not a machine lock |
| E05 | [Publisher](../../engineering/agents/publisher.md), Initial Trusted Route and Trusted-Context Ownership Handoff | Dispatch establishes initial owner only after independent persisted readback; handoff does not create a new publication permission |
| E06 | [PROCESS-002](../../engineering/PROCESS-002-DOCUMENTATION-SYNCHRONIZATION.md), Publication-State Synchronization, line 154 onward | Project navigation stores stable facts, not live ownership or temporary checkpoints |
| E07 | [TASK-096](https://github.com/dsdred/universal-websocket-platform/blob/82cf49627e54881599bb60a55f28c8e9c6b0c4db/docs/tasks/TASK-096-FAILED-OWNER-RECOVERY-ASSESSMENT.md), F096-01–F096-12 and decision matrix | Existing paths are conditional; missing evidence is not automatically a governance gap |
| E08 | TASK-096 F096-05/F096-06 | TASK-095 PR #99 merged at c8047fd08caa79480c2bc73b5c382f4aa5079db7; P0–P4 proven; specific P5 invocation did not execute |
| E09 | Historical TASK-096 native InitialDispatch msg_00e5a0948e1deb9d016ad39287fa6c81919a9fae6d99885cb7 | Owning run 01a14a6e-5244-7431-9057-d6acc18bef5f; historical initial binding, not present continuity |
| E10 | Historical TASK-096 BlockerReport msg_00e5a0948e1deb9d016ad3936236688191a3d7a062b10016e4 | P0–P3 completed; that P4 merge command did not execute; no claim about all future effects |

E09/E10 are historical qualified-native locators supplied to this design, not freshly probed remote facts.
TASK-096 target is 82cf49627e54881599bb60a55f28c8e9c6b0c4db; accepted manifest
1e065935ded73cd69c89b7f60a8359488509697d; qualified base is the E08 merge OID.
PR #100 was OPEN at its last confirmed observation; current PR/checks outcome is Unknown here.
TASK-095 remaining P5–P10 and TASK-096 remaining P4–P10 are offline fixtures only.
Their historical permissions are neither reused nor transferred by this proposal.

## 4. Provider capability matrix

No authoritative provider contract qualified in this repository establishes the following guarantees.
Each Unknown is a required research result, not an assertion that the feature does not exist.
Official documented semantics, supported transport/version and observable failure behavior are prerequisites.
A test can corroborate a documented guarantee; a successful sample does not establish universal atomicity.
Runtime DP-023/024 local containment/claim models are not Publisher-provider capability evidence.

| Capability | A evidence required | B evidence required | Present qualification |
| --- | --- | --- | --- |
| Conditional claim / CAS | Exact predecessor/version, durable success/conflict semantics | Same, for gateway authority and succession ledger | Unknown |
| Old-executor fencing | Every local/remote/subprocess/legacy/queued mutation path rejects old epoch | Exclusive mediated credentials and enforced single active gateway across replicas/crashes | Unknown |
| Durable ordering | Claims, revokes, admissions and calls share a proven order | Durable sequencer/journal with recovery and no stale leader writes | Unknown |
| In-flight outcome | Request identities and authoritative completion/absence evidence | Same; journal alone cannot cancel an external accepted request | Unknown |
| Lost response reconciliation | Readback identifies exact write/request without blind repetition | Intent/result linkage survives response loss and process death | Unknown |
| Git/GitHub conditional effects | Documented exact-ref/head predicates and rejection semantics per operation | Gateway validates predicates and records outcomes; external atomicity still unproven | Unknown |
| Credential mediation | All cached/shared credentials and bypass paths closed | No direct caller/legacy credentials outside enforcement | Unknown |
| Clock/lease behavior | Expiry plus enforced revocation, not elapsed time alone | Monotonic epochs and enforced fencing despite pauses/partitions | Unknown |

Unique ledger claim, machine fencing and atomic external mutation are three different properties.
CAS proves at most one winning conditional ledger update in its stated serialization domain.
It does not stop an old process using retained Git/GitHub authority.
A token check followed by an external request leaves a check-then-act race.
No cross-system transaction, exactly-once GitHub execution or automatic cancellation is assumed.

## 5. A/B/C comparison

| Boundary | A — Direct Publisher | B — Fenced Mutation Gateway | C — current STOP |
| --- | --- | --- | --- |
| Provider assumptions | Complete enforcement on every direct mutation path | Fully mediated gateway plus proven durable sequencing/fencing | Existing qualified native store and current contracts |
| Authorization | Explicit recovery scope and named successor; carry-over policy unresolved | Same; gateway must order authorization/revoke with admission and execution | No new succession permission |
| Admission | Proven fence, reconciliation and conditional successor claim | Gateway-enforced fence, reconciled journal and conditional successor admission | Proven continuity or existing source handoff only |
| Ownership point | Durable conditional admission, independently read back | Same point in enforced gateway authority domain | Existing InitialDispatch / Accept rules |
| Old executor | Reject stale authority at every effect boundary | Remove direct capabilities; fence old gateway replicas and queued intents | No replacement owner inferred |
| In-flight Git/GitHub | Inventory/drain/reconcile each old request | Durable intent/result journal plus external reconciliation | Unknown effects remain STOP |
| Crash consistency | Fence survives restart; ledger/remote split explicitly recovered | Persist intent before execution; uncertain calls block conflicting work | No new writes |
| Competing successors | Exact predecessor CAS and enforced epoch, no ABA | Serialized conditional admission and fenced leadership | Neither obtains new ownership |
| Lost CAS/API response | Read exact winning claim/result before retry | Reconcile exact intent/request and current authority before retry | Inspect only |
| Target/provenance | Exact Target retained; successor attributed as itself | Same, with journal correlation to actual executor and request | Preserve existing history |
| Prospective activation | Separate approved contract, enforcement and verification | Same plus complete gateway deployment/migration proof | Already effective |
| Historical migration | Only individually proven complete legacy fencing | Legacy credentials/paths must be removed or effectively fenced | Remains available when proof fails |

A fails qualification if any subprocess, queued command, cached credential or alternate API bypasses fencing.
Same Windows SID or shared credential possession is not exclusive authority.
B also requires an exclusive gateway-owned Git workspace: every filesystem/index/ref writer,
including P6–P9 local operations, must be isolated or mediated with no direct caller bypass.
Single serial dispatch must remain enforced across gateway restarts and replicas; GitHub
credential mediation alone cannot protect local Git state.
B is a research direction only: a central service without enforced single leadership is insufficient.
Gateway restart, replicas and delayed requests must not resurrect obsolete authority.
A/B cannot gain safety merely by naming the component a fence or gateway.
C preserves current safety obligations and accepts potentially indefinite blocked publication.

## 6. Authority, identity and invariants

Candidate recovery authorization names original transaction/Target, bounded remaining phases and named candidate.
It is distinct from task intake, architecture approval and original publication authorization.
Whether original authorization survives successor admission is an unresolved policy decision for a separate gate.
User intent cannot manufacture source-authored Release, old run linkage or provider capabilities.
No new InitialDispatch substitutes for succession of an already-owned transaction.
Future recoveryAttempt/claim/epoch are separate axes from existing transferAttempt
Unissued/Released/Accepted/Closed, authorization and ownership. RecoveryAuthorized alone
neither consumes/revokes the old gate nor certifies terminal closure. Successful admission
does not resolve the open permission carry-over policy.

| Invariant | Required design proof |
| --- | --- |
| I01 | At most one mutation-eligible owner for an exact transaction across all epochs; admission of a newer epoch requires enforced rejection of every prior epoch and reconciliation/drain of its in-flight operations. |
| I02 | Immutable Target and accepted scope are preserved; changes invalidate the applicable authority |
| I03 | New actor emits only its own attributed records; historical source records remain unchanged |
| I04 | Claim reservation alone grants no effect permission |
| I05 | Revoked authorization blocks new dispatch and queued execution; accepted requests are reconciled |
| I06 | Every effect is correlated to current authority, exact intent and operation outcome |
| I07 | Unknown outcome is not safe retry; proven completed effects are not replayed |
| I08 | P10 needs proven current owner, clean synchronized main and independent persisted terminal readback |

Fencing must survive claimant failure; cancellation does not silently restore the previous owner.
Irrecoverability is a bounded evidence finding, never inferred from timeouts or completed-run status alone.
Native identity includes actor, original history, owning run, route and actual execution principal.
Successor proof must also bind actual calls/results to its admitted context.

## 7. Candidate state machine

All names below are design labels, not operational events authorized for emission.
Candidate flow: ObservedBlocked → RecoveryAuthorized → ClaimReserved → PriorExecutorFenced →
EffectsReconciled → SuccessorAdmitted → RemainingExecution → TerminalVerified.

| State / actor | Required predecessor and evidence | Crash/reconciliation obligation |
| --- | --- | --- |
| ObservedBlocked / observer | Exact transaction, current complete chain, phase inventory | Missing lineage remains Unknown; do not choose a convenient owner |
| RecoveryAuthorized / authority recorder | Exact actual user recovery scope and candidate | Read durable authorization; no inferred old permission carry-over |
| ClaimReserved / qualified coordinator | Conditional CAS on exact predecessor/version and fresh claim identity | Lost response: inspect winning record; no blind second claim |
| PriorExecutorFenced / enforcement authority | Complete stale capability inventory and enforced rejection proof | Persist fence across coordinator/gateway restart; fail closed on partial fence |
| EffectsReconciled / independent verifier | Fence plus local/remote request inventory and authoritative outcomes | Unknown/inconsistent request blocks successor execution |
| SuccessorAdmitted / admission authority | Unchanged authorization/Target, same claim, proven fence and reconciled effects | Conditional durable admission is proposed ownership point; independent readback required |
| RemainingExecution / admitted successor | Persisted admission plus fresh phase-specific checks | Revalidate authority and enforcement for every operation; never replay completed phases |
| TerminalVerified / current successor and reader | Proven owned predecessor, required phase facts and terminal persistence | Lost terminal response: read exact record, do not synthesize duplicate closure |

Reservation does not fence; fencing does not prove prior external outcomes; reconciliation does not confer ownership.
If ledger and enforcement live in different systems, define durable partial states and recovery ordering explicitly.
No state advances on a lost result, stale prefix, failed readback or unqualified identity.
A crash after fencing must leave prior authority closed while the claim is reconciled.
A crash after admission must distinguish committed ownership from unissued admission.
The admission CAS must atomically validate exact transaction/Target, recovery authorization and its version,
revocation disposition and predecessor claim/version within its proven serialization domain.
Revocation and Target change must invalidate queued work at execution as well as at enqueue time.
Already accepted GitHub operations may still complete: block dependent execution until their outcomes are known.

## 8. Operation-level crash and idempotence boundaries

| Operation family | Required before candidate execution | Uncertain completion treatment |
| --- | --- | --- |
| Push / PR creation | Exact refs/head and existing PR discovery; authorized range | Inspect exact remote ref/PR/request; do not create duplicates |
| Merge | Fresh exact head/base, required checks and permitted strategy | Read PR merge outcome and exact merge commit before any retry |
| Remote ref deletion | Exact task ref still points to authorized head; no recreated ref | Compare current OID and operation semantics; unrelated/ref-recreated deletion forbidden |
| Local switch/pull/delete | Exact repository state, local intent and allowed phase | Inspect refs/index/worktree; never reset/rebase to manufacture expected state |
| Terminal record | Exact newest owner predecessor and all terminal prerequisites | Read independently persisted identity and resulting disposition |

Local Git mutation and remote GitHub acceptance have separate durability domains.
A journal records observed facts and uncertainties; it cannot make external effects atomic.
Gateway single-flight sequencing must account for requests already accepted before leadership changed.
Idempotent endpoint behavior must be documented for the exact operation, not assumed from similar APIs.
No phase completion is inferred solely from branch absence or a successful earlier HTTP connection.

## 9. Adversarial trace matrix

These are mandatory design-level traces; no live operation is executed by this document.
A/B PASS below means a future candidate rejects unsafe execution, not present implementation success.

| ID | Scenario / candidates | Required result |
| --- | --- | --- |
| T01 | Two successors race same predecessor / A,B | One conditional winner, loser never mutation-eligible |
| T02 | No supported CAS / A,B,C | A/B unavailable; retain C |
| T03 | CAS succeeds but old executor retains credentials / A,B | Reject admission; ledger is not fencing |
| T04 | Old run resumes after claim / A,B | Every stale mutation path rejects |
| T05 | Subprocess or cached credential bypass / A,B | Complete mediation proof fails; STOP |
| T06 | Gateway replicas partition / B | Fenced leadership rejects stale replica, otherwise STOP |
| T07 | Crash before reservation commit / A,B | No ownership; exact readback before reissue |
| T08 | Lost CAS success response / A,B | Discover exact winner, no duplicate claim |
| T09 | Crash after reservation before fence / A,B | Reserved claim grants no execution |
| T10 | Crash during partial fence / A,B | Fail closed; no successor execution |
| T11 | Crash after fence before reconciliation / A,B | Fence remains enforced; reconcile known claim |
| T12 | Lost admission response / A,B | Independent exact readback decides ownership |
| T13 | ABA / reused predecessor or claim / A,B | Monotonic version and unique identities reject reuse |
| T14 | Merge accepted, response lost / A,B,C | Inspect exact PR/merge outcome, never blind remerge |
| T15 | Gateway queued intent survives revoke / B | Execution rejects; accepted external request separately reconciled |
| T16 | Revoke races claim/admission / A,B | Defined ordering rejects authority after revoke |
| T17 | Revoke races already accepted API / A,B,C | Preserve uncertain effect; no false no-effects assertion |
| T18 | Immutable Target changes after enqueue / A,B | Reject old intent/admission; no retargeting |
| T19 | Delayed old Git push / A,B | Fence or drain with outcome evidence before remaining work |
| T20 | TASK-095 post-merge fixture / A,B,C | Preserve completed P0–P4; no merge replay; current STOP until proof |
| T21 | TASK-096 pre-merge fixture / A,B,C | Preserve P0–P3; fresh P4 facts required; no merge in this task |
| T22 | New policy approved without deployment / A,B,C | No activation or owner; retain C |
| T23 | Legacy capabilities cannot be fenced / A,B,C | Historical migration rejected; retain C |
| T24 | P10 requested without current owner / A,B,C | Reject terminalization |
| T25 | Terminal append response lost / A,B,C | Independent persisted readback; no duplicate terminal event |
| T26 | Clock expiry/TTL or completed run only / A,B | Insufficient loss/fence proof; STOP |
| T27 | Remote task ref recreated at other OID / A,B,C | Do not delete replacement ref; reconciliation required |
| T28 | Observer fabricates source Release / A,B,C | Reject provenance; no ownership transition |
| T29 | Gateway journal committed, API not known / B | Intent remains uncertain until authoritative reconciliation |
| T30 | Claim cancelled after old executor fenced / A,B | No automatic return of old owner/capability |

## 10. Prospective activation and historical migration

Default is prospective-only activation after separately approved normative and implementation stages.
Required before activation: authoritative capability qualification, architecture decision, contract alignment,
complete enforcement implementation, independent crash/concurrency proofs and explicit activation scope.
Documentation Acceptance, commit or publication alone satisfies none of these operational conditions.
New authoring authority does not transfer either fixture's publication permission.

Historical migration requires an individually authorized exact transaction and complete attempt/request inventory.
Retain original records and dispositions unchanged; new actor uses its own explicit attribution.
Prove complete old capability fencing, reconcile every potentially in-flight effect and qualify successor admission.
If legacy shared credentials, queued requests or missing history defeat proof, migration stays blocked.
No retrospective Release, synthetic owner continuity or historical terminal rewrite is allowed.
Offline fixture success cannot authorize live TASK-095/096 recovery.

## 11. Candidate evidence record fields

This is a schema inventory for later design, not a new native event format.
Each record must distinguish logical transition identity from physical persisted item identity.

| Field group | Required candidate binding |
| --- | --- |
| Transaction | Original gate/Target identity, accepted scope, ordered range and fixed base |
| Authority | Actual scoped recovery authorization, version, revocation and actor attribution |
| Claim | Exact predecessor/version, unique candidate claim and conditional operation outcome |
| Enforcement | Fence epoch, covered paths/capabilities, durable enforcement and drain evidence |
| Effects | Exact intent/request identity, Git/GitHub outcome, observation cut and uncertainty |
| Readback | Independent full persisted payload identity, digest, locator and complete current tail |

Unknown or contradictory fields prevent admission; timestamps do not repair missing causality.

## 12. Evidence, acceptance and next gates

A later provider investigation must record official source/version, supported interface, exact guarantees,
serialization domain, enforcement coverage, negative cases and reproducible read-only evidence.
Missing evidence remains Unknown; absence of records is not proven absence of actions.
Tests of a model cannot certify provider guarantees or exclusive credential mediation.
No external provider capability was fetched, configured or exercised for this Draft.

Design acceptance requires mirrored A/B/C coverage, complete capability/evidence matrix, explicit authority axes,
state predecessors, crash windows, adversarial outcomes and legacy limitations.
Independent Architecture Confirmation, Verification/Tester, non-author Review and Scope Audit are task gates.
Their acceptance assesses this bounded document; it does not approve an operational contract.
Impact inventory for any future normative stage: PROCESS-001/002, AGENT, Publisher/Coordinator roles,
recovery/Publisher scenarios, task template and mirrored process guides; changes require separate agreement.
Implementation/tooling, credentials, provider configuration and live migration remain outside current scope.
Next candidate: evidence research or architecture decision, Not Activated, without an automatically assigned task.
STOP at unknown enforcement, incomplete lineage, competing claim, stale predecessor, Target conflict,
lost unqualified outcome or any demand to impersonate a historical source.

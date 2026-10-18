# DP-026: Fenced Mutation Gateway

[Russian version](../../ru/design/DP-026-fenced-mutation-gateway.md)

## 1. Status and authority

- **Design Status:** Draft
- **Implementation Status:** Planned (not implemented)
- **Task:** [TASK-098](../../tasks/TASK-098-FENCED-MUTATION-GATEWAY-DESIGN.md)
- **Parent research:** [DP-025](DP-025-controlled-publisher-recovery.md), unchanged Draft / Planned.

This document records the Architect's candidate B design for a new project-owner requirement,
not a confirmed governance gap. It does not amend PROCESS-001/002 or agent contracts.
C, existing fail-closed STOP, remains the only operative failed-owner baseline.
A Direct Publisher remains UNKNOWN for complete fencing; B is CONDITIONALLY FEASIBLE
for prospective research under the proof obligations below, not ready for deployment.
All state names, fields, permits and transitions here are proposed design labels,
not native event formats or permission to emit events. Design acceptance/publication
does not activate DP-025/026, restore an old run or authorize a recovery.

## 2. Problem, threat model and non-goals

A failed Publisher may leave credentials, child processes, queued commands and accepted
Git/GitHub requests alive. The new design separates user authority, native ownership,
claim reservation, gateway admission, actual mutation eligibility and terminal outcome.
A native CAS or epoch read is not fencing of external mutations. In particular,
checking epoch before an API call does not prevent a paused executor calling later.

Threats include agent/gateway revival, competing successors, stale reads, retained
credentials, ABA, response loss, partitions, process/host failure and backup rollback.
The candidate assumes a trusted host administrator, credential authority and authenticated
admission issuer; protection against their compromise is outside the stated model.
A future design must reject activation if these assumptions cannot be sustained.

Safety means exclusive authorized mutation, preserved provenance and no blind replay.
Liveness is conditional: irreducible uncertainty may block forever. Timeouts, lease expiry,
completed-run status and absence of a record do not prove absence of effects.
Non-goals: implementation, provider configuration, mutation PoC, live TASK-095/096 exercise,
normative amendment, automatic HA, retrospective source Release and historical rewrite.
Runtime DP-023/024 local models do not qualify Publisher provider capabilities.

## 3. Evidence and provider capability ledger

Repository precedence is unchanged: [PROCESS-001](../../engineering/PROCESS-001-AI-DEVELOPMENT-WORKFLOW.md)
and [Publisher](../../engineering/agents/publisher.md) govern native ownership and P0–P10;
[PROCESS-002](../../engineering/PROCESS-002-DOCUMENTATION-SYNCHRONIZATION.md) governs durable navigation.
PROCESS-001 owning-context tuple (1208 onward), source-authored handoff (1484 onward),
canonical store/readback (1559 onward) and terminal owner prerequisites (1714 onward)
are existing contracts, not gateway capabilities. Line anchors describe this task's baseline;
section names remain the interpretation boundary.

This ledger preserves bounded prior Provider Capability Assessment observations.
It does not claim fresh deployment qualification. PROVEN is a narrow documented
guarantee or bounded observation; NOT SUPPORTED concerns the named interface only;
UNKNOWN means evidence missing, not a claim that a feature is absent.

| ID | Authoritative source / observation | Qualified domain and limit | Deployment qualification |
| --- | --- | --- | --- |
| G01 | [Codex App Server](https://learn.chatgpt.com/docs/app-server): thread/read, thread/resume, turn/interrupt | Reading/reopening/cancelling are not documented restoration of the original owning execution binding or external request fencing | UNKNOWN; exact supported endpoint/version and identity bridge required |
| G02 | [git-update-ref](https://git-scm.com/docs/git-update-ref): old OID and transactions | Conditional local refs; concurrent multi-ref readers may observe a subset, not one global snapshot | UNKNOWN for gateway authority store and filesystem fencing |
| G03 | [git-push](https://git-scm.com/docs/git-push): explicit expected-OID lease, atomic if server supports it | Ref transport predicates only; no authorization to force-push, no API/credential fence | UNKNOWN exact-server atomic capability |
| G04 | [GitHub GraphQL Git](https://docs.github.com/en/graphql/reference/git): updateRefs / beforeOid / afterOid | Documented atomic multi-ref conditional update/delete, not PR creation/merge or native events | UNKNOWN end-to-end deployment; read-only schema observation 192e51, 2026-10-17T22:45:29Z, is bounded evidence |
| G05 | [REST best practices](https://docs.github.com/en/rest/using-the-rest-api/best-practices-for-using-the-rest-api) | Generic unsafe-method conditional requests NOT SUPPORTED unless endpoint explicitly provides them; ETag GET is not write CAS | No generic conditional mutation guarantee |
| G06 | [PR API](https://docs.github.com/en/rest/pulls/pulls), merge head predicate | Expected head is not Publisher epoch, owner or fixed-base CAS; async merge request UUID/status is documented but not qualified here | UNKNOWN deployment and lost-response correlation |
| G07 | [GraphQL mutations](https://docs.github.com/en/graphql/reference/mutations), clientMutationId | Correlation does not establish deduplication or exactly-once PR creation | UNKNOWN exact operation retry guarantee |
| G08 | [REST refs](https://docs.github.com/en/rest/git/refs), ref deletion | REST DELETE has no expected-OID input; read-then-delete has a race; G04 is a conditional candidate | UNKNOWN adopted delete path; no mutation test |
| G09 | [GitHub App installation token](https://docs.github.com/en/rest/apps/installations) | Credential revocation is distinct from cancelling accepted requests or fencing every cached/alternate credential | UNKNOWN complete capability revocation/drain |
| G10 | [SQLite isolation](https://www.sqlite.org/isolation.html), [atomic commit](https://www.sqlite.org/atomiccommit.html), [WAL](https://www.sqlite.org/wal.html) | Candidate transactional store only; filesystem/configuration assumptions, WAL not a network-filesystem protocol | UNKNOWN selected store/configuration; no external cross-system transaction |
| G11 | Existing native persisted readback and trusted route contracts | Actual actor/history/run/route/principal and independently read-back items, not SID/sessionId alone | UNKNOWN new gateway issuer/native bridge and conditional append |

Latest documentation and schema do not certify installed-client compatibility.
The assessment observed Desktop 0.162.0-alpha.17.2 and npm CLI 0.128.0; these are historical
version observations, not an asserted current pair. No App Server is started/resumed here.
Native identity, durable readback, store CAS, linearizability, old-executor fence,
trusted admission and crash recovery each need separate evidence. Git/GitHub conditional
refs do not certify the other domains. No universal exactly-once, immutable storage,
atomic external acceptance or automatic cancellation guarantee is assumed.

## 4. Trust boundary and complete mutation-path inventory

The candidate is one single-active gateway holding mutation credentials and a dedicated
Git workspace, index and refs. Callers submit authenticated intents, not shell commands,
arbitrary URLs or credentials. The gateway alone validates, records, dispatches and reconciles.
Read-only observers have no write capability. Credential and filesystem isolation are both
required: API mediation alone does not protect checkout or local refs.

The coverage inventory must close Git CLI, REST, GraphQL, browser/MCP, hooks, subprocesses,
automation, alternate workspaces, cached/shared credentials and direct native emitters.
A privileged bypass is either removed or explicitly inside the trusted enforcement boundary.
Unknown bypass coverage rejects activation. Hooks/child processes need bounded approved
behavior and lifetime tracking; disabling an observed path does not prove no alternative exists.

| Path | Executor / credential boundary | Predicate and observable effect | In-flight, retry and fencing obligation |
| --- | --- | --- | --- |
| Push | Gateway Git transport credential / remote ref policy | Exact Target range and task ref; remote OID equals authorized head | Delayed old push must be fenced or drained; uncertain reply requires exact ref/request readback |
| PR creation | Gateway API credential / PR-write scope | Discover exact repo/head/base first; record unique PR ID and exact head/base | Lost reply is not safe create retry; clientMutationId is correlation only |
| Merge | Gateway API credential / contents-write and effective rules | Fresh head/base/checks, MERGEABLE / CLEAN, permitted merge commit; observe MERGED SHA/parents | Expected head is not owner/base CAS; accepted request may finish after revoke; reconcile before retry |
| Remote ref deletion | Gateway Git/API credential / exact ref | Authorized head still current; conditional G04 candidate or separately qualified path; observe absence | Recreated/moved ref must survive; REST read/delete race cannot be called fencing |
| Fetch/prune | Gateway remote-read credential plus exclusive local writer | Exact origin; objects, tracking refs and FETCH_HEAD observed separately | Local writes are effects; partial state and subprocess lifetime require reconciliation |
| Checkout main | Gateway exclusive workspace/index/HEAD writer | Exact clean preconditions; observe main, index and worktree | Git locks do not fence every writer; no caller or stale gateway filesystem access |
| Fast-forward | Gateway transport plus exclusive local writer | Only ff-only, allowed ancestry; observe main/origin/main and worktree | Fetch and local update are different crash domains; never reset/rebase to repair |
| Local branch deletion | Gateway exclusive ref writer | Exact task branch, merged proof, safe -d only; observe absence | Recreated/moved ref blocks delete; ref lock is not global ownership |
| Terminal event | Future approved attributed native transport / authenticated issuer | Proven current native owner, P9 and exact terminal payload; independent persisted readback | Journal cannot fabricate Publisher event; lost append needs exact readback before any retry |

The inventory includes all P0–P10 writers, not merely remote operations. No gateway credential
is delivered to agents. The final credential choice, scopes, issuance, rotation, revocation,
secret storage and process isolation remain UNKNOWN until separately qualified.
A host path or SID is not an authenticated caller identity or exclusive mutation authority.

## 5. Identity, user authority and immutable Target

Admission must bind actual actor, original history/branch identity, native owning run,
authenticated route, actual principal and gateway caller identity. Trusted transport/issuer
evidence must establish these fields; a client-supplied tuple is only a claim.
Read-only transcript access, same conversation/session, same SID or completed run does not
establish continuity. The gateway cannot impersonate the old Publisher.

Use independent axes: scoped user recovery authorization and version; revocation version;
native ownership; unique claim/predecessor version; fence evidence; operation state;
terminal disposition. Old publication permission never automatically transfers.
A future approved carry-over policy and explicit scoped recovery authorization are required;
this Draft resolves neither by silently adopting historical permissions.

Immutable Target includes publication class/task/repository, exact branch, fixed base,
ordered commits, head/tree, accepted manifest and acceptance identity. Every intent
binds exact payload, phase and Target digest. Target change invalidates affected authority
and queued work; there is no implicit retargeting or source commit rewrite.
Original records remain immutable; successors emit only self-attributed records with
original/replacement provenance and explicit observation cuts.

## 6. Candidate state machine, claim and admission

All transitions require independent authoritative readback; Unknown or conflict blocks
execution at every state. The following proposed admission point does not change today's
InitialDispatch / Release / Route / Accept ownership protocol.

| State / actor | Exact predecessor and proof | Crash / rejection rule |
| --- | --- | --- |
| BlockedObserved / observer | Complete transaction chain, exact Target, last proven phase and operation inventory | Missing lineage remains UNKNOWN; select no convenient owner |
| Authorized / authority recorder | Actual scoped user recovery authorization, candidate, authority/revoke versions | No inferred permission carry-over; revoke or Target change rejects |
| ClaimReserved / conditional store | Unique claim plus expected predecessor/version, Target, authority and revoke versions in one qualified serialization domain | Reservation is not permission; one winner only, lost reply read back |
| PriorExecutorFenced / enforcer | Proven rejection of all prior capabilities, children, queues and stale gateway instances; accepted requests inventoried | Partial fence, retained credential or missing request blocks admission |
| EffectsReconciled / reconciler | Authoritative outcomes/drain of every potentially in-flight request and partial local effect | Uncertain remains blocked; completed effect never replayed |
| SuccessorAdmitted / admission issuer | Conditional commit validates unchanged claim, fence, outcome inventory, Target and authority/revoke versions | Proposed linearization point: durable admission commit plus independent readback; native bridge requires future normative approval |
| Executing / gateway | Exact current admitted epoch/context, fresh phase prerequisites and durable operation intent | Stale caller/queued intent rejects; inspect-first after interruption |
| TerminalVerified / attributed emitter and reader | Current native owner, all phases proven, clean synchronized main, independently persisted terminal | No ownerless P10 or fabricated source event; uncertain append reconciles |

Unique claim identity plus monotonic version must prevent ABA. An OID-only predecessor check
does not prove no intermediate changes. Claim/admission/revoke order is atomic only inside
the qualified store's serialization domain, never across GitHub and the native provider.
Cancelling a claim after fencing does not resurrect the prior owner or unfreeze credentials.

Agent failure while the gateway remains proven active differs from gateway failure.
The former may reject stale agent intents and reconcile its existing journal before a
separately authorized successor; the latter requires proof that the old gateway, children,
credentials and accepted requests are fenced/drained before any new dispatch.
No automatic failover or TTL-based takeover is proposed.

## 7. Durable journal and revoke/dispatch ordering

The proposed journal is authoritative only after store, durability, rollback and access
controls are qualified. Records distinguish logical operation ID from physical record ID.
Required fields: transaction/Target digest, scoped authorization/version, native actor/context,
claim/predecessor, epoch/incarnation, phase, exact request payload/digest, preconditions,
intent ID, dispatch/result state, provider request locator, observed cut, result digest,
uncertainty and independent readback locator. Secrets must not enter the journal.

| Journal state | Required durable meaning | Recovery action |
| --- | --- | --- |
| IntentRecorded | Exact operation proposed, no proven dispatch decision | Revalidate current authority; never infer effect from intent |
| DispatchReserved | Serialized authority/revoke/epoch/precondition decision | No independent worker may dispatch after authority changes |
| MayHaveDispatched | Persist before external call, including the crash-before-call window | Treat as potentially accepted; never auto-resend on restart |
| OutcomeVerified | Exact authoritative outcome linked to intent/request | Preserve completed effect or documented rejection |
| Finalized | Outcome and evidence durably recorded | Advance only the proven next phase |
| RejectedBeforeDispatch / Uncertain | Proven rejection before send / insufficient evidence distinguished | First may be re-planned under fresh authority; second blocks conflicting work |

**Current authority rule:** an actual observed native user revoke invalidates current
authority and requires STOP under existing contracts. Neither DispatchReserved nor
MayHaveDispatched grants entitlement to send after that revoke. Neither state proves
external acceptance; a reservation is not a request already accepted by GitHub.
This Draft cannot reinterpret or delay the effect of a current actual user revoke.

**Unapproved candidate policy — activation blocker:** a future gateway could propose a
durable RevokeCommitted record in the same execution serialization domain as its revoke
linearization point. Trusted receipt/message arrival, effective commit and acknowledgement
would then be distinct. Acknowledgement would follow commit and independent readback only.
This receipt-versus-commit authority policy is NOT decided or approved here; it requires
separate explicit user/normative approval before implementation or activation.

Under that conditional proposal only, RevokeCommitted before dispatch would reject queued/new
work; dispatch before RevokeCommitted might have an accepted external request to reconcile.
A gateway paused before sending while holding the execution lock could leave a trusted revoke
receipt pending indefinitely: that receipt is neither a committed revoke nor proof of an
accepted request. It must not be reported as completed revocation, external acceptance or
terminal closure. Reservation alone supplies no post-effective-revoke send permission.
No atomic transaction joins gateway revocation with external API acceptance.

If policy instead requires immediate effect at trusted receipt, this candidate B does not
prove that property without a stronger enforced external fence; admission stays blocked and
C remains operative. The execution lock must cover the dispatch gap: releasing it after a
token check and before the call is insufficient. Holding it through a call can indefinitely
delay the proposed commit/acknowledgement, a disclosed liveness cost, not an exception to the
current authority rule. Any uncertain reserved operation requires drain/reconciliation;
it cannot be called absent or automatically resent.

Crash after persisting MayHaveDispatched but before sending is observationally ambiguous
with accepted-call/response-loss. Journal ordering deliberately sacrifices automatic retry.
Transport cancellation, timeout and credential revoke are not proof that a request was
not accepted. Proven terminal/ref/PR state must be linked to the intent; absence of records
is weaker than authoritative no-effects evidence.

## 8. P0–P10 phase-specific obligations

The gateway is a proposed executor boundary, not permission to skip existing Publisher gates.
Every phase retains current authority/context checks and exact operation journal linkage.
Only merge commit is studied; this Draft grants no squash/rebase/force permissions.

| Phase | Fresh prerequisites / executor | Proven completion and recovery boundary |
| --- | --- | --- |
| P0 | Native owner and proposed admission bridge, trusted identity/route, API+Git probes, exact Target/branch/base/range, permissions and effective rules | Full current P0 proof, not collector capability alone; context change stops |
| P1 | Gateway exact authorized head/ref, no conflicting remote effect | Remote OID equals Target; discover existing completion before push |
| P2 | Gateway exact head/base/repo PR discovery | One matching PR ID/URL; ambiguous creation reconciles before retry |
| P3 | Gateway/observer fresh checks and conflict inspection | Required Success or No CI distinct from CI PASS; Pending/Failed block; No CI requires MERGEABLE / CLEAN |
| P4 | Gateway fresh head/base/checks, MERGEABLE / CLEAN and effective merge-commit policy | MERGED plus exact merge SHA/parents; accepted/unknown response inspected, never blind remerge |
| P5 | Gateway exact remote task ref still at authorized head; separate delete and fetch/prune intents | Deletion/absence and refreshed refs each proven; moved/recreated ref not removed |
| P6 | Gateway exclusive local workspace, clean phase facts | Current main with exact index/worktree; partial checkout reconciled |
| P7 | Gateway ff-only and ancestry proof | Synchronized main/origin/main; partial transport/local update separately inspected |
| P8 | Gateway merged exact local task branch, safe -d only | Exact local ref absence; no glob/-D/cleanup of other tasks |
| P9 | Gateway/independent observer fresh main/ref/worktree facts | main == origin/main == authoritative remote main, task refs absent, clean index/worktree |
| P10 | Proven current native owner and all terminal prerequisites; future approved self-attributed native emission | Persisted terminal plus independent exact readback; gateway journal alone is not P10 |

P5–P9 failures never undo P4. Local mutation and GitHub acceptance have separate durability
domains. A fresh read may qualify an outcome but does not restore a lost native owner.
Terminal execution remains blocked until the future gateway/native ownership bridge is
explicitly approved, implemented and qualified.

## 9. Crash consistency, rollback and HA limits

| Crash point | Safe candidate continuation |
| --- | --- |
| Before durable authorization/claim | No new owner; inspect exact authoritative history |
| After reservation before fence | Keep non-executing reservation; no effect permission |
| During partial fence | STOP; finish proving fencing, never assume completion |
| After fence before reconciliation | Preserve fence; reconcile known claim, never restore old owner |
| After intent/reservation before call | Distinguish proven not dispatched from MayHaveDispatched; uncertain blocks |
| After external acceptance before reply | Discover exact request/effect; no automatic resend |
| After reply before journal completion | Reconstruct exact outcome and bind it to intent |
| After admission before acknowledgement | Read exact winning admission; no second owner |
| After terminal append before receipt | Independent exact persisted readback; no synthetic duplicate |
| Restart from old backup | STOP until non-rollback authority/fence and current external outcomes are proven |

A same-store epoch counter can roll back with the backup. A hash chain detects some
modifications but does not prove freshness or prevent replay of a complete old snapshot.
Candidate restart therefore requires an independently trusted non-rollback high-water mark,
unique incarnation and external fence evidence, or another separately proven equivalent.
No such provider mechanism is qualified here: UNKNOWN blocks reactivation.

Single-active is an enforcement property, not a process name. A stale gateway replica with
credentials or writable workspace is an unfenced executor. Supervisor, child lifetime,
credential revocation and accepted-request drain need proof before restart.
HA, automatic leader replacement and partition tolerance require a separate protocol;
a DB lease/TTL alone is insufficient. Initial single-active design is an explicit SPOF.
Provider outage preserves uncertain outcomes and may stop recovery indefinitely.
Local Git locks, store transactions and remote conditional refs have different domains;
none makes the entire journal/Git/GitHub/native pipeline atomic.

## 10. Formal invariant coverage

These are design proof obligations, not runtime test results.

| ID | DP-025 invariant | Candidate mechanism / proof and rejection |
| --- | --- | --- |
| I01 | At most one actual mutation-eligible owner across epochs | Exclusive boundary + prior capability fence/drain + conditional admission; any bypass/stale gateway rejects |
| I02 | Immutable Target and accepted scope | Digest in authority, claim, journal and per-phase predicate; mismatch rejects queued/executing intent |
| I03 | Self-attribution and preserved history | Authenticated issuer/native bridge, immutable original locators; fabricated Release/continuity rejects |
| I04 | Reservation is not permission | ClaimReserved has no dispatch transition until fence/reconciliation/admission/readback |
| I05 | Revoke blocks new/queued execution; accepted requests reconciled | Current native revoke requires STOP; future receipt/commit policy remains unapproved; serialized execution and exact in-flight proof, never reservation-as-acceptance |
| I06 | Every effect linked to current authority, intent and outcome | Durable journal, authenticated execution linkage and independent authoritative result |
| I07 | Unknown is not retry; completed effects not replayed | MayHaveDispatched blocks resend; operation-specific reconciliation |
| I08 | P10 needs current owner, clean synchronized main and terminal readback | P9 plus native bridge and independent persisted terminal receipt; missing proof rejects |

For formal/model work, safety properties must quantify over stale agents, gateway incarnations,
queues, subprocesses and accepted requests, not just ledger rows. Liveness claims require
stated fairness, provider availability and outcome observability assumptions.
Manual trace coverage demonstrates document consistency only, not universal provider behavior.

## 11. Adversarial traces T01–T30

Each row specifies a future proof/rejection obligation; none reports a live test PASS.

| ID | Scenario and state/predecessor | Invariants | Expected result / evidence |
| --- | --- | --- | --- |
| T01 | Competing ClaimReserved on same predecessor | I01,I04 | One conditional winner; loser cannot dispatch; exact winning readback |
| T02 | Claim CAS unavailable / Authorized | I01,I04 | B admission unavailable; retain C, capability UNKNOWN |
| T03 | Claim succeeds, old credentials retained / fence | I01,I04 | Reject admission; complete revocation/bypass proof required |
| T04 | Old run revives after claim / Executing | I01,I06 | Stale caller/capability rejected at actual mutation boundary |
| T05 | Child/cached credential bypass / fence | I01 | Mediation proof fails; STOP until complete inventory/fence |
| T06 | Gateway replicas partition / admission | I01 | Stale replica fenced or STOP; no lease-only HA |
| T07 | Crash before claim commit / Authorized | I04,I07 | No owner inferred; exact store readback before reissue |
| T08 | Lost successful claim reply / ClaimReserved | I01,I07 | Read unique winner/version, no duplicate claim |
| T09 | Crash after reservation before fence | I04 | Reservation remains non-executing |
| T10 | Crash during partial fence | I01,I04 | Unknown fence blocks successor; persist covered/uncovered paths |
| T11 | Crash after fence before reconciliation | I01,I07 | Preserve fence, resume exact inventory; no old-owner restoration |
| T12 | Lost admission reply / SuccessorAdmitted | I01,I06 | Independent winning admission readback determines candidate |
| T13 | ABA/reused claim or predecessor | I01,I02 | Unique identity + monotonic non-rollback version reject reuse |
| T14 | Merge accepted, reply lost / P4 uncertain | I06,I07 | Exact PR/request/merge SHA reconciliation, no blind merge |
| T15 | Queued intent survives revoke / dispatch | I05 | Current actual native revoke requires STOP; reject unaccepted work; reservation/MayHaveDispatched is not acceptance or post-revoke send permission; future receipt/commit policy needs separate approval |
| T16 | Revoke races claim/admission | I01,I05 | Serialization order/authority version rejects post-revoke admission |
| T17 | Revoke races accepted API request | I05,I07 | Preserve uncertain effect; no fabricated no-effects claim |
| T18 | Target changes after enqueue | I02,I05 | Reject stale intent/admission; no silent retarget |
| T19 | Delayed old push / fence/drain | I01,I06,I07 | Proven rejection or exact outcome before remaining work |
| T20 | TASK-095 post-merge offline fixture | I02,I07,I08 | Keep P0–P4, do not replay merge; P5–P10 blocked without proof |
| T21 | TASK-096 pre-merge offline fixture | I02,I07 | Keep P0–P3, fresh P4 facts/admission required; no live merge |
| T22 | Approved design without deployment | I01,I04 | No operational owner/activation; retain C |
| T23 | Legacy capabilities cannot be fenced | I01,I03 | Historical migration rejected; retain C |
| T24 | P10 without proven current owner | I08 | Reject terminalization even if external merge completed |
| T25 | Terminal append reply lost | I03,I07,I08 | Read exact persisted item independently, no duplicate terminal |
| T26 | TTL/completed run only as loss proof | I01,I04 | Insufficient fence/drain evidence; STOP |
| T27 | Task ref recreated at different OID / P5 | I02,I07 | Do not delete new ref; exact conditional predicate/reconcile |
| T28 | Observer fabricates source Release | I03 | Reject provenance; no ownership transition |
| T29 | Journal committed, API outcome unknown | I06,I07 | Preserve MayHaveDispatched/Uncertain; no automatic resend |
| T30 | Claim cancelled after old executor fenced | I01,I04 | Do not restore old capability; new authorized disposition required |

Supplemental future traces must cover old-backup rollback, process pause inside the
dispatch gap and stale gateway restart; these extend, rather than replace, T01–T30.
Every future proof must record exact predecessor, request/epoch, observation cut,
authoritative result and limitation.

## 12. Activation, fixtures, PoCs and next gates

Default activation is prospective-only after separate architecture decision, normative
alignment, implementation, independent verification, deployment qualification and explicit
activation scope. Required impact inventory: PROCESS-001/002, AGENT, Coordinator/Publisher
contracts, general/Publisher recovery scenarios, task template and mirrored process guides.
This design task changes none of them. Existing native admission and terminal contracts
remain effective until an approved bridge exists; gateway admission cannot replace them.

Offline fixtures use [immutable TASK-096 evidence](https://github.com/dsdred/universal-websocket-platform/blob/82cf49627e54881599bb60a55f28c8e9c6b0c4db/docs/tasks/TASK-096-FAILED-OWNER-RECOVERY-ASSESSMENT.md)
and DP-025 E07–E10. TASK-095 PR #99 merge c8047fd08caa79480c2bc73b5c382f4aa5079db7
has historical P0–P4 proof and P5–P10 obligations. TASK-096 has historical P0–P3,
P4 blocker and remaining P4–P10; current PR/checks outcome is UNKNOWN here.
A specific command not executing is not proof that all later actions were absent.
No live fixture read/recovery, old permission reuse or terminal inference occurs.

Historical migration needs separate exact authorization, full original lineage/attempt/request
inventory, complete legacy credential/workspace/child fencing, authoritative in-flight outcomes,
an approved native bridge and successor self-attribution. Gateway deployment cannot retroactively
prove legacy mediation. Missing source history or unfenceable capability keeps that transaction
blocked. No synthetic Release, InitialDispatch, continuity or historical terminal rewrite.

| Proposed isolated PoC | Required evidence / STOP boundary |
| --- | --- |
| Conditional claim/ref contention | Concurrent winner, ABA, lost reply and exact version/readback; no universal guarantee from a sample |
| Credential/workspace isolation | Every enumerated bypass denied, including children/alternate workspace; unknown path blocks |
| Crash-point injection | Every section 9 boundary, uncertain dispatch and journal restart; never automatic resend |
| Revival/partition | Paused old agent/gateway and children cannot mutate; no automatic HA assumption |
| Backup rollback | Independent freshness/incarnation/fence evidence survives old snapshot; UNKNOWN blocks restart |
| Merge request UUID/status | Exact supported deployment, intent correlation, async outcome and lost response; not assumed cancellation |
| Conditional delete | Exact recreated-ref race rejected, documented predicate and response-loss reconciliation |
| Formal/model verification | I01–I08, T01–T30 plus rollback/dispatch-gap traces under explicit assumptions |

These PoCs require separate authorization and disposable resources. No mutation probe,
implementation or provider setup is performed here. Read-only research may narrow a
capability gap but cannot substitute for exclusive enforcement proof.

Design acceptance requires Architect Confirmation, independent Tester, non-author Review,
EN/RU parity, valid links, PROCESS-002, eight-path Scope Audit, canonical accepted manifest
and post-decision integrity. Manual design verification is not runtime/provider certification.
B remains conditional; A remains UNKNOWN; C remains active.
Next candidate is a separately agreed evidence/architecture gate, Not Activated, with no
automatic Task ID. STOP on unknown enforcement, stale predecessor, competing claim,
Target/provenance mismatch, uncertain outcome, bypass, rollback ambiguity or demand for
unapproved normative/runtime work.

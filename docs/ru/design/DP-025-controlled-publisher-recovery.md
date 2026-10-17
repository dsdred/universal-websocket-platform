# DP-025: Controlled Publisher Recovery

[English version](../../en/design/DP-025-controlled-publisher-recovery.md)

## 1. Статус и полномочия

- **Design Status:** Draft
- **Implementation Status:** Planned (не реализовано)
- **Task:** [TASK-097](../../tasks/TASK-097-CONTROLLED-PUBLISHER-RECOVERY-DESIGN.md)

Предложение исследует новое требование владельца проекта, не confirmed governance gap.
Architecture Confirmation разрешает bounded comparison, но не утверждает operational
recovery contract. Действующие PROCESS-001/002 и Publisher rules сохраняют силу.
Provider capabilities, ownership transitions, перенос permission и migration не активируются.
Слово «требуется» ниже задаёт candidate proof obligation, не действующее разрешение.

## 2. Проблема, границы и решение

Исследуется unfinished publication после невосстановимой потери original owning execution run.
Authorization, mutation ownership, recovery claim, operation outcomes и terminal state разделяются.
Недоступный run не доказывает неактивность subprocesses, credentials и принятых requests.
Safety означает отсутствие неразрешённых/concurrent mutations, повторов proven effects и выдуманной provenance.
Liveness означает eventual remaining-phase completion при явно названных доказанных условиях.
Постоянная неопределённость может правомерно препятствовать liveness.

Сравниваются A Direct Publisher с proven fencing, B Fenced Mutation Gateway и C current fail-closed STOP.
Рекомендация: C остаётся единственной active policy; B исследуется условно; A требует полного доказательства.
Это рекомендация для отдельных design decisions, не утверждение B и не отклонение текущих contracts.
Implementation, provider setup, live fixture exercise и нормативные amendments исключены из задачи.

## 3. Действующая семантика и evidence ledger

Repository contracts имеют приоритет над Draft. Historical evidence ограничен observed cut.
Native API и provider backing — views одного store, не два независимых ownership authorities.
Чтение transcript не восстанавливает owning execution context.

| Evidence | Источник и exact scope | Следствие / ограничение |
| --- | --- | --- |
| E01 | [PROCESS-001](../../engineering/PROCESS-001-AI-DEVELOPMENT-WORKFLOW.md), owning-context tuple и per-call linkage, строки 1208–1223 | Same SID, conversation или role сами по себе не доказывают continuity |
| E02 | PROCESS-001, Trusted-Context Release Handoff, от строки 1484 | Только proven source выпускает Release; затем exact user Route и durable destination Accept |
| E03 | PROCESS-001, canonical store/source emission, строки 1606–1670 | Reader receipt не создаёт source event; inaccessible provenance остаётся Unknown/STOP |
| E04 | PROCESS-001, terminal semantics, строки 1714–1736 | P10 требует непосредственно предшествующего Active/Owned(exact actor); protocol procedural, не machine lock |
| E05 | [Publisher](../../engineering/agents/publisher.md), Initial Trusted Route и Trusted-Context Ownership Handoff | Dispatch устанавливает initial owner только после independent persisted readback; handoff не создаёт новое publication permission |
| E06 | [PROCESS-002](../../engineering/PROCESS-002-DOCUMENTATION-SYNCHRONIZATION.md), Publication-State Synchronization, от строки 154 | Project navigation хранит stable facts, не live ownership или временные checkpoints |
| E07 | [TASK-096](https://github.com/dsdred/universal-websocket-platform/blob/82cf49627e54881599bb60a55f28c8e9c6b0c4db/docs/tasks/TASK-096-FAILED-OWNER-RECOVERY-ASSESSMENT.md), F096-01–F096-12 и decision matrix | Existing paths условны; missing evidence не становится governance gap автоматически |
| E08 | TASK-096 F096-05/F096-06 | TASK-095 PR #99 merged c8047fd08caa79480c2bc73b5c382f4aa5079db7; P0–P4 proven; конкретная P5 invocation не выполнялась |
| E09 | Historical TASK-096 native InitialDispatch msg_00e5a0948e1deb9d016ad39287fa6c81919a9fae6d99885cb7 | Owning run 01a14a6e-5244-7431-9057-d6acc18bef5f; historical initial binding, не current continuity |
| E10 | Historical TASK-096 BlockerReport msg_00e5a0948e1deb9d016ad3936236688191a3d7a062b10016e4 | P0–P3 completed; та P4 merge command не выполнялась; не вывод обо всех будущих effects |

E09/E10 — historical qualified-native locators, переданные design, не freshly probed remote facts.
TASK-096 target: 82cf49627e54881599bb60a55f28c8e9c6b0c4db; accepted manifest:
1e065935ded73cd69c89b7f60a8359488509697d; qualified base — merge OID из E08.
PR #100 был OPEN при последнем подтверждённом наблюдении; current PR/checks outcome здесь Unknown.
Remaining P5–P10 TASK-095 и remaining P4–P10 TASK-096 — только offline fixtures.
Их historical permissions этим proposal не используются и не передаются.

## 4. Матрица provider capabilities

Ни один qualified authoritative provider contract в repository не устанавливает следующие гарантии.
Каждый Unknown — обязательный результат исследования, не утверждение отсутствия feature.
Предпосылки: официально documented semantics, supported transport/version и observable failure behavior.
Тест может corroborate documented guarantee; успешный sample не устанавливает универсальную atomicity.
Локальные Runtime DP-023/024 containment/claim models не доказывают Publisher-provider capabilities.

| Capability | Требуемое evidence A | Требуемое evidence B | Текущая квалификация |
| --- | --- | --- | --- |
| Conditional claim / CAS | Exact predecessor/version, durable success/conflict semantics | То же для gateway authority и succession ledger | Unknown |
| Old-executor fencing | Каждый local/remote/subprocess/legacy/queued mutation path отклоняет old epoch | Exclusive mediated credentials и enforced single active gateway через replicas/crashes | Unknown |
| Durable ordering | Claims, revokes, admissions и calls имеют proven общий порядок | Durable sequencer/journal с recovery без stale leader writes | Unknown |
| In-flight outcome | Request identities и authoritative completion/absence evidence | То же; journal сам не отменяет external accepted request | Unknown |
| Lost response reconciliation | Readback определяет exact write/request без слепого повтора | Intent/result linkage сохраняется после response loss и process death | Unknown |
| Git/GitHub conditional effects | Documented exact-ref/head predicates и rejection semantics каждой operation | Gateway проверяет predicates и записывает outcomes; external atomicity всё ещё не доказана | Unknown |
| Credential mediation | Все cached/shared credentials и bypass paths закрыты | Direct caller/legacy credentials вне enforcement отсутствуют | Unknown |
| Clock/lease behavior | Expiry вместе с enforced revocation, не одно прошедшее время | Monotonic epochs и enforced fencing при pauses/partitions | Unknown |

Unique ledger claim, machine fencing и atomic external mutation — три разные properties.
CAS доказывает максимум один winning conditional ledger update в указанном serialization domain.
Он не останавливает old process, сохранивший Git/GitHub authority.
Token check перед external request оставляет check-then-act race.
Cross-system transaction, exactly-once GitHub execution и automatic cancellation не предполагаются.

## 5. Сравнение A/B/C

| Boundary | A — Direct Publisher | B — Fenced Mutation Gateway | C — current STOP |
| --- | --- | --- | --- |
| Provider assumptions | Полное enforcement каждого direct mutation path | Полностью mediated gateway с proven durable sequencing/fencing | Existing qualified native store и current contracts |
| Authorization | Explicit recovery scope и named successor; carry-over policy unresolved | То же; gateway упорядочивает authorization/revoke с admission и execution | Новое succession permission отсутствует |
| Admission | Proven fence, reconciliation и conditional successor claim | Gateway-enforced fence, reconciled journal и conditional successor admission | Только proven continuity либо existing source handoff |
| Ownership point | Durable conditional admission с independent readback | Та же точка в enforced gateway authority domain | Existing InitialDispatch / Accept rules |
| Old executor | Stale authority отклоняется на каждом effect boundary | Direct capabilities устранены; old gateway replicas и queued intents ограждены | Replacement owner не выводится |
| In-flight Git/GitHub | Inventory/drain/reconcile каждого old request | Durable intent/result journal и external reconciliation | Unknown effects сохраняют STOP |
| Crash consistency | Fence переживает restart; ledger/remote split явно восстанавливается | Intent persisted до execution; uncertain calls блокируют conflicting work | Новые writes отсутствуют |
| Competing successors | Exact predecessor CAS и enforced epoch, без ABA | Serialized conditional admission и fenced leadership | Никто не получает новое ownership |
| Lost CAS/API response | Читать exact winning claim/result до retry | Reconcile exact intent/request и current authority до retry | Только inspect |
| Target/provenance | Exact Target сохранён; successor attributed от собственного имени | То же, с journal correlation к actual executor и request | Existing history сохранена |
| Prospective activation | Отдельно approved contract, enforcement и verification | То же плюс complete gateway deployment/migration proof | Уже действует |
| Historical migration | Только individually proven complete legacy fencing | Legacy credentials/paths устранены либо effectively fenced | Доступен при недостаточном proof |

A не квалифицирован, если subprocess, queued command, cached credential или alternate API обходят fencing.
Same Windows SID или владение shared credential не означает exclusive authority.
B также требует exclusive gateway-owned Git workspace: каждый filesystem/index/ref writer,
включая local operations P6–P9, изолирован либо mediated без direct caller bypass.
Single serial dispatch сохраняется через gateway restarts и replicas; одно GitHub
credential mediation не защищает local Git state.
B — только research direction: central service без enforced single leadership недостаточен.
Gateway restart, replicas и delayed requests не должны восстанавливать obsolete authority.
Назвать component fence или gateway недостаточно для safety A/B.
C сохраняет действующие safety obligations и допускает indefinite blocked publication.

## 6. Authority, identity и invariants

Candidate recovery authorization называет original transaction/Target, bounded remaining phases и named candidate.
Она отличается от task intake, architecture approval и original publication authorization.
Сохранение original authorization после successor admission — открытое policy decision отдельного gate.
User intent не создаёт source-authored Release, old run linkage или provider capabilities.
Новый InitialDispatch не заменяет succession уже owned transaction.
Будущие recoveryAttempt/claim/epoch — отдельные axes относительно existing transferAttempt
Unissued/Released/Accepted/Closed, authorization и ownership. RecoveryAuthorized сам
не consumes/revokes old gate и не certifies terminal closure. Successful admission
не решает открытую permission carry-over policy.

| Invariant | Необходимое design proof |
| --- | --- |
| I01 | Не более одного mutation-eligible owner для exact transaction среди всех epochs; admission нового epoch требует enforced rejection каждого prior epoch и reconciliation/drain его in-flight operations. |
| I02 | Immutable Target и accepted scope сохранены; изменения invalidates applicable authority |
| I03 | Новый actor пишет только собственные attributed records; historical source records неизменны |
| I04 | Claim reservation сама не даёт effect permission |
| I05 | Revoked authorization блокирует новый dispatch и queued execution; accepted requests reconciled |
| I06 | Каждый effect связан с current authority, exact intent и operation outcome |
| I07 | Unknown outcome не разрешает retry; proven completed effects не повторяются |
| I08 | P10 требует proven current owner, clean synchronized main и independent persisted terminal readback |

Fencing переживает claimant failure; cancellation не возвращает previous owner автоматически.
Irrecoverability — bounded evidence finding, не следствие одних timeouts или completed-run status.
Native identity включает actor, original history, owning run, route и actual execution principal.
Successor proof также связывает actual calls/results с admitted context.

## 7. Candidate state machine

Все имена ниже — design labels, не operational events, разрешённые к emission.
Candidate flow: ObservedBlocked → RecoveryAuthorized → ClaimReserved → PriorExecutorFenced →
EffectsReconciled → SuccessorAdmitted → RemainingExecution → TerminalVerified.

| State / actor | Required predecessor и evidence | Crash/reconciliation obligation |
| --- | --- | --- |
| ObservedBlocked / observer | Exact transaction, complete current chain, phase inventory | Missing lineage остаётся Unknown; нельзя выбирать удобного owner |
| RecoveryAuthorized / authority recorder | Exact actual user recovery scope и candidate | Читать durable authorization; old permission carry-over не выводится |
| ClaimReserved / qualified coordinator | Conditional CAS по exact predecessor/version и fresh claim identity | Lost response: inspect winning record; без слепого второго claim |
| PriorExecutorFenced / enforcement authority | Полный stale capability inventory и enforced rejection proof | Сохранить fence через coordinator/gateway restart; partial fence fail closed |
| EffectsReconciled / independent verifier | Fence, local/remote request inventory и authoritative outcomes | Unknown/inconsistent request блокирует successor execution |
| SuccessorAdmitted / admission authority | Unchanged authorization/Target, same claim, proven fence и reconciled effects | Conditional durable admission — proposed ownership point; independent readback обязателен |
| RemainingExecution / admitted successor | Persisted admission и fresh phase-specific checks | Каждая operation revalidates authority/enforcement; completed phases не повторяются |
| TerminalVerified / current successor и reader | Proven owned predecessor, required phase facts и terminal persistence | Lost terminal response: читать exact record, не создавать duplicate closure |

Reservation не fencing; fencing не доказывает prior external outcomes; reconciliation не даёт ownership.
Если ledger и enforcement в разных системах, явно определить durable partial states и recovery ordering.
Lost result, stale prefix, failed readback и unqualified identity не разрешают state advance.
Crash после fencing сохраняет prior authority закрытой, пока claim проходит reconciliation.
Crash после admission требует различить committed ownership и unissued admission.
Admission CAS должен атомарно проверять exact transaction/Target, recovery authorization и её version,
revocation disposition и predecessor claim/version внутри proven serialization domain.
Revocation и Target change должны invalidates queued work при execution и при enqueue.
Уже принятые GitHub operations могут завершиться: dependent execution блокируется до известных outcomes.

## 8. Operation-level crash и idempotence boundaries

| Operation family | Что требуется до candidate execution | Обращение с uncertain completion |
| --- | --- | --- |
| Push / PR creation | Exact refs/head и discovery existing PR; authorized range | Inspect exact remote ref/PR/request; duplicate creation запрещён |
| Merge | Fresh exact head/base, required checks и permitted strategy | Читать PR merge outcome и exact merge commit до любого retry |
| Remote ref deletion | Exact task ref всё ещё указывает на authorized head; ref не recreated | Сверить current OID и operation semantics; чужой/recreated ref удалять нельзя |
| Local switch/pull/delete | Exact repository state, local intent и allowed phase | Inspect refs/index/worktree; не reset/rebase ради ожидаемого state |
| Terminal record | Exact newest owner predecessor и все terminal prerequisites | Independent readback persisted identity и resulting disposition |

Local Git mutation и remote GitHub acceptance имеют разные durability domains.
Journal записывает observed facts и uncertainties; он не делает external effects атомарными.
Gateway single-flight sequencing учитывает requests, принятые ещё до leadership change.
Idempotent endpoint behavior документируется для exact operation, не выводится из похожих API.
Branch absence или успешное прежнее HTTP connection сами по себе не доказывают phase completion.

## 9. Adversarial trace matrix

Это обязательные design-level traces; документ не выполняет live operations.
A/B PASS ниже означает отвержение unsafe execution будущим candidate, не успех текущей реализации.

| ID | Scenario / candidates | Required result |
| --- | --- | --- |
| T01 | Два successors race same predecessor / A,B | Один conditional winner, loser никогда не mutation-eligible |
| T02 | Нет supported CAS / A,B,C | A/B недоступны; сохранить C |
| T03 | CAS успешен, old executor сохраняет credentials / A,B | Reject admission; ledger не fencing |
| T04 | Old run resumes после claim / A,B | Каждый stale mutation path отклоняется |
| T05 | Subprocess или cached credential bypass / A,B | Complete mediation proof провален; STOP |
| T06 | Gateway replicas partition / B | Fenced leadership отклоняет stale replica, иначе STOP |
| T07 | Crash до reservation commit / A,B | Ownership отсутствует; exact readback до reissue |
| T08 | Lost CAS success response / A,B | Найти exact winner, без duplicate claim |
| T09 | Crash после reservation до fence / A,B | Reserved claim не разрешает execution |
| T10 | Crash во время partial fence / A,B | Fail closed; successor execution запрещён |
| T11 | Crash после fence до reconciliation / A,B | Fence остаётся enforced; reconcile known claim |
| T12 | Lost admission response / A,B | Independent exact readback определяет ownership |
| T13 | ABA / reused predecessor или claim / A,B | Monotonic version и unique identities отклоняют reuse |
| T14 | Merge принят, response потерян / A,B,C | Inspect exact PR/merge outcome, без blind remerge |
| T15 | Gateway queued intent пережил revoke / B | Execution отклоняется; accepted external request отдельно reconciled |
| T16 | Revoke races claim/admission / A,B | Defined ordering отклоняет authority после revoke |
| T17 | Revoke races already accepted API / A,B,C | Сохранить uncertain effect; без false no-effects assertion |
| T18 | Immutable Target меняется после enqueue / A,B | Reject old intent/admission; retargeting запрещён |
| T19 | Delayed old Git push / A,B | Fence либо drain с outcome evidence до remaining work |
| T20 | TASK-095 post-merge fixture / A,B,C | Сохранить completed P0–P4; merge не повторять; current STOP до proof |
| T21 | TASK-096 pre-merge fixture / A,B,C | Сохранить P0–P3; нужны fresh P4 facts; merge в задаче запрещён |
| T22 | New policy approved без deployment / A,B,C | Нет activation/owner; сохранить C |
| T23 | Legacy capabilities невозможно fence / A,B,C | Historical migration отклонена; сохранить C |
| T24 | P10 запрошен без current owner / A,B,C | Reject terminalization |
| T25 | Terminal append response потерян / A,B,C | Independent persisted readback; без duplicate terminal event |
| T26 | Только clock expiry/TTL или completed run / A,B | Loss/fence proof недостаточен; STOP |
| T27 | Remote task ref recreated на другом OID / A,B,C | Replacement ref не удалять; reconciliation обязателен |
| T28 | Observer создаёт source Release / A,B,C | Reject provenance; ownership transition отсутствует |
| T29 | Gateway journal committed, API outcome неизвестен / B | Intent остаётся uncertain до authoritative reconciliation |
| T30 | Claim cancelled после old executor fencing / A,B | Нет automatic return старого owner/capability |

## 10. Prospective activation и historical migration

По умолчанию activation prospective-only после отдельно approved normative и implementation stages.
До activation нужны authoritative capability qualification, architecture decision, contract alignment,
complete enforcement implementation, independent crash/concurrency proofs и explicit activation scope.
Documentation Acceptance, commit или publication сами не выполняют operational conditions.
Новая authoring authority не передаёт publication permission ни одного fixture.

Historical migration требует individually authorized exact transaction и полного attempt/request inventory.
Original records и dispositions неизменны; новый actor имеет собственную явную attribution.
Доказать complete old capability fencing, reconcile все potentially in-flight effects и successor admission.
Если legacy shared credentials, queued requests или missing history исключают proof, migration blocked.
Retrospective Release, synthetic owner continuity и historical terminal rewrite запрещены.
Offline fixture success не разрешает live TASK-095/096 recovery.

## 11. Candidate evidence record fields

Это schema inventory будущего design, не новый native event format.
Каждый record различает logical transition identity и physical persisted item identity.

| Field group | Required candidate binding |
| --- | --- |
| Transaction | Original gate/Target identity, accepted scope, ordered range и fixed base |
| Authority | Actual scoped recovery authorization, version, revocation и actor attribution |
| Claim | Exact predecessor/version, unique candidate claim и conditional operation outcome |
| Enforcement | Fence epoch, covered paths/capabilities, durable enforcement и drain evidence |
| Effects | Exact intent/request identity, Git/GitHub outcome, observation cut и uncertainty |
| Readback | Independent full persisted payload identity, digest, locator и complete current tail |

Unknown/conflicting fields запрещают admission; timestamps не восстанавливают missing causality.

## 12. Evidence, acceptance и следующие gates

Будущее provider investigation фиксирует official source/version, supported interface, exact guarantees,
serialization domain, enforcement coverage, negative cases и reproducible read-only evidence.
Missing evidence остаётся Unknown; отсутствие records не означает proven absence of actions.
Model tests не сертифицируют provider guarantees или exclusive credential mediation.
External provider capabilities для Draft не загружались, не настраивались и не проверялись.

Design acceptance требует зеркального A/B/C coverage, полной capability/evidence matrix, явных authority axes,
state predecessors, crash windows, adversarial outcomes и legacy limitations.
Independent Architecture Confirmation, Verification/Tester, non-author Review и Scope Audit — task gates.
Их acceptance оценивает bounded document, не утверждает operational contract.
Impact inventory будущего normative stage: PROCESS-001/002, AGENT, Publisher/Coordinator roles,
recovery/Publisher scenarios, task template и зеркальные process guides; изменения отдельно согласуются.
Implementation/tooling, credentials, provider configuration и live migration исключены из current scope.
Next candidate: evidence research либо architecture decision, Not Activated, без automatic task assignment.
STOP при unknown enforcement, incomplete lineage, competing claim, stale predecessor, Target conflict,
lost unqualified outcome или требовании impersonate historical source.

# DP-024: Граница recovery claim и admission Runtime

[English version](../../en/design/DP-024-runtime-recovery-claim-admission-boundary.md)

## 1. Статус

- **Design Status:** Approved
- **Implementation Status:** Partial (isolated local test-owner foundation, closed DP-014/DP-015 participation and client replacement)

TASK-079 ввела private boundary. Независимое Architect Confirmation TASK-080
утверждает уточнённый design. TASK-081 реализует только
`internal/runtimerecoverycoordination`: exact-scope explicit participant identity,
closed aggregate/command test-owner models, detached complete membership/revisions,
startup-closed/sticky fence, local reservation и client/fatal epoch invalidation.
Outer serialization отделён от коротких state/aggregate/command locks;
reservation sessions сохраняют только outer token при simulated external I/O.
Один local revalidation disposition и idempotent disposal не создают claim/permit.

TASK-084 добавляет minimal dependency leaf `internal/runtimecoordination` и
complete closed participation реального DP-014 owner: construction до allocation,
все direct/narrow writers, ordered detached readers, coherent complete history и
allocation/absence stamps. Все coordinated writes reject; missing commands,
current authority и qualified no-claim не могут открыть admission. Legacy
constructor regressions сохраняют successful isolated publication behavior.

TASK-085 добавляет complete closed real DP-015 participation, detached command
membership/link/presence stamps и storage-wide client replacement. Явно supplied
inert `AuthorityFence` и storage-owned `ClientEpoch` дают только local ordering/
invalidation; actual containment authority и fatal coupling отсутствуют.
Authorized replay/read доступен, каждый coordinated mutation reject;
successful fixtures не доказывают coordinated startup.

Implementation Status остаётся Partial для TASK-081 models и closed local slices
TASK-084/TASK-085. Gates/closure каждой задачи определяет только newest valid
matching Recovery Evidence Envelope. Integrated P1–P14 остаются Planned, item7
unproven. Authentic producer provenance, persistent stores, durable claim/readback/
resume и live recovery permits отсутствуют. DP-017 остаётся Approved / Planned
overall: реализован только isolated read-only assessment TASK-075. Durable recovery
barrier и production recovery отсутствуют.

## 2. Назначение и источники

Ответить на четыре gap TASK-078 для DP-023 §19 item 7: joint ordering,
authentic prior-authority loss, durable claim/readback/resume и claim-aware
assessment. Предложение уточняет [DP-017](DP-017-runtime-recovery-reconciliation.md)
§§7–9,16–19,23 в рамках
[ARCH-004](../architecture/ARCH-004-runtime-deployment-and-identity-model.md).
[DP-014](DP-014-runtime-operational-identity-persistence.md) владеет aggregate и
attempt truth; [DP-015](DP-015-runtime-management-command-idempotency.md) —
command/parent/phase truth. [DP-022](DP-022-runtime-execution-containment-and-evidence.md)
и [DP-023](DP-023-runtime-process-containment-bootstrap.md) владеют containment
и generation evidence. Это Approved уточнение не переопределяет перечисленные источники.

Approved-source amendment или ADR не предлагаются: это заменяемая private
реализация существующих требований ordering/ownership без постоянного выбора
технологии. Изменение этих требований требует отдельной архитектурной работы.

## 3. Scope и non-goals

Предлагаемая начальная реализация — single-node, in-process, restart-based.
Она определяет coordination semantics, участие owners, private provenance,
closed admission, claim issuance/resume и read-only observation. DB, schema,
adapter, identifier format и production provisioner не выбираются.
Lifecycle/command repair, claim release, API, discovery, reporting, automatic
restart, production integration и activation не реализуются.

Same-process потеря client, Owner, stack или permit остаётся unresolved. Будущий
revocation-and-drain protocol должен независимо доказать невозможность сохранения
всех прежних capabilities до поддержки same-process recovery. NewBoundary,
cancellation, timeout и пустая live map недостаточны.

## 4. Ownership и exact scope

Exact scope: operational/containment domain, Workspace, Configuration и Runtime
Instance. Composition явно передаёт один и тот же private per-Instance ordering
participant owners DP-014, DP-015 и recovery. Разные Instances независимы.
Participant владеет ordering и local fence; records, validation и conditional
publications остаются у owners. Recovery владеет claim/barrier coordination и
non-transferable permit; composition предоставляет current containment authority
и prior-process-loss proof. Raw maps/locks, generic transaction registry, service
locator, произвольный callback и ordinary lifecycle capability границу не пересекают.

Это complete participation inventory. TASK-084/TASK-085 поставляют closed
DP-014/DP-015 participation и client replacement; recovery и actual containment
seams остаются Planned:

| Owner/source | Required participation |
| --- | --- |
| `runtimeidentity/store.go` | Instance creation/candidate identity allocation при влиянии на absence/identity; attempt claim/history append, execution binding, Running publication, Stop claim, все authority-specific terminal publishers включая recovery |
| `runtimecommandidempotency/store.go` | Primitive inspect/claim, tracked-Start Stop admission, permit consumption/delegation gates, terminal publication |
| `managed_start.go`, `managed_parent.go`, `orchestration_admission.go` | Managed candidate revalidation, tracked parent/preclaimed StopOld atomic admission, binding/rendezvous continuation, Satisfied terminal publication |
| `parent_store.go` | Parent admission, StopOld/StartTarget creation/delegation, phase/parent terminal publication, omission/order decisions |
| `assessment_snapshot.go` и DP-014 detached readers | Complete coherent fact collection; pending/unknown coordination не является qualified no-claim/Clean; reader snapshots не раскрывают participant, lock, map или authority |
| `MemoryStorage.nextGeneration` / `NewBoundary` | Client-epoch invalidation, никогда process-termination proof |
| `runtimerecoveryassessment` | Planned detached claim/provenance reader и ReleaseOnly, без mutation/issuance |
| `runtimecontainmentcomposition` | Явная передача одного participant, current containment authority и immutable supersession observation; без synthetic producer provenance или containment-ledger extension |
| Future recovery owner/store | Claim CAS, exact readback, conditional resume, local single issuance; future release участвует, но не активирован |

Все writers и delegation gates должны участвовать до заявления integrated
ordering. Ordinary terminal completion не обходит participant. Live callback
может работать вне locks; takeover всё равно требует proven termination каждого
authentic relevant process origin.

## 5. Joint ordering protocol

Outer exact-scope serialization token и owner-local locks — разные boundaries.
Каждый seam §4 участвует; recovery claim/resume и future release разделяют
ordering. Захват: outer participant, затем fixed owner-local lock order;
обратный вход запрещён. DP-014 aggregate locks, DP-015 active-client-generation/
ledger locks и Runtime Owner locks защищают только короткие internal transitions.
Все owner-local locks освобождаются до external persistence/readback,
authorization, evidence queries, waits или lifecycle work.

Recovery может удерживать outer token во время qualified bounded CAS/readback;
это никогда не разрешает external I/O под owner-local lock. Разные Instances
независимы. Unknown completion сохраняет closed admission и unresolved
disposition даже после освобождения physical token. Lifecycle/evidence callback
не выполняется под token.

DP-015 §13.1 остаётся authoritative: parent, preclaimed StopOld, sole Stop
exception occupant, rendezvous и private live phase permit образуют один atomic
internal transition под active-generation/ledger locks. Этот design не разделяет
transition и не переносит ordinary parent admission в recovery storage.
Future persistent DP-015 adapter отдельно доказывает compatibility.
Новый ordinary-command publication protocol или adapter здесь не утверждён.

Detached facts/evidence собираются вне owner-local locks. Под participant locks
кратко проверяют exact aggregate/history, complete command membership,
absence/links/revisions, claim/coordination revisions, authentic provenance,
current unfenced containment authority и local client/admission epoch. Записывают
только closed candidate reservation §9, затем освобождают locks.
Stale/conflicting validation имеет zero durable mutation и не выдаёт permit.

Token исключает всех conforming writers пока recovery publication pending;
validated DP-023 authority исключает другого authorized process writer.
Captured owner revisions поэтому не меняются через conforming writer между
validation и recovery CAS. Adapter без этой гарантии unqualified. Нынешние
раздельные owner mutexes/stable rereads не реализуют протокол. Cross-store
distributed transaction и physical layout не выбираются.

## 6. Admission fence до claim

Startup admission по умолчанию закрыт до чтения coordination, aggregate и command
truth. При required recovery assessment по DP-017 §7 (matching live ownership
не доказано либо есть unresolved claim), первое non-clean или unknown observation
вводит путь в participant
и закрывает sticky local fence до возврата observation или попытки claim.
Если command первой прошла область, assessment stale и перечитывает весь set.
Stale claim rejection имеет zero durable mutation, local fence остаётся закрытым.
Observation само по себе не создаёт claim.

Crash до claim не открывает admission: successor startup закрыт. Только fresh
authoritative fully clean no-claim verification под joint ordering разрешает
normal admission без recovery mutation. Missing storage не является qualified
no-claim read. Existing unresolved claim остаётся закрытым до будущего DP-023
item 10 release. Все state-changing/later-phase paths соблюдают fence;
tracked-Start Stop exception не переживает ownership loss. Authorized same-key
observation/terminal replay остаётся read-only без permit. Ordinary observation
normally managed live Starting/Running Instance не является этим recovery trigger
и не отключает normal tracked-Start Stop path.

## 7. Authentic origin и потеря прежней authority

До normal command/phase permit issuance или delegation durable private
coordination provenance связывает exact scope, command/phase identity,
admission revision и parent links с authoritative process execution generation.
DP-014 lifecycle producers также связывают operation kind, candidate attempt
при наличии, expected/resulting aggregate revision и related command/phase со
своей authentic producer generation. Recovery issuer generation и opaque
incarnation сохраняются до recovery permit issuance.

Provenance — coordination metadata, а не второй lifecycle/command truth store,
attempt execution binding или extension containment ledger. Lock separation и
reservation/CAS/readback discipline §§5,9 обязательны; external provenance
publication/inspection не выполняются под owner-local locks. Оно должно commit
вместе с owner publication либо как write-ahead certificate перед ней;
оба варианта требуют exact inspection до capability release. Orphan certificate
сам по себе не является admission или claim truth. Proven publication absence
при unchanged expected revision разрешает retry того же candidate; exact
matching publication связывает provenance с owner facts. Unknown/newer/foreign/
ambiguous state остаётся закрытым без capability. Certificate deletion или
unconditional reinterpretation не разрешены.

Validated current containment authority и exact prior-generation supersession
могут доказать process-wide loss только когда каждый relevant non-terminal
producer и normal/recovery capability origin связан с proven terminated
generation. Это покрывает CommandOnly/UnboundAttempt без предположения attempt
binding. Missing/legacy provenance unresolved; его нельзя изготовить из current
generation, client generation, отсутствия maps, PID или времени. Relevant
current-generation provenance запрещает takeover в restart-only protocol.

Этот proof отделён от exact attempt-bound resource absence и Host shutdown
completion. Producer provenance не предоставляет execution binding или Stopped.
Outside-region evidence — immutable exact supersession observation; внутри
revalidate-ятся provenance binding, local current-authority identity и unfenced
state до commit и issuance. Evidence query callback внутри не выполняется.
Authority loss следует DP-023 fatal fencing/termination, без in-place takeover.

## 8. Durable coordination record и storage obligations

Abstract record сохраняет immutable scope и non-reused opaque claim identity,
claim revision, starting aggregate revision и attempt/history correlation,
полный captured non-terminal primitive/parent/phase identity/revision set с
membership/absence proof, creation generation, current issuer generation и
opaque incarnation, observed producer-provenance set и unresolved/closed barrier
state. Resume сохраняет original claim identity/observations и conditional
append current monotonic observations и fresh issuer; original observations
не переписываются так, будто reconciliation не было.

Заменяемый persistent boundary обязан обеспечить complete atomic record CAS,
monotonic revisions, exact detached readback, inspectable committed/absent/unknown
results, process-restart retention и existing-only original storage-authority
validation. Missing/replaced/rolled-back/corrupt/uncertain storage не является
empty и не даёт authority. Provisioning должен предотвращать undetectable joint
rollback/clone; candidate storage не self-attest expected authority.
Только qualified authoritative absence доказывает отсутствие claim.

DP-014 и DP-015 truth также должны переживать restart. Persist только claim при
исчезновении aggregate/command stores не удовлетворяет item 7. Нынешние Store
и MemoryStorage process-local. Containment ledger содержит только generation/
supersession facts. Raw permits, Host pointers, payloads, Secrets, PID и
wall-clock takeover metadata не входят в coordination records.

## 9. Claim, confirmation и single issuance

После detached collection и validation §5 закрыть/сохранить sticky local
admission fence и записать private candidate reservation с exact validated facts.
Освободить все owner-local locks. Этот local transition — только fence/reservation:
без committed claim, success, reopened admission, usable permit или delegation.
Reservation не является recovery truth.

Удерживая outer token, выполнить qualified recovery-store conditional atomic
publication. Successful durable CAS — claim commit/publication point: один complete
record содержит immutable claim identity, captured facts/provenance, issuer
generation/incarnation, monotonic revision и unresolved closed barrier.
Exact candidate readback проходит вне owner-local locks. Затем кратко reacquire
local locks под participant и revalidate reservation, epoch, exact owner revisions,
claim revision и current unfenced authority. Записать одну local issuance disposition.
Освободить locks и participant до synchronous delivery original candidate-owning
path. Observers не получают permit. Stored identity не восстанавливает issuance;
private permit non-transferable и не может invoke Start, Stop, Flow, Load, Build,
Launcher или adopt Host.

Каждый issued/delivered permit навсегда связан с exact claim revision, issuer
incarnation, local client/admission epoch и current containment authority.
После освобождения issuance locks/token synchronous delivery проходит новый короткий
closed internal acceptance gate, atomically ordered с epoch replacement/fatal
invalidation. Gate выполняет только bounded local validation/acceptance: без external
I/O, lifecycle или arbitrary callback под owner-local locks. Invalidation до acceptance
подавляет successful delivery и оставляет issuance exhausted/unresolved. Если acceptance
первой прошла gate, последующая invalidation немедленно disables handle; удержание
или возврат handle не continuing authority. Delivery linearizes на gate, а не physical
return. Каждый последующий permitted conditional recovery use проходит fresh
exact-epoch/current-unfenced-authority gate под тем же participant, atomically ordered
с invalidation. Stale handles reject без новой authorized mutation/delegation.
External persistence остаётся вне всех owner-local locks по reservation/publication/
confirmation §§5,9. Уже in-flight uncertain publication inspect как possibly committed,
никогда не reported successful от invalidated authority. Gate не retries/recreates
issued permit; epoch invalidation не proof termination old process/callback.

Durable commit и live issuance различны. Atomic semantic boundary DP-017 сохранена:
normal admission/другой issuer не interleave, reservation не даёт authority,
interrupted completion остаётся unresolved/closed. Client-generation replacement
или fatal fencing invalidate pending reservation и disable issuance даже во время
persistence. Invalidation никогда не ждёт external storage под owner-local locks
и не захватывает locks обратно. Она не переписывает captured owner revisions,
не удаляет possibly committed claim и не доказывает prior-process termination.
Inspection может завершиться, но invalidated candidate не возвращает permit успешно.

Unknown publication/readback остаётся closed. Сначала inspect same candidate.
Retry требует proven exact absence, unchanged expected durable/owner facts,
valid original storage authority и fresh current authorization; используется тот
же candidate. Другой tail/claim не перезаписывается. Если issuance могла состояться,
same-generation reissue/adoption запрещены. Cancellation/loss сохраняет unresolved
claim; fenced issuer не может continue/delegate.

| Cut | Required recovery behavior |
| --- | --- |
| До reservation/fence | Нет claim/permit; successor admission starts closed |
| Local fence/reservation, до durable CAS | Нет committed claim/success; successor inspect qualified storage, reservation не reconstruct permit |
| Durable publication uncertain | Closed admission; exact-candidate inspection до retry |
| Claim committed, до local confirmation | Authoritative unresolved claim сохраняется; successor доказывает termination prior origins до conditional resume |
| Confirmed claim, до issuance | Stored identity не доказывает issuance; same-generation uncertainty fails closed |
| Local issuance recorded, до/uncertain delivery | Invalidation до acceptance подавляет success; после acceptance disables handle; нет duplicate в этой generation/incarnation |
| Permit delivered, cancellation/client-epoch replacement/fatal fence | Claim unresolved; invalidated handle не continue/delegate |
| Restart | Old permit не revived; fresh conditional resume требует termination всех relevant origins и сохраняет original observations |

## 10. Conditional resume

Successor сначала доказывает termination prior issuer и всех relevant normal
capability origins, читает full current aggregate/command/claim/provenance set,
revalidate exact revisions под joint ordering. Сохраняет claim identity/original
observations, conditional advances claim revision с fresh issuer generation/
incarnation и current observations, подтверждает exact readback, затем использует
reservation, unlocked CAS/readback и single-issuance protocol §§5,9. External I/O
не проходит под owner-local lock. Старый permit не восстанавливается.

Current-generation issuer, incomplete origin set, stale facts, missing original
authority или uncertainty возможности сохранения prior issuer/capability
запрещают resume. Proven prior-generation termination устраняет неопределённость
historical possible issuance; current-generation uncertain issuance никогда не разрешает reissue.
Inspect-first convergence
после каждого unknown cut предшествует retry. Только later items 8–10 выполняют
monotonic reconciliation/release; claim/resume не может clear attempt, terminalize
command/parent, manufacture Owner completion basis или открыть claimed barrier.

## 11. Claim-aware read-only assessment

Private assessment расширяется detached exact claim/provenance observation и
revision, явно переданными recovery reader. Stable rereads включают claim/
coordination revision вокруг evidence. Missing/foreign/stale/unknown reads дают
Unknown, а не absent claim. Pending reservations/unknown disposition никогда
не qualified no-claim или Clean. Assessment не выдаёт mutation capability.

Clean требует no active attempt, no non-terminal primitive/parent/existing phase,
coherent terminal aggregate/command facts и qualified absence
unresolved claim. Те же terminal facts с exact unresolved claim классифицируются
ReleaseOnly, никогда Clean; одна classification не выполняет release/reopening.
Nonterminal classifications сохраняют exact claim observation и closed admission.
Нынешний TASK-075 не имеет claim reader или ReleaseOnly result; эти additions
Planned, а не existing capability.

## 12. Risk-oriented proof obligations

| ID | Будущий executable proof | Источник |
| --- | --- | --- |
| P1 | Все owner writers используют один exact participant; нет admission/phase bypass | DP-017 §§8,19 |
| P2 | Command-vs-observation/claim race; stale full-set/absence revisions дают zero durable mutation | DP-014 §16; DP-017 §§8–9 |
| P3 | Первое non-clean/unknown observation и restart сохраняют closed fence | DP-017 §§7–9,20 |
| P4 | Concurrent claim paths создают один committed claim и не более одного live permit | DP-017 §9 |
| P5 | CommandOnly/unbound origins доказывают old process loss без synthetic binding | DP-017 §§10–11 |
| P6 | Unlocked live callback, client expiry, same-generation issuer или missing provenance запрещают takeover | DP-015 permits; DP-017 §9 |
| P7 | Crash на provenance/publication/readback/issuance и issuance→unlock→delivery invalidation cuts; exact inspection, без duplicate capability | DP-014 §17; DP-017 §§9,20 |
| P8 | Persistent aggregate, complete commands и claim переживают real process restart | DP-017 §§9,17,23 |
| P9 | Missing/replaced/rollback/clone/corrupt authority никогда не становится empty | DP-023 §8.1; DP-017 §23 |
| P10 | Fresh conditional successor resume сохраняет original observations; stale/foreign input не меняет facts | DP-017 §§9,16 |
| P11 | Clean vs ReleaseOnly; stable detached claim-aware reads не дают permit | DP-017 §§7,12 |
| P12 | Cancellation/unknown result/fatal fence сохраняют closed admission; epoch replacement отзывает issued и delivered permits | DP-017 §§20–21; DP-023 §12 |
| P13 | Independent Instances, fixed lock order, evidence/lifecycle вне области, race/stress | DP-017 §19 |
| P14 | Claim не даёт lifecycle/reconciliation/release/Owner-basis promotion или new attempt | DP-017 §§13–17 |

Все четырнадцать rows — Planned obligations, не PASS. In-memory simulations
могут доказать только local ordering/issuance mechanics. Они не доказывают P8/P9
durable conformance, process-origin production coverage или production recovery
readiness.


### Coverage первого foundation slice

| Proof | TASK-081 isolated local model coverage | Remaining prerequisite |
| --- | --- | --- |
| P1 | Explicit same participant и exact owner identity только test models | Каждый real writer/delegation gate |
| P2 | Detached complete membership/revisions и stale/foreign rejection | Real complete owners и durable conditional publication |
| P3 | Startup closed; fresh clean-only opening и sticky unknown/recovery fence | Real restart с qualified owner/claim stores |
| P4 | Reservation/revalidation — только local data/disposition, без authority | Durable one-claim/single live issuance |
| P5 | Deferred; origin input/loss proof отсутствуют | Authentic durable producer provenance |
| P6 | Client epoch invalidation не даёт takeover authority | Complete origins/surviving real callbacks |
| P7 | Pending local candidate invalidation при simulated external I/O | Real provenance/CAS/readback/issuance crash cuts |
| P8 | Deferred | Persistent aggregate/complete commands/claim restart |
| P9 | Deferred | Existing-only authority, anti-clone/rollback provisioning |
| P10 | Resume operation/authority отсутствуют | Conditional durable successor resume |
| P11 | Unknown/incomplete/claim-present model facts не открывают admission | Actual claim-aware assessment/ReleaseOnly |
| P12 | Local epoch/fatal invalidation закрывает admission и rejects old candidate | Actual fatal authority/uncertain storage и delivered-permit revocation |
| P13 | Independent Instances, fixed owner lock order, owner locks released при simulated I/O | Real integration, race/stress/persistent I/O |
| P14 | Нет lifecycle/reconciliation/release capability | Continued integrated negative proofs |

Это partial local model coverage, никогда full P1–P14 PASS, real-owner participation
или durability qualification. Восемь cuts §9 по-прежнему требуют integrated
executable evidence. Reservations и successful local revalidation не дают authority.

## 13. Prerequisites и следующее решение

TASK-081 поставляет только isolated closed test-owner models. TASK-082 завершает
real-owner/generation readiness: full integration SPLIT REQUIRED / NOT READY
AS ONE BOUNDED TASK. TASK-083 фиксирует concrete private realization в §15.
TASK-084 реализует minimal leaf и complete closed DP-014 local owner coverage;
TASK-085 добавляет complete closed DP-015 participation, membership stamps и
storage-wide client replacement из §15. Следующая рекомендация — fresh bounded
actual containment-source fatal coupling и explicit composition supply readiness,
Not Activated, со fresh Size Guard и independent review. Missing current authority/
qualified no-claim оставляет coordinated writes closed; inert root construction
не квалифицирует containment authority.

Durable provenance, persistent aggregate/complete command/recovery storage и
original storage-authority/provisioning qualification — отдельные prerequisites;
DB, schema, adapter и provisioner не выбраны. Только later fresh intake с integrated
owners, authentic durable origins и qualified stores может заявить item7 completion.
Items 8–10 reconciliation/release и item11 reporting/integration остаются позже.
Design task не активирует implementation или successor.

## 14. Граница решения

Независимый Architect TASK-080 явно утверждает уточнённый private restart-based
design: Approved; Implementation остаётся Partial для local models TASK-081 и closed DP-014/DP-015 participation plus client replacement TASK-084/TASK-085. Internal atomic transitions и external-I/O prohibition
DP-015 не изменены. TASK-079/TASK-080 не поставили код, durable store, recovery permit
или production recovery. Все integrated proof rows остаются Planned. Unknown provenance,
storage, authority или publication outcome оставляет admission closed;
design Approval не активирует implementation.

## 15. Конкретный план local owner participation

TASK-083 фиксирует эту private realization существующей Approved семантики.
TASK-084 реализует minimal typed leaf, closed real DP-014 participation и
coherent detached aggregate observations. TASK-085 добавляет complete closed real
DP-015 participation, membership observations и local client replacement.
Actual fatal-source coupling, reservation-session и integrated recovery protocol
остаются Planned. Design Status остаётся Approved; Implementation Status Partial.
Эти slices не дают claim, permit, takeover authority или production startup
qualification.

### 15.1 Typed binding и явное coverage

Dependency leaf `internal/runtimecoordination` не импортирует aggregate, command,
containment, lifecycle или recovery owner. TASK-084 реализует immutable validated
`Scope`, copy-safe `Participant`, non-rebinding `AggregateAttachment` и short
`ObservationRegion`; transition region всегда reject. TASK-085 реализует
copy-safe `CommandAttachment`, `ClientEpoch`, inert `AuthorityFence` и bounded
registration/replacement regions. Reservation-session handles остаются Planned.
Composition явно поставляет identical participant. Attachment однократно связывает
exact owner identity и scope; duplicate, foreign и rebinding reject. Complete-owner
markers принадлежат conforming owner construction, а не caller claims. Через seam
не проходят raw locks, generic callbacks, service discovery или ordinary lifecycle
capability. Leaf latch не query-ит и не заменяет containment authority.

Реализованный `runtimeidentity.NewCoordinatedStore(scope, participant)` связывает одну
Instance до allocation: historical `AllocateCandidateIdentity(id)` не передаёт
domain/Workspace/Configuration. Каждый direct method, включая narrow terminal
publishers, rejects foreign identity и сам входит в conforming region. Historical
`NewStore`, `NewMemoryStorage` и uncoordinated composition остаются supported, но
не считаются DP-024 conformance. TASK-085
`NewCoordinatedMemoryStorage(authority)` использует typed `AttachInstance(scope, participant)`
до capture ledger absence. Staged registration commit происходит только после short
owner catalog checks; rejected unattached calls не раскрывают absence и не poison attachment. Attachment не
зависит от существования aggregate или ledger.

Complete DP-014 coverage включает allocation/create, attempt claim/history append,
execution binding, Running, Stop и все authority-specific terminal publishers,
aggregate/history readers и exact-attempt capture/revalidation. TASK-085 complete closed DP-015
coverage включает primitive inspect/claim и tracked Stop; managed admission,
binding и live gates; parent/phase admission, creation, delegation, terminal и
expiry; atomic tracked parent/preclaimed StopOld/slot/rendezvous/private permit
admission и consumption; replay-first orchestration; pending Stop и managed
rendezvous transitions; complete assessment snapshots и client replacement. Actual
containment source fatal paths должны участвовать в later separate slice. Wrapper не доказывает coverage direct calls.

### 15.2 Полные detached observations и revisions

Detached stamps содержат exact scope, opaque attachment identity, aggregate presence/
allocation revision, aggregate/history revision, command-ledger presence, command
membership revision, client epoch и participant epoch. Capture/revalidation не
меняют domain facts и не возвращают handle/authority. Allocation/create повышают
presence revision; successful aggregate publications повышают существующую
aggregate revision. Ledger creation, primitive/parent/phase membership/publication,
live-permit consumption/expiry, Stop-slot и rendezvous changes повышают отдельную
coordination membership revision. Existing domain record revisions сохраняют semantics.

Capture включает все terminal/non-terminal members и links; empty set содержит
явный absence stamp. Missing, stale, foreign или partial coverage не открывает
admission. Create-after-absence обязан invalidate stamp. Counters никогда не wrap:
exhaustion закрывает admission и rejects future authorized mutation.

TASK-084 `CaptureCoordinationSnapshot` и `RevalidateCoordinationSnapshot`
реализуют только DP-014 часть: exact scope, отдельную opaque identity без
owner/lock links, presence revision, allocation/aggregate presence, aggregate view
и все terminal/nonterminal history members. Capture проверяет complete index,
immutable binding, active/terminal и desired/actual coherence; nil, incomplete или
contradictory owner facts reject без partial result. Coherent unbound early facts
и retained StopFailure остаются valid local observations; отдельный exact execution
reader по-прежнему rejects unbound selected attempt. Command, participant/client
epoch или qualified no-claim stamp не фабрикуются. Presence/domain counters rejects
exhaustion до mutation. Create-after-absence и successful-publication observations
используют явно обозначенные package-local fixtures; они не доказывают successful
coordinated mutation.

TASK-085 command capture/revalidation копирует все terminal/nonterminal members,
live/slot/rendezvous links и explicit ledger absence под одним ordered region.
Opaque snapshot identity не содержит owner/lock links; mutable permit/channel/
bridge references удалены. Exact scope, intent/authorization, derived phase order
и retained resolved/blocked diagnostic bridges проверяются без трактовки retained
tokens как live authority. Private value/identity accounting повышает membership
только при actual changes; reads не fence и не меняют facts. New coordination
counters reject exhaustion до mutation; existing domain revision semantics
сохранены, с covered pre-mutation exhaustion guards. Это не global legacy counter
или authority guarantee. Successful changes используют labelled local fixtures
или отдельные legacy regressions, а не coordinated mutation evidence.

### 15.3 Фиксированный порядок коротких regions

Complete planned protocol acquire `Instance outer -> AuthorityFence read gate ->
ClientEpoch read gate -> participant state -> owner-local locks`. Owner catalog
lookup/registration предшествует aggregate lock. DP-015 сохраняет existing
`clientMu.RLock -> ledger.mu` atomic transition, включая parent, preclaimed
StopOld, sole occupant, rendezvous и private live phase permit.

TASK-085 прозрачно переводит previously attached aggregate observations на root
order после command registration commit. Rootless short observation сохраняет
participant state до Close, поэтому registration не пересекает owner read; это
не external-I/O reservation. Rooted command regions сохраняют fixed root/state
gates до `clientMu.RLock` и ledger locks. Public wrappers делегируют private
short-region helpers, освобождая каждый region до nested entry/callbacks/waits.
Missing current authority/qualified no-claim оставляет transitions closed.
Shared root/attachment/region copies используют один disposition; zero/nil Close
безопасен. Raw mutex или arbitrary callback не предоставляется.

Aggregate/command facts можно копировать последовательно, пока identical outer
token исключает writers; raw owner locks не пересекают package boundaries. Все
authority/client/participant-state/owner gates освобождаются до authorization,
evidence/storage/provider calls, lifecycle work или waits. Reservation может
сохранять только outer при qualified external work. Direct owner calls и narrow
publishers сами получают conforming region; composition wrapper недостаточен.
Reverse entry из owner/client lock к outer token запрещён.

### 15.4 Storage-wide client replacement

`ClientEpoch` — отдельный bounded local gate, не замена DP-015 `clientMu`.
Replacement получает exclusive gate, closes/invalidates все typed attached
participants под каждым participant state lock, затем публикует новую generation
под `clientMu.Lock`. Он не получает outer или owner ledger/aggregate lock и не
ждёт external I/O. Retained outer/I/O другой Instance не блокирует этот bounded fan-out.

Registration следует `AuthorityFence.R -> ClientEpoch.W -> participant.state`,
после чего выполняется short owner registration. Она не может escape concurrent
replacement или inherit prior opening. Overflow permanently fences client root и
attached participants и запрещает visibility новой generation. Fan-out ждёт только
bounded local gates; это не process-loss proof и не разрешение recovery.

TASK-085 реализует этот local registration/replacement order, copy-safe disposal
и permanent exhaustion fencing до new client generation visibility. Inert root
не доказывает prior-process loss, current authority или fatal coupling.
Fixture-only outer retention показывает, что fan-out не получает outer token;
short observation regions освобождаются до external work.

### 15.5 Fatal coupling во всех actual source paths

Containment keeper владеет явно associated `AuthorityFence`. Каждый source
fatal-publication path связывает existing `keeper.fenced` publication и complete
attached-participant invalidation под authority root exclusive gate до возврата
fatal result. Registration/internal acceptance используют тот же root read gate:
fatal-first подавляет acceptance; acceptance-first немедленно invalidated после
этого. Terminal completion, expiry и later phases obey resulting fence; authorized
replay остаётся read-only.

External `IsAuthoritative`, provisioning validation и `ReadGeneration` storage
work остаются вне всех short gates. Roots предоставляют только bounded local
invalidation, без evidence callback или termination claim. Leaf не заменяет
containment truth и не меняет DP-023 fatal shutdown obligations.

### 15.6 Наименьшая независимая последовательность и executable obligations

TASK-084 реализует first slice: leaf binding/local observation regions и complete
closed DP-014 direct-writer/narrow-publisher/reader/absence coverage. Его 13 proof
tests и два regression tests покрывают copy/zero handles, attachment до
allocation/absence, direct/narrow rejection, detached/stale/foreign/partial facts,
counter exhaustion, independent Instances и released owner locks. Success fixtures
package-local; historical constructor regressions сохраняют existing successful
publication behavior. Focused stress, full regression и vet проходят; race
verification недоступна в текущем CGO-disabled/no-C-compiler environment. Evidence
доказывает только local slice, не integrated P1–P14.

TASK-085 реализует следующий closed slice с 28 added tests: 25 proof и три
regression tests. Они покрывают complete direct/callback rejection, authorized
primitive/managed/replay-first terminal replay с zero work callbacks, detached
membership/absence invalidation, exact contradictory-link rejection, retained
diagnostic bridges, copy/zero/abort handles, attachment/replacement races и
overflow без new generation visibility. Successful mutation fixtures остаются
package-local; actual legacy Stop-first/atomicity/counter regressions отделены.
Full supported regression/vet и serial focused stress проходят. Race verification
недоступна при CGO disabled; earlier concurrent legacy bounded-yield failure и
последующие serial passes сохранены как limitation, не diagnosed fact.
Эти proofs подтверждают только closed local mechanics, не integrated P1–P14/item7.

Next candidate: fresh bounded actual containment-source fatal coupling и explicit
composition supply readiness, Not Activated. Каждый successor требует fresh
readiness, Size Guard и independent review; successor здесь не активируется.

Missing command attachment, current authority или qualified no-claim verification
оставляет каждый coordinated write closed. Production test-opening API отсутствует.
Synthetic opening в existing closed model tests остаётся model-only; successful
real-owner startup требует later complete qualification.

Remaining integrated executable proofs: invalidate stamps после каждого
command/live/rendezvous change; сохранение tracked-parent atomicity; two Instances
и racing attachment при replacement; overflow без new-generation visibility;
fatal до/после acceptance; closed-fence authorized replay с callback count zero;
external work вне всех short gates. Это только local participation mechanics,
не integrated P1–P14 или item7 completion.

Durable producer provenance, persistent aggregate/complete commands/claim,
independently trusted original storage authority, anti-clone/rollback qualification,
claim-aware assessment, conditional claim/resume и live recovery issuance остаются
отдельными full item7 prerequisites. Reconciliation, release, reporting, production
wiring и activation находятся вне этой local sequence.

## 16. Готовность source-bound fatal coupling и closed composition

TASK-086 фиксирует planned private realization разделов 4–9 и 15.5. Он заменяет
historical TASK-085 next-readiness recommendation в разделах 13/15 без изменения
этих разделов, Design Status или Implementation Status. Design не предоставляет
runtime capability, current-authority qualification, qualified no-claim, recovery
permit или item7 proof.

### 16.1 Actual source inventory и gap

На baseline TASK-086 существуют семь production fatal publications:

| Actual source | Fatal condition |
| --- | --- |
| `containment.go`, `IsAuthoritative` | Invalid held capability, nil DB или failed provisioning revalidation |
| `containment.go`, `AcquireProcessContainment`, first acquired guard | Acquired capability invalid |
| `containment.go`, `AcquireProcessContainment`, existing-store guard | Existing provisioned store open fails или held capability invalid |
| `containment.go`, `AcquireProcessContainment`, successor guard | Generation creation, successor bootstrap, post-append provisioning revalidation или held-capability validation fails |
| `evidence.go`, `ReadGeneration` | Ledger tail противоречит current generation |
| `evidence.go`, `validateForEvidence` | Held capability invalid или DB unavailable |
| `evidence.go`, `fenceEvidenceError` | Provisioning/ledger/storage inspection fails |

Direct test stores — fixtures, не source coverage. Current authority root содержит
только RWMutex; participant registry принадлежит client. Participant roots появляются
только после command registration commit, поэтому aggregate-only participants
остаются rootless. Existing `runtimecontainmentcomposition.New` принимает prebuilt
dependencies и не supplies shared participant или coordinated stores. Existing
authority/evidence checks не реализуют source-local invalidation.

### 16.2 Source association до owner attachment

Keeper privately создаёт и сохраняет один immutable domain-bound authority root
до first acquired-capability fatal branch. Caller inert root не принимается как
source-associated. Copies сохраняют identity; source/domain rebinding запрещён.
Private composition получает exact keeper-associated root через exact active
authority и existing external validation. Сам root никогда не attests containment
authority.

Additive source-bound participant construction регистрирует exact-domain
participant под `Authority.R -> participant.state` до любого owner attachment или
absence capture. Authority registry включает каждый participant, в том числе
aggregate-only и failed partial construction, и каждый associated ClientEpoch.
Detach/rebind или removal-as-absence запрещены. Root/client construction,
participant registration и staged command registration rejects sticky fatal под
тем же authority read gate. Abort command registration не добавляет command fact
и не удаляет already source-registered participant. Historical rootless/inert
construction остаётся supported, но unqualified.

### 16.3 Typed source-local publication и complete invalidation

Один private keeper helper покрывает каждый actual fatal branch. Он получает opaque
copy-safe fatal-publication region, сохраняющий `Authority.W`, публикует existing
`keeper.fenced=true`, вызывает bounded typed `Invalidate()` и только затем closes
region и returns. Raw lock или arbitrary callback не предоставляется. Handle не
предоставляет source association, evidence, opening, generation, release или
recovery authority. Nil/zero handles safe; copies используют один idempotent
invalidation/close disposition.

Каждый source-associated fast return, увидевший `keeper.fenced=true`, до return
join-ит completed publication через bounded typed same-root completion read.
Read получает Authority.R вне owner/outer gates и после external inspection;
in-flight Authority.W сначала завершает flag publication, complete invalidation
и release. Repeated fatal observations safe; atomic flag сам по себе не доказывает
completed coupling. Это включает IsAuthoritative final false из-за concurrent
fencing. Nil/no-source caller errors не фабрикуют root или authority.

Fan-out следует `Authority.W -> ClientEpoch.W` (один client за раз), затем
`participant.state`, и также покрывает separately registered partial participants.
Root/clients/participants permanently terminal, epochs MaxUint64. Deduplication
или idempotent terminal assignments исключают пропуск participant или counter wrap
при repeated registry associations. Fan-out не получает Instance outer,
aggregate/ledger/clientMu lock и не вызывает callback, storage/evidence/lifecycle
work или external wait. Ожидаются только bounded local gates. Existing replacement
publication и fatal fan-out serializes через каждый client gate; replacement не
может clear terminal state или resurrect participant.

Authority write упорядочен с registration/observation/internal-acceptance read
gates: fatal-first suppresses acceptance; acceptance-first invalidated до fatal
return. Actual live recovery issuance отсутствует; region ordering или labelled
fixtures не доказывают delivered-permit revocation. Authorized replay остаётся
read-only с zero work callbacks; все coordinated writes closed.

`IsAuthoritative`, provisioning validation и `ReadGeneration` fault detection
остаются вне каждого short gate. Source results/reasons, caller-error nonfencing,
fatal-over-cancellation precedence, retained capability, no reacquisition и
DP-023 process-termination obligations не меняются. Local invalidation никогда не
является process-loss evidence или termination proof.

### 16.4 Explicit opaque closed assembly

Additive private assembly получает exact keeper-associated root вне short gates
и проверяет canonical operational/containment domain и exact target scope. Он
регистрирует один source-bound participant до создания своего coordinated DP-014
Store и coordinated DP-015 MemoryStorage/attachment, передавая identical participant
и source root обоим owners. Возвращается только complete closed assembly.
Partial owners/absence не exposed; participant/root, caller completeness assertion,
storage adoption или unqualified owner injection не пересекают его surface.

Этот separate opaque local assembly владеет coordinated dependencies и предоставляет
только detached read observations. Activation methods или Flow/Owner/private-invoker
adoption не добавляются. Legacy activation/evidence `Composition` и `New` остаются
supported, но unqualified. Assembly не предоставляет ordinary runtime owner creation,
successful startup, production wiring или activation.

### 16.5 Next implementation admission и proof obligations

Assessment Ready. После independent acceptance этого plan smallest implementation
candidate Ready to fresh intake / Not Activated: source-bound leaf registry/fatal
handle, complete actual containment fatal coverage и explicit closed composition
supply. Three existing packages, zero new packages, estimate 300–500 production
lines и 13–17 total paths с documentation требуют fresh Size Guard; более 500
production lines или 15 paths triggers reassessment до дальнейшей работы. Source
publication нельзя отделить от complete source association/fan-out/supply без bypass.

| Future obligation | Required executable proof и limitation |
| --- | --- |
| Все actual sources | Static inventory плюс fault injection во всех семи sites, включая bootstrap до active authority exposure |
| Registration ordering | Fatal-first rejection, registration-first invalidation, aggregate-only и aborted command attachment coverage |
| Completed fatal return | Pause после keeper flag publication до fan-out: repeated-fatal/fast-return readers ждут complete invalidation до false/Unknown return |
| Complete fan-out | Two clients/two Instances, partial participants, exact domain/root mismatch, copy/zero handles |
| Sticky terminal state | Fatal versus replacement не resurrect state и не wrap epochs; exhaustion не exposes new client generation |
| Bounded gates | Retained outer не блокирует fatal; fan-out не получает owner locks; external authority/evidence work вне short gates |
| Explicit supply | Identical source/participant/root для обоих real owners, foreign scope rejection, без partial exposure или absence leak |
| Closed compatibility | Direct writers reject, authorized replay zero callbacks; separate legacy constructor regressions |

Это planned proofs, не PASS. Integrated P1–P14 остаются Planned, item7 unproven;
full integration Split Required / Not Ready. Authentic durable producer provenance,
persistent aggregate/complete commands/claim, independently trusted original storage
authority, anti-clone/rollback qualification, qualified current authority/no-claim,
claim-aware assessment, conditional claim/readback/resume и live issuance остаются
отдельными prerequisites. Reservation, reconciliation, release, reporting,
production wiring и activation остаются later work. Design или его acceptance
не активируют successor.

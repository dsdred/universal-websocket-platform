# DP-024: Граница recovery claim и admission Runtime

[English version](../../en/design/DP-024-runtime-recovery-claim-admission-boundary.md)

## 1. Статус

- **Design Status:** Approved
- **Implementation Status:** Planned

TASK-079 ввела предлагаемую private boundary. Независимое Architect Confirmation
TASK-080 утверждает уточнённый design этого документа; Implementation остаётся
Planned. Approval не поставляет код или storage qualification.
DP-017 остаётся Approved / Planned overall: изолированно реализован только
read-only assessment TASK-075. Durable claim, recovery permit и barrier отсутствуют.

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

Это present owners и prospective participation seams, а не поставленные APIs:

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

| Proof | Planned first foundation coverage | Remaining prerequisite |
| --- | --- | --- |
| P1 | Shared identity/no bypass только test participants | Каждый real writer/delegation gate |
| P2 | Local race/reservation/revision rejection | Real complete owners и durable conditional publication |
| P3 | Local startup/sticky fence | Real restart с qualified owner/claim stores |
| P4 | Reservations не дают authority | Durable one-claim/single live issuance |
| P5 | Reject missing/current synthetic origins | Authentic durable producer provenance |
| P6 | Client expiry не takeover proof | Complete origins/surviving real callbacks |
| P7 | Только local cut/invalidation model | Real provenance/CAS/readback/issuance crash cuts |
| P8 | Deferred | Persistent aggregate/complete commands/claim restart |
| P9 | Deferred | Existing-only authority, anti-clone/rollback provisioning |
| P10 | Resume authority не раскрывается | Conditional durable successor resume |
| P11 | Pending/unknown не qualified Clean в model | Actual claim-aware assessment/ReleaseOnly |
| P12 | Local cancellation/epoch/fence negative gates | Actual fatal authority/uncertain storage cuts |
| P13 | Independent scopes/fixed order/no owner locks during simulated I/O | Real integration, race/stress/persistent I/O |
| P14 | Нет lifecycle/reconciliation/release capability | Continued integrated negative proofs |

Это planned partial model coverage, никогда PASS или real-owner/durability
qualification. Восемь cuts §9 требуют future executable evidence.

## 13. Prerequisites и следующее решение

Следующая рекомендация — отдельная isolated private ordering/fence foundation,
Not Activated. Fresh independent intake ограничит scope и Size Guard. Допустимы
явно переданная exact-scope participant identity, startup-closed/sticky fence,
closed typed reservation/epoch invalidation, fixed lock order и test owner
participants. Negative dispositions проверяются без claim-store mutation,
provenance publication, live recovery permit, lifecycle delegation или production
composition. Package/API/file-count до intake не обещаются.

Test owner ports доказывают только local semantics, а не participation каждого
real DP-014/DP-015 writer. Real-owner/generation-transition integration — отдельный
bounded slice. Persistent aggregate/complete command/recovery adapter и original
storage-authority/provisioning qualification остаются отдельными prerequisites;
DB, schema, adapter и provisioner здесь не выбираются.
Только later fresh intake с integrated owners, authentic durable provenance и
qualified stores может заявить item 7 completion. Items 8–10 reconciliation/release
и item 11 reporting/integration остаются позже. Следующая task здесь не начинается.

## 14. Граница решения

Независимый Architect TASK-080 явно утверждает уточнённый private restart-based
design: Approved / Planned. Internal atomic transitions и external-I/O prohibition
DP-015 не изменены. TASK-079/TASK-080 не поставили код, durable store, recovery permit
или production recovery. Все proof rows остаются Planned. Unknown provenance,
storage, authority или publication outcome оставляет admission closed;
design Approval не активирует implementation.

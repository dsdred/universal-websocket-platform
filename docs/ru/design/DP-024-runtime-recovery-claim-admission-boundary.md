# DP-024: Граница recovery claim и admission Runtime

[English version](../../en/design/DP-024-runtime-recovery-claim-admission-boundary.md)

## 1. Статус

- **Design Status:** Draft
- **Implementation Status:** Planned

TASK-079 фиксирует предлагаемый private implementation boundary. Architecture
Confirmation разрешает документировать предложение, но не Approval или код.
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
и generation evidence. Этот Draft не переопределяет перечисленные источники.

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
Instance. Один composition-private per-Instance ordering participant явно
передаётся DP-014, DP-015 и recovery owners. Composition должна передать один
и тот же participant всем writers этого scope, а не независимые gates одного
Instance. Разные Instances имеют независимые participants.

Participant владеет только ordering и local admission fence. DP-014 и DP-015
сохраняют records, validation и conditional publications. Recovery владеет
claim/barrier coordination и private non-transferable permit. Composition
предоставляет current containment authority и prior-process-loss proof.
Raw maps/locks, generic transaction registry, service locator, произвольный
callback и ordinary lifecycle capability не пересекают эту границу.

## 5. Joint ordering protocol

Каждый DP-014 conditional writer и каждый DP-015 primitive/managed claim,
candidate-to-claim path, tracked-Start Stop exception, parent с preclaimed
StopOld, later-phase creation/delegation gate и ordinary terminal publication
входят в одну короткую область. Recovery claim/resume и будущий release также
участвуют. Один обход делает joint-ordering proof недействительным.

Порядок захвата: outer exact-scope participant, затем owner-local locks;
обратный вход запрещён. Внутри выполняются только закрытые owner-defined bounded
validation/publication operations. Authorization, evidence queries и
Owner/Flow/Load/Host work выполняются снаружи. Долгая lifecycle/evidence operation
не удерживает область. Persistent publication использует только qualified bounded
owner operation; unknown storage completion закрывает/fences, а не позволяет
другому writer переинтерпретировать частичный результат.

Перед commit проверяются exact aggregate/history identity и revision,
полные primitive/parent/phase membership и revisions (включая absence и links),
claim revision/absence, coordination revision и current authority. Detached
reads/evidence собираются снаружи и revalidate-ятся внутри. Stale/conflicting/
unknown validation не меняет durable facts и не выдаёт permit; требуется
reassessment. Local fence при этом может оставаться закрытым.

Нынешние DP-014 aggregate mutex и DP-015 client/ledger locks раздельны;
stable rereads сами по себе не реализуют протокол. Cross-process writer
exclusion требует validated current DP-023 authority и conforming storage
boundary. Cross-store distributed transaction не предполагается: owner
publications остаются отдельными, все valid writers разделяют ordering,
restart admission начинается закрытым. Physical layout не выбран.

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
attempt execution binding или extension containment ledger. Оно должно commit
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

Под joint region fresh exact non-clean facts и complete prior-authority-loss
proof разрешают один conditional persistent claim и closed recovery barrier
в одной logical publication point. Claim фиксирует exact captured revisions,
scope, provenance и candidate issuer. Stale/conflicting input имеет zero durable
mutation. Fence и claim исключают всех valid normal writers.

Durable commit сам по себе не возвращает permit. Exact candidate readback
должен подтвердить весь record. Затем под тем же participant с revalidated
current authority/revisions фиксируется одна local issuance disposition для exact
claim revision/issuer incarnation. Только synchronous path, владеющий candidate,
получает один private non-transferable permit; observers его не получают.
Issuance не восстанавливается из stored identity. Recovery authority не может
вызвать Start, Stop, Flow, Load, Build, Launcher или adopt Host.

Unknown commit/readback оставляет admission закрытым без issuance. Сначала
inspect того же candidate; confirmed absence при unchanged expected state
разрешает только exact-candidate retry. Другой tail/claim или unknown authority
не перезаписывается. Если local issuance могла произойти без proven disposition,
duplicate запрещён; same-generation reissue/adoption первоначально не поддержан.
Cancellation/loss оставляет claim unresolved. Claim, committed до последующего
fatal fencing, остаётся для successor inspection; fenced authority не может
использовать permit или delegate work.

## 10. Conditional resume

Successor сначала доказывает termination prior issuer и всех relevant normal
capability origins, читает full current aggregate/command/claim/provenance set,
revalidate exact revisions под joint ordering. Сохраняет claim identity/original
observations, conditional advances claim revision с fresh issuer generation/
incarnation и current observations, подтверждает exact readback, затем использует
single-issuance protocol. Старый permit не восстанавливается.

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
Unknown, а не absent claim. Assessment не выдаёт mutation capability.

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
| P7 | Crash на provenance/publication/readback/issuance cuts; exact inspection, без duplicate capability | DP-014 §17; DP-017 §§9,20 |
| P8 | Persistent aggregate, complete commands и claim переживают real process restart | DP-017 §§9,17,23 |
| P9 | Missing/replaced/rollback/clone/corrupt authority никогда не становится empty | DP-023 §8.1; DP-017 §23 |
| P10 | Fresh conditional successor resume сохраняет original observations; stale/foreign input не меняет facts | DP-017 §§9,16 |
| P11 | Clean vs ReleaseOnly; stable detached claim-aware reads не дают permit | DP-017 §§7,12 |
| P12 | Cancellation/unknown result/fatal fence сохраняют closed admission и запрещают issuance | DP-017 §§20–21; DP-023 §12 |
| P13 | Independent Instances, fixed lock order, evidence/lifecycle вне области, race/stress | DP-017 §19 |
| P14 | Claim не даёт lifecycle/reconciliation/release/Owner-basis promotion или new attempt | DP-017 §§13–17 |

Все четырнадцать rows — Planned obligations, не PASS. In-memory simulations
могут доказать только local ordering/issuance mechanics. Они не доказывают P8/P9
durable conformance, process-origin production coverage или production recovery
readiness.

## 13. Prerequisites и следующее решение

Первый candidate: отдельное design-only DP-024 Design Status и conformance
decision против Approved sources и всех proof obligations. Он Not Activated;
Acceptance этого Draft не разрешает implementation. Решение должно ограничить
first isolated mechanics slice, exact owner seams/proof coverage, затем отделить
его от persistent DP-014/DP-015/recovery storage qualification. Concrete durable
adapter/schema/provisioning остаётся отдельным prerequisite.

Только later fresh intake с approved mechanics, exact owner integration,
authentic provenance и qualified durable stores может заявить item 7 completion.
Items 8–10 reconciliation/release и item 11 reporting/integration остаются позже.
Один ordering/claim behavior сохраняется целостным; evidence scanning, второй
adapter, production wiring и unrelated functionality разделяются. Production-line
и file-count Size Guard требует fresh оценки при implementation intake.

## 14. Граница решения

Зафиксировать предлагаемый private restart-based protocol как Draft / Planned.
Existing Approved statuses и delivered capabilities не изменены. TASK-079 не
поставила код, durable store или production recovery. Unknown provenance,
storage, authority или commit outcome всегда оставляет admission закрытым.

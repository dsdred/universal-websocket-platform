# DP-014: Персистентность operational identity Runtime

[English version](../../en/design/DP-014-runtime-operational-identity-persistence.md)

## 1. Статус

- **Design Status:** Approved
- **Implementation Status:** Implemented in isolation

Этот approved design определяет durable boundary identity и history Runtime
Instance и Launch Attempt. Package `internal/runtimeidentity` реализует все
девять conceptual operations из §21 и удовлетворяет всем acceptance proofs из
§22 как изолированный in-process in-memory store. Внешний storage, HTTP API,
production wiring и второй lifecycle owner отсутствуют.

## 2. Назначение

Определить минимальный persistence contract, сохраняющий operational identity
model ARCH-004 между lifetime процессов без создания второго Runtime
Lifecycle Owner.

Contract делает durable binding Runtime Instance, history Launch Attempt и
последние подтверждённые Owner desired и actual facts. Stored facts не
становятся доказательством текущей liveness process.

## 3. Источники полномочий

Proposal уточняет, но не переопределяет:

- [ADR-0002](../adr/0002-configuration-dsl.md);
- [ADR-0003](../adr/0003-runtime-architecture.md);
- [ADR-0004](../adr/0004-handshake-runtime-dependencies.md);
- [ARCH-002](../architecture/ARCH-002-runtime-foundation-freeze.md);
- [ARCH-004](../architecture/ARCH-004-runtime-deployment-and-identity-model.md);
- [ARCH-005](../architecture/ARCH-005-runtime-configuration-snapshot-and-loading-model.md);
- [DP-010](DP-010-runtime-lifecycle-owner-contract.md);
- [DP-011](DP-011-runtime-launch-pipeline-integration.md);
- [DP-012](DP-012-runtime-source-composition.md);
- [DP-013](DP-013-runtime-management-routing.md).

Active ARCH-004 имеет приоритет над этим Approved design. DP-010 остаётся
contract process-local Owner и ownership live Host. DP-013 остаётся
Draft по Design Status и реализован изолированно
`internal/runtimemanagement`; production integration отсутствует.

## 4. Scope и non-goals

Этот design определяет:

- один durable aggregate Runtime Instance;
- immutable binding Workspace, Configuration и Runtime Instance;
- append-only history Launch Attempt;
- durable boundaries allocation opaque identity;
- atomic и conditional publication lifecycle facts;
- coherent reads aggregate и history;
- правила failure и indeterminate outcome;
- инварианты concurrency, uniqueness, security и redaction.

Этот design не определяет:

- database, repository product, table, document, index, migration или ORM;
- Go interfaces, packages, HTTP endpoints, DTO или status codes;
- durable idempotency management commands;
- activation, replacement, rollback, recovery или reconciliation;
- operational reporting, logging, metrics, retention audit или alerting;
- deletion, retention, automatic restart, scheduling, clustering или process
  isolation;
- хранение payload Configuration или Secret.

## 5. Термины

**Operational management domain** — одна authority boundary, в которой
Runtime Instance IDs уникальны, а management routing разрешает ID в один
aggregate.

**Aggregate Runtime Instance** — durable consistency boundary с root в одной
identity Runtime Instance.

**Aggregate revision** — opaque монотонно изменяющееся значение, определяющее
одно committed состояние aggregate. Это concurrency token, а не timestamp или
business identity.

**History attempt** — полное append-only membership committed children Launch
Attempt, принадлежащих одному Runtime Instance. Membership, parent identity,
child identity и exact version pin immutable. Lifecycle phase и outcome facts
могут conditionally продвигаться внутри того же child.

**Последний подтверждённый fact** — desired или actual lifecycle fact, явно
опубликованный Runtime Lifecycle Owner в определённой linearization point.

**Indeterminate outcome** означает, что caller не может определить, была ли
requested atomic publication committed.

## 6. Ownership и aggregate boundary

Runtime Instance является aggregate root. Его immutable identity binding,
revision, desired и actual facts, optional reference active attempt и полная
history Launch Attempt образуют одну consistency boundary.

Launch Attempt является owned child ровно одного aggregate Runtime Instance.
Он не является независимо изменяемым aggregate и не может перемещаться между
Runtime Instances.

Runtime Lifecycle Owner остаётся единственным lifecycle decision maker и
единственным owner live reference Runtime Host. Persistence проверяет и
условно публикует facts; она не запускает и не останавливает Host, не выбирает
Configuration, не владеет ресурсами, не маршрутизирует произвольные services
и не становится registry или service locator.

## 7. Durable facts

Aggregate сохраняет только facts, требуемые ARCH-004:

- immutable binding identity Workspace, Configuration и Runtime Instance;
- текущую aggregate revision;
- последний подтверждённый Owner desired state;
- последний подтверждённый Owner actual state;
- optional identity active Launch Attempt;
- полную append-only history Launch Attempt;
- для каждого attempt immutable identity и exact pin Published
  ConfigurationVersion;
- для claimed attempt, который может войти в external preparation, optional
  immutable opaque execution-generation binding, требуемый Approved
  [DP-017](DP-017-runtime-recovery-reconciliation.md);
- committed phase и terminal outcome facts, необходимые для различения
  claimed, running, stop-claimed, stopped и failed attempts.

Persistence contract не копирует payload Configuration, Snapshot, Secret
values, pointer Host, context, goroutine, PID, socket или state Session.

## 8. Opaque IDs и namespaces

RuntimeInstanceID и LaunchAttemptID являются opaque non-zero identities. Их
representation и generation algorithm находятся вне этого design.

RuntimeInstanceID уникален внутри одного operational management domain.
Candidate ID, уже определяющий любой текущий или historical Runtime Instance
в этом domain, отклоняется.

LaunchAttemptID уникален только внутри полной history owning Runtime Instance.
Его durable child key:

```text
(RuntimeInstanceID, LaunchAttemptID)
```

Global uniqueness LaunchAttemptID между разными Runtime Instances или
operational management domains не требуется. Attempt ID никогда не
переиспользуется внутри history одного Runtime Instance, включая состояния
после failure, stop, replacement или будущего retention transition,
сохраняющего identity aggregate.

## 9. Создание Runtime Instance

Creation атомарно публикует:

- один новый RuntimeInstanceID;
- exact immutable binding WorkspaceID и ConfigurationID;
- начальные desired `Stopped` и actual `Stopped` facts;
- отсутствие active attempt;
- пустую history attempt;
- начальную aggregate revision.

Creation публикует полный initial aggregate либо не публикует ничего.
Existing identity, invalid identity, invalid binding или stale creation
precondition выполняет zero mutation. Rebinding существующего Runtime Instance
к другому Workspace или Configuration запрещён.

Allocation candidate identity может происходить до publication, но allocation
само по себе не доказывает существование aggregate.

## 10. Claim Launch Attempt

Claim launch атомарно:

- проверяет expected aggregate identity и revision;
- проверяет, что текущие lifecycle facts разрешают Start;
- проверяет отсутствие active Launch Attempt;
- проверяет отсутствие candidate LaunchAttemptID в полной history;
- добавляет один новый Launch Attempt с exact immutable pin Published
  ConfigurationVersion;
- устанавливает его как единственный active attempt;
- публикует соответствующие desired/actual claim facts;
- один раз продвигает aggregate revision.

Работа Loader, Builder, Launcher или Host не начинается из claim, который не
подтверждён как committed. Retry или replacement создаёт distinct candidate
attempt identity только после того, как outcome предыдущего claim известен.

### Execution generation binding

Control Service composition allocate одну opaque execution generation для
своей process-containment boundary. Persistence не allocate её и не infer из
PID, time, address или caller identity.

После launch claim и до любого Load, Build, Launcher или Host work original
tracked Start path через planned continuation DP-011 может conditionally bind
exact active attempt к этой generation. Publication:

- проверяет aggregate identity и exact expected revision;
- проверяет exact active non-terminal attempt и immutable version pin;
- проверяет отсутствие binding к другой generation;
- сохраняет immutable correlation attempt-to-generation;
- один раз продвигает aggregate revision.

Exact already-present same-generation binding — zero-mutation satisfied
observation. Любой conditional rejection выполняет zero mutation и не разрешает
external preparation, но сам rejection не доказывает absence binding. Different
generation, stale revision, inactive/terminal attempt, conflicting fact или
unavailable store требует coherent exact re-read. Exact existing terminal
outcome converges; different binding/unresolved conflict даёт `Blocked` и не
входит в resource-free failure path.

Только coherent read, доказывающий отсутствие binding у exact still-active
attempt на expected revision, разрешает `BindingFailed`. Сам attempt claim уже
является durable lifecycle mutation. Тот же Start path затем обязан converge
process-local attempt через existing Owner.Start с authentic token и
`FailedPreparation(bindingFailure)`. Mutex Owner упорядочивает failure и
concurrent Stop. Только exact returned Owner outcome может запросить conditional
durable terminal publication: preparation failure публикует resource-free
Failed; выигравший Stop публикует exact Owner-confirmed stopped-before-running
fact. Terminalization command/phase следует только после confirmed publication.

После indeterminate binding publication path coherently inspect exact aggregate,
attempt, candidate generation и expected/new revision. Exact same-generation
presence подтверждает binding; coherently proven absence для exact still-active
attempt/expected revision разрешает Owner-owned failure convergence path, а не
direct persistence mutation, новую generation или blind binding retry;
different generation, stale/conflicting/inactive facts, unavailable state или
unknown остаётся unresolved, если exact existing terminal outcome нельзя
converge. Concurrent Stop сначала
упорядочивается final continuation gate DP-011/DP-016, затем, если winning
binding-failure path, existing mutex Owner против failure acceptance.
Absent/indeterminate Owner/durable terminal outcome остаётся unresolved.

Binding сохраняется в attempt history после terminalization. Он доказывает
только correlation, но не liveness, readiness, preparation, ownership или
graceful shutdown.

## 11. Publication Running

Только Runtime Lifecycle Owner может запросить publication Running после
завершения startup и readiness Host для exact claimed attempt.

Publication Running атомарно проверяет aggregate identity, expected revision,
identity active attempt и допустимую prior phase attempt; публикует actual
`Running` fact attempt и aggregate; сохраняет exact version pin; и один раз
продвигает revision.

Она не может создать attempt, заменить active attempt, восстановить
отсутствующую history или опубликовать Running для stale, terminal или другого
attempt.

## 12. Publication Stop и terminal

Claim Stop атомарно фиксирует transfer shutdown responsibility exact active
attempt, публикует соответствующие desired и phase facts и продвигает
revision. Repeated или stale stop claims не создают другого shutdown owner.

Phase-sensitive publication Stop следует этим правилам:

- attempt в `Preparing`, который не создал owned Host, может атомарно
  опубликовать desired `Stopped`, actual `Stopped` и historical outcome
  stopped-before-running;
- attempt в `Launching` или `Running` сначала публикует same-attempt claim Stop
  и actual `Stopping`, затем публикует terminal outcome только из confirmed
  shutdown result;
- failure Stop или unproven cleanup сохраняет association active attempt и
  публикует только правдивые `Failed` или `Stopping` facts, не заявляющие
  release ресурсов.

Confirmed terminal publication атомарно:

- проверяет exact aggregate, revision и identity attempt;
- фиксирует phase-sensitive stopped или failed outcome;
- публикует последний подтверждённый Owner actual fact;
- очищает reference active attempt только когда Owner доказывает отсутствие
  Host resources или что startup не создал owned Host;
- сохраняет полную immutable history attempt;
- один раз продвигает revision.

Actual `Stopped` публикуется только после подтверждения Owner, что owned Host
resources освобождены или startup не создал их. Fact failure stop-operation
или cleanup-unproven не является terminal historical outcome attempt: attempt
остаётся active в `AttemptStopping`, а его association сохраняется. Terminal
historical outcome `Failed` может быть опубликован только когда
phase-sensitive contract также доказывает отсутствие Host resources или что
startup не создал owned Host.

## 13. Desired, actual и liveness

Desired и actual state остаются раздельными. Desired state фиксирует последнюю
принятую management intent в определённой lifecycle publication. Actual state
фиксирует последний lifecycle fact, подтверждённый Owner.

Persisted actual `Running`, `Stopping`, `Stopped` или `Failed` является
historical operational knowledge. После потери Owner или process Control
Service он не доказывает текущую liveness Host, process, socket или ресурсов.

Только утверждённый recovery и reconciliation contract может сравнить durable
facts с external execution evidence и опубликовать reconciled state. Approved
[DP-017](DP-017-runtime-recovery-reconciliation.md) определяет этот recovery
contract; его Planned implementation остаётся отсутствующей. Этот design не
выводит liveness из stored PID, address, time или более раннего Running fact.

## 14. Atomicity

Отдельными обязательными atomic publications являются:

1. initial aggregate Runtime Instance;
2. claim единственного active Launch Attempt и append history;
3. `ConditionalBindExecutionGeneration` exact claimed attempt;
4. publication Running;
5. claim Stop;
6. phase-sensitive terminal publication.

Каждая publication commit все указанные aggregate и attempt facts с одной
новой revision либо не commit ничего. Observer не может увидеть partial
identity binding, attempt без history, history entry без exact version pin,
два active attempts, historical terminal или cleared attempt, остающийся
active, или lifecycle fact без commit-нувшей его revision. Retained active
`AttemptStopping` с fact stop-failure или cleanup-unproven явно разрешён и не
является historical terminal attempt.

Atomicity является semantic requirement и не предписывает реализацию
transaction, locking, consensus или storage.

## 15. Uniqueness и history

Durable boundary обеспечивает:

1. один immutable aggregate для каждого RuntimeInstanceID в management domain;
2. один permanent binding Workspace и Configuration этого aggregate;
3. не более одного active Launch Attempt;
4. один immutable version pin каждого attempt;
5. отсутствие reuse LaunchAttemptID внутри history aggregate;
6. append-only membership history: committed child никогда не удаляется и не
   заменяется;
7. immutable parent identity, LaunchAttemptID и exact Published
   ConfigurationVersion pin каждого child;
8. lifecycle phase и outcome facts продвигаются только conditionally внутри
   того же child, никогда не регрессируют и не rewrite его immutable facts;
9. отсутствие transition terminal historical attempt обратно в active.

Deletion и retention отложены. Пока focused contract не определит их, design
предполагает доступность identity и attempt history для uniqueness и
inspection.

## 16. Concurrency и conditional revisions

Все mutations одного aggregate Runtime Instance используют одну serialization
boundary и exact expected revision. Successful publication монотонно
продвигает revision. Concurrent operations разных Runtime Instances могут
выполняться независимо.

Stale revision, wrong active attempt, wrong phase или mismatched immutable
binding отклоняется с zero mutation. Persistence boundary не должна молча
re-read и переинтерпретировать operation относительно более нового state,
выбирать latest attempt, merge conflicting lifecycle intents или выполнять
retry detached от caller.

Representation revision, размер increment и storage mechanism не определены.
Единственный observable contract — equality expected revision и monotonic
change после каждой successful publication.

## 17. Failure и indeterminate outcomes

Definitive rejection или definitive commit failure ничего не публикует и
возвращает достаточно category information, чтобы caller мог различить
invalid, missing, stale, conflicting или unavailable outcomes без раскрытия
sensitive data.

При indeterminate outcome caller не должен:

- выполнять blind retry с новым RuntimeInstanceID или LaunchAttemptID;
- предполагать отсутствие mutation;
- запускать Host из unconfirmed claim;
- публиковать более позднюю phase до разрешения более ранней publication.

Caller сначала выполняет coherent inspection exact aggregate identity,
candidate child identity и expected/new revision. Затем он определяет,
присутствует requested publication, отсутствует или всё ещё неизвестна. Это
правило inspect-after-indeterminate предотвращает duplicate identity и
history; оно не определяет command deduplication ARCH-004 section 19(3) или
process recovery section 19(5).

## 18. Coherent reads

Read одного Runtime Instance возвращает одну coherent aggregate revision:
immutable binding, desired и actual facts, reference active attempt и
предоставленные attempt facts не могут происходить из несовместимых revisions.

History reads сохраняют каждый committed attempt и его exact version pin.
Pagination, streaming, snapshots и read consistency mechanisms являются
implementation choices, но не должны создавать вымышленный ordering,
пропускать committed entries при заявленной completeness или представлять
child под другим aggregate.

Reads являются только observation. Они не получают lifecycle ownership, не
обновляют liveness, не исправляют state и не продвигают revision.

## 19. Security и redaction

Каждая operation scoped к exact operational management domain и identity
Runtime Instance. Binding Workspace и Configuration нельзя изменить для
пересечения authorization boundary.

Durable model содержит opaque identities и lifecycle facts, а не:

- Secret или credential values;
- полные payload Configuration или Snapshot;
- authentication material;
- raw internal errors или stack traces;
- pointers Host или process-local capabilities.

Errors и inspection results не раскрывают существование или state другого
unauthorized aggregate. Concrete authorization и operational
reporting/redaction policy остаются отдельными обязательными designs.

## 20. Technology Neutrality

Этот contract может быть реализован любой storage technology, которая
доказывает его requirements atomicity, conditional revision, uniqueness,
coherent read, durability и failure.

Термины aggregate, append, conditional publication и revision являются
semantic. Они не требуют relational tables, document storage, event sourcing,
compare-and-swap instructions, distributed consensus, UUID или конкретного
transaction isolation level.

Implementation может добавить private mechanics, но не должна раскрывать их
как новую architecture или ослаблять contract.

## 21. Conceptual operations

Design требует capability, эквивалентной следующим conceptual operations:

```text
AllocateCandidateIdentity
CreateRuntimeInstance
ReadRuntimeInstance
ReadLaunchAttemptHistory
ConditionalClaimLaunchAttempt
ConditionalBindExecutionGeneration
ConditionalPublishRunning
ConditionalClaimStop
ConditionalPublishTerminal
```

Эти имена являются explanatory, а не definitions API или Go interface.
Operations принимают exact identity и expected revision, где применимо, и
возвращают committed revision или правдивый definitive/indeterminate outcome.

Generic CRUD repository, dynamic entity registry, universal transaction
manager или service locator не разрешены.

## 22. Acceptance proofs

Implementation должна доказать минимум:

1. atomic complete creation Instance и immutable binding;
2. uniqueness RuntimeInstanceID внутри management domain;
3. atomic claim единственного active attempt с exact version pin;
4. uniqueness child key и non-reuse внутри полной history Instance;
5. append-only history после start failure, stop и последующих attempts;
6. exact immutable execution-generation binding после claim и до Load;
7. binding mismatch, stale revision или indeterminate inspection не разрешает
   external preparation;
8. exact conditional publications Running, Stop и terminal;
9. stale и mismatched operations выполняют zero mutation;
10. concurrent claims одного Instance создают не более одной accepted mutation;
11. разные Instances выполняются независимо;
12. coherent reads соответствуют одной committed revision;
13. definitive failure ничего не публикует;
14. indeterminate outcomes разрешаются inspection exact identity/revision без
    blind retry с новым ID;
15. persisted actual state или execution binding никогда не используется как
    liveness proof после потери Owner;
16. redaction и domain isolation предотвращают cross-scope disclosure;
17. не появляется второй lifecycle owner, ownership Host, schema promise или
    hidden service locator.

Proofs включают технически доступные concurrency, race, failure-injection,
durability и restart-of-storage-client scenarios. Они не разрешают production
activation.

## 23. Formal и последующие gates ARCH-004 section 19

Этот Approved design закрывает focused architecture design gate ARCH-004
section 19(2). Approved DP-015–DP-018 закрывают downstream focused design gates
sections 19(3)–(6). Эти status decisions сами по себе не разрешают
implementation.

Conditional aggregate revision предотвращает stale mutation, но не является
client command idempotency. Inspection indeterminate write не является
recovery или reconciliation. Хранение terminal facts не является operational
reporting. Approved [DP-015](DP-015-runtime-management-command-idempotency.md),
[DP-016](DP-016-runtime-activation-replacement-rollback.md),
[DP-017](DP-017-runtime-recovery-reconciliation.md) и
[DP-018](DP-018-runtime-operational-error-reporting-redaction.md) определяют
эти отдельные ответственности. Primitive boundary DP-015 реализован
изолированно `internal/runtimecommandidempotency`; partial parent/phase
sequential core DP-019 также реализован там изолированно, а
Continue/pending-Stop extension и DP-016–DP-018 остаются Planned.

## 24. Явно отложено

За пределы isolated implementations DP-014/DP-015 отложены:

- transport command-key fields, retention/deduplication windows, caller retry
  policy и integrated replay delivery;
- выбор, activation, replacement или rollback version;
- hydration Owner после restart и reconciliation с execution evidence;
- diagnostic taxonomy, redaction policy, audit, metrics и alerting;
- management API create/list/delete и concrete authorization;
- storage product, schema, migrations, serialization, backup и retention;
- automatic restart, scheduling, process supervision и clustering.

Ничто из этого не может быть скрыто внутри identity allocation, aggregate
revision, history или conceptual operations.

## 25. Implementation boundary

Implementation Status — Implemented in isolation. Package
`internal/runtimeidentity` реализует девять conceptual operations §21 как
process-local in-memory Runtime Instance aggregate store с proof conditional
revision и append-only history Launch Attempt. Он не подключён к management
routing DP-013 или production composition.

External durable storage, schema, adapter, API, hydration, recovery, management
wiring и Production Activation отсутствуют. Изолированный package не заявляет
persistence через restart процесса и не меняет downstream integration gates.

## 26. Решение

UWP будет сохранять operational identity как один aggregate Runtime Instance с
immutable binding Workspace/Configuration/Instance, монотонной conditional
revision, последними подтверждёнными Owner desired и actual facts, не более
чем одним active attempt и полной append-only history Launch Attempt.

Каждый Launch Attempt является owned child с key
`(RuntimeInstanceID, LaunchAttemptID)`, pin ровно одну Published
ConfigurationVersion и никогда не переиспользует child identity внутри
history Instance. RuntimeInstanceID уникален внутри operational management
domain.

До external preparation claimed attempt может conditionally получить один
immutable opaque execution-generation binding, owned DP-014. Planned
continuation DP-011 координирует binding capability, DP-016 определяет final
binding/load gate, а DP-017 использует binding во время recovery. Binding
сохраняется как correlation history и никогда не доказывает liveness или
shutdown.

Runtime Lifecycle Owner остаётся единственным lifecycle и live Host owner.
Atomic persistence фиксирует правдивые facts, но не доказывает liveness после
потери Owner. Stale operations выполняют zero mutation; indeterminate outcomes
требуют inspection exact identity и revision без blind retry с новым ID.
Design не выбирает external durable schema, storage technology, integration
API или production composition. Isolated process-local in-memory
implementation остаётся package, описанным в §25.

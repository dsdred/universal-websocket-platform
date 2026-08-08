# DP-017: Восстановление и сверка Runtime

[English version](../../en/design/DP-017-runtime-recovery-reconciliation.md)

## 1. Статус

- **Design Status:** Approved
- **Implementation Status:** Planned

Этот approved design определяет planned boundary восстановления и сверки после потери process-local Runtime
ownership Control Service. Этот документ не создаёт recovery package, store,
schema, execution adapter, API, scanner или production wiring.

## 2. Назначение

Определить, как один Runtime Instance снова становится безопасным для
management после termination Control Service или эквивалентной потери всех
live lifecycle и command permits. Contract сопоставляет exact durable facts с
authoritative execution evidence, публикует только доказанные outcomes и
открывает command admission лишь после terminal и coherent состояния всего
связанного lifecycle/command set.

## 3. Основания

Предложение уточняет, но не переопределяет:

- [ARCH-004](../architecture/ARCH-004-runtime-deployment-and-identity-model.md),
  особенно sections 9, 12, 14 и 19(5);
- [DP-013](DP-013-runtime-management-routing.md) для exact Target routing и
  authorization-before-mutation;
- [DP-014](DP-014-runtime-operational-identity-persistence.md) для coherent
  aggregate revision, attempt history и conditional lifecycle publication;
- [DP-015](DP-015-runtime-management-command-idempotency.md) для command facts,
  non-transferable permits, linked command sets и unresolved admission;
- [DP-016](DP-016-runtime-activation-replacement-rollback.md) для phase order,
  proven release и process-loss cut points.

Accepted ADR и Active/Frozen architecture остаются authoritative. DP-013
остаётся Draft и реализован изолированно. Approved DP-014 и DP-015 реализованы
изолированно; Approved DP-016 и DP-017 остаются Planned.

## 4. Область действия

Design охватывает:

- exact per-Instance restart assessment;
- один durable recovery claim и один non-transferable recovery permit;
- разделение durable facts, live capabilities и execution evidence;
- evidence classification для initial in-process topology;
- reconciliation active Launch Attempt и primitive/linked commands;
- conditional publication order и crash-resumable convergence;
- reopening command admission и unresolved outcomes;
- concurrency, cancellation, security и Technology Neutrality constraints.

Design не определяет discovery scans, database schema, process supervision,
PID files, socket probes, child-worker protocol, public recovery API, transport
mapping, automatic restart, reporting или retention.

## 5. Термины

**Restart assessment** — coherent read, определяющий, уже clean ли один Runtime
Instance или требует recovery до admission любой state-changing command.

**Recovery claim** — durable internal coordination fact, связанный с одним
exact Runtime Instance, его starting aggregate revision и exact revisions всех
non-terminal commands, наблюдавшихся при claim. Это не management command,
Launch Attempt, lease по истечению времени или authority вызвать lifecycle work.

**Recovery permit** — non-transferable process-local capability, возвращаемый
только path, который подтвердил новый или resumable recovery claim. Он
разрешает conditional reconciliation publications для этого claim, но никогда
не разрешает Start, Stop, Load, Build, Launch, Host adoption или command replay.

**Execution generation** — opaque identity одной execution-containment
boundary. В initial topology он обозначает одну Control Service process
generation. Это не PID, timestamp, liveness lease, Runtime identity или command
identity.

**Execution binding** — durable immutable correlation одного exact Launch
Attempt с execution generation, которой разрешено prepare или own его Host. Он
confirmed после attempt claim и до Load. Он доказывает correlation, а не
liveness или начало preparation.

**Execution evidence** — observation от execution boundary, чей contract может
связать его с exact Launch Attempt и execution generation и различить proven
termination, proven live execution, proven Host-owned shutdown completion, где
это поддержано, и unknown. PID, address, time, stored actual state или успешное
соединение сами по себе таким evidence не являются.

**Clean set** означает отсутствие active attempt, non-terminal command/phase и
unresolved recovery claim при coherent terminal facts на одном verified наборе
revisions.

**Unresolved set** означает, что любой exact aggregate, attempt, primitive
command, parent, phase, recovery claim или execution fact, необходимый для
truthful outcome, missing, contradictory, stale, unavailable или indeterminate.

## 6. Граница ответственности

Recovery владеет assessment, exclusive reconciliation coordination, evidence
classification, conditional repair publication и финальным решением сохранить
или открыть per-Instance admission barrier. Он не владеет Runtime lifecycle
policy, live Host resources, version selection, authorization policy,
transport, diagnostics presentation или automatic execution.

Runtime Lifecycle Owner остаётся единственным normal lifecycle decision maker
и live Host owner. Recovery никогда не hydrate Owner из stored fields и не
переносит старый execution/phase permit. Persistence проверяет facts, execution
evidence сообщает execution truth. Ни один из них не становится вторым Owner.

Exact Control Service composition создаёт opaque execution generation и владеет
proof её containment/termination. DP-014 владеет conditional durable execution
binding внутри aggregate Runtime Instance. Planned Start-claim continuation
DP-011 координирует borrowed binding capability и pending-Stop gate DP-015.
DP-017 читает binding при assessment, но не allocate generation/binding.

## 7. Триггер восстановления и оценка

Assessment обязателен до state-changing command, когда current Control Service
process не может доказать process-local ownership, соответствующий durable
non-terminal facts. Сюда входят process start, loss Owner/command call stack,
indeterminate lifecycle/command publication и existing unresolved recovery
claim.

Один coherent assessment читает:

- exact Runtime Instance identity, binding, revision, desired/actual facts;
- active-attempt identity и его exact phase, outcome и version pin;
- все non-terminal primitive, parent и phase commands этого Instance;
- exact command revisions, immutable intents и linked identities;
- immutable execution binding, если он committed для active attempt;
- existing recovery claim и его observed-revision set;
- available execution evidence для exact attempt/generation.

Clean set не требует recovery mutation. Read-only Observe остаётся доступным с
обычной authorization и truthful маркировкой stale facts. Assessment не
изменяет lifecycle state и не доказывает liveness.

## 8. Fail-closed admission

От первого наблюдения non-clean set до verified release recovery claim каждая
новая state-changing command и каждая дальнейшая parent phase отклоняется без
mutation. Tracked-Start Stop exception не переживает process loss. Same-key
observation может сообщить non-terminal status, но не получает permit и не
делегирует work.

Admission и creation recovery claim разделяют одну per-Instance atomic ordering
boundary. Command не проходит из stale clean observation одновременно с
recovery claim, а два recovery path не могут одновременно получить permits.

## 9. Durable recovery claim

Claim recovery атомарно проверяет exact Instance/command revisions, подтверждает
non-clean set, устанавливает один recovery claim, закрывает command admission и
возвращает один live recovery permit. Definitive conflict или stale revision
выполняет zero mutation и требует reassessment.

Existing claim может быть resumed только новым conditional recovery step,
который повторно читает все exact facts и устанавливает, что old permit не мог
сохраниться. Он не переиспользует и не пересоздаёт этот permit. Loss recovery
permit оставляет claim unresolved и admission closed; следующий recovery pass
продолжает из durable publications, а не повторяет lifecycle work.

Representation recovery identity, storage mechanism и claim layout остаются
implementation decisions. Wall-clock timeout сам по себе не доказывает permit
loss, execution termination или safe takeover.

## 10. Иерархия evidence

Recovery отдельно оценивает три класса truth:

1. durable facts DP-014/DP-015 доказывают лишь committed состояние на exact
   revisions;
2. current process-local capabilities доказывают лишь ownership, созданный в
   current process, и не восстанавливаются из durable identity;
3. execution evidence доказывает лишь liveness/termination property,
   гарантированное approved adapter contract.

Один класс не заменяет другой. Persisted `Running`, `Stopping`, command
`Claimed`, PID, Listener address, port-bind result, elapsed time, log record,
health response или configuration version сами по себе не доказывают present
Host ownership, resource release или lifecycle completion.

Contradictory evidence остаётся unresolved. Recovery не выбирает самое новое,
удобное или majority observation.

До любого Load, Build, Launcher или Host work planned continuation DP-011
должен подтвердить DP-014 publication одного execution binding для exact
already-claimed attempt и expected aggregate revision. Только coherent exact
read, доказывающий отсутствие binding у этого exact still-active attempt на
expected revision, может дать `BindingFailed`: external preparation не
начинается, но уже committed attempt/Starting mutation не стирается. Different
generation, stale revision, conflicting или inactive state, unavailable store
либо unknown result требуют exact re-read и затем exact terminal convergence
или `Blocked`, но никогда не `BindingFailed`. Final per-Instance gate
continuation затем упорядочивает Stop и release к Load. Этот prerequisite
уточняет planned DP-016 ordering, не разрешая implementation.

Пока Owner live, такое coherently proven exact binding absence не является
recovery и не разрешает direct durable terminalization. Continuation возвращает
`BindingFailed`; Flow передаёт `FailedPreparation` с authentic token в existing
Owner.Start. Mutex Owner упорядочивает failure и Stop. Только exact Owner outcome
может публиковаться durably и затем terminalize command. DP-017 обрабатывает
unbound attempt лишь после потери этой Owner convergence.

## 11. Initial in-process topology

В текущей ARCH-004 single-node in-process topology Runtime Host не может
пережить termination владевшего им process. Approved containment boundary
сначала должен установить unique current generation и доказать termination
exact prior generation из execution binding attempt. Это доказывает, что Host
prior generation больше не owned и не runnable. Это не доказывает graceful
Runtime cleanup, successful Stop, readiness или потерянную terminal command
publication.

Replacement Control Service не фабрикует Host reference, не hydrate Owner, не
probe port с выводом Running и не adopt execution. Proven generation termination
используется только для phase-sensitive process-loss failure facts и clearing
active attempt association, когда все необходимые exact facts coherent.
Resource absence отдельно никогда не доказывает successful Stop или Host
shutdown completion.

Если future child-process/remote adapter сообщает live orphan, такое execution
остаётся non-manageable, а admission closed до отдельного approved adoption или
termination protocol. DP-017 не разрешает этот protocol.

## 12. Классификация восстановления

Assessment или claimed recovery path классифицирует exact set:

- **Clean:** все attempts/commands terminal и recovery claim отсутствует;
  assessment не создаёт claim или reconciliation mutation;
- **Release-only:** все lifecycle/command facts terminal и coherent, и только
  exact recovery claim осталось conditionally release;
- **Command only:** command claimed, но exact aggregate/attempt facts доказывают,
  что Owner не claim Launch Attempt и lifecycle mutation не началась;
- **Unbound attempt:** Owner claim exact attempt и publish Starting, но exact
  conditional read доказывает, что execution binding не commit; это доказывает
  no external preparation, а не no lifecycle mutation;
- **Execution terminated:** active attempt существовал, и authoritative
  evidence доказывает termination его exact execution generation;
- **Resource absence:** Stop/replacement release phase шла, и evidence с exact
  aggregate facts доказывают отсутствие old Host resources, но не Host shutdown
  completion;
- **Shutdown completed:** exact authoritative evidence доказывает completion
  Host-owned shutdown contract для exact attempt;
- **Live orphan:** authoritative evidence сообщает execution без ownership
  current process;
- **Unknown:** evidence absent, stale, contradictory, indeterminate или не
  связывается с exact attempt/generation.

Clean возвращается до claim. Release-only, Command only, Unbound attempt,
Execution terminated и evidence-backed shutdown completion могут продвигаться
в initial topology. Resource absence без shutdown-completion proof, Live orphan
и Unknown сохраняют truthful failure/unresolved outcomes и не фабрикуют Stopped.

## 13. Phase-sensitive reconciliation attempt

Recovery никогда не создаёт и не переиспользует Launch Attempt. Definitively
unbound active attempt уже является durable lifecycle mutation в actual
`Starting`. Поскольку gate DP-011 запрещает external preparation до binding,
recovery может conditionally publish exact attempt historical `Failed` со
stable unprepared-process-loss category и clear его как resource-free. Если
binding publication indeterminate или не читается coherently, attempt остаётся
active и unresolved.

Для proven terminated bound generation:

- attempt для desired `Running` без confirmed resource-free terminal fact
  становится historical `Failed` со stable process-loss category;
- persisted `Running` становится actual `Failed`, никогда не restored Running;
- `Preparing`/`Launching` становится failed даже без readiness publication;
  recovery не выводит, насколько далеко прошёл startup;
- Stop-claimed attempt становится resource-free `Failed`/interrupted, когда
  generation termination доказывает только resource absence;
- historical Stopped разрешён только если exact authoritative evidence
  доказывает completion Host-owned shutdown contract этого attempt, а не просто
  исчезновение process/resources;
- иначе active association и truthful non-terminal/Failed fact сохраняются, а
  set остаётся unresolved.

Clearing active-attempt reference использует exact DP-014 revision и разрешён
только при proven resource absence. Historical identity/version pin неизменны.

## 14. Reconciliation primitive command

Claimed primitive command terminalizes только из exact durable lifecycle facts:

- если Owner не claim attempt, публикуется stable recovered-no-mutation outcome;
- если Owner claim exact attempt, но binding definitively absent, сначала
  terminalize этот resource-free attempt Failed, затем публикуется stable
  failed-before-preparation command outcome;
- если Start владеет exact attempt, terminated при process loss, публикуется
  stable failed outcome, связанный с attempt после его terminal publication;
- если Stop target — exact attempt и доказана только generation termination/
  resource absence, после terminal Failed publication публикуется stable
  interrupted/process-loss failure;
- stopped/satisfied публикуется только при exact proof completion Host-owned
  shutdown contract;
- если command-to-attempt relationship или outcome не доказаны, Claimed и
  closed barrier сохраняются.

Recovery не вызывает DP-013, Flow или Owner, не replay command и не фабрикует
identities, отсутствующие в existing outcome contract.

## 15. Reconciliation parent и phase

DP-016 parent orchestration сверяется из immutable linked phase set:

- каждая claimed phase resolved до parent;
- unclaimed later phase остаётся absent и никогда не создаётся recovery;
- old-Stop с exact Host-shutdown completion и без Start phase оставляет Instance
  Stopped и terminalizes parent как interrupted after release;
- old-Stop только с process termination/resource absence оставляет old attempt
  и Instance Failed и terminalizes phase/parent interrupted by process loss;
  Start phase не создаётся;
- claimed Start phase с terminated exact attempt terminalizes failed после
  attempt/phase facts;
- pending Stop без surviving permit resolves только при proven exact
  no-mutation, Host-shutdown-complete или process-loss failure outcome;
- missing phase link, indeterminate publication или contradictory ordering
  сохраняет весь linked set unresolved.

Parent становится Terminal лишь после terminal всех existing phases и их
outcomes, образующих один valid DP-016 ordering. Recovery не продолжает later
phase.

## 16. Порядок reconciliation publications

Distributed transaction между aggregate/command stores не предполагается.
Пока durable recovery barrier закрыт, publications идут monotonic:

1. повторно прочитать и проверить exact aggregate, attempt, command и recovery
   revisions;
2. conditionally publish phase-sensitive attempt/aggregate terminal fact;
3. conditionally terminalize primitive или linked phase commands из этого fact;
4. conditionally terminalize parent после всех existing phases;
5. coherently проверить весь set на resulting revisions;
6. release recovery claim и открыть admission атомарно относительно exact
   verified set.

Каждый step idempotent по exact identity, immutable outcome и conditional
revision. Crash/indeterminate publication на любом step оставляет admission
closed. Следующий pass inspect и resume, никогда не повторяя lifecycle work.

## 17. Открытие barrier

Barrier открывается только когда один coherent verification доказывает:

- active attempt отсутствует;
- каждый primitive command, parent и existing phase Terminal;
- relevant contradictory/unknown execution evidence отсутствует;
- aggregate, attempt, command и recovery revisions совпадают с verified set;
- release recovery claim committed.

Partial command terminalization, cleared attempt отдельно или successful
publication response с indeterminate commit admission не открывают. После
reopening последующий explicitly authorized Start может создать fresh attempt
через обычный DP-015/DP-016 ordering. Recovery этого никогда не делает.

## 18. Desired и actual facts

Recovery не вводит новый public desired/actual state. Desired остаётся last
accepted management intent. Actual reconciles существующими
`Stopped|Starting|Running|Stopping|Failed` facts:

- proven process loss при desired Running публикует actual Failed;
- exact proof completion Host-owned shutdown contract для accepted desired
  Stopped может публиковать actual Stopped;
- process termination/resource absence без такого proof публикует actual
  Failed/interrupted, никогда Stopped;
- unknown release никогда не публикует Stopped;
- persisted Running не refresh/preserve как live только из-за durability.

Recovery categories и claim status — internal coordination/audit facts, не
дополнительные Runtime lifecycle states.

## 19. Concurrency и linearization

Обязательные semantic linearization points:

1. assessment против new command admission;
2. один recovery claim и один newly issued recovery permit;
3. каждая exact conditional attempt/command reconciliation publication;
4. final coherent-set verification и recovery release против new admission.

Different Runtime Instances прогрессируют независимо. Long evidence collection
не держит aggregate/command locks, но revisions revalidate до каждой
publication. Concurrent observer никогда не получает mutation authority.

## 20. Матрица failure и indeterminate outcomes

| Failure cut | Truthful result | Forbidden consequence |
| --- | --- | --- |
| clean assessment | zero recovery mutation | create recovery claim или lifecycle work |
| stale assessment/claim conflict | reassess exact revisions | overwrite newer facts |
| recovery permit loss | recovery claim unresolved | recreate permit по timeout |
| command claimed, attempt отсутствует | terminal outcome recovered no-lifecycle-mutation | create attempt или invoke lifecycle |
| active attempt, binding definitively absent | resource-free Failed attempt и failed-before-preparation command | назвать claim no mutation или начать Load |
| binding publication/inspection indeterminate | active attempt и linked set unresolved | infer absence или bind new generation |
| execution generation termination proven | reconcile exact attempt phase | заявить graceful cleanup |
| Stop claimed, доказана только resource absence | Failed/interrupted process-loss outcome | publish Stopped или stopped/satisfied |
| exact Host shutdown completion proven | phase-sensitive Stopped может publish | infer proof только из process termination |
| live orphan observed | barrier remains closed | hydrate Owner или adopt Host |
| evidence unavailable/contradictory | unresolved | infer release из PID, port или time |
| attempt publication indeterminate | inspect exact revision | terminalize commands из assumption |
| phase terminal, parent publication lost | resume из durable phase | rerun Stop или Start |
| все commands terminal, release indeterminate | barrier remains closed | admit new command |
| recovery cancellation | caller может прекратить wait; claim остаётся | erase claim или transfer permit |

## 21. Caller cancellation

Cancellation до recovery claim выполняет zero mutation. После claim он может
завершить только wait текущего caller. Он не удаляет recovery claim, не
доказывает execution outcome, не открывает admission и не переносит permit.

Если cancellation wins до conditional publication, path может вернуть claim
unresolved. Если publication могла commit, exact inspection обязателен.
Следующий pass resume из durable truth.

## 22. Security и scope isolation

Assessment/reconciliation scoped к одному exact operational management domain,
Workspace, Configuration, Runtime Instance, attempt и linked command set.
Recovery не использует cross-Workspace evidence и не раскрывает существование
unauthorized identity.

Evidence/outcomes содержат opaque identities и stable semantic categories, но
не credentials, Secrets, Configuration/Snapshot payloads, raw internal errors,
stack traces, Host pointers, process-local permits или unrestricted process
metadata. Concrete operator reporting/redaction остаются section 19(6).

## 23. Technology Neutrality

Recovery claim, revision, generation, evidence, conditional publication и
barrier — semantic requirements. Они не требуют database product, transaction
coordinator, lease service, workflow engine, queue, PID file, signal mechanism,
process supervisor, clock или identifier format.

Implementation может добавить private mechanics только при доказательстве exact
binding, single recovery authority, crash-resumable publication и fail-closed
admission без generic registry/service locator.

## 24. Концептуальные операции

Design требует capabilities, эквивалентные:

```text
AssessRuntimeRecovery
ConditionalClaimRecovery
ReadExactExecutionEvidence
ReadExactExecutionBinding
ConditionalReconcileUnboundAttempt
ConditionalReconcileAttempt
ConditionalReconcileCommandOrPhase
ConditionalReconcileParent
ConditionalReleaseRecovery
```

Это explanatory semantic capabilities, не API или Go interface names. Ни одна
не вызывает Runtime lifecycle work.

## 25. Проверки приёмки

Future implementation должна доказать как минимум:

1. clean Stopped Instance выполняет zero recovery mutation;
2. non-clean assessment атомарно исключает new command admission;
3. concurrent recovery paths выдают не более одного live recovery permit;
4. persisted Running, PID, port, time и health response по отдельности не
   доказывают ownership/liveness;
5. execution binding commits до любого Load и никогда не доказывает liveness;
6. coherently proven exact binding absence для exact still-active attempt на
   expected revision converge через Owner.Start и race Stop под mutex Owner до
   durable command terminalization;
7. exact in-process generation termination не разрешает Host adoption;
8. process loss до Owner claim resolves как no lifecycle mutation;
9. definitively absent binding после Owner claim resolves exact already-Starting
   attempt как resource-free Failed, никогда no mutation;
10. indeterminate binding остаётся active/unresolved без Load;
11. process loss во время Starting terminalizes exact attempt Failed только
   после proven resource absence;
12. process loss после Running никогда не сохраняет Running как current truth;
13. Stop-in-progress публикует Stopped только из exact proof Host-owned shutdown
    completion, никогда только resource absence;
14. primitive command outcome следует exact reconciled attempt fact;
15. каждая existing phase terminalizes до DP-016 parent;
16. recovery не claims missing later phase и не вызывает DP-013;
17. crash после каждой publication resumes без lifecycle replay;
18. stale revisions/contradictory evidence выполняют no mutation;
19. barrier открывается только для coherent fully terminal set;
20. cancellation/indeterminate outcomes оставляют admission closed;
21. different Instances recover независимо;
22. EN/RU contract, matrices, gates и Planned status aligned.

Proofs включают технически доступные concurrency, race, failure-injection,
durability, process-restart и recovery-restart scenarios. Они не разрешают
production activation.

## 26. Formal и downstream gates ARCH-004 section 19

Этот Approved design закрывает focused architecture design gate ARCH-004
section 19(5). Approved DP-014–DP-016 и DP-018 закрывают остальные focused
design gates sections 19(2)–(4) и 19(6). Полный approved set определяет
fail-closed recovery ordering. Isolated process-local stores DP-014/DP-015
существуют, но process-restart recovery store, execution adapter, recovery
executor, reporting, integration и Production Activation отсутствуют.

## 27. Явно отложенные вопросы

Отложены до focused designs или implementation tasks:

- startup discovery, enumeration, scheduling и recovery worker lifecycle;
- storage schema, transaction mechanics, migrations и adapters;
- child/remote process protocol, supervision, adoption и forced termination;
- automatic restart, rollback, retry/backoff, failover и policy evaluation;
- public recovery controls, API/DTO/status mapping и concrete authorization;
- operational error taxonomy, reporting, audit, metrics, alerting и redaction
  policy;
- retention/deletion и Production Activation.

## 28. Граница реализации

Implementation Status — Planned. Repository содержит isolated process-local
Runtime aggregate и command stores, но не содержит external durable или
process-restart store, recovery claim, execution-evidence adapter, recovery
executor, public management API или production wiring.

Текущие in-process Runtime components не переживают Control Service process
termination и не предоставляют restart-time recovery capability. Создание
approval закрывает design gate section 19(5), но не реализует и не подключает
contract.

## 29. Решение

UWP восстанавливает один Runtime Instance через exact durable fail-closed
reconciliation claim. Durable lifecycle/command facts остаются last-confirmed
history; current liveness/release требует authoritative execution evidence,
связанного с exact attempt и generation. Stored Running fact, PID, address,
clock или probe result отдельно не пересоздают ownership.

Recovery не выполняет Start, Stop, Load, adoption, retry или new Launch Attempt.
Он conditionally terminalizes exact attempt и linked command set из proven
facts, остаётся crash-resumable при closed admission barrier и открывает barrier
только после coherent fully terminal verification. Unknown/contradictory truth
остаётся unresolved. Automatic restart и operational reporting остаются
отдельными решениями.

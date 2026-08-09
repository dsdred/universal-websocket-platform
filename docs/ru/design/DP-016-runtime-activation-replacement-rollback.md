# DP-016: Runtime Activation, Replacement, and Rollback

[English version](../../en/design/DP-016-runtime-activation-replacement-rollback.md)

## 1. Статус

- **Design Status:** Approved
- **Implementation Status:** Planned

Этот approved design определяет planned
ordering contract activation, replacement и explicit rollback одного Runtime
Instance. Этот документ не создаёт activation/replacement orchestrator или
его workflow persistence, API, recovery worker или production wiring.
Implementation architecture-blocked, пока оставшиеся exact authorization,
private managed Start-claim continuation, Owner-claim view и binding DP-014 из
Approved/Planned DP-019 не реализованы и независимо не приняты.

## 2. Назначение

Определить, как одна exact Published ConfigurationVersion становится source
execution Runtime Instance, как active execution заменяется без overlap Host и
как caller явно выполняет rollback к другой exact Published version без reuse
history или создания automatic fallback policy.

## 3. Authority

Proposal уточняет, но не переопределяет:

- [ARCH-004](../architecture/ARCH-004-runtime-deployment-and-identity-model.md),
  особенно section 19(4);
- [DP-013](DP-013-runtime-management-routing.md) для exact routing и
  authorization-before-mutation;
- [DP-014](DP-014-runtime-operational-identity-persistence.md) для aggregate,
  Launch Attempt, revision и semantics lifecycle publication;
- [DP-015](DP-015-runtime-management-command-idempotency.md) для durable
  command identity, execution permits, replay и unresolved barriers.
- [DP-011](DP-011-runtime-launch-pipeline-integration.md) для Owner-owned Start
  claim и planned private extension claim-continuation.
- [DP-017](DP-017-runtime-recovery-reconciliation.md) для recovery boundary,
  которая требует и использует DP-014-owned execution-generation binding.
- [DP-019](DP-019-runtime-activation-orchestration-prerequisites.md) для exact
  internal API, связывающего authorization, parent/phase admission, Owner
  claim, durable attempt publication, generation binding и Continue.

Принятые ADR и Active/Frozen architecture остаются authoritative. DP-013
остаётся Draft и реализован изолированно. Approved DP-014 и primitive boundary
DP-015, partial parent/phase sequential core DP-019 и command-boundary
Continue/pending-Stop rendezvous реализованы изолированно; managed continuation
и Approved DP-016/DP-017 остаются Planned.

## 4. Область

Design охватывает initial activation, already-satisfied activation, ordered
replacement, explicit rollback, exact target-version selection, Stop races,
caller cancellation, failure cut points и indeterminate outcomes.

Design не определяет public commands, HTTP resources, storage schema,
zero-downtime replacement, overlapping Hosts, in-place reload, automatic
rollback, restart policy, recovery, reconciliation, reporting или production
activation.

## 5. Термины

**Activation target** — exact Published ConfigurationVersion, выбранная одним
authorized immutable command intent для одного Runtime Instance.

**Activation** — ordered creation и Start одного нового Launch Attempt из exact
target, когда не требуется сначала release conflicting active execution.

**Replacement** — один orchestration intent, освобождающий другой active или
starting attempt до activation нового attempt с pin target.

**Rollback** — explicit replacement, target которого является exact более
ранней Published ConfigurationVersion. Это не reversal history и не
resurrection старого Launch Attempt.

**Proven release** означает подтверждение Lifecycle Owner, что предыдущий
attempt не владеет Host resources согласно ARCH-004 и DP-014.

**Continue gate** — per-Instance linearization boundary между accepted Stop
intent и claim linked Start phase orchestration.

**Parent orchestration command** — durable replacement или rollback intent. Его
process-local permit может продвигать только конечные linked phases, определённые
здесь, и никогда не вызывает DP-013 прямо.

**Linked phase claim** — durable parent-derived fact `Stop old` или `Start
target`. Его newly issued phase permit разрешает не более одного exact
invocation DP-013. У него нет caller-selected key, и он не переиспользуется
между phases.

## 6. Граница ответственности

Runtime Lifecycle Owner остаётся sole lifecycle decision maker и owner live
Host reference. Он выполняет или сводит Start/Stop exact attempts.
Orchestration contract упорядочивает существующие responsibilities, но не
становится вторым lifecycle owner.

Persistence DP-014 фиксирует truthful aggregate и attempt facts. DP-015 владеет
command claim, execution permit, replay и unresolved barriers. DP-016 владеет
только bounded parent/phase shape, ordering и preconditions между этими
boundaries. Parent permit не является lifecycle authority; только linked phase
permit достигает DP-013, где authoritative остаётся Lifecycle Owner. Runtime
Host не читает management state, не выбирает versions и не выполняет
replacement.

## 7. Поддерживаемые intents

Design определяет semantic intents, а не public API:

- activate exact target, когда не требуется release другого active attempt;
- replace current exact attempt новым attempt с pin другой exact target;
- явно roll back через replacement к caller-selected exact Published version.

Observe и Stop сохраняют существующие semantics. Create, Delete, Restart,
automatic rollback, latest-version activation и policy-driven deployment не
вводятся.

## 8. Preconditions target

До durable command claim или lifecycle mutation boundary проверяет exact
Workspace, Configuration и Runtime Instance Target, а также что requested
ConfigurationVersion exact, non-zero, Published и принадлежит этой
Configuration.

Authorization текущего caller для exact action, Target и target version
происходит до command mutation. Latest, previous, recommended, fallback или
inferred version не выбирается. Failed validation, lookup или authorization
выполняет zero command и lifecycle mutation.

## 9. Coherent starting observation

Ordering начинается с одного coherent observation aggregate Runtime Instance и
current Owner facts. Он включает expected aggregate revision, desired/actual
facts, identity active attempt, когда она существует, exact version pin и
наличие ownership соответствующей live operation текущим process.

Persisted Running или Stopping fact после потери Owner не является liveness
proof. Если live ownership не установлена, activation, replacement или rollback
не продолжается; authoritative остаётся unresolved barrier DP-015, а Approved
[DP-017](DP-017-runtime-recovery-reconciliation.md) определяет planned recovery
contract section 19(5); его implementation остаётся отсутствующей.

## 10. Initial activation

Из confirmed `Stopped` без active attempt или eligible resource-free `Failed`
activation следует этому порядку:

```text
authorize exact target
    -> claim durable command and execution permit
    -> revalidate aggregate identity and revision
    -> claim one new Launch Attempt with exact target pin
    -> подтвердить DP-014-owned execution-generation binding, required DP-017
    -> Load -> Build -> Start through the existing Flow and Owner
    -> publish Running only after Host readiness
    -> publish replay-equivalent command outcome
```

Loader, Builder, Launcher или Host work не начинается до confirmed command и
Launch Attempt claims. Принимается не более одного attempt. Startup failure
сохраняется как truthful outcome этого attempt и никогда не публикует Running.
Initial activation является одним primitive Start command по DP-015, а не
two-phase parent orchestration.

## 11. Already-satisfied и in-progress targets

Если Runtime Instance confirmed Running на exact target version с desired
Running, activation уже satisfied и выполняет zero lifecycle/aggregate
mutation. Command может опубликовать terminal satisfied outcome.

Если exact target уже Starting под tracked execution, новый attempt не
создаётся. Same-command replay наблюдает existing command; другой activation
intent получает non-mutating in-progress outcome.

Running или Starting на другой target никогда не считается satisfied молча.
Требуется ordered replacement.

## 12. Ordering replacement

Replacement сохраняет identity Runtime Instance и использует строгий порядок:

```text
claim exact parent replacement command and orchestration permit
    -> claim linked Stop-old phase and its one DP-013 Stop permit
    -> invoke or converge Stop for the old active attempt
    -> prove release of all old Host resources
    -> publish old attempt terminal, aggregate Stopped, and Stop-phase outcome
    -> pass the Continue gate
    -> claim linked Start-target phase and its one DP-013 Start permit
    -> invoke DP-013 Start; Owner claims the exact new Launch Attempt
    -> resolve pending Stop at the private Start-claim continuation before Load
    -> подтвердить DP-014-owned execution-generation binding, required DP-017, до Load
    -> publish Start-phase and parent outcomes truthfully
```

Новый attempt имеет fresh identity и immutable target pin. Он никогда не
переиспользует old attempt, Host, Snapshot, context, Listener или readiness.
Service gap между proven release и new readiness принимается initial
single-node contract.

## 13. Replacement во время Starting

Если old active attempt ещё Starting на другой target, command admission
атомарно claim parent replacement и его linked Stop-old phase в single
exception tracked-Start Stop DP-015. Permit этой phase вызывает Stop, который
захватывает тот же attempt, предотвращает Running publication и ждёт
convergence startup rollback или shutdown Host. Independent Stop не может
одновременно занять exception.

Replacement attempt не claim до тех пор, пока old attempt не historical и
release resources не proven. Если old attempt Starting на exact target,
применяется section 11 и replacement ничего не создаёт.

## 14. Existing Stop и Stopping state

Replacement/rollback не обходят already claimed independent Stop. Пока active
attempt Stopping под другим command, replacement/rollback не может claim parent
или новый attempt и получает non-mutating in-progress outcome по admission
DP-015. Same-parent replay только наблюдает exact tracked parent и linked phase.

Только confirmed terminal Stop и proven release разрешают evaluation Continue
gate. Stop failure или unproven cleanup сохраняет active attempt и блокирует
replacement.

## 15. Continue gate и concurrent Stop

Непосредственно перед продолжением replacement/rollback одна per-Instance
atomic boundary упорядочивает claim newly authorized independent Stop и claim
linked Start-target phase parent:

- если Stop linearizes первым, orchestration сохраняет stopped/cancelled
  terminal outcome, а Start phase и новый attempt не claim;
- если Start-phase claim linearizes первым, permit phase является единственной
  authority для одного Start DP-013; ровно один более поздний Stop может claim
  существующий exception tracked-Start и становится pending до claim Owner;
- concurrent observations не видят одновременно “Stop won with no attempt” и
  accepted Start phase.

То же правило закрывает gap после old release и до Start-phase claim.
Definitive failure phase claim оставляет parent terminal и Instance Stopped;
indeterminate claim делает linked command set unresolved. Затем DP-013 вызывает
existing Flow, и только Owner может claim Launch Attempt. Gate не pre-create
attempt. Planned private continuation ниже является coordination seam
observation claim, а не ownership handoff.

Planned private continuation Start-claim DP-011/DP-013 выполняется синхронно
после claim Owner и до Load. При pending Stop continuation сигнализирует original
blocked claimant Stop. Тот call stack сохраняет permit, проверяет cancellation,
единолично вызывает exact Stop DP-013, публикует outcome и сигнализирует result.
Continuation никогда не получает Stop permit или caller context.

Если pending Stop не остаётся, continuation conditionally publish exact
execution-generation binding DP-014. Exact same-generation presence продолжает.
Только отсутствие, доказанное coherent exact read для exact всё ещё active
attempt на expected revision, входит в final Stop-ordering gate и может вернуть
`BindingFailed` без terminal publication. Другое generation, stale revision,
conflicting или inactive state, unavailable store либо unknown result требуют
exact re-read и затем либо convergence к exact terminal outcome, либо `Blocked`;
ни один из этих случаев не становится `BindingFailed`. Flow передаёт failure
через existing Owner.Start с
authentic preparation token. Mutex Owner упорядочивает failure acceptance и
later Stop, и только exact returned outcome может вести DP-014 и terminal
publication command/phase. После confirmed
binding final per-Instance gate упорядочивает new Stop claim и `Continue`.
Выигравший Stop converge до Load; выигравший `Continue` releases Flow, а later
Stop обычно достигает claimed attempt. Admission/Owner lock не удерживается во
время persistence, wait или convergence Stop. Current isolated Flow не реализует
continuation/binding gate, поэтому DP-016 остаётся Planned.

## 16. Explicit rollback

Rollback принимается только как новый authorized/idempotent intent, называющий
одну exact Published ConfigurationVersion той же Configuration. Target могла
использоваться historical attempts, но rollback создаёт fresh command identity
и fresh Launch Attempt identity.

Previous-version pointer, stack, latest-minus-one rule, timestamp ordering или
automatic candidate не выводятся. Если exact rollback target уже Running,
result satisfied с zero mutation. Иначе rollback следует тому же replacement
ordering и failure rules, что любая другая target.

## 17. Desired и actual facts

Contract не вводит `Replacing`, `RollingBack`, `Activating` или другой новый
operational state. Во время release old attempt existing Stop semantics
публикуют desired Stopped и actual Stopping, затем actual Stopped только после
proven release.

Claim нового attempt публикует desired Running и actual Starting через DP-014.
Running публикуется только после readiness. Immutable command intent сохраняет
overall target между phases; он не falsify current desired/actual aggregate
fact.

## 18. Linearization points

Обязательные semantic linearization points:

1. durable primitive command claim или parent orchestration claim и permit;
2. linked old-attempt Stop phase claim и phase permit, когда active attempt
   существует;
3. old-attempt terminal publication с proven release;
4. Continue gate ordering claim independent Stop и claim linked Start phase;
5. один invocation DP-013 newly issued Start-phase permit, после которого Owner
   claim exact new Launch Attempt/version pin;
6. conditional binding DP-014 exact attempt к composition-owned execution
   generation, включая exact inspection после indeterminate;
7. final Start-claim continuation gate ordering pending Stop против release
   confirmed binding в external preparation;
8. publication Running или terminal failure;
9. publication linked phase и replay-equivalent terminal parent.

Каждая publication DP-014 использует exact expected aggregate revision. Long
Load, Build, Start, Stop или wait work не выполняется под command-admission или
aggregate lock.

## 19. Failure matrix

| Failure cut | Truthful result | Forbidden consequence |
| --- | --- | --- |
| validation/lookup/authorization | zero mutation | command claim или lifecycle call |
| command claim definitive failure | zero lifecycle mutation | detached retry |
| linked Stop phase claim definitive failure | parent становится terminal без Stop | lifecycle invocation |
| old Stop failure или cleanup unproven | old attempt остаётся active/Stopping или Failed truthfully | new attempt claim |
| old Stop indeterminate | command остаётся unresolved | assume release или continue |
| proven release, Start-phase claim definitive failure | aggregate остаётся Stopped, old attempt historical | resurrect old Host |
| Start-phase claim indeterminate | linked command set остаётся unresolved | Start, Stop exception или another claim |
| phase claimed, Start DP-013 cancelled до Owner claim | parent и phase terminalize без нового attempt | fabricate Starting или Running |
| cancellation pending Stop claimant до delegation | Stop terminal no-mutation; continuation может перейти к binding | transfer или invoke его permit |
| return caller pending Stop или permit loss | linked command set остаётся unresolved до Load | second permit или preparation work |
| Owner claim, coherent exact binding absence | final gate может вернуть BindingFailed; Flow converge exact token через Owner.Start | continuation публикует lifecycle/command facts или начинает Load |
| другое generation, stale/conflicting/inactive, unavailable или unknown binding facts | exact re-read; converge к exact terminal outcome или Blocked | BindingFailed или resource-free inference |
| Owner claim, binding/inspection indeterminate | linked command set остаётся unresolved до Load | preparation work, new generation или blind retry |
| BindingFailed, Owner failure acceptance wins | Owner возвращает preparation failure; затем persist Failed и command/phase outcome | publish до Owner outcome |
| BindingFailed, later Stop wins mutex Owner | Owner возвращает stopped-before-running; persist exact outcome | overwrite binding failure |
| binding confirmed, final Stop gate wins | original Stop claimant converge exact attempt до Load | вернуть Continue или transfer Stop permit |
| binding confirmed, final release gate wins | Flow может начать Load; later Stop использует ordinary tracked route | bypass exact attempt или second permit |
| Owner/durable terminal publication indeterminate | linked command set остаётся unresolved до Load | terminal command или preparation work |
| new startup failure | new attempt становится historical failure после release resources | automatic rollback или Running |
| Running/terminal publication indeterminate | exact command остаётся unresolved | another replacement или fabricated success |
| command terminal publication indeterminate | inspect exact command и aggregate facts | re-execute с другим key |

Definitive failure после old release может оставить Runtime Instance Stopped.
Это truthful supported failure outcome, а не permission revive old execution.

## 20. Caller cancellation

Cancellation, видимая до command claim, выполняет zero mutation. Между parent
claim и linked old-Stop phase claim она может terminalize parent без lifecycle
mutation, если existing gate первым подтверждает cancellation.

После победы linked old-Stop phase claim authoritative является convergence
Stop DP-010. Caller cancellation может interrupt только waiter; без definitive
outcome parent и phase становятся unresolved и не переходят к new attempt.
После proven release cancellation снова проверяется до Continue gate; если она
выигрывает, parent terminalizes, а Runtime Instance остаётся truthfully Stopped.

После победы Caller Cancellation Gate нового Start authoritative является
synchronous wait DP-011, и caller cancellation больше не сокращает operation.
Если cancellation выигрывает до Owner claim при pending Stop, Start и parent
terminalize без attempt, а Stop terminalizes satisfied; его permit не вызывает
DP-013.

Cancellation caller pending Stop до signal Owner claim публикуется original
claimant как terminal no-mutation и позволяет continuation перейти к binding и
final gate.
После signal действует обычное cancellation ordering DP-010. Caller return,
permit loss, unproven convergence или indeterminate result сигнализирует
`Blocked`; Flow не начинает Load, а linked command set остаётся unresolved.
Если linked Start definitive возвращается до claim Owner, его path сигнализирует
`StartNoClaim`; original Stop claimant consume свой permit как terminal satisfied
без DP-013. Lost/indeterminate signaling является `Blocked`.

## 21. Concurrency и command admission

Same-key replay и different-intent conflict DP-015 сохраняются. Одна
per-Instance admission boundary сериализует replacement/rollback commands и
Continue gate. Tracked parent может продвигать только следующую immutable linked
phase. Он может сосуществовать только с одним exact Stop exception section 13.
В section 15 Stop либо выигрывает до linked Start phase, либо после победы phase
ровно один Stop claim exception tracked-Start. Pre-claim Stop остаётся pending
до observation claim Owner private continuation; post-continuation Stop сразу
делегируется тому же Owner. Unresolved parent, phase или pending Stop не
разрешает exception.

Разные Runtime Instances выполняются независимо. Commands не получают
cross-Instance global lock. Concurrent publication новой ConfigurationVersion
не retarget already claimed command или Launch Attempt.

## 22. Indeterminate и recovery boundary

Любая потеря exact parent/phase permit, termination Control Service или
indeterminate publication parent/phase/attempt оставляет linked command set
unresolved. New activation, replacement или rollback не проходят admission
этого Runtime Instance, пока coherent inspection и утверждённый recovery
contract section 19(5) не разрешат parent, каждую phase и truth live resources.
DP-017 имеет status Approved и определяет fail-closed reconciliation exact
facts без lifecycle replay; его Planned implementation остаётся отсутствующей.

Design не hydrate Owner, не inspect process, не probe socket, не reconcile
persisted Running и не выбирает, завершилась ли orphan operation. Это mandatory
recovery decisions, а не hidden ordering mechanics.

## 23. Security и redaction

Validation, observation, command claim, replay и каждое continuation происходят
в exact authorized scope Workspace, Configuration, Runtime Instance, action и
target version. Retry авторизуется повторно; earlier authorization result не
является durable authority.

Outcomes раскрывают только opaque identities и redacted semantic categories.
Они не содержат credential, Secret, raw Configuration payload, Snapshot,
internal error, stack trace, Host pointer, process-local permit или cross-scope
state. Concrete reporting/redaction остаются section 19(6).

## 24. Technology Neutrality

Contract не требует transaction product, database, workflow engine, queue,
distributed lock, saga framework, clock или identifier format. Термины claim,
gate, revision и publication описывают observable semantic ordering.

Generic deployment engines, universal command buses, dynamic registries и
service locators не требуются и не разрешаются. Private implementation
mechanics должны доказать contract без его расширения.

## 25. Acceptance proofs

Будущая implementation должна доказать минимум:

1. initial activation создаёт один exact version-pinned attempt;
2. exact Running target возвращает satisfied с zero mutation;
3. different Running target не меняется in place;
4. replacement никогда не владеет old и new Hosts одновременно;
5. Stop во время old Starting захватывает тот же attempt;
6. new claim происходит только после old proven release;
7. Continue gate имеет одного winner между linked Start phase и concurrent Stop;
8. Stop, выигравший до new claim, предотвращает любой новый attempt;
9. Start phase, выигравшая первой, разрешает ровно один pending Stop, который
   вызывает DP-013 из original claiming path только после claim Owner и до Load;
10. Stop после continuation достигает exact tracked attempt и предотвращает или
    завершает Running по rules Owner;
11. Stop failure или unproven cleanup запрещает new claim;
12. startup failure не resurrect old Host и не запускает automatic rollback;
13. explicit rollback использует exact Published target и fresh attempt identity;
14. same-target rollback/activation является zero-mutation satisfied;
15. cancellation на каждой phase сохраняет truthful state и ownership;
16. indeterminate outcomes закрывают admission DP-015 до recovery;
17. exact DP-014-owned execution-generation binding, required и consumed
    DP-017, commit после attempt claim и до Load, иначе preparation не
    начинается;
18. разные Instances выполняются независимо;
19. EN/RU contract, failure matrix, gates и planned status aligned.

Proofs включают доступные concurrency, race, failure injection и
storage-client-restart scenarios. Они не разрешают production activation.

## 26. Formal и последующие gates ARCH-004 section 19

Этот Approved design закрывает focused architecture design gate ARCH-004
section 19(4). Approved DP-014, DP-015, DP-017 и DP-018 закрывают остальные
focused design gates sections 19(2), 19(3), 19(5) и 19(6). Полный approved set
определяет ordering. Isolated process-local stores DP-014/DP-015 существуют,
но activation orchestrator, external durable workflow persistence, recovery,
reporting, integration и Production Activation отсутствуют.

## 27. Явно отложено

До focused designs или implementation tasks отложены:

- public Activate, Replace, Rollback, Restart или deployment API;
- storage schema, workflow persistence, adapters и migrations;
- process-restart recovery, hydration, orphan resolution и reconciliation;
- diagnostics, audit, metrics, alerting и redaction policy;
- zero-downtime replacement, Listener transfer и connection draining policy;
- automatic rollback/restart, retry/backoff, scheduling и supervision;
- safe command retention, deletion и Production Activation.

## 28. Implementation boundary

Implementation Status — Planned. Repository содержит isolated Lifecycle Owner,
launch flow, source adapter, routing Draft DP-013, aggregate storage Approved
DP-014 и command storage Approved DP-015, включая isolated parent/phase
Continue/pending-Stop rendezvous. DP-016–DP-018 остаются Planned.
Activation/replacement orchestrator, external durable
command/aggregate/workflow storage, public management API, recovery executor и
production wiring отсутствуют.

Approval закрывает design gate section 19(4), но не реализует и не подключает
contract. TASK-026 остаётся Blocked до implementation и acceptance всех
оставшихся prerequisites DP-019; reduced slice DP-016 запрещён.

## 29. Решение

UWP будет активировать один Runtime Instance только из exact authorized
Published ConfigurationVersion. Replacement и explicit rollback сохраняют
identity Runtime Instance, но всегда создают fresh Launch Attempt после proven
release prior Host. Overlapping Hosts, in-place reload, inferred previous/latest
selection и automatic rollback запрещены.

Continue gate атомарно упорядочивает concurrent Stop и claim linked Start phase.
Stop либо предотвращает эту phase, либо занимает один exception tracked-Start;
private continuation сигнализирует его original claimant только после claim
Owner и до Load. Только этот path использует Stop permit, а eventual exact
attempt claim только Owner. Failures остаются truthful и могут оставить
Instance Stopped или Failed; indeterminate outcomes закрывают command admission
до separate recovery.

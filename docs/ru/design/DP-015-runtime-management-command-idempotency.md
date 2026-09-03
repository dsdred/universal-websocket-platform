# DP-015: Runtime Management Command Idempotency

[English version](../../en/design/DP-015-runtime-management-command-idempotency.md)

## 1. Статус

- **Design Status:** Approved
- **Implementation Status:** Primitive boundary Start/Stop, parent/phase
  sequential core Approved DP-019 и command-boundary Continue/pending-Stop
  rendezvous реализованы изолированно; managed gates и continuation Среза 3
  реализованы и независимо приняты изолированно; replay-first/late-generation
  admission TASK-057 реализован изолированно под verification; полное
  extension DP-019 остаётся Planned

TASK-049 завершила design-only refinement contract replay-first orchestration
admission и позднего выделения generation в разделе 13.2; Coordinator
Acceptance получена 2026-08-28. TASK-057 реализует этот отдельный contract
изолированно и остаётся projected `In Progress` под post-sync verification.
DP-015 остаётся Approved с Partial implementation.

Этот approved design определяет durable idempotency boundary для
state-changing management commands Runtime. Package
`internal/runtimecommandidempotency` реализует boundary изолированно на
process-local in-memory storage без management integration, external schema,
API, recovery worker или production wiring.

## 2. Назначение

Определить, как повторные, concurrent и неоднозначно завершившиеся submissions
одного авторизованного state-changing management intent Runtime сходятся на
одном durable command execution и одном replayable outcome без создания второй
lifecycle mutation или Launch Attempt.

## 3. Authority

Этот proposal уточняет, но не переопределяет:

- [ARCH-004](../architecture/ARCH-004-runtime-deployment-and-identity-model.md),
  особенно section 19(3);
- [DP-013](DP-013-runtime-management-routing.md) для exact routing и
  authorization-before-mutation;
- [DP-014](DP-014-runtime-operational-identity-persistence.md) для durable
  aggregate identity, conditional revision и publication lifecycle facts.
- [DP-019](DP-019-runtime-activation-orchestration-prerequisites.md) для exact
  planned callback-scoped parent/phase API и authorization integration.

Принятые ADR и Active или Frozen architecture остаются authoritative. DP-013
остаётся Draft; Approved DP-014, primitive boundary DP-015 и partial
parent/phase sequential core DP-019 и command-boundary Continue/pending-Stop
rendezvous реализованы изолированно. TASK-031/TASK-032/TASK-035 также реализуют
изолированно exact authorization value, dependency-leaf binding, primitive
managed adapter и seam managed Flow/OwnerClaimView. TASK-037 реализует
изолированно managed parent/StartTarget adapter, общие managed gates, concrete
continuation, binding sequence attempt/generation DP-014 и exact адаптацию
outcomes managed Flow, реализованные и независимо принятые изолированно.
TASK-043 реализует и независимо верифицирует concrete private composition
invoker изолированно. Последующая terminal publication, orchestrator и
Approved DP-016 сохраняют Implementation Status Planned.
TASK-046 фиксирует additive contract admission tracked-Start managed-parent, а
TASK-047 реализует его изолированно через
`Boundary.ExecuteManagedParentFromTrackedStart` и callback-scoped capability
preclaimed `StopOld`. Fresh reassessment TASK-026 принимает `READY — UNBLOCK
TASK-026` с 7 Direct / 10 Compositional / 2 Missing core / 0 Missing
prerequisite / 0 Missing external / 0 Deferred. Текущий цикл TASK-026
superseded эту readiness для live execution: repeat Architecture Confirmation
вернула `NEEDS DECISION` / `SPLIT REQUIRED`, потому что тогдашние eager
generation и combined inspect/claim не обеспечивали exact replay-first admission и late
allocation. TASK-026 заблокирована. TASK-049 — завершённая и Coordinator-
Accepted design-only DP-015/DP-020 refinement; TASK-057 реализует её isolated
prerequisite под verification. Статус DP-015 не меняется, а TASK-026 требует
отдельной post-acceptance readiness reassessment.

## 4. Область

Design охватывает:

- opaque client-supplied command identity;
- immutable binding command intent;
- ordering validation, authorization, durable claim и lifecycle delegation;
- same-key replay и different-intent conflict;
- concurrent submission, in-progress observation и terminal replay;
- observation lifecycle outcome и barrier unresolved command;
- definitive и indeterminate outcomes;
- ограничения retention, isolation и redaction.

Design не определяет HTTP headers, DTO, status codes, SDK behavior, database
schema, storage technology, activation, replacement, rollback, recovery,
reconciliation или operational reporting.

## 5. Термины

**Command key** — opaque значение, предоставленное caller для identity одного
state-changing management intent в одном exact command scope.

**Command scope** — tuple operational management domain, Workspace,
Configuration, Runtime Instance и kind management operation.

**Immutable intent** — полный normalized semantic input, способный повлиять на
авторизованную lifecycle mutation. Он исключает transport representation,
credentials, request context, tracing metadata и mutable observation data.

**Command record** — durable idempotency fact, связывающий один command key и
scope с одним immutable intent, command state, observed lifecycle facts, когда
они известны, и replayable outcome при terminal state.

**Execution permit** — одна non-transferable process-local capability,
возвращаемая только path, который commit новый claim. Она разрешает один exact
invocation DP-013. Это не durable identity, lifecycle ownership или proof
начала mutation.

Focused downstream orchestration design может определить один **parent
orchestration permit** и durable linked **phase claims**. Parent permit
разрешает только конечные phase-claim transitions, объявленные этим design; он
не разрешает invocation DP-013. Каждый newly committed phase claim возвращает
собственный non-transferable phase permit не более чем для одного exact
invocation DP-013. Phase identity выводится из parent command identity и
immutable kind/ordinal phase, не выбирается caller и не может replay как другая
phase.

**Replay-equivalent outcome** — стабильный semantic result, необходимый для
ответа на same-intent repeat без повторной delegation lifecycle work. Это не
сохранённый HTTP response и не raw internal error.

## 6. Граница ответственности

Command idempotency boundary владеет только command claim, equality intent,
command state, barrier unresolved command, observed outcome facts и replay
facts. Он не владеет lifecycle решениями Runtime, live Host resources,
authorization policy, transport mapping или recovery.

Runtime Lifecycle Owner остаётся единственным lifecycle decision maker и
owner live Host. DP-014 остаётся owner durable facts Runtime Instance
и Launch Attempt. Idempotency boundary не может выводить текущую liveness,
выбирать version, повторять lifecycle work или становиться service locator.

## 7. Охватываемые commands

Этот contract применяется только к state-changing management operations,
semantics которых определены authoritative design. В текущем isolated surface
DP-013 это Start и Stop. Observe является read-only и не создаёт durable
command record.

Contract не вводит Create, Delete, Restart, Replace, Rollback или другую
operation. Future operation должна сначала определить lifecycle semantics и
immutable intent, прежде чем использовать эту boundary.

## 8. Command identity и namespace

Command identity является парой:

```text
(CommandScope, CommandKey)
```

Representation и allocation CommandKey являются implementation decisions. Key
никогда не является RuntimeInstanceID, LaunchAttemptID, aggregate revision,
ConfigurationVersionID, Principal, credential, timestamp, PID, pointer или
transport request identity.

Один raw key может существовать в другом command scope без collision. В одном
scope committed key никогда не перепривязывается к другому intent. Cross-scope
lookup или replay запрещены.

## 9. Immutable intent

Equality intent является semantic и exact. Intent включает verified Target и
каждый operation input, способный изменить lifecycle mutation:

- Start включает exact identity Published ConfigurationVersion, запрошенную
  command;
- Stop включает exact Target без inferred version или latest attempt;
- future commands включают только inputs, определённые их approved lifecycle
  contract.

Intent исключает Principal, credential, authorization result, deadline,
cancellation state, trace ID, transport encoding, retry count и aggregate
observations после submission. Mechanics canonicalization или fingerprint
должны обеспечивать deterministic collision-safe equality, но design не
требует hashing или wire format.

## 10. Ordering validation и authorization

Каждый submission следует этому порядку:

1. validate command key, exact Target, operation и operation inputs;
2. resolve exact Runtime Instance scope без mutation;
3. authorize текущего caller для exact action и Target;
4. выполнить final caller-cancellation gate, требуемый command contract;
5. inspect или claim durable command record;
6. только newly committed claim может перейти к lifecycle delegation.

Invalid, missing, mismatched, canceled-before-claim, denied или failed
authorization выполняет zero command и lifecycle mutation. Authorization
выполняется для каждого submission, включая replay. Его result не сохраняется
как durable authority и не используется повторно для другого caller.

## 11. Durable claim

Новый command claim атомарно публикует:

- exact command identity;
- immutable binding intent;
- state `Claimed`;
- отсутствие terminal outcome;
- отсутствие fabricated lifecycle mutation;
- monotonic command revision или equivalent conditional token.

Complete claim commit происходит до вызова любого lifecycle method Flow или
Owner. Definitive claim failure ничего не делегирует. Allocation candidate key
или private representation intent до commit не доказывает существование
command.

## 12. Матрица решений same-key

После validation и authorization inspection одной command identity имеет ровно
такие semantic outcomes:

| Existing record | Submitted intent | Result | Lifecycle delegation |
| --- | --- | --- | --- |
| absent | valid intent | atomic claim нового command | разрешена один раз после confirmed claim |
| non-terminal | same intent | truthful non-terminal observation | запрещена |
| terminal | same intent | replay-equivalent terminal outcome | запрещена |
| любой | different intent | conflict command key с zero mutation | запрещена |

Same-intent repeat никогда не обновляет authority, не меняет intent, не
продвигает lifecycle phase, не создаёт Launch Attempt и не ожидает completion
неявно. Waiting или polling behavior является future API concern.

## 13. Concurrency и serialization

Concurrent submissions одной command identity используют одну conditional
serialization boundary. Не более одного submission создаёт claim. Каждый
остальной same-intent submission наблюдает этот claim или более поздний state;
каждый different-intent submission получает conflict.

Commands разных Runtime Instances могут продвигаться независимо. Для одного
Runtime Instance evaluation каждого non-terminal record, permitted exception
tracked Start и insertion нового claim имеют одну atomic linearization point
command admission. Concurrent different keys не могут оба пройти stale barrier
check.

Claimed record с exact live execution permit является **tracked** в текущем
process. Пока tracked Start выполняется, ровно один distinct Stop command может
claim собственный permit и делегироваться в тот же scope DP-013. Это требуемая
ARCH-004 convergence Stop-during-Starting; Owner захватывает тот же Launch
Attempt и остаётся authoritative. Другой Start или другой Stop после появления
tracked Stop получает non-mutating in-progress conflict.

Focused contract DP-016 уточняет то же single exception для tracked parent
replacement или rollback; он не создаёт общий orchestration bypass barrier.
Если replacement/rollback принимается во время tracked earlier Start, parent
claim и его первый linked Stop phase claim атомарно занимают один Stop
exception. Independent Stop не может занять его одновременно. После release
old attempt tracked parent может claim только свой declared linked Start phase.
На Continue gate DP-016 один independent Stop упорядочен против этого
Start-phase claim: Stop либо terminalizes parent до появления phase, либо phase
выигрывает. После победы phase ровно один distinct Stop может claim exception
tracked-Start. До claim Owner он записывается pending, а его permit не может
invoke Stop DP-013; continuation Start-claim DP-011/DP-013 сначала сигнализирует
Owner claim original claiming path Stop. Тот же blocked call stack сохраняет
non-transferable permit, проверяет собственную cancellation, выполняет свой один
invocation Stop DP-013, публикует outcome и сигнализирует continuation.
Continuation никогда не получает и не вызывает этот permit.

Если pending Stop не остаётся, continuation координирует exact
execution-generation binding DP-014. Та же per-Instance admission boundary затем
атомарно упорядочивает final Stop claim против `Continue` для confirmed binding
или `BindingFailed` только для coherently proven absence у exact still-active
attempt/expected revision. Different generation, stale/conflicting/inactive
facts, unavailable state или unknown перечитывается и converge к exact terminal
outcome либо остаётся `Blocked`; он никогда не получает permit или BindingFailed.
Выигравший Stop converge original
claimant; выигравший `Continue` разрешает preparation, после чего later Stop
может claim ordinary tracked-Start exception. `BindingFailed` не terminalize
command напрямую: Flow
converge exact token через Owner.Start с `FailedPreparation`, а command/phase
terminalizes только из exact Owner outcome после confirmed DP-014 terminal
publication. Stop, выигравший mutex Owner, вместо этого даёт
stopped-before-running outcome. Indeterminate binding, Owner convergence или
terminal publication даёт `Blocked` и unresolved. Другой Stop/lifecycle command
получает non-mutating in-progress conflict.

### 13.1 Admission managed-parent поверх tracked-Start

Одна dedicated internal operation может admit новый replacement или rollback
parent поверх ровно одного blocking primitive Start, только когда этот Start
находится в `Claimed`, имеет exact live permit и revision текущей generation в
том же operational domain, Workspace, Configuration и Runtime Instance, а его
sole Stop exception не занят. Validation, current authorization, final
pre-claim cancellation gate и inspection того же parent предшествуют новому
claim. Same intent наблюдает InProgress или Replay без callback и новой
authority; different intent возвращает key conflict.

Под active-generation read lock и одним per-Instance ledger lock единый atomic
transition commits parent как Claimed, его derived ordinal-zero phase
`StopOld` как Claimed с одним private live phase permit, occupation sole Stop
exception tracked Start этой phase и rendezvous parent. Эти internal mutations
записей ledger/storage DP-015 являются обязательным atomic transition и
выполняются под этими locks. Callback, lifecycle invocation, wait, external
storage callback/I/O и иная external work под ними не выполняются. Occupant
внутренне является либо одним primitive Stop
command, либо одной derived phase `StopOld`; новый path не создаёт primitive
Stop record или identity.

Только call нового claim получает callback-scoped, generation-bound capability
managed-parent. Она сохраняет existing managed StartTarget и parent-terminal
operations и добавляет одну operation, consume уже выданный permit `StopOld`.
Она не раскрывает parent или phase permit. Ровно один callback-scoped consumer
может synchronously и не более одного раза invoke exact Stop; repeated
consumers и replay только наблюдают. StartTarget остаётся illegal до durable
Terminal preclaimed phase.

Independent primitive Stop и этот parent admission разделяют одну ledger-lock
linearization point. Stop first не создаёт parent или phase. Parent first
атомарно раскрывает обе records и заставляет independent Stop завершиться без
mutation. Invalid, unauthorized, pre-claim-cancelled или stale submissions не
изменяют state. Callback error, invalid outcome, panic, `runtime.Goexit`,
return без consumption/publication, generation loss, cancellation после claim
или indeterminate publication не фабрикуют success, не передают authority и не
выдают permit повторно; durable non-terminal facts остаются fail-closed и
unresolved. Non-returning callback удерживает только private live capability,
но не admission lock. Existing ordinary admission paths не меняются, а разные
Runtime Instances остаются независимыми.

### 13.2 Replay-first admission оркестрации и позднее выделение generation

Additive orchestration admission сохраняет эту границу и разделяет наблюдение
и claim. Каждая submission валидируется, авторизуется и проходит финальную
cancellation-проверку до входа в существующую per-Instance точку
линеаризации. Сначала инспектируется exact command identity: same-intent
non-terminal возвращает `InProgress`, terminal — exact replay, different intent
— conflict. Для существующей записи не вызываются absent-intent decision,
generation provider или lifecycle callback.

Только absent identity может вызвать принадлежащее оркестратору read-only
решение вне command, aggregate и Owner locks. Closed decision возвращает только
`SatisfiedCandidate` с exact revision/attempt/version facts,
`ExecutePrimitiveCandidate` с exact expected aggregate revision,
`ExecuteParentCandidate` (включая tracked-Starting parent и preclaimed
`StopOld`) либо definitive `NoClaim`. Он не содержит generation, permit,
rendezvous, lifecycle authority или непроверенную terminal truth. Claiming path
возвращается в ту же admission boundary и атомарно повторно проверяет identity
и candidate; проигравший race только наблюдает победителя.

`SatisfiedCandidate` сначала создаёт durable command claim, затем повторно
проверяет exact aggregate facts до terminal publication. Stale, unavailable или
ambiguous revalidation оставляет запись unresolved и не фабрикует satisfied
truth. Для execution candidate composition-owned generation provider вызывается
ровно один раз на winning synchronous call stack, только после claim primitive
или `StartTarget` и победы final cancellation/admission gate. Результат обязан
быть непустым; DP-015 устанавливает immutable binding и rendezvous до вызова
managed Flow. Replay, losing race, satisfied outcome и pre-claim failure provider
не вызывают.

Ошибка provider или пустой результат, panic, `runtime.Goexit`, non-return,
замена generation, cancellation после winning gate, неопределённость установки
binding/rendezvous и любая indeterminate publication оставляют durable command
или phase в `Claimed` и unresolved. Capability исчерпывается, не retry-ится и
не выдаётся повторно; Owner, Load, Build, Launcher и Host не вызываются.
Обычная legacy admission не меняется, разные Runtime Instances остаются
независимыми.

Pending claimant ждёт ровно один process-local signal: `OwnerClaimed` или
`StartNoClaim`, когда linked Start path definitive возвращается до claim Owner.
`StartNoClaim` позволяет тому же Stop path consume permit как terminal satisfied
без Stop DP-013. Потеря signal является unresolved, а не implicit choice.

Claimed primitive, parent или phase record без exact live permit является
**unresolved**. Unresolved parent или любая его phase является durable barrier
для каждого нового state-changing command и дальнейшей phase до тех пор, пока
утверждённый recovery contract не сделает linked command set Terminal. Approved
[DP-017](DP-017-runtime-recovery-reconciliation.md) определяет fail-closed
resolution exact facts; его Planned implementation остаётся отсутствующей.
Observe остаётся read-only. Ни
один tracked exception не действует после restart process, потери claiming
call stack или indeterminate claim/terminal publication.

## 14. Lifecycle delegation

Только execution path, подтвердивший создание нового durable claim, может
делегировать command exact scope DP-013. Он делегирует не более одного раза в
этом process execution и сохраняет существующие semantics cancellation,
outcome и failure Flow и Owner.

Parent orchestration path никогда не делегирует прямо в DP-013. Он может
продвигаться только conditional commit следующей immutable linked phase,
объявленной focused orchestration contract. Только path с newly issued permit
этой phase может выполнить её один exact invocation DP-013. Outcomes parent и
phase продвигаются монотонно; missing или indeterminate phase outcome закрывает
barrier parent, а не разрешает другую phase или permit.

Pending Stop claiming path остаётся synchronously blocked и сохраняет свой
permit. Signal Owner claim только открывает downstream gate этого path и не
передаёт authority. Только тот же path вызывает Stop DP-013, затем публикует и
сигнализирует один definitive no-mutation, converged, failed или indeterminate
outcome. Return без такого outcome теряет live permit и делает pending Stop и
parent unresolved.

Claimed command не доказывает, что lifecycle mutation началась. Caller
cancellation после claim не удаляет record и не позволяет другому request
делегировать его снова. Существующий gate Flow или Owner решает, начинается ли
mutation; command outcome должен правдиво различать отсутствие mutation и
начатую или завершённую mutation.

## 15. Lifecycle outcome и barrier unresolved command

Design не меняет exact surface Start или Stop DP-013 и не передаёт command
identity в Flow, Owner или aggregate publication DP-014. После durable command
claim только path с execution permit один раз вызывает exact operation DP-013.
Same-key replay никогда не получает этот permit.

Если этот synchronous invocation возвращает definitive outcome, command
boundary может опубликовать его replay-equivalent Terminal outcome. Любые
Launch Attempt, version, Stop origin или aggregate fact из этого outcome
сохраняются только как observed immutable facts; boundary не создаёт identity,
которую существующий outcome не раскрывает.

Promise atomic commit между текущим call DP-013 и command record намеренно
отсутствует. Если lifecycle mutation могла произойти, но terminal command
publication отсутствует или indeterminate, record остаётся Claimed. После
исчезновения execution permit он является unresolved и закрывает per-Instance
barrier. Ни retry, ни другой key не могут делегировать lifecycle work. Approved
DP-017 определяет contract section 19(5) для inspection exact command,
lifecycle и execution-evidence facts и truthful barrier resolution; его
Planned implementation остаётся отсутствующей.

## 16. Состояния command

Минимальные semantic states:

```text
Claimed -> Terminal
```

`Claimed` означает durable command ownership, но replay-equivalent terminal
outcome ещё не является durable. Matching live execution permit отличает
tracked execution от unresolved Claim без изменения durable identity.
Lifecycle mutation может отсутствовать, выполняться, быть завершённой или
indeterminate; Claim сам не выбирает одно из этих состояний. `Terminal`
означает durable replay-equivalent outcome, после которого per-Instance barrier
может открыться для следующего command.

Implementation может использовать private substates, но не может выполнять
regression, пропускать обязательную truth или представлять Claimed как
successful terminal completion.

## 17. Terminal outcome

Terminal publication conditionally проверяет command identity, immutable
intent, current command state, command revision и definitive operation outcome.
Затем она сохраняет один immutable replay-equivalent outcome и один раз
продвигает command state в Terminal.

Outcome фиксирует stable domain categories и identities, нужные для semantic
replay. Он не сохраняет credentials, Principal, raw internal error, stack
trace, Host pointer, context, transport response или mutable live observation.
Replay terminal outcome выполняет zero lifecycle и aggregate mutation.

## 18. Definitive failures

Failures validation, lookup, authorization или pre-claim cancellation не
создают command record. Definitive claim failure ничего не создаёт и не
делегирует.

После существования claim definitive no-mutation lifecycle rejection может
быть опубликован как terminal command outcome. Definitive committed lifecycle
outcome должен быть linked, затем опубликован terminal. Wording failure или
transport mapping могут меняться; stored semantic category и identity facts
должны оставаться replay-equivalent.

## 19. Indeterminate outcomes

Если claim, lifecycle invocation, command observation или terminal publication
имеет indeterminate outcome, caller не должен:

- создавать replacement command key для того же intent;
- повторно делегировать lifecycle work;
- создавать другой Launch Attempt;
- предполагать отсутствие, failure или terminal command;
- создавать replay result из stale in-memory state.

Caller инспектирует exact command identity и доступные exact Runtime Instance,
revision, Observation и facts Launch Attempt. Coherent read может установить
absent, tracked Claimed, unresolved Claimed, Terminal или still unknown в
текущем process. Durable state сам по себе никогда не создаёт live permit.
Unresolved record блокирует каждый новый state-changing command этого Runtime
Instance. Restart-time resolution и convergence orphan commands относятся к
ARCH-004 section 19(5), а не к этому design.

## 20. Caller cancellation и retry

Cancellation, видимая до durable claim, выполняет zero mutation. Между claim и
downstream gate Flow или Owner видимая cancellation ещё может выиграть этот
существующий gate и дать definitive result без lifecycle mutation.

Для Start после победы Caller Cancellation Gate DP-011 тот же synchronous
invocation Flow больше не проверяет caller context и ожидает один exact Owner
outcome или operation error. Idempotency boundary не может вернуть управление
раньше или detach эту work. Для Stop locked cancellation gate DP-010 решает,
начинается ли mutation; после победы nil check и locked mutation поздняя
cancellation может прервать только wait этого caller, пока convergence Owner
продолжается. Если definitive terminal outcome недоступен и caller Stop
возвращается, его permit исчезает; command остаётся unresolved Claimed и
сохраняет закрытым per-Instance barrier. Пока Start permit остаётся live,
отдельно claimed exception Stop остаётся доступным и достигает того же Owner.

Для pending Stop DP-016 cancellation до signal Owner claim проверяется и
terminally публикуется original claiming path. Если она definitive выигрывает
до delegation, permit consumes без mutation DP-013, а Start continuation может
перейти к execution binding и final gate. Cancellation, видимая только после
signal Owner claim,
управляется обычным gate Stop DP-010. Если pending caller возвращается, теряет
permit, не может доказать no mutation или опубликовать definitive outcome, он
сигнализирует `Blocked`; Flow не начинает Load, а linked set остаётся unresolved.
Admission/Owner lock не удерживается, пока любой call stack ждёт signal/result.

Cancellation никогда не удаляет command, не передаёт command ownership retry и
не разрешает duplicate delegation.

Retry обязан использовать ту же command identity и immutable intent. Новый key
является новым command, а не retry, и остаётся subject текущим lifecycle
preconditions. SDK retry counts, backoff, deadlines, polling и transport status
находятся вне design.

## 21. Retention и reuse key

Safe forgetting зависит от retry horizons caller, indeterminate outcomes,
audit requirements и Approved recovery semantics, Planned implementation
которых остаётся отсутствующей. Поэтому в рамках этого Approved contract
command record нельзя удалять, а command identity нельзя использовать
повторно.

Будущий focused retention contract может разрешить bounded expiry только если
докажет, что expired key не приведёт к повторному выполнению старого intent
или смешению старого terminal result с новым command. Один time-to-live не
является таким proof.

## 22. Security и redaction

Каждый inspection, claim, conflict, in-progress observation и replay происходит
только после current authorization для exact Target и action. Responses не
раскрывают существование того же raw key в другом scope или для unauthorized
target.

Durable command records содержат opaque identity, normalized intent facts,
state, bounded observed lifecycle facts и redacted semantic outcomes. Они не содержат credential,
Secret, authentication material, raw Configuration payload, Snapshot, raw
error, stack trace, Host reference или process-local capability. Concrete
reporting и redaction policy остаются обязательными в section 19(6).

## 23. Technology Neutrality

Encoding command key, comparison intent, representation revision, storage,
locking, transaction mechanics и serialization являются implementation
choices. Observable requirements: durable claim-before-delegation, exact
binding intent, один accepted claim, atomic per-Instance command admission,
один non-transferable permit на accepted primitive или phase claim, отсутствие
lifecycle delegation у parent permit, bounded phase/pending-Stop exceptions
DP-016, Start-claim continuation до external preparation work, barrier
unresolved command set, at-most-once phase delegation, monotonic command state,
truthful inspection и replay без mutation.

Generic CRUD repositories, distributed lock services, universal command buses,
dynamic registries и service locators не требуются и не разрешаются.

## 24. Acceptance proofs

Implementation должна доказать минимум:

1. один same-key/same-intent claim при concurrent submission;
2. same-key/different-intent conflict с zero mutation;
3. authorization на каждом initial и replay submission;
4. отсутствие command claim и lifecycle mutation до authorization;
5. durable claim до lifecycle delegation;
6. at-most-once delegation для concurrent и repeated submissions;
7. evaluation barrier и different-key claim имеют одну per-Instance atomic
   linearization point;
8. tracked Start разрешает ровно один distinct Stop claim и делегирует его тому
   же Owner, тогда как unresolved Claim блокирует каждую новую mutation;
9. non-terminal replay никогда не сообщает terminal success;
10. terminal replay возвращает тот же semantic outcome с zero mutation;
11. caller cancellation после claim не разрешает duplicate delegation;
12. definitive failures сохраняют specified zero-mutation boundary;
13. indeterminate outcomes требуют exact inspection и запрещают blind retry;
14. разные command identities сохраняют one-Instance lifecycle serialization;
15. restart storage client сохраняет claim и terminal replay facts;
16. domain isolation и redaction предотвращают cross-scope disclosure.
17. admission managed-parent поверх tracked-Start атомарно раскрывает ровно
    один parent и его derived ordinal-zero phase `StopOld` либо не раскрывает
    ни одного;
18. independent Stop и managed-parent admission доказывают оба legal winner
    orders с ровно одним occupant Stop exception;
19. только new claim получает callback-scoped authority preclaimed phase;
    same-key observation и replay не получают её;
20. preclaimed permit вызывает exact Stop не более одного раза, не consume
    через ordinary phase path и expires при возврате callback;
21. panic, `runtime.Goexit`, non-return, cancellation, generation loss и
    indeterminate publication сохраняют truthful fail-closed barriers;
22. ordinary admission behavior не меняется, а разные Runtime Instances
    продолжают выполняться независимо.
23. replay и conflict возвращаются до absent-intent decision и generation
    provider; concurrent absent decisions имеют одного atomic winner;
24. satisfied candidate claim-ится и точно revalidate-ится до terminal
    publication, а stale/ambiguous facts остаются unresolved;
25. generation выделяется один раз после winning claim и cancellation gate и до
    managed Flow; provider/binding failure не вызывает Owner или Load и не
    retry-ится;
26. replay не принимает authority provider/callback, а reconstruction
    восстанавливает durable facts без live capability.

Proofs включают технически доступные concurrency, race, failure-injection,
durability и storage-client-restart scenarios. Они не разрешают Production
Activation.

## 25. Formal и последующие gates ARCH-004 section 19

Этот Approved design закрывает focused architecture design gate ARCH-004
section 19(3). Approved DP-014 и DP-016–DP-018 закрывают остальные focused
design gates sections 19(2) и 19(4)–(6). Полный approved set определяет
dependency ordering. Сами status decisions не создают implementation; current
isolated package предоставляет process-local in-memory command storage, а
external persistence, recovery, reporting, integration и Production Activation
остаются отсутствующими.

## 26. Явно отложено

За пределы isolated command implementation отложены:

- transport idempotency field, DTO, status code и behavior client SDK;
- external durable command schema, migration, storage adapter и production
  integration API;
- activation, replacement, rollback и policy version selection;
- recovery после restart process, resolution orphan commands и reconciliation;
- taxonomy diagnostics, reporting, audit, metrics и redaction policy;
- safe command retention и deletion;
- concrete authorization policy и Production Activation.

## 27. Implementation boundary

Implementation Status primitive Start/Stop — Implemented in isolation.
Parent/phase sequential core DP-019 также Implemented in isolation, а полное
extension остаётся Planned. Package
`internal/runtimecommandidempotency` реализует exact Scope/CommandKey identity,
immutable Start/Stop intent, authorization-before-claim, atomic per-Instance
admission, claim-before-delegation, one-shot process-local execution permit,
tracked-Start Stop exception, unresolved barrier и terminal semantic replay.
Отдельный `MemoryStorage` сохраняет claim/replay facts при reconstruction
`Boundary`, но не обещает persistence через restart процесса и не
восстанавливает live permits. `Boundary.Execute` сохраняет primitive permit
private на synchronous claiming call stack, поэтому caller не может потерять
его между claim и delegation. Client-generation transition атомарно
сериализован с admission; stale Boundary не может создать новый Claim.
`ExecuteParent` добавляет exact Replace/Rollback intent, durable
parent/derived-phase records, generation-bound callback capability, strict
порядок optional `StopOld` затем `StartTarget`, phase replay, parent terminal
gating и тот же unresolved barrier. `ContinueOrExecuteStartTarget` добавляет
non-bypassable pre-phase Continue gate и synchronous pending-Stop rendezvous с
immutable signal cause и fail-closed callback/reconstruction behavior.

TASK-031/TASK-032/TASK-035 добавляют изолированно exact authorization value,
dependency-leaf binding, primitive adapter `ExecuteManagedStart` и seam managed
Flow/OwnerClaimView. TASK-037 добавляет изолированно managed
parent/StartTarget adapter, общие primitive/linked managed rendezvous gates,
stateless continuation OwnerClaim-to-DP-014 и отображение outcomes managed
Flow, реализованные и независимо принятые изолированно. TASK-043 добавляет
concrete private composition invoker изолированно. External durable
storage/schema, API, DP-016 orchestration, DP-017 recovery, последующая
terminal publication DP-014 и terminalization command/phase DP-015, management
wiring и Production Activation отсутствуют. TASK-046 определяет bounded
contract tracked-Start managed-parent плюс preclaimed `StopOld` admission в
section 13.1; TASK-047 реализует изолированно его atomic admission,
discriminated occupant sole Stop exception, callback-scoped consumption,
replay, expiry и proofs winner ordering. Fresh reassessment TASK-026 принимает
READY boundary как historical evidence. Repeat Architecture Confirmation теперь
блокирует TASK-026 отдельным DP-015/DP-020 refinement replay-first admission и
late generation, описанным выше. Design refinement завершена как TASK-049 и
принята Coordinator 2026-08-28; TASK-057 реализует её отдельный replay-first/
late-generation prerequisite изолированно под verification. Isolated package
не изменяет lifecycle contracts и не подключён к DP-013 Directory. TASK-026
остаётся Blocked до отдельной readiness reassessment после TASK-057 Acceptance.

## 28. Решение

UWP будет идентифицировать один state-changing management intent Runtime через
opaque command key внутри exact authorized command scope. Complete durable
claim связывает эту identity с одним immutable intent до lifecycle delegation.
Same-intent repeats наблюдают или replay тот же command без delegation;
different intent с тем же key конфликтует с zero mutation.

Один immutable replay-equivalent terminal outcome сохраняется. Live execution
permit отслеживает claiming path и сохраняет mandatory exception
Stop-during-Start. Claimed record без exact permit является unresolved и
блокирует каждый новый state-changing command того же Runtime Instance.
Cancellation или indeterminate outcome никогда не разрешают blind
re-execution. Runtime Lifecycle Owner остаётся единственным lifecycle decision
maker, а truthful barrier resolution, recovery, retention, API mapping и
production wiring остаются отдельной work.

# DP-022: Граница containment исполнения Runtime и evidence

[English version](../../en/design/DP-022-runtime-execution-containment-and-evidence.md)

## 1. Статус

- **Design Status:** Approved
- **Implementation Status:** Planned

Это предложение определяет execution-containment и evidence boundary, которую
Approved DP-017 section 11 требует до реализации recovery reconciliation. Оно
Approved как эта граница, поэтому prerequisite DP-017 section 11 о
существовании approved containment boundary удовлетворён на уровне дизайна. Это
ничего не говорит о реализации, которая отсутствует: реализация DP-017 остаётся
неактивированной.

[DP-023](DP-023-runtime-process-containment-bootstrap.md) — Approved,
Implemented-in-isolation boundary initial bootstrap capability/ledger/
generation. TASK-069 реализует этот Windows-only substrate в worktree. Latest
verification, review и Acceptance checkpoint и subject identity resolve-ятся
только из newest valid matching envelope TASK-069 и здесь не дублируются. Slice
не реализует evidence outcomes этого предложения.

Evidence adapter, scanner, supervisor, production wiring и composed runtime
behavior отсутствуют. Ничто здесь не утверждает и не подразумевает, что Control
Service уже способен наблюдать termination процесса.

## 2. Назначение

Определить один bounded, technology-neutral ответ на вопрос, который задаёт
Approved DP-017 section 11 и на который до этого документа не отвечал ни один
authoritative source: что делает execution generation уникальной и текущей, и
какая authority может доказать, что exact generation, названный durable
execution binding attempt, уже terminated.

Design должен позволить поздней replacement Control Service различать для одного
exact Runtime Instance между execution, которое всё ещё live, execution, чьё
containing generation доказано terminated, ресурсами, которые отсутствуют,
Host shutdown contract, чьё completion доказано, и истиной, которую она не
знает. Он также должен точно указывать, какие observations недостаточны, чтобы
ни одна реализация не могла подставить probe, clock или stored state вместо
доказательства.

## 3. Полномочия

Это предложение уточняет, не переопределяя:

- [ARCH-004](../architecture/ARCH-004-runtime-deployment-and-identity-model.md)
  sections 9, 11, 12, 14, 17 и 19 — single-node in-process topology, ownership
  Host, publication of `Stopped`, semantics потери process и запрет на process
  metadata как identity;
- [ADR-0003](../adr/0003-runtime-architecture.md) и freeze
  [ARCH-002](../architecture/ARCH-002-runtime-foundation-freeze.md) для
  composition-root ownership и отсутствия Host restart или in-place reload;
- [DP-011](DP-011-runtime-launch-pipeline-integration.md) и
  [DP-013](DP-013-runtime-management-routing.md) для правила, что exact Control
  Service composition создаёт opaque execution generation и что Directory, Flow,
  Owner и Host не allocate его и не persist его binding;
- [DP-014](DP-014-runtime-operational-identity-persistence.md) sections 10 и 13
  для durable attempt-to-generation binding и его явной неспособности доказать
  liveness;
- [DP-016](DP-016-runtime-activation-replacement-rollback.md) sections 5, 10,
  12 и 22 для proven release, phase order и отнесения process inspection к
  recovery;
- [DP-017](DP-017-runtime-recovery-reconciliation.md) sections 5, 6, 10, 11, 12
  и 20 для определений execution generation, execution binding и execution
  evidence, для fail-closed ordering и для authority классифицировать reconciled
  set;
- [DP-019](DP-019-runtime-activation-orchestration-prerequisites.md),
  [DP-020](DP-020-runtime-orchestration-binding-sequence-readiness.md) section
  8.5 и [DP-021](DP-021-private-exact-scope-managed-start-invoker.md) для
  существующего request-exactly-once generation provider seam, binding sequence
  и invocation custody.

DP-022 потребляет те seams и никогда не reallocates, не переопределяет и не
заменяет их. DP-017 section 12 остаётся единственной authority, которая
объединяет facts в recovery classification; DP-022 поставляет facts и их
гарантии. ARCH-004 section 19(5) закрыт Approved DP-017 и этим документом не
заявляется и не переоткрывается. Approved источники сохраняют свой статус; Draft
не может их переопределить.

## 4. Область действия

Design определяет:

- границу containment, покрываемые ею классы ресурсов и гарантию containment,
  которую она может заявить в initial single-node in-process topology;
- containment capability, которая устанавливает одну unique current execution
  generation на containment domain;
- правила выдачи generation, допустимых потребителей и запрет переиспользования;
- durable containment ledger и facts, которые он может и не может содержать;
- точное обязательство доказательства, которое устанавливает termination
  названного prior generation, и то, что такое доказательство никогда не
  устанавливает;
- отдельную authority для Host-owned shutdown completion;
- закрытую модель результатов evidence, её варианты `Unknown` и их приоритет;
- перечень observations, которые не являются evidence, и запрещённые выводы,
  построенные на них;
- уровни гарантий adapter, доверие, scope isolation, security, cancellation и
  границы concurrency;
- явно отложенные вопросы и acceptance proofs.

## 5. Не-цели

Design не определяет:

- recovery claim, permit, assessment, reconciliation, publication order,
  reopening barrier или любое behavior DP-017;
- storage engine, schema, transaction, migration, lock primitive, file layout,
  named object, signal, process table query или format identifier;
- протоколы child-process или remote-worker, supervision, adoption, forced
  termination, обработку live-orphan или cross-node quorum;
- restart, retry, backoff, failover, scheduling, watchdog или self-replacement;
- public или internal API, DTO, transport mapping, health endpoint либо operator
  reporting и redaction, которые остаются за DP-018 и ARCH-004 section 19(6);
- sequencing запуска Control Service, deployment, process management либо
  Production Activation;
- любое новое Runtime lifecycle state, desired state, actual state или
  категорию.

## 6. Термины

**Containment domain** — durable scope, внутри которого решается
execution-generation containment. В initial topology это operational management
domain, обслуживаемый одним Control Service, вместе с durable identity state,
которым он владеет. Domain — это не Runtime Instance, Workspace, Configuration
или метка host-машины.

Для initial process boundary этот durable scope включает один immutable
authoritative storage root и заранее provisioned привязку Domain-to-storage
authority. DP-023 section 8.1 определяет её bootstrap и deployment trust
contract. Отсутствие storage не доказывает, что Domain является новой.

**Containment capability** — исключительно удерживаемая authority быть live
generation одного containment domain. Она acquired одним Control Service process
ровно один раз, удерживается на всё authoritative lifetime этого process и
никогда не передаётся, не renew, не истекает по таймауту и не разделяется.

**Containment boundary** — набор execution-ресурсов, чьё lifetime ограничен
держателем containment capability. Ресурсы вне этого набора находятся вне
гарантии containment.

**Execution generation** сохраняет свой Approved DP-017 section 5 смысл: opaque
identity одной execution-containment boundary, которая в initial topology
обозначает одну Control Service process generation. Это не PID, timestamp,
liveness lease, Runtime identity или command identity.

**Generation authority** — composition-owned responsibility, которая держит
containment capability, выдала current generation identity и отвечает на closed
evidence вопросы о generation собственного domain. Это не lifecycle owner, не
persistence component и не второй источник истины об attempt.

**Containment ledger** — durable, append-only запись о том, какие opaque
generation identity держали containment capability одного domain и какие были
superseded более поздним exclusive acquisition.

**Termination proof** — производный, несущий evidence вывод о том, что exact
названный generation, отличный от current, больше не держит containment
capability и поэтому terminated.

**Shutdown-completion evidence** — coherently прочитанный durable fact,
принадлежащий DP-014 и DP-015, о том, что Runtime Lifecycle Owner завершил
Host-owned shutdown contract для одного exact Launch Attempt внутри одного exact
bound generation.

**Unbound observation** — любой signal, который невозможно соотносить с exact
`(domain, generation, attempt)` tuple. Это не execution evidence.

**Unknown** — семейство результатов, которые ничего не доказывают: absent,
unavailable, stale, scope-mismatched, contradictory, indeterminate, cancelled
либо полученные adapter, чьи объявленные гарантии не покрывают данный случай.

## 7. Модель границы containment

В initial ARCH-004 topology Runtime Host компонуется внутри процесса Control
Service и не может пережить этот процесс. Поэтому граница — это процесс, а
containment является свойством классов ресурсов, а не компонента, который следит
за процессами.

1. **Covered class.** Ресурсы, чьё lifetime платформа завершает вместе с
   процессом: in-process память и goroutine, in-process session state, listeners
   и sockets, открытые этим процессом, и удерживаемые им handles. Для этого
   класса доказанное termination generation является доказательством того, что
   эти ресурсы больше не удерживаются и недостижимы.
2. **Uncovered class.** Всё, что может пережить процесс: дочерние процессы,
   ресурсы, которые другой процесс держит от его имени, или регистрации во
   внешних системах. Initial topology не определяет гарантии containment для
   этого класса и не определяет протокол adoption, termination или cleanup для
   него.

Компонент Runtime, который acquired ресурс из uncovered класса, пока live orphan
path DP-017 не approved, получает evidence, которое граница containment не может
интерпретировать. Поэтому design рассматривает acquisition непокрытых ресурсов
как нарушение границы: evidence для того generation разрешается в `Unknown`,
admission barrier остаётся закрытым, и ситуация эскалируется как architecture
gap, а не ремонтируется выводом. ARCH-004 section 11 запрещает имитировать
supervision через hidden state в Runtime Host, и эта граница не создаёт такого
state.

Гарантия containment исключительно о существовании и достижимости ресурсов. Она
никогда не утверждает, что terminated generation выполнил cleanup gracefully,
завершил stop, закончил command или достиг какого-либо lifecycle outcome.

## 8. Containment capability и одна live generation

1. Один containment domain имеет не более одного live holder containment
   capability. Acquisition является exclusive, atomic и либо succeeds, либо
   fails; не существует shared, partial или «probably held» acquisition, и
   acquisition не повторяется процессом, который уже держит capability или уже её
   потерял.
2. Composition acquired capability один раз, до открытия command admission, до
   принятия management command, до claim любого Launch Attempt и до записи
   любого execution binding. Ordering является normative: до acquisition не
   существует ни одного generation и ничего не может быть связано binding.
3. Holder никогда не release capability, пока продолжает действовать как
   authoritative, и никогда не re-acquires её на месте. Release выполняется
   только платформой как часть termination процесса, и adapter может объявлять
   эту гарантию только если release не может быть потерян, отложен или выдан
   дважды.
4. Capability не является lease. У неё нет expiry, renewal, heartbeat, clock и
   elapsed-time вывода. Никакое wall-clock observation не вносит вклад в любой
   вывод, из неё производный.
5. Failed, ambiguous, lost или revoked acquisition выполняет fail-closed:
   процесс не allocate generation, не persist binding, не открывает admission, не
   invokes lifecycle work и сообщает domain как unavailable. Потеря во время
   жизни закрывает admission и запрещает дальнейшее binding до тех пор, пока
   новый процесс не выполнит новый exclusive acquisition. Затем действует
   recovery assessment по DP-017; данный design не разрешает automatic restart,
   self-replacement или takeover.

Exclusive re-acquisition — единственный механизм, которым данная граница
устанавливает, что prior generation ушёл, и именно поэтому доказательство не
зависит от timing.

## 9. Выдача generation

1. Ровно одно opaque execution generation identity выдаётся на каждый успешный
   acquisition capability, generation authority в composition Control Service, в
   соответствии с DP-011, DP-013 и DP-014 section 10.
2. Identity является non-empty, unguessable, opaque и immutable на всё lifetime
   процесса. Оно никогда не выводится из PID, имени процесса, executable или
   working path, address, port, значения clock, boot identifier, host name,
   версии Configuration, identity Runtime Instance, identity Launch Attempt или
   identity command, не равно им и не composed из них.
3. Identity никогда не переиспользуется внутри domain, пока существует любая
   durable запись о нём. Containment ledger обеспечивает уникальность: выдача
   identity, уже присутствующего в ledger, является дефектом, который
   разрешается в `Unknown` и закрывает admission, а не молча supersede
   generation.
4. Значение, предлагаемое provider seam DP-020 section 8.5, — это ровно current
   generation identity. Provider может быть опрошен не более одного раза на
   winning claim, как уже требует DP-020; он должен вернуть current identity или
   fail. Он никогда не должен возвращать другое generation, remembered
   generation, реконструированное значение или значение, minted после потери.
5. Ни один потребитель не может cache, substitute, merge или replace generation.
   После потери capability процесс не имеет authoritative generation и не выдаёт
   нового без нового acquisition, который данный design запрещает на месте.

### 9.1 Допустимые потребители

Generation authority поставляет identity только тем путям, которые уже владеют
им: binding sequence DP-015/DP-019/DP-020, conditional binding операция DP-014 и
closed evidence reads. Flow, Lifecycle Owner, DP-013 Directory, Runtime Host и
recovery code никогда не allocate его, не persist вне DP-014 и не раскрывают.

## 10. Durable корреляция attempt-to-generation

DP-014 исключительно владеет durable immutable корреляцией одного exact Launch
Attempt с одним execution generation, публикуемую conditionally внутри aggregate
Runtime Instance до Load, как требуют DP-014 section 10 и DP-016 section 10.
DP-022 не добавляет новой durable корреляции, параллельной записи и второго пути
записи. Он лишь фиксирует, что эта корреляция может и не может означать:

1. binding может называть только generation, которое acquiring процесс сейчас
   держит;
2. binding доказывает только корреляцию, ровно как утверждает DP-014, и никогда
   liveness, preparation, readiness, ownership, release или shutdown;
3. durable binding, называющий generation без записи containment-ledger в том же
   domain, является scope mismatch, а не evidence, и разрешается в `Unknown`;
4. binding — это вход, называющий generation, termination которого должен быть
   доказан; сам по себе он никогда не является этим доказательством;
5. ledger и binding должны читаться как один coherent domain-scoped набор; если
   они расходятся, либо если любое чтение stale, partial или contradictory,
   никакое заключение о termination не является допустимым.

## 11. Containment ledger

Ledger — это durable, append-only containment запись одного domain. Он содержит
для каждого generation: его opaque identity, то, что оно держало containment
capability, и то, что более поздний exclusive acquisition его superseded.
Больше он ничего не содержит.

1. Владение: generation authority — единственный writer. Persistence DP-014
   остаётся единственным durable хранилищем facts об aggregate, attempt и
   binding. Recovery является reader и никогда не writer.
2. Порядок: ledger не требует timestamp, sequence clock или сравнения
   elapsed-time. Supersession записывается как явный containment fact во время
   acquisition, поэтому рассуждение о termination является чисто реляционным:
   current holder плюс записанная superseded запись для другого identity.
3. Запрещённое содержимое: lifecycle или desired/actual state, command state,
   outcome, attempt history, PID, address, port, path, credential, payload
   configuration или Snapshot, diagnostic текст или любое поле, которое сделало
   бы ledger конкурирующим источником истины о Runtime.
4. Durability и scope: ledger и attempt state должны принадлежать одному
   containment domain. Ledger, читаемый из более чем одного независимого durable
   state, или domain, чьи два кандидатных хранилища расходятся, дают `Unknown`;
   design не разрешает cross-store неоднозначность предпочтением или
   свежестью.
5. Retention здесь не определяется; удаление или compaction, которые могли бы
   стереть supersession fact generation, всё ещё названного durable binding, не
   разрешены.
6. Ledger — это containment fact set, который DP-017 section 11 требует, чтобы
   назвать exact prior generation. Он не является process supervision, PID
   persistence, scheduling, clustering или самим crash recovery, что ARCH-004
   section 11 оставляет вне области действия, и не хранит ни одного process
   identifier, адреса или пути.

Для initial bootstrap `ProvisionedEmpty` — это существующий, проверенный,
заранее provisioned anchor и ledger без generation entry согласно DP-023
section 8.1. Missing, replaced, relocated, stale или alternate store не
является empty ledger и не даёт authority. Anchor — metadata storage provenance
вне generation records; он не добавляет Runtime facts в ledger. Доверенная
deployment/storage граница предотвращает ненаблюдаемый joint rollback или
cloning anchor и ledger; runtime validation сама по себе не доказывает, что
согласованная копия является исходной.

## 12. Доказательство termination

Для exact prior generation, названного execution binding attempt,
`GenerationTerminated` доказывается тогда и только тогда, когда все следующие
условия выполнены для одного coherent чтения одного domain:

1. читающий процесс сейчас держит containment capability этого domain, acquired
   эксклюзивно в этой process generation;
2. ledger записывает, что названный prior generation держало эту capability;
3. названный prior generation отличается от identity current generation;
4. используемая capability не может держаться двумя live generation и не может
   быть release иначе как через termination процесса, как объявлено уровнем
   гарантий adapter в section 17;
5. generation authority не сообщает о потере или revocation собственного
   acquisition.

Это доказательство устанавливает только то, что ни одно execution, contained этим
generation, не остаётся live или достижимым внутри covered класса ресурсов. Оно
не устанавливает:

- когда, почему и насколько cleanly завершился процесс, ни любую классификацию
  crash;
- что какой-либо Host остановился, закрыл listeners, drain'ил sessions или
  выполнял shutdown contract;
- что Stop, replacement phase или command завершились либо были удовлетворены;
- что ресурсы uncovered класса отсутствуют;
- что attempt может быть terminalized, admitted, replay, adopt либо restarted,
  что остаётся решениями DP-017 и DP-016.

Само-доказательство исключено: current generation никогда не может сообщить о
собственном termination, и ни один процесс не может сообщить termination для
generation, которое он не записал в этом domain. Отсутствие записи в ledger
никогда не является termination proof.

## 13. Shutdown-completion evidence

Host-owned shutdown completion — это другой fact с другим производителем.

1. Оно устанавливается только durable terminal fact, который сам Runtime
   Lifecycle Owner опубликовал для exact attempt внутри exact bound generation,
   после того как Host завершил свой owned shutdown contract, как уже требуют
   DP-016 и DP-017 section 13.
2. Его допустимость в качестве recovery evidence дополнительно требует
   `GenerationTerminated` для того же generation, чтобы запись нельзя было
   прочитать, пока Host, который она описывает, всё ещё может быть live.
3. Не вводится ни нового store, ни типа записи, ни второго пути публикации.
   Durable attempt и command facts, принадлежащие DP-014 и DP-015, и есть тот
   fact; DP-022 определяет только, когда их чтение является законным
   evidence.
4. Fact, записанный recovery или reconciliation путём, никогда не является
   shutdown-completion evidence. Выход recovery не может удостоверить то, о чём
   его просят судить, поэтому данный design запрещает такое циклическое
   использование явно.
5. Отсутствие записи не является evidence незавершения. Termination плюс
   отсутствующий допустимый shutdown-completion fact дают лишь truthful
   resource-absence и interrupted outcomes, которые DP-017 уже предписывает, и
   никогда `Stopped`.

Следовательно `resource absence != Host-owned shutdown completion` здесь
структурно: у двух заключений разные производители, разные доказательства и
отсутствие вывода одного из другого.

## 14. Evidence query contract

Evidence производится одной read-only capability, эквивалентной
`ReadExactExecutionEvidence` DP-017 section 24. Её semantic contract:

- входом является exact tuple `(containment domain, Runtime Instance, Launch
  Attempt, execution generation)`; неполный или частично разрешённый tuple
  отклоняется как `Unknown(ScopeMismatch)` до любого observation;
- выходом является ровно один результат из закрытой модели в section 15;
- операция не выполняет mutation, ничего не allocate и не bind, ничего не
  открывает и не предоставляет authority за пределами чтения;
- результат привязан к tuple, который его породил, и одноразовый: поздний
  потребитель должен выполнить повторное чтение, потому что cached или replayed
  результат является `Unknown(Stale)`;
- запрос может быть отвечен только generation authority того domain, который он
  называет, либо adapter, чей объявленный уровень гарантий покрывает этот
  случай;
- конкурентные чтения разрешены и независимы; долгие чтения не держат lock
  aggregate, command, Owner или admission, и любая последующая conditional
  publication заново валидирует revisions по DP-017 section 16;
- не существует операций batch, scan, discovery, enumeration или «nearest
  match».

## 15. Закрытые outcomes evidence

| Результат | Доказывает | Никогда не доказывает |
| --- | --- | --- |
| `GenerationLive` | опрашиваемое generation является собственным current generation читателя и его capability удерживается | что любое более раннее generation terminated; любое lifecycle completion |
| `GenerationTerminated` | exact названный prior generation superseded через exclusive re-acquisition в этом domain | graceful cleanup, успех Stop, outcome command, shutdown completion |
| `CoveredResourcesAbsent` | ни один ресурс covered класса terminated generation не остаётся удерживаемым или достижимым | отсутствие uncovered класса; успешный release приложением |
| `HostShutdownCompleted` | durable terminal fact Owner доказывает, что owned shutdown contract Host завершился для exact attempt в exact terminated generation | что recovery может reopen admission, что более поздняя phase может выполниться, либо любое более раннее claim о readiness |
| `LiveUnownedExecution` | adapter, чей approved уровень сообщает об этом, наблюдал execution, которым current процесс не владеет | manageability, права adoption, identity владельца или безопасное действие termination |
| `Unknown(reason)` | ничего | любой из выводов выше |

`reason` является закрытым и равняется одному из: `Absent`, `Unavailable`,
`Stale`, `ScopeMismatch`, `Contradictory`, `Indeterminate`, `Cancelled`,
`UnsupportedTopology`, `GuaranteeNotDeclared`.

Приоритет намеренно минимален:

1. unbound observation вообще не является входом;
2. любое несовпадение tuple, ledger или binding разрешается в
   `Unknown(ScopeMismatch)`;
3. любые два входа, которые расходятся, разрешаются в `Unknown(Contradictory)`;
4. никакое правило ordering, свежести, большинства, уверенности или стоимости не
   может выбрать winner между расходящимися facts;
5. `Unknown` является терминальным для этого чтения: он никогда не понижается до
   догадки и не повышается повторением.

`LiveUnownedExecution` недостижим в initial in-process topology, которая не
определяет уровня adapter, способного о нём сообщить; он назван, чтобы будущий
adapter не мог молча переопределить containment, и он сохраняет barrier закрытым
согласно DP-017 section 11.

## 16. Недостаточные observations

Ничто из перечисленного ниже, по отдельности или в сочетании, не является
execution evidence или частью какого-либо доказательства: process identifier или
запись process table, имя процесса или service, executable или working
directory, listening address или port, успешный bind, успешное connect, refused
connect, elapsed или wall-clock time, clock skew, boot identifier, host name,
строка лога, health или readiness response, memory dump, stored desired или
actual `Running` либо `Stopping`, persisted `Running`, обновлённый при чтении,
command `Claimed`, in-flight callback, версия configuration или Snapshot,
aggregate revision, identity Runtime Instance или Launch Attempt, либо
отсутствие чего-либо из этого.

Запрещённые выводы:

| Вывод | Статус |
| --- | --- |
| connect не удался, поэтому prior generation terminated | запрещено |
| port свободен, поэтому ресурсы Host освобождены и Stop завершился успешно | запрещено |
| процесс исчез из таблицы, поэтому shutdown contract завершён | запрещено |
| elapsed time превысил порог, поэтому permit, capability или execution утрачены | запрещено |
| durable binding существует, поэтому execution живо или жило | запрещено |
| durable state говорит `Running`, поэтому Host управляем сейчас | запрещено |
| нет записи shutdown-completion, поэтому Host не завершил shutdown | запрещено |
| публикация recovery существует, поэтому её precondition был доказан | запрещено |
| evidence stale, но favorable, поэтому используется favorable чтение | запрещено |

## 17. Доверие к adapter и уровни гарантий

Adapter отвечает на evidence запросы только на объявленном уровне гарантий и не
может его превышать.

| Уровень | Покрывает | Может сообщать | Требуемые гарантии |
| --- | --- | --- | --- |
| `None` | adapter отсутствует либо не может связать tuple | только `Unknown` | никаких; default состояние репозитория |
| `ProcessContainment` | initial single-node in-process границу | `GenerationLive`, `GenerationTerminated`, `CoveredResourcesAbsent`, `HostShutdownCompleted`, читаемые через facts DP-014/DP-015 | exclusive single-holder acquisition; release только при termination процесса без потери или двойной выдачи; отсутствие использования clock; locality ledger и pre-provisioned storage authority по section 11 и DP-023 section 8.1 |
| `ExecutionIsolation` | будущую approved child или remote границу | дополнительно `LiveUnownedExecution` | всё из `ProcessContainment` плюс approved протокол adoption и termination, которого не существует |

Каждый adapter должен объявить, для каждого уровня: какой domain он связывает,
как результат привязывается к exact tuple, что он может и не может наблюдать, на
чём основана его гарантия release-on-termination, и как он ведёт себя при
unavailability, contradiction, partial read, panic и cancellation. Любое
необъявленное или непроверяемое property делает уровень непригодным и даёт
`Unknown(GuaranteeNotDeclared)`. Adapter никогда не mutate, никогда не закрывает
barrier, никогда не классифицирует recovery set и никогда не становится owner;
generation authority, который его потребляет, остаётся внутри composition Control
Service, поэтому single composition root ADR-0003 и freeze ARCH-002
сохраняются.

## 18. Scope isolation и security

Чтения evidence и ledger scoped ровно к одному containment domain — одному
operational management domain, обслуживаемому одним Control Service вместе с
принадлежащим ему durable identity state, — и внутри него — к одному Workspace,
Configuration, Runtime Instance, Launch Attempt и execution generation.
Cross-domain evidence запрещены даже когда они доступны и favorable.

Результаты несут только opaque identities и закрытые semantic категории. Они
никогда не несут credentials, Secrets, payload configuration или Snapshot, raw
internal errors, stack traces, host pointers, process-local permits, command
callbacks или unrestricted process metadata. Читатель, который не может доказать
привязку tuple, отбрасывает вход вместо его сужения. Существование execution,
принадлежащего другому tenant, никогда не раскрывается, а concrete operator
reporting и redaction остаются за DP-018 и ARCH-004 section 19(6).

## 19. Cancellation и concurrency

1. Cancellation чтения evidence даёт `Unknown(Cancelled)` и не выполняет
   mutation. Он никогда не release capability, никогда не доказывает termination,
   никогда не разрешает contradiction и никогда не авторизует вызывающую сторону
   продолжать.
2. Cancellation lifecycle пути регулируется DP-016 и DP-017; он не изменяет
   containment facts.
3. Acquisition сериализуется самой исключительностью: конкурентные acquisitions в
   одном domain дают ровно один holder и один failure, детерминированно и без
   clock или приоритетов.
4. Конкурентные чтения evidence независимы и никогда не блокируют admission; они
   никогда не держат locks через чтение, и любая последующая conditional
   publication заново валидирует revisions по DP-017 section 16.
5. Конкурентный читатель никогда не получает authority mutation, а writer,
   который изменяет tuple, ledger или binding между чтением и использованием,
   инвалидирует это чтение.

## 20. Матрица failure

| Ситуация | Truthful результат | Запрещённое следствие |
| --- | --- | --- |
| acquisition fails или ambiguous | ни generation, ни binding, ни admission, domain unavailable | proceed с предполагаемым generation |
| acquisition succeeds, запись ledger indeterminate | `Unknown(Indeterminate)` | bind или admit на partial containment state |
| capability lost или revoked во время жизни | admission closed, никакого нового binding, никакого in-place re-acquisition | продолжать обслуживание, self-heal или restart автоматически |
| названное generation не имеет записи в ledger | `Unknown(ScopeMismatch)` | трактовать отсутствие как termination |
| запись ledger равна current generation | `GenerationLive` | сообщить о собственном termination |
| ledger и current holder различаются и гарантии исключительности держатся | `GenerationTerminated` | выводить cleanup или успех Stop |
| adapter не может доказать release-on-termination | `Unknown(GuaranteeNotDeclared)` | откатиться к clock, PID или probe |
| два чтения расходятся | `Unknown(Contradictory)` | выбрать свежее, большинство или удобное |
| cached чтение переиспользовано позже | `Unknown(Stale)` | переиспользовать как current truth |
| запрос cancelled | `Unknown(Cancelled)` | брать favorable outcome по умолчанию |
| binding называет generation другого domain | `Unknown(ScopeMismatch)` | cross-domain inference |
| Host отсутствует, shutdown-completion fact нет | только отсутствие covered-ресурсов | публиковать или выводить `Stopped` |
| durable terminal fact Owner плюс termination proof | `HostShutdownCompleted` | пропускать любую из двух половин доказательства |
| durable terminal fact записан recovery | не является evidence | циклическая сертификация |
| будущий adapter сообщает unowned live execution | `LiveUnownedExecution` | adoption, replay или forced termination |
| ресурс uncovered класса может существовать | `Unknown`, и barrier остаётся закрытым | полагать, что containment покрывает его |
| process table говорит, что старый PID исчез | `Unknown` | termination proof |
| port отказывает в соединениях | `Unknown` | release ресурсов или успешный Stop |

## 21. Сводка ownership

| Fact | Owner | Роль DP-022 |
| --- | --- | --- |
| containment capability и её исключительность | composition Control Service (generation authority) | определяет semantics и failure behavior |
| current generation identity | generation authority | определяет правила выдачи и переиспользования |
| durable attempt-to-generation binding | DP-014 | ограничивает допустимые значения и смысл |
| containment ledger | generation authority как writer, recovery как reader | определяет содержимое и запреты |
| live Host ownership и lifecycle decisions | Runtime Lifecycle Owner (ARCH-004 section 9) | не затронуто; evidence никогда не передаёт это |
| ресурсы Host | Runtime Host | не затронуто; containment лишь ограничивает их lifetime |
| recovery classification, claim, permit, barrier | DP-017 | поставляет только facts и гарантии |
| durable outcomes command и attempt | DP-014, DP-015, DP-016 | определяет, когда они считаются evidence |
| operator reporting и redaction | DP-018, ARCH-004 section 19(6) | явно вне области действия |

## 22. Technology Neutrality

Containment capability, generation identity, ledger, supersession и закрытые
результаты evidence — это semantic requirements. Design не выбирает и не требует
database, file lock, PID file, named object, semaphore, lease service, consensus
или quorum механизм, clock, boot identifier, process supervisor, operating-system
primitive, container или VM средство, vendor либо format identifier. Он не
авторизует generic registry, service locator, process manager или background
scanner и не добавляет dependency или package.

Реализация может выбрать любой механизм, который доказывает exclusive
single-holder acquisition, release только при termination процесса, durable
append-only supersession, точную привязку tuple и fail-closed поведение
`Unknown`. Механизм, который не может доказать что-либо из этого, не является
частичной реализацией данного design; по определению он даёт `Unknown`.

## 23. Явно отложенные вопросы

Отложены в отдельные approved designs или implementation tasks:

- adapter для child-process, remote-worker и container, включая любое
  обнаружение live-orphan, adoption и протокол termination;
- process supervision, scheduling, clustering, cross-node quorum и любой
  multi-holder domain;
- concrete storage engine, schema encoding, migrations, retention, compaction и
  deployment layout сверх guarantee class и atomic bootstrap, утверждённых
  DP-023;
- sequencing запуска Control Service, signal handling, shutdown orchestration,
  production wiring и Production Activation;
- recovery claim, permit, assessment, reconciliation и mechanics barrier,
  которые принадлежат DP-017;
- public или internal API, DTO, status mapping, authorization policy, health
  endpoint, operator reporting и redaction;
- metrics, logging, tracing, alerting и retention;
- автоматический restart, retry, backoff, failover и remediation policy.

## 24. Acceptance Proofs

Будущая реализация этой границы должна доказать как минимум:

1. не более одного live процесса Control Service держит capability для одного
   domain, при конкурентном acquisition;
2. current generation выдаётся ровно один раз на успешный acquisition, до
   admission, claim attempt и binding;
3. ни одно generation identity не выводится из PID, address, port, clock, host
   name, версии Configuration или любого identity Runtime либо Attempt и не
   равно им;
4. переиспользование identity внутри domain обнаруживается и разрешается в
   `Unknown` с закрытым admission;
5. failure или неоднозначность acquisition не дают ни binding, ни admission, ни
   вызова lifecycle;
6. потеря capability закрывает admission и не может быть восстановлена на
   месте;
7. ни один durable fact, probe, лог, результат port или elapsed time сами по
   себе не дают никакого положительного результата section 15;
8. terminated prior generation доказывается только через exclusive
   re-acquisition плюс совпадающая запись ledger;
9. termination proof никогда не даёт `HostShutdownCompleted`, cleanup, успех Stop
   или outcome command;
10. отсутствие covered-класса сообщается без заявления отсутствия
    uncovered-класса;
11. `HostShutdownCompleted` требует durable terminal fact Owner и termination
    proof того же exact generation;
12. durable fact, записанный recovery, никогда не принимается как shutdown
    evidence;
13. binding, называющий generation, отсутствующий в ledger domain, даёт
    `Unknown(ScopeMismatch)`;
14. contradictory, stale, cancelled, unavailable и unsupported-topology входы
    каждый дают свой точный reason `Unknown` и никакой mutation;
15. чтения evidence не выполняют mutation, не открывают admission и не выдают
    authority;
16. конкурентные читатели никогда не сериализуются и не обходят revalidation
    revisions DP-017;
17. ни один adapter ниже `ProcessContainment` не может дать положительный
    результат;
18. ledger не содержит ни одного из запрещённых полей section 11;
19. cross-domain и cross-Workspace evidence никогда не используются, и результаты
    не раскрывают запрещённого payload;
20. EN/RU contract, матрицы, outcome names, уровни гарантий и статус Planned
    остаются aligned.

## 25. Граница реализации

Реализованы изолированно сегодня: opaque
identity type `ExecutionGeneration`, conditional attempt-to-generation binding
операция DP-014, request-exactly-once provider seam DP-020 и Windows-only
bootstrap package slice DP-023 из TASK-069. Последний устанавливает private
capability, ledger и generation authority только внутри package. Latest
verification, review и Acceptance checkpoint определяется только newest valid
matching envelope TASK-069. Evidence adapter, production composition wiring и
code path, который мог бы expose termination evidence Control Service,
отсутствуют.

DP-023 — Approved/Implemented in isolation по explicit Coordinator status
decision через TASK-069. Mutable role verdicts и identities resolve-ятся из
newest valid matching envelope этой task. Все evidence и downstream gates
остаются без изменений; сам DP-022 остаётся Planned.

Этот документ есть Approved design граница, поэтому DP-017 section 11 теперь
имеет authoritative containment boundary для потребления; сам DP-017 остаётся
Approved/Planned и неактивированным. Статус получен явным решением через project
design status процесс; acceptance задачи со стороны Documentation, Tester,
Reviewer или Coordinator не повышает Design Status этого документа и никогда не
повышает Implementation Status. Isolated bootstrap candidate не активирует
evidence: evidence adapter и composition отсутствуют, поэтому exact
prior-generation termination proof, требуемый DP-017 section 11, по-прежнему не
может потребляться ни одним компонентом, а DP-017 recovery, DP-018 reporting,
production integration и Production Activation остаются `Not Activated` и
отсутствуют; downstream consumption containment evidence относится к более
поздней, отдельно approved границе. Ни один gate ARCH-004 section 19 здесь не
заявляется и не переоткрывается.

## 26. Решение

UWP делает execution containment свойством исключительности одного процесса
Control Service на containment domain. Composition acquired одну exclusive
containment capability ровно один раз, выдаёт под ней одно opaque execution
generation и никогда не release эту capability, пока является authoritative.
Termination exact prior generation, названного durable attempt binding,
доказывается только exclusive re-acquisition, записанным как supersession в
durable containment ledger — никогда с помощью clock, lease, записи process
table, PID, address, port, probe или stored lifecycle state.

Host-owned shutdown completion доказывается другим fact, произведённым Runtime
Lifecycle Owner и durable через DP-014 и DP-015, и допустимым в качестве evidence
только вместе с доказательством termination того же generation. Поэтому resource
absence никогда не доказывает shutdown completion, а отсутствие записи никогда не
доказывает незавершения. Любой другой случай — это `Unknown`, и `Unknown` ничего
не adopt, ничего не replay, ничего не admits и не изобретает terminal truth.
Containment для child-process и remote, adoption, supervision и production wiring
остаются отдельными approved решениями.

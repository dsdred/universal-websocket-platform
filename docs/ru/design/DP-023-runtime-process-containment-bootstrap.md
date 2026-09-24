# DP-023: Bootstrap process-containment Runtime

[English version](../../en/design/DP-023-runtime-process-containment-bootstrap.md)

## 1. Статус

- **Design Status:** Approved
- **Implementation Status:** Implemented in isolation

TASK-068 утверждает эту implementation boundary и decomposition. Approval делает
один последующий code slice Ready для отдельного intake; он не создаёт package,
adapter, ledger, generation authority, positive evidence или production
capability.

TASK-069 реализует в worktree изолированный Windows-only первый slice:
validation независимого trusted descriptor,
existing-only заранее provisioned anchor/bbolt ledger, non-inheritable Windows
`Global\\` Event, exact successor append/inspection, identity checks и
fail-closed tests. Coordinator явно установил этот Implementation Status через
TASK-069. Mutable Tester/Reviewer verdicts и subject identities resolve-ятся
только из newest valid matching envelope этой task и здесь не дублируются;
latest verification, review и Acceptance checkpoint определяется этим
envelope. Package не wired в Control Service; evidence reading, recovery,
reporting, provisioning и Production Activation отсутствуют.

## 2. Назначение

Определить минимальный implementable bootstrap, который может достоверно заявить
гарантию `ProcessContainment`, требуемую DP-022 и позднее потребляемую DP-017.
До admission любой management work Runtime bootstrap обязан установить один
authoritative process, одну fresh opaque execution generation и один durable
supersession transition как единое safety behavior.

## 3. Полномочия и приоритет

Это предложение уточняет, но не заменяет:

- ownership, lifecycle, serialization и activation gates ARCH-004;
- fail-closed lifecycle semantics ADR-0003;
- attempt-to-generation binding DP-014;
- command truth DP-015 и orchestration DP-016;
- recovery ownership и ordering DP-017;
- outcomes containment и exact evidence DP-022.

Если этот документ противоречит Accepted ADR, Active architecture guide или
Approved ownership rule выше, побеждает источник с более высоким приоритетом,
а implementation останавливается до нового решения.

## 4. Область действия

Этот design владеет только:

- initial local process-containment guarantee class;
- одним safety-atomic acquisition, generation и ledger bootstrap;
- правилами retention capability и fatal fencing;
- private package и adapter boundary;
- executable conformance requirements первого implementation slice.

Он не определяет evidence reader, recovery action, public API, production wiring
или activation.

## 5. Не-цели

- выбор storage vendor, database product или именованного operating-system lock;
- release, unlock, transfer, lease, renewal, expiry или in-place reacquisition;
- PID, process-name, address, port, health, wall-clock или elapsed-time probes;
- attempt binding, lifecycle mutation, command mutation или recovery mutation;
- `ReadExactExecutionEvidence` или composition shutdown-completion evidence;
- child, remote, container, cluster, adoption или supervision protocols;
- retention, compaction, migration, deletion или arbitrary repair ledger;
- реализация production provisioning domain или production startup wiring;
- integration Control Service, public DTO, authorization или Production
  Activation.

## 6. Initial guarantee source

Initial implementation guarantee:

> Stable-domain, host-kernel-enforced process-scoped exclusive capability,
> приватно удерживаемая в течение authoritative lifetime Control Service и
> автоматически освобождаемая только termination процесса, в паре с
> crash-consistent, single-writer, durable append-only containment ledger,
> закреплённым в том же stable domain namespace.

Гарантия требует все следующие свойства:

1. exclusivity enforced вне обычной application memory;
2. второй live process не может acquire capability того же domain;
3. application code не может явно release или transfer capability;
4. retention остаётся strong до process exit и не зависит от finalizer или
   caller-owned handle;
5. namespace не может молча разрешиться в две независимые authority;
6. commit ledger crash-consistent и inspectable как точно committed или absent;
7. capability holder является единственным configured writer ledger для domain;
8. неопределённость guarantee, namespace, storage, corruption или inspection не
   даёт authority и positive evidence.
9. authoritative storage root и заранее provisioned anchor являются immutable
   для domain; loss или replacement нельзя принять за first use.

Database session lock, file marker, PID file, expiring lease или process-local
mutex недостаточны, если их concrete contract и subprocess tests не доказывают
каждое свойство выше.

## 7. Ownership

| Fact или authority | Единственный owner |
| --- | --- |
| Exclusive containment capability | generation authority `runtimecontainment` |
| Current execution generation | generation authority `runtimecontainment` |
| Append-only generation/supersession ledger | та же authority как sole writer |
| Attempt-to-generation binding | DP-014 / `runtimeidentity` |
| Host и lifecycle decisions | Runtime Lifecycle Owner |
| Command truth | command boundary DP-015 |
| Recovery classification, claim и barrier | DP-017 |
| Operator projection | DP-018 |

Ledger является отдельным containment fact set. Это не второй Runtime aggregate,
lifecycle store, command store, attempt store или recovery store.

## 8. Containment domain и identities

Containment domain — stable operational namespace для одной initial authority
Control Service. Domain identity и execution-generation identity являются
opaque, non-zero и non-reusable. Ни одна из них не выводится из PID, времени,
адреса, hostname, process name или mutable Runtime state.

Одна normalized domain identity выбирает и exclusive capability, и её ledger.
Namespace mismatch, aliasing, replacement или uncertain resolution являются
fatal; implementation не продолжает работу под заново разрешённым namespace.

### 8.1 Provisioned domain и storage authority

Authoritative immutable storage root входит в provisioning containment domain.
Доверенный deployment provisioning до Runtime bootstrap связывает одну
normalized Domain с одной storage authority и существующим domain anchor/ledger
в этом root. Anchor содержит Domain и opaque, non-reused storage-authority
identity как metadata namespace/provenance; он не является generation record
или вторым источником истины о Runtime. Canonical path, полученный из Domain
внутри root, лишь находит storage, но не доказывает, что это исходная authority.
Смена root или повторный provisioning той же Domain не создают legitimate
first use.

Runtime bootstrap получает от deployment boundary immutable trusted
provisioning descriptor независимо от candidate root, anchor, ledger и любых
bytes внутри этого root. Descriptor связывает normalized Domain с expected
canonical root, expected physical root identity и expected opaque
storage-authority identity. Candidate storage предоставляет только observed
values; он никогда не поставляет, не выводит, не переопределяет и не repair-ит
expected values descriptor и не может удостоверять собственную authority.
Перед `ProvisionedEmpty` или выдачей authority bootstrap сверяет observed
canonical/physical root identities, Domain и storage-authority identity anchor,
а также binding ledger с descriptor. Alternate или copied store выполняет fail
closed, даже если его anchor и ledger внутренне согласованы.

`ProvisionedEmpty` означает, что trusted anchor и ledger уже существуют и
проверены как provisioned authority, но ledger ещё не содержит generation
entry. Только из этого состояния можно commit first generation без
predecessor. Отсутствующий anchor, ledger или directory никогда не являются
`ProvisionedEmpty`. Runtime открывает provisioned storage только existing-only;
он не создаёт, не ремонтирует, не мигрирует, не перемещает, не заменяет и не
перепривязывает Domain, anchor, root или ledger.

Deployment/provisioning authority обязана сохранять привязку Domain-to-root и
storage-authority через restarts и предотвращать несанкционированные deletion,
replacement, cloning, relocation и rollback anchor и ledger как единого
целого. Она никогда не переиспользует storage-authority identity и не считает
потерю ранее provisioned Domain созданием новой Domain. Bootstrap может
проверить наблюдаемые несовпадения root, anchor, Domain, authority identity и
ledger; он не может обнаружить идеально согласованный joint rollback или
clone, читая только откатанное или клонированное storage. Поэтому заявленная
гарантия `ProcessContainment` опирается на доверенную deployment/storage
границу, предотвращающую этот ненаблюдаемый случай. Среда, которая не может
установить эту границу, не может заявлять гарантию или выдавать authority.
Реализация production provisioning и wiring вне первого implementation slice.

Missing, mismatched, replaced, relocated, stale или alternate candidate
storage и неопределённость authoritative root выполняют fail closed. Failure
до acquisition не даёт authority; после acquisition domain переходит в
`FatalFenced`, а process завершается. Fallback root и automatic reprovisioning
запрещены в обоих случаях.

## 9. Safety-atomic bootstrap

Acquisition capability, создание fresh generation и durable ledger transition —
одно неделимое safety behavior, даже если kernel capability и durable append не
разделяют physical transaction.

Authority существует только после выполнения всех условий:

1. process исключительно держит capability domain;
2. existing provisioned anchor и ledger открыты и проверены под этой
   capability в authoritative storage root;
3. создан ровно один fresh opaque generation candidate;
4. exact successor entry durably committed при удерживаемой capability;
5. exact inspection подтверждает candidate как current ledger tail.

Durable successor commit — logical linearization point. Raw acquisition
capability и private candidate являются provisional и не дают generation
authority.

## 10. Протокол перехода ledger

Для previous tail `Gprev` и fresh candidate `Gnew` единственный успешный
transition:

```text
expected tail = Gprev
append current generation = Gnew
record predecessor = Gprev
thereby record Gprev superseded
```

First generation не имеет predecessor только тогда, когда expected state —
проверенное `ProvisionedEmpty`. Append сравнивает exact expected tail либо это
явное empty state. Committed generation никогда не rewritten, removed,
reordered или reused.
После inspected definite absence можно retry только exact candidate.

## 11. Crash и indeterminate cuts

| Cut | Обязательный результат |
| --- | --- |
| до exclusive acquisition | нет mutation, generation или authority |
| acquisition fails или ambiguous | domain unavailable; нет ledger write или downstream work |
| provisioned anchor/ledger отсутствует, mismatched, replaced, relocated, stale, alternate или uninspectable | нет вывода о first use; fail closed, а при удерживаемой capability — fatal fence |
| capability held до создания candidate | только provisional; crash не оставляет durable generation |
| candidate создан до append | candidate остаётся private и unbound |
| definite append failure и inspected tail не изменился | retry exact candidate при удерживаемой capability |
| append outcome indeterminate | сначала inspect под held capability; не создавать другой candidate |
| exact candidate является current tail | converge to success |
| candidate absent и exact prior tail не изменился | retry только exact append и candidate |
| другой tail, partial/corrupt ledger или inspection unavailable | fatal fence для domain и termination |
| crash после commit до возврата authority | следующий process может append successor; attempt не был bound |
| crash после возврата authority | process termination releases capability; следующий acquisition append successor |
| live capability loss, revocation или namespace uncertainty | закрыть authority, не reacquire in place и terminate |
| caller cancellation после provisional acquisition | reconcile exact transition или fatal-fence; не release/transfer |

Post-acquisition fatal path не может вернуть process в обычный service.

## 12. Retention capability и fatal fencing

Authoritative process удерживает capability в private strongly reachable state
всю свою lifetime. Ни один consumer не получает raw handle. Normal cleanup
operation для release отсутствует.

Loss, revocation, guarantee downgrade, namespace ambiguity, unreconciled append,
storage-authority uncertainty или corruption навсегда переводят process-domain
state в `FatalFenced` после acquisition.
Fencing закрывает admission, отключает generation provision и positive evidence
и требует process termination. Process не может reacquire authority in place.

## 13. Модель ledger

Каждая запись содержит только:

- identity containment domain;
- opaque current generation identity;
- optional exact predecessor identity.

Она не содержит PID, time, address, Host state, lifecycle phase, command,
attempt, recovery claim, operator annotation, payload или user data. Ledger
append-only, single-writer, domain-isolated и durable через process termination.
Loss, partial state, duplicate identity, impossible predecessor или unverified
tail означают unavailable/fatal, а не `ProvisionedEmpty` и не positive termination
evidence.

## 14. Концептуальный API

Public package semantics намеренно узки:

```text
AcquireProcessContainment(domain, trustedProvisioningDescriptor) -> ActiveAuthority | Unavailable | FatalFenced

ActiveAuthority.CurrentGeneration() -> exact committed current generation
ActiveAuthority.GuaranteeLevel() -> ProcessContainment
ActiveAuthority.IsAuthoritative() -> true only while unfenced
```

Private driver semantics:

```text
AcquireExclusiveProcessLifetime(domain)
OpenExistingProvisionedAnchor(heldCapability, domain, trustedProvisioningDescriptor)
ReadLedgerTail(heldCapability, domain)
AppendSuccessor(heldCapability, domain, expectedTail, exactCandidate)
InspectExactAppend(heldCapability, domain, exactCandidate)
```

Ни одна surface не раскрывает release, unlock, transfer, renew, expiry,
arbitrary ledger mutation, PID lookup, evidence classification, attempt binding,
lifecycle, command или recovery mutation.

## 15. Package и dependency boundary

Первый implementation target:

```text
internal/runtimecontainment
```

Он владеет opaque domain/generation types, invalid-zero semantics, capability
state machine, existing-only проверкой anchor, records ledger и expected-tail
append, candidate creation, guarantee identity, fatal fencing, одним concrete
initial local adapter и его subprocess/crash conformance harness.

Dependency direction:

```text
OS and durable adapter
        ↓
internal/runtimecontainment
        ↓ later composition adapter
existing ProvideExecutionGeneration seam
        ↓
runtimeactivation and DP-014 binding path
```

`runtimecontainment` не импортирует runtimeactivation, command, lifecycle,
recovery, HTTP или reporting package. Downstream packages не allocate и не
replace его generation.

## 16. Admission и composition boundary

До authoritative bootstrap completion composition не должна открыть management
admission, принять state-changing command, предоставить generation, bind attempt,
load Runtime или вызвать lifecycle work. Wiring committed generation в
существующий provider seam и proof no bypass относятся к последующему slice, а
не к bootstrap package.

Следовательно, Slice 1 доказывает local authority semantics, но не заявляет, что
current composition Control Service уже их обеспечивает.

## 17. Первый implementation slice

Первый implementation slice — **Runtime Process-Containment Bootstrap over a
Pre-Provisioned Domain Anchor**:

- один package `internal/runtimecontainment`;
- один real local adapter, удовлетворяющий section 6;
- existing-only open и validation заранее provisioned domain anchor и ledger
  под trusted storage authority из section 8.1;
- safety-atomic protocol sections 9–11;
- fatal fencing из section 12;
- subprocess, restart, crash-cut, concurrency, durability, identity-reuse и
  corruption proofs, включая missing/deleted anchor, alternate root,
  replacement, relocation, stale copy и два candidate stores.

Evidence reading, изменения DP-014, provider wiring, production provisioner,
recovery, reporting, public API, production integration и activation остаются
явными non-goals.

## 18. Матрица доказательств

| Требование | Покрытие первого slice |
| --- | --- |
| DP-022 proof 1: не более одного live holder | direct concurrent-subprocess proof |
| proof 2: одна generation на acquisition | bootstrap direct; admission/binding deferred to composition |
| proofs 3–4: opaque unique identity и reuse fencing | direct structural и restart proofs |
| proof 5: failure не даёт authority | direct; downstream no-work proof deferred to composition |
| proof 6: loss закрывает authority без in-place recovery | direct |
| proof 7: probes/time не устанавливают evidence | direct API absence; positive evidence deferred |
| proof 8: termination substrate | direct exclusive reacquisition плюс ledger; classification deferred |
| proofs 9–16 | deferred to evidence и recovery slices |
| proof 17: lower guarantee никогда не positive | direct adapter conformance |
| proof 18: forbidden ledger fields отсутствуют | direct structural proof |
| proof 19: domain isolation | direct для ledger; cross-Workspace evidence deferred |
| DP-017 section 11 и proofs 7, 11–13 | compositional prerequisite only; нет recovery claim |

Required scenarios включают N concurrent processes для одного domain с ровно
одним winner, independent domains, forced termination, один successor на
restart, каждый crash cut, inspection indeterminate append,
committed-but-unacknowledged append convergence, corruption и namespace
mismatch, rejection identity reuse, отсутствие release/reacquire API, race
checks и repository regression tests.
Harness provision-ит anchor до вызова bootstrap и доказывает first generation
только из проверенного `ProvisionedEmpty`; missing/deleted anchor, alternate
root, mismatched identity, replacement, relocation, stale copy и два candidate
stores выполняют fail closed. Тесты наблюдаемого mismatch не заявляют
обнаружение ненаблюдаемого joint rollback/clone; adapter обязан объявить и
проверить свою deployment/storage trust precondition до заявления guarantee.
Тесты также получают expected canonical root, physical root identity и
storage-authority identity только через независимый trusted provisioning
descriptor. Candidate anchor или store, копирующий либо self-reporting
совпадающие expected values, не заменяет descriptor; alternate/copied storage
fail closed against it, даже если внутренне согласован.

## 19. Ordered downstream decomposition

1. DP-023 process-containment bootstrap.
2. DP-022 exact generation evidence reader.
3. DP-022 composition shutdown-completion evidence.
4. containment composition/admission/provider gate.
5. DP-017 read-only recovery assessment.
6. DP-017 durable recovery claim и admission barrier.
7. DP-017 reconciliation attempt и primitive command.
8. DP-017 reconciliation linked phase и parent.
9. DP-017 coherent release и reopening barrier.
10. DP-018 reporting, затем production integration и Production Activation.

Каждый последующий slice требует fresh intake и prerequisite check. Этот порядок
не активирует ни один из них.

## 20. Size Guard

Verdict: **ACCEPT — ONE INDIVISIBLE ATOMIC IMPLEMENTATION BEHAVIOR**.

Target — один package и одно independently shipped behavior. Capability, ledger
и generation issuance — неразделимые internal parts одного safety transition.
Evidence reading и все consumers отделены. Вторая adapter family, evidence
reader, wiring или recovery behavior требует split; превышение обычного
production-line threshold требует re-evaluation, а не exposure partial
authority.

## 21. Граница реализации

Implementation Status — `Implemented in isolation`. Worktree TASK-069 содержит
Windows-only slice `internal/runtimecontainment` с conforming
capability, pre-provisioned anchor/bbolt ledger, authoritative generation
bootstrap, fail-closed stub для unsupported platforms и focused proofs. Он не
composed в Control Service; positive evidence reader отсутствует. Поэтому
current production composition behavior и DP-017 не меняются; evidence,
recovery, reporting, provisioning, integration и Production Activation
остаются последующими работами.

## 22. Решение

UWP установит initial `ProcessContainment` через одну process-lifetime exclusive
capability и один same-domain durable append-only ledger в доверенной, заранее
provisioned immutable storage authority. Первый implementation slice обязан
acquire capability, проверить existing anchor и ledger, создать одну opaque
generation, durably append и inspect её exact successor record до exposure
authority. First generation требует проверенного `ProvisionedEmpty`;
отсутствующее или неопределённое storage никогда не является first use. Любая
ambiguous guarantee или unreconciled state выполняет fail closed; после
acquisition process fatal-fenced и завершается. Evidence, composition,
recovery, reporting, реализация provisioning и activation остаются
последующими работами.

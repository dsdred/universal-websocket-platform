# DP-023: Bootstrap process-containment Runtime

[English version](../../en/design/DP-023-runtime-process-containment-bootstrap.md)

## 1. Статус

- **Design Status:** Approved
- **Implementation Status:** Planned

TASK-068 утверждает эту implementation boundary и decomposition. Approval делает
один последующий code slice Ready для отдельного intake; он не создаёт package,
adapter, ledger, generation authority, positive evidence или production
capability.

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

## 9. Safety-atomic bootstrap

Acquisition capability, создание fresh generation и durable ledger transition —
одно неделимое safety behavior, даже если kernel capability и durable append не
разделяют physical transaction.

Authority существует только после выполнения всех условий:

1. process исключительно держит capability domain;
2. создан ровно один fresh opaque generation candidate;
3. exact successor entry durably committed при удерживаемой capability;
4. exact inspection подтверждает candidate как current ledger tail.

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

First generation не имеет predecessor. Append сравнивает exact expected tail.
Committed generation никогда не rewritten, removed, reordered или reused.
После inspected definite absence можно retry только exact candidate.

## 11. Crash и indeterminate cuts

| Cut | Обязательный результат |
| --- | --- |
| до exclusive acquisition | нет mutation, generation или authority |
| acquisition fails или ambiguous | domain unavailable; нет ledger write или downstream work |
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

Loss, revocation, guarantee downgrade, namespace ambiguity, unreconciled append
или corruption навсегда переводят process-domain state в `FatalFenced`.
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
tail означают unavailable/fatal, а не empty ledger и не positive termination
evidence.

## 14. Концептуальный API

Public package semantics намеренно узки:

```text
AcquireProcessContainment(domain) -> ActiveAuthority | Unavailable | FatalFenced

ActiveAuthority.CurrentGeneration() -> exact committed current generation
ActiveAuthority.GuaranteeLevel() -> ProcessContainment
ActiveAuthority.IsAuthoritative() -> true only while unfenced
```

Private driver semantics:

```text
AcquireExclusiveProcessLifetime(domain)
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
state machine, records ledger и expected-tail append, candidate creation,
guarantee identity, fatal fencing, одним concrete initial local adapter и его
subprocess/crash conformance harness.

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

Следующий допустимый intake — **Runtime Process-Containment Bootstrap
Implementation**:

- один package `internal/runtimecontainment`;
- один real local adapter, удовлетворяющий section 6;
- safety-atomic protocol sections 9–11;
- fatal fencing из section 12;
- subprocess, restart, crash-cut, concurrency, durability, identity-reuse и
  corruption proofs.

Evidence reading, изменения DP-014, provider wiring, recovery, reporting, public
API, production integration и activation остаются явными non-goals.

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

Implementation Status остаётся `Planned`. Repository не содержит package
`internal/runtimecontainment`, conforming capability, containment ledger,
authoritative generation bootstrap или positive evidence reader. Approval этого
design и Acceptance TASK-068 могут только обосновать отдельный code-task intake.
Они не меняют current runtime behavior и не удовлетворяют DP-017.

## 22. Решение

UWP установит initial `ProcessContainment` через одну process-lifetime exclusive
capability и один same-domain durable append-only ledger. Первый implementation
slice обязан acquire capability, создать одну opaque generation, durably append
и inspect её exact successor record до exposure authority. Любая ambiguous
guarantee или unreconciled state fail closed и terminate process; evidence,
composition, recovery, reporting и activation остаются последующими решениями.

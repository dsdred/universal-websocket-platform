# ARCH-003: Runtime Foundation Migration Revision

[English version](../../en/architecture/ARCH-003-runtime-migration-revision.md)

**Статус:** Active

**Область:** migration architecture фундамента Runtime

**Стабильность:** Approved migration sequence

## 1. Purpose

ARCH-003 фиксирует пересмотренную последовательность реализации фундамента Runtime, определённого в [DP-003](../design/DP-003-runtime-session-manager.md) и [DP-004](../design/DP-004-per-session-execution-boundary.md).

Target architecture остаётся неизменной. Документ изменяет только migration architecture: порядок и границы, через которые репозиторий может достичь утверждённой цели без временного появления некорректного ownership, publication или shutdown behavior.

Документ дополняет [ARCH-002](ARCH-002-runtime-foundation-freeze.md), [Master Engineering Plan](../roadmap/MASTER_PLAN.md) и зафиксированное [текущее состояние](../../../spec/current-state.md). Он не заменяет DP-003 или DP-004.

Только для migration sequencing ARCH-003 заменяет последовательность из section 29 DP-004. Каждый нормативный contract target architecture в DP-003 и DP-004 остаётся неизменным.

## 2. Decision Context

Предыдущая оставшаяся последовательность рассматривала Dormant Execution, интеграцию Execution Binding, migration Dispatcher, интеграцию Runtime callback и shutdown Runtime как отдельно активируемые шаги. Такая последовательность не могла быть безопасно исполнена, хотя утверждённая target architecture оставалась согласованной.

Дефект находился в migration sequencing, а не в DP-003 или DP-004. Несколько primitives могут быть независимо реализованы и протестированы, но их production effects ownership и publication неразделимы.

## 3. Completed Migration Tasks

Следующие migration tasks завершены:

1. Session Core;
2. Provisional Session;
3. Cleanup Acknowledgement;
4. Pre-Commit Session Bundle.

Эти tasks сохраняют synchronous production Dispatcher. Private pre-Commit bundle является структурно полным подготовленным Session-side object graph, а не нормативным Manager Commit result. Ownership transfer, execution publication и Runtime integration не выполнялись.

## 4. Primitive and Production Responsibility

Независимо реализуемый **primitive** имеет один локальный инвариант, который можно полностью проверить без активации неполного production behavior. Существующие примеры включают one-shot Execution Binding, bound Completion Adapter и accounting Owner Lifetime Lease.

Неразделимая **production responsibility** объединяет effects, которые должны становиться observable одновременно. Её нельзя разделять на production steps, если это создаёт resource без owner, partial Commit, untracked Session, неполный terminal path или ложную convergence shutdown.

Независимая реализация primitive не разрешает его изолированную production activation.

## 5. Inseparable Migration Boundaries

### Dormant Handoff

Dormant Execution нельзя отделить от:

- интеграции одного Execution Binding;
- одного domain-specific `CommitHandoff`, который связывает one-shot outcome с immutable `CommitResult`, созданным Manager;
- pre-Commit ownership Dispatcher;
- одной publication `NotCommitted` для каждого non-committed outcome;
- возврата dormant path до продолжения Dispatcher;
- disposal prepared values;
- abort Reservation.

Dormant launch path не является passive value или alias Execution Owner. Без Binding и принадлежащей Dispatcher последовательности disposal path может стать orphan или остаться заблокированным после recoverable pre-Commit failure.

### Atomic Commit Publication

Successful publication Manager Commit нельзя отделить от:

- publication Registration;
- Completion capability, привязанной к Registration;
- publication Owner Lifetime Lease;
- Stop-publication capability;
- хранения identity `CommitHandoff` для validation repeated Commit;
- publication `Committed` с полным immutable `CommitResult` через этот handoff;
- irreversible ownership transfer.

Эти effects используют одну synchronization boundary. Failure не публикует ни один из них. Observable Registration не может существовать без уже eligible единственного execution path, а dormant path не может наблюдать `Committed` без RegistrationID, Completion Adapter и Owner Lifetime Lease.

### Complete Owner Lifecycle

Production handoff не должен происходить до появления полного terminal lifecycle Owner. Освобождение committed path, останавливающегося в `Terminalizing`, оставило бы нерешёнными обязательства Cleanup, Complete, Observer, callback drain, Terminal и release lease. Accounting Manager не смог бы правдиво сойтись.

### Production Cutover and Shutdown

Production activation Runtime должна объединяться с правдивой shutdown integration. Runtime не должен начинать создавать Manager-tracked Sessions, пока shutdown Host не может запросить Stop через захваченный Snapshot, отменить root Runtime context, выполнить drain handlers Listener и дождаться accounting Manager.

## 6. Revised Remaining Migration Roadmap

### Task 5: Complete Atomic Commit Publication Foundation

**Objective:** завершить Manager-side publication contract и совместимый capability-bearing Shutdown Snapshot без активации production Session handoff.

**Architectural invariant:** successful Manager Commit создаёт один immutable `CommitResult`, содержащий RegistrationID, Registration-bound Completion Adapter и Owner Lifetime Lease, и публикует его как `Committed` через узкую Manager-facing capability `CommitHandoff` transaction под одной synchronization boundary. Binding Stop остаётся state Registration, а не owner payload. Failure не публикует committed state.

**Dependencies:** завершённые Tasks 1–4 и существующие foundations Manager, Execution Binding, Completion Adapter, Lifetime Lease и control Owner.

**Out of scope:** создание dormant launch path, выбор production Dispatcher, lifecycle Runtime callback, полный terminal execution Owner и cutover shutdown Runtime.

Реализованный foundation Task 5 использовал raw committed-only Execution publisher до появления полного dormant contract. До создания или выбора transaction-capable Dispatcher в Task 9 эта input boundary должна быть сужена до Manager-facing publisher `CommitHandoff`, а repeated Commit должен проверять ту же identity handoff. Эта prerequisite correction сохраняет linearization point и accounting Task 5 и сама по себе не активирует dormant execution или production handoff.

### Task 6: Runtime-Cancellation and Control-Call Lifecycle

**Objective:** завершить Owner-local contracts observation Runtime cancellation и admission, accounting, unregister и drain control calls.

**Architectural invariant:** Explicit Stop и Runtime cancellation разделяют одно first-writer causal state, а registration и drain callback имеют один bounded lifecycle.

**Dependencies:** необходимые contracts capability control Owner, заданные существующими foundations и Task 5.

**Out of scope:** migration Session Start или Run, invocation Manager Commit, publication Terminal Result, dormant execution, выбор production Dispatcher и activation shutdown Runtime.

### Task 7: Terminal Result and Observer Contracts

**Objective:** добавить границы immutable terminal outcome и synchronous Observer без активации production handoff.

**Architectural invariant:** один committed Owner создаёт не более одного immutable Terminal Result, а одна synchronous invocation Observer не получает lifecycle ownership.

**Dependencies:** необходимые committed capability contracts, заданные Task 5.

**Out of scope:** реализация Runtime callback, migration execution Session, полная terminal orchestration, dormant execution, migration Dispatcher и composition Runtime.

Tasks 6 и 7 логически параллельны. Обычно их следует commit-ить последовательно, чтобы каждый concurrency и ownership contract получил сфокусированное review.

### Task 8: Complete Execution Owner Terminal Lifecycle

**Objective:** расширить существующий skeleton Execution Owner до полной committed terminal sequence.

**Architectural invariant:** каждое committed execution следует порядку:

```text
Cleanup
    -> Complete
    -> Terminal Result
    -> Observer
    -> UnregisterAndDrain
    -> seal
    -> Terminal
    -> conditional Lifetime Lease release
```

**Dependencies:** Tasks 6 и 7, Cleanup Acknowledgement из Task 3 и committed capabilities из Task 5.

**Out of scope:** создание dormant launch path, выбор production Dispatcher, composition Runtime, shutdown Listener и orchestration Manager Wait.

### Task 9: Transactional Dispatcher and Dormant Handoff

**Objective:** ввести domain-specific `CommitHandoff`, выполнить prerequisite correction Commit input из Task 5 и создать и полностью протестировать полную pre-Commit transaction, dormant launch path, transactional Dispatcher и boundary accepted result.

**Architectural invariant:** ровно один `CommitHandoff`, один underlying Execution Binding и один dormant launch path принадлежат одной pre-Commit transaction Dispatcher. Dispatcher может публиковать только `NotCommitted`; Manager может публиковать только `Committed` с полным `CommitResult`; dormant path может только ожидать. Каждый non-committed path выполняет:

```text
NotCommitted publication
    -> dormant path return
    -> prepared-value disposal
    -> Reservation abort
    -> accepted=false
```

Successful Commit делает eligible ровно один execution path, доставляет тому path и caller Commit тот же logical `CommitResult` и необратимо передаёт ownership. Post-Commit activation или capability-delivery step отсутствует.

**Dependencies:** Tasks 5 и 8, завершённый pre-Commit Session bundle и prerequisite correction Commit input Task 5, выполняемая как первый focused sub-step Task 9.

**Out of scope:** выбор production composition Runtime, shutdown ordering Host, orchestration shutdown Listener и production Manager Wait.

Полный transaction-capable Dispatcher может временно существовать вне production path Runtime. Он должен быть полным и полностью протестированным; это не частично активная альтернативная ownership model.

### Task 10: Atomic Runtime Composition and Shutdown Cutover

**Objective:** выбрать уже полный transaction-capable Dispatcher и path `CommitHandoff` в production composition и активировать правдивое Runtime-wide shutdown accounting в одном production cutover.

**Architectural invariant:** каждая production accepted Session отслеживается Manager от Commit до completion и release Lifetime Lease. Shutdown следует порядку:

```text
close Admission
    -> BeginShutdown
    -> Snapshot RequestStop
    -> cancel root Runtime context
    -> Listener Stop
    -> Manager Wait
```

**Dependencies:** Task 9 и frozen contracts Host, Admission Gate, Runtime context, Listener и startup rollback из ARCH-002.

**Out of scope:** изменения утверждённой target architecture, Router, Delivery, Persistence, Plugins, Metrics, diagnostics backends, restart и reload.

## 7. Dependency Graph

```text
Completed:
Task 1: Session Core
    -> Task 2: Provisional Session
    -> Task 3: Cleanup Acknowledgement
    -> Task 4: Pre-Commit Session Bundle
                |
                v
Task 5: Complete Atomic Commit Publication Foundation
                |
                +-------------------+
                |                   |
                v                   v
Task 6: Runtime-Cancellation   Task 7: Terminal Result
and Control-Call Lifecycle     and Observer Contracts
                |                   |
                +---------+---------+
                          |
                          v
Task 8: Complete Execution Owner Terminal Lifecycle
                          |
                          v
Task 9: Transactional Dispatcher and Dormant Handoff
                          |
                          v
Task 10: Atomic Runtime Composition and Shutdown Cutover
```

## 8. Retired Remaining Sequence

Следующие standalone boundaries оставшихся tasks отменены:

- Dormant Execution без integration Binding и ownership Dispatcher;
- оставшийся standalone task Execution Binding;
- Runtime callback как изолированный production integration step;
- migration Dispatcher без ownership и disposal dormant path;
- shutdown Runtime отдельно от production activation.

Историческая информация завершённых tasks остаётся действительной. Существующий primitive Execution Binding остаётся реализованным и протестированным; только его integration переносится в tasks atomic publication и handoff.

## 9. Compatibility and Production Status

Эта revision не вводит Runtime capability и не изменяет production behavior. Текущий production Dispatcher остаётся synchronous, а composition Runtime пока не создаёт и не координирует Session Manager.

DP-003 и DP-004 остаются нормативной target architecture. ARCH-002 остаётся неизменным: production cutover должен сохранить его frozen semantics lifecycle Host, readiness, Admission Gate, Runtime context, startup rollback и ordering Listener.

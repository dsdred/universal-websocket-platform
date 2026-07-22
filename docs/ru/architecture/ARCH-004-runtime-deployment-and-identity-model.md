# ARCH-004: Runtime Deployment and Identity Model

[English version](../../en/architecture/ARCH-004-runtime-deployment-and-identity-model.md)

**Статус:** Active

**Область:** operational identity, ownership, lifecycle и management boundary Runtime в single-node среде

**Стабильность:** Approved architectural model

## 1. Назначение

ARCH-004 определяет operational identity и модель ownership между Published ConfigurationVersion и одним работающим Runtime Host.

Документ закрывает разрыв identity и lifecycle, обнаруженный TASK-M11-DISCOVERY-001. Он фиксирует стабильные понятия, необходимые до того, как Control Service сможет создавать, запускать, останавливать, заменять или наблюдать экземпляры Runtime. Он не определяет implementation API, persistence schema, загрузку Configuration или создание Snapshot.

Эта модель дополняет [ADR-0002](../adr/0002-configuration-dsl.md), [ADR-0003](../adr/0003-runtime-architecture.md), [ARCH-002](ARCH-002-runtime-foundation-freeze.md) и фундамент Runtime, определённый [DP-003](../design/DP-003-runtime-session-manager.md), [DP-004](../design/DP-004-per-session-execution-boundary.md) и [DP-006](../design/DP-006-runtime-production-integration.md). Она не изменяет их Runtime-internal contracts ownership или lifecycle.

## 2. Контекст

В репозитории независимо реализованы две границы:

```text
Configuration
    -> ConfigurationVersion
    -> Published ConfigurationVersion
```

и:

```text
Immutable Runtime Snapshot
    -> Runtime Bootstrap
    -> Runtime Host
```

Production owner перехода между ними отсутствует. ConfigurationVersion не может безопасно быть identity одновременно declarative configuration и operational execution, потому что одна Configuration может запускаться многократно, может иметь несколько независимо управляемых экземпляров Runtime и может переходить между Published versions без изменения identity управляемого сервера.

Runtime Host намеренно ограничен одним execution lifetime. Он владеет ресурсами Runtime, но не владеет Control Plane identity, desired state, persistence, process supervision или выбором Configuration.

## 3. Область

Документ определяет:

- identity Runtime Instance;
- identity Launch Attempt;
- связь между Workspace, Configuration, ConfigurationVersion, Runtime Instance, Launch Attempt и Runtime Host;
- operational ownership до, во время и после запуска;
- desired и actual lifecycle state;
- management boundary для запуска, остановки и замены Runtime execution;
- инварианты concurrency и identity;
- начальную single-node execution topology.

Документ не определяет:

- Go types или публичные HTTP endpoints;
- repository или PostgreSQL schema;
- Configuration Loader или поля immutable Snapshot;
- разрешение Secret;
- authorization management operations;
- automatic restart, failover, scheduling, clustering или federation;
- zero-downtime replacement;
- in-place reload;
- deployment-specific overrides Configuration;
- metrics, хранилище diagnostics или retention.

## 4. Архитектурные понятия

### Configuration

Configuration является стабильным declarative parent внутри одного Workspace. Она владеет историей ConfigurationVersion и не содержит operational state Runtime.

### ConfigurationVersion

ConfigurationVersion является одним immutable declarative payload после публикации. Она является источником поведения Runtime, а не identity Runtime, записью запуска или записью процесса.

### Runtime Instance

Runtime Instance является стабильной operational identity одного независимо управляемого WebSocket-сервера. Он принадлежит ровно одному Workspace и ровно одной Configuration.

Runtime Instance переживает отдельные запуски, остановки, неуспешные запуски и замены. Он содержит или ссылается на operational desired и actual state, но не дублирует payload Configuration.

Удаление Runtime Instance, удаление его истории Launch Attempt и retention policy намеренно находятся вне scope документа. До реализации удаления эти semantics должен определить отдельный сфокусированный Design Proposal.

### Launch Attempt

Launch Attempt является одной immutable execution identity, создаваемой каждый раз, когда lifecycle owner пытается запустить Runtime Instance. Launch Attempt принадлежит ровно одному Runtime Instance и фиксирует ровно одну Published ConfigurationVersion как свой источник.

Неуспешный startup остаётся отдельным Launch Attempt. Restart или replacement создаёт новый Launch Attempt и никогда не переиспользует identity предыдущей попытки.

### Runtime Host

Runtime Host является Runtime-internal владельцем composition и lifecycle одного Launch Attempt. Он владеет копией immutable Snapshot, Listener, Runtime context, Session Manager и component graph, определённым существующей архитектурой Runtime.

Identity Host не является Control Plane identity. Pointer, process address, goroutine, context, socket или PID не должны использоваться как identity Runtime Instance или Launch Attempt.

### Runtime Lifecycle Owner

Runtime Lifecycle Owner является orchestration responsibility на стороне Control Service, которая сериализует management operations одного Runtime Instance, создаёт Launch Attempts, вызывает Runtime Launcher, владеет ссылкой на активный Host и фиксирует правдивый actual state.

Эта ответственность может быть реализована сфокусированными компонентами. Это роль, а не разрешение превратить Control Service или Runtime Host в universal manager.

## 5. Модель identity

Нормативный граф identity:

```text
Workspace
    -> Configuration
        -> Runtime Instance
            -> Launch Attempt
                -> Runtime Host

Configuration
    -> ConfigurationVersion history

Launch Attempt
    -> exactly one Published ConfigurationVersion
```

Действуют следующие правила cardinality:

- один Workspace может владеть многими Configurations;
- одна Configuration может владеть многими ConfigurationVersions;
- одна Configuration может иметь много Runtime Instances;
- Runtime Instance принадлежит одной Configuration весь свой lifetime;
- Runtime Instance может иметь много исторических Launch Attempts;
- Runtime Instance может иметь не более одного активного Launch Attempt;
- Launch Attempt имеет не более одного Runtime Host;
- Launch Attempt фиксирует одну точную Published ConfigurationVersion;
- несколько Runtime Instances могут использовать одну ConfigurationVersion, если их effective Listener resources не конфликтуют.

Runtime Instance ID и Launch Attempt ID являются стабильными opaque identifiers. Их конкретное представление и стратегия выделения являются implementation decisions. PID является необязательной наблюдаемой process metadata и никогда не является domain identity.

Активный Launch Attempt остаётся активным всё время, пока Runtime Lifecycle Owner владеет связанным с ним Runtime Host, включая startup и shutdown. После наблюдения terminal completion и освобождения ссылки Host владельцем Launch Attempt становится историческим. Startup attempt, не создавший owned Host, становится историческим после фиксации failure. Исторический Launch Attempt никогда не становится активным повторно. Эти понятия описывают длительность ownership и не вводят отдельную state machine Launch Attempt.

## 6. Привязка Configuration

Runtime Instance привязан к Configuration, а не навсегда к одной ConfigurationVersion. Это сохраняет стабильную operational identity, пока Configuration развивается через новые immutable versions.

Каждый Launch Attempt фиксирует точную source version, выбранную до construction Runtime. Публикация новой ConfigurationVersion не изменяет существующий Launch Attempt, Snapshot, Host или работающий Runtime.

Начальная архитектура не допускает overrides Runtime Instance для полей Configuration, влияющих на поведение. Host и port Listener остаются частью ConfigurationVersion. Поддержка placement или binding overrides требует отдельного архитектурного решения, иначе такие overrides станут вторым источником поведения Runtime.

## 7. Operational state

Desired state и actual state разделены.

Начальные desired states:

- **Stopped:** Runtime execution не запрошен;
- **Running:** запрошен один Runtime execution.

Добавление другого desired state, включая такие понятия, как Maintenance, Paused или Draining, требует отдельного утверждённого Design Proposal. Реализация не должна выводить дополнительные desired states из actual state или management commands.

Начальные actual states:

- **Stopped:** нет активного Launch Attempt, владеющего ресурсами Runtime;
- **Starting:** один Launch Attempt получает и запускает ресурсы Runtime;
- **Running:** активный Host завершил startup и находится в Ready;
- **Stopping:** lifecycle owner ожидает освобождения ресурсов активным Host;
- **Failed:** последний Launch Attempt не запустился или завершился без выполнения запрошенного lifecycle outcome.

Это operational lifecycle за пределами Runtime Host. Он не изменяет lifecycle Host, замороженный ARCH-002, и не вводит новые состояния в Runtime, Session Manager, Session или Execution Owner.

## 8. Lifecycle transitions

Поддерживаются следующие operational transitions:

```text
Stopped
    -> Starting
    -> Running
    -> Stopping
    -> Stopped

Starting
    -> Failed

Starting
    -> Stopping
    -> Stopped

Running
    -> Failed

Failed
    -> Starting

Failed
    -> Stopped
```

Правила:

- Start запрашивает desired `Running`.
- Start, вызванный когда тот же Runtime Instance уже находится в `Starting` или `Running`, не должен создавать другой Launch Attempt.
- Startup success линеаризуется только после успешного Host Start и открытия readiness Runtime.
- Startup failure не публикует Running и не оставляет active Host без owner.
- Stop во время `Starting` захватывает тот же Launch Attempt, запрещает publication Running и ожидает startup rollback или shutdown Host до publication `Stopped`.
- Stop запрашивает desired `Stopped` и применяется только к текущему активному Launch Attempt.
- Stop completion публикуется только после завершения Host своего shutdown contract.
- Повторные или concurrent Stop должны сходиться на одном active Launch Attempt и не создавать второго shutdown owner.
- Restart не является операцией Host. Если он будет представлен позже, это orchestration Stop с последующим новым Launch Attempt.
- Replacement не является in-place reload. Он требует новый Launch Attempt с новым immutable Snapshot.

Automatic restart после failure не входит в начальную модель. Будущая policy может реагировать на расхождение desired/actual только после утверждения semantics retry, backoff и failure.

## 9. Модель ownership

| Объект или state | Создатель | Owner до запуска | Owner во время активности | Terminal responsibility |
|---|---|---|---|---|
| Workspace | Control Plane | Control Plane | Control Plane | Persistence policy Control Plane |
| Configuration | Control Plane | Control Plane | Control Plane | Persistence policy Control Plane |
| ConfigurationVersion | Control Plane | Control Plane | Control Plane; заимствуется как immutable source | Lifecycle ConfigurationVersion |
| Identity Runtime Instance | Management boundary Control Plane | Runtime Lifecycle Owner | Runtime Lifecycle Owner | Сохраняется до explicit domain deletion |
| Desired state | Management command boundary | Runtime Lifecycle Owner | Runtime Lifecycle Owner | Сохраняется как operational intent |
| Identity Launch Attempt | Runtime Lifecycle Owner | Runtime Lifecycle Owner | Runtime Lifecycle Owner | Сохраняется как launch history согласно будущей persistence policy |
| Immutable Snapshot | Runtime boundary | Launch preparation | Runtime Host | Value lifetime завершается вместе с Host |
| Runtime Host | Runtime Launcher | Runtime Lifecycle Owner во время construction | Runtime Host владеет ресурсами Runtime; Lifecycle Owner владеет ссылкой Host | Lifecycle Owner ожидает Host completion и освобождает ссылку |
| Actual state | Runtime Lifecycle Owner | Runtime Lifecycle Owner | Runtime Lifecycle Owner на основании наблюдаемых outcomes Host | Должен оставаться правдивым после success или failure |
| PID или worker metadata | Execution adapter, если применимо | Отсутствует | Наблюдается Runtime Lifecycle Owner | Очищается или сохраняется как non-identity history |

Configuration services никогда не владеют ресурсами Runtime. Runtime Host никогда не владеет identity Runtime Instance, desired state, repository state или publication Configuration.

## 10. Management boundary

Management operations адресуются Runtime Instance ID. Они никогда не адресуются pointer Host, PID, адресом Listener, Session или ConfigurationVersion как execution identity.

Management boundary отвечает за:

- проверку ownership Workspace и Configuration;
- сериализацию state-changing operations одного Runtime Instance;
- загрузку выбранной Published ConfigurationVersion через будущую boundary Loader;
- создание одного Launch Attempt;
- запрос Runtime Launcher на construction и startup одного Host;
- publication actual state из наблюдаемых outcomes;
- остановку только активного Host;
- сохранение identity и state при retry caller.

Runtime Launcher принимает подготовленный launch input и возвращает lifecycle ownership одного Host или startup failure. Он не выбирает Configuration, не изменяет desired state, не читает management repositories и не определяет retry policy.

Runtime Launcher является stateless construction boundary. Он не владеет Runtime Instance, не хранит lifecycle state, не содержит registry Runtime, не принимает management decisions и не становится вторым Runtime Lifecycle Owner.

Точные commands, HTTP resources, error contracts, authorization и repository mutations требуют сфокусированных Design Proposals.

## 11. Начальная execution topology

Начальная Beta topology является single-node и in-process: Control Service владеет Runtime Lifecycle Owner, который запускает Runtime Host через explicit Runtime Launcher boundary.

Модель identity не приравнивает in-process execution к domain identity. Будущий утверждённый execution adapter может разместить Runtime Host в child process без изменения identity Runtime Instance или Launch Attempt.

Domain identity полностью независима от execution topology: переход между single-process и multi-process execution не изменяет модель Runtime Instance или Launch Attempt.

Separate process supervision, persistence PID, crash recovery после restart Control Service, remote workers, scheduling и clustering находятся вне scope. Их нельзя имитировать через скрытый state в Runtime Host.

## 12. Concurrency и linearization

Все state-changing management operations одного Runtime Instance должны использовать одну serialization boundary.

Архитектура требует отдельные linearization points для:

- **Launch claim:** создание единственного active Launch Attempt из состояния, допускающего Start;
- **Running publication:** наблюдение успешного startup Host и readiness;
- **Stop claim:** передача shutdown responsibility активного Launch Attempt одной lifecycle operation;
- **Stopped publication:** наблюдение завершения shutdown Host и отсутствия активных ресурсов;
- **Failure publication:** фиксация невозможности Launch Attempt стать или оставаться Running.

Ни один observer не может наблюдать:

- два активных Launch Attempts одного Runtime Instance;
- actual `Running` до readiness Host;
- actual `Stopped`, пока owned Host всё ещё удерживает ресурсы;
- active Host без identity Launch Attempt;
- один Launch Attempt, связанный с разными ConfigurationVersions;
- повторно использованную identity Launch Attempt после Stop или failure.

Operations разных Runtime Instances могут выполняться независимо с учётом внешних resource conflicts, таких как ownership адреса Listener.

## 13. Publication, replacement и rollback

Publication новой ConfigurationVersion изменяет только state Configuration. Она автоматически не:

- запускает Runtime Instance;
- останавливает Runtime Instance;
- изменяет работающий Snapshot;
- заменяет Host;
- изменяет source identity активного Launch Attempt.

Configuration rollback и Runtime replacement являются разными ответственностями:

- Configuration rollback выбирает immutable version для будущей launch preparation;
- Runtime replacement создаёт новый Launch Attempt из явно выбранной version;
- существующий Launch Attempt продолжает использовать зафиксированный Snapshot до остановки Host.

Ordering, availability и failure semantics replacement и rollback требуют сфокусированного DP. Эта модель не допускает in-place reload.

## 14. Границы failure и recovery

Неуспешный start создаёт failed Launch Attempt и правдивый actual `Failed`. Он не должен публиковать readiness Runtime или сохранять ресурсы Host без owner.

Неожиданное завершение активного Runtime создаёт actual `Failed` после наблюдения Runtime-owned cleanup. Оно не создаёт replacement Launch Attempt молча.

Cancellation caller ограничивает только ожидание caller, когда это определяет сфокусированный command contract. Она не должна передавать ownership активного Host неотслеживаемой goroutine или позволять actual state заявлять о cleanup, которого не произошло.

Recovery после завершения process Control Service зависит от будущего persistent state Runtime Instance и Launch Attempt и утверждённого reconciliation contract. Документ не выводит существование работающего Host из stale PID data.

## 15. Security и граница Workspace

Runtime Instance принадлежит тому же Workspace, что и его Configuration. Management operation не должна перепривязывать существующий Runtime Instance к Configuration другого Workspace.

Operational identity содержит references и lifecycle state, а не Secret values. Записи Runtime Instance и Launch Attempt не должны становиться альтернативным хранилищем Configuration или credentials.

Authorization policy для создания или управления Runtime Instances находится вне scope, но должна применяться на management boundary до lifecycle mutation.

## 16. Совместимость с существующей архитектурой Runtime

ARCH-004 не изменяет:

- Published ConfigurationVersion как источник поведения;
- ownership immutable Runtime Snapshot;
- Runtime Host как composition root одного execution lifetime;
- startup, readiness, Admission Gate, rollback или shutdown Host;
- Runtime Context;
- Transactional Dispatcher;
- accounting Session Manager;
- lifecycle Execution Owner;
- contracts Listener, Router, Authentication или Secret Resolver.

В текущем репозитории отсутствуют production Runtime Lifecycle Owner, integration Runtime Launcher, repository Runtime Instance, management API и process supervision. Документ определяет только архитектуру и не объявляет эти capabilities реализованными.

## 17. Обязательные инварианты

1. Runtime Instance является operational identity, а ConfigurationVersion является declarative identity.
2. Runtime Instance постоянно принадлежит одному Workspace и одной Configuration.
3. Runtime Instance имеет не более одного активного Launch Attempt.
4. Каждый Launch Attempt фиксирует ровно одну Published ConfigurationVersion.
5. Каждый Runtime Host принадлежит ровно одному Launch Attempt.
6. Start, Stop и будущие replacement operations адресуются Runtime Instance ID.
7. PID, socket address, pointer Host, context, goroutine и Session ID не являются identity Runtime.
8. Actual `Running` невозможен до readiness Host.
9. Actual `Stopped` невозможен до освобождения всех owned Host ресурсов.
10. Publication не изменяет и не перезапускает активный Runtime.
11. Restart и replacement создают новый Launch Attempt и Snapshot.
12. Runtime Host не читает repositories Control Plane и не изменяет desired state.
13. Control Plane не получает ownership Runtime-internal lifecycle Session или execution.
14. Operational state не становится hidden Configuration.
15. Behavior-affecting deployment override не существует без отдельного утверждённого решения.

## 18. Последствия

### Преимущества

- Runtime execution получает стабильную operational identity без искажения identity Configuration.
- Retry Start, restart, replacement и failure становятся различимыми.
- Одна Configuration может поддерживать независимо управляемые Runtime Instances.
- Desired и actual state могут оставаться правдивыми, не входя в Runtime Host.
- Provenance Snapshot может ссылаться на конкретный Launch Attempt.
- Будущий process adapter не требует изменения domain identity.

### Издержки

- Control Service требуется новая domain model и lifecycle owner.
- Production persistence должна сохранять инварианты identity и state transitions.
- Management concurrency требует explicit serialization и idempotency.
- Replacement, rollback и recovery требуют сфокусированных contracts вместо implicit side effects.

## 19. Последующая архитектура

Реализация не должна начинаться до того, как focused design определит минимум:

1. Configuration Loader, provenance Snapshot и schema compatibility;
2. persistence contracts Runtime Instance и Launch Attempt;
3. contracts management commands и idempotency;
4. ordering activation, replacement и rollback;
5. recovery и reconciliation после завершения Control Service;
6. operational error reporting и redaction.

Process isolation, automatic restart, scheduling и clustering остаются более поздними решениями и не являются prerequisites начальной in-process single-node реализации.

## 20. Решение

UWP принимает отдельный Runtime Instance как стабильную operational identity между Configuration и execution Runtime.

Runtime Instance принадлежит одному Workspace и одной Configuration. Каждый start создаёт новый immutable Launch Attempt, который фиксирует одну Published ConfigurationVersion и владеет не более чем одним execution Runtime Host. Runtime Lifecycle Owner на стороне Control Service сериализует management operations, владеет desired и actual state и удерживает ссылку на active Host. Runtime Host остаётся существующим владельцем composition и ресурсов одного execution lifetime и не получает ответственности Control Plane.

Начальная topology является single-node и in-process за explicit Runtime Launcher boundary. Publication, restart, replacement, rollback и process recovery никогда не изменяют active Snapshot и требуют explicit lifecycle operations или последующих focused designs.

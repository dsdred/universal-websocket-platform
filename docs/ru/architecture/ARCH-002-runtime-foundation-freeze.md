# ARCH-002: Runtime Foundation Freeze

[English version](../../en/architecture/ARCH-002-runtime-foundation-freeze.md)

**Статус:** Active

**Область:** реализованный фундамент Runtime

**Стабильность:** Frozen

## 1. Purpose

ARCH-002 фиксирует, что реализованный фундамент Runtime считается архитектурно стабильным. Документ не предлагает новую архитектуру. Он определяет компоненты и инварианты, которые уже реализованы, подтверждены тестами и могут считаться стабильной основой для следующей Beta-работы.

Документ дополняет [ARCH-001](ARCH-001-runtime-architectural-pattern.md), [DP-001](../design/DP-001-runtime-handshake-pipeline.md), [DP-002](../design/DP-002-runtime-host-composition-root.md), [Master Engineering Plan](../roadmap/MASTER_PLAN.md) и зафиксированное [текущее состояние](../../../spec/current-state.md). Он замораживает только реализованную часть этих designs.

## 2. Scope

Freeze охватывает композицию Runtime Host и lifecycle-инфраструктуру:

- Runtime Host;
- production composition root;
- координацию lifecycle Host;
- корневой Runtime context;
- startup transaction и rollback;
- Runtime readiness;
- Admission Gate, зависящий только от lifecycle.

Freeze распространяется на архитектурные обязанности, ownership, порядок операций и наблюдаемую lifecycle-семантику. Он не замораживает каждую private-деталь реализации и не требует превращать текущие private-типы в публичные контракты.

## 3. Frozen Runtime Components

### Runtime Host

Runtime Host владеет независимой копией Snapshot, своим Container, ссылкой на собранный Listener, Runtime context, lifecycle-состоянием, readiness и Admission Gate. Он координирует создание, запуск, rollback и остановку, не забирая бизнес-логику Authentication, Session, Message или transport.

### Composition Root

Runtime Bootstrap создает и подготавливает Host. Host является production composition root для одного экземпляра Runtime. Во время Start явные конструкторы собирают поддерживаемый граф Authentication, connection dispatch, Session handoff, Message Handler и Listener. Service locator, reflection, DI framework и generic component registry не используются.

Замороженным свойством является наличие единственного явного production composition root и направление его зависимостей. Будущий порядок Handshake и будущий граф подсистем этим документом не замораживаются.

### Lifecycle

Host владеет единым потокобезопасным lifecycle и не поддерживает restart или in-place reload. Длительные операции Listener выполняются вне lifecycle mutex Host.

### Runtime Context

Host владеет одним Runtime context успешно запущенного экземпляра. Вызывающий код может его наблюдать, но не получает функцию cancellation.

### Startup Transaction and Rollback

Start получает собранный Listener как startup resource, запускает его, выполняет commit только после успеха и откатывает полученные ресурсы при ошибке. Ошибки startup и rollback остаются различимыми через обычные механизмы wrapping и joining ошибок Go.

### Runtime Readiness

Readiness является lifecycle-фактом Host. Она становится true только после startup commit и перехода Host в `Running`.

### Admission Gate

Host владеет небольшим потокобезопасным Admission Gate, зависящим только от lifecycle. Он отвечает только на вопрос, разрешает ли текущий lifecycle Host новый admission. В нем нет Authentication, Origin, rate-limit, maintenance или configuration policy.

## 4. Frozen Architectural Invariants

Замораживаются следующие реализованные инварианты:

- Host является production composition root и lifecycle coordinator, а не сервисом бизнес-логики.
- Зависимости соединяются явно через конструкторы и сфокусированные Bootstrap-компоненты.
- Container остается хранилищем Snapshot, а не service locator.
- Host хранит и возвращает независимые копии Snapshot.
- `Build` подготавливает lifecycle-состояние без открытия сетевых ресурсов.
- Runtime dependencies собираются во время `Start`; Listener является внешне наблюдаемым компонентом, запускаемым Host.
- Startup публикует Runtime resources только после успешного запуска Listener.
- Ошибка startup не оставляет Host в Running, Ready или с открытым admission.
- Полученные startup resources откатываются после ошибки запуска.
- Shutdown сбрасывает readiness в false и закрывает admission до потенциально длительной остановки Listener.
- Lifecycle-методы и state accessors Host безопасны для конкурентного использования.
- Restart и in-place reload экземпляра Host не поддерживаются.

Эти инварианты описывают существующее поведение. Они не определяют будущие контракты Handshake, Session Manager, diagnostics или supervision.

## 5. Guaranteed Runtime Lifecycle

Реализованный lifecycle Host:

```text
Created
    -> Built
    -> Starting
    -> Running
    -> Stopping
    -> Stopped
```

- `Build` выполняет однократный переход `Created -> Built`.
- `Start` допустим только из `Built`.
- При ошибке startup Host после rollback возвращается в `Built`, позволяя повторить запуск после исправления причины.
- Успешный startup ровно один раз фиксирует состояние `Running`.
- `Stop` из `Running` или конкурентно со `Starting` координирует единственную операцию shutdown.
- Конкурентные и повторные вызовы Stop наблюдают одинаковый итоговый результат shutdown.
- Stop до Start является no-op и не запрещает первый Start.
- `Stopped` является terminal-состоянием; restart не поддерживается.

Состояния `Initialized` и `Failed` не входят в замороженную реализацию. Они остаются предложением DP-002, а не гарантией этого документа.

## 6. Guaranteed Startup Semantics

Startup имеет следующую реализованную семантику:

1. `NewHost` проверяет обязательные входы Host и сохраняет независимый Snapshot через Container.
2. `Build` отмечает Host как подготовленный, не получая и не запуская Runtime components.
3. `Start` переходит в `Starting` и сохраняет readiness и admission закрытыми.
4. Поддерживаемый component graph собирается явно.
5. Listener создается и запускается.
6. При ошибке composition или запуска Listener полученные ресурсы откатываются, а исходная ошибка сохраняется.
7. После успешного запуска Listener Host публикует Listener, создает Runtime context, переходит в `Running`, открывает admission и становится Ready в рамках одного lifecycle commit.

Context, переданный в `Start`, ограничивает startup-работу, которая его учитывает. Он не становится lifecycle context успешно запущенного Runtime.

## 7. Guaranteed Runtime Context Semantics

- `RuntimeContext()` равен nil до успешного startup commit.
- Host создает Runtime context только после успешного запуска Listener.
- Runtime context не зависит от context, переданного в `Start`.
- Cancellation startup context после успешного Start не останавливает Runtime.
- Host владеет Runtime cancellation function и никогда ее не раскрывает.
- Stop отменяет Runtime context до вызова Listener Stop.
- После Stop остается доступен тот же отмененный context.
- Повторный Stop не создает и не отменяет другой Runtime context.

Этот freeze не определяет будущие дочерние contexts для Handshake или Session.

## 8. Guaranteed Readiness Semantics

`Ready()` возвращает true ровно тогда, когда Host находится в `Running`.

Readiness равна false:

- после `NewHost` и `Build`;
- на всем протяжении `Starting`;
- при ошибке composition, ошибке запуска Listener и во время rollback;
- с момента начала Stop;
- на всем протяжении `Stopping` и после `Stopped`.

Конкурентные readers безопасно наблюдают lifecycle boundary. Readiness пока не означает health зависимостей, traffic probes, supervision или endpoint здоровья Control Service.

## 9. Guaranteed Admission Gate Semantics

`CanAccept()` возвращает boolean lifecycle decision. Admission открыт только тогда, когда Host одновременно находится в `Running` и Ready.

Gate закрыт:

- после создания Host и Build;
- на всем протяжении Starting;
- при ошибке composition и во время rollback;
- до startup commit;
- в начале Stop, до вызова Listener Stop;
- на всем протяжении длительного Listener Stop;
- после Stop и после отклоненного restart.

Gate не создает goroutine и не выполняет network I/O. Сейчас он является lifecycle boundary, принадлежащим Host. Его атомарное применение внутри будущего pre-Upgrade Handshake commit остается частью открытой архитектуры Handshake.

## 10. Architecture Freeze

Runtime Host, граница production composition root, реализованный lifecycle, Runtime context, startup transaction и rollback, readiness и lifecycle-only Admission Gate считаются стабильными.

Их архитектурные обязанности, ownership и lifecycle-семантика могут изменяться только через новый сфокусированный Design Proposal или новое Architecture Decision. Обычные bug fixes, усиление тестов и внутренний refactoring допустимы, если они сохраняют эти замороженные инварианты и наблюдаемую семантику.

Freeze не является заявлением о production readiness и не переводит в реализованную архитектуру содержимое Draft DP, которое еще не реализовано.

## 11. Open Architecture

Следующие области остаются открытыми и не замораживаются ARCH-002:

- Handshake;
- Authentication Pipeline, включая pre-Upgrade Authentication;
- ownership Session на границе Handshake handoff и ее учет в Runtime shutdown;
- Router;
- Delivery;
- Persistence;
- Operational Diagnostics;
- supervision Runtime.

Этим областям требуются сфокусированный design и подтверждение реализацией. ARCH-002 не определяет их API или внутренние component models.

## 12. Explicitly Out of Scope

Документ не замораживает и не проектирует:

- Handshake или WebSocket admission commit;
- Router или route configuration;
- Delivery guarantees, addressing, groups или topics;
- Persistence или database contracts;
- Plugin ABI или plugin lifecycle;
- Session Manager;
- выполнение TLS и применение Listener timeouts;
- operational diagnostics, monitoring или supervision;
- reload, restart, failover или замену Runtime на уровне процесса.

Он не вводит новые API, interfaces, lifecycle states, configuration fields или требования Runtime.

## 13. Change Policy

Предлагаемое изменение сначала должно быть классифицировано:

- совместимое исправление реализации может выполняться с тестами, сохраняющими этот freeze;
- добавление подсистемы вне замороженного scope требует собственного сфокусированного design, если ее контракты еще не стабильны;
- изменение замороженной ответственности, ownership boundary, startup commit, rollback, Runtime context, readiness, Admission Gate или lifecycle-семантики требует нового DP или ADR до реализации.

Сам ARCH-002 следует обновлять только для отражения явно принятого архитектурного изменения или более сильного подтверждения реализацией. Его нельзя использовать как обход review через DP или ADR.

## 14. Next Beta Phase

Замороженный фундамент Runtime является основой следующей Beta-фазы. Дальнейшая работа переходит в открытую архитектуру, начиная с границы Handshake, описанной в Draft DP-001, и ее интеграции с lifecycle и admission Runtime.

Router, Delivery, Persistence, Plugins и Session Manager остаются более поздними сфокусированными эпиками [Master Engineering Plan](../roadmap/MASTER_PLAN.md). Их будущие designs должны опираться на замороженный фундамент, не превращая Host в компонент бизнес-логики, а Container — в service locator.

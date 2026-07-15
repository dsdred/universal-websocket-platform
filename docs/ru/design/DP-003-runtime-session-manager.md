# DP-003: Runtime Session Manager

[English version](../../en/design/DP-003-runtime-session-manager.md)

## 1. Статус

**Статус:** Draft

## 2. Постановка проблемы

Runtime умеет создавать и запускать Session после успешного Handshake, но в нем нет компонента, владеющего множеством живых Session. Сейчас Session выполняется синхронно в пути handoff и не регистрируется в общем Runtime shutdown wait set. Поэтому Runtime не может в едином месте определить, активна ли Session, исключить выход принятого соединения из shutdown tracking или дождаться завершения всех переданных Session.

[DP-001](DP-001-runtime-handshake-pipeline.md) требует, чтобы Session вошла в Runtime shutdown wait set до завершения ownership handoff. [ARCH-002](../architecture/ARCH-002-runtime-foundation-freeze.md) намеренно оставляет ownership Session в этом wait set открытым вопросом. Этот документ определяет недостающую границу, не изменяя замороженные Runtime Host, Admission Gate, Runtime context и startup transaction.

## 3. Цели

- Определить единственного lifecycle owner для множества Session одного экземпляра Runtime.
- Сделать регистрацию точной границей, на которой Session становится видимой Runtime и входит в shutdown wait set.
- Сделать удаление детерминированным и выполняемым ровно один раз.
- Поддержать lookup по минимально необходимому стабильному идентификатору.
- Координировать graceful shutdown Runtime, не отбирая у Session ownership транспорта.
- Сохранить явный ownership, ограниченный lifecycle и обычный dependency injection.

## 4. Не-цели

Документ не проектирует:

- Router или policy маршрутизации сообщений;
- Delivery, очереди, повторные попытки или backpressure;
- Persistence или восстановление Session;
- Presence, Groups, Topics или Broadcast;
- cluster membership, federation или распределенный ownership Session;
- Plugin SDK или Plugin ABI;
- Metrics или framework операционной диагностики;
- публичный Control Plane или HTTP API;
- конкретные Go-интерфейсы или структуры хранения.

Эти системы смогут интегрироваться с lifecycle Session в будущем, но они не входят в Session Manager.

## 5. Ответственность

Session Manager отвечает только за:

- принятие lifecycle ownership успешно созданной Session;
- регистрацию и дерегистрацию этой Session;
- ведение авторитетного множества зарегистрированных Session одного экземпляра Runtime;
- точный lookup зарегистрированной Session по стабильному SessionID;
- координацию закрытия и завершения Session при shutdown Runtime;
- соблюдение определенных здесь инвариантов Runtime Session lifecycle.

Session Manager не аутентифицирует клиентов, не принимает WebSocket, не читает и не отправляет сообщения, не маршрутизирует трафик, не сохраняет состояние, не оценивает policy и не владеет корневым Runtime context.

## 6. Модель владения

Цепочка ownership выглядит так:

```text
Runtime Host
    -> Session Manager
        -> Session
            -> WebSocket connection
```

Цепочка описывает lifecycle authority, а не совместный ownership транспорта:

- Runtime Host собирает Session Manager и вызывает его shutdown coordination. Host не хранит и не закрывает отдельные Session.
- Session Manager владеет авторитетным membership и completion tracking зарегистрированных Session. Он может запросить остановку Session, но не закрывает WebSocket напрямую.
- Session единолично владеет своим WebSocket после handoff, определенного DP-001. Session выполняет фактическое закрытие транспорта и освобождает собственные ресурсы соединения.

До регистрации Upgrade boundary владеет кандидатом Session, derived connection context и WebSocket. Успешная регистрация является единственной атомарной концептуальной linearization point: Session входит в registry Manager и shutdown wait set, а Manager принимает lifecycle ownership. Только после этого Upgrade boundary может сообщить об успешном handoff.

Если регистрация не удалась, ownership не передается и Upgrade boundary закрывает соединение. После успешной регистрации верхний слой не должен закрывать WebSocket или дерегистрировать Session. Завершение Session проходит через принадлежащий Manager путь удаления.

## 7. Runtime Session Lifecycle

Концептуальный lifecycle:

```text
Created
    -> Registering
    -> Registered
    -> Running
    -> Closing
    -> Closed
    -> Removed
```

Это архитектурные состояния, а не обязательный Go enum.

- **Created:** Upgrade boundary создал кандидата Session и еще владеет им.
- **Registering:** Manager пытается выполнить атомарный переход ownership и membership. Это состояние не видно через lookup.
- **Registered:** ownership передан, Session видна Runtime через lookup и входит в shutdown wait set.
- **Running:** Session исполняет lifecycle соединения и сообщений. Она остается зарегистрированной.
- **Closing:** Session завершается из-за закрытия peer, ошибки, собственного lifecycle или shutdown Runtime. Она остается в wait set.
- **Closed:** транспорт Session и работа соединения завершены. Session остается в tracking только до окончания удаления.
- **Removed:** Manager удалил Session ровно один раз. Она больше не видна и не входит в wait set.

Ошибка после успешной регистрации не обходит `Closing`, `Closed` или `Removed`. Ошибка до регистрации оставляет cleanup Upgrade boundary. Restart или повторная регистрация удаленной Session не поддерживаются.

## 8. Регистрация

Регистрация — единственная операция, после которой Session существует в области управления Runtime. Она должна атомарно установить все следующие свойства:

- SessionID допустим и еще не зарегистрирован;
- Session становится видимой через lookup;
- Session входит в Runtime shutdown wait set;
- Session Manager принимает lifecycle ownership;
- конкурентный shutdown либо увидит Session, либо отклонит регистрацию.

Успешный handoff не может ссылаться на незарегистрированную Session. Регистрация принимается не более одного раза для SessionID за время жизни одного Manager. Регистрация после закрытия admission Manager для shutdown отклоняется; ответственность за cleanup остается у Upgrade boundary.

Регистрация не должна начинать сетевой I/O под synchronization Manager. Manager не должен вызывать произвольное поведение Session под lock registry.

## 9. Удаление

Удаление происходит только после закрытия транспорта Session и завершения ее выполнения. Manager владеет переходом удаления, даже если завершение вызвано peer или ошибкой Session.

Удаление должно:

- выполняться ровно один раз;
- атомарно убирать всю видимость через lookup;
- удалять Session из shutdown wait set;
- пробуждать shutdown coordination, ожидающую опустошения множества;
- сохранять исходный результат завершения Session для будущей границы диагностики, не превращая registry в эту границу.

Session не должна напрямую изменять хранилище Manager. Принадлежащий Manager путь выполнения и завершения наблюдает terminal completion Session и выполняет удаление. Повторные уведомления о завершении безопасны и не могут дважды уменьшить tracking. Для первой реализации tombstone не требуется; после удаления lookup сообщает об отсутствии.

## 10. Координация shutdown

Session Manager участвует в shutdown Runtime, но не владеет им. Замороженный Host остается lifecycle coordinator и владеет закрытием Admission Gate и cancellation корневого Runtime context.

Концептуальная последовательность shutdown:

```text
Host закрывает Admission Gate
    -> новый Handshake admission commit не может начаться
    -> Session Manager закрывает регистрацию
    -> незавершенная регистрация становится принятой и отслеживаемой либо отклоняется
    -> cancellation Runtime context достигает Session
    -> Manager запрашивает закрытие всех зарегистрированных Session
    -> Manager ожидает, пока каждая зарегистрированная Session станет Removed
    -> Runtime Stop может завершиться
```

Shutdown wait set в точности равен множеству Session, для которых регистрация успешна, а удаление еще не завершено. Session входит в множество в linearization point регистрации до успешного handoff и выходит только при удалении.

Закрытие регистрации Manager и регистрация Session должны иметь взаимно однозначный порядок. Регистрация, линеаризованная до закрытия, включается в tracking и ожидание. Регистрация, линеаризованная после закрытия, завершается ошибкой, а ownership остается у Upgrade boundary. Это правило закрывает разрыв между финальной проверкой Admission Gate и ownership handoff, не перенося Gate в Session Manager.

Ожидание shutdown должно учитывать границу shutdown caller, когда соответствующая policy будет определена. Этот документ не выбирает forced-close escalation или timeout policy. Независимо от результата timeout Manager не должен ложно объявлять неудаленную Session удаленной.

## 11. Lookup

Обязательный первый ключ lookup — **SessionID**. Он стабилен, уникален внутри одного экземпляра Runtime, уже относится к metadata Session и не вводит семантику identity или routing.

Lookup видит только зарегистрированные и еще не удаленные Session. Он не должен раскрывать изменяемый registry Manager или передавать ownership Session. Caller получает только capability, необходимую будущей интеграции, и не может дерегистрировать Session или закрывать транспорт в обход lifecycle Manager.

Следующие ключи отложены:

- **ConnectionID:** отдельное стабильное значение в Runtime пока не определено.
- **Principal ID:** один Principal может владеть несколькими Session, а anonymous Session требуют явной семантики.
- **UserID:** домен User отсутствует в Runtime-контрактах.
- перечисление и secondary indexes: требования зависят от Delivery, Presence и операционных use cases.

Эти отложенные пути lookup потребуют сфокусированного проектирования при появлении реальных потребителей.

## 12. Runtime Context

Session Manager не создает, не заменяет и не отменяет корневой Runtime context. Host остается его единственным владельцем согласно [ADR-0004](../adr/0004-handshake-runtime-dependencies.md).

Upgrade boundary создает connection context, производный от активного Runtime context. До регистрации его cancellation принадлежит Upgrade boundary. При успешной регистрации и ownership handoff cancellation соединения становится частью lifecycle Session. Manager отслеживает и координирует этот lifecycle, но никогда не получает root cancellation function.

Cancellation корневого Runtime context сигнализирует каждому производному Session context при shutdown. Manager все равно ожидает фактического завершения и удаления Session; cancellation context сама по себе не доказывает завершение cleanup.

## 13. Архитектурные инварианты

- Каждая переданная Session зарегистрирована ровно в одном Session Manager.
- Каждая зарегистрированная Session входит в Runtime shutdown wait set.
- Успешная регистрация является единственной точкой передачи lifecycle ownership.
- После передачи Session является единственным владельцем и closer своего WebSocket.
- SessionID регистрируется не более одного раза за время жизни одного Manager.
- Удаление выполняется ровно один раз и только после завершения Session.
- Lookup никогда не возвращает удаленную или еще не зарегистрированную Session.
- Закрытие регистрации атомарно упорядочено с конкурентной регистрацией.
- Shutdown не может завершиться, пока shutdown wait set не пуст.
- Manager не удерживает synchronization registry при вызове сетевой или lifecycle-работы Session.
- Manager не владеет Authentication, routing, message handling или cancellation корневого Runtime context.
- Host не становится registry Session и не закрывает транспорт отдельных Session.
- Ошибка не может незаметно оставить зарегистрированную Session вне lookup и shutdown tracking одновременно.

## 14. Будущая интеграция

Будущие системы интегрируются на границе Session Manager без изменения ownership:

- **Router** может адресовать текущую зарегистрированную Session, но не владеет ее lifecycle.
- **Delivery** может разрешать адресата через будущий контракт lookup или index; семантика очередей и retry остается отдельной.
- **Presence** может выводить состояние из событий регистрации и удаления; Manager не хранит Presence.
- **Groups** могут индексировать SessionID, не владея Session или транспортом.
- **Persistence** может сохранять durable metadata или события; сохраненная запись никогда не становится owner живой Session.

Раздел только называет точки интеграции. Он не проектирует эти подсистемы или их API.

## 15. Альтернативы

### Альтернатива A — Без Session Manager

**Преимущества:** сохраняется небольшой текущий путь handoff и не добавляется компонент.

**Недостатки:** отсутствуют авторитетное множество живых Session, lookup, exactly-once removal и полный shutdown wait set. Требования handoff из DP-001 остаются невыполненными.

**Решение:** отклонено, потому что Runtime не может доказать ownership или завершение shutdown.

### Альтернатива B — Runtime Host напрямую владеет Session

**Преимущества:** меньше верхнеуровневых компонентов и прямой доступ при Stop.

**Недостатки:** Host становится изменяемым registry Session и получает lifecycle-логику отдельных соединений. Это противоречит его замороженной роли composition и lifecycle coordination и ведет к god object.

**Решение:** отклонено. Host собирает и координирует Manager.

### Альтернатива C — Listener владеет Session

**Преимущества:** Listener близок к принятым соединениям и shutdown транспорта.

**Недостатки:** HTTP/WebSocket transport связывается с Runtime Session identity, lookup и будущими потребителями. Также размывается граница ownership между Upgrade и Session.

**Решение:** отклонено, поскольку Listener должен оставаться транспортным компонентом.

### Альтернатива D — Распределенный ownership Session

**Преимущества:** может обеспечить межэкземплярный lookup и координацию в cluster.

**Недостатки:** требует membership, failure detection, consistency, remote addressing и partition semantics до стабилизации lifecycle одного экземпляра.

**Решение:** отклонено для текущего Runtime. Cluster и federation отложены за рамки документа.

### Альтернатива E — Раздельные Registry и Manager

**Преимущества:** отделяет indexing от lifecycle coordination и может позже поддержать специализированные indexes.

**Недостатки:** первая реализация разделит атомарный инвариант регистрации и удаления между двумя owners без независимого use case. Возрастут coordination и failure modes.

**Решение:** пока отклонено. Выделение можно пересмотреть после появления нескольких реальных потребителей lookup или indexing.

## 16. Явно вне рамок

- Router, Delivery, очереди, backpressure и persistence;
- Presence, Groups, Topics, Broadcast и fan-out;
- cluster, federation, horizontal scaling и remote lookup;
- Plugin contracts или загрузка extensions;
- Metrics, tracing, logging policy и diagnostics sinks;
- решения Authentication и Authorization;
- транспортный протокол Session и поведение message loop;
- публичный API, дополнения Configuration DSL и database schema;
- конкретные concurrency primitives или сигнатуры Go-методов.

## 17. Стратегия миграции

Миграция выполняется поэтапно и сохраняет текущее поведение:

1. Собрать один Session Manager на экземпляр Runtime через Runtime Host без изменения замороженных lifecycle states Host.
2. Направить успешное создание Session после Upgrade через регистрацию Manager.
3. Сделать успех регистрации точкой принятия handoff из DP-001; при ошибке сохранить cleanup на Upgrade boundary.
4. Выполнять существующий lifecycle Session под completion tracking Manager с exactly-once removal.
5. Включить wait set Manager в shutdown coordination Runtime после закрытия Admission Gate.
6. Добавить lookup по SessionID только после доказательства ownership invariants тестами регистрации, удаления и shutdown.

Ни один шаг миграции не меняет Authentication, транспортное поведение Listener, Message handling или Configuration Snapshot. Каждый шаг должен сохранять единственного closer WebSocket и независимо проверяться при конкурентных handoff и shutdown.

## 18. Открытые вопросы

- Какая ограниченная escalation применяется, если Session не завершается до deadline shutdown caller?
- Как категоризировать ошибку запуска Session после успешной регистрации для операционной диагностики?
- Имеет ли будущий ConnectionID семантику, отличную от SessionID?
- Какую форму capability должен возвращать lookup, не раскрывая lifecycle mutation?
- Какая гарантия consistency потребуется при конкурентном удалении, когда реальным потребителям понадобится перечисление?
- Какой будущий компонент публикует события lifecycle Session, не превращая Manager в diagnostics framework?

Эти вопросы не блокируют согласование ownership, регистрации, удаления и инварианта shutdown wait set. Для них потребуется отдельное implementation design или сфокусированные proposals, когда станут известны потребители.

## 19. Готовность к ревью

Draft готов к архитектурному ревью, когда reviewers могут подтвердить, что:

- регистрация является однозначной linearization point ownership и shutdown tracking;
- каждый успешный handoff из DP-001 либо отслеживается, либо очищается Upgrade boundary;
- Session остается единственным владельцем WebSocket после handoff;
- shutdown не может пропустить конкурентную регистрацию;
- правила удаления и lookup детерминированы при concurrency;
- Manager не поглощает ответственность Host, Listener, Authentication, Router, Delivery или diagnostics;
- миграция возможна без изменения замороженной семантики Runtime Foundation.

Реализация не должна начинаться, пока ownership transfer, порядок shutdown и решение о минимальном lookup не получат архитектурное одобрение.

## 20. Ссылки

- [ARCH-001: Runtime Architectural Pattern](../architecture/ARCH-001-runtime-architectural-pattern.md)
- [ARCH-002: Runtime Foundation Freeze](../architecture/ARCH-002-runtime-foundation-freeze.md)
- [DP-001: Runtime Handshake Pipeline](DP-001-runtime-handshake-pipeline.md)
- [DP-002: Runtime Host Composition Root](DP-002-runtime-host-composition-root.md)
- [ADR-0004: Handshake Runtime Dependency Boundary](../adr/0004-handshake-runtime-dependencies.md)
- [Master Engineering Plan](../roadmap/MASTER_PLAN.md)


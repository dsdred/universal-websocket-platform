# DP-005: Маршрутизатор Runtime-сообщений

[English version](../../en/design/DP-005-runtime-message-router.md)

## 1. Статус

**Статус:** Approved

Этот документ определяет минимальный детерминированный Router для первого Beta-эпика обработки сообщений. Он не определяет Delivery, выбор целевых Session, Persistence или универсальный policy framework.

## 2. Контекст

Текущий Runtime компонует один `message.Handler` и передаёт его каждой Session. Session преобразует каждый текстовый или бинарный WebSocket frame в неизменяемое Runtime Message и синхронно вызывает этот Handler. Echo доказывает этот вертикальный срез, но Published Configuration пока не может выбирать поведение для отдельного сообщения.

[ARCH-001](../architecture/ARCH-001-runtime-architectural-pattern.md) описывает маршрутизацию как `Message Context -> route matchers -> selected Handler or destination -> Handler execution`, но явно оставляет API Router, модель route, контракт matcher и политику ошибок отдельному DP. [Главный инженерный план](../roadmap/MASTER_PLAN.md) размещает Router после четырёх foundation gates Runtime и перед Session Manager, Delivery и Persistence.

Реализация уже содержит необходимую границу: transport-neutral Handler вызывается Session и явно подключается Runtime composition. Runtime Message Context, routing Configuration и Router пока отсутствуют.

## 3. Проблема

Runtime требуется детерминированный выбор одного сконфигурированного Handler без проникновения транспортных деталей, изменяемого состояния Session или репозиториев Control Plane в обработку сообщений. Дизайн должен сохранять текущее поведение Echo при отсутствии routing metadata и различать намеренно пустую routing configuration, невалидную configuration и сообщение, которому не соответствует ни один route.

Без нормативной модели реализация должна была бы самостоятельно решить:

- какие значения Session и Principal доступны маршрутизации;
- принадлежит ли выполнение Handler компоненту Router;
- правила порядка, пересечения, default и no-match;
- способ разрешения ссылок на Handler;
- завершает ли routing outcome работу Session;
- где выполняются проверки Configuration и исполнимости.

## 4. Цели

- Выбирать ровно один Handler или один явный no-match outcome для каждого Runtime Message Context.
- Обеспечить детерминированный выбор для одного неизменяемого Published Snapshot и одного неизменяемого Context.
- Сохранить существующую границу Runtime composition на основе Handler.
- Определить минимальный transport-neutral Context, который может заполнить текущий Session pipeline.
- Определить обратно совместимую и ограниченную routing Configuration.
- Разрешать все ссылки на Handler до перехода Runtime в Ready.
- Оставить выбор read-only, синхронным и безопасным для concurrent Sessions.
- Сохранять `errors.Is` при распространении ошибок через Router и Session.

## 5. Не-цели

Router не выполняет следующие функции:

- доставку сообщений клиентам или выбор Session-получателей;
- управление, регистрацию, поиск, остановку или удержание Session;
- владение goroutine, lifecycle, очередями, retry, backpressure или acknowledgement;
- сохранение сообщений или создание DLQ;
- реализацию Groups, Topics, Broadcast, Presence, Delivery, Plugins, Metrics или rate limiting;
- анализ HTTP request, HTTP headers, WebSocket frames, connections или transport addresses;
- разбор или сопоставление байтов пользовательского payload;
- динамическую регистрацию Handler после запуска Runtime;
- создание универсального matcher, middleware или policy framework.

## 6. Терминология

- **Runtime Message Context:** неизменяемый per-message envelope, создаваемый Session из значений, которыми владеет Runtime.
- **Route:** неизменяемое сконфигурированное правило, объединяющее matchers и называющее одну ссылку на Handler.
- **Matcher:** один предикат равенства над разрешённым routable-полем Context.
- **Compiled Route:** проверенный Route, ссылка которого разрешена в экземпляр Handler.
- **Legacy Handler:** Handler, который передавался Runtime composition до появления Router.
- **Default Handler:** необязательный Handler, выбираемый только тогда, когда ни один включённый явный Route не совпал.
- **No Match:** нормальный routing outcome, при котором Handler не вызывается.
- **Selection:** чистая упорядоченная проверка, возвращающая один compiled Handler или No Match.

## 7. Ограничения текущей архитектуры

- Runtime Host остаётся единственным composition root и не выполняет проверку routes.
- Session остаётся владельцем connection, read loop, lifecycle и outbound-capability `Sender`.
- Router никогда не получает конкретные значения Session, Listener, Handshake, HTTP, WebSocket или Session Manager.
- Payload Runtime Message остаётся во владении неизменяемого `message.Message`.
- Выполнение Handler остаётся синхронным в execution path Session.
- Ошибка Handler сейчас завершает `Session.Run`; Router обязан сохранить этот контракт, а не скрывать ошибку.
- Published Configuration копируется через `runtimeconfig.Builder`; Router не читает репозитории.
- Проверка исполнимости Runtime завершается до захвата Listener socket.
- Identity Session Manager и `RegistrationView` не являются контрактами targeting для Router.

## 8. Runtime Message Context

Дизайн вводит новый неизменяемый Runtime Message Context и не добавляет поля в `message.Message`.

Концептуально Context содержит:

| Поле | Источник | Routable в первой версии | Контракт |
|---|---|---:|---|
| Runtime Message | Session read loop | Только тип сообщения | Существующее неизменяемое text/binary Message; payload и receive time не меняются |
| Sender | Текущая Session | Нет | Transport-neutral outbound-capability, доступная только выбранному Handler |
| Session ID | Текущая Session | Нет | Непрозрачное скопированное значение для корреляции в Handler; не является capability поиска или targeting Session |
| Флаг Authenticated | Неизменяемый Principal Session | Через principal-kind | Ровно один из флагов authenticated и anonymous равен true |
| Флаг Anonymous | Неизменяемый Principal Session | Через principal-kind | Явная anonymous identity остаётся различимой |
| Тип Authentication | Неизменяемый Principal Session | Да | Тип Provider, например `api-key` или `jwt`; отсутствует для anonymous identity |
| Authentication provider | Неизменяемый Principal Session | Да | Безопасное сконфигурированное имя Provider; отсутствует для anonymous identity |

Исходный Context не содержит HTTP headers, query values, remote address, WebSocket value, request, connection, исходную Session, Principal ID или name, карты Principal, claims, roles, attributes, metadata или произвольные пользовательские metadata. Principal представлен только двумя взаимоисключающими kind-флагами и безопасными идентификаторами типа Authentication и Provider. Этот набор нельзя расширять только потому, что данные уже где-то удерживаются: каждая будущая категория identity или routable-данных требует отдельного Configuration и security review.

Служебные metadata и пользовательское содержимое разделены: scalar-поля Context создаются компонентами Runtime, а байты payload остаются только внутри Runtime Message. Matchers никогда не читают байты payload.

При создании все scalar-значения копируются. Accessors возвращают значения, Runtime Message сохраняет существующий контракт копирования payload, а Context не раскрывает изменяемые maps или slices. Sender является стабильной execution-capability, а не routable-данными; Router его не вызывает.

Создание Context успешно только для валидного text или binary Message, ненулевого Sender, непустого opaque Session ID и ровно одного флага Principal kind. Для authenticated Context обязательны и тип Authentication, и Provider; для anonymous Context оба должны отсутствовать. Нарушение этих инвариантов является internal construction error, а не routable no-match value.

Контракт Handler согласованно развивается: он получает Runtime Message Context вместе со стандартным cancellation `context.Context`. Session создаёт ровно один Runtime Message Context для каждого принятого Message. Echo читает Message и Sender из этого Context и в остальном сохраняет текущее поведение. Не существует двойного legacy/context dispatch Handler и не используется `context.Value`.

## 9. Модель routing configuration

В `ConfigurationVersion` появляется необязательная секция Routing. Наличие представляется явно, чтобы отсутствие отличалось от присутствующего пустого объекта.

Концептуальная модель:

```text
Routing
    Routes []Route
    DefaultHandlerRef optional string

Route
    ID string
    Enabled bool
    Priority uint32
    Matchers []Matcher
    HandlerRef string

Matcher
    Type MatcherType
    Value string
```

Первая реализация использует следующие ограничения:

- не более 256 Routes;
- не более одного Matcher каждого поддерживаемого типа на Route;
- следовательно, не более четырёх Matchers на Route;
- Route ID и ссылка на Handler обрезаются по краям, имеют от 1 до 128 ASCII-символов и соответствуют `[A-Za-z][A-Za-z0-9._-]*`;
- Priority каждого Route положителен и уникален среди всех Routes, включая отключённые;
- Route ID уникален среди всех Routes;
- каждый Route, включая отключённый, должен быть структурно валиден;
- включённым Routes требуются непустой список Matcher и активная разрешимая ссылка на Handler;
- отключённые Routes не компилируются и не вычисляются при selection; их синтаксически валидная ссылка на Handler может не разрешаться до более поздней Published Configuration, включающей Route;
- присутствующий `DefaultHandlerRef` должен быть синтаксически валиден и разрешён до readiness;
- два включённых Route с идентичными нормализованными наборами Matcher невалидны независимо от Priority, поскольку Route с меньшим precedence, проверяемый позже в порядке возрастания Priority, иначе никогда не мог бы быть выбран;
- неподдерживаемые типы Matcher невалидны, в том числе в отключённых Routes, поскольку схема не содержит контракта сохранения opaque extensions.

Пустой список Matcher никогда не означает catch-all Route. Catch-all поведение представляется только `DefaultHandlerRef`.

Проверка повторяющихся наборов включает только включённые Routes. Отключённые Routes всё равно проходят всю структурную validation отдельного Route, включая правила типа и значения Matcher и одного Matcher каждого типа, но исключаются из сравнения с включёнными и другими отключёнными Routes. Последующее включение Route требует validation всей новой Configuration, и тогда он участвует в проверке повторяющихся наборов.

Владение configuration разделено:

- `ConfigurationVersion` владеет редактируемыми декларативными Routing metadata и Control Plane validation;
- `runtimeconfig.Snapshot` владеет глубокой копией Published Routing metadata;
- Router владеет compiled immutable table, содержащей только включённые Routes, нормализованные matchers, identity Route и разрешённые значения Handler.

Router не удерживает ни изменяемый ConfigurationVersion, ни временный ввод Handler bindings, использованный при composition.

## 10. Семантика Matcher

Исходный набор Matcher закрыт и содержит четыре типа.

| Matcher | Значения | Равенство | Отсутствующее значение Context |
|---|---|---|---|
| `message-type` | `text`, `binary` | Точное, case-sensitive enum equality | Невозможно для валидного Runtime Message |
| `principal-kind` | `authenticated`, `anonymous` | Точное, case-sensitive enum equality | Невалидный Context является internal failure, а не non-match |
| `authentication-type` | Поддерживаемый тип Authentication Provider | Точное, case-sensitive enum equality | Non-match, включая anonymous Principal |
| `authentication-provider` | Сконфигурированное имя Provider | Точное, case-sensitive string equality | Non-match, включая anonymous Principal |

Нормализация Matcher использует один точный алгоритм:

1. Type и Value должны присутствовать. Отсутствующие и пустые значения Configuration невалидны и не считаются wildcard.
2. Начальные и конечные Unicode whitespace, определённые Go-функцией `strings.TrimSpace`, удаляются из Type и Value. Внутренние whitespace сохраняются и никогда не объединяются.
3. Обрезанный Type должен точно совпадать с одним из четырёх канонических lowercase-токенов из таблицы. Он никогда не приводится к lower case и не подвергается case folding.
4. Enum values должны точно совпадать со своим каноническим lowercase-токеном: `text` или `binary`, `authenticated` или `anonymous`, а для типа Authentication — `jwt`, `api-key` или `basic`. Значения никогда не приводятся к lower case и не подвергаются case folding.
5. Значение Authentication Provider сохраняет написание и регистр после обрезки и точно сравнивается с нормализованным сконфигурированным именем Provider.
6. Type или Value, ставшие пустыми после обрезки, невалидны. При проверке сообщения отсутствующее необязательное значение Authentication отличается от пустого сконфигурированного значения Matcher: отсутствие даёт non-match согласно таблице, а пустое сконфигурированное значение никогда не проходит validation.
7. Нормализованный набор Matcher не зависит от порядка: он определяется отсортированной последовательностью нормализованных пар `(Type, Value)`. Поэтому перестановка Matchers не позволяет обойти проверку повторяющихся наборов.

Control Plane применяет этот алгоритм до принятия и публикации Routing metadata. `runtimeconfig.Builder` применяет тот же алгоритм защитно при создании своей deep copy. Router compilation снова применяет его к приватной compiled copy, не изменяя Snapshot. Все три слоя принимают и отклоняют одинаковые исходные Type и Value и создают побайтово эквивалентные канонические значения; ни один слой не выполняет дополнительное преобразование регистра.

Несколько Matchers одного Route объединяются логическим AND. OR представляется отдельными Routes. Отсутствующее необязательное значение Context даёт false только для двух matchers Authentication metadata.

Regex, glob, prefix, payload, claim, role, arbitrary attribute, HTTP header, query, remote-address и Session-ID matchers в этой версии не поддерживаются. Активная Configuration с ними отклоняется до readiness.

## 11. Алгоритм выбора

Composition один раз выполняет следующие действия:

1. Проверяет Routing metadata.
2. Исключает отключённые Routes из runtime table.
3. Разрешает ссылку Handler каждого включённого Route и необязательную ссылку Default Handler.
4. Нормализует значения Matcher.
5. Сортирует compiled Routes по возрастанию Priority.
6. Публикует неизменяемую compiled table только как часть успешной Runtime composition.

Для каждого Runtime Message Context Router:

1. возвращает cancellation, если context вызова уже отменён;
2. проверяет compiled Routes по возрастанию Priority;
3. выбирает первый Route, для которого истинны все Matchers;
4. если совпадений нет, выбирает разрешённый Default Handler, если он настроен;
5. иначе возвращает No Match.

Результатом является ровно один Handler или No Match. Router никогда не вызывает второй Handler. Совпадение нескольких Routes с разными priorities намеренно разрешается выбором первого priority. Равные priorities отклоняются при запуске. Два включённых Route с идентичными нормализованными наборами Matcher также отклоняются независимо от Priority: Route с меньшим precedence, проверяемый позже в порядке возрастания Priority, иначе был бы недостижим. Отключённые Routes не участвуют в этом сравнении повторов. Поэтому runtime ambiguity не является поддерживаемым состоянием.

Одна и та же compiled table и равные значения Context всегда дают один и тот же результат выбора.

## 12. Обратная совместимость

Четыре случая различаются:

| Состояние Configuration | Поведение Runtime |
|---|---|
| Секция Routing отсутствует | Composition устанавливает один неявный legacy default с HandlerRef `legacy`, связанный с существующим переданным Handler; текущий Echo vertical не меняется |
| Секция Routing присутствует и пуста | Валидная намеренная reject-all configuration; каждое Message даёт No Match, Router возвращает nil, Session продолжает работу |
| Секция Routing присутствует, но невалидна | Запуск Runtime завершается ошибкой до захвата Listener socket |
| Валидная Routing присутствует, но ни один Route не совпал и default отсутствует | No Match; Handler не вызывается, Router возвращает nil, Session продолжает работу без fallback |

Неявная compatibility routing entry существует только при отсутствии Routing. Она представлена Default Handler reference на `legacy`, а не обычным сконфигурированным Route. Runtime не добавляет её скрыто в явно присутствующую секцию Routing.

## 13. Модель выполнения Handler

Router реализует transport-neutral роль `message.Handler` и синхронно вызывает выбранный Handler. Выбрана модель B из рассмотренных вариантов.

Она выбрана потому, что Runtime уже компонует один Handler в Session. Подстановка Router на этой границе не добавляет второй dispatcher в Session и оставляет выбор route вне Session. Возврат Handler в Session заставил бы Session знать об outcomes выбора Router и дублировать обработку выполнения и ошибок.

Создание Handler и разрешение ссылок выполняются во время Runtime composition. Исходный локальный для composition registry Handler содержит ровно одну поддерживаемую ссылку: `legacy`. Runtime composition связывает существующий переданный экземпляр Handler с этим именем до Router compilation. При отсутствии Routing неявный compatibility default ссылается на тот же binding `legacy` и не создаёт другой Handler.

Registry является конечным construction input, а не dynamic registry, и Router не удерживает его после compilation. Каждый включённый Route и явный `DefaultHandlerRef` должны разрешиться в нём до Runtime readiness и до создания Listener или захвата socket. Поэтому в исходной реализации их единственным допустимым разрешимым значением является `legacy`; любая другая активная ссылка вызывает startup failure категории unresolved Handler. Синтаксически валидная ссылка отключённого Route не разрешается до более поздней Configuration, включающей этот Route.

Compiled Routes могут совместно использовать один экземпляр Handler. Каждая реализация Handler владеет и документирует собственную concurrency safety. Router не добавляет сериализацию и не удерживает lock во время вызова Handler.

Для выбранного route Router ровно один раз вызывает ровно один Handler с исходным cancellation context и неизменяемым Runtime Message Context. При No Match Handler не вызывается, а существующий return contract `message.Handler` получает `nil`. Поэтому Session продолжает read loop. После выбора явной секции Routing при composition результат No Match никогда не приводит к fallback на legacy Handler.

После начала вызова Router возвращает результат выбранного Handler и не заменяет его отдельно наблюдаемым cancellation outcome. Выбранный Handler владеет наблюдением того же переданного context во время своего выполнения. Router не выполняет fallback selection и никогда не вызывает другой Handler из-за конкуренции cancellation с этим вызовом.

## 14. Модель ошибок

Реализации требуются стабильные категории, совместимые с `errors.Is`.

| Outcome | Категория | Эффект для Runtime/Session |
|---|---|---|
| Невалидная форма Routing, повтор identity/priority, запрещённый пустой matcher или статическая ambiguity | Invalid routing configuration | Не допускает readiness Runtime до запуска Listener |
| Неподдерживаемый Matcher | Unsupported routing capability | Не допускает readiness Runtime до запуска Listener |
| Активная неразрешённая ссылка Handler | Unresolved Handler | Не допускает readiness Runtime до запуска Listener |
| Нет явного совпадения и default | No Match, представленный значением `nil` из `message.Handler` | Handler не вызывается, legacy fallback отсутствует; текущий read loop Session продолжается |
| Runtime ambiguity | Недостижима после успешной validation | Рассматривается как internal failure, если дефект инварианта сделал её наблюдаемой |
| Выбранный Handler вернул ошибку | Handler execution error с сохранением cause | Возвращается через Router; текущий `Session.Run` завершается и сохраняет `errors.Is` при wrapping |
| Context вызова отменён до selection | `context.Canceled` или `context.DeadlineExceeded` | Handler не вызывается; ошибка возвращается в execution path Session |
| Невалидный Runtime Message Context | Internal Router failure | Завершает текущий execution path Session; никогда не считается No Match |

Router не создаёт protocol response, acknowledgement, retry, альтернативный route или Delivery action ни для одного outcome. Текст ошибки может содержать безопасный Route ID или ссылку Handler, но не payload, Principal identity, credentials, claims или Secrets.

## 15. Конкурентность и lifecycle

Router не имеет lifecycle states, `Start` или `Stop`. Runtime composition создаёт его до запуска Listener и освобождает вместе с component graph Runtime.

Compiled route slice, значения Matcher, Route IDs, Handler references и разрешённые значения Handler никогда не меняются после создания. Selection выполняет read-only iteration и не требует mutex Router. Concurrent-вызовы из разных Sessions безопасны и не разделяют изменяемое per-message state.

Router не запускает goroutines, не владеет очередями и не удерживает lock во время выполнения Handler. Каждый вызов получает собственные Runtime Message Context и cancellation context; значения одной Session нельзя повторно использовать или записать в Context другой. Реализации Handler остаются ответственными за собственный контракт concurrent-вызовов.

## 16. Startup validation

Control Plane validation владеет декларативной формой и нормализацией:

- наличием секции;
- ограничениями collections;
- синтаксисом Route и Handler reference;
- уникальностью Route ID и Priority;
- поддерживаемыми именами Matcher и domains значений;
- одним Matcher каждого типа на Route;
- непустым списком Matcher для включённых Routes;
- точными повторами наборов Matcher включённых Routes;
- единственным необязательным полем Default Handler; scalar-схема не позволяет выразить несколько defaults.

`runtimeconfig.Builder` является defensive publication boundary. Он отказывается создавать Snapshot из Published ConfigurationVersion, нарушающего перечисленные выше структурные инварианты Routing, и выполняет deep copy только валидного представления. Это не передаёт владение declarative validation от Control Plane: проверка не позволяет malformed programmatic input стать доверенным Runtime Snapshot.

Runtime executability validation владеет:

- поддержкой версии Routing Snapshot и всех активных типов Matcher;
- разрешением каждого Handler включённого Route и настроенного Default Handler;
- компиляцией без изменения Snapshot;
- установкой неявного legacy default только при отсутствии Routing;
- доказательством возможности опубликовать compiled Router до запуска Listener.

Validation выполняется в Runtime composition до создания Listener и захвата socket. Routing validation участвует в существующей модели startup capability: активная невалидная или неподдерживаемая routing не допускает Ready; отключённые Routes не исполняются, но всё равно проходят структурную Control Plane validation. Инварианты deep copy Snapshot и compiled table проверяются независимо.

Правила validation имеют одно нормативное определение в DP. Тесты Control Plane и `runtimeconfig.Builder` должны использовать общие cases или зеркальные proof tables, чтобы defensive validation не расходилась незаметно. Runtime composition не повторяет проверки repository или editable state; она проверяет поддержку Snapshot и разрешает активные runtime capabilities.

## 17. Границы пакетов

Ожидаемое направление зависимостей:

```text
ConfigurationVersion
    -> runtimeconfig.Builder
    -> immutable Routing Snapshot
    -> Runtime composition and Handler bindings
    -> Router compiled table

Session
    -> immutable Runtime Message Context
    -> Router as message.Handler
    -> selected message.Handler
```

Новый пакет `internal/router` может зависеть от:

- transport-neutral контрактов Message, Context, Sender и Handler из `internal/message`;
- immutable routing snapshot или локального для Router construction input;
- стандартных пакетов Go для context и errors.

Запрещённые imports и dependencies:

- `net/http`;
- любой пакет реализации WebSocket;
- `internal/listener`;
- конкретные типы `internal/session`;
- `internal/sessionmanager` и его Registration или Lookup internals;
- `internal/handshake` и `internal/connection`;
- Delivery, Persistence, repositories, Control Service handlers, logging implementations и plugin infrastructure.

Runtime composition может зависеть от Router для wiring graph. Router никогда не зависит от Runtime Host или Container.

## 18. Соображения безопасности

- Router не видит credentials, Authorization headers, Secret References, request values или resolved Secrets.
- Байты payload не являются входом matcher и никогда не включаются в ошибки Router.
- Claims, roles, attributes, metadata, ID и name Principal не являются routable в первой версии; routable только явные metadata классификации authentication.
- Имена Authentication Provider рассматриваются как безопасные Configuration identifiers, а не credential material.
- Route IDs и Handler references имеют ограниченную длину и ограниченный синтаксис.
- Количество Route и Matcher ограничено, чтобы Configuration не могла создать неограниченную per-message evaluation.
- Router не ослабляет Authentication, admission, ownership Session или Runtime cancellation.

## 19. Соображения наблюдаемости

Этот эпик не вводит logging, Metrics, tracing framework или event bus. Стабильные категории возвращаемых ошибок позволяют существующему terminal-error path наблюдать failures выбранного Handler. Безопасная диагностика может указывать Route ID и Handler reference, но не должна включать Message payload или значения Principal identity.

No Match является нормальным outcome и не сообщается как terminal error. Будущий Metrics может считать только ограниченные Route IDs после того, как эпик Metrics определит cardinality и ownership. Этот DP не требует такой instrumentation.

## 20. Стратегия тестирования

Proof реализации должен включать:

- каждый исходный тип Matcher, правило equality, case rule и missing-value rule;
- логический AND внутри одного Route;
- возрастающий Priority, пересекающиеся Routes и first-match selection;
- отклонение duplicate ID, duplicate Priority, duplicate Matcher type и идентичного нормализованного Matcher set включённых Routes независимо от Priority;
- исключение отключённого Route из selection и сравнения повторяющихся наборов при сохранении его structural validation;
- эквивалентность whitespace, missing/empty, canonical enum, case-preservation, order-independent set и cross-layer normalization;
- implicit legacy behavior при отсутствующем Routing;
- reject-all behavior при присутствующем пустом Routing;
- настроенный Default Handler;
- No Match как `nil` без вызова Handler, legacy fallback или завершения Session;
- единственный исходный binding Handler `legacy`, implicit compatibility reference, отклонение неизвестной активной ссылки, отсрочку разрешения ссылки отключённого Route и совместное использование Handler instances;
- сохранение payload text и binary Message;
- вызов ровно одного выбранного Handler;
- ошибку выбранного Handler и сохранение `errors.Is` через Session;
- cancellation до selection и во время выполнения Handler;
- concurrent routing через одну immutable compiled table;
- изоляцию значений Context разных Sessions;
- deep-copy Context и Snapshot;
- отклонение неразрешённого Handler и неподдерживаемого Matcher до readiness;
- реальный поток `ConfigurationVersion -> Published -> runtimeconfig.Builder -> Runtime Host -> WebSocket -> Session -> Router -> Handler`;
- совместимость текущего Echo vertical при отсутствии Routing;
- обновление исчерпывающей матрицы поддержки Runtime Snapshot;
- проверки package/import boundaries, если их поддерживают инструменты репозитория.

Тесты используют channels, barriers, injected Handlers и реальные integration clients, где это уместно. Произвольные sleeps не являются синхронизацией, и каждая запущенная goroutine имеет ограниченный cleanup path.

## 21. Рассмотренные альтернативы

### Router возвращает Handler для выполнения Session

Отклонено. Это делает Session осведомлённой о решениях Router и добавляет selection-specific branching компоненту, владеющему transport execution.

### Расширить `message.Message` полями Session и Principal

Отклонено. Runtime Message остаётся application data; execution и identity metadata принадлежат отдельному immutable Context.

### Хранить значения Context в `context.Context`

Отклонено. `context.Value` создал бы скрытые зависимости от данных и ослабил compile-time contracts.

### Маршрутизировать по payload, claims, произвольным metadata или HTTP values

Отклонено для первой версии. Это расширяет security, normalization, compatibility и matcher semantics без доказанного use case.

### Вызывать каждый совпавший Handler

Отклонено. Выполнение нескольких Handler является Delivery или fan-out behavior и требует semantics порядка, partial failure и backpressure вне этого эпика.

### Разрешать равные priorities по Route ID или declaration order

Отклонено. Startup rejection понятнее неявного tie-break rule и предотвращает случайные изменения поведения при редактировании Configuration.

### Dynamic Handler registry

Отклонено. Runtime composition разрешает конечный явный набор до readiness; dynamic registration вводит ненужные Router lifecycle и synchronization.

### Считать присутствующий пустой Routing legacy-поведением

Отклонено. Это уничтожило бы семантическое различие между отсутствующей configuration и явным выбором оператора не маршрутизировать ничего.

## 22. Последствия

Положительные последствия:

- Router встраивается в существующую границу Handler и не меняет transport ownership.
- Selection детерминирован, ограничен, lock-free и configuration-first.
- Текущие deployments сохраняют Echo behavior при отсутствующем Routing.
- Невалидные routes и неразрешённые Handlers отклоняются до traffic.
- Router остаётся изолированным от будущих контрактов Session Manager и Delivery.

Издержки и ограничения:

- Вход Handler должен согласованно мигрировать на Runtime Message Context.
- Исходные routing inputs намеренно ограничены.
- Первая production composition предоставляет только существующий Handler как `legacy`, пока не реализованы дополнительные явные Handler bindings.
- No Match молча не выполняет application action, поскольку message-level acknowledgement protocol отсутствует.
- Concurrency Handler остаётся отдельным implementation contract.

## 23. Этапы реализации

1. Ввести immutable Runtime Message Context и перенести Handler, Echo и тесты Session без изменения поведения.
2. Добавить необязательные Routing metadata и Control Plane validation в ConfigurationVersion.
3. Выполнять deep copy Routing в Runtime Snapshot и расширить покрытие support matrix.
4. Реализовать compilation Router, закрытый набор Matcher, детерминированный selection и категории ошибок.
5. Связать существующий Handler как `legacy` и скомпоновать Router до запуска Listener.
6. Добавить unit, concurrency, startup-failure и full WebSocket integration proofs.
7. Обновить фактические current-state и changelog только после прохождения проверок реализации.

Каждый этап сохраняет компилируемый vertical и не вводит Delivery, интеграцию Session Manager или dynamic Handler registration.

## 24. Открытые вопросы

Отсутствуют. Все нормативные решения, необходимые для первой реализации Router, закрыты этим proposal. Дополнительные matchers, Handler bindings, routing outcomes или targeting semantics потребуют сфокусированной ревизии или отдельного design после появления конкретных use cases.

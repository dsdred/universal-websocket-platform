# Review контракта Runtime Integration DP-005

[English version](../../en/reviews/dp-005-runtime-integration-contract-review.md)

**Статус:** Completed

**Дата review:** 2026-07-20

**Оценка:** Approved для продолжения Runtime integration

## 1. Scope

Это review фиксирует устранение blocker границы создания, обнаруженного до Runtime integration DP-005. Оно проверяет только контракт между Runtime composition, строгим compiler Router, compatibility path для отсутствующего Routing и существующим переданным legacy Handler.

Этот документ не оценивает Go-реализацию. Selection semantics Router, поведение Matcher, lifecycle Runtime, структура Snapshot и выполнение downstream Handler остаются под управлением [DP-005](../design/DP-005-runtime-message-router.md) и не меняются, кроме уточнённого контракта создания.

## 2. Blocker, обнаруженный до реализации

DP-005 требовал, чтобы отсутствующий Routing устанавливал неявный compatibility Router на основе существующего переданного legacy Handler. Реализованный строгий compiler требует ненулевой валидный `RoutingSnapshot`, а Runtime одновременно допускает nil переданного Handler как намеренную discard configuration.

Поэтому существующие публичные контракты не определяли, как Runtime может создать Router для отсутствующего Routing, не изменяя semantics строгого compiler, не обходя Router и не изобретая поведение nil Handler во время реализации. Runtime integration остановилась до изменения кода, поскольку каждая интерпретация означала бы неутверждённое архитектурное решение.

## 3. Отклонённые интерпретации

### Изменить строгий compiler для приёма nil

Отклонено. Nil не описывает явную Routing configuration и не может удовлетворять validation contract строгого compiler. Трактовка `router.New(nil, ...)` как compatibility construction объединила бы declarative compilation с поведением отсутствующей Configuration и изменила бы установленное значение невалидного compiler input.

Строгий compiler продолжает требовать валидный `RoutingSnapshot` и отклонять nil Routing input.

### Обходить Router при отсутствующем Routing

Отклонено. Прямая передача legacy Handler в Session создала бы два пути выполнения Message и заставила бы наличие Runtime Configuration управлять message-time composition. Это также нарушило бы инвариант, согласно которому каждое Message после startup проходит через `Router.Handle`.

После startup Runtime не должен проверять routes или обходить Router.

### Установить synthetic discard Handler

Отклонено. Synthetic no-op Handler ввёл бы execution object, который не требуется для сохранения существующего поведения nil Handler. No Match уже предоставляет требуемый нетерминальный результат `nil` без вызова Handler.

## 4. Утверждённый контракт compatibility construction

Router предоставляет отдельный явный compatibility construction path. Его концептуальная ответственность эквивалентна:

```text
Compatibility Router construction
    input: injected legacy Handler
    compiled Routes: zero
    Default Handler: injected Handler when non-nil; absent when nil
    output: one immutable Router
```

Эта factory представляет только отсутствующий Routing. Она не принимает настроенные Routes, не нормализует Matchers, не разрешает registry и не меняет строгий compiler. Точное Go-имя остаётся выбором уровня реализации только тогда, когда оно следует существующим соглашениям репозитория и сохраняет эту единственную ответственность.

## 5. Поведение nil legacy Handler

Когда Routing отсутствует, а переданный legacy Handler равен nil:

- startup Runtime завершается успешно;
- compatibility Router содержит ноль compiled Routes;
- compatibility Router не имеет Default Handler;
- каждое Message даёт No Match;
- `Router.Handle` возвращает nil;
- Session продолжает read loop;
- synthetic Handler не создаётся и не вызывается.

Это сохраняет существующее discard behavior, оставляя Router на message path.

## 6. Создание Runtime и путь Message

Runtime выполняет ровно одну ветвь создания во время startup:

```text
Routing present
    -> strict Router compilation from RoutingSnapshot

Routing absent
    -> compatibility Router construction from injected legacy Handler

Either branch
    -> store one immutable Router
    -> inject Router as message.Handler
    -> reuse Router.Handle for every Message
```

Создание завершается до создания Listener и захвата socket. Compilation Router, sorting, normalization и resolution Handler не выполняются во время обработки Message. Итоговый Router является immutable и не требует принадлежащей Runtime synchronization на hot path.

## 7. Устранённые findings

| Finding | Решение |
|---|---|
| Отсутствует контракт создания при отсутствующем Routing | Устранено отдельной compatibility factory |
| Возможное изменение `router.New(nil, ...)` | Отклонено; semantics строгого compiler не меняется |
| Возможный обход Router в Runtime | Отклонено; обе startup branches публикуют Router |
| Не определён nil переданный legacy Handler | Устранено compatibility Router без Default Handler и обычным No Match |
| Не определены количество созданий и повторное использование | Устранено ровно одним startup construction и повторным использованием одного immutable Router |

Это уточнение не объявляет устранёнными несвязанные findings DP-005.

## 8. Оценка approval

**Вердикт: Approved для продолжения Runtime integration.**

Граница создания теперь однозначна. Реализация может добавить узкую compatibility factory, выбрать один путь создания Router во время startup Runtime и внедрить итоговый Router на существующей границе `message.Handler`. Она должна сохранить поведение строгого compiler, совместимость nil discard, существующий lifecycle Runtime и единственный post-startup путь `Router.Handle`.

Реализация по-прежнему подлежит независимому review кода и тестов. Это approval не утверждает, что Runtime integration уже реализована.

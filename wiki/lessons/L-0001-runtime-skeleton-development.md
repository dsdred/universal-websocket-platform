# L-0001 — Runtime Skeleton Development

## 1. Context

Задача Runtime Skeleton состояла в создании минимального post-Commit контура
Execution Owner. Контур должен был занять единственный execution path, вызвать
`Session.Start()`, при допустимом результате вызвать `Session.Run()` и завершить
текущий этап в состоянии `Terminalizing`.

В scope не входили terminal chain, Cleanup, Complete, Observer, Terminal Result,
Runtime callback lifecycle, Lease Release и интеграция с Dispatcher, Runtime или
Session Manager.

Источниками архитектурных требований были:

- DP-003 Runtime Session Manager;
- DP-004 Per-Session Execution Boundary.

## 2. Initial Assumption

Первоначальная модель рассматривала lifecycle transition как общий примитив,
доступный наряду с Runtime Execution Loop. Сам Loop последовательно переводил
Owner через `Committed`, `Starting`, `Running` и `Terminalizing`, а общий примитив
сохранял возможность выполнить допустимый переход независимо от Loop.

Первая реализация успешно проходила `go test`, `go vet` и `gofmt`. Проверки
подтвердили компилируемость, отсутствие обнаруживаемых стандартными тестами
ошибок и корректное форматирование. Они не доказали исключительное владение
post-Commit lifecycle.

Независимое implementation review выявило архитектурный дефект, который не был
виден в исходном наборе тестов.

## 3. Review Finding

BLOCKER F-01 описывал допустимую последовательность, в которой Runtime Execution
Loop уже перевёл Owner в `Starting` и выполнял `Session.Start()`, а внешний код в
это время мог выполнить переход `Starting → Running`.

Если `Start()` после этого возвращал ошибку, Owner мог остаться в `Running`, хотя
успешного Start не было и `Run()` не вызывался. Тем самым публично доступная
mutation нарушала связь между состоянием Owner и фактическим Session lifecycle.

## 4. Investigation

Исследование выполнялось последовательно:

```text
Independent Review
    ↓
предположение о конкурентном race
    ↓
повторное чтение DP-003 и DP-004
    ↓
анализ ownership до и после Commit
    ↓
обнаружение двух post-Commit lifecycle writers
```

Первичная формулировка указывала на interleaving между двумя операциями. Анализ
только средств синхронизации мог бы привести к сериализации этих операций, но
не отвечал бы на вопрос, почему обе операции вообще имели право изменять одно
состояние.

DP-анализ разделил три полномочия:

- Commit публикует ownership и переводит execution binding в `Committed`;
- Runtime Execution Loop управляет post-Commit lifecycle;
- внешние сигналы публикуют termination intent, но не выполняют lifecycle
  transitions.

После такого разделения дефект был классифицирован как нарушение ownership, а
не как недостаток механизма синхронизации.

## 5. Root Cause

Главная причина:

> После Commit существовало более одного post-Commit lifecycle writer.

Runtime Execution Loop изменял lifecycle как владелец execution path. Одновременно
общая transition operation позволяла внешнему caller выполнить те же изменения.
Обе стороны могли использовать нормативно допустимые переходы, но совокупность
этих полномочий не соответствовала DP.

DP-004 назначает единственный заранее созданный path владельцем Session Start,
Run и terminal progression после Commit. DP-003 исключает Session Manager из
Session execution ownership. Следовательно, корректность отдельного перехода
не давала внешнему компоненту права выполнять этот переход.

## 6. Architectural Resolution

Итоговый инвариант сформулирован так:

> После Commit существует ровно один post-Commit lifecycle writer:
> Runtime Execution Loop.

Из него следуют ограничения:

- внешний Transition ограничен publication boundary `PreCommit → Committed`;
- post-Commit lifecycle mutation является внутренней обязанностью Runtime
  Execution Loop;
- Stop публикует только termination intent;
- Session возвращает результаты Start и Run, но не изменяет lifecycle Owner;
- Runtime callback публикует causal termination state, но не изменяет lifecycle;
- копирование Owner не создаёт отдельного writer или отдельный execution claim.

Это разделяет lifecycle authority и termination signaling. Источник завершения
может повлиять на решение Loop, но не получает право выполнить переход сам.

## 7. Implementation Changes

Реализация была приведена к архитектурному контракту следующими изменениями:

- публичная transition boundary ограничена переходом `PreCommit → Committed`;
- post-Commit transition mechanism оставлен внутри пакета Execution Owner;
- Runtime Execution Loop сохранил последовательность Start, Run и перехода в
  `Terminalizing`;
- комментарий Owner обновлён в соответствии с фактическим lifecycle ownership;
- тесты state machine отделены от package-external proof-тестов authority;
- blocking test harness получил обязательные unblock и join операции.

Terminal chain и интеграционные обязанности при этом не добавлялись.

## 8. Proof Strategy

Тестовая стратегия была построена вокруг архитектурных инвариантов, а не вокруг
отдельных функций.

Proof tests подтверждают:

- существует один execution path;
- post-Commit lifecycle имеет единственного writer;
- одновременный claim из `Committed` имеет одного победителя;
- проигравшие claims не изменяют lifecycle;
- `Start()` вызывается не более одного раза;
- `Run()` вызывается не более одного раза;
- `Running` достижим только после успешного Start;
- ошибка Start запрещает Run и приводит к `Terminalizing`;
- возврат или ошибка Run приводит к `Terminalizing`;
- Stop остаётся termination intent и не становится lifecycle writer;
- копии Owner используют общий lifecycle, control cell и execution claim;
- публичные post-Commit, обратные, self и неизвестные transitions отклоняются
  без изменения состояния;
- blocking gates освобождаются, а test goroutines присоединяются даже при
  промежуточном failure.

Для simultaneous claim все callers подготавливаются до открытия общего barrier.
Победитель удерживается в `Starting`, пока остальные callers не подтвердят
проигрыш. Такой порядок доказывает atomicity claim, а не только отсутствие второго
Start после того, как первый execution уже завершился.

## 9. Process Lessons

Из разработки Runtime Skeleton следуют процессные выводы:

1. Успешные `go test`, `go vet` и formatting checks не доказывают корректность
   ownership model.
2. Независимое review необходимо рассматривать как проверку архитектурных
   инвариантов, а не только качества реализации.
3. Найденный concurrency-сценарий сначала следует проверить на уровне модели
   ownership и lifecycle authority.
4. Исправление должно исключать недопустимое полномочие, а не только упорядочивать
   конкурирующие операции.
5. Proof tests являются исполняемой частью архитектурного доказательства: они
   фиксируют linearization boundaries, exactly-once свойства и запрещённые пути.
6. Blocking concurrency tests должны сохранять cleanup guarantees и при падении
   промежуточной проверки.
7. Commit допустим после завершения реализации, независимого review и полного
   набора проверок.

## 10. Result

Runtime Skeleton завершает определённый для него этап lifecycle в
`Terminalizing` и не принимает обязанности terminal chain.

Production-модель соответствует ownership contract DP-003 и DP-004. После
Commit Runtime Execution Loop является единственным post-Commit lifecycle writer.
Внешние компоненты ограничены publication boundary и termination signaling.

Proof tests подтверждают single execution path, exclusive lifecycle authority,
Start/Run exactly-once свойства, корректную Start/Run ordering, copy safety и
atomic simultaneous claim.

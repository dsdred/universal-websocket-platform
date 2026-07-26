# PROCESS-001 — AI Development Workflow

## Purpose

Определить единый воспроизводимый процесс разработки с Coordinator и
специализированными агентами. Процесс предотвращает потерю контекста,
смешение ответственности и завершение задачи при рассинхроне репозитория.

## Scope

Процесс применяется к изменениям production-кода, тестов, архитектуры, API,
моделей данных, конфигурации, документации, сборки и поставки.

Coordinator может исключить неприменимый этап только с явным обоснованием в
task record. Architecture Analysis, Verification, Review, Documentation
Synchronization и Closure нельзя считать пройденными неявно.

## Principles

1. Репозиторий является источником истины.
2. Approved и Frozen архитектура имеет приоритет над текущей реализацией.
3. Документация является частью результата.
4. Архитектурное решение предшествует реализации.
5. Каждый агент действует только в пределах своей роли.
6. Каждый этап имеет вход, проверяемый результат и handoff.
7. Планируемое и реализованное состояние описываются раздельно.
8. Критический drift или архитектурное противоречие останавливает работу.

## Roles

Процесс использует роли:

- [Coordinator](agents/coordinator.md);
- [Architect](agents/architect.md);
- [Documentation Agent](agents/documentation.md);
- [Developer](agents/developer.md);
- [Tester](agents/tester.md);
- [Reviewer](agents/reviewer.md).

Один агент может выполнять несколько непересекающихся ролей только при
явном назначении Coordinator. Автор изменения не выполняет независимый
final review того же изменения.

## Source of Truth and Status

При конфликте применяется следующий порядок:

1. Approved ADR;
2. Active или Frozen ARCH;
3. Approved или Accepted DP в пределах его scope;
4. спецификации;
5. task contract;
6. production-код и тесты как доказательство реализованного состояния;
7. навигационные и статусные документы.

Draft не является нормативным и не переопределяет Approved, Active или Frozen
документ. Commit, сообщение commit и история чата не меняют статус документа.
Статус меняется только явным изменением поля `Status` по утверждённому
процессу; Documentation Agent не повышает его самостоятельно.

Для проектных документов различаются:

- **Design Status** — Draft, Approved, Accepted или иной явно установленный
  статус решения;
- **Implementation Status** — Planned, Partial, Implemented или Completed.

Accepted design может оставаться не реализованным. Реализация не повышает
Design Status автоматически.

## Language Policy

Публичная и проектная документация в `docs/en/` и `docs/ru/` поддерживается
зеркально: совпадают набор документов, структура, статусы и нормативный смысл.

`docs/engineering/` и `docs/tasks/` являются внутренними operational
документами. На текущем этапе их канонический язык — русский; обязательное
EN-зеркало для них отсутствует. Отсутствие такого зеркала не является drift.

## Autonomous Continuation

Точная bare-команда `Продолжай проект.`, определённая в
[`AGENT.md`](AGENT.md), запускает следующий алгоритм.

### Preflight

Coordinator до выбора работы:

1. читает branch, status и локальную историю без изменения репозитория;
2. читает task index и все active task records;
3. сверяет `.ai/PROJECT_CONTEXT.md`, `spec/current-state.md`,
   `spec/decisions.md` и MASTER_PLAN;
4. не очищает, не сбрасывает и не переопределяет происхождение необъяснённых
   изменений.

Attributed dirty worktree разрешён только для возобновления ровно одной active
task, когда branch и изменения однозначно соответствуют её record. При выборе
или подготовке новой task любой dirty baseline останавливает autonomous
continuation, включая изменения завершённой task до их отдельно разрешённого
commit. Неатрибутированный dirty worktree и diverged либо необъяснимый
branch/history всегда останавливают autonomous continuation.

### Deterministic Work Selection

Coordinator выбирает ровно один bounded slice:

1. возобновляет однозначно атрибутированную `In Progress` task; `Blocked` task
   возобновляется только при наличии в репозитории evidence, что blocker
   устранён, иначе autonomous continuation останавливается;
2. иначе использует явную current или next task из project-state documents
   либо closure последней завершённой task после проверки readiness;
3. иначе пересекает dependency ordering текущей milestone из MASTER_PLAN,
   фактические gaps из `spec/current-state.md`, открытые решения из
   `spec/decisions.md` и существующие ADR, ARCH и DP.

MASTER_PLAN задаёт зависимости, но не является очередью задач. Candidate
считается `Ready`, только если:

- это наименьший независимо проверяемый bounded slice;
- все обязательные продуктовые и архитектурные решения уже существуют;
- его prerequisites подтверждены repository evidence;
- scope, non-goals и verification можно однозначно записать;
- работа не требует неразрешённого изменения Approved или Frozen источника.

Architecture refinement может быть `Ready`, когда production implementation
ещё не готова.

Ready candidates ранжируются последовательно:

1. dependency current milestone;
2. prerequisite order;
3. наименьший независимо проверяемый scope;
4. наименьший unresolved risk;
5. порядок первого появления в authoritative repository documents.

Если остаются materially different candidates или выбор является продуктовой
приоритизацией, Coordinator останавливается и запрашивает решение пользователя.
Отсутствие Ready candidate, конфликт источников или отсутствующее критическое
решение также являются stop condition.

### Task and Branch Preparation

После выбора новой task Coordinator сначала готовит безопасную локальную
task-ветку, если она требуется. Task record является первым content change в
этой ветке и создаётся либо актуализируется до остальной работы. Record
обязательно содержит:

- evidence выбора и отклонённые альтернативы;
- scope, non-goals, sources, roles и verification;
- branch decision и stop conditions;
- следующий candidate, который не становится active автоматически.

Новая task и её ветка начинаются только с чистого, понятного baseline и при
отсутствии другой active task. Для production work используется локальная
ветка с префиксом `feature/`, для documentation-only work — с префиксом
`docs/`; новый slug должен включать Task ID, если он уже назначен. Если
подходящая task-ветка уже активна, работа продолжается в ней. Создавать или
переключаться на ветку можно только без перезаписи, удаления, rebase или
изменения `main`.

Bare-команда не разрешает stage, commit, push, merge, удаление веток, fetch,
pull, remote mutation или иное неявное изменение git history. Исключение для
attributed dirty worktree применяется только к resume единственной active task;
оно никогда не разрешает выбор или подготовку новой task. Unattributed dirty
или diverged baseline всегда требует остановки.

### Autonomous Full Cycle

После подготовки task Coordinator проводит применимые стадии:

```text
Task Intake
    -> Documentation Baseline
    -> explicit Architecture Confirmation
    -> Pre-Implementation Documentation, если меняется contract
    -> Developer, только если меняется production code
    -> Verification
    -> Independent Review and Rework
    -> PROCESS-002
    -> Scope Audit
    -> Final Checks and Independent Review
    -> Coordinator Acceptance
    -> Project-State Update
    -> Next-Task Recommendation
```

Роли и независимость review сохраняются. Неприменимый этап пропускается только
с явным обоснованием в task record. Следующая рекомендация не запускает новую
task и не создаёт для неё branch.

## Task Intake

Coordinator:

1. фиксирует цель, scope, запреты и ожидаемый результат;
2. определяет затрагиваемые источники истины;
3. назначает необходимые роли;
4. создаёт или подтверждает task record;
5. останавливает intake, если задача противоречит источнику более высокого
   уровня.

Результат: однозначный task contract и список обязательных evidence.

## Documentation Baseline

До проектирования или реализации Documentation Agent либо назначенный auditor:

1. собирает inventory затрагиваемых документов;
2. проверяет статусы, EN/RU parity, индексы и ссылки;
3. отделяет documented architecture от implemented state;
4. фиксирует drift.

Критический drift передаётся Coordinator и блокирует новую функциональность.

## Architecture Analysis

Architect:

1. проверяет соответствие ADR, ARCH и DP;
2. определяет ownership, responsibility, lifecycle, validation и failure
   boundaries;
3. подтверждает существующее решение либо формирует отдельное архитектурное
   изменение;
4. задаёт acceptance criteria и implementation constraints.

Architect не реализует код. Неразрешённое противоречие возвращается
Coordinator со статусом `Blocked`.

## Pre-Implementation Documentation

Documentation Agent фиксирует только утверждённое решение:

1. обновляет необходимые EN/RU документы;
2. сохраняет их semantic parity;
3. обновляет индексы и ссылки;
4. не выдаёт planned state за implemented;
5. не меняет статус без явного решения.

Результат передаётся Reviewer для проверки до начала реализации, если task
меняет архитектурный контракт.

## Implementation

Developer:

1. реализует только утверждённый scope;
2. не принимает новых архитектурных решений;
3. сохраняет совместимость и ownership invariants;
4. добавляет необходимые тесты;
5. останавливается при обнаружении архитектурного пробела.

Результат: минимальный reviewable diff и список изменённого поведения.

## Verification

Tester либо Developer по назначению Coordinator:

1. изучает существующие тесты;
2. сопоставляет их с acceptance criteria;
3. добавляет недостающие proof tests в разрешённом scope;
4. выполняет применимые formatter, test, race, vet, lint и diff checks;
5. сохраняет точные результаты и причины недоступных проверок.

Проверка считается доказательством только при воспроизводимом результате.

## Review

Независимый Reviewer проверяет:

- соответствие источникам истины;
- ownership и lifecycle invariants;
- scope;
- качество proof tests;
- отсутствие скрытых изменений API или архитектуры;
- полноту документации.

Reviewer возвращает `Approved`, `Approved with Findings` либо `Needs
Revision` согласно task contract. Неустранённый blocking finding запускает
Rework Loop.

## Final Documentation Synchronization

После реализации Documentation Agent выполняет PROCESS-002:

1. отражает только фактически реализованное состояние;
2. сохраняет отдельно Design Status и Implementation Status;
3. синхронизирует EN/RU, индексы, current state, roadmap и changelog только
   там, где они применимы;
4. фиксирует сознательно отложенную работу.

## Scope Audit

После PROCESS-002 и до финального gate Coordinator проверяет полный diff.
Каждый изменённый production, test, documentation или generated-файл
классифицируется как:

- `Required` — необходим для acceptance criteria либо обязательной
  синхронизации;
- `Questionable` — связь со scope требует отдельного доказательства;
- `Removable` — не требуется task contract.

Audit отдельно проверяет:

- не началась ли следующая task или преждевременная pipeline integration;
- нет ли unrelated behavior, refactoring или изменения архитектуры;
- нет ли generated, formatting-only либо случайных файлов;
- не описывает ли документация planned behavior как implemented.

`Removable` change должен быть удалён владельцем изменения. `Questionable`
change должен получить доказательство необходимости либо также быть удалён.
Результат audit и disposition каждого finding фиксируются в task record.

## Closure

Coordinator закрывает задачу только после получения всех обязательных
handoff и evidence.

Closure record содержит:

- выполненный scope;
- изменённые файлы;
- архитектурное соответствие;
- результаты проверок;
- известные ограничения;
- итоговый статус;
- следующий разрешённый шаг;
- следующую рекомендуемую ready work либо причину отсутствия рекомендации.

Commit выполняется только при явном разрешении.

## Rework Loop

При finding или failed verification:

```text
Reviewer or Tester
        ↓
Coordinator
        ↓
Architect, Documentation Agent, or Developer
        ↓
Verification
        ↓
Independent Review
```

Coordinator направляет проблему владельцу соответствующей ответственности.
Повторная работа ограничивается finding и не расширяет исходный scope.

## Handoff Contracts

Каждый handoff содержит:

- исходный task и scope;
- использованные источники истины;
- принятые и запрещённые решения;
- изменённые или ожидаемые файлы;
- acceptance criteria;
- выполненные проверки;
- открытые findings и риски;
- требуемое действие следующей роли.

Устный или чатовый контекст не заменяет handoff в репозитории.

## Stop Conditions

Этап останавливается, если:

- отсутствует обязательный источник истины;
- Draft конфликтует с Approved, Active или Frozen архитектурой;
- ownership или responsibility невозможно определить однозначно;
- требуется решение вне task scope;
- обнаружен критический documentation drift;
- обязательная проверка завершилась ошибкой;
- Reviewer выдал blocking finding;
- изменение затронет неразрешённые файлы или публичный контракт.

## Failure Handling

Агент:

1. прекращает только затронутую работу;
2. сохраняет репозиторий в проверяемом состоянии;
3. фиксирует факты, воспроизводимый сценарий и нарушенный инвариант;
4. не подменяет архитектурное решение предположением;
5. возвращает проблему Coordinator.

Coordinator выбирает Rework Loop либо статус `Blocked`.

## Required Evidence

До Closure должны существовать:

- traceability от task к архитектурным требованиям;
- review result;
- результаты применимых проверок;
- подтверждение documentation parity и ссылок;
- явная применимость `spec/current-state.md`, `spec/decisions.md`,
  `.ai/PROJECT_CONTEXT.md`, roadmap, root README и `CHANGELOG.md`;
- scope audit всех изменённых файлов;
- подтверждение отсутствия неожиданных файлов в diff.

## Definition of Done

Задача завершена, только если:

- contract complete: Critical отсутствуют, Major устранены или явно приняты;
- language complete для документов с обязательным EN/RU mirror;
- navigation complete: индексы, ссылки и orphans проверены;
- project-state complete: каждый обязательный status-файл проверен явно;
- scope complete: все изменённые файлы классифицированы, а `Questionable` и
  `Removable` changes разрешены;
- repository complete: проверки успешны, conflict markers и trailing
  whitespace отсутствуют;
- final gate выполнен независимым Reviewer;
- Coordinator зафиксировал Closure;
- commit не выполнен без явного разрешения.

## Process Invariants

Запрещается:

- писать код до необходимого архитектурного анализа;
- менять архитектуру внутри implementation stage;
- документировать предположение как факт;
- повышать статус документа по факту commit или реализации;
- смешивать planned и implemented state;
- закрывать задачу при критическом drift;
- пропускать независимый review;
- создавать конкурирующие источники истины;
- полагаться на память диалога для продолжения работы.

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
- [Reviewer](agents/reviewer.md);
- [Publisher](agents/publisher.md).

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
   возобновляется при наличии в репозитории evidence, что blocker устранён,
   либо только для bounded `Blocked Closure Certified` cycle, когда её exact
   attributed evidence diff удовлетворяет eligibility этого процесса. Второй
   случай запрещает product implementation, изменение blocker scope и
   активацию prerequisite; иначе autonomous continuation останавливается;
2. если clean synchronized `main` содержит exact certified checkpoint и ровно
   один записанный `Not Activated` prerequisite, Coordinator read-only
   reconstruct-ит terminal publication: checkpoint/ordered target находятся в
   ancestry `main`, exact PR имеет `MERGED` с ожидаемыми head/base/merge OID,
   local/remote task refs отсутствуют и `main == origin/main`. Только при полном
   доказательстве P0–P10 outcome blocked record считается sealed evidence, а не
   resumable active work, и только для admission этого exact prerequisite.
   Coordinator проверяет обычную readiness и создаёт новую task normal intake;
   TASK сохраняет `Blocked` и не считается Accepted/Completed. Отсутствующий,
   неоднозначный или не-Ready prerequisite останавливает selection;
3. иначе использует явную current или next task из project-state documents
   либо closure последней завершённой task после проверки readiness;
4. иначе пересекает dependency ordering текущей milestone из MASTER_PLAN,
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

- Task Contract: `Task Mode`, `Why Now`, `Definition of Done`, `Out of
  Scope` и `Verification Plan`;
- evidence выбора и отклонённые альтернативы;
- scope, non-goals, sources, roles и verification;
- branch decision и stop conditions;
- следующий candidate, который не становится active автоматически.

Новая task и её ветка начинаются только с чистого, понятного baseline и при
отсутствии другой active task. Единственное исключение — terminally published
sealed blocked-evidence record из шага 2: он не является active work только для
admission своего exact prerequisite. Для production work используется локальная
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

Для eligible blocked-evidence resume цикл заканчивается PROCESS-002, Scope
Audit, Verification, independent final Review и `Blocked Closure Certified`
вместо Coordinator Acceptance. Это не является новой task или сокращённым
product acceptance cycle.

### User Command Gates

Пользовательская модель содержит ровно три permission/gate-команды:

```text
Продолжай проект.
    -> Coordinator формирует Task Contract
    -> выполняет task, verification и независимую приёмку
    -> STOP

Разрешаю коммит.
    -> один проверенный accepted task commit либо Blocked Evidence Checkpoint
    -> STOP

Разрешаю публиковать.
    -> Publisher выполняет полный P0–P10
    -> STOP только после terminal success либо при реальном external blocker
```

Внутренние P-steps не являются пользовательскими командами или отдельными
permissions. Publisher resume-сигнал после устранения blocker не является
четвёртым gate: он не выдаёт нового разрешения и только возобновляет ранее
разрешённый immutable target.

`Blocked Closure Certified` не добавляет пользовательскую команду и не
является Coordinator Acceptance. Оно только выбирает альтернативный,
ограниченный closure/commit path, определённый ниже.

## Execution Interruption Recovery

Этот contract применяется при model usage/time limit, потере сети, закрытии
или restart session, crash процесса, restart/reboot host, tool timeout/failure,
GitHub outage/authentication failure и любом ином внешнем interruption на
любой стадии PROCESS-001.

Interruption сам по себе не является `PASS`, `FAIL`, `Approved`, `Accepted`,
`Completed`, `Blocked Closure Certified` или завершённым checkpoint.

Обязательные инварианты:

- `Started != Completed`;
- `Outcome Unknown != Failed`;
- `Outcome Unknown != Safe to Retry`;
- status claim или recovery note не заменяет independently reproducible
  evidence;
- уже доказанный checkpoint не повторяется без необходимости;
- side effect с неизвестным outcome никогда не повторяется вслепую.

### Recovery Reconstruction Gate

Новый или восстановленный агент до любой дальнейшей mutation:

1. читает repository entry contracts, branch/status/history, task index,
   единственную active task и её persistent recovery anchor;
2. проверяет происхождение baseline, staged/unstaged/untracked state, refs и
   применимое remote/GitHub state read-only;
3. связывает evidence с exact task scope и exact content identity, а не с
   именем stage или chat assertion;
4. классифицирует каждый затронутый checkpoint как:
   - `Proven Completed` — результат и все обязательные evidence доказаны;
   - `Proven Not Started` — side effect и completion evidence доказанно
     отсутствуют;
   - `Outcome Unknown` — момент interruption относительно операции либо её
     результат не доказан;
   - `Inconsistent` — наблюдаемые facts конфликтуют с task contract,
     authorization tuple или друг с другом;
5. продолжает с первого checkpoint, завершение которого не доказано.

`Proven Completed` не replay-ится. `Proven Not Started` может выполняться
только при действующих scope, prerequisites и permission. `Outcome Unknown`
сначала проходит operation-specific reconciliation ниже. `Inconsistent`
является stop condition и возвращается Coordinator без cleanup, reset или
предположения о preferred state.

Read-only reconstruction не повышает status и не создаёт verdict. Если exact
diff, scope, baseline или prerequisite изменились, агент повторяет только
затронутые и downstream checks; stale evidence не переносится.

### Persistent Recovery Anchor

До первой task mutation task record обязан содержать минимум:

- repository, Task ID/status, exact branch и trusted baseline OID;
- Task Contract, scope/non-goals, sources, roles и ordered applicable stages;
- разрешённые и запрещённые operations, stop conditions и next candidate как
  `Not Activated`;
- Existing Coverage Report и Verification Plan до test mutation;
- для каждого заявленного completed role checkpoint — exact content identity,
  command/result либо role verdict, findings и следующий stage;
- после Independent Review и Coordinator Acceptance — exact reviewed/accepted
  evidence subject, HEAD/base OID и canonical subject-manifest identity;
- до commit/publication — accepted subject/tree identity и immutable Target;
  после side effect exact local commit и Git/GitHub PR/merge/ref outcome
  reconstruct-ятся из refs/history/external repository evidence и не
  встраиваются в bytes того же commit. Их durable запись допустима только
  отдельным позднейшим authorized synchronization transition; immutable target
  не изменяется ради recovery note.

Canonical subject manifest включает новые/untracked paths без преждевременного
stage. Evidence subject — exact projected bytes, которые attests role handoff;
сам role verdict/Acceptance/evidence envelope не может self-attest.

Для любого evidence-bearing checkpoint Coordinator сначала фиксирует exact
subject path set и строит его staging-invariant canonical manifest. Subject
не является raw `git diff`: task record входит в set с projection
`task-record-v1`, каждый другой present path — с projection `full`, а
необходимый deleted path — с projection `full`, baseline mode и OID `-`.
Terminal `Recovery Evidence Envelope` является metadata envelope, а не
отдельным subject path; его bytes исключаются только через общую
`task-record-v1` projection task record. Поэтому envelope может быть
дописан только append-only после фиксации subject, не меняя его projection.
Любое изменение task record вне envelope, кроме специально исключённого
status evidence body, либо любого другого subject path изменяет identity и
invalidates затронутые verification/review/acceptance gates. Status evidence
body остаётся исключённым из manifest identity по `task-record-v1`, но его
изменение всё равно требует отдельной status/contract reconciliation и не
может изменить Blocked Closure state.

Task record использует обязательную projection `task-record-v1`:

1. headings `## Status`, `## Task Contract` и terminal
   `## Recovery Evidence Envelope` встречаются ровно по одному разу, в этом
   порядке и с начала строки;
2. envelope heading является последним top-level `##` heading; все bytes от
   его первого `#` до EOF исключаются;
3. status evidence body между окончанием строки `## Status` и первым `#`
   строки `## Task Contract` исключается;
4. projected byte stream равен: raw bytes от BOF до и включая фактический line
   terminator строки `## Status`, затем exact UTF-8 bytes
   `STATUS-EVIDENCE-EXCLUDED` и один NUL byte, затем raw bytes от первого `#`
   строки `## Task Contract` до byte перед первым `#` envelope heading;
5. никакая newline normalization, decoding, trimming или path filtering не
   применяется. Missing/duplicate/out-of-order heading либо последующий `##`
   heading после envelope делает identity `Inconsistent`.

Для каждого subject path в ascending unsigned UTF-8 path-byte order
(case-sensitive и locale-independent; сравнение выполняется по каждому byte
как `0..255`, без Unicode/case/locale collation) repository bytes
manifest содержит NUL-separated `path`, projection (`full` или
`task-record-v1`), state (`present`/`deleted`), mode и Git blob OID projected
bytes. Full present path использует current/intended Git mode и
`git hash-object --no-filters`; projected task record передаёт exact stream из
шага 4 в `git hash-object --stdin`; deleted path использует projection `full`,
baseline tree mode и OID `-`. Записи также заканчиваются NUL. Raw manifest
передаётся `git hash-object --stdin`; record сохраняет anchor HEAD, exact
subject paths/projections, path order, manifest rows, object format и OID.

Evidence envelope связывает verdict/Acceptance с subject-manifest OID, но не
объявляет hash собственных final bytes. До commit, если после interruption
append-only envelope provenance, exact subject либо отсутствие последующей
mutation нельзя независимо доказать, затронутые Review/Acceptance
классифицируются `Outcome Unknown` и выполняются повторно. После commit exact
tree/commit OID из Git history является immutable proof final
task-record/evidence bytes; task record не пытается содержать OID самого себя.
Перед commit staged tree сверяется с accepted subject, allowed evidence
envelope и exact file set. Эта двухфазная схема не создаёт self-referential
digest и не ослабляет Review/Commit Gate.

Task record является recovery anchor, но его утверждение `completed` само по
себе не доказывает completion. Chat history, незаписанный terminal output,
process memory и model memory recovery state не являются.

### Stage Reconstruction

- **Implementation / Documentation mutation:** агент сначала inspect-ит exact
  current content и полный diff. Частично применённое изменение продолжается
  от фактического состояния; patch или generator не replay-ится наугад.
- **Verification Matrix / Tester:** started command без сохранённого exact exit
  и результата считается незавершённым. Проверка безопасно запускается заново
  на неизменном content identity; material files или external state, которые
  могла изменить проверка, сначала reconciled. `PASS`/`FAIL` существует только
  как завершённый reproducible result.
- **Independent Review:** review завершён только при explicit verdict,
  findings и exact reviewed subject-manifest identity в repository handoff.
  Interruption во время чтения или анализа не создаёт verdict.
- **Rework:** фактический diff reconciled; rework invalidates все затронутые
  verification/review/scope/acceptance evidence. После rework обязательны
  повторные applicable Verification, Scope Audit и Independent Review.
- **Coordinator Acceptance:** существует только как explicit Acceptance,
  связанная с exact reviewed subject и canonical subject-manifest identity.
  Interruption между review и Acceptance оставляет Acceptance непройденной;
  interruption после Acceptance требует доказать неизменность tuple.
- **Commit Gate:** до stage/commit повторно reconstruct-ятся exact accepted или
  certified tuple, index, worktree, HEAD и permission. Partial staging не
  является завершённым commit checkpoint.
- **Publisher:** после publish permission применяется специализированный
  phase-aware Resume Reconstruction Guard. Он расширяет этот общий gate и не
  заменяется им.

### Side-Effect Reconciliation Before Retry

Для side-effecting operation с `Outcome Unknown` обязательны следующие
read-only/inspect-first proofs:

| Operation | Reconciliation before any retry |
|---|---|
| File mutation | Сравнить exact file bytes/content, expected pre/postcondition и полный diff; продолжать только отсутствующую часть без overwrite чужих changes |
| Stage | Inspect index, worktree и exact accepted/certified path set; partial/unexpected index сначала классифицировать, не считать commit выполненным |
| Commit | Inspect HEAD, parents, tree, message, log/reflog и exact accepted/certified subject-manifest identity; существующий exact commit принять как completed, duplicate commit запретить |
| Push | Inspect exact remote ref/OID; совпадающий OID доказывает completion, moved/ambiguous ref блокирует blind push |
| PR creation | Искать exact repository/head OID/base PR до create; ambiguous response не создаёт duplicate |
| Merge | Inspect exact PR state/head/base и merge OID; confirmed `MERGED` не merge-ится повторно |
| Branch deletion | Проверить существование и OID exact local/remote ref; already absent может доказать completed outcome, moved/recreated ref не удаляется |
| Documentation/status transition | Сверить actual bytes, source precedence, required evidence и все mirrors/indexes; partial либо ложное повышение status исправляется через обычный review, а не принимается как checkpoint |

Локальный side effect при отсутствии remote effect и remote side effect при
устаревшем local state классифицируются раздельно. Например, local commit без
push не доказывает P1, а remote push/merge сначала обновляет factual
reconstruction и не replay-ится из-за stale local refs.

### User Permissions After Interruption

Persistent recovery anchor хранит exact target/readiness, но не создаёт и не
подменяет пользовательское permission. Permission нельзя выводить из status,
существующего side effect или недоступной истории чата.

- interruption **до** permission gate оставляет gate непройденным;
- authority уже активированной PROCESS-001 task сохраняется до её terminal
  `STOP` только для exact task contract/scope; новый агент может применить её,
  когда current user input явно просит продолжить/resume эту active task.
  Bare `Продолжай проект.` также детерминированно возобновляет её через обычный
  preflight. Без current continue/resume input task record сам выполнение не
  запускает;
- если exact permission присутствует в текущем пользовательском вводе и tuple
  неизменен, оно применяется по обычному contract;
- для one-shot Commit Gate, если permission было дано в потерянной session, а
  commit доказанно не создан, новый агент запрашивает `Разрешаю коммит.` снова;
- если outcome commit неизвестен, сначала выполняется reconciliation; exact
  существующий commit не повторяется и новое permission не запрашивается для
  его создания post factum;
- Publisher permission сохраняется только для неизменного immutable Target по
  действующему Publisher contract. Новый агент без session history требует
  текущий explicit resume-сигнал, ссылающийся на ранее разрешённую exact
  publication; при отсутствии такой ссылки либо reconstructable Target
  требуется обычное `Разрешаю публиковать.`;
- permission, данное до изменения diff/branch/commit/base/scope, не переносится
  на изменённый tuple;
- если operation доказанно выполнена, но user report не отправлен, агент
  reconstruct-ит outcome и отправляет только truthful report/следующий
  обязательный checkpoint; operation не повторяется.

Повторный запрос permission после доказанного `Proven Not Started` не является
retry side effect. Он не разрешает расширение scope и не лечит invalidation.

Process behavior проверяется
[Execution Interruption Recovery Acceptance Scenarios](EXECUTION-INTERRUPTION-RECOVERY-ACCEPTANCE-SCENARIOS.md),
а Publisher-specific behavior — отдельными Publisher scenarios.

## Task Intake

Coordinator:

1. до проектирования, реализации или изменения тестов фиксирует Task Contract:
   `Task Mode` (`Design-only`, `Design-update`, `Implementation` или
   `Documentation-only`), `Why Now`, проверяемый `Definition of Done`, явный
   `Out of Scope` и risk-based `Verification Plan`;
2. фиксирует цель, scope, запреты и ожидаемый результат;
3. определяет затрагиваемые источники истины;
4. назначает необходимые роли;
5. создаёт или подтверждает task record;
6. применяет Size Guard;
7. до создания или изменения тестов убеждается, что Existing Coverage Report
   уже заполнен; выполнение отчёта может быть делегировано Tester или
   Developer, но gate принадлежит Coordinator;
8. останавливает intake, если задача противоречит источнику более высокого
   уровня.

Результат: однозначный task contract и список обязательных evidence.

### Size Guard

Жёсткий лимит строк или файлов не устанавливается. Coordinator обязан
остановиться и переоценить scope при одном или нескольких признаках:

- более 15 изменённых файлов;
- более 500 строк production-кода;
- более одного нового package;
- более одного нового архитектурного контракта;
- более одного независимо поставляемого поведения.

Coordinator либо доказывает целостность slice через Definition of Done и
verification, либо разделяет task до начала остальной работы.

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

Coordinator не разрешает создание или изменение тестов до Existing Coverage
Report. Tester либо Developer по назначению Coordinator:

1. до создания или изменения тестов формирует Existing Coverage Report:
   `Existing Coverage`, `Coverage Gap`, а после изменений — `Added Proof
   Tests`, `Added Regression Tests` и `Remaining Limitations`;
2. сопоставляет существующее покрытие с Definition of Done;
3. добавляет недостающие proof и regression tests в разрешённом scope;
4. применяет Verification Matrix;
5. выполняет применимые formatter, test, race, vet, lint, smoke и diff checks;
6. сохраняет точные результаты и причины недоступных проверок.

Проверка считается доказательством только при воспроизводимом результате.
Недоступная обязательная проверка не может быть молча отмечена `PASS`: отчёт
указывает точную причину, доступные substitute/stress checks и итоговый статус
`PASS WITH LIMITATION`.

### Verification Matrix

| Изменение | Обязательная verification |
|---|---|
| Concurrency, synchronization, lifecycle, cancellation, goroutines или shared state | Race detector и доступные stress checks; при технической недоступности race — `PASS WITH LIMITATION` с точной причиной |
| API, CLI, UI, configuration или production wiring | Manual smoke либо обоснованный executable proof-сценарий |
| Imports, dependencies или module boundaries | `go mod tidy`; каждое изменение `go.mod`/`go.sum` объясняется, случайный diff запрещён |
| Public API | Необходимость и godoc каждого exported identifier, proof через public behavior; иначе identifier делается unexported |
| Documentation | Применимая EN/RU parity, проверка ссылок и отсутствие противоречий нормативных источников |

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

Coordinator после каждой task явно проверяет применимость task record,
`spec/current-state.md`, зеркальных MASTER_PLAN, связанных Design Proposal,
`.ai/PROJECT_CONTEXT.md` при изменении фундаментального состояния и
`CHANGELOG.md` только для user-facing или release changes. Неприменимость
фиксируется с причиной; отсутствие diff не считается неявной проверкой.

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
Для каждого `Questionable` change record указывает:

- какой пункт Definition of Done он обеспечивает;
- почему без него task некорректна;
- почему его нельзя вынести в отдельную task.

При отсутствии любого доказательства change становится `Removable`. Final
Reviewer явно отвечает: «Можно ли удалить это изменение и сохранить
выполнение Definition of Done?»
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

### Blocked Closure Certified

Если выполнение остановлено отсутствующим prerequisite, Coordinator может
сертифицировать blocked closure вместо Acceptance только при одновременном
выполнении всех условий:

1. task record сохраняет статус `Blocked`, точный blocker, нарушенный invariant
   и отдельный следующий prerequisite как `Not Activated`;
2. Coordinator Acceptance явно записан как `не пройден` и не подменён review
   evidence;
3. полный attributed diff содержит только необходимые blocking-discovery,
   closure и project-state evidence; production code, exploratory code,
   generated, temporary, staged и untracked files отсутствуют;
4. Documentation Synchronization, применимая Verification Matrix, Scope Audit
   и независимый final Review завершены; blocking evidence признано полным и
   непротиворечивым, но product result не принят. Durable Tester handoff
   обязан быть воспроизводим из repository и содержать exact tested subject
   identity, commands с exit/results, limitations, scope/coverage counts и
   применимые proof artifacts;
5. зафиксирован immutable certification tuple: repository, Task ID, branch,
   base branch/OID, current HEAD OID, exact ordered evidence path set с
   projection/state/mode/OID rows, canonical manifest command/object
   format/OID, blocker identity, Tester handoff identity и
   verification/review results.

Canonical evidence identity — Git blob OID staging-invariant manifest stream,
а не hash от `git diff` или другой diff representation. Для exact evidence
set Coordinator строит в ascending unsigned UTF-8 path-byte order
(case-sensitive, locale-independent) NUL-separated rows:

`path\0projection\0state\0mode\0oid\0`

Для present `full` path `oid` — Git blob OID raw current/intended bytes,
полученный через `git hash-object --no-filters`; для present task record
`projection` — `task-record-v1`, а `oid` — Git blob OID exact projected stream,
переданный в `git hash-object --stdin`; для deleted path `state=deleted`,
`projection=full`, `mode` берётся из baseline tree, а `oid=-`. Raw manifest
stream хешируется через `git hash-object --stdin`; tuple сохраняет exact
ordered path/projection set, все rows, command, certified HEAD, repository
object format и manifest OID. Manifest OID, вычисленный от сырых diff bytes,
нормализованного или сокращённого set, не является взаимозаменяемым.

`task-record-v1` исключает status evidence body и terminal envelope, поэтому
certification tuple может быть дописан в envelope без self-reference и без
изменения certified subject. Envelope обязана повторить exact ordered set,
projection/state/mode/OID rows и свежий durable Tester handoff: tested
identity/manifest OID, exact commands и exit/results, limitations,
scope/coverage counts и путь к воспроизводимым evidence. До tuple append
любая mutation вне envelope требует повторить затронутые verification,
review и scope gates; envelope не объявляет hash собственных bytes.

Результат называется `Blocked Closure Certified`. Он не меняет task status,
не выполняет commit, не разрешает publication, не устраняет blocker и не
активирует следующую task. Любое изменение tuple, file set или evidence diff
аннулирует certification и требует повторных verification, review и
сертификации.

После отдельной команды `Разрешаю коммит.` альтернативный Commit Gate может
создать ровно один `Blocked Evidence Checkpoint`. Checkpoint является
evidence/transition artifact, а не task Acceptance или Completion.

## Commit Gate

Точная команда `Разрешаю коммит.` после Coordinator Acceptance разрешает
создание ровно одного проверенного task commit из принятого diff. Она не
разрешает push, PR, merge или публикацию.

Та же точная команда после `Blocked Closure Certified` разрешает ровно один
`Blocked Evidence Checkpoint` из exact certified evidence diff. В этом режиме
Coordinator обязан дополнительно подтвердить неизменность certification tuple,
статус task `Blocked`, отсутствие Coordinator Acceptance, отсутствие product
implementation и exact staged set, равный certified evidence file set.
Сообщение commit должно обозначать blocked evidence, а не completion.

Непосредственно перед commit Coordinator:

1. проверяет соответствие commit message принятой policy;
2. повторно проверяет полный exact file set;
3. убеждается, что после acceptance либо certification не появились
   неожиданные изменения;
4. исключает временные, generated и посторонние файлы;
5. повторяет `git diff --check` и применимые final checks.

GPG, DCO и sign-off не требуются, если отдельно не приняты проектом.

## Publication

Publication readiness не является publication completion. После отдельного
разрешения commit Coordinator передаёт Publisher immutable tuple: publication
class (`Accepted Task` или `Blocked Evidence Recovery`), Task ID, exact branch,
ordered commit target, base `main`, verification/scope и publication readiness.
Для `Accepted Task` target содержит exact task commit и accepted scope. Для
`Blocked Evidence Recovery` target содержит exact evidence checkpoint, его
certification tuple и при необходимости contiguous process-amendment commit
между base OID и checkpoint; task остаётся `Blocked`, а scope называется
certified, не accepted.

Сообщение, всё содержимое которого после удаления начальных и конечных
пробельных символов равно `Разрешаю публиковать.`, разрешает одну полную
публикацию этого tuple:

```text
P0 read-only preflight
    -> P1 push exact branch/OID
    -> P2 inspect then create/discover one PR to main
    -> P3 inspect checks
    -> P4 reconfirm gate and merge
    -> P5 delete exact remote branch and fetch --prune
    -> P6 checkout main
    -> P7 pull --ff-only
    -> P8 safely delete exact local branch
    -> P9 verify main/origin/main, refs and clean worktree
    -> P10 terminal report and STOP
```

Initial P0 до первой mutation проверяет clean staged/unstaged/untracked state,
current exact branch/HEAD, immutable Target `{publication class, TaskID,
repository, branch, ordered commit target, base main, scope}`, origin
URL/repository, noninteractive SSH и
`git ls-remote --exit-code origin`, `gh auth status` и repository/default
branch через `gh repo view` либо equivalent. Failure любого transport/auth/
repository subcheck внутри P0 оставляет P0 первым незавершённым, zero completed
pipeline steps и P1 not attempted.

Resume использует отдельный non-checkpoint Resume Reconstruction Guard и не
регрессирует P0. Guard сначала reconstruct-ит completed checkpoints, PR head
OID и merge OID:

- до confirmed P6 current branch/HEAD обычно остаются exact target branch/head
  OID; remote ref допустим до P4 и отсутствует после P5;
- после confirmed P6 current branch обязан быть clean `main`; task branch/HEAD
  больше не требуются, local task branch существует только до P8, remote уже
  отсутствует, `main` может отставать до P7, а equality требуется лишь P9;
- ambiguous P6 доказывается состоянием: clean current `main` означает P6
  complete, exact current task branch — P6 first unfinished.

Resume никогда не recreates/checkout-ит удалённую task branch и сначала
inspect-ит ambiguous outcomes.

P1 подтверждает `remote OID == target head OID` и немедленно переходит к P2. P2
сначала ищет exact PR по head branch/OID и base `main`, затем при необходимости
создаёт ровно один PR; ambiguous response inspect-ится до retry. Успешный push
не является terminal outcome и не разрешает запрос `Создать Pull Request`.

P3 различает `Required Success`, `No CI`, `Pending` и `Failed`. Отсутствие
`.github/workflows` либо zero registered checks фиксируется как `No CI` и не
блокирует merge при `MERGEABLE / CLEAN`. Required pending/failed блокируют P3;
checks и protection не обходятся.

После push Publisher проверяет отсутствие merge conflict. Непосредственно
перед merge он повторяет CI, exact head/base OID и mergeability; устаревший
результат P3 не используется как разрешение на P4.

P4 повторно подтверждает exact base/head OID, checks и
`MERGEABLE / CLEAN`, выполняет merge без implicit branch deletion, затем
подтверждает `MERGED` и merge commit. `UNKNOWN`, conflict, non-clean gate,
protection refusal или неподтверждённый merge блокируют P4. Merge является
checkpoint, после которого обязательны P5–P10.
Merge strategy называется явно и соответствует действующей policy
репозитория; Publisher не выбирает новую policy самостоятельно.

P5–P9 выполняют отдельный безопасный cleanup только после confirmed merge:
remote ref удаляется лишь если всё ещё равен authorized target head OID; затем
P5 выполняет `git fetch --prune`; затем P6 checkout `main`, P7 только
`pull --ff-only`, P8 local delete
через `-d` и final
verification. Recreated/moved remote ref не удаляется. Force, `-D`, reset,
rebase, non-fast-forward pull, globs и удаление unmerged/unconfirmed branch
запрещены.

External blocker не расходует publish authority. Blocker report перечисляет
completed steps, exact first unfinished P-step, factual error/check state,
известные PR/target-head/merge IDs и publication class, current
branch/HEAD/worktree/refs, unblock action
и сохранение разрешения. Команда
`Авторизация готова. Продолжай ранее разрешённую публикацию.` либо столь же
явная unblock/resume ссылка выполняет phase-aware Resume Reconstruction Guard
и продолжает первый незавершённый шаг без нового publish permission. Blind
replay uncertain mutation запрещён.

Dirty/ambiguous baseline является safety failure; изменение publication class,
branch, ordered commit target/target head OID, base или accepted/certified scope
invalidates exact authority. Они отличаются от SSH, `gh`,
repository, PR, checks, merge-gate/protection и cleanup external blockers.

P10 сообщает PR number/URL, publication class, target head commit, merge
commit, CI/checks state,
наблюдавшийся `MERGEABLE / CLEAN`, удаление remote/local task branches,
`main == origin/main` с OID, current `main`, clean worktree и затем STOP.
Отчёт только о push или merge запрещён.

Для `Blocked Evidence Recovery` все P0–P10 выполняются без послаблений. P0
дополнительно доказывает exact ordered commit range, наличие evidence
checkpoint, неизменность certification tuple и отсутствие ложного Accepted/
Completed state. P10 называет publication class и подтверждает, что task
остаётся `Blocked`. После P9 clean synchronized `main` является допустимым
baseline для последующей отдельной команды `Продолжай проект.`; prerequisite
не выбирается и не активируется Publisher автоматически.
Только read-only reconstructable terminal P10 outcome делает blocked record
sealed для узкого selection exception; post-merge documentation commit для
admission не требуется. Checkpoint без terminal publication сохраняет
dirty/active-baseline stop.

Publisher фиксирует последний успешно выполненный P-checkpoint. Blocker
report содержит exact первый незавершённый checkpoint, factual blocker,
сохранённые branch/HEAD/worktree/refs и требуемое unblock action. Resume
начинает с первого незавершённого checkpoint после read-only reconstruction.

Process behavior проверяется
[Publisher Acceptance Scenarios](PUBLISHER-ACCEPTANCE-SCENARIOS.md).

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

External interruption сначала проходит Recovery Reconstruction Gate. До
reconciliation `Outcome Unknown` не классифицируется как failure и не запускает
Rework/Blocked автоматически. После reconstruction доказанный failed check,
inconsistent state или factual external blocker обрабатывается обычными
правилами ниже.

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

## Process Health Review

Coordinator запускает лёгкий documentation-only Process Health Review:

- после каждых десяти завершённых task;
- после rollback;
- после дефекта, попавшего в `main`;
- после повторяющегося Publisher failure;
- после task, более двух раз возвращённой с review.

Review фиксирует без production metric tooling:

- `Questionable` и `Removable`, найденные Scope Audit;
- дефекты, найденные после первичной verification;
- CI failures после локального PASS;
- post-merge fixes;
- недоступные проверки;
- повторяющиеся источники лишней работы.

Результат — bounded process finding или подтверждение отсутствия изменения.
Review не создаёт fast path, обязательную статистическую систему или новую
пользовательскую команду.

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

Для `Blocked Closure Certified` этот список применяется к достоверности и
полноте evidence, а не к выполнению product Definition of Done. Task не
считается завершённой; terminal blocked closure достигается только через
certification, отдельно разрешённый checkpoint и, если разрешена, полную
P0–P10 publication с clean synchronized baseline.

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
- считать external interruption результатом stage либо завершённым checkpoint;
- считать unknown outcome failed или safe to retry;
- повторять side-effecting operation до reconciliation фактического outcome;
- считать `Blocked Closure Certified` либо `Blocked Evidence Checkpoint`
  Coordinator Acceptance, Completion или доказательством устранения blocker;
- активировать subsequent prerequisite до clean synchronized baseline после
  publication blocked evidence либо иного отдельно определённого clean path.

# Руководство по разработке с использованием LLM

## Назначение

Это руководство определяет инженерный процесс разработки с использованием языковых моделей. Оно является живым стандартом проекта для участников, сопровождающих проект специалистов, Владельца проекта и инженеров, реализующих утверждённые Design Proposals.

Проект использует структурированный процесс вместо ситуативной генерации кода, чтобы обеспечить:

- детерминированную разработку;
- прослеживаемость архитектурных решений;
- воспроизводимую реализацию;
- независимую проверку;
- долгосрочную сопровождаемость.

Использование вспомогательных средств не переносит инженерную ответственность. К каждому изменению применяются те же требования к архитектуре, scope, тестированию, review и ownership, что и к любому другому вкладу в проект.

## Основные принципы

### Сначала архитектура, затем код

До начала реализации должны быть определены предполагаемая модель компонентов, ответственность, ownership, lifecycle, инварианты и границы.

### Design Proposal как источник истины

Утверждённый Design Proposal является нормативным источником поведения и архитектуры в своём scope. Реализация должна соответствовать ему и не должна неявно переосмысливать его положения.

### Однократная компиляция

Зависящие от Configuration validation, normalization, ordering и resolution следует выполнять один раз на соответствующей границе создания. Опубликованные Runtime-структуры должны быть готовы к непосредственному выполнению без defensive recompilation на operational paths.

### Неизменяемость по умолчанию

Опубликованная Configuration, compiled structures, identity snapshots и значения message context должны быть immutable. Изменяемое состояние требует явных контрактов ownership и synchronization.

### Небольшие независимо проверяемые шаги

Реализация должна быть разделена на минимальные целостные шаги, каждый из которых компилируется, сохраняет существующее поведение и может проходить независимое review.

### Одна завершённая идея на commit

Каждый commit должен содержать одну завершённую инженерную идею, её тесты и документацию, необходимую для описания этой идеи. Частичные или несвязанные изменения не должны объединяться.

### Требования к производительности относятся к архитектуре

Ограничения аллокаций, bounded work, compilation boundaries и свойства concurrency должны быть определены до того, как реализация начнёт на них полагаться.

### Независимая проверка

Реализация и review являются разными обязанностями. Review должно непосредственно проверять состояние репозитория и доказательства, а не полагаться на отчёт о реализации.

### Blocker предпочтительнее неверной реализации

Реализация должна остановиться, если утверждённая архитектура неоднозначна, противоречива или недостаточна. Недостающие решения должны быть явно определены до продолжения разработки.

### Hot path требует явного performance review

Обработка сообщений, routing, admission и другие часто выполняемые пути требуют целенаправленного review аллокаций, synchronization, копирования данных, boundedness и скрытой работы.

## Роли

### Архитектура

Архитектурная роль:

- определяет ответственность и границы компонентов;
- назначает ownership и lifecycle authority;
- задаёт наблюдаемое поведение и инварианты;
- фиксирует существенные решения в соответствующем архитектурном документе;
- устраняет неоднозначность до начала реализации;
- утверждает или отклоняет предлагаемые архитектурные изменения.

### Реализация

Роль реализации:

- следует утверждённому Design Proposal и scope реализации;
- изучает существующий код и тесты до внесения изменений;
- не вводит скрытых архитектурных решений;
- сохраняет совместимость, если изменение явно не утверждено;
- реализует детерминированные proof tests;
- выполняет все обязательные проверки;
- точно описывает итоговое состояние репозитория.

### Review

Роль review:

- независимо сопоставляет код и тесты с утверждённой архитектурой;
- проверяет утверждения об ownership, lifecycle, concurrency, immutability и boundaries;
- подтверждает, что тесты доказывают требуемые инварианты;
- классифицирует findings по влиянию;
- различает архитектурные дефекты и дефекты реализации;
- не выдаёт approval при недостаточных доказательствах.

### Владелец проекта

Владелец проекта:

- управляет scope проекта и приоритетами roadmap;
- утверждает архитектуру и её существенные изменения;
- разрешает компромиссы уровня продукта;
- определяет, блокируют ли findings дальнейшую работу;
- разрешает commit и интеграцию;
- обеспечивает соответствие проектных записей завершённой работе.

Ни одна роль не может принимать полномочия другой роли только ради продолжения реализации.

### Publisher

Publisher интегрирует один допустимый accepted либо evidence target после явного разрешения на
публикацию. Точная команда `Разрешаю публиковать.` разрешает весь pipeline
immutable target от read-only preflight через push, Pull Request, checks,
merge, cleanup веток, синхронизированный локальный `main`, terminal report и
STOP. Push и merge являются checkpoints, а не terminal outcomes. Создание
commit остаётся отдельно разрешаемым действием.

## Стандартный процесс

```text
MASTER PLAN
    |
    v
Design Proposal
    |
    v
Architecture Review
    |
    v
Implementation Prompt
    |
    v
Implementation
    |
    v
Independent Review
    |
    v
Architecture Fix, if required
    |
    v
Commit
    |
    v
Publication
    |
    v
Post-Implementation Architecture Review
    |
    v
CHANGELOG
```

### MASTER PLAN

MASTER PLAN определяет инженерную последовательность и назначение milestone. Он выбирает проблемную область, но не заменяет Design Proposal и не предписывает будущие API.

### Design Proposal

Design Proposal определяет архитектуру в заданном scope, контракты, инварианты, модель отказов, требования совместимости и исключённую работу. Он должен содержать достаточно информации для реализации без изобретения новых решений.

### Architecture Review

Независимое Architecture Review пытается опровергнуть proposal. Блокирующая неоднозначность, пробелы ownership, недопустимые lifecycle transitions и недоказанная concurrency semantics должны быть устранены до реализации.

### Implementation Prompt

Implementation Prompt преобразует утверждённый design в один ограниченный инженерный шаг. Он задаёт разрешённые файлы и поведение, исключения, обязательные тесты, команды проверки и ожидаемый отчёт.

### Implementation

Implementation изменяет только утверждённый scope. Она включает proof tests и сохраняет репозиторий в компилируемом состоянии, пригодном для review.

### Independent Review

Independent Review проверяет фактический код и тесты по Design Proposal и Implementation Prompt. Успешная сборка или отчёт о реализации не считаются доказательством архитектурной корректности.

### Architecture Fix

Если review обнаруживает архитектурный дефект, реализация приостанавливается. Сначала анализируется модель, затем утверждается требуемое архитектурное изменение, после чего исправление поставляется отдельно от несвязанной работы. Дефекты только уровня реализации не требуют изменения архитектуры.

### Commit

Commit создаётся только после завершения реализации в заданном scope и обязательных проверок, когда findings устранены или явно приняты. Commit фиксирует одну завершённую идею.

Для evidence-only paths отдельная точная команда `Разрешаю коммит.` может
вместо этого разрешить один Blocked Evidence Checkpoint после Blocked Closure
Certified либо один Negative Disposition Checkpoint после Negative Disposition
Recorded и post-decision integrity. Это не implementation Acceptance.
Полные gates PROCESS-001 обязательны, включая exact staged-tree match;
LF/CRLF mismatch не является equivalence. Decision не выдаёт permission.

### Publication

Разрешение публикации связано с exact class, task/repository/branch,
ordered commit target, base `main` и scope соответствующего class:

- `Accepted Task`: accepted task commit/scope;
- `Blocked Evidence Recovery`: certified checkpoint/recovery-chain и scope;
- `Negative Disposition`: один exact Negative Disposition Checkpoint прямо
  поверх fixed base, его disposition tuple и negative scope.

Negative Disposition следует ND-1–ND-5
[PROCESS-001](../../engineering/PROCESS-001-AI-DEVELOPMENT-WORKFLOW.md).
Обязательны proven ownership/preservation, exact mandatory provenance blocker,
недоступность обычных Acceptance/BCC, независимо доказанный bounded Required
Recovery Exhausted, полные governance testing/review, PROCESS-002 и Scope Audit.
Известный feasible recovery route, unknown outcome обязательного источника,
product/test changes или unresolved blocking finding запрещают этот path.
Not Proven сохраняет uncertainty; Disproven сохраняет явное опровержение.
Ни одно не является positive downstream evidence или successful implementation.

Отдельная команда `Разрешаю публиковать.` по-прежнему нужна после отдельно
разрешённого checkpoint. Все classes выполняют один полный pipeline:

```text
P0 read-only preflight
    -> P1 push exact branch/commit
    -> P2 inspect then create or discover one PR to main
    -> P3 inspect required checks
    -> P4 reconfirm MERGEABLE / CLEAN and merge
    -> P5 delete the exact remote task branch
    -> P6 checkout main
    -> P7 pull --ff-only
    -> P8 safely delete the exact local task branch
    -> P9 verify main == origin/main and a clean worktree
    -> P10 full terminal report and STOP
```

Initial P0 проверяет чистое staged/unstaged/untracked состояние, current exact
task branch и commit, immutable Target, origin и noninteractive SSH access,
`gh auth status`, а также доступ к текущему GitHub repository/default branch.
Auth, transport или repository failure внутри P0 оставляет P0 первым
незавершённым, zero completed pipeline steps и P1 not attempted. Успешный push
непосредственно переходит к discovery/creation Pull Request без нового
permission или запроса `Создать Pull Request`.

Exact execution context, который будет выполнять publication, обязан доказать
usable GitHub capability двумя успешными read-only probes: decisive GitHub API
user/repository operation и Git remote authentication/read exact origin. Оба
обязательны; `gh auth status` — только supporting diagnostics. User profile path, account/helper configuration,
keyring reference, установленный credential tool или успешный probe другой
Windows identity не являются capability evidence. Non-secret evidence называет
context и результаты; credentials, tokens, headers и credential payloads
никогда не попадают в prompts, logs, task records или repository.

Если source context не обладает capability, он может выпустить Release Handoff
с unique opaque non-secret transfer ID для одной release instance и неизменным
immutable Target, затем стать observation-only. Пользователь обязан явно
маршрутизировать exact ID и Target этой ранее разрешённой publication в trusted destination. Destination выполняет `Inspect ->
Reconstruct -> Reconcile`, доказывает Target и все local/remote checkpoints,
проходит оба probe из exact context и затем фиксирует Accept Handoff с тем же
ID и Target. Accept
Handoff является procedural ownership linearization point: только destination
может resume P0-P10; source и duplicate destinations остаются
observation-only. Это exclusivity contract, а не machine lock; transfer ID
также не является secret, credential или permission.

Normative Transfer Identity — immutable tuple `{transfer ID, Target, source
execution identity, Release checkpoint snapshot}`. ID — fresh canonical
lowercase UUIDv4, отсутствующий во всех доступных operational handoff records
этой publication; он opaque и не выводится из Target, identity или времени.
Target содержит publication class, Task ID, repository/origin, exact branch,
ordered range/head OID, base OID и scope identity. Snapshot содержит P0-P10
classifications, known refs/PR/merge OID и first unfinished step. Explicit user
route и Accept event связывают exact destination identity с этими неизменными
полями.

State model имеет три независимые оси: authorization `Active | Consumed(P10) |
RevokedByUser | InvalidatedByTargetChange`; mutation ownership `Owned(context)
| InTransitNone | NoneTerminal | Unknown`; transfer attempt `Unissued |
Released | Accepted | Closed(reason)`. Release устанавливает
`Active/InTransitNone/Released`, Accept —
`Active/Owned(destination)/Accepted`.

`CancelledBeforeAccept` требует explicit user directive с exact ID и
reconciliation, доказавшего отсутствие Accept/side effect destination. Event
закрывает ID, сохраняет authorization `Active`, возвращает ownership recorded
releasing source и требует повторного capability proof. После Accept reverse
может начать только current destination-owner через release fresh ID с потерей
ownership; cancellation возвращает этого releasing destination. Factual Target
mismatch устанавливает `InvalidatedByTargetChange/NoneTerminal` и закрывает все
attempts; user revoke — `RevokedByUser/NoneTerminal`, а позже нужен новый gate.
P10 может terminalize publication только когда predecessor chain доказывает
`Active/Owned(execution-context)` у exact actor. `Released/InTransitNone`
запрещает P10 до exact Accept, установившего destination owner, либо valid
`CancelledBeforeAccept`, вернувшего releasing owner. Proven P10 затем всегда
даёт `Consumed(P10)/NoneTerminal`, recovery только report-only. Без handoff
attempt остаётся `Unissued`, publication-level P10 event использует `transfer
ID: none`; current `Accepted` закрывает этот exact attempt; already
`Closed(reason)` сохраняет reason, а P10 является отдельным event только пока
current ownership доказан. `NoneTerminal`, `Unknown` либо missing owner proof
не могут создать P10.

Release, user route, Accept и closed events образуют append-only non-secret
operational record вне immutable Target и project-state documents. Каждый event
цитирует Target, actor identity, predecessor event ID либо digest/tail,
resulting three-axis state, owner и terminal reason/disposition. Transfer event
цитирует exact ID; publication-level P10 явно цитирует `transfer ID: none` и не
фабрикует attempt.
Record обязан переживать interruption и быть independently inspectable source,
destination и Coordinator. Если он недоступен, неоднозначен или противоречив,
ownership = `Unknown` и все publication mutations STOP. Это procedural
exclusivity, а не machine или distributed lock.

Projected live-state sources сохраняют verification-stable `In Progress`.
Exact latest verdict, identity и first incomplete checkpoint берутся только из
newest valid terminal envelope entry, совпадающей с independently recomputed
canonical manifest. Missing, stale, conflicting или mismatched evidence
означает STOP и не создаёт Acceptance, commit или publication.

Handoff сохраняет authorization только для неизменных publication class, Task
ID, repository, branch, ordered commit range/head OID, base и scope. Он не
создаёт новое permission, Commit Gate или Coordinator Acceptance. Другой
release/destination, unknown/reused/mismatched/duplicate/already-accepted ID,
ambiguous ownership, отсутствие explicit user routing либо interruption до
Accept Handoff запрещают side effects. Repeat, reverse и cancellation следуют
только exact transitions выше; каждая новая attempt использует fresh ID.
Найденный при reconciliation completed
remote effect не повторяется вслепую.

Credential unavailability для exact execution identity классифицируется
отдельно от invalid/expired credentials, repository permission denial,
network/transport failure, GitHub outage и tool/session failure. Login, secret
transfer, authentication bypass и undocumented elevation не являются handoff
механизмами.

Required checks должны завершиться успешно. Отсутствие workflows или zero
registered checks фиксируется как `No CI` и не блокирует только при merge gate
`MERGEABLE / CLEAN`. Pending или failed required checks, недоказанный либо
non-mergeable PR, conflicts и branch protection refusal блокируют merge и не
могут обходиться.

External blockers не расходуют publication authorization. Blocker report
перечисляет завершённые шаги, exact первый незавершённый шаг, factual state,
сохранённые refs/worktree и требуемое unblock action. После
`Авторизация готова. Продолжай ранее разрешённую публикацию.` Publisher
использует non-checkpoint phase-aware Resume Reconstruction Guard и продолжает
без новой команды публикации. До confirmed P6 он обычно ожидает clean
task-branch/commit phase. После confirmed P6 он требует clean current `main`,
никогда не требует и не recreates task branch, допускает отставание `main` до
P7 и требует equality только в P9. Изменение target commit, branch, base или
scope вместо этого invalidates exact authorization.

После confirmed merge обязательны remote/local branch cleanup и
синхронизированный clean `main`. Cleanup использует exact branch, проверяет,
что remote ref всё ещё указывает на authorized commit, применяет только
fast-forward pull и safe local deletion и никогда не использует force, reset
или rebase. Terminal success сообщает PR number/URL, task и merge commits,
checks state, наблюдавшийся `MERGEABLE / CLEAN`, удаление обеих веток,
`main == origin/main`, clean worktree/current `main`, а затем STOP.

Для Negative Disposition P0 дополнительно проверяет exact checkpoint/base,
decision tuple, negative facts и отсутствие Acceptance/BCC/Completed claims.
Все capability, ownership/handoff, invalidation и phase-aware recovery rules
выше сохраняются. P10 доказывает только публикацию negative evidence.
Negative Disposition Recorded останавливает original work, но сохраняет
active-task barrier. Только reconstructed terminal P10, exact merged PR/ancestry,
удалённые task refs и clean synchronized main дают Sealed Negative Disposition.
Последующий отдельный ordinary intake требует собственной readiness;
publication его не активирует. Immutable checkpoint не изменяется ради записи
его публикации. До/после decision или commit отсутствующее authority не
выводится из status; unknown effects inspect-ятся до retry, changed targets
invalidates прежнее authority.
До proven P10 новое конкретное provenance evidence/pointer останавливает
следующую mutation для независимой eligibility revalidation. Active authority
недостаточно; failed eligibility запрещает remaining effects/P10/intake даже
при неизменном Git target. Нельзя форсировать cleanup ради clean baseline.
Только фактическое изменение tuple/target даёт TargetChanged. После proven
P10 новое evidence требует отдельного authorized normal intake.

### Prospective Acceptance immutable опубликованного subject

IPSPA — новый evidence event для уже опубликованных immutable Git bytes. Он
никогда не исправляет historical Acceptance: Historical Equivalence (`Proven`,
`Not Proven` или `Disproven`) остаётся независимой от prospective event.

Source — exact repository, полные commit/tree OID, optional fixed deletion
base, ordered path set и full-only canonical manifest, прочитанные через Git
object API. Working-tree, checkout/filter, normalized, decoded, archive или
diff bytes не authoritative. Отдельный Evidence Record цитирует source, но не
может входить в его tree, path set или manifest.

Обязательны fresh independent source verification, применимое Testing,
documentation synchronization, Scope Audit, Independent Review, explicit
Coordinator Prospective Acceptance и post-decision integrity. Historical
verdicts не переносятся. Mutation source либо evidence invalidates затронутые
gates и проходит inspect-first recovery. Accepted event может удовлетворить
только downstream contract, который явно называет exact source и claims, после
чего требуется отдельная repository-first reassessment. Он не создаёт четвёртый
publication class и не активирует работу автоматически.

### Post-Implementation Architecture Review

Завершённая реализация может быть проверена относительно более широкой модели компонентов, чтобы подтвердить, что локальная корректность не привела к boundary drift и не нарушила последующие допущения.

### CHANGELOG

CHANGELOG обновляется только тогда, когда репозиторий содержит завершённую capability. Он фиксирует реализованное состояние, а не предполагаемое или частично реализованное поведение.

## Правила Design Proposal

- Реализация не должна изменять архитектуру.
- Proposal должен отделять нормативные требования от примеров и будущей работы.
- Любая неоднозначность, влияющая на ownership, lifecycle, поведение, concurrency, обработку отказов или публичные контракты, останавливает реализацию.
- Архитектурные изменения требуют явного review и утверждения в соответствующем документе.
- Удобство реализации не является достаточным обоснованием изменения утверждённого инварианта.
- Отложенное поведение должно отсутствовать; placeholder API не должны создавать впечатление поддержки отсутствующей capability.
- Код является доказательством состояния реализации, но не заменяет утверждённую архитектурную модель.

## Правила реализации

- Работа должна оставаться в одном явно ограниченном scope.
- Каждый компонент и изменение должны иметь одну заявленную ответственность.
- Существующее поведение должно оставаться совместимым, если prompt не разрешает изменение.
- Публичные API не должны добавляться для предположительного будущего использования.
- Зависимости должны следовать утверждённым границам пакетов и ownership.
- Operational paths не должны получать скрытые затраты на compilation, normalization, I/O, concurrency или allocations.
- Тесты должны доказывать заданные инварианты успеха, отказов, lifecycle, concurrency и совместимости.
- Synchronization tests должны быть детерминированными и не должны использовать ordering на основе sleep.
- Все запущенные goroutines должны детерминированно освобождаться и присоединяться тестами.
- Обязательные formatting, tests, static analysis и diff checks должны пройти до сообщения о завершении.
- Итоговый отчёт должен описывать фактические изменения и результаты проверок, включая проверки, которые выполнить не удалось.

## Правила review

Reviewers должны проверять следующие аспекты, где они применимы:

- направление зависимостей и границы пакетов;
- immutable ownership и отсутствие aliasing;
- ownership ресурсов, connections, contexts и lifecycle;
- допустимые state transitions и terminal semantics;
- linearization points concurrency и exactly-once guarantees;
- bounded execution и algorithmic cost;
- аллокации, копирование, locking и скрытую работу на hot path;
- identity ошибок и наблюдаемое поведение;
- backward compatibility и default behavior;
- согласованность архитектуры, реализации, тестов и документации current state;
- соответствие явному scope и исключениям.

Review должно ссылаться на конкретные документы, файлы и тесты. Успешные formatting, compilation или tests сами по себе не доказывают ownership или архитектурную корректность. Finding должен указывать нарушенный инвариант и его влияние.

## Стандарт prompt

Каждый Implementation Prompt должен содержать следующие разделы.

### Objective

Определяет единственный результат задачи и предотвращает включение в изменение несвязанных целей.

### Scope

Определяет разрешённые поведение, пакеты, файлы и архитектурную ответственность. Он задаёт полномочия, доступные реализации.

### Out of Scope

Перечисляет смежные capabilities, которые должны отсутствовать. Это предотвращает speculative API, скрытую интеграцию и преждевременную работу над подсистемами.

### Architecture Constraints

Повторяет применимые документы-источники, правила ownership, lifecycle invariants, требования совместимости, dependency boundaries и performance constraints.

### Tests

Определяет свойства, требующие доказательства, включая случаи отказов и concurrency. Тесты должны демонстрировать инварианты, а не только выполнять строки кода.

### Verification

Перечисляет точные formatting, test, static-analysis, race, allocation, documentation и diff checks, обязательные для задачи. Недоступные проверки должны быть точно указаны и не могут представляться выполненными.

### Final Report

Определяет доказательства, необходимые для передачи результата: изменённые файлы, итоговые контракты, тесты, результаты проверок, ограничения, риски и предлагаемый commit message. Он должен описывать итоговое состояние, а не ход разработки.

## Накопленные уроки

Уроки являются объективным инженерным знанием, извлечённым из завершённых Design Proposals и их реализации. Они дополняют управляющую архитектуру, но не изменяют её.

DP-005 предоставляет следующие примеры:

- Independent Review необходимо, даже когда compilation и tests проходят успешно: оно проверяет, сохраняет ли реализация предполагаемую модель.
- Immutable compiled Runtime structures отделяют validation и preparation от выполнения при обработке сообщения.
- Normalization Configuration, resolution Handler и ordering routes относятся к construction boundary, а не к routing hot path.
- Priority ordering должно быть явным и в compiled representation, и в его proof tests; declaration order не должно становиться случайным Runtime rule.
- Поведение hot path следует подтверждать сфокусированными proofs аллокаций и concurrency в дополнение к функциональным тестам.
- Архитектурные исправления следует отделять от реализации capability, чтобы их intent, evidence и rollback boundary оставались ясными.

Новые уроки должны фиксировать повторно используемые инженерные результаты. Они не должны заменять Design Proposals, decision records, отчёты о реализации или историю проекта.

## Дальнейшее развитие

Это руководство развивается вместе с проектом. Будущие Design Proposals, implementation reviews и подтверждённые инженерные ограничения могут расширять процесс. Изменения руководства должны оставаться согласованными с архитектурой проекта, процессом принятия решений и стандартами проверки репозитория.

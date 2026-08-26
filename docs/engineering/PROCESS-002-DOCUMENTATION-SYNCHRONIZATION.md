# PROCESS-002 — Documentation Synchronization

## Purpose

Определить единый процесс синхронизации документации с фактическим состоянием репозитория.

Цель процесса:

- устранить documentation drift;
- обеспечить возможность продолжения работы новым агентом без истории чата;
- сделать репозиторий единственным источником знаний о проекте.

---

# When to Run

Процесс выполняется:

- перед началом новой функциональности;
- после завершения реализации;
- после архитектурных изменений;
- после изменения публичных API;
- после изменения модели данных;
- при передаче задачи другому агенту;
- перед `Blocked Closure Certified`;
- по запросу Coordinator.

---

# Inputs

Documentation Agent использует:

- Approved ADR;
- Active и Frozen ARCH;
- Approved и Accepted DP;
- исходный код;
- тесты;
- архитектурную документацию;
- roadmap;
- ADR;
- документацию проекта;
- результаты предыдущих задач.

История чата может использоваться только как вспомогательный источник.

Порядок источников истины, правила статусов и языковая политика определены в
[PROCESS-001](PROCESS-001-AI-DEVELOPMENT-WORKFLOW.md). Этот процесс не может
синхронизировать Approved architecture с отклоняющейся реализацией в пользу
реализации: такое расхождение возвращается Architect.

---

# Synchronization Workflow

## Step 1 — Collect Evidence

Изучить:

- текущую реализацию;
- существующую документацию;
- тесты;
- архитектурные решения.

---

## Step 2 — Detect Drift

Для каждого документа определить:

- соответствует ли он источникам истины более высокого уровня;
- отделяет ли planned state от implemented state;
- соответствует ли описание implemented state реализации и тестам;
- содержит ли устаревшую информацию;
- отсутствует ли описание новой функциональности;
- существуют ли противоречия между документами.

Все найденные расхождения фиксируются.

---

## Step 3 — Resolve Drift

Для каждого расхождения определить действие.

Возможные действия:

- обновить документацию;
- удалить устаревшую информацию;
- запросить решение Architect;
- вернуть задачу Coordinator.

Documentation Agent не принимает архитектурные решения самостоятельно.

---

## Step 4 — Update Documentation

После устранения расхождений обновляются только необходимые документы.

Изменения должны отражать:

- текущее состояние системы;
- подтверждённые архитектурные решения;
- фактически реализованное поведение.

Если архитектура утверждена, но ещё не реализована, документ сохраняет
утверждённый Design Status и отдельно указывает незавершённый Implementation
Status. Documentation Agent не повышает статус самостоятельно.

### Mandatory Applicability Record

После каждой task Coordinator и Documentation Agent фиксируют проверку:

- task record — всегда;
- `spec/current-state.md` — при factual capability, milestone или task-state
  change;
- зеркальные MASTER_PLAN — при изменении milestone boundary, engineering
  dependency или durable roadmap status;
- связанные Design Proposal — при изменении design или implementation status;
- `.ai/PROJECT_CONTEXT.md` — при изменении фундаментального состояния,
  текущей/последней task или durable operational governance;
- `CHANGELOG.md` — только для user-facing или release changes.

Документ либо синхронизируется, либо получает явное `Not applicable` с
причиной в task record. CHANGELOG не обновляется автоматически для каждой
внутренней task.

---

## Step 5 — Validate

Проверить:

- документация соответствует коду;
- документация соответствует тестам;
- реализация не представлена как источник, переопределяющий Approved
  architecture;
- EN/RU parity проверен только для дерева, где mirror обязателен согласно
  PROCESS-001;
- отсутствуют противоречия между документами;
- новый агент способен продолжить работу без истории переписки.
- active task recovery anchor содержит exact branch/baseline/scope/content
  identity, а каждый claimed checkpoint имеет independently reproducible
  evidence;
- обязательная applicability record заполнена.

---

# Publication-State Synchronization

Project-state документы сохраняют только устойчивые repository facts:

- accepted task и factual closure;
- accepted task commit либо certified blocked evidence checkpoint и
  подтверждённый merge/PR outcome после terminal publication;
- для blocked closure — неизменный статус `Blocked`, непройденный Coordinator
  Acceptance, blocker identity и prerequisite `Not Activated`;
- после terminal blocked publication — при следующем применимом PROCESS-002
  устойчивые merged PR/OID и sealed-evidence facts; их ещё не записанное
  post-merge состояние не блокирует exact prerequisite admission, если полный
  terminal outcome read-only reconstructable по PROCESS-001;
- текущую active task и product implemented/planned boundary;
- следующую рекомендацию, если она не активирована.

В них не фиксируются ephemeral Publisher states: live auth failure, pending
checks, push pending, temporary branch/worktree condition, первый
незавершённый cleanup step или инструкция «разрешён только commit». Blocker и
resume state принадлежат Publisher blocker/terminal report и при resume
реконструируются read-only из Git/GitHub.

Причина: accepted task commit либо blocked recovery-chain является immutable
publication target.
Записывать transient blocker state в этот commit после выдачи publish
authority означало бы изменить OID и invalidate разрешение.

После terminal publication Documentation Agent при следующем применимом
PROCESS-002 сверяет main/GitHub evidence и удаляет stale pre-commit/pre-merge
operational gates. Historical task closure может правдиво говорить, что на
момент closure commit или publication не выполнялись; это не является live
инструкцией после последующего merge.

Coordinator отдельно различает publication readiness, active blocked
Publisher run и terminal publication completion. Устойчивый project state не
подменяет inspect-first Publisher reconstruction.

Live blocker/terminal report хранит immutable Target, известные PR/merge OID,
completed checkpoints и phase. Phase-aware Resume Reconstruction Guard до
confirmed P6 обычно ожидает current task branch/HEAD; после P6 truthful phase
использует current `main`, допускает его отставание до P7 и не
требует/не восстанавливает task branch. Эти ephemeral checkpoint facts не
записываются изменением immutable publication target.

---

# Interruption-Recovery Synchronization

Documentation/project-state sources хранят durable recovery facts, но не
подменяют Execution Interruption Recovery gate PROCESS-001.

Task record до первой mutation хранит persistent recovery anchor: exact task,
branch/baseline, scope, roles, ordered stages и reproducible evidence только
для действительно завершённых checkpoints. После review/acceptance он хранит
exact evidence subject, exclusions и canonical subject-manifest identity
PROCESS-001, включая untracked/new files без преждевременного stage. Evidence
envelope не self-attest-ит свои final bytes; до commit не доказанная после
interruption envelope mutation возвращает affected Review/Acceptance, а после
commit final bytes доказывает tree OID. Started command, partial output, memory
session, live tool/network/auth condition и unknown outcome не
записываются как `PASS`, `FAIL`, `Approved`, `Accepted` или `Completed`.

После interruption Documentation Agent сначала сверяет actual bytes,
source-precedence, mirrors/indexes и required evidence. Partial
documentation/status mutation отражается как factual incomplete diff и
возвращается в обычные Verification/Review gates; одно наличие нового status
text не доказывает допустимый transition.

Project-state documents не хранят user permission как authority. Они могут
хранить exact accepted/certified/publication target и доказанный outcome, но
permission применяется только по текущему user input и правилам PROCESS-001.
Это предотвращает одновременно потерю factual target и перенос chat-only
authority на изменённый diff.

---

# Outputs

Результатом процесса является один из статусов.

## Synchronized

Документация соответствует репозиторию.

## Drift Detected

Обнаружены расхождения.

Новая функциональность не должна начинаться до устранения критических расхождений.

## Blocked

Невозможно определить корректное состояние проекта.

Требуется решение Coordinator или Architect.

## Blocked Evidence Synchronized

Для candidate `Blocked Closure Certified` Documentation Agent дополнительно
подтверждает, что:

- task record, task index и project-state documents одинаково сохраняют
  статус `Blocked` и точную missing-prerequisite причину;
- Coordinator Acceptance и Completion нигде не заявлены;
- следующий prerequisite обозначен только как recommendation / `Not
  Activated` и не получил неявный Task ID;
- diff содержит только стабильное blocking-discovery, closure и navigation
  evidence, без transient workstation, auth или Publisher state;
- exact evidence file set и canonical evidence digest по алгоритму PROCESS-001
  готовы для certification tuple.

Выход `Blocked Evidence Synchronized` разрешает Coordinator продолжить
certification checks, но сам не является certification, commit или
publication permission.

---

# Responsibilities

Documentation Agent:

- выявляет documentation drift;
- синхронизирует документацию;
- поддерживает актуальность знаний проекта.

Coordinator:

- определяет необходимость запуска процесса;
- принимает итоговый статус.

Architect:

- принимает решения при архитектурных противоречиях.

Developer:

- предоставляет информацию о реализации при необходимости.

Reviewer:

- подтверждает корректность синхронизации.

---

# Definition of Done

Процесс считается завершённым, если:

- все критические расхождения устранены;
- документация соответствует реализации;
- отсутствуют противоречия между документами;
- следующий агент может продолжить работу только по репозиторию;
- обязательная applicability record заполнена.

Для blocked closure дополнительно выполнен выход `Blocked Evidence
Synchronized`; product Definition of Done при этом не считается выполненным.

---

# Invariants

Запрещается:

- документировать предположения как факты;
- скрывать известные расхождения;
- смешивать планируемое и реализованное состояние;
- продолжать разработку при критическом documentation drift.

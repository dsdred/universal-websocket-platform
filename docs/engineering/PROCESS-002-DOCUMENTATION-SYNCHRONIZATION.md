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
- перед `Negative Disposition Recorded`;
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

Причина: accepted task commit, blocked recovery-chain либо Negative Disposition
Checkpoint является immutable
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

Trusted-context `Release Handoff`/`Accept Handoff` является non-secret
operational evidence, а не credential store и не новым project capability.
Immutable target commit не изменяется ради записи handoff. Source/destination
reports обязаны сохранять exact Target, execution identities, ownership phase,
dual-probe non-secret results и first unfinished P-step; tokens, passwords,
authorization headers и credential payload запрещены. Project-state документы
фиксируют сам принятый governance contract и terminal publication facts, но не
копируют transient handoff, auth или workstation state.

Operational handoff record хранится вне immutable Target/project-state и
append-only связывает exact UUIDv4 transfer ID с immutable Target, source
identity, Release checkpoint snapshot, explicit user route, destination
identity, Accept и closed events. Он обязан переживать interruption и быть
independently inspectable. Authorization `Active/Consumed/Revoked/Invalidated`,
ownership `Owned/InTransitNone/NoneTerminal/Unknown` и attempt
`Unissued/Released/Accepted/Closed(reason)` не выводятся из project-state text.
Недоступный или неоднозначный record означает ownership `Unknown` и STOP.

Projected task/context/current-state/index сохраняют verification-stable
`In Progress` и envelope-resolution rule. Они не копируют изменчивый latest
verdict/identity/checkpoint либо mutable Acceptance/commit/publication state:
current value даёт newest valid terminal envelope entry, совпадающая с
independently recomputed manifest. Stale, missing, conflicting или mismatched
evidence означает STOP. P10 evidence сначала доказывает
`Active/Owned(execution-context)` у exact actor; `Released/InTransitNone`
запрещает terminalization до exact Accept либо valid
`CancelledBeforeAccept`. Затем evidence различает no-handoff
`Unissued`/publication-level no-ID event, exact Accepted closure и сохранение
already-closed reason с отдельным P10 event при доказанном current owner.

При applicability review изменения Publisher execution governance требуют
проверки PROCESS-001, AGENT, Coordinator/Publisher roles, Publisher и general
recovery scenarios, task template и зеркальных EN/RU process guides.

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

### Durable Blocked-Closure Evidence

Для Attributed New-Record Bootstrap Recovery из PROCESS-001 Documentation
Agent сначала сохраняет lossless premutation capture и исходные raw/projection/
manifest identities. Ownership, exact byte retention и доказанный порядок
creation/prior observation -> blocker discovery фиксируются раздельно.
Отсутствующее historical chronology evidence нельзя заменить текущим capture,
timestamp, user ownership assertion либо утверждением внутри самого record.
Исторический record не переписывается как будто bootstrap уже существовал:
prospective recovery contract явно отделяется, оригинальные bytes сохраняются
losslessly, current identity считается заново без normalization.

Applicability inventory включает PROCESS-001/002, task template, general
recovery scenarios, Coordinator/Tester/Reviewer contracts, task navigation и
project state; AGENT, Publisher/scenarios и EN/RU process guides проверяются
на противоречия, а не меняются автоматически. Полный subject включает bounded
process amendment и каждый eligible record, без arbitrary untracked paths.
Подготовленная/проверенная process amendment при Not Proven provenance не
получает выход Blocked Evidence Synchronized: результат Blocked, certification
не пройдена. Смена status требует отдельной reconciliation; readiness и
historical acceptance не выводятся из process repair.

Перед выходом `Blocked Evidence Synchronized` Documentation Agent фиксирует
точный subject, который будет проверяться certification pipeline. В subject
входит task record с projection `task-record-v1`; каждый другой present
evidence path имеет projection `full`, а необходимый deleted path — baseline
mode и OID `-`. Paths и rows воспроизводятся в ascending unsigned UTF-8
path-byte order как `path\0projection\0state\0mode\0oid\0`, после чего полный
NUL-separated manifest stream хешируется `git hash-object --stdin`. Raw
`git diff`, normalized diff и сокращённый path set не являются canonical
identity.

Terminal `Recovery Evidence Envelope` task record — metadata, не отдельный
subject: его bytes исключены через общую `task-record-v1` projection. Поэтому
envelope может быть дописан только append-only после фиксации manifest; любое
изменение projected subject вне envelope invalidates downstream
verification/review/scope evidence. Status evidence body также исключён из
manifest identity, но его изменение требует отдельной status/contract
reconciliation. Durable handoff обязан перечислять exact ordered paths,
projection/state/mode/OID rows, repository/Task/branch/base/current HEAD,
object format, manifest command/OID и blocker identity. Он также обязан
содержать свежий Tester handoff с exact tested identity, командами и exit/
results, limitations, scope/coverage counts и воспроизводимыми proof
artifacts. Эти сведения должны быть доступны из repository без chat history.

Project-state documents не хранят user permission как authority. Они могут
хранить exact accepted/certified/publication target и доказанный outcome, но
permission применяется только по текущему user input и правилам PROCESS-001.
Это предотвращает одновременно потерю factual target и перенос chat-only
authority на изменённый diff.

---

# Outputs

Результатом процесса является один из статусов.

## Negative Disposition Synchronized

Отдельный выход для class C по ND-1–ND-5 PROCESS-001, не Blocked Evidence
Synchronized. Он требует согласованного полного disposition subject:

- projected scope/provenance proposition, Not Proven либо Disproven и их
  negative semantics; original result STOP, без Acceptance/BCC/Completed;
- lossless historical preservation, independently verified ownership, bounded
  Required Recovery Exhausted inventory с exact outcomes/limitations и без
  unresolved known feasible evidence retrieval;
- exact current paths/rows, неизменные task-record-v1/canonical raw identity,
  durable Evidence Verifier/Tester/Recovery Audit handoffs; отдельные Scope
  Audit, final Review и Coordinator decision ещё не выводятся из sync;
- applicability PROCESS-001/002, AGENT, Coordinator/Publisher/Tester/Reviewer,
  task template, recovery и Publisher scenarios, EN/RU process guides,
  task/index/context/current-state; остальные documents с явными причинами N/A;
- projected live In Progress и newest-matching-envelope resolution, отсутствие
  duplicated mutable decision/checkpoint; intake barrier до ND-4 P10 sealing;
- stable negative publication facts отражаются только при последующем
  применимом sync после Git/GitHub reconstruction, не edit immutable target.

General-rule synchronization/Approval оцениваются до application eligibility.
Not Proven запрещает BCC, но само по себе не запрещает этот отдельный C output,
если все ND-1 conditions доказаны. Missing ownership, preservation, recovery
exhaustion, stale identity либо blocking disposition finding дают Blocked.
Этот output не является Coordinator decision, commit или publication authority.
При неизменном subject applicable verified checks не повторяются; material
application edit invalidates affected downstream gates. Product DoD не выполнен.

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
- exact ordered evidence path set и canonical subject-manifest rows по
  алгоритму PROCESS-001 готовы для certification tuple;
- task record включён с `task-record-v1`, terminal envelope исключён из
  projected subject и не self-attest-ит собственные bytes;
- свежий durable Tester handoff точно привязан к manifest identity и содержит
  команды, exit/results, limitations, scope/coverage counts и reproducible
  evidence.
- для нового untracked record eligibility, historical chronology и lossless
  preservation по PROCESS-001 доказаны независимо; scope содержит только
  атрибутированные records и допустимый bounded process repair. Иначе результат
  Blocked, а не Blocked Evidence Synchronized.

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

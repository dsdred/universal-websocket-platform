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

---

# Publication-State Synchronization

Project-state документы сохраняют только устойчивые repository facts:

- accepted task и factual closure;
- task commit и подтверждённый merge/PR outcome после terminal publication;
- текущую active task и product implemented/planned boundary;
- следующую рекомендацию, если она не активирована.

В них не фиксируются ephemeral Publisher states: live auth failure, pending
checks, push pending, temporary branch/worktree condition, первый
незавершённый cleanup step или инструкция «разрешён только commit». Blocker и
resume state принадлежат Publisher blocker/terminal report и при resume
реконструируются read-only из Git/GitHub.

Причина: accepted task commit является immutable publication target.
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
записываются изменением accepted task commit.

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
- следующий агент может продолжить работу только по репозиторию.

---

# Invariants

Запрещается:

- документировать предположения как факты;
- скрывать известные расхождения;
- смешивать планируемое и реализованное состояние;
- продолжать разработку при критическом documentation drift.

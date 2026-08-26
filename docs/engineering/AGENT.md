# Agent Contract

## Purpose

Этот документ является детальным operational contract для любого AI-агента.
Каноническая корневая точка входа — [`AGENTS.md`](../../AGENTS.md).

Перед выполнением любой задачи агент обязан изучить данный документ.

---

# Project Context

Universal WebSocket Platform — open-source платформа для создания, настройки,
развёртывания и эксплуатации независимых WebSocket-серверов без написания
инфраструктурного кода.

После process и role contracts, но до изменения проекта агент обязан изучить:

1. [`.ai/PROJECT_CONTEXT.md`](../../.ai/PROJECT_CONTEXT.md);
2. [`spec/README.md`](../../spec/README.md);
3. [`spec/current-state.md`](../../spec/current-state.md);
4. [`spec/decisions.md`](../../spec/decisions.md);
5. все спецификации затрагиваемой подсистемы.

---

# Repository First

История чата не является источником истины.

Источником истины является репозиторий.

При противоречии между историей чата и репозиторием необходимо сообщить Coordinator о конфликте.

---

# Autonomous Continuation Entry

Сообщение, всё содержимое которого после удаления только начальных и конечных
пробельных символов равно `Продолжай проект.`, является точкой входа в
автономное продолжение по PROCESS-001. Дополнительный текст, другая команда
или отсутствие точки не получают эти полномочия автоматически.

Эта команда разрешает Coordinator:

- выполнить read-only intake репозитория;
- выбрать следующую готовую работу по правилам PROCESS-001;
- создать или актуализировать task record;
- при необходимости создать и переключиться на безопасную локальную
  task-ветку по правилам PROCESS-001.

После фиксации task contract назначенные роли могут выполнять его in-scope
изменения по PROCESS-001. В части управления git команда не разрешает
автоматически выполнять stage, commit, push, merge, удаление ветки, fetch,
pull, изменение remote или изменение `main`.

---

# Commit Entry

Сообщение, всё содержимое которого после удаления начальных и конечных
пробельных символов равно `Разрешаю коммит.`, после Coordinator Acceptance
разрешает создать ровно один проверенный task commit из принятого diff.
Альтернативно, после `Blocked Closure Certified` эта же команда разрешает
ровно один `Blocked Evidence Checkpoint` из certified evidence-only diff;
задача при этом остаётся `Blocked` и не считается Accepted или Completed.

Перед commit Coordinator выполняет Commit Gate PROCESS-001: повторно проверяет
message policy, exact file set, отсутствие post-acceptance, временных,
generated и посторонних changes и применимые final checks. Для blocked
closure вместо post-acceptance проверяется неизменность certified tuple и
отсутствие product implementation. Команда не
разрешает push, PR, merge или публикацию. GPG, DCO и sign-off не добавляются
без отдельной policy проекта.

---

# Publisher Entry

Сообщение, всё содержимое которого после удаления начальных и конечных
пробельных символов равно `Разрешаю публиковать.`, после отдельно разрешённого
и созданного accepted task commit либо `Blocked Evidence Checkpoint` даёт
Publisher одно разрешение на весь exact pipeline
`preflight -> push -> create/discover PR -> inspect checks -> merge -> cleanup
-> synchronized main -> terminal report -> STOP`.

Разрешение связано с publication class, exact branch, ordered commit target,
base `main` и accepted либо certified scope. Для blocked recovery target может включать
предшествующий process-amendment commit и exact evidence checkpoint; это не
создаёт Coordinator Acceptance. Push и merge являются checkpoint: здоровый
pipeline продолжается без
дополнительного запроса. Реальный внешний blocker сохраняет разрешение и
позволяет inspect-first phase-aware resume командой
`Авторизация готова. Продолжай ранее разрешённую публикацию.` без повторного
`Разрешаю публиковать.`. Изменение target tuple или scope invalidates
разрешение.

Initial P0 требует current exact target branch/head commit. Resume Reconstruction
Guard reconstruct-ит checkpoints по immutable Target и после confirmed P6
ожидает clean current `main`, а не target branch/commit. Полный P0–P10, blocker
report, CI/merge gate, safe cleanup и terminal evidence определены в
[PROCESS-001](PROCESS-001-AI-DEVELOPMENT-WORKFLOW.md) и
[Publisher contract](agents/publisher.md).

Auth/transport/repository failure внутри initial P0 оставляет P0 первым
незавершённым и P1 not attempted.

---

# Read Before Work

После перехода из корневого `AGENTS.md` агент обязан изучить:

1. [PROCESS-001](PROCESS-001-AI-DEVELOPMENT-WORKFLOW.md)
2. [PROCESS-002](PROCESS-002-DOCUMENTATION-SYNCHRONIZATION.md)
3. документ своей роли в [`agents/`](agents/)
4. документацию затрагиваемой подсистемы

Только после этого разрешается приступать к задаче.

---

# External Interruption Entry

После model/time limit, потери сети, закрытия/restart session, crash/tool
failure, restart host либо неизвестного результата внешней операции агент не
продолжает с remembered stage и не повторяет mutation.

До любой дальнейшей mutation он выполняет общий **Execution Interruption
Recovery** gate PROCESS-001: reconstruct-ит task/branch/baseline/diff и
local/remote facts, классифицирует checkpoints и продолжает первый, completion
которого не доказан. `Started` не означает `Completed`, interruption не создаёт
verdict/status, а unknown side effect сначала reconciled inspect-first.

Chat history не является recovery state. Task record обязан давать persistent
anchor, но status claim без independently reproducible evidence не доказывает
checkpoint и не выдаёт user permission. Для Publisher сохраняется его более
строгий phase-aware Resume Reconstruction Guard.

Authority exact active task сохраняется через interruption только до её
terminal `STOP` и только в записанном scope. Новый агент применяет её после
current explicit user request продолжить/resume эту task; task record без
такого current input не запускает выполнение. Commit/Publisher permissions
сохраняют отдельные более строгие правила PROCESS-001.

---

# Responsibilities

Каждый агент работает только в рамках своей роли.

Описание ролей находится в [`docs/engineering/agents/`](agents/).

---

# Documentation

Любое изменение должно сопровождаться анализом необходимости изменения документации.

При обнаружении documentation drift задача возвращается Coordinator.

---

# Architecture

Архитектурные решения принимает только Architect в пределах утверждённого
процесса и передаёт их Writer или Documentation Agent для фиксации.

Developer не изменяет архитектуру самостоятельно.

---

# Implementation

Изменения должны быть небольшими.

После каждого значимого изменения должна существовать возможность проверки.

Агент также обязан:

- не выходить за scope текущей задачи и milestone;
- сохранять Technology Neutrality до утверждённого решения;
- фиксировать значимые архитектурные решения в зеркальных EN/RU ADR;
- обновлять `spec/current-state.md`, когда capability существенно меняется;
- не добавлять secrets, credentials, generated artifacts или локальные
  environment files;
- не выполнять commit без явного указания.

---

# Testing

Перед завершением задачи агент обязан:

- найти существующие тесты;
- оценить покрытие;
- создать недостающие тесты;
- выполнить проверки проекта.

---

# Completion

Задача считается завершённой только после успешного завершения PROCESS-001.

---

# Main Rule

Если возникает существенная неопределённость, агент не документирует
предположение как факт. Он останавливает затронутый этап и сообщает
Coordinator.

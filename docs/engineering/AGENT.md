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

Третий отдельный class по ND-1–ND-3 PROCESS-001: после `Negative Disposition
Recorded` и post-decision integrity эта же точная команда разрешает ровно один
`Negative Disposition Checkpoint`. Exact decision tuple и staged-tree match
обязательны; это не Acceptance/BCC/Completed. Без отдельной команды stage и
commit запрещены. LF/CRLF mismatch не получает equivalence.

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
и созданного accepted task commit, `Blocked Evidence Checkpoint` либо
`Negative Disposition Checkpoint` даёт
Publisher одно разрешение на весь exact pipeline
`preflight -> push -> create/discover PR -> inspect checks -> merge -> cleanup
-> synchronized main -> terminal report -> STOP`.

Разрешение связано с publication class, exact branch, ordered commit target,
base `main` и accepted, certified либо negative disposition scope. Для blocked recovery target может включать
предшествующий process-amendment commit и exact evidence checkpoint; это не
создаёт Coordinator Acceptance. Push и merge являются checkpoint: здоровый
pipeline продолжается без
дополнительного запроса. Реальный внешний blocker сохраняет разрешение и
позволяет inspect-first phase-aware resume командой
`Авторизация готова. Продолжай ранее разрешённую публикацию.` без повторного
`Разрешаю публиковать.`. Изменение target tuple или scope invalidates
разрешение.

Class `Negative Disposition` содержит один exact checkpoint непосредственно
поверх base и его disposition tuple; обычные P0–P10, context/ownership,
invalidation и recovery сохраняются. P10 означает публикацию отрицательного
checkpoint, не Acceptance/BCC/Completion. Только reconstructed terminal P10 на
clean synchronized main даёт `Sealed Negative Disposition`, после которого
допустим отдельный normal intake без positive prerequisite proof. До этого
projected In Progress не разрешает resume original work или новый intake.

Initial P0 требует current exact target branch/head commit. Resume Reconstruction
Guard reconstruct-ит checkpoints по immutable Target и после confirmed P6
ожидает clean current `main`, а не target branch/commit. Полный P0–P10, blocker
report, CI/merge gate, safe cleanup и terminal evidence определены в
[PROCESS-001](PROCESS-001-AI-DEVELOPMENT-WORKFLOW.md) и
[Publisher contract](agents/publisher.md).

Auth/transport/repository failure внутри initial P0 оставляет P0 первым
незавершённым и P1 not attempted.

### Publisher Execution Context

Publisher выполняет side effects только из exact execution context, который
до них успешно прошёл оба обязательных read-only capability probe: decisive
GitHub API user/repository access и Git remote authentication/read для exact
origin. `gh auth status` — supporting diagnostics, не decisive proof.
Совпадение `USERPROFILE`, account metadata, `gh` configuration,
`credential.helper`, keyring reference или credentials другой Windows identity
не доказывает capability.

Если текущий context неспособен пройти оба probe, Publisher не запрашивает
secret через prompt, не записывает credential в repository и не использует
непредусмотренную elevation. Уже разрешённый immutable Target может быть
передан только через trusted-context Release Handoff PROCESS-001: явная
маршрутизация пользователя с exact unique non-secret transfer ID и immutable
Target, destination `Inspect -> Reconstruct -> Reconcile`,
оба успешных probe и затем `Accept Handoff`. Release Handoff не является новой
publication permission; Accept Handoff цитирует тот же transfer ID и Target.
Unknown, reused, mismatched, duplicate либо already-accepted ID fail closed.
Source становится observation-only, а после Accept Handoff ровно destination
владеет P0-P10. ID не является secret или machine lock; exclusivity —
обязательный procedural contract.

Normative Transfer Identity, UUIDv4 freshness/uniqueness, immutable Target и
Release snapshot, независимые authorization/ownership/attempt axes и
append-only durable operational record определены PROCESS-001. Любой closed
attempt фиксирует reason и authorization/owner disposition; недоступный или
неоднозначный record означает ownership `Unknown` и запрещает все publication
mutations.

P10 требует доказанного `Active/Owned(execution-context)` у exact actor;
`Released/InTransitNone` запрещает terminalization до exact Accept либо valid
`CancelledBeforeAccept`, вернувшего releasing owner. Proven P10 всегда
consumption/no-owner, но не фабрикует transfer: no-handoff остаётся `Unissued`,
current Accepted закрывает exact attempt, already-closed reason сохраняется
отдельным publication event только при доказанном current owner. Projected
live state остаётся `In Progress`; newest valid envelope matching recomputed
manifest является единственным latest checkpoint source, mismatch/conflict
означает STOP.

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

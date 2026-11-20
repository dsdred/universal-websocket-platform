# TASK-143 — Диалоги подтверждения Studio

## Status

`In Progress`. Текущий gate определяется Recovery Evidence Envelope.

## Task Contract

- Режим: Implementation. Owner активировал пункт 4
  [UX-плана TASK-138](TASK-138-STUDIO-OWNER-ACCEPTANCE.md). Номер TASK-143
  свободен в task registry и Git refs. TASK-141 (темы) остаётся отдельной.
- Один диалог Studio на EN/RU для ухода по внутренним ссылкам с правками
  и существующих действий сброса. Cancel/Escape сохраняют выбранные
  ресурсы, Draft, раздел и поля. Confirm выполняет одно выбранное действие.
- Нативный `beforeunload` сохраняет защиту reload, закрытия вкладки и
  внешнего перехода. Уход по внутренней ссылке после явного подтверждения
  сбрасывает подтверждённые локальные правки перед переходом.
  Постоянный или временный обход `beforeunload` отсутствует.
- Браузерная проверка на 1280/390 px, EN/RU, мышь/touch, клавиатура,
  повторная активация, Cancel/Confirm, прерванный переход и native reload.
- Вне scope: API, Runtime, TASK-131, TASK-134, темы и сервера Owner.
  TASK-142 Owner PASS сохраняется. Stage, commit и publication требуют
  отдельных команд Owner. STOP на Pre-Commit Gate.

## Recovery anchor

- Ветка `feature/task-143-styled-confirmations`. Чистый `main == origin/main`
  на `11003f2af44bf661d5105e843c6a0015bb124309`, PR #142.
- Источники: [TASK-142](TASK-142-RESPONSIVE-NAVIGATION.md), TASK-138,
  существующие Studio и `beforeunload`. Порядок: intake → Architecture →
  implementation → verification → PROCESS-002 → independent review → STOP.
- Coordinator назначил основную сессию на реализацию и тесты. Architect
  и Reviewer работают независимо.

## Architecture Confirmation

Независимый Architect: `APPROVED`, 0 blocking findings. Нативный
`<dialog>`, один отложенный action, сохранение активного состояния до
Confirm. Для подтверждённого ухода очищаются только уже подтверждённые
к сбросу правки. Неопределённый Save требует свежего GET при повторном
открытии Draft. Обработчик `beforeunload` не отключается. ADR/DP и
изменение публичных контрактов неприменимы.

## Existing Coverage Report

151 Node-тест проверяет локализацию, раскрытие навигации, состояние Draft,
native discard, pending save и неизвестный результат. Go проверяет
embedded delivery. Новые доказательства: один диалог, точный отложенный
переход, прерывание, стандартные ссылки, фокус и нативная защита reload.

## Scope audit и PROCESS-002

12 Required, 0 Questionable, 0 Removable. Четыре frontend-ресурса Studio
и каталога переводов меняют только подтверждения. Один Node-файл покрывает
новое поведение и существующие сценарии. Семь документов покрывают intake,
project context, task index, current state, CHANGELOG и Quick Start EN/RU.
План TASK-138 и опубликованная запись TASK-142 остаются историческими
источниками. MASTER_PLAN, ADR/DP, `spec/decisions.md` и корневые README
неприменимы: вехи, зависимости, API и порядок запуска сохраняются.

Для Confirm на переходе локальный сброс происходит до попытки перехода.
Если переход прерван, значения предыдущего чтения остаются в редакторе.
Если результат предыдущего Save неизвестен, редактор закрывает выбор
Draft и перед новым Save требует повторного открытия со свежим GET.
Cancel/Escape выполняется до локального сброса. Busy блокирует внутренний
переход до завершения сохранения и сохраняет браузерную защиту выхода.

## Recovery Evidence Envelope

### Pre-Commit Report — 2026-11-20

- Coordinator Acceptance: `Accepted` для Pre-Commit Gate. Ветка
  `feature/task-143-styled-confirmations`, базовый HEAD
  `11003f2af44bf661d5105e843c6a0015bb124309`. Canonical SHA-1
  `e85a2d6eb735640fd5cf96e7e54ebf9b2862c8ec`. Scope Audit 12/0/0.
  STOP на Pre-Commit Gate. Index пуст.
- Независимый Reviewer: `APPROVED`, 0 blocking и 0 unresolved findings.
  Reviewer независимо воспроизвёл exact 12-path manifest, 155 Node-тестов,
  Go-пакет Studio, семь относительных ссылок в документации и diff check.
  Замечания по ссылке на текущую страницу в desktop и озвучиванию
  предупреждения о неизвестном Save устранены. Reviewer проверил
  доказательства браузерного smoke и полного Go suite для того же subject.
- Node 155/155. Команда `node --test --test-reporter=dot internal/configurationstudio/studio_test.js internal/configurationstudio/dashboard_test.js internal/configurationstudio/control_navigation_test.js`.
  `go test -mod=readonly -p 6 -count=1 ./...`,
  `go vet -mod=readonly ./...`, сборка отдельного Control Service через
  `go build -mod=readonly -buildvcs=false` и
  `git -c core.autocrlf=false -c core.whitespace=cr-at-eol diff --check` прошли.
- Chrome 158.0.8127.36, изолированный Control Service на 18144. EN при 1280 px,
  RU при 390 px. Проверены мышь, touch, Tab, Shift+Tab, Enter, Escape,
  возврат фокуса, отсутствие overflow, Cancel, повторная активация,
  переход Studio → Dashboard через Confirm и сброс внутренней смены раздела
  через Cancel. Отмена сохранила Workspace, Configuration, Draft, раздел
  и введённые значения. Переход по Confirm загрузил Dashboard один раз.
- CDP прервал один подтверждённый переход. Studio осталась открытой со
  значениями предыдущего чтения. Новая правка снова вызвала нативный
  beforeunload при reload. Ещё два нативных предупреждения проверены
  на reload и внешней ссылке. Закрытие вкладки покрыто тем же обработчиком
  и отдельно в браузере не воспроизводилось. Браузерных записей API и
  JS-ошибок не было.
- PID Owner-инстансов 18138–18143 до и после проверки совпали
  (`37772`, `21044`, `24148`, `30028`, `26932`, `26260`).
  TASK-142 Owner PASS сохраняется. TASK-141 остаётся отдельной.

#### Exact file manifest

Пути отсортированы по исходным UTF-8 bytes. Каждая строка сериализуется
как `path\0projection\0state\0mode\0oid\0`. Итоговый поток хешируется
через `git hash-object --stdin`. Git object format: SHA-1.

```text
.ai/PROJECT_CONTEXT.md | full | present | 100644 | 117c7bb0e92c6cbb105762a20f1cc61f9418b770
CHANGELOG.md | full | present | 100644 | 70dc44525d307afef33e6984a4b7338b17987842
docs/en/quick-start.md | full | present | 100644 | 58c80f0b124179f64d2c3ba228431b9670103e73
docs/ru/quick-start.md | full | present | 100644 | 644868bef058a5b0f9bfcabf8d9edfc1ede47f33
docs/tasks/README.md | full | present | 100644 | abc1d4aaa678e801763955bfbaee2ce592c41f19
docs/tasks/TASK-143-STYLED-CONFIRMATION-DIALOGS.md | task-record-v1 | present | 100644 | 659207f9f90752738dd6c0651b9843a9bee6acd8
internal/configurationstudio/assets/control-language.js | full | present | 100644 | b89196f4d33dbef1a8812df0a70b676861babce7
internal/configurationstudio/assets/index.html | full | present | 100644 | 1eaf9cc08dc454ffdb009f2630f83a098a1f34ee
internal/configurationstudio/assets/studio.css | full | present | 100644 | 9d852511c5f12fc5233e25d3bfab3bdb870d8dcc
internal/configurationstudio/assets/studio.js | full | present | 100644 | 8e3e531a286e972b9566be3aac3364f96bbd7512
internal/configurationstudio/studio_test.js | full | present | 100644 | 20fbe12134421de678d2332780916c037370b112
spec/current-state.md | full | present | 100644 | 402ba079296e562eeb1010ed793743f9d4ce26da
```

### Owner Acceptance — прерванный переход

Owner подтвердил действующую семантику TASK-143. Cancel (включая Escape)
сохраняет несохранённые поля, выбранный Draft и раздел редактирования.
Confirm явно отбрасывает локальные правки и инициирует выбранный переход
один раз. Если переход прерван и Studio остаётся открытой, отброшенные
правки автоматически не восстанавливаются. Нативный обработчик
`beforeunload` остаётся активным и защищает последующие правки. Owner
принял это ограничение и отклонил временную копию и восстановление правок.

- Target: ветка `feature/task-143-styled-confirmations`, HEAD
  `11003f2af44bf661d5105e843c6a0015bb124309`, canonical SHA-1
  `e85a2d6eb735640fd5cf96e7e54ebf9b2862c8ec`. Те же 12 строк в разделе
  [Exact file manifest](#exact-file-manifest) выше.
- Запись добавлена в Recovery Evidence Envelope по PROCESS-001.
  Coordinator Acceptance (`Accepted`), независимый Review (`APPROVED`)
  и результаты Pre-Commit Report относятся к тому же subject.
- STOP на Commit Gate. Index пуст. Решение Owner о поведении
  не является разрешением на stage, commit или publication.

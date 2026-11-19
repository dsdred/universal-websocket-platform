# TASK-140 — Корректировка переключателя языка по замечанию Owner

## Status

`In Progress`. Текущий gate определяется последней применимой записью
Recovery Evidence Envelope.

## Task Contract

- Task Mode: Implementation, ограниченная корректировка интерфейса в рамках
  ручной приёмки TASK-140. Owner отклонил подпись «Language» и выпадающий
  список опубликованного интерфейса.
- Definition of Done: одна кнопка с иконкой языков и текущим кодом EN/RU
  в том же месте в обеих шапках. Нажатие мышью, Enter и Space переключают
  язык. Доступное имя отражает текущий язык и действие, фокус виден. На
  ширине 390 px кнопка доступна. Смена языка сохраняет текущий Draft,
  несохранённые поля, ошибки и предпочтение без перезагрузки или запросов.
- Out of Scope: прочие элементы интерфейса, API, Runtime, запуск задач
  TASK-141 и следующих, обновление действующих Owner-инстансов.
- Verification: существующие и целевые Node-тесты, Go-проверки, Chrome
  smoke на desktop/390 px с сетевым счётчиком, независимый review.

## Recovery anchor

- Репозиторий Universal WebSocket Platform. Ветка
  `feature/task-140-owner-language-toggle` от чистого опубликованного
  `main == origin/main` на `4fcbbacfdfacbd8a39b4ab2233136bb2cb546d29` (PR #139).
- Источники: [опубликованная TASK-140](TASK-140-CONTROL-LOCALIZATION.md),
  [ADR-0001](../en/adr/0001-bootstrap-control-service.md),
  [Quick Start (EN)](../en/quick-start.md) и [Quick Start (RU)](../ru/quick-start.md).
- Coordinator назначил реализацию, тесты и синхронизацию основной сессии.
  Архитектурная проверка и независимый review выполняются отдельно.
- Этапы: intake → Architecture Confirmation → implementation → verification
  → PROCESS-002 → scope audit и independent review → Pre-Commit STOP.
  Stage, commit и publication требуют отдельных Owner gates.
- TASK-141 Not Activated. TASK-134, TASK-131, TASK-102, Gate 11 и защищённые
  этапы вне scope. Неожиданный diff, блокирующая проверка или новая
  область работ приостанавливают затронутый этап.

## Architecture Confirmation

Независимый Architect: `APPROVED`, 0 blocking findings. Нативная
`button type="button"` меняет текущий язык через существующий обработчик
отрисовки и хранилище. Встроенная декоративная иконка не требует HTTP.
Элемент использует нативную клавиатурную активацию и общий focus-visible.
Существующие контракты хранения, маршрутов и данных действуют. Новый ADR/DP
и отдельная предреализационная спецификация не применяются.

## Existing Coverage Report

148 Node-тестов, включая локализацию Draft, ошибки Listener/Timeouts,
обратную навигацию, Dashboard, сохранение языка и fallback. Go проверяет
встроенные ресурсы. Пробелы: семантика новой кнопки, доступное имя,
нажатие мышью и клавиатурой, фокус и одинаковая геометрия шапки при 390 px.
Текущие выверенные сценарии переключения должны выполняться через click.

## Scope audit и PROCESS-002

13 Required, 0 Questionable, 0 Removable. Один существующий пакет и одна
корректировка интерфейса. Четыре файла production меняют только элемент
в шапке, его стиль и обработчик. Два файла Node-тестов адаптируют сценарии к
нажатию кнопки и проверяют видимый код, доступное имя и сохранение состояния.
Семь файлов документации включают task record, task index, project context,
current state, changelog и зеркальный Quick Start EN/RU.
Историческая запись публикации TASK-140 сохранена. MASTER_PLAN, ADR/DP,
`spec/decisions.md` и корневые README неприменимы: зависимости, нормативные
решения и способы запуска сохраняются. Точные пути, дерево baseline и
проверяемый манифест приведены в Recovery Evidence Envelope.

## Recovery Evidence Envelope


### Pre-Commit Report — 2026-11-19

- Coordinator Acceptance: `Accepted` для `adf2ea0f01dde16c6b5d5d59c13c915101315e9d`.
  Branch `feature/task-140-owner-language-toggle`, baseline и HEAD
  `4fcbbacfdfacbd8a39b4ab2233136bb2cb546d29`. Index пуст. STOP на Pre-Commit Gate.
- Независимый final Reviewer: `APPROVED`, 0 findings. Пересчитал все 13
  строк манифеста, повторил Node и diff check. Go-пакет Studio,
  15 относительных ссылок и парность заголовков Quick Start проверены.
- TASK-141 остаётся Not Activated. Сервисы на 18138, 18139 и 18140
  продолжают работу с исходными PID 37772, 21044 и 24148.

| Проверка | Результат |
|---|---|
| `node --test --test-reporter=dot internal/configurationstudio/studio_test.js internal/configurationstudio/dashboard_test.js` | Exit 0, 148/148. Открытый Draft, локальная проверка таймаутов, отказ Listener, результат Save, сохранение Dashboard, fallback и восстановление из кэша истории. |
| `go build -mod=readonly -buildvcs=false -o <temp-exe> ./cmd/control-service` | Exit 0, сборка временного проверочного инстанса. |
| `go test -mod=readonly -p 6 -count=1 ./...` и `go vet -mod=readonly ./...` | Exit 0. Новых путей Go concurrency нет, отдельный race run неприменим к изменению интерфейса. |
| `git -c core.autocrlf=false -c core.whitespace=cr-at-eol diff --check` | Exit 0. Новый task record отдельно проверен на пробелы и окончания строк. |

Chrome 157.0.8103.51 через CDP, отдельный сервис на 18142 и чистый
профиль. Для повтора собрать текущую ветку, открыть обе страницы,
нажать кнопку мышью, перейти к ней с Tab и переключить язык Enter
и Space. Проверить доступное имя, видимость рамки и текущий EN/RU
после каждого действия. Проверка на 1280 и 390 px подтвердила
совпадение размеров и стиля кнопки в обеих шапках на 390 px,
помещение в viewport и отсутствие горизонтального overflow страницы.

При открытом Draft проверены неизменные выборы ресурсов и секции,
введённый Listener, ошибка IPv4, ошибка диапазона таймаута и
перевод результата Save. Смена языка после загрузки Dashboard
сохранила выборы и записи таблицы. Refresh и переходы между
страницами сохранили RU. CDP зафиксировал 13 переключений и 0
запросов при них, четыре запроса документов от явных переходов и
refresh, 0 JS exceptions и 12 ожидаемых API-записей от явных действий
создания и Save. Шесть встроенных ресурсов совпали с checkout
byte-for-byte. Проверочный сервис на 18142 остановлен.

#### Exact file manifest

Object format SHA-1. Пути отсортированы по raw UTF-8. Строки
`path\0projection\0state\0mode\0oid\0` хешируются `git hash-object --stdin`.
OID файлов `full` вычисляются без фильтров; task record использует
проекцию `task-record-v1` PROCESS-001.

```text
.ai/PROJECT_CONTEXT.md | full | present | 100644 | 624e8a9022b6bb36af1a53dfdf91d2a7c1ebd8be
CHANGELOG.md | full | present | 100644 | 0d3d64e4a624b8652e7a689c38a33726fc8550dd
docs/en/quick-start.md | full | present | 100644 | 340d9abdbf04f3beed89787ef43a92220073c3ed
docs/ru/quick-start.md | full | present | 100644 | 622f789fea476e8c250010b341401d97a8d48190
docs/tasks/README.md | full | present | 100644 | 93f24319d04ddcf983b29881513f400b7b2c7e1c
docs/tasks/TASK-140-OWNER-UX-CORRECTION.md | task-record-v1 | present | 100644 | ba5d364c0aeefe89b13f40dc3db965a3233a1526
internal/configurationstudio/assets/control-language.js | full | present | 100644 | d56d5d55c1c58d594f2606a55693c6298082d89e
internal/configurationstudio/assets/control.css | full | present | 100644 | 75ebcfd1e46902dd8cc3e0f2189f68cc7abfab3c
internal/configurationstudio/assets/dashboard.html | full | present | 100644 | 309fed0db5ed45e67f01552af50d31455521a592
internal/configurationstudio/assets/index.html | full | present | 100644 | a69507d2cd8c7b5962da12c0bdd34bb74710222a
internal/configurationstudio/dashboard_test.js | full | present | 100644 | d1dcad459bb1f4446d1bb038eb7b34ca63f0c5cb
internal/configurationstudio/studio_test.js | full | present | 100644 | 56b6a31f9e862995c192af625453dcf522b3fbec
spec/current-state.md | full | present | 100644 | f4fa42c6d12eda5a8e140995c77c73bd6392291b
```

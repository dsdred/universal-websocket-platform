# TASK-142 — Адаптивная навигация Studio и Dashboard

## Status

`In Progress`. Текущий gate определяется Recovery Evidence Envelope.

## Task Contract

- Режим: Implementation. Owner указал на многострочную мобильную навигацию
  и низкое расположение EN/RU после TASK-140. Выбран пункт 3
  [утверждённого UX-плана TASK-138](TASK-138-STUDIO-OWNER-ACCEPTANCE.md).
- Нумерация: task index, task records, локальные и tracking refs и история
  проверены. TASK-141 упоминается как Not Activated в записях TASK-140,
  отдельное назначение отсутствует. Этот номер не переиспользуется.
  Свободный TASK-142 назначен адаптивной навигации. Темы вне scope.
- Результат: при ширине до 700 px один глобальный ряд с логотипом, UWP Control,
  EN/RU и кнопкой навигации. Меню по умолчанию закрыто, текущая страница
  выделена. Клик, touch, Enter/Space, Escape, закрытие после выбора и
  возврат фокуса работают. На 390 px нет горизонтального overflow.
- На 1280 px сохранены боковая навигация и EN/RU в шапке страницы.
  Сохранены localStorage и защищённые несохранённые правки Draft.
  Текущий URL в мобильном меню закрывает его без перезагрузки, переход
  с изменённым Draft использует существующий браузерный prompt.
- Вне scope: API, Runtime, темы, TASK-134, TASK-131, TASK-141 и действующие
  инстансы Control Service. Commit и publication требуют отдельных команд.

## Recovery anchor

- Чистый `main == origin/main`, `4ce41f98b64d74efb2928fdcfbb9b7ffa8b38583`
  (PR #141). Ветка `feature/task-142-responsive-navigation`.
- Источники: TASK-138, [TASK-140 Header Cleanup](TASK-140-OWNER-HEADER-CLEANUP.md),
  [ADR-0001](../en/adr/0001-bootstrap-control-service.md), существующие
  HTML/CSS/JS и правила защиты Draft.
- Coordinator назначил основную сессию на реализацию, тесты и документацию.
  Architect и Reviewer независимы. Intake → Architecture → implementation
  → verification → PROCESS-002 → scope audit и review → Pre-Commit STOP.

## Architecture Confirmation

Независимый Architect: `APPROVED`, 0 blocking findings. Общий нативный
disclosure управляет существующим nav через `hidden` и `aria-expanded`.
Один элемент EN/RU перемещается по breakpoint. Ссылки и `aria-current`
сохраняются. Cross-page переход использует прежний `beforeunload`.
Новый ADR/DP неприменим. Stage, commit и publication вне текущего gate.

## Existing Coverage Report

148 Node-тестов Studio/Dashboard покрывают локализацию, состояние Draft,
GET-only Dashboard и API-последовательности. Go проверяет доставку
встроенных ресурсов. Пробелы: новый asset, состояние раскрытия, клавиатура
и фокус, touch, смена viewport, фактические переходы и отмена ухода с Draft.
План: Node, Go, Chrome 1280/390 EN/RU с нативными событиями и счётчиком сети.
В действующие инстансы тесты не записывают.

## Scope audit и PROCESS-002

15 Required, 0 Questionable, 0 Removable. Две страницы переиспользуют
существующие ссылки, кнопку языка и breakpoint 700 px. Общие CSS и JS
управляют глобальной шапкой, disclosure, Escape, фокусом и возвратом
из истории. Перевод добавляет доступное имя кнопки EN/RU. Go изменён
только для доставки статического ресурса; новый Node-тест и Go proof
покрывают состояние и встроенную поставку.

Синхронизированы task record/index, project context, current state,
CHANGELOG и зеркальный Quick Start. Исторические записи, включая
план TASK-138, сохраняются. MASTER_PLAN, ADR/DP, `spec/decisions.md`
и корневые README неприменимы: вехи, зависимости, контракты и
порядок запуска сохраняются.

## Recovery Evidence Envelope


### Pre-Commit Report — 2026-11-20

- Coordinator Acceptance: `Accepted`, canonical SHA-1
  `4be342f0613c189df8b0127d0f8b956fd8327145`. HEAD и база
  `4ce41f98b64d74efb2928fdcfbb9b7ffa8b38583`, ветка
  `feature/task-142-responsive-navigation`. Index пуст. STOP на Pre-Commit Gate.
- Независимый Reviewer: `APPROVED`, 0 findings. Воспроизведены 15 строк,
  151 Node-тест, тесты Go-пакета Studio, восемь относительных ссылок,
  EN/RU Quick Start parity и поведение JS в DOM harness. Scope Audit 15/0/0.
- Node 151/151, из них 148 baseline и три regression proof. Команда
  `node --test --test-reporter=dot internal/configurationstudio/control_navigation_test.js internal/configurationstudio/studio_test.js internal/configurationstudio/dashboard_test.js`.
- `go test -mod=readonly -p 6 -count=1 ./...` с loopback access,
  `go vet -mod=readonly ./...`, `gofmt -l` для двух изменённых Go-файлов
  и `git -c core.autocrlf=false -c core.whitespace=cr-at-eol diff --check`
  — exit 0. Новые Go concurrency paths отсутствуют.
- `go build -mod=readonly -buildvcs=false -o <temp-exe> ./cmd/control-service`
  — exit 0. Chrome 158.0.8127.36 через CDP с чистым профилем и отдельным
  Control Service на 18143, ширина 1280 и 390 px, EN/RU. На обеих
  страницах проверены основной маршрут, активная ссылка, один EN/RU,
  сохранение языка после refresh, отсутствие overflow. Мышь, touch,
  Tab, Enter, Space и Escape проверены нативными событиями.
- На 390 px шапка в одну строку, меню изначально скрыто и при открытии
  вертикальное. Tab пропускает закрытые ссылки, Escape возвращает фокус,
  переход и ссылка на раздел закрывают меню. Фокус проверен при смене
  ширины. Активный URL закрывает меню без загрузки. Уход с изменённым
  Draft вызвал прежний `beforeunload`. После отмены подтверждены поле
  порта, выбранный Workspace, Configuration, Draft и секция.
  Подтверждены 0 вызовов сети при открытии, смене языка и выборе
  текущей страницы, 0 браузерных записей API и 0 JS exceptions.
- Тестовый 18143 остановлен. Owner-инстансы на 18138–18142 сохранили PID
  `37772`, `21044`, `24148`, `30028` и `26932`. TASK-141 Not Activated.

#### Exact file manifest

Пути отсортированы по raw UTF-8. Каждая строка хешируется в общем
потоке как `path\0projection\0state\0mode\0oid\0` через
`git hash-object --stdin`. Для `full` использованы байты без фильтров,
task record использует `task-record-v1` PROCESS-001.

```text
.ai/PROJECT_CONTEXT.md | full | present | 100644 | 93c0808a262afd0d51175cb5a5851106f7e87d63
CHANGELOG.md | full | present | 100644 | 33d1cbc044a4e5ee936d00bfcabb9623e706eeda
docs/en/quick-start.md | full | present | 100644 | 538ddf6e5a13cf36e9b27fea145ffcf74e4a2575
docs/ru/quick-start.md | full | present | 100644 | 642f317d6c03a0633594d8a4876af6c1cdf9995e
docs/tasks/README.md | full | present | 100644 | d56ecbffd7de63a4c5415a85f5582e7722b5df2d
docs/tasks/TASK-142-RESPONSIVE-NAVIGATION.md | task-record-v1 | present | 100644 | a5f5e22e522c4dc14769ae6f19c46d0dbd2812a3
internal/configurationstudio/assets/control-language.js | full | present | 100644 | cc85a076e2f603d3e86dd24d5e4f7dc20efd33cf
internal/configurationstudio/assets/control-navigation.js | full | present | 100644 | f27adb187ef3feb7ac927bfa70e2d0c0e788da39
internal/configurationstudio/assets/control.css | full | present | 100644 | cc2ef9376d1718af192a76fa209a1189f9057f41
internal/configurationstudio/assets/dashboard.html | full | present | 100644 | e13bcdb71d8d2ab4991055402f3188a251d57307
internal/configurationstudio/assets/index.html | full | present | 100644 | 96e1a3149c48606856bba1997ec2f6deabfade69
internal/configurationstudio/control_navigation_test.js | full | present | 100644 | c4cadee298035d02a70374f398048fd44b97bbe8
internal/configurationstudio/handler.go | full | present | 100644 | bf5a39aacefd2d33f8f28322413094d8f1671691
internal/configurationstudio/handler_test.go | full | present | 100644 | c11d22b3959f0678c77875b2ee15342213b6a7ec
spec/current-state.md | full | present | 100644 | db7554579478c6ccb874a29668d21675d5269974
```

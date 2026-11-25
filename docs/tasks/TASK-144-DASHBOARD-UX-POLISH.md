# TASK-144 — Dashboard UX Polish

## Status

`In Progress`. Текущий gate определяется Recovery Evidence Envelope.

## Task Contract

- Task Mode `Implementation`. Owner одобрил коррекцию Dashboard после
  визуальной приёмки опубликованной TASK-141. Один срез оформления:
  ручное обновление и заголовки существующей навигации Dashboard.
- База `694ac794914da4d2b9c19f9787c353348ee965f7`, ветка
  `feature/task-144-dashboard-ux-polish`. Изолированное Git worktree
  сохраняет локальную документационную запись TASK-141 на её ветке.
- Scope — `internal/configurationstudio/assets/dashboard.html`,
  `control.css`, `control-language.js`, `dashboard.js` в том же каталоге,
  `internal/configurationstudio/dashboard_test.js`, обе версии Quick Start,
  `CHANGELOG.md`, `docs/tasks/README.md` и этот record. В `dashboard.js`
  разрешено только восстановление фокуса.
- Definition of Done — refresh-cw, поверхность 36 × 36 px, цель 44 × 44 px,
  подсказка и доступное имя EN/RU, видимый фокус, нативные клавиатура и touch.
  Единственный прежний GET-цикл, прежняя блокировка во время загрузки,
  сохранённые выбранные ресурсы, фокус возвращается только после потери
  из-за отключения кнопки и при отсутствии перехода к другой цели.
  Разделители и отступы только в Dashboard, включая открытое мобильное
  меню. Порядок, имена, ссылки и активный пункт сохраняются.
- Out of Scope — дизайн Studio, изменения API, темы и маршрутов,
  обработчики навигации, polling, деплой и действующие Control Services.
  Stop при необходимости расширения десяти файлов. Stage, commit и
  publication требуют отдельных команд Owner.
- Architecture Confirmation — существующие нативные элементы и
  embedded assets. Переиспользуются EN/RU словарь, CSS-токены темы,
  блокировка Refresh и номер текущего запроса. API и CSP сохраняются.
- Existing Coverage — Node Dashboard проверяет однократную загрузку,
  GET-only, ручной Refresh, сохранение выбранных ID, отмену, timeout и
  результаты ошибок. Studio и mobile navigation уже покрыты своими
  тестами, Go проверяет выдачу страниц. Coverage Gap — лейбл/подсказка,
  условный возврат фокуса, реальные касания, геометрия и CSS sidebar.
- Verification Plan — Node и полный Go, Chrome на отдельной сборке,
  Dashboard 1280/390/320 × EN/RU × Light/Dark, Studio regression,
  задержанный GET, Enter/Space/Tab/touch, PROCESS-002, независимый
  Reviewer, Scope Audit и точный manifest. STOP на Commit Gate.
  Следующая задача не активируется.

## Recovery Evidence Envelope

### Pre-Commit Report — 2026-11-24

- Coordinator Acceptance для Commit Gate — `Accepted`. Ветка
  `feature/task-144-dashboard-ux-polish`, HEAD и опубликованная база
  `694ac794914da4d2b9c19f9787c353348ee965f7`. Canonical SHA-1
  `c4ccef6bae5535d5659d1c59e47e09097e7e8295`. Scope Audit
  10 Required, 0 Questionable, 0 Removable. Индекс пуст. Новый task
  record остаётся ожидаемым untracked-файлом до отдельного разрешения.
- Dashboard использует нативную кнопку с иконкой refresh-cw, видимой
  поверхностью 36 × 36 px и целью 44 × 44 px. Имя и подсказка используют
  общий EN/RU перевод. Разделители и отступы ограничены навигацией
  Dashboard, в том числе раскрытым мобильным меню. Порядок, назначения
  и активное состояние сохранены. `dashboard.js` изменён только для
  восстановления фокуса при синхронной или отложенной потере из-за
  отключения кнопки. Действие относится к текущему поколению загрузки.
  Другое назначение фокуса, касание вне кнопки и потеря фокуса окна
  отменяют восстановление.
- Независимый Reviewer — `Approved` по указанному манифесту. Открытых
  замечаний нет. Независимо сверены 10 путей, пустой индекс, ссылки,
  зеркала EN/RU, whitespace, Node 164/164, Go-пакет Studio и
  финальное браузерное доказательство.
- `node --test --test-reporter=spec internal/configurationstudio/studio_test.js internal/configurationstudio/dashboard_test.js internal/configurationstudio/control_navigation_test.js internal/configurationstudio/control_theme_test.js`
  — 164/164, exit 0. Включены delayed blur, отсутствие перехвата фокуса,
  повторная активация и смена поколения. `go test -mod=readonly -p 6 -count=1 ./...`
  с loopback — exit 0, 40 пакетов с тестами.
- `go build -mod=readonly -buildvcs=true -o bin/task144-smoke/control-service-focus.exe ./cmd/control-service`
  прошла на Go 1.26.5. `vcs.revision` равен базе, `vcs.modified=true`
  соответствует проверяемому diff. SHA-256 бинарника
  `da33861da3f8af34c0402e74d149fd4a74e9513510199f0b019ccddab104f9bb`.
- `bin/task144-smoke/browser-run.ps1` на Chrome 158.0.8127.36 и
  127.0.0.1:18151 — exit 0. Dashboard прошёл 12 сочетаний 1280/390/320 px,
  EN/RU и Light/Dark. Проверены 36/44 px, видимость подсказки на hover
  и фокусе, доступное имя, Tab, Enter, Space и эмулированное касание.
  Подтверждены разделители, прежние ссылки и мобильные раскрытие и Escape.
  Горизонтального переполнения нет.
- Задержанный GET в Chrome блокировал повторную активацию. Один цикл
  включил только ожидаемые GET health, Workspace, Configuration и
  Versions. Выбор Workspace и Configuration с тестовым Draft сохранился.
  Перезагрузки и лишних запросов не было. Фокус возвращался после
  нативной потери при блокировке и сохранялся на новом элементе при
  перемещении. Дополнительно проверены уход с нового элемента и нажатие
  вне кнопки. Studio проверена при всех трёх ширинах, её оформление
  заголовков и поведение переключателей сохранены.
- Изолированный PID `45792` остановлен. Действующие PID в диапазоне
  18138–18150 совпали до и после. Chrome работал с эмулированным касанием.
  Другие браузеры и физические сенсорные устройства не использовались.
  Журналы, сборка и сценарии находятся в игнорируемом `bin/task144-smoke/`.
- PROCESS-002. Синхронизированы task record, индекс, Quick Start EN/RU и
  CHANGELOG. Current state, MASTER_PLAN, ADR/DP и контекст проверены на
  применимость. Возможности, API и границы milestone сохранены.
  Корневые README и решения не требуют правок. Документация закрытия
  TASK-141 сохранена в отдельном рабочем дереве. Whitespace и ссылки
  прошли. Stage, commit и publication не выполнялись. STOP на Commit Gate.

#### Exact changed-file manifest

Git object format SHA-1. Поля и конец каждой строки разделены NUL. Порядок
строк по исходным UTF-8 байтам пути. Весь поток передаётся
`git hash-object --stdin`.

```text
CHANGELOG.md | full | present | 100644 | d38acb11020e9b0317bf504802a278946fd5498f
docs/en/quick-start.md | full | present | 100644 | 7b6b9b4780342e1973be24e841b2213387384d34
docs/ru/quick-start.md | full | present | 100644 | 297bae099759b98298e75e7b97213bf87601e105
docs/tasks/README.md | full | present | 100644 | 9c9daa9cff8192ff001f2870c64d3abdd8c20a81
docs/tasks/TASK-144-DASHBOARD-UX-POLISH.md | task-record-v1 | present | 100644 | d45ba3deafc5a80548992565315caa218a5556e8
internal/configurationstudio/assets/control-language.js | full | present | 100644 | 0c80ccd17ddb5e8535592b4227e5dae4f573bfee
internal/configurationstudio/assets/control.css | full | present | 100644 | 2c4daa251a9ff8c75167ee2340ae5166a94e1e4d
internal/configurationstudio/assets/dashboard.html | full | present | 100644 | 7f120980c3b72dbeba2c3f772c551d80789c560f
internal/configurationstudio/assets/dashboard.js | full | present | 100644 | 04d0bdfcaccead325f7a205b8c5daf1205102c39
internal/configurationstudio/dashboard_test.js | full | present | 100644 | 1ad4e09ae4457a6ac239b3a02726ebff2de11e2b
```

### Owner Acceptance Closure — 2026-11-25

Решение Owner по TASK-144 — `ACCEPTED`.
Визуальная проверка пройдена. Замечаний Owner нет.

Проверенный опубликованный baseline — `main@f89efa82c6416ffa0c6d28866e4acfc4a9a148d5`.
Для совместной приёмки TASK-144, TASK-145 и TASK-146 предоставлена
изолированная сборка Control Service на `127.0.0.1:18157`.
Принят интерфейс Dashboard с кнопкой Refresh, локализованной подсказкой
и оформлением разделов навигации.

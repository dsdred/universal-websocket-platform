# TASK-141 — Темы UWP Control

## Status

`In Progress`. Текущий gate определяется Recovery Evidence Envelope.

## Task Contract

- Owner одобрил план и активировал пункт 2
  [UX-плана TASK-138](TASK-138-STUDIO-OWNER-ACCEPTANCE.md).
  TASK-141 зарезервирована для Light / Dark / System.
- Контракт UI: System по умолчанию, выбор хранится локально, System
  реагирует на изменения ОС. Studio и Dashboard синхронизируются между
  вкладками одного origin и после восстановления из BFCache. Смена темы
  не перезагружает страницу и не выполняет API-запросов.
- Проверяются ранний выбор темы, EN/RU, 1280/390/320 px, клавиатура, touch,
  контраст, сохранение фокуса, Draft и валидации. Диалог TASK-143 и
  native beforeunload сохраняют принятое поведение.
- Граница: утверждённые 20 файлов frontend, static asset delivery,
  тестов и документации. API, Runtime, TASK-131, TASK-134 и работающие
  Control Service инстансы вне scope. Commit и publication требуют
  отдельных команд Owner. STOP на Commit Gate.

## Recovery anchor

- Ветка `codex/task-141-control-themes`; базовый `main == origin/main`
  `dbe5ac48f67b1ee80628389ce01e054d42741c98`, PR #144.
- Источники: TASK-138, [TASK-142](TASK-142-RESPONSIVE-NAVIGATION.md),
  [TASK-143](TASK-143-STYLED-CONFIRMATION-DIALOGS.md), embedded Control UI.
  TASK-143 Owner Acceptance PASS 6/6 сохраняется.

## Architecture Confirmation

Независимый Architect: `APPROVED`, 0 blocking findings. Один ранний
same-origin скрипт выбирает CSS-палитру до загрузки стилей. Один
нативный select с переводом перемещается вместе с переключателем языка
по существующему breakpoint. Выбор темы не заменяет DOM редактора и
не трогает обработчики навигации Draft. CSP и Control API сохраняются.
ADR/DP не требуется для изменения presentation layer.

## Existing Coverage Report

155 Node-тестов baseline прошли. Studio, Dashboard и общая навигация
покрывают переходы, локализацию, Draft, dialog и beforeunload.
Go проверяет embedded delivery. Новые сценарии покрывают boot,
выбор и синхронизацию темы, фокус и наличие раннего asset на обеих страницах.

## Scope audit и PROCESS-002

Size Guard: 20 файлов укладываются в утверждённый Owner scope одного
presentation slice. Палитры, общий контроль, ранняя загрузка и
синхронизация должны работать на обеих страницах одного origin.
8 файлов реализации включают 7 frontend-ресурсов и выдачу static asset,
3 файла проверок покрывают новые риски, 9 документов фиксируют
task contract, статусы и пользовательские инструкции. ADR/DP,
`spec/decisions.md` и корневые README не меняются: архитектура,
публичные контракты и запуск сохраняются. EN/RU Quick Start и MASTER_PLAN
синхронизированы. Исторические записи TASK-138, TASK-142 и TASK-143 сохранены.

## Owner UX refinement — compact theme selector

- Task Mode: `Implementation`. Owner одобрил замену списка на три постоянно
  видимые иконки рядом с EN/RU. Причина — текстовый список занимает слишком много места.
- Baseline: `67a1a18c8b95d3dd0ef6533ee4725c5995bb3a1c`, ветка
  `feature/task-141-theme-selector-icons`. Scope: `assets/index.html`,
  `assets/dashboard.html`, `assets/control.css`, `assets/control-theme.js`,
  `control_theme_test.js`, `control_navigation_test.js` в `internal/configurationstudio`,
  зеркальные `docs/{en,ru}/quick-start.md` и этот task record.
- Definition of Done: нативная EN/RU радиогруппа Light/Dark/System, цели от
  44 × 44 px, отчётливые выбор и фокус, доступное имя UWP Control, ширины
  1280/390/320 px. Выбор хранится под прежним ключом, System следует ОС,
  вкладки и BFCache синхронизируются. Смена темы сохраняет Draft и валидацию
  без навигации или API-запросов. TASK-143 проверяется на регрессию.
- Out of Scope: изменения API, палитр, архитектуры, TASK-143 и действующих
  Control Services на 18138–18145. Stage, commit и publication требуют
  отдельных команд Owner. STOP на Commit Gate.
- Verification Plan: Node и Go, Chrome на изолированной сборке, независимое
  review, PROCESS-002 и Scope Audit с exact manifest. Роли: Coordinator,
  Developer/Tester, независимый Reviewer. Ранняя установка темы и responsive
  перенос сохраняются в существующей frontend-архитектуре.
- Existing Coverage: Node покрывает раннюю тему, выбор, ОС, вкладки, BFCache,
  недоступное хранилище, перенос предпочтений и Draft/диалоги. Go покрывает
  ранний скрипт и embedded delivery. Coverage Gap: состояние трёх радиокнопок,
  фокус на вложенной радиокнопке, реальная клавиатура, касание и размеры.

## Owner UX refinement — single theme toggle

- Task Mode `Implementation`. Owner одобрил одну кнопку выбора следующей
  палитры вместо группы из трёх иконок. База
  `5312192f919dfbd60a966d49945e5666f6fc0a26`, ветка
  `feature/task-141-single-theme-toggle`.
- Scope в `internal/configurationstudio` включает `assets/index.html`,
  `assets/dashboard.html`, `assets/control.css`, `assets/control-theme.js`,
  `assets/control-language.js`, `control_theme_test.js`,
  `control_navigation_test.js` и `handler_test.go`. Документация включает
  этот record, обе версии Quick Start и MASTER_PLAN, `spec/current-state.md`
  и `CHANGELOG.md`. API, навигация, TASK-143, архитектура и существующие
  Control Services на 18138–18147 вне scope.
- Definition of Done — нативная кнопка рядом с EN/RU с солнцем в тёмной
  палитре и луной в светлой, локализованным именем действия, фокусом и целью
  44 × 44 px на 1280/390/320 px. Сохранённый `system` и пустое или неверное
  значение следуют ОС без записи. Первый клик сохраняет явный Light/Dark под
  прежним ключом. Вкладки, BFCache и отказ хранилища покрыты проверками.
  Переключение сохраняет Draft и валидацию без сетевых запросов и навигации.
- Архитектура сохраняет раннюю установку палитры, локальное предпочтение,
  существующие маршруты embedded assets и CSP. Роли — Coordinator,
  Developer/Tester и независимый Reviewer.
- Existing Coverage включает раннюю установку, System, storage events,
  BFCache, отказ хранилища, перенос фокуса, выдачу встроенных страниц
  и защиту Draft. Coverage Gap — переключение на противоположную палитру,
  динамическое имя действия, миграция старого `system` одним нажатием и
  нативная активация кнопки на реальных страницах.
- Verification Plan — Node, полный Go, изолированный Chrome, независимое
  ревью, PROCESS-002 и точный манифест. STOP на Commit Gate. Stage, commit
  и publication требуют отдельных команд Owner.

## Owner UX correction — matching preference buttons

- Task Mode `Implementation`. Owner одобрил одинаковый видимый размер
  кнопок темы и EN/RU. Опубликованная база
  `7b0a1d3d193583dcbc606ca0ceecbc0692793e19`, рабочая ветка
  `feature/task-141-match-preference-buttons`.
- Scope — `internal/configurationstudio/assets/control.css` и этот record.
  Обе страницы используют общие стили и существующие нативные кнопки.
  Разметка, JavaScript, API и конфигурация действующих Control Services
  остаются вне scope. Архитектура загрузки и применения палитры сохраняется.
- Definition of Done — видимая поверхность 36 × 36 px, цель 44 × 44 px,
  визуальный зазор 8 px, единое оформление и центрирование. Зоны касания
  и фокус не перекрывают соседние элементы. Переключение сохраняет семантику
  и не вызывает перезагрузки или API-запросов.
- Existing Coverage — Node проверяет тему, локализацию и перенос фокуса,
  Go проверяет выдачу встроенных страниц и стилей. Coverage Gap — пиксельная
  геометрия, касания у границ и размещение обводки в реальном браузере.
- Verification Plan — Node и Go, Chrome на изолированной сборке с матрицей
  Studio/Dashboard × 320/390/1280 px × Light/Dark × EN/RU, независимый
  Reviewer, PROCESS-002 и точный манифест. Роли — Coordinator,
  Developer/Tester и независимый Reviewer. STOP на Commit Gate.
  Stage, commit и publication требуют отдельных команд Owner.

## Recovery Evidence Envelope

### Pre-Commit Report — 2026-11-21

- Coordinator Acceptance: `Accepted` для Commit Gate. Ветка
  `codex/task-141-control-themes`, HEAD и база
  `dbe5ac48f67b1ee80628389ce01e054d42741c98`, canonical SHA-1
  `90ff50010c3e58d17599f06ede3098437442695e`. Scope Audit 20 Required,
  0 Questionable, 0 Removable. Index пуст. STOP на Commit Gate.
- Независимый Reviewer: `APPROVED`, 0 blocking и 0 unresolved findings.
  Reviewer независимо воспроизвёл 20 строк, 160 Node-тестов, Go-пакет
  Studio, 11 ссылок и EN/RU parity. Независимая проверка контраста
  дала минимум 4,59:1 для проверенного текста, 3,09:1 для границ полей
  и 5,90:1 для фокуса. Chrome evidence проверена Reviewer.
- `node --test --test-reporter=spec internal/configurationstudio/studio_test.js internal/configurationstudio/dashboard_test.js internal/configurationstudio/control_navigation_test.js internal/configurationstudio/control_theme_test.js`
  — 160/160 (155 baseline + 5 новых). Полный
  `go test -mod=readonly -p 6 -count=1 ./...`, отдельный
  `go test -mod=readonly -count=1 ./internal/configurationstudio`,
  `go vet -mod=readonly ./...` и сборка Control Service прошли.
  `gofmt -l` для двух Go-файлов не вывел путей.
- `git -c core.autocrlf=false -c core.whitespace=cr-at-eol diff --check`
  прошла. В Studio index.html сохранён опубликованный CRLF; обычный
  diff check называет CR на новых строках trailing whitespace.
- Chrome 158.0.8127.36, изолированный Control Service 18145. В 24
  сочетаниях Studio/Dashboard × EN/RU × 1280/390/320 px × Light/Dark
  подтверждены отсутствие горизонтального переполнения и одна строка
  мобильной шапки с целями 44 px. Проверены клавиатурный выбор темы,
  видимый и восстановленный фокус, touch, 20 пар токенов контраста
  в каждой палитре. В других браузерах smoke не запускался.
- На первом кадре приложения подтверждены System Light по умолчанию,
  сохранённый Dark против OS Light на Dashboard и Light против OS
  Dark в Studio, включая cold reload обеих страниц. В System обе
  смены OS применились live. Реальные storage event из второй вкладки
  и восстановление из BFCache подтверждены Chrome. Ошибочные значения
  и недоступное хранилище покрыты Node.
- Смена темы сохранила Draft, секцию, несохранённое поле и ошибку ввода.
  Нулевые API/document-запросы при переключении, нулевые браузерные
  записи и JS exceptions. Проверены стили TASK-143 dialog в обеих
  палитрах, Cancel и Escape, возврат фокуса, один переход Confirm
  и нативный beforeunload при reload. CSP сохранён. API/Runtime,
  TASK-131, TASK-134 и TASK-143 Owner PASS не изменены.
- PID 18138–18144 до и после smoke совпали: `37772`, `21044`,
  `24148`, `30028`, `26932`, `26260`, `11184`.
  Smoke-процесс на 18145 остановлен. Отдельное Owner разрешение
  требуется для stage/commit и publication.

#### Exact file manifest

Git object format: SHA-1. Порядок путей по исходным UTF-8 байтам.
Сериализация каждой строки: `path\0projection\0state\0mode\0oid\0`.
Task record использует `task-record-v1`; остальные файлы хешируются
без фильтров.

```text
.ai/PROJECT_CONTEXT.md | full | present | 100644 | b1366e768cc0ded4a393a5f8fa124220dc4f9a39
CHANGELOG.md | full | present | 100644 | a3e48ba538ac58c95dc9f84eafceb1657946d59c
docs/en/quick-start.md | full | present | 100644 | 93ac6ec33d1b658277c56d4e38d19d4b1dc2bc12
docs/en/roadmap/MASTER_PLAN.md | full | present | 100644 | b145ecbe613ba58a8e7d63f78d7c0a69cdefdc6b
docs/ru/quick-start.md | full | present | 100644 | f3de2d3ac6e59665c7dd0715a77e5733b0fef02d
docs/ru/roadmap/MASTER_PLAN.md | full | present | 100644 | 9977affaf6c2f43ff874c2e781f0df38fa09ddfa
docs/tasks/README.md | full | present | 100644 | 3fca10d48f1be86ccff6651053384e58bf76e0d5
docs/tasks/TASK-141-CONTROL-THEMES.md | task-record-v1 | present | 100644 | 293cdab1d2b507deb60a9e0b134d8a5e63aca5b1
internal/configurationstudio/assets/control-language.js | full | present | 100644 | 7eab2cce979e56e228e2e56076febf6bb62badbd
internal/configurationstudio/assets/control-navigation.js | full | present | 100644 | 4c69fe179ea7c6f482d57c87b17982d97220a9c5
internal/configurationstudio/assets/control-theme.js | full | present | 100644 | bc26b6db8669396c2c841123d1a2daae35d85c5c
internal/configurationstudio/assets/control.css | full | present | 100644 | 3b012d72ced24ab81d32b47a2a574650f730148a
internal/configurationstudio/assets/dashboard.html | full | present | 100644 | 17d7e917bb142367b9fafd37ae6b707985313617
internal/configurationstudio/assets/index.html | full | present | 100644 | f3945ff155ea9cadbe81cd8bef5867b84dafacdd
internal/configurationstudio/assets/studio.css | full | present | 100644 | a27268c2badbf77d0f67be74a473f8538ce68441
internal/configurationstudio/control_navigation_test.js | full | present | 100644 | 1d857a636036e77614078970f569c4ec281f5bca
internal/configurationstudio/control_theme_test.js | full | present | 100644 | e891d93997435737a80436786e7d77c062b52d94
internal/configurationstudio/handler.go | full | present | 100644 | c8cdcea462948c539bfe8f83a498ee9cafc7f824
internal/configurationstudio/handler_test.go | full | present | 100644 | 60b65610f26b43f63d0f0372687cccb6633d156b
spec/current-state.md | full | present | 100644 | c6d5fee378fb239b3a77a8f0e5ce489adde67f57
```

### UX Refinement Pre-Commit Report — 2026-11-21

- Coordinator Acceptance для Commit Gate: `Accepted`. Ветка
  `feature/task-141-theme-selector-icons`. HEAD и опубликованная база
  `67a1a18c8b95d3dd0ef6533ee4725c5995bb3a1c`. Canonical manifest
  `50df418fd62ba4143a816166d41974a7062db425`. Scope Audit
  9 Required, 0 Questionable, 0 Removable. Index пуст. STOP на Commit Gate.
- Нативные радиокнопки Light/Dark/System рядом с EN/RU используют прежнюю
  локализацию, ключ `uwp-control-theme` и раннее применение палитры.
  Мобильная шапка сохраняет доступное имя ссылки UWP Control при скрытой
  подписи логотипа. Тесты радиогруппы проверяют один выбранный режим, а
  тест переноса шапки проверяет фокус на вложенной радиокнопке.
- Независимый Reviewer — `Approved`, 0 findings, exact manifest выше.
  Независимо подтверждены 9 файлов, пустой index, Node 160/160, Go-пакет
  Studio и EN/RU parity. По обоим наборам токенов минимальный контраст
  контура выбора с фоном 4,59:1, а контура фокуса с выбранным фоном 5,90:1.
- `node --test --test-reporter=spec internal/configurationstudio/studio_test.js internal/configurationstudio/dashboard_test.js internal/configurationstudio/control_navigation_test.js internal/configurationstudio/control_theme_test.js`
  — 160/160, exit 0.
  `go test -mod=readonly -p 6 -count=1 ./...` — exit 0, 40 пакетов с
  тестами. Первый прогон в песочнице не имел доступа к loopback. Полный
  прогон на хосте с loopback прошёл, результат сохранён в
  `bin/task141-icons-smoke/go-test.log`.
- `go build -mod=readonly -buildvcs=true -o bin/task141-icons-smoke/control-service-final.exe ./cmd/control-service`
  — exit 0. Go 1.26.5, `vcs.revision` совпадает с базой,
  `vcs.modified=true` соответствует рабочему diff. SHA-256 бинарника
  `a7c80757076904aa7390ed95fa37d09894fb230d9629c064c8813d86055e61de`.
- Браузерный прогон `bin/task141-icons-smoke/browser-run.ps1`, Chrome
  158.0.8127.36, Control Service на 127.0.0.1:18146, exit 0. В 36 сочетаниях
  Studio/Dashboard × EN/RU × 1280/390/320 px × Light/Dark/System проверены
  отсутствие переполнения, соседство EN/RU, все цели 44 × 44 px и единственный
  визуально выбранный режим. Проверены дерево доступности и имя ссылки
  UWP Control при 320 px, Shift+Tab, стрелка, Space, обводка и фокус после
  смены брейкпоинта. Касания проверены эмуляцией ввода Chrome.
- Chrome подтвердил сохранение режима после reload, live System при
  эмуляции настройки ОС, фактическую синхронизацию вкладок и восстановление
  Dashboard из BFCache. Смена всех трёх тем сохранила несохранённый Draft
  и ошибку валидации с нулём запросов API и документа. Проверены Cancel,
  Escape, возврат фокуса, native beforeunload и однократный переход Confirm.
  Финальный лог — `bin/task141-icons-smoke/browser-proof.log`. Другие браузеры
  и физическое сенсорное устройство не использовались.
- Порты 18138–18145 до и после smoke сохранили PID
  `37772, 21044, 24148, 30028, 26932, 26260, 11184, 30708` соответственно.
  Изолированный PID `41708` остановлен, порт 18146 свободен.
- PROCESS-002. Обновлены этот record и обе версии Quick Start.
  `spec/current-state.md`, зеркала MASTER_PLAN, `.ai/PROJECT_CONTEXT.md`,
  `spec/decisions.md`, ADR/DP, корневые README и CHANGELOG не требуют
  синхронизации в этом срезе. Возможности, контракты, пользовательский
  сценарий и границы milestone описаны существующими записями. Проверка
  whitespace с `core.whitespace=cr-at-eol` прошла. Сохранены исходные окончания
  строк файлов, включая CRLF в `index.html`. Сборка и логи находятся в
  игнорируемом `bin/` и в subject не входят. Stage, commit и publication
  не выполнялись. Следующая задача не активирована.

#### Exact UX refinement manifest

Git object format SHA-1. Для каждой строки используются поля в порядке ниже.
Между полями и в конце строки записывается NUL. Порядок строк по UTF-8 байтам
пути. OID полного потока вычисляется `git hash-object --stdin`.

```text
docs/en/quick-start.md | full | present | 100644 | d591b243a315432d6c77ae465ebe8f5043cca517
docs/ru/quick-start.md | full | present | 100644 | 1a962e06c1386952c39d116ca2df40c2b748b721
docs/tasks/TASK-141-CONTROL-THEMES.md | task-record-v1 | present | 100644 | 65e28fa36e13a9aacdaa32140ca079b52ce117ea
internal/configurationstudio/assets/control-theme.js | full | present | 100644 | a2bfba7e3d46bb184e6905c16f7b28fb34833149
internal/configurationstudio/assets/control.css | full | present | 100644 | 727e803c38e863098ccc1e858b53d2a939b236ca
internal/configurationstudio/assets/dashboard.html | full | present | 100644 | 16193b57386a8bc1ed8c8d7cfbf4965555afb92b
internal/configurationstudio/assets/index.html | full | present | 100644 | 09a177a7c5986e905ed5519fc02f65f736859a3d
internal/configurationstudio/control_navigation_test.js | full | present | 100644 | b0337d81e7dd0ad07d2332702489df41cd27a279
internal/configurationstudio/control_theme_test.js | full | present | 100644 | c84e51f9ccef29ec40bf86797dbc32d13b1743d1
```

### Single Toggle Pre-Commit Report — 2026-11-22

- Coordinator Acceptance для Commit Gate — `Accepted`. Ветка
  `feature/task-141-single-theme-toggle`, HEAD и опубликованная база
  `5312192f919dfbd60a966d49945e5666f6fc0a26`. Canonical SHA-1
  `8b622a5da5c4c81424daef9cc60816da3cc1c097`. Scope Audit
  15 Required, 0 Questionable, 0 Removable. Index пуст. STOP на Commit Gate.
- Одна нативная кнопка использует солнце в тёмной палитре и луну в светлой.
  Локализованное имя описывает следующее действие. Внутренний System
  сохраняет реакцию на ОС, пока под прежним ключом не записан явный выбор.
  Отказ записи в хранилище сохраняет выбор на странице и при её возврате
  из BFCache, что проверено Node.
- Independent Review — `Approved` по указанному манифесту. Проверены
  15 файлов, пробельные изменения, EN/RU parity, Node 160/160, Go-пакет Studio
  и финальный Chrome log. Уточнена документация об уже сохранённом Light/Dark.
  Открытых замечаний нет.
- `node --test --test-reporter=spec internal/configurationstudio/studio_test.js internal/configurationstudio/dashboard_test.js internal/configurationstudio/control_navigation_test.js internal/configurationstudio/control_theme_test.js`
  — 160/160, exit 0. `go test -mod=readonly -p 6 -count=1 ./...` с доступом
  к loopback — exit 0, 40 пакетов с тестами. `gofmt -l` для
  `handler_test.go` ничего не вывел. Журналы `node-test.log` и `go-test.log`
  сохранены в `bin/task141-single-toggle-smoke/`.
- `go build -mod=readonly -buildvcs=true -o bin/task141-single-toggle-smoke/control-service.exe ./cmd/control-service`
  — exit 0. Go 1.26.5, `vcs.revision` равен базе, `vcs.modified=true`
  отражает проверяемый рабочий diff. SHA-256 бинарника
  `93af8742820e2f07483324f424b93ce0d14a64f7cb62cea76cf8098abd9dd767`.
- `bin/task141-single-toggle-smoke/browser-run.ps1` прошёл в Chrome
  158.0.8127.36 на 127.0.0.1:18148, exit 0. В 24 сочетаниях
  Studio/Dashboard × EN/RU × 1280/390/320 px × Light/Dark подтверждены
  одиночная кнопка рядом с EN/RU, иконка следующего действия, доступное имя,
  цель 44 × 44 px и отсутствие переполнения. Проверены Tab, Enter, Space,
  эмулированное касание, индикация и сохранение фокуса при смене ширины.
- Chrome подтвердил отсутствие записи до клика для пустого предпочтения
  и сохранённого `system`, переключение живой темы ОС, явный выбор после
  клика, сохранение после reload, реальную синхронизацию вкладок и возврат
  Dashboard из BFCache. Темы переключались без API- и document-запросов.
  CSP совпал с опубликованным, Draft и ошибка валидации сохранились.
  TASK-143 Cancel, Escape, возврат фокуса, native beforeunload и однократный
  переход Confirm прошли. Другие браузеры и физические сенсорные устройства
  не использовались. Журнал — `bin/task141-single-toggle-smoke/browser-proof.log`.
- Изолированный PID `2016` остановлен. Порты 18138–18145 и 18147
  сохранили PID `37772, 21044, 24148, 30028, 26932, 26260, 11184, 30708, 15720`
  соответственно. 18146 был свободен до и после прогона.
- PROCESS-002. Обновлены этот record, обе версии Quick Start и MASTER_PLAN,
  current state и CHANGELOG. Индекс задач и `.ai/PROJECT_CONTEXT.md` уже
  ссылаются на тот же TASK-141 и текущий статус. ADR/DP, решения и корневые
  README проверены на применимость. Границы архитектуры и запуска сохраняются.
  `git -c core.autocrlf=false -c core.whitespace=cr-at-eol diff --check -- <15 paths>`
  прошла. Окончания строк унаследованы от базы. Сборка и доказательства в
  игнорируемом `bin/` за пределами манифеста. Stage, commit и publication
  не выполнялись. Следующая задача не активирована.

#### Exact single toggle manifest

Git object format SHA-1. Порядок строк по исходным UTF-8 байтам пути.
Поля разделяются NUL, строка также заканчивается NUL.
OID вычисляется `git hash-object --stdin` для полного потока.

```text
CHANGELOG.md | full | present | 100644 | 162b6cbe9636b4997be0aef79d97638853939c66
docs/en/quick-start.md | full | present | 100644 | 738388b50192ffad492e45d8b971106dad8cb017
docs/en/roadmap/MASTER_PLAN.md | full | present | 100644 | f23487f16bfd585e43011481762798df50d2d00d
docs/ru/quick-start.md | full | present | 100644 | 2342c44af23423c1e5a793570e44e76491d76fc2
docs/ru/roadmap/MASTER_PLAN.md | full | present | 100644 | 4e7de4227d44e7b59a0231a1397ffa7e72ea8f5e
docs/tasks/TASK-141-CONTROL-THEMES.md | task-record-v1 | present | 100644 | 21031184af6cdeebb1b6f0ca8740a18b841e54eb
internal/configurationstudio/assets/control-language.js | full | present | 100644 | c2f261cc5f40af91e2a9be0ebf18c9d3ec86b2f8
internal/configurationstudio/assets/control-theme.js | full | present | 100644 | d1e3503a056a314b4360e53e5d9e8e40f0bd2c5f
internal/configurationstudio/assets/control.css | full | present | 100644 | 4dfa532f51085906af24106686f51a4be0b8be17
internal/configurationstudio/assets/dashboard.html | full | present | 100644 | 8e3420404cf55c007b1fedca68fd2a7372307d97
internal/configurationstudio/assets/index.html | full | present | 100644 | b9590f64082e2cc81010bacc10935abe1e3c804a
internal/configurationstudio/control_navigation_test.js | full | present | 100644 | 4c29210a6f1e893a95c2e979a9d23fa31d6cc781
internal/configurationstudio/control_theme_test.js | full | present | 100644 | 9984feb28b31c622eb261f6b036488e291e70f75
internal/configurationstudio/handler_test.go | full | present | 100644 | 79a4ccda00f4e8a8505f95961523352130bf04b0
spec/current-state.md | full | present | 100644 | a505eec7f7fd02fc5934f171d8028ddbf4af4282
```

### Preference Button Geometry Pre-Commit Report — 2026-11-22

- Coordinator Acceptance для Commit Gate — `Accepted`. Ветка
  `feature/task-141-match-preference-buttons`, HEAD и опубликованная база
  `7b0a1d3d193583dcbc606ca0ceecbc0692793e19`. Canonical SHA-1
  `7dd122377986e63bad4555499e606e73423192ff`. Scope Audit
  2 Required, 0 Questionable, 0 Removable. Git-индекс пуст.
- Обе нативные кнопки занимают по 44 × 44 px. Оформление и видимый фокус
  отрисовываются на внутренних поверхностях 36 × 36 px с визуальным
  зазором 8 px. Содержимое центрируется общей сеткой. Навигация, разметка,
  обработчики и сохранение настроек используют прежнюю реализацию.
- Независимый Reviewer — `Approved` по указанному манифесту, без открытых
  замечаний. Независимо сверены оба файла, пустой Git-индекс, whitespace,
  Node 160/160, пакет Go `internal/configurationstudio` и доказательства Chrome.
- `node --test --test-reporter=spec internal/configurationstudio/studio_test.js internal/configurationstudio/dashboard_test.js internal/configurationstudio/control_navigation_test.js internal/configurationstudio/control_theme_test.js`
  — 160/160, exit 0. `go test -mod=readonly -p 6 -count=1 ./...` с loopback
  — exit 0, 40 пакетов с тестами.
- `go build -mod=readonly -buildvcs=true -o bin/task141-preference-geometry-smoke/control-service.exe ./cmd/control-service`
  прошла на Go 1.26.5. `vcs.revision` совпал с базой, `vcs.modified=true`
  соответствует рабочему diff. SHA-256 бинарника
  `67affc2ae6b17d985d3d994f0752a6bbb6414c39d391e1397f2307eaf57dcad6`.
- Chrome 158.0.8127.36 на 127.0.0.1:18150, exit 0. Проверены 24 сочетания
  Studio/Dashboard × 320/390/1280 px × Light/Dark × EN/RU. Совпали фон,
  рамка, скругление и отступы. Зоны 44 × 44 px соприкасаются без перекрытия
  друг с другом, меню и брендом. Переполнения по горизонтали нет.
  192 эмулированных касания по целочисленным координатам внутри всех четырёх
  краёв активировали нужную кнопку. Проверены Tab, Enter и Space.
  На 48 снимках пиксели обводки видны по всем четырём сторонам.
  Переключение темы и языка выполнилось без запросов API или документа и
  без перезагрузки. Физическое сенсорное устройство не использовалось.
- Изолированный PID `31776` остановлен, порт 18150 свободен. Инвентарь
  PID действующих портов 18138–18149 совпал до и после. Скрипты и логи
  сохранены в игнорируемом `bin/task141-preference-geometry-smoke/`.
- PROCESS-002. Task record описывает согласованные размеры и доказательства.
  Quick Start, текущая спецификация, MASTER_PLAN, changelog и контекст
  проверены на применимость. Действующие описания сценариев и контрактов
  точны для этой коррекции размеров. `git diff --check` по двум точным путям
  прошла. Stage, commit, публикация и следующая задача не запускались.

#### Exact preference geometry manifest

Git object format SHA-1. Порядок строк по UTF-8 байтам пути. Поля разделяются
NUL, строка также заканчивается NUL. OID вычисляется `git hash-object --stdin`
для полного потока.

```text
docs/tasks/TASK-141-CONTROL-THEMES.md | task-record-v1 | present | 100644 | 1e83fb34852d9ed921f487fe937a6395a5bae412
internal/configurationstudio/assets/control.css | full | present | 100644 | 5e6a8f9c56604c9465f19d0ce3b87411d78c577e
```

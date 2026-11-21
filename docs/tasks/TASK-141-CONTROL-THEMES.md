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

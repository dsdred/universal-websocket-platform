# TASK-145 — Общая навигация Studio и Dashboard

## Status

`In Progress`. Текущий gate определяется Recovery Evidence Envelope.

## Task Contract

- Task Mode `Implementation`. Owner визуально принял Dashboard TASK-144 и
  одобрил выравнивание оболочки. Опубликованный Dashboard задаёт оформление
  разделов навигации.
- Baseline `8bd9b2c6af37f59c92643ae42569fda39a4cce52`, локальные
  `main` и `origin/main` совпадают с live `main`. Ветка
  `feature/task-145-shared-navigation` в отдельном worktree сохраняет
  локальные правки TASK-141. TASK-145 свободен в индексе и refs.
- Scope: `control.css`, `control-navigation.js`, `dashboard.html`,
  `control_navigation_test.js`, `CHANGELOG.md`, индекс и этот record.
- Definition of Done: общие отступы и разделители, sidebar 220 px на
  desktop, прежние размеры ссылок, active/hover, адреса и `aria-current`.
  На desktop позиция controls учитывает фактическую ширину scrollbar в
  пределах header. На mobile и при overlay дополнительный отступ равен
  нулю. Сохраняются раскрытие, Escape, focus, EN/RU, Light/Dark,
  размеры form и Dashboard content при закрытом меню.
- Architecture Confirmation: сохранить embedded HTML/CSS/JS и
  существующую общую навигацию TASK-142. Две страницы используют
  `#control-navigation`. Выравнивание header использует разницу между
  шириной нативного scrollbar и уже занятой шириной viewport.
  Overlay решается фактической нулевой шириной. Глобальный scrollbar
  gutter отклонён, поскольку меняет ширину контента.
- Out of Scope: изменение DOM контента, API, theme/locale, destinations,
  механики disclosure, архитектуры, действующих Control Services и данных.
  STOP на Commit Gate. Stage, commit и публикация отдельно авторизуются.
- Existing Coverage: Node подтверждает disclosure, Escape/focus, ссылки,
  переходы breakpoint, Dashboard GET-only, Refresh, Studio workflows и
  тему. Go проверяет HTTP и embedded assets. Coverage Gap: adaptive header
  gutter, переход появления scrollbar, кросс-страничная геометрия и
  отсутствие влияния на контент. План: targeted Node regression, Go и
  Chrome до/после на отдельных процессах и пустом хранилище.
- Verification Plan: Chrome 1280/390/320 × EN/RU × Light/Dark, classic
  scrollbar с прокруткой и без неё, overlay, открытое и закрытое mobile
  меню, Escape, клавиатурный focus, навигация и API. PROCESS-002, Scope
  Audit и независимый Reviewer привязываются к exact manifest.
- Roles: Coordinator, Developer и Tester для implementation и evidence,
  отдельный Reviewer для независимого итогового решения. Следующая работа
  не активируется.

## Scope Audit и PROCESS-002

- Scope Audit: 7 Required, 0 Questionable, 0 Removable. Три embedded
  assets обеспечивают общий селектор, удаление прежнего marker и
  измерение scrollbar только для desktop header. Node regression
  проверяет появление и исчезновение gutter, mobile и overlay.
- Синхронизированы индекс задач и CHANGELOG. Quick Start EN/RU уже
  описывает общие навигацию, mobile menu и предпочтения; новых действий
  пользователя нет. `spec/current-state.md`, `spec/decisions.md`,
  ADR/DP, оба MASTER_PLAN, `.ai/PROJECT_CONTEXT.md` и корневые README
  проверены на применимость. Изменение не затрагивает capability,
  milestone, architectural decision или следующий этап roadmap.
- Тестовые бинарники, Chrome screenshots и журналы находятся в
  игнорируемом `bin/task145-proof/`. Исходная ветка TASK-141 и
  опубликованные refs сохраняются.

## Recovery Evidence Envelope

### Verification handoff

- `feature/task-145-shared-navigation`; HEAD и baseline
  `8bd9b2c6af37f59c92643ae42569fda39a4cce52`. Subject manifest
  SHA-1 `3215e298d2d390cfccbe9293cc844660d7683b0f`. Git index пуст.
- `node --test internal/configurationstudio/studio_test.js internal/configurationstudio/dashboard_test.js internal/configurationstudio/control_navigation_test.js internal/configurationstudio/control_theme_test.js`
  — 166/166, exit 0. Два новых regression-сценария покрывают 19 px,
  наличие и исчезновение scrollbar, mobile и нулевую ширину overlay.
- `go test -mod=readonly -p 6 -count=1 ./...` — exit 0, 40 packages
  with tests. `go build -mod=readonly -buildvcs=false -o <isolated-binary> ./cmd/control-service`
  — exit 0, SHA-256
  `6c6e0d8b442ce100d40c8e4673cff7284ad16e7b93c1c583b49af9fb0ef4cd05`.
  Studio, Dashboard и health вернули HTTP 200 на обеих изолированных
  сборках. Workspaces при старте `[]`.
- Chrome 158.0.8127.36, `bin/task145-proof/run-browser.ps1`:
  20 сравнений classic и 20 overlay, EN/RU × Light/Dark при
  1280×720/900/1100, 390×900 и 320×900. Overlay измерен как 0 px.
  Проверены естественная смена scrollbar от высоты документа и viewport,
  desktop/mobile focus, Escape, toggle и переход Dashboard ↔ Studio.
  Горизонтального overflow нет. Сравнены размеры и относительные
  координаты видимого контента и computed styles скрытых элементов.
  При открытом mobile menu форма Studio сдвигается вниз вместе с
  нижней границей меню. Внутренняя раскладка формы и Dashboard сохранена.
- Classic, Dashboard/Studio: desktop nav 267.16/276.94 px до,
  267.16/267.16 после. Header controls x при 1280×900
  1167.50/1152.50 до, 1152.50/1152.50 после. Тот же x после при
  1280×720 и 1280×1100. Desktop sidebar 220 px на обеих страницах.
  Mobile nav при 390 и 320 px 297.91/243.38 px до, 297.91/297.91
  после, строки 44 px. Mobile controls x сохранён: 231.11 и 161.11 px.
  До/после сохранены в `classic-geometry.json`, `overlay-geometry.json`
  и PNG внутри каталога proof. Проверка mobile выполнена в Chrome
  с эмулированными viewport.
- `git -c core.autocrlf=false -c core.whitespace=cr-at-eol diff --check -- <changed tracked paths>`
  — exit 0, ссылки проверены. Bytes tracked docs сохраняют baseline
  line endings, product/test files используют LF. Будущий stage должен
  сохранить approved raw bytes, и staged tree сверяется отдельно.

#### Exact ordered changed-file manifest

Git SHA-1. Поля и строки разделены NUL. Для task record применяется
`task-record-v1`; остальные строки содержат raw Git blob OID.

```text
CHANGELOG.md | full | present | 100644 | b24be38d44a4f1c76ff1cce3af06556dee3673cd
docs/tasks/README.md | full | present | 100644 | 42d9fc836d013601aa802b3a87707c2e1fe50e0f
docs/tasks/TASK-145-SHARED-NAVIGATION.md | task-record-v1 | present | 100644 | 534b9d7326d092e2af90f0cf48d090cb0f90a424
internal/configurationstudio/assets/control-navigation.js | full | present | 100644 | db29102ef13fdceb2099d063a622fdc5ebe425ea
internal/configurationstudio/assets/control.css | full | present | 100644 | 1abe6702648bae758035073f4749423c1202f59d
internal/configurationstudio/assets/dashboard.html | full | present | 100644 | c3aff11549591e3d7d68e82c07c505cbb513ac54
internal/configurationstudio/control_navigation_test.js | full | present | 100644 | 2626a16d3f5e5580b209c32d43d4a9fa89e93ec5
```

### Pre-Commit Report — 2026-11-25

- Supplemental Chrome `run-browser.ps1 -RefreshOnly` — exit 0.
  12 сценариев 1280/390/320 × EN/RU × Light/Dark сохранили GET-цикл
  при задержанном запросе, блокировку повторной активации и выбор
  Workspace/Configuration. Проверены Tab, Enter, Space,
  эмулированное касание, восстановление фокуса после native blur и
  сохранение нового фокуса после перехода. Только тестовый процесс
  получал fixtures. Оба proof-процесса остановлены; PID и порты
  действующих Control Services совпали до и после.
- Independent Reviewer — `Approved`, 0 открытых findings по
  `3215e298d2d390cfccbe9293cc844660d7683b0f`. Независимо
  пересчитаны семь rows и projection, проверены index и whitespace,
  SHA-256 сборки, Node 166/166, Go 40 packages, две матрицы Chrome,
  скриншоты на трёх ширинах и Refresh 12/12. Scope Audit подтверждён:
  7 Required, 0 Questionable, 0 Removable.
- Remaining Limitations: автоматизированный Chrome на Windows с
  измеренным classic gutter и переключённым в браузере overlay mode.
  Физические мобильные устройства и другие браузеры не использовались.
- Coordinator Acceptance для Commit Gate — `Accepted` по тому же
  манифесту. Следующая задача не активирована. Stage, commit, push и
  publication не выполнены. STOP на Commit Gate.

# TASK-146 — Выравнивание кнопок мобильной шапки

## Status

`In Progress`. Текущий gate определяется Recovery Evidence Envelope.

## Task Contract

- Task Mode `Implementation`. Owner подтвердил, что мобильная кнопка меню
  выглядит крупнее кнопок темы и языка, и одобрил общий стиль для трёх
  кнопок. Chrome подтвердил расхождение на Dashboard и Studio.
- Baseline `bac94b86d0a9a37877b1d355cc3378c5930e67a9` совпадает с
  локальными `main`, `origin/main` и live `main`. Ветка
  `feature/task-146-mobile-header-controls` в отдельном worktree.
  Локальные изменения TASK-141 остаются в исходном worktree.
- Scope: единственный production-файл `control.css`, этот record,
  индекс задач и CHANGELOG. Size Guard не срабатывает.
- Definition of Done: три кнопки на mobile имеют видимую поверхность
  36 × 36 px, область ввода 44 × 44 px, видимые зазоры 8 px,
  border 1 px, radius 7 px и одинаковую вертикальную позицию.
  Прозрачные края и углы доступны; области ввода не перекрываются.
- Architecture Confirmation: переиспользовать существующие CSS-правила
  поверхности, hover и focus-visible. Сохранить нативные кнопки,
  `aria-expanded`, `aria-controls`, доступные имена, breakpoint 700 px,
  общий скрипт навигации TASK-142 и выравнивание desktop TASK-145.
- Out of Scope: HTML, JavaScript, API, desktop/content layout,
  рабочие Control Services и их данные. STOP на Commit Gate.
  Stage, commit и публикация требуют отдельных команд Owner.
- Existing Coverage: 166 Node-тестов покрывают disclosure, Escape,
  focus, темы, локализацию, Dashboard Refresh и Studio workflows.
  Go покрывает HTTP и встроенные assets. Baseline Chrome на четырёх
  ширинах подтвердил различие размеров и текущие сценарии ввода.
  Coverage Gap: фактическое попадание указателя в прозрачные углы,
  совпадение геометрии и CSS focus/hover. Proof — реальный Chrome,
  поскольку Node harness не выполняет CSS layout или hit testing.
- Verification Plan: 320/390/768/1280 × EN/RU × Light/Dark ×
  classic/overlay, Dashboard и Studio, compare baseline/candidate.
  Снимки, CSS/DOM размеры, hit-test краёв и углов, мышь, эмуляция
  touch, Tab/Enter/Space/Escape, theme/locale, links и active state.
  Node, Go, PROCESS-002, Scope Audit и независимый Reviewer.
- Roles: Coordinator, Developer и Tester выполняют реализацию и proof.
  Независимый Reviewer проверяет exact manifest. Следующая работа
  не активируется.

## Scope Audit и PROCESS-002

- 4 Required, 0 Questionable, 0 Removable. В `control.css` квадратный
  mobile hit box сохраняет доступность прозрачных углов. Скругление
  осталось на общей видимой поверхности; hover и фокус используют
  существующие CSS-токены. Убраны отдельный размер и зазор hamburger.
  Общий конечный отступ мобильной шапки `1.25rem` оставляет место
  для нативного overlay scrollbar в Chrome. Начальный отступ и
  высота шапки сохраняются. Размеры viewport и scrollbar в CSS не
  вычисляются, отдельный breakpoint для этого отступа не добавлен.
- Индекс и CHANGELOG синхронизированы. Quick Start EN/RU не требует
  изменений, поскольку назначение и использование кнопок прежние.
  `spec/current-state.md`, `spec/decisions.md`, ADR/DP, оба MASTER_PLAN,
  `.ai/PROJECT_CONTEXT.md` и корневые README проверены на применимость.
  Capability, публичный контракт и milestones сохраняются.
- Browser regression проверяется воспроизводимым proof-сценарием
  `bin/task146-proof/browser.cjs`. Код проверки, скриншоты, бинарники
  и журналы находятся в игнорируемом proof-каталоге. Existing Node
  и Go tests используются без изменений.

## Recovery Evidence Envelope

### Verification handoff

- Branch `feature/task-146-mobile-header-controls`; HEAD и baseline
  `bac94b86d0a9a37877b1d355cc3378c5930e67a9`. Subject manifest
  SHA-1 `643553e907c99d46fa6d8f0c1df621460c833918`; index пуст.
- `node --test internal/configurationstudio/studio_test.js internal/configurationstudio/dashboard_test.js internal/configurationstudio/control_navigation_test.js internal/configurationstudio/control_theme_test.js`
  — 166/166, exit 0. `go test -mod=readonly -p 6 -count=1 ./...`
  — exit 0, 40 packages with tests. `go build -mod=readonly -buildvcs=false -o <isolated-binary> ./cmd/control-service`
  — exit 0. Candidate SHA-256
  `501e1de8251ba20ea5f00a771af8c7db45eadc06e6a59cdfd97cb86f124fc9e3`.
- `bin/task146-proof/run-browser.ps1` — exit 0. Published baseline
  и candidate стартовали на 18155 и 18156 с пустыми stores. Studio,
  Dashboard и health вернули HTTP 200, Workspaces `[]`. После
  проверки эти процессы остановлены. PIDs и порты существующих
  Control Services совпали, локальный diff TASK-141 сохранён.
- Chrome 158.0.8127.36: 48 пар classic и 48 пар overlay,
  Dashboard и Studio, EN/RU × Light/Dark, 320/390/768/1280 px.
  На 1280 дополнительно проверены высоты 720 и 1100 px.
  Сохранены desktop header, sidebar, обе области контента,
  позиции ссылок при раскрытии, active state, ARIA и доступные имена.
  Горизонтального overflow нет. Overlay gutter измерен как 0 px.
- Mobile до: поверхности Theme/Language 36 × 36 px, Menu 44 × 44 px,
  верхние границы 11/11/7 px, зазоры 8/8.89 px. После: 36 × 36 px,
  границы 11/11/11 px, зазоры 8/8 px. Области ввода 44 × 44 px
  остались смежными без перекрытия, border 1 px, radius 7 px.
  Высота mobile header 58 px на 320 и 390. Видимый focus outline 2 px.
- По 384 hit-test точки в каждом scrollbar mode: четыре угла и четыре
  края каждой кнопки. Click мышью и эмулированный touch в прозрачных
  углах переключили тему, язык и меню. Проверены Tab, Enter, Space,
  Escape, возврат фокуса, переходы между страницами, закрытие меню и
  сохранение RU/Light. Дополнительно проверен переход breakpoint 700/701.
- `git -c core.autocrlf=false -c core.whitespace=cr-at-eol diff --check -- <changed tracked paths>`
  — exit 0. Exact patch проверен через `git apply --check --reverse`.
  Снимки и точные исходные координаты в `classic-geometry.json`,
  `overlay-geometry.json` и PNG из `bin/task146-proof/`.
  Chrome headless на Windows, мобильные viewport и touch эмулированы.

#### Exact ordered changed-file manifest

Git SHA-1. Поля и записи разделены NUL. Для task record применяется
`task-record-v1`; остальные строки содержат raw Git blob OID.

```text
CHANGELOG.md | full | present | 100644 | 54fa9ad8ac70b338cbc6a632625f0a49b8dd4fc4
docs/tasks/README.md | full | present | 100644 | edf0f37f157b6a382929c25114f834963e3c15c9
docs/tasks/TASK-146-MOBILE-HEADER-CONTROLS.md | task-record-v1 | present | 100644 | 386b4074769093867e636925a6913c84c02af973
internal/configurationstudio/assets/control.css | full | present | 100644 | 927c3c3460d0a28594e64371cd3b49177e0b81ed
```

### Pre-Commit Report — 2026-11-25

- Independent Reviewer — `Approved`, 0 открытых и blocking findings,
  exact subject `643553e907c99d46fa6d8f0c1df621460c833918`.
  Независимо проверены 4 raw/projection rows, binary SHA-256, index
  и whitespace; повторены 166 Node tests и сверены 40 Go packages.
  Проверены 96 пар browser geometry и 768 hit-test points, скриншоты,
  navigation supplement и совпадение listener inventories.
  Scope Audit 4 Required, 0 Questionable, 0 Removable.
- Chrome proof использует измеренные classic и browser-enabled overlay
  scrollbars, эмулированный input. Другие браузеры и физические touch
  устройства не использовались.
- Coordinator Acceptance для Commit Gate — `Accepted` по тому же
  манифесту. Следующая задача не активирована. Stage, commit, push и
  publication не выполнялись. STOP на Commit Gate.

### Owner Acceptance Closure — 2026-11-25

Решение Owner по TASK-146 — `ACCEPTED`.
Визуальная проверка пройдена. Замечаний Owner нет.

Проверенный опубликованный baseline — `main@f89efa82c6416ffa0c6d28866e4acfc4a9a148d5`.
Для совместной приёмки TASK-144, TASK-145 и TASK-146 предоставлена
изолированная сборка Control Service на `127.0.0.1:18157`.
Принято выравнивание кнопок темы, языка и меню в мобильной шапке
Dashboard и Studio.

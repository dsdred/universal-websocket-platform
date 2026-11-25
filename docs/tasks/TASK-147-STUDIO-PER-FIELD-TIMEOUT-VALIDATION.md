# TASK-147 — Поэлементная валидация таймаутов Studio

## Status

`In Progress` для документационного закрытия. Реализация опубликована,
Owner Acceptance подтверждена. Точное состояние документационного gate
определяется по Recovery Evidence Envelope.

## Task Contract

- **Task Mode** — Implementation.
- **Why Now** — пункт 5 [плана UX TASK-138](TASK-138-STUDIO-OWNER-ACCEPTANCE.md#prioritized-studio-uiux-plan).
  TASK-137 уже сохраняет четыре таймаута, а общая ошибка показывает только
  первое нарушение. Owner выбрал этот ограниченный срез перед изменениями
  сохранения, публикации и переходов между страницами.
- **Definition of Done** — Save показывает все ошибки рядом с полями,
  фокусирует первое неверное поле и до исправления не отправляет запросы
  записи. Поля связаны с сообщениями через ARIA. Правка очищает только
  соответствующую ошибку, отмена Discard сохраняет ошибки, успешное
  восстановление значений очищает их. EN/RU применяется на месте. Валидный
  Save отправляет один существующий timeout PUT на выбранный Draft.
- **Out of Scope** — API, Runtime, Dashboard, существующие экземпляры
  Control Service, TASK-131, автоматические повторы, stage/commit/publish.
- **Verification Plan** — Node regression, Go embedded markup и repo checks,
  изолированный mobile browser smoke, PROCESS-002, scope audit, независимый
  Review, Coordinator decision, STOP на Commit Gate.

## Baseline and roles

- HEAD и опубликованный baseline
  `a51fd7ce1eb030f92dfda71802fd266a91d81960` (PR #153),
  локальные main/origin/main и live origin совпали.
- Ветка `feature/task-147-timeout-validation` в worktree
  `bin/task147-worktree`. Исходная ветка с шестью локальными правками
  сохранена. Это первая правка в task worktree.
- Coordinator назначает текущему агенту intake, реализацию, документацию и
  verification. Независимый Reviewer проверяет точный subject.
- Architecture confirmation — UI повторяет паттерн Listener и контракт
  [TASK-137](TASK-137-EXTENDED-DRAFT-EDITING.md). ADR-0002 и API не требуют
  изменения. Publisher вне scope.
- Size Guard — один independently deliverable UI-срез, до 12 утверждённых
  путей, без новых packages и архитектурных контрактов.
- Stop — расхождение baseline/scope, обязательный failed check,
  unresolved blocking review или достижение Commit Gate.
- Кандидат после TASK-147 — отдельная команда публикации сохранённого Draft
  в Studio, `Not Activated`. Требует нового выбора Owner.

## Existing Coverage Report

- Existing Coverage — baseline Node 166/166. Studio уже проверяет неверные
  значения, границы и zero, scoped PUT, смену секции, отмену Discard,
  неизвестный результат и локализацию. Go проверяет embedded markup и
  существующий timeout API. Dashboard имеет отдельный regression suite.
- Coverage Gap — одновременные ошибки, фокус первого поля, адресная
  очистка, ассоциации ARIA, перевод видимых ошибок и сохранение ошибок
  после Cancel. Mobile smoke проверит раскладку и клавиатурный сценарий.
- Added Proof Tests — Node: одновременные ошибки/фокус/исправление и
  Cancel/Reload; Go: привязка ошибок к четырём полям в embedded HTML.
- Added Regression Tests — текущие Node проверки неверного ввода и EN/RU
  проверяют размещение ошибки возле поля. Существующие boundary, scoped PUT,
  dirty/unknown outcome, Listener и Dashboard сценарии сохранены.
- Remaining Limitations — headless Chrome проверяет эмуляцию мобильного
  viewport, ввод, вставку, перемещение фокуса и numeric input hint.
  Физическая экранная клавиатура и реальный screen reader не проверяются.

## Implementation and scope audit

Поля сохраняют введённый текст и запрашивают цифровую клавиатуру через
`inputmode="numeric"`. JavaScript проверяет непустую строку десятичных
цифр и прежний диапазон. Поля используют существующие CSS-правила Listener и
`ControlLanguage`. Save использует прежний scoped request и проверку ответа.

| Path | Scope |
|---|---|
| `internal/configurationstudio/assets/index.html` | Required — связи полей с сообщениями |
| `internal/configurationstudio/assets/studio.js` | Required — валидация, фокус, очистка ошибок |
| `internal/configurationstudio/assets/control-language.js` | Required — локализация общего сообщения |
| `internal/configurationstudio/studio_test.js` | Required — поведение и regression |
| `internal/configurationstudio/handler_test.go` | Required — проверка embedded accessible markup |
| `docs/tasks/TASK-147-STUDIO-PER-FIELD-TIMEOUT-VALIDATION.md` | Required — contract и evidence |
| `docs/tasks/README.md` | Required — реестр задачи |
| `.ai/PROJECT_CONTEXT.md` | Required — текущий срез |
| `spec/current-state.md` | Required — фактическая возможность Studio |
| `docs/en/quick-start.md` | Required — инструкция Save timeouts |
| `docs/ru/quick-start.md` | Required — зеркальная инструкция |
| `CHANGELOG.md` | Required — пользовательское изменение |

PROCESS-002 — обновлены task record, index, context, current-state, EN/RU
Quick Start и Unreleased changelog. MASTER_PLAN, README и ADR/DP не требуют
правок: scope milestone, runtime capability и публичные контракты не меняются.
Проверки зависимостей и race неприменимы к синхронной клиентской валидации,
которая не изменяет зависимости, cancellation или lifecycle. Go build, test и
vet проверяют встроенные assets и текущие пакеты.

## Recovery

Рабочий anchor — task branch и указанный baseline. Порядок этапов:
intake → implementation → verification → documentation/scope → independent
review → Coordinator decision. При прерывании сверить реальные HEAD, diff,
test evidence и content identity до продолжения. Изменения разрешены только
в согласованных 12 путях. Точный subject фиксируется после проверки.

## Owner Acceptance Closure

26 ноября 2026 года Owner подтвердил ручную приёмку Studio на опубликованной
версии `9ff3ea2f32e31cfa3030a2728ffd1699a03a615e`. Owner сообщил, что ошибки
отображаются корректно, исправление поля очищает его ошибку и сохранение
таймаутов работает. Результат — **Owner Acceptance PASS**. Источник —
сообщение Owner `01a21448-af49-7381-875f-ed8dbd1e3f46` в turn
`01a21448-aea6-7cf0-bf79-98de14a83ca3` чата
`01a213e4-f648-78f1-916b-f951563e8778`. Результат фиксирует именно
подтверждённые Owner наблюдения.

- [PR #154](https://github.com/dsdred/universal-websocket-platform/pull/154)
  слит в `main`. Исходный коммит
  `affee889c46d20b35fa9f5c03194b994c90aa01a`, дерево исходного и merge-коммитов
  `f40a51f2835a616055c7d9407b2ffd52b99071b0`. CI прошёл на
  [PR](https://github.com/dsdred/universal-websocket-platform/actions/runs/45685781589)
  и на [main](https://github.com/dsdred/universal-websocket-platform/actions/runs/45686136047).
  P10 публикации реализации завершён.
- Изолированный стенд `127.0.0.1:18158`, PID `23396`, был собран из указанного
  merge-коммита с `vcs.modified=false`. Бинарник SHA-256
  `4ae92afde85530aafe0bd9a84a66dba136a1b9b27c42d573db33f21665e35000`.
  Стенд содержал `TASK-147 Owner Acceptance → Timeout Validation → Draft v1 (#1)`.
  Подготовительная API-проверка Save/readback и возврат начальных значений
  завершились перед ручной проверкой.
- После подтверждения Owner проверены PID, путь и hash бинарника,
  HTTP health и владение loopback listener. Остановлен PID `23396`, порт
  `18158` свободен. Инвентарь всех остальных 14 слушателей диапазона
  `18138–18169` совпал до и после остановки.

### Документационный Task Contract

- **Mode / Why Now** — Documentation-only. Зафиксировать подтверждённую
  приёмку опубликованной реализации и закрытие её изолированного стенда.
- **Branch / baseline** — `docs/task-147-owner-acceptance` в отдельном
  worktree `bin/task147-owner-closure`, HEAD и базовый `main` на merge SHA
  выше. Имеющиеся локальные правки сохранены в исходном worktree.
- **Definition of Done** — отражены сообщение Owner, exact PR/merge, итог
  остановки стенда и следующий шаг. PROCESS-002, exact scope и
  независимый Review пройдены до документационного Commit Gate.
- **Scope** — эта запись, `docs/tasks/README.md`, `.ai/PROJECT_CONTEXT.md`
  и `spec/current-state.md`. Product/API changes, другие Control Services,
  stage/commit/publish и активация TASK-148 вне scope.
- **Verification Plan** — сверить Git/GitHub, источники Owner evidence,
  host/PID и listener inventory, локальные ссылки, целостность исторического
  envelope и достижимость exact manifest; затем независимый Review и решение
  Coordinator. Product regression неприменим к документационному diff.
- **PROCESS-002** — task record, index, current-state и project context
  синхронизируются по факту приёмки. MASTER_PLAN EN/RU, ADR/DP, decisions,
  Quick Start, корневые README и CHANGELOG проверяются на применимость.
  Приёмка уже выпущенного поведения сохраняет milestone, API, руководство
  пользователя и состав релиза, поэтому эти файлы за пределами diff.
- **Roles / STOP** — Coordinator, Documentation Agent, независимый Reviewer.
  Size Guard — четыре документационных пути и отдельный Commit Gate.
  Следующее действие после закрытия TASK-147 — планирование
  **TASK-148 — Publish Existing Draft**, `Not Activated`.

## Recovery Evidence Envelope

### V147-001 / R147-002 / CA147-001 — Commit Gate STOP

Baseline и HEAD `a51fd7ce1eb030f92dfda71802fd266a91d81960`,
ветка `feature/task-147-timeout-validation`. Формат Git SHA-1.
Итоговый subject manifest `12a613f04510cba406747f7a8f4586a87479b153`.
Точные строки subject (path | projection | state | mode | blob OID):

```text
.ai/PROJECT_CONTEXT.md | full | present | 100644 | 7f586dfa6b5ca8f474710421cddc19c38aa85463
CHANGELOG.md | full | present | 100644 | d6b16939b6b83e907085689f9fd3dd9271c1a27c
docs/en/quick-start.md | full | present | 100644 | 72533c53406a174105483ea9fcecd664da94f9e7
docs/ru/quick-start.md | full | present | 100644 | 2d7de58c082f8279030b679625fb966150492a3a
docs/tasks/README.md | full | present | 100644 | 95e0dddc7a1e9f71049bf39ff0841f925c6c4e50
docs/tasks/TASK-147-STUDIO-PER-FIELD-TIMEOUT-VALIDATION.md | task-record-v1 | present | 100644 | 81f048fd7a7862e7727a9ca923bbbb1330b46a9f
internal/configurationstudio/assets/control-language.js | full | present | 100644 | 4eb51eef1cf10c923c94e31da86aa37a41c5b6e5
internal/configurationstudio/assets/index.html | full | present | 100644 | 94abc478fe15fc106cb4771ddce833d9a16845e7
internal/configurationstudio/assets/studio.js | full | present | 100644 | 13a8f54054f9885c95b950487f0c02f46f04a5d3
internal/configurationstudio/handler_test.go | full | present | 100644 | dbb049f15bc32da31929ec7b4d024fcd27e54aca
internal/configurationstudio/studio_test.js | full | present | 100644 | 95b9a20be692ac0e8f08fc92140ebfab4ab832b6
spec/current-state.md | full | present | 100644 | 526514920e030a96cf1c52d146ab6bff068c5cfa
```

- Verification PASS. `node --test internal/configurationstudio/*_test.js` —
  168/168. Проверены все четыре поля, фокус, изолированная очистка, EN/RU,
  Cancel/Reload, zero и границы, отсутствие записей при ошибках, один
  scoped PUT при успехе, прежние Listener/Dashboard и unknown-outcome tests.
- Go PASS. `go test -mod=readonly -count=1 ./...`,
  `go build -mod=readonly ./...` и `go vet -mod=readonly ./...`.
  `gofmt -l` для изменённого Go test-файла не вывел путей.
- Mobile smoke PASS. Chrome 158.0.8127.36, 320/390 px, реальные Tab/Enter/Escape,
  ввод и вставка, программно заданные значения, numeric input hint,
  видимые сообщения EN/RU без горизонтального переполнения, ноль записей
  при invalid Save и один PUT при valid Save. Временный loopback fixture
  использовал текущие assets, живые экземпляры Control Service не использованы.
- PROCESS-002 и Scope Audit PASS, 12 Required / 0 Questionable / 0 Removable.
  `git -c core.whitespace=cr-at-eol diff --check` PASS. Один minor finding
  первого Review (повтор предложения в RU Quick Start) исправлен, frontend
  tests повторены. Runtime/API, Dashboard и пользовательские правки в
  исходной ветке сохранены.
- R147-002 Independent Review **Approved**, 0 blocking / 0 nonblocking.
  Reviewer независимо пересчитал 12 строк и сравнил NUL manifest побайтно.
  Результат привязан к итоговому subject выше.
- CA147-001 Coordinator Acceptance — **Accepted** для указанного subject.
  Следующий шаг — отдельная команда Owner на Commit Gate. Стандартный
  Windows autocrlf меняет prospective filtered blobs для семи исторических
  CRLF/mixed путей. Read-only hash check с command-scoped
  `core.autocrlf=false` сохраняет exact raw bytes всех путей. При отдельном
  разрешении stage требует exact staged-tree сверки по принятому subject.
  Git config не менялся. Индекс пуст, commit/push/publication не выполнялись.

### V147-OA-001 / R147-OA-001 / CA147-OA-001 — документационный Commit Gate

Ветка `docs/task-147-owner-acceptance`, HEAD/base
`9ff3ea2f32e31cfa3030a2728ffd1699a03a615e`. SHA-1 subject manifest
`a8a4a0706c500e6648f577c2034dd3782e2a3b3c`.

```text
.ai/PROJECT_CONTEXT.md | full | present | 100644 | dd31a0d736d2bb3323b62a27a3b4f6d43d0c7bcb
docs/tasks/README.md | full | present | 100644 | 1ae3cff1dee15180c83713b542a3e5539261f5d5
docs/tasks/TASK-147-STUDIO-PER-FIELD-TIMEOUT-VALIDATION.md | task-record-v1 | present | 100644 | 4da521b352e0d04b455fb2288c4852caac1202ed
spec/current-state.md | full | present | 100644 | 5e109b72d8791504aa2489022127bddfd77b2f82
```

- V147-OA-001 PASS. Git ancestry, source/merge tree и опубликованный PR #154
  сверены. Подтверждены пять новых относительных ссылок и сохранение
  исходного Recovery Evidence Envelope побайтно. `git -c
  core.whitespace=cr-at-eol diff --check` PASS. Инвентарь diff содержит
  четыре документационных пути. Product regression неприменим к этому diff.
- PROCESS-002 PASS. Scope Audit — 4 Required / 0 Questionable / 0 Removable.
  Task record фиксирует решение Owner и стенд, index обеспечивает навигацию,
  current-state фиксирует принятую возможность, project context фиксирует
  очередь следующего планирования.
- R147-OA-001 **Approved**, 0 blocking / 0 nonblocking. Независимый Reviewer
  пересчитал все четыре строки и canonical NUL manifest, подтвердил
  источник и границы Owner Acceptance, сохранность исторического envelope
  и отключение только подтверждённого экземпляра на 18158.
- CA147-OA-001 Coordinator Acceptance — **Accepted** для документационного
  subject выше. Предлагаемое сообщение коммита
  `docs: record TASK-147 owner acceptance`. Индекс пуст. Для будущего stage
  использовать command-scoped `core.autocrlf=false` и сверить staged tree с
  принятыми bytes, включая итоговый envelope. Commit и публикация требуют
  отдельных разрешений. TASK-148 остаётся `Not Activated`.

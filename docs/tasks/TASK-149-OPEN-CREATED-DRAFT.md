# TASK-149 — Открытие созданного Draft из Studio

## Status

Реализация опубликована. Owner Acceptance — **Accepted** для опубликованной
версии. Документационное закрытие — `In Progress`. Его актуальный gate
определяется по последней записи в Recovery Evidence Envelope, совпадающей с
пересчитанным manifest документационного subject.

## Task Contract

- **Task Mode** — Implementation. **Why Now** — опубликованная TASK-148 прошла
  Owner Acceptance. Owner выбрал кандидат №1, Open Created Draft. Другие
  кандидаты и TASK-131 не активированы.
- **Definition of Done** — успешный Create Draft показывает действие открытия.
  Оно выполняет свежий scoped GET, проверяет точный ID и состояние Draft и
  открывает существующий редактор. При недоступности версии или ошибке чтения
  результат создания сохраняется. Неизвестный исход создания исключает действие.
  Проверяются повторные действия, EN/RU, клавиатура, фокус и регрессии Create/Edit.
- **Out of Scope** — новые API, изменения Control Service и Runtime, TASK-131,
  редизайн Studio, архитектура навигации и соседние задачи.
- **Verification Plan** — Node Studio/Control UI suites, smoke клавиатуры и
  фокуса, Go test/build/vet и оценка применимости race, PROCESS-002, scope audit
  и independent review итогового diff.

## Baseline and Roles

Baseline и HEAD `81cd47c5b80c3cca8a8b3300fee65fafe0f507cb`, branch
`codex/feature/task-149-open-created-draft`, worktree `bin/task149-worktree`.
Исходный checkout `docs/task-141-owner-acceptance-closure` на
`694ac794914da4d2b9c19f9787c353348ee965f7` содержит шесть исключённых путей.
Изоляция сохраняет оригинальные файлы и позволяет проверять diff от точного
baseline. Порядок стадий: intake → design confirmation → implementation →
verification → PROCESS-002 → scope audit → independent review → acceptance →
Commit Gate. Отдельные разрешения нужны для stage/commit и публикации.

Coordinator фиксирует границы; Developer реализует slice; Tester проверяет
сценарии; Documentation Agent синхронизирует документацию. Независимый
Reviewer проверяет итоговый subject. Publisher в этой задаче не активируется.

## Architecture Confirmation

[ADR-0002](../en/adr/0002-configuration-dsl.md) сохраняет lifecycle версии.
[ARCH-005](../en/architecture/ARCH-005-runtime-configuration-snapshot-and-loading-model.md)
закрепляет границу опубликованной конфигурации и Runtime. Studio использует
существующие контракты [TASK-135](TASK-135-STUDIO-EXISTING-CONFIGURATION.md),
[TASK-136](TASK-136-EXISTING-DRAFT-MANAGEMENT.md) и
[TASK-148](TASK-148-PUBLISH-EXISTING-DRAFT.md). Workspace подтверждён scoped
списком Configurations, ответ POST подтверждает Configuration и ID Draft.
Свежий GET списка версий этого scope проверяет точный ID, исходный `createdAt`
и состояние Draft. Сравнение точной строки `createdAt` отклоняет наблюдаемое
повторное использование ID после перезапуска in-memory Control Service.
Переход использует существующий редактор и его проекцию. Изменения контрактов
или предварительной архитектурной документации не требуются.

## Existing Coverage Report

- Existing Coverage: `studio_test.js` проверяет создание Draft без parent
  mutations или publish, неизвестный исход и неверный scope, Draft-only выбор,
  fresh GET, гонки открытия, Save с закреплённым ID, локализацию и управление
  фокусом диалогов. `handler_test.go` проверяет embedded Studio.
- Coverage Gap: переход из подтверждённого результата, выбор точного ID при
  нескольких Drafts, fresh editor values, stale/missing/network failure,
  сохранение результата и отсутствие повторных запросов при pending.

Size Guard: одна Studio capability, существующие пакеты и API, без нового
архитектурного контракта. STOP у Commit Gate либо при blocking finding.

## Implementation and Scope Audit

| Классификация | Пути | Назначение |
| --- | --- | --- |
| Required production | `internal/configurationstudio/assets/index.html`, `studio.js`, `control-language.js` | Кнопка результата, закрепление подтверждённого scope, GET и существующая проекция редактора, доступный фокус и EN/RU. |
| Required verification | `internal/configurationstudio/studio_test.js`, `handler_test.go` | Проверки transition, отсутствия подмены и побочных записей, pending, ошибок и embedded accessible button. |
| Required documentation | `docs/tasks/TASK-149-OPEN-CREATED-DRAFT.md`, `docs/tasks/README.md`, `.ai/PROJECT_CONTEXT.md`, `spec/current-state.md`, `docs/en/quick-start.md`, `docs/ru/quick-start.md`, `CHANGELOG.md` | Persistent anchor, навигация, состояние среза, зеркальное руководство и пользовательское изменение. |

PROCESS-002: перечисленные документы синхронизированы. MASTER_PLAN EN/RU
неприменимы: границы milestone и зависимости сохраняются. ADR, ARCH, DP и
`spec/decisions.md` неприменимы: architecture и API сохраняются. Корневой
README неприменим: пользовательский сценарий описан в Quick Start.

Added Proof Tests: точный Draft среди нескольких, свежие listener values,
один GET, защита от повторного действия и Save в точный Draft. Added Regression
Tests: Published, Archived, deleted, повторное использование ID, ошибка GET
и безопасный retry, отсутствие действия при unknown либо malformed creation.

Исключения из subject в исходном checkout: `.ai/PROJECT_CONTEXT.md`,
`docs/en/roadmap/MASTER_PLAN.md`, `docs/ru/roadmap/MASTER_PLAN.md`,
`docs/tasks/README.md`, `docs/tasks/TASK-141-CONTROL-THEMES.md`,
`spec/current-state.md`. Это локальные версии в отдельном checkout, без
изменений в их байтах в ходе TASK-149. Ни один следующий кандидат не активирован.

## Owner Acceptance Closure

26 ноября 2026 года Owner принял опубликованную реализацию TASK-149.
Источник ручных наблюдений — сообщение `01a21784-6348-7be0-a898-ab56a9025ba2`,
turn `01a21784-6229-7543-98a4-93c8d2128a6f`, чат
`01a21714-1a24-7db0-8b81-23a965744e3b`.

[PR #158](https://github.com/dsdred/universal-websocket-platform/pull/158)
слит в `main`. Проверенная Owner версия `main` — merge
`dcaa017eddc002de83c89af7c4f5e259ba871eab`. Коммит реализации
`2539a4e32eb39e5b3bcd885fce7879a1e456787a` и merge имеют общее дерево
`23d645c5fef9766b846ce1eb5048480a8152e9d4`. В ходе публикации `go-ci` прошёл
на [PR](https://github.com/dsdred/universal-websocket-platform/actions/runs/45799769069)
и на [main](https://github.com/dsdred/universal-websocket-platform/actions/runs/45800119832).

| Ручная проверка | Результат Owner |
| --- | --- |
| Create Draft v2 (#2) | PASS |
| Open created Draft | PASS |
| Открытие именно Draft v2 (#2) | PASS |
| Изменение Listener port на `19149` | PASS |
| Save и Reload Draft | PASS |

Owner вручную не проверял выбор среди нескольких Drafts, устаревший или
отсутствующий Draft, ошибку GET и повтор чтения, а также screen reader.
Эти сценарии не входят в ручной PASS. Для первых трёх сценариев отдельные
автоматизированные доказательства сохранены в evidence реализации ниже.
Headless-проверка клавиатуры и фокуса не включала screen reader.

### Контракт документационного закрытия

- **Task Mode** — Documentation-only. **Why Now** — Owner подтвердил
  опубликованную версию. Branch `codex/docs/task-149-owner-acceptance`,
  worktree `bin/task149-owner-closure`, baseline/HEAD перед изменениями
  `dcaa017eddc002de83c89af7c4f5e259ba871eab`.
- **Definition of Done** — записать решение Owner и ограничения, сверить
  опубликованные Git identities, синхронизировать состояние и навигацию,
  подтвердить exact subject независимым Review.
- **Scope** — task record, `docs/tasks/README.md`, `.ai/PROJECT_CONTEXT.md`
  и `spec/current-state.md`. Эти четыре пути Required. Product code, tests,
  API, Runtime и следующая задача вне scope. Шесть посторонних локальных
  правок остаются в исходном checkout.
- **Verification Plan** — сверить запись с сообщением Owner, локальные Git
  identities, исторический envelope, ссылки, четыре пути и whitespace.
  Product tests остаются evidence опубликованной реализации.
- **PROCESS-002** — task record, task index, project context и current state
  применимы. MASTER_PLAN EN/RU, ADR, ARCH, DP, `spec/decisions.md`, Quick Start,
  корневой README и CHANGELOG неприменимы: приёмка фиксирует состояние уже
  опубликованной функции. Milestone, архитектура, руководство и релиз не меняются.
- **Size Guard и роли** — четыре документационных пути, один срез. Coordinator
  координирует приёмку, Documentation Agent синхронизирует записи, независимый
  Reviewer проверяет итоговый subject. Следующий кандидат не активирован.

## Recovery Evidence Envelope

### Verification and exact subject

- Repository `universal-websocket-platform`, branch
  `codex/feature/task-149-open-created-draft`, base/HEAD
  `81cd47c5b80c3cca8a8b3300fee65fafe0f507cb`, Git object format SHA-1.
- `node --test internal/configurationstudio/*_test.js` — PASS, 186/186.
  Шесть добавленных сценариев и усиленные проверки Create/Publish.
- `go test -mod=readonly -count=1 ./...`, `go build -mod=readonly ./...`,
  `go vet -mod=readonly ./...` — PASS, exit 0. Для существующих тестов
  с loopback потребовался контекст, разрешающий локальные соединения.
- `CGO_ENABLED=1 go test -mod=readonly -race -count=1 ./internal/configurationstudio`
  — Windows toolchain не нашёл `gcc`, exit 1 на этапе сборки. Verification
  PASS WITH LIMITATION. Pending/duplicate async guards проверены в Node.
- `node bin/task149-browser-proof.cjs` из исходного checkout — PASS, exit 0.
  Реальный headless Edge и локальный mock существующего API. EN и RU: Enter
  создаёт Draft, фокус переходит на действие открытия, клавиатурная активация
  открывает точный Draft со свежим портом через один GET. RU: Space активирует
  действие, 503 переводит фокус на сообщение и сохраняет результат, Tab
  возвращает на действие, Enter повторяет только GET. Screen reader
  и ручная Owner Acceptance не проводились.
- `git -c core.whitespace=cr-at-eol diff --check` и отдельная проверка нового
  task record — PASS; `gofmt -l internal/configurationstudio/handler_test.go`
  без вывода. Scope audit: 12 Required / 0 Questionable / 0 Removable.
  SHA-256 всех шести исключённых файлов совпали до и после.
- Точность подтверждается на момент свежего GET. Если версия изменится после
  чтения, дальнейшие операции остаются под существующими правилами редактора
  и проверки lifecycle на сервере.

Manifest `cf6ad2ff88296784a5d53e7d2b64624d7a3f3dbe` вычислен через `git hash-object --stdin`
из отсортированных NUL-separated rows PROCESS-001. Ниже те же rows с
разделителями для чтения.

```text
.ai/PROJECT_CONTEXT.md | full | present | 100644 | a25cf7a67ab7a337429284ef19ab1e0b55ebbfdb
CHANGELOG.md | full | present | 100644 | 71239d25a380b0150be2201c3cc014a0a865b063
docs/en/quick-start.md | full | present | 100644 | 79c3a2b667bc331c25c9614d8a889f0df9f1baae
docs/ru/quick-start.md | full | present | 100644 | 6cc8f3cbdbdec8141377e150f02dfd04ca637c0b
docs/tasks/README.md | full | present | 100644 | 02c60e61232bd615ac28677d3657dfd73cddc903
docs/tasks/TASK-149-OPEN-CREATED-DRAFT.md | task-record-v1 | present | 100644 | 508770730f97b8838ec6859b4f21657c178057a2
internal/configurationstudio/assets/control-language.js | full | present | 100644 | e9072bb1f5f39478466f20d5a5e8a09ab520fe81
internal/configurationstudio/assets/index.html | full | present | 100644 | 99c9156282027abde25b842b1047179ca54fe8b6
internal/configurationstudio/assets/studio.js | full | present | 100644 | 0fc55ff2bb15d331210840b968816469c4998b5f
internal/configurationstudio/handler_test.go | full | present | 100644 | b5b872db38609dbd30f0f982f8c826e9353f5a9b
internal/configurationstudio/studio_test.js | full | present | 100644 | d0e6bb5f15e73dbe33ca21e8962e36c1e2ead8d2
spec/current-state.md | full | present | 100644 | 9632d91d3dbaf12dc28693c067661d5b07ae8bac
```

### Independent Review and Coordinator Acceptance

- Final Independent Review — `Approved` (0 blocking, 0 nonblocking) для
  manifest `cf6ad2ff88296784a5d53e7d2b64624d7a3f3dbe`. Reviewer независимо
  пересчитал все 12 rows, повторил Node 186/186 и сверил keyboard proof,
  EN/RU Quick Start и восемь локальных ссылок. Все пути Required.
  Предварительные замечания о повторном использовании ID и проверке
  продолжения Save закрыты в reviewed subject.
- Coordinator Acceptance — `Accepted`, 2026-11-26, тот же manifest и baseline.
  PROCESS-002 `Synchronized`. Следующий gate — Commit Gate, состояние STOP до
  отдельной команды Owner. Stage, commit, push, merge и следующая задача не
  выполнялись.

### Documentation-only Owner Acceptance evidence

- Branch `codex/docs/task-149-owner-acceptance`, baseline/HEAD
  `dcaa017eddc002de83c89af7c4f5e259ba871eab`. Git object format SHA-1.
- Verification PASS: exact five ручных PASS и четыре исключения сверены с
  сообщением Owner. `git rev-parse` подтвердил точный merge, второго
  родителя и совпадение деревьев реализации и merge. Исторический envelope
  байт-в-байт сохранён, SHA-256
  `4745683c6334b09ca4b767a6c7acfdb36f34b774465931f2f099d98daabe919c`.
- `git -c core.whitespace=cr-at-eol diff --check` PASS. Пять локальных
  ссылок в diff разрешаются. Scope Audit 4 Required / 0 Questionable /
  0 Removable. Product code, tests и публичные EN/RU документы не изменялись.
  SHA-256 шести исключённых локальных файлов совпали с исходными.

Manifest документационного subject `fc2d244b734b346c2770ffbe26bb256e010b6cf1` вычислен
через `git hash-object --stdin` из отсортированных NUL-separated rows
PROCESS-001.

```text
.ai/PROJECT_CONTEXT.md | full | present | 100644 | 0e03b6cdd9fca5c7b0ce45ca5119df797827b4a0
docs/tasks/README.md | full | present | 100644 | 69c65bf749d5182abf86f7cd7c62588a0d081545
docs/tasks/TASK-149-OPEN-CREATED-DRAFT.md | task-record-v1 | present | 100644 | 2110ac32dbd0926154c83416a3a8e31324160b1b
spec/current-state.md | full | present | 100644 | 59d79313a07e97a2fa9381e805317638834e3ca9
```

- Final Independent Review — `Approved`, 0 blocking / 0 nonblocking,
  exact documentation manifest `fc2d244b734b346c2770ffbe26bb256e010b6cf1`. Reviewer независимо
  пересчитал четыре rows, сверил native Owner evidence, Git-публикацию,
  CI, исторические байты, scope и closure links.
- Coordinator Acceptance документационного subject — `Accepted`,
  2026-11-26, тот же manifest и baseline. PROCESS-002 `Synchronized`.
  Следующий gate — Commit Gate. На момент этой записи stage, commit
  и публикация документационного slice не выполнялись. Следующая задача
  не активирована.

# TASK-150 — Защита Studio при недоступности открытого Draft

## Status

Реализация опубликована. Owner Acceptance — **Accepted** для опубликованной
версии. Документационное закрытие — `In Progress`. Актуальный gate определяется
по последней записи в Recovery Evidence Envelope, совпадающей с пересчитанным
manifest документационного subject.

## Task Contract

- **Task Mode** — Implementation.
- **Why Now** — Owner утвердил Studio-only исправление. Открытый Draft может
  перейти в Published или Archived в другой вкладке. После подтверждённого
  отказа Save сейчас снова доступен.
- **Baseline / HEAD** — `2854f32244155ed84d1ac37ed9a832103f69efe6`.
  Branch `codex/task-150-stale-draft-protection`, worktree `bin/task150-worktree`.
- **Definition of Done** — авторитетный scoped ответ о недоступности точной
  открытой версии блокирует оба Save и Publish, включая Enter и повторные
  действия. Введённые значения и выбор сохраняются, другая версия
  автоматически не открывается. EN/RU, фокус, явный переход и read-only
  recovery проверяются. Ошибки сети и неоднозначные ответы сохраняют свои
  отдельные состояния.
- **Out of Scope** — новые API, изменения lifecycle Control Service и Runtime,
  polling, live sync, ETag/CAS, TASK-131 и посторонние изменения Studio.
- **Verification Plan** — focused и полные Node suites, Go test/build/vet и
  оценка race, smoke двух вкладок на отдельном loopback Control Service,
  PROCESS-002, scope audit, независимый review точного manifest.
- **Roles** — Coordinator ведёт intake и gate. Developer реализует утверждённый
  срез, Tester проверяет его, Documentation Agent сверяет документы.
  Независимый Reviewer проверяет финальный subject. STOP у Commit Gate.

## Architecture Confirmation

[ADR-0002](../en/adr/0002-configuration-dsl.md) и
[ARCH-005](../en/architecture/ARCH-005-runtime-configuration-snapshot-and-loading-model.md)
сохраняют существующие lifecycle и Runtime boundaries. Studio использует
scoped Versions GET и typed 404/409 существующего API. Правила Draft-only
мутаций принадлежат Control Service. [TASK-136](TASK-136-EXISTING-DRAFT-MANAGEMENT.md),
[TASK-137](TASK-137-EXTENDED-DRAFT-EDITING.md),
[TASK-148](TASK-148-PUBLISH-EXISTING-DRAFT.md) и
[TASK-149](TASK-149-OPEN-CREATED-DRAFT.md) описывают существующие Save, Publish,
recovery и открытие точной версии. Новый архитектурный контракт не требуется.

## Existing Coverage Report

- **Existing Coverage** — 155 Studio Node tests на baseline. Покрыты scoped
  open/save, отмена и поздние GET, сохранение ввода при 404/409, unknown Save,
  stale Publish preflight, Publish race, read-only recovery, EN/RU, Escape,
  фокус и dirty navigation. Go покрывает lifecycle rejection и atomic batch.
- **Coverage Gap** — запрет обоих Save и Publish после typed отказа или
  подтверждённой недоступности, сохранение этого запрета при failed follow-up
  GET и при чистой форме, отдельные ambiguous outcomes, отсутствие подмены.
- **Size Guard** — одна пользовательская возможность в существующей Studio.

Шесть правок исходного checkout исключены из subject. Исходный checkout
`docs/task-141-owner-acceptance-closure` на `694ac794914da4d2b9c19f9787c353348ee965f7`;
binary diff fingerprint `d247a12d68547c07e5cd15ea0dfed8ff2f42bc40`.

## Implementation and Scope Audit

| Классификация | Пути | Назначение |
| --- | --- | --- |
| Required production | `internal/configurationstudio/assets/studio.js`, `internal/configurationstudio/assets/control-language.js` | Точное сопоставление scoped evidence, блокировка обеих мутаций, сохранение формы, доступный фокус и EN/RU. |
| Required verification | `internal/configurationstudio/studio_test.js` | Proof typed lifecycle, scoped absence, read-only recovery, гонки выбора, форма и фокус. |
| Required documentation | `docs/tasks/TASK-150-STUDIO-STALE-DRAFT-PROTECTION.md`, `docs/tasks/README.md`, `.ai/PROJECT_CONTEXT.md`, `spec/current-state.md`, `CHANGELOG.md` | Persistent anchor, навигация, состояние ветки и пользовательское исправление. |
| Required documentation | `docs/en/quick-start.md`, `docs/ru/quick-start.md` | Зеркальное описание блокировки, сохранения ввода и явного выбора другого Draft. |

PROCESS-002: зеркальные MASTER_PLAN неприменимы, границы milestone и
зависимости сохраняются. ADR, ARCH, DP и `spec/decisions.md` неприменимы,
архитектура и API сохраняются. Корневой README неприменим, пользовательский
сценарий описан в Quick Start.

**Added Proof Tests** — обе секции Save после typed 409, scoped 404,
Published/Archived/пропавшая версия, наблюдаемая подмена ID и отказ follow-up
GET после typed Publish conflict. Проверяется сохранение формы и блокировка
после её очистки, read-only reload и явный выбор другой версии.
**Added Regression Tests** — неверный код конфликта, bare/malformed response,
неудачная Publish preflight, поздний scoped 404, EN/RU, фокус и подтверждение
discard.
**Remaining Limitations** — Studio обнаруживает смену состояния при получении
scoped evidence. Для изменений ещё активного Draft действует существующая
конкурентная семантика без CAS.

## Owner Acceptance Closure

27 ноября 2026 года Owner принял опубликованную реализацию TASK-150.
Ручные наблюдения зафиксированы в сообщении `01a21957-563e-7be0-acd8-82104e11fb35`,
turn `01a21957-5378-7453-b4f0-c5544296ef50`, чат
`01a218ba-778d-7c80-9b5a-7d467077f34f`.

[PR #160](https://github.com/dsdred/universal-websocket-platform/pull/160)
слит в `main`. Owner проверял опубликованный merge
`67b74d7760188a073d139a339fef38a7975069a5`. Его второй родитель — коммит
реализации `c1a5f98882730d5a2aac12b8138cc3a6b5d3d4aa`. Оба имеют дерево
`946be9bedb4f6f06b1677645664e6376faf1903a`.

### Ручная проверка Owner

| Сценарий | Наблюдение | Результат |
| --- | --- | --- |
| Draft v2 (#2), Save listener | Draft опубликовали в другой вкладке. Save listener отклонил сохранение. Несохранённый порт `19150` и выбранный Draft сохранились. Save и Publish стали недоступны. | PASS |
| Draft v3 (#3), Reload Draft | Draft опубликовали в другой вкладке. Reload Draft обнаружил недоступность версии. Save и Publish стали недоступны, выбранный Draft сохранился. | PASS |

За пределами ручной проверки Owner остались следующие сценарии.

- ошибка сети и восстановление после неё;
- сохранение несохранённого ввода при Reload уже недоступного Draft;
- Save в секции Timeouts после перехода версии из Draft;
- полная проверка клавиатурой и проверка со screen reader.

### Автоматизированная проверка реализации

Опубликованный [набор Studio-тестов](../../internal/configurationstudio/studio_test.js)
покрывает сетевые ошибки, повторное чтение и сверку после неопределённого
исхода, сохранение несохранённого ввода при stale Reload и stale Save для
обеих секций.
Отдельная двухвкладочная проверка Edge покрывает клавиатурные действия и
фокус в проверенных сценариях, включая stale Save в секции Timeouts на RU.
Screen reader в эту проверку не входил. Результаты реализации зафиксированы
в [Recovery Evidence Envelope](#recovery-evidence-envelope).

### Контракт документационного закрытия

- **Task Mode** — Documentation-only. **Why Now** — Owner принял версию
  из опубликованного `main`. Ветка `codex/docs/task-150-owner-acceptance`,
  worktree `bin/task150-owner-closure`, исходный HEAD
  `67b74d7760188a073d139a339fef38a7975069a5`.
- **Definition of Done** — зафиксировать решение Owner, оба ручных сценария
  и пределы этой проверки, отдельно указать автоматизированное покрытие,
  синхронизировать статус и навигацию, передать точный diff на независимый Review.
- **Scope** — Required пути включают этот task record, `docs/tasks/README.md`,
  `.ai/PROJECT_CONTEXT.md`, `spec/current-state.md`. Код, тесты, API, Runtime,
  TASK-131 и следующие задачи вне scope. Шесть локальных правок остаются
  в исходном checkout.
- **Verification Plan** — сверить ручные наблюдения с сообщением Owner,
  удостоверить Git-идентичность публикации, исторические bytes, локальные
  ссылки, канонический manifest и чистоту diff. Повторить Node suite для
  опубликованного кода как отдельное автоматизированное доказательство.
- **PROCESS-002** — этот record, индекс задач, project context и current state
  применимы. Границы milestone, зависимости и roadmap status сохраняются,
  поэтому зеркальные MASTER_PLAN неприменимы. Приёмка опирается на существующий
  архитектурный контракт, поэтому ADR, ARCH, DP и `spec/decisions.md` неприменимы.
  Quick Start, корневой README и CHANGELOG неприменимы к записи о приёмке
  опубликованной функции.
- **Size Guard и роли** — четыре документационных файла, один срез.
  Coordinator ведёт gate, Documentation Agent синхронизирует записи,
  Reviewer независимо проверяет subject. STOP у Commit Gate.

## Recovery Evidence Envelope

Append-only evidence для exact subject записывается здесь.

### TASK-150 — Verification and Independent Review

Anchor HEAD `2854f32244155ed84d1ac37ed9a832103f69efe6`, branch
`codex/task-150-stale-draft-protection`. Object format `sha1`; exact canonical
manifest `dcde7290e75ffbbf7fa544738e724bcce87a8011`.

| Path | Projection | State | Mode | Blob OID |
| --- | --- | --- | --- | --- |
| .ai/PROJECT_CONTEXT.md | full | present | 100644 | bd7a8b2ad50a458c4bb2ed3a12a1101fcc8d0b73 |
| CHANGELOG.md | full | present | 100644 | a2c783d91344613ea8ddf7b2d0264324e20869cb |
| docs/en/quick-start.md | full | present | 100644 | 76c3eb7cd029a82708f4a13e32fc97b1567712a9 |
| docs/ru/quick-start.md | full | present | 100644 | 09af9afe7380e3dc2e50d5138ce433b5ba38fdf2 |
| docs/tasks/README.md | full | present | 100644 | 49e343d305914a376d7ffbb650587b4074bf7614 |
| docs/tasks/TASK-150-STUDIO-STALE-DRAFT-PROTECTION.md | task-record-v1 | present | 100644 | 2aa3f762c2832a1df5baf23716fbbe1dd5c9c22c |
| internal/configurationstudio/assets/control-language.js | full | present | 100644 | 7aefb5e721c40ef45595337fb30cfaff5e5a9b52 |
| internal/configurationstudio/assets/studio.js | full | present | 100644 | 7562bc6600e8c8e54413277d949e98c28f0410bf |
| internal/configurationstudio/studio_test.js | full | present | 100644 | abd9485768d5e9519896f17cb382f76e570ae82c |
| spec/current-state.md | full | present | 100644 | aea3023e4ce2d1889b0fd86a1dd14e8f08ac54a7 |

- **Tester** — `node --test` для всех пяти Control UI `*_test.js` —
  PASS 199/199. Из них 14 focused проверок прошли пять последовательных
  прогонов, 70/70. Поведение обеих кнопок и прямого submit проверено в Node,
  реальный ввод Enter и фокус — в Edge.
- **Go** — `go test -buildvcs=false -mod=readonly -count=1 ./...` PASS с
  доступом к loopback. Первый запуск в sandbox отказал WebSocket тестам
  на localhost с socket access denied. Повтор с localhost access завершился
  успешно. `CGO_ENABLED=1 go test -mod=readonly -race -count=1 ./internal/configurationstudio ./internal/configurationversion`
  недоступен: в Windows PATH нет `gcc`. Итог verification —
  PASS WITH LIMITATION. Go concurrency code в этом subject отсутствует.
- **Browser smoke** — PASS для финальной сборки Control Service. Локальный
  Edge, две вкладки, отдельный временный loopback server с fixture данными.
  EN: вкладка B публикует Draft через UI, проверены Cancel autofocus и Escape.
  Вкладка A получает lifecycle 409 на Save listener, сохраняет выбор и ввод,
  фокусирует статус. Повторные Enter и clicks не отправляют мутации. Reload
  выполняет только GET. RU: B архивирует Draft через существующий scoped API.
  A сохраняет введённый timeout и выбор, показывает RU статус, фокусирует его
  и блокирует повторные действия. Проверка выполнена через локальный
  `../task150-proof/two-tab-smoke.cjs` с `TASK150_CONTROL_BINARY=control-service-final.exe`.
- **PROCESS-002 and scope audit** — EN/RU Quick Start и пользовательские
  сообщения согласованы. Два production файла, один test, семь documentation
  paths. Generated и посторонние файлы исключены. CRLF-aware
  `git -c core.whitespace=cr-at-eol diff --check` PASS.
- **Independent Reviewer** — Approved для указанного manifest. На предыдущем
  subject Reviewer обнаружил malformed timestamp и неверный уровень scoped
  404. Оба findings исправлены. Reviewer независимо подтвердил 10 paths,
  пересчитал manifest и повторил все 199 тестов. Blocking findings отсутствуют.

### TASK-150 — Coordinator Acceptance and Commit Gate STOP

**Coordinator Acceptance — Accepted** для exact subject. Первый
незавершённый checkpoint — Commit Gate после отдельной команды Owner
`Разрешаю коммит.`. Stage, commit, push, merge не выполнялись. Другие задачи
не активированы. Исходный checkout подтвердил исходный binary diff fingerprint
`d247a12d68547c07e5cd15ea0dfed8ff2f42bc40` всех шести локальных правок.

Финальные `go build -buildvcs=false -mod=readonly ./...` и
`go vet -mod=readonly ./...` PASS. Уточнение к Tester evidence выше:
199/199 — результат всех четырёх Control UI `*_test.js`, число тестов и verdict
сохраняются. Independent Review, PROCESS-002 и scope audit закрыты.
Go race остаётся техническим ограничением. При разрешённом staging требуется
проверить exact bytes с учётом Git newline filters. Смена LF/CRLF недопустима.

Target: `codex/task-150-stale-draft-protection` at
`2854f32244155ed84d1ac37ed9a832103f69efe6`. Object format `sha1`.
Canonical manifest `dcde7290e75ffbbf7fa544738e724bcce87a8011`.
Ordered rows:

```text
.ai/PROJECT_CONTEXT.md | full | present | 100644 | bd7a8b2ad50a458c4bb2ed3a12a1101fcc8d0b73
CHANGELOG.md | full | present | 100644 | a2c783d91344613ea8ddf7b2d0264324e20869cb
docs/en/quick-start.md | full | present | 100644 | 76c3eb7cd029a82708f4a13e32fc97b1567712a9
docs/ru/quick-start.md | full | present | 100644 | 09af9afe7380e3dc2e50d5138ce433b5ba38fdf2
docs/tasks/README.md | full | present | 100644 | 49e343d305914a376d7ffbb650587b4074bf7614
docs/tasks/TASK-150-STUDIO-STALE-DRAFT-PROTECTION.md | task-record-v1 | present | 100644 | 2aa3f762c2832a1df5baf23716fbbe1dd5c9c22c
internal/configurationstudio/assets/control-language.js | full | present | 100644 | 7aefb5e721c40ef45595337fb30cfaff5e5a9b52
internal/configurationstudio/assets/studio.js | full | present | 100644 | 7562bc6600e8c8e54413277d949e98c28f0410bf
internal/configurationstudio/studio_test.js | full | present | 100644 | abd9485768d5e9519896f17cb382f76e570ae82c
spec/current-state.md | full | present | 100644 | aea3023e4ce2d1889b0fd86a1dd14e8f08ac54a7
```

Повторное вычисление manifest локальным `node ../task150-proof/manifest.cjs`
выполняется с raw bytes без staging и без `--normalize`.

### TASK-150 — Documentation-only Owner Acceptance evidence

- Branch `codex/docs/task-150-owner-acceptance`, baseline/HEAD
  `67b74d7760188a073d139a339fef38a7975069a5`. Git object format `sha1`.
- Verification PASS. Две ручные проверки и четыре исключения сверены с
  сообщением Owner. Git подтвердил второй родитель merge и общее дерево с
  коммитом реализации. Исторический envelope сохранён побайтово, SHA-256
  `b25131c03a7d74dd39b93c75bf793018b33e9af87ba6da59652ee7e11e10ca09`.
- Отдельная автоматизированная проверка на опубликованном коде —
  `node --test` для четырёх Control UI `*_test.js`, PASS 199/199.
- `git -c core.whitespace=cr-at-eol diff --check` PASS. Шесть локальных
  ссылок проверены. Scope Audit — 4 Required / 0 Questionable / 0 Removable.
  Fingerprint исходных шести правок `d247a12d68547c07e5cd15ea0dfed8ff2f42bc40`
  подтверждён повторно.

Канонический manifest документационного subject
`2f414363495621a72b9f6a5072bd6390a9786c42` вычислен по NUL-separated rows PROCESS-001.

```text
.ai/PROJECT_CONTEXT.md | full | present | 100644 | 317b495758678b54d8a172e62f49bb34cfb40e86
docs/tasks/README.md | full | present | 100644 | 9f8acd754319649c9857cdb3faaaef041c27dc0b
docs/tasks/TASK-150-STUDIO-STALE-DRAFT-PROTECTION.md | task-record-v1 | present | 100644 | 8483b5296224182d57f55961d52440ef06789525
spec/current-state.md | full | present | 100644 | 7b01e07a7f0ba93ae031e19ff2a5b8d5348b9d8b
```

- Independent Review — **Approved**, 0 blocking / 0 nonblocking findings.
  Reviewer независимо подтвердил все четыре raw rows, сверил Owner evidence,
  опубликованные Git identities, ссылки, PROCESS-002 и границы ручного PASS.
- Coordinator Acceptance документационного subject — **Accepted**
  27 ноября 2026 года для указанного manifest. PROCESS-002 — Synchronized.
  Первый незавершённый checkpoint — Commit Gate после отдельной команды
  Owner `Разрешаю коммит.`. Staging, commit, push и merge этого среза ещё не
  выполнялись. Следующие задачи не активированы.

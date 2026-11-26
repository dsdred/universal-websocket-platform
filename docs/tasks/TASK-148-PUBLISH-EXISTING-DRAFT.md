# TASK-148 — Публикация существующего Draft из Studio

## Status

Реализация опубликована. **Owner Acceptance PASS** для опубликованной версии.
Документационное закрытие — `In Progress`. Состояние Commit Gate определяется
по актуальной записи в Recovery Evidence Envelope.

## Task Contract

- Task Mode: Implementation. Why Now: TASK-147 Owner Acceptance PASS, TASK-148 указан следующим срезом в task index.
- Branch: `feature/task-148-publish-existing-draft`. Baseline: `f2632cae9970c85d920c81c458fd6618c669b18b`. Изолированный worktree `bin/task148-worktree` сохраняет изменения основной рабочей копии.
- Coordinator фиксирует intake и scope. Developer реализует утверждённый срез, Tester проверяет критерии, Documentation Agent синхронизирует документы. Независимый Reviewer проверяет итоговый diff перед Coordinator Acceptance.
- Definition of Done: отдельная публикация открытого, чистого Draft после EN/RU подтверждения с идентификаторами и предупреждением об архивировании; свежий scoped GET, один POST, обновление списка; блокировка при dirty, unknown Save, pending, stale или смене scope; read-only reconciliation при unknown outcome; TLS-enabled Draft и отсутствие Runtime запросов.
- Out of Scope: изменения Control Service / API, CAS, новый Runtime lifecycle, TASK-131, редизайн, commit, push и merge.
- Verification Plan: Studio proof tests, Go test/build/vet/race, браузерный smoke с двумя вкладками, EN/RU и клавиатурой, PROCESS-002, scope audit и независимый review. STOP у Commit Gate.

## Architecture Confirmation

Owner утвердил Studio-only reuse существующих `GET .../versions` и `POST .../versions/{versionID}/publish`. `configurationversion.Service.Publish` сохраняет ownership и atomic Published/Archived transition. ARCH-005 разделяет публикацию и запуск Runtime. `updatedAt` и открытые поля используются для best-effort freshness. GET после uncertain POST показывает наблюдаемое состояние; исход POST остаётся неизвестным. Новый POST требует свежего чтения и нового подтверждения.

## Existing Coverage Report

- Existing Coverage: `studio_test.js` покрывает Draft-only discovery, чистую/dirty форму, listener/timeout Save, unknown Save, stale/aborted GET, защиту scope и dialog Escape. `handler_test.go` проверяет embedded UI, разрешённые маршруты и create-and-publish. `configurationversion` проверяет publish/archives, routing validation и атомарность batch.
- Coverage Gap: отдельная команда публикации, preflight concurrency, повторный click/confirm, reconciliation, TLS-enabled Draft, refresh failure и локализация нового диалога. Новые proof/regression tests добавляются в существующие Studio suite и embedded contract.

## Implementation and Scope Audit

| Классификация | Файлы | Обоснование |
| --- | --- | --- |
| Required production | `internal/configurationstudio/assets/index.html`, `studio.css`, `studio.js`, `control-language.js` | Отдельное действие, переиспользование modal и его визуального стиля, EN/RU, preflight, один scoped POST, refresh/reconciliation. |
| Required verification | `internal/configurationstudio/studio_test.js`, `internal/configurationstudio/handler_test.go` | Сценарии UI и проверки embedded surface, включая Cancel autofocus. |
| Required documentation | `docs/tasks/TASK-148-PUBLISH-EXISTING-DRAFT.md`, `docs/tasks/README.md`, `docs/en/quick-start.md`, `docs/ru/quick-start.md`, `spec/current-state.md`, `.ai/PROJECT_CONTEXT.md`, `CHANGELOG.md` | Task/recovery anchor, navigation, зеркальная инструкция, фактическая capability и user-facing changelog. |

Size Guard: 13 файлов, один independently deliverable slice, без новых package или архитектурного контракта. Generated и unrelated files в subject отсутствуют.

PROCESS-002: task record, current state, project context, task index, EN/RU quick-start и changelog применимы. MASTER_PLAN N/A: milestone boundary и engineering dependencies сохраняются. Design Proposals, ADR, `spec/decisions.md` и root README N/A: архитектурный и API contracts сохраняются, пользовательский сценарий документирован в quick-start.

Ограничения: между preflight GET и POST нет CAS. Отображаемые поля и точная строка `updatedAt` обнаруживают наблюдаемые изменения. Успешный section Save обновляет `updatedAt`; переключение других редактируемых значений в Save response требует отдельного reload. Для изменений, которые предшествуют Save в неотображаемых полях, и для изменений после preflight автоматическая гарантия отсутствует. Unknown publication хранится в текущей странице; после нового открытия Draft выполняется новое явное чтение и подтверждение. Состояние из GET не атрибутирует прежний POST.

## Owner Acceptance Closure

26 ноября 2026 года Owner принял опубликованную реализацию TASK-148. Источник
ручных наблюдений — сообщение Owner `01a216ab-850a-74d2-985b-389c23a00acf`
в turn `01a216ab-8382-7c33-9cb0-79139a98af18` чата
`01a21469-86e6-79e3-aff7-6520e76504e9`.

[PR #156](https://github.com/dsdred/universal-websocket-platform/pull/156)
слит в `main`. Исходный коммит `01e6c811afb19e7ee0b9770d565fc56d634c9c15`,
merge SHA `998eefb8dff8231b4eb0dff0bb98ed3e824d1597`, общее дерево
`98aa3829ad37c746cd1df77e414c9ef31ea82e08`. Обязательный `go-ci`, включая
race-тесты, прошёл на [PR](https://github.com/dsdred/universal-websocket-platform/actions/runs/45764656019)
и на [main](https://github.com/dsdred/universal-websocket-platform/actions/runs/45764851162).

Ручная проверка выполнена на изолированном Control Service `127.0.0.1:18148`
для Workspace `TASK-148 Acceptance` (#1) и Configuration
`Publish Existing Draft` (#1). Начальное состояние — v1 (#1) `Published`
на `127.0.0.1:19148`, v2 (#2) `Draft` на `127.0.0.1:19248`.

| Ручной сценарий | Результат Owner |
| --- | --- |
| Идентичность перед публикацией | Publish Draft показывает выбранную Configuration и Draft v2 (#2). |
| Публикация | После подтверждения Draft v2 (#2) стала `Published`. |
| Предыдущая версия и Dashboard | v1 (#1) стала `Archived`. Dashboard показывает `Published` для v2 и `Archived` для v1. |
| Runtime на проверяемых портах | Порты 19148 и 19248 не слушались. Публикация не запустила Runtime на этих портах. |
| Cancel | Отдельного визуального подтверждения Owner нет. Сценарий не засчитан в ручную приёмку. |

**Owner Acceptance PASS** основана на перечисленных наблюдениях. Покрытие
Cancel в автоматических тестах записано отдельно в evidence реализации.
Наблюдение о Runtime ограничено портами 19148 и 19248.

### Документационный контракт

- **Mode** — Documentation-only. Branch `docs/task-148-owner-acceptance`,
  worktree `bin/task148-owner-closure`, baseline и HEAD перед изменениями
  `998eefb8dff8231b4eb0dff0bb98ed3e824d1597`.
- **Why Now** — Owner подтвердил вручную опубликованную реализацию.
- **Definition of Done** — зафиксировать решение Owner и ограничения,
  синхронизировать статус и навигацию, сверить ссылки и exact manifest,
  завершить независимый Review документационного diff.
- **Scope** — текущий task record, `docs/tasks/README.md`,
  `.ai/PROJECT_CONTEXT.md` и `spec/current-state.md`. Product, API, Runtime и
  Control Services вне scope. Следующая задача не активируется.
- **Verification Plan** — проверить локальные Git objects и scope, сверить
  наблюдения с сообщением Owner, сохранить исторический envelope, проверить
  четыре документа и передать exact subject независимому Reviewer.
- **PROCESS-002** — task record, index, project context и current state
  применимы. MASTER_PLAN EN/RU, ADR/DP, `spec/decisions.md`, Quick Start,
  корневой README и CHANGELOG неприменимы: приёмка фиксирует статус уже
  опубликованной функции без изменения milestone, архитектуры, инструкций
  пользователя или состава релиза.
- **Size Guard / STOP** — четыре документационных пути. Coordinator,
  Documentation Agent и независимый Reviewer. Остановка до stage и commit.

## Recovery Evidence Envelope

### Verification and scope audit

- Repository `universal-websocket-platform`, branch `feature/task-148-publish-existing-draft`, baseline/HEAD `f2632cae9970c85d920c81c458fd6618c669b18b`, Git object format SHA-1.
- `node --test internal/configurationstudio/*_test.js` PASS, 180/180. Добавлены 12 scenarios в существующей suite: success и archive, Cancel/Escape и changed scope, dirty/unknown Save, two-tab stale, unavailable target, reload, отдельный Save, 409, unknown publish и read-only reconciliation, refresh failure, supersession после POST и EN/RU. Embedded Go проверяет отдельную кнопку и Cancel autofocus.
- `go test -mod=readonly -count=1 ./...` PASS в разрешённом loopback context. Default sandbox отклонял loopback WebSocket в существующих Go tests. `go build -mod=readonly ./...` и `go vet -mod=readonly ./...` PASS.
- `CGO_ENABLED=1 go test -mod=readonly -race -count=1 ./...` — недоступен: Windows Go не нашёл `gcc`. Verification verdict: PASS WITH LIMITATION. Async Studio guards отдельно проверены в Node и браузере.
- Edge headless + временный Control Service с новым loopback port и synthetic data: PASS. Вкладки A/B открывают TLS-enabled Draft при существующей Published версии. В A клавиатурный Enter открывает dialog, Cancel получает начальный фокус, Escape возвращает фокус, Enter на Cancel закрывает dialog. В B dirty блокирует Publish и Save меняет port. Подтверждение в A останавливается на stale preflight. Reload в A показывает port из B. RU dialog, Tab на Publish и Enter дают один POST, focus на status, Published target с прежним TLS, Archived predecessor, ноль Runtime requests. Это proof на запущенном бинарнике worktree; действующие сервисы не изменялись.
- `git -c core.whitespace=cr-at-eol diff --check` PASS; diff scope ограничен 13 paths. Без stage, commit, push, merge или нового API. LF/CRLF bytes сохранены; для последующего exact Commit Gate raw hashes совпадают с `git -c core.autocrlf=false hash-object` на всех путях. Default conversion для части путей меняет bytes, поэтому будущий staged tree проверяется с сохранением raw identity.

### Exact subject

`bfa1c5baf1249fb8f07fb6aaa0d3e475acae3f7f` — `git hash-object --stdin` от отсортированных UTF-8 NUL rows по PROCESS-001. Task record использует `task-record-v1`; остальные OID вычислены `git hash-object --no-filters`. Envelope исключён из projection.

```text
.ai/PROJECT_CONTEXT.md | full | present | 100644 | 553cba2a2dd29285cf79407cfdc32f3bb1d1eb73
CHANGELOG.md | full | present | 100644 | 275e7401760dbe6de9c6d52a4e004171170a6bed
docs/en/quick-start.md | full | present | 100644 | bc42c2e36c3d12390ee9da108fcb697cea5944eb
docs/ru/quick-start.md | full | present | 100644 | 43af396959850091b6a55684855ed69d68116b7f
docs/tasks/README.md | full | present | 100644 | 9ce63b1ecc5da396826ceafa7238cdd66ad00165
docs/tasks/TASK-148-PUBLISH-EXISTING-DRAFT.md | task-record-v1 | present | 100644 | 13b0ed2312aaf05fd26e29d6abf7d5a175a87df5
internal/configurationstudio/assets/control-language.js | full | present | 100644 | dd90ab11acc0d1830d196adaa53e272916db4015
internal/configurationstudio/assets/index.html | full | present | 100644 | 73e4a765f94512d89738ee97382d7aa2abf05d9b
internal/configurationstudio/assets/studio.css | full | present | 100644 | 2f45492f3e872c05c49ac8b06a65dac67ff756d1
internal/configurationstudio/assets/studio.js | full | present | 100644 | a7d765d848e67a34edc480ba79c1b38192939080
internal/configurationstudio/handler_test.go | full | present | 100644 | 85dbf58806e830d769140aa9f37c4751e32d81c5
internal/configurationstudio/studio_test.js | full | present | 100644 | 1fd6c21e0b70812a67a46a595d85cc52ce13f2fb
spec/current-state.md | full | present | 100644 | 3f5e96731d1c2cfadc621edf5b31f2eccd91a231
```

### Independent Review and Coordinator Acceptance

Reviewer **Approved**, 0 open findings, exact 13-path subject
`bfa1c5baf1249fb8f07fb6aaa0d3e475acae3f7f`. Reviewer независимо пересчитал
manifest, подтвердил scope каждого файла и заново выполнил Node tests 180/180,
CRLF-aware diff check и проверку пустого staged set. Исправления result heading,
current state после supersession и снятие eligibility после stale reload
подтверждены. Ограничение Windows race сохраняется.

Coordinator **Accepted with verification limitation** для этого subject.
PROCESS-002 завершён, Scope Audit: 13 Required, 0 Questionable, 0 Removable.
STOP у Commit Gate в ожидании отдельного `Разрешаю коммит.`. До него отсутствуют
stage, commit, push, merge и publication. Рекомендация после принятого diff —
отдельный commit gate, затем публикация по отдельному разрешению. Следующая
task не активирована.

### Final review and acceptance after restoring baseline line endings

Эта запись supersedes предыдущий exact subject. При проверке patch восстановлены
baseline LF только в `control-language.js` и `studio.css`. Обратное преобразование
этих двух файлов в CRLF воспроизводит оба предыдущих raw OID. Остальные
11 projected rows совпадают. Поведение не менялось.

На актуальном subject Studio Node 180/180, targeted Go
`go test -mod=readonly -count=1 ./internal/configurationstudio ./internal/configurationversion`
и CRLF-aware diff check прошли. Ранее завершённые полный Go suite и двухвкладочный
Edge smoke проверили семантически идентичную реализацию. Windows race остаётся
недоступным из-за отсутствия `gcc`.

Свежий independent Review: **Approved**, 0 open findings. Reviewer независимо
пересчитал manifest, подтвердил EOL-only delta и Scope Audit, повторил Node
180/180 и scoped diff check, подтвердил пустой staged set.

Coordinator **Accepted with verification limitation**. Тот же branch и HEAD,
13 Required paths, 0 Questionable, 0 Removable. Exact SHA-1 manifest
`3f2c46b2c61ec35deb98b5f6badb70493ee1e675`. Будущий Commit Gate проверяет raw
identity с command-local `core.autocrlf=false` только для exact path set и
отдельно проверяет envelope. STOP у Commit Gate до `Разрешаю коммит.`.
Stage, commit и publication не выполнялись.

```text
.ai/PROJECT_CONTEXT.md | full | present | 100644 | 553cba2a2dd29285cf79407cfdc32f3bb1d1eb73
CHANGELOG.md | full | present | 100644 | 275e7401760dbe6de9c6d52a4e004171170a6bed
docs/en/quick-start.md | full | present | 100644 | bc42c2e36c3d12390ee9da108fcb697cea5944eb
docs/ru/quick-start.md | full | present | 100644 | 43af396959850091b6a55684855ed69d68116b7f
docs/tasks/README.md | full | present | 100644 | 9ce63b1ecc5da396826ceafa7238cdd66ad00165
docs/tasks/TASK-148-PUBLISH-EXISTING-DRAFT.md | task-record-v1 | present | 100644 | 13b0ed2312aaf05fd26e29d6abf7d5a175a87df5
internal/configurationstudio/assets/control-language.js | full | present | 100644 | e22cbaf6afef2db7e602f5f11cd3f76546d93ddb
internal/configurationstudio/assets/index.html | full | present | 100644 | 73e4a765f94512d89738ee97382d7aa2abf05d9b
internal/configurationstudio/assets/studio.css | full | present | 100644 | 5b693290f84814370b3078ccdbdcf75c41c69337
internal/configurationstudio/assets/studio.js | full | present | 100644 | a7d765d848e67a34edc480ba79c1b38192939080
internal/configurationstudio/handler_test.go | full | present | 100644 | 85dbf58806e830d769140aa9f37c4751e32d81c5
internal/configurationstudio/studio_test.js | full | present | 100644 | 1fd6c21e0b70812a67a46a595d85cc52ce13f2fb
spec/current-state.md | full | present | 100644 | 3f5e96731d1c2cfadc621edf5b31f2eccd91a231
```

### Owner Acceptance documentation — Review and Commit Gate

- Branch `docs/task-148-owner-acceptance`, HEAD и base
  `998eefb8dff8231b4eb0dff0bb98ed3e824d1597`. SHA-1 manifest
  `2ab8b65ce085a190e057169fa706916b88f05e99`, UTF-8 NUL-separated rows,
  `task-record-v1` для task record, raw blob OID для остальных файлов.
- Проверены Git ancestry и дерево опубликованной реализации, состояние PR #156
  и успешные hosted `go-ci` с race на PR и на main. Проверены локальные ссылки,
  исходные окончания строк, четыре пути и пустой индекс.
  CRLF-aware `git diff --check` прошёл. Product tests повторно не запускались
  для документационного diff.
- Неизменный исторический envelope реализации — 8667 bytes, SHA-256
  `00424abcbe9b2dd12671fa93c141204d97e929e325dcc25d1cd55b74220d5601`.
  Все шесть исходных локальных правок сохранены побайтово.
- Независимый Reviewer **Approved**, 0 findings. Reviewer независимо
  пересчитал manifest, сверил сообщение Owner, ограничения Cancel и Runtime,
  опубликованную идентичность, статус, ссылки и scope. PROCESS-002
  синхронизирован. Scope Audit — 4 Required, 0 Questionable, 0 Removable.
- Coordinator **Accepted** для этого документационного subject.
  **STOP на Commit Gate**. Stage, commit, push, PR и merge этого diff
  не выполнялись. Состояние Control Services не изменялось. Следующая задача
  не активирована.

```text
.ai/PROJECT_CONTEXT.md | full | present | 100644 | 4a259f890a0081b037b22d475f39c2bf2e4f64f7
docs/tasks/README.md | full | present | 100644 | ac0eafff8587facb3a54f90f9e4679b22380a804
docs/tasks/TASK-148-PUBLISH-EXISTING-DRAFT.md | task-record-v1 | present | 100644 | 45ee2a90530042c5c7d605885c6124b2b1c9315c
spec/current-state.md | full | present | 100644 | 683639b115443e029ddfa49880572d6025a2e51b
```

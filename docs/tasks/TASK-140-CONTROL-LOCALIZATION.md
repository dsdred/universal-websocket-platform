# TASK-140 — Локализация Studio и Dashboard

## Status

`In Progress`. Текущий gate определяется последней применимой записью
Recovery Evidence Envelope.

## Task Contract

- Task Mode: Implementation в существующем встроенном интерфейсе Control.
- Why Now: Owner выбрал первый UX-кандидат TASK-138 после 12/12 PASS.
  PR #138 присутствует в локальном Git как merge
  `47b19f97f2fbf3e01e99ab261e20fc1d59b5582d`.
- Definition of Done: обе страницы полностью доступны на EN/RU. При
  отсутствии сохранённого выбора и недостающем переводе используется EN.
  Смена языка сохраняет выборы, введённые поля, ошибки и результаты Save,
  не перезагружает страницу и не отправляет запросы. Сценарии TASK-138,
  данные и правила проверки сохраняются.
- Out of Scope: темы, hamburger, оформление диалогов, переработка проверки
  таймаутов, новый layout Studio, Runtime management и следующие UX-задачи.
- Verification Plan: Node proof и регрессии, Go tests/build/vet, браузерный
  smoke на desktop/mobile, сверка области изменений и независимый review.

## Recovery anchor

- Репозиторий: Universal WebSocket Platform, `E:\wikiPRJ\universal-websocket-platform`.
- Ветка `feature/task-140-control-localization` создана с чистого локального
  `main == origin/main` на `47b19f97f2fbf3e01e99ab261e20fc1d59b5582d`.
- Изменения ограничены встроенными ресурсами и тестами
  `internal/configurationstudio`, task record/index, project context, factual
  current state, CHANGELOG и зеркальным Quick Start.
- Coordinator назначил реализацию, тесты и синхронизацию основной сессии.
  Архитектурная проверка и независимый review выполняются отдельными агентами.
  Publisher не назначен.
- Источники: [ADR-0001](../en/adr/0001-bootstrap-control-service.md),
  [ADR-0002](../en/adr/0002-configuration-dsl.md),
  [TASK-076](TASK-076-LOCAL-CONFIGURATION-STUDIO.md),
  [TASK-138](TASK-138-STUDIO-OWNER-ACCEPTANCE.md),
  [TASK-139](TASK-139-LISTENER-VALIDATION-UX.md).
- Этапы: intake → Documentation Baseline → Architecture Confirmation →
  implementation → verification → independent review/rework → PROCESS-002 →
  scope audit/final checks & review → Coordinator decision → Pre-Commit STOP.
- TASK-134 и TASK-102 вне scope, TASK-131 DEFERRED, Gate 11 OPEN, защищённые
  этапы Not Activated. Следующий UX-кандидат (темы) Not Activated.
- Stage, commit, push, PR, merge, fetch/pull и publication запрещены текущим
  разрешением. STOP при конфликте источников, выходе за scope, блокирующей
  проверке или на Pre-Commit Gate.

## Architecture Confirmation

Независимый Architect, `APPROVED`, 0 blocking findings, baseline
`47b19f97f2fbf3e01e99ab261e20fc1d59b5582d`. Общий plain JavaScript и встроенный
same-origin статический ресурс укладываются в существующую границу.
Выбор языка хранится только в браузере. Представление переводит известные
статусы и ошибки без изменения идентификаторов, данных и API. Новый ADR/DP и
pre-implementation documentation не применяются: контракт архитектуры и API
сохраняется.

## Existing Coverage Report

Node-тесты покрывают три режима создания, выборы, Draft open/save, локальные
ошибки и серверные отказы, защиту несохранённых значений, конкурентные
ответы и Dashboard GET-only browsing. Go-тесты проверяют точные маршруты,
встроенные ресурсы и CSP. TASK-138 содержит 12 ручных PASS на базовой версии.
Пробелы: переключение языка при текущих данных и feedback, сохранение выбора
между страницами, fallback, сетевое отсутствие побочных эффектов и размеры
локализованного интерфейса.

## Реализация и scope audit

Селектор в шапке обеих страниц управляет локальным предпочтением
`uwp-control-language`. EN задаёт исходный текст и fallback. Каталог
обновляет текст, заголовок документа, `lang` и доступные названия на месте.
Ключи и параметры уже показанных сообщений остаются доступны для повторного
перевода. При восстановлении страницы из кэша истории перечитывается
сохранённый выбор. DOM-элементы формы и запросы ведут прежний жизненный цикл.
Идентификаторы, пользовательские имена, параметры ресурса и payload
передаются как значения. Известные причины отказа Listener отображаются
локализованно, для новой причины используется общий текст для того же поля.

Size Guard пересмотрен при 17 файлах. Это один проверяемый срез: сквозная
настройка языка требует обе страницы, общий каталог, регистрацию статического
ресурса и тесты текущих сценариев. Документы дают навигацию, инструкцию
пользователю и доказательства работы по PROCESS-001/002. Разделение между
страницами оставило бы частично переведённую оболочку при переходе между ними.
Один существующий пакет, новых архитектурных контрактов нет.

- Production (7 Required): `assets/control-language.js`, `assets/control.css`,
  `assets/index.html`, `assets/studio.js`, `assets/dashboard.html`,
  `assets/dashboard.js`, `handler.go` в `internal/configurationstudio`.
- Tests (3 Required): `studio_test.js`, `dashboard_test.js`, `handler_test.go`
  в том же пакете. Новые проверки покрывают Draft + feedback, отказ Listener,
  EN fallback и статический каталог, Dashboard и восстановление из кэша истории.
- Documentation (7 Required): task record/index, `.ai/PROJECT_CONTEXT.md`,
  `spec/current-state.md`, `CHANGELOG.md` и обе страницы Quick Start.
  Canonical paths и bytes перечислены в terminal envelope. Scope Audit: 17
  Required, 0 Questionable, 0 Removable.

PROCESS-002: task record и index нужны для intake/recovery, project context
для активной задачи, current-state и CHANGELOG для новой возможности,
EN/RU Quick Start для инструкции на обеих страницах. MASTER_PLAN N/A:
зависимости и границы milestones сохраняются. DP/ADR и `spec/decisions.md`
N/A: новые нормативные решения отсутствуют. Корневые README N/A: назначение,
ссылки и способ запуска актуальны. Исторические записи TASK-138/139 сохраняются.

## Recovery Evidence Envelope


### Pre-Commit Report — 2026-11-19

- Coordinator Acceptance: `Accepted` для subject
  `4b06ceb593b03298834f34a6f2c715c5fd9f427c`. Следующий gate — отдельное
  разрешение Owner на commit. Текущая сессия остановлена на Pre-Commit Gate
  с пустым index. Разрешение на публикацию требуется отдельно.
- Baseline и HEAD: `47b19f97f2fbf3e01e99ab261e20fc1d59b5582d` (PR #138),
  ветка `feature/task-140-control-localization`. Object format: SHA-1.
- Verification: PASS для Node 148/148, Go build/test/vet, форматирования
  изменённых Go-файлов, browser smoke, проверки ссылок и diff.
  Race: локальный запуск недоступен из-за отсутствующего `gcc` в PATH
  (`CGO_ENABLED=1`). Новые concurrency и synchronization paths отсутствуют.
- Independent final Review: `APPROVED`, 0 open findings. Reviewer
  независимо пересчитал те же 17 ordered rows, повторил Node (148/148),
  Go-тесты `internal/configurationstudio`, diff-check и проверку 11 ссылок
  и парности структуры/семантики Quick Start. Finding для кэша истории
  закрыт изменением и отдельной проверкой. Scope Audit: 17/0/0.
- TASK-134 и TASK-102 вне изменений. TASK-131 DEFERRED, Gate 11 OPEN,
  защищённые этапы Not Activated. Тема интерфейса остаётся отдельным
  кандидатом Not Activated.

#### Проверки и воспроизводимость

| Проверка | Результат |
|---|---|
| `node --test internal/configurationstudio/studio_test.js internal/configurationstudio/dashboard_test.js` | Exit 0, 148 PASS. Регрессии трёх режимов создания, выбора/открытия/сохранения Draft, Cancel и discard, валидации, запаздывающих ответов и Dashboard GET. Новые сценарии проверяют локализованное состояние, fallback и кэш истории. |
| `go build -mod=readonly ./...`, `go vet -mod=readonly ./...`, `gofmt -l internal/configurationstudio/handler.go internal/configurationstudio/handler_test.go` | Exit 0. Gofmt без вывода. |
| `go test -mod=readonly -p 6 -count=1 ./...` | Exit 0 с loopback access. Первый запуск под сетевым sandbox получил запрет на локальный TCP; полный повтор завершился успешно. |
| `go mod tidy -diff` с `GOFLAGS=-modfile=.tmp/task140-go.mod` | Exit 0 на временных копиях `go.mod`/`go.sum` с LF из Git HEAD. Обычный запуск на Windows checkout показал только CRLF/LF в `go.sum`. `git diff --exit-code HEAD -- go.mod go.sum` — exit 0. |
| `CGO_ENABLED=1 go test -mod=readonly -race -count=1 ./...` | Exit 1 до исполнения тестов: `cgo: C compiler "gcc" not found`. |
| `git -c core.autocrlf=false -c core.whitespace=cr-at-eol diff --check` | Exit 0. Отдельно проверены raw bytes обоих новых файлов, пустой index и exact path set. |

Browser smoke: Chrome 157.0.8103.51 (CDP), чистый профиль, отдельный
Control Service на `127.0.0.1:18140`, тестовые данные в памяти. Для
повтора собрать `go build -mod=readonly -buildvcs=false ./cmd/control-service`
и запустить сервис с `UWP_HTTP_HOST=127.0.0.1`, `UWP_HTTP_PORT=18140`.
Открыть Studio и Dashboard с пустым localStorage. Проверенные шаги:

1. EN по умолчанию. RU → EN → RU обновляет заголовки, доступное имя навигации
   и форму. Ввод имени `Published {0}` сохраняется при смене языка.
2. Создать и опубликовать конфигурацию в новой области, затем создать и
   опубликовать другую конфигурацию в этой области. В первой конфигурации
   создать только Draft. Результаты видны на RU, всего 10 ожидаемых записей API.
3. Открыть Draft ID 3. Published ID 1 отсутствует в редакторе. Ввести
   `127.0.1.2.3`, выполнить Save, сменить RU ↔ EN. Сохраняются значение,
   область, конфигурация, Draft и выбранный раздел, а inline ошибка и
   сообщение меняют язык. Исправить на `localhost:8139` и сохранить (один
   PUT в тот же Draft). Сменить язык после Save и проверить сообщение.
4. В разделе таймаутов ввести `301` и выполнить Save. Сменить язык,
   отменить перезагрузку в локализованном confirm, убедиться в сохранении
   ввода и отсутствии записи. Исправить на `25`, сохранить (один PUT), сменить
   язык и перечитать Draft. Тот же ID остаётся Draft. GET подтвердил
   `localhost:8139` и таймаут рукопожатия `25`.
5. Перейти в Dashboard с сохранённым RU, выбрать первую конфигурацию,
   проверить состояния и адрес, сменить язык с сохранением выбора, выполнить
   Refresh и вернуться через шапку в Studio с RU.
6. Ширина 1280 и 390 px (эмуляция Chrome). Переключатель помещается в шапке,
   у документа нет горизонтального overflow, таблица прокручивается внутри
   своей области. Шесть затронутых ресурсов совпали byte-for-byte с
   содержимым репозитория.
7. CDP зафиксировал 13 смен языка, 0 запросов при них, 3 ожидаемых запроса
   документов, 0 JS exceptions и ровно 12 записей API от явных действий.
   Сервис и браузерный профиль для проверки были отдельными от Owner fixture.

#### Exact file manifest

Пути в возрастающем порядке UTF-8. Для каждой строки последовательность
`path\0projection\0state\0mode\0oid\0` хешируется через
`git hash-object --stdin`. OID full-файлов получены без фильтров,
task record использует `task-record-v1` PROCESS-001.

```text
.ai/PROJECT_CONTEXT.md | full | present | 100644 | 785079ba662027d814f5e3d1891f58b727a2557f
CHANGELOG.md | full | present | 100644 | d0e4a7576fc7161050f3ab3f9c99994549f9e760
docs/en/quick-start.md | full | present | 100644 | 35fc62a087b2a8a7c42bc37a5caebae9ad989205
docs/ru/quick-start.md | full | present | 100644 | 7c7c85a0ae1cf63207a4d3aa8ea039963c5350b1
docs/tasks/README.md | full | present | 100644 | 7cabb085ef6134ea89fc4f020a65e01d863dc275
docs/tasks/TASK-140-CONTROL-LOCALIZATION.md | task-record-v1 | present | 100644 | d156ce37a74cc8cf27480f91656b0b800f8d53d8
internal/configurationstudio/assets/control-language.js | full | present | 100644 | cff60dbc77bfc6cce637ee9f272af7db0c52afd3
internal/configurationstudio/assets/control.css | full | present | 100644 | 57fe8fd89314b789875702de97eceb7f55677aad
internal/configurationstudio/assets/dashboard.html | full | present | 100644 | 9be0d11934fe7b33897df17451e27accc283bc5c
internal/configurationstudio/assets/dashboard.js | full | present | 100644 | 73e0ae6030b6351f3a86a342e3c8a80838557d87
internal/configurationstudio/assets/index.html | full | present | 100644 | 3787bbcfaa1855b22164d278b0bb1968e72ff7c6
internal/configurationstudio/assets/studio.js | full | present | 100644 | b4834aad0a1ccccad533bfb687270f34d918f0c9
internal/configurationstudio/dashboard_test.js | full | present | 100644 | 1644b4b8d77fb10025d6ef656699e39a9e353461
internal/configurationstudio/handler.go | full | present | 100644 | e57e52c2470663d3657e437691cbb3234e935a27
internal/configurationstudio/handler_test.go | full | present | 100644 | cab43ba74ae870c186d075152c655270e642a995
internal/configurationstudio/studio_test.js | full | present | 100644 | 35c2355d9291ac293fd7289cbd25a549e32b8286
spec/current-state.md | full | present | 100644 | 75ee97757cd0f429c5cd8c1f21fd83bebe29f933
```

### CG140-001 — Owner Commit Gate

Получена точная команда Owner `Разрешаю коммит.`. Повторные проверки
подтвердили исходный HEAD, принятый subject из 17 файлов, 148 Node-тестов,
Go-тесты пакета Studio и форматирование изменённых Go-файлов.
Первичная проверка index обнаружила пустую строку в конце нового task record.
Эта append-only запись сохраняет прежние байты и принятую проекцию.
Перед коммитом повторно сверяются все байты и режимы index и его tree.
Сообщение коммита: `feat(control): add English and Russian localization`.

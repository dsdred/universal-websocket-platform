# TASK-140 — Упрощение шапки по замечанию Owner

## Status

`In Progress`. Текущий gate определяется Recovery Evidence Envelope.

## Task Contract

- Режим: Implementation в рамках текущей ручной приёмки TASK-140.
  После PR #140 Owner указал на лишний символ рядом с EN/RU и неясную
  кнопку «Панель управления» в шапке Studio.
- Результат: в шапках Studio и Dashboard переключатель показывает текущий
  EN или RU текстом. Ссылки между страницами остаются в основной навигации.
  Клик, Enter/Space, локализованное доступное имя, фокус, сохранённый язык
  и правки Draft работают на обеих страницах и при ширине 390 px.
- Scope: две шапки и их общие стили, устаревший ключ перевода, применимые
  документы. API, Runtime, TASK-141 и работающие Owner-инстансы вне scope.
- Проверки: Node-регрессии, Go-тесты, браузерный smoke и независимый review.

## Recovery anchor

- Ветка `feature/task-140-owner-header-cleanup` от чистого
  `main == origin/main`, `1872bb59bc76f59ada2241f01f3603161990761c` (PR #140).
- Основание: [предыдущая корректировка](TASK-140-OWNER-UX-CORRECTION.md),
  [ADR-0001](../en/adr/0001-bootstrap-control-service.md) и уточнение Owner.
- Coordinator назначил основную сессию на реализацию, тесты и синхронизацию.
  Architect и Reviewer проверяют независимо. Intake → Architecture →
  implementation → tests → PROCESS-002 → final review → Pre-Commit STOP.
- Stage, commit и publication требуют отдельных команд Owner.
  TASK-141 Not Activated. TASK-134, TASK-131, TASK-102, Gate 11 и защищённые
  этапы вне scope.

## Existing Coverage Report

148 Node-тестов проверяют переключение, состояние Draft и Dashboard,
локализованную обратную связь и localStorage. Go проверяет встроенные ресурсы.
Браузерная проверка покрывает изменение состава шапки, переходы через
основную навигацию, переключение в обеих локалях и ширину 390 px.

## Architecture Confirmation и scope audit

Независимый Architect: `APPROVED`, 0 blocking findings. Существующий
нативный переключатель и его обработчик сохраняются. Обе страницы используют
основную навигацию и общий адаптивный стиль. Удалены декоративный SVG,
дублирующие ссылки в шапках и ключ `Open Studio`, использовавшийся там.
Отдельный ADR/DP и изменения API неприменимы.

11 Required, 0 Questionable, 0 Removable. Четыре frontend-ресурса нужны
для двух симметричных шапок и удаления их неиспользуемых стилей и текста.
Семь документов покрывают task record, task index, project context, current
state, changelog и зеркальный Quick Start. Исторические task records
сохранены. MASTER_PLAN, `spec/decisions.md` и корневые README неприменимы:
границы вех, архитектурные решения и порядок запуска сохранены.

## Recovery Evidence Envelope


### Pre-Commit Report — 2026-11-19

- Coordinator Acceptance: `Accepted`, манифест `1e5466ebb7e48445cad610f2b3cced33d73dfc09`.
  Ветка `feature/task-140-owner-header-cleanup`, HEAD и база
  `1872bb59bc76f59ada2241f01f3603161990761c`. Index пуст. STOP на Pre-Commit Gate.
- Independent final Reviewer: `APPROVED`, 0 findings. Повторный расчёт
  байтового манифеста, Node 148/148, восемь относительных ссылок и
  парность Quick Start подтверждены. Scope Audit 11/0/0.
- `node --test --test-reporter=dot internal/configurationstudio/studio_test.js internal/configurationstudio/dashboard_test.js`,
  `go test -mod=readonly -count=1 ./internal/configurationstudio` и
  `git -c core.autocrlf=false -c core.whitespace=cr-at-eol diff --check` — exit 0.
- `go build -mod=readonly -buildvcs=false -o <temp-exe> ./cmd/control-service`
  — exit 0. Smoke на отдельном 18143 через CDP, Chrome 158.0.8127.36.
  Шесть встроенных ресурсов совпали с checkout byte-for-byte. При 1280 и
  390 px в шапках по одной нативной кнопке EN/RU без SVG и ссылок;
  переключение мышью, Enter и Space, фокус, основная навигация, locale после
  refresh и правки Draft проверены. Зафиксированы 13 переключений, 0
  запросов при них и 0 исключений JavaScript. Четыре запроса документов
  вызваны явными переходами и refresh; 12 записей API относятся к явным
  действиям с тестовыми данными на 18143. Тестовый сервис остановлен.
- Owner-инстансы на 18138–18141 сохранили PID
  `37772`, `21044`, `24148` и `30028`. TASK-141 Not Activated.

#### Exact file manifest

SHA-1. Пути отсортированы по raw UTF-8. Строки
`path\0projection\0state\0mode\0oid\0` хешируются `git hash-object --stdin`.
Записи `full` используют байты без фильтров, task record использует
`task-record-v1` PROCESS-001.

```text
.ai/PROJECT_CONTEXT.md | full | present | 100644 | 6fac8a09afa65277bcd93d53823de9e44527ee44
CHANGELOG.md | full | present | 100644 | 5a5990cdb9f9e31692a2666fc578a8749c0d90b4
docs/en/quick-start.md | full | present | 100644 | 41eec741706da7f337954440afd3767e5b6bcc62
docs/ru/quick-start.md | full | present | 100644 | 03084b11e5ac65e5e8ecf3ed1c8c2e53f2b81477
docs/tasks/README.md | full | present | 100644 | 249cf99c60f8fb834185212b69c585c7dee2170b
docs/tasks/TASK-140-OWNER-HEADER-CLEANUP.md | task-record-v1 | present | 100644 | 6636ea6304b2f79abb9c678b068dd5d59baf3e1b
internal/configurationstudio/assets/control-language.js | full | present | 100644 | f0b03207b9a0af301bdf2883a9c5924cfe619328
internal/configurationstudio/assets/control.css | full | present | 100644 | 8ce068a9402758fe7dcb32e523cc87be1414a6fc
internal/configurationstudio/assets/dashboard.html | full | present | 100644 | 45114cc05008bc03a5e79171af80e44102b4e34e
internal/configurationstudio/assets/index.html | full | present | 100644 | 6921bf6b5067ad9a9da53c46c74f6db6387514c1
spec/current-state.md | full | present | 100644 | dd137201382a7d5dd0f350fc210da9a9fa2f5c5b
```

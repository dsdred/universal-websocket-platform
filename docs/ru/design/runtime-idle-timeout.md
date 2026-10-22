# Применение Runtime IdleSeconds — TASK-106

[English version](../../en/design/runtime-idle-timeout.md)

Design status: Architecture Confirmation PASS (independent Architect).
Implementation status: implemented candidate; verification и acceptance
определяются Recovery Evidence Envelope TASK-106. Commit/publication не разрешены.
Focused note не изменяет Approved DP-003/004 или Frozen ARCH-002.
Effective timeout behavior находится вне freeze ARCH-002.

## Конфигурация и владение

Runtime проверяет Published IdleSeconds в диапазоне 0..86400 до ресурсов; ноль
отключает policy. Composition передаёт immutable duration в Dispatcher,
который назначает его подготовленной Session до dormant launch и Commit.
Существующие constructors сохраняют отключённый idle. Новый explicit
NewTransactionalDispatcherWithIdleTimeout принимает read, write и idle duration,
сохраняя сигнатуры существующих constructors.

## Idle interval и активность

Интервал начинается, когда единственный допущенный Run создаёт private idle
scope после Commit и Start. До Run timer и watchdog не работают. Успешный полный
text/binary application Read или успешный application Write может обновить
deadline. Queue admission, partial fragments, ping/pong, failed I/O и завершение
Handler не обновляют его. Время Handler учитывается в idle.

Expiry и activity сериализуются одним private mutex. Только activity строго до
текущего deadline может обновить его; на deadline или позже побеждает expiry,
даже если watchdog ещё не получил scheduling. Expiry необратим: late I/O success
не возобновляет scope. Успешный полный Read может быть обработан; Handler
наблюдает отменённый idle scope. Transport success не является delivery guarantee.

## Contexts и остальные timeouts

При включённом idle Read и Handler используют private Run-scoped child исходного
Run context. Per-message ReadSeconds child остаётся lexical и отменяется до
Handler; Handler никогда не получает read child. При отключённом idle сохраняется
identity исходного Handler context. Это additive idle clarification focused
ReadSeconds note, без изменения Owner или Manager contract.

ReadSeconds сохраняет fixed full-message budget: успешные Writes обновляют idle,
но не продлевают выполняющийся read budget. WriteSeconds сохраняет caller
per-Send budget для queue и network. Idle не получает новой cancellation authority
над независимо допущенными Send: Owner cleanup закрывает transport; допущенные
Writes могут завершиться после expiry, не возобновляя scope. Исходный Run/root
cancellation сохраняется; idle никогда не отменяет root context. Run раскрывает
idle expiry через errors.Is context.DeadlineExceeded; при отмене исходный parent
error имеет приоритет. Transport/Handler errors сохраняют wrapping и при
необходимости join с idle cause. Existing Owner first-writer
termination categories и cleanup order сохраняются. Нет retry или новой category.

## Watchdog lifetime и shutdown

Session владеет одним private watchdog goroutine и timer на enabled Run. Он
содержит только bounded deadline/expired/closed state, private child cancellation
и coalesced wake notifications. Он никогда не вызывает Stop, не закрывает
transport, не меняет Manager и не записывает Owner termination intent. Network
I/O и Handler не выполняются под watchdog или Session lifecycle lock.

Любой return или panic Run закрывает и join-ит watchdog до публикации readLoopDone.
Stop закрывает и join-ит тот же idempotent watchdog до transport shutdown вне
lifecycle mutex; Run также сохраняет обязанность join. Concurrent activity не открывает closed tracking.
Parent cancellation также будит watchdog. Owner Cleanup остаётся единственным
terminal cleanup path, поэтому успешный Manager Wait покрывает отсутствие idle
work. Uncooperative Handler/transport может задержать Run/cleanup; hard-kill
guarantee отсутствует.

## Acceptance и verification

Доказать disabled identity, bounds 0/1/86400 и invalid guards; immutable
Snapshot-to-Dispatcher-to-Session propagation до Commit; отсутствие activity до
Run; expiry и успешные Read/Write resets; отсутствие reset от partial/control/
failed activity; Handler time/context isolation; независимость read/write budgets;
expiry/activity/Stop/root races; timer join при return/panic, единственный Owner
cleanup и Manager convergence. Сохранить text/binary/echo regressions. Выполнить
focused, real WebSocket, full regression, stress, build/vet/format/tidy/module
checks и race при поддержке existing tooling. Недоступный race — явная limitation.

Non-goals: TLS, новые Configuration fields/dependencies/lifecycle states/terminal
categories, native toolchain, Architecture B, PoC и Approved/Frozen amendments.

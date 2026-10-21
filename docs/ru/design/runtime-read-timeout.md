# Применение Runtime ReadSeconds — TASK-104

[English version](../../en/design/runtime-read-timeout.md)

Design status: Architecture Confirmation PASS (independent Architect).
Implementation status: implemented candidate. Commit/publication не разрешены.
Эта focused implementation note не меняет Approved DP-003/004/008, Frozen ARCH-002
или Active ARCH-003. Effective Listener timeout behavior исключён из freeze
[ARCH-002](../architecture/ARCH-002-runtime-foundation-freeze.md); существующий
[DP-008](DP-008-snapshot-builder-contract.md) задаёт ReadSeconds 0..86400.

## Конфигурация и ownership

Runtime проверяет ReadSeconds 0..86400 до приобретения ресурсов через private
validateReadSeconds guard, вызываемый validateExecutableSnapshot. Ноль отключает
configured read timeout. Legacy Session/Dispatcher constructors сохраняют нулевую
configured read duration; existing write-only constructor сохраняет семантику.
Добавляется NewTransactionalDispatcherWithTimeouts с аргументами duration сначала
read, затем write: read 0..86400 seconds, write positive и не более 300 seconds.
Production composition использует combined constructor. Dispatcher копирует обе
private durations в prepared Session до dormant goroutine и Commit; setter и
изменения Core, Owner или Manager не вводятся.

## Deadline каждого сообщения

Непосредственно перед каждым connection.Read Run создаёт child context.WithTimeout
от исходного Run context только при положительной configured read duration.
Ноль передаёт исходный context без изменения. Каждый positive timeout даёт один
фиксированный бюджет ожидания, всех fragments и чтения полного body, включая
обработку control frames до возврата полного application message. Fragment или
ping/pong activity не сбрасывает бюджет. Более ранний caller deadline/cancellation
сохраняется. Это не idle watchdog и не per-frame timeout.

Lexical read helper через defer отменяет child при successful, failed и panicking
Read. Failed Read классифицируется по child context до deferred cancellation;
созданный cleanup context.Canceled не должен маскировать normal close или transport
error. Raw failure отличается от successful complete-message result: normal close,
отображённый в nil, должен вернуть Run nil, а не создать сообщение или продолжить
чтение. Successful Read обрабатывает возвращённое полное сообщение даже при гонке
expiry с возвратом; guaranteed scheduling-time cutoff, retry, delivery или rollback
не обещаются.

Child cleanup завершается до Message processing и Handler invocation. Handler
получает исходный Run context, никогда не отменённый read child. Время Handler не
расходует следующий read budget, который начинается только перед следующим Read.
Read panic передаётся существующей Owner panic boundary после timer cleanup.

## Ошибки и lifecycle

Configured expiry остаётся распознаваемым через errors.Is DeadlineExceeded;
original cancellation сохраняет errors.Is semantics; unexpected transport errors
сохраняют existing wrapping. Normal/GoingAway close успешно завершает Run.
Read expiry — обычный Run failure; Session не вызывает свой Stop и не вводит
terminal category. Existing Owner first-writer causal Stop/root-cancellation
versus ExecutionFailure rules, cleanup, completion, observation, drain и lifetime
release не меняются. Stop не приобретает новый lock и не ждёт read timers.
Дополнительная goroutine или cancellation authority не вводится.

Existing coder/websocket Read convenience читает полный body. Context expiry
может закрыть transport; продолжение Read после expiry не поддерживается.
Existing WriteSeconds behavior, transport ownership и text/binary handling
сохраняются в действующих contracts.

## Acceptance и verification

- Принимаются 0/1/86400; 86401/max uint32 отвергаются до ресурсов. Direct Runtime
  guard tests проверяют errors.Is ErrInvalidRuntimeConfiguration отдельно от
  Builder rejection. Guard-only tests не доказывают полный Snapshot path;
  production composition/pre-Commit propagation дают отдельный wiring proof.
- Blocked Read и partial fragmented body истекают в одном бюджете; control frames
  не сбрасывают его. Каждый complete-message read получает новый бюджет.
- Доказываются earlier caller deadline/cancellation, disabled zero,
  success/error/panic child cleanup, original Handler context и excluded Handler time.
- Normal close, transport errors, concurrent Run/Stop/root cancellation и single
  Owner cleanup/Manager convergence сохраняют existing semantics.
- Проходят text/binary/echo и WriteSeconds regression. Focused tests, real
  WebSocket proof, full regression, concurrency stress, vet и formatting выполняются
  offline. Race — только при доступном tooling; отсутствие является явным
  limitation, stress не доказывает race freedom. Independent review обязателен.

Non-goals: IdleSeconds, TLS, новая configuration или lifecycle categories, native
toolchain, Architecture B, PoC и Approved/Frozen contract amendments.

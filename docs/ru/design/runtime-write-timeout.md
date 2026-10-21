# Runtime WriteSeconds enforcement — TASK-103

Design status: Architecture Confirmation PASS (независимый Architect).
Implementation status: Delivered via PR #103, source
`a2742ab927f7a23fd10a92db986ce36ff6a987e8`, merge
`aed7cc8f1693cf8bb48d71532fe543f09d2d1191`.
Focused implementation note не изменяет Approved DP-003/004, Frozen
ARCH-002 и Active ARCH-003. Effective Listener timeout behavior исключён из freeze ARCH-002.

Runtime проверяет Published Snapshot WriteSeconds в диапазоне 1..300 до получения
ресурсов. Composition передаёт положительный duration transactional Dispatcher;
Session получает неизменяемый duration до Commit. Manager, Owner, handoff,
владение cancellation и terminal categories не меняются. Legacy конструкторы
Session/Dispatcher сохраняют отсутствие configured write timeout.

Каждый Send при входе создаёт один context с более ранним из configured timeout
и caller deadline. Caller cancellation сохраняется. Context-aware gate
сериализует записи; ожидание очереди и network Write используют общий context и
budget. После получения gate Send проверяет context error, затем Running под
lifecycle lock; во время network I/O lock не удерживается. Копирование payload и
text/binary behavior сохраняются. Wrapping сохраняет errors.Is.

Queue timeout/cancellation не вызывает network Write и не останавливает Session.
Network failure следует существующему пути Send/handler/Run/Owner. Нет automatic
retry, root cancellation или новой termination category. Stop не получает write
gate. Уже допущенный Write может завершиться после concurrent Stop; policy не
может принудительно завершить transport, игнорирующий context. Success может
совпасть с expiry и означает transport success, без гарантий delivery/rollback.

Acceptance: 1/300 приняты, 0/301/overflow отвергнуты до ресурсов; production
Snapshot propagation; network deadline; более ранний caller deadline/cancellation;
отмена queued Send без Write; общий budget queue/I/O; сериализация Send;
независимый concurrent Stop; сохранение payload copy, text/binary, lifecycle,
transaction cleanup и error sentinels. Focused/regression/stress проверки;
недоступность race tooling фиксируется явно.

Non-goals: ReadSeconds, IdleSeconds, TLS, новая configuration, native toolchain,
Architecture B, PoC и изменения contracts.

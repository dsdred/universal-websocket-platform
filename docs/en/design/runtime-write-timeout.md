# Runtime WriteSeconds enforcement — TASK-103

Design status: Architecture Confirmation PASS (independent Architect).
Implementation status: implemented candidate; commit/publication not authorized.
This focused implementation note does not amend Approved DP-003/004, Frozen
ARCH-002 or Active ARCH-003. Effective Listener timeout behavior is outside ARCH-002 freeze.

Runtime validates Published Snapshot WriteSeconds in 1..300 before acquiring
resources. Composition passes a positive duration to the transactional Dispatcher;
each Session receives its immutable duration before Commit. Manager, Owner,
handoff, cancellation ownership and terminal categories remain unchanged. Legacy
Session/Dispatcher constructors retain no configured write timeout.

Each Send creates one context at entry with the earlier of the configured timeout
and caller deadline. Caller cancellation remains effective. A context-aware gate
serializes writes; queue waiting and network Write share that context and budget.
After gate acquisition Send checks context error, then Running state under the
lifecycle lock; no lifecycle lock is held during network I/O. Payload copying and
text/binary behavior remain unchanged. Errors retain errors.Is through wrapping.

Queue timeout/cancellation performs no network Write and does not Stop Session.
Network failure follows existing Send/handler/Run/Owner propagation. No automatic
retry, root cancellation or new termination category is introduced. Stop never
acquires the write gate. An admitted Write can finish after concurrent Stop;
context-ignoring transports cannot be forcibly terminated by this policy. Success
may race expiry and means transport success, not delivery or rollback guarantees.

Acceptance: bounds 1/300 accepted and 0/301/overflow rejected before resources;
production Snapshot propagation; configured network deadline; earlier caller
deadline/cancellation; cancellable queued Send without Write; one budget across
queue and I/O; serialized concurrent Send; independent concurrent Stop; unchanged
payload copy, text/binary, lifecycle, transaction cleanup and error sentinels.
Run focused/regression/stress checks; record unavailable race tooling explicitly.

Non-goals: ReadSeconds, IdleSeconds, TLS, new configuration, native toolchain,
Architecture B, PoC and contract amendments.

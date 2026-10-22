# Runtime IdleSeconds enforcement — TASK-106

[Russian version](../../ru/design/runtime-idle-timeout.md)

Design status: Architecture Confirmation PASS (independent Architect).
Implementation status: published by TASK-106, source
`909abc7f67a36f6709ba343d0cf5b63bc3509c85`, PR #106, merge
`9151db9e32e0b351d403087e4f1361b31b134fac`. Historical verification and
acceptance remain in the unchanged TASK-106 Recovery Evidence Envelope.
This focused note does not amend Approved DP-003/004 or Frozen ARCH-002.
Effective timeout behavior is outside the ARCH-002 freeze.

## Configuration and ownership

Runtime validates Published IdleSeconds in 0..86400 before resources; zero
disables the policy. Composition passes the immutable duration to Dispatcher,
which assigns it to the prepared Session before dormant launch and Commit.
Existing constructors retain disabled idle behavior. A new explicit
NewTransactionalDispatcherWithIdleTimeout constructor accepts read, write and
idle durations and preserves the existing constructor signatures.

## Idle interval and activity

The interval begins when the sole admitted Run establishes its private idle
scope, after Commit and Start. No timer or watchdog runs before Run. Each
successful complete text/binary application Read or successful application Write
may reset the deadline. Queue admission, partial fragments, ping/pong, failed
I/O and Handler completion do not reset it. Handler time counts toward idle.

Expiry and activity serialize under one private mutex. Activity strictly before
the current deadline can reset it; at or after the deadline expiry wins even
if the watchdog has not yet been scheduled. Expiry is irreversible: late I/O
success cannot revive the scope. A successful complete Read may still be processed;
Handler then observes the cancelled idle scope. Transport success is not a delivery guarantee.

## Contexts and other timeouts

With idle enabled, Read and Handler use a private Run-scoped child of the
original Run context. The per-message ReadSeconds child is still lexical and
cancelled before Handler; Handler never receives that read child. With idle
disabled the original Handler context identity is preserved. This is an additive
idle policy clarification to the focused ReadSeconds note, not an Owner or
Manager contract change.

ReadSeconds retains its fixed full-message budget: successful Writes reset idle
but never extend an in-progress read budget. WriteSeconds retains the caller's
per-Send queue plus network budget. Idle does not grant new cancellation authority
over independently admitted Send calls; Owner cleanup closes the transport, and
already admitted Writes may complete after expiry without reviving the scope.
Original Run/root cancellation remains effective and root context is never
cancelled by idle. Idle expiry is exposed by Run through errors.Is
context.DeadlineExceeded; original parent error takes precedence when cancelled.
Transport/Handler errors retain wrapping, joined with idle cause when necessary.
Existing Owner first-writer termination categories and
cleanup order remain unchanged. No retry or new terminal category is introduced.

## Watchdog lifetime and shutdown

Session owns one private watchdog goroutine and timer per enabled Run. It
contains only bounded deadline/expired/closed state, private child cancellation
and coalesced wake notifications. It never invokes Stop, closes transport,
mutates Manager or writes Owner termination intent. Network I/O and Handler work
never execute under the watchdog or Session lifecycle lock.

Every Run return or panic closes and joins the watchdog before readLoopDone is
published. Stop closes and joins the same idempotent watchdog before transport shutdown,
outside the lifecycle mutex; Run also retains its join obligation. Concurrent
activity cannot reopen closed tracking. Parent cancellation also wakes the
watchdog. Owner Cleanup remains the sole terminal cleanup path, so successful
Manager Wait covers absence of idle-owned work. An uncooperative Handler or
transport can delay Run/cleanup; there is no hard-kill guarantee.

## Acceptance and verification

Prove disabled identity, bounds 0/1/86400 and invalid guards; immutable
Snapshot-to-Dispatcher-to-Session propagation before Commit; no activity until
Run; expiry and successful Read/Write resets; rejected partial/control/failed
activity; Handler time and context isolation; read/write budget independence;
expiry/activity/Stop/root races; return/panic timer join, one Owner cleanup and
Manager convergence. Retain text/binary/echo regressions. Run focused, real
WebSocket, full regression, stress, build/vet/format/tidy/module checks and race
where existing tooling supports it. Unavailable race is an explicit limitation.

Non-goals: TLS, new Configuration fields/dependencies/lifecycle states/terminal
categories, native toolchain, Architecture B, PoC and Approved/Frozen amendments.

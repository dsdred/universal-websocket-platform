# Runtime ReadSeconds enforcement — TASK-104

[Russian version](../../ru/design/runtime-read-timeout.md)

Design status: Architecture Confirmation PASS (independent Architect).
Implementation status: published by TASK-104, source
`63d99906924195115cb9dc260762ca050cb675b8`, PR #104, merge
`fb916df77e697fb7a45dc7bb6e8d0150411f2d82`. Historical verification and acceptance
remain in the unchanged TASK-104 evidence. Gate 9 technical closure and retained
limitations: [closure record](../reviews/gate-9-closure.md).
This focused implementation note does not amend Approved DP-003/004, Draft DP-008, Frozen
ARCH-002 or Active ARCH-003. Effective Listener timeout behavior is outside
[ARCH-002](../architecture/ARCH-002-runtime-foundation-freeze.md) freeze; existing
[DP-008](DP-008-snapshot-builder-contract.md) ReadSeconds range is 0..86400.

## Configuration and ownership

Runtime validates ReadSeconds in 0..86400 before acquiring resources, through a
private validateReadSeconds guard called by validateExecutableSnapshot. Zero
disables the configured read timeout. Legacy Session/Dispatcher constructors
retain zero configured read duration; the existing write-only constructor keeps
its semantics. Add NewTransactionalDispatcherWithTimeouts with read then write
duration arguments: read 0..86400 seconds, write positive and at most 300 seconds.
Production composition uses this combined constructor. Dispatcher copies both
private durations into the prepared Session before the dormant goroutine and
Commit; no setter or changes to Core, Owner or Manager are introduced.

## Per-message deadline

Immediately before each connection.Read, Run creates a child context.WithTimeout
from the original Run context only when the configured read duration is positive.
Zero passes the original context unchanged. Each positive timeout provides one
fixed budget for waiting, every fragment and complete body reading, including
control-frame handling until a full application message returns. Fragment or
ping/pong activity does not reset it. The earlier caller deadline/cancellation
remains effective. This is not an idle watchdog or a per-frame timeout.

A lexical read helper defers child cancellation for successful, failed and
panicking Read. It classifies a failed Read against the child context before
deferred cancellation; cleanup-created context.Canceled must not mask a normal
close or transport error. Raw failure and successful complete-message results
remain distinguishable: mapping normal close to nil must return Run nil rather
than invent a message or continue reading. A successful Read processes the
returned complete message even if expiry races its return; there is no guaranteed
scheduling-time cutoff, retry, delivery or rollback promise.

Child cleanup finishes before Message processing and Handler invocation. Handler
receives the original Run context, never the cancelled read child. Handler time
does not consume the next read budget, which starts only before the next Read.
Read panic propagates to the existing Owner panic boundary after timer cleanup.

## Errors and lifecycle

Configured expiry remains discoverable through errors.Is DeadlineExceeded;
original cancellation retains errors.Is semantics; unexpected transport errors
retain existing wrapping. Normal/GoingAway close terminates Run successfully.
Read expiry is an ordinary Run failure; Session does not Stop itself or introduce
a terminal category. Existing Owner first-writer causal Stop/root-cancellation
versus ExecutionFailure rules, cleanup, completion, observation, drain and lifetime
release remain unchanged. Stop acquires no new lock and does not wait for read
timers. No extra goroutine or cancellation authority is introduced.

The existing coder/websocket Read convenience reads the complete body. Context
expiry can close transport; continuing Read after expiry is not supported.
Existing WriteSeconds behavior, transport ownership and text/binary handling stay
within their current contracts.

## Acceptance and verification

- Bounds 0/1/86400 accepted; 86401/max uint32 rejected before resources. Direct
  Runtime guard tests verify errors.Is ErrInvalidRuntimeConfiguration separately
  from Builder rejection. Guard-only tests do not prove the full Snapshot path;
  production composition/pre-Commit propagation supplies separate wiring proof.
- Blocked Read and partial fragmented body expire within one budget; control
  frames do not reset it. Each complete-message read receives a fresh budget.
- Caller earlier deadline/cancellation, disabled zero, success/error/panic child
  cleanup, original Handler context and excluded Handler time are demonstrated.
- Normal close, transport errors, concurrent Run/Stop/root cancellation and
  single Owner cleanup/Manager convergence preserve their existing semantics.
- Text/binary/echo and WriteSeconds regression pass. Focused tests, real
  WebSocket proof, full regression, concurrency stress, vet and formatting run
  offline. Race runs only with available tooling; absence is an explicit
  limitation and stress is not race-freedom proof. Independent review is required.

Non-goals: IdleSeconds, TLS, new configuration or lifecycle categories, native
toolchain, Architecture B, PoC and Approved/Frozen contract amendments.

## Additive IdleSeconds policy qualification (TASK-106)

When IdleSeconds is enabled, the original Run context is wrapped in a private
Run-scoped idle child used by Read and Handler. Handler never receives the
per-read child. Zero idle preserves original context identity. ReadSeconds still
excludes Handler time; idle includes it. See [idle design](runtime-idle-timeout.md).

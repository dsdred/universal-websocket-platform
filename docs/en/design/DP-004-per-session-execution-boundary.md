# DP-004: Per-Session Execution Boundary

[Russian version](../../ru/design/DP-004-per-session-execution-boundary.md)

## 1. Status

**Status:** Approved

This approved design defines the per-Session execution boundary after successful WebSocket Upgrade. It separates synchronous activation, transport ownership, asynchronous Session execution, terminal cleanup, and Runtime shutdown accounting without changing the frozen Runtime Host lifecycle.

## 2. Problem Statement

The current Session Dispatcher constructs a Session with an already-owned WebSocket and synchronously performs `Start`, `Run`, and `Stop` in the HTTP handler goroutine. That model cannot support Runtime-owned asynchronous execution:

- the handler remains blocked for the entire Session lifetime;
- Runtime has no stable per-Session Stop-request capability;
- Session Manager cannot account for execution through terminal observation;
- connection-context cancellation belongs to a Dispatcher defer that disappears when Dispatcher becomes asynchronous;
- registration completion can precede the end of Runtime-owned work.

[DP-003](DP-003-runtime-session-manager.md) defines registration identity and accounting. This document defines the execution boundary and its narrow integration with those contracts.

## 3. Goals

- Preserve exactly one WebSocket owner and one connection-cancellation owner.
- Define one synchronous activation executor.
- Define one per-Session Execution Owner and one owned goroutine.
- Make Start, Run, termination, cleanup, and completion linearizable.
- Give Runtime shutdown a stable non-owning Stop-request capability.
- Keep Session Manager free from Session and transport behavior.
- Make successful Manager Wait cover the full tracked owner lifetime.

## 4. Scope

This document designs only:

- ownership after WebSocket Upgrade;
- transport-independent provisional construction;
- package-private transport attachment;
- Execution Owner creation, launch, activation, and lifecycle;
- Session Start, Run, and terminal cleanup;
- connection-context cancellation acknowledgement;
- completion, terminal observation, Stop requests, Snapshot capability integration, and owner-lifetime accounting.

It does not design Router, Delivery, Persistence, Presence, Groups, Diagnostics infrastructure, Session limits, Cluster behavior, Configuration fields, public management API, or a generic supervision framework.

## 5. Existing Constraints

The design preserves:

- [ARCH-001](../architecture/ARCH-001-runtime-architectural-pattern.md) ownership and explicit-execution principles;
- the frozen Host lifecycle, Runtime context, Admission Gate, startup, rollback, readiness, and Listener ownership in [ARCH-002](../architecture/ARCH-002-runtime-foundation-freeze.md);
- the pre-Upgrade Authentication and Upgrade-boundary ownership in [DP-001](DP-001-runtime-handshake-pipeline.md);
- the composition direction in [DP-002](DP-002-runtime-host-composition-root.md);
- the read-only Runtime capability boundary in [ADR-0004](../adr/0004-handshake-runtime-dependencies.md);
- DP-003 Reserve, Commit, Complete, Lookup, registration accounting, and identity Snapshot semantics.

The current Session implementation requires migration because its constructor requires a WebSocket and its `Created` state always has transport. That implementation detail is not treated as the target provisional contract.

## 6. Selected Architecture

Dispatcher is the only handoff executor:

```text
Upgrade boundary enters Dispatcher
    -> Reserve SessionID
    -> create transport-independent Session Core
    -> construct Execution Owner
    -> create exactly one dormant launch goroutine blocked on its one-shot Commit gate
    -> provisionally attach transport and cancellation
    -> activate owner in PreCommit without starting Session execution
    -> Commit
       -> under one Manager synchronization boundary:
          -> publish Registration, lease accounting, and Stop capability
          -> transfer Session, WebSocket, cancellation, and execution ownership
          -> transition the one-shot Commit gate from PreCommit to Committed
       -> release the synchronization boundary only after all publication is complete
    -> return accepted=true immediately

Dormant launch goroutine
    -> cannot pass the Commit gate before successful Commit
    -> becomes the owned execution path when the gate reaches Committed
    -> install Runtime-cancellation observation with a race-safe post-registration check
    -> Start
    -> Run when permitted
    -> Terminalizing
    -> Session Cleanup
    -> Complete
    -> Terminal Observer
    -> close control-call admission
    -> UnregisterAndDrain Runtime-cancellation callback and entered control calls
       -> if callback absence is not confirmed: remain Terminalizing with lease active
    -> seal detached Stop control cell
    -> Terminal
    -> release Owner Lifetime Lease
    -> return with no Runtime-owned epilogue
```

Commit is the sole irreversible publication point for both Registration and execution eligibility. Every operation before Commit is provisional and may fail through Abort plus Dispatcher cleanup. No recoverable operation after Commit can return `accepted=false` or restore pre-Commit state.

“Owner started” has one normative meaning: the one-shot Commit gate reached `Committed`, so the already existing dormant launch goroutine is eligible to enter owner lifecycle. Scheduler execution, `Session.Start`, and `Session.Run` are later effects and are not publication points.

## 7. Ownership Domains

### Upgrade Boundary

Ownership begins after successful `websocket.Accept`. Dispatcher, acting for the Upgrade boundary, exclusively owns WebSocket, connection cancellation, and pre-handoff cleanup.

Ownership ends only when Commit succeeds. Commit is the single transport-ownership and registration-publication linearization point. Before it, Dispatcher is the sole closer and cancellation owner. After it, Dispatcher never closes or cancels those resources.

### Session Core and Session

Session Core owns only immutable identity, Principal, metadata, and Message Handler. It is not a Session, exposes no `Start`, `Run`, `Send`, or `Stop`, and owns no transport.

Before Commit, one Core may be provisionally formed into a fully configured Session and bound to a dormant owner and one dormant launch goroutine, but Dispatcher still owns transport, the provisional launch obligation, and cleanup. The goroutine can observe only its private Commit gate and cannot call Session lifecycle. Successful Commit publishes the committed identity, transfers the prepared Session, WebSocket, connection cancellation, and execution responsibility, and makes that same goroutine eligible. No Runtime-owned Session execution exists before Commit.

### Reservation and Registration

Dispatcher owns the Reservation obligation through Commit or Abort. Before Commit, Registration does not exist, Lookup cannot observe it, Snapshot cannot contain it, and owner-lifetime accounting does not include it.

Commit consumes Reservation and irreversibly publishes committed identity, owner-only completion, owner lifetime, Snapshot Stop capability, and ownership acceptance. After Commit, Registration may disappear only through normal Complete.

### Execution Owner

Owner construction and pre-Commit activation acquire no Manager or transport ownership. Owner remains dormant in `PreCommit`; Session Start and Run are forbidden. Dispatcher creates exactly one launch goroutine and exactly one one-shot Commit gate for that owner. Successful Commit gives owner the attached Session, committed bundle, and execution responsibility by changing that gate to `Committed` within the same publication boundary. From Commit until Terminal, that single path is the sole production caller of Session `Start`, `Run`, and terminal cleanup.

### Session Manager

Manager owns reservation accounting, registration identity and visibility, Snapshot publication, completion mutation, and Owner Lifetime Lease accounting. It owns no Session, WebSocket, Context, goroutine, or Terminal Result.

### Terminal Result

Owner creates one immutable categorized Terminal Result and lends it synchronously to one observer call. Observer owns no resource or lifecycle capability.

## 8. Connection Cancellation and Cleanup Acknowledgement

Activation wraps the derived connection context in one narrow, idempotent cancellation cell owned with the WebSocket. The cell uses the actual Runtime-derived context and its private cancellation function; it exposes neither value to owner or observer.

Cancellation terminology is:

- **invocation:** an attempt to invoke the cancellation cell; repeated invocation is permitted;
- **effective cancellation:** the single state transition after which the derived context is observably canceled;
- **acknowledgement:** immutable cleanup output confirming the effective canceled state.

The first effective cancellation changes context state once. Repeated invocations are safe no-ops. The terminal requirement is effective canceled state, not exactly one invocation.

The attached Session exposes one owner-facing synchronous terminal operation:

```text
Cleanup(cleanupContext) CleanupResult
```

`CleanupResult` is immutable and contains categorized transport-close outcome, effective-cancellation acknowledgement, and cleanup-panic category. It contains no raw transport, Context, callback, or arbitrary error.

Session Cleanup is itself the panic-safe boundary. Internal close, join, and cancellation steps may panic, but the Cleanup wrapper recovers each internal panic and always returns a `CleanupResult`. Outward panic from the Cleanup contract is forbidden.

Session Cleanup:

1. performs the existing Stop/transport-close work;
2. always invokes the cancellation cell through an internal final guard;
3. observes the derived context in canceled state;
4. returns acknowledgement only after that observation;
5. records sanitized categories for every recovered internal anomaly.

Repeated Cleanup returns the same detached terminal result and performs no second effective close or cancellation. Cleanup never exposes partial mutable state.

The cancellation cell is constructed only from the standard Runtime-derived cancellation primitive. Its safe invocation isolates internal cancellation panic, performs the private cancellation operation, and confirms the canceled state. A failure to confirm cancellation is an architecture anomaly in `CleanupResult`; owner does not release its lease until Cleanup returns. Owner proves canceled state solely from the returned acknowledgement and never accesses the private cancellation cell.

## 9. Provisional Session Model

The selected model is **transport-independent Session Core followed by provisional Session formation before Commit**.

Core is not a lifecycle object. Therefore:

- Start, Run, Send, Stop, and Cleanup before provisional formation are impossible by type boundary;
- there is no transport-free `Created` Session;
- no operation can dereference nil transport;
- disposal of an unattached Core requires no lifecycle cleanup.

The package-private preparation operation is conceptually:

```text
prepareOwner(
    core,
    WebSocket,
    ConnectionCancellation,
) (PreCommitOwner, error)
```

Only Dispatcher handoff code can call it. Production composition permits exactly one caller and forbids concurrent invocation. The operation is not a public attachment or ownership-transfer API.

Its contract:

- one Core can be prepared once;
- Session formation, owner control-cell installation, and Stop binding preparation complete before Commit;
- Dispatcher creates exactly one dormant launch goroutine and binds it to exactly one one-shot Commit gate;
- WebSocket and cancellation remain owned by Dispatcher throughout preparation;
- the prepared owner remains in `PreCommit`, and the launch goroutine cannot pass its gate or perform Session lifecycle work;
- no `Start`, `Run`, or `Cleanup` can race before Commit;
- it performs bounded in-memory construction and no network I/O;
- validation failure or recovered panic leaves Registration absent and both supplied resources with Dispatcher;
- every terminal pre-Commit outcome resolves the gate as non-committed exactly once, waits for the dormant goroutine to return, disposes owner-local prepared values, and then performs Abort before transport cleanup;
- no Runtime-cancellation callback registration exists before Commit;
- no lease or Snapshot capability exists before Commit;
- Start, Run, Stop, and Cleanup can begin only after successful Commit.

The prepared Session has the values required by the existing transport-owning `Created` state, but ownership and lifecycle authority remain with Dispatcher until Commit. At successful Commit, the Session becomes Runtime-owned in `Created`; existing `Created -> Running -> Stopping -> Stopped` semantics then apply.

## 10. Creator and Composition Boundary

Runtime Host remains the production composition root. It constructs:

- Session Manager;
- a narrow activation/registration adapter;
- Execution Owner factory;
- Terminal Observer;
- Dispatcher.

Execution Owner receives only:

- owner control cell, Runtime-derived execution context, and read-only root Runtime context observation input;
- lifecycle-only Session after activation;
- owner-scoped Completion Adapter;
- owner-scoped Lifetime Lease;
- one Terminal Observer;
- immutable RegistrationID.

Owner does not import concrete Runtime Host, Manager, Listener, Handshake, HTTP, WebSocket library, Runtime Snapshot, Router, Delivery, Presence, or Persistence.

No service locator, global registry, reflection, `context.Value` dependency injection, generic executor pool, or generic supervisor is introduced.

## 11. Owner Preparation and Start Boundary

Owner construction is synchronous. Dispatcher still owns Reservation, WebSocket, cancellation, and cleanup.

Pre-Commit activation prepares:

1. the fully formed but not yet Runtime-owned Session;
2. the owner control cell and causal termination state;
3. the Stop-publication binding required by Commit;
4. the immutable read-only root Runtime context observation input required by the future owned execution path;
5. exactly one dormant launch goroutine and its private one-shot Commit gate.

The owner is then in `PreCommit`. The launch goroutine exists, but it is still provisional Dispatcher-owned control flow and cannot pass the Commit gate. It does not call Session Start, Run, Stop, Cleanup, Complete, observer, or lease release.

Dispatcher owns the entire pre-Commit transaction through one panic-safe orchestration boundary beginning no later than dormant-path creation. Ordinary error, context cancellation, a Commit lost to BeginShutdown, attachment or preparation panic, and panic immediately before Commit all use one terminal sequence:

1. publish the non-committed gate outcome exactly once;
2. wait for the dormant path to observe that outcome and return;
3. dispose all owner-local prepared values;
4. execute Abort;
5. retain transport and connection-cancellation ownership;
6. return `accepted=false` with a safe error so Handshake performs transport cleanup.

Abort is not a gate-completion mechanism. Dispatcher never returns a pre-Commit outcome before the dormant path has returned. The dormant path never invokes committed Complete, observer, Runtime-cancellation observation, or lease operations. No callback exists to unregister or join on a failed Commit. The orchestration boundary converts recoverable panic to the same safe error result; it does not propagate that panic through the transport contract.

Successful Commit is the owner start boundary and the only execution publication point. Within the same Manager synchronization boundary, it changes the gate `PreCommit -> Committed`, transfers ownership, publishes the owner-scoped bundle, and makes exactly one already existing path eligible. The operation does not return success and does not release the synchronization boundary before all of those effects are complete.

There is no second launch call, launch acknowledgement, or fallible post-Commit activation step. The goroutine becoming scheduled is not part of Commit, and neither are Session Start, Run, Cleanup, Complete, observer, or lease release.

The Commit gate is not another lifecycle, coordinator, supervisor, or owner. It is a narrow one-shot publication binding representing the existing `PreCommit -> Committed` transition. Dispatcher constructs it and Manager receives only permission to invoke its one publication operation. `Committed` makes the single dormant path eligible. The non-committed outcome terminates that path without execution. Repeated publication returns the already fixed outcome and never releases a second path. The binding provides no lifecycle control after publication.

## 12. Commit Bundle and Lease Identity

Commit is the only Manager and handoff publication point. It either fails without publishing committed state or irreversibly publishes the complete owner bundle:

```text
RegistrationID
CompletionAdapter
OwnerLifetimeLease
StopPublicationBinding
ExecutionPublicationBinding
```

All potentially fallible validation, construction, and allocation required by this bundle completes before the publication critical section. The critical section performs only bounded mutation of Manager-owned maps/accounting and the already prepared panic-free one-shot publication binding. It calls no external code, Session or owner method, callback, goroutine scheduler operation, or blocking channel send while holding the Manager lock.

The bundle has these properties:

- RegistrationID is never reused for Manager lifetime;
- Lease identity is bound one-to-one to RegistrationID and one owner control cell;
- exactly one lease exists for one committed Registration;
- Completion Adapter accepts no caller-supplied RegistrationID and can complete only its bound Registration;
- Lifetime Lease can release only its bound lease;
- Stop publication binds Snapshot capability to the same already prepared owner control cell;
- Execution publication binds exactly one dormant launch goroutine to the same RegistrationID and publishes eligibility once through a panic-free state mutation;
- failure to return the complete bundle means Commit did not occur;
- no partially committed or partially returned bundle exists.

The panic-contained integrated Commit operation consists only of:

1. validating that Reservation and Manager state still permit Commit;
2. consuming Reservation and publishing Registration identity;
3. publishing Completion Adapter, Owner Lifetime Lease, and Stop binding;
4. invoking the narrow one-shot execution-publication binding with `Committed`;
5. releasing the shared Manager synchronization boundary;
6. returning the complete bound result.

Steps 2–4 are not externally separable. Lookup and BeginShutdown use the same synchronization boundary and cannot observe Registration before the execution binding is committed. Manager holds only the narrow publication right; it does not own the binding's path or invoke Session lifecycle.

Commit has exactly two recoverable outcomes:

- **non-committed failure:** validation fails before the first observable mutation; Registration, lease, Stop binding, and committed execution outcome do not exist, and Dispatcher publishes the non-committed gate outcome;
- **successful Commit:** Registration and the complete bundle are published and the gate outcome is `Committed`; rollback is forbidden.

No recoverable panic exists between partial publication mutations. External or arbitrary code is not invoked there, and every prepared publication mutation is panic-free by contract. Process termination, unrecoverable Go runtime failure, memory corruption, and equivalent failures are accepted unrecoverable limitations rather than a recoverable Commit-panic branch.

Commit does not include goroutine scheduling, `Session.Start`, `Session.Run`, Cleanup, Complete, observer invocation, lease release, or Dispatcher return to Handshake.

Before Commit:

- Registration, lease accounting, and Snapshot capability do not exist;
- Lookup cannot observe the prepared owner;
- BeginShutdown cannot capture it;
- the dormant launch goroutine cannot enter owner lifecycle;
- Dispatcher retains transport and cancellation ownership.

After Commit:

- Registration is immediately visible to Lookup and shutdown accounting;
- BeginShutdown may capture the immutable identity and bound Stop capability;
- exactly one committed launch path is already eligible to enter owner lifecycle;
- Session, WebSocket, cancellation, and execution responsibility belong to owner;
- Registration is never rolled back and may be removed only by Complete;
- Dispatcher must return `accepted=true` and must not perform cleanup.

There is no post-Commit failure branch in Dispatcher. A panic, cancellation, Start failure, or any other recoverable event after Commit is a normal owner terminal cause and follows Cleanup, Complete, observer, and lease-release ordering.

An interval may exist between Commit return and actual scheduler execution of the launch goroutine. It is safe and is not an orphan interval: Registration, lease, Stop capability, ownership, and an already eligible execution path all exist. Shutdown may request Stop during that interval; the owner observes termination before Start and enters the normal terminal path.

Repeated Commit for the same still-existing Registration returns the same logical bound publication result: the same RegistrationID, Completion Adapter, Owner Lifetime Lease identity, Stop-publication binding, and committed execution outcome. It does not invoke publication again, cause another callback installation, create a lease, Stop capability, or goroutine, or change accounting. Commit after Complete retains the `ErrRegistrationRemoved` semantics defined by DP-003.

Lease release is exposed only through an owner-bound panic-safe adapter with an explicit outcome. Release by the owner is effective once. Repeated release is a stable anomaly result. Unknown, stale, or foreign release cannot affect another lease. Never-reused RegistrationID plus owner-bound lease identity prevents ABA.

Adapter panic or unsuccessful release leaves that lease accounting active. It cannot empty another lease or allow Wait to report success.

## 13. Dispatcher Handoff Contract

Dispatcher executes one serialized control flow:

1. Reserve SessionID;
2. create Session Core;
3. create Execution Owner in `PreCommit`;
4. create exactly one dormant launch goroutine blocked on the owner Commit gate;
5. provisionally attach WebSocket and connection cancellation;
6. activate all owner control structures without starting Session execution;
7. inspect termination and shutdown eligibility;
8. Commit the prepared owner binding.

Commit is the sole Registration publication, execution publication, ownership-transfer, and handoff-success linearization point. Its internal Manager mutation and Commit-gate transition are protected by the same synchronization boundary and are externally indivisible. There is no later activation mutation and no recoverable state “Commit succeeded but execution publication did not”.

Dispatcher guarantees exactly-once execution by creating one launch goroutine before Commit and supplying one one-shot gate in the Commit binding. Manager guarantees atomic visibility of the opaque binding with Registration, but does not create or own the goroutine. After Commit returns success, Dispatcher has no launch, cleanup, Session, or cancellation obligation and only returns `accepted=true`.

Observable `DispatchAuthenticated(...)(accepted, error)` semantics are:

- `accepted=false, error=nil` — no ownership transfer occurred; Handshake/Upgrade boundary retains and cleans transport;
- `accepted=false, error!=nil` — no ownership transfer occurred; Handshake/Upgrade boundary retains and cleans transport after recording the operational failure;
- `accepted=true, error=nil` — Commit succeeded, ownership transfer completed, and Dispatcher returns immediately while owner progresses independently.

Production Dispatcher does not return `accepted=true` with an error. Recoverable post-Commit execution outcomes fixed before Terminal Result construction are delivered through Terminal Observer. Callback outcomes arising after that construction belong only to the existing terminal-accounting path and never return through Dispatcher. Request cancellation before Commit causes Abort and `accepted=false`; cancellation after Commit is a normal owner termination source.

## 14. Execution Owner Lifecycle

The conceptual state machine is:

```text
PreCommit
    -> Committed

Committed
    -> Starting
    -> Terminalizing

Starting
    -> Running
    -> Terminalizing

Running
    -> Terminalizing

Terminalizing
    -> Terminal
```

`PreCommit` means the owner, Session values, control cell, Stop binding, read-only root Runtime context observation input, and one dormant launch goroutine are prepared while Dispatcher still owns Reservation, transport, and provisional launch control. No Runtime-cancellation callback exists. The goroutine is blocked on the Commit gate, so Session execution has not started.

`Committed` begins inside successful Commit before the shared synchronization boundary is released. Registration, lease, Stop capability, Session ownership, and execution eligibility are published together, while Start has not yet linearized. Actual scheduling may occur later without creating another lifecycle state.

`Starting` and `Running` are the explicit Start and Run linearization states.

`Terminalizing` is mandatory for every committed terminal path. It runs the panic-safe terminal chain and closes control-call admission at the defined post-observer step. No post-Commit path transitions directly to Terminal.

`Terminal` means Session execution and observer work are complete, `UnregisterAndDrain` confirmed that no future or entered callback work remains, control-call admission is closed, every entered control call returned, and the control cell is sealed. Only the conditional lease-release attempt and immediate technical return remain. An unconfirmed callback lifetime remains `Terminalizing` and never claims Terminal. Restart is forbidden.

The dormant goroutine belongs to `PreCommit`; it does not require another lifecycle state. A non-committed gate outcome terminates that path without transition to `Committed` and without Complete, committed observer, callback entry, or Manager owner-lifetime accounting. Every post-Commit terminal outcome enters `Terminalizing` before the single definition of Terminal.

## 15. Start and Run Linearization

After handoff and before Start linearization, owner installs cancellation observation on the Host-owned root Runtime context. This root context is the single normative observation source; the callback never observes the derived execution context. Callback creation and ownership belong exclusively to owner; Dispatcher and Manager neither create nor retain it. The callback receives only the narrow causal-cell operation and no Session, Manager, WebSocket, lifecycle control, or completion capability.

Installation uses one race-safe contract: register observation first, then synchronously check the root Runtime context, or use an equivalent primitive with the same guarantee. Root cancellation before or during registration is therefore either delivered by the context mechanism or observed by the post-registration check. Both paths attempt the same `RuntimeCanceled` first-writer mutation, so cancellation is effective at most once. Repeated callback invocation is safe.

If callback installation returns an error or panics, the owner wrapper records a sanitized callback-installation anomaly, attempts `ExecutionFailure` as termination intent, and enters `Terminalizing`. Registration and ownership remain committed; rollback is forbidden. If the root Runtime context is observed canceled, `RuntimeCanceled` is attempted instead. Callback cleanup still runs through the common terminal contract, including when installation created no registration.

After installation, owner checks unified termination state under its control lock.

- If termination already exists, it enters `Terminalizing`; Start and Run are skipped.
- Otherwise `Committed -> Starting` is the Start linearization point.

Lock is released before `Session.Start`. Start is called at most once. Error or panic enters `Terminalizing`; Run is not called.

After successful Start, owner checks the same termination state again:

- existing termination enters `Terminalizing`;
- otherwise `Starting -> Running` is the Run linearization point.

Lock is released before `Session.Run`.

If termination linearizes first, Run is forbidden. If `Running` linearizes first, Run is officially started even if cancellation occurs before the next machine instruction invokes it. Calling Run with an already-canceled execution context is then the valid running path.

No lifecycle lock is held during Start, Run, or Cleanup.

## 16. Unified Termination Intent

Every termination source competes through one owner control cell and one first-writer linearization point.

Termination sources are:

- `ExplicitStop`;
- `RuntimeCanceled`;
- `NaturalCompletion`;
- `ExecutionFailure`;
- `RecoveredPanic`.

The first source to write the empty causal cell becomes the primary termination source. `ExplicitStop`, `RuntimeCanceled`, `NaturalCompletion`, `ExecutionFailure`, and `RecoveredPanic` use the same mutation. Simultaneous attempts have exactly one winner under the control lock.

`RequestStop() bool` returns `true` only when that invocation first establishes termination intent. It returns `false` after earlier Runtime cancellation, earlier explicit Stop, Terminalizing, or Terminal.

Runtime-cancellation timing has one normative model. No callback registration exists before Commit. After the dormant owner observes `Committed`, it installs observation of the root Runtime context before Start linearization using the race-safe registration-and-check contract above. A root context canceled before Commit is rejected synchronously by Dispatcher and never creates a callback. A root context canceled after Commit but before installation is observed during installation and competes through the causal cell as `RuntimeCanceled`. Session Cleanup cancels only the derived execution context through its private cancellation cell; because the callback observes the distinct root Runtime context, normal Session Cleanup cannot itself generate `RuntimeCanceled`.

Owner owns callback registration from creation through cleanup. Callback invocation and explicit RequestStop share the same admission and outstanding control-call accounting. The callback wrapper is panic-safe: outward panic is forbidden, a panic becomes a sanitized callback anomaly, termination intent is still attempted, and entered-call accounting is decremented in a final guard. Callback performs no Session I/O and waits for neither Cleanup, Complete, observer, nor lease release.

Later signals are reduced to bounded secondary categories. They do not replace the first cause, change the original authentication result, affect an already returned `RequestStop` value, or create another Stop, Complete, observer, or lease obligation.

## 17. RequestStop Contract

`RequestStop`:

- is thread-safe;
- waits for no Session I/O;
- waits for neither Start, Run, Cleanup, Complete, observer, nor terminal completion;
- performs only bounded owner-state mutation plus local context cancellation;
- does not promise a hard time bound for standard context cancellation;
- returns `false` for repeated, concurrent losing, post-Terminalizing, and post-Terminal calls.

Before Start it prevents Start and Run. During Start it cancels execution and prevents Run unless Run already linearized. During Run it cancels the execution context. It never invokes Session Cleanup in caller goroutine.

The stable capability contains only a detached control cell. It exposes no Session, WebSocket, Context, CancelFunc, Send operation, callback registration, Runtime, Listener, or Manager record.

Before lease release, explicit RequestStop and Runtime callback enter through the same termination operation. Terminalization cleanup closes admission and drains already-entered calls through the callback cleanup contract. After confirmed callback cleanup, the detached cell is sealed: every later Stop-capability invocation returns `false` without reading or mutating released owner state.

## 18. Session Cleanup Semantics

Owner calls Session Cleanup once on every activated terminal path, including:

- termination before Start;
- Start error or panic;
- Run return, error, or panic;
- explicit Stop;
- Runtime cancellation.

Execution Owner creates the cleanup context. It is independent from the execution context, is not automatically canceled by root Runtime cancellation, and exists only for cooperative cleanup and join operations. Expiration of that context does not authorize Complete, lease release, or successful Wait while Cleanup has not actually returned its acknowledgement.

Cleanup incorporates the existing Session Stop semantics:

- closing transport;
- waiting for read loop when it exists;
- idempotent repeated observation;
- stable terminal result.

The current primary Stop/read-loop wait may block indefinitely independently of cleanup-context cancellation. This remains an accepted limitation. While Cleanup is blocked, lease remains active and Manager Wait remains pending.

`CleanupResult` has two cancellation outcomes:

- `Confirmed` permits the terminal chain to test lease-release eligibility;
- `Anomaly` means effective connection cancellation could not be proved.

For cancellation `Anomaly`, owner still follows the single terminal order: it attempts Complete, builds Terminal Result, invokes observer, closes control admission, successfully unregisters and drains the Runtime callback, seals the control cell, and reaches Terminal. It does not release the owner lease because effective cancellation was not confirmed. Manager Wait therefore remains blocked. If callback cleanup is itself unconfirmed, owner instead remains in `Terminalizing` with the lease active. Neither incomplete-shutdown outcome has a hidden retry. Operational intervention is outside this proposal.

## 19. Exclusive Completion Ownership

Execution Owner is the only production terminal completion caller for its Registration.

Owner does not receive general `Manager.Complete(RegistrationID)`. It receives one owner-scoped Completion Adapter created by Runtime composition from the Commit bundle:

```text
CompleteBoundRegistration() CompleteOutcome
```

The adapter cannot address any other Registration. Handshake, Dispatcher after activation, Host, Listener, Session, Snapshot, and observer do not receive it.

The technically exported internal `Manager.Complete` may remain as the low-level adapter mechanism, but production composition must not call or distribute it outside construction of the bound adapter.

Dependency and composition tests must prove:

- only the adapter package imports the low-level completion method for production wiring;
- owner receives the bound adapter, not Manager;
- other production constructors receive no completion capability;
- one owner invocation produces at most one effective Manager mutation.

`CompleteOutcome` is categorized as `Completed` or `AccountingAnomaly`. False/unknown removal becomes the anomaly category and is never retried.

## 20. Full Owner Lifetime Lease

Owner Lifetime Lease is created and published atomically by Commit. It remains active through the entire Runtime-owned lifetime.

The single terminal order for every committed path is:

```text
enter Terminalizing
    -> Session Cleanup returns
    -> Complete bound Registration
    -> build immutable Terminal Result
    -> synchronous Terminal Observer returns
    -> invoke panic-safe UnregisterAndDrain
       -> close admission to RequestStop and Runtime-callback entries
       -> prevent future callback entry
       -> unregister Runtime-cancellation callback
       -> wait for already-entered control calls
       -> return immutable CallbackCleanupResult
    -> if callback absence is not confirmed: remain Terminalizing with lease active
    -> seal and detach the control cell
    -> transition owner to Terminal
    -> if every release condition holds, invoke panic-safe Lease release adapter
    -> perform no further Runtime-owned work
    -> owned goroutine returns
```

Lease release is permitted only when all of these conditions hold:

- Cleanup returned and confirmed effective cancellation;
- the bound Complete invocation returned;
- Terminal Observer returned;
- callback installation and cleanup lifecycle completed;
- `UnregisterAndDrain` confirmed no future or entered callback work;
- every entered control call returned;
- control-call admission was closed;
- the causal cell was sealed.
- owner reached Terminal.

If callback cleanup cannot confirm absence of future and entered callback work, owner remains in `Terminalizing`, lease is not released, and Manager Wait cannot succeed. If owner reaches Terminal but another release condition, including effective-cancellation acknowledgement, is absent, lease likewise remains active. Successful lease release is the final Runtime-owned operation performed by the launch wrapper and the linearization point at which Manager may remove lease accounting. After it returns, the goroutine performs no Runtime-owned work and returns immediately.

Manager Wait guarantees absence of remaining Runtime-owned work; it does not claim scheduler-level observation that the goroutine stack has physically disappeared.

A blocked mandatory step keeps lease active. Panic in a mandatory step is isolated and recorded; later mandatory steps still execute when their prerequisites remain provable. Normal completion, explicit Stop, Runtime cancellation, callback installation failure, callback invocation panic, Start or Run failure, recovered execution panic, cancellation anomaly, Complete anomaly, and observer anomaly all use this same order. Only an unconfirmed callback cleanup prevents the `Terminalizing -> Terminal` transition.

## 21. Panic-Safe Terminal Chain

The launch wrapper owns one outer recover boundary. Each terminal operation also uses an independent safe-invocation boundary so a panic cannot skip later obligations.

Safe invocation rules:

| Operation | Panic treatment | Subsequent obligations |
| --- | --- | --- |
| Session Cleanup internals | Cleanup wrapper records a sanitized category and returns `CleanupResult`; outward panic is forbidden | Complete, observer, and callback cleanup continue; Terminal still requires Confirmed callback cleanup, and release requires confirmed cancellation |
| Cancellation cell internals | Cleanup wrapper recovers panic, invokes the private cancellation primitive, and records `CancellationPanic`; inability to confirm cancellation becomes `Anomaly` | Remaining chain continues; release is skipped for `Anomaly` |
| Runtime-callback installation | Recover panic or error as a sanitized installation anomaly; attempt the appropriate termination source | Common terminal chain begins; no rollback or Dispatcher result change |
| Runtime-callback invocation | Final guard decrements entered-call accounting; outward panic is converted to a sanitized callback anomaly and termination intent is still attempted | Callback returns without Session I/O or terminal waits |
| Completion Adapter | Record `CompletePanic` and accounting anomaly | Observer and callback cleanup continue; Terminal and release remain conditional |
| Terminal Observer adapter | Isolate panic as local operational anomaly outside Terminal Result; do not re-invoke | Callback cleanup continues; Terminal and release remain conditional |
| `UnregisterAndDrain` | Close entry admission, recover internal panic, unregister, and wait for entered calls; on completion return immutable `CallbackCleanupResult` | Confirmed result permits seal and Terminal; unconfirmed result keeps owner Terminalizing and lease active without retry |
| Lease release adapter | Recover panic and return explicit unsuccessful release outcome outside Terminal Result; do not re-panic | No hidden retry; lease accounting remains active and Wait cannot succeed |

Start and Run panic are recovered by the owned-goroutine boundary, categorized, and routed into Terminalizing.

No panic is rethrown from the owned goroutine. No diagnostics backend is defined. Process termination, unrecoverable Go runtime failure, memory corruption, and a permanently blocked operation remain outside recoverable guarantees.

## 22. Outstanding Termination Calls

RequestStop and the post-Commit Runtime-cancellation callback use the same control cell.

An invocation:

1. enters only while termination admission is open;
2. increments an invocation count;
3. attempts the first-writer termination mutation;
4. performs local execution-context cancellation if it won;
5. decrements count before return.

Entry into `Terminalizing` fixes lifecycle direction but does not itself close control-call admission. After observer returns, owner invokes idempotent `UnregisterAndDrain() CallbackCleanupResult`. The operation closes admission under the control lock, prevents future callback entry, performs panic-safe unregister, and waits outside Session, lifecycle, and Manager locks for already-entered calls to return. If the operation completes, it returns exactly one immutable result and never propagates panic. A permanently blocked entered call keeps the operation in progress, owner in Terminalizing, lease active, and Manager Wait blocked.

`Confirmed` proves that the registration cannot produce future callback work and that every entered callback/control call returned. Owner may then seal the cell and reach Terminal. `Unconfirmed` records the callback-lifetime anomaly; owner remains in Terminalizing, lease remains active, Manager Wait remains blocked, and no automatic retry occurs. Repeated cleanup returns the same detached result and has no second unregister or accounting effect.

Detached Stop capabilities may outlive Terminal. Immediately before lease release, their control cell is sealed and detached from owner state; subsequent calls always return `false`, including when the release adapter reports an unsuccessful outcome.

## 23. Terminal Result and Observer

Terminal Observer is synchronous and invoked once. Its return is covered by Owner Lifetime Lease.

Terminal Result is an immutable value containing only bounded enums and booleans:

- Start category: `NotAttempted`, `Succeeded`, `Failed`, `Panicked`;
- Run category: `NotStarted`, `Returned`, `Failed`, `Panicked`;
- Cleanup category: `NotRequired`, `Succeeded`, `Failed`, `Panicked`, `Blocked` only while no result is yet published;
- cancellation category: `Confirmed`, `Anomaly`;
- Complete category: `Completed`, `AccountingAnomaly`, `Panicked`;
- recovered execution-panic phase category without raw panic value;
- primary termination source;
- bounded secondary termination categories.

It stores no arbitrary raw error, credentials, headers, request, WebSocket, Context, callback, Session, or mutable collection. It intentionally represents only execution-lifecycle outcomes known when it is constructed before observer invocation: Runtime-callback installation or invocation anomaly already observed by that point, Start, Run, Cleanup/Stop, recovered execution panic, Complete, and termination source.

Callback invocation outcomes first arising after Terminal Result construction, `UnregisterAndDrain` outcome, Observer invocation outcome, lease-release outcome, and goroutine-return outcome are deliberately outside Terminal Result because they occur after the value is built. They never mutate the published result and never cause a second Observer invocation. Late callback outcomes remain bounded local terminal-accounting facts; callback cleanup alone determines whether callback lifetime permits Terminal. Observer panic is isolated by its adapter and does not prevent later callback cleanup or eligible lease release. Lease-release anomaly likewise does not mutate Terminal Result. These are local operational anomalies; diagnostics backend remains outside scope.

Observer blocking keeps lease active and Wait pending. Observer cannot call Complete, release lease, request Stop, or retain an execution capability through its contract.

## 24. Capability-Bearing Shutdown Snapshot

The current identity-only internal public Snapshot API is extended compatibly.

Each immutable Shutdown Registration retains existing:

- `SessionID`;
- `RegistrationID`.

It gains one read-only accessor returning a stable non-owning Stop-request capability bound to the committed owner control cell. Existing identity accessors and identity-only callers retain behavior.

Snapshot rules remain:

- first BeginShutdown atomically fixes membership;
- reservations are excluded;
- Commit winning before BeginShutdown is present; losing Commit is absent;
- repeated BeginShutdown returns the same logical Snapshot;
- Complete does not mutate captured Snapshot;
- capability remains safe after Complete and Terminal;
- capability remains safe after lease release and returns `false` without owner-state access;
- entry exposes no Session, Context, Runtime, Send, callback registration, or mutable owner state.

Existing Snapshot tests migrate by preserving all identity assertions and adding capability lifetime/concurrency assertions.

## 25. Failure Matrix

Every pre-Commit row uses the panic-safe sequence: publish non-committed, wait for dormant return, dispose owner-local values, Abort, and return `accepted=false`. No callback exists on that path. Every row beginning with successful Commit has one eligible owner path and uses the single terminal order. Confirmed callback cleanup permits Terminal; unconfirmed callback cleanup leaves owner Terminalizing.

| Failure or race | Dormant path outcome | Transport owner | Cancellation owner | Reservation outcome | Registration outcome | accepted/error | Complete | Observer | Callback outcome | Owner final state | Lease outcome | Manager.Wait outcome |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| Owner construction failure | Not created | Dispatcher | Dispatcher | Aborted | None | `false`, safe construction error | Not invoked | Not invoked | Never created | No owner | None | Unaffected after Abort |
| Dormant goroutine readiness failure | Non-committed path returns before Dispatcher | Dispatcher | Dispatcher | Aborted | None | `false`, safe readiness error | Not invoked | Not invoked | Never created | `PreCommit` disposed | None | Unaffected after Abort |
| Dispatcher panic after dormant creation | Non-committed published once; path joined | Dispatcher | Dispatcher | Aborted | None | `false`, sanitized panic error | Not invoked | Not invoked | Never created | `PreCommit` disposed | None | Unaffected after Abort |
| Attachment panic | Non-committed published once; path joined | Dispatcher | Dispatcher | Aborted | None | `false`, sanitized attachment error | Not invoked | Not invoked | Never created | `PreCommit` disposed | None | Unaffected after Abort |
| Runtime context canceled before Commit | Non-committed published once; path joined | Dispatcher | Dispatcher | Aborted | None | `false`, cancellation error | Not invoked | Not invoked | Never created | `PreCommit` disposed | None | Unaffected after Abort |
| BeginShutdown wins Commit | Non-committed published once; path joined | Dispatcher | Dispatcher | Aborted | None; Snapshot excludes it | `false`, Commit rejection | Not invoked | Not invoked | Never created | `PreCommit` disposed | None | Succeeds after Abort if otherwise empty |
| Recoverable Commit validation failure | Non-committed published once; path joined | Dispatcher | Dispatcher | Aborted | None | `false`, validation error | Not invoked | Not invoked | Never created | `PreCommit` disposed | None | Unaffected after Abort |
| Successful Commit | Committed once; one path eligible | Owner | Owner | Consumed | Registered with complete bundle | `true`, nil | Once at terminalization | Once | Installed only by owner | `Committed`, then normal lifecycle | Active until eligible release | Succeeds only after Registration and lease clear |
| Repeated Commit before Complete | Existing committed outcome returned | Owner | Owner | Already consumed | Same Registration and logical bundle | Same prior successful result | No new invocation | No new invocation | No second installation | Existing state unchanged | Same lease; no accounting change | Unchanged |
| Runtime context canceled after Commit before callback installation | Owner observes cancellation during race-safe install | Owner | Owner control cell | Consumed | Registered until Complete | Prior `true`, nil | Once | Once with RuntimeCanceled | Registration or post-registration check records RuntimeCanceled once | `Committed -> Terminalizing -> Terminal` after Confirmed cleanup | Released if all conditions hold | Pending until Complete and release |
| Callback installation succeeds | Owner continues after race-safe check | Owner | Owner control cell | Consumed | Registered until Complete | Prior `true`, nil | Once at later terminalization | Once | One owner-owned registration | `Committed -> Starting` or `Terminalizing` if termination won | Active until eligible release | Pending until terminal accounting clears |
| Callback installation panic or error | Owner enters common terminal path | Owner | Owner control cell | Consumed | Registered until Complete | Prior `true`, nil | Once | Once with installation anomaly | Partial or absent registration handled by cleanup contract | Terminal only after Confirmed cleanup | Released only after Terminal and all conditions | Blocked permanently if cleanup is Unconfirmed |
| RequestStop before owner scheduling | Committed path observes `ExplicitStop` | Owner | Owner control cell | Consumed | Registered until Complete | Prior `true`, nil | Once | Once with ExplicitStop | Installation and cleanup remain owner obligations | `Committed -> Terminalizing -> Terminal` after Confirmed cleanup | Released if all conditions hold | Pending until Complete and release |
| Callback entry during Starting | RuntimeCanceled competes before Run linearization | Owner | Owner control cell | Consumed | Registered until Complete | Prior `true`, nil | Once | Once | Entry accounted; Run forbidden if callback wins | `Starting -> Terminalizing -> Terminal` after Confirmed cleanup | Released if all conditions hold | Pending until Complete and release |
| Callback entry during Running | RuntimeCanceled cancels execution | Owner | Owner control cell | Consumed | Registered until Complete | Prior `true`, nil | Once | Once | Entry accounted and returns without Session I/O | `Running -> Terminalizing -> Terminal` after Confirmed cleanup | Released if all conditions hold | Pending until Complete and release |
| Callback wrapper panic | Panic sanitized; termination attempted; entry count decremented | Owner | Owner control cell | Consumed | Registered until Complete | Prior `true`, nil | Once | Once with callback anomaly only if known before Terminal Result construction; no second invocation for a later anomaly | Wrapper returns; a late anomaly remains a bounded terminal-accounting fact; Confirmed cleanup still required | Terminal after Confirmed cleanup; otherwise remains Terminalizing | Released if all conditions hold; otherwise remains active | Pending until Complete and release; blocked while cleanup is Unconfirmed |
| Normal completion | Committed path runs normally | Owner | Owner | Consumed | Removed by Complete | Prior `true`, nil | Once, effective once | Once with NaturalCompletion | `UnregisterAndDrain` Confirmed | `Running -> Terminalizing -> Terminal` | Released once | Succeeds after release |
| Callback already entered during terminalization | Owner waits outside lifecycle and Manager locks | Owner | Entered callback until return | Consumed | Complete already attempted | Prior `true`, nil | Once | Returned once | Drain waits for entry; no future entry | Terminalizing until drain, then Terminal | Active until Terminal | Blocked until entry returns and lease releases |
| Unregister success | Confirmed cleanup result | Owner | Sealed owner cell after drain | Consumed | Complete already attempted | Prior `true`, nil | Once | Returned once | No future or entered callback work | `Terminalizing -> Terminal` | Eligible subject to remaining conditions | Pending until release |
| Unregister panic or internal anomaly | Unconfirmed cleanup result; no outward panic | Owner | Owner control cell | Consumed | Complete already attempted | Prior `true`, nil | Once | Returned once | Absence of callback work unproved | Remains Terminalizing | Retained; no retry | Blocked |
| Unregister cannot prove non-entry | Unconfirmed cleanup result | Owner | Owner control cell | Consumed | Complete already attempted | Prior `true`, nil | Once | Returned once | Future or entered work cannot be excluded | Remains Terminalizing | Retained; no retry | Blocked |
| Repeated unregister | Existing immutable cleanup result returned | Owner | Same as first outcome | Consumed | Unchanged | Prior `true`, nil | No new invocation | No new invocation | No second unregister or drain | Unchanged | No accounting change | Unchanged |
| Cleanup cancellation anomaly | Committed path terminalizes | Owner | Cleanup wrapper | Consumed | Complete attempted once | Prior `true`, nil | Once | Once with Anomaly | Confirmed callback cleanup still required | Terminal after Confirmed callback cleanup | Retained because cancellation unconfirmed | Blocked |
| Complete false | Committed path continues | Owner | Owner control cell | Consumed | Accounting anomaly; this call removes nothing | Prior `true`, nil | Once, AccountingAnomaly | Once with Complete anomaly | Confirmed cleanup still required | Terminal after Confirmed callback cleanup | Released only if all conditions hold | Registration or lease prevents false success |
| Complete panic | Panic isolated; path continues | Owner | Owner control cell | Consumed | Accounting outcome may remain active | Prior `true`, nil | Once, Panicked | Once with CompletePanic | Confirmed cleanup still required | Terminal after Confirmed callback cleanup | Released only if all conditions hold | Remaining accounting prevents false success |
| Observer blocks | Path blocked in observer | Owner | Owner | Consumed | Complete already attempted | Prior `true`, nil | Once | In progress | Cleanup not yet reached in terminal order | Terminalizing | Active | Blocked |
| Observer panic | Adapter returns sanitized anomaly | Owner | Owner control cell | Consumed | Complete already attempted | Prior `true`, nil | Once | Once; panic isolated | `UnregisterAndDrain` follows | Terminal after Confirmed cleanup | Eligible subject to remaining conditions | Pending until release |
| Lease release anomaly | All earlier steps complete | No remaining transport work | Sealed cell | Consumed | Complete already attempted | Prior `true`, nil | Once | Returned once | Confirmed; no callback work | Terminal | Retained; no retry | Blocked |

An unrecoverable low-level completion panic may leave Registration accounting active even after owner lease exit. An unconfirmed lease release leaves lease accounting active. In either case Manager Wait remains blocked rather than reporting false success.

## 26. Manager Wait Guarantee

After successful `Manager.Wait`:

- Reservation set is empty;
- Registration set is empty;
- Owner Lifetime Lease set is empty;
- every tracked owner reached Terminal;
- every Runtime-cancellation callback registration was removed;
- all entered control calls returned;
- every Session Cleanup returned its immutable acknowledgement;
- every tracked connection context is effectively canceled;
- every synchronous Terminal Observer returned;
- after successful lease release no Runtime-owned work remains.

If Cleanup, observer, control-call drain, Registration removal, or lease release does not complete as required, accounting remains active and Wait does not succeed. A caller deadline may end that caller's Wait without changing truthful Manager state.

Wait does not promise scheduler-level proof that an owned goroutine stack has physically returned after its final non-Runtime instruction.

## 27. Shutdown Ordering

The normative order is:

```text
Host closes Admission Gate
    -> Manager.BeginShutdown captures capability-bearing Snapshot
    -> invoke Snapshot RequestStop capabilities
    -> cancel root Runtime context
    -> initiate Listener Stop
       |-> Listener handler drain -> Listener Stop returns
       |-> owners terminalize -> eligible leases release
    -> after Listener Stop returns, Manager Wait
    -> remaining Runtime components stop
```

The handler-drain and owner-drain branches are concurrent; neither is ordered before the other. Manager Wait begins only after Listener Stop returns. Owners committed before BeginShutdown appear in Snapshot. An in-flight Reservation whose Commit loses aborts and Dispatcher cleans transport. A Commit that wins is irreversibly published, appears in Snapshot, transfers ownership, returns `accepted=true`, and follows the normal owner terminal lifecycle.

Explicit Stop and root cancellation share one first-writer state, so their ordering creates no second Stop obligation.

This preserves ARCH-002: admission closes first and root Runtime context is canceled before Listener Stop.

## 28. Deadlock Analysis

- The pre-Commit proof is `panic-safe Dispatcher boundary -> non-committed publication -> dormant path return -> owner-local disposal -> Abort -> accepted=false`. No callback exists on this path, and Dispatcher cannot return while a recoverable pre-Commit dormant path remains blocked.
- The Commit proof is `precompute complete bundle -> panic-free atomic publication under Manager lock -> committed gate outcome -> unlock -> one owner path eligible`.
- The post-Commit proof is `owner path -> install callback -> lifecycle -> Terminalizing -> Cleanup -> Complete -> observer -> UnregisterAndDrain -> seal -> Terminal -> lease outcome`. Unconfirmed callback cleanup stops at Terminalizing.
- Dispatcher does not wait for owner after successful Commit, and owner does not wait for Dispatcher after observing `Committed`.
- Manager never waits for the dormant path. The dormant path waits only for its one-shot outcome and never waits for Manager after that outcome exists.
- Commit performs only bounded panic-free in-memory publication and no callback, channel send, goroutine scheduling, Session I/O, or owner method while holding Manager locks.
- One-shot publication stores one terminal outcome before waking its single waiter; repeated publication observes the existing outcome, so lost wakeup and double release are impossible by contract.
- Owner Start, Run, and Cleanup execute without owner or Manager locks.
- Callback entry performs only causal-cell mutation and returns; it waits for no owner terminal operation.
- RequestStop and Runtime cancellation do not wait for Session operations.
- After observer, owner holds no lifecycle or Manager lock while `UnregisterAndDrain` waits for entered calls. Callback never waits for that cleanup and therefore cannot form a lock cycle.
- Completion Adapter performs only Manager mutation and never waits for Manager Wait.
- Observer cannot call Complete, release lease, or wait for owner through its contract.
- Lease release never waits for Manager Wait.
- Listener Stop may wait for HTTP handlers, while owners terminalize in parallel; owner does not wait for Listener Stop.
- Manager Wait waits only for accounting convergence and holds no lock while blocked.

There is no circular wait between Dispatcher and prepared owner, and no recoverable Dispatcher panic can orphan the dormant path. Manager Wait never participates in callback cleanup, and Dispatcher has no callback or owner obligation after Commit. Permanently blocked Session Cleanup, observer, callback entry, or unconfirmed callback cleanup may delay truthful shutdown, but none creates a lock cycle under these contracts.

## 29. Migration from Current Dispatcher

Migration is sequential:

1. introduce transport-independent Core without changing current Session;
2. add package-private provisional Session and owner preparation;
3. add cleanup acknowledgement around current Stop and connection cancellation;
4. add Manager Commit bundle, opaque one-shot execution binding, and compatible Snapshot capability accessor;
5. add dormant launch-path preparation and Commit-boundary tests;
6. add owner-only post-Commit Runtime-callback installation and `UnregisterAndDrain` proof;
7. move Start/Run/Cleanup into owner after Commit succeeds;
8. wire Manager shutdown ordering through Runtime composition.

At every step, Handshake still sees `DispatchAuthenticated(...)(accepted, error)`. `accepted=true` is returned only after successful Commit. `accepted=false` always means Commit did not occur, Dispatcher still owns cleanup, and no Registration, lease, or Snapshot capability exists.

## 30. Dependency Boundaries and God-Object Control

Execution Owner coordinates one Session lifecycle only. Each additional entity has one independent invariant:

- cancellation cell proves effective connection cancellation;
- Completion Adapter prevents cross-registration completion;
- Lifetime Lease covers goroutine lifetime beyond registration visibility;
- Terminal Result transports immutable terminal categories;
- Terminal Observer consumes exactly one result;
- Stop capability requests termination without Session ownership.

None is a service locator or generic coordination framework. Owner does not gain routing, delivery, persistence, presence, limits, or supervision responsibilities.

## 31. Architectural Invariants

- Dispatcher is the sole handoff coordinator.
- Dispatcher receives no ambiguous outcome: failed Commit means no transfer; successful Commit means irreversible transfer.
- Core is not Session and owns no transport.
- Dispatcher creates exactly one dormant launch goroutine and one one-shot Commit gate per prospective owner.
- Dispatcher owns one panic-safe pre-Commit orchestration boundary, creates no Runtime callback, and cannot return until a non-committed dormant path has returned and prepared values are disposed.
- Commit is the only Registration, lease, Stop-capability, ownership, and execution publication point.
- Commit publication contains no fallible external operation and has only complete non-committed or complete committed outcomes.
- Commit does not return success or release its synchronization boundary until the execution binding is `Committed`.
- No observable Registration exists without an already eligible execution path.
- Complete is the only Registration removal point.
- Owner preparation precedes Commit; Session execution begins only after Commit.
- Every committed path enters Terminalizing before Terminal.
- Start and Run each have one linearization point.
- Owner installs observation of only the root Runtime context after Commit and before Start, using race-safe register-and-check semantics.
- Explicit Stop and Runtime cancellation share one first-writer termination state.
- Session Cleanup synchronously acknowledges effective canceled state.
- Owner alone receives bound completion and lease capabilities.
- Terminal Observer is synchronous and invoked once.
- Confirmed `UnregisterAndDrain` proves callback absence, entered control calls drain, and the cell seals before Terminal; unconfirmed cleanup remains Terminalizing.
- Lease release is the final Runtime-owned operation.
- Successful Manager Wait covers Terminal and absence of remaining Runtime-owned work, not scheduler-level goroutine return.
- Snapshot capability exposes no Session or mutable execution state.

## 32. Alternatives

### Synchronous Dispatcher

Rejected because it cannot provide independent Runtime execution or shutdown capability.

### Owner Executes Attachment

Rejected because it requires a transport-transfer capability in owner and splits the synchronous Upgrade cleanup boundary.

### Dispatcher Activation Executor

Selected because Dispatcher already owns Upgrade resources, can make Commit itself the ownership-acceptance and execution-publication point, and can return one explicit handoff outcome.

### Provisional Session in `Created`

Rejected because one state would represent both missing and owned transport.

### Transport-Independent Core

Selected because no Session lifecycle exists before attachment.

### Complete as Final Accounting

Rejected because observer and full goroutine lifetime would not be represented after registration removal.

### Global Supervisor

Rejected because one per-Session invariant does not justify a generic framework.

## 33. TASK-REV-004 Traceability

| Finding | Status | Resolution |
| --- | --- | --- |
| F-01 | Resolved | Session Cleanup returns effective-cancellation acknowledgement. |
| F-02 | Resolved | Lease release is the final Runtime-owned operation before immediate goroutine return. |
| F-03 | Resolved | Run linearizes at `Starting -> Running`. |
| F-04 | Resolved | Core is not Session; attachment creates the transport-owning Session. |
| F-05 | Resolved | One shutdown order preserves ARCH-002. |
| F-06 | Resolved | Owner receives only a bound Completion Adapter. |
| F-07 | Resolved | Observer is synchronous, one-shot, categorized, and lease-covered. |
| F-08 | Accepted limitation | Permanently blocked Cleanup keeps accounting active. |

## 34. TASK-REV-005 Traceability

| Finding | Status | Resolution |
| --- | --- | --- |
| F-01 | Resolved | CleanupResult synchronously acknowledges effective cancellation. |
| F-02 | Resolved | Lease release is the final Runtime-owned operation with no Runtime epilogue. |
| F-03 | Resolved | Every cleanup step has independent safe-invocation semantics. |
| F-04 | Resolved | Dispatcher cannot abandon the complete activation transaction. |
| F-05 | Resolved | Dispatcher is the only attachment/activation executor. |
| F-06 | Resolved | Transport-independent Core is not a `Created` Session. |
| F-07 | Resolved | Attachment is package-private, atomic, one-use, and concurrency-defined. |
| F-08 | Resolved | Reservation remains with Dispatcher until owner readiness. |
| F-09 | Resolved | Explicit Stop and Runtime cancellation share first-writer intent. |
| F-10 | Resolved | All committed terminal paths enter Terminalizing. |
| F-11 | Clarified | Invocation count and effective cancellation are distinct. |
| F-12 | Clarified | RequestStop promises no Session wait, not a hard time bound. |
| F-13 | Resolved | Owner receives registration-bound Completion Adapter only. |
| F-14 | Resolved | Commit atomically returns one identity-bound complete bundle. |
| F-15 | Clarified | Snapshot gains a compatible immutable capability accessor. |
| F-16 | Resolved | Terminal Result contains bounded categories and no raw error. |

No TASK-REV-005 Blocker or High finding is Deferred.

## 35. TASK-REV-006 Traceability

| Finding | Status | Resolution |
| --- | --- | --- |
| Attachment versus acknowledgement ownership | Resolved | Successful Commit transfers transport, cancellation, Session, and execution ownership together. |
| Activation linearization | Resolved | Commit returns either irreversible publication or no committed state. |
| Terminal and lease ordering | Resolved | Owner closes callback admission, unregisters, drains entered calls, seals the cell, reaches Terminal, and then performs the final lease-release operation; Wait guarantees no remaining Runtime-owned work, not scheduler-level goroutine disappearance. |
| Cleanup-panic acknowledgement | Resolved | Cleanup is the outward panic-safe wrapper and always returns immutable cancellation acknowledgement. |
| Observer and lease late panic outcomes | Resolved | They are isolated operational anomalies outside the already-built Terminal Result. |
| Runtime callback lifetime | Resolved by TASK-REV-011 | Owner creates observation only after Commit; confirmed `UnregisterAndDrain` is required before Terminal. |
| Pre-Commit observation | Resolved | Failure returns through Dispatcher/Handshake as ordinary error; committed Terminal Observer is not invoked. |
| DP-002 shutdown ordering | Resolved | DP-002 is synchronized to the ARCH-002-compatible active-Session order defined here. |
| Cleanup-context ownership | Resolved | Owner creates an independent cooperative cleanup context; its expiry cannot authorize false completion. |

No TASK-REV-006 Blocker or High finding is Deferred.

## 36. TASK-REV-007 Traceability

| Finding | Status | Resolution |
| --- | --- | --- |
| F-01 | Resolved | Commit is the single ownership and Registration publication point. |
| F-02 | Resolved | Cancellation `Anomaly` reaches Terminal but keeps lease active and Wait blocked without retry. |
| F-03 | Resolved | Listener handler drain and owner drain are parallel after Listener Stop is initiated. |
| F-04 | Resolved | All termination sources compete through one causal cell; later sources are bounded secondary categories. |
| F-05 | Resolved | The failure matrix states transport, cancellation, Reservation, Registration, owner, lease, and Wait outcomes explicitly. |
| F-06 | Clarified | Exact lease-release conditions and the no-Runtime-work-after-release boundary are normative. |

No TASK-REV-007 Blocker or High finding is Deferred.

## 37. TASK-REV-008 Traceability

| Finding | Status | Resolution |
| --- | --- | --- |
| F-01 — Commit visibility before handoff success | Resolved | Commit itself is handoff success and the only irreversible publication point. |
| F-02 — BeginShutdown captures an unfinished handoff | Resolved | A Commit winning BeginShutdown is already a successful handoff; a losing Commit publishes nothing and must Abort. |
| F-03 — Undefined post-Commit lifecycle branch | Resolved | The branch is removed. Every post-Commit path uses the normal `Committed -> Terminalizing -> Terminal` lifecycle. |
| F-04 — Missing Commit/BeginShutdown matrix outcome | Resolved | The matrix contains explicit rows for both race winners and their Snapshot, ownership, lease, and Wait outcomes. |
| F-05 — `accepted=true` with synchronous error | Resolved by TASK-REV-010 | Generic and target-production Dispatcher semantics are now explicitly separated. |
| F-06 — anomaly ordering wording | Resolved by TASK-REV-010 | One Terminal order now applies to every committed path. |

Only F-01 through F-04 are closed by this revision.

## 38. TASK-REV-009 Commit-to-Execution Traceability

| Finding | Status | Resolution |
| --- | --- | --- |
| F-01 — Commit-to-owner execution boundary | Resolved | Dispatcher creates exactly one dormant launch goroutine. The integrated Commit operation publishes Registration and changes its one-shot execution binding to `Committed` under the same synchronization boundary before returning success. |
| F-02 — Runtime-cancellation callback timing | Superseded by TASK-REV-011 | The final model removes pre-Commit registration and gives post-Commit installation to owner. |
| F-03 — cancellation-anomaly Terminal ordering | Resolved by TASK-REV-010 | Cancellation anomaly follows the single common Terminal order. |
| F-04 — DP-003 later registration failure after ownership transfer | Resolved | DP-003 now states that Commit itself simultaneously publishes Registration and transfers Session ownership; no later registration failure exists. |
| F-05 — repeated Commit and owner bundle | Resolved | Repeated Commit for an existing Registration returns the same bound publication observation and cannot create or release another execution path. |
| F-06 — incomplete failure-matrix execution outcome | Clarified | The matrix now states the common pre-Commit and post-Commit accepted and launch-path outcomes; its unrelated terminal columns are unchanged. |
| F-07 — common-interface `accepted=true, error!=nil` | Resolved by TASK-REV-010 | Generic interface ownership semantics remain valid while the target Dispatcher uses only `true,nil` after Commit. |

Only the Commit-to-execution findings F-01, F-04, and F-05 are closed. F-06 is clarified but not claimed as a full redesign of the matrix.

## 39. TASK-REV-010 Traceability

| Finding | Status | Resolution |
| --- | --- | --- |
| F-01 — pre-Commit terminal guarantee | Resolved | One panic-safe Dispatcher boundary publishes non-committed once, joins the dormant path, disposes owner-local values, aborts, and returns `accepted=false`; no callback exists before Commit. |
| F-02 — Commit panic atomicity | Resolved | Fallible work completes before the critical section; publication invokes no external code and has only complete non-committed or complete committed outcomes. |
| F-03 — Runtime-cancellation callback timing | Superseded by TASK-REV-011 | Owner installs observation only after Commit and before Start through one race-safe registration-and-check contract. |
| F-04 — Terminal ordering | Resolved by TASK-REV-011 | Every committed terminal cause uses one order, and only confirmed callback cleanup permits Terminal. |
| F-05 — incomplete Failure Matrix | Resolved by TASK-REV-011 | The matrix now separates callback installation, invocation, cleanup, Complete, Observer, owner-state, lease, and Wait outcomes. |
| F-06 — execution-binding terminology | Clarified | The binding is a narrow one-shot publication capability, not a passive value or general lifecycle control. |
| F-07 — accepted/error contract | Resolved | The generic interface permits true plus error as an ownership result; the target production Dispatcher emits only pre-Commit false plus error or successful Commit true plus nil. |

No TASK-REV-010 Blocker or High finding is Deferred.

## 40. TASK-REV-011 Traceability

| Finding | Status | Resolution |
| --- | --- | --- |
| F-01 — unregister anomaly versus Terminal | Resolved | `UnregisterAndDrain` must confirm absence of future and entered callback work before Terminal; unconfirmed cleanup remains Terminalizing with lease and Wait active. |
| F-02 — pre-Commit callback join | Resolved | Runtime callback does not exist before Commit; only the dormant owner path must be joined on non-committed outcomes. |
| F-03 — two pre-Commit disposal orders | Resolved | The only order is non-committed publication, dormant return, owner-local disposal, Abort, and `accepted=false`. |
| F-04 — incomplete Failure Matrix | Resolved | Separate rows cover pre/post-Commit cancellation, installation, invocation, cleanup, Complete, Observer, and lease anomalies. |

No TASK-REV-011 Blocker or High finding is Deferred.

## 41. Impact on DP-003

DP-003 remains authoritative for Reserve, Abort, registration identity, Lookup, Complete mutation, Manager lifecycle, BeginShutdown, Wait, and identity Snapshot semantics.

This document adds the normative integrated Commit bundle, owner-lifetime lease accounting, and compatible capability-bearing Snapshot accessor. These additions do not change existing identity or completion linearization points.

DP-003 does not duplicate execution state, attachment, cleanup, or termination behavior.

## 42. Open Questions

- Exact Go names and private package placement of Core, Commit bundle, adapters, and categorized values.
- Exact enum names for terminal categories.
- Test-only instrumentation needed to prove that no Runtime-owned work occurs after successful lease release.

These remaining questions are implementation representation and proof-instrumentation choices. Callback timing, installation-race semantics, callback cleanup, terminal ordering, accepted/error ownership semantics, and the Commit-to-execution publication boundary are no longer open.

## 43. Complexity Self-Check

- Conceptual owner states: six — one pre-Commit state and five committed states.
- Committed accounting dimensions: two — Registration and Owner Lifetime Lease. Reservation remains pre-Commit transaction accounting.
- Ownership-transfer points: one.
- Execution-publication points: one, inside Commit.
- Dormant launch paths per prospective owner: one.
- Terminal Observer invocations: one.
- Completion callers per Registration: one bound owner.
- Generic supervisors or policy frameworks introduced: zero.

## 44. Review Readiness

All TASK-REV-010 and TASK-REV-011 Blocker and High findings are resolved normatively. No externally observable committed Registration can exist without one already eligible execution path, every recoverable pre-Commit path terminates without creating a callback, and owner-only callback installation has one race-safe model. Terminal is reachable only after confirmed callback cleanup.

Accepted limitations remain process termination or unrecoverable Go runtime failure, scheduler starvation, permanently blocked Session Cleanup, observer, or entered callback, unconfirmed callback cleanup, cancellation acknowledgement anomaly, and unsuccessful lease release. Callback-lifetime anomaly remains Terminalizing; cancellation acknowledgement anomaly may reach Terminal, but both keep lease accounting and Wait active rather than producing false completion.

**Approval decision:** Approved.

Approval follows the completed review trail in Sections 33–40. Callback timing is singular, pre-Commit callback registration is absent, and Terminal has one provable definition. TASK-REV-013 Codex concluded Approved with one non-blocking clarity finding, TASK-REV-013 Kiro concluded Approved, and TASK-DOC-016 resolved the remaining clarity and synchronization findings without changing the architecture. The items in Section 42 are implementation representation and proof-instrumentation choices, not unresolved architectural decisions. No Blocker or High architectural finding remains open.

## 45. References

- [ARCH-001: Runtime Architectural Pattern](../architecture/ARCH-001-runtime-architectural-pattern.md)
- [ARCH-002: Runtime Foundation Freeze](../architecture/ARCH-002-runtime-foundation-freeze.md)
- [DP-001: Runtime Handshake Pipeline](DP-001-runtime-handshake-pipeline.md)
- [DP-002: Runtime Host Composition Root](DP-002-runtime-host-composition-root.md)
- [DP-003: Runtime Session Manager](DP-003-runtime-session-manager.md)
- [ADR-0004: Handshake Runtime Dependency Boundary](../adr/0004-handshake-runtime-dependencies.md)

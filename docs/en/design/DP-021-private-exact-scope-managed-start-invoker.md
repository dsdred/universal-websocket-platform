# DP-021: Private Exact-Scope Managed Start Invoker

[Russian version](../../ru/design/DP-021-private-exact-scope-managed-start-invoker.md)

## 1. Status

- **Design Status:** Draft
- **Implementation Status:** Partial — implemented in isolation

This focused proposal defines one composition-private DP-013 invocation
contract. TASK-043 implements that object in isolation inside the existing
package; it adds no public API, transport, policy, persistence, or production
wiring. Approved DP-014 through DP-019 remain unchanged, and TASK-026 remains
unimplemented. Completed and Coordinator-Accepted TASK-044 (2026-08-24)
historically records `UNBLOCK TASK-026`. A superseding TASK-026 recheck
confirms one missing DP-015 tracked-Start managed-parent plus preclaimed
`StopOld` admission prerequisite. TASK-026 is blocked, the prerequisite is not
activated, and this DP remains Draft/Partial.

## 2. Purpose

Define the exact internal object that joins an immutable DP-013 management
scope to one already-constructed scope-bound managed Flow and the accepted
authorization and command-binding seams as the sole lifecycle subcall of a
future TASK-026 orchestrator-owned callback closure, without making the
invoker itself a callback or creating another Flow, lifecycle, authorization,
command, or orchestration owner.

## 3. Authority

This proposal refines, without overriding:

- [ARCH-004](../architecture/ARCH-004-runtime-deployment-and-identity-model.md)
  for management authorization-before-mutation and lifecycle ownership;
- [DP-011](DP-011-runtime-launch-pipeline-integration.md) for synchronous
  managed Flow execution;
- [DP-013](DP-013-runtime-management-routing.md) for exact Target ownership and
  composition-private routing;
- [DP-014](DP-014-runtime-operational-identity-persistence.md) for attempt and
  execution-generation facts;
- [DP-015](DP-015-runtime-management-command-idempotency.md) for authorization,
  claim, live callback capability, replay, and unresolved barriers;
- [DP-016](DP-016-runtime-activation-replacement-rollback.md) for unchanged
  activation, replacement, and rollback ordering;
- [DP-019](DP-019-runtime-activation-orchestration-prerequisites.md) for the
  Approved exact-scope private-invocation obligations;
- [DP-020](DP-020-runtime-orchestration-binding-sequence-readiness.md) for the
  dependency-neutral binding and managed primitive/linked adapters.

Higher-authority statuses and semantics do not change.

## 4. Scope

The design covers only:

- construction of one immutable invoker around one already-constructed managed
  Flow for one exact Runtime Instance scope;
- exact invocation validation before Owner mutation;
- synchronous delegation to one stored managed Flow;
- primitive Start and linked parent/`StartTarget` upstream mappings;
- the type-correct boundary between the future TASK-026 orchestrator-owned
  DP-015 callback closure, invoker `StartOutcome`, and out-of-DP-021 terminal
  mapping/publication;
- per-call capability custody and lifetime;
- cancellation, failure, panic, `runtime.Goexit`, and indeterminate behavior;
- dependency direction and the absence of legacy fallback.

## 5. Non-goals

This proposal does not define:

- production code, exported product API, HTTP/CLI/DTO routes, or transport;
- a Principal model or concrete authorization policy;
- DP-014 terminal publication or DP-015 command/phase terminalization;
- mapping a returned Owner `StartOutcome` to a DP-015 `TerminalOutcome`;
- the DP-016 orchestrator or activation/replacement/rollback selection policy;
- persistence, recovery, reporting, supervision, or production composition;
- any change to public DP-013 `Directory.Start`, `Stop`, or `Observe`;
- adoption of legacy command records or unmanaged Flow execution.

## 6. Ownership and Construction

The invoker is owned by the existing internal `runtimemanagement` composition
boundary. Its conceptual construction contract is:

```text
NewManagedStartInvoker(
    OperationalDomain,
    runtimemanagement.Target,
    alreadyConstructed *runtimelaunchflow.ManagedFlow,
) -> ManagedStartInvoker, error
```

Construction validates:

1. `OperationalDomain` is non-empty and opaque;
2. `Target` is one valid immutable DP-013 scope;
3. `alreadyConstructed` is non-nil.

Before this constructor, composition calls `runtimelaunchflow.NewManaged`
exactly once with the same Binding's Owner and Loader plus the accepted
`StartClaimContinuation`. The composition audit proves that the copied Target,
Binding Owner/Loader, and preconstructed Flow belong to the same immutable
scope. The invoker neither repeats nor replaces that construction.

On success the invoker stores only:

- the exact immutable `OperationalDomain`;
- a copy of the exact immutable DP-013 `Target`;
- one borrowed immutable reference to the already-constructed scope-bound
  managed Flow.

It accepts and retains no Binding, Owner, Loader, continuation, authorization
policy, command/parent/phase record, command key, permit,
`StartExecutionBinding`, rendezvous, caller context, current-operation slot,
Owner claim view, Host, Snapshot, or terminal publisher. It creates zero Flow,
registry entry, goroutine, detached callback, or second Flow per call.

No Flow introspection, accessor, scope token, or new identity is added to prove
construction alignment. That proof belongs to the composition audit over the
original Binding inputs and object identities.

The conceptual stable errors are `ErrInvalidManagedStartInvoker` for failed
domain/Target/non-nil-Flow construction validation and
`ErrInvalidManagedStartInvocation` for receiver, context, request, or binding
validation at invocation. An upstream `NewManaged` error occurs before invoker
construction and is returned by composition unchanged. An exact downstream
Flow or Owner error is not converted to either sentinel.

## 7. Sole Invocation Contract

The invoker exposes one composition-private operation:

```text
InvokeManagedStart(
    context.Context,
    runtimelifecycle.StartRequest,
    runtimeorchestrationbinding.StartExecutionBinding,
) -> runtimelifecycle.StartOutcome, error
```

No Stop, Observe, parent, phase, terminal-publication, or generic Execute
operation is added.

The invoker is not a DP-015 callback and never returns a DP-015
`TerminalOutcome`. Future TASK-026 composition owns the callback closure passed
to a managed DP-015 primitive or linked adapter. That closure makes
`InvokeManagedStart` its sole lifecycle subcall, receives the exact
`StartOutcome` and error, and performs any later mapping, DP-014 terminal
publication, and DP-015 command/phase terminalization outside DP-021.

The invoker never calls a DP-015 `Boundary`, creates a Flow, creates or
inspects a command, maps a terminal result, publishes identity, or terminalizes
a command/phase.

## 8. Exact Validation Before Owner Mutation

Before delegating to the stored managed Flow, invocation validates in order:

1. receiver construction is valid;
2. context is non-nil;
3. `StartRequest` contains non-zero Workspace, Configuration, and target
   ConfigurationVersion identities;
4. `StartExecutionBinding.Valid()` is structurally true;
5. binding authorization `OperationalDomain` equals the stored domain;
6. binding Workspace, Configuration, and Runtime Instance equal the stored
   Target;
7. `StartRequest` Workspace, Configuration, and target version equal the
   binding authorization tuple;
8. the closed binding shape is either primitive `ActivateExactTarget` with no
   linked identity, or exact Replace/Rollback with its all-or-none derived
   `StartTarget` identity.

Any observable mismatch returns the exact invocation-validation error before
`Owner.PrepareStart`, Load, Build, Launcher, Host, DP-014 write, or any other
lifecycle mutation. Validation does not inspect mutable aggregate or command
state and cannot retarget an invocation.

The invoker cannot and must not introspect the preconstructed Flow to re-prove
its Owner or Loader. Exact Flow-to-Target alignment is the preceding
composition audit obligation from section 6; adding an accessor or validation
token would widen the managed Flow contract and is forbidden.

Structural `Valid()` proves only that immutable fields and their cross-field
shape are well formed. It is not a live permit, rendezvous, callback-authority,
generation, or custody proof. The invoker cannot distinguish a freshly
callback-delivered binding from a retained structurally valid value.

## 9. Cancellation Linearization

A nil context is invalid and causes zero mutation.

A non-nil context is never rejected by the invoker merely because
`ctx.Err()` is already non-nil. Once DP-015 has entered the newly claimed
future TASK-026 orchestrator-owned callback closure and that closure makes its
sole lifecycle subcall, the invoker must call the managed Flow with that
context. Flow owns
the pre-Owner cancellation check and the mandatory exact `StartNoClaim` signal
to the command-owned rendezvous. An invoker early return would lose that
signal and could strand an independently admitted Stop.

After authentic Owner claim, the existing managed Flow preserves context
values while ignoring caller cancellation and deadlines through
`context.WithoutCancel`; the invoker neither adds cancellation authority nor
changes that behavior.

## 10. Authorization and Command Ordering

Authorization is not an invoker responsibility.

For every initial, in-progress, and replay submission, DP-015 evaluates the
exact six-field `AuthorizeOrchestration` request before command inspection or
mutation. Only a newly committed claim gives the future orchestrator-owned
callback closure its complete binding. The closure calls the invoker once as
its sole lifecycle subcall. The invoker validates that accepted tuple against its stored scope; it
does not invoke policy a second time, cache authority, reinterpret denial, or
authorize a linked phase independently from its accepted parent.

Therefore authorization-before-mutation is compositional:

```text
DP-015 validate + authorize + pre-claim cancellation
    -> claim and orchestrator-owned callback closure, only for a new claim
        -> DP-021 exact scope/request validation, sole lifecycle subcall
        -> ManagedFlow.StartManaged
        -> Owner.PrepareStart
        -> exact StartOutcome/error back to the closure
    -> closure-owned mapping/publication/terminalization outside DP-021
```

Replay and in-progress observations receive no callback closure and therefore
cannot invoke the invoker through this protocol.

## 11. Primitive and Linked Call Paths

### 11.1 Primitive Start

Future TASK-026 composition supplies an orchestrator-owned callback closure to
`Boundary.ExecuteManagedStart`. DP-015 constructs the primitive binding with
`ActivateExactTarget`, no parent/phase identity, one exact aggregate revision,
execution generation, and live rendezvous. The closure invokes the one
scope-bound invoker exactly once as its sole lifecycle subcall, receives exact
`StartOutcome`/error, and performs terminal mapping/publication outside DP-021.

### 11.2 Replace/Rollback StartTarget

Composition first enters `Boundary.ExecuteManagedParent`, then uses only its
callback-scoped `ManagedParentExecution` and
`ContinueOrExecuteManagedStartTarget`. DP-015 derives the exact parent and
ordinal-one `StartTarget` identities and constructs the linked binding. That
future orchestrator-owned phase callback closure invokes the same scope-bound
`InvokeManagedStart` operation as its sole lifecycle subcall, then owns result
mapping/publication outside DP-021.

The invoker contains no primitive-versus-parent branch beyond validating the
closed binding shape. Both paths converge on the same exact Target, Start
request, binding, and managed Flow contract.

Legacy `Execute`, `ExecuteParent`, `ContinueOrExecuteStartTarget`, public
`Directory.Start`, and unmanaged `Flow.Start` cannot be adopted, upgraded, or
used as fallback.

## 12. Per-call Capability Lifecycle

`StartExecutionBinding` is an immutable structural value delivered by DP-015
to the future orchestrator-owned callback closure on the original synchronous
permit-holding stack. The closure lends it to the invoker for its sole
lifecycle subcall; the invoker passes it unchanged to `StartManaged`.

The invoker:

- never stores, copies into long-lived state, caches, indexes, replaces, or
  reuses the binding or rendezvous;
- cannot manufacture a permit or resolve the opaque rendezvous;
- returns before the DP-015 callback capability expires;
- exposes no binding, Flow, continuation, Owner, or Target accessor that could
  widen custody.

The stored Flow reference is borrowed immutably for the lifetime of the one
scope composition. The invoker does not own, close, reconstruct, swap, or
duplicate it. Composition must keep the preconstructed Flow alive for the
invoker lifetime and retire both together with that immutable scope.

Return, panic, `runtime.Goexit`, boundary generation loss, or callback expiry
expires the live permit, rendezvous lookup, and callback authority under
DP-015 ownership. It does not mutate or invalidate the immutable
`StartExecutionBinding` value; that value may remain structurally `Valid()`.
No-reuse is therefore a callback-custody and no-bypass invariant, not an
invoker-enforced liveness check. A caller retaining the value violates the
contract, and the invoker cannot distinguish that retained value from a fresh
structurally valid binding.

## 13. Dependency Direction

Construction direction is:

```text
runtimemanagement composition Binding(Target, Owner, Loader)
    -> runtimelaunchflow.NewManaged(Owner, Loader, continuation), exactly once
    -> composition audit of the same Binding/Target/object identities
    -> NewManagedStartInvoker(domain, Target, preconstructed ManagedFlow)
```

Invocation direction is:

```text
runtimecommandidempotency managed adapter
    -> future TASK-026 orchestrator-owned callback closure
        -> runtimemanagement private invoker, sole lifecycle subcall
            -> runtimelaunchflow.ManagedFlow
            -> exact StartOutcome/error back to closure
        -> mapping/publication/terminalization outside DP-021
```

`runtimemanagement` may depend on `runtimelaunchflow`,
`runtimelifecycle`, and dependency-leaf `runtimeorchestrationbinding`. It must
not import `runtimecommandidempotency`, `runtimeidentity`, transport, or a
future orchestrator to implement this invoker. The invoker never calls upward
into the command boundary. The dependency-leaf binding package continues to
depend on none of these higher packages. `runtimelaunchflow` does not depend
back on `runtimemanagement`; the preconstructed reference introduces no cycle.

## 14. Results, Failures, Panic, and Indeterminate Outcomes

The design distinguishes:

- **upstream managed Flow construction error** — returned unchanged by
  composition before the invoker constructor is called;
- **invoker construction error** — invalid domain, Target, or nil
  preconstructed managed Flow; no invoker is returned and no Flow is created;
- **invocation-validation error** — nil context or exact scope/request/binding
  mismatch; zero Owner mutation;
- **managed Flow outcome/error** — returned unchanged by the invoker;
- **panic from validation dependencies or Flow** — not converted to success;
  it propagates through the future orchestrator-owned closure to the existing
  DP-015 panic-safe callback boundary, which leaves the command/phase
  unresolved and expires permit/rendezvous/callback authority;
- **`runtime.Goexit`** — unwinds the synchronous closure; DP-015 deferred
  expiry removes live permit/rendezvous/callback authority and leaves durable
  work unresolved;
- **indeterminate callback return or missing terminal publication** — remains
  indeterminate at DP-015, with no retry, adoption, or fallback.

The invoker does not wrap or relabel an exact downstream Owner outcome/error,
return a `TerminalOutcome`, or infer terminal command truth. It returns exact
`StartOutcome`/error to the future TASK-026 orchestrator-owned closure. That
closure owns mapping to a DP-015 terminal outcome, DP-014 terminal publication,
and DP-015 command/phase terminalization outside DP-021.

## 15. Privacy and Custody Limitation

Repository `internal` visibility and a composition-private constructor are
encapsulation, not an authentication boundary. The invoker cannot prove that a
caller has current DP-015 callback authority because structural binding
validity deliberately carries no live permit proof. The security and
at-most-once proof therefore depends on custody:

- production composition constructs one managed Flow exactly once from the
  scope Binding, audits the same Owner/Loader/Target, then creates the invoker
  with that preconstructed Flow;
- only the future TASK-026 orchestrator-owned callback closure receiving a
  fresh DP-015 binding may call the invoker, exactly once as its sole lifecycle
  subcall;
- no transport, public Directory method, registry, service locator, long-lived
  command record, or recovery reconstruction exposes the invoker;
- reconstruction restores durable records but no callback authority to invoke
  the invoker.

If production composition cannot prove this custody and absence of bypass,
integration fails closed and TASK-026 remains blocked.

## 16. Acceptance Proofs

A later implementation must prove all 18 rows:

1. composition calls `NewManaged` exactly once with the same Binding
   Owner/Loader/Target, returns its error unchanged, and creates no duplicate
   or unmanaged Flow before constructing the invoker;
2. empty domain, invalid Target, or nil preconstructed Flow returns the
   invoker-construction error; success stores only copied domain/Target and one
   borrowed immutable Flow reference with no per-call fact;
3. nil context or any structural domain, Target, request, authorization,
   action, or linked-identity mismatch fails before Owner mutation and performs
   zero Load, Build, Launcher, Host, DP-014, or terminal mutation;
4. `StartExecutionBinding.Valid()` is structural only and is never treated as
   proof of a live permit, rendezvous, callback, generation, or custody;
5. DP-015 authorizes every initial/replay submission before inspection, gives
   only a newly committed claim to the future TASK-026 callback closure, and
   the invoker performs no second policy call;
6. the primitive adapter invokes the future orchestrator-owned closure, whose
   sole lifecycle subcall reaches the stored Flow once through the invoker with
   unchanged request and binding;
7. the linked parent/`StartTarget` adapter invokes the future
   orchestrator-owned phase closure, whose sole lifecycle subcall reaches the
   same invoker/Flow once with unchanged request and binding;
8. the invoker is never the DP-015 callback, never calls Boundary, and returns
   `StartOutcome`/error rather than `TerminalOutcome`;
9. exact Flow `StartOutcome` and error identities return unchanged through the
   invoker to the future orchestrator-owned closure;
10. mapping to `TerminalOutcome`, DP-014 terminal publication, and DP-015
    command/phase terminalization occur only in that closure outside DP-021;
11. an already-cancelled non-nil context reaches Flow and emits exact
    `StartNoClaim`; post-Owner-claim cancellation preserves existing
    continuation and Owner convergence semantics;
12. in-progress, replay, legacy, and reconstructed records receive no callback
    closure and therefore no authorized invoker call;
13. the closure lends the immutable binding for one synchronous lifecycle
    subcall; the invoker never stores, indexes, or exposes it or its rendezvous;
14. callback return, panic, `runtime.Goexit`, or generation loss expires the
    live permit, rendezvous lookup, and callback authority but does not mutate
    or invalidate the structurally valid binding value;
15. retained structurally valid binding reuse is forbidden by callback custody
    and no-bypass composition; the invoker cannot detect liveness or
    distinguish retained from fresh structural values;
16. panic, `runtime.Goexit`, error, and indeterminate outcomes cannot become
    success, duplicate work, retry, adoption, or legacy/unmanaged fallback and
    leave unresolved work under existing DP-015 rules;
17. construction and invocation imports remain acyclic; the invoker creates no
    Flow and directly calls neither DP-015, DP-014, transport, nor orchestrator;
18. no public/private bypass, accessor/token, second policy check, terminal
    owner, or production capability is introduced, and EN/RU mirrors remain
    semantically equal.

## 17. Implementation Boundary

Implementation Status is Partial — implemented in isolation. TASK-043 adds the
concrete `ManagedStartInvoker`, `NewManagedStartInvoker`,
`InvokeManagedStart`, `ErrInvalidManagedStartInvoker`, and
`ErrInvalidManagedStartInvocation` to existing package
`internal/runtimemanagement`. Focused proofs cover constructor and invocation
validation, exact primitive/linked delegation, already-cancelled context
delivery, and unchanged downstream outcome/error identity. The invoker stores
only copied domain/Target and one borrowed preconstructed managed Flow and does
not duplicate Flow construction.

This isolated implementation does not provide the future DP-015 callback
closure, callback custody/replay integration, terminal result mapping, DP-014
terminal publication, DP-015 command/phase terminalization, the DP-016/TASK-026
orchestrator, production composition audit, or production wiring.

TASK-043 is Completed — Coordinator Accepted (2026-08-21) and does not
activate the next task. TASK-044 historically records `UNBLOCK TASK-026`; the
superseding TASK-026 recheck confirms the missing DP-015 tracked-Start
managed-parent plus preclaimed `StopOld` admission prerequisite. TASK-026 is
blocked, the prerequisite is not activated, and no TASK-026 implementation is
asserted here.

## 18. Decision

UWP will use one immutable scope-bound managed Start invoker owned by the
existing internal DP-013 management composition. Composition constructs one
managed Flow exactly once beforehand and passes its borrowed immutable
reference with copied domain/Target into the invoker. The invoker creates zero
Flow, validates exact domain, Target, request, and structural binding, then
synchronously delegates once to that Flow as the sole lifecycle subcall of a
future TASK-026 orchestrator-owned callback closure. Exact `StartOutcome` and
error return to that closure; mapping, terminal publication, and
terminalization remain outside DP-021. The invoker is not the DP-015 callback
and cannot prove live callback authority from a structurally valid binding.
DP-015 retains authorization, command, replay, permit, rendezvous, and callback
authority ownership; callback expiry does not mutate or invalidate the binding
value. Owner retains lifecycle authority; no-reuse rests on closure custody and
no-bypass composition.

There is no second policy check, command owner, lifecycle owner, public route,
detached work, terminal publication, or legacy fallback.

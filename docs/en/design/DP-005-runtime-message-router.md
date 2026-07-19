# DP-005: Runtime Message Router

[Russian version](../../ru/design/DP-005-runtime-message-router.md)

## 1. Status

**Status:** Approved

This proposal defines the minimal deterministic Router for the first Beta message-processing epic. It does not define Delivery, Session targeting, persistence, or a general policy framework.

## 2. Context

The current Runtime composes one `message.Handler` and injects it into every Session. Session converts each WebSocket text or binary frame into an immutable Runtime Message and invokes that Handler synchronously. Echo demonstrates this vertical, but Published Configuration cannot select behavior per message.

[ARCH-001](../architecture/ARCH-001-runtime-architectural-pattern.md) describes routing as `Message Context -> route matchers -> selected Handler or destination -> Handler execution`, while explicitly leaving the Router API, route model, matcher contract, and error policy to a focused DP. The [Master Engineering Plan](../roadmap/MASTER_PLAN.md) places Router after the four Runtime foundation gates and before Session Manager, Delivery, and Persistence.

The implementation already provides the required seam: a transport-neutral Handler invoked by Session and explicitly wired by Runtime composition. It does not yet provide a Runtime Message Context, routing Configuration, or Router.

## 3. Problem

Runtime needs deterministic selection of one configured Handler without allowing transport details, mutable Session state, or Control Plane repositories to enter message processing. The design must preserve the current Echo behavior when routing metadata is absent and must distinguish an intentionally empty routing configuration from an invalid one and from a message that matches no route.

Without a normative model, implementation would have to invent:

- which Session and Principal values are visible to routing;
- whether Handler execution belongs to Router;
- route ordering, overlap, default, and no-match behavior;
- how Handler references are resolved;
- whether a routing outcome terminates Session;
- where Configuration and executability validation occur.

## 4. Goals

- Select exactly one Handler or one explicit no-match outcome for each Runtime Message Context.
- Make selection deterministic for one immutable Published Snapshot and one immutable Context.
- Preserve the existing Handler-based Runtime composition boundary.
- Define the smallest transport-neutral Context that the current Session pipeline can populate.
- Define a backward-compatible, bounded routing Configuration.
- Resolve all Handler references before Runtime becomes Ready.
- Keep selection read-only, synchronous, and safe for concurrent Sessions.
- Preserve `errors.Is` through Router and Session error propagation.

## 5. Non-goals

Router does not:

- deliver messages to clients or choose recipient Sessions;
- manage, register, look up, stop, or retain Sessions;
- own goroutines, lifecycle, queues, retries, backpressure, or acknowledgements;
- persist messages or produce a DLQ;
- implement Groups, Topics, Broadcast, Presence, Delivery, Plugins, Metrics, or rate limiting;
- inspect HTTP requests, HTTP headers, WebSocket frames, connections, or transport addresses;
- parse or match user payload bytes;
- dynamically register Handlers after Runtime startup;
- define a generic matcher, middleware, or policy framework.

## 6. Terminology

- **Runtime Message Context:** immutable per-message envelope created by Session from Runtime-owned values.
- **Route:** immutable configured rule that combines matchers and names one Handler reference.
- **Matcher:** one equality predicate over an approved routable Context field.
- **Compiled Route:** validated Route with its Handler reference resolved to a Handler instance.
- **Legacy Handler:** the Handler already supplied to Runtime composition before Router exists.
- **Default Handler:** optional Handler selected only when no enabled explicit Route matches.
- **No Match:** normal routing outcome in which no Handler is invoked.
- **Selection:** pure ordered evaluation that returns one compiled Handler or No Match.

## 7. Current Architecture Constraints

- Runtime Host remains the only composition root and does not perform route evaluation.
- Session remains the owner of its connection, read loop, lifecycle, and outbound `Sender` capability.
- Router never receives concrete Session, Listener, Handshake, HTTP, WebSocket, or Session Manager values.
- Runtime Message payload remains owned by immutable `message.Message`.
- Handler execution remains synchronous in the Session execution path.
- A Handler error currently terminates `Session.Run`; Router must preserve that contract rather than hide it.
- Published Configuration is copied through `runtimeconfig.Builder`; Router reads no repositories.
- Runtime executability validation completes before Listener socket acquisition.
- Session Manager identity and `RegistrationView` are not Router targeting contracts.

## 8. Runtime Message Context

The design introduces a new immutable Runtime Message Context; it does not add fields to `message.Message`.

Conceptually the Context contains:

| Field | Source | Routable in the first version | Contract |
|---|---|---:|---|
| Runtime Message | Session read loop | Message type only | Existing immutable text/binary Message; payload and receive time are unchanged |
| Sender | Current Session | No | Transport-neutral outbound capability available only to the selected Handler |
| Session ID | Current Session | No | Opaque copied value for Handler correlation; not a Session lookup or targeting capability |
| Authenticated flag | Immutable Session Principal | Through principal-kind | Exactly one of authenticated or anonymous is true |
| Anonymous flag | Immutable Session Principal | Through principal-kind | Explicit anonymous identity remains distinguishable |
| Authentication type | Immutable Session Principal | Yes | Provider type such as `api-key` or `jwt`; absent for anonymous identity |
| Authentication provider | Immutable Session Principal | Yes | Safe configured Provider name; absent for anonymous identity |

The initial Context contains no HTTP headers, query values, remote address, WebSocket value, request, connection, raw Session, Principal ID or name, Principal maps, claims, roles, attributes, metadata, or arbitrary user metadata. The Principal is represented only by the two mutually exclusive kind flags and the safe Authentication type and Provider identifiers. Those values cannot be expanded merely by exposing data already retained elsewhere; each future identity or routable category requires focused Configuration and security review.

Service metadata and user content remain separate: Context scalar fields are generated by Runtime components, while payload bytes remain only inside Runtime Message. Matchers never read payload bytes.

Construction copies all scalar values. Accessors return values, Runtime Message preserves its existing payload-copy contract, and the Context exposes no mutable maps or slices. Sender is a stable execution capability, not routable data; Router does not invoke it.

Context construction succeeds only for a valid text or binary Message, a non-nil Sender, a non-empty opaque Session ID, and exactly one Principal-kind flag. An authenticated Context requires both Authentication type and Provider; an anonymous Context requires both to be absent. Failure to satisfy these invariants is an internal construction error, never a routable no-match value.

The Handler contract evolves coherently to receive the Runtime Message Context with the standard cancellation `context.Context`. Session constructs exactly one Runtime Message Context per accepted Message. Echo reads Message and Sender from that Context and otherwise preserves its current behavior. There is no dual legacy/context Handler dispatch and no use of `context.Value`.

## 9. Routing Configuration Model

`ConfigurationVersion` gains an optional Routing section. Optionality is represented explicitly so absence differs from a present empty object.

The conceptual model is:

```text
Routing
    Routes []Route
    DefaultHandlerRef optional string

Route
    ID string
    Enabled bool
    Priority uint32
    Matchers []Matcher
    HandlerRef string

Matcher
    Type MatcherType
    Value string
```

The first implementation uses these bounds:

- at most 256 Routes;
- at most one Matcher of each supported type per Route;
- therefore at most four Matchers per Route;
- Route ID and Handler reference are trimmed, 1 to 128 ASCII characters, and match `[A-Za-z][A-Za-z0-9._-]*`;
- every Route Priority is positive and unique across all Routes, including disabled Routes;
- Route ID is unique across all Routes;
- every Route, including a disabled Route, must be structurally valid;
- enabled Routes require a non-empty Matcher list and an active resolvable Handler reference;
- disabled Routes are not compiled or evaluated; their syntactically valid Handler reference need not resolve until a later Published Configuration enables the Route;
- `DefaultHandlerRef`, when present, must be syntactically valid and resolve before readiness;
- two enabled Routes with identical normalized Matcher sets are invalid regardless of Priority, because the Route with lower precedence, evaluated later in ascending Priority order, could never be selected;
- unsupported Matcher types are invalid, including on disabled Routes, because this schema has no opaque extension preservation contract.

An empty Matcher list is never shorthand for a catch-all Route. Catch-all behavior is represented only by `DefaultHandlerRef`.

Duplicate-set detection includes enabled Routes only. Disabled Routes still undergo all per-Route structural validation, including Matcher type, value, and one-Matcher-per-type rules, but are excluded from comparisons with enabled or other disabled Routes. Enabling one later requires validation of the complete new Configuration, at which point it participates in duplicate-set detection.

Configuration ownership is separated:

- `ConfigurationVersion` owns editable declarative Routing metadata and Control Plane validation;
- `runtimeconfig.Snapshot` owns a deep copy of Published Routing metadata;
- Router owns a compiled immutable table containing only enabled Routes, normalized matchers, Route identity, and resolved Handler values.

Router retains neither the mutable ConfigurationVersion nor the temporary Handler-binding input used during composition.

## 10. Matcher Semantics

The initial Matcher set is closed and contains four types.

| Matcher | Values | Equality | Missing Context value |
|---|---|---|---|
| `message-type` | `text`, `binary` | Exact, case-sensitive enum equality | Impossible for a valid Runtime Message |
| `principal-kind` | `authenticated`, `anonymous` | Exact, case-sensitive enum equality | Invalid Context is an internal failure, not a non-match |
| `authentication-type` | Supported Authentication Provider type | Exact, case-sensitive enum equality | Non-match, including anonymous Principal |
| `authentication-provider` | Configured Provider name | Exact, case-sensitive string equality | Non-match, including anonymous Principal |

Matcher normalization is one exact algorithm:

1. Type and Value must both be present. Missing and empty Configuration values are invalid rather than equivalent to a wildcard.
2. Leading and trailing Unicode whitespace, as defined by Go `strings.TrimSpace`, is removed from Type and Value. Internal whitespace is preserved and never collapsed.
3. The trimmed Type must exactly equal one of the four canonical lowercase type tokens in the table. It is never lowercased or case-folded.
4. Enum values must exactly equal their canonical lowercase token: `text` or `binary`, `authenticated` or `anonymous`, and `jwt`, `api-key`, or `basic` for Authentication type. Values are never lowercased or case-folded.
5. An Authentication Provider value preserves its trimmed spelling and case and compares exactly with the normalized configured Provider name.
6. A Type or Value empty after trimming is invalid. At message-evaluation time, an absent optional Authentication value is distinct from an empty configured Matcher value: absence yields non-match as specified in the table, while an empty configured value can never pass validation.
7. A normalized Matcher set is order-independent: it is identified by the sorted sequence of normalized `(Type, Value)` pairs. Reordering Matchers therefore cannot evade duplicate-set detection.

Control Plane applies this algorithm before accepting and publishing Routing metadata. `runtimeconfig.Builder` applies the same algorithm defensively while producing its deep copy. Router compilation applies it again into the private compiled copy without mutating Snapshot. All three layers accept and reject the same raw Type and Value inputs and produce byte-for-byte equivalent canonical values; none performs additional case conversion.

Multiple Matchers in one Route use logical AND. OR is represented by separate Routes. A missing optional Context value produces false only for the two Authentication metadata matchers.

Regex, glob, prefix, payload, claim, role, arbitrary attribute, HTTP header, query, remote-address, and Session-ID matchers are not supported in this version. An active Configuration using them is rejected before readiness.

## 11. Selection Algorithm

Composition performs these steps once:

1. Validate Routing metadata.
2. Remove disabled Routes from the runtime table.
3. Resolve each enabled Route Handler reference and optional Default Handler reference.
4. Normalize Matcher values.
5. Sort compiled Routes by ascending Priority.
6. Publish the immutable compiled table only as part of successful Runtime composition.

For each Runtime Message Context, Router:

1. returns cancellation if the call context is already canceled;
2. evaluates compiled Routes in ascending Priority;
3. selects the first Route for which every Matcher is true;
4. if none match, selects the resolved Default Handler when configured;
5. otherwise returns No Match.

Exactly one Handler or No Match is produced. Router never invokes a second Handler. Multiple matching Routes with different priorities are intentionally resolved by the first priority. Equal priorities are rejected at startup. Two enabled Routes with identical normalized Matcher sets are also rejected regardless of Priority: the Route with lower precedence, evaluated later in ascending Priority order, would otherwise be unreachable. Disabled Routes do not participate in this duplicate comparison. Consequently runtime ambiguity is not a supported state.

The same compiled table and equal Context values always produce the same selection.

## 12. Backward Compatibility

The four cases are distinct:

| Configuration state | Runtime behavior |
|---|---|
| Routing section absent | Composition installs one implicit legacy default with HandlerRef `legacy`, bound to the existing injected Handler; current Echo vertical is unchanged |
| Routing section present and empty | Valid intentional reject-all configuration; every Message produces No Match, Router returns nil, and Session continues |
| Routing section present but invalid | Runtime startup fails before Listener socket acquisition |
| Valid Routing present but no Route matches and no default exists | No Match; no Handler invocation, Router returns nil, and Session continues without fallback |

The implicit compatibility routing entry exists only when Routing is absent. It is represented as a Default Handler reference to `legacy`, not as an ordinary configured Route. Runtime does not silently add it to an explicitly present Routing section.

## 13. Handler Execution Model

Router implements the transport-neutral `message.Handler` role and invokes the selected Handler synchronously. This is model B from the considered choices.

The model is selected because Runtime already composes one Handler into Session. Substituting Router at that seam avoids adding a second dispatcher to Session and keeps route selection outside Session. Returning a Handler to Session would make Session know Router selection outcomes and duplicate execution/error plumbing.

Handler construction and reference resolution occur during Runtime composition. The initial composition-local Handler registry contains exactly one supported reference: `legacy`. Runtime composition binds the existing injected Handler instance to that name before Router compilation. When Routing is absent, the implicit compatibility default refers to the same `legacy` binding; it does not create another Handler.

The registry is a finite construction input, not a dynamic registry, and Router does not retain it after compilation. Every enabled Route and explicit `DefaultHandlerRef` must resolve in it before Runtime readiness and before Listener construction or socket acquisition. Therefore, in the initial implementation, their only accepted resolvable value is `legacy`; any other active reference causes startup failure in the unresolved-Handler category. A syntactically valid reference on a disabled Route is not resolved until a later Configuration enables that Route.

Compiled Routes may share one Handler instance. Each Handler implementation owns and documents its concurrency safety. Router adds no serialization and holds no lock while calling a Handler.

For a selected route, Router invokes exactly one Handler exactly once with the original cancellation context and immutable Runtime Message Context. No Match invokes no Handler and maps to `nil` from the existing `message.Handler` return contract. Session therefore continues its read loop. Once an explicit Routing section has been selected for composition, No Match never falls back to the legacy Handler.

After invocation starts, Router returns the selected Handler result and does not replace it with a separately observed cancellation outcome. The selected Handler owns observation of the same passed context during its execution. Router performs no fallback selection and never invokes another Handler because cancellation races with that call.

## 14. Error Model

The implementation requires stable categories compatible with `errors.Is`.

| Outcome | Category | Runtime/Session effect |
|---|---|---|
| Invalid Routing shape, duplicate identity/priority, forbidden empty matcher, or static ambiguity | Invalid routing configuration | Prevents Runtime readiness before Listener startup |
| Unsupported Matcher | Unsupported routing capability | Prevents Runtime readiness before Listener startup |
| Active unresolved Handler reference | Unresolved Handler | Prevents Runtime readiness before Listener startup |
| No explicit match and no default | No Match, represented by `nil` from `message.Handler` | No Handler call, no legacy fallback; current Session read loop continues |
| Runtime ambiguity | Unreachable after successful validation | Treated as internal failure if an invariant defect makes it observable |
| Selected Handler returns error | Handler execution error preserving the cause | Returned through Router; current `Session.Run` terminates and preserves `errors.Is` through wrapping |
| Call context canceled before selection | `context.Canceled` or `context.DeadlineExceeded` | No Handler call; returned to Session execution path |
| Invalid Runtime Message Context | Internal Router failure | Terminates current Session execution path; never treated as No Match |

Router does not produce a protocol response, acknowledgement, retry, alternative route, or Delivery action for any outcome. Error text may contain a safe Route ID or Handler reference but never payload, Principal identity, credentials, claims, or Secrets.

## 15. Concurrency and Lifecycle

Router has no lifecycle states and no `Start` or `Stop`. Runtime composition constructs it before Listener startup and releases it with the Runtime component graph.

The compiled route slice, Matcher values, Route IDs, Handler references, and resolved Handler values never change after construction. Selection performs read-only iteration and requires no Router mutex. Concurrent calls from different Sessions are safe and share no per-message mutable state.

Router starts no goroutines, owns no queues, and holds no lock during Handler execution. Each call receives its own Runtime Message Context and cancellation context; values from one Session cannot be reused or written into another Context. Handler implementations remain responsible for their own concurrent-call contract.

## 16. Startup Validation

Control Plane validation owns declarative shape and normalization:

- section presence;
- collection bounds;
- Route and Handler-reference syntax;
- Route ID and Priority uniqueness;
- supported Matcher names and value domains;
- one Matcher per type per Route;
- non-empty Matcher list for enabled Routes;
- exact duplicate enabled Matcher sets;
- the single optional Default Handler field; the scalar schema cannot express multiple defaults.

`runtimeconfig.Builder` is the defensive publication boundary. It refuses a Published ConfigurationVersion that violates the structural Routing invariants above and deep-copies only a valid representation. This does not transfer declarative-validation ownership from Control Plane: it prevents malformed programmatic input from becoming a trusted Runtime Snapshot.

Runtime executability validation owns:

- support for the Routing Snapshot version and all active Matcher types;
- resolution of every enabled Route Handler and configured Default Handler;
- compilation without mutation of Snapshot;
- installation of the implicit legacy default only when Routing is absent;
- proof that the compiled Router can be published before Listener startup.

Validation occurs in Runtime composition before Listener construction and socket acquisition. Routing validation participates in the existing startup capability model: active invalid or unsupported routing prevents Ready; disabled Routes are not executable but still pass Control Plane structural validation. Snapshot and compiled-table deep-copy invariants are verified independently.

Validation rules have one normative definition in the DP. Control Plane and `runtimeconfig.Builder` tests must use shared cases or mirrored proof tables so defensive validation cannot silently diverge. Runtime composition does not repeat repository or editable-state checks; it validates Snapshot support and resolves active runtime capabilities.

## 17. Package Boundaries

Expected dependency direction:

```text
ConfigurationVersion
    -> runtimeconfig.Builder
    -> immutable Routing Snapshot
    -> Runtime composition and Handler bindings
    -> Router compiled table

Session
    -> immutable Runtime Message Context
    -> Router as message.Handler
    -> selected message.Handler
```

The new `internal/router` package may depend on:

- `internal/message` transport-neutral Message, Context, Sender, and Handler contracts;
- immutable routing snapshot or a Router-local construction input;
- Go standard-library context and error packages.

Forbidden imports and dependencies are:

- `net/http`;
- any WebSocket implementation package;
- `internal/listener`;
- concrete `internal/session` types;
- `internal/sessionmanager` and its Registration or Lookup internals;
- `internal/handshake` and `internal/connection`;
- Delivery, Persistence, repositories, Control Service handlers, logging implementations, and plugin infrastructure.

Runtime composition may depend on Router to wire the graph. Router never depends on Runtime Host or Container.

## 18. Security Considerations

- Router sees no credentials, Authorization headers, Secret References, request values, or resolved Secrets.
- Payload bytes are not matcher input and are never included in Router errors.
- Principal claims, roles, attributes, metadata, ID, and name are not routable in the first version; only explicit authentication classification metadata is routable.
- Authentication Provider names are treated as safe Configuration identifiers, not credential material.
- Route IDs and Handler references have bounded length and restricted syntax.
- Route and Matcher counts are bounded to prevent unbounded per-message evaluation from Configuration.
- Router does not weaken Authentication, admission, Session ownership, or Runtime cancellation.

## 19. Observability Considerations

This epic introduces no logging, Metrics, tracing framework, or event bus. Stable returned error categories allow the existing terminal-error path to observe selected Handler failures. Safe diagnostics may identify Route ID and Handler reference, but must not include Message payload or Principal identity values.

No Match is a normal outcome and is not reported as a terminal error. Future Metrics may count bounded Route IDs only after the Metrics epic defines cardinality and ownership. This DP does not require that instrumentation.

## 20. Test Strategy

Implementation proof must include:

- every initial Matcher type, equality rule, case rule, and missing-value rule;
- logical AND within one Route;
- ascending Priority, overlapping Routes, and first-match selection;
- duplicate ID, duplicate Priority, duplicate Matcher type, and identical normalized Matcher-set rejection across enabled Routes regardless of Priority;
- disabled Route exclusion from selection and duplicate-set comparison while retaining its structural validation;
- whitespace, missing/empty, canonical enum, case-preservation, order-independent set, and cross-layer normalization equivalence;
- absent Routing implicit legacy behavior;
- present empty Routing reject-all behavior;
- configured Default Handler;
- No Match as `nil`, without Handler invocation, legacy fallback, or Session termination;
- the sole initial `legacy` Handler binding, implicit compatibility reference, unknown active-reference rejection, disabled-reference deferral, and shared Handler instances;
- text and binary Message payload preservation;
- exactly one selected Handler invocation;
- selected Handler error and `errors.Is` preservation through Session;
- cancellation before selection and during Handler execution;
- concurrent routing through one immutable compiled table;
- isolation of Context values belonging to different Sessions;
- Context and Snapshot deep-copy behavior;
- unresolved Handler and unsupported Matcher rejection before readiness;
- real `ConfigurationVersion -> Published -> runtimeconfig.Builder -> Runtime Host -> WebSocket -> Session -> Router -> Handler` flow;
- compatibility of the current Echo vertical when Routing is absent;
- update of the exhaustive Runtime Snapshot support matrix;
- package/import-boundary checks where repository tooling supports them.

Tests use channels, barriers, injected Handlers, and real integration clients where appropriate. Arbitrary sleeps are not synchronization, and every spawned goroutine has a bounded cleanup path.

## 21. Alternatives Considered

### Router returns a Handler for Session to execute

Rejected. It makes Session aware of Router decisions and adds selection-specific branching to the component that owns transport execution.

### Extend `message.Message` with Session and Principal fields

Rejected. Runtime Message remains application data; execution and identity metadata belong in a separate immutable Context.

### Store Context values in `context.Context`

Rejected. `context.Value` would create hidden data dependencies and weaken compile-time contracts.

### Route on payload, claims, arbitrary metadata, or HTTP values

Rejected for the first version. It expands security, normalization, compatibility, and matcher semantics without a proven use case.

### Invoke every matching Handler

Rejected. Multi-handler execution is Delivery or fan-out behavior and requires ordering, partial-failure, and backpressure semantics outside this epic.

### Resolve priority ties by Route ID or declaration order

Rejected. Startup rejection is clearer than a subtle tie-break rule and prevents accidental behavior changes during Configuration editing.

### Dynamic Handler registry

Rejected. Runtime composition resolves a finite explicit set before readiness; dynamic registration introduces lifecycle and synchronization not required by Router.

### Treat present empty Routing as legacy behavior

Rejected. It would erase the semantic distinction between absent configuration and an explicit operator choice to route nothing.

## 22. Consequences

Positive consequences:

- Router fits the existing Handler seam and does not alter transport ownership.
- Selection is deterministic, bounded, lock-free, and configuration-first.
- Current deployments retain Echo behavior when Routing is absent.
- Invalid routes and unresolved Handlers fail before traffic.
- Router remains isolated from future Session Manager and Delivery contracts.

Costs and limitations:

- Handler input must migrate coherently to Runtime Message Context.
- Initial routing inputs are intentionally narrow.
- The first production composition exposes only the existing Handler as `legacy` until additional explicit Handler bindings are implemented.
- No Match silently performs no application action because no message-level acknowledgement protocol exists.
- Handler concurrency remains a separate implementation contract.

## 23. Implementation Stages

1. Introduce immutable Runtime Message Context and migrate Handler, Echo, and Session tests without changing behavior.
2. Add optional Routing metadata and Control Plane validation to ConfigurationVersion.
3. Deep-copy Routing into Runtime Snapshot and extend support-matrix coverage.
4. Implement Router compilation, closed Matcher set, deterministic selection, and error categories.
5. Bind the existing Handler as `legacy` and compose Router before Listener startup.
6. Add unit, concurrency, startup-failure, and full WebSocket integration proofs.
7. Update factual current-state and changelog documentation only after implementation passes verification.

Each stage preserves a buildable vertical and does not introduce Delivery, Session Manager integration, or dynamic Handler registration.

## 24. Open Questions

None. All normative decisions required for the first Router implementation are closed by this proposal. Additional matchers, Handler bindings, routing outcomes, or targeting semantics require a focused revision or a separate design after concrete use cases exist.

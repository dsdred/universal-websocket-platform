# DP-005 Runtime Integration Contract Review

[Russian version](../../ru/reviews/dp-005-runtime-integration-contract-review.md)

**Status:** Completed

**Review date:** 2026-07-20

**Assessment:** Approved to continue Runtime integration

## 1. Scope

This review records the resolution of the construction-boundary blocker found before DP-005 Runtime integration. It reviews only the contract between Runtime composition, the strict Router compiler, the absent-Routing compatibility path, and the existing injected legacy Handler.

No Go implementation is assessed by this document. Router selection semantics, Matcher behavior, Runtime lifecycle, Snapshot structure, and downstream Handler execution remain governed by [DP-005](../design/DP-005-runtime-message-router.md) and are unchanged except for the clarified construction contract.

## 2. Blocker Found Before Implementation

DP-005 required absent Routing to install an implicit compatibility Router backed by the existing injected legacy Handler. The implemented strict compiler requires a non-nil valid `RoutingSnapshot`, while Runtime also permits the injected Handler to be nil as an intentional discard configuration.

The existing public contracts therefore did not determine how Runtime could construct a Router for absent Routing without either changing strict compiler semantics, bypassing Router, or inventing nil-Handler behavior during implementation. Runtime integration stopped before code changes because each interpretation would have made an unapproved architectural decision.

## 3. Rejected Interpretations

### Change the Strict Compiler to Accept Nil

Rejected. Nil does not describe an explicit Routing configuration and cannot satisfy the strict compiler's validation contract. Treating `router.New(nil, ...)` as compatibility construction would combine declarative compilation with absent-configuration behavior and would change the established meaning of invalid compiler input.

The strict compiler continues to require a valid `RoutingSnapshot` and continues to reject nil Routing input.

### Bypass Router When Routing Is Absent

Rejected. Passing the legacy Handler directly to Session would create two message execution paths and make Runtime configuration presence control message-time composition. It would also prevent the invariant that every post-startup Message traverses `Router.Handle`.

Runtime must not inspect routes or bypass Router after startup.

### Install a Synthetic Discard Handler

Rejected. A synthetic no-op Handler would introduce an execution object that is not required to preserve existing nil-Handler behavior. No Match already provides the required non-terminal `nil` result without Handler invocation.

## 4. Approved Compatibility Construction Contract

Router exposes a separate explicit compatibility construction path. Its conceptual responsibility is equivalent to:

```text
Compatibility Router construction
    input: injected legacy Handler
    compiled Routes: zero
    Default Handler: injected Handler when non-nil; absent when nil
    output: one immutable Router
```

This factory represents only absent Routing. It does not accept configured Routes, normalize Matchers, resolve a registry, or change the strict compiler. Exact Go naming remains an implementation-level choice only when it follows existing repository conventions and preserves this single responsibility.

## 5. Nil Legacy Handler Behavior

When Routing is absent and the injected legacy Handler is nil:

- Runtime startup succeeds;
- the compatibility Router contains zero compiled Routes;
- the compatibility Router has no Default Handler;
- each Message produces No Match;
- `Router.Handle` returns nil;
- Session continues its read loop;
- no synthetic Handler is created or invoked.

This preserves the existing discard behavior while keeping Router on the message path.

## 6. Runtime Construction and Message Path

Runtime performs exactly one construction branch during startup:

```text
Routing present
    -> strict Router compilation from RoutingSnapshot

Routing absent
    -> compatibility Router construction from injected legacy Handler

Either branch
    -> store one immutable Router
    -> inject Router as message.Handler
    -> reuse Router.Handle for every Message
```

Construction completes before Listener construction and socket acquisition. Router compilation, sorting, normalization, and Handler resolution do not occur during message processing. The resulting Router is immutable and requires no Runtime-owned synchronization on the hot path.

## 7. Resolved Findings

| Finding | Resolution |
|---|---|
| No absent-Routing construction contract | Resolved by the separate compatibility factory |
| Potential change to `router.New(nil, ...)` | Rejected; strict compiler semantics remain unchanged |
| Potential Runtime bypass of Router | Rejected; both startup branches publish Router |
| Nil injected legacy Handler unspecified | Resolved as compatibility Router without Default Handler and normal No Match |
| Construction count and reuse unspecified | Resolved as exactly one startup construction and reuse of one immutable Router |

No unrelated DP-005 finding is claimed as resolved by this clarification.

## 8. Approval Assessment

**Verdict: Approved to continue Runtime integration.**

The construction boundary is now unambiguous. Implementation may add the narrow compatibility factory, select one Router construction path during Runtime startup, and inject the resulting Router at the existing `message.Handler` seam. It must preserve strict compiler behavior, nil discard compatibility, existing Runtime lifecycle, and the single post-startup `Router.Handle` path.

Implementation remains subject to independent code and test review. This approval does not claim that Runtime integration is already implemented.

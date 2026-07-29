# DP-012: Runtime Source Composition

[Russian version](../../ru/design/DP-012-runtime-source-composition.md)

## 1. Status

- **Design Status:** Draft
- **Implementation Status:** Implemented in isolation

This proposal is non-normative until approved. It defines a
repository-backed `configurationloader.Source`. The adapter is implemented
and tested in isolation; management composition and Production Activation do
not exist.

## 2. Purpose

Define the smallest concrete Source composition that adapts the existing
in-memory Configuration and ConfigurationVersion repositories to
`configurationloader.Source`, so a future application composition root can
construct `Source -> Loader -> Flow` without starting a Runtime.

## 3. Authority

This proposal refines, without overriding:

- [ADR-0002](../adr/0002-configuration-dsl.md);
- [ADR-0003](../adr/0003-runtime-architecture.md);
- [ARCH-004](../architecture/ARCH-004-runtime-deployment-and-identity-model.md);
- [ARCH-005](../architecture/ARCH-005-runtime-configuration-snapshot-and-loading-model.md);
- [DP-007](DP-007-configuration-loader-contract.md) through
  [DP-011](DP-011-runtime-launch-pipeline-integration.md).

Approved ADR and Active architecture remain authoritative.

## 4. Scope

The design covers one in-process adapter over the existing concrete in-memory
repositories, exact lookup, consistency confinement, schema facts,
detachment, error classification, construction, dependency direction,
concurrency, lifetime, and future implementation proofs.

## 5. Package and responsibility

The planned package is `internal/configurationloadsource`.

It owns:

- exact source addressing and the confined consistency proof;
- static source representation facts;
- defensive detachment;
- classification of repository outcomes into existing Loader source errors.

It does not own Loader validation, Builder semantic validation, Flow
orchestration, lifecycle, routing, or activation.

## 6. Exact planned API

```go
type MemorySource struct {
    // private borrowed dependencies
}

func NewMemorySource(
    configurations *configuration.MemoryConfigurationRepository,
    versions *configurationversion.MemoryConfigurationVersionRepository,
) *MemorySource

func (s *MemorySource) LoadExact(
    workspaceID uint64,
    configurationID uint64,
    configurationVersionID uint64,
) (configurationloader.SourceObservation, error)
```

`*MemorySource` implements the existing `configurationloader.Source`.
No new error, interface, option, context, or lifecycle declaration is added.

## 7. Constructor

`NewMemorySource` only stores borrowed repository references. It performs no
read, write, goroutine creation, lock acquisition, validation, cache
initialization, or lifecycle action. A nil receiver or either nil dependency
causes `LoadExact` to return `configurationloader.ErrSourceUnavailable`.

## 8. Exact lookup algorithm

For one call, after validating the binding, Source must:

1. call `versions.Get(configurationVersionID)` exactly once;
2. normalize a returned error;
3. verify returned Version ID and Configuration ID;
4. reject known non-Published states, unknown state, or zero Number;
5. retain a detached exact Version value;
6. call `configurations.Get(configurationID)` exactly once;
7. normalize a returned error;
8. verify returned Configuration ID and Workspace ID;
9. return a fresh complete observation.

The requested Workspace ID is not used to select another entity; it is an
identity assertion over the exact parent.

## 9. Prohibited selection

Source must not call `GetPublished`, list versions, select latest/current,
retry, re-read, replace the requested identity, or fall back to another
ConfigurationVersion. A failure of the exact chain is terminal for that load.

## 10. Published validation

`Draft`, `Validated`, and `Archived` map to
`configurationloader.ErrVersionNotPublished`. An unknown lifecycle state or
zero Version Number maps to `configurationloader.ErrSourceIntegrity`.
Only the exact observed `Published` Version is eligible.

## 11. Consistency linearization

The successful Version `Get` is linearization point **L**. The later parent
Configuration `Get`, **C**, proves that the parent existed at L only under the
single-instance mutation topology and immutability invariants in the next
section. The single Version Service serializes lifecycle operations around L;
the single Configuration Service cannot change parent identity or resurrect a
deleted parent. Two independent repository locks do not by themselves form a
cross-repository snapshot.

## 12. Mandatory composition confinement

One application composition root must own one repository pair and construct
exactly one `*configuration.Service` and exactly one
`*configurationversion.Service` over that pair. These two instances are the
sole mutation authorities. Handlers receive only these exact Service
instances; repository references are not exposed.

After bootstrap/setup there must be no direct repository
`Create`/`Update`/`UpdateBatch`/`Delete`, second Service instance, importer,
migration, or alternate writer. MemorySource performs only `Get`.

The single ConfigurationVersion Service uses its one `lifecycleMu` to
serialize all Create, update, publish, and archive operations. This excludes a
stale Draft update overwriting a Published value. The single Configuration
Service has no additional mutex, but its Update changes only Name,
Description, and UpdatedAt: ID and Workspace ID remain immutable. An Update
racing Delete cannot resurrect the parent because repository Update returns
Not Found after deletion.

The composition must preserve all of these facts:

- IDs are monotonic and never reused;
- a parent Configuration is created before its Versions;
- Configuration ID and Workspace ID are immutable;
- Version ID, Configuration ID, Number, and Published payload are immutable;
- each record `Get` and lifecycle transition is atomic;
- each `Get` returns a detached value.

These facts are established by a mandatory Composition Audit before
MemorySource construction and again before Production Activation. The
constructor cannot introspect or prove topology. If the Audit cannot prove the
topology, or finds a second Service/direct writer, Source is not constructed
and activation is blocked. A violation introduced after construction is a
programming/composition contract violation, not a detectable runtime Source
failure. Neither two locks nor retry can manufacture the required snapshot.

## 13. Parent races

If the parent is deleted before C, Source conservatively returns Not Found.
If deletion happens after C, success remains a truthful observation at L
under confinement.

## 14. Version lifecycle races

Archive before L returns Version Not Published. Archive after L does not
change the detached exact Published Version already observed at L, so the load
may succeed under confinement.

## 15. Source observation

Success returns a fresh `configurationloader.SourceObservation` with:

| Field | Value |
|---|---|
| `WorkspaceID` | requested and verified Workspace ID |
| `Configuration` | detached exact parent Configuration |
| `ConfigurationVersion` | detached exact Published Version |
| `SchemaIdentity` | literal `uwp.configuration` |
| `SchemaVersion` | literal `1` |
| `RepresentationComplete` | `true` |

Schema facts are adapter-owned static representation facts. The adapter does
not import `runtimeconfig`, negotiate schemas, or perform semantic validation.

## 16. Detachment

The observation must share no mutable logical content with either repository
or caller. Defense-in-depth copies all nested Authentication provider
collections, API Key/Basic/JWT pointers, JWT slices, Routing, Routes, and
Matchers. Mutating repository input after the call cannot change the
observation, and mutating a returned observation cannot change repository
state or a later observation.

## 17. Failure matrix

| Condition | Exact returned identity |
|---|---|
| nil receiver or dependency | `ErrSourceUnavailable` |
| Version `Get` not found | `ErrSourceNotFound` |
| Configuration `Get` not found | `ErrSourceNotFound` |
| any other `Get` error | `ErrSourceUnavailable` |
| returned Version ID or Configuration ID mismatch | `ErrIdentityMismatch` |
| returned parent ID or Workspace ID mismatch | `ErrIdentityMismatch` |
| exact `Draft`, `Validated`, or `Archived` Version | `ErrVersionNotPublished` |
| unknown state or zero Version Number | `ErrSourceIntegrity` |
| incomplete static schema or representation facts | `ErrSourceIntegrity` |

Mappings use `errors.Is`. Source exposes no raw repository error, wrapped
detail, configuration data, or transport diagnostic.

`ErrInconsistentSourceObservation` remains part of the generic Loader contract
for other Source implementations. MemorySource never synthesizes it. Its
runtime outcomes are only Not Found, Identity Mismatch, Version Not Published,
Source Integrity, Source Unavailable, or success.

## 18. Defense-in-depth validation

Source classifies repository and representation facts. Loader remains
responsible for independently revalidating request completeness, ownership
chain, lifecycle state, schema completeness, and the closed failure contract.
Source validation does not weaken or replace Loader validation.

## 19. Dependency graph

```text
future cmd composition root
    ├── configuration.MemoryConfigurationRepository
    ├── configurationversion.MemoryConfigurationVersionRepository
    └── configurationloadsource.MemorySource
            │ implements
            ▼
        configurationloader.Source
            ▼
        configurationloader.Loader
            ▼
        runtimelaunchflow.Flow

Flow.Start / Runtime Host are not invoked by construction.
```

The adapter may depend on `configuration`, `configurationversion`, and
`configurationloader`. Loader depends only on its Source boundary. Runtime
configuration, lifecycle, Flow, and Runtime packages must not depend on the
adapter. No dependency cycle is permitted.

## 20. Ownership

Repositories remain owned by the application composition root. MemorySource
borrows them and owns no entity or lifecycle. Loader borrows Source; Flow
borrows Loader and Owner. Constructing this chain transfers no resource
ownership and performs no activation.

## 21. Lifetime

Both repositories must outlive MemorySource; MemorySource must outlive every
Loader that borrows it; the Loader must outlive every Flow that borrows it.
Shutdown ordering belongs to future application composition and introduces no
adapter hook.

## 22. Concurrency

MemorySource is stateless after construction and safe for concurrent calls
because it adds no mutable state and relies on the existing concurrent-safe,
detaching repository reads. It adds no mutex, cache, goroutine, retry, or
background work.

## 23. Future composition

A future `cmd` composition root may construct the repositories, services,
MemorySource, Loader, Owner, and Flow. This proposal authorizes only the
dependency chain; it does not authorize routing a management request, calling
`Flow.Start`, publishing a Host, or otherwise activating Runtime.

## 24. Implementation proofs

TASK-014 provides local proof of the applicable implementation contract:

1. compile-time Source interface conformance;
2. zero constructor side effects and nil-binding behavior;
3. exactly one Version `Get` and at most one Configuration `Get`;
4. no list, `GetPublished`, fallback, retry, or re-read;
5. exhaustive identity, state, schema, and error mapping;
6. deep two-way detachment;
7. equivalent repeated and concurrent loads;
8. publish/archive/delete race outcomes at L and C;
9. exactly one Configuration Service and one ConfigurationVersion Service over
   the same repository pair;
10. rejection by Composition Audit of two Version Service instances, any
    direct writer, importer, or migration;
11. regression proof that a stale Draft cannot overwrite Published through
    the single Version Service;
12. single-Service update/publish serialization;
13. Configuration update/delete identity and non-resurrection invariants;
14. Composition Audit completion before Source construction and activation;
15. no test expecting MemorySource to return
    `ErrInconsistentSourceObservation`;
16. Loader integration over the real adapter;
17. isolated `Source -> Loader -> Flow` construction without Start or Host;
18. absence of detector, registry, global lock, retry, repository extension,
    dependency cycle, cache, or goroutine;
19. targeted tests, stress, and race detector when technically available.

The implementation includes the exact concrete constructor and private
getter-function test seam, Version-first lookup, closed error mapping, static
schema facts, two-way deep detachment including non-nil empty collections,
concurrent/repeated loads, Loader integration, and isolated
`Source -> Loader -> Flow` construction without Start or Host. Service
regressions cover stale Draft protection and Configuration identity/delete
invariants.

Race execution is technically unavailable with `CGO_ENABLED=0`, and enabling
CGO fails because `gcc` is absent; targeted concurrency stress is the recorded
substitute. Composition Audit remains static/manual future application
evidence, not adapter runtime behavior.

## 25. Activation gate

Before Source construction and again before Production Activation, a
Composition Audit must prove the exact reference graph: one repository pair,
one Service instance of each type, handler references to only those instances,
MemorySource read-only access, service-only mutation, and absence of direct or
alternate writers.

If any item is unproven, Source is not constructed and activation is blocked.
The Audit is composition evidence, not a runtime detector, registry, global
lock, retry loop, or repository extension.

## 26. Non-goals

- management HTTP routing, authentication, or authorization;
- persistence or Runtime identity storage;
- Production Activation;
- recovery, reconciliation, retry, supervision, or caching;
- diagnostics transport;
- schema migration or negotiation;
- repository interfaces, extensions, or redesign;
- lifecycle hooks or background work.

## 27. Implementation boundary

Implementation Status is Implemented in isolation. Package
`internal/configurationloadsource`, the exact `MemorySource` constructor and
`LoadExact`, compile assertion, deep detachment, local tests, Loader
integration, and construction proof exist. Application/Control Service wiring,
management routing, `Flow.Start`, Host creation, and Production Activation do
not exist.

## 28. Decision

The implemented-in-isolation minimal Source is a stateless adapter over the two existing
concrete in-memory repositories. It reads the exact Version before its exact
parent, treats the Version read as L under the audited single-instance
mutation topology, returns a deeply detached `uwp.configuration` v1
observation, and maps runtime failures into the existing applicable Loader
error set without synthesizing inconsistent-observation errors. It enables
future construction without authorizing management routing, persistence, or
Production Activation.

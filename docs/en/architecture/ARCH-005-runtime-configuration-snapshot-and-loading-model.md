# ARCH-005: Runtime Configuration Snapshot and Loading Model

[Russian version](../../ru/architecture/ARCH-005-runtime-configuration-snapshot-and-loading-model.md)

**Status:** Active

**Scope:** Runtime Configuration Snapshot and the loading boundary between Control Plane and Runtime

**Stability:** Approved architectural model

## 1. Purpose

ARCH-005 defines what starts a Runtime: one immutable Runtime Configuration Snapshot constructed from the exact Published ConfigurationVersion pinned by one Launch Attempt.

The document establishes the boundaries between Configuration sources, Configuration Loader, Builder, Snapshot, Runtime Bootstrap, Runtime Host, and Runtime Services. It defines architecture and ownership, not a Loader implementation.

This model complements [ADR-0002](../adr/0002-configuration-dsl.md), [ADR-0003](../adr/0003-runtime-architecture.md), [ARCH-002](ARCH-002-runtime-foundation-freeze.md), and [ARCH-004](ARCH-004-runtime-deployment-and-identity-model.md). It does not change the Runtime Foundation or the operational identity model.

## 2. Context and Scope

The Control Plane owns declarative Configuration and its version lifecycle. Runtime owns execution from a stable input. A direct dependency from Runtime Services on Configuration repositories, HTTP models, YAML representations, or publication state would merge these responsibilities and make running behavior dependent on mutable management state.

The required boundary is:

```text
Configuration source
    -> Published ConfigurationVersion
    -> Configuration Loader
    -> Builder
    -> immutable Runtime Configuration Snapshot
    -> Runtime Bootstrap
    -> Runtime Host
    -> Runtime Services
```

This document defines that boundary, Snapshot content and provenance, construction and ownership, validation and normalization responsibilities, schema compatibility, concurrency, and failure semantics.

It does not define:

- Go APIs or interfaces;
- repository, PostgreSQL, or HTTP schemas;
- caching, polling, or retry policies;
- Runtime Instance persistence;
- Runtime Launcher implementation;
- reload, replacement, or reconciliation;
- secret storage backends;
- configuration schema migration or version negotiation.

## 3. Runtime Configuration Snapshot

Runtime Configuration Snapshot is the immutable, detached, execution-ready input for one Runtime Host.

Snapshot is not:

- Configuration or ConfigurationVersion;
- a Control Plane repository entity;
- an HTTP DTO;
- a YAML document model;
- a persistence record;
- a Runtime service or mutable Runtime state.

Snapshot contains the effective configuration required by the supported Runtime component graph and the provenance required to identify its declarative and operational origin. It contains neither management history nor mutable operational state.

One Snapshot belongs to one Launch Attempt. A Runtime starts from that Snapshot and never consults the source ConfigurationVersion again during that Host lifetime.

## 4. Configuration-to-Runtime Boundary

Control Plane responsibility ends after it has made an exact Published ConfigurationVersion available through the Configuration Loader boundary. Runtime responsibility begins with construction of a detached Snapshot from that selected source.

The boundary permits immutable data transfer. It does not transfer:

- repository access;
- publication authority;
- Configuration lifecycle mutation;
- ConfigurationVersion history;
- Control Plane service ownership;
- source-specific parsing or storage behavior.

Runtime must not infer Configuration by querying management services. Control Plane must not mutate Snapshot or Runtime-owned state after handoff.

## 5. Configuration Source Model

A Configuration source is an adapter that can supply the declarative model required by the Loader boundary. PostgreSQL, YAML, tests, in-memory repositories, imports, and future adapters are possible sources.

Source type is an implementation and deployment concern. It must not create a second Configuration language or change Runtime semantics. Every source must represent the same ConfigurationVersion domain model and preserve its identity, lifecycle state, schema identity, and configuration values.

The repository currently implements in-memory Control Plane storage and direct values in tests. PostgreSQL and YAML are architectural representations, not production loading capabilities declared by this document.

## 6. Published ConfigurationVersion

Only an exact Published ConfigurationVersion may be selected as the declarative source of a new Snapshot.

Draft, Validated, and Archived versions cannot be selected for a new Launch Attempt. The Loader must establish the Published state and exact identity as one consistent source observation. A later publication or archival does not change the already selected version, its Snapshot, or the active Launch Attempt.

ConfigurationVersion remains owned by the Control Plane. Runtime borrows its immutable content only long enough to construct a detached Snapshot. Runtime never changes its state or writes derived values back to it.

## 7. Configuration Loader Boundary

Configuration Loader is the architectural boundary that obtains the exact Published ConfigurationVersion selected for one Launch Attempt and supplies it for Snapshot construction.

The Loader is responsible for:

- preserving the requested Workspace, Configuration, ConfigurationVersion, Runtime Instance, and Launch Attempt identity relationship;
- returning only a consistently observed Published source;
- preserving Configuration schema identity;
- reporting load, identity, lifecycle-state, and source-integrity failures before Runtime construction;
- preventing Runtime components from receiving repository or management API access.

The Loader does not:

- choose desired state or make launch decisions;
- select an arbitrary newer version after the Launch Attempt has pinned one;
- build Runtime components;
- normalize Runtime configuration;
- resolve Secret values;
- retain ownership of Snapshot;
- define caching, polling, persistence, or retry policy.

ARCH-005 defines this responsibility without defining a Loader interface, methods, or concrete adapters.

## 8. Builder Boundary

Builder transforms one loaded Published ConfigurationVersion plus the already established operational provenance into one validated, normalized, detached Snapshot.

Builder is responsible for:

- defensive validation of the input required for Snapshot construction;
- deterministic normalization according to the Configuration domain semantics;
- complete conversion of supported behavior-affecting configuration;
- deep detachment from all caller-owned mutable memory;
- construction of the required provenance;
- rejection of unsupported or malformed input without publishing a partial Snapshot.

Builder is not a Loader. It does not know repositories, HTTP, YAML, PostgreSQL, publication history, or management commands. It does not select Configuration or decide whether a Runtime should launch. It does not retain the source or the constructed Snapshot after returning it.

## 9. Snapshot Construction Ownership

Snapshot construction is owned by launch preparation at the Runtime boundary. Runtime Lifecycle Owner first creates and pins the Launch Attempt and exact Published ConfigurationVersion. The Loader obtains that source, and Builder constructs the Snapshot for that attempt.

The construction flow is:

```text
Runtime Lifecycle Owner
    -> pins Launch Attempt and Published ConfigurationVersion
    -> invokes Configuration Loader
    -> supplies loaded source and operational provenance to Builder
    -> receives complete immutable Snapshot
    -> passes Snapshot through Runtime Launcher to Bootstrap
```

No Snapshot is architecturally visible until construction succeeds completely. Builder creates the value, but it does not own the launch decision or the resulting lifetime.

## 10. Snapshot Content Model

Snapshot contains exactly two architectural categories of information:

1. effective, behavior-affecting Runtime configuration represented by the selected Configuration schema;
2. stable provenance that identifies the declarative source and the execution identity for which the Snapshot was constructed.

Snapshot must contain every supported value required to compose and run the Runtime without returning to Control Plane. It may preserve configured capabilities whose execution belongs to an approved later stage, provided existing startup capability rules explicitly permit them to remain inactive.

Snapshot must not contain:

- Secret values;
- repositories, services, loaders, or resolver implementations;
- mutable Runtime state;
- desired or actual lifecycle state;
- ConfigurationVersion history;
- source adapter type;
- load timestamps, cache metadata, or transport metadata;
- deployment overrides not represented by an approved architecture.

Concrete field layout is an implementation concern and is not defined here.

## 11. Snapshot Provenance

Every Snapshot must carry enough immutable provenance to establish all of the following relationships:

- the Workspace that owns the Configuration;
- the Configuration identity;
- the exact ConfigurationVersion identity and version number;
- the Configuration schema identity and version represented by the Snapshot;
- the Runtime Instance for which launch was requested;
- the Launch Attempt for which the Snapshot was constructed.

Provenance is identity, not telemetry. Source adapter type, load time, construction duration, process identity, Host pointer, PID, and socket address are not Snapshot provenance.

The concrete representation and field names require focused implementation design. Their semantic identity and consistency are mandatory.

## 12. Schema Compatibility Model

Snapshot represents one concrete Configuration schema. Its provenance identifies that schema so the Runtime boundary can prove that the selected configuration is interpretable by the current Builder and Runtime component graph.

An unsupported, unknown, or incompatible schema must fail launch preparation before Runtime Host acquires resources or becomes Ready. Runtime must not guess missing semantics, silently downgrade input, or reinterpret the source through hidden defaults.

This document does not define schema negotiation, migration, compatibility ranges, or adapter versioning. Those mechanisms require a focused Design Proposal.

## 13. Validation Layering

Validation is layered without transferring ownership between components:

1. Configuration source adapters validate representation integrity and preserve the domain model.
2. Control Plane validates Configuration domain and publication rules.
3. Loader validates requested identity, Published eligibility, consistent source observation, and source integrity.
4. Builder defensively validates completeness, supported schema, cross-field Snapshot invariants, and deterministic conversion.
5. Runtime Bootstrap validates startup-critical capabilities before acquiring externally visible resources and before readiness.

Validation at an earlier layer does not permit a later trust boundary to omit defensive checks. Layers must not contradict the Configuration domain's acceptance semantics. A failure stops launch preparation or startup at the layer that owns the relevant responsibility.

## 14. Normalization Boundary

Configuration domain rules define semantic equality and canonical values. Source adapters must preserve those semantics and may perform only representation-level decoding.

Builder is the authoritative Runtime-boundary normalization step. It produces one canonical Snapshot without mutating the loaded ConfigurationVersion. Equivalent valid inputs from different sources must produce semantically equivalent Snapshot values.

Loader does not normalize Runtime configuration. Bootstrap and Runtime Services do not repeat Configuration normalization or establish source-specific defaults. Defensive validation in multiple layers is permitted, but its acceptance semantics must remain equivalent.

## 15. Snapshot Immutability

Snapshot becomes immutable at successful construction. From that point:

- no owner may mutate it in place;
- no caller-owned mutable memory may alias its logical content;
- readers receive immutable views or detached copies;
- Runtime Services may read concurrently without synchronization for Snapshot mutation;
- derived operational state must be stored outside Snapshot;
- publication of a new ConfigurationVersion cannot change it.

Immutability is an ownership guarantee, not merely a convention that callers should avoid mutation.

## 16. Snapshot Ownership and Lifetime

Ownership transfers in one direction:

| Stage | Ownership and permitted action |
|---|---|
| ConfigurationVersion | Control Plane owns the immutable published source |
| Loader | Temporarily owns the load operation and detached loaded material; never owns Snapshot lifetime |
| Builder | Owns construction only; retains neither source nor result |
| Launch preparation | Owns the complete Snapshot before Runtime construction |
| Runtime Bootstrap | Accepts the Snapshot for one Host construction and transfers an independent immutable value to Host |
| Runtime Host | Owns Snapshot for the entire Host lifetime |
| Runtime Services | Read only; they do not own, replace, or mutate Snapshot |

If Bootstrap fails before ownership transfer completes, launch preparation releases its values and no Host-visible Snapshot exists. After successful construction, Host retains Snapshot through startup, Running, shutdown, and terminal completion. Its value lifetime ends when the terminal Host and all permitted readers become unreachable.

Snapshot has no explicit destruction protocol because it contains no Secret values or independently owned Runtime resources.

## 17. Secret Reference Boundary

ConfigurationVersion and Snapshot contain only Secret References. Loader and Builder preserve and validate references but never resolve or embed Secret values.

Secret values enter only through the explicitly composed Secret Resolver after Snapshot construction and remain owned by the consuming Runtime capability. They must not be written back to ConfigurationVersion or Snapshot, included in provenance, logged, persisted, or exposed through Snapshot inspection.

ARCH-005 does not change existing Authentication semantics. ADR-0003 describes startup resolution for required references, while some current provider contracts resolve references during request processing. The exact resolution time and value lifetime for each consuming capability remain a focused architectural follow-up; this document resolves neither by moving Secret values into Snapshot nor by silently changing provider behavior.

## 18. Launch Attempt Integration

Runtime Lifecycle Owner creates one Launch Attempt and pins one exact Published ConfigurationVersion before loading begins. Snapshot construction uses that fixed pair and records the Runtime Instance and Launch Attempt provenance.

One Launch Attempt may produce at most one successful Snapshot and at most one Runtime Host. A failed construction leaves no runnable Host and is recorded as a failed historical Launch Attempt according to ARCH-004.

A retry or replacement creates a new Launch Attempt and a new Snapshot. Snapshot reuse across Launch Attempts is forbidden even when both attempts select the same ConfigurationVersion, because their operational provenance differs.

## 19. Publication Consistency and Concurrency

The source selection linearization point is the consistent observation that the exact version pinned by the Launch Attempt is Published. Everything before that point is selection and validation; everything after it uses that immutable version identity.

The architecture requires:

- no Draft, Validated, or already Archived version may cross the selection boundary for a new attempt;
- publication of another version after selection does not redirect the attempt;
- concurrent publication cannot produce a Snapshot composed from multiple versions;
- concurrent starts for one Runtime Instance remain governed by ARCH-004's single active Launch Attempt rule;
- starts for different Runtime Instances may independently select the same Published ConfigurationVersion;
- no running Snapshot is replaced, refreshed, patched, or normalized again;
- readers observe only the complete constructed Snapshot, never partial construction.

Publication consistency concerns the selected declarative source. It does not introduce an in-place reload mechanism.

## 20. Failure Boundaries

Loading or construction failure occurs before Runtime Host ownership and before Runtime resource acquisition. Such failure:

- publishes no Snapshot to Runtime Services;
- starts no Listener and opens no admission;
- leaves no partially composed Runtime graph;
- does not mutate the source ConfigurationVersion;
- does not select a fallback or newer version implicitly;
- produces a truthful failed Launch Attempt through the ARCH-004 lifecycle owner.

Bootstrap failures after a Snapshot has been handed to Runtime remain governed by ARCH-002 startup transaction and rollback. ARCH-005 does not define retry, backoff, fallback, or recovery policy.

## 21. Source Equivalence

All conforming sources must produce semantically equivalent Snapshots for the same ConfigurationVersion, schema, Runtime Instance, and Launch Attempt identities.

Equivalence includes effective Runtime configuration, normalization, ordering, absence-versus-presence semantics, Secret References, and provenance. Source-specific formatting, storage metadata, retrieval path, and adapter type must not affect the result.

Tests and in-memory sources are subject to the same boundary. They must not gain privileged fields, relaxed validation, or alternate defaults unavailable to production sources.

## 22. Required Invariants

1. Runtime starts from exactly one complete immutable Snapshot.
2. Every Snapshot originates from exactly one Published ConfigurationVersion selected for a new Launch Attempt.
3. Draft, Validated, and Archived versions cannot become the source of a new Snapshot.
4. One Snapshot belongs to exactly one Runtime Instance and one Launch Attempt.
5. One Launch Attempt produces at most one successful Snapshot and one Runtime Host.
6. Snapshot identifies its Workspace, Configuration, exact ConfigurationVersion, Configuration schema, Runtime Instance, and Launch Attempt provenance.
7. Snapshot contains effective Runtime configuration and provenance only.
8. Snapshot contains Secret References and never Secret values.
9. Loader and Builder never expose Control Plane repositories or management APIs to Runtime.
10. Loader does not choose Configuration, make lifecycle decisions, normalize Runtime values, or build Runtime components.
11. Builder does not load, persist, select, publish, or mutate ConfigurationVersion.
12. Successful construction fully detaches Snapshot from caller-owned mutable memory.
13. Snapshot never changes after construction and is never reloaded in place.
14. Runtime Services read Snapshot and never own, replace, or mutate it.
15. Publication of a newer ConfigurationVersion does not affect an existing Snapshot or Host.
16. Equivalent valid source representations produce semantically equivalent Snapshots.
17. Unsupported schema or malformed input fails before Runtime resource acquisition and readiness.
18. No observer can access a partially constructed Snapshot.
19. Bootstrap and Host receive Snapshot without receiving Loader, repository, or publication authority.
20. Snapshot lifetime is bounded by its Host and permitted readers and owns no independently destructible resource.
21. Source adapter type, load timestamp, PID, Host pointer, and socket identity do not become Snapshot provenance.
22. Retry, replacement, and restart require a new Launch Attempt and a new Snapshot.

## 23. Follow-up Design Questions

Focused Design Proposals must define implementation contracts for:

- Configuration Loader commands, errors, and adapter integration;
- concrete Snapshot provenance representation;
- Configuration schema compatibility ranges, negotiation, and migration;
- Runtime Instance and Launch Attempt persistence;
- secret-resolution timing and value lifetime per consuming capability where current contracts differ from ADR-0003's general startup model;
- operational diagnostics and redaction for loading and construction failures;
- replacement, rollback, reconciliation, and retry policies.

These questions must not be answered through hidden Loader or Builder behavior.

## 24. Decision

UWP adopts Runtime Configuration Snapshot as the sole immutable Runtime input between Published ConfigurationVersion and Runtime Host.

Runtime Lifecycle Owner pins one exact Published ConfigurationVersion for one Launch Attempt. Configuration Loader obtains that exact published source without exposing persistence or management infrastructure. Builder defensively validates, normalizes, detaches, and constructs one Snapshot with effective Runtime configuration and mandatory declarative and operational provenance. Bootstrap transfers the complete Snapshot to one Host, which owns it for its lifetime; Runtime Services are read-only consumers.

All Configuration sources must be semantically equivalent at this boundary. Snapshot never contains Secret values, never changes in place, and is never redirected by later publication. Loading, construction, schema incompatibility, and validation failures prevent Runtime resource acquisition and produce no partial Runtime input.

# LLM Development Guide

## Purpose

This guide defines the engineering workflow for development assisted by language models. It is a living project standard for contributors, maintainers, the Project Owner, and engineers implementing approved Design Proposals.

The project uses a structured workflow instead of ad hoc code generation to provide:

- deterministic development;
- traceable architecture decisions;
- reproducible implementation;
- independent verification;
- long-term maintainability.

Assistance does not transfer engineering accountability. Every change remains subject to the same architecture, scope, testing, review, and ownership requirements as any other project contribution.

## Core Principles

### Architecture Before Code

The intended component model, responsibilities, ownership, lifecycle, invariants, and boundaries must be established before implementation begins.

### Design Proposal as the Source of Truth

An approved Design Proposal is the normative source for the behavior and architecture in its scope. Implementation must conform to it and must not silently reinterpret it.

### Compile Once

Configuration-dependent validation, normalization, ordering, and resolution should occur once at the appropriate construction boundary. Published runtime structures should be ready for direct execution without defensive recompilation on operational paths.

### Immutable by Default

Published configuration, compiled structures, identity snapshots, and message context values should be immutable. Mutable state requires explicit ownership and synchronization contracts.

### Small, Independently Reviewable Steps

Implementation must be decomposed into the smallest coherent steps that compile, preserve existing behavior, and can be reviewed independently.

### One Completed Idea per Commit

Each commit should contain one complete engineering idea, its tests, and any documentation required to describe that idea. Partial or unrelated changes must not be combined.

### Performance Requirements Belong to Architecture

Allocation limits, bounded work, compilation boundaries, and concurrency properties must be specified before they are relied upon by implementation.

### Independent Verification

Implementation and review are distinct responsibilities. Review must inspect the repository state and evidence directly rather than rely on the implementation report.

### Blockers Are Preferable to Incorrect Implementations

Implementation must stop when the approved architecture is ambiguous, contradictory, or insufficient. Missing decisions must be resolved explicitly before coding continues.

### Hot Paths Require Explicit Performance Review

Message processing, routing, admission, and other frequently executed paths require deliberate review of allocations, synchronization, data copying, boundedness, and hidden work.

## Roles

### Architecture

The architecture role:

- defines component responsibilities and boundaries;
- assigns ownership and lifecycle authority;
- specifies observable behavior and invariants;
- records material decisions in the appropriate architectural document;
- resolves ambiguity before implementation begins;
- approves or rejects proposed architectural changes.

### Implementation

The implementation role:

- follows the approved Design Proposal and implementation scope;
- inspects existing code and tests before making changes;
- introduces no hidden architectural decisions;
- preserves compatibility unless a change is explicitly approved;
- implements deterministic proof tests;
- runs all required verification;
- reports the resulting repository state accurately.

### Review

The review role:

- independently compares code and tests with the approved architecture;
- challenges ownership, lifecycle, concurrency, immutability, and boundary claims;
- verifies that tests prove the required invariants;
- classifies findings by impact;
- distinguishes architectural defects from implementation defects;
- withholds approval when evidence is incomplete.

### Project Owner

The Project Owner:

- controls project scope and roadmap priority;
- approves architecture and material changes to it;
- resolves product-level trade-offs;
- decides whether findings block progress;
- authorizes commits and integration;
- ensures that project records reflect completed work.

No role may assume the authority assigned to another role merely to keep implementation moving.

## Standard Workflow

```text
MASTER PLAN
    |
    v
Design Proposal
    |
    v
Architecture Review
    |
    v
Implementation Prompt
    |
    v
Implementation
    |
    v
Independent Review
    |
    v
Architecture Fix, if required
    |
    v
Commit
    |
    v
Post-Implementation Architecture Review
    |
    v
CHANGELOG
```

### Master Plan

The Master Plan identifies the engineering sequence and milestone intent. It selects the problem area but does not replace a Design Proposal or prescribe future APIs.

### Design Proposal

The Design Proposal defines the scoped architecture, contracts, invariants, failure model, compatibility requirements, and excluded work. It must contain enough information for implementation without inventing new decisions.

### Architecture Review

Independent architecture review attempts to disprove the proposal. Blocking ambiguity, ownership gaps, invalid lifecycle transitions, and unproven concurrency semantics must be resolved before implementation.

### Implementation Prompt

The implementation prompt converts the approved design into one bounded engineering step. It states the permitted files and behavior, exclusions, required tests, verification commands, and expected report.

### Implementation

Implementation changes only the approved scope. It includes proof tests and preserves the repository in a compiling, reviewable state.

### Independent Review

Independent review verifies actual code and tests against the Design Proposal and implementation prompt. It does not treat a successful build or implementation report as proof of architectural correctness.

### Architecture Fix

If review exposes an architectural defect, implementation pauses. The model is analyzed first, the required architectural change is approved, and the correction is delivered separately from unrelated work. Implementation-only defects do not require an architecture change.

### Commit

A commit is created only after the scoped implementation and required verification are complete and review findings have been resolved or explicitly accepted. The commit records one completed idea.

### Post-Implementation Architecture Review

The completed implementation may be reviewed against the broader component model to confirm that local correctness did not introduce boundary drift or invalidate downstream assumptions.

### Changelog

The changelog is updated only when the repository contains the completed capability. It records the implemented state, not intended or partially implemented behavior.

## Design Proposal Rules

- Implementation must not change architecture.
- The proposal must distinguish normative requirements from examples and future work.
- Any ambiguity that affects ownership, lifecycle, behavior, concurrency, failure handling, or public contracts stops implementation.
- Architecture changes require explicit review and approval in the appropriate document.
- An implementation convenience is not sufficient justification for changing an approved invariant.
- Deferred behavior must remain absent; placeholder APIs must not imply unsupported capability.
- Code is evidence of implementation state, not a substitute for the approved architectural model.

## Implementation Rules

- Work must remain within one explicitly bounded scope.
- Each component and change must have a single stated responsibility.
- Existing behavior must remain compatible unless the prompt authorizes a change.
- Public APIs must not be added for speculative future use.
- Dependencies must follow approved package and ownership boundaries.
- Operational paths must not acquire hidden compilation, normalization, I/O, concurrency, or allocation costs.
- Tests must prove the specified success, failure, lifecycle, concurrency, and compatibility invariants.
- Synchronization tests must be deterministic and must not use sleep-based ordering.
- All started goroutines must be deterministically released and joined by tests.
- Required formatting, tests, static analysis, and diff checks must pass before completion is reported.
- The final report must describe actual changes and verification results, including checks that could not be performed.

## Review Rules

Reviewers must verify the following where applicable:

- dependency direction and package boundaries;
- immutable ownership and absence of aliasing;
- resource, connection, context, and lifecycle ownership;
- legal state transitions and terminal semantics;
- concurrency linearization points and exactly-once guarantees;
- bounded execution and algorithmic cost;
- hot-path allocations, copying, locking, and hidden work;
- error identity and observable behavior;
- backward compatibility and default behavior;
- consistency between architecture, implementation, tests, and current-state documentation;
- compliance with the explicit scope and exclusions.

Review must cite concrete documents, files, and tests. Successful formatting, compilation, or tests do not by themselves prove ownership or architectural correctness. Findings must identify the violated invariant and its impact.

## Prompt Standard

Every implementation prompt should contain the following sections.

### Objective

Defines the single outcome of the task and prevents unrelated goals from entering the change.

### Scope

Identifies permitted behavior, packages, files, and architectural responsibilities. It establishes the authority available to implementation.

### Out of Scope

Lists adjacent capabilities that must remain absent. This prevents speculative APIs, hidden integration, and premature subsystem work.

### Architecture Constraints

Restates the applicable source documents, ownership rules, lifecycle invariants, compatibility requirements, dependency boundaries, and performance constraints.

### Tests

Specifies the properties that require proof, including failure and concurrency cases. Tests should demonstrate invariants rather than only exercise lines of code.

### Verification

Lists the exact formatting, test, static-analysis, race, allocation, documentation, and diff checks required for the task. Unavailable checks must be reported precisely and must not be represented as completed.

### Final Report

Defines the evidence required for handoff: changed files, resulting contracts, tests, verification results, limitations, risks, and a suggested commit message. It must describe the resulting state rather than the development narrative.

## Lessons Learned

Lessons are objective engineering knowledge extracted from completed Design Proposals and their implementation. They complement the governing architecture but do not alter it.

DP-005 provides the following examples:

- Independent review is necessary even when compilation and tests pass; it verifies whether implementation preserves the intended model.
- Immutable compiled runtime structures separate validation and preparation from message-time execution.
- Configuration normalization, Handler resolution, and route ordering belong at the construction boundary rather than the routing hot path.
- Priority ordering must be explicit in both the compiled representation and its proof tests; declaration order must not become an accidental runtime rule.
- Hot-path behavior benefits from focused allocation and concurrency proofs in addition to functional tests.
- Architecture corrections should be isolated from feature implementation so that their intent, evidence, and rollback boundary remain clear.

New lessons should record reusable engineering outcomes. They must not replace Design Proposals, decision records, implementation reports, or project history.

## Future Evolution

This guide evolves with the project. Future Design Proposals, implementation reviews, and demonstrated engineering constraints may extend the workflow. Changes to this guide must remain consistent with the project's architecture, decision process, and repository verification standards.

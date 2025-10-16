# Roadmap for GQL (ISO/IEC 39075:2024) Support in Cayley

## Context and Goals

Cayley currently offers Gizmo, a GraphQL-inspired language, and MQL query languages through the shared query session interface. Supporting the official ISO/IEC 39075:2024 Graph Query Language (GQL) will broaden interoperability with industry standards while keeping the modular design of Cayley intact.【F:README.md†L18-L27】【F:query/session.go†L14-L114】

This roadmap outlines the recommended phases to deliver a compliant and production-ready GQL implementation that integrates cleanly with existing query infrastructure and storage back ends.【F:graph/quadstore.go†L15-L112】

## Phase 0: Standards Alignment and Discovery

1. **Requirements Deep Dive**
   - Produce a specification summary that maps GQL constructs to Cayley capabilities and identifies gaps in existing components.
2. **Success Metrics and Acceptance Tests**
   - Define compliance tests (e.g., open-source GQL suites, bespoke validation sets) and performance benchmarks that must be met before GA.

## Phase 1: Architecture Preparation

1. **Language Registration and Session Lifecycle**
   - Extend the `query.Language` registry with a `gql` entry and design a `Session` implementation that reuses existing options (collations, limits) while honoring GQL result shapes.【F:query/session.go†L51-L114】
   - Ensure HTTP and REPL entry points can negotiate the new language without breaking existing clients.
2. **Quad Store Capability Audit**
   - Review `graph.QuadStore` interfaces and back ends for compatibility with GQL requirements such as labeled paths, temporal data, and schema constraints.【F:graph/quadstore.go†L15-L112】
   - Create a backlog of storage-layer enhancements (e.g., new indexes, statistics) needed to satisfy anticipated query plans.
3. **Schema and Type System Foundations**
   - Evaluate the `schema` and `inference` packages for reuse in modeling GQL types, constraints, and validation flows.
   - Prototype a metadata repository that tracks catalogs, graphs, roles, and authorization semantics mandated by GQL.

## Phase 2: Parser, AST, and Validation

1. **Grammar Implementation**
   - Implement or integrate a parser for ISO/IEC 39075:2024 (ANTLR, goyacc, or hand-written) that outputs an abstract syntax tree (AST) aligned with Cayley execution needs.
   - Establish incremental parsing facilities to support multi-statement scripts and interactive sessions.
2. **Semantic Analysis Layer**
   - Build a semantic validator that resolves graph objects, views, and schema references against Cayley metadata before execution.
   - Enforce access control, typing, and conformance rules at this stage to keep the executor focused on query planning.
3. **Error Reporting and Tooling**
   - Design diagnostics that align with GQL terminology and integrate with existing HTTP error handlers and REPL messaging.【F:query/session.go†L80-L106】
   - Provide developer tooling (linting, formatting) to ease adoption and testing.

## Phase 3: Query Planning and Execution

1. **Logical Plan Construction**
   - Translate validated ASTs into logical plans that express pattern matching, graph transformations, and result shaping in Cayley’s iterator language.
   - Reuse and extend existing iterator and path packages to cover GQL features such as composable graph patterns, path expressions, and tabular projections.
2. **Optimizer Extensions**
   - Implement cost estimation strategies that leverage `QuadStore` statistics for cardinality and selectivity estimates.【F:graph/quadstore.go†L63-L101】
   - Add rewrite rules and heuristics tailored to GQL constructs (e.g., shortest paths, graph views, recursive patterns).
3. **Execution Engine Enhancements**
   - Ensure the runtime can deliver GQL result formats (tables, graph views, JSON) mapped to Cayley’s collation modes.【F:query/session.go†L51-L79】
   - Optimize iterators for streaming large result sets and pipelining across distributed back ends where applicable.

## Phase 4: Ecosystem Integration

1. **API and Client Updates**
   - Expose GQL through existing HTTP endpoints, gRPC (if applicable), and client SDKs alongside Gizmo/MQL.
2. **Tooling and UI**
   - Update the built-in query editor and visualization tools in `ui/` to support GQL syntax highlighting, auto-completion, and explain plans.
   - Add REPL commands for managing catalogs, roles, and other GQL administrative tasks.
3. **Operational Readiness**
   - Extend configuration schemas, deployment manifests, and monitoring dashboards to expose GQL-specific knobs and metrics.

## Phase 5: Verification and Release

1. **Compliance and Interoperability Testing**
   - Execute comprehensive compliance suites, regression tests across all back ends, and interoperability checks with external GQL tools.
   - Benchmark performance and compare to baseline Gizmo/MQL queries to validate parity or improvements.
2. **GA Criteria and Support Plan**
   - Define clear release gates (test coverage, SLAs, bug burndown) and prepare an ongoing maintenance cadence for the GQL subsystem.
   - Establish triage guidelines for GQL-related issues and feedback loops for future enhancements.

## Suggested Timeline and Dependencies

- **Quarter 1**: Complete Phase 0 deliverables and architectural audits.
- **Quarter 2**: Implement parser, AST, and semantic validation; land foundational session and storage updates.
- **Quarter 3**: Deliver execution engine integration, optimizer enhancements, and begin end-to-end testing.
- **Quarter 4**: Focus on ecosystem integration, compliance validation, and general availability launch.

Dependencies across phases should be tracked in the project board with explicit cross-links to Cayley packages (`query`, `graph`, `schema`, `ui`) to maintain visibility and enable incremental delivery.

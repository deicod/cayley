# Cayley GQL Support Specification

## Purpose
Provide an actionable specification for adding ISO/IEC 39075:2024 Graph Query Language (GQL) support to Cayley while preserving
compatibility with existing query languages and storage back ends.

## Objectives
- Deliver a standards-aligned GQL interface that plugs into Cayley query sessions and HTTP APIs.
- Reuse and extend existing iterator, schema, and storage capabilities where practical.
- Provide clear migration paths and tooling for users of Gizmo and MQL.
- Establish compliance, performance, and operational criteria for a production-ready launch.

## Non-Goals
- Replacing or deprecating Gizmo or MQL language support.
- Implementing vendor-specific GQL extensions outside the ISO/IEC 39075:2024 scope.
- Building new storage engines; focus is on enhancing current QuadStore implementations.

## Stakeholders
- Core Cayley maintainers responsible for query execution, storage adapters, and API surfaces.
- Developer tooling and UI contributors supporting editors, REPLs, and documentation.
- Early adopters providing feedback on standards coverage and interoperability.

## Requirements Summary
| Area | Requirement |
| --- | --- |
| Language Registration | Register `gql` in `query.Language` with session lifecycle parity with Gizmo/MQL. |
| Parsing & Validation | Implement grammar, AST, and semantic analysis that enforce schema, typing, and security constraints. |
| Query Planning | Map validated ASTs to iterator-based logical plans with optimizer support for GQL constructs. |
| Execution | Produce result sets in GQL-compliant formats (tables, graph views, JSON) with streaming optimizations. |
| Storage | Audit existing QuadStores for feature coverage; design enhancements for statistics, indexes, and metadata. |
| Tooling & UX | Update HTTP endpoints, REPL, UI components, and developer tooling to recognize GQL syntax and workflows. |
| Compliance | Define and execute conformance, performance, and interoperability test suites before GA. |
| Operations | Extend configuration, deployment, and monitoring assets with GQL-specific controls and metrics. |

## High-Level Architecture
1. **Session Integration**
   - Extend `query/session.go` to instantiate a `GQLSession` implementing the `Session` interface with language negotiation through HTTP and REPL entry points.
   - Share common session options (collation, limit handling) while mapping GQL result projections to Cayley iterators.
2. **Parsing and Semantic Layer**
   - Introduce a `gql/parser` package producing an AST tailored for execution planning.
   - Add a `gql/semantic` component that resolves catalog metadata, enforces access control, and performs type validation using existing `schema` and `inference` packages.
3. **Planning and Execution**
   - Create a `gql/planner` responsible for transforming semantic graphs into iterator trees, reusing `path` and `iterator` packages.
   - Extend optimizer infrastructure with GQL-aware cost estimators and rewrite rules.
4. **Metadata and Catalog Management**
   - Provide a metadata service for catalogs, graphs, roles, and grants, integrating with storage-specific capabilities and configuration.
5. **Client & Tooling Surface**
   - Update HTTP handlers, REPL commands, and the `ui/` query editor for language selection, syntax highlighting, and explain plans.

## Deliverables
- `gql` package hierarchy (`parser`, `semantic`, `planner`, `executor`) wired into `query` registry.
- Storage capability matrix documenting required enhancements per backend.
- Test harness covering parser conformance, semantic validation, planner correctness, and execution integration.
- Documentation updates (usage guides, API references, migration notes).
- Operational artifacts (configuration schema updates, monitoring metrics, deployment guides).

## Risks and Mitigations
- **Standards Complexity**: Mitigate by maintaining a living specification summary and aligning with community compliance suites.
- **Performance Regression**: Introduce benchmarking pipelines comparing GQL to existing Gizmo/MQL workloads; optimize iterators accordingly.
- **Storage Gaps**: Prioritize audits and prototype indexes/statistics early; provide fallbacks where full feature support is not feasible.
- **Adoption Friction**: Deliver tooling, tutorials, and migration guides in parallel with core implementation.

## Acceptance Criteria
- `gql` language selectable via HTTP API, REPL, and client SDKs with parity in session options.
- End-to-end queries pass conformance suites covering pattern matching, path expressions, tabular projections, and administrative commands.
- QuadStore back ends document and expose feature support, with tests demonstrating compliant behavior or explicit limitations.
- Query editor and documentation updated to include GQL usage examples and troubleshooting guidance.
- Monitoring dashboards expose GQL-specific metrics (query counts, latencies, error classes).


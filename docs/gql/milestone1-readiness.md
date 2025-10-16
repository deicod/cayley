# GQL Milestone 1 — Standards Alignment and Architecture Readiness

Milestone 1 prepares Cayley for a first-class GQL implementation by aligning the graph
engine with ISO/IEC 39075:2024 expectations and scoping the required infrastructure. The
tasks below summarize the current state, gaps, and planned follow-ups.

## 1. Standards Coverage Summary

The table maps notable GQL constructs to existing Cayley features and highlights required
enhancements. References to Cayley packages are provided for cross-checking.

| GQL construct | Cayley support | Gaps / follow-up work |
| ------------- | -------------- | --------------------- |
| Catalogs & Graph spaces | Multiple named graphs are supported via `QuadStore` contexts and configuration profiles. | Need formal catalog metadata describing graph spaces, owners, and lifecycle, exposed through admin APIs. |
| Schema types (node, edge, tuple) | `schema` package supports JSON-based schema definitions with type validation. | Extend schema to express GQL tuple and multigraph constraints; align terminology and error reporting. |
| Pattern matching (`MATCH`, `WHERE`) | `path` and `iterator` packages provide pattern execution with filters. | Parser/semantic layers must translate GQL syntax; add support for named path values and variable bindings. |
| Path patterns (quantifiers, concatenation) | Gizmo path API already supports repetition, alternation, and saved paths. | Introduce parser lowering to iterator plans; ensure all quantifier semantics (`*`, `+`, `{m,n}`) are covered. |
| Tabular projections (`SELECT`, `GROUP`) | `query/mql` and `query/graphql` demonstrate projection and grouping. | Need GQL-specific projection planner, grouping semantics, and aggregation catalogue. |
| Data manipulation (`INSERT`, `DELETE`, `UPDATE`) | `graph/transaction` and writer packages handle quad mutations. | Define GQL transaction boundaries, IF EXISTS clauses, and optimistic locking semantics. |
| Constraints & integrity (`ASSERT`, `CONSTRAINT`) | `schema` and `inference` enforce basic validation. | Build constraint registry and enforcement pipeline triggered from GQL DDL. |
| Temporal & versioned queries | Backends expose quads with timestamp metadata (`log` and `nosql` backends). | Standardize temporal predicates and ensure planner can exploit timestamp indexes. |
| Security & roles | RBAC hooks exist in `internal/http` middleware, with coarse query authorization. | Integrate role metadata with catalog entries; enforce per-graph privileges in the semantic analyzer. |

## 2. Compliance and Benchmarking Plan

* **Conformance suites**
  * Author reusable fixtures describing GQL scripts with expected iterator plans and results.
  * Mirror ISO/IEC 39075 annex examples and the LDBC GQL benchmark once released.
  * Integrate suites into `go test ./...` by adding `gql/testdata` with YAML-based expectations.
* **Performance benchmarks**
  * Extend existing `clog` benchmarks to capture iterator composition costs for MATCH-heavy workloads.
  * Track latency and allocation deltas when running equivalent Gizmo, MQL, and GQL workloads on
    Memstore, Bolt, and PostgreSQL backends.
  * Automate benchmark runs in CI nightly jobs and publish regressions to the observability dashboards.

## 3. QuadStore Backend Audit

| Backend | Capabilities | Follow-up actions |
| ------- | ------------ | ----------------- |
| `graph/memstore` | Full in-memory feature set, rapid iteration. Supports path iterators and transactions. | Add optional statistics collector for planner costing; evaluate memory pressure under large MATCH queries. |
| `graph/kv` (BoltDB, Badger) | Persistent key-value stores with ordered iterators and history logs. | Confirm labeled path persistence; extend encoding to store temporal metadata and path names. |
| `graph/sql` (PostgreSQL, MySQL) | Relational backing with existing SQL schema. | Verify support for labeled paths and schema-qualified graphs; add materialized view hooks for projections. |
| `graph/http` | Remote execution against Cayley instances. | Specify protocol extensions for GQL result shapes and error payloads. |
| `graph/nosql` (MongoDB) | Document store integration, eventually consistent. | Assess feasibility of constraint enforcement and path statistics; potentially limit GQL feature set here. |

## 4. Metadata Catalog Requirements

To deliver GQL catalog semantics we need a central metadata registry describing graphs,
roles, schemas, and access policies. The following responsibilities are identified:

1. **Graph descriptors** — unique IDs, storage backend bindings, lifecycle state, retention
   policy, and optional temporal windows. Extend `configurations` to emit this metadata.
2. **Role & principal mappings** — map users/groups to catalog roles, providing CRUD APIs
   to administrative tooling. Reuse `internal/http/admin` authentication hooks.
3. **Schema linkage** — associate `schema` and `inference` definitions with catalog entries so
   semantic validation can resolve type definitions, key constraints, and inference rules.
4. **Authorization checks** — semantic analyzer (`gql/semantic`) must enforce per-role
   permissions before planning, mirroring existing authorization middleware.
5. **Introspection surfaces** — expose catalog state via HTTP/GraphQL endpoints and CLI, to
   support GQL `SHOW` statements and administrative audits.

These requirements inform the design of `gql/semantic` and planner components in later milestones.

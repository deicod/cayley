# GQL Milestone 2 — Parser, AST, and Semantic Validation

Milestone 2 establishes the front-end pipeline for Cayley's ISO/IEC 39075 GQL
implementation. The parser, diagnostics, and semantic checker provide the inputs
required for later planning milestones.

## Parser and AST Support

* `query/gql/parser` tokenizes multi-statement scripts, tracks source positions,
  and builds strongly-typed statement nodes (e.g. `UseGraphStatement`,
  `MatchStatement`, `CommandStatement`).【F:query/gql/parser/parser.go†L12-L332】
* The parser preserves quoted strings, nested projection expressions, and
  provides detailed diagnostics for unterminated strings, empty statements, and
  missing clauses.【F:query/gql/parser/parser.go†L64-L193】【F:query/gql/parser/parser.go†L236-L332】

## Semantic Validation

* `semantic.Validator` resolves active graphs, default schemas, and role-based
  permissions against a catalog abstraction to ensure scripts respect catalog
  metadata.【F:query/gql/semantic/semantic.go†L33-L148】
* Authorization and projection checks surface structured diagnostics for missing
  graph selections, unauthorized writes, and undefined variables in RETURN
  clauses.【F:query/gql/semantic/semantic.go†L88-L171】
* The in-memory catalog enables milestone prototypes without external
  dependencies while mirroring the APIs required for persistent metadata
  services.【F:query/gql/semantic/catalog.go†L7-L78】

## Diagnostics and Session Integration

* `diagnostic.Error` records severity, codes, and positional data, enabling HTTP
  and REPL surfaces to display consistent GQL terminology.【F:query/gql/diagnostic/diagnostic.go†L12-L86】
* `query/gql/session` wires the parser and validator into the query language
  registry, enforces Cayley's collation contract, and exposes structured HTTP
  error responses until execution support lands.【F:query/gql/session.go†L19-L126】

## Testing and Tooling

* Unit tests cover parser splitting, projection parsing, semantic authorization,
  and diagnostic propagation to guard against regressions.【F:query/gql/parser/parser_test.go†L1-L44】【F:query/gql/session_test.go†L1-L55】【F:query/gql/semantic/semantic_test.go†L1-L200】

These components complete the front-end readiness required by Milestone 2 and
unlock the planner and execution work tracked in Milestone 3.

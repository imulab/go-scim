# go-scim v3 (next)

## What This Is

A code-generation toolkit for building SCIM v2 protocol servers in Go. Users describe their SCIM resources (built-in User/Group/EnterpriseUser plus arbitrary custom types) via a definition that is parsed into a model, and the generator emits a working `http.Handler` plus SQL persistence and migration DDL. The output server is a plain Go module the user owns, builds, and runs.

Successor to the legacy `go-scim` (under `.legacy/`), which used a generic in-memory tree-like data model. This version replaces runtime genericity with build-time code generation.

## Core Value

A developer can describe their SCIM resource set once and get a spec-compliant SCIM v2 server with SQL persistence, mountable on any `net/http`-compatible mux, with no runtime schema-interpretation overhead.

## Requirements

### Validated

(None yet — ship to validate. Legacy capabilities are reference only; nothing inherited automatically.)

### Active

- [ ] Definition input that describes SCIM resources, attributes, and constraints (format TBD — fluent Go builder is the working hypothesis)
- [ ] Parser that ingests the definition into an internal model the generator consumes
- [ ] Code generator that emits a `http.Handler` implementing SCIM v2 endpoints (CRUD, /Schemas, /ResourceTypes, /ServiceProviderConfig, PATCH, bulk, filter, sort, ETag)
- [ ] Pluggable SQL persistence abstraction (driver interface)
- [ ] SQLite as the first reference SQL driver implementation
- [ ] Generator emits SQL migration DDL alongside the server code
- [ ] Generated server uses a typed `Config` struct (constructed by the user's `main`)
- [ ] CLI tool (`scimgen` or similar) that wraps the generator library and emits a runnable repo
- [ ] Library API exposing the same generation surface for in-process use
- [ ] Generated code includes unit tests, integration tests against SQLite, and a SCIM v2 compliance test suite
- [ ] Pluggable observability interfaces (logging/tracing/metrics) — no concrete impl shipped
- [ ] Support custom resource types (beyond User/Group/EnterpriseUser) declared via the definition
- [ ] Persistence layer designed tenant-id aware to enable later multi-tenancy (single-tenant only in v1)
- [ ] MIT license

### Out of Scope

- IdP non-compliance accommodations (e.g., Microsoft capital-letter attribute names) — strict spec only in v1; revisit later
- Authentication / authorization built into the generated server — assume an upstream gateway terminates auth
- Multi-tenancy at the request layer — single-tenant per generated server in v1
- Group-membership change eventing (the legacy `groupsync` over RabbitMQ) — implementation detail of the old project, not in the SCIM spec
- Backward compatibility with legacy `go-scim` consumers — clean break, different module path; legacy is frozen reference
- Concrete observability implementations (OTel, Prometheus, etc.) — interfaces only; user wires concrete deps
- Schema-evolution migrations between generator runs — generator emits fresh DDL each run; user owns diffing/migration in v1
- Explicit performance/scale targets — correctness over throughput in v1
- Concrete non-SQLite SQL drivers in v1 (Postgres/MySQL ship later, behind the same interface)
- Concrete HTTP-mux integrations beyond `net/http` — handler is plain `http.Handler`, user mounts on chi/gorilla/gin/etc.

## Context

**Predecessor (`.legacy/`):** Go SCIM v2 implementation built around a generic in-memory tree-like data model where every property node carried schema-attribute references, with a persistence adapter interface. MongoDB adapter (`.legacy/mongo/v2/`) worked because BSON is also tree-like; SQL adapters were difficult. Codebase mapped under `.planning/codebase/` for reference only — see `.planning/codebase/CONCERNS.md` for the three documented rewrite drivers.

**Three rewrite drivers (from user, confirmed by codebase analysis):**
1. **Feasibility:** generic tree model is hard to port to SQL. Code-gen generates per-resource SQL repos directly, sidestepping the tree-to-relational impedance.
2. **Efficiency:** maintaining a generic in-memory data model duplicates effort and still requires hacks for SCIM rules (e.g., "exactly one element of a multi-valued complex attribute can be `primary`"). Code-gen makes invariants explicit in the generated types.
3. **Portability:** mainstream IdPs have quirks (Microsoft capitalization, etc.); legacy hacks made the codebase brittle. v1 punts on quirks (out of scope) but the definition format is the natural future home for compatibility profiles.

**Working hypothesis on definition format:** a fluent Go builder that constructs a model object the parser ingests. This keeps definitions in-language, type-checked, and refactor-friendly. YAML/JSON is not ruled out and remains a Pending decision.

**Repo state:** legacy code preserved under `.legacy/` on the `next` branch. Top of working tree starts empty. Multi-module workspace planned (separate `go.mod` for generator vs. runtime support library).

## Constraints

- **Language:** Go — continuity with legacy and target ecosystem
- **Spec:** SCIM v2 (RFC 7643 schema, RFC 7644 protocol) — full coverage in scope for v1
- **HTTP:** generated handler is a plain `net/http` `http.Handler` — user mounts on any compatible mux; no opinionated router baked in
- **Persistence:** SQL only in v1; pluggable abstraction with SQLite shipped as reference; Postgres/MySQL deferred but must fit the same interface
- **Testing:** generated server must include unit, integration (against SQLite), and SCIM v2 compliance suites
- **Distribution:** both CLI tool and importable library; CLI is a thin wrapper over the library
- **License:** MIT, matching legacy
- **Tenancy:** single-tenant per generated server in v1; persistence layer must keep tenant-id awareness so multi-tenancy can be added later without rewriting repos

## Key Decisions

| Decision | Rationale | Outcome |
|----------|-----------|---------|
| Code generation over generic in-memory tree | Eliminates tree-to-SQL impedance; makes SCIM invariants explicit in generated Go types; addresses all three legacy pain points | — Pending |
| Pluggable SQL driver abstraction; SQLite first | Keeps the door open to Postgres/MySQL; SQLite is the easiest reference impl and useful for tests/local dev | — Pending |
| Plain `net/http` handler, no opinionated router | Maximum mux portability; user picks chi/gorilla/gin and mounts | — Pending |
| Both CLI and library distribution | Library covers in-process generation; CLI covers repo-emission workflows; CLI = thin wrapper | — Pending |
| Full SCIM v2 coverage in v1 (no subset MVP) | A "code-gen for SCIM" that doesn't cover the spec isn't credible; partial coverage hides architectural cliffs | — Pending |
| Generator emits migration DDL; user owns runtime correctness | Single source of truth for schema; user opts in or supplies their own migration tooling | — Pending |
| Spec-defined + arbitrary custom resource types | Custom types are a primary value-prop of code-gen; declaring them in the definition unifies treatment | — Pending |
| Observability via pluggable interfaces only | Avoids dragging OTel/Prometheus deps into generated servers; user wires concrete impls | — Pending |
| Single-tenant v1, design-extensible | Ships faster, but persistence layer carries tenant-id awareness so multi-tenancy is a later additive change | — Pending |
| Clean break from legacy `go-scim` | Different module path / major version; legacy frozen for reference; no compat shim | — Pending |
| Multi-module workspace | Separate `go.mod` for generator vs. runtime support library; matches legacy `pkg/v2`/`mongo/v2` split style | — Pending |
| Definition format: fluent Go builder (working hypothesis) | In-language, type-checked, refactor-friendly; YAML/JSON not ruled out | — Pending (revisit in arch phase) |
| IdP quirks out of v1 scope | Strict spec keeps the model honest; quirks become a definition-level feature later (compatibility profiles) | — Pending |
| Auth / authorization out of v1 scope | Assume upstream gateway; keeps the generator focused on the SCIM data plane | — Pending |
| Drop legacy `groupsync` | Not a SCIM spec requirement; was an implementation detail of the old project | — Pending |
| Config = typed `Config` struct, library-style; CLI-emitted main wires env/flags | Library users get type safety; CLI users get 12-factor ergonomics on top | — Pending |

---
*Last updated: 2026-05-07 after initialization*

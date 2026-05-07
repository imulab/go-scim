# Feature Research

**Domain:** SCIM v2 server code-generation toolkit
**Researched:** 2026-05-07
**Confidence:** HIGH

## Stance Summary (Read This First)

This project sits at the intersection of two product categories:

1. **A SCIM v2 server.** The features users expect here are dictated by RFC 7643 (core schema) and RFC 7644 (protocol). Compliance is binary — partial implementations are unusable for the obvious customers (interop with Okta/Azure AD/etc. as identity sources). The project goal locks in **full SCIM v2 coverage in v1**; this rules out a "spec subset" MVP.
2. **A code generator.** The features users expect here are a clean definition format, idempotent generation, generated code that compiles and tests, and emitted artifacts that fit a normal Go workflow (modules, build, test).

Three opinionated stances follow from those two framings:

- **Filter grammar coverage stance: full RFC 7644 §3.4.2 grammar in v1.** Subsetting the filter grammar is a cliff because IdPs send arbitrary filters and the grammar is a closed formal grammar — partial parsers fail on real client traffic. Cost is one-time (write the parser once), so subset has no ongoing payoff.
- **PATCH path-expression stance: full path-expression PATCH in v1, but flagged as the single highest-complexity feature.** RFC 7644 §3.5.2 PATCH with value-path filters (`emails[type eq "work"].value`) is the most complex single requirement in the spec. It must work end-to-end (parse → validate against schema → translate to SQL update) before the server is shippable. Treat it as its own roadmap phase.
- **Code-gen value-prop stance: custom resource types, emitted DDL, and emitted compliance suite are the differentiators that justify code-gen over runtime interpretation.** Without these, generation is just a build-time optimization with no user-visible benefit.

## Feature Landscape

### Table Stakes (RFC-mandated; server is non-compliant without these)

These map directly to RFC 7643 / 7644 requirements. Missing any of them means the generated server fails interop with mainstream IdPs.

#### Resource endpoints & CRUD

| Feature | Why Expected | Complexity | Notes |
|---------|--------------|------------|-------|
| `POST /Users`, `GET /Users/{id}`, `PUT /Users/{id}`, `DELETE /Users/{id}` | RFC 7644 §3.3–§3.6; without it there is no SCIM server | LOW | Per-resource-type generated handlers; PUT is full replace, not partial |
| `POST /Groups`, `GET /Groups/{id}`, `PUT /Groups/{id}`, `DELETE /Groups/{id}` | RFC 7643 §4.2; Group is one of two mandatory core resources | LOW | Identical mechanics to Users with different schema |
| `GET /Users` (list with query params) | RFC 7644 §3.4.2 — listing is the main read path for IdPs | MEDIUM | Driven by filter/sort/pagination/projection, see below |
| `GET /Groups` (list with query params) | Same | MEDIUM | Same |
| `POST /Users/.search`, `POST /Groups/.search`, `POST /.search` | RFC 7644 §3.4.3 — POST-based search for filters too long for URL | LOW | Same logic as GET list, body-parsed instead |

#### Discovery endpoints

| Feature | Why Expected | Complexity | Notes |
|---------|--------------|------------|-------|
| `GET /Schemas` | RFC 7644 §4 — IdPs hit this first to learn what attributes exist | LOW | Statically generated from definition; no runtime construction |
| `GET /Schemas/{id}` | Same | LOW | Same |
| `GET /ResourceTypes` | RFC 7644 §4 — maps endpoint paths to schemas | LOW | Statically generated |
| `GET /ResourceTypes/{id}` | Same | LOW | Same |
| `GET /ServiceProviderConfig` | RFC 7644 §4 — declares which optional features (filter, bulk, etag, sort, patch) the server supports | LOW | Static JSON; values derived from generator config (e.g., bulk on/off) |

#### Filter grammar (RFC 7644 §3.4.2.2)

| Feature | Why Expected | Complexity | Notes |
|---------|--------------|------------|-------|
| Comparison operators: `eq`, `ne`, `co`, `sw`, `ew`, `gt`, `lt`, `ge`, `le`, `pr` | Spec-mandated; IdPs use all of them | MEDIUM | Each maps to a SQL operator/predicate; `co/sw/ew` need LIKE; `pr` is IS NOT NULL |
| Logical operators: `and`, `or`, `not`, parentheses for grouping | Same | MEDIUM | Standard recursive-descent parse; AST → SQL WHERE |
| Value-path expressions: `emails[type eq "work"]` | Spec-mandated; very common in real traffic | HIGH | Filters on multi-valued complex sub-attributes; in normalized SQL this means JOIN on the child table with the inner predicate |
| Attribute-path filters with extension URIs: `urn:ietf:params:scim:schemas:extension:enterprise:2.0:User:employeeNumber eq "X"` | Spec-mandated; required for EnterpriseUser interop | MEDIUM | Parser must accept full URN-prefixed paths |
| Case-insensitivity for string types except `caseExact` attributes | RFC 7643 §2.2; `id` is caseExact, most others are not | MEDIUM | Drives which DB column collation / `LOWER()` wrapping to emit |

#### Sort, pagination, projection

| Feature | Why Expected | Complexity | Notes |
|---------|--------------|------------|-------|
| `sortBy`, `sortOrder=ascending|descending` | RFC 7644 §3.4.2.3 | LOW | Maps to SQL ORDER BY; sortBy may be dotted path |
| `startIndex` (1-based), `count`, response `totalResults` and `itemsPerPage` | RFC 7644 §3.4.2.4 | LOW | Standard offset pagination; `totalResults` requires a COUNT query |
| `attributes` and `excludedAttributes` projection | RFC 7644 §3.9 | MEDIUM | Server still has to honor `returned: always` regardless; complex attr projection means partially returning sub-attributes |

#### PATCH (RFC 7644 §3.5.2)

| Feature | Why Expected | Complexity | Notes |
|---------|--------------|------------|-------|
| `PATCH` with `add`/`replace`/`remove` ops, no path (whole-resource merge) | Spec-mandated | MEDIUM | Treat body as partial resource; merge into current; honor mutability |
| `PATCH` with attribute path: `path: "name.givenName"` | Spec-mandated | MEDIUM | Resolve dotted path against schema, apply value at that path |
| `PATCH` with value-path filter: `path: "emails[type eq \"work\"].value"` | Spec-mandated; **the hardest single feature in the spec** | HIGH | Filter must select element(s) of multi-valued attr, op applies to the filtered subset; `remove` removes elements; `add` on existing path replaces; `replace` replaces matched values |
| Atomicity: all ops succeed or resource is restored | RFC 7644 §3.5.2 | MEDIUM | Requires DB transaction wrapping the whole PATCH |
| Mutability enforcement during PATCH (immutable rejected, readOnly ignored, writeOnly accepted but never returned) | RFC 7643 §2.2 + RFC 7644 §3.5.2 | MEDIUM | Generator emits per-attribute checks; failure returns `mutability` scimType |

#### Schema-driven attribute behavior

| Feature | Why Expected | Complexity | Notes |
|---------|--------------|------------|-------|
| `mutability`: `readOnly`, `readWrite`, `immutable`, `writeOnly` enforcement | RFC 7643 §2.2 | MEDIUM | Generator emits per-field check sites at create/replace/patch boundaries |
| `returned`: `always`, `never`, `default`, `request` honored on every response | RFC 7643 §2.2 | MEDIUM | `always` can never be excluded; `never` (e.g. password) can never be returned; interacts with `attributes`/`excludedAttributes` |
| `uniqueness`: `none`, `server`, `global` enforcement | RFC 7643 §2.2 | MEDIUM | `server`/`global` require DB unique index; conflict returns 409 with `uniqueness` scimType |
| Multi-valued complex `primary: true` invariant (at most one element per attribute may be primary) | RFC 7643 §2.4 | MEDIUM | Must be enforced on every mutation path (create, replace, patch) — easy to miss in PATCH |
| `required` enforcement on create | RFC 7643 §2.2 | LOW | Standard validation |
| `caseExact` semantics for string equality and uniqueness | RFC 7643 §2.3.1 | MEDIUM | Drives DB collation choice and filter SQL emission |
| Type system: `string`, `boolean`, `decimal`, `integer`, `dateTime`, `binary`, `reference`, `complex` (and the multi-valued flag on each) | RFC 7643 §2.3 | MEDIUM | Each type maps to a Go type and a SQL column type; DDL emitter must cover all |

#### Concurrency & versioning

| Feature | Why Expected | Complexity | Notes |
|---------|--------------|------------|-------|
| `meta.version` (ETag) on every resource | RFC 7644 §3.14 + RFC 7643 §3.1 | LOW | Hash of resource state or monotonic version column; emit weak ETag (`W/"..."`) |
| `If-Match` on PUT/PATCH/DELETE → 412 on mismatch | RFC 7644 §3.14 | LOW | Optimistic concurrency; refuses overwrite when version moved |
| `If-None-Match: *` on POST → 412 if resource already exists | Less common but spec-allowed | LOW | Cheap; do it |
| `ETag` response header echoing `meta.version` | RFC 7644 §3.14 | LOW | Same value as `meta.version` |

#### Common attributes (RFC 7643 §3.1)

| Feature | Why Expected | Complexity | Notes |
|---------|--------------|------------|-------|
| `id` — server-issued, immutable, caseExact, always returned | Spec-mandated | LOW | Generator emits UUID generation in create handler |
| `externalId` — client-issued, optional | Spec-mandated | LOW | Plain column |
| `schemas` array on every resource | Spec-mandated | LOW | Static per resource type, plus extension URIs when extensions are populated |
| `meta.resourceType`, `meta.created`, `meta.lastModified`, `meta.location`, `meta.version` | Spec-mandated; always returned | LOW | Server-managed; populated on create/update |

#### EnterpriseUser extension (RFC 7643 §4.3)

| Feature | Why Expected | Complexity | Notes |
|---------|--------------|------------|-------|
| `urn:ietf:params:scim:schemas:extension:enterprise:2.0:User` extension on User | Most IdPs (Okta, Azure AD, OneLogin) populate this; functionally table stakes despite being "an extension" | LOW | Generator treats it as a registered extension on the User resource; generic extension mechanism handles arbitrary extensions, this is just the canonical one |

#### Errors, content type, status codes

| Feature | Why Expected | Complexity | Notes |
|---------|--------------|------------|-------|
| `application/scim+json` content type on requests and responses | RFC 7644 §3.1 | LOW | Single header; most clients still send `application/json` and servers should accept both for input but emit `application/scim+json` |
| Error JSON format with `schemas`, `status`, `scimType`, `detail` | RFC 7644 §3.12 | LOW | Per-error-class scimType values: `uniqueness`, `mutability`, `tooMany`, `invalidFilter`, `invalidPath`, `invalidValue`, `invalidVers`, `invalidSyntax`, `noTarget`, `sensitive` |
| Status codes: 200/201/204 for success; 304 (with ETag); 400/401/403/404/409/412/413/500/501 | RFC 7644 §3.12 | LOW | Each error type has a specific status; `501 Not Implemented` for unsupported optional features (e.g., bulk if generator emits without it) |
| `Location` header on 201 Created pointing to `meta.location` | RFC 7644 §3.3 | LOW | Trivial |

#### Bulk operations (RFC 7644 §3.7)

| Feature | Why Expected | Complexity | Notes |
|---------|--------------|------------|-------|
| `POST /Bulk` accepting an array of operations with `bulkId` cross-references | Spec-listed in `/ServiceProviderConfig`; declared in scope by PROJECT.md | HIGH | Operations execute sequentially; `bulkId` lets a later op reference a resource created by an earlier op (e.g., add a just-created user to a group); requires sub-router that internally calls the same per-resource handlers; `failOnErrors` short-circuit; partial-success reporting |

#### `/Me` alias

| Feature | Why Expected | Complexity | Notes |
|---------|--------------|------------|-------|
| `/Me` redirects/aliases to `/Users/{authenticated-id}` | RFC 7644 §3.11 | LOW (but blocked) | **Cannot implement in v1** — auth is out of scope per PROJECT.md, and `/Me` is meaningless without an authenticated principal. Either skip and advertise as unsupported in `/ServiceProviderConfig`, or document that the user must wire it in their main. Recommend: don't generate it; document the gap. |

### Differentiators (Code-Gen Value-Prop)

These are the features that make a code-generation toolkit better than a runtime SCIM library. They are the answer to "why generate instead of interpret a schema at runtime?"

| Feature | Value Proposition | Complexity | Notes |
|---------|-------------------|------------|-------|
| **Custom resource types via definition** | Users can declare arbitrary resources (Devices, Apps, Tenants, …) alongside User/Group and get the full SCIM treatment — handlers, schema endpoint, repo, DDL, tests, compliance suite — as if they were built-in. The legacy library required hand-wiring per-resource code on top of a generic schema. | MEDIUM | Definition input describes attributes/sub-attributes; generator iterates over all declared resources uniformly; mandatory User/Group are just pre-bundled definitions |
| **Generated per-resource SQL repositories (no runtime tree-to-SQL translation)** | Eliminates the legacy's tree-to-relational impedance. Each resource gets a typed Go repo with hand-shaped SQL. No reflection, no generic property tree at runtime. | HIGH | One-time generator complexity, ongoing runtime simplicity. This is the single biggest win over the legacy architecture per PROJECT.md drivers |
| **DDL / migration emission** | Generator emits SQL CREATE TABLE / CREATE INDEX scripts alongside the server. User runs them once on their DB, no separate "design your schema to match the SCIM model" step. | MEDIUM | One DDL file per driver (SQLite first, Postgres/MySQL deferred); table layout: parent table per resource, child table per multi-valued complex attribute |
| **Generated SCIM v2 compliance test suite** | Generator emits tests that hit every required endpoint with every required behavior, run against the user's compiled binary + SQLite. Users can prove compliance in CI without writing the tests. | HIGH | Test suite is itself derived from the schema (knows attribute mutability, uniqueness, etc.) so it can produce schema-accurate fixtures and expectations |
| **Generated unit + integration tests** | Beyond the compliance suite: generator emits unit tests for handlers, repos, filter compilation, and integration tests against SQLite. Same one-time generation cost, ongoing user benefit. | MEDIUM | Largely templated; adds confidence that generation is correct |
| **Pluggable observability interfaces in generated code** | Logging/tracing/metrics surfaces are emitted as Go interfaces only — no OTel, no Prometheus dependency dragged in. Users wire concrete implementations in their main. | LOW | Define a small interface set (Logger, Tracer, Metrics or a single Observer); inject via Config; generator emits no-op default |
| **Plain `http.Handler` output, no opinionated mux** | User mounts the generated server on chi/gorilla/gin/std net/http/whatever they already use. Library does not pick their HTTP stack. | LOW | Generator emits a `func New(cfg Config) http.Handler`; routing is internal but does not leak |
| **Typed `Config` struct (library-style) + CLI-emitted `main` (12-factor)** | Library users get compile-time type safety on configuration; CLI users who run `scimgen new` get a generated `main.go` that wires environment variables and flags into the same `Config`. | LOW | Two thin layers; CLI is a wrapper over the library |
| **Tenant-id-aware persistence layer (single-tenant in v1, multi-tenant ready)** | Generated repos carry a tenant-id column from day one. v1 generator hardcodes a single tenant value; v2 will enable multi-tenant routing without rewriting the repos. | LOW (v1) / MEDIUM (when activated) | Per PROJECT.md: out of scope at the request layer in v1, but the persistence schema must support it |
| **Pluggable SQL driver interface; SQLite shipped first** | Same generator, multiple backends. SQLite ships in v1; Postgres and MySQL slot in behind the same interface later. | MEDIUM (v1 — design the interface right) | The interface must be expressive enough to cover dialect differences (RETURNING, JSON columns, collations) without leaking driver specifics into handlers |
| **Strict-spec stance (no IdP quirk hacks in generated code)** | The generated server is a clean reference implementation. No `if Microsoft { capitalize }` branches in handlers. Quirks become a future definition-level "compatibility profile" feature. | LOW (it's the absence of code) | Per PROJECT.md "Out of Scope" |

### Anti-Features (Explicitly Not Building)

These are features that look obvious but the project has decided against, either in PROJECT.md or as a consequence of the code-gen architecture.

| Anti-Feature | Why Avoided | Alternative |
|--------------|-------------|-------------|
| **IdP-specific quirk accommodations** (Microsoft capital-letter attribute names, Azure AD non-standard PATCH shapes, Okta-specific filter dialect) | Per PROJECT.md "Out of Scope". Quirk hacks rotted the legacy codebase; embedding them per-IdP creates a combinatorial maintenance burden. | Strict spec compliance in v1. Future: definition-level compatibility profiles that the generator switches on, keeping the generated code clean per profile |
| **Authentication / authorization in the generated server** | Per PROJECT.md "Out of Scope". Auth is a deployment concern (gateway, mesh, sidecar) and bundling it would force a choice the user already made elsewhere. | Document that an upstream gateway is expected to terminate auth; emit hooks (middleware slot in `Config`) for users who want to inject their own |
| **Multi-tenancy at the request layer in v1** | Per PROJECT.md "Out of Scope" for v1. Tenant routing has knock-on effects for filter/pagination/uniqueness scoping that are easier to design once than to retrofit twice. | Persistence layer is tenant-id aware from day one; v2 adds the request-layer routing additively |
| **Group-membership change eventing (legacy `groupsync`/RabbitMQ)** | Per PROJECT.md "Out of Scope". Not a SCIM spec requirement; it was an implementation quirk of the legacy project. | Users who need group-membership eventing wire it themselves over their existing event bus; SCIM PATCH on `members` is the source of truth |
| **Runtime schema reload / hot-swap** | Defeats the entire code-gen value-prop. If schemas are hot-swappable at runtime, you are back to the legacy generic-tree design. | Re-run the generator and redeploy. Schema changes are a build-time concern |
| **Generic ORM passthrough (e.g., GORM auto-migration)** | A generic ORM forces back the impedance mismatch the code-gen approach exists to escape, and SCIM's mutability/uniqueness/multi-valued-complex semantics are not expressible as plain ORM annotations. | Hand-shaped, generated SQL behind a narrow driver interface. `database/sql` directly, or `sqlc`-style typed queries |
| **Schema-evolution diffing / migrations between generator runs** | Per PROJECT.md "Out of Scope" for v1. Schema-diff tooling is its own large project. | Generator emits fresh DDL on every run; user owns diffing using their existing migration tool (golang-migrate, atlas, sqitch). v2 may add hooks |
| **Concrete observability implementations (OTel SDK, Prometheus client, zerolog/zap dependency)** | Per PROJECT.md "Out of Scope". Forces a vendor dependency on every consumer. | Pluggable interfaces only; user wires concrete impls. Default is a no-op |
| **Concrete non-SQLite SQL drivers in v1** (Postgres, MySQL) | Per PROJECT.md "Out of Scope" for v1. Driver count is a force multiplier on test surface; ship one, prove the interface, then add. | Design the SQL driver interface with Postgres/MySQL in mind; ship SQLite as the reference; second driver is a follow-on milestone |
| **Backward compatibility with legacy `go-scim`** | Per PROJECT.md "Out of Scope". Different mental model entirely (generated typed repos vs. generic property tree); a compat shim would carry the legacy's complexity into v3. | Clean break; new module path; legacy frozen under `.legacy/` for reference |
| **Built-in HTTP mux opinion** (chi, gorilla, gin chosen by the library) | Per PROJECT.md "Out of Scope". Forces a routing-library dependency; users already have one. | Emit a plain `http.Handler`; user mounts wherever |
| **Performance / scale targets in v1** | Per PROJECT.md "Out of Scope". Optimizing prematurely against unknown workloads warps the design. | Correctness first; the generated SQL layout (per-resource tables, indexed unique columns) is naturally efficient and we measure once we have real users |
| **Subset filter grammar** | Filter grammar is closed; partial parsers reject real IdP traffic with `invalidFilter`, which is worse than not advertising filter at all | Implement the full RFC 7644 §3.4.2.2 grammar in v1. Cost is one parser; benefit is interop |
| **Whole-resource-only PATCH (no path expressions)** | Real IdPs send path-expression PATCHes constantly (every Azure AD group-membership change is one). Path-less-only PATCH means failing real traffic | Implement full path expressions including value-path filters in v1; flag this as the highest-complexity work item |

## Feature Dependencies

```
[Definition format & parser]
    └──requires──> [Internal model (resource type, attribute, schema)]
                       └──requires──> [Code generator core]
                                          ├──emits──> [HTTP handlers per resource type]
                                          │              └──requires──> [JSON (de)serialization]
                                          │              └──requires──> [Mutability/returned/uniqueness enforcement]
                                          ├──emits──> [SQL repositories per resource type]
                                          │              └──requires──> [SQL driver interface]
                                          │              └──requires──> [DDL emission]
                                          ├──emits──> [Filter compiler (SCIM filter AST → SQL WHERE)]
                                          │              └──requires──> [Filter grammar parser]
                                          ├──emits──> [Path-expression engine (for PATCH)]
                                          │              └──requires──> [Filter grammar parser (value-path subset)]
                                          ├──emits──> [Discovery endpoints (/Schemas, /ResourceTypes, /ServiceProviderConfig)]
                                          ├──emits──> [Compliance test suite]
                                          └──emits──> [Unit + integration tests]

[/Schemas]              ──precedes──> [/ResourceTypes]
                                            └──precedes──> [/ServiceProviderConfig]
                                                                   (which advertises filter, patch, bulk, etag, sort capabilities)

[Filter compiler]       ──precedes──> [GET /Users list with ?filter=]
                        ──precedes──> [PATCH path-expression engine (value-path filters reuse the parser)]
                        ──precedes──> [POST /Bulk]   (bulk ops include filtered queries)

[CRUD on a single resource type]
    └──precedes──> [Bulk operations]   (bulk dispatches to per-resource handlers)
    └──precedes──> [PATCH]             (PATCH semantics build on the same write path)

[ETag emission on responses]
    └──precedes──> [If-Match enforcement on PUT/PATCH/DELETE]

[Mutability enforcement]
    └──precedes──> [PATCH]   (PATCH is the hardest place to get mutability right)

[Multi-valued complex `primary` invariant]
    └──enforced-at──> [Create, Replace, Patch]   (must hold on every mutation path)

[SQLite driver]
    └──proves──> [SQL driver interface design]
                       └──unblocks──> [Postgres/MySQL drivers (post-v1)]

[Custom resource types]
    └──requires──> [Definition format expressive enough to declare arbitrary attributes]
    └──requires──> [Generator that treats User/Group as just two pre-bundled custom types]
```

### Dependency Notes

- **Discovery endpoints depend on the schema model, not on CRUD.** They can be implemented and tested standalone, very early. Doing so de-risks the schema model before handlers consume it.
- **Filter compiler is upstream of three big features** (list endpoints, PATCH path expressions, bulk). Build it once, well, before any of those.
- **PATCH path expressions reuse the filter parser's value-path subset.** Do not build two parsers; the filter grammar's bracketed predicate is the same grammar PATCH paths use.
- **`primary` invariant is a mutation-time invariant, not a schema-time one.** Easy to enforce in create/replace; easy to forget in PATCH. Generator must emit the check at every write path.
- **`returned: never` (e.g., `password`) is a serialization invariant.** It must be enforced in the JSON serializer, not just in the handler — otherwise it leaks in error responses, debug logs, etc.
- **ETag/If-Match is independently shippable** but pointless without write endpoints. Sequence after CRUD.
- **Bulk depends on every other write endpoint.** It is essentially a fan-out over the existing handlers; ship it last in the protocol layer.
- **Compliance suite depends on every protocol feature it tests.** It can be built incrementally alongside each feature and consolidated at the end.
- **DDL emission is independent of HTTP** but must agree with the SQL repos on column names/types. Generate them from the same internal model.

## MVP Definition

**The locked-in scope per PROJECT.md is "full SCIM v2 coverage in v1."** That removes the usual "subset for MVP" lever. What remains is sequencing — the order in which features land within v1 — and what stays out (deferred) vs. what ships.

### Launch With (v1) — full SCIM v2 + code-gen value-prop

Definition & generator core:
- [ ] Definition format (fluent Go builder per working hypothesis) — every other feature is downstream of this
- [ ] Internal model (ResourceType, Schema, Attribute, type system, mutability/returned/uniqueness)
- [ ] Code generator core (template engine, per-resource emission loop)
- [ ] Multi-module workspace layout (generator vs. runtime support library)
- [ ] CLI (`scimgen` or similar) as a thin wrapper over the library

Schema layer (RFC 7643):
- [ ] Built-in User schema
- [ ] Built-in Group schema
- [ ] Built-in EnterpriseUser extension
- [ ] Custom resource types declared via definition (the differentiator)
- [ ] All attribute characteristics: mutability, returned, uniqueness, required, caseExact, multiValued, type
- [ ] Multi-valued complex `primary` invariant enforcement

Protocol layer (RFC 7644):
- [ ] CRUD on `/Users` and `/Groups` (plus any custom resources from the definition)
- [ ] Discovery: `/Schemas`, `/ResourceTypes`, `/ServiceProviderConfig`
- [ ] List with full filter grammar, sort, pagination, projection (`attributes`/`excludedAttributes`)
- [ ] `POST /.search` and per-resource `.search`
- [ ] PATCH with no-path, attribute-path, and value-path-filter forms (the highest-complexity feature)
- [ ] `/Bulk` with `bulkId` cross-references and `failOnErrors`
- [ ] ETag emission, `If-Match`, `If-None-Match`
- [ ] `application/scim+json` content type
- [ ] RFC 7644 §3.12 error format with full set of `scimType` values
- [ ] Correct status codes (200/201/204/304/400/401/403/404/409/412/500/501)

Persistence layer:
- [ ] SQL driver interface (designed to absorb Postgres/MySQL later)
- [ ] SQLite driver implementation
- [ ] Generated per-resource SQL repos (no runtime tree-to-SQL)
- [ ] DDL / migration script emission
- [ ] Tenant-id-aware schema (single tenant value in v1)

Generated artifacts:
- [ ] Generated unit tests
- [ ] Generated integration tests against SQLite
- [ ] Generated SCIM v2 compliance test suite
- [ ] Generated `Config` struct
- [ ] Generated `main.go` (CLI workflow) with env/flag wiring
- [ ] Pluggable observability interfaces (no concrete impl shipped)

Distribution:
- [ ] Plain `net/http` `http.Handler` output (no mux baked in)
- [ ] MIT license

### Add After Validation (v1.x)

- [ ] Postgres driver behind the existing SQL driver interface — trigger: SQLite design is proven and the interface has been stress-tested against a second dialect
- [ ] MySQL driver — trigger: Postgres lands cleanly
- [ ] `/Me` alias — trigger: an auth story emerges (likely a `Config` hook for "who is the current user?")
- [ ] Definition format alternatives (YAML/JSON) if the fluent Go builder reveals friction in user research — trigger: actual user feedback, not speculation
- [ ] Schema-evolution / migration-diff helpers — trigger: users hit pain re-running the generator on existing databases

### Future Consideration (v2+)

- [ ] Multi-tenancy at the request layer (persistence is already ready) — trigger: a user who needs it
- [ ] Compatibility profiles in the definition (Microsoft / Okta / Azure AD quirks selectable per profile, generator emits per-profile branches in clean code) — trigger: enough demand to justify the maintenance cost; only attempt once the v1 strict-spec generator is stable
- [ ] Observability adapter packages (separate modules: `go-scim-otel`, `go-scim-prom`) so users opting into a stack don't need to write the glue — trigger: a stable observability interface and concrete user requests
- [ ] Authentication middleware adapters (separate modules; same pattern) — trigger: same as above
- [ ] Group-membership change eventing as an opt-in emitter — trigger: a user whose architecture genuinely needs it (unclear there is one)

## Feature Prioritization Matrix

| Feature | User Value | Implementation Cost | Priority |
|---------|------------|---------------------|----------|
| Definition format & parser | HIGH (gates everything) | MEDIUM | P1 |
| Code generator core (template engine, emission loop) | HIGH (gates everything) | HIGH | P1 |
| Built-in User/Group/EnterpriseUser schemas | HIGH (compliance) | LOW | P1 |
| Custom resource types | HIGH (differentiator) | MEDIUM | P1 |
| CRUD on resources (POST/GET/PUT/DELETE) | HIGH (compliance) | LOW | P1 |
| Discovery endpoints (`/Schemas`, `/ResourceTypes`, `/ServiceProviderConfig`) | HIGH (compliance) | LOW | P1 |
| Mutability/returned/uniqueness enforcement | HIGH (compliance) | MEDIUM | P1 |
| Filter grammar (full RFC 7644 §3.4.2.2) | HIGH (compliance) | MEDIUM | P1 |
| Sort, pagination, projection | HIGH (compliance) | LOW | P1 |
| Multi-valued complex `primary` invariant | HIGH (compliance, easy to miss) | MEDIUM | P1 |
| ETag / If-Match / If-None-Match | HIGH (compliance) | LOW | P1 |
| Error format + scimType + status codes | HIGH (compliance) | LOW | P1 |
| `application/scim+json` content type | HIGH (compliance) | LOW | P1 |
| **PATCH with full path-expression support** | HIGH (compliance) | **HIGH** | **P1 (deserves its own phase)** |
| `/Bulk` operations | HIGH (compliance) | HIGH | P1 |
| `POST /.search` | MEDIUM (compliance) | LOW | P1 |
| SQL driver interface | HIGH (architecture) | MEDIUM | P1 |
| SQLite driver | HIGH (reference impl) | MEDIUM | P1 |
| Generated per-resource SQL repos | HIGH (differentiator) | HIGH | P1 |
| DDL / migration emission | HIGH (differentiator) | MEDIUM | P1 |
| Generated unit + integration tests | MEDIUM (differentiator) | MEDIUM | P1 |
| Generated SCIM v2 compliance suite | HIGH (differentiator) | HIGH | P1 |
| Pluggable observability interfaces | MEDIUM (differentiator) | LOW | P1 |
| Typed `Config` struct + CLI-emitted `main` | MEDIUM (differentiator) | LOW | P1 |
| Tenant-id-aware persistence (v1: single tenant) | MEDIUM (future-proofing) | LOW | P1 |
| Plain `http.Handler` output | HIGH (differentiator) | LOW | P1 |
| `/Me` alias | LOW (auth-blocked) | LOW | P3 |
| Postgres driver | HIGH (post-v1 demand) | MEDIUM | P2 |
| MySQL driver | MEDIUM | MEDIUM | P2 |
| YAML/JSON definition alternatives | LOW (revisit after user feedback) | MEDIUM | P3 |
| Multi-tenancy at request layer | MEDIUM (future) | HIGH | P3 |
| Compatibility profiles (IdP quirks) | MEDIUM (future) | HIGH | P3 |
| Concrete observability adapters (OTel/Prom) | LOW (users wire their own) | MEDIUM | P3 |
| Schema-evolution migration diffing | MEDIUM (future) | HIGH | P3 |

**Priority key:**
- P1 — Must ship in v1
- P2 — Add after v1 validates
- P3 — Future / out of scope for v1 per PROJECT.md

## "Competitor" Feature Analysis

There are not many SCIM-server-as-code-gen projects to compare against. The relevant comparison points are (a) the legacy `go-scim` (this project's predecessor), (b) other Go SCIM libraries, and (c) other code-generation toolkits (gRPC/protobuf, OpenAPI generators, sqlc) for stylistic reference.

| Feature | Legacy `go-scim` (`.legacy/`) | Other Go SCIM libs (e.g., elimity-com/scim, scim2/filter-parser) | Code-gen reference (sqlc, oapi-codegen) | Our Approach |
|---------|------------------------------|------------------------------------------------------------------|------------------------------------------|--------------|
| Resource model | Generic in-memory tree, schema-aware nodes | Generic struct, marshal/unmarshal | N/A (different domain) | Generated typed Go structs per resource type, no runtime tree |
| Persistence | Pluggable interface; MongoDB adapter; SQL adapters considered hard | Mostly user-supplied; in-memory examples | Generated typed queries (sqlc) | Generated per-resource SQL repos behind a pluggable driver interface; SQLite reference; no MongoDB |
| Definition format | JSON schema files registered at startup | Code or JSON, varies | sqlc: SQL+queries.sql; oapi-codegen: OpenAPI YAML | Fluent Go builder (working hypothesis); JSON/YAML not ruled out |
| Custom resource types | Possible (define a schema) but boilerplate-heavy | Possible, varying levels of friction | N/A | First-class; declared in the definition, generator treats them identically to User/Group |
| PATCH path expressions | Implemented via crud/expr package | Coverage varies; some libraries punt on value-path filters | N/A | Full RFC 7644 §3.5.2 including value-path filters; flagged as the single hardest feature |
| Filter grammar | Full grammar with expr compiler | Varies; some are subset-only | N/A | Full RFC 7644 §3.4.2.2 grammar; subsetting rejected |
| Bulk | Not detected as fully implemented in legacy | Often deferred | N/A | Full `/Bulk` with `bulkId` cross-references in v1 |
| ETag / If-Match | Limited / not first-class | Varies | N/A | First-class; emitted on every response and enforced on every conditional write |
| Compliance test suite | Internal test suite, not emitted to consumers | Not typically shipped | sqlc/oapi-codegen ship test scaffolds | Emitted into the user's generated repo; runs against their compiled binary |
| HTTP routing | Custom in `cmd/api` | Varies (chi, gorilla, custom) | N/A | Plain `http.Handler`, no mux dependency |
| Auth | Not in legacy | Not typically in scope | N/A | Out of scope; gateway terminates upstream |
| IdP quirks | Some hacks accumulated | Varies; some libraries embed Microsoft accommodations | N/A | Strict spec only in v1; future "compatibility profile" feature in the definition |
| Observability | zerolog injected | Varies | N/A | Pluggable interfaces, no concrete impl shipped |

## Sources

**Primary (HIGH confidence):**
- RFC 7643 — System for Cross-domain Identity Management: Core Schema (https://datatracker.ietf.org/doc/html/rfc7643) — verified directly for: User/Group/EnterpriseUser attributes, mutability/returned/uniqueness enums, multi-valued complex `primary` invariant, common attributes (id/externalId/schemas/meta), type system
- RFC 7644 — System for Cross-domain Identity Management: Protocol (https://datatracker.ietf.org/doc/html/rfc7644) — verified directly for: endpoint set, HTTP methods, PATCH operations and path expressions, filter grammar, sort/pagination/projection, ETag/If-Match/If-None-Match, error format §3.12, status codes, content type, bulk operations §3.7

**Project context (HIGH confidence):**
- `.planning/PROJECT.md` — locked scope, locked exclusions, three rewrite drivers (feasibility/efficiency/portability), key decisions
- `.planning/codebase/ARCHITECTURE.md` — legacy feature surface for reference (what the old generic-tree implementation covered, where it had gaps, what the SQL-impedance pain looked like)

**Confidence assessment:**

| Area | Confidence | Reason |
|------|------------|--------|
| Spec endpoint coverage | HIGH | Direct RFC verification |
| PATCH path-expression complexity claim | HIGH | Cross-checked against RFC 7644 §3.5.2 examples and against legacy `crud/expr/` implementation |
| Filter grammar coverage stance | HIGH | RFC defines the grammar formally; partial parsers fail real traffic — this is well-known in the SCIM community |
| Code-gen-specific feature value-prop | HIGH | Directly grounded in PROJECT.md's three rewrite drivers and the legacy's documented impedance pain |
| Anti-features list | HIGH | All anti-features are explicitly enumerated in PROJECT.md "Out of Scope" |
| Bulk operations complexity | MEDIUM | Spec is clear, but real-world cross-IdP variation in `bulkId` usage may surface edge cases not visible from the RFC alone |
| `/Me` recommendation (skip in v1) | MEDIUM | Spec-grounded reasoning (auth required), but the recommendation is a judgment call — alternative is to wire a hook |
| EnterpriseUser as "table stakes despite being an extension" | MEDIUM | RFC labels it an extension; community/IdP behavior makes it de-facto required. State as opinion, not as RFC mandate |

---
*Feature research for: SCIM v2 server code-generation toolkit*
*Researched: 2026-05-07*

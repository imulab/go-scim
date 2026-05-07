# Project Research Summary

**Project:** go-scim v3 (`next` branch) — Go code-generation toolkit emitting spec-compliant SCIM v2 servers with SQL persistence
**Domain:** Go code-generator toolkit (definition → typed Go server + SQL repos + DDL + tests)
**Researched:** 2026-05-07
**Confidence:** HIGH

## Executive Summary

go-scim v3 is a build-time code generator that takes a SCIM resource definition and emits a plain `net/http` `http.Handler` server, per-resource typed Go structs, per-resource `database/sql` repositories, SQLite DDL, and a SCIM v2 compliance test suite. The four research dimensions converge tightly: this is a **classic three-stage codegen pipeline** (Definition → IR → Multi-Emitter → gofmt) with a **two-module split** (heavyweight generator vs. lean runtime support library imported by generated servers). The pattern is well-trodden across `ent`, `sqlc`, `oapi-codegen`, `kubebuilder`, and `protoc-gen-connect-go`; the SCIM-specific work is the IR shape, the filter/PATCH grammars, and the per-attribute mutability/returned/uniqueness emission.

The recommended approach is opinionated and converges across all four researchers: ship a **fluent Go builder** definition format (in-language, type-checked, refactor-friendly), validate everything at IR time (so user typos error at `go generate`, never in generated code), emit per-resource typed structs and repos against `database/sql` directly (no embedded ORM, no runtime tree), default to **`modernc.org/sqlite`** (CGo-free) as the v1 reference driver, and put the PATCH engine, filter parser, ETag helpers, and SCIM error envelope in a separate runtime support module that the generated server imports. The CLI (`scimgen`, Cobra-based) is a thin wrapper over the library API.

The dominant risk is **architectural relapse into the legacy generic-tree model** — the rewrite exists precisely because that model was incompatible with SQL and required reactive subscribers (`ExclusivePrimarySubscriber`) for SCIM invariants. Every researcher independently flags this as catastrophic and as the single most important rule to encode early. Two adjacent risks: **scope creep into IdP-quirk accommodations** (Microsoft Entra capitalization, Okta non-standard fields) which legacy CONCERNS.md §3 documents as a brittleness driver and which PROJECT.md explicitly excludes from v1; and **PATCH path-expression correctness** (RFC 7644 §3.5.2 with value-path filters like `emails[type eq "work"].value`), which is universally identified as the single hardest piece of work and deserves its own roadmap phase with a dedicated test corpus.

## Key Findings

### Recommended Stack

The stack is split into two `go.mod` modules: a heavyweight **generator** (only built by toolkit developers and CLI users) and a lean **runtime support library** (imported by every generated server, deliberately tiny so consumers can audit every dep). Generator stack centers on Go 1.24 with tool directives, Cobra for the CLI, testify (require only) for tests, and a debate over emission engine (see Architecture section). Runtime stack is essentially stdlib (`net/http`, `database/sql`, `log/slog`, `encoding/json`) plus `google/uuid` and `modernc.org/sqlite`.

**Core technologies:**
- **Go 1.24+** — tool directives in `go.mod`, ServeMux pattern syntax (1.22+) covers SCIM's REST shape with no third-party router
- **`modernc.org/sqlite` v1.46.1+** (default runtime driver) — CGo-free, single static binary, December 2025 prepared-statement fix closes the historical perf gap
- **`mattn/go-sqlite3`** (opt-in alternative) — documented swap for write-heavy workloads where users accept CGo
- **`net/http` + `ServeMux`** in generated code — plain `http.Handler` contract; user mounts on chi/gorilla/gin at the application level
- **`database/sql`** directly, no ORM — code-gen IS the ORM; sqlc/ent/sqlboiler/gorm are explicitly rejected because they are competing generators
- **`log/slog`** stdlib — pluggable observability via stdlib handlers; no third-party logger ships into generated code
- **`spf13/cobra`** for the `scimgen` CLI — de-facto standard, thin wrapper over the library
- **`google/uuid`** — RFC 4122 IDs (legacy used dead `satori/go.uuid`; do not inherit)
- **Plain numbered `.sql` files** (golang-migrate naming convention) for emitted DDL — runner-agnostic; user picks golang-migrate/goose/atlas
- **stdlib `testing` + `testify/require`** — assertions only, no `suite`, no `mock`

**Conflict to resolve in arch phase:** STACK.md recommends **`dave/jennifer`** as the primary code-emission engine (auto-managed imports across conditional emission paths); ARCHITECTURE.md recommends **`text/template` + `go/format`** (used by ent, sqlc, oapi-codegen, kubebuilder, protoc-gen-go; templates resemble their output; refactoring is cheap). Both researchers agree templates win for SQL/Markdown/Makefile non-Go output; the disagreement is about Go source emission specifically. See "Gaps to Address" below.

See [STACK.md](./STACK.md) for full rationale, version pins, and alternatives.

### Expected Features

PROJECT.md locks "full SCIM v2 coverage in v1," removing the usual subset-MVP lever. Sequencing within v1 is the only flexibility. Features research divides cleanly into RFC-mandated table stakes (server is non-compliant without them, fails IdP interop), code-gen-specific differentiators (the answer to "why generate instead of interpret a schema?"), and explicit anti-features (per PROJECT.md "Out of Scope").

**Must have (RFC 7643/7644 table stakes — all P1, all v1):**
- CRUD on `/Users` and `/Groups` (POST/GET/PUT/DELETE) plus per-resource and `/.search` POST endpoints
- Discovery: `/Schemas`, `/ResourceTypes`, `/ServiceProviderConfig` — generated statically from the definition
- **Full RFC 7644 §3.4.2.2 filter grammar** (eq/ne/co/sw/ew/gt/lt/ge/le/pr, and/or/not, parens, value-path expressions with bracketed predicates, URN-prefixed extension paths) — subset rejected because IdPs send arbitrary filters and partial parsers fail real traffic
- Sort, 1-based pagination, `attributes`/`excludedAttributes` projection
- **Full RFC 7644 §3.5.2 PATCH** including no-path, attribute-path, and value-path-filter forms — atomicity via transactions, mutability enforcement on every write path
- Schema-driven attribute behavior: `mutability` (4 values), `returned` (4 values), `uniqueness` (3 values), `required`, `caseExact`, multi-valued complex `primary` invariant (at most one element per attribute may be primary)
- ETag (weak `W/"..."`) on every response; `If-Match` enforcement on PUT/PATCH/DELETE returning 412; `If-None-Match: *` on POST
- `application/scim+json` content type; RFC 7644 §3.12 error envelope with full set of `scimType` values; correct status codes (200/201/204/304/400/401/403/404/409/412/500/501)
- Common attributes (`id`, `externalId`, `schemas`, `meta.{resourceType,created,lastModified,location,version}`)
- EnterpriseUser extension (functionally table stakes despite RFC labeling it an extension)
- `/Bulk` with `bulkId` cross-references and `failOnErrors`

**Should have (code-gen differentiators — all P1, all v1):**
- **Custom resource types** declared via the definition — User/Group/EnterpriseUser are just pre-bundled definitions; users get the full SCIM treatment for arbitrary resources
- **Generated per-resource SQL repositories** with hand-shaped SQL — eliminates the legacy tree-to-relational impedance
- **DDL / migration emission** — single-file `schema.sql` per resource set, deterministic
- **Generated SCIM v2 compliance test suite** emitted into the user's repo
- **Pluggable observability interfaces** in the runtime library (Logger/Tracer/Meter SPIs only, no concrete impl)
- **Plain `http.Handler` output** — no mux baked in; consumer composes
- **Typed `Config` struct** + CLI-emitted `main.go` (12-factor wiring on top of the same library API)
- **Tenant-id-aware persistence** from day one (single tenant value in v1; multi-tenancy at request layer is a future additive change)

**Defer (explicit v2+ per PROJECT.md):**
- IdP compatibility profiles (Microsoft/Okta/Azure AD quirks selectable per profile)
- Authentication/authorization in the generated server (gateway terminates upstream)
- Multi-tenancy at the request layer
- Postgres + MySQL drivers (interface designed for them; implementation post-v1)
- YAML/JSON definition format alternatives (only if user research shows the Go builder produces friction)
- Schema-evolution diffing between generator runs
- Concrete observability adapters (separate `go-scim-otel` / `go-scim-prom` modules)
- `/Me` alias (auth-blocked; document the gap)
- Group-membership eventing (legacy `groupsync` — not a SCIM requirement)

See [FEATURES.md](./FEATURES.md) for the full prioritization matrix and dependency graph.

### Architecture Approach

A **three-stage pipeline** with a thin separately-versioned runtime support library — the dominant pattern across mature Go codegen toolkits. Definition (fluent Go builder) → Loader/Validator (compile-time invariant checks with file:line attribution back to the builder call site) → IR (`Catalog` of typed structs, JSON-dumpable for debugging via `--dump-ir`) → multiple emitters running concurrently via `errgroup` (domain, persistence, server, spec, migrations, compliance) → `gofmt`/`goimports` post-processing → files on disk. Generated code imports only the runtime support module (`scimrt`); it never depends on the generator (`scimgen`). The runtime module ships the SCIM filter parser, PATCH engine, error envelope, ETag helpers, content-negotiation helpers, and the SQL driver SPI.

**Major components:**
1. **`scimdef` (in `scimgen` module)** — fluent Go builder DSL the user imports to author definitions; user code never reaches the generated server's runtime image
2. **`ir.Catalog`** — flat normalized typed structs; single source of truth for all emitters; never reach back into definition objects
3. **Validator** — pure functions over the IR; SCIM invariants (URN format, primary cardinality, sub-attribute name collisions, reserved attributes) become generator-time errors
4. **Multi-emitter pipeline** — one emitter per concern (domain types, persistence repos, HTTP server + handlers, spec/discovery payloads, migration DDL, compliance tests); composable via a `Generator` interface (cf. ent's `gen.Generator`); runs in parallel
5. **Code-emission engine** — recommendation conflict between `text/template`+`go/format` and `dave/jennifer` (see Gaps); both agree on `text/template` for non-Go output (SQL/Markdown/Makefile)
6. **`scimrt/sqldriver`** — small SPI (Tx, Exec, Query, Migrate, Close) in the runtime module; SQLite reference impl in `scimrt/sqldriver/sqlite`; dialect awareness lives in the **emitter** (not the runtime), so generated SQL is concrete and debuggable
7. **`scimrt/filter` + `scimrt/path` + `scimrt/patch`** — SCIM-spec parsers and PATCH engine; PATCH operates on typed resources via generated visitor methods (`(*User).ApplyPatch(op)`) rather than reflection
8. **`scimrt/observe`** — Logger/Tracer/Meter interfaces only; no-op defaults; user wires zerolog/OTel/Prometheus in their `main.go`
9. **Generated module layout** — user owns `cmd/server/main.go`; `internal/{server,domain,persistence,spec,migrations,scimtest}` are regenerated; `*_gen.go` naming distinguishes generated from user-owned files

See [ARCHITECTURE.md](./ARCHITECTURE.md) for the full system diagram, component table, and the eight-phase build-order recommendation.

### Critical Pitfalls

1. **Reintroducing a generic in-memory tree "for flexibility"** — the exact failure mode that drove the rewrite. Code review checklist must reject any `Property` interface, runtime `Resource.Get(path)`, or subscriber/visitor pattern for SCIM rule enforcement. Encode the rule in contributor docs at Phase 0; enforce by what the emitter produces.
2. **Reintroducing runtime schema interpretation** — generated server reads JSON schemas at startup, registers in a global mutable `schemaRegistry`, dispatches PATCH/filter through a runtime path resolver. Avoid via `//go:embed schemas/*.json` for discovery payloads only; PATCH/filter machinery asks the generated typed validator (`(*User).validatePatchTarget`), never a runtime schema record.
3. **Conflating spec compliance with IdP compatibility (v1 scope creep)** — Microsoft Entra sends `"op": "Add"` (capitalized) and stringified booleans; each accommodation feels small but the legacy died from this. v1 stance: strict spec, return 400 with `scimType: invalidSyntax` on non-compliant input; document what each IdP requires. Compatibility profiles are a v2 design that lands no implementation in v1. PR template asks "does this add IdP-specific tolerance?" — answer must be no.
4. **PATCH path-expression edge cases (RFC 7644 §3.5.2)** — filter-targeted Remove on multi-valued (`{"op":"remove","path":"members[value eq \"abc\"]"}`), Add on non-existent target, Replace-treated-as-Add when missing, multi-valued add via `value: [...]` array form, removing required sub-attributes. Avoid via a dedicated PATCH conformance corpus drawing from the spec, the `scim-patch` test corpus, and adversarial cases. PATCH logic lives once in the runtime library, parameterized by per-resource validation hooks — not regenerated per resource.
5. **Filter parser correctness (precedence, valuePath nesting, FILTER vs PATH grammar conflation)** — RFC operator precedence is `not > and > or`; ad-hoc parsers get `a eq 1 or b eq 2 and c eq 3` wrong; valuePath `emails[type eq "work" and (value ew "@x.com" or value ew "@y.com")]` requires correct nesting; reusing the FILTER parser as the PATH parser with a "mode" boolean is a tempting bug. Avoid via a generated parser (participle/PEG/ANTLR) from a committed grammar, property-based testing for parse-print round-trip, differential testing against another implementation, and Go's built-in fuzzer running 24h+ in CI before declaring filter parsing done. Distinct types for `FilterExpr` vs `PathExpr`.
6. **Pluggable SQL abstraction designed ahead of one working driver** — six weeks designing a 30-method interface that has never been exercised against a real DB. Build SQLite end-to-end as a concrete (non-interface) implementation passing the compliance suite; only then extract the interface from the working code. Postgres validates the SPI later; if it doesn't fit, that's a flag the SPI was wrong.
7. **Mutability/returned/uniqueness rule violations** — `password` (`returned: never`) leaking via a debug code path; `uniqueness` enforcement via raw DB constraint errors instead of pre-flight 409 with `scimType: uniqueness`; `returned: always` attributes (e.g. `id`) being filtered out by `?attributes=userName`. Avoid via per-attribute generated filter functions (`applyMutabilityForPATCH`, `filterForReturned`), a mechanical (POST × PUT × PATCH) × (server-set × client-set × omitted) test matrix, and a raw-JSON-substring test for forbidden attribute names.

See [PITFALLS.md](./PITFALLS.md) for the full pitfall-to-phase mapping, recovery strategies, "looks done but isn't" checklist, and explicit coverage of the three documented rewrite drivers.

## Implications for Roadmap

All four researchers independently propose roughly **8–11 phases** with strikingly similar dependency ordering. The convergence is high enough to treat the phase structure below as confident; what varies between proposals is grouping granularity, not order.

### Phase 0: Architecture Spike & Repo Setup
**Rationale:** Three pitfalls (#1 generic tree relapse, #2 runtime schema interpretation, #3 IdP scope creep) are catastrophic and must be encoded as written architectural rules and PR-template gates **before any code is written**. Multi-module workspace setup (Pitfall #18) and the `text/template` vs `dave/jennifer` decision (see Gaps) belong here too.
**Delivers:** Contributor architecture rules; PR template; `go.work` + two empty modules (`scimgen`, `scimrt`); CI builds with both `GOWORK=on` and `GOWORK=off`; emission-engine decision recorded.
**Avoids:** Pitfalls #1, #2, #3, #18.

### Phase 1: Definition + IR + Validator (no emission yet)
**Rationale:** Every later phase consumes the IR; get its shape right before any emitter depends on it. Validator running early (Pitfall #12) means user typos error at `go generate` with file:line attribution back to the builder call, not in machine-generated code 3000 lines deep. Internal model must be typed structs (Pitfall #21), not `map[string]interface{}`.
**Delivers:** `scimdef` fluent builder for User; `ir.Catalog` typed structs; validator covering URN format, primary uniqueness, attribute name collisions, sub-attribute restrictions; `dump-ir` command (cf. sqlc `dumpcatalog`) for debugging.
**Uses:** Stack: stdlib only; testify/require; google/go-cmp for golden tests.
**Avoids:** Pitfalls #12, #21.

### Phase 2: Domain Emitter (one resource type, no SQL, no HTTP)
**Rationale:** Smallest possible end-to-end emit slice. Validates the template/format/snapshot loop that every later emitter inherits. Generated struct is pure data — no runtime library needed yet. Establishes the "generated code always passes through `go/format.Source`" rule (Pitfall #9) and the `// Code generated ... DO NOT EDIT.` marker convention.
**Delivers:** Single emitter producing `domain/user.go` with typed struct + `Validate()` method; golden-file snapshot tests; `gofmt` + `goimports` pipeline.
**Implements:** Domain emitter component.
**Avoids:** Pitfall #9.

### Phase 3: Runtime Library Skeleton + Persistence Emitter (SQLite, concrete-first)
**Rationale:** Persistence is the longest pole and the architectural commitment that defines the runtime/generated split. Per Pitfall #4, build SQLite as a **concrete non-interface** implementation first; extract the SPI later from the working code. Establishes the per-resource generated repo pattern that eliminates the legacy tree-to-SQL impedance. ETag versioning (Pitfall #8) lands here as a per-resource monotonic counter incremented in the same transaction as the write.
**Delivers:** `scimrt/sqldriver/sqlite` reference impl; `migrations` emitter producing deterministic single-file `schema.sql` (Pitfall #19); `persistence` emitter producing `UserRepo` with Create/Get/Replace/Delete; ETag/version column emission. Postgres deferred.
**Uses:** `modernc.org/sqlite` v1.46.1+, `database/sql`, `google/uuid`.
**Implements:** Persistence emitter, runtime SQL driver, ETag mechanics.
**Avoids:** Pitfalls #4, #8, #19.

### Phase 4: HTTP Server Emitter + Discovery Endpoints + List/Sort/Pagination/Projection
**Rationale:** Needs Phase 3's persistence. Server emission, discovery payloads (`/Schemas`, `/ResourceTypes`, `/ServiceProviderConfig` — all generated from the same IR as the types, embedded via `//go:embed`, addressing Pitfalls #2, #16, #20), and the filter parser all share IR work and ship together. Filter parser is upstream of three big features (list, PATCH path expressions, bulk) so it must be solid here. List semantics (Pitfall #13: 1-based startIndex, integer not string `totalResults`, ListResponse wrapper) and content-type/error-format (Pitfall #14: `application/scim+json`, string `status`, single error helper) lock in.
**Delivers:** Generated `internal/server` (router + per-resource handlers); `scimrt/filter` parser + AST + SQLite translator; discovery endpoint emission; list endpoint with full filter/sort/pagination/projection; `scimrt/scimerr` error envelope; `scimrt/scimjson` content negotiation.
**Implements:** Server emitter, discovery emitter, filter compiler.
**Avoids:** Pitfalls #2, #13, #14, #16, #20.

### Phase 5: PATCH + ETag/If-Match + Bulk
**Rationale:** Universally identified as the highest-complexity phase. Schedule slack. PATCH semantics (Pitfall #5) are gnarly enough that doing them after the simpler CRUD cycle is well-debugged is the right order. PATCH engine lives in the runtime library, parameterized by generated per-resource visitor methods. ETag enforcement on conditional writes (Pitfall #8) and Bulk (which fans out to existing handlers and uses the filter parser for queried operations) ship together because they share the write-path concurrency surface.
**Delivers:** `scimrt/patch` engine with full RFC 7644 §3.5.2 coverage including value-path filters; PATCH conformance corpus drawn from spec examples + `scim-patch` corpus + adversarial cases; `If-Match` middleware returning 412 on stale ETag; `If-None-Match: *` on POST; `/Bulk` endpoint emission with `bulkId` cross-references and `failOnErrors`; mutability/returned enforcement on every PATCH path (Pitfall #7).
**Implements:** PATCH engine, ETag enforcement, Bulk emitter.
**Avoids:** Pitfalls #5, #7, #8, #15.

### Phase 6: Multi-Resource Support — Group + EnterpriseUser + Custom Types
**Rationale:** At this point the emitter is mature on User; broadening to N resources mostly exercises code paths already proven. Group introduces the `User.groups` read-only-mirror invariant (Pitfall #15: canonical store is `Group.members`, `User.groups` is a read-only projection). EnterpriseUser exercises the schema extension mechanism. Custom resource types prove the differentiator and stress-test the emitter on cross-resource references.
**Delivers:** Built-in Group and EnterpriseUser definitions; extension schema mechanism; custom resource type support; multi-resource integration tests.
**Implements:** Schema extension support, multi-resource emission.
**Avoids:** Pitfall #15.

### Phase 7: CLI + Observability SPI + Compliance Test Emitter
**Rationale:** Productisation, not architecture. Easier once the inner loop is stable. CLI (Cobra) is a thin wrapper over the library API. Observability (Pitfall #5 from Architecture: bare interfaces, no-op defaults). Compliance suite emission packages the test corpora developed in Phases 4–5 into a generated `scimtest/` package the user runs against their compiled binary. `scimgen check` subcommand (Pitfall #11) warns on definition diffs that would break user code.
**Delivers:** `cmd/scimgen` CLI with `generate`/`check`/`dump-ir` subcommands; CLI-emitted `cmd/server/main.go` with env/flag wiring; `scimrt/observe` interfaces + no-op defaults; `scimtest` compliance suite emitter; regen workflow that preserves user-owned files (Pitfall #10).
**Uses:** `spf13/cobra`.
**Implements:** CLI, observability SPI, compliance emitter.
**Avoids:** Pitfalls #10, #11.

### Phase 8: Postgres Dialect + Driver (Architectural Validation Gate)
**Rationale:** Validates that the SQL dialect SPI extracted in Phase 3 actually accommodates a second backend. If the SPI requires changes to fit Postgres, that's a signal the abstraction was wrong — and it's better to find out before more emitters depend on it. Per PROJECT.md "Out of Scope for v1" — but the design milestone of "interface stress-tested against Postgres" should land in v1 even if the implementation ships post-v1. Could alternatively be the first v1.x phase.
**Delivers:** `scimgen/sqldialect/postgres.go`; `scimrt/sqldriver/postgres` impl using `pgx` underneath the `database/sql` shape; SPI refactor if needed; Postgres-specific compliance run.
**Note:** Roadmapper should decide whether this lands in v1 or v1.x. All four researchers flag it as the natural validation gate; PROJECT.md defers it.

### Phase Ordering Rationale

- **Architecture rules before code (Phase 0):** Three of the most catastrophic pitfalls are reversibility traps; they must be encoded as norms before they can be violated.
- **IR before emitters (Phase 1):** Every emitter consumes IR; getting it wrong means rewriting all emitters.
- **One emitter end-to-end before broadening (Phases 2–3):** Establish the template/format/snapshot loop and the runtime/generated split on the simplest case (User domain → User repo) before any of it is locked in by additional emitters.
- **Filter before PATCH (Phase 4 before Phase 5):** PATCH value-path expressions reuse the filter parser; do not build two parsers.
- **PATCH gets its own phase (Phase 5):** All four researchers independently flag PATCH as the highest-complexity single piece of work; the spec is genuinely ambiguous in places, vendor payloads diverge in the wild, and the "looks done but isn't" surface is enormous.
- **Multi-resource after single-resource (Phase 6):** Stress-test the emitter on N=1 first; broadening to N=many is mostly proving existing patterns generalize.
- **CLI/observability/compliance late (Phase 7):** Productisation polish on a stable inner loop.
- **Postgres as the SPI gate (Phase 8):** The architecture's portability claim is worthless until proven against a second backend.

### Research Flags

**Phases likely needing deeper research during planning (`/gsd:research-phase` recommended):**
- **Phase 0:** the `text/template` vs `dave/jennifer` decision is the single most material unresolved question; both researchers make defensible cases. Worth a focused research-phase to look at concrete output shapes for the gnarliest emitter (filter compiler or PATCH visitor methods) and decide on evidence.
- **Phase 5 (PATCH):** spec ambiguity in §3.5.2.1 ("if the target location does not exist…"); vendor payload divergence; need to assemble the test corpus from `scim-patch`, real captured payloads, and adversarial cases. Visitor-method-vs-reflection decision for the PATCH applier shape (flagged in STACK.md Open Question #1) lands here.
- **Phase 4 (filter parser):** parser-generator choice (participle vs pigeon PEG vs ANTLR-Go vs hand-rolled with property tests) deserves a focused look; the legacy filter.go was 1100 lines and that is a warning shape.
- **Phase 8 (Postgres):** dialect differences (RETURNING, JSON columns, citext for caseExact, weakly-defined LIMIT/OFFSET semantics under concurrent writes) need a paper exercise even if the implementation defers.

**Phases with standard patterns (research-phase likely unnecessary):**
- **Phase 1 (IR + validator):** standard codegen pattern; reference ent/sqlc directly.
- **Phase 2 (domain emitter):** template-and-format loop is well-trodden.
- **Phase 6 (multi-resource):** at this point the emitter is mature; new resources exercise existing paths.
- **Phase 7 (CLI):** Cobra is the de-facto pattern.

## Confidence Assessment

| Area | Confidence | Notes |
|------|------------|-------|
| Stack | HIGH | Core choices (Go 1.24, modernc.org/sqlite, net/http, log/slog, database/sql, Cobra, testify) verified against go.dev release notes, official package docs, 2026-dated benchmarks (cvilsmeier/go-sqlite-bench), and broad Go ecosystem signal. MEDIUM only on the definition-format hypothesis (Go builder vs YAML — locked as Pending in PROJECT.md) and the compliance-suite recommendation (no Go-native SCIM compliance suite in active maintenance; recommendation is informed inference). |
| Features | HIGH | All P1 features verified directly against RFC 7643 and RFC 7644. Anti-features explicitly enumerated in PROJECT.md "Out of Scope". MEDIUM only on bulk-operations complexity (real-world cross-IdP `bulkId` variation may surface edge cases not visible from the RFC alone) and the `/Me` skip recommendation (judgment call; alternative is wiring a Config hook). |
| Architecture | HIGH | Three-stage pipeline + two-module split is the dominant pattern across ent/sqlc/oapi-codegen/kubebuilder/protoc-gen-connect-go; well-documented and battle-tested. SCIM-specific specifics (PATCH applier shape, filter compiler shape) flagged for arch-phase deep-dive. |
| Pitfalls | HIGH | SCIM/RFC claims verified against RFCs 7643/7644/9865 and IdP vendor docs (Microsoft Learn, Okta). Go tooling verified against go.dev, pkg.go.dev, and 2026-dated benchmarks. MEDIUM only on the "common mistakes" attributed to vendor implementations (multiple corroborating sources but not exhaustive). |

**Overall confidence:** HIGH

### Gaps to Address

1. **`text/template` vs `dave/jennifer` for Go source emission** — STACK.md and ARCHITECTURE.md disagree explicitly. STACK argues Jennifer's auto-managed imports across conditional emission paths are decisive for handlers/repos/validators/filter compilers/PATCH appliers. Architecture argues templates win because most output is mostly-static, templates resemble their output, refactoring is cheap, and every other major Go codegen toolkit (ent, sqlc, oapi-codegen, kubebuilder, protoc-gen-go) chose templates. Both agree on `text/template` for non-Go output (SQL/Markdown/Makefile). **Action:** make this a Phase 0 decision; consider a small spike emitting one gnarly artifact (e.g., the per-resource filter compiler with conditional CASE expressions on attribute types) in both styles and choosing on evidence. The choice is reversible at significant cost; making it once and committing is better than churning.

2. **Definition format: fluent Go builder vs YAML/JSON** — locked as Pending in PROJECT.md; STACK.md and FEATURES.md both adopt the Go builder as the working hypothesis but flag YAML as a not-ruled-out alternative. The architecture (Architecture Pattern 1 — definition is "just Go code that calls Builder methods") accommodates either; the YAML path is "parser layer feeds the builder calls." **Action:** ship Go builder in v1 per hypothesis; revisit only if user research shows friction.

3. **PATCH applier shape: generated visitor methods vs reflection** — STACK.md Open Question #1 and Architecture's PATCH flow both flag this. Visitor methods avoid reflection entirely but add per-resource emit complexity; reflection is simpler in the runtime but slower and harder to verify. **Action:** Phase 5 research; default lean toward generated visitor methods (consistent with the "no reflection on definition objects at runtime" rule from Architecture Anti-Pattern #4).

4. **Filter compiler shape: per-resource compiled evaluator vs generic SCIM-filter-to-SQL translator in runtime** — STACK.md Open Question #2. Either works; tradeoff is generated-code size vs runtime-library complexity. Both researchers lean toward "filter parser in `scimrt/filter` produces an AST; per-dialect SQL translator in `scimrt/filter/sqlite` consumes it; per-resource generated code passes the AST to the translator." **Action:** Phase 4 confirms the lean.

5. **Down-migration emission feasibility** — STACK.md Open Question #3; Pitfalls #19 also touches this. Generator emits fresh DDL each run per PROJECT.md "Out of Scope for schema-evolution diffing." Whether to emit `down.sql` at all in v1 is uncertain. **Action:** Phase 3 decides; lean toward single `schema.sql` per Pitfall #19 (numbered migration files conflict with regen workflow).

6. **External SCIM compliance suite integration in CI** — STACK.md Open Question #4. `python-scim/scim2-tester` is the closest active external suite; integrating adds a Python toolchain to generator development. **Action:** Phase 7 decides; defer to post-v1 unless a CI cross-validation gate becomes important earlier.

## Sources

### Primary (HIGH confidence)
- RFC 7643 — System for Cross-domain Identity Management: Core Schema (https://datatracker.ietf.org/doc/html/rfc7643) — verified directly for User/Group/EnterpriseUser, attribute characteristics, multi-valued complex `primary` invariant, common attributes
- RFC 7644 — System for Cross-domain Identity Management: Protocol (https://datatracker.ietf.org/doc/html/rfc7644) — verified directly for endpoint set, PATCH §3.5.2, filter grammar §3.4.2.2, sort/pagination/projection §3.4.2, ETag §3.14, error envelope §3.12, content type §3.1, bulk §3.7
- RFC 9865 — Cursor-Based Pagination of SCIM Resources (October 2025)
- RFC 7232 — HTTP Conditional Requests
- Go 1.24 Release Notes (https://go.dev/doc/go1.24)
- ent/entgo.io — canonical reference for Go schema-driven codegen
- sqlc.dev docs — SQL → catalog IR → codegen plugin pipeline; `dumpcatalog` debug command
- Buf Connect (connectrpc.com) — IDL-driven codegen with thin runtime support library
- oapi-codegen GitHub — OpenAPI → typed schema model → ParseTemplates emitter
- protoc-gen-connect-go — runtime/generated split pattern
- modernc.org/sqlite docs — pure-Go transpilation, no CGo
- text/template Go docs (https://pkg.go.dev/text/template)
- dave/jennifer GitHub
- Microsoft Learn — Known issues with SCIM 2.0 protocol compliance (Entra ID)
- thomaspoignant/scim-patch — JS PATCH library with extensive edge-case test corpus

### Secondary (MEDIUM confidence — multiple corroborating sources)
- Encore — Comparing the best Go ORMs (2026)
- Rost Glukhov — Which ORM to use in Go (March 2025)
- cvilsmeier/go-sqlite-bench (March 2026 driver benchmarks)
- DataStation — SQLite in Go, with and without CGo
- Atlas — Picking a database migration tool for Go
- Alex Edwards — Which Go Router Should I Use?
- Calhoun.io — Go's 1.22+ ServeMux vs Chi Router
- JetBrains — The Go Ecosystem in 2025 (testify adoption stats)
- scim2/filter-parser DeepWiki analysis
- Microsoft Q&A — Entra SCIM Provisioning sends invalid PATCH requests

### Tertiary (LOW confidence individually, HIGH when corroborated)
- python-scim/scim2-tester — closest active SCIM compliance suite
- WSO2 SCIM2 Compliance Test Suite — reference catalog
- elimity-com/scim, arturoeanton/goscim — existing Go SCIM library architectural references

### Internal references
- `.planning/PROJECT.md` — locked scope, locked exclusions, three rewrite drivers, key decisions
- `.planning/codebase/ARCHITECTURE.md` — legacy generic-tree architecture (anti-pattern reference)
- `.planning/codebase/CONCERNS.md` — three documented rewrite drivers (feasibility, efficiency, portability)
- `.planning/codebase/STRUCTURE.md` — legacy module organization

---
*Research completed: 2026-05-07*
*Synthesized from: STACK.md, FEATURES.md, ARCHITECTURE.md, PITFALLS.md*
*Ready for roadmap: yes*

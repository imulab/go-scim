# Requirements: go-scim v3 (next)

**Defined:** 2026-05-07
**Core Value:** A developer can describe their SCIM resource set once and get a spec-compliant SCIM v2 server with SQL persistence, mountable on any `net/http`-compatible mux, with no runtime schema-interpretation overhead.

## v1 Requirements

### Definition & IR

- [ ] **DEF-01**: User can describe a SCIM resource set via a fluent Go builder (working hypothesis; YAML/JSON deferred)
- [ ] **DEF-02**: Builder constructs a typed internal model the parser ingests (no `map[string]interface{}` in the IR)
- [ ] **DEF-03**: Validator runs against the IR before any emission and surfaces invariant violations with file:line attribution back to the builder call site
- [ ] **DEF-04**: Validator enforces SCIM invariants — URN format, attribute name collisions, sub-attribute name restrictions, primary uniqueness on multi-valued complex attributes, reserved attribute names
- [ ] **DEF-05**: Generator exposes a `dump-ir` command that emits the IR as JSON for debugging

### Code Generation Engine

- [ ] **GEN-01**: Generator implements a three-stage pipeline (Definition → IR → Multi-Emitter → gofmt → files)
- [ ] **GEN-02**: Generated Go source always passes through `go/format.Source` (and `goimports` where applicable) before being written
- [ ] **GEN-03**: All generated files carry the `// Code generated ... DO NOT EDIT.` marker on the first non-blank line
- [ ] **GEN-04**: Generated files use a `*_gen.go` naming convention to distinguish them from user-owned files
- [ ] **GEN-05**: Multiple emitters run concurrently (e.g. via `errgroup`); failure of any emitter aborts the run with a clear error
- [ ] **GEN-06**: Regeneration is deterministic — identical IR produces byte-identical output (no timestamps, map iteration nondeterminism, etc.)

### Domain Emitter

- [ ] **DOM-01**: Per-resource typed Go struct emitted (no generic `Resource` map)
- [ ] **DOM-02**: Per-resource `Validate()` method emitted, enforcing required/caseExact/canonicalValues/multi-valued primary cardinality at write time
- [ ] **DOM-03**: Generated struct field tags carry attribute metadata sufficient for JSON serialization per RFC 7643

### Persistence (SQL)

- [ ] **PER-01**: Per-resource SQL repository emitted with hand-shaped SQL — no embedded ORM, no runtime tree
- [ ] **PER-02**: Repository exposes Create / Get / Replace / Delete operations that participate in a single transaction with the version-counter increment
- [ ] **PER-03**: Persistence layer is tenant-id aware in its column shape (single tenant value used in v1; design supports later request-layer multi-tenancy)
- [ ] **PER-04**: SQL driver abstraction exists in `rt/sqldriver` exposing a small SPI (Tx, Exec, Query, Migrate, Close)
- [ ] **PER-05**: `modernc.org/sqlite` reference driver implements the SPI and ships in v1
- [ ] **PER-06**: SQL driver SPI is extracted only after the SQLite implementation passes the compliance suite (concrete-first)

### Migration / DDL Emission

- [ ] **DDL-01**: Generator emits a deterministic single-file `schema.sql` per resource set (or numbered migration files behind a flag — decided in Phase 3)
- [ ] **DDL-02**: Emitted DDL is consumable by golang-migrate / goose / atlas conventions
- [ ] **DDL-03**: User can opt to use the emitted DDL as-is or supply their own migration scripts; runtime correctness is the user's responsibility

### HTTP Server

- [ ] **HTTP-01**: Generated handler is a plain `net/http` `http.Handler`; no third-party router baked into generated code
- [ ] **HTTP-02**: Generated routing covers all RFC 7644 endpoints (per-resource CRUD, `/.search`, `/Bulk`, discovery endpoints)
- [ ] **HTTP-03**: Generated server enforces `application/scim+json` content type on requests with bodies and returns it on all responses
- [ ] **HTTP-04**: Generated error responses use the RFC 7644 §3.12 envelope with the correct `schemas` URN, `status` as a string, and the appropriate `scimType` value
- [ ] **HTTP-05**: Generated handlers return correct HTTP status codes per RFC 7644 (200/201/204/304/400/401/403/404/409/412/500/501)
- [ ] **HTTP-06**: Common attributes (`id`, `externalId`, `schemas`, `meta.{resourceType,created,lastModified,location,version}`) populated correctly on every response

### Discovery Endpoints

- [ ] **DISC-01**: `/Schemas` endpoint generated from the same IR as the typed structs (single source of truth)
- [ ] **DISC-02**: `/ResourceTypes` endpoint generated and consistent with what the server actually serves
- [ ] **DISC-03**: `/ServiceProviderConfig` endpoint generated and accurately reflects supported features (PATCH=true, bulk=true, filter=true, etc.)
- [ ] **DISC-04**: Discovery payloads embedded via `//go:embed` (no runtime schema interpretation)

### List / Filter / Sort / Pagination / Projection

- [ ] **LIST-01**: List endpoint supports 1-based `startIndex`, `count`, and reports `totalResults` as an integer reflecting the filter
- [ ] **LIST-02**: Response uses the `ListResponse` wrapper with the correct `schemas` URN
- [ ] **LIST-03**: Filter parser implements the full RFC 7644 §3.4.2.2 grammar — `eq`/`ne`/`co`/`sw`/`ew`/`gt`/`lt`/`ge`/`le`/`pr`, `and`/`or`/`not`, parens, value-path expressions with bracketed predicates, URN-prefixed extension paths
- [ ] **LIST-04**: Filter parser produces a typed `FilterExpr` AST distinct from the `PathExpr` AST used by PATCH
- [ ] **LIST-05**: Filter AST → SQL translator emits parameterized SQL safe against injection
- [ ] **LIST-06**: Sort supports `sortBy` and `sortOrder` per RFC 7644 §3.4.2.3
- [ ] **LIST-07**: Attribute projection respects `attributes` and `excludedAttributes`; `returned: always` attributes are never filtered out; `returned: never` attributes are never returned
- [ ] **LIST-08**: `POST /.search` supported per RFC 7644 §3.4.3 (per-resource and root-level)

### PATCH

- [ ] **PATCH-01**: Full RFC 7644 §3.5.2 PATCH operation set: `add`, `replace`, `remove`
- [ ] **PATCH-02**: Path forms supported: no path, attribute path, sub-attribute path, value-path filter (e.g. `emails[type eq "work"].value`)
- [ ] **PATCH-03**: Multi-valued add via `value: [...]` array form supported
- [ ] **PATCH-04**: Filter-targeted Remove on multi-valued attributes works correctly
- [ ] **PATCH-05**: Replace on a non-existent target behaves per RFC (treated as Add where applicable)
- [ ] **PATCH-06**: Removing a required sub-attribute returns the correct error; removing a `returned: never` attribute is a no-op or rejected per spec
- [ ] **PATCH-07**: PATCH is atomic — entire op set succeeds or the resource is unchanged
- [ ] **PATCH-08**: PATCH engine lives in `rt/patch` and operates on typed resources via generated visitor methods (e.g. `(*User).ApplyPatch(op)`); no reflection over generated structs at runtime
- [ ] **PATCH-09**: PATCH conformance test corpus drawn from spec examples + `scim-patch` corpus + adversarial cases ships with the runtime

### ETag / Concurrency

- [ ] **ETAG-01**: Every response carries a weak ETag (`W/"..."`) sourced from the resource's `meta.version`
- [ ] **ETAG-02**: `meta.version` is a per-resource monotonic counter incremented in the same transaction as the write
- [ ] **ETAG-03**: `If-Match` enforced on PUT, PATCH, and DELETE; mismatch returns 412
- [ ] **ETAG-04**: `If-None-Match: *` enforced on POST for upsert protection where applicable

### Bulk

- [ ] **BULK-01**: `/Bulk` endpoint emitted per RFC 7644 §3.7
- [ ] **BULK-02**: `bulkId` cross-references resolved correctly within a single bulk request
- [ ] **BULK-03**: `failOnErrors` honored
- [ ] **BULK-04**: Bulk operations fan out to the existing per-resource handlers (no duplicated logic)

### Resource Types

- [ ] **RES-01**: Built-in `User` resource type definition shipped
- [ ] **RES-02**: Built-in `Group` resource type definition shipped, with `User.groups` correctly modeled as a read-only projection of `Group.members`
- [ ] **RES-03**: Built-in `EnterpriseUser` extension shipped (functionally table stakes despite RFC labeling it an extension)
- [ ] **RES-04**: User-defined custom resource types supported via the same definition mechanism (the code-gen value-prop)

### CLI

- [ ] **CLI-01**: `scim` CLI built on Cobra wraps the library API (CLI = thin wrapper)
- [ ] **CLI-02**: `scim generate` runs the full pipeline against a definition entry point and writes a runnable repo
- [ ] **CLI-03**: `scim check` warns on definition diffs that would break user-owned code
- [ ] **CLI-04**: `scim dump-ir` emits the IR as JSON for debugging
- [ ] **CLI-05**: CLI emits `cmd/server/main.go` that wires environment variables and flags into the typed `Config` struct (12-factor on top of the library API)

### Library

- [ ] **LIB-01**: Generator is importable as a Go library; in-process generation supported
- [ ] **LIB-02**: Generated server consumes a typed `Config` struct constructed by the user's `main`
- [ ] **LIB-03**: Generator and runtime support library are separately versioned (semver)
- [ ] **LIB-04**: Generated code imports only the `rt` runtime module — never the `gen` generator module

### Observability

- [ ] **OBS-01**: `rt/observe` exposes Logger / Tracer / Meter SPIs only (no concrete impl)
- [ ] **OBS-02**: No-op default implementations supplied so generated servers compile and run with no observability dependencies
- [ ] **OBS-03**: User can wire concrete adapters (slog handler, OTel, Prometheus) via the `Config` struct

### Testing

- [ ] **TEST-01**: Generator emits unit tests for generated handlers and repos
- [ ] **TEST-02**: Generator emits integration tests that run against the SQLite reference driver
- [ ] **TEST-03**: Generator emits a SCIM v2 compliance test suite into the user's repo (`internal/scimtest/`) covering RFC 7643/7644 conformance
- [ ] **TEST-04**: Generator's own test suite covers golden-file snapshots of generated output
- [ ] **TEST-05**: Filter parser is fuzz-tested (`go test -fuzz`) for at least 24h in CI before being declared done
- [ ] **TEST-06**: Generator self-test includes a "compiled examples" CI module that compiles and runs the generated quickstart server

### Repo / Workspace / License

- [ ] **REPO-01**: Multi-module workspace via `go.work` (gitignored); CI runs with `GOWORK=off` to validate each module standalone
- [ ] **REPO-02**: Two `go.mod` modules: `gen` (generator at `github.com/imulab/go-scim/gen`) and `rt` (runtime support library at `github.com/imulab/go-scim/rt`)
- [ ] **REPO-03**: License is MIT (matching legacy `.legacy/LICENSE`)
- [ ] **REPO-04**: Contributor architecture rules document encodes anti-relapse rules (no generic tree, no runtime schema interpretation, no IdP accommodation)
- [ ] **REPO-05**: PR template asks reviewers to verify changes do not introduce IdP-specific tolerance

## v2 Requirements

Deferred to future release. Tracked but not in current roadmap.

### Persistence

- **PER-V2-01**: PostgreSQL driver implementation behind the same SPI
- **PER-V2-02**: MySQL/MariaDB driver implementation behind the same SPI
- **PER-V2-03**: Schema-evolution migrations between generator runs (diffing prior IR vs current)
- **PER-V2-04**: Cursor-based pagination per RFC 9865

### Compatibility

- **COMPAT-V2-01**: IdP compatibility profiles declared in the definition (e.g., Microsoft Entra capitalization aliasing, Okta extra-field tolerance)

### Auth

- **AUTH-V2-01**: Built-in OAuth2 Bearer / JWT validation in the generated server
- **AUTH-V2-02**: `/Me` endpoint emission once auth is in scope

### Multi-Tenancy

- **TENANT-V2-01**: Multi-tenant resolution at the request layer (header / path / token claim) on top of the tenant-id-aware persistence already shipped in v1

### Observability

- **OBS-V2-01**: Concrete adapter modules (`go-scim-otel`, `go-scim-prom`) outside the runtime support library

### Definition Format

- **DEF-V2-01**: YAML/JSON definition format alternative if user research shows the Go builder produces friction

### Eventing

- **EVT-V2-01**: Group-membership change eventing (only if a clear consumer demand emerges; not a SCIM spec requirement)

## Out of Scope

| Feature | Reason |
|---------|--------|
| IdP non-compliance accommodations in v1 | Strict spec keeps the model honest; quirks become a definition-level feature later (compatibility profiles) |
| Authentication / authorization built into the generated server (v1) | Assume an upstream gateway terminates auth; keeps the generator focused on the SCIM data plane |
| Multi-tenancy at the request layer (v1) | Single-tenant per generated server in v1; persistence layer is tenant-id aware so multi-tenancy is a later additive change |
| Group-membership change eventing (legacy `groupsync`) | Implementation detail of the old project, not in the SCIM spec |
| Backward compatibility with legacy `go-scim` consumers | Clean break, different module path; legacy is frozen for reference |
| Concrete observability implementations in the runtime | Interfaces only; user wires concrete deps so the runtime stays lean |
| Schema-evolution migrations between regen runs (v1) | Generator emits fresh DDL each run; user owns diffing/migration in v1 |
| Explicit performance/scale targets | Correctness over throughput in v1 |
| Concrete non-SQLite SQL drivers in v1 | Postgres/MySQL ship later behind the same interface |
| Concrete HTTP-mux integrations beyond `net/http` | Handler is plain `http.Handler`; user mounts on chi/gorilla/gin/etc. |
| Subset SCIM v2 coverage | Full coverage in v1 — partial coverage hides architectural cliffs |
| Generic in-memory tree / runtime schema interpretation | This is the legacy architecture being replaced; reintroducing it defeats the rewrite |
| Embedded ORM (sqlc/ent/sqlboiler/gorm) | Code-gen IS the ORM; embedding another schema-aware Go-emitter is two generators fighting |
| Reflection over generated structs at runtime | Visitor methods generated per resource; reflection defeats the typed-codegen value-prop |

## Traceability

Mapped during roadmap creation (2026-05-07). All v1 REQ-IDs map to exactly one phase.

| Requirement | Phase | Status |
|-------------|-------|--------|
| DEF-01 | Phase 1 | Pending |
| DEF-02 | Phase 1 | Pending |
| DEF-03 | Phase 1 | Pending |
| DEF-04 | Phase 1 | Pending |
| DEF-05 | Phase 1 | Pending |
| GEN-01 | Phase 2 | Pending |
| GEN-02 | Phase 2 | Pending |
| GEN-03 | Phase 2 | Pending |
| GEN-04 | Phase 2 | Pending |
| GEN-05 | Phase 2 | Pending |
| GEN-06 | Phase 2 | Pending |
| DOM-01 | Phase 2 | Pending |
| DOM-02 | Phase 2 | Pending |
| DOM-03 | Phase 2 | Pending |
| PER-01 | Phase 3 | Pending |
| PER-02 | Phase 3 | Pending |
| PER-03 | Phase 3 | Pending |
| PER-04 | Phase 3 | Pending |
| PER-05 | Phase 3 | Pending |
| PER-06 | Phase 3 | Pending |
| DDL-01 | Phase 3 | Pending |
| DDL-02 | Phase 3 | Pending |
| DDL-03 | Phase 3 | Pending |
| HTTP-01 | Phase 4 | Pending |
| HTTP-02 | Phase 4 | Pending |
| HTTP-03 | Phase 4 | Pending |
| HTTP-04 | Phase 4 | Pending |
| HTTP-05 | Phase 4 | Pending |
| HTTP-06 | Phase 4 | Pending |
| DISC-01 | Phase 4 | Pending |
| DISC-02 | Phase 4 | Pending |
| DISC-03 | Phase 4 | Pending |
| DISC-04 | Phase 4 | Pending |
| LIST-01 | Phase 4 | Pending |
| LIST-02 | Phase 4 | Pending |
| LIST-03 | Phase 4 | Pending |
| LIST-04 | Phase 4 | Pending |
| LIST-05 | Phase 4 | Pending |
| LIST-06 | Phase 4 | Pending |
| LIST-07 | Phase 4 | Pending |
| LIST-08 | Phase 4 | Pending |
| PATCH-01 | Phase 5 | Pending |
| PATCH-02 | Phase 5 | Pending |
| PATCH-03 | Phase 5 | Pending |
| PATCH-04 | Phase 5 | Pending |
| PATCH-05 | Phase 5 | Pending |
| PATCH-06 | Phase 5 | Pending |
| PATCH-07 | Phase 5 | Pending |
| PATCH-08 | Phase 5 | Pending |
| PATCH-09 | Phase 5 | Pending |
| ETAG-01 | Phase 5 | Pending |
| ETAG-02 | Phase 3 | Pending |
| ETAG-03 | Phase 5 | Pending |
| ETAG-04 | Phase 5 | Pending |
| BULK-01 | Phase 5 | Pending |
| BULK-02 | Phase 5 | Pending |
| BULK-03 | Phase 5 | Pending |
| BULK-04 | Phase 5 | Pending |
| RES-01 | Phase 6 | Pending |
| RES-02 | Phase 6 | Pending |
| RES-03 | Phase 6 | Pending |
| RES-04 | Phase 6 | Pending |
| CLI-01 | Phase 7 | Pending |
| CLI-02 | Phase 7 | Pending |
| CLI-03 | Phase 7 | Pending |
| CLI-04 | Phase 7 | Pending |
| CLI-05 | Phase 7 | Pending |
| LIB-01 | Phase 1 | Pending |
| LIB-02 | Phase 4 | Pending |
| LIB-03 | Phase 1 | Pending |
| LIB-04 | Phase 1 | Pending |
| OBS-01 | Phase 7 | Pending |
| OBS-02 | Phase 7 | Pending |
| OBS-03 | Phase 7 | Pending |
| TEST-01 | Phase 7 | Pending |
| TEST-02 | Phase 7 | Pending |
| TEST-03 | Phase 7 | Pending |
| TEST-04 | Phase 7 | Pending |
| TEST-05 | Phase 7 | Pending |
| TEST-06 | Phase 7 | Pending |
| REPO-01 | Phase 0 | Pending |
| REPO-02 | Phase 0 | Pending |
| REPO-03 | Phase 0 | Pending |
| REPO-04 | Phase 0 | Pending |
| REPO-05 | Phase 0 | Pending |

**Coverage:**
- v1 requirements: 85 total
- Mapped to phases: 85
- Unmapped: 0

**Per-phase counts:**

| Phase | REQ Count |
|-------|-----------|
| Phase 0: Architecture Spike & Repo Setup | 5 |
| Phase 1: Definition + IR + Validator | 8 |
| Phase 2: Domain Emitter + Generator Pipeline | 9 |
| Phase 3: Persistence + DDL + ETag (SQLite) | 10 |
| Phase 4: HTTP Server + Discovery + List | 19 |
| Phase 5: PATCH + ETag/If-Match + Bulk | 16 |
| Phase 6: Multi-Resource (Group/EU/Custom) | 4 |
| Phase 7: CLI + Observability + Compliance | 14 |
| **Total** | **85** |

---
*Requirements defined: 2026-05-07*
*Last updated: 2026-05-07 — traceability populated by roadmapper*
*Updated 2026-05-07 (Phase 0 plan) — CONTEXT.md overrides applied: module names scimgen→gen, scimrt→rt; CLI binary scimgen→scim. Requirement IDs unchanged.*

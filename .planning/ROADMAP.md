# Roadmap: go-scim v3 (next)

## Overview

go-scim v3 is a build-time code generator that takes a SCIM resource definition and emits a spec-compliant SCIM v2 server (typed Go structs, per-resource SQL repositories, SQLite DDL, HTTP handler, compliance tests). The journey moves from architectural rules and a multi-module repo skeleton through an end-to-end emission slice on a single resource (User), then broadens horizontally to the full RFC 7643/7644 surface (PATCH, Bulk, multi-resource, custom types) before productising with a CLI, observability SPIs, and an emitted compliance suite. The phase ordering encodes hard dependency rules from research: anti-relapse rules before code, IR before emitters, one resource end-to-end before broadening, filter parser before PATCH (PATCH value-paths reuse the filter grammar), concrete SQLite passing compliance before extracting the SQL SPI.

**Depth:** standard (config) — natural decomposition is 8 phases (research convergence). Phase 0 is non-optional per PITFALLS.md (three catastrophic, irreversible pitfalls must be encoded as written rules and PR gates before any emitted code). Phase 8 (Postgres dialect) is deferred to v1.x per PROJECT.md "Out of Scope"; tracked as the architectural validation gate.

## Phases

**Phase Numbering:**
- Integer phases (0-7): Planned milestone work for v1
- Decimal phases (e.g., 2.1): Urgent insertions if they arise (marked INSERTED)

- [ ] **Phase 0: Architecture Spike & Repo Setup** - Encode anti-relapse rules; multi-module workspace; emission-engine decision
- [ ] **Phase 1: Definition + IR + Validator** - Fluent Go builder, typed IR, generator-time invariant checks
- [ ] **Phase 2: Domain Emitter + Generator Pipeline** - End-to-end emit of typed User struct via the multi-emitter pipeline
- [ ] **Phase 3: Persistence + DDL + ETag (SQLite concrete-first)** - Per-resource generated repo, deterministic DDL, transactional version counter
- [ ] **Phase 4: HTTP Server + Discovery + List/Filter/Sort/Pagination** - Plain http.Handler with discovery endpoints and full filter grammar
- [ ] **Phase 5: PATCH + ETag/If-Match + Bulk** - RFC 7644 §3.5.2 PATCH engine, conditional writes, /Bulk fan-out
- [ ] **Phase 6: Multi-Resource (Group + EnterpriseUser + Custom Types)** - Built-in Group/EnterpriseUser definitions plus user-defined custom resources
- [ ] **Phase 7: CLI + Observability SPI + Compliance Test Emitter** - scim CLI, no-op observability defaults, emitted compliance suite

**Deferred to v1.x:** Phase 8 (Postgres dialect + driver) — architectural validation gate for the SQL SPI; PROJECT.md defers concrete non-SQLite drivers to post-v1.

## Phase Details

### Phase 0: Architecture Spike & Repo Setup
**Goal**: Encode the three catastrophic anti-relapse rules in writing and stand up the multi-module workspace before any emitted code exists.
**Depends on**: Nothing (first phase)
**Requirements**: REPO-01, REPO-02, REPO-03, REPO-04, REPO-05
**Success Criteria** (what must be TRUE):
  1. Contributor architecture rules document exists and explicitly forbids (a) generic Property/tree models, (b) runtime schema interpretation in generated servers, (c) IdP-specific accommodations
  2. PR template asks reviewers to confirm the change does not introduce IdP-specific tolerance and answer must be no
  3. Two `go.mod` modules exist (`gen` and `rt`) linked by a `go.work` file, with `go.work` gitignored
  4. CI pipeline builds every module both with `GOWORK=on` (workspace mode) and `GOWORK=off` (production-like) and a disagreement between the two fails the build
  5. Code-emission engine decision (`text/template` + `go/format` vs `dave/jennifer`) recorded in PROJECT.md Key Decisions with the spike evidence that drove it
**Plans**: 4 plans
  - [ ] 00-01-PLAN.md — CONTRIBUTING.md anti-relapse rules + PR template + CODEOWNERS + lefthook hooks
  - [ ] 00-02-PLAN.md — gen/rt modules + go.work gitignore + LICENSE + README + Makefile + GitHub Actions CI
  - [ ] 00-03-PLAN.md — emission-engine spike (text/template vs dave/jennifer) + PROJECT.md Key Decisions
  - [ ] 00-04-PLAN.md — reconcile REQUIREMENTS.md / ROADMAP.md / STACK.md per CONTEXT.md overrides

### Phase 1: Definition + IR + Validator
**Goal**: A user can describe a SCIM resource via a typed Go builder; mistakes surface at `go generate` time with file:line attribution back to the builder call, never in machine-generated code.
**Depends on**: Phase 0
**Requirements**: DEF-01, DEF-02, DEF-03, DEF-04, DEF-05, LIB-01, LIB-03, LIB-04
**Success Criteria** (what must be TRUE):
  1. User can author a User-resource definition by importing `scimdef` and chaining typed builder calls in a Go file (no `map[string]interface{}` in the IR)
  2. Validator rejects definitions with URN format errors, attribute name collisions, sub-attribute name violations, multi-valued complex `primary` ambiguities, or reserved attribute names — and the error message points to the builder call site (file:line)
  3. `scim dump-ir` emits the validated IR as JSON for debugging
  4. Generator is importable as a Go library (in-process generation works without invoking the CLI)
  5. Generator (`gen`) and runtime support library (`rt`) are separately versioned modules; `gen` never appears in any generated file's import list
**Plans**: TBD

### Phase 2: Domain Emitter + Generator Pipeline
**Goal**: Establish the three-stage Definition → IR → Multi-Emitter → gofmt pipeline by emitting a single artifact end-to-end: the typed `domain/user.go` with a `Validate()` method enforcing SCIM invariants at write time.
**Depends on**: Phase 1
**Requirements**: GEN-01, GEN-02, GEN-03, GEN-04, GEN-05, GEN-06, DOM-01, DOM-02, DOM-03
**Success Criteria** (what must be TRUE):
  1. Running the generator against a User definition produces a typed `User` struct (no generic `Resource` map) with a `Validate()` method that enforces required, caseExact, canonicalValues, and multi-valued primary cardinality
  2. Every generated file passes through `go/format.Source` (and `goimports` where applicable) before being written to disk; CI fails if `make regen && go vet ./... && git diff --exit-code` produces a diff
  3. Every generated file carries the `// Code generated ... DO NOT EDIT.` marker on its first non-blank line and uses the `*_gen.go` naming convention
  4. Multiple emitters run concurrently via `errgroup`; the failure of any emitter aborts the whole run with a clear error pointing to the offending emitter
  5. Two consecutive generator runs against an identical IR produce byte-identical output (deterministic regeneration; no map iteration nondeterminism, no embedded timestamps)
**Plans**: TBD

### Phase 3: Persistence + DDL + ETag (SQLite Concrete-First)
**Goal**: Generated per-resource SQL repositories hand-shaped against `database/sql`, with deterministic DDL emission and a per-resource monotonic version counter incremented in the same transaction as every write — built as a concrete SQLite implementation first, with the SPI extracted only after compliance passes.
**Depends on**: Phase 2
**Requirements**: PER-01, PER-02, PER-03, PER-04, PER-05, PER-06, DDL-01, DDL-02, DDL-03, ETAG-02
**Success Criteria** (what must be TRUE):
  1. Generated `UserRepo` exposes Create/Get/Replace/Delete operations, all participating in the same transaction as the version-counter increment (no embedded ORM, no runtime tree)
  2. `meta.version` is a per-resource monotonic counter incremented in the same transaction as every write — two consecutive writes always produce strictly increasing versions
  3. Generator emits a deterministic single-file `schema.sql` per resource set; two regen runs produce byte-identical SQL; the DDL is consumable by golang-migrate / goose / atlas naming conventions
  4. Persistence column shape carries a `tenant_id` (single value used in v1) so multi-tenancy can be added later without rewriting the repos
  5. `modernc.org/sqlite` reference driver implements the SQL SPI in `rt/sqldriver` and the SPI is extracted only after the SQLite implementation passes integration tests against the User repo (concrete-first; SPI shape informed by working seams)
**Plans**: TBD

### Phase 4: HTTP Server + Discovery + List/Filter/Sort/Pagination
**Goal**: Generated `http.Handler` exposing the full RFC 7644 read surface — CRUD endpoints, discovery payloads embedded at build time, full filter grammar producing parameterized SQL, sort/pagination/projection per spec.
**Depends on**: Phase 3
**Requirements**: HTTP-01, HTTP-02, HTTP-03, HTTP-04, HTTP-05, HTTP-06, DISC-01, DISC-02, DISC-03, DISC-04, LIST-01, LIST-02, LIST-03, LIST-04, LIST-05, LIST-06, LIST-07, LIST-08, LIB-02
**Success Criteria** (what must be TRUE):
  1. Generated handler is a plain `net/http` `http.Handler`; user can mount it on chi/gorilla/gin/std mux without third-party router dependencies leaking into generated code
  2. Discovery endpoints (`/Schemas`, `/ResourceTypes`, `/ServiceProviderConfig`) are served from `//go:embed` payloads generated from the same IR as the typed structs (single source of truth; no `os.ReadFile` for schema content at runtime)
  3. List endpoint accepts the full RFC 7644 §3.4.2.2 filter grammar (eq/ne/co/sw/ew/gt/lt/ge/le/pr, and/or/not, parens, value-path expressions, URN-prefixed extension paths) and translates the typed `FilterExpr` AST to parameterized SQL safe against injection; filter parser passes a 24h+ fuzz target with no crashes
  4. List responses use the `ListResponse` wrapper with correct `schemas` URN, 1-based `startIndex`, and integer (not string) `totalResults` reflecting the filter; `attributes`/`excludedAttributes` projection respects `returned: always` (never filtered out) and `returned: never` (never returned)
  5. Generated server enforces `application/scim+json` content type, returns RFC 7644 §3.12 error envelopes with string `status`, populates common attributes (`id`, `externalId`, `schemas`, `meta.*`) on every response, and returns the correct status codes (200/201/204/304/400/401/403/404/409/412/500/501); generated server consumes a typed `Config` struct constructed by the user's `main`
**Plans**: TBD

### Phase 5: PATCH + ETag/If-Match + Bulk
**Goal**: Ship the highest-complexity SCIM features — full RFC 7644 §3.5.2 PATCH (including value-path filters), `If-Match` enforcement on conditional writes, and `/Bulk` operations that fan out to existing handlers — with a comprehensive PATCH conformance corpus drawn from spec examples, the `scim-patch` corpus, and adversarial cases.
**Depends on**: Phase 4
**Requirements**: PATCH-01, PATCH-02, PATCH-03, PATCH-04, PATCH-05, PATCH-06, PATCH-07, PATCH-08, PATCH-09, ETAG-01, ETAG-03, ETAG-04, BULK-01, BULK-02, BULK-03, BULK-04
**Success Criteria** (what must be TRUE):
  1. PATCH supports the full RFC 7644 §3.5.2 op set (`add`, `replace`, `remove`) across all path forms — no path, attribute path, sub-attribute path, value-path filter (e.g. `emails[type eq "work"].value`) — and the multi-valued `value: [...]` array add form merges rather than replaces
  2. PATCH is atomic: all ops succeed or the resource is unchanged; filter-targeted Remove on multi-valued attributes removes only matching elements; Replace on a non-existent target behaves per RFC; removing required sub-attributes returns the correct error; `returned: never` and `mutability: readOnly` violations are rejected with the correct `scimType`
  3. PATCH engine lives in `rt/patch` and operates on typed resources via generated visitor methods (e.g. `(*User).ApplyPatch(op)`); no reflection over generated structs at runtime; PATCH conformance test corpus drawn from spec examples + `scim-patch` + adversarial cases ships with the runtime
  4. Every response carries a weak ETag (`W/"..."`) sourced from `meta.version`; `If-Match` on PUT/PATCH/DELETE returns 412 on mismatch; `If-None-Match: *` on POST is enforced for upsert protection
  5. `/Bulk` endpoint executes operations sequentially, resolves `bulkId` cross-references within a single request, honors `failOnErrors`, and fans out to the same per-resource handlers (no duplicated logic)
**Plans**: TBD

### Phase 6: Multi-Resource (Group + EnterpriseUser + Custom Types)
**Goal**: Prove the code-gen value-prop by broadening from one resource to many — built-in Group with the `User.groups` read-only-mirror invariant correctly handled, EnterpriseUser as the canonical schema extension, and user-declared custom resource types getting the full SCIM treatment.
**Depends on**: Phase 5
**Requirements**: RES-01, RES-02, RES-03, RES-04
**Success Criteria** (what must be TRUE):
  1. Built-in `User`, `Group`, and `EnterpriseUser` resource type definitions ship; running the generator against the default catalog produces a server that handles all three end-to-end (CRUD, discovery, filter, PATCH, Bulk)
  2. `Group.members` is the canonical store; `User.groups` is emitted as a read-only projection — PATCH targeting `User.groups` returns 400 with `scimType: mutability`; GET on `/Users/{id}` derives the `groups` field from the membership join
  3. EnterpriseUser appears as a `schemaExtension` of User in `/ResourceTypes` and round-trips correctly through CRUD/PATCH/list with extension URN-prefixed paths
  4. A user-defined custom resource type (e.g., `Device` or `Tenant`) declared in the definition gets the same emitted artifacts (typed struct, repo, handler, DDL, discovery payload, validator) as built-in types — proving custom types are not a special case
**Plans**: TBD

### Phase 7: CLI + Observability SPI + Compliance Test Emitter
**Goal**: Productisation — wrap the library in a Cobra-based CLI, ship pluggable observability interfaces with no-op defaults, and emit a SCIM v2 compliance test suite into the user's generated repo so consumers can prove conformance in CI without writing the tests themselves.
**Depends on**: Phase 6
**Requirements**: CLI-01, CLI-02, CLI-03, CLI-04, CLI-05, OBS-01, OBS-02, OBS-03, TEST-01, TEST-02, TEST-03, TEST-04, TEST-05, TEST-06
**Success Criteria** (what must be TRUE):
  1. `scim` CLI built on Cobra exposes `generate`, `check`, and `dump-ir` subcommands (CLI is a thin wrapper over the library API); `scim generate` runs the full pipeline against a definition entry point and writes a runnable repo; `scim check` warns on definition diffs that would break user-owned code
  2. CLI emits `cmd/server/main.go` that wires environment variables and flags into the typed `Config` struct (12-factor on top of the library API); the generated `main.go` is regenerable but user-editable (file naming distinguishes it from `*_gen.go`)
  3. `rt/observe` exposes Logger / Tracer / Meter SPIs with no-op default implementations — generated servers compile and run with no observability dependencies; user can wire concrete adapters (slog handler, OTel, Prometheus) via the `Config` struct
  4. Generator emits unit tests for handlers and repos, integration tests against the SQLite reference driver, golden-file snapshots of generator output, and a SCIM v2 compliance test suite into the user's repo at `internal/scimtest/` covering RFC 7643/7644 conformance
  5. Filter parser fuzz target (`go test -fuzz`) runs at least 24h in CI without crash before being declared done; "compiled examples" CI module compiles and runs the generated quickstart server end-to-end
**Plans**: TBD

## Progress

**Execution Order:**
Phases execute in numeric order: 0 → 1 → 2 → 3 → 4 → 5 → 6 → 7

| Phase | Plans Complete | Status | Completed |
|-------|----------------|--------|-----------|
| 0. Architecture Spike & Repo Setup | 2/4 | In Progress |  |
| 1. Definition + IR + Validator | 0/TBD | Not started | - |
| 2. Domain Emitter + Generator Pipeline | 0/TBD | Not started | - |
| 3. Persistence + DDL + ETag (SQLite) | 0/TBD | Not started | - |
| 4. HTTP Server + Discovery + List | 0/TBD | Not started | - |
| 5. PATCH + ETag/If-Match + Bulk | 0/TBD | Not started | - |
| 6. Multi-Resource (Group/EU/Custom) | 0/TBD | Not started | - |
| 7. CLI + Observability + Compliance | 0/TBD | Not started | - |

## Notes

**Deferred to v1.x (post-v1 architectural validation gate):**
- Phase 8: Postgres dialect + Postgres driver. Validates that the SQL SPI extracted in Phase 3 actually accommodates a second backend; per PROJECT.md "Out of Scope for v1", concrete non-SQLite drivers ship later behind the same interface. All four research dimensions flag this as the natural validation gate; the design milestone of "interface stress-tested against Postgres on paper" should land before v1 freeze even if the implementation defers.

**Open decisions to resolve in Phase 0 (per research Gaps to Address):**
- `text/template` + `go/format` vs `dave/jennifer` for Go source emission. STACK.md and ARCHITECTURE.md disagree explicitly; both agree `text/template` wins for non-Go output (SQL/Markdown/Makefile). Phase 0 spike emits one gnarly artifact in both styles to decide on evidence.

**Phases flagged for `/gsd:research-phase` during planning** (per SUMMARY.md):
- Phase 0: emission-engine decision spike
- Phase 4: filter parser-generator choice (participle vs pigeon PEG vs ANTLR-Go vs hand-rolled with property tests)
- Phase 5: PATCH applier shape (visitor methods vs reflection); spec ambiguity in §3.5.2.1; corpus assembly from `scim-patch` and real captured payloads

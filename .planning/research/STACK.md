# Stack Research

**Project:** Go SCIM v2 server code-generation toolkit (go-scim v3 / `next` branch)
**Domain:** Go code-generator toolkit — emits `http.Handler` + SQL persistence + migration DDL + tests for SCIM v2 servers
**Researched:** 2026-05-07
**Overall confidence:** HIGH (core choices); MEDIUM (definition format, compliance suite)

---

## Summary Recommendation (TL;DR)

| Dimension | Pick | One-liner |
|-----------|------|-----------|
| Go toolchain | **Go 1.25+** (CI tests 1.25 + 1.26) | Tool directives in `go.mod`, generic type aliases, Swiss Tables runtime, `go.work` workspaces |
| Multi-module workflow | **`go.work`** (gitignored) + per-module `go.mod` | Iterate generator + runtime in lockstep without `replace` hacks |
| Code generation | **`dave/jennifer`** (primary) + **`text/template`** (string-shaped fragments) | Programmatic Go AST with auto-managed imports; `gofmt` always |
| Definition format (v1) | **Fluent Go builder** ingested into an internal model | In-language, type-checked, refactor-friendly; YAML deferred |
| HTTP routing | **`net/http`** + Go 1.22+ enhanced `ServeMux` (no router dep in generated code) | Plain `http.Handler` is the contract; user mounts on chi/gorilla/whatever |
| SQL access in generated code | **`database/sql` + manually emitted typed query layer** (do NOT reach for sqlc/ent/sqlboiler) | Code-gen IS our ORM; embedding another generator is the wrong shape |
| SQLite driver | **`modernc.org/sqlite`** (CGo-free), pluggable | Cross-compile, single static binary, "good enough" perf; ship `mattn/go-sqlite3` as opt-in for write-heavy workloads |
| Migration emission | **Plain numbered `.sql` files** following golang-migrate naming | Standard format, runtime-tool-agnostic, user picks runner |
| Logging interface | **`log/slog`** as the canonical pluggable shape | Stdlib, handler-based, no third-party logger leakage into generated code |
| Testing | **stdlib `testing`** + **`testify/require`** (assertions only, no `suite`) | Idiomatic, ~27% adoption baseline, low ceremony |
| HTTP integration tests | **`net/http/httptest`** + in-memory SQLite | No containers needed for v1; SQLite IS our reference |
| SCIM compliance | **Internal table-driven RFC 7643/7644 conformance suite** (generated alongside server) | External suites are scaffolding-tests, not unit-test-grade; cite `python-scim/scim2-tester` and WSO2 as references |
| YAML (if added later) | **`goccy/go-yaml`** | `gopkg.in/yaml.v3` archived April 2025 |
| CLI framework | **`spf13/cobra`** for the `scimgen` CLI | Standard for Go CLIs; library remains the canonical entry point |

**Two stacks, one project:**

1. **Generator stack** (the toolkit itself, one `go.mod`): Jennifer, Cobra, slog, testify, go/parser+go/format for golden-file tests of emitted code.
2. **Runtime support stack** (imported by generated servers, second `go.mod`): zero third-party deps in v1 except the SQL drivers. Generated code consumes `net/http`, `database/sql`, `log/slog`, `encoding/json`, and a thin runtime support library exposing the persistence-driver interface, ETag helpers, filter/PATCH evaluators, and pluggable observability hooks.

The reason "generator stack" and "runtime stack" are separated below is the same reason this project needs a multi-module workspace: **the generator is heavyweight and the runtime must stay lean enough that consumers can audit every dep.** Generated SCIM servers should not transitively pull in `dave/jennifer`.

---

## Generator Stack (toolkit `go.mod`)

### Core Technologies

| Technology | Version | Purpose | Why Recommended |
|------------|---------|---------|-----------------|
| Go | 1.25.x floor; CI matrix 1.25 + 1.26 | Toolchain | Tool directives in `go.mod`, generic type aliases, Swiss Tables runtime, `go.work` workspaces. Go 1.25 floor per Phase 0 CONTEXT.md decision; user runs 1.26.1. [Confidence: HIGH] |
| `dave/jennifer` | v1.7.x | Programmatic Go source generation | Auto-tracks imports across conditional emission paths (the killer feature templates can't match); produces `gofmt`-clean output; widely used as the foundation of Go code generators (zerogen, go-contentful-generator, jennifer's own genjen). The code we're emitting (per-resource handlers, repos, validators, filter compilers, PATCH appliers) is exactly the "complex with conditional logic and many imports" case the project documentation flags as Jennifer-shaped. [Confidence: HIGH] |
| `text/template` | stdlib | String-shaped fragments only | For SQL DDL, `main.go` skeleton, README, Makefile — output that isn't Go AST. Mark all rendered files with the `// Code generated ... DO NOT EDIT.` regex from `golang.org/s/generatedcode`. Run all `.go` output through `go/format.Source` before write. [Confidence: HIGH] |
| `go/parser` + `go/format` | stdlib | Validate emitted Go before writing to disk | Parse-then-write catches malformed templates at generation time, not at user-build time. Cheap insurance. [Confidence: HIGH] |
| `spf13/cobra` | latest v1.x | CLI scaffolding for `scimgen` | De-facto standard for Go CLIs (kubectl, hugo, gh). The CLI is a thin wrapper over the library API per project decisions; cobra adds subcommand structure without coupling the library to it. [Confidence: HIGH] |
| `log/slog` | stdlib (Go 1.21+) | Generator's own logging | Use the same logging surface we ask generated code to use; eat our own dog food. Default `JSONHandler` for CI, `TextHandler` for local. [Confidence: HIGH] |

### Supporting Libraries (Generator)

| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| `stretchr/testify` | v1.10.x | `require`/`assert` for generator tests | Use `require` only — no `suite`, no `mock` (testify mocks become a maintenance pit). Idiomatic with `t *testing.T`. [Confidence: HIGH] |
| `google/go-cmp` | v0.6.x | Deep-equality diffs in golden tests | Better diff output than `reflect.DeepEqual`; pairs naturally with golden-file tests of generated code. [Confidence: HIGH] |

### What NOT to Use in the Generator

| Avoid | Why | Use Instead |
|-------|-----|-------------|
| `kong`, `urfave/cli/v2`, `alecthomas/kingpin` | Cobra is the larger ecosystem; legacy used `urfave/cli/v2` and we're explicitly making a clean break | `spf13/cobra` |
| `goccy/go-yaml` (in v1) | Definition format is fluent Go builder per project hypothesis; YAML support is a Pending decision | Defer — add only when YAML/JSON path is greenlit |
| `dolmen-go/codegen` (template wrapper) | Adds a third style on top of `text/template` + `dave/jennifer`; we already have two | `text/template` directly when you need it |
| `astutil` / `go/ast` for emission | `dave/jennifer` IS an AST builder — using `go/ast` to emit code is harder, no import management, no value-add | `dave/jennifer` |

---

## Runtime Support Stack (separate `go.mod`, imported by generated code)

The hard rule: **the runtime support library must stay drastically smaller than the generator.** Every dep here ships into every generated SCIM server. This list is deliberately short.

### Core Technologies

| Technology | Version | Purpose | Why Recommended |
|------------|---------|---------|-----------------|
| Go | 1.25+ | Same as generator | Generated code targets the same minimum |
| `net/http` | stdlib | HTTP handler contract | Project decision: generated server is a plain `http.Handler`. Go 1.22's enhanced `ServeMux` (`mux.HandleFunc("GET /Users/{id}", ...)`) is now sufficient for SCIM's REST surface — no third-party router needed inside the generated code. The user mounts on chi/gorilla/gin/whatever they want at the application level. [Confidence: HIGH] |
| `database/sql` | stdlib | SQL driver interface | The pluggable persistence abstraction is built on `database/sql`'s `driver.Driver` indirection. Each generated repo speaks `database/sql` directly; the SQLite/Postgres/MySQL choice is a runtime config of which `_ import` the user adds. [Confidence: HIGH] |
| `log/slog` | stdlib | Pluggable logging | Generated code accepts an `*slog.Logger` (or a small interface that `*slog.Logger` satisfies). Stdlib means zero deps in the runtime; user wires Datadog/OTel/Loki via slog handlers. The Handler/Record/Logger split is the canonical "pluggable backend" shape — exactly what the project's "Pluggable observability" requirement asks for. [Confidence: HIGH] |
| `encoding/json` | stdlib | SCIM payload serialization | SCIM is JSON-over-HTTP. Stdlib is fine for v1; revisit `goccy/go-json` or `bytedance/sonic` only if profiling shows JSON is the bottleneck. Generated structs carry `json:` tags emitted from the definition. [Confidence: HIGH] |
| `modernc.org/sqlite` | v1.46+ (latest 1.49.x) | SQLite driver (default) | **CGo-free.** Single static binary, trivial cross-compilation, no `gcc` required to build a generated SCIM server. December 2025 prepared-statement fix closed a long-standing perf gap (~39% faster on prepared reads). 2026 benchmarks show modernc is competitive with mattn on most workloads, slightly behind on heavy single-connection inserts, slightly ahead on concurrent reads. The distribution win (no CGo) is decisive for a code-generator: we cannot saddle every generated repo with a CGo toolchain requirement. [Confidence: HIGH] |
| `mattn/go-sqlite3` | v1.14.x | SQLite driver (opt-in alternative) | Document as the swap-in for write-heavy workloads. Same `database/sql` interface — user changes one blank import. We do not pick a default that requires CGo. [Confidence: HIGH] |

### Supporting Libraries (Runtime)

| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| `google/uuid` | v1.6.x | RFC 4122 UUID generation for resource IDs | Stdlib has no UUID; `google/uuid` is the de-facto standard with active maintenance (legacy used the dead `satori/go.uuid` — do not inherit). [Confidence: HIGH] |

That's the entire runtime third-party surface. Anything else (filter parsing, PATCH evaluation, ETag helpers, SCIM error envelope, conditional-request handling per RFC 7232) is internal code in the runtime support library — no external dep needed.

### What NOT to Use in the Runtime

| Avoid | Why | Use Instead |
|-------|-----|-------------|
| **`sqlc` / `ent` / `sqlboiler` / `gorm`** | We ARE a code generator. Embedding another schema-driven Go-emitter inside our emission pipeline is two generators fighting over the same generated package. sqlc wants SQL-first; ent wants schema-as-Go-struct; sqlboiler wants schema-from-live-DB; gorm wants reflection. None of them know about SCIM's filter grammar, PATCH semantics, ETag concurrency rules, or the multi-valued `primary` invariant. We emit per-resource repos directly against `database/sql` — that's the whole point of choosing code-gen over a generic data model. | Hand-emit `database/sql` query methods using `dave/jennifer` |
| `julienschmidt/httprouter` | Legacy choice. Go 1.22 `ServeMux` covers the same ground in stdlib; SCIM endpoints are simple REST patterns. | `net/http` `ServeMux` |
| `gorilla/mux` | Archived December 2022 (read-only). Not a stable foundation for a new project. | `net/http` `ServeMux`; user mounts on chi at app level if they want middleware groups |
| `chi` (in generated code) | We emit a plain `http.Handler` — the user is free to wrap it in chi/gorilla/gin in their own `main`. Baking chi into the runtime forces it on every consumer. | Plain `http.Handler` — let consumers compose |
| `rs/zerolog` | Faster than slog but a third-party logger surface in the generated runtime is exactly the dep we don't want. slog is good enough; user can wire zerolog underneath via a slog handler if they really care. | `log/slog` |
| `gopkg.in/yaml.v3` | Archived April 2025 by upstream. Not a stable dep. | `goccy/go-yaml` (only if YAML definitions land later) |
| `satori/go.uuid` | Unmaintained; legacy used it; known security issues in older versions. | `google/uuid` |

---

## Migration Emission

The generator emits SQL migration DDL. **Choice: emit plain numbered files matching the `golang-migrate` naming convention** (`0001_create_users.up.sql` / `0001_create_users.down.sql`).

**Rationale:**
- `golang-migrate` and `goose` both consume this format; `atlas` can also import it. Users pick whichever runner fits their CI; we don't force one.
- The generator owns `up.sql`. Emitting `down.sql` (best-effort reverse) is a v1 nice-to-have; if reverse can't be safely synthesized, omit it and document.
- We do NOT ship a runtime migration runner inside the generated server. Per project decisions ("user owns runtime correctness", "schema-evolution migrations between generator runs out of scope"), the generator emits DDL fresh each run; running it is the user's problem with their tool of choice.

**What NOT to use:**
| Avoid | Why | Use Instead |
|-------|-----|-------------|
| Atlas HCL as the emission format | Couples users to Atlas. Atlas can read plain `.up.sql`/`.down.sql` files just fine. | Plain SQL files with golang-migrate naming |
| Embedded migration runner (e.g., `golang-migrate` as a runtime dep) | Drags a dep into every generated server for a feature 80% of users will replace with their own pipeline | Emit files; user runs them |
| `gorm.AutoMigrate` | Reflection-based; we already rejected GORM | N/A |

[Confidence: HIGH on format choice; MEDIUM on always-emit-down-sql — depends on what we can synthesize reliably]

---

## SCIM v2 Compliance Testing

**There is no Go-native SCIM v2 compliance suite in active maintenance** as of 2026-05. The actively-maintained ones are:

| Suite | Language | Status | Use For |
|-------|----------|--------|---------|
| `python-scim/scim2-tester` | Python | Active (PyPI 0.2.5) | External end-to-end checks via CLI; can run in CI as a black-box |
| WSO2 SCIM2 Compliance Test Suite | Java (WAR) | Active but heavyweight | Reference for which RFC clauses to cover |
| `suvera/scim2-compliance-test-utility` | Java | Maintained | Reference |
| `ltsch/mock-scim-server` | Python | Useful as a test harness, not a suite | IdP simulation |

**Recommendation:** Generate an **internal Go conformance test suite** alongside each server. Table-driven tests against RFC 7643 (schema) and RFC 7644 (protocol) — Discovery endpoints, CRUD round-trips, PATCH semantics, filter grammar, sort, pagination, ETag/If-Match, error envelope. Reference the WSO2 suite's test catalog for coverage; do not run it in CI (Java + WAR deployment is too much friction).

Optionally (post-v1) wire `python-scim/scim2-tester` into a CI job for cross-validation against an external implementation — flag as a research task for the conformance phase.

[Confidence: MEDIUM — recommendation is informed by the gap in Go-native suites; deeper research warranted at the conformance phase]

---

## HTTP / SCIM-Specific Conventions

The project decision is "plain `http.Handler`, no opinionated router." Confirming best practices:

- **Routing:** Go 1.22+ `ServeMux` supports `mux.HandleFunc("GET /Users/{id}", h)` and `r.PathValue("id")`. Sufficient for SCIM's URL shape. No third-party router in generated code.
- **Content negotiation:** SCIM mandates `application/scim+json`. Generated handlers read `Content-Type` and emit `Content-Type: application/scim+json` per RFC 7644 §3.1.
- **ETags (RFC 7232 + RFC 7644 §3.14):** SCIM resources carry `meta.version` which doubles as the HTTP `ETag`. Emit weak ETags (`W/"..."`) — generated dynamically, not byte-perfect. Strong-comparison for `If-Match` (PUT/PATCH/DELETE concurrency); weak-comparison for `If-None-Match` on GET. **Always quote** the value (`"\"abc123\""`) — Go's `http.ServeContent` requires RFC 7232-compliant quoting and will silently misbehave if you forget.
- **Errors:** SCIM error response is a specific JSON envelope (`{"schemas":["urn:ietf:params:scim:api:messages:2.0:Error"], "status":"400", "scimType":"...", "detail":"..."}`) — emit a typed `scim.Error` in the runtime support library.
- **PATCH:** SCIM PATCH (RFC 7644 §3.5.2) is a custom format, NOT JSON Patch RFC 6902. Operations are `add`/`replace`/`remove` with SCIM filter `path` expressions. Generated PATCH appliers walk the typed resource — this is one of the major code-gen wins (the legacy tree-model approach pays for it with the `ExclusivePrimarySubscriber` hack documented in CONCERNS.md §2).

[Confidence: HIGH on routing/ETag mechanics; MEDIUM on PATCH applier shape — flag for arch-phase deep dive]

---

## Multi-Module Workspace Layout

```
go-scim/
  go.work                    # gitignored
  go.work.sum                # gitignored
  generator/
    go.mod                   # module github.com/imulab/go-scim/generator
    cmd/scimgen/main.go      # CLI entry point (cobra)
    pkg/...                  # library API
    internal/...             # jennifer-based emitters
  runtime/
    go.mod                   # module github.com/imulab/go-scim/runtime
    scim/...                 # types, errors, ETag helpers
    persist/...              # Driver interface; SQLite reference impl
    httpkit/...              # ServeMux helpers, content negotiation
  examples/
    go.mod                   # module github.com/imulab/go-scim/examples
    quickstart/...           # generated output committed for reference
```

**Rules:**
- `go.work` is **gitignored** — it lists local paths that don't exist on other machines. CI runs with `GOWORK=off` so each module is validated against its `go.mod`.
- `runtime` releases first, then `generator` — generated code imports `runtime` by its public path, so the import must resolve before generator users can build.
- The `examples` module exists so generated output can be committed and continuously compiled in CI. This is the canary that detects "generator changed but generated code stopped compiling" before users hit it.

[Confidence: HIGH]

---

## Versions Pin Recommendation

```
# Generator go.mod
go 1.25

require (
    github.com/dave/jennifer v1.7.1
    github.com/spf13/cobra v1.8.1
    github.com/stretchr/testify v1.10.0
    github.com/google/go-cmp v0.6.0
)

# Tool directive (Go 1.25+)
tool (
    github.com/imulab/go-scim/generator/cmd/scimgen
)
```

```
# Runtime go.mod
go 1.25

require (
    github.com/google/uuid v1.6.0
    modernc.org/sqlite v1.49.1   // default driver (CGo-free)
)

// Optional, documented as opt-in alternative
// require github.com/mattn/go-sqlite3 v1.14.42
```

[Confidence: HIGH on `dave/jennifer`, `cobra`, `testify`, `go-cmp`, `google/uuid`; HIGH on `modernc.org/sqlite` recency. Verify exact patch versions at implementation time — these were current as of 2026-05.]

---

## Alternatives Considered

| Recommended | Alternative | When the Alternative Makes Sense |
|-------------|-------------|----------------------------------|
| `dave/jennifer` for code-gen | `text/template` only | If output is mostly fixed boilerplate with simple substitutions and you don't generate Go (e.g., emitting SQL/Markdown/Makefile). We use `text/template` for exactly those non-Go artifacts. |
| `dave/jennifer` for code-gen | `statictemplate` | Runtime HTML rendering performance — irrelevant to our build-time emission |
| `database/sql` directly | `sqlc` | If you wanted SQL-first generic code-gen for an arbitrary schema. We need SCIM-aware emission, not generic. |
| `database/sql` directly | `ent` | If you wanted a general-purpose schema-as-Go-code ORM. We define schemas as SCIM resources, not Ent entities — the abstraction layers don't compose. |
| `modernc.org/sqlite` | `mattn/go-sqlite3` | Heavy single-connection write workloads where the ~30% insert perf gap matters AND you can require CGo in your build environment |
| `modernc.org/sqlite` | `ncruces/go-sqlite3` (WASM) | Read-heavy workloads where its 3x read advantage matters AND you accept a less-mainstream driver |
| `net/http` ServeMux | `chi` | If your application needs middleware groups, subrouters, custom 404/405 handlers — but at the app level, not inside our generated handler. Generated code stays plain `http.Handler`. |
| Plain SQL migration files | Atlas HCL | Atlas is excellent if your team is already standardized on it. Plain SQL is the lowest-common-denominator format Atlas can also consume. |
| Internal Go conformance tests | `python-scim/scim2-tester` in CI | Cross-validation against an external implementation as a post-v1 quality gate |
| `testify` (require only) | `ginkgo`/`gomega` | Cloud-native ecosystem projects where BDD DSL is the team norm (Kubernetes-adjacent code). Not idiomatic for a code-generator toolkit. |

---

## Stack Patterns by Variant

**If a future user wants Postgres or MySQL:**
- Add a sibling driver implementation in `runtime/persist/postgres/` and `runtime/persist/mysql/`. Same `Driver` interface. `database/sql` abstracts the wire-level differences.
- Generator emits per-driver SQL DDL when the user picks the driver in the definition (or in the CLI flag).
- Use `pgx` underneath (not `lib/pq`, which is in maintenance mode) — but expose only the `database/sql` shape to the generated code to keep the dep contained.

**If YAML/JSON definition format ships later:**
- Reach for `goccy/go-yaml` (active, spec-compliant, replaces archived `gopkg.in/yaml.v3`) for YAML.
- Use stdlib `encoding/json` + JSON Schema validation (`santhosh-tekuri/jsonschema/v6`) for JSON.
- Both feed the same internal model the fluent Go builder produces — definition format is a parser layer, not an architectural fork.

**If multi-tenancy lands:**
- Persistence-layer tenant-id awareness is already a v1 requirement (project decisions). The interface shape — `Driver.GetByID(ctx, tenantID, id)` etc. — should be designed in v1 even though only the `default` tenant is used.
- No additional library is needed; the change is in interface design, not stack.

---

## Version Compatibility Notes

| Package A | Compatible With | Notes |
|-----------|-----------------|-------|
| `modernc.org/sqlite` v1.49+ | Go 1.25+ | December 2025 prepared-statement fix is in v1.46.1+ — pin minimum at 1.46.1 to get the perf fix |
| `dave/jennifer` v1.7+ | Go 1.18+ | No generics requirement; safe to bump Go floor without re-evaluating |
| Go 1.22+ ServeMux pattern syntax | Go 1.22+ | If we drop `1.25` floor for any reason, the ServeMux pattern API still requires 1.22 minimum. Don't go below 1.22. |
| `log/slog` | Go 1.21+ | Stdlib since 1.21; safe assumption |
| `testify` v1.10 | Go 1.18+ | Drop `suite` package; use `require`/`assert` only |
| `goccy/go-yaml` (if used) | Go 1.20+ | Active maintenance; not blocked by archived upstream |

---

## What This Stack Solves vs. Legacy

Mapping back to `.planning/codebase/CONCERNS.md`:

| Legacy Concern | This Stack's Answer |
|----------------|---------------------|
| §1 Tree-model → SQL impedance | Code-gen emits per-resource `database/sql` repos — no tree-to-relational translation needed. We reject sqlc/ent/sqlboiler/gorm precisely because the SCIM-aware emission is the value-add. |
| §2 Raw()/Hash() inefficiency, primary-rule subscribers | Generated typed structs make the "exactly one primary" rule a compile-checkable invariant in the generated PATCH applier — no event-propagating subscriber walking siblings. |
| §3 No portability accommodation | Out of v1 scope per project decisions, but the definition format is the natural future home for compatibility profiles. The fluent Go builder makes per-attribute case-policies trivial to add. |
| §4.1 Projection bug | Per-resource generated repos make projection a code-gen concern; the bug class disappears when projection is statically known per resource. |
| §6 Visitor traversal cost | Generated code doesn't visit a tree — it knows the shape. PATCH/serialize/validate become direct field access. |
| §7.1 JSON deserializer state machine | Stdlib `encoding/json` with generated structs replaces the 715-line manual scanner. |
| §10.1 Deprecated `io/ioutil` | Go 1.25 floor; we never write `io/ioutil`. |

---

## Sources

### Primary (HIGH confidence)
- [Go 1.24 Release Notes — go.dev](https://go.dev/doc/go1.24) — toolchain features, version verification
- [dave/jennifer GitHub](https://github.com/dave/jennifer) — code generator API
- [sqlc 1.31.1 docs](https://docs.sqlc.dev/en/latest/) — verified plugin architecture and SQLite tutorial recommends `modernc.org/sqlite`
- [chi v5.2.3 GitHub](https://github.com/go-chi/chi) — release recency check
- [Atlas Releases](https://github.com/ariga/atlas/releases) — current declarative HCL features
- [modernc.org/sqlite Go Packages](https://pkg.go.dev/modernc.org/sqlite) — pure-Go transpilation, no CGo
- [RFC 7232 — HTTP Conditional Requests](https://www.rfc-editor.org/rfc/rfc7232.html) — ETag semantics
- [RFC 7643 — SCIM Schema](https://datatracker.ietf.org/doc/html/rfc7643)
- [RFC 7644 — SCIM Protocol](https://datatracker.ietf.org/doc/html/rfc7644)
- [Go Workspaces Tutorial](https://go.dev/doc/tutorial/workspaces) — multi-module workflow
- [slog package — Go Blog](https://go.dev/blog/slog) — Handler/Record/Logger architecture

### Comparative (MEDIUM confidence — multiple independent sources triangulated)
- [Encore — Comparing the best Go ORMs (2026)](https://encore.cloud/resources/go-orms)
- [Rost Glukhov — Which ORM to use in Go (March 2025)](https://www.glukhov.org/post/2025/03/which-orm-to-use-in-go/)
- [johal.in — GORM 2 vs Ent 0.14 vs SQLBoiler 4.10](https://www.johal.in/go-orm-comparison-gorm-vs-ent-014-vs/)
- [cvilsmeier/go-sqlite-bench (March 2026 run)](https://github.com/cvilsmeier/go-sqlite-bench) — modernc vs mattn benchmarks
- [DataStation — SQLite in Go, with and without CGo](https://datastation.multiprocess.io/blog/2022-05-12-sqlite-in-go-with-and-without-cgo.html) — historical baseline
- [Atlas — Picking a database migration tool for Go (2023)](https://atlasgo.io/blog/2022/12/01/picking-database-migration-tool)
- [Alex Edwards — Which Go Router Should I Use?](https://www.alexedwards.net/blog/which-go-router-should-i-use)
- [Calhoun.io — Go's 1.22+ ServeMux vs Chi Router](https://www.calhoun.io/go-servemux-vs-chi/)
- [JetBrains — The Go Ecosystem in 2025](https://blog.jetbrains.com/go/2025/11/10/go-language-trends-ecosystem-2025/) — testify adoption stats

### Ecosystem Status (LOW confidence individually, HIGH when corroborated)
- [Issue: gopkg.in/yaml.v3 archived](https://github.com/go-task/task/issues/2171) — corroborated by multiple project migration discussions
- [goccy/go-yaml GitHub](https://github.com/goccy/go-yaml) — active maintenance verification
- [python-scim/scim2-tester GitHub](https://github.com/python-scim/scim2-tester) — closest active SCIM compliance suite
- [WSO2 SCIM2 Compliance Test Suite](https://github.com/wso2-incubator/scim2-compliance-test-suite) — reference catalog

---

## Open Questions for Downstream Phases

1. **PATCH applier shape** — generated typed PATCH applier vs. a runtime helper that takes a typed resource and a SCIM PATCH AST. Affects how much code Jennifer emits per resource. Flag for architecture phase.
2. **Filter compiler shape** — emit a per-resource compiled filter evaluator vs. emit a generic SCIM-filter-to-SQL translator the runtime owns. Either works; tradeoff is generated-code size vs. runtime-library complexity. Flag for architecture phase.
3. **Down-migration emission feasibility** — can the generator reliably synthesize `down.sql` for arbitrary SCIM resource diffs? If not, document and emit `up.sql` only. Flag for migration-emission phase.
4. **External compliance suite integration** — should `python-scim/scim2-tester` be a CI job? Adds Python toolchain dep to generator development. Defer to conformance-test phase.
5. **Postgres/MySQL driver timing** — v1 ships SQLite only per project decisions, but the persistence interface must accommodate them. Verify the interface design with a paper Postgres implementation early to catch leaky abstractions.

---

*Stack research for: Go SCIM v2 server code-generation toolkit*
*Researched: 2026-05-07*
*Researcher: gsd-project-researcher (Stack dimension)*
*Updated 2026-05-07 (Phase 0 plan) — Go floor bumped 1.24 → 1.25 per CONTEXT.md decision. CI matrix tests 1.25 + 1.26.*

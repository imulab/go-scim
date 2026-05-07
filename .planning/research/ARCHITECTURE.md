# Architecture Research

**Domain:** Go code-generation toolkit (SCIM v2 server emitter)
**Researched:** 2026-05-07
**Confidence:** HIGH (architectural pattern is well-established across ent/sqlc/Buf/oapi-codegen; SCIM-specific specifics are MEDIUM)

## Standard Architecture

This research surveys how mature Go code-generation toolkits (ent/entc, sqlc, oapi-codegen, kubebuilder, Buf+protoc-gen-connect-go, Smithy, AWS SDK code-gen) are architected, then opinionates a recommended shape for a SCIM v2 server-emitting toolkit.

The dominant pattern across all six toolkits is a **three-stage pipeline**: `Definition → Loader/Parser → IR (typed in-memory model) → Emitter → Generated artifacts`, with a thin separately-versioned **runtime support library** that the generated code imports.

### System Overview

```
┌──────────────────────────────────────────────────────────────────────────┐
│                        DEFINITION LAYER (user-authored)                   │
│   ┌──────────────────────────────────────────────────────────────┐       │
│   │  Fluent Go builder DSL (in-language, type-checked)           │       │
│   │  scim.Resource("User").                                       │       │
│   │      Attribute(scim.String("userName").Required().Unique()).  │       │
│   │      MultiValued("emails", scim.Complex(...).PrimaryUnique()) │       │
│   └──────────────────────────────────┬───────────────────────────┘       │
└──────────────────────────────────────┼───────────────────────────────────┘
                                       │
┌──────────────────────────────────────▼───────────────────────────────────┐
│                      LOADER / PARSER LAYER (generator)                    │
│   ┌──────────────────────┐   ┌─────────────────────────────────────┐     │
│   │  Definition runner    │──►│  Validator (compile-time invariants)│     │
│   │  (executes builder    │   │  - schema URN uniqueness            │     │
│   │   to materialise spec)│   │  - primary cardinality on multi-VC  │     │
│   └──────────────────────┘   │  - sub-attribute name collisions    │     │
│                              │  - reserved-attribute presence       │     │
│                              └─────────────────┬───────────────────┘     │
└────────────────────────────────────────────────┼─────────────────────────┘
                                                 │
┌────────────────────────────────────────────────▼─────────────────────────┐
│                      IR LAYER (in-memory typed model)                     │
│   ┌──────────────────────────────────────────────────────────────┐       │
│   │  Catalog                                                       │       │
│   │  ├─ Schemas[]   (URN, attributes, sub-attributes)              │       │
│   │  ├─ ResourceTypes[]                                            │       │
│   │  │   ├─ Endpoint, Schema URN, Extensions                       │       │
│   │  │   └─ ResolvedAttributes[] (flat, name-mangled)              │       │
│   │  ├─ ServiceProviderConfig                                      │       │
│   │  └─ Indexes, UniqueConstraints, FilterableAttrs                │       │
│   └──────────────────────────────────────────────────────────────┘       │
└────────────────────────────────────────────────┬─────────────────────────┘
                                                 │
┌────────────────────────────────────────────────▼─────────────────────────┐
│                            EMITTER LAYER                                   │
│   ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐      │
│   │ domain   │ │persistence│ │  server  │ │   spec   │ │migrations│      │
│   │ emitter  │ │  emitter  │ │ emitter  │ │ emitter  │ │ emitter  │      │
│   └────┬─────┘ └─────┬────┘ └────┬─────┘ └────┬─────┘ └────┬─────┘      │
│        │             │           │            │            │             │
│        └─────────────┴───────────┴────────────┴────────────┘             │
│                              │                                            │
│                       gofmt + goimports                                   │
└──────────────────────────────┼────────────────────────────────────────────┘
                               │
┌──────────────────────────────▼────────────────────────────────────────────┐
│                  GENERATED MODULE (user owns, builds, runs)                │
│  cmd/server/main.go            (typed Config wiring; user edits freely)   │
│  internal/server/router.go     (http.Handler routing; regenerated)        │
│  internal/server/handlers/     (per-resource HTTP handlers)               │
│  internal/domain/{user,group,...}.go   (typed Go structs + validation)   │
│  internal/persistence/{user,group,...}.go  (typed repos, SQL queries)    │
│  internal/migrations/0001_init.sql                                        │
│  internal/spec/embed.go        (//go:embed of /Schemas, /ResourceTypes)  │
│  internal/scimtest/compliance_test.go                                     │
└──────────────────────────────┬────────────────────────────────────────────┘
                               │ imports
┌──────────────────────────────▼────────────────────────────────────────────┐
│              RUNTIME SUPPORT LIBRARY (separate go.mod, semver-stable)      │
│  scimrt/filter        (SCIM filter parser → AST + evaluator interface)    │
│  scimrt/path          (SCIM path parser)                                  │
│  scimrt/patch         (RFC7644 PATCH operation engine)                    │
│  scimrt/scimerr       (error → HTTP status mapping per RFC7644)           │
│  scimrt/etag          (Weak ETag computation, version field handling)     │
│  scimrt/scimjson      (application/scim+json content negotiation, errors)│
│  scimrt/sqldriver     (driver SPI: Tx, Exec, Query, Migrate)             │
│  scimrt/sqldriver/sqlite  (reference impl)                                │
│  scimrt/observe       (Logger, Tracer, Meter SPIs — interfaces only)     │
└───────────────────────────────────────────────────────────────────────────┘
```

### Component Responsibilities

| Component | Responsibility | Typical Implementation |
|-----------|----------------|------------------------|
| **Definition (user-facing)** | Express SCIM resources, attributes, constraints | Fluent Go builder. User imports `scimdef` package and writes a Go file that returns a `*scimdef.Catalog`. |
| **Loader/Parser** | Execute the definition; surface a normalized model | Plain Go function call (definition is already Go code). For YAML alternative, a parser would deserialize → builder calls. |
| **Validator** | Enforce SCIM invariants at generator time | Pure functions over the IR. Errors stop generation with file:line attribution from the builder call site. |
| **IR (Catalog)** | Single source of truth for all emitters | Flat, normalized structs. No Go AST nodes. JSON-serializable for debugging (cf. sqlc `dumpcatalog`, ent `load.SchemaSpec`). |
| **Emitters (multi)** | Produce one concern each (domain, persistence, server, spec, migrations, tests) | Each is `func Emit(*Catalog, *Config) ([]File, error)`. Composable via a `Generator` interface (cf. ent's `gen.Generator`). |
| **Code-emission engine** | Turn structured intent into Go source | `text/template` + `go/format` (recommendation; see Pattern 2) |
| **Runtime support lib (`scimrt`)** | What's identical across all generated servers | Separate `go.mod`. Filter parser, PATCH engine, errors, ETag, content negotiation, driver SPIs. |
| **SQL driver SPI** | Pluggable persistence backend | Small interface (`Exec`, `Query`, `Tx`, `Migrate`). SQLite is reference impl in v1; Postgres/MySQL slot in later. |
| **Observability SPI** | No-op-by-default Logger/Tracer/Meter interfaces | Generated code calls them; user wires zerolog/OTel/Prometheus in `main.go`. Library ships **only interfaces**. |
| **Generated `cmd/server/main.go`** | Typed `Config` wiring; user-editable | Emitted once, never overwritten on regen (or regenerated only if user opts in). |

## Recommended Project Structure

### Repo layout (the toolkit itself — multi-module workspace)

```
go-scim/                                    # workspace root
├── go.work                                 # links all modules below
├── scimgen/                                # MODULE 1: generator (the tool)
│   ├── go.mod                              # module github.com/imulab/go-scim/scimgen
│   ├── cmd/scimgen/main.go                 # CLI entry point (Cobra/urfave)
│   ├── scimdef/                            # fluent builder DSL (user-imported by definitions)
│   │   ├── catalog.go                      # Catalog, Resource, Attribute, Schema builders
│   │   ├── types.go                        # String/Int/Bool/Complex/MultiValued constructors
│   │   ├── validate.go                     # invariant checks at builder time
│   │   └── builtin.go                      # User, Group, EnterpriseUser presets
│   ├── ir/                                 # intermediate representation
│   │   ├── catalog.go                      # post-validation IR types (flat, JSON-friendly)
│   │   ├── resolve.go                      # name resolution, attribute path indexing
│   │   └── debug.go                        # dump-ir command (cf. sqlc dumpcatalog)
│   ├── emit/                               # emitters (one subpackage per artifact group)
│   │   ├── domain/                         # typed structs + validation methods
│   │   ├── persistence/                    # SQL repo + queries (driver-aware)
│   │   ├── server/                         # http.Handler, routing, middleware wiring
│   │   ├── spec/                           # /Schemas, /ResourceTypes, /SPC payload embedders
│   │   ├── migrations/                     # SQL DDL files
│   │   ├── compliance/                     # SCIM compliance test fixtures
│   │   └── templates/                      # text/template files (//go:embed'd)
│   ├── sqldialect/                         # SQL dialect awareness for the persistence emitter
│   │   ├── dialect.go                      # interface (Quote, Limit/Offset, JSON column type, etc.)
│   │   ├── sqlite.go                       # reference dialect
│   │   └── (postgres.go, mysql.go later)
│   └── pipeline.go                         # top-level: Generate(catalog, cfg) error
├── scimrt/                                 # MODULE 2: runtime support library
│   ├── go.mod                              # module github.com/imulab/go-scim/scimrt
│   ├── filter/                             # SCIM filter parser + evaluator interface
│   ├── path/                               # SCIM path parser
│   ├── patch/                              # RFC7644 PATCH op engine (operates on driver SPI)
│   ├── scimerr/                            # spec-defined errors + HTTP mapping
│   ├── scimjson/                           # application/scim+json helpers, error payloads
│   ├── etag/                               # version field + weak ETag helpers
│   ├── observe/                            # Logger, Tracer, Meter interfaces (no impls)
│   └── sqldriver/                          # SQL driver SPI
│       ├── driver.go                       # the interface
│       └── sqlite/                         # reference impl (uses modernc.org/sqlite — pure Go, no cgo)
└── examples/                               # MODULE(S) 3+: full generated examples for testing/demo
    ├── basic-user-group/                   # generated server demonstrating core resources
    │   ├── go.mod
    │   ├── definition.go                   # uses scimdef
    │   └── (generated server tree)
    └── custom-resource/                    # generated server with a custom resource type
```

### Generated module layout (what `scimgen` emits — user owns this)

```
my-scim-server/                             # user's repo
├── go.mod                                  # depends on scimrt only (NOT scimgen)
├── definition.go                           # user-edited (scimdef DSL)
├── cmd/server/
│   └── main.go                             # generated once, user-editable; wires Config
├── internal/
│   ├── server/
│   │   ├── router.go                       # http.Handler factory, mounts handlers
│   │   ├── handlers/                       # one file per resource type (regenerated)
│   │   │   ├── user.go
│   │   │   └── group.go
│   │   └── middleware.go                   # generated stubs; user-extensible
│   ├── domain/                             # typed Go structs (one file per resource)
│   │   ├── user.go
│   │   ├── group.go
│   │   └── validation.go                   # SCIM invariant enforcement (primary, etc.)
│   ├── persistence/                        # typed repos
│   │   ├── user_repo.go                    # CRUD + filter + sort + paginate
│   │   ├── group_repo.go
│   │   └── tx.go                           # transactional patch helpers
│   ├── migrations/
│   │   ├── 0001_init.sql                   # full DDL for all resource tables
│   │   └── embed.go                        # //go:embed *.sql
│   ├── spec/                               # discovery payloads
│   │   ├── schemas.json                    # embedded /Schemas response
│   │   ├── resource_types.json             # embedded /ResourceTypes response
│   │   ├── service_provider_config.json    # embedded /ServiceProviderConfig response
│   │   └── embed.go
│   └── scimtest/                           # SCIM compliance suite, pre-fixtured for the user's resources
│       ├── compliance_test.go
│       └── fixtures/
└── Makefile                                # generated convenience targets (test, run, regen)
```

### Structure Rationale

- **Two modules, not one:** `scimgen` (generator) and `scimrt` (runtime) are versioned independently. Generator changes (new emitter feature) shouldn't bump every existing generated server's runtime dep. `scimrt` is semver-stable; `scimgen` can iterate. This is the same split as protoc / google.golang.org/protobuf, sqlc / sqlc-gen-go, and ent / ent runtime.
- **`scimdef` lives inside the `scimgen` module** because the user's definition is consumed *by the generator*, not the generated server. The generated server has zero dependency on `scimdef`. This avoids the trap of pulling builder-DSL types into a server's runtime image.
- **Generated module imports `scimrt` only.** No reflection on definition objects at runtime. No tree model. Strong, statically-typed call sites against thin SPIs.
- **`internal/` for everything generated.** Prevents downstream consumers from importing the generated server's internals; user can break them at will. Only `cmd/server/main.go` is the public surface.
- **Migrations as SQL files (not Go AST).** The user can run them via any migration tool (golang-migrate, goose, atlas) or pipe them through their CI. Generator emits fresh DDL; user owns diffing in v1 (per PROJECT.md out-of-scope).
- **`scimtest/` co-located with the generated code.** Fixtures know the user's resource shape. Externalizing them would require runtime introspection — exactly what we're avoiding.

## Architectural Patterns

### Pattern 1: Multi-Stage Pipeline with In-Memory IR

**What:** Definition → Loader → IR → Multiple Emitters → gofmt → Files. Each stage has a single, testable responsibility. Emitters never read the definition directly; they consume the IR.

**When to use:** Always. This is the universal pattern across ent (`load.Schema` → `gen.Graph` → templates), sqlc (SQL → `catalog.Catalog` → codegen plugin via protobuf), oapi-codegen (OpenAPI YAML → typed schema model → `ParseTemplates`), kubebuilder (Go AST + markers → `APIs` struct → generators), and Smithy (IDL → JSON AST → language plugins).

**Trade-offs:**
- **Pro:** Each stage independently unit-testable. IR can be dumped (`dump-ir` command) for debugging — sqlc's `dumpcatalog` is the model.
- **Pro:** Multiple emitters can read the same IR concurrently (cf. ent uses `errgroup` to parallelize).
- **Pro:** New artifacts (e.g., add OpenAPI emission later) plug in as new IR consumers without touching the loader.
- **Con:** Two model representations (definition + IR) means a translation layer. Worth it.

**Example (pseudo-Go):**
```go
// Definition (user code)
catalog := scimdef.NewCatalog().
    Resource("User", scimdef.UserResourceType()).
    Resource("Group", scimdef.GroupResourceType()).
    Build()

// Pipeline (scimgen library API)
ir, err := scimgen.LoadIR(catalog)        // validation happens here
if err != nil { return err }
err = scimgen.Generate(ir, scimgen.Config{
    OutDir:   "./",
    Module:   "github.com/me/my-scim",
    Dialect:  sqldialect.SQLite,
})
```

### Pattern 2: Code Emission via `text/template` + `go/format`

**What:** Use Go's standard `text/template` to write source files as templated text, then post-process every file through `go/format` (programmatic `gofmt`) and `golang.org/x/tools/imports` (programmatic `goimports`).

**When to use:** This is the recommended approach. Justification:

- **Most Go code generators in production use templates.** ent, sqlc, oapi-codegen, kubebuilder, protoc-gen-go all use Go templates. The pattern is well-trodden.
- **Templates resemble their output.** A reviewer reading the User struct template can predict what gets emitted. With `dave/jennifer`, a 3-line struct becomes 8 lines of `.Id().Type().Struct(...)` chained calls — output is opaque without mentally executing the calls.
- **Most output is mostly-static for SCIM.** A SCIM HTTP handler's structure is highly regular (decode → validate → call repo → encode). Templates excel at "mostly static, some holes."
- **Refactoring is cheap.** Changing the shape of generated code = editing a template. With Jennifer, it's restructuring chained calls across emitter functions.

**Trade-offs:**
- **Pro:** Lowest dependency footprint (stdlib only for emission core).
- **Pro:** Generated-code changes are diff-able as template diffs.
- **Pro:** Template helpers (Funcs map) handle all type-name mangling, import resolution.
- **Con:** No compile-time check that the generator produces valid Go — you only find out at `go/format` time. Mitigation: every emitter unit test runs `go/format` on output and fails on errors.
- **Con:** Import management is manual; fix by running `goimports` programmatically on every file before write.

**Decision matrix from research:**

| Need | Best Tool | Use here? |
|------|-----------|-----------|
| Generate brand-new files with mostly-static structure | `text/template` | **YES — primary** |
| Generate brand-new files with highly dynamic structure | `dave/jennifer` | No (only ~5% of our output is dynamic) |
| Modify existing files preserving comments | `dave/dst` | No (we never modify; we always emit) |
| Read/analyze existing Go code | `go/ast`+`go/types` | No (we don't introspect Go source) |

**Real-world signal:** The Go community pattern that emerged in maintenance retrospectives is *"`go/ast` is for reading, jennifer/templates for writing, and they're frequently used together."* Since we don't read Go source, we don't need `go/ast` at all. Since most output is mostly-static, `text/template` wins.

**Example:**
```go
//go:embed templates/*.tmpl
var templates embed.FS

var domainTmpl = template.Must(
    template.New("domain").Funcs(funcMap).ParseFS(templates, "templates/domain.tmpl"))

func emitDomain(rt *ir.ResourceType, w io.Writer) error {
    var buf bytes.Buffer
    if err := domainTmpl.Execute(&buf, rt); err != nil {
        return err
    }
    formatted, err := format.Source(buf.Bytes())   // gofmt
    if err != nil {
        return fmt.Errorf("emit %s: invalid Go produced: %w\n--- raw output ---\n%s",
            rt.Name, err, buf.String())
    }
    _, err = w.Write(formatted)
    return err
}
```

### Pattern 3: Compile-Time Validation, Runtime Defense

**What:** Push every invariant possible into the validator that runs against the IR before emission. Generated code carries minimal runtime checks — only those that depend on input data (e.g., per-request validation that a payload doesn't violate `mutability=readOnly`).

**When to use:** Always. This is the explicit antidote to legacy `CONCERNS.md` Section 2 (efficiency: ExclusivePrimarySubscriber walking siblings at runtime to enforce "only one primary").

**Trade-offs:**
- **Pro:** Cardinality, uniqueness, name-collision, schema-URN format, attribute-path resolution, etc. all become generator errors with file:line on the `scimdef` call. User sees mistakes at `go generate` time, not at request time.
- **Pro:** Generated code is leaner — no subscriber tree, no event propagation, no reactive rule engine.
- **Pro:** "Exactly one element with `primary=true`" becomes a generated method `func (es Emails) Validate() error { ... count primaries, error if >1 }` called once per request, O(n).
- **Con:** Definition mistakes only caught at `go generate` time (acceptable — that's a developer-time check).

**Example:**
```go
// In ir/resolve.go (validator):
func ValidatePrimaryUniqueness(rt *ResourceType) error {
    for _, attr := range rt.MultiValuedComplex() {
        if attr.HasSubAttribute("primary") && !attr.HasGeneratorMarker("PrimaryUnique") {
            return fmt.Errorf(
                "%s.%s: multi-valued complex attribute has 'primary' sub-attribute "+
                "but builder did not assert PrimaryUnique() — "+
                "SCIM allows at most one primary; assert it explicitly",
                rt.Name, attr.Name)
        }
    }
    return nil
}

// In emit/domain/template.tmpl (runtime check, narrow scope):
func ({{.Receiver}} {{.Type}}) Validate() error {
    {{- range .MultiValuedAttrs}}
    {{- if .PrimaryUnique}}
    if err := scimrt.ExactlyOnePrimary({{$.Receiver}}.{{.GoFieldName}}); err != nil {
        return scimerr.Invalid("{{.Path}}", err)
    }
    {{- end}}
    {{- end}}
    return nil
}
```

### Pattern 4: Driver SPI in Runtime Library, Dialect Awareness in Emitter

**What:** The runtime library defines a small `sqldriver` interface that handles connection lifecycle, transactions, and execution. The *emitter* knows the SQL dialect (table syntax, type names, JSON column form) and emits driver-agnostic SQL strings the driver simply executes.

**When to use:** Always. This mirrors sqlc's split: the `sqlite` driver in sqlc-gen is a code-generation-time concern (type mapping, SQL dialect quirks), but the runtime DB connection is a thin `database/sql`-style abstraction.

**Trade-offs:**
- **Pro:** Adding Postgres = new dialect file in `scimgen/sqldialect/postgres.go` + new driver impl in `scimrt/sqldriver/postgres/`. No churn in the IR or in other emitters.
- **Pro:** Generated code uses concrete SQL strings (fast, debuggable) rather than runtime AST translation (legacy mongo/v2/filter.go pattern, which was a translator and is exactly what we're avoiding).
- **Pro:** The runtime driver SPI is tiny (Tx, Exec, Query, Close, Migrate). Less to mock; less to break.
- **Con:** SCIM filter expressions still need runtime translation to SQL `WHERE` (we can't bake that into the emitter because filter strings are user data). Solution: `scimrt/filter` parses to an AST; the generated repo passes the AST to a per-dialect translator (also in `scimrt`, dialect-tagged).

**Example interface shape:**
```go
// scimrt/sqldriver/driver.go
package sqldriver

type Driver interface {
    Exec(ctx context.Context, query string, args ...any) (Result, error)
    Query(ctx context.Context, query string, args ...any) (Rows, error)
    BeginTx(ctx context.Context) (Tx, error)
    Migrate(ctx context.Context, ddl []string) error
    Close() error
}

// Generated repo uses it:
func (r *UserRepo) Create(ctx context.Context, u *domain.User) error {
    _, err := r.db.Exec(ctx, sqlInsertUser, u.ID, u.UserName, u.Active /* ... */)
    return err
}
```

### Pattern 5: Observability via Bare Interfaces in Runtime Library

**What:** Define `Logger`, `Tracer`, `Meter` interfaces in `scimrt/observe`. Ship no concrete implementations. Provide no-op defaults so generated code never panics on a nil logger.

**When to use:** Always for libraries the user owns. Avoids dragging zerolog/OTel/Prometheus dependencies into every generated server.

**Trade-offs:**
- **Pro:** User picks zerolog vs slog vs zap; OTel vs Datadog; Prometheus vs StatsD. No transitive-dep flame wars.
- **Pro:** Generated code remains tiny.
- **Con:** Users must wire concrete impls in `cmd/server/main.go`. Mitigation: emit a commented-out example block in `main.go` with the most common wiring (zerolog + OTel) under `// uncomment to enable observability`.

## Data Flow

### Generator-time data flow

```
User's definition.go
        │
        │  go run ./cmd/scimgen generate ./definition.go
        ▼
┌──────────────────────────────┐
│  scimdef.Catalog (in memory)  │
└──────────────┬───────────────┘
               ▼
┌──────────────────────────────┐
│  Validator (invariant checks) │ ─── error ──► fail with file:line
└──────────────┬───────────────┘
               ▼
┌──────────────────────────────┐
│  ir.Catalog (typed IR)        │ ─── --dump-ir ──► JSON for debugging
└──────────────┬───────────────┘
               ▼
   ┌───────┬───┴───┬───────┬───────┐
   ▼       ▼       ▼       ▼       ▼
domain  persist  server   spec   migrations  (parallel emit via errgroup)
   │       │       │       │       │
   └───────┴───┬───┴───────┴───────┘
               ▼
       gofmt + goimports
               ▼
           Files on disk
```

### Generated-server-runtime data flow (HTTP CREATE)

```
HTTP POST /scim/v2/Users  (application/scim+json)
        │
        ▼
internal/server/router.go      (mux dispatches to user.go handler)
        │
        ▼
internal/server/handlers/user.go
   1. scimrt/scimjson.Decode(body, &domain.User)
   2. domain.User.Validate()                 ◄── runtime SCIM invariants
   3. scimrt/etag.Compute(&user)
   4. persistence.UserRepo.Create(ctx, &user)
        │
        ▼
internal/persistence/user_repo.go
   - executes parameterised INSERT via scimrt/sqldriver
        │
        ▼
scimrt/sqldriver/sqlite (or pg, mysql)
        │
        ▼
SQL database
        │
        ▲ row
        │
        ▼
scimrt/scimjson.Encode(user, w)
        │
        ▼
HTTP 201 Created  (Location, ETag, application/scim+json body)
```

### PATCH flow (the hardest one)

```
HTTP PATCH /scim/v2/Users/{id}
        │
        ▼
handlers/user.go
   1. scimrt/scimjson.DecodePatch(body) → []scimrt.PatchOp
   2. repo.GetForUpdate(tx, id) → domain.User
   3. for each op:
        scimrt/patch.Apply(&user, op, scimdef.UserSchema())
          ├─ scimrt/path.Parse(op.Path)
          ├─ apply per spec semantics (add/replace/remove)
          └─ enforce mutability=readOnly server-side
   4. user.Validate()
   5. repo.Replace(tx, &user)
   6. tx.Commit()
```

The PATCH engine is in the **runtime library** (`scimrt/patch`) because patch semantics are spec-defined and identical for every resource. It operates on the typed `domain.User` via reflection or — preferably — via generated visitor methods (`func (u *User) ApplyPatch(op PatchOp) error`) that the emitter can produce. The visitor approach avoids reflection entirely; this is a research item for the emit phase.

## Scaling Considerations

This is a generator toolkit, not a SaaS. "Scale" here means **scale of emitted code maintenance and breadth of definitions supported**, not request throughput.

| Scale | Architecture Adjustments |
|-------|--------------------------|
| 1-3 resource types (MVP) | Single emit pass, no parallelism, single `text/template`. Validate end-to-end manually. |
| 5-15 resource types (typical) | Parallelize emit per resource via `errgroup` (cf. ent). Add `--watch` mode for definition iteration. |
| 30+ resource types or large extension chains | Cache parsed IR keyed on definition file hash. Profile template execution. Consider per-resource file outputs to keep diffs small. |

### Scaling Priorities

1. **First bottleneck (DX, not perf):** generator compile errors that point to the *generated file* instead of the *definition*. Fix early by attaching `caller.File():caller.Line()` to every builder method invocation and threading it through to error messages.
2. **Second bottleneck (correctness):** template + IR drift. Fix by snapshot tests (`testdata/golden/`) for every supported resource type — every emit is diffed against a checked-in golden output. This is industry standard for codegen tools.
3. **Third bottleneck (perf):** template execution time on 30+ resources. Fix only if measured; `errgroup` parallelism is a 5-line change.

## Anti-Patterns

### Anti-Pattern 1: Generic Runtime Tree Model (the legacy mistake)

**What people do:** Maintain a `Property` interface with N implementations and a `Navigator` that walks at runtime. Treat resources as "schema-aware nodes" that interpret their `spec.Attribute` reference on every operation.

**Why it's wrong:** This is exactly what `.planning/codebase/ARCHITECTURE.md` documents and `.planning/codebase/CONCERNS.md` flags as the rewrite driver. Specifically:
- Tree → SQL impedance is fundamental, not bridgeable (Concern #1).
- Constraint enforcement becomes reactive subscribers walking siblings (Concern #2: `ExclusivePrimarySubscriber` is O(n) per primary set).
- `Raw()` rebuilds slices on every call; `Hash()` is O(n²) (Concern #2).
- Five+ traversals per PATCH (Concern #6.1).

**Do this instead:** Generate concrete typed Go structs per resource type. SCIM invariants compile into method calls, not subscriber tree traversals. The runtime never sees the `spec.Attribute` graph — it's baked into types.

### Anti-Pattern 2: Generator Library Imports the Runtime Library (or vice versa)

**What people do:** `scimgen` depends on `scimrt`, or `scimrt` imports types from `scimgen`. Often justified as "we need to share the schema types."

**Why it's wrong:**
- Generator code (templates, AST manip) leaks into runtime image of every generated server. Bloat.
- Versioning the two together means every generator iteration breaks every prior generated server.
- Exactly opposite of how protoc/protobuf-go and sqlc/sqlc-gen-go are structured.

**Do this instead:** Strict separation. `scimgen` produces text. `scimrt` provides services to the generated text. They share zero types. If something feels like it needs to be shared (e.g., schema URN constants), put it in `scimrt` and have the emitter reference it by literal string.

### Anti-Pattern 3: One Mega-Template for the Whole Server

**What people do:** Single `server.go.tmpl` that emits routing + handlers + domain + persistence in one file.

**Why it's wrong:**
- Templates become unreadable past ~200 lines.
- Single failure mode (one bad branch breaks the whole emit).
- User can't review a focused diff when adding a resource — every regen rewrites the world.

**Do this instead:** One template per concern (one per file in the generated output). Each template has a tight, named purpose. Templates compose data via shared `Funcs` (e.g., `{{ goFieldName .Path }}`).

### Anti-Pattern 4: Runtime Reflection on Definition Objects

**What people do:** Generated code holds a reference to the `scimdef.Catalog` and reflects over it at request time to drive validation.

**Why it's wrong:** This re-creates the legacy generic-tree problem under a different name. The whole point of code generation is to *eliminate* runtime introspection.

**Do this instead:** Invariants either become validator-time errors (invalid definition rejected at `go generate`), or they become straightforward boolean checks in the generated `Validate()` method (e.g., `if countPrimary(u.Emails) > 1 { return ... }`).

### Anti-Pattern 5: Using `dave/jennifer` for Mostly-Static Output

**What people do:** Reach for `jennifer` because "it's the typed AST way" and "templates are old-fashioned."

**Why it's wrong:**
- 3 lines of Go become 8 lines of Jennifer chains.
- Reviewer has to mentally execute the generator to predict the output.
- Refactoring the generated code shape becomes restructuring emitter functions instead of editing a template.
- Real-world maintenance reports describe moving *off* AST-based generation back to templates due to maintenance cost.

**Do this instead:** Use `text/template`. Keep `jennifer` as an option only for the small slice of output that's truly type-driven and dynamic (e.g., generating filter-evaluator switch tables on attribute paths) — and even then, evaluate whether a template helper function suffices first.

## Integration Points

### External Services (the generator)

| Service | Integration Pattern | Notes |
|---------|---------------------|-------|
| User's filesystem | Read definition file, write generated tree | Use `io/fs` abstractions so we can unit-test with `fstest.MapFS`. |
| `gofmt` / `goimports` | In-process via `go/format` and `golang.org/x/tools/imports` | Never shell out. Failure should attach raw output for debugging. |
| `go.work` | Generated example modules link to `scimgen`/`scimrt` via go.work for dev | Production users consume `scimrt` via versioned go.mod. |

### External Services (the generated server)

| Service | Integration Pattern | Notes |
|---------|---------------------|-------|
| SQL database | `scimrt/sqldriver.Driver` interface | SQLite reference impl uses `modernc.org/sqlite` (pure Go, no cgo). |
| HTTP framework | Plain `http.Handler` | User mounts on chi/gorilla/gin/std mux. Zero opinion. |
| Identity provider (Okta, Entra) | Standard SCIM v2 protocol | Quirks (Microsoft capitalization) explicitly out of v1 scope per PROJECT.md. |
| Logging/Tracing/Metrics | `scimrt/observe` interfaces; user wires concrete in `main.go` | Library ships no impls. |

### Internal Boundaries

| Boundary | Communication | Notes |
|----------|---------------|-------|
| `scimdef` ↔ `ir` | Function call (`LoadIR(catalog)`) | One-way. IR never reaches back. |
| `ir` ↔ emitters | Each emitter takes `*ir.Catalog` | Emitters share IR; do not mutate. |
| Emitters ↔ templates | Templates read structured data + Funcs map | Funcs map is the only escape hatch. |
| `scimgen` ↔ `scimrt` | **None** at compile/import time | Templates emit *string references* to `scimrt` packages; no Go-level dependency. |
| Generated server ↔ `scimrt` | Standard Go import | Versioned via go.mod. |
| Generated server ↔ user code (`main.go`) | Typed `Config` struct passed to `server.New(cfg)` | User's `main` constructs Config from env/flags/file. |

## Build Order / Phase Decomposition

This is the most actionable output for the roadmap. Based on the dependency graph above and the canonical pattern across ent/sqlc/oapi-codegen, the build order is:

### Phase Ordering (recommended for roadmap)

1. **Definition + IR + Validator (no emission yet)**
   - Build `scimdef` fluent builder for the User resource only.
   - Build `ir.Catalog` types.
   - Build validator with the SCIM invariants we know upfront (URN format, primary uniqueness, attribute name collisions).
   - **Test artifact:** `dump-ir` command produces JSON; round-trip a hand-written User definition.
   - **Why first:** every later phase consumes the IR. Get its shape right before building emitters against it.

2. **Domain emitter only (one resource type, no SQL, no HTTP)**
   - Single `text/template` emits `domain/user.go` with typed struct + `Validate()`.
   - Snapshot-test the output (golden file).
   - No runtime library needed yet — generated struct is pure data.
   - **Why second:** smallest possible end-to-end emit slice. Validates the template/format/snapshot loop.

3. **Runtime library skeleton + persistence emitter (SQLite only)**
   - `scimrt/sqldriver` interface + SQLite impl.
   - `scimrt/filter` parser to AST (no evaluator yet).
   - `migrations` emitter (DDL for User table).
   - `persistence` emitter (UserRepo with Create/Get/Replace/Delete; filter/sort/paginate Phase 4).
   - **Test artifact:** can do CRUD against generated repo + SQLite.
   - **Why third:** persistence is the longest pole; doing it for *one* resource forces the runtime/generated split decisions.

4. **HTTP server emitter + filter/sort/pagination + spec endpoints**
   - `internal/server` emission (router + User handlers).
   - Filter AST → SQL translator (in `scimrt/filter/sqlite`).
   - `internal/spec` discovery payloads (/Schemas, /ResourceTypes, /SPC).
   - **Test artifact:** end-to-end SCIM CRUD + List+filter against running server.
   - **Why fourth:** needs Phase 3's persistence. Server + filter + spec all share the discovery-payload IR work.

5. **PATCH + bulk + ETag + RFC7644 errors**
   - `scimrt/patch` engine (most complex piece in runtime).
   - `scimrt/etag` weak ETag support.
   - `scimrt/scimerr` error → status mapping covering all RFC7644 cases.
   - Bulk endpoint emission.
   - **Why fifth:** PATCH semantics are gnarly; doing them after the simpler CRUD cycle is well-debugged is the right order.

6. **Group + EnterpriseUser + custom resource type support**
   - Extend `scimdef` builtins for Group, EnterpriseUser.
   - Stress-test the emitter on multi-resource setups (extension schemas, references between resources).
   - **Why sixth:** at this point the emitter is mature; broadening from 1→N resources mostly exercises code paths already proven.

7. **CLI + observability SPIs + compliance test fixture emitter**
   - `cmd/scimgen` CLI (Cobra/urfave/cli) on top of the library.
   - `scimrt/observe` interfaces + no-op defaults.
   - `scimtest` compliance suite emission.
   - **Why seventh:** these are productisation, not architecture. Easier to do once the inner loop is stable.

8. **Postgres dialect + Postgres driver impl**
   - Validates the SQL dialect SPI design with a second backend.
   - **Why last:** if the SPI is right, this is a configuration + dialect file + driver impl. If the SPI is wrong, this phase will surface it — and it's better to find out before more emitters depend on it.

### Critical phase dependencies

- Phase 2 (domain emitter) **must** prove the template/snapshot/format loop before anything else is built. Get this loop right; everything else inherits it.
- Phase 3 (persistence) **must** establish the runtime/generated split and the driver SPI shape. Every later phase respects these boundaries.
- Phase 5 (PATCH) is the **hardest single piece**. Schedule slack here.
- Phase 8 (Postgres) is the **architectural validation gate**. If it requires changes to the SPI, that's a flag the SPI was wrong; don't paper over.

## How This Architecture Solves the Three Legacy Rewrite Drivers

(Per `.planning/codebase/CONCERNS.md` — this is mandatory per the quality gate.)

| Legacy Concern | Recommended Architecture's Answer |
|----------------|-----------------------------------|
| **#1 Feasibility: tree model incompatible with SQL** | Generated typed structs per resource → straightforward generated typed repos with parameterised SQL. No tree-to-relational impedance. SQLite is the reference; Postgres slots into the same SPI. |
| **#2 Efficiency: generic model duplication + reactive rule enforcement** | Invariants compile into validator-time errors (definition rejected) or generated `Validate()` methods (single O(n) pass). No subscribers, no event propagation, no runtime tree traversal. `Raw()` and `Hash()` simply don't exist — the generated struct *is* the data. |
| **#3 Portability: no IdP-quirk accommodation hooks** | Out of v1 scope per PROJECT.md, but the architecture leaves the door wide open: the **definition layer** is the natural future home for compatibility profiles (e.g., `scimdef.Profile(scimdef.MicrosoftEntra)` that toggles capitalization rules at IR time). The emitter remains spec-strict; profiles inject normalization into the generated decode path. |

## Sources

### Primary architectural references (HIGH confidence)

- [ent (entgo.io)](https://entgo.io/) — multi-stage `load → graph → generate` pipeline; canonical reference for Go schema-driven codegen
- [ent/entc/load package docs](https://pkg.go.dev/entgo.io/ent/entc/load) — the loader pattern (parse Go schemas → JSON-marshallable IR)
- [ent/entc/gen package docs](https://pkg.go.dev/entgo.io/ent/entc/gen) — the IR (Graph + Type) and template-based emitter
- [sqlc GitHub](https://github.com/sqlc-dev/sqlc) — SQL → catalog IR → codegen plugin pipeline
- [sqlc generate docs](https://docs.sqlc.dev/en/latest/howto/generate.html) — `dumpcatalog` / `dumpast` debug commands as model for IR introspection
- [sqlc SQLite tutorial](https://docs.sqlc.dev/en/latest/tutorials/getting-started-sqlite.html) — pluggable database backend pattern (SQLite + Postgres + MySQL behind same config)
- [Announcing sqlc-gen-go](https://sqlc.dev/posts/2023/11/06/publishing-sqlc-gen-go/) — separation of generator from runtime; protobuf-IPC plugin contract
- [Buf Connect getting started (connectrpc.com)](https://connectrpc.com/docs/go/getting-started/) — IDL-driven approach (schema → generated stubs + thin runtime library)
- [protoc-gen-connect-go docs](https://pkg.go.dev/connectrpc.com/connect/cmd/protoc-gen-connect-go) — runtime/generated split (connect-go runtime, protoc-gen-connect-go emitter)
- [oapi-codegen GitHub](https://github.com/oapi-codegen/oapi-codegen) — OpenAPI YAML → typed schema model → ParseTemplates emitter
- [kubebuilder codegen](https://pkg.go.dev/sigs.k8s.io/kubebuilder/cmd/kubebuilder-gen/codegen) — Generator interface pattern; markers + Go AST as definition source
- [Smithy](https://smithy.io/) — IDL-first protocol-agnostic codegen; AWS SDK uses Smithy server SDKs as runtime support libs

### Code-emission approach (HIGH confidence)

- [dave/jennifer GitHub](https://github.com/dave/jennifer) — typed Go AST emission library (alternative considered, rejected for this use case)
- [text/template Go docs](https://pkg.go.dev/text/template) — recommended emission engine (used by ent, sqlc, oapi-codegen, kubebuilder, protoc-gen-go)
- [Code generation pros/cons (Mehdi Khalili)](https://www.mehdi-khalili.com/code-generation-pros-cons-t4-template) — broader maintainability discussion
- [Metaprogramming with Go (DEV Community, hlubek)](https://dev.to/hlubek/metaprogramming-with-go-or-how-to-build-code-generators-that-parse-go-code-2k3j) — real-world report on `go/ast` IR maintenance cost
- [Reducing boilerplate with go generate (Gopher Academy)](https://blog.gopheracademy.com/advent-2015/reducing-boilerplate-with-go-generate/) — community pattern for `text/template + go/format`
- [Code generation in Go: tools and use cases (Developers Heaven)](https://developers-heaven.net/blog/code-generation-in-go-tools-and-use-cases/) — landscape comparison

### Multi-module workspace + CLI patterns (HIGH confidence)

- [Tutorial: Getting started with multi-module workspaces (go.dev)](https://go.dev/doc/tutorial/workspaces) — official guidance on `go.work`
- [How to Manage Multi-Module Go Projects with Workspaces](https://oneuptime.com/blog/post/2026-01-25-multi-module-go-projects-workspaces/view) — 2026 best-practice synthesis
- [chi router GitHub](https://github.com/go-chi/chi) — example of stdlib-`http.Handler`-compatible routing (model for the generated server's mux-portability constraint)
- [Making and Using HTTP Middleware in Go (Alex Edwards)](https://www.alexedwards.net/blog/making-and-using-middleware) — middleware composition pattern

### SCIM-specific (MEDIUM confidence — combined with primary spec)

- [RFC 7643 — SCIM Core Schema](https://datatracker.ietf.org/doc/html/rfc7643)
- [RFC 7644 — SCIM Protocol](https://datatracker.ietf.org/doc/html/rfc7644) (referenced in [Wikipedia overview](https://en.wikipedia.org/wiki/System_for_Cross-domain_Identity_Management))
- [elimity-com/scim](https://github.com/elimity-com/scim) — existing Go SCIM library (architectural reference, not a target)
- [arturoeanton/goscim](https://github.com/arturoeanton/goscim) — alternative Go SCIM impl (uses ANTLR for filters)
- [Okta SCIM 2.0 reference](https://developer.okta.com/docs/api/openapi/okta-scim/guides/scim-20) — IdP-side expectations
- [How to build a SCIM 2.0 endpoint (Scalekit Blog)](https://www.scalekit.com/blog/build-scim-endpoint) — implementation pitfalls

### Internal references (HIGH confidence)

- `.planning/PROJECT.md` — project goals, constraints, decisions
- `.planning/codebase/ARCHITECTURE.md` — legacy tree model (anti-pattern reference)
- `.planning/codebase/STRUCTURE.md` — legacy module organisation
- `.planning/codebase/CONCERNS.md` — three rewrite drivers this architecture must address

---

*Architecture research for: Go SCIM v2 server code-generation toolkit*
*Researched: 2026-05-07*

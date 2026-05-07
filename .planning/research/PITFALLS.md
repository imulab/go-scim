# Pitfalls Research

**Domain:** Go code-generation toolkit for SCIM v2 servers
**Researched:** 2026-05-07
**Confidence:** HIGH for SCIM/RFC claims (verified against RFC 7643, 7644, 9865 and IdP vendor docs); HIGH for Go tooling (verified against go.dev, pkg.go.dev, multiple 2026-dated benchmarks); MEDIUM for "common mistakes" attributed to vendor implementations (multiple corroborating sources but not exhaustive).

This file deliberately calls out three classes of pitfall:
- **A. SCIM v2 spec/protocol pitfalls** — what RFC 7643/7644/9865 demand and what gets implemented wrong.
- **B. Go code-generator pitfalls** — Go-specific failure modes (not generic codegen advice).
- **C. Project-specific traps** — relapses into the legacy `go-scim` failure modes documented in `.planning/codebase/CONCERNS.md`.

The downstream consumer (roadmap) needs each pitfall to map to (1) detection, (2) prevention, and (3) the phase that should own the prevention. Phase names below are illustrative — the roadmap may rename them but should preserve ordering.

---

## Critical Pitfalls

### Pitfall 1: Reintroducing a generic in-memory tree "for flexibility"

**What goes wrong:**
The team starts code-generating per-resource Go types, then halfway through someone proposes a `Resource` interface with `Get(path string) any` / `Set(path string, v any)` "so generic services can operate on any resource type." Filter evaluation, validation, JSON shaping all migrate to this generic layer. Six months later, the project has rebuilt the legacy tree model with extra indirection, every "primary uniqueness" rule lives in a runtime subscriber, the SQL adapter still doesn't fit cleanly, and code-gen has become a thin veneer over a runtime interpreter.

**Why it happens:**
- Code that operates on a known concrete type (`*User`) feels less reusable than code that operates on `Resource`. The reuse instinct points back toward genericity.
- SCIM endpoints (`/Schemas`, `/ResourceTypes`, `/ServiceProviderConfig`, error paths) genuinely do need to handle unknown resource types at runtime. It's tempting to extend that runtime polymorphism into the data plane.
- Generated code feels "dumb" — engineers want the elegance of reflection or dynamic dispatch.

**Why it's catastrophic for THIS project:**
This is the exact failure mode that drove the rewrite (CONCERNS.md §1, §2). The tree model is what made SQL infeasible. Allowing it back means the rewrite produced nothing.

**How to avoid:**
- Architectural rule, written into the contributor docs: **the generated code is the data plane; runtime SCIM-aware code lives only in (a) discovery endpoints and (b) the `crud/expr` filter parser (a pure tree of values, not properties).** No `Property` interface. No `Navigator` outside filter evaluation. No `Subscriber` system.
- Filter evaluation walks Go reflect/codegen-emitted accessors, not a property tree. Even better: emit per-resource filter compilers that produce parameterized SQL plus an in-Go fallback.
- Code review checklist item: "Does this introduce a runtime-typed Resource interface?" If yes, justify in writing.

**Warning signs:**
- A new `interface { Path(string) Property }` appears in the generator-runtime support library.
- Generated structs gain a `func (u *User) Get(path string) any` method that wraps an internal map.
- A subscriber/event/visitor pattern is proposed for SCIM rule enforcement (e.g., "primary exclusivity").
- Anyone uses the phrase "we can do this generically with reflection."

**Phase to address:**
Phase 0 (architecture spike) — encode the rule before the first line of generator output is written. Phase 2 (core generator) — enforce by what is emitted. Every subsequent phase — code review.

---

### Pitfall 2: Reintroducing runtime schema interpretation

**What goes wrong:**
Definitions get parsed at generator time, but the generated server still loads a JSON schema file at startup, registers it in a global `schemaRegistry`, and dispatches PATCH/filter operations through a path resolver that consults the registry at request time. Result: per-request schema lookups, panics on missing registrations, cannot ship the binary without the schema files alongside, schema/code drift becomes possible.

**Why it happens:**
- The legacy code did this (`pkg/v2/spec/schema.go`, CONCERNS.md §7.3) and is the available mental model.
- `/Schemas` and `/ResourceTypes` endpoints have to serve schema content, so engineers reason "we already need schemas at runtime, might as well use them everywhere."
- "What if the user wants to hot-reload a schema?" — a feature request that has no business being satisfied in v1.

**How to avoid:**
- Schemas are inputs to the generator, not inputs to the generated server. The generated server embeds (`embed.FS`) the canonical schema JSON it serves on `/Schemas`. That JSON is generated from the same definition the types are generated from — single source of truth.
- No global mutable registry. If discovery endpoints need a registry, it is constructed once at server start from compile-time-embedded data and is read-only.
- The PATCH/filter machinery never asks "is this attribute mutable?" at request time by looking up a schema record. It asks the generated, type-specific validator method (e.g., `(*User).validatePatchTarget(path, op)`).

**Warning signs:**
- Generator output references `os.ReadFile("schemas/User.json")` at server runtime.
- A `package spec` in the generated module has package-level mutable state.
- Engineers ask, "where do we register the User schema?" — there should be no register step.

**Phase to address:**
Phase 0 (architecture spike) and Phase 2 (core generator: schema embedding strategy).

---

### Pitfall 3: Conflating spec compliance with IdP compatibility (v1 scope creep)

**What goes wrong:**
First time someone tests the generated server against Microsoft Entra, it fails: Entra sends `"op": "Add"` (capitalized) instead of `"op": "add"`, and stringifies booleans as `"True"`. ([Microsoft Learn: Known issues with SCIM 2.0 protocol compliance](https://learn.microsoft.com/en-us/entra/identity/app-provisioning/application-provisioning-config-problem-scim-compatibility)). The team adds case-insensitive matching for `op`, then a tolerant boolean parser, then accommodation for Entra's `members` Remove-via-`value`-array quirk, then Okta's extra-fields-on-User. By v1 ship, the spec layer is riddled with "if Entra send X" branches and the project's positioning ("strict spec only") is dead.

**Why it happens:**
- Real users will report compatibility issues from day one because Entra/Okta own the market.
- Each accommodation feels small and sympathetic ("just one if-statement").
- The boundary between "strict spec" and "real-world tolerance" has no natural enforcement point.

**Why it's catastrophic for THIS project:**
PROJECT.md explicitly lists IdP non-compliance accommodations as **out of scope** for v1, and CONCERNS.md §3 documents that legacy brittleness from ad-hoc IdP hacks was one of the three rewrite drivers. Repeating it produces the same brittle codebase under a new module path.

**How to avoid:**
- Documented v1 stance: the generated server implements the RFC. If an IdP sends non-compliant payloads, it gets a 400 with `scimType: invalidSyntax`. The README contains a table: "Tested with Entra: requires their compliance feature flag. Tested with Okta: requires X." Not "tolerant of."
- Plan v2's compatibility-profile design now (so engineers stop asking "where does this go?"), but ship nothing. Profiles will be a definition-level concept (e.g., `def.Compat(def.Entra2026)`) that produces an alternate generated parser path. No runtime middleware of accommodations in v1.
- PR template includes "Does this PR add IdP-specific tolerance?" — answer must be "no" or PR is closed.

**Warning signs:**
- Any commit message containing "Microsoft", "Entra", "Azure AD", or "Okta".
- A `compat/` or `tolerance/` package appearing in the generator-runtime library.
- Tests named `TestEntraQuirk_*` outside an explicitly-quarantined "future v2 design notes" location.

**Phase to address:**
Phase 0 (charter / scope statement). Every phase — gate via PR template.

---

### Pitfall 4: Pluggable SQL abstraction ahead of one working driver

**What goes wrong:**
Team starts by designing the SQL driver interface — connection management, dialect quoting, JSON column emulation, cursor pagination, transaction semantics — to "support Postgres and MySQL later." Six weeks in, the interface has 30 methods, none of them have been exercised against a real database, the SQLite implementation keeps rediscovering interface mismatches, and the migration DDL emitter is still unwritten because "we need to nail the abstraction first."

**Why it happens:**
- PROJECT.md says "pluggable SQL abstraction (driver interface)" and "Postgres/MySQL ship later, behind the same interface" — the natural reading is "design the abstraction first."
- Engineers prefer interface design over driver-specific debugging.
- "We don't want to refactor the interface later" — but you will refactor it later regardless, and refactoring against one working driver is much easier than designing in the abstract.

**How to avoid:**
- Build SQLite end-to-end first as a concrete (non-interface) implementation. CRUD, filter translation, sort, pagination, ETag-on-version, migration DDL. All passing the SCIM compliance test suite.
- Only then extract an interface — and the extraction is informed by the seams that actually existed in the working code.
- Postgres support comes via "implement the interface again, see what breaks, refactor the interface to fit both." Don't try to design for it preemptively.

**Warning signs:**
- A `db/driver.go` file with 20+ methods exists before any concrete driver works against `sqlite-utils-style` integration tests.
- Discussions about "how Postgres handles X" before SQLite handles X.
- The phrase "let's keep the interface clean for future drivers."

**Phase to address:**
Phase 3 (persistence) — the phase plan must explicitly say "concrete SQLite first, then extract."

---

### Pitfall 5: PATCH path-expression edge cases (RFC 7644 §3.5.2)

**What goes wrong:**
The generated server handles common PATCH cases (Add a top-level attribute, Replace a scalar) but breaks on the edge cases that real provisioning workflows hit constantly. Examples:

1. **Filter-targeted Remove on multi-valued complex:** `{"op":"remove","path":"members[value eq \"abc\"]"}` — must remove only the matching element, not all members. Vendors mishandle this both ways: some apply `value` as a payload-side filter and remove all members ([Microsoft Q&A](https://learn.microsoft.com/en-us/answers/questions/5524268/entra-scim-provisioning-sends-invalid-patch-reques)), some scope correctly only with the path filter.
2. **Add on a non-existent target:** §3.5.2.1 says "if the target location does not exist, the attribute and value are added." If `path` is `emails[type eq "work"].value` and no such email exists, must we synthesize the parent? Spec is ambiguous. Most libraries synthesize; document the choice.
3. **Replace treated as Add when target missing:** Common library behavior (`scim-patch` defaults to `treatMissingAsAdd: true`). Must be explicit and documented.
4. **Removing required attributes:** Spec says removal of a required attribute "SHALL fail" for top-level required, but is silent on required sub-attributes of optional complex attributes. Implementations differ.
5. **Multi-valued add via the `value: [...]` form:** Some clients send `{"op":"add","path":"phoneNumbers","value":[{"type":"work","value":"555"}]}` — must merge, not replace. Easy to get wrong.
6. **Capitalized `op`:** Spec is silent on case sensitivity of `op` values; Entra sends `"Add"`. Generated server must decide: case-sensitive (strict, will fail Entra) or case-insensitive. Per Pitfall 3, v1 = strict.

**Why it happens:**
- The spec itself is ambiguous in places (§3.5.2.1's "target location" definition, per the Microsoft Q&A discussion).
- Vendor PATCH payloads in the wild differ from spec, and there is no widely-accepted authoritative test corpus.
- It's easy to write tests that cover `op: replace, path: userName` and call PATCH "done."

**How to avoid:**
- Build a PATCH conformance test suite drawing from: RFC 7644 §3.5.2 examples, the [`scim-patch` test corpus](https://github.com/thomaspoignant/scim-patch), real captured payloads from public SCIM SDKs, and adversarial cases (filter matching nothing, filter matching all, removal of required sub-attribute of unset complex parent).
- Generated PATCH machinery should have a single tested implementation (in the runtime support library) parameterized by per-resource validation hooks — not regenerated PATCH logic per resource. PATCH logic is generic enough.
- Every documented edge-case decision (synthesize on add-missing? treat-missing-replace as add?) lives in a public table in the docs so behavior is predictable.

**Warning signs:**
- PATCH tests only exercise top-level scalar Replace.
- No tests with filter expressions in the path.
- "We'll handle that case when a user reports it."

**Phase to address:**
Phase 4 (PATCH semantics) — must include a test corpus phase exit criterion.

---

### Pitfall 6: Filter parser correctness — operator precedence, valuePath nesting, FILTER vs PATH grammars

**What goes wrong:**
The filter parser handles the happy path (`userName eq "alice"`) but produces wrong ASTs for compound expressions, fails on nested value-path filters, conflates the FILTER grammar with the PATCH PATH grammar, or panics on unbalanced parens. Real-world examples from the legacy go-scim and other parsers ([scim2/filter-parser DeepWiki analysis](https://deepwiki.com/scim2/filter-parser)):

- Precedence order per RFC 7644 §3.4.2.2: `not` > `and` > `or`. Parsers that left-fold `and`/`or` at the same precedence get wrong results on `a eq 1 or b eq 2 and c eq 3`.
- Parenthesization: `(a eq 1 or b eq 2) and c eq 3` requires correct grouping. The shunting-yard algorithm (or a recursive-descent parser with explicit precedence climbing) handles this; ad-hoc parsers don't.
- Value-path nesting: `emails[type eq "work" and (value ew "@x.com" or value ew "@y.com")]` — Errata 4690 restricts recursion levels but real filters use one level of nesting. Naive grammars hit left-recursion when allowing this.
- Trailing sub-attribute after value path: `emails[type eq "work"].value` — the segment after `]` is part of PATH grammar; parsers that treat the whole thing as a FILTER expression fail.
- Case sensitivity: attribute names and operators are case-insensitive (`UserName Eq "x"` == `username eq "x"`); attribute values are case-sensitive.
- Escape handling in string literals: `\"`, `\\`, surrogate pairs in JSON-style escapes. Easy to miss.

**Why it happens:**
- Filter parsing looks like a small task — engineers underestimate it and write recursive-descent parsers without a formal grammar.
- The RFC's BNF (§3.4.2.2 of RFC 7644) is non-trivial and parsers that don't follow it precisely diverge.
- The PATH grammar (PATCH paths) is a different production with similar tokens; reusing the FILTER parser is a tempting bug.

**How to avoid:**
- Use a generated parser (participle, pigeon PEG, ANTLR-Go) from a grammar that is committed and reviewable, not a hand-written recursive-descent parser.
- Property-based testing (`gopter` or `quick.Check`) generating well-formed filter expressions and asserting parse-then-print round-trip equivalence.
- Differential testing against another implementation (e.g., `scim2/filter-parser` Go package, or `scim2-filter-parser` Python) on a shared corpus.
- A fuzzing target (`go test -fuzz`) — Go's built-in fuzzer is well-suited; should run for hours in CI before declaring filter parsing done.
- Distinct types for `FilterExpr` and `PathExpr`. A parser entry point per grammar. Compile errors catch confusion.

**Warning signs:**
- Filter parser is a hand-written single function over 200 lines (the legacy filter.go was 1100 lines).
- No fuzz test target exists.
- Tests are all positive cases ("does this parse?") with no negative cases ("does this reject malformed input cleanly without panicking?").
- The parser is shared between PATCH paths and search filters with a "mode" boolean.

**Phase to address:**
Phase 5 (filter / sort) — entry criterion: grammar committed; exit criterion: 24 hours of fuzzing without crash + diff-test corpus passes.

---

### Pitfall 7: Mutability/Returned/Uniqueness rule violations

**What goes wrong:**
The generated server returns password hashes in GET responses ("just for that one debug case"), allows PATCH to modify `id` or `userName` after creation, fails to enforce server-scoped uniqueness on `userName` until it gets a database constraint error and surfaces the raw error, returns `meta.location` as readWrite, or omits `id` (returned: always) when the client requests `?attributes=userName`.

**Why it happens:**
- These rules are per-attribute properties on the schema (RFC 7643 §2.2) and are easy to ignore at the data-flow level if the generator emits naive marshalling.
- `mutability` has four values (`readOnly`, `readWrite`, `immutable`, `writeOnly`) and the implications for each operation (POST/PUT/PATCH) are non-obvious. `immutable` famously means "set once, then no changes" — not "never settable" ([RFC 7644 §3.5.2](https://datatracker.ietf.org/doc/html/rfc7644)).
- `returned` has four values (`always`, `never`, `default`, `request`) and applies even when the client uses the `attributes` query parameter — `returned: never` attributes must be filtered after attribute selection, not before.
- `uniqueness` enforcement requires a server-side check before insert, not just a database unique constraint (the constraint is fallback; the SCIM error must be `409` with `scimType: uniqueness`).

**How to avoid:**
- Generate per-attribute filter functions: `applyMutabilityForPATCH(*User, op, path) error`, `filterForReturned(*User, requestedAttrs, excludedAttrs) *User`. The rules are compile-time evaluated against the definition.
- Test matrix: for each attribute, test (POST, PUT, PATCH) × (server-set, client-set, omitted). The matrix is large but mechanical and can itself be code-generated.
- Uniqueness: pre-flight check inside the same transaction as the insert; on conflict, translate to SCIM error before returning to the user.
- `password`-class attributes (`mutability: writeOnly`, `returned: never`): generated marshaller must not emit them, even via reflection on a debug code path. Add a test that inspects the marshalled JSON for any `password` substring.

**Warning signs:**
- A test passes that does `POST /Users` and asserts `response.Password == ""` — but the server never set it in the response struct, so the test passes trivially. Instead, assert the raw JSON does not contain "password".
- "We'll add the uniqueness check later" — leaks DB errors to the wire.
- No test for `attributes=userName` filtering when `id` is `returned: always` (must still appear).

**Phase to address:**
Phase 2 (core generator: emit mutability/returned filters) and Phase 6 (compliance test suite).

---

### Pitfall 8: ETag/version mismanagement (RFC 7644 §3.14, RFC 7232)

**What goes wrong:**
- ETag is generated from a timestamp with second-level resolution; two updates in the same second produce the same ETag and the second goes undetected.
- ETag is exposed as a non-opaque value (e.g., the raw `updated_at` timestamp), and clients start parsing it.
- `meta.version` and the HTTP `ETag` header drift — RFC 7643 requires they match exactly, including the `W/` prefix (case-sensitive).
- ETag is set on the response but `If-Match` is not validated on PATCH/PUT/DELETE — silent loss of optimistic concurrency.
- Bulk PATCH operations that mutate multiple sub-attributes produce one ETag transition, but the server emits the new ETag from a stale cached resource.
- Strong vs. weak: a database-backed implementation cannot generally satisfy strong-validator semantics (byte-identical representations); per RFC 7643 the server MUST mark such ETags as weak with the `W/` prefix.

**Why it happens:**
- Versioning is "optional" in the spec (§3.14 of RFC 7644), so it's easy to ship without it and add later — at which point clients have already learned to ignore conflicts.
- The relationship between `ETag` header, `meta.version`, `If-Match`, and `If-None-Match` is spread across RFC 7644 §3.14 and RFC 7232 §2.1.
- "Just hash the JSON" works until the server's JSON serialization is non-deterministic (key ordering, optional whitespace).

**How to avoid:**
- Pick a deterministic version source: a per-resource monotonic counter incremented in the same transaction as the write. Hash it with the resource ID for opacity.
- Mark all ETags weak (`W/"..."`) — never claim strong-validator semantics; the spec accepts weak.
- Centralize ETag generation in the generated repository methods so the same value is used in the response header and `meta.version` on a single read-after-write pull.
- `If-Match` validation is a single shared middleware/helper, not duplicated per handler. Returns 412 Precondition Failed.
- Test: write resource, capture ETag, modify outside the API, attempt PATCH with old ETag, assert 412. And: two concurrent PATCHes with the same `If-Match`; exactly one succeeds.

**Warning signs:**
- ETag string contains a recognizable timestamp or integer.
- No 412 test exists.
- `meta.version` is computed in the JSON serializer and `ETag` header is computed in the handler.

**Phase to address:**
Phase 3 (persistence — version columns) and Phase 4 (PATCH/PUT — If-Match middleware).

---

## Moderate Pitfalls

### Pitfall 9: Generated code that doesn't `gofmt`

**What goes wrong:**
`text/template` is text-based; a stray space, a missing line, or an unbalanced brace from a template-control structure produces invalid Go that `gofmt` rejects. Worse, sometimes it produces *valid* but ugly Go that diffs noisily on every regeneration.

**How to avoid:**
- Always render to a `bytes.Buffer`, then call `go/format.Source()` from the standard library before writing to disk. `format.Source` returns an error pointing to the offending line, which is invaluable for template debugging.
- Use `goimports` on the formatted output to auto-manage imports (so the template doesn't have to enumerate them perfectly).
- CI step: re-run the generator and `git diff --exit-code` — catches any regeneration churn.
- Use `text/template`, not `html/template` (the latter HTML-escapes identifiers).
- Mark every generated file with `// Code generated by scimgen. DO NOT EDIT.` as the first non-blank line — Go convention, recognized by gopls and others ([go.dev — go fmt your code](https://go.dev/blog/gofmt)).

**Phase to address:** Phase 1 (generator scaffolding).

---

### Pitfall 10: Regeneration overwrites user customization

**What goes wrong:**
User customizes a generated handler to add an audit log call. Regenerates. Customization vanishes silently.

**How to avoid:**
- The generated module is meant to be checked in by the user and edited at the seams the generator deliberately exposes. Two strategies, used together:
  1. **Composition over modification:** generated types are concrete structs; users wrap them via embedding or via interface-implementing wrappers in *non-generated* sibling files.
  2. **Hooks, not protected regions:** the generator emits clearly-named extension points (e.g., `func (c *Config) BeforeUserCreate func(ctx, *User) error`) rather than `// USER CODE BEGIN/END` comment markers. Comment-marker preservation is fragile and adds parser complexity to the generator.
- The CLI emits a `Makefile` (or just-recipe) with `regen` and `regen-clean` targets; `regen` refuses to overwrite files lacking the `DO NOT EDIT` marker (i.e., user files).
- README clearly names which files are regenerated and which are user-owned (`*_gen.go` regenerated, everything else preserved).

**Phase to address:** Phase 1 (generator scaffolding — file naming convention) and Phase 7 (CLI tool — regen workflow).

---

### Pitfall 11: Stale test fixtures when the definition changes

**What goes wrong:**
User changes the definition (adds a new attribute to User). Regenerates. Generated tests are updated, but the user's hand-written integration tests in `users_test.go` still construct `User{}` literals missing the new field, or assert on JSON shapes that no longer match. Compile errors at best, silent skip at worst.

**How to avoid:**
- Generated test fixtures are in `*_fixtures_gen.go` files; user tests should call fixture builders (`NewUserFixture()`) rather than struct-literal-construct `User`. Builders absorb schema changes.
- Field additions are backward-compatible in struct-literals only when the literal uses field names. Don't allow positional struct literals in generated docs/examples.
- Provide a `scimgen check` subcommand that diffs the current definition against the last-generated state and warns about removals/renames (additive changes are safe, removals/renames break user code).

**Phase to address:** Phase 7 (CLI — `check` subcommand) and Phase 6 (test infrastructure).

---

### Pitfall 12: Definition validation runs too late

**What goes wrong:**
A user writes a definition with a typo: `mutability: "immuteable"`. The parser accepts it (it's just a string). The generator runs. The generated code references `MutabilityImmuteable` which doesn't exist. The user sees a compile error in machine-generated code 3000 lines deep and has no idea where their typo is.

**How to avoid:**
- The definition parser is the single point of validation. It validates *everything* the generator might rely on — enum values, attribute name uniqueness within a schema, sub-attribute restrictions (no complex sub-attributes per RFC 7643), no two attributes with the same canonical-case name, schema URN format, etc.
- Validation errors point to the user's definition source location (line/column for YAML, struct-tag location for builder DSL via `runtime.Caller` if necessary), never to generated code.
- Snapshot test the validator: a corpus of intentionally-broken definitions, asserting on the human-readable error each produces.
- Generator panics if it receives an invalid model. The validator is the only legitimate error source.

**Phase to address:** Phase 1 (parser/validator — must precede the generator).

---

### Pitfall 13: SCIM listing semantics — startIndex 1-based, totalResults reflects filter, integers not strings

**What goes wrong:**
- Implementing `startIndex` as 0-based (every other API in the world is 0-based; SCIM is 1-based per RFC 7644 §3.4.2.4).
- `totalResults` returning the table count instead of the filtered count.
- Returning `"totalResults": "100"` (string) instead of `100` (integer) — the spec requires integer per §3.4.2.4 ([RFC 7644](https://datatracker.ietf.org/doc/html/rfc7644)).
- No default for `count` — RFC 7644 doesn't mandate a specific default but most implementations cap at 100-200; unbounded responses are a DoS vector.
- Returning `Resources: null` instead of `Resources: []` when no results.
- Forgetting `schemas: ["urn:ietf:params:scim:api:messages:2.0:ListResponse"]` on the wrapper.

**How to avoid:**
- Generated list-response builder enforces correct types and field names; user code never constructs the wrapper.
- Compliance tests for: `startIndex < 1` (treat as 1 per practice), `count > server max` (clamp), `count = 0` (return only totalResults, no Resources), filter that matches nothing.
- Decide a default `count` (recommend 100; many vendor SDKs use it) and document.

**Consider RFC 9865 (cursor pagination, October 2025):** new RFC defines `cursor`/`nextCursor` query/response attributes and `pagination` capability in `/ServiceProviderConfig`. v1 should support index-based pagination (RFC 7644) but design data layer so cursor pagination is a Phase-N additive change. Don't paint into a corner that requires materializing the full result set ([RFC 9865](https://datatracker.ietf.org/doc/rfc9865/)).

**Phase to address:** Phase 5 (query/list endpoints).

---

### Pitfall 14: Wrong content type and error format

**What goes wrong:**
- Returning `Content-Type: application/json` instead of `application/scim+json` — RFC 7644 §3.1 requires the latter. Some clients reject the response.
- Error responses return Go's default JSON error envelope (`{"error": "..."}`) instead of the SCIM error format with `schemas: ["urn:ietf:params:scim:api:messages:2.0:Error"]`, `status: "409"` (string!), and `scimType` for 400-class errors.
- `status` field as integer (`409`) instead of string (`"409"`) — spec requires string per §3.12 of RFC 7644.

**How to avoid:**
- Single error-writing helper in the runtime support library, used by all generated handlers. Never construct the error envelope ad-hoc.
- Content-Type is set at the handler-write helper level, not per-handler.
- Compliance tests assert content-type header and error envelope shape on every error path.

**Phase to address:** Phase 2 (handler scaffolding).

---

### Pitfall 15: Group membership round-trip — `User.groups` is read-only mirror

**What goes wrong:**
Server stores `groups` on User. PATCH on `/Users/.../groups` is accepted, mutating the user. Group's `members` list is now out of sync with User's `groups` list. Per RFC 7643 §4.1.2: `User.groups` is `mutability: readOnly` — the canonical store is `Group.members`; `User.groups` is a derived projection.

**Why it happens:**
- It's symmetric data; not obvious which side is canonical.
- Storing the join in both directions feels efficient for reads.

**How to avoid:**
- Generated User type does not have a writable `groups` field at the storage layer. The field is computed on read (LEFT JOIN to group_members).
- PATCH/PUT operations on `User` validate that the request does not target `groups` (mutability: readOnly returns 400).
- Compliance test: PATCH a user with a `groups` operation; assert 400 with `scimType: mutability`.

**Phase to address:** Phase 2 (User/Group resource generation) and Phase 6 (compliance suite).

---

### Pitfall 16: Resource-type discovery diverges from what's served

**What goes wrong:**
`/ResourceTypes` advertises a `User` resource type with a particular schema URN. The actual server has been customized and serves a slightly different shape. Clients use `/ResourceTypes` to drive their behavior and get inconsistent results.

**How to avoid:**
- `/Schemas` and `/ResourceTypes` content is generated from the same definition the type code is generated from. Single source of truth, embedded as `embed.FS` in the generated server.
- Compliance test: parse the `/Schemas` response, then for each defined attribute, POST a resource with that attribute and assert it round-trips.

**Phase to address:** Phase 2 (discovery endpoints generated alongside resource handlers).

---

## Minor Pitfalls

### Pitfall 17: SQLite driver choice — modernc.org/sqlite vs mattn/go-sqlite3

**What goes wrong:**
Choosing `mattn/go-sqlite3` because "it's the canonical one" forces CGO on every consumer of the generated code. Cross-compilation breaks. CI gets harder. Or, choosing `modernc.org/sqlite` for the CGO-free win, then hitting a perf cliff on large dataset scans (mattn ~3x faster on 200K-row queries per [cvilsmeier/go-sqlite-bench, March 2026](https://github.com/cvilsmeier/go-sqlite-bench)).

**Recommendation:**
- Default to `modernc.org/sqlite` for the v1 reference driver. PROJECT.md says "correctness over throughput in v1." The CGO-free property matters more than the perf gap at v1 scale.
- Note that as of `modernc.org/sqlite` v1.46.1 (Dec 2025) the prepared-statement perf gap closed significantly; that fix is now widely deployed.
- Datetime handling: both drivers accept `time.Time` but parse stored ISO 8601 strings differently on read; pick a single canonical text storage format (ISO 8601 UTC) and write a normalization helper. Test it.

**Phase to address:** Phase 3 (persistence — driver choice).

---

### Pitfall 18: Multi-module workspace footguns (`go.work`, `replace`)

**What goes wrong:**
- `replace` directives in `go.mod` (instead of `go.work`) leak to consumers — they try to `go get` your module and Go fails because the local path doesn't exist on their machine.
- `go.work` committed accidentally enables workspace mode for everyone cloning the repo, masking module-resolution bugs that production builds (which use `GOWORK=off`) hit.
- Stale `go.sum` after dependency changes — `go work sync` not run.
- A `use` and a local-path `replace` for the same module produces the cryptic error "workspace module is replaced at all versions in the go.work file" ([golang/go#54264](https://github.com/golang/go/issues/54264)).

**How to avoid:**
- `replace` lives in `go.work` only. `go.mod` files in published modules contain only real version constraints.
- CI runs both `go build ./...` (workspace mode) AND `GOWORK=off go build ./...` (production-like) — disagreement is a bug.
- `.gitignore` `go.work.sum` (it's per-developer-machine state).
- Decide once whether to commit `go.work`. For a multi-module project where the workspace topology is the same for every contributor, commit it; otherwise `.gitignore` it.

**Phase to address:** Phase 0 (repo setup) and Phase 1 (generator scaffolding — module layout).

---

### Pitfall 19: Migration ordering and idempotency

**What goes wrong:**
Generator emits `001_create_users.sql`, `002_add_emails.sql`. User regenerates after editing definition. Generator emits `001_create_users.sql` (now with the `emails` column included). Migration tooling sees the same filename with different content, behaviors range from "silently skips" to "fails checksum." Or, generated migrations include `DROP TABLE IF EXISTS` and a dev environment loses data.

**How to avoid:**
- v1 scope per PROJECT.md: "generator emits fresh DDL each run; user owns diffing/migration." Document this clearly. Generator emits `schema.sql` (single file, full DDL, idempotent via `CREATE TABLE IF NOT EXISTS`), not numbered migrations. User integrates with their own migration tool.
- Make this a documented constraint, not a feature — say in the README: "We do not generate incremental migrations. If you need them, use [tool] against successive schema.sql snapshots."
- The single `schema.sql` is deterministic — same definition produces identical output. Test it.

**Phase to address:** Phase 3 (DDL emission) and Phase 7 (CLI/docs).

---

### Pitfall 20: Embedded vs. runtime-loaded schema files

**What goes wrong:**
Generator emits schema JSON files in `schemas/` directory and the generated server reads them via `os.ReadFile` at startup. User deploys without the schemas directory. Server panics on first `/Schemas` request. Or, dev edits a schema file by hand to "tweak something quickly" and forgets to regenerate the types — silent drift between served schema and actual behavior.

**How to avoid:**
- Generated server uses `//go:embed schemas/*.json` so the binary is self-contained.
- The embedded schemas are the same artifact the generator used to produce types. If the user wants to inspect them, they can extract them from the binary or read the `*_gen.go`.
- No `os.ReadFile` for schema content at runtime, ever.

**Phase to address:** Phase 2 (discovery endpoint generation).

---

### Pitfall 21: Loss of type information at generation boundaries

**What goes wrong:**
The generator passes data around as `map[string]interface{}` between stages, then serializes to template strings. Type errors that should be caught at generator-build time instead surface at generated-code-build time, with stack traces that point into templates and `interface{}` assertions.

**How to avoid:**
- The internal model that the parser produces and the generator consumes is a typed tree of Go structs (no `map[string]any`). Adding a new attribute kind requires touching the type, which forces awareness of every codegen path.
- Templates receive concrete typed structs as their `.` data. Template field references then fail at template-parse time (good) rather than at template-execute time (less good).

**Phase to address:** Phase 1 (model design).

---

## Technical Debt Patterns

| Shortcut | Immediate Benefit | Long-term Cost | When Acceptable |
|----------|-------------------|----------------|-----------------|
| Skip ETag/If-Match in v1 | One less moving part for first integration | Clients build up workflows assuming no concurrency; retrofit breaks them | Never — versioning is too central to PUT/PATCH semantics |
| Hand-write filter parser instead of using a parser generator | No external dep | Fuzz crashes, precedence bugs, eventual rewrite | Never for SCIM — the grammar is published and non-trivial |
| Use `map[string]interface{}` in the generator's internal model | Looks flexible, fast to start | Loses every static check Go gives you; bugs surface in generated code | Never — the whole point of code-gen is type safety |
| Add IdP-specific tolerance "just for now" | One vendor works in test | Pitfall 3 — the file becomes unreviewable | Never in v1; design profiles for v2 |
| Ship without `/Schemas` and `/ResourceTypes` | Faster to first POST /Users | Clients have to be hardcoded to your shape; not really a SCIM server | Only as scaffolding within Phase 2; must be done by Phase 6 |
| Generate handlers per resource but PATCH logic per resource too | Symmetry feels right | PATCH is the trickiest single piece; six copies = six places to fix every bug | Never — PATCH is generic over resource type with hooks |
| Skip mutability/returned filtering ("clients can ignore fields they don't need") | Simpler marshaller | Returns passwords; immutable fields silently mutate; spec violation | Never — these are spec MUSTs |
| Generate Postgres support before SQLite passes compliance | "Doing it right from the start" | Pitfall 4 — abstraction without reality | Never — one driver passing compliance first |
| Use `text/template` without `format.Source` post-processing | "Templates are simpler if I just emit valid code" | Diff noise; broken builds when templates have edge cases | Never — always run through `format.Source` |
| Skip integration tests against real SQLite ("unit tests cover it") | Faster feedback loop | SQL dialect bugs, driver quirks, transaction semantics — all undetected | Only if a separate compliance test phase covers them |

---

## Integration Gotchas

| Integration | Common Mistake | Correct Approach |
|-------------|----------------|------------------|
| Microsoft Entra ID | Adding case-insensitive `op` matching to "fix" Entra | Strict per spec; document that Entra requires its compliance feature flag (per [Microsoft Learn](https://learn.microsoft.com/en-us/entra/identity/app-provisioning/application-provisioning-config-problem-scim-compatibility)) |
| Okta | Accepting Okta-specific extra fields on User | Strict; reject unknown fields with `scimType: invalidSyntax` (or ignore unknown fields with a documented warning — pick one and document) |
| Any IdP for `members` PATCH | Treating `value: [...]` array on Remove as a filter (deletes everything) | Reject — Remove must use a filter in `path`, not a value-array filter |
| `net/http` mux | Assuming the user's mux preserves query strings, body length, etc. | Generated handler accepts a standard `http.ResponseWriter`/`*http.Request` and reads only via the standard interface |
| Database migration tools | Trying to be clever about generating numbered migrations | Don't — emit one `schema.sql`; user owns migration |
| Reverse proxies | Stripping `If-Match` or `ETag` headers | Document that proxy must preserve conditional-request headers |
| Logging middleware | Logging raw request bodies (passwords) | Generated server's request-logging extension point gets a redacted body where `password` and other `returned: never` attributes are scrubbed |

---

## Performance Traps

PROJECT.md explicitly says "Explicit performance/scale targets — out of scope for v1; correctness over throughput." So this table is short and only flags traps that cause *correctness* problems too, or that are cheap to avoid.

| Trap | Symptoms | Prevention | When It Breaks |
|------|----------|------------|----------------|
| Cursor-translation in pagination (materializing full result set to serve index pages) | Memory blowup on large filters | Use SQL `LIMIT/OFFSET` directly; design for RFC 9865 cursor pagination as a future addition | At ~10k+ matched resources |
| `SELECT *` on User with denormalized JSON columns | Slow queries even for `attributes=userName` | Project columns based on `attributes` query param; only fetch what's requested | At ~1k resources with large JSON payloads |
| Filter compiled per-request | CPU spent re-parsing same filter | LRU cache of compiled filter ASTs by string key | At sustained query load (low-priority for v1) |
| No SQL index on filter-able attributes | Sequential scans for `filter=userName eq "x"` | Generator emits indexes for attributes marked `uniqueness != none` and for explicitly-tagged "indexed" attributes | At ~10k+ resources |
| ETag computed by re-marshalling the resource | Expensive per write | Use a per-resource version counter incremented in the same transaction | Always — and it's a correctness issue (non-determinism) |

---

## Security Mistakes

| Mistake | Risk | Prevention |
|---------|------|------------|
| Echoing client-supplied `id` on POST | Clients spoof IDs; collision attacks | Per RFC 7643, `id` is server-generated; ignore client-supplied value silently or reject |
| Returning `password` (or any `returned: never` attribute) in any response | Credential disclosure | Generated marshaller never emits these; integration test inspects raw JSON for forbidden substrings |
| SCIM filter injection — passing user-supplied filter into raw SQL | SQL injection | Filter compiler emits parameterized SQL only; never string-concatenates user input |
| Verbose error messages on bcrypt failure leaking attribute name | Confirms password attribute exists | Generic error message for `writeOnly` attribute failures |
| No request body size limit | Memory exhaustion via huge JSON | Generated server has a configurable body-size limit (default low, e.g., 1 MB) |
| No JSON nesting depth limit | Stack overflow on deeply-nested JSON | Use a depth-limited JSON decoder, or pre-validate depth |
| `If-Match` weak ETag confusion | A client trusting weak ETag for security checks may proceed past unintended changes | Document weak-ETag semantics; never use ETag for authorization decisions |
| Bulk endpoint allowing unbounded operations | DoS via huge bulk payload | Enforce `bulkMaxOperations` and `bulkMaxPayloadSize` from `/ServiceProviderConfig` |
| Auth out-of-scope but no explicit `Forbidden`-by-default | Server runs unauthenticated by default if user forgets to wire auth | README explicitly warns; perhaps emit a stub middleware that returns 401 unless explicitly disabled |
| Logging full request body | Credentials, PII in logs | Generated logging hook gets a sanitized body with `returned: never` fields scrubbed |

---

## "Looks Done But Isn't" Checklist

- [ ] **PATCH endpoint:** Often missing filter-targeted Remove on multi-valued — verify with `{"op":"remove","path":"emails[type eq \"work\"]"}` against a user with multiple emails of different types
- [ ] **PATCH endpoint:** Often missing the `value: [...]` add form for multi-valued — verify with `{"op":"add","path":"phoneNumbers","value":[{"type":"work","value":"555"}]}`
- [ ] **PATCH endpoint:** Often missing rejection of mutations to `mutability: readOnly` — verify with PATCH on `meta.created` returning 400 with `scimType: mutability`
- [ ] **PATCH endpoint:** Often missing rejection of second-write to `mutability: immutable` — verify with two consecutive Replace ops on an immutable field
- [ ] **Filter parsing:** Often missing operator precedence — verify `not a eq 1 and b eq 2` parses as `(not (a eq 1)) and (b eq 2)`, not `not ((a eq 1) and (b eq 2))`
- [ ] **Filter parsing:** Often missing case-insensitive attribute names — verify `UserName eq "x"` matches `userName eq "x"`
- [ ] **List response:** Often missing 1-based `startIndex` — verify GET with `?startIndex=2` returns the second resource (not third)
- [ ] **List response:** Often missing integer types in JSON — verify `totalResults`, `startIndex`, `itemsPerPage` are JSON numbers, not strings
- [ ] **Error response:** Often missing `application/scim+json` content type — verify on every 4xx and 5xx response
- [ ] **Error response:** Often has integer `status` — verify it is a JSON string ("409", not 409)
- [ ] **ETag:** Often missing on response headers and `meta.version` — verify both present and equal on POST/GET/PUT/PATCH
- [ ] **ETag:** Often missing `If-Match` enforcement — verify PATCH with stale ETag returns 412
- [ ] **`/Schemas`:** Often inconsistent with served resources — verify that every attribute in `/Schemas/urn:ietf:params:scim:schemas:core:2.0:User` is acceptable on POST /Users
- [ ] **`/ResourceTypes`:** Often missing the schema extensions array — verify EnterpriseUser appears as a `schemaExtension` of User
- [ ] **`/ServiceProviderConfig`:** Often inaccurate — verify advertised capabilities (`patch.supported`, `bulk.supported`, `filter.supported`, `etag.supported`) match actual behavior
- [ ] **Password attribute:** Often returned in some path — verify the substring "password" never appears in any GET/PUT/PATCH/POST response body
- [ ] **`User.groups`:** Often writable — verify PATCH on `groups` returns 400; verify membership read derives from Group.members
- [ ] **Generated code:** Often diffs noisily on regen — verify `make regen && git diff --exit-code` passes
- [ ] **Generated code:** Often missing `// Code generated ... DO NOT EDIT.` header — verify every `*_gen.go` has it on line 1
- [ ] **Migration:** Often non-deterministic — verify two regen runs produce byte-identical `schema.sql`
- [ ] **Compliance suite:** Often run only against happy paths — verify a known-bad-input corpus exists (e.g., malformed filters, oversized bodies, unknown attributes)

---

## Recovery Strategies

| Pitfall | Recovery Cost | Recovery Steps |
|---------|---------------|----------------|
| Reintroduced generic tree (Pitfall 1) | HIGH | Stop adding features; revert to the commit before `Property` interface; rewrite affected services to use generated types directly |
| Runtime schema interpretation (Pitfall 2) | MEDIUM | Move schema content into `embed.FS`; remove `os.ReadFile`; smoke test; promote registry to read-only |
| IdP accommodations crept in (Pitfall 3) | MEDIUM | Identify all "if Entra/Okta then" branches; remove; document removal as breaking; design v2 profile system |
| Pluggable SQL too early (Pitfall 4) | MEDIUM | Inline the abstraction back into the SQLite driver; finish features against concrete driver; re-extract interface from working code |
| PATCH edge cases break in production (Pitfall 5) | LOW per case | Add failing test; fix; add to compliance corpus permanently |
| Filter parser bug found (Pitfall 6) | LOW per case | Add failing test; fix grammar (not the parser); regenerate parser; ensure fuzz target covers the case |
| Mutability violation shipped (Pitfall 7) | LOW–MEDIUM | Hot-fix the marshaller; audit all `returned: never` attributes; add the audit to CI as a permanent test |
| ETag silently broken (Pitfall 8) | MEDIUM | Audit version-source determinism; switch to per-resource counter if currently timestamp-based; add 412 tests |
| Generated code stops formatting (Pitfall 9) | LOW | Wrap template output in `format.Source`; CI step `make regen && go vet ./... && git diff --exit-code` |
| Migration files conflict | HIGH | Switch to single-file deterministic `schema.sql` model; users own diffing |

---

## Pitfall-to-Phase Mapping

Phase names below are illustrative for the roadmap consumer. Adjust to the roadmap's actual phase nomenclature.

| Pitfall | Prevention Phase | Verification |
|---------|------------------|--------------|
| 1. Generic tree relapse | Phase 0 (architecture spike, written rule) | Code review checklist; periodic architecture audit |
| 2. Runtime schema interpretation | Phase 0 + Phase 2 (generator emits embedded schemas) | grep for `os.ReadFile.*schema` returns nothing |
| 3. IdP scope creep | Phase 0 (charter); every phase (PR template) | Search commit history for vendor names |
| 4. SQL abstraction too early | Phase 3 (persistence — concrete-first plan) | Single working driver passes compliance before interface extraction |
| 5. PATCH edge cases | Phase 4 (PATCH semantics) | Compliance corpus, including filter-targeted Remove and value-array Add |
| 6. Filter parser correctness | Phase 5 (filter / sort) | Committed grammar; fuzz target ran 24h+; diff-test against another impl |
| 7. Mutability/returned/uniqueness | Phase 2 (generator) + Phase 6 (compliance) | Per-attribute test matrix; raw-JSON inspection for forbidden substrings |
| 8. ETag mismanagement | Phase 3 (version columns) + Phase 4 (If-Match) | 412-on-stale test; deterministic version-source test |
| 9. Generated code doesn't gofmt | Phase 1 (generator scaffolding) | CI: `make regen && git diff --exit-code` |
| 10. Regeneration overwrites customization | Phase 1 (file naming) + Phase 7 (CLI workflow) | Test: customize a non-`_gen.go` file, regen, customization preserved |
| 11. Stale test fixtures | Phase 6 (test infrastructure) + Phase 7 (`scimgen check`) | `scimgen check` warns on definition diff |
| 12. Late definition validation | Phase 1 (parser/validator before generator) | Snapshot test corpus of broken definitions with expected errors |
| 13. List semantics wrong | Phase 5 (query/list) | Compliance suite covers startIndex=0/1, count=0, integer types |
| 14. Wrong content type / error format | Phase 2 (handler scaffolding) | Single error helper; compliance assertions on every error path |
| 15. `User.groups` writable | Phase 2 (User/Group gen) + Phase 6 (compliance) | PATCH `groups` returns 400 |
| 16. Discovery diverges from served | Phase 2 (discovery generated alongside resources) | Round-trip test: every advertised attribute is acceptable |
| 17. SQLite driver choice | Phase 3 (persistence) | Document choice; benchmark recorded |
| 18. Multi-module workspace | Phase 0 (repo setup) | CI builds with `GOWORK=on` and `GOWORK=off` |
| 19. Migrations non-deterministic | Phase 3 (DDL emission) | Two regens produce byte-identical schema.sql |
| 20. Runtime-loaded schemas | Phase 2 (embed.FS) | grep for `os.ReadFile.*schema` returns nothing |
| 21. Untyped generator internals | Phase 1 (model design) | Internal model is structs, not maps |

---

## Explicit Coverage of the Three Documented Rewrite Drivers

From `.planning/codebase/CONCERNS.md`:

### Driver 1: Feasibility (generic tree → SQL incompatible)
**How v1 avoids repeating it:** Pitfall 1 (no generic tree), Pitfall 2 (no runtime schema interpretation), Pitfall 4 (concrete SQLite first). The data model is generated Go structs with explicit fields; SQL DDL is emitted directly from the same definition; filter expressions compile to parameterized SQL via per-resource generated functions. There is no tree-to-relational translation step because there is no tree.

### Driver 2: Efficiency (in-memory tree duplication, brittle SCIM rule enforcement)
**How v1 avoids repeating it:** Pitfall 1 (no Property tree means no Raw/Hash/visit traversal cost), Pitfall 7 (mutability/returned/uniqueness rules emitted as compile-time-evaluated functions, not runtime subscribers). SCIM invariants like "exactly one primary in a multi-valued complex" are enforced by generated validator functions that have direct access to the typed slice, not by a subscriber walking a tree of properties.

### Driver 3: Portability (no IdP-quirk accommodation, brittle hacks)
**How v1 avoids repeating it:** Pitfall 3 (strict spec only in v1, by charter). The brittleness in legacy came from accommodations bolted on without a coherent extension model. v1 ships none. v2 will introduce a definition-level "compatibility profile" concept that produces alternative generator output paths — the right place to handle this. v1 must resist piecemeal accommodation; the design exists, but no implementation lands until v2.

---

## Sources

### SCIM Specifications
- [RFC 7644 — System for Cross-domain Identity Management: Protocol](https://datatracker.ietf.org/doc/html/rfc7644) — PATCH (§3.5.2), filtering (§3.4.2.2), pagination (§3.4.2.4), error responses (§3.12), versioning (§3.14)
- [RFC 7643 — System for Cross-domain Identity Management: Core Schema](https://www.rfc-editor.org/rfc/rfc7643.html) — attribute characteristics (§2.2), User schema, Group schema, EnterpriseUser
- [RFC 9865 — Cursor-Based Pagination of SCIM Resources (October 2025)](https://datatracker.ietf.org/doc/rfc9865/) — recent and material for pagination design
- [RFC 7232 — HTTP Conditional Requests](https://datatracker.ietf.org/doc/html/rfc7232) — strong vs weak validators, If-Match semantics

### SCIM Implementation Pitfalls
- [Microsoft Learn — Known issues with SCIM 2.0 protocol compliance (Entra ID)](https://learn.microsoft.com/en-us/entra/identity/app-provisioning/application-provisioning-config-problem-scim-compatibility)
- [Microsoft Q&A — Entra SCIM Provisioning sends invalid PATCH requests](https://learn.microsoft.com/en-us/answers/questions/5524268/entra-scim-provisioning-sends-invalid-patch-reques)
- [Microsoft Q&A — SCIM PATCH of Complex Multi-Valued Attribute Includes Filter and Sub Attribute in Path](https://learn.microsoft.com/en-us/answers/questions/708183/scim-patch-of-complex-multi-valued-attribute-inclu)
- [thomaspoignant/scim-patch — JS PATCH library with extensive edge-case handling](https://github.com/thomaspoignant/scim-patch)
- [scim2/filter-parser — Go SCIM filter parser; documented limitations](https://github.com/scim2/filter-parser) and [DeepWiki analysis](https://deepwiki.com/scim2/filter-parser)
- [Okta SCIM 2.0 Developer Guide](https://developer.okta.com/docs/api/openapi/okta-scim/guides/scim-20)
- [WorkOS — SCIM 2.0 vs 1.0 differences](https://workos.com/guide/scim2-vs-scim1) — mutability/returned/uniqueness explanation
- [Scalekit blog — How to build a SCIM 2.0 endpoint](https://www.scalekit.com/blog/build-scim-endpoint)

### Go Code Generation
- [go.dev — go fmt your code](https://go.dev/blog/gofmt) — formatter conventions
- [pkg.go.dev — text/template](https://pkg.go.dev/text/template) — template engine
- [pkg.go.dev — cmd/gofmt](https://pkg.go.dev/cmd/gofmt) — formatter and `-s` flag behavior
- [pkg.go.dev — github.com/dolmen-go/codegen](https://pkg.go.dev/github.com/dolmen-go/codegen) — text/template + gofmt + DO NOT EDIT marker pattern
- [Anti-patterns in code generation (Abstratt blog)](https://blog.abstratt.com/2011/04/05/anti-patterns-in-code-generation-part-1/) — preserved-region and customization patterns
- [golang/go#49555 — gopls behavior on generated files](https://github.com/golang/go/issues/49555)

### Go Multi-Module Workspaces
- [go.dev — Tutorial: Getting started with multi-module workspaces](https://go.dev/doc/tutorial/workspaces)
- [go.dev — Get familiar with workspaces](https://go.dev/blog/get-familiar-with-workspaces)
- [golang/go#54264 — `replace` + `use` conflict in go.work](https://github.com/golang/go/issues/54264)

### SQLite Drivers
- [cvilsmeier/go-sqlite-bench — March 2026 driver benchmarks](https://github.com/cvilsmeier/go-sqlite-bench)
- [multiprocessio/sqlite-cgo-no-cgo — DataStation benchmark](https://github.com/multiprocessio/sqlite-cgo-no-cgo)
- [DataStation blog — SQLite in Go, with and without cgo](https://datastation.multiprocess.io/blog/2022-05-12-sqlite-in-go-with-and-without-cgo.html)
- [lbe/sqlite-read-benchmark — modernc prepared-statement fix verification](https://github.com/lbe/sqlite-read-benchmark)

### Project-Internal Sources
- `/Users/imulab/workspace/go-scim/.planning/PROJECT.md` — scope and out-of-scope statements
- `/Users/imulab/workspace/go-scim/.planning/codebase/CONCERNS.md` — three documented rewrite drivers
- `/Users/imulab/workspace/go-scim/.planning/codebase/ARCHITECTURE.md` — legacy architecture (anti-pattern reference)

---
*Pitfalls research for: Go SCIM v2 server code-generation toolkit*
*Researched: 2026-05-07*

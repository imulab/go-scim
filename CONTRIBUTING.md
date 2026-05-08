# Contributing to go-scim

## Why this file exists

This project is a from-scratch rewrite of an earlier SCIM v2 implementation that collapsed under three architectural pitfalls: a generic in-memory property tree that could not be cleanly mapped to SQL, a runtime schema-interpretation layer that made the binary depend on external schema files, and a creeping accumulation of vendor-specific tolerances that made the codebase unreviewable. The legacy code is preserved under `.legacy/` as a frozen reference so that the failure modes can be cited concretely. The rules below exist so the rewrite does not regress into them.

The rules are hard prohibitions, not principles. They are written to be quoted verbatim in PR review comments. If you believe you need to break a rule, see "If you absolutely need to break a rule" at the end.

---

## Rule 1: No generic Property / tree models in `rt` or generated code

**You MUST NOT introduce a `Property` interface, a generic-tree node type, a path-walking navigator, or a subscriber/observer pattern over a generic resource model.**

### Why

The legacy implementation modelled every SCIM resource as a tree of typed `Property` nodes carrying schema-attribute references. Concrete artifacts in this repo:

- `.legacy/pkg/v2/prop/property.go` — defines the `Property` interface that every resource node implemented. Operating on a `Property` meant calling `Path(...)`, `Raw()`, `Hash()`, etc., instead of touching a typed Go field.
- `.legacy/pkg/v2/prop/navigator.go` — tree traversal with event propagation. Every read or write threaded through this navigator.
- `.legacy/pkg/v2/prop/subscriber.go` — the `ExclusivePrimarySubscriber` (lines 106-171) enforced "at most one element of a multi-valued complex attribute can be `primary`" by reacting to events the navigator emitted. A SCIM invariant that should be a single line in a generated `Validate()` method became a global subscriber, an event system, and a tree walk.

This shape made the persistence layer untenable: BSON (legacy MongoDB adapter) is also tree-shaped, so the impedance was hidden; SQL is not, so the impedance was lethal. The rewrite exists because that impedance could not be paid down. Reintroducing the tree, even "just for the runtime support library," means the rewrite produced nothing.

### Warning signs

- A new `interface { Path(string) ... }` declaration anywhere in `rt/` or generated code.
- A `Property` interface declaration in any form.
- A `Navigator`, `Visitor`, or `Walker` type that operates on a generic resource (filter evaluation walking generated typed accessors is fine; walking a tree of `Property` nodes is not).
- A subscriber / observer / event-bus pattern proposed for SCIM rule enforcement (e.g., "primary exclusivity," "auto-compact," "schema sync"). SCIM invariants belong in generated `Validate()` / `ApplyPatch()` methods on typed structs.
- Factory functions like `NewString(...)`, `NewComplex(...)`, `NewMulti(...)` that return a generic node type instead of a concrete typed Go struct.
- Anyone using the phrase "we can do this generically with reflection" or "let's add a `Resource` interface so services are reusable."

### If you think you need this

Open an issue describing the use case. Do not bypass this rule in a PR. Filter evaluation has a legitimate need for runtime polymorphism (the filter expression itself is a value tree); discovery endpoints (`/Schemas`, `/ResourceTypes`, `/ServiceProviderConfig`) have a legitimate need for schema content at runtime — but both are served by compile-time-evaluated artifacts (typed filter compilers; embedded schema JSON), not by a generic resource tree.

---

## Rule 2: No runtime schema interpretation — discovery payloads are `//go:embed` only

**You MUST NOT load schema content from disk at runtime in `rt/` or generated code, register schemas in a global mutable registry, or dispatch SCIM operations through a path resolver that consults a schema record at request time.**

### Why

The legacy server loaded schema JSON files from disk at startup and registered them in a mutable package-level registry. Concrete artifacts in this repo:

- `.legacy/cmd/internal/args/scim.go` — reads schema JSON files from a configurable directory and registers them before the server boots. The binary cannot be deployed without those files alongside.
- `.legacy/public/schemas/` — the schema JSON content that was read at runtime (`core_schema.json`, `user_schema.json`, `group_schema.json`, `user_enterprise_extension_schema.json`).

Consequences: the binary is not self-contained, schema and code can drift silently (a hand-edit of `user_schema.json` does not regenerate types), per-request operations pay schema-lookup cost, and panics on missing registrations are easy to trigger and hard to diagnose. In v3, schemas are inputs to the generator, not inputs to the generated server. The same definition produces the typed Go structs and the canonical schema JSON; the generated server embeds (`//go:embed`) the JSON and serves it on `/Schemas` / `/ResourceTypes`. There is one source of truth.

### Warning signs

- Any `os.ReadFile(...)` whose path argument names or aliases "schema" (e.g., `os.ReadFile(schemaPath)`, `os.ReadFile("schemas/User.json")`, `os.ReadFile(filepath.Join(schemaDir, ...))`).
- A `--schemas-dir`, `--schema-path`, or `SCIM_SCHEMA_DIR` flag / env var on the generated server or anything in `rt/`.
- A package-level mutable variable named `schemaRegistry`, `schemas`, `schemaCache` that is written to from multiple call sites.
- A `RegisterSchema(...)` function exposed by `rt/`.
- Code that takes a schema URN at request time, looks up an `Attribute` record, and dispatches PATCH / filter / validation through that record — instead of calling a generated, type-specific method on the typed struct.
- Any `json.Unmarshal` of a schema document inside `rt/` or generated code (parsing schemas at server start).

### If you think you need this

Open an issue. The two legitimate "I need schemas at runtime" cases — serving discovery endpoints, and validating extension-URN-prefixed paths in PATCH — are served by `//go:embed` of generator-produced JSON and by generated typed methods, respectively. If your case does not fit either, the rule is the conversation, not the code.

---

## Rule 3: No IdP-specific accommodations (Entra, Okta, Azure AD, etc.) in v1

**You MUST NOT add code that accommodates an identity-provider-specific deviation from RFC 7643 / RFC 7644 in v1. Vendor names MUST NOT appear in source code, commit messages, file names, or branch names. The generated server enforces the spec strictly; non-compliant payloads receive `400 invalidSyntax`.**

### Why

The legacy codebase accreted ad-hoc tolerances for IdP quirks — case-insensitive `op` matching, lenient boolean parsing, vendor-specific `members` PATCH handling — and the spec layer became unreviewable. Each accommodation looked small in isolation; the cumulative effect was a codebase no one trusted. PROJECT.md explicitly lists IdP non-compliance accommodations as out of scope for v1. The legacy code does not contain a single canonical artifact to cite (the brittleness was diffuse, not localised to one file); see `.planning/research/PITFALLS.md` Pitfall 3 for the documented failure mode and the v2 plan that replaces piecemeal accommodation with a definition-level "compatibility profile" concept.

This rule is preventive. v1 ships with strict compliance. v2 will introduce profiles as a definition-level feature that produces alternative generator output paths — the right place for vendor variance. v1 must resist the piecemeal accommodation.

### Warning signs

- Vendor names (`Entra`, `Okta`, `Azure AD`, `AzureAD`, `Microsoft`, `OneLogin`, `Ping`, `JumpCloud`) appearing anywhere in `gen/`, `rt/`, generated code, commit messages, branch names, or file names. (They appear legitimately in `.planning/`, `.legacy/`, `CONTRIBUTING.md`, `PROJECT.md`, `README.md` — those locations are exempt.)
- Conditional logic keyed on `User-Agent`, vendor-specific request headers, or any "is this Entra?" detection.
- Boolean flags or struct fields named `Lenient`, `Tolerant`, `IdPCompat`, `LooseParsing`, `CaseInsensitive` (when applied to spec-defined values like `op` or attribute values).
- Case-insensitive comparisons of spec-defined enumerations (`op`, `mutability`, `returned`, etc.) — RFC values are case-sensitive in the wire format.
- Tolerant parsers for boolean / integer values (e.g., accepting `"True"` as `true`).
- A `compat/`, `tolerance/`, `vendors/`, or `quirks/` package or directory anywhere in the project.
- Tests named `TestEntraQuirk_*`, `TestOktaCompat_*`, etc., outside an explicitly-quarantined "future v2 design notes" location.
- Commit messages of the form "fix Entra PATCH issue" / "make Okta happy" / "allow Azure AD capitalization."

### If you think you need this

Open an issue describing the IdP and the deviation. The issue is the v2 profile-design backlog, not the v1 patch queue. If the IdP is non-compliant, the README's compatibility table records the fact ("Tested with Entra: requires their compliance feature flag") — the server does not bend.

---

## How this is enforced

Four levers, defense in depth. No single lever is the security boundary; the combination is.

1. **PR template** (`.github/pull_request_template.md`) — auto-applied to every new PR. Three checkboxes (one per rule) must be answered "No" before review proceeds. Honor system at the PR-author level, but combined with CODEOWNERS and CI it is hard to bypass without notice.
2. **CODEOWNERS** (`.github/CODEOWNERS`) — pins `@imulab` as required reviewer for `/CONTRIBUTING.md`, `/PROJECT.md`, `/.github/`, `/gen/`, `/rt/`. Sensitive paths cannot merge without owner review (when paired with branch protection, configured in the GitHub UI).
3. **lefthook hooks** (`lefthook.yml`) — local pre-commit and commit-msg hooks reject staged Go files in `gen/` or `rt/` that contain forbidden symbols (Rules 1 and 2 regex set), and reject commit messages containing vendor names (Rule 3). Bypassable via `git commit --no-verify`; the CI mirror (lever 4) is the actual gate.
4. **CI grep gate** (`.github/workflows/ci.yml`, planned in Plan 02) — re-runs the same forbidden-symbol regex set on every PR commit and the PR's commit-message range. Cannot be bypassed by a PR author. This is the security boundary.

The forbidden-symbol regex set is committed verbatim in `lefthook.yml` and is mirrored exactly in CI. The set:

- `os\.ReadFile.*[Ss]chema` — Rule 2 (runtime schema loading)
- `interface\s*\{\s*Path\(` — Rule 1 (Property-tree relapse signature)
- `Property\s+interface` — Rule 1 (interface declaration form)
- `Entra|Okta|Azure\s*AD|AzureAD` — Rule 3 (vendor-name accommodation)

---

## If you absolutely need to break a rule

The rules are amendable via project decision, not bypassable via PR.

1. Open an issue. Describe the use case, the rule it conflicts with, and the alternative you considered.
2. The amendment lands in `PROJECT.md` Key Decisions and in `CONTRIBUTING.md` here, in the same PR. The CI grep set in `lefthook.yml` and the workflow are updated to match.
3. PROJECT.md "Out of Scope" already reserves the right to scope additions; rule amendments use the same channel.

If you are about to merge a PR that violates one of these rules without an accompanying amendment in `CONTRIBUTING.md` and `PROJECT.md`, stop. The rule is the conversation.
< codeowners test >

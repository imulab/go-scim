# Phase 0: Architecture Spike & Repo Setup - Context

**Gathered:** 2026-05-07
**Status:** Ready for planning

<domain>
## Phase Boundary

Encode the three catastrophic anti-relapse rules in writing, stand up the multi-module workspace (`gen` + `rt`), and decide the Go-source emission engine via a small spike — all before any emitted code exists.

In scope:
- Architecture rules document and enforcement plumbing (PR template, CI grep gates, CODEOWNERS, commit-message linter)
- Two `go.mod` modules (`gen`, `rt`) linked by a gitignored `go.work`
- CI pipeline that builds every module both with and without `GOWORK`
- Emission-engine spike (`text/template` + `go/format` vs `dave/jennifer`) producing one representative artifact, decision recorded in PROJECT.md Key Decisions

Out of scope (other phases):
- Any actual generator code, IR, validator, emitters (Phase 1+)
- Filter parser, PATCH applier, persistence, HTTP server (Phases 4–5)
- CLI subcommand bodies (Phase 7)

</domain>

<decisions>
## Implementation Decisions

### Anti-relapse rules document
- Location: `CONTRIBUTING.md` at repo root (single source of truth, GitHub-recognized)
- Tone: hard prohibitions + rationale + warning signs (mirrors PITFALLS.md structure; e.g., "You MUST NOT introduce a generic `Property` interface. Reason: … Warning signs: …")
- Rules enumerated explicitly:
  1. No generic `Property`/tree models in `rt` or generated code
  2. No runtime schema interpretation (no `os.ReadFile` of schema content; discovery payloads are `//go:embed`-only)
  3. No IdP-specific accommodations (Entra/Okta/Azure AD/etc.) in v1
- Legacy `.legacy/` referenced with concrete file paths from `.planning/codebase/CONCERNS.md` (e.g., `pkg/v2/spec/schema.go`, `ExclusivePrimarySubscriber`) so the "why" is visceral

### Enforcement (all four levers, per user choice)
- **PR template** (`.github/pull_request_template.md`) asks: "Does this PR introduce IdP-specific tolerance? [must be no]" plus "Does this PR introduce a runtime-typed Resource interface? [must be no]"
- **CI grep gate** (forbidden-symbol job) fails the build on any of:
  - `os\.ReadFile.*schema` in `rt/` or generated code
  - `interface\s*{\s*Path\(` (Property-tree relapse signature)
  - `Property\s+interface` declarations
  - `Entra|Okta|Azure\s*AD|AzureAD` in source files (vendor names anywhere outside research/PITFALLS docs)
- **CODEOWNERS** pinning `rt/` core packages, `CONTRIBUTING.md`, and `.github/` to repo owner so sensitive edits force review
- **Commit-message linter** (pre-commit hook + CI mirror) rejects commits containing vendor names (Entra/Okta/AzureAD) per Pitfall 3 warning sign

### Module path & directory layout
- Module URI base: `github.com/imulab/go-scim`
- Two modules: `github.com/imulab/go-scim/gen` and `github.com/imulab/go-scim/rt`
  - **Override of REQ-REPO-02** (which named modules `scimgen` + `scimrt`): user prefers short names without `scim` repetition; users alias on import if desired. Planner must update REQUIREMENTS.md / ROADMAP.md references during planning.
- Directory layout: `gen/` and `rt/` matching module names (zero translation cost)
- CLI binary name: `scim` (top-level command: `scim generate`, `scim check`, `scim dump-ir`)
  - **Override of REQ-CLI-01** (which named CLI `scimgen`): user prefers `scim` as a shorter top-level command. Planner reconciles REQUIREMENTS.md and Phase 7.
- `go.work` and `go.work.sum` gitignored (REQ-REPO-01 honored)
- `examples/` module deferred to Phase 2 (when first emitted artifact exists to commit and continuously compile)
- `.legacy/` kept as-is, excluded from `go.work` `use` directives, never built

### Emission-engine spike
- Spike artifact: typed `User` struct + `Validate()` method — exercises conditional imports (`time`, `regexp`, `errors`), nested types, branching validation (required, canonicalValues, multi-valued primary cardinality)
- Both implementations live in `spike/` (excluded from `go.work` `use`):
  - `spike/template/` — `text/template` rendering into `bytes.Buffer`, then `go/format.Source`
  - `spike/jennifer/` — `dave/jennifer` programmatic AST
- Output Go file from both must be byte-identical after `gofmt`
- Evaluation criteria (all four weighted):
  1. **Import management** — does the engine auto-track imports across conditional emission paths?
  2. **Source readability** — can a contributor reading the emitter understand what it produces?
  3. **Diff stability** — two regen runs produce byte-identical Go (REQ-GEN-06 dry run)
  4. **Mixed-output compatibility** — Go-emission winner does not preclude using `text/template` for non-Go artifacts (SQL/Makefile/README) per STACK.md
- Decision + evidence recorded in `PROJECT.md` under a new **Key Decisions** section (short prose summary citing the spike directory)
- Both spike implementations retained in-repo (not deleted post-decision) so the comparison stays auditable

### CI shape
- Provider: GitHub Actions
- Go version matrix: **1.25** and **1.26** (user is on 1.26.1)
  - **Override of STACK.md** which set Go floor at 1.24. Planner: bump `go` directive in both `go.mod` files to 1.25 and update STACK.md / REQUIREMENTS notes.
- OS: `ubuntu-latest` only in Phase 0 (cross-OS expansion deferred until generator emits real code)
- GOWORK gate: **two parallel jobs**
  - Job A (workspace mode): `go build ./...` from repo root with `go.work` present (recreated in CI from `go.work.example` or via `go work init && go work use ./gen ./rt`)
  - Job B (isolated mode): `cd gen && GOWORK=off go build ./...` then same for `rt/`
  - Disagreement between A and B fails the build
- Phase 0 CI scope: build only + forbidden-symbol grep gate. No tests (no code), no lint (yet). golangci-lint introduced when first real emitter lands.

### Claude's Discretion
- Exact wording / heading structure inside `CONTRIBUTING.md`
- Pre-commit hook implementation (lefthook vs husky-style vs plain bash in `.git/hooks/`)
- CODEOWNERS exact path globs
- GitHub Actions YAML structure (composite vs reusable workflow vs inline)
- Spike scoring rubric (qualitative vs LOC-counted)
- Whether to ship a `Makefile` or `justfile` for `regen` / `lint` / `ci-local` targets

</decisions>

<specifics>
## Specific Ideas

- User wants short module names (`gen`, `rt`) and short CLI (`scim`) — explicit "don't repeat scim references in sub-modules"
- All four enforcement levers (PR template + CI grep + CODEOWNERS + commit-msg linter) chosen — strong stance; no soft enforcement
- Hard prohibition tone preferred over principles — direct citability in PR comments
- Legacy `.legacy/` to remain referenced with concrete file paths in CONTRIBUTING.md (anchors visceral motivation for the rules)
- Spike retained in-repo as audit trail; future "but jennifer is bigger / templates are uglier" arguments must point to evidence
- Emission-engine decision recorded in `PROJECT.md` Key Decisions, not a new ADR system

</specifics>

<code_context>
## Existing Code Insights

### Reusable Assets
- None. Repo is empty under v3 working tree (only `CLAUDE.md` at root, `.planning/` planning docs, `.legacy/` frozen reference).

### Established Patterns
- `.planning/` discipline already in place: research → roadmap → phase context → plan → execute → verify
- Codebase docs at `.planning/codebase/` describe intended layout (STACK.md, ARCHITECTURE.md, etc.) — informational, not source of truth (ROADMAP + REQUIREMENTS supersede)

### Integration Points
- **Phase 0 → Phase 1**: scaffolded `gen/` module receives generator, IR types, validator
- **Phase 0 → Phase 2**: emission-engine choice locked; first emitter (domain) ships in `gen/internal/emit/...`; `examples/` module added
- **Phase 0 → all later phases**: anti-relapse rules + CI grep gates apply continuously; PR template + CODEOWNERS active from first PR after Phase 0 merges
- **Legacy boundary**: `.legacy/` excluded from build via go.work; CONTRIBUTING.md links into it for anti-pattern citation only

</code_context>

<deferred>
## Deferred Ideas

- **Cross-OS CI matrix (macos / windows)** — deferred until generator emits real code where OS-specific behavior could regress (Phase 2+)
- **golangci-lint / go vet hygiene jobs** — added when first real emitter exists (Phase 1 or 2)
- **`examples/` module + continuous-compile canary** — Phase 2 (needs first emitted artifact)
- **ADR system (`docs/adr/...`)** — not started in Phase 0; PROJECT.md Key Decisions section is sufficient for v1. Revisit if decision count grows.
- **Branch protection rules / required reviewers** — outside scope of repo files; configured in GitHub UI by repo owner.
- **Makefile vs justfile choice** — Claude's discretion during planning.

</deferred>

<requirements_overrides>
## Requirements Overrides Flagged for Planner

User decisions in this discussion deviate from the written REQUIREMENTS / STACK. Planner must reconcile during `/gsd:plan-phase 0`:

| Override | From | To | Action |
|----------|------|-----|--------|
| Module names | `scimgen` + `scimrt` (REQ-REPO-02) | `gen` + `rt` | Update REQUIREMENTS.md REQ-REPO-02; update ROADMAP.md phase descriptions; update STACK.md draft layout |
| CLI binary name | `scimgen` (REQ-CLI-01) | `scim` | Update REQ-CLI-01 wording; planner notes for Phase 7 |
| Go version floor | 1.24 (STACK.md) | 1.25 (CI matrix 1.25 + 1.26) | Update STACK.md version pin; both `go.mod` files declare `go 1.25` |

These overrides are user-driven decisions made during phase context gathering. Planner applies them as part of Phase 0 plan; no separate approval needed.

</requirements_overrides>

---

*Phase: 00-architecture-spike-repo-setup*
*Context gathered: 2026-05-07*

# Phase 0: Architecture Spike & Repo Setup - Research

**Researched:** 2026-05-07
**Domain:** Go multi-module workspace setup, Go-source code-emission engine selection, anti-relapse rule enforcement (PR template, CODEOWNERS, CI grep gates, commit-msg linter), GitHub Actions CI for Go monorepos
**Confidence:** HIGH

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions

**Anti-relapse rules document**
- Location: `CONTRIBUTING.md` at repo root (single source of truth, GitHub-recognized)
- Tone: hard prohibitions + rationale + warning signs (mirrors PITFALLS.md structure; e.g., "You MUST NOT introduce a generic `Property` interface. Reason: … Warning signs: …")
- Rules enumerated explicitly:
  1. No generic `Property`/tree models in `rt` or generated code
  2. No runtime schema interpretation (no `os.ReadFile` of schema content; discovery payloads are `//go:embed`-only)
  3. No IdP-specific accommodations (Entra/Okta/Azure AD/etc.) in v1
- Legacy `.legacy/` referenced with concrete file paths from `.planning/codebase/CONCERNS.md` (e.g., `pkg/v2/spec/schema.go`, `ExclusivePrimarySubscriber`) so the "why" is visceral

**Enforcement (all four levers)**
- **PR template** (`.github/pull_request_template.md`) asks: "Does this PR introduce IdP-specific tolerance? [must be no]" plus "Does this PR introduce a runtime-typed Resource interface? [must be no]"
- **CI grep gate** (forbidden-symbol job) fails the build on any of:
  - `os\.ReadFile.*schema` in `rt/` or generated code
  - `interface\s*{\s*Path\(` (Property-tree relapse signature)
  - `Property\s+interface` declarations
  - `Entra|Okta|Azure\s*AD|AzureAD` in source files (vendor names anywhere outside research/PITFALLS docs)
- **CODEOWNERS** pinning `rt/` core packages, `CONTRIBUTING.md`, and `.github/` to repo owner so sensitive edits force review
- **Commit-message linter** (pre-commit hook + CI mirror) rejects commits containing vendor names (Entra/Okta/AzureAD) per Pitfall 3 warning sign

**Module path & directory layout**
- Module URI base: `github.com/imulab/go-scim`
- Two modules: `github.com/imulab/go-scim/gen` and `github.com/imulab/go-scim/rt`
  - **Override of REQ-REPO-02** (which named modules `scimgen` + `scimrt`): user prefers short names without `scim` repetition
- Directory layout: `gen/` and `rt/` matching module names (zero translation cost)
- CLI binary name: `scim` (top-level command: `scim generate`, `scim check`, `scim dump-ir`)
  - **Override of REQ-CLI-01** (which named CLI `scimgen`)
- `go.work` and `go.work.sum` gitignored (REQ-REPO-01 honored)
- `examples/` module deferred to Phase 2
- `.legacy/` kept as-is, excluded from `go.work` `use` directives, never built

**Emission-engine spike**
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
- Decision + evidence recorded in `PROJECT.md` under a new **Key Decisions** section
- Both spike implementations retained in-repo (not deleted post-decision) so the comparison stays auditable

**CI shape**
- Provider: GitHub Actions
- Go version matrix: **1.25** and **1.26** (user is on 1.26.1)
  - **Override of STACK.md** which set Go floor at 1.24
- OS: `ubuntu-latest` only in Phase 0 (cross-OS expansion deferred)
- GOWORK gate: **two parallel jobs**
  - Job A (workspace mode): `go build ./...` from repo root with `go.work` present (recreated in CI)
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

### Deferred Ideas (OUT OF SCOPE)
- **Cross-OS CI matrix (macos / windows)** — until Phase 2+
- **golangci-lint / go vet hygiene jobs** — added when first real emitter exists
- **`examples/` module + continuous-compile canary** — Phase 2
- **ADR system (`docs/adr/...`)** — PROJECT.md Key Decisions section sufficient for v1
- **Branch protection rules / required reviewers** — outside scope of repo files; configured in GitHub UI
- **Makefile vs justfile choice** — Claude's discretion during planning
</user_constraints>

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|-----------------|
| REPO-01 | Multi-module workspace via `go.work` (gitignored); CI runs with `GOWORK=off` to validate each module standalone | "go.work + GOWORK=off" section below: official Go reference confirms `GOWORK=off` puts the command into single-module context, ignoring `go.work`. Two-job CI pattern (workspace vs isolated) is documented best practice. |
| REPO-02 | Two `go.mod` modules (per CONTEXT.md override: `gen` + `rt`, not `scimgen` + `scimrt`) | "Workspace structure" section below: `go work init && go work use ./gen ./rt` produces the expected workspace. Each module declares its own `go.mod` with `module github.com/imulab/go-scim/{gen,rt}`. Planner must update REQUIREMENTS.md REQ-REPO-02 wording. |
| REPO-03 | License is MIT | Single LICENSE file at repo root copying `.legacy/LICENSE` text. No research needed. |
| REPO-04 | Contributor architecture rules document encodes anti-relapse rules | "Enforcement levers" section below: `CONTRIBUTING.md` at repo root is the canonical GitHub-recognized location for contributor rules. Hard-prohibition tone aligns with PITFALLS.md Pitfalls 1, 2, 3. |
| REPO-05 | PR template asks reviewers to verify changes do not introduce IdP-specific tolerance | "PR template" section below: `.github/pull_request_template.md` is the GitHub-recognized location; auto-applies to every new PR. |

**Note on requirements overrides** (from CONTEXT.md `<requirements_overrides>` block):
- `scimgen`/`scimrt` → `gen`/`rt` (REQ-REPO-02): planner updates REQUIREMENTS.md, ROADMAP.md, STACK.md
- `scimgen` CLI → `scim` (REQ-CLI-01): planner updates REQ-CLI-01 wording, planner notes for Phase 7
- Go floor 1.24 → 1.25 (STACK.md): both `go.mod` files declare `go 1.25`; CI matrix tests 1.25 + 1.26
</phase_requirements>

## Summary

Phase 0 is purely organisational: stand up the repo skeleton, encode written rules, and pick a Go-source emission engine via a small spike. There is **no SCIM logic** in this phase. Research focused on five concrete domains: (1) Go workspace + `GOWORK=off` semantics, (2) `dave/jennifer` vs `text/template` + `go/format` for code emission, (3) GitHub Actions setup-go matrix patterns and Go 1.26 availability, (4) GitHub conventions for `CONTRIBUTING.md` / `.github/pull_request_template.md` / `CODEOWNERS`, (5) commit-msg linter implementation choices for a Go-only project.

The dominant finding for the spike: **Jennifer (v1.7.1, Sept 2024)** offers automatic import management via `Qual()` — exactly the capability that becomes painful with `text/template` once conditional imports (`time`/`regexp`/`errors` based on attribute kinds) enter the picture. **`text/template` + `go/format.Source`** is simpler for static boilerplate but pushes import bookkeeping into the template author. The user's spike protocol (build the same `User` struct + `Validate()` artifact in both, score on import management / readability / diff stability / mixed-output compatibility) is the right shape because it isolates exactly the dimensions where the two engines diverge. **`golang.org/x/tools/imports.Process`** can rescue `text/template` from manual import bookkeeping (it is the library form of `goimports`), which is a third option worth surfacing in the spike report even if it isn't itself a separate spike implementation.

For workspace + CI: `go.work` is gitignored (per REQ-REPO-01), `go.work.sum` is also gitignored (per-developer state), CI re-creates the workspace inline via `go work init && go work use ./gen ./rt` for the workspace-mode job, and runs `GOWORK=off go build ./...` from each module root for the isolated job. Go 1.26 stable shipped 2026-02-10 (latest patch 1.26.2 from 2026-04-07), so the `1.25 + 1.26` matrix uses `go-version: '1.26'` (resolves to latest 1.26.x patch on `actions/setup-go@v6`).

**Primary recommendation:** Build the spike in `spike/template/` and `spike/jennifer/` (both excluded from `go.work`), evaluate against the four user-defined criteria, but explicitly note that **`text/template` + `imports.Process`** is the practical fallback if Jennifer's API ergonomics outweigh its import-management win. Record the decision and evidence in `PROJECT.md` Key Decisions (no separate ADR system per CONTEXT.md). Wire the four enforcement levers (PR template, CI grep, CODEOWNERS, commit-msg linter) using **lefthook** for git hooks (single Go binary, no Node.js dependency) — strictly Claude's discretion per CONTEXT.md but the only hook manager that fits a Go-only project's runtime profile.

## Standard Stack

### Core

| Library / Tool | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| Go | 1.25 (floor) / 1.26 (CI matrix top) | Toolchain | User runs 1.26.1; STACK.md override per CONTEXT.md bumps floor from 1.24 to 1.25; matrix tests both. Go 1.26 stable shipped 2026-02-10 ([Go 1.26 release blog](https://go.dev/blog/go1.26)). |
| `dave/jennifer` (`github.com/dave/jennifer/jen`) | v1.7.1 (Sept 2024) | Programmatic Go AST emission — spike candidate A | 3.6k stars, 1,198 importers, MIT, "stability-stable" badge. Automatic import management via `Qual(path, name)` — solves the conditional-imports problem cleanly ([pkg.go.dev/github.com/dave/jennifer/jen](https://pkg.go.dev/github.com/dave/jennifer/jen)). |
| `text/template` (stdlib) + `go/format.Source` | bundled with Go | Template-based emission — spike candidate B | Zero external deps; `go/format.Source` (the in-process equivalent of `gofmt`) catches template syntax errors at template-execute time and refuses to write invalid Go ([pkg.go.dev/text/template](https://pkg.go.dev/text/template)). |
| `golang.org/x/tools/imports` (`imports.Process`) | tracked with `golang.org/x/tools` | Auto-manage imports for `text/template` output (post-process) | Library form of `goimports`. Lets `text/template` win on simplicity without losing on import bookkeeping when conditional imports happen ([pkg.go.dev/golang.org/x/tools/imports](https://pkg.go.dev/golang.org/x/tools/imports)). |
| `actions/setup-go@v6` | v6 | GitHub Actions Go install | Resolves `go-version: '1.26'` to the latest 1.26.x patch on Ubuntu runners; tool cache speeds installs ([actions/setup-go](https://github.com/actions/setup-go)). |
| `actions/checkout@v4` | v4 | Source checkout | Standard. |

### Supporting

| Library / Tool | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| `lefthook` (CLI) | latest stable | Git hook manager (pre-commit, commit-msg) | Single Go binary, no Node.js dependency. Parallel hook execution. Configured via `lefthook.yml` at repo root ([evilmartians/lefthook](https://github.com/evilmartians/lefthook)). |
| GNU Make (or `justfile`) | system | `regen` / `lint` / `ci-local` task runner | Per CONTEXT.md, choice is Claude's discretion. Makefile is more universally available; `just` has cleaner syntax. |

### Alternatives Considered

| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| `dave/jennifer` for Go emission | `text/template` + `go/format.Source` only | Simpler if no conditional imports; harder once imports vary by attribute kind. |
| `dave/jennifer` for Go emission | `text/template` + `imports.Process` (goimports as library) | Compromise: keep templates for readability, let `imports.Process` do import management. Spike should produce evidence either way. |
| `dolmen-go/codegen` | `text/template` directly | Tiny wrapper around `text/template` + `gofmt` + DO-NOT-EDIT marker enforcement. Marginal value over hand-rolled wrapper; only adopt if its API matches the generator's needs. |
| `lefthook` for hooks | `pre-commit` (Python) | Requires Python runtime — unwanted for a Go-only project. |
| `lefthook` for hooks | `husky` (Node) | Requires Node.js runtime — unwanted for a Go-only project. |
| `lefthook` for hooks | Plain bash in `.git/hooks/` (committed via a setup script) | Zero deps, but no parallelism, no `staged_files` magic, manual install dance. Acceptable for a project with only 1–2 hooks. |
| Makefile | `justfile` | `just` is more readable, but Make is universally available on dev machines and CI runners. Either works. |

**Installation (CLI tools used during planning execution):**
```bash
# Go toolchain (developer machine — confirmed user is on 1.26.1)
# (No install step needed for stdlib text/template, go/format)

# Jennifer — added as a dep to spike/jennifer/go.mod only (not gen or rt)
go get github.com/dave/jennifer@v1.7.1

# imports.Process — added as a dep wherever the chosen emission engine consumes it
# (Likely gen/go.mod once the spike concludes; spike/template/go.mod for the spike itself)
go get golang.org/x/tools@latest

# lefthook — installed by contributor; binary on PATH
go install github.com/evilmartians/lefthook@latest
# OR: brew install lefthook   # OR: download release binary
```

## Architecture Patterns

### Recommended Project Structure (Phase 0 deliverable)

```
go-scim/                           # repo root, github.com/imulab/go-scim
├── .github/
│   ├── workflows/
│   │   └── ci.yml                 # build (workspace + isolated), forbidden-symbol grep, commit-msg lint mirror
│   ├── pull_request_template.md   # IdP-tolerance + Resource-interface gates (must-be-no)
│   └── CODEOWNERS                 # rt/, .github/, CONTRIBUTING.md pinned to repo owner
├── .legacy/                        # frozen reference; NOT in go.work; NEVER built
│   └── (existing legacy code untouched)
├── spike/                          # emission-engine spike; excluded from go.work `use`
│   ├── template/
│   │   ├── go.mod
│   │   ├── main.go                 # text/template + go/format.Source emitter
│   │   ├── user.go.tmpl
│   │   └── output/
│   │       └── user_gen.go         # spike output — committed for diff comparison
│   ├── jennifer/
│   │   ├── go.mod
│   │   ├── main.go                 # dave/jennifer programmatic emitter
│   │   └── output/
│   │       └── user_gen.go         # spike output — committed for diff comparison
│   └── README.md                   # spike protocol + evaluation criteria + result link
├── gen/                            # github.com/imulab/go-scim/gen — generator module (empty in Phase 0)
│   └── go.mod                      # `go 1.25`, module path, no source files yet
├── rt/                             # github.com/imulab/go-scim/rt — runtime support module (empty in Phase 0)
│   └── go.mod                      # `go 1.25`, module path, no source files yet
├── lefthook.yml                    # pre-commit + commit-msg hooks
├── Makefile                        # regen / lint / ci-local targets (or justfile)
├── CONTRIBUTING.md                 # three anti-relapse rules with rationale + warning signs
├── CLAUDE.md                       # existing — keep
├── LICENSE                         # MIT (copied from .legacy/LICENSE)
├── README.md                       # short — links to PROJECT.md, CONTRIBUTING.md
├── PROJECT.md                      # add new "Key Decisions" section for spike outcome
├── go.work                         # GITIGNORED — recreated locally and in CI
├── go.work.sum                     # GITIGNORED — per-developer state
└── .gitignore                      # ignores go.work, go.work.sum, build artifacts
```

### Pattern 1: `go.work` + GOWORK=off two-job CI gate

**What:** Run two CI jobs in parallel: one with `go.work` present (workspace mode), one with `GOWORK=off` (each module isolated). If they disagree, fail the build.

**When to use:** Always for a multi-module repo where `go.work` is gitignored. This pattern catches the class of bug where a developer's local `go.work` workspace masks a missing `go.mod` `require` line that breaks a downstream consumer (per [Pitfall 18 in PITFALLS.md](file://./.planning/research/PITFALLS.md)).

**Example (GitHub Actions workflow shape):**
```yaml
# Source: official setup-go matrix docs + go.work reference
# https://github.com/actions/setup-go
# https://go.dev/ref/mod#workspaces

name: ci
on: [push, pull_request]

jobs:
  build-workspace:
    name: build (workspace mode, go ${{ matrix.go }})
    runs-on: ubuntu-latest
    strategy:
      fail-fast: false
      matrix:
        go: ['1.25', '1.26']
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v6
        with:
          go-version: ${{ matrix.go }}
      - name: Recreate workspace
        run: |
          go work init
          go work use ./gen ./rt
      - name: Build workspace
        run: go build ./...

  build-isolated:
    name: build (GOWORK=off, go ${{ matrix.go }})
    runs-on: ubuntu-latest
    strategy:
      fail-fast: false
      matrix:
        go: ['1.25', '1.26']
        module: [gen, rt]
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v6
        with:
          go-version: ${{ matrix.go }}
      - name: Build module in isolation
        env:
          GOWORK: 'off'
        run: |
          cd ${{ matrix.module }}
          go build ./...

  forbidden-symbols:
    name: forbidden-symbol grep gate
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Grep for relapse signatures
        run: |
          ./.github/scripts/forbidden-symbols.sh
          # Script greps for: os\.ReadFile.*schema (in rt/, gen/output/, generated code)
          #                   interface\s*\{\s*Path\(
          #                   Property\s+interface
          #                   Entra|Okta|Azure\s*AD|AzureAD (excluding .planning/, .legacy/, CONTRIBUTING.md)
          # Exits 1 on any match.

  commit-msg-lint:
    name: commit message vendor-name gate
    runs-on: ubuntu-latest
    if: github.event_name == 'pull_request'
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0
      - name: Check PR commit messages for vendor names
        run: |
          base=${{ github.event.pull_request.base.sha }}
          head=${{ github.event.pull_request.head.sha }}
          if git log "${base}..${head}" --format=%B | grep -iE 'Entra|Okta|Azure\s*AD|AzureAD'; then
            echo "::error::Commit message contains vendor name (per Pitfall 3)."
            exit 1
          fi
```

### Pattern 2: Spike directory excluded from workspace

**What:** Place `spike/template/` and `spike/jennifer/` as standalone modules with their own `go.mod` files, but **do not** list them in `go.work`'s `use` directive.

**When to use:** Whenever the spike depends on a library (Jennifer) you don't want leaking into `gen`'s or `rt`'s dependency graph.

**Example:**
```
# repo root
go work init
go work use ./gen ./rt
# (spike/template/ and spike/jennifer/ deliberately NOT added)

# In spike/jennifer/go.mod:
module github.com/imulab/go-scim/spike/jennifer

go 1.25

require github.com/dave/jennifer v1.7.1

# spike/jennifer is built independently:
cd spike/jennifer && go build ./...
```

This isolates the spike from the production graph: `gen` never accidentally `import`s Jennifer, and `git grep dave/jennifer gen/ rt/` returns nothing — a useful CI gate post-decision.

### Pattern 3: PR template forced-checkbox gate

**What:** Use `.github/pull_request_template.md` to put two checkboxes that must be answered with `[x] No` before review proceeds.

**When to use:** When you need a **human-attested** confirmation that a specific anti-pattern wasn't introduced. Checkboxes are not enforced by GitHub itself but become a shared norm; CODEOWNERS + branch protection enforce the review.

**Example:**
```markdown
<!-- Source: GitHub PR template conventions
     https://docs.github.com/en/pull-requests/collaborating-with-pull-requests/getting-started/managing-and-standardizing-pull-requests -->

## What does this PR do?

<!-- One sentence. -->

## Why is this change needed?

<!-- Link to issue / decision in PROJECT.md / context. -->

## Anti-relapse self-attestation (per CONTRIBUTING.md)

- [ ] **Does this PR introduce IdP-specific tolerance** (Entra, Okta, Azure AD, etc.)?
      Answer: **No** (if Yes, this PR will be closed; see CONTRIBUTING.md §3 and PITFALLS.md Pitfall 3).

- [ ] **Does this PR introduce a runtime-typed `Resource` interface or a generic `Property` tree**?
      Answer: **No** (if Yes, this PR will be closed; see CONTRIBUTING.md §1 and PITFALLS.md Pitfall 1).

- [ ] **Does this PR add `os.ReadFile` of schema content at runtime** (or any non-`embed.FS` schema loading)?
      Answer: **No** (if Yes, this PR will be closed; see CONTRIBUTING.md §2 and PITFALLS.md Pitfall 2).

## Tests

<!-- What tests were added / what was verified manually. -->
```

### Pattern 4: CODEOWNERS pinning sensitive paths

**What:** `.github/CODEOWNERS` (or `CODEOWNERS` at root) maps file patterns to GitHub users/teams who become required reviewers for changes to those paths (when combined with branch protection).

**Important:** **Last matching pattern wins** ([GitHub docs — About code owners](https://docs.github.com/en/repositories/managing-your-repositorys-settings-and-features/customizing-your-repository/about-code-owners)). Order matters.

**Example:**
```
# Source: GitHub CODEOWNERS docs
# https://docs.github.com/en/repositories/managing-your-repositorys-settings-and-features/customizing-your-repository/about-code-owners

# Default: nobody (no auto-assignment for unmatched files)
# (Empty line means no owner; that's intentional — most files don't need codeowner gating)

# Architectural rules: any change to CONTRIBUTING.md must be reviewed by repo owner
/CONTRIBUTING.md @imulab

# CI / repo-meta: any change to .github/ must be reviewed
/.github/ @imulab

# Runtime support library core: any change to rt/ requires review
/rt/ @imulab

# Generator core: any change to gen/ requires review
/gen/ @imulab

# Project decisions: PROJECT.md is the single source of truth for Key Decisions
/PROJECT.md @imulab
```

Branch protection (configured in GitHub UI, deferred per CONTEXT.md `<deferred>`) enforces that CODEOWNERS reviews are required.

### Pattern 5: lefthook commit-msg + pre-commit configuration

**What:** A single `lefthook.yml` at repo root configures both pre-commit hooks (e.g., grep for forbidden patterns in staged files) and commit-msg hooks (e.g., reject vendor names in commit messages).

**Example:**
```yaml
# Source: lefthook docs
# https://lefthook.dev/configuration/

# lefthook.yml at repo root

pre-commit:
  parallel: true
  commands:
    forbidden-symbols:
      glob: "**/*.go"
      run: |
        if echo {staged_files} | xargs -r grep -lEn 'os\.ReadFile.*schema|interface\s*\{\s*Path\(|Property\s+interface'; then
          echo "Forbidden symbol detected. See CONTRIBUTING.md."
          exit 1
        fi

commit-msg:
  commands:
    vendor-names:
      run: |
        if grep -iE 'Entra|Okta|Azure\s*AD|AzureAD' {1}; then
          echo "Commit message contains vendor name. See CONTRIBUTING.md §3."
          exit 1
        fi
```

`{1}` is the path to the commit message file (lefthook passes it positionally for `commit-msg` hooks). `{staged_files}` expands to the list of staged paths matching the glob.

### Anti-Patterns to Avoid

- **Committing `go.work`:** Per official Go reference, "It is generally inadvisable to commit go.work files into version control systems." A checked-in `go.work` masks the production-like build path (`GOWORK=off`) from CI and contributors. (REQ-REPO-01 already mandates `go.work` gitignored — this is the **why**.)
- **`replace` directives in `go.mod`:** They leak to consumers. All local-path overrides live in `go.work` (which is gitignored). Per the [Go module reference](https://go.dev/ref/mod#workspaces): wildcard `replace` in `go.work` overrides version-specific `replace` in `go.mod`.
- **Single-job CI without `GOWORK=off`:** Workspace mode masks missing `require` lines. The two-job parallel pattern (Pattern 1) is the only reliable way to catch this class of bug.
- **Using `html/template`:** It HTML-escapes identifiers. Use `text/template` for source generation. Mistake is easy because both packages share an import name (`template`).
- **Skipping `go/format.Source` post-processing:** A stray space or unbalanced template control structure produces invalid Go. `go/format.Source` returns an error pointing to the offending line. PITFALLS.md Pitfall 9 documents this.
- **CODEOWNERS pattern ordering:** Last matching pattern wins. Putting a broad `@team` line before a narrow `@person` line means the broad pattern overrides for files matched by both. Order specific-to-general → general-to-specific reading is wrong; correct order is **most-specific last**.
- **PR-template-only enforcement:** Checkboxes are honor-system. They must be paired with CI grep gates (Pattern 1's `forbidden-symbols` job) and CODEOWNERS to actually block bad PRs.
- **Pre-commit hooks without a CI mirror:** Hooks can be bypassed with `git commit --no-verify`. The CI workflow MUST re-run the same checks (commit-msg gate as a CI job mirroring the lefthook commit-msg hook) so the hook is a developer-experience optimization, not the security boundary.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Go AST emission with conditional imports | A bespoke `bytes.Buffer` writer that tracks imports manually via a side `map[string]struct{}` | `dave/jennifer` `Qual()` (auto-tracks imports) **OR** `text/template` + `imports.Process` (post-processes imports) | Hand-rolled import tracking is a known foot-gun. The legacy `pkg/v2` codebase had several places where attribute additions caused import drift; this is the second-most-cited cause of "regen produces noisy diff" in the [Go-codegen blog ecosystem](https://eli.thegreenplace.net/2021/a-comprehensive-guide-to-go-generate). |
| Post-emission Go formatting | A custom AST-walk reformatter | `go/format.Source` (stdlib, free) | `go/format.Source` is the in-process equivalent of `gofmt`. It also returns errors on invalid Go, which is invaluable for template debugging. |
| Goimports as a subprocess | `exec.Command("goimports", ...)` shelling out | `golang.org/x/tools/imports.Process` (library form) | Subprocess adds binary-installation requirements for every contributor. The library form is in-process and stable. |
| Git hook management | Hand-rolled bash in `.git/hooks/*` with a `setup.sh` install script | `lefthook` (single Go binary, declarative `lefthook.yml`) | Hooks are per-developer state. lefthook handles the install dance, parallel execution, and `staged_files` expansion. Two of the four enforcement levers (pre-commit forbidden-symbol grep, commit-msg vendor-name gate) live in the same config file. |
| Workspace recreation in CI | A custom shell script that parses `go.mod` files and writes `go.work` | Two trivial commands: `go work init && go work use ./gen ./rt` | `go work use` is officially supported, idempotent, and survives module additions. Hand-rolling is a maintenance liability. |
| GitHub Actions Go installation | A `curl`-the-tarball step | `actions/setup-go@v6` | It manages the tool cache, version aliases (`stable`, `oldstable`), and falls back to the official Go distribution. |
| GitHub PR template enforcement | A custom GitHub App | Combination of: PR template (Markdown) + CODEOWNERS + branch protection (UI-configured, deferred) | GitHub already supports all three primitives; combining them yields a self-enforcing sequence ([DevToolbox PR guide](https://devtoolbox.dedyn.io/blog/github-pull-requests-complete-guide)). |

**Key insight:** Phase 0 is "configure existing tools" not "write code." Every "Don't Hand-Roll" entry above maps to a Phase 0 task that should be a configuration line in YAML/Go-mod, not a function in Go. The temptation to write a tiny shell script for X is the trap — the maintained tool already handles X plus three edge cases you haven't considered.

## Common Pitfalls

### Pitfall 1: Spike concludes before evidence is captured

**What goes wrong:** The team writes both spike implementations, eyeballs them, declares "Jennifer wins" or "templates are simpler," and deletes one before any objective comparison is recorded. Six weeks later in Phase 2, someone proposes switching engines based on a new gut feeling and there's no audit trail to push back.

**Why it happens:** The spike feels like a tax on getting to "real code." The decision is small enough to make on intuition, large enough to regret.

**How to avoid:**
- **Both spike implementations stay in-repo permanently** (per CONTEXT.md). They are the audit trail.
- The decision document in `PROJECT.md` Key Decisions cites the spike directory by path **and** records the four scoring criteria with concrete observations (e.g., "Jennifer: 0 lines of import bookkeeping; templates: 12 lines including a TODO for handling `errors.Is` aliasing").
- Two regen runs of the chosen winner produce byte-identical Go (REQ-GEN-06 dry run); evidence is a `git diff --exit-code` showing zero output.

**Warning signs:**
- The spike `README.md` says "we picked Jennifer" without citing the criteria.
- One spike implementation is `git rm`'d before the decision lands in PROJECT.md.

### Pitfall 2: `go.work` accidentally committed

**What goes wrong:** A contributor `git add .`s and commits `go.work`. Now every CI run uses workspace mode, the `GOWORK=off` job is masked (because `go.work` overrides `GOWORK=off`... no wait, `GOWORK=off` overrides `go.work` per the official reference, but the symptom is still that contributors get inconsistent local builds).

**Why it happens:** `go.work` is auto-created by `go work init`; `git status` lists it; muscle memory adds it.

**How to avoid:**
- `.gitignore` lists `go.work` AND `go.work.sum` from the first commit ([per Go module reference](https://go.dev/ref/mod#workspaces) — `go.work.sum` is automatically maintained per-developer).
- Pre-commit hook (lefthook) refuses to stage `go.work` or `go.work.sum`:
  ```yaml
  pre-commit:
    commands:
      no-go-work:
        run: |
          if echo {staged_files} | grep -qE '(^|/)go\.work(\.sum)?$'; then
            echo "Refusing to commit go.work / go.work.sum (per .gitignore convention)."
            exit 1
          fi
  ```
- README.md explicitly says: "If you cloned this repo and `go build ./...` complains about missing modules, run `go work init && go work use ./gen ./rt`."

**Warning signs:**
- `git ls-files | grep go.work` returns hits.
- A new contributor reports "build worked when I had it but not after I cleaned up."

### Pitfall 3: CI grep gates have false negatives

**What goes wrong:** The `forbidden-symbols` CI job greps for `os\.ReadFile.*schema` but a contributor writes `os.ReadFile(schemaPath)` where `schemaPath` is a variable. The pattern doesn't match the variable name, the relapse sneaks in, the gate looks like it works but doesn't.

**Why it happens:** Regex grep is line-oriented and identifier-blind. It catches obvious cases, misses indirected ones.

**How to avoid:**
- Make the regex broad: `os\.ReadFile.*[Ss]chema|os\.ReadFile.*\.json` catches both literal-path and variable-name patterns when the variable name contains "schema" or the file extension is `.json`.
- Pair the grep gate with a CONTRIBUTING.md rule that says "any new use of `os.ReadFile` in `gen/` or `rt/` requires repo-owner approval (CODEOWNERS-pinned)." This makes the human-in-the-loop the second line of defense.
- Add a positive-coverage test in Phase 1+ that spins up a generated server in a chroot/container with no schema files on disk; if it boots, embedded schemas worked.

**Warning signs:**
- Grep gate hasn't fired in months (could mean nobody's tried, could mean nobody's noticed when they did).
- A contributor's PR adds a helper function `loadFromDisk(path)` and the grep doesn't see the indirection.

### Pitfall 4: Commit-msg linter bypassed via `--no-verify`

**What goes wrong:** A contributor commits with vendor names in the message using `git commit --no-verify`. The lefthook hook is silent because it was bypassed; the commit lands.

**Why it happens:** `--no-verify` is sometimes legitimately needed (e.g., the hook itself is broken). Contributors learn it as a "make hook errors go away" reflex.

**How to avoid:**
- **CI mirror is the security boundary.** The GitHub Actions `commit-msg-lint` job (Pattern 1 above) re-runs the same regex on every commit in the PR's commit range. The pre-commit hook is a developer-experience optimization, not the gate.
- `--no-verify` is fine on local commits; the CI mirror catches the message before merge.
- README.md explicitly says: "The pre-commit hook can be bypassed with `--no-verify`. Don't bother — the same checks run in CI on PR open."

**Warning signs:**
- A PR's commit history shows a commit with a vendor name; CI didn't fail (the mirror is misconfigured).
- A contributor reports "I `--no-verify`'d to get past a hook bug" — the bug should be fixed, not normalized.

### Pitfall 5: CODEOWNERS misconfiguration silently disables the gate

**What goes wrong:** CODEOWNERS lists `@imulab` for `/rt/` but `@imulab` is not a member of the repo with write access (e.g., team renamed, account deleted, or the YAML has a typo like `@iamulab`). GitHub silently skips the auto-assignment, no review is required, the gate is dead.

**Why it happens:** Per [GitHub CODEOWNERS docs](https://docs.github.com/en/repositories/managing-your-repositorys-settings-and-features/customizing-your-repository/about-code-owners): "If you specify a user or team that doesn't exist or has insufficient access, a code owner will not be assigned."

**How to avoid:**
- After creating CODEOWNERS, **navigate to the file in the GitHub UI** — invalid lines are highlighted with a red marker.
- A list of errors is also accessible via the GitHub API (`GET /repos/{owner}/{repo}/codeowners/errors`).
- For Phase 0: only one owner (`@imulab`); the risk is small. For later phases, periodic CODEOWNERS audits.

**Warning signs:**
- CODEOWNERS file in the GitHub UI shows red squiggles.
- PRs to `/rt/` paths don't auto-request the expected reviewer.

### Pitfall 6: Spike scope creep into "real" generator code

**What goes wrong:** The spike for "User struct + Validate()" grows tendrils — "while we're here, let's emit the JSON marshaller too" — and now the spike is half a generator and never converges.

**Why it happens:** The spike artifact's choice (typed User + Validate) is deliberately representative; from there, every additional emitted artifact feels small.

**How to avoid:**
- The spike artifact is **frozen** at: typed `User` struct + `Validate()` method exercising conditional imports (`time`, `regexp`, `errors`), nested types, and branching validation (per CONTEXT.md). Anything beyond that is Phase 2's `gen/internal/emit/`.
- Both spike implementations have a hard LOC budget the reviewer enforces (e.g., < 500 lines each).
- The spike `README.md` ends with "if you're tempted to add X, X belongs in Phase 2."

**Warning signs:**
- Spike grows past the artifact contract (e.g., starts emitting `*_repo.go` for SQL).
- Spike implementations diverge in what they emit (one has the validator, the other has marshalling) — apples-to-oranges defeats the comparison.

### Pitfall 7: Contributor environment drift on the Go version floor

**What goes wrong:** The user runs Go 1.26.1 locally and writes idiomatic 1.26 code (e.g., uses a stdlib feature added in 1.26). CI passes the 1.26 matrix job, fails the 1.25 job. The contributor doesn't have 1.25 installed, can't reproduce locally.

**Why it happens:** `go.mod`'s `go 1.25` directive is a **minimum version**, not a maximum. Code that compiles with 1.26 may break on 1.25 if it uses 1.26-only stdlib calls.

**How to avoid:**
- Both `gen/go.mod` and `rt/go.mod` declare `go 1.25` (the floor per CONTEXT.md).
- CI matrix includes 1.25 explicitly (per CONTEXT.md, matrix is `1.25 + 1.26`).
- README.md "Contributing" section says: "If your code compiles locally but fails CI on 1.25, install 1.25 via `go install golang.org/dl/go1.25@latest && go1.25 download` to reproduce."
- Optional: add a `gotoolchain` directive to pin (vs. auto-upgrade); see [go.dev — toolchains](https://go.dev/doc/toolchain) — Phase 0 doesn't need this complication.

**Warning signs:**
- 1.25 CI job fails on a feature only 1.26 has.
- Contributors complain about "weird CI failures I can't reproduce."

## Code Examples

Verified patterns from official sources. These are templates the planner can lift into Phase 0 task descriptions.

### Example 1: Workspace creation (one-time setup, repeated in CI)

```bash
# Source: https://go.dev/doc/tutorial/workspaces
# https://go.dev/ref/mod#workspaces

# At repo root, after gen/ and rt/ have go.mod files:
go work init
go work use ./gen ./rt

# Output: go.work file with:
#   go 1.25
#   use (
#       ./gen
#       ./rt
#   )

# This file is GITIGNORED per REQ-REPO-01.

# To verify:
go env GOWORK
# /Users/.../go-scim/go.work    (workspace mode)

GOWORK=off go env GOWORK
# off                            (single-module mode)
```

### Example 2: Minimal `gen/go.mod` and `rt/go.mod`

```go
// Source: https://go.dev/ref/mod#go-mod-file
// gen/go.mod
module github.com/imulab/go-scim/gen

go 1.25
```

```go
// rt/go.mod
module github.com/imulab/go-scim/rt

go 1.25
```

(No `require` lines yet — Phase 0 has no source files.)

### Example 3: `text/template` + `go/format.Source` skeleton (spike candidate B)

```go
// Source: https://pkg.go.dev/text/template
// https://pkg.go.dev/go/format

package main

import (
	"bytes"
	"fmt"
	"go/format"
	"os"
	"text/template"
)

const userTmpl = `// Code generated by spike/template. DO NOT EDIT.

package domain

{{- if .HasTime }}
import "time"
{{- end }}

type User struct {
	ID       string ` + "`json:\"id\"`" + `
	UserName string ` + "`json:\"userName\"`" + `
	{{- if .HasTime }}
	Created  time.Time ` + "`json:\"created\"`" + `
	{{- end }}
}

func (u *User) Validate() error {
	if u.UserName == "" {
		return fmt.Errorf("userName is required")
	}
	return nil
}
`

type userData struct {
	HasTime bool
}

func main() {
	tmpl := template.Must(template.New("user").Parse(userTmpl))
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, userData{HasTime: true}); err != nil {
		panic(err)
	}
	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		// format.Source returns the offending line — invaluable for template debugging.
		fmt.Fprintf(os.Stderr, "format error: %v\nraw output:\n%s\n", err, buf.String())
		os.Exit(1)
	}
	if err := os.WriteFile("output/user_gen.go", formatted, 0644); err != nil {
		panic(err)
	}
}
```

**Note:** The template hand-writes the `import "time"` block. With more conditional imports (e.g., `regexp`, `errors`, third-party), this block gets gnarly fast. Hence Pattern 3's option to swap `format.Source` for `imports.Process`:

```go
// Source: https://pkg.go.dev/golang.org/x/tools/imports
import "golang.org/x/tools/imports"

// Replaces format.Source(buf.Bytes()) with:
formatted, err := imports.Process("output/user_gen.go", buf.Bytes(), nil)
// imports.Process formats AND adds/removes imports automatically.
```

### Example 4: `dave/jennifer` skeleton (spike candidate A)

```go
// Source: https://pkg.go.dev/github.com/dave/jennifer/jen
// https://github.com/dave/jennifer

package main

import (
	"github.com/dave/jennifer/jen"
)

func main() {
	f := jen.NewFile("domain")
	f.HeaderComment("Code generated by spike/jennifer. DO NOT EDIT.")

	hasTime := true

	// type User struct { ... }
	fields := []jen.Code{
		jen.Id("ID").String().Tag(map[string]string{"json": "id"}),
		jen.Id("UserName").String().Tag(map[string]string{"json": "userName"}),
	}
	if hasTime {
		// jen.Qual handles the import for "time" automatically.
		fields = append(fields, jen.Id("Created").Qual("time", "Time").Tag(map[string]string{"json": "created"}))
	}
	f.Type().Id("User").Struct(fields...)

	// func (u *User) Validate() error { ... }
	f.Func().
		Params(jen.Id("u").Op("*").Id("User")).
		Id("Validate").Params().
		Error().
		Block(
			jen.If(jen.Id("u").Dot("UserName").Op("==").Lit("")).Block(
				// jen.Qual handles "fmt" import.
				jen.Return(jen.Qual("fmt", "Errorf").Call(jen.Lit("userName is required"))),
			),
			jen.Return(jen.Nil()),
		)

	// f.Save calls f.Render which runs go/format internally.
	if err := f.Save("output/user_gen.go"); err != nil {
		panic(err)
	}
}
```

**Note:** Jennifer's `Qual(path, name)` automatically adds `import "time"` and `import "fmt"` to the file's import block. No conditional `{{- if .HasTime }}import "time"{{- end }}` ceremony. This is the key win the spike must measure.

### Example 5: `.gitignore` for the workspace + spike state

```gitignore
# Source: https://go.dev/ref/mod#workspaces — go.work / go.work.sum advisory

# Go workspace state — per-developer, never committed.
go.work
go.work.sum

# Build artifacts
*.exe
*.test
/bin/

# Editor / OS
.DS_Store
.idea/
.vscode/
```

### Example 6: lefthook configuration

```yaml
# Source: https://github.com/evilmartians/lefthook
# https://lefthook.dev/configuration/
# Place at repo root as lefthook.yml.

pre-commit:
  parallel: true
  commands:
    no-go-work:
      run: |
        if echo {staged_files} | tr ' ' '\n' | grep -qE '(^|/)go\.work(\.sum)?$'; then
          echo "Refusing to commit go.work / go.work.sum."
          exit 1
        fi

    forbidden-symbols-source:
      glob: "{gen,rt}/**/*.go"
      run: |
        if echo {staged_files} | xargs -r grep -lEn 'os\.ReadFile.*[Ss]chema|interface\s*\{\s*Path\(|Property\s+interface'; then
          echo "Forbidden symbol detected. See CONTRIBUTING.md."
          exit 1
        fi

    forbidden-vendor-names-source:
      glob: "{gen,rt}/**/*.go"
      run: |
        if echo {staged_files} | xargs -r grep -liE 'Entra|Okta|Azure\s*AD|AzureAD'; then
          echo "Vendor name in source. See CONTRIBUTING.md §3."
          exit 1
        fi

commit-msg:
  commands:
    vendor-names:
      run: |
        if grep -iE 'Entra|Okta|Azure\s*AD|AzureAD' {1}; then
          echo "Commit message contains vendor name. See CONTRIBUTING.md §3."
          exit 1
        fi
```

Install: `lefthook install` after `lefthook.yml` lands.

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| Single-module Go repo with subpackages | Multi-module workspace via `go.work` | Go 1.18 (2022) introduced `go.work`; widely adopted by 2024 | Cleaner separation of generator vs. runtime; each module versions independently |
| `replace` directives in `go.mod` for local-path overrides | `replace` directives in `go.work` only | Go 1.18+ | `go.mod replace` leaks to consumers; `go.work replace` is local-only |
| `goimports` as a subprocess | `golang.org/x/tools/imports.Process` library call | Go modules era | In-process, no binary install needed |
| Hand-rolled bash hooks in `.git/hooks/` + setup script | `lefthook` (Go binary) declarative `lefthook.yml` | ~2022–2024 became dominant for non-Node projects | Single config file, parallel exec, language-agnostic |
| `actions/setup-go@v3` / `@v4` / `@v5` | `actions/setup-go@v6` | 2025–2026 | Better tool-cache behavior, `stable`/`oldstable` aliases, `go-version-file` support |
| Go 1.13 (legacy `.legacy/go.mod`) | Go 1.25 floor (CI matrix 1.25 + 1.26) | Per CONTEXT.md override of STACK.md | Generics (1.18+), `slices`/`maps` packages (1.21+), `go.work` (1.18+), iter (1.23+), all available |

**Deprecated/outdated:**
- `io/ioutil` — replaced by `io` and `os` packages (deprecated since Go 1.16). Per [Pitfall 4.2 in CONCERNS.md](file://./.planning/codebase/CONCERNS.md), legacy code uses it; new code (Phase 0+) MUST NOT.
- `mattn/go-sqlite3` as the default — `modernc.org/sqlite` is the v1 reference per Pitfall 17 of PITFALLS.md (CGO-free wins for v1 correctness focus). Not Phase 0's concern, but the planner should know.
- `goimports` as a separate binary on `PATH` — library form (`imports.Process`) is the modern path.

## Open Questions

1. **`text/template` + `imports.Process` as a third spike option?**
   - What we know: `imports.Process` (library form of `goimports`) handles auto-import management on top of `text/template` output. This compromises between "templates feel readable" and "imports get tracked automatically."
   - What's unclear: CONTEXT.md explicitly defines the spike as `text/template` + `go/format` vs `dave/jennifer` — adding a third leg is scope creep for the spike, but ignoring `imports.Process` may make the spike's "import management" criterion unfairly favor Jennifer.
   - Recommendation: The planner should **document `imports.Process` as a fallback option in the spike `README.md` and the `PROJECT.md` Key Decisions entry**, even if the spike itself only builds the two CONTEXT.md-specified implementations. If Jennifer wins on import management but loses on readability, `imports.Process` becomes the natural compromise to evaluate in Phase 2 if Jennifer's API ergonomics turn out worse than expected.

2. **Should `spike/` itself have a `go.work` (excluded from the main one)?**
   - What we know: `spike/template/` and `spike/jennifer/` are separate modules with separate `go.mod` files, neither in the main `go.work`.
   - What's unclear: Whether to wire them together in `spike/go.work` for spike-internal convenience, or build each independently with `cd spike/X && go build ./...`.
   - Recommendation: **Don't add a `spike/go.work`**. The spikes are deliberately independent (apples-to-apples comparison demands no shared deps); a separate workspace adds nothing and confuses contributors who already learned the main workspace pattern. Each spike is built in isolation: `cd spike/template && go build ./...`.

3. **Where does the spike runner live?**
   - What we know: Both spike implementations need to be runnable so the artifact comparison happens.
   - What's unclear: Add `make spike` / `just spike` target? Add a CI job? Or document manual `cd spike/X && go run .` commands?
   - Recommendation: **Add `make spike` (or `just spike`) target that runs both, diffs the outputs, and exits non-zero on diff.** This makes the byte-identical-after-`gofmt` criterion (REQ-GEN-06 dry run) executable. Optional: a CI job that runs `make spike` to keep the comparison alive even if the spike directories aren't otherwise built.

4. **Vendor-name regex false positives in CI?**
   - What we know: The CI grep gate fails on `Entra|Okta|Azure\s*AD|AzureAD` in source files. PITFALLS.md, CONTRIBUTING.md, and `.planning/` docs cite these names legitimately.
   - What's unclear: How does the gate avoid false-positive on the docs that are SUPPOSED to mention these names?
   - Recommendation: **Scope the grep with explicit excludes**:
     ```bash
     git grep -iE 'Entra|Okta|Azure\s*AD|AzureAD' -- \
       ':(exclude).planning/' \
       ':(exclude).legacy/' \
       ':(exclude)CONTRIBUTING.md' \
       ':(exclude)PROJECT.md' \
       ':(exclude)README.md'
     ```
     The exclusion list is committed in `.github/scripts/forbidden-symbols.sh`; CONTRIBUTING.md documents the rationale ("PITFALLS.md / CONTRIBUTING.md may name vendors as warnings; source code may not").

5. **Phase 0 success-criterion #4 says "every module" — what counts as a module?**
   - What we know: Phase 0 produces `gen/` and `rt/` modules. `spike/template/` and `spike/jennifer/` also have `go.mod` files but are excluded from `go.work`.
   - What's unclear: Should the `GOWORK=off` build job iterate over `gen, rt, spike/template, spike/jennifer`, or only `gen, rt`?
   - Recommendation: **Phase 0 CI builds `gen` and `rt` only.** Spike modules are run by `make spike` (or equivalent) when the spike is being evaluated, not on every CI run — they don't change after the decision lands and would just be CI noise. Document this in `spike/README.md`.

## Sources

### Primary (HIGH confidence)
- [Go module reference — Workspaces](https://go.dev/ref/mod#workspaces) — `GOWORK=off` semantics, `go.work` precedence, `go.work.sum` non-tracking advisory
- [Go tutorial — Multi-module workspaces](https://go.dev/doc/tutorial/workspaces) — `go work init` / `go work use` examples
- [Go 1.26 release blog](https://go.dev/blog/go1.26) — release date 2026-02-10, current stable
- [pkg.go.dev — text/template](https://pkg.go.dev/text/template) — template syntax, `Execute` semantics
- [pkg.go.dev — go/format](https://pkg.go.dev/go/format) — `format.Source` API
- [pkg.go.dev — golang.org/x/tools/imports](https://pkg.go.dev/golang.org/x/tools/imports) — `imports.Process` API and `Options` struct
- [pkg.go.dev — github.com/dave/jennifer/jen](https://pkg.go.dev/github.com/dave/jennifer/jen) — v1.7.1, Sept 2024, MIT, automatic imports via `Qual()`
- [GitHub — actions/setup-go README](https://github.com/actions/setup-go) — v6 matrix syntax, Go version aliasing
- [GitHub Docs — About code owners](https://docs.github.com/en/repositories/managing-your-repositorys-settings-and-features/customizing-your-repository/about-code-owners) — file location precedence, last-pattern-wins, error reporting
- [GitHub Docs — Managing and standardizing pull requests](https://docs.github.com/en/pull-requests/collaborating-with-pull-requests/getting-started/managing-and-standardizing-pull-requests) — `pull_request_template.md` location and behavior
- [evilmartians/lefthook README](https://github.com/evilmartians/lefthook) — installation, `lefthook.yml` syntax, parallel execution
- `.planning/research/PITFALLS.md` — Pitfalls 1, 2, 3 (anti-relapse rules), 9 (gofmt), 18 (workspace footguns)
- `.planning/codebase/CONCERNS.md` — three rewrite drivers; legacy artifact paths cited in CONTRIBUTING.md
- `.planning/REQUIREMENTS.md` — REPO-01..05 wording (incl. CONTEXT.md overrides)

### Secondary (MEDIUM confidence)
- [oneuptime — How to Use Go Workspaces for Monorepos (2026-02-01)](https://oneuptime.com/blog/post/2026-02-01-go-workspaces-monorepos/view) — `GOWORK=off` CI pattern advice (corroborates Go reference)
- [oneuptime — Multi-Module Go Projects with Workspaces (2026-01-25)](https://oneuptime.com/blog/post/2026-01-25-multi-module-go-projects-workspaces/view) — module organisation patterns
- [Eli Bendersky — A comprehensive guide to go generate](https://eli.thegreenplace.net/2021/a-comprehensive-guide-to-go-generate) — text/template + DO-NOT-EDIT marker pattern, `goimports` post-processing
- [pkg.go.dev — github.com/dolmen-go/codegen](https://pkg.go.dev/github.com/dolmen-go/codegen) — text/template wrapper enforcing gofmt + DO-NOT-EDIT marker (alternative to hand-rolling)
- [DEV — Metaprogramming with Go (Jennifer + go/types)](https://dev.to/hlubek/metaprogramming-with-go-or-how-to-build-code-generators-that-parse-go-code-2k3j) — Jennifer usage with `go:generate`
- [DevToolbox — GitHub Pull Requests Complete Guide](https://devtoolbox.dedyn.io/blog/github-pull-requests-complete-guide) — PR template + CODEOWNERS + branch protection composition
- [PullNotifier — Pull request template github](https://blog.pullnotifier.com/blog/pull-request-template-github-a-guide-to-github-prs) — template structure conventions
- [Pockit — GitHub Actions in 2026: Monorepo CI/CD](https://pockit.tools/blog/github-actions-monorepo-runners-guide-2026/) — monorepo CI patterns
- [Boot.dev — Format on Save in Go](https://www.boot.dev/blog/golang/format-on-save-vs-code-golang/) — gofmt vs goimports tradeoff

### Tertiary (LOW confidence — flagged for verification if used)
- [hyr.mn — Formatting Go code with goimports](https://hyr.mn/gofmt/) — single source on `goimports` / `goreturns` toolchain split (corroborated by Boot.dev so not strictly LOW but only one independent voice)
- [pkgpulse — husky vs lefthook vs lint-staged 2026](https://www.pkgpulse.com/blog/husky-vs-lefthook-vs-lint-staged-git-hooks-nodejs-2026) — feature comparison (commercial blog, biased toward Node projects but cross-checks with lefthook docs)

## Metadata

**Confidence breakdown:**
- Standard Stack: **HIGH** — every library/tool verified against pkg.go.dev or the official GitHub README; versions current as of 2026-05-07
- Architecture Patterns: **HIGH** — workflows derived from Go module reference + actions/setup-go README; CODEOWNERS pattern from official GitHub Docs; lefthook example from official lefthook.dev docs
- Common Pitfalls: **HIGH** — every pitfall traces to either an official-docs warning, a documented Go ecosystem behavior, or a concrete recommendation in PITFALLS.md (which itself was researched in the previous wave)
- Code Examples: **HIGH** — all snippets are minimal-modification adaptations of official-doc examples; cited inline

**Verification protocol followed:**
- Negative claims verified ("don't commit go.work" — Go reference; "lefthook bypassable via --no-verify" — git man page convention; "CODEOWNERS silently skips invalid users" — GitHub Docs)
- Multiple sources cross-referenced for every Standard Stack entry (pkg.go.dev + project GitHub + at least one independent blog or guide)
- Publication dates checked (Go 1.26 release Feb 2026; Jennifer v1.7.1 Sept 2024 — flagged as "stable but not aggressively updated", which is itself the relevant data point for the spike)

**Research date:** 2026-05-07
**Valid until:** ~2026-08-07 (90 days; Phase 0 deliverables are configuration files against well-stabilised tools — Go workspace semantics, GitHub features, lefthook config syntax — none of which churn rapidly. Re-verify only if Go 1.27 ships or actions/setup-go releases v7.)

---
phase: 00-architecture-spike-repo-setup
plan: 02
subsystem: infra
tags: [go-modules, go-workspace, github-actions, makefile, ci, license, gitignore]

# Dependency graph
requires:
  - phase: 00-architecture-spike-repo-setup
    provides: nothing (Plan 02 is wave-1 — runs in parallel with Plan 01; Plan 01's lefthook.yml lands the local-hook side of the four enforcement levers, Plan 02 lands the CI mirror side)
provides:
  - Two-module Go workspace (gen + rt) with module paths github.com/imulab/go-scim/{gen,rt} and go 1.25 directive
  - go.work + go.work.sum gitignored (REPO-01)
  - LICENSE at repo root (MIT, byte-identical to .legacy/LICENSE) (REPO-03)
  - README.md with build instructions for fresh-clone contributors
  - .github/workflows/ci.yml with four jobs (workspace build, isolated build, forbidden-symbol grep, commit-msg vendor-name lint)
  - .github/scripts/forbidden-symbols.sh — single source of truth for the four forbidden-symbol regexes
  - Makefile with ci-local target reproducing CI without GitHub
affects: [01-Definition-IR-Validator, all-future-phases (build infrastructure foundational)]

# Tech tracking
tech-stack:
  added: [go workspaces (go.work), GitHub Actions (actions/checkout@v4, actions/setup-go@v6), GNU Make]
  patterns:
    - "Two-build CI gate: every module compiled both with go.work (workspace mode) and with GOWORK=off (isolated mode) on each PR — workspace cannot mask a missing require line"
    - "Forbidden-symbol grep gate as a separately-versioned shell script invoked from CI AND from the local hooks (single source of truth for the regex set)"
    - "ci-local Makefile target mirrors CI exactly so contributors can reproduce CI without push"
    - "Workspace build uses explicit module-tree enumeration (go build ./gen/... ./rt/...) because repo root is not itself a module"

key-files:
  created:
    - gen/go.mod
    - gen/doc.go
    - rt/go.mod
    - rt/doc.go
    - .gitignore
    - LICENSE
    - README.md
    - .github/workflows/ci.yml
    - .github/scripts/forbidden-symbols.sh
    - Makefile
  modified: []

key-decisions:
  - "Module paths are github.com/imulab/go-scim/{gen,rt} (CONTEXT.md override of REPO-02 wording — Plan 04 reconciles REQUIREMENTS.md text)"
  - "go directive is 1.25 in both modules; CI matrix is ['1.25', '1.26']"
  - "Each module ships a placeholder doc.go containing only a package declaration so go build matches a package — empty modules cause `go build ./...` to error under a workspace"
  - "Workspace build invocation is `go build ./gen/... ./rt/...` not `go build ./...`; the latter errors with `pattern ./...: directory prefix . does not contain modules` when run from a non-module repo root"

patterns-established:
  - "Repo-root non-module + multi-module workspace: build commands enumerate module trees explicitly"
  - "CI mirrors lefthook hooks via shared regex set — same script surface for both, so a regex tweak ripples to both gates atomically"

requirements-completed: [REPO-01, REPO-02, REPO-03]

# Metrics
duration: 6min
completed: 2026-05-08
---

# Phase 0 Plan 02: gen/rt modules + GitHub Actions CI Summary

**Two-module Go workspace (gen + rt, go 1.25) with gitignored go.work, MIT LICENSE, README, and a four-job GitHub Actions CI (workspace + isolated builds on Go 1.25/1.26 matrix, plus forbidden-symbol and commit-msg vendor-name grep gates).**

## Performance

- **Duration:** 6 min
- **Started:** 2026-05-08T03:32:00Z
- **Completed:** 2026-05-08T03:38:30Z
- **Tasks:** 3
- **Files created:** 10

## Accomplishments

- Two empty Go modules (`gen` + `rt`) with the agreed module paths and `go 1.25` directive — both build under `go work` AND with `GOWORK=off` in isolation.
- `.gitignore` excludes `go.work` and `go.work.sum` so the workspace state never enters git history (REPO-01); `git ls-files | grep go.work` returns zero results.
- `LICENSE` at repo root is byte-identical to `.legacy/LICENSE` (REPO-03); `cmp` exits 0.
- `README.md` documents the fresh-clone build sequence (workspace recreation + isolated build) — verified correct against actual Go behavior.
- `.github/workflows/ci.yml` with four jobs: `build-workspace` (Go 1.25 + 1.26), `build-isolated` (Go 1.25 + 1.26 × {gen, rt}), `forbidden-symbols` (calls the extracted script), `commit-msg-lint` (PR-only, mirrors lefthook commit-msg regex).
- `.github/scripts/forbidden-symbols.sh` extracted as a single source of truth for the four regexes — exits 0 on the current tree; CI and `make forbidden-symbols` both call it.
- `Makefile` with `help`, `workspace`, `build-workspace`, `build-isolated`, `forbidden-symbols`, `ci-local`, `spike`, `clean` targets. `make ci-local` reproduces CI locally and succeeds.

## Task Commits

Each task was committed atomically:

1. **Task 1: Stand up two-module workspace, .gitignore, LICENSE, README** — `f41da48` (feat)
2. **Task 2: Write GitHub Actions CI workflow + forbidden-symbols script** — `904742f` (feat)
3. **Task 3: Write Makefile with ci-local and spike targets** — `79a45ea` (feat)

(See "Issues Encountered" for a note on Task 3's commit accidentally absorbing other plans' staged files due to a parallel-wave race.)

## Files Created/Modified

- `gen/go.mod` — `module github.com/imulab/go-scim/gen`, `go 1.25`.
- `gen/doc.go` — placeholder package declaration so `go build` has at least one package to compile in Phase 0.
- `rt/go.mod` — `module github.com/imulab/go-scim/rt`, `go 1.25`.
- `rt/doc.go` — placeholder package declaration (same rationale as gen/doc.go).
- `.gitignore` — excludes `go.work`, `go.work.sum`, `*.exe`, `*.test`, `/bin/`, `.DS_Store`, `.idea/`, `.vscode/`. Spike directory deliberately NOT ignored.
- `LICENSE` — MIT, copied byte-for-byte from `.legacy/LICENSE`.
- `README.md` — project tagline, status, modules description, fresh-clone build steps, contributing pointer, license link.
- `.github/workflows/ci.yml` — four CI jobs as described above; uses `actions/checkout@v4` and `actions/setup-go@v6` on Go 1.25 + 1.26.
- `.github/scripts/forbidden-symbols.sh` — executable shell script implementing the forbidden-symbol grep gate.
- `Makefile` — `help`, `workspace`, `build-workspace`, `build-isolated`, `forbidden-symbols`, `ci-local`, `spike`, `clean` targets with TAB indentation.

## Decisions Made

- **Module paths use the CONTEXT.md override of REPO-02:** `github.com/imulab/go-scim/{gen,rt}` (short names, no `scim` repetition). REQUIREMENTS.md text reconciliation is Plan 04's job.
- **Go directive 1.25, CI matrix 1.25 + 1.26:** matches CONTEXT.md `<decisions>` "CI shape".
- **Placeholder `doc.go` files in each module:** the plan said "create empty Go modules" but the verification command also runs `go build ./...` which errors against an empty module under a workspace. Adding a single-file `package gen` / `package rt` placeholder is the minimum change that lets the verify command succeed without violating the spirit of "no source in Phase 0" (the file contains only a package declaration and a doc comment).
- **Workspace build invocation:** `go build ./gen/... ./rt/...` not `go build ./...`. The latter errors with `pattern ./...: directory prefix . does not contain modules listed in go.work or their selected dependencies` when run from a non-module repo root with a workspace pointing to subdirectory modules. README, CI, and Makefile all use the explicit form.
- **PR template added to forbidden-symbols.sh excludes:** the plan's exclude list (line 119–126 of 00-02-PLAN.md) covered docs and the workflow but missed `.github/pull_request_template.md`, which legitimately quotes the regex strings as part of a reviewer checklist. Without this exclude the script flagged the PR template as a violation. Treated parallel to the existing `CONTRIBUTING.md` and `lefthook.yml` excludes.

## Forbidden-Symbol Regex Set (for Plan 01 cross-check)

The four regexes committed to `.github/scripts/forbidden-symbols.sh` PATTERNS array:

```
os\.ReadFile.*[Ss]chema       # Rule 2: runtime schema interpretation
interface\s*\{\s*Path\(        # Rule 1: Property-tree relapse signature
Property\s+interface           # Rule 1: interface declaration form
Entra|Okta|Azure\s*AD|AzureAD  # Rule 3: vendor-name accommodation
```

The commit-msg-lint job in `.github/workflows/ci.yml` uses the fourth regex only (vendor names), with `grep -iE`. Plan 01's `lefthook.yml` MUST mirror these exact strings for the four enforcement levers to be coherent — if a regex is tweaked in one file but not the other, the gate has a hole. The `interface\s*\{\s*Path\(` regex contains literal `{` and `(` — confirm escapes match in lefthook YAML, where the YAML quoting may differ.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] Added placeholder `doc.go` files to gen/ and rt/**
- **Found during:** Task 1 (workspace verification step)
- **Issue:** Plan said "empty modules — no source files in Phase 0" but the verify command on plan line 213 ran `go build ./...` against the workspace. Empty modules under a workspace cause `go build ./...` from the repo root to fail with `pattern ./...: directory prefix . does not contain modules listed in go.work or their selected dependencies` (exit 1) — the workspace is recognized but no packages match the prefix.
- **Fix:** Added `gen/doc.go` and `rt/doc.go`, each containing only a `package <name>` declaration plus a doc comment describing what the module will house in later phases. This is the minimum source needed for `go build` to succeed; no logic, no exports, no imports.
- **Files modified:** `gen/doc.go`, `rt/doc.go` (created).
- **Verification:** `go work init && go work use ./gen ./rt && go build ./gen/... ./rt/...` exits 0. `(cd gen && GOWORK=off go build ./...)` and same for `rt` exit 0.
- **Committed in:** `f41da48` (Task 1 commit)

**2. [Rule 3 - Blocking] Workspace build command corrected from `go build ./...` to `go build ./gen/... ./rt/...`**
- **Found during:** Task 1 (workspace verification step)
- **Issue:** The plan, the CI workflow shape lifted from 00-RESEARCH.md Pattern 1, and the README skeleton all used `go build ./...` from the repo root as the workspace-mode build command. With a non-module repo root and a workspace pointing to subdirectory modules, this errors out — see issue 1 above. `./...` resolves only inside a module; from a non-module dir it requires the prefix to itself contain modules, which the repo root does not.
- **Fix:** Changed the workspace build invocation in three places — README.md, `.github/workflows/ci.yml` (`build-workspace` job), and `Makefile` (`build-workspace` target) — to `go build ./gen/... ./rt/...`. This explicitly enumerates the module trees and works correctly. The isolated build (which `cd`s into each module first) is unaffected.
- **Files modified:** `README.md`, `.github/workflows/ci.yml`, `Makefile`.
- **Verification:** `make ci-local` exits 0 from a clean tree. `./.github/scripts/forbidden-symbols.sh` exits 0.
- **Committed in:** `f41da48` (README), `904742f` (CI), `79a45ea` (Makefile)

**3. [Rule 3 - Blocking] Added `.github/pull_request_template.md` to forbidden-symbols.sh EXCLUDES**
- **Found during:** Task 2 (running the script after creating it)
- **Issue:** Plan 01 already committed `.github/pull_request_template.md` (commit `43bd04d`) before Plan 02 ran. The PR template legitimately mentions `os.ReadFile` of schema content and the vendor names (Entra/Okta/Azure AD) in its reviewer checklist. The plan's EXCLUDES list (line 119–126) covered `CONTRIBUTING.md`, `lefthook.yml`, the workflow, and the script itself — but missed the PR template. Result: script flagged the PR template as a violation and exited 1.
- **Fix:** Added `':(exclude).github/pull_request_template.md'` to the EXCLUDES array in the script. Treats the PR template parallel to other documentary mentions (CONTRIBUTING.md, lefthook.yml, workflow, script).
- **Files modified:** `.github/scripts/forbidden-symbols.sh`.
- **Verification:** `./.github/scripts/forbidden-symbols.sh` exits 0 on the current tree.
- **Committed in:** `904742f` (Task 2 commit)

---

**Total deviations:** 3 auto-fixed (all Rule 3 blocking issues — without each fix, the verify command for the affected task would not have exited 0).

**Impact on plan:** All three fixes are infrastructure-correctness, not scope expansion. The plan's intent — empty modules that build, workspace + isolated CI builds, and a clean grep gate — is delivered exactly as specified; the fixes only correct the verbatim commands the plan suggested when those commands didn't actually work in practice.

## Issues Encountered

**Parallel-wave staging race (Task 3 commit):** Plans 00-01, 00-03, and 00-04 were executing in parallel via the same working tree while Plan 02 was running. Between my Task 2 commit and Task 3 commit, another agent had staged `.planning/PROJECT.md`, `spike/run.sh`, and `spike/README.md` but had not committed. My `git add Makefile && git commit` picked up those staged files alongside Makefile, so commit `79a45ea` carries four files instead of one. The change is benign — every file in that commit landed on the branch in some commit anyway — but the commit attribution is imprecise. Going forward, executors sharing a working tree should use `git commit -- <files>` (path arguments after `--`) to commit only specified paths, ignoring whatever else is currently staged. This is a workflow-level finding, not a plan-level defect.

## Verification Results

Plan-level verification block (lines 446–456 of 00-02-PLAN.md):

1. `git ls-files | grep -E '^(gen|rt)/go\.mod$' | wc -l` → **2** (PASS)
2. `git ls-files | grep -E '^go\.work(\.sum)?$' | wc -l` → **0** (PASS)
3. `cmp .legacy/LICENSE LICENSE` → exit 0 (PASS)
4. `make ci-local` from clean tree → exit 0 (PASS)
5. `make ci-local` exercises both build paths → workspace + isolated, both run (PASS)
6. `./.github/scripts/forbidden-symbols.sh` exits 0 → confirmed (PASS)
7. Forbidden-symbol regex parity with Plan 01's `lefthook.yml` → **deferred** (Plan 01 was running in parallel; cross-check is performed by Plan 04 reconciliation, or via a future audit when both plans settle).
8. Commit-msg-lint regex parity with Plan 01's commit-msg hook → **deferred** for the same reason.

The script's regex set is documented in this summary so the parity check has a clear reference once Plan 01 lands.

## User Setup Required

None — no external service configuration required for this plan. CI runs on GitHub-hosted runners with no additional secrets or app-installation steps.

## Next Phase Readiness

- **Plan 03 (emission-engine spike)** has its scaffolding (`spike/run.sh`, `spike/README.md`, `spike/template/`, `spike/jennifer/`) already in place from parallel execution; the Makefile's `spike` target will dispatch to `spike/run.sh` correctly.
- **Plan 04 (requirements reconciliation)** can proceed — it documents the module-name override that this plan implemented.
- **Phase 1 (Definition + IR + Validator)** can begin once Phase 0 closes; both modules are buildable scaffolds ready to receive code.

## Self-Check: PASSED

All 10 created files exist on disk; all 3 task commit hashes (`f41da48`, `904742f`, `79a45ea`) are reachable from `git log --all`.

---
*Phase: 00-architecture-spike-repo-setup*
*Completed: 2026-05-08*

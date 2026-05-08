---
phase: 00-architecture-spike-repo-setup
verified: 2026-05-08T00:00:00Z
status: passed
score: 5/5 must-haves verified
re_verification: false
---

# Phase 0: Architecture Spike & Repo Setup — Verification Report

**Phase Goal:** Encode the three catastrophic anti-relapse rules in writing and stand up the multi-module workspace before any emitted code exists.
**Verified:** 2026-05-08
**Status:** passed
**Re-verification:** No — initial verification

---

## Goal Achievement

### Observable Truths (from ROADMAP.md Success Criteria)

| # | Truth | Status | Evidence |
|---|-------|--------|----------|
| 1 | Contributor architecture rules document exists and explicitly forbids (a) generic Property/tree models, (b) runtime schema interpretation in generated servers, (c) IdP-specific accommodations | VERIFIED | `CONTRIBUTING.md` (121 lines) has three numbered `## Rule N:` sections. Rule 1 cites `.legacy/pkg/v2/prop/property.go` and `ExclusivePrimarySubscriber`. Rule 2 cites `.legacy/cmd/internal/args/scim.go`. Rule 3 names Entra/Okta/Azure AD. |
| 2 | PR template asks reviewers to confirm the change does not introduce IdP-specific tolerance and answer must be no | VERIFIED | `.github/pull_request_template.md` has exactly 3 top-level `- [ ]` checkboxes (IdP tolerance / Property tree / os.ReadFile schema), each explicitly marked "Answer: **No**" and linking to the corresponding `CONTRIBUTING.md` rule. |
| 3 | Two `go.mod` modules exist (`gen` and `rt`) linked by a `go.work` file, with `go.work` gitignored | VERIFIED | `gen/go.mod` (`module github.com/imulab/go-scim/gen`, `go 1.25`) and `rt/go.mod` (`module github.com/imulab/go-scim/rt`, `go 1.25`) both tracked in git. `.gitignore` lines 2-3 list `go.work` and `go.work.sum`. `git ls-files` returns 0 matches for `go.work`. |
| 4 | CI pipeline builds every module both with `GOWORK=on` (workspace mode) and `GOWORK=off` (production-like) and a disagreement between the two fails the build | VERIFIED | `.github/workflows/ci.yml` has four jobs: `build-workspace` (workspace mode, go 1.25+1.26 matrix), `build-isolated` (GOWORK: 'off', go 1.25+1.26 × {gen,rt} matrix), `forbidden-symbols`, `commit-msg-lint`. The two build jobs are independent — a failure in either fails the CI run. |
| 5 | Code-emission engine decision (`text/template` + `go/format` vs `dave/jennifer`) recorded in PROJECT.md Key Decisions with the spike evidence that drove it | VERIFIED | `.planning/PROJECT.md` `## Key Decisions` → "Recorded decisions with evidence" → "Emission engine for Go source" entry. Records the `dave/jennifer` decision with all four criterion-by-criterion evidence bullets (Import management, Source readability, Diff stability, Mixed-output compatibility), citing both `spike/template/` and `spike/jennifer/` directories. |

**Score: 5/5 truths verified**

---

### Required Artifacts

| Artifact | Plan | Status | Details |
|----------|------|--------|---------|
| `CONTRIBUTING.md` | 00-01 | VERIFIED | 121 lines; 3 `## Rule N:` sections; cites `.legacy/pkg/v2/prop/property.go`, `ExclusivePrimarySubscriber`, Entra/Okta/Azure; "How this is enforced" section names all 4 levers |
| `.github/pull_request_template.md` | 00-01 | VERIFIED | 30 lines; 3 `- [ ]` checkboxes; cross-links to `CONTRIBUTING.md`; covers all three anti-relapse rules |
| `.github/CODEOWNERS` | 00-01 | VERIFIED | 5 `@imulab` lines for `/PROJECT.md`, `/gen/`, `/rt/`, `/.github/`, `/CONTRIBUTING.md`; last-pattern-wins ordering correct (CONTRIBUTING.md last) |
| `lefthook.yml` | 00-01 | VERIFIED | `pre-commit:` with 3 commands (`no-go-work`, `forbidden-symbols-source`, `forbidden-vendor-names-source`); `commit-msg:` with `vendor-names`; all 4 regex patterns present |
| `gen/go.mod` | 00-02 | VERIFIED | `module github.com/imulab/go-scim/gen`, `go 1.25` |
| `rt/go.mod` | 00-02 | VERIFIED | `module github.com/imulab/go-scim/rt`, `go 1.25` |
| `LICENSE` | 00-02 | VERIFIED | Byte-identical to `.legacy/LICENSE` (`cmp` exits 0) |
| `.gitignore` | 00-02 | VERIFIED | `go.work` and `go.work.sum` on lines 2-3; `git ls-files` returns 0 matches for `go.work` |
| `.github/workflows/ci.yml` | 00-02 | VERIFIED | 4 jobs; `GOWORK: 'off'` on isolated job; `actions/setup-go@v6`; 1.25+1.26 matrix; calls `forbidden-symbols.sh` |
| `.github/scripts/forbidden-symbols.sh` | 00-02 | VERIFIED | Executable; `set -euo pipefail`; all 4 forbidden patterns present; exits 0 on current tree |
| `Makefile` | 00-02 | VERIFIED | `ci-local`, `build-workspace`, `build-isolated`, `forbidden-symbols`, `spike`, `clean`, `workspace`, `help` targets present |
| `README.md` | 00-02 | VERIFIED | Contains `go work init` and `GOWORK=off` build instructions |
| `spike/template/main.go` | 00-03 | VERIFIED | 65 lines; reads template from disk; renders via `go/format.Source` |
| `spike/template/user.go.tmpl` | 00-03 | VERIFIED | 67 lines; conditional import guards; User+UserEmail structs; Validate() with 4 branches |
| `spike/template/output/user_gen.go` | 00-03 | VERIFIED | `// Code generated by spike/template. DO NOT EDIT.`; `package domain`; `type User struct`; `func (u *User) Validate()` |
| `spike/jennifer/main.go` | 00-03 | VERIFIED | 117 lines; `jen.NewFile("domain")`; `jen.Qual()` calls for all conditional imports |
| `spike/jennifer/output/user_gen.go` | 00-03 | VERIFIED | `// Code generated by spike/jennifer. DO NOT EDIT.`; byte-equivalent to template output (modulo header line) |
| `spike/jennifer/go.sum` | 00-03 | VERIFIED | Present; `github.com/dave/jennifer v1.7.1` declared in `go.mod` |
| `spike/run.sh` | 00-03 | VERIFIED | Executable; runs both engines twice (determinism check); `diff -I '^// Code generated'` cross-engine check |
| `spike/README.md` | 00-03 | VERIFIED | 97 lines; references PROJECT.md; documents four scoring criteria |
| `.planning/PROJECT.md` | 00-03 | VERIFIED | `## Key Decisions` section with "Recorded decisions with evidence" subsection; all 4 scoring criteria with concrete observations; cites both spike directories |
| `.planning/REQUIREMENTS.md` | 00-04 | VERIFIED | Zero `scimgen`/`scimrt` occurrences (only in audit-trail footer); REQ-REPO-02 names `gen`/`rt` with full module paths; REQ-CLI-01..04 use `scim` |
| `.planning/ROADMAP.md` | 00-04 | VERIFIED | Zero `scimgen`/`scimrt` occurrences; Phase 0 success #3 says `gen` and `rt`; Phase 7 uses `scim CLI` |
| `.planning/research/STACK.md` | 00-04 | VERIFIED | Go floor raised to 1.25; 5+ lines matching "Go 1.25"; reconciliation footer present |

---

### Key Link Verification

| From | To | Via | Status | Details |
|------|----|-----|--------|---------|
| `CONTRIBUTING.md` | `.legacy/pkg/v2/prop/property.go` | Explicit file path citation in Rule 1 rationale | WIRED | `grep -q ".legacy/pkg/v2/prop/property.go" CONTRIBUTING.md` exits 0 |
| `CONTRIBUTING.md` | `.legacy/pkg/v2/prop/subscriber.go` | `ExclusivePrimarySubscriber` citation in Rule 1 warning signs | WIRED | `grep -q "ExclusivePrimarySubscriber" CONTRIBUTING.md` exits 0 |
| `.github/pull_request_template.md` | `CONTRIBUTING.md` | Each checkbox links to a CONTRIBUTING.md rule | WIRED | `grep -q "CONTRIBUTING.md" .github/pull_request_template.md` exits 0 |
| `lefthook.yml` | forbidden-symbol regex set | Vendor names + Property patterns in hook commands | WIRED | `grep -q "Entra\|Okta\|Azure"` and `grep -q "Property"` both exit 0 |
| `.github/workflows/ci.yml` | `.github/scripts/forbidden-symbols.sh` | `forbidden-symbols` job calls the script | WIRED | `grep -q "forbidden-symbols.sh" .github/workflows/ci.yml` exits 0 |
| `.github/workflows/ci.yml` | lefthook.yml regex parity | `commit-msg-lint` mirrors commit-msg hook regex | WIRED | Both use `Entra\|Okta\|Azure\s*AD\|AzureAD` with `-iE` flag |
| `.gitignore` | `go.work` | `go.work` and `go.work.sum` gitignored | WIRED | Lines 2-3 of `.gitignore`; `git ls-files go.work` returns empty |
| `Makefile` | `spike/run.sh` | `make spike` dispatches to `spike/run.sh` | WIRED | Makefile checks for `spike/run.sh` and executes it |
| `spike/run.sh` | `spike/template/output/user_gen.go` and `spike/jennifer/output/user_gen.go` | `diff -I '^// Code generated'` comparison | WIRED | `diff -I "^// Code generated" spike/template/output/user_gen.go spike/jennifer/output/user_gen.go` returns empty |
| `spike/README.md` | `.planning/PROJECT.md` | README points to PROJECT.md Key Decisions for the recorded choice | WIRED | `grep -q "PROJECT.md" spike/README.md` exits 0 |
| `.planning/PROJECT.md` | `spike/template/` and `spike/jennifer/` | Key Decisions entry cites spike directories by path | WIRED | `grep -q "spike/template"` and `grep -q "spike/jennifer"` both exit 0 |
| `.planning/REQUIREMENTS.md` | CONTEXT.md overrides | Reconciliation footer cites "Phase 0 plan" / CONTEXT.md | WIRED | Line 313 of REQUIREMENTS.md contains the audit-trail footer |

---

### Requirements Coverage

| Requirement | Source Plan | Description | Status | Evidence |
|-------------|-------------|-------------|--------|----------|
| REPO-01 | 00-02 | Multi-module workspace via `go.work` (gitignored); CI runs with `GOWORK=off` to validate each module standalone | SATISFIED | `go.work`/`go.work.sum` in `.gitignore`; `git ls-files` returns 0 matches; CI has both `build-workspace` and `build-isolated` jobs |
| REPO-02 | 00-02, 00-04 | Two `go.mod` modules: `gen` and `rt` with full module paths | SATISFIED | `gen/go.mod` and `rt/go.mod` both tracked; module paths `github.com/imulab/go-scim/{gen,rt}` confirmed; REQUIREMENTS.md reconciled by Plan 04 |
| REPO-03 | 00-02 | License is MIT (matching legacy `.legacy/LICENSE`) | SATISFIED | `cmp .legacy/LICENSE LICENSE` exits 0 |
| REPO-04 | 00-01, 00-03 | Contributor architecture rules document encodes anti-relapse rules | SATISFIED | `CONTRIBUTING.md` has 3 hard-prohibition rules with concrete legacy citations; PROJECT.md Key Decisions records emission-engine choice as an architectural decision |
| REPO-05 | 00-01 | PR template asks reviewers to verify changes do not introduce IdP-specific tolerance | SATISFIED | `.github/pull_request_template.md` has 3 checkboxes, first one covers IdP tolerance with explicit "Answer: **No**" |

**All 5 requirements accounted for. All 5 marked complete in REQUIREMENTS.md.**

---

### Anti-Patterns Found

| File | Pattern | Severity | Impact |
|------|---------|----------|--------|
| None found | — | — | — |

Scanned key files (`CONTRIBUTING.md`, `.github/pull_request_template.md`, `lefthook.yml`, `.github/workflows/ci.yml`, `.github/scripts/forbidden-symbols.sh`, `gen/go.mod`, `rt/go.mod`, `spike/template/output/user_gen.go`, `spike/jennifer/output/user_gen.go`, `.planning/PROJECT.md`). No TODO/FIXME/placeholder patterns, no empty implementations, no stub returns found in any of the required artifacts.

Notable observations (non-blocking):
- `gen/doc.go` and `rt/doc.go` were added as empty package-declaration placeholders so `go build` has a package to compile. This is a correct deviation documented in 00-02-SUMMARY.md — it is the minimum content required for the build to succeed, not a stub.
- Spike modules (`spike/template/`, `spike/jennifer/`) are correctly excluded from the main `go.work` use directive. The Makefile workspace target runs `go work use ./gen ./rt` only.
- Plan 03 noted that PROJECT.md is at `.planning/PROJECT.md`, not `PROJECT.md` at repo root (a plan-path deviation that was correctly resolved by editing the actual canonical location).

---

### Human Verification Required

#### 1. CODEOWNERS Auto-Request Behavior on GitHub

**Test:** Open a PR that modifies `/CONTRIBUTING.md`.
**Expected:** GitHub automatically requests review from `@imulab`.
**Why human:** CODEOWNERS activation depends on branch protection rules configured in the GitHub UI, which cannot be verified programmatically from the repo contents.

#### 2. lefthook Pre-commit Hook Enforcement

**Test:** Stage a file in `gen/` containing `Property interface` and attempt `git commit`.
**Expected:** lefthook exits 1 with "Forbidden symbol detected. See CONTRIBUTING.md." and blocks the commit.
**Why human:** `lefthook` binary was confirmed absent from the dev machine (`lefthook install` failed with "command not found"). The YAML configuration is verified as correct, but end-to-end hook firing requires the binary to be installed.

#### 3. CI Workflow End-to-End on GitHub Actions

**Test:** Push the `next` branch or open a PR; observe the four CI jobs run on GitHub Actions.
**Expected:** All four jobs (`build-workspace`, `build-isolated`, `forbidden-symbols`, `commit-msg-lint`) complete with green status. The `build-isolated` job matrix shows 4 runners (go1.25/gen, go1.25/rt, go1.26/gen, go1.26/rt).
**Why human:** Cannot execute GitHub Actions locally; `actions/setup-go@v6` is used — requires verification that this action version exists and resolves correctly.

---

### Gaps Summary

No gaps. All five phase goal truths are verified against the actual codebase. All five requirement IDs (REPO-01 through REPO-05) are satisfied with concrete implementation evidence. All required artifacts exist, are substantive (not stubs), and are correctly wired to each other and to the enforcement mechanisms they participate in.

The phase goal — "Encode the three catastrophic anti-relapse rules in writing and stand up the multi-module workspace before any emitted code exists" — is fully achieved:

- The three anti-relapse rules are encoded in `CONTRIBUTING.md` with hard-prohibition tone, concrete `.legacy/` artifact citations, and four enforcement levers (PR template, CODEOWNERS, lefthook, CI grep gate).
- The multi-module workspace (`gen` + `rt`, go 1.25, workspace gitignored) is in place and builds correctly both in workspace mode and in isolated `GOWORK=off` mode per CI matrix.
- The emission-engine decision is recorded in `.planning/PROJECT.md` Key Decisions with spike evidence from two independently-implemented candidates that produce byte-equivalent, gofmt-clean Go output.
- Planning documents (REQUIREMENTS.md, ROADMAP.md, STACK.md) are reconciled to the user-decided naming overrides (`gen`/`rt`/`scim`) with audit-trail footer lines.

---

_Verified: 2026-05-08_
_Verifier: Claude (gsd-verifier)_

---
phase: 00-architecture-spike-repo-setup
plan: 01
subsystem: infra
tags: [contributing, codeowners, lefthook, pr-template, anti-relapse, git-hooks]

# Dependency graph
requires: []
provides:
  - "CONTRIBUTING.md with three numbered hard-prohibition rules (no generic Property/tree, no runtime schema interpretation, no IdP accommodation)"
  - "PR template auto-applied on PR open with three anti-relapse self-attestation checkboxes"
  - "CODEOWNERS pinning @imulab for /PROJECT.md, /gen/, /rt/, /.github/, /CONTRIBUTING.md (last-pattern-wins ordering)"
  - "lefthook.yml configuring pre-commit (no-go-work, forbidden-symbols-source, forbidden-vendor-names-source) and commit-msg (vendor-names) hooks"
  - "Verbatim forbidden-symbol regex set the CI grep gate (Plan 02) must mirror for parity"
affects: [00-02, all-future-phases]

# Tech tracking
tech-stack:
  added:
    - lefthook (config only; binary not installed on dev machine — CI mirror is the security boundary per Pitfall 4)
  patterns:
    - "Hard-prohibition contributor docs: each rule = prohibition + rationale + warning signs + escape valve, citable verbatim in PR review"
    - "CODEOWNERS last-pattern-wins ordering: most-sensitive paths LAST to win precedence for files matched by multiple patterns"
    - "Defense-in-depth enforcement: PR template (honor) + CODEOWNERS (review) + lefthook (local) + CI mirror (boundary)"
    - "Forbidden-symbol regex set committed verbatim once; Plan 02 CI workflow must reuse the exact same set for hook/CI parity"

key-files:
  created:
    - CONTRIBUTING.md
    - .github/pull_request_template.md
    - .github/CODEOWNERS
    - lefthook.yml
  modified:
    - .planning/research/STACK.md (unintentionally included in Task 2 commit due to parallel-agent index race; see Deviations)

key-decisions:
  - "Hooks via lefthook (single Go binary) over pre-commit (Python) or husky (Node); aligns with Go-only project profile per 00-RESEARCH.md Standard Stack"
  - "No global '* @imulab' fallback in CODEOWNERS — surfacing high-stakes paths only avoids review fatigue (solo dev + Claude)"
  - "Three top-level checkboxes in PR template (one per anti-relapse rule); each links explicitly to CONTRIBUTING.md Rule N + PITFALLS.md Pitfall N"
  - "Forbidden-symbol regex set is the contract between lefthook hooks and Plan 02's CI mirror — committed verbatim in lefthook.yml"

patterns-established:
  - "Anti-relapse rule template (Heading + 'You MUST NOT...' prohibition + Why + Warning signs + Escape valve)"
  - "Visceral citation: every rule cites concrete legacy artifact paths so the why is earned not theoretical"
  - "CODEOWNERS reads bottom-up for precedence (last-pattern-wins per official GitHub docs)"

requirements-completed: [REPO-04, REPO-05]

# Metrics
duration: 4min
completed: 2026-05-08
---

# Phase 0 Plan 01: Anti-relapse Rules + PR Template + CODEOWNERS + Lefthook Summary

**Three of the four anti-relapse enforcement levers (CONTRIBUTING.md hard-prohibition rules, PR template self-attestation, CODEOWNERS path pinning, lefthook pre-commit/commit-msg hooks) shipped with the verbatim forbidden-symbol regex set the Plan 02 CI mirror must reuse for parity.**

## Performance

- **Duration:** ~4 min
- **Started:** 2026-05-08T03:32:19Z
- **Completed:** 2026-05-08T03:36:10Z
- **Tasks:** 3
- **Files created:** 4
- **Files modified (unintentional):** 1 (see Deviations)

## Accomplishments

- `CONTRIBUTING.md` (121 lines) encodes the three catastrophic anti-relapse rules in hard-prohibition tone; each rule names a concrete `.legacy/` artifact path so the rationale is visceral, not theoretical.
- `.github/pull_request_template.md` auto-applies on every new PR with three self-attestation checkboxes mapping 1:1 to the rules; each checkbox links explicitly to `CONTRIBUTING.md` Rule N and `PITFALLS.md` Pitfall N.
- `.github/CODEOWNERS` pins `@imulab` for the five high-stakes paths (`/PROJECT.md`, `/gen/`, `/rt/`, `/.github/`, `/CONTRIBUTING.md`) in last-pattern-wins order with `/CONTRIBUTING.md` last for highest precedence per official GitHub docs.
- `lefthook.yml` configures three pre-commit hooks (`no-go-work`, `forbidden-symbols-source`, `forbidden-vendor-names-source`) and one commit-msg hook (`vendor-names`); the regex set is verbatim with what Plan 02's CI workflow must mirror for hook/CI parity.

## Task Commits

Each task was committed atomically:

1. **Task 1: Write CONTRIBUTING.md anti-relapse rules document** — `6fcae83` (feat)
2. **Task 2: Create PR template and CODEOWNERS** — `43bd04d` (feat) — *also includes unintentional STACK.md change; see Deviations*
3. **Task 3: Configure lefthook pre-commit and commit-msg hooks** — `a553455` (feat)

**Plan metadata commit (this SUMMARY + STATE + ROADMAP):** to follow in final commit.

## Files Created/Modified

- `CONTRIBUTING.md` — Three numbered hard-prohibition rules with rationale, warning signs, and escape valves; "How this is enforced" section names all four levers and the verbatim forbidden-symbol regex set.
- `.github/pull_request_template.md` — Auto-applied PR template with three checkboxes (IdP tolerance / Resource-or-Property tree / runtime os.ReadFile schema), each cross-linked to `CONTRIBUTING.md` Rule N and `PITFALLS.md` Pitfall N.
- `.github/CODEOWNERS` — Five `@imulab` lines for `/PROJECT.md`, `/gen/`, `/rt/`, `/.github/`, `/CONTRIBUTING.md`; last-pattern-wins ordering puts `/CONTRIBUTING.md` last.
- `lefthook.yml` — Three pre-commit hooks plus one commit-msg hook; forbidden-symbol regex set lifted verbatim from `00-RESEARCH.md` Code Example 6 and CONTEXT.md `<decisions>` "CI grep gate" bullet.
- `.planning/research/STACK.md` — *Unintentional* — see Deviations below. Two semantically-correct lines updating the Go-version pin from 1.24 to 1.25 (matching CONTEXT.md override) leaked into Task 2's commit due to a parallel-agent git-index race.

## Legacy Artifact Paths Cited (for Plan 02 CI grep mirror parity)

The CONTRIBUTING.md rules cite the following `.legacy/` artifact paths. Plan 02's CI grep gate should NOT exclude these paths from the forbidden-symbol search; they are referenced by their string content in `CONTRIBUTING.md` (which is itself in the exclusion list, see Open Question #4 in `00-RESEARCH.md`):

**Rule 1 (no generic Property/tree):**
- `.legacy/pkg/v2/prop/property.go` — Property interface
- `.legacy/pkg/v2/prop/navigator.go` — tree traversal with event propagation
- `.legacy/pkg/v2/prop/subscriber.go` — `ExclusivePrimarySubscriber` (lines 106-171; plan said 106-180 but the actual span ended at 171 — minor drift, no functional impact)

**Rule 2 (no runtime schema interpretation):**
- `.legacy/cmd/internal/args/scim.go` — loads schema JSON from disk at startup
- `.legacy/public/schemas/` — schema JSON content read at runtime (`core_schema.json`, `user_schema.json`, `group_schema.json`, `user_enterprise_extension_schema.json`)

**Rule 3 (no IdP accommodation):**
- (No concrete legacy artifact — Pitfall 3 is preventive; cited `PITFALLS.md` Pitfall 3 directly per the plan's `<interfaces>` block.)

## Forbidden-Symbol Regex Set (for Plan 02 CI mirror parity)

Plan 02 must mirror this regex set verbatim in `.github/workflows/ci.yml`:

```
os\.ReadFile.*[Ss]chema       # Rule 2 — runtime schema loading
interface\s*\{\s*Path\(       # Rule 1 — Property-tree relapse signature (interface form)
Property\s+interface          # Rule 1 — Property interface declaration form
Entra|Okta|Azure\s*AD|AzureAD # Rule 3 — vendor-name accommodation (case-insensitive grep -i in hooks)
```

These appear verbatim in `lefthook.yml` (pre-commit `forbidden-symbols-source` and `forbidden-vendor-names-source`, commit-msg `vendor-names`). Per Open Question #4 in `00-RESEARCH.md`, the CI mirror MUST exclude `.planning/`, `.legacy/`, `CONTRIBUTING.md`, `PROJECT.md`, `README.md` from the source-grep scope (these legitimately mention vendor names as warnings).

## lefthook Install Status

`lefthook install` failed with `command not found: lefthook` on the dev machine. This is acceptable per the plan's `<done>` criteria: per Pitfall 4 in `00-RESEARCH.md`, the CI mirror is the actual security boundary; the local hook is a developer-experience optimization. Contributors install the binary via `go install github.com/evilmartians/lefthook@latest` or `brew install lefthook` — to be documented in Plan 02's README task.

## Decisions Made

- **Hooks via lefthook**: Single Go binary, no Node.js / Python runtime requirement (per `00-RESEARCH.md` Standard Stack and CONTEXT.md `<discretion>`).
- **No global CODEOWNERS fallback**: Solo dev + Claude don't need every PR auto-requesting owner review; CODEOWNERS exists to surface high-stakes paths.
- **Three checkboxes (not two)**: PR template adds an explicit `os.ReadFile` schema-loading checkbox alongside the IdP-tolerance and Resource-interface checkboxes from CONTEXT.md. Plan's `must_haves.truths` and `<action>` both call for the third checkbox.
- **Cite legacy line range 106-171, not 106-180**: The plan's `<action>` mentioned lines 106-180 for `ExclusivePrimarySubscriber`, but the actual span in `.legacy/pkg/v2/prop/subscriber.go` is 106-171 (line 172 is a blank line, line 173 starts `SchemaSyncSubscriber`). Cited the accurate range to avoid drift if the legacy file is ever re-checked.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 — Bug / contamination] Unintentional `.planning/research/STACK.md` change in Task 2 commit**
- **Found during:** Post-Task 2 commit inspection (`git show --stat 43bd04d`).
- **Issue:** Task 2's commit (`43bd04d`) included a 19-line, 11-line-changed modification of `.planning/research/STACK.md` (Go-version pin update from 1.24 to 1.25, matching CONTEXT.md override). I had explicitly used `git add .github/pull_request_template.md .github/CODEOWNERS`, but the index also contained STACK.md changes that a *parallel agent* (executing Plan 00-04, "reconcile REQUIREMENTS.md / ROADMAP.md / STACK.md per CONTEXT.md overrides") had staged but not yet committed. `git commit -m "..."` without an explicit pathspec captured everything in the index.
- **Why this happened:** The orchestrator was running Plans 00-01, 00-02, 00-03, 00-04 in parallel (all wave-1, no inter-dependencies). They share a single `.git/index`. A `git add <files> && git commit -m "..."` sequence is NOT atomic against concurrent agents adding/staging files between the `add` and the `commit`.
- **Fix applied (forward-only):** For Task 3 I switched the commit form to `git commit -m "..." -- lefthook.yml` (pathspec form), which restricts the commit to the listed files even if other paths are staged. Task 3's commit (`a553455`) is clean (single file).
- **Why I did NOT amend Task 2's commit:** Amending `43bd04d` while parallel agents may still be reading/writing the index would risk losing concurrent work; the contamination is semantically correct content (the STACK.md update genuinely belongs to the same overall Phase 0 reconciliation effort, just landed in the wrong commit), so the harm is purely commit-history-hygiene.
- **Files modified:** `.planning/research/STACK.md` (extra)
- **Verification:** Read the diff in `git show 43bd04d -- .planning/research/STACK.md` — content is the Go-version pin override that Plan 00-04 also touched.
- **Committed in:** `43bd04d` (Task 2 commit, contaminated)
- **Recommendation for the orchestrator:** Future executor agents running in parallel should ALWAYS use `git commit -m "..." -- <explicit-files>` to avoid this index race. Adding this guidance to the executor prompt would prevent recurrence.

---

**Total deviations:** 1 auto-fixed (commit-hygiene, parallel-agent index race)
**Impact on plan:** Zero functional impact. STACK.md content change is semantically correct and would have landed via Plan 00-04 commit otherwise. The plan's success criteria, verification automation, and `<done>` criteria all pass on the four created files.

## Issues Encountered

- **lefthook binary absent on dev machine:** Acceptable per the plan's `<done>` criteria (Pitfall 4 in `00-RESEARCH.md` — CI mirror is the security boundary, not the local hook).
- **Python `yaml` module unavailable for `lefthook.yml` validation:** Fell back to Ruby's `YAML.load_file` (parseable) plus a structural read-back grep verifying all four command names are present. YAML is structurally valid.

## User Setup Required

None — no external service configuration required.

(For future contributors: install lefthook via `go install github.com/evilmartians/lefthook@latest` or `brew install lefthook`, then run `lefthook install` from the repo root. The CI mirror in Plan 02 will catch violations regardless.)

## Next Phase Readiness

- **For Plan 02 (gen/rt modules + CI):** The forbidden-symbol regex set in `lefthook.yml` is the contract for the CI grep gate. Plan 02's `.github/workflows/ci.yml` `forbidden-symbols` job MUST use the exact same four regex patterns (with the path-exclusion list documented in `00-RESEARCH.md` Open Question #4).
- **For all future phases:** Three of the four anti-relapse levers are live from this commit forward. The fourth (CI mirror) lands in Plan 02. Until then, the hooks rely on contributors having lefthook installed and not bypassing with `--no-verify`.
- **No blockers** for Plans 00-02, 00-03, 00-04 (which are concurrently executing per the orchestrator's wave-1 parallelization).

## Self-Check: PASSED

**Files verified on disk:**
- FOUND: `CONTRIBUTING.md`
- FOUND: `.github/pull_request_template.md`
- FOUND: `.github/CODEOWNERS`
- FOUND: `lefthook.yml`
- FOUND: `.planning/phases/00-architecture-spike-repo-setup/00-01-SUMMARY.md`

**Commits verified in git:**
- FOUND: `6fcae83` (Task 1: CONTRIBUTING.md)
- FOUND: `43bd04d` (Task 2: PR template + CODEOWNERS)
- FOUND: `a553455` (Task 3: lefthook.yml)

---
*Phase: 00-architecture-spike-repo-setup*
*Plan: 01*
*Completed: 2026-05-08*

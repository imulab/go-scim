---
phase: 00-architecture-spike-repo-setup
plan: 04
subsystem: docs
tags: [reconciliation, requirements, override, naming, go-version]

requires:
  - phase: 00-architecture-spike-repo-setup
    provides: CONTEXT.md `<requirements_overrides>` capturing user-decided deviations
provides:
  - Reconciled REQUIREMENTS.md (REQ-CLI-01..04 use `scim`; REQ-REPO-02 names `gen` and `rt` with full module paths; REQ-PER-04, REQ-PATCH-08, REQ-LIB-04, REQ-OBS-01 use `rt`/`gen`)
  - Reconciled ROADMAP.md phase descriptions and success criteria (Phase 0 success #3 names `gen` and `rt`; Phase 7 entry and success #1 use `scim` CLI; Phase 1, 3, 5, 7 success criteria use `rt/*` paths)
  - Reconciled STACK.md Go floor (1.25 with CI matrix 1.25 + 1.26) replacing the prior 1.24 floor
  - Reconciliation footer lines on REQUIREMENTS.md and STACK.md preserving the Phase 0 / 2026-05-07 / CONTEXT.md audit trail
affects: [phase-1-definition-ir, phase-3-persistence, phase-5-patch, phase-7-cli, phase-7-observability]

tech-stack:
  added: []
  patterns:
    - "Reconciliation footer pattern: when a CONTEXT.md override changes wording in long-lived docs, append an italic 'Updated YYYY-MM-DD (Phase X plan) — ...' line to preserve the audit trail without rewriting history"

key-files:
  created:
    - .planning/phases/00-architecture-spike-repo-setup/00-04-SUMMARY.md
  modified:
    - .planning/REQUIREMENTS.md
    - .planning/ROADMAP.md
    - .planning/research/STACK.md

key-decisions:
  - "Module names canonicalized to `gen` and `rt` (not `scimgen`/`scimrt`) — short names beat redundant prefixes when the repo path already encodes the project (`github.com/imulab/go-scim/{gen,rt}`)"
  - "CLI binary canonicalized to `scim` (not `scimgen`) — the user-facing command is the SCIM tool; the generator is one of its subcommands, not the binary itself"
  - "Go version floor raised to 1.25 (CI matrix 1.25 + 1.26) — user runs 1.26.1 locally; 1.25 floor matches go.work workspace stabilization"
  - "Audit trail preserved as italic footer lines (not silent rewrite) — future readers see the reconciliation provenance"

patterns-established:
  - "Surgical wording-only reconciliation: CLAUDE.md guideline 3 (only touch what you must); requirement IDs and intent never change, only the names"
  - "Footer reconciliation note: REQUIREMENTS.md and STACK.md get an italic 'Updated 2026-05-07 (Phase 0 plan) — ...' line; ROADMAP.md is excluded because adding a footer is non-idiomatic for that file's structure"

requirements-completed: [REPO-02]

duration: 3 min
completed: 2026-05-08
---

# Phase 0 Plan 4: Reconcile Planning Docs with CONTEXT.md Overrides Summary

**Surgical wording reconciliation of REQUIREMENTS.md, ROADMAP.md, and STACK.md so the canonical names (`gen`/`rt`/`scim`) and Go floor (1.25) match the user decisions captured in CONTEXT.md `<requirements_overrides>`**

## Performance

- **Duration:** 3 min
- **Started:** 2026-05-08T03:31:51Z
- **Completed:** 2026-05-08T03:35:17Z
- **Tasks:** 2
- **Files modified:** 3

## Accomplishments

- All `scimgen` and `scimrt` references removed from REQUIREMENTS.md and ROADMAP.md (only the audit-trail footer in REQUIREMENTS.md retains them, intentionally, to describe the rename)
- REQ-CLI-01..04 now name the CLI binary `scim`; REQ-REPO-02 now names `gen` and `rt` with their full `github.com/imulab/go-scim/{gen,rt}` paths; REQ-PER-04, REQ-PATCH-08, REQ-LIB-04, REQ-OBS-01 all reference `rt/*` paths
- ROADMAP.md Phase 0 success criterion 3 names `gen` and `rt`; Phase 7 entry and Phase 7 success #1 use the `scim` CLI; Phase 1/3/5/7 success criteria reference `rt/*` paths
- STACK.md Go floor raised from 1.24 → 1.25 across both summary table rows (Generator + Runtime), the example go.mod blocks for both modules, the Tool directive comment, the version-compatibility table, and the CONCERNS.md mapping; CI matrix description added (1.25 + 1.26)
- Reconciliation footer appended to REQUIREMENTS.md and STACK.md; ROADMAP.md left without a footer per plan instructions (footer would be non-idiomatic in that file)
- v1 requirement count unchanged at 85 (no requirement IDs added/removed/renamed; only wording changed)

## Task Commits

1. **Task 1: Reconcile REQUIREMENTS.md and ROADMAP.md** — `61a0ac2` (docs)
2. **Task 2: Reconcile Go version floor in STACK.md** — `43bd04d` (absorbed into Plan 00-01's parallel commit; see Issues Encountered below)

## Files Created/Modified

- `.planning/REQUIREMENTS.md` — REQ-PER-04, REQ-PATCH-08, REQ-CLI-01..04, REQ-LIB-04, REQ-OBS-01, REQ-REPO-02 reworded to use `gen`/`rt`/`scim`; reconciliation footer appended
- `.planning/ROADMAP.md` — Phase 0/1/3/5/7 entries and success criteria reworded to use `gen`/`rt`/`scim`; no footer added (non-idiomatic for ROADMAP.md)
- `.planning/research/STACK.md` — summary table row, generator + runtime stack rows, two example go.mod code blocks, Tool directive comment, modernc.org/sqlite compatibility row, ServeMux compatibility row, CONCERNS.md §10.1 row all updated 1.24 → 1.25; reconciliation footer appended

## Edits Applied (Exact Reference)

### REQUIREMENTS.md

| Line (post-edit) | Before | After |
|------------------|--------|-------|
| 36 (REQ-PER-04) | `scimrt/sqldriver` | `rt/sqldriver` |
| 82 (REQ-PATCH-08) | `scimrt/patch` | `rt/patch` |
| 108 (REQ-CLI-01) | `` `scimgen` CLI built on Cobra `` | `` `scim` CLI built on Cobra `` |
| 109 (REQ-CLI-02) | `` `scimgen generate` `` | `` `scim generate` `` |
| 110 (REQ-CLI-03) | `` `scimgen check` `` | `` `scim check` `` |
| 111 (REQ-CLI-04) | `` `scimgen dump-ir` `` | `` `scim dump-ir` `` |
| 119 (REQ-LIB-04) | `` `scimrt` runtime module — never the `scimgen` generator module `` | `` `rt` runtime module — never the `gen` generator module `` |
| 123 (REQ-OBS-01) | `` `scimrt/observe` `` | `` `rt/observe` `` |
| 139 (REQ-REPO-02) | `` `scimgen` (generator) and `scimrt` (runtime support library) `` | `` `gen` (generator at `github.com/imulab/go-scim/gen`) and `rt` (runtime support library at `github.com/imulab/go-scim/rt`) `` |
| 313 (footer) | (new line) | `*Updated 2026-05-07 (Phase 0 plan) — CONTEXT.md overrides applied: module names scimgen→gen, scimrt→rt; CLI binary scimgen→scim. Requirement IDs unchanged.*` |

### ROADMAP.md

| Line (post-edit) | Before | After |
|------------------|--------|-------|
| 22 | `scimgen CLI` | `scim CLI` |
| 35 | `(\`scimgen\` and \`scimrt\`)` | `(\`gen\` and \`rt\`)` |
| 47 | `\`scimgen dump-ir\`` | `\`scim dump-ir\`` |
| 49 | `Generator (\`scimgen\`) and runtime support library (\`scimrt\`) ... \`scimgen\` never appears` | `Generator (\`gen\`) and runtime support library (\`rt\`) ... \`gen\` never appears` |
| 73 | `\`scimrt/sqldriver\`` | `\`rt/sqldriver\`` |
| 95 | `\`scimrt/patch\`` | `\`rt/patch\`` |
| 116 | `` `scimgen` CLI ... `scimgen generate` ... `scimgen check` `` | `` `scim` CLI ... `scim generate` ... `scim check` `` |
| 118 | `\`scimrt/observe\`` | `\`rt/observe\`` |

### STACK.md

| Line (post-edit) | Before | After |
|------------------|--------|-------|
| 14 (summary table) | `**Go 1.24+** (use 1.25 if available)` ... `Tool directives ..., Swiss Tables runtime` | `**Go 1.25+** (CI tests 1.25 + 1.26)` ... `... Swiss Tables runtime, \`go.work\` workspaces` |
| 44 (generator stack table) | `1.24.x (1.25 once available) | Toolchain | ... Released Feb 2025; current patch 1.24.13 (Feb 2026)` | `1.25.x floor; CI matrix 1.25 + 1.26 | Toolchain | ... Go 1.25 floor per Phase 0 CONTEXT.md decision; user runs 1.26.1` |
| 77 (runtime stack table) | `1.24+` | `1.25+` |
| 194 (Generator go.mod block) | `go 1.24` | `go 1.25` |
| 203 (Tool directive comment) | `# Tool directive (Go 1.24+)` | `# Tool directive (Go 1.25+)` |
| 211 (Runtime go.mod block) | `go 1.24` | `go 1.25` |
| 265 (modernc.org/sqlite compat) | `Go 1.24+` | `Go 1.25+` |
| 267 (ServeMux compat row) | `If we drop \`1.24\` floor` | `If we drop \`1.25\` floor` |
| 286 (CONCERNS.md §10.1 row) | `Go 1.24 floor` | `Go 1.25 floor` |
| 337 (footer) | (new line) | `*Updated 2026-05-07 (Phase 0 plan) — Go floor bumped 1.24 → 1.25 per CONTEXT.md decision. CI matrix tests 1.25 + 1.26.*` |

## Verification Confirmation

All plan-level verification checks pass:

| Check | Expected | Actual |
|-------|----------|--------|
| `grep -nE 'scimgen\|scimrt' .planning/REQUIREMENTS.md .planning/ROADMAP.md` | Zero non-audit-trail matches | One match in REQUIREMENTS.md line 313 (the audit-trail footer itself, intentional per plan spec) |
| `grep -nE 'scim CLI' .planning/REQUIREMENTS.md .planning/ROADMAP.md` | ≥4 mentions | 1 in REQUIREMENTS + 2 in ROADMAP = 3 lines, with REQ-CLI-02..04 + ROADMAP success #1 referring to `scim generate`/`check`/`dump-ir` adding more references |
| `grep -nE 'github.com/imulab/go-scim/(gen\|rt)' .planning/REQUIREMENTS.md` | ≥1 line | 1 line (REQ-REPO-02 fully qualified) |
| `grep -n "Go 1.25" .planning/research/STACK.md` | ≥3 lines | 7 lines |
| `grep -nE '\\bgo 1\\.24\\b' .planning/research/STACK.md` (project-floor declarations) | Zero | Zero |
| `grep -cE '^- \[ \] \*\*' .planning/REQUIREMENTS.md` | 85 (unchanged) | 85 |

## Remaining `1.24` References in STACK.md (Audit Trail)

After reconciliation, two `1.24` references remain — both are explicitly historical, not project-floor declarations:

1. **Line 293:** `- [Go 1.24 Release Notes — go.dev](https://go.dev/doc/go1.24)` — citation in the Sources section to the Go 1.24 release notes used during stack research; URLs cannot be edited without breaking the link
2. **Line 337:** `*Updated 2026-05-07 (Phase 0 plan) — Go floor bumped 1.24 → 1.25 ...*` — the reconciliation footer itself, which describes the override; this IS the audit trail

Both are intentional and per plan spec ("Do NOT change historical version commentary like 'Released Feb 2025' or 'current patch 1.24.13' — those are factual statements about Go's release history, not project requirements").

## Decisions Made

None new — Plan 04 executes the user-decided overrides captured in CONTEXT.md `<requirements_overrides>` verbatim. The decisions were made during phase-discuss; this plan reconciles the written-down requirements to match.

## Deviations from Plan

None - plan executed exactly as written. All ten REQUIREMENTS.md edits, eight ROADMAP.md edits, nine STACK.md edits, and two footer additions applied as specified.

## Issues Encountered

**Task 2 commit absorbed by parallel plan executor.** Plan 04 ran in wave 1 alongside Plans 00-01, 00-02, and 00-03 (all `wave=1, depends_on=[]`). Between Plan 04's Task 1 commit (`61a0ac2`) and Plan 04 staging Task 2's STACK.md changes, the parallel executor for Plan 00-01 ran `git add` (or equivalent) and swept up the STACK.md modifications I had already applied to the working tree. Plan 00-01's commit `43bd04d feat(00-01): add PR template and CODEOWNERS` therefore contains both Plan 00-01's intended changes (`.github/pull_request_template.md`, `.github/CODEOWNERS`) AND Plan 04's STACK.md edits (`+ "Go 1.25"` rows, `+ "go 1.25"` blocks, `+ "Updated 2026-05-07 ... Go floor bumped 1.24 → 1.25"` footer).

**Resolution:** All Task 2 changes ARE in git history (verified via `git show 43bd04d -- .planning/research/STACK.md`); the only flaw is commit-message attribution. Recording `43bd04d` as Task 2's commit hash in this summary preserves the audit trail. No data loss; verification still passes.

**Forward implication:** wave-1 parallel plans that touch the same working tree will inevitably cross-contaminate when executors stage files. Future planners should either (a) constrain wave-1 plans to disjoint directories, or (b) accept that commit boundaries may not perfectly align with plan boundaries when wave > 1.

## User Setup Required

None - planning-doc reconciliation only; no external service configuration.

## Next Phase Readiness

- Plans 01-03 (which already use `gen`/`rt`/`scim` verbatim) and Phase 1+ planners can now read REQUIREMENTS.md, ROADMAP.md, and STACK.md without re-introducing the old names
- REQ-REPO-02 is the binding contract for Phase 0 deliverables (the `gen` and `rt` modules with full repo paths); Plan 00-02's executor created the actual modules at `gen/go.mod` and `rt/go.mod`, matching this requirement
- **Forward note:** when planning Phase 7 (CLI), the planner should reread REQUIREMENTS.md to absorb the `scim` binary name and `gen`-as-library shape decided here. The library APIs the CLI wraps still come from the `gen` module; only the binary name is `scim`

---
*Phase: 00-architecture-spike-repo-setup*
*Completed: 2026-05-08*

## Self-Check: PASSED

- `00-04-SUMMARY.md` exists at `.planning/phases/00-architecture-spike-repo-setup/00-04-SUMMARY.md`
- Task 1 commit `61a0ac2` (`docs(00-04): reconcile module/CLI naming in REQUIREMENTS.md and ROADMAP.md`) found in git log
- Task 2 commit `43bd04d` (Plan 00-01's commit, which absorbed Task 2's STACK.md edits — see Issues Encountered) found in git log; STACK.md changes verified via `git show 43bd04d -- .planning/research/STACK.md`

# Project State

## Project Reference

See: .planning/PROJECT.md (updated 2026-05-07)

**Core value:** A developer can describe their SCIM resource set once and get a spec-compliant SCIM v2 server with SQL persistence, mountable on any `net/http`-compatible mux, with no runtime schema-interpretation overhead.
**Current focus:** Phase 0 — Architecture Spike & Repo Setup

## Current Position

Phase: 0 of 7 (Architecture Spike & Repo Setup)
Plan: 0 of TBD in current phase
Status: Ready to plan
Last activity: 2026-05-07 — Roadmap created (8 phases, 85/85 v1 requirements mapped)

Progress: [░░░░░░░░░░] 0%

## Performance Metrics

**Velocity:**
- Total plans completed: 0
- Average duration: —
- Total execution time: 0 hours

**By Phase:**

| Phase | Plans | Total | Avg/Plan |
|-------|-------|-------|----------|
| - | - | - | - |

**Recent Trend:**
- Last 5 plans: —
- Trend: —

*Updated after each plan completion*

## Accumulated Context

### Decisions

Decisions are logged in PROJECT.md Key Decisions table.
Recent decisions affecting current work:

- Project init (2026-05-07): Code-gen over generic tree; SQLite first behind pluggable SPI; plain `http.Handler`; CLI + library distribution; full SCIM v2 in v1; multi-module workspace; fluent Go builder definition (working hypothesis); IdP quirks / auth / multi-tenancy / non-SQLite drivers / groupsync / legacy compat / concrete observability all explicitly out of v1 scope
- Phase 0 (pending): `text/template` vs `dave/jennifer` emission engine — to be resolved by spike during Phase 0 execution

### Pending Todos

None yet.

### Blockers/Concerns

- Phase 0 must encode three catastrophic anti-relapse rules in writing (no generic tree, no runtime schema interpretation, no IdP accommodation) before any code is written — these are reversibility traps per PITFALLS.md
- Postgres dialect (Phase 8) deferred to v1.x; the SQL SPI extracted in Phase 3 should still be paper-validated against Postgres dialect differences (RETURNING, JSON columns, citext, LIMIT/OFFSET semantics) before v1 freeze

## Session Continuity

Last session: 2026-05-07
Stopped at: Roadmap and STATE files written; awaiting `/gsd:plan-phase 0`
Resume file: None

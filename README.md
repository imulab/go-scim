# go-scim

A code-generation toolkit for building SCIM v2 protocol servers in Go: describe your SCIM resource set once and get a spec-compliant SCIM v2 server with SQL persistence, mountable on any `net/http`-compatible mux, with no runtime schema-interpretation overhead.

## Status

Pre-alpha. Phase 0 in progress. See `.planning/ROADMAP.md`.

## Modules

This repository is a multi-module Go workspace:

- **`gen/`** — `github.com/imulab/go-scim/gen`: the build-time generator (parses definitions, walks the IR, runs emitters, writes files).
- **`rt/`** — `github.com/imulab/go-scim/rt`: the runtime support library imported by generated servers (SQL SPI, observability SPIs, PATCH engine, filter parser, common HTTP plumbing).

The two modules are linked by a `go.work` file at the repo root. `go.work` is gitignored (see REPO-01) — every contributor recreates it locally. CI builds both modules in workspace mode AND in isolation (`GOWORK=off`) to catch missing `require` lines that a workspace would mask.

## Building from a fresh clone

```bash
# Recreate the workspace (go.work is gitignored per REPO-01)
go work init
go work use ./gen ./rt

# Workspace build (explicit module prefixes — repo root is not itself a module,
# so plain `go build ./...` will not match; enumerate module trees instead)
go build ./gen/... ./rt/...

# Or isolated build (matches CI's GOWORK=off job)
(cd gen && GOWORK=off go build ./...)
(cd rt  && GOWORK=off go build ./...)
```

## Contributing

See [CONTRIBUTING.md](./CONTRIBUTING.md) for the architecture rules and workflow. Four enforcement levers keep the project on the rails: a PR template that asks reviewers about IdP-specific tolerance and runtime tree relapses, a CODEOWNERS file that pins sensitive paths, lefthook hooks that grep for forbidden symbols and vendor names locally, and a CI mirror of those grep gates so the rules are enforced even if a contributor skipped `lefthook install`. Run `lefthook install` once after cloning to wire up the local hooks.

## License

MIT — see [LICENSE](./LICENSE).

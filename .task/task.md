# `temporal sample` — Task

## Status

The CLI implementation is functionally complete: `temporal sample init` and
`temporal sample list` work for all 6 languages with tests against synthetic
HTTP fixtures. Code lives in `internal/temporalcli/`:

- `commands.yaml` — command declarations
- `commands.gen.go` — generated structs
- `commands.sample.go` — implementation
- `commands.sample_test.go` — unit tests

## Remaining work

See [spec.md](spec.md) for design constraints and rationale.
See [plan.md](plan.md) for concrete implementation details.

1. **CLI bug fixes** (4 items) — operator precedence bug, missing
   "sample not found" error, missing manifest fallback, `list` ignoring
   `sample_path`.

2. **Per-sample manifests** (65 total across 6 repos) — root
   `temporal-samples.yaml` and per-sample `temporal-sample.yaml` files
   covering hello, nexus, standalone-activity, agentic AI, saga,
   encryption, expense, DSL, and other core patterns.

3. **Integration tests** — one per language, downloading from real GitHub
   repos, verifying structure, and running each language's build command.
   Gated by build tag.

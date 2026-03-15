# `temporal sample` — Task

## Motivation

A new Temporal user wants to go from zero to a running sample in under two minutes.
Today that requires: finding the right samples repo, cloning it, understanding the
project structure, installing language-specific tooling, figuring out which commands
to run, and understanding the paths. `temporal sample init` should collapse this to
one command.

The purpose is for a user to quickly be able to get up and running. They run
`temporal server start-dev` (if they're not using cloud), then
`temporal sample init ...`, then run a couple of commands that are printed to the
screen and open the localhost or cloud URL that's printed to the screen to view
the UI.

The rule is that whatever you name, it has to be a directory containing a README.md,
i.e. a "sample" as Temporal loosely/implicitly defines.

This needs to work for all SDK languages that have samples repos: Go, Java, Python,
TypeScript, .NET, Ruby. PHP is deferred (the repo is a single Docker-based
application, not a collection of independent samples).

Remember that we're working in the CLI project that hitherto has known nothing of
any language other than Go. The samples repos are where we define or suggest the
toolchain for setting up a project in each language. But, in the interests of user
experience, we are prepared to hard-code some language logic in the CLI repo if it's
inevitable. Ideally the CLI acquires that logic from structured data
(`temporal-samples.yaml` manifests) in the samples repo.

## Status

The CLI implementation is functionally complete: `temporal sample init` and
`temporal sample list` work for all 6 languages with tests against synthetic
HTTP fixtures. Code lives in `internal/temporalcli/`:

- `commands.yaml` — command declarations
- `commands.gen.go` — generated structs
- `commands.sample.go` — implementation
- `commands.sample_test.go` — unit tests
- `commands.sample_integration_test.go` — integration tests (build tag `sample_integration`)

See [plan.md](plan.md) for design, architecture, and implementation details.
See [result.md](result.md) for what has been done and what remains.

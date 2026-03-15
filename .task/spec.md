# `temporal sample` — Specification for Remaining Work

## What's done

The CLI implementation is complete and tested against synthetic HTTP fixtures.
`temporal sample init` and `temporal sample list` work for all 6 languages
(Python, Go, TypeScript, Java, .NET, Ruby). See `commands.sample.go` and
`commands.sample_test.go` in `internal/temporalcli/`.

This spec covers what remains: CLI bug fixes, manifest content for each sample
repo, and integration tests that prove the scaffolded projects actually build.

---

## CLI bug fixes

### 1. Simplify manifest parsing logic

`commands.sample.go` line 314 has a confusing condition with wrong operator
precedence. The `||` branch with `HasSuffix` is dead code — it enters the
outer `if` but the inner `if` never matches, falling through to the skip
block. Simplify to a single condition.

### 2. Error when sample not found

If the sample name doesn't match anything in the tarball, the CLI silently
creates a directory containing only scaffold files. It should detect that
zero files were extracted and return an error.

### 3. Fallback when repo manifest is missing

If `fetchRepoManifest` gets a 404, fall back to extracting the directory as-is
(flat copy, no scaffold) and print a warning. This is critical for the
transition period while manifest PRs land across 6 repos — the CLI and the
manifests will not all merge atomically.

### 4. `list` should respect `sample_path`

Currently `list` scans the entire tarball for `temporal-sample.yaml` at any
depth. For repos with `sample_path` (Java, .NET), it should first fetch the
repo manifest and restrict scanning to that subtree. Without this, Java's
`list` will surface entries from the wrong locations.

---

## Manifest design per language

Each repo gets a root `temporal-samples.yaml` (scaffold templates, language
config) and a `temporal-sample.yaml` in each sample directory (description,
optional dependencies or extra template variables).

### Python

Scaffold generates `pyproject.toml`. Dependencies vary per sample — most need
only `temporalio>=1.23.0,<2`, but AI and encryption samples have extras. The
root `pyproject.toml` already declares these as `[project.optional-dependencies]`
groups; the per-sample manifest mirrors that data.

### Go

Scaffold generates `go.mod`. Import rewriting replaces
`github.com/temporalio/samples-go/<sample>` with `<name>/<sample>` in all
`.go` files. No per-sample dependencies — `go mod tidy` resolves them.

The generated `go.mod` will not have a `go.sum`. The user must run
`go mod tidy` before `go build`. This is a documented step, not automated
by the CLI.

**Nested sample issue**: `standalone-activity/helloworld` is two levels deep.
The CLI currently takes a single path component as the sample name. Either
support slash-separated sample paths, or place the manifest at
`standalone-activity/` and treat the whole directory as the sample.

### TypeScript

Empty scaffold — each sample already has its own `package.json`. Copied flat,
no nesting.

### Java

Scaffold generates `build.gradle`. Samples live deep in
`core/src/main/java/io/temporal/samples/<sample>/`. The CLI strips the first
path component (`core/`) and preserves the rest under the project root. Each
per-sample manifest supplies `sdk_version` as an extra template variable.

### .NET

Scaffold generates `Directory.Build.props` with framework target and SDK
references. The real `Directory.Build.props` includes analyzers and diagnostic
packages; the scaffold is intentionally minimal. `sample_path: src`.

### Ruby

Scaffold generates `Gemfile`. Ruby samples use `require_relative` for local
files, so no import rewriting is needed.

---

## Integration test requirements

Each language must have at least one integration test that:

1. Downloads from the real GitHub sample repo (on a feature branch initially,
   then `main` after manifest PRs merge).
2. Verifies the output directory structure (scaffold files present, sample
   files in the right place, manifests excluded).
3. Runs the language's build command to prove the scaffolded project is valid:

| Language | Build check |
|----------|-------------|
| Python | `uv sync` |
| Go | `go mod tidy && go build ./...` |
| TypeScript | `npm install && npx tsc --noEmit` |
| Java | `gradle compileJava` |
| .NET | `dotnet build` |
| Ruby | `bundle install` |

These tests require network access and language toolchains, so they must be
gated (build tag or env var) and not run in default CI.

---

## Scope boundaries

**In scope**: CLI bug fixes, manifests for ~10 samples per language (covering
hello, nexus, standalone-activity, agentic AI, saga, encryption, expense, DSL
where they exist), integration tests.

**Not in scope (later)**: Tarball caching, `--ref` flag, `temporal sample update`,
prerequisite checking (`which uv`), PHP support, config injection from
`temporal env`, Gradle wrapper in Java scaffold.

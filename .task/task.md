# `temporal sample` — Remaining Work

## Status

The CLI implementation is functionally complete: `temporal sample init` and `temporal sample list`
are wired up in commands.yaml, codegen is done, and all 10 tests pass against synthetic HTTP
fixtures. The code lives in:

- `internal/temporalcli/commands.yaml` — command declarations
- `internal/temporalcli/commands.gen.go` — generated structs
- `internal/temporalcli/commands.sample.go` — implementation
- `internal/temporalcli/commands.sample_test.go` — tests

## 1. Code issues to fix

### 1a. Dead/confusing manifest parsing logic (line 314)

```go
if relToSample == "temporal-sample.yaml" || strings.HasSuffix(rel, "/temporal-sample.yaml") && !parsedSampleManifest {
    if relToSample == "temporal-sample.yaml" {
```

Due to operator precedence, the `||` branch with `HasSuffix` can enter the outer `if`
but never does anything (the inner `if` only checks `relToSample ==`). The whole block
should just be:

```go
if relToSample == "temporal-sample.yaml" && !parsedSampleManifest {
    ...parse...
    continue
}
```

### 1b. No error when sample doesn't exist

If you run `temporal sample init python nonexistent_sample`, the tarball streams through,
nothing matches, and the CLI creates an empty directory with only scaffold files. It should
detect that zero files were extracted and return an error like
`sample "nonexistent_sample" not found in temporalio/samples-python`.

### 1c. No fallback when repo manifest is missing

The spec says: "Fallback when no manifest exists: extract the directory as-is, print the README,
warn that additional setup may be required." Currently `fetchRepoManifest` failure is fatal.
This matters for the transition period before all sample repos have manifests.

### 1d. `list` doesn't respect `sample_path`

The list command scans the entire tarball for any `temporal-sample.yaml` at any depth. For
repos with `sample_path` (Java, .NET), this could pick up unintended files. It should fetch
the repo manifest first and, if `sample_path` is set, restrict scanning to that subtree.
(Currently works because test fixtures are clean, but will matter with real repos.)

## 2. Add manifest files to sample repos

This is the bulk of remaining work. Each of the 6 sample repos needs:

1. A root-level `temporal-samples.yaml` (repo manifest with scaffold templates)
2. A `temporal-sample.yaml` in each sample directory (description + deps)

At minimum, a representative subset of samples per repo should have manifests for the
initial release, with the full set populated incrementally.

### Repos and what each needs

| Repo | Root manifest | Per-sample manifests | Notes |
|------|--------------|---------------------|-------|
| `samples-python` | `pyproject.toml` scaffold, dependencies template | ~38 dirs | Each needs `dependencies` listing |
| `samples-go` | `go.mod` scaffold, `rewrite_imports` | ~48 dirs | No per-sample deps needed |
| `samples-typescript` | Empty scaffold (`scaffold: {}`) | ~40 dirs | Simplest — just descriptions |
| `samples-java` | `build.gradle` scaffold, `sample_path` | ~30 packages under `core/src/.../samples/` | Needs `sdk_version` extra var |
| `samples-dotnet` | `Directory.Build.props` scaffold, `sample_path: src` | Projects under `src/` | |
| `samples-ruby` | `Gemfile` scaffold | ~12 dirs | |

The spec (now archived in git history) has worked examples for each language's manifest format.

## 3. Integration testing

After manifests are in the sample repos (even on a branch), test the real end-to-end flow:

```
temporal sample list python
temporal sample init python hello
cd hello && uv run hello/worker.py
```

Repeat for each language. This validates that the manifest content and the CLI's
interpretation of it actually produce runnable projects.

## 4. PR workflow

- **CLI PR**: Squash the implementation commits on the `init` branch, open PR against `main`.
  The CLI changes are self-contained and can merge independently of the sample repo changes
  (the fallback from 1c would make this smoother).
- **Sample repo PRs**: One PR per repo adding manifest files. These can land incrementally;
  the CLI gracefully handles missing manifests (after 1c is fixed).

## 5. Not in scope (later)

- Tarball caching
- `--ref` flag to pin a branch/tag
- `temporal sample update`
- Prerequisite checking (`which uv`, etc.)
- PHP support
- Config injection (Temporal address/namespace from `temporal env`)

# Result: `temporal sample` implementation

## What changed

Three files added/modified in `internal/temporalcli/`:

### `commands.yaml` — command declarations

Added `temporal sample`, `temporal sample init`, and `temporal sample list`
entries. The parent command has `docs.keywords` and `docs.description-header`
(required by the code generator for depth-2 commands). `sample list` uses
`maximum-args: 1` rather than `exact-args: 1` because cobra's `ExactArgs`
validation causes a double-call to the `Fail` handler when args are missing
(the arg validator fails, then the "unknown command" fallback fires since
`ActuallyRanCommand` was never set). The run method validates `len(args) < 1`
instead.

### `commands.gen.go` — regenerated

`make gen` produced `TemporalSampleCommand`, `TemporalSampleInitCommand`
(with `OutputDir string`), and `TemporalSampleListCommand`. The parent
`TemporalCommand` now wires in the sample subcommand.

### `commands.sample.go` — implementation (new file)

All logic in one file, no new packages. Key design:

- **Manifest types**: `repoManifest` (repo-level `temporal-samples.yaml`)
  and `sampleManifest` (per-sample `temporal-sample.yaml`). Parsed with
  `gopkg.in/yaml.v3` (already in go.mod).

- **Language → repo map**: Hardcoded `langRepos` for 6 languages.

- **URL construction**: `rawContentURL` and `tarballURL` with
  `TEMPORAL_SAMPLES_BASE_URL` env var override for testing.

- **`list` command**: Downloads tarball, streams through scanning for
  `temporal-sample.yaml` entries, prints aligned name+description table.

- **`init` command**: Parses args (2 positional or 1 GitHub URL), fetches
  repo manifest, streams tarball extracting sample files. Two extraction
  modes:
  - **Nested** (scaffold non-empty): sample files go under
    `<outputDir>/<sample>/` (or deeper for `sample_path`), scaffold files
    at project root, README promoted to root.
  - **Flat** (scaffold empty, e.g. TypeScript): sample files extracted
    directly into `<outputDir>/`.

- **Template expansion**: `strings.ReplaceAll` for `{{name}}`,
  `{{dependencies}}`, and extra keys from sample manifests. Dependencies
  are quoted on expansion.

- **Import rewriting**: Walks files matching glob, replaces old monorepo
  import prefix with new standalone module path.

- **GitHub URL parsing**: Extracts owner/repo/ref/sample from
  `https://github.com/OWNER/REPO/tree/REF/SAMPLE`.

## Test results

All 10 tests pass:

```
go test -run 'TestSample_' -count=1 -v ./internal/temporalcli/
```

## Notable decisions

1. Used `maximum-args` + runtime check instead of `exact-args` to avoid
   the cobra double-Fail bug (this is the first `exact-args` usage in the
   codebase; the framework's error handling at `commands.go:371-393` doesn't
   return after the first `Fail` call).

2. Dependencies are quoted during template expansion (`"dep1", "dep2"`)
   because YAML parsing strips the quotes from the manifest values.

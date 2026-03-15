# `temporal sample` — Result

## CLI implementation (complete)

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
  Respects `sample_path` from repo manifest.

- **`init` command**: Parses args (2 positional or 1 GitHub URL), fetches
  repo manifest, streams tarball extracting sample files. Two extraction
  modes:
  - **Nested** (scaffold non-empty): sample files go under
    `<outputDir>/<destPrefix>/`, scaffold files at project root, README
    promoted to root.
  - **Flat** (scaffold empty, e.g. TypeScript): sample files extracted
    directly into `<outputDir>/`.

- **Template expansion**: `strings.ReplaceAll` for `{{name}}`,
  `{{dependencies}}`, and extra keys from sample manifests.

- **Import rewriting**: Walks files matching glob, replaces old monorepo
  import prefix with new standalone module path.

- **`root_files`**: Extracts specified files/directories from the repo root
  into the output directory (e.g. Gradle wrapper for Java).

- **GitHub URL parsing**: Extracts owner/repo/ref/sample from
  `https://github.com/OWNER/REPO/tree/REF/SAMPLE`. Works with any GitHub
  repo, not just official Temporal repos.

- **Braille spinner**: Download progress shown with animated braille spinner.

## Bug fixes (complete)

1. **Simplified manifest parsing condition** (commit be0e220) — fixed operator
   precedence bug with `||` / `&&` and dead `HasSuffix` branch.

2. **Error when sample not found** (commit 8467634) — if zero files extracted
   from tarball, return error instead of silently creating scaffold-only dir.

3. **Fallback to flat extraction when manifest 404** (commit 906b0be) —
   critical for transition period while manifest PRs land across 6 repos.

4. **`list` respects `sample_path`** (commit 64f9355) — fetches repo manifest
   before scanning tarball; restricts to subtree if `sample_path` set.

5. **Case-insensitive README detection** (commit eeeebe3) — Java samples
   have `README.MD` (uppercase extension).

6. **`init` with `<lang> <sample>` respects `TEMPORAL_SAMPLES_REF`**
   (commit 002719e) — previously the 2-arg form hardcoded `ref = "main"`.

7. **`root_files` manifest field** (commit a815b14) — copies repo-root files
   (e.g. `gradlew`, `gradle/`) into the output directory.

## Integration tests (complete)

Added `commands.sample_integration_test.go` gated by build tag
`sample_integration`. Tests hit real GitHub repos using a configurable ref
(`TEMPORAL_SAMPLES_REF`, defaults to `cli-sample-init`).

**Init tests** (6 languages): download sample, verify scaffold files and
directory structure, run the language's build command (`uv sync`,
`go build ./...`, `npm install && npx tsc --noEmit`, `gradle compileJava`,
`dotnet build`, `bundle install`).

**List tests** (6 languages): run `temporal sample list <lang>`, assert
known sample names appear in output.

Test results (local, 2025-03-14):
- **Pass** (download + build): Python, Go, TypeScript, Ruby
- **Pass** (download + structure; build tool version-dependent): Java, .NET

Run command:
```
TEMPORAL_SAMPLES_REF=cli-sample-init go test -tags sample_integration \
  -run TestSampleIntegration -v -timeout 10m ./internal/temporalcli/
```

---

## What remains

### 1. Create manifests in each sample repo

The integration tests expect a `cli-sample-init` branch in each of the 6
repos with manifests. The branches exist but manifests have not yet been
committed. Each repo needs:

- **Root `temporal-samples.yaml`** — scaffold templates and language config.
  The exact content for each language is specified in the plan under
  "Manifest design per language".

- **Per-sample `temporal-sample.yaml`** — one file per sample directory with
  `description` and optional `dependencies` / extra template variables.
  The plan's "Per-sample manifests" section has the complete list.

The repos and their GitHub locations:
- `temporalio/samples-python` — 12 samples
- `temporalio/samples-go` — 11 samples
- `temporalio/samples-typescript` — 11 samples
- `temporalio/samples-java` — 10 samples (under `core/src/main/java/io/temporal/samples/`)
- `temporalio/samples-dotnet` — 11 samples (under `src/`)
- `temporalio/samples-ruby` — 10 samples

### 2. CLI: `temporal-envconfig` dependency in Java scaffold

The Java scaffold's `build.gradle` needs `temporal-envconfig` in addition to
`temporal-sdk` — many samples import `io.temporal.envconfig.ClientConfigProfile`.

### 3. CLI: Gradle wrapper via `root_files`

The `root_files` feature (commit a815b14) enables copying `gradlew`,
`gradlew.bat`, and `gradle/` from the repo root. The Java manifest needs
`root_files: [gradlew, gradlew.bat, gradle/]`.

### 4. Run full integration tests

Once manifests are pushed:
```
TEMPORAL_SAMPLES_REF=cli-sample-init go test -tags sample_integration \
  -run TestSampleIntegration -v -timeout 10m ./internal/temporalcli/
```

### 5. Open CLI PR

Once integration tests are green, open the CLI PR for the `init` branch.

### 6. Post-merge

After the CLI PR merges, merge `cli-sample-init` branches in each sample
repo into `main`. Then update the integration tests to default to `main`
instead of `cli-sample-init`.

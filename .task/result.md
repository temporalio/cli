# `temporal sample` — Result

## What was done

### CLI bug fix: `list` respects `sample_path` (commit 64f9355)

`temporal sample list` now fetches the repo manifest before scanning the
tarball. If the manifest specifies `sample_path` (Java, .NET), only
`temporal-sample.yaml` entries under that subtree are considered. A 404 on
the manifest is tolerated — the command falls back to scanning everywhere.

Added `TestSample_List_SamplePath` unit test with a synthetic tarball
containing entries at two depths, asserting only the ones under
`sample_path` are listed.

### Integration tests (commit 21b54b4)

Added `commands.sample_integration_test.go` gated by build tag
`sample_integration`. Tests hit real GitHub repos using a configurable ref
(`TEMPORAL_SAMPLES_REF`, defaults to `sample-init`).

**Init tests** (6 languages): download sample, verify scaffold files and
directory structure, run the language's build command (`uv sync`,
`go build ./...`, `npm install && npx tsc --noEmit`, `gradle compileJava`,
`dotnet build`, `bundle install`).

**List tests** (6 languages): run `temporal sample list <lang>`, assert
known sample names appear in output.

Run command:
```
TEMPORAL_SAMPLES_REF=sample-init go test -tags sample_integration -run TestSampleIntegration -v ./internal/temporalcli/
```

### Bug fixes #1–3 (already done in prior commits)

These were completed before this session:
- Simplified manifest parsing condition (commit be0e220)
- Error when sample not found in tarball (commit 8467634)
- Fallback to flat extraction when repo manifest returns 404 (commit 906b0be)

---

## What remains

### 1. Create `sample-init` branches in each sample repo

The integration tests expect a `sample-init` branch in each of the 6 repos
with manifests. Each repo needs:

- **Root `temporal-samples.yaml`** — scaffold templates and language config.
  The exact content for each language is specified in the plan under
  "Manifest design per language".

- **Per-sample `temporal-sample.yaml`** — one file per sample directory with
  `description` and optional `dependencies` / extra template variables.
  The plan's "Per-sample manifests" section has the complete list of
  directories, descriptions, and dependencies for all 65 samples across
  6 repos.

The repos and their GitHub locations:
- `temporalio/samples-python` — 12 samples
- `temporalio/samples-go` — 11 samples
- `temporalio/samples-typescript` — 11 samples
- `temporalio/samples-java` — 10 samples (under `core/src/main/java/io/temporal/samples/`)
- `temporalio/samples-dotnet` — 11 samples (under `src/`)
- `temporalio/samples-ruby` — 10 samples

For each repo: create branch `sample-init` from `main`, add root manifest +
all per-sample manifests in a single commit, push. The exact YAML content is
in the plan — use it verbatim. Pay particular attention to:

- **Go**: `rewrite_imports` field in root manifest.
- **Java**: `sample_path: core/src/main/java/io/temporal/samples` and
  `sdk_version` in each per-sample manifest's extra vars.
- **TypeScript**: `scaffold: {}` (empty — each sample has its own
  `package.json`).
- **.NET**: `sample_path: src`.

### 2. Run integration tests against `sample-init` branches

Once manifests are pushed:
```
TEMPORAL_SAMPLES_REF=sample-init go test -tags sample_integration \
  -run TestSampleIntegration -v -timeout 10m ./internal/temporalcli/
```

Debug failures by examining the actual directory structure in the temp dir
(add `t.Log(os.Getwd())` if needed). The most likely failures:
- Manifest YAML typos (wrong indentation, missing fields)
- Sample directory names not matching expectations in the test
- Build tool not installed on the machine

### 3. Open CLI PR

Once integration tests are green, open the CLI PR for the `init` branch.

### 4. Post-merge

After the CLI PR merges, merge `sample-init` branches in each sample repo
into `main`. Then update the integration tests to default to `main` instead
of `sample-init` (change the default in `samplesRef()`).

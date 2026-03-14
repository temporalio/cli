# Implementation Plan: `temporal sample`

## Overview

Add `temporal sample init` and `temporal sample list` commands to the CLI.
All languages except PHP are supported from the start: **Python, TypeScript,
Go, Java, .NET, Ruby.** The CLI is a manifest interpreter: language-specific
knowledge lives in sample repo manifests, not the CLI.

## Architecture

```
commands.yaml           ─── declares temporal sample {init,list} commands
commands.gen.go         ─── generated structs + Cobra wiring
commands.sample.go      ─── run() implementations for init and list
commands.sample_test.go ─── tests (already committed, currently failing)
```

No new packages. Everything lives in `internal/temporalcli/`.

## Steps

### 1. Add command declarations to `commands.yaml`

Add three entries after the `temporal workflow unpause` block, before
`option-sets:`:

```yaml
- name: temporal sample
  summary: Initialize or list sample projects
  description: |
    Create a new project from a Temporal sample, or browse
    available samples for a given language.

    List all Python samples:

        temporal sample list python

    Initialize a sample:

        temporal sample init python hello

    Initialize from a GitHub URL:

        temporal sample init \
            https://github.com/temporalio/samples-python/tree/main/hello

- name: temporal sample init
  summary: Create a project from a Temporal sample
  description: |
    Download a sample from the Temporal samples repository and set up
    a standalone project. Provide a language and sample name, or a
    GitHub URL:

        temporal sample init python hello

        temporal sample init \
            https://github.com/temporalio/samples-python/tree/main/hello

    Use "--output-dir" to control where the project is created.
    Defaults to "./<sample-name>/" in the current directory.
  maximum-args: 2
  options:
    - name: output-dir
      type: string
      description: Directory to create the project in.

- name: temporal sample list
  summary: List available Temporal samples for a language
  description: |
    Download and scan the samples repository for a given language,
    then print each sample's name and description:

        temporal sample list python
  exact-args: 1
```

### 2. Run `make gen`

Regenerates `commands.gen.go` with the new structs:
- `TemporalSampleCommand`
- `TemporalSampleInitCommand` (with `OutputDir string`)
- `TemporalSampleListCommand`

### 3. Create `commands.sample.go`

This file contains all implementation logic. Key types and functions:

#### Types

```go
// repoManifest is the repo-level temporal-samples.yaml.
type repoManifest struct {
    Version        int               `yaml:"version"`
    Language       string            `yaml:"language"`
    Repo           string            `yaml:"repo"`
    Scaffold       map[string]string `yaml:"scaffold"`
    RewriteImports *rewriteRule      `yaml:"rewrite_imports"`
    SamplePath     string            `yaml:"sample_path"`
}

type rewriteRule struct {
    From string `yaml:"from"`
    Glob string `yaml:"glob"`
}

// sampleManifest is the per-sample temporal-sample.yaml.
type sampleManifest struct {
    Description  string            `yaml:"description"`
    Dependencies []string          `yaml:"dependencies"`
    Extra        map[string]string `yaml:",inline"`
}
```

#### Language → repo mapping (hardcoded)

```go
var langRepos = map[string]string{
    "go":         "temporalio/samples-go",
    "java":       "temporalio/samples-java",
    "python":     "temporalio/samples-python",
    "typescript": "temporalio/samples-typescript",
    "dotnet":     "temporalio/samples-dotnet",
    "ruby":       "temporalio/samples-ruby",
}
```

#### URL construction

Two GitHub endpoints:
- **Raw content**: `https://raw.githubusercontent.com/{owner}/{repo}/{ref}/{path}`
- **Tarball**: `https://codeload.github.com/{owner}/{repo}/tar.gz/{ref}`

Both share a single base URL override via `TEMPORAL_SAMPLES_BASE_URL` env var
(for testing). When set, both endpoints use `{base}/{owner}/{repo}/...` paths.

```go
func samplesBaseURL() string {
    if u := os.Getenv("TEMPORAL_SAMPLES_BASE_URL"); u != "" {
        return u
    }
    return ""
}

func rawContentURL(repo, ref, path string) string {
    if base := samplesBaseURL(); base != "" {
        return base + "/" + repo + "/" + ref + "/" + path
    }
    return "https://raw.githubusercontent.com/" + repo + "/" + ref + "/" + path
}

func tarballURL(repo, ref string) string {
    if base := samplesBaseURL(); base != "" {
        return base + "/" + repo + "/tar.gz/" + ref
    }
    return "https://codeload.github.com/" + repo + "/tar.gz/" + ref
}
```

#### `temporal sample list` implementation

```
func (c *TemporalSampleListCommand) run(cctx *CommandContext, args []string) error
```

1. Resolve `args[0]` to repo via `langRepos`.
2. Download tarball from `tarballURL(repo, "main")`.
3. Scan tar entries for paths matching `*/temporal-sample.yaml` (one
   directory deep, or under `sample_path` if the repo manifest specifies one).
4. Parse each as `sampleManifest`, collect name + description.
5. Sort by name, print as aligned two-column table to `cctx.Options.Stdout`.
6. Print repo URL at the end.

#### `temporal sample init` implementation

```
func (c *TemporalSampleInitCommand) run(cctx *CommandContext, args []string) error
```

1. Parse args: if 1 arg starting with `https://`, parse as GitHub URL
   extracting owner/repo/ref/sample. Derive language from repo name.
   If 2 args, treat as language + sample; resolve repo, use ref=main.
   If 0 args, return usage error.
2. Fetch repo manifest from `rawContentURL(repo, ref, "temporal-samples.yaml")`.
   Parse as `repoManifest`.
3. Determine output directory: `c.OutputDir` if set, else `./<sample>`.
   Error if it already exists.
4. Download tarball from `tarballURL(repo, ref)`.
5. Determine the sample's path in the tarball. If `repoManifest.SamplePath`
   is set, the sample lives at `<sample_path>/<sample>/`; otherwise `<sample>/`.
6. Determine extraction mode:
   - If `scaffold` is non-empty: **nested** — create `<outputDir>/<sample>/`
     and extract sample files there. Copy `README.md` to `<outputDir>/`.
   - If `scaffold` is empty: **flat** — extract directly into `<outputDir>/`.
7. Stream through tarball, extracting files whose path is within the sample
   directory. Skip `temporal-sample.yaml` and `temporal-samples.yaml`.
8. For each scaffold template in `repoManifest.Scaffold`:
   - Expand `{{name}}` → sample directory name.
   - Expand `{{dependencies}}` → comma-joined, quoted dependency strings
     from `sampleManifest.Dependencies`.
   - Expand any other `{{key}}` from `sampleManifest.Extra`.
   - Write to `<outputDir>/<filename>`.
9. If `repoManifest.RewriteImports` is set, walk all extracted files matching
   the glob and replace the old import prefix with the new module path.
   For Go: `github.com/temporalio/samples-go/<sample>` → `<name>/<sample>`.
10. Print instructions to stdout:
    ```
    Downloading <sample> from <repo>...
    Created ./<outputDir>/

      cd <outputDir>
      cat README.md
    ```

#### Import rewriting (Go)

When `rewrite_imports` is present in the repo manifest:

```go
func rewriteImports(dir string, rule rewriteRule, oldPrefix, newPrefix string) error
```

Walk files matching `rule.Glob` under `dir`. For each file, replace all
occurrences of `oldPrefix` with `newPrefix`. This handles both direct imports
(`"github.com/temporalio/samples-go/helloworld"`) and nested package imports
(`"github.com/temporalio/samples-go/helloworld/worker"`).

The old prefix is `rule.From + "/" + sample`. The new prefix is
`name + "/" + sample` (where `name` is the project directory name, which
equals the sample name by default).

#### Java: deep `sample_path`

When `sample_path` is set (e.g. `core/src/main/java/io/temporal/samples`),
the sample lives at `<sample_path>/<sample>/` in the tarball. The CLI:

1. Extracts sample files to `<outputDir>/src/main/java/io/temporal/samples/<sample>/`
   — preserving the Java source tree structure under the project root.
2. The `src/main/java/...` prefix is derived from `sample_path` by stripping
   the module prefix (`core/`). More precisely: `sample_path` minus the first
   path component gives the Java source tree layout.

The per-sample `temporal-sample.yaml` can supply extra template variables
(e.g. `sdk_version`) that the scaffold template references.

#### .NET: `sample_path: src`

Similar to Java but simpler. Samples live at `src/<SampleName>/` in the repo.
The CLI extracts to `<outputDir>/<SampleName>/` (nested under project root).
The scaffold generates `Directory.Build.props` at the project root, providing
the shared SDK version and framework target that the repo-level file normally
supplies.

#### GitHub URL parsing

```go
// parseGitHubURL extracts (repo, ref, sample) from a URL like
// https://github.com/temporalio/samples-python/tree/main/hello
func parseGitHubURL(rawURL string) (repo, ref, sample string, err error)
```

Path segments: `/{owner}/{repo}/tree/{ref}/{sample}`.
Derive language from repo name: `samples-python` → `python`.

#### Tarball reading

```go
// downloadTarball downloads and opens a gzipped tarball, returning a
// *tar.Reader. The caller must close the response body.
func downloadTarball(ctx context.Context, url string) (io.ReadCloser, *tar.Reader, error)

// stripTarPrefix removes the top-level GitHub prefix directory from a tar
// entry name, returning the path relative to the repo root.
func stripTarPrefix(name string) string
```

GitHub tarballs have a single top-level directory (`{owner}-{repo}-{shortsha}/`).
`stripTarPrefix` removes it by finding the first `/` and returning everything after.

### 4. Add manifests to sample repos

Each samples repo (on branch `cli-sample-init`) gets:
- `temporal-samples.yaml` at the repo root
- `temporal-sample.yaml` in each sample directory (at minimum a representative
  subset for initial testing; the full set can be populated incrementally)

The Python repo already has a `temporal-samples.yaml` and two per-sample
manifests. The remaining repos need manifests created from scratch.

**Repo-level manifests (summary):**

| Repo | scaffold files | sample\_path | rewrite\_imports |
|------|---------------|-------------|-----------------|
| samples-python | `pyproject.toml` | — | — |
| samples-typescript | (empty) | — | — |
| samples-go | `go.mod` | — | `from: github.com/temporalio/samples-go`, `glob: "*.go"` |
| samples-java | `build.gradle` | `core/src/main/java/io/temporal/samples` | — |
| samples-dotnet | `Directory.Build.props` | `src` | — |
| samples-ruby | `Gemfile` | — | — |

Per-sample manifests contain `description` and optionally `dependencies`
or extra template variables (e.g. `sdk_version` for Java).

### 5. Dependencies

Add `gopkg.in/yaml.v3` for manifest parsing (check if already in `go.mod`
— it's likely already there via transitive deps).

No other new dependencies. The implementation uses:
- `archive/tar`, `compress/gzip` (stdlib)
- `net/http` (stdlib)
- `strings.ReplaceAll` for `{{var}}` expansion
- `path/filepath.WalkDir` + `strings.ReplaceAll` for import rewriting

For template expansion: `strings.ReplaceAll` is simpler and sufficient.
The spec defines only `{{name}}`, `{{dependencies}}`, and per-sample extra
keys. No control flow, no escaping. A simple replace loop is preferable
to `text/template`.

### 6. Verify tests pass

Run:
```
go test -run 'TestSample_' -count=1 -v ./internal/temporalcli/
```

Expected: all 10 tests pass (8 feature tests + 2 error-case tests).

### 7. Run full test suite, linter, type checker

```
make gen
go vet ./...
go test ./internal/temporalcli/ -count=1
```

### 8. Commit

Single commit for CLI changes: YAML declarations + generated code +
implementation file. Separate commits for each samples repo manifest.

## Key design decisions

1. **No interactive mode.** The spec is explicit: CLIs must be scriptable.
   Missing args → usage help + non-zero exit.

2. **Env var for test URL override.** `TEMPORAL_SAMPLES_BASE_URL` replaces
   both `raw.githubusercontent.com` and `codeload.github.com`. Simpler than
   injecting an HTTP client or adding hidden flags.

3. **Streaming tarball extraction.** Don't download-then-extract. Stream the
   tar reader, writing only matching entries. Keeps memory bounded.

4. **Nested vs flat extraction.** Determined by whether `scaffold` is
   non-empty. When scaffolding, sample files go one level deeper to preserve
   absolute imports. When empty (TypeScript), sample IS the project.

5. **README promotion.** In nested mode, README.md is copied to the project
   root (not the nested package dir). This is the first thing users should
   read, and it matches the spec's example output.

6. **`strings.ReplaceAll` over `text/template`.** The template language is
   trivially simple (just `{{name}}` and `{{dependencies}}`). A replace loop
   is easier to understand, has no edge cases around delimiters, and needs
   no documentation for sample repo maintainers.

7. **All languages from day one.** Each language exercises a different
   combination of manifest features (scaffold, sample_path, rewrite_imports,
   flat vs nested). Implementing all six validates that the manifest design
   generalises rather than accumulating special cases.

## What this plan does NOT cover (later)

- Caching downloaded tarballs
- `temporal sample update`
- `--ref` flag to pin a branch/tag
- Prerequisite checking (`which uv`)
- PHP support

## Running the tests

From the repo root (`/Users/dan/worktrees/cli/init/cli`):

```bash
# Run only the sample command tests
go test -run 'TestSample_' -count=1 -v ./internal/temporalcli/

# After implementation, all 10 must pass:
#   TestSample_List                         (list discovers samples from tarball)
#   TestSample_Init_Python                  (scaffold + nested extraction)
#   TestSample_Init_TypeScript_FlatCopy     (empty scaffold → flat copy)
#   TestSample_Init_Go_ImportRewrite        (scaffold + import rewriting)
#   TestSample_Init_Java                    (scaffold + deep sample_path)
#   TestSample_Init_DotNet                  (scaffold + sample_path: src)
#   TestSample_Init_Ruby                    (scaffold, no import rewrite)
#   TestSample_Init_GitHubURL               (URL parsing variant)
#   TestSample_Init_NoArgs                  (error: missing args)
#   TestSample_List_NoArgs                  (error: missing args)
```

Without the implementation, the first 8 tests fail with "unknown command"
(the `temporal sample` subcommand does not exist). The last 2 pass trivially
because unknown commands already produce errors.

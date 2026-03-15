# `temporal sample` — Plan

## Prior art

| Tool | Mechanism | Post-download | Metadata | Docs |
|------|-----------|---------------|----------|------|
| **Stripe CLI** | `go-git` clone to cache | `.env` injection of API keys | Two-tier: central `samples.json` + per-sample `.cli.json` | [wiki](https://github.com/stripe/stripe-cli/wiki/Samples-command), [gallery](https://docs.stripe.com/samples), [source](https://github.com/stripe/stripe-cli/tree/master/pkg/samples) |
| **`sam init`** | Git clone | Full project + tests | GitHub repo ([templates](https://github.com/aws/aws-sam-cli-app-templates)) + custom URLs | [docs](https://docs.aws.amazon.com/serverless-application-model/latest/developerguide/using-sam-cli-init.html) |
| **`pulumi new`** | Git clone | Dep install + stack init | GitHub repo ([templates](https://github.com/pulumi/templates)) + arbitrary git URLs | [docs](https://www.pulumi.com/docs/iac/cli/commands/pulumi_new/) |
| **`dotnet new`** | NuGet download | Dep restore | `.template.config/template.json`; `dotnet new search` queries NuGet.org | [docs](https://learn.microsoft.com/en-us/dotnet/core/tools/dotnet-new) |
| **Spring Initializr** | HTTP API → zip | Full Maven/Gradle project | Centralized server API; web UI at start.spring.io | [docs](https://docs.spring.io/spring-boot/cli/using-the-cli.html) |
| **create-next-app** | GitHub tarball + tar filter | Installs deps, copies `.gitignore` | None; checks dir existence via GitHub API | — |
| **degit** | GitHub tarball + tar filter | None (pure copy) | None | — |
| **gonew** | `go mod download` | Rewrites module path + imports | None; any Go module is a template | — |
| **cargo-generate** | Git clone | Liquid template substitution | `cargo-generate.toml` in template | — |

**Key observations:**

- **Stripe CLI** is the closest structural analog (SDK vendor, multi-language, CLI-driven
  sample scaffolding) and the only one with two-tier metadata.
- Two download strategies dominate: **GitHub tarballs** (create-next-app, degit) and
  **git clone** (Stripe, cargo-generate). Tarballs are simpler for one-shot extraction.
- **Competitors:** Restate, Inngest, and Dagger have no CLI sample scaffolding.

### Stripe CLI — closest prior art

[CLI source](https://github.com/stripe/stripe-cli/tree/master/pkg/samples) ·
[Registry](https://github.com/stripe-samples/samples-list) ·
[Wiki](https://github.com/stripe/stripe-cli/wiki/Samples-command)

Two-tier registry: central `samples.json` (name → repo URL) + per-sample `.cli.json`
with integrations, client/server language lists, and post-install config. Downloads
via `go-git` clone to `~/.config/stripe/samples-cache/`. Injects API keys into `.env`.

**Key structural difference:** Stripe has one sample per repo with multiple language
implementations inside. Temporal has one repo per language with multiple samples inside.
This inverts the registry: Stripe's index points to repos; ours points to directories
within repos.

**Adopted from Stripe:** Two-tier metadata (repo-level + per-sample).
**Different from Stripe:** Our metadata lives in the samples repos (not a separate
registry repo). We need project scaffolding (Stripe samples are already standalone).

---

## Architecture

Two halves:

1. **Samples-side**: A repo-level manifest (`temporal-samples.yaml`) at each repo
   root with language identifier and scaffold templates, plus a per-sample manifest
   (`temporal-sample.yaml`) in each sample directory with description and dependencies.
2. **CLI-side**: Downloads sample code, reads both manifests, generates scaffolding,
   prints the README. The CLI is a **manifest interpreter**, not a language expert.

This keeps language-specific knowledge in the samples repos (where maintainers know
their structure best) and out of the CLI (where release cadence is slower).

### Temporal samples repos — current state

| Language | Repo | Build system | Sample unit | Standalone? | Run commands |
|----------|------|-------------|-------------|-------------|--------------|
| **Python** | `samples-python` | Root `pyproject.toml` (uv/hatch) | Top-level dirs (~38) | No — need generated `pyproject.toml` | `uv run <sample>/worker.py` |
| **Go** | `samples-go` | Root `go.mod` | Top-level dirs (~48) | No — need `go.mod` + import rewrite | `go run <sample>/worker/main.go` |
| **TypeScript** | `samples-typescript` | pnpm workspace | Top-level dirs (~40), each with own `package.json` | **Yes** | `npm run start` / `npm run workflow` |
| **Java** | `samples-java` | Root Gradle multi-module | Classes in `core/src/.../samples/` | No — complex; see below | `./gradlew execute -PmainClass=...` |
| **.NET** | `samples-dotnet` | `.sln` + `Directory.Build.props` | Projects in `src/` | Almost — need `Directory.Build.props` content | `dotnet run` from sample dir |
| **Ruby** | `samples-ruby` | Root `Gemfile` | Top-level dirs (~12) | No — need generated `Gemfile` | `bundle exec ruby worker.rb` |

### Commands

```
temporal sample init <language> <sample> [--output-dir DIR]
temporal sample init <github-url> [--output-dir DIR]
temporal sample list <language>
```

Missing arguments produce usage help and exit non-zero. No interactive mode —
CLIs must be scriptable.

### Extraction

For languages where samples use absolute imports (Python, Go), the sample
directory is nested inside the project to preserve the import structure:

```
hello_standalone_activity/          ← project root (created by CLI)
  pyproject.toml                    ← generated from scaffold template
  README.md                         ← copied from sample dir
  hello_standalone_activity/        ← package dir (extracted from repo)
    __init__.py
    my_activity.py
    worker.py
```

For TypeScript (empty `scaffold`), the sample directory IS the project — copied
flat, no nesting.

### Download mechanism

1. Fetch repo-level manifest via GitHub raw content API
2. Parse manifest, validate sample exists
3. Download repo tarball from `codeload.github.com`, filter to sample directory
4. Apply scaffold templates, write output

**Fallback when no manifest exists**: extract the directory as-is, print the README,
warn that additional setup may be required.

### What the CLI hardcodes

1. Language → repo mapping (`python` → `temporalio/samples-python`)
2. GitHub URL parsing
3. Manifest schema (`temporal-samples.yaml` v1)
4. `{{variable}}` template substitution
5. Default ref: `main`

---

## CLI implementation

```
commands.yaml           ─── declares temporal sample {init,list} commands
commands.gen.go         ─── generated structs + Cobra wiring
commands.sample.go      ─── run() implementations for init and list
commands.sample_test.go ─── tests (already committed, currently failing)
```

No new packages. Everything lives in `internal/temporalcli/`.

### Types

```go
// repoManifest is the repo-level temporal-samples.yaml.
type repoManifest struct {
    Version        int               `yaml:"version"`
    Language       string            `yaml:"language"`
    Repo           string            `yaml:"repo"`
    Scaffold       map[string]string `yaml:"scaffold"`
    RewriteImports *rewriteRule      `yaml:"rewrite_imports"`
    SamplePath     string            `yaml:"sample_path"`
    RootFiles      []string          `yaml:"root_files"`
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

### Language → repo mapping (hardcoded)

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

### URL construction

Two GitHub endpoints:
- **Raw content**: `https://raw.githubusercontent.com/{owner}/{repo}/{ref}/{path}`
- **Tarball**: `https://codeload.github.com/{owner}/{repo}/tar.gz/{ref}`

Both share a single base URL override via `TEMPORAL_SAMPLES_BASE_URL` env var
(for testing).

### `temporal sample list` implementation

1. Resolve `args[0]` to repo via `langRepos`.
2. Download tarball from `tarballURL(repo, ref)`.
3. Scan tar entries for `temporal-sample.yaml` (respecting `sample_path`).
4. Parse each as `sampleManifest`, collect name + description.
5. Sort by name, print as aligned two-column table.
6. Print repo URL at the end.

### `temporal sample init` implementation

1. Parse args: if 1 arg starting with `https://`, parse as GitHub URL
   extracting owner/repo/ref/sample. If 2 args, treat as language + sample;
   resolve repo, use ref=main.
2. Fetch repo manifest from raw content URL.
3. Determine output directory: `c.OutputDir` if set, else `./<sample>`.
4. Download tarball, stream through extracting sample files.
5. Two extraction modes:
   - **Nested** (scaffold non-empty): sample files go under
     `<outputDir>/<destPrefix>/`, scaffold files at project root, README
     promoted to root.
   - **Flat** (scaffold empty): sample files extracted directly into
     `<outputDir>/`.
6. Expand `{{var}}` placeholders in scaffold templates.
7. If `rewrite_imports` is set, walk files matching glob and replace old
   monorepo import prefix with new module path.
8. If `root_files` is set, extract those paths from the tarball into the
   output directory.

### Import rewriting (Go)

When `rewrite_imports` is present in the repo manifest, walk files matching
`rule.Glob` under the output dir. For each file, replace all occurrences of
`rule.From + "/" + sample` with `name + "/" + sample`. This handles both
direct imports and nested package imports.

### Java: deep `sample_path`

When `sample_path` is set (e.g. `core/src/main/java/io/temporal/samples`),
the sample lives at `<sample_path>/<sample>/` in the tarball. The CLI
extracts to `<outputDir>/src/main/java/io/temporal/samples/<sample>/` —
preserving the Java source tree, stripping the first path component (`core/`).

### .NET: `sample_path: src`

Similar to Java but simpler. Samples live at `src/<SampleName>/` in the repo.
The CLI extracts to `<outputDir>/<SampleName>/` (nested under project root).

---

## Manifest design per language

Each repo gets a root `temporal-samples.yaml` (scaffold templates, language
config) and a `temporal-sample.yaml` in each sample directory (description,
optional dependencies or extra template variables).

| Language | scaffold | sample\_path | rewrite\_imports | root\_files | Extraction |
|----------|----------|-------------|-----------------|-------------|------------|
| Python | `pyproject.toml` | — | — | — | nested |
| Go | `go.mod` | — | yes | — | nested |
| TypeScript | empty | — | — | — | flat |
| Java | `build.gradle` | `core/src/main/java/io/temporal/samples` | — | `gradlew`, `gradlew.bat`, `gradle/` | nested (deep) |
| .NET | `Directory.Build.props` | `src` | — | — | nested |
| Ruby | `Gemfile` | — | — | — | nested |

### Python

Scaffold generates `pyproject.toml`. Dependencies vary per sample — most need
only `temporalio>=1.23.0,<2`, but AI and encryption samples have extras.

### Go

Scaffold generates `go.mod`. Import rewriting replaces
`github.com/temporalio/samples-go/<sample>` with `<name>/<sample>` in all
`.go` files. No per-sample dependencies — `go mod tidy` resolves them.

### TypeScript

Empty scaffold — each sample already has its own `package.json`. Copied flat.

### Java

Scaffold generates `build.gradle`. Samples live deep in
`core/src/main/java/io/temporal/samples/<sample>/`. Each per-sample manifest
supplies `sdk_version` as an extra template variable. Root `root_files` copies
the Gradle wrapper (`gradlew`, `gradlew.bat`, `gradle/`) into the output.

### .NET

Scaffold generates `Directory.Build.props` with framework target and SDK
references. `sample_path: src`.

### Ruby

Scaffold generates `Gemfile`. Ruby samples use `require_relative` for local
files, so no import rewriting is needed.

---

## Per-sample manifests

Each repo gets a `cli-sample-init` branch with a root `temporal-samples.yaml`
and a `temporal-sample.yaml` in each selected sample directory.

### Python (`samples-python`) — 12 samples

Root `temporal-samples.yaml`:

```yaml
version: 1
language: python
repo: temporalio/samples-python
scaffold:
  pyproject.toml: |
    [project]
    name = "{{name}}"
    version = "0.1.0"
    requires-python = ">=3.10"
    dependencies = [{{dependencies}}]
    [build-system]
    requires = ["hatchling"]
    build-backend = "hatchling.build"
```

Per-sample `temporal-sample.yaml` files:

| Directory | `description` | `dependencies` |
|-----------|---------------|----------------|
| `hello` | Basic hello world samples (activity, signal, query, etc.) | `"temporalio>=1.23.0,<2"` |
| `hello_standalone_activity` | Execute Activities directly from a Client, without a Workflow | `"temporalio>=1.23.0,<2"` |
| `hello_nexus` | Nexus service definition, operation handlers, and workflow calls | `"temporalio>=1.23.0,<2"`, `"nexus-rpc>=1.1.0,<2"` |
| `nexus_cancel` | Fan out Nexus operations, take first result, cancel the rest | `"temporalio>=1.23.0,<2"`, `"nexus-rpc>=1.1.0,<2"` |
| `nexus_multiple_args` | Map a Nexus operation to a handler workflow with multiple arguments | `"temporalio>=1.23.0,<2"`, `"nexus-rpc>=1.1.0,<2"` |
| `nexus_sync_operations` | Nexus service backed by a long-running workflow with updates and queries | `"temporalio>=1.23.0,<2"`, `"nexus-rpc>=1.1.0,<2"` |
| `openai_agents` | OpenAI Agents SDK with Temporal durable execution | `"openai-agents[litellm]==0.3.2"`, `"temporalio[openai-agents]>=1.18.0"`, `"requests>=2.32.0,<3"` |
| `langchain` | Orchestrate LangChain workflows with LangSmith tracing | `"langchain>=0.1.7,<0.2"`, `"langchain-openai>=0.0.6,<0.0.7"`, `"openai>=1.4.0,<2"` |
| `bedrock` | Amazon Bedrock AI chatbot with durable execution | `"temporalio>=1.23.0,<2"`, `"boto3>=1.34.92,<2"` |
| `encryption` | End-to-end encryption codec, compatible with TypeScript and Go | `"temporalio>=1.23.0,<2"`, `"cryptography>=38.0.1,<39"`, `"aiohttp>=3.8.1,<4"` |
| `dsl` | Workflow interpreter for arbitrary steps defined in YAML DSL | `"temporalio>=1.23.0,<2"`, `"pyyaml>=6.0.1,<7"`, `"dacite>=1.8.1,<2"` |
| `schedules` | Schedule a Workflow Execution and control actions | `"temporalio>=1.23.0,<2"` |

### Go (`samples-go`) — 11 samples

Root `temporal-samples.yaml`:

```yaml
version: 1
language: go
repo: temporalio/samples-go
scaffold:
  go.mod: |
    module {{name}}
    go 1.23.0
    require (
        go.temporal.io/sdk v1.41.0
        go.temporal.io/sdk/contrib/envconfig v0.1.0
    )
rewrite_imports:
  from: github.com/temporalio/samples-go
  glob: "*.go"
```

Per-sample `temporal-sample.yaml` files:

| Directory | `description` |
|-----------|---------------|
| `helloworld` | Basic hello world workflow |
| `standalone-activity` | Execute Activities directly from a Client without a Workflow |
| `nexus` | Nexus service definition and cross-namespace operation calls |
| `nexus-cancelation` | Cancel a Nexus operation using WaitRequested cancellation type |
| `nexus-context-propagation` | Propagate context through client calls, workflows, and Nexus headers |
| `nexus-multiple-arguments` | Map a Nexus operation to a workflow with multiple input arguments |
| `expense` | Asynchronous activity completion |
| `saga` | Microservice orchestration using the Saga pattern |
| `shoppingcart` | Update-with-Start and WORKFLOW_ID_CONFLICT_POLICY_USE_EXISTING |
| `dsl` | DSL workflow interpreter driven by YAML step definitions |
| `encryption` | Remote codec server for end-to-end encryption |

No `dependencies` field — Go deps resolved by `go mod tidy`.

### TypeScript (`samples-typescript`) — 11 samples

Root `temporal-samples.yaml`:

```yaml
version: 1
language: typescript
repo: temporalio/samples-typescript
scaffold: {}
```

Per-sample `temporal-sample.yaml` files:

| Directory | `description` |
|-----------|---------------|
| `hello-world` | Default scaffolded hello world project |
| `nexus-hello` | Nexus service definition and cross-namespace calls |
| `nexus-cancellation` | Cancel a Nexus operation from a caller workflow |
| `ai-sdk` | Vercel AI SDK integration for LLM applications |
| `saga` | Microservice orchestration using the Saga pattern |
| `expense` | Expense workflow with Signal-based approval/rejection |
| `child-workflows` | Child workflow execution patterns |
| `signals-queries` | Signals, Queries, and Workflow Cancellation |
| `dsl-interpreter` | DSL workflow interpreter driven by YAML step definitions |
| `encryption` | Custom data converter with AES encryption |
| `food-delivery` | Production-like distributed app from blog post |

No scaffold — each sample already has its own `package.json`. Copied flat.

### Java (`samples-java`) — 10 samples

Root `temporal-samples.yaml`:

```yaml
version: 1
language: java
repo: temporalio/samples-java
scaffold:
  build.gradle: |
    plugins { id 'java' }
    repositories { mavenCentral() }
    java { sourceCompatibility = JavaVersion.VERSION_11 }
    dependencies {
        implementation "io.temporal:temporal-sdk:{{sdk_version}}"
        implementation "io.temporal:temporal-envconfig:{{sdk_version}}"
    }
    task execute(type: JavaExec) {
        mainClass = findProperty("mainClass") ?: ""
        classpath = sourceSets.main.runtimeClasspath
    }
sample_path: core/src/main/java/io/temporal/samples
root_files:
  - gradlew
  - gradlew.bat
  - gradle/
```

Per-sample `temporal-sample.yaml` files (all under `core/src/main/java/io/temporal/samples/`):

| Directory | `description` | Extra vars |
|-----------|---------------|-----------|
| `hello` | Single-file hello world samples demonstrating core SDK features | `sdk_version: "1.32.1"` |
| `nexus` | Nexus service definition and operation handlers | `sdk_version: "1.32.1"` |
| `nexuscancellation` | Cancel a Nexus operation using WaitRequested cancellation | `sdk_version: "1.32.1"` |
| `nexuscontextpropagation` | Propagate MDC context from Workflows to Nexus operations | `sdk_version: "1.32.1"` |
| `nexusmultipleargs` | Map a Nexus operation to a workflow with multiple input arguments | `sdk_version: "1.32.1"` |
| `bookingsaga` | Trip booking Saga pattern with compensation | `sdk_version: "1.32.1"` |
| `moneytransfer` | Separate processes for workflows, activities, and transfer requests | `sdk_version: "1.32.1"` |
| `dsl` | DSL-driven workflow steps defined in JSON | `sdk_version: "1.32.1"` |
| `fileprocessing` | Route tasks to specific Workers for host-local download/process/upload | `sdk_version: "1.32.1"` |
| `encryptedpayloads` | End-to-end payload encryption | `sdk_version: "1.32.1"` |

### .NET (`samples-dotnet`) — 11 samples

Root `temporal-samples.yaml`:

```yaml
version: 1
language: dotnet
repo: temporalio/samples-dotnet
scaffold:
  Directory.Build.props: |
    <Project>
      <PropertyGroup>
        <TargetFramework>net8.0</TargetFramework>
        <ImplicitUsings>enable</ImplicitUsings>
        <Nullable>enable</Nullable>
      </PropertyGroup>
      <ItemGroup>
        <PackageReference Include="Temporalio" Version="1.11.1" />
        <PackageReference Include="Temporalio.Extensions.Hosting" Version="1.11.1" />
      </ItemGroup>
    </Project>
sample_path: src
```

Per-sample `temporal-sample.yaml` files (all under `src/`):

| Directory | `description` |
|-----------|---------------|
| `ActivitySimple` | Workflow executing synchronous static and asynchronous instance activity methods |
| `NexusSimple` | Nexus service definition and operation calls from a workflow |
| `NexusCancellation` | Cancel a Nexus operation using WaitCancellationRequested |
| `NexusContextPropagation` | Propagate context through client calls, workflows, and Nexus headers |
| `NexusMultiArg` | Convert a single Nexus argument into multiple workflow arguments |
| `Bedrock` | Amazon Bedrock AI chatbot with durable execution |
| `Saga` | Microservice orchestration using the Saga pattern |
| `Encryption` | Custom payload codec for end-to-end encryption |
| `Dsl` | Workflow interpreter for arbitrary steps defined in a DSL |
| `Polling` | Three best practices for polling |
| `AspNet` | Generic host worker + ASP.NET web application |

### Ruby (`samples-ruby`) — 10 samples

Root `temporal-samples.yaml`:

```yaml
version: 1
language: ruby
repo: temporalio/samples-ruby
scaffold:
  Gemfile: |
    source 'https://rubygems.org'
    gem 'temporalio'
```

Per-sample `temporal-sample.yaml` files:

| Directory | `description` |
|-----------|---------------|
| `activity_simple` | Simple activity and workflow calling |
| `activity_worker` | Go workflow calling Ruby activity |
| `activity_heartbeating` | Activity heartbeating and cancellation handling with progress resume |
| `message_passing_simple` | Workflow accepting signals, queries, and updates |
| `polling` | Polling best practices |
| `context_propagation` | Thread/fiber local propagation via interceptor |
| `worker_specific_task_queues` | Unique Task Queue per Worker for host-specific activities |
| `client_mtls` | Mutual TLS authentication |
| `rails_app` | API-only Rails app with shopping cart workflow |
| `sorbet_generic` | Sorbet type-checked workflow and activity definitions |

---

## Integration tests

File: `internal/temporalcli/commands.sample_integration_test.go`

Gated by build tag `sample_integration` so they don't run in default CI.

### Per-language `init` tests

Each language's test:
1. Downloads from the real GitHub sample repo (on `cli-sample-init` branch).
2. Verifies the output directory structure (scaffold files present, sample
   files in the right place, manifests excluded).
3. Runs the language's build command to prove the scaffolded project is valid.

| Test | Sample | Build command |
|------|--------|---------------|
| `TestSampleIntegration_Init_Python` | `hello` | `uv sync` |
| `TestSampleIntegration_Init_Go` | `helloworld` | `go mod tidy && go build ./...` |
| `TestSampleIntegration_Init_TypeScript` | `hello-world` | `npm install && npx tsc --noEmit` |
| `TestSampleIntegration_Init_Java` | `hello` | `gradle compileJava` |
| `TestSampleIntegration_Init_DotNet` | `ActivitySimple` | `dotnet build` |
| `TestSampleIntegration_Init_Ruby` | `activity_simple` | `bundle install` |

### Per-language `list` tests

Each language's test runs `temporal sample list <lang>` against the real repo
and asserts that known sample names appear in the output.

Run command:
```
TEMPORAL_SAMPLES_REF=cli-sample-init go test -tags sample_integration \
  -run TestSampleIntegration -v -timeout 10m ./internal/temporalcli/
```

---

## Key design decisions

1. **No interactive mode.** CLIs must be scriptable. Missing args → usage
   help + non-zero exit.

2. **Env var for test URL override.** `TEMPORAL_SAMPLES_BASE_URL` replaces
   both `raw.githubusercontent.com` and `codeload.github.com`.

3. **Streaming tarball extraction.** Don't download-then-extract. Stream the
   tar reader, writing only matching entries. Keeps memory bounded.

4. **Nested vs flat extraction.** Determined by whether `scaffold` is
   non-empty. When scaffolding, sample files go one level deeper to preserve
   absolute imports. When empty (TypeScript), sample IS the project.

5. **README promotion.** In nested mode, README.md is copied to the project
   root. This is the first thing users should read.

6. **`strings.ReplaceAll` over `text/template`.** The template language is
   trivially simple (`{{name}}`, `{{dependencies}}`). A replace loop is
   easier to understand and has no edge cases.

7. **All languages from day one.** Each language exercises a different
   combination of manifest features (scaffold, sample_path, rewrite_imports,
   root_files, flat vs nested), validating that the manifest design generalises.

8. **Arbitrary repos via URL form.** `temporal sample init <github-url>`
   works with any GitHub repo, not just official Temporal repos. The `<lang>
   <sample>` form is a convenience that maps to the official repos.

---

## Scope boundaries

**In scope**: CLI implementation, manifests for ~10 samples per language (covering
hello, nexus, standalone-activity, agentic AI, saga, encryption, expense, DSL
where they exist), integration tests, `root_files` for Gradle wrapper.

**Not in scope (later)**: Tarball caching, `--ref` flag, `temporal sample update`,
prerequisite checking (`which uv`), PHP support, config injection from
`temporal env`.

## Long-term direction

The v1 manifest duplicates dependency information that already exists in the root
`pyproject.toml` / `go.mod`. Long-term, the manifest becomes the single source of
truth: a build step in each samples repo generates per-sample build files from it,
each sample becomes self-contained in the repo, and CLI extraction becomes a trivial
copy. TypeScript already matches this end state.

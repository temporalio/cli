# `temporal sample` — Specification for Remaining Work

## What's done

The CLI implementation is complete and tested against synthetic HTTP fixtures.
`temporal sample init` and `temporal sample list` work for all 6 languages
(Python, Go, TypeScript, Java, .NET, Ruby). See [commands.sample.go](internal/temporalcli/commands.sample.go)
and [commands.sample_test.go](internal/temporalcli/commands.sample_test.go).

This spec covers what remains: CLI bug fixes, manifest files for each sample
repo, and integration tests that download from the real repos and verify the
scaffolded projects build/run.

---

## Part 1: CLI bug fixes

### 1a. Simplify manifest parsing logic

`commands.sample.go` line 314 has a confusing condition with wrong precedence.
Replace:

```go
if relToSample == "temporal-sample.yaml" || strings.HasSuffix(rel, "/temporal-sample.yaml") && !parsedSampleManifest {
    if relToSample == "temporal-sample.yaml" {
```

With:

```go
if relToSample == "temporal-sample.yaml" && !parsedSampleManifest {
```

The `HasSuffix` branch is dead code — it enters the outer `if` but the inner
`if` never matches, so it falls through to the skip block below.

### 1b. Error when sample not found

After the tarball extraction loop, if no files were extracted, return:

```
sample "nonexistent" not found in temporalio/samples-python
```

Currently the CLI silently creates a directory containing only scaffold files.

### 1c. Fallback when repo manifest is missing

If `fetchRepoManifest` gets a 404, fall back to extracting the directory as-is
(flat copy, no scaffold) and print a warning:

```
Warning: no temporal-samples.yaml found; extracting as-is.
Additional setup may be required — see README.md.
```

This is critical for the transition period while manifest PRs land.

### 1d. `list` should respect `sample_path`

Currently `list` scans the entire tarball for `temporal-sample.yaml` at any
depth. For repos with `sample_path` (Java: `core/src/main/java/io/temporal/samples`,
.NET: `src`), it should first fetch the repo manifest, then restrict scanning to
entries under `sample_path` (if set). Without this, Java's `list` will pick up
files from the wrong locations.

---

## Part 2: Manifest files for each sample repo

Each repo gets two kinds of manifest:

1. **`temporal-samples.yaml`** at the repo root — language, scaffold templates,
   optional `sample_path` and `rewrite_imports`.
2. **`temporal-sample.yaml`** in each sample directory — `description` and
   optional `dependencies` or extra template variables.

Initial target: ~10 manifests per repo covering hello, all nexus samples,
standalone activity (where it exists), all agentic AI samples, and a
representative selection of other core patterns.

All manifests go on a branch named `sample-init` in each repo so the
integration tests can target that branch via `--ref` or URL parsing. The
remaining samples can be populated incrementally after initial merge.

### Python (`samples-python`)

Root manifest:

```yaml
# temporal-samples.yaml
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

Per-sample manifests (12 samples):

| Directory | Description | Dependencies |
|-----------|-------------|-------------|
| `hello` | Basic hello world samples (activity, signal, query, etc.) | `temporalio>=1.23.0,<2` |
| `hello_standalone_activity` | Execute Activities directly from a Client, without a Workflow | `temporalio>=1.23.0,<2` |
| `hello_nexus` | Define a Nexus service, implement operation handlers, call operations from a workflow | `temporalio>=1.23.0,<2` |
| `nexus_cancel` | Fan out Nexus operations, take first result, cancel the rest | `temporalio>=1.23.0,<2` |
| `nexus_multiple_args` | Map a Nexus operation to a handler workflow with multiple input arguments | `temporalio>=1.23.0,<2` |
| `nexus_sync_operations` | Nexus service backed by a long-running workflow exposing updates and queries | `temporalio>=1.23.0,<2` |
| `openai_agents` | OpenAI Agents SDK with Temporal durable execution | `openai-agents[litellm]==0.3.2`, `temporalio[openai-agents]>=1.18.0`, `requests` |
| `langchain` | Orchestrate LangChain workflows with LangSmith tracing | `langchain>=0.1.7,<0.2`, `langchain-openai>=0.0.6,<0.0.7`, `openai>=1.4.0,<2` |
| `bedrock` | Amazon Bedrock AI chatbot with durable execution | `boto3>=1.34.92,<2` |
| `encryption` | End-to-end encryption codec, compatible with TypeScript and Go samples | `cryptography>=38.0.1,<39`, `aiohttp>=3.8.1,<4` |
| `dsl` | Workflow interpreter for arbitrary steps defined in YAML DSL | `pyyaml>=6.0.1,<7`, `dacite>=1.8.1,<2` |
| `schedules` | Schedule a Workflow Execution and control actions | `temporalio>=1.23.0,<2` |

Dependency notes: the root `pyproject.toml` declares these as optional
dependency groups (`[project.optional-dependencies]`). Every sample needs
at least `temporalio>=1.23.0,<2`; the extras are additive.

### Go (`samples-go`)

Root manifest:

```yaml
# temporal-samples.yaml
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

Per-sample manifests (11 samples):

| Directory | Description |
|-----------|-------------|
| `helloworld` | Basic hello world workflow |
| `standalone-activity/helloworld` | Execute Activities directly from a Client without a Workflow; ListActivities and CountActivities APIs |
| `nexus` | Nexus service definition and cross-namespace operation calls |
| `nexus-cancelation` | Cancel a Nexus operation using WaitRequested cancellation type |
| `nexus-context-propagation` | Propagate context through client calls, workflows, and Nexus headers |
| `nexus-multiple-arguments` | Map a Nexus operation to a workflow with multiple input arguments |
| `expense` | Asynchronous activity completion |
| `saga` | Microservice orchestration using the Saga pattern |
| `shoppingcart` | Update-with-Start and WORKFLOW_ID_CONFLICT_POLICY_USE_EXISTING |
| `dsl` | DSL workflow interpreter driven by YAML step definitions |
| `encryption` | Remote codec server for end-to-end encryption |

No per-sample `dependencies` — Go deps are resolved by `go mod tidy`.

**Import rewriting**: The CLI replaces `github.com/temporalio/samples-go/<sample>`
with `<name>/<sample>` in all `.go` files.

**Open question**: The generated `go.mod` won't have a `go.sum`. The user will
need to run `go mod tidy` before `go build`. For now, this is a documented step.

**Note on `standalone-activity/helloworld`**: This is a nested sample (two
levels deep). The CLI's current `sample` argument is a single path component.
Either the CLI needs to support `temporal sample init go standalone-activity/helloworld`
or we place the manifest at `standalone-activity/` and treat the whole directory
as the sample. Check the actual repo structure to decide.

### TypeScript (`samples-typescript`)

Root manifest:

```yaml
# temporal-samples.yaml
version: 1
language: typescript
repo: temporalio/samples-typescript
scaffold: {}
```

Per-sample manifests (11 samples):

| Directory | Description |
|-----------|-------------|
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

No scaffold needed — each sample already has its own `package.json`. Copied flat.

### Java (`samples-java`)

Root manifest:

```yaml
# temporal-samples.yaml
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
    }
    task execute(type: JavaExec) {
        mainClass = findProperty("mainClass") ?: ""
        classpath = sourceSets.main.runtimeClasspath
    }
sample_path: core/src/main/java/io/temporal/samples
```

Per-sample manifests (10 samples, all under `core/src/main/java/io/temporal/samples/`):

| Directory | Description | Extra vars |
|-----------|-------------|-----------|
| `hello` | Single-file hello world samples demonstrating core SDK features | `sdk_version: "1.32.1"` |
| `nexus` | Nexus service definition and operation handlers | `sdk_version: "1.32.1"` |
| `nexuscancellation` | Cancel a Nexus operation using WaitRequested cancellation | `sdk_version: "1.32.1"` |
| `nexuscontextpropagation` | Propagate MDC context from Workflows to Nexus operations | `sdk_version: "1.32.1"` |
| `nexusmultipleargs` | Map a Nexus operation to a workflow with multiple input arguments | `sdk_version: "1.32.1"` |
| `bookingsaga` | Trip booking Saga pattern with compensation | `sdk_version: "1.32.1"` |
| `moneytransfer` | Separate processes for workflows, activities, and transfer requests | `sdk_version: "1.32.1"` |
| `dsl` | DSL-driven workflow steps defined in JSON | `sdk_version: "1.32.1"` |
| `fileprocessing` | Route tasks to specific Workers for download, process, upload on same host | `sdk_version: "1.32.1"` |
| `encryptedpayloads` | End-to-end payload encryption | `sdk_version: "1.32.1"` |

**Extraction**: The sample at `core/src/main/java/io/temporal/samples/hello/`
is extracted to `hello/src/main/java/io/temporal/samples/hello/`. The CLI
strips the first path component (`core/`). `build.gradle` generated at
`hello/build.gradle`.

### .NET (`samples-dotnet`)

Root manifest:

```yaml
# temporal-samples.yaml
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

Per-sample manifests (11 samples, all under `src/`):

| Directory | Description |
|-----------|-------------|
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

The real `Directory.Build.props` includes analyzers and diagnostic packages.
The scaffold is minimal — just the framework target and SDK references.

### Ruby (`samples-ruby`)

Root manifest:

```yaml
# temporal-samples.yaml
version: 1
language: ruby
repo: temporalio/samples-ruby
scaffold:
  Gemfile: |
    source 'https://rubygems.org'
    gem 'temporalio'
```

Per-sample manifests (10 samples — the repo has 12 total, so nearly complete):

| Directory | Description |
|-----------|-------------|
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

## Part 3: Integration tests

These tests download from real GitHub repos (on the `sample-init` branch) and
verify the scaffolded project is structurally correct. They replace the current
synthetic-HTTP tests for the same scenarios but are additive — keep the
synthetic tests as fast unit tests.

### Test design

Each integration test:

1. Uses `TEMPORAL_SAMPLES_REF` (or similar env var, defaulting to `sample-init`)
   to target the branch with manifests. This avoids hardcoding `main` and lets
   us test before merging manifests.
2. Runs `temporal sample init` against the real GitHub endpoint (no
   `TEMPORAL_SAMPLES_BASE_URL` override).
3. Verifies the output directory structure matches expectations.
4. Runs the language's build/check command to verify the project is valid.

Guard these tests behind a build tag or env var (e.g.
`TEMPORAL_SAMPLE_INTEGRATION=1`) so they don't run in CI by default (they
require network access and language toolchains).

### Per-language integration tests

**Python** — `temporal sample init python hello`:
```
hello/
  pyproject.toml          contains name = "hello", temporalio dep
  README.md               present
  hello/
    __init__.py            present
    hello_activity.py      present
```
Build check: `cd hello && uv sync` (or `uv pip install -e .` — just verify
deps resolve). Optionally: `uv run hello/hello_activity.py` — this starts a
worker that connects to localhost:7233; if no server is running it exits with
a connection error, but that proves the code loads and imports work.

**Go** — `temporal sample init go helloworld`:
```
helloworld/
  go.mod                  module helloworld
  README.md               present
  helloworld/
    helloworld.go          no "github.com/temporalio/samples-go" imports
    worker/main.go         imports "helloworld/helloworld"
    starter/main.go        imports "helloworld/helloworld"
```
Build check: `cd helloworld && go mod tidy && go build ./...`

**TypeScript** — `temporal sample init typescript hello-world`:
```
hello-world/
  package.json            present, has @temporalio deps
  tsconfig.json           present
  src/
    activities.ts          present
    workflows.ts           present
```
Build check: `cd hello-world && npm install && npx tsc --noEmit`

**Java** — `temporal sample init java hello`:
```
hello/
  build.gradle            contains temporal-sdk:1.32.1
  README.md               present
  src/main/java/io/temporal/samples/hello/
    HelloActivity.java     present
```
Build check: `cd hello && gradle compileJava` (requires Gradle installed or
the scaffold could include the Gradle wrapper — open question).

**.NET** — `temporal sample init dotnet ActivitySimple`:
```
ActivitySimple/
  Directory.Build.props   contains net8.0, Temporalio ref
  README.md               present
  ActivitySimple/
    TemporalioSamples.ActivitySimple.csproj
    Program.cs
```
Build check: `cd ActivitySimple && dotnet build`

**Ruby** — `temporal sample init ruby activity_simple`:
```
activity_simple/
  Gemfile                 contains gem 'temporalio'
  README.md               present
  activity_simple/
    worker.rb              present
    my_workflow.rb          present
```
Build check: `cd activity_simple && bundle install`

### Test implementation sketch

```go
//go:build sample_integration

func TestSampleIntegration_Python(t *testing.T) {
    ref := envOr("TEMPORAL_SAMPLES_REF", "sample-init")
    t.Chdir(t.TempDir())

    h := NewCommandHarness(t)
    res := h.Execute("sample", "init",
        fmt.Sprintf("https://github.com/temporalio/samples-python/tree/%s/hello", ref))
    require.NoError(t, res.Err)

    // Structure checks
    assert.FileExists(t, "hello/pyproject.toml")
    assert.FileExists(t, "hello/README.md")
    assert.FileExists(t, "hello/hello/__init__.py")
    assert.FileExists(t, "hello/hello/hello_activity.py")

    // Build check
    cmd := exec.Command("uv", "sync")
    cmd.Dir = "hello"
    out, err := cmd.CombinedOutput()
    require.NoError(t, err, "uv sync failed: %s", out)
}
```

---

## Part 4: PR workflow

1. **Sample repos first**: Push `sample-init` branch to each of the 6 repos
   with root manifest + at least one per-sample manifest. This unblocks
   integration testing.

2. **CLI PR**: Fix bugs (Part 1), add integration tests (Part 3), open PR
   against `main`. The synthetic unit tests continue to provide fast CI
   coverage; integration tests run on-demand.

3. **Merge sample repos**: Once the CLI PR is validated, merge `sample-init`
   branches. Then switch integration tests to target `main`.

4. **Populate remaining manifests**: Add `temporal-sample.yaml` to remaining
   sample directories in each repo. This is incremental — `temporal sample list`
   shows only samples that have manifests, so incomplete coverage is fine.

---

## Not in scope (later)

- Tarball caching
- `--ref` flag
- `temporal sample update`
- Prerequisite checking (`which uv`)
- PHP support
- Config injection from `temporal env`
- Gradle wrapper in Java scaffold

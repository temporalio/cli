# `temporal sample` — Implementation Plan

## Ordering

1. Fix CLI bugs (spec §CLI bug fixes) — prerequisite for everything else
2. Push `sample-init` branch to each sample repo with manifests
3. Add integration tests to CLI, targeting the `sample-init` branches
4. Open CLI PR
5. Merge sample repo branches
6. Switch integration tests to target `main`
7. Populate remaining per-sample manifests incrementally

---

## Step 1: CLI bug fixes

### 1a. Manifest parsing (line 314)

Replace:

```go
if relToSample == "temporal-sample.yaml" || strings.HasSuffix(rel, "/temporal-sample.yaml") && !parsedSampleManifest {
    if relToSample == "temporal-sample.yaml" {
```

With:

```go
if relToSample == "temporal-sample.yaml" && !parsedSampleManifest {
    if err := yaml.NewDecoder(tr).Decode(&sm); err != nil {
        return fmt.Errorf("parsing sample manifest: %w", err)
    }
    parsedSampleManifest = true
    continue
}
```

### 1b. Sample-not-found error

Add a file counter in the extraction loop. After the loop:

```go
if filesExtracted == 0 {
    return fmt.Errorf("sample %q not found in %s", sample, repo)
}
```

Also clean up the empty output directory on this path.

### 1c. Manifest-missing fallback

In `init`, if `fetchRepoManifest` returns a 404 (detect via HTTP status),
set `manifest` to a zero-value `repoManifest` (empty scaffold → flat copy)
and print a warning to stderr. Don't fail.

### 1d. `list` respects `sample_path`

In `list`, fetch the repo manifest first (same 404 fallback). If
`manifest.SamplePath` is non-empty, only match tar entries under that prefix.

### Tests

Update existing synthetic tests to cover:
- `init` with a nonexistent sample name → error
- `init` when manifest fetch returns 404 → flat extraction + warning
- `list` with `sample_path` set → only samples under that path appear

---

## Step 2: Sample repo manifests

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
| `hello` | Basic hello world samples (activity, signal, query, etc.) | `temporalio>=1.23.0,<2` |
| `hello_standalone_activity` | Execute Activities directly from a Client, without a Workflow | `temporalio>=1.23.0,<2` |
| `hello_nexus` | Define a Nexus service, implement operation handlers, call operations from a workflow | `temporalio>=1.23.0,<2` |
| `nexus_cancel` | Fan out Nexus operations, take first result, cancel the rest | `temporalio>=1.23.0,<2` |
| `nexus_multiple_args` | Map a Nexus operation to a handler workflow with multiple input arguments | `temporalio>=1.23.0,<2` |
| `nexus_sync_operations` | Nexus service backed by a long-running workflow exposing updates and queries | `temporalio>=1.23.0,<2` |
| `openai_agents` | OpenAI Agents SDK with Temporal durable execution | `openai-agents[litellm]==0.3.2`, `temporalio[openai-agents]>=1.18.0`, `requests` |
| `langchain` | Orchestrate LangChain workflows with LangSmith tracing | `langchain>=0.1.7,<0.2`, `langchain-openai>=0.0.6,<0.0.7`, `openai>=1.4.0,<2` |
| `bedrock` | Amazon Bedrock AI chatbot with durable execution | `boto3>=1.34.92,<2` |
| `encryption` | End-to-end encryption codec, compatible with TypeScript and Go samples | `temporalio>=1.23.0,<2`, `cryptography>=38.0.1,<39`, `aiohttp>=3.8.1,<4` |
| `dsl` | Workflow interpreter for arbitrary steps defined in YAML DSL | `temporalio>=1.23.0,<2`, `pyyaml>=6.0.1,<7`, `dacite>=1.8.1,<2` |
| `schedules` | Schedule a Workflow Execution and control actions | `temporalio>=1.23.0,<2` |

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
| `standalone-activity/helloworld` | Execute Activities directly from a Client without a Workflow |
| `nexus` | Nexus service definition and cross-namespace operation calls |
| `nexus-cancelation` | Cancel a Nexus operation using WaitRequested cancellation type |
| `nexus-context-propagation` | Propagate context through client calls, workflows, and Nexus headers |
| `nexus-multiple-arguments` | Map a Nexus operation to a workflow with multiple input arguments |
| `expense` | Asynchronous activity completion |
| `saga` | Microservice orchestration using the Saga pattern |
| `shoppingcart` | Update-with-Start and WORKFLOW_ID_CONFLICT_POLICY_USE_EXISTING |
| `dsl` | DSL workflow interpreter driven by YAML step definitions |
| `encryption` | Remote codec server for end-to-end encryption |

**Note**: `standalone-activity/helloworld` is nested. Decide at implementation
time: either support slashes in sample names or place the manifest at
`standalone-activity/` level.

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
    }
    task execute(type: JavaExec) {
        mainClass = findProperty("mainClass") ?: ""
        classpath = sourceSets.main.runtimeClasspath
    }
sample_path: core/src/main/java/io/temporal/samples
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
| `fileprocessing` | Route tasks to specific Workers for download, process, upload on same host | `sdk_version: "1.32.1"` |
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

## Step 3: Integration tests

File: `commands.sample_integration_test.go`, guarded by `//go:build sample_integration`.

Each test uses `TEMPORAL_SAMPLES_REF` (default `sample-init`) to target the
feature branch, invokes `temporal sample init` via `CommandHarness`, then
verifies structure and runs the build command.

### Per-language tests

**Python** — init `hello`, check structure, run `uv sync`:

```go
func TestSampleIntegration_Python(t *testing.T) {
    ref := envOr("TEMPORAL_SAMPLES_REF", "sample-init")
    t.Chdir(t.TempDir())

    h := NewCommandHarness(t)
    res := h.Execute("sample", "init",
        fmt.Sprintf("https://github.com/temporalio/samples-python/tree/%s/hello", ref))
    require.NoError(t, res.Err)

    assert.FileExists(t, "hello/pyproject.toml")
    assert.FileExists(t, "hello/README.md")
    assert.FileExists(t, "hello/hello/__init__.py")
    assert.FileExists(t, "hello/hello/hello_activity.py")

    pyproject, _ := os.ReadFile("hello/pyproject.toml")
    assert.Contains(t, string(pyproject), `name = "hello"`)
    assert.Contains(t, string(pyproject), `temporalio`)

    cmd := exec.Command("uv", "sync")
    cmd.Dir = "hello"
    out, err := cmd.CombinedOutput()
    require.NoError(t, err, "uv sync failed: %s", out)
}
```

**Go** — init `helloworld`, check imports rewritten, run `go mod tidy && go build ./...`:

```go
func TestSampleIntegration_Go(t *testing.T) {
    ref := envOr("TEMPORAL_SAMPLES_REF", "sample-init")
    t.Chdir(t.TempDir())

    h := NewCommandHarness(t)
    res := h.Execute("sample", "init",
        fmt.Sprintf("https://github.com/temporalio/samples-go/tree/%s/helloworld", ref))
    require.NoError(t, res.Err)

    assert.FileExists(t, "helloworld/go.mod")
    assert.FileExists(t, "helloworld/README.md")
    assert.FileExists(t, "helloworld/helloworld/helloworld.go")
    assert.FileExists(t, "helloworld/helloworld/worker/main.go")

    worker, _ := os.ReadFile("helloworld/helloworld/worker/main.go")
    assert.Contains(t, string(worker), `"helloworld/helloworld"`)
    assert.NotContains(t, string(worker), "github.com/temporalio/samples-go")

    cmd := exec.Command("sh", "-c", "go mod tidy && go build ./...")
    cmd.Dir = "helloworld"
    out, err := cmd.CombinedOutput()
    require.NoError(t, err, "go build failed: %s", out)
}
```

**TypeScript** — init `hello-world`, check flat copy, run `npm install && npx tsc --noEmit`:

```go
func TestSampleIntegration_TypeScript(t *testing.T) {
    ref := envOr("TEMPORAL_SAMPLES_REF", "sample-init")
    t.Chdir(t.TempDir())

    h := NewCommandHarness(t)
    res := h.Execute("sample", "init",
        fmt.Sprintf("https://github.com/temporalio/samples-typescript/tree/%s/hello-world", ref))
    require.NoError(t, res.Err)

    assert.FileExists(t, "hello-world/package.json")
    assert.FileExists(t, "hello-world/tsconfig.json")
    assert.FileExists(t, "hello-world/src/workflows.ts")
    assert.NoDirExists(t, "hello-world/hello-world")

    cmd := exec.Command("sh", "-c", "npm install && npx tsc --noEmit")
    cmd.Dir = "hello-world"
    out, err := cmd.CombinedOutput()
    require.NoError(t, err, "tsc failed: %s", out)
}
```

**Java** — init `hello`, check deep path, run `gradle compileJava`:

```go
func TestSampleIntegration_Java(t *testing.T) {
    ref := envOr("TEMPORAL_SAMPLES_REF", "sample-init")
    t.Chdir(t.TempDir())

    h := NewCommandHarness(t)
    res := h.Execute("sample", "init",
        fmt.Sprintf("https://github.com/temporalio/samples-java/tree/%s/hello", ref))
    require.NoError(t, res.Err)

    assert.FileExists(t, "hello/build.gradle")
    assert.FileExists(t, "hello/README.md")
    assert.FileExists(t, "hello/src/main/java/io/temporal/samples/hello/HelloActivity.java")

    gradle, _ := os.ReadFile("hello/build.gradle")
    assert.Contains(t, string(gradle), "temporal-sdk:1.32.1")

    cmd := exec.Command("gradle", "compileJava")
    cmd.Dir = "hello"
    out, err := cmd.CombinedOutput()
    require.NoError(t, err, "gradle compileJava failed: %s", out)
}
```

**.NET** — init `ActivitySimple`, check nested structure, run `dotnet build`:

```go
func TestSampleIntegration_DotNet(t *testing.T) {
    ref := envOr("TEMPORAL_SAMPLES_REF", "sample-init")
    t.Chdir(t.TempDir())

    h := NewCommandHarness(t)
    res := h.Execute("sample", "init",
        fmt.Sprintf("https://github.com/temporalio/samples-dotnet/tree/%s/ActivitySimple", ref))
    require.NoError(t, res.Err)

    assert.FileExists(t, "ActivitySimple/Directory.Build.props")
    assert.FileExists(t, "ActivitySimple/README.md")
    assert.FileExists(t, "ActivitySimple/ActivitySimple/Program.cs")

    cmd := exec.Command("dotnet", "build")
    cmd.Dir = "ActivitySimple"
    out, err := cmd.CombinedOutput()
    require.NoError(t, err, "dotnet build failed: %s", out)
}
```

**Ruby** — init `activity_simple`, check nested structure, run `bundle install`:

```go
func TestSampleIntegration_Ruby(t *testing.T) {
    ref := envOr("TEMPORAL_SAMPLES_REF", "sample-init")
    t.Chdir(t.TempDir())

    h := NewCommandHarness(t)
    res := h.Execute("sample", "init",
        fmt.Sprintf("https://github.com/temporalio/samples-ruby/tree/%s/activity_simple", ref))
    require.NoError(t, res.Err)

    assert.FileExists(t, "activity_simple/Gemfile")
    assert.FileExists(t, "activity_simple/README.md")
    assert.FileExists(t, "activity_simple/activity_simple/worker.rb")

    cmd := exec.Command("bundle", "install")
    cmd.Dir = "activity_simple"
    out, err := cmd.CombinedOutput()
    require.NoError(t, err, "bundle install failed: %s", out)
}
```

### Helper

```go
func envOr(key, fallback string) string {
    if v := os.Getenv(key); v != "" {
        return v
    }
    return fallback
}
```

---

## Step 4: CLI PR

Open PR against `main` on the CLI repo containing:
- Bug fixes (step 1)
- Integration test file (step 3)
- Any updates to existing synthetic tests

The synthetic unit tests provide fast CI coverage. Integration tests run
manually via `go test -tags sample_integration -run TestSampleIntegration`.

---

## Step 5–7: Merge and populate

5. Merge `sample-init` branches in each sample repo.
6. Update `TEMPORAL_SAMPLES_REF` default from `sample-init` to `main` in the
   integration tests. (Or just remove the env var override if we're confident.)
7. Add `temporal-sample.yaml` to remaining sample directories in each repo.
   This is incremental — `temporal sample list` only shows samples that have
   manifests.

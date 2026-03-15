# `temporal sample` — Implementation Plan

This plan implements the spec. It covers three phases: CLI bug fixes,
per-sample manifest content, and integration tests.

---

## Phase 1: CLI bug fixes

All changes in `internal/temporalcli/commands.sample.go`. Add corresponding
unit tests in `commands.sample_test.go` for each fix.

### 1. Simplify manifest parsing (line 314)

Replace:

```go
if relToSample == "temporal-sample.yaml" || strings.HasSuffix(rel, "/temporal-sample.yaml") && !parsedSampleManifest {
    if relToSample == "temporal-sample.yaml" {
```

With:

```go
if relToSample == "temporal-sample.yaml" && !parsedSampleManifest {
```

Remove the inner `if` and its enclosing block.

### 2. Error when sample not found

After the tarball extraction loop, track whether any files were extracted.
If zero files were written and no sample manifest was parsed, return:

```
sample "nonexistent" not found in temporalio/samples-python
```

Add test: `TestSample_Init_NotFound` — request a nonexistent sample name,
assert error message.

### 3. Fallback when repo manifest is missing

In `fetchRepoManifest`, if 404, return `(nil, nil)` instead of an error.
In `run`, if manifest is nil, set `nested = false`, `scaffold = nil`, print
a warning, and proceed with flat extraction.

Add test: `TestSample_Init_NoManifest` — server returns 404 for
`temporal-samples.yaml`, assert files extracted flat with warning on stdout.

### 4. `list` should respect `sample_path`

Change `list` to fetch the repo manifest first (tolerate 404 → scan
everywhere). If `manifest.SamplePath` is set, only consider
`temporal-sample.yaml` entries whose path starts with that prefix.

Add test: `TestSample_List_SamplePath` — synthetic tarball with entries at
two depths, assert only the ones under `sample_path` are listed.

---

## Phase 2: Per-sample manifests

Each repo gets a `sample-init` branch with a root `temporal-samples.yaml`
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

**`standalone-activity`**: The manifest goes at `standalone-activity/temporal-sample.yaml`.
The sample name for the CLI is `standalone-activity`. The directory contains
`helloworld/` as a subdirectory with all the code. The existing extraction
logic copies everything under `standalone-activity/`, preserving the
`helloworld/` subdirectory. Import rewriting replaces
`github.com/temporalio/samples-go/standalone-activity/helloworld` with
`standalone-activity/helloworld`.

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

## Phase 3: Integration tests

File: `internal/temporalcli/commands.sample_integration_test.go`

Gated by build tag `sample_integration` so they don't run in default CI.

### Test harness

```go
//go:build sample_integration

package temporalcli_test

func samplesRef() string {
    if v := os.Getenv("TEMPORAL_SAMPLES_REF"); v != "" {
        return v
    }
    return "sample-init"
}
```

Each test uses `t.Chdir(t.TempDir())` and passes a full GitHub URL with
`samplesRef()` as the ref, so it hits real GitHub (no `TEMPORAL_SAMPLES_BASE_URL`).

### Per-language tests

**TestSampleIntegration_Python** — `temporal sample init` via URL for `hello`:
- Assert: `hello/pyproject.toml` contains `name = "hello"` and `temporalio` dep
- Assert: `hello/README.md` exists
- Assert: `hello/hello/__init__.py` exists
- Build: `cd hello && uv sync`

**TestSampleIntegration_Go** — `temporal sample init go helloworld`:
- Assert: `helloworld/go.mod` contains `module helloworld`
- Assert: `helloworld/helloworld/worker/main.go` contains `"helloworld/helloworld"`,
  does not contain `github.com/temporalio/samples-go`
- Build: `cd helloworld && go mod tidy && go build ./...`

**TestSampleIntegration_TypeScript** — `temporal sample init typescript hello-world`:
- Assert: `hello-world/package.json` exists with `@temporalio` deps
- Assert: `hello-world/src/workflows.ts` exists
- Assert: no `hello-world/hello-world/` nesting (flat copy)
- Build: `cd hello-world && npm install && npx tsc --noEmit`

**TestSampleIntegration_Java** — `temporal sample init java hello`:
- Assert: `hello/build.gradle` contains `temporal-sdk:1.32.1`
- Assert: `hello/src/main/java/io/temporal/samples/hello/` has `.java` files
- Build: `cd hello && gradle compileJava`

**TestSampleIntegration_DotNet** — `temporal sample init dotnet ActivitySimple`:
- Assert: `ActivitySimple/Directory.Build.props` contains `net8.0` and `Temporalio`
- Assert: `ActivitySimple/ActivitySimple/*.csproj` exists
- Build: `cd ActivitySimple && dotnet build`

**TestSampleIntegration_Ruby** — `temporal sample init ruby activity_simple`:
- Assert: `activity_simple/Gemfile` contains `gem 'temporalio'`
- Assert: `activity_simple/activity_simple/worker.rb` exists
- Build: `cd activity_simple && bundle install`

Run command:
```
TEMPORAL_SAMPLES_REF=sample-init go test -tags sample_integration -run TestSampleIntegration -v ./internal/temporalcli/
```

---

## Execution order

1. **Phase 1** (CLI bug fixes) — one commit per fix, all in this repo.
   Run existing unit tests after each.

2. **Phase 2** (manifests) — create `sample-init` branch in each of the 6
   sample repos. Push root manifest + all per-sample manifests in a single
   commit per repo.

3. **Phase 3** (integration tests) — add the test file, run against the
   `sample-init` branches. Once green, open the CLI PR.

4. **Post-merge**: merge `sample-init` branches in sample repos, switch
   integration tests to default to `main`.

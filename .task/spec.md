# `temporal sample` — Design Specification

## Problem

A new Temporal user wants to go from zero to a running sample in under two minutes.
Today that requires: finding the right samples repo, cloning it, understanding the
project structure, installing language-specific tooling, figuring out which commands
to run, and understanding the paths. `temporal sample init` should collapse this to
one command.

## Prior art

### Prior art surveyed

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
| **`cdk init`** | Built into CLI (3 templates) | Dep install + git init | Hardcoded | [docs](https://docs.aws.amazon.com/cdk/v2/guide/ref-cli-cmd-init.html) |
| **`serverless`** | Bundled with CLI | Creates `serverless.yml` + handler | Hardcoded | [docs](https://www.serverless.com/framework/docs-providers-aws-cli-reference-create) |
| **C3 (Cloudflare)** | Bundled + delegates to framework CLIs | Dep install + optional deploy | Hardcoded | [docs](https://developers.cloudflare.com/pages/get-started/c3/) |
| **`flutter create`** | Built into SDK (6 templates) | Dep install | Hardcoded | [docs](https://docs.flutter.dev/reference/create-new-app) |
| **`projen new`** | npm/PyPI packages | Ongoing synthesis from `.projenrc` | Package registry | [docs](https://projen.io/docs/introduction/getting-started/) |
| **Cookiecutter** | Git clone | Jinja2 substitution in files + filenames | `cookiecutter.json` | — |
| **Yeoman (`yo`)** | npm packages (`generator-*`) | Generator-defined | npm registry | [docs](https://yeoman.io/learning/) |
| **Vite** | Bundled templates (18) | `npm install` | Hardcoded | — |
| **Vercel Templates** | Web gallery + one-click deploy | Creates repo, configures deployment | Web gallery at vercel.com/templates | — |
| **Firebase** | Per-platform quickstart repos | IDE-openable projects | GitHub repo structure | — |
| **Supabase** | Docs-driven quickstarts | Per-framework tutorial | Curated docs pages | — |

**Key observations:**

- **Stripe CLI** is the closest structural analog: SDK vendor, multi-language, CLI-driven
  sample scaffolding. Covered in detail below.
- **`sam init`** is the closest structural model for multi-language serverless templates
  in a GitHub repo with custom URL support. It defaults to interactive mode, which is
  convenient for exploration but makes it impossible to script without `--no-interactive`.
  We should avoid this.
- **`dotnet new`** shows what a fully mature template ecosystem looks like: searchable
  registry, parameterized templates, community publishing. Worth aspiring to long-term.
- **Spring Initializr** demonstrates the "thin CLI over server API" alternative — the
  web UI is the real product, the CLI is for scripting.

### Competitors

- **Restate**: No CLI scaffolding. Manual clone from examples repo.
- **Inngest**: No CLI scaffolding. Doc-driven quickstart only.
- **Dagger**: `dagger init --sdk {lang}` creates empty module. `--blueprint` for module templates. Not example-oriented.

### Stripe CLI — closest prior art

Docs:
- CLI wiki: https://github.com/stripe/stripe-cli/wiki/Samples-command
- Samples gallery (web, no CLI docs): https://docs.stripe.com/samples
- CLI source: https://github.com/stripe/stripe-cli/tree/master/pkg/samples
- Registry repo: https://github.com/stripe-samples/samples-list

Stripe's `stripe samples create` command is the single most relevant prior art. Like
Temporal, Stripe is an SDK vendor with a CLI, multi-language samples, and the goal of
getting users from zero to a running integration quickly.

**Architecture (two-tier registry):**

1. **Central registry**: A separate repo (`stripe-samples/samples-list`) containing a
   single `samples.json` file — a flat array of `{name, description, URL}` entries
   pointing to individual sample repos. ~35 samples.

2. **Per-sample metadata**: Each sample repo has a `.cli.json` at its root:
   ```json
   {
     "name": "accept-a-payment",
     "configureDotEnv": true,
     "postInstall": {"message": "..."},
     "integrations": [
       {
         "name": "payment-element",
         "clients": ["html", "react-cra"],
         "servers": ["ruby", "node", "python", "java", "go", "dotnet"]
       }
     ]
   }
   ```

**Download**: `go-git` PlainClone to `~/.config/stripe/samples-cache/<name>/`. Pulls
on subsequent use. No tarballs.

**Multi-language handling**: Convention-based directory layout
(`{integration}/server/{lang}/`, `{integration}/client/{lang}/`). The CLI prompts the
user to select an integration, then client language, then server language, and copies
only the selected directories into the target.

**Config injection**: Parses `.env.example`, injects the user's Stripe API keys (from
`~/.config/stripe` profile), writes `server/.env`.

**UX**: Interactive prompts via `promptui.Select` (arrow-key selection). Spinners for
download/copy/configure phases. Final output: "You're all set. To get started: cd {dest}".

**Key structural difference from Temporal**: Stripe has **one sample per repo** with
multiple language implementations inside. Temporal has **one repo per language** with
multiple samples inside. This inverts the registry design: Stripe's central registry
points to repos; Temporal's would point to directories within repos.

**What we should adopt from Stripe:**
- The two-tier metadata pattern (central discovery + per-sample config)
- `go-git` for download (the CLI already uses Go; go-git is battle-tested)
- Local caching of downloaded samples
- Config injection (Temporal address/namespace instead of API keys)

**What we should do differently:**
- Our registry is per-language-repo, not a separate registry repo (simpler)
- Our metadata lives alongside the samples, not in separate repos
- We need project scaffolding (Stripe samples are already standalone per-language)

### Key insight from prior art

Two download strategies dominate: **GitHub tarballs** (create-next-app, degit) and
**git clone to cache** (Stripe, cargo-generate). Tarballs are faster for one-shot use;
git clone is better when samples are reused or updated. Since `temporal sample init` is
typically a one-shot operation, tarballs are slightly better, but `go-git` (Stripe's
approach) is simpler to implement in Go and handles private repos/auth naturally.

The most relevant models are **Stripe CLI** (closest structural analog: SDK vendor,
multi-language, CLI-driven) and **create-next-app** (monorepo of examples, subdirectory
extraction). Stripe's two-tier metadata approach is the most mature design for our
problem space.

## Current state of Temporal samples repos

Eight samples repos exist: `samples-{go,java,python,typescript,dotnet,ruby,php}` and
`samples-server`. A Rust repo is expected.

### Per-language extractability

The central design constraint is that most samples repos are monorepos where samples
share root-level build configuration. Extracting a single sample directory and making it
independently buildable requires language-specific scaffolding.

| Language | Root build files | Per-sample build file | Extractable standalone? |
|----------|------------------|-----------------------|------------------------|
| **TypeScript** | `pnpm-workspace.yaml`, root `package.json` | Own `package.json` with all deps | **Yes** — each sample is self-contained |
| **.NET** | `Directory.Build.props`, `.sln` | Own `.csproj` (minimal, inherits from `Directory.Build.props`) | **Almost** — need `Directory.Build.props` or inline its content into `.csproj` |
| **Python** | `pyproject.toml` (single, with dependency groups) | None (just `.py` files) | **No** — need to generate a `pyproject.toml` |
| **Go** | `go.mod`, `go.sum` | None (share root module) | **No** — need to generate `go.mod` + rewrite imports |
| **Ruby** | `Gemfile` | None (just `.rb` files) | **No** — need to generate `Gemfile` |
| **Java** | `build.gradle`, `settings.gradle`, `gradle.properties` | None (share root build) | **No** — need to generate full Gradle scaffolding |
| **PHP** | `docker-compose.yml`, `composer.json` | None | **No** — Docker-based setup, complex |

### README conventions

Sample READMEs vary significantly:

- **Python**: Well-structured with "Quickstart" sections, but commands assume repo-root
  working directory (e.g., `uv run hello_standalone_activity/worker.py`)
- **Go**: Terse, assumes repo-root (e.g., `go run helloworld/worker/main.go`)
- **TypeScript**: Has `.post-create` files in `.shared/` with post-scaffold instructions
- **Java**: Commands use Gradle from repo root (`./gradlew -q execute -PmainClass=...`)
- **.NET**: Commands use `dotnet run` from sample directory — works standalone

### Existing metadata

- TypeScript has `.scripts/list-of-samples.json` — a flat list of sample names
- TypeScript has `.shared/.post-create` — post-scaffold instructions template
- Python `pyproject.toml` has dependency groups for samples needing extra deps
- No other repos have structured sample metadata

## Design

### Core principle: manifest-driven with progressive enhancement

The design has two halves:

1. **Samples-side**: A manifest file (`temporal-samples.yaml`) at the root of each
   samples repo, describing samples and how to make them runnable standalone.
2. **CLI-side**: Download logic + manifest interpreter. Minimal hardcoded language
   knowledge; the manifest provides the instructions.

This follows the user's stated preference: "the CLI would acquire that logic from some
sort of structured data in the samples repo."

### Why not hardcode language logic in the CLI?

- 8 languages today, Rust coming, more possible
- Build tool conventions change (poetry→uv, npm→pnpm, Maven→Gradle)
- Samples repo maintainers know their structure best
- CLI releases are slower than samples repo updates

The CLI should be a **manifest interpreter**, not a language expert.

### Why not just clone the whole repo?

This was considered (it's what users do manually today, and README commands work as-is).
Rejected because:

- Defeats the "quick standalone project" goal
- Downloads 10-100x more than needed
- User is inside someone else's repo, not their own project
- Doesn't answer the "initialize the project for them" question

### The manifest: `temporal-samples.yaml`

Located at the root of each samples repo. Describes available samples and provides the
CLI with everything it needs to create a standalone project from any sample.

```yaml
# temporal-samples.yaml
version: 1
language: python
repo: temporalio/samples-python

# Files from the repo root that every sample needs (beyond its own directory).
# These are copied into the created project directory.
common_files: []

# Project scaffold: files to generate in the project root.
# These are templates with simple variable substitution.
scaffold:
  pyproject.toml: |
    [project]
    name = "{{sample_name}}"
    version = "0.1.0"
    requires-python = ">=3.10"
    dependencies = ["temporalio>=1.23.0,<2", {{extra_deps}}]

    [build-system]
    requires = ["hatchling"]
    build-backend = "hatchling.build"

samples:
  hello_standalone_activity:
    description: "Execute Activities directly from a Temporal Client, without a Workflow"
    # Extra dependencies beyond the base SDK
    extra_deps: []
    # Directory containing the sample files (relative to repo root)
    path: hello_standalone_activity
    # Commands to run after extraction
    setup: []
    # Commands the user should run (printed to terminal)
    run:
      - title: "Run the Worker"
        command: "uv run worker.py"
        background: true
      - title: "Execute the Activity"
        command: "uv run execute_activity.py"

  encryption:
    description: "End-to-end encryption with a custom codec"
    extra_deps: ["cryptography>=38.0.1,<39", "aiohttp>=3.8.1,<4"]
    path: encryption
    setup: []
    run:
      - title: "Run the Worker"
        command: "uv run worker.py"
        background: true
      - title: "Run the Starter"
        command: "uv run starter.py"
```

For Go, the manifest would look different:

```yaml
version: 1
language: go
repo: temporalio/samples-go

scaffold:
  go.mod: |
    module {{module_name}}

    go 1.23.0

    require go.temporal.io/sdk v1.33.0

samples:
  helloworld:
    description: "Basic hello world workflow"
    path: helloworld
    # Files from repo root needed by this sample
    common_files: []
    # Import path rewriting (like gonew)
    rewrite_imports:
      from: "github.com/temporalio/samples-go"
      to: "{{module_name}}"
    setup: ["go mod tidy"]
    run:
      - title: "Run the Worker"
        command: "go run ./worker/main.go"
        background: true
      - title: "Run the Starter"
        command: "go run ./starter/main.go"
```

For TypeScript (simplest case — samples are already standalone):

```yaml
version: 1
language: typescript
repo: temporalio/samples-typescript

scaffold: {}  # No scaffolding needed

samples:
  hello-world:
    description: "Basic hello world workflow"
    path: hello-world
    setup: ["npm install"]
    run:
      - title: "Run the Worker"
        command: "npm run start.watch"
        background: true
      - title: "Run the Workflow"
        command: "npm run workflow"
```

For Java:

```yaml
version: 1
language: java
repo: temporalio/samples-java

scaffold:
  build.gradle: |
    plugins {
        id 'application'
    }
    repositories {
        mavenCentral()
    }
    dependencies {
        implementation 'io.temporal:temporal-sdk:{{sdk_version}}'
    }
  settings.gradle: |
    rootProject.name = '{{sample_name}}'
  gradle.properties: |
    # Generated by temporal sample init

samples:
  hello:
    description: "Hello World samples demonstrating individual SDK features"
    path: core/src/main/java/io/temporal/samples/hello
    # Java samples need the Gradle wrapper too
    common_files:
      - gradlew
      - gradlew.bat
      - gradle/
    setup: ["./gradlew build"]
    run:
      - title: "Run HelloActivity"
        command: "./gradlew run --args='HelloActivity'"
```

### Template variables

The manifest supports simple `{{variable}}` substitution. Variables are:

| Variable | Source | Example |
|----------|--------|---------|
| `sample_name` | Directory name (or user-specified) | `hello_standalone_activity` |
| `module_name` | User-specified or derived | `example.com/hello` (Go), `hello_standalone_activity` (Python) |
| `extra_deps` | From sample's `extra_deps` list | `"cryptography>=38.0.1"` |
| `sdk_version` | From manifest or latest release | `1.33.0` |

This is deliberately simple — no Jinja2, no Liquid. If a sample needs complex scaffolding,
that's a signal that the samples repo should restructure itself to be more extractable.

### Command naming: `temporal sample`

**Decision: `temporal sample init` and `temporal sample list`.**

Reasoning:

- **Not `temporal init`** — `init` at root level implies initializing `temporal` itself
  (like `git init` initializes a git repo). Using it for "download a sample" is a
  semantic mismatch. It's also a valuable command name to reserve for future use
  (e.g. initializing CLI configuration, onboarding wizard).

- **Not `temporal python ...`** or `temporal <language> ...`** — language as a subcommand
  creates a sprawling namespace that conflicts with the CLI's domain-oriented structure
  (workflow, activity, schedule, etc). Makes `--help` discovery worse.

- **`temporal sample <verb>`** follows the CLI's existing pattern: `temporal workflow start`,
  `temporal schedule create`, `temporal server start-dev`. The noun is the subcommand,
  the verb is the leaf.

- **`init` over `create`** — you're *initializing a project from* a sample, not *creating*
  a sample. `create` implies authoring. (`sam init`, `cdk init`, `cargo init` use `init`
  for this reason.)

- **`temporal sample list`** follows naturally for discovery and mirrors `stripe samples list`.

- Extensible later: `temporal sample run`, `temporal sample update`, etc.

### CLI extension context

The CLI has a PATH-based extension system (shipped v1.6.0): any executable named
`temporal-<subcommand>` on PATH becomes `temporal <subcommand>`. The first consumer is
`temporal-cloud` (the cloud-cli repo at `temporalio/cloud-cli`).

**Should `sample` be built-in or an extension?** Either works — the UX is identical
thanks to the extension system. Arguments for built-in: it's part of the core
getting-started story alongside `server start-dev`, and having it ship with the binary
means zero extra installation. Arguments for extension: keeps the core CLI lean,
allows independent release cadence. **Recommendation: built-in**, because the
getting-started flow should work out of the box with no extra installation.

### CLI UX

Language is a positional argument, not a flag. Users nearly always care about one
language — it's the primary dimension of their mental model. Making it positional
keeps commands concise and natural.

```
# Standard form: language + sample name
temporal sample init python hello_standalone_activity

# From a GitHub URL (language inferred from repo name)
temporal sample init https://github.com/temporalio/samples-python/tree/main/hello_standalone_activity

# Specify output directory
temporal sample init python hello_standalone_activity --output-dir ./my-project

# List available samples for a language (language is required)
temporal sample list python
```

Missing required arguments produce usage help and exit non-zero — no interactive
prompts by default. CLIs must be scriptable.

```
$ temporal sample init
Error: requires language argument

Usage: temporal sample init <language> <sample> [flags]
       temporal sample init <github-url> [flags]

Languages: go, java, python, typescript, dotnet, ruby, php

$ temporal sample list
Error: requires language argument

Usage: temporal sample list <language> [flags]

Languages: go, java, python, typescript, dotnet, ruby, php
```

#### Detailed flow

```
$ temporal sample init python hello_standalone_activity

Downloading hello_standalone_activity from temporalio/samples-python...
Created ./hello_standalone_activity/

To get started:

  cd hello_standalone_activity

  # 1. Start the dev server (if not using Temporal Cloud)
  temporal server start-dev

  # 2. Run the Worker (in a new terminal)
  uv run worker.py

  # 3. Execute the Activity (in a new terminal)
  uv run execute_activity.py

  # View in the Web UI
  http://localhost:8233
```

#### `temporal sample list`

```
$ temporal sample list python

Available Python samples:

  hello_standalone_activity    Execute Activities directly from a Client, without a Workflow
  encryption                   End-to-end encryption with a custom codec
  dsl                          YAML-based DSL workflow interpreter
  ...

https://github.com/temporalio/samples-python
```

Language is required. This reads the manifest from the language's samples repo and
displays sample names + descriptions. The manifest is fetched (and cached briefly)
from GitHub.

#### URL parsing

When given a GitHub URL, the CLI parses it to extract:
- Owner/repo → determines language
- Branch/ref
- Path → sample directory

This means `temporal sample init https://github.com/temporalio/samples-python/tree/main/hello_standalone_activity`
is equivalent to `temporal sample init python hello_standalone_activity`.

### Download mechanism

Following create-next-app and degit:

1. Fetch `https://codeload.github.com/temporalio/samples-{lang}/tar.gz/{ref}`
2. Stream through tar extraction, filtering entries to:
   - The sample directory (from manifest `path`)
   - Any `common_files` specified in the manifest
   - The manifest file itself (to read it)
3. Strip the leading path components so files are written relative to the output directory

For step 2, we actually need the manifest first to know what to extract. So the flow is:

1. Fetch the manifest file via GitHub raw content API:
   `https://raw.githubusercontent.com/temporalio/samples-{lang}/{ref}/temporal-samples.yaml`
2. Parse it, find the sample, determine what to download
3. Download and extract via tarball API

This two-step approach avoids downloading the entire tarball when we only need a subdirectory.
(create-next-app downloads the whole tarball and filters; we can be smarter since we have a manifest.)

Actually, the tarball approach is still simpler and more robust — the manifest tells us
what to *keep* after extraction, but the tarball download + filter is a single HTTP request
regardless. The manifest file is small; fetching it first is one extra request but gives us
the sample metadata before we start.

**Fallback when no manifest exists**: If a samples repo hasn't adopted the manifest yet,
the CLI can fall back to:
1. Download the sample directory via tarball
2. Copy files as-is (no scaffolding)
3. Print the raw README.md content
4. Warn: "This sample may require additional setup. See README.md for details."

This provides a degraded but functional experience while repos adopt the manifest.

### No default interactive mode

Bare `temporal sample init` and `temporal sample list` print usage and exit non-zero.
No wizard launches unless the user explicitly opts in. This is a deliberate design
choice: CLIs should be scriptable, and implicit interactive prompts break pipelines,
CI scripts, and documentation that assumes deterministic behavior.

If we later decide interactive discovery is valuable, it should be behind an explicit
flag (e.g. `temporal sample init --interactive`) or a separate command
(`temporal sample browse`). But the core commands are deterministic.

### What about users already in a project?

The task asks: "do we assume the user is already in a functional language project or do
we initialize the project in the dir we're creating for them?"

**Answer: always create a new directory.** Reasons:

1. Merging sample code into an existing project is inherently language-specific and
   build-tool-specific (poetry vs uv vs pip, npm vs yarn vs pnpm, Maven vs Gradle).
   This is exactly the kind of thing that's better left to AI agents or manual work.
2. A standalone sample directory is independently valuable: it works, you can study it,
   then manually integrate patterns you want into your project.
3. create-next-app, gonew, cargo-generate, and degit all create new directories.

If the user wants to integrate sample code into an existing project, that's a different
(harder, more ambiguous) problem that AI agents handle well.

## What the CLI needs to hardcode

Despite the manifest-driven approach, the CLI needs some hardcoded knowledge:

1. **Language → repo mapping**: `python` → `temporalio/samples-python`, etc.
2. **GitHub URL parsing**: Extract owner, repo, ref, path from GitHub URLs.
3. **Manifest format**: The `temporal-samples.yaml` schema.
4. **Template variable substitution**: The `{{variable}}` expansion logic.
5. **Default ref**: `main` branch when not specified.

This is minimal and stable. The language-specific knowledge lives in the manifests.

## Config injection (inspired by Stripe)

Stripe's best UX innovation is injecting the user's API keys into `.env` so the sample
works immediately. We should do the equivalent for Temporal:

- If the user has a configured Temporal environment (via `temporal env`), inject the
  address and namespace into the sample's configuration.
- For local dev server: inject `localhost:7233` (the default).
- For Temporal Cloud: inject the address, namespace, and cert paths or API key.

The mechanism: each sample's `run` commands could reference environment variables that
the CLI prints as `export` statements, or the CLI could write a `.env` file. The manifest
specifies which variables the sample expects:

```yaml
env:
  TEMPORAL_ADDRESS: "{{temporal_address}}"
  TEMPORAL_NAMESPACE: "{{temporal_namespace}}"
```

The CLI resolves these from the user's current `temporal env` configuration.

## Phased rollout

### Phase 1: Core mechanism + Python/TypeScript

- Implement CLI command with GitHub tarball download
- Define manifest schema (`temporal-samples.yaml` v1)
- Add manifests to `samples-python` and `samples-typescript`
- `temporal sample init <language> <sample>` and URL form
- `temporal sample list <language>`

### Phase 2: Go, .NET, Java

- Add manifests to remaining repos
- Go requires import-path rewriting (follow gonew's approach)
- Java requires Gradle scaffolding generation
- .NET requires `Directory.Build.props` generation or inlining

### Phase 3: Ruby, PHP, Rust

- Add manifests as these repos are ready
- PHP may need special handling (Docker-based)

### Phase 4: Polish

- Caching of manifests and tarballs
- Offline mode (if tarball cached)
- Version pinning / branch selection
- `temporal sample update` to refresh a previously-initialized sample

## Scope boundary: classical vs AI

The classical `temporal sample init` command handles the **deterministic, well-defined path**:
download a known sample from a known repo into a new directory with known scaffolding.
This is fast, offline-capable (with caching), reproducible, and reliable.

**What classical programming handles well:**
- Downloading and extracting sample code
- Generating project scaffolding from manifest templates
- Printing setup/run instructions
- Listing available samples

**What's better left to AI agents:**
- Integrating sample code into an existing project with a different build tool
- Adapting sample code to the user's specific context (different package manager,
  different project structure, different SDK version)
- Troubleshooting when setup commands fail
- Explaining what the sample code does

These are complementary. `temporal sample init` gives you a working starting point.
An AI agent can then help you adapt it to your context or understand it.

## Open questions

1. **Manifest ownership**: Who maintains the manifests? Likely the samples repo maintainers,
   with CI validation that the manifest stays in sync with actual directories.

2. **SDK version pinning**: Should the manifest hardcode SDK versions, or should the CLI
   resolve "latest stable" at init time? The manifest should probably specify a minimum,
   and the CLI could optionally upgrade to latest.

3. **Branch/tag support**: Should `temporal sample init python hello_standalone_activity@v1.0`
   work? Useful for reproducibility, but adds complexity.

4. **Prerequisites checking**: Should the CLI check that `uv`, `go`, `npm`, etc. are
   installed before proceeding? Helpful for UX, but adds language-specific knowledge
   to the CLI. The manifest could specify `prerequisites: ["uv"]` and the CLI just
   checks `which`.

5. ~~**Naming**~~: Resolved — `temporal sample init` and `temporal sample list`.

6. **Monorepo sample references**: Java's `hello` samples are at
   `core/src/main/java/io/temporal/samples/hello/` — a deeply nested path. The manifest
   allows the sample `name` to be a short alias (`hello`) while `path` is the full path.
   But what about Python samples that have sub-samples (e.g., `message_passing/` contains
   multiple sub-directories)?

7. **The "Rust" question**: Should we wait for a `samples-rust` repo to exist, or define
   the manifest format now to accommodate it?

8. **Manifest bootstrapping**: How do we get manifests into all 7+ samples repos? This
   is a coordination cost. Should we start with just Python and TypeScript (the most
   extractable) and let the manifest format prove itself?

## Recommendation

Start with Phase 1 (Python + TypeScript) to validate the manifest-driven approach. These
two languages have the most extractable samples and the largest user base. The manifest
format is simple enough that adding other languages is incremental work in the samples
repos, not the CLI.

The CLI implementation is straightforward Go: HTTP client, tar extraction, YAML parsing,
template substitution, terminal output. Estimated at ~500-800 lines of Go in a single
`commands.sample.go` file plus the YAML command definition.

The manifest format should be designed now for all languages (even if only Python and
TypeScript adopt it first) to ensure it's general enough. The Go import-rewriting and
Java Gradle-scaffolding use cases should be representable in v1 of the format.

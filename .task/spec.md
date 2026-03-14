# `temporal sample` — Design Specification

## Problem

A new Temporal user wants to go from zero to a running sample in under two minutes.
Today that requires: finding the right samples repo, cloning it, understanding the
project structure, installing language-specific tooling, figuring out which commands
to run, and understanding the paths. `temporal sample init` should collapse this to
one command.

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

- **Stripe CLI** is the closest structural analog (SDK vendor, multi-language, CLI-driven
  sample scaffolding) and the only one with two-tier metadata. Covered below.
- **`sam init`** is closest for multi-language templates in a GitHub repo. Defaults to
  interactive mode — we should avoid this.
- Two download strategies dominate: **GitHub tarballs** (create-next-app, degit) and
  **git clone** (Stripe, cargo-generate). Tarballs are simpler for one-shot extraction.

**Competitors:** Restate, Inngest, and Dagger have no CLI sample scaffolding.

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

**Adopted from Stripe:** Two-tier metadata (repo-level + per-sample), caching.
**Different from Stripe:** Our metadata lives in the samples repos (not a separate
registry repo). We need project scaffolding (Stripe samples are already standalone).

## Temporal samples repos — current state

Eight repos: `samples-{go,java,python,typescript,dotnet,ruby,php}` plus `samples-server`.

| Language | Repo | Build system | Sample unit | Standalone? | Run commands |
|----------|------|-------------|-------------|-------------|--------------|
| **Python** | `samples-python` | Root `pyproject.toml` (uv/hatch) | Top-level dirs (~38) | No — need generated `pyproject.toml` | `uv run <sample>/worker.py` |
| **Go** | `samples-go` | Root `go.mod` | Top-level dirs (~48) | No — need `go.mod` + import rewrite | `go run <sample>/worker/main.go` |
| **TypeScript** | `samples-typescript` | pnpm workspace | Top-level dirs (~40), each with own `package.json` | **Yes** | `npm run start` / `npm run workflow` |
| **Java** | `samples-java` | Root Gradle multi-module | Classes in `core/src/.../samples/` | No — complex; see below | `./gradlew execute -PmainClass=...` |
| **.NET** | `samples-dotnet` | `.sln` + `Directory.Build.props` | Projects in `src/` | Almost — need `Directory.Build.props` content | `dotnet run` from sample dir |
| **Ruby** | `samples-ruby` | Root `Gemfile` | Top-level dirs (~12) | No — need generated `Gemfile` | `bundle exec ruby worker.rb` |
| **PHP** | `samples-php` | Docker + Composer | `app/` dir | No — Docker-based | Docker-based |

### Java: special case

Java samples are not top-level directories. The `core` module contains ~30 samples as
Java classes within `core/src/main/java/io/temporal/samples/` (e.g.,
`hello/HelloActivity.java`). Each sample is a self-contained class with a `main()`
method. The `springboot` and `springboot-basic` modules are separate Gradle modules
with their own `build.gradle`, closer to the directory-per-sample model.

Making a Java sample standalone requires generating a full Gradle project:
`build.gradle`, `settings.gradle`, wrapper scripts, and preserving the Java package
hierarchy. This is the hardest extraction of any language.

### PHP: special case

PHP samples use Docker Compose and RoadRunner. The repo structure is a single
application (`app/` directory), not a collection of independent samples. Supporting
PHP may require a different approach or deferral.

## Design

### Architecture

Two halves:

1. **Samples-side**: A repo-level manifest (`temporal-samples.yaml`) at each repo
   root with language identifier and scaffold templates, plus a per-sample manifest
   (`temporal-sample.yaml`) in each sample directory with description and dependencies.
2. **CLI-side**: Downloads sample code, reads both manifests, generates scaffolding,
   prints the README. The CLI is a **manifest interpreter**, not a language expert.

This keeps language-specific knowledge in the samples repos (where maintainers know
their structure best) and out of the CLI (where release cadence is slower).

### The manifests

**Repo-level: `temporal-samples.yaml`** at the repo root — language identifier,
scaffold template, and optional post-extraction operations.

**Per-sample: `temporal-sample.yaml`** in each sample directory — description and
any sample-specific metadata (e.g. dependencies).

Template variables in scaffold files: `{{name}}` (directory name) and
`{{dependencies}}` (from per-sample manifest, quoted and joined).

No `run` commands in manifests. The CLI prints the README; READMEs are the source
of truth for how to run a sample.

#### Worked examples by language

**Python** — scaffold generates `pyproject.toml`; sample nested inside project dir
to preserve absolute imports (`from hello.hello_activity import ...`):

```yaml
# temporal-samples.yaml (repo root)
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
```yaml
# hello/temporal-sample.yaml
description: Basic hello world samples (activity, signal, query, etc.)
dependencies:
  - "temporalio>=1.23.0,<2"
```
```
Result:  hello/pyproject.toml  (generated)
         hello/README.md       (copied from sample)
         hello/hello/          (sample files)
```

**Go** — scaffold generates `go.mod`; imports must be rewritten because the monorepo
module path (`github.com/temporalio/samples-go`) no longer applies. The manifest
specifies the old prefix; the CLI replaces it with the new module name in all `.go`
files. This follows `gonew`'s approach.

```yaml
# temporal-samples.yaml (repo root)
version: 1
language: go
repo: temporalio/samples-go
scaffold:
  go.mod: |
    module {{name}}
    go 1.23.0
    require go.temporal.io/sdk v1.38.0
rewrite_imports:
  from: github.com/temporalio/samples-go
  glob: "*.go"
```
```yaml
# helloworld/temporal-sample.yaml
description: Basic hello world workflow
```
```
Result:  helloworld/go.mod              (generated, module = "helloworld")
         helloworld/README.md           (copied)
         helloworld/helloworld/         (sample files)
           worker/main.go               "github.com/temporalio/samples-go/helloworld"
                                        → "helloworld/helloworld"
```

**TypeScript** — no scaffolding needed; each sample already has `package.json`.
Copied flat.

```yaml
# temporal-samples.yaml (repo root)
version: 1
language: typescript
repo: temporalio/samples-typescript
scaffold: {}
```
```yaml
# hello-world/temporal-sample.yaml
description: Basic hello world workflow
```
```
Result:  hello-world/          (sample dir copied as-is)
           package.json        (already present)
           src/activities.ts
           src/workflows.ts
           ...
```

**Java** — scaffold generates Gradle wrapper and build files. Samples live deep
in the source tree (`core/src/main/java/io/temporal/samples/hello/`); the manifest
maps a short name to the deep path. No import rewriting needed — the Java package
declaration and intra-package imports survive extraction.

```yaml
# temporal-samples.yaml (repo root)
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
```
```yaml
# core/src/main/java/io/temporal/samples/hello/temporal-sample.yaml
description: Basic hello world samples (activity, retry, signal, etc.)
sdk_version: "1.32.1"
```
```
Result:  hello/build.gradle                                       (generated)
         hello/README.md                                          (copied)
         hello/src/main/java/io/temporal/samples/hello/           (sample files)
           HelloActivity.java
           HelloSignal.java
           ...
```

**.NET** — scaffold inlines what `Directory.Build.props` normally provides into
the `.csproj`, or generates a minimal `Directory.Build.props` alongside it.

```yaml
# temporal-samples.yaml (repo root)
version: 1
language: dotnet
repo: temporalio/samples-dotnet
scaffold:
  Directory.Build.props: |
    <Project>
      <PropertyGroup>
        <TargetFramework>net8.0</TargetFramework>
      </PropertyGroup>
    </Project>
sample_path: src
```
```yaml
# src/ActivitySimple/temporal-sample.yaml
description: Simple activity execution from a workflow
```
```
Result:  ActivitySimple/Directory.Build.props  (generated)
         ActivitySimple/README.md              (copied)
         ActivitySimple/ActivitySimple/        (sample files)
           ActivitySimple.csproj
           ...
```

**Ruby** — scaffold generates `Gemfile`. Uses `require_relative` for local files,
so no import rewriting needed.

```yaml
# temporal-samples.yaml (repo root)
version: 1
language: ruby
repo: temporalio/samples-ruby
scaffold:
  Gemfile: |
    source 'https://rubygems.org'
    gem 'temporalio'
```
```yaml
# activity_simple/temporal-sample.yaml
description: Simple activity execution from a workflow
```
```
Result:  activity_simple/Gemfile     (generated)
         activity_simple/README.md   (copied)
         activity_simple/            (sample files)
           my_workflow.rb
           my_activities.rb
           worker.rb
           starter.rb
```

**Discovery:** `temporal sample list` fetches the repo tarball and scans for
directories containing `temporal-sample.yaml`. One HTTP request, no index to maintain.

### Commands

`temporal sample init` and `temporal sample list`, following the CLI's existing
`temporal <noun> <verb>` pattern. `init` over `create` because you're initializing
a project *from* a sample, not authoring one. Language is a positional argument.

```
temporal sample init <language> <sample> [--output-dir DIR]
temporal sample init <github-url> [--output-dir DIR]
temporal sample list <language>
```

Missing arguments produce usage help and exit non-zero. No interactive mode by
default — CLIs must be scriptable.

#### Example session

```
$ temporal sample init python hello_standalone_activity

Downloading hello_standalone_activity from temporalio/samples-python...
Created ./hello_standalone_activity/

  cd hello_standalone_activity
  cat README.md
```

```
$ temporal sample list python

Available Python samples:

  hello_standalone_activity    Execute Activities directly from a Client, without a Workflow
  encryption                   End-to-end encryption with a custom codec
  dsl                          YAML-based DSL workflow interpreter
  ...

https://github.com/temporalio/samples-python
```

#### URL parsing

`temporal sample init https://github.com/temporalio/samples-python/tree/main/hello_standalone_activity`
parses to language=python, sample=hello_standalone_activity, ref=main. Equivalent
to the positional form.

### Extraction

The CLI creates a project directory. For languages where samples use absolute imports
(Python, Go), the sample directory is nested inside the project to preserve the
import structure:

```
hello_standalone_activity/          ← project root (created by CLI)
  pyproject.toml                    ← generated from scaffold template
  README.md                         ← copied from sample dir
  hello_standalone_activity/        ← package dir (extracted from repo)
    __init__.py
    my_activity.py
    worker.py
```

This mirrors the monorepo layout for a single sample. README commands like
`uv run hello_standalone_activity/worker.py` work identically in both contexts.

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

### Built-in, not an extension

The CLI has a PATH-based extension system (`temporal-<subcommand>`), but `sample`
should be built-in: it's part of the core getting-started story alongside
`temporal server start-dev` and must work with zero extra installation.

## Scope

All languages except PHP are supported from the initial release:
**Python, TypeScript, Go, Java, .NET, Ruby.**

Each language exercises a different combination of the manifest features,
validating that the design generalises:

| Language | scaffold | sample\_path | rewrite\_imports | Extraction |
|----------|----------|-------------|-----------------|------------|
| Python | `pyproject.toml` | — | — | nested |
| Go | `go.mod` | — | yes | nested |
| TypeScript | empty | — | — | flat |
| Java | `build.gradle` | `core/src/main/java/io/temporal/samples` | — | nested (deep) |
| .NET | `Directory.Build.props` | `src` | — | nested |
| Ruby | `Gemfile` | — | — | nested |

PHP is deferred: the repo is a single Docker-based application, not a
collection of independent samples.

**Polish (later):** Caching, version pinning, config injection (Temporal address/
namespace from `temporal env`), `temporal sample update`.

## Open questions

1. **Java sample granularity**: The unit is a package directory (`hello/`)
   under `core/src/main/java/io/temporal/samples/`. Each package may contain
   multiple classes, all sharing a `main()` entry point. The `temporal-sample.yaml`
   lives inside the package directory.

2. **PHP**: Deferred. The repo structure is a single Docker-based application,
   not a collection of independent samples.

3. **Prerequisites checking**: Should the manifest declare `prerequisites: ["uv"]` so
   the CLI can `which`-check before proceeding?

## Long-term direction

The v1 manifest duplicates dependency information that already exists in the root
`pyproject.toml` / `go.mod`. Long-term, the manifest becomes the single source of
truth: a build step in each samples repo generates per-sample build files from it,
each sample becomes self-contained in the repo, and CLI extraction becomes a trivial
copy. TypeScript already matches this end state.

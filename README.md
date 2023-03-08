# Temporal CLI

[![Go Reference](https://pkg.go.dev/badge/github.com/temporalio/cli.svg)](https://pkg.go.dev/github.com/temporalio/cli)
[![ci](https://github.com/temporalio/cli/actions/workflows/ci.yml/badge.svg)](https://github.com/temporalio/cli/actions/workflows/ci.yml)

> ⚠️ Temporal CLI's API is still a subject for changes. ⚠️

The Temporal CLI is a command-line interface for running Temporal Server and interacting with Workflows, Activities, Namespaces, and other parts of Temporal.

## Getting Started

### Installation

#### curl

`curl -sSf https://temporal.download/cli.sh | sh`

#### Homebrew

`brew install temporal`

#### GitHub Releases

Download and extract the [latest release](https://github.com/temporalio/cli/releases/latest) from [GitHub releases](https://github.com/temporalio/cli/releases).

### Start Temporal server:

```bash
temporal server start-dev
```

At this point you should have a server running on `localhost:7233` and a web interface at <http://localhost:8233>.

Run individual commands to interact with the local Temporal server.

```bash
temporal operator namespace list
temporal workflow list
```

## Configuration

Use the help flag to see all available options:

```bash
temporal server start-dev -h
```

### Namespace Registration

Namespaces are pre-registered at startup so they're available to use right away.

By default, the "default" namespace is registered. To customize the pre-registered namespaces, start the server with:

```bash
temporal server start-dev --namespace foo --namespace bar
```

Registering namespaces the old-fashioned way via `temporal operator namespace create foo` works too!

### Persistence Modes

### In-memory

By default `temporal server start-dev` run in an in-memory mode.

#### File on Disk

To persist the state to a file use `--db-filename`:

```bash
temporal server start-dev --db-filename my_test.db
```

### Temporal UI

By default the Temporal UI is started with Temporal CLI. The UI can be disabled via a runtime flag:

```bash
temporal server start-dev --headless
```

To build without static UI assets, use the `headless` build tag when running `go build`.

### Dynamic Config

Some advanced uses require Temporal dynamic configuration values which are usually set via a dynamic configuration file inside the Temporal configuration file. Alternatively, dynamic configuration values can be set via `--dynamic-config-value KEY=JSON_VALUE`.

For example, to disable search attribute cache to make created search attributes available for use right away:

```bash
temporal server start-dev --dynamic-config-value system.forceSearchAttributesCacheRefreshOnRead=true
```

## Auto-completion

Running `temporal completion SHELL` will output the related completion SHELL code. See the following
sections for more details for each specific shell / OS and how to enable it.

### zsh auto-completion

Add the following to your `~/.zshrc` file:

```sh
source <(temporal completion zsh)
```

or from your terminal run:

```sh
echo 'source <(temporal completion zsh)' >> ~/.zshrc
```

Then run `source ~/.zshrc`.

### Bash auto-completion (linux)

Bash auto-completion relies on [bash-completion](https://github.com/scop/bash-completion#installation). Make sure
you follow the instruction [here](https://github.com/scop/bash-completion#installation) and install the software or
use a package manager to install it like `apt install bash-completion`, `pacman -S bash-completion` or `yum install bash-completion`, etc. For example
on alpine linux:

-   apk update
-   apk add bash-completion
-   source /etc/profile.d/bash_completion.sh

Verify that bash-completion is installed by running `type _init_completion` add the following to your `.bashrc`
file to enable completion for temporal

```
echo 'source <(temporal completion bash)' >>~/.bashrc
source ~/.bashrc
```

### Bash auto-completion (macos)

For macos you can install it via brew `brew install bash-completion@2` and add the following line to
your `~/.bashrc`:

```sh
[[ -r "/usr/local/etc/profile.d/bash_completion.sh" ]] && . "/usr/local/etc/profile.d/bash_completion.sh"
```

Verify that bash-completion is installed by running `type _init_completion` and add the following to your `.bashrc`
file to enable completion for temporal

```
echo 'source <(temporal completion bash)' >> ~/.bashrc
source ~/.bashrc
```

## Development

To compile the source run:

```bash
go build -o dist/temporal ./cmd/temporal
```

To compile the documentation, run:

```bash
go build -o dist/temporal-docgen ./cmd/temporal-docgen
```

To run all tests:

```bash
go test ./...
```

## Known Issues

- When consuming Temporal as a library in go mod, you may want to replace grpc-gateway with a fork to address URL escaping issue in UI. See <https://github.com/temporalio/temporalite/pull/118>

- When running the executables from the Releases page in macOS you will want to allowlist `temporal` binary in `Security & Privacy` settings:

<img width="654" alt="image (1)" src="https://user-images.githubusercontent.com/11838981/203155541-f33395f9-9ed2-4d53-a4ac-c61098cf19ef.png">

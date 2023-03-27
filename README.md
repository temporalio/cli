# Temporal CLI

[![Go Reference](https://pkg.go.dev/badge/github.com/temporalio/cli.svg)](https://pkg.go.dev/github.com/temporalio/cli)
[![ci](https://github.com/temporalio/cli/actions/workflows/ci.yml/badge.svg)](https://github.com/temporalio/cli/actions/workflows/ci.yml)

> ⚠️ Temporal CLI's API is still subject to change. ⚠️

Use the CLI to run a Temporal Server and interact with it.

**Table of Contents**

- [Install](#install)
  - [cURL](#curl)
  - [Homebrew](#homebrew)
  - [GitHub releases](#github-releases)
  - [CDN](#cdn)
- [Start the Temporal Server](#start-the-temporal-server)
- [Configure](#configure)
  - [Namespace registration](#namespace-registration)
  - [Persistence modes](#persistence-modes)
  - [Enable or disable Temporal UI](#enable-or-disable-temporal-ui)
  - [Dynamic configuration](#dynamic-configuration)
- [Auto-completion](#auto-completion)
  - [zsh auto-completion](#zsh-auto-completion)
  - [Bash auto-completion](#bash-auto-completion)
    - [macOS installation](#macos-installation)
    - [Linux installation](#linux-installation)
- [Development](#development)
  - [Compile and run CLI](#compile-and-run-cli)
  - [Compile docs](#compile-docs)
  - [Run tests](#run-tests)
- [Known Issues](#known-issues)


## Install

Temporal CLI can be installed through several different methods.

### cURL

Run the following command to install the latest version of Temporal CLI.

`curl -sSf https://temporal.download/cli.sh | sh`

### Homebrew

Run the following command to install the CLI on macOS.

`brew install temporal`

### GitHub releases

Download and extract the latest release from [GitHub releases](https://github.com/temporalio/cli/releases), and add it to your PATH.

### CDN

To install the Temporal CLI from CDN:

  1. Download:
     - <a href="https://temporal.download/cli/archive/latest?platform=linux&arch=amd64">Download for Linux amd64</a>
     - <a href="https://temporal.download/cli/archive/latest?platform=linux&arch=arm64">Download for Linux arm64</a>
     - <a href="https://temporal.download/cli/archive/latest?platform=darwin&arch=amd64">Download for macOS amd64</a>
     - <a href="https://temporal.download/cli/archive/latest?platform=darwin&arch=arm64">Download for macOS arm64</a> (Apple silicon)
     - <a href="https://temporal.download/cli/archive/latest?platform=windows&arch=amd64">Download for Windows amd64</a>
     - <a href="https://temporal.download/cli/archive/latest?platform=windows&arch=arm64">Download for Windows arm64</a>
  2. Extract the downloaded archive.
  3. Add the `temporal` binary to your PATH. (`temporal.exe` for Windows)

## Start the Temporal Server

```bash
temporal server start-dev
```

This:

- Starts the Server on `localhost:7233`
- Runs the Web UI at <http://localhost:8233>
- Creates a `default` Namespace

In another terminal, run commands to interact with the Server:

```bash
temporal workflow list
temporal operator namespace list
```

## Configure

Use the help flag to see a full list of CLI options:

```bash
temporal -h
temporal server start-dev -h
```

Configure the environment with `env` commands:

```bash
temporal env set [environment options]
```

### Namespace registration

Namespaces are pre-registered at startup so they're available to use right away.
To customize the pre-registered namespaces, start the server with:

```bash
temporal server start-dev --namespace foo --namespace bar
```

You can also register Namespaces with the following command:

```bash
temporal operator namespace create foo
```

### Persistence modes

By default, `temporal server start-dev` runs in an in-memory mode.

To persist the state to a file on disk, use `--db-filename`:

```bash
temporal server start-dev --db-filename my_test.db
```

### Enable or disable Temporal UI

By default, the Temporal UI is started with Temporal CLI. The UI can be disabled via a runtime flag:

```bash
temporal server start-dev --headless
```

To build without static UI assets, use the `headless` build tag when running `go build`.

<!--TODO: add go example -->

### Dynamic configuration

Advanced configuration of the Temporal CLI requires the use of a dynamic configuration file.
This file is created outside of the Temporal CLI; it is usually located with the service's config files.

Dynamic configuration values can also be set via `--dynamic-config-value KEY=JSON_VALUE`.
For example, to disable the search attribute cache, run:

```bash
temporal server start-dev --dynamic-config-value system.forceSearchAttributesCacheRefreshOnRead=true
```

This setting makes created search attributes immediately available for use.

## Auto-completion

The Temporal CLI has the capability to auto-complete commands.

Running `temporal completion SHELL` will output the related completion SHELL code.

### zsh auto-completion

<!-- TODO: add more information about zsh to make comparable to bash section -->

Add the following code snippet to your `~/.zshrc` file:

```sh
source <(temporal completion zsh)
```

If you're running auto-completion from the terminal, run the command below:

```sh
echo 'source <(temporal completion zsh)' >> ~/.zshrc
```

After setting the variable, run:

`source ~/.zshrc`.

### Bash auto-completion

Bash auto-completion relies on `bash-completion`.

Install the software with the steps provided [here](https://github.com/scop/bash-completion#installation), or use your preferred package manager on your operating system.

#### macOS installation

Install `bash-completion` through Homebrew:
`brew install bash-completion@2`

Follow the instruction printed in the "Caveats" section, which will say to add one of the following lines to your `~/.bashrc` file:

```sh
[[ -r "/opt/homebrew/etc/profile.d/bash_completion.sh" ]] && . "/opt/homebrew/etc/profile.d/bash_completion.sh"
```

or:

```sh
[[ -r "/usr/local/etc/profile.d/bash_completion.sh" ]] && . "/usr/local/etc/profile.d/bash_completion.sh"
```

Verify that `bash-completion` is installed by running `type _init_completion`. 
It should say `_init_completion is a function` and print the function.

Enable completion for Temporal by adding the following code to your bash file:

```bash
echo 'source <(temporal completion bash)' >> ~/.bashrc
source ~/.bashrc
```

Now test by typing `temporal`, space, and then tab twice. You should see:

```bash
$ temporal 
activity    completion  h           operator    server      workflow    
batch       env         help        schedule    task-queue
```

#### Linux installation

Use any of the following package managers to install `bash-completion`:
`apt install bash-completion`
`pacman -S bash-completion`
`yum install bash-completion`

Verify that `bash-completion` is installed by running `type _init_completion`.

To install the software on Alpine Linux, run:

```bash
apk update
apk add bash-completion
source /etc/profile.d/bash_completion.sh
```

Finally, enable completion for Temporal by adding the following code to your bash file:

```
echo 'source <(temporal completion bash)' >> ~/.bashrc
source ~/.bashrc
```

## Development

### Compile and run CLI

```bash
go build -o dist/temporal ./cmd/temporal
dist/temporal
```

### Compile docs

```bash
go build -o dist/temporal-docgen ./cmd/temporal-docgen
```

### Run tests

```bash
go test ./...
```

## Known Issues

- When consuming Temporal as a library in go mod, you may want to replace grpc-gateway with a fork to address URL escaping issue in UI. See <https://github.com/temporalio/temporalite/pull/118>

- When running the executables from the Releases page in macOS you will need to click "Allow Anyway" in `Security & Privacy` settings:

<img width="654" alt="image (1)" src="https://user-images.githubusercontent.com/11838981/203155541-f33395f9-9ed2-4d53-a4ac-c61098cf19ef.png">

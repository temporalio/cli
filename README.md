# Temporal CLI

Temporal command-line interface and development server.

**[DOCUMENTATION](https://docs.temporal.io/cli)**

## Quick Install

Reference [the documentation](https://docs.temporal.io/cli) for detailed install information.

### Install via Homebrew

    brew install temporal

### Install via download

1. Download the version for your OS and architecture:
   - [Linux amd64](https://temporal.download/cli/archive/latest?platform=linux&arch=amd64)
   - [Linux arm64](https://temporal.download/cli/archive/latest?platform=linux&arch=arm64)
   - [macOS amd64](https://temporal.download/cli/archive/latest?platform=darwin&arch=amd64)
   - [macOS arm64](https://temporal.download/cli/archive/latest?platform=darwin&arch=arm64) (Apple silicon)
   - [Windows amd64](https://temporal.download/cli/archive/latest?platform=windows&arch=amd64)
2. Extract the downloaded archive.
3. Add the `temporal` binary to your `PATH` (`temporal.exe` for Windows).

### Run via Docker

[Temporal CLI on DockerHub](https://hub.docker.com/r/temporalio/temporal)

    docker run --rm temporalio/temporal --help

Note that for dev server to be accessible from host system, it needs to listen on external IP and the ports need to be forwarded:

    docker run --rm -p 7233:7233 -p 8233:8233 temporalio/temporal:latest server start-dev --ip 0.0.0.0
    # UI is now accessible from host at http://localhost:8233/

### Build

1. Install [Go](https://go.dev/)
2. Clone repository
3. Switch to cloned directory, and run `go build ./cmd/temporal`

The executable will be at `temporal` (`temporal.exe` for Windows).

### Build with Nix

1. [Install Nix](https://docs.determinate.systems/getting-started/individuals#install)
2. Clone repository
3. Switch to cloned directory, and run `nix build`

The executable will be at `result/bin/temporal`.

### Nix Development Environment

1. [Install Nix](https://docs.determinate.systems/getting-started/individuals#install)
2. Clone repository
3. Switch to cloned directory, and run `nix develop`

Go and related tools will be made available in this shell. This can be further automated by direnv:

1. Install direnv: `nix profile add nixpkgs#direnv`
2. Run `direnv allow` in the project directory

Now every time you enter the project directory, all the tools will be available.

## Usage

Reference [the documentation](https://docs.temporal.io/cli) for detailed usage information.

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md).

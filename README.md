# Temporal CLI

Temporal command-line interface and development server.

⚠️ Under active development and inputs/outputs may change ⚠️

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

### Build

1. Install [Go](https://go.dev/)
2. Clone repository
3. Switch to cloned directory, and run `go build -tags protolegacy ./cmd/temporal`

The executable will be at `temporal` (`temporal.exe` for Windows).

## Usage

Reference [the documentation](https://docs.temporal.io/cli) for detailed usage information.

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md).

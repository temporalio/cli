# Contributing

## Building

With the latest `go` version installed, simply run the following:

    go build ./cmd/temporal

## Testing

Uses normal `go test`, e.g.:

    go test ./...

See other tests for how to leverage things like the command harness and dev server suite.

Example to run a single test case:

    go test ./... -run TestSharedServerSuite/TestOperator_SearchAttribute

## Adding/updating commands

First, update [commands.yaml](internal/temporalcli/commands.yaml) following the rules in that file. Then to regenerate the
[commands.gen.go](internal/commands.gen.go) file from code, simply run:

    go run ./internal/cmd/gen-commands

This will expect every non-parent command to have a `run` method, so for new commands developers will have to implement
`run` on the new command in a separate file before it will compile.

Once a command is updated, the CI will automatically generate new docs
and create a PR in the Documentation repo with the corresponding updates. To generate these docs locally, run:

    go run ./internal/cmd/gen-docs

This will auto-generate a new set of docs to `dist/docs/`. If a new root command is added, a new file will be automatically generated, like `temporal activity` and `activity.mdx`.

## Inject additional build-time information

To add build-time information to the version string printed by the binary, use

    go build -ldflags "-X github.com/temporalio/cli/internal.buildInfo=<MyString>"

This can be useful if, for example, you've used a `replace` statement in go.mod pointing to a local directory.
Note that inclusion of space characters in the value supplied via `-ldflags` is tricky.
Here's an example that adds branch info from a local repo to the version string, and includes a space character:

    go build -ldflags "-X 'github.com/temporalio/cli/internal.buildInfo=ServerBranch $(git -C ../temporal rev-parse --abbrev-ref HEAD)'" -o temporal ./cmd/temporal/main.go

## Building Docker image

Docker image build requires [Goreleaser](https://goreleaser.com/) to build the binaries first, although it doesn't use
Goreleaser for the Docker image itself.

First, run the Goreleaser build:

    goreleaser build --snapshot --clean

Then, run the Docker build using the following command:

    docker build --tag temporalio/temporal:snapshot --platform=<platform> .

Currently only `linux/amd64` and `linux/arm64` platforms are supported.

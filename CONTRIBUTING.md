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

First, update [commands.md](temporalcli/commandsmd/commands.md) following the rules in that file. Then to regenerate the
[commands.gen.go](temporalcli/commands.gen.go) file from code, simply run:

    go run ./temporalcli/internal/cmd/gen-commands

This will expect every non-parent command to have a `run` method, so for new commands developers will have to implement
`run` on the new command in a separate file before it will compile.

## Inject additional build-time information
To add build-time information to the version string printed by the binary, use

    go build -ldflags "-X github.com/temporalio/cli/temporalcli.buildInfo=<MyString>"

This can be useful if, for example, you've used a `replace` statement in go.mod pointing to a local directory.
Note that inclusion of space characters in the value supplied via `-ldflags` is tricky.
Here's an example that adds branch info from a local repo to the version string, and includes a space character:

    go build -ldflags "-X 'github.com/temporalio/cli/temporalcli.buildInfo=ServerBranch $(cd ../temporal && git rev-parse --abbrev-ref HEAD)'" -o temporal ./cmd/temporal/main.go

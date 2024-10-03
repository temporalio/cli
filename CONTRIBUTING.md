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

First, update [commands.yml](temporalcli/commandsgen/commands.yml) following the rules in that file. Then to regenerate the
[commands.gen.go](temporalcli/commands.gen.go) file from code, simply run:

    go run ./temporalcli/internal/cmd/gen-commands

This will expect every non-parent command to have a `run` method, so for new commands developers will have to implement
`run` on the new command in a separate file before it will compile.

Once a command is updated, new docs must also be generated:

    go run ./temporalcli/internal/cmd/gen-docs

This will auto-generate a new set of docs to `temporalcli/docs/`. If a new root command is added,
be sure to add it to git tracking, as each doc file represents a top level command, like `temporal activity`.

## Inject additional build-time information
To add build-time information to the version string printed by the binary, use

    go build -ldflags "-X github.com/temporalio/cli/temporalcli.buildInfo=<MyString>"

This can be useful if, for example, you've used a `replace` statement in go.mod pointing to a local directory.
Note that inclusion of space characters in the value supplied via `-ldflags` is tricky.
Here's an example that adds branch info from a local repo to the version string, and includes a space character:

    go build -ldflags "-X 'github.com/temporalio/cli/temporalcli.buildInfo=ServerBranch $(git -C ../temporal rev-parse --abbrev-ref HEAD)'" -o temporal ./cmd/temporal/main.go

# Contributing

## Building

With the latest `go` version installed, simply run the following:

    go build ./cmd/temporal

## Testing

Uses normal `go test`, e.g.:

    go test ./...

See other tests for how to leverage things like the command harness and dev server suite.

## Adding/updating commands

First, update [commands.md](temporalcli/commandsmd/commands.md) following the rules in that file. Then to regenerate the
[commands.gen.go](temporalcli/commands.gen.go) file from code, simply run:

    go run ./temporalcli/internal/cmd/gen-commands

This will expect every non-parent command to have a `run` method, so for new commands developers will have to implement
`run` on the new command in a separate file before it will compile.
.PHONY: all gen gen-docs build

all: gen build

gen: internal/temporalcli/commands.gen.go cliext/flags.gen.go

# Generate CLI commands code from commands.yaml, using option sets from cliext
internal/temporalcli/commands.gen.go: internal/temporalcli/commands.yaml cliext/option-sets.yaml
	go run ./cmd/gen-commands \
		-input internal/temporalcli/commands.yaml \
		-input cliext/option-sets.yaml \
		-pkg temporalcli \
		-context "*CommandContext" > $@

# Generate cliext flags code from option-sets.yaml (no commands = cliext mode)
cliext/flags.gen.go: cliext/option-sets.yaml
	go run ./cmd/gen-commands \
		-input cliext/option-sets.yaml \
		-pkg cliext > $@

gen-docs: internal/temporalcli/commands.yaml cliext/option-sets.yaml
	go run ./cmd/gen-docs \
		-input internal/temporalcli/commands.yaml \
		-input cliext/option-sets.yaml \
		-output dist/docs

build:
	go build ./cmd/temporal

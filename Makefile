.PHONY: all gen gen-docs build

all: gen build

gen: internal/temporalcli/commands.gen.go cliext/flags.gen.go internal/temporalcli/commands.system_nexus.gen.go

internal/temporalcli/commands.gen.go: internal/temporalcli/commands.yaml
	go run ./cmd/gen-commands \
		-input internal/temporalcli/commands.yaml \
		-pkg temporalcli \
		-context "*CommandContext" > $@

cliext/flags.gen.go: cliext/option-sets.yaml
	go run ./cmd/gen-commands \
		-input cliext/option-sets.yaml \
		-pkg cliext > $@

internal/temporalcli/commands.system_nexus.gen.go:
	go run ./cmd/gen-system-nexus \
		-pkg temporalcli \
		-package go.temporal.io/api/workflowservice/v1/workflowservicenexus > $@

gen-docs: internal/temporalcli/commands.yaml cliext/option-sets.yaml
	go run ./cmd/gen-docs \
		-input internal/temporalcli/commands.yaml \
		-input cliext/option-sets.yaml \
		-output dist/docs

build:
	go build ./cmd/temporal

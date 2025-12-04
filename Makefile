.PHONY: all gen gen-docs build

all: gen build

gen: internal/temporalcli/commands.gen.go

internal/temporalcli/commands.gen.go: internal/temporalcli/commands.yaml
	go run ./internal/cmd/gen-commands -input $< -pkg temporalcli -context "*CommandContext" > $@

gen-docs: internal/temporalcli/commands.yaml
	go run ./internal/cmd/gen-docs -input $< -output dist/docs

build:
	go build ./cmd/temporal

.PHONY: all gen gen-docs build

all: gen build

gen: internal/temporalcli/commands.gen.go

internal/temporalcli/commands.gen.go: internal/temporalcli/commands.yaml
	go run ./cmd/gen-commands -pkg temporalcli -context "*CommandContext" < $< > $@

gen-docs: internal/temporalcli/commands.yaml
	go run ./cmd/gen-docs -output dist/docs < $<

build:
	go build ./cmd/temporal

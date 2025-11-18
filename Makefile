.PHONY: all gen build fmt-imports

all: gen build

gen: internal/commands.gen.go

internal/commands.gen.go: internal/temporalcli/commands.yaml
	go run ./internal/cmd/gen-commands

build:
	go build ./cmd/temporal

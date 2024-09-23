.PHONY: all gen build

all: gen build

gen: temporalcli/commands.gen.go

temporalcli/commands.gen.go: temporalcli/commandsgen/commands.md
	go run ./temporalcli/internal/cmd/gen-commands

build:
	go build ./cmd/temporal

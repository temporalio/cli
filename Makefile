.PHONY: all gen build fmt-imports

all: gen build

gen: internal/commands.gen.go

internal/commands.gen.go: internal/commandsgen/commands.yml
	go run ./internal/cmd/gen-commands

build:
	go build ./cmd/temporal

fmt-imports:
	go install golang.org/x/tools/cmd/goimports@v0.39.0
	go run golang.org/x/tools/cmd/goimports -local github.com/temporalio/cli -w .

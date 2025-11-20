.PHONY: all gen build fmt-imports update-alpine

all: gen build

gen: internal/commands.gen.go

internal/commands.gen.go: internal/commandsgen/commands.yml
	go run ./internal/cmd/gen-commands

build:
	go build ./cmd/temporal

update-alpine:
	$(MAKE) -C .github/docker update-alpine

.PHONY: all gen build edit open

all: gen build

gen:
	@echo "Generating"
	go run ./temporalcli/internal/cmd/gen-commands

build:
	@echo "Building"
	go build ./cmd/temporal

edit:
	open ./temporalcli/commandsmd/commands.md

open:
	open ./temporalcli/commandsmd/commands.md

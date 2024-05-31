.PHONY: all gen build edit open

all: gen build

gen: temporalcli/commands.gen.go

temporalcli/commands.gen.go: temporalcli/commandsmd/commands.md
	go run ./temporalcli/internal/cmd/gen-commands

build:
	go build ./cmd/temporal

# For my convenience. And an alias so I don't have to remember
# which one is right. Intention: remove Makefile or remove
# these lines.

edit:
	open ./temporalcli/commandsmd/commands.md

open:
	open -a Xcode ./temporalcli/commandsmd/commands.md

path:
	@echo $(shell readlink -f './temporalcli/commandsmd/commands.md')

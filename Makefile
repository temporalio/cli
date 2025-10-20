.PHONY: all gen gen-docs build

all: gen build

gen: temporalcli/commands.gen.go temporalcloudcli/commands.gen.go

temporalcli/commands.gen.go: temporalcli/commands.yml temporalcli/commandsgen/*.go temporalcli/internal/cmd/gen-commands/main.go
	go run ./temporalcli/internal/cmd/gen-commands

temporalcloudcli/commands.gen.go: temporalcloudcli/commands.yml temporalcli/commandsgen/*.go temporalcloudcli/internal/cmd/gen-commands/main.go
	go run ./temporalcloudcli/internal/cmd/gen-commands

gen-docs: temporalcli/docs temporalcloudcli/docs

temporalcli/docs: temporalcli/commands.yml temporalcli/commandsgen/*.go temporalcli/internal/cmd/gen-docs/main.go
	go run ./temporalcli/internal/cmd/gen-docs

temporalcloudcli/docs: temporalcloudcli/commands.yml temporalcli/commandsgen/*.go temporalcloudcli/internal/cmd/gen-docs/main.go
	go run ./temporalcloudcli/internal/cmd/gen-docs

build:
	go build ./cmd/temporal
	go build ./cmd/temporal-cloud

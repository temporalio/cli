.PHONY: all gen build man

all: gen build man

gen: temporalcli/commands.gen.go

temporalcli/commands.gen.go: temporalcli/commandsgen/commands.yml
	go run ./temporalcli/internal/cmd/gen-commands

build:
	go build ./cmd/temporal

man: build
	@mkdir -p man
	@if [ -f ./temporal ]; then \
		help2man \
		    -N \
		    --name="1.1.0 (Server 1.25.0, UI 2.30.3)" \
		    --version-string="temporal CLI" \
		    --manual="temporal CLI" \
		    --section=7 ./temporal > man/temporal.7; \
	else \
		echo "Error: './temporal' executable not found."; \
		exit 1; \
	fi

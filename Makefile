.PHONY: all gen build

all: gen build

gen: temporalcli/commands.gen.go

temporalcli/commands.gen.go: temporalcli/commandsgen/commands.yml
	go run ./temporalcli/internal/cmd/gen-commands

build:
	go build ./cmd/temporal

##### Auxiliary #####
# Pinning modernc.org/sqlite to this version until https://gitlab.com/cznic/sqlite/-/issues/196 is resolved.
PINNED_DEPENDENCIES := \
	modernc.org/sqlite@v1.34.1 \
	modernc.org/libc@v1.55.3


update-dependencies:
	@printf $(COLOR) "Update dependencies..."
	@go get -u -t $(PINNED_DEPENDENCIES) ./...
	@go mod tidy

############################# Main targets #############################
# Install all tools, recompile proto files, run all possible checks and tests (long but comprehensive).
all: clean build
########################################################################


##### Variables ######

COLOR := "\e[1;36m%s\e[0m\n"

ifndef GOOS
GOOS := $(shell go env GOOS)
endif

ifndef GOARCH
GOARCH := $(shell go env GOARCH)
endif

# go.opentelemetry.io/otel/sdk/metric@v0.31.0 - there are breaking changes in v0.32.0.
# github.com/urfave/cli/v2@v2.23.6             - newer version regressed reading JSON values in subcommands. TODO apply this to subcommands https://github.com/urfave/cli/commit/dc6dfb7851fbaa6519a9691ac921c9c7e072abc8#diff-6c4b6ed7dc8834cef100f50dae61c30ffe7775a3f3f6f5a557517cb740c44a2dR237
PINNED_DEPENDENCIES := \
	go.opentelemetry.io/otel/sdk/metric@v0.31.0 \
	github.com/urfave/cli/v2@v2.23.6

##### Build #####

build:
	@printf $(COLOR) "Building Temporal CLI with OS: $(GOOS), ARCH: $(GOARCH)..."
	CGO_ENABLED=0 go build -ldflags "-s -w" ./cmd/temporal
	@printf $(COLOR) "Building docs generation tool: $(GOOS), ARCH: $(GOARCH)..."
	CGO_ENABLED=0 go build -o temporal-doc-gen ./cmd/docgen

clean:
	@printf $(COLOR) "Clearing binaries..."
	@rm -f temporal

##### Test #####
test:
	@printf $(COLOR) "Running unit tests..."
	go test ./... -count 1

##### Misc #####

update-dependencies:
	@printf $(COLOR) "Update dependencies..."
	@go get -u -t $(PINNED_DEPENDENCIES) ./...
	@go mod tidy

lint:
	@printf $(COLOR) "Run linters..."
	@golangci-lint run --verbose --timeout 10m --fix=false --new-from-rev=HEAD~ --config=.golangci.yml



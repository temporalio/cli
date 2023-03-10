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
# github.com/urfave/cli/v2@v2.4.0             - needs to accept comma in values before unlocking https://github.com/urfave/cli/pull/1241.
PINNED_DEPENDENCIES := \
	go.opentelemetry.io/otel/sdk/metric@v0.31.0 \
	github.com/urfave/cli/v2@v2.4.0

##### Build #####

build:
	@printf $(COLOR) "Building Temporal CLI with OS: $(GOOS), ARCH: $(GOARCH)..."
	CGO_ENABLED=0 go build -ldflags "-s -w -X github.com/temporalio/cli/headers.Version=0.0.0" ./cmd/temporal

clean:
	@printf $(COLOR) "Clearing binaries..."
	@rm -f temporal

##### Test #####
test:
	@printf $(COLOR) "Running unit tests..."
	go test ./... -race -count 1

##### Misc #####

update-dependencies:
	@printf $(COLOR) "Update dependencies..."
	@go get -u -t $(PINNED_DEPENDENCIES) ./...
	@go mod tidy

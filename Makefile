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

##### Build #####

build:
	@printf $(COLOR) "Build tctl with OS: $(GOOS), ARCH: $(GOARCH)..."
	CGO_ENABLED=0 go build ./cmd/tctl
	@printf $(COLOR) "Build tctl-authorization-plugin with OS: $(GOOS), ARCH: $(GOARCH)..."
	CGO_ENABLED=$(CGO_ENABLED) go build ./cmd/plugins/tctl-authorization-plugin

clean:
	@printf $(COLOR) "Clearing binaries..."
	@rm -f tctl tctl-authorization-plugin

##### Test #####
test:
	@printf $(COLOR) "Running unit tests..."
	go test ./... -race

##### Checks #####

check: copyright-check

copyright-check:
	@printf $(COLOR) "Fix license header..."
	@go run ./cmd/copyright/licensegen.go --verifyOnly

copyright:
	@printf $(COLOR) "Fix license header..."
	@go run ./cmd/copyright/licensegen.go

############################# Main targets #############################
# Install all tools, recompile proto files, run all possible checks and tests (long but comprehensive).
all: clean build
########################################################################


##### Variables ######
ifndef GOOS
GOOS := $(shell go env GOOS)
endif

ifndef GOARCH
GOARCH := $(shell go env GOARCH)
endif

COLOR := "\e[1;36m%s\e[0m\n"

##### Build #####
build:
	@printf $(COLOR) "Build tctl with OS: $(GOOS), ARCH: $(GOARCH)..."
	CGO_ENABLED=0 go build -o tctl cmd/main.go

clean:
	@printf $(COLOR) "Clearing binaries..."
	@rm -f tctl
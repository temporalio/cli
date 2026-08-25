ROOT_DIR := ..
BIN_DIR := $(CURDIR)/bin
TEMPORAL_BIN := $(BIN_DIR)/temporal
WORKER_BIN := $(BIN_DIR)/worker

.PHONY: all build build-server build-worker \
	demo-start demo-start-v1 demo-start-v2 demo-describe demo-status demo-logs demo-clean \
	server-start server-status server-logs server-stop

all: build

build: build-server build-worker

build-server:
	@mkdir -p $(BIN_DIR)
	@echo "Building the CLI with the embedded local Temporal server..."
	@cd $(ROOT_DIR) && go build -o $(TEMPORAL_BIN) ./cmd/temporal

build-worker:
	@mkdir -p $(BIN_DIR)
	@echo "Building the one-timer Workflow Worker..."
	@cd $(ROOT_DIR) && go build -o $(WORKER_BIN) ./demo/worker

demo-start: demo-start-v2

demo-start-v1: build
	@TEMPORAL_SCHEDULER_VERSION=v1 ./demo.sh demo-start

demo-start-v2: build
	@TEMPORAL_SCHEDULER_VERSION=v2 ./demo.sh demo-start

demo-describe:
	@./demo.sh demo-describe

demo-status:
	@./demo.sh demo-status

demo-logs:
	@./demo.sh demo-logs

demo-clean:
	@./demo.sh demo-clean
	@rm -f $(TEMPORAL_BIN) $(WORKER_BIN)
	@rmdir $(BIN_DIR) 2>/dev/null || true

server-start: build-server
	@./demo.sh server-start

server-status:
	@./demo.sh server-status

server-logs:
	@./demo.sh server-logs

server-stop:
	@./demo.sh server-stop

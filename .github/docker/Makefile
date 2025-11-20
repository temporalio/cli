IMAGE_REPO ?= temporalio

CLI_SHA := $(shell git rev-parse HEAD)
IMAGE_SHA_TAG := $(shell git rev-parse --short HEAD)
IMAGE_BRANCH_TAG := $(shell git rev-parse --abbrev-ref HEAD)

BAKE_ENV := CLI_SHA=$(CLI_SHA) \
	IMAGE_REPO=$(IMAGE_REPO) \
	IMAGE_SHA_TAG=$(IMAGE_SHA_TAG) \
	IMAGE_BRANCH_TAG=$(IMAGE_BRANCH_TAG)

.PHONY: help
help: ## Display this help message
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-20s %s\n", $$1, $$2}'

.PHONY: build
build: ## Build Docker images locally
	cd ../.. && goreleaser release --snapshot --clean --skip=publish
	$(BAKE_ENV) docker buildx bake --file docker-bake.hcl cli

.PHONY: load
load: ## Build and load image into local Docker
	cd ../.. && goreleaser release --snapshot --clean --skip=publish
	$(BAKE_ENV) docker buildx bake --file docker-bake.hcl --load cli

.PHONY: push
push: ## Build and push Docker images
	cd ../.. && goreleaser release --snapshot --clean --skip=publish
	$(BAKE_ENV) docker buildx bake --file docker-bake.hcl --push cli

.PHONY: clean
clean: ## Clean build artifacts
	rm -rf ../../dist

.PHONY: update-alpine
update-alpine: ## Update Alpine base image to latest version and digest (usage: make update-alpine [ALPINE_TAG=3.22])
	@if [ -n "$(ALPINE_TAG)" ]; then \
		LATEST_TAG=$(ALPINE_TAG); \
	else \
		echo "Fetching latest Alpine version from Docker Hub..."; \
		LATEST_TAG=$$(curl -s https://registry.hub.docker.com/v2/repositories/library/alpine/tags\?page_size=100 | \
			jq -r '.results[].name' | grep -E '^3\.[0-9]+$$' | sort -V | tail -1); \
	fi && \
	echo "Alpine version: $$LATEST_TAG" && \
	DIGEST=$$(docker buildx imagetools inspect alpine:$$LATEST_TAG 2>/dev/null | grep "Digest:" | head -1 | awk '{print $$2}') && \
	DIGEST_HASH=$${DIGEST#sha256:} && \
	echo "Digest: sha256:$$DIGEST_HASH" && \
	ALPINE_FULL="alpine:$$LATEST_TAG@sha256:$$DIGEST_HASH" && \
	if sed --version 2>&1 | grep -q GNU; then \
		sed -i "s|default = \"alpine:[^\"]*\"|default = \"$$ALPINE_FULL\"|" docker-bake.hcl; \
	else \
		sed -i '' "s|default = \"alpine:[^\"]*\"|default = \"$$ALPINE_FULL\"|" docker-bake.hcl; \
	fi && \
	echo "Updated docker-bake.hcl with $$ALPINE_FULL"

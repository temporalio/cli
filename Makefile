.PHONY: all gen build fmt-imports update-alpine

all: gen build

gen: internal/commands.gen.go

internal/commands.gen.go: internal/commandsgen/commands.yml
	go run ./internal/cmd/gen-commands

build:
	go build ./cmd/temporal

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
		sed -i "s|default = \"alpine:[^\"]*\"|default = \"$$ALPINE_FULL\"|" .github/docker/docker-bake.hcl; \
	else \
		sed -i '' "s|default = \"alpine:[^\"]*\"|default = \"$$ALPINE_FULL\"|" .github/docker/docker-bake.hcl; \
	fi && \
	echo "Updated .github/docker/docker-bake.hcl with $$ALPINE_FULL"

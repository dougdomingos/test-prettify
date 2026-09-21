.PHONY: build release help

build: ## Generates program binary
	go build ./cmd/test-prettify

TAG := $(shell git describe --tags --abbrev=0 2>/dev/null || echo dev)

release: ## Builds a stripped release binary tagged with the current version
	CGO_ENABLED=0 go build -ldflags="-s -w" -o bin/test-prettify-$(TAG) ./cmd/test-prettify

help: ## Show help for each make command
	@echo 'Makefile commands:'
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

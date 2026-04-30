.DEFAULT_GOAL := help

.PHONY: help generate-api mock test cov lint tidy

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## ' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

generate-api: ## Regenerate protobuf code (runs scripts/generate.sh)
	@bash scripts/generate.sh

mock: ## Regenerate minimock files
	@go generate ./internal/usecase/...

test: ## Run unit tests for the usecase layer
	@go test -count=1 ./internal/usecase/...

cov: ## Run unit tests with coverage
	@go test -cover ./internal/usecase/...

lint: ## go vet ./...
	@go vet ./...

tidy: ## go mod tidy
	@go mod tidy

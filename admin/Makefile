.DEFAULT_GOAL := help

.PHONY: help generate-api mock test cov race lint tidy clean

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## ' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

generate-api: ## Regenerate protobuf code (runs scripts/generate.sh)
	@bash scripts/generate.sh

mock: ## Regenerate minimock files (mkdir is defensive — minimock won't create the parent dir)
	@mkdir -p internal/usecase/mocks
	@go generate ./internal/usecase

test: ## Run unit tests for the usecase layer
	@go test -count=1 ./internal/usecase

cov: ## Run unit tests with coverage
	@go test -cover ./internal/usecase

race: ## Run unit tests with the race detector
	@go test -race -count=1 ./internal/usecase

lint: ## go vet (excludes generated mocks/)
	@go vet $$(go list ./... | grep -v /mocks)

tidy: ## go mod tidy
	@go mod tidy

clean: ## Remove generated artifacts (pb, mocks, bin); recoverable via generate-api + mock
	@rm -rf internal/pb internal/usecase/mocks bin

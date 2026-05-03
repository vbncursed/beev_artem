.PHONY: help generate-api lint tidy

help: ## Show available targets
	@grep -hE '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN{FS=":.*?## "}{printf "  \033[36m%-18s\033[0m %s\n", $$1, $$2}'

generate-api: ## Regenerate protobuf + grpc + grpc-gateway stubs
	bash scripts/generate.sh

lint: ## go vet (excludes mocks)
	go vet $$(go list ./... | grep -v /mocks)

tidy: ## go mod tidy
	go mod tidy

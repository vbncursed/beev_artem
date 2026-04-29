.PHONY: generate-api
generate-api:
	@./scripts/generate.sh

.PHONY: run
run:
	@APP_ENV=dev go run ./cmd/app

.PHONY: generate-api
generate-api:
	@./scripts/generate.sh

.PHONY: run
run:
	@APP_ENV=dev go run ./cmd/app

.PHONY: cov
cov:
	go test -cover ./internal/services/vacancy_service

.PHONY: mock
mock:
	mockery

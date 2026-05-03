.DEFAULT_GOAL := help

DOCKER_COMPOSE ?= docker compose
SERVICES := auth gateway vacancy resume analysis multiagent
# Subset that has an internal/usecase package — gateway is a transport-only
# edge with no business logic, so test/cov/race are scoped here.
USECASE_SERVICES := auth vacancy resume analysis multiagent

.PHONY: help up up-prod up-build up-build-prod down down-v restart restart-prod ps logs pull rebuild admin-promote admin-demote test cov race lint generate-api mock clean

help: ## Show available targets
	@grep -E '^[a-zA-Z_-]+:.*?## ' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-18s\033[0m %s\n", $$1, $$2}'

up: ## Start dev stack (gateway + services + local postgres/redis)
	APP_ENV=dev COMPOSE_PROFILES=dev $(DOCKER_COMPOSE) up -d

up-prod: ## Start prod stack (no local infra; uses external hosts from config.docker.prod.yaml)
	APP_ENV=prod $(DOCKER_COMPOSE) up -d

up-build: ## up + --build
	APP_ENV=dev COMPOSE_PROFILES=dev $(DOCKER_COMPOSE) up -d --build

up-build-prod: ## up-prod + --build
	APP_ENV=prod $(DOCKER_COMPOSE) up -d --build

down: ## Stop and remove containers
	# COMPOSE_PROFILES=dev so the dev-profiled postgres/redis containers
	# are part of the teardown — without it they stay running, hold the
	# `hr-net` network alive, and the next `make up` reuses stale state.
	COMPOSE_PROFILES=dev $(DOCKER_COMPOSE) down

down-v: ## Full reset: down + remove volumes/orphans
	COMPOSE_PROFILES=dev $(DOCKER_COMPOSE) down -v --remove-orphans

restart: down up ## Restart dev stack

restart-prod: down up-prod ## Restart prod stack

ps: ## List containers
	$(DOCKER_COMPOSE) ps

logs: ## Tail logs (last 200 lines, follow)
	$(DOCKER_COMPOSE) logs -f --tail=200

pull: ## Pull updated images from registry
	$(DOCKER_COMPOSE) pull

# rebuild replaces 12 hand-written per-service targets with one parameterised
# call. The dev branch flips on the "dev" profile so postgres+redis come up
# alongside; prod relies on external infra hosts pinned in
# config.docker.prod.yaml.
rebuild: ## Rebuild and restart one service: make rebuild SVC=auth [ENV=prod]
	@if [ -z "$(SVC)" ]; then echo "Usage: make rebuild SVC=<service> [ENV=prod]"; exit 1; fi
	@if [ "$(ENV)" = "prod" ]; then \
		APP_ENV=prod $(DOCKER_COMPOSE) up -d --build $(SVC); \
	else \
		APP_ENV=dev COMPOSE_PROFILES=dev $(DOCKER_COMPOSE) up -d --build $(SVC); \
	fi

# admin-promote / admin-demote: shells into the running auth container and
# invokes the in-image `admin` CLI. Use to bootstrap the first admin (the
# UpdateUserRole RPC requires an existing admin to call it).
admin-promote: ## Promote a user to admin: make admin-promote EMAIL=user@example.com
	@if [ -z "$(EMAIL)" ]; then echo "Usage: make admin-promote EMAIL=<email>"; exit 1; fi
	@docker exec hr-auth admin promote --email=$(EMAIL)

admin-demote: ## Demote an admin back to user: make admin-demote EMAIL=user@example.com
	@if [ -z "$(EMAIL)" ]; then echo "Usage: make admin-demote EMAIL=<email>"; exit 1; fi
	@docker exec hr-auth admin demote --email=$(EMAIL)

test: ## go test -count=1 against each service's internal/usecase
	@for s in $(USECASE_SERVICES); do echo "=== $$s ==="; (cd $$s && go test -count=1 ./internal/usecase) || exit 1; done

cov: ## go test -cover against each service's internal/usecase
	@for s in $(USECASE_SERVICES); do echo "=== $$s ==="; (cd $$s && go test -cover ./internal/usecase) || exit 1; done

race: ## go test -race against each service's internal/usecase
	@for s in $(USECASE_SERVICES); do echo "=== $$s ==="; (cd $$s && go test -race -count=1 ./internal/usecase) || exit 1; done

lint: ## go vet across all services (excludes mocks/)
	@for s in $(SERVICES); do echo "=== $$s ==="; (cd $$s && go vet $$(go list ./... | grep -v /mocks)) || exit 1; done

generate-api: ## Regenerate protobuf code for all services (calls each service's scripts/generate.sh)
	@for s in $(SERVICES); do echo "=== $$s ==="; bash $$s/scripts/generate.sh; done

mock: ## Regenerate minimock files in every service that has a usecase layer
	@for s in $(USECASE_SERVICES); do echo "=== $$s ==="; (cd $$s && $(MAKE) --no-print-directory mock) || exit 1; done

clean: ## Remove generated artifacts in every service (pb, mocks, local bin); recoverable via generate-api + mock
	@for s in $(SERVICES); do echo "=== $$s ==="; (cd $$s && $(MAKE) --no-print-directory clean) || exit 1; done

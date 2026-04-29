DOCKER_COMPOSE ?= docker compose
APP_ENV ?= dev

.PHONY: help
help: ## Показать список доступных команд
	@echo "Usage: make <target>"
	@echo ""
	@grep -E '^[a-zA-Z0-9_-]+:.*?## ' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-18s %s\n", $$1, $$2}'

.PHONY: up
up: ## Поднять dev-окружение (gateway + auth + vacancy + resume + analysis + multiagent + локальные postgres/redis)
	APP_ENV=dev COMPOSE_PROFILES=dev $(DOCKER_COMPOSE) up -d

.PHONY: up-dev
up-dev: ## То же, что up: поднять dev-окружение
	APP_ENV=dev COMPOSE_PROFILES=dev $(DOCKER_COMPOSE) up -d

.PHONY: up-prod
up-prod: ## Поднять prod-профиль (gateway + auth + vacancy + resume + analysis + multiagent, внешние postgres/redis)
	APP_ENV=prod $(DOCKER_COMPOSE) up -d

.PHONY: rebuild-dev
rebuild-dev: ## Пересобрать и перезапустить весь dev-стек
	APP_ENV=dev COMPOSE_PROFILES=dev $(DOCKER_COMPOSE) up -d --build

.PHONY: rebuild-prod
rebuild-prod: ## Пересобрать и перезапустить весь prod-стек
	APP_ENV=prod $(DOCKER_COMPOSE) up -d --build

.PHONY: auth-rebuild-dev
auth-rebuild-dev: ## Пересобрать и перезапустить только auth в dev
	APP_ENV=dev COMPOSE_PROFILES=dev $(DOCKER_COMPOSE) up -d --build auth

.PHONY: auth-rebuild-prod
auth-rebuild-prod: ## Пересобрать и перезапустить только auth в prod
	APP_ENV=prod $(DOCKER_COMPOSE) up -d --build auth

.PHONY: gateway-rebuild-dev
gateway-rebuild-dev: ## Пересобрать и перезапустить только gateway в dev
	APP_ENV=dev COMPOSE_PROFILES=dev $(DOCKER_COMPOSE) up -d --build gateway

.PHONY: gateway-rebuild-prod
gateway-rebuild-prod: ## Пересобрать и перезапустить только gateway в prod
	APP_ENV=prod $(DOCKER_COMPOSE) up -d --build gateway

.PHONY: vacancy-rebuild-dev
vacancy-rebuild-dev: ## Пересобрать и перезапустить только vacancy в dev
	APP_ENV=dev COMPOSE_PROFILES=dev $(DOCKER_COMPOSE) up -d --build vacancy

.PHONY: vacancy-rebuild-prod
vacancy-rebuild-prod: ## Пересобрать и перезапустить только vacancy в prod
	APP_ENV=prod $(DOCKER_COMPOSE) up -d --build vacancy

.PHONY: resume-rebuild-dev
resume-rebuild-dev: ## Пересобрать и перезапустить только resume в dev
	APP_ENV=dev COMPOSE_PROFILES=dev $(DOCKER_COMPOSE) up -d --build resume

.PHONY: resume-rebuild-prod
resume-rebuild-prod: ## Пересобрать и перезапустить только resume в prod
	APP_ENV=prod $(DOCKER_COMPOSE) up -d --build resume

.PHONY: analysis-rebuild-dev
analysis-rebuild-dev: ## Пересобрать и перезапустить только analysis в dev
	APP_ENV=dev COMPOSE_PROFILES=dev $(DOCKER_COMPOSE) up -d --build analysis

.PHONY: analysis-rebuild-prod
analysis-rebuild-prod: ## Пересобрать и перезапустить только analysis в prod
	APP_ENV=prod $(DOCKER_COMPOSE) up -d --build analysis

.PHONY: multiagent-rebuild-dev
multiagent-rebuild-dev: ## Пересобрать и перезапустить только multiagent в dev
	APP_ENV=dev COMPOSE_PROFILES=dev $(DOCKER_COMPOSE) up -d --build multiagent

.PHONY: multiagent-rebuild-prod
multiagent-rebuild-prod: ## Пересобрать и перезапустить только multiagent в prod
	APP_ENV=prod $(DOCKER_COMPOSE) up -d --build multiagent

.PHONY: down
down: ## Остановить и удалить контейнеры текущего compose проекта
	$(DOCKER_COMPOSE) down

.PHONY: restart
restart: ## Перезапустить dev-окружение (down + up)
restart: down up

.PHONY: restart-prod
restart-prod: ## Перезапустить prod-профиль (down + up-prod)
restart-prod: down up-prod

.PHONY: ps
ps: ## Показать список контейнеров compose проекта
	$(DOCKER_COMPOSE) ps

.PHONY: logs
logs: ## Смотреть логи всех сервисов (follow, последние 200 строк)
	$(DOCKER_COMPOSE) logs -f --tail=200

.PHONY: pull
pull: ## Обновить Docker-образы из реестра
	$(DOCKER_COMPOSE) pull

.PHONY: down-v
down-v: ## Полный сброс compose-проекта (down -v --remove-orphans)
	$(DOCKER_COMPOSE) down -v --remove-orphans

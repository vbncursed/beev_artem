# Multiagent Service

gRPC микросервис генерации HR-решения и обратной связи кандидату.

## Метод

- `GenerateDecision`

## Что делает

- принимает контекст вакансии и кандидата;
- формирует рекомендацию `hire|maybe|no`;
- возвращает confidence, rationale, feedback и agent trace;
- сохраняет нормализованный request/response в PostgreSQL (`multiagent_decisions`).

## Конфиги

Порядок выбора:

1. `configPath` (если задан)
2. `APP_ENV=prod` -> `config.docker.prod.yaml`
3. иначе -> `config.docker.dev.yaml`

## Запуск

```bash
make multiagent-rebuild-dev
```

## Порт

- gRPC: `50055`

# Analysis Service

gRPC микросервис анализа резюме и скоринга кандидатов.

## Методы

- `StartAnalysis`
- `GetAnalysis`
- `ListCandidatesByVacancy`

## HTTP маршруты через gateway

- `POST /api/v1/resumes/{resume_id}/analyze`
- `GET /api/v1/analyses/{analysis_id}`
- `GET /api/v1/vacancies/{vacancy_id}/candidates`

## Поведение StartAnalysis

- `use_llm=false`: сохраняется эвристический скоринг и базовый `ai` блок.
- `use_llm=true`: после базового скоринга сервис запрашивает `multiagent-service` (`GenerateDecision`) и обновляет `analysis.ai`.

## Конфиги

Порядок выбора:

1. `configPath` (если задан)
2. `APP_ENV=prod` -> `config.docker.prod.yaml`
3. иначе -> `config.docker.dev.yaml`

Ключевые параметры:

- `server.grpc_addr` (по умолчанию `:50054`)
- `multiagent.grpc_addr` (по умолчанию `multiagent:50055`)

## Запуск

```bash
make analysis-rebuild-dev
```

## Порт

- gRPC: `50054`

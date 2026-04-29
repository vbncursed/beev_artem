# Gateway Service

HTTP gateway для внутренних gRPC-сервисов.
Сейчас проксирует запросы в `auth-service` и `vacancy-service`.

## Что делает

- принимает HTTP/JSON запросы от клиентов;
- конвертирует HTTP -> gRPC через grpc-gateway;
- отправляет запросы в `auth-service` по адресу из конфига.
- отдаёт OpenAPI (`/swagger.json`) и UI-документацию через Scalar (`/docs`).

## Конфиги

Порядок выбора:

1. `configPath` (если задан)
2. `APP_ENV=prod` -> `config.docker.prod.yaml`
3. иначе -> `config.docker.dev.yaml`

Файлы:

- `config.docker.dev.yaml`
  - `server.http_addr: :8080`
  - `auth.grpc_addr: auth:50050`
  - `vacancy.grpc_addr: vacancy:50051`
  - `resume.grpc_addr: resume:50052`
  - `analysis.grpc_addr: analysis:50054`
- `config.docker.prod.yaml`
  - `server.http_addr: :8080`
  - `auth.grpc_addr: auth:50050`
  - `vacancy.grpc_addr: vacancy:50051`
  - `resume.grpc_addr: resume:50052`
  - `analysis.grpc_addr: analysis:50054`

## HTTP маршруты (через auth)

- `POST /api/v1/auth/register`
- `POST /api/v1/auth/login`
- `POST /api/v1/auth/refresh`
- `POST /api/v1/auth/logout`
- `POST /api/v1/auth/logout-all`
- `GET /api/v1/auth/me`
- `POST /api/v1/vacancies`
- `POST /api/v1/vacancies/{vacancy_id}/candidates`
- `POST /api/v1/vacancies/{vacancy_id}/candidates/from-resume`
- `GET /api/v1/vacancies/{vacancy_id}`
- `GET /api/v1/vacancies`
- `PATCH /api/v1/vacancies/{vacancy_id}`
- `POST /api/v1/vacancies/{vacancy_id}/archive`
- `GET /api/v1/candidates/{candidate_id}`
- `GET /api/v1/resumes/{resume_id}`
- `POST /api/v1/resumes/intake`
- `POST /api/v1/resumes/intake/batch`
- `POST /api/v1/resumes/{resume_id}/analyze`
- `GET /api/v1/analyses/{analysis_id}`
- `GET /api/v1/vacancies/{vacancy_id}/candidates`

## Запуск

Из корня репозитория:

```bash
make up-dev
```

Проверка:

```bash
curl -sS -X POST http://localhost:8080/api/v1/auth/register \
  -H 'Content-Type: application/json' \
  -d '{"email":"test@example.com","password":"Password123@"}'
```

Документация:

- [http://localhost:8080/swagger.json](http://localhost:8080/swagger.json)
- [http://localhost:8080/docs](http://localhost:8080/docs)

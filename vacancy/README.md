# Vacancy Service

gRPC микросервис управления вакансиями.

## Методы

- `CreateVacancy`
- `GetVacancy`
- `ListVacancies`
- `UpdateVacancy`
- `ArchiveVacancy`

## HTTP маршруты через gateway

- `POST /api/v1/vacancies`
- `GET /api/v1/vacancies/{vacancy_id}`
- `GET /api/v1/vacancies`
- `PATCH /api/v1/vacancies/{vacancy_id}`
- `POST /api/v1/vacancies/{vacancy_id}/archive`

Все маршруты требуют `Authorization: Bearer <access_token>`.

## Доступ (RBAC)

- `admin`:
  - видит все вакансии;
  - может обновлять и архивировать любые вакансии.
- обычный пользователь:
  - видит только свои вакансии;
  - может обновлять и архивировать только свои вакансии.

## Связь с Resume Service

- `vacancy_id` используется как контекст при загрузке резюме в конкретную вакансию:
  - `POST /api/v1/vacancies/{vacancy_id}/candidates/from-resume`.
- Также поддерживается загрузка резюме в общий пул без вакансии:
  - `POST /api/v1/resumes/intake`
  - `POST /api/v1/resumes/intake/batch`

## Конфиги

Порядок выбора:

1. `configPath` (если задан)
2. `APP_ENV=prod` -> `config.docker.prod.yaml`
3. иначе -> `config.docker.dev.yaml`

## Запуск

```bash
make vacancy-rebuild-dev
```

## Порт

- gRPC: `50051`

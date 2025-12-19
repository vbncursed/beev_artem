## MVP: HR-сервис анализа резюме

Сервис умеет:
- хранить вакансии (с навыками и весами)
- хранить резюме (файл + распарсенный текст)
- извлекать структурный профиль резюме (summary/skills/experience/education) через LLM
- сравнивать резюме с вакансией, считать score и формировать отчёт
- выдавать “кандидатов по вакансии” с фильтрацией/сортировкой

## Запуск

### 1) Конфигурация (.env)

Пример `.env`:

```bash
PORT=8080
DATABASE_URL=postgres://postgres:postgres@localhost:5432/hr?sslmode=disable

JWT_SECRET=dev-secret-change
JWT_ISSUER=hr-service
JWT_TTL_MINUTES=60

# LLM (OpenRouter)
OPENROUTER_API_KEY=sk-or-...
OPENROUTER_MODEL=qwen/qwen2.5-32b-instruct
# OPENROUTER_BASE_URL=https://openrouter.ai/api/v1
# OPENROUTER_HTTP_REFERER=
# OPENROUTER_APP_TITLE=hr-service
```

### 2) Запуск сервера

```bash
go run ./cmd/server
```

По умолчанию сервер слушает `:8080`.

## Swagger

Swagger UI доступен по пути: `/swagger/index.html`.

Генерация спецификаций:

```bash
go install github.com/swaggo/swag/cmd/swag@latest
swag init -g cmd/server/main.go -o ./docs
```

Авторизация в Swagger:
- Нажмите **Authorize** и вставьте JWT в заголовок `Authorization`
- Поддерживаются форматы: `Bearer <JWT>` или просто `<JWT>`

## Основной flow (MVP)

1) Регистрация/логин → получаем JWT  
2) Создаём вакансию  
3) Загружаем резюме (получаем `resumeId` + профиль строится автоматически)  
4) Проверяем профиль резюме  
5) Создаём анализ `resumeId + vacancyId` → получаем score + report  
6) Смотрим кандидатов по вакансии (`/vacancies/{id}/candidates`)

## Эндпоинты (основные)

### Health
- `GET /api/v1/health`
- `GET /api/v1/ready`

### Auth
- `POST /api/v1/auth/register`
- `POST /api/v1/auth/login`

### Vacancies
- `POST /api/v1/vacancies`
- `GET /api/v1/vacancies?limit=&offset=`
- `GET /api/v1/vacancies/{id}`
- `PUT /api/v1/vacancies/{id}/skills`
- `DELETE /api/v1/vacancies/{id}`

### Resumes
- `POST /api/v1/resumes`
- `GET /api/v1/resumes?limit=&offset=`
- `GET /api/v1/resumes/{id}`
- `GET /api/v1/resumes/{id}/file`
- `DELETE /api/v1/resumes/{id}`
- `GET /api/v1/resumes/{id}/profile`
- `POST /api/v1/resumes/{id}/profile/rebuild`

### Analyses
- `POST /api/v1/analyses`
- `GET /api/v1/analyses?limit=&offset=`
- `GET /api/v1/analyses/{id}`
- `DELETE /api/v1/analyses/{id}`
- `GET /api/v1/vacancies/{id}/analyses?limit=&offset=`
- `GET /api/v1/vacancies/{id}/candidates?minScore=&skill=&sort=&limit=&offset=`

## Примеры curl

### 1) Register/Login

```bash
curl -sS -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"test123"}'
```

```bash
TOKEN=$(curl -sS -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"test123"}' | jq -r .token)
```

### 2) Создать вакансию

```bash
VACANCY_ID=$(curl -sS -X POST http://localhost:8080/api/v1/vacancies \
  -H "Authorization: $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "title":"Python Backend Developer",
    "description":"Backend разработка, API, БД",
    "skills":[
      {"skill":"Python","weight":1.0},
      {"skill":"FastAPI","weight":0.9},
      {"skill":"PostgreSQL","weight":0.8}
    ]
  }' | jq -r .id)
```

### 3) Загрузить резюме

```bash
RESUME_ID=$(curl -sS -X POST http://localhost:8080/api/v1/resumes \
  -H "Authorization: $TOKEN" \
  -F "file=@/path/to/resume.pdf" | jq -r .id)
```

### 4) Посмотреть профиль резюме

```bash
curl -sS -H "Authorization: $TOKEN" \
  http://localhost:8080/api/v1/resumes/$RESUME_ID/profile | jq .
```

### 5) Создать анализ (resumeId + vacancyId)

```bash
curl -sS -X POST http://localhost:8080/api/v1/analyses \
  -H "Authorization: $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{\"resumeId\":\"$RESUME_ID\",\"vacancyId\":\"$VACANCY_ID\"}" | jq .
```

### 6) Кандидаты по вакансии

```bash
curl -sS -H "Authorization: $TOKEN" \
  "http://localhost:8080/api/v1/vacancies/$VACANCY_ID/candidates?minScore=0.6&sort=score_desc&limit=50&offset=0" | jq .
```

## Схема/таблицы

В dev-режиме репозитории создают/дополняют таблицы автоматически при старте (через `CREATE TABLE IF NOT EXISTS` и `ALTER TABLE ... IF NOT EXISTS`).


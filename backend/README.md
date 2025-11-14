## Запуск

```bash
go run ./cmd/server
```

Сервер слушает `:8080` по умолчанию (можно задать `PORT`).

## Эндпоинты

- `GET /api/v1/health` – liveness
- `GET /api/v1/ready` – readiness
- `POST /api/v1/auth/register` – `{ "email": string, "password": string }`
- `POST /api/v1/auth/login` – `{ "email": string, "password": string }`

## Swagger
## Конфигурация (.env)

Далее отредактируйте `.env` (переменные `PORT`, `DATABASE_URL`, `JWT_SECRET`, `JWT_ISSUER`, `JWT_TTL_MINUTES`). Приложение автоматически загрузит значения из `.env` при старте.

## Миграции/Схема

В dev-режиме репозиторий автоматически создаёт таблицу `users` при старте. Для продакшена используйте миграции. Базовая схема:

```sql
CREATE TABLE IF NOT EXISTS users (
  id UUID PRIMARY KEY,
  email TEXT NOT NULL UNIQUE,
  password_hash TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL,
  is_admin BOOLEAN NOT NULL DEFAULT FALSE
);
```


Проект уже содержит swagger-комментарии в обработчиках. Для генерации спеков:

1. Установите CLI:

```bash
go install github.com/swaggo/swag/cmd/swag@latest
```

2. Сгенерируйте документацию из корня проекта:

```bash
swag init -g cmd/server/main.go -o ./docs
```


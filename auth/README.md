# Auth Service

Микросервис аутентификации и авторизации.
Работает только по gRPC.

## Что делает сервис

- регистрация пользователя
- логин и выдача `access`/`refresh` токенов
- refresh токенов (ротация refresh)
- logout текущей сессии
- logout всех сессий пользователя
- изменение роли пользователя (admin-only)

## Технологии

- Go 1.26+
- gRPC + Protocol Buffers
- PostgreSQL (пользователи)
- Redis (сессии и rate limit)
- JWT + bcrypt

## Архитектура

```text
[gRPC :50050]
      |
[API layer: validation, mapping, errors, rate limit]
      |
[Service layer: auth business logic]
      |
[Storage: PostgreSQL + Redis]
```

## gRPC API

Proto-файл: `api/auth_api/auth.proto`

Сервис: `auth.service.v1.AuthService`

Методы:

- `Register`
- `Login`
- `Refresh`
- `Logout`
- `LogoutAll`
- `Me`
- `ValidateAccessToken` (внутренний RPC для gateway/сервисов)
- `UpdateUserRole`

Пример проверки доступности:

```bash
grpcurl -plaintext localhost:50050 list
```

## Конфиги

Сервис выбирает конфиг так:

1. `configPath` (если задан)
2. `APP_ENV=prod` -> `config.docker.prod.yaml`
3. иначе -> `config.docker.dev.yaml`

Файлы:

- `config.docker.dev.yaml` — для dev в Docker
  - `database.host=postgres`
  - `redis.host=redis`
- `config.docker.prod.yaml` — для prod в Docker
  - внешний PostgreSQL
  - внешний Redis

Важно для prod: перед запуском заменить плейсхолдеры (`CHANGE_ME`, хосты) на реальные значения.

## Запуск через общий Docker Compose (рекомендуется)

Из корня репозитория:

```bash
make up-dev
```

Что поднимется:

- `auth`
- локальные `postgres` и `redis`

Prod-профиль:

```bash
make up-prod
```

Что поднимется:

- только `auth`
- PostgreSQL и Redis должны быть внешними и доступны по настройкам в `config.docker.prod.yaml`

Полезные команды:

- `make help` — список всех команд
- `make logs` — логи
- `make ps` — статус контейнеров
- `make down` — остановка
- `make down-v` — остановка + удаление volume

## Порты

- gRPC: `50050`

## Хранилища

### PostgreSQL

Таблица пользователей создаётся автоматически при старте.

```sql
CREATE TABLE IF NOT EXISTS auth_users (
    id BIGSERIAL PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    role VARCHAR(50) NOT NULL DEFAULT 'user',
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_auth_users_email ON auth_users(email);
```

### Redis

Ключи:

- сессии: `session:<hex(refresh_hash)>`
- rate limit: `ratelimit:<endpoint>:<ip>`

## Безопасность

- пароль хранится только в виде bcrypt-хеша
- access token — JWT (короткий TTL), claims: `sub`, `user_id`, `email`, `role`, `exp`
- refresh token — opaque, хранится в Redis
- лимиты запросов на login/register/refresh задаются в конфиге

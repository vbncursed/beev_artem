# auth — техническое задание

## Назначение

Сервис идентификации и сессий пользователей beev. Источник истины по всему,
что касается «кто этот вызов»: владеет таблицами `users` и `sessions`,
выпускает и валидирует JWT, поддерживает refresh-rotation и принудительный
выход. Является клиентом для каждого другого backend-сервиса
(vacancy / resume / analysis / multiagent / gateway) — все они вызывают
`ValidateAccessToken` для проверки Bearer-токенов.

## Архитектура

Clean architecture, четыре слоя:

```
auth/internal/
├── domain/                       чистые типы (User, Session, JWT claims)
├── usecase/                      бизнес-логика (Register / Login / Refresh /
│                                 Logout / ValidateAccessToken / Me)
├── infrastructure/               адаптеры портов
│   ├── persistence/              pgx + goose миграции, реализует UserStorage
│   │                             и SessionStorage
│   └── tokens/                   реализация TokenIssuer на JWT (HS256)
└── transport/
    ├── grpc/                     gRPC-обработчики, errdetails.ErrorInfo
    └── middleware/                Recovery + Logging interceptors. Auth-
                                  interceptor сюда не подключён: auth — сам
                                  источник истины и проверяет токены внутри
                                  use-case при необходимости (`Me`).
```

## API

| RPC | HTTP | Описание |
|---|---|---|
| `Register` | `POST /api/v1/auth/register` | Создаёт пользователя + первую сессию. Возвращает access + refresh + userId. Email+пароль в теле. |
| `Login` | `POST /api/v1/auth/login` | Проверяет пароль (bcrypt), выдаёт новую пару токенов. Rate-limited. |
| `Refresh` | `POST /api/v1/auth/refresh` | Меняет refreshToken на новую пару. Старый refresh инвалидируется (rotation). |
| `Logout` | `POST /api/v1/auth/logout` | Удаляет конкретную сессию по refreshToken. |
| `LogoutAll` | `POST /api/v1/auth/logout-all` | Удаляет все сессии пользователя — нужен пароль. |
| `Me` | `GET /api/v1/auth/me` | Возвращает identity по access-токену. Используется фронтом для bootstrap-валидации. |
| `ValidateAccessToken` | (gRPC-only) | Внутренний RPC для остальных сервисов: парсит JWT, проверяет ревокацию, возвращает `(valid, userId, role)`. Не торчит наружу через grpc-gateway. |

## Domain model

```go
type User struct {
    ID           uint64
    Email        string
    PasswordHash string  // bcrypt
    Role         string  // "user" | "admin"
    CreatedAt    time.Time
}

type Session struct {
    ID           string
    UserID       uint64
    RefreshToken string  // хешированный, не plain
    UserAgent    string
    IP           string
    CreatedAt    time.Time
    ExpiresAt    time.Time
}
```

## JWT claims

```json
{
  "sub": "1234",
  "role": "user",
  "iat": 1746201200,
  "exp": 1746204800,
  "jti": "<sessionID>"
}
```

`sub` хранит `userID` строкой; `jti` равен `Session.ID` — позволяет отозвать
конкретную сессию через `LogoutAll`.

## Зависимости

- **PostgreSQL** — таблицы `users`, `sessions`. Миграции goose
  (`internal/infrastructure/persistence/migrations`).
- **Redis** — rate-limit для `/login`, `/register`, `/refresh`. Token-bucket
  через `golang.org/x/time/rate` поверх Redis-counter'а.
- **Внешних gRPC зависимостей нет** — auth самодостаточен.

## Конфигурация

Файл: `config.docker.dev.yaml` / `config.docker.prod.yaml`. Селектор
выбирается по env `APP_ENV` (`prod`/`production` → prod, иначе dev) или
явному `configPath`.

```yaml
database: { host, port, username, password, name, ssl_mode }
redis:    { host, port, password, db }
server:
  grpc_addr: ":50050"
  tls: { cert_file, key_file }   # opt-in
jwt:
  issuer: "beev-auth"
  access_ttl: 15m
  refresh_ttl: 720h               # 30 days
rate_limit:
  login:    { rps, burst, window }
  register: { rps, burst, window }
  refresh:  { rps, burst, window }
```

### Секреты через env

| Env | Required | Описание |
|---|---|---|
| `AUTH_JWT_SECRET` | ✓ (≥32 байт, не плейсхолдер) | подпись JWT |
| `AUTH_DB_PASSWORD` | dev: optional, prod: required | пароль PostgreSQL |
| `AUTH_REDIS_PASSWORD` | пусто в dev | пароль Redis |

`docker-compose.yaml` блокирует старт контейнера если `AUTH_JWT_SECRET`
не задан или равен плейсхолдеру.

## Запуск и сборка

Авторизованный поток развёрнут только через docker-compose:

```bash
make rebuild SVC=auth        # пересобрать и перезапустить
make logs                    # хвост логов всех сервисов
docker exec hr-auth psql ... # подключиться к БД
```

Локальный запуск без compose не предусмотрен (нет `make run`) — сервис
зависит от postgres+redis, поднимаемых compose-профилем `dev`.

## Тестирование

Unit-тесты только для `internal/usecase` (`package usecase`, white-box):

- assertions: `gotest.tools/v3/assert`
- suites: `github.com/stretchr/testify/suite`
- mocks: `github.com/gojuno/minimock` (генерируется через `make mock`)
- mocked интерфейсы: `AuthStorage`, `SessionStorage`, `TokenIssuer`, `RateLimiter`
- coverage цель: ≥90% (текущая 93.8%)

Никаких integration / handler / storage / e2e тестов — это политика
монорепо, не пробел в покрытии.

## Безопасность

- Пароли — `bcrypt` cost 12 (настраивается)
- Refresh-токены хранятся хешированными (`sha256`) — утечка БД не
  компрометирует активные сессии
- Rate-limits применяются ДО проверки пароля чтобы не нагружать bcrypt
  при брутфорсе
- TLS опционален в YAML (для k8s обычно делегируется service mesh)

## Известные ограничения

- Нет восстановления пароля через email (TODO)
- Нет двух-факторной аутентификации (TODO)
- Email верификация не требуется при регистрации (упрощение MVP)
- Роли захардкожены (`user`, `admin`) — RBAC с произвольными ролями вне
  scope текущего MVP

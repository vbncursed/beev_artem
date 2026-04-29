# Auth Service

Микросервис аутентификации и авторизации.
Работает только по gRPC.

## Что делает сервис

- регистрация пользователя
- логин и выдача `access`/`refresh` токенов
- refresh токенов (одноразовая ротация — повторное использование refresh-токена отклоняется)
- logout текущей сессии
- logout всех сессий пользователя
- изменение роли пользователя (admin-only)
- валидация access-токенов (внутренний RPC для gateway)

## Технологии

- Go 1.26+
- gRPC + Protocol Buffers
- PostgreSQL (пользователи) — миграции через `pressly/goose` v3, embedded через `embed.FS`
- Redis (сессии, индекс `user_sessions:<uid>`, rate-limit)
- JWT (HS256, требуется `exp`) + bcrypt (cost конфигурируемый)
- gotest.tools/v3/assert + testify/suite + gojuno/minimock — для unit-тестов бизнес-логики

## Архитектура

```text
[gRPC :50050]
      |
[API layer: interceptors (recovery + logging + auth) → handlers]
      |
[Service layer: auth business logic, storage interfaces]
      |
[Storage: PostgreSQL (auth_users) + Redis (sessions, user_sessions index, rate-limit)]
```

Аутентификация для не-public RPC (`Logout`, `LogoutAll`, `Me`, `UpdateUserRole`) выполняется одним interceptor-ом — handler получает `*tokenClaims` через `ClaimsFromContext(ctx)` и не парсит JWT сам.

## gRPC API

Proto-файл: `api/auth_api/auth.proto`. Сервис: `auth.service.v1.AuthService`.

| Метод | Описание | Auth |
|---|---|---|
| `Register` | создать пользователя, выдать пару токенов | публичный |
| `Login` | проверить пароль, выдать пару токенов | публичный |
| `Refresh` | ротация refresh-токена (одноразовая, атомарная через Redis `GETDEL`) | публичный (refresh в теле) |
| `Logout` | отозвать одну сессию (по refresh-токену) | требует `Authorization: Bearer <access>` |
| `LogoutAll` | отозвать все сессии пользователя | требует Bearer |
| `Me` | текущий пользователь по access-токену | требует Bearer |
| `UpdateUserRole` | изменить роль (admin-only) | требует Bearer (роль `admin`) |
| `ValidateAccessToken` | проверить access-токен, вернуть user_id/email/role | публичный (используется gateway) |

Также зарегистрирован стандартный `grpc.health.v1.Health` (для k8s/compose readiness probes).

Пример проверки:

```bash
grpcurl -plaintext localhost:50050 list
grpcurl -plaintext localhost:50050 grpc.health.v1.Health/Check
```

## Конфигурация

Файл выбирается так:

1. env `configPath` если задан;
2. иначе `APP_ENV=prod|production` → `config.docker.prod.yaml`;
3. иначе → `config.docker.dev.yaml`.

### Секреты — только через env (не в YAML)

Конфиг-файлы хранят только non-secret параметры. Секреты обязательно задаются переменными окружения и **накладываются поверх YAML на старте**:

| Env | Назначение | Обязательность |
|---|---|---|
| `AUTH_JWT_SECRET` | секрет для подписи JWT (≥ 32 байта, не плейсхолдер) | **обязательная** — сервис не стартует без неё |
| `AUTH_DB_PASSWORD` | пароль PostgreSQL | обязательная для prod; dev defaults to `admin` через compose |
| `AUTH_REDIS_PASSWORD` | пароль Redis | опциональная (пусто, если Redis без auth) |

`docker-compose.yaml` пробрасывает эти переменные через `${AUTH_JWT_SECRET:?…}` — compose откажется стартовать, если они не определены. Образец — `.env.example` в корне репо.

Сгенерировать JWT-секрет: `openssl rand -base64 48`.

### Параметры в YAML

```yaml
auth:
  jwt_secret: ""           # из AUTH_JWT_SECRET, оставлять пустым в файле
  access_ttl_seconds: 3600
  refresh_ttl_seconds: 2592000
  rate_limit_login_per_minute: 10
  rate_limit_register_per_minute: 5
  rate_limit_refresh_per_minute: 30
  bcrypt_cost: 12          # 0 = bcrypt.DefaultCost (10); 12 — разумный prod-уровень

server:
  grpc_addr: ":50050"
  tls:
    cert_file: ""          # см. раздел TLS ниже; пусто = plaintext gRPC
    key_file: ""
```

Валидация на старте отклоняет: пустой/коротенький `jwt_secret`, литеральный `CHANGE_ME_IN_PRODUCTION`, `bcrypt_cost` вне `[4..31]`, нулевой/отрицательный TTL.

## Запуск

Из корня репозитория:

```bash
# dev
AUTH_JWT_SECRET=$(openssl rand -base64 48) make up-dev

# prod (внешние postgres/redis)
AUTH_JWT_SECRET=… AUTH_DB_PASSWORD=… AUTH_REDIS_PASSWORD=… make up-prod
```

Пересобрать только auth: `make auth-rebuild-dev` / `make auth-rebuild-prod`.

Полезные команды: `make help`, `make logs`, `make ps`, `make down`, `make down-v`.

## Порты

- gRPC: `50050` (внутренний, expose в docker-сети `hr-net`).

## Хранилища

### PostgreSQL — миграции через goose

Схема живёт в `internal/storage/auth_storage/migrations/NNNNN_<name>.sql` и **встроена в бинарь** через `embed.FS`. На старте `NewAuthStorage`:

1. Парсит conn-string, открывает `pgxpool` под `context.WithTimeout(30s)`.
2. `Ping()` — реальный TCP/auth handshake (pgx ленив без него).
3. `goose.UpContext(...)` — применяет неприменённые миграции, ведёт служебную таблицу `goose_db_version`.

Текущая схема (миграция `00001_initial_schema.sql`):

```sql
CREATE TABLE IF NOT EXISTS auth_users (
    id            BIGSERIAL    PRIMARY KEY,
    email         VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    role          VARCHAR(50)  NOT NULL DEFAULT 'user',
    created_at    TIMESTAMP    NOT NULL DEFAULT NOW()
);
```

Чтобы добавить миграцию: создать `00002_<name>.sql` с блоками `-- +goose Up` / `-- +goose Down`, пересобрать образ. На следующем старте контейнера goose накатит её автоматически. Откат миграций (`goose down`) в бинарь не зашит — это сознательно, откат на проде должен быть осознанным шагом, не побочным эффектом запуска.

### Redis — ключи

| Ключ | Тип | TTL | Назначение |
|---|---|---|---|
| `session:<hex(refresh_hash)>` | STRING (JSON) | refresh_ttl | сессия пользователя |
| `user_sessions:<user_id>` | SET | — | вторичный индекс активных refresh-хешей пользователя; нужен, чтобы `LogoutAll` работал за O(sessions-of-user), а не SCAN по всему keyspace |
| `rl:login:ip:<addr>` | INT | 60s | rate-limit логина по IP клиента (приходит из gateway через metadata `x-client-ip`) |
| `rl:login:email:<sha256-prefix>` | INT | 60s | rate-limit логина по email-хешу (защита от credential stuffing) |
| `rl:register:ip:…`, `rl:register:email:…` | INT | 60s | то же для register |
| `rl:refresh:ip:…` | INT | 60s | rate-limit refresh по IP |

Запись и удаление сессии атомарны: `CreateSession` делает `SET + SADD` одним pipeline; `Refresh` использует `GETDEL` плюс best-effort `SREM` из индекса; `LogoutAll` делает `SMEMBERS → pipeline DEL session:<...> + DEL user_sessions:<uid>`.

## Безопасность

- **Пароли:** только bcrypt-хеш, cost через `auth.bcrypt_cost` (по умолчанию 10, prod-baseline 12).
- **Access-токены:** JWT HS256 с обязательным `exp`. Парсер пинит алгоритм через `WithValidMethods(["HS256"])` + `WithExpirationRequired()` — токен без `exp` отвергается. Claims: `sub`, `user_id`, `email`, `role`, `iat`, `exp`.
- **Refresh-токены:** opaque (32 случайных байта, hex). В Redis хранится только sha256-хеш; сам токен видит только клиент. Ротация **одноразовая**: `Refresh` атомарно `GETDEL`-ит запись, повторное использование того же токена → `Unauthenticated/INVALID_TOKEN`.
- **Rate-limit:** двойной — по IP и по email-хешу (для login/register). Атомарность счётчика обеспечена Lua-скриптом в Redis (INCR + EXPIRE без TTL-гонок).
- **Логи:** raw email никогда не пишется. В error-кейсах login/register логируется только `email_hash` (sha256-prefix). Mismatch refresh-токена при logout логируется server-side (`slog.Warn` с `caller_user_id` + `session_user_id`) и **не** уходит клиенту — клиент видит единый `INVALID_TOKEN` для «сессия не найдена» и «сессия чужого пользователя».
- **Interceptors:** `UnaryRecoveryInterceptor` (panic → `Internal` со стеком в логах), `UnaryLoggingInterceptor` (method/code/duration), `UnaryAuthInterceptor` (проверяет JWT для приватных RPC).
- **Graceful shutdown:** `signal.NotifyContext(SIGINT, SIGTERM)` → health flips to NOT_SERVING → `GracefulStop` (15s timeout, fallback `Stop`) → cleanup hooks LIFO (закрытие pgxpool, redis-клиента).

### TLS / mTLS

По умолчанию gRPC-сервер работает в plaintext — нормально внутри изолированной docker-сети. Для прода:

1. **Service mesh / sidecar (рекомендуется).** Istio, Linkerd, k8s + cilium и прочее даёт mTLS автоматически — приложение не знает о сертификатах. Это правильный путь, потому что ротация ключей и проверка SAN/SPIFFE-идентичности живут в инфраструктуре.
2. **Встроенный TLS как escape hatch.** Если service mesh недоступен:
   ```yaml
   server:
     grpc_addr: ":50050"
     tls:
       cert_file: "/run/secrets/auth.crt"
       key_file:  "/run/secrets/auth.key"
   ```
   При обоих заданных полях сервер поднимется с TLS. **Все клиенты (gateway в первую очередь) должны быть переконфигурированы на dial с `credentials.NewClientTLSFromFile(...)` — иначе RPC отвалятся `Unavailable`.** mTLS (проверка клиентского сертификата) этой опцией не настраивается — для этого тоже см. service mesh.

## Тестирование

Unit-тесты только на бизнес-логику в `internal/services/auth_service/` — без интеграционных, без real-DB/Redis, без handler-тестов. Storage интерфейсы мокаются через minimock.

```bash
make cov          # go test -cover ./internal/services/auth_service
make mock         # go generate ./internal/services/...  (перегенерация моков)
```

Стек: `gotest.tools/v3/assert` (ассерты) + `testify/suite` (группировка SetupTest) + `minimock` v3 (моки). Текущее покрытие пакета `auth_service` — 93.5%. Среди тестов — регрессии для refresh-replay (одноразовость) и uniform-error при logout чужой сессии.

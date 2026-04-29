# Auth Service — Audit (2026-04-29)

Аудит по навыкам: `golang-security`, `golang-grpc`, `golang-design-patterns`, `golang-modernize`, `golang-mastery`, `golang-testing`. Скоуп тестов: только unit на `internal/services/auth_service`.

## Done log

- **C1** (2026-04-29) — `config/config.go` теперь оверлеит секреты из env (`AUTH_JWT_SECRET`, `AUTH_DB_PASSWORD`, `AUTH_REDIS_PASSWORD`) и валидирует JWT-секрет на старте (>=32 байт, не плейсхолдер). YAML-конфиги вычищены, `docker-compose.yaml` пробрасывает env с `:?required`, `.env.example` обновлён.
- **C2** (2026-04-29) — добавлен атомарный `SessionStorage.ConsumeSessionByRefreshHash` через Redis `GETDEL`; `Refresh` теперь сначала consume, потом issue — повторный использованный refresh даёт `ErrInvalidRefreshToken`. Старый `RevokeSessionByRefreshHash` оставлен для Logout (поправим в H4).

## Сводка приоритетов

| #   | Severity | Файлы                                                              | Источник (skill)              | Фикс одной строкой                                                                |
| --- | -------- | ------------------------------------------------------------------ | ----------------------------- | --------------------------------------------------------------------------------- |
| ~~C1~~ | ✅       | ~~`config.docker.{dev,prod}.yaml`, `Dockerfile:14-15`~~              | ~~golang-security~~           | ~~Секреты (`jwt_secret`, db/redis password) — в env, валидировать на старте~~        |
| ~~C2~~ | ✅       | ~~`internal/services/auth_service/refresh.go:17-37`, `session_storage/revoke_by_hash.go:12-20`~~ | ~~golang-security, golang-design-patterns~~ | ~~Refresh: атомарный Redis `GETDEL` + revoke-then-issue (одноразовость refresh)~~     |
| H1  | 🟠       | `internal/api/auth_service_api/helpers.go:60-78`, `redis_ratelimit.go:46-81` | golang-security, golang-grpc  | rate-limit ключ — клиентский IP из metadata (gateway forward), не peer; + по email |
| H2  | 🟠       | `auth_service_api/login.go:29,47`, `register.go:29,47`, `me.go`, `validate_access_token.go` | golang-security (PII)         | Email убрать из логов или хешировать; user_id оставить                             |
| H3  | 🟠       | `auth_service_api/jwt_helpers.go:143-165`                          | golang-security, golang-modernize | `jwt.NewParser(WithValidMethods([]string{"HS256"}), WithExpirationRequired())`    |
| H4  | 🟠       | `auth_service/logout.go:14-21`, `logout_all.go:14-21`              | golang-security (info leak)   | Одна ошибка для «session not found» и «session not yours»                          |
| H5  | 🟠       | `bootstrap/server.go:13-30`                                        | golang-grpc, golang-mastery   | `signal.NotifyContext` + `GracefulStop` 15s fallback `Stop`; закрыть pgxpool/redis |
| M1  | 🟡       | `bootstrap/server.go:25-26`                                        | golang-grpc                   | `healthpb.RegisterHealthServer(s, health.NewServer())`                            |
| M2  | 🟡       | `auth_service_api/{logout,logout_all,me,update_user_role}.go`     | golang-grpc, golang-design-patterns | Interceptors: recovery + logging + auth (claims в ctx)                            |
| M3  | 🟡       | `session_storage/revoke_by_user_id.go:11-50`                       | golang-design-patterns (bounded) | Вторичный индекс `user_sessions:{user_id}` (SADD на create) вместо SCAN          |
| M4  | 🟡       | `storage/auth_storage/auth_storage.go:38-55`                       | golang-mastery                | Migrations отдельным инструментом (`pressly/goose` / `golang-migrate`)            |
| M5  | 🟡       | `storage/auth_storage/auth_storage.go:21,49`                       | golang-mastery, golang-context | `context.WithTimeout(..., 30s)` на pool init и initTables                         |
| M6  | 🟡       | `bootstrap/server.go:15`, `bootstrap/pgstorage.go:17`              | golang-mastery                | Bootstrap возвращает error; `panic`/`log.Panicf` → в `main` через `os.Exit(1)`    |
| M7  | 🟡       | `bootstrap/server.go:25`                                           | golang-grpc                   | TLS на gRPC или явно задокументировать «mTLS делает service mesh»                  |
| L1  | 🟢       | все `auth_storage/*.go` (`pkgerrors.Wrap`)                         | golang-modernize              | `pkg/errors` → `fmt.Errorf("...: %w", err)`                                       |
| L2  | 🟢       | `session_storage/get.go:20`, `revoke_by_hash.go:14`, `revoke_by_user_id.go:26` | golang-modernize             | `err == redis.Nil` → `errors.Is(err, redis.Nil)`                                  |
| L3  | 🟢       | `auth_service_api/jwt_helpers.go:105`                              | golang-modernize, golang-mastery | `fmt.Sscanf("%d", ...)` → `strconv.ParseUint(v, 10, 64)`                          |
| L4  | 🟢       | `revoke_by_user_id.go:13`, `redis_ratelimit.go:59-60`              | golang-modernize              | `var cursor uint64 = 0` → `var cursor uint64`; `ctx, cancel :=` одной строкой    |
| L5  | 🟢       | `auth_service/issue_tokens.go:65-72`                               | golang-design-patterns        | Удалить `tokenToHash` (alias на `hashRefreshToken`)                                |
| L6  | 🟢       | `auth_service/auth_service.go:33`                                  | golang-design-patterns        | `int64` секунды → `time.Duration` (конверсия в bootstrap); опционально func opts  |
| L7  | 🟢       | `auth_service/register.go:20`                                      | golang-security               | `bcrypt.DefaultCost` → конфигурируемый `AuthConfig.BcryptCost`                    |
| L8  | 🟢       | `auth_service_api/auth_api.go:30-47`                               | golang-design-patterns        | Гарантировать non-nil limiters в bootstrap, убрать `denyAllLimiter` fallback       |

## 🧪 Testing — пробелы

Skill: `golang-testing` (audit-mode). Scope: только unit на `internal/services/auth_service`. Стек: `gotest.tools/v3/assert` + `testify/suite`.

| Метод            | Что покрыть                                                                              |
| ---------------- | ---------------------------------------------------------------------------------------- |
| `Register`       | happy; email exists; bcrypt error; `CreateUser` error; `CreateSession` error             |
| `Login`          | happy; user not found → `ErrInvalidCredentials`; wrong password; storage error           |
| `Refresh`        | happy; expired; user gone; revoke fails; replay (после фикса C2)                          |
| `Logout`         | happy; чужой user → единая ошибка после H4; storage error                                |
| `LogoutAll`      | аналогично `Logout`                                                                      |
| `UpdateUserRole` | admin → success; non-admin → denied; self → `ErrCannotChangeOwnRole`; invalid role; target not found |
| `GetUserByID`    | found; not found                                                                         |
| `issueTokens`    | shape access/refresh, claims, корректный hash + expiresAt в session storage              |

Сопутствующее:

- `.mockery.yaml` мокает только `AuthStorage` — нужен мок `SessionStorage` (после миграции на minimock тоже).
- `validate_test.go:131,164` — дубль кейса «password too short».
- `assert.NilError(s.T(), err)` повторяется — можно вынести `t := s.T()` в начало метода.

## Рекомендованный порядок работ

1. **C1 + C2** — секреты в env, refresh-replay через GETDEL.
2. **H1 + H5** — клиентский IP через metadata + graceful shutdown.
3. **H2 + H3 + H4** — PII, JWT parser options, info leak.
4. **M1 + M2** — health-check + interceptors (часть M2 закрывает M6 для логирования).
5. **Tests** — покрытие методов по таблице выше (расширить `.mockery.yaml`).
6. **M3 + M4 + M5 + M7** — operability и schema migrations.
7. **L\*** — модернизация, чистка `pkg/errors`, `errors.Is`, `time.Duration`, мелочи.

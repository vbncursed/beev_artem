# admin — техническое задание

## Назначение

Операционный дашборд-сервис: aggregate-статистика платформы (счётчики
пользователей / вакансий / кандидатов / анализов), список HR-аккаунтов с
их активностью, изменение ролей. Доступен **только admin'ам** —
auth-interceptor отбивает любой запрос с `IsAdmin=false` ещё до handler'а
кодом `PermissionDenied`.

## Архитектура

Clean architecture, четыре слоя:

```
admin/internal/
├── domain/                       SystemStats, AdminUserView,
│                                 UpdateRoleInput
├── usecase/
│   └── admin_service.go          ports + service struct + методы
├── infrastructure/
│   ├── persistence/              read-only pgx queries
│   │   ├── stats_storage.go      pool init
│   │   ├── get_stats.go          один UNION ALL для всех счётчиков
│   │   └── list_users.go         JOIN auth_users + vacancies + candidates
│   └── auth_client/              gRPC клиент → auth (UpdateUserRole)
└── transport/
    ├── grpc/                     handlers
    │   ├── admin_api.go          server type + service interface
    │   ├── errors.go             errdetails.ErrorInfo с reason+domain
    │   ├── get_overview.go
    │   ├── list_users.go
    │   └── promote_user.go       PromoteUser + DemoteUser
    └── middleware/               Recovery + Logging + Auth (admin-only!)
```

## API

| RPC | HTTP | Описание |
|---|---|---|
| `GetOverview` | `GET /api/v1/admin/overview` | Aggregate-счётчики: users / admins / vacancies / candidates / analyses (total + done + failed) |
| `ListUsers` | `GET /api/v1/admin/users` | Все HR-аккаунты с ролью + активностью (количество вакансий и кандидатов) |
| `PromoteUser` | `POST /api/v1/admin/users/{user_id}/promote` | Обёртка над `auth.UpdateUserRole(role=admin)` |
| `DemoteUser` | `POST /api/v1/admin/users/{user_id}/demote` | То же, role=user |

Все требуют `Authorization: Bearer <admin-jwt>`. Не-admin токен → 403
`PermissionDenied`.

## Domain model

```go
type SystemStats struct {
    UsersTotal       uint64
    AdminsTotal      uint64
    VacanciesTotal   uint64
    CandidatesTotal  uint64
    AnalysesTotal    uint64
    AnalysesDone     uint64
    AnalysesFailed   uint64
}

type AdminUserView struct {
    ID                 uint64
    Email              string
    Role               string         // "admin" | "user"
    CreatedAt          time.Time
    VacanciesOwned     uint64         // COUNT через JOIN
    CandidatesUploaded uint64         // COUNT через JOIN
}
```

## Дизайн-решение: прямой read-only доступ к общей БД

В отличие от других сервисов, admin **читает напрямую** из таблиц всех
сервисов (`auth_users`, `vacancies`, `candidates`, `analyses`), а не зовёт
их gRPC API. Причины:

- На MVP все сервисы делят один postgres-контейнер с одной БД `hr` —
  cross-table SELECT возможен и быстр.
- Альтернатива — добавить `Count*` RPC в auth/vacancy/resume/analysis. Это
  +4 PR'а ради одного `SELECT count(*)`. Не стоит inkrement complexity.
- Admin — операционный сервис, его задача читать состояние всей системы.
  "Чистоту boundaries" нарушаем сознательно и документируем здесь.

Если архитектура переедет на per-service DB — придётся:
- Либо открывать пул на каждую БД (admin продолжает читать напрямую,
  только DBs стало больше)
- Либо добавить gRPC count-эндпоинты и переписать `get_stats.go` на
  fan-out gRPC

## Поток `PromoteUser` / `DemoteUser`

```
1. gateway проксирует POST → admin gRPC
2. middleware.UnaryAuthInterceptor валидирует JWT + проверяет IsAdmin
   → не admin → PermissionDenied
3. handler → usecase.UpdateRole(in)
4. usecase валидирует (role enum, ID != 0, IsAdmin=true)
5. authClient.UpdateUserRole(ctx, userID, newRole) → gRPC к auth
6. auth ВТОРОЙ раз проверяет admin (defense in depth) и обновляет БД
7. возвращаем UpdateRoleResponse → handler → gateway → frontend
```

## Зависимости

- **PostgreSQL** — read-only по 4 таблицам в `hr` БД (документировано как
  exception)
- **auth** (gRPC) — `ValidateAccessToken` для middleware, `UpdateUserRole`
  для proxy
- **Внешних gRPC зависимостей нет** (vacancy / resume / analysis не зовём)

## Конфигурация

```yaml
database: { host, port, username, password, name, ssl_mode }
auth:
  grpc_addr: "auth:50050"
server:
  grpc_addr: ":50056"
  tls: { cert_file, key_file }
```

| Env | Required | Описание |
|---|---|---|
| `ADMIN_DB_PASSWORD` | dev: optional, prod: required | пароль PostgreSQL |

## Запуск

```bash
make rebuild SVC=admin
docker exec hr-admin grpc_health_probe ...   # (если включить healthcheck)
```

## Тестирование

Пока не написаны. Когда появятся — конвенция как у других сервисов:
`internal/usecase` (white-box) + `gotest.tools/v3/assert` +
`testify/suite` + `gojuno/minimock`.

## Известные ограничения

- **Нет audit log** действий admin'а. `CallerUserID` в `UpdateRoleInput`
  зарезервирован под будущую таблицу `admin_audit_log(actor, action,
  target, at, payload)`.
- **Нет per-service breakdown** в `GetOverview` — один общий счётчик. Для
  filter-by-vacancy / filter-by-user понадобится отдельный endpoint.
- **`ListUsers` без пагинации** — на ≥10k пользователей понадобится
  cursor-based pagination.
- **Нет endpoint'ов `Get/Delete vacancy by ID`** — admin использует
  существующие vacancy/resume API с admin-bypass'ом
  (`($isAdmin OR owner_user_id = $caller)` в WHERE) уже работающим.

# beev

HR-платформа на микросервисах. Кандидат загружает резюме → backend
извлекает текст → эвристика + LLM (Yandex Cloud Foundation Models)
выставляют match score и обоснование → HR видит кандидатов
ранжированными по соответствию вакансии.

Шесть Go-сервисов общаются по gRPC; единственная HTTP-точка входа —
**gateway** — транслирует HTTP/JSON в gRPC через `grpc-gateway`.
Поверх лежит SPA на React 19 (**Cadence**), которая ходит только в
gateway.

## Сервисы

| Модуль | Назначение | gRPC | Публичный | Документ |
|---|---|---|---|---|
| [`auth/`](auth/README.md) | Идентификация, JWT, сессии, rate-limit | `:50050` | — | [README](auth/README.md) |
| [`vacancy/`](vacancy/README.md) | CRUD вакансий, авто-определение роли | `:50051` | — | [README](vacancy/README.md) |
| [`resume/`](resume/README.md) | Кандидаты + резюме (PDF/DOCX/TXT), `pdftotext` | `:50052` | — | [README](resume/README.md) |
| [`analysis/`](analysis/README.md) | Скоринг + аудит-аналитика | `:50054` | — | [README](analysis/README.md) |
| [`multiagent/`](multiagent/README.md) | LLM-вердикт через Yandex Cloud, role-aware промпты | `:50055` | — | [README](multiagent/README.md) |
| [`gateway/`](gateway/README.md) | HTTP edge (`grpc-gateway`), CORS, OpenAPI на `/docs` | — | **`:8080`** | [README](gateway/README.md) |
| [`frontend/`](frontend/README.md) | Cadence — React 19 + TS 6 + Vite + Tailwind v4 | — | **`:3000`** | [README](frontend/README.md) |

Каждый README содержит полный технический спек: clean-architecture
слои, API-контракт, доменная модель, зависимости, конфигурация,
тестирование, известные ограничения.

## Архитектура

```
[Browser] ──HTTP──▶ [frontend (nginx :80)] ─/api/*──▶ [gateway :8080]
                                                            │
                                                            │ gRPC
                                                ┌───────────┼───────────────┐
                                                ▼           ▼               ▼
                                          [auth]     [vacancy/resume]   [analysis]
                                            │              │                │ gRPC
                                            ▼              ▼                ▼
                                         Redis        PostgreSQL       [multiagent] ──HTTPS──▶ Yandex Cloud
                                                                            │              (Foundation Models)
                                                                            ▼
                                                                       PostgreSQL
```

Подробнее про auth flow, page rhythm и cleanup-протоколы — см.
[`CLAUDE.md`](CLAUDE.md).

## Быстрый старт

```bash
# 1. Скопировать .env.example, заполнить секреты
cp .env.example .env
# обязательно задать: AUTH_JWT_SECRET (openssl rand -base64 48)
# и YANDEX_API_KEY (console.cloud.yandex.ru → service account → API key)

# 2. Поднять весь стек (постгрес + редис + 6 сервисов + frontend)
make up

# 3. Открыть в браузере
open http://localhost:3000           # Cadence SPA
open http://localhost:8080/docs      # Scalar UI с OpenAPI спекой
```

Если фронт не нужен — `make up` всё равно поднимет его, можно
игнорировать. Для отдельных перезапусков:

```bash
make rebuild SVC=resume     # пересобрать один сервис
make logs                   # хвост логов всех контейнеров
make ps                     # статус контейнеров
make down                   # остановить
make down-v                 # остановить + удалить данные (postgres + redis)
```

## Admin

`Register` всегда создаёт пользователя с `role="user"`. Промоутить
кого-то в admin может только другой admin через `UpdateUserRole` RPC —
chicken-and-egg. Для bootstrap первого admin'а в auth-контейнере вшита
операционная CLI:

```bash
make admin-promote EMAIL=you@example.com    # обёртка над docker exec hr-auth admin promote
make admin-demote  EMAIL=you@example.com    # вернуть обратно в "user"
```

После смены роли существующий JWT всё ещё несёт старую — нужно
sign out + log in заново, чтобы получить токен с обновлёнными claims.
Admin видит **все** вакансии и резюме всех пользователей через те же
эндпоинты — отдельных admin-ручек нет, persistence-слой проверяет
`($isAdmin OR owner_user_id = $caller)`. Подробнее — в
[`auth/README.md`](auth/README.md#admin-cli).

## Тестирование

Юнит-тесты только для `internal/usecase` каждого сервиса (политика
монорепо: никаких integration / handler / storage тестов):

```bash
make test          # fan-out по всем сервисам
make cov           # coverage
make race          # с -race
make lint          # go vet
```

Текущее покрытие: auth 93.8%, vacancy 100%, resume 97.2%,
analysis 100%, multiagent 97.6%.

## Конфигурация

Каждый сервис имеет `config.docker.dev.yaml` и `config.docker.prod.yaml`.
Селектор: env `configPath` → иначе `APP_ENV=prod*` → иначе dev.

**Секреты только через env** (не в YAML):

| Env | Required | Описание |
|---|---|---|
| `AUTH_JWT_SECRET` | ✓ | подпись JWT (≥32 байт). Без него auth не стартует. |
| `YANDEX_API_KEY` | ✓ | Yandex Cloud Foundation Models. Без него multiagent не стартует. |
| `AUTH_DB_PASSWORD` | dev: optional, prod: required | пароль PostgreSQL |
| `AUTH_REDIS_PASSWORD` | пусто в dev | пароль Redis |

`docker-compose.yaml` использует `${VAR:?msg}` синтаксис — compose
откажется стартовать если обязательная env не задана.

## Стек

| Слой | Технология |
|---|---|
| Языки | Go 1.26 (бэкенд) · TypeScript 6 (фронт) |
| RPC | Protocol Buffers + gRPC + grpc-gateway |
| HTTP | стандартный `net/http` (gateway), nginx (frontend) |
| СУБД | PostgreSQL 17 (per-service БД) |
| Кэш / rate-limit | Redis 8 (только auth) |
| Миграции | `pressly/goose` v3, embed.FS |
| Тесты | `gotest.tools/v3/assert` + `testify/suite` + `gojuno/minimock` |
| LLM | Yandex Cloud Foundation Models (Responses API) |
| Frontend bundler | Vite 8 |
| UI стек | React 19, Tailwind v4 |
| Шрифты | General Sans (Fontshare), JetBrains Mono (self-hosted) |
| PDF извлечение | `pdftotext` из poppler-utils (внутри resume Docker image) |

## Документы в репо

- [`CLAUDE.md`](CLAUDE.md) — полное руководство по архитектуре, конвенциям,
  layout каждого сервиса, gotchas. Источник истины для AI-агентов и
  для нового разработчика.
- [`DESIGN.md`](DESIGN.md) — дизайн-система Cadence (Coinbase-style):
  токены цвета / типографики / радиусов / spacing, контракты компонентов,
  do/don't правила.
- [`<service>/README.md`](#сервисы) — технический спек по каждому модулю.

## Лицензия

Internal — TBD.

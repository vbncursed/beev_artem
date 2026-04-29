# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Repository overview

Multi-service HR system written in Go 1.26. Each service is an independent Go module with its own `go.mod`, `Dockerfile`, configs, and `internal/` tree. Services talk to each other over gRPC; one of them (`gateway`) exposes HTTP/JSON to the outside world via grpc-gateway.

Modules (top-level directories):

- `gateway/` — HTTP/JSON edge, only public-facing service. Translates HTTP → gRPC via grpc-gateway, serves OpenAPI/Scalar docs. Listens on `:8080`.
- `auth/` — auth & sessions. gRPC `:50050`. Owns users, JWT issue/refresh, rate limit, role checks. Backed by PostgreSQL + Redis.
- `vacancy/` — vacancies CRUD. gRPC `:50051`. PostgreSQL.
- `resume/` — candidates & resume files (PDF/DOCX/TXT). gRPC `:50052`. PostgreSQL.
- `analysis/` — resume scoring / candidate analysis. gRPC `:50054`. Calls `multiagent` for LLM-backed decisions.
- `multiagent/` — generates HR decision (`hire|maybe|no`) + feedback. gRPC `:50055`. PostgreSQL.

Each Go module is named `github.com/artem13815/hr/<service>`. There is no Go workspace file — modules are independent and compiled separately.

## Architecture

```
clients ──HTTP──▶ gateway:8080 ──gRPC──▶ auth:50050
                                   ├──▶ vacancy:50051
                                   ├──▶ resume:50052
                                   └──▶ analysis:50054 ──gRPC──▶ multiagent:50055
                                                                            │
                                                          PostgreSQL ◀──────┴── Redis (auth only)
```

Each service follows the same internal layout:

```
<service>/
  api/                    *.proto contracts (auth_api, vacancy_api, …) + vendored google/api
  cmd/app/main.go         entry point: load config → bootstrap.Init* → AppRun
  config/                 config struct + LoadConfig
  config.docker.dev.yaml  dev compose config (postgres=postgres, redis=redis)
  config.docker.prod.yaml prod config (external hosts; placeholders CHANGE_ME)
  internal/
    api/<service>_api/    gRPC server impl: validation, mapping, error translation
    bootstrap/            wiring: pgstorage.go, redis.go, server.go, <service>.go …
    domain/               clean-architecture domain entities and value objects (gateway has no domain layer; multiagent currently has none)
    pb/                   generated protobuf, grpc, grpc-gateway, swagger
    services/<svc>_svc/   business logic / use cases; defines storage interfaces (mocked for tests)
    storage/              concrete PostgreSQL/Redis impls of those interfaces
  scripts/
    command.mk            per-service make targets (run, mock, cov, generate-api)
    generate.sh           protoc invocation (uses grpc-gateway path from `go list -m`)
  Makefile                only `include ./scripts/command.mk`
  buf.gen.yaml            buf alternative (protoc is the actual generator used)
  .mockery.yaml           mock generation config (see "Mocks" below)
```

Config selection (uniform across services), in `cmd/app/main.go`:
1. env `configPath` if set;
2. else `APP_ENV=prod|production` → `config.docker.prod.yaml`;
3. else → `config.docker.dev.yaml`.

Service-to-service addresses are read from config (e.g. `auth.grpc_addr: auth:50050` in gateway, `multiagent.grpc_addr: multiagent:50055` in analysis). Inside the compose network they resolve via container names.

## Docker / Compose

`docker-compose.yaml` at repo root defines 7 services. Build context for every Go service is the repo root (`.`); each `Dockerfile` lives inside its service dir but is referenced as `dockerfile: <svc>/Dockerfile`. This is required because each Dockerfile does `COPY <svc>/go.mod <svc>/go.sum ./` and `COPY <svc>/ ./`.

Common Dockerfile pattern (identical across all 6 Go services, only paths/ports differ):

```dockerfile
FROM golang:1.26-alpine AS builder
WORKDIR /src/<svc>
COPY <svc>/go.mod <svc>/go.sum ./
RUN go mod download
COPY <svc>/ ./
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/<svc>-service ./cmd/app

FROM alpine:3.22
WORKDIR /app
COPY --from=builder /out/<svc>-service /usr/local/bin/<svc>-service
COPY --from=builder /src/<svc>/config.docker.dev.yaml ./config.docker.dev.yaml
COPY --from=builder /src/<svc>/config.docker.prod.yaml ./config.docker.prod.yaml
EXPOSE <port>
CMD ["<svc>-service"]
```

Static binary (`CGO_ENABLED=0`) → final image is plain `alpine:3.22`, no Go toolchain. Both dev and prod config files are baked into the image; `APP_ENV` selects which one is used at runtime. The gateway image additionally copies the swagger directory (`internal/pb/swagger`) for the `/swagger.json` and `/docs` endpoints.

Compose specifics:
- `postgres` (16-alpine) and `redis` (7-alpine) are gated by `profiles: ["dev"]` — they only start in dev. In prod, services connect to external hosts configured in `config.docker.prod.yaml`.
- Service `depends_on` infra uses `required: false` so prod runs without local postgres/redis. Gateway depends on all backend services with `required: true`.
- Only `gateway` publishes a port (`8080:8080`); all backends use `expose:` only — reachable on the internal `hr-net` bridge network.
- Volumes: `hr_postgres_data`, `hr_redis_data`.

The repo root `Makefile` is the entrypoint for compose:

| Target | What it does |
|---|---|
| `make up` / `make up-dev` | `APP_ENV=dev COMPOSE_PROFILES=dev docker compose up -d` (starts everything incl. local postgres/redis) |
| `make up-prod` | `APP_ENV=prod docker compose up -d` (no infra, expects external hosts) |
| `make rebuild-dev` / `make rebuild-prod` | same as up but `--build` |
| `make <svc>-rebuild-dev` / `-prod` | rebuild & restart a single service (`auth`, `gateway`, `vacancy`, `resume`, `analysis`, `multiagent`) |
| `make down` / `make down-v` | stop / full reset (`-v --remove-orphans`) |
| `make ps` / `make logs` | status / `logs -f --tail=200` |
| `make restart` / `restart-prod` | down + up |

`.env.example` only contains `APP_ENV=dev`.

## Per-service development (inside each service directory)

```bash
make run            # APP_ENV=dev go run ./cmd/app  (runs against local-host configured services)
make cov            # go test -cover ./internal/services/<svc>_service
make mock           # regenerate mocks (see below)
make generate-api   # regenerate protobuf via scripts/generate.sh
```

Standard Go commands inside a service module:

```bash
go test ./...
go test -run TestName ./internal/services/<svc>_service       # single test
go test -race ./...
go build ./cmd/app
```

`generate.sh` uses `protoc` (not buf) and resolves the grpc-gateway include path with `go list -m -f '{{.Dir}}' github.com/grpc-ecosystem/grpc-gateway/v2`. It produces three things: gRPC stubs (`internal/pb/...`), grpc-gateway handlers, and OpenAPI v2 (`internal/pb/swagger/`).

## Mocks (minimock)

Each service uses [gojuno/minimock](https://github.com/gojuno/minimock) v3.4.7. Mocks live under `internal/services/<svc>_service/mocks/<iface>_mock.go` (one file per interface, named `<Iface>Mock`).

Generation is driven by a `//go:generate` directive at the top of `internal/services/<svc>_service/<svc>_service.go`. The `-g` flag suppresses minimock's auto-emitted directive in the generated file so the only canonical source is the service file. The pinned `@v3.4.7` keeps the version reproducible without a tool entry in `go.mod`.

To regenerate per service: `make mock` (which runs `go generate ./internal/services/...`). Run after changing any storage interface.

Interfaces currently mocked (one per service, two for auth):
- auth: `AuthStorage`, `SessionStorage`
- vacancy: `VacancyStorage`
- resume: `ResumeStorage`
- analysis: `AnalysisStorage`
- multiagent: `DecisionStorage`

**Test API (minimock vs testify-mock):** minimock uses `EXPECT()` builder chains, not testify's `On(...).Return(...)`. Example:

```go
mock := mocks.NewAuthStorageMock(t)
mock.GetUserByEmailMock.Expect(ctx, "a@b").Return(&domain.User{ID: 42}, nil)
```

The mock struct registers a cleanup hook with the test that asserts all expectations were satisfied at the end of the test, so `t.Cleanup(...)` and `mock.AssertExpectations(t)` are not needed by the caller.

## Testing conventions

- **Scope: unit tests only, and only for the business-logic layer** (`internal/services/<svc>_service`). Do **not** add integration tests, HTTP/gRPC handler tests, storage tests against real Postgres/Redis, or end-to-end tests. The API layer, bootstrap, storage, and pb packages stay untested.
- Storage and other collaborators are mocked via interfaces defined in `services/<svc>_service` (see Mocks section). Tests must not hit a real DB, network, or filesystem.
- Assertions: **`gotest.tools/v3/assert`** (e.g. `assert.NilError(t, err)`, `assert.Equal(t, got, want)`, `assert.DeepEqual(...)`). **Do not** use `github.com/stretchr/testify/assert` or `testify/require` — even though `testify` appears in `go.mod`, it's only there for `suite`.
- Test suites: **`github.com/stretchr/testify/suite`** for grouping setup/teardown.
- Mocks: testify-mock-based today (mockery), planned migration to minimock (see Mocks section).
- Use `t.Context()` (Go 1.24+) instead of `context.Background()` in tests.
- Coverage target is the services package only — `make cov` already runs `go test -cover ./internal/services/<svc>_service`.

## Conventions worth knowing

- Protobuf source of truth lives under `<svc>/api/`. Generated Go is committed under `<svc>/internal/pb/`. Don't edit generated code; rerun `make generate-api`.
- `config.docker.prod.yaml` files contain `CHANGE_ME` placeholders and external host names — these MUST be replaced before a real prod deploy.
- gRPC API layer (`internal/api/<svc>_api`) is the only place that does request validation, error mapping (to gRPC status codes), and rate limiting. The `services/<svc>_service` layer is pure business logic against storage interfaces.
- All inter-service Go code uses `github.com/pkg/errors` for wrapping.
- Auth: rate limiters (login/register/refresh) are wired in `bootstrap/ratelimit.go` and use Redis.
- README files inside each service are written in Russian and are kept reasonably up to date — they're a good source of method-level intent.

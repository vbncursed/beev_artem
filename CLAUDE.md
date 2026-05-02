# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Repository overview

Multi-service HR platform written in Go 1.26. Each service is an independent Go module with its own `go.mod`, `Dockerfile`, configs, migrations, and `internal/` tree following clean architecture (domain / usecase / transport / infrastructure). Services communicate over gRPC; one of them (`gateway`) exposes HTTP/JSON to the outside world via grpc-gateway. The repo also ships a single-page **`frontend/`** (React 19 + TypeScript 6 + Vite + Tailwind v4) that consumes the gateway exclusively — frontend has no per-service network knowledge, only the public OpenAPI contracts under `frontend/api/*.json`.

Modules (top-level directories):

| Module | Role | gRPC port | Public? |
|---|---|---|---|
| `gateway/` | HTTP/JSON edge, only public-facing. Translates HTTP → gRPC via grpc-gateway, serves OpenAPI/Scalar docs. Listens on `:8080` (HTTP). | — | ✓ HTTP |
| `auth/` | Auth & sessions. Owns users, JWT issue/refresh, rate limit, role checks. Backed by PostgreSQL + Redis. | `:50050` | — |
| `vacancy/` | Vacancies CRUD. Carries the `role` field that drives multiagent prompt selection. PostgreSQL. | `:50051` | — |
| `resume/` | Candidates & resume files (PDF/DOCX/TXT). PostgreSQL. | `:50052` | — |
| `analysis/` | Resume scoring / candidate analysis. Calls `multiagent` for LLM-backed HR decisions. PostgreSQL. | `:50054` | — |
| `multiagent/` | Generates HR decision (`hire/maybe/no`) + structured feedback via Yandex Cloud Foundation Models, role-aware prompts. PostgreSQL. | `:50055` | — |
| `frontend/` | Vite-built SPA (Cadence brand) on gateway:8080. Clean-architecture mirror in TypeScript: domain → application → infrastructure → presentation. ru/en i18n, light/dark theme, drag&drop resume upload, AI analysis details. | — | — |

Each Go module is named `github.com/artem13815/hr/<service>`. There is no Go workspace file — modules are independent and compiled separately.

## Architecture

```
clients ──HTTP──▶ gateway:8080 ──gRPC──▶ auth:50050  (JWT issue/validate, sessions)
                                   ├──▶ vacancy:50051 (CRUD; carries role)
                                   ├──▶ resume:50052  (candidates + files)
                                   └──▶ analysis:50054 ──gRPC──▶ multiagent:50055 ──HTTPS──▶ Yandex Cloud
                                                                                          (Foundation Models)
                            PostgreSQL ◀───────┴── Redis (auth only)
```

### Auth flow (defense in depth)

1. Client sends `Authorization: Bearer <jwt>` to gateway.
2. **Gateway** does an edge fast-fail via `auth.ValidateAccessToken` for gated paths (see `gateway/internal/transport/http/auth.go:requiresAuth`). Bad token → 401 in one hop.
3. Bearer is propagated as gRPC metadata via `IncomingHeaderMatcher` to backend services.
4. **Each backend service** (auth/vacancy/resume/analysis) **independently** validates the JWT in its `UnaryAuthInterceptor` by calling `auth.ValidateAccessToken` again. Identity (`UserContext{UserID, Role, IsAdmin}`) is placed on `context.Context` via an unexported key. Handlers read it with `middleware.Get(ctx)` — never from raw metadata.
5. Multiagent has **no auth interceptor** — it is internal-only, reached only by analysis on the docker-compose network. If it ever becomes public-facing, copy the auth interceptor pattern from any data service.

The `x-user-id` / `x-user-role` / `x-user-email` headers gateway used to inject have been **removed** — they were a pre-clean-arch artifact and a spoofing vector. Bearer-only flows.

### Multiagent LLM flow

1. `analysis.StartAnalysis` runs the heuristic `Scorer` against the resume + vacancy skills, persists the result, and returns immediately.
2. If the caller passed `UseLLM=true`, analysis fans out to multiagent (synchronously today; could be moved to an outbox-fed worker later).
3. `multiagent.GenerateDecision` looks up `assets/prompts/<role>.txt` (programmer / manager / accountant / default), renders the request as JSON, and calls Yandex Cloud Foundation Models (`/v1/responses`) via the `infrastructure/llm/yandex` adapter.
4. The JSON response is parsed (with markdown-fence stripping for chatty models), validated against the schema baked into the prompts, and returned. The (request, response) pair is persisted as an audit row in `multiagent_decisions`.
5. Back in analysis, if multiagent succeeded, the heuristic AI in the analysis row is overwritten with the LLM's decision via `UpdateAIDecision`. **LLM failures are swallowed** — heuristic AI stays as the authoritative fallback.

Rate limit on the multiagent side: token-bucket via `golang.org/x/time/rate`, default **10 rps / burst 5**, defends Yandex billing from a runaway loop in analysis.

## Clean architecture layout (per service)

All five backend services follow the same shape. Gateway is a transport-only edge — it has no domain or usecase layer.

```
<service>/
├── api/                                   *.proto contracts (versioned per service)
│   ├── <svc>_api/<svc>.proto              public service contract
│   ├── auth_api/auth.proto                client copy when service consumes auth
│   ├── multiagent_api/multiagent.proto    (analysis only) client copy of multiagent
│   ├── models/*.proto                     domain message protos
│   ├── common/common.proto                shared (Page, Sort, etc.)
│   └── google/api/                        vendored grpc-gateway annotations
├── cmd/app/main.go                        run() error pattern, slog.Error+os.Exit
├── config/config.go                       Config + LoadConfig + validate()
├── config.docker.dev.yaml                 dev compose config (postgres=postgres, redis=redis)
├── config.docker.prod.yaml                prod config (external hosts; CHANGE_ME placeholders)
├── Dockerfile                             multi-stage golang:1.26-alpine → alpine:3.22 (CGO_ENABLED=0)
├── go.mod
├── internal/
│   ├── bootstrap/                         wiring + lifecycle
│   │   ├── server.go                      AppRun: SIGINT/SIGTERM → GracefulStop + LIFO cleanup hooks
│   │   ├── pgstorage.go                   InitPGStorage (returns *Storage, error)
│   │   ├── <svc>_service.go               InitXxxService — wires storage + ports → usecase
│   │   ├── <svc>_api.go                   InitXxxServiceAPI — wraps usecase in transport
│   │   └── auth_client.go                 (data services) gRPC dial to auth for ValidateAccessToken
│   ├── domain/                            transport-independent types (no pb, no pgx)
│   │   ├── domain.go                      core entities
│   │   ├── errors.go                      (resume) sentinel errors
│   │   └── …                              service-specific (e.g. analysis_pipeline.go)
│   ├── infrastructure/                    adapters that implement usecase ports
│   │   ├── persistence/                   pgxpool + goose migrations
│   │   │   ├── <svc>_storage.go           Pool init, Close, embed.FS migrations
│   │   │   ├── migrations/00001_*.sql     goose Up/Down
│   │   │   ├── <method>.go                one method per file (create.go, get.go, …)
│   │   │   └── helpers.go                 newID / marshalJSON / etc.
│   │   ├── auth_client/client.go          (data services) gRPC client to auth
│   │   ├── multiagent_client/client.go    (analysis only) gRPC client to multiagent
│   │   ├── extractor/, profile/           (resume only) text + structured-profile extraction
│   │   ├── scorer/scorer.go               (analysis only) heuristic Scorer port impl
│   │   ├── llm/yandex/client.go           (multiagent only) Yandex Foundation Models adapter
│   │   └── prompts/                       (multiagent only) embedded role prompts
│   │       ├── store.go                   Store implementing usecase.PromptStore
│   │       └── templates/<role>.txt       programmer / manager / accountant / default
│   ├── pb/                                generated protobuf, grpc, grpc-gateway (gateway also has openapi/openapi.yaml)
│   ├── transport/
│   │   ├── grpc/                          gRPC handlers, conversions, errors
│   │   │   ├── <svc>_api.go               server type + service interface (consumer port)
│   │   │   ├── <method>.go                one handler per file
│   │   │   ├── helpers.go                 pb<->domain mapping
│   │   │   ├── conversions.go             (multiagent only) pb<->domain
│   │   │   └── errors.go                  newError using errdetails.ErrorInfo (Reason+Domain)
│   │   └── middleware/middleware.go       Recovery + Logging + Auth interceptors
│   └── usecase/                           business logic, no I/O
│       ├── <svc>_service.go               port interfaces + service struct + constructor
│       ├── errors.go                      ErrInvalidArgument / ErrNotFound / ErrUnauthorized
│       ├── <method>.go                    one method per file
│       ├── llm_port.go                    (multiagent only) LLM + PromptStore ports
│       └── mocks/                         minimock-generated, regenerated by `make mock`
└── scripts/generate.sh                    protoc invocation (uses go list -m for grpc-gateway path)
```

### Gateway-specific layout

Gateway has **no domain or usecase** (it's a pure transport composition):

```
gateway/internal/
├── bootstrap/
│   ├── server.go                          AppRun: HTTP server lifecycle
│   ├── auth_client.go                     init wrapper
│   ├── gateway_mux.go                     register all 4 backend grpc-gateway handlers
│   ├── swagger.go                         load merged OpenAPI 3 spec, YAML→JSON for /swagger.json
│   └── http_handler.go                    compose root handler tree
├── infrastructure/auth_client/client.go   gRPC client to auth (same shape as data services)
└── transport/http/
    ├── middleware.go                      logging / clientIP / jsonContentType / IncomingHeaderMatcher
    ├── auth.go                            WithAuthContext + extractBearerToken + requiresAuth + writeUnauthorized
    ├── swagger.go                         /swagger.json + /docs handlers
    └── health.go                          /healthz
```

### Service variations

| Aspect | auth | vacancy | resume | analysis | multiagent | gateway |
|---|---|---|---|---|---|---|
| Has `domain/` | ✓ | ✓ | ✓ | ✓ | ✓ | ✗ |
| Has `usecase/` | ✓ | ✓ | ✓ | ✓ | ✓ | ✗ |
| Has auth interceptor | client-side via JWT validator (it's the source of truth) | ✓ | ✓ | ✓ | ✗ (internal-only) | edge fast-fail only |
| Has streaming RPC | — | — | ✓ (UploadResume) | — | — | — |
| Persists data | users + sessions | vacancies + skills | candidates + resumes | analyses | multiagent_decisions (audit) | — |
| External calls | — | — | — | multiagent | Yandex Cloud | — |

## Config

`config/config.go` per service. All five data services + gateway use the same `LoadConfig(filename) (*Config, error)` pattern with **`validate()`** that rejects empty required fields at boot.

### Config selection (in `cmd/app/main.go`)

```go
configPath := cmp.Or(os.Getenv("configPath"), defaultConfigPathByEnv(os.Getenv("APP_ENV")))
```

1. env `configPath` if set (overrides everything);
2. else `APP_ENV=prod|production` → `config.docker.prod.yaml`;
3. else (incl. `APP_ENV=dev`) → `config.docker.dev.yaml`.

### Common fields across services

```yaml
database:
  host: postgres   # in dev profile
  port: 5432
  username: admin
  password: admin
  name: hr
  ssl_mode: disable

server:
  grpc_addr: ":<port>"
  tls:                 # opt-in; both empty → plaintext
    cert_file: ""
    key_file: ""
```

Service-specific blocks: `auth.grpc_addr` (resume / vacancy / analysis / gateway), `multiagent.grpc_addr` (analysis), `yandex.{folder_id, model, request_timeout, max_output_tokens}` + `rate_limit.{rps, burst}` (multiagent), `redis` (auth).

### Secrets via env (never YAML)

| Var | Required | Used by | Generate / source |
|---|---|---|---|
| `AUTH_JWT_SECRET` | ✓ | auth | `openssl rand -base64 48` |
| `AUTH_DB_PASSWORD` | dev: optional (defaults `admin` in compose), prod: required | auth | secret manager |
| `AUTH_REDIS_PASSWORD` | empty in dev | auth | secret manager |
| `YANDEX_API_KEY` | ✓ | multiagent | console.cloud.yandex.ru → service account → API key |

`AUTH_JWT_SECRET` is enforced by compose itself: `${AUTH_JWT_SECRET:?AUTH_JWT_SECRET is required (>=32 bytes, not the placeholder)}` — the auth container will not start without it.

## Docker / Compose

`docker-compose.yaml` at repo root defines 8 services. Build context for every Go service is the repo root (`.`); each `Dockerfile` lives inside its service dir but is referenced as `dockerfile: <svc>/Dockerfile`. This is required because each Dockerfile does `COPY <svc>/go.mod <svc>/go.sum ./` and `COPY <svc>/ ./`.

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

Static binary (`CGO_ENABLED=0`) → final image is plain `alpine:3.22`, no Go toolchain. Both dev and prod config files are baked into the image; `APP_ENV` selects which one is used at runtime. Multiagent additionally embeds the `assets/prompts/*.txt` via `//go:embed` (no extra Dockerfile changes needed). Gateway image additionally copies `internal/pb/openapi/openapi.yaml` (the merged OpenAPI 3.0 document gnostic produces) so it can serve `/swagger.json` and `/docs`.

### Port topology (host vs compose-internal)

| Service | Host | Compose | Why |
|---|---|---|---|
| `gateway` | `8080:8080` | — | Public HTTP edge |
| `postgres` | `5432:5432` | — | Dev convenience for DBeaver/psql |
| `redis` | `6379:6379` | — | Dev convenience for redis-cli |
| `auth` | — | `expose: 50050` | Internal — only gateway and other backends call it |
| `vacancy` | — | `expose: 50051` | Internal |
| `resume` | — | `expose: 50052` | Internal |
| `analysis` | — | `expose: 50054` | Internal |
| `multiagent` | — | `expose: 50055` | Internal — only analysis calls it |

`expose:` only exposes inside the `hr-net` bridge; Docker does not allocate a host-side NAT entry. Backend gRPC ports are unreachable from `localhost`. For prod hardening, also drop `ports:` from postgres / redis (use a bastion or SSH tunnel).

### Compose specifics

- `postgres` (17-alpine) and `redis` (8-alpine) are gated by `profiles: ["dev"]` — they only start in dev. In prod, services connect to external hosts configured in `config.docker.prod.yaml`.
- Service `depends_on` infra uses `required: false` so prod runs without local postgres/redis. Gateway depends on all backend services with `required: true`.
- Volumes: `postgres_data`, `redis_data`.
- Env injection: `AUTH_JWT_SECRET`, `AUTH_DB_PASSWORD`, `AUTH_REDIS_PASSWORD` flow into auth; `YANDEX_API_KEY` flows into multiagent.

### Root Makefile

| Target | What it does |
|---|---|
| `make help` | List all targets (auto-derived from `## ` comments) |
| `make up` | `APP_ENV=dev COMPOSE_PROFILES=dev docker compose up -d` |
| `make up-prod` | `APP_ENV=prod docker compose up -d` (no infra; external hosts) |
| `make up-build` / `make up-build-prod` | Same as above with `--build` |
| `make rebuild SVC=<name> [ENV=prod]` | Rebuild + restart one service. Replaces 12 hand-written per-service targets |
| `make down` / `make down-v` | Stop / full reset (`-v --remove-orphans`) |
| `make restart` / `make restart-prod` | down + up |
| `make ps` / `make logs` / `make pull` | status / `logs -f --tail=200` / pull |
| `make test` | Fan out `go test -count=1 ./internal/usecase` over services that have a usecase layer (skips gateway). |
| `make cov` | Same, but `go test -cover`. |
| `make race` | Same, with `-race`. |
| `make lint` | `go vet $(go list ./... \| grep -v /mocks)` over all 6 services. |
| `make generate-api` | Re-run `bash <svc>/scripts/generate.sh` for every service. |

## Per-service development (inside each service directory)

Each service has its own `Makefile`. Services with a usecase layer (auth/vacancy/resume/analysis/multiagent) share the same target set:

| Target | Command |
|---|---|
| `make help` | List targets |
| `make generate-api` | `bash scripts/generate.sh` (protoc + go + grpc + gateway; gateway additionally emits OpenAPI 3 via gnostic) |
| `make mock` | `go generate ./internal/usecase` (regenerates minimock files) |
| `make test` | `go test -count=1 ./internal/usecase` |
| `make cov` | `go test -cover ./internal/usecase` |
| `make race` | `go test -race -count=1 ./internal/usecase` |
| `make lint` | `go vet $(go list ./... \| grep -v /mocks)` |
| `make tidy` | `go mod tidy` |

Gateway has only `help / generate-api / lint / tidy` (no usecase / no business logic).

**No `make run`** — services boot only through docker-compose. Use `make rebuild SVC=<name>` for fast iteration.

`generate.sh` uses `protoc` (not buf) and resolves the grpc-gateway include path with `go list -m -f '{{.Dir}}' github.com/grpc-ecosystem/grpc-gateway/v2`. Backend services emit two artifacts: gRPC stubs (`internal/pb/...`) and grpc-gateway handlers. Only the gateway emits docs: a single merged **OpenAPI 3.0** document at `internal/pb/openapi/openapi.yaml` produced by gnostic's `protoc-gen-openapi` (one invocation across all four service protos). The gateway loads it at boot and serves it as JSON at `/swagger.json` (Scalar UI at `/docs` reads that endpoint).

## Migrations (goose + embedded SQL)

Each persistence package owns its own `migrations/NNNNN_name.sql` with goose `Up` / `Down` blocks, embedded into the binary via `//go:embed migrations/*.sql`.

Lifecycle in `persistence/<svc>_storage.go`:

1. `pgxpool.NewWithConfig(ctx, config)` — pool created with a 30-second `initTimeout` ctx.
2. `db.Ping(ctx)` — pgx v5 connects lazily, force a real handshake so the timeout applies to TCP/auth.
3. `applyMigrations(ctx, connString)` — opens a short-lived `*sql.DB` (goose's API requires it), runs `goose.UpContext` against the embed.FS.
4. Pool returned, ready for handlers.

Adding a column: drop a new `NNNNN_descriptive.sql` with `ALTER TABLE … ADD COLUMN IF NOT EXISTS …`, redeploy. Goose tracks state in its own `goose_db_version` table.

## Mocks (minimock)

Each service uses [gojuno/minimock](https://github.com/gojuno/minimock) v3.4.7. Mocks live under `internal/usecase/mocks/<iface>_mock.go` (one file per interface, named `<Iface>Mock`). The `//go:generate` directive lives at the top of `internal/usecase/<svc>_service.go` (and a few sibling files like `multiagent/internal/usecase/llm_port.go`).

Examples of generated mocks per service:

| Service | Interfaces mocked |
|---|---|
| auth | `AuthStorage`, `SessionStorage`, `TokenIssuer`, `RateLimiter` |
| vacancy | `VacancyStorage` |
| resume | `ResumeStorage`, `TextExtractor`, `ProfileExtractor` |
| analysis | `AnalysisStorage`, `Scorer` |
| multiagent | `DecisionStorage`, `LLM`, `PromptStore` |

To regenerate per service: `make mock` (which runs `go generate ./internal/usecase`).

**API: minimock vs testify-mock.** minimock uses `EXPECT()` builder chains, **not** testify's `On(...).Return(...)`. Example:

```go
mock := mocks.NewVacancyStorageMock(t)
mock.GetVacancyMock.Expect(ctx, "v-1", uint64(7), false).Return(want, nil)
// or with arg-by-arg overrides:
mock.GetVacancyMock.ExpectVacancyIDParam2("v-1").Return(want, nil)
// or with Inspect for asserting captured args:
mock.SaveAnalysisMock.Inspect(func(_ context.Context, in domain.SaveAnalysisInput) {
    assert.Equal(t, in.VacancyID, "v-override")
}).Return(nil)
```

The mock struct registers a cleanup hook with the test that asserts all expectations were satisfied, so callers don't need `t.Cleanup(...)` or `mock.AssertExpectations(t)`.

For the `Inspect` callback, the function signature **must match the interface method exactly** (including `context.Context` as the first param — `any` doesn't satisfy the type check).

## Testing conventions

- **Scope: unit tests only, and only for the business-logic layer (`internal/usecase`).** Do **not** add integration tests, HTTP/gRPC handler tests, storage tests against real Postgres/Redis, or end-to-end tests. The API layer, bootstrap, persistence, infrastructure, and pb packages stay untested.
- **Tests live in `package usecase`** (white-box). The exception is multiagent where the LLM mock subpackage references `domain.CompletionRequest` to break a cycle — see the comment in `multiagent/internal/usecase/llm_port.go`.
- Storage and other collaborators are mocked via interfaces defined in `internal/usecase`. Tests must not hit a real DB, network, or filesystem.
- **Assertions: `gotest.tools/v3/assert`** (e.g. `assert.NilError(t, err)`, `assert.Equal(t, got, want)`, `assert.DeepEqual(...)`, `assert.ErrorIs(t, err, sentinel)`).  **Do not** use `github.com/stretchr/testify/assert` or `testify/require` — even though `testify` appears in `go.mod`, it's only there for `suite`.
- **Test suites: `github.com/stretchr/testify/suite`** for grouping setup/teardown. Each per-method test file declares a `XxxSuite struct{ baseSuite }` and a `TestXxxSuite(t *testing.T) { suite.Run(t, new(XxxSuite)) }` runner.
- A shared `service_test.go` per service holds a `baseSuite` that wires fresh mocks in `SetupTest`. Each suite embeds `baseSuite`.
- **Mocks: minimock** (see above).
- **Use `t.Context()`** (Go 1.24+) instead of `context.Background()` in tests.
- Coverage target is the usecase package only — `make cov` runs `go test -cover ./internal/usecase`.
- Current coverage: auth 93.8%, vacancy 100%, resume 97.2%, analysis 100%, multiagent 97.6%.

## Modern Go conventions (Go 1.26)

- `any` everywhere (no `interface{}`)
- `cmp.Or(a, b, …)` for default values (Go 1.22)
- `min` / `max` builtins (Go 1.21); `min(cmp.Or(in.Limit, 20), 100)` is the canonical limit-clamp idiom
- `slices.SortFunc` + `cmp.Compare` (Go 1.21) — never `sort.Slice`
- `for i := range N` (Go 1.22) — never `for i := 0; i < N; i++`
- `t.Context()` in tests (Go 1.24)
- `errors.Is` / `errors.AsType[T]` — never `==` for errors, never `errors.As(err, &target)` for new code
- `strings.Cut` over `strings.SplitN(s, sep, 2)`
- `errors.Is(err, http.ErrServerClosed)` — never `err == http.ErrServerClosed`
- `crypto/rand` for tokens (no `math/rand`)
- `log/slog` for structured logging
- `errdetails.ErrorInfo{Reason, Domain}` for gRPC errors via `status.WithDetails`, never JSON in the message field

## Conventions worth knowing

- Protobuf source of truth lives under `<svc>/api/`. Generated Go is committed under `<svc>/internal/pb/`. Don't edit generated code; rerun `make generate-api`.
- Multiple services share proto files (e.g. `auth_api/auth.proto` exists in auth, vacancy, resume, analysis, gateway — each with the appropriate `go_package` so each compiles to its own `internal/pb/`). When changing such a proto, update **every** copy and regenerate **every** consumer.
- `config.docker.prod.yaml` files contain `CHANGE_ME` placeholders and external host names — these MUST be replaced before a real prod deploy.
- gRPC API layer (`internal/transport/grpc`) is the **only place** that does request validation (gRPC-level), error mapping (`errdetails.ErrorInfo`), pb↔domain conversion, and rate limiting. The `usecase` layer is pure business logic against ports. The `infrastructure/persistence` layer is dumb SQL — no business logic.
- All inter-service Go code uses `fmt.Errorf("…: %w", err)` for wrapping. The `github.com/pkg/errors` import is **legacy** — don't add it to new files.
- Auth: rate limiters (login/register/refresh) are wired in `bootstrap/ratelimit.go` and use Redis.
- Gateway TLS is opt-in via `config.docker.*.yaml`'s `server.tls`. For real production, prefer service-mesh-managed mTLS (Istio, Linkerd) over hand-rolled cert files in YAML.
- Cleanup hooks run **LIFO** in `bootstrap.AppRun`. Construction order matters: the auth-client conn must close after the gRPC server stops accepting requests, otherwise in-flight handlers calling `auth.ValidateAccessToken` would race the conn close.
- Multiagent: prompts live in `internal/infrastructure/prompts/templates/<role>.txt`. Adding a new role = drop a new `.txt` file + rebuild. No proto/schema/storage changes. Roles are case-insensitive, fall back to `default.txt` on miss.
- Multiagent does NOT carry an internal heuristic fallback. If the LLM fails (provider down, malformed JSON), the error propagates to analysis, which keeps its **heuristic AI** as the authoritative answer in the analysis row.
- READMEs inside each service are written in Russian and are kept reasonably up to date — they're a good source of method-level intent.
- Each service ships a `SPEC.md` (Russian) at its root with the full technical specification: ports, endpoints, domain model, dependencies, configuration, deployment notes. Frontend has the same at `frontend/SPEC.md`. When a service contract changes, update the corresponding SPEC.md in the same commit.

## Recent significant decisions

- **2026-04-29:** Full clean architecture refactor for auth and resume — `services/<svc>_service` → `internal/usecase`, `storage/<svc>_storage` → `internal/infrastructure/persistence` with goose migrations, `api/<svc>_service_api` → `internal/transport/grpc` with `errdetails.ErrorInfo`. New `internal/transport/middleware/middleware.go` (Recovery + Logging + Auth) and `internal/infrastructure/auth_client/`.
- **2026-04-29:** Vacancy mirrors the same shape; `make rebuild` collapses 12 per-service targets.
- **2026-04-30:** Analysis + multiagent migrated to clean architecture. Multiagent's pb leak closed (introduced `domain.DecisionRequest/Response`); analysis's scoring algorithm extracted from `infrastructure/persistence/helpers.go` into a `Scorer` port + `infrastructure/scorer/` adapter.
- **2026-04-30:** Vacancy gains `Role string` field (proto + migration `00002_add_role.sql` + storage SQL + transport mapping). Analysis tunnels role through `ResumeContext.VacancyRole` to multiagent's `DecisionRequest.Role`.
- **2026-04-30:** Multiagent heuristic stub replaced with Yandex Cloud Foundation Models adapter. New `usecase.LLM` and `usecase.PromptStore` ports, `infrastructure/llm/yandex/` adapter (rate-limited, ctx-aware, JSON-mode prompts, error mapping HTTP→domain), and `infrastructure/prompts/` (embed.FS templates per role: programmer / manager / accountant / default).
- **2026-04-30:** Gateway clean refactor (transport-only edge: no domain/usecase). Added graceful shutdown, `/healthz`, full HTTP timeouts, config validate(). Removed dead `x-user-id` header injection — backends ignore those headers post-refactor.
- **2026-05-02:** Multiagent prompts hard-pin Russian replies via `languageDirective` constant appended to every role prompt. `hr_recommendation` enum stays English (`hire`/`maybe`/`no`); rationale, feedback, soft-skills notes are RU. Vacancy auto-detects role from title+description via `usecase.DetectRole` (keyword tables for accountant / doctor / electrician / analyst / manager / programmer + `default` fallback).
- **2026-05-02:** Cadence frontend shipped under `frontend/`. Coinbase-style design system locked in `DESIGN.md` (single accent #0052ff, type at weight 400, pill-rounded CTAs, 96px section rhythm). React 19 + Tailwind v4 with CSS-var theming, ru/en i18n with proper RU plural forms, no second brand color.
- **2026-05-02:** Resume gains two endpoints — `GET /api/v1/resumes/:id/download` (returns original file bytes for manual review) and `DELETE /api/v1/candidates/:id` (cascades to resumes via FK). Frontend wires both to AnalysisDetails with inline-confirm delete UX.
- **2026-05-02:** Analysis heuristic strings russified (HRRationale / CandidateFeedback / SoftSkillsNotes) and `Profile.YearsExperience` extracted via regex (`\d+ лет/год/year`). Extra-skills tokenizer trims trailing `-_.` and applies a 16-word stop-list to filter generic noise (tool / info / data / team …). `infrastructure/scorer/scorer.go` split by responsibility into `scorer.go` (algorithm) + `extras.go` (tokenization) + `profile.go` (years/summary helpers).
- **2026-05-02:** Resume PDF extraction switched to external `pdftotext` from poppler-utils (Dockerfile installs `poppler-utils`). The in-process `ledongthuc/pdf` cannot recover word boundaries on PDFs without positional metadata in the content stream — typical for Cyrillic résumés where the result was gibberish like "ЭдуардКурочкинGo-разработчик". `extractViaLedongthuc` is kept as a best-effort fallback when the binary is missing.
- **2026-05-02:** Per-service refactor — every source file ≤200 LOC. Three identical `middleware.go` (analysis/resume/vacancy, 209 LOC each) split into `middleware.go` (types) + `recovery.go` + `logging.go` + `auth.go`. `multiagent/internal/usecase/multiagent_service.go` (214 LOC) split into the service (with constants + GenerateDecision) and `decision_parser.go` (LLM wire-shape + JSON fence stripping).

## Help with feedback

If the user asks for help or wants to give feedback, point them at:
- `/help` — built-in Claude Code help
- Issues: `https://github.com/anthropics/claude-code/issues`

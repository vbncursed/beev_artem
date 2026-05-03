# gateway — техническое задание

## Назначение

Единственная HTTP-точка входа в платформу beev. Транслирует HTTP/JSON в
gRPC через `grpc-gateway`, валидирует Bearer-токены на edge-уровне (fast-
fail), проксирует Authorization заголовок дальше как gRPC-метаданные.
Также сервит OpenAPI-спеки и Scalar UI на `/docs`. Не содержит бизнес-
логики и не имеет своего хранилища данных.

## Архитектура

**Transport-only edge**:

```
gateway/internal/
├── bootstrap/                    композиция и lifecycle
│   ├── server.go                 AppRun: HTTP server + graceful shutdown
│   ├── auth_client.go            init wrapper над auth_client
│   ├── gateway_mux.go            регистрирует grpc-gateway хэндлеры
│   │                             всех 4 backend сервисов
│   ├── swagger.go                загружает merged OpenAPI 3 spec
│   │                             (yaml→json для /swagger.json)
│   └── http_handler.go           композиция middleware:
│                                 cors → log → ip → json-content →
│                                 auth-fastfail → mux
├── infrastructure/
│   └── auth_client/              gRPC client → auth.ValidateAccessToken
└── transport/http/
    ├── middleware.go              logging / clientIP / jsonContentType /
    │                              IncomingHeaderMatcher / CORS
    ├── auth.go                    extractBearerToken + requiresAuth +
    │                              writeUnauthorized + WithAuthContext
    ├── swagger.go                 /swagger.json + /docs handlers
    └── health.go                  /healthz endpoint
```

Нет `domain/` и `usecase/` — это намеренно: gateway чистый транспортный
слой.

## Эндпоинты

### Системные

| Path | Описание |
|---|---|
| `GET /healthz` | Liveness probe; всегда 200 OK. |
| `GET /docs` | Scalar UI для интерактивной документации. |
| `GET /swagger.json` | Merged OpenAPI 3.0 spec (4 сервиса). |

### Прокси на бэкенды (через grpc-gateway)

Все ниже идут с auth fast-fail на edge-уровне (требуют валидный JWT
кроме `/auth/login`, `/auth/register`, `/auth/refresh`).

| Path | Backend |
|---|---|
| `POST /api/v1/auth/login` | auth |
| `POST /api/v1/auth/register` | auth |
| `POST /api/v1/auth/refresh` | auth |
| `GET /api/v1/auth/me` | auth |
| `POST /api/v1/auth/logout` | auth |
| `POST /api/v1/auth/logout-all` | auth |
| `GET\|POST\|PATCH /api/v1/vacancies/...` | vacancy |
| `POST /api/v1/vacancies/{id}/archive` | vacancy |
| `POST /api/v1/vacancies/{id}/candidates*` | resume |
| `GET /api/v1/candidates/{id}` | resume |
| `DELETE /api/v1/candidates/{id}` | resume |
| `GET /api/v1/resumes/{id}` | resume |
| `GET /api/v1/resumes/{id}/download` | resume |
| `POST /api/v1/resumes/intake` | resume |
| `POST /api/v1/resumes/intake/batch` | resume |
| `POST /api/v1/resumes/{id}/analyze` | analysis |
| `GET /api/v1/analyses/{id}` | analysis |
| `GET /api/v1/vacancies/{id}/candidates` | analysis |

multiagent **не выставлен** — internal-only, доступен только analysis.

## Auth flow на edge

`transport/http/auth.go`:

1. `requiresAuth(path)` определяет, нужно ли проверять Bearer (всё кроме
   public auth-эндпоинтов).
2. На gated path: `extractBearerToken(r)` → `auth.ValidateAccessToken`
   через grpc-client.
3. Невалид → 401 (`writeUnauthorized`) с короткой ошибкой; токен НЕ
   попадает в backend-цепочку, лишний RPC не делается.
4. Валид → запрос проксируется в grpc-gateway mux через
   `IncomingHeaderMatcher` который пропускает `authorization` как
   gRPC-метадату.
5. Backend сервис **повторно** валидирует тот же токен в своём
   auth-interceptor'е (defense in depth).

Это даёт две полезных свойства:
- Bad-token каскадных вызовов нет (gateway отбрасывает его за 1 hop).
- Если злоумышленник попадёт в docker-net и пойдёт мимо gateway —
  каждый сервис всё равно проверит JWT сам.

## CORS

Middleware с явным allowlist (см. `transport/http/middleware.go`).
Origins берутся из конфига:

```yaml
cors:
  allowed_origins:
    - "http://localhost:5173"      # dev frontend
    - "https://app.beev.example"   # prod
```

Wildcard `*` запрещён (Bearer-headers требуют explicit origin). Methods:
`GET, POST, PUT, PATCH, DELETE, OPTIONS`. Headers: `Authorization,
Content-Type`. Preflight кэшируется на 86400s.

## OpenAPI

- Каждый бэкенд эмитит per-service spec через `protoc-gen-openapi`
  (gnostic).
- gateway собирает merged spec во время `bash scripts/generate.sh` —
  один YAML файл на всё API.
- На рантайме читается из embed-asset, конвертируется в JSON и
  отдаётся на `/swagger.json`.
- Scalar UI на `/docs` подхватывает этот JSON и рендерит интерактивную
  документацию с Try-it-out.

## Конфигурация

```yaml
http:
  addr: ":8080"
  read_header_timeout: 10s
  read_timeout: 30s
  write_timeout: 60s
  idle_timeout: 120s
auth:
  grpc_addr: "auth:50050"
  insecure: true
backends:
  auth:       "auth:50050"
  vacancy:    "vacancy:50051"
  resume:     "resume:50052"
  analysis:   "analysis:50054"
cors:
  allowed_origins: [...]
tls:
  cert_file: ""
  key_file: ""
```

`config/config.go::validate()` отбраковывает пустые обязательные поля
на старте.

## Запуск

```bash
make rebuild SVC=gateway
curl http://localhost:8080/healthz       # должен ответить 200
curl http://localhost:8080/docs           # открыть в браузере
```

## Тестирование

Gateway не имеет usecase-слоя → `make test` ничего не запускает для
него. Контрактное тестирование — на стороне backend-сервисов через
их protobuf-стабы. E2E на gateway отдельно не покрыто (политика
монорепо: только usecase unit-тесты).

## Известные ограничения

- Нет встроенного rate-limit'а на edge — backend-сервисы лимитируют
  сами (auth: login/register/refresh; multiagent: Yandex).
- Нет request-tracing (OpenTelemetry) — slog access-log даёт
  method/path/duration/code, но без correlation-id.
- TLS опционален; для prod рекомендуется делегировать service mesh /
  ingress, а не настраивать TLS в YAML.

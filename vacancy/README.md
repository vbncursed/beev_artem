# vacancy — техническое задание

## Назначение

CRUD-сервис вакансий. Хранит позицию (название + описание + взвешенные
навыки) и владеет полем `role` — ключом, по которому multiagent выбирает
промпт для AI-оценки кандидата. Роль определяется LLM через
`multiagent.ClassifyRole` (та же модель, что и пишет HR-вердикты);
keyword-табличный `DetectRole` остаётся как детерминированный fallback
на случай недоступности LLM.

## Архитектура

Clean architecture:

```
vacancy/internal/
├── domain/                       Vacancy, SkillWeight, ListInput, ...
├── usecase/                      бизнес-логика
│   ├── vacancy_service.go        ports + service struct
│   ├── create.go                 CreateVacancy (validate → resolveRole → store)
│   ├── get.go                    GetVacancy (owner-only / admin-bypass)
│   ├── list.go                   ListVacancies (с pagination + query)
│   ├── update.go                 UpdateVacancy (versioned, optimistic)
│   ├── archive.go                ArchiveVacancy
│   ├── role_classifier.go        порт RoleClassifier (LLM-based)
│   ├── resolve_role.go           wrapper: LLM with timeout → DetectRole fallback
│   ├── role_detector.go          keyword-based DetectRole (deterministic fallback)
│   ├── validate.go               правила валидации (см. ниже)
│   └── *_test.go                 unit-тесты, 100% coverage
├── infrastructure/
│   ├── persistence/              pgx + goose
│   │   ├── vacancy_storage.go    pool init + applyMigrations
│   │   ├── migrations/00001_*    initial schema
│   │   ├── migrations/00002_*    add_role
│   │   └── *.go                  по одному методу на файл
│   ├── multiagent_client/        gRPC client → multiagent.ClassifyRole
│   │   ├── client.go             dial + cleanup hook
│   │   └── classifier.go         RoleClassifier impl (wraps RPC errors as ErrLLMUnavailable)
│   └── auth_client/              gRPC client → auth.ValidateAccessToken
└── transport/
    ├── grpc/                     handlers + errdetails.ErrorInfo
    └── middleware/               Recovery + Logging + Auth (4 файла)
```

## API

| RPC | HTTP | Описание |
|---|---|---|
| `CreateVacancy` | `POST /api/v1/vacancies` | Создаёт вакансию. Бэкенд авто-нормализует веса (если все 0 → 1/N) и определяет роль через `multiagent.ClassifyRole` (с fallback на `DetectRole`). |
| `GetVacancy` | `GET /api/v1/vacancies/{vacancy_id}` | Только владелец или admin. |
| `ListVacancies` | `GET /api/v1/vacancies` | Page-based pagination + опциональный `query` для full-text по title+description. |
| `UpdateVacancy` | `PATCH /api/v1/vacancies/{vacancy_id}` | Обновляет; роль пересчитывается на каждом update. Optimistic concurrency через `version`. |
| `ArchiveVacancy` | `POST /api/v1/vacancies/{vacancy_id}/archive` | Soft delete: статус `archived`. |

## Domain model

```go
type Vacancy struct {
    ID            string         // varchar(64) UUID
    OwnerUserID   uint64
    Title         string         // ≤255 рун (utf8.RuneCountInString)
    Description   string         // ≤4000 рун
    Skills        []SkillWeight  // ≥1
    Role          string         // авто-determined; источник для multiagent
    Status        Status         // draft / open / archived
    Version       uint32         // optimistic concurrency
    CreatedAt     time.Time
    UpdatedAt     time.Time
}

type SkillWeight struct {
    Name       string             // ≤64 chars, non-empty after trim
    Weight     float32            // ∈ [0, 1]
    MustHave   bool
    NiceToHave bool               // взаимоисключаемо с MustHave (UI-уровень)
}
```

## Правила валидации (`usecase/validate.go`)

Источник истины — backend, frontend дублирует через zod на стороне UX:

- `title`: trim non-empty, ≤255 рун (`utf8.RuneCountInString`, не байт)
- `description`: optional, ≤4000 рун — соответствует counter'у на фронте,
  Cyrillic не "съедает" лимит вдвое
- `skills`: ≥1 элемент
- `skill.name`: trim non-empty
- `skill.weight`: ∈ [0, 1]
- `OwnerUserID` ≠ 0 — иначе `ErrUnauthorized`

### Нормализация весов (`normalizeSkills`)

Если все навыки имеют `Weight == 0` — backend ставит каждому `1/N`,
чтобы скоринг работал без ручной разметки. UI отдельно подсказывает
пользователю про это поведение.

## Определение роли

Роль выбирается из закрытого набора имён prompt-шаблонов, лежащих в
`multiagent/internal/infrastructure/prompts/templates/`:

```
accountant / analyst / doctor / electrician / manager / programmer / default
```

### Основной путь — LLM (`usecase/resolve_role.go`)

`resolveRole(ctx, title, description)`:

1. оборачивает контекст 5-секундным `WithTimeout` — мы не блокируем CRUD
   ради классификатора;
2. вызывает `RoleClassifier.Classify` (порт), реализованный в
   `infrastructure/multiagent_client/classifier.go` поверх gRPC RPC
   `multiagent.ClassifyRole`;
3. multiagent шлёт LLM-запрос с `temperature=0`, валидирует JSON-ответ
   `{"role": "<one>"}` против `PromptStore.ListRoles() ∪ {"default"}` и
   отсекает галлюцинации;
4. при любой ошибке (`Unavailable`, `DeadlineExceeded`, парсер,
   неизвестная роль) или пустом ответе — fallback на `DetectRole`.

Контракт: `resolveRole` всегда возвращает валидное имя из набора —
вакансия никогда не сохраняется с пустым или неизвестным `role`.

### Fallback — `DetectRole` (`usecase/role_detector.go`)

Keyword-table приоритезирована (специфичные роли — раньше generic'а):

```
1. accountant   — бухгалт / accountant / финансист / главбух
2. doctor       — врач / doctor / физиотерапевт / хирург / кардиолог
3. electrician  — электрик / electrician / электромонт
4. analyst      — аналитик / analyst / data scientist / data engineer
5. manager      — менедж / руководит / тимлид / product owner / директор
6. programmer   — програм / разработ / developer / engineer / Go / Python / ...
                  (catch-all для tech-ролей)
default         — если ни одно ключевое слово не сработало
```

Используется только когда LLM недоступен. Сохранять синхронизацию с
`templates/` имеет смысл, но критичность ниже — happy-path уходит в LLM.

### Добавление новой роли

1. положить файл промпта в
   `multiagent/internal/infrastructure/prompts/templates/<role>.txt` —
   `PromptStore.ListRoles()` подхватит автоматически, классификатор
   увидит новую роль на следующем запросе без сборки;
2. (опционально, для оффлайн-fallback) дополнить `roleKeywords` в
   `vacancy/internal/usecase/role_detector.go`;
3. обновить `KNOWN_ROLES` на фронте
   (`frontend/src/domain/vacancy/types.ts`).

## Зависимости

- **PostgreSQL** — таблица `vacancies`, JSONB-колонка `skills`. Миграции
  `00001_initial_schema.sql` (базовая схема), `00002_add_role.sql` (TEXT
  колонка).
- **auth** (gRPC) — каждый запрос проверяется через
  `auth.ValidateAccessToken` в auth-interceptor'е.
- **multiagent** (gRPC) — `ClassifyRole` для определения роли вакансии.
  Soft-зависимость: при недоступности vacancy продолжает работать через
  `DetectRole`-fallback.
- **Redis не используется**.

## Конфигурация

```yaml
database: { host, port, username, password, name, ssl_mode }
auth:
  grpc_addr: "auth:50050"
multiagent:
  grpc_addr: "multiagent:50055"
server:
  grpc_addr: ":50051"
  tls: { cert_file, key_file }
```

Секретов через env у vacancy нет (БД пароль идёт через compose).

## Тестирование

```bash
make test     # 100% coverage
make race
make cov
```

mocks: `VacancyStorage`, `RoleClassifier` (через minimock).

## Известные ограничения

- Нет full-text индекса в PostgreSQL — `query` использует `ILIKE %x%`,
  что неоптимально на больших объёмах. Для ≥100k вакансий понадобится
  GIN/tsvector.
- Нет пагинации по cursor (только offset) — для глубоких страниц это
  будет медленно.
- Optimistic concurrency через `version` реализована, но frontend пока
  не показывает 409-конфликты (nice-to-have).

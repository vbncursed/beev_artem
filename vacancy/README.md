# vacancy — техническое задание

## Назначение

CRUD-сервис вакансий. Хранит позицию (название + описание + взвешенные
навыки) и владеет полем `role` — ключом, по которому multiagent выбирает
промпт для AI-оценки кандидата. Авто-определяет роль по тексту вакансии
через локальный keyword-table — пользователь не задаёт её руками.

## Архитектура

Clean architecture:

```
vacancy/internal/
├── domain/                       Vacancy, SkillWeight, ListInput, ...
├── usecase/                      бизнес-логика
│   ├── vacancy_service.go        ports + service struct
│   ├── create.go                 CreateVacancy (с DetectRole + normalize)
│   ├── get.go                    GetVacancy (owner-only / admin-bypass)
│   ├── list.go                   ListVacancies (с pagination + query)
│   ├── update.go                 UpdateVacancy (versioned, optimistic)
│   ├── archive.go                ArchiveVacancy
│   ├── role_detector.go          keyword-based DetectRole
│   ├── validate.go               правила валидации (см. ниже)
│   └── *_test.go                 unit-тесты, 100% coverage
├── infrastructure/
│   ├── persistence/              pgx + goose
│   │   ├── vacancy_storage.go    pool init + applyMigrations
│   │   ├── migrations/00001_*    initial schema
│   │   ├── migrations/00002_*    add_role
│   │   └── *.go                  по одному методу на файл
│   └── auth_client/              gRPC client → auth.ValidateAccessToken
└── transport/
    ├── grpc/                     handlers + errdetails.ErrorInfo
    └── middleware/               Recovery + Logging + Auth (4 файла)
```

## API

| RPC | HTTP | Описание |
|---|---|---|
| `CreateVacancy` | `POST /api/v1/vacancies` | Создаёт вакансию. Бэкенд авто-нормализует веса (если все 0 → 1/N) и определяет роль через `DetectRole`. |
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

## DetectRole (`usecase/role_detector.go`)

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

`DetectRole(title, description)` лоуэркейсит haystack и возвращает первое
совпадение. Multiagent использует возвращённое значение для подгрузки
`assets/prompts/<role>.txt` с fallback'ом на `default.txt`.

Добавление новой роли: дополнить `roleKeywords` + положить новый файл
промпта в `multiagent/internal/infrastructure/prompts/templates/`.
Frontend `KNOWN_ROLES` тоже обновить (`frontend/src/domain/vacancy/types.ts`).

## Зависимости

- **PostgreSQL** — таблица `vacancies`, JSONB-колонка `skills`. Миграции
  `00001_initial_schema.sql` (базовая схема), `00002_add_role.sql` (TEXT
  колонка).
- **auth** (gRPC) — каждый запрос проверяется через
  `auth.ValidateAccessToken` в auth-interceptor'е.
- **Redis не используется**.

## Конфигурация

```yaml
database: { host, port, username, password, name, ssl_mode }
auth:
  grpc_addr: "auth:50050"
  insecure: true
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

mocks: `VacancyStorage` (через minimock).

## Известные ограничения

- Нет full-text индекса в PostgreSQL — `query` использует `ILIKE %x%`,
  что неоптимально на больших объёмах. Для ≥100k вакансий понадобится
  GIN/tsvector.
- Нет пагинации по cursor (только offset) — для глубоких страниц это
  будет медленно.
- Optimistic concurrency через `version` реализована, но frontend пока
  не показывает 409-конфликты (nice-to-have).

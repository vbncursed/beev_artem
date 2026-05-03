# multiagent — техническое задание

## Назначение

Генератор HR-вердикта (`hire` / `maybe` / `no`) + структурированного
обоснования через Yandex Cloud Foundation Models. Получает на вход
`(резюме-текст, профиль кандидата, требования вакансии, эвристический
скор)`, выбирает промпт по `role` и возвращает строго типизированный
JSON. Является internal-only сервисом — наружу через gateway не
выставлен. Единственный клиент — `analysis`.

## Архитектура

Clean architecture с **тремя** driven-портами:

```
multiagent/internal/
├── domain/
│   ├── decision.go               DecisionRequest, DecisionResponse,
│   │                             AgentResult
│   └── completion.go             CompletionRequest (port input для LLM)
├── usecase/
│   ├── multiagent_service.go     service struct + GenerateDecision +
│   │                             buildInput + isEmptyRequest +
│   │                             constants (105 LOC)
│   ├── decision_parser.go        llmDecision wire-shape +
│   │                             parseDecision + stripJSONFences (77 LOC)
│   ├── llm_port.go               LLM port + PromptStore port
│   └── *_test.go                 unit-тесты, 97.6% coverage
├── infrastructure/
│   ├── persistence/              pgx + goose
│   │   └── audit_decision.go     INSERT в multiagent_decisions
│   ├── llm/yandex/               реализация LLM port
│   │   └── client.go             rate-limited HTTP client с error mapping
│   └── prompts/                  реализация PromptStore port
│       ├── store.go              embed.FS lookup
│       └── templates/<role>.txt  programmer / manager / accountant /
│                                 doctor / electrician / analyst /
│                                 default
└── transport/
    └── grpc/                     handler — без auth-interceptor!
                                  multiagent доступен только с docker-
                                  net'а (analysis), не с публичного.
```

## API

Только **internal gRPC** — через grpc-gateway не пробрасывается.

| RPC | Описание |
|---|---|
| `GenerateDecision` | Принимает `DecisionRequest`, возвращает `DecisionResponse`. Записывает аудит в `multiagent_decisions`. |

## Domain model

```go
type DecisionRequest struct {
    Model              string         // override Yandex model id
    Role               string         // ключ для PromptStore
    VacancyMustHave    []string
    VacancyNiceToHave  []string
    CandidateSkills    []string
    CandidateSummary   string
    ResumeText         string
    MatchScore         float32
    MissingSkills      []string
    ScoreExplanation   string
}

type DecisionResponse struct {
    HRRecommendation  string         // "hire" | "maybe" | "no"
    Confidence        float32
    HRRationale       string         // RU (enforced)
    CandidateFeedback string         // RU
    SoftSkillsNotes   string         // RU
    AgentResults      []AgentResult
    RawTrace          string         // "yandex-llm-v1"
    CreatedAt         time.Time
}
```

## Поток `GenerateDecision`

```
1. usecase.GenerateDecision(ctx, req)
2. isEmptyRequest? → ErrInvalidArgument (защита от пустого запроса)
3. instructions = prompts.Get(req.Role) + languageDirective
4. input = JSON-сериализация структурированного payload (vacancy +
   candidate + score + missing_skills)
5. llm.Complete(ctx, CompletionRequest{
     Instructions, Input,
     Temperature: 0.3,             // низкая — это аналитика, не creative
     MaxOutputTokens: 1500,
   })
6. parseDecision(completion):
   а. stripJSONFences — убрать ```json ... ``` обёртку (chatty модели)
   б. Unmarshal в llmDecision wire-shape
   в. mapping → domain.DecisionResponse
7. RawTrace = "yandex-llm-v1", CreatedAt = now
8. storage.StoreDecision(ctx, req, resp) — аудит
9. Возврат resp
```

## Промпты

`infrastructure/prompts/templates/<role>.txt` — embed.FS, баклый
строковый шаблон. Каждый промпт описывает:

- роль агента (HR-аналитик специализированного домена)
- ожидаемый JSON-формат ответа (фиксированная схема)
- критерии оценки специфичные для роли

К каждому промпту runtime приклеивает `languageDirective` — ВАЖНУЮ
секцию из ~6 строк, требующую:

- все текстовые поля (`hr_rationale`, `candidate_feedback`,
  `soft_skills_notes`, `agent_results[*].summary`) — **строго на русском**
- `hr_recommendation` остаётся английским enum
- технологии и аббревиатуры (Go, PostgreSQL, gRPC, ИВЛ, МКБ-10) —
  не переводятся
- английские формулировки в резюме — пересказывать по-русски

Эта директива пин-нится в коде (а не в шаблонах) специально: новый
шаблон промпта не сможет случайно обнулить языковое требование.

### Текущие роли

| role | файл |
|---|---|
| `programmer` | software / data eng / devops / frontend |
| `manager` | менеджмент / тимлид / product / директор |
| `accountant` | бухгалтер / аудитор |
| `doctor` | врач / медицина |
| `electrician` | электрики и смежные ИТР |
| `analyst` | data scientist / аналитик / ML |
| `default` | catch-all для не-классифицированных вакансий |

Добавление новой роли = новый `.txt` файл + расширение
`vacancy/usecase/role_detector.go` keyword-table'а + `KNOWN_ROLES`
массив на frontend.

## Yandex LLM adapter

`infrastructure/llm/yandex/client.go`:

- HTTPS POST → `https://llm.api.cloud.yandex.net/foundationModels/v1/responses`
- Auth: `Authorization: Api-Key <YANDEX_API_KEY>`
- Body: `{model, messages, temperature, max_tokens}` где messages =
  `[{role: system, content: instructions}, {role: user, content: input}]`
- Rate limit (token-bucket): по умолчанию **10 rps / burst 5** —
  defends against runaway loop в analysis
- Mapping HTTP-ошибок:
  - 401/403 → `ErrLLMAuth`
  - 429 → `ErrLLMRateLimited`
  - 5xx → `ErrLLMUpstream`
  - таймаут / connection error → `ErrLLMUpstream`
  - malformed JSON в ответе → `ErrLLMInvalidResponse`

Все ошибки попадают в `analysis`, который их **молча проглатывает** —
эвристика остаётся authoritative.

## Зависимости

- **PostgreSQL** — таблица `multiagent_decisions` (JSONB request/response
  для аудита). Миграции goose.
- **Yandex Cloud Foundation Models** — внешний HTTPS endpoint.
- **Внутренних gRPC зависимостей нет** (multiagent ничего не зовёт).

## Конфигурация

```yaml
database: { ... }
yandex:
  folder_id: "b1g..."
  model: "aliceai-llm/latest"        # либо "yandexgpt/latest"
  request_timeout: 30s
  max_output_tokens: 1500
rate_limit:
  rps: 10
  burst: 5
server:
  grpc_addr: ":50055"
```

### Секреты

| Env | Required |
|---|---|
| `YANDEX_API_KEY` | ✓ |

Без auth-interceptor'а: multiagent доверяет docker-net'у (только
analysis может в него ходить). Если когда-нибудь multiagent станет
public-facing — копировать паттерн auth-interceptor'а из любого data-
сервиса.

## Тестирование

mocks: `DecisionStorage`, `LLM`, `PromptStore`. Реальный Yandex
endpoint в тестах не дёргается.

```bash
make test
make cov     # 97.6%
```

Так как multiagent ссылается на `domain.CompletionRequest` из своего
LLM-mock'а, есть единственный special-case: comment в `llm_port.go`
объясняет, почему нельзя завернуть mock через `package usecase_test` —
мы оставляем mock внутри `package usecase`, чтобы избежать import
cycle.

## Известные ограничения

- AgentResults пока всегда пустой (multi-agent режим — placeholder).
  Каркас domain.AgentResult и поле в proto зарезервированы.
- Нет per-tenant rate-limit'ов — глобальный bucket. Для multi-tenancy
  понадобится отдельный лимит на user_id.
- Нет retry с exponential backoff — `analysis` сам решает, когда
  перезвать. Защита от runaway-loop'ов на стороне rate-limiter'а.
- Yandex может вернуть JSON без обязательных полей — мы валидируем
  только `hr_recommendation` (без него `ErrLLMInvalidResponse`),
  остальные поля могут быть пустыми (UI это нормально обрабатывает).

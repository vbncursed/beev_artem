# analysis — техническое задание

## Назначение

Сервис скоринга и анализа кандидатов. Получает на вход пару
`(resume_id, vacancy_id)`, JOIN'ит данные через persistence-слой,
прогоняет эвристику (свой `Scorer`) для немедленного ответа, опционально
зовёт `multiagent` за LLM-обоснованной оценкой и сохраняет аудит в
таблицу `analyses`. Является единственным потребителем multiagent —
multiagent сам по себе наружу не торчит.

## Архитектура

Clean architecture с двумя driven-портами:

```
analysis/internal/
├── domain/
│   ├── domain.go                 Analysis, AIDecision, ScoreBreakdown,
│   │                             CandidateProfile, ListCandidatesByVacancyInput
│   └── analysis_pipeline.go      ResumeContext, AnalysisPayload,
│                                 SaveAnalysisInput
├── usecase/                      бизнес-логика
│   ├── analysis_service.go       ports + service
│   ├── start.go                  StartAnalysis (heuristic + опц LLM)
│   ├── get.go                    GetAnalysis
│   ├── list_by_vacancy.go        ListCandidatesByVacancy (sorted)
│   ├── update_ai_decision.go     перезаписать AI часть после LLM
│   └── *_test.go                 unit-тесты, 100%
├── infrastructure/
│   ├── persistence/              pgx + goose
│   ├── scorer/                   реализация Scorer (heuristic)
│   │   ├── scorer.go             Score method (115 LOC)
│   │   ├── extras.go             tokenizer + extractExtraSkills (75 LOC)
│   │   └── profile.go            yearsRe + summarize + round2 (55 LOC)
│   ├── multiagent_client/        gRPC client → multiagent
│   └── auth_client/              gRPC client → auth
└── transport/
    ├── grpc/                     handlers
    └── middleware/               Recovery + Logging + Auth (4 файла)
```

## API

| RPC | HTTP | Описание |
|---|---|---|
| `StartAnalysis` | `POST /api/v1/resumes/{resume_id}/analyze` | Запускает анализ. Эвристика всегда; LLM — если `useLlm: true`. Сохраняет в БД, возвращает `analysis_id`. |
| `GetAnalysis` | `GET /api/v1/analyses/{analysis_id}` | Полный объект Analysis (profile + breakdown + ai). |
| `ListCandidatesByVacancy` | `GET /api/v1/vacancies/{vacancy_id}/candidates` | Сортированный список кандидатов с их аналитикой. Фильтры: `minScore`, `requiredSkill`. Сортировка: `scoreOrder=SORT_ORDER_DESC` (proto enum по полному имени). |

## Domain model

```go
type Analysis struct {
    ID             string
    VacancyID      string
    CandidateID    string
    ResumeID       string
    VacancyVersion uint32         // снимок версии вакансии — для аудита
    Status         AnalysisStatus // queued / running / done / failed
    MatchScore     float32        // [0, 100]
    Profile        CandidateProfile
    Breakdown      ScoreBreakdown
    AI             AIDecision
    ErrorMessage   string         // не пусто при Status=failed
    CreatedAt      time.Time
    UpdatedAt      time.Time
}

type CandidateProfile struct {
    Skills          []string
    YearsExperience float32
    Positions       []string
    Technologies    []string
    Education       []string
    Summary         string         // ≤320 chars
}

type ScoreBreakdown struct {
    MatchedSkills   []string
    MissingSkills   []string
    ExtraSkills     []string       // что в резюме сверху требований
    BaseScore       float32
    MustHavePenalty float32         // -10 за каждый missed must-have
    NiceToHaveBonus float32         // +2 за каждый matched nice-to-have
    Explanation     string          // короткая тех. заметка
}

type AIDecision struct {
    HRRecommendation  string         // "hire" | "maybe" | "no" (enum-like)
    Confidence        float32
    HRRationale       string         // RU
    CandidateFeedback string         // RU
    SoftSkillsNotes   string         // RU
    AgentResults      []AgentResult  // мульти-агентный режим (placeholder)
    RawTrace          string         // "yandex-llm-v1" / "heuristic"
}
```

## Алгоритм скоринга (`Scorer`)

Реализация: `infrastructure/scorer/scorer.go`. Чистая функция, без I/O,
тред-безопасна (stateless).

1. Лоуэркейсим резюме и каждое требуемое skillName
2. Для каждого `skill` проверяем `strings.Contains(lowerText, lowerName)`:
   - hit → `matched += weight`, увеличиваем `niceMatched` если `NiceToHave`
   - miss → `missing`, увеличиваем `mustMissing` если `MustHave`
3. `baseScore = (matchedWeight / totalWeight) * 100` (clamp totalWeight ≥1)
4. `mustPenalty = mustMissingCount * 10`
5. `niceBonus = niceMatchedCount * 2`
6. `matchScore = clamp(baseScore - mustPenalty + niceBonus, 0..100)`
7. Tier:
   - `matchScore ≥ 75` → `hire` (confidence 0.82)
   - `matchScore ≥ 45` → `maybe` (0.67)
   - иначе → `no` (0.55)

### `extractExtraSkills` (`extras.go`)
Tokenizer `[a-zA-Z][a-zA-Z0-9+.#_-]{1,}`, частота ≥ 2, минимум 3 символа.
Тримит хвостовые `-_.` (артефакт PDF reflow). Stop-list из 16 слов
(`tool`, `info`, `data`, `team`, `code` и т.д.) — отфильтровывает шум.
Топ-8 по убыванию частоты, с детерминистическим тай-брейкером по алфавиту.

### `extractYearsExperience` (`profile.go`)
Regex `(?i)(\d{1,2})\s*\+?\s*-?\s*[xх]?\s*(?:год|лет|year)` ловит
"5 лет", "7+ years", "опыт 3 года", "3-х лет". Берёт максимум,
ограничивает [0, 50] чтобы не словить даты вроде "2024 год".
**Fallback** — у регексa нет шансов отличить «25 лет опыта» от
«(20 лет)» в строке «Дата рождения: 22.09.2004 (20 лет)». На
успешном LLM-проходе analysis перезаписывает значение через
`UpdateProfileYearsExperience` — именно LLM теперь authoritative.

### `summarize` (`profile.go`)
Превью для `Profile.Summary` — collapse whitespace + truncate до
**320 рун** (не байт; иначе срез ловит середину кириллической буквы
и в `gRPC` уезжает битый UTF-8). На успешном LLM-вызове
перезаписывается LLM-сгенерированным `candidate_summary` (1–2
предложения) через `UpdateProfileSummary`.

## Поток `StartAnalysis`

```
1. usecase.StartAnalysis(ctx, in)
2. JOIN: resume + candidate + vacancy → ResumeContext
3. Scorer.Score(text, vacancy.Skills) → AnalysisPayload (heuristic)
4. SaveAnalysis(ctx, payload, status=done) → возврат analysis_id
5. (если useLlm)
   а. ResumeText, Profile.Skills/Summary, Breakdown.{Missing,Matched,
      Explanation} → strings.ToValidUTF8 (defensive: pdftotext иногда
      выдаёт битые байты, gRPC marshaling роняет string field)
   б. multiagent.GenerateDecision(ctx, DecisionRequest{
        ...VacancyRole, ResumeText: <ПОЛНЫЙ текст резюме, не Summary>
      }) → DecisionResponse  (timeout 45s — Yandex full decision 5–15s)
   в. UpdateProfileYearsExperience(ctx, analysisID, yoe) — если
      maResp.years_experience > 0
   г. UpdateProfileSummary(ctx, analysisID, summary) — если
      maResp.candidate_summary не пустой
   д. UpdateAIDecision(ctx, analysisID, llmDecision) — перезаписываем
      AI-часть; profile.{years_experience, summary} уже обновлены
      выше, остальной profile / breakdown / score остаётся от
      эвристики
6. Если LLM упал: ничего не делаем (ошибка свопаем) → эвристика
   остаётся authoritative. Все три ветки `if err != nil` логируются
   через `slog.WarnContext` чтобы Yandex 4xx / timeouts / parse-errors
   не терялись.
```

**Почему ResumeText, а не Summary**: модели нужны полные периоды работы
(`Февраль 2025 — Октябрь 2025`, `11 МЕСЯЦЕВ`) чтобы посчитать
`years_experience`. 320-рунный preview обрезает их.

LLM-failures **намеренно проглатываются**: эвристика — fallback.

## Зависимости

- **PostgreSQL** — таблица `analyses` (JSONB-колонки для profile /
  breakdown / ai). Миграции goose.
- **auth** (gRPC) — каждый RPC проходит через auth-interceptor.
- **multiagent** (gRPC) — `infrastructure/multiagent_client/` тонкий
  адаптер. Используется только если `useLlm=true`. Кап вызова —
  `multiagentTimeout = 45s` (под Yandex `request_timeout=60s`,
  с headroom). Контракт `GenerateDecisionResponse` сейчас включает
  опциональные `years_experience` (float, поле 9) и
  `candidate_summary` (string, поле 10) — analysis применяет их
  к `profile_json` через `jsonb_set` (один метод на файл в
  `persistence/update_profile_years.go` и `update_profile_summary.go`).

## Конфигурация

```yaml
database: { ... }
auth:
  grpc_addr: "auth:50050"
multiagent:
  grpc_addr: "multiagent:50055"
  insecure: true
server:
  grpc_addr: ":50054"
```

## Тестирование

mocks: `AnalysisStorage`, `Scorer`. multiagent клиент через
multiagent-mock в нём же.

```bash
make test
make cov     # 100%
```

## Известные ограничения

- LLM-вызов синхронный с `StartAnalysis` — даёт latency 2–10s. Для prod
  стоит вынести в outbox-worker (структурно подготовлено: status
  поле уже поддерживает `queued`/`running`/`done`/`failed`).
- `Scorer` — простая keyword-based эвристика. Не учитывает синонимы
  ("PostgreSQL" vs "Postgres"), регистр в pre-trim только частичный.
- Нет re-scoring при обновлении вакансии — старые анализы остаются на
  `vacancyVersion` старом снимке (это by-design, для аудита).

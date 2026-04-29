# Раздел 1. Архитектура, бизнес-логика и технологическая реализация системы

---

## 1.1 Название проекта

**Система автоматизированного анализа резюме и оценки соответствия кандидатов требованиям вакансий (AI Resume Screening System).**

---

## 1.2 Цель разработки

Разработать систему автоматизированного анализа резюме, которая:

* извлекает структурированные данные из резюме кандидатов;
* сопоставляет навыки кандидата с требованиями вакансии;
* рассчитывает коэффициент соответствия;
* формирует рекомендации HR-специалисту;
* генерирует обратную связь для кандидата.

Система предназначена для **ускорения процесса подбора персонала и повышения качества предварительного отбора кандидатов**.

---

## 1.3 Проблема существующих решений

В традиционных HR-процессах:

* рекрутеры вручную просматривают десятки или сотни резюме;
* сравнение навыков кандидата и требований вакансии выполняется субъективно;
* значительное время тратится на первичный отбор.

Основные проблемы:

* высокая трудоёмкость анализа резюме;
* субъективность оценки кандидатов;
* невозможность быстро обрабатывать большой поток откликов.

Это приводит к увеличению времени закрытия вакансии и снижению эффективности рекрутинга.

---

## 1.4 Предлагаемое решение

В разработанной системе:

* резюме автоматически анализируются системой;
* текст проходит NLP-обработку;
* извлекаются навыки, опыт и ключевые технологии;
* происходит автоматическое сопоставление с требованиями вакансии;
* рассчитывается коэффициент соответствия кандидата.

Дополнительно используется **LLM-модель** для:

* генерации рекомендаций HR;
* анализа soft-skills кандидата;
* формирования обратной связи кандидату.

---

# 2. Архитектура системы

---

## 2.1 Архитектурный стиль

Система реализуется в виде **микросервисной архитектуры**, включающей:

* Gateway-service (входная HTTP точка);
* Auth-service (внутренний gRPC-сервис пользователей);
* Vacancy-service (внутренний gRPC-сервис вакансий);
* Resume-service (внутренний gRPC-сервис резюме);
* Analysis-service (внутренний gRPC-сервис анализа);
* MultiAgent-service (внутренний gRPC-сервис мультиагентного ИИ);
* Redis;
* базу данных PostgreSQL.

Взаимодействие между внутренними сервисами осуществляется через **gRPC + Protocol Buffers**.

Внешние клиенты взаимодействуют с системой только через **Gateway-service**, который:

* принимает HTTP/JSON запросы;
* транслирует их в gRPC вызовы (grpc-gateway);
* возвращает JSON ответы.

---

## 2.2 Компоненты системы

### 1. Gateway-service

Функции:

* внешняя HTTP/JSON точка входа;
* трансляция HTTP в gRPC (grpc-gateway);
* маршрутизация запросов;
* rate limit;
* выдача OpenAPI/Scalar-документации.

Gateway-service является единственной точкой входа для клиента.

---

### 2. Auth-service

Функции:

* регистрация пользователей;
* аутентификация;
* выдача access/refresh токенов;
* получение данных текущего пользователя (`Me`);
* аудит событий;
* управление сессиями (refresh в Redis).

---

### 3. Vacancy-service

Функции:

* создание вакансий;
* редактирование вакансий;
* архивирование/удаление вакансий;
* хранение требований к кандидатам;
* хранение ключевых навыков и весов.

---

### 4. Resume-service

Функции:

* создание кандидата (профиль);
* загрузка файлов резюме;
* извлечение текста (PDF/DOCX/TXT);
* хранение оригинального файла и извлечённого текста;
* выдача метаданных резюме.

---

### 5. Analysis-service

Функции:

* запуск анализа резюме;
* NLP-предобработка;
* извлечение сущностей;
* расчёт скоринга соответствия;
* сохранение результатов анализа;
* выдача отчётов и списка кандидатов по вакансии.

---

### 6. MultiAgent-service

Функции:

* генерация рекомендаций HR;
* генерация обратной связи кандидату;
* анализ soft-skills;
* хранение нормализованного результата LLM.

---

# 3. Бизнес-логика системы

---

## 3.1 Регистрация и авторизация HR пользователя

Процесс:

1. пользователь вводит email и пароль;
2. система проверяет уникальность email;
3. пароль хэшируется;
4. пользователь сохраняется в базе данных;
5. пользователю выдаётся JWT токен и refresh token (refresh хранится в Redis).

---

## 3.2 Создание вакансии

HR создаёт вакансию и указывает:

* название вакансии;
* описание вакансии;
* список навыков;
* веса навыков (опционально).

На основе этих данных формируется **эталонный профиль вакансии**.

---

## 3.3 Загрузка резюме

Процесс:

1. HR создаёт кандидата в контексте вакансии;
2. HR загружает файл резюме;
3. Gateway передаёт файл в Resume-service;
4. Resume-service извлекает текст и сохраняет данные;
5. Resume-service инициирует анализ (или HR запускает анализ отдельно).

---

## 3.4 Анализ резюме и скоринг

Analysis-service выполняет:

* очистку текста;
* нормализацию и NLP-пайплайн;
* извлечение навыков/опыта/образования;
* сопоставление с вакансией;
* расчёт match_score (0..100);
* формирование breakdown (matched/missing/extra).

---

## 3.5 Генерация рекомендаций (мультиагентная система)

После расчёта скоринга Analysis-service обращается к MultiAgent-service:

* передаёт данные вакансии и профиля кандидата;
* получает рекомендацию HR (hire/maybe/no) + аргументы и confidence;
* получает фидбек кандидату (что улучшить) и результаты агентов (trace).

---

# 4. Протоколы взаимодействия

---

## 4.1 Внешний контур — HTTP/JSON

HTTP используется только на уровне **Gateway-service**.

Реализация выполняется через **grpc-gateway** на основе `.proto` контрактов (google.api.http).

---

## 4.2 Внутренний контур — gRPC

Взаимодействие между:

* Gateway-service и Auth/Vacancy/Resume/Analysis;
* Analysis-service и MultiAgent-service;

осуществляется по **gRPC с использованием Protocol Buffers**.

---

# 5. Полноценные Protocol Buffers контракты (proto3)

Ниже приведены **полные** `.proto` контракты (services + messages + enums + HTTP annotations для gateway).

---

## 5.1 common/v1/common.proto

```proto
syntax = "proto3";

package hr.common.v1;

option go_package = "github.com/your-org/hr-system/gen/go/hr/common/v1;commonv1";

import "google/protobuf/timestamp.proto";

message PageRequest {
  uint32 page_size = 1;      // 1..200
  string page_token = 2;     // opaque token
}

message PageResponse {
  string next_page_token = 1;
}

enum SortOrder {
  SORT_ORDER_UNSPECIFIED = 0;
  SORT_ORDER_ASC = 1;
  SORT_ORDER_DESC = 2;
}

message AuditMeta {
  string request_id = 1;
  string ip = 2;
  string user_agent = 3;
  google.protobuf.Timestamp occurred_at = 4;
}
```

---

## 5.2 auth/v1/auth.proto

```proto
syntax = "proto3";

package hr.auth.v1;

option go_package = "github.com/your-org/hr-system/gen/go/hr/auth/v1;authv1";

import "google/api/annotations.proto";
import "google/protobuf/timestamp.proto";

service AuthService {
  rpc Register(RegisterRequest) returns (AuthResponse) {
    option (google.api.http) = {
      post: "/api/v1/auth/register"
      body: "*"
    };
  }

  rpc Login(LoginRequest) returns (AuthResponse) {
    option (google.api.http) = {
      post: "/api/v1/auth/login"
      body: "*"
    };
  }

  rpc Refresh(RefreshRequest) returns (AuthResponse) {
    option (google.api.http) = {
      post: "/api/v1/auth/refresh"
      body: "*"
    };
  }

  rpc Logout(LogoutRequest) returns (LogoutResponse) {
    option (google.api.http) = {
      post: "/api/v1/auth/logout"
      body: "*"
    };
  }

  rpc Me(MeRequest) returns (MeResponse) {
    option (google.api.http) = {
      get: "/api/v1/auth/me"
    };
  }

  // Внутренний метод для gateway/сервисов: проверка токена.
  rpc ValidateAccessToken(ValidateAccessTokenRequest) returns (ValidateAccessTokenResponse);
}

message RegisterRequest {
  string email = 1;        // RFC 5322 simplified
  string password = 2;     // min 8
}

message LoginRequest {
  string email = 1;
  string password = 2;
}

message RefreshRequest {
  string refresh_token = 1;
}

message LogoutRequest {
  string refresh_token = 1;
}

message LogoutResponse {
  string status = 1; // "logged_out"
}

message AuthResponse {
  string user_id = 1;          // uuid
  string access_token = 2;     // JWT
  string refresh_token = 3;    // opaque
  bool requires_2fa = 4;       // если включено 2FA (опционально в будущем)
  string challenge_id = 5;     // если requires_2fa=true
}

message MeRequest {}

message MeResponse {
  string user_id = 1;
  string email = 2;
  string role = 3; // "HR" (MVP)
  google.protobuf.Timestamp created_at = 4;
}

message ValidateAccessTokenRequest {
  string access_token = 1;
}

message ValidateAccessTokenResponse {
  bool valid = 1;
  string user_id = 2;
  string email = 3;
  string role = 4;
}
```

---

## 5.3 vacancy/v1/vacancy.proto

```proto
syntax = "proto3";

package hr.vacancy.v1;

option go_package = "github.com/your-org/hr-system/gen/go/hr/vacancy/v1;vacancyv1";

import "google/api/annotations.proto";
import "google/protobuf/timestamp.proto";
import "hr/common/v1/common.proto";

service VacancyService {
  rpc CreateVacancy(CreateVacancyRequest) returns (VacancyResponse) {
    option (google.api.http) = {
      post: "/api/v1/vacancies"
      body: "*"
    };
  }

  rpc GetVacancy(GetVacancyRequest) returns (VacancyResponse) {
    option (google.api.http) = {
      get: "/api/v1/vacancies/{vacancy_id}"
    };
  }

  rpc ListVacancies(ListVacanciesRequest) returns (ListVacanciesResponse) {
    option (google.api.http) = {
      get: "/api/v1/vacancies"
    };
  }

  rpc UpdateVacancy(UpdateVacancyRequest) returns (VacancyResponse) {
    option (google.api.http) = {
      patch: "/api/v1/vacancies/{vacancy_id}"
      body: "*"
    };
  }

  rpc ArchiveVacancy(ArchiveVacancyRequest) returns (ArchiveVacancyResponse) {
    option (google.api.http) = {
      post: "/api/v1/vacancies/{vacancy_id}/archive"
      body: "*"
    };
  }
}

enum VacancyStatus {
  VACANCY_STATUS_UNSPECIFIED = 0;
  VACANCY_STATUS_ACTIVE = 1;
  VACANCY_STATUS_ARCHIVED = 2;
}

message SkillWeight {
  string name = 1;   // e.g. "Docker"
  float weight = 2;  // 0..1 (если 0 -> система нормализует автоматически)
  bool must_have = 3;
  bool nice_to_have = 4;
}

message Vacancy {
  string id = 1;                 // uuid
  string owner_user_id = 2;      // uuid (HR)
  string title = 3;
  string description = 4;
  repeated SkillWeight skills = 5;
  VacancyStatus status = 6;
  uint32 version = 7;            // инкремент при изменении требований (для пересчёта)
  google.protobuf.Timestamp created_at = 8;
  google.protobuf.Timestamp updated_at = 9;
}

message CreateVacancyRequest {
  string title = 1;
  string description = 2;
  repeated SkillWeight skills = 3;
}

message GetVacancyRequest {
  string vacancy_id = 1;
}

message ListVacanciesRequest {
  hr.common.v1.PageRequest page = 1;
  string query = 2; // поиск по title/description
}

message ListVacanciesResponse {
  repeated Vacancy vacancies = 1;
  hr.common.v1.PageResponse page = 2;
}

message UpdateVacancyRequest {
  string vacancy_id = 1;
  string title = 2;
  string description = 3;
  repeated SkillWeight skills = 4;
}

message ArchiveVacancyRequest {
  string vacancy_id = 1;
}

message ArchiveVacancyResponse {
  string status = 1; // "archived"
}

message VacancyResponse {
  Vacancy vacancy = 1;
}
```

---

## 5.4 resume/v1/resume.proto

```proto
syntax = "proto3";

package hr.resume.v1;

option go_package = "github.com/your-org/hr-system/gen/go/hr/resume/v1;resumev1";

import "google/api/annotations.proto";
import "google/protobuf/timestamp.proto";
import "hr/common/v1/common.proto";

service ResumeService {
  // Создание кандидата в контексте вакансии.
  rpc CreateCandidate(CreateCandidateRequest) returns (CandidateResponse) {
    option (google.api.http) = {
      post: "/api/v1/vacancies/{vacancy_id}/candidates"
      body: "*"
    };
  }

  rpc GetCandidate(GetCandidateRequest) returns (CandidateResponse) {
    option (google.api.http) = {
      get: "/api/v1/candidates/{candidate_id}"
    };
  }

  // Загрузка резюме как stream: first message = metadata, next = chunks.
  rpc UploadResume(stream UploadResumeRequest) returns (UploadResumeResponse) {
    option (google.api.http) = {
      post: "/api/v1/candidates/{candidate_id}/resume"
      body: "*"
    };
  }

  rpc GetResume(GetResumeRequest) returns (ResumeResponse) {
    option (google.api.http) = {
      get: "/api/v1/resumes/{resume_id}"
    };
  }
}

message Candidate {
  string id = 1;                // uuid
  string vacancy_id = 2;        // uuid
  string full_name = 3;
  string email = 4;
  string phone = 5;
  string source = 6;            // "hh", "linkedin", "manual", etc.
  string comment = 7;           // HR note
  google.protobuf.Timestamp created_at = 8;
}

message Resume {
  string id = 1;                // uuid
  string candidate_id = 2;      // uuid
  string file_name = 3;
  string file_type = 4;         // "pdf" | "docx" | "txt"
  uint64 file_size_bytes = 5;
  string storage_path = 6;      // path / object key
  string extracted_text = 7;    // optional: can be empty until parsed
  google.protobuf.Timestamp created_at = 8;
}

message CreateCandidateRequest {
  string vacancy_id = 1;
  string full_name = 2;
  string email = 3;
  string phone = 4;
  string source = 5;
  string comment = 6;
}

message GetCandidateRequest {
  string candidate_id = 1;
}

message CandidateResponse {
  Candidate candidate = 1;
}

message UploadResumeMeta {
  string candidate_id = 1;
  string file_name = 2;
  string file_type = 3; // "pdf"|"docx"|"txt"
}

message UploadResumeChunk {
  bytes data = 1;
}

message UploadResumeRequest {
  oneof payload {
    UploadResumeMeta meta = 1;
    UploadResumeChunk chunk = 2;
  }
}

message UploadResumeResponse {
  Resume resume = 1;
}

message GetResumeRequest {
  string resume_id = 1;
}

message ResumeResponse {
  Resume resume = 1;
}
```

---

## 5.5 analysis/v1/analysis.proto

```proto
syntax = "proto3";

package hr.analysis.v1;

option go_package = "github.com/your-org/hr-system/gen/go/hr/analysis/v1;analysisv1";

import "google/api/annotations.proto";
import "google/protobuf/timestamp.proto";
import "hr/common/v1/common.proto";

service AnalysisService {
  // Ставим задачу анализа (async) и возвращаем analysis_id.
  rpc StartAnalysis(StartAnalysisRequest) returns (StartAnalysisResponse) {
    option (google.api.http) = {
      post: "/api/v1/resumes/{resume_id}/analyze"
      body: "*"
    };
  }

  rpc GetAnalysis(GetAnalysisRequest) returns (AnalysisResponse) {
    option (google.api.http) = {
      get: "/api/v1/analyses/{analysis_id}"
    };
  }

  // Список кандидатов по вакансии с сортировкой по score.
  rpc ListCandidatesByVacancy(ListCandidatesByVacancyRequest) returns (ListCandidatesByVacancyResponse) {
    option (google.api.http) = {
      get: "/api/v1/vacancies/{vacancy_id}/candidates"
    };
  }
}

enum AnalysisStatus {
  ANALYSIS_STATUS_UNSPECIFIED = 0;
  ANALYSIS_STATUS_QUEUED = 1;
  ANALYSIS_STATUS_RUNNING = 2;
  ANALYSIS_STATUS_DONE = 3;
  ANALYSIS_STATUS_FAILED = 4;
}

message CandidateProfile {
  repeated string skills = 1;
  float years_experience = 2;
  repeated string positions = 3;
  repeated string technologies = 4;
  repeated string education = 5;
  string summary = 6; // краткое резюме (если нужно)
}

message ScoreBreakdown {
  repeated string matched_skills = 1;
  repeated string missing_skills = 2;
  repeated string extra_skills = 3;
  float base_score = 4;
  float must_have_penalty = 5;
  float nice_to_have_bonus = 6;
  string explanation = 7; // "почему такой скор"
}

message AgentResult {
  string agent_name = 1;          // "ExtractorAgent" и т.п.
  string summary = 2;             // краткий результат
  string structured_json = 3;     // JSON (если агент возвращает структуру)
  float confidence = 4;           // 0..1
}

message AIDecision {
  string hr_recommendation = 1;     // "hire"|"maybe"|"no"
  float confidence = 2;             // 0..1
  string hr_rationale = 3;

  string candidate_feedback = 4;
  string soft_skills_notes = 5;

  repeated AgentResult agent_results = 6; // опционально: результаты агентов
  string raw_trace = 7;                   // опционально: сырой трейс
}

message Analysis {
  string id = 1;                  // uuid
  string vacancy_id = 2;          // uuid
  string candidate_id = 3;        // uuid
  string resume_id = 4;           // uuid
  uint32 vacancy_version = 5;
  AnalysisStatus status = 6;
  float match_score = 7;          // 0..100
  CandidateProfile profile = 8;
  ScoreBreakdown breakdown = 9;
  AIDecision ai = 10;
  string error_message = 11;      // если failed
  google.protobuf.Timestamp created_at = 12;
  google.protobuf.Timestamp updated_at = 13;
}

message StartAnalysisRequest {
  string resume_id = 1;
  string vacancy_id = 2;          // можно не передавать, если резюме уже связано с вакансией
  bool use_llm = 3;               // true/false
}

message StartAnalysisResponse {
  string analysis_id = 1;
  AnalysisStatus status = 2;      // queued
}

message GetAnalysisRequest {
  string analysis_id = 1;
}

message AnalysisResponse {
  Analysis analysis = 1;
}

message ListCandidatesByVacancyRequest {
  string vacancy_id = 1;
  hr.common.v1.PageRequest page = 2;
  float min_score = 3;            // filter
  string required_skill = 4;      // filter
  hr.common.v1.SortOrder score_order = 5; // ASC/DESC
}

message CandidateWithAnalysis {
  string candidate_id = 1;
  string full_name = 2;
  string email = 3;
  string phone = 4;
  float match_score = 5;
  string analysis_id = 6;
  AnalysisStatus analysis_status = 7;
  google.protobuf.Timestamp created_at = 8;
}

message ListCandidatesByVacancyResponse {
  repeated CandidateWithAnalysis candidates = 1;
  hr.common.v1.PageResponse page = 2;
}
```

---

## 5.6 multiagent/v1/multiagent.proto

```proto
syntax = "proto3";

package hr.multiagent.v1;

option go_package = "github.com/your-org/hr-system/gen/go/hr/multiagent/v1;multiagentv1";

import "google/protobuf/timestamp.proto";

service MultiAgentService {
  rpc GenerateDecision(GenerateDecisionRequest) returns (GenerateDecisionResponse);
}

enum AgentMode {
  AGENT_MODE_UNSPECIFIED = 0;
  AGENT_MODE_FAST = 1;        // быстро, меньше агентов/контекста
  AGENT_MODE_BALANCED = 2;    // default
  AGENT_MODE_STRICT = 3;      // максимум проверок и аудита
}

message GenerateDecisionRequest {
  string model = 1;                 // e.g. "qwen-chat"
  AgentMode mode = 2;

  // Контекст вакансии
  string vacancy_title = 3;
  string vacancy_description = 4;
  repeated string vacancy_must_have = 5;
  repeated string vacancy_nice_to_have = 6;

  // Контекст кандидата
  repeated string candidate_skills = 7;
  repeated string missing_skills = 8;
  string candidate_summary = 9;

  // Контекст скоринга
  string score_explanation = 10;    // explanation из breakdown
  float match_score = 11;           // 0..100

  // Доп. контекст (опционально)
  string resume_text = 12;          // можно обрезать
}

message AgentResult {
  string agent_name = 1;
  string summary = 2;
  string structured_json = 3;       // JSON
  float confidence = 4;             // 0..1
}

message GenerateDecisionResponse {
  string hr_recommendation = 1;     // "hire"|"maybe"|"no"
  float confidence = 2;             // 0..1
  string hr_rationale = 3;

  string candidate_feedback = 4;
  string soft_skills_notes = 5;

  repeated AgentResult agent_results = 6;
  string raw_trace = 7;

  google.protobuf.Timestamp created_at = 8;
}

```

---

# 6. Хранение данных

---

## 6.1 PostgreSQL

Основное хранилище данных.

Хранятся:

* пользователи;
* вакансии;
* кандидаты;
* резюме;
* результаты анализа;
* аудит (по необходимости).

---

## 6.2 Redis

Используется для:

* хранения refresh-сессий;
* кеширования;
* очередей задач анализа;
* rate limiting.

---

# 7. Используемые технологии

---

## 7.1 Backend

* Go (Golang)
* gRPC + Protocol Buffers
* grpc-gateway (Gateway-service)

---

## 7.2 Frontend

* React
* Next.js
* HeroUI

---

## 7.3 База данных

* PostgreSQL
* Redis

---

## 7.4 Контейнеризация

* Docker
* Docker Compose

---

# 8. Развёртывание системы

Система разворачивается в Docker.

Сервисы:

* gateway-service
* auth-service
* vacancy-service
* resume-service
* analysis-service
* multiagent-service
* postgres
* redis
* frontend

Запуск системы:

```bash
docker-compose up
```

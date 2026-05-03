# resume — техническое задание

## Назначение

Хранит кандидатов и резюме-файлы. Принимает PDF / DOCX / TXT в
multipart-форме, извлекает текст, выделяет минимальный профиль
(имя / email / телефон) и возвращает упорядоченную пару
`(Candidate, Resume)`. Файл хранится в БД как `BYTEA` (для MVP — потом
можно вынести в S3-совместимое хранилище).

## Архитектура

Clean architecture:

```
resume/internal/
├── domain/                       Candidate, Resume, NewResumeData,
│                                 DownloadResumeInput, ResumeFile, ...
├── usecase/                      бизнес-логика
│   ├── resume_service.go         ports + service struct
│   ├── create_candidate.go       без файла — добавить вручную
│   ├── create_candidate_from_resume.go  primary path: файл → кандидат+resume
│   ├── ingest_resume_batch.go    bulk-import (CRM-импорты)
│   ├── upload_resume.go          stream RPC, для существующего кандидата
│   ├── get_candidate.go          / get_resume.go
│   ├── download_resume.go        возвращает file_data байтами
│   ├── delete_candidate.go       cascade на resumes через FK
│   └── *_test.go                 unit-тесты, 97.2% coverage
├── infrastructure/
│   ├── persistence/              pgx + goose
│   │   ├── resume_storage.go     pool init
│   │   ├── migrations/00001_*    candidates + resumes (FK ON DELETE CASCADE)
│   │   └── *.go                  один метод на файл
│   ├── extractor/                PDF/DOCX/TXT парсинг
│   │   ├── extract.go            primary: pdftotext shell-out;
│   │   │                         fallback: ledongthuc/pdf
│   │   └── detect.go             magic-byte type detection
│   ├── profile/profile_extractor.go  regex-based name/email/phone
│   └── auth_client/              gRPC → auth
└── transport/
    ├── grpc/                     один handler на файл
    └── middleware/               Recovery + Logging + Auth (4 файла)
```

## API

| RPC | HTTP | Описание |
|---|---|---|
| `CreateCandidate` | `POST /api/v1/vacancies/{vacancy_id}/candidates` | Кандидат без резюме. |
| `CreateCandidateFromResume` | `POST /api/v1/vacancies/{vacancy_id}/candidates/from-resume` | Primary path — base64-encoded `file_data` в JSON. Backend сам определит тип, извлечёт текст, заполнит профиль. |
| `IngestResume` | `POST /api/v1/resumes/intake` | То же, но без `vacancy_id` (сначала кандидат в общий пул). |
| `IngestResumeBatch` | `POST /api/v1/resumes/intake/batch` | Массовый импорт; каждый файл получает отдельный результат с error-полем. |
| `UploadResume` | (gRPC-only stream) | Дозалить резюме существующему кандидату чанками — для очень больших файлов. |
| `GetCandidate` | `GET /api/v1/candidates/{candidate_id}` | Метаданные кандидата. |
| `GetResume` | `GET /api/v1/resumes/{resume_id}` | Метаданные резюме (НЕ файл). |
| `DownloadResume` | `GET /api/v1/resumes/{resume_id}/download` | **Файл целиком** в proto `bytes` (base64 в JSON через grpc-gateway). |
| `DeleteCandidate` | `DELETE /api/v1/candidates/{candidate_id}` | Удаляет кандидата; resumes уходят через FK CASCADE. |

## Domain model

```go
type Candidate struct {
    ID          string  // VARCHAR(64) UUID
    VacancyID   string  // optional — пустой для intake
    OwnerUserID uint64
    FullName    string
    Email       string
    Phone       string
    Source      string  // ручной / интеграция / batch
    Comment     string
    CreatedAt   time.Time
}

type Resume struct {
    ID            string
    CandidateID   string
    FileName      string
    FileType      string   // "pdf" | "docx" | "txt"
    FileSizeBytes uint64
    StoragePath   string   // зарезервировано под S3; пока пустое
    ExtractedText string   // результат extractor — не PII-safe!
    CreatedAt     time.Time
}

type ResumeFile struct {     // только для DownloadResume
    FileName string
    FileType string
    Data     []byte
}
```

## Поток обработки `CreateCandidateFromResume`

1. Декодируем base64 → `[]byte`
2. `extractor.DetectFileType(fileName, data)` — magic bytes / расширение
3. `extractor.ExtractText(fileType, data)` — реальное извлечение
4. `profile.ExtractCandidateProfile(text, fileName)` — regex для name/email/phone
5. Single-transaction insert: candidate row + resume row → возврат пары

### Извлечение текста

#### PDF (primary): `pdftotext` (poppler-utils)
- В Dockerfile: `apk add --no-cache poppler-utils`
- Shell out: `pdftotext -layout -nopgbrk -enc UTF-8 - -` (stdin → stdout)
- 30-секундный таймаут через `context.WithTimeout`
- `-layout` сохраняет визуальные колонки, `-nopgbrk` убирает form-feed
  между страницами (иначе heuristic-extractors анализа ловят шум)

#### PDF (fallback): `github.com/ledongthuc/pdf`
- Используется если `pdftotext` отсутствует / упал / вернул пусто
- `slog.Warn` фиксирует переход на fallback
- Качество хуже: на PDF без позиционной разметки склеивает слова без
  пробелов ("ЭдуардКурочкин" вместо "Эдуард Курочкин")
- Гарантия для dev-окружения без poppler

#### DOCX
- `archive/zip` → `word/document.xml`
- `encoding/xml` token stream
- Между `<w:p>`/`<w:br>` — `\n`, между `<w:tab>` — `\t`, между фрагментами
  одного абзаца — пробел через `needsSpace` flag
- Лимит на распакованный размер 2 MB (защита от zip-бомб)

#### TXT
- Тривиально: `strings.TrimSpace(string(data))`

### Лимит на размер

`MaxExtractedTextBytes = 2 * 1024 * 1024` — после извлечения, защита от
PDF-ов с раздутым текстовым слоем.

## Зависимости

- **PostgreSQL** — таблицы `candidates`, `resumes`. FK
  `resumes.candidate_id REFERENCES candidates(id) ON DELETE CASCADE`.
- **auth** (gRPC) — auth-interceptor.
- **poppler-utils** (system binary) — `pdftotext` в Dockerfile.

## Конфигурация

```yaml
database: { host, port, username, password, name, ssl_mode }
auth:
  grpc_addr: "auth:50050"
  insecure: true
server:
  grpc_addr: ":50052"
  tls: { cert_file, key_file }
```

Секретов через env у resume нет (БД пароль через compose).

## Тестирование

mocks: `ResumeStorage`, `TextExtractor`, `ProfileExtractor`. Извлечение
из реальных файлов не покрыто unit-тестами (по политике монорепо — нет
infrastructure тестов); ручной smoke через `cmd/pdf-debug` (был, удалён
после починки) при подозрительных PDF-ах.

```bash
make test
make cov          # ≥97.2%
```

## Безопасность

- Файлы хранятся в БД как BYTEA — backup-ы шифруются на уровне диска / pg_dump
- `extracted_text` содержит исходный текст резюме — это PII; логировать
  его нельзя; в slog.Info нет полей с этим текстом
- Размер файла ограничен на frontend (10 MB) и на extractor
  (2 MB на текст)

## Известные ограничения

- Файлы в БД, не в S3. Для prod-деплоя с десятками тысяч резюме нужен
  blob-storage adapter (storage_path колонка зарезервирована).
- `ExtractCandidateProfile` парсит только name/email/phone — реальные
  поля (skills/years/positions) заполняет analysis-сервис через эвристику
  или LLM. Resume не делает это сам по соображениям границ ответственности.
- Нет анти-вирусной проверки загружаемых файлов — потенциально опасно
  если эти файлы будут отдаваться обратно по `DownloadResume`.
- Re-extraction для существующих резюме (после улучшения extractor'а) не
  реализован — нужно перезаливать вручную.

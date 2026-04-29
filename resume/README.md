# Resume Service

gRPC микросервис управления кандидатами и резюме.

## Методы

- `CreateCandidate`
- `CreateCandidateFromResume`
- `IngestResume`
- `IngestResumeBatch`
- `GetCandidate`
- `UploadResume`
- `GetResume`

## Форматы резюме

- `PDF`
- `DOCX`
- `TXT`

При загрузке файл сохраняется целиком, а также заполняется `extracted_text` из содержимого файла.

## HTTP маршруты через gateway

- `POST /api/v1/vacancies/{vacancy_id}/candidates`
- `POST /api/v1/vacancies/{vacancy_id}/candidates/from-resume`
- `POST /api/v1/resumes/intake`
- `POST /api/v1/resumes/intake/batch`
- `GET /api/v1/candidates/{candidate_id}`
- `GET /api/v1/resumes/{resume_id}`

## Сценарии

1. В конкретную вакансию:
- `POST /api/v1/vacancies/{vacancy_id}/candidates/from-resume`
- body: только `file_data` (base64).

2. В общий пул (без вакансии):
- `POST /api/v1/resumes/intake`
- body: только `file_data` (base64).

3. Пакетная загрузка:
- `POST /api/v1/resumes/intake/batch`
- body:

```json
{
  "files": [
    {
      "external_id": "cv-1",
      "file_data": "<BASE64_FILE_1>"
    },
    {
      "external_id": "cv-2",
      "file_data": "<BASE64_FILE_2>"
    }
  ]
}
```

`external_id` опционален и нужен для сопоставления результата с исходным файлом.

## Что извлекается автоматически

- тип файла (PDF/DOCX/TXT);
- `full_name`;
- `email`;
- `phone`;
- `extracted_text`.

Ручное поле `file_type` для intake-методов не требуется.

## Ограничения

- максимальный размер одного файла: `10MB`;
- максимальный размер batch: `50` файлов.

## Конфиги

Порядок выбора:

1. `configPath` (если задан)
2. `APP_ENV=prod` -> `config.docker.prod.yaml`
3. иначе -> `config.docker.dev.yaml`

## Запуск

```bash
make resume-rebuild-dev
```

## Порт

- gRPC: `50052`

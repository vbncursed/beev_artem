package resume

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// Resume хранит метаданные загруженного файла.
type Resume struct {
	ID         uuid.UUID
	Filename   string
	MimeType   string
	Size       int64
	StorageURI string
	CreatedAt  time.Time
}

// Parsed хранит извлечённый из резюме текст.
type Parsed struct {
	ResumeID uuid.UUID
	Text     string
}

// Repository — порт доступа к резюме.
type Repository interface {
	Create(ctx context.Context, r Resume) error
	SaveParsed(ctx context.Context, p Parsed) error
	GetParsed(ctx context.Context, resumeID uuid.UUID) (Parsed, error)
}

// ParsedResume — извлечённый чистый текст из резюме.
type ParsedResume struct {
	ResumeID uuid.UUID
	Text     string
}

// ResumeRepository — порт для хранения резюме и извлечённых текстов.
type ResumeRepository interface {
	Create(ctx context.Context, r Resume) error
	SaveParsed(ctx context.Context, parsed ParsedResume) error
	GetByID(ctx context.Context, id uuid.UUID) (Resume, *ParsedResume, error)
}

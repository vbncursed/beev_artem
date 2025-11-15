package resume

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// Resume хранит метаданные загруженного файла.
type Resume struct {
	ID         uuid.UUID
	OwnerID    uuid.UUID
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
	// meta
	GetMetaForOwner(ctx context.Context, ownerID, id uuid.UUID) (Resume, error)
	ListByOwner(ctx context.Context, ownerID uuid.UUID, limit, offset int) ([]Resume, error)
	// admin
	GetMetaAny(ctx context.Context, id uuid.UUID) (Resume, error)
	ListAll(ctx context.Context, limit, offset int) ([]Resume, error)
	// delete (returns deleted meta for file cleanup)
	DeleteForOwner(ctx context.Context, ownerID, id uuid.UUID) (Resume, error)
	DeleteAny(ctx context.Context, id uuid.UUID) (Resume, error)
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

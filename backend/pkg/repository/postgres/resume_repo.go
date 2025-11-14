package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/artem13815/hr/pkg/resume"
)

// ResumeRepository хранит резюме и извлечённый текст.
type ResumeRepository struct {
	pool *pgxpool.Pool
}

func NewResumeRepository(pool *pgxpool.Pool) (*ResumeRepository, error) {
	r := &ResumeRepository{pool: pool}
	if err := r.ensureSchema(context.Background()); err != nil {
		return nil, err
	}
	return r, nil
}

func (r *ResumeRepository) ensureSchema(ctx context.Context) error {
	_, err := r.pool.Exec(ctx, `
CREATE TABLE IF NOT EXISTS resumes (
	id UUID PRIMARY KEY,
	owner_id UUID,
	filename TEXT NOT NULL,
	mime_type TEXT NOT NULL,
	size_bytes BIGINT NOT NULL,
	storage_uri TEXT NOT NULL,
	created_at TIMESTAMPTZ NOT NULL
);
CREATE TABLE IF NOT EXISTS parsed_resumes (
	resume_id UUID PRIMARY KEY REFERENCES resumes(id) ON DELETE CASCADE,
	text TEXT NOT NULL
);
-- backfill for older schemas
ALTER TABLE resumes ADD COLUMN IF NOT EXISTS owner_id UUID;
`)
	return err
}

func (r *ResumeRepository) Create(ctx context.Context, rs resume.Resume) error {
	if rs.ID == uuid.Nil {
		rs.ID = uuid.New()
	}
	if rs.CreatedAt.IsZero() {
		rs.CreatedAt = time.Now().UTC()
	}
	_, err := r.pool.Exec(ctx, `
INSERT INTO resumes (id, owner_id, filename, mime_type, size_bytes, storage_uri, created_at)
VALUES ($1, $2, $3, $4, $5, $6, $7)
`, rs.ID, rs.OwnerID, rs.Filename, rs.MimeType, rs.Size, rs.StorageURI, rs.CreatedAt)
	return err
}

func (r *ResumeRepository) SaveParsed(ctx context.Context, p resume.Parsed) error {
	_, err := r.pool.Exec(ctx, `
INSERT INTO parsed_resumes (resume_id, text)
VALUES ($1, $2)
ON CONFLICT (resume_id) DO UPDATE SET text = EXCLUDED.text
`, p.ResumeID, p.Text)
	return err
}

func (r *ResumeRepository) GetParsed(ctx context.Context, resumeID uuid.UUID) (resume.Parsed, error) {
	row := r.pool.QueryRow(ctx, `
SELECT resume_id, text FROM parsed_resumes WHERE resume_id = $1
`, resumeID)
	var p resume.Parsed
	if err := row.Scan(&p.ResumeID, &p.Text); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return resume.Parsed{}, pgx.ErrNoRows
		}
		return resume.Parsed{}, err
	}
	return p, nil
}

func (r *ResumeRepository) GetMetaForOwner(ctx context.Context, ownerID, id uuid.UUID) (resume.Resume, error) {
	row := r.pool.QueryRow(ctx, `
SELECT id, owner_id, filename, mime_type, size_bytes, storage_uri, created_at
FROM resumes WHERE id = $1 AND owner_id = $2
`, id, ownerID)
	var m resume.Resume
	var created time.Time
	if err := row.Scan(&m.ID, &m.OwnerID, &m.Filename, &m.MimeType, &m.Size, &m.StorageURI, &created); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return resume.Resume{}, pgx.ErrNoRows
		}
		return resume.Resume{}, err
	}
	m.CreatedAt = created.UTC()
	return m, nil
}

func (r *ResumeRepository) ListByOwner(ctx context.Context, ownerID uuid.UUID, limit, offset int) ([]resume.Resume, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := r.pool.Query(ctx, `
SELECT id, owner_id, filename, mime_type, size_bytes, storage_uri, created_at
FROM resumes WHERE owner_id = $3
ORDER BY created_at DESC
LIMIT $1 OFFSET $2
`, limit, offset, ownerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var res []resume.Resume
	for rows.Next() {
		var m resume.Resume
		var created time.Time
		if err := rows.Scan(&m.ID, &m.OwnerID, &m.Filename, &m.MimeType, &m.Size, &m.StorageURI, &created); err != nil {
			return nil, err
		}
		m.CreatedAt = created.UTC()
		res = append(res, m)
	}
	return res, nil
}

func (r *ResumeRepository) GetMetaAny(ctx context.Context, id uuid.UUID) (resume.Resume, error) {
	row := r.pool.QueryRow(ctx, `
SELECT id, owner_id, filename, mime_type, size_bytes, storage_uri, created_at
FROM resumes WHERE id = $1
`, id)
	var m resume.Resume
	var created time.Time
	if err := row.Scan(&m.ID, &m.OwnerID, &m.Filename, &m.MimeType, &m.Size, &m.StorageURI, &created); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return resume.Resume{}, pgx.ErrNoRows
		}
		return resume.Resume{}, err
	}
	m.CreatedAt = created.UTC()
	return m, nil
}

func (r *ResumeRepository) ListAll(ctx context.Context, limit, offset int) ([]resume.Resume, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := r.pool.Query(ctx, `
SELECT id, owner_id, filename, mime_type, size_bytes, storage_uri, created_at
FROM resumes
ORDER BY created_at DESC
LIMIT $1 OFFSET $2
`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var res []resume.Resume
	for rows.Next() {
		var m resume.Resume
		var created time.Time
		if err := rows.Scan(&m.ID, &m.OwnerID, &m.Filename, &m.MimeType, &m.Size, &m.StorageURI, &created); err != nil {
			return nil, err
		}
		m.CreatedAt = created.UTC()
		res = append(res, m)
	}
	return res, nil
}

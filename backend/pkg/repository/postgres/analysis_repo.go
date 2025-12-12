package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/artem13815/hr/pkg/analysis"
)

// AnalysisRepository сохраняет результаты анализа.
type AnalysisRepository struct {
	pool *pgxpool.Pool
}

func NewAnalysisRepository(pool *pgxpool.Pool) (*AnalysisRepository, error) {
	r := &AnalysisRepository{pool: pool}
	if err := r.ensureSchema(context.Background()); err != nil {
		return nil, err
	}
	return r, nil
}

func (r *AnalysisRepository) ensureSchema(ctx context.Context) error {
	_, err := r.pool.Exec(ctx, `
CREATE TABLE IF NOT EXISTS analyses (
	id UUID PRIMARY KEY,
	resume_id UUID NOT NULL REFERENCES resumes(id) ON DELETE CASCADE,
	vacancy_id UUID NOT NULL REFERENCES vacancies(id) ON DELETE CASCADE,
	score REAL NOT NULL,
	model TEXT NOT NULL,
	report JSONB NOT NULL,
	created_at TIMESTAMPTZ NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_analyses_vacancy ON analyses(vacancy_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_analyses_resume ON analyses(resume_id, created_at DESC);
`)
	return err
}

func (r *AnalysisRepository) Create(ctx context.Context, a analysis.Analysis) (analysis.Analysis, error) {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	if a.CreatedAt.IsZero() {
		a.CreatedAt = time.Now().UTC()
	}
	reportJSON, err := json.Marshal(a.Report)
	if err != nil {
		return analysis.Analysis{}, err
	}
	_, err = r.pool.Exec(ctx, `
INSERT INTO analyses (id, resume_id, vacancy_id, score, model, report, created_at)
VALUES ($1, $2, $3, $4, $5, $6, $7)
`, a.ID, a.ResumeID, a.VacancyID, a.Score, a.Model, reportJSON, a.CreatedAt)
	if err != nil {
		return analysis.Analysis{}, err
	}
	return a, nil
}

func (r *AnalysisRepository) GetByID(ctx context.Context, id uuid.UUID) (analysis.Analysis, error) {
	row := r.pool.QueryRow(ctx, `
SELECT id, resume_id, vacancy_id, score, model, report, created_at
FROM analyses WHERE id = $1
`, id)
	var a analysis.Analysis
	var reportBytes []byte
	var created time.Time
	if err := row.Scan(&a.ID, &a.ResumeID, &a.VacancyID, &a.Score, &a.Model, &reportBytes, &created); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return analysis.Analysis{}, pgx.ErrNoRows
		}
		return analysis.Analysis{}, err
	}
	_ = json.Unmarshal(reportBytes, &a.Report)
	a.CreatedAt = created.UTC()
	return a, nil
}

func (r *AnalysisRepository) GetByIDForOwner(ctx context.Context, ownerID, id uuid.UUID) (analysis.Analysis, error) {
	row := r.pool.QueryRow(ctx, `
SELECT a.id, a.resume_id, a.vacancy_id, a.score, a.model, a.report, a.created_at
FROM analyses a
JOIN resumes r ON r.id = a.resume_id
WHERE a.id = $1 AND r.owner_id = $2
`, id, ownerID)
	var a analysis.Analysis
	var reportBytes []byte
	var created time.Time
	if err := row.Scan(&a.ID, &a.ResumeID, &a.VacancyID, &a.Score, &a.Model, &reportBytes, &created); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return analysis.Analysis{}, pgx.ErrNoRows
		}
		return analysis.Analysis{}, err
	}
	_ = json.Unmarshal(reportBytes, &a.Report)
	a.CreatedAt = created.UTC()
	return a, nil
}

func (r *AnalysisRepository) ListByOwner(ctx context.Context, ownerID uuid.UUID, limit, offset int) ([]analysis.Analysis, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := r.pool.Query(ctx, `
SELECT a.id, a.resume_id, a.vacancy_id, a.score, a.model, a.report, a.created_at
FROM analyses a
JOIN resumes r ON r.id = a.resume_id
WHERE r.owner_id = $3
ORDER BY a.created_at DESC
LIMIT $1 OFFSET $2
`, limit, offset, ownerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var res []analysis.Analysis
	for rows.Next() {
		var a analysis.Analysis
		var reportBytes []byte
		var created time.Time
		if err := rows.Scan(&a.ID, &a.ResumeID, &a.VacancyID, &a.Score, &a.Model, &reportBytes, &created); err != nil {
			return nil, err
		}
		_ = json.Unmarshal(reportBytes, &a.Report)
		a.CreatedAt = created.UTC()
		res = append(res, a)
	}
	return res, nil
}

func (r *AnalysisRepository) ListByVacancyForOwner(ctx context.Context, ownerID, vacancyID uuid.UUID, limit, offset int) ([]analysis.Analysis, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := r.pool.Query(ctx, `
SELECT a.id, a.resume_id, a.vacancy_id, a.score, a.model, a.report, a.created_at
FROM analyses a
JOIN vacancies v ON v.id = a.vacancy_id
WHERE a.vacancy_id = $3 AND v.owner_id = $4
ORDER BY a.created_at DESC
LIMIT $1 OFFSET $2
`, limit, offset, vacancyID, ownerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var res []analysis.Analysis
	for rows.Next() {
		var a analysis.Analysis
		var reportBytes []byte
		var created time.Time
		if err := rows.Scan(&a.ID, &a.ResumeID, &a.VacancyID, &a.Score, &a.Model, &reportBytes, &created); err != nil {
			return nil, err
		}
		_ = json.Unmarshal(reportBytes, &a.Report)
		a.CreatedAt = created.UTC()
		res = append(res, a)
	}
	return res, nil
}

func (r *AnalysisRepository) ListAll(ctx context.Context, limit, offset int) ([]analysis.Analysis, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := r.pool.Query(ctx, `
SELECT id, resume_id, vacancy_id, score, model, report, created_at
FROM analyses
ORDER BY created_at DESC
LIMIT $1 OFFSET $2
`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var res []analysis.Analysis
	for rows.Next() {
		var a analysis.Analysis
		var reportBytes []byte
		var created time.Time
		if err := rows.Scan(&a.ID, &a.ResumeID, &a.VacancyID, &a.Score, &a.Model, &reportBytes, &created); err != nil {
			return nil, err
		}
		_ = json.Unmarshal(reportBytes, &a.Report)
		a.CreatedAt = created.UTC()
		res = append(res, a)
	}
	return res, nil
}

func (r *AnalysisRepository) ListByVacancyAny(ctx context.Context, vacancyID uuid.UUID, limit, offset int) ([]analysis.Analysis, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := r.pool.Query(ctx, `
SELECT id, resume_id, vacancy_id, score, model, report, created_at
FROM analyses
WHERE vacancy_id = $3
ORDER BY created_at DESC
LIMIT $1 OFFSET $2
`, limit, offset, vacancyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var res []analysis.Analysis
	for rows.Next() {
		var a analysis.Analysis
		var reportBytes []byte
		var created time.Time
		if err := rows.Scan(&a.ID, &a.ResumeID, &a.VacancyID, &a.Score, &a.Model, &reportBytes, &created); err != nil {
			return nil, err
		}
		_ = json.Unmarshal(reportBytes, &a.Report)
		a.CreatedAt = created.UTC()
		res = append(res, a)
	}
	return res, nil
}

func (r *AnalysisRepository) DeleteForOwner(ctx context.Context, ownerID, id uuid.UUID) error {
	cmd, err := r.pool.Exec(ctx, `
DELETE FROM analyses a
USING resumes r
WHERE a.id = $1 AND a.resume_id = r.id AND r.owner_id = $2
`, id, ownerID)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (r *AnalysisRepository) DeleteAny(ctx context.Context, id uuid.UUID) error {
	cmd, err := r.pool.Exec(ctx, `DELETE FROM analyses WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

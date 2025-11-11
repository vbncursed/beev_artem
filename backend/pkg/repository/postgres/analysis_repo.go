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
`)
	return err
}

func (r *AnalysisRepository) Create(ctx context.Context, a analysis.Analysis) error {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	if a.CreatedAt.IsZero() {
		a.CreatedAt = time.Now().UTC()
	}
	reportJSON, err := json.Marshal(a.Report)
	if err != nil {
		return err
	}
	_, err = r.pool.Exec(ctx, `
INSERT INTO analyses (id, resume_id, vacancy_id, score, model, report, created_at)
VALUES ($1, $2, $3, $4, $5, $6, $7)
`, a.ID, a.ResumeID, a.VacancyID, a.Score, a.Model, reportJSON, a.CreatedAt)
	return err
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

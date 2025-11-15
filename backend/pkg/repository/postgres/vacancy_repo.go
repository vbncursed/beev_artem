package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/artem13815/hr/pkg/vacancy"
)

// VacancyRepository хранит вакансии и их навыки с весами.
type VacancyRepository struct {
	pool *pgxpool.Pool
}

func NewVacancyRepository(pool *pgxpool.Pool) (*VacancyRepository, error) {
	r := &VacancyRepository{pool: pool}
	if err := r.ensureSchema(context.Background()); err != nil {
		return nil, err
	}
	return r, nil
}

func (r *VacancyRepository) ensureSchema(ctx context.Context) error {
	_, err := r.pool.Exec(ctx, `
CREATE TABLE IF NOT EXISTS vacancies (
	id UUID PRIMARY KEY,
	owner_id UUID,
	title TEXT NOT NULL,
	description TEXT NOT NULL,
	created_at TIMESTAMPTZ NOT NULL
);
CREATE TABLE IF NOT EXISTS vacancy_skills (
	vacancy_id UUID NOT NULL REFERENCES vacancies(id) ON DELETE CASCADE,
	skill TEXT NOT NULL,
	weight REAL NOT NULL CHECK (weight >= 0 AND weight <= 1),
	PRIMARY KEY (vacancy_id, skill)
);
-- backfill for older schemas
ALTER TABLE vacancies ADD COLUMN IF NOT EXISTS owner_id UUID;
CREATE INDEX IF NOT EXISTS idx_vacancies_owner ON vacancies(owner_id);
`)
	return err
}

func (r *VacancyRepository) Create(ctx context.Context, v vacancy.Vacancy) error {
	if v.ID == uuid.Nil {
		v.ID = uuid.New()
	}
	if v.CreatedAt.IsZero() {
		v.CreatedAt = time.Now().UTC()
	}
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	_, err = tx.Exec(ctx, `
INSERT INTO vacancies (id, owner_id, title, description, created_at)
VALUES ($1, $2, $3, $4, $5)
`, v.ID, v.OwnerID, strings.TrimSpace(v.Title), v.Description, v.CreatedAt)
	if err != nil {
		return err
	}
	for _, sw := range v.Skills {
		_, err = tx.Exec(ctx, `
INSERT INTO vacancy_skills (vacancy_id, skill, weight)
VALUES ($1, $2, $3)
ON CONFLICT (vacancy_id, skill) DO UPDATE SET weight = EXCLUDED.weight
`, v.ID, strings.ToLower(strings.TrimSpace(sw.Skill)), sw.Weight)
		if err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

func (r *VacancyRepository) UpdateSkillsForOwner(ctx context.Context, ownerID, id uuid.UUID, skills []vacancy.SkillWeight) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	// Ensure ownership
	row := tx.QueryRow(ctx, `SELECT 1 FROM vacancies WHERE id = $1 AND owner_id = $2`, id, ownerID)
	var one int
	if err := row.Scan(&one); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return pgx.ErrNoRows
		}
		return err
	}
	_, err = tx.Exec(ctx, `DELETE FROM vacancy_skills WHERE vacancy_id = $1`, id)
	if err != nil {
		return err
	}
	for _, sw := range skills {
		_, err = tx.Exec(ctx, `
INSERT INTO vacancy_skills (vacancy_id, skill, weight)
VALUES ($1, $2, $3)
`, id, strings.ToLower(strings.TrimSpace(sw.Skill)), sw.Weight)
		if err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

func (r *VacancyRepository) GetByIDForOwner(ctx context.Context, ownerID, id uuid.UUID) (vacancy.Vacancy, error) {
	row := r.pool.QueryRow(ctx, `
SELECT id, owner_id, title, description, created_at FROM vacancies WHERE id = $1 AND owner_id = $2
`, id, ownerID)
	var v vacancy.Vacancy
	var created time.Time
	if err := row.Scan(&v.ID, &v.OwnerID, &v.Title, &v.Description, &created); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return vacancy.Vacancy{}, pgx.ErrNoRows
		}
		return vacancy.Vacancy{}, err
	}
	v.CreatedAt = created.UTC()
	// skills
	rows, err := r.pool.Query(ctx, `
SELECT skill, weight FROM vacancy_skills WHERE vacancy_id = $1
`, id)
	if err != nil {
		return vacancy.Vacancy{}, err
	}
	defer rows.Close()
	for rows.Next() {
		var sw vacancy.SkillWeight
		if err := rows.Scan(&sw.Skill, &sw.Weight); err != nil {
			return vacancy.Vacancy{}, err
		}
		v.Skills = append(v.Skills, sw)
	}
	// stable order
	sort.Slice(v.Skills, func(i, j int) bool { return v.Skills[i].Skill < v.Skills[j].Skill })
	return v, nil
}

func (r *VacancyRepository) ListByOwner(ctx context.Context, ownerID uuid.UUID, limit, offset int) ([]vacancy.Vacancy, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := r.pool.Query(ctx, `
SELECT v.id, v.owner_id, v.title, v.description, v.created_at,
	COALESCE(
		json_agg(json_build_object('skill', vs.skill, 'weight', vs.weight) ORDER BY vs.skill)
			FILTER (WHERE vs.skill IS NOT NULL),
		'[]'
	) AS skills
FROM vacancies v
LEFT JOIN vacancy_skills vs ON vs.vacancy_id = v.id
WHERE v.owner_id = $3
GROUP BY v.id
ORDER BY v.created_at DESC
LIMIT $1 OFFSET $2
`, limit, offset, ownerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var res []vacancy.Vacancy
	for rows.Next() {
		var v vacancy.Vacancy
		var created time.Time
		var skillsJSON []byte
		if err := rows.Scan(&v.ID, &v.OwnerID, &v.Title, &v.Description, &created, &skillsJSON); err != nil {
			return nil, err
		}
		v.CreatedAt = created.UTC()
		_ = json.Unmarshal(skillsJSON, &v.Skills)
		res = append(res, v)
	}
	return res, nil
}

// Admin: без фильтра владельца
func (r *VacancyRepository) GetByIDAny(ctx context.Context, id uuid.UUID) (vacancy.Vacancy, error) {
	row := r.pool.QueryRow(ctx, `
SELECT id, owner_id, title, description, created_at FROM vacancies WHERE id = $1
`, id)
	var v vacancy.Vacancy
	var created time.Time
	if err := row.Scan(&v.ID, &v.OwnerID, &v.Title, &v.Description, &created); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return vacancy.Vacancy{}, pgx.ErrNoRows
		}
		return vacancy.Vacancy{}, err
	}
	v.CreatedAt = created.UTC()
	rows, err := r.pool.Query(ctx, `SELECT skill, weight FROM vacancy_skills WHERE vacancy_id = $1`, id)
	if err != nil {
		return vacancy.Vacancy{}, err
	}
	defer rows.Close()
	for rows.Next() {
		var sw vacancy.SkillWeight
		if err := rows.Scan(&sw.Skill, &sw.Weight); err != nil {
			return vacancy.Vacancy{}, err
		}
		v.Skills = append(v.Skills, sw)
	}
	sort.Slice(v.Skills, func(i, j int) bool { return v.Skills[i].Skill < v.Skills[j].Skill })
	return v, nil
}

func (r *VacancyRepository) ListAll(ctx context.Context, limit, offset int) ([]vacancy.Vacancy, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := r.pool.Query(ctx, `
SELECT v.id, v.owner_id, v.title, v.description, v.created_at,
	COALESCE(
		json_agg(json_build_object('skill', vs.skill, 'weight', vs.weight) ORDER BY vs.skill)
			FILTER (WHERE vs.skill IS NOT NULL),
		'[]'
	) AS skills
FROM vacancies v
LEFT JOIN vacancy_skills vs ON vs.vacancy_id = v.id
GROUP BY v.id
ORDER BY v.created_at DESC
LIMIT $1 OFFSET $2
`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var res []vacancy.Vacancy
	for rows.Next() {
		var v vacancy.Vacancy
		var created time.Time
		var skillsJSON []byte
		if err := rows.Scan(&v.ID, &v.OwnerID, &v.Title, &v.Description, &created, &skillsJSON); err != nil {
			return nil, err
		}
		v.CreatedAt = created.UTC()
		_ = json.Unmarshal(skillsJSON, &v.Skills)
		res = append(res, v)
	}
	return res, nil
}

func (r *VacancyRepository) DeleteForOwner(ctx context.Context, ownerID, id uuid.UUID) error {
	cmd, err := r.pool.Exec(ctx, `DELETE FROM vacancies WHERE id = $1 AND owner_id = $2`, id, ownerID)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (r *VacancyRepository) DeleteAny(ctx context.Context, id uuid.UUID) error {
	cmd, err := r.pool.Exec(ctx, `DELETE FROM vacancies WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

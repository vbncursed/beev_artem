package persistence

import (
	"context"
	"fmt"

	"github.com/artem13815/hr/admin/internal/domain"
)

// GetSystemStats issues seven COUNT(*) queries in a single round trip
// via UNION ALL so the dashboard fetch is one network hop. PG plans
// each as a sequential scan on a single column → cheap until tens of
// millions of rows, at which point we'd add a materialised view.
func (s *AdminStorage) GetSystemStats(ctx context.Context) (*domain.SystemStats, error) {
	const query = `
SELECT
  (SELECT count(*) FROM auth_users)                                   AS users_total,
  (SELECT count(*) FROM auth_users WHERE role = 'admin')              AS admins_total,
  (SELECT count(*) FROM vacancies)                                    AS vacancies_total,
  (SELECT count(*) FROM candidates)                                   AS candidates_total,
  (SELECT count(*) FROM analyses)                                     AS analyses_total,
  (SELECT count(*) FROM analyses WHERE status = 'done')               AS analyses_done,
  (SELECT count(*) FROM analyses WHERE status = 'failed')             AS analyses_failed
`
	var stats domain.SystemStats
	err := s.db.QueryRow(ctx, query).Scan(
		&stats.UsersTotal,
		&stats.AdminsTotal,
		&stats.VacanciesTotal,
		&stats.CandidatesTotal,
		&stats.AnalysesTotal,
		&stats.AnalysesDone,
		&stats.AnalysesFailed,
	)
	if err != nil {
		return nil, fmt.Errorf("count system stats: %w", err)
	}
	return &stats, nil
}
